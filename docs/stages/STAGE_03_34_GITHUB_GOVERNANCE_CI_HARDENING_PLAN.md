# Stage 3.34 — GitHub Governance and CI/Security Hardening Plan

| Field | Value |
| --- | --- |
| Status | Planning/remediation gate only; implementation not yet authorized |
| Baseline | `develop` at `71a1faeb97d33d05f2936111b53f1285edddabe9` |
| Audit findings | P2-16, P2-17 |
| Scope | Repository governance and CI/security hardening only |
| Architecture impact | None |
| Product impact | None |
| Privacy impact | None; Stage 3.25 remains separate |

## Purpose

Stage 3.34 is the final P2 remediation block from the original 32-finding repository audit. It is
limited to GitHub repository governance enforcement and CI/security/concurrency hardening. It does
not authorize product features, financial logic changes, database/schema changes, OpenAPI changes,
provider selection, privacy-lifecycle implementation, dependency upgrades unrelated to the security
gates, mobile, tax, AI, or broker integrations.

No implementation starts until this planning gate is independently reviewed and approved.

## Current evidence

### P2-16 — GitHub governance is not enforced

Read-only GitHub API inspection of canonical `develop` after Stage 3.33 closure shows:

- `develop` points to `71a1faeb97d33d05f2936111b53f1285edddabe9`;
- branch protection is disabled (`protected: false` / protection disabled);
- repository default branch is `develop`;
- repository settings currently allow merge commits, rebase merges, and squash merges.

Therefore the repository does not mechanically enforce the frozen delivery policy. P2-16 remains
OPEN.

GitHub's current documented feature matrix states that protected branches/rulesets for private
repositories require a plan that supports private-repository protection (for example GitHub Pro,
Team, or Enterprise). The connected repository API does not expose the account's billing plan.
Therefore Stage 3.34 must fail closed on feature availability: if protection/ruleset controls are not
available for this private repository, P2-16 remains OPEN. Changing repository visibility merely to
obtain protection is not authorized by this stage.

### P2-17 — CI/security class is incomplete

The canonical repository currently has one workflow, `.github/workflows/ci.yml`, triggered only by
pull requests to `develop` or `main`.

The workflow already has useful concurrency control:

- deterministic workflow/PR-or-ref concurrency group;
- `cancel-in-progress: true`.

The existing six CI jobs cover:

1. PostgreSQL-backed Go tests;
2. Python tests;
3. frontend typecheck/tests/build;
4. OpenAPI validation;
5. Docker Compose config validation;
6. PostgreSQL migration/rollback/runtime-ACL validation.

However the current CI does not provide the remaining audit-required security/concurrency class:

- no mandatory `go vet ./...` PR gate;
- no mandatory `go test -race ./...` PR gate;
- no `govulncheck` gate;
- no dependency-diff/security review gate;
- no scheduled/nightly vulnerability/security run.

P2-17 therefore remains OPEN. Existing concurrency is retained and is not falsely described as
missing.

## Proposed remediation

### Track A — P2-17 CI/security hardening

Implementation must preserve the existing six jobs and add narrowly scoped, reproducible security
checks without changing application behavior.

Required implementation outcomes:

1. **Go static analysis**
   - add a required PR job running `go vet ./...` against the canonical Go module;
   - use the repository Go version contract rather than an unrelated toolchain.

2. **Go race detection**
   - add a required PostgreSQL-backed PR job running `go test -race ./...`;
   - provision the same migrations and least-privilege runtime role required by the normal Go suite;
   - do not replace the existing non-race Go suite.

3. **Go vulnerability analysis**
   - add a reproducibly pinned `govulncheck` execution against `./...`;
   - fail on actionable known vulnerabilities in reachable Go code;
   - tool installation/version must be explicit and reviewable.

4. **Dependency/security review**
   - add a PR-time dependency-change security gate covering the repository lock/module manifests;
   - prefer a GitHub-native dependency-review gate when it is actually supported for this private
     repository;
   - GitHub's documented dependency-review action requires GitHub Code Security or GitHub Advanced
     Security for private repositories; do not assume that entitlement exists;
   - if repository-plan/API support makes the native gate unavailable, use a pinned, deterministic
     equivalent scanner and document the exact coverage rather than claiming unsupported GitHub
     security-product behavior;
   - include the Go, pnpm, and Python locked dependency surfaces in the chosen supported approach.

5. **Scheduled security verification**
   - add a scheduled/nightly workflow or scheduled mode that re-runs vulnerability/security checks
     even when no PR is open;
   - keep permissions least-privilege and concurrency bounded;
   - scheduled failure must be visible in GitHub Actions and must not silently auto-modify source or
     dependency files.

