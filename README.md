# OpenInvest

OpenInvest is an independent, privacy-first investment analytics platform. It is not a broker, bank, asset manager, trading system, or investment adviser.

This repository contains the closed Stage 2 contract/canonical-model baseline, the accepted Next.js
Web presentation-layer baseline, the merged Stage 3.2 Go API vertical-slice backend, the closed
Stage 3.3 Web presentation slice, the closed Stage 3.4 end-to-end verification layer, the closed
Stage 3.5 broker-file import/reconciliation design, and the closed Stage 3.6 import reconciliation
slice.

Stage 3.6 added an internal CSV parser/review/append-plan slice only: no public import API, broker
API, upload UI, SQL migration, worker, or automatic ledger append was introduced.
Stage 3.7 is closed and added internal atomic append of user-approved import rows with atomic
PostgreSQL persistence, duplicate revalidation, idempotency protection, audit evidence, and
deterministic snapshot rebuilds. Stage 3.8 planning is active and is limited to defining the future
internal import review → append flow integration. Public import API, upload UI, import-session
persistence, workers, provider integrations, tax, mobile, and AI remain out of scope.
Product-risk refinement is closed and remains part of the MVP governance baseline.

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
