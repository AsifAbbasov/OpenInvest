# Stage 3.51 — P3-07 Transaction Form Fixture / Default Semantics Closure and Forensic Record

| Field | Value |
| --- | --- |
| Status | COMPLETE / CANONICAL — P3-07 CLOSED through PR #113 squash merge `072350205b2746bdcd83f20718eb59efcd0478ef` |
| Date | 2026-09-01 |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Stage 3.49 plan | PR #109 squash merge `cfcc384a97327cc8b74aa05567b9629abf40a5fb` / plan blob `20921a00c6669fb0c39523e783c7015a7d016f80` |
| Stage 3.50 implementation | PR #110 / head `be774b3a8423ffba98633b257983856b2c990b95` / squash merge `915d42f614121959fface9846a07cc1b412febe2` |
| Stage 3.50 exact CI | #306 / run `33499393962` — 10/10 SUCCESS |
| Stage 3.53 prerequisite disposition | PR #112 squash merge `ea1f204eab47bf16566096722d6390557b8141af` — P2-GOV-01 disposition effective |
| Stage 3.51 published head | `1868749965fa1c113875592668afaa8f1f1ca35e` |
| Stage 3.51 published tree | `c83120dcfdf1cab281fea10818860827eb699b64` |
| Stage 3.51 exact CI | #309 / run `33540615693` — 10/10 SUCCESS |
| Stage 3.51 exact-published-head review | APPROVED |
| Stage 3.51 squash merge | `072350205b2746bdcd83f20718eb59efcd0478ef` |
| Final original audit | 31/32 = 96.875%; only P3-08 remains OPEN |
| Runtime change in this post-merge synchronization | None |
| Branch deletion | Not authorized by this record |

## 1. Finding and root cause

Original audit finding P3-07 identified fixture/default semantics in the production transaction form.

The production `AddTransactionForm` inherited plausible demonstration business values from the early
Stage 3.3 Web presentation slice. At the Stage 3.49 planning base, the initial editable values included
`SBER`, quantity `10.00000000`, unit price `280.00000000`, gross amount `2800.00000000`,
commission `2.80000000`, tax `0.00000000`, fixed historical trade/settlement dates and the note
`Stage 3.3 Web presentation slice`.

The defect was not that every initial UI state value is forbidden. The defect was that **user business
facts** were pre-populated with fixture data. `transactionType="BUY"` is a control default required to
select a render/applicability branch, and fixed RUB is an existing contract invariant.

Root cause: demonstration data from an early presentation slice survived into a production data-entry
surface without a formal taxonomy separating user-supplied business facts from UI control defaults and
contract invariants.

## 2. Failure scenario and project impact

A user can open the production form, overlook one or more plausible prefilled values and submit an
immutable ledger transaction containing stale example facts.

The fixed historical dates become more misleading with time, but replacing them with browser-local
"today" or a derived T+N settlement date would not solve the underlying problem: the client would still
manufacture a financial business fact that the user did not provide.

Impact was therefore correctness/UX debt with potential immutable bad-ledger input. The remediation
never claimed a demonstrated auth bypass, security exploit or backend-integrity bypass; the Go API
remained authoritative and continued to validate submitted requests.

## 3. Initial remediation direction and rejected alternatives

The first durable design decision was made in Stage 3.49:

- user business facts must start empty;
- `BUY` may remain as a UI control default;
- RUB remains the frozen contract invariant;
- placeholder text may guide entry but must never become a serialized value.

The following alternative defaults were explicitly rejected because they still invent business facts:

- current/browser-local date;
- T+N settlement date;
- last-used transaction values;
- portfolio-derived values;
- asset-search-derived values;
- automatically calculated or looked-up prices/amounts.

This was a design rejection, not a claim that a technical reviewer first approved and then rejected a
different runtime implementation. The runtime implementation itself was narrow and later passed its
technical published-head review.

## 4. Final technical design

Stage 3.50 changed exactly two frontend files:

1. `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx`
2. `frontend-next/tests/add-transaction-form.component.test.tsx`

Nine editable business initial values became empty strings:

- ticker;
- quantity;
- unit price;
- gross amount;
- commission;
- tax;
- trade date;
- settlement date;
- note.

The final design intentionally preserved:

- `transactionType="BUY"` as the control default;
- fixed `currency: "RUB"` behavior;
- ticker trim/uppercase normalization where applicable;
- nullable/applicability behavior;
- existing payload construction;
- idempotency intent/retry behavior;
- Unicode note validation;
- backend/database/OpenAPI authority and contracts.

No backend, database, migration, OpenAPI, dependency, financial-arithmetic, security/privacy or P3-08
runtime work was included.

## 5. Regression and adversarial evidence

The governed regression design required evidence that:

1. all nine business-value inputs start empty;
2. BUY remains the planned control default;
3. Stage 3.3 fixture values are no longer production initial values;
4. empty optional settlement date serializes as `null`;
5. empty normalized note serializes as `null`;
6. hidden values cannot leak into payload fields where the selected transaction type makes them
   inapplicable;
7. user-entered values, not placeholders, reach the request payload;
8. the existing idempotency path remains intact;
9. Unicode note validation remains intact.

A particularly important adversarial scenario was transaction-type switching. A value may remain in
React state after its field becomes hidden; the payload builder must still gate/null that field so stale
hidden state cannot leak into an inapplicable transaction payload.

Focused component-test blob:
`12df8f19344829772f7eea3412edd33e131257bc`.

Production form blob:
`8968fc9c5a91ba9d314c5f1fb29368793d6c6f61`.

## 6. Stage 3.49 planning evidence

Stage 3.49 / PR #109 froze the narrow semantics before implementation.

- planning merge: `cfcc384a97327cc8b74aa05567b9629abf40a5fb`;
- approved plan blob: `20921a00c6669fb0c39523e783c7015a7d016f80`;
- implementation remained frontend-only;
- P3-08 and Stage 3.25 remained explicitly out of scope.

The plan also rejected silent scope expansion into backend logic, migrations, OpenAPI changes,
automatic market-data lookup, frontend financial calculation or transaction-form redesign.

## 7. Stage 3.50 implementation evidence

Stage 3.50 was published as PR #110.

- exact published head: `be774b3a8423ffba98633b257983856b2c990b95`;
- exact published tree: `f3a77245ea06b1fddc25e80c83e50aeda2551447`;
- exact CI: #306 / run `33499393962` — 10/10 SUCCESS;
- technical published-head review: APPROVED;
- squash merge: `915d42f614121959fface9846a07cc1b412febe2`.

The technical behavior was therefore accepted and merged. The later difficulty was a governance
evidence-lifecycle problem, not a newly discovered runtime P3-07 defect.

## 8. Governance failure and review history

The governance history is intentionally separated from the technical remediation history. Nothing in
this section rewrites Stage 3.50 into historical compliance.

### 8.1 P2-GOV-01 — premature Internal evidence publication

PR #110 disclosed Internal approval in the Draft PR before the External verdict. Because the PR later
merged without an evidence-only follow-up head, the required temporal sequence could never be recreated:

- no later evidence-only head existed;
- no CI existed on such a head;
- no same-chat no-semantic-drift verification could occur on a head that never existed.

Threat to governance: a later reader could otherwise confuse technical correctness with proof that the
mandatory evidence chronology was followed.

Resolution: Stage 3.52 introduced the narrow non-retroactive v1.4.0 disposition mechanism; Stage 3.53
then dispositioned P2-GOV-01. Historical noncompliance remains permanently preserved.

`P2-GOV-01 = DISPOSITIONED — HISTORICAL NONCOMPLIANCE PRESERVED / RESIDUAL GOVERNANCE RISK ACCEPTED`

### 8.2 P2-GOV-02 — cross-surface current-state drift

Review found that canonical surfaces could disagree about whether Stage 3.51 was blocked, eligible or
closed.

Failure scenario: one high-priority document could state a different lifecycle state from another,
making the repository's current authority ambiguous.

Resolution: exact structured state was synchronized across canonical closure surfaces and identity
became a checked invariant.

