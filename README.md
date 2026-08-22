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