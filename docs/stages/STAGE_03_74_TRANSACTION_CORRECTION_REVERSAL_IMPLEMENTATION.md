# Stage 3.74 — Transaction Correction & Reversal / Ledger Repair UX

| Field | Value |
| --- | --- |
| Stage | 3.74 |
| Status | IMPLEMENTATION CANDIDATE — canonical only after reviewed protected merge |
| Canonical base | `develop@18fff1150ac8d1d3c4eafbf7065d38d0bbb47f55` |
| Runtime budget | 0 ₽ |

## Purpose

Stage 3.74 closes the user-facing ledger-repair gap without weakening the immutable financial ledger. A correction appends the next revision of the same logical transaction. A reversal appends a separate immutable command row with an explicit BusinessDate. No existing financial row is updated or deleted.

## Frozen contract reused

The implementation wires the already-frozen `correctTransaction` PATCH and `reverseTransaction` DELETE operations. `expectedRevision`, mandatory reasons, reversal `effectiveDate`, `Idempotency-Key`, the standard error envelope and the existing transaction/reversal response schemas remain authoritative. DELETE is command semantics only.

## Effective ledger materialization

Position and snapshot consumers no longer interpret raw immutable rows as independent economic events. The supported pipeline is:

`raw immutable entries → validated revision chains → latest logical transaction → reversal effective-date filter → deterministic effective ledger → Stage 3.71 position.Apply/Rebuild`.

Corrections are retroactive truth: the latest valid revision supplies the transaction fields for historical reconstruction. Reversal becomes effective on its explicit `effectiveDate`; before that BusinessDate the logical transaction remains visible in Time Machine, on/after it the transaction is excluded. Same-BusinessDate ordering uses the logical transaction's first ledger sequence, so a correction does not accidentally move an existing logical transaction merely because its revision was appended later.

Malformed revision chains, malformed reversal targets or duplicate reversals fail closed with `ErrUnsupportedPositionLedger`.

## Atomicity and concurrency

All correction/reversal mutations reuse the existing portfolio row `FOR UPDATE` serialization. `UNIQUE(transaction_id, revision)` remains the database backstop against duplicate revision numbers. Under the portfolio lock the runtime checks `expectedRevision` and reversal state, appends exactly one candidate entry, rebuilds the effective position ledger, rejects historical oversell, rebuilds affected snapshots, persists the exact replay artifact and commits atomically.

A migration is not required. The existing schema already contains `revision`, `prior_entry_id`, `correction_reason`, `reverses_transaction_id` and the unique revision constraint. The runtime role remains `SELECT, INSERT` only on `investment.transaction_entries`; UPDATE/DELETE/TRUNCATE stay revoked.

## Oversell safety

A backdated correction or reversal is validated against the full effective portfolio history before commit. If changing/reversing a BUY makes a later SELL impossible, the command returns the existing insufficient-position conflict and the new ledger row plus snapshot rebuilds roll back together.

## Snapshots and Time Machine

Stage 3.71 snapshot cash/invested-capital inputs and Stage 3.72/3.73 positions now consume the same effective ledger semantics. Correction rebuild planning starts from both the old and corrected trade dates; reversal rebuild planning starts from the reversal effective date. Historical positions therefore remain active before a later reversal and disappear on/after the reversal date.

Existing snapshot overflow behavior remains part of the frozen error contract: effective-ledger snapshot calculations preserve the established `ErrInvalidInput` classification when cumulative financial values exceed `NUMERIC(28,8)` storage precision.

## Transaction list UX

The main transaction list projects one row per logical transaction. Raw revisions and reversal command rows are not presented as independent trades. The current row exposes `ACTIVE`, `CORRECTED` with current revision, or `REVERSED`. Edit and Reverse actions use the frozen API. The correction command response also returns the new logical revision as `CORRECTED`, rather than exposing the raw-entry selector's transient `ACTIVE` interpretation. The reversal confirmation explicitly states that history is preserved.

The Stage 3.74 UI intentionally edits canonical financial fields needed for ledger repair without adding a forensic revision-history endpoint. A future dedicated audit-detail view would require separately reviewed contract scope.

## Idempotency

PATCH and DELETE are financial commands and use the existing exact-response replay system. Same key + same payload replays the original success without another revision/reversal. Same key + different payload conflicts. Replay resolution occurs before mutable business-state checks so a completed command remains exactly replayable after later ledger changes.

The Web repair controls also reuse the Stage 3.32 browser retry journal rather than generating a fresh key for every click. Correction and reversal have distinct principal-scoped technical scopes. The same unresolved intent therefore retains its key across an ambiguous transport retry/reload, changed intent rotates the key, confirmed success clears the journal, and explicit client-side/rejected 4xx outcomes clear retry state. Raw principal, portfolio and financial payload values are not persisted by the journal because the existing helper stores only a hashed technical scope and the opaque key.

## Security

Portfolio ownership is checked by the existing subject-scoped portfolio lock before target transaction state is read or mutated. A transaction in another subject/portfolio therefore preserves the existing not-found/anti-enumeration boundary.

## Verification vectors

The implementation adds PostgreSQL witnesses for correction WAC/basis, revision 1→2→3, stale revision conflict, correction-induced historical oversell rollback, reversal effective-date Time Machine behavior, reversal-induced oversell rollback and concurrent correction races. Replay witnesses use valid request/trace metadata so they exercise the same exact-response storage contract as production handlers. Existing Stage 3.29 snapshot-overflow tests and Stage 3.71–3.73 tests continue to cover storage bounds, close/reopen, same-date ordering, position projection, market-unavailable semantics and rapid historical-load protection. Frontend contract tests cover PATCH/DELETE, Edit/Reverse copy, mandatory reasons, destructive-action accessibility text, and reuse of the Stage 3.32 retry-safe idempotency journal.

## Non-scope

No market provider, market value, P/L, XIRR, inflation, corporate-action source activation, broker sync, imported SELL expansion, tax-basis methodology, bond NKD/YTM, AI, notifications, Redis, Kafka, workers, new SaaS or paid dependency is introduced.

## Expected result

A user can correct a mistaken price/quantity/date with a reason, see the current logical revision and recalculated position/WAC/basis, and reverse an operation with reason/effectiveDate while retaining full immutable history. Time Machine reflects the corrected/reversed economic truth, no accepted repair can create an oversold portfolio, and ambiguous browser retries do not silently become duplicate repair commands.
