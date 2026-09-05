# Stage 3.68 — Dividend Calculator Implementation

## Status

**HISTORICAL IMPLEMENTATION / EVIDENCE RECORD — current lifecycle status is governed by Stage 3.69 closure.**

Stage 3.68 is **COMPLETE / MERGED**. PR #133 was squash-merged into protected `develop` at `6fb395ffcef12840133dac27294f653276adcdf6` with protected tree `be09503ceaafb8781cc82829f98e37cda5c6be6b`. The implementation passed CI #352 / run `33972964583` 10/10, fresh External published-head review `APPROVED`, evidence-head CI #353 / run `33974141987` 10/10, exact evidence-publication verification `APPROVED`, and explicit human Ready/squash-merge authorization. Stage 3.68 lifecycle/documentation closure is canonical through Stage 3.69 / PR #134, squash-merged into protected `develop` at `fee7de358f0919802e16a398b19c8947bc852645`, tree `e7c9ddb96a5bf7add5204ccdf54aa20c190a0014`, after exact-head closure CI #354 / run `33983374426` 10/10 and exact published-head closure verification `APPROVED`.

Historical pre-merge status snapshot preserved for chronology:

> Development implementation is published in Draft PR #133. Exact-head CI #352 and the fresh External published-head review are complete and approved on pre-evidence head `c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c`. This Stage is not merged or accepted for merge yet; the mandatory evidence-only follow-up, evidence-head CI/verification, and separate human Ready/squash-merge authorization remain outstanding governance gates.

Canonical base:

- `develop@393f782b72347f9e98026940ce31b11c7cfbfcc6`
- tree `06a7dc77eca18b5fc3c6c4ef832c5adbb3388f6a`

## Scope

Stage 3.68 implements the existing `POST /api/v1/dividends/calculate` contract without changing its business meaning or OpenAPI files.

The calculator is public and idempotent. It accepts user-supplied ticker, quantity, dividend per unit, and optional position cost. The Go backend exclusively owns financial arithmetic.

The implementation does not activate or depend on Corporate Actions Feature 3D or any external financial-data provider.

## Financial semantics

- Decimal arithmetic uses the existing base-10 fixed scale 8 implementation and half-even multiplication/division.
- `quantity` must be positive and fit the canonical Decimal storage domain.
- `dividendPerUnit` must be non-negative RUB.
- `positionCost`, when supplied, must be positive RUB.
- `grossDividend = quantity × dividendPerUnit`.
- `grossYield = grossDividend / positionCost` when position cost is supplied; otherwise `grossYield = null`.
- Derived gross dividend and yield must fit the canonical `NUMERIC(28,8)` precision domain.
- `taxIncluded = false`.
- methodology is `dividend-calculator-v1`.

Canonical vector:

```text
quantity          = 1000.00000000
dividendPerUnit   = 34.84000000 RUB
positionCost      = 280000.00000000 RUB
grossDividend     = 34840.00000000 RUB
grossYield        = 0.12442857
```

## Public idempotency

The existing mandatory `Idempotency-Key` contract is preserved.

The calculator does not create a user, cookie, IP-derived identity, browser fingerprint, or portfolio record. After the key is validated, the handler derives a deterministic UUIDv8 technical replay scope from SHA-256 over a calculator-specific domain separator plus the key. Therefore:

- the same validated key resolves to the same anonymous technical scope;
- distinct keys are domain-separated and hashed into independently derived UUIDv8 technical scopes; this is a cryptographic collision-resistance property, not a mathematical injectivity claim, and the original idempotency key remains part of the replay uniqueness tuple;
- the technical scope is separated from `devSubjectID` and authenticated subject identity;
- replay remains bound by existing `(principal, method, canonical path, idempotency key, request hash)` semantics.

The existing 24-hour `investment.command_deduplication` exact-response replay mechanism is reused. No new table, migration, or calculator business record is introduced.

### Privacy and retention boundary

Exact replay is technical persistence, not zero persistence. For up to the existing 24-hour idempotency window, `investment.command_deduplication` can retain the calculator's technical replay principal, idempotency key, request hash, response metadata, and exact response body. Because the response body contains the submitted ticker/quantity/dividend-per-unit/optional position-cost values and derived result, those financial values can therefore be present in the replay artifact for that retention window.

Stage 3.68 does not introduce a new retention class or extend that window. It reuses the privacy/correlation surface and operational retention controls already established for `investment.command_deduplication` by the existing privacy/idempotency stages. This Stage makes no claim that calculator requests are never persisted; the narrower claim is that no calculator business-domain record, portfolio mutation, user profile, or provider-derived record is created.

