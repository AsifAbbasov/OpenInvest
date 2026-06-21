# OpenInvest Architecture Freeze v1.0

> Historical freeze record. The Vite SPA and mandatory Redux Toolkit portions of the Web frontend
> decision below are superseded by accepted ADR-007. React and TypeScript remain through Next.js;
> pnpm is the Web package manager. All other frozen decisions remain subject to the documented
> priority rules.

**Status:** Approved for implementation
**Version:** 1.0

## Rule

The OpenInvest architecture is frozen. Fundamental decisions may not change without an accepted architecture decision record (ADR).

## Frozen decisions

- Monorepo: `OpenInvest/`
- Backend: Go + Fiber
- Analytics workers: Python
- Database: PostgreSQL
- Cache: Redis + RAM cache
- Frontend: React + Vite + TypeScript + Redux Toolkit
- Future mobile clients: SwiftUI + Jetpack Compose
- Architecture: API First, Domain-Driven Design, Clean Architecture, Event Driven
- Data: canonical database, snapshots, immutable transactions
- Security: Zero Trust and Privacy by Design
- External data: official and free sources accessed only through backend collectors
- The frontend never calls MOEX, CBR, or Rosstat directly
- Business data is never stored in LocalStorage
- Changes are never pushed automatically without user confirmation

## ADR required for

- backend language
- database
- API style
- snapshot strategy
- tax logic
- privacy model
- mathematical engine
- mobile strategy
- external data policy

## Stage 1 scope

Stage 1 contains repository structure, documentation, ADRs, minimal Go/Fiber, React/Vite/TypeScript/Redux Toolkit and Python/FastAPI skeletons, plus local PostgreSQL and Redis configuration. It contains no product business logic.
