# ADR-001: Use Go and Fiber for the backend

- Status: Accepted
- Date: 2026-06-19

## Context

OpenInvest requires a fast API-first backend with explicit server-side business logic and predictable resource use.

## Decision

Use Go 1.24 or newer with Fiber for the primary backend API. Keep domain logic independent from the HTTP framework through Clean Architecture boundaries.

## Consequences

The backend gains strong concurrency and deployment characteristics. Fiber-specific code must remain at the transport boundary, and changing the backend language requires a superseding ADR.
