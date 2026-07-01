# Stage 3.4 — End-to-End Verification

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-04 |
| Version | 0.1.1 |
| Status | Complete / closed |
| Owner | Builder Engineer |
| Supersedes | Stage 3.4 roadmap placeholder |
| Dependencies | Stage 3.1; Stage 3.2; Stage 3.3; ADR-006; ADR-007 |
| Last Review Date | 2026-07-02 |
| Next Review Date | 2027-01-02 |

## Purpose

Stage 3.4 proves that the first vertical slice works as one local path:

```text
Next.js presentation shell
→ Go API
→ PostgreSQL
→ immutable transaction append
→ snapshot rebuild
→ API summary response
→ Web-ready rendered state
```

This stage is intentionally short. It does not add new product features.

## Scope

Allowed:

- root-level developer commands;
- local smoke verification script;
- README/onboarding updates;
- implementation-log and roadmap synchronization;
- evidence that the existing Stage 3.1–3.3 slice can run locally.

Forbidden:

- new business logic;
- new portfolio/domain services;
- SQL migrations;
- provider integrations;
- workers;
- authentication implementation;
- frontend feature expansion;
- mobile implementation;
- Stage 3.5 implementation.

## Root commands

The root `package.json` provides the canonical local developer entrypoints:

```bash
pnpm run infra:up
pnpm run dev:api
pnpm run dev:web
pnpm run verify
pnpm run verify:e2e
```

`pnpm run verify:e2e` runs `scripts/stage-03-04-smoke.sh`.

If local port `5432` is already occupied, run the smoke script with an alternate PostgreSQL port:

```bash
POSTGRES_PORT=55432 pnpm run verify:e2e
```

## Smoke verification contract

The smoke script:

1. starts local PostgreSQL and Redis through Docker Compose;
2. applies the Stage 3.1 migration only when the schema is missing;
3. fails if another API is already responding at `http://localhost:8080`;
4. starts a controlled Go API process with the exported `DATABASE_URL`;
5. creates a portfolio through `POST /api/v1/portfolios`;
6. appends a cash deposit through `POST /transactions`;
7. appends a BUY transaction through `POST /transactions`;
8. reads `GET /summary`;
9. reads `GET /transactions`;
10. verifies expected decimal-string summary values and transaction count;
11. verifies in Docker PostgreSQL that the created portfolio, transactions, and calculated snapshots
    exist in the same database prepared by the script.

The script deliberately verifies backend-owned financial output instead of recalculating portfolio
values in the Web layer.

## Expected local evidence

Successful smoke output must prove:

- PostgreSQL is reachable;
- migrations are present;
- Go API readiness succeeds;
- portfolio data is persisted;
- immutable transaction entries are appended;
- snapshot rebuild produces the expected Stage 3.2 methodology version;
- summary values are returned as decimal strings;
- Docker PostgreSQL contains the created portfolio, transaction entries, and calculated snapshots.

## Boundary confirmation

Stage 3.4 does not change the Next.js implementation. Next.js remains a presentation layer only:

- no Route Handlers for business domains;
- no Server Actions with business behavior;
- no direct PostgreSQL, Redis, MOEX, CBR, Rosstat, broker, or provider access;
- no LocalStorage, SessionStorage, or IndexedDB business persistence;
- no financial, tax, dividend, or inflation calculations.

## Remaining risks

- `pnpm run verify` covers Web typecheck/build. `pnpm run verify:e2e` covers the local
  PostgreSQL/Go API/transaction/snapshot/summary smoke path. Stage 3.4 does not add a browser
  automation dependency. Browser-click automation remains a candidate for a later dedicated test
  stage if the project accepts the dependency and maintenance cost.
- The local API still uses the Stage 3 development subject boundary. Production authentication is
  intentionally outside Stage 3.4.

## Completion checklist

- `pnpm run verify` — passed.
- `POSTGRES_PORT=55432 pnpm run verify:e2e` — passed.
- forbidden-boundary scan — passed.
- Internal Review Agent review — approved after REQUEST CHANGES fixes.
- Draft PR review — approved.

## Closure evidence

- PR: #13 — `feature/stage-03-04-e2e-verification` → `develop`
- Merge method: squash merge
- Merge commit: `86582efaa420b2c38465a5d0da041814149392c7`
- CI status before merge: green
- Scope result: verification and onboarding only; no business logic, SQL migrations, provider
  integrations, workers, frontend feature expansion, mobile code, or Stage 3.5 implementation.
