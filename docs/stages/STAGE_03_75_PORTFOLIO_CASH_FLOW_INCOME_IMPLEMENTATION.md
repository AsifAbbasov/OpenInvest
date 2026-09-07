# Stage 3.75 — Portfolio Cash Flow & Income Truth

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION CANDIDATE / NON-CANONICAL |
| Canonical base | `develop@7043871e44eeb25e516a40b059f89d3cb25e03c1` |
| Budget | `0 RUB` |
| Runtime providers | none |
| New database migration | none |
| Review path | development path under `docs/REVIEW_WORKFLOW.md` |

## Purpose

Stage 3.75 closes the financial gap between the immutable transaction ledger and user-visible cash/income truth. It reuses Stage 3.74 effective-ledger correction/reversal semantics and adds one backend-owned, rebuildable aggregate projection for deposits, withdrawals, trades, income, fees and recorded taxes.

## Canonical methodology

For each effective-ledger row, transaction-level `commission` is counted once in `fees` and transaction-level `tax` is counted once in `taxes`. The row gross amount is then categorized exactly once:

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

The eight component aggregates are non-negative magnitudes. `netExternalFlow`, `netInvestmentIncome`, and `netCashFlow` are signed and may legitimately be negative.

`netInvestmentIncome` attributes deductions only when they are recorded on the same `DIVIDEND` or `COUPON` row. BUY/SELL commissions and standalone `FEE`/`TAX` rows remain real cash outflows and are included in `fees`, `taxes`, and `netCashFlow`, but are not silently attributed to investment income without a reviewed linkage model. This metric is not investment return, market performance, or tax advice.

## Date and correction semantics

- grouping uses canonical `tradeDate`, never `created_at`;
- periods are backend-produced `YYYY-MM` buckets in ascending order;
- `fromDate` and `toDate` are optional BusinessDate filters;
- explicit `toDate` uses Stage 3.74 historical effective-ledger semantics, including reversal `effectiveDate`;
- corrections replace earlier revisions economically instead of being added to them;
- reversals remove the target economic effect when effective under the requested cutoff;
- omitted `toDate` preserves Stage 3.74 current effective-ledger semantics, including accepted future-dated rows; no wall-clock cutoff is introduced;
- with explicit `toDate`, `calculation.inputsAsOf` is that requested cutoff;
- with omitted `toDate`, `calculation.inputsAsOf` is the maximum included effective-ledger `tradeDate`, or `null` when no row is included.

## API

```http
GET /api/v1/portfolios/{portfolioId}/cash-flow?fromDate=YYYY-MM-DD&toDate=YYYY-MM-DD
```

The response contains bounded aggregate data: totals plus monthly periods. Financial aggregation remains backend-owned; the browser does not fetch the full ledger to calculate financial truth.

## Summary and snapshot integration

The same cash-flow arithmetic drives snapshot cash value. Stage 3.71 invested-capital semantics remain unchanged: BUY gross + BUY commission + BUY tax. `PortfolioSummary.dividendsReceived` and `couponsReceived` are gross recorded amounts from the canonical effective ledger under the summary cutoff/as-of boundary. Associated commission/tax deductions are not subtracted from those fields; they are represented separately through cash-flow fees/taxes. These summary fields are not net receipts and are not performance metrics. Summary reads snapshot and income truth under one repeatable-read transaction.

## Manual transaction entry

The existing Web form exposes the already-frozen `DIVIDEND`, `COUPON`, `FEE`, and `TAX` transaction types in addition to the existing trade and cash-flow types. No second form and no new transaction type are introduced. Income uses ticker, optional quantity and gross amount according to the frozen contract. Standalone FEE/TAX UI writes gross amount with null ticker/quantity/unitPrice and zero nested commission/tax, preventing nested deductions from being double-counted.

## UI boundary

The portfolio page renders backend totals and monthly buckets only. BUY outflows and SELL inflows are displayed alongside deposits, withdrawals, gross dividend/coupon income, fees and taxes so backend `netCashFlow` is reconcilable without frontend arithmetic. Summary labels are `Gross dividends recorded` and `Gross coupons recorded`. The current no-date-filter view states that accepted future-dated ledger rows are included. Legacy nominal-return/XIRR/real-gain/purchasing-power cards are not presented as investment performance because approved market valuation and return methodology remain unavailable.

## Contract hardening

`PortfolioCashFlowTotals` uses `NonNegativeMoney` for `deposits`, `withdrawals`, `buyOutflows`, `sellInflows`, `dividendsGross`, `couponsGross`, `fees`, and `taxes`. Only `netExternalFlow`, `netInvestmentIncome`, and `netCashFlow` use signed `Money`. The canonical OpenAPI validator rejects negative component magnitudes and positively verifies that negative net values remain valid.

## Replacement-governance provenance

Historical Draft PR #152 is preserved as non-canonical review/remediation evidence. Its original pre-publication Internal Review evidence could not be recovered, so this candidate does **not** claim retroactive compliance and does **not** inherit PR #152's review verdicts as approval of this candidate.

This replacement candidate restarts the mandatory development-path sequence from the same canonical base. It must receive a fresh read-only pre-publication Internal review of the complete changed-file set. The resulting Internal report remains withheld from PR/repository evidence until after a genuinely fresh External published-head verdict, exactly as required by `docs/REVIEW_WORKFLOW.md`.

