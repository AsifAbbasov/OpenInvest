# Stage 3.46 — P3-06 HTTP API Decomposition Plan

| Field | Value |
| --- | --- |
| Status | Planning-only governance scope; this artifact grants no runtime refactor authorization |
| Date | 2026-08-30 |
| Canonical planning base | `develop@c029cd62715b15614e82972309bdc53669ec02ee` |
| Protected-base tree | `cf4cc749d2065c73eea7df4e46eca972a2036794` |
| Finding | Original audit `P3-06 — httpapi/api.go decomposition` |
| Prior closure | P3-10 CLOSED after Stage 3.45 closure PR #105 squash-merged the approved closure surfaces into protected `develop` at `c029cd62715b15614e82972309bdc53669ec02ee` |
| Planning-base audit state | At `develop@c029cd62715b15614e82972309bdc53669ec02ee`: 29 / 32 closed = 90.625%; remaining P3-06, P3-07, P3-08 |
| Runtime implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / Ready / merge authorized here | No |

## 1. Objective

Close original audit finding P3-06 by decomposing the oversized
`backend-go/internal/httpapi/api.go` source file into cohesive same-package HTTP transport files while
preserving the already-accepted runtime, API, security, idempotency, import, pagination and error
semantics.

This stage is planning-only. It does not move runtime declarations, change routes, modify tests,
change OpenAPI, alter dependencies, or close P3-06.

## 2. Exact repository baseline

At protected `develop@c029cd62715b15614e82972309bdc53669ec02ee`:

- protected-base tree:
  `cf4cc749d2065c73eea7df4e46eca972a2036794`;
- `backend-go/internal/httpapi/api.go` blob:
  `f05b7288311918e95c7484dc2cc7385f2de3c03c`;
- `backend-go/internal/httpapi/api.go` size:
  `67204` bytes;
- `docs/SOURCE_OF_TRUTH.md` blob:
  `e5d3d042d4f65520bb8d64557e73504852708a1a`;
- `docs/ROADMAP.md` blob:
  `5e8bf10c8fef0cba264528b090aaa6fcab7ae1f5`;
- `docs/REVIEW_WORKFLOW.md` blob:
  `06f9cabd04e6791be1892ae1f0eae8d915fddc02`;
- Stage 3.45 closure dossier blob:
  `727f2d29ad27f662891c5f4e15a7391e941bdc54`.

At the protected Stage 3.46 planning base, the `httpapi` package is not literally a one-file package. Existing companion files already
separate some concerns, including password JSON handling, replay/idempotency handling and related
tests. P3-06 therefore targets the remaining concentration inside `api.go`; it does not authorize
rewriting already-separated helpers merely to achieve stylistic uniformity.

## 2.1 Planning review history — append-only

The first complete Stage 3.46 planning review returned `REQUEST CHANGES` with two P3 findings:

### `PLAN-STAGE-03-46-P3-01` — asymmetric declaration-inventory scopes

Problem:
the v1 plan required a baseline inventory only for top-level declarations in `api.go`, but compared
that with a post-decomposition package-wide inventory. Because `httpapi` already contains production
companion files, those scopes were not directly comparable and could not deterministically prove that
package-level declarations outside `api.go` were preserved.

Remediation in this v2 candidate:
- freeze one deterministic **whole-production-package baseline inventory** and one deterministic
  **whole-production-package target inventory** using the exact same scope;
- record declaration name, receiver, signature, source file and deterministic structural/source
  identity sufficient to detect loss, duplication, signature change, body change or unexpected new
  production declarations;
- keep tests outside the production declaration inventory and, when test inventory is used, compare
  baseline and target using the same test scope;
- separately preserve an `api.go` declaration-to-target-file relocation map.

Fresh v2 re-review outcome:
the designated reviewer found this substantive remediation sufficient and marked
`PLAN-STAGE-03-46-P3-01` `RESOLVED`. The same v2 re-review nevertheless returned overall
`REQUEST CHANGES` because it discovered separate publication-stability finding
`PLAN-STAGE-03-46-P3-03`.

This historical outcome is preserved as review evidence; it is not a mutable active-state gate.

### `PLAN-STAGE-03-46-P3-02` — incomplete canonical review-size budget

Problem:
the v1 plan froze the repository file-count budget but omitted the canonical
`<=800 changed lines of hand-written business logic` limit and the required Principal Architect
exception before review begins when a default budget is exceeded.

Remediation in this v2 candidate:
- freeze both canonical default limits:
  - `<=25 changed files`;
  - `<=800 changed lines of hand-written business logic`;
- require a fail-closed stop before Internal review when either limit is exceeded unless the canonical
  exception is documented and explicitly approved by the Principal Architect before review begins;
- require a return to planning before silently splitting the remediation into multiple implementation
  PRs; any staged strategy must define lifecycle, evidence and closure semantics first.

