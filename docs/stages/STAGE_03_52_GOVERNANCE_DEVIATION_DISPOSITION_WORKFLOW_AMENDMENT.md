# Stage 3.52 — Governance Deviation Disposition Workflow Amendment

| Field | Value |
| --- | --- |
| Status | PROPOSAL ONLY — canonical `REVIEW_WORKFLOW.md` remains v1.3.0 until this amendment is squash-merged into protected `develop` |
| Date | 2026-09-01 |
| Canonical base | protected `develop@915d42f614121959fface9846a07cc1b412febe2` |
| Base tree | `f3a77245ea06b1fddc25e80c83e50aeda2551447` |
| Current workflow | `docs/REVIEW_WORKFLOW.md` v1.3.0 |
| Proposed workflow | v1.4.0 |
| Trigger | Stage 3.51 Governance / Closure review established that v1.3.0 has no retrospective disposition/acceptance mechanism for irreversible historical governance deviation `P2-GOV-01` |
| Development surface | None — documentation/governance only |
| Current original audit | 30/32 closed = 93.75%; P3-07 and P3-08 remain OPEN |
| Stage 3.51 status | BLOCKED / not publication-eligible while `P2-GOV-01` is unresolved |
| Self-bootstrap | Forbidden — this proposal cannot use its own v1.4.0 mechanism to disposition `P2-GOV-01` |
| Next after protected amendment activation | Separate Stage 3.53 `P2-GOV-01` disposition under effective v1.4.0 |
| P3-08 | Unaffected |

## 1. Why this amendment exists

Stage 3.50's runtime/UI remediation for original audit P3-07 was technically reviewed, passed exact-head
CI and was squash-merged. A later Stage 3.51 Governance / Closure review found a separate historical
process defect, `P2-GOV-01`: the development-path Internal evidence lifecycle required by v1.3.0 was
not followed before PR #110 merged.

The missed ordering cannot be recreated after merge. The same review established that canonical
v1.3.0 contains no explicit retrospective waiver, disposition or acceptance procedure for such an
irreversible historical governance deviation.

General Principal Architect risk ownership is not enough by itself because v1.3.0 also states that a
direct process directive cannot suppress a finding and that historical review events keep their
original meaning.

Therefore a canonical mechanism must be added before `P2-GOV-01` can be dispositioned.

## 2. Design goal

The amendment creates a narrow **irreversible historical governance deviation disposition** mechanism.

It does **not** make historical noncompliance compliant. It creates a controlled way to say:

- the deviation happened;
- it cannot be replayed or repaired in its original temporal order;
- exact evidence is preserved;
- technical/product correctness is evaluated separately;
- residual governance risk is explicitly reviewed and accepted by the Principal Architect;
- compensating controls are recorded;
- the deviation changes from `UNRESOLVED_BLOCKER` to `DISPOSITIONED — HISTORICAL NONCOMPLIANCE
  PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED`;
- dependent governance work may proceed only after the disposition record itself is activated on
  protected `develop`.

## 3. Non-goals

This mechanism MUST NOT:

- retroactively label a violated mandatory control as performed;
- erase, rewrite or downgrade historical findings;
- excuse an unmerged/open PR from a control that can still be performed;
- substitute for code/runtime/security/privacy/data-integrity remediation;
- waive a red CI check;
- authorize direct protected-branch writes;
- bypass current review, CI, Ready, merge or human-authorization gates;
- be used by this Stage 3.52 proposal before v1.4.0 becomes canonical;
- close P3-07 by itself;
- touch P3-08.

## 4. Eligibility

A deviation is eligible for disposition only if all are true:

1. a mandatory governance/process control was historically missed;
2. the governed action is already immutable/merged or otherwise temporally irreversible;
3. performing the original control now cannot recreate the evidence property the control was designed
   to establish at the required historical time;
4. the deviation is not a runtime/product/security/privacy/financial/data-integrity defect masquerading
   as process debt;
5. all available original evidence is preserved;
6. current technical state has sufficient independent evidence to assess what risk remains;
7. the same deviation is not still preventable on an open/unmerged subject;
8. no narrower canonical remediation exists.

If any item fails, disposition is forbidden and ordinary remediation remains required.

## 5. Required disposition evidence

A disposition record must include:

- stable deviation ID;
- affected stage, PR, head(s), merge SHA and protected-base identity;
- exact canonical workflow version violated;
- exact mandatory control(s) missed;
- immutable chronology showing what happened and when;
- why original temporal compliance cannot be recreated;
- technical evidence that remains valid and its exact limits;
- residual governance risk;
- affected and unaffected product/audit scope;
- compensating controls already implemented or required;
- recurrence-prevention controls;
- explicit statement that history remains noncompliant;
- current dependent blockers;
- exact activation rule.

Missing evidence blocks disposition.

## 6. Review and activation sequence

A disposition is a **separate post-development governance action** after this v1.4.0 mechanism is
already canonical.

Required sequence:

```text
effective REVIEW_WORKFLOW v1.4.0
  → separate disposition branch/record
  → local deterministic evidence checks
  → read-only Governance / Closure review
  → remediate findings and rerun checks
  → APPROVED prepublication disposition review
  → separate human permission to commit/push
  → Draft PR
  → required exact-head CI green
  → same designated review chat exact-published-head verification
  → explicit Principal Architect residual-governance-risk acceptance
     bound to the deviation ID and exact published disposition head
  → separate explicit squash-merge authorization
  → squash merge to protected develop
  → disposition becomes effective
```

The Principal Architect's residual-risk acceptance is not equivalent to merge authorization and neither
one implies the other.

## 7. Disposition semantics

An effective disposition uses the status:

`DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED`

It MUST NOT use `COMPLIANT`, `RETROACTIVELY COMPLIANT`, `CONTROL PERFORMED` or equivalent language.

Disposition resolves the **blocking effect** of the specified historical governance deviation only.
It does not erase the deviation from history and does not close unrelated audit findings.

## 8. Compensating-control requirements

Compensating controls must be specific to the missed evidence property.

For a missed review-evidence lifecycle, examples include:

- machine-enforced Draft-PR withholding of Internal evidence;
- an explicit `NEXT_GATE=EVIDENCE_ONLY_PUBLICATION` transition after External verdict;
- mandatory distinct evidence-only head;
- CI authority bound to the evidence head;
- same-chat exact evidence-publication verification;
- merge fail-closed on machine-readable lifecycle prerequisites;
- durable error ledger and executable regression tests.

A disposition cannot rely only on a promise to "be more careful."

## 9. Prohibition on amendment self-bootstrap

This Stage 3.52 proposal is governed entirely by **canonical v1.3.0** until protected activation.

Its proposed v1.4.0 disposition rules:

- cannot justify this amendment's own review path;
- cannot waive any v1.3.0 adoption requirement;
- cannot disposition `P2-GOV-01` before this amendment is squash-merged;
- cannot make Stage 3.51 publication-eligible before separate Stage 3.53 disposition activation.

## 10. Stage 3.51 / P3-07 relationship

Current state remains:

- original audit: 30/32 = 93.75%;
- P3-07: OPEN;
- P3-08: OPEN / unaffected;
- Stage 3.51: BLOCKED.

If this amendment is eventually protected-merged:

- v1.4.0 becomes canonical;
- `P2-GOV-01` is still unresolved;
- P3-07 remains OPEN;
- Stage 3.51 remains BLOCKED;
- next action is separate Stage 3.53 disposition.

Only after the Stage 3.53 disposition is itself protected-merged may Stage 3.51 be revised/re-reviewed.

## 11. Adoption path for this amendment

Because this candidate changes only documentation/governance and no development surface, it follows
the v1.3.0 post-development Governance / Closure path:

1. local deterministic checks;
2. same designated review chat prepublication Governance / Closure review;
3. separate human commit/push permission;
4. Draft PR;
5. required exact-head CI;
6. same designated review chat exact-published-head verification;
7. separate human Ready authorization if repository workflow requires Ready;
8. separate explicit squash-merge authorization;
9. protected merge activates v1.4.0.

Nothing in this proposal authorizes any remote mutation.

## 12. Decision

Propose `REVIEW_WORKFLOW.md` v1.4.0 with a narrow irreversible-historical-governance-deviation
disposition mechanism.

Before protected activation, v1.3.0 remains authoritative.

After protected activation, v1.4.0 only makes a future separate disposition **possible**. It does not
perform that disposition.

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
