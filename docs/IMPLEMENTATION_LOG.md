# OpenInvest Implementation Log

| Field | Value |
| --- | --- |
| Document ID | REG-IMP-001 |
| Version | 1.1.57 |
| Status | Current |
| Owner | Builder Engineer |
| Supersedes | Informal stage-status notes |
| Dependencies | `SOURCE_OF_TRUTH.md`; `REVIEW_WORKFLOW.md` |
| Last Review Date | 2026-08-23 |
| Next Review Date | Before Stage 3.25 evidence-collection plan review, evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

This log is the index of implementation stages. Every stage must document its purpose, scope, decisions, completed work, verification, known risks, and recommended next step. At the end of each stage, implementation stops for a user-facing report and confirmation before any push.

| Stage | Purpose | Status | Report |
| --- | --- | --- | --- |
| 0 — Foundation | Establish a reproducible, architecture-aligned repository skeleton | Complete | [Stage 0 report](stages/STAGE_00_FOUNDATION.md) |
| 1 — Documentation Consolidation | Establish the repository-owned Source of Truth and freeze v1.2 | Complete; awaiting review | [Stage 1 report](stages/STAGE_01_DOCUMENTATION_CONSOLIDATION.md) |
| 2 — Contract and Canonical Model Freeze | Freeze the MVP API, canonical DTOs, logical ER model, and migration strategy | Complete / closed; merged into `develop` at `bfde623552ebea6eac7bdaabf0d1a2263883de12` | [Stage 2 report](stages/STAGE_02_CONTRACT_AND_CANONICAL_MODEL.md) |
| Web architecture amendment | Replace the Web skeleton with presentation-only Next.js under ADR-007 | Complete / closed; merged into `develop` at `6a7748cc24fc852d42b90b0e0cb843b6020f3973` | [Amendment report](stages/WEB_FRONTEND_ARCHITECTURE_AMENDMENT.md) |
| 3 — First Vertical Slice | Implement the first thin MVP path after contract and Web baseline approval | Implementation closed through Stage 3.16 audit-fix closure; Stage 3.17-3.24 proposals are merged; Stage 3.25 privacy evidence planning remains separate; Stages 3.27-3.31 audit remediation are closed; Stage 3.32 P2-09/P2-13 implementation is merged and closure governance is tracked on the current docs branch | [Stage 3 plan](stages/STAGE_03_FIRST_VERTICAL_SLICE.md) |
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
| 3.12 — Web Authentication UI Planning | Define the future Next.js presentation-only auth/session UI boundary before implementation | Complete / closed; merged into `develop` at `25be13ce84844562e0381b79f4b81cbfed7eb44d` | [Stage 3.12 plan](stages/STAGE_03_12_AUTH_UI_PLANNING.md) |
| 3.12 — Web Authentication UI Slice | Implement the approved Next.js presentation-only auth/session UI boundary | Complete / closed; merged into `develop` at `b4840b60346109e3cd54a07d9e1e131fc0cfad23` | [Stage 3.12 implementation report](stages/STAGE_03_12_AUTH_UI_SLICE.md) |
| 3.13 — Instrument Catalog Planning | Define the canonical MVP MOEX share/bond identity boundary before implementation | Complete / closed; merged into `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92` | [Stage 3.13 plan](stages/STAGE_03_13_INSTRUMENT_CATALOG_PLANNING.md) |
| 3.13 — Instrument Catalog Slice | Resolve approved MOEX share/bond tickers through the backend-owned catalog boundary | Complete / closed; merged into `develop` at `b9c05fb14d0ee03e6de4dfc04ff67c16da33040b` | [Stage 3.13 implementation report](stages/STAGE_03_13_INSTRUMENT_CATALOG_SLICE.md) |
| 3.14 — Asset Search/Card API Boundary Planning | Define the future Go API asset search/detail boundary before implementation | Complete / closed; merged into `develop` at `2c4f7853599a455bb0cc04114b338a1145baf39c` | [Stage 3.14 plan](stages/STAGE_03_14_ASSET_API_BOUNDARY_PLANNING.md) |
| 3.14 — Asset Search/Card API Boundary Slice | Expose the public Go API asset search boundary without fabricated market data or detail provenance | Complete / closed; merged into `develop` at `57a9404952cb65693614109dd4a14d41fa5c4295` | [Stage 3.14 implementation report](stages/STAGE_03_14_ASSET_API_BOUNDARY_SLICE.md) |
| 3.15 — Web Asset Discovery UI Planning | Define the future Next.js presentation-only asset search entry and honest deferred card-state boundary | Complete / closed; merged into `develop` at `dfeab109b2825fe0e0317e87a7abf2e706a29ea6` | [Stage 3.15 plan](stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_PLANNING.md) |
| 3.15 — Web Asset Discovery UI Slice | Implement the reviewed Next.js presentation-only asset discovery boundary | Complete / closed; merged into `develop` at `22bede651a646d0e8b06568bda457d0626891e63` | [Stage 3.15 implementation report](stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_SLICE.md) |
| 3.16 — Repository Audit Planning | Plan the mandatory full repository audit before the next implementation stage | Complete / merged into `develop` at `74eebe9ec8231764f21ce384c4690d073d0273da` | [Stage 3.16 plan](stages/STAGE_03_16_REPOSITORY_AUDIT_PLANNING.md) |
| 3.16 — Repository Audit Report | Record mandatory full repository audit coverage, manifest, and verdict | Complete / returned `REQUEST CHANGES` | [Stage 3.16 audit report](stages/STAGE_03_16_REPOSITORY_AUDIT_REPORT.md) |
| 3.16 — Repository Audit Fixes | Fix mandatory repository audit `REQUEST CHANGES` findings | Complete / closed; merged into `develop` at `9e6b8a753bf73ef020ce40461df25a5878344d92` | [Stage 3.16 audit fixes](stages/STAGE_03_16_REPOSITORY_AUDIT_FIXES.md) |
| 3.17 — Privacy Lifecycle Planning | Define the future account-deletion, anonymization, backup-destruction, and retention execution boundary | Complete / merged through PR #46 at `1e8c240` | [Stage 3.17 plan](stages/STAGE_03_17_PRIVACY_LIFECYCLE_PLANNING.md) |
| 3.18 — Privacy Contract and Security Proposal | Define the candidate account-deletion contract, security, cryptographic-erasure, restore, and operational gates | Complete / merged through PR #47 at `4680e9c1b7b916169972c84ad8c3879955c7f509` | [Stage 3.18 proposal](stages/STAGE_03_18_PRIVACY_CONTRACT_SECURITY_PROPOSAL.md) |
| 3.19 — Privacy Security and ADR Proposal | Define provider-neutral cryptographic-erasure, deletion-marker, restore, and separation-of-duties controls | Complete / merged through PR #48 at `fdf74c1` | [Stage 3.19 dossier](stages/STAGE_03_19_PRIVACY_SECURITY_ADR_PROPOSAL.md) |
| 3.20 — Privacy Lifecycle Threat-Model Proposal | Define the future privacy-lifecycle threat boundary, residual risks, and review evidence | Complete / merged through PR #49 at `849d934906f878a6d79ba89e940e5ba470e64c09` | [Stage 3.20 threat model](stages/STAGE_03_20_PRIVACY_THREAT_MODEL_PROPOSAL.md) |
| 3.21 — Privacy Data-Inventory Proposal | Map observed privacy-relevant fields and external evidence gaps before any deletion/anonymization design | Complete / merged through PR #50 at `207325e0497cc2608b99366f7f840472d270b6ed`; internal and blind external review evidence recorded | [Stage 3.21 inventory](stages/STAGE_03_21_PRIVACY_DATA_INVENTORY_PROPOSAL.md) |
| 3.22 — Privacy Key-Custody and Destruction-Proof Proposal | Define provider-neutral custody, irreversible destruction proof, and fail-closed evidence requirements | Complete / merged through PR #51 at `5f42d32db1e045c23fb99a5af8f136b7a49e3bc2` | [Stage 3.22 proposal](stages/STAGE_03_22_PRIVACY_KEY_CUSTODY_PROPOSAL.md) |
| 3.23 — Privacy Deletion-Marker Control-Plane Proposal | Define a restricted non-identifying marker lifecycle, snapshot integrity, and fail-closed restore replay | Complete / merged through PR #52 at `f7f23bce33038f259c976db6375079c68209a7aa` | [Stage 3.23 proposal](stages/STAGE_03_23_PRIVACY_DELETION_MARKER_PROPOSAL.md) |
| 3.24 — Privacy Security Review Readiness Dossier | Define the mandatory evidence package, questions, outcomes, and residual decision boundary before formal Security Review | Complete / merged through PR #53 at `544ad8cc7371caf93913ea7716f3feb68be0ea44` | [Stage 3.24 dossier](stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md) |
| 3.25 — Privacy Security Review Evidence-Collection Plan | Define minimal, integrity-protected, independently verified evidence collection before formal Security Review | Active / proposal only | [Stage 3.25 plan](stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md) |
| 3.27 — Import Financial Identity and Cash-Flow Semantics Remediation | Remediate repository-audit P1-02/P1-03/P1-04 across import identity, reconciliation, cash-flow semantics, PostgreSQL, and OpenAPI | Complete / closed; implementation merged through PR #55 at `6e8c806de857f844954f1db513487357dfe90187`; closure governance recorded through PR #58 | [Stage 3.27 report](stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md) |
| 3.28 — Authentication Security Remediation | Remediate repository-audit P1-01/P1-05 across refresh-token replay/session-family containment and bounded Argon2 work | Complete / closed; implementation merged through PR #59 at `dc83f5f3a11da164e6809593861d96ccf47b29ca` after CI #114, renewed independent `APPROVED`, and human approval; closure governance merged through PR #60 at `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` | [Stage 3.28 report](stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md) |
| 3.29 — Input and Contract Hardening | Remediate audit P2-05/P2-06/P2-07/P2-08/P2-15 across client validation, exact-decimal/storage bounds, strict JSON commands, note length, CSV schema ambiguity, and snapshot aggregate arithmetic | Complete / closed; implementation merged through PR #61 at `7331d3f34783baec3997497d1a79b78eaa558bd4`; closure governance merged through PR #62 at `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` | [Stage 3.29 report](stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md) |
| 3.30 — Import Review Integrity | Remediate audit P2-02/P2-03/P2-04 across review-token semantics, parser-owned row bounds, and full-history targeted reconciliation | Complete / closed; implementation merged through PR #63 at `8f68dd18800918e6a9882e995e13dba2723dc929`; closure governance merged through PR #64 at `ae6497050692798795efb85678af64db97cc5f53` | [Stage 3.30 report](stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md) |
| 3.31 — Authentication Operational Hardening | Remediate audit P2-01/P2-14 across logout admission and bounded auth-limiter lifecycle | Complete / closed; implementation merged through PR #65 at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`; closure governance merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` | [Stage 3.31 report](stages/STAGE_03_31_AUTH_OPERATIONAL_HARDENING.md) |
| 3.32 — Exact Idempotency Replay and Browser Retry Recovery | Remediate audit P2-09/P2-13 across exact original-response replay and browser retry continuity/isolation | Implementation merged through PR #67 at `0623d5ef326cd783b7dc0417dbcb02f18c506171` after CI #181 and repeat independent `APPROVED`; closure governance closes the findings when canonical and leaves 5 P2 plus 10 P3 | [Stage 3.32 report](stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md) |

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
- Kept public import endpoints, upload UI, SQL import-session persistence, broker/provider
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
- Started documentation-only planning for a future Next.js CSV import upload/review UI over the existing
  Stage 3.9 Go API boundary.
- Kept Next.js implementation, OpenAPI changes, Go handlers, SQL migrations, import-session persistence,
  raw file persistence, workers, provider integrations, tax logic, mobile, AI, and implementation out of
  scope.

## 2026-07-09 — Stage 3.10 import upload/review UI slice started

- Squash-merged Stage 3.10 planning into `develop` at
  `27480d6ff22e2929e33aeac352aef8a1b01bb448`.
- Started the implementation slice for a presentation-only Next.js CSV import upload/review panel.
- Scoped implementation to typed Go API calls, transient in-memory file handling, review display,
  explicit row approval, append submission, and UI feedback only.
- Kept OpenAPI changes, backend handlers, SQL migrations, import-session persistence, raw file
  persistence, provider integrations, workers, tax, mobile, AI, and Stage 3.11 out of scope.

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

## 2026-07-09 — Stage 3.12 Web authentication UI planning started

- Squash-merged Stage 3.11 closure governance into `develop` at
  `2febb6f49224ec6252368d2195a4e3054ea24278`.
- Started documentation-only planning for a future Next.js presentation-only registration/login,
  session-shell, refresh, and logout UI over the existing Stage 3.11 Go API auth boundary.
- Kept Next.js implementation, Go handler changes, OpenAPI changes, SQL migrations, token-storage
  changes, business logic, provider integrations, workers, tax, mobile, AI, and Stage 3.13 out of
  scope.

