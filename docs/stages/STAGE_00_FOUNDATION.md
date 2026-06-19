# Stage 0 — Foundation

**Status:** Complete
**Completed:** 2026-06-19
**Source scope:** “First Implementation Stage” in Architecture Freeze v1.0

## Why

Create a small, reproducible foundation that proves the frozen technology choices can coexist before any financial business logic is introduced.

## Scope

- Initialize the monorepo and `develop` branch.
- Record the Architecture Freeze and initial ADRs.
- Create minimal Go/Fiber, React/Vite/TypeScript/Redux Toolkit, and Python/FastAPI services.
- Configure local PostgreSQL and Redis.
- Standardize package management and reproducible dependency locks.
- Add only health-level checks; no product business logic.

## Decisions

- Frontend dependencies use pnpm with a committed `pnpm-lock.yaml` and a pinned package-manager version.
- Python dependencies and environments use uv with a committed `uv.lock`.
- Go dependencies use Go Modules with committed `go.mod` and `go.sum`.
- Package-manager choices are engineering workflow standards and do not alter the fundamental frozen architecture.

## Completed work

- Repository structure and documentation folders created.
- ADR template and ADR-001 through ADR-005 added.
- Minimal Go/Fiber API and health check added.
- Minimal React 19/Vite/TypeScript/Redux Toolkit application added.
- Minimal FastAPI analytics service and health check added.
- Local Docker Compose services for PostgreSQL and Redis added.
- Local secrets excluded; `.env.example` contains development placeholders only.
- Dependency and build artifacts excluded from Git.

## Verification

- `go test ./...`
- `go vet ./...`
- `pnpm build`
- `uv run pytest`
- `docker compose config --quiet`

## Known risks

- Docker service startup and runtime health have not yet been exercised; only Compose validation is in Stage 0.
- CI is not present yet.
- FastAPI/Starlette currently emits an upstream TestClient deprecation warning about the future `httpx2` transition.

## Recommended next step

Commit the foundation after review, then define the next approved stage around CI and the first OpenAPI contract without adding financial business logic prematurely.
