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

cleanup_controlled_api() {
  local pid="${API_PID:-}"
  local i=0

  if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
    kill -TERM "$pid" >/dev/null 2>&1 || true
    while kill -0 "$pid" >/dev/null 2>&1 && [[ "$i" -lt 40 ]]; do
      sleep 0.25
      i=$((i + 1))
    done
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill -KILL "$pid" >/dev/null 2>&1 || true
    fi
  fi

  if [[ -n "$pid" ]]; then
    wait "$pid" >/dev/null 2>&1 || true
  fi

  if [[ -n "${API_TMP_DIR:-}" && -d "$API_TMP_DIR" ]]; then
    rm -rf "$API_TMP_DIR"
  fi

  API_PID=""
  API_BIN=""
  API_TMP_DIR=""
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

  API_TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/openinvest-stage-03-04-api.XXXXXX")"
  API_BIN="$API_TMP_DIR/api"

  echo "Building disposable Go API binary."
  (
    cd "$ROOT_DIR/backend-go"
    go build -o "$API_BIN" ./cmd/api
  )

  echo "Starting Go API at ${API_BASE_URL}"
  (
    export OPENINVEST_ENV="${OPENINVEST_ENV:-development}"
    export OPENINVEST_DEV_AUTH_BYPASS="${OPENINVEST_DEV_AUTH_BYPASS:-true}"
    exec "$API_BIN"
  ) >"$API_LOG" 2>&1 &
  API_PID=$!

  trap cleanup_controlled_api EXIT
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
echo "Bootstrapping local infrastructure and the canonical PostgreSQL schema."
bash "$ROOT_DIR/scripts/bootstrap-local.sh"
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

if summary["calculation"]["methodologyVersion"] != "stage-03-71-position-cost-snapshot-v1":
    raise SystemExit("unexpected snapshot methodology version")
PY

assert_database_state "$portfolio_id"

cleanup_controlled_api
trap - EXIT

echo "Stage 3.4 smoke verification passed:"
echo "- Next.js-compatible Go API base URL: ${API_BASE_URL}"
echo "- portfolio persisted in PostgreSQL"
echo "- immutable transactions appended"
echo "- snapshot rebuilt"
echo "- summary returned expected decimal-string values"
echo "- Docker PostgreSQL contains the created portfolio, transactions, and snapshots"
