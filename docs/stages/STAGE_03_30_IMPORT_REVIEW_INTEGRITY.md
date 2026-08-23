# Stage 3.30 — Import Review Integrity

| Field | Value |
| --- | --- |
| Status | Complete / closed for P2-02/P2-03/P2-04 |
| Owner | Principal Architect |
| Baseline | `develop` at `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` |
| Branch | `fix/stage-03-30-import-review-integrity` |
| Implementation PR | #63 |
| Implementation merge | `8f68dd18800918e6a9882e995e13dba2723dc929` |
| Reviewed exact head | `2f788e0811d78c9def0502676a74bee2f9922bf5` |
| Exact-head CI | GitHub Actions #128 — SUCCESS |
| Independent final review | `APPROVED` |
| Human implementation merge authorization | 2026-08-23 |
| Closure PR | #64 |
| Trigger | Repository-audit P2-02, P2-03, P2-04 |
| Scope | Review-token semantic binding, parser row admission, targeted full-history reconciliation, HTTP/OpenAPI and regression coverage |
| Out of scope | Remaining P2/P3 findings, Stage 3.25 privacy work, product-scope expansion |

## Purpose

Stage 3.30 remediates three medium-severity import-review integrity defects that sit between transient
CSV interpretation, user approval, mutable ledger history, and the append transaction.

The design keeps the immutable/locked store as the final append authority while making the review
itself honest and bounded. It deliberately avoids making idempotent append retries depend on a
mutable pre-store review result.

## P2-02 — review token did not bind normalized financial meaning

### Observed defect

The previous HMAC token bound subject, portfolio, source account label, file hash, and row number/hash.
It did not bind the parser version, normalized candidate/fingerprint/status semantics, or expiry.

The same raw CSV and row hashes could therefore survive a deployment where parser semantics changed
and be appended under a meaning different from the one the user reviewed.

### Root cause

The token authenticated raw-source identity, not the interpretation contract. Review and append
reparsed the file, but there was no cryptographic assertion that normalized semantics remained the
same.

### Remediation

The review flow now has two phases:

1. a bounded parser review with no mutable ledger history;
2. a final review using only relevant full-history reconciliation rows.

The signed token is versioned and short-lived. It binds token/parser version, issue/expiry timestamps,
subject/portfolio/source label/file hash, all row identities, a stable parser-review semantic digest,
the final review semantic digest, and the exact row numbers that were APPENDABLE at review time.

Append reparses the same payload with no mutable ledger dependency and recomputes the parser semantic
digest. Any parser/candidate/fingerprint drift forces a fresh review. APPROVE is accepted only for a
row that the signed final review marked APPENDABLE.

### Why this method

The stable parser digest is recomputable during append without consulting mutable ledger state.
This matters for idempotency: after a successful append, the ledger necessarily changes. Requiring
the complete mutable-ledger final review to remain identical would make an idempotent retry fail
before the store can recognize the already-completed command.

The store still performs the final locked duplicate/identity check after idempotency reservation.

### Rejected alternatives

- Sign only raw file/row hashes: does not bind normalized financial meaning.
- Recompute the complete mutable-ledger final review at append and require equality: breaks safe
  idempotent retries after the first successful append changes ledger state.
- Persist server-side review sessions: adds state/cleanup/storage complexity for a transient 100-row
  workflow without being required for the current threat model.
- Put raw broker operation IDs in the token: unnecessary disclosure; the semantic digest uses the
  privacy-minimized broker-operation key.
- Non-expiring tokens: permits stale approvals to survive indefinitely across operational change.

### Trade-offs

`ReviewParserVersion` must be bumped whenever code changes can alter normalized import semantics. A
short token lifetime may require a fresh user review, which is preferable to applying stale financial
approval.

## P2-03 — the 100-row limit was not a computational limit

### Observed defect

`ReviewCSV` previously parsed until EOF and built all rows; HTTP rejected `TotalRows > 100` only after
parsing completed. The 2 MiB payload limit bounded bytes but did not bound review work to 100 rows.

### Root cause

The row-count invariant lived in the transport response path rather than the parser that performs the
work.

### Remediation

`ReviewCSV` now owns `MaxReviewRows = 100` and fails immediately when the 101st data row is read.
Malformed rows count as data rows. Every caller, including HTTP review and importflow append, inherits
the same computational admission bound. The HTTP count check remains defense in depth.

### Why this method

The CSV parser is the earliest authoritative layer that knows how many semantic records were consumed.

### Rejected alternatives

