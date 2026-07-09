# OpenInvest Document Index

| Field | Value |
| --- | --- |
| Document ID | REG-DOC-001 |
| Version | 1.1.23 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal attachment-only inventory |
| Dependencies | `SOURCE_OF_TRUTH.md` |
| Last Review Date | 2026-07-01 |
| Next Review Date | 2027-01-01 |

## Canonical control documents

| ID | Title | Version | Status | Location |
| --- | --- | --- | --- | --- |
| SOT-001 | Source of Truth | 1.4.23 | Approved | `SOURCE_OF_TRUTH.md` |
| 42 | Architecture Amendments | 1.1.0 | Approved | `specifications/current/DOCUMENT_42_ARCHITECTURE_AMENDMENTS_v1.1.md` |
| 43 | Architecture Decision Closure | 1.2.0 | Final | `specifications/current/DOCUMENT_43_ARCHITECTURE_CLOSURE_v1.2.md` |

## Legacy specification registry

Documents 1–41 remain normative only where they do not conflict with Documents 42–43 or accepted ADRs. Historical number 14 has two source documents and is disambiguated as 14A/14B.

| ID | Title | Version | Repository state |
| --- | --- | --- | --- |
| 00 | Project Manifest | 1.0 | Consolidated legacy source; cancelled draft excluded |
| 01 | Product Vision & PRD | 1.0 | Consolidated legacy source |
| 02 | System Architecture Blueprint | 1.0 | Archived source |
| 03 | Domain Model & Business Logic | 1.0 | Archived source |
| 04 | System Design & Engineering Architecture | 1.0 | Archived source |
| 05 | Database Architecture & PostgreSQL Design | 1.0 | Archived source |
| 06 | API Contract & Backend Architecture | 1.0 | Archived source |
| 07 | Frontend Architecture & Product UX | 1.0 | Archived source |
| 08 | Security, Privacy & Trust by Design | 1.0 | Consolidated legacy source |
| 09–13 | Product, engineering, security, backend, domain | 1.0 | Archived sources |
| 14A | Frontend Architecture | 1.0 | Archived source; duplicate ID disambiguated |
| 14B | Mathematical Engine | 1.0 | Archived source; duplicate ID disambiguated |
| 15–27 | Backend through Final Product Blueprint | 1.0 | Archived sources |
| 28–39 | Architecture constitutions | 2.0 | Archived sources |
| 40 | Codex Execution Manifest | 3.0 | Archived source |
| 41 | Anti-Patterns / Red Book | 1.0 | Archived source |

Individual archived files live under `specifications/legacy/`. They are preserved for traceability and must not be silently edited.

## Governance documents

| Document | Purpose |
| --- | --- |
| `VERSION_MATRIX.md` | Version, ownership, review, and precedence matrix |
| `CHANGELOG.md` | Architecture/documentation change history |
| `OPEN_QUESTIONS.md` | Controlled unresolved-decision register |
| `registries/DATA_SOURCE_REGISTRY.md` | Approved external-source register |
| `BACKLOG_V2.md` | Ideas excluded from MVP |
| `ROADMAP.md` | Ordered implementation stages |
| `IMPLEMENTATION_LOG.md` | Completed-stage index and completion protocol |
| `REVIEW_WORKFLOW.md` | Mandatory branch, PR, CI, specialist review, approval, and merge process |
| `product/MVP_PRODUCT_RISK_REFINEMENT.md` | Proposed product-risk response, ICP sharpening, import/tax/purchasing-power guardrails |

## Stage 2 contract documents

| Document | Status | Location |
| --- | --- | --- |
| ADR-006 | Accepted | `ADR/ADR-006-contract-and-canonical-model-freeze.md` |
| API contract | Closed / canonical Stage 2 baseline | `api/API_CONTRACT_STAGE_02.md` |
| Canonical model | Closed / canonical Stage 2 baseline | `domain/CANONICAL_MODEL_STAGE_02.md` |
| Logical ER model | Closed / canonical Stage 2 baseline | `database/ER_MODEL_STAGE_02.md` |
| Migration strategy | Closed / canonical Stage 2 baseline | `database/MIGRATION_STRATEGY_STAGE_02.md` |
| Stage report | Closed / merged into `develop` | `stages/STAGE_02_CONTRACT_AND_CANONICAL_MODEL.md` |

## Stage 3 planning documents

| Document | Status | Location |
| --- | --- | --- |
| Stage 3 plan | Active / staged implementation | `stages/STAGE_03_FIRST_VERTICAL_SLICE.md` |
| Stage 3.1 database foundation | Complete / merged into `develop` | `stages/STAGE_03_01_DATABASE_FOUNDATION.md` |
| Stage 3.2 Go API vertical slice | Complete / merged into `develop` | `stages/STAGE_03_02_GO_API_VERTICAL_SLICE.md` |
| Stage 3.3 Next.js presentation slice | Complete / merged into `develop` | `stages/STAGE_03_03_NEXTJS_PRESENTATION_SLICE.md` |
| Stage 3.4 end-to-end verification | Complete / merged into `develop` | `stages/STAGE_03_04_END_TO_END_VERIFICATION.md` |
| Stage 3.5 broker file import and reconciliation design | Complete / merged into `develop` | `stages/STAGE_03_05_BROKER_FILE_IMPORT_RECONCILIATION_DESIGN.md` |
| Stage 3.6 broker file import reconciliation slice | Complete / merged into `develop` | `stages/STAGE_03_06_IMPORT_RECONCILIATION_SLICE.md` |
| Stage 3.7 import append planning | Complete / merged into `develop` | `stages/STAGE_03_07_IMPORT_APPEND_PLANNING.md` |
| Stage 3.7 import append slice | Complete / merged into `develop` | `stages/STAGE_03_07_IMPORT_APPEND_SLICE.md` |
| Stage 3.8 import review append flow planning | Complete / merged into `develop` | `stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_PLANNING.md` |
| Stage 3.8 import review append flow slice | Complete / merged into `develop` | `stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_SLICE.md` |
| Stage 3.9 import API boundary planning | Complete / merged into `develop` | `stages/STAGE_03_09_IMPORT_API_BOUNDARY_PLANNING.md` |
| Stage 3.9 import API boundary slice | Complete / merged into `develop` | `stages/STAGE_03_09_IMPORT_API_BOUNDARY_SLICE.md` |
| Stage 3.10 import upload/review UI planning | Complete / merged into `develop` | `stages/STAGE_03_10_IMPORT_UPLOAD_UI_PLANNING.md` |
| Stage 3.10 import upload/review UI slice | Complete / merged into `develop` | `stages/STAGE_03_10_IMPORT_UPLOAD_REVIEW_UI_SLICE.md` |
| Stage 3.11 authentication and privacy-boundary planning | Draft / planning active | `stages/STAGE_03_11_AUTH_PRIVACY_PLANNING.md` |

## Product refinement documents

| Document | Status | Location |
| --- | --- | --- |
| MVP product risk refinement | Approved / merged into `develop` | `product/MVP_PRODUCT_RISK_REFINEMENT.md` |

## Web presentation amendment documents

| Document | Status | Location |
| --- | --- | --- |
| ADR-007 | Accepted | `ADR/ADR-007-use-nextjs-for-web-frontend.md` |
| Architecture freeze note | Superseded Web frontend target only | `ARCHITECTURE_FREEZE_v1.md` |
| Amendment report | Closed / merged into `develop` | `stages/WEB_FRONTEND_ARCHITECTURE_AMENDMENT.md` |
