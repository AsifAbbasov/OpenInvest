# Deployment-global abuse control ownership

OpenInvest keeps process-local limiters as fail-fast safety controls. They are not deployment-global
when the API is horizontally replicated. Outside `development`/`local`, startup therefore requires:

```text
OPENINVEST_DEPLOYMENT_ABUSE_CONTROL=edge-v1
```

The value is an explicit operator contract: the API origin must only be reachable through an
operator-controlled edge/gateway that enforces the aggregate ceilings below. Direct public origin
reachability violates this contract. Proxy identity remains governed independently by
`OPENINVEST_TRUST_PROXY` and its validated allowlist.

## edge-v1 aggregate ceilings

The edge policy must aggregate across all API replicas for one deployment:

| Surface | Required deployment-global ceiling |
| --- | ---: |
| `/api/v1/auth/*` | <= 2000 requests/minute |
| `/api/v1/portfolios/*/imports/review` + `imports/append` | <= 60 requests/minute combined |
| `/api/v1/dividends/calculate` | <= 1200 requests/minute |
| `positions` + `cash-flow` + `returns` + `summary` | <= 120 expensive reads/minute |
| `/api/v1/corporate-actions/projection` while T-Invest is enabled | <= 1 request/minute |

The Corporate Actions edge ceiling is intentionally conservative: one request accepts at most
50 instrument IDs and the provider implementation issues at most one provider request per canonical
instrument. The existing provider-local 60/minute budget and four-concurrent fail-fast semaphore remain
per-process safety controls, while provider HTTP 429 / rate-limit headers remain authoritative remote
budget evidence.

The import edge ceiling is deliberately no greater than the application's fresh-command ceiling. Exact
idempotency replay remains durable and correct; an edge may rate-limit a replay for availability, but
must never rewrite or strip `Idempotency-Key`.

## Layered local controls

The application still enforces:
- auth per-IP emergency and normalized-credential buckets;
- Argon2 process capacity;
- import per-subject execution/fresh budgets and process capacity;
- dividend per-key/process fresh admission;
- expensive-read per-subject/process budgets and process capacity;
- T-Invest process-local request/concurrency bounds.

These controls protect one process and provide defense in depth. They must never be described as
deployment-global.

## Failure model

Staging/production startup fails closed unless `OPENINVEST_DEPLOYMENT_ABUSE_CONTROL=edge-v1` is
configured. Configuration does not make an unprotected origin safe: deployment verification must prove
that the origin is not directly reachable around the declared edge policy.
