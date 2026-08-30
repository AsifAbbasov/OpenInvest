# Stage 3.43 — P3-10 Fiber Maintenance Plan

| Field | Value |
| --- | --- |
| Status | Planning/review candidate only; dependency/runtime implementation not authorized |
| Date | 2026-08-30 |
| Canonical planning base | `develop@8861d49580c92eabe5f859729b3777175134a4e2` |
| Finding | Original audit `P3-10 — Fiber maintenance` |
| Prior closure | P3-09 CLOSED by merge-activated Stage 3.42 governance after PR #102 squash merge placed the approved closure surfaces on protected `develop` |
| Current audit state | 28 / 32 closed = 87.5%; remaining P3-06, P3-07, P3-08, P3-10 |
| Runtime/dependency implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / Ready / merge authorized here | No |

## 1. Objective

Close original audit finding P3-10 by updating the repository's direct Go Fiber v3 dependency from
the current `v3.3.0` baseline to the current stable upstream `v3.5.0` release while preserving
OpenInvest HTTP/API behavior and avoiding unrelated backend, architecture, financial, database,
OpenAPI, frontend, or audit-remediation work.

This planning artifact does not itself modify dependencies and does not close P3-10.

## 2. Exact repository baseline

At protected `develop@8861d49580c92eabe5f859729b3777175134a4e2`:

- `backend-go/go.mod` blob:
  `9342530d300392173c9ebeca0ce42a8ec9d86459`;
- `backend-go/go.sum` blob:
  `1a206517eae6e65a91fabbfa8c4ba440cf8b526a`;
- direct Fiber dependency:
  `github.com/gofiber/fiber/v3 v3.3.0`;
- Go directive:
  `go 1.25.14`;
- other root `go.mod` direct requirements:
  - `github.com/google/uuid v1.6.0`;
  - `github.com/jackc/pgx/v5 v5.9.2`;
  - `golang.org/x/crypto v0.51.0`;
  - `gopkg.in/yaml.v3 v3.0.1`;
- OpenInvest directly imports `golang.org/x/crypto/argon2` in authentication, so `x/crypto` is not
  merely an incidental transitive module; it is a security-sensitive shared-direct dependency;
- `backend-go/internal/httpapi/api.go` blob:
  `f05b7288311918e95c7484dc2cc7385f2de3c03c`.

The HTTP API imports only the Fiber v3 root package. Repository inspection did not find imports of
Fiber's built-in `cache`, `helmet`, `basicauth`, `proxy`, or `idempotency` middleware packages.

The current API constructs the app with:

`fiber.New(fiber.Config{AppName: "OpenInvest API"})`

and uses core Fiber route registration, `fiber.Ctx`, request body/header/cookie/query/path access,
response status/header/cookie functions, and `c.Context()` propagation into OpenInvest services.

## 3. Upstream maintenance basis

The current upstream latest non-prerelease Fiber release was independently read from the upstream
GitHub repository as:

- tag: `v3.5.0`;
- prerelease: false;
- published: 2026-08-13.

Upstream `v3.4.0`, published 2026-07-02, contains a substantial set of core, middleware, routing,
binding, concurrency, proxy, timeout, session, and security/correctness fixes.

Upstream `v3.5.0` adds further fixes and maintenance, including routing, binding, redirect, proxy,
cache, listener/error-path, HTTP semantics, and dependency updates.

This stage does not assert that every upstream fixed issue is reachable in OpenInvest.

### Security applicability note

One confirmed upstream advisory relevant to version identity is:

- `GHSA-gv83-gqw6-9j2c` — Fiber helmet HSTS header not set;
- affected Fiber v3 versions: `<= 3.3.0`;
- patched version: `3.4.0`.

OpenInvest currently pins `3.3.0`, so version-range applicability exists.

However, repository inspection did not find use of Fiber's `helmet` middleware, therefore this plan
does **not** claim a demonstrated OpenInvest HSTS exploit path from that advisory.

Other historical Fiber v3 advisories with patched ranges below `3.3.0` must not be presented as
current OpenInvest vulnerabilities merely because they exist in upstream history.

Any newly discovered OpenInvest-reachable security exposure during implementation/review must be
reported as a separate material finding instead of silently rewriting P3-10.

## 4. Known shared-direct module graph movement

The Fiber `v3.5.0` target is already known at planning time to require:

`golang.org/x/crypto v0.54.0`

The current OpenInvest root module directly declares:

`golang.org/x/crypto v0.51.0`

and OpenInvest directly uses:

`golang.org/x/crypto/argon2`

for password hashing and verification.

Therefore the Fiber maintenance scope is **not accurately described as a Fiber-only selected-module
change**. The approved target is a Fiber-attributable module-graph update that includes the already
known security-sensitive shared-direct movement:

`golang.org/x/crypto: v0.51.0 -> selected v0.54.0`

The implementation must distinguish:

1. the version text declared in root `backend-go/go.mod`;
2. the effective version selected by Go Minimal Version Selection (MVS);
3. the version actually used by the compiled module graph.

