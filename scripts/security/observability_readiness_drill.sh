#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_BASE_URL:?set OPENINVEST_BASE_URL, e.g. http://127.0.0.1:8080}"
expected_ready="${OPENINVEST_EXPECT_READY_CODE:-200}"

tmp_headers="$(mktemp)"
tmp_body="$(mktemp)"
trap 'rm -f "$tmp_headers" "$tmp_body"' EXIT

request() {
  local path="$1"
  curl \
    --silent \
    --show-error \
    --max-time 5 \
    --dump-header "$tmp_headers" \
    --output "$tmp_body" \
    --write-out '%{http_code}' \
    "$OPENINVEST_BASE_URL$path"
}

echo "[health] liveness must answer 200"
health_code="$(request /api/v1/health)"
if [[ "$health_code" != "200" ]]; then
  echo "FAIL: health returned HTTP $health_code" >&2
  cat "$tmp_body" >&2
  exit 1
fi

echo "[readiness] checking expected dependency state"
ready_code="$(request /api/v1/ready)"
if [[ "$ready_code" != "$expected_ready" ]]; then
  echo "FAIL: readiness returned HTTP $ready_code, expected $expected_ready" >&2
  cat "$tmp_body" >&2
  exit 1
fi

if [[ "$expected_ready" == "503" ]]; then
  if ! grep -q 'SERVICE_NOT_READY' "$tmp_body"; then
    echo "FAIL: degraded readiness did not expose stable SERVICE_NOT_READY classification" >&2
    cat "$tmp_body" >&2
    exit 1
  fi
  if grep -Eqi 'postgres|password|database_url|stack trace|panic|sqlstate' "$tmp_body"; then
    echo "FAIL: degraded readiness leaked infrastructure detail" >&2
    cat "$tmp_body" >&2
    exit 1
  fi
fi

echo "OBSERVABILITY_READINESS_DRILL=PASS health=200 ready=$ready_code"
