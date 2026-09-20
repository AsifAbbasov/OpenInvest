# OI-NEW-09 — Graceful shutdown remediation

## Status

OI-NEW-09 remediation candidate implemented.

Independent review identified follow-up blocker:

`OI-NEW-09-R1 — shutdown timeout does not guarantee bounded process termination`

The R1 remediation is now implemented and locally verified.

Awaiting fresh GitHub CI and independent re-review.

This document does not declare OI-NEW-09 CLOSED.

## Finding

- Finding: `OI-NEW-09 — Graceful shutdown отсутствует`
- Original severity: `P3`
- Independent-review blocker: `OI-NEW-09-R1`
- R1 reviewer severity: `P2`
- Target: `develop`
- Baseline SHA: `cf58624d4b58ed54e4eb4867570f95091752cf56`
- Branch: `fix/oi-new-09-graceful-shutdown`
- Pull request: `#193`

## Original root cause

The API composition root originally constructed the Fiber application and
immediately called `Listen(":8080")`.

The process had no explicit SIGINT/SIGTERM lifecycle, no bounded graceful
shutdown boundary, and no composition-root ownership that closed PostgreSQL
resources after server termination.

Initialization also used process termination from dependency construction,
which was incompatible with reliable resource ownership and cleanup.

## Independent review remediation — OI-NEW-09-R1

The first OI-NEW-09 candidate bounded only the wait performed by
`Fiber.ShutdownWithContext`.

That was insufficient for the full application lifecycle.

Fiber v3 delegates shutdown to fasthttp. When the shutdown deadline expires,
fasthttp may return `context.DeadlineExceeded` while a busy handler worker is
still executing.

A request can therefore still own application/database work after
`ShutdownWithContext` has returned.

OpenInvest handlers also propagate `fiber.Ctx.Context()` into service and
PostgreSQL calls. Without an explicit context bridge, that context is not tied
to the process shutdown lifecycle.

The consequence was a possible sequence:

SIGTERM ->
graceful timeout ->
HTTP shutdown returns ->
runtime cleanup begins ->
`sql.DB.Close()` waits for already-started database work ->
process shutdown becomes unbounded.

## Final lifecycle design

The remediation introduces an explicit `RequestLifecycle` owned by the API
composition root.

The lifecycle middleware:

- is installed before application routes;
- tracks admitted request work;
- preserves an already configured request context;
- supplies a cancellable standard-library context to downstream handlers;
- prevents new tracked work from being admitted after forced cancellation;
- exposes bounded waiting through `WaitContext`;
- exposes request ownership state through `Idle`.

### Graceful phase

On SIGINT or SIGTERM:

1. Fiber graceful shutdown begins.
2. Existing request contexts remain alive during the graceful phase.
3. The graceful phase has a 10-second timeout.
4. Requests are allowed to finish normally during that interval.
5. Runtime resources are closed only after tracked request work is quiescent.

Request contexts are deliberately not canceled merely because the process
received SIGTERM.

### Forced phase

If graceful shutdown returns an error or reaches its deadline:

1. request admission is sealed;
2. tracked request contexts are canceled;
3. the listener is defensively closed;
4. tracked request work receives a bounded 2-second forced-drain window;
5. the serving goroutine is joined within the forced lifecycle boundary.

Cooperative PostgreSQL work receives cancellation through the propagated
request context and can terminate before runtime resources are closed.

### Defensive ownership guard

The production timeout contract relies on request-owned operations honoring the
propagated cancellation context.

OpenInvest PostgreSQL request paths use cancellation-aware operations such as
`QueryContext`, `ExecContext`, and `BeginTx`, so the forced phase can unwind
database work before runtime cleanup.

Arbitrary Go or third-party code that ignores context cancellation indefinitely
cannot be forcibly terminated by the Go runtime while simultaneously
guaranteeing clean in-process resource cleanup. This remediation does not claim
otherwise.

As a fail-closed ownership invariant, `applicationRuntime.Close()` refuses to
enter the underlying resource closer while tracked request work still owns
runtime resources and returns `errRuntimeResourcesStillInUse`.

