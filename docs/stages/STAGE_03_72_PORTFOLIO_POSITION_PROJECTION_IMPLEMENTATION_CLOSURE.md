# Stage 3.72 — Portfolio Position Projection / Cost Basis View implementation closure

Canonical record: commit(s) `2144c52f4dc5ba9e917947d396ced1f3c572fe51`, `b6cd7766a6a584440dfdd7d00c35c34c8553a9ff`, `e935f19b68624f0b8be6ca293ead8f4395cf555e`, `3775b533867c2fb6de473da84a105cf3367cf848`.

## 1. Why Stage 3.72 exists

Stage 3.71 made portfolio position math authoritative, deterministic and financially safe, but deliberately stopped short of a public position projection. A user could record BUY/SELL transactions and the backend could rebuild quantity, Weighted Average Cost and remaining acquisition basis, yet the product still lacked an honest read model that exposed those facts directly.

Stage 3.72 closes that product gap without pretending that approved live market data exists. Its purpose is to let the user answer four cost-basis questions from the immutable ledger itself:

1. what open instruments are currently owned;
2. how many units are open;
3. what the authoritative average acquisition price is;
4. how much acquisition basis remains and what share of total acquisition basis that position represents.

Market valuation is not required to answer those questions. Therefore Stage 3.72 explicitly represents market valuation as unavailable instead of inventing prices, values, profit/loss or market weights.

## 2. What was implemented

### 2.1 Canonical backend projection

A dedicated portfolio-position projection was added on top of the Stage 3.71 engine. The implementation reuses canonical `position.Apply`; it does not introduce a second WAC or cost-basis calculator.

The PostgreSQL replay seam reconstructs position state from accepted BUY/SELL ledger rows in deterministic order and returns open STOCK/BOND positions with:

- `ticker`;
- `assetType`;
- exact Decimal quantity;
- authoritative Weighted Average Cost;
- remaining acquisition basis;
- independent acquisition-basis weight;
- explicit unavailable market-valuation state.

The aggregate denominator is the sum of acquisition basis for all open STOCK/BOND positions. Each weight is independently calculated with scale-8 Half-Even rounding. Final rounded weights are not force-normalized. If open positions exist but total acquisition basis is exactly `0.00000000`, the position weight is `null` rather than fabricated as zero.

Closed positions are omitted. Asset catalog lifecycle status does not erase an owned open position from the projection.

### 2.2 Deterministic as-of semantics

`GET /api/v1/portfolios/{portfolioId}/positions` accepts an endpoint-local optional `asOfDate`.

When supplied, only position-affecting ledger rows with `tradeDate <= asOfDate` are replayed, and `inputsAsOf` is the requested BusinessDate even when no qualifying rows exist.

When omitted, all accepted position-affecting rows are replayed, including accepted future-dated trades, and `inputsAsOf` is the maximum included trade date. An empty position ledger returns `items=[]`, zero RUB total acquisition basis and `inputsAsOf=null`.

Unsupported correction/reversal state fails closed. The endpoint does not reuse the older shared OpenAPI `AsOfDate` parameter whose default semantics refer to the latest completed MOEX business date.

### 2.3 Ownership and consistency

The PostgreSQL read executes inside one read-only Repeatable Read transaction. Portfolio ownership/subject isolation is checked before ledger replay in that same transaction. This preserves anti-enumeration semantics and prevents a projection from mixing authorization and financial reads across inconsistent snapshots.

Stable public ordering is `ticker ASC`, with an internal deterministic tie-breaker where needed.

### 2.4 Honest market-unavailable contract

Every Stage 3.72 position returns:

```text
status = UNAVAILABLE
reason = NO_APPROVED_MARKET_PRICE_SOURCE
marketPrice = null
marketValue = null
unrealizedGain = null
marketWeight = null
provider = null
asOf = null
```

No provider lookup occurs. No frontend or backend code derives market values from transaction prices or acquisition basis.

