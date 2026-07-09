# Stage 3.10 — Import Upload and Review UI Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-10-UI |
| Version | 0.1.1 |
| Status | Complete / merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Stage 3.10 planning-only state |
| Dependencies | Stage 3.10 import upload/review UI planning; Stage 3.9 import API boundary slice; ADR-007 |
| Last Review Date | 2026-07-09 |
| Next Review Date | Before changing import-session persistence or browser file-retention behavior |

## Purpose

Stage 3.10 implements the smallest Web presentation path for user-supplied CSV import review and
explicit append.

The UI exists to make the already-approved Go import API usable from the portfolio detail page. It
does not parse CSV, reconcile transactions, calculate financial values, mutate the ledger directly,
persist raw files, or replace the Go API.

## Implemented scope

The implementation adds:

- a portfolio-detail import panel for CSV file selection;
- a typed frontend API boundary for the existing Go import review and append endpoints;
- transient review display for appendable, duplicate, conflict, and invalid rows;
- explicit user selection of appendable rows;
- append submission for selected `APPROVE` decisions only;
- success and error presentation states;
- CSS for import review status and success feedback;
- documentation and governance updates.

## Merge and review evidence

- PR: #27
- Branch: `feature/stage-03-10-import-upload-review-ui`
- Merge target: `develop`
- Merge commit: `e19a1a0ea4b0b183687bd89daabdfbc973daea71`
- Local verification: `pnpm run verify`
- GitHub CI: Go tests, Python tests, Frontend build/typecheck, OpenAPI contract, PostgreSQL
  migration validation, and Docker Compose config passed.
- Independent review: approved after fixes for duplicate-row React keys and file-input clearing.

## Route and navigation decision

Stage 3.10 uses the existing portfolio detail route:

```text
/portfolios/[portfolioId]
```

The import panel is colocated with the portfolio ledger view because the import action is portfolio
scoped. A dedicated import route is deferred until the UI becomes large enough to justify a separate
navigation surface.

## Boundary guarantees

- Next.js remains presentation only.
- The browser reads the selected file only to send text to the Go API.
- Raw CSV content is held only in React component state for the current interaction.
- Raw CSV content is cleared from component state after successful append.
- No `localStorage`, `sessionStorage`, IndexedDB, cookies, Route Handlers, Server Actions, database
  access, external provider calls, or workers are introduced.
- Review is preflight only; append reruns backend validation.
- Only selected appendable rows are sent as `APPROVE` decisions.
- Non-selected rows are not converted into implicit business decisions.

## Files changed

- `frontend-next/src/common/api/openinvest.ts`
- `frontend-next/src/features/portfolio/components/ImportUploadReviewPanel.tsx`
- `frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx`
- `frontend-next/src/app/styles.css`
- governance documentation listed in this report's PR.

## Verification plan

Before merge, this slice must pass:

- `git diff --check`;
- `pnpm run verify`;
- GitHub CI;
- strict independent review confirming that the Web layer remains presentation-only.

## Known limitations

- The UI currently supports CSV only.
- There is no import-session persistence; this is intentional for the stateless Stage 3.9 API.
- Refreshing or navigating away clears the selected file and review state.
- There is no browser E2E test harness yet; verification relies on typecheck/build, API boundary
  tests, and review for this slice.

## Forbidden scope not added

Stage 3.10 does not add:

- backend handlers;
- OpenAPI changes;
- SQL migrations;
- import-session tables;
- raw file persistence;
- CSV parsing business logic in Next.js;
- financial calculations;
- tax logic;
- provider integrations;
- workers;
- mobile code;
- AI assistance.
