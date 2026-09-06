# Stage 3.71 — Portfolio Position & Cost Basis Engine implementation

Status: implementation candidate; not merge-activated until the repository review/CI/governance gates complete.

Base authority: protected `develop@b772e52221fbb694b3116bd1b579db99d4e56302`, with Stage 3.70 / ADR-009 already merge-activated.

## 1. Scope

Stage 3.71 implements the runtime boundary frozen by Stage 3.70:

- deterministic portfolio-local ledger ordering using persisted `ledger_sequence`;
- exact Decimal(28,8), Half-Even Weighted Average Cost (`position-wac-v1`);
- remaining acquisition basis as `quantity × authoritative WAC`;
- manual BUY/SELL position rebuild and atomic oversell rejection;
- backdated historical replay validation;
- SELL-safe local-cost snapshots under `stage-03-71-position-cost-snapshot-v1`;
- manual Web SELL exposure only after the backend path exists;
- one narrow POST-transaction 409 contract extension for `INSUFFICIENT_POSITION_QUANTITY`.

Explicitly excluded: public position DTO activation, market price/value, unrealized P/L, XIRR, FIFO/tax lots, transaction correction/reversal runtime, imported SELL, provider activation, notifications, AI, Redis/workers, and Feature3D.

## 2. Financial invariants

Canonical state for an open position is quantity plus authoritative WAC. Remaining acquisition basis is derived and is never an independent repricing input.

BUY:

```text
oldBasis = Round8HalfEven(oldQty × oldWAC)
buyBasis = Round8HalfEven(buyQty × buyPrice)
newQty   = oldQty + buyQty
newWAC   = Round8HalfEven((oldBasis + buyBasis) / newQty)
newBasis = Round8HalfEven(newQty × newWAC)
```

SELL:

```text
sellQty > oldQty -> reject
partial SELL      -> WAC unchanged; basis = remainingQty × oldWAC
full SELL         -> position closed
later BUY         -> fresh WAC state
```

Commission and recorded tax remain separate cash/performance amounts and do not enter position WAC. Every persistence-bound derived value is checked against NUMERIC(28,8).

The fractional ADR-009 witness remains canonical: quantity `0.20000000`, WAC `275.12345678`, basis `55.02469136`, while `basis / quantity` rounds to `275.12345680`; that quotient must never reprice WAC.

## 3. Deterministic ledger ordering

Stage 3.71 adds nullable `investment.transaction_entries.ledger_sequence BIGINT` as an Expand-compatible schema change and then a portfolio-local unique index:

```text
UNIQUE(portfolio_id, ledger_sequence)
```

The column remains nullable during Expand so an old runtime is not broken merely by applying the schema migration. Runtime activation is fail-closed until the owner-only population step has completed and the unique index exists.

Pre-Stage-3.71 backfill uses exactly the ADR-009 frozen tuple:

```text
PARTITION BY portfolio_id
ORDER BY trade_date ASC,
         created_at ASC,
         transaction_id ASC,
         revision ASC,
         entry_id ASC
```

After population, runtime position math orders only by:

```text
trade_date ASC,
ledger_sequence ASC
```

The backfill command refuses partial population, historical SELL, and correction/reversal state. It updates only `ledger_sequence`; no financial fact is rewritten. Re-running after completed population is verify-only and never reorders later runtime rows.

## 4. Runtime allocation and atomicity

All accepted writes reuse the existing per-portfolio `FOR UPDATE` lock. The next sequence is allocated from the persisted portfolio-local maximum and fails closed if any prior row lacks a positive sequence or if BIGINT is exhausted.

Manual append flow:

```text
idempotency reservation
-> portfolio lock
-> allocate ledgerSequence
-> insert immutable transaction entry
-> replay affected asset history in (tradeDate, ledgerSequence) order
-> reject oversell/overflow
-> rebuild affected snapshots
-> persist success replay artifact
-> commit
```

The Stage 3.38 replay-aware path reaches the same sequence/rebuild helper through `insertTransactionEntryWithID`; no parallel financial implementation is introduced.

An oversell returns:

```text
HTTP 409
INSUFFICIENT_POSITION_QUANTITY
```

Because the ledger insert, snapshot rebuild, and success replay artifact share one SQL transaction, rejection leaves none of those artifacts committed.

## 5. Import boundary

Removing the old general SELL rejection would otherwise make a crafted import request eligible for the shared transaction validator. Stage 3.71 therefore keeps an explicit import-level SELL rejection. `TRANSACTION_TYPE_REQUIRES_FUTURE_REVIEW` remains the importer product gate and imported SELL is not activated.

Accepted import batches still receive consecutive portfolio-local sequence values in exact append-plan order because the existing batch transaction holds the portfolio lock and calls the same insertion helper once per request row.

## 6. Snapshot methodology

Stage 3.71 stops valuing remaining assets as `BUY gross - SELL gross` and uses position acquisition basis:

```text
stockValue = sum(open STOCK acquisitionBasis)
bondValue  = sum(open BOND acquisitionBasis)
```

Cash semantics remain:

```text
BUY  -> -(gross + commission + tax)
SELL -> +(gross - commission - tax)
```

The new snapshot methodology is:

```text
stage-03-71-position-cost-snapshot-v1
```

Existing Stage 3.02 rows are preserved as historical rows and are not relabelled. A later affected rebuild inserts a new Stage 3.71 methodology row for the required BusinessDate. Snapshot-date planning considers existing dates independent of old/new methodology so backdated writes preserve the Stage 3.33 exact affected-date behavior.

Public `summary.positions` remains empty because market-derived fields do not yet have an approved honest unavailable/quote source model.

## 7. Database rollout

The governed paired-SQL validator currently supports Expand DDL only. Stage 3.71 therefore does not smuggle Populate DML into migration SQL.

Rollout order:

1. apply `000008_stage_03_71_ledger_sequence` (nullable additive column);
2. apply `000009_stage_03_71_ledger_sequence_unique` (concurrent unique index);
3. run `go run ./cmd/backfill-ledger-sequence` with owner-only `OPENINVEST_DATABASE_OWNER_URL`;
4. verify zero NULL/non-positive/duplicate sequence rows;
5. deploy the Stage 3.71 runtime;
6. runtime readiness additionally verifies the unique index and complete positive sequence population.

The runtime database role remains append/read only for `investment.transaction_entries`; it never receives UPDATE/DELETE/TRUNCATE authority.

Rollback before runtime activation leaves the additive structure unused or uses the exact paired DOWN migrations. After Stage 3.71 financial writes are accepted, rollback must be application/configuration roll-forward aware and must not erase accepted immutable financial history.

## 8. Verification matrix

Required evidence covers:

- canonical BUY + BUY WAC;
- partial SELL with unchanged WAC;
- full close and reopen;
- fractional WAC non-invertibility witness;
- Decimal derived overflow;
- oversell rejection;
- backdated oversell against a later SELL;
- deterministic same-BusinessDate sequence order;
- concurrent SELL where only one request can legally commit;
- no ledger/snapshot/success-replay artifact after rejected oversell;
- imported SELL still rejected;
- import sequence range follows request order;
- snapshot stock/bond values use remaining acquisition basis;
- old Stage 3.02 rows are not relabelled;
- migration backfill tuple and structural verification;
- repeated position rebuild returns identical financial state.

## 9. Governance state

This document records implementation intent/evidence only. It does not claim CI success, published-head review approval, Ready state, merge, or protected-branch activation before those events actually occur.
