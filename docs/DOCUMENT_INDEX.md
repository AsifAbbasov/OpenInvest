# OpenInvest Document Index

| Field | Value |
| --- | --- |
| Document ID | REG-DOC-001 |
| Version | 1.1.77 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal attachment-only inventory |
| Dependencies | `SOURCE_OF_TRUTH.md` |
| Last Review Date | 2026-09-10 |
| Next Review Date | Before any Stage 3.78+ product/runtime or architecture-changing scope, Stage 3.25 evidence-collection plan review, evidence collection, formal Security Review, ADR-008 acceptance, Feature 3D runtime activation or broader T-Invest source/use scope, privacy-lifecycle migration proposal, or the next separately reviewed audit-remediation scope |

## Canonical control documents

| ID | Title | Version | Status | Location |
| --- | --- | --- | --- | --- |
| SOT-001 | Source of Truth | 1.5.1 | Approved | `SOURCE_OF_TRUTH.md` |
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
| `audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` | Canonical cross-finding status for all 32 original Stage 3.16 audit findings |
| `governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md` | Historical reconciliation record through Stage 3.72; later active lifecycle state is carried by the canonical registries and stage dossiers |
| `governance/HISTORICAL_FEATURE_FORENSIC_DOCUMENTATION_RECONCILIATION.md` | Cross-stage forensic reconciliation for Stages 3.57–3.77; preserves material review/remediation history, explicit evidence gaps and the Feature 3D reference standard without rewriting historical dossiers |
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

## Proposed architecture decisions

| Document | Status | Location |
| --- | --- | --- |
| ADR-008 privacy-lifecycle erasure and restore controls | Proposed / non-normative pending Security Review and human acceptance | `ADR/ADR-008-privacy-lifecycle-erasure-and-restore.md` |

## Accepted architecture decisions

