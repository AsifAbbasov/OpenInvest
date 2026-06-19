# OpenInvest

OpenInvest is an independent, privacy-first investment analytics platform. It is not a broker, bank, asset manager, trading system, or investment adviser.

This repository currently contains the Stage 1 engineering skeleton only. There is no business logic yet.

## Components

- `backend-go/` — Go 1.24+ API using Fiber.
- `frontend-react/` — React 19, Vite, TypeScript, and Redux Toolkit web client.
- `microservice-python/` — FastAPI analytics worker skeleton.
- `infrastructure/` — local infrastructure configuration.
- `docs/` — frozen architecture and architecture decision records.

## Local checks

```bash
cd backend-go && go test ./...
cd frontend-react && pnpm install --frozen-lockfile && pnpm build
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
