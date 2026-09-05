# Stage 3.69 — Dividend Calculator Closure

| Field | Value |
| --- | --- |
| Status | CANONICAL / PROTECTED-ACTIVATED — Stage 3.68 lifecycle/documentation closure COMPLETE |
| Date | 2026-09-05 |
| Closed implementation stage | Stage 3.68 — Dividend Calculator Implementation |
| Detailed implementation/evidence record | `docs/stages/STAGE_03_68_DIVIDEND_CALCULATOR_IMPLEMENTATION.md` |
| Implementation record blob on closure base | `29b4a6a29f3a6c609a3dba38e6904c973f19be67` |
| Implementation PR | PR #133 — `feat: implement Stage 3.68 Dividend Calculator` |
| Frozen candidate identity SHA-256 | `06619d5ce086812868cd5f2469d8c735d22dc1236a4e6e0a7a6844a3ab898a84` |
| Frozen implementation manifest SHA-256 | `98bd9844ee3f3c135e167928225aaa88d756ac4ee9d760735974fd8b701cd0db` |
| Frozen complete patch SHA-256 | `a796960cf124293f0f77acabb26e02d12ae2bbc8ba65b34536319c8b7f2d5f1e` |
| Internal review report SHA-256 | `d35491b34e6948b42d89761193061fd6bc686cc4817d78a25be5f9d6d37ba347` |
| External-reviewed semantic head | `c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c` |
| External-reviewed semantic tree | `4eaf6be3616dae1bec593127b0115e1d6e7f39f3` |
| Final evidence head | `014a594695b94b9af424b06a3e38590fbb5281ff` |
| Final evidence tree | `be09503ceaafb8781cc82829f98e37cda5c6be6b` |
| Implementation squash merge | `6fb395ffcef12840133dac27294f653276adcdf6` |
| Protected post-merge tree | `be09503ceaafb8781cc82829f98e37cda5c6be6b` |
| Implementation parent / prior protected `develop` | `393f782b72347f9e98026940ce31b11c7cfbfcc6` |
| Semantic CI | CI #352 / run `33972964583` — 10/10 required jobs SUCCESS |
| Evidence CI | CI #353 / run `33974141987` — 10/10 required jobs SUCCESS |
| Internal review | APPROVED; P0=0 / P1=0 / P2 blocking=0 / P3 blocking=0 |
| Fresh External review | APPROVED; P0=0 / P1=0 / P2 blocking=0 / P3 blocking=0 |
| Exact evidence verification | APPROVED; runtime/product/API/database/math/security/dependency semantic drift NONE |
| Closure base | protected `develop@6fb395ffcef12840133dac27294f653276adcdf6` |
| Closure base tree | `be09503ceaafb8781cc82829f98e37cda5c6be6b` |
| Closure PR | PR #134 — `docs: close Stage 3.68 Dividend Calculator lifecycle` |
| Closure published head | `e98bad4b2431eb7f45a5bfff17c3601c49554b45` |
| Closure CI | CI #354 / run `33983374426` — 10/10 required jobs SUCCESS |
| Exact published-head closure verification | APPROVED; changed scope exactly three documentation files; runtime/product/API/database/math/security/dependency drift NONE |
| Closure squash merge | `fee7de358f0919802e16a398b19c8947bc852645` |
| Closure protected tree | `e7c9ddb96a5bf7add5204ccdf54aa20c190a0014` |
| Closure parent | `6fb395ffcef12840133dac27294f653276adcdf6` |
| Closure runtime scope | None — documentation/governance synchronization only |
| Canonical closure surfaces | this record; `docs/ROADMAP.md`; `docs/SOURCE_OF_TRUTH.md`; Stage 3.68 implementation record now carries an explicit historical/current-status boundary |
| Feature 3D | NOT STARTED by this closure; separate source/use planning gate required |
| External source activation | None |

## 1. Closure basis

Stage 3.68 delivered the provider-independent Dividend Calculator that was already part of the frozen MVP contract.
The governed implementation was squash-merged through PR #133 into protected `develop` at:

