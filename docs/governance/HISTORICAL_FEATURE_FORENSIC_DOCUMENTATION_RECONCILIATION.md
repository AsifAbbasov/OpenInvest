# Historical Feature Forensic Documentation Reconciliation — Stages 3.57–3.77

| Field | Value |
| --- | --- |
| Document ID | GOV-HIST-FEATURE-FORENSIC-001 |
| Version | 1.0.0 |
| Status | APPROVED / CANONICAL |
| Owner | Principal Architect |
| Repository | `AsifAbbasov/OpenInvest` |
| Canonical base | `develop@5d8ed09ff1061a57281b231cd4f448df174a625a` |
| Canonical base tree | `78579ebb3d39501a28bb87a8d843791a5983b325` |
| Reviewed PR head | `fa2f0a164c7e2c4fe414a7bb9469eedf5ea7df54` |
| Reviewed tree | `e742d1084fdcc5751acf896738ba9065a4b0f741` |
| Protected develop merge | `5da679f0d15f8742239659ff9cebc451a213894d` |
| Merged tree | `e742d1084fdcc5751acf896738ba9065a4b0f741` |
| Content drift | NONE |
| Exact-head CI | 10/10 SUCCESS |
| Historical scope | Stages 3.57–3.77 inclusive |
| Reference standard | Feature 3D post-merge forensic closure; referenced, not rewritten |
| Runtime / API / SQL / dependency / provider activation change | NONE |
| Stage 3.78 authorization | NONE |
| Date | 2026-09-10 |

## 1. Purpose

This document reconciles the historical engineering-review record for the post-audit feature sequence from Stage 3.57 through Stage 3.77 into one forensic navigation surface.

It does **not** rewrite the historical stage dossiers. The original planning, implementation, review-evidence, governance-disposition and closure documents remain the time-specific primary evidence. This reconciliation adds a normalized cross-stage index so later reviewers can answer:

- what problem or defect was actually found;
- what design assumption caused it;
- what failure or attack scenario was demonstrated or explicitly considered;
- what project impact mattered;
- what insufficient design was replaced;
- what remediation was selected;
- what regression/CI/review evidence closed the issue;
- what evidence is missing and therefore must remain missing.

Feature 3D raised the documentation standard by preserving problem → root cause → failure scenario → remediation → evidence → residual limitation in one closure dossier. Earlier stages contain much of the same substance, but it is distributed unevenly across planning dossiers, implementation records, PR comments, CI chronology and later closure documents. This reconciliation normalizes those historical facts **without inventing missing review history**.

## 2. Non-retroactive evidence rule

The governing rule for this document is:

> contemporary evidence outranks later narrative completeness.

Accordingly:

1. a finding is recorded only when the contemporary repository/PR/CI evidence demonstrates it;
2. later code state is not used to invent an earlier reviewer rationale;
3. a later fix does not prove what the earlier reviewer thought unless the earlier evidence says so;
4. missing Internal/External evidence is never reconstructed from memory or final code;
5. an irreversible governance deviation remains historically noncompliant even after disposition;
6. a green CI run proves the checks that ran on that exact head, not every semantic invariant;
7. implementation presence does not imply provider/runtime activation or source/use permission.

Where a required forensic field cannot be supported, the literal marker is:

```text
NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
```

This reconciliation contains exactly **60** such explicit field markers across the 47 material forensic records below.

## 3. Classification vocabulary

- **FULL FORENSIC COVERAGE** — the material problem, remediation and verification chain is recoverable from contemporary evidence. An optional second-order field may still be `NOT RECORDED`.
- **STRONG PARTIAL** — the principal material finding is recoverable, but some surrounding historical reviewer/causal fields are not preserved.
- **PARTIAL** — final defect/remediation state is known, but multiple original-cause/reviewer fields are missing.
- **MINIMAL / NO MATERIAL REVIEW FINDINGS RECORDED** — no material remediation finding is preserved; this document does not manufacture one.
- **HISTORICAL EVIDENCE GAP** — a mandatory historical evidence artifact is known to be unavailable and must remain explicitly unavailable.

A **preventive design decision** or **prepublication hardening** is not relabeled as an External/Internal finding unless the contemporary source says it was one.

## 4. Evidence corpus and precedence

The reconciliation audit read the canonical Stage 3.57–3.77 dossiers, relevant closures/governance dispositions, ADR-009, current registries, `docs/REVIEW_WORKFLOW.md`, and Feature 3D reference evidence. It also cross-checked material chronology against the relevant GitHub PR evidence for #123, #126, #128, #138, #142, #150, #152, #153, #155 and #157.

Evidence precedence for this document is:

1. exact contemporary stage/review/governance dossier on protected history;
2. exact PR comment/review and exact-head CI/commit identity;
3. accepted ADR or canonical lifecycle registry for scope/status;
4. later closure/reconciliation text only as navigation or current-state confirmation.

No chat-only reconstruction is treated as canonical historical evidence.

## 5. Audit summary

```text
Historical stages inspected:           21
Feature 3D reference inspected:         YES
Primary repository evidence files:      41
GitHub PR evidence sets cross-checked:   10
Material forensic records:              47
Full forensic records:                  33
Partial forensic records:               14
Explicit NOT RECORDED fields:           60
Historical dossiers rewritten:          NO
Runtime/API/DB/dependency changes:       NONE
Provider/source activation changes:     NONE
Stage 3.78 started:                      NO
```

The 33/14 full-versus-partial count is record-level. Stage-level labels can still be `STRONG PARTIAL` where the main material finding is reconstructable but secondary chronology is incomplete.

## 6. Stage coverage matrix

| Stage | Subject | Coverage | Material records | Primary evidence |
| --- | --- | --- | ---: | --- |
| 3.57 | Market Data Provider Boundary | MINIMAL / NO MATERIAL REVIEW FINDINGS RECORDED | 0 | `docs/stages/STAGE_03_57_MARKET_DATA_PROVIDER_BOUNDARY_PLANNING.md`, `...IMPLEMENTATION.md` |
| 3.58 | Market Data Provider Boundary Closure | MINIMAL / NO MATERIAL REVIEW FINDINGS RECORDED | 0 | `docs/stages/STAGE_03_58_MARKET_DATA_PROVIDER_BOUNDARY_CLOSURE.md` |
| 3.59 | MOEX ISS Quote Provider Adapter | FULL FORENSIC COVERAGE | 4 | `docs/stages/STAGE_03_59_MOEX_ISS_QUOTE_PROVIDER_PLANNING.md`, `...IMPLEMENTATION.md`, PR #123 |
| 3.60 | MOEX ISS Runtime / Source Activation Decision | MINIMAL / DESIGN-SOURCE GOVERNANCE | 0 | `docs/stages/STAGE_03_60_MOEX_ISS_RUNTIME_SOURCE_ACTIVATION_PLANNING.md`, Data Source Registry |
| 3.61 | Corporate Actions Calendar Planning | MINIMAL / DESIGN-SOURCE GOVERNANCE | 0 | `docs/stages/STAGE_03_61_CORPORATE_ACTIONS_CALENDAR_PLANNING.md` |
| 3.62 | Corporate Action Boundary / Feature 3A | STRONG PARTIAL | 1 | `docs/stages/STAGE_03_62_CORPORATE_ACTION_BOUNDARY_IMPLEMENTATION.md`, PR #126 |
| 3.63 | Calendar + Heatmap Projection / Feature 3B | STRONG PARTIAL | 1 | `docs/stages/STAGE_03_63_CORPORATE_ACTION_CALENDAR_HEATMAP_PROJECTION_IMPLEMENTATION.md`, PR #127 |
| 3.64 | Corporate Actions API / UI / Feature 3C | FULL FORENSIC COVERAGE | 5 | `docs/stages/STAGE_03_64_CORPORATE_ACTIONS_API_UI_IMPLEMENTATION.md`, PR #128 |
| 3.65 | Corporate Actions API / UI Closure | MINIMAL / CLOSURE ONLY | 0 | `docs/stages/STAGE_03_65_CORPORATE_ACTIONS_API_UI_CLOSURE.md` |
| 3.66 | Corporate Actions Request Cancellation | FULL FORENSIC COVERAGE | 2 | `docs/stages/STAGE_03_66_CORPORATE_ACTIONS_REQUEST_CANCELLATION_IMPLEMENTATION.md`, PR #130 |
| 3.67 | Request Cancellation Closure | MINIMAL / CLOSURE ONLY | 0 | `docs/stages/STAGE_03_67_CORPORATE_ACTIONS_REQUEST_CANCELLATION_CLOSURE.md` |
| 3.68 | Dividend Calculator | FULL FORENSIC COVERAGE | 6 | `docs/stages/STAGE_03_68_DIVIDEND_CALCULATOR_IMPLEMENTATION.md`, PR #133 |
| 3.69 | Dividend Calculator Closure | MINIMAL / CLOSURE ONLY | 0 | `docs/stages/STAGE_03_69_DIVIDEND_CALCULATOR_CLOSURE.md`, PR #134 |
| 3.70 | Portfolio Position & Cost Basis Planning / ADR-009 | FULL FORENSIC COVERAGE | 2 | `docs/stages/STAGE_03_70_PORTFOLIO_POSITION_COST_BASIS_PLANNING.md`, evidence appendix, ADR-009, PR #137 |
| 3.71 | Portfolio Position & Cost Basis Engine | FULL FORENSIC COVERAGE | 5 | Stage 3.71 implementation/review/rollout/disposition dossiers, PR #138 |
| 3.72 | Portfolio Position Projection / Cost Basis View | STRONG PARTIAL | 4 | Stage 3.72 planning/runtime/disposition dossiers, PRs #142/#143/#145 |
| 3.73 | Portfolio Time Machine | MINIMAL / NO MATERIAL REVIEW FINDINGS RECORDED | 0 | `docs/stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md`, PR #148 |
| 3.74 | Transaction Correction & Reversal | PARTIAL | 4 | `docs/stages/STAGE_03_74_TRANSACTION_CORRECTION_REVERSAL_IMPLEMENTATION.md`, PR #150 |
| 3.75 | Portfolio Cash Flow & Income Truth | HISTORICAL EVIDENCE GAP | 4 | `docs/stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md`, historical PR #152, replacement PR #153 |
| 3.76 | Manual Market Valuation & Portfolio P/L | STRONG PARTIAL | 2 | `docs/stages/STAGE_03_76_MANUAL_MARKET_VALUATION_IMPLEMENTATION.md`, PR #155 |
| 3.77 | Portfolio Money-Weighted Return / XIRR | FULL FORENSIC COVERAGE | 7 | `docs/stages/STAGE_03_77_PORTFOLIO_XIRR_IMPLEMENTATION.md`, PR #157 |

