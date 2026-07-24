#!/usr/bin/env bash
# End-to-end stress test: fires a large OTLP load at the running stack, then waits
# for the worker to drain it into ClickHouse and reports throughput + correctness.
#
# Prereqs: `docker compose up -d` and migrations applied. This script starts the
# server + worker itself, runs the load, verifies ClickHouse counts, and cleans up.
#
# Usage: scripts/stress_test.sh [TOTAL_EVENTS] [BATCH] [CONCURRENCY] [DISTINCT]
set -euo pipefail
cd "$(dirname "$0")/.."

TOTAL=${1:-100000}
BATCH=${2:-100}
CONCURRENCY=${3:-64}
DISTINCT=${4:-50}

CH() { docker compose exec -T clickhouse clickhouse-client -q "$1"; }

echo "==> building binaries"
go build -o /tmp/errlog-server ./cmd/server
go build -o /tmp/errlog-worker ./cmd/worker
go build -o /tmp/errlog-loadtest ./cmd/loadtest

echo "==> truncating ClickHouse tables for a clean measurement"
for t in logs_local error_events_local issue_stats_local service_stats_local release_health_local; do
  CH "TRUNCATE TABLE error_logging.${t} ON CLUSTER errlog_cluster" >/dev/null 2>&1 || true
done

echo "==> recreating Kafka topic + clearing consumer group (clean-slate measurement)"
PARTS=${KAFKA_PARTITIONS:-6}
docker compose exec -T kafka kafka-topics --bootstrap-server localhost:29092 --delete --topic error-logs >/dev/null 2>&1 || true
sleep 2
docker compose exec -T kafka kafka-topics --bootstrap-server localhost:29092 --create --if-not-exists --topic error-logs --partitions "$PARTS" --replication-factor 1 >/dev/null 2>&1 || true
docker compose exec -T kafka kafka-consumer-groups --bootstrap-server localhost:29092 --delete --group error-logging-worker >/dev/null 2>&1 || true

echo "==> starting server + worker"
/tmp/errlog-server >/tmp/errlog-server.log 2>&1 & SRV=$!
/tmp/errlog-worker >/tmp/errlog-worker.log 2>&1 & WKR=$!
cleanup() { kill -TERM "$SRV" "$WKR" 2>/dev/null || true; }
trap cleanup EXIT
until curl -sf http://localhost:8080/api/health/ >/dev/null 2>&1; do sleep 1; done

echo "==> registering a load-test service"
curl -s -X POST http://localhost:8080/api/projects -H 'Content-Type: application/json' -d '{"name":"stress"}' >/dev/null || true
RESP=$(curl -s -X POST http://localhost:8080/api/projects/1/services -H 'Content-Type: application/json' -d '{"name":"stress-svc"}')
KEY=$(echo "$RESP" | sed -E 's/.*"api_key":"([^"]+)".*/\1/')
PUBID=$(echo "$RESP" | sed -E 's/.*"public_id":"([^"]+)".*/\1/')

echo "==> firing load"
/tmp/errlog-loadtest -url http://localhost:8080 -key "$KEY" -public-id "$PUBID" \
  -events "$TOTAL" -batch "$BATCH" -concurrency "$CONCURRENCY" -distinct "$DISTINCT"

echo "==> waiting for the worker to drain into ClickHouse"
DRAIN_START=0
LAST=0; STABLE=0
while true; do
  CNT=$(CH "SELECT count() FROM error_logging.logs" 2>/dev/null || echo 0)
  # start the drain clock when the worker actually begins processing
  if [ "$DRAIN_START" -eq 0 ] && [ "$CNT" -gt 0 ]; then DRAIN_START=$(date +%s); fi
  if [ "$CNT" -ge "$TOTAL" ]; then break; fi
  # stop if the count stops advancing for ~30s (avoids hanging on lost events)
  if [ "$CNT" -eq "$LAST" ]; then STABLE=$((STABLE+1)); else STABLE=0; fi
  if [ "$STABLE" -ge 15 ]; then echo "   (count stalled at ${CNT}/${TOTAL})"; break; fi
  LAST=$CNT
  echo "   ClickHouse logs: ${CNT}/${TOTAL}"
  sleep 2
done
NOW=$(date +%s)
DRAIN_SECS=$(( DRAIN_START > 0 ? NOW - DRAIN_START : 0 ))
FINAL=$(CH "SELECT count() FROM error_logging.logs" 2>/dev/null || echo 0)
if [ "$DRAIN_SECS" -gt 0 ]; then
  echo "==> worker drained ${FINAL} events in ~${DRAIN_SECS}s → ~$(( FINAL / DRAIN_SECS )) events/sec"
fi

echo ""
echo "==> results"
CH "SELECT
  (SELECT count() FROM error_logging.logs)                          AS logs_stored,
  (SELECT countMerge(event_count) FROM error_logging.issue_stats)   AS counted_via_mv,
  (SELECT count() FROM error_logging.error_events)                  AS error_events_kept,
  (SELECT uniqExact(issue_id) FROM error_logging.logs WHERE issue_id>0) AS distinct_issues
FORMAT Vertical"
echo "worker drain time after load: ~${DRAIN_SECS}s"
echo "MySQL issues:"
docker compose exec -T mysql mysql -uroot error_logging -e "SELECT count(*) AS issue_rows FROM issues;" 2>/dev/null | grep -v "Using a password" || true
