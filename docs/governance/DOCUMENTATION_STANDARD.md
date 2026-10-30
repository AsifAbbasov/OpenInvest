# OpenInvest documentation standard — ADR, MADR, Diátaxis, Docs-as-Code
**Goal:** Compress EXISTING documents; no duplicate truths or reports; new documentation files MUST stay within 15 physical lines total (including blanks/title).
**Diátaxis:** Tutorials = README/onboarding; how-to = operations/scripts; reference = OpenAPI/DB/registries; explanation = ADR/decisions. Keep task-oriented navigation.
**ADR / Nygard:** Context → Decision → Consequences; record why, not just what.
**MADR:** Document problem/root cause, risk, genuinely considered alternatives, rationale, trade-offs and residuals; never invent deliberations.
**A / Compact decisions:** Target 10, MAX 15 physical lines incl title; include status/owner, context, problem, cause, options, decision, rationale, consequences, PR, exact-head verification and remaining risks. Combine labels, not facts.
**B / Existing documents:** Add/change at most 15 physical lines per task, including API, DB, runbooks, registers and onboarding; preserve all live requirements, contracts and evidence. If a lossless edit needs more, stop for owner's separate decision.
**C / Evidence:** Preserve frozen bytes/hashes, original RED/FAIL and dated history; new provider research is recorded prospectively, not rewritten retroactively.
**Authority:** Documents 43 → 42 → accepted ADRs → applicable legacy; [Source of Truth](../SOURCE_OF_TRUTH.md), OpenAPI and [Data Source Registry](../registries/DATA_SOURCE_REGISTRY.md) keep distinct roles.
**Docs-as-Code:** Edit existing files on a branch/Draft PR; check links/fragments, SHA, version, line/byte deltas, protected blocks and exact-head CI; owner controls merge/deploy.
**Governance:** Technical PASS ≠ formal closure ≠ deploy; proposed ≠ approved; history-only Git links cannot replace live contracts; B/C classification does not exempt removable narrative.
**Template:** Existing [ADR/MADR template](../ADR/README.md); full tests, contracts and operations remain in current authoritative files.
