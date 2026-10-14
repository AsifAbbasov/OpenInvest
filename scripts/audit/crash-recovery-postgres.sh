#!/usr/bin/env bash
# Independent crash-recovery probe. Disposable GitHub Actions postgres:18 only.
set -euo pipefail
: "${PG_CONTAINER:?required isolated container ID}"
: "${OPENINVEST_DATABASE_TEST_URL:?required isolated postgres URL}"
if [[ "$PG_CONTAINER" != *postgres* && "$PG_CONTAINER" != "pg-chaos-"* ]]; then echo 'ISOLATION_GUARD: unexpected container name'; exit 90; fi
export PGCONNECT_TIMEOUT=3
sql(){ psql "$OPENINVEST_DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -Atc "$1"; }
echo "PG_CRASH_TEST_BEGIN=$(date -u +%FT%TZ)"
marker="crash_audit_$(date +%s)_$$"
sql "CREATE TABLE IF NOT EXISTS public.independent_chaos_markers(id TEXT PRIMARY KEY, value BIGINT NOT NULL);"
sql "INSERT INTO public.independent_chaos_markers VALUES ('$marker',101)"
test "$(sql "SELECT value FROM public.independent_chaos_markers WHERE id='$marker'")" = 101
# Crash PostgreSQL postmaster without graceful shutdown. The orchestrator
# must NOT restart the container automatically; startup is explicit.
docker exec -u postgres "$PG_CONTAINER" sh -c 'kill -9 "$(head -1 /var/lib/postgresql/18/docker/postmaster.pid)"' || true
for n in $(seq 1 25); do
 if ! pg_isready -h 127.0.0.1 -p 5432 -U openinvest -d openinvest >/dev/null 2>&1; then break; fi
 sleep .2
done
if pg_isready -h 127.0.0.1 -p 5432 -U openinvest -d openinvest >/dev/null 2>&1; then echo 'FAIL_POSTMASTER_NOT_CRASHED'; exit 91; fi
echo 'POSTMASTER_SIGKILL_CONFIRMED=YES'
docker restart "$PG_CONTAINER" >/dev/null
ready=0
for n in $(seq 1 100); do
 if pg_isready -h 127.0.0.1 -p 5432 -U openinvest -d openinvest >/dev/null 2>&1; then ready=1; break; fi
 sleep .3
done
test "$ready" -eq 1 || { echo 'POSTGRES_RECOVERY_TIMEOUT'; exit 92; }
test "$(sql "SELECT value FROM public.independent_chaos_markers WHERE id='$marker'")" = 101 || { echo 'DURABLE_COMMITTED_MARKER_LOST'; exit 93; }
test "$(sql 'SELECT count(*) FROM identity.registration_intents WHERE generation <> (SELECT j.generation FROM identity.email_delivery_jobs j WHERE j.intent_id=identity.registration_intents.id) AND intent_state=\x27pending\x27 AND expires_at>statement_timestamp()')" = 0 || { echo 'CROSS_GENERATION_MISMATCH_AFTER_RESTART'; exit 94; }
sql "DELETE FROM public.independent_chaos_markers WHERE id='$marker'"
echo 'POSTGRES_SIGKILL_RESTART_COMMITTED_DURABILITY=PASS'
