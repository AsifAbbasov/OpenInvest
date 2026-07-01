# OpenInvest

OpenInvest is an independent, privacy-first investment analytics platform. It is not a broker, bank, asset manager, trading system, or investment adviser.

This repository contains the closed Stage 2 contract/canonical-model baseline, the accepted Next.js
Web presentation-layer baseline, the merged Stage 3.2 Go API vertical-slice backend, and the closed
Stage 3.3 Web presentation slice.

The current implementation focus is Stage 3.4 end-to-end verification: proving the complete local
path from Next.js through the Go API, PostgreSQL, snapshots, API responses, and rendered Web state.
Product-risk refinement is closed and remains part of the MVP governance baseline.

## Components

- `backend-go/` — Go 1.24+ API using Fiber.
- `frontend-next/` — Next.js App Router, TypeScript, and pnpm Web presentation layer.
- `microservice-python/` — FastAPI analytics worker skeleton.
- `infrastructure/` — local infrastructure configuration.
- `docs/` — frozen architecture and architecture decision records.

## Local checks

```bash
cd backend-go && go test ./...
cd backend-go && go run ./cmd/validate-openapi && go run ./cmd/validate-migrations
cd frontend-next && corepack pnpm install --frozen-lockfile && corepack pnpm run typecheck && corepack pnpm run build
cd microservice-python && uv sync --extra dev --locked && uv run pytest
POSTGRES_PASSWORD=openinvest-local docker compose config --quiet
```

## Local infrastructure

Copy `.env.example` to `.env`, replace the development-only values, then run:

```bash
docker compose up -d
```

Architecture changes require an ADR and Source of Truth update. Start with `docs/SOURCE_OF_TRUTH.md` and `docs/ARCHITECTURE_FREEZE_v1.2.md`.

Implementation progress and completed-stage reports are recorded in `docs/IMPLEMENTATION_LOG.md`.
Product risk decisions are recorded in `docs/product/MVP_PRODUCT_RISK_REFINEMENT.md`.
