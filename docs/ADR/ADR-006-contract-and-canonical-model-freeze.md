# ADR-006: Freeze the MVP Contract and Canonical Model Before Implementation

| Field | Value |
| --- | --- |
| Document ID | ADR-006 |
| Version | 1.0.0 |
| Status | Proposed / Awaiting Stage 2 Review |
| Owner | Principal Architect |
| Supersedes | Conflicting API/DTO/ER details in Documents 1–41 |
| Dependencies | Documents 42–43; ADR-002; ADR-003; ADR-004; ADR-005 |
| Last Review Date | 2026-06-20 |
| Next Review Date | At Stage 2 approval |

## Context

OpenInvest cannot safely implement portfolio, transaction, analytics, or dashboard behavior while
the API envelope, financial wire formats, auth boundary, immutable transaction semantics, and
logical data ownership remain implicit or contradictory across legacy specifications.

The product requires one API for web and later native clients, exact financial values, explainable
calculations, privacy-by-design, immutable history, rebuildable snapshots, and schema isolation.
Implementation-first choices would harden accidental details and create expensive migrations.

## Decision

Freeze the Stage 2 artifacts as the pre-implementation contract when this ADR is approved:

1. OpenAPI 3.1 is the normative MVP web HTTP contract under `/api/v1`.
2. Success uses `data + meta`; errors use `error + meta`.
3. Request/trace identity is returned in headers and body metadata.
4. Financial amounts/rates are Decimal strings; Money is RUB-only in MVP.
5. Financial dates are BusinessDate; technical timestamps are UTC SystemTimestamp.
6. MVP web auth uses short-lived bearer access tokens, rotating Secure HttpOnly refresh cookie,
   and CSRF token for cookie-authenticated refresh/logout.
7. Cursor pagination is opaque and deterministic.
8. Financial writes are idempotent; transaction PATCH/DELETE append correction/reversal history.
9. Canonical DTO meaning is independent from Go structs, SQL rows, provider payloads, and UI models.
10. Logical ER ownership uses isolated identity, investment, analytics, audit, and future tax schemas.
11. Identity-to-financial linkage is isolated and must be irreversibly destroyed on deletion;
    retained ledger becomes Anonymous Financial History.
12. Snapshots are versioned, deterministic, and rebuildable from immutable transactions.
13. Database evolution follows Expand → Populate → Switch → Validate → Contract.
14. No SQL table migration or business implementation is authorized by this ADR.

Normative artifacts:

- `openapi/openapi.yaml` and its referenced component/example files;
- `docs/api/API_CONTRACT_STAGE_02.md`;
- `docs/domain/CANONICAL_MODEL_STAGE_02.md`;
- `docs/database/ER_MODEL_STAGE_02.md`;
- `docs/database/MIGRATION_STRATEGY_STAGE_02.md`.

## Alternatives considered

### Implement backend first and document afterward

Rejected. It violates API First, makes clients depend on accidental Go/SQL shapes, and delays
discovery of privacy and immutable-history conflicts until migrations are expensive.

### Expose provider-shaped market endpoints

Rejected. It leaks MOEX/other provider implementation details, couples clients to external
licenses/rate limits, and violates the backend collector boundary.

### JSON numbers for convenience

Rejected. Binary floating point cannot provide reproducible financial arithmetic and contradicts
the approved Decimal/Half-Even standard.

### Mutable transaction CRUD

Rejected. Updating or deleting financial facts destroys auditability and snapshot reproducibility.
CRUD-like HTTP interaction is preserved through appended corrections and reversals.

### Offset pagination

Rejected. It produces unstable pages under concurrent append-only writes and degrades at scale.

### Store refresh tokens in browser storage

Rejected for MVP web. It increases token-exfiltration impact. Rotating HttpOnly cookie plus CSRF
protection is the frozen browser contract; native mobile auth remains future work.

### Create SQL tables during contract freeze

Rejected. Physical types, constraints, roles, RLS, event state, anonymization keys, and migration
tooling require a separate infrastructure review and executable database tests.

## Consequences

### Positive

- Backend and frontend can be reviewed/generated against one versioned boundary.
- Financial precision and date semantics are testable before algorithms exist.
- Immutable ledger and snapshot semantics cannot be accidentally weakened by CRUD scaffolding.
- Identity separation/anonymization requirements shape storage before personal data is collected.
- Provider and persistence details remain replaceable behind anti-corruption boundaries.
- Migration risk is managed through an explicit zero-destructive-change lifecycle.

### Costs and constraints

- Contract changes require review before implementation.
- DTOs may intentionally differ from persistence and domain entities, requiring mapping code.
- Cookie/CSRF web auth needs explicit browser integration and security testing.
- Append-only correction/reversal is more complex than mutable CRUD.
- External reference resolution and example validation must become CI requirements before merge.
- Stage 4 must still choose exact PostgreSQL types, roles, migration tooling, and encryption design.

## Compatibility and rollback

Before implementation, rollback is removal/reversion of the Stage 2 contract commit. After clients
or servers implement `/api/v1`, breaking semantic changes require a new ADR and versioned contract.
No database rollback is involved because Stage 2 creates no tables or data.

## Security and privacy impact

- Reduces secret exposure through HttpOnly refresh cookie and bounded error DTOs.
- Prevents direct identity fields in investment/analytics schemas.
- Requires a reviewed cryptographic-erasure/restore design before deletion implementation.
- Prohibits secrets, identity documents, and raw financial documents in logs/audit payloads.
- Introduces no collection of personal or financial data by itself.

## Performance and cost impact

The contract introduces no runtime cost. Cursor pagination, server-composed dashboard responses,
and rebuildable snapshots establish performance-compatible boundaries. Actual SLO evidence,
caching, materialization, and infrastructure cost remain later-stage responsibilities.

## Approval condition

On Principal Architect and human approval, change status to `Accepted`, register ADR-006 and the
Stage 2 freeze in `docs/SOURCE_OF_TRUTH.md`, and require contract validation in CI. Until then the
artifacts are proposed and business implementation remains forbidden.
