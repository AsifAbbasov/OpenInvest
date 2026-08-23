# Stage 3.33 — Snapshot Rebuild Accuracy and PostgreSQL Runtime Immutability

| Field | Value |
| --- | --- |
| Status | Implementation candidate; exact-head CI and independent review pending |
| Owner | Principal Architect |
| Baseline | `develop` at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` |
| Branch | `fix/stage-03-33-snapshot-immutability` |
| Implementation PR | #69 |
| Trigger | Repository-audit P2-10, P2-11, and P2-12 |
| Scope | Exact import snapshot rebuild reporting, one-pass affected-snapshot rebuild planning, PostgreSQL application-runtime append-only privilege enforcement, startup privilege validation, regression/CI evidence |
| Out of scope | P2-16/P2-17, all P3 findings, Stage 3.25 privacy Security Review evidence work, provider credential provisioning, product-scope expansion, snapshot methodology redesign |

## Purpose

Stage 3.33 closes three repository-audit gaps without changing the financial methodology or product
surface. P2-10 and P2-11 concern the correctness and efficiency of the derived snapshot rebuild path.
P2-12 concerns the gap between the frozen append-only ledger architecture and the privileges actually
used by the running application.

The remediation keeps PostgreSQL as the authoritative transaction boundary. It does not introduce a
queue, Redis dependency, background worker, new microservice, or new financial feature.

## P2-10 — `snapshotDatesRebuilt` did not report the dates actually rebuilt

### Observed defect

The import response previously derived `snapshotDatesRebuilt` from the approved import rows' trade
dates. PostgreSQL used a different rule: for each imported trade date it rebuilt that date plus every
existing snapshot for the portfolio on or after that date under the active methodology.

A backdated import could therefore rebuild later historical snapshots while the HTTP response omitted
those dates. The response described requested input dates, not the database work that actually
committed.

### Remediation

Stage 3.33 moves affected-date ownership to the PostgreSQL boundary.

`planAffectedSnapshotDates` computes one deterministic ordered set:

`all distinct imported trade dates ∪ all existing active-methodology snapshot dates >= earliest imported trade date`

Every imported trade date remains present even when no snapshot existed before the command, preserving
the previous rule that the trade date itself receives a snapshot. Existing later snapshots are included
because their ledger prefix changes after a backdated financial entry.

The PostgreSQL append operation returns an `ImportAppendOutcome` containing both the inserted
transactions and the exact ordered `SnapshotDatesRebuilt` set. `importflow` no longer reconstructs or
guesses that list from its input requests.

For the replay-aware production path, the HTTP result is built from this exact outcome while the same
PostgreSQL transaction remains open. The Stage 3.32 atomic replay contract is preserved: financial
entries, snapshot rebuilds, the exact reported date list encoded in the response artifact, and command
completion commit together.

Legacy completed commands that do not contain an exact response artifact are not used to reconstruct a
new snapshot-date result from mutable projection state.

## P2-11 — one import batch could rebuild the same later snapshot repeatedly

### Observed defect

The old batch path deduplicated imported trade dates, but then called `rebuildAffectedSnapshots` once
for every distinct trade date. Each call cascaded across all existing snapshots on or after that date.

For example, if a batch contained 15 June and 25 June while existing snapshots included 20 June and
30 June, the 30 June snapshot could be rebuilt in both cascades. The final values remained derivable,
but the same command performed redundant projection writes and incremented snapshot versions more than
once for the same affected date.

### Remediation

The new planner computes the union before any snapshot is rebuilt. `rebuildSnapshotPlan` then walks the
sorted affected-date list exactly once. The portfolio row lock remains held for the financial command,
so same-portfolio financial writes cannot interleave between planning and rebuild execution.

The planner runs after the new ledger entries are inserted and before snapshot mutation. Every rebuilt
snapshot therefore reads the full command-local ledger state inside the same transaction.

### Regression proof

The PostgreSQL-backed Stage 3.33 regression creates baseline snapshots on 10, 20, and 30 June and then
imports backdated entries on 15 and 25 June.

The exact outcome must report:

- 15 June;
- 20 June;
- 25 June;
- 30 June.

The test then verifies snapshot versions. The existing 20 and 30 June snapshots advance exactly once,
while newly introduced 15 and 25 June snapshots have version 1. Under the previous repeated-cascade
implementation, the later snapshot could advance more than once within the same batch, so the version
assertion detects the original defect rather than only comparing final monetary values.

## P2-12 — append-only ledger policy was not a runtime database privilege boundary

### Observed defect

The frozen Stage 2 ER model states that `investment.transaction_entries` is append-only and that the
application ledger writer receives no UPDATE or DELETE privilege. Runtime configuration previously did
not establish or verify that separation. An API deployed with the migration/schema-owner connection
could retain PostgreSQL authority to mutate or delete canonical ledger rows even though application
code treated them as immutable.

That was an architecture-enforcement gap: code convention and constraints were not equivalent to a
least-privilege database capability boundary.

### Runtime role

Stage 3.33 adds `infrastructure/postgres/runtime/openinvest_runtime_role.sql`, intended to be applied by
the privileged migration/operator connection after schema migrations.

The script owns a NOLOGIN capability role named `openinvest_runtime`. A provider-managed dedicated API
LOGIN role is granted membership in that capability role. Credentials and provider passwords are not
stored in the repository.

The capability role receives the existing application table privileges required by the current runtime,
then explicitly removes mutable capabilities from protected append-only tables:

- `investment.transaction_entries`: SELECT + INSERT, no UPDATE/DELETE/TRUNCATE;
- `audit.events`: SELECT + INSERT, no UPDATE/DELETE/TRUNCATE.

The script deliberately does not create broad default privileges for future tables. Newly introduced
tables therefore remain unavailable until their runtime capabilities are reviewed and the runtime role
script is reapplied/extended explicitly.

### Production/staging startup guard

`postgres.OpenRuntime` validates PostgreSQL's effective privileges before returning a Store. Outside
explicit `development`/`local` mode, the API uses this runtime-opening path.

Startup fails closed if the effective database principal:

- is a PostgreSQL superuser;
- owns, or inherits the owner role of, a protected append-only table;
- lacks schema usage or the required SELECT/INSERT capability;
- can UPDATE, DELETE, or TRUNCATE a protected append-only table.

The validation is capability-based rather than tied to a provider-specific login name. A managed
provider can therefore provision its own LOGIN role while inheriting `openinvest_runtime`.

Explicit local/development mode may continue using the schema-owner connection for migration and local
workflow convenience. That exception is not accepted for staging or production.

### PostgreSQL-backed privilege proof

CI provisions a dedicated `openinvest_runtime_ci` LOGIN role, grants it membership in
`openinvest_runtime`, and supplies both owner and runtime database URLs to the Go integration suite.

The Stage 3.33 integration test proves that:

1. `OpenRuntime` rejects the schema-owner connection;
2. `OpenRuntime` accepts the least-privilege runtime connection;
3. the runtime connection can create a portfolio and append a financial transaction through the normal
   service path;
4. direct UPDATE of `investment.transaction_entries` fails with PostgreSQL permission denial;
5. direct DELETE of `investment.transaction_entries` fails with PostgreSQL permission denial;
6. the runtime role has no TRUNCATE privilege.

The migration-validation job independently applies the runtime role script and validates the protected
table ACL shape after the full migration rollback/reapply rehearsal.

## Atomicity, idempotency, and concurrency boundaries

Stage 3.33 does not weaken Stage 3.32 exact replay semantics. A new import command still reserves its
idempotency command before mutable business work, acquires the portfolio lock, validates import
identity/conflicts, inserts ledger entries, rebuilds snapshots, constructs the exact response artifact,
persists command completion, and commits atomically. A completed duplicate still resolves the stored
response artifact before current mutable portfolio state.

Same-portfolio writes remain serialized by the existing portfolio row lock. The exact snapshot plan is
therefore stable for the transaction in which it is executed.

The pre-existing compatibility methods remain for older internal test/service interfaces, but the
canonical `cmd/api` runtime uses the replay-aware composition and the importflow boundary now consumes
the exact PostgreSQL-owned outcome rather than deriving snapshot dates itself.

## Deployment boundary

Stage 3.33 introduces a required separation between privileged schema administration and the API
runtime connection for non-development environments:

1. apply migrations with the owner/migration connection;
2. apply `infrastructure/postgres/runtime/openinvest_runtime_role.sql` with a role permitted to manage
   the capability role and grants;
3. provision a dedicated API LOGIN through the database provider;
4. grant that LOGIN membership in `openinvest_runtime`;
5. configure staging/production `DATABASE_URL` with the dedicated LOGIN.

A provider environment that cannot apply the role/grant script has not satisfied P2-12 operationally;
the production API will fail its runtime privilege validation rather than silently falling back to the
schema-owner connection.

## Regression evidence

Stage 3.33 adds or updates coverage for:

- exact database-owned import snapshot-date reporting;
- a backdated multi-date batch rebuilding each affected date once;
- snapshot-version assertions that detect repeated rebuild work;
- preservation of exact replay artifact construction from the database-owned import outcome;
- existing expired-review-token replay behavior under the new outcome contract;
- legacy importflow tests consuming an appender-owned date list rather than recomputing it;
- owner connection rejection by runtime privilege validation;
- successful business append under the dedicated runtime role;
- PostgreSQL permission denial for runtime UPDATE and DELETE on the canonical ledger;
- absence of runtime TRUNCATE capability;
- CI-level ACL verification after migration apply/rollback/reapply;
- the existing frontend, Python, OpenAPI, Docker Compose, migration, and full Go suites.

An intermediate CI run failed only because the new snapshot regression initially used a fabricated
`SourceFingerprint`; the existing Stage 3.27 normalized-import identity validation correctly rejected
that fixture. The test was corrected to derive the fingerprint with the production
`NormalizedTransactionFingerprint` function. No runtime validation was weakened.

## Scope boundary and next gate

Stage 3.33 does not close P2-16 or P2-17. The CI changes in this stage are only the minimum evidence
needed to exercise the new PostgreSQL runtime-role boundary; general branch protection, race/vet/
vulnerability/dependency-review hardening remain the separate Stage 3.34 scope.

Stage 3.25 privacy Security Review evidence planning remains separate and is not superseded. This stage
makes no privacy-lifecycle, provider-retention, deletion, backup, tax, mobile, or product-scope claim.

P2-10, P2-11, and P2-12 remain OPEN until the final immutable PR head passes exact-head CI, independent
review returns `APPROVED` with the three findings closed and no blocking regression, explicit human
squash-merge authorization is received, implementation PR #69 is squash-merged into `develop`, and a
separate closure-governance record becomes canonical.
