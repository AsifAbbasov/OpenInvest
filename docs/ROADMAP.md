# Implementation Roadmap

| Field | Value |
| --- | --- |
| Document ID | ENG-ROADMAP-001 |
| Version | 1.1.3 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal stage ordering |
| Dependencies | Architecture Freeze v1.2 |
| Last Review Date | 2026-06-27 |
| Next Review Date | 2026-12-26 |

| Stage | Outcome | State |
| --- | --- | --- |
| 0 — Foundation / Bootstrap | Local monorepo skeleton, toolchain, health checks, local PostgreSQL/Redis definition | Complete; awaiting review/commit |
| 1 — Documentation Consolidation | Repository-owned Source of Truth and frozen MVP/architecture registers | Complete; awaiting review/commit |
| 2 — Contract and Canonical Model Freeze | Reviewed MVP API contract, schemas, canonical DTOs, ER draft, migration strategy | Complete |
| Web architecture amendment | Replace Vite skeleton with presentation-only Next.js under ADR-007 | Complete |
| 3 — First Vertical Slice Planning | Plan the first portfolio → transaction → snapshot → API → Web path | Complete |
| 3.1 — Local database foundation | Minimal PostgreSQL schemas/migrations for the first slice | Complete |
| 3.2 — Go API vertical-slice backend | Portfolio, transaction append, snapshot rebuild, summary read | Complete |
| 3.3 — Next.js presentation slice | Render the first slice through the Go API only | Planned after backend slice approval |
| 3.4 — End-to-end verification | Prove the complete path and update onboarding docs | Planned |
| 3.5 — Broker file import and reconciliation design | Reduce public-MVP manual-entry risk before broad release | Recommended / requires approval |
| 3.6 — Broker file import vertical slice | User-supplied file import with review and append-only reconciliation | Candidate / requires design approval |

The repository already exists because Stage 0 was executed before the refined roadmap. Stage 3
therefore implements the first vertical slice incrementally instead of recreating the repository.

No AI, Tax Export, email, mobile, premium, direct broker API synchronization, credential scraping, or
unnecessary worker implementation enters these stages without separate approval.
