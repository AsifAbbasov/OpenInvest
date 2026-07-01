#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_BASE_URL="${OPENINVEST_API_BASE_URL:-http://localhost:8080}"
CONTROLLED_API_BASE_URL="http://localhost:8080"
POSTGRES_DB="${POSTGRES_DB:-openinvest}"
POSTGRES_USER="${POSTGRES_USER:-openinvest}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-openinvest-local}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
DATABASE_URL="${DATABASE_URL:-postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable}"
API_LOG="${TMPDIR:-/tmp}/openinvest-stage-03-04-api.log"

export POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD POSTGRES_PORT DATABASE_URL

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

json_field() {
  python3 -c 'import json, sys; print(json.load(sys.stdin)'"$1"')'
}

wait_for_api() {
  for _ in $(seq 1 40); do
    if curl --fail --silent "${API_BASE_URL}/api/v1/ready" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "Go API did not become ready. Last API log:" >&2
  tail -n 80 "$API_LOG" >&2 || true
  return 1
}

start_controlled_api() {
  if [[ "$API_BASE_URL" != "$CONTROLLED_API_BASE_URL" ]]; then
    echo "Stage 3.4 smoke starts the Go API itself and currently supports only ${CONTROLLED_API_BASE_URL}." >&2
    echo "Unset OPENINVEST_API_BASE_URL or update the Go API listen address before using another URL." >&2
    exit 1
  fi

  if curl --fail --silent "${API_BASE_URL}/api/v1/health" >/dev/null 2>&1; then
    echo "A Go API is already responding at ${API_BASE_URL}." >&2
    echo "Stop the existing API before running this smoke test so the script can prove its own DATABASE_URL wiring." >&2
    exit 1
  fi

  echo "Starting Go API at ${API_BASE_URL}"
  (
    cd "$ROOT_DIR/backend-go"
    go run ./cmd/api
  ) >"$API_LOG" 2>&1 &
  API_PID=$!
  trap 'kill "${API_PID:-}" >/dev/null 2>&1 || true' EXIT
  wait_for_api
}

assert_database_state() {
  local portfolio_id="$1"
  local counts
  local portfolio_count
  local transaction_count
  local snapshot_count

  counts="$(
    docker compose exec -T postgres psql \
      -v ON_ERROR_STOP=1 \
      -v portfolio_id="$portfolio_id" \
      -U "$POSTGRES_USER" \
      -d "$POSTGRES_DB" \
      -At \
      -F '|' <<'SQL'
SELECT
    (SELECT count(*) FROM investment.portfolios WHERE id = :'portfolio_id'::uuid),
    (SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id = :'portfolio_id'::uuid),
    (SELECT count(*) FROM analytics.portfolio_snapshots WHERE portfolio_id = :'portfolio_id'::uuid AND snapshot_status = 'calculated');
SQL
  )"

  IFS='|' read -r portfolio_count transaction_count snapshot_count <<< "$counts"

  if [[ "$portfolio_count" != "1" ]]; then
    echo "Expected portfolio to exist in Docker PostgreSQL, got count ${portfolio_count}" >&2
    exit 1
  fi
  if [[ "$transaction_count" != "2" ]]; then
    echo "Expected 2 transaction entries in Docker PostgreSQL, got ${transaction_count}" >&2
    exit 1
  fi
  if [[ "$snapshot_count" -lt "2" ]]; then
    echo "Expected at least 2 calculated snapshots in Docker PostgreSQL, got ${snapshot_count}" >&2
    exit 1
  fi
}

apply_migration_if_needed() {
  local exists
  exists="$(
    docker compose exec -T postgres psql \
      -U "$POSTGRES_USER" \
      -d "$POSTGRES_DB" \
      -Atqc "SELECT to_regclass('investment.portfolios') IS NOT NULL;"
  )"

  if [[ "$exists" == "t" ]]; then
    echo "Database schema already exists; skipping Stage 3.1 migration replay."
    return 0
  fi

  echo "Applying Stage 3.1 migration."
  docker compose exec -T postgres psql \
    -v ON_ERROR_STOP=1 \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    < "$ROOT_DIR/infrastructure/postgres/migrations/000001_stage_03_01_vertical_slice.up.sql"
}

api_post() {
  local path="$1"
  local key="$2"
  local body="$3"
  curl --fail --silent \
    -X POST "${API_BASE_URL}${path}" \
    -H "Content-Type: application/json" \
    -H "Idempotency-Key: ${key}" \
    --data "$body"
}

api_get() {
  curl --fail --silent "${API_BASE_URL}$1"
}

require_command curl
require_command docker
require_command go
require_command python3

cd "$ROOT_DIR"
echo "Starting local PostgreSQL and Redis."
docker compose up -d postgres redis >/dev/null

echo "Waiting for PostgreSQL readiness."
for _ in $(seq 1 40); do
  if docker compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
docker compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null

apply_migration_if_needed
start_controlled_api

suffix="$(date +%s)"
portfolio_response="$(
  api_post "/api/v1/portfolios" "stage-03-04-portfolio-${suffix}" \
    '{"name":"Stage 3.4 E2E portfolio","baseCurrency":"RUB"}'
)"
portfolio_id="$(printf '%s' "$portfolio_response" | json_field "['data']['id']")"
echo "Created portfolio: ${portfolio_id}"

api_post "/api/v1/portfolios/${portfolio_id}/transactions" "stage-03-04-deposit-${suffix}" \
  '{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"10000.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-01-09","settlementDate":"2026-01-09","note":"Stage 3.4 smoke deposit"}' >/dev/null

api_post "/api/v1/portfolios/${portfolio_id}/transactions" "stage-03-04-buy-${suffix}" \
  '{"transactionType":"BUY","ticker":"SBER","quantity":"10.00000000","unitPrice":{"amount":"280.00000000","currency":"RUB"},"grossAmount":null,"commission":{"amount":"2.80000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-01-10","settlementDate":"2026-01-13","note":"Stage 3.4 smoke buy"}' >/dev/null

summary_response="$(api_get "/api/v1/portfolios/${portfolio_id}/summary")"
transactions_response="$(api_get "/api/v1/portfolios/${portfolio_id}/transactions")"

python3 - "$summary_response" "$transactions_response" <<'PY'
import json
import sys

summary = json.loads(sys.argv[1])["data"]
transactions = json.loads(sys.argv[2])["data"]["items"]

expected = {
    "totalValue": "9997.20000000",
    "cashValue": "7197.20000000",
    "stockValue": "2800.00000000",
    "investedCapital": "2802.80000000",
}

for field, amount in expected.items():
    actual = summary[field]["amount"]
    if actual != amount:
        raise SystemExit(f"{field} expected {amount}, got {actual}")

if len(transactions) != 2:
    raise SystemExit(f"expected 2 transactions, got {len(transactions)}")

if summary["calculation"]["methodologyVersion"] != "stage-03-02-local-cost-snapshot-v1":
    raise SystemExit("unexpected snapshot methodology version")
PY

assert_database_state "$portfolio_id"

echo "Stage 3.4 smoke verification passed:"
echo "- Next.js-compatible Go API base URL: ${API_BASE_URL}"
echo "- portfolio persisted in PostgreSQL"
echo "- immutable transactions appended"
echo "- snapshot rebuilt"
echo "- summary returned expected decimal-string values"
echo "- Docker PostgreSQL contains the created portfolio, transactions, and snapshots"
