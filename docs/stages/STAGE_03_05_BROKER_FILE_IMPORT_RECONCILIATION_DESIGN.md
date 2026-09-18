# Stage 3.5 — Broker File Import and Reconciliation Design

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-05 |
| Version | 0.1.1 |
| Status | Complete / closed |
| Supersedes | Stage 3.5 roadmap placeholder |
| Dependencies | Stage 3.4; `product/MVP_PRODUCT_RISK_REFINEMENT.md`; ADR-003; ADR-006; Documents 42–43 |

## Purpose

Stage 3.5 designs the smallest safe broker-file import path needed before public MVP.

The product risk is simple: manual transaction entry proved the architecture, but it is not enough
for users with real portfolio-accounting pain. A public MVP must let a user bring historical ledger
data into OpenInvest without credential scraping, silent mutation, or unreviewed automation.

This stage was documentation only and did not authorize parser implementation. It is now closed and
merged into `develop` at `072d38d94b529221d6467502f82f03a674a7d805`.

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


The Data Source Registry remains unchanged: a user-uploaded file is not a production external-source
approval for a broker, MOEX, CBR, Rosstat, or any provider.

## Minimal data model concepts

Stage 3.6 may need persistence for import sessions, but Stage 3.5 does not create tables.

Conceptual entities:


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

## Failure, retry, and partial append semantics

Stage 3.6 must not introduce ambiguous import recovery behavior.

Normative rules:


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


## Stage 3.6 candidate scope

If this design is approved, Stage 3.6 may implement the smallest vertical import slice:


Anything beyond this requires separate approval.

## Acceptance criteria

Stage 3.5 is complete when:
