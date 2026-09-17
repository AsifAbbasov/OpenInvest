# OI-NEW-04 C′ replay-policy operations

This document is operational guidance for the additive OI-NEW-04 replay-state schema. The immutable `investment.transaction_entries` ledger remains the source of truth. Replay epochs are derived state and are append-only after creation.

The `000012_oi_new_04_replay_epochs.down.sql` migration is an exact migration-test inverse only. It is **not** an ordinary production rollback after a generation or replay state has been populated or activated. After activation, rollback is performed by draining financial writes and portfolio creation, appending `INVALIDATED(N)`, stopping A2 processes, converging the runtime role to R0, proving the exact R0 capability set, and only then starting the legacy application. The invalidated generation is never reactivated; a later bounded activation uses N+1.

Generation existence is not activation. Owner tooling allocates generation N in PREACTIVE state before any epoch references N. Provisional backfill may then append epochs while policy remains inactive. Finalization freezes financial mutations and portfolio creation, rereads the exact active portfolio set, validates every ledger, appends a final epoch for every portfolio, creates the deterministic activation manifest, verifies it, and only then appends `ACTIVATED(N)`. Any portfolio failure aborts finalization and no activation event is written.

Runtime capability transitions require a full process drain. R0 has no replay-table capability. R1 has SELECT on replay policy/state tables only. R2 adds INSERT on `portfolio_replay_epochs`, `portfolio_replay_positions`, and `portfolio_replay_financial_state`; policy generations/events remain owner-only. No R profile grants UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, CREATE, sequence access, or grant option for replay state.

After `ACTIVATED(N)`, only A2/R2 may serve financial mutations. The application dispatcher must route append, import, correction, and reversal through the bounded C′ engine. Missing or stale replay state fails closed. There is no automatic whole-history reconstruction and no legacy fallback while policy N is active.
