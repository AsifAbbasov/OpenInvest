# OpenInvest Implementation Log

| Field | Value |
| --- | --- |
| Document ID | REG-IMP-001 |
| Version | 1.1.17 |
| Status | Active |
| Owner | Builder Engineer |
| Supersedes | Informal stage-status notes |
| Dependencies | `SOURCE_OF_TRUTH.md`; `REVIEW_WORKFLOW.md` |
| Last Review Date | 2026-07-01 |
| Next Review Date | 2027-01-01 |

This log is the index of implementation stages. Every stage must document its purpose, scope, decisions, completed work, verification, known risks, and recommended next step. At the end of each stage, implementation stops for a user-facing report and confirmation before any push.

| Stage | Purpose | Status | Report |
| --- | --- | --- | --- |
| 0 — Foundation | Establish a reproducible, architecture-aligned repository skeleton | Complete | [Stage 0 report](stages/STAGE_00_FOUNDATION.md) |
| 1 — Documentation Consolidation | Establish the repository-owned Source of Truth and freeze v1.2 | Complete; awaiting review | [Stage 1 report](stages/STAGE_01_DOCUMENTATION_CONSOLIDATION.md) |
| 2 — Contract and Canonical Model Freeze | Freeze the MVP API, canonical DTOs, logical ER model, and migration strategy | Complete / closed; merged into `develop` at `bfde623552ebea6eac7bdaabf0d1a2263883de12` | [Stage 2 report](stages/STAGE_02_CONTRACT_AND_CANONICAL_MODEL.md) |
| Web architecture amendment | Replace the Web skeleton with presentation-only Next.js under ADR-007 | Complete / closed; merged into `develop` at `6a7748cc24fc852d42b90b0e0cb843b6020f3973` | [Amendment report](stages/WEB_FRONTEND_ARCHITECTURE_AMENDMENT.md) |
| 3 — First Vertical Slice | Implement the first thin MVP path after contract and Web baseline approval | Planning complete; implementation split into small PRs | [Stage 3 plan](stages/STAGE_03_FIRST_VERTICAL_SLICE.md) |
| 3.1 — Local Database Foundation | Add minimal PostgreSQL structures and migration validation for the first vertical slice | Complete / closed; merged into `develop` at `b1a3f23` | [Stage 3.1 report](stages/STAGE_03_01_DATABASE_FOUNDATION.md) |
| 3.2 — Go API Vertical-Slice Backend | Implement portfolio create, transaction append, snapshot rebuild, and summary read in Go | Complete / closed; merged into `develop` at `8971918c8046fb9a2d6bf9f97897432cf08fbde1` | [Stage 3.2 report](stages/STAGE_03_02_GO_API_VERTICAL_SLICE.md) |
| Product risk refinement | Convert hard PRD criticism into controlled MVP risk decisions | Complete / closed; merged into `develop` at `65bdf6537b44ed57e1c00bf68d2dacd70aa09702` | [MVP product risk refinement](product/MVP_PRODUCT_RISK_REFINEMENT.md) |
| 3.3 — Next.js Presentation Slice | Render the first Web path through the Go API only | Complete / closed; implementation merged into `develop` at `11805cc298bba13f09f7f7af8b1e1178dc351209`; closure docs merged at `fe402030359459f909c156a1e993f18ceed257bf` | [Stage 3.3 report](stages/STAGE_03_03_NEXTJS_PRESENTATION_SLICE.md) |
| 3.4 — End-to-End Verification | Prove the full local path from Next.js to Go API, PostgreSQL, snapshots, API responses, and rendered Web state | Complete / closed; merged into `develop` at `86582efaa420b2c38465a5d0da041814149392c7` | [Stage 3.4 report](stages/STAGE_03_04_END_TO_END_VERIFICATION.md) |
| 3.5 — Broker File Import and Reconciliation Design | Design the safe user-supplied broker-file import path before parser implementation | Complete / closed; merged into `develop` at `072d38d94b529221d6467502f82f03a674a7d805` | [Stage 3.5 report](stages/STAGE_03_05_BROKER_FILE_IMPORT_RECONCILIATION_DESIGN.md) |
| 3.6 — Broker File Import Reconciliation Slice | Parse CSV broker files into reviewable normalized candidates and explicit append plans | Complete / closed; merged into `develop` at `e2b05650a4422b97d4bd924254367106b6a4686b` | [Stage 3.6 report](stages/STAGE_03_06_IMPORT_RECONCILIATION_SLICE.md) |
| 3.7 — Import Append Planning | Define the safe atomic append boundary before any import ledger mutation implementation | Complete / closed; merged into `develop` at `36d86c7ff2a9c75478de155d4f60b979b8da9376` | [Stage 3.7 plan](stages/STAGE_03_07_IMPORT_APPEND_PLANNING.md) |
| 3.7 — Import Append Slice | Internally append user-approved import rows atomically into the immutable ledger | Complete / closed; merged into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720` | [Stage 3.7 implementation report](stages/STAGE_03_07_IMPORT_APPEND_SLICE.md) |
| 3.8 — Import Review Append Flow Planning | Define the internal orchestration from reviewed import candidates to atomic append | Complete / closed; merged into `develop` at `a35af2f5207bd564647d2a3fc032f4f940e62ddd` | [Stage 3.8 plan](stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_PLANNING.md) |
| 3.8 — Import Review Append Flow Slice | Internally orchestrate import review decisions into atomic append | Complete / closed; merged into `develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79` | [Stage 3.8 implementation report](stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_SLICE.md) |
| 3.9 — Import API Boundary Planning | Define future public Go API boundary for user-supplied broker-file import | Complete / closed; merged into `develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab` | [Stage 3.9 plan](stages/STAGE_03_09_IMPORT_API_BOUNDARY_PLANNING.md) |
| 3.9 — Import API Boundary Slice | Expose user-supplied CSV import review/append through the Go API boundary | Complete / closed; merged into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898` | [Stage 3.9 implementation report](stages/STAGE_03_09_IMPORT_API_BOUNDARY_SLICE.md) |
| 3.10 — Import Upload/Review UI Planning | Define the future Next.js presentation-only import upload/review UI boundary | Complete / closed; merged into `develop` at `27480d6ff22e2929e33aeac352aef8a1b01bb448` | [Stage 3.10 plan](stages/STAGE_03_10_IMPORT_UPLOAD_UI_PLANNING.md) |
| 3.10 — Import Upload/Review UI Slice | Expose CSV import review/append through the Next.js presentation layer only | Complete / closed; merged into `develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71` | [Stage 3.10 implementation report](stages/STAGE_03_10_IMPORT_UPLOAD_REVIEW_UI_SLICE.md) |
| 3.11 — Authentication and Privacy-Boundary Planning | Define the future auth/session/privacy-default implementation boundary before replacing the local development subject | Complete / closed; merged into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125` | [Stage 3.11 plan](stages/STAGE_03_11_AUTH_PRIVACY_PLANNING.md) |
| 3.11 — Authentication and Privacy-Boundary Slice | Implement the approved Go API auth/session/privacy-default boundary without frontend auth UI | Complete / closed; merged into `develop` at `5c49173ac858995929f266c2de991282dd194dec` | [Stage 3.11 implementation report](stages/STAGE_03_11_AUTH_PRIVACY_SLICE.md) |

## Stage completion protocol

1. Finish only the approved stage scope.
2. Run checks proportionate to the changes.
3. Update this log and the stage report.
4. Report created or changed files, commands, checks, risks, and the recommended next step.
5. Stop and request explicit confirmation before commit/push or the next stage when required.

## 2026-07-02 — Stage 3.6 closed

- Squash-merged PR #15 into `develop` at `e2b05650a4422b97d4bd924254367106b6a4686b`.
- Closed the Stage 3.6 broker-file import reconciliation slice after green CI, requested-changes
  fixes, and independent follow-up review approval.
- Added internal CSV parse, normalization, duplicate/conflict detection, safe review model, and
  explicit append-plan generation.
- Kept public import endpoints, upload UI, SQL import-session tables, workers, broker/provider
  integrations, credential scraping, XLSX/PDF parsing, automatic ledger append, and Stage 3.7 work
  out of scope.

## 2026-07-02 — Stage 3.7 import append planning started

- Started the documentation-only Stage 3.7 planning scope for atomic append of user-approved import
  rows into the immutable ledger.
- Kept Stage 3.7 implementation, public import endpoints, upload UI, SQL import-session
  persistence, broker/provider integrations, workers, tax logic, mobile, and AI out of scope.
- Required a separate reviewed implementation PR before any imported rows can mutate the ledger.

## 2026-07-02 — Stage 3.7 import append slice started

- Started the internal implementation slice for atomic append of user-approved import rows.
- Added no public import API, frontend upload UI, SQL import-session table, worker, provider
  integration, tax logic, mobile, or AI scope.
- Required live PostgreSQL verification for atomicity, idempotency, snapshot rebuild, rollback, and
  audit evidence.

## 2026-07-02 — Stage 3.7 import append slice closed

- Squash-merged PR #18 into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`.
- Added internal atomic append of user-approved import rows with duplicate revalidation,
  idempotency protection, minimal audit evidence, and deterministic snapshot rebuilds.
