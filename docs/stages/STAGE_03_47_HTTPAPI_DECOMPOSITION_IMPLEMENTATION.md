# Stage 3.47 — P3-06 HTTP API Decomposition Implementation

| Field | Value |
| --- | --- |
| Stage | 3.47 |
| Finding | Original audit `P3-06 — httpapi/api.go decomposition` |
| Development path | Yes — runtime-source structural refactor |
| Canonical implementation base | `develop@546f0406d1353c13673be4ab97c4a527a9b58116` |
| Base tree | `ec46d70104e9091ebd94084920b8523de5f762c5` |
| Approved planning artifact | `docs/stages/STAGE_03_46_HTTPAPI_DECOMPOSITION_PLAN.md` blob `9e028f817220973458b28a2393ee61bdd2eb83a0` |
| Planning-base audit state | 29 / 32 closed = 90.625%; remaining P3-06, P3-07, P3-08 |
| P3-06 lifecycle | OPEN through implementation; only a later separately reviewed docs-only closure activation may close it |

## 1. Problem

At the canonical Stage 3.47 implementation base, `backend-go/internal/httpapi/api.go` is the
Stage 3.46 planning-base 67,204-byte concentration point for multiple HTTP transport responsibilities.
The original audit finding is maintainability/structure debt: bootstrap, route registration, CORS,
auth rate limiting, health/readiness, auth transport, assets, portfolios, transactions, imports,
security/token helpers, pagination and shared response/DTO helpers are concentrated in one source file.

This implementation does not characterize the baseline as a demonstrated user-visible defect or
security exploit. The risk is structural: excessive responsibility concentration increases review
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

## 10. Review and Builder preparation history

Stage 3.46 planning review history is preserved in the canonical planning artifact. The Builder does
not issue an Internal reviewer verdict. The permanent Stage 3.47 dossier preserves Builder preparation
history but does not publish the current designated Internal verdict or finding details before the
External published-head verdict, because canonical workflow keeps that review evidence withheld from
the Draft PR/repository evidence surface until then.

The following Stage 3.47 Builder preparation/tooling chronology is append-only because each event
materially explains why later evidence-generation controls exist. These are Builder/tooling or evidence
failures unless an event is explicitly identified as a project/runtime defect:

1. The initial local full `go test -race -count=1 ./...` gate returned a non-zero exit. That non-zero
   result remains a failed execution event and is never rewritten as `PASS`.
2. Preparation v1 removed its temporary worktree before preserving the race stdout/stderr needed to
   classify that failure. The run was therefore diagnostically insufficient. This was an evidence-
   generation defect, not proof of a runtime decomposition defect.
3. Preparation v2 was intended to add baseline-versus-candidate diagnostics, but the Builder packaging
   step did not persist the modified runner bytes before ZIP creation. The delivered archive therefore
   still contained the older v1 runner. This was a Builder packaging defect.
4. Preparation v3 corrected archive-byte persistence, emitted an explicit runner-version marker and
   preserved a diagnostic archive before cleanup. It then executed the same full-race command against
   the pristine exact protected baseline and the decomposed candidate on the same machine.
5. The v3 baseline and candidate both returned non-zero full-race exits with byte-equal normalized
   non-empty failure signatures. The observed class was the same authentication-test
   `POST /api/v1/auth/register: i/o timeout`, and neither side emitted `WARNING: DATA RACE`. This is
   `BASELINE_PARITY_LIMITATION`, not a successful race run and not a demonstrated candidate regression.
6. Builder analysis after v3 identified a separate scope-accounting weakness: generated Go files were
   still untracked when a plain Git diff path inventory was collected, so a review package could omit
   candidate files even though the runtime transformation itself was present in the worktree. This was
   an evidence-scope defect.
7. Preparation v4 remediated complete-diff accounting by registering generated files with local
   intent-to-add inside the isolated worktree and independently requiring the Git diff path inventory
   to equal the porcelain-status path inventory. The resulting review subject contained all 16 changed
   candidate files.
8. Preparation v6 strengthened the physical-line/assertion/lifecycle evidence but its first Mac execution
   still stopped fail-closed at `ERROR[70]` with `baseline_rc=1` and `candidate_rc=1` because the local
   race-control comparator could not establish parity. Code inspection showed that the comparator compared
   sorted matched failure lines with multiplicity and therefore could false-block when repeated or volatile
   instances of the same semantic timeout class differed between runs. The preserved v6 non-zero event is a
   Builder tooling/evidence-control limitation, not a successful race run and not proof of a project/runtime defect.
9. Preparation v7 replaces that fragile line-multiplicity comparison with normalized unique semantic failure
   identities derived from both stdout and stderr. If the first semantic sets differ, v7 preserves the mismatch
   and runs one fresh pristine-baseline/candidate confirmation pair. Continuation is allowed only when the
   two-pair baseline and candidate semantic unions are exactly equal, both candidate-only and baseline-only
   sets are empty, every observed race-warning flag is zero, and all non-zero exits remain recorded as failures.
   Otherwise v7 stops fail-closed and preserves diagnostics. This does not retroactively relabel v6 as PASS.
