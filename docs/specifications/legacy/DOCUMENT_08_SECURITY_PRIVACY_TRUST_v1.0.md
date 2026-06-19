# Document 08 — Security, Privacy & Trust by Design

| Field | Value |
| --- | --- |
| Document ID | 08 |
| Version | 1.0.0 |
| Status | Approved legacy; consolidated edition |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | Document 00; amended by Documents 30, 42, and 43 |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |

## Core rule

The user owns the data. If a feature can work without personal data, that data must not be collected. New accounts start with Privacy Mode on, tax profile off, notifications off, and anonymous analytics.

Required account data is limited to email, password hash, language, theme, and timezone. Identity and tax details remain optional. Temporary, encrypted stored, and manual/no-storage tax modes preserve human confirmation.

## Security baseline

- Argon2id password hashing, AES-256-GCM sensitive-field encryption, TLS 1.3 transport.
- Secrets in an approved secret manager/environment, never Git or logs.
- Separate identity, investment, tax, audit, and notification boundaries; Document 42 maps these to PostgreSQL schemas.
- Authenticate, authorize, validate, audit, then execute every request.
- Short-lived access tokens, rotated refresh tokens, rate limits, replay protection, prepared SQL statements, XSS/CSRF controls, and CORS allowlists.
- Private portfolio data never enters public CDN caches.
- Passwords, tokens, passport data, INN, XML/PDF contents, and sensitive tax data are forbidden in logs.

## User control

Users can inspect and terminate sessions, export data, and request deletion with an approved grace period. Generated tax files use expiring downloads and are not retained without explicit user choice. Document 43 governs anonymization and retention after identity deletion.

## Trust

Every portfolio number is explainable, every tax calculation traceable, and every dividend source-linked. AI cannot transact or modify a portfolio; final decisions belong to the user. Legal policies and third-party/open-source notices are required before production.