Fresh v2 re-review outcome:
the designated reviewer found this substantive remediation sufficient and marked
`PLAN-STAGE-03-46-P3-02` `RESOLVED`. The same v2 re-review nevertheless returned overall
`REQUEST CHANGES` because it discovered separate publication-stability finding
`PLAN-STAGE-03-46-P3-03`.

This historical outcome is preserved as review evidence; it is not a mutable active-state gate.

No other finding from the first Stage 3.46 planning review is erased or rewritten.

### `PLAN-STAGE-03-46-P3-03` — lifecycle-sensitive review-state wording

The fresh complete v2 planning re-review returned `REQUEST CHANGES` for one new P3 finding after
confirming `PLAN-STAGE-03-46-P3-01` and `PLAN-STAGE-03-46-P3-02` substantively resolved.

Problem:
the v2 planning document retained mutable active-state wording that would become false immediately
after the very re-review required before publication. Specifically, it described itself as a
`Planning/review candidate only` and permanently encoded P3-01/P3-02 as
`PENDING FRESH COMPLETE RE-REVIEW`.

Root cause:
the Builder had documented a lifecycle/publication-stability control but did not apply a final
transition simulation to the exact candidate after incorporating the previous reviewer findings.
The control existed as prose but was not enforced against the artifact that was handed to review.

Impact:
publishing v2 unchanged would have made the canonical planning record stale at the moment the fresh
review completed.

Remediation in this v3 candidate:
- replace candidate-state status wording with stable planning-scope wording;
- preserve P3-01/P3-02 as historical resolved findings from the completed v2 re-review rather than
  mutable pending states;
- keep the v2 overall `REQUEST CHANGES` and P3-03 discovery append-only;
- do not encode a mutable `P3-03 = pending review` field in the permanent planning artifact;
- establish P3-03 disposition through designated review evidence bound to the exact v3 candidate,
  not through a self-staling active-state sentence inside the candidate;
- strengthen the mandatory Builder anti-regression preflight with an exact-candidate
  post-remediation transition simulation and a prohibition on reviewer-state leakage into permanent
  artifacts.

The Builder does not self-issue a canonical verdict for P3-03. Its disposition is determined only by
the designated review chat over the exact v3 candidate.

Fresh v3 re-review outcome:
the designated reviewer found `PLAN-STAGE-03-46-P3-03` `RESOLVED` and reviewer-result/permanent-artifact
separation `PASS`, but returned overall `REQUEST CHANGES` for new P3 finding
`PLAN-STAGE-03-46-P3-04`.

### `PLAN-STAGE-03-46-P3-04` — incomplete semantic lifecycle simulation

Problem:
the v3 exact-candidate scan was intentionally focused on recurrence of the P3-03 reviewer-state
wording class. It did not semantically evaluate every lifecycle-sensitive assertion across the full
planning -> implementation -> closure lifecycle required by the plan itself.

The v3 permanent planning artifact therefore retained unqualified mutable facts such as
`Current audit state`, `Current concentration` and `The current api.go...`. Those statements were
true at the planning base but would become false after the planned implementation and protected
P3-06 closure activation.

Root cause:
the Builder treated a keyword recurrence scan as if it were equivalent to a complete semantic
transition simulation. The process checked dangerous words, but did not classify each mutable fact by
its temporal anchor and terminal lifecycle truth.

Impact:
publishing v3 unchanged would have created a permanent planning record that self-staled when the very
remediation described by the plan completed.

Remediation in this v4 candidate:
- anchor mutable repository/audit facts explicitly to the immutable Stage 3.46 planning base;
- rename `Current audit state` to `Planning-base audit state`;
- rename the concentration section to `Planning-base concentration`;
- rewrite `The current api.go...` as an explicit protected-planning-base fact;
- rename the later audit block to `Stage 3.46 planning-base original audit state`;
- replace the narrow recurrence-only lifecycle scan with a complete semantic transition simulation
  covering the exact final candidate through review, commit, push, Draft PR, CI, External verdict,
  evidence publication, exact verification, Ready, implementation merge and closure;
- add a permanent anti-regression class for temporal anchoring of mutable baseline facts.

The Builder does not self-issue a canonical verdict for P3-04. Its disposition is determined only by
the designated review chat over the exact v4 candidate.

Fresh v4 re-review outcome:
the designated reviewer confirmed that v4 materially improved temporal anchoring but found
`PLAN-STAGE-03-46-P3-04` **NOT RESOLVED**. No new finding ID was created.

The remaining defect was one unqualified mutable source-layout statement in the final Planning
Decision:

`split the current 67,204-byte httpapi/api.go...`

The v4 temporal assertion inventory detected that line, but the v4 semantic lifecycle simulation did
not explicitly classify it and nevertheless returned a Builder `PASS`. This exposed a distinct
process weakness inside the P3-04 class: detection coverage and semantic-classification coverage were
not required to be identical.

v5 remediation:
- anchor the Planning Decision to the immutable Stage 3.46 planning base;
- require one-to-one coverage between every temporal assertion inventoried and every assertion
  semantically classified;
