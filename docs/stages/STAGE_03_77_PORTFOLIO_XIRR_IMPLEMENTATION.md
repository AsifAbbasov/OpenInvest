# Stage 3.77 — Portfolio Money-Weighted Return / XIRR Implementation

> **Historical Stage 3.77 implementation/review snapshot**
>
> This dossier preserves the Stage 3.77 implementation/review state that was current when it was written.
> Draft-PR, authorization, CI/evidence-head and next-stage statements below are historical evidence; they do not define the current repository state.
> Current product/runtime status: [`../SOURCE_OF_TRUTH.md`](../SOURCE_OF_TRUTH.md).
> Current planning/status: [`../ROADMAP.md`](../ROADMAP.md).
> Implementation chronology: [`../IMPLEMENTATION_LOG.md`](../IMPLEMENTATION_LOG.md).

## Status

Stage 3.77 is published as Draft PR #157 against the immutable Stage 3.76 canonical base. Ready-for-review, merge, and Stage 3.78+ remain unauthorized. Full canonical closure is not claimed by this document.

Internal Review Evidence: **PUBLISHED — APPROVED**.

External Published-Head Review: **APPROVED** for implementation/remediation head `3a976568567c478a20ed51e85fb35da020132328`, tree `aa7be937517cdd96a2cb1c3689b80e6cf531b258`, after protected exact-head CI completed 10/10 GREEN.

This document update is the required **evidence-only follow-up**. It changes review/governance evidence only. It must receive its own protected exact-head CI pass and no-drift verification before a separate Ready-for-review authorization may be requested.

## Canonical base

- repository: `AsifAbbasov/OpenInvest`
- protected base branch: `develop`
- base commit: `12b41d2bb03999e3adefacf1ebb408fb30d732b8`
- base tree: `8e8c183912cf87fe0e68f3d73c1032aec158a409`
- base state: Stage 3.76 fully closed/canonical before Stage 3.77 work

## Published review evidence

### Pre-publication Internal review

The original 22-file Stage 3.77 implementation candidate received a complete read-only Internal review with verdict **APPROVED** before publication authorization.

- approved candidate tree: `2ba4bc559fd8d9417128f69a38b69bf94d7c3c5b`
- first implementation commit: `1d236c878499707ce76cd4f9ee27b6f815a386cb`

The Internal cycle found and remediated numerical, precision, performance, PostgreSQL integration, Replay-route, and frontend stale-response issues before the first publication.

### CI-remediation review

The first published CI run exposed two non-financial blockers: a Stage 3.77 HTTP test fixture with non-canonical zero Decimal fields and dependency advisories in the existing frontend dependency graph. A narrow remediation review approved the four product/dependency files changed to resolve those blockers.

- remediation commit: `88b7e08592f694d5f56c9ba275a34474f7cb677f`
- remediation tree: `a2bdc857f2a3ffa6e407614845e811f816a42577`
- protected CI on that published head: **10/10 GREEN**

### External published-head review

Fresh External review was performed on the exact published `88b7e085...` bytes across all 25 changed files, with no sampling. It returned **REQUEST CHANGES** for two concrete findings:

1. repeated `asOfDate` query parameters were collapsed by the shared `QueryArgs.Peek` behavior instead of being rejected as ambiguous Stage 3.77 input;
2. this implementation dossier still described the pre-publication 22-file state instead of the actual 25-file Draft PR.

Both findings were remediated without adding a 26th changed path.

External remediation changed only:

- `backend-go/internal/httpapi/returns.go`;
- `backend-go/internal/httpapi/stage_03_77_returns_test.go`;
- `docs/stages/STAGE_03_77_PORTFOLIO_XIRR_IMPLEMENTATION.md`.

The final remediation implementation head before this evidence-only follow-up was:

- commit: `3a976568567c478a20ed51e85fb35da020132328`
- tree: `aa7be937517cdd96a2cb1c3689b80e6cf531b258`
- protected CI run: `34303636331`
- protected CI result: **10/10 GREEN**

A compare from `88b7e085...` to `3a976568...` showed only those three remediation paths changed. Therefore the other 22 already-reviewed published files were byte-identical; the three changed files were reread in full. With the duplicate-query finding fixed, the regression passing in the full Go suite and race suite, and no remaining review findings, the External verdict became **APPROVED**.

## Scope

Stage 3.77 implements one backend-owned money-weighted portfolio return projection using XIRR semantics and adds:

`GET /api/v1/portfolios/{portfolioId}/returns?asOfDate=YYYY-MM-DD`

The endpoint is authenticated. `asOfDate` is mandatory, explicit, strict, and never inferred from the wall clock or a market-calendar default.

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

Internal activity may change canonical portfolio cash and therefore terminal portfolio value, but it is not reclassified as investor contribution/withdrawal. Effective rows after the requested `asOfDate` are excluded by the Stage 3.74 cutoff semantics. Unknown future transaction types fail closed.

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

