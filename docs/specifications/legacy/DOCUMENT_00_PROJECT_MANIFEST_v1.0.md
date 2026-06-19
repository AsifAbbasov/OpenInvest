# Document 00 — Project Manifest

| Field | Value |
| --- | --- |
| Document ID | 00 |
| Version | 1.0.0 |
| Status | Approved legacy; consolidated edition |
| Owner | Principal Architect |
| Supersedes | Cancelled OpenInvest OS draft originally attached to the project |
| Dependencies | Documents 42–43 |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |

## Mission

OpenInvest is an independent investment analytics platform. It is not a broker, asset manager, investment adviser, bank, or automated trading system. It gives users transparent analysis of their own capital, dividends, taxes, purchasing power, and long-term investment effectiveness.

## Principles

- Privacy First: collect the minimum personal data; passport, INN, address, and phone are optional.
- Trust First: every figure has a source and every calculation is reproducible and auditable.
- API First: business logic runs on the server and all clients use one contract.
- Performance First: use appropriate RAM/HTTP caching, ETags, compression, background jobs, and snapshots.
- User First: every screen answers a concrete user question.

## Differentiators

OpenInvest reports nominal, after-fee, after-tax, and inflation-adjusted returns; shows purchasing-power equivalents; exposes formula/source/date/rate/fee/tax details; and keeps a human in the loop for sensitive operations.

## Trust, security, and engineering

Design uses minimal collection, encryption, separation of identity and investment data, export, deletion, Zero Trust, least privilege, defense in depth, audit logs, immutable financial history, and versioned calculations. Engineering follows SOLID where useful, KISS, DRY, YAGNI, explicitness, separation of concerns, and composition over inheritance.

## Product questions

Within seconds, a user should understand capital value, real profit, received and expected dividends, applicable taxes, and purchasing power. Features that do not support those questions are secondary.

## Source-of-truth rule

Code cannot override approved documentation. Architecture changes require an approved ADR and the update process established by Document 43.
