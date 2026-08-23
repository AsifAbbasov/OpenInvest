# PostgreSQL runtime role boundary

Stage 3.33 separates the migration/schema-owner connection from the API runtime connection.

## Required deployment order

1. Apply repository migrations with the database owner or migration role.
2. Run `openinvest_runtime_role.sql` with that same privileged connection.
3. Create or provision a dedicated LOGIN role for the API through the database provider.
4. Grant that LOGIN role membership in `openinvest_runtime`.
5. Configure staging/production `DATABASE_URL` with the dedicated LOGIN role, never the migration/schema owner.

Example operator SQL after the provider has created the LOGIN role:

```sql
GRANT openinvest_runtime TO openinvest_app_login;
```

Passwords and provider credentials are intentionally not stored in this repository.

## Ledger invariant

The `openinvest_runtime` capability role has `SELECT` and `INSERT` on
`investment.transaction_entries` but no `UPDATE`, `DELETE`, or `TRUNCATE`. Corrections and reversals
remain append operations. `audit.events` is protected with the same append/read-only runtime shape.

The API additionally validates the effective runtime principal at startup outside explicit local or
development mode. Startup fails closed when the connected role is a superuser, owns the protected
table (or inherits its owner role), lacks required `SELECT`/`INSERT`, or can `UPDATE`, `DELETE`, or
`TRUNCATE` the protected table.

The runtime grants script is intentionally rerun after migrations. It does not establish broad default
privileges for future tables; a new table therefore remains unavailable to the runtime role until its
required capabilities are reviewed explicitly.
