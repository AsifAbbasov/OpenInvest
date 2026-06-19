# Implementation Roadmap

| Field | Value |
| --- | --- |
| Document ID | ENG-ROADMAP-001 |
| Version | 1.0.0 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Informal stage ordering |
| Dependencies | Architecture Freeze v1.2 |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |

| Stage | Outcome | State |
| --- | --- | --- |
| 0 — Foundation / Bootstrap | Local monorepo skeleton, toolchain, health checks, local PostgreSQL/Redis definition | Complete; awaiting review/commit |
| 1 — Documentation Consolidation | Repository-owned Source of Truth and frozen MVP/architecture registers | Complete; awaiting review/commit |
| 2 — OpenAPI Freeze | Reviewed MVP API contract, schemas, errors, idempotency, generated validation | Planned |
| 3 — Bootstrap Hardening | Align existing Stage 0 skeleton with frozen OpenAPI, CI and repository standards | Planned |
| 4 — Infrastructure | PostgreSQL schemas, migrations, Redis boundaries, outbox/inbox and test infrastructure | Planned |
| 5 — First Vertical Slice | Add transaction → PostgreSQL → snapshot → API → React dashboard with tests | Planned |

The repository already exists because Stage 0 was executed before the refined roadmap. Stage 3 therefore hardens rather than recreates it.

No AI, Tax Export, email, mobile, premium, or unnecessary worker implementation enters these stages.
