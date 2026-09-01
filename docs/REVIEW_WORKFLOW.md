# Review and Delivery Workflow

| Field | Value |
| --- | --- |
| Document ID | ENG-REVIEW-001 |
| Version | 1.4.0 |
| Status | Approved / Mandatory |
| Owner | Principal Architect |
| Supersedes | Informal review and push process |
| Dependencies | Architecture Freeze v1.2; Document 40; Document 43 |
| Last Review Date | 2026-09-01 |
| Next Review Date | 2026-12-19 |

## Purpose

Separate implementation, automated verification, specialist review, human approval, and merge authority. No builder may approve its own architectural work, and no review bypass may write directly to protected branches.

This workflow has two review paths:

- **development path** — changes to runtime code, executable API contracts, database/schema/migrations,
  dependencies, CI/workflows, security/privacy behavior, financial/mathematical semantics, or other
  implementation-affecting artifacts;
- **post-development governance/closure path** — documentation-only governance, evidence, or closure
  synchronization that changes none of the development surfaces above.

All model review work is conducted in **one designated review chat**. For development-path changes,
`Internal` and `External` are two sequential review phases inside that same chat, not separate chats.
For eligible post-development governance/closure changes, the same chat performs one Governance /
Closure review phase, plus CI and explicit human merge authorization.

### Review-chat topology and workflow transition rule

The Principal Architect defines the review-channel topology: **one designated review chat** is used
for repository review work. `Internal`, `External`, and `Governance / Closure` are review phases,
not requirements for separate chats.

For development-path review, the Principal Architect explicitly accepts one substantive replacement:
the former **strict information-isolation / blind non-disclosure** property between Internal and
External reviewers is replaced by **same-chat phase independence**. Earlier Internal messages may be
visible in chat history. The External published-head phase must nevertheless perform a fresh
evidentiary review and MUST NOT cite, inherit, or rely on the Internal verdict/findings as supporting
evidence for its conclusion.

This accepted replacement does not waive any other required control: Internal evidence remains
withheld from the Draft PR/repository evidence surface until the External verdict; required evidence
publication/verification, CI, finding remediation, protected-branch restrictions, and explicit human
merge authorization remain mandatory.

`docs/REVIEW_WORKFLOW.md` is governance-critical. A proposed replacement version becomes canonical
only after it is squash-merged into protected `develop`; until then, the previous repository version
remains the canonical text. The proposed version MUST NOT use its own content to bypass substantive
review, CI, or human-authorization gates required for its adoption.

A direct Principal Architect process directive may clarify review logistics such as chat topology
before repository text is synchronized, because the Principal Architect owns the review process and
final risk/merge authority. Such a directive MUST be explicit, narrow, preserved as transition
evidence, and MUST NOT authorize commit, push, Ready, merge, protected-branch mutation, or finding
suppression by itself.

After a new workflow version becomes effective on protected `develop`, review actions performed
thereafter use that effective version according to the PR's actual current scope, including for PRs
that were already open but unmerged. Review events completed before adoption keep their original
historical meaning and are not retroactively reclassified.

## Roles

### Builder Engineer — Codex

Owns scoped implementation, refactoring, tests, terminal execution, builds, documentation, feature branches, local commits, Draft PR preparation, review fixes, and stage reports. The Builder proposes a push but never pushes or merges without the required approval.

### Automated CI

Runs deterministic lint, formatting, unit, integration, contract, security, build, and relevant performance checks. Full expensive performance, resilience, and architecture suites may run nightly rather than on every PR when the PR gate remains risk-appropriate.

### Internal Review Agent

For **development-path** changes, the designated review chat first performs an **Internal review
phase**: mandatory read-only, line-by-line review of every changed file after Builder checks and
before commit authorization. The reviewer is independent from the active Builder turn, must not edit
files, run auto-fixes, stage changes, create commits, or silently resolve findings. It produces
evidence and exactly one verdict: `APPROVED`, `REQUEST CHANGES`, or
`BLOCKED — insufficient evidence`.