- forbid an overall semantic `PASS` when any inventoried assertion lacks an explicit classification;
- explicitly test the Planning Decision at T9 implementation merge and T11 protected closure;
- add a permanent cross-class protection for detector/classifier coverage completeness.

The Builder does not self-issue a canonical resolution for P3-04.

Fresh v5 re-review outcome:
the designated reviewer kept `PLAN-STAGE-03-46-P3-04` **NOT RESOLVED** with no new finding ID.
The corrected Planning Decision and its dedicated T9/T11 checks passed, but the reviewer found that
the v5 temporal inventory was still not semantically exhaustive:

- the inventory/classifier equations proved only equality inside the detector's own population;
- keyword-based discovery omitted semantically important state assertions that lacked those keywords;
- the exact 16-route frozen state was not inventoried as one complete assertion;
- planning-base `api.go` blob/size and complete audit arithmetic were only partially inventoried;
- some future/workflow statements were classified too generically as historical facts.

v6 remediation:
- semantic discovery starts from the **entire exact candidate source**, not from keywords;
- every candidate line is assigned exactly once to a semantic coverage block or explicit
  formatting/non-assertion class;
- every semantic coverage block receives an explicit temporal category and T0→T11 truth behavior;
- the exact 16-route set is one compound invariant with all 16 method/path pairs;
- the planning-base `api.go` identity/size is one complete compound baseline assertion;
- pre-closure and conditional post-closure audit arithmetic each include count, percentage and
  remaining-finding components;
- source-to-ledger coverage equations are mandatory in addition to assertion-to-classifier equality;
- a new permanent cross-class protection is added for **semantic discovery completeness /
  source-to-ledger coverage**.

The Builder does not self-issue a canonical resolution for P3-04. Its disposition remains exclusively
a designated-review outcome bound to the exact candidate under review.

## 3. Planning-base concentration that P3-06 addresses

At the protected Stage 3.46 planning base, `api.go` simultaneously owns multiple distinct transport responsibilities, including:

- `API` construction and Fiber app bootstrap;
- route registration;
- local-development CORS admission;
- authentication rate limiting;
- health/readiness handlers;
- auth register/login/refresh/logout handlers;
- asset handlers;
- portfolio handlers;
- transaction handlers;
- import review/append HTTP orchestration;
- pagination/cursor support;
- import-review token/security support;
- request/response transport helpers, DTO/mapping/error support that still resides in this file.

The finding is maintainability/structure debt. This plan does not claim that the Stage 3.46 planning-base file is a
demonstrated security exploit, API-contract defect, or user-visible functional failure.

## 4. Frozen remediation strategy

The implementation SHALL be a **same-package mechanical decomposition**.

Core rules:

1. Keep package name `httpapi`.
2. Preserve exported package API.
3. Preserve existing handler/helper symbol names unless a new private routing helper is required.
4. Prefer moving complete declarations intact rather than rewriting them.
5. Do not introduce a new abstraction layer, framework, router wrapper, service boundary, interface
   hierarchy, generic handler framework or dependency.
6. Do not move domain/application logic into HTTP transport.
7. Do not move HTTP transport logic into domain/application packages.
8. Do not combine this finding with behavior cleanup.
9. Do not change an expectation merely to make a test pass after refactoring.
10. Any behavior difference discovered during implementation must fail closed and be reviewed before
    scope expansion.

## 5. Target responsibility boundaries

The implementation should separate the Stage 3.46 planning-base declarations into cohesive families equivalent to the
following logical boundaries. Exact private file names may vary if review finds a clearer same-package
layout, but the responsibility boundaries themselves are frozen.

### A. Bootstrap and API state

`api.go` after remediation should contain only the minimal package bootstrap/state responsibility:

- `API` state required by HTTP transport;
- `New`;
- `NewDevelopment`;
- minimal app construction;
- a call into route registration if route registration is moved.

`api.go` must not retain endpoint handler bodies merely to avoid moving code.

### B. Route registration

A dedicated routing responsibility should own the exact existing method/path registrations.

The public route set frozen by this plan is:

1. `GET /api/v1/health`
2. `GET /api/v1/ready`
3. `POST /api/v1/auth/register`
4. `POST /api/v1/auth/login`
5. `POST /api/v1/auth/refresh`
6. `POST /api/v1/auth/logout`
7. `GET /api/v1/assets/search`
8. `GET /api/v1/assets/:ticker`
9. `GET /api/v1/portfolios`
10. `POST /api/v1/portfolios`
11. `GET /api/v1/portfolios/:portfolioId`
12. `GET /api/v1/portfolios/:portfolioId/summary`
13. `GET /api/v1/portfolios/:portfolioId/transactions`
14. `POST /api/v1/portfolios/:portfolioId/transactions`
15. `POST /api/v1/portfolios/:portfolioId/imports/review`
16. `POST /api/v1/portfolios/:portfolioId/imports/append`