- Resolved independent review findings for concurrent duplicate-batch serialization and live
  PostgreSQL evidence.
- Kept public import endpoints, upload UI, SQL import-session persistence, workers,
  broker/provider integrations, tax logic, mobile, AI, and Stage 3.8 work out of scope.

## 2026-07-03 — Stage 3.8 import review append flow planning started

- Started the documentation-only Stage 3.8 planning scope for the future internal orchestration
  between Stage 3.6 import review output and Stage 3.7 atomic append.
- Kept public import endpoints, OpenAPI changes, upload UI, SQL import-session persistence, raw
  file persistence, workers, broker/provider integrations, tax logic, mobile, AI, and implementation
  out of scope.

## 2026-07-03 — Stage 3.8 import review append flow slice started

- Started the internal implementation slice for parse/review/approve/append orchestration.
- Added no public import API, OpenAPI changes, upload UI, SQL import-session table, raw file
  persistence, worker, provider integration, tax logic, mobile, or AI scope.
- Required live PostgreSQL verification for full parse/review/approve/append behavior and stale
  duplicate rollback.

## 2026-07-03 — Stage 3.8 import review append flow slice closed

- Squash-merged PR #21 into `develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`.
- Added internal import review → append orchestration with bounded in-memory payload handling,
  explicit approved decisions, Stage 3.7 atomic append invocation, and non-sensitive result
  metadata.
