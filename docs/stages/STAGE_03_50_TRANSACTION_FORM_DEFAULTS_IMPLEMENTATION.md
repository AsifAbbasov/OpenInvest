# Stage 3.50 — P3-07 Transaction Form Fixture / Default Semantics Implementation

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION MERGED — P3-07 later CLOSED through Stage 3.51 / PR #113; this record remains the immutable implementation-evidence record |
| Date | 2026-09-01 |
| Planning prerequisite | Stage 3.49 / PR #109 squash merge `cfcc384a97327cc8b74aa05567b9629abf40a5fb` |
| Approved plan blob | `20921a00c6669fb0c39523e783c7015a7d016f80` |
| Implementation PR | #110 |
| Published implementation head | `be774b3a8423ffba98633b257983856b2c990b95` |
| Published implementation tree | `f3a77245ea06b1fddc25e80c83e50aeda2551447` |
| Implementation squash merge | `915d42f614121959fface9846a07cc1b412febe2` |
| Exact required CI | #306 / run `33499393962` — 10/10 SUCCESS on exact published head |
| Production form blob | `8968fc9c5a91ba9d314c5f1fb29368793d6c6f61` |
| Focused component-test blob | `12df8f19344829772f7eea3412edd33e131257bc` |
| Historical governance disposition | P2-GOV-01 dispositioned through Stage 3.53 / PR #112 merge `ea1f204eab47bf16566096722d6390557b8141af`; historical noncompliance remains preserved |
| Final P3-07 closure | Stage 3.51 / PR #113 exact head `1868749965fa1c113875592668afaa8f1f1ca35e`, CI #309 / run `33540615693` 10/10, exact-published-head APPROVED, squash merge `072350205b2746bdcd83f20718eb59efcd0478ef` |
| Current original audit | 31/32 closed = 96.875%; only P3-08 remains OPEN |
| Original audit before closure | 30/32 closed = 93.75%; P3-07 and P3-08 remain OPEN |
| Runtime change in this record | None — this file records the already-merged Stage 3.50 implementation |

## 1. Scope actually implemented

Stage 3.50 implemented the Stage 3.49-approved narrow P3-07 change in exactly two frontend files:

1. `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx`
2. `frontend-next/tests/add-transaction-form.component.test.tsx`

The production change removed fixture/business data from the nine initial editable values:

- ticker;
- quantity;
- unit price;
- gross amount;
- commission;
- tax;
- trade date;
- settlement date;
- note.

Those fields start empty. `transactionType="BUY"` remains the intentional default. RUB remains the
existing currency invariant.

No today/browser-derived date, T+N date, last-used value, portfolio-derived value, or hidden stale
business value is introduced.

## 2. Preserved semantics

The implementation intentionally preserved:

- nullable/applicability behavior;
- existing payload construction;
- trim/uppercase semantics;
- idempotency behavior;
- Unicode behavior;
- transaction-type visibility/applicability;
- server-owned persistence;
- existing backend, database and OpenAPI contracts.

No backend, database, migration, OpenAPI, dependency, security/privacy or P3-08 work was included.

## 3. Exact technical evidence

Stage 3.50 was published as PR #110 from exact head `be774b3a8423ffba98633b257983856b2c990b95`.

The exact reviewed blobs were:

- form: `8968fc9c5a91ba9d314c5f1fb29368793d6c6f61`;
- focused component test: `12df8f19344829772f7eea3412edd33e131257bc`.

GitHub Actions CI #306 / run `33499393962` completed SUCCESS on that exact head with all ten required
jobs successful.

Technical published-head review returned APPROVED. The implementation was squash-merged at
`915d42f614121959fface9846a07cc1b412febe2` and the reviewed tree was preserved.

## 4. Historical governance deviation

The implementation's technical evidence remains valid, but the Stage 3.50 evidence lifecycle was
historically noncompliant with REVIEW_WORKFLOW v1.3.0.

The exact historical defect is P2-GOV-01:

- Internal approval was disclosed in the Draft PR before the External verdict;
- no later evidence-only head was created;
- no CI existed on such an evidence-only head;
- no same-chat no-semantic-drift verification could occur on a head that never existed.

Stage 3.53 did not rewrite those facts. Under canonical REVIEW_WORKFLOW v1.4.0, PR #112 protected-merged
the disposition record at `ea1f204eab47bf16566096722d6390557b8141af` after exact-head verification, explicit residual-governance-risk
acceptance and separate merge authorization.

Therefore:

`P2-GOV-01 = DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED`

That disposition resolves only the blocking governance effect. It does not itself close P3-07.

## 5. Closure activation outcome

Stage 3.51 / PR #113 completed the separately governed closure lifecycle on exact published head `1868749965fa1c113875592668afaa8f1f1ca35e`. CI #309 / run `33540615693` completed 10/10 SUCCESS, same designated reviewer exact-published-head verification returned APPROVED, separate Ready and squash-merge authorizations were given, and protected squash merge `072350205b2746bdcd83f20718eb59efcd0478ef` activated closure.

Current result:

- P3-07 = CLOSED;
- original audit = 31/32 = 96.875%;
- P3-08 = OPEN / unaffected;
- historical P2-GOV-01 noncompliance remains permanently preserved;
- this implementation record does not claim any P3-08 or Stage 3.25 work.

<!-- OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1
CANONICAL_WORKFLOW=1.4.0
CLOSURE_BASE=ea1f204eab47bf16566096722d6390557b8141af
CLOSURE_BASE_TREE=488cce09e87d42b3e5e03441336197bd10228c51
STAGE_03_49_PLAN_PR=109
STAGE_03_49_PLAN_MERGE_SHA=cfcc384a97327cc8b74aa05567b9629abf40a5fb
STAGE_03_50_IMPLEMENTATION_PR=110
STAGE_03_50_IMPLEMENTATION_HEAD=be774b3a8423ffba98633b257983856b2c990b95
STAGE_03_50_IMPLEMENTATION_MERGE_SHA=915d42f614121959fface9846a07cc1b412febe2
P2_GOV_01_STATUS=DISPOSITIONED_HISTORICAL_NONCOMPLIANCE_PRESERVED_RESIDUAL_GOVERNANCE_RISK_ACCEPTED
P2_GOV_01_DISPOSITION_PR=112
P2_GOV_01_DISPOSITION_MERGE_SHA=ea1f204eab47bf16566096722d6390557b8141af
HISTORICAL_COMPLIANCE=NONCOMPLIANT_PRESERVED
DISPOSITION_DOES_NOT_RETROACTIVELY_COMPLY=TRUE
P2_GOV_02_TO_05=REMEDIATED_AND_RETAINED
P2_GOV_06=REMEDIATED_AND_VERIFIED
CLOSURE_PR=113
CLOSURE_PUBLISHED_HEAD=1868749965fa1c113875592668afaa8f1f1ca35e
CLOSURE_PUBLISHED_TREE=c83120dcfdf1cab281fea10818860827eb699b64
CLOSURE_CI_RUN_NUMBER=309
CLOSURE_CI_RUN_ID=33540615693
CLOSURE_CI=10/10_SUCCESS
EXACT_PUBLISHED_HEAD_REVIEW=APPROVED
CLOSURE_MERGE_SHA=072350205b2746bdcd83f20718eb59efcd0478ef
CLOSURE_MERGE_TREE=c83120dcfdf1cab281fea10818860827eb699b64
P3_07_STATE=CLOSED
AUDIT_CLOSED=31/32
AUDIT_PERCENT=96.875%
REMAINING_ORIGINAL_FINDING=P3-08
P3_08_STATE=OPEN_UNAFFECTED
RUNTIME_CHANGE=NONE
FORENSIC_NARRATIVE=STAGE_03_51_DOCUMENT
BRANCH_DELETION_AUTHORIZED=NO
<!-- OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1_END -->