The decomposition may not add, remove, rename, reorder semantically significant middleware, or
retarget any route.

### C. CORS

Local-development CORS admission belongs in a focused transport file.

The implementation must preserve:

- `OPENINVEST_ALLOWED_WEB_ORIGINS`;
- the existing localhost defaults;
- whitespace handling;
- exact credential behavior;
- allowed methods;
- allowed headers;
- exposed headers;
- `Vary: Origin`;
- OPTIONS handling and status.

### D. Authentication transport

Authentication transport should own:

- register/login/refresh/logout HTTP handlers;
- refresh-cookie transport helpers that belong to these HTTP flows at the Stage 3.46 planning base.

No authentication-domain policy is changed.

### E. Authentication rate limiting

Rate-limiter state and helpers should be isolated from general handlers.

The decomposition must preserve:

- per-key limit behavior;
- global limit behavior;
- max-key bound;
- time window behavior;
- cleanup/sweep behavior;
- path plus IP keying;
- fail-closed handling for invalid/zero capacity;
- retry/error mapping behavior.

### F. Assets

Asset HTTP handling should own asset search/detail transport and asset-specific pagination/cursor
coordination.

No asset-search validation, Unicode bound, cursor identity or result-envelope behavior changes.

### G. Portfolios

Portfolio HTTP handling should own list/create/detail/summary transport and portfolio-specific
pagination/cursor coordination.

Idempotency behavior remains exactly governed by the already-existing replay/idempotency components.

### H. Transactions

Transaction HTTP handling should own list/append transport and transaction-specific pagination/cursor
coordination.

This stage must not absorb P3-07 transaction-form fixture/default semantics.

### I. Imports

Import HTTP handling should own review/append orchestration and the import-review token/security
helpers that are still part of `api.go`, unless an already-existing focused companion file is the
clearer home.

The decomposition must preserve all import review-token, row-count, payload-size, source identity,
file-hash, replay/idempotency and parser-version semantics.

### J. Shared HTTP transport helpers

Only genuinely cross-handler transport helpers may live in a shared HTTP helper/DTO/response file.

This is not authorization to create a generic framework. A helper used by one functional family
should stay with that family when practical.

## 6. Behavior-preservation contract

P3-06 is approved only as a behavior-preserving structural remediation.

Implementation must preserve all of the following:

- exported constructors and their signatures;
- Fiber app configuration;
- all 16 route method/path pairs;
- middleware order and CORS placement;
- HTTP status codes;
- error codes/messages/envelopes;
- response DTO JSON field names and omission behavior;
- strict JSON decoding behavior;
- request body/header/query/path/cookie access semantics;
- `fiber.Ctx` usage;
- `c.Context()` propagation into application/auth services;
- authentication register/login/refresh/logout behavior;
- refresh-cookie set/clear behavior;
- CSRF-related transport behavior;
- auth rate-limit semantics;
- asset query validation and pagination;
- portfolio pagination and idempotent create behavior;
- transaction filtering/pagination/append behavior;
- import review and append behavior;
- import payload and row limits;
- pagination cursor encoding/validation/version/limits;
- import-review token encoding/signing/validation/version/TTL/limits;
- request metadata and tracing/error-envelope behavior;
- idempotency replay behavior;
- OpenAPI-visible behavior.

A mechanical move is not permission to normalize an odd-looking existing behavior.

## 7. Explicitly forbidden semantic changes

The implementation SHALL NOT:

- change a route;
- rename a JSON field;
- change a status code or error code;
- change validation order;
- change trim/Unicode semantics;
- change token/cursor wire format or secret derivation;
- change token/cursor versions or TTLs;
- change authentication/session policy;
- change CORS policy;
- change idempotency request identity or replay authority;
- change import parser/version semantics;
- change financial calculations;
- change database/schema/migrations;
- change executable OpenAPI;
- change Go dependencies;
- change Fiber version;
- change frontend behavior;
- absorb P3-07 or P3-08.

If implementation discovers a real functional defect, it must be reported separately and must not be
silently fixed inside the decomposition.

## 8. Existing companion-file protection

Existing focused `httpapi` files are not automatically in implementation scope merely because they are
in the same package.

Known already-separated surfaces include, among others:

- `auth_password_json.go`;
- `idempotency_replay.go`;
- `idempotent_handlers.go`;
- `idempotent_import_handler.go`;
- `import_replay_recovery.go`;
- `replay_app.go`;
- existing focused test files.

A companion file may change only when a move from `api.go` requires a minimal compile/gofmt/test
adjustment that is directly attributable to the decomposition. Unrelated cleanup is forbidden.

## 9. Expected implementation shape

The future implementation is expected to:

1. change `backend-go/internal/httpapi/api.go`;
2. add a bounded set of cohesive `.go` files under the same
   `backend-go/internal/httpapi` package;
3. add or minimally adjust focused tests only where needed to prove structural and behavior
   invariants;
4. add a Stage 3.47 implementation/evidence dossier.