- Resolved independent review finding that initially exposed full review rows through the append
  result.
- Kept public import endpoints, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, AI, and Stage 3.9 work out of scope.

## 2026-07-08 — Stage 3.9 import API boundary planning started

- Started the documentation-only Stage 3.9 planning scope for a future public Go API boundary for
  user-supplied broker-file import.
- Kept OpenAPI changes, Go handlers, frontend upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax logic, mobile, AI, and implementation out of
  scope.

## 2026-07-08 — Stage 3.9 import API boundary slice started

- Squash-merged PR #23 into `develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`.
- Started the implementation slice for public Go API import review/append endpoints.
- Scoped the implementation to OpenAPI contract additions, Go HTTP handlers, DTOs, tests, and
  documentation only.
- Chose a stateless review/append boundary for this slice: no review IDs, no import-session table,
  no raw CSV persistence, and append reruns review before invoking the atomic store boundary.
- Kept frontend upload UI, SQL import-session persistence, workers, broker/provider integrations,
  tax, mobile, AI, and Stage 3.10 work out of scope.

## 2026-07-08 — Stage 3.9 import API boundary slice closed

- Squash-merged PR #24 into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`.
- Closed the Stage 3.9 public Go API boundary after green CI and independent external review
  approval.
- Added transient CSV import review and explicit append endpoints backed by the Stage 3.8
  review→append flow and Stage 3.7 atomic store append.
- Resolved independent review findings for current-ledger revalidation, exact replay of original
  imported transactions, complete append idempotency hashing, and append payload validator coverage.