If normal Go module resolution updates the root `go.mod` requirement to `v0.54.0`, that declaration
change is allowed only as the already-known Fiber-attributable shared-direct movement described here.

If the root declaration remains `v0.51.0` but MVS selects `v0.54.0`, the implementation evidence must
state both facts explicitly rather than describing the older declaration as the effective runtime
version.

The following unrelated root direct requirements remain frozen exactly:

- `github.com/google/uuid v1.6.0`;
- `github.com/jackc/pgx/v5 v5.9.2`;
- `gopkg.in/yaml.v3 v3.0.1`.

Any other direct-dependency movement requires explicit attribution and review. An unrelated movement
must fail closed.

## 5. Severity and scope discipline

The original audit severity remains **P3**.

Upstream maintenance/security information can raise operational priority without retroactively
rewriting the original audit severity.

This stage SHALL NOT absorb:

- P3-06 — `httpapi/api.go` decomposition;
- P3-07 — transaction-form fixture/default semantics;
- P3-08 — migration-validator policy hardening;
- Stage 3.25 privacy work;
- HTTP/API redesign;
- auth/session redesign;
- financial calculations;
- database/schema/migrations;
- OpenAPI contract changes;
- Next.js/frontend changes.

## 6. Frozen implementation target

The proposed implementation target SHALL be:

`github.com/gofiber/fiber/v3: v3.3.0 -> v3.5.0`

Rules:

1. Use stable `v3.5.0`, not `main`, a pseudo-version, prerelease, or canary-like build.
2. Keep Go directive at `1.25.14`.
3. Freeze the already-known Fiber-attributable shared-direct selected-module movement:
   `golang.org/x/crypto v0.51.0 -> selected v0.54.0`.
4. Keep `github.com/google/uuid v1.6.0`, `github.com/jackc/pgx/v5 v5.9.2`, and
   `gopkg.in/yaml.v3 v3.0.1` unchanged.
5. Distinguish root `go.mod` declaration from MVS-selected effective module version.
6. Allow only additional module changes demonstrably produced by the exact Fiber update and normal Go
   module resolution.
7. Enumerate every changed direct/shared-direct/transitive module in implementation evidence.
8. No application-source change is expected.
9. If the Fiber update requires a material application-source/API behavior change, stop and return to
   planning/review before expanding scope.
10. Do not combine P3-10 with P3-06 decomposition even if `api.go` is touched by compilation concerns.

## 7. Expected implementation surface

Expected mandatory files:

- `backend-go/go.mod`;
- `backend-go/go.sum`;
- one Stage 3.44 implementation/evidence dossier under `docs/stages/`.

Application source changes are not expected.

A source change may occur only if all of the following are true:

- directly required by Fiber `v3.5.0` compatibility;
- minimal;
- behavior-preserving;
- explicitly reviewed;
- covered by targeted regression tests;
- does not perform P3-06 decomposition.

## 8. Go module invariants

Implementation must prove:

- direct Fiber dependency resolves exactly to `v3.5.0`;
- no active `github.com/gofiber/fiber/v3 v3.3.0` entry remains;
- `golang.org/x/crypto` effective selected version is exactly the Fiber-attributable target required by
  the resolved graph, expected `v0.54.0` for the approved Fiber `v3.5.0` target;
- root `go.mod` declaration and MVS-selected effective `x/crypto` version are reported separately;
- `go list -m all` or equivalent exact resolved-module evidence is captured;
- implementation evidence identifies why `x/crypto` moved and confirms OpenInvest's direct Argon2 use;
- `go.mod` and `go.sum` are reproducible through the pinned Go toolchain;
- unrelated root direct dependency versions remain unchanged;
- every other direct/shared-direct/transitive change is enumerated and attributable;
- `go mod verify` succeeds;
- no accidental module downgrade or pseudo-version enters the graph.

## 9. Required local verification

At minimum from `backend-go`:

```text
go mod tidy
go mod verify
go test ./...
go vet ./...
```

The implementation must also run the repository's applicable vulnerability/dependency scan using the
same pinned policy used by CI.

Race/postgres-backed tests must be executed where the repository's existing environment supports
them; exact-head GitHub CI remains authoritative for all ten required protected-branch checks.

Every command outcome must be recorded exactly. A non-zero process exit must never be summarized as
a successful execution merely because earlier substantive checks passed.

## 10. Required HTTP regression coverage

The update must preserve existing OpenInvest API behavior across at least:

- app construction and route registration;
- health and readiness endpoints;
- auth register/login/refresh/logout;
- refresh-cookie set/clear behavior;
- request headers and credentials behavior;
- local-development CORS behavior;
- path parameters and route matching;
- query parsing and pagination;
- request-body handling and strict JSON admission;
- `fiber.Ctx` header/cookie/query/body semantics used by current code;
- `c.Context()` service propagation behavior as currently tested/implemented;
- asset search/detail;
- portfolio list/create/detail/summary;
- transaction list/append;
- import review/append;
- response status, headers, cookies, and error envelopes;
- OpenAPI-visible behavior remaining unchanged.