## 7. Stages with no material remediation finding recorded

The following stages are deliberately **not** assigned fictional defect histories:

- **Stage 3.57 — Market Data Provider Boundary.** Contemporary evidence records preventive provider-neutral design and validation boundaries. No material remediation finding is preserved.
- **Stage 3.58 — Market Data Provider Boundary Closure.** Closure/evidence synchronization only; no new material runtime finding.
- **Stage 3.60 — MOEX ISS Runtime / Source Activation Decision.** The material outcome is a source-governance `NO-GO` for shipped/public activation under reviewed rights/cost constraints, not a repaired runtime defect.
- **Stage 3.61 — Corporate Actions Calendar Planning.** Design/source-governance stage; no material implementation remediation finding preserved.
- **Stage 3.65 — Corporate Actions API/UI Closure.** Closure-only stage; no new material defect.
- **Stage 3.67 — Request Cancellation Closure.** Closure-only stage; Stage 3.66 contains the substantive remediation history.
- **Stage 3.69 — Dividend Calculator Closure.** Closure-only stage; Stage 3.68 contains the substantive remediation history.
- **Stage 3.73 — Portfolio Time Machine.** Contemporary evidence records a deliberately minimal reuse design over the Stage 3.72 `asOfDate` projection, including prevention of a second financial engine/API/schema/provider. No material remediation finding is preserved.

Preventive architecture for these stages remains important, but it is classified as design rationale rather than retroactively promoted to a review finding.

## 8. Material forensic records

## Stage 3.59 — MOEX ISS Quote Provider Adapter

### HF-359-01 — Prepublication provider-boundary deterministic hardening

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The first local MOEX adapter candidate still contained weaker deterministic-boundary choices: unnecessary nondeterministic `time.Now()` use in negative-path tests, missing caller-deadline and NUMERIC(28,8)-overflow proof, and a provider-aware constructor that did not yet delegate the canonical base `Service` constructor.
- **Root cause:** The adapter and DI seam were new, and the first candidate had not yet reduced every clock/construction assumption to existing OpenInvest authority seams.
- **Failure / attack scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Project impact:** Tests could remain timing-sensitive, cancellation/overflow invariants under-proven, and future base-constructor initialization could drift.
- **Original / insufficient design:** Use weaker negative-path timing/evidence and construct provider-aware service state separately.
- **Why review rejected or challenged it:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Remove unnecessary nondeterministic time use; add caller-deadline and NUMERIC(28,8)-overflow proof; make `NewServiceWithQuoteProvider` delegate `NewService`.
- **Why the remediation was selected:** Reuse existing clock/service authority and avoid duplicate initialization.
- **Regression coverage:** Focused provider/service tests plus later exact-head Go tests/race/vet.
- **CI / review evidence:** Stage 3.59 dossier §15.1; PR #123; Internal record SHA-256 `cd712cf4a5093c7e35fea47966a6ff41a923e3c0013accae6111a1bdf16d0ab2`.
- **Residual limitations:** The Internal review still missed the later redirect and typed-nil defects; approval is chronological evidence only.

### HF-359-02 — Prepublication transport/provenance fail-closed hardening

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The first local candidate could inherit an HTTP CookieJar and did not yet validate `dataversion` evidence before classifying quote absence.
- **Root cause:** A copied caller `http.Client` can carry ambient cookie state, and absence handling can short-circuit before validating same-response provenance/time evidence.
- **Failure / attack scenario:** A supplied client with cookies could widen the adapter access context; malformed `dataversion` could be hidden behind an empty quote result.
- **Project impact:** Ambient state leakage and false quote-absence classification would weaken the fixed provider boundary.
- **Original / insufficient design:** Copy the client without the final no-cookie guarantee and permit quote absence before complete `dataversion` validation.
- **Why review rejected or challenged it:** The dossier records both as prepublication hardening; exact per-item review prose is not preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Clear inherited CookieJar state and validate required `dataversion` evidence before quote-absence classification.
- **Why the remediation was selected:** Fail closed on malformed source evidence and avoid ambient HTTP state.
- **Regression coverage:** CookieJar/timeout tests and malformed-dataversion-before-absence regression.
- **CI / review evidence:** Stage 3.59 dossier §§3, 8, 10, 15.1.
- **Residual limitations:** No production MOEX runtime/public activation was authorized; the adapter remained dormant.

### HF-359-03 — HTTP redirect could escape the fixed-host provider boundary

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P2
- **Problem:** The provider-owned HTTP client inherited automatic redirect following.
- **Root cause:** The fixed production host was enforced only on the initial request URL; default redirects could send the follow-up request elsewhere.
- **Failure / attack scenario:** A provider 3xx response pointing to another origin could make OpenInvest contact an unreviewed host.
- **Project impact:** Source/use, privacy and trust-boundary assumptions could be violated even though the initial URL was fixed.
- **Original / insufficient design:** Use copied `http.Client` timeout/body bounds but retain default redirect behavior.
- **Why review rejected or challenged it:** Fresh External review demonstrated that fixed-host URL construction is insufficient if redirects can leave the host.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Override `CheckRedirect` with `http.ErrUseLastResponse`; classify the resulting non-200 through existing provider-data semantics.
- **Why the remediation was selected:** The smallest safe policy is no automatic redirect.
- **Regression coverage:** Dedicated redirect regression proves the target receives zero requests; corrected exact-head CI #331 / run `33900381845` 10/10.
- **CI / review evidence:** PR #123: initial head `fe265e9127b44b2fe5899b62b1fc47c8429075e7`, CI #327 green; External COMMENT `5116007026` REQUEST CHANGES; corrected head `4605753d68b56227c646663446700271042cc299`; re-review `5116042645` APPROVED.
- **Residual limitations:** Any future redirect policy requires separate source-boundary review.

### HF-359-04 — Typed-nil `QuoteProvider` could bypass fail-closed unconfigured state

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P2
- **Problem:** An interface containing a typed-nil `QuoteProvider` was non-nil at the interface level and could bypass a simple nil guard.
- **Root cause:** Go interface nil semantics differ from a nil underlying pointer.
- **Failure / attack scenario:** A typed-nil provider whose method dereferences its receiver could be invoked and panic.
- **Project impact:** A normal provider-unavailable condition could become a process-level panic.
- **Original / insufficient design:** Check only `s.quoteProvider == nil`.
- **Why review rejected or challenged it:** Fresh External review demonstrated that interface equality does not cover typed-nil values.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Normalize nil and typed-nil providers in `NewServiceWithQuoteProvider` to the existing unconfigured service state.
- **Why the remediation was selected:** Dependency validity belongs at construction; callers then get deterministic unavailable behavior.
- **Regression coverage:** Typed-nil provider regression would panic if invoked but instead returns `ErrMarketQuoteProviderUnavailable`; full CI green.
- **CI / review evidence:** PR #123 External COMMENT `5116007026`; corrected head `4605753...`; CI #331; re-review `5116042645` APPROVED.
- **Residual limitations:** Future interface-typed dependencies need the same typed-nil awareness if nil is invalid.


## Stage 3.62 — Corporate Action Boundary / Feature 3A

### HF-362-01 — Opaque SourceEventID validation missed Unicode C1 controls

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P3
- **Problem:** `validateOpaqueCorporateActionID` claimed to reject control characters but rejected only C0 controls and DEL.
- **Root cause:** The implementation encoded an ASCII-oriented control check instead of the Unicode semantic rule stated by the contract.
- **Failure / attack scenario:** Provider evidence containing U+0080..U+009F could pass validation.
- **Project impact:** Malformed provenance identifiers could enter the canonical boundary and later surface in logs/evidence handling.
- **Original / insufficient design:** Reject runes `<0x20` and `0x7f` only.
- **Why review rejected or challenged it:** External review compared implementation with the documented opaque-ID contract and found the mismatch.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use `unicode.IsControl` and add a C1 regression.
- **Why the remediation was selected:** Use the language's Unicode semantic category rather than maintain partial ranges.
- **Regression coverage:** Dedicated C1 regression; CI #337 / run `33909927892` 10/10; re-review approved.
- **CI / review evidence:** PR #126 review `5116906103` REQUEST CHANGES; corrected head `97089869e8f3c5ccaf14a9aecd4927d8e2c2eb85`; re-review `5116978636`; evidence verification `5117045812`.
- **Residual limitations:** `SourceEventID` remains opaque/internal; no provider activation or persistence was authorized.

#### Non-material metadata note

The same External review also found that the dossier named `ErrCorporateActionsUnavailable` instead of the actual `ErrCorporateActionsProviderUnavailable`. The text was corrected on the same remediation head. This reconciliation does not count that sentinel-name typo as a second material forensic record because it did not change runtime semantics.


## Stage 3.63 — Calendar + Heatmap Projection / Feature 3B

### HF-363-01 — Supersession could remove an unrelated instrument or action kind

- **Classification:** INTERNAL REVIEW FINDING
- **Severity:** P2
- **Problem:** The first projection candidate did not require a present predecessor to share `InstrumentID` and `Kind` with its superseding event.
- **Root cause:** Graph-shape validation did not include the economic identity dimensions of a supersession edge.
- **Failure / attack scenario:** A dividend for one ticker could supersede a coupon or an event for another ticker and remove that unrelated dated event.
- **Project impact:** Calendar/heatmap truth could silently lose valid events through a malformed cross-entity link.
- **Original / insufficient design:** Allow any well-formed `SupersedesEventID` edge when the predecessor is present.
- **Why review rejected or challenged it:** Internal review identified cross-instrument/cross-kind removal as a demonstrated design gap.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** When the predecessor is present, require same `InstrumentID` and same `Kind`; otherwise fail closed with `ErrInvalidCorporateActionProjection`.
- **Why the remediation was selected:** Preserve partial-batch support while preventing cross-entity semantic corruption.
- **Regression coverage:** Cross-instrument/kind, fork and cycle regressions; focused statement coverage 100%; CI #339 / run `33914229650` 10/10.
- **CI / review evidence:** Stage 3.63 dossier §9; PR #127; External COMMENT `5117337573` APPROVED.
- **Residual limitations:** Cross-batch persistence/revision uniqueness remains deferred; missing predecessors are allowed.


