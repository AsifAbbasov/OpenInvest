# ADR-009: Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics

| Field | Value |
| --- | --- |
| Document ID | ADR-009 |
| Version | 0.1.0-candidate |
| Status | Merge-activated decision — PROPOSED/non-normative before required gates; ACCEPTED only after explicit Principal Architect acceptance and squash merge of this exact decision to protected `develop` |
| Owner | Principal Architect |
| Supersedes | Implicit same-BusinessDate ordering and incomplete Stage 3.02 local-cost position semantics |
| Dependencies | Documents 42–43; ADR-006; Stage 2 canonical/API/ER freeze; Stage 3.02; Stage 3.33; Stage 3.70 |
| GitHub issue | #136 — `https://github.com/AsifAbbasov/OpenInvest/issues/136` |
| Date | 2026-09-06 |

## Context

Stage 3.02 intentionally rejects SELL until WAC/cost-basis position rebuild exists. The remaining ambiguity is financially material:

- multiple trades may share one BusinessDate;
- UUID is identity, not economic order;
- SystemTimestamp must not become hidden financial ordering;
- BUY/SELL order can change oversell validity and remaining WAC;
- the old local-cost snapshot formula subtracts SELL proceeds instead of remaining acquisition basis;
- commission/tax and bond NKD must not be silently mixed into WAC;
- public positions cannot honestly expose market-derived fields while market-price runtime is dormant.

## Decision

### 1. Source of truth

The immutable transaction ledger remains source of truth. Position state is a deterministic projection; no mutable balance replaces the ledger.

### 2. Ordering

Every accepted ledger row receives an immutable positive portfolio-local `ledgerSequence`.

Canonical replay order is:

```text
tradeDate ASC
ledgerSequence ASC
```

Sequence is assigned under the existing serialized portfolio write transaction. Concurrent writes become one linearizable accepted order. `createdAt` and UUID ordering are not recurring position-math inputs.

The frozen API has no intraday execution-order field. Therefore `ledgerSequence` is the canonical OpenInvest accounting order within a BusinessDate; it does not claim to reconstruct broker execution time. Manual same-day trades must be appended in the intended economic order. Reordering accepted rows is outside this ADR, and imported SELL remains blocked until broker/source ordering semantics are separately reviewed.

### 3. Methodology

Methodology version:

```text
position-wac-v1
```

All arithmetic uses Decimal scale 8 / precision 28 / Half Even; persistence-bound derived values must fit NUMERIC(28,8).

Canonical open-position state is:

```text
quantity
acquisitionBasis
```

with:

```text
WAC = Round8HalfEven(acquisitionBasis / quantity)
```

`acquisitionBasis` means remaining RUB trade-price basis only. It is not tax basis and excludes commission, recorded tax, FX, inflation and bond accrued coupon/NKD.

### 4. BUY

```text
buyBasis = Round8HalfEven(buyQty × buyPrice)
newQty   = oldQty + buyQty
newBasis = oldBasis + buyBasis
newWAC   = Round8HalfEven(newBasis / newQty)
```

Basis-first accumulation avoids unnecessary cumulative rounding and makes BUY-only accumulation commutative.

### 5. SELL

```text
sellQty > oldQty
→ reject
```

Partial SELL:

```text
newQty   = oldQty - sellQty
newWAC   = oldWAC
newBasis = Round8HalfEven(newQty × oldWAC)
```

Full close ends open WAC state. A later BUY starts a new position.

SELL price, commission and recorded tax do not change remaining WAC.

### 6. Historical oversell invariant

A candidate trade is valid only if ordered replay of the affected asset history including that candidate never produces quantity below zero. This includes later entries when the candidate is backdated.

Insufficient quantity uses the existing Conflict surface:

```text
HTTP 409
INSUFFICIENT_POSITION_QUANTITY
```

Ledger append, affected snapshot rebuild and successful replay artifact remain one atomic transaction. Oversell rolls all of them back.

### 7. Migration/backfill

Stage 3.71 may add `ledger_sequence BIGINT` with portfolio-local uniqueness using Expand → Populate → Validate.

For **pre-Stage-3.71 rows only**, backfill is frozen as:

```text
PARTITION BY portfolio_id
ORDER BY trade_date ASC,
         created_at ASC,
         transaction_id ASC,
         revision ASC,
         entry_id ASC
→ ledgerSequence = ROW_NUMBER() starting at 1
```

