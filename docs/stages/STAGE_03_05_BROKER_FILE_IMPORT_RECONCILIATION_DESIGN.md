# Stage 3.5 — Broker File Import and Reconciliation Design

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-05 |
| Version | 0.1.0 |
| Status | In progress / design only |
| Owner | Builder Engineer |
| Supersedes | Stage 3.5 roadmap placeholder |
| Dependencies | Stage 3.4; `product/MVP_PRODUCT_RISK_REFINEMENT.md`; ADR-003; ADR-006; Documents 42–43 |
| Last Review Date | 2026-07-02 |
| Next Review Date | Before Stage 3.6 implementation |

## Purpose

Stage 3.5 designs the smallest safe broker-file import path needed before public MVP.

The product risk is simple: manual transaction entry proved the architecture, but it is not enough
for users with real portfolio-accounting pain. A public MVP must let a user bring historical ledger
data into OpenInvest without credential scraping, silent mutation, or unreviewed automation.

This stage is documentation only. It does not authorize parser implementation.

## Non-goals

Forbidden in Stage 3.5:

- parser implementation;
- backend endpoints;
- frontend upload screens;
- SQL migrations;
- workers or collectors;
- direct broker API synchronization;
- credential collection or scraping;
- PDF parsing;
- automatic transaction mutation;
- external provider integrations;
- tax calculation;
- mobile implementation;
- Stage 3.6 implementation.

## Import principle

The import flow must preserve OpenInvest's canonical ledger model:

```text
User-supplied file
→ Parse
→ Normalize
→ Match
→ Detect duplicates
→ Detect conflicts
→ User review
→ Append only
→ Snapshot rebuild
```

No imported row may silently overwrite, delete, or rewrite an existing financial record.

## MVP import scope

The first import implementation candidate should support only:

- user-supplied broker files;
- CSV first;
- XLSX only if the dependency and file-safety model are approved in review;
- Russian ruble-denominated ordinary brokerage-account statements;
- transaction types already understood by the Stage 3 ledger model or explicitly mapped to a
  review-needed state;
- manual user confirmation before append.

Out of scope until later:

- direct broker API;
- credentials, tokens, passwords, or scraping;
- PDF statements;
- foreign securities;
- multi-currency;
- tax-lot optimization;
- automatic corporate-action inference;
- automatic fixing of existing ledger records.

## File trust model

Uploaded broker files are user-provided private data, not approved external data sources.

Rules:

- files are treated as untrusted input;
- no macros are executed;
- formulas are ignored or resolved only through safe library primitives if XLSX is later approved;
- spreadsheet-compatible previews, exports, diagnostics, and downloaded review files must neutralize
  formula-injection payloads;
- untrusted cell text beginning with `=`, `+`, `-`, `@`, tab, carriage return, or line feed must
  be escaped or rendered as inert text before export;
- original files are temporary by default;
- parsed rows must not be logged with personal or financial details;
- generated diagnostics must use safe row numbers and error codes;
- user can cancel the import before append;
- import preview must be exportable or inspectable before confirmation.

The Data Source Registry remains unchanged: a user-uploaded file is not a production external-source
approval for a broker, MOEX, CBR, Rosstat, or any provider.

## Minimal data model concepts

Stage 3.6 may need persistence for import sessions, but Stage 3.5 does not create tables.

Conceptual entities:

- `ImportSession`
  - user/subject boundary;
  - source kind: `USER_UPLOADED_FILE`;
  - file metadata: safe filename, size, media type, hash;
  - state: uploaded, parse_failed, parsed, validation_failed, normalized, review_required,
    cancelled, approved, append_in_progress, append_failed_retryable, append_failed_terminal,
    appended, rejected, expired;
  - created/updated timestamps.
- `ImportRow`
  - source row number;
  - persisted row hash, normalized candidate, and safe diagnostics only by default;
  - raw row text may exist only transiently in memory during parse/review rendering;
  - persistent raw row text storage requires a separate explicit retention approval and privacy review;
  - normalized candidate transaction;
  - validation status;
  - matching status;
  - duplicate/conflict reason.
- `ImportDecision`
  - user action: approve, ignore, map manually, reject;
  - decision timestamp;
  - append result when approved.

These are design concepts only. Any schema requires a separate Stage 3.6 migration PR and review.

## Normalized transaction candidate

Every parsed row should normalize into a candidate shape close to the Stage 2/3 transaction contract:

- transaction type;
- ticker, when applicable;
- quantity as decimal string;
- unit price as decimal string money;
- gross amount as decimal string money;
- commission as decimal string money;
- tax as decimal string money;
- trade date as BusinessDate;
- settlement date as nullable BusinessDate;
- broker operation identifier if present;
- source row number;
- source file hash;
- safe note.

Rules:

- binary float is forbidden;
- decimal parsing must preserve the source scale until normalized to 8 decimal places;
- ambiguous signs must become review-required, not guessed;
- dates are business dates, not UTC timestamps;
- unknown transaction types become review-required, not silently ignored.

