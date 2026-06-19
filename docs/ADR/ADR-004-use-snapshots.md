# ADR-004: Use versioned portfolio snapshots

- Status: Accepted
- Date: 2026-06-19

## Context

Portfolio analytics must load quickly while remaining reproducible against immutable financial history.

## Decision

Use versioned, rebuildable snapshots derived from canonical immutable transactions. Snapshots are projections, not the source of truth.

## Consequences

Snapshot algorithms need versions, deterministic rebuilds, and invalidation rules. Changing this strategy requires a superseding ADR.
