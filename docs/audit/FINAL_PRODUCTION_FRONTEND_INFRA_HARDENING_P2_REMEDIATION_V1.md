# Module 12 P2 remediation — scoped verification

## Baseline and change control

Protected `develop` was independently read before this follow-up:

```text
DEVELOP_SHA=7631e9fe62a8be2a527d12c608c8539771dd3277
DEVELOP_TREE=b97fe7bd2e03b6a9251e7cb581450366b69c66ef
DEVELOP_PROTECTED=YES
STARTING_PR_245_HEAD=2179de251e40be638ad3954fab573d526086cebf
STARTING_PR_245_TREE=07dae48f409fd871dde5f087f0ccb3222ff2124e
PR_244_HEAD=29ad44cadfdd8f072b65da0235de540b34afb462
```

PR #244 remains open, draft and unmerged; it was not modified. PR #245 is not
merged. Module #12 closure and the final repository-wide assault are not started.
Previous run `37999113179` applies only to starting head `2179de...`, not this
follow-up. The new exact-head CI result is reported separately after commit.

## Recovered local inventory

The isolated worktree was detached at the exact starting head. No changes were
reset or discarded. Inventory before further edits:

- `provider.go`: detached caller-cancellation lifetime change.
- `shared_budget_cancellation_regression_test.go`: partial fixture adjustment.
- `sharedbudget/redis.go`: unfinished rolling-window implementation.
- `sharedbudget/rolling_window_test.go`: untracked rolling tests.
- The attempted new socket test file was absent because its creation failed.

The recovered changes were completed and verified as described below.

## Files changed in this follow-up

1. `.github/workflows/ci.yml`
2. `backend-go/internal/provider/tinvest/provider.go`
3. `backend-go/internal/provider/tinvest/shared_budget_cancellation_regression_test.go`
4. `backend-go/internal/provider/tinvest/detached_operation_socket_test.go`
5. `backend-go/internal/sharedbudget/redis.go`
6. `backend-go/internal/sharedbudget/rolling_window_test.go`
7. `docs/operations/ABUSE_PROTECTION_DEPLOYMENT.md`
8. `docs/audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_P2_REMEDIATION_V1.md`

The earlier Redis wiring, dependency, runtime activation and shared endpoint
changes remain in PR #245. No new dependency is added in this follow-up. Pinned
Redis client remains `github.com/redis/go-redis/v9 v9.22.0`.

## Rolling-window budget authority

Limits remain shared across replicas:

- Endpoint global: 48 in a rolling 60-second interval.
- Per canonical client: 12 in a rolling 60-second interval.
- Provider: 60 in a rolling 60-second interval.

Admission stores a bounded Redis sorted-set journal rather than a fixed-window
counter. Redis `TIME` supplies microsecond scores. Window durations are rounded
up to microseconds, key TTLs rounded up to milliseconds. Each attempt receives a
128-bit random member identity to distinguish simultaneous admissions.

One endpoint Lua execution removes expired entries, checks both global and
client journals, and appends to both only if both limits permit admission.
Rejection does not partially consume an admission. Provider admission atomically
checks its rolling journal and shared conservative remote allowance, then
appends and decrements the remote allowance when present.

Expiration uses `(now - window, now]` semantics. Accepted entries refresh the
journal's bounded TTL to one window. Rejected attempts do not prolong it. A
journal never grows beyond its configured admission limit.

Old string counters lack admission timestamps. Conversion conservatively
reserves an entire allowance at Redis's current time for one complete rolling
window, denying the triggering request. It does not merely wait for the old
fixed-window TTL, which could expire too early. This can temporarily restrict
availability during conversion; it never grants a fresh allowance early.
Unrecognized key types deny admission.

The supported authority is the existing single Redis backend; Redis Cluster
routing and real production topology are not verified. Redis-server time avoids
replica clock comparisons; production Redis clock integrity remains an operator
responsibility.

Production requires shared storage and fails closed on admission errors. There
is no production fallback to local accounting. The shared conservative provider
quota observation remains intact. Process recreation does not clear Redis state.

## Rolling verification

Real Redis tests use a two-second window and the actual limits 48, 12 and 60.
An initial admission establishes the old fixed-window boundary; later admissions
fill the remainder. Just past the original boundary, only the expired initial
admission can be replaced, rather than issuing a new full allowance. Scheduling
that invalidates the measurement interval fails the test instead of passing it.