## Stage 3.64 — Corporate Actions API / UI / Feature 3C

### HF-364-01 — Provider fan-out was not bounded before invocation

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The early API/UI candidate did not yet freeze a hard request cardinality cap before provider invocation.
- **Root cause:** A public list query can multiply provider work unless admission is bounded before the provider seam.
- **Failure / attack scenario:** A client could submit a very large instrument list and force an eventual provider adapter to perform excessive work.
- **Project impact:** Performance/cost and provider-rate-rights risk before Feature 3D.
- **Original / insufficient design:** Accept the domain query without the final `<=50` public bound.
- **Why review rejected or challenged it:** The dossier records provider fan-out bounding as resolved prepublication hardening; separate reviewer prose is not preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Reject more than 50 instruments before provider invocation.
- **Why the remediation was selected:** Fail-fast admission preserves explicit request scope and makes provider demand auditable.
- **Regression coverage:** HTTP tests prove oversized input is rejected before provider invocation; External review rechecked cost controls; CI #344 green.
- **CI / review evidence:** Stage 3.64 dossier §§3, 12, 14; PR #128.
- **Residual limitations:** The 50-instrument bound is an application contract, not a provider-specific production rate policy.

### HF-364-02 — Frontend stale responses could overwrite a newer Corporate Actions query

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The first UI candidate did not yet have the final request-generation stale-result guard.
- **Root cause:** Asynchronous completion order is independent of user query/update order.
- **Failure / attack scenario:** An older request could finish after a newer one and replace the latest calendar/heatmap or source-unavailable state.
- **Project impact:** Users could see data for the wrong instruments/date window.
- **Original / insufficient design:** Rely on request completion without a final generation identity check.
- **Why review rejected or challenged it:** The dossier records stale request-generation protection as prepublication hardening; per-finding reviewer narrative is not preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use AbortController plus monotonically invalidated request generation; commit state only for the current generation.
- **Why the remediation was selected:** Independent transport-cancel and stale-completion guards cover different races.
- **Regression coverage:** Component tests cover replacement/newer-result wins and stale completion suppression.
- **CI / review evidence:** Stage 3.64 dossier §§10, 12; final pre-evidence CI #344 / run `33927609434`.
- **Residual limitations:** Component-unmount cancellation was intentionally retained as a non-blocking P3 and closed in Stage 3.66.

### HF-364-03 — Provider-owned source identity and caching rights were too exposed/implicit

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The early candidate still exposed provider-owned `SourceEventID` in the public DTO and had not yet frozen `Cache-Control: no-store` while source caching rights were unapproved.
- **Root cause:** The API/UI surface was designed before a real source/use contract existed, so provider-specific evidence and retention rights could accidentally become public contract.
- **Failure / attack scenario:** A future adapter could lock opaque provider IDs into clients or allow caching despite source terms not authorizing it.
- **Project impact:** Provider coupling and source-rights violations would be harder to reverse.
- **Original / insufficient design:** Expose source event identity and rely on ordinary caching defaults.
- **Why review rejected or challenged it:** The dossier records removal of `SourceEventID` and `no-store` as prepublication hardening; exact rationale per sub-item is not separately recorded.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Remove `SourceEventID` from the public DTO and add regression protection; send `Cache-Control: no-store` on every endpoint response.
- **Why the remediation was selected:** Keep the public contract provider-neutral and conservative until exact source rights are reviewed.
- **Regression coverage:** OpenAPI/DTO/client regressions and forbidden-surface scans; External review rechecked identity minimization/no-cache.
- **CI / review evidence:** Stage 3.64 dossier §§6, 7, 12, 14.
- **Residual limitations:** Future source-specific attribution/cache policy requires separate approval.

### HF-364-04 — Provider over-return could widen the public date scope

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The first candidate did not yet explicitly bound final dated output after supersession resolution.
- **Root cause:** Providers can return events outside the requested window; filtering at the wrong point can either leak scope or break supersession semantics.
- **Failure / attack scenario:** An out-of-window event could appear publicly, or premature filtering could hide a cancellation needed to remove an in-window predecessor.
- **Project impact:** The response could exceed requested scope or show obsolete events.
- **Original / insufficient design:** Project provider results without the final reviewed ordering of full-batch supersession resolution followed by date filtering.
- **Why review rejected or challenged it:** The dossier records effective-date output filtering as resolved prepublication hardening; detailed reviewer narrative is not preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Resolve supersession over the complete validated batch, then filter Calendar/Heatmap output to `from <= effectiveDate <= to`.
- **Why the remediation was selected:** Preserve correction/cancellation integrity while enforcing request scope.
- **Regression coverage:** Provider over-return regression plus Stage 3.63 projection tests; CI #344 green.
- **CI / review evidence:** Stage 3.64 dossier §§3, 12, 14.
- **Residual limitations:** Undated current evidence remains outside dated projection by design.

### HF-364-05 — Frontend component test environment did not match Next/React browser assumptions

- **Classification:** CI / TEST-ENVIRONMENT REMEDIATION FAMILY
- **Severity:** CI blockers; product severity not recorded
- **Problem:** The first published component-test setup failed on CSS-module import; subsequent review/CI exposed ReactDOM import-before-globals and missing browser `self` assumptions.
- **Root cause:** The Node/tsx test runner lacked the CSS/browser environment that the styled Next.js component and `Link` expected.
- **Failure / attack scenario:** CI component tests fail before exercising feature behavior.
- **Project impact:** Required protected checks could not prove the UI regression surface.
- **Original / insufficient design:** Use the existing Node component test path without a CSS loader and with static ReactDOM imports before JSDOM globals.
- **Why review rejected or challenged it:** CI #341 and #343 demonstrated concrete failures; review also identified import-order fragility.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Add a test-only CSS-module loader; install JSDOM/browser globals before dynamic React/ReactDOM/component imports; provide `self: dom.window`; run tests sequentially.
- **Why the remediation was selected:** Confine remediation to test infrastructure with no production runtime/dependency expansion.
- **Regression coverage:** CI #341 exposed CSS failure; commits `cee7e1...`, `78e7aaa...`; CI #343 exposed `self`; final head `9bbcf6d...`; CI #344 / run `33927609434` 10/10.
- **CI / review evidence:** PR #128 External review `5118470329` APPROVED; evidence head `f4631c04...`, CI #345 / run `33927918258` green.
- **Residual limitations:** Test shims are evidence infrastructure only.


## Stage 3.66 — Corporate Actions Request Cancellation

### HF-366-01 — In-flight Corporate Actions request survived component unmount

- **Classification:** CARRY-FORWARD P3 / IMPLEMENTATION HARDENING
- **Severity:** P3
- **Problem:** Stage 3.64 protected query-change/resubmit races but did not explicitly abort the active request when the owning component unmounted.
- **Root cause:** Request lifetime was tied to query generations, not component lifetime.
- **Failure / attack scenario:** Navigation/unmount during a real request could leave network/provider work in flight and a late continuation still needed suppression.
- **Project impact:** Future rate/cost-limited provider quota/resources could be wasted and lifecycle correctness weakened.
- **Original / insufficient design:** Abort old requests on input change/resubmit and rely on generation guards; no mount cleanup.
- **Why review rejected or challenged it:** Stage 3.64 External review explicitly retained this as a non-blocking P3 before Feature 3D.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Unmount cleanup increments request generation first, aborts the controller, clears it, and acceptance checks both `signal.aborted` and generation identity.
- **Why the remediation was selected:** Smallest owner-local fix; preserves shared client semantics and provides redundant race protection.
- **Regression coverage:** Component tests prove active signal aborted on unmount, aborted requests do not surface false errors, replacement wins, genuine errors remain visible; CI #347 / run `33952598235` 10/10.
- **CI / review evidence:** Stage 3.66 dossier; PR #130; External COMMENT `5550298926` APPROVED.
- **Residual limitations:** Shared client still maps low-level fetch exceptions generically; broader typed cancellation is separate scope.

### HF-366-02 — Unmount regression test could leak React root after early assertion failure

- **Classification:** INTERNAL TEST-QUALITY FINDING
- **Severity:** P3
- **Problem:** The initial unmount test did not guarantee React root cleanup if an early assertion failed.
- **Root cause:** Teardown was not fail-safe around intermediate assertions.
- **Failure / attack scenario:** One failed test could contaminate later sequential component tests or leave lifecycle state/resources alive.
- **Project impact:** Evidence quality and reproducibility would be weaker even though production runtime was not defective.
- **Original / insufficient design:** Cleanup only on the ordinary success path.
- **Why review rejected or challenged it:** Internal review recorded a non-runtime P3 test-quality finding.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Add fail-safe teardown and retain sequential isolation; add a genuine network-failure regression so abort suppression cannot hide normal transport failures.
- **Why the remediation was selected:** Cancellation tests must themselves be deterministic and distinguish abort from genuine failure.
- **Regression coverage:** Focused preflight plus exact-head frontend tests under CI #347.
- **CI / review evidence:** Stage 3.66 dossier §13.2; Internal review SHA-256 `70e1b8fb0f40941137d5350831aa1bdb0c35db135ba57610ac5e8bbade5c426c`.
- **Residual limitations:** No runtime finding remained after remediation.


## Stage 3.68 — Dividend Calculator

### HF-368-01 — Feature-specific frontend API client duplicated canonical HTTP authority

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The first calculator candidate introduced a feature-specific frontend HTTP client instead of canonical `frontend-next/src/common/api/openinvest.ts`.
- **Root cause:** Feature convenience started duplicating shared transport/auth/error/idempotency conventions.
- **Failure / attack scenario:** Calculator behavior could drift from shared auth/error/retry handling.
- **Project impact:** Frontend architecture would fragment into multiple transport authorities.
- **Original / insufficient design:** Add a calculator-specific API client module.
- **Why review rejected or challenged it:** Internal hardening removed it; detailed per-item reviewer rationale is not separately preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use only canonical `openinvest.ts`; calculator component remains presentation/orchestration only.
- **Why the remediation was selected:** Keep one HTTP client authority and reuse established semantics.
- **Regression coverage:** Architecture/scope scans, frontend tests, External review, CI #352 green.
- **CI / review evidence:** Stage 3.68 dossier Internal Review Evidence.
- **Residual limitations:** Feature helpers may compose typed calls, but transport authority remains shared.

