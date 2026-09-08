# Stage 3.77 — Portfolio Money-Weighted Return / XIRR Implementation

## Status

Stage 3.77 implementation candidate is assembled locally against the immutable Stage 3.76 canonical base. Publication, Draft PR creation, Ready-for-review transition, merge, and Stage 3.78+ remain unauthorized. Full canonical closure is not claimed by this document.

Internal Review Evidence: **WITHHELD — external published-head phase pending**.

## Canonical base

- repository: `AsifAbbasov/OpenInvest`
- protected base branch: `develop`
- base commit: `12b41d2bb03999e3adefacf1ebb408fb30d732b8`
- base tree: `8e8c183912cf87fe0e68f3d73c1032aec158a409`
- base state: Stage 3.76 fully closed/canonical before Stage 3.77 work

The implementation candidate is reconstructed only from immutable base blobs plus Stage 3.77 deltas. Modified pre-existing files are byte-checked against their canonical base blobs when the Stage 3.77 delta is removed.

## Scope

Stage 3.77 implements one backend-owned money-weighted portfolio return projection using XIRR semantics.

It adds the authenticated endpoint:

`GET /api/v1/portfolios/{portfolioId}/returns?asOfDate=YYYY-MM-DD`

`asOfDate` is mandatory and explicit. No wall-clock date is substituted.

The existing public `PortfolioSummary.xirr` field is activated only by the same canonical projection engine. `nominalReturnRate` and `realReturn` remain inactive/null in the public HTTP contract.

## Canonical financial semantics

External investor cash flows come only from the Stage 3.74 correction/reversal-aware effective ledger:

- `DEPOSIT` => negative XIRR cash flow;
- `WITHDRAWAL` => positive XIRR cash flow.

The following are intentionally not external XIRR flows:

- `BUY`;
- `SELL`;
- `DIVIDEND`;
- `COUPON`;
- `FEE`;
- `TAX`.

Those internal ledger events may change portfolio cash and therefore terminal portfolio value, but they are not reclassified as investor contributions or withdrawals.

Future-dated effective rows after the requested `asOfDate` are excluded by the canonical Stage 3.74 effective-ledger cutoff.

## Terminal portfolio value

The terminal XIRR value is the exact requested `asOfDate` Stage 3.76 manual-valuation portfolio value.

The implementation reuses the Stage 3.76 position/manual-valuation path through the shared transaction-level `portfolioPositionsProjectionTx` helper and requires:

- valuation coverage `COMPLETE`;
- every open position valued for the exact requested BusinessDate;
- source `USER_SUPPLIED` only;
- RUB only;
- positive `currentPortfolioValue`.

There is no live, MOEX, broker, provider, interpolation, carry-forward, carry-back, or hidden current-price substitution.

## PostgreSQL architecture

`GetPortfolioReturns` runs in one PostgreSQL transaction configured as:

- `REPEATABLE READ`;
- `READ ONLY`.

Within the same database snapshot it performs:

1. subject/portfolio ownership verification;
2. Stage 3.74 effective-ledger read at the exact `asOfDate`;
3. extraction of DEPOSIT/WITHDRAWAL external flows;
4. the canonical Stage 3.76 transaction-level position/manual-valuation/cash projection via `portfolioPositionsProjectionTx`;
5. Stage 3.76 valuation coverage/terminal-value check;
6. canonical domain XIRR solve.

`GetPortfolioPositions` and Stage 3.77 both call the same transaction-level Stage 3.76 helper; Stage 3.77 does not duplicate its rebuild/manual-valuation/cash orchestration. The repository does not implement SQL-owned XIRR arithmetic, a second ledger interpretation, or a second valuation engine.

`PortfolioSummary.xirr` calls the same transaction-level `portfolioReturnProjectionTx` using `summary.AsOfDate`; there is no second summary XIRR calculation.

## XIRR mathematics

The equation is evaluated for `r > -1` with ACT/365 year fractions:

`sum(CF_i / (1 + r)^((date_i - date_0) / 365)) = 0`

BusinessDate differences are based on normalized UTC-midnight calendar-day arithmetic / Unix-day differences, avoiding duration overflow for large spans.

Financial Decimal values remain canonical scale-8 values at the domain/public boundary. `float64` is used only inside the numerical solver.

The solver works in `x = log(1+r)` space and uses derivative-recursive monotonic-interval root isolation. It does not use a fixed sampling grid to infer uniqueness.

At each recursion level the minimum exponent is factored out before differentiation. The removed positive exponential factor cannot change roots and guarantees that the derivative recursion reduces the effective term count.

Finite dominance bounds are constructed in log-space. The two-term fast path also solves coefficient ratios in log-space to avoid overflow/underflow from direct tiny/huge division.

