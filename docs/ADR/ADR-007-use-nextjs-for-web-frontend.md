# ADR-007: Use Next.js App Router for the Web Frontend

| Field | Value |
| --- | --- |
| Document ID | ADR-007 |
| Version | 1.0.0 |
| Status | Accepted |
| Owner | Principal Architect |
| Supersedes | Vite SPA and mandatory Redux Toolkit portions of the Web target in `ARCHITECTURE_FREEZE_v1.md`; conflicting legacy Web frontend specifications |
| Dependencies | Documents 42–43; ADR-003; ADR-005 |
| Last Review Date | 2026-06-20 |
| Next Review Date | 2026-12-20 |

## Context

OpenInvest needs SEO-friendly public pages, static legal/marketing content, an authenticated web
shell, and deliberate server/client rendering boundaries. The original React + Vite SPA target can
provide the UI but does not make these web delivery capabilities first-class.

This change must not weaken the established API First boundary. Go remains the canonical backend
and the only business API. Python remains a future analytics and collector worker layer. Native
mobile clients remain future phases and will consume the same Go API.

## Decision

The Web MVP uses **Next.js App Router + TypeScript + pnpm** in `frontend-next/`.

React and TypeScript remain part of the Web stack through Next.js. Vite is replaced. Redux Toolkit
is no longer mandatory: this presentation-only skeleton needs no client state library, and any
future state dependency must be justified by an approved feature rather than installed preemptively.

Next.js is allowed only as the Presentation Layer. Its responsibilities are:

- UI rendering and web routing;
- SSR, SSG, and ISR where useful;
- Server Components for presentation and Go API data-fetch orchestration only;
- explicit hydration boundaries;
- SEO-friendly public pages;
- static marketing and legal pages;
- the authenticated web shell.

Next.js is strictly forbidden from:

- implementing business or domain logic;
- accessing PostgreSQL, Redis, or any database directly;
- calculating portfolios, financial results, dividends, inflation, purchasing power, or tax;
- integrating directly with MOEX, CBR, Rosstat, brokers, or other external data providers;
- acting as the main backend or replacing the Go API;
- storing business data in LocalStorage;
- creating Route Handlers/API routes for business domains.

Next.js Route Handlers are allowed only if a later approved need exists for frontend-only technical
concerns, such as a presentation health endpoint or proxy-free static metadata. No Route Handler is
created in the current scope. Any future Route Handler must remain outside business domains and be
reviewed against this ADR.

The architecture boundary is:

```text
Browser / Next.js Web
        ↓ OpenAPI contract
Go API — canonical backend and business API
        ↓
PostgreSQL / Redis / future Python analytics and collector workers
```

Future iOS SwiftUI and Android Jetpack Compose applications will consume the same Go API. No mobile
code is authorized now.

The target source layout uses one App Router root, not duplicate `app/` and `src/app/` trees:

```text
frontend-next/
├── src/
│   ├── app/
│   ├── common/
│   └── features/
├── public/
├── package.json
├── pnpm-lock.yaml
├── next.config.ts
└── tsconfig.json
```

## Alternatives considered

### Keep React + Vite SPA

Rejected for the current Web MVP target. It keeps the client simple but requires additional design
for SSR/SSG/ISR, public SEO pages, and server/client rendering boundaries.

### Use Next.js as a backend-for-frontend or primary backend

Rejected. It would duplicate Go responsibilities, create two business APIs, weaken OpenAPI First,
and increase security, correctness, and operational cost.

### Implement mobile clients now

Rejected as scope expansion. SwiftUI and Jetpack Compose remain future phases.

## Consequences

### Positive

- Public, marketing, and legal pages can use framework-native rendering and metadata capabilities.
- Server and Client Component boundaries can minimize unnecessary browser JavaScript.
- Web routing and authenticated shell composition use one supported App Router model.
- The Go API remains reusable by web and future native clients.

### Costs and constraints

- Next.js adds a Node.js web runtime and framework upgrade surface.
- Server Components must not become a hidden business/service layer.
- Dependency, caching, rendering, and deployment behavior require framework-specific tests.
- Route Handler creation requires explicit review against the presentation-only boundary.

## Dependency and license review

| Dependency | Pinned version | License | Purpose |
| --- | --- | --- | --- |
| Next.js | 16.2.9 | MIT | App Router Web presentation runtime |
| React / React DOM | 19.2.7 | MIT | Declarative UI rendering |
| TypeScript | 5.9.3 | Apache-2.0 | Static type checking |
| `@types/*` | pinned | MIT | Development-only type declarations |
| sharp | 0.34.5 transitive | Apache-2.0 | Next.js image/build support; only allowed dependency build script |

No database, external-provider, state-management, financial, or analytics dependency is added.
pnpm `allowBuilds` permits only sharp's installation check; all other dependency build scripts
remain denied by default.

## Security and privacy impact

- Server-rendered presentation must not log access tokens, cookies, identity documents, or private
  financial payloads.
- Browser business data must not be persisted in LocalStorage.
- Authentication remains governed by the Go API OpenAPI contract and approved cookie/token model.
- Next.js receives no database credentials or external-provider credentials.

## Performance and cost impact

SSR/SSG/ISR may improve delivery and SEO when applied deliberately, but server rendering consumes
Node.js compute. Static generation is preferred for stable public content. Authenticated financial
views call the Go API and must respect its caching/privacy rules. Deployment cost and Web Vitals are
measured before production; Next.js does not change backend SLO definitions.

## Compatibility and rollback

The current Vite application is only a non-business skeleton. Rollback restores `frontend-react/`
and the previous CI working directory. No API, database, financial history, or user data migration
is involved.

## Implementation guardrails

- No business screens or calculations are added during this architecture update.
- No business Route Handlers, database clients, external-provider SDKs, or server actions are added.
- All future data access is generated or implemented against the Go OpenAPI contract.
- pnpm is the only JavaScript package manager for the web project.
- Stage 3 remains unauthorized.
