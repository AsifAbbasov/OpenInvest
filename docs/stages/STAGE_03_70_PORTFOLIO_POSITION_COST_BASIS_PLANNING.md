# Stage 3.70 — Portfolio Position & Cost Basis Engine Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-70-PORTFOLIO-POSITION-COST-BASIS-PLAN |
| Version | 0.1.1-candidate |
| Status | Merge-activated planning decision — candidate/non-normative before required review/CI/evidence gates; COMPLETE/CANONICAL only after explicit Principal Architect acceptance and squash merge of this exact decision to protected `develop`; no Stage 3.71 runtime/Ready/merge authorization by this document |
| Owner | Principal Architect / Portfolio & Analytics |
| Canonical planning base | `develop@d2258134433fe214695db43db7de3b6bf003e9cf` |
| Protected-base tree | `9b73746a298a4d36c6dcb0e1037ba45b17dfb0e6` |
| Dependencies | Documents 42–43; ADR-006; Stage 2 API/canonical/ER freeze; Stage 3.02; Stage 3.33; Stage 3.69; `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Proposed decision | ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics |
| Architecture Issue | GitHub #136 — `https://github.com/AsifAbbasov/OpenInvest/issues/136` |
| Date | 2026-09-06 |

## 1. Purpose and why now

Freeze the smallest correct boundary needed to rebuild positions from the immutable transaction ledger and safely activate `SELL`.

The engine must answer one question:

> Given the immutable ledger, what are the exact remaining quantity, Weighted Average Cost (WAC), and remaining acquisition basis for each asset at any BusinessDate?

This is the next dependency below SELL, position summary, P/L, XIRR, real return and attribution. It requires no external provider, paid source, scraping, Redis, worker or AI service.

Current repository evidence already establishes:

- canonical BUY/SELL transaction types;
- append-only `investment.transaction_entries`;
- Decimal scale 8 / precision 28 / Half Even;
- portfolio-level PostgreSQL write serialization;
- rebuildable snapshots;
- canonical position vocabulary with quantity/WAC;
- Stage 3.02 runtime rejection of SELL until cost-basis rebuild exists.

## 2. Governance classification

This is documentation-only but changes implementation-affecting financial semantics, so it follows the **development path** in `REVIEW_WORKFLOW.md` v1.4.0.

```text
local candidate
→ local checks
→ Internal read-only full review
→ explicit human commit/push permission
→ Draft PR + CI
→ fresh External published-head review
→ Internal evidence publication + evidence CI/verification
→ explicit human ADR acceptance + Ready/squash-merge authorization
→ protected develop
```

Stage 3.71 runtime coding is forbidden until Stage 3.70/ADR-009 are canonical and separately authorized.

## 3. Canonical terminology

### Position quantity

Remaining quantity of one canonical asset after ordered ledger replay.

### Acquisition basis

Remaining RUB **trade-price basis** of the open position.

It is not tax basis and excludes:

```text
commission
tax
inflation
FX
bond accrued coupon / NKD
```

### WAC

Per-unit acquisition cost under methodology:

```text
position-wac-v1
```

WAC is not market price.

The immutable ledger remains source of truth; position is only a rebuildable projection.

## 4. Exact Decimal rules

All position arithmetic uses the existing standard:

```text
NUMERIC(28,8)
scale 8
precision 28
Half Even
no float32 / float64
```

Every persistence-bound derived value must pass `FitsStorage()` after arithmetic. Overflow fails closed.

## 5. BUY methodology

Canonical open-position state is:

```text
quantity
weightedAverageCost
```

Remaining acquisition basis is derived, not independently carried:

```text
acquisitionBasis = Round8HalfEven(quantity × weightedAverageCost)
```

For a new/reopened position:

```text
newQty   = buyQty
newWAC   = Round8HalfEven(buyPrice)
newBasis = Round8HalfEven(newQty × newWAC)
```

For an existing position:

```text
oldBasis = Round8HalfEven(oldQty × oldWAC)
buyBasis = Round8HalfEven(buyQty × buyPrice)
newQty   = oldQty + buyQty
newWAC   = Round8HalfEven((oldBasis + buyBasis) / newQty)
newBasis = Round8HalfEven(newQty × newWAC)
```