6. **Workflow supply-chain discipline**
   - continue pinning third-party GitHub Actions to immutable commit SHAs;
   - keep `GITHUB_TOKEN` permissions at the minimum needed by each workflow/job;
   - no untrusted PR code may receive write-capable repository credentials.

### Track B — P2-16 GitHub governance enforcement

After the Stage 3.34 CI implementation is merged and final required check names are stable, GitHub
repository settings must be configured and independently verified.

Required `develop` enforcement:

1. require changes to reach `develop` through pull requests;
2. require the final Stage 3.34-defined CI checks before merge;
3. require review conversations to be resolved before merge;
4. block force pushes to `develop`;
5. block deletion of `develop`;
6. require linear history where supported;
7. leave squash merge enabled and disable merge-commit and rebase-merge repository methods so the
   repository policy is squash-only;
8. disable the normal administrator/owner bypass for the required PR/check/protection policy and
   enforce the reviewed protection rules against administrators/owners whenever the selected GitHub
   protection mechanism supports that control;
9. verify through the GitHub API that the effective protection/ruleset state applies to the
   administrator/owner path as well as ordinary contributors;
10. if the repository/account configuration cannot mechanically prevent administrator/owner bypass,
    P2-16 remains OPEN; disclosure of that bypass is not sufficient for closure.

The connected GitHub tool available to this workflow can inspect repository and branch protection
state but does not expose a mutation for branch protection/rulesets or repository merge-method
settings. Therefore Track B contains one unavoidable human GitHub Settings action. The exact values
and click path must be provided only after Track A establishes the final required check names. The
assistant must then re-read the GitHub API state and refuse to close P2-16 if the effective settings do
not match the reviewed policy, including the administrator/owner enforcement path.

If the current GitHub plan does not expose branch protection/rulesets for this private repository, or
if the available protection mechanism cannot enforce the reviewed policy against the repository
administrator/owner path, the human settings step must stop rather than weaken the acceptance
criteria. P2-16 remains OPEN until the repository has a supported mechanical protection mechanism
that also covers administrators/owners. Purchasing a GitHub plan or changing repository visibility is
an account/product decision outside Stage 3.34 and requires a separate explicit user choice.

## Execution order

1. independently review and approve this Stage 3.34 plan;
2. create the Stage 3.34 implementation branch from the then-canonical `develop`;
3. implement Track A CI/security changes only;
4. run exact-head CI and independent implementation review;
5. obtain explicit human authorization and squash-merge the CI implementation;
6. apply the one human GitHub Settings change for Track B using the exact final check names, if the
   repository's GitHub plan supports the required protection controls;
7. verify branch protection/repository merge settings read-only through GitHub API, including that
   the required policy applies to the administrator/owner path;
8. if any protection requirement, required GitHub feature, or administrator/owner enforcement is
   missing, keep P2-16 OPEN;
9. record final independent evidence and closure governance;
10. only after canonical closure may the project claim P2=0.

## Acceptance criteria

### P2-17 may be CLOSED only if

- `go vet ./...` is a green required PR check;
- PostgreSQL-backed `go test -race ./...` is a green required PR check;
- `govulncheck` is reproducibly executed and green under the documented policy;
- dependency/security review has real, supported repository coverage and is green;
- a scheduled/nightly security workflow exists and is valid;
- existing CI gates remain green;
- workflow permissions and action pinning do not introduce a new P1/P2 security regression.

### P2-16 may be CLOSED only if

- GitHub API verification shows `develop` is effectively protected;
- direct unreviewed merge paths are constrained by the approved rules;
- required CI checks are enforced;
- conversation resolution is enforced;
- force-push and branch deletion are blocked;
- repository merge methods enforce squash-only policy;
- the protection/ruleset configuration mechanically applies the required PR/check/protection policy
  to administrators/owners rather than leaving the normal administrator bypass available;
- API verification confirms the effective administrator/owner path is covered by the reviewed
  enforcement policy;
- the protection mechanism is actually supported and active for this private repository; a plan or
  account limitation cannot be documented away as if enforcement existed;
- if administrator/owner bypass cannot be mechanically disabled under the available repository/account
  configuration, P2-16 remains OPEN. Disclosure alone is not sufficient for closure.

## Closure target

If both P2-16 and P2-17 satisfy the acceptance criteria and the final implementation/governance
review returns APPROVED, the original repository-audit severity backlog becomes:

- P0: 0
- P1: 0
- P2: 0
- P3: 10
- total remaining: 10

This does not imply production readiness. The ten P3 findings and the separate Stage 3.25 privacy
Security Review evidence work remain open.