### External Review Agent — ChatGPT

For **development-path** changes, the same designated review chat later performs an **External
published-head review phase** after Draft PR publication and CI. This is a fresh evidentiary
re-evaluation of the published diff/evidence and must not inherit, cite, or use the Internal-phase
verdict/findings as supporting evidence for its conclusion, even though those earlier messages remain
visible in the same chat. It addresses architecture/DDD/SOLID, API contracts, security/privacy,
performance/cost, mathematical correctness when relevant, scope/YAGNI, migrations, rollback, and
documentation.

Specialist Architecture, Security, Performance, API, and Mathematical reviews may be separate reports
or clearly separated sections in one review. A model's self-review is advisory; the human remains
accountable.

### Governance / Closure Review Agent

For an eligible **post-development governance/closure** change, the designated review chat performs one
independent read-only Governance / Closure review phase covering the complete changed-file set. No
second Internal/External phase is required for that path.

The governance/closure reviewer must verify full documentation/evidence consistency, exact repository
identity where relevant, CI evidence, lifecycle/activation semantics, audit arithmetic, scope, rollback,
and absence of development-surface changes. If the candidate touches a development surface or a
material finding requires implementation-affecting remediation, the change returns to the development
path and the separate Internal + External review sequence applies.

### Principal Architect / Human Reviewer

Owns final scope, product, architecture, risk acceptance, review verdict, and merge authorization. Only the human may declare the stage accepted for merge.

## Mandatory delivery sequence

### Development path

```text
Approved development scope
  → feature branch
  → implementation and documentation
  → local quality gates
  → Internal Review Agent line-by-line review (read-only)
  → Builder resolves findings
  → local quality gates rerun
  → Builder stage report and diff summary
  → human permission to commit/push feature branch
  → push feature branch
  → Draft PR targeting develop with current Internal evidence withheld from PR/repository publication
  → required GitHub CI green
  → External published-head review phase in the same designated review chat
     (fresh evidentiary review; Internal verdict/findings are not supporting evidence)
  → Builder fixes and CI rerun
  → External verdict
  → after External verdict, Builder publishes required Internal evidence in an evidence-only follow-up
  → required CI on the evidence-follow-up head
  → same designated review chat verifies the evidence-only publication for exactness/no semantic drift
  → human review and approval
  → squash merge to develop
```

### Post-development governance / closure path

This path is allowed only when the complete change is documentation/evidence/governance-only and does
not change any development surface defined in **Purpose**.

```text
Approved governance/closure scope
  → documentation branch
  → documentation/evidence candidate
  → local deterministic checks
  → Governance / Closure review in the designated review chat
  → Builder resolves findings and reruns affected checks
  → human permission to commit/push
  → push branch
  → Draft PR targeting develop
  → required GitHub CI green
  → synchronize live PR metadata/evidence to the actual published head and CI
  → same single review chat performs exact-published-head verification
  → Builder resolves any new material finding and repeats affected gates
  → human review and explicit merge approval
  → squash merge to develop
```

The same designated review chat performs both pre-publication governance/closure review and
exact-published-head verification. A no-new-finding final `APPROVED` verdict does not require another
repository commit
solely to embed that verdict. Any new material finding must still be remediated and preserved before
merge.

GitHub Actions cannot evaluate an unpushed branch. Therefore local gates run before Draft PR;
authoritative repository CI runs after the feature branch is pushed. This is not permission to push
to `develop` or `main`.

## Mandatory Internal Line-by-Line Review — development path

Every changed file must be reviewed in full after the Builder finishes the scoped change and runs
the first local quality gate. Sampling is insufficient. Generated, vendored, specification,
documentation, configuration, example, test, and migration files are not exempt; the reviewer may
adapt emphasis to the artifact but must account for every changed line.

For an eligible post-development governance/closure change, the Governance / Closure phase in the
designated review chat subsumes the full changed-file review. No second Internal/External phase is
required. Sampling is still insufficient.

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