## 2026-07-11 — Stage 3.12 Web authentication UI slice closed

- Squash-merged PR #32 into `develop` at `b4840b60346109e3cd54a07d9e1e131fc0cfad23`.
- Closed the Stage 3.12 presentation-only Web authentication UI slice after green GitHub CI and
  strict independent follow-up review approval.
- Added registration/login UI, authenticated shell gating, refresh/logout controls, in-memory
  access-token propagation to existing portfolio/import Go API calls, stale-token load guards,
  local credentialed CORS support for the HttpOnly refresh cookie, and frontend API/session tests.
- Kept Route Handlers, Server Actions, OpenAPI changes, SQL migrations, direct datastore access,
  refresh-token JavaScript storage, durable browser token storage, provider integrations, workers,
  tax, mobile, AI, and Stage 3.13 work out of scope.

## 2026-07-11 — Stage 3.13 instrument catalog planning started

- Squash-merged Stage 3.12 closure governance into `develop` at
  `321eaf4f75df83d85fd356a8d6a454e49bbc4db4`.
- Started documentation-only planning for a future backend-owned MVP instrument catalog boundary for
  MOEX shares and bonds.
- Kept implementation, provider integrations, workers, market-data ingestion, financial
  calculations, tax, mobile, AI, and frontend business authority out of scope.

## 2026-07-12 — Stage 3.13 instrument catalog slice started

- Squash-merged Stage 3.13 planning into `develop` at
  `ca16af9adba249fc8c32c9b246b5f92f7e290b92` after strict follow-up review approval and green CI.
- Started the backend-only implementation slice on `feature/stage-03-13-instrument-catalog`.
- Scoped implementation to approved local asset fixtures, backend ticker resolution through the
  existing `investment.assets` table, unsupported ticker rejection, stock/bond asset-type
  preservation in existing snapshot buckets, tests, and documentation.