### HF-368-02 — Anonymous calculator replay identity/privacy model was too weakly specified

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The original anonymous replay subject design and uniqueness/retention wording did not yet provide the final domain-separated technical principal/privacy semantics.
- **Root cause:** A public idempotent endpoint needs replay isolation without creating a user profile or overclaiming mathematical uniqueness.
- **Failure / attack scenario:** A shared/poorly scoped subject could collide conceptually with other principals; docs could claim injective uniqueness or zero persistence despite exact-response retention.
- **Project impact:** Audit/privacy semantics would be misleading and could set unsafe precedent.
- **Original / insufficient design:** Use the original anonymous subject design plus stronger uniqueness/zero-persistence language.
- **Why review rejected or challenged it:** Internal review required domain-separated technical replay scope and corrected privacy/retention wording.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Derive deterministic UUIDv8 scope from SHA-256(domain separator + validated key); keep original key in replay uniqueness; describe collision resistance, not absolute uniqueness; disclose existing 24-hour exact-response persistence.
- **Why the remediation was selected:** Technical replay identity without user/cookie/IP fingerprinting, with accurate cryptographic/retention claims.
- **Regression coverage:** Replay scope/privacy tests, PostgreSQL integration, External review; CI #352/#353.
- **CI / review evidence:** Stage 3.68 dossier Public idempotency / Privacy and retention / Internal Review Evidence.
- **Residual limitations:** Accepted only for this public calculator; not precedent for sensitive anonymous surfaces.

### HF-368-03 — Replay/rate-limit ordering could weaken exact replay and create writes for denied fresh commands

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The early flow did not yet place exact replay resolution before fresh-command limiting and fresh admission before writable replay reservation.
- **Root cause:** Idempotency and abuse-admission concerns were ordered around the write path rather than semantic authority.
- **Failure / attack scenario:** A completed exact retry could receive 429; a denied fresh request could create a provisional dedup row; races could hide stronger replay/conflict state.
- **Project impact:** Breaks exact-response replay expectations and increases public write amplification.
- **Original / insufficient design:** Run limiter/reservation before complete read-only replay resolution.
- **Why review rejected or challenged it:** Internal hardening moved exact replay ahead of limiting and admission ahead of writable reservation.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use read-only replay lookup first; return completed/conflict/in-flight; only no-state requests enter bounded in-memory admission; denied requests race-recheck read-only; only admitted fresh requests reserve transactionally.
- **Why the remediation was selected:** Prioritize idempotent truth and keep denied fresh requests read-only.
- **Regression coverage:** PostgreSQL replay integration/cleanup, limiter/race tests, Go/race CI, External review.
- **CI / review evidence:** Stage 3.68 dossier Public idempotency / Internal Review Evidence.
- **Residual limitations:** Limiter is in-process, not a distributed/edge DoS shield.

### HF-368-04 — Backend trim-first ticker admission contradicted the OpenAPI grammar

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The first calculator backend accepted ticker input after trimming whitespace.
- **Root cause:** Transport normalization silently widened an exact canonical grammar.
- **Failure / attack scenario:** A padded ticker rejected by contract could be accepted by runtime and normalized to another value.
- **Project impact:** API/runtime drift and ambiguous replay/request-hash identity.
- **Original / insufficient design:** Trim ticker before canonical validation.
- **Why review rejected or challenged it:** Internal hardening removed trim-first acceptance so OpenAPI grammar fails closed.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Validate the raw supplied ticker under canonical grammar; do not silently normalize padding.
- **Why the remediation was selected:** Contract-first exactness and deterministic request identity.
- **Regression coverage:** HTTP/validation vectors; route/OpenAPI parity; CI #352.
- **CI / review evidence:** Stage 3.68 dossier Internal Review Evidence.
- **Residual limitations:** No new normalization policy introduced.

### HF-368-05 — Browser retry identity was not retained for ambiguous failed unchanged-payload attempts

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The early Web flow could release/rotate the Idempotency-Key after a failed attempt even when payload had not changed.
- **Root cause:** Retry identity was treated as per-click state rather than unresolved user intent.
- **Failure / attack scenario:** Transport uncertainty/in-flight response followed by retry could become a second command instead of exact replay.
- **Project impact:** Weakens idempotency continuity.
- **Original / insufficient design:** Generate/release key around each attempt.
- **Why review rejected or challenged it:** Internal hardening required retaining identity for failed unchanged payload and releasing only on success/input change.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Bind key to current payload; retain across failed attempts; clear on confirmed success; rotate when any input changes.
- **Why the remediation was selected:** Matches established retry-intent semantics without duplicating business logic.
- **Regression coverage:** Frontend retry tests; External review; CI #352.
- **CI / review evidence:** Stage 3.68 dossier Web architecture / Internal Review Evidence.
- **Residual limitations:** Key identifies technical retry intent, not a user account.

### HF-368-06 — Money presentation used scale-8 trim-only output instead of exact two-decimal Half-Even UI semantics

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** Initial UI displayed returned money by trimming scale-8 strings rather than applying reviewed two-decimal money presentation.
- **Root cause:** Canonical calculation scale and human display precision were conflated.
- **Failure / attack scenario:** Values requiring third-decimal rounding could display inconsistently or with excessive precision.
- **Project impact:** User-visible financial presentation could disagree with project monetary formatting.
- **Original / insufficient design:** Trim canonical scale-8 Decimal for display.
- **Why review rejected or challenged it:** Internal hardening changed the presentation path; detailed reviewer prose is not separately preserved.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Implement exact decimal-string two-decimal Half-Even money display; percentage text uses decimal-string shifting.
- **Why the remediation was selected:** Preserve exact backend truth and deterministic presentation without client financial recomputation.
- **Regression coverage:** Frontend formatting tests; External review explicitly checked presentation; CI #352.
- **CI / review evidence:** Stage 3.68 dossier Financial semantics / Web architecture / Internal Review Evidence.
- **Residual limitations:** Presentation rounding does not change backend scale-8 results or replay body.


## Stage 3.70 — Portfolio Position & Cost Basis Planning / ADR-009

### HF-370-01 — WAC scale-8 state model was self-contradictory

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P1
- **Problem:** The first planning candidate treated `quantity + acquisitionBasis` as canonical state, re-derived WAC from rounded basis/quantity, and also required partial SELL to preserve WAC.
- **Root cause:** Scale-8 rounding makes rounded acquisition basis non-invertible for fractional quantities.
- **Failure / attack scenario:** For `quantity=0.20000000`, WAC `275.12345678` gives basis `55.02469136`; dividing basis by quantity yields `275.12345680`, silently repricing WAC.
- **Project impact:** Financial state would drift after partial SELL and violate the stated WAC methodology.
- **Original / insufficient design:** Canonical quantity+basis state with inverse WAC reconstruction.
- **Why review rejected or challenged it:** Fresh External review demonstrated the mathematical contradiction and classified it P1.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Make `quantity + weightedAverageCost` authoritative; derive `acquisitionBasis = Round8HalfEven(quantity × WAC)`; never inverse-reprice after SELL; add the fractional non-invertibility vector.
- **Why the remediation was selected:** Matches the rule that partial SELL preserves WAC and makes rounding one-way/deterministic.
- **Regression coverage:** ADR-009 fractional witness; Stage 3.71 runtime vectors; CI #357 and fresh External re-review green.
- **CI / review evidence:** PR #137 initial head `a5c7a507a586271b7117580585c88191693144e3`; CI #356 / run `33998907479`; remediation head `2af826392e081eb07d3234be1c94833bae768aa7`; CI #357 / run `34000772097`.
- **Residual limitations:** Acquisition basis remains trade-price basis, not tax basis; non-invertibility is intentional.

### HF-370-02 — Planned transaction conflict contract overstated existing 409/429 API semantics

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P2
- **Problem:** The first plan described a generic business-conflict surface/429 behavior that the frozen POST transaction contract did not actually expose.
- **Root cause:** Planning language assumed stronger existing API semantics instead of reconciling with canonical OpenAPI.
- **Failure / attack scenario:** Stage 3.71 could ship insufficient-position conflict without an approved schema or falsely claim 429 support.
- **Project impact:** Contract-first governance would be violated and clients could receive undocumented behavior.
- **Original / insufficient design:** Treat existing POST 409 as generic enough and describe 429 as existing.
- **Why review rejected or challenged it:** External review compared the plan with the frozen contract and found the mismatch.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Plan one narrow POST-transaction 409 contract change: `TransactionConflict` preserves `IDEMPOTENCY_CONFLICT` and adds `INSUFFICIENT_POSITION_QUANTITY`; remove false 429 claim.
- **Why the remediation was selected:** Adds only the conflict needed for SELL while preserving HTTP 409/envelope.
- **Regression coverage:** Stage 3.71 OpenAPI/HTTP regressions later prove both conflict codes; CI #357/re-review approved.
- **CI / review evidence:** Stage 3.70 evidence appendix §17; PR #137.
- **Residual limitations:** The separate P3 stale SOT version in `DOCUMENT_INDEX.md` is retained as a non-material registry note, not a separate material record.

#### Non-material registry note

The first External Stage 3.70 review also found a P3 registry-version mismatch: `DOCUMENT_INDEX.md` identified SOT-001 as `1.4.63` while the candidate Source of Truth was `1.4.90`. The remediation synchronized the registry. It is preserved here as metadata drift, not counted among the two material financial/API findings.


## Stage 3.71 — Portfolio Position & Cost Basis Engine

### HF-371-01 — Exact-index readiness/backfill SQL compared `name[]` with `text[]`

- **Classification:** ADVERSARIAL REVIEW FINDING
- **Severity:** Substantive; severity label not preserved
- **Problem:** The exact-index catalog query used `array_agg(pg_attribute.attname ...) = text[]`, which fails because `attname` is PostgreSQL type `name`.
- **Root cause:** Static reasoning missed the real array element type mismatch.
- **Failure / attack scenario:** Backfill/readiness fails with `operator does not exist: name[] = text[]` even when schema/data are otherwise valid.
- **Project impact:** Stage 3.71 cannot safely populate/activate; non-executing tests could give false assurance.
- **Original / insufficient design:** Compare aggregated `attname` values directly with a `text[]` literal.
- **Why review rejected or challenged it:** Adversarial review exercised/inspected the real PostgreSQL path and found the exact error.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Cast `pg_attribute.attname::text` before ordered array comparison in both backfill and runtime readiness.
- **Why the remediation was selected:** Preserve strict exact-index verification without weakening readiness.
- **Regression coverage:** Real PostgreSQL backfill coverage and `Stage371Ready` integration regression; CI #381 all 10 contexts.
- **CI / review evidence:** Stage 3.71 review evidence §2.1; PR #138 adversarial review `5125058519`; External `5125093621` APPROVED.
- **Residual limitations:** Missing/invalid/not-ready/partial/expression/differently ordered indexes still fail closed.

