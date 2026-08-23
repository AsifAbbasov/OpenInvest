# Stage 3.34 — CI/Security Hardening Implementation

| Field | Value |
| --- | --- |
| Status | Implementation candidate; exact-head CI and independent review pending |
| Canonical base | `develop` at `b4299bcdc28202c27388642dc7b426b159bb315c` |
| Branch | `fix/stage-03-34-ci-security-hardening-final` |
| Finding | P2-17 |
| Scope | CI/security verification plus the minimum dependency patches required for the new gates to become honestly green |
| Out of scope | P2-16 repository settings, P3 closure claims, Stage 3.25 privacy work, SQL/OpenAPI/product behavior changes, unrelated dependency upgrades |

## 1. Finding / symptom

The repository had six useful functional CI jobs but no mandatory Go static-analysis gate, no race-enabled PostgreSQL-backed suite, no Go vulnerability gate, no dependency-security audit for the locked Go/pnpm/Python dependency surfaces, and no scheduled security verification when no pull request was open.

The existing workflow already had bounded concurrency with `cancel-in-progress: true`; Stage 3.34 does not misclassify concurrency as absent.

## 2. Root cause

The original CI evolved around functional correctness and migration/runtime-boundary verification. Security tooling and race detection were never promoted into first-class merge gates, and the workflow only ran for pull requests. When the new gates were executed, they also exposed a stale security baseline that the old CI could not detect.

## 3. Failure scenarios

Without this remediation:

- vet-detectable defects can pass ordinary tests;
- concurrency defects can escape non-race test runs;
- reachable Go vulnerabilities can remain in the toolchain or dependency graph;
- vulnerable pnpm/Python locked dependencies can remain undetected;
- newly disclosed advisories can remain invisible while the repository is idle.

## 4. Project impact

The impact is operational and security-oriented rather than a change to financial methodology. Missing gates increase the chance of concurrency defects and vulnerable dependencies surviving normal pull-request CI. The first real execution proved this risk was present: the baseline already contained reachable Go vulnerabilities and vulnerable frontend dependencies.

## 5. Severity rationale

P2 is retained because the gap affects repository-wide prevention and detection controls. The newly discovered dependency advisories are remediated only as necessary to make the approved P2-17 gate truthful; this stage does not silently close separate P3 dependency-maintenance findings.

## 6. Existing guarantees violated

The frozen delivery workflow requires reproducible verification, least-privilege CI credentials, required checks before merge, and review evidence strong enough for independent validation. The original workflow did not cover static analysis, races, vulnerability checks, dependency audit, or scheduled re-evaluation.

## 7. Considered solutions

1. GitHub-native Dependency Review;
2. third-party all-ecosystem scanners;
3. ecosystem-native pinned tooling;
4. a new standalone security workflow;
5. extending the existing canonical CI workflow.

## 8. Chosen remediation

The canonical `.github/workflows/ci.yml` keeps all six existing jobs and adds four stable jobs:

- `Go vet` — `go vet ./...`;
- `Go race tests` — PostgreSQL-backed `go test -race ./...` using the same migrations and least-privilege runtime role as ordinary Go integration tests;
- `Go vulnerability scan` — `govulncheck` pinned to v1.7.0;
- `Dependency security scan` — pnpm audit plus Python dependency audit derived from the frozen uv lock.

The same workflow gains nightly and manual triggers while repository permission remains `contents: read`.

The first exact execution exposed real stale dependencies. The minimum security remediation selected from disposable builder evidence is:

- Go language/toolchain declaration: `1.25.0` -> `1.25.14`;
- `github.com/jackc/pgx/v5`: `v5.7.6` -> `v5.9.2`;
- `golang.org/x/net`: `v0.54.0` -> `v0.55.0`;
- `golang.org/x/text`: `v0.37.0` -> `v0.39.0`;
- transitive Go modules adjusted by `go mod tidy` (`x/sync`, `x/sys`, and tidy-only metadata);
- Next.js: `16.2.9` -> `16.3.2` with generated pnpm lockfile.

Next.js 16.2.11 was explicitly tested as the smaller candidate but rejected because its locked dependency graph still produced 7 audit findings (5 high, 2 moderate), including vulnerable `sharp`, `postcss`, and `nanoid`. Next.js 16.3.2 passed typecheck, tests, build, and `pnpm audit` with zero advisories in the builder evidence.

For Python, `pip-audit --locked .` was rejected after real execution because pip-audit does not treat `uv.lock` as a supported lockfile. The final gate exports the existing `uv.lock` without re-locking using `uv export --frozen --format requirements.txt`, then audits that pinned export with `pip-audit -r ... --strict`. `uv.lock` remains the only committed Python lock source of truth.

## 9. Why this solution was chosen

The ecosystem-native approach avoids assuming unavailable GitHub Code Security entitlement. Extending the already-canonical CI workflow avoids the first-introduction bootstrap gap observed with a new standalone workflow.

The dependency patches are not discretionary modernization. The new gates demonstrated that the prior baseline could not pass an honest security scan. Candidate generation was therefore isolated in disposable, never-merged builder PRs and selected from measured results rather than version preference.

## 10. Rejected alternatives

### GitHub-native Dependency Review as the only gate

Rejected because availability for this private repository depends on security entitlement that has not been established.

### Standalone security workflow