- Keep the HTTP post-parse check only: leaves the original computational gap.
- Infer row count from bytes/newlines: quoted CSV fields can contain newlines and rows are variable
  length.

## P2-04 — review used only the latest 100 ledger rows

### Observed defect

`/imports/review` loaded `ListTransactions(...Limit: 100)`. A duplicate/conflict older than that page
could be shown as APPENDABLE. The locked store could later reject it, but only after misleading user
approval.

### Root cause

A public recency-pagination primitive was reused as reconciliation history. Recency is not equivalent
to financial identity relevance.

### Remediation

Stage 3.30 parses the bounded batch first, derives a targeted reconciliation filter, and queries the
full ledger only for rows capable of changing review classification:

- all transactions on trade dates in the batch, covering legacy/manual exact and near-match checks;
- current-version imports for the same source-account label whose broker-operation key or normalized
  source fingerprint matches a batch identity key.

The query has no arbitrary recency limit but does not transfer the entire portfolio ledger. The final
review reruns the same parser against this targeted history.

### Why this method

The current financial identity model defines which existing rows can influence duplicate/conflict
classification. Querying exactly those dimensions is complete for current semantics without an
unbounded full-ledger application scan.

### Rejected alternatives

- Raise 100 to a larger page: moves the bug.
- Remove LIMIT and load the entire ledger: creates unbounded query/memory transfer.
- Rely only on append-time store revalidation: protects persistence but leaves review misleading.
- Feed mutable history into pre-store append re-review: can break idempotent retry before command
  reservation after a successful append.

### Trade-offs

The targeted-query dimensions must evolve together with reconciliation semantics and parser version.

## Regression evidence

- exactly 100 rows accepted; 101st fails during parse;
- semantic digest changes with normalized candidate/status changes;
- targeted filter contains only unique relevant dates and privacy-minimized identity keys;
- review token expires and rejects parser semantic drift;
- signed non-APPENDABLE rows cannot be approved;
- HTTP review uses targeted history, not public transaction pagination;
- over-limit review fails before any history query;
- PostgreSQL integration proves an old row omitted from latest-100 remains visible to targeted
  reconciliation;
- Stage 3.29 duplicate-header fail-closed behavior remains intact and now fails before history lookup.

## Verification and implementation merge evidence

- Local targeted importer/verticalslice/httpapi/postgres tests passed after correcting the stale-ledger
  regression to model the actual race as `review → concurrent ledger mutation → append`.
- `go vet ./...`, migration validation, OpenAPI validation, and full `go test ./...` passed locally.
- Extending `verticalslice.Store` exposed the development-only `unavailableStore` interface gap;
  it was completed with the same fail-closed `database url is not configured` behavior before commit.
- Implementation PR #63 was squash-merged into `develop` at `8f68dd18800918e6a9882e995e13dba2723dc929`.
- Final independently reviewed implementation head: `2f788e0811d78c9def0502676a74bee2f9922bf5`.
- Exact-head GitHub Actions CI #128 completed `SUCCESS`; all six workflow jobs passed.
- The Go CI job exported `OPENINVEST_DATABASE_TEST_URL`, so PostgreSQL integration tests executed
  against the service container rather than taking the no-database skip path.
- Independent final implementation review returned `APPROVED`.
- Explicit human authorization was received before the implementation squash merge.

## Residual boundaries

- Review tokens intentionally expire after 15 minutes; an expired token requires a fresh review.
- Stage 3.30 does not close P2-09 original-response idempotent replay or P2-13 browser idempotency
  persistence/recovery.
- The targeted-history dimensions are coupled to reconciliation semantics; future semantic changes
  must update the filter and bump `ReviewParserVersion` when normalized meaning changes.
- The locked PostgreSQL append path remains the final authority for races that occur after review.

## Canonical closure statement

Closure governance PR #64 passed exact-head CI #132 on `7d97f5f967074f98311adcd4b8f7962e0584c719`, received independent
closure `APPROVED` review and fresh explicit human squash-merge authorization, and was squash-merged
into `develop` at `ae6497050692798795efb85678af64db97cc5f53`.

Stage 3.30 is therefore CLOSED only for P2-02/P2-03/P2-04. At Stage 3.30 closure, the original audit
backlog contained 9 P2 and 10 P3 findings. Later remediation does not broaden or retroactively change
the Stage 3.30 closure scope. Stage 3.25 privacy Security Review evidence planning remains separate
and is not superseded.

## Scope boundary

P2-01/P2-09/P2-10/P2-11/P2-12/P2-13/P2-14/P2-16/P2-17 and all P3 findings remain separate.
Stage 3.25 privacy evidence planning remains separate.
