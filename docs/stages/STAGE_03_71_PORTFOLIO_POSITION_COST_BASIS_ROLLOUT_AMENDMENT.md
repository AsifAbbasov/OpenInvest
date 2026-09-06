# Stage 3.71 — Portfolio Position & Cost Basis Engine rollout amendment

Status: required activation amendment discovered during final adversarial rollout review. This amendment is additive to the frozen Stage 3.71 implementation authority and does not change ADR-009 financial semantics, migration SQL, WAC arithmetic, or ledger ordering.

Canonical base: protected `develop@b772e52221fbb694b3116bd1b579db99d4e56302`.

Related authority: `docs/stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_IMPLEMENTATION.md`.

## 1. Problem discovered

The Stage 3.71 Expand/Populate design intentionally leaves `investment.transaction_entries.ledger_sequence` nullable until activation so the pre-Stage-3.71 runtime is not broken merely by applying migration `000008`.

The owner-only backfill then populates all historical rows and intentionally refuses any partially populated state.

A cutover race exists if the legacy runtime remains writable after the backfill commits:

```text
000008 + 000009 applied
-> backfill completes with every existing row populated
-> backfill transaction releases its table lock
-> legacy runtime accepts a new transaction
-> legacy INSERT writes ledger_sequence = NULL
-> Stage 3.71 runtime starts
-> Stage371Ready detects incomplete sequence population and fails closed
```

A second backfill invocation also fails closed because the database is now partially populated. This behavior is correct for data safety but means the original rollout steps were operationally incomplete.

## 2. Required activation invariant

Stage 3.71 activation requires a continuous **portfolio-write quiescence window** spanning the Populate-to-runtime transition.

The quiescence window MUST begin before the owner-only backfill starts and MUST remain in force until the Stage 3.71 runtime is deployed and its readiness check succeeds.

During this window no pre-Stage-3.71 process may append portfolio transaction rows.

Acceptable zero-budget/MVP mechanisms include stopping the API instance, disabling the portfolio write route at the deployment edge, or otherwise draining/blocking portfolio writes. The mechanism chosen for an environment must be observable by the operator; merely assuming low traffic is not sufficient.

## 3. Revised rollout sequence

The activation sequence is:

1. apply `000008_stage_03_71_ledger_sequence`;
2. apply `000009_stage_03_71_ledger_sequence_unique`;
3. enter portfolio-write quiescence / maintenance mode;
4. prove legacy portfolio writes are stopped or blocked;
5. run `go run ./cmd/backfill-ledger-sequence` with owner-only `OPENINVEST_DATABASE_OWNER_URL`;
6. verify the exact unique index and zero NULL/non-positive/duplicate `ledger_sequence` rows;
7. deploy the Stage 3.71 runtime while write quiescence remains active;
8. require Stage 3.71 readiness success;
9. only then reopen portfolio writes;
10. monitor the first accepted writes for positive portfolio-local sequences and normal snapshot/replay behavior.

The backfill's `SHARE ROW EXCLUSIVE` lock protects the table while Populate itself runs. The operational quiescence requirement covers the period after that transaction commits and before the new runtime is proven ready.

## 4. Abort conditions

Activation MUST abort before reopening writes if any of the following occurs:

- backfill reports partial population;
- any NULL/non-positive sequence row exists after Populate;
- the exact `000009` index is missing, invalid, not ready, partial, expression-based, or structurally wrong;
- a legacy runtime/process is still capable of portfolio writes;
- Stage 3.71 readiness fails;
- deployment identity does not match the reviewed Stage 3.71 candidate.

Do not perform an ad hoc partial sequence repair in place. The fail-closed partial-population rule remains authoritative.

## 5. Recovery before Stage 3.71 activation

If a legacy write occurs after a completed backfill but before Stage 3.71 activation, keep portfolio writes closed and treat the cutover as aborted.

Because Stage 3.71 financial writes have not yet been accepted, the additive Stage 3.71 schema can be returned to a clean pre-activation state using the governed paired DOWN/UP path as appropriate for the environment, then the full quiesced rollout can be repeated. Immutable legacy financial facts must not be deleted or rewritten.

The operator must not drop Stage 3.71 structures after Stage 3.71 financial writes have been accepted merely to repair activation state; after activation, recovery is roll-forward aware.

## 6. Verification evidence required for Ready/activation

Repository CI proves migration structure, backfill behavior, readiness logic and runtime semantics. Environment activation additionally requires operator evidence for:

- write quiescence began before backfill;
- no legacy portfolio writer remained active during cutover;
- backfill completed successfully;
- structural sequence verification passed;
- the reviewed Stage 3.71 runtime revision was deployed;
- Stage 3.71 readiness passed before writes reopened.

For the current zero-budget development/MVP environment this can be a concise deployment log/checklist; a separate paid orchestration system is not required.

## 7. Governance effect

This amendment closes an operational activation gap only. It does not authorize merge or production activation by itself and does not modify:

- `trade_date ASC, ledger_sequence ASC` runtime ordering;
- the historical ADR-009 backfill tuple;
- WAC/remaining acquisition-basis semantics;
- imported SELL exclusion;
- public position DTO availability;
- market-data behavior;
- later-stage scope.

Stage 3.71 remains Draft until the current exact PR head is green and review/governance gates complete.