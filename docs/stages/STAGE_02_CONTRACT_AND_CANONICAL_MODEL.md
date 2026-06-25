# Stage 2 — Contract and Canonical Model Freeze

| Field | Value |
| --- | --- |
| Document ID | STAGE-02 |
| Version | 1.0.4 |
| Status | Implemented / REQUEST CHANGES fixes applied |
| Owner | Builder Engineer |
| Supersedes | Stage 1 non-business OpenAPI skeleton after approval |
| Dependencies | Stage 1 approval; Architecture Freeze v1.2; ADR-003; proposed ADR-006 |
| Last Review Date | 2026-06-21 |
| Next Review Date | At Stage 2 approval |

## Goal

Freeze the MVP web API, canonical DTO vocabulary, logical ER ownership, and safe migration
strategy before backend business implementation.

## Scope completed

- OpenAPI 3.1 `/api/v1` contract for every authorized MVP endpoint.
- Reusable schemas for response envelopes, errors, tracing, auth, pagination, exact financial
  values, assets, portfolios, transactions, summaries, snapshots, dividends, returns, purchasing
  power, and dashboard.
- Domain-grouped JSON request/response examples covering every endpoint.
- Canonical model documentation independent from implementation and persistence.
- Logical ER diagram for identity, investment, analytics, audit, and future tax isolation.
- Draft relationship/index inventory with immutable ledger and rebuildable snapshot semantics.
- Expand → Populate → Switch → Validate → Contract migration strategy.
- ADR-006 proposal defining the contract/canonical-model freeze and its approval gate.

## Explicitly not implemented

- no Go services, repositories, handlers, or domain calculations;
- no Python workers or calculations;
- no SQL migrations, tables, ORM entities, or database writes;
- no MOEX, CBR, Rosstat, or other external integration;
- no frontend/mobile screen or client generation;
- no tax export, foreign security, prediction, or AI behavior.

## Review-size exception

The complete Stage 2 PR contains 26 files, one above the default 25-file review budget. The contract,
canonical/ER/migration records, OpenAPI validator CI, five synchronized governance registries,
implementation log, and source-registry correction must remain consistent in one reviewable freeze.
Splitting one governance file would create a knowingly inconsistent Source of Truth and would not
reduce contract complexity. Generated/example artifacts remain fully reviewable and are not hidden
from either reviewer.

Auditable evidence: Principal Architect / Human Reviewer approval for this exact PR #2 26-file
review-size exception is recorded in GitHub PR comment
`https://github.com/AsifAbbasov/OpenInvest/pull/2#issuecomment-4800618888`.

The approval covers the review-size exception only. It is not merge approval, not ADR-006
acceptance, not authorization to start Stage 3, and not a reusable exception for future PRs.

## Contract decisions

- versioned `/api/v1` paths;
- `data + meta` success and `error + meta` failure envelopes;
- request ID and trace ID in headers and response metadata;
- Decimal strings, RUB Money, BusinessDate, UTC SystemTimestamp, half-even/8-digit precision;
- short-lived bearer access token and rotating HttpOnly refresh cookie with CSRF for MVP web;
- opaque cursor pagination;
- idempotency for financial commands;
- transaction PATCH/DELETE append correction/reversal rather than mutate/delete history;
- transaction reversal requires an explicit `effectiveDate` BusinessDate; reversal economics are
  never derived from request/system/worker timestamps;
- official dividend events only; gross dividend calculator excludes tax;
- no provider/persistence implementation detail in public DTOs.

## Assumptions

1. The Stage 2 branch is based on current `develop` commit `75af67d`. Any later `develop` changes
   must be incorporated through the approved update/rebase workflow followed by complete
   revalidation; this report does not claim that other feature branches are included.
2. MVP is web-only; secure native refresh-token transport is intentionally deferred.
3. MOEX ticker is sufficient as the public asset path key for MVP while internal source IDs remain
   private.
4. Portfolio deletion is removal from active use, not destruction of immutable financial history.
5. Tax amounts may be recorded from broker history, but tax calculation/export/advice is excluded.
6. Exact PostgreSQL types, migration library, RLS/roles, and encryption-key hierarchy remain Stage 4
   physical-design decisions constrained by the Stage 2 semantics.

## Open questions

None. All decisions needed to begin post-approval implementation are explicit. Any new question
must use Issue → ADR → review → approval and cannot be resolved ad hoc in production code.

## Verification

- `ruby scripts/validate_openapi.rb`: passed; 22 operations, 2,501 resolved references,
  11 OpenAPI/component/example documents. The validator now includes focused mutation guards for
  UUID, BusinessDate, traceparent, `unevaluatedProperties`, non-negative aggregates, reversal date
  semantics, command invariants, and dividend calculation invariants.
- YAML parse: passed for root contract and both component files.
- JSON parse: passed for all eight domain-grouped example files.
- Required endpoint and success-example coverage: passed.
- `go test ./...`: passed.
- `uv run pytest`: passed, 1 test; one known upstream FastAPI/Starlette TestClient deprecation
  warning.
- `pnpm install --frozen-lockfile`: passed with pnpm 11.8.0.
- `pnpm run typecheck`: passed.
- `pnpm run build`: passed.
- `docker compose config --quiet`: passed with a validation-only password.
- Pull-request CI includes `ruby scripts/validate_openapi.rb` as a required contract job.
- Markdown internal-link check: passed.
- whitespace check for every new file: passed.

The repository-owned validator is intentionally not a complete JSON Schema 2020-12 implementation
and must not be represented as one. It provides structural/reference/example checks plus focused
mutation guards for UUID, BusinessDate, traceparent, `unevaluatedProperties`, non-negative
financial aggregates, transaction command invariants, and dividend calculation invariants. A
standards-compliant OpenAPI/JSON Schema ruleset such as Redocly remains recommended in connected CI
or review before merge, but the repository validator no longer claims full standards compliance.

## Known risks

- Full Redocly/spec ruleset evidence remains pending because npm registry access was unavailable;
  this is explicitly documented and not hidden by the repository validator.
- The OpenAPI CI job uses the GitHub-hosted runner Ruby. Pinning Ruby through an additional setup
  action is deferred as a non-blocking operational hardening item to avoid adding a new supply-chain
  dependency inside this Stage 2 blocker fix.
- GitHub CI evidence remains pending until the updated feature branch is pushed and PR #2 runs.
- OpenAPI documents response shape but does not prove calculation correctness; financial vectors
  remain mandatory before algorithms are implemented.
- Authentication, anonymization key destruction, event reliability, and physical database roles
  require specialist security/infrastructure review before implementation.

## Internal Review Evidence

- Changed files reviewed: 12 tracked Stage 2 files in the current diff.
- Review verdict: `APPROVED`.
- Blocking findings: none.
- Resolved findings: external Stage 2 blockers 1–7 addressed in this blocker-fix diff.
- Remaining non-blocking notes: Redocly/full JSON Schema ruleset evidence and Ruby setup pinning
  remain documented operational hardening items before merge/release.
- Reviewer edit confirmation: Internal Review Agent confirmed read-only review and made no edits.

External ChatGPT review and explicit human approval remain required before merge.

## Rollback

Revert the Stage 2 contract commits. No runtime behavior, database object, persisted data,
external source, or client has been changed by this stage.

## Stop condition

After validation, report files/checks/risks and wait for explicit permission before commit, push,
or Draft PR creation. Backend implementation remains forbidden.
