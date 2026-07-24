# Observability Platform — Design Plan

An OTel-native, ClickHouse-backed observability platform — delivered as an **error-tracking
vertical slice first** (Sentry-style issues UI), built on an **all-logs foundation**
(SigNoz/Datadog-style), designed so a large org could run it at scale while everything runs
locally in Docker.

**Identity:** ingest ALL logs (every severity) via standard OTLP; error tracking
(fingerprinting, issues, grouping) is a **derived view** over the exception-bearing subset.
Later pillars — logs search, traces, metrics, dashboards/alerting — extend the same
foundation without rearchitecting. **No custom SDKs**: clients use their existing
OpenTelemetry SDKs/collectors; we publish a "point your OTel exporter here" doc per language.

---

## 1. High-level architecture

```
Client (any OTel SDK / collector — OUR SDK DOES NOT EXIST, by design)
   │  OTLP logs (JSON first, protobuf later) + X-API-Key
   ▼
┌──────────────────────────────────────────────────────────────┐
│ Ingest API (Gin) — stateless, N replicas behind LB           │
│   1. API-key middleware (Redis cache → MySQL fallback)       │
│   2. Validate / split OTLP batch                             │
│   3. Produce records → Kafka `logs` topic (records INLINE,   │
│      partitioned by hash(service_id))                        │
│   4. 202 Accepted — nothing heavy on the hot path            │
└──────────────────────────────────────────────────────────────┘
   ▼
┌──────────────────────────────────────────────────────────────┐
│ Workers — stateless Kafka consumer group (≤ partition count) │
│   EVERY record:                                              │
│     • batched bulk INSERT → ClickHouse `logs`                │
│     • async archive batch → S3 (audit/replay, NOT hot path)  │
│   EXCEPTION-bearing records (exception.type present) ALSO:   │
│     • parse exception.stacktrace → structured stack_frames   │
│     • fingerprint → dedupe/create issue (MySQL unique upsert)│
│     • regression check (resolved → regressed)                │
│     • rate-limit (10/s/fingerprint): gates FULL-FIDELITY row │
│       in `error_events` only — counting is never gated       │
└──────────────────────────────────────────────────────────────┘
   ▼
┌──────────────────────────────────────────────────────────────┐
│ ClickHouse — durable source of truth for ALL analytics       │
│   logs           (all severities, OTel log model, short TTL) │
│   error_events   (full fidelity, rate-limited, 90d TTL)      │
│   logs ─MV→ issue_stats     (counts + uniq users/sessions)   │
│   logs ─MV→ service_stats   (service overview rollup)        │
│   logs ─MV→ release_health  (crash-free sessions per release)│
└──────────────────────────────────────────────────────────────┘
   ▼ sync job (sharded across workers, idempotent, ~10s)
  MySQL issues: denormalized event_count / last_seen / affected_*

Read path: services/ read layer → Dashboard API (graphs)
                                → (future) MCP server → AI fix agent
                                → OpenReplay deep-links (session_id)
```

**Why this shape**
- All-logs volume is 10–100× errors, so the old per-event S3→Kafka-pointer→S3-fetch pattern
  is wrong here; records flow **through Kafka into batched ClickHouse inserts**, and S3
  becomes an **async archive sink**, off the hot path.
- ClickHouse is the durable source of truth for every number. **Redis is cache + rate-limit
  windows only — wiping Redis loses zero analytics** (anything with TTL/eviction can never
  be a source of truth; HLL distinct-state in particular is unreconstructible once lost).
- Kafka absorbs big-org ingest spikes as consumer lag, never as dropped data or a stalled API.

---

## 2. Storage responsibilities

