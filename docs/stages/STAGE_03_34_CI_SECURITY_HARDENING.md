# Stage 3.34 — CI/Security Hardening Implementation

| Field | Value |
| --- | --- |
| Status | Implementation candidate; exact-head CI and independent review pending |
| Canonical base | `develop` at `b4299bcdc28202c27388642dc7b426b159bb315c` |
| Final review branch | `fix/stage-03-34-ci-security-hardening-final` |
| Historical builder/bootstrap PR | #74 |
| Final implementation PR | Pending at branch freeze; PR metadata is the authoritative final PR identifier |
| Finding | P2-17 |
| Scope | CI/security verification only |
| Out of scope | P2-16 repository settings, all P3 findings, Stage 3.25 privacy work, runtime financial behavior, SQL/OpenAPI/product changes, unrelated dependency upgrades |

## 1. Finding / symptom

The repository had six useful functional CI jobs but no mandatory Go static-analysis gate, no race-enabled PostgreSQL-backed suite, no Go vulnerability gate, no dependency-security audit for the locked Go/pnpm/Python dependency surfaces, and no scheduled security verification when no pull request was open.

The existing workflow already had bounded concurrency with `cancel-in-progress: true`; Stage 3.34 does not misclassify concurrency as absent.

## 2. Root cause

The original CI evolved around functional correctness and migration/runtime-boundary verification. Security tooling and race detection were never promoted into first-class merge gates, and the workflow only ran for pull requests. Dependency advisories or race regressions could therefore appear after a previously green PR without scheduled re-evaluation.

## 3. Failure scenarios

Without remediation:

- a Go change can introduce a vet-detectable defect while tests remain green;
- a concurrent path can contain a data race invisible to ordinary `go test`;
- reachable Go code can depend on a newly disclosed vulnerable package/version;
- a locked pnpm or Python dependency can become known-vulnerable after merge;
- vulnerability databases can change while the repository remains unchanged and no new CI run occurs.

## 4. Project impact

The impact is repository-wide operational/security risk rather than a change to financial methodology. Missing gates increase the chance of concurrency defects, vulnerable dependencies, and security regressions surviving normal pull-request verification. This matters because authentication, immutable-ledger, import, and PostgreSQL boundaries are correctness-sensitive.

## 5. Severity rationale

P2 is retained because the gap weakens repository-wide prevention/detection controls but is not evidence of an already exploited vulnerability or known P1 financial/security failure.

## 6. Existing guarantees violated

The delivery workflow requires reproducible required checks, least-privilege CI credentials, and evidence strong enough for independent review. The original workflow did not include the security/race class required by the repository audit.

## 7. Considered solutions

1. GitHub-native Dependency Review for all ecosystems.
2. A third-party all-ecosystem GitHub Action scanner.
3. Ecosystem-native pinned tools in a new dedicated security workflow.
4. Extending the already-canonical CI workflow with narrowly scoped security jobs while preserving its six existing jobs.

## 8. Chosen remediation

The existing `.github/workflows/ci.yml` remains canonical. Its six pre-existing jobs/check names are preserved. The workflow gains scheduled/manual triggers and four stable jobs:

- `Go vet` — `go vet ./...` using `backend-go/go.mod` toolchain selection;
- `Go race tests` — PostgreSQL-backed `go test -race ./...` after real migrations and the same least-privilege runtime-role setup as ordinary Go CI;
- `Go vulnerability scan` — official `govulncheck` pinned to `golang.org/x/vuln` v1.7.0 and run against `./...`;
- `Dependency security scan` — pinned pnpm 11.8.0 `pnpm audit` plus pinned `pip-audit` 2.10.1 against the Python locked project with `--locked --strict`.

The workflow continues on pull requests to `develop`/`main`, also runs nightly at `03:17 UTC`, and supports `workflow_dispatch`. Repository-level permissions remain `contents: read`; the existing bounded concurrency policy remains active.

Go dependency vulnerability coverage is provided by `govulncheck`, which evaluates the module graph and reachable vulnerable code. pnpm and Python surfaces are audited from locked dependency state.

## 9. Why this solution was chosen

Ecosystem-native tools avoid assuming private-repository GitHub Code Security/Advanced Security entitlement while still providing real blocking vulnerability checks with explicit versions.

Extending the existing canonical workflow also solves a governance bootstrap problem: a workflow already present on `develop` can produce pre-merge evidence when modified by a PR. The six existing jobs stay intact, and the new stable job names can later become P2-16 required checks.

## 10. Rejected alternatives

### GitHub-native Dependency Review as the only dependency gate

Rejected as unconditional scope because private-repository availability depends on an entitlement not established for this repository. Unsupported product behavior must not be claimed as enforcement.

### Separate newly introduced security workflow as the final shape

