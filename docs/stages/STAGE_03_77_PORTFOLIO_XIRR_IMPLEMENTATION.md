# Stage 3.77 — Portfolio Money-Weighted Return / XIRR Implementation

## Status

Stage 3.77 is published as Draft PR #157 against the immutable Stage 3.76 canonical base. Ready-for-review, merge, and Stage 3.78+ remain unauthorized. Full canonical closure is not claimed by this document.

Internal Review Evidence: **WITHHELD — publish only after External published-head approval in the evidence-only follow-up phase**.

The previously published head `88b7e08592f694d5f56c9ba275a34474f7cb677f` completed the protected ten-job CI matrix 10/10 GREEN. A fresh External review of those exact published bytes returned `REQUEST CHANGES` for two issues: duplicate `asOfDate` query parameters were not rejected at the Stage 3.77 HTTP boundary, and this dossier still described the earlier pre-publication 22-file candidate instead of the actual 25-file Draft PR. The remediation changes existing PR files only and therefore does not expand the 25-file scope. Any new published head requires a fresh exact-head CI run before External approval can be granted.

## Canonical base

- repository: `AsifAbbasov/OpenInvest`
- protected base branch: `develop`
- base commit: `12b41d2bb03999e3adefacf1ebb408fb30d732b8`
- base tree: `8e8c183912cf87fe0e68f3d73c1032aec158a409`
- base state: Stage 3.76 fully closed/canonical before Stage 3.77 work

## Scope

Stage 3.77 implements one backend-owned money-weighted portfolio return projection using XIRR semantics and adds:

`GET /api/v1/portfolios/{portfolioId}/returns?asOfDate=YYYY-MM-DD`

The endpoint is authenticated. `asOfDate` is mandatory, explicit, strict, and never inferred from the wall clock or a market calendar default.

The existing public `PortfolioSummary.xirr` field is activated only through the same canonical return engine. Public `nominalReturnRate` and `realReturn` remain null.

## Canonical financial semantics

External investor cash flows are derived only from the Stage 3.74 correction/reversal-aware effective ledger:

- `DEPOSIT` → negative XIRR flow;
- `WITHDRAWAL` → positive XIRR flow.

These ledger types are internal portfolio activity and are not external XIRR flows:

- `BUY`;
- `SELL`;
- `DIVIDEND`;
- `COUPON`;
- `FEE`;
- `TAX`.

Internal activity may change canonical portfolio cash and therefore terminal portfolio value, but it is not reclassified as investor contribution/withdrawal. Effective rows after the requested `asOfDate` are excluded by the Stage 3.74 cutoff semantics. Unknown future transaction types fail closed instead of being silently ignored.

## Terminal portfolio value

The terminal value is the exact requested `asOfDate` Stage 3.76 portfolio valuation. Stage 3.77 reuses the canonical Stage 3.76 transaction-level projection and requires:

- valuation coverage `COMPLETE`;
- exact requested BusinessDate for every open-position valuation;
- source `USER_SUPPLIED`;
- RUB;
- positive terminal portfolio value.

There is no live/provider/MOEX/broker price substitution, interpolation, carry-forward, carry-back, FX conversion, or hidden current-price input.

## PostgreSQL architecture

`GetPortfolioReturns` runs in one PostgreSQL `REPEATABLE READ`, `READ ONLY` transaction. The same consistent snapshot performs ownership verification, effective-ledger extraction, Stage 3.76 position/manual-valuation/cash projection, terminal-value qualification, and canonical XIRR construction.

`PortfolioSummary.xirr` invokes the same transaction-level return projection using `summary.AsOfDate`. There is no second summary XIRR engine, SQL-owned XIRR formula, alternative ledger interpretation, or alternative valuation engine.

## XIRR mathematics

The engine evaluates for `r > -1` with ACT/365:

`sum(CF_i / (1 + r)^((date_i - date_0) / 365)) = 0`

Canonical financial values remain Decimal scale 8 at the domain/public boundary. `float64` exists only inside the numerical layer.

The solver operates in `x = log(1+r)` space. It uses coefficient-aware scaled exponential evaluation, finite dominance bounds, a one-sign-change direct bisection fast path, and derivative-recursive monotonic-interval isolation for multi-sign-change series. It does not infer uniqueness from a fixed sampling grid.

Derivative recursion normalizes/factors the minimum exponent so the effective term count shrinks. Simple roots are refined by bracket width rather than a loose NPV tolerance. Tangent/critical near-zero candidates fail closed as `AMBIGUOUS_MULTIPLE_ROOTS` when a confidently unique simple root cannot be established.

Public scale-8 quantization uses Half-Even semantics. It guards float ULP/exact-integer precision limits, evaluates integer-year midpoint cases exactly with `big.Rat`, and fails closed for numerically unresolved irregular-date midpoint decisions rather than guessing the eighth decimal place.

The multi-sign-change recursive path is capped at 160 terms to prevent adversarial request monopolization; the ordinary one-sign-change contribution path is not subject to that cap.

## Solver hardening evidence

Defects found and remediated during Stage 3.77 include:

- fixed-grid close-root misses;
- non-shrinking derivative recursion;
- premature NPV-tolerance bisection termination;
- binary-float Half-Even midpoint drift;
- two-term extreme coefficient-ratio overflow/underflow;
- scale-8 float exact-integer/ULP precision boundaries;
- coefficient-magnitude loss during generalized exponential scaling;
- same-date external-flow aggregate overflow beyond canonical `NUMERIC(28,8)`;
- exact near-midpoint discrimination beyond float64 precision;
- adversarial multi-sign-change complexity behavior;
- long calendar spans without `time.Duration` saturation.

Independent numerical evidence includes:

- **20,000** random polynomial-equivalent degree 1–6 cases cross-checked against NumPy roots with **0 root-count label mismatches**;
- **1,000** irregular ACT/365 cases cross-checked against an independent 90-digit `mpmath` oracle with **0 scale-8 output mismatches**.

These are strong engineering cross-checks for the tested domains, not a formal proof for every generalized exponential series.

## Required unavailable semantics

The public projection uses exactly these fail-closed reasons:

1. `INCOMPLETE_VALUATION`
2. `NO_EXTERNAL_CONTRIBUTIONS`
3. `INSUFFICIENT_DATE_SPAN`
4. `NO_SIGN_CHANGE`
5. `NON_POSITIVE_TERMINAL_VALUE`
6. `AMBIGUOUS_MULTIPLE_ROOTS`
7. `NUMERICAL_SOLUTION_FAILED`

Unavailable projection is never converted to `0%`.

## HTTP strictness and External remediation

The Stage 3.77 endpoint validates the raw BusinessDate and rejects missing, empty, whitespace-padded, invalid-calendar, and duplicate `asOfDate` inputs before repository work.

The duplicate-query defect was found during the fresh External review of published head `88b7e08592f694d5f56c9ba275a34474f7cb677f`: the shared helper used `QueryArgs.Peek`, which could collapse repeated query keys to a single value. Stage 3.77 now counts occurrences at its own HTTP boundary and requires exactly one `asOfDate`. A dedicated regression sends two conflicting `asOfDate` parameters and verifies HTTP 400 with zero store calls.

The narrow fix is Stage 3.77-specific and does not change global query semantics for unrelated endpoints.

## PostgreSQL lifecycle witnesses

Integration tests cover:

- cash-only XIRR and summary mirroring;
- exclusion of BUY/SELL/DIVIDEND/COUPON/FEE/TAX from external-flow evidence;
- exclusion of future-dated withdrawals after the cutoff;
- subject isolation;
- exact-date manual valuation completeness;
- partial valuation coverage;
- correction-aware external flow replacement;
- reversal effective-date historical behavior;
- full-close/reopen lifecycle generation preventing stale valuation reuse;
- a real-ledger multiple-root case where an internal FEE changes terminal cash without becoming an external flow.

## HTTP / Replay compatibility

The normal route registry exposes Stage 3.77 `/returns`.

The production Replay constructor is narrowly remediated to expose:

- Stage 3.76 PUT manual valuation;
- Stage 3.76 DELETE manual valuation;
- Stage 3.77 GET `/returns`.

Replay/idempotency semantics are not changed.

## OpenAPI

The contract includes:

- endpoint-local required `asOfDate`;
- authenticated `getPortfolioReturns`;
- 200/400/401/404 responses;
- AVAILABLE/UNAVAILABLE response union;
- the exact seven-reason enum;
- canonical Decimal-string XIRR;
- external-flow evidence;
- terminal portfolio value;
- methodology `portfolio-xirr-v1` and day count `ACT/365`;
- AVAILABLE and UNAVAILABLE examples;
- validator registration and Stage 3.77 contract witnesses.

## Frontend

Portfolio detail includes a focused `PerformanceBlock`. The browser:

- requires explicit BusinessDate selection;
- calls the backend `/returns` projection;
- renders server XIRR Decimal strings without implementing XIRR/NPV/root solving;
- distinguishes AVAILABLE from UNAVAILABLE;
- renders all seven unavailable reasons;
- never substitutes unavailable with `0%`;
- displays server terminal value and ACT/365 semantics;
- refreshes selected return projection after accepted ledger/valuation mutations;
- protects principal/portfolio/date/token transitions with generation and identity guards.