### 2.5 OpenAPI and HTTP boundary

The implementation adds:

```text
GET /api/v1/portfolios/{portfolioId}/positions
```

with `getPortfolioPositions` as the operation ID and responses for:

- `200` success;
- `400` invalid endpoint-local `asOfDate`;
- `401` unauthenticated request;
- `404` inaccessible/not-found portfolio with existing anti-enumeration behavior.

The contract adds explicit Stage 3.72 position, calculation and market-unavailable schemas plus a financially consistent example. `acquisitionBasisWeight` is constrained to `[0,1] | null`.

### 2.6 Next.js product surface

The portfolio detail view now loads portfolio metadata, summary, the dedicated Stage 3.72 positions projection and transactions together.

The positions UI provides explicit loading, positions-specific error and empty states, desktop table and mobile cards, quantity, average acquisition price, acquisition basis, acquisition-basis weight and a visible `Market valuation unavailable` state.

The UI never reconstructs positions from summary or transaction rows when the positions request fails.

A dedicated quantity formatter preserves fractional position precision instead of rounding tiny valid quantities to `0.00`. Ratio-to-percent conversion is presentation-only and does not recalculate allocation.

The stale-load guard binds results to request generation, access token, authenticated principal identity and portfolio identity before committing state. Accepted transaction/import mutations reload the positions projection.

## 3. Expected user-visible behavior

For the canonical example:

```text
BUY 100 @250
BUY 100 @300
SELL 50
```

Stage 3.72 exposes:

```text
SBER
Quantity               150
Average price           275 RUB
Acquisition basis    41,250 RUB
Acquisition-basis weight   derived from total open acquisition basis

Market valuation unavailable
```

The user receives a real cost-basis view based only on accepted ledger facts. They do not receive invented market price, market value, unrealized P/L or market allocation.

For the fractional witness, authoritative state remains:

```text
quantity = 0.20000000
WAC = 275.12345678
acquisitionBasis = 55.02469136
```

The implementation never re-derives WAC as `basis / quantity`.

## 4. Why this architecture was chosen

The Stage 3.72 projection is deliberately additive rather than a rewrite of Stage 3.71. This preserves four important boundaries:


A provider was not activated merely to make the screen look richer because that would combine source-rights, provenance, freshness, outage, currency and valuation-methodology decisions with a stage whose approved responsibility is ledger-derived position projection.

## 6. Preserved boundaries and non-goals

Stage 3.72 does **not** authorize or implement:

- approved/live market-price provider activation;
- market value or unrealized P/L;
- market-weight allocation;
- XIRR/return expansion;
- imported SELL expansion;
- correction/reversal runtime;
- FIFO/tax lots or tax basis;
- bond accrued interest/NKD;
- Corporate Actions Feature 3D source activation;
- notifications or ;
- Redis, worker or microservice expansion;
- production provider rollout.

`analytics.snapshot_positions` is untouched. Existing Stage 3.71 snapshot totals and legacy `PortfolioSummary.positions` semantics are intentionally preserved.

## 7. Residual risk and next decision boundary


That limitation is preferable to fabricated or weakly sourced market data. Stage 3.72 is complete when users can reliably inspect their ledger-derived open positions and acquisition basis while the system explicitly states that market valuation is unavailable.

Any later step that introduces a market provider, valuation, unrealized P/L, market weights, returns, tax basis or broader position semantics is a new separately reviewed stage and must not be inferred from this closure.

## 8. Closure statement

Stage 3.72 runtime implementation is complete and canonical through PR #145 squash merge `e935f19b68624f0b8be6ca293ead8f4395cf555e`. The implementation delivers the intended Portfolio Position Projection / Cost Basis View over the canonical Stage 3.71 engine, with deterministic replay, endpoint-local as-of behavior, acquisition-basis allocation, ownership isolation, Web presentation, OpenAPI contract and explicit honest market-unavailable semantics.