| Document | Status | Location |
| --- | --- | --- |
| ADR-009 deterministic portfolio ledger ordering and WAC position semantics | Accepted/canonical through Issue #136 and PR #137 squash merge `b772e52221fbb694b3116bd1b579db99d4e56302`; runtime scope remained separately gated and was later implemented by Stage 3.71 | `ADR/ADR-009-deterministic-portfolio-ledger-ordering-and-wac.md` |

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
| Stage 3.32 exact idempotency replay and browser retry recovery | Complete / closure canonical through PR #68 at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be`; P2-09/P2-13 CLOSED | `stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md` |
| Stage 3.33–3.56 repository-audit remediation continuation | Complete; all remaining P2/P3 findings closed and original audit 32/32 = 100% | `audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` |
| Stage 3.57–3.60 market-data lifecycle | Provider-neutral boundary and delayed MOEX ISS adapter implemented; production/public activation NO-GO | `governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md` |
| Stage 3.61–3.67 Corporate Actions lifecycle | Features 3A/3B/3C and request-cancellation closure complete; provider-neutral boundary remains canonical | `governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md` |
| Feature 3D T-Invest Corporate Actions implementation | Complete / merged through PR #164; exact constrained `GetDividends`/`GetBondCoupons` adapter only; runtime activation remains NO | `stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_IMPLEMENTATION.md` |
| Feature 3D review evidence | Internal/External review chronology, CI #499, evidence publication and exact-head evidence CI #501 | `stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE.md` |
| Feature 3D evidence errata | Append-only correction for `SourceEventID` ownership and observation-time `AsOf` semantics | `stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE_ERRATA.md` |
| Feature 3D post-merge closure | Canonical lifecycle distinction: source/use CONDITIONAL-GO, implementation complete, runtime activation NO, live token/production traffic not authorized/claimed | `stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_CLOSURE.md` |
| Stage 3.68–3.69 Dividend Calculator lifecycle | Complete / canonical | `governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md` |
| Stage 3.70 Portfolio Position & Cost Basis Engine planning | Complete/canonical through PR #137 squash merge `b772e52221fbb694b3116bd1b579db99d4e56302`; ADR-009 accepted; Stage 3.70 itself granted no runtime authorization | `stages/STAGE_03_70_PORTFOLIO_POSITION_COST_BASIS_PLANNING.md` |
| Stage 3.71 Portfolio Position & Cost Basis Engine implementation | Canonical through PR #138 squash merge `827f49f909ace5a3f7bcb2a3f51ce7c638c458ad`; final evidence head `6f66f7f3da982b2f83637452bfd1f556b92b3956`; CI #382 / run `34028563925` 10/10 SUCCESS | `stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_IMPLEMENTATION.md` |
| Stage 3.71 rollout amendment | Required activation amendment preserving continuous portfolio-write quiescence through Stage 3.71 readiness; no change to ADR-009 WAC math or ledger ordering | `stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_ROLLOUT_AMENDMENT.md` |
| Stage 3.71 review evidence | Historical implementation/review/evidence chronology, including permanently preserved pre-External evidence-withholding deviation | `stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_REVIEW_EVIDENCE.md` |
| STAGE-03-71-GOV-01 historical governance deviation disposition | Effective through PR #139 squash merge `1da2d3b3b33a9424b06f2f157b2997f20e335954`; historical noncompliance preserved; residual governance risk accepted | `stages/STAGE_03_71_GOV_01_HISTORICAL_GOVERNANCE_DEVIATION_DISPOSITION.md` |
| STAGE-03-72-GOV-01 historical governance deviation disposition | Effective through PR #143 squash merge `be654b938484d4ef4e4796b247bb04bf7efd2049`; protected tree `d90ad717da04c7a28f538a4a0fea2781409dd44e`; historical noncompliance preserved; residual governance risk accepted; the disposition itself grants no Stage 3.72 planning acceptance/Ready/merge or runtime authorization | `stages/STAGE_03_72_GOV_01_HISTORICAL_GOVERNANCE_DEVIATION_DISPOSITION.md` |
| Stage 3.72 Portfolio Position Projection / Honest Market-Unavailable Semantics planning | Complete/canonical through PR #142 squash merge `2144c52f4dc5ba9e917947d396ced1f3c572fe51`; planning authority activated without provider/runtime activation by the planning merge itself | `stages/STAGE_03_72_PORTFOLIO_POSITION_PROJECTION_PLANNING.md` |
| Stage 3.72 Portfolio Position Projection / Cost Basis View runtime | Complete/canonical through PR #145 squash merge `e935f19b68624f0b8be6ca293ead8f4395cf555e` from exact runtime head `b6cd7766a6a584440dfdd7d00c35c34c8553a9ff` after CI #413 / run `34091588982` 10/10 SUCCESS | `stages/STAGE_03_72_PORTFOLIO_POSITION_PROJECTION_IMPLEMENTATION_CLOSURE.md` |
| Stage 3.72 lifecycle/documentation closure | Complete/canonical through PR #146 squash merge `ac0396eff47ce5cd862f41a5ec8ab2803de7fa58`; records purpose, implementation, expected behavior, verification and preserved boundaries | `stages/STAGE_03_72_PORTFOLIO_POSITION_PROJECTION_IMPLEMENTATION_CLOSURE.md` |
| Stage 3.73 Portfolio Time Machine / Historical Position View | Complete/canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; Current ↔ Historical UI, stale-result protection and historical financial witnesses over the existing Stage 3.72 `asOfDate` contract; no new endpoint/schema/provider | `stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md` |
| Stage 3.74 Transaction Correction & Reversal / Ledger Repair UX | Complete/canonical through PR #150 squash merge `0580bf7e98c532202f84bbf9ceacd97aedbe4140` from exact final head `ff6d3efb8bb0d63f8cab55ac82fdf1052946aa02` after CI #435 / run `34117662576` 10/10 SUCCESS; append-only revisions/reversals, effective-ledger materialization, oversell rollback, exact replay, retry-safe Edit/Reverse UX and HTTP contract witnesses | `stages/STAGE_03_74_TRANSACTION_CORRECTION_REVERSAL_IMPLEMENTATION.md` |
| Stage 3.75 Portfolio Cash Flow & Income Truth | Complete/canonical through PR #153 squash merge `57faae841805a5ab11ac959018a315cbada69207`; implementation head `f97f44016d22e783980c47bbde4799a3124623f4` passed CI #459 / run `34155729975` 10/10 SUCCESS; evidence head `fae674518e37d11cad0da28d91682c35bf8bf99a` passed CI #460 / run `34156290105` 10/10 SUCCESS; fresh External review and evidence/no-drift verification APPROVED; effective-ledger cash-flow/income truth, no provider/migration/cache table | `stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md` |
| Stage 3.76 Manual Market Valuation & Portfolio P/L | Complete/canonical through PR #155 squash merge `e00699f8d455bcbaea0c1dc69ce534460fea6ff9` from exact final head `f89e1e793ef505c31bf76de7d43c4d40fdd9c38b`, tree `5971ae4bf5461dd562d7a73289b5942b02f962ba`, after CI #467 / run `34174642209` 10/10 SUCCESS; authenticated `USER_SUPPLIED` RUB valuation, backend-derived market value/unrealized P/L, lifecycle-safe generation and exact historical-date semantics; external provider activation remains NO-GO | `stages/STAGE_03_76_MANUAL_MARKET_VALUATION_IMPLEMENTATION.md` |
| Stage 3.77 Portfolio Money-Weighted Return / XIRR | Complete/canonical through PR #157 squash merge `63d916b447e91c4de54efee5c66b27cf7727be92` from final evidence head `b3bc17e710e306f9d1a4d7be95dfe56b3f7202ba`, tree `e91e28e96416d9510962a66c1399459844574bba`, after protected exact-head CI run `34303895226` 10/10 SUCCESS; backend-owned ACT/365 XIRR, exact seven unavailable reasons, strict explicit BusinessDate, Stage 3.74 external-flow truth and Stage 3.76 exact-date manual terminal valuation; External review APPROVED and no implementation drift | `stages/STAGE_03_77_PORTFOLIO_XIRR_IMPLEMENTATION.md` |

