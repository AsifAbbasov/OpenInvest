# Final Production / Frontend / Infrastructure Hardening — P2 Remediation V1

## Scope

Authorized remediation for:
- `M12-P2-01 MULTI_REPLICA_PROVIDER_BUDGET_MULTIPLICATION`
- `M12-P2-02 DISCONNECT_CANCELLATION_DOES_NOT_BOUND_UPSTREAM_PROVIDER_WORK`

Frozen baseline:
- SHA: `7631e9fe62a8be2a527d12c608c8539771dd3277`
- tree: `b97fe7bd2e03b6a9251e7cb581450366b69c66ef`
- branch: `develop`

Audit PR #244 remains open/draft/unmerged and is not part of this remediation.

## Changed files

1. `.env.example`
2. `.github/workflows/ci.yml`
3. `backend-go/cmd/api/main.go`
4. `backend-go/cmd/api/main_test.go`
5. `backend-go/cmd/api/shared_budget_runtime.go`
6. `backend-go/cmd/api/tinvest_runtime.go`
7. `backend-go/cmd/api/tinvest_runtime_test.go`
8. `backend-go/go.mod`
9. `backend-go/go.sum`
10. `backend-go/internal/httpapi/api.go`
11. `backend-go/internal/httpapi/corporateactions.go`
12. `backend-go/internal/httpapi/corporateactions_abuse_control.go`
13. `backend-go/internal/httpapi/corporateactions_shared_budget_integration_test.go`
14. `backend-go/internal/httpapi/oi_if_001_import_admission_test.go`
15. `backend-go/internal/httpapi/replay_app.go`
16. `backend-go/internal/provider/tinvest/provider.go`
17. `backend-go/internal/provider/tinvest/shared_budget_cancellation_regression_test.go`
18. `backend-go/internal/sharedbudget/redis.go`
19. `docs/operations/ABUSE_PROTECTION_DEPLOYMENT.md`
20. `docs/audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_P2_REMEDIATION_V1.md`

No frontend source, business feature, migration, OpenAPI, caching, retry, polling, or Module #12 closure work is included.

## Shared budget architecture

`backend-go/internal/sharedbudget.RedisAuthority` uses pinned Redis client `github.com/redis/go-redis/v9 v9.22.0`.

The shared authority owns:
- Corporate Actions endpoint per-client budget: 12 / 60 seconds across replicas.
- Corporate Actions endpoint global budget: 48 / 60 seconds across replicas.
- T-Invest provider safety budget: 60 / 60 seconds across provider instances.
- Conservative shared provider `X-RateLimit-Remaining` observations.

Namespace:
`openinvest:<normalized-environment>:security-budget:v1`.

All security-budget keys have Redis TTLs. Client identities are SHA-256-derived before key construction.

## Atomicity semantics

Endpoint per-client + global admission is performed by one server-side Redis Lua script.

The script reads both counters, rejects before either increment when either limit is exhausted, then increments both counters atomically in the same Redis execution. Therefore a rejected global admission does not consume only a client token and a rejected client admission does not consume only a global token.

Provider admission uses a server-side Lua script that atomically checks the shared local safety ceiling and conservative shared remote allowance before increment/decrement.

No application-wall-clock timestamp comparison is used for admission; Redis key TTL owns window lifecycle.

```text
SHARED_BUDGET_BACKEND=REDIS
SHARED_BUDGET_ATOMICITY=SERVER_SIDE_REDIS_LUA
```

## Production fail-closed behavior

New runtime configuration:
`OPENINVEST_SHARED_BUDGET_REDIS_URL`.

Outside explicit development/local mode, enabled T-Invest Corporate Actions requires a configured shared budget authority. Redis configuration is parsed and pinged during runtime dependency construction; unavailable/misconfigured shared Redis causes startup/configuration failure instead of process-local fallback.

At request time, shared endpoint admission errors map to unavailable/fail-closed behavior and do not reach the provider.

