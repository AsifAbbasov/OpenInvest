# OpenInvest Document Index

| Field | Value |
| --- | --- |
| Document ID | REG-DOC-001 |
| Version | 1.1.65 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal attachment-only inventory |
| Dependencies | `SOURCE_OF_TRUTH.md` |
| Last Review Date | 2026-08-24 |
| Next Review Date | Before Stage 3.25 evidence-collection plan review, evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

## Canonical control documents

| ID | Title | Version | Status | Location |
| --- | --- | --- | --- | --- |
| SOT-001 | Source of Truth | 1.4.65 | Approved | `SOURCE_OF_TRUTH.md` |
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
| `audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` | Single cross-stage index for the original 32 repository-audit findings: root-cause/impact summary, remediation rationale, evidence, and status |
| `audit/REMEDIATION_DOCUMENTATION_STANDARD.md` | Mandatory 18-part causal/risk/remediation/evidence standard for Stage 3.34 and all remaining P3 remediation dossiers |
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

## Proposed architecture decisions

| Document | Status | Location |
| --- | --- | --- |
| ADR-008 privacy-lifecycle erasure and restore controls | Proposed / non-normative pending Security Review and human acceptance | `ADR/ADR-008-privacy-lifecycle-erasure-and-restore.md` |

## Stage 3 planning documents

