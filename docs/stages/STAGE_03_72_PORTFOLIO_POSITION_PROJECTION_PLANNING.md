# Stage 3.72 — Portfolio Position Projection / Honest Market-Unavailable Semantics Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-72-PORTFOLIO-POSITION-PROJECTION-PLAN |
| Version | 0.1.1-candidate |
| Status | MERGE-ACTIVATED PLANNING DECISION — NON-NORMATIVE BEFORE PROTECTED MERGE / CANONICAL ONLY AFTER REQUIRED GATES + SQUASH MERGE / NO IMPLEMENTATION AUTHORIZATION |
| Canonical planning base | `develop@69be7fe12f53d016019c8a29297d7c656830670d` |
| Protected-base tree | `a595a8a90cbf5319fcf33f93bb81e5bae083e976` |
| Architecture decision | No new ADR proposed; this scope implements the public projection explicitly deferred by ADR-009 without changing ledger ordering or WAC methodology |
| Date | 2026-09-06 |

## 1. Purpose

Freeze the smallest correct product/API/UI boundary that exposes the deterministic open positions already computed from the Stage 3.71 immutable ledger and WAC engine.

The user-facing responsibility is:

> Show the portfolio's canonical open-position ledger projection and how much remaining trade-price acquisition basis belongs to each open position, while representing every market-derived valuation fact as explicitly unavailable.

Stage 3.72 is not a market-price feature and does not redesign returns.

## 2. Why this is the next dependency

Stage 3.71 already provides the authoritative financial mechanics:

```text
immutable transaction ledger
→ tradeDate ASC
→ ledgerSequence ASC
→ canonical position engine
→ quantity
→ weightedAverageCost
→ acquisitionBasis
```

ADR-009 and Stage 3.70 deliberately kept public positions off because the existing `PortfolioPosition` / `analytics.snapshot_positions` model requires market-derived values that cannot be fabricated.

Stage 3.72 closes only that presentation/read-model gap.

## 3. Governance classification


Protected merge of this planning decision makes the plan canonical only. It does not itself authorize Stage 3.72 runtime implementation.

## 4. Frozen responsibility

One responsibility:

> Deterministically expose open portfolio positions from the canonical Stage 3.71 ledger/WAC engine, including quantity, WAC, remaining trade-price acquisition basis, and acquisition-basis allocation, while making unavailable market valuation explicit and non-fabricated.

The implementation must reuse the canonical `position` engine. It must not introduce a second WAC/cost-basis algorithm in HTTP, service, PostgreSQL, Web, or mobile code.

## 5. New additive API resource

Planned endpoint:

```http
GET /api/v1/portfolios/{portfolioId}/positions
```

Optional query:

```text
asOfDate=YYYY-MM-DD
```

The endpoint is additive. It does **not** activate or reshape the existing `PortfolioSummary.positions` contract.

Expected standard API envelope:

```json
{
  "data": {
    "items": [],
    "totalAcquisitionBasis": {
      "amount": "0.00000000",
      "currency": "RUB"
    },
    "calculation": {
      "methodologyVersion": "portfolio-position-projection-v1",
      "inputsAsOf": null
    }
  }
}
```

Planned operation identifier:

```text
getPortfolioPositions
```

Planned responses:

```text
200 success
400 invalid asOfDate
401 unauthenticated
404 portfolio not found / not accessible under the existing subject-isolation semantics
```

Stage 3.72 does not add a new 429 contract or a new error-envelope shape.

The v1 positions resource is not paginated. Its public cardinality is one item per open canonical asset, not one item per ledger entry. If measured portfolio/catalog cardinality later proves pagination necessary, that is a separately reviewed contract change.

## 6. BusinessDate / asOfDate semantics

`asOfDate` reuses the existing canonical `BusinessDate` grammar and server-side validation.

When `asOfDate` is supplied:

```text
include position-affecting ledger entries where tradeDate <= asOfDate
calculation.inputsAsOf = requested asOfDate
```

When `asOfDate` is omitted:

```text
include all accepted supported position-affecting ledger entries
calculation.inputsAsOf = maximum included tradeDate
```