When a **development-path** verdict is not `APPROVED`, only the Builder applies fixes. The Builder
reruns every affected check and sends the revised complete diff back to an Internal Review Agent. No
human permission to commit/push may be requested until the current complete diff has an `APPROVED`
Internal verdict.

For an eligible **post-development governance/closure** change, the same rule applies using the single
Governance / Closure Review Agent: only the Builder fixes findings, affected checks rerun, and
commit/push permission may be requested only after that review chat returns `APPROVED`.

The review record states files reviewed, verdict, blocking findings, resolved findings, remaining
non-blocking notes, and explicit confirmation that the reviewer made no edits.

On the development path, Internal review does not replace CI, independent Draft PR External review,
or human accountability. It is a pre-commit defect gate; External Draft PR review is the independent
published-commit gate.

## Review levels and independence

- Level A — automated build, lint, tests, coverage where applicable, OpenAPI, YAML, Markdown, and
  other deterministic checks;
- Level B — Internal review phase in the designated review chat for development-path changes;
- Level C — External published-head review phase in that same chat for development-path changes;
- Level G — one Governance / Closure review phase in the designated review chat for eligible
  post-development documentation/evidence/governance-only changes;
- Level D — human business/product judgment and final merge authorization.

For the **development path**, the designated review chat keeps the two phases logically and
evidentially distinct:

- the Internal phase reviews the complete pre-publication candidate;
- before the External verdict, the current Internal verdict/findings remain **withheld from the
  Draft PR and repository evidence surface**;
- because one chat is used, earlier Internal messages may remain visible in chat history; this is the
  explicitly accepted replacement for strict blind information-isolation, not a claim that blindness
  is preserved;
- the External phase MUST perform a fresh published-head review and MUST NOT cite, inherit, or rely on
  the Internal verdict/findings as supporting evidence;
- the Stage report keeps the `Internal Review Evidence` fields
  `WITHHELD — external published-head phase pending` until the External verdict exists;
- after the External verdict, the Builder publishes the required Internal evidence in a follow-up
  documentation/evidence commit;
- required CI runs on that evidence-follow-up head;
- the same designated review chat verifies that the published Internal evidence is complete, accurate,
  and evidence-only, with no semantic/runtime drift before human merge authorization.

This preserves evidentiary non-reliance, repository/PR withholding, post-External evidence
publication, CI, and verification while using one chat. It does **not** claim to preserve strict
pre-verdict information-isolation; that specific control is deliberately replaced by same-chat phase
independence under the Principal Architect decision above. Same-chat visibility is not permission to
treat an earlier phase as proof.

An evidence-publication verification that finds **no new material finding** may remain as live review
evidence and does not require another repository commit merely to embed its own verdict. Any new
material finding must still be remediated, permanently preserved, and the affected gates repeated.

The two-phase development sequence does **not** apply to an eligible post-development
governance/closure-only change. That path uses one Governance / Closure phase in the same designated
review chat and must still preserve all material findings/remediations, exact published-head evidence,
CI evidence, and human merge authority.


## Irreversible historical governance deviation disposition

This section applies only to a mandatory **governance/process** control that was historically missed
and whose original temporal evidence property cannot be recreated because the governed action is
already immutable/merged or otherwise temporally irreversible.

It does not permit retroactive compliance.

### Eligibility

Disposition is allowed only when all conditions below are independently verified:

1. the missed item is a governance/process control, not a runtime, product, security, privacy,
   financial, mathematical, data-integrity, contract or migration defect;
2. the governed action is already immutable/merged or otherwise temporally irreversible;
3. replaying the control now cannot recreate the evidentiary property it was required to establish at
   the original time;
4. all available original evidence and failed chronology are preserved append-only;
5. the subject's current technical state has sufficient independent evidence to bound residual risk;
6. the control is not still performable on an open/unmerged subject;
7. no narrower canonical remediation exists.

If any condition is false or uncertain, disposition is forbidden.

### Mandatory semantics