| Store          | Role                                         | Holds |
|----------------|----------------------------------------------|-------|
| **ClickHouse** | All telemetry + analytics (**source of truth for numbers**) | logs, error_events, issue_stats, service_stats, release_health |
| **MySQL**      | Control plane + transactional issue state    | projects, services, issues (status, dedupe constraint, denormalized counts for list UI) |
| **Kafka**      | Elastic ingest buffer                        | `logs` topic, records inline, partitioned by hash(service_id) |
| **S3 / MinIO** | Async archive (audit / re-fingerprint replay)| batched raw OTLP, keyed by service/date |
| **Redis**      | **Cache + rate-limit windows ONLY**          | api-key cache, fingerprint→issue_id cache, per-fingerprint windows |

---

## 3. Data models

### 3.1 MySQL (control plane)

**projects** — `id, name, owner_id (nullable v1 — auth open question), created_at, updated_at`

**services**
```
id, project_id FK, name,
public_id      VARCHAR UNIQUE     -- in ingest URL, non-secret
api_key_hash   CHAR(64)           -- sha256(key); raw key shown ONCE at creation
api_key_prefix VARCHAR            -- "elk_live_ab12" for display
created_at, updated_at
```

**issues** (the group — one per fingerprint)
```
id BIGINT PK, service_id BIGINT FK,
fingerprint CHAR(64), title VARCHAR, culprit VARCHAR,   -- culprit = top in-app frame
level  ENUM(debug,info,warning,error,fatal),
status ENUM(unresolved,resolved,ignored,regressed),
first_seen DATETIME, last_seen DATETIME,
event_count BIGINT,                        -- ┐ denormalized for sortable/paginated lists;
affected_users_estimate BIGINT,            -- │ source of truth = ClickHouse issue_stats,
affected_sessions_estimate BIGINT,         -- ┘ synced ~10s by the sync job
regressed_at DATETIME NULL,
metadata JSON,                             -- representative sample: exception type/value,
                                           -- top frames, sample session_id
created_at, updated_at,
UNIQUE KEY (service_id, fingerprint),      -- dedupe guarantee; kills the need for any
                                           -- distributed lock on first-sighting races
INDEX (service_id, status, last_seen)
```

### 3.2 ClickHouse

> **Cluster-ready DDL from day one:** every table is written as `ReplicatedMergeTree`-family
> local table + `Distributed` table on top, coordinated by ClickHouse Keeper, shard key
> `cityHash64(service_id)` (tenant locality + cross-tenant spread). Locally this runs as a
> small 2-shard docker-compose cluster so scalability is demonstrated, not claimed. MVs run
> per-shard; reads go through Distributed with count/uniqMerge.

**A. `logs`** — the foundation. Every record, every severity, OTel log data model.
```
timestamp DateTime64(9), observed_at DateTime64(9),
project_id UInt64, service_id UInt64,
severity_number UInt8, severity_text LowCardinality(String),
body String,
trace_id String, span_id String,
-- promoted facets (typed columns = fast filters; the long tail stays in maps)
environment LowCardinality(String),        -- deployment.environment
release LowCardinality(String),            -- service.version
user_id String,                            -- user.id (enduser.id is deprecated)
session_id String,                         -- session.id (client OTel/OpenReplay provides it)
exception_type LowCardinality(String),     -- '' for non-error logs
exception_message String,
attributes Map(LowCardinality(String), String),
resource_attributes Map(LowCardinality(String), String)

ENGINE ReplicatedMergeTree  PARTITION BY toDate(timestamp)
ORDER BY (service_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 14 DAY     -- raw logs short; rollups keep history
```

**B. `error_events`** — full-fidelity debugging rows, **rate-limited/sampled**.
```
event_id UUID, issue_id UInt64, service_id UInt64, project_id UInt64,
timestamp DateTime64(9), fingerprint CHAR(64) → String,
severity_number UInt8, severity_text LowCardinality(String),
environment LowCardinality(String), release LowCardinality(String),
exception_type LowCardinality(String), exception_message String,
user_id String, session_id String,
stack_frames Array(Tuple(file String, function String, line UInt32, col UInt32, in_app UInt8)),
raw_stacktrace String,                     -- unparsed fallback
trace_id String, span_id String,
attributes Map(LowCardinality(String), String),
resource_attributes Map(LowCardinality(String), String),
s3_key String                              -- pointer into the archive batch

ENGINE ReplicatedMergeTree  PARTITION BY toYYYYMM(timestamp)
ORDER BY (service_id, issue_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
```

