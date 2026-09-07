# Stage 3.76 — Manual Market Valuation & Portfolio P/L

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION CANDIDATE / NON-CANONICAL |
| Owner | Builder Engineer |
| Canonical base | `develop@e7c15c6587f114abc9eb9e601dd060b8e720c100` |
| Budget | 0 RUB |
| External market-data provider | NONE |
| Authorized price provenance | `USER_SUPPLIED` only |
| Next stage | Stage 3.77 NOT AUTHORIZED |

## Purpose

Stage 3.76 lets an authenticated portfolio owner persist an explicit manual RUB price for an open STOCK/BOND position and see backend-derived market value, unrealized gain/loss, unrealized return, portfolio valuation coverage, valued-position allocation, cash, and a complete current portfolio value only when every open position is valued.

Manual input is never represented as an official, live, MOEX, broker, or provider quote.

## Architecture

The implementation reuses the canonical chain:

`effective ledger -> WAC/acquisition basis -> positions -> manual valuation -> derived P/L`

Manual valuation is valuation input, not a transaction and not a second financial ledger. Stage 3.57 `QuoteProvider` remains dormant. No provider activation, HTTP market feed, scraping, polling, WebSocket, worker, Redis, Kafka, SaaS, or paid infrastructure is added.

## Persistence and rollout

One current-state table, `investment.portfolio_manual_valuations`, stores one user-supplied valuation per `(portfolio_id, asset_id)`:

- portfolio and asset identity;
- current position generation (`position_opened_ledger_sequence`);
- exact `NUMERIC(28,8)` positive price;
- RUB currency;
- explicit BusinessDate;
- source `USER_SUPPLIED`;
- update timestamp.

The existing paired-SQL migration validator intentionally accepts only a narrow frozen DDL grammar. Stage 3.76 does not weaken that validator. The schema therefore rolls out through two additive governed migrations:

1. `000010_stage_03_76_manual_valuation` creates the empty table and unique `(portfolio_id, asset_id)` index.
2. `000011_stage_03_76_manual_valuation_fks` adds portfolio and asset foreign keys as `NOT VALID`; PostgreSQL enforces them for new writes without requiring a historical table scan.

The physical columns created by `000010` are nullable because inline NOT NULL/CHECK constraints are outside the frozen validator grammar. Runtime writes always provide every field, typed scans reject NULLs, and read validation fails closed for invalid price, generation, currency, source, or date. No accepted Stage 3.76 behavior treats malformed stored data as usable valuation input.

After migrations, the existing runtime-role grant script must be rerun. Stage 3.76 readiness performs a zero-row projection over every required column, so an instance fails readiness when the table is missing or the runtime role lacks access.

The table is current state, not price history. PUT overwrites the current manual valuation deterministically. DELETE is idempotent and can clear a stale valuation even after a position has fully closed.

## Position lifecycle guard

The position-generation marker is the first effective BUY ledger sequence in the current continuous open-position lifecycle.

- correction of the logical BUY preserves its logical sequence and generation;
- partial SELL preserves the generation;
- reversal/correction recomputes through the effective ledger;
- full close followed by a new BUY creates a new generation.

A stored valuation whose generation does not match the reconstructed open position is ineligible, so a stale price cannot silently reactivate after close/reopen.

## API

Authenticated portfolio-scoped endpoints:

- `PUT /api/v1/portfolios/{portfolioId}/valuations/{ticker}`
- `DELETE /api/v1/portfolios/{portfolioId}/valuations/{ticker}`

PUT body contains exact Money plus explicit `asOfDate`. PUT requires the ticker to identify an open position in the owned portfolio. Price must be positive RUB and fit canonical Decimal storage. PUT and DELETE are naturally retry-safe current-state operations and do not use the financial command replay/idempotency-key subsystem.

`GET /api/v1/portfolios/{portfolioId}/positions` remains the authoritative projection and exposes either:

- `UNAVAILABLE / NO_APPROVED_MARKET_PRICE_SOURCE`; or
- `AVAILABLE / USER_SUPPLIED`.

## Financial formulas

For an eligible valued open position:

- `marketValue = quantity * manualPrice`
- `unrealizedGain = marketValue - acquisitionBasis`
- `unrealizedReturn = unrealizedGain / acquisitionBasis` when acquisition basis is non-zero; otherwise `null`
- `marketWeight = marketValue / valuedPositionsMarketValue`

All arithmetic uses the existing exact Decimal scale-8 Half-Even implementation. Browser code performs no financial arithmetic.

Portfolio valuation summary exposes:

- valued position count / total open position count;
- `PARTIAL` or `COMPLETE` coverage;
- valued positions market value;
- valued positions acquisition basis;
- total acquisition basis;
- unrealized gain over valued positions;
- canonical effective-ledger cash;
- `currentPortfolioValue = cash + valuedPositionsMarketValue` only when coverage is `COMPLETE`; otherwise `null`.

Partial coverage is never presented as complete portfolio value. Market allocation is allocation among valued positions while coverage is partial.

## Historical and date semantics

Current projection with omitted `asOfDate` may use the current manual valuation when its stored position generation matches the current open-position generation. Its explicit valuation BusinessDate is always displayed.

Historical projection may use a manual valuation only when:

1. stored position generation matches the reconstructed historical generation; and
2. stored manual valuation `asOfDate` exactly equals the requested historical BusinessDate.

There is no carry-back, carry-forward, interpolation, or use of today's price for a different historical date. Because Stage 3.76 stores current state only, overwriting a price does not preserve prior price history.

Stage 3.76 does not apply the wall clock as a hidden financial cutoff. A user may record a future BusinessDate; the system preserves that explicit user-supplied fact and does not present it as provider-confirmed market data.

## Security and isolation

Existing authenticated subject and portfolio anti-enumeration boundaries remain authoritative. PUT resolves the owned portfolio and then requires an open position matching the ticker. DELETE resolves the owned portfolio and removes only a valuation linked to that portfolio and ticker. No cross-user valuation lookup or mutation path is introduced.

## Frontend

Positions remains the user-visible surface. Current mode supports Add/Edit/Clear manual price, explicit price date, `Source: Manual`, market value, unrealized P/L, unrealized return, coverage and valued-position allocation. Historical mode remains read-only and never fabricates valuation for unmatched dates.

Frontend sends string Decimal/Money values and renders backend-derived financial values only.

## Explicit non-scope

No MOEX/broker/scraping source, live/automatic prices, FX engine, XIRR, TWR, nominal/real/inflation return, purchasing-power methodology, tax assistant, notifications, AI, mobile work, full dashboard redesign, provider activation, price-history warehouse, cache table, worker, Redis, Kafka, or Stage 3.77+ work.

Existing legacy `nominalReturnRate`, `xirr`, and `realReturn` placeholders are not populated from unrealized P/L.

## Verification plan

Required evidence before protected activation:

- simple valuation and exact Decimal/Half-Even vectors;
- multiple BUY/WAC and partial SELL;
- correction/reversal recomputation;
- zero acquisition-basis return `null`;
- overflow fail-closed behavior;
- unavailable/add/update/clear;
- partial/complete coverage and current value gating;
- close/reopen generation invalidation;
- historical exact-date eligibility and mismatch unavailability;
- explicit future/backdated manual BusinessDate preservation;
- subject/portfolio/ticker isolation;
- invalid/zero/negative price and currency mismatch;
- migration validator exact pair/hash/authority evidence;
- runtime readiness and privilege surface;
- frontend loading/error/add/edit/clear and positive/negative/zero P/L rendering;
- stale-request protection;
- exact-head CI and mandatory review gates.

## Governance state

Active lifecycle registries remain unchanged until Stage 3.76 is protected-merged. This dossier is part of the implementation candidate and is not canonical by publication alone.

The build environment could not clone GitHub because DNS resolution for `github.com` failed. No claim of local `pnpm verify`, Go tests, frontend build, or PostgreSQL integration execution is made. Deterministic candidate identity plus static/manual checks are used before publication; the full repository gates must run in GitHub CI on the exact published head.

Internal review evidence is intentionally withheld from this dossier until after fresh External published-head review, per `docs/REVIEW_WORKFLOW.md`.