- Kept OpenAPI changes, SQL migrations, Go handler changes, frontend work, provider integrations,
  workers, market-data ingestion, stock/bond cards, dividend/coupon scope, tax, mobile, and AI out
  of scope.

## 2026-07-12 — Stage 3.13 implementation hardening updated

- Changed the catalog boundary to enforce literal ticker contract matching, seed approved assets
  without reactivating or rewriting existing rows, resolve import-batch assets in deterministic
  ticker order, and assert approved fixture metadata in tests.
- Required existing active catalog rows to match approved canonical metadata while legacy internal
  UUIDs remain compatible when all canonical metadata matches.
- Required catalog-mutation integration tests to restore shared database state with checked cleanup
  operations.
- Kept the fixes inside the backend/store/test/documentation implementation slice; no OpenAPI,
  migration, handler, frontend, provider, worker, market-data, tax, mobile, or AI scope was added.

## 2026-07-13 — Stage 3.13 instrument catalog slice closed

- Squash-merged PR #35 into `develop` at `b9c05fb14d0ee03e6de4dfc04ff67c16da33040b`.
- Closed the backend-owned instrument catalog boundary after green GitHub CI and strict
  separate-window review approval.
- Preserved the narrow approved fixture set and explicit unsupported ticker rejection while keeping
  provider ingestion, market-data collection, instrument cards, dividend/coupon scope, frontend,
  mobile, tax, AI, OpenAPI changes, SQL migrations, and Go handler changes out of scope.
- Advanced the governance baseline to Stage 3.14 planning.

## 2026-07-13 — Stage 3.14 asset search/card API boundary planning started

- Squash-merged Stage 3.13 closure governance into `develop` at
  `45a298e3ba36dbe711fa27b8d044d80a77cfd74a`.
- Started documentation-only planning for a future Go API implementation of the frozen
  `GET /api/v1/assets/search` and `GET /api/v1/assets/{ticker}` contract over the Stage 3.13
  backend-owned catalog.
- Kept implementation, frontend stock/bond cards, OpenAPI changes, SQL migrations, provider
  integrations, workers, market-data ingestion, financial calculations, tax, mobile, and AI out of
  scope.

## 2026-07-13 — Stage 3.14 asset search/card API boundary slice started

- Squash-merged Stage 3.14 planning PR #37 into `develop` at
  `2c4f7853599a455bb0cc04114b338a1145baf39c`.
- Started the backend-only implementation slice on `feature/stage-03-14-asset-api-boundary`.
- Added the implementation report for a public Go API asset search/detail boundary over the
  Stage 3.13 backend-owned local catalog.
- Scoped runtime behavior to active canonical catalog summaries with `lastPrice: null`; asset-card
  detail remains deferred with `404 NOT_FOUND` until mandatory source/detail fields can be
  populated without fabricated data.
- Kept OpenAPI changes, SQL migrations, frontend stock/bond cards, providers, workers,
  market-data ingestion, financial calculations, tax, mobile, AI, and Stage 3.15 out of scope.

## 2026-07-14 — Stage 3.14 asset search/card API boundary slice closed

- Squash-merged PR #38 into `develop` at `57a9404952cb65693614109dd4a14d41fa5c4295`.
- Closed the Go API asset search/detail boundary after green CI and strict separate-window review
  approval.
- Added public asset search over active canonical approved catalog rows with `lastPrice: null`.
- Kept asset detail honest and contract-safe by returning `404 NOT_FOUND` until approved runtime
  source provenance and mandatory detail fields exist.
- Kept OpenAPI changes, SQL migrations, frontend stock/bond cards, provider integrations,
  market-data ingestion, workers, financial calculations, tax, mobile, AI, and Stage 3.15 out of
  scope.

## 2026-07-26 — Stage 3.15 Web asset discovery UI planning started

- Squash-merged Stage 3.14 closure governance PR #39 into `develop` at
  `f5289eb604b8ba31aa422d0d09950da02e0f48b3` after green CI and strict separate-window review
  approval.
- Started documentation-only planning for a future Next.js presentation-only asset discovery entry
  over the existing Go asset search API.
- Scoped planning to UI state, typed frontend API calls, routing, accessibility, and honest deferred
  asset-card behavior while asset detail remains unavailable.
- Kept implementation, OpenAPI changes, SQL migrations, Route Handlers, Server Actions, direct
  database access, market-data/provider integrations, workers, stock/bond financial calculations,
  tax, mobile, and AI out of scope.

## 2026-07-26 — Stage 3.15 planning review findings fixed

- Addressed strict separate-window review findings by making the planned audit target reproducible:
  after the planning PR merge, the audit stage must record the full post-planning `develop` SHA.
- Required a tracked-file coverage manifest where every path is audited or excluded with a narrow
  generated, vendored, binary, or archival rationale.
- Added explicit SOLID, cost, and ADR-consistency coverage to the audit method and acceptance
  criteria.