The timestamp/identity/UUID fields are one-time migration tie-breakers only; they do not become ongoing financial ordering semantics. Publication tests must prove that pre-Stage-3.71 supported runtime/import paths contain no accepted SELL history whose former order would be reinterpreted, BUY-only basis accumulation is order-independent, repeated clean migration yields the same assignment, existing supported import rows are preserved, and all later calculations use persisted sequence only. If any premise is false, implementation stops for redesign.

Future manual writes receive the next positive portfolio-local sequence under the existing lock. One accepted import batch receives a consecutive range preserving exact `request.Transactions` append-plan order. Sequence exhaustion fails closed without wrap. Imported SELL remains blocked.

### 8. Snapshot methodology

Once SELL is active, local-cost asset values become:

```text
stockValue = sum(open STOCK acquisitionBasis)
bondValue  = sum(open BOND acquisitionBasis)
```

A new methodology version is required. Old Stage 3.02 snapshots are historical and must not be relabelled.

### 9. Public positions remain deferred

Stage 3.71 does not activate public `PortfolioSummary.positions` or populate fake market fields in `analytics.snapshot_positions`.

Market price/value, unrealized gain and weight require honest market-data/unavailable semantics later. WAC must never be substituted as market price.

### 10. Deferred boundaries

This ADR does not activate:

```text
transaction correction/reversal runtime
imported SELL
FIFO / tax lots
bond NKD-aware basis
XIRR / returns / P&L
market-price provider
public position projection
```

Future correction/reversal must feed an effective append-only ledger projection into the same ordering/WAC engine and preserve BusinessDate semantics.

## Rationale

- **Portfolio-local sequence** is the smallest auditable order because writes are already serialized per portfolio.
- **System time** is rejected as financial order because it is technical metadata.
- **UUID order** is rejected because identity is not business sequence.
- **Basis-first BUY accumulation** gives explicit Decimal round points and avoids avoidable rounding drift.
- **SELL preserves WAC** because sale proceeds do not reprice remaining acquired units under WAC.
- **Commission/tax stay separate** because current documented WAC is price-based and those values belong to separate performance/tax semantics.
- **No public positions yet** because the frozen DTO requires market-derived fields that cannot be fabricated.

## Alternatives rejected

### Latest-quantity oversell check only

Rejected: a backdated SELL can invalidate a later historical SELL.

### Mutable position/balance as source of truth

Rejected: violates append-only auditability and deterministic historical rebuild.

### `createdAt` or UUID ordering

Rejected: both are accidental technical ordering, not a frozen economic contract.

### FIFO now

Rejected: separate accounting methodology and scope expansion.

### Include commission/tax in WAC

Rejected for v1: silently changes the currently documented price-based WAC methodology.

### Global sequence / Redis / distributed lock

Rejected: unnecessary infrastructure; correctness is portfolio-local and PostgreSQL already supplies the transaction boundary.

## Consequences

### Positive

- safe manual SELL foundation;
- deterministic historical positions;
- auditable same-day/concurrent order;
- exact Decimal WAC vectors;
- SELL-safe local-cost snapshots;
- no external service or licensing cost;
- clean dependency for later position projection, P/L, XIRR and attribution.

### Costs

- additive migration/backfill;
- sequence allocation in manual/import append paths;
- affected-history replay for oversell validation;
- new snapshot methodology version;
- public positions remain deferred;
- bond NKD/tax-basis completeness remains future work.

## Compatibility and rollback

Before runtime implementation, rollback is reversion of this planning decision.

After Stage 3.71, rollback must never delete/rewrite accepted ledger rows. New sequence metadata is additive and old/new snapshot methodology versions remain distinguishable. If runtime rollback is required, disabling new SELL acceptance is safer than destructive history rollback.

## Acceptance condition

ADR-009 is not accepted by drafting, local review, commit or push. It becomes canonical only after:

1. the required GitHub architecture issue exists and is referenced here;
2. development-path review/CI/evidence gates pass;
3. the Principal Architect explicitly accepts this exact decision and authorizes Ready/squash merge;
4. the accepted text is squash-merged into protected `develop`.

Until then, the Stage 3.02 SELL rejection remains authoritative runtime behavior.
