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
gates, mobile, tax, or broker integrations.


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


If the current GitHub plan does not expose branch protection/rulesets for this private repository, or
if the available protection mechanism cannot enforce the reviewed policy against the repository
administrator/owner path, the human settings step must stop rather than weaken the acceptance
criteria. P2-16 remains OPEN until the repository has a supported mechanical protection mechanism
that also covers administrators/owners. Purchasing a GitHub plan or changing repository visibility is
an account/product decision outside Stage 3.34 and requires a separate explicit user choice.

## Execution order


## Acceptance criteria

### P2-17 may be CLOSED only if

- `go vet ./...` is a green required PR check;
- PostgreSQL-backed `go test -race ./...` is a green required PR check;
- `govulncheck` is reproducibly executed and green under the documented policy;
- a scheduled/nightly security workflow exists and is valid;
- existing CI gates remain green;
- workflow permissions and action pinning do not introduce a new P1/P2 security regression.

### P2-16 may be CLOSED only if


## Closure target

If both P2-16 and P2-17 satisfy the acceptance criteria and the final implementation/governance

- P0: 0
- P1: 0
- P2: 0
- P3: 10
- total remaining: 10

This does not imply production readiness. The ten P3 findings and the separate Stage 3.25 privacy
