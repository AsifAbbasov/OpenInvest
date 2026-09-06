# Stage 3.71 — Historical Governance Deviation Disposition

| Field | Value |
| --- | --- |
| Document type | Post-development governance / historical deviation disposition |
| Status | PUBLISHED DRAFT — NOT EFFECTIVE / UNRESOLVED BLOCKER |
| Date | 2026-09-06 |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Disposition ID | `STAGE-03-71-GOV-01` |
| Affected stage | Stage 3.71 — Portfolio Position & Cost Basis Engine |
| Affected PR | `#138` |
| Historical compliance | `NONCOMPLIANT — PRE-EXTERNAL EVIDENCE WITHHOLDING CONTROL MISSED` |
| Target disposition status after protected activation | `DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED` |
| Residual-governance-risk acceptance | `PENDING — explicit Principal Architect acceptance required on exact published disposition head` |
| Squash-merge authorization | `NOT GRANTED` |
| Production/runtime authorization | `NONE — this record is governance-only` |

## 1. Purpose and non-retroactivity

This record proposes the narrow disposition of one irreversible Stage 3.71 governance/process deviation:
required Internal/adversarial review evidence was published on the Draft PR/repository evidence surface
before the formal External published-head verdict.

This record does **not** make that historical event compliant. It does not state that the withholding
control was performed, that the deviation never existed, or that technical correctness substitutes for
the missing temporal evidence property.

The only allowed effective outcome, and only after every activation gate in Section 11 completes, is:

```text
DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
```

The Stage 3.71 runtime is already squash-merged into protected `develop`; this disposition changes no
runtime, financial, API, migration, frontend, security/privacy, dependency, CI/workflow, or provider
behavior.

## 2. Exact repository and historical identity

Current disposition base:

```text
PROTECTED_BRANCH=develop
DISPOSITION_BRANCH=docs/stage-03-71-governance-deviation-disposition
BASE_COMMIT=827f49f909ace5a3f7bcb2a3f51ce7c638c458ad
BASE_TREE=7c772c5fbaac63ea9fb28ca6f4088f6d2e981ab6
BASE_COMMIT_VERIFICATION=VALID
```

Historical Stage 3.71 subject:

```text
PR=138
HISTORICAL_DEVELOP_BASE=b772e52221fbb694b3116bd1b579db99d4e56302
IMPLEMENTATION_CODE_HEAD=6a1668fa9ecf6833f2d25fbdbe6b8bd9f5092e64
PREMATURE_EVIDENCE_COMMIT=86b0beaca8716cc3f9b75061890829aa92575cb3
EXTERNAL_REVIEWED_HEAD=2e18197e532671ca1c75803e0b4b025069c9c47e
POST_EXTERNAL_EVIDENCE_HEAD=6f66f7f3da982b2f83637452bfd1f556b92b3956
PROTECTED_SQUASH_MERGE=827f49f909ace5a3f7bcb2a3f51ce7c638c458ad
MERGED_TREE=7c772c5fbaac63ea9fb28ca6f4088f6d2e981ab6
```

The prematurely published evidence path was:

```text
docs/stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_REVIEW_EVIDENCE.md
```

The first publication commit added that review-evidence file to the already-open PR branch before the
formal External verdict.

## 3. Exact missed mandatory control

Canonical `docs/REVIEW_WORKFLOW.md` v1.4.0 requires, on the development path:

- the Draft PR to be published with current Internal evidence withheld from the PR/repository evidence
  surface until the External verdict; and
- the required Internal evidence to be published only after the External verdict, followed by fresh CI
  and exact evidence-only verification.

Stage 3.71 violated the first requirement.

The exact failure was not “missing review” and was not “missing CI.” It was an **evidence-timing /
withholding failure**:

```text
2026-09-06T10:17:13Z
  commit 86b0beaca8716cc3f9b75061890829aa92575cb3
  "docs: record Stage 3.71 adversarial review evidence"
  -> review evidence becomes repository/PR-visible

2026-09-06T10:48:43Z
  PR review 5125093621
  -> formal External published-head verdict recorded as APPROVED
```

Therefore the required temporal proposition “the Internal/adversarial evidence was not published on
the PR/repository surface before the External verdict” is false for Stage 3.71 and remains permanently
false.

The following controls were **not** missed and are not being dispositioned:

- Stage 3.71 implementation review and remediation;
- exact-head required CI before External review;
- fresh External published-head review;
- post-External evidence-only follow-up;
- fresh required CI on the evidence-only head;
- exact no-semantic-drift verification of that evidence follow-up;
- explicit human Ready authorization;
- explicit human squash-merge authorization;
- protected squash merge to `develop`.