```text
6fb395ffcef12840133dac27294f653276adcdf6
```

with protected tree:

```text
be09503ceaafb8781cc82829f98e37cda5c6be6b
```

That protected tree exactly equals the final evidence tree. The squash merge therefore introduced no content drift
beyond the exact evidence-approved PR tree.

Stage 3.69 created no second implementation event. It synchronized lifecycle documentation after the already-completed
Stage 3.68 merge and was itself protected-activated through PR #134 at
`fee7de358f0919802e16a398b19c8947bc852645`, tree `e7c9ddb96a5bf7add5204ccdf54aa20c190a0014`.
The Stage 3.68 implementation record preserves its original pre-merge evidence chronology, while its active top-level
status now explicitly points to this canonical closure so historical wording cannot be mistaken for current state.

## 2. Why Stage 3.68 was implemented

Dividend calculation is a central MVP user capability that does not require a live Corporate Actions source.

The stage was intentionally implemented before Feature 3D because the user supplies all calculation inputs directly.
That gives OpenInvest immediate product value without pretending that dividend announcements, record dates, or payment
dates are available from an approved external source.

The resulting separation is deliberate:

```text
user-supplied calculation
  → available now through Stage 3.68

provider-derived Corporate Actions data
  → still gated by Feature 3D source/use approval
```

This keeps the calculator useful under the project's zero-budget/source-rights constraints while preserving the
fail-closed external-data policy.

## 3. What Stage 3.68 delivered

The public Web/API path is:

```text
/dividends/calculator
  → frontend-next/src/common/api/openinvest.ts
  → POST /api/v1/dividends/calculate
  → Go Dividend vertical slice
  → exact Decimal calculation
  → replay-aware exact response
```

The user supplies:

- ticker;
- quantity;
- dividend per unit;
- optional position cost.

The backend returns:

- the canonicalized calculator inputs;
- gross dividend;
- optional gross yield;
- `taxIncluded = false`;
- methodology `dividend-calculator-v1`.

The Dashboard exposes an unconditional navigation entry to `/dividends/calculator`.

No external financial-data source, portfolio mutation, tax calculation, or market-data enrichment participates in this
calculation.

## 4. Expected product behavior

The frozen financial semantics are:

- ticker must match `^[A-Z0-9]{1,32}$`;
- quantity must be positive;
- dividend per unit must be non-negative RUB;
- optional position cost must be positive RUB;
- `grossDividend = quantity × dividendPerUnit`;
- `grossYield = grossDividend / positionCost` when position cost is supplied;
- `grossYield = null` when position cost is absent;
- `taxIncluded = false`;
- derived overflow fails closed;
- calculation uses fixed-scale base-10 Decimal with scale 8 / precision 28 and half-even multiplication/division.

Canonical vector:

```text
ticker            = SBER
quantity          = 1000.00000000
dividendPerUnit   = 34.84000000 RUB
positionCost      = 280000.00000000 RUB
grossDividend     = 34840.00000000 RUB
grossYield        = 0.12442857
taxIncluded       = false
methodology       = dividend-calculator-v1
```

The frontend does not recalculate gross dividend or yield. It renders backend-owned exact decimal strings. Monetary
display rounds to two decimal places with exact half-even string arithmetic; yield presentation uses decimal-string
shifting rather than binary floating point.

## 5. Idempotency, retry, and concurrency behavior

`Idempotency-Key` remains mandatory.

For this public calculator, a calculator-specific domain separator plus the validated key is SHA-256-derived into a
technical UUIDv8 replay principal. The technical principal is not a user identity and is not derived from IP address,
cookie, or browser fingerprint.

The service ordering is:

```text
validate request/key
→ read-only replay lookup
→ exact replay / conflict / in-flight resolution
→ fresh-command admission
→ transactional replay reservation
→ calculate and persist exact replay artifact
```

An exact completed replay therefore wins over fresh-command rate limiting. If fresh admission is denied, one read-only
race recheck can still return a stronger replay/conflict/in-flight state; otherwise the request returns `429` without
creating a provisional replay row.

