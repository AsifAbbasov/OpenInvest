# OpenInvest documentation index
Navigation only; current contracts and original historical outcomes remain in the repository. Technical acceptance is not deployment or new formal module closure.

## Start here

- [Project overview, architecture diagram and local onboarding](../README.md); [GitHub configuration and exact-head CI](../.github/README.md).
- [Current navigation reconciliation](governance/DOCUMENTATION_NAVIGATION_RECONCILIATION.md); [Documentation standard](governance/DOCUMENTATION_STANDARD.md).
- [migration inventory](governance/DOCUMENTATION_MIGRATION_INVENTORY.csv); [Phase 3 purpose review](governance/DOCUMENTATION_PHASE3_DISPOSITION.md).

## Module #1

- [Technical acceptance and integration record](audit/MODULE1_TECHNICAL_CLOSURE.md); [Architecture acceptance and integration summary](audit/MODULE1_ARCHITECTURE_ACCEPTANCE_AND_INTEGRATION_READINESS.md).
- [Approved email-ownership addendum](ADR/ADR-010-email-ownership-approval-candidate.md); [approved access-JWT identity contract](api/ACCESS_JWT_IDENTITY_POLICY_V1.md).
- [JWT02/JWT03 disposition](audit/MODULE1_JWT02_JWT03_POLICY_DECISION.md); [PR265 supplemental evidence provenance](audit/MODULE1_PR265_EVIDENCE_PROVENANCE.md).
- [original owner approval record](audit/MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md); [CORS remediation](audit/MODULE12_V2_CORS_01_REMEDIATION_V1.md).

## Contracts and authority

- [Source of Truth](SOURCE_OF_TRUTH.md); [Architecture Freeze v1.2](ARCHITECTURE_FREEZE_v1.2.md).
- [Document 42 amendments](specifications/current/DOCUMENT_42_ARCHITECTURE_AMENDMENTS_v1.1.md); [Document 43 closure](specifications/current/DOCUMENT_43_ARCHITECTURE_CLOSURE_v1.2.md).
- [OpenAPI](../openapi/openapi.yaml); [component schemas](../openapi/components/schemas.yaml).
- [Architecture decisions](ADR/); [original ADR-010](ADR/ADR-010-email-ownership-before-account-activation.md).
- [Stage 2 API baseline](api/API_CONTRACT_STAGE_02.md); [canonical model](domain/CANONICAL_MODEL_STAGE_02.md).
- [ER model](database/ER_MODEL_STAGE_02.md); [migration strategy](database/MIGRATION_STRATEGY_STAGE_02.md).
- [Legacy specifications](specifications/legacy/); [English reference translations](specifications/legacy-translations/en/).

## Operations and data

- [Data Source Registry](registries/DATA_SOURCE_REGISTRY.md); [Corporate Actions research](research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md).
- [other source research](research/); [Browser/HTTP deployment](operations/HTTP_BROWSER_SECURITY_DEPLOYMENT.md).
- [abuse protection deployment](operations/ABUSE_PROTECTION_DEPLOYMENT.md); [replay policy operations](operations/OI_NEW_04_REPLAY_POLICY_OPERATIONS.md).
- [Infrastructure](../infrastructure/README.md); [PostgreSQL runtime role](../infrastructure/postgres/runtime/README.md).
- [scripts](../scripts/README.md); [financial-import fixtures](../tests/financial/import/README.md).

## Planning and historical evidence

- [Historical auth/session/privacy closure](audit/AUTH_SESSION_PRIVACY_ADVERSARIAL_AUDIT_CLOSURE.md); [email-ownership remediation](audit/AUTH_ADV_02_EMAIL_OWNERSHIP_REMEDIATION_V1.md).
- [Roadmap](ROADMAP.md); [product-risk refinement](product/MVP_PRODUCT_RISK_REFINEMENT.md).
- [backlog](BACKLOG_V2.md); [Decision questions](OPEN_QUESTIONS.md).
- [version matrix](VERSION_MATRIX.md); [Implementation Log](IMPLEMENTATION_LOG.md).
- [changelog](CHANGELOG.md); [stage dossiers](stages/).
- [Historical Stage 3.72 reconciliation](governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md); [historical feature reconciliation](governance/HISTORICAL_FEATURE_FORENSIC_DOCUMENTATION_RECONCILIATION.md).
- [Full prior index and stage/PR/commit registry at the Phase 2 baseline](https://github.com/AsifAbbasov/OpenInvest/blob/ab39b4979fc2921d29b62c8862658dc4a2ad1f6a/docs/DOCUMENT_INDEX.md); [Original 32-finding remediation register](audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md).
- [audit directory](audit/); [Module #10 provider audit closure](audit/PROVIDERS_EXTERNAL_API_BUDGET_RESILIENCE_ADVERSARIAL_AUDIT_CLOSURE.md).
- [Module #12 repository-side remediation/revalidation](audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_ADVERSARIAL_AUDIT_CLOSURE.md); [P2 remediation](audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_P2_REMEDIATION_V1.md).
- [OI-NEW P2 technical closure](audit/OI_NEW_P2_REMEDIATION_CLOSURE.md).

OpenAPI, accepted decisions and the source registry retain precedence. Read exact dates/SHA/run evidence; preserve RED/FAIL; SMTP, recovery, immediate JWT revocation and deployment remain separately gated.