That condition is an ownership/lifecycle error, not a successful graceful
shutdown.

The production-equivalent timeout path verified by this remediation is:

forced request cancellation ->
cancellation-aware request/resource operation exits ->
request lifecycle reaches idle ->
HTTP serve goroutine exits ->
underlying Store.Close runs exactly once ->
runApplication returns with the shutdown error still discoverable.

### Unexpected post-start Listener termination

A Listener failure after the server has entered its accept loop follows the
same ownership invariant.

The listener is explicitly closed, the request lifecycle is force-canceled, and
active request work is given the bounded forced-drain window before runtime
resources may be closed.

This prevents an unexpected serving failure from leaving the listener owned by
the abandoned serve path or releasing PostgreSQL/runtime resources while
request-owned work is still active.

## Runtime resource ownership invariant

The required invariant is:

`runtime resources must not be closed while tracked request work still owns them`

`applicationRuntime.Close()` therefore:

1. force-cancels the request lifecycle;
2. checks `RequestLifecycle.Idle()`;
3. returns `errRuntimeResourcesStillInUse` while ownership remains active;
4. calls the underlying runtime closer only after request ownership reaches
   zero.

The underlying resource closer is not started speculatively in another
goroutine.

Normal cleanup therefore remains synchronous and explicit.

## Error propagation

Lifecycle errors remain discoverable through `errors.Is`.

The implementation preserves joined errors for combinations such as:

- HTTP shutdown failure + serving failure;
- HTTP shutdown failure + forced-drain timeout;
- serving failure + active request ownership;
- serving/shutdown failure + runtime cleanup failure.

Startup failures continue to propagate without starting the HTTP lifecycle.

Runtime cleanup is performed exactly once on safe cleanup paths.

## Verification

Final focused lifecycle verification:

- direct active-resource ownership guard, `count=20` — PASS
- forced-cancellation request/resource drain regression, `count=20` — PASS
- post-start Listener failure ordering regression, `count=20` — PASS
- `go test ./cmd/api -run 'TestOINew09' -count=1` — PASS
- `go test -race ./cmd/api -run 'TestOINew09' -count=20` — PASS
- `go test -race ./internal/httpapi -run 'TestRequestLifecycle' -count=20` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
- `go run ./cmd/validate-openapi` — PASS
- `pnpm run verify` — PASS
- `git diff --check` — PASS

`pnpm run verify` includes:

- Go repository tests;
- Python dependency sync and tests;
- frontend typecheck;
- frontend tests;
- production frontend build;
- OpenAPI validation;
- migration validation;
- Docker Compose configuration validation.

Local full `go test -race ./...` has previously hit one-second timeout failures
in pre-existing `internal/httpapi` authentication tests.

The same timeout class was previously reproduced against the untouched
baseline, while all OI-NEW-09 focused race suites pass repeatedly.

Fresh GitHub CI for the new R1 head remains authoritative.

The previous 10/10-success CI run belonged to the pre-R1 reviewed candidate and
must not be treated as fresh evidence for this final remediation.

## Regression boundaries

The remediation does not intentionally change:

- PostgreSQL schema or migrations;
- financial calculations;
- ledger semantics;
- authentication semantics;
- OpenAPI contracts;
- frontend behavior;
- OI-NEW-03 behavior;
- OI-NEW-05 HTTP timeout behavior;
- OI-NEW-06 trusted-proxy behavior.

## Self-review

Reviewed for:

- bounded whole-application shutdown;
- request-context propagation;
- preservation of existing request contexts;
- graceful versus forced cancellation ordering;
- database-operation cancellation;
- resource ownership;
- listener ownership;
- serving-goroutine termination;
- post-start Listener failure;
- defensive active-request ownership guard;
- cleanup exactly once;
- startup failure;
- shutdown error propagation;
- joined-error discoverability;
- nil/double close;
- races;
- goroutine leakage;
- scope creep.

## Candidate state

`OI_NEW_09_STATUS=AWAITING_INDEPENDENT_REVIEW`
