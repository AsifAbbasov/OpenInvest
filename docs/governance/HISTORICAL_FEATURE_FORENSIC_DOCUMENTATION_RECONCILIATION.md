# Historical Feature Forensic Documentation Reconciliation — Stages 3.57–3.77

Canonical record: commit(s) `5d8ed09ff1061a57281b231cd4f448df174a625a`, `78579ebb3d39501a28bb87a8d843791a5983b325`, `fa2f0a164c7e2c4fe414a7bb9469eedf5ea7df54`, `e742d1084fdcc5751acf896738ba9065a4b0f741`, `5da679f0d15f8742239659ff9cebc451a213894d`.

## 1. Purpose


- what problem or defect was actually found;
- what design assumption caused it;
- what failure or attack scenario was demonstrated or explicitly considered;
- what project impact mattered;
- what insufficient design was replaced;
- what remediation was selected;
- what evidence is missing and therefore must remain missing.


## 2. Non-retroactive evidence rule

The governing rule for this document is:

> contemporary evidence outranks later narrative completeness.

Accordingly:

1. a finding is recorded only when the contemporary repository/PR/CI evidence demonstrates it;
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
- **MINIMAL / NO MATERIAL REVIEW FINDINGS RECORDED** — no material remediation finding is preserved; this document does not manufacture one.
- **HISTORICAL EVIDENCE GAP** — a mandatory historical evidence artifact is known to be unavailable and must remain explicitly unavailable.

A **preventive design decision** or **prepublication hardening** is not relabeled as an External/Internal finding unless the contemporary source says it was one.

## 4. Evidence corpus and precedence


Evidence precedence for this document is:

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

Canonical record: PR #123, PR #126, PR #127, PR #128, PR #130, PR #133, PR #134, PR #137, PR #138, PR #148, PR #150, PR #152, PR #153, PR #155, PR #157.

## 7. Stages with no material remediation finding recorded

The following stages are deliberately **not** assigned fictional defect histories:

- **Stage 3.57 — Market Data Provider Boundary.** Contemporary evidence records preventive provider-neutral design and validation boundaries. No material remediation finding is preserved.
- **Stage 3.58 — Market Data Provider Boundary Closure.** Closure/evidence synchronization only; no new material runtime finding.
- **Stage 3.61 — Corporate Actions Calendar Planning.** Design/source-governance stage; no material implementation remediation finding preserved.
- **Stage 3.65 — Corporate Actions API/UI Closure.** Closure-only stage; no new material defect.
- **Stage 3.67 — Request Cancellation Closure.** Closure-only stage; Stage 3.66 contains the substantive remediation history.
- **Stage 3.69 — Dividend Calculator Closure.** Closure-only stage; Stage 3.68 contains the substantive remediation history.
- **Stage 3.73 — Portfolio Time Machine.** Contemporary evidence records a deliberately minimal reuse design over the Stage 3.72 `asOfDate` projection, including prevention of a second financial engine/API/schema/provider. No material remediation finding is preserved.


## 8. Material forensic records

## Stage 3.59 — MOEX ISS Quote Provider Adapter

### HF-359-01 — Prepublication provider-boundary deterministic hardening

Canonical record: PR #123.

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
- **Residual limitations:** No production MOEX runtime/public activation was authorized; the adapter remained dormant.

### HF-359-03 — HTTP redirect could escape the fixed-host provider boundary

Canonical record: PR #123; commit(s) `fe265e9127b44b2fe5899b62b1fc47c8429075e7`, `4605753d68b56227c646663446700271042cc299`.

### HF-359-04 — Typed-nil `QuoteProvider` could bypass fail-closed unconfigured state

- **Severity:** P2
- **Problem:** An interface containing a typed-nil `QuoteProvider` was non-nil at the interface level and could bypass a simple nil guard.
- **Root cause:** Go interface nil semantics differ from a nil underlying pointer.
- **Failure / attack scenario:** A typed-nil provider whose method dereferences its receiver could be invoked and panic.
- **Project impact:** A normal provider-unavailable condition could become a process-level panic.
- **Original / insufficient design:** Check only `s.quoteProvider == nil`.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Normalize nil and typed-nil providers in `NewServiceWithQuoteProvider` to the existing unconfigured service state.
- **Why the remediation was selected:** Dependency validity belongs at construction; callers then get deterministic unavailable behavior.
- **Regression coverage:** Typed-nil provider regression would panic if invoked but instead returns `ErrMarketQuoteProviderUnavailable`; full CI green.
Canonical record: PR #123.
- **Residual limitations:** Future interface-typed dependencies need the same typed-nil awareness if nil is invalid.


