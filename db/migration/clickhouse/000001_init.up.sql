-- logs: the all-severities foundation table. Every OTLP log record lands here;
-- error tracking derives from the subset where exception_type != ''. issue_id is
-- resolved by the worker before insert (0 for non-error records) so the issue_stats
-- MV is a plain projection.
CREATE TABLE IF NOT EXISTS logs_local ON CLUSTER errlog_cluster
(
    timestamp           DateTime64(9),
    observed_at         DateTime64(9),
    project_id          UInt64,
    service_id          UInt64,
    issue_id            UInt64,
    severity_number     UInt8,
    severity_text       LowCardinality(String),
    body                String,
    trace_id            String,
    span_id             String,
    environment         LowCardinality(String),
    release             LowCardinality(String),
    user_id             String,
    session_id          String,
    exception_type      LowCardinality(String),
    exception_message   String,
    attributes          Map(LowCardinality(String), String),
    resource_attributes Map(LowCardinality(String), String)
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/logs_local', '{replica}')
PARTITION BY toDate(timestamp)
ORDER BY (service_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 14 DAY;

CREATE TABLE IF NOT EXISTS logs ON CLUSTER errlog_cluster AS logs_local
ENGINE = Distributed(errlog_cluster, currentDatabase(), logs_local, cityHash64(service_id));

-- error_events: full-fidelity debugging rows, rate-limited/sampled per fingerprint.
CREATE TABLE IF NOT EXISTS error_events_local ON CLUSTER errlog_cluster
(
    event_id            UUID,
    issue_id            UInt64,
    service_id          UInt64,
    project_id          UInt64,
    timestamp           DateTime64(9),
    ingested_at         DateTime64(9),
    severity_number     UInt8,
    severity_text       LowCardinality(String),
    environment         LowCardinality(String),
    release             LowCardinality(String),
    exception_type      LowCardinality(String),
    exception_message   String,
    user_id             String,
    session_id          String,
    stack_frames        Array(Tuple(file String, function String, line UInt32, col UInt32, in_app UInt8)),
    raw_stacktrace      String,
    trace_id            String,
    span_id             String,
    attributes          Map(LowCardinality(String), String),
    resource_attributes Map(LowCardinality(String), String),
    s3_key              String
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/error_events_local', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (service_id, issue_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS error_events ON CLUSTER errlog_cluster AS error_events_local
ENGINE = Distributed(errlog_cluster, currentDatabase(), error_events_local, cityHash64(service_id));

-- issue_stats: durable analytics source of truth (counts + distinct users/sessions).
CREATE TABLE IF NOT EXISTS issue_stats_local ON CLUSTER errlog_cluster
(
    service_id  UInt64,
    issue_id    UInt64,
    hour        DateTime,
    event_count AggregateFunction(count),
    users       AggregateFunction(uniq, String),
    sessions    AggregateFunction(uniq, String)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/{database}/issue_stats_local', '{replica}')
PARTITION BY toYYYYMM(hour)
ORDER BY (service_id, issue_id, hour);

CREATE TABLE IF NOT EXISTS issue_stats ON CLUSTER errlog_cluster AS issue_stats_local
ENGINE = Distributed(errlog_cluster, currentDatabase(), issue_stats_local, cityHash64(service_id));

CREATE MATERIALIZED VIEW IF NOT EXISTS issue_stats_mv ON CLUSTER errlog_cluster
TO issue_stats_local AS
SELECT
    service_id,
    issue_id,
    toStartOfHour(timestamp) AS hour,
    countState()             AS event_count,
    uniqState(user_id)       AS users,
    uniqState(session_id)    AS sessions
FROM logs_local
WHERE exception_type != ''
GROUP BY service_id, issue_id, hour;

-- service_stats: per-service overview rollup across all severities.
CREATE TABLE IF NOT EXISTS service_stats_local ON CLUSTER errlog_cluster
(
    service_id      UInt64,
    hour            DateTime,
    severity_number UInt8,
    event_count     AggregateFunction(count),
    issues          AggregateFunction(uniq, UInt64),
    users           AggregateFunction(uniq, String)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/{database}/service_stats_local', '{replica}')
PARTITION BY toYYYYMM(hour)
ORDER BY (service_id, hour, severity_number);

CREATE TABLE IF NOT EXISTS service_stats ON CLUSTER errlog_cluster AS service_stats_local
ENGINE = Distributed(errlog_cluster, currentDatabase(), service_stats_local, cityHash64(service_id));

CREATE MATERIALIZED VIEW IF NOT EXISTS service_stats_mv ON CLUSTER errlog_cluster
TO service_stats_local AS
SELECT
    service_id,
    toStartOfHour(timestamp) AS hour,
    severity_number,
    countState()                                AS event_count,
    uniqStateIf(issue_id, exception_type != '') AS issues,
    uniqState(user_id)                          AS users
FROM logs_local
GROUP BY service_id, hour, severity_number;

-- release_health: crash-free-session rate per release (severity >= 17 == ERROR).
CREATE TABLE IF NOT EXISTS release_health_local ON CLUSTER errlog_cluster
(
    service_id       UInt64,
    release          LowCardinality(String),
    environment      LowCardinality(String),
    hour             DateTime,
    sessions_total   AggregateFunction(uniq, String),
    sessions_errored AggregateFunction(uniq, String)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/{database}/release_health_local', '{replica}')
PARTITION BY toYYYYMM(hour)
ORDER BY (service_id, release, environment, hour);

CREATE TABLE IF NOT EXISTS release_health ON CLUSTER errlog_cluster AS release_health_local
ENGINE = Distributed(errlog_cluster, currentDatabase(), release_health_local, cityHash64(service_id));

CREATE MATERIALIZED VIEW IF NOT EXISTS release_health_mv ON CLUSTER errlog_cluster
TO release_health_local AS
SELECT
    service_id,
    release,
    environment,
    toStartOfHour(timestamp) AS hour,
    uniqStateIf(session_id, session_id != '')                           AS sessions_total,
    uniqStateIf(session_id, session_id != '' AND severity_number >= 17) AS sessions_errored
FROM logs_local
GROUP BY service_id, release, environment, hour;