**C. `issue_stats`** — trend/count/affected source of truth (MV over `logs` WHERE exception_type != '').
```
service_id UInt64, issue_id UInt64, hour DateTime,
event_count AggregateFunction(count),
users    AggregateFunction(uniq, String),
sessions AggregateFunction(uniq, String)
ENGINE ReplicatedAggregatingMergeTree
PARTITION BY toYYYYMM(hour)  ORDER BY (service_id, issue_id, hour)
-- effectively permanent retention; aggregate states are tiny and mergeable
```
The MV needs `issue_id` on the row — the worker resolves fingerprint→issue_id BEFORE the
ClickHouse insert (Redis-cached; MySQL on miss), so `logs` carries `issue_id UInt64`
(0 for non-error records) purely to make this MV a plain projection. Every exception
record is counted — the `error_events` rate limit never touches this path.

**D. `service_stats`** — MV over all of `logs`: per service+hour event counts by severity,
distinct issues, distinct users. Powers the service-overview page.

**E. `release_health`** — crash-free sessions, derived (MV over `logs`).
```
service_id UInt64, release LowCardinality(String), environment LowCardinality(String),
hour DateTime,
sessions_total   AggregateFunction(uniq, String),   -- uniqState(session_id)
sessions_errored AggregateFunction(uniq, String)    -- uniqStateIf(session_id, severity>=17)
ENGINE ReplicatedAggregatingMergeTree
PARTITION BY toYYYYMM(hour)  ORDER BY (service_id, release, environment, hour)
```
`crash_free_rate = 1 − errored/total`. Renders only where clients stamp `session.id`
(OTel browser/mobile session tracking or OpenReplay); N/A gracefully otherwise — sessions
are a client-side concept we cannot synthesize server-side, and that's correct behavior
(crash-free is a frontend/mobile metric).

### 3.3 Why THREE error-path tables (logs / error_events / issue_stats)

- The rate limit must cut **row count**, not row size — columnar storage compresses empty
  fat columns to ~nothing, so "one table with skinny throttled rows" still accrues unbounded
  rows in a crash loop (10k/s × 90d = tens of billions). → fat `error_events` is sampled.
- But **counting must see every event** or trend graphs flatline exactly during the incident
  they exist to show. → complete `logs` stream feeds the MVs; rate limit gates fidelity only.
- Distinct users/sessions need `uniqState` (HLL) — ClickHouse-internal aggregate state,
  awkward to fabricate from Go. MVs are the idiomatic way; this mirrors Sentry's Snuba.

### 3.4 Partitioning & tenant isolation (the "one bad service" question)

- **Never `PARTITION BY service_id`** — high-cardinality partition keys cause part-count
  explosion / merge pressure / `TOO_MANY_PARTS`; that IS the "one bad service takes the
  system down" failure, caused by service-id partitioning rather than prevented by it.
- **PARTITION BY time** (toDate short-TTL, toYYYYMM long-lived): bounded partition count,
  cheap retention via whole-partition drops.
- **ORDER BY service_id first** = the real per-tenant query-isolation lever (primary-index
  granule skipping).
- Ingest-side noisy neighbors: per-service quotas/rate limits, Kafka partition spread,
  batched/async inserts (prevents tiny-insert part explosions), per-tenant query
  concurrency limits on the read path.

### 3.5 Kafka / S3

- Kafka `logs` topic: records inline (no S3 pointer round-trip), keyed by
  `hash(service_id)`, multiple partitions day one (partition count = max worker parallelism).
  Later pillars get their own topics (`traces`, `metrics`) so signals scale independently.