Provider shared-budget admission errors fail closed as `ErrCorporateActionsProviderUnavailable` and perform zero upstream calls.

Operator acknowledgements remain configuration acknowledgements only:
- `OPENINVEST_DEPLOYMENT_GLOBAL_ABUSE_CONTROL`
- `OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER`

They are not treated as shared enforcement.

```text
PRODUCTION_FAIL_CLOSED=YES
SHARED_BACKEND_FAILURE=FAIL_CLOSED
```

## Shared endpoint regression campaign

Protected-CI evidence run: `37985206336`.

Real Redis service + real HTTP boundary:

```text
instances=1 total_http=65 http_2xx=48 http_429=17 http_5xx=0 total_provider_calls=48
instances=2 total_http=65 http_2xx=48 http_429=17 http_5xx=0 total_provider_calls=48
instances=4 total_http=65 http_2xx=48 http_429=17 http_5xx=0 total_provider_calls=48

SHARED_ENDPOINT_GLOBAL_LIMIT_1_REPLICA=48
SHARED_ENDPOINT_GLOBAL_LIMIT_2_REPLICAS=48
SHARED_ENDPOINT_GLOBAL_LIMIT_4_REPLICAS=48
GLOBAL_LIMIT_DOES_NOT_MULTIPLY=PASS
HTTP_5XX_FROM_BUDGET_RACE=0
```

Same canonical client was routed across four replicas:

```text
SHARED_PER_CLIENT_LIMIT_ACROSS_REPLICAS=12
PER_CLIENT_LIMIT_DOES_NOT_MULTIPLY=PASS
```

After exhausting the window, a newly created application replica using the same Redis namespace remained limited:

```text
PROCESS_RESTART_SHARED_BUDGET_RESET=NO
```

Closed/shared backend regression:

```text
M12_SHARED_BUDGET_BACKEND_UNAVAILABLE=FAIL_CLOSED
PROVIDER_CALLS=0
```

## Shared provider budget regression campaign

Real Redis service with 1, 2, and 4 independent Provider objects:

```text
instances=1 total_provider_calls=60 provider_budget_multiplication=NO
instances=2 total_provider_calls=60 provider_budget_multiplication=NO
instances=4 total_provider_calls=60 provider_budget_multiplication=NO

SHARED_PROVIDER_LIMIT_1_INSTANCE=60
SHARED_PROVIDER_LIMIT_2_INSTANCES=60
SHARED_PROVIDER_LIMIT_4_INSTANCES=60
PROVIDER_BUDGET_MULTIPLICATION=NO
PROCESS_RESTART_SHARED_BUDGET_RESET=NO
```

Shared provider backend outage:

```text
M12_SHARED_PROVIDER_BACKEND_UNAVAILABLE=FAIL_CLOSED
PROVIDER_CALLS=0
```

Provider quota response observation is shared conservatively:

```text
M12_SHARED_PROVIDER_REMOTE_ALLOWANCE=GLOBAL_CONSERVATIVE
```

## Cancellation/resource design

The provider retains local concurrency occupancy when caller cancellation returns before the underlying upstream transport has reached a safe terminal state.

A cancelled call returns promptly to the caller where possible, but the provider slot is quarantined for a bounded safety hold of `requestTimeout + 500ms` when abandoned upstream work may still exist. This prevents admitting replacement work while the old upstream operation can still consume resources.

The finite provider HTTP timeout remains 5 seconds.

## Cancellation/resource regression campaign

Real socket regression:
- rounds: 5
- concurrent per round: 4
- total cancellations: 20
- configured provider concurrency: 4

Protected-CI measurements:

```text
TOTAL_CANCELLATIONS=20
UPSTREAM_CONTEXT_CANCELLATIONS=0
UPSTREAM_ACTIVE_AFTER_750MS=4
PROVIDER_MAX_ACTIVE_UPSTREAM=4
LOCAL_CONCURRENCY_IN_USE_AFTER_CALLER_CANCEL=4

GOROUTINES_BASELINE=3
GOROUTINES_PEAK=27
GOROUTINES_RECOVERY=6

FD_BASELINE=10
FD_PEAK=18
FD_RECOVERY=12

SUBSEQUENT_LEGITIMATE_REQUEST=PASS
```

