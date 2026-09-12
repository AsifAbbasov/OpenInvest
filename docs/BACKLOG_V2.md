# Backlog v2.0

| Field | Value |
| --- | --- |
| Document ID | PROD-BL-002 |
| Version | 2.0.2 |
| Status | Frozen until after MVP |
| Owner | Principal Architect |
| Supersedes | Feature ideas mixed into MVP scope |
| Dependencies | Documents 42–43 |
| Last Review Date | 2026-09-11 |
| Next Review Date | After MVP release |

The following ideas cannot alter MVP implementation without an approved architecture/product update: AI Assistant, scenario analysis, premium analytics, Tax XML export, email automation, forecasting, family accounts, public API, foreign securities/markets, iOS, Android, and desktop applications.

Adding an item here is not approval to implement it.

## Clarified backlog boundaries

- Full multi-broker API synchronization remains backlog until source licenses, security model,
  rate limits, and user consent are approved.
- Credential scraping is not an approved strategy.
- LLM-driven tax calculation is rejected; future AI may explain deterministic tax outputs only.
- Purchasing Power entertainment-style equivalents are backlog unless product review confirms they
  improve understanding without distracting from real return.

## Historical MVP-readiness adjustment

Broker file import and reconciliation was moved out of the late generic backlog into public-MVP
readiness scope. The subsequent design and implementation stages are complete; current lifecycle
status is recorded in [ROADMAP.md](ROADMAP.md), and implementation chronology is recorded in
[IMPLEMENTATION_LOG.md](IMPLEMENTATION_LOG.md). This section preserves the product decision that
import moved earlier; it is not a current implementation gate.