## 2026-07-26 — Stage 3.15 Web asset discovery UI slice started

- Squash-merged Stage 3.15 planning PR #40 into `develop` at
  `dfeab109b2825fe0e0317e87a7abf2e706a29ea6` after green CI and strict separate-window review
  approval.
- Started the Next.js presentation-only implementation slice on
  `feature/stage-03-15-asset-discovery-ui`.
- Scoped implementation to typed public asset API calls, UI state, stale-response/pagination guards,
  unavailable price rendering, deferred detail handling, routing, accessibility behavior, tests, and
  documentation.
- Kept OpenAPI changes, SQL migrations, Go handlers, Route Handlers, Server Actions, direct
  datastore access, provider/market-data integrations, workers, financial calculations, tax, mobile,
  and AI out of scope.

## 2026-07-26 — Stage 3.15 implementation review findings fixed

- Addressed strict separate-window review findings for stale detail-response resurrection, missing
  focus transfer during detail loading, incomplete live-region behavior, weak component wiring
  coverage, incorrect successful detail typing, and incomplete verification evidence.
- Added detail-generation invalidation on search reset and detail close, loading-state focus entry,
  separate search/detail polite announcements, assertive error-only alerts, frozen `Asset` detail
  typing, and source-level component wiring checks.
- Addressed follow-up strict review findings by aligning `Asset` detail types with the frozen
  `SourceReference`, `AssetStatus`, and optional bond coupon-rate schema, making successful detail
  heading copy distinct from deferred detail, and preventing async detail outcomes from stealing
  focus after the initial detail entry.
- Tightened the detail focus helper so retrying the same ticker re-enters the loading detail region
  while same-ticker async outcomes still avoid stealing focus.

## 2026-07-27 — Stage 3.15 Web asset discovery UI slice closed

- Squash-merged PR #41 into `develop` at `22bede651a646d0e8b06568bda457d0626891e63`.
- Closed the reviewed Next.js presentation-only asset discovery boundary after strict
  separate-window review approval and green CI.
- Added `/assets` discovery UI, typed public asset API calls, stale search/detail guards, honest
  unavailable-price and deferred-detail rendering, keyboard/focus/live-region behavior, and focused
  frontend tests.
- Updated the umbrella Stage 3 plan and Stage 3.15 planning report so no current governance
  document still points to Stage 3.15 implementation approval after closure.
- Kept OpenAPI changes, SQL migrations, Go handlers, Route Handlers, Server Actions, direct
  datastore access, provider/market-data integrations, workers, financial calculations, tax, mobile,
  and AI out of scope.

## 2026-07-27 — Stage 3.16 repository audit planning started

- Squash-merged Stage 3.15 closure governance PR #42 into `develop` at
  `9eec98c36d7aeffb21dc2d7e7e0eb1681106901d`.
- Started documentation-only planning for the mandatory full repository audit before the next
  implementation stage.
- Scoped the planned audit to architecture, DDD, API, security, privacy, performance, dependencies,
  tests, documentation, cost, and ADR consistency.
- Kept implementation work, financial algorithms, OpenAPI changes, SQL migrations, dependency
  changes, market data, providers, workers, tax, mobile, and AI out of scope.

## 2026-07-27 — Stage 3.16 planning review findings fixed

- Addressed strict separate-window review findings by making the planned audit target reproducible:
  after the planning PR merge, the audit stage must record the full post-planning `develop` SHA.
- Required a tracked-file coverage manifest where every path is audited or excluded with a narrow
  generated, vendored, binary, or archival rationale.
- Added explicit SOLID, cost, and ADR-consistency coverage to the audit method and acceptance
  criteria.

## 2026-08-04 — Stage 3.16 repository audit fixes started

- Squash-merged Stage 3.16 planning PR #43 into `develop` at
  `74eebe9ec8231764f21ce384c4690d073d0273da`.
- Completed strict separate-window full repository audit on that immutable SHA; verdict:
  `REQUEST CHANGES`.
- Started `stage-03-16-audit-fixes` and fixed the first blocking audit findings:
  - import append approvals now require the reviewed source file hash and row hashes;
  - Web import review state is invalidated on file/source-label changes and stale review responses;
  - portfolio summary return fields are unavailable instead of exposing placeholder calculations;
  - browser write intents preserve idempotency keys across ambiguous retry attempts;
  - CI now runs frontend tests and PostgreSQL-backed Go integration tests.
- Kept financial algorithms, market data, providers, workers, tax, mobile, and AI out of scope.

## 2026-08-08 — Stage 3.16 audit-fix hardening completed pending review

- Made import review-token configuration fail closed outside explicitly named local/development
  environments; the service now requires a 32-byte-or-longer secret and hashes it before signing.
- Added source-aware transaction provenance migration and integration coverage: independent matching
  manual ledger entries remain allowed, while equivalent imports remain rejected under the shared lock.
- Added an exact SHA-256 source-file-hash validation boundary for imported batches.
- Replaced client-controlled list offsets with signed opaque keyset cursors bound to subject,
  endpoint, portfolio/filter scope, and deterministic page anchors; the Web layer forwards only the
  returned opaque token.
