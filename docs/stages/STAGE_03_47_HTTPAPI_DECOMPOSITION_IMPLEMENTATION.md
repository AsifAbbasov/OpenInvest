# Stage 3.47 — P3-06 HTTP API Decomposition Implementation

| Field | Value |
| --- | --- |
| Stage | 3.47 |
| Status | IMPLEMENTATION MERGED — PR #107 squash-merged into protected `develop` at `332f7cd2ec40caf0760b97b806f637e4c89dbb96`; P3-06 closure is merge-activated under Stage 3.48 |
| Finding | Original audit `P3-06 — httpapi/api.go decomposition` |
| Development path | Yes — runtime-source structural refactor |
| Canonical implementation base | `develop@546f0406d1353c13673be4ab97c4a527a9b58116` |
| Base tree | `ec46d70104e9091ebd94084920b8523de5f762c5` |
| Approved planning artifact | `docs/stages/STAGE_03_46_HTTPAPI_DECOMPOSITION_PLAN.md` blob `9e028f817220973458b28a2393ee61bdd2eb83a0` |
| Planning-base audit state | 29 / 32 closed = 90.625%; remaining P3-06, P3-07, P3-08 |
| Final evidence-publication head | `657afbde74b79db6966333e27d52f0320660d6b3` |
| Implementation squash merge | `332f7cd2ec40caf0760b97b806f637e4c89dbb96` |
| Final exact-head CI | CI #301 / run `33343890109`, 10/10 required jobs successful |
| P3-06 lifecycle | Before the approved Stage 3.48 closure record and synchronized canonical surfaces are present on protected `develop`, P3-06 is OPEN; once present, P3-06 is CLOSED |

## 1. Problem

At the canonical Stage 3.47 implementation base, `backend-go/internal/httpapi/api.go` is the
Stage 3.46 planning-base 67,204-byte concentration point for multiple HTTP transport responsibilities.
The original audit finding is maintainability/structure debt: bootstrap, route registration, CORS,
auth rate limiting, health/readiness, auth transport, assets, portfolios, transactions, imports,
security/token helpers, pagination and shared response/DTO helpers are concentrated in one source file.

This implementation does not characterize the baseline as a demonstrated user-visible defect or
cost, accidental-coupling risk and the blast radius of later HTTP changes.

## 2. Root cause

The first vertical-slice implementation accumulated transport responsibilities in one package file as
features were added. Later focused companion files already extracted idempotency/replay and password
JSON concerns, but the remaining `api.go` concentration was not decomposed.

## 3. Failure / attack scenarios considered

The remediation is designed to reduce structural risk without changing behavior. Relevant failure
classes include:

- route loss, duplication or method/path drift during a move;
- CORS middleware moving after route registration or changing policy;
- authentication/session/cookie/CSRF behavior changing through cleanup disguised as refactor;
- rate-limit state or timing semantics changing;
- cursor/token signing, validation, TTL/version or size-limit drift;
- idempotency/replay authority drifting into newly moved handlers;
- import review/append safety or parser/version semantics changing;
- DTO JSON/status/error-envelope drift;
- accidental edits to pre-existing companion production files;
- declaration loss, duplication, receiver/signature changes or unexpected helper growth.

## 4. Impact boundary

Allowed impact is file ownership and one bounded private routing delegation only. No product behavior,
public API contract, database/schema/migration, dependency, frontend, financial calculation or
external-data behavior is intentionally changed.

P3-07 transaction-form fixture/default semantics and P3-08 migration-validator work are explicitly
out of scope.

## 5. Initial / proposed remediation

The canonical Stage 3.46 plan approved a same-package mechanical decomposition. Stage 3.47 implements
that decision by moving intact production declarations into cohesive `httpapi` files and reducing
`api.go` to API state/bootstrap responsibility.

One approved structural rewrite is used: `newApp` keeps Fiber app construction and CORS placement in
`api.go`, then delegates the exact route registrations to a new private `registerRoutes` helper.

## 6. Design rationale

Same-package movement avoids new boundaries, dependencies, interfaces and runtime indirection.
Responsibility files make review ownership explicit while preserving package-private access and
existing handler/helper names.

The candidate intentionally does not redesign transport abstractions. Shared helpers remain concrete
HTTP helpers rather than becoming a generic framework.

## 7. Implemented responsibility layout

- `api.go` — `API` state, `New`, `NewDevelopment`, `nowUTC`, minimal `newApp` bootstrap;
- `routes.go` — exact canonical route registration through private `registerRoutes`;
- `cors.go` — local-development CORS admission;
- `auth_rate_limit.go` — authentication rate-limiter state and helpers;
- `health.go` — health/readiness handlers;
- `auth_handlers.go` — auth transport, development-subject/cookie helpers and auth DTO mapping;
- `assets.go` — asset transport and asset cursor coordination;
- `portfolios.go` — portfolio transport, portfolio cursor coordination and portfolio DTO mapping;
- `transactions.go` — transaction transport, transaction cursor coordination and transaction DTO mapping;
- `imports.go` — import review/append transport and import DTO mapping;
- `import_security.go` — import-review token/hash/row-limit security helpers;
- `pagination.go` — shared signed pagination cursor payload/sign/verify helpers;
- `money_transport.go` — shared transport money/decimal conversion and mapping;
- `transport_helpers.go` — concrete shared HTTP query/response/error/metadata/strict-JSON helpers;
- `transport_constants.go` — the pre-existing grouped transport constants moved intact as one block.

