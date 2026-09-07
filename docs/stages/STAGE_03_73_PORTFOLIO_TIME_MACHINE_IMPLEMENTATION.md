# Stage 3.73 — Portfolio Time Machine / Historical Position View

| Field | Value |
| --- | --- |
| Stage | 3.73 |
| Status | MERGE-ACTIVATED IMPLEMENTATION — canonical only when this implementation is present on protected `develop` |
| Product scope | Historical open-position reconstruction by explicit BusinessDate |
| Canonical base | `develop@f76b03e01621114d1dbb96cd035e50de43c64687` |
| Runtime budget | 0 ₽ |
| External provider required | No |
| New API endpoint | No |
| New database table/cache | No |
| New financial engine | No |

## 1. Purpose

Stage 3.73 turns the already-approved Stage 3.72 `asOfDate` projection semantics into a user-visible product capability.

The user can switch the portfolio positions view between:

```text
CURRENT
```

and:

```text
AS OF YYYY-MM-DD
```

and see the exact position state reconstructed from the canonical immutable transaction ledger for that BusinessDate.

The feature is intentionally not a market-performance feature. It shows historical holdings and acquisition-basis accounting only.

## 2. Why this is the next product increment

Stage 3.72 already provides the hard financial primitive:

```text
Immutable transaction ledger
  ↓
portfolio-local ledger_sequence
  ↓
Stage 3.71 position.Apply / deterministic rebuild
  ↓
Stage 3.72 position projection
  ↓
GET /api/v1/portfolios/{portfolioId}/positions?asOfDate=YYYY-MM-DD
```

Creating a second historical endpoint, historical-position table, frontend WAC calculator, cache, worker, paid provider or new service would duplicate authority and create drift risk. Stage 3.73 therefore reuses the existing contract directly.

## 3. API and backend decision

No new API is introduced.

Stage 3.73 reuses:

```http
GET /api/v1/portfolios/{portfolioId}/positions?asOfDate=YYYY-MM-DD
```

The existing Stage 3.72 semantics remain authoritative:

- explicit `asOfDate` includes accepted position-affecting ledger rows with `tradeDate <= asOfDate`;
- `inputsAsOf` equals the requested BusinessDate for an explicit date;
- omitted `asOfDate` remains the existing current/latest-accepted-ledger projection and is not redefined as wall-clock “today”;
- same-BusinessDate replay remains ordered by canonical portfolio-local ledger sequence;
- backdated accepted trades alter subsequent reconstruction because immutable ledger truth is replayed again;
- accepted future-dated trades remain governed by the frozen Stage 3.72 omitted/explicit semantics;
- unsupported correction/reversal state continues to fail closed;
- closed positions are omitted for the selected projection date.

No Go production code, OpenAPI contract, SQL schema or migration is required for Stage 3.73.

## 4. Web UX

The portfolio positions panel adds two explicit modes:

```text
Current | Historical
```

Historical mode uses the native HTML date input and sends the selected BusinessDate to the existing typed `getPortfolioPositions` client.

The UI states are:

- Current;
- Historical with no date selected;
- loading historical projection;
- historical projection available;
- empty portfolio on the selected date;
- backend validation/error response;
- switch back to Current.

Historical mode displays the exact represented date in the positions heading and status text.

The word `Today` is intentionally not used. `Current` means the existing Stage 3.72 omitted-date ledger projection, not a statement about wall-clock date or market valuation.

## 5. Summary isolation

The existing portfolio summary is not an as-of Stage 3.73 projection. Showing its current metrics beside historical positions could imply unsupported historical market/cash/performance semantics.

Therefore Historical mode hides the current summary metrics and explicitly states that Stage 3.73 reconstructs positions only.

This prevents current summary values from being visually misread as historical values.

## 6. Financial authority

The frontend does not calculate:

- quantity;
- weighted-average cost;
- acquisition basis;
- acquisition-basis weight.

It does not infer historical values from the transaction list or legacy portfolio summary.

All displayed financial values remain backend-owned exact Decimal results from the Stage 3.71/3.72 engine and projection.

## 7. Canonical historical witness

Ledger:

```text
2026-01-10 BUY  100 SBER @ 250
2026-03-01 BUY  100 SBER @ 300
2026-06-01 SELL  50 SBER @ 350
```