## Stage 3.62 — Corporate Action Boundary / Feature 3A

### HF-362-01 — Opaque SourceEventID validation missed Unicode C1 controls

- **Severity:** P3
- **Problem:** `validateOpaqueCorporateActionID` claimed to reject control characters but rejected only C0 controls and DEL.
- **Root cause:** The implementation encoded an ASCII-oriented control check instead of the Unicode semantic rule stated by the contract.
- **Failure / attack scenario:** Provider evidence containing U+0080..U+009F could pass validation.
- **Project impact:** Malformed provenance identifiers could enter the canonical boundary and later surface in logs/evidence handling.
- **Original / insufficient design:** Reject runes `<0x20` and `0x7f` only.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use `unicode.IsControl` and add a C1 regression.
- **Why the remediation was selected:** Use the language's Unicode semantic category rather than maintain partial ranges.
Canonical record: PR #126; commit(s) `97089869e8f3c5ccaf14a9aecd4927d8e2c2eb85`.
- **Residual limitations:** `SourceEventID` remains opaque/internal; no provider activation or persistence was authorized.

#### Non-material metadata note


## Stage 3.63 — Calendar + Heatmap Projection / Feature 3B

### HF-363-01 — Supersession could remove an unrelated instrument or action kind

- **Severity:** P2
- **Problem:** The first projection candidate did not require a present predecessor to share `InstrumentID` and `Kind` with its superseding event.
- **Root cause:** Graph-shape validation did not include the economic identity dimensions of a supersession edge.
- **Failure / attack scenario:** A dividend for one ticker could supersede a coupon or an event for another ticker and remove that unrelated dated event.
- **Project impact:** Calendar/heatmap truth could silently lose valid events through a malformed cross-entity link.
- **Original / insufficient design:** Allow any well-formed `SupersedesEventID` edge when the predecessor is present.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** When the predecessor is present, require same `InstrumentID` and same `Kind`; otherwise fail closed with `ErrInvalidCorporateActionProjection`.
- **Why the remediation was selected:** Preserve partial-batch support while preventing cross-entity semantic corruption.
- **Regression coverage:** Cross-instrument/kind, fork and cycle regressions; focused statement coverage 100%; CI #339 / run `33914229650` 10/10.
Canonical record: PR #127.
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
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Reject more than 50 instruments before provider invocation.
- **Why the remediation was selected:** Fail-fast admission preserves explicit request scope and makes provider demand auditable.
Canonical record: PR #128.
- **Residual limitations:** The 50-instrument bound is an application contract, not a provider-specific production rate policy.

### HF-364-02 — Frontend stale responses could overwrite a newer Corporate Actions query

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** The first UI candidate did not yet have the final request-generation stale-result guard.
- **Root cause:** Asynchronous completion order is independent of user query/update order.
- **Failure / attack scenario:** An older request could finish after a newer one and replace the latest calendar/heatmap or source-unavailable state.
- **Project impact:** Users could see data for the wrong instruments/date window.
- **Original / insufficient design:** Rely on request completion without a final generation identity check.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use AbortController plus monotonically invalidated request generation; commit state only for the current generation.
- **Why the remediation was selected:** Independent transport-cancel and stale-completion guards cover different races.
- **Regression coverage:** Component tests cover replacement/newer-result wins and stale completion suppression.
- **Residual limitations:** Component-unmount cancellation was intentionally retained as a non-blocking P3 and closed in Stage 3.66.

### HF-364-03 — Provider-owned source identity and caching rights were too exposed/implicit


### HF-364-04 — Provider over-return could widen the public date scope


### HF-364-05 — Frontend component test environment did not match Next/React browser assumptions