The solver operates in `x = log(1+r)` space. It uses coefficient-aware scaled exponential evaluation, finite dominance bounds, a one-sign-change direct-bisection fast path, and derivative-recursive monotonic-interval isolation for multi-sign-change series. It does not infer uniqueness from a fixed sampling grid.

Derivative recursion factors the minimum exponent so the effective term count shrinks. Simple roots are refined by bracket width rather than a loose NPV tolerance. Tangent/critical candidates fail closed as `AMBIGUOUS_MULTIPLE_ROOTS` when a confidently unique simple root cannot be established.

Public scale-8 quantization uses Half-Even semantics. It guards float ULP/exact-integer precision limits, evaluates integer-year midpoint cases exactly with `big.Rat`, and fails closed for numerically unresolved irregular-date midpoint decisions instead of guessing the eighth decimal place.

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
- same-date aggregate overflow beyond canonical `NUMERIC(28,8)`;
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

## HTTP strictness

The Stage 3.77 HTTP boundary rejects missing, empty, whitespace-padded, invalid-calendar, and duplicate `asOfDate` inputs before repository work.

The duplicate-query remediation is Stage 3.77-specific: it counts occurrences at the endpoint boundary and requires exactly one `asOfDate`. A dedicated regression sends conflicting repeated values and proves HTTP 400 with zero store calls. Global query semantics for unrelated endpoints are not changed.

## PostgreSQL lifecycle witnesses

Integration tests cover:

- cash-only XIRR and summary mirroring;
- exclusion of BUY/SELL/DIVIDEND/COUPON/FEE/TAX from external-flow evidence;
- future-withdrawal cutoff;
- subject isolation;
- exact-date and partial manual-valuation coverage;
- correction-aware external-flow replacement;
- reversal effective-date historical behavior;
- full-close/reopen lifecycle generation preventing stale valuation reuse;
- a real-ledger multiple-root case where an internal FEE changes terminal cash without becoming an external flow.

## HTTP / Replay compatibility

The normal route registry exposes Stage 3.77 `/returns`. The production Replay constructor is narrowly remediated to expose Stage 3.76 valuation PUT/DELETE and Stage 3.77 GET `/returns`. Replay/idempotency semantics are not changed.

## OpenAPI

The contract includes endpoint-local required `asOfDate`, authenticated `getPortfolioReturns`, 200/400/401/404 responses, AVAILABLE/UNAVAILABLE union, the exact seven-reason enum, Decimal-string XIRR, external-flow evidence, terminal portfolio value, methodology `portfolio-xirr-v1`, ACT/365, examples, validator registration, and dedicated contract witnesses.

## Frontend

Portfolio detail includes a focused `PerformanceBlock`. The browser requires explicit BusinessDate selection, calls only the backend projection, renders server Decimal strings without XIRR/NPV/root solving, distinguishes AVAILABLE/UNAVAILABLE, renders all seven reasons, never substitutes unavailable with `0%`, refreshes after accepted ledger/valuation mutations, and guards principal/portfolio/date/token transitions against stale responses.

TWR, nominal-return activation, real return, inflation and FX remain outside this stage.

## CI security remediation

The first CI cycle required:

- Next.js `16.3.3 → 16.3.4`;
- resolved Sharp `0.35.4`;
- `baseline-browser-mapping` override `2.11.21` in `pnpm-workspace.yaml`;
- regenerated pnpm lockfile;
- canonical zero-Money initialization in the Stage 3.77 HTTP test fixture.

The remediation produced `pnpm audit: No known vulnerabilities`, frontend typecheck/tests/build PASS, and full protected CI 10/10 GREEN on `88b7e085...`. The subsequent External remediation also passed the complete protected matrix 10/10 GREEN on `3a976568...`.

## Out of scope / NO-GO

Stage 3.77 does not add or activate TWR, nominal-return methodology, real return/inflation, FX, live/provider/MOEX prices, broker sync, price history, workers/cron, Redis/Kafka, notifications, tax, AI/mobile, or Stage 3.78+.

The seven canonical closure registries (`README.md`, `docs/SOURCE_OF_TRUTH.md`, `docs/ROADMAP.md`, `docs/DOCUMENT_INDEX.md`, `docs/IMPLEMENTATION_LOG.md`, `docs/VERSION_MATRIX.md`, `docs/OPEN_QUESTIONS.md`) remain outside the implementation PR and belong only to the separate post-merge closure workflow.

## Published Draft PR changed-file set

Draft PR #157 remains exactly at the **25-file governance ceiling**. The evidence-only follow-up changes an already-counted dossier path only.

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

## Governance stop condition

Implementation/remediation External review is **APPROVED**. This evidence-only follow-up must now pass exact-head protected CI and prove no code/contract/dependency drift relative to `3a976568...`.

Only after that verification may a separate explicit human authorization be requested for the Ready-for-review transition. Ready is not merge authorization. Squash merge requires another explicit authorization. After implementation merge, a separate docs-only closure PR is mandatory before Stage 3.77 may be called **FULLY CLOSED / CANONICAL**.
