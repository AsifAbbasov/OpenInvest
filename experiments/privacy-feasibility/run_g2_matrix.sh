#!/usr/bin/env bash
# Disposable localhost PostgreSQL only; no existing database or credential is consumed.
set -euo pipefail
mode="${1:-normal}"
if [[ "$mode" != normal && "$mode" != race ]]; then echo 'usage: run_g2_matrix.sh normal|race' >&2; exit 2; fi
command -v docker >/dev/null
command -v go >/dev/null
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
container="oi-g2-gates-${GITHUB_RUN_ID:-local}-$$"
cleanup() { docker rm -fv "$container" >/dev/null 2>&1 || true; }
trap cleanup EXIT INT TERM
# Same major version as repository CI; parameters apply ONLY to this disposable container.
docker run --detach --rm --name "$container" --memory=256m --cpus=1 --pids-limit=128 \
  --publish 127.0.0.1::5432 \
  --env POSTGRES_USER=fixture_owner --env POSTGRES_PASSWORD=synthetic-local-only \
  --env POSTGRES_DB=postgres postgres:18 \
  -c max_prepared_transactions=8 -c max_connections=24 -c shared_buffers=32MB >/dev/null
docker inspect "$container" --format 'G2_CONTAINER_IMAGE_ID={{.Image}} MEMORY_LIMIT={{.HostConfig.Memory}} PID_LIMIT={{.HostConfig.PidsLimit}}'
ready=0
for attempt in $(seq 1 60); do
  if docker exec "$container" pg_isready -h 127.0.0.1 -U fixture_owner -d postgres >/dev/null 2>&1; then ready=1; break; fi
  sleep 0.2
done
if [[ "$ready" != 1 ]]; then docker logs "$container"; exit 1; fi
binding="$(docker port "$container" 5432/tcp)"
port="${binding##*:}"
if [[ ! "$binding" =~ ^127\.0\.0\.1:[0-9]+$ ]]; then echo 'non-loopback port refused' >&2; exit 1; fi
export OPENINVEST_DATABASE_TEST_URL="postgres://fixture_owner:synthetic-local-only@127.0.0.1:${port}/postgres?sslmode=disable"
export OPENINVEST_G2_TRANSACTION_MATRIX=1
export OPENINVEST_PRIVACY_FEASIBILITY_DISPOSABLE=1
printf 'G2_MATRIX_MODE=%s ISOLATION=DEDICATED_LOOPBACK_CONTAINER PREPARED_CAPACITY=8\n' "$mode"
cd "$root/backend-go"
if [[ "$mode" == race ]]; then
  go test -race -v ./experiments/privacyfeasibility -run '^TestG2TransactionGates$' -count=1 -timeout=120s
else
  go test -v ./experiments/privacyfeasibility -run '^TestG2TransactionGates$' -count=1 -timeout=120s
fi