Pre-existing companion production files are not rewritten by the decomposition.

## 8. Behavior-preservation contract

The exact 16 canonical method/path registrations remain unchanged and in the same order in
`registerRoutes`. `newApp` still constructs the same Fiber app, installs `localDevelopmentCORS` before
route registration and returns the same app.

All moved declarations other than the approved `newApp` structural rewrite are required to retain the
same normalized declaration/body identity. The only approved new production declaration is the
private `registerRoutes` helper containing the moved route registration statements.

The whole-production-package comparator fails closed on declaration loss, unexpected new declarations,
unapproved body/signature/receiver drift or changes to pre-existing companion declarations.

## 9. Review-size budget

Canonical default limits are `<=25` changed files and `<=800` changed lines of hand-written business
logic. Mechanical relocation is reported separately from semantic business-logic change.

The candidate uses relocation-aware declaration comparison. Intact moved declarations are not counted
as changed business logic. The `newApp` delegation and `registerRoutes` extraction are structural
routing edits and do not introduce or alter business decision logic. Raw textual diff additions and
deletions are still reported separately and are not hidden from review.


designated Internal verdict/findings were intentionally withheld from the Draft PR and repository
evidence surface. The External verdict now exists on the exact published implementation head, so this
evidence-only follow-up publishes the completed Internal record in Section 14 without changing runtime
source or using the Internal verdict as support for the already-completed External conclusion.

failures unless an event is explicitly identified as a project/runtime defect:

Canonical record: PR #107; commit(s) `5ea1f7f29cf0ab9225460e01076255f32e2cf4cf`, `edfa751ea7daef8ecb44defcb8eb84f04156e2d3`.

No tooling/evidence failure above is reclassified as a project/runtime defect, and no project/runtime
External verdict existed; Section 14 now publishes that completed record as required by the canonical
workflow.

## 11. Rejected approaches

The following approaches are rejected by the approved plan and were not used:

- new HTTP framework/router abstraction;
- new package or Go module;
- service/interface redesign;
- behavior cleanup mixed into decomposition;
- changes to replay/idempotency companion logic;
- silent P3-07 or P3-08 work;
- silent multi-PR split to evade review-size limits.

## 12. Regression strategy

The candidate uses three complementary proof layers:

1. symmetric whole-production-package declaration inventory before and after decomposition;
2. exact route registration verification plus focused responsibility-shape verification;
3. repository local gates including pinned-Go tests, race tests where runnable, vet, OpenAPI validation,
   gofmt verification and `git diff --check`.

Existing tests remain authoritative; no expectation is changed merely to make the decomposition pass.

## 13. Evidence provenance

Machine evidence in the review package is generated from the isolated worktree rooted at the exact
Stage 3.47 base. Tool output and exit status are recorded as machine evidence. Repository facts are
bound to exact Git identities or generated inventory artifacts.

No unsupported connector/tooling mechanics are elevated into canonical repository facts.

## 15. Published implementation identity and exact-head CI

After separate explicit merge gate, the Internal-approved candidate was published as:

- Draft PR: `#107`;
- feature branch: `refactor/stage-03-47-p3-06-httpapi-decomposition`;
- implementation commit: `edfa751ea7daef8ecb44defcb8eb84f04156e2d3`;
- implementation tree: `02e113e8ef00ea8c3e6867af7cbc31771ecbfd4b`;
- commit parent/base: `546f0406d1353c13673be4ab97c4a527a9b58116`;
- changed files: `16`;
- published diff: `2369 additions`, `1921 deletions`.

Canonical record: PR #107.

GitHub Actions CI run `33341679669` / run number `300`, triggered by the pull request for exact head
`edfa751ea7daef8ecb44defcb8eb84f04156e2d3`, completed with overall conclusion `success`.
All 10 protected required jobs succeeded:

1. Go tests;
2. Python tests;
3. Frontend build and typecheck;
4. OpenAPI contract;
5. Docker Compose config;
6. PostgreSQL migration validation;
7. Go vet;
8. Go race tests;
9. Go vulnerability scan;
10. Dependency security scan.

The successful exact-head Go race and PostgreSQL-backed jobs are authoritative for that published
implementation head. The earlier local race baseline limitation and local Docker skip remain historical
prepublication evidence and are not rewritten as successful local executions.

This evidence-only follow-up intentionally changes only this dossier. Its own publication commit/head
and CI identity are not predicted here; they must be recorded only after they actually exist.

## 17. Closure activation

Stage 3.47 implementation is now squash-merged through PR #107 into protected `develop` at
`332f7cd2ec40caf0760b97b806f637e4c89dbb96`.

That implementation merge does not itself close P3-06. P3-06 remains OPEN until the separate Stage 3.48
verified and squash-merged to protected `develop`.

