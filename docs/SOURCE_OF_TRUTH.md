# OpenInvest Source of Truth

| Field | Value |
| --- | --- |
| Document ID | SOT-001 |
| Version | 1.4.31 |
| Status | Approved / Architecture Freeze Active |
| Owner | Principal Architect |
| Supersedes | Disconnected source-of-truth declarations in legacy documents |
| Dependencies | Documents 42–43 and accepted ADRs |
| Last Review Date | 2026-07-12 |
| Next Review Date | Before Stage 3.13 implementation review |

## Architecture status

**Architecture Freeze v1.2: ACTIVE**
**Documentation Freeze: ACTIVE**
**Last completed implementation stage: Stage 3.12 — Web Authentication UI Slice**
**Last completed architecture amendment: Next.js Web Presentation Amendment**
**Current canonical implementation baseline: `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92`**
**Current active work item: Stage 3.13 instrument catalog implementation on `feature/stage-03-13-instrument-catalog`**
**Stage 2 status: Closed / merged into `develop`; ADR-006 accepted**
**Web presentation amendment status: Closed / merged into `develop`; ADR-007 accepted**

## Document priority

`Document 43 → Document 42 → accepted ADRs → Documents 1–41 → README → comments`

No lower-priority document, code, or comment may override a higher-priority decision. An architecture change requires an ADR, review, approval, and update to this file.

## Product definition

OpenInvest is a **Personal Capital Operating System**. It is not a broker, bank, trading terminal, investment adviser, or tax service.

The first public MVP targets investors with real portfolio-accounting pain: long-term, dividend,
FIRE, and multi-account investors who need independent, explainable real-return analytics. It is not
optimized for casual brokerage-app users who only need a simple green/red return badge.

## Frozen stack and architecture

- Monorepo: `OpenInvest/`
- Backend: Go 1.24+ and Fiber
- Analytics: Python managed with uv
- Database: PostgreSQL; schemas `identity`, `investment`, `analytics`, `tax`, `audit`
- Cache: Redis plus process-local RAM cache
- Web frontend: Next.js App Router, TypeScript, and pnpm; presentation layer only under ADR-007
- Current client implementation scope: Web MVP only
- Mobile future: iOS SwiftUI and Android Jetpack Compose; no current mobile implementation
- Style: API First, DDD, Clean Architecture, Event Driven
- Data: canonical database, immutable transactions, rebuildable versioned snapshots
- Events: at-least-once delivery and idempotent business processing through outbox/inbox
- Security: Zero Trust and Privacy by Design
- External sources: official/permitted sources registered before use and accessed only through backend collectors
- Client: no external market-data calls and no LocalStorage for business data
- Boundary: Browser/Next.js → OpenAPI-defined Go API → PostgreSQL/Redis/future Python workers;
  Next.js never replaces the Go business API or accesses data stores directly
- Delivery: no automatic commit or push without user review and approval
- Review: mandatory feature branch → local checks → read-only Internal Review Agent line-by-line
  approval → human commit/push permission → Draft PR → green CI → independent ChatGPT external
  approval → human approval → squash merge; see `REVIEW_WORKFLOW.md`.

## Mandatory quality gates

- Builder Agent cannot approve its own work.
- Every changed line is reviewed internally before commit permission is requested.
- Internal Review Agent produces findings only and cannot edit, stage, commit, or push.
- Draft PR cannot merge without green CI, approved internal review evidence, approved independent
  ChatGPT Draft PR review, and explicit human approval.
- External ChatGPT receives the Draft PR diff without prior internal verdict disclosure and reaches
  an independent conclusion before review evidence is compared.
- Every fifth completed stage requires a full repository line-by-line audit before proceeding.
- A review gate may add evidence and reject scope; it cannot silently change frozen architecture.

## MVP scope

Included: registration; portfolio; transactions; stock card; bond card; dividend calculator; dividend calendar; snapshots; weighted-average cost; XIRR; real return; inflation-adjusted return; purchasing-power card; dashboard.

Excluded: AI Assistant; scenarios; premium analytics; Tax XML export; email automation; forecasts; family accounts; public API; foreign securities; mobile applications.

