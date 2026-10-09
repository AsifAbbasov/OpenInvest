#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_DR_POSTGRES_CONTAINER:?set OPENINVEST_DR_POSTGRES_CONTAINER to a disposable DR PostgreSQL container}"
: "${OPENINVEST_DATABASE_TEST_URL:?set OPENINVEST_DATABASE_TEST_URL for the disposable DR database}"
: "${OPENINVEST_ALLOW_DESTRUCTIVE_DRILL:?set OPENINVEST_ALLOW_DESTRUCTIVE_DRILL=YES}"

if [[ "$OPENINVEST_ALLOW_DESTRUCTIVE_DRILL" != "YES" ]]; then
  echo "refusing DR outage injection without OPENINVEST_ALLOW_DESTRUCTIVE_DRILL=YES" >&2
  exit 2
fi

if [[ ! "$OPENINVEST_DR_POSTGRES_CONTAINER" =~ [Oo]pen[Ii]nvest.*[Dd][Rr]|[Dd][Rr].*[Oo]pen[Ii]nvest ]]; then
  echo "refusing to stop container whose name does not clearly identify an OpenInvest DR fixture" >&2
  exit 2
fi

command -v docker >/dev/null || { echo "docker is required" >&2; exit 2; }
command -v psql >/dev/null || { echo "psql is required" >&2; exit 2; }

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "[preflight] database must be reachable before outage injection"
psql "$OPENINVEST_DATABASE_TEST_URL" -Atv ON_ERROR_STOP=1 -c 'SELECT 1' | grep -qx '1'

echo "[fault] stopping disposable PostgreSQL container"
docker stop "$OPENINVEST_DR_POSTGRES_CONTAINER" >/dev/null

if psql "$OPENINVEST_DATABASE_TEST_URL" -At -c 'SELECT 1' >/dev/null 2>&1; then
  echo "FAIL: database remained reachable after declared outage" >&2
  docker start "$OPENINVEST_DR_POSTGRES_CONTAINER" >/dev/null || true
  exit 1
fi

echo "[recovery] restarting PostgreSQL container"
docker start "$OPENINVEST_DR_POSTGRES_CONTAINER" >/dev/null

for _ in $(seq 1 30); do
  if psql "$OPENINVEST_DATABASE_TEST_URL" -At -c 'SELECT 1' >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
psql "$OPENINVEST_DATABASE_TEST_URL" -Atv ON_ERROR_STOP=1 -c 'SELECT 1' | grep -qx '1'

echo "[recovery] running restart/idempotency and ledger-readiness witnesses"
cd "$repo_root/backend-go"
go test ./internal/postgres \
  -run '^(TestStage0332ReplaySurvivesStoreRestart|TestStage371RuntimeReadinessExecutesExactIndexGate)$' \
  -count=1 \
  -v

echo "DISASTER_RECOVERY_DRILL=PASS"
