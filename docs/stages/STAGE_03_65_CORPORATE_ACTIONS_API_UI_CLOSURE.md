# Stage 3.65 — Corporate Actions API / UI Closure

| Field | Value |
| --- | --- |
| Status | MERGE-ACTIVATED CLOSURE RECORD — before protected activation this document is a closure candidate; once present on protected `develop`, Stage 3.64 / Feature 3C documentation closure is complete |
| Date | 2026-09-05 |
| Closed implementation stage | Stage 3.64 / Feature 3C — Corporate Actions API / UI |
| Planning authority | `docs/stages/STAGE_03_61_CORPORATE_ACTIONS_CALENDAR_PLANNING.md` |
| Domain dependency | Stage 3.62 / Feature 3A — Corporate Action Boundary |
| Projection dependency | Stage 3.63 / Feature 3B — Calendar + Heatmap Projection |
| Detailed implementation/evidence record | `docs/stages/STAGE_03_64_CORPORATE_ACTIONS_API_UI_IMPLEMENTATION.md` |
| Implementation PR | PR #128 — `feat: implement Stage 3.64 corporate actions API/UI` |
| Initial frozen V3 manifest SHA-256 | `112a4f94ce038255211e28f3b7f23b3980312dbc80a23e8b83fc664b6661bf66` |
| Final reviewed semantic/remediation head | `9bbcf6d3f0f4a3b87e06a944869fb6e4ef722784` |
| Final reviewed semantic/remediation tree | `1d24e2cd02b216b4caedb29089cd86dc97af26f9` |
| Final evidence head | `f4631c04efd0ae47eaa46d7f38ef916f350c100f` |
| Final evidence tree | `33dd90d3928286c7d2628dd56e7d9f55eece08b5` |
| Implementation squash merge | `c204ee9eee320e6171b55983cfde5cf74a2008df` |
| Protected post-merge tree | `33dd90d3928286c7d2628dd56e7d9f55eece08b5` |
| Closure base | protected `develop@c204ee9eee320e6171b55983cfde5cf74a2008df` |
| Closure base tree | `33dd90d3928286c7d2628dd56e7d9f55eece08b5` |
| Closure runtime scope | None — documentation/governance synchronization only |
| Synchronized canonical surfaces | `docs/ROADMAP.md`; `docs/SOURCE_OF_TRUTH.md`; this Stage 3.65 closure record |
| External source activation | None; Feature 3D remains separately gated |
| Architecture/API/UI debt after protected activation | CLOSED |
| Remaining Corporate Actions blocker | Feature 3D — real source adapter / source rights / licensing / rate-cost policy |

## 1. Closure basis

Stage 3.62 established the canonical Corporate Action boundary. Stage 3.63 added deterministic current Calendar and
event-density Heatmap projection semantics. Stage 3.64 exposed those reviewed semantics through a provider-neutral Go
HTTP boundary, OpenAPI contract, typed Next.js client, and presentation UI.

PR #128 completed the governed development path and was squash-merged into protected `develop` at:

```text
c204ee9eee320e6171b55983cfde5cf74a2008df
```

Post-merge protected tree:

```text
33dd90d3928286c7d2628dd56e7d9f55eece08b5
```

The post-merge tree exactly equals the approved final evidence tree. The squash merge therefore introduced no
additional runtime, test, OpenAPI, dependency, or documentation bytes beyond the approved final PR tree.

Stage 3.65 creates no second implementation event. It exists only to synchronize the final lifecycle state after the
already-completed Stage 3.64 implementation merge.

## 2. What was delivered

The merged product path is:

```text
CorporateActionProvider
        ↓
validated CorporateActionEvent[]
        ↓
Calendar projection
        ↓
Heatmap projection
        ↓
GET /api/v1/corporate-actions/projection
        ↓
typed Next.js client
        ↓
Corporate Actions Calendar / Heatmap UI
```

Stage 3.64 delivered:

- provider-neutral Corporate Actions HTTP/OpenAPI/UI integration;
- explicit source-unavailable versus legitimate-empty semantics;
- Calendar lifecycle presentation for `ANNOUNCED`, `CONFIRMED`, `PAID`, and `CANCELLED`;
- event-density Heatmap presentation without monetary aggregation;
- a 50-instrument pre-provider request cap;
- deterministic request/output scope behavior;
- `Cache-Control: no-store` while source rights remain unapproved;
- public DTO minimization that keeps provider-owned `SourceEventID` internal;
- frontend stale-result protection;
- focused backend/frontend tests and exact contract validation;
- no runtime activation of a real corporate-actions source.

The complete contract, failure mapping, DTO/UI semantics, test scope, and remediation chronology remain canonical in
the Stage 3.64 implementation/evidence record.

## 3. How and why the boundary works

