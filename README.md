# OpenInvest

OpenInvest is an independent, privacy-first investment analytics platform. It is not a broker, bank, asset manager, trading system, or investment adviser.

This repository contains the closed Stage 2 contract/canonical-model baseline, the accepted Next.js
Web presentation-layer baseline, and the staged Stage 3 MVP vertical-slice implementation.

Stage 3 currently has closed slices for the local database foundation, portfolio/transaction
vertical slice, Next.js presentation shell, end-to-end verification, CSV import/reconciliation,
authentication/privacy boundaries, Web authentication UI, backend-owned instrument catalog, the Go
API asset search/detail boundary over the approved local catalog, and the reviewed Web asset
discovery UI. Asset search returns
backend-owned catalog summaries with `lastPrice: null`; asset-card detail remains intentionally
deferred until mandatory source provenance and required detail fields can be populated without
fabricated data. Stage 3.16 repository audit planning and its audit-fix closure are closed; the
fixes were squash-merged into `develop` through PR #44 at
`9e6b8a753bf73ef020ce40461df25a5878344d92`. Stage 3.17 privacy-lifecycle planning and the
Stage 3.18 contract/security proposal are closed through PRs #46 and #47. Stage 3.19 privacy
security/ADR proposal is closed through PR #48, and Stage 3.20 privacy threat model is closed through
PR #49, Stage 3.21 privacy data inventory is closed through PR #50, and Stage 3.22 key custody is
closed through PR #51, and Stage 3.23 deletion-marker control-plane planning is closed through PR
#52, and Stage 3.24 Security Review readiness planning is closed through PR #53. Stage 3.25 is the
active documentation-only evidence-collection plan; it does not collect evidence, perform Security
Review, accept ADR-008, or authorize an implementation stage, provider, schema, or operational change.
Product-risk refinement is closed and remains part of the MVP governance baseline.

## Completed audit remediation

Stage 3.27 is an audit-remediation slice for import financial identity and cash-flow semantics. The
repository audit identified three P1 defects: broker-operation identity was not persisted through the
append boundary, cash near-match classification omitted the transaction amount, and deposits/withdrawals
accepted fee fields whose economic effect was undefined. The remediation uses versioned persisted import
identity, amount-aware cash reconciliation, and fail-closed zero-fee cash-flow semantics. Detailed root
cause, design rationale, migration impact, regression coverage, and verification evidence are recorded in
[`docs/stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md`](docs/stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md).
Stage 3.27 is closed for P1-02, P1-03, and P1-04 after implementation PR #55 was squash-merged into `develop` at `6e8c806de857f844954f1db513487357dfe90187` following exact-head CI #90, renewed independent `APPROVED` review on `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c`, explicit human merge approval, and closure governance recorded through PR #58.

Stage 3.28 subsequently remediated the remaining P1-01 refresh-token replay/session-family issue and P1-05 Argon2 resource-admission issue; implementation PR #59 was squash-merged into `develop` at `dc83f5f3a11da164e6809593861d96ccf47b29ca` after exact-head CI #114, renewed independent `APPROVED` review on `92edab5d3e93dafe2fcc6247644e38e878a4202f`, and explicit human merge approval. Detailed root cause, failure modes and security impact, chosen remediation, decision rationale, rejected alternatives, concurrency trade-offs, regression evidence, review history, and verification evidence are recorded in [`docs/stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md`](docs/stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md). Closure governance is recorded through PR #60 when this documentation is canonical on `develop`. Stage 3.25 privacy evidence planning and the P2/P3 audit backlog remain separate.

## Components

- `backend-go/` — Go 1.24+ API using Fiber.
- `frontend-next/` — Next.js App Router, TypeScript, and pnpm Web presentation layer.
- `microservice-python/` — FastAPI analytics worker skeleton.
- `infrastructure/` — local infrastructure configuration.
- `docs/` — frozen architecture and architecture decision records.

## Root commands

Use pnpm from the repository root for common local workflows:

```bash
pnpm run infra:up
pnpm run dev:api
pnpm run dev:web
pnpm run verify
pnpm run verify:e2e
```

`dev:api` and `dev:web` are intentionally separate long-running commands. Run them in two terminal
tabs when manually checking the Web UI.

`dev:api` sets `OPENINVEST_ENV=development` by default so the local auth bypass and insecure local
refresh cookie cannot accidentally become production behavior. An API process without `DATABASE_URL`
starts only with an explicit `OPENINVEST_ENV=development` or `local`; staging and production fail
closed. Production or staging runs with a configured `DATABASE_URL` must keep unsafe local auth flags disabled and provide
`OPENINVEST_ACCESS_TOKEN_SECRET` and `OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET`. Both secrets must
contain at least 32 high-entropy bytes and must be different values.

## Local checks

```bash
pnpm run verify
```

For the Stage 3.4 vertical-slice smoke proof, run:

```bash
pnpm run verify:e2e
```

If local port `5432` is already used by another PostgreSQL process, use an alternate local port:

```bash
POSTGRES_PORT=55432 pnpm run verify:e2e
```

The smoke script starts its own Go API on `http://localhost:8080` with the matching `DATABASE_URL`.
Stop any already-running local Go API before running `verify:e2e`.

## Local infrastructure

Copy `.env.example` to `.env`, replace the development-only values, then run:

```bash
docker compose up -d
```

Architecture changes require an ADR and Source of Truth update. Start with `docs/SOURCE_OF_TRUTH.md` and `docs/ARCHITECTURE_FREEZE_v1.2.md`.

Implementation progress and completed-stage reports are recorded in `docs/IMPLEMENTATION_LOG.md`.
Product risk decisions are recorded in `docs/product/MVP_PRODUCT_RISK_REFINEMENT.md`.