`weightedAverageCost` is authoritative at scale 8. After a partial SELL, the engine MUST NOT re-derive WAC by dividing rounded basis by fractional quantity. Consequently the accounting replay order is part of the deterministic contract; Stage 3.70 does not claim commutativity for arbitrary rounded BUY permutations.

## 6. SELL methodology

Given `oldQty > 0`, `oldWAC`, and `sellQty > 0`:

### Oversell

```text
sellQty > oldQty
→ reject
```

### Partial SELL

```text
newQty   = oldQty - sellQty
newWAC   = oldWAC
newBasis = Round8HalfEven(newQty × oldWAC)
```

SELL price does not change remaining WAC.

### Full close

```text
sellQty == oldQty
→ quantity = 0
→ acquisitionBasis = 0
→ open WAC state ends
```

A later BUY starts a new WAC history.

SELL price, commission and recorded tax remain separate cash/performance facts. Realized P/L is not part of Stage 3.71.

## 7. Commission, tax and bond boundary

Freeze the separation:

```text
WAC / acquisition basis = trade-price basis only
commission               = separate expense
recorded tax              = separate financial fact
```

BUY commission/tax do not increase WAC. SELL commission/tax do not change remaining WAC.

The engine may rebuild trade-price quantity/WAC for canonical STOCK and BOND rows, but Stage 3.71 does not claim bond NKD, YTM, duration, amortization or tax-grade bond basis. Any later NKD-aware basis requires a new methodology/version.

## 8. Deterministic same-BusinessDate order

`tradeDate` alone is insufficient. UUID and `createdAt` must not remain hidden financial ordering rules.

ADR-009 proposes an immutable positive portfolio-local ledger sequence.

Canonical replay order:

```text
tradeDate ASC
ledgerSequence ASC
```

Rules:

- sequence is portfolio-scoped and immutable;
- allocated inside the existing serialized portfolio write transaction;
- concurrent accepted writes become one linearizable order;
- backdated append receives the next sequence but replays on its own BusinessDate;
- system timestamps and UUID sort are not ongoing position-math inputs.

Because the frozen transaction API has no intraday execution timestamp/order, Stage 3.71 treats accepted `ledgerSequence` as the **canonical OpenInvest accounting order** inside one BusinessDate, not as a claim about broker execution time. Manual same-day trades must therefore be entered in the intended economic order. Reordering already-accepted same-day entries is outside Stage 3.71; broker-import SELL stays blocked until broker/source ordering semantics are reviewed.

## 9. Migration/backfill boundary

Stage 3.71 may add an additive `ledger_sequence BIGINT` plus uniqueness equivalent to:

```text
UNIQUE(portfolio_id, ledger_sequence)
```

Migration must be Expand → Populate → Validate and must not rewrite/delete financial facts.

For **pre-Stage-3.71 rows only**, the deterministic backfill order is frozen as:

```text
PARTITION BY portfolio_id
ORDER BY trade_date ASC,
         created_at ASC,
         transaction_id ASC,
         revision ASC,
         entry_id ASC
→ ledgerSequence = ROW_NUMBER() starting at 1
```

`created_at`, transaction/revision identity and UUID are allowed here only as one-time deterministic migration tie-breakers. After backfill, they are not position-math ordering inputs.

Before implementation publication, migration tests must prove:

1. pre-Stage-3.71 manual SELL is rejected;
2. imported SELL remains blocked as future-review;
3. no accepted historical SELL order exists to reinterpret;
4. the frozen backfill establishes the first canonical WAC replay order for historical BUY-only rows;
5. backfill follows the frozen tuple exactly and repeated clean migration produces the same sequence assignment;
6. existing supported imported rows are preserved;
7. after backfill, runtime calculations use persisted sequence only.

If any premise is false, stop and redesign the migration before publication.

