# MVP Product Risk Refinement

| Field | Value |
| --- | --- |
| Document ID | PROD-RISK-001 |
| Version | 1.0.0 |
| Status | Approved / merged into `develop` |
| Owner | Principal Architect |
| Supersedes | Unstructured PRD criticism outside the repository |
| Dependencies | `SOURCE_OF_TRUTH.md`; Documents 42–43; ADR-003; ADR-006; ADR-007 |
| Last Review Date | 2026-06-27 |
| Next Review Date | Before public MVP scope lock |

## Purpose

This document converts hard product criticism into controlled MVP risk management. It does not
change Architecture Freeze v1.2 and does not authorize new implementation by itself.

The goal is to strengthen OpenInvest's advantages while reducing the most credible delivery,
adoption, legal, and UX risks.

## Sharpened product thesis

OpenInvest is a **Personal Capital Operating System** for private investors who need a reliable,
independent view of their own capital.

The first version must not try to serve every retail investor. The first version targets users whose
pain is already strong enough to justify setup effort:

- long-term investors with multiple brokers or accounts;
- dividend and FIRE-oriented investors;
- investors who care about real return after fees, taxes, and inflation;
- users with enough transactions or capital that broker-app summaries and spreadsheets are no
  longer trustworthy;
- users who value privacy, auditability, export, and deterministic calculations.

The first version is not optimized for casual users who only need a broker app's green/red return
badge.

## MVP value hierarchy

The first-release product value order is:

1. Correct capital ledger.
2. Explainable portfolio summary.
3. Real return after fees and inflation.
4. Dividend/coupon visibility.
5. Import and reconciliation to reduce manual entry.
6. Purchasing-power insight as a secondary explanatory layer.
7. Tax preparation data only after deterministic rules and privacy boundaries are ready.

Features that do not support this hierarchy move to `BACKLOG_V2.md`.

## Accepted criticism and mitigation

### Import cannot remain late

Risk: manual entry is acceptable for engineering validation, but a public MVP that requires users to
type hundreds of historical transactions will fail adoption.

Decision:

- keep Stage 3 manual append for the vertical slice;
- add a near-term **Broker File Import and Reconciliation** candidate before public MVP;
- start with file-based import, not broker API integrations:
  - CSV/XLSX first where possible;
  - PDF only after parser evidence exists;
  - no credential scraping;
  - no direct broker API unless licensed and registered.

Import must follow the existing append-only reconciliation model:

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

### Tax must be deterministic, not LLM-driven

Risk: presenting AI as a tax calculator creates correctness, legal, and user-trust risk.

Decision:

- the tax calculation core, if enabled later, must be a deterministic rules engine with test vectors;
- LLM/AI may only explain, summarize, assist review, or guide user input;
- AI must never be the source of tax truth;
- Tax XML/PDF export remains outside MVP and behind a future feature flag;
- any tax feature requiring personal data must support temporary in-memory entry or encrypted
  opt-in storage according to Privacy by Design.

### ICP must be narrower

Risk: a generic "all investors" target produces weak messaging and weak product decisions.

Decision:

- public MVP messaging must target investors with portfolio-accounting pain, not casual brokerage
  app users;
- onboarding and examples should emphasize multi-account tracking, real return, dividends, and
  auditability.

### Server-side calculations need snapshots, not request-time heroics

Risk: calculating all metrics from raw transactions on every request will violate SLOs as portfolios
grow.

Decision:

- raw immutable ledger remains canonical;
- expensive analytics must be materialized through rebuildable snapshots, background jobs, or
  controlled caches;
- synchronous API responses should read precomputed or bounded-scope data where possible;
- no new analytics feature may bypass the snapshot/caching strategy without ADR review.

## Partially accepted criticism

### Purchasing power can become noise

Risk: "iPhone counts" and similar equivalents can feel unserious if placed above core financial
truth.

Decision:

- keep Purchasing Power because it is a differentiator and explains inflation in human terms;
- demote it from core dashboard truth to a secondary insight card;
- always show real/inflation-adjusted return before consumer-good equivalents;
- support sober categories first:
  - average salary;
  - food basket;
  - rent;
  - utilities;
  - square meter / housing affordability where reliable data exists;
- consumer-electronics equivalents are optional examples, not primary financial analysis.

### Public MVP should not depend on broad broker integrations

Risk: full broker API synchronization is expensive, legally sensitive, and source-dependent.

Decision:

- broker import moves earlier as a product requirement candidate;
- full multi-broker API synchronization remains outside early implementation;
- the first practical step is user-supplied file import with explicit review and no credentials.

## Rejected or already mitigated criticism

### MongoDB / polyglot persistence

Not applicable to the frozen architecture. OpenInvest uses PostgreSQL as the canonical database.
Additional datastores require ADR approval.

### "All server calculations make SLO impossible"

Accepted as a risk, rejected as a conclusion. The current architecture already uses immutable
transactions, rebuildable snapshots, Redis/RAM cache, and background-work boundaries to avoid
unbounded request-time calculations.

### "The product must become a broker-adjacent upsell to survive"

Rejected for MVP. Independence is a core trust property. Monetization and pricing must be tested
without compromising data ownership, calculation transparency, or product neutrality.

## Product guardrails

- Do not sell the product as a trading terminal.
- Do not sell AI as financial, investment, or tax truth.
- Do not make Purchasing Power the primary dashboard metric.
- Do not require passport, INN, phone, address, or stored tax profile for portfolio analytics.
- Do not add external provider data without the Data Source Registry.
- Do not add import behavior that mutates history automatically.
- Do not add a feature unless it improves capital, return, dividends, taxes-as-data, or purchasing
  power understanding.

## Near-term roadmap adjustment

The implementation order remains architecture-first, but public-MVP readiness requires import
earlier than the legacy roadmap suggested.

Recommended sequence after the first vertical slice:

1. Stage 3.3 — Next.js presentation slice for the current Go API.
2. Stage 3.4 — End-to-end verification and onboarding.
3. Stage 3.5 — Broker file import and reconciliation design.
4. Stage 3.6 — File-import vertical slice, if approved after design review.
5. Later — WAC/XIRR/real-return algorithms with financial test vectors.

This is a product-risk recommendation, not automatic implementation authorization.

## Success criteria for public MVP

Public MVP should not be declared ready until:

- a user can create/import enough ledger data without unreasonable manual effort;
- every visible number has an explanation path;
- core financial calculations have deterministic test vectors;
- Purchasing Power is clearly secondary to real return;
- tax functionality is either absent or deterministic and privacy-scoped;
- performance relies on snapshots/caches rather than full recalculation on every request.
