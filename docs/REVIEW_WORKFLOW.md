# Review and Delivery Workflow

| Field | Value |
| --- | --- |
| Document ID | ENG-REVIEW-001 |
| Version | 1.0.0 |
| Status | Approved / Mandatory |
| Owner | Principal Architect |
| Supersedes | Informal review and push process |
| Dependencies | Architecture Freeze v1.2; Document 40; Document 43 |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |

## Purpose

Separate implementation, automated verification, specialist review, human approval, and merge authority. No builder may approve its own architectural work, and no review bypass may write directly to protected branches.

## Roles

### Builder Engineer — Codex

Owns scoped implementation, refactoring, tests, terminal execution, builds, documentation, feature branches, local commits, Draft PR preparation, review fixes, and stage reports. The Builder proposes a push but never pushes or merges without the required approval.

### Automated CI

Runs deterministic lint, formatting, unit, integration, contract, security, build, and relevant performance checks. Full expensive performance, resilience, and architecture suites may run nightly rather than on every PR when the PR gate remains risk-appropriate.

### Review Agent — ChatGPT

Reviews the Draft PR diff and evidence independently from the Builder. A review must address architecture/DDD/SOLID, API contracts, security/privacy, performance/cost, mathematical correctness when relevant, scope/YAGNI, migrations, rollback, and documentation.

Specialist Architecture, Security, Performance, API, and Mathematical reviews may be separate reports or clearly separated sections in one review. A model's self-review is advisory; the human remains accountable.

### Principal Architect / Human Reviewer

Owns final scope, product, architecture, risk acceptance, review verdict, and merge authorization. Only the human may declare the stage accepted for merge.

## Mandatory delivery sequence

```text
Approved stage scope
  → feature branch
  → implementation and documentation
  → local quality gates
  → Builder stage report and diff summary
  → human permission to commit/push feature branch
  → push feature branch
  → Draft PR targeting develop
  → required GitHub CI green
  → ChatGPT specialist review(s)
  → Builder fixes and CI rerun
  → ChatGPT approval/evidence
  → human review and approval
  → squash merge to develop
  → nightly verification
  → periodic architecture audit
```

GitHub Actions cannot evaluate an unpushed branch. Therefore local gates run before Draft PR; authoritative repository CI runs after the feature branch is pushed. This is not permission to push to `develop` or `main`.

## Branch convention

Branches describe intent, not author:

- `feature/openapi-stage-02`
- `feature/portfolio-crud`
- `feature/xirr-engine`
- `feature/dividend-calendar`
- `fix/snapshot-cache`
- `refactor/domain-events`
- `docs/review-workflow`

Use lowercase kebab-case. One branch and PR must represent one responsibility. Direct commits or pushes to `develop` and `main` are forbidden.

## Pull-request size

Default review budget:

- at most 25 changed files;
- at most 800 changed lines of hand-written business logic;
- one responsibility;
- one independently reversible outcome.

Generated code, lockfiles, snapshots, migrations, and imported specifications are reported separately and excluded from the hand-written business-line budget, but never hidden from review. An exception requires an explanation in the PR and Principal Architect approval before review begins.

## Required PR disclosure

Every PR states:

- stage and responsibility;
- user value and why the change is needed now;
- ADR affected or `None`;
- DDD/bounded contexts affected;
- OpenAPI changed;
- database/schema/migration changed;
- mathematical calculations affected;
- performance and cost impact;
- security and privacy impact;
- external data sources affected;
- backward-compatibility impact;
- rollback availability and procedure;
- local/CI evidence;
- files and line-count budget;
- explicit out-of-scope items.

## ADR gate

An ADR is mandatory when a change affects a frozen architecture decision, introduces or replaces a major framework/service/database/protocol, changes privacy/security/tax/math/snapshot/event semantics, creates vendor lock-in, changes external-data policy, or materially changes operational cost or rollback characteristics.

A routine, well-scoped library does not automatically require an ADR. It still requires dependency rationale, license/security review, maintenance assessment, alternatives considered, and removal/rollback notes in the PR. This distinction prevents governance from becoming ceremony that hides real risk.

## CI gate

Required checks are proportional to affected surfaces and may include:

- formatting and lint;
- static analysis and type checking;
- unit tests;
- integration/database tests;
- OpenAPI/contract compatibility;
- migration validation;
- dependency, secret, SAST, and license scanning;
- production build;
- financial golden vectors;
- focused performance regression checks.

A red required check blocks review approval and merge. Flaky checks are fixed or explicitly quarantined through an approved issue; they are not casually rerun until green.

## Review verdict

Each Review Agent report ends with exactly one verdict:

- `APPROVED`
- `REQUEST CHANGES`
- `BLOCKED — insufficient evidence`

Findings include severity, file/line where possible, violated principle, user/operational impact, and a minimal recommendation. Style preferences do not block a PR unless they violate an approved standard.

## Merge and push policy

- Feature-branch push requires explicit human permission after the Builder report.
- Merge requires green CI, resolved blocking findings, Review Agent evidence, and human approval.
- Merge method into `develop` is squash merge.
- Direct push and force push to `develop`/`main` are forbidden.
- No Builder self-approval.
- `main` receives release PRs from `develop` under the same or stricter controls.

## Target branch protection

### `develop`

- pull request required;
- required status checks;
- conversations resolved;
- force push and deletion disabled;
- squash merge only;
- Review Agent evidence plus human approval required.

### `main`

- pull request required;
- required status checks;
- signed/verified commits or platform-verified squash commits;
- force push and deletion disabled;
- two independent approvals when two eligible GitHub reviewers exist;
- release provenance and rollback evidence required.

GitHub can enforce only identities/checks it recognizes. Until a Review Agent submits a formal GitHub review or status check, its verdict is review evidence and the human GitHub approval is the enforceable gate. This limitation must not be misrepresented as automated enforcement.

## Post-merge controls

- Nightly: extended tests, dependency/security scans, financial vectors, migration rehearsal where applicable, and targeted performance baselines.
- Weekly: architecture drift, dependency growth, module coupling, Source of Truth consistency, unresolved risks, and MVP-scope audit.

## Scope rule

Every PR must answer: **Does this increase first-release user value or directly reduce a material correctness, security, privacy, or delivery risk?** If not, move it to `BACKLOG_V2.md`.