For future writes, sequence allocation reuses the existing portfolio lock; no global counter service, Redis, Kafka or distributed lock is justified. A manual append receives the next positive portfolio-local sequence. An accepted import batch receives one consecutive sequence range in the exact existing `request.Transactions` append-plan order. Sequence exhaustion fails closed; values never wrap.

Existing supported import rows therefore receive deterministic sequence values without enabling imported SELL.

## 10. Oversell, backdated writes and atomicity

Oversell is a historical invariant, not only a check against the latest quantity.

A backdated SELL can invalidate a later existing SELL. Therefore a candidate position-affecting transaction is accepted only if replay of the affected asset history with the candidate included never produces quantity below zero.

Recommended transactional flow:

```text
reserve/enter DB transaction
→ lock portfolio
→ assign sequence
→ insert candidate
→ replay affected ordered asset history
→ reject on any negative quantity/overflow
→ rebuild affected snapshots
→ complete success replay artifact
→ commit
```

Any failure rolls the whole transaction back.

For insufficient owned quantity, use the Stage 3.70 frozen business conflict semantics:

```text
HTTP 409
INSUFFICIENT_POSITION_QUANTITY
```

The current POST transaction OpenAPI binds 409 specifically to `IdempotencyConflict`, so this error is **not** already representable by the frozen endpoint contract. Stage 3.71 is authorized to make one narrow OpenAPI change for POST transaction conflicts: preserve `IDEMPOTENCY_CONFLICT` and add `INSUFFICIENT_POSITION_QUANTITY` under a transaction-specific `TransactionConflict` 409 response while preserving the existing `ErrorResponse` envelope.

No committed transaction, corrupted snapshot or successful replay artifact may remain after rejection.

Concurrent operations are linearizable through the portfolio lock/sequence. The system does not promise which simultaneous request wins the lock; it promises that committed history is equivalent to one serial order and can never contain a negative position.

## 11. Snapshot boundary

Stage 3.02 temporarily treats local acquisition cost as asset value because live market data is unavailable. The current aggregate formula `BUY gross - SELL gross` becomes invalid once sale proceeds differ from remaining acquisition cost.

Stage 3.71 must therefore version the local-cost snapshot methodology so that:

```text
stockValue = sum(open STOCK acquisitionBasis)
bondValue  = sum(open BOND acquisitionBasis)
```

Use a new methodology version, e.g.:

```text
stage-03-71-position-cost-snapshot-v1
```

Do not relabel old `stage-03-02-local-cost-snapshot-v1` history.

Backdated writes rebuild their BusinessDate and every existing later affected snapshot date. Preserve the Stage 3.33 one-pass principle where practical instead of N full-history scans.

“Rebuild twice = identical” means identical financial state/methodology/input watermark, not identical generated UUIDs, versions or timestamps.

## 12. Public position projection stays off

The existing public `PortfolioPosition` and `analytics.snapshot_positions` require market-derived fields:

```text
marketPrice
marketValue
unrealizedGain
weight
```

There is no approved live market-price source. Stage 3.71 must not fill these with WAC, zero, examples or fabricated prices.

Therefore Stage 3.71 may use internal/transient position rebuild for SELL validation and local-cost snapshot aggregates, but **must not activate public `summary.positions`**.

A later Portfolio Position Projection stage must solve honest market-unavailable semantics.

## 13. API, Web, import and reversal scope

The frozen transaction OpenAPI already contains SELL, so Stage 3.71 may:

- remove the Stage 3.02 runtime SELL rejection;
- return the existing canonical transaction DTO;
- preserve existing 201/400/401/404 semantics;
- make the one explicitly planned OpenAPI change to POST transaction 409 so it can represent both idempotency conflict and insufficient-position business conflict;
- expose SELL in the current Web Add Transaction form only after backend correctness is proven.

The current POST transaction contract does not list 429. Stage 3.70 does not authorize silently adding a 429 response as part of this feature; any such contract parity work requires separate evidence/scope justification.

Stage 3.71 must not:

- change public position DTOs;
- calculate WAC in Web/mobile;
- enable imported SELL (`TRANSACTION_TYPE_REQUIRES_FUTURE_REVIEW` remains);
- implement transaction PATCH correction;
- implement transaction DELETE reversal.