- **Classification:** CI / TEST-ENVIRONMENT REMEDIATION FAMILY
- **Severity:** CI blockers; product severity not recorded
- **Root cause:** The Node/tsx test runner lacked the CSS/browser environment that the styled Next.js component and `Link` expected.
- **Failure / attack scenario:** CI component tests fail before exercising feature behavior.
- **Project impact:** Required protected checks could not prove the UI regression surface.
- **Original / insufficient design:** Use the existing Node component test path without a CSS loader and with static ReactDOM imports before JSDOM globals.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Add a test-only CSS-module loader; install JSDOM/browser globals before dynamic React/ReactDOM/component imports; provide `self: dom.window`; run tests sequentially.
- **Why the remediation was selected:** Confine remediation to test infrastructure with no production runtime/dependency expansion.
- **Regression coverage:** CI #341 exposed CSS failure; commits `cee7e1...`, `78e7aaa...`; CI #343 exposed `self`; final head `9bbcf6d...`; CI #344 / run `33927609434` 10/10.
Canonical record: PR #128.
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
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Unmount cleanup increments request generation first, aborts the controller, clears it, and acceptance checks both `signal.aborted` and generation identity.
- **Why the remediation was selected:** Smallest owner-local fix; preserves shared client semantics and provides redundant race protection.
- **Regression coverage:** Component tests prove active signal aborted on unmount, aborted requests do not surface false errors, replacement wins, genuine errors remain visible; CI #347 / run `33952598235` 10/10.
Canonical record: PR #130.
- **Residual limitations:** Shared client still maps low-level fetch exceptions generically; broader typed cancellation is separate scope.

### HF-366-02 — Unmount regression test could leak React root after early assertion failure

- **Classification:** INTERNAL TEST-QUALITY FINDING
- **Severity:** P3
- **Problem:** The initial unmount test did not guarantee React root cleanup if an early assertion failed.
- **Root cause:** Teardown was not fail-safe around intermediate assertions.
- **Failure / attack scenario:** One failed test could contaminate later sequential component tests or leave lifecycle state/resources alive.
- **Project impact:** Evidence quality and reproducibility would be weaker even though production runtime was not defective.
- **Original / insufficient design:** Cleanup only on the ordinary success path.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Add fail-safe teardown and retain sequential isolation; add a genuine network-failure regression so abort suppression cannot hide normal transport failures.
- **Why the remediation was selected:** Cancellation tests must themselves be deterministic and distinguish abort from genuine failure.
- **Regression coverage:** Focused preflight plus exact-head frontend tests under CI #347.
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
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use only canonical `openinvest.ts`; calculator component remains presentation/orchestration only.
- **Why the remediation was selected:** Keep one HTTP client authority and reuse established semantics.
- **Residual limitations:** Feature helpers may compose typed calls, but transport authority remains shared.

### HF-368-02 — Anonymous calculator replay identity/privacy model was too weakly specified

- **Classification:** PREPUBLICATION HARDENING FAMILY
- **Severity:** Severity not separately recorded
- **Problem:** The original anonymous replay subject design and uniqueness/retention wording did not yet provide the final domain-separated technical principal/privacy semantics.
- **Root cause:** A public idempotent endpoint needs replay isolation without creating a user profile or overclaiming mathematical uniqueness.
- **Failure / attack scenario:** A shared/poorly scoped subject could collide conceptually with other principals; docs could claim injective uniqueness or zero persistence despite exact-response retention.
- **Project impact:** Audit/privacy semantics would be misleading and could set unsafe precedent.
- **Original / insufficient design:** Use the original anonymous subject design plus stronger uniqueness/zero-persistence language.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Derive deterministic UUIDv8 scope from SHA-256(domain separator + validated key); keep original key in replay uniqueness; describe collision resistance, not absolute uniqueness; disclose existing 24-hour exact-response persistence.
- **Why the remediation was selected:** Technical replay identity without user/cookie/IP fingerprinting, with accurate cryptographic/retention claims.
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
- **Residual limitations:** Limiter is in-process, not a distributed/edge DoS shield.

### HF-368-04 — Backend trim-first ticker admission contradicted the OpenAPI grammar


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
- **Residual limitations:** Key identifies technical retry intent, not a user account.

### HF-368-06 — Money presentation used scale-8 trim-only output instead of exact two-decimal Half-Even UI semantics

- **Classification:** PREPUBLICATION HARDENING
- **Severity:** Severity not separately recorded
- **Problem:** Initial UI displayed returned money by trimming scale-8 strings rather than applying reviewed two-decimal money presentation.
- **Root cause:** Canonical calculation scale and human display precision were conflated.
- **Failure / attack scenario:** Values requiring third-decimal rounding could display inconsistently or with excessive precision.
- **Project impact:** User-visible financial presentation could disagree with project monetary formatting.
- **Original / insufficient design:** Trim canonical scale-8 Decimal for display.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Implement exact decimal-string two-decimal Half-Even money display; percentage text uses decimal-string shifting.
- **Why the remediation was selected:** Preserve exact backend truth and deterministic presentation without client financial recomputation.
- **Residual limitations:** Presentation rounding does not change backend scale-8 results or replay body.