No package move and no new Go module is expected.

The implementation must remain within the complete canonical default review-size budget defined by
`docs/REVIEW_WORKFLOW.md` v1.3.0:

- `<=25 changed files`; and
- `<=800 changed lines of hand-written business logic`.

Before Internal review begins, the Builder must calculate both limits against the complete candidate.

If either default limit is exceeded, the implementation must **fail closed before review** unless a
canonical exception is documented with the reason, bounded review strategy and explicit Principal
Architect approval obtained before review begins.

The Builder may not evade the limit by silently splitting one reviewed responsibility across multiple
implementation PRs. If a staged implementation becomes necessary, stop and return to planning so the
multi-PR lifecycle, exact boundaries, review evidence, intermediate protected states and eventual
P3-06 closure semantics are explicitly frozen before code is changed.

If the decomposition requires an architecture change, stop and return to planning.

## 10. Structural acceptance criteria

Implementation evidence must prove:

1. `api.go` no longer contains endpoint handler bodies;
2. `api.go` is reduced to bootstrap/API-state responsibility;
3. route registration is readable as one coherent transport responsibility;
4. auth rate limiting is not mixed into unrelated handler files;
5. auth, asset, portfolio, transaction and import handler responsibilities are separable by file;
6. import-review token/cursor/security helpers have a clear focused home;
7. common transport helpers are not turned into a generic mini-framework;
8. no duplicate implementation of moved declarations remains;
9. no production symbol was accidentally lost;
10. exported `httpapi` package surface is unchanged;
11. all moved code remains in package `httpapi`;
12. no import cycle or new dependency is introduced.

A numeric line-count target is intentionally not the closure criterion. Responsibility separation and
behavior preservation are the criteria.

## 11. Symmetric whole-production-package declaration-inventory proof

Before changing runtime code, implementation must capture a deterministic **whole-production-package
baseline inventory** for every production `.go` file in `backend-go/internal/httpapi`, excluding
`*_test.go`.

After decomposition, implementation must capture the **whole-production-package target inventory**
using the exact same production-file scope and the same inventory algorithm.

For each production declaration, the inventory must record enough deterministic identity to detect
semantic or structural drift, including at minimum:

- declaration kind;
- declaration name;
- receiver identity for methods;
- normalized signature/type identity;
- source file;
- deterministic declaration/body structural or source identity.

The exact representation may use Go AST/source hashing or an equivalent deterministic repository-local
mechanism, but baseline and target must use the same algorithm.

The production-package comparison must classify every difference as one of:

- declaration moved intact from `api.go` to an approved target file;
- explicitly approved new private routing/bootstrap helper;
- import-only/gofmt/file-boundary consequence that does not alter a declaration;
- material unexpected difference.

The following are material and fail closed unless explicitly reviewed:

- production declaration loss;
- duplicate production declaration;
- signature/type/receiver change;
- body/structural identity change not attributable to an approved mechanical relocation;
- unexpected new production helper or business logic;
- modification of a pre-existing companion-file declaration.

Separately, implementation must produce an **`api.go` declaration-to-target-file relocation map**
covering every baseline production declaration originally located in `api.go`.

Tests are not mixed into the production declaration inventory. If a test declaration inventory is
used, it must use one symmetric baseline/target `*_test.go` scope and be reported separately.

This symmetric whole-package proof is required specifically because the package already contains
production companion files and an `api.go`-only baseline cannot be compared rigorously with a
package-wide target.

## 12. Required regression coverage

At minimum, implementation verification must preserve/execute coverage for:

- app construction;
- health/readiness;
- exact route availability;
- CORS;
- auth register/login/refresh/logout;
- auth rate limiting;
- refresh cookies and CSRF transport behavior;
- asset search/detail;
- portfolio list/create/detail/summary;
- portfolio cursor behavior;
- transaction list/append;
- transaction cursor/filter behavior;
- import review/append;
- import-review token behavior;
- import limits;
- idempotency/replay behavior;
- error envelopes and request metadata;
- strict JSON admission;
- OpenAPI contract parity.

Existing tests are authoritative where they already cover these behaviors.

Test expectation changes require explicit review. A decomposition that passes only because existing
expected behavior was rewritten is not acceptable.

## 13. Required local verification

Future implementation must run at minimum, from the appropriate repository/backend directories:

```text
gofmt verification for every changed/new Go file
go test ./...
go vet ./...
git diff --check
```

Where supported by the repository environment, also run the applicable race/PostgreSQL-backed tests
and the same dependency/vulnerability policy used by CI.

Exact-head GitHub CI remains authoritative for all ten required protected-branch checks.

Every command outcome must be recorded exactly. A tooling failure or non-zero exit may not be
summarized as a successful process.

## 14. Required implementation evidence

The Stage 3.47 implementation/evidence dossier must record:

1. exact protected implementation base;
2. baseline `api.go` blob and size;
3. exact changed/new file list;
4. symmetric whole-production-package baseline/target declaration inventories;
5. complete `api.go` declaration-to-target-file relocation map;
6. exact route method/path preservation;
7. responsibility mapping from old `api.go` declarations to new files;
8. exported package-surface comparison;
9. exact review-size arithmetic for both canonical limits and any approved exception evidence;
10. completed Builder cross-class anti-regression preflight result;
11. confirmation that no dependency/OpenAPI/schema/frontend change occurred;
12. targeted regression evidence;
13. local test/vet/gofmt/diff-check outcomes;
14. exact published implementation head after publication;
15. exact-head GitHub CI result;
16. complete Internal review history after the External verdict is available for publication;
17. External published-head review history and remediation;
18. any tooling failure classified separately from project defects;
19. no unresolved material finding.

## 15. Review focus

Internal and External review must pay particular attention to semantic drift hidden inside a
supposedly mechanical move:

- altered route registration;
- changed middleware placement;
- changed init/default values;
- changed validation order;
- changed error mapping;
- changed cookie/header behavior;
- cursor/token helper edits;
- moved constants with changed values;
- new helper abstractions that subtly change control flow;
- tests edited to follow new behavior instead of detecting it;
- accidental modification of already-separated idempotency/replay code;
- unrelated P3-07/P3-08 work.

Large move-only diffs should be reviewed by responsibility and declaration identity, not waived as
"just refactoring."

Planning/publication review must also inspect the planning/implementation dossier itself for
review-state leakage: a document must not retain `pending review` or `candidate-only` current-state
wording that becomes false when the required review completes.

## 16. Mandatory Builder cross-class anti-regression preflight

Before **every** future P3-06 implementation package, pre-commit review package, publication commit,
Draft PR publication, remediation publication and closure package, the Builder must run one explicit
fail-closed preflight over the following recurring error classes.

This preflight is preventive. It is not a substitute for the designated reviewer.

### A. Complete review subject / scope identity

- complete candidate file list is known and packaged;
- exact protected base is known;
- exact diff/patch or equivalent complete subject is available;
- no unreviewed file is hidden by generated, untracked or companion-file state;
- review-size arithmetic covers the complete candidate.

### B. Symmetric structural comparison

- baseline and target declaration inventories use the exact same production-package scope;
- tests, if inventoried, use their own exact symmetric baseline/target scope;
- `api.go` relocation map is complete;
- no declaration is lost, duplicated, silently rewritten or introduced outside approved scope.

### C. Canonical review-size budget

- changed files `<=25`;
- changed hand-written business-logic lines `<=800`;
- if either is exceeded, stop before Internal review unless a documented exception has explicit
  Principal Architect approval before review begins;
- no silent multi-PR split is used to bypass the budget.

### D. Forensic chronology integrity

- every command/process exit is recorded exactly;
- a non-zero exit is never summarized as successful;
- substantive test success is distinguished from runner/tooling failure;
- failed tooling attempts remain tooling failures unless evidence shows a project defect;
- earlier findings, `REQUEST CHANGES`, `BLOCKED` states and remediations remain append-only.

### E. Evidence provenance / supportability

- canonical claims come only from self-contained package/repository/immutable evidence;
- operator transcriptions are labelled as transcriptions;
- session/tool mechanics are not promoted into permanent forensic facts without immutable evidence;
- unsupported claims such as client guard internals or connector-only mechanics are excluded.

### F. Authorization boundary

- review verdict is not mutation authorization;
- authorization language is not inferred or strengthened;
- before every remote write, the intended action family and every field must fit the current explicit
  human authorization;
- no-op write attempts still count as writes and are forbidden outside authorization.

### G. Lifecycle / publication stability and review-state leakage

Run an **exact-candidate semantic self-staleness simulation after all known reviewer findings have
been incorporated and immediately before handing those exact bytes to the next review or publication
step**.

This is not a keyword scan. The Builder must classify every statement that describes mutable
repository state, audit state, lifecycle state, authorization state, review state, implementation
shape or closure state and determine whether it remains true at every later transition. A statement
that is intentionally historical/baseline-specific must carry an explicit immutable temporal anchor.

Semantically inspect permanent dossier/PR/evidence wording across:

`local candidate -> review completed -> commit -> push -> Draft PR -> CI complete -> External verdict
-> evidence publication -> exact verification -> Ready -> merge -> closure`.

Search at minimum for lifecycle-sensitive wording such as:

`candidate only`, `requires`, `pending`, `pending review`, `pending re-review`, `before commit`,
`before push`, `will`, `next`, `current head`, `Draft PR`, `awaiting`, `to be published`,
`must be completed before`, `not yet`, `future`, `review required`, `resolution pending`.

Mandatory rules:

- no permanent artifact may encode a mutable reviewer-state field whose truth changes merely because
  the required review is performed;
- findings may be preserved historically as `found`, `remediated`, `resolved by review X`, or
  `overall verdict REQUEST CHANGES because of finding Y`, but not as a live `pending re-review`
  state intended to survive publication;
