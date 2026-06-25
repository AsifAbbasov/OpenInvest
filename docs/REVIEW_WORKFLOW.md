# Review and Delivery Workflow

| Field | Value |
| --- | --- |
| Document ID | ENG-REVIEW-001 |
| Version | 1.2.0 |
| Status | Approved / Mandatory |
| Owner | Principal Architect |
| Supersedes | Informal review and push process |
| Dependencies | Architecture Freeze v1.2; Document 40; Document 43 |
| Last Review Date | 2026-06-20 |
| Next Review Date | 2026-12-19 |

## Purpose

Separate implementation, automated verification, specialist review, human approval, and merge authority. No builder may approve its own architectural work, and no review bypass may write directly to protected branches.

## Roles

### Builder Engineer — Codex

Owns scoped implementation, refactoring, tests, terminal execution, builds, documentation, feature branches, local commits, Draft PR preparation, review fixes, and stage reports. The Builder proposes a push but never pushes or merges without the required approval.

### Automated CI

Runs deterministic lint, formatting, unit, integration, contract, security, build, and relevant performance checks. Full expensive performance, resilience, and architecture suites may run nightly rather than on every PR when the PR gate remains risk-appropriate.

### Internal Review Agent

Performs a mandatory read-only, line-by-line review of every changed file after Builder checks and
before commit authorization. The Internal Review Agent is independent from the active Builder turn,
must not edit files, run auto-fixes, stage changes, create commits, or silently resolve findings.
It produces evidence and exactly one verdict: `APPROVED`, `REQUEST CHANGES`, or
`BLOCKED — insufficient evidence`.

### External Review Agent — ChatGPT

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
  → Internal Review Agent line-by-line review (read-only)
  → Builder resolves findings
  → local quality gates rerun
  → Builder stage report and diff summary
  → human permission to commit/push feature branch
  → push feature branch
  → Draft PR targeting develop
  → required GitHub CI green
  → independent ChatGPT external review of PR diff without internal verdict disclosure
  → Builder fixes and CI rerun
  → ChatGPT approval/evidence
  → human review and approval
  → squash merge to develop
  → nightly verification
  → periodic architecture audit
```

GitHub Actions cannot evaluate an unpushed branch. Therefore local gates run before Draft PR; authoritative repository CI runs after the feature branch is pushed. This is not permission to push to `develop` or `main`.

## Mandatory Internal Line-by-Line Review

Every changed file must be reviewed in full after the Builder finishes the scoped change and runs
the first local quality gate. Sampling is insufficient. Generated, vendored, specification,
documentation, configuration, example, test, and migration files are not exempt; the reviewer may
adapt emphasis to the artifact but must account for every changed line.

The Internal Review Agent must check, where applicable:

- SOLID, including SRP, OCP, LSP, ISP, and DIP;
- DRY, KISS, YAGNI, and Law of Demeter;
- DDD ownership and bounded-context boundaries;
- API First compliance and OpenAPI consistency;
- security and Privacy by Design;
- performance and cost impact;
- testability and evidence quality;
- documentation and Source of Truth consistency;
- scope control and absence of unrelated work;
- absence of business logic forbidden by the currently approved stage.

The reviewer does not modify code or documentation. Each finding contains severity, file and line
where possible, violated principle/risk, impact, and a minimal recommendation. Its report ends with
exactly one verdict:

- `APPROVED` — no blocking finding remains;
- `REQUEST CHANGES` — one or more actionable blocking findings must be fixed;
- `BLOCKED — insufficient evidence` — a complete review cannot be supported by the available diff,
  files, checks, or requirements.

When the verdict is not `APPROVED`, only the Builder applies fixes. The Builder reruns every
affected check and sends the revised complete diff back to an Internal Review Agent. No human
permission to commit/push may be requested until the current complete diff has an `APPROVED`
internal verdict. The stage report records files reviewed, verdict, blocking findings, resolved
findings, remaining non-blocking notes, and explicit confirmation that the reviewer made no edits.

Internal review does not replace CI, independent Draft PR review, or human accountability. It is a
pre-commit defect gate; ChatGPT Draft PR review is the independent external gate on the published
commit and authoritative GitHub diff.

## Review levels and independence

- Level A — automated build, lint, tests, coverage where applicable, OpenAPI, YAML, Markdown, and
  other deterministic checks;
- Level B — read-only Internal Review Agent line-by-line review;
- Level C — independent external ChatGPT review of architecture, security, privacy, DDD, API,
  performance, cost, mathematics when relevant, and governance;
- Level D — human business/product judgment and final merge authorization.

The external reviewer receives the Draft PR, changed files, diff, governing documentation, and CI
evidence. Before the external verdict, internal findings and verdict remain in the private review
thread or another access-controlled, out-of-band review record and are not committed into the PR.
The Stage report retains the required `Internal Review Evidence` section but marks its fields
`WITHHELD — blind external review pending`. After ChatGPT records its independent verdict, the
Builder publishes the internal evidence in a follow-up documentation commit and both reviewers
verify that evidence-only change. This prevents anchoring while preserving the permanent audit
record; it does not hide code, requirements, tests, or repository history.

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

- Feature-branch commit/push permission may be requested only after local checks and the mandatory
  Internal Review Agent verdict are `APPROVED`.
- Merge requires green CI, an approved internal review, an approved independent ChatGPT Draft PR
  review, resolved blocking findings, and explicit human approval.
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
- Every fifth completed stage: a full repository line-by-line audit covering architecture, DDD,
  SOLID, API, security, privacy, performance, dependencies, tests, documentation, cost, and ADR
  consistency. Findings are resolved or explicitly accepted by the human before the next stage.

## Scope rule

Every PR must answer: **Does this increase first-release user value or directly reduce a material correctness, security, privacy, or delivery risk?** If not, move it to `BACKLOG_V2.md`.
