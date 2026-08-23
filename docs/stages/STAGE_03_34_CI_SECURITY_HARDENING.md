# Stage 3.34 — CI/Security Hardening Implementation

| Field | Value |
| --- | --- |
| Status | Implementation candidate; exact-head CI and independent review pending |
| Canonical base | `develop` at `b4299bcdc28202c27388642dc7b426b159bb315c` |
| Branch | `fix/stage-03-34-ci-security-hardening` |
| Implementation PR | #74 |
| Finding | P2-17 |
| Scope | CI/security verification only |
| Out of scope | P2-16 repository settings, all P3 findings, Stage 3.25 privacy work, runtime financial behavior, SQL/OpenAPI/product changes, unrelated dependency upgrades |

## 1. Finding / symptom

The repository had six useful functional CI jobs but no mandatory Go static-analysis gate, no race-enabled PostgreSQL-backed suite, no Go vulnerability gate, no dependency-security audit for the locked Go/pnpm/Python dependency surfaces, and no scheduled security verification when no pull request was open.

The existing workflow already had bounded concurrency with `cancel-in-progress: true`; Stage 3.34 does not misclassify concurrency as absent.

## 2. Root cause

The original CI evolved around functional correctness and migration/runtime-boundary verification. Security tooling and race detection were never promoted into first-class merge gates, and the workflow only ran for pull requests. As a result, dependency advisories or race regressions could appear after a previously green PR without any scheduled re-evaluation.

## 3. Failure scenarios

Without this remediation, several classes of regression could reach or remain on `develop` without mechanical detection:

- a Go code change introduces a vet-detectable defect but all tests still pass;
- a concurrent code path contains a data race that ordinary `go test` does not detect;
- reachable Go code depends on a package/version with a newly disclosed vulnerability;
- a locked pnpm or Python dependency becomes known-vulnerable after merge;
- a repository remains unchanged for days while vulnerability databases change, leaving no new CI run to surface the new advisory.

## 4. Project impact

The impact is operational and security-oriented rather than a change to financial methodology. Missing gates increase the chance of concurrency defects, vulnerable dependencies, and security regressions surviving normal pull-request CI. For OpenInvest this matters because authentication, immutable-ledger, import, and PostgreSQL boundaries are correctness-sensitive and should not rely only on ordinary unit/integration execution.

## 5. Severity rationale

This remains P2 because the gap affects repository-wide prevention and detection controls but was not evidence of an already exploited vulnerability or a known P1 financial/security failure. The remediation strengthens the delivery system rather than changing product behavior.

## 6. Existing guarantees violated

The frozen delivery workflow requires required checks before merge, reproducible verification, least-privilege CI credentials, and evidence strong enough for independent review. The original workflow did not include the security/race class required by the repository audit.

## 7. Considered solutions

The implementation considered:

1. GitHub-native Dependency Review for all dependency ecosystems;
2. a third-party all-ecosystem GitHub Action scanner;
3. ecosystem-native pinned tools in a new dedicated security workflow;
4. extending the already-canonical CI workflow with narrowly scoped security jobs while preserving its six existing jobs.

## 8. Chosen remediation

The existing `.github/workflows/ci.yml` remains the canonical workflow. Its six pre-existing jobs and check names are preserved, and four additional jobs plus scheduled/manual triggers are added to that same workflow.

The workflow continues to run on pull requests to `develop`/`main`, and now also runs nightly at `03:17 UTC` and manually through `workflow_dispatch`. Repository-level permissions remain `contents: read`, and the existing bounded concurrency policy remains active.

The four stable added check names are:

- `Go vet` — `go vet ./...` using the Go version declared by `backend-go/go.mod`;
- `Go race tests` — PostgreSQL-backed `go test -race ./...` with the same migration and least-privilege runtime-role setup as the existing Go suite;
- `Go vulnerability scan` — official `govulncheck` pinned to `golang.org/x/vuln` v1.7.0 and run against `./...`;
- `Dependency security scan` — pinned pnpm 11.8.0 `pnpm audit` against `frontend-next/pnpm-lock.yaml`, plus `pip-audit` 2.10.1 against the Python locked project with `--locked --strict`.

The Go dependency surface is covered by `govulncheck`, which evaluates the module graph and narrows blocking findings to vulnerabilities reachable from the source build configuration. The pnpm and Python dependency surfaces are audited directly from their locked dependency state.

## 9. Why this solution was chosen

The ecosystem-native approach avoids assuming that this private repository has GitHub Code Security / Advanced Security entitlement. It provides real vulnerability gates using public ecosystem services while keeping tool versions explicit and reviewable.

Extending the existing CI workflow has an additional governance advantage: the workflow already exists on canonical `develop`, so a pull request changing it can produce exact-head evidence for the newly added jobs before merge. This avoids a bootstrap gap where a brand-new scheduled/security workflow might not provide usable pre-merge evidence on its first introduction.

The existing six functional CI jobs remain intact. Race testing duplicates the proven PostgreSQL/runtime-role setup rather than creating a weaker in-memory or owner-only test environment.

## 10. Rejected alternatives

### GitHub-native Dependency Review as the only dependency gate

Rejected as an unconditional implementation because private-repository availability depends on GitHub security entitlement that has not been established. Stage 3.34 must fail closed rather than claim an unavailable product feature.

### Separate new security workflow as the final shape