Active canonical lifecycle registries (`README.md`, `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, `DOCUMENT_INDEX.md`, `IMPLEMENTATION_LOG.md`, `VERSION_MATRIX.md`, and `OPEN_QUESTIONS.md`) deliberately remain unchanged in this implementation candidate. They continue to describe protected `develop` through Stage 3.74 until Stage 3.75 is actually accepted and merged. This avoids publishing candidate state as canonical registry state and keeps the implementation review set within the default file budget. Post-merge lifecycle synchronization is a separate governance/closure action.

## File-budget discipline

The replacement implementation candidate contains 24 changed files: 23 development/test/contract files plus this stage dossier. It is therefore within the default `<=25 changed files` development review budget and requires no file-count exception.

## Non-scope

No market prices/value, unrealized P/L, XIRR/return methodology, inflation, provider activation, broker-sync expansion, tax advice/declaration, bond YTM/duration/NKD, notifications, AI, Redis, Kafka, workers, paid services, new transaction table, aggregate cache table, Docker/container restructuring, new development orchestration, or Stage 3.76+ work is authorized.

## Verification expectations

The implementation includes backend integration vectors for mixed ledgers, standalone and nested fees/taxes, income-only net-investment-income attribution, monthly/range aggregation, summary cash/dividend/coupon truth, correction, reversal effective dates, backdated income, empty periods and subject isolation; HTTP contract witnesses; frontend manual-type and backend-owned-math contract tests; OpenAPI financial guard vectors; and the repository's full required CI matrix after publication.

## Internal Review evidence — published after External verdict

This section publishes the genuine replacement-candidate Internal Review record only after the fresh External published-head review of PR #153 returned `APPROVED` in PR conversation comment `5574951592`. It is not a reconstruction of the missing historical PR #152 review.

### Reviewed candidate identity

- review phase: `Internal Review Agent`
- canonical base: `7043871e44eeb25e516a40b059f89d3cb25e03c1`
- frozen unpublished candidate tree: `a25d940d729e95072c1bdbc0179aca5c24b85bad`
- candidate manifest SHA-256: `e2f4e31d67347a4f78f74a9bd78bb71009eec76c7c2bba8d4675b7225fd1aea8`
- candidate state during review: no branch ref, commit, or PR published for this replacement candidate
- changed-file count reviewed: `24 / 24`
- sampling: `NO`

### Files reviewed

1. `backend-go/cmd/validate-openapi/main.go`
2. `backend-go/internal/httpapi/cash_flow.go`
3. `backend-go/internal/httpapi/replay_app.go`
4. `backend-go/internal/httpapi/routes.go`
5. `backend-go/internal/httpapi/stage_03_75_cash_flow_test.go`
6. `backend-go/internal/postgres/effective_ledger.go`
7. `backend-go/internal/postgres/stage_03_71_summary.go`
8. `backend-go/internal/postgres/stage_03_75_cash_flow.go`
9. `backend-go/internal/postgres/stage_03_75_cash_flow_integration_test.go`
10. `backend-go/internal/verticalslice/service.go`
11. `backend-go/internal/verticalslice/stage_03_71_store_adapter.go`
12. `backend-go/internal/verticalslice/stage_03_75_cash_flow.go`
13. `docs/stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md`
14. `frontend-next/src/common/api/openinvest.ts`
15. `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx`
16. `frontend-next/src/features/portfolio/components/CashFlowIncomeBlock.tsx`
17. `frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx`
18. `frontend-next/tests/add-transaction-form.component.test.tsx`
19. `frontend-next/tests/stage-03-72-positions-component-contract.test.mjs`
20. `frontend-next/tests/stage-03-75-cash-flow-income.test.mjs`
21. `openapi/components/responses.yaml`
22. `openapi/components/schemas.yaml`
23. `openapi/examples/cash-flow.json`
24. `openapi/openapi.yaml`

### Internal review result

- financial formulas / Decimal ownership: `PASS`
- OpenAPI/runtime consistency: `PASS`
- security / portfolio ownership boundary: `PASS`
- Stage 3.74 effective-ledger reuse and date semantics: `PASS`
- frontend financial-calculation boundary: `PASS — backend-owned`
- DDD / SOLID / DRY / KISS / YAGNI / scope: `PASS`
- file budget: `PASS — 24 <= 25`
- blocking findings: `0`
- resolved findings in this final frozen candidate: `none outstanding`
- remaining non-blocking notes requiring remediation: `0`
- reviewer made edits: `NO`

Shell-based full-project local execution was unavailable in the connector-only review environment, and no `pnpm verify` or equivalent local result was claimed. The candidate was frozen and reviewed by exact tree/file identity; authoritative full-project verification was explicitly deferred to the exact published replacement head. That published head `f97f44016d22e783980c47bbde4799a3124623f4` subsequently passed CI #459 / run `34155729975` 10/10 before the External verdict. Historical PR #152 review verdicts are not used as approval of this replacement candidate.

**Internal VERDICT: APPROVED**

This dossier grants no Ready, merge, branch-deletion, or Stage 3.76+ authorization.
