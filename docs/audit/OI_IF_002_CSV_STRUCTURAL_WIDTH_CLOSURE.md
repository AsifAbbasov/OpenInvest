# OI-IF-002 — CSV Structural Width / Resource-Amplification Closure

Status: CLOSURE_CANDIDATE
Finding: OI_IF_002
Provisional finding: CANDIDATE-IMPORT-002
Severity: P2
Original evidence state: CONFIRMED

## 1. Finding identity

`OI_IF_002` is a security-relevant availability and resource-exhaustion input
boundary finding. It concerned attacker-controlled CSV structural cardinality
below the existing 2 MiB HTTP import-payload limit. No production outage was
demonstrated.

## 2. Original evidence

The initial Import Flow deep audit reviewed 65/65 files and 17,410/17,410
reviewable lines (100.0000%) and identified two confirmed P2 candidates. The
independent candidate review confirmed `CANDIDATE-IMPORT-002` as P2.

The audit reproduction accepted 90,012 CSV header columns in a 2,059,023-byte
payload, below the existing 2 MiB import limit. The active import semantics
also accepted arbitrarily wide data rows while consuming only canonical indexed
fields.

## 3. Root cause

Before remediation, CSV parsing used `FieldsPerRecord = -1`, read the complete
attacker-controlled header, and passed arbitrary header cardinality to
`mapColumns`. Row-count limits did not bound header or per-row field
cardinality. In addition, `hashRecord(record)` processed the complete data
record, so an over-wide data row could amplify downstream work.

## 4. Independent review correction

The initial remediation V1 added a header-cardinality bound, bumped the parser
from v2 to v3, and designed historical v1/v2 replay. Independent precommit
review returned `REQUEST_CHANGES`: it found that a wide data row remained an
active-parser bypass. R2 added the exact active-v3 12-field data-row structural
bound before downstream OpenInvest processing. Independent R2 precommit review
returned `APPROVE_FOR_COMMIT`.

## 5. Final remediation

The merged active parser defines:

```text
ReviewParserVersion=3
previousReviewParserVersion=2
legacyReviewParserVersion=1
MaxCSVHeaderColumns=12
MaxCSVHeaderFieldBytes=128
ExpectedCSVDataRowFields=12
ACTIVE_V3_SCHEMA_WIDTH_POLICY=EXACT_12_FIELDS
```

Active v3 rejects structurally excessive headers and data rows before
`reviewRow`, `hashRecord` for an over-wide row, candidate normalization,
history lookup, review-token issuance, append authorization, and material
store work. Go `encoding/csv` necessarily parses a raw record before this
application-level cardinality validation; this record does not claim zero
CSV-parser allocation.

## 6. Parser-version / historical replay compatibility

v3 is the only parser for fresh review and fresh append admission. v2 is
limited to exact historical completed-command reconstruction with strict
Decimal grammar and the prior structural semantics. v1 is limited to exact
historical completed-command reconstruction with legacy Decimal grammar and
the prior structural semantics. An old parser token cannot authorize a fresh
financial write. Exact completed-command replay remains available only where
the signed historic proof and replay identity match.

## 7. HTTP/store regression evidence

Stage 3.40 regression coverage proves the following for both an excessive
header and an excessive data row:

```text
review: 400 VALIDATION_ERROR; history/store calls=0
append: 400 VALIDATION_ERROR; material store calls=0
```

It additionally proves that a fresh v2 structural token is
`REJECTED_FOR_NEW_WRITE`, while a completed v2 exact replay is `PRESERVED`.

## 8. Technical PR and CI

Technical PR: #208
Technical source head: `f143575a72f081fcaa3f2548f2da6212e1d8d0b7`
Technical source tree: `1ecb3df225c580cbc1445594a1db236454ebe1dd`
Technical squash: `67ff5b3a328c54cc216e8944f4d3a6ef2dbba044`
Technical squash parent: `e16a710fa0fe2bc4cf77f4e14a5ecc60de44d5b4`

The PR had one commit and exactly five files (`+415/-18`). PR CI run #574
(`35960530346`, attempt 1) was `10/10 SUCCESS`. Independent PR/CI review
returned `APPROVE_FOR_MERGE`.

## 9. Post-merge verification

The squash tree is
`1ecb3df225c580cbc1445594a1db236454ebe1dd`, identical to the reviewed source
tree. Independent post-merge verification returned `PASS`. No workflow exists
for the final squash SHA (`POST_MERGE_WORKFLOW=NONE`); this record does not
claim post-merge CI success.

## 10. Scope exclusions

This record closes neither the Import Flow module nor the repository-wide
audit. It does not modify or authorize import rate limiting, import concurrency
control, Redis/Valkey, gateway, or distributed-limiter work. It does not alter
the historical Stage 3.16 remediation register, whose 32/32 findings are
already closed and which is a separate audit stream.

## 11. Remaining OI-IF-001 finding

`OI_IF_001` remains `OPEN_P2`. It is the separately confirmed finding for
missing server-side import rate/concurrency admission control. This
`OI_IF_002` record does not close or weaken `OI_IF_001`.

## 12. Closure activation

```text
PRE_MERGE_STATUS=CLOSURE_CANDIDATE
POST_MERGE_STATUS=CLOSED
CANONICAL_ACTIVATION=THIS_EXACT_CLOSURE_RECORD_ON_PROTECTED_DEVELOP_AFTER_REQUIRED_GATES
```

Before this closure-record PR is merged, `OI_IF_002` remains technically
remediated but formally open.