The browser retains the same key for a failed retry of an unchanged payload, releases it after success, and releases it
when calculator input changes. Input replacement/unmount aborts obsolete work, and a generation guard prevents stale
completion from overwriting a newer request.

## 6. Privacy, retention, performance, and cost boundaries

Stage 3.68 creates no calculator business-domain record, user/profile record, provider record, or portfolio mutation.

It does reuse the existing technical replay persistence. For the existing 24-hour replay authority window,
`investment.command_deduplication` may retain the technical replay principal, idempotency key, request hash, response
metadata, and exact response body. Because the response body contains submitted and derived calculator values, those
values can be present in the replay artifact for that existing window.

This is technical persistence, not zero persistence. Stage 3.68 introduces no new retention class or duration.

Fresh-command writable amplification is bounded per process by:

- 20 requests per idempotency key per minute;
- 1200 fresh admissions globally per minute;
- at most 4096 tracked keys;
- `Retry-After: 60` on the calculator rate-limit response.

The limiter is intentionally not represented as a distributed or edge DoS shield. A valid request still performs the
read-only PostgreSQL replay lookup required by exact idempotency semantics.

No paid API, Redis addition, worker, cron, new cache service, AI service, or new dependency was introduced.

## 7. What Stage 3.68 did not do

Stage 3.68 did not:

- activate Corporate Actions Feature 3D;
- connect MOEX, Finam, BCS, T-Invest, Interfax, NSD, CBR, issuer sites, or another external provider;
- add scraping;
- change OpenAPI;
- add a database schema or migration;
- add tax calculation/export;
- mutate portfolio state;
- add Redis, workers, cron, AI, or a new production dependency;
- authorize mobile, notifications, broker synchronization, or another product scope.

The calculator is therefore an exact user-supplied-input calculation surface, not a live dividend-data service.

## 8. Verification and review evidence

The frozen prepublication candidate identity is:

```text
candidate identity:
06619d5ce086812868cd5f2469d8c735d22dc1236a4e6e0a7a6844a3ab898a84

manifest:
98bd9844ee3f3c135e167928225aaa88d756ac4ee9d760735974fd8b701cd0db

complete patch:
a796960cf124293f0f77acabb26e02d12ae2bbc8ba65b34536319c8b7f2d5f1e
```

Internal review covered all 21/21 changed files, was read-only, and concluded:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The semantic publication was:

```text
head c4a87bf8cf4eeefc3dbf3e130e1a9e21b623952c
tree 4eaf6be3616dae1bec593127b0115e1d6e7f39f3
```

CI #352 / run `33972964583` completed with all ten required jobs successful.

Fresh External published-head review independently re-reviewed the published 21-file implementation and concluded:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The evidence-only follow-up was:

```text
head 014a594695b94b9af424b06a3e38590fbb5281ff
tree be09503ceaafb8781cc82829f98e37cda5c6be6b
```

CI #353 / run `33974141987` completed with all ten required jobs successful. Exact evidence-publication verification
confirmed that the follow-up changed only the Stage 3.68 documentation record and introduced no runtime/product/API/
database/math/security/dependency semantic drift.

## 9. Human authorization and protected merge

The Principal Architect explicitly authorized publication of the implementation candidate, publication of the
post-External evidence-only follow-up, synchronization of PR #133 metadata, and finally Ready + squash merge of exact
evidence head `014a594695b94b9af424b06a3e38590fbb5281ff`.

The Ready transition preserved the authorized head. Squash merge was executed fail-closed with `expected_head_sha`.
GitHub created the verified protected merge commit:

```text
6fb395ffcef12840133dac27294f653276adcdf6
```

Post-merge verification confirmed:

- PR #133 is closed and `merged=true`;
- protected `develop` points to `6fb395ffcef12840133dac27294f653276adcdf6` at the implementation activation point;
- its parent is `393f782b72347f9e98026940ce31b11c7cfbfcc6`;
- its tree is exactly `be09503ceaafb8781cc82829f98e37cda5c6be6b`;
- the merge commit is GitHub-verified;
- the protected branch retains the required CI checks.

