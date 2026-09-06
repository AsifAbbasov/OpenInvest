# Stage 3.72 — Historical Governance Deviation Disposition

| Field | Value |
| --- | --- |
| Document type | Post-development governance / historical deviation disposition |
| Status | DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED |
| Date | 2026-09-06 |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Disposition ID | `STAGE-03-72-GOV-01` |
| Affected stage | Stage 3.72 — Portfolio Position Projection / Honest Market-Unavailable Semantics planning |
| Affected PR | `#142` |
| Historical compliance | `NONCOMPLIANT — PRE-EXTERNAL INTERNAL-VERDICT WITHHOLDING CONTROL MISSED` |
| Effective disposition status | `DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED` |
| Residual-governance-risk acceptance | `ACCEPTED — Principal Architect review 5126035577 on exact published head 29be318f4e6ffdd79a9d3f68d5c1c10ac0858f36` |
| Squash-merge authorization | `GRANTED AND CONSUMED — explicit Principal Architect authorization preserved by PR #143 comment 5560966901; PR #143 squash-merged at be654b938484d4ef4e4796b247bb04bf7efd2049` |
| Protected activation | `PR #143 / merge be654b938484d4ef4e4796b247bb04bf7efd2049 / tree d90ad717da04c7a28f538a4a0fea2781409dd44e / post-merge verification PASS comment 5560958299` |
| Stage 3.72 planning PR merge | `PR #142 remains DRAFT / UNMERGED — disposition blocker resolved; current-base refresh/reverification plus separate planning acceptance, Ready and squash-merge authorization remain required` |
| Stage 3.72 runtime authorization | `NONE` |
| Production/provider authorization | `NONE` |

## 1. Purpose and non-retroactivity

This record addresses one irreversible governance/process deviation in the Stage 3.72 planning lifecycle:

> the Draft PR body disclosed the current Internal review verdict `APPROVED` before the External
> published-head verdict, even though `docs/REVIEW_WORKFLOW.md` v1.4.0 requires the current Internal
> verdict and findings to remain withheld from the Draft PR/repository evidence surface until the
> External verdict.

The detailed Internal findings/evidence were not published before External review. The failure is narrower:
the verdict itself was exposed too early.

This record does **not** make the historical event compliant. It does not state or imply that:

- the verdict was actually withheld;
- the original pre-External review surface became compliant later;
- editing the PR body repaired the historical interval;
- fresh External review, CI, or later evidence publication substitutes for the missed temporal property;
- Stage 3.72 planning is accepted for merge;
- Stage 3.72 runtime implementation is authorized.

Every disposition activation gate has now succeeded through PR #143 squash merge `be654b938484d4ef4e4796b247bb04bf7efd2049` and post-merge verification. The exact named deviation therefore has the workflow-defined effective status:

```text
DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
```

This post-activation status does not make the historical event compliant, recreate the missed temporal withholding property, or authorize Stage 3.72 planning merge/runtime implementation. Historical pre-activation states remain preserved in their explicitly identified snapshots and chronology.

## 2. Exact repository, protected-base, and subject identity

Disposition branch identity at candidate preparation:

```text
PROTECTED_BRANCH=develop
DISPOSITION_BRANCH=docs/stage-03-72-governance-deviation-disposition
BASE_COMMIT=db589eb074468f352e61b82bbad8f33307b3fc89
BASE_TREE=e72fdd5a6577d81cb3b5413ec7c754cabb3bd7bc
BASE_COMMIT_VERIFICATION=VALID
BASE_BRANCH_PROTECTED=YES
```

Affected Stage 3.72 planning subject:

```text
PR=142
PR_STATE=OPEN_DRAFT
PR_MERGED=NO
PR_BASE=develop
PR_BASE_SHA=db589eb074468f352e61b82bbad8f33307b3fc89

INITIAL_PUBLISHED_HEAD=595877215528eb0d3b71c31076d6520e899d5182
INITIAL_PUBLISHED_TREE=1aca22110304ae3e472f67244df09cb02b2e329e
INITIAL_CI_RUN=387
INITIAL_CI_RUN_ID=34045426467

REMEDIATION_HEAD=9bc2229625a47df26244421002e7f45d93a96cea
REMEDIATION_TREE=c76e0e354c5df9db856caa7f70f8859ee2b2826f
REMEDIATION_CI_RUN=390
REMEDIATION_CI_RUN_ID=34045656905

EVIDENCE_HEAD=c4fbceb722946ceb364ba8c7f41e69848179fe7c
EVIDENCE_TREE=8a4ee1d183715fcb85f17fb03f8d00e1e5f05bd2
EVIDENCE_CI_RUN=399
EVIDENCE_CI_RUN_ID=34046213771

SUBJECT_MERGE_SHA=NONE
SUBJECT_PROTECTED_MERGE=NOT_PERFORMED
```

`SUBJECT_MERGE_SHA=NONE` is intentional and material. PR #142 remains open and Draft. This disposition
exists **before** any planning merge because the missed withholding property is already temporally
irreversible and must be resolved before the subject may proceed to Ready/merge.

## 3. Exact missed mandatory control

Canonical `docs/REVIEW_WORKFLOW.md` v1.4.0 requires the development path to publish the Draft PR with
current Internal evidence withheld from the PR/repository evidence surface until the External verdict.

The workflow further states:

```text
before the External verdict, the current Internal verdict/findings remain
WITHHELD from the Draft PR and repository evidence surface
```

The Stage 3.72 planning PR violated that control at Draft PR creation.

The initial PR body contained:

```text
Internal review result: APPROVED
```

while simultaneously stating that Internal evidence was withheld pending External review.

Therefore the historical proposition:

```text
the current Internal verdict was absent from the Draft PR/repository evidence surface
until the External verdict
```

is false.

The deviation is precisely:

```text
MISSED_CONTROL = PRE-EXTERNAL INTERNAL-VERDICT WITHHOLDING
MISSED_FINDINGS_WITHHOLDING = NO
MISSED_DETAILED_EVIDENCE_WITHHOLDING = NO
MISSED_EXTERNAL_REVIEW = NO
MISSED_CI = NO
```

The following controls are **not** dispositioned:

- complete prepublication Internal review;
- Builder remediation of Internal findings before publication;
- exact-head CI after Draft PR publication;
- fresh External published-head review;
- remediation of External findings;
- fresh CI on the remediation head;
- fresh External verdict on the remediation head;
- post-External full Internal evidence publication;
- exact-head CI after evidence publication;
- same-chat exact evidence-head no-semantic-drift verification;
- Principal Architect planning acceptance;
- Ready transition;
- squash-merge authorization;
- protected merge;
- Stage 3.72 runtime implementation authorization.

## 4. Immutable chronology

The relevant chronology is preserved as follows.

1. A complete prepublication four-file Stage 3.72 planning candidate was prepared against
   `develop@db589eb074468f352e61b82bbad8f33307b3fc89`.
2. Internal review found two prepublication issues:
   - `INT-372-F1` — undefined zero-denominator acquisition-basis allocation;
   - `INT-372-F2` — Stage 3.71 / Stage 3.72 authorization wording contradiction.
3. The Builder corrected both issues before publication. Internal verdict became `APPROVED`.
4. Human permission authorized commit/push of the four-file planning candidate and opening a Draft PR.
5. At `2026-09-06T16:26:04Z`, Draft PR `#142` was created. Its body disclosed:
   `Internal review result: APPROVED`.
6. At that moment, no External published-head verdict yet existed.
7. Initial published head:
   `595877215528eb0d3b71c31076d6520e899d5182`, tree
   `1aca22110304ae3e472f67244df09cb02b2e329e`.
8. CI `#387` / run `34045426467` completed successfully on that head.
9. At `2026-09-06T16:29:02Z`, External review `5125938134` returned `REQUEST CHANGES` and identified:
   - `EXT-372-P3-01` — published planning status still said PREPUBLICATION;
   - `EXT-372-P3-02` — Stage 3.71 current-state lifecycle wording remained stale.
10. The Builder remediated both findings without runtime/product/API scope expansion.
11. Remediation head became:
    `9bc2229625a47df26244421002e7f45d93a96cea`, tree
    `c76e0e354c5df9db856caa7f70f8859ee2b2826f`.
12. CI `#390` / run `34045656905` passed all 10 required contexts on that exact head.
13. At `2026-09-06T16:35:15Z`, fresh External published-head review `5125951620` returned `APPROVED`.
    It independently rechecked the published planning content and did not use the earlier Internal
    verdict/findings as supporting evidence.
14. Only after that External verdict, the full Internal findings/verdict evidence was published in an
    evidence-only follow-up.
15. Evidence head became:
    `c4fbceb722946ceb364ba8c7f41e69848179fe7c`, tree
    `8a4ee1d183715fcb85f17fb03f8d00e1e5f05bd2`.
16. The evidence record explicitly preserved `STAGE-03-72-GOV-01`,
    `HISTORICAL_COMPLIANCE=NONCOMPLIANT`, and `DISPOSITION_EFFECTIVE=NO`.
17. CI `#399` / run `34046213771` passed all 10 required contexts on exact evidence head `c4fbceb...`.
18. At `2026-09-06T16:47:18Z`, exact evidence-head review `5125981931` returned
    `APPROVED_WITH_EXISTING_GOVERNANCE_BLOCKER_PRESERVED`, confirming:
    - evidence-only exactness;
    - no financial/API/product semantic drift;
    - the historical deviation remains unresolved and blocks Ready/merge.
19. PR #142 remains open, Draft, unmerged, and blocked by this deviation.

No later action changes step 5: the Internal verdict was already visible before the first and final External
verdicts.

## 5. Why original temporal compliance cannot be recreated

The missed property is not merely “a line in the current PR body.” It is a historical condition over a
specific time interval:

```text
Draft PR publication
        ↓
External verdict
```

During that interval, the current Internal verdict was required to be absent from the Draft PR/repository
evidence surface.

That property cannot now be recreated because:

- the PR was already created with the verdict visible;
- the first External review already occurred after that disclosure;
- the final fresh External approval already occurred after that disclosure;
- editing the current PR body does not erase the earlier publication event;
- closing/reopening PR #142 does not make the earlier interval compliant;
- creating a replacement PR would not make PR #142's historical interval compliant and would fragment
  the preserved audit trail;
- repeating External review can add assurance but cannot prove that the verdict had been absent at the
  required historical time;
- rewriting or suppressing history would reduce forensic transparency and still would not constitute
  compliance.

Therefore this is an `otherwise temporally irreversible` governance deviation under
`docs/REVIEW_WORKFLOW.md` v1.4.0.

## 6. Why the open/unmerged subject does not make the missed control replayable

PR #142 is open and unmerged. That fact does **not** make this specific control still performable.

The exact control was:

```text
withhold the current Internal verdict until the External verdict
```

The relevant External verdict now already exists, and the verdict had already been disclosed before it.
There is no future action on this open PR that can make the historical pre-verdict interval satisfy the
withholding condition.

Still-performable controls remain fully live and are not waived:

```text
Principal Architect planning acceptance
Ready transition
planning PR squash-merge authorization
protected merge
Stage 3.72 runtime implementation authorization
```

Disposition applies only to the already-lost temporal withholding property. It cannot be used to bypass
any control that remains performable on PR #142.

## 7. Technical and review evidence that remains valid — and its limits

The current subject has substantial independent evidence that bounds technical/product risk.

### Initial published-head evidence

