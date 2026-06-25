# Documentation Changelog

| Field | Value |
| --- | --- |
| Document ID | REG-CHG-001 |
| Version | 1.0.3 |
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

## 2026-06-25 — Stage 2 final review blockers

- Required explicit reversal `effectiveDate` BusinessDate so immutable-ledger reversals and
  snapshot rebuilds do not depend on system timestamps.
- Changed economically non-negative aggregate values from signed `Money` to `NonNegativeMoney`.
- Tightened `traceparent` validation to reject W3C-invalid version `ff`, all-zero trace IDs, and
  all-zero parent IDs.
- Documented repository OpenAPI validator limitations and added focused mutation guards instead of
  claiming complete JSON Schema 2020-12 compliance.
- Recorded auditable Principal Architect approval for the Stage 2 26-file review-size exception.
- Documented the GitHub runner Ruby version as a remaining non-blocking operational hardening risk.

## 2026-06-25 — Stage 2 closure and ADR-006 acceptance

- Squash-merged PR #2 into `develop` at `bfde623552ebea6eac7bdaabf0d1a2263883de12`.
- Accepted ADR-006 after external review, green CI, and human approval.
- Marked Stage 2 Contract and Canonical Model Freeze as closed.
- Declared `develop` at the Stage 2 merge commit as the canonical Stage 2 baseline.
