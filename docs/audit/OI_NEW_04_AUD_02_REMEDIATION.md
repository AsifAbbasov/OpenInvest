# OI-NEW-04-AUD-02 — Snapshot persistence fan-out remediation candidate

## Status

`OI_NEW_04_AUD_02_STATUS=AWAITING_INDEPENDENT_REVIEW`

## Problem

ACTIVE C′ bounded raw-ledger replay capped mutable history at 5,000 rows, but snapshot persistence still issued one PostgreSQL `INSERT` statement per affected snapshot date. High distinct-date cardinality therefore amplified transaction duration and database round trips while the portfolio row remained serialized.

## Root cause

`executeFinancialReplayPlanTx` derived the exact ordered replay states correctly, then persisted each state through a separate `ExecContext` call.

## Failure scenario

A valid mutation can affect many existing snapshot dates at or after its earliest trade date. The previous implementation could therefore execute O(D) snapshot statements even though ledger replay itself remained bounded.

## Impact

Transaction/lock duration and database work could scale with affected snapshot-date cardinality independently of the existing raw-history bound.

## Initial remediation

N/A — this is the first remediation candidate for `OI-NEW-04-AUD-02`.

## Why review rejected it

N/A.

## Second attack scenario

N/A.

## Final remediation candidate

ACTIVE C′ keeps the existing deterministic replay-state calculation in Go, enforces `D <= replayBoundRawRows` (5,000), serializes the already-computed exact Decimal states, and persists all admitted snapshot versions with one set-based PostgreSQL `INSERT`. Snapshot rows remain append-only; no historical snapshot is updated or deleted.

## Why this solution

The date ceiling reuses the already-approved C′ work bound rather than inventing a smaller product limit. The independent SQL-statement bound is one statement per admitted ACTIVE C′ snapshot rebuild, so PostgreSQL round trips do not scale one-for-one with D. Financial methodology, WAC, immutable ledger semantics, snapshot versioning, R0/R1/R2 privileges, and activation governance remain unchanged.

## Regression evidence

The candidate adds a canonical ACTIVE-R2 import using 100 distinct trade dates, plus separate evidence for 5,000 distinct affected snapshot dates, 5,001 fail-closed atomic rollback, cancellation after the high-D set-based snapshot statement, same-portfolio high-D concurrency, and separate `SnapshotWrites` versus `SnapshotSQLStatements` instrumentation.

## Residual limitations

The accepted reversal R*B in-memory replay trade-off remains unchanged. `detectStaleRowsAgainstEpoch` / `logicalFamilyFrozenTx` query amplification remains a separate hardening lead and is not addressed here.

## Status

REMEDIATION CANDIDATE / AWAITING INDEPENDENT REVIEW
