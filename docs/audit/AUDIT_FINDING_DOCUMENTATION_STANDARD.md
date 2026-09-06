# Audit Finding Documentation Standard

| Field | Value |
| --- | --- |
| Status | Mandatory for all repository-audit remediation work |
| Scope | Original audit findings P0/P1/P2/P3 and any blocking regressions discovered during remediation review |
| Applies from | Stage 3.34 onward, and retroactively as a consistency target for already-closed remediation stages |

## Purpose

Every audit finding must leave behind a durable engineering record that explains not only what changed, but why the defect existed, how it could affect OpenInvest, why the chosen remediation is appropriate, what alternatives were rejected, how the fix is proven, and what residual risk remains.

README is a high-level index only. Detailed evidence belongs in `docs/stages/` and/or `docs/audit/`. `docs/audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` is the cross-finding index.

## Mandatory 18-part finding record

A finding may not be declared canonically CLOSED unless its detailed remediation documentation covers all applicable sections below. If a section is not applicable, the document must say why rather than silently omit it.

1. **Finding / symptom** — exact review finding and observable defect.
2. **Root cause** — technical reason the defect existed.
3. **Failure scenario** — concrete reproduction or adversarial path.
4. **Impact** — effect on financial correctness, consistency, security, privacy, availability, performance, maintainability, or user experience.
5. **Severity rationale** — why the finding is P0/P1/P2/P3.
6. **Existing guarantees violated** — architecture, financial, API, security, privacy, database, or governance contract that was not being upheld.
7. **Considered solutions** — credible remediation options evaluated.
8. **Chosen remediation** — exact implementation/design change selected.
9. **Why this solution** — decision rationale tied to OpenInvest constraints and canonical architecture.
10. **Rejected alternatives** — alternatives not selected and why.
11. **Trade-offs** — cost, complexity, operational impact, performance impact, compatibility, or maintenance burden introduced by the fix.
12. **Regression tests** — tests or checks that prevent recurrence and what failure mode each proves.
13. **Adversarial review findings** — blockers or weaknesses found by independent review after the first implementation.
14. **Remediation iterations** — first fix, later corrections, and why earlier attempts were insufficient.
15. **Residual risk / limitations** — remaining non-blocking risk, deferred cleanup, assumptions, or out-of-scope paths.
16. **Operational / deployment consequences** — rollout, credentials, migrations, runtime configuration, provider requirements, or manual controls, where applicable.
17. **Exact evidence** — branch, PR, reviewed head SHA, CI run, review verdict, merge SHA, and closure PR/commit where applicable.
18. **Final canonical status** — OPEN/CLOSED and the exact canonical evidence that makes the status authoritative.

## Review-discovered blockers

A blocker discovered while remediating an original finding must not be erased from history after it is fixed. The stage document must preserve:

- the original proposed remediation;
- the independent review objection;
- why the objection was valid;
- the minimum corrective action requested;
- the actual corrective action implemented;
- new regression/adversarial tests;
- exact-head CI and repeat review evidence.

This applies even when the blocker is ultimately considered part of the same original finding rather than assigned a new audit ID.

## Documentation layers

### README

README should provide only a concise remediation summary: affected findings, broad remediation, implementation/closure status, and a link to the detailed dossier. It must not become the authoritative forensic record.

### Stage dossier

The stage dossier in `docs/stages/` is the primary detailed engineering narrative for a remediation slice and must satisfy the 18-part standard for all findings in that slice.

### Audit remediation register

`docs/audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` is the single cross-finding index. It records status, stage, high-level root cause/impact, chosen remediation, evidence location, and canonical closure state for all 32 original findings.

### Source of Truth / Roadmap / Implementation Log

These remain governance/state documents. They should reference canonical stage and closure evidence without duplicating the full forensic narrative.

## Closure rule

A code change alone is not sufficient to close an audit finding. Canonical closure requires:

1. implementation within approved scope;
2. regression/verification evidence;
3. exact-head green CI;
4. independent review approval;
5. explicit human squash-merge authorization where required by `docs/REVIEW_WORKFLOW.md`;
6. squash merge to canonical `develop`;
7. closure governance synchronized across the relevant documentation surfaces;
8. remediation register status updated with canonical evidence.

If any required enforcement or evidence is unavailable, the finding remains OPEN. Documentation must never convert an unavailable control into a claimed closed control.