- if reviewer disposition has not yet occurred, permanent wording must state the invariant that
  disposition is established by designated review evidence bound to the exact candidate, rather than
  predicting a future pending state;
- completed gates must not remain described as future/pending;
- no sentence may falsely predict a future PR/head/CI/merge identity;
- the exact candidate must be re-scanned after every remediation because adding review history can
  itself introduce new stale lifecycle wording;
- mutable facts such as audit counts, remaining findings, file size/blob, route layout, package
  concentration, PR state or implementation state must never be labelled only as `current` in a
  permanent artifact when the artifact is expected to outlive that state;
- such facts must instead be anchored to an immutable event/base, for example
  `At Stage 3.46 planning base develop@...`, `Before P3-06 closure`, or
  `After protected P3-06 closure activation`;
- projected states must be explicitly conditional/hypothetical and must not be presented as already
  active;
- baseline descriptive sections must remain semantically true even after implementation and closure
  because their wording refers to the historical planning baseline, not the moving repository head.

Failure of this simulation blocks the package before designated review/publication.

### H. PR metadata / head / CI synchronization

- PR body facts match actual publication state;
- head/base identities are exact;
- CI is attributed to the exact head;
- metadata-only synchronization must not be confused with repository-byte change;
- stale “current head”, “pending CI” or “verification still required” wording is removed when the
  lifecycle advances.

### I. Runtime/dependency/contract isolation

For P3-06 specifically:

- no Go dependency drift;
- no executable OpenAPI drift;
- no schema/migration drift;
- no frontend drift;
- no P3-07/P3-08 work;
- no auth/session/idempotency/import semantic redesign;
- every semantic deviation from the mechanical decomposition fails closed.

### J. Canonical state / audit arithmetic synchronization

- `SOURCE_OF_TRUTH`, `ROADMAP`, stage implementation dossier and closure dossier agree when those
  surfaces are in scope;
- closed findings do not reappear;
- P3-06 remains OPEN through planning and implementation;
- the Stage 3.46 planning-base state is 29/32 = 90.625%; later audit arithmetic must be recomputed from protected state, and if no other finding changes, protected P3-06 closure yields 30/32 = 93.75%;
- only protected closure activation may produce 30/32 = 93.75%, leaving P3-07/P3-08.

### K. Role separation

- Builder prepares/fixes/packages/publishes only when authorized;
- designated review chat alone issues canonical review verdicts;
- Builder never self-issues `APPROVED`, `REQUEST CHANGES` or `BLOCKED`.

### L. Toolchain / artifact-generation safety

- repository-pinned toolchain and ambient toolchain are not conflated;
- machine-readable output is not parsed from blindly merged stdout/stderr;
- Markdown/evidence generation cannot execute backticks or command substitutions;
- Git worktree validity is checked with Git semantics rather than assuming `.git` is a directory;
- generated package manifest and exact candidate identities are verified before publication.

### M. Reviewer-result / permanent-artifact separation

- the permanent candidate never relies on `pending reviewer resolution` as its current-state truth;
- designated reviewer verdicts live in review evidence and may later be recorded only as completed
  historical facts;
- when a prior finding is already resolved in a completed review, the artifact must not continue to
  say the finding is pending;
- when a current finding awaits review, the artifact uses stable invariant wording instead of a
  self-staling pending field;
- after adding any review-history paragraph, rerun the lifecycle scan and transition simulation on
  the exact resulting bytes.

This specifically prevents the class exemplified by `PLAN-STAGE-03-46-P3-03`: a document can contain
a correct publication-stability policy and still violate it if the exact artifact is not simulated
through the next lifecycle transition.

### N. Temporal anchoring of mutable baseline facts

- every repository/audit/lifecycle fact is classified as one of:
  `immutable historical fact`, `planning-base fact`, `conditional projected fact`,
  `workflow invariant`, or `live metadata fact`;
- permanent planning artifacts may contain planning-base facts only when explicitly anchored to the
  exact immutable planning base or historical event;
- unqualified `current` wording is forbidden for facts expected to change during the planned
  lifecycle;
- audit counts/remaining findings are anchored to pre-closure or post-closure states;
- source-layout descriptions are anchored to the planning base when the implementation is intended to
  change that layout;
- live PR/head/CI state belongs in live metadata/evidence, not in permanent prose unless recorded as a
  completed historical event;
- the exact candidate must survive semantic simulation through the terminal closure state without any
  baseline statement becoming false.

This class is the permanent protection added after `PLAN-STAGE-03-46-P3-04`.

### O. Temporal-inventory / semantic-simulation coverage completeness

- every temporal/lifecycle-sensitive assertion in the exact candidate receives a stable assertion ID;
- every assertion ID must have exactly one semantic classification;
- every assertion classification must state its temporal anchor/category and T0..T11 truth behavior
  or an explicit reason the assertion is invariant;