This omitted-date mode means **latest accepted ledger projection**, not a wall-clock claim of "today". If the existing transaction contract permits and contains a future-dated accepted trade, omitted-date projection includes it. Stage 3.72 does not redesign transaction-date admission. The UI must expose `calculation.inputsAsOf` when non-null so the user can see the exact BusinessDate represented.

If no position-affecting ledger row exists and `asOfDate` is omitted:

```text
items = []
totalAcquisitionBasis = 0 RUB
calculation.inputsAsOf = null
```

If an explicit `asOfDate` is supplied but no position-affecting row exists on/before it:

```text
items = []
totalAcquisitionBasis = 0 RUB
calculation.inputsAsOf = requested asOfDate
```

No system timestamp, UUID order, asset map iteration, or frontend clock may become a hidden financial ordering input.

Stage 3.72 does not add date navigation/history charts. Historical Position Projection / Portfolio Time Machine remains a later product scope.

## 7. Position projection contract

Planned item shape for a positive aggregate acquisition-basis denominator:

```json
{
  "ticker": "SBER",
  "assetType": "STOCK",
  "quantity": "150.00000000",
  "weightedAverageCost": {
    "amount": "275.00000000",
    "currency": "RUB"
  },
  "acquisitionBasis": {
    "amount": "41250.00000000",
    "currency": "RUB"
  },
  "acquisitionBasisWeight": "0.34210000",
  "marketValuation": {
    "status": "UNAVAILABLE",
    "reason": "NO_APPROVED_MARKET_PRICE_SOURCE",
    "marketPrice": null,
    "marketValue": null,
    "unrealizedGain": null,
    "marketWeight": null,
    "provider": null,
    "asOf": null
  }
}
```

Canonical API `assetType` values remain:

```text
STOCK
BOND
```

All money in this Stage 3.72 projection is RUB.

Closed positions are excluded.

The implementation must not filter an owned open position merely because its catalog `lifecycle_status` later becomes inactive.

## 8. Existing PortfolioPosition remains untouched

The existing canonical `PortfolioPosition` requires:

```text
ticker
assetType
quantity
weightedAverageCost
marketPrice
marketValue
unrealizedGain
weight
```

Stage 3.72 must not make those market fields optional, fill them with zero, or reinterpret `weight`.

Therefore:

```text
PortfolioSummary.positions
```

remains unchanged/empty under the current Stage 3.71 summary behavior.

Stage 3.72 introduces a distinct projection DTO/resource instead of overloading the old market-valued position model.

## 9. Acquisition basis semantics

Per ADR-009 / Stage 3.71:

```text
acquisitionBasis = Round8HalfEven(quantity × weightedAverageCost)
```

It means remaining RUB **trade-price acquisition basis** only.

It is not:

```text
tax basis
market value
cash invested net of all flows
commission-inclusive basis
inflation-adjusted basis
FX-adjusted basis
bond NKD-aware basis
```

Commission, recorded tax, FX, inflation, and bond accrued coupon/NKD remain outside this projection.

## 10. Total acquisition basis

The response exposes the denominator explicitly:

```text
totalAcquisitionBasis
=
Σ acquisitionBasis of every open STOCK/BOND position
```

The total is calculated server-side.

The frontend must not recompute financial totals from display-rounded values.

Every intermediate total remains exact Decimal scale 8 / precision 28 / Half Even and must satisfy the existing storage-compatible numeric bound. Overflow fails closed.

## 11. Acquisition-basis allocation

For each open position:

```text
acquisitionBasisWeight
=
Round8HalfEven(
  position.acquisitionBasis / totalAcquisitionBasis
)
```

Properties:

- if `totalAcquisitionBasis > 0`, every open item receives an exact scale-8 Decimal weight;
- if open items exist but `totalAcquisitionBasis = 0.00000000` because all positive trade-price bases rounded to zero at canonical scale 8, `acquisitionBasisWeight = null` for those items; the API must not fabricate `0` for the undefined `0/0` allocation;
- if no open item exists, no division occurs;
- exact server-side Decimal arithmetic only;
- range `0.00000000..1.00000000`;
- this is **not** market portfolio weight;
- API/UI terminology must preserve that distinction.

The planned OpenAPI type for `acquisitionBasisWeight` is `Decimal | null`; `null` is allowed only for the zero-denominator case above.

