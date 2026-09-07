# Repository Documentation Reconciliation — through Stage 3.72

| Field | Value |
| --- | --- |
| Document ID | GOV-DOC-RECON-001 |
| Version | 1.0.0 |
| Status | MERGE-ACTIVATED — CANONICAL ONLY AFTER PROTECTED MERGE |
| Owner | Principal Architect |
| Baseline | Stage 3.72 lifecycle closure / PR #146 squash merge `ac0396eff47ce5cd862f41a5ec8ab2803de7fa58` |
| Date | 2026-09-07 |

## Purpose

This document closes cross-registry documentation drift that accumulated while runtime and stage dossiers continued to advance. It does not replace the detailed stage dossiers. It makes the repository entrypoints and governance registries agree on what has actually been implemented, what is deliberately dormant/unavailable, and what remains separately gated.

This record remains non-canonical while it exists only on the reconciliation branch; its synchronized status becomes canonical only when this exact documentation-only change set is squash-merged into protected `develop`.

## Canonical product and engineering state

- Stage 3 runtime is complete through **Stage 3.72 — Portfolio Position Projection / Cost Basis View**.
- Authentication, portfolio/transaction flows, CSV import/reconciliation, instrument catalog and Web asset discovery are implemented.
- The original Stage 3.16 repository audit is **32/32 CLOSED** through Stage 3.56.
- Market Data Provider Boundary is implemented. The delayed MOEX ISS TQBR adapter exists, but production/public activation is **NO-GO** and the adapter remains dormant.
- Corporate Actions provider-neutral Features 3A/3B/3C, Calendar/Heatmap API/UI and request cancellation are implemented and documentation-closed. Feature 3D real-source activation remains separately gated.
- Dividend Calculator is implemented with backend-owned exact Decimal arithmetic and durable idempotency/replay.
- ADR-009 is accepted. Manual SELL, deterministic portfolio-local ledger ordering, WAC and remaining acquisition basis are canonical Stage 3.71 semantics.
- Stage 3.72 exposes open STOCK/BOND positions and acquisition-basis allocation through a dedicated API/UI. Market price, market value, unrealized P/L and market weight remain explicitly unavailable because no approved market-price source is activated.
- Stage 3.25 privacy evidence collection remains documentation-only and authorizes no privacy-lifecycle implementation.

## Reconciled surfaces

The following surfaces are synchronized by this documentation-only closure:

1. `README.md` — project entrypoint and current product state;
2. `docs/SOURCE_OF_TRUTH.md` — authoritative current architecture/lifecycle status;
3. `docs/ROADMAP.md` — ordered stage history through 3.72;
4. `docs/IMPLEMENTATION_LOG.md` — implementation continuation through 3.72;
5. `docs/CHANGELOG.md` — current reconciliation milestone and historical chronology;
6. `docs/VERSION_MATRIX.md` — canonical lifecycle ranges and review gates;
7. `docs/OPEN_QUESTIONS.md` — no stale pre-3.72 trigger;
8. `docs/DOCUMENT_INDEX.md` — navigation to the reconciliation and audit register;
9. `docs/audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` — all 32 original findings indexed as closed.

## Documentation precedence

Detailed stage dossiers remain the source for exact implementation rationale, test vectors, review findings, residual risk and commit/CI evidence. When a short registry summary conflicts with a canonical stage dossier or `SOURCE_OF_TRUTH.md`, the Source of Truth and accepted ADR/stage authority take precedence according to the existing documentation freeze rules.

## No runtime change

This reconciliation changes no Go, PostgreSQL schema/migration, OpenAPI runtime behavior, Next.js runtime behavior, dependency, CI policy, provider wiring, credentials, infrastructure or production deployment. It is documentation/governance only.

## Next gate

No Stage 3.73 work is implied. Any new product/runtime stage must be separately planned and reviewed. Architecture-changing provider/public activation, market valuation, privacy lifecycle, tax-basis semantics, imported SELL expansion or other frozen-boundary changes must enter the existing Issue → ADR/plan → review → explicit authorization workflow.