## 4. Immutable chronology

The relevant immutable chronology is:

1. PR `#138` was opened against protected `develop@b772e52221fbb694b3116bd1b579db99d4e56302`.
2. Stage 3.71 implementation/remediation reached code head
   `6a1668fa9ecf6833f2d25fbdbe6b8bd9f5092e64`; CI `#375` completed successfully.
3. At `2026-09-06T10:17:13Z`, commit
   `86b0beaca8716cc3f9b75061890829aa92575cb3` published the adversarial review-evidence document on
   the PR branch before the formal External verdict.
4. A pre-External adversarial review comment was later preserved as review `5125058519`.
5. The published candidate reached exact head
   `2e18197e532671ca1c75803e0b4b025069c9c47e`; CI `#381` passed all 10 protected contexts.
6. After explicit human authorization, PR `#138` moved from Draft to Ready without changing that
   External-review subject.
7. At `2026-09-06T10:48:43Z`, formal External published-head review `5125093621` returned governance
   verdict `APPROVED` on exact head `2e18197e...`, explicitly stating that the earlier adversarial
   evidence was not used as supporting evidence.
8. Human permission was then given for the mandatory post-External evidence-only publication.
9. At `2026-09-06T10:51:07Z`, commit
   `6f66f7f3da982b2f83637452bfd1f556b92b3956` published the post-External evidence follow-up.
10. CI `#382` completed `SUCCESS` on exact evidence head `6f66f7f3...` across all 10 protected
    contexts.
11. Review `5125104746` verified that `2e181... -> 6f66f7f3...` was exactly one Markdown evidence file
    and found no semantic/runtime drift.
12. The Principal Architect explicitly authorized squash merge of PR `#138`.
13. At `2026-09-06T10:55:55Z`, PR `#138` was squash-merged into protected `develop` as
    `827f49f909ace5a3f7bcb2a3f51ce7c638c458ad`.
14. The protected squash commit tree `7c772c5f...` is identical to the final CI-green evidence-head
    tree, so the merge introduced no additional bytes.

No step above rewrites or erases the pre-External publication failure.

## 5. Why original temporal compliance cannot be recreated

The missed property is historical and temporal: the review evidence needed to be absent from the
published PR/repository surface **before** the External verdict.

That condition cannot now be recreated because:

- commit `86b0beac...` and the PR history already prove the evidence was published before the verdict;
- PR `#138` is already squash-merged and closed;
- deleting, reverting, renaming, or superseding the evidence now would not make the earlier
  publication disappear and would only reduce forensic transparency;
- repeating an External review now could add current assurance but cannot prove that the original
  evidence had been withheld at the required historical time;
- rewriting Git history or suppressing the preserved chronology would violate the repository's
  append-only governance/evidence intent and would not constitute compliance.

The correct canonical mechanism is therefore disposition of the historical process deviation, not
retroactive “repair.”

## 6. Technical evidence that remains valid — and its limits

Valid technical evidence includes:

- implementation/code head `6a1668fa...` passed CI `#375`;
- exact External-review head `2e18197e...` passed CI `#381` across all 10 protected contexts;
- formal External review `5125093621` returned `APPROVED` with no P0/P1 product-code blocker and
  stated that the earlier adversarial evidence was not used as supporting evidence;
- post-External evidence head `6f66f7f3...` passed CI `#382` across all 10 protected contexts;
- evidence verification `5125104746` confirmed an evidence-only one-file delta and
  `SEMANTIC_RUNTIME_DRIFT=NONE_FOUND`;
- protected squash merge `827f49f9...` is GitHub-verified and has the same tree
  `7c772c5f...` as the final CI-green evidence head.

This evidence bounds current technical risk but does **not** recreate the missed withholding property.
In particular:

- a reviewer statement of fresh non-reliance is not equivalent to proof that the repository evidence
  was unavailable before the verdict;
- technical CI cannot test a historical “was not published yet” condition after the fact;
- the External GitHub review state was `COMMENTED`; this disposition does not relabel it as a
  GitHub-native independent human approval.

## 7. Residual governance risk

Residual risk is governance/evidence risk, not a newly identified product/runtime defect.

The lost assurance is that the formal External phase encountered a repository/PR surface from which
the earlier Internal/adversarial evidence had been withheld. The External review explicitly stated
fresh evidentiary non-reliance, which reduces but does not erase the audit concern created by the
premature publication.

Therefore residual risk remains non-zero:

- historical review-surface independence is less strongly evidenced than the canonical workflow
  requires;
- an auditor cannot truthfully verify strict repository/PR withholding for the original External
  review window;
- the exact temporal control cannot be replayed on the already-merged subject.

The Principal Architect must explicitly accept this residual governance risk on the exact published
disposition head after CI and exact-published-head Governance / Closure verification. That acceptance
must not be inferred from permission to prepare, publish, mark Ready, or merge another artifact.

## 8. Disposition eligibility

`STAGE-03-71-GOV-01` satisfies the current workflow eligibility test only as follows:

| Requirement | Evidence | Result |
| --- | --- | --- |
| Missed item is governance/process only | The defect is pre-External review-evidence publication timing; no runtime/product/security/privacy/financial/data-integrity/contract/migration defect is being waived | PASS |
| Governed action is immutable/merged | PR `#138` is merged into protected `develop` at `827f49f9...` | PASS |
| Replay cannot recreate temporal property | Later deletion/review cannot make the evidence historically absent before the original verdict | PASS |
| Original evidence and failed chronology preserved | Commit `86b0beac...`, reviews `5125058519` / `5125093621`, evidence follow-up and PR history remain preserved | PASS |
| Current technical state has sufficient independent evidence to bound residual risk | CI `#381`, fresh External `APPROVED`, CI `#382`, exact evidence verification and tree equality | PASS |
| Missed control is no longer performable on the original open/unmerged subject | PR `#138` is closed/merged; the original pre-verdict interval cannot be reopened | PASS |
| No narrower canonical remediation exists | A normal docs correction cannot restore a missed historical withholding interval; disposition is the canonical mechanism | PASS |

Eligibility verdict:

```text
ELIGIBILITY=7/7_PASS
DISPOSITION_ALLOWED=YES
RETROACTIVE_COMPLIANCE=FORBIDDEN
```

This eligibility verdict authorizes only the disposition workflow. It does not self-accept residual
risk and does not authorize publication or merge.

## 9. Affected and unaffected scope

Affected:

- Stage 3.71 governance/evidence chronology;
- auditability of the pre-External withholding control;
- Stage 3.71 lifecycle closure status until this disposition becomes effective.

Unaffected:

- ADR-009 financial semantics;
- Decimal(28,8) / Half-Even WAC and acquisition-basis calculations;
- transaction-ledger ordering and oversell rejection;
- migrations `000008` / `000009`, owner-only population and readiness semantics;
- OpenAPI transaction-conflict semantics;
- Web manual SELL behavior;
- post-External CI and evidence-only verification;
- protected merge tree contents;
- privacy/security runtime behavior;
- external provider/source policy;
- the original 32-finding repository audit, already separately closed 32/32.

Current canonical `ROADMAP.md`, `SOURCE_OF_TRUTH.md`, and `DOCUMENT_INDEX.md` still contain
pre-activation Stage 3.70/3.71 wording. That is separate current-state documentation debt. This
disposition does not silently treat those files as synchronized and does not expand its own scope to a
Stage 3.71 closure rewrite. A separately reviewed documentation/closure synchronization remains
required before Stage 3.71 may be called fully lifecycle/documentation-closed.

## 10. Compensating and recurrence-prevention controls

Compensating controls already completed for the historical Stage 3.71 subject:

- fresh External review on exact published head `2e181...`;
- explicit External statement of fresh evidentiary non-reliance;
- post-External evidence-only publication;
- fresh CI `#382` on that evidence head;
- exact evidence-only/no-drift verification;
- exact tree equality between the final evidence head and protected squash merge.

Recurrence-prevention requirements for future development-path PRs:

1. Before Draft PR publication, the Builder must expose only the workflow-approved withholding marker,
   not Internal verdict/findings/evidence content.
2. Before the External verdict, any discovered repository/PR publication of Internal evidence is a
   governance blocker; it must not be treated as harmless because the reviewer promises not to rely on
   it.
3. The designated review chat must verify the published PR/repository evidence surface as part of the
   External phase and explicitly distinguish “visible but not relied upon” from the required
   repository-withholding property.
4. Required Internal evidence is published only after the External verdict, then receives fresh CI and
   exact no-semantic-drift verification.
5. No current tooling is claimed to machine-enforce this temporal property; these are mandatory
   procedural controls unless a separately governed enforcement mechanism is later adopted.

## 11. Activation rule and remaining gates

Creating this candidate, creating its branch, reviewing it, publishing a Draft PR, or obtaining green
CI does **not** disposition the deviation.