Future correction/reversal must feed an effective append-only ledger projection into the same position engine and preserve explicit reversal BusinessDate semantics.

## 14. Mandatory financial vectors and tests

Document 42 requires canonical WAC vectors under `tests/financial/` before production readiness.

Core vector:

```text
BUY 100 SBER @ 250
BUY 100 SBER @ 300

quantity = 200.00000000
basis    = 55000.00000000 RUB
WAC      = 275.00000000 RUB

SELL 50

quantity = 150.00000000
basis    = 41250.00000000 RUB
WAC      = 275.00000000 RUB

SELL 150
→ closed

BUY 20 @ 310
quantity = 20.00000000
basis    = 6200.00000000 RUB
WAC      = 310.00000000 RUB
```

Mandatory fractional-quantity regression vector proving SELL preserves authoritative WAC without inverse-rounding it from basis:

```text
BUY 0.30000000 @ 275.12345678
WAC   = 275.12345678
basis = 82.53703703 RUB

SELL 0.10000000
quantity = 0.20000000
WAC      = 275.12345678
basis    = 55.02469136 RUB

Round8(55.02469136 / 0.20000000) = 275.12345680, but canonical WAC remains 275.12345678.
Do not calculate WAC as Round8(basis / quantity) after this SELL.
```

Minimum Stage 3.71 matrix:

```text
BUY
BUY + BUY
multiple-price BUY
BUY + SELL
multiple SELL
full close
reopen
oversell
zero quantity / zero price rejection
max Decimal
buy-basis overflow
WAC Half Even tie
fractional SELL preserves authoritative WAC without inverse repricing
remaining-basis overflow
same BusinessDate ordering
backdated BUY
backdated SELL invalidating later history
rebuild twice → same financial state
concurrent BUY + BUY
concurrent BUY + SELL
concurrent SELL + SELL
manual append + import serialization
same idempotency key replay
same key / different payload conflict
oversell rollback leaves no successful business/replay state
migration/backfill integration
```

PostgreSQL integration tests are mandatory for ordering, concurrency, rollback, migration/backfill and affected snapshot rebuild.

## 15. Explicit exclusions

```text
XIRR
Nominal / Real Return activation
inflation / purchasing power
realized or unrealized P/L
market-price provider activation
public PortfolioPosition activation
portfolio weight / attribution
correction PATCH / reversal DELETE
broker-import SELL
DIVIDEND/COUPON/FEE/TAX import expansion
bond NKD / YTM / duration / amortization
FIFO / tax lots
Corporate Actions Feature 3D
notifications / AI / Monte Carlo
Redis / workers / cron
new paid or external service
```

## 16. Stage 3.71 responsibility and stop conditions

One implementation responsibility:

> Deterministically rebuild quantity, WAC and remaining trade-price acquisition basis from an immutable ordered portfolio ledger; reject oversell atomically; make existing local-cost snapshots SELL-safe.

Expected surfaces: pure Go engine/tests, financial vectors, additive sequence migration, transaction/import append integration, snapshot methodology update, the narrow POST transaction 409 contract/mapping change frozen here, focused Web SELL option, integration/concurrency evidence and Stage 3.71 implementation record.

No review-size exception is pre-authorized.

STOP before Stage 3.71 if:

- ADR-009 is not canonical;
- required Issue → ADR governance is incomplete;
- historical accepted SELL evidence contradicts the backfill assumptions;
- Decimal round points differ across paths;
- oversell cannot be atomic with ledger/replay/snapshot writes;
- implementation needs fabricated market values or public-position contract drift;
- scope expands into reversal/correction, imported SELL, XIRR, tax, market data or bond NKD.

GitHub Issue #136 carries the admitted architecture question, owner, due date, affected decisions and proposed ADR-009. The Issue does not authorize implementation or repository publication by itself.

After Stage 3.70/ADR-009 is accepted and squash-merged to protected `develop`, the next action is a separate explicit human authorization for Stage 3.71 runtime implementation. No additional multi-stage planning programme is required unless review finds a new blocking architecture question.