Allowed UI labels:

```text
Доля вложенного капитала
Доля по стоимости приобретения
Acquisition basis weight
```

Forbidden UI/API relabelling:

```text
Portfolio weight
Market weight
Asset allocation by market value
```

Independent scale-8 rounding means the sum of item weights is not force-normalized to exactly `1.00000000`.

The implementation must not alter the final item merely to compensate for prior rounding.

## 12. Honest market-unavailable model

Stage 3.72 runtime may emit only:

```text
marketValuation.status = UNAVAILABLE
marketValuation.reason = NO_APPROVED_MARKET_PRICE_SOURCE
```

and:

```text
marketPrice = null
marketValue = null
unrealizedGain = null
marketWeight = null
provider = null
asOf = null
```

Forbidden substitutions:

```text
marketPrice = weightedAverageCost
marketPrice = 0
marketValue = acquisitionBasis
marketValue = 0
unrealizedGain = 0
marketWeight = acquisitionBasisWeight
provider = example/provider placeholder
```

No `AVAILABLE` market-valuation runtime state is authorized by Stage 3.72.

A later approved market-price source stage may extend the market-valuation contract, but Stage 3.72 must not pre-activate provider semantics or production source identifiers.

## 13. Deterministic ordering

Public `items` order is:

```text
ticker ASC
```

The canonical asset table already requires unique ticker identity for the current MVP catalog.

The response must not expose incidental order from:

```text
Go map iteration
asset UUID
database physical order
createdAt
ledgerSequence across different assets
```

Ledger order remains `tradeDate ASC, ledgerSequence ASC` **within each asset replay**.

## 14. PostgreSQL / engine design boundary

No new table is planned.

No change to `analytics.snapshot_positions` is planned.

No persisted position cache is planned.

Expected read path:

```text
authenticated subject / portfolio authorization
→ one portfolio-scoped ledger read for BUY/SELL rows
→ canonical asset metadata join
→ exact persisted ledgerSequence validation
→ existing correction/reversal fail-closed checks
→ existing position.Apply / position.Rebuild semantics
→ open position states
→ server-side total + acquisitionBasisWeight
→ deterministic ticker sort
→ API DTO
```

Stage 3.72 should extract/reuse a common internal ledger-to-position projection seam where practical so Stage 3.71 snapshot aggregation and Stage 3.72 public projection do not drift into separate financial algorithms.

Any refactor of the Stage 3.71 read/rebuild helper must preserve exact existing financial results and Stage 3.71 tests.

## 15. Concurrency and read consistency

The projection is a read model over immutable accepted ledger rows.

A request must observe one internally consistent ledger view for its returned items/total/weights. It must not combine positions from one database view with totals from a later view.

Implementation may use one SQL statement plus in-memory deterministic replay, or an equivalent read-only transaction/snapshot boundary.

A concurrent write may be visible or not visible depending on the read snapshot, but the response must correspond to one coherent accepted-ledger state.

## 16. Security / ownership boundary

The endpoint is subject-owned exactly like existing portfolio reads.

Mandatory properties:

```text
subject A cannot read subject B portfolio positions
portfolio not found / inaccessible behavior preserves the existing anti-enumeration boundary
no asset/ledger row leaks before authorization
```

No public/anonymous position endpoint is introduced.

## 17. Empty / unsupported / fail-closed behavior

Empty portfolio:

```text
200
items = []
totalAcquisitionBasis = 0 RUB
```

Full close:

```text
closed item omitted
```

Close → reopen:

```text
later BUY starts the new canonical WAC state already defined by ADR-009
```

If the ledger contains:

```text
missing/non-positive ledgerSequence
unsupported correction/reversal state
unsupported asset type
invalid Decimal
derived overflow
historical oversell
```

the endpoint fails closed through the existing mapped error discipline. It must not return a partial position set.

Stage 3.72 does not define correction/reversal runtime semantics.

## 18. OpenAPI surfaces planned for implementation

Implementation is expected to add only the additive contract surfaces required by the new read resource, for example:

```text
openapi/openapi.yaml
openapi/components/schemas.yaml
openapi/components/responses.yaml
openapi/examples/...
backend-go/cmd/validate-openapi/main.go
```

