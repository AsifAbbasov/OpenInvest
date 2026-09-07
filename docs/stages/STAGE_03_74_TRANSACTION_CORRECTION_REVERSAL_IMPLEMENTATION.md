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

## Transaction list UX

The main transaction list projects one row per logical transaction. Raw revisions and reversal command rows are not presented as independent trades. The current row exposes `ACTIVE`, `CORRECTED` with current revision, or `REVERSED`. Edit and Reverse actions use the frozen API. The reversal confirmation explicitly states that history is preserved.

The Stage 3.74 UI intentionally edits canonical financial fields needed for ledger repair without adding a forensic revision-history endpoint. A future dedicated audit-detail view would require separately reviewed contract scope.

## Idempotency

PATCH and DELETE are financial commands and use the existing exact-response replay system. Same key + same payload replays the original success without another revision/reversal. Same key + different payload conflicts. Replay resolution occurs before mutable business-state checks so a completed command remains exactly replayable after later ledger changes.

## Security

Portfolio ownership is checked by the existing subject-scoped portfolio lock before target transaction state is read or mutated. A transaction in another subject/portfolio therefore preserves the existing not-found/anti-enumeration boundary.

## Verification vectors

The implementation adds PostgreSQL witnesses for correction WAC/basis, revision 1→2→3, stale revision conflict, correction-induced historical oversell rollback, reversal effective-date Time Machine behavior, reversal-induced oversell rollback and concurrent correction races. Existing Stage 3.71–3.73 tests continue to cover close/reopen, same-date ordering, position projection, market-unavailable semantics and rapid historical-load protection. Frontend contract tests cover PATCH/DELETE, Edit/Reverse copy, mandatory reasons and destructive-action accessibility text.

## Non-scope

No market provider, market value, P/L, XIRR, inflation, corporate-action source activation, broker sync, imported SELL expansion, tax-basis methodology, bond NKD/YTM, AI, notifications, Redis, Kafka, workers, new SaaS or paid dependency is introduced.

## Expected result

A user can correct a mistaken price/quantity/date with a reason, see the current logical revision and recalculated position/WAC/basis, and reverse an operation with reason/effectiveDate while retaining full immutable history. Time Machine reflects the corrected/reversed economic truth, and no accepted repair can create an oversold portfolio.