## Stage 3.70 — Portfolio Position & Cost Basis Planning / ADR-009

### HF-370-01 — WAC scale-8 state model was self-contradictory

- **Severity:** P1
- **Problem:** The first planning candidate treated `quantity + acquisitionBasis` as canonical state, re-derived WAC from rounded basis/quantity, and also required partial SELL to preserve WAC.
- **Root cause:** Scale-8 rounding makes rounded acquisition basis non-invertible for fractional quantities.
- **Failure / attack scenario:** For `quantity=0.20000000`, WAC `275.12345678` gives basis `55.02469136`; dividing basis by quantity yields `275.12345680`, silently repricing WAC.
- **Project impact:** Financial state would drift after partial SELL and violate the stated WAC methodology.
- **Original / insufficient design:** Canonical quantity+basis state with inverse WAC reconstruction.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Make `quantity + weightedAverageCost` authoritative; derive `acquisitionBasis = Round8HalfEven(quantity × WAC)`; never inverse-reprice after SELL; add the fractional non-invertibility vector.
- **Why the remediation was selected:** Matches the rule that partial SELL preserves WAC and makes rounding one-way/deterministic.
Canonical record: PR #137; commit(s) `a5c7a507a586271b7117580585c88191693144e3`, `2af826392e081eb07d3234be1c94833bae768aa7`.
- **Residual limitations:** Acquisition basis remains trade-price basis, not tax basis; non-invertibility is intentional.

### HF-370-02 — Planned transaction conflict contract overstated existing 409/429 API semantics

- **Severity:** P2
- **Problem:** The first plan described a generic business-conflict surface/429 behavior that the frozen POST transaction contract did not actually expose.
- **Root cause:** Planning language assumed stronger existing API semantics instead of reconciling with canonical OpenAPI.
- **Failure / attack scenario:** Stage 3.71 could ship insufficient-position conflict without an approved schema or falsely claim 429 support.
- **Project impact:** Contract-first governance would be violated and clients could receive undocumented behavior.
- **Original / insufficient design:** Treat existing POST 409 as generic enough and describe 429 as existing.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Plan one narrow POST-transaction 409 contract change: `TransactionConflict` preserves `IDEMPOTENCY_CONFLICT` and adds `INSUFFICIENT_POSITION_QUANTITY`; remove false 429 claim.
- **Why the remediation was selected:** Adds only the conflict needed for SELL while preserving HTTP 409/envelope.
Canonical record: PR #137.
- **Residual limitations:** The separate P3 stale SOT version in `DOCUMENT_INDEX.md` is retained as a non-material registry note, not a separate material record.

#### Non-material registry note

The first External Stage 3.70 review also found a P3 registry-version mismatch: `DOCUMENT_INDEX.md` identified SOT-001 as `1.4.63` while the candidate Source of Truth was `1.4.90`. The remediation synchronized the registry. It is preserved here as metadata drift, not counted among the two material financial/API findings.


## Stage 3.71 — Portfolio Position & Cost Basis Engine

### HF-371-01 — Exact-index readiness/backfill SQL compared `name[]` with `text[]`

Canonical record: PR #138.

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
- **Residual limitations:** Quiescence is an operational deployment requirement; no online distributed migration is claimed.

### HF-371-04 — Attempted second machine-bound `staged_rollout` authority conflicted with single-authority migration policy

- **Classification:** CI / GOVERNANCE FINDING
- **Severity:** CI blocker
- **Problem:** A governance-only change attempted to add a second machine-bound `staged_rollout` authority.
- **Root cause:** Deployment rollout evidence was being encoded in a machine-enforced authority model designed around one authority source.
- **Failure / attack scenario:** Validator must either reject the policy or be weakened to accept multiple authorities.
- **Project impact:** Weakening would make migration authority ambiguous; leaving it blocks protected CI.
- **Original / insufficient design:** Add second `staged_rollout` authority.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Restore policy manifest byte-for-byte to valid single-authority model; keep rollout requirements in amendment/evidence docs.
- **Why the remediation was selected:** Separate machine-bound migration authority from operational rollout instructions.
- **Regression coverage:** CI #380 failure followed by exact-head CI #381 success.
Canonical record: PR #138.

### HF-371-05 — Internal/adversarial evidence was published before the formal External verdict


## Stage 3.72 — Portfolio Position Projection / Cost Basis View

### HF-372-01 — Acquisition-basis allocation was undefined for nonempty zero-total-basis projection