A bounded in-memory admission limiter caps public fresh-command writable amplification without IP tracking. Before any fresh writable reservation, the service uses the existing read-only replay lookup to resolve an exact completed replay, idempotency conflict, or in-flight command. Only a request with no current replay state invokes the fresh-command admission callback. If admission is denied, the service performs one read-only race recheck and returns any stronger replay/conflict/in-flight state that appeared concurrently; otherwise it returns `429` without inserting a provisional deduplication row. A successful admission then enters the existing transactional reservation path, which still resolves races atomically. Therefore completed exact replay is never replaced by `429`, and rate-limited fresh requests do not create transient database writes. The mandatory replay check is still a read-only PostgreSQL lookup, so this in-process control is not claimed to be a distributed or edge denial-of-service shield; that would be a separate deployment-level concern if the service later scales horizontally.

## Web architecture

The only API client authority remains:

```text
frontend-next/src/common/api/openinvest.ts
```

No feature-specific `src/common/api/dividends.ts` client exists.

The Web route is:

```text
/dividends/calculator
```

The component sends exact decimal strings to Go and renders the returned result. It does not recompute `grossDividend` or `grossYield`. Money presentation rounds the returned amount strings to the canonical two-decimal UI precision with exact half-even string arithmetic, while percentage display converts the returned ratio to presentation text using decimal-string shifting. Neither presentation path uses binary floating point.

Request lifecycle rules:

- input change aborts an obsolete request and invalidates its generation;
- unmount aborts the active request;
- stale completions cannot commit UI state;
- every failed attempt retains the key for retry of the exact same unchanged payload, including transport uncertainty and an in-flight HTTP response;
- successful completion releases the current idempotency key;
- changing input releases the retry identity and the next submission receives a new key.

The Dashboard exposes an unconditional link to `/dividends/calculator` from its hero section.

## Explicit exclusions

Stage 3.68 adds none of the following:

- Feature 3D provider activation;
- MOEX, Finam, BCS, T-Invest, or scraping;
- Redis;
- worker or cron processing;
- artificial intelligence;
- tax calculation or tax export;
- portfolio mutation;
- database schema or migration changes;
- OpenAPI changes;
- new dependencies.

## Verification expectations

Pre-publication evidence must cover focused calculator arithmetic, HTTP route behavior, replay behavior, shared frontend client wiring, component cancellation/retry semantics, route/OpenAPI parity, Dashboard navigation, exact candidate file identity, and scope scans.

The local execution environment available during candidate preparation does not satisfy the repository toolchain contract (`go 1.25.14`, Node `>=22.22.2`, pnpm `11.8.0`) and has no canonical PostgreSQL/frontend dependency tree. Therefore full repository Go, PostgreSQL, React, Next.js, OpenAPI, and security gates remain mandatory on the exact published head before external review and merge.

## Internal Review Evidence

The mandatory prepublication Internal review evidence was withheld from the Draft PR/repository until the fresh External published-head verdict, as required by `docs/REVIEW_WORKFLOW.md` v1.4.0.

Internal review report SHA-256:

```text
d35491b34e6948b42d89761193061fd6bc686cc4817d78a25be5f9d6d37ba347
```

Final Internal verdict before publication authorization:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The Internal reviewer performed complete read-only review of all 21 changed files and made no repository edits. Builder remediation was applied before final freeze. Resolved prepublication findings/hardening included:

- removing the separate feature-specific frontend HTTP client in favor of canonical `openinvest.ts`;
- replacing the original anonymous replay subject design with deterministic domain-separated technical replay scope;
- strengthening PostgreSQL replay integration evidence and cleanup behavior;
- removing trim-first backend ticker acceptance so the OpenAPI ticker grammar fails closed;
- retaining browser retry identity across failed unchanged-payload attempts and releasing it on success/input change;
- moving exact replay ahead of fresh-command rate limiting;
- moving fresh-command admission ahead of writable replay reservation by reusing the canonical read-only replay lookup;
- correcting privacy/retention documentation to disclose the existing 24-hour technical replay persistence boundary;
- synchronizing generic replay-lookup comments with the legitimate fresh-admission use;
- changing money UI presentation from scale-8 trim-only display to exact two-decimal half-even display;
- correcting hash-to-UUID documentation from absolute uniqueness language to cryptographic collision-resistance language.

Frozen prepublication evidence identity:

```text
Candidate identity SHA-256:
06619d5ce086812868cd5f2469d8c735d22dc1236a4e6e0a7a6844a3ab898a84

Complete candidate patch SHA-256:
a796960cf124293f0f77acabb26e02d12ae2bbc8ba65b34536319c8b7f2d5f1e

Frozen manifest SHA-256:
98bd9844ee3f3c135e167928225aaa88d756ac4ee9d760735974fd8b701cd0db
```

## Published-head verification chronology

The exact 21-file candidate was published as one implementation commit:

```text
c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c

tree:
4eaf6be3616dae1bec593127b0115e1d6e7f39f3
```