Simple roots are refined by bracket width, not by prematurely accepting a small absolute NPV value. Critical/tangent candidates are fail-closed: a near-zero derivative at the only candidate is classified as `AMBIGUOUS_MULTIPLE_ROOTS`, because multiplicity means the engine cannot publish it as a confidently unique simple XIRR.

The final public result is quantized to scale 8 with Half-Even semantics. Midpoint handling does not rely on binary floating-point formatting. General multi-term evaluation scales each signed term by `log(abs(coefficient)) - years*x`, so coefficient magnitude participates in the numerical scaling instead of only the exponential factor.

## Solver defects found and remediated during Stage 3.77

### Fixed-grid close-root miss

The initial 4096-interval `log(1+r)` scan could miss two nearby roots inside one sample interval. It was replaced by derivative-recursive root isolation rather than a denser heuristic grid.

### Non-shrinking derivative recursion

An intermediate reconstruction differentiated terms without first factoring the minimum exponent, so the recursive representation did not necessarily shrink and could stack-overflow. Each recursive level now normalizes its minimum exponent to zero before differentiation.

### Premature NPV-tolerance termination

Simple-root bisection originally stopped when `|NPV| <= 1e-12`; a scale-8 midpoint case could therefore stop on the wrong side of the Half-Even boundary. Simple-root refinement now terminates by bracket width.

### Binary-float Half-Even boundary

A binary representation such as `0.123456795` can format as `0.12345679` even though the canonical decimal midpoint rounds Half-Even to `0.12345680`. Public quantization now resolves scale-8 midpoints explicitly and then enters the canonical Decimal boundary.

### Two-term tiny/huge coefficient ratio

Internal review found that the general isolator used log-space bounds while the two-term fast path still evaluated `-c0/c1` directly. Extreme coefficient ratios could overflow or underflow and fail numerically. The two-term solution now computes the coefficient ratio in log-space. A dedicated extreme-ratio regression covers the repaired path.

### Scale-8 float exact-integer boundary

Internal review also found that an extreme rate can make `rate * 1e8` exceed the IEEE-754 exact-integer range. In that regime a float64 cannot honestly guarantee the requested eighth decimal place even if the value would fit the storage Decimal range. Public quantization now checks two precision boundaries before publishing: the local float64 ULP must remain below one 128th of the `1e-8` output quantum, and `|rate * 1e8|` must remain below the exact-integer `2^53` boundary. Values outside either boundary return `NUMERICAL_SOLUTION_FAILED` rather than false precision. After Half-Even chooses an exact scale unit, the canonical Decimal text is assembled from the exact integer scale units; it is not produced by dividing back through float64 and formatting the quotient.

### Coefficient-magnitude loss in generalized exponential evaluation

Internal review found that the multi-term evaluator originally normalized only the exponential factor. With sufficiently different coefficient magnitudes, a mathematically material small coefficient could underflow before the compensated sum saw it. Evaluation now normalizes the full log-magnitude `log(abs(coefficient)) - years*x`, preserves the sign separately, and then performs compensated summation on the scaled terms. A dedicated regression verifies that coefficient magnitude participates in the scale selection.

### Same-date external-flow aggregate storage boundary

External flows originate as canonical Decimal amounts, but multiple individually valid DEPOSIT/WITHDRAWAL rows on one BusinessDate can sum beyond the canonical `NUMERIC(28,8)` boundary. Stage 3.77 now validates both each included external flow and every same-date aggregate with the canonical Decimal storage predicate. Overflow fails closed as invalid input rather than allowing non-canonical evidence into the XIRR solver or public response.

### Exact scale-8 midpoint beyond float64 discrimination

Internal review reproduced a stronger Half-Even defect using valid `NUMERIC(28,8)` cash flows: the exact XIRR was only `1e-26` above a scale-8 midpoint, while float64 collapsed it onto the midpoint and would have rounded to the wrong scale unit. For integer-year ACT/365 series, midpoint classification now evaluates the transformed polynomial exactly with `big.Rat`; an exact zero applies Half-Even, while a non-zero exact sign determines the root side. For irregular ACT/365 exponents, a numerically unresolved midpoint fails closed as `NUMERICAL_SOLUTION_FAILED` rather than guessing a last decimal. Dedicated just-above and just-below regressions cover the exact polynomial path.

### One-sign-change performance path and adversarial complexity guard

Internal performance review showed that derivative recursion is unnecessary for ordinary contribution histories and can be adversarial for hundreds of alternating-sign dates. The solver now counts coefficient sign changes after normalization. Zero sign changes means no admissible root; exactly one sign change uses dominance bounds plus direct bisection and does not recurse. A regression with 1,500 distinct negative contribution dates plus a positive terminal value exercises this uncapped ordinary-history path.

