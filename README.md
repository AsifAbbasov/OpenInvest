# OpenInvest

OpenInvest is a privacy-first **Personal Capital Operating System** for tracking investments, understanding returns, dividends and inflation, and preparing reliable financial information.

It is not a broker, bank, trading terminal, asset manager, investment adviser, or tax service. OpenInvest does not execute trades or provide investment recommendations.

> **Project status:** Stage 1 documentation consolidation is approved. The current bootstrap transition adds the first CI quality gates and the non-business OpenAPI skeleton. Product business logic has not been implemented yet.

## What we are building

The MVP is intended to give an investor one trustworthy view of personal capital:

- registration with privacy-preserving defaults;
- portfolios and immutable transaction history;
- MOEX stock and bond cards;
- RUB cash tracking;
- weighted-average cost and portfolio snapshots;
- XIRR, real return and inflation-adjusted return;
- dividend calculator and dividend calendar;
- purchasing-power analysis;
- a dashboard built from server-calculated financial results.

AI assistants, forecasts, scenario analysis, premium analytics, public APIs, foreign securities, tax XML export, family accounts and mobile applications are deliberately outside the MVP.

## Product principles

- **Correctness before features.** Financial values use decimal arithmetic, half-even rounding and canonical test vectors.
- **Privacy by Design.** Only necessary data is collected; identity and financial records are isolated and deletion must prevent re-identification.
- **API First.** HTTP contracts are reviewed in OpenAPI before business endpoints are implemented.
- **Canonical data.** PostgreSQL is the durable source of truth; Redis and RAM are acceleration layers only.
- **Reproducible analytics.** Transactions are immutable and snapshots are versioned, deterministic projections.
- **Reliable events.** Delivery is at least once, with outbox/inbox processing and idempotent business effects.
- **Backend-owned calculations.** Clients display server-calculated results and do not duplicate financial logic.

## What exists today

The repository currently proves the selected toolchain without introducing product behavior:

| Area | Implemented | Why it exists now |
| --- | --- | --- |
| Go backend | Minimal Fiber application and `/health` test | Verifies the primary API runtime and testing path |
| Python analytics | Minimal FastAPI application and `/health` test | Verifies the isolated analytics runtime and locked Python environment |
| React frontend | Vite, TypeScript, React and Redux Toolkit skeleton | Verifies the web toolchain and production build path |
| Local infrastructure | PostgreSQL and Redis Docker Compose services | Provides the future canonical database and cache development baseline |
| API contract | OpenAPI 3.1 operational skeleton for `/health` | Establishes the contract location before the Stage 2 API freeze |
| CI | Go tests, Python tests, frontend typecheck/build and Compose validation | Prevents unverified changes from entering protected branches |
| Architecture | Source of Truth, accepted ADRs and review workflow | Keeps implementation aligned with approved product and engineering decisions |

There are no portfolio aggregates, financial algorithms, database migrations, authentication flows, collectors, or other business modules yet. Those are introduced only in their approved stages.

## Repository structure

```text
OpenInvest/
├── backend-go/            # Primary Go/Fiber API
├── frontend-react/        # React/Vite/TypeScript web client
├── microservice-python/   # Python analytics service
├── openapi/               # Versioned HTTP contract
├── infrastructure/        # Infrastructure notes and future deployment assets
├── docs/                  # Architecture, ADRs, stage reports and traceability
├── scripts/               # Automation added only when a concrete workflow needs it
└── docker-compose.yml     # Local PostgreSQL and Redis
```

## Technology baseline

- Go 1.24+ with Fiber;
- Python 3.12+ managed with uv;
- React 19, Vite, TypeScript and Redux Toolkit;
- pnpm 11.8.0 with a frozen lockfile;
- PostgreSQL and Redis;
- Docker Compose;
- OpenAPI 3.1;
- GitHub Actions.

## Local development

Prerequisites: Go, Python/uv, Node.js with Corepack, and Docker Compose.

Create local infrastructure configuration:

```bash
cp .env.example .env
```

Replace the development placeholders in `.env`, then start PostgreSQL and Redis:

```bash
docker compose up -d
```

Run the same baseline checks used by CI:

```bash
(cd backend-go && go test ./...)
(cd frontend-react && corepack pnpm install --frozen-lockfile)
(cd frontend-react && corepack pnpm typecheck && corepack pnpm build)
(cd microservice-python && uv sync --extra dev --locked && uv run pytest)
POSTGRES_PASSWORD=openinvest-local docker compose config --quiet
```

The services are development skeletons; there is no supported end-user workflow yet.

## Delivery workflow

Changes follow this mandatory path:

```text
feature branch → local checks → Draft PR → CI → independent review
→ human approval → squash merge into develop
```

Direct pushes to `develop` and `main` are forbidden. Architecture changes require an ADR, impact analysis and Source of Truth update.

## Roadmap

| Stage | Outcome | Status |
| --- | --- | --- |
| 0 — Foundation | Monorepo skeleton, health checks and local infrastructure | Complete |
| 1 — Documentation Consolidation | Repository-owned architecture and governance | Approved |
| Stage 1 Bootstrap transition | Basic CI and operational OpenAPI skeleton | In review |
| 2 — OpenAPI Freeze | Reviewed MVP contracts, errors, idempotency and validation | Not started |
| 3 — Bootstrap Hardening | Align the skeleton with frozen contracts and repository standards | Planned |
| 4 — Infrastructure | PostgreSQL schemas, migrations, cache boundaries and outbox/inbox | Planned |
| 5 — First Vertical Slice | Transaction-to-dashboard flow with tests | Planned |

Implementation does not advance to the next stage without explicit approval.

## Documentation

Start here:

- [`docs/SOURCE_OF_TRUTH.md`](docs/SOURCE_OF_TRUTH.md) — authoritative product scope, architecture and standards;
- [`docs/REVIEW_WORKFLOW.md`](docs/REVIEW_WORKFLOW.md) — branch, CI, review and merge requirements;
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — approved implementation order;
- [`docs/IMPLEMENTATION_LOG.md`](docs/IMPLEMENTATION_LOG.md) — completed-stage reports and evidence;
- [`docs/DOCUMENT_INDEX.md`](docs/DOCUMENT_INDEX.md) — canonical and legacy documentation registry;
- [`docs/ADR/`](docs/ADR/) — accepted architecture decisions;
- [`openapi/openapi.yaml`](openapi/openapi.yaml) — versioned API contract.

Documents 42–43 and the Source of Truth resolve conflicts in earlier specifications. Legacy documents remain in the repository for decision traceability and must not be treated as higher-priority than the current architecture.

## Current limitations

- no production deployment or runtime SLO evidence;
- no business or financial calculation code;
- no database schema or migration;
- no authentication or authorization implementation;
- no approved external market-data source;
- no mobile, AI, tax-export or public API implementation.

These limitations are intentional at the current stage.
