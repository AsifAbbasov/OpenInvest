# Stage 3.16 — Repository Audit Fixes

| Field | Value |
| --- | --- |
| Status | Active / audit fixes |
| Owner | Principal Architect |
| Audit Target SHA | `74eebe9ec8231764f21ce384c4690d073d0273da` |
| Audit Verdict | `REQUEST CHANGES` |
| Branch | `stage-03-16-audit-fixes` |

## Scope

This stage fixes the blocking findings from the mandatory Stage 3.16 full repository audit before any
new implementation stage can begin.

## Implemented fixes awaiting independent review

- Fail closed when `OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET` is absent or shorter than 32 bytes;
  a static development secret is available only through the explicit development constructor or an
  explicitly named `development`/`local` runtime environment.
- Bind import append approvals to a server-signed review token covering subject, portfolio, source
  label, CSV file hash, row identities, and token signature integrity.
- Prevent stale Web import reviews and appends from committing after a changed CSV payload or source
  label invalidates their generation.
- Stop publishing placeholder nominal, XIRR, and real-return figures as ordinary portfolio outputs.
- Preserve browser idempotency keys across ambiguous retry attempts for serialized financial write
  intents, while rotating the key when the intent changes.
- Run frontend tests and PostgreSQL-backed Go integration tests in CI.
- Make portfolio and transaction continuation pages reachable in the Web UI using opaque cursors.
- Replace client-controlled offsets with signed opaque keyset cursors bound to subject, endpoint,
  portfolio/filter scope, and the deterministic page anchor.
- Store transaction provenance in a migration and allow same-value manual entries while rejecting
  duplicate imports and preserving the shared append lock.
- Add the durable 200-path audit coverage manifest, linked to the report and governance registers.
- Fail closed for an absent database URL outside explicitly named local/development environments,
  enforce OpenAPI token/hash bounds exactly, and rehearse every migration rollback and full reapply in CI.
- Preserve import append idempotency across ledger changes by leaving mutable duplicate classification
  to the locked store, use the internal transaction-entry ID for SQL keyset anchors, apply the
  provenance migration in the smoke workflow, and invalidate stale Web import work on session or
  portfolio changes.
- Sign public asset continuations with query/type-bound HMAC keysets, and synchronously invalidate
  stale Web import work before passive effects run after a session or portfolio change.
- Rehearse full migration rollback/reapply through null-delimited CI traversal and assert provenance
  columns, default, constraint, index, and invalid imported-row rejection in disposable PostgreSQL.
- Reject explicitly supplied empty or whitespace private-list cursors, and reject undeclared fields in
  import review/append payloads, including nested import decisions, to match the OpenAPI cursor and
  `additionalProperties: false` contract.

## Still requiring disposition

- Registration exists, but account deletion, anonymization, backup destruction, and retention
  execution remain non-production blockers until implemented or explicitly accepted by the human owner
  with expiry and compensating controls.
- Dependency advisories must be remediated or explicitly accepted in a separately reviewed dependency
  update.
- Snapshot rebuild performance and DDD/SOLID boundary pressure remain follow-up audit findings before
  any financial algorithm, market data, or worker stage.

## Prohibitions

This fix stage does not authorize WAC, XIRR, real return, inflation, purchasing power, dividends,
coupons, market data, providers, workers, tax, mobile, AI, or production rollout.
