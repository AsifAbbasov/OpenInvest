# Stage 3.66 — Corporate Actions In-Flight Request Cancellation Implementation

| Field | Value |
| --- | --- |
| Status | Draft PR #130 published; fresh External review APPROVED; this evidence-only follow-up publishes withheld Internal evidence; final evidence-head CI/verification and separate Ready/merge authorization remain pending |
| Date | 2026-09-05 |
| Canonical implementation base | `develop@1c30a4bf637c933e7c210cff6e26fabd91d8bab1` |
| Protected-base tree | `4ba610c1d20c95f10a5a16a6a0ece6caceb3236e` |
| Prior feature authority | Stage 3.64 / Feature 3C — Corporate Actions API/UI |
| Closure authority | Stage 3.65 — Corporate Actions API/UI Closure |
| Carry-forward finding | Stage 3.64/3.65 non-blocking P3: explicit component-unmount request cancellation before any real rate/cost-limited provider activation |
| Feature 3D | NOT STARTED / NOT AUTHORIZED |
| External source activation | None |

## 1. Governance and stage identity

Protected `develop` already contains canonical Stage 3.65 as the documentation/governance closure for Stage 3.64.
Therefore this implementation must not reuse the number `3.65`. The next free stage number is `3.66`.

`docs/REVIEW_WORKFLOW.md` v1.4.0 requires an **Approved development scope** for development-path changes; it does not
mandate a separate planning-only PR for every narrow implementation. This scope is already bounded by the preserved
Stage 3.64/3.65 P3 and the Principal Architect directive that authorizes only Corporate Actions frontend request
cancellation hardening while explicitly forbidding Feature 3D and real source work.

Therefore Stage 3.66 may proceed directly as a narrow development-path implementation candidate. Commit/push/Draft PR
remain separately gated after local checks and mandatory Internal review.

## 2. Purpose

Close the preserved Corporate Actions frontend lifecycle P3 before Feature 3D by ensuring that an in-flight
`GET /api/v1/corporate-actions/projection` request is explicitly cancelled when its owning component unmounts, while
preserving the Stage 3.64 stale-result and error semantics.

This stage is transport-lifecycle hardening only. It does not change Corporate Action domain events, projection rules,
HTTP/OpenAPI contracts, provider/source semantics, persistence, caching, or financial calculations.

## 3. Contract

The Stage 3.66 frontend contract is:

1. every Corporate Actions request remains owned by its request-local `AbortController`;
2. the existing `AbortSignal` continues to flow through `getCorporateActionProjection` into native `fetch`;
3. component unmount invalidates the active request generation and aborts the current controller;
4. query replacement or resubmit continues to abort the previous controller;
5. a result belonging to an aborted controller is ignored even if the lower-level typed HTTP client converts the
   fetch rejection into its existing generic failed `ApiResult`;
6. generation mismatch remains a second independent stale-result guard;
7. cancellation is a UI lifecycle event and MUST NOT become `unavailable` or `error` presentation state;
8. genuine `503` source-unavailable, `502` source-invalid, other HTTP/API failures and network failures that belong to
   the current non-aborted request preserve existing Stage 3.64 behavior.

No new global request abstraction, state manager, query library, dependency, retry mechanism, cache, worker, or provider
surface is introduced.

## 4. Existing architecture reused

Stage 3.64 already provides:

```text
CorporateActionsSlice
  → AbortController per request
  → getCorporateActionProjection({ ..., signal })
  → request(... RequestInit.signal ...)
  → native fetch
```

It also already invalidates request generations when Corporate Actions inputs change and ignores stale completions.
The missing lifecycle edge is component unmount.

Because `CorporateActionProjectionParams.signal?: AbortSignal` and the existing typed client already forward the signal,
Stage 3.66 does **not** modify `frontend-next/src/common/api/openinvest.ts`. Reusing that seam is smaller and safer than
introducing a Corporate-Actions-specific HTTP wrapper or changing the shared `ApiResult` union for unrelated consumers.

## 5. Implementation

Candidate runtime change:

`frontend-next/src/features/corporate-actions/components/CorporateActionsSlice.tsx`

Add one mount-lifetime cleanup effect:

```text
component cleanup
  → increment requestGenerationRef
  → abortRef.current?.abort()
  → clear abortRef
```

The generation is invalidated before abort so any continuation scheduled by abort rejection is stale before it can
attempt a state transition.

After awaiting the typed client, acceptance becomes:

```text
if controller.signal.aborted
  → ignore
else if generation mismatch
  → ignore
else
  → preserve existing Stage 3.64 response handling
```

This is intentionally redundant with the generation guard: controller state proves explicit cancellation, while the
generation guard still protects against stale completion races and mocks/transports that do not reject immediately on
abort.

## 6. Failure semantics