### 8.3 P2-GOV-03 — machine-enforcement overclaim

Earlier wording claimed machine enforcement without sufficient reviewable executable proof.

Failure scenario: prose could claim that an invariant was mechanically guaranteed when no executable
checker demonstrated that guarantee.

Resolution: machine-enforcement claims were narrowed to executable checker behavior backed by explicit
positive and negative tests.

### 8.4 P2-GOV-04 — formatting-sensitive checker

A checker depended too strongly on prose formatting/line wrapping.

Failure scenario: a semantic contradiction could evade detection because equivalent prose was wrapped
or formatted differently.

Resolution: authoritative lifecycle state was represented by exact structured identity rather than by
formatting-sensitive prose matching.

### 8.5 P2-GOV-05 — duplicate/existential mutation gap

A checker could prove that an expected token existed without proving uniqueness or complete state
identity.

Failure scenario: a valid state block and a contradictory duplicate block could coexist, or one surface
could omit the authoritative block while another remained valid.

Resolution:

- exactly one authoritative block per governed surface;
- exact key/value maps;
- cross-surface byte identity;
- missing-block rejection;
- duplicate-block rejection;
- state-value mutation rejection.

The reviewed v6 controls retained 8/8 state-value mutations rejected, 4/4 missing-block cases rejected
and duplicate-block rejection.

### 8.6 P2-GOV-06 — stale high-priority current authority

The first final Stage 3.51 v7 candidate still contained stale high-priority SOT assertions that:

- P2-GOV-01 was still unresolved;
- Stage 3.53 was candidate-only;
- P2-GOV-01 remained an unresolved blocker;
- Stage 3.51 remained blocked.

Those statements contradicted the already protected-merged Stage 3.53 disposition.

Why the previous checker missed it: it validated structured state well but did not sufficiently scan the
human-readable high-priority current-authority region.

Resolution in v8:

- stale assertions removed;
- `CURRENT_AUTHORITY_SCAN=PASS`;
- `STALE_CURRENT_AUTHORITY=NONE`;
- four stale-authority mutations independently rejected;
- previous structured-state controls retained.

Repeat prepublication review returned APPROVED with no new material finding.

### 8.7 P2-GOV-07 — residual stale-current-authority after first post-merge forensic synchronization

The first post-merge forensic synchronization candidate passed its local structural checker, including
byte-identical post-merge state and 7/7 negative mutations, but independent prepublication review found
one additional material governance failure mode.

The candidate still allowed contradictory or ambiguously scoped legacy authority to survive outside the
new authoritative block:

- Stage 3.53 top-level metadata said `CANONICAL / EFFECTIVE` while `## 13. Current decision` still
  explicitly denied that the disposition was effective and described it as candidate-only;
- the SOT section heading still called Stage 3.51 a `closure candidate after Stage 3.53 disposition`
  despite the body correctly recording PR #113 and P3-07 CLOSED;
- legacy Stage 3.52/3.53 structured snapshots in SOT/ROADMAP still contained activation-time
  `P3_07_STATE=OPEN` / `CURRENT_AUDIT_CLOSED=30/32` values without an explicit machine-readable outer
  classification proving that those blocks were historical snapshots and not current authority.

Failure scenario: a human reader or a simple documentation parser can consume the lower stale
`Current decision` prose or a legacy `CURRENT_*` field and reach the opposite conclusion from the
actual protected repository state, even though the new post-merge block itself is correct.

Root cause: the first checker proved existence/identity of the new authoritative state and rejected the
seven mutations it knew about, but its stale-authority scan was not global enough. It did not require
historical classification of every surviving legacy state block and did not reject the stale Stage 3.53
current-decision sentence or stale SOT heading.

Remediation in this revised candidate:

- replace the entire Stage 3.53 `Current decision` section with effective post-merge facts;
- rename the SOT Stage 3.51 section to an explicit post-merge closure / forensic record;
- preserve Stage 3.52/3.53 legacy structured blocks byte-for-byte but wrap each SOT/ROADMAP copy in an
  explicit `HISTORICAL_ACTIVATION_TIME_SNAPSHOT` boundary with `CURRENT_AUTHORITY=NO`;
