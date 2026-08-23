# Stage 3.33 — Snapshot Rebuild Accuracy and PostgreSQL Runtime Immutability

| Field | Value |
| --- | --- |
| Status | Implementation candidate; repeat independent review and human merge authorization pending |
| Owner | Principal Architect |
| Baseline | `develop` at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` |
| Branch | `fix/stage-03-33-snapshot-immutability` |
| Implementation PR | #69 |
| Trigger | Repository-audit P2-10, P2-11, and P2-12 |
| Scope | Exact import snapshot rebuild reporting, one-pass affected-snapshot rebuild planning, authenticated PostgreSQL runtime append-only privilege enforcement, startup credential-graph validation, regression/CI evidence |
| Out of scope | P2-16/P2-17, all P3 findings, Stage 3.25 privacy Security Review evidence work, provider credential provisioning, product-scope expansion, snapshot methodology redesign |

## Purpose

Stage 3.33 addresses three repository-audit gaps without changing the financial methodology or product surface. P2-10 and P2-11 concern correctness and efficiency of derived snapshot rebuilds. P2-12 concerns enforcing the frozen append-only ledger architecture at the database privilege boundary used by the running application.

PostgreSQL remains the authoritative transaction boundary. The stage introduces no queue, Redis dependency, background worker, microservice, or new financial feature.

## P2-10 — exact `snapshotDatesRebuilt`

The old import response derived `snapshotDatesRebuilt` from approved input trade dates even though PostgreSQL could also rebuild later existing snapshots. A backdated import could therefore commit projection work not reported by the response.

Stage 3.33 moves affected-date ownership to PostgreSQL. `planAffectedSnapshotDates` computes the deterministic ordered union:

`all distinct imported trade dates ∪ all existing active-methodology snapshot dates >= earliest imported trade date`

The PostgreSQL mutation returns `ImportAppendOutcome` containing inserted transactions and the exact ordered `SnapshotDatesRebuilt` set. `importflow` copies that database-owned result instead of recomputing dates from input rows.

On the replay-aware production path, the exact HTTP artifact is built from this outcome before the same PostgreSQL transaction commits. Stage 3.32 exact replay ordering is preserved: reservation, portfolio lock, ledger mutation, exact snapshot plan/rebuild, response-artifact construction, command completion, commit. A completed duplicate still resolves the persisted artifact before mutable portfolio state.

Both independent Stage 3.33 reviews to date marked P2-10 CLOSED.

## P2-11 — one rebuild per affected date

The old batch path called `rebuildAffectedSnapshots` once per distinct imported trade date. Each call cascaded over later snapshots, so one later snapshot could be rebuilt repeatedly in the same command.

The new canonical import path computes one union before any rebuild. `rebuildSnapshotPlan` then walks the sorted unique dates once while the portfolio row lock remains held. All new ledger rows are inserted before planning/rebuild, so every projection reads the complete command-local ledger state.

The PostgreSQL regression creates baseline snapshots on 10, 20, and 30 June, then imports entries on 15 and 25 June. The exact result must report 15, 20, 25, and 30 June. Snapshot versions prove work count, not only final balances:

- 10 June remains version 1;
- 15 June is version 1;
- 20 June advances to version 2;
- 25 June is version 1;
- 30 June advances to version 2.

Both independent Stage 3.33 reviews to date marked P2-11 CLOSED. The first review noted the old direct non-replay PostgreSQL compatibility method still contains the historical cascading implementation, but confirmed that method is not the canonical production HTTP/importflow path. Removal/deprecation remains a non-blocking maintainability cleanup.

## P2-12 — PostgreSQL runtime append-only boundary

### Initial remediation

`infrastructure/postgres/runtime/openinvest_runtime_role.sql` defines a repository-owned `openinvest_runtime` capability role as NOLOGIN/NOSUPERUSER/NOCREATEDB/NOCREATEROLE/NOREPLICATION. A dedicated provider-managed API LOGIN inherits that capability role; schema-owner/migration credentials remain separate.

The protected append-only tables are runtime read/append only:

- `investment.transaction_entries`: SELECT + INSERT, no UPDATE/DELETE/TRUNCATE;
- `audit.events`: SELECT + INSERT, no UPDATE/DELETE/TRUNCATE.

Outside explicit `development`/`local` mode, `cmd/api` opens PostgreSQL through `postgres.OpenRuntime`. Local/development may use the owner connection only as an explicit development convenience.

CI provisions a separate runtime LOGIN and proves normal portfolio/transaction writes succeed while direct ledger UPDATE and DELETE fail with PostgreSQL permission denial and TRUNCATE capability is absent. Migration validation independently checks the ACL shape after migration rollback/reapply.

### First independent review finding — masked authenticated principal

The first independent Stage 3.33 review marked P2-12 NOT CLOSED because the startup validator inspected only `current_user`.

PostgreSQL distinguishes the authenticated `session_user` from the effective `current_user`. An overprivileged authenticated session can execute or start with `SET ROLE openinvest_runtime`, making `current_user` appear least-privileged while `session_user` remains privileged and can later recover its original authority. The same problem exists for a clean-looking LOGIN that can `SET ROLE` into a non-inherited role carrying ledger-mutation privileges.

This was a valid P2 blocker because it meant the startup check could prove only the currently selected effective role, not that the authenticated runtime credentials themselves were incapable of escaping the append-only boundary.

### First post-review remediation

`ValidateRuntimePrivileges` was changed to perform validation on one physical `database/sql` connection so every identity and ACL check describes the same PostgreSQL session.

It now:

1. reads both `session_user` and `current_user`;
2. rejects startup unless the authenticated and effective identities are equal;
3. validates the authenticated LOGIN itself is not SUPERUSER, CREATEDB, CREATEROLE, REPLICATION, or BYPASSRLS;
4. rejects CREATE capability in every runtime schema: `identity`, `investment`, `analytics`, and `audit`;
5. validates required schema USAGE plus SELECT/INSERT on each protected append-only table;
6. rejects owner/owner-role membership and UPDATE/DELETE/TRUNCATE/TRIGGER capability on the protected tables;
7. enumerates every role reachable from the authenticated LOGIN through `pg_has_role(session_user, role, 'SET')`;
8. applies the elevated-role, schema-CREATE, owner-membership, and protected-table mutation checks to every SET-reachable role.

Two adversarial PostgreSQL regressions were added:

**Privileged session masked by `SET ROLE`:** owner/superuser credentials connect with a connection-time role setting that produces `session_user != current_user` and `current_user = openinvest_runtime`. The fixture verifies the split exists, then proves `OpenRuntime` rejects it.

**Latent SET-role escalation:** a NOLOGIN role receives schema USAGE plus UPDATE on `investment.transaction_entries`, is granted to the clean runtime LOGIN with `INHERIT FALSE, SET TRUE`, and the fixture proves UPDATE is not directly inherited while SET capability exists. `OpenRuntime` rejects the LOGIN.

The code-only remediation head `64190a2cc42dfc50f747e63b508a23aa0d6a79da` passed CI #190 with all six jobs successful. Evidence-only follow-up head `35960c7821e8fce9577bd674ee5f2c7e06be2f61` passed CI #191 with all six jobs successful.

### Second independent review finding — latent ADMIN OPTION escalation

The repeat independent review reconfirmed P2-10 and P2-11 CLOSED and confirmed the previous P2-12 blocker was fixed, but correctly kept P2-12 OPEN because the credential graph still considered only roles already reachable through `SET ROLE`.

PostgreSQL role membership has a separate `ADMIN OPTION`. A runtime LOGIN can hold membership in a ledger-mutating role with `ADMIN TRUE, INHERIT FALSE, SET FALSE`. That role is neither directly inherited nor currently SET-reachable, so the first credential-graph validator could accept the LOGIN. However, ADMIN OPTION allows the runtime principal to administer membership in that role and manufacture a later SET/INHERIT path for itself. The API has no legitimate need to administer PostgreSQL roles, so any such capability is incompatible with the fail-closed runtime boundary.

### ADMIN OPTION remediation

The runtime validator now rejects role-administration capability as a class, rather than attempting to decide whether an administrable role is currently dangerous.

`validateNoRoleAdministration` uses PostgreSQL's role-membership capability model to reject any role for which the checked runtime identity holds `MEMBER WITH ADMIN OPTION`. The check is applied to:

1. the authenticated `session_user` itself; and
2. every role already reachable from that authenticated principal through `SET ROLE`.

This closes both direct and reachable role-administration paths. A runtime credential therefore cannot pass startup validation while retaining authority to grant itself a new SET/INHERIT path later.

The clean `openinvest_runtime` capability membership remains valid because it is granted without ADMIN OPTION.

### ADMIN OPTION adversarial regression

A third PostgreSQL attack regression creates a NOLOGIN role with UPDATE on `investment.transaction_entries` and grants it to the clean runtime LOGIN using:

`ADMIN TRUE, INHERIT FALSE, SET FALSE`.

Before calling `OpenRuntime`, the fixture proves all three required facts independently:

- direct ledger UPDATE is false;
- `pg_has_role(session_user, role, 'SET')` is false;
- `pg_has_role(session_user, role, 'MEMBER WITH ADMIN OPTION')` is true.

`OpenRuntime` must reject that credential despite the dangerous role not yet being SET-reachable. The original clean runtime regression and the prior masked-session/latent-SET regressions remain active.

The code-plus-regression ADMIN OPTION remediation passed GitHub Actions CI #193 with all six jobs successful, including the full PostgreSQL-backed Go suite. Final immutable repeat-review head and final exact-head CI are recorded in PR metadata after this evidence update rather than self-referenced in this mutable stage document.

## Atomicity and concurrency boundaries

Stage 3.33 does not weaken Stage 3.32 idempotency/replay semantics. New imports reserve the command, lock the portfolio, perform identity/conflict checks, append all ledger entries, calculate one exact snapshot plan, rebuild the plan once, construct the response artifact from the DB-owned outcome, complete the command, and commit atomically. Completed duplicates still return the persisted artifact before mutable portfolio state.

Same-portfolio financial commands remain serialized by the existing portfolio row lock, so the plan cannot be interleaved by another same-portfolio financial write.

## Deployment boundary

For non-development environments the required sequence is:

1. apply migrations with the owner/migration connection;
2. apply `infrastructure/postgres/runtime/openinvest_runtime_role.sql` with the privileged operator connection;
3. provision a dedicated API LOGIN through the provider;
4. grant that LOGIN ordinary membership in `openinvest_runtime` without ADMIN OPTION;
5. ensure the LOGIN has no separate elevated memberships, SET-reachable mutation roles, schema CREATE capability, or role ADMIN OPTION memberships;
6. configure staging/production `DATABASE_URL` with that dedicated LOGIN.

A provider setup that cannot establish these grants has not satisfied P2-12. Production/staging fails startup rather than falling back to owner credentials.

## Review and verification history

- Initial implementation CI #186 failed only because the new snapshot regression used a fabricated import `SourceFingerprint`; existing Stage 3.27 validation correctly rejected it.
- The fixture was changed to derive the production `NormalizedTransactionFingerprint`; no runtime validation was weakened.
- CI #187 passed all six jobs.
- Evidence-only update produced CI #188, also 6/6.
- First independent review: P2-10 CLOSED, P2-11 CLOSED, P2-12 NOT CLOSED; REQUEST CHANGES because `current_user` validation did not prove authenticated/session principal safety or latent SET-role safety.
- First post-review remediation added `session_user` validation, same-connection identity checks, runtime-schema CREATE checks, protected-table owner/mutation checks, SET-reachable role enumeration, and two PostgreSQL attack regressions.
- CI #190 and evidence-head CI #191 passed all six jobs.
- Second independent review: P2-10 CLOSED, P2-11 CLOSED, P2-12 NOT CLOSED; REQUEST CHANGES because a role membership with ADMIN TRUE / INHERIT FALSE / SET FALSE could manufacture its own later escalation path.
- Second post-review remediation rejects all ADMIN OPTION memberships held by the authenticated principal or any SET-reachable role and adds a PostgreSQL ADMIN-option attack regression.
- Code-plus-regression CI #193 passed all six jobs.
- Repeat independent review on the final immutable head remains required before merge.

## Scope boundary and next gate

Stage 3.33 does not close P2-16 or P2-17. The CI changes here are only the evidence required for this PostgreSQL boundary. General branch protection, race/vet/vulnerability/dependency-review hardening remain separate work.

Stage 3.25 privacy Security Review evidence planning also remains separate.

P2-10, P2-11, and P2-12 are not canonically closed until the final immutable PR head passes exact-head CI, repeat independent review returns APPROVED with all three findings closed and no blocking regression, explicit human squash-merge authorization is received, PR #69 is squash-merged into `develop`, and separate closure governance becomes canonical.
