#!/usr/bin/env bash
set -euo pipefail

suite="${1:-}"
minutes="${OPENINVEST_CAMPAIGN_MINUTES:-}"
maximum_minutes="${OPENINVEST_CAMPAIGN_MAX_MINUTES:-1440}"
mode="${OPENINVEST_CAMPAIGN_MODE:-collect}"
campaign_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
log_dir="${OPENINVEST_CAMPAIGN_LOG_DIR:-$campaign_root/security-campaign-logs}"
mkdir -p "$log_dir"
log_file="$log_dir/${suite}.log"

if [[ ! "$suite" =~ ^(auth|provider|financial|postgres|frontend|breaker)$ ]]; then
  echo "usage: $0 {auth|provider|financial|postgres|frontend|breaker}" >&2
  exit 64
fi
if [[ ! "$minutes" =~ ^[1-9][0-9]*$ ]]; then
  echo "OPENINVEST_CAMPAIGN_MINUTES must be a positive integer" >&2
  exit 64
fi
if [[ ! "$maximum_minutes" =~ ^[1-9][0-9]*$ ]] || (( minutes > maximum_minutes )); then
  echo "OPENINVEST_CAMPAIGN_MINUTES must not exceed $maximum_minutes" >&2
  exit 64
fi
if [[ "$mode" != "collect" && "$mode" != "fail-fast" ]]; then
  echo "OPENINVEST_CAMPAIGN_MODE must be collect or fail-fast" >&2
  exit 64
fi

deadline=$(( $(date +%s) + minutes * 60 ))
iteration=0
failures=0

snapshot() {
  {
    echo "RESOURCE_SNAPSHOT suite=$suite at=$(date -u +%FT%TZ)"
    free -m 2>/dev/null || true
    df -h / 2>/dev/null || true
    ps -eo pid,rss,vsz,etime,cmd --sort=-rss 2>/dev/null | head -n 25 || true
    if [[ -n "${OPENINVEST_DATABASE_TEST_URL:-}" ]]; then
      psql "$OPENINVEST_DATABASE_TEST_URL" -At -v ON_ERROR_STOP=1 -c         "SELECT 'pg_activity='||count(*) FROM pg_stat_activity WHERE datname=current_database();" 2>/dev/null || true
    fi
  } >>"$log_file"
}

run_suite() {
  case "$suite" in
    auth)
      (cd "$campaign_root/backend-go" && go test -race ./internal/auth ./internal/httpapi -count=1)
      ;;
    provider)
      (cd "$campaign_root/backend-go" && go test -race ./internal/provider/... -count=1)
      ;;
    financial)
      (cd "$campaign_root/backend-go" && go test -race ./internal/decimal ./internal/importer ./internal/position ./internal/verticalslice -count=1)
      ;;
    postgres)
      : "${OPENINVEST_DATABASE_TEST_URL:?postgres suite requires OPENINVEST_DATABASE_TEST_URL}"
      (cd "$campaign_root/backend-go" && go test -race ./internal/postgres -count=1)
      ;;
    frontend)
      (cd "$campaign_root/frontend-next" && pnpm test && pnpm run typecheck && pnpm run build)
      ;;
    breaker)
      (
        cd "$campaign_root/backend-go"
        export OPENINVEST_SECURITY_BREAKER_TESTS=1
        go test -race ./internal/httpapi -run='^TestSecurityBreakerAuthRate' -count=1
        go test -race ./internal/auth -run='^(TestSecurityBreakerJWT|TestAuthTheft)' -count=1
        go test -race ./internal/provider/tinvest -run='^TestProviderResilience' -count=1
        if [[ -n "${OPENINVEST_DATABASE_TEST_URL:-}" ]]; then
          go test -race ./internal/postgres -run='^TestAuthTheftPostgres' -count=1
        fi
      )
      ;;
  esac
}

: >"$log_file"
echo "CAMPAIGN_START suite=$suite mode=$mode minutes=$minutes at=$(date -u +%FT%TZ)" | tee -a "$log_file"
snapshot

while (( $(date +%s) < deadline )); do
  iteration=$((iteration + 1))
  echo "ITERATION_START suite=$suite iteration=$iteration at=$(date -u +%FT%TZ)" | tee -a "$log_file"

  set +e
  run_suite 2>&1 | tee -a "$log_file"
  status=${PIPESTATUS[0]}
  set -e

  if (( status != 0 )); then
    failures=$((failures + 1))
    echo "ITERATION_FAILURE suite=$suite iteration=$iteration status=$status failures=$failures" | tee -a "$log_file"
    snapshot
    if [[ "$mode" == "fail-fast" ]]; then
      exit "$status"
    fi
  else
    echo "ITERATION_SUCCESS suite=$suite iteration=$iteration" | tee -a "$log_file"
  fi

  if (( iteration % 10 == 0 )); then
    snapshot
  fi
done

snapshot
echo "CAMPAIGN_END suite=$suite iterations=$iteration failures=$failures at=$(date -u +%FT%TZ)" | tee -a "$log_file"
test "$failures" -eq 0