- strengthen the checker to reject the stale Stage 3.53 sentence, stale SOT heading, missing historical
  wrappers, unscoped legacy snapshots, post-merge state mutation and duplicate authoritative blocks.

This finding concerns only the post-merge documentation synchronization. It does not reopen P3-07,
does not alter 31/32 = 96.875%, and does not change P3-08.

## 9. Final Stage 3.51 publication and closure evidence

The approved post-P2-GOV-06 candidate was published without semantic drift:

- PR: #113;
- base: `develop@ea1f204eab47bf16566096722d6390557b8141af`;
- base tree: `488cce09e87d42b3e5e03441336197bd10228c51`;
- exact published head: `1868749965fa1c113875592668afaa8f1f1ca35e`;
- exact published tree: `c83120dcfdf1cab281fea10818860827eb699b64`;
- commit count: 1;
- changed-file set: exactly four closure documentation surfaces;
- published blobs equaled the approved prepublication blobs;
- semantic drift: NONE;
- CI #309 / run `33540615693`: 10/10 SUCCESS on the exact published head;
- same designated reviewer exact-published-head verification: APPROVED;
- separate human Ready authorization: completed;
- separate exact-head squash-merge authorization: completed;
- protected squash merge: `072350205b2746bdcd83f20718eb59efcd0478ef`;
- protected `develop` tree after merge: `c83120dcfdf1cab281fea10818860827eb699b64`.

The prepublication reviewer ZIP SHA256 was
`9993ab6bdf5e869307ef27402952d2c9fa1c8fe487858319a6b48badec70c642`.

The exact-published-head verification ZIP SHA256 was
`59482d3662ee330a339f8e4941ed5b1125ff674414efed9969f7a0779d56d6f8`.

## 10. Why the final design was accepted

The final design was accepted because it closes the original user-facing defect with the smallest
runtime surface while preserving existing contracts and making the governance history explicit rather
than rewriting it.

Technically:

- fixture business facts are gone;
- no replacement inferred business facts were introduced;
- field applicability prevents hidden stale-value leakage;
- existing validation/idempotency/Unicode/backend authority remains intact;
- exact implementation CI and review evidence is preserved.

Governance-wise:

- the irreversible Stage 3.50 lifecycle deviation is preserved rather than falsely repaired;
- the blocking effect was separately dispositioned under a workflow that did not self-bootstrap;
- structured state has mutation/absence/duplicate defenses;
- human-readable current authority received a dedicated stale-authority scan;
- the final published head had exact-head CI and exact-published-head independent verification.

## 11. Residual limitations and explicit non-claims

P3-07 closure does **not** claim or implement any of the following:

- BUY is still an intentional UI control default; the remediation does not redesign transaction-type UX;
- RUB remains a fixed MVP contract invariant and is not made editable;
- the client does not auto-derive trade date, settlement date, price, amount or other business facts;
- backend validation remains authoritative;
- no broader transaction-form calculation or workflow redesign is included;
- P3-08 migration-validator policy hardening remains OPEN and unaffected;
- Stage 3.25 privacy/security evidence work remains separate;
- Stage 3.50 remains historically noncompliant with the v1.3.0 evidence lifecycle;
- Stage 3.53 disposition does not retroactively make Stage 3.50 compliant;
- branch deletion is not authorized by this record.

These are intentional boundaries, not hidden closure claims.

## 12. Final canonical state

After PR #113 squash merge:

- P3-07 = CLOSED;
- original audit = 31/32 CLOSED = 96.875%;
- P0 = 0;
- P1 = 0;
- P2 = 0;
- remaining original finding = P3-08 only;
- P3-08 = OPEN / UNAFFECTED;
- runtime change from this post-merge documentation synchronization = NONE.

This document is the single canonical forensic narrative for P3-07. Stage 3.49 remains the immutable
planning record, Stage 3.50 remains the implementation evidence record, and Stages 3.52/3.53 remain the
governance-amendment/disposition records.

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