- `inventory_count` must equal `classified_count`;
- `unclassified_count` must be exactly `0`;
- `duplicate_classification_count` must be exactly `0`;
- an overall semantic `PASS` is mechanically forbidden when either count rule fails;
- assertions may not be silently ignored merely because their wording does not contain a known
  temporal keyword;
- terminal-state checks must explicitly include every planning-base source-layout assertion whose
  subject is expected to change during implementation.

This class permanently addresses the v4 failure mode where an inventoried assertion was omitted from
semantic simulation.

### P. Semantic discovery completeness / source-to-ledger coverage

Keyword detection is never the authority for semantic completeness.

The exact candidate itself is the authority. Before review/publication:

- every physical line of the exact candidate is assigned exactly once to:
  - a named semantic assertion group; or
  - `STRUCTURAL/FORMATTING`; or
  - `NON_ASSERTION`;
- `candidate_line_count == source_coverage_line_count`;
- `uncovered_source_lines == 0`;
- `multiply_covered_source_lines == 0`;
- every named semantic assertion group appears in the semantic assertion catalog exactly once;
- every catalogued group has one correct temporal category and explicit T0→T11 behavior;
- compound state assertions must be complete as compounds, including:
  - all 16 frozen route method/path pairs;
  - planning-base `api.go` blob plus byte size;
  - planning-base closed count, percentage and remaining findings;
  - conditional post-closure closed count, percentage and remaining findings;
- no generic fallback classification such as `historical/normative; reviewer should verify` is
  sufficient for a state-sensitive assertion;
- an overall lifecycle `PASS` is mechanically forbidden unless both source-to-ledger coverage and
  assertion-to-classifier coverage pass.

This class permanently addresses the v5 failure where `inventory_count == classified_count` was true
inside a keyword-derived population while the exact candidate still contained semantic assertions
that never entered that population.

Any failed item in this preflight blocks progression to the next lifecycle mutation/review step until
remediated.

## 17. Rollback

Before protected merge, abandon the implementation branch/PR.

After protected implementation merge, rollback is a normal protected-branch revert of the
decomposition implementation. Because P3-06 is planned as structure-only with no schema/data/wire-format
change, no database migration or data rewrite should be required.

If implementation introduces a change that makes rollback require data/schema/protocol migration,
that is evidence of scope expansion and must fail closed.

## 18. Development-path governance

After this plan is approved and separately published/merged, implementation remains on the canonical
development path defined by `docs/REVIEW_WORKFLOW.md` v1.3.0:

1. implementation branch from the exact then-current protected `develop`;
2. scoped same-package decomposition;
3. local gates and deterministic structural evidence;
4. complete Internal line-by-line read-only review in the designated review chat;
5. Builder remediation and rerun of affected gates;
6. explicit human commit/push authorization;
7. Draft PR to `develop`;
8. exact-head required GitHub CI;
9. fresh External published-head review in the same designated review chat;
10. evidence-only Internal evidence publication after the External verdict;
11. CI on the evidence head;
12. same-chat exactness/no-semantic-drift verification;
13. separate explicit human Ready + squash-merge authorization;
14. separate docs-only closure-governance activation.

Internal evidence remains withheld from the Draft PR/repository evidence surface until the External
verdict.

No review verdict by itself authorizes a protected mutation.

## 19. P3-06 lifecycle and audit arithmetic

P3-06 remains OPEN during Stage 3.46 planning and throughout the later implementation stage.

P3-06 becomes CLOSED only after:

- an approved behavior-preserving decomposition implementation is actually squash-merged into
  protected `develop`; and
- a separate reviewed documentation/governance closure activation is present on protected `develop`.

Stage 3.46 planning-base original audit state:

- closed: 29 / 32;
- completion: 90.625%;
- remaining: P3-06, P3-07, P3-08.

If no other audit finding changes concurrently, after eventual P3-06 closure:

- closed: 30 / 32;
- completion: 93.75%;
- remaining: P3-07, P3-08.

No future PR number, CI run, implementation head or merge SHA is predicted by this plan.

## 20. Explicitly out of scope

Stage 3.46 and the later P3-06 implementation do not authorize:

- P3-07 — transaction-form fixture/default semantics;
- P3-08 — migration-validator policy hardening;
- Stage 3.25 privacy Security Review evidence work;
- new product functionality;
- new routes;
- route redesign;
- HTTP contract redesign;
- auth/session redesign;
- idempotency/replay redesign;
- import-format redesign;
- cursor/token version changes;
- financial logic;
- database/schema/migrations;
- frontend work;
- dependency maintenance;
- architecture changes.

## 21. Planning decision

The proposed remediation is intentionally narrow:

**split the Stage 3.46 planning-base 67,204-byte `httpapi/api.go` into cohesive same-package
transport files while preserving behavior and proving that preservation through declaration
inventory, route invariants, existing regressions and exact-head CI.**

This plan does not authorize runtime implementation, commit, push, PR creation, Ready, merge or
closure.
