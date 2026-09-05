# Stage 3.68 — Dividend Calculator Implementation

## Status

Implementation candidate. Publication, commit, push, Draft Pull Request creation, external review, and merge remain separately authorized governance steps.

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

`WITHHELD — external published-head phase pending`

Per `docs/REVIEW_WORKFLOW.md` v1.4.0, current pre-publication Internal findings/verdict are not published into the repository-visible Stage report before the External published-head verdict. They remain review-channel evidence only until the required post-External evidence-only publication step.
