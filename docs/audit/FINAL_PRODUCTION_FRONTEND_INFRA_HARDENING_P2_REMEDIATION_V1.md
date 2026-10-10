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

## Final non-cooperative upstream evidence gate

This evidence-only follow-up starts from independently reviewed PR head
`b3b872795342d29312714c40a3fc8fcf3f1abe6c`, tree
`7f959f844ae228c8884e97fba8208ec1b3ada16c`, whose protected CI run
`38063030700` passed ten of ten jobs. That previous run is not credited as CI
for the new evidence commit. No provider production or Redis implementation is
changed by this follow-up. Only the permanent test, its explicit CI evidence
invocation, and this report change.

The accepted original audit evidence is retained without rerunning vulnerable
code: 20 concurrent cancellations, local semaphore occupancy zero, 20 upstream
handlers still active after 750 milliseconds, and zero upstream context
cancellations. These are accepted pre-remediation audit measurements, not new
measurements from this follow-up.

`TestNoncooperativeUpstreamCannotRecycleCapacityOnCallerCancellation` uses a
real httptest HTTP server and the normal Go HTTP transport. Its blocked upstream
handler waits only for the explicit test release channel. It never exits on
request-context cancellation, connection closure or a self-terminating timer.
Remote active, maximum and total-started counters are independent of the client
transport counters. Local operation observation extends through response-body
closure, using the existing transparent transport instrumentation.

Four requests enter the remote fixture; all four callers are then cancelled and
must return context.Canceled within one shared one-second promptness deadline.
After an additional 500 milliseconds, the test verifies four occupied slots,
four local operations and four remote handlers. Four replacement attempts must
fail closed without reaching the remote server. The measurement must complete
at least one second before the provider's five-second timeout; scheduler delays
that violate that bound fail the test rather than count as evidence.

The test records its security counters before explicitly releasing handlers.
It then observes zero slots, zero local operations and zero remote active
handlers, switches the fixture to immediate success, and verifies a legitimate
request plus bounded goroutine and descriptor recovery. Deferred release exists
only as failure cleanup; successful measurements precede explicit release.
Remote total/max counters below describe the blocked challenge phase, before
the subsequent legitimate request in immediate-success mode.

The final challenger passed locally with the race detector:

```text
HOSTILE_NONCOOPERATIVE_UPSTREAM_TEST=PASS
FIRST_WAVE_REMOTE_ACTIVE=4
CALLERS_CANCELLED=4
CALLERS_RETURNED_PROMPTLY=YES
OBSERVATION_DELAY_MS=500
PROVIDER_SEMAPHORE_IN_USE_AFTER_CANCEL=4
LOCAL_PROVIDER_OPERATIONS_AFTER_CANCEL=4
SECOND_WAVE_ATTEMPTS=4
SECOND_WAVE_REACHED_REMOTE=0
REMOTE_TOTAL_STARTED=4
REMOTE_MAX_ACTIVE=4
SEMAPHORE_AFTER_RELEASE=0
LOCAL_PROVIDER_OPERATIONS_AFTER_RELEASE=0
SUBSEQUENT_LEGITIMATE_REQUEST=PASS
M12_P2_02_CALLER_CANCELLATION_AMPLIFICATION=REMEDIATED_REPOSITORY_SIDE
LOCAL_PROVIDER_TRANSPORT_CONCURRENCY_BOUNDED=YES
CALLER_CANCELLATION_CANNOT_RECYCLE_PROVIDER_CAPACITY_EARLY=YES
M12_P2_01_REPOSITORY_SIDE=REMEDIATED
M12_P2_01_ROLLING_WINDOW_SEMANTICS=PASS
REAL_EXTERNAL_PROVIDER_SERVER_SIDE_CANCELLATION=NOT_VERIFIED
REMOTE_PROVIDER_SERVER_SIDE_TERMINATION=NOT_VERIFIED
```

This proves caller cancellation cannot recycle local capacity before the
independently bounded operation reaches its local terminal condition. After
that timeout closes the local transport, arbitrary remote computation may
continue. Neither this challenger nor the shared request budget proves that
external computation terminates. Remote provider termination remains a residual
production/provider property.

The permanent challenger is invoked explicitly in the existing protected CI
Module 12 evidence step. Existing cooperative socket, hard-timeout, raw TCP,
local lifetime, error recovery and all rolling/shared-replica regressions remain
unchanged and are rerun. Exact final-head CI results are reported separately
once all ten protected jobs complete.

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
