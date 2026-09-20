# OI-NEW-09 — Graceful shutdown remediation

## Status

OI-NEW-09 remediation candidate implemented.

Verification completed.

Awaiting independent review / merge.

This document does not declare OI-NEW-09 CLOSED.

## Finding

- Finding: `OI-NEW-09 — Graceful shutdown отсутствует`
- Severity: `P3`
- Target: `develop`
- Baseline SHA: `cf58624d4b58ed54e4eb4867570f95091752cf56`
- Branch: `fix/oi-new-09-graceful-shutdown`

## Root cause

The API composition root constructed the Fiber application and immediately called
`Listen(":8080")`.

The process had no explicit SIGINT/SIGTERM lifecycle, no bounded graceful-shutdown
boundary, and no composition-root ownership that closed the PostgreSQL store after
server termination.

Initialization also used `log.Fatal` inside dependency construction, which was not
compatible with resource cleanup once owned resources had been opened.

## Remediation

The composition root now:

- owns the Fiber app and PostgreSQL close lifecycle explicitly;
- returns initialization errors instead of terminating inside construction;
- closes PostgreSQL when later initialization fails;
- handles SIGINT and SIGTERM with `signal.NotifyContext`;
- performs Fiber `ShutdownWithContext`;
- bounds graceful shutdown to 10 seconds;
- drains HTTP work before releasing owned runtime resources;
- propagates startup, serving, shutdown, and cleanup errors.

Startup integrity, `/health`, `/ready`, authentication, OI-NEW-05 HTTP timeouts,
and OI-NEW-06 trusted-proxy behavior remain unchanged.

No migrations, schema, financial semantics, auth semantics, OpenAPI, or frontend
behavior were changed.

## Verification

Focused lifecycle verification:

- `go test ./cmd/api -run 'TestOINew09' -count=1` — PASS
- `go test -race ./cmd/api -run 'TestOINew09' -count=20` — PASS
- `go test ./cmd/api -count=1` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
- `pnpm run verify` — PASS
- `pnpm audit` — PASS
- `pip-audit==2.10.1` — PASS
- `govulncheck@v1.7.0` with `GOTOOLCHAIN=go1.25.14` — PASS

Local `go test -race ./...` fails in pre-existing `internal/httpapi`
authentication tests due to request timeouts.

The same failure was reproduced on an untouched detached `origin/develop`
worktree at baseline SHA `cf58624d4b58ed54e4eb4867570f95091752cf56`.

Focused OI-NEW-09 race tests pass repeatedly.

The workstation-default Go toolchain is Go 1.26.2. Under that toolchain,
govulncheck reports standard-library vulnerabilities on both baseline and
candidate. Under repository/CI Go 1.25.14, govulncheck passes on both.

Required GitHub PR CI remains authoritative.

## Runtime contract

Runtime lifecycle:

process start -> dependency initialization -> startup integrity validation ->
HTTP serving -> SIGINT/SIGTERM -> bounded graceful shutdown ->
in-flight request drain -> PostgreSQL cleanup -> clean return or propagated error.

Signals handled:

- SIGINT
- SIGTERM

Shutdown timeout:

- 10 seconds

## Self-review

Reviewed for:

- goroutine leakage;
- races;
- shutdown deadlock;
- double/nil close;
- shutdown-before-start;
- signal cleanup;
- lifecycle error propagation;
- test flakiness;
- P2 regression;
- scope creep;
- resource ownership.

An intermediate `os.Exit(realMain())` implementation was removed during
self-review before publication.

## Candidate state

`OI_NEW_09_STATUS=AWAITING_INDEPENDENT_REVIEW`