Series with more than one sign change still use derivative-recursive isolation because multiple admissible roots must be detected. A measured alternating-sign stress case showed a sharp runtime cliff beyond roughly 180–200 terms, so the ambiguous recursive path has an explicit safety ceiling of 160 terms. Larger multi-sign-change series fail closed as `NUMERICAL_SOLUTION_FAILED` instead of monopolizing an API process. This ceiling does not apply to the one-sign-change fast path.

### Strict BusinessDate query boundary

Internal API review found that trimming `asOfDate` would accept whitespace-padded values despite the documented strict `YYYY-MM-DD` contract. Stage 3.77 now validates the raw query value without trimming and rejects whitespace-padded dates before repository work.

## Required unavailable semantics

The endpoint and domain projection expose explicit fail-closed states:

- `INCOMPLETE_VALUATION`;
- `NO_EXTERNAL_CONTRIBUTIONS`;
- `INSUFFICIENT_DATE_SPAN`;
- `NO_SIGN_CHANGE`;
- `NON_POSITIVE_TERMINAL_VALUE`;
- `AMBIGUOUS_MULTIPLE_ROOTS`;
- `NUMERICAL_SOLUTION_FAILED`.

`NO_SIGN_CHANGE` means that no mathematically admissible XIRR root exists for the constructed series. `NUMERICAL_SOLUTION_FAILED` is reserved for numerical/isolation failure rather than ordinary absence of a root.

Unavailable projection is never converted to `0%`.

Fail-closed reason precedence is deterministic: incomplete exact-date valuation first; then non-positive terminal value; then absence of an effective negative external contribution; then date-span/root-isolation outcomes. This prevents a lower-level solver label from hiding a more fundamental missing valuation or invalid terminal-value prerequisite.

## Verification vectors

The domain regression suite includes:

- simple `+10%`;
- simple `-20%`;
- an irregular ACT/365 vector yielding `0.12048717`;
- same-date aggregation;
- future-flow cutoff;
- no-root / `NO_SIGN_CHANGE`;
- `r > -1` boundary fail-closed behavior;
- conventional multiple roots `[-100, +230, -132]` with roots 10% and 20%;
- close multiple roots based on `[+123, +403, -429, -175, +165]` with roots approximately `-0.29519336` and `-0.28653470`;
- tangent/double-root fail-closed behavior;
- scale-8 Half-Even midpoint cases;
- extreme two-term log-space ratio handling;
- scale-8 float exact-integer-boundary fail-closed behavior;
- coefficient-magnitude-aware generalized exponential scaling;
- same-date external-flow aggregate canonical-storage rejection;
- exact just-above / just-below scale-8 midpoint discrimination beyond float64 precision;
- strict whitespace-padded BusinessDate rejection;
- a 1,500-date ordinary contribution history on the one-sign-change fast path;
- fail-closed protection for adversarial multi-sign-change series above the recursive complexity ceiling.

An independent NumPy polynomial-root oracle was regenerated/checked against the current derivative-recursive solver path for exactly **20,000 random polynomial-equivalent cases of degree 1–6**. Recalculation with `numpy.roots` produced **0 root-count label mismatches**, and the isolated current Go candidate solver passed those same 20,000 cases with no unexpected solver failure. This is strong engineering evidence for the tested polynomial-equivalent domain, not a formal mathematical proof for every arbitrary irregular-date generalized exponential series.

A separate high-precision irregular-date oracle generated **1,000 random ACT/365 cases** with non-integer year fractions, solved them independently with `mpmath` at 90 decimal digits, and quantized expected outputs with Decimal Half-Even. The current Go candidate matched all 1,000 expected scale-8 XIRR outputs. This directly complements the polynomial-equivalent oracle in the irregular-date domain.

## PostgreSQL lifecycle witnesses

Integration tests exercise real repository/service paths for:

- cash-only XIRR and summary mirroring;
- exclusion of BUY, SELL, DIVIDEND, COUPON, FEE and TAX from external-flow evidence while their canonical ledger effects remain part of terminal portfolio value where applicable;
- exclusion of a future-dated WITHDRAWAL after the requested `asOfDate`;
- subject isolation;
- exact-date manual valuation completeness;
- partial valuation coverage returning `INCOMPLETE_VALUATION`;
- correction-aware external flow replacement;
- reversal effective-date historical behavior;
- full-close/reopen lifecycle generation preventing a stale manual valuation from becoming terminal valuation for the reopened position;
- a real-ledger multiple-root case producing `[-100, +230, -132]`, where an internal FEE changes terminal value but is not itself an external investor flow.

## HTTP and Replay route remediation

The normal route registry exposes `/returns`.

The production Replay constructor is narrowly remediated to expose:

- Stage 3.76 PUT manual valuation;
- Stage 3.76 DELETE manual valuation;
- Stage 3.77 GET `/returns`.

The change does not alter replay/idempotency semantics and does not expand into unrelated historical route drift. HTTP regression witnesses prove route availability in normal and Replay constructors.

## OpenAPI