- S3/MinIO: workers flush batched raw OTLP async — `{service_id}/{yyyy}/{mm}/{dd}/{batch}.json`
  — for audit + re-fingerprinting replay. Never a synchronous hot-path step.

---

## 4. Worker pipeline (per record batch)

1. Consume batch from Kafka.
2. Every record → normalize to OTel log model.
3. If `exception.type` present:
   a. structured `stack_frames`: parse `exception.stacktrace` per-runtime (OTel only gives
      an unstructured string — this parser is OUR value-add, needed for fingerprinting,
      culprit, and structured AI/MCP output; raw string kept as fallback).
   b. fingerprint = hash(exception.type + normalized message template + top-N in-app
      frames' file+function), stripping dynamic tokens (numbers, UUIDs, hex, quoted strings);
      manual override honored via `log.fingerprint` attribute.
   c. resolve issue: Redis fingerprint→issue_id cache; on miss
      `INSERT ... ON DUPLICATE KEY UPDATE id=id` + SELECT on (service_id, fingerprint) —
      the unique constraint settles concurrent first-sightings, no distributed lock.
   d. regression: if issue.status == resolved → status = regressed, stamp regressed_at.
   e. rate-limit `INCR ratelimit:{fingerprint}:{sec}` EXPIRE 2; ≤10/s → queue full-fidelity
      `error_events` row (limit config-driven, per-service overridable later).
4. Batched inserts: `logs` (always, all records) + `error_events` (gated) via bulk/async
   insert. Async S3 archive flush.
5. **Sync job** (in-worker goroutine, ~10s tick, sharded by hash(service_id), idempotent —
   reads current ClickHouse aggregates and upserts absolute values, so concurrent/duplicate
   runs converge; no leader election): issue_stats → MySQL issues (event_count, last_seen,
   affected_users/sessions_estimate) as one batched UPDATE.

---

## 5. Ingest API

- `POST /api/v1/logs/:service_public_id` — OTLP `ExportLogsServiceRequest` JSON (same shape
  as the OTel Collector's `/v1/logs` HTTP receiver, so any OTel SDK/collector points at us
  with just an endpoint + header change). Protobuf later via `go.opentelemetry.io/proto/otlp`.
- `X-API-Key` middleware: sha256(key) → Redis → MySQL; attaches service+project to context.
- Validate, produce to Kafka, **202**. Stateless; scale = add replicas.

---

## 6. Product surface (error-tracking slice — Sentry/Bugsnag/Datadog parity)

**Issue list** — sort by volume / last-seen / users-affected; status filters; faceted search
on environment / release / exception_type (the promoted columns).

**Issue detail**
- trend graph from `issue_stats` (never flatlines — counting is pre-rate-limit)
- first/last seen, total events, affected users & sessions
- structured stack trace, in-app frames highlighted, culprit
- **breadcrumbs: derived at read time** — query `logs` for records sharing the error's
  `trace_id` (backend; free with standard OTel context propagation) or `session_id`
  (frontend), before the error timestamp, limit N. No stored column, no client buffer.
- facet breakdowns (environment/release/user), trace_id/span_id correlation
- OpenReplay deep-link via session_id; status + regression indicator

**Service overview** — top-N issues, error rate over time, distinct issues/users, DoD/WoW
(from `service_stats`).

**Release health** — crash-free-session rate per release (`release_health`), errors &
affected users by release, regression-on-deploy.

All reads go through the `services/` layer (`ListIssues`, `GetIssueDetail`,
`GetIssueTimeseries`, `GetBreadcrumbs`, `SearchEvents`, `GetServiceOverview`,
`GetReleaseHealth`) — controllers never touch gorm/ClickHouse directly. This layer is the
single product API for: dashboard now, **MCP tools later** (thin adapter, no duplicated
query logic, summary-vs-detail shaping for LLM context limits).

---

## 7. Scalability topology (designed distributed, run single-box)

| Layer      | Local (docker-compose)      | Scale-out (config/topology change, no rewrite) |
|------------|-----------------------------|------------------------------------------------|
| Ingest API | 1 container                 | N replicas behind LB (stateless)               |
| Kafka      | 1 broker, ≥6 partitions     | brokers + partitions; consumer lag absorbs spikes |
| Workers    | 1 container                 | scale to partition count (stateless group)     |
| ClickHouse | 2 shards × 1 replica + Keeper | add shards/replicas; DDL already Replicated+Distributed, shard key cityHash64(service_id) |
| MySQL      | 1 node                      | read replicas; writes stay light by design (new-issue inserts are rare — hot path is Redis-cached; counts never touch MySQL per-event) |
| Redis      | 1 node                      | Redis Cluster; loss costs zero analytics       |

---

## 8. Path to full observability (architected now, built later)

Error tracking is the first vertical; the foundation is deliberately signal-agnostic:
1. **Logs search UI** — the `logs` table already holds everything; add query endpoints + UI.
2. **Traces** — OTLP traces receiver → `traces` topic → spans table (SigNoz-style);
   trace_id/span_id are already first-class on logs for correlation.
3. **Metrics** — OTLP metrics receiver → dedicated time-series schema (label-set
   fingerprints + samples). Heaviest lift; explicitly deferred.
4. **Dashboards & alerting** — over the same read layer.
Ingest routes are namespaced (`/v1/logs`, later `/v1/traces`, `/v1/metrics`) so new signals
are new receivers + topics + tables — not a redesign.

## 9. Future integrations

- **OpenReplay**: frontend attaches `OpenReplay.getSessionID()` as `session.id` — no
  protocol change; we deep-link `/{project}/session/{session_id}?jumpto={offset}` from issue
  detail; later enrich AI context with session metadata via its API.
- **MCP + AI fix suggestions**: MCP server exposes the read layer as tools
  (ListIssues/GetIssueDetail/GetIssueTimeseries/SearchEvents/GetBreadcrumbs/
  GetSessionReplayContext); agent maps stack frames → repo source (GitHub API) → suggests
  root cause / where to debug. Schema already carries everything it needs (structured
  frames, typed exception fields, environment/release/user/session, trace correlation) —
  designed today precisely so this is never a re-migration.

---

## 10. Decisions log (what we chose and why — so we don't relitigate)

1. **All-logs foundation, error tracking as derived view** — Datadog/SigNoz direction;
   unlocks breadcrumbs + crash-free server-side and makes logs-search a query away.
2. **No custom SDKs** — OTel compliance means the ecosystem's SDKs are our SDKs; we ship
   per-language config docs instead. Breadcrumbs/sessions ride on standard OTel
   (trace propagation, browser/mobile session tracking), not proprietary clients.
3. **Records through Kafka; S3 async archive** — the per-event S3→pointer→fetch pattern
   died with all-logs volume.
4. **ClickHouse = source of truth for every number; Redis = cache only** — TTL/eviction
   can't corrupt analytics; HLL state is unreconstructible so it must live durable.
5. **MVs over Go-side aggregation** — uniqState/HLL is ClickHouse-internal; declarative,
   replayable, decoupled from worker bugs; the Snuba pattern.
6. **Rate limit gates fidelity, never counting** — trend graphs must not flatline during
   the incident they exist for.
7. **Unique constraint over Redlock** for first-sighting dedupe — cheaper, no TTL/clock
   assumptions, race is rare by construction.
8. **PARTITION BY time + ORDER BY service_id**, never PARTITION BY service_id.
9. **Crash-free included but honest** — renders where clients provide session.id; N/A
   otherwise. Stability score without a client-guaranteed denominator is not faked.
10. Facet/promoted columns locked from research: user_id, session_id, environment, release,
    exception_type/message, structured stack_frames, regressed_at + regressed status.

---

## 11. Folder structure (existing dig/Gin conventions)

```
pkg/models/            gorm: Project, Service, Issue
repositories/ + /impl  interfaces + gorm / clickhouse impls
services/ + /impl      ingest orchestration, fingerprinting, analytics reads, sync job,
                       ingest_processor (extracted from cmd/worker — testable, dig-injected)
pkg/client/s3/         S3-compatible client (MinIO ↔ real S3 = config swap)
pkg/otel/              OTLP JSON ↔ internal LogRecord + attribute extraction conventions
pkg/fingerprint/       grouping algorithm + per-runtime stacktrace parsers
controllers/impl/      Project, Service, Ingest, Analytics controllers
router/                one route file per controller, registered in di/container.go
```

## 12. docker-compose additions

MinIO + bucket-init; ClickHouse → 2-shard + Keeper topology; Kafka topic with ≥6 partitions.
MySQL/Redis already present.

---

## 13. Phased roadmap

1. **Control plane** — migrations (projects/services/issues), gorm repos, registration
   endpoints, API-key generation (hash + prefix, shown once).
2. **Infra wiring** — MinIO + bucket init, ClickHouse 2-shard compose + Keeper, cluster DDL
   (logs, error_events, issue_stats, service_stats, release_health + MVs), Kafka topic.
3. **Ingest API** — API-key middleware, OTLP JSON parsing, Kafka producer, 202 path.
4. **Worker** — consume/normalize, stacktrace parser, fingerprint, issue upsert + regression,
   rate limiter, batched CH inserts, async S3 archive.
5. **Sync job** — CH issue_stats → MySQL denormalized columns (sharded, idempotent).
6. **Read layer + Analytics API** — issues list/detail, timeseries, breadcrumbs-at-read,
   facets, service overview, release health.
7. **Dashboard** — separate frontend project (use the dataviz skill for chart design).
8. **OpenReplay** — session deep-links.
9. **MCP + AI fix agent** — tools over the read layer.
10. **Next pillars** — logs search UI → traces → metrics → alerting.

**Open question to settle in phase 1:** auth for project registration (user/org concept) vs
open registration — `projects.owner_id` is reserved either way; ingest-path API-key auth is
independent of this call.

---

## 14. Worker delivery semantics & resource bounds

**Delivery = at-least-once.** The batch consumer uses `FetchMessage` + explicit
`CommitMessages` *after* a successful ClickHouse flush. A crash between flush and commit
replays the cycle on restart. Consequences, accepted deliberately:
- **No duplicate issues** — `UNIQUE(service_id, fingerprint)` + idempotent `ResolveOrCreate`,
  and Kafka keys by `service_id` so one service maps to one partition/consumer.
- **Bounded duplicate rows** in `logs`/`error_events` on crash-replay → transient count drift
  in the MVs (they aggregate at insert time). Not fixed by `ReplacingMergeTree` (the MV has
  already counted); the honest options are "accept the small drift" (chosen) or de-dupe
  before the MV (stateful, not worth it). `event_id` is currently random, so replays are not
  retro-dedupable — a future deterministic `event_id` would be required for that.

**Memory is bounded per cycle** (only one cycle is in flight at a time):
- `fetch_max_bytes` caps raw payload bytes per cycle (not just `fetch_max_messages`), so a
  few large messages end a cycle early instead of buffering GBs.
- `flush_chunk_rows` caps rows per ClickHouse insert so the driver's batch buffer never holds
  a whole large cycle.
- `GOMEMLIMIT` (env / container) makes the GC defend a ceiling; docker-compose runs the
  worker with `mem_limit` above `GOMEMLIMIT` for headroom.
All four are env-overridable (`ERRLOG_WORKER_*`).

**Scaling to N workers:** safe — partition-per-service exclusivity + idempotent upserts +
atomic Redis ops mean no races. Requirement: **Kafka partitions ≥ worker replicas**
(`KAFKA_PARTITIONS` env), else extra workers idle.