- Kept frontend upload UI, SQL import-session persistence, raw CSV persistence, workers,
  broker/provider integrations, tax, mobile, AI, and Stage 3.10 implementation out of scope.

## 2026-07-08 — Stage 3.10 import upload/review UI planning started

- Squash-merged Stage 3.9 closure governance into `develop` at
  `682ffd856395a6e3e988817551a512898fda2d38`.
- Started the documentation-only Stage 3.10 planning scope for a future Web import upload/review UI.
- Kept Next.js implementation, OpenAPI changes, Go handlers, SQL migrations, import-session
  persistence, business logic, provider integrations, workers, tax, mobile, and AI out of scope.

## 2026-07-09 — Stage 3.10 import upload/review UI slice started

- Squash-merged Stage 3.10 planning into `develop` at
  `27480d6ff22e2929e33aeac352aef8a1b01bb448`.
- Started the implementation slice for a presentation-only Next.js CSV import upload/review panel.
- Scoped implementation to typed Go API calls, transient in-memory file handling, review display,
  explicit row approval, append submission, and UI feedback only.
- Kept OpenAPI changes, backend handlers, SQL migrations, import-session persistence, raw file
  persistence, provider integrations, workers, tax, mobile, and AI out of scope.

## 2026-07-09 — Stage 3.10 import upload/review UI slice closed

- Squash-merged PR #27 into `develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71`.
- Added the presentation-only Next.js import upload/review panel over the existing Go import API.
- Resolved independent review findings for duplicate-row React keys and file-input clearing after
  successful append or oversized-file rejection.
- Kept OpenAPI changes, backend handlers, SQL migrations, import-session persistence, raw file
  persistence, provider integrations, workers, tax, mobile, AI, and Stage 3.11 implementation out
  of scope.

## 2026-07-09 — Stage 3.11 authentication and privacy-boundary planning started

- Started the documentation-only planning scope for replacing the local development subject with
  the approved MVP web auth, session, CSRF, and privacy-default boundary.
- Kept auth implementation, schema migrations, password hashing, token issuance, frontend session
  code, business logic, workers, tax, mobile, AI, and provider integrations out of scope.

## 2026-07-09 — Stage 3.11 authentication and privacy-boundary slice started

- Squash-merged PR #28 into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125`.
- Started the implementation slice for the approved Go API auth/privacy boundary.
- Added no frontend auth UI, email verification, OAuth/passkeys/2FA, worker, provider integration,
  tax logic, mobile implementation, AI functionality, or Stage 3.12 scope.
- Required strict independent review to verify refresh-token secrecy, CSRF enforcement, replay
  rejection, privacy defaults, migration safety, and absence of Next.js business logic.

## 2026-07-09 — Stage 3.11 independent review findings fixed

- Added a production guard requiring `OPENINVEST_ENV=development` or `local` before unsafe local auth
  flags can run with a configured `DATABASE_URL`.
- Added non-secret audit events for auth/session lifecycle actions and rejected refresh/logout
  attempts, including missing cookie/CSRF paths that return before token lookup.
- Added `Retry-After` to rate-limited auth responses and removed logout rate limiting to preserve
  the frozen logout contract.
- Tightened auth email validation to reject display-name forms and addresses above the OpenAPI
  254-character limit.

## 2026-07-09 — Stage 3.11 authentication and privacy-boundary slice closed

- Squash-merged PR #29 into `develop` at `5c49173ac858995929f266c2de991282dd194dec`.
- Closed the Stage 3.11 Go API auth/privacy boundary after green GitHub CI on
  `8a8052c18768dbad0aa0e724836f3c9252d257e3` and strict independent review with no remaining code
  findings.
- Added registration, login, refresh, and logout handlers; Argon2id password hashing; short-lived
  access tokens; rotating HttpOnly refresh sessions; CSRF enforcement; privacy-default persistence;
  additive identity schema migration; and non-secret auth/session audit evidence.
- Kept frontend auth UI, email verification, OAuth/passkeys/2FA, workers, provider integrations,
  tax, mobile, AI, and Stage 3.12 out of scope.
