# OpenInvest documentation index

This is a navigation map, not a contract, deployment approval or replacement for dated decision/evidence records. Start with the task below; do not infer current status from an old document title or embedded candidate flag.

## Start here

- [Project overview, architecture diagram and local onboarding](../README.md).
- [GitHub configuration and exact-head CI](../.github/README.md).
- [Current navigation reconciliation](governance/DOCUMENTATION_NAVIGATION_RECONCILIATION.md).
- [Documentation standard](governance/DOCUMENTATION_STANDARD.md) and [migration inventory](governance/DOCUMENTATION_MIGRATION_INVENTORY.csv): ten-line explanations, full normative contracts and preserved historical evidence.
- [Phase 3 purpose review](governance/DOCUMENTATION_PHASE3_DISPOSITION.md): all ten remaining candidates are full registers/contracts or preserved research; dated provider proposals do not supersede the source registry.

## Module #1 — verified technical integration

- [Technical acceptance and integration record](audit/MODULE1_TECHNICAL_CLOSURE.md): `MODULE1_TECHNICAL_ACCEPTANCE=PASS`, `MODULE1_INTEGRATED=YES`; PR #268 and exact post-merge CI. This does not establish new formal module closure or deployment approval.
- [Architecture acceptance and integration summary](audit/MODULE1_ARCHITECTURE_ACCEPTANCE_AND_INTEGRATION_READINESS.md): scoped acceptance matrix and residuals through immutable references.
- [Approved email-ownership addendum](ADR/ADR-010-email-ownership-approval-candidate.md) and [approved access-JWT identity contract](api/ACCESS_JWT_IDENTITY_POLICY_V1.md).
- [JWT02/JWT03 disposition](audit/MODULE1_JWT02_JWT03_POLICY_DECISION.md): approved future-iat residual, unchanged parser and historical RED evidence.
- [PR265 supplemental evidence provenance](audit/MODULE1_PR265_EVIDENCE_PROVENANCE.md) and [original owner approval record](audit/MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md): later procedural authorization does not extend the original campaign deadline.
- [Historical auth/session/privacy closure](audit/AUTH_SESSION_PRIVACY_ADVERSARIAL_AUDIT_CLOSURE.md) and [email-ownership remediation](audit/AUTH_ADV_02_EMAIL_OWNERSHIP_REMEDIATION_V1.md): preserved dated evidence, not substitutes for the current technical record.

## Contracts and authority

- [Source of Truth](SOURCE_OF_TRUTH.md): established human-readable product/runtime authority; its dated audit/lifecycle paragraphs must be read with later explicitly scoped records, not treated as a fresh global verification.
- [Architecture Freeze v1.2](ARCHITECTURE_FREEZE_v1.2.md), [Document 42 amendments](specifications/current/DOCUMENT_42_ARCHITECTURE_AMENDMENTS_v1.1.md) and [Document 43 closure](specifications/current/DOCUMENT_43_ARCHITECTURE_CLOSURE_v1.2.md): architecture boundaries; navigation does not amend them.
- [OpenAPI](../openapi/openapi.yaml) and [component schemas](../openapi/components/schemas.yaml): executable HTTP contract, including registration/completion/session operations.
- [Architecture decisions](ADR/): accepted decisions retain their authority and historical bytes. The [original ADR-010](ADR/ADR-010-email-ownership-before-account-activation.md) is migration-hash-pinned and unchanged; the approved addendum is a distinct record.
- [Stage 2 API baseline](api/API_CONTRACT_STAGE_02.md), [canonical model](domain/CANONICAL_MODEL_STAGE_02.md), [ER model](database/ER_MODEL_STAGE_02.md) and [migration strategy](database/MIGRATION_STRATEGY_STAGE_02.md): full reference contracts; later OpenAPI/accepted decisions govern superseded portions.
- [Legacy specifications](specifications/legacy/) and [English reference translations](specifications/legacy-translations/en/): preserve originals; Documents 42–43 and accepted decisions govern conflicts. Historical document number 14 is split into 14A/14B.

## Operations and external data

- [Data Source Registry](registries/DATA_SOURCE_REGISTRY.md): sole provider/source-use authority; implemented adapters do not imply authorized production traffic.
- [Corporate Actions research](research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md) and [other source research](research/): provider rights, costs and unresolved activation questions.
- [Browser/HTTP deployment](operations/HTTP_BROWSER_SECURITY_DEPLOYMENT.md), [abuse protection deployment](operations/ABUSE_PROTECTION_DEPLOYMENT.md) and [replay policy operations](operations/OI_NEW_04_REPLAY_POLICY_OPERATIONS.md): complete operational requirements remain outside the ten-line explanatory limit.
- [Infrastructure](../infrastructure/README.md), [PostgreSQL runtime role](../infrastructure/postgres/runtime/README.md), [scripts](../scripts/README.md) and [financial-import fixtures](../tests/financial/import/README.md): setup, permissions and verification instructions.

## Planning and history

- [Roadmap](ROADMAP.md), [product-risk refinement](product/MVP_PRODUCT_RISK_REFINEMENT.md) and [backlog](BACKLOG_V2.md): sequencing and product constraints, not new runtime/deployment authorization.
- [Decision questions](OPEN_QUESTIONS.md) and [version matrix](VERSION_MATRIX.md): decision/reference metadata, not current runtime authority.
- [Implementation Log](IMPLEMENTATION_LOG.md), [changelog](CHANGELOG.md) and [stage dossiers](stages/): dated planning, implementation, closure and evidence; no historical decisions are rewritten by this index.
- [Historical Stage 3.72 reconciliation](governance/REPOSITORY_DOCUMENTATION_RECONCILIATION.md) and [historical feature reconciliation](governance/HISTORICAL_FEATURE_FORENSIC_DOCUMENTATION_RECONCILIATION.md): preserved records, not present-day lifecycle summaries.
- [Full prior index and stage/PR/commit registry at the Phase 2 baseline](https://github.com/AsifAbbasov/OpenInvest/blob/ab39b4979fc2921d29b62c8862658dc4a2ad1f6a/docs/DOCUMENT_INDEX.md): exact historical mappings and metadata remain accessible without duplicating their chronology here.

## Security evidence

- [Original 32-finding remediation register](audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md) and [audit directory](audit/): historical scope-specific reports; a closure record is not a universal production-security claim.
- [Module #10 provider audit closure](audit/PROVIDERS_EXTERNAL_API_BUDGET_RESILIENCE_ADVERSARIAL_AUDIT_CLOSURE.md): original repository-side disposition.
- [Module #12 repository-side remediation/revalidation](audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_ADVERSARIAL_AUDIT_CLOSURE.md), [P2 remediation](audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_P2_REMEDIATION_V1.md) and [CORS remediation](audit/MODULE12_V2_CORS_01_REMEDIATION_V1.md): later evidence; embedded review/closure flags remain historical. Actual edge/CDN normalization, provider quotas and deployed topology/TLS/network controls remain deployment-specific evidence gaps.
- [OI-NEW P2 technical closure](audit/OI_NEW_P2_REMEDIATION_CLOSURE.md): bounded findings and immutable baseline, not the mutable develop tip.

## Reading and maintenance rules

Use exact-SHA source snapshots and run IDs when interpreting evidence. Preserve historical FAIL/RED as observed; later remediation or owner risk acceptance is a separate disposition. Check approval dates and scope before resolving conflicting statuses. Navigation/onboarding retain useful diagrams, commands and maps; standalone explanations follow the ten-line standard. Class B/C documents in the inventory require preservation-first review, not automatic shortening. SMTP activation, account recovery, immediate JWT revocation and deployment remain separately gated.