TWR, nominal-return activation, real return, inflation and FX remain outside this stage.

## CI security remediation

The first published Stage 3.77 CI run exposed dependency advisories already present in the frontend dependency graph and a Stage 3.77 HTTP test fixture that constructed non-canonical zero-value Decimal fields.

The remediation:

- updated Next.js `16.3.3 → 16.3.4`;
- updated the resolved Sharp chain to `0.35.4`;
- pinned `baseline-browser-mapping` to `2.11.21` through `pnpm-workspace.yaml`;
- regenerated the pnpm lockfile;
- corrected the HTTP test fixture to use canonical `decimal.Zero()` Money values.

A dedicated GitHub Actions remediation run produced `pnpm audit: No known vulnerabilities`, frontend typecheck/tests/build PASS, and targeted Go HTTP tests PASS. Published head `88b7e08592f694d5f56c9ba275a34474f7cb677f` subsequently completed the complete protected CI matrix 10/10 GREEN.

Because External remediation changes published bytes after that run, final approval requires the same protected matrix to pass again on the new exact head.

## Out of scope / NO-GO

Stage 3.77 does not add or activate:

- TWR;
- nominal-return methodology;
- real return or inflation;
- FX;
- live/provider/MOEX prices;
- broker synchronization;
- price-history warehouse;
- workers/cron;
- Redis/Kafka;
- notifications/tax/AI/mobile;
- Stage 3.78+.

The seven canonical closure registries (`README.md`, `docs/SOURCE_OF_TRUTH.md`, `docs/ROADMAP.md`, `docs/DOCUMENT_INDEX.md`, `docs/IMPLEMENTATION_LOG.md`, `docs/VERSION_MATRIX.md`, `docs/OPEN_QUESTIONS.md`) remain outside the implementation PR and belong only to the separate post-merge closure workflow.

## Published Draft PR changed-file set

Draft PR #157 remains exactly at the **25-file governance ceiling**. External remediation modifies existing files only.

1. `backend-go/cmd/validate-openapi/stage_03_77_registration.go`
2. `backend-go/cmd/validate-openapi/stage_03_77_returns_contract_test.go`
3. `backend-go/internal/httpapi/portfolios.go`
4. `backend-go/internal/httpapi/replay_app.go`
5. `backend-go/internal/httpapi/returns.go`
6. `backend-go/internal/httpapi/routes.go`
7. `backend-go/internal/httpapi/stage_03_77_returns_test.go`
8. `backend-go/internal/postgres/stage_03_71_summary.go`
9. `backend-go/internal/postgres/stage_03_72_positions.go`
10. `backend-go/internal/postgres/stage_03_77_returns.go`
11. `backend-go/internal/postgres/stage_03_77_returns_integration_test.go`
12. `backend-go/internal/verticalslice/stage_03_77_returns.go`
13. `backend-go/internal/verticalslice/stage_03_77_returns_test.go`
14. `backend-go/internal/verticalslice/stage_03_77_store_adapter.go`
15. `docs/stages/STAGE_03_77_PORTFOLIO_XIRR_IMPLEMENTATION.md`
16. `frontend-next/package.json`
17. `frontend-next/pnpm-lock.yaml`
18. `frontend-next/pnpm-workspace.yaml`
19. `frontend-next/src/common/api/openinvest.ts`
20. `frontend-next/src/features/portfolio/components/PerformanceBlock.tsx`
21. `frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx`
22. `frontend-next/tests/stage-03-77-portfolio-xirr.test.mjs`
23. `openapi/components/stage_03_77.yaml`
24. `openapi/examples/portfolio-returns.json`
25. `openapi/openapi.yaml`

## Verification and limitations

Local evidence includes solver regressions, oracle cross-checks, frontend static/contract witnesses, YAML/JSON parsing, scope checks, and reconstruction checks. Local runtime constraints are not used to claim the protected repository gates.

Authoritative whole-repository verification is GitHub Actions on the exact published PR head. The previous published head completed 10/10 GREEN; the External remediation head must independently repeat that result before approval.

## Governance stop condition

Pre-publication Internal Review completed with `APPROVED`, publication authorization was received, and Draft PR #157 was created.

The current gate is the External remediation cycle: publish the repaired exact head, require all protected CI jobs to pass on that head, and complete fresh External verification with no unresolved findings. Internal Review evidence remains withheld until External approval and the evidence-only follow-up phase.

Ready-for-review requires separate explicit human authorization. Squash merge requires another separate explicit authorization. After implementation merge, a separate docs-only closure PR is required before Stage 3.77 may be called **FULLY CLOSED / CANONICAL**.