The permanent tests also verify concurrent admission, bounded TTL/journal size,
atomic client/global accounting and conservative legacy-state conversion.

```text
BOUNDARY_BURST_ENDPOINT_GLOBAL=BLOCKED_TO_ROLLING_LIMIT
BOUNDARY_BURST_PER_CLIENT=BLOCKED_TO_ROLLING_LIMIT
BOUNDARY_BURST_PROVIDER=BLOCKED_TO_ROLLING_LIMIT
ROLLING_WINDOW_GLOBAL_MAX_OBSERVED=48
ROLLING_WINDOW_PER_CLIENT_MAX_OBSERVED=12
ROLLING_WINDOW_PROVIDER_MAX_OBSERVED=60
```

The existing real-HTTP 1/2/4-replica, same-client, provider-instance, restart and
backend-failure regressions passed again locally with real Redis and the race
detector on the follow-up code. Exact-head protected CI remains the final gate.
They are not inferred from the old CI run.

```text
M12_P2_01_REPLICA_MULTIPLICATION=REMEDIATED
M12_P2_01_WINDOW_SEMANTICS=ROLLING_WINDOW_RESTORED
HTTP_GLOBAL_1_REPLICA=48
HTTP_GLOBAL_2_REPLICAS=48
HTTP_GLOBAL_4_REPLICAS=48
HTTP_PER_CLIENT_ACROSS_REPLICAS=12
PROVIDER_1_INSTANCE=60
PROVIDER_2_INSTANCES=60
PROVIDER_4_INSTANCES=60
PROCESS_RESTART_RESETS_SHARED_BUDGET=NO
SHARED_BACKEND_UNAVAILABLE=FAIL_CLOSED
SHARED_BACKEND_UNAVAILABLE_PROVIDER_CALLS=0
```

## Provider operation lifetime

An admitted operation owns its semaphore slot in a bounded worker until HTTP
request execution, body processing and body closure finish. The completion
channel is buffered so an abandoned caller cannot prevent worker recovery.

The operation context is now:

```go
context.WithTimeout(context.WithoutCancel(callerContext), requestTimeout)
```

Caller cancellation returns promptly without cancelling this operation context.
The provider operation retains its independent five-second deadline. Capacity
is released when the local operation actually returns, not by a quarantine timer
or by the caller's return. Shared quota observation uses this bounded context.
At most four provider operation workers can own outbound work.

A custom RoundTripper that violates context/deadline semantics remains occupied
rather than causing capacity to be released on elapsed time. Runtime transports
must honor the normal `net/http` cancellation/deadline contract.

## Scoped cancellation and socket verification

The custom local lifetime regression remains: ten rounds of four cancellations,
with the first round keeping a controlled local operation alive past six seconds.
Its explicit terminal signal releases ownership. The fixture was adjusted so
release no longer depends on caller cancellation reaching the operation context.

A new post-fix real-socket regression performs ten rounds of four cancellations.
Its upstream cooperates with transport cancellation and supports explicit test
release. It counts locally owned HTTP operations through RoundTrip errors or
response-body Close. Replacement attempts must fail while the four original
operations remain owned. Each subsequent round starts only after actual local
terminal conditions are observed. Goroutine and file-descriptor recovery are
measured for this real-socket campaign.

A separate real-socket check observes that prompt caller cancellation leaves the
local operation occupied, and that its independent hard timeout subsequently
terminates the local transport and releases capacity. Server-side cleanup uses
an explicit release and does not depend on remote context cancellation.

The existing Fiber/raw `net.Dial` test verifies four downstream disconnects,
HTTP 503 while capacity is occupied, and HTTP 200 after local completion.
Downstream disconnect cancellation is not exposed by the observed Fiber path;
this is not claimed as upstream cancellation.

Partial response, connection reset and upstream timeout recovery tests passed.

Race-enabled local real-socket measurements (cooperative fixture only):