- **Severity:** Material; `INT-372-F1`
- **Problem:** The first plan did not define `acquisitionBasisWeight` when open positions exist but every scale-8 acquisition basis rounds to zero.
- **Root cause:** The allocation formula assumed a positive denominator.
- **Failure / attack scenario:** Tiny valid open positions can produce `totalAcquisitionBasis=0.00000000`, making each weight `0/0`.
- **Project impact:** Runtime/API could fabricate zero, divide by zero, or diverge between backend/UI.
- **Original / insufficient design:** Always return Decimal allocation for every open item.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Make `acquisitionBasisWeight` `Decimal | null`; null only for nonempty-open-position/zero-total-basis.
- **Why the remediation was selected:** Represent mathematical undefined state explicitly.
- **Regression coverage:** Planning vectors and Stage 3.72 runtime/OpenAPI tests later verify zero-denominator/null behavior.
- **Residual limitations:** Independent rounded weights are not force-normalized to exactly one.

### HF-372-02 — Stage 3.71 closure wording contradicted a separately gated Stage 3.72 plan


### HF-372-03 — Published Draft still labeled itself PREPUBLICATION

Canonical record: PR #142; commit(s) `9bc2229625a47df26244421002e7f45d93a96cea`.

### HF-372-04 — Stage 3.71 lifecycle text remained future-tense after closure had already merged

Canonical record: PR #142, PR #140.

#### Historical governance chronology note — STAGE-03-72-GOV-01


```text
HISTORICAL_COMPLIANCE = NONCOMPLIANT
DISPOSITION = HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
```

The disposition does not recreate the missed temporal property and does not turn the historical interval into compliant history. This event is preserved in its dedicated governance-disposition dossier and is not counted as a fifth Stage 3.72 feature finding.


## Stage 3.74 — Transaction Correction & Reversal

### HF-374-01 — Correction command response exposed transient raw-entry status instead of logical `CORRECTED`

Canonical record: PR #150.

### HF-374-02 — Snapshot overflow error classification drifted from canonical `ErrInvalidInput`

Canonical record: PR #150.

### HF-374-03 — Replay/idempotency evidence and browser retry identity were insufficient

Canonical record: PR #150.

### HF-374-04 — HTTP and frontend contract evidence was not robust enough

Canonical record: PR #150; commit(s) `0580bf7e98c532202f84bbf9ceacd97aedbe4140`.


## Stage 3.75 — Portfolio Cash Flow & Income Truth

### HF-375-01 — OpenAPI allowed negative component aggregates that runtime defined as non-negative magnitudes

- **Severity:** P1
- **Problem:** `PortfolioCashFlowTotals` exposed deposits, withdrawals, buy/sell legs, gross income, fees and taxes as signed `Money` although runtime semantics require non-negative component magnitudes.
- **Root cause:** The public schema reused a generic signed money type instead of the stronger invariant already implemented by the accumulator.
- **Failure / attack scenario:** Clients/examples could treat negative component values as valid even though canonical runtime never permits them.
- **Project impact:** API-first contract was weaker than runtime financial truth and could allow future drift.
- **Original / insufficient design:** Use signed `Money` for all totals.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Use `NonNegativeMoney` for the eight components; keep the three net fields signed; add validator negative/positive witnesses.
- **Why the remediation was selected:** Schema now exactly mirrors accumulator sign semantics.
- **Regression coverage:** Validator rejects negative components and accepts negative net values; CI #458 / run `34152257188` 10/10.
Canonical record: PR #152; commit(s) `2f8882cc94becbec838b98f63640e2653f593eae`.
Canonical record: PR #152, PR #153.

### HF-375-02 — Gross dividend/coupon summary semantics were activated without explicit API description

- **Severity:** P1
- **Problem:** Stage 3.75 populated `PortfolioSummary.dividendsReceived` / `couponsReceived` from gross effective-ledger amounts, but the OpenAPI descriptions did not state gross/cutoff/non-net/non-performance meaning.
- **Root cause:** Implementation/UI semantics advanced without equally precise public API documentation.
- **Failure / attack scenario:** A client could interpret the fields as net receipts or investment-return metrics.
- **Project impact:** Financial meaning at the API boundary would be ambiguous.
- **Original / insufficient design:** Keep legacy field names/descriptions without explicit gross semantics.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Document gross effective-ledger amounts under summary cutoff/as-of; deductions are separate; fields are not net receipts/performance.
- **Why the remediation was selected:** Clarify semantics without breaking field names or formulas.
- **Regression coverage:** OpenAPI contract checks; replacement PR #153 revalidated the definitions; CI #459/#460 green.
Canonical record: PR #152, PR #153.
- **Residual limitations:** Tax/performance interpretation remains out of scope.

