# MVP Product Risk Refinement

| Field | Value |
| --- | --- |
| Document ID | PROD-RISK-001 |
| Version | 1.0.1 |
| Status | Approved / merged into `develop` |
| Supersedes | Unstructured PRD criticism outside the repository |
| Dependencies | `SOURCE_OF_TRUTH.md`; Documents 42–43; ADR-003; ADR-006; ADR-007 |

## Context and decision

OpenInvest is an independent, privacy-first Personal Capital Operating System. This approved risk
refinement does not change Architecture Freeze v1.2 or authorize implementation.
Target long-term, multi-account, dividend/FIRE investors whose accounting needs exceed broker badges
and spreadsheets; emphasize real return after fees, taxes and inflation, auditability and export.
Onboarding/examples should highlight multi-account tracking, real return, dividends and auditability.

## Value hierarchy

1. Correct capital ledger.
2. Explainable portfolio summary.
3. Real return after fees and inflation.
4. Dividend/coupon visibility.
5. Import/reconciliation instead of repetitive manual entry.
6. Purchasing Power as a secondary explanatory layer.
7. Tax preparation only after deterministic rules and privacy boundaries are ready.

Other ideas belong in [Backlog](../BACKLOG_V2.md), not automatic MVP scope.

## Accepted risks and decisions

### Import and broker integration

Manual entry is suitable for vertical-slice validation, not hundreds of public-MVP transactions.
The approved recommendation moved file import/reconciliation earlier, while retaining manual append.
Prefer CSV/XLSX; PDF needs parser evidence. Credential scraping is forbidden; direct broker APIs
require licensing and registry approval. Broad synchronization remains later scope because its cost,
legal exposure and source dependence are unsuitable for the early MVP.
Use user-supplied files with explicit review and no credentials:

```text
Import
→ Normalization
→ Matching
→ Duplicate Detection
→ Conflict Detection
→ User Review
→ Append Only
```

No import may silently mutate historical records.

### Deterministic tax and privacy

LLM-driven tax truth risks incorrect results, legal exposure and lost trust. Future tax calculations
must use deterministic rules and test vectors; AI may explain, summarize, assist review or guide input,
never supply tax truth. Tax XML/PDF export remains outside MVP behind a future feature flag.
Personal tax input must support temporary in-memory entry or encrypted opt-in storage under Privacy by Design.

### Bounded server-side analytics

Full-ledger recalculation on every request threatens SLOs. Keep the immutable ledger canonical;
materialize expensive analytics through rebuildable snapshots, background jobs or controlled caches.
Read precomputed/bounded-scope data synchronously; bypassing this strategy requires ADR review.
This mitigates the risk without moving financial truth into clients.

### Purchasing Power — partially accepted criticism

Retain its explanatory value but place it below real/inflation-adjusted return, never as primary truth.
Prioritize salary, food basket, rent, utilities and housing/square-meter affordability where data are reliable.
Consumer electronics are optional examples, not primary analysis.

## Rejected alternatives and rationale

- PostgreSQL remains canonical; polyglot persistence/MongoDB requires an ADR.
- Server-side calculations do not inherently make SLOs impossible: use the bounded strategy above.
- A broker-adjacent upsell is rejected for MVP; monetization must preserve independence, data ownership,
  transparent calculations and neutrality.

## Product guardrails

- No trading-terminal positioning or AI presented as financial, investment or tax truth.
- No primary Purchasing Power dashboard metric.
- No passport, INN, phone, address or stored tax profile required for portfolio analytics.
- No external data without the [Data Source Registry](../registries/DATA_SOURCE_REGISTRY.md).
- No automatic history mutation; features must improve capital, return, dividends, taxes-as-data or purchasing-power understanding.

## Historical sequencing and verification

The recommendation was Stage 3.3 presentation → 3.4 end-to-end verification → 3.5 import design →
3.6 file-import slice if separately approved → later WAC/XIRR/real-return algorithms with test vectors.
Stages 3.3–3.6 subsequently completed through separate reviews; [Roadmap](../ROADMAP.md) owns current
sequencing and [Implementation Log](../IMPLEMENTATION_LOG.md) the evidence. This was not runtime authority.

## Public-MVP acceptance and residuals

Do not declare readiness until users can create/import sufficient data without unreasonable effort,
every visible number is explainable, core financial calculations have deterministic vectors,
Purchasing Power is secondary, tax is absent or deterministic/privacy-scoped, and performance uses
snapshots/caches rather than full recalculation per request. Licensing and future tax approvals remain separate.