Immediate upstream context cancellation is therefore not claimed. The security invariant is instead provided by conservative slot occupancy: abandoned upstream work remains bounded by the configured concurrency capacity.

```text
CANCELLED_OR_DISCONNECTED_CALLS_CANNOT_CREATE_UNBOUNDED_UPSTREAM_CONCURRENCY=YES
UPSTREAM_ACTIVE_MUST_NOT_EXCEED_CONFIGURED_PROVIDER_CONCURRENCY_BOUND=YES
CANCELLATION_AMPLIFICATION=NO
```

Transport failure regressions:

```text
PARTIAL_RESPONSE_RECOVERY=PASS
CONNECTION_RESET_RECOVERY=PASS
UPSTREAM_TIMEOUT_RECOVERY=PASS
```

## Dependency and configuration changes

New direct dependency:
- `github.com/redis/go-redis/v9 v9.22.0`

New indirect module requirements:
- `github.com/cespare/xxhash/v2 v2.3.0`
- `go.uber.org/atomic v1.11.0`

Additional new go.sum-only dependency/test artifacts:
- `github.com/bsm/ginkgo/v2 v2.12.0`
- `github.com/bsm/gomega v1.27.10`
- `github.com/klauspost/cpuid/v2 v2.2.10`
- `github.com/zeebo/xxh3 v1.1.0`

`github.com/valyala/fasthttp v1.73.0` remains the same version and is promoted from indirect to direct module declaration.

Configuration:
- added `OPENINVEST_SHARED_BUDGET_REDIS_URL` documentation/example;
- Go test and race-test protected CI jobs now run a real Redis 8-alpine service;
- permanent shared-budget regressions run with `OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL`;
- no production Docker Compose topology is claimed from the CI Redis service.

## Protected CI evidence

Run `37985206336` on remediation head `c482bc0ed40449f1019026ce0f7ba70b25fb8c03`:

```text
REQUIRED_JOBS=10/10_SUCCESS
GOVULNCHECK=0_REACHABLE_VULNERABILITIES
PNPM_AUDIT=NO_KNOWN_VULNERABILITIES
PYTHON_AUDIT=NO_KNOWN_VULNERABILITIES
GO_RACE=PASS
DATA_RACE=NONE
```

The final report-only PR head must still pass protected CI before merge review.

## Remediation disposition

```text
M12_P2_01_REPOSITORY_SIDE=REMEDIATED
M12_P2_02_REPOSITORY_SIDE=REMEDIATED

REPLICA_COUNT_MUST_NOT_MULTIPLY_SECURITY_BUDGET=YES
PROCESS_RESTART_MUST_NOT_RESET_SHARED_BUDGET=YES
SHARED_BACKEND_FAILURE=FAIL_CLOSED
CANCELLATION_AMPLIFICATION=NO
PROVIDER_MAX_ACTIVE_UPSTREAM=4
CONFIGURED_MAX_CONCURRENCY=4
```

## Production-only residual boundaries

```text
REAL_EXTERNAL_PROVIDER_BUDGET_TEST_EXECUTED=NO
PRODUCTION_PROVIDER_ACCOUNT_ENFORCEMENT=NOT_VERIFIED
REAL_PRODUCTION_MULTI_HOST_TOPOLOGY=NOT_VERIFIED
REAL_PRODUCTION_EDGE_TLS_HSTS=NOT_VERIFIED
```

Repository-side Redis shared enforcement is not proof that the actual production deployment uses the required Redis authority, routes all replicas through the same namespace/backend, or that the real provider account enforces any additional account-wide quota.

Module #12 is not formally closed. The final repository-wide fuzz/property/security assault has not started.