Stage 3.71 lifecycle/documentation closure is canonical through PR #140 squash merge `97c35a7b0fe7e2cd487c63d03b8ba79f80dcfc7b`. `STAGE-03-72-GOV-01` is effective through PR #143 squash merge `be654b938484d4ef4e4796b247bb04bf7efd2049`, with historical noncompliance preserved and residual governance risk accepted. Stage 3.72 planning is canonical through PR #142 squash merge `2144c52f4dc5ba9e917947d396ced1f3c572fe51`, and Stage 3.72 runtime is canonical through PR #145 squash merge `e935f19b68624f0b8be6ca293ead8f4395cf555e` after CI #413 / run `34091588982` 10/10 SUCCESS. Stage 3.72 lifecycle/documentation closure is canonical through PR #146 squash merge `ac0396eff47ce5cd862f41a5ec8ab2803de7fa58`; it authorizes no later Portfolio Position / Cost Basis expansion, production rollout, imported SELL expansion, provider-derived market-valued positions, XIRR/returns, correction/reversal, tax basis, bond NKD, notifications or AI.

Stage 3.73 Portfolio Time Machine is complete/canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS. It is a Web/product projection over the existing Stage 3.72 `asOfDate` contract and authorizes no market provider activation, historical provider market-value/return calculation, new financial engine, API/schema expansion, tax-basis semantics, imported SELL expansion, privacy lifecycle, infrastructure expansion, notifications or AI.

Stages 3.74–3.76 are complete/canonical through protected PRs #150, #153 and #155. Stage 3.76 adds authenticated explicit `USER_SUPPLIED` RUB manual valuation and backend-derived current market value/unrealized P/L while preserving exact-date historical eligibility and lifecycle-safe position generations. It does not activate MOEX, broker, scraping, quote providers, polling, workers, Redis, Kafka, FX, XIRR/TWR, tax logic, notifications, AI, mobile or paid infrastructure. No Stage 3.77+ product/runtime scope is authorized by this registry synchronization.

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
