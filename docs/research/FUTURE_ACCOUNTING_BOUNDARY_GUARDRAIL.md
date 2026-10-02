# Future Accounting Boundary Guardrail

| Field | Value |
| --- | --- |
| Status | Proposal / non-authorizing |
| Scope | Future architecture guardrail only |
| Runtime impact | None |
| Product impact | None |
| Current architecture change | No |
| Requires separate ADR before implementation | Yes |

## Purpose

This document preserves one architectural conclusion from external donor/repository analysis without changing the current OpenInvest architecture or roadmap.

OpenInvest is an investment portfolio accounting and analytics product. It is not a bank, payment processor, wallet, custody platform, clearing system, or payment-rail implementation.

The current immutable investment transaction ledger is designed around investment events, portfolio reconstruction, cost basis, positions, corrections/reversals, cash-flow/income analytics, valuation, returns, broker-import reconciliation, and related investment-domain truth.

## Guardrail

**Do not convert the existing OpenInvest investment ledger into a generic banking or payment double-entry ledger merely because double-entry accounting is common in banking, wallet, payment, clearing, or settlement systems.**

A future double-entry accounting subsystem may be considered only when a separately reviewed product requirement proves that OpenInvest has a bounded context that genuinely needs accounting semantics such as debits/credits, chart-of-accounts treatment, cash-account balancing, settlement, custody, payable/receivable accounting, or another domain where double-entry invariants materially improve correctness.

If such a need appears, the preferred default is:

1. keep the existing investment-event ledger canonical for its current bounded context;
2. introduce the accounting capability as a separate bounded context or narrowly scoped subsystem;
3. define an explicit mapping between investment events and accounting postings rather than silently replacing the current ledger model;
4. preserve append-only correction/reversal semantics and exact-money rules;
5. require a dedicated ADR, threat model, migration/replay analysis, reconciliation design, and executable acceptance tests before implementation;
6. reject any proposal whose primary justification is architectural fashion, donor similarity, or the phrase “bank-grade”.

## Why this boundary exists

External reviews of double-entry systems such as Blnk, pgledger, clearing, Atlas, bank0 and other payment/ledger projects demonstrate strong techniques that OpenInvest can reuse selectively: database-level invariants, deterministic lock ordering, reconciliation, immutability, property testing, chaos testing, tamper evidence, and least-privilege database boundaries.

Those projects nevertheless solve materially different problems: custody, payment movement, settlement, wallet balances, clearing, or general accounting. Their domain model should not be imported wholesale into OpenInvest.

The correct reuse strategy is therefore:

> **Transfer invariants and engineering methods, not an unrelated financial domain model.**

## Examples of reusable techniques that do not require ledger replacement

- database-enforced immutability and least privilege;
- deterministic lock ordering where multiple protected rows must be acquired;
- idempotency and replay protection;
- reconciliation between canonical and derived state;
- property-based and concurrency tests;
- PREVENTED vs DETECTED invariant classification;
- controlled fault/chaos testing;
- tamper-evident audit research with an independent trust boundary;
- strict separation of event time and ingestion time;
- explicit provider and source provenance.

## Future decision gate

Any future proposal to introduce double-entry accounting must answer, with evidence:

- What new product capability cannot be represented safely by the existing investment-event ledger?
- What exact bounded context owns the double-entry model?
- Does it coexist with, derive from, or replace any existing canonical truth?
- What migration and historical replay semantics are required?
- What new failure modes, reconciliation obligations, and security boundaries appear?
- What measurable correctness benefit justifies the added complexity?

Until those questions are answered through a separately reviewed ADR and implementation gate, the default decision is **NO ARCHITECTURAL CONVERSION**.

## Non-authorization statement

This proposal does not authorize Stage 3.78+, schema changes, Go changes, frontend changes, migrations, a new accounting engine, a new service, or any change to the current canonical investment ledger. It exists only to preserve a future architectural boundary discovered during donor analysis.
