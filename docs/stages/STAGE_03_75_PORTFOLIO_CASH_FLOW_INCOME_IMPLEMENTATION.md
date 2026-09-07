# Stage 3.75 — Portfolio Cash Flow & Income Truth

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION CANDIDATE — canonical only after reviewed protected merge |
| Canonical base | `develop@7043871e44eeb25e516a40b059f89d3cb25e03c1` |
| Budget | `0 RUB` |
| Runtime providers | none |
| New database migration | none |

## Purpose

Stage 3.75 closes the financial gap between the immutable transaction ledger and user-visible cash/income truth. It reuses Stage 3.74 effective-ledger correction/reversal semantics and adds one backend-owned, rebuildable aggregate projection for deposits, withdrawals, trades, income, fees and recorded taxes.

## Canonical methodology

For each effective ledger row, transaction-level `commission` is counted once in `fees` and transaction-level `tax` is counted once in `taxes`. Then the row gross amount is categorized exactly once:

- `DEPOSIT` -> `deposits`
- `WITHDRAWAL` -> `withdrawals`
- `BUY` -> `buyOutflows`
- `SELL` -> `sellInflows`
- `DIVIDEND` -> `dividendsGross`
- `COUPON` -> `couponsGross`
- standalone `FEE` gross -> `fees`
- standalone `TAX` gross -> `taxes`

Derived identities are:

```text
netExternalFlow = deposits - withdrawals
netInvestmentIncome = Σ(DIVIDEND gross - that row commission - that row tax)
                    + Σ(COUPON gross - that row commission - that row tax)
netCashFlow = deposits - withdrawals - buyOutflows + sellInflows
              + dividendsGross + couponsGross - fees - taxes
```

`netInvestmentIncome` therefore attributes deductions only when they are recorded on the same `DIVIDEND` or `COUPON` row. BUY/SELL commissions and standalone `FEE`/`TAX` rows remain real cash outflows and are included in `fees`, `taxes`, and `netCashFlow`, but they are not silently attributed to investment income without a reviewed linkage model. This metric is not investment return, market performance, or tax advice.

## Date and correction semantics

- grouping uses canonical `tradeDate`, never `created_at`;
- periods are backend-produced `YYYY-MM` buckets in ascending order;
- `fromDate` and `toDate` are optional BusinessDate filters;
- explicit `toDate` uses Stage 3.74 historical effective-ledger semantics, including reversal `effectiveDate`;
- corrections replace earlier revisions economically instead of being added to them;
- reversals remove the target economic effect when effective under the requested cutoff;
- omitted `toDate` preserves Stage 3.74 current effective-ledger semantics and therefore includes accepted future-dated rows rather than silently applying a wall-clock cutoff.

## API

```http
GET /api/v1/portfolios/{portfolioId}/cash-flow?fromDate=YYYY-MM-DD&toDate=YYYY-MM-DD
```

The response is bounded aggregate data: totals plus monthly periods. The browser does not fetch the full ledger to calculate financial truth.

## Summary and snapshot integration

The same cash-flow arithmetic now drives snapshot cash value. Stage 3.71 invested-capital semantics remain unchanged: BUY gross + BUY commission + BUY tax. `PortfolioSummary.dividendsReceived` and `couponsReceived` are populated from effective-ledger gross dividend/coupon truth instead of hardcoded zero placeholders; associated deductions remain separately represented in cash-flow fees/taxes and cash value. Summary reads snapshot and income truth under one repeatable-read transaction.

## Manual transaction entry

The existing Web form now exposes the already-frozen `DIVIDEND`, `COUPON`, `FEE`, and `TAX` types. No second form or new transaction type was created. Income follows the frozen ticker / optional quantity / gross amount contract. Pure standalone FEE/TAX UI writes their gross amount with null ticker/quantity/unitPrice and zero nested commission/tax; nested commission/tax remain available on trade/income rows.

The component contract explicitly distinguishes canonical zero from fixture-derived business values: `0.00000000` is permitted only where the request contract requires zero nested commission/tax for standalone expense or pure cash-flow rows. This keeps the no-fixture guard intact without treating a required financial zero as fake data.

## UI boundary

The portfolio page renders backend totals and monthly buckets only. BUY outflows and SELL inflows are displayed alongside deposits, withdrawals, income, fees and taxes so the user can reconcile backend `netCashFlow` without hidden frontend arithmetic. Gross dividend/coupon wording is explicit, and the current no-date-filter view states that accepted future-dated ledger rows are included. Legacy nominal-return/XIRR/real-gain/purchasing-power cards are not presented as investment performance in this stage because approved market valuation and return methodology remain unavailable.

## Non-scope

No market prices/value, unrealized P/L, XIRR/return methodology, inflation, provider activation, broker sync, tax advice/declaration, bond YTM/duration/NKD, notification, AI, Redis, Kafka, worker, paid service, new transaction table, aggregate cache table, or production rollout is authorized.

## Verification

The implementation includes backend integration vectors for mixed ledgers, standalone and nested fees/taxes, income-only net-investment-income attribution, monthly/range aggregation, summary cash/dividend/coupon truth, correction, reversal effective dates, backdated income, empty periods and subject isolation; HTTP contract witnesses; frontend manual-type and backend-owned-math contract tests; and the repository's full protected CI matrix.