Initially attempted, then rejected during builder verification because its first-introduction PR did not provide the independent workflow run needed for exact-head pre-merge evidence. Keeping unverified checks until after merge would invert the reviewed workflow gate. The jobs were therefore moved into the already-canonical CI workflow and the temporary standalone workflow was removed.

### Replacing or weakening the existing six CI jobs

Rejected because P2-17 does not require redesigning already-green functional gates. The implementation extends the existing workflow and preserves those jobs.

### Running race tests without PostgreSQL

Rejected because significant Go integration behavior depends on PostgreSQL transactions, migrations, and the least-privilege runtime role. A reduced race suite would not prove the same production-relevant paths.

### Unpinned `latest` security tools

Rejected because tool drift between PRs would make failures harder to reproduce and would violate the repository's supply-chain discipline.

## 11. Trade-offs

The extended workflow increases GitHub Actions time. Because the schedule is attached to the canonical CI workflow, nightly runs execute both the six functional jobs and the four security jobs. This costs more runner time than a security-only scheduled workflow, but it provides one mechanically testable workflow surface and stronger nightly regression evidence without a first-merge bootstrap exception.

Nightly dependency/vulnerability checks can fail without a source change when a new advisory is published; that behavior is intentional.

`govulncheck` focuses on reachable Go vulnerabilities and therefore is intentionally lower-noise than a module-only block on every advisory in unused code. Dependency audit results still require engineering judgment before any remediation upgrade; Stage 3.34 does not authorize unrelated dependency churn.

## 12. Regression / verification coverage

The implementation itself is verified by GitHub Actions on PR #74. Required evidence before P2-17 can close:

- all six pre-existing CI jobs remain green;
- `Go vet` is green;
- `Go race tests` is green against PostgreSQL and the least-privilege runtime role;
- `Go vulnerability scan` is green with pinned govulncheck v1.7.0;
- `Dependency security scan` is green for pnpm and Python locked dependencies;
- GitHub accepts the modified workflow and exposes the added check names on the PR;
- the nightly schedule is present in the canonical workflow definition;
- no job receives write-capable repository permissions.

## 13. Adversarial review requirements

Independent review must verify at minimum:

- no existing CI check was removed or weakened;
- race testing uses the same database/migration/runtime-role boundary as ordinary Go integration tests;
- `govulncheck` is pinned and actually fails on reachable known vulnerabilities;
- pnpm and Python audits use locked dependency state;
- scheduled runs do not require secrets and cannot write repository contents;
- all referenced GitHub Actions remain pinned to immutable commit SHAs;
- added check names are stable enough for later P2-16 branch protection;
- moving the jobs into existing `ci.yml` does not alter the established least-privilege/concurrency semantics.

Any blocking review finding must be preserved here with its root cause, impact, remediation iteration, and new exact-head evidence.

## 14. Remediation iterations

**Builder verification iteration 1:** a standalone `.github/workflows/security.yml` was created with the four new jobs. After PR #74 opened, the existing CI ran but the newly introduced standalone workflow did not provide the required pre-merge PR evidence. Because P2-17 acceptance requires exact-head proof rather than trusting a workflow that would run only after becoming canonical, the standalone file was rejected as the final shape.

**Builder verification iteration 2:** the four jobs and nightly/manual triggers were integrated into the already-canonical `.github/workflows/ci.yml`, preserving all six existing jobs. The temporary standalone security workflow was deleted. Exact-head CI on this shape is pending.

No independent-review remediation has occurred yet. This section must be extended if the independent implementation review returns `REQUEST CHANGES`.

## 15. Residual risk / limitations

P2-17 closure does not itself protect `develop`; P2-16 remains OPEN until GitHub repository settings mechanically require the reviewed checks and prevent bypass, including the administrator/owner path.

Vulnerability databases are external and can change independently of source control. A green scan proves the database/tool result at the run time, not the permanent absence of future advisories.

The workflow does not add SAST, secret scanning, container-image scanning, or license-policy enforcement because those were not part of P2-17's approved audit scope.

## 16. Operational / deployment consequences

There is no application deployment or database migration consequence. Repository operations gain four additional PR checks and a nightly full CI/security run. After the implementation is merged, the exact final check names become inputs to the manual P2-16 GitHub Settings enforcement step.

## 17. Exact evidence

Planning gate:

- planning PR #71;
- approved planning head `0583fda8da92cbc15efd7e0497bd36027956c87e`;
- exact-head planning CI #205: SUCCESS, six of six existing jobs;
- independent planning review: `APPROVED` after the administrator/owner bypass blocker was corrected;
- explicit human squash-merge authorization;
- planning squash merge: `b4299bcdc28202c27388642dc7b426b159bb315c`.

Implementation:

- PR #74;
- implementation branch `fix/stage-03-34-ci-security-hardening`;
- final exact implementation head, workflow run, reviewer verdict, and merge SHA are pending and must be recorded before closure.

## 18. Canonical status

P2-17 remains **OPEN** while this implementation is only a candidate. It may be marked CLOSED only after exact-head green CI/security evidence, independent implementation review `APPROVED`, explicit human squash-merge authorization, squash merge into `develop`, and closure governance.

P2-16 remains **OPEN** regardless of P2-17 implementation status until the repository's effective GitHub protection and squash-only settings are mechanically enforced and API-verified.