Rejected after builder verification because its first-introduction PR did not provide the required pre-merge exact-head evidence.

### Suppressing audit findings or lowering severity

Rejected because that would make P2-17 mechanically present but semantically ineffective.

### Next.js 16.2.11

Rejected after real builder verification: typecheck/tests/build passed, but audit still returned 5 high and 2 moderate vulnerabilities.

### `pip-audit --locked` directly against `uv.lock`

Rejected after real execution: the scanner reported `no lockfiles found`. The approved replacement preserves `uv.lock` as source of truth and audits a frozen export.

### Broad dependency modernization

Rejected. Only versions needed to remove findings discovered by the new P2-17 gates are changed. Fiber remains unchanged; P3-10 remains separate. P3-09 is not silently closed merely because a security-required Next.js patch is included here.

## 11. Trade-offs

The workflow consumes more Actions time, and nightly runs can fail without a source change when vulnerability databases are updated. That is intentional. The schedule currently executes the full ten-job workflow rather than only security jobs, trading runner cost for a single mechanically verified workflow surface.

## 12. Regression / verification coverage

Closure evidence must prove on one immutable implementation head:

- all six pre-existing CI jobs green;
- `Go vet` green;
- PostgreSQL-backed `Go race tests` green;
- `Go vulnerability scan` green;
- `Dependency security scan` green for pnpm and Python frozen-lock-derived dependencies;
- frontend typecheck/tests/build green after Next.js patch;
- PostgreSQL migration/runtime-boundary jobs unchanged and green;
- repository token permission remains read-only in the canonical workflow;
- nightly schedule remains defined.

## 13. Adversarial review requirements

Independent review must verify:

- no existing CI job was removed or weakened;
- race tests use the same PostgreSQL/migration/runtime-role boundary;
- govulncheck is pinned and blocking;
- dependency audit cannot silently skip Python or frontend surfaces;
- `uv export --frozen` does not alter `uv.lock`;
- action references are immutable SHAs;
- the dependency patches are narrowly justified by recorded audit evidence;
- P3-09/P3-10 are not falsely marked CLOSED;
- stable check names are suitable for P2-16 enforcement.

## 14. Remediation iterations

**Iteration 1 — standalone workflow bootstrap:** a new `security.yml` was created. The original CI ran, but the first-introduction standalone workflow did not provide usable pre-merge evidence. It was removed.

**Iteration 2 — canonical CI integration:** the four jobs were moved into existing `ci.yml`. A YAML scalar error in the govulncheck command prevented parsing; this builder defect was corrected before any merge.

**Iteration 3 — exact execution:** CI run #216 exposed real baseline debt. `Go vet` and `Go race tests` passed. `Dependency security scan` found 16 frontend advisories (9 high, 7 moderate). `govulncheck` reported 32 reachable vulnerabilities, dominated by the Go 1.25.0 standard library plus pgx/x/net/x/text.

**Iteration 4 — disposable dependency builders:** builder PRs generated real package-manager outputs. The final useful artifact showed:

- patched Go candidate: tests=0, vet=0, govulncheck=0; govulncheck output `No vulnerabilities found`;
- Next 16.2.11: typecheck=0, tests=0, build=0, audit=1 with 7 residual advisories;
- Next 16.3.2: typecheck=0, tests=0, build=0, audit=0 with zero advisories;
- direct `pip-audit --locked .` against the uv project: exit=1 because the tool supports pylock, not uv.lock.

The generated Next 16.3.2 lockfile SHA-256 is `f492c1d06aff6bed6e21d839fc510e3402615fa624ef59215d9b12c444314336`. A disposable never-merged builder workflow is used only to regenerate that lockfile from the reviewed package manifest and push it to the implementation branch if and only if the digest matches exactly.

## 15. Residual risk / limitations

P2-17 closure does not protect `develop`; P2-16 remains OPEN until GitHub settings mechanically require the final checks and restrict bypass.

Vulnerability databases can change after any green run. P3 dependency-maintenance findings remain separately governed even where a package received a security patch here.

## 16. Operational / deployment consequences

No application feature, SQL schema, financial methodology, or API contract changes. Delivery gains four PR security/race gates and scheduled verification. Runtime dependencies change only to remove vulnerabilities exposed by those gates.

## 17. Exact evidence

Planning:

- PR #71;
- approved planning head `0583fda8da92cbc15efd7e0497bd36027956c87e`;
- exact-head planning CI #205: SUCCESS, six of six;
- independent planning review: APPROVED;
- explicit human squash-merge authorization;
- planning squash merge `b4299bcdc28202c27388642dc7b426b159bb315c`.

Implementation history:

- #74/#75: historical builder/bootstrap PRs, closed without merge;
- #76: active implementation PR;
- exact run #216 proved ten jobs execute and exposed the dependency baseline failures;
- disposable dependency builder PRs #77/#78 are never-merge tooling only;
- final immutable implementation head, final green workflow run, independent reviewer verdict, and merge SHA remain pending.

## 18. Canonical status

P2-17 remains **OPEN** until the final immutable implementation head has all ten checks green, independent implementation review returns `APPROVED`, explicit human squash-merge authorization is given, the PR is squash-merged, and closure governance is completed.

P2-16 remains **OPEN** regardless of P2-17 status until effective GitHub protection and squash-only settings are mechanically enforced and API-verified.