Initially attempted on historical builder PR #74, then rejected because the new workflow did not provide the required pre-merge exact-head run on first introduction. Accepting it without execution evidence would invert the reviewed workflow gate.

### Replacing or weakening the six existing jobs

Rejected because P2-17 requires additive hardening, not redesign of already-reviewed functional gates.

### Race tests without PostgreSQL

Rejected because important Go paths depend on PostgreSQL transactions, migrations, and runtime-role behavior.

### Unpinned `latest` tools

Rejected because tool drift undermines reproducibility and supply-chain discipline.

## 11. Trade-offs

Nightly runs execute all ten jobs, consuming more Actions time than a security-only schedule. This is accepted to keep one mechanically testable canonical workflow and stronger nightly regression evidence.

Security scans depend on changing external advisory databases, so a nightly run can fail without source changes. That is intentional. `govulncheck` prioritizes reachable vulnerabilities to reduce noise; dependency remediation still requires engineering judgment and does not authorize unrelated upgrade churn.

## 12. Regression / verification coverage

P2-17 cannot close until exact-head GitHub Actions evidence proves:

- all six pre-existing jobs remain green;
- `Go vet` is green;
- `Go race tests` is green against PostgreSQL and the least-privilege runtime role;
- `Go vulnerability scan` is green with pinned govulncheck v1.7.0;
- `Dependency security scan` is green for locked pnpm/Python dependencies;
- GitHub accepts the modified workflow and exposes the four stable added check names;
- the nightly schedule exists in the canonical workflow definition;
- no job receives write-capable repository permissions.

## 13. Adversarial review requirements

Independent review must verify at minimum:

- no existing CI check was removed or weakened;
- race testing uses the same database/migration/runtime-role boundary as ordinary Go integration tests;
- govulncheck is pinned and blocking for reachable known vulnerabilities;
- pnpm/Python audits use locked dependency state;
- scheduled runs require no write credentials or secrets;
- GitHub Actions references remain immutable-SHA pinned;
- added check names are stable for later P2-16 protection;
- least-privilege and concurrency semantics remain intact.

Any blocking review finding must be preserved here with root cause, impact, remediation iteration, and renewed exact-head evidence.

## 14. Remediation iterations

**Builder iteration 1:** a standalone `.github/workflows/security.yml` was introduced. Historical PR #74 opened and the pre-existing CI ran, but the new standalone workflow supplied no usable first-introduction pre-merge evidence. The design was rejected rather than trusted without execution.

**Builder iteration 2:** the four jobs plus nightly/manual triggers were integrated into the already-canonical `.github/workflows/ci.yml`; the temporary standalone workflow was deleted. This preserves six prior jobs and makes all ten jobs part of one existing workflow.

**Builder iteration 3:** connector-generated commits after PR #74 opened did not create a new `synchronize` workflow run, and close/reopen also produced no exact-head run. To avoid treating CI #210 from an older head as evidence, a clean final review branch was frozen before opening the final implementation PR. The final PR `opened` event is used only to obtain honest exact-head execution evidence; no code is weakened or bypassed.

No independent-review remediation has occurred yet.

## 15. Residual risk / limitations

P2-17 closure does not protect `develop`; P2-16 remains OPEN until repository settings mechanically require the reviewed checks and prevent bypass, including administrator/owner paths.

A green vulnerability scan is point-in-time evidence, not permanent absence of future advisories. SAST, secret scanning, container-image scanning, and license-policy enforcement are outside the approved P2-17 audit scope.

## 16. Operational / deployment consequences

No application deployment or database migration changes. Repository operations gain four PR checks and a nightly full CI/security run. After implementation merge, the ten stable final check names become inputs to P2-16 GitHub Settings enforcement.

## 17. Exact evidence

Planning gate:

- planning PR #71;
- approved planning head `0583fda8da92cbc15efd7e0497bd36027956c87e`;
- planning CI #205: SUCCESS, six of six existing jobs;
- repeat independent planning review: `APPROVED`;
- explicit human squash-merge authorization;
- planning squash merge `b4299bcdc28202c27388642dc7b426b159bb315c`.

Implementation builder history:

- historical PR #74 records initial workflow/bootstrap iterations;
- final review branch: `fix/stage-03-34-ci-security-hardening-final`;
- final PR number, exact head, exact-head workflow run, reviewer verdict, merge authorization, and merge SHA remain pending and are recorded through PR metadata/closure governance once known.

## 18. Canonical status

P2-17 remains **OPEN**. It may become CLOSED only after final exact-head green CI/security evidence, independent implementation review `APPROVED`, explicit human squash-merge authorization, squash merge into `develop`, and closure governance.

P2-16 remains **OPEN** regardless of P2-17 implementation status until effective GitHub protection and squash-only repository settings are mechanically enforced and API-verified.