### HF-371-02 — Snapshot summary selected financial truth by wall-clock `calculated_at`

- **Classification:** ADVERSARIAL REVIEW FINDING
- **Severity:** Substantive; severity label not preserved
- **Problem:** Legacy summary selected newest same-date snapshot by `calculated_at DESC`.
- **Root cause:** `command.Now` is request-derived before the serialized portfolio DB lock; serialized financial order can differ from timestamp order.
- **Failure / attack scenario:** A later committed snapshot version can carry an earlier timestamp and be ignored.
- **Project impact:** Public summary could return stale quantity/acquisition-basis truth despite correct ledger/snapshot history.
- **Original / insufficient design:** Use calculated timestamp as primary newest-snapshot authority.
- **Why review rejected or challenged it:** Review recognized system time as technical metadata, not serialized financial order.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Order by snapshot date, Stage 3.71 methodology priority, `snapshot_version DESC`, then timestamp/id as tie-breakers.
- **Why the remediation was selected:** Use business date/methodology/serialized version as financial authority.
- **Regression coverage:** Regression inverts timestamps and includes Stage 3.02 version 99; expected Stage 3.71 version 2 remains selected.
- **CI / review evidence:** Stage 3.71 review evidence §2.2; CI #381; External review approved.
- **Residual limitations:** Timestamp remains deterministic tie-breaker only.

### HF-371-03 — Populate-to-runtime cutover allowed a legacy NULL-sequence write between backfill and readiness

- **Classification:** ROLLOUT REVIEW FINDING
- **Severity:** Substantive; severity label not preserved
- **Problem:** Nullable Expand migration plus owner-only backfill left an interval where legacy runtime could append `ledger_sequence=NULL` after backfill.
- **Root cause:** Schema compatibility kept legacy writes valid, but the first rollout wording did not maintain write quiescence through new-runtime readiness.
- **Failure / attack scenario:** Backfill succeeds; lock releases; legacy write inserts NULL; readiness fails; rerun backfill refuses partial population.
- **Project impact:** Cutover becomes stuck. Financial corruption is prevented by fail-closed gates, but activation is operationally incomplete.
- **Original / insufficient design:** Quiesce/populate/deploy without explicitly requiring continuous quiescence through readiness.
- **Why review rejected or challenged it:** Final rollout review identified the gap between Populate completion and Stage 3.71 activation.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Maintain portfolio-write quiescence from before backfill through deployment and successful readiness; abort cutover on any legacy write rather than ad hoc repair.
- **Why the remediation was selected:** Preserve deterministic backfill and fail-closed activation without rewriting accepted ledger history.
- **Regression coverage:** Rollout amendment plus readiness/backfill integration evidence.
- **CI / review evidence:** Stage 3.71 review evidence §2.3; rollout amendment.
- **Residual limitations:** Quiescence is an operational deployment requirement; no online distributed migration is claimed.

### HF-371-04 — Attempted second machine-bound `staged_rollout` authority conflicted with single-authority migration policy

- **Classification:** CI / GOVERNANCE FINDING
- **Severity:** CI blocker
- **Problem:** A governance-only change attempted to add a second machine-bound `staged_rollout` authority.
- **Root cause:** Deployment rollout evidence was being encoded in a machine-enforced authority model designed around one authority source.
- **Failure / attack scenario:** Validator must either reject the policy or be weakened to accept multiple authorities.
- **Project impact:** Weakening would make migration authority ambiguous; leaving it blocks protected CI.
- **Original / insufficient design:** Add second `staged_rollout` authority.
- **Why review rejected or challenged it:** CI #380 rejected it; project explicitly chose not to weaken the validator.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Restore policy manifest byte-for-byte to valid single-authority model; keep rollout requirements in amendment/evidence docs.
- **Why the remediation was selected:** Separate machine-bound migration authority from operational rollout instructions.
- **Regression coverage:** CI #380 failure followed by exact-head CI #381 success.
- **CI / review evidence:** Stage 3.71 review evidence §6; PR #138.
- **Residual limitations:** Future authority-model changes require separate governance review.

### HF-371-05 — Internal/adversarial evidence was published before the formal External verdict

- **Classification:** HISTORICAL GOVERNANCE DEVIATION
- **Severity:** Governance noncompliance; dispositioned
- **Problem:** A pre-External adversarial review/evidence document was already present on the published PR before formal External review.
- **Root cause:** The temporal withholding control in `REVIEW_WORKFLOW.md` was missed.
- **Failure / attack scenario:** External reviewer could see earlier review evidence before issuing the formally required fresh External verdict.
- **Project impact:** Repository cannot prove strict pre-verdict evidence withholding for that interval.
- **Original / insufficient design:** Publish adversarial evidence before formal External phase.
- **Why review rejected or challenged it:** Current workflow requires Internal evidence withheld until after External; the temporal condition cannot be repaired retroactively.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Preserve chronology, perform fresh External review without using earlier findings as supporting evidence, publish post-External evidence honestly, and disposition the irreversible deviation with residual risk accepted.
- **Why the remediation was selected:** Non-retroactive evidence is more auditable than rewriting history.
- **Regression coverage:** External review `5125093621` explicitly states non-reliance; evidence verification `5125104746`; separate disposition.
- **CI / review evidence:** Stage 3.71 review evidence §8 and `STAGE_03_71_GOV_01_HISTORICAL_GOVERNANCE_DEVIATION_DISPOSITION.md`.
- **Residual limitations:** Historical compliance remains noncompliant after disposition; technical semantics unaffected.


## Stage 3.72 — Portfolio Position Projection / Cost Basis View

### HF-372-01 — Acquisition-basis allocation was undefined for nonempty zero-total-basis projection

- **Classification:** INTERNAL REVIEW FINDING
- **Severity:** Material; `INT-372-F1`
- **Problem:** The first plan did not define `acquisitionBasisWeight` when open positions exist but every scale-8 acquisition basis rounds to zero.
- **Root cause:** The allocation formula assumed a positive denominator.
- **Failure / attack scenario:** Tiny valid open positions can produce `totalAcquisitionBasis=0.00000000`, making each weight `0/0`.
- **Project impact:** Runtime/API could fabricate zero, divide by zero, or diverge between backend/UI.
- **Original / insufficient design:** Always return Decimal allocation for every open item.
- **Why review rejected or challenged it:** Internal review identified the zero-denominator case as a real financial contract gap.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Make `acquisitionBasisWeight` `Decimal | null`; null only for nonempty-open-position/zero-total-basis.
- **Why the remediation was selected:** Represent mathematical undefined state explicitly.
- **Regression coverage:** Planning vectors and Stage 3.72 runtime/OpenAPI tests later verify zero-denominator/null behavior.
- **CI / review evidence:** Stage 3.72 planning dossier Internal evidence; `INT-372-F1`.
- **Residual limitations:** Independent rounded weights are not force-normalized to exactly one.

### HF-372-02 — Stage 3.71 closure wording contradicted a separately gated Stage 3.72 plan

- **Classification:** INTERNAL REVIEW FINDING
- **Severity:** Material; `INT-372-F2`
- **Problem:** Canonical Stage 3.71 lifecycle wording still broadly prohibited Stage 3.72 scope while the new plan attempted to define it.
- **Root cause:** Closure/non-scope wording was treated as a durable prohibition rather than a successor gate boundary.
- **Failure / attack scenario:** Repository documents could simultaneously say Stage 3.72 is planned and Stage 3.71 forbids it.
- **Project impact:** Authorization provenance becomes ambiguous.
- **Original / insufficient design:** Carry forward the broad Stage 3.71 prohibition unchanged.
- **Why review rejected or challenged it:** Internal review identified the authorization contradiction.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Clarify Stage 3.71 is complete and Stage 3.72 planning is a separate merge-activated decision; runtime remains separately unauthorized.
- **Why the remediation was selected:** Preserve historical boundary while allowing an explicit successor gate.
- **Regression coverage:** Documentation consistency review and later lifecycle closure.
- **CI / review evidence:** Stage 3.72 planning dossier; `INT-372-F2`.
- **Residual limitations:** Planning merge itself does not authorize runtime/provider work.

### HF-372-03 — Published Draft still labeled itself PREPUBLICATION

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P3; `EXT-372-P3-01`
- **Problem:** Top-level status remained `PREPUBLICATION CANDIDATE` after publication in Draft PR #142.
- **Root cause:** Lifecycle metadata was not advanced when temporal state changed.
- **Failure / attack scenario:** Reader sees false current-state metadata in the published PR.
- **Project impact:** Audit chronology and later evidence gates are weakened.
- **Original / insufficient design:** Keep prepublication label unchanged after publication.
- **Why review rejected or challenged it:** External review compared file status with actual PR state and returned REQUEST CHANGES.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Change only lifecycle label to the published Draft state without changing implementation authorization.
- **Why the remediation was selected:** Minimal metadata correction restores fact without altering semantics.
- **Regression coverage:** Fresh re-review verified lifecycle state and no semantic drift.
- **CI / review evidence:** PR #142 review `5125938134`; remediation head `9bc2229625a47df26244421002e7f45d93a96cea`; CI #390 / run `34045656905`; re-review `5125951620` APPROVED.
- **Residual limitations:** Historical earlier labels remain true for their earlier commits.

### HF-372-04 — Stage 3.71 lifecycle text remained future-tense after closure had already merged

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P3; `EXT-372-P3-02`
- **Problem:** ROADMAP/SOURCE_OF_TRUTH lines edited by PR #142 still said Stage 3.71 closure would become complete once merged, although PR #140 was already merged.
- **Root cause:** Registry text was not synchronized with completed protected history.
- **Failure / attack scenario:** Publishing Stage 3.72 would also republish known-false Stage 3.71 current state.
- **Project impact:** Canonical lifecycle registries drift from actual protected history.
- **Original / insufficient design:** Leave future-tense Stage 3.71 closure wording.
- **Why review rejected or challenged it:** External review required correction of known-false current metadata in touched canonical registry lines.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Replace future-tense wording with actual Stage 3.71 closure evidence while keeping Stage 3.72 planning/runtime separate.
- **Why the remediation was selected:** Touched canonical registries must describe current state accurately.
- **Regression coverage:** External re-review `5125951620` verified current-state metadata.
- **CI / review evidence:** PR #142 initial review `5125938134`; CI #390; re-review approved.
- **Residual limitations:** A separate Stage 3.72 withholding deviation is preserved/dispositioned as governance chronology rather than counted as a fifth material feature record.