- Made missing database configuration fail closed outside explicit local/development mode, matched
  runtime import token/hash validation to OpenAPI bounds, and added CI rollback/reapply rehearsal for
  every migration in disposable PostgreSQL.
- Preserved HTTP import retry replay through the atomic store, switched SQL transaction cursors to
  internal entry anchors, applied migration `000003` in smoke, and scope-bound in-flight Web import
  operations to the active portfolio/session.
- Added a durable 200-path Stage 3.16 coverage manifest and synchronized the audit report, document
  index, and version matrix. The original audit verdict remains `REQUEST CHANGES` as historical
  evidence pending the later fix disposition.
- Signed asset-search continuations with query/type-bound HMAC keysets and synchronously invalidate
  stale import operations before passive effects run after session or portfolio changes.
- Hardened migration-rehearsal traversal and schema evidence, and added a mounted React lifecycle test
  for stale review and append promises after a portfolio or session change.

## 2026-08-09 — Stage 3.16 audit fixes closed

- PR #44 was squash-merged into `develop` at
  `9e6b8a753bf73ef020ce40461df25a5878344d92` after read-only review approval and green GitHub
  Actions verification run `31300786551`.
- CI initially exposed PostgreSQL `CHECK` NULL semantics and a stale rollback-test row count; the
  follow-up fix corrected both before the final green rerun.
- The immutable audit report retains its original `REQUEST CHANGES` verdict; its in-scope blocking
  findings are resolved in the merged fix set.
- No financial algorithms, market data, providers, workers, tax, mobile, AI, or subsequent
  implementation stage is authorized. The next work requires separately reviewed planning.
- Follow-up closure review synchronized the umbrella Stage 3 plan status, recorded the location of
  human merge authorization, and distinguished it from an independently recorded external review.

## 2026-08-09 — Stage 3.17 privacy lifecycle planning started

- Began documentation-only planning for the remaining Stage 3.16 account-deletion, anonymization,
  backup-destruction, and retention-execution blocker.
- Anchored the future lifecycle in the existing identity state, anonymous subject state, and sole
  identity-to-subject link while requiring a reviewed contract, migration, security, and restore path.
- Kept runtime code, OpenAPI, SQL migrations, backup/KMS configuration, market data, financial
  calculations, tax, mobile, AI, and all implementation work out of scope.

## 2026-08-09 — Stage 3.18 privacy contract and security proposal started

- Squash-merged Stage 3.17 privacy-lifecycle planning PR #46 into `develop` at `1e8c240` after
  an approved read-only review and green GitHub Actions verification.
- Started a documentation-only candidate contract and security proposal for account deletion,
  30-day grace lifecycle, cryptographic erasure, deletion-marker replay, backup retention, and
  restore fail-closed behavior.
- Closed the historical Stage 3.17 plan record at PR #46 / `1e8c240` and removed its stale active
  duplicate from the version matrix.
- Kept the OpenAPI, runtime, PostgreSQL schema, key-management/provider configuration, operations,
  market data, financial calculations, tax, mobile, AI, and all implementation work out of scope.

## 2026-08-09 — Stage 3.19 privacy security and ADR proposal started

- Squash-merged Stage 3.18 privacy contract/security proposal PR #47 into `develop` at
  `4680e9c1b7b916169972c84ad8c3879955c7f509` after green GitHub Actions and the explicitly
  recorded Principal Architect review-process exception.
- Started proposed ADR-008 and a documentation-only security dossier for the per-subject erasure
  boundary, non-identifying deletion markers, fail-closed restore, separation of duties, backup
  evidence, and cross-system partial failure.
- Kept ADR acceptance, Security Review approval, OpenAPI, runtime, PostgreSQL schema, key-management
  or provider selection, backup operations, and all implementation work out of scope.

## 2026-08-09 — Stage 3.20 privacy threat-model proposal started

- Recorded Stage 3.19 as squash-merged through PR #48 at
  `fdf74c16446e7623f76882aa7add64554141abc6`.
- Started the documentation-only privacy threat-model proposal for future browser/session authority,
  application and privileged-data access, key custody, marker controls, backup/restore, partial
  failure, evidence redaction, and indirect-reidentification risks.
- Kept ADR acceptance, Security Review approval, runtime, OpenAPI, PostgreSQL schema, providers,
  key-management configuration, backup operations, and all implementation work out of scope.

## 2026-08-09 — Stage 3.20 review evidence recorded

- Published the previously withheld internal-review evidence after the dedicated blind external task
  independently reviewed PR #49 and returned `APPROVED`.
- Recorded both `APPROVED` verdicts as governance evidence only; ADR-008, Security Review, runtime,
  OpenAPI, PostgreSQL schema, providers, backup operations, and implementation remain unchanged.

## 2026-08-09 — Stage 3.21 privacy data-inventory proposal started

