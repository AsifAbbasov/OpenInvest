# Stage 3.31 — Authentication Operational Hardening

| Field | Value |
| --- | --- |
| Status | Closure record; when canonical on `develop`, Stage 3.31 is closed for P2-01/P2-14 |
| Owner | Principal Architect |
| Baseline | `develop` at `ae6497050692798795efb85678af64db97cc5f53` |
| Branch | `fix/stage-03-31-auth-operational-hardening` |
| Implementation PR | #65 |
| Implementation merge | `9bf4d1d31597918eacf0c3358bf6caa2aa9db897` |
| Reviewed exact head | `82557c55c0772a66707088b858ec9eafc2073119` |
| Exact-head CI | GitHub Actions #133 — SUCCESS |
| Independent final review | `APPROVED` |
| Human implementation merge authorization | 2026-08-23 |
| Closure PR | #66 |
| Trigger | Repository-audit P2-01 and P2-14 |
| Scope | Logout admission, rejected-auth audit-write amplification, bounded auth limiter lifecycle, OpenAPI 429 parity, regression coverage |
| Out of scope | P2-09/P2-13 idempotency replay/browser recovery, remaining P2/P3, Stage 3.25 privacy work |

## P2-01 — logout audit-write amplification

Invalid logout paths can persist `AUTH_LOGOUT_REJECTED` audit events, but `/api/v1/auth/logout`
previously bypassed the HTTP auth limiter because its OpenAPI operation did not expose 429.

Repeated unauthenticated logout requests could therefore create sustained audit writes and database
transactions without valid credentials or a valid session.

Logout now passes through auth admission before body decoding or auth-service work. The limiter has
both a per-key budget and a finite global downstream-attempt budget. Once exhausted, requests return
the existing 429 + `Retry-After` response and do not reach rejected-auth audit persistence. OpenAPI
now exposes `RateLimited` for logout.

HTTP admission is the earliest boundary that can stop unauthenticated traffic before session lookup,
audit persistence, or other downstream work. Removing audit events was rejected because it would lose
security evidence; PostgreSQL-only limiting was rejected because it would already consume database
work; Redis was rejected as unnecessary MVP infrastructure expansion.

Admission remains process-local. A future multi-replica deployment may require a distributed limiter.
Global fail-closed exhaustion can temporarily reject legitimate auth traffic; this is intentional
overload behavior and `Retry-After` remains explicit.

## P2-14 — unbounded auth limiter key map

The previous limiter stored `map[path|IP][]time.Time` and pruned only the current key. Distinct keys
that were never revisited remained indefinitely, leaving key cardinality unbounded.

The limiter now has three finite dimensions: per-key attempts, global downstream attempts per window,
and maximum active key buckets. Expired current-key timestamps are pruned on access. Expired buckets
are swept periodically and whenever a new key encounters capacity. New keys fail closed if active
capacity remains full. The global slice cannot grow past its configured budget, and each key bucket
cannot grow past its per-key budget.

Whole-map timer clears and active-bucket eviction were rejected because they discard live protection;
TTL without cardinality bounds was rejected because it still permits memory spikes; Redis is not
required for the current single-process closure criterion.

## Regression evidence

- existing auth 429 response still includes `Retry-After`;
- first invalid logout can record one rejected audit event;
- repeated invalid logout is blocked before a second audit write;
- unique active-key cardinality cannot exceed configured capacity;
- new keys fail closed while active capacity is full;
- expired key buckets are reclaimed;
- a global budget caps downstream auth work across unique keys;
- the global budget becomes available again after expiry;
- logout OpenAPI advertises 429 `RateLimited`.

## Scope boundary

Stage 3.31 closes only P2-01/P2-14 after exact-head CI, independent final review, explicit human merge
authorization, implementation squash merge, and separate closure governance.

P2-09/P2-13 are intentionally deferred to the next idempotency stage because the normative contract
requires an identical replay to return the original status and body. P2-10/P2-11/P2-12/P2-16/P2-17
and all P3 findings remain separate. Stage 3.25 privacy evidence planning remains separate.


## Verification and implementation merge evidence

- Local targeted HTTP/auth/PostgreSQL tests passed.
- `go vet ./...`, migration validation, OpenAPI validation, and full `go test ./...` passed locally.
- Implementation PR #65 was squash-merged into `develop` at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`.
- Final independently reviewed implementation head: `82557c55c0772a66707088b858ec9eafc2073119`.
- Exact-head GitHub Actions CI #133 completed `SUCCESS`; all six workflow jobs passed.
- Independent final implementation review returned `APPROVED`.
- Explicit human authorization was received before the implementation squash merge.

## Canonical closure statement

Implementation prerequisites are satisfied, but this report is not canonical closure until PR #66
passes its own exact-head CI, independent closure-governance review, fresh explicit human merge
authorization, and squash merge into `develop`.

When that closure record is canonical, Stage 3.31 is closed only for P2-01/P2-14. The original audit
backlog then contains 7 P2 and 10 P3 findings. P2-09/P2-13 remain separate for the next idempotency
stage. Stage 3.25 privacy Security Review evidence planning remains separate and is not superseded.