The implementation deliberately follows this order:

1. validate query and canonical identifiers/dates;
2. cap the request at 50 unique instruments before any provider call;
3. validate returned canonical provider evidence;
4. resolve correction/cancellation supersession over the complete returned evidence batch;
5. filter the final dated projection to the requested inclusive effective-date range;
6. expose Calendar/Heatmap through stable transport/OpenAPI/typed-client contracts.

The order is important. Filtering provider evidence before supersession could lose a correction/cancellation needed to
remove an obsolete event. Resolving first and filtering the final dated projection preserves evidence integrity while
still enforcing the public request scope.

The API also keeps two product states separate:

```text
no approved/configured source
→ 503 CORPORATE_ACTIONS_SOURCE_UNAVAILABLE
```

versus:

```text
approved provider queried successfully, no dated current events
→ 200
→ calendar=[]
→ heatmap=[]
```

This prevents OpenInvest from falsely telling users that no dividends/coupons exist when the system actually lacks an
authorized source.

## 4. Why no money is aggregated in the Heatmap

The first Heatmap is intentionally event-density only:

```text
TotalCount
DividendCount
CouponCount
AnnouncedCount
ConfirmedCount
PaidCount
CancelledCount
```

It performs no money, yield, FX, tax, portfolio-income, or investment-return aggregation.

This keeps lifecycle certainty explicit and prevents an `ANNOUNCED` action from being presented as guaranteed income.
Monetary forecasting is a separate product/calculation problem and requires separate evidence and review.

## 5. Why provider/source details remain constrained

Three source-governance decisions are preserved:

- provider-owned opaque `SourceEventID` remains internal rather than becoming a premature public-contract dependency;
- every Corporate Actions projection response uses `Cache-Control: no-store`;
- no real provider, scraper, polling loop, worker, background ingestion, or production cache was activated.

These choices keep Feature 3C independent of a future Feature 3D provider and avoid assuming public-display,
retention, caching, or rate rights before an actual source contract is approved.

## 6. Review, remediation, and CI evidence

The authorized V3 frozen manifest was:

```text
112a4f94ce038255211e28f3b7f23b3980312dbc80a23e8b83fc664b6661bf66
```

The initial published candidate exposed real frontend test-environment defects. They were preserved and remediated
rather than hidden:

- CI #341 / run `33926915501`: frontend component tests could not load CSS modules under Node/tsx;
- remediation `cee7e1aab8b02d98e14148939773e89c11ec8bcb`: test-only CSS-module loader;
- review hardening `78e7aaade9ea7de97fbc6f91093993b88e35ef37`: stable JSDOM before dynamic React/ReactDOM imports;
- CI #343 / run `33927419572`: Next.js component test required browser-global `self`;
- remediation/final reviewed semantic head `9bbcf6d3f0f4a3b87e06a944869fb6e4ef722784`: `self: dom.window` in the test environment.

No runtime/API/domain semantic remediation was required after publication.

Final semantic/remediation CI:

```text
CI #344
run 33927609434
10/10 required jobs SUCCESS
```

Fresh External published-head review COMMENT:

```text
5118470329
VERDICT = APPROVED
P0=0 / P1=0 / P2 blocking=0 / P3 blocking=0
```

The evidence-only follow-up was:

```text
f4631c04efd0ae47eaa46d7f38ef916f350c100f
tree 33dd90d3928286c7d2628dd56e7d9f55eece08b5
```

Evidence-head CI:

```text
CI #345
run 33927918258
10/10 required jobs SUCCESS
```

Exact evidence-publication verification COMMENT:

```text
5547451403
runtime/test semantic drift = NONE
VERDICT = APPROVED
```

No unresolved review threads remained.

## 7. Human authorization and implementation merge

After exact evidence verification, the Principal Architect explicitly authorized Ready and squash merge of exact
evidence head:

```text
f4631c04efd0ae47eaa46d7f38ef916f350c100f
```

and tree:

```text
33dd90d3928286c7d2628dd56e7d9f55eece08b5
```

conditional on no drift.

Immediately before mutation, PR #128 remained open/mergeable at that exact head, CI #345 was successful, and protected
`develop` remained at the authorized base `a8f9e95c065ee708885461166e1e992d1f4aae22`.

The PR was transitioned to Ready. A second read-back confirmed no head/base drift.

Squash merge was executed with `expected_head_sha`. GitHub returned:

```text
c204ee9eee320e6171b55983cfde5cf74a2008df
```

Post-merge verification confirmed:

- PR #128 is closed and `merged=true`;
- protected `develop` points to `c204ee9eee320e6171b55983cfde5cf74a2008df`;
- its parent is the previous canonical `develop@a8f9e95c065ee708885461166e1e992d1f4aae22`;
- its tree is exactly `33dd90d3928286c7d2628dd56e7d9f55eece08b5`;
- the merge commit is GitHub-verified;
- the same ten protected-branch required checks remain configured.

