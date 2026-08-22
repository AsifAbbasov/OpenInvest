# Implementation Roadmap

| Field | Value |
| --- | --- |
| Document ID | ENG-ROADMAP-001 |
| Version | 1.1.53 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal stage ordering |
| Dependencies | Architecture Freeze v1.2 |
| Last Review Date | 2026-08-18 |
| Next Review Date | Before Stage 3.25 evidence-collection plan review, evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

| Stage | Outcome | State |
| --- | --- | --- |
| 0 — Foundation / Bootstrap | Local monorepo skeleton, toolchain, health checks, local PostgreSQL/Redis definition | Complete; awaiting review/commit |
| 1 — Documentation Consolidation | Repository-owned Source of Truth and frozen MVP/architecture registers | Complete; awaiting review/commit |
| 2 — Contract and Canonical Model Freeze | Reviewed MVP API contract, schemas, canonical DTOs, ER draft, migration strategy | Complete |
| Web architecture amendment | Replace Vite skeleton with presentation-only Next.js under ADR-007 | Complete |
| 3 — First Vertical Slice Planning | Plan the first portfolio → transaction → snapshot → API → Web path | Complete |
| 3.1 — Local database foundation | Minimal PostgreSQL schemas/migrations for the first slice | Complete |
| 3.2 — Go API vertical-slice backend | Portfolio, transaction append, snapshot rebuild, summary read | Complete |
| 3.3 — Next.js presentation slice | Render the first slice through the Go API only | Complete |
| 3.4 — End-to-end verification | Prove the complete path and update onboarding docs | Complete |
| 3.5 — Broker file import and reconciliation design | Reduce public-MVP manual-entry risk before broad release | Complete |
| 3.6 — Broker file import vertical slice | User-supplied CSV import review and explicit append-plan generation | Complete |
| 3.7 — Import append planning | Define the reviewed atomic append scope before any ledger mutation implementation | Complete |
| 3.7 — Import append slice | Internal atomic append of user-approved import rows into immutable ledger | Complete |
| 3.8 — Import review append flow planning | Define the internal parse/review/approve/append orchestration boundary | Complete |
| 3.8 — Import review append flow slice | Internal parse/review/approve/append orchestration | Complete |
| 3.9 — Import API boundary planning | Define future public Go API boundary for user-supplied broker-file import | Complete |
| 3.9 — Import API boundary slice | Expose transient CSV review and explicit append through the Go API | Complete |
| 3.10 — Import upload/review UI planning | Define future Next.js presentation-only import upload and review UI before implementation | Complete |
| 3.10 — Import upload/review UI slice | Expose the CSV import review/append flow in Next.js presentation only | Complete |
| 3.11 — Authentication and privacy-boundary planning | Define the future replacement of the local development subject with the approved web auth/session/privacy model | Complete |
| 3.11 — Authentication and privacy-boundary slice | Implement the approved Go API auth/session/privacy-default boundary without frontend auth UI | Complete |
| 3.12 — Web authentication UI planning | Define the future Next.js presentation-only auth/session UI boundary | Complete |
| 3.12 — Web authentication UI slice | Expose registration, login, session shell, refresh, and logout in Next.js presentation only | Complete |
| 3.13 — Instrument catalog planning | Define canonical MVP MOEX share/bond identity before implementation | Complete |
| 3.13 — Instrument catalog slice | Resolve approved MOEX share/bond tickers through the backend-owned catalog boundary | Complete |
| 3.14 — Asset search/card API boundary planning | Define the future Go API asset search/detail boundary before implementation | Complete |
| 3.14 — Asset search/card API boundary slice | Expose the public Go API asset search boundary without fabricated market data or detail provenance | Complete |
| 3.15 — Web asset discovery UI planning | Define the future Next.js presentation-only asset search entry and deferred card-state boundary before implementation | Complete |
| 3.15 — Web asset discovery UI slice | Implement the reviewed Next.js presentation-only asset discovery boundary | Complete |
| 3.15 — Closure governance | Close Stage 3.15 implementation governance after PR #42 merge | Complete |
| 3.16 — Repository audit planning | Plan the mandatory full repository audit before the next implementation stage | Complete |
| 3.16 — Repository audit fixes | Fix mandatory audit `REQUEST CHANGES` findings before the next implementation stage | Complete / merged |
| 3.16 — Closure governance | Record audit-fix completion and preserve the next planning gate | Complete |
| 3.17 — Privacy lifecycle planning | Define the reviewed account deletion, anonymization, backup-destruction, and retention execution boundary | Complete / merged through PR #46 |
| 3.18 — Privacy contract and security proposal | Define the candidate account-deletion contract, security, cryptographic-erasure, restore, and operational gates | Complete / merged through PR #47 |
| 3.19 — Privacy security and ADR proposal | Define provider-neutral cryptographic-erasure, deletion-marker, restore, and separation-of-duties controls | Complete / merged through PR #48 |
| 3.20 — Privacy lifecycle threat-model proposal | Define the future privacy-lifecycle threat boundary, residual risks, and review evidence | Complete / merged through PR #49 |
| 3.21 — Privacy data-inventory proposal | Map observed privacy-relevant fields and external evidence gaps before any deletion/anonymization design | Complete / merged through PR #50 |
| 3.22 — Privacy key-custody and destruction-proof proposal | Define provider-neutral custody, irreversible destruction proof, and fail-closed evidence requirements | Complete / merged through PR #51 |
| 3.23 — Privacy deletion-marker control-plane proposal | Define restricted marker lifecycle, snapshot integrity, and fail-closed isolated restore replay | Complete / merged through PR #52 |
| 3.24 — Privacy Security Review readiness dossier | Define required evidence, questions, outcomes, and record for the mandatory formal Security Review | Complete / merged through PR #53 |
| 3.25 — Privacy Security Review evidence-collection plan | Define minimized, integrity-protected, independently verified evidence collection before formal Security Review | Active / proposal only |
| 3.27 — Import financial identity and cash-flow semantics remediation | Close audit P1-02/P1-03/P1-04 with persisted import identity, amount-aware cash reconciliation, and fail-closed cash-flow fee semantics | Implementation verified / pre-commit review approved / implementation commit `19a8abb` pushed / Draft PR #55 open / CI run #83 passed on implementation head / closure requires final-head green CI, required PR review, explicit human approval, and squash merge |

The repository already exists because Stage 0 was executed before the refined roadmap. Stage 3
therefore implements the first vertical slice incrementally instead of recreating the repository.

Stage 3.27 is a separately authorized narrow repository-audit remediation and does not authorize
product-scope expansion. No further implementation stage begins without a separately reviewed
planning or remediation gate.

No AI, Tax Export, email, mobile, premium, direct broker API synchronization, credential scraping, or
unnecessary worker implementation enters these stages without separate approval.
