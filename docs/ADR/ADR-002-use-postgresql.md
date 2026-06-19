# ADR-002: Use PostgreSQL as the canonical database

- Status: Accepted
- Date: 2026-06-19

## Context

Financial history needs transactional consistency, auditability, constraints, and reproducible calculations.

## Decision

Use PostgreSQL as the canonical durable database. Redis and RAM caches are non-canonical acceleration layers and must not become sources of truth.

## Consequences

Schemas and migrations require careful review. Database replacement requires a superseding ADR and a verified migration strategy.