## 8. What debt is closed

Once this closure record is present on protected `develop`, the Corporate Actions implementation chain is
documentation-closed as:

```text
Stage 3.62 / Feature 3A
Canonical Corporate Action Boundary       CLOSED

Stage 3.63 / Feature 3B
Calendar + Heatmap Projection              CLOSED

Stage 3.64 / Feature 3C
HTTP API + OpenAPI + Frontend UI           CLOSED
```

Therefore:

```text
Corporate Actions architecture/API/UI implementation debt = CLOSED
```

The purpose of the 3A → 3B → 3C chain was to create an honest provider-neutral Corporate Actions product surface from
canonical evidence validation through deterministic projections to API and UI, without coupling product correctness to
an unapproved external feed.

## 9. What remains

Feature 3D is a separate source/runtime-activation problem, not unfinished Feature 3C implementation.

It remains gated by:

- exact provider selection;
- source/use rights and licensing/cost acceptance;
- public-display rights;
- caching/retention rights;
- rate/traffic policy;
- failure/retry policy;
- Data Source Registry approval;
- runtime composition.

One non-blocking P3 is carried forward specifically to Feature 3D hardening: explicitly abort an in-flight Corporate
Actions request on component unmount before activating a real rate/cost-limited provider.

This closure does not approve or activate Interfax, NSD, CBR, MOEX corporate-actions ingestion, issuer scraping, or
any other provider.

## 10. Why this closure is documentation-only

Stage 3.65 changes documentation/governance only. Its complete closure candidate synchronizes exactly three canonical documentation surfaces:

1. `docs/stages/STAGE_03_65_CORPORATE_ACTIONS_API_UI_CLOSURE.md`;
2. `docs/ROADMAP.md`;
3. `docs/SOURCE_OF_TRUTH.md`.

`docs/stages/STAGE_03_64_CORPORATE_ACTIONS_API_UI_IMPLEMENTATION.md` is intentionally left byte-unchanged as historical implementation/evidence history.

It changes no:

- Go runtime/domain code;
- tests or test infrastructure;
- executable OpenAPI contract;
- frontend runtime;
- dependencies;
- CI/workflows;
- SQL/migrations;
- source adapter or source registry decision;
- caching or financial-calculation semantics.

The implementation bytes are already present on protected `develop`.

The closure exists so current lifecycle state does not have to be inferred from the historical pre-merge Status/Gate
language preserved in the Stage 3.64 implementation record.

## 11. Relationship to Stage 3.64

`STAGE_03_64_CORPORATE_ACTIONS_API_UI_IMPLEMENTATION.md` remains authoritative for the detailed technical contract,
implementation reasoning, failure semantics, review findings, remediation history, and pre-merge evidence chain.

Its pre-merge lifecycle wording is retained as immutable historical evidence of the state when that evidence record was
published.

Stage 3.65 supersedes only the **current lifecycle/closure status**:

```text
Stage 3.64 / Feature 3C = CLOSED
Corporate Actions architecture/API/UI debt = CLOSED
Feature 3D real source adapter = NOT ACTIVATED / separately gated
```

## 12. Governance / Closure path

This is an eligible post-development governance/closure change under `docs/REVIEW_WORKFLOW.md` v1.4.0.

The complete candidate is documentation-only and contains exactly the Stage 3.65 closure record plus synchronized
`ROADMAP.md` and `SOURCE_OF_TRUTH.md`. The designated review chat performs one read-only Governance / Closure review;
no second development-path Internal/External cycle is required.

Protected activation sequence:

```text
documentation candidate
→ deterministic documentation checks
→ Governance / Closure review
→ explicit human commit/push authorization
→ Draft PR targeting develop
→ required GitHub CI
→ exact-published-head Governance / Closure verification
→ explicit human merge authorization
→ squash merge to protected develop
```

Draft publication, green CI, review approval, or Ready state alone does not activate this record. Protected `develop`
is the activation boundary.

## 13. Closure decision

The Stage 3.64 implementation is already technically complete and merged.

Before Stage 3.65 protected activation, the only remaining Stage 3.64 gap is documentation lifecycle synchronization.

Once this exact closure record and its synchronized `ROADMAP.md` / `SOURCE_OF_TRUTH.md` state are squash-merged into protected `develop`:

```text
Stage 3.64 / Feature 3C documentation closure = COMPLETE
Corporate Actions architecture/API/UI implementation debt = CLOSED
Remaining Corporate Actions work = Feature 3D source/licensing/runtime-source activation only
```

Nothing in this candidate authorizes Feature 3D implementation, source activation, branch deletion, direct protected
mutation, Ready, or merge.