#### Historical governance chronology note — STAGE-03-72-GOV-01

Stage 3.72 planning also missed the workflow's temporal requirement to withhold the current Internal verdict until the External verdict. The detailed Internal findings were not published early, but the Draft PR body exposed `Internal review result: APPROVED`. The later disposition explicitly preserves:

```text
HISTORICAL_COMPLIANCE = NONCOMPLIANT
DISPOSITION = HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
```

The disposition does not recreate the missed temporal property and does not turn the historical interval into compliant history. This event is preserved in its dedicated governance-disposition dossier and is not counted as a fifth Stage 3.72 feature finding.


## Stage 3.74 — Transaction Correction & Reversal

### HF-374-01 — Correction command response exposed transient raw-entry status instead of logical `CORRECTED`

- **Classification:** PRE-MERGE REVIEW FINDING
- **Severity:** Severity not recorded
- **Problem:** The correction command response initially exposed the raw-entry selector's transient `ACTIVE` interpretation rather than logical revision state `CORRECTED`.
- **Root cause:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Failure / attack scenario:** A successful correction could appear to clients as an ordinary active transaction.
- **Project impact:** Client/UI truth would disagree with the effective logical transaction model.
- **Original / insufficient design:** Return selector's raw/transient status.
- **Why review rejected or challenged it:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Map correction response to logical current revision status `CORRECTED`.
- **Why the remediation was selected:** Align command response with one-row-per-logical-transaction public model.
- **Regression coverage:** Final review and HTTP/DTO witnesses; CI #435 / run `34117662576`.
- **CI / review evidence:** Stage 3.74 dossier post-merge closure; PR #150 final review comment `5570125403`.
- **Residual limitations:** Original root cause/reviewer rationale is not preserved beyond final-review summary.

### HF-374-02 — Snapshot overflow error classification drifted from canonical `ErrInvalidInput`

- **Classification:** PRE-MERGE REVIEW FINDING
- **Severity:** Severity not recorded
- **Problem:** Stage 3.74 effective-ledger snapshot handling did not initially preserve the established overflow error classification.
- **Root cause:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Failure / attack scenario:** A cumulative NUMERIC(28,8) overflow could map differently after correction/reversal than under Stage 3.29.
- **Project impact:** Public/API error semantics drift for the same financial storage-bound failure.
- **Original / insufficient design:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Why review rejected or challenged it:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Preserve canonical `ErrInvalidInput` for snapshot overflow.
- **Why the remediation was selected:** Correction/reversal should not redefine established storage-bound errors.
- **Regression coverage:** Existing Stage 3.29 overflow tests plus Stage 3.74 final suite; CI #435.
- **CI / review evidence:** Stage 3.74 dossier; PR #150 comment `5570125403`.
- **Residual limitations:** Detailed initial implementation and reviewer rationale are not preserved.

### HF-374-03 — Replay/idempotency evidence and browser retry identity were insufficient

- **Classification:** PRE-MERGE REVIEW FINDING FAMILY
- **Severity:** Severity not recorded
- **Problem:** Replay witnesses used invalid request/trace metadata, and Edit/Reverse retries initially did not fully reuse the Stage 3.32 principal-scoped retry journal for the same ambiguous intent.
- **Root cause:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Failure / attack scenario:** Tests could miss production replay rules; ambiguous retry could receive a new key and append another repair command.
- **Project impact:** Weak evidence plus duplicate immutable correction/reversal risk.
- **Original / insufficient design:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Why review rejected or challenged it:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use production-shaped request/trace metadata; reuse distinct correction/reversal journal scopes, retaining key across unresolved retries and rotating on changed intent/success/rejected client outcome.
- **Why the remediation was selected:** Exercise the real replay contract and preserve idempotent user intent.
- **Regression coverage:** Replay integration and frontend retry tests; CI #435.
- **CI / review evidence:** Stage 3.74 dossier and PR #150 final review comment `5570125403`.
- **Residual limitations:** Separate severities/root causes for the sub-items are not recorded.

### HF-374-04 — HTTP and frontend contract evidence was not robust enough

- **Classification:** PRE-MERGE TEST/CONTRACT EVIDENCE FINDING FAMILY
- **Severity:** Severity not recorded
- **Problem:** Final review required robust frontend multiline assertions and dedicated HTTP witnesses for PATCH/DELETE routing, DTO preservation, stale-revision 409, statuses, and required nested `corrected.settlementDate`.
- **Root cause:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Failure / attack scenario:** Handler wiring or nested-settlement drift could pass broader tests; brittle assertions could fail for irrelevant formatting.
- **Project impact:** Frozen correction/reversal API could drift unnoticed.
- **Original / insufficient design:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Why review rejected or challenged it:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Harden frontend assertions and add dedicated HTTP tests for routing/DTO/error/status/settlement-date semantics.
- **Why the remediation was selected:** Target the exact public boundary without changing product design.
- **Regression coverage:** Dedicated HTTP witnesses and frontend contract tests; CI #435 10/10.
- **CI / review evidence:** Stage 3.74 dossier; PR #150 comment `5570125403`; merge `0580bf7e98c532202f84bbf9ceacd97aedbe4140`.
- **Residual limitations:** Root cause, second-order alternatives and reviewer rationale are NOT RECORDED.


## Stage 3.75 — Portfolio Cash Flow & Income Truth

### HF-375-01 — OpenAPI allowed negative component aggregates that runtime defined as non-negative magnitudes

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P1
- **Problem:** `PortfolioCashFlowTotals` exposed deposits, withdrawals, buy/sell legs, gross income, fees and taxes as signed `Money` although runtime semantics require non-negative component magnitudes.
- **Root cause:** The public schema reused a generic signed money type instead of the stronger invariant already implemented by the accumulator.
- **Failure / attack scenario:** Clients/examples could treat negative component values as valid even though canonical runtime never permits them.
- **Project impact:** API-first contract was weaker than runtime financial truth and could allow future drift.
- **Original / insufficient design:** Use signed `Money` for all totals.
- **Why review rejected or challenged it:** External review classified this P1 because the public financial invariant must match canonical runtime truth.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use `NonNegativeMoney` for the eight components; keep the three net fields signed; add validator negative/positive witnesses.
- **Why the remediation was selected:** Schema now exactly mirrors accumulator sign semantics.
- **Regression coverage:** Validator rejects negative components and accepts negative net values; CI #458 / run `34152257188` 10/10.
- **CI / review evidence:** PR #152 External comment `5574239091` REQUEST CHANGES; remediation head `2f8882cc94becbec838b98f63640e2653f593eae`; approval comment `5574580529`.
- **Residual limitations:** Historical PR #152 remained noncanonical; replacement PR #153 was independently reviewed.

### HF-375-02 — Gross dividend/coupon summary semantics were activated without explicit API description

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** P1
- **Problem:** Stage 3.75 populated `PortfolioSummary.dividendsReceived` / `couponsReceived` from gross effective-ledger amounts, but the OpenAPI descriptions did not state gross/cutoff/non-net/non-performance meaning.
- **Root cause:** Implementation/UI semantics advanced without equally precise public API documentation.
- **Failure / attack scenario:** A client could interpret the fields as net receipts or investment-return metrics.
- **Project impact:** Financial meaning at the API boundary would be ambiguous.
- **Original / insufficient design:** Keep legacy field names/descriptions without explicit gross semantics.
- **Why review rejected or challenged it:** External review classified the ambiguity P1 because the stage materially activated financial fields.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Document gross effective-ledger amounts under summary cutoff/as-of; deductions are separate; fields are not net receipts/performance.
- **Why the remediation was selected:** Clarify semantics without breaking field names or formulas.
- **Regression coverage:** OpenAPI contract checks; replacement PR #153 revalidated the definitions; CI #459/#460 green.
- **CI / review evidence:** PR #152 comment `5574239091`; replacement PR #153 External comment `5574951592` APPROVED.
- **Residual limitations:** Tax/performance interpretation remains out of scope.

### HF-375-03 — Original PR exceeded changed-file budget without completed exception/disclosure gate

- **Classification:** EXTERNAL GOVERNANCE FINDING
- **Severity:** Ready blocker
- **Problem:** PR #152 had 31 changed files versus default `<=25` and initially lacked the required exception explanation/Principal Architect approval and full disclosure fields.
- **Root cause:** The broad vertical-slice candidate crossed the workflow threshold before governance metadata/approval caught up.
- **Failure / attack scenario:** Review could proceed on a subject outside default budget without explicit acceptance of the review-size trade-off.
- **Project impact:** Violates scope-control governance and weakens auditability.
- **Original / insufficient design:** Publish 31-file candidate without completed exception gate.
- **Why review rejected or challenged it:** External review blocked Ready until exception/disclosure requirements were satisfied.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Add complete PR disclosures and obtain Principal Architect approval for exactly 31 files in comment `5574540750`.
- **Why the remediation was selected:** Make the trade-off explicit without weakening the global default.
- **Regression coverage:** Fresh exact-head review after approval; no technical finding suppressed.
- **CI / review evidence:** PR #152 comments `5574239091` and `5574540750`.
- **Residual limitations:** The original PR still did not become canonical; replacement #153 returned to 24 files and needed no exception.

### HF-375-04 — Original prepublication Internal review evidence could not be recovered

- **Classification:** HISTORICAL EVIDENCE GAP
- **Severity:** Governance gap; not a reconstructed defect
- **Problem:** Historical PR #152 prepublication Internal review record was unavailable when the canonical replacement was prepared.
- **Root cause:** The required evidence artifact was not recoverable from contemporary repository/PR sources.
- **Failure / attack scenario:** Inferring the missing review from later code, External comments or memory would create false chronology.
- **Project impact:** An auditor cannot prove the original candidate completed the mandatory Internal phase as required.
- **Original / insufficient design:** Treat later/historical review outcomes as if they reconstruct the missing Internal review.
- **Why review rejected or challenged it:** Replacement dossier explicitly forbids retroactive compliance claims and does not inherit PR #152 verdicts.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Preserve the gap; restart development path from same canonical base with replacement PR #153, fresh Internal review, exact-head CI and fresh External review.
- **Why the remediation was selected:** Fact over narrative completeness; independently reviewed replacement is stronger than invented evidence.
- **Regression coverage:** Replacement 24-file tree `a25d940...`; head `f97f440...`; CI #459 / run `34155729975` 10/10; External `5574951592`; evidence CI #460 green.
- **CI / review evidence:** Stage 3.75 canonical dossier and PR #153 evidence.
- **Residual limitations:** The missing original Internal review remains permanently missing and must stay labeled as such.