| Condition | Required UI effect |
| --- | --- |
| component unmount with active request | abort transport; no state commit |
| input/query change with active request | abort old request; invalidate generation; return to existing idle semantics |
| replacement request | old request aborted/stale; newest valid generation alone may commit |
| aborted fetch rejected as `AbortError` and shared client maps it to generic failed `ApiResult` | ignored because owning controller is aborted / generation stale; no false user error |
| stale request completes despite abort | ignored by aborted-signal or generation guard |
| current request returns HTTP 503 | existing `unavailable` state |
| current request returns other HTTP/API failure | existing `error` state with server message |
| current request has genuine network failure | existing generic Go API unavailable message |
| current request succeeds empty | existing legitimate `empty` state |
| current request succeeds populated | existing `ready` state |

Cancellation does not alter provider/application failure taxonomy because no provider or backend contract changes in
this stage.

## 7. Changed files

The local candidate is limited to exactly three files:

1. `frontend-next/src/features/corporate-actions/components/CorporateActionsSlice.tsx`;
2. `frontend-next/tests/corporate-actions-slice.component.test.tsx`;
3. `docs/stages/STAGE_03_66_CORPORATE_ACTIONS_REQUEST_CANCELLATION_IMPLEMENTATION.md`.

No `openinvest.ts`, OpenAPI, backend Go, dependency, package lock, SQL, migration, CI workflow, Data Source Registry,
provider adapter, cache, worker, or source activation file is changed.

## 8. Test contract

The component test suite retains all existing Stage 3.64 Corporate Actions tests and adds focused proof for:

- active request receives an `AbortSignal` and is aborted on component unmount;
- an aborted request cannot update state after query invalidation;
- abort does not surface the shared client's generic network-failure text as a false application/provider error;
- a replacement request aborts its predecessor and the newer result wins;
- a genuine Corporate Actions HTTP/API error still renders the existing alert/error state;
- a genuine non-abort network failure still renders the existing generic Go API unavailable error;
- the existing stale-completion test remains unchanged in intent and continues proving a late completion cannot
  overwrite a newer result.

The abort-aware fetch fixture rejects with a browser `DOMException` named `AbortError`, matching native fetch abort
semantics rather than resolving an artificial success value.

## 9. Required quality gates

Before publication the development path requires, where executable in the local environment:

- focused Corporate Actions component tests;
- complete frontend test suite;
- frontend typecheck;
- frontend production build;
- TypeScript/TSX syntax/static preflight;
- changed-file/scope audit;
- mandatory Internal read-only review across all changed lines.

Available deterministic prepublication evidence:

- exact base reconstruction for `CorporateActionsSlice.tsx`: Git blob `6a0f092ada3613052acee7a00cb7d08e642ba272`, matching protected `develop`;
- exact base reconstruction for `corporate-actions-slice.component.test.tsx`: Git blob `d93df735bab8a16cb37bff76ee62009585d7f2f6`, matching protected `develop`;
- TypeScript/TSX parser preflight on the changed runtime and test files: PASS;
- candidate runtime stub typecheck for the modified lifecycle surface: PASS;
- whitespace / diff-check preflight: PASS;
- scope guard: no provider/source, direct fetch, polling, cache, DB/migration, or worker code added to production component: PASS;
- focused component-test inventory: 10 tests total (all 5 existing Stage 3.64 tests retained plus 5 new cancellation/error tests): PASS as source inventory.

The current execution environment does not contain the repository checkout or frontend runtime dependencies
(Next/React/jsdom/tsx), and its Node runtime is `22.16.0` while the repository requires `>=22.22.2`. Therefore the real
component tests, complete frontend test suite, canonical frontend typecheck and production build cannot be truthfully
reported as local PASS. Those execution results remain `UNKNOWN` until the exact candidate runs in the repository
environment after separately authorized publication triggers authoritative GitHub CI.

After publication authorization, required GitHub CI must pass all protected-branch jobs on the exact PR head before a
fresh External published-head review.

## 10. Architectural consequences

Positive consequences:

- frontend request lifetime now matches component lifetime;
- abandoned navigation no longer leaves avoidable Corporate Actions HTTP work in flight;
- the future cost/rate impact of a real provider is reduced before Feature 3D exists;
- stale-result correctness is strengthened without changing the API, domain, provider, or state architecture;
- the existing typed HTTP client's `AbortSignal` seam is proven sufficient; no new abstraction is needed.

Deliberately unchanged consequences:

- the shared `request()` helper still maps low-level fetch exceptions to its existing generic network `ApiResult`;
- Stage 3.66 does not globally redesign cancellation semantics for Asset Discovery or other callers;
- Corporate Actions suppresses cancellation at the owning component boundary using controller/generation identity;
- if a future cross-feature requirement emerges for typed cancellation as a first-class `ApiResult`, that must be a
  separately reviewed shared-client scope rather than being smuggled into this P3.

## 11. Explicit deferred / forbidden scope

Stage 3.66 does NOT authorize or implement:

- Feature 3D;
- any real Corporate Actions provider;
- Interfax, NSD, CBR, MOEX Corporate Actions, issuer scraping or other external source;
- Data Source Registry activation;
- backend provider composition;
- OpenAPI/API changes;
- SQL, database, migration or persistence;
- Redis or any cache;
- worker, polling, retry loop, background task or prefetch;
- new production dependency;
- Redux, TanStack Query or other client-state framework;
- financial calculations, monetary heatmap, FX, tax, yield or portfolio-income forecast;
- unrelated Asset Discovery lifecycle remediation.

## 12. Review standard

Mandatory review order:

```text
contract
→ implementation
→ failure cases
→ tests
→ CI
→ architectural consequences
```

The Internal reviewer must classify only demonstrated findings as P0/P1/P2/P3 and finish with exactly one verdict:
`APPROVED`, `REQUEST CHANGES`, or `BLOCKED — insufficient evidence`.

If Internal review is `APPROVED`, human permission is still required before commit/push/Draft PR. Publication then
requires exact-head GitHub CI, fresh External review, any demonstrated remediation, evidence-only publication and
verification, and a separate human Ready/squash-merge gate under `docs/REVIEW_WORKFLOW.md` v1.4.0.


## 13. Published review and evidence chronology

This section is evidence-only and is published only after the fresh External published-head verdict, as required by
`docs/REVIEW_WORKFLOW.md` v1.4.0. Sections 1–12 preserve the Stage 3.66 implementation contract and prepublication
reasoning. This follow-up changes no runtime, tests, typed client, dependency, OpenAPI, backend, source or Feature 3D
surface.

### 13.1 Frozen prepublication subject

- canonical base: `develop@1c30a4bf637c933e7c210cff6e26fabd91d8bab1`;
- base tree: `4ba610c1d20c95f10a5a16a6a0ece6caceb3236e`;
- frozen three-file manifest SHA-256: `f6097cc169e046e598135877eafbab23a11cb2035fbd06972f293541048f1bbe`;
- production blob: `c4e36a8cef8965fea21223a9b455820a734221b5`;
- component-test blob: `b9f9d2eb40dc1c9f09b8ab4dd3cf15755a1e35f1`;
- initial implementation-record blob: `08f325c49b74b0b0a75737f15245718ec5349686`.

The exact semantic publication was one commit ahead / zero behind its authorized base and changed exactly the three
files named in Section 7.

### 13.2 Internal review evidence — published after External verdict

The mandatory prepublication Internal review was read-only and made no repository edits. Its evidence was withheld
from the Draft PR/repository until the fresh External published-head verdict.

Internal review SHA-256:

```text
70e1b8fb0f40941137d5350831aa1bdb0c35db135ba57610ac5e8bbade5c426c
```

Final prepublication Internal verdict:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

One non-runtime P3 test-quality finding was resolved before freeze: the initial unmount test did not guarantee React
root cleanup after an early assertion failure. Builder remediation added fail-safe teardown, reran the affected local
preflight, and retained sequential test isolation. A separate genuine network-failure regression was also added so
abort suppression cannot accidentally hide normal transport failures. No runtime finding remained.

### 13.3 Published semantic head and authoritative CI

Draft PR #130 initially published exact semantic head:

```text
97694ddfe49a1587aa4e86a6c0258a57fd95a708

tree:
067cdb5a69d9aa9e4d18ec76a372d836d1b3c13c
```

CI #347 / run `33952598235` completed on that exact head with all ten required jobs `SUCCESS`:

```text
Go tests                         PASS
Go race tests                    PASS
Go vet                           PASS
Go vulnerability scan           PASS
Python tests                     PASS
Frontend build and typecheck     PASS
OpenAPI contract                 PASS
Docker Compose config            PASS
PostgreSQL migration validation  PASS
Dependency security scan         PASS
```

The frontend job specifically passed all three relevant steps: `Typecheck`, complete `Test`, and production `Build`.
This is the authoritative execution evidence for the real React/JSDOM cancellation tests that were unavailable in the
prepublication sandbox.

### 13.4 Fresh External published-head review

Fresh External review COMMENT `5550298926` reviewed the exact semantic head and did not use the earlier Internal
verdict/findings as supporting evidence. It re-evaluated the contract, implementation, failure/race semantics, tests,
CI and architectural consequences.

External verdict:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The External review confirmed that unmount/replacement cancellation is confined to the Corporate Actions frontend
ownership boundary, the existing generation guard is preserved, normal API/network errors keep their prior semantics,
and Feature 3D/source activation remains outside the PR.

### 13.5 Evidence-only publication rule

This follow-up publishes only the previously withheld Internal evidence and the already-observed published-head
CI/External chronology in this implementation record. The semantic runtime/test blobs from Section 13.1 remain
unchanged.

The evidence head must pass the same required GitHub CI. The designated review chat must then verify that the
semantic-head → evidence-head transition changes only this documentation record and introduces no runtime/test
semantic drift. Ready/squash merge remains a separate explicit human authorization gate. Feature 3D remains not
authorized.