Public-MVP readiness additionally requires an approved import/reconciliation path so users are not
forced to enter large transaction histories manually. The preferred first import path is user-supplied
broker files with explicit review, not credential scraping or direct broker API synchronization.

Purchasing Power remains an MVP differentiator, but it is a secondary explanatory insight. Real
return, capital, dividends/coupons, and inflation-adjusted performance stay above consumer-good
equivalents in dashboard priority.

Tax export remains outside MVP. Any future tax calculation core must be deterministic and covered by
financial/legal test vectors; AI may explain or assist review, but must never be the source of tax
truth.

Product risk refinement is closed and merged into `develop` at
`65bdf6537b44ed57e1c00bf68d2dacd70aa09702`. Stage 3.3 is closed and merged into `develop` at
`11805cc298bba13f09f7f7af8b1e1178dc351209`, with closure documentation merged at
`fe402030359459f909c156a1e993f18ceed257bf`. Stage 3.4 is closed and merged into `develop` at
`86582efaa420b2c38465a5d0da041814149392c7`; it added end-to-end local verification and root
developer commands only. Stage 3.5 is closed and merged into `develop` at
`072d38d94b529221d6467502f82f03a674a7d805`; it approved the design guardrails for user-supplied
broker-file import. Stage 3.6 implementation is closed and merged into `develop` at
`e2b05650a4422b97d4bd924254367106b6a4686b`, with closure governance merged at
`fb651632036fabaa31ec92e9d28b5782ca0f92e5`; it added an internal CSV parse/normalize/review/
append-plan slice only. It does not authorize public import endpoints, broker API integration,
upload UI, SQL migrations, workers, or automatic ledger append. Stage 3.7 import append planning is
merged into `develop` at `36d86c7ff2a9c75478de155d4f60b979b8da9376`. Stage 3.7 implementation is
closed and merged into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`; it added internal
atomic append of user-approved import rows with duplicate revalidation, idempotency protection,
minimal audit evidence, and deterministic snapshot rebuilds. Public import endpoints, upload UI,
import-session persistence, broker/provider integrations, workers, tax, mobile, and AI remain out of
scope.

Stage 3.8 import review append flow planning is merged into `develop` at
`a35af2f5207bd564647d2a3fc032f4f940e62ddd`. Stage 3.8 implementation is closed and merged into
`develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`; it added internal orchestration between
Stage 3.6 review output and Stage 3.7 atomic append. Public import endpoints, OpenAPI changes,
upload UI, SQL import-session persistence, raw file persistence, workers, broker/provider
integrations, tax, mobile, AI, and automatic append without explicit approved decisions remain out
of scope.

Stage 3.8 closure governance is merged into `develop` at
`cb9a392eb90ede954d9cc68b247bada13a1540d9`. Stage 3.9 import API boundary planning is merged into
`develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`. Stage 3.9 implementation is closed and
merged into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`, with closure governance
merged at `682ffd856395a6e3e988817551a512898fda2d38`; it added only the public Go API boundary,
OpenAPI contract, DTOs, tests, and documentation for user-supplied CSV import review/append.
Stage 3.10 import upload/review UI planning is closed and merged into `develop` at
`27480d6ff22e2929e33aeac352aef8a1b01bb448`. Stage 3.10 implementation is closed and merged into
`develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71`; it added only the presentation-only
Next.js import upload/review UI over the existing Go API boundary. SQL import-session persistence,
raw file persistence, backend contract changes, workers, broker/provider integrations, tax, mobile,
and AI remain out of scope. Stage 3.11 authentication and privacy-boundary planning is closed and
merged into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125`. Stage 3.11 implementation is
closed and merged into `develop` at `5c49173ac858995929f266c2de991282dd194dec`; it implemented only
the approved Go API auth, privacy-default, session, CSRF, audit, and persistence boundary. It does
not authorize frontend auth UI, email verification, OAuth/passkeys/2FA, tax, mobile, AI, workers,
provider integrations, or portfolio feature expansion. Unsafe local auth flags with a configured
`DATABASE_URL` require `OPENINVEST_ENV=development` or `local`; production authorization must use
signed access tokens. Stage 3.11 closure governance is merged into `develop` at
`2febb6f49224ec6252368d2195a4e3054ea24278`. Stage 3.12 Web authentication UI planning is closed
and merged into `develop` at `25be13ce84844562e0381b79f4b81cbfed7eb44d`. Stage 3.12 Web
authentication UI implementation is closed and merged into `develop` at
`b4840b60346109e3cd54a07d9e1e131fc0cfad23`. Stage 3.12 closure governance is merged into
`develop` at `321eaf4f75df83d85fd356a8d6a454e49bbc4db4`. Stage 3.13 instrument catalog planning is
merged into `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92`. Stage 3.13 instrument
catalog implementation is active on `feature/stage-03-13-instrument-catalog`; it is limited to backend-owned
approved asset-catalog fixture resolution, supported ticker validation, existing asset table usage,
snapshot stock/bond bucket classification from backend-owned asset type, tests, and documentation.
It does not authorize OpenAPI changes, SQL migrations, Go handler changes, frontend work, provider
integrations, workers, market-data ingestion, stock/bond cards, dividend/coupon scope, tax, mobile,
or AI.

## Financial standard

- Decimal only; binary float forbidden for financial values.
- Half-even rounding.
- Eight decimal places internally; two for monetary display.
- SQL `DATE` for financial business dates.
- UTC `TIMESTAMP WITH TIME ZONE` for system timestamps.
- MOEX calendar for MVP market events.
- Canonical financial test vectors are mandatory before production algorithms.

## Privacy definitions

- **Personal Data:** information that identifies a person directly or indirectly.
- **Pseudonymized Data:** data linked through a reversible identifier.
- **Anonymous Data:** data that cannot be linked back to an individual by any reasonable technical or organizational means.

Deleting a user removes identity data and irreversibly destroys its link to the financial ledger. The detached ledger becomes **Anonymous Financial History**. OpenInvest retains no re-identification mechanism.

## Retention

- Identity/personal data: deleted completely after the approved deletion lifecycle.
- Audit: 10 years.
- Anonymous transactions and snapshots: no fixed expiration.
- Backups: maximum 90 days, then automatic destruction.

## Feature matrix

| Capability | MVP | State |
| --- | --- | --- |
| Registration and privacy defaults | Yes | Stage 3.12 closed; Go API auth plus presentation-only Web auth UI |
| Portfolio and transactions | Yes | Stage 3.4 verification closed |
| MOEX shares and bonds | Yes | Stage 3.13 instrument catalog implementation active |
| Dashboard and snapshots | Yes | Stage 3.4 verification closed |
| WAC, XIRR, real/inflation returns | Yes | Planned |
| Dividend calculator/calendar | Yes | Planned |
| Broker file import and reconciliation | Public-MVP readiness candidate | Stage 3.10 upload/review UI slice closed; no import-session persistence |
| Purchasing power | Yes | Planned as secondary insight |
| Tax export | No | Experimental; feature flag off |
| Foreign securities | No | Backlog v2.0 |
| AI, mobile, premium, public API | No | Backlog v2.0 |

## ADR registry

| ADR | Decision | Status |
| --- | --- | --- |
| ADR-001 | Go and Fiber backend | Accepted |
| ADR-002 | PostgreSQL canonical database | Accepted |
| ADR-003 | OpenAPI-first contracts | Accepted |
| ADR-004 | Versioned rebuildable snapshots | Accepted |
| ADR-005 | Privacy by Design | Accepted; interpreted with Document 43 anonymization terminology |
| ADR-006 | Stage 2 MVP contract and canonical model freeze | Accepted |
| ADR-007 | Next.js App Router for Web presentation only | Accepted |

## Version matrix

The complete version and legacy-document matrix is maintained in `VERSION_MATRIX.md`. `DOCUMENT_INDEX.md` is the navigation registry.

## Open questions

No unresolved architecture questions exist at Freeze v1.2 activation. See `OPEN_QUESTIONS.md` for the controlled process.