Expected reconstruction:

```text
as of 2026-02-01
quantity = 100.00000000
WAC      = 250.00000000
basis    = 25000.00000000

as of 2026-04-01
quantity = 200.00000000
WAC      = 275.00000000
basis    = 55000.00000000

as of 2026-07-01
quantity = 150.00000000
WAC      = 275.00000000
basis    = 41250.00000000
```

These are backend-owned results and are covered by the Stage 3.73 PostgreSQL integration witness.

## 8. Historically-open / currently-closed invariant

A position that is closed now but was open on the requested historical date must appear in the historical projection.

Example:

```text
2026-01-10 BUY 5 SBER @ 100
2026-05-01 SELL 5 SBER @ 150
```

Expected:

```text
Current        → SBER absent
As of 2026-03 → SBER present, quantity 5, WAC 100, basis 500
```

This is a primary Stage 3.73 product invariant and has a dedicated PostgreSQL integration test.

## 9. Stale-request and identity protection

Historical requests have an independent generation guard based on the existing portfolio load-guard primitive.

A result is committed only when all of these still match:

- request generation;
- access token;
- principal identity;
- portfolio identity;
- requested view key (`AS_OF:YYYY-MM-DD`).

Therefore a late March response cannot overwrite a newer May selection, and a response from a prior principal or portfolio cannot be committed into the current view.

Accepted transaction/import mutations refresh both the canonical current projection and the selected historical projection so a newly accepted backdated transaction is reflected without frontend-generated historical state.

## 10. Empty, invalid and error semantics

- Historical mode with no selected date prompts for a BusinessDate and sends no historical request.
- A valid date before the first position-affecting transaction renders an honest empty historical portfolio.
- Invalid explicit dates remain backend-contract errors; the UI surfaces the positions error and does not fall back to summary/transaction-derived values.
- Repeating the same historical request against an unchanged ledger returns the same financial result.

## 11. Market-data boundary

Historical mode preserves the Stage 3.72 market state exactly:

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

The UI continues to say:

```text
Market valuation unavailable
```

No market-price provider, MOEX runtime activation, Corporate Actions source activation, external network data or paid API is introduced.

## 12. Test coverage

Stage 3.73 adds focused tests for:

- the three canonical historical WAC/basis vectors above;
- deterministic repeated projection for the same date;
- historically-open/currently-closed reconstruction;
- Current/Historical UI modes;
- native date input and exact date display;
- historical loading/empty/error states;
- independent generation + principal + portfolio + date stale-result protection;
- refresh after accepted transaction/import mutation;
- continued market-unavailable semantics;
- no frontend financial reconstruction and no new endpoint.

The existing Stage 3.72 suite remains authoritative for the broader replay matrix already implemented before Stage 3.73, including:

- BUY + BUY;
- partial/full SELL;
- close → reopen;
- multiple assets;
- STOCK + BOND;
- same-BusinessDate ordering;
- backdated BUY/SELL;
- fractional quantity and Half-Even arithmetic;
- explicit cutoff before first transaction;
- explicit and omitted future-date semantics;
- stable ordering;
- ownership/subject isolation;
- fail-closed ledger metadata;
- market-unavailable projection.

## 13. Explicit non-scope

Stage 3.73 does not add or authorize:

- market price/value;
- unrealized P/L;
- market or nominal return;
- XIRR;
- inflation/real-return history;
- portfolio-value history chart;
- acquisition-basis history chart in this first version;
- MOEX provider activation;
- Corporate Actions provider activation;
- transaction correction/reversal activation;
- imported SELL expansion;
- tax basis/tax lots;
- bond NKD/YTM/duration;
- notifications;
- AI;
- Redis/Kafka/workers/new servers/new SaaS.

## 14. Expected user-visible result

Definition of Done for the product experience:

1. open a portfolio;
2. see current positions;
3. choose Historical;
4. choose a BusinessDate;
5. see positions that actually existed on that date;
6. see exact historical quantity;
7. see exact historical WAC;
8. see exact historical acquisition basis;
9. see historical acquisition-basis allocation;
10. see a position that is currently closed when it was historically open;
11. switch back to Current;
12. never receive fabricated historical market valuation, P/L or return.

No later feature is authorized by Stage 3.73.
