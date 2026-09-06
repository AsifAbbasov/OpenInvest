# Stage 3.71 — Portfolio Position & Cost Basis Engine review evidence

Status: adversarial review evidence for the Stage 3.71 implementation candidate. This record does not merge, activate, or independently approve Stage 3.71.

Canonical base: protected `develop@b772e52221fbb694b3116bd1b579db99d4e56302`.

Frozen implementation authority: `docs/stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_IMPLEMENTATION.md`, content SHA-256 `3a32bcd287211338348900458dfb5787acdeb777858acc8da1032a36a3124758`. That hash-bound authority document remains unchanged so migration policy references do not become a moving target.

Required rollout amendment: `docs/stages/STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_ROLLOUT_AMENDMENT.md`.

Reviewed implementation/code head before governance-only evidence commits: `6a1668fa9ecf6833f2d25fbdbe6b8bd9f5092e64`.

Pull request: `#138` — `feat: implement Stage 3.71 portfolio position and cost basis runtime`.

## 1. Review scope

The review covered the complete Stage 3.71 diff against the canonical base, with emphasis on:

- WAC and remaining acquisition-basis arithmetic;
- deterministic same-BusinessDate ledger ordering;
- manual SELL admission and oversell rollback;
- replay/idempotency transaction boundaries;
- PostgreSQL migration/backfill/readiness activation order;
- concurrent BUY/SELL and SELL/SELL behavior;
- snapshot rebuilding and summary read selection;
- OpenAPI/HTTP/Web 409 behavior;
- import SELL exclusion;
- compatibility impact on prior Stage 3.x tests and contracts;
- deployment/cutover behavior while a pre-Stage-3.71 runtime may still be writable.

The reviewed code diff remained limited to Stage 3.71 runtime, migrations, tests, OpenAPI/Web exposure, and implementation documentation. No public position DTO, market-price provider, unrealized P/L, XIRR, FIFO/tax lots, correction/reversal runtime, imported SELL, notifications, AI, Redis/workers, or Feature3D was activated.

## 2. Substantive defects found and remediated

### 2.1 Exact-index readiness/backfill SQL type defect

The first exact-index catalog check compared the result of `array_agg(pg_attribute.attname ...)`, whose PostgreSQL element type is `name`, with a `text[]` literal. Real PostgreSQL execution failed with:

```text
operator does not exist: name[] = text[]
```

Impact if left unresolved:

- owner-only Stage 3.71 backfill could not prove the required `000009` index;
- runtime Stage 3.71 readiness could fail with a SQL execution error even when schema/data were otherwise correct;
- a superficially green test set without direct execution of these paths would not prove activation safety.

Remediation:

- cast `pg_attribute.attname::text` before ordered array comparison in both backfill and runtime readiness;
- add real PostgreSQL backfill coverage;
- add a real PostgreSQL `Stage371Ready` integration regression.

The backfill now verifies the exact Stage 3.71 unique index before population and fails closed if the index is missing, invalid, not ready, partial, expression-based, or structurally different from ordered `(portfolio_id, ledger_sequence)`.

### 2.2 Snapshot summary wall-clock ordering defect

The legacy summary read selected the newest snapshot on a date by `calculated_at DESC`. Stage 3.71 writes are serialized by the portfolio database lock, but `command.Now` is request-derived before that lock. Therefore a later serialized financial write can legally carry an earlier timestamp than a preceding write.

Impact if left unresolved:

- the ledger could contain both committed writes;
- Stage 3.71 could have produced the correct later snapshot version;
- the public summary read could still return the previous snapshot merely because its wall-clock timestamp was newer.

Remediation:

Stage 3.71 summary selection now orders calculated snapshots by:

```text
snapshot_date DESC
-> methodology priority (Stage 3.71 before Stage 3.02)
-> snapshot_version DESC
-> calculated_at DESC
-> id DESC
```

Financial/business selection therefore uses date, methodology, and serialized `snapshot_version`. Timestamp and UUID are only deterministic tie-breakers.

The regression fixture intentionally gives Stage 3.71 version 1 a newer `calculated_at` than version 2 and also stores a Stage 3.02 version 99 with an even newer timestamp. The required result remains Stage 3.71 version 2 with the expected acquisition-basis values.

### 2.3 Populate-to-runtime legacy-write cutover race

The schema intentionally keeps `ledger_sequence` nullable during Expand so applying `000008` does not itself break the legacy runtime. The backfill atomically populates all existing rows and intentionally refuses a partially populated state.

A final rollout review exposed this sequence:

```text
backfill succeeds
-> table lock is released
-> legacy runtime accepts one more transaction with ledger_sequence = NULL
-> Stage 3.71 readiness fails closed
-> second backfill refuses partial population
```

This does not silently corrupt financial state; both readiness and backfill fail closed. However, the original rollout wording did not close the operational gap between Populate completion and new-runtime readiness.