10. Preparation v8 separates physical-source coverage from semantic-proposition discovery. Wrapped
    Markdown paragraphs/list items/table rows are assembled into complete logical assertion blocks before
    classification, each substantive dossier line maps to exactly one semantic assertion, and temporal
    qualifiers such as `until`, `after` and `if` therefore govern the complete proposition rather than one
    physical line. Each assertion carries an explicit temporal rule and phase contract; transition rules
    produce different expected states on opposite sides of T9/T10 or T15 instead of mechanically repeating
    one classification-level PASS across all 16 phases.
11. Preparation v8 also replaces the v7 selected-pattern race signature with structured `go test -json`
    evidence. For every failed package, the complete structured event surface is retained (`run`, `pass`, `skip`, `pause`, `cont`, `fail`, and `output`; only the package `start` envelope is excluded), while timing-only
    fields are narrowly normalized, stderr is retained, and the canonical comparison includes failure-detail
    text rather than only failed-test/package identities. Adversarial self-tests require timing-only variance
    to compare equal while different `expected/got`, candidate-only detail, candidate-only panic and an extra
    failure event compare unequal. Raw JSON/stderr and exact non-zero exits remain preserved.
12. The first local v8 full-suite baseline/candidate pair still differed. v8 preserved the mismatch and
    ran one confirmation pair. Both baseline full-suite runs produced the exact same structured event hash,
    while the candidate retry reproduced that stable baseline hash exactly; the first candidate run remained
    different and included three `missing cookie "oi_refresh"` details plus one auth rate-limit PASS. v8 then
    stopped fail-closed at `ERROR[70]` because its two-run union policy permanently retained that one-off
    candidate outcome. The supplied diagnostic archive SHA256 was
    `ddc01dfe3eb8e6878194acd1f9ad279866d7a6571fbb5e2b061f38aa42b6a30f`. This is a Builder evidence-control
    event; it does not by itself prove a runtime decomposition defect.
13. Preparation v9 preserves that first anomaly instead of erasing it and replaces union-equality acceptance with
    controlled recurrence characterization. Continuation requires: an exact stable two-run pristine baseline
    full-suite profile; TWO subsequent candidate full-suite runs matching that stable profile exactly; symmetric
    `go test -json -race -count=3 ./internal/httpapi` characterization on pristine baseline and candidate; and zero
    data-race warnings in every observed run. Any failure remains `ERROR[70]`. A successful characterization is
    still `BASELINE_PARITY_LIMITATION`, never a successful race run, and exact-head CI remains authoritative.
14. The designated v9 Internal re-review resolved P3-01 and P3-02 but kept P3-03 open. It demonstrated that
    v9 still normalized duration-looking substrings inside arbitrary assertion/error output, so semantically
    different values such as `got (2.00s)` versus `got (3.00s)` could canonicalize equal. The same review also
    showed an executable-policy bypass: because the current v9 baseline/candidate pair compared equal, the direct
    branch accepted `DIRECT_STRUCTURED_FAILURE_EVENT_EQUAL` and skipped the recurrence characterization that the
    preserved v8 anomaly made mandatory. This was an evidence-control defect, not evidence of runtime drift.
15. Preparation v10 narrows normalization to syntactically unambiguous Go harness/runtime metadata only. Arbitrary
    assertion/error payload is preserved exactly, and adversarial tests explicitly require semantic duration-like
    values and semantic goroutine-number values to compare unequal. The pinned v8 diagnostic is now an executable
    control input: its verified non-empty candidate-only anomaly activates recurrence characterization before any
    race-control success path. v10 therefore always requires a stable two-run pristine baseline, TWO subsequent
    candidate full-suite confirmations matching that stable profile, symmetric pinned
    `go test -json -race -count=3 ./internal/httpapi` characterization, and zero data-race warnings. A later clean
    first pair cannot bypass that sequence.

No tooling/evidence failure above is reclassified as a project/runtime defect, and no project/runtime
defect may be silently reclassified as tooling. Current designated Internal findings and verdicts stay
in review-package evidence until the workflow permits their publication.

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

## 14. Internal Review Evidence publication rule

The Builder does not issue an Internal verdict. For the initial Draft PR publication state, canonical
workflow requires the stage-report value `WITHHELD — external published-head phase pending` and keeps
the actual Internal verdict/findings out of the repository/PR evidence surface until the External
published-head verdict. This paragraph states that workflow rule rather than claiming a future verdict.

After an actual External verdict, the required evidence-only follow-up may append the completed
Internal history and must receive exact-head CI plus same-chat no-semantic-drift verification.

## 15. Published exact-head identity

No implementation PR number, publication commit, CI run, External verdict or merge identifier is
asserted by this prepublication implementation record. If publication is later separately authorized,
only observed identifiers may be recorded.

## 16. External Review Evidence

No External verdict is asserted in the prepublication candidate. The External phase, if reached after
separate publication authorization and green exact-head CI, is a fresh designated-review phase and may
not use the Internal verdict/findings as supporting evidence.

## 17. Closure activation

Stage 3.47 implementation does not close P3-06. Even after an eventual protected implementation merge,
P3-06 remains OPEN until a separate docs-only closure record and synchronized canonical audit surfaces
are independently reviewed, published, verified and squash-merged to protected `develop`.

If no other original audit finding changes concurrently, that later closure activation would move the
original audit from the Stage 3.47 planning-base state 29/32 = 90.625% to 30/32 = 93.75%, leaving
P3-07 and P3-08.

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