The OpenAPI contract adds:

- required endpoint-local `asOfDate`;
- authenticated GET operation `getPortfolioReturns`;
- 200/400/401/404 responses;
- AVAILABLE and UNAVAILABLE discriminated response schemas;
- exact unavailable-reason enum;
- Decimal-string XIRR, not a JSON numeric financial field;
- external-flow evidence, terminal portfolio value, and ACT/365 methodology metadata;
- AVAILABLE and UNAVAILABLE examples;
- validator registration and dedicated Stage 3.77 contract witnesses.

## Frontend

Portfolio detail gains a focused `PerformanceBlock`.

The Web client:

- starts with no implicit date and requires explicit BusinessDate selection;
- calls only the backend `/returns?asOfDate=...` projection;
- renders backend XIRR Decimal strings and converts the decimal string to percent presentation without binary numeric root/return calculation;
- distinguishes AVAILABLE from UNAVAILABLE;
- maps each backend unavailable reason to a human-readable explanation;
- never substitutes unavailable with `0%`;
- states ACT/365 and explicit manual terminal valuation semantics;
- contains no live/provider/MOEX-price wording for XIRR;
- refreshes the selected return projection after ledger and valuation mutations;
- protects against stale principal/portfolio/date responses using a generation guard plus principal, portfolio and `asOfDate` identity;
- rechecks that performance identity after the awaited mutation refresh before issuing the follow-up XIRR request, so an old mutation callback cannot invalidate a newer BusinessDate load.

The prior portfolio-detail copy is corrected so it no longer says all return methodology is unavailable: XIRR is implemented by Stage 3.77, while TWR, nominal return activation, real return and inflation remain outside the implemented surface.

## Out of scope / NO-GO

This implementation does not add or activate:

- TWR;
- nominal-return methodology;
- real return or inflation;
- FX;
- live/provider/MOEX prices;
- broker synchronization;
- historical price warehouse;
- workers/cron;
- Redis/Kafka;
- notifications/tax/AI/mobile work;
- Stage 3.78+.

No implementation-PR edit is made to the seven canonical closure registries (`README.md`, `docs/SOURCE_OF_TRUTH.md`, `docs/ROADMAP.md`, `docs/DOCUMENT_INDEX.md`, `docs/IMPLEMENTATION_LOG.md`, `docs/VERSION_MATRIX.md`, `docs/OPEN_QUESTIONS.md`). Those belong only to the separate post-merge closure workflow.

## Candidate changed-file set

The implementation candidate is intentionally below the 25-file ceiling:

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
15. `frontend-next/src/common/api/openinvest.ts`
16. `frontend-next/src/features/portfolio/components/PerformanceBlock.tsx`
17. `frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx`
18. `frontend-next/tests/stage-03-77-portfolio-xirr.test.mjs`
19. `openapi/openapi.yaml`
20. `openapi/components/stage_03_77.yaml`
21. `openapi/examples/portfolio-returns.json`
22. `docs/stages/STAGE_03_77_PORTFOLIO_XIRR_IMPLEMENTATION.md`

## Local verification and limitations

Available local evidence includes:

- byte-exact reconstruction checks for all modified pre-existing files;
- `gofmt` cleanliness;
- isolated Stage 3.77 Go solver regressions;
- the 20,000-case independent NumPy/Go polynomial root-count cross-check;
- the 1,000-case independent high-precision irregular ACT/365 `mpmath`/Go scale-8 cross-check;
- adversarial multi-sign-change runtime characterization and explicit fail-closed complexity protection;
- frontend Stage 3.77 contract tests;
- standalone TypeScript surface checking for the API client;
- OpenAPI YAML / JSON parsing and Stage 3.77 static contract witnesses;
- scope, no-wall-clock and no-browser-XIRR static checks.

The local runtime provides Go 1.23.2 while the repository requires Go 1.24. `OPENINVEST_DATABASE_TEST_URL` is not available in this runtime, and `pnpm` is not installed. Therefore this dossier does **not** claim that the complete repository Go 1.24 suite, PostgreSQL integration suite, full Next.js build/typecheck, race/vet/vulnerability/security jobs, or the protected ten-job CI matrix passed locally.

Those authoritative whole-repository checks must run against the exact published candidate head in GitHub Actions after explicit publication authorization. Until that phase succeeds, Stage 3.77 is not merged or canonically closed.

## Governance stop condition

Before publication, the candidate must receive a complete read-only Internal Review over every changed file and every changed line. Only an Internal `APPROVED` verdict permits asking for explicit commit/push + Draft PR authorization.

After publication, exact-head protected CI and a fresh independent External published-head review are still mandatory. Ready-for-review and squash merge each require their own explicit human authorization. A separate docs-only closure PR is required after implementation merge before Stage 3.77 may be called **FULLY CLOSED / CANONICAL**.