```text
HEAD=595877215528eb0d3b71c31076d6520e899d5182
TREE=1aca22110304ae3e472f67244df09cb02b2e329e
CI=#387 / 34045426467 / SUCCESS
EXTERNAL_REVIEW=5125938134 / REQUEST CHANGES
```

Materially, the first External review did not rubber-stamp the leaked Internal `APPROVED` verdict. It
independently found two documentation-state defects and required remediation.

### Remediation-head evidence

```text
HEAD=9bc2229625a47df26244421002e7f45d93a96cea
TREE=c76e0e354c5df9db856caa7f70f8859ee2b2826f
CI=#390 / 34045656905 / PASS_ALL_10_REQUIRED_CONTEXTS
EXTERNAL_REVIEW=5125951620 / APPROVED
```

That fresh External review independently rechecked:

- additive positions read-model boundary;
- authoritative Stage 3.71 WAC/acquisition-basis semantics;
- `acquisitionBasisWeight` zero-denominator/null behavior;
- deterministic BusinessDate/as-of semantics;
- subject isolation;
- explicit `UNAVAILABLE / NO_APPROVED_MARKET_PRICE_SOURCE`;
- no market-value fabrication;
- no provider activation;
- no executable OpenAPI/database/runtime drift.

### Evidence-head evidence

```text
HEAD=c4fbceb722946ceb364ba8c7f41e69848179fe7c
TREE=8a4ee1d183715fcb85f17fb03f8d00e1e5f05bd2
CI=#399 / 34046213771 / PASS_ALL_10_REQUIRED_CONTEXTS
EVIDENCE_VERIFICATION=5125981931
EVIDENCE_ONLY_EXACTNESS=PASS
NO_SEMANTIC_DRIFT=PASS
```

The evidence follow-up changes one planning file relative to the External-approved head and records
Internal evidence/governance chronology without changing frozen product/API/financial semantics.

### Limits of this evidence

None of the evidence above recreates the missed temporal withholding property.

In particular:

- a reviewer statement of non-reliance is not the same as proof that the verdict was unavailable;
- the fact that the first External review found two issues strongly bounds rubber-stamp risk but does not
  make the historical withholding control compliant;
- CI cannot verify after the fact that a PR body did not expose a verdict at an earlier time;
- the current planning content being technically sound does not substitute for the missed governance
  evidence property.

## 8. Residual governance risk

Residual risk is governance/evidence risk.

The lost assurance is:

> the External reviewer should have encountered a Draft PR/repository evidence surface from which the
> current Internal verdict and findings were withheld.

The Internal verdict was visible. Therefore an auditor cannot truthfully verify full repository/PR
withholding for the original Stage 3.72 External-review window.

Risk is bounded by:

- the first External review independently returning `REQUEST CHANGES` despite the leaked `APPROVED`;
- remediation of both External findings;
- fresh exact-head CI;
- fresh External `APPROVED` review after remediation;
- post-External full Internal evidence publication;
- exact evidence-head CI;
- exact no-semantic-drift verification;
- PR #142 remaining Draft and unmerged while the deviation is unresolved.

Residual governance risk remains non-zero because visibility itself cannot be undone.

At candidate preparation time:

```text
RESIDUAL_GOVERNANCE_RISK_ACCEPTANCE=PENDING
```

No Builder/model action may self-accept that risk.

## 9. Disposition eligibility — 7/7 test

`STAGE-03-72-GOV-01` satisfies the canonical eligibility test as follows.

| # | Requirement | Evidence | Result |
| ---: | --- | --- | --- |
| 1 | Missed item is governance/process only | Failure is timing of Internal-verdict publication; no runtime/product/security/privacy/financial/math/data-integrity/contract/migration defect is being waived | PASS |
| 2 | Governed action is immutable/merged or otherwise temporally irreversible | PR #142 is open, but the pre-External publication interval and both External review events have already occurred; the required historical absence property is irreversible | PASS |
| 3 | Replay cannot recreate original evidentiary property | A later edit, new review, replacement PR, or re-run cannot prove the verdict was absent before the already-completed External verdict | PASS |
| 4 | Original evidence and failed chronology are preserved append-only | PR #142 creation timestamp/body history, reviews `5125938134` / `5125951620`, evidence section, CI and exact-head review `5125981931` preserve the failure and later controls | PASS |
| 5 | Current technical state has sufficient independent evidence to bound residual risk | CI #390, fresh External `APPROVED`, CI #399, evidence-only verification, and the first External `REQUEST CHANGES` independently bound technical/review risk | PASS |
| 6 | The missed control is no longer performable on the open/unmerged subject | The only dispositioned control is “withhold verdict until External verdict”; the verdict was already disclosed and External verdict already occurred. Ready/merge/runtime gates remain performable and are explicitly not dispositioned | PASS |
| 7 | No narrower canonical remediation exists | Editing current metadata cannot restore the missed historical interval; abandoning/replacing the PR would not make the historical event compliant and would weaken traceability. Canonical disposition is the narrow mechanism | PASS |

Eligibility verdict:

```text
ELIGIBILITY=7_OF_7_PASS
DISPOSITION_ALLOWED=YES
RETROACTIVE_COMPLIANCE=FORBIDDEN
```

This verdict authorizes only continuation of the disposition lifecycle. It does not authorize commit,
push, risk acceptance, Ready, merge, branch deletion, Stage 3.72 runtime, provider activation, production
rollout, or later-stage work.

## 10. Affected and unaffected scope

Affected:

- Stage 3.72 planning governance/evidence chronology;
- auditability of the pre-External Internal-verdict withholding control on PR #142;
- PR #142 merge eligibility while this deviation remains unresolved.

Unaffected:

- Stage 3.71 runtime and lifecycle/documentation closure;
- ADR-009 WAC/acquisition-basis semantics;
- the Stage 3.72 planned additive endpoint/read-model semantics;
- quantity/WAC/acquisitionBasis financial vectors;
- zero-denominator `acquisitionBasisWeight = null` rule;
- BusinessDate/as-of semantics;
- subject isolation;
- market-unavailable semantics;
- OpenAPI executable contract, which Stage 3.72 planning has not yet changed;
- PostgreSQL schema/migrations;
- frontend runtime;
- provider/source activation;
- Feature 3D Corporate Actions source due diligence;
- XIRR/returns/tax/imported SELL/correction/reversal;
- production rollout;
- Stage 3.72 runtime implementation authorization.

This disposition does not make Stage 3.72 planning canonical by itself.

## 11. Compensating controls completed

The following controls already bound residual risk on the affected planning subject:

1. Exact-head CI on the initial published planning head.
2. Fresh External review `5125938134` that independently found two defects and returned `REQUEST CHANGES`.
3. Builder-only remediation of both External findings.
4. Exact-head CI #390 across all 10 required protected contexts.
5. Fresh External review `5125951620` on remediation head `9bc222...` with `APPROVED`.
6. Post-External publication of complete Internal findings/verdict evidence.
7. Explicit append-only preservation of `STAGE-03-72-GOV-01` and historical `NONCOMPLIANT` state.
8. Exact evidence-head CI #399 across all 10 required protected contexts.
9. Exact evidence-only verification `5125981931` confirming no financial/API/product semantic drift.
10. PR #142 kept Draft/unmerged after discovery of the deviation.

These controls reduce current risk but do not cure the historical noncompliance.

## 12. Recurrence-prevention controls

For future development-path PRs:

1. Before Draft PR publication, the PR/repository surface must use only an exact withholding marker such
   as:

   ```text
   Internal Review Evidence = WITHHELD — external published-head phase pending
   ```

   It must not expose `APPROVED`, `REQUEST CHANGES`, `BLOCKED`, Internal findings, or a summary that
   reveals the current Internal verdict.
2. The Builder must inspect the complete Draft PR body before creation and search specifically for
   `Internal`, `APPROVED`, `REQUEST CHANGES`, and `BLOCKED`.
3. If the current Internal verdict/findings are accidentally exposed before External verdict, the
   temporal deviation must be recorded immediately; editing the body is containment, not retroactive
   compliance.
4. The External phase must remain fresh and evidentiary and must not cite or inherit the leaked Internal
   material as support.
5. Full Internal evidence is published only after the External verdict, followed by exact-head CI and
   evidence-only verification.
6. No tooling is claimed to machine-enforce this temporal property today. Any automated enforcement
   requires a separately governed workflow/tooling change.

## 13. Dependent blocker and activation rule

This disposition must become effective **before** PR #142 may proceed to its final human planning
acceptance / Ready / squash-merge gates.

Before protected disposition merge:

```text
STAGE_03_72_PLANNING_PR=142
STAGE_03_72_PLANNING_PR_STATE=DRAFT_OPEN
STAGE_03_72_PLANNING_CONTENT_REVIEW=APPROVED
STAGE_03_72_GOV_01=UNRESOLVED_BLOCKER
HISTORICAL_COMPLIANCE=NONCOMPLIANT
DISPOSITION_EFFECTIVE=NO
RESIDUAL_GOVERNANCE_RISK_ACCEPTANCE=PENDING
DISPOSITION_MERGE_AUTHORIZATION=NOT_GRANTED
PLANNING_ACCEPTANCE=NOT_GRANTED
PLANNING_READY_AUTHORIZATION=NOT_GRANTED
PLANNING_MERGE_AUTHORIZATION=NOT_GRANTED
STAGE_03_72_IMPLEMENTATION_AUTHORIZATION=NOT_GRANTED
PRODUCTION_PROVIDER_AUTHORIZATION=NONE
LATER_STAGE_AUTHORIZATION=NONE
```

Required disposition sequence:

```text
this separate disposition candidate
→ deterministic checks
→ Governance / Closure review
→ findings remediated/checks rerun if needed
→ APPROVED prepublication review
→ separate human commit/push permission
→ disposition Draft PR
→ exact-head required CI green
→ exact-published-head Governance / Closure verification
→ explicit Principal Architect residual-governance-risk acceptance
   bound to STAGE-03-72-GOV-01 and exact disposition head
→ separate explicit squash-merge authorization
→ squash merge disposition PR to protected develop
→ post-merge verification
→ disposition becomes effective
```

Only after protected disposition activation may the named deviation become:

```text
DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
```

Even then, all of these remain separate:

```text
PR #142 planning acceptance
PR #142 Ready transition
PR #142 squash-merge authorization
Stage 3.72 runtime implementation authorization
```

Disposition merge must not be interpreted as granting any of them.

Because protected `develop` necessarily advances when the disposition is merged, PR #142 must then be
rechecked against the new protected base before any planning merge decision. At minimum:

```text
verify PR #142 remains conflict-free and exact-scope
verify the four-file planning diff has no semantic drift against the new protected base
verify the disposition record is present in the new protected base
rerun or re-establish required current-base CI evidence as required by the repository gate
perform a fresh exact-current-base no-drift verification
```

No pre-disposition CI/review evidence is allowed to silently stand in for this post-base-change gate.
The base refresh/reverification does not itself grant planning acceptance, Ready, merge, or runtime
authorization.

## 14. Rollback

Before protected merge, rollback is simple abandonment of the disposition branch/candidate. The
historical deviation remains unresolved.

After protected merge, the disposition is an append-only governance fact. Reverting the record would
not make the historical event compliant and would remove audit context. A future correction, if needed,
must preserve the original deviation, disposition, acceptance evidence, and correction chronology.

No rollback action may rewrite PR #142 history, erase the leaked-verdict chronology, or alter Stage 3.71
runtime/financial state.

## 15. Prepublication decision state

At candidate preparation time:

```text
DISPOSITION_ID=STAGE-03-72-GOV-01
ELIGIBILITY=7_OF_7_PASS
HISTORICAL_COMPLIANCE=NONCOMPLIANT
RETROACTIVE_COMPLIANCE_CLAIM=NONE
DISPOSITION_EFFECTIVE=NO
DEVIATION_STATE=UNRESOLVED_BLOCKER

GOVERNANCE_CLOSURE_REVIEW=PENDING
HUMAN_COMMIT_PUSH_PERMISSION=NOT_GRANTED
DRAFT_PR=NOT_OPENED
EXACT_HEAD_CI=NOT_RUN
EXACT_PUBLISHED_HEAD_VERIFICATION=NOT_RUN
RESIDUAL_GOVERNANCE_RISK_ACCEPTANCE=PENDING
SQUASH_MERGE_AUTHORIZATION=NOT_GRANTED

PR_142_READY_AUTHORIZATION=NOT_GRANTED
PR_142_MERGE_AUTHORIZATION=NOT_GRANTED
STAGE_03_72_IMPLEMENTATION_AUTHORIZATION=NOT_GRANTED
PRODUCTION_ROLLOUT_AUTHORIZATION=NONE
LATER_STAGE_AUTHORIZATION=NONE
```

This candidate is not effective and is not self-approving.

## 16. Protected activation and current state

The disposition lifecycle completed through protected squash merge of PR #143.

```text
DISPOSITION_ID=STAGE-03-72-GOV-01
FINAL_DISPOSITION_HEAD=29be318f4e6ffdd79a9d3f68d5c1c10ac0858f36
FINAL_DISPOSITION_TREE=d90ad717da04c7a28f538a4a0fea2781409dd44e
FINAL_DISPOSITION_CI=#401 / 34047566741 / PASS_ALL_10_REQUIRED_CONTEXTS
EXACT_PUBLISHED_HEAD_VERIFICATION=5126029859 / APPROVED
RESIDUAL_GOVERNANCE_RISK_ACCEPTANCE=5126035577 / ACCEPTED
MERGE_AUTHORIZATION_EVIDENCE=5560966901
PROTECTED_SQUASH_MERGE=be654b938484d4ef4e4796b247bb04bf7efd2049
MERGE_PARENT=db589eb074468f352e61b82bbad8f33307b3fc89
MERGED_TREE=d90ad717da04c7a28f538a4a0fea2781409dd44e
POST_MERGE_VERIFICATION=PASS / 5560958299
```

The merged tree is identical to the exact CI-green disposition head tree. The historical withholding failure remains permanently noncompliant; protected activation does not recreate the missed temporal property.

Current effective state:

```text
STAGE-03-72-GOV-01=DISPOSITIONED
HISTORICAL_COMPLIANCE=NONCOMPLIANT_PRESERVED
RESIDUAL_GOVERNANCE_RISK=ACCEPTED
DISPOSITION_EFFECTIVE=YES
RETROACTIVE_COMPLIANCE_CLAIM=NONE

PR_142_STATE=DRAFT_OPEN_UNMERGED
PR_142_DISPOSITION_BLOCKER=RESOLVED
PR_142_CURRENT_BASE_REFRESH_REVERIFICATION=REQUIRED
PR_142_PLANNING_ACCEPTANCE=NOT_GRANTED
PR_142_READY_AUTHORIZATION=NOT_GRANTED
PR_142_MERGE_AUTHORIZATION=NOT_GRANTED
STAGE_03_72_IMPLEMENTATION_AUTHORIZATION=NOT_GRANTED
PRODUCTION_PROVIDER_AUTHORIZATION=NONE
LATER_STAGE_AUTHORIZATION=NONE
```

Sections 8, 13 and 15 contain explicitly historical/pre-activation snapshots. Their pre-merge values remain preserved for chronology and are not current authority. This Section 16 and the active top table are the current-state authority for the disposition lifecycle.
