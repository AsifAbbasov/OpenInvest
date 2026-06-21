# Documentation Changelog

| Field | Value |
| --- | --- |
| Document ID | REG-CHG-001 |
| Version | 1.0.1 |
| Status | Active |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | `SOURCE_OF_TRUTH.md` |
| Last Review Date | 2026-06-21 |
| Next Review Date | 2026-12-21 |

## 2026-06-19 — Architecture Freeze v1.2

- Approved Documents 42 and 43 as the two highest-priority architecture sources.
- Resolved business-date versus UTC timestamp semantics.
- Replaced exactly-once transport language with at-least-once delivery and idempotent business processing.
- Corrected privacy terminology from pseudonymization to anonymization when re-identification is impossible.
- Froze MVP scope, asset scope, financial precision, retention, SLO boundaries, data schemas, and document precedence.
- Consolidated source documents into the repository and activated Documentation Freeze.
- Established mandatory Builder/CI/Review Agent/Human separation, Draft PR review gates, PR size budgets, ADR triggers, branch conventions, and squash-merge policy.

## 2026-06-21 — Stage 2 governance hardening

- Registered proposed ADR-006 and all Stage 2 contract artifacts without approving the ADR.
- Added the repository-owned OpenAPI validator to the pull-request CI gate.
- Reserved explicit `EXAMPLE_*` source identifiers so contract examples cannot be mistaken for
  approved MOEX, Rosstat, CBR, or other production sources.
- Synchronized the Stage 2 status across governance registries and the implementation log.