| Document | Status | Location |
| --- | --- | --- |
| Stage 3 plan | Closed through Stage 3.24 dossier; Stage 3.25 evidence-collection plan is active without implementation authorization | `stages/STAGE_03_FIRST_VERTICAL_SLICE.md` |
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
| Stage 3.11 authentication and privacy-boundary planning | Complete / merged into `develop` | `stages/STAGE_03_11_AUTH_PRIVACY_PLANNING.md` |
| Stage 3.11 authentication and privacy-boundary slice | Complete / merged into `develop` | `stages/STAGE_03_11_AUTH_PRIVACY_SLICE.md` |
| Stage 3.12 Web authentication UI planning | Complete / merged into `develop` | `stages/STAGE_03_12_AUTH_UI_PLANNING.md` |
| Stage 3.12 Web authentication UI slice | Complete / merged into `develop` | `stages/STAGE_03_12_AUTH_UI_SLICE.md` |
| Stage 3.13 instrument catalog planning | Complete / merged into `develop` | `stages/STAGE_03_13_INSTRUMENT_CATALOG_PLANNING.md` |
| Stage 3.13 instrument catalog slice | Complete / merged into `develop` | `stages/STAGE_03_13_INSTRUMENT_CATALOG_SLICE.md` |
| Stage 3.14 asset search/card API boundary planning | Complete / merged into `develop` | `stages/STAGE_03_14_ASSET_API_BOUNDARY_PLANNING.md` |
| Stage 3.14 asset search/card API boundary slice | Complete / merged into `develop` | `stages/STAGE_03_14_ASSET_API_BOUNDARY_SLICE.md` |
| Stage 3.15 Web asset discovery UI planning | Complete / merged into `develop` | `stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_PLANNING.md` |
| Stage 3.15 Web asset discovery UI slice | Complete / merged into `develop` | `stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_SLICE.md` |
| Stage 3.16 repository audit planning | Complete / merged into `develop` | `stages/STAGE_03_16_REPOSITORY_AUDIT_PLANNING.md` |
| Stage 3.16 repository audit report | Complete / returned `REQUEST CHANGES` | `stages/STAGE_03_16_REPOSITORY_AUDIT_REPORT.md` |
| Stage 3.16 repository audit coverage manifest | Complete / immutable 200-path coverage record | `stages/STAGE_03_16_REPOSITORY_AUDIT_MANIFEST.md` |
| Stage 3.16 repository audit fixes | Complete / merged into `develop` at `9e6b8a753bf73ef020ce40461df25a5878344d92` | `stages/STAGE_03_16_REPOSITORY_AUDIT_FIXES.md` |
| Stage 3.17 privacy lifecycle planning | Complete / merged through PR #46 | `stages/STAGE_03_17_PRIVACY_LIFECYCLE_PLANNING.md` |
| Stage 3.18 privacy contract and security proposal | Complete / merged through PR #47 | `stages/STAGE_03_18_PRIVACY_CONTRACT_SECURITY_PROPOSAL.md` |
| Stage 3.19 privacy security and ADR proposal | Complete / merged through PR #48 | `stages/STAGE_03_19_PRIVACY_SECURITY_ADR_PROPOSAL.md` |
| Stage 3.20 privacy lifecycle threat-model proposal | Complete / merged through PR #49; internal and blind external review evidence recorded | `stages/STAGE_03_20_PRIVACY_THREAT_MODEL_PROPOSAL.md` |
| Stage 3.21 privacy data-inventory proposal | Complete / merged through PR #50; internal and blind external review evidence recorded | `stages/STAGE_03_21_PRIVACY_DATA_INVENTORY_PROPOSAL.md` |
| Stage 3.22 privacy key-custody and destruction-proof proposal | Complete / merged through PR #51; internal and external review evidence recorded | `stages/STAGE_03_22_PRIVACY_KEY_CUSTODY_PROPOSAL.md` |
| Stage 3.23 privacy deletion-marker control-plane proposal | Complete / merged through PR #52; internal corrective and non-blind external review evidence recorded | `stages/STAGE_03_23_PRIVACY_DELETION_MARKER_PROPOSAL.md` |
| Stage 3.24 privacy Security Review readiness dossier | Complete / merged through PR #53; internal corrective and non-blind external review evidence recorded | `stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md` |
| Stage 3.25 privacy Security Review evidence-collection plan | Active / proposal only | `stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md` |
| Stage 3.27 import financial identity and cash-flow semantics remediation | Complete / closed for P1-02/P1-03/P1-04; implementation merged through PR #55 at `6e8c806de857f844954f1db513487357dfe90187` after exact-head CI #90, renewed independent `APPROVED`, and explicit human merge approval; closure governance recorded through PR #58 | `stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md` |
| Stage 3.28 authentication security remediation | Complete / closed for P1-01/P1-05; implementation merged through PR #59 at `dc83f5f3a11da164e6809593861d96ccf47b29ca` after CI #114, renewed independent `APPROVED`, and explicit human approval; closure governance merged through PR #60 at `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` | `stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md` |
| Stage 3.29 input and contract hardening | Complete / closed for P2-05/P2-06/P2-07/P2-08/P2-15; implementation merged through PR #61 at `7331d3f34783baec3997497d1a79b78eaa558bd4` after CI #124 and renewed independent `APPROVED`; closure governance merged through PR #62 at `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` | `stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` |
| Stage 3.30 import review integrity | Complete / closed for P2-02/P2-03/P2-04; implementation merged through PR #63 at `8f68dd18800918e6a9882e995e13dba2723dc929`; closure governance merged through PR #64 at `ae6497050692798795efb85678af64db97cc5f53` | `stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md` |
| Stage 3.31 authentication operational hardening | Complete / closed for P2-01/P2-14; implementation merged through PR #65 at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`; closure governance merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` | `stages/STAGE_03_31_AUTH_OPERATIONAL_HARDENING.md` |
| Stage 3.32 exact idempotency replay and browser retry recovery | Complete / closed for P2-09/P2-13; implementation merged through PR #67 at `0623d5ef326cd783b7dc0417dbcb02f18c506171`; closure governance merged through PR #68 at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` | `stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md` |
| Stage 3.33 snapshot rebuild accuracy and PostgreSQL runtime immutability | Complete / closed for P2-10/P2-11/P2-12; implementation PR #69 merged at `87a7c38e16062a5f3fcef3727f60c0c6741eb805` after CI #199 and final independent `APPROVED`; canonical closure governance PR #70 merged at `71a1faeb97d33d05f2936111b53f1285edddabe9` | `stages/STAGE_03_33_SNAPSHOT_REBUILD_POSTGRES_IMMUTABILITY.md` |
| Stage 3.33 closure governance | Complete / canonical; preserves both P2-12 `REQUEST CHANGES` cycles, final implementation approval, closure CI/review, and separate human squash-merge authorization | `stages/STAGE_03_33_CLOSURE.md` |

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
