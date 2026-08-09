# Stage 3.16 - Repository Audit Coverage Manifest

| Field | Value |
| --- | --- |
| Audit target SHA | `74eebe9ec8231764f21ce384c4690d073d0273da` |
| Tracked path count | 200 |
| Path-list SHA-256 | `1fd740a8d5bca3afd05daa3268c079bae3a7a331a043ecdff5d35734ac77604e` |
| Audit task | `019fcdf5-b722-7513-9860-286e83f4c44c` |

Each path below is classified against the immutable audit target. `Excluded` is reserved for archived legacy material that is not active runtime or control authority.

| Path | Disposition | Evidence |
| --- | --- | --- |
| `.editorconfig` | Audited | E-SOURCE |
| `.env.example` | Audited | E-SOURCE |
| `.github/PULL_REQUEST_TEMPLATE.md` | Audited | E-SOURCE |
| `.github/README.md` | Audited | E-SOURCE |
| `.github/workflows/ci.yml` | Audited | E-CONFIG |
| `.gitignore` | Audited | E-SOURCE |
| `README.md` | Audited | E-DOC |
| `backend-go/cmd/api/main.go` | Audited | E-SOURCE |
| `backend-go/cmd/api/main_test.go` | Audited | E-TEST |
| `backend-go/cmd/validate-migrations/main.go` | Audited | E-SOURCE |
| `backend-go/cmd/validate-openapi/main.go` | Audited | E-SOURCE |
| `backend-go/go.mod` | Audited | E-CONFIG |
| `backend-go/go.sum` | Audited | E-LOCK |
| `backend-go/internal/auth/models.go` | Audited | E-SOURCE |
| `backend-go/internal/auth/password.go` | Audited | E-SOURCE |
| `backend-go/internal/auth/service.go` | Audited | E-SOURCE |
| `backend-go/internal/auth/service_test.go` | Audited | E-TEST |
| `backend-go/internal/auth/tokens.go` | Audited | E-SOURCE |
| `backend-go/internal/decimal/decimal.go` | Audited | E-SOURCE |
| `backend-go/internal/decimal/decimal_test.go` | Audited | E-TEST |
| `backend-go/internal/httpapi/api.go` | Audited | E-SOURCE |
| `backend-go/internal/httpapi/api_test.go` | Audited | E-TEST |
| `backend-go/internal/httpapi/auth_test.go` | Audited | E-TEST |
| `backend-go/internal/importer/importer.go` | Audited | E-SOURCE |
| `backend-go/internal/importer/importer_test.go` | Audited | E-TEST |
| `backend-go/internal/importflow/importflow.go` | Audited | E-SOURCE |
| `backend-go/internal/importflow/importflow_test.go` | Audited | E-TEST |
| `backend-go/internal/postgres/auth_store.go` | Audited | E-SOURCE |
| `backend-go/internal/postgres/auth_store_integration_test.go` | Audited | E-TEST |
| `backend-go/internal/postgres/store.go` | Audited | E-SOURCE |
| `backend-go/internal/postgres/store_integration_test.go` | Audited | E-TEST |
| `backend-go/internal/postgres/store_test.go` | Audited | E-TEST |
| `backend-go/internal/verticalslice/models.go` | Audited | E-SOURCE |
| `backend-go/internal/verticalslice/service.go` | Audited | E-SOURCE |
| `backend-go/internal/verticalslice/service_test.go` | Audited | E-TEST |
| `docker-compose.yml` | Audited | E-CONFIG |
| `docs/ADR/ADR-001-use-go-fiber.md` | Audited | E-DOC |
| `docs/ADR/ADR-002-use-postgresql.md` | Audited | E-DOC |
| `docs/ADR/ADR-003-use-openapi-first.md` | Audited | E-DOC |
| `docs/ADR/ADR-004-use-snapshots.md` | Audited | E-DOC |
| `docs/ADR/ADR-005-privacy-by-design.md` | Audited | E-DOC |
| `docs/ADR/ADR-006-contract-and-canonical-model-freeze.md` | Audited | E-DOC |
| `docs/ADR/ADR-007-use-nextjs-for-web-frontend.md` | Audited | E-DOC |
| `docs/ADR/README.md` | Audited | E-DOC |
| `docs/ARCHITECTURE_FREEZE_v1.2.md` | Audited | E-DOC |
| `docs/ARCHITECTURE_FREEZE_v1.md` | Audited | E-DOC |
| `docs/BACKLOG_V2.md` | Audited | E-DOC |
| `docs/CHANGELOG.md` | Audited | E-DOC |
| `docs/DOCUMENT_INDEX.md` | Audited | E-DOC |
| `docs/IMPLEMENTATION_LOG.md` | Audited | E-DOC |
| `docs/OPEN_QUESTIONS.md` | Audited | E-DOC |
| `docs/REVIEW_WORKFLOW.md` | Audited | E-DOC |
| `docs/ROADMAP.md` | Audited | E-DOC |
| `docs/SOURCE_OF_TRUTH.md` | Audited | E-DOC |
| `docs/VERSION_MATRIX.md` | Audited | E-DOC |
| `docs/api/API_CONTRACT_STAGE_02.md` | Audited | E-DOC |
| `docs/database/ER_MODEL_STAGE_02.md` | Audited | E-DOC |
| `docs/database/MIGRATION_STRATEGY_STAGE_02.md` | Audited | E-DOC |
| `docs/domain/CANONICAL_MODEL_STAGE_02.md` | Audited | E-DOC |
| `docs/product/MVP_PRODUCT_RISK_REFINEMENT.md` | Audited | E-DOC |
| `docs/registries/DATA_SOURCE_REGISTRY.md` | Audited | E-DOC |
| `docs/specifications/current/DOCUMENT_42_ARCHITECTURE_AMENDMENTS_v1.1.md` | Audited | E-DOC |
| `docs/specifications/current/DOCUMENT_43_ARCHITECTURE_CLOSURE_v1.2.md` | Audited | E-DOC |
| `docs/specifications/legacy/DOCUMENT_00_PROJECT_MANIFEST_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_01_PRODUCT_VISION_PRD_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_02_SYSTEM_ARCHITECTURE_BLUEPRINT_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_03_DOMAIN_MODEL_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_04_SYSTEM_DESIGN_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_05_DATABASE_POSTGRESQL_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_06_API_BACKEND_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_07_FRONTEND_UX_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_08_SECURITY_PRIVACY_TRUST_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_09_PRODUCT_VISION_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_10_ENGINEERING_STANDARDS_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_11_SECURITY_BLUEPRINT_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_12_BACKEND_PIPELINE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_13_DATABASE_DOMAIN_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_14A_FRONTEND_ARCHITECTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_14B_MATHEMATICAL_ENGINE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_15_BACKEND_ARCHITECTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_16_DATABASE_EVENT_SOURCING_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_17_BUSINESS_LOGIC_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_18_SYSTEM_INTEGRATION_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_19_FRONTEND_ARCHITECTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_20_TESTING_STRATEGY_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_21_SECURITY_PRIVACY_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_22_LEGAL_ARCHITECTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_23_MOBILE_ARCHITECTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_24_AI_AGENTS_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_25_PRODUCT_ANALYTICS_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_26_DEVOPS_INFRASTRUCTURE_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_27_FINAL_PRODUCT_BLUEPRINT_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_28_ARCHITECTURE_REFINEMENT_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_29_DDD_CLEAN_ARCHITECTURE_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_30_SECURITY_ZERO_TRUST_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_31_DATABASE_BIBLE_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_32_OPENAPI_FIRST_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_33_DESIGN_SYSTEM_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_34_TESTING_CONSTITUTION_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_35_MOBILE_BIBLE_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_36_AI_ARCHITECTURE_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_37_DEVOPS_WORKFLOW_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_38_PRODUCT_ANALYTICS_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_39_SCALING_DR_v2.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_40_CODEX_EXECUTION_v3.0.md` | Excluded | E-ARCHIVE |
| `docs/specifications/legacy/DOCUMENT_41_ANTI_PATTERNS_v1.0.md` | Excluded | E-ARCHIVE |
| `docs/stages/STAGE_00_FOUNDATION.md` | Audited | E-DOC |
| `docs/stages/STAGE_01_DOCUMENTATION_CONSOLIDATION.md` | Audited | E-DOC |
| `docs/stages/STAGE_02_CONTRACT_AND_CANONICAL_MODEL.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_01_DATABASE_FOUNDATION.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_02_GO_API_VERTICAL_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_03_NEXTJS_PRESENTATION_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_04_END_TO_END_VERIFICATION.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_05_BROKER_FILE_IMPORT_RECONCILIATION_DESIGN.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_06_IMPORT_RECONCILIATION_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_07_IMPORT_APPEND_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_07_IMPORT_APPEND_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_09_IMPORT_API_BOUNDARY_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_09_IMPORT_API_BOUNDARY_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_10_IMPORT_UPLOAD_REVIEW_UI_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_10_IMPORT_UPLOAD_UI_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_11_AUTH_PRIVACY_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_11_AUTH_PRIVACY_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_12_AUTH_UI_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_12_AUTH_UI_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_13_INSTRUMENT_CATALOG_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_13_INSTRUMENT_CATALOG_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_14_ASSET_API_BOUNDARY_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_14_ASSET_API_BOUNDARY_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_15_WEB_ASSET_DISCOVERY_UI_SLICE.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_16_REPOSITORY_AUDIT_PLANNING.md` | Audited | E-DOC |
| `docs/stages/STAGE_03_FIRST_VERTICAL_SLICE.md` | Audited | E-DOC |
| `docs/stages/WEB_FRONTEND_ARCHITECTURE_AMENDMENT.md` | Audited | E-DOC |
| `frontend-next/AGENTS.md` | Audited | E-SOURCE |
| `frontend-next/next.config.ts` | Audited | E-SOURCE |
| `frontend-next/package.json` | Audited | E-CONFIG |
| `frontend-next/pnpm-lock.yaml` | Audited | E-SOURCE |
| `frontend-next/pnpm-workspace.yaml` | Audited | E-SOURCE |
| `frontend-next/public/.gitkeep` | Audited | E-SOURCE |
| `frontend-next/src/app/assets/page.tsx` | Audited | E-SOURCE |
| `frontend-next/src/app/layout.tsx` | Audited | E-SOURCE |
| `frontend-next/src/app/page.tsx` | Audited | E-SOURCE |
| `frontend-next/src/app/portfolios/[portfolioId]/page.tsx` | Audited | E-SOURCE |
| `frontend-next/src/app/styles.css` | Audited | E-SOURCE |
| `frontend-next/src/common/README.md` | Audited | E-SOURCE |
| `frontend-next/src/common/api/openinvest.ts` | Audited | E-SOURCE |
| `frontend-next/src/common/presentation/format.ts` | Audited | E-SOURCE |
| `frontend-next/src/features/README.md` | Audited | E-SOURCE |
| `frontend-next/src/features/assets/assetAccessibility.ts` | Audited | E-SOURCE |
| `frontend-next/src/features/assets/assetSearchState.ts` | Audited | E-SOURCE |
| `frontend-next/src/features/assets/components/AssetDiscoverySlice.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/auth/components/AuthForm.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/auth/components/AuthShell.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/auth/session.ts` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/components/CreatePortfolioForm.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/components/DashboardSlice.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/components/ImportUploadReviewPanel.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx` | Audited | E-SOURCE |
| `frontend-next/src/features/portfolio/loadGuard.ts` | Audited | E-SOURCE |
| `frontend-next/tests/asset-accessibility.test.mjs` | Audited | E-TEST |
| `frontend-next/tests/asset-component-contract.test.mjs` | Audited | E-TEST |
| `frontend-next/tests/asset-search-state.test.mjs` | Audited | E-TEST |
| `frontend-next/tests/auth-session.test.mjs` | Audited | E-TEST |
| `frontend-next/tests/openinvest-api.test.mjs` | Audited | E-TEST |
| `frontend-next/tests/portfolio-load-guard.test.mjs` | Audited | E-TEST |
| `frontend-next/tsconfig.json` | Audited | E-SOURCE |
| `infrastructure/README.md` | Audited | E-SOURCE |
| `infrastructure/postgres/migrations/000001_stage_03_01_vertical_slice.down.sql` | Audited | E-MIGRATION |
| `infrastructure/postgres/migrations/000001_stage_03_01_vertical_slice.up.sql` | Audited | E-MIGRATION |
| `infrastructure/postgres/migrations/000002_stage_03_11_auth_privacy.down.sql` | Audited | E-MIGRATION |
| `infrastructure/postgres/migrations/000002_stage_03_11_auth_privacy.up.sql` | Audited | E-MIGRATION |
| `microservice-python/openinvest_analytics/__init__.py` | Audited | E-SOURCE |
| `microservice-python/openinvest_analytics/main.py` | Audited | E-SOURCE |
| `microservice-python/pyproject.toml` | Audited | E-CONFIG |
| `microservice-python/tests/test_health.py` | Audited | E-TEST |
| `microservice-python/uv.lock` | Audited | E-LOCK |
| `openapi/components/responses.yaml` | Audited | E-SOURCE |
| `openapi/components/schemas.yaml` | Audited | E-SOURCE |
| `openapi/examples/assets.json` | Audited | E-SOURCE |
| `openapi/examples/auth.json` | Audited | E-SOURCE |
| `openapi/examples/dashboard.json` | Audited | E-SOURCE |
| `openapi/examples/dividends.json` | Audited | E-SOURCE |
| `openapi/examples/errors.json` | Audited | E-SOURCE |
| `openapi/examples/imports.json` | Audited | E-SOURCE |
| `openapi/examples/operations.json` | Audited | E-SOURCE |
| `openapi/examples/portfolios.json` | Audited | E-SOURCE |
| `openapi/examples/transactions.json` | Audited | E-SOURCE |
| `openapi/openapi.yaml` | Audited | E-SOURCE |
| `package.json` | Audited | E-CONFIG |
| `pnpm-lock.yaml` | Audited | E-LOCK |
| `scripts/README.md` | Audited | E-SOURCE |
| `scripts/stage-03-04-smoke.sh` | Audited | E-SOURCE |
| `tests/financial/import/README.md` | Audited | E-SOURCE |
| `tests/financial/import/conflicts_stage_03_06.csv` | Audited | E-SOURCE |
| `tests/financial/import/formula_injection_stage_03_06.csv` | Audited | E-SOURCE |
| `tests/financial/import/valid_stage_03_06.csv` | Audited | E-SOURCE |

## Evidence Key

| Code | Review evidence |
| --- | --- |
| E-SOURCE | Manual review of active implementation, interfaces, dependency direction, and runtime behavior; checked with the repository verification suite. |
| E-TEST | Manual review of test intent and boundary coverage; executed by the applicable Go, Node, Python, or end-to-end verification command. |
| E-MIGRATION | SQL safety and rollback-pair review; checked by `go run ./cmd/validate-migrations` and the PostgreSQL CI job. |
| E-CONFIG | Manual review of CI, Docker, package, and runtime configuration; checked by configuration and build validation. |
| E-LOCK | Structured lockfile review and frozen-install consistency check. |
| E-DOC | Manual review for Source of Truth, roadmap, architecture, API, privacy/security, test, dependency, cost, and ADR consistency. |
| E-ARCHIVE | Excluded: archival legacy material retained for traceability and not active runtime/control authority under SOT-001. |

## Audit Boundary

This manifest records audit coverage only. The associated audit report retains the `REQUEST CHANGES` verdict until every blocking item is independently re-reviewed and approved.