Existing tests are authoritative where they cover these surfaces.

Because the approved Fiber target is already known to move the selected `golang.org/x/crypto`
version used by OpenInvest authentication, implementation must include a targeted historical-password
compatibility regression proving all of the following:

- a password hash encoded before the dependency update remains verifiable after the update;
- the same wrong-password case remains rejected;
- OpenInvest's Argon2id parameters remain unchanged;
- OpenInvest's encoded password-hash format remains unchanged;
- no password rehash/migration policy is silently introduced by this maintenance stage.

This regression is compatibility evidence for the shared-direct module movement. It is not evidence
that the pre-update authentication implementation was defective.

If an upstream semantic change causes a regression, do not normalize it as "expected maintenance";
treat it as a blocking compatibility issue.

## 11. Upstream-change review focus

Review must pay specific attention to `v3.3.0 -> v3.4.0 -> v3.5.0` changes that can touch OpenInvest's
actual core usage:

- routing and route matching;
- HTTP method/status behavior;
- request/response header semantics;
- content negotiation;
- request-body access;
- binder behavior only where OpenInvest actually calls Fiber binders;
- cookie behavior;
- `fiber.Ctx` behavior;
- app initialization/listen/shutdown behavior;
- indirect `fasthttp`, `schema`, `utils`, compression, and related module updates.

Middleware not imported by OpenInvest is still relevant to dependency provenance but must not be
misrepresented as an active application path.

## 12. Required implementation evidence

Before implementation can receive final approval, evidence must show:

1. exact base identity;
2. exact pre-update `go.mod` / `go.sum` blobs;
3. exact Fiber target `v3.5.0`;
4. explicit known `x/crypto v0.51.0 -> selected v0.54.0` shared-direct movement;
5. root `go.mod` declaration versus effective MVS-selected version;
6. `go list -m all` or equivalent exact resolved-module evidence;
7. reproducible module graph;
8. enumerated direct/shared-direct/transitive drift;
9. targeted pre-update Argon2id historical-hash compatibility regression;
10. unchanged Argon2 parameters and encoded password-hash format;
11. required tests/vet/module verification pass;
12. vulnerability/dependency scan result;
13. no unrelated direct dependency drift;
14. no unrelated source change;
15. exact-head GitHub CI with all ten required jobs successful after publication;
16. no unresolved material review finding.

## 13. Rollback policy

If Fiber `v3.5.0` causes a blocking regression before merge:

- do not merge;
- first determine whether a narrow behavior-preserving compatibility fix is possible;
- if the required change materially expands scope, stop and return to planning.

A downgrade after protected activation is an incident/recovery action, not a normal long-lived
completion strategy for P3-10.

## 14. Mandatory development-path review gates

After this plan is accepted, implementation remains on the canonical development path:

1. feature branch from the exact then-current protected `develop`;
2. exact scoped dependency change and evidence;
3. local quality gates;
4. Internal line-by-line read-only review in the designated review chat;
5. Builder remediation of all findings;
6. rerun required gates;
7. explicit human commit/push authorization;
8. Draft PR to `develop`;
9. required exact-head GitHub CI;
10. fresh External published-head review in the same designated review chat;
11. evidence-only Internal-evidence publication after External verdict as required by
    `REVIEW_WORKFLOW.md` v1.3.0;
12. CI on the evidence head;
13. same-chat exact verification;
14. separate explicit human Ready + squash-merge authorization;
15. separate docs-only closure-governance activation.

Internal evidence remains withheld from the Draft PR/repository evidence surface until the External
verdict, exactly as required by the effective canonical workflow.

No review verdict by itself authorizes a protected mutation.

## 15. Publication-stability and forensic rules

All implementation and closure records must preserve append-only factual history.

They must distinguish:

- substantive gate outcomes;
- tooling failures;
- process exit status;
- manual reruns;
- Internal review findings;
- External review findings;
- remediation;
- exact published heads;
- CI;
- actual merge facts.

Tool/client-session mechanics must not be promoted into permanent canonical history unless they are
self-contained in review evidence or independently observable.

Temporary lifecycle phrases such as `PRE-COMMIT`, `pending CI`, `current head`, or unqualified
`next step` must not be left in publication-stable records where they will become false after the
next lifecycle transition.

## 16. Closure semantics

P3-10 remains OPEN throughout Stage 3.43 planning and Stage 3.44 implementation.

It becomes CLOSED only after:

- approved Fiber maintenance implementation is actually squash-merged into protected `develop`; and
- a separate reviewed docs-only closure-governance record and synchronized canonical surfaces are
  activated on protected `develop`.

If no other audit finding changes concurrently, after P3-10 closure:

- closed: 29 / 32;
- completion: 90.625%;
- remaining: 3 / 32;
- remaining findings: P3-06, P3-07, P3-08.

No future PR number, CI run, implementation head, or merge SHA is predicted by this plan.