Planned schema vocabulary:

```text
PortfolioPositionProjection
PortfolioPositionsProjection
MarketValuationUnavailable
PortfolioPositionsResponse
```


The Stage 2 frozen baseline remains historical authority; implementation records the additive Stage 3.72 contract rather than silently rewriting old architectural history.

## 19. Go application surfaces planned for implementation

Expected areas:

```text
verticalslice store/service read contract
PostgreSQL projection/replay helper
HTTP handler + DTO mapper
route wiring
tests
```

Financial arithmetic remains backend-owned.

No WAC, acquisition-basis, or weight calculation is permitted in Web/mobile code.

## 20. Web product boundary

The Portfolio detail page adds a Positions block backed only by the new Go API. The request participates in the existing authenticated portfolio load/principal-generation guard so stale responses from an earlier token/principal/portfolio load cannot overwrite the current view. A positions-specific API failure must render an explicit positions warning without fabricating values from summary or transaction data.

Desktop target:

| Asset | Quantity | Average acquisition price | Acquisition basis | Acquisition-basis weight |
| --- | ---: | ---: | ---: | ---: |
| SBER | 150 | 275.00 ₽ | 41,250.00 ₽ | 34.2% |

Mobile target: cards containing the same canonical values.

Each item shows an explicit unavailable state such as:

```text
Market valuation unavailable
```

The UI must not show:

```text
current price
market value
unrealized profit/loss
market portfolio weight
return percentage derived from missing market value
```

The existing summary return placeholders remain outside this feature.

## 21. Mandatory canonical financial vectors

Stage 3.72 must reuse Stage 3.71 financial vectors.

Core vector:

```text
BUY 100 @ 250
BUY 100 @ 300

quantity = 200.00000000
WAC = 275.00000000
acquisitionBasis = 55000.00000000
```

After:

```text
SELL 50
```

must return:

```text
quantity = 150.00000000
WAC = 275.00000000
acquisitionBasis = 41250.00000000
```

Fractional authoritative-WAC witness remains:

```text
quantity = 0.20000000
WAC = 275.12345678
acquisitionBasis = 55.02469136
```

The implementation must not re-derive WAC as `acquisitionBasis / quantity`.

## 22. Mandatory test matrix

Minimum implementation evidence:

```text
single BUY
BUY + BUY
BUY + partial SELL
BUY + full SELL
close → reopen
multiple assets
multiple STOCK
multiple BOND
mixed STOCK + BOND
backdated BUY
backdated SELL
same BusinessDate ordering
fractional quantity
Half-Even WAC witness
maximum Decimal
derived basis overflow
aggregate total overflow
open position whose rounded acquisitionBasis is zero
nonempty open positions with zero total basis -> nullable acquisitionBasisWeight
zero closed positions excluded
all positions closed
empty portfolio
rebuild twice identical financial projection
stable ticker ordering
explicit asOfDate cutoff
future/later transaction excluded by earlier asOfDate
omitted asOfDate resolves maximum included tradeDate
acquisitionBasisWeight denominator correctness
acquisitionBasisWeight Half-Even rounding
rounded weights are not force-normalized
subject isolation
portfolio ownership / anti-enumeration
market status always UNAVAILABLE
market reason always NO_APPROVED_MARKET_PRICE_SOURCE
all market-derived fields null
no fabricated market valuation
no provider/source lookup
existing PortfolioSummary.positions remains unchanged/empty
analytics.snapshot_positions untouched
Stage 3.71 snapshot totals remain financially identical after any helper extraction
OpenAPI schema/example validator
typed frontend client
frontend loading/error/empty states
frontend principal/token/portfolio stale-load guard
positions failure does not fabricate/fallback from summary
desktop table
mobile/card presentation
reload after accepted transaction
```

PostgreSQL integration evidence is mandatory for ownership, deterministic replay, as-of cutoff, consistency under concurrent write/read, fail-closed ledger metadata, and Stage 3.71 regression preservation.

## 23. Performance and cost boundary

Budget:

```text
external API = 0
paid provider = 0
new cloud service = 0
Redis/cache = 0
worker/cron = 0
new table = 0
 = 0
```