Remediation is a required rollout amendment:

- enter portfolio-write quiescence before backfill;
- keep legacy portfolio writes stopped/blocked throughout Populate and deployment;
- deploy the reviewed Stage 3.71 runtime while quiescence remains active;
- require Stage 3.71 readiness success;
- reopen portfolio writes only after readiness succeeds;
- abort cutover rather than attempting ad hoc partial sequence repair if a legacy write is detected.

The full contract is recorded in `STAGE_03_71_PORTFOLIO_POSITION_COST_BASIS_ROLLOUT_AMENDMENT.md`.

## 3. Backfill and rollout evidence

The owner-only backfill now has direct PostgreSQL tests rather than `[no test files]` coverage. Evidence includes:

- exact ADR-009 historical tuple:
  `trade_date ASC, created_at ASC, transaction_id ASC, revision ASC, entry_id ASC`;
- portfolio-local sequence reset;
- exact-index requirement before mutation;
- rejection of partial population without mutation;
- rejection of pre-Stage-3.71 historical SELL;
- rejection of correction/reversal/unsupported revision state;
- completed-population rerun as verify-only;
- structural verification of positive, non-null, unique portfolio-local sequence values.

The activation order is now interpreted together with the rollout amendment:

1. apply `000008` nullable `ledger_sequence BIGINT` Expand migration;
2. apply `000009` concurrent unique index;
3. enter and prove portfolio-write quiescence;
4. run owner-only deterministic population;
5. verify exact index and complete positive sequence state;
6. deploy Stage 3.71 while writes remain quiesced;
7. require runtime readiness success;
8. only then reopen portfolio writes.

The runtime role remains read/append-only on the immutable ledger and does not receive UPDATE/DELETE/TRUNCATE authority.

## 4. Financial and concurrency evidence

The candidate includes evidence for:

- canonical BUY + BUY Weighted Average Cost;
- partial SELL preserving authoritative WAC exactly;
- full close and later reopen;
- fractional Half-Even non-invertibility witness;
- derived NUMERIC(28,8) overflow rejection;
- backdated oversell against a later SELL;
- same-BusinessDate `ledger_sequence` ordering;
- concurrent SELL + SELL where only one illegal-overlap outcome can commit;
- concurrent BUY + SELL accepting only valid serialized outcomes;
- rejected oversell leaving no ledger row, snapshot mutation, or success replay artifact;
- exact oversell retry with the same idempotency key;
- import sequence allocation in append-plan order;
- crafted/imported SELL remaining rejected;
- STOCK/BOND snapshot values using remaining acquisition basis;
- deterministic repeated rebuild.

The public `summary.positions` projection remains unavailable; Stage 3.71 does not invent market price/value fields.

## 5. API and Web evidence

`POST /api/v1/portfolios/{portfolioId}/transactions` uses `TransactionConflict` for HTTP 409 and covers both:

- existing idempotency conflict;
- `INSUFFICIENT_POSITION_QUANTITY`.

The Web transaction form exposes manual SELL only. On an oversell 409, the browser does not clear the idempotency intent; exact retry can therefore retain the same key. Imported SELL remains outside Stage 3.71 scope.

## 6. CI evidence

GitHub Actions run `#375` completed successfully for exact implementation/code head:

```text
6a1668fa9ecf6833f2d25fbdbe6b8bd9f5092e64
```

Passing gates:

- PostgreSQL migration validation;
- Go tests;
- Go race tests;
- Go vet;
- Go vulnerability scan;
- dependency security scan;
- frontend tests/typecheck/build;
- OpenAPI contract;
- Python tests;
- Docker Compose validation.

Governance-only evidence/amendment commits were created after that code head. The current PR head must receive a fresh complete green CI run before Ready or merge decisions; CI #375 must not be misrepresented as validation of later bytes.

## 7. Review disposition

No unresolved inline review threads or submitted external pull-request reviews existed when this evidence was recorded.

No known Stage 3.71 product-code blocker remained after the code remediations and green CI on the reviewed implementation head.

The cutover race is an activation/governance blocker unless the rollout amendment is followed; it does not require changing the reviewed financial runtime implementation.

One architectural compatibility constraint remains explicit: the production API uses `*postgres.Store`, which implements Stage 3.71 capabilities. Legacy/test Store implementations may use compatibility fallback and must not be treated as authorization for a future non-Stage-3.71-capable production persistence backend.

## 8. Governance gate

This record does **not** claim:

- independent reviewer approval;
- Ready-for-review state;
- merge approval;
- protected `develop` activation;
- production rollout;
- closure of Stage 3.71;
- authorization for Stage 3.72 or any later feature.

The candidate may move from Draft to Ready only after the current exact PR head has complete green CI and the repository's remaining human/governance review requirement is satisfied. Merge remains a separate explicit human action.