Before protected disposition activation:

```text
STAGE_03_71_RUNTIME=MERGED_IN_DEVELOP
STAGE_03_71_GOVERNANCE_CLOSURE=BLOCKED
STAGE_03_71_GOV_01=UNRESOLVED_BLOCKER
HISTORICAL_COMPLIANCE=NONCOMPLIANT
DISPOSITION_EFFECTIVE=NO
RESIDUAL_GOVERNANCE_RISK_ACCEPTANCE=PENDING
DISPOSITION_MERGE_AUTHORIZATION=NOT_GRANTED
PRODUCTION_ROLLOUT_AUTHORIZATION=NONE
LATER_STAGE_AUTHORIZATION=NONE
```

Required sequence from this candidate:

1. complete prepublication Governance / Closure review of the exact candidate;
2. remediate any finding and repeat affected deterministic checks/review;
3. obtain separate human permission to commit/push;
4. publish a Draft PR from the separate disposition branch to protected `develop`;
5. require all protected CI contexts green on the exact disposition head;
6. perform same-chat exact-published-head Governance / Closure verification;
7. obtain explicit Principal Architect residual-governance-risk acceptance bound to
   `STAGE-03-71-GOV-01` and that exact published head;
8. obtain a **separate** explicit squash-merge authorization;
9. squash merge the exact accepted disposition into protected `develop`;
10. verify protected activation.

Only after step 10 may the deviation state become:

```text
STAGE_03_71_GOV_01=DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED
HISTORICAL_COMPLIANCE=NONCOMPLIANT
DISPOSITION_EFFECTIVE=YES
```

Disposition activation resolves this specific historical governance blocker only. It does not itself:

- authorize production rollout;
- authorize branch deletion;
- authorize Stage 3.72 or any later feature;
- rewrite the Stage 3.71 historical evidence file;
- synchronize stale current-state wording in `ROADMAP.md`, `SOURCE_OF_TRUTH.md`, or
  `DOCUMENT_INDEX.md`;
- constitute a separate Stage 3.71 lifecycle/documentation closure.

## 12. Rollback and failure semantics

Before protected merge, rollback is to close/abandon the disposition PR and leave
`STAGE-03-71-GOV-01=UNRESOLVED_BLOCKER`.

After protected activation, reverting the disposition through the normal protected-branch PR process
does not erase the historical risk-acceptance event from Git history, but it removes the canonical
active disposition record. Any dependent Stage 3.71 closure assertion must then fail closed until a
valid canonical disposition is restored.

Any red CI, material Governance / Closure review finding, head drift, incomplete risk acceptance, or
missing separate merge authorization stops activation.

## 13. Prepublication decision state

At candidate preparation time:

```text
OPENINVEST_STAGE_03_71_GOV_01_DISPOSITION_V1
CANONICAL_WORKFLOW=REVIEW_WORKFLOW_V1_4_0
DISPOSITION_ID=STAGE-03-71-GOV-01
AFFECTED_STAGE=3.71
AFFECTED_PR=138
HISTORICAL_BASE=b772e52221fbb694b3116bd1b579db99d4e56302
PREMATURE_EVIDENCE_COMMIT=86b0beaca8716cc3f9b75061890829aa92575cb3
EXTERNAL_REVIEWED_HEAD=2e18197e532671ca1c75803e0b4b025069c9c47e
POST_EXTERNAL_EVIDENCE_HEAD=6f66f7f3da982b2f83637452bfd1f556b92b3956
HISTORICAL_MERGE=827f49f909ace5a3f7bcb2a3f51ce7c638c458ad
HISTORICAL_MERGE_TREE=7c772c5fbaac63ea9fb28ca6f4088f6d2e981ab6
MISSED_CONTROL=PRE_EXTERNAL_INTERNAL_EVIDENCE_REPOSITORY_WITHHOLDING
ELIGIBILITY=7_OF_7_PASS
HISTORICAL_COMPLIANCE=NONCOMPLIANT
DISPOSITION_STATUS=UNRESOLVED_BLOCKER
TARGET_STATUS=DISPOSITIONED_HISTORICAL_NONCOMPLIANCE_PRESERVED_RESIDUAL_GOVERNANCE_RISK_ACCEPTED
RESIDUAL_RISK_ACCEPTANCE=PENDING
MERGE_AUTHORIZATION=NOT_GRANTED
RUNTIME_CHANGE=NONE
PRODUCTION_ROLLOUT_AUTHORIZATION=NONE
LATER_STAGE_AUTHORIZATION=NONE
```