Initial implementation may replay the portfolio's accepted position-affecting ledger in-process because correctness and zero-cost scope take priority over speculative persistence.

No database migration/index is pre-authorized by this plan.


## 24. Explicit exclusions

```text
market-price provider
market price
market value
unrealized P/L
market weight
AVAILABLE market valuation
XIRR
nominal return redesign
real return
inflation
purchasing power redesign
realized P/L
performance attribution
concentration analytics
tax basis
FIFO/tax lots
commission-inclusive WAC
bond NKD
YTM
duration
bond amortization semantics
Corporate Actions provider / Feature 3D
imported SELL activation
transaction correction
transaction reversal
new position table
snapshot_positions redesign
Redis
cache/workers/cron
notifications

mobile-native implementation
```

## 25. Stop conditions

Stop Stage 3.72 implementation and return to planning if:

- the implementation would require fabricating any market-derived fact;
- the existing position engine cannot be reused as the financial authority;
- the endpoint would require weakening ADR-009 ordering/WAC semantics;
- subject isolation cannot be preserved;
- the read cannot return one coherent ledger projection;
- a database migration/new table/provider becomes necessary;
- Stage 3.71 regression vectors change;
- scope expands into returns, P/L, market data, tax, correction/reversal, imported SELL, Corporate Actions, notifications, or .

## 26. Rollback

Before runtime implementation, rollback is reversion of this planning decision.

After a future Stage 3.72 implementation, the new endpoint/UI is a derived read projection only. Rollback must not rewrite/delete ledger entries, ledgerSequence, WAC history, snapshots, or source evidence.

If public projection must be disabled, remove/disable the additive endpoint/UI surface while preserving canonical Stage 3.71 financial history.

## 27. Planning acceptance and implementation gate


It becomes canonical only after:

2. explicit human commit/push authorization;
3. Draft PR publication;
4. all required exact-head CI succeeds;
6. required Internal evidence is published only after the External verdict;
7. exact-head evidence-only CI and single-context verification pass;
8. Principal Architect explicitly accepts the Stage 3.72 planning decision and authorizes squash merge;
9. the exact accepted planning tree is squash-merged into protected `develop`.

Canonical record: PR #142, PR #143, PR #144; commit(s) `69be7fe12f53d016019c8a29297d7c656830670d`.

Even after planning activation:

```text
STAGE_03_72_IMPLEMENTATION_AUTHORIZATION = NOT_GRANTED
```

Runtime implementation requires a new, separate explicit merge gate.

## 29. Post-disposition current-base refresh state

Canonical record: PR #142, PR #143, PR #144; commit(s) `db589eb074468f352e61b82bbad8f33307b3fc89`.

Current refresh identity:

```text
PREVIOUS_EVIDENCE_HEAD=c4fbceb722946ceb364ba8c7f41e69848179fe7c
CURRENT_CANONICAL_BASE=69be7fe12f53d016019c8a29297d7c656830670d
CURRENT_CANONICAL_BASE_TREE=a595a8a90cbf5319fcf33f93bb81e5bae083e976
STAGE-03-72-GOV-01=DISPOSITIONED
HISTORICAL_COMPLIANCE=NONCOMPLIANT_PRESERVED
RESIDUAL_GOVERNANCE_RISK=ACCEPTED
DISPOSITION_EFFECTIVE=YES
DISPOSITION_PR=143
DISPOSITION_CLOSURE_PR=144
```

This refresh preserves the four-file planning scope and all frozen Stage 3.72 financial/API/product semantics. It reconciles only current-base and governance/lifecycle metadata required by the already-canonical PR #143/#144 state. It does not activate `PortfolioSummary.positions`, a market provider, market valuation, XIRR/returns, imported SELL, correction/reversal, tax basis, Corporate Actions, notifications, a new table, Redis, cache, workers, or production rollout.


```text
PR_142_READY_AUTHORIZATION=NOT_GRANTED
PLANNING_ACCEPTANCE=NOT_GRANTED
SQUASH_MERGE_AUTHORIZATION=NOT_GRANTED
STAGE_03_72_IMPLEMENTATION_AUTHORIZATION=NOT_GRANTED
```