If no other original audit finding changes concurrently, Stage 3.48 protected activation moves the
original audit from 29/32 = 90.625% to 30/32 = 93.75%, leaving P3-07 and P3-08.

## 18. Residual limitations

This stage does not remove the `httpapi` package's inherent transport coupling and does not attempt a
new architecture. It only reduces one-file concentration under the frozen same-package design.
Exact-head GitHub CI remains authoritative after publication; local evidence cannot substitute for it.

## 19. Local evidence summary

The preparation runner appends observed local evidence below. A non-zero required gate blocks package
creation and is not rewritten as success.

### 19.1 Exact prepublication identity

- exact base: `546f0406d1353c13673be4ab97c4a527a9b58116`
- exact base tree: `ec46d70104e9091ebd94084920b8523de5f762c5`
- baseline `api.go` blob: `f05b7288311918e95c7484dc2cc7385f2de3c03c` / `67204` bytes
- approved planning blob: `9e028f817220973458b28a2393ee61bdd2eb83a0`
- pinned toolchain observation: `go version go1.25.14 darwin/arm64`
- prepublication preparation runner remote mutations: `NONE` (read-only `fetch origin develop` only)

### 19.2 Structural preservation evidence

- symmetric whole-production-package inventory: `PASS`
- pre-existing companion declaration/blob drift: `NONE`
- exact `api.go` relocation map: `EVIDENCE/api_go_relocation_map.tsv`
- exact 16-route `registerRoutes` verification: `PASS`
- responsibility-shape verification: `PASS`
- approved structural differences only: `newApp` routing delegation + new private `registerRoutes`

### 19.3 Local quality gates

- gofmt stability: `PASS`
- `git diff --check`: `PASS`
- pinned `go test ./...`: `PASS`
- pinned `go vet ./...`: `PASS`
- pinned `go test -race -count=1 ./...`: `BASELINE_PARITY_LIMITATION`; parity mode=`HISTORICAL_V8_ANOMALY_FORCED_RECURRENCE_CHARACTERIZED`; baseline rc=`1`, candidate rc=`1`; non-zero baseline-parity exits are preserved and not relabeled success
- OpenAPI validator: `PASS`
- PostgreSQL-backed local gate: `SKIPPED — Docker unavailable or daemon not running`
- `go.mod` blob remains `a6be6f2266a428b52206ee3890c7d5b199dab97b`; `go.sum` blob remains `bdd8e6edfd45a2725fb1f8dc0831d45ae9f39cd0`

### 19.4 Canonical review budget

- changed files: `16 / 25`
- changed hand-written business-logic lines: `0 / 800`
- raw textual Go relocation diff for transparency: `2050 additions`, `1921 deletions`
- moved declaration bodies are separately proven intact by structural identity and are not hidden from line-by-line review

### 19.5 Published implementation-head verification

- Draft PR: `#107`;
- implementation head: `edfa751ea7daef8ecb44defcb8eb84f04156e2d3`;
- implementation tree: `02e113e8ef00ea8c3e6867af7cbc31771ecbfd4b`;
- GitHub CI run: `33341679669` / run number `300`;
- exact-head required CI: `10 / 10 SUCCESS`;
- External P0/P1/P2/P3: `0 / 0 / 0 / 0`;
- P3-06 lifecycle after External approval: `OPEN`.

## 20. Governance / closure activation

Stage 3.47 development-path work is complete and merged.

Immutable post-evidence history:

1. final evidence-publication head `657afbde74b79db6966333e27d52f0320660d6b3`;
2. final Stage 3.47 dossier blob `df81494d72490ea03a8c3ab71645649a9645b6d3`;
3. evidence-head CI #301 / run `33343890109` — 10/10 SUCCESS;
4. single-context evidence-publication verification returned `APPROVED`, confirmed `1 DOC ONLY`,
   runtime freeze `15/15 MATCH`, complete Internal chronology, accurate External verdict,
   External independence, publication stability and zero findings;
5. explicit human Ready authorization;
6. connected Ready mutation failed before state change due a GraphQL schema error;
Canonical record: PR #107.
8. the already-authorized CLI fallback transitioned PR #107 to Ready;
9. post-Ready read-back confirmed same head, mergeability and CI #301 still 10/10;
10. separate explicit merge gate;
11. actual squash merge `332f7cd2ec40caf0760b97b806f637e4c89dbb96`;
12. protected `develop` read-back at that exact SHA and tree `6d073ad2530b2f6c9f59c973fbbe5ee284b16692`.

The implementation merge is immutable; moving `develop` is not hard-coded as permanently equal to it.

P3-06 closure remains separate. Stage 3.48 is docs/governance-only merge-activated closure:
before its approved surfaces reach protected `develop`, P3-06 remains OPEN; once present, P3-06 is CLOSED.

Tooling anti-regression controls are preserved:
- machine-verified exact local bytes outrank manual connector reconstruction for an approved candidate;
- a failed connector mutation requires live read-back before any already-authorized fallback.

No statement here authorizes Stage 3.48 commit, push, Draft PR, Ready, merge or protected mutation.