## Matching and duplicate detection

The import engine should compute a deterministic candidate fingerprint from normalized business
fields:

```text
portfolio_id
transaction_type
ticker
quantity
unit_price
gross_amount
commission
tax
trade_date
settlement_date
broker_operation_id_or_empty
```

Duplicate detection levels:

1. Exact broker operation ID match, when present and scoped to the same authenticated subject,
   portfolio, user-selected source account or broker label, and source kind.
2. Exact normalized fingerprint match.
3. Near match requiring user review:
   - same date/ticker/type/quantity but different fee;
   - same cash amount/date/type but missing broker ID;
   - same row imported from same file hash.

Duplicate candidates must not be appended automatically.

Broker operation identifiers are untrusted user-file data. They may be used for matching scope, but
their presence does not approve or authenticate a broker, account, or external provider.

## Conflict detection

Conflicts require user review and must not append automatically.

Examples:

- row maps to a transaction type outside current implementation scope;
- ticker does not satisfy canonical ticker rules;
- amount sign conflicts with transaction type;
- quantity is zero or negative where not allowed;
- currency is not RUB;
- trade date is missing or invalid;
- settlement date precedes trade date when the source does not explicitly justify it;
- imported candidate appears to reverse or correct an existing transaction but lacks explicit user
  confirmation.

## User review workflow

The product UI for Stage 3.6 should present:

- file-level summary;
- row count;
- parsed rows;
- invalid rows;
- duplicate candidates;
- conflict candidates;
- appendable candidates;
- per-row decision controls;
- final confirmation.

The final append action must be explicit:

```text
Review
→ Confirm append
→ Append immutable transactions
→ Rebuild affected snapshots
→ Show import result
```

No automatic push, notification, declaration, tax export, or external submission is allowed.

## Failure, retry, and partial append semantics

Stage 3.6 must not introduce ambiguous import recovery behavior.

Normative rules:

- parse and validation failures are terminal for the affected raw file version until the user uploads
  a corrected file or changes mapping decisions;
- user cancellation leaves no ledger effect;
- cancellation, rejection, expiration, or terminal validation failure must delete transient raw row text;
- rejected rows leave no ledger effect;
- approved rows enter `append_in_progress` before ledger writes start;
- the append operation must be atomic for all approved rows in the session;
- if any approved row cannot be appended, the entire append operation must roll back and the session
  becomes `append_failed_retryable` or `append_failed_terminal`;
- `append_failed_retryable` may be retried only with the same approved row set and the same
  normalized fingerprints;
- retry must use an idempotency key derived from import session ID, approved row IDs, and normalized
  fingerprints;
- `append_failed_terminal` requires a new user review before any future append attempt;
- partial ledger append is forbidden in the MVP import path;
- snapshots rebuild only after the append transaction commits successfully;
- if snapshot rebuild fails after successful append, the immutable ledger remains committed and the
  snapshot rebuild must be retried idempotently from canonical transactions;
- import diagnostics must expose row numbers, safe error codes, and decision status, not raw private
  row contents.

## Security and privacy requirements

- imported files are private user data;
- default retention is temporary;
- persistent storage of original files requires separate approval;
- no file contents in logs;
- no passport, INN, address, phone, or tax profile is required;
- malware/macro execution is forbidden;
- parser errors must avoid leaking full row contents;
- import sessions must be scoped to the authenticated subject when authentication exists;
- anonymous financial history rules must apply after account deletion.

## Test-vector plan

Stage 3.6 must introduce test vectors before parser behavior is accepted.

Required vector categories:

- valid BUY;
- valid DEPOSIT;
- valid WITHDRAWAL;
- fee/commission normalization;
- duplicate exact match;
- near duplicate requiring review;
- invalid ticker;
- invalid date;
- non-RUB currency rejected or review-required;
- unknown transaction type review-required;
- amount sign ambiguity;
- malformed CSV row;
- spreadsheet formula-injection payloads in text fields and diagnostics;
- safe filename/media-type handling.

Test vectors should live under:

```text
tests/financial/import/
```

No production import parser should merge without these vectors or an explicit reviewed exception.

## Stage 3.6 candidate scope

If this design is approved, Stage 3.6 may implement the smallest vertical import slice:

- upload or local test fixture ingestion;
- CSV parse only;
- normalize into candidate transactions;
- duplicate/conflict detection;
- user-review representation through API or CLI test fixture;
- append only after explicit approval;
- snapshot rebuild after append;
- tests and examples.

Anything beyond this requires separate approval.

## Acceptance criteria

Stage 3.5 is complete when:

- import scope is explicitly limited;
- credential scraping and direct broker API are explicitly forbidden;
- append-only reconciliation is documented;
- duplicate and conflict rules are documented;
- privacy/security rules are documented;
- test-vector plan is documented;
- Stage 3.6 candidate scope is documented;
- governance registries point to this design;
- no implementation code, migrations, workers, or UI were added;
- strict review approves the design.
