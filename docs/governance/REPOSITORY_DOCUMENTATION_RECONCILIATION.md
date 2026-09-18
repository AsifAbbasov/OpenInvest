# Repository Documentation Reconciliation — through Stage 3.72

| Field | Value |
| --- | --- |
| Document ID | GOV-DOC-RECON-001 |
| Version | 1.0.0 |
| Status | MERGE-ACTIVATED — CANONICAL ONLY AFTER PROTECTED MERGE |
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


## Documentation precedence


## No runtime change

This reconciliation changes no Go, PostgreSQL schema/migration, OpenAPI runtime behavior, Next.js runtime behavior, dependency, CI policy, provider wiring, credentials, infrastructure or production deployment. It is documentation/governance only.

## Next gate

Canonical record: PR #148, PR #150; commit(s) `0580bf7e98c532202f84bbf9ceacd97aedbe4140`, `ff6d3efb8bb0d63f8cab55ac82fdf1052946aa02`.
