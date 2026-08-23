# Implementation Roadmap

| Field | Value |
| --- | --- |
| Document ID | ENG-ROADMAP-001 |
| Version | 1.1.64 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal stage ordering |
| Dependencies | Architecture Freeze v1.2 |
| Last Review Date | 2026-08-24 |
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
| 3.27 — Import financial identity and cash-flow semantics remediation | Close audit P1-02/P1-03/P1-04 with persisted import identity, amount-aware cash reconciliation, and fail-closed cash-flow fee semantics | Complete / merged through PR #55 at `6e8c806de857f844954f1db513487357dfe90187` |
| 3.27 — Closure governance | Record canonical merge, final review, CI, and human approval evidence | Complete / PR #58 |
| 3.28 — Authentication security remediation | Close audit P1-01/P1-05 with refresh-family replay containment and bounded Argon2 admission | Implementation complete / merged through PR #59 at `dc83f5f3a11da164e6809593861d96ccf47b29ca` |
| 3.28 — Closure governance | Record canonical merge, CI #114, renewed independent approval, human authorization, and residual P2/P3 scope | Complete / merged through PR #60 at `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` |
| 3.29 — Input and contract hardening | Close audit P2-05/P2-06/P2-07/P2-08/P2-15 across client/import contracts and snapshot persistence bounds | Implementation complete / merged through PR #61 at `7331d3f34783baec3997497d1a79b78eaa558bd4` |
| 3.29 — Closure governance | Record canonical implementation merge, CI #124, first `REQUEST CHANGES`, blocker remediation, renewed independent approval, human authorization, and remaining 12 P2 / 10 P3 scope | Complete / merged through PR #62 at `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` |
| 3.30 — Import review integrity | Close audit P2-02/P2-03/P2-04 with semantic review-token binding, parser-owned row admission, and targeted full-history reconciliation | Implementation complete / merged through PR #63 at `8f68dd18800918e6a9882e995e13dba2723dc929` |
| 3.30 — Closure governance | Record canonical implementation merge, CI #128, independent approval, human authorization, and remaining 9 P2 / 10 P3 scope | Complete / merged through PR #64 at `ae6497050692798795efb85678af64db97cc5f53` |
| 3.31 — Authentication operational hardening | Close audit P2-01/P2-14 with bounded logout admission and finite auth-limiter lifecycle | Implementation complete / merged through PR #65 at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897` |
| 3.31 — Closure governance | Record canonical implementation merge, CI #133, independent approval, human authorization, and remaining 7 P2 / 10 P3 scope | Complete / merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` |
| 3.32 — Exact idempotency replay and browser retry recovery | Close audit P2-09/P2-13 with exact original-response replay and principal-isolated browser retry identity | Implementation complete / merged through PR #67 at `0623d5ef326cd783b7dc0417dbcb02f18c506171` |
| 3.32 — Closure governance | Record canonical implementation merge, CI #181, first `REQUEST CHANGES`, P2-13 remediation, repeat independent `APPROVED`, human authorization, and remaining 5 P2 / 10 P3 scope | Complete / merged through PR #68 at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` |
| 3.33 — Snapshot rebuild accuracy and PostgreSQL runtime immutability | Close audit P2-10/P2-11/P2-12 with DB-owned exact snapshot rebuild reporting, one-pass rebuild planning, and fail-closed runtime credential-graph enforcement | Implementation complete / merged through PR #69 at `87a7c38e16062a5f3fcef3727f60c0c6741eb805` |
| 3.33 — Closure governance | Preserve CI #199, two independent P2-12 `REQUEST CHANGES` cycles, final independent `APPROVED`, separate closure CI/review/human authorization, and remaining 2 P2 / 10 P3 scope | Complete / canonical through PR #70 at `71a1faeb97d33d05f2936111b53f1285edddabe9` |

The repository already exists because Stage 0 was executed before the refined roadmap. Stage 3
therefore implements the first vertical slice incrementally instead of recreating the repository.

Stages 3.27, 3.28, 3.29, 3.30, 3.31, 3.32, and 3.33 are separately authorized narrow repository-audit remediations and do not
authorize product-scope expansion. Stage 3.25 privacy evidence planning remains separate. After
Stage 3.33 canonical closure, the original audit backlog contains exactly 2 P2 and 10 P3 findings. The remaining P2 findings are P2-16 and P2-17. The cross-stage audit register is `audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md`, and all future remediation dossiers must satisfy `audit/REMEDIATION_DOCUMENTATION_STANDARD.md`.

No further implementation stage begins without a separately reviewed planning or remediation gate.

No AI, Tax Export, email, mobile, premium, direct broker API synchronization, credential scraping, or
unnecessary worker implementation enters these stages without separate approval.