```text
TOTAL_CALLER_CANCELLATIONS=40
CANCELLATION_DRIVEN_REPLACEMENT_BEFORE_LOCAL_TERMINAL=0
MAX_LOCAL_PROVIDER_TRANSPORTS=4
MAX_SEMAPHORE_IN_USE=4
COOPERATIVE_FIXTURE_MAX_REMOTE_HANDLERS=4
LOCAL_PROVIDER_TRANSPORT_CONCURRENCY_BOUNDED=YES
M12_P2_02_LOCAL_TRANSPORT_RESOURCE_AMPLIFICATION=REMEDIATED_IN_TESTED_SCOPE
DETACHED_OPERATION_HARD_TIMEOUT=PASS
RAW_TCP_DISCONNECT_TEST=PASS
DOWNSTREAM_DISCONNECT_CONTEXT_PROPAGATION=NO
SUBSEQUENT_REQUEST_WHILE_PROVIDER_OCCUPIED=503
SUBSEQUENT_REQUEST_AFTER_TERMINAL=200
GOROUTINES_BASELINE=3
GOROUTINES_PEAK=27
GOROUTINES_RECOVERY=3
FD_BASELINE=9
FD_PEAK=17
FD_RECOVERY=9
```

## Explicitly unexecuted hostile reproduction

The requested pre-fix/post-fix reproduction using an upstream that ignores socket
cancellation indefinitely was not executed. The cooperative post-fix socket
fixture and controlled local RoundTripper are not substitutes for that evidence.
No pre-fix remote counts, hostile post-fix remote counts or reproduction outcome
are invented.

```text
PRE_FIX_CALLER_CANCELLATION_AMPLIFICATION=NOT_EXECUTED
PRE_FIX_FIRST_WAVE_REMOTE_ACTIVE=NOT_MEASURED
PRE_FIX_SECOND_WAVE_REACHED_REMOTE=NOT_MEASURED
PRE_FIX_REMOTE_MAX_ACTIVE=NOT_MEASURED
POST_FIX_SECOND_WAVE_REACHED_REMOTE=NOT_MEASURED_FOR_HOSTILE_FIXTURE
M12_P2_02_CALLER_CANCELLATION_AMPLIFICATION=DEFENSIVE_SCOPE_ONLY_PENDING_HOSTILE_VERIFICATION
REAL_EXTERNAL_PROVIDER_SERVER_SIDE_CANCELLATION=NOT_VERIFIED
REMOTE_PROVIDER_SERVER_SIDE_TERMINATION=NOT_VERIFIED
```

External computations may continue after local transport closure. Neither local
operation bounds nor shared request budgets prove arbitrary remote termination.

## Execution and dependency evidence

Initial local Redis runs could not connect because server and tests were run in
separate isolated execution environments. Those runs are failures, not proof of
limiter behavior. Redis and tests were then launched together in one execution.

A first timeout fixture incorrectly waited for the server's disconnect context
as proof of local completion, causing teardown to hang. It was repaired to
observe local transport completion and explicitly release its server fixture.
That failed run is not counted as successful evidence.

Final local commands include race-enabled sharedbudget/provider/HTTP tests with
real Redis, full Go tests/race/vet, and verbose govulncheck. Tests requiring real
PostgreSQL are authoritatively executed by protected CI, not credited from local
skips. CI retains all ten protected checks and its real Redis services. An
additional evidence step prints the rolling and detached socket results.

Known module advisory, without a reachable product call path:

```text
NON_REACHABLE_GOVULN_ID=GO-2026-5932
NON_REACHABLE_GOVULN_MODULE=golang.org/x/crypto@v0.57.0
AFFECTED_PACKAGE=golang.org/x/crypto/openpgp
FIXED_VERSION=N/A
```

No scanner suppression is introduced. The new exact-head scan must retain zero
reachable vulnerabilities; a clean reachable scan does not erase this advisory.

## Production boundaries and stop gate

```text
REAL_EXTERNAL_PROVIDER_BUDGET_TEST_EXECUTED=NO
PRODUCTION_PROVIDER_ACCOUNT_ENFORCEMENT=NOT_VERIFIED
REAL_PRODUCTION_MULTI_HOST_TOPOLOGY=NOT_VERIFIED
REAL_PRODUCTION_EDGE_TLS_HSTS=NOT_VERIFIED
REAL_EXTERNAL_PROVIDER_SERVER_SIDE_CANCELLATION=NOT_VERIFIED
REMOTE_PROVIDER_SERVER_SIDE_TERMINATION=NOT_VERIFIED
APPROVE_FOR_MERGE=NO
FORMAL_MODULE_12_CLOSED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO
```

Review the exact new PR head and its completed protected CI. This report does not
approve merge or declare Module #12 closed.
