# OpenInvest documentation standard — ADR · MADR · Diátaxis · Docs-as-Code
**Purpose:** Maximize compression of EXISTING documentation; one live authority per contract, no duplicate reports or needless new Markdown.
**Diátaxis / tutorial:** [README](../../README.md) keeps working onboarding steps, prerequisites and architecture diagrams.
**Diátaxis / how-to:** Operations, infrastructure and scripts retain complete procedures, security prerequisites and rollback.
**Diátaxis / reference:** OpenAPI, schemas, approved contracts and registries stay complete; past Git snapshots cannot replace a live normative rule.
**Diátaxis / explanation:** ADRs and compact decisions explain context, root cause, problem, considered alternatives, choice and consequences.
**Nygard + MADR:** Context → Decision → Consequences; distinguish evidenced alternatives from suggestions, explain rationale, trade-offs and residuals.
**Class A / compact decision:** Target ≤10 physical lines; MAX 15 including title. Include Status/Version/Owner, Context, Root Cause, Problem, Risk, Alternatives, Decision, Rationale, Trade-offs, Implementation, Verification, Residuals; combine fields rather than omit facts.
**Class B / normative or mixed:** No forced cap on APIs, database rules, operational guides, registers, onboarding or navigation; open with a ≤15-line overview when useful, eliminate duplicate narrative.
**Class C / evidence:** Preserve hash-pinned bytes, historical RED/FAIL and dated decisions; ongoing research may add dated evidence, not rewrite previous findings.
**Docs-as-Code:** Edit existing files on a branch in a Draft PR; check syntax, links/fragments, exact SHAs, byte/line deltas, normative coverage and applicable exact-head CI before owner-controlled merge.
**Authority:** Documents 43 → 42 → accepted ADRs → applicable legacy; [Source of Truth](../SOURCE_OF_TRUTH.md), [Document Index](../DOCUMENT_INDEX.md) and [Data Source Registry](../registries/DATA_SOURCE_REGISTRY.md) retain distinct authority.
**Governance:** Technical PASS ≠ formal closure ≠ deployment approval; accepted ADRs stay immutable, proposals are not approvals; never fabricate historic rationale or reclassify RED as PASS.
**Template:** Use the existing [ADR / MADR template](../ADR/README.md); full tests/contracts stay in their current authoritative files, not a new duplicated summary.
**Preservation:** B/C classification never automatically exempts removable narrative from review, but no shortening may remove operational obligations, rights, approval boundaries or immutable evidence.