### HF-375-03 — Original PR exceeded changed-file budget without completed exception/disclosure gate

Canonical record: PR #152.

## Stage 3.76 — Manual Market Valuation & Portfolio P/L

### HF-376-01 — PR description identified superseded candidate SHA/tree/manifest

Canonical record: PR #155.

### HF-376-02 — Evidence publication changed dossier hash while migration authority_refs still pinned old hash

Canonical record: PR #155; commit(s) `f89e1e793ef505c31bf76de7d43c4d40fdd9c38b`, `e00699f8d455bcbaea0c1dc69ce534460fea6ff9`.


## Stage 3.77 — Portfolio Money-Weighted Return / XIRR

### HF-377-01 — Fixed-grid root discovery could miss close XIRR roots and derivative recursion did not shrink

Canonical record: PR #157.

### HF-377-02 — Bisection could terminate on loose NPV tolerance instead of root/bracket precision


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
- **Residual limitations:** float64 remains internal; canonical/public values are Decimal strings.

### HF-377-04 — Exponential evaluation lost coefficient magnitude or overflowed/underflowed on extreme series


### HF-377-05 — Adversarial multi-sign-change series could monopolize solver work


### HF-377-06 — Duplicate `asOfDate` query parameters were silently collapsed

- **Severity:** Blocking finding; severity not separately recorded
- **Problem:** Shared query accessor used first/peek semantics, so repeated `asOfDate` values were collapsed rather than rejected.
- **Root cause:** Stage 3.77 requires exactly one explicit BusinessDate but generic query helper did not express multiplicity.
- **Failure / attack scenario:** `?asOfDate=2026-01-01&asOfDate=2026-02-01` could use one value and ignore the conflict.
- **Project impact:** Ambiguous financial cutoff/request identity.
- **Original / insufficient design:** Read `asOfDate` through shared single-value `QueryArgs.Peek`.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** At Stage 3.77 endpoint count occurrences and require exactly one; reject missing/empty/whitespace/invalid/duplicate before store work.
- **Why the remediation was selected:** Localizes strictness to the financial endpoint.
- **Regression coverage:** Conflicting duplicate regression proves HTTP 400 and zero store calls; full Go/race CI.
Canonical record: PR #157; commit(s) `3a976568567c478a20ed51e85fb35da020132328`.
- **Residual limitations:** Unrelated endpoints keep existing query behavior.

### HF-377-07 — Published evidence state and first CI remediation did not match the actual final PR subject

- **Classification:** CI / DOCUMENTATION INTEGRITY FAMILY
- **Severity:** Blocking evidence finding plus CI blockers
- **Root cause:** Implementation evolved through dependency/test remediation, while evidence inventory lagged actual published bytes.
- **Original / insufficient design:** Keep original dependency/test fixture and 22-file dossier after remediation expanded the PR.
- **Second-order scenario:** NOT RECORDED IN CONTEMPORANEOUS EVIDENCE
- **Final remediation:** Update Next.js `16.3.3→16.3.4`, resolve Sharp `0.35.4`, pin `baseline-browser-mapping 2.11.21`, regenerate lockfile, fix canonical zero-Money fixture; update already-counted dossier path to actual 25-file state.
- **Why the remediation was selected:** Keep security gates green and evidence identity synchronized without a 26th path.
- **Regression coverage:** Protected CI 10/10 on `88b7e085...`; final `3a976568...` run `34303636331` 10/10; evidence head `b3bc17e...` run `34303895226` 10/10; evidence-only no drift.
Canonical record: PR #157; commit(s) `63d916b447e91c4de54efee5c66b27cf7727be92`.


## 9. Historical evidence gaps and irreversible governance deviations

### 9.2 Stage 3.71 evidence-withholding deviation


### 9.3 Stage 3.72 Internal-verdict withholding deviation


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
→ residual limitation
```

This reconciliation does not duplicate or rewrite Feature 3D. It uses that shape only as the normalization template for older feature history.

## 12. Future forensic documentation standard — proposed, not workflow-authoritative


1. preserve each material finding in the canonical stage/closure evidence;
3. record second-order scenarios only when they were genuinely considered;
4. use the literal `NOT RECORDED IN CONTEMPORANEOUS EVIDENCE` for historical gaps rather than inference;
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