## 10. Remaining non-blocking debt

Two non-blocking notes remain recorded without changing the Stage 3.68 correctness verdict:

1. the pre-existing generic replay error helper can emit `503 SERVICE_NOT_READY`, while the frozen calculator
   operation explicitly lists `200/400/409/429`; current production composition supplies the required replay
   capabilities, so this remains repo-wide contract debt rather than a normal Stage 3.68 runtime path;
2. the anonymous technical replay principal derived from `Idempotency-Key` is acceptable for this public calculator
   surface, but it must not be copied to a future sensitive anonymous response surface without a separately reviewed
   client-scope boundary.

Neither note reopens Stage 3.68.

## 11. Canonical documentation synchronization

The protected Stage 3.69 activation through PR #134 synchronized the canonical closure state in:

1. `docs/stages/STAGE_03_69_DIVIDEND_CALCULATOR_CLOSURE.md`;
2. `docs/ROADMAP.md`;
3. `docs/SOURCE_OF_TRUTH.md`.

A final documentation-only textual synchronization additionally marks the Stage 3.68 implementation record as an
explicit historical/evidence record with a current Stage 3.69 closure pointer. This does not create Stage 3.70 and does
not reopen the implementation or closure lifecycle.

The Stage 3.69 synchronization also corrected the already-activated Stage 3.67 lifecycle wording: Stage 3.67 is
canonical through PR #131 / protected squash merge `c885f6e57ea08e4583103fe2f22f142bf13a8560`, tree
`53f823e92b02721211e1d89f8af6374fffc252ae`.

No runtime/test/OpenAPI/dependency/CI/database/provider/source behavior changes.

## 12. Feature 3D boundary after closure

Feature 3D remains a separate source/use planning gate.

Stage 3.69 does not authorize:

- a real Corporate Actions adapter;
- provider HTTP;
- runtime source activation;
- public-display/redistribution rights;
- caching/persistence rights for provider-derived data;
- provider cost or rate policy.

Those decisions still require an exact provider/use mode, Data Source Registry approval, legal/contractual production
use, public-display rights, cost acceptance, traffic/failure policy, retention/caching rights, provenance/freshness
obligations, and separately reviewed runtime composition.

The user-supplied Dividend Calculator is complete independently of that unresolved source gate.

## 13. Governance path

Stage 3.69 followed the `docs/REVIEW_WORKFLOW.md` v1.4.0 post-development governance/closure path as a
documentation/governance-only closure.

Completed activation sequence:

```text
documentation candidate
→ deterministic documentation checks
→ Governance / Closure review APPROVED
→ explicit human commit/push authorization
→ Draft PR #134 targeting develop
→ exact-head CI #354 / run 33983374426 — 10/10 SUCCESS
→ exact-published-head Governance / Closure verification APPROVED
→ explicit human Ready/squash-merge authorization
→ protected develop activation at fee7de358f0919802e16a398b19c8947bc852645
```

Protected `develop` was the activation boundary. Activation of Stage 3.69 authorizes no Feature 3D implementation,
external HTTP/provider use, source activation, tax, portfolio mutation, or unrelated product work.

## 14. Closure decision

Protected activation occurred through PR #134, exact published head
`e98bad4b2431eb7f45a5bfff17c3601c49554b45`, CI #354 / run `33983374426` 10/10 SUCCESS, and squash merge
`fee7de358f0919802e16a398b19c8947bc852645` with protected tree
`e7c9ddb96a5bf7add5204ccdf54aa20c190a0014`.

Therefore the canonical state is:

```text
Stage 3.68 implementation = COMPLETE / MERGED
Stage 3.68 lifecycle/documentation closure = COMPLETE
Stage 3.69 closure governance = COMPLETE / PROTECTED-ACTIVATED
Dividend Calculator = canonical MVP functionality on develop
Provider-derived live dividend data = NOT AUTHORIZED by this closure
Feature 3D = separate source/use planning gate
Next product work = separately reviewed and explicitly authorized
```