Draft PR #133 targets the unchanged canonical base `develop@393f782b72347f9e98026940ce31b11c7cfbfcc6`. Before the Draft PR was created, the published commit was independently reconciled against the frozen manifest: all 21/21 changed paths and 21/21 Git blob identities matched, with exactly 7 modified and 14 added files and no extras.

GitHub Actions run `33972964583` / CI #352 completed successfully on exact pre-evidence head `c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c`. All ten required jobs passed:

```text
Go tests                         PASS
Go race tests                    PASS
Go vet                           PASS
Go vulnerability scan           PASS
Python tests                     PASS
Frontend build and typecheck     PASS
OpenAPI contract                 PASS
Docker Compose config            PASS
PostgreSQL migration validation  PASS
Dependency security scan         PASS
```

The frontend job specifically passed Typecheck, Test, and Build. The Go test and race jobs ran against migrated PostgreSQL with the configured least-privilege runtime role.

## External published-head review

After CI #352 was green, the designated review chat performed a fresh External review of the complete published PR #133 diff/evidence on pre-evidence head `c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c`. The External phase did not use the earlier Internal verdict/findings as supporting evidence.

Review coverage included:

- architecture, DDD/SOLID/KISS/YAGNI boundaries and reuse of the canonical shared frontend client;
- API/OpenAPI parity, public idempotency semantics, replay/conflict/in-flight ordering and failure behavior;
- exact Decimal financial arithmetic, overflow handling, nullable yield, gross-only/tax semantics, and two-decimal UI money presentation;
- PostgreSQL replay reservation ordering, fresh-command admission, race recheck and no-write behavior on denied fresh commands;
- security/privacy, anonymous technical replay scope, 24-hour replay retention disclosure and public cost/amplification controls;
- frontend abort/generation correctness, retry-key lifecycle, stale completion suppression and unmount cancellation;
- scope/YAGNI, absence of Feature 3D/provider/schema/OpenAPI/dependency drift, and exact 21-file PR scope;
- exact-head CI #352 and absence of unresolved PR comments/review threads at verdict time.

Final External verdict on pre-evidence head `c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c`:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

Remaining non-blocking notes:

- the pre-existing generic replay error helper can emit `503 SERVICE_NOT_READY`, while the frozen calculator operation explicitly lists `200/400/409/429`; current production composition provides the required replay capabilities, so this is recorded as repo-wide contract debt rather than a new normal Stage 3.68 runtime path;
- deriving the calculator's anonymous technical replay principal from the validated `Idempotency-Key` is acceptable for this public calculator surface, but it must not be treated as precedent for a future sensitive anonymous response surface without a separately reviewed client-scope boundary.

## Historical governance state and next gate (pre-merge snapshot)

Development implementation, exact publication, exact-head CI #352, and the fresh External published-head review are complete.

This documentation change is the mandatory evidence-only follow-up required by `docs/REVIEW_WORKFLOW.md` v1.4.0. It publishes the previously withheld Internal evidence only after the External verdict and records the exact published-head CI/External-review evidence. It changes no runtime, product, API, database, mathematical, security/privacy behavior, dependency, or Feature 3D activation state. It does not authorize Ready, merge, branch deletion, or any protected-branch mutation.

Historical remaining sequence at that evidence-publication point:

```text
required GitHub CI on evidence-only head
→ same designated review chat exact evidence-publication verification
→ explicit Principal Architect Ready + squash-merge authorization
→ squash merge to protected develop
```

A no-new-finding exact evidence verification may remain as live review evidence and does not require another repository commit solely to embed its own verdict.

## Current lifecycle status

Those historical gates were subsequently completed. Final evidence head `014a594695b94b9af424b06a3e38590fbb5281ff` passed CI #353 / run `33974141987` 10/10, exact evidence-publication verification returned `APPROVED`, the Principal Architect explicitly authorized Ready + squash merge, and PR #133 was squash-merged into protected `develop` at `6fb395ffcef12840133dac27294f653276adcdf6` with tree `be09503ceaafb8781cc82829f98e37cda5c6be6b`.

The documentation/governance lifecycle was then closed through Stage 3.69 / PR #134. Exact closure head `e98bad4b2431eb7f45a5bfff17c3601c49554b45` passed CI #354 / run `33983374426` 10/10 and exact published-head closure verification `APPROVED`; PR #134 was explicitly authorized and squash-merged into protected `develop` at `fee7de358f0919802e16a398b19c8947bc852645`, tree `e7c9ddb96a5bf7add5204ccdf54aa20c190a0014`.

Therefore Stage 3.68 implementation and lifecycle/documentation closure are **COMPLETE**. Feature 3D remains a separate source/use planning gate and was not activated by Stage 3.68 or Stage 3.69.