## Stage 3.76 — Manual Market Valuation & Portfolio P/L

### HF-376-01 — PR description identified superseded candidate SHA/tree/manifest

- **Classification:** EXTERNAL GOVERNANCE FINDING
- **Severity:** Blocking governance finding
- **Problem:** PR #155 description still identified a superseded pre-correction commit/tree/manifest while actual published head/tree were different.
- **Root cause:** Candidate identity metadata was not synchronized after prepublication correction.
- **Failure / attack scenario:** Reviewer/auditor could bind conclusions to bytes that were not the actual PR head.
- **Project impact:** Evidence-chain integrity becomes ambiguous even though runtime code is sound.
- **Original / insufficient design:** Leave old candidate identity in PR metadata.
- **Why review rejected or challenged it:** Fresh External review found no blocking runtime defect but returned REQUEST CHANGES solely for identity drift.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Correct PR description metadata only to `14a824b...` / tree `781f14e...`; do not reuse superseded manifest; leave repository head unchanged.
- **Why the remediation was selected:** Metadata should identify actual reviewed bytes.
- **Regression coverage:** Fresh re-review confirmed no code/head drift and 25/25 paths; CI #465 / run `34169041943` 10/10.
- **CI / review evidence:** PR #155 comments `5576561274` REQUEST CHANGES and `5576571231` APPROVED.
- **Residual limitations:** Governance/evidence identity drift only, not a manual-valuation runtime defect.

### HF-376-02 — Evidence publication changed dossier hash while migration authority_refs still pinned old hash

- **Classification:** CI / AUTHORITY-BINDING GOVERNANCE FINDING
- **Severity:** CI blocker
- **Problem:** After Internal evidence was published into the Stage 3.76 dossier, migration policy `authority_refs` still referenced the pre-evidence dossier SHA-256.
- **Root cause:** The migration governance model hash-binds authority documentation, and evidence publication changed those document bytes.
- **Failure / attack scenario:** CI #466 rejects the evidence head even though runtime/SQL semantics are unchanged.
- **Project impact:** Protected merge remains blocked; stale refs make migration provenance unverifiable.
- **Original / insufficient design:** Publish evidence without synchronizing bound authority hash.
- **Why review rejected or challenged it:** The migration validator correctly treated hash mismatch as failure; project did not weaken it.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Update only Stage 3.76 authority hash references in `policy_manifest.json`; keep runtime, SQL migration files, frontend, OpenAPI and tests unchanged.
- **Why the remediation was selected:** Preserve machine-enforced authority binding while allowing evidence bytes to become current authority.
- **Regression coverage:** No-drift compare from `14a824b...` to final `f89e1e...`; CI #467 / run `34174642209` 10/10 including migration validation.
- **CI / review evidence:** PR #155 chronology; final head `f89e1e793ef505c31bf76de7d43c4d40fdd9c38b`; merge `e00699f8d455bcbaea0c1dc69ce534460fea6ff9`.
- **Residual limitations:** Future changes to hash-bound authority docs must synchronize refs instead of weakening validator.


## Stage 3.77 — Portfolio Money-Weighted Return / XIRR

### HF-377-01 — Fixed-grid root discovery could miss close XIRR roots and derivative recursion did not shrink

- **Classification:** PREPUBLICATION NUMERICAL HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** Early solver designs could miss closely spaced roots with a fixed sampling grid, and derivative recursion did not initially guarantee a smaller effective term set.
- **Root cause:** Generalized exponential roots cannot be proven complete by a fixed grid; naive derivative recursion can repeat equivalent complexity.
- **Failure / attack scenario:** A multi-sign-change series could be labeled unique/no-root despite multiple close roots, or recursion could become inefficient/non-terminating.
- **Project impact:** Wrong money-weighted-return classification or excessive CPU.
- **Original / insufficient design:** Grid-sample the domain and recursively analyze a derivative without factoring minimum exponent.
- **Why review rejected or challenged it:** The Stage 3.77 dossier records these as solver defects found/remediated before publication.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Solve in `x=log(1+r)`; derivative-recursive monotonic interval isolation; factor minimum exponent so effective term count shrinks; one-sign-change direct-bisection fast path.
- **Why the remediation was selected:** Use mathematical structure rather than arbitrary sampling.
- **Regression coverage:** Close-root/multiple-root vectors; 20,000 NumPy polynomial-equivalent cases with 0 root-count label mismatches; Go/race CI.
- **CI / review evidence:** Stage 3.77 dossier Solver hardening evidence; PR #157.
- **Residual limitations:** Strong engineering cross-check, not a formal proof for every generalized exponential series.

### HF-377-02 — Bisection could terminate on loose NPV tolerance instead of root/bracket precision

- **Classification:** PREPUBLICATION NUMERICAL HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** Early refinement could stop because evaluated NPV was small even when the rate interval was not precise enough for scale-8 output.
- **Root cause:** Function-value tolerance is coefficient/scale dependent and does not directly bound rate error near flat regions.
- **Failure / attack scenario:** A simple root could round incorrectly; tangent/critical behavior could be mistaken for a confident solution.
- **Project impact:** Incorrect XIRR or false uniqueness.
- **Original / insufficient design:** Use NPV residual threshold as primary stopping condition.
- **Why review rejected or challenged it:** The dossier lists premature NPV-tolerance termination as a remediated defect.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Refine simple roots by bracket width; tangent/critical candidates fail closed as `AMBIGUOUS_MULTIPLE_ROOTS` unless confidently unique.
- **Why the remediation was selected:** Bracket width directly bounds rate interval; ambiguity is represented instead of guessed.
- **Regression coverage:** Close-root/tangent/double-root tests; high-precision oracle; exact-head CI.
- **CI / review evidence:** Stage 3.77 dossier XIRR mathematics / solver hardening.
- **Residual limitations:** Numerically unresolved cases intentionally return unavailable.

### HF-377-03 — Binary float could choose the wrong scale-8 Half-Even side near a midpoint

- **Classification:** PREPUBLICATION NUMERICAL HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** Direct float64 quantization could drift at Half-Even midpoints, including scale-8 integer/ULP limits and near-midpoint irregular-date roots.
- **Root cause:** Binary representation cannot exactly encode many decimal midpoints; ULP spacing grows with magnitude.
- **Failure / attack scenario:** A mathematically equal/near-equal XIRR can round to the wrong eighth decimal.
- **Project impact:** Public financial result violates canonical Decimal/Half-Even contract.
- **Original / insufficient design:** Round the float64 solution directly to eight decimals.
- **Why review rejected or challenged it:** Dossier records midpoint drift, ULP boundary and near-midpoint discrimination defects.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Guard float exact-integer/ULP limits; evaluate integer-year midpoint cases exactly with `big.Rat`; fail closed for unresolved irregular-date midpoint decisions.
- **Why the remediation was selected:** Use exact arithmetic where structure permits; otherwise preserve uncertainty.
- **Regression coverage:** Half-Even midpoint vectors; 1,000 irregular ACT/365 cases against 90-digit `mpmath` with 0 scale-8 mismatches.
- **CI / review evidence:** Stage 3.77 dossier Solver hardening evidence.
- **Residual limitations:** float64 remains internal; canonical/public values are Decimal strings.

### HF-377-04 — Exponential evaluation lost coefficient magnitude or overflowed/underflowed on extreme series

- **Classification:** PREPUBLICATION NUMERICAL HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** Extreme coefficient ratios/magnitudes, same-date aggregation and long calendar spans exposed range/precision hazards.
- **Root cause:** Naive exponentiation/scaling can overflow/underflow; exponent-only scaling can erase coefficient dominance; duration assumptions can saturate; aggregated flows can exceed NUMERIC(28,8).
- **Failure / attack scenario:** Solver sign/evaluation becomes wrong/non-finite or canonical input overflows silently.
- **Project impact:** Incorrect root count/XIRR or false valid output.
- **Original / insufficient design:** Evaluate raw exponentials/coefficient sums and rely on ordinary float/time range.
- **Why review rejected or challenged it:** Dossier lists ratio overflow/underflow, coefficient-magnitude loss, same-date overflow and long-span saturation as remediated.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use coefficient-aware scaled exponential evaluation with finite dominance bounds; guard canonical aggregation overflow; compute date exponents without saturating duration assumptions.
- **Why the remediation was selected:** Preserve sign/root structure across large dynamic ranges and fail closed at storage limits.
- **Regression coverage:** Extreme-ratio/magnitude, same-date overflow and long-span vectors; NumPy/mpmath oracles; CI.
- **CI / review evidence:** Stage 3.77 dossier.
- **Residual limitations:** Inputs outside validated bounds may return `NUMERICAL_SOLUTION_FAILED`.

### HF-377-05 — Adversarial multi-sign-change series could monopolize solver work

- **Classification:** PREPUBLICATION PERFORMANCE HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** Recursive root isolation for many sign changes can become expensive as term count grows.
- **Root cause:** Multi-root completeness work scales much worse than ordinary one-sign-change cases.
- **Failure / attack scenario:** A crafted large alternating cash-flow series consumes excessive CPU.
- **Project impact:** Availability/resource exhaustion at a public financial endpoint.
- **Original / insufficient design:** Allow recursive multi-sign-change path to scale with arbitrary term count.
- **Why review rejected or challenged it:** Dossier records adversarial complexity behavior as a remediated defect.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Cap only multi-sign-change recursive path at 160 terms; keep ordinary one-sign-change fast path outside that cap.
- **Why the remediation was selected:** Bound expensive ambiguity work without penalizing normal contribution histories.
- **Regression coverage:** Adversarial complexity regression and 1,500-date one-sign-change performance witness; CI green.
- **CI / review evidence:** Stage 3.77 dossier XIRR mathematics / solver hardening.
- **Residual limitations:** The cap is a safety boundary, not a proof that every <=160 series is cheap.

### HF-377-06 — Duplicate `asOfDate` query parameters were silently collapsed

- **Classification:** EXTERNAL REVIEW FINDING
- **Severity:** Blocking finding; severity not separately recorded
- **Problem:** Shared query accessor used first/peek semantics, so repeated `asOfDate` values were collapsed rather than rejected.
- **Root cause:** Stage 3.77 requires exactly one explicit BusinessDate but generic query helper did not express multiplicity.
- **Failure / attack scenario:** `?asOfDate=2026-01-01&asOfDate=2026-02-01` could use one value and ignore the conflict.
- **Project impact:** Ambiguous financial cutoff/request identity.
- **Original / insufficient design:** Read `asOfDate` through shared single-value `QueryArgs.Peek`.
- **Why review rejected or challenged it:** Fresh External review returned REQUEST CHANGES for the ambiguity.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** At Stage 3.77 endpoint count occurrences and require exactly one; reject missing/empty/whitespace/invalid/duplicate before store work.
- **Why the remediation was selected:** Localizes strictness to the financial endpoint.
- **Regression coverage:** Conflicting duplicate regression proves HTTP 400 and zero store calls; full Go/race CI.
- **CI / review evidence:** PR #157: published remediation head `88b7e085...` received REQUEST CHANGES; final remediation head `3a976568567c478a20ed51e85fb35da020132328`; CI run `34303636331` 10/10; final External verdict APPROVED. GitHub review/comment ID: NOT RECORDED IN CONTEMPORANEOUS GITHUB EVIDENCE.
- **Residual limitations:** Unrelated endpoints keep existing query behavior.

### HF-377-07 — Published evidence state and first CI remediation did not match the actual final PR subject

- **Classification:** CI / DOCUMENTATION INTEGRITY FAMILY
- **Severity:** Blocking evidence finding plus CI blockers
- **Problem:** First CI exposed dependency advisories and a non-canonical zero-Money fixture; later External review found the dossier still described the original 22-file state although PR had 25 files.
- **Root cause:** Implementation evolved through dependency/test remediation, while evidence inventory lagged actual published bytes.
- **Failure / attack scenario:** CI/security fails and reviewer can rely on stale changed-file/subject description.
- **Project impact:** Protected-merge evidence is not cleanly bound to the reviewed 25-file head.
- **Original / insufficient design:** Keep original dependency/test fixture and 22-file dossier after remediation expanded the PR.
- **Why review rejected or challenged it:** CI required dependency/test fixes; External review required dossier to describe actual 25-file subject.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Update Next.js `16.3.3→16.3.4`, resolve Sharp `0.35.4`, pin `baseline-browser-mapping 2.11.21`, regenerate lockfile, fix canonical zero-Money fixture; update already-counted dossier path to actual 25-file state.
- **Why the remediation was selected:** Keep security gates green and evidence identity synchronized without a 26th path.
- **Regression coverage:** Protected CI 10/10 on `88b7e085...`; final `3a976568...` run `34303636331` 10/10; evidence head `b3bc17e...` run `34303895226` 10/10; evidence-only no drift.
- **CI / review evidence:** Stage 3.77 dossier and PR #157 body; merge `63d916b447e91c4de54efee5c66b27cf7727be92`.
- **Residual limitations:** Final External GitHub review/comment ID is not recorded in contemporary GitHub discussion; verdict remains preserved by canonical dossier/PR/registries.


## 9. Historical evidence gaps and irreversible governance deviations

### 9.1 Stage 3.75 original Internal Review evidence

The most important evidence gap is the original Stage 3.75 PR #152 prepublication Internal review record.

Known facts:

```text
Historical PR #152:
  external findings/review chronology = AVAILABLE
  exact-head CI chronology             = AVAILABLE
  Principal Architect file exception   = AVAILABLE
  original prepublication Internal     = UNAVAILABLE
```

The canonical replacement PR #153 did **not** treat PR #152 review outcomes as approval. It restarted the mandatory development path with a fresh 24-file candidate, a fresh Internal review, exact-head CI #459, fresh External review, evidence publication and exact-head CI #460.

No later artifact may claim that the missing PR #152 Internal review was reconstructed.

### 9.2 Stage 3.71 evidence-withholding deviation

The earlier adversarial/Internal evidence was visible before the formal External verdict. The later disposition preserves noncompliance and bounded residual governance risk. Fresh External review explicitly did not use the earlier Internal/adversarial result as supporting evidence.

### 9.3 Stage 3.72 Internal-verdict withholding deviation

The Draft PR body exposed the current Internal verdict before the External verdict. Editing the PR later, repeating review or creating a disposition cannot make the original temporal interval compliant. The dedicated disposition preserves this fact.

## 10. Cross-stage recurring root-cause families

The historical sequence shows recurring engineering families rather than isolated one-off defects:

### 10.1 Financial authority drift

Examples:
- Stage 3.70 rounded-basis/WAC inverse repricing contradiction;
- Stage 3.71 wall-clock snapshot selection;
- Stage 3.72 undefined zero-denominator allocation;
- Stage 3.75 API sign/gross-income semantics;
- Stage 3.77 scale-8 Half-Even numerical boundary.

Recurring lesson: define one authoritative financial state, make derived values one-way, and never let transport/presentation/time metadata become financial authority.

### 10.2 Provider and external-source trust boundaries

Examples:
- Stage 3.59 CookieJar, redirect, provider-data validation and typed-nil failure;
- Stage 3.64 request fan-out, source identity minimization and no-store caching;
- Stage 3.60 explicit MOEX activation NO-GO;
- Feature 3D constrained source/use rights and disabled-by-default runtime.

Recurring lesson: technical adapter availability is not source/use approval or runtime activation.

### 10.3 Temporal and asynchronous state integrity

Examples:
- Stage 3.64 stale browser results;
- Stage 3.66 unmount cancellation;
- Stage 3.71 cutover quiescence;
- Stage 3.72 lifecycle metadata chronology;
- Stage 3.77 exact explicit BusinessDate and duplicate-query rejection.

Recurring lesson: wall-clock/order/lifecycle assumptions must be explicit and tied to the authoritative state transition, not incidental timing.

### 10.4 Idempotency and replay ordering

Examples:
- Stage 3.68 exact replay before rate limiting and browser retry-key continuity;
- Stage 3.74 correction/reversal retry journal;
- Stage 3.75 effective-ledger reuse;
- Stage 3.77 external-flow construction from correction/reversal-aware ledger truth.

Recurring lesson: resolve prior durable intent before admitting a new mutation and preserve the same retry identity through ambiguity.

### 10.5 PostgreSQL rollout and authority binding

Examples:
- Stage 3.71 `name[]`/`text[]` readiness defect;
- Stage 3.71 single migration-authority model and continuous cutover quiescence;
- Stage 3.76 hash-bound migration authority references.

Recurring lesson: migration/readiness evidence must execute against real PostgreSQL and governance authority must remain machine-verifiable.

### 10.6 Evidence identity and current-state metadata

Examples:
- Stage 3.72 PREPUBLICATION label after Draft publication;
- Stage 3.72 stale Stage 3.71 closure text;
- Stage 3.75 missing original Internal record;
- Stage 3.76 superseded PR identity;
- Stage 3.77 stale 22-file dossier.

Recurring lesson: exact SHA/tree/file-set identity is part of review evidence, not decorative metadata.

## 11. Feature 3D as the reference standard

Feature 3D remains the strongest existing example of the desired forensic shape because its post-merge closure explicitly records, for each material Internal finding:

```text
problem
→ root cause
→ failure / attack scenario
→ project impact
→ original design
→ why review rejected it
→ second-order scenario when actually considered
→ final remediation
→ rationale
→ regression coverage
→ CI/review evidence
→ residual limitation
```

This reconciliation does not duplicate or rewrite Feature 3D. It uses that shape only as the normalization template for older feature history.

## 12. Future forensic documentation standard — proposed, not workflow-authoritative

This reconciliation does **not** modify `docs/REVIEW_WORKFLOW.md`.

A later, separately reviewed governance amendment should consider requiring the following for future substantive stages with material findings:

1. preserve each material finding in the canonical stage/closure evidence;
2. distinguish review finding, preventive hardening, CI/test failure and metadata/governance finding;
3. record second-order scenarios only when they were genuinely considered;
4. use the literal `NOT RECORDED IN CONTEMPORANEOUS EVIDENCE` for historical gaps rather than inference;
5. bind review verdicts to exact commit/tree/file-set and exact-head CI;
6. never call a stage `CLOSED/CANONICAL` in a candidate document before protected merge if closure is merge-activated;
7. preserve irreversible process deviations rather than claiming retroactive compliance.

This proposal is explanatory only. It creates no new mandatory control until separately accepted.

## 13. Reconciliation verdict

```text
POST_MERGE_CANONICAL_STATE

HISTORICAL_SCOPE = STAGES_3_57_THROUGH_3_77
FEATURE_3D_REFERENCE = YES

MATERIAL_RECORDS = 47
FULL_FORENSIC_RECORDS = 33
PARTIAL_FORENSIC_RECORDS = 14
NOT_RECORDED_FIELDS = 60

ORIGINAL_REVIEWED_BASE_COMMIT =
5d8ed09ff1061a57281b231cd4f448df174a625a

ORIGINAL_REVIEWED_BASE_TREE =
78579ebb3d39501a28bb87a8d843791a5983b325

REVIEWED_HEAD =
fa2f0a164c7e2c4fe414a7bb9469eedf5ea7df54

REVIEWED_TREE =
e742d1084fdcc5751acf896738ba9065a4b0f741

MERGED_DEVELOP =
5da679f0d15f8742239659ff9cebc451a213894d

MERGED_TREE =
e742d1084fdcc5751acf896738ba9065a4b0f741

CONTENT_DRIFT = NONE
CI = 10/10 SUCCESS

HISTORICAL_DOSSIERS_REWRITTEN = NO

RUNTIME_DRIFT = NONE
OPENAPI_DRIFT = NONE
DATABASE_OR_MIGRATION_DRIFT = NONE
DEPENDENCY_DRIFT = NONE
PROVIDER_ACTIVATION_DRIFT = NONE

STAGE_3_78_AUTHORIZATION = NONE
STAGE_3_78_STARTED = NO

DOCUMENT_STATUS = APPROVED / CANONICAL
FORENSIC_RECONCILIATION = CLOSED
```
