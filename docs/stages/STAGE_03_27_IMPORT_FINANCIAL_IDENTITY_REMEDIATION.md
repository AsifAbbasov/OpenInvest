# Stage 3.27 — Import Financial Identity and Cash-Flow Semantics Remediation

| Field | Value |
| --- | --- |
| Status | Complete / closed for P1-02, P1-03, and P1-04; implementation merged through PR #55; closure governance recorded through PR #58 |
| Baseline | `develop` at `213d1d9b4369a5e046b26c3a08990aa571603eaa` |
| Branch | `fix/stage-03-27-import-financial-identity` |
| Merge | PR #55; squash commit `6e8c806de857f844954f1db513487357dfe90187` |
Canonical record: commit(s) `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c`.
| Final CI | GitHub Actions #90 `SUCCESS` on `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c` |
| Closure Governance | PR #58 |
| Trigger | Repository audit P1 findings P1-02, P1-03, and P1-04 |
| Scope | Import identity/provenance, duplicate/conflict classification, cash-flow fee semantics, PostgreSQL constraints, OpenAPI contract, regression coverage |
| Out of scope | Authentication/session hardening, Argon2 availability controls, unrelated P2/P3 findings, product feature expansion |

## Purpose

Stage 3.27 remediates three high-severity financial-integrity findings identified during repository
audit. The defects shared a common boundary problem: import review semantics were richer than the
persisted ledger identity and cash-flow rules. The remediation therefore treats them as one coherent
financial-identity increment instead of three unrelated patches.

This document records the findings, root causes, selected remediation methods, rejected/avoided
alternatives, persistence and contract impact, regression evidence, and remaining verification gates.

## Findings summary

| Finding | Observed defect | Risk |
| --- | --- | --- |
| P1-02 | `broker_operation_id` participated in review-time identity but was dropped before ledger persistence. | Distinct broker operations with identical economics could collapse, while later imports could not reliably reconstruct the same broker identity. |
| P1-03 | Cash near-match identity used transaction type and date but omitted gross amount. | Two legitimate deposits or withdrawals on the same date but for different amounts could be classified as near-conflicts. |
| P1-04 | Deposits and withdrawals accepted and persisted `commission`/`tax`, while snapshot cash calculation ignored those fields. | The ledger could contain fee values whose effect on cash and derived financial state was undefined and inconsistent. |

## Root-cause analysis

### P1-02 — Review identity was not a persistence contract

The importer computed fingerprints with `broker_operation_id` and could use broker-operation scope
while reviewing a file. The append DTO and persisted `transaction_entries` schema, however, did not
carry an equivalent per-row import identity. Once a reviewed row crossed into the append/store layer,
that identity was lost. Existing-row reconciliation therefore had to fall back to economic fields
only.

The root cause was architectural, not a single missing assignment: identity existed as importer-local
logic instead of a versioned contract shared across importer, domain validation, store persistence,
and database uniqueness.

### P1-03 — Cash near-match reused an asset-oriented key

The near-match key was built from transaction type, date, ticker, and quantity. That shape works as a
coarse asset-transaction candidate key, but cash transactions have no ticker or quantity. For
`DEPOSIT` and `WITHDRAWAL` rows the key therefore reduced effectively to type plus date. Gross amount
was checked only as a differing field after rows had already been grouped as near matches.

The root cause was reuse of one reconciliation key across transaction classes whose meaningful
identity dimensions are different.

### P1-04 — Accepted schema exceeded defined financial semantics

The API/domain validation allowed non-negative commission and tax for cash flows, and PostgreSQL
persisted them. Snapshot cash logic, however, defined deposits as `+gross` and withdrawals as
`-gross` without applying those fields. No approved financial rule specified whether a cash-flow fee
should reduce the deposit, increase the withdrawal, become a separate ledger entry, or be represented
in another way.

The root cause was an overly permissive write contract in the absence of an approved economic model.

## Remediation decisions

### 1. Versioned persisted import identity

Stage 3.27 introduces a canonical import identity carried through the append boundary and persisted
with imported ledger entries. The identity contains:

- normalized `source_account_label` scope;
- a SHA-256 `source_broker_operation_key` derived from the broker operation identifier when present;
- a normalized `source_fingerprint` for the financial row;
- an explicit `source_identity_version` so future identity changes are migrations/contracts rather
  than silent algorithm changes.

The raw broker operation identifier is not persisted. It is converted to a deterministic SHA-256 key
for equality/uniqueness checks. This keeps stable broker-operation identity without unnecessarily
propagating the source identifier through stored ledger records.

PostgreSQL enforces the new imported identity boundary through Stage 3.27 constraints and partial
unique indexes. The store still serializes import append under the existing portfolio lock, while the
database provides a final concurrency-safe uniqueness boundary.

#### Why this method was chosen

A database-backed identity was selected because review-only deduplication cannot protect later imports,
process restarts, concurrent requests, or future store implementations. A versioned contract prevents
identity semantics from drifting independently between importer and persistence.

Hashing the broker identifier was preferred over storing it raw because the system needs stable
equality, not display or recovery of the original identifier. SHA-256 gives deterministic fixed-size
indexable material while minimizing propagation of source-specific identifiers.

A purely economic fingerprint was not used as the sole identity because two real broker operations may
legitimately have identical economics. When broker identity is available it is the stronger operation
identity; the normalized financial fingerprint remains the fallback and conflict evidence.

### 2. Amount-aware cash reconciliation

Cash near-match classification now includes gross amount as an identity dimension. Two cash
transactions of the same type on the same date but with different amounts are therefore independent
candidates rather than automatic near-conflicts.

Identical broker identity with changed financial content is treated fail-closed as an identity
conflict, including when conflicting rows appear in the same CSV. This distinguishes "same operation
seen again" from "same claimed operation identity with inconsistent economics."

#### Why this method was chosen

For a cash transaction, amount is a primary business dimension. Adding it to cash reconciliation is
more precise than weakening near-match checks globally, and it preserves conservative duplicate
handling for asset transactions.

### 3. Fail-closed cash-flow fee semantics

`DEPOSIT` and `WITHDRAWAL` now require zero `commission` and zero `tax` at importer/domain contract
boundaries, with matching OpenAPI restrictions and a PostgreSQL constraint.

The remediation deliberately does **not** invent a formula for applying those fees to cash snapshots.
If product requirements later need cash-transfer fees or withholding, that behavior must be introduced
as an explicitly approved financial model with its own ledger semantics, migration/contract changes,
and golden vectors.

#### Why this method was chosen

Rejecting unsupported economics is safer than silently storing values that derived-state calculations
ignore. It preserves ledger explainability and prevents the database from accepting a state for which
the product has no canonical accounting interpretation.

## Persistence and migration impact

Stage 3.27 adds migration pair:

- `000004_stage_03_27_import_financial_identity.up.sql`
- `000004_stage_03_27_import_financial_identity.down.sql`

The migration extends `investment.transaction_entries` with import identity metadata, adds constraints
for imported identity and cash-flow fee semantics, and adds partial uniqueness indexes for imported
broker/fingerprint identity.

Existing migrations remain immutable. Legacy rows are not assigned fabricated broker identifiers.
The new identity contract applies prospectively to Stage 3.27 imported entries while preserving
existing ledger history.

## Application-layer impact

The remediation updates the following boundaries:

- importer review and append planning;
- vertical-slice append models and validation;
- PostgreSQL duplicate/conflict detection and transaction insertion;
- OpenAPI cash transaction semantics and examples;
- unit/regression tests and PostgreSQL integration tests.

The change does not expose broker operation identifiers through public response DTOs and does not add
broker API integration or raw broker-file persistence.

## Regression and verification matrix

The Stage 3.27 test set covers the following cases:

| Case | Expected result |
| --- | --- |
| Same economics, broker operation A vs B | Both operations may persist as distinct broker operations. |
| Repeated broker operation in the same source scope | Classified/rejected as duplicate according to persisted identity. |
| Same broker identity submitted later | Existing persisted identity is detected. |
| Same broker identity with changed financial fields | Fail-closed identity conflict. |
| Existing fallback fingerprint `F`, then strong broker identity `B/F` | Fail closed; ambiguous identity-strength transition cannot auto-append. |
| Existing strong broker identity `B/F`, then fallback fingerprint `F` | Rejected; result is independent of import order. |
| Distinct strong broker identities `A/F` and `B/F` | Both may persist as independently identified broker operations. |
| Conflicting same broker identity inside one CSV | Fail-closed identity conflict before append. |
| Same-day deposits of 1,000 and 2,000 | Both are valid independent cash operations. |
| Non-zero deposit/withdrawal commission or tax | Rejected before persistence; database constraint provides defense in depth. |
| Concurrent append of the same broker identity | Database/store boundary permits only one canonical imported operation. |
| Migration up/down/reapply | Schema must apply, roll back, and reapply cleanly on disposable PostgreSQL. |
| OpenAPI canonical examples | Validator must match the Stage 3.27 normalized fingerprint algorithm. |

## Verification status

At the time this document was introduced, implementation and static diff checks were complete but
Stage 3.27 was **not closed**. Closure requires all of the following to be recorded as passing evidence:

1. migration static validator;
2. OpenAPI validator;
3. PostgreSQL 18 migration apply;
4. Stage 3.27 schema/constraint/index inspection;
5. full migration rollback and reapply;
6. targeted importer/verticalslice/PostgreSQL remediation tests;
7. full `go test ./...` with PostgreSQL integration tests enabled;
8. `go vet ./...`;
9. final `git diff --check`;
11. repository governance updates and explicit merge gate before commit/push.

The first runtime attempt successfully established Docker/PostgreSQL 18, Go 1.25, and migration
validation but exposed a stale OpenAPI import-example fingerprint. The example was corrected to the
new canonical Stage 3.27 fingerprint before the runtime suite was rerun. This is recorded as a caught
contract regression, not as a bypassed test.

## Alternatives not selected


## Final-review corrective cycle

Canonical record: PR #55; commit(s) `c6c3a4c91a108426448a2bc230873ab9e479a335`.

The correction makes mixed-strength identity fail closed at four boundaries: importer reconciliation against persisted rows, same-file review, vertical-slice batch validation, and PostgreSQL store/database enforcement. The database guard takes a transaction-scoped advisory lock derived from portfolio, source-account scope, identity version, and financial fingerprint before checking for a fallback/strong collision, preventing concurrent direct inserts from racing through the check. Distinct non-null broker-operation identities with the same financial fingerprint remain allowed.


## Closure governance

- Final merge-candidate head `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c` passed GitHub Actions CI #90
  across frontend, PostgreSQL migration apply/rollback/reapply, Go tests,
  Docker Compose, Python, and OpenAPI gates.
- The earlier `changes required` on
  `c6c3a4c91a108426448a2bc230873ab9e479a335` remains historical evidence of
  the order-dependent fallback-to-strong identity defect; the approved head
  contains its correction and regression coverage.
- Explicit merge gate approved squash merge of PR #55.
- PR #55 was squash-merged into `develop` at `6e8c806de857f844954f1db513487357dfe90187`.
- PR #58 records the closure governance for Stage 3.27. Once this document is
  canonical on `develop`, P1-02/P1-03/P1-04 are closed.
- Closure scope is P1-02, P1-03, and P1-04 only.
- P1-01 and P1-05 remain unresolved separate audit-remediation items.
- Stage 3.25 remains separate proposal-only privacy evidence planning.

## Residual risks and next work

Stage 3.27 closes only P1-02, P1-03, and P1-04 after all closure gates pass. It does not address the
remaining P1 authentication/availability findings or the audit P2/P3 backlog.

The next high-priority remediation increment is expected to address refresh-token replay/session-family
revocation and bounded Argon2 resource consumption. That work must remain a separate stage so auth
security changes are independently reviewable from financial import identity.

## Closure rule

This document must not be changed to `Complete / closed` merely because the implementation exists or a
workflow. Until then, Stage 3.27 remains open.