A disposition MUST preserve the historical deviation as noncompliant.

The only effective status is:

`DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED`

A disposition MUST NOT state or imply that:

- the missed control was performed;
- the historical event became compliant;
- the finding never existed;
- technical correctness substitutes for the missed governance evidence;
- the disposition closes unrelated findings.

### Required evidence

The disposition record MUST identify:

- stable deviation ID;
- affected stage, PR, exact published head(s), merge SHA and protected-base identity;
- canonical workflow version and exact missed mandatory control(s);
- immutable chronology;
- why original temporal compliance cannot be recreated;
- exact technical evidence that remains valid and its limits;
- residual governance risk;
- affected/unaffected audit and product scope;
- compensating and recurrence-prevention controls;
- dependent blockers and activation rule.

### Required disposition sequence

A disposition is a separate post-development governance action and uses the effective workflow version
already present on protected `develop`.

```text
separate disposition branch/record
  → local deterministic checks
  → Governance / Closure review
  → findings remediated and checks rerun
  → APPROVED prepublication review
  → separate human commit/push permission
  → Draft PR
  → required exact-head CI green
  → same designated review chat exact-published-head verification
  → explicit Principal Architect residual-governance-risk acceptance
     bound to the deviation ID and exact published disposition head
  → separate explicit squash-merge authorization
  → squash merge to protected develop
  → disposition becomes effective
```

Risk acceptance and merge authorization are distinct explicit human acts.

### Activation

Before the disposition record is squash-merged into protected `develop`, the deviation remains an
unresolved blocker.

Protected merge activates only the disposition of the exact named deviation. Historical noncompliance
remains preserved permanently.

### Prohibitions

This mechanism MUST NOT be used:

- prospectively to skip a control;
- to bypass a control that can still be performed;
- to self-bootstrap an amendment before the amendment is canonical;
- to waive red CI, unresolved runtime/security/privacy/data-integrity defects or protected-branch rules;
- by Builder self-approval;
- as implicit authorization for commit, push, Ready, merge, branch deletion or another finding.


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

- For development-path changes, feature-branch commit/push permission may be requested only after
  local checks and the mandatory Internal Review Agent verdict are `APPROVED`.
- For eligible post-development governance/closure changes, commit/push permission may be requested
  only after local deterministic checks and the single Governance / Closure Review Agent verdict are
  `APPROVED`.
- Development-path merge requires green CI, an approved Internal phase, an approved External
  published-head phase in the designated review chat, publication and verification of required
  Internal evidence after the External verdict, resolved blocking findings, and explicit human approval.
- Eligible post-development governance/closure merge requires green CI, an `APPROVED` exact-published-
  head Governance / Closure verdict from the designated review chat, resolved blocking findings, and
  explicit human approval.
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

<!-- OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1
CANONICAL_WORKFLOW_BEFORE_ACTIVATION=1.3.0
PROPOSED_WORKFLOW=1.4.0
AMENDMENT_STATUS=PROPOSAL_NOT_CANONICAL
ADOPTION_PATH=V1_3_POST_DEVELOPMENT_GOVERNANCE
SELF_BOOTSTRAP_NEW_RULES=FORBIDDEN
P2_GOV_01=UNRESOLVED_BLOCKER
P2_GOV_02_TO_05=REMEDIATED_IN_STAGE_03_51_V6_REVIEW
P3_07_STATE=OPEN
STAGE_03_51_PUBLICATION_ELIGIBILITY=BLOCKED
P3_08_STATE=OPEN_UNAFFECTED
CURRENT_AUDIT_CLOSED=30/32
CURRENT_AUDIT_PERCENT=93.75%
NEW_MECHANISM_AVAILABLE_AFTER=PROTECTED_DEVELOP_SQUASH_MERGE
NEXT_AFTER_AMENDMENT=SEPARATE_P2_GOV_01_DISPOSITION
DISPOSITION_STAGE=3.53
THEN=REVISE_AND_REREVIEW_STAGE_03_51
<!-- OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1_END -->