- Recorded Stage 3.20 as squash-merged through PR #49 at
  `849d934906f878a6d79ba89e940e5ba470e64c09` after internal and blind external review evidence.
- Started a repository-derived, documentation-only inventory of observed privacy-relevant database,
  code, browser, and import surfaces, with explicit external backup/log/provider/CI evidence gaps.
- Kept Security Review approval, ADR-008 acceptance, production data discovery, runtime, OpenAPI,
  PostgreSQL schema, providers, key-management, backup operations, and implementation out of scope.

## 2026-08-09 — Stage 3.21 review evidence recorded

- Published the previously withheld internal-review evidence after the dedicated blind external task
  independently reviewed Draft PR #50 and returned `APPROVED`.
- Recorded both `APPROVED` verdicts as governance evidence only; ADR-008, Security Review, runtime,
  OpenAPI, PostgreSQL schema, providers, backup operations, and implementation remain unchanged.

## 2026-08-18 — Stage 3.22 privacy key-custody and destruction-proof proposal started

- Recorded Stage 3.21 as squash-merged through PR #50 at
  `207325e0497cc2608b99366f7f840472d270b6ed` after internal and blind external `APPROVED` verdicts.
- Added the next documentation-only evidence proposal: provider-neutral key-custody roles,
  per-subject erasure-material lifecycle, integrity-verifiable non-identifying destruction proof,
  provider evaluation, restore gate, and required adversarial evidence.
- No provider, Security Review, ADR-008 acceptance, implementation, API, schema, migration,
  credential, backup, CI, infrastructure, or product-scope change is authorized.

## 2026-08-18 — Stage 3.23 privacy deletion-marker control-plane proposal started

- Recorded Stage 3.22 as squash-merged through PR #51 at
  `5f42d32db1e045c23fb99a5af8f136b7a49e3bc2` after recorded internal and external review
  evidence and green GitHub Actions.
- Started documentation-only design of a restricted deletion-marker control plane: monotonic
  lifecycle, serialized-intent binding, marker integrity and availability, isolated restore replay,
  fail-closed release, and retention boundaries.
- Kept Security Review approval, ADR-008 acceptance, runtime, OpenAPI, PostgreSQL schema,
  migrations, providers, custody configuration, backup operations, and implementation out of scope.

## 2026-08-18 — Stage 3.24 privacy Security Review readiness dossier started

- Recorded Stage 3.23 as squash-merged through PR #52 at
  `f7f23bce33038f259c976db6375079c68209a7aa` after internal corrective and non-blind external
  review evidence and green GitHub Actions.
- Started documentation-only readiness work that distinguishes repository proposals from required
  operational evidence and defines formal Security Review questions, evidence rules, outcomes, and
  record requirements.
- Kept Security Review execution, ADR-008 acceptance, runtime, OpenAPI, PostgreSQL schema,
  migrations, providers, custody configuration, backup operations, and implementation out of scope.

## 2026-08-18 — Stage 3.25 privacy Security Review evidence-collection plan started

- Recorded Stage 3.24 as squash-merged through PR #53 at
  `544ad8cc7371caf93913ea7716f3feb68be0ea44` after internal corrective and non-blind external
  review evidence and green GitHub Actions.
- Started the documentation-only plan for accountable, minimized, integrity-protected, independently
  verified collection of each mandatory Security Review input, including restricted handling and
  fail-closed treatment of missing, stale, conflicting, or unverifiable evidence.
- Kept evidence collection, formal Security Review, ADR-008 acceptance, provider selection, runtime,
  OpenAPI, PostgreSQL schema, migrations, credentials, backup operations, and implementation out of
  scope.

## 2026-08-22 — Stage 3.27 closure governance

- PR #55 passed exact-head CI #90 on `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c`.
- Renewed independent external review returned `APPROVED` on that exact head.
- The earlier `REQUEST CHANGES` remains historical evidence of the corrected
  fallback-to-strong identity defect.
- Explicit human authorization approved squash merge of PR #55.
- PR #55 was squash-merged into `develop` at `6e8c806de857f844954f1db513487357dfe90187`.
- PR #58 records the closure governance for Stage 3.27. Once this closure
  record is canonical on `develop`, P1-02/P1-03/P1-04 are closed.
- P1-01 and P1-05 remain separate future remediation.
- Stage 3.25 remains separate proposal-only privacy evidence planning.

## 2026-08-23 — Stage 3.28 authentication security remediation closure governance

- P1-01 and P1-05 implementation was squash-merged through PR #59 at `dc83f5f3a11da164e6809593861d96ccf47b29ca`.
- Exact implementation head `92edab5d3e93dafe2fcc6247644e38e878a4202f` passed GitHub Actions CI #114.
- Initial independent review returned `REQUEST CHANGES` for one governance-status mismatch only;
  renewed independent review returned `APPROVED` after the one-line correction.
- Explicit human squash-merge authorization was given before PR #59 merged.
- Closure governance is delivered through PR #60 and becomes authoritative when this record is
  canonical on `develop`.
- Stage 3.25 privacy evidence planning and the P2/P3 audit backlog remain separate.

## 2026-08-23 — Stage 3.29 input and contract hardening implementation merged

- Squash-merged implementation PR #61 into `develop` at `7331d3f34783baec3997497d1a79b78eaa558bd4`.
- Closed the implementation work for P2-05, P2-06, P2-07, P2-08, and P2-15 after exact-head CI #124.
- Recorded the first independent `REQUEST CHANGES` for incomplete P2-07 aggregate snapshot
  arithmetic admission, the same-transaction PostgreSQL remediation on
  `f9e70e70956c76edbc2ab02c52d45124b2dea525`, and renewed independent `APPROVED`.
- Preserved atomic rollback for rejected snapshot-bound writes and kept the SQL snapshot methodology
  as the single calculation source.
- Recorded detailed root cause, failure mode, rationale, rejected alternatives, regression evidence,
  and review history in `STAGE_03_29_INPUT_CONTRACT_HARDENING.md`.
- Closure governance is tracked through PR #62; when canonical, the remaining original audit backlog
  is 12 P2 and 10 P3 findings. Stage 3.25 privacy evidence planning remains separate.

## 2026-08-23 — Stage 3.30 import review integrity implementation merged

- Squash-merged implementation PR #63 into `develop` at `8f68dd18800918e6a9882e995e13dba2723dc929`.
- Remediated P2-02 with versioned, expiring review-token semantic binding and signed review-time
  APPENDABLE status while retaining locked-store authority over post-review ledger races.
- Remediated P2-03 with a parser-owned 100-data-row computational bound.
- Remediated P2-04 with complete targeted historical reconciliation rather than latest-100 public
  transaction pagination.
- Exact-head CI #128 completed `SUCCESS` on `2f788e0811d78c9def0502676a74bee2f9922bf5` across all six workflow jobs; the Go job
  ran with PostgreSQL and executed the Stage 3.30 integration regression.
- Independent final implementation review returned `APPROVED`; explicit human squash-merge
  authorization was received before PR #63 merged.
- Local verification also exposed and corrected two harness/interface issues before commit: the stale
  ledger race regression was adjusted to mutate ledger state after review, and development
  `unavailableStore` was extended fail-closed for the new history-read interface.
- Closure governance is tracked through PR #64. When canonical, the remaining original audit backlog
  is 9 P2 and 10 P3 findings. Stage 3.25 privacy evidence planning remains separate.

## 2026-08-23 — Stage 3.31 authentication operational hardening implementation merged

- Squash-merged implementation PR #65 into `develop` at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`.
- Remediated P2-01 by placing logout behind auth admission before rejected-auth persistence.
- Remediated P2-14 with finite per-key, global-attempt, and active-key bounds plus expired-bucket reclamation.
- Exact-head CI #133 completed `SUCCESS` on `82557c55c0772a66707088b858ec9eafc2073119` across all six workflow jobs.
- Independent final implementation review returned `APPROVED`; explicit human squash-merge authorization
  was received before PR #65 merged.
- No Redis, distributed limiter, migration, or product-scope expansion was introduced.
- Closure governance was squash-merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746`; P2-01/P2-14 are CLOSED and 7 P2 plus 10 P3 findings remained.

## 2026-08-23 — Stage 3.32 exact idempotency replay and browser retry recovery implementation merged

- Squash-merged implementation PR #67 into `develop` at `0623d5ef326cd783b7dc0417dbcb02f18c506171`.
- P2-09 now persists the exact original versioned HTTP response artifact in the same PostgreSQL transaction as the financial mutation and returns stored response bytes/status/request/trace identity on duplicate replay rather than rereading mutable resource state.
- P2-13 now preserves unresolved browser retry identity across reload/remount using a short-lived `sessionStorage` journal scoped by stable authenticated principal + operation + optional portfolio before SHA-256 slot derivation; raw principal, portfolio scope, financial payload, CSV, review token, and auth tokens are not stored.
- The first independent remediation review returned `REQUEST CHANGES`: P2-09 was CLOSED, but P2-13 remained open because User A/User B could share the same browser-tab retry slot.
- The remediation added stable-principal scoping across Create Portfolio, Add Transaction, and Import Append plus an explicit A→B→A isolation regression.
- Repeat independent review on exact head `02aa2417a3caca79e2afc4e7b598b92055de96b7` returned `APPROVED`, marking P2-09 and P2-13 CLOSED with no new blocking P1/P2 regression.
- Exact-head CI #181 passed all six jobs, including PostgreSQL-backed Go tests, migration validation/apply/rollback/reapply, frontend typecheck/tests/build, Python tests, OpenAPI validation, and Docker Compose validation.
- Explicit human squash-merge authorization was received before PR #67 merged.
- Closure governance is tracked on `docs/stage-03-32-closure`; when canonical on `develop`, the remaining original audit backlog is 5 P2 and 10 P3 findings: P2-10/P2-11/P2-12/P2-16/P2-17 plus all P3 findings. Stage 3.25 remains separate.
