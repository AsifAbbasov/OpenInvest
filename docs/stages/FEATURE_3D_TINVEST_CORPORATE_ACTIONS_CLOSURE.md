# Feature 3D — T-Invest Corporate Actions Post-Merge Closure

| Field | Value |
| --- | --- |
| Status | Post-development governance / closure candidate; becomes canonical only after this docs-only closure PR is squash-merged into protected `develop` |
| Date | 2026-09-09 |
| Repository | `AsifAbbasov/OpenInvest` |
| Canonical implementation base | `develop@247081a95a7daf33c0077c88c5f41cb2e8161865` |
| Implementation PR | `#164` |
| Implementation head reviewed by External phase | `7c021e74db3a66488c4c6d87412729f524d9023f` |
| Final evidence head | `ddc5b5be36f0b5ae127c6ee430918a9dba1b453e` |
| Canonical squash merge | `247081a95a7daf33c0077c88c5f41cb2e8161865` |
| Canonical merged tree | `ec7bc9152210913b1a6ef742bddd599abee50bc7` |
| Source/use mode | `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` |
| Source/use rights | `CONDITIONAL-GO` for the exact constrained mode only |
| Adapter implementation | COMPLETE / MERGED |
| Runtime activation | NO |
| Live T-Invest token | NOT USED by the Feature 3D implementation/review workflow; deployed secret stores were not inspected by this closure |
| Production provider traffic | NOT AUTHORIZED / NOT CLAIMED by Feature 3D; production telemetry was not inspected by this closure |
| Persistent provider storage | NO |
| Background polling/synchronization | NO |
| Ledger auto-mutation | NO |
| Stage 3.78 | NOT STARTED / NOT AUTHORIZED |

## 1. Purpose and authority

This document closes the **documentation lifecycle** of Feature 3D after the implementation and evidence chain were squash-merged through PR `#164`. It does not rewrite the historical implementation dossier, review-evidence publication, or evidence errata. Those remain time-specific evidence:

- `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_IMPLEMENTATION.md`
- `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE.md`
- `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE_ERRATA.md`

The post-merge documentation problem was repository drift: the codebase already contained the constrained T-Invest adapter while several canonical registries still described Feature 3D as a future implementation/source-use gate.

This closure keeps four states separate:

```text
source/use rights      = CONDITIONAL-GO for one exact constrained mode
adapter implementation = COMPLETE / CANONICAL after PR #164 merge
documentation closure  = PENDING this docs-only governance path
runtime activation     = NO
production traffic     = NOT AUTHORIZED / NOT CLAIMED by Feature 3D
```

`implementation present` must never be read as `provider activated`.

## 2. Canonical implementation identity and evidence

Feature 3D implementation was published as PR `#164` from the reviewed implementation commit:

```text
7c021e74db3a66488c4c6d87412729f524d9023f
feat: add constrained T-Invest corporate actions provider
```

The evidence-only follow-up chain ended at:

```text
ddc5b5be36f0b5ae127c6ee430918a9dba1b453e
```

PR `#164` was squash-merged into protected `develop` as:

```text
247081a95a7daf33c0077c88c5f41cb2e8161865
```

with direct parent:

```text
e4e676f8a168bf1a2618185ee345cab245d1225d
```

and tree:

```text
ec7bc9152210913b1a6ef742bddd599abee50bc7
```

Authoritative CI evidence:

```text
implementation head CI #499 / run 34353668492 = 10/10 SUCCESS
evidence head CI       #501 / run 34356672643 = 10/10 SUCCESS
```

Review evidence:

```text
final Internal review                   = APPROVED
fresh External published-head review    = APPROVED
blocking P0/P1/P2                       = 0 / 0 / 0
post-evidence no-semantic-drift review  = APPROVED
```

## 3. Exact source/use and runtime boundary

Only this source/use mode is approved:

```text
TINVEST_CORPORATE_ACTIONS_CONSTRAINED
```

Allowed methods:

```text
GetDividends   -> canonical DIVIDEND
GetBondCoupons -> canonical COUPON
```

Static provider mapping:

```text
SBER         -> SBER_TQBR          -> SHARE
GAZP         -> GAZP_TQBR          -> SHARE
SU26238RMFS4 -> SU26238RMFS4_TQOB  -> BOND
```

Not authorized by Feature 3D:

- `GetBondEvents`;
- `FindInstrument` / `GetInstrumentBy` discovery;
- `MarketDataService`, `OrdersService`, `OperationsService`;
- account or portfolio synchronization;
- trading, prices, forecasts;
- scraping or unapproved-source fallback;
- background polling/synchronization;
- persistent raw provider payloads, DB archive, historical raw archive;
- cross-request provider cache or Redis provider cache;
- SDK/gRPC expansion;
- provider-specific public API fields;
- new canonical Corporate Action kinds.

Broader T-Invest use requires a new source/use review and explicit registry approval.

## 4. Financial-truth and runtime activation boundary

T-Invest Corporate Actions are external reference/event evidence only. They do not automatically become realized portfolio truth and must not automatically create or mutate `DIVIDEND`/`COUPON` ledger transactions, cash balances, realized income, tax transactions, cost basis, or market valuation.

The provider is disabled by default. Construction requires both:

```text
OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED=true
OPENINVEST_TINVEST_READONLY_TOKEN=<valid server-side read-only token>
```

The token alone does not activate the provider. Enablement without a valid non-empty token fails closed. Feature 3D implementation/review used no live T-Invest token. This closure neither activates a provider nor claims to inspect deployed secret stores or production telemetry.

Provider transport remains bounded: Go `net/http`, production REST base `https://invest-public-api.tbank.ru/rest`, 5-second timeout, redirects forbidden, cookie jar nil, response <=256 KiB, no automatic retry, internal <=60 provider requests/minute, maximum concurrency 4 with fail-fast overflow, request-lifetime provider processing only, no persistence/polling/cache.

Money normalization is exact: no floating-point conversion, canonical Decimal scale 8, mixed-sign `units`/`nano` rejected, unknown optional fields remain `nil`, unsupported dividend types fail closed, and dates never imply `PAID`/`CONFIRMED`.

For this adapter:

```text
AsOf        = OpenInvest observation/normalization time
RetrievedAt = the same OpenInvest observation time
```

This is not a provider update timestamp or freshness/SLA claim. Internal `SourceEventID` is an application-generated deterministic digest of normalized provider evidence, not a native T-Invest event ID and not a public API field.

## 5. Engineering review and remediation history

The history below uses the required sequence: **problem -> root cause -> failure/attack scenario -> project impact -> original design -> why review rejected it -> second-order scenario -> final remediation -> rationale -> regression coverage -> CI/review evidence -> residual limitations**. A second-order scenario is identified as such only when it was genuinely considered; no fictional review history is introduced.

### INT-3D-001 — dividend lifecycle/type mapping was too permissive

**Problem.** The first candidate ignored `Dividend.dividend_type`.

**Root cause.** A `GetDividends` row was initially treated as enough evidence for an ordinary announced cash dividend.

**Failure/attack scenario.** Provider evidence such as `Cancelled`, `Return of Capital`, `Daily Accrual`, or a future unknown type could be emitted as `DIVIDEND / ANNOUNCED`.

**Project impact.** OpenInvest could convert provider evidence into false event/financial truth and show a cancelled or semantically different event as expected cash income.

**Original design.** Every returned dividend row mapped to canonical `DIVIDEND / ANNOUNCED`.

**Why review rejected it.** T-Invest explicitly supplies type/lifecycle evidence; discarding it was an unjustified semantic upgrade.

**Second-order scenario.** A weaker fix of `Cancelled -> CANCELLED; everything else -> ANNOUNCED` would still silently reinterpret `Return of Capital`, `Daily Accrual`, and future unknown types.

**Final remediation.** `Regular Cash -> ANNOUNCED`, explicit `Cancelled -> CANCELLED`, all other/unknown types -> provider-data-invalid/fail closed. Dividend type also participates in deterministic identity.

**Why selected.** Conservative normalization is safer than inventing canonical meaning that the domain has not approved.

**Regression coverage.** Regular Cash, explicit Cancelled, unsupported provider types, and past payment dates that must not infer `PAID`.

**Evidence.** Internal finding `INT-3D-001` RESOLVED; Internal final `APPROVED`; External `APPROVED`; CI #499 and #501 green.

**Residual limitations.** Newly introduced provider dividend types intentionally fail closed until separately reviewed.

### INT-3D-002 — retrieval timestamp could precede actual retrieval

**Problem.** `RetrievedAt` could be sampled before the outbound provider response completed.

**Root cause.** The clock was sampled too early in the request lifecycle.

**Failure scenario.** A slow/blocking provider request could finish after the timestamp already recorded as retrieval time.

**Impact.** Freshness/audit metadata could claim observation earlier than the actual response acquisition.

**Original design.** Timestamp acquisition occurred before retrieval completed.

**Why review rejected it.** Audit timestamps must describe observation, not request initiation.

**Second-order scenario.** Merely moving the clock after request start but before decode would still allow malformed/slow response processing to be stamped too early.

**Final remediation.** For a non-empty decoded event batch, sample observation time after response decode and use it for both `AsOf` and `RetrievedAt` under the documented observation-time policy.

**Regression coverage.** A controlled test blocks the provider response and proves the clock is not sampled while retrieval is still blocked.

**Evidence.** `INT-3D-002` RESOLVED; final reviews APPROVED; exact-head CI green.

**Residual limitations.** These timestamps are OpenInvest observation time, not T-Invest publication/update time.

### INT-3D-003 — remote rate-limit evidence was ignored

**Problem.** The first candidate enforced the local 60/minute budget but did not use provider `x-ratelimit-remaining/reset` evidence.

**Root cause.** Local admission and provider-reported allowance were treated as unrelated controls.

**Failure scenario.** T-Invest could report the current allowance exhausted while OpenInvest continued sending calls permitted by its local counter.

**Impact.** Avoidable 429s, unnecessary provider pressure, and weaker compliance with the constrained traffic policy.

**Original design.** Local `<=60/minute` limiter only.

**Why review rejected it.** A stricter provider-side state must be honored when explicitly reported.

**Second-order scenario.** Naively trusting every later header could let an out-of-order response with a higher `remaining` value increase a previously observed lower allowance.

**Final remediation.** Fold provider remaining/reset evidence conservatively into admission; never raise the stricter local ceiling; stale/out-of-order higher values cannot increase a known lower allowance; reset evidence expires after the advertised interval; no retry/wait loop.

**Regression coverage.** Remote exhaustion/reset handling and stale higher-header non-relaxation.

**Evidence.** `INT-3D-003` RESOLVED; final reviews APPROVED; CI #499/#501 green.

**Residual limitations.** Limiting is provider-instance/process-local, not a distributed global quota coordinator.

### INT-3D-004 — mixed-sign MoneyValue could be silently reinterpreted

**Problem.** Positive `units` with negative `nano`, or the inverse, could be arithmetically combined.

**Root cause.** The initial conversion focused on numerical result rather than validating provider encoding consistency first.

**Failure scenario.** `units=1, nano=-100000000` could become a valid-looking decimal instead of invalid source evidence.

**Impact.** Malformed provider money could be silently transformed into a different canonical financial amount.

**Original design.** Arithmetic normalization before sign-consistency validation.

**Why review rejected it.** Financial normalization must preserve source truth, not repair contradictory encoding.

**Second-order scenario.** Checking only one sign direction would still leave the symmetric invalid form accepted.

**Final remediation.** Reject contradictory `units`/`nano` signs before exact Decimal conversion; preserve no-float scale-8 exactness checks.

**Regression coverage.** Both mixed-sign directions plus values requiring forbidden rounding.

**Evidence.** `INT-3D-004` RESOLVED; reviews APPROVED; exact-head CI green.

**Residual limitations.** Values not exactly representable at canonical scale 8 fail closed instead of being rounded.

### INT-3D-005 — optional provider facts were treated as required

**Problem.** Missing dates or amounts could become provider errors instead of canonical unknowns.

**Root cause.** The first normalization path conflated absent optional evidence with malformed present evidence.

**Failure scenario.** A valid partial dividend/coupon row could fail the whole provider request solely because a date or amount was absent.

**Impact.** Loss of valid source evidence and violation of the canonical rule that unknown values remain absent rather than fabricated.

**Original design.** Some optional source fields were effectively mandatory.

**Why review rejected it.** Unknown is a legitimate domain state and must not be converted to zero/date/error without evidence.

**Second-order scenario.** Converting absence to zero or substituting another date would avoid the error while creating false financial truth.

**Final remediation.** Absent optional dates/MoneyValue -> canonical `nil`; malformed present values -> provider-data-invalid.

**Regression coverage.** Missing dividend and coupon fields remain nil; malformed present values still fail closed.

**Evidence.** `INT-3D-005` RESOLVED; reviews APPROVED; CI green.

**Residual limitations.** Dated projections may omit undated events according to existing projection semantics; the adapter does not invent missing dates.

### INT-3D-006 — coupon identity accepted non-positive couponNumber

**Problem.** Zero/negative coupon numbers could participate in deterministic event identity.

**Root cause.** Identity presence was checked more weakly than semantic validity.

**Failure scenario.** Malformed provider coupon rows could generate ambiguous or unstable canonical identities.

**Impact.** Incorrect deduplication/identity behavior and acceptance of provider data outside the expected contract.

**Original design.** Non-positive values were not rejected strictly enough.

**Why review rejected it.** Coupon number is a provider identity component and must be positive before canonical identity is built.

**Second-order scenario.** Falling back to amount/date-only identity could collapse distinct schedule rows sharing economic fields.

**Final remediation.** Require `couponNumber > 0` and include it in deterministic coupon event identity.

**Regression coverage.** Zero and negative coupon numbers fail closed.

**Evidence.** `INT-3D-006` RESOLVED; reviews APPROVED; CI green.

**Residual limitations.** The adapter still relies on the reviewed provider field semantics; no provider discovery or secondary identity service is introduced.

### INT-3D-007 — active-call cap still allowed unbounded waiters

**Problem.** A semaphore limited active HTTP calls to four but callers could wait indefinitely for a slot.

**Root cause.** Concurrency control bounded network concurrency, not total in-flight demand/resource retention.

**Failure/attack scenario.** Thousands of callers could produce only four active provider calls but thousands of waiting goroutines.

**Impact.** Memory/back-pressure exhaustion and degraded service under burst or abuse traffic.

**Original design.** Blocking semaphore admission.

**Why review rejected it.** Bounding downstream concurrency while leaving the waiter population unbounded does not bound process resource use.

**Second-order scenario.** Adding a large fixed queue would merely move the exhaustion boundary and introduce latency/retry ambiguity not required by Feature 3D.

**Final remediation.** Admission is fail-fast when four provider calls are active; overflow returns provider-unavailable instead of queueing.

**Regression coverage.** Hold four provider calls active and prove additional calls fail immediately rather than wait.

**Evidence.** `INT-3D-007` RESOLVED; reviews APPROVED; race/CI checks green.

**Residual limitations.** The four-call gate is per provider/process instance, not a distributed cross-replica admission system.

### INT-3D-008 — empty protobuf-JSON response compatibility

**Problem.** Legitimate empty repeated fields could be classified as malformed provider data.

**Root cause.** Initial decoding expectations were stricter than protobuf JSON omission/null semantics.

**Failure scenario.** HTTP 200 body `{}` or a `null` repeated field could become a 502-style provider-data error instead of an empty result.

**Impact.** Valid empty source responses would be misreported as provider corruption.

**Original design.** Required explicit non-null repeated arrays.

**Why review rejected it.** Omitted/null repeated fields can represent an empty collection in the provider JSON contract.

**Second-order scenario.** Relaxing top-level JSON generally would risk accepting top-level `null`, arrays, trailing JSON values, or oversized content.

**Final remediation.** Accept omitted/null/empty `dividends` or `events` as empty results while still rejecting malformed JSON, top-level null/array, trailing values, and oversized responses.

**Regression coverage.** `{}`, null/empty repeated fields accepted; malformed/trailing/top-level-null/array/oversize rejected.

**Evidence.** `INT-3D-008` RESOLVED; reviews APPROVED; CI green.

**Residual limitations.** Compatibility is intentionally limited to the reviewed response envelope; schema expansion remains fail closed unless reviewed.

### INT-3D-009 — concurrent error result depended on goroutine completion order

**Problem.** A multi-instrument request could return different public error classes depending on which concurrent call finished first.

**Root cause.** First-observed error was effectively authoritative.

**Failure scenario.** One call returns provider-unavailable and another provider-data-invalid; reversed completion timing could switch the public result between the two classes.

**Impact.** Same provider facts could produce nondeterministic API behavior, making client handling and incident diagnosis unreliable.

**Original design.** Completion-order-sensitive aggregation.

**Why review rejected it.** Concurrency scheduling must not change externally observable semantics.

**Second-order scenario.** Cancelling all siblings on first failure would also make the final class depend on timing and discard already-authorized bounded evidence.

**Final remediation.** Complete the bounded authorized set and classify deterministically: caller context cancellation first; otherwise provider-data-invalid takes precedence over provider-unavailable.

**Regression coverage.** Reverse completion order while requiring the same final error class.

**Evidence.** `INT-3D-009` RESOLVED; reviews APPROVED; race/CI checks green.

**Residual limitations.** The precedence is intentionally conservative and specific to the current provider-neutral error classes.

### INT-3D-010 — evidence wording overclaimed some semantics

**Problem.** Documentation/evidence wording did not precisely match every runtime semantic, especially optional fields, rate headers, mixed-sign money, review-size arithmetic, `AsOf`, and internal `SourceEventID` ownership.

**Root cause.** Implementation and governance evidence evolved through multiple remediation cycles, leaving some text broader than the final code justified.

**Failure scenario.** Future reviewers could treat an application-generated digest as a provider-owned native ID or interpret `AsOf` as T-Invest source-update time.

**Impact.** Governance records could become technically misleading even while runtime code was correct.

**Original design.** The frozen implementation dossier attempted to summarize all final behavior in one prepublication record.

**Why review rejected it.** Evidence must not claim stronger provenance/time semantics than the implementation/provider contract proves.

**Second-order scenario.** Silently rewriting the historical evidence after publication would destroy chronology and obscure what was corrected later.

**Final remediation.** Publish an append-only evidence record followed by an append-only errata: `SourceEventID` is an application-generated deterministic digest; `AsOf = RetrievedAt` is OpenInvest observation time, not native provider update/freshness evidence.

**Regression/evidence coverage.** Cumulative evidence diff from implementation head was verified documentation-only with no runtime semantic drift; CI #501 passed all ten protected jobs.

**Evidence.** `INT-3D-010` RESOLVED; External verdict remained APPROVED; no-drift verification APPROVED.

**Residual limitations.** The historical implementation dossier intentionally remains immutable and must be read together with review evidence and errata.

## 6. PRE-3D-ROUTE-001 — shipped replay composition omitted the canonical route

**Problem.** The ordinary Corporate Actions route registry included `GET /api/v1/corporate-actions/projection`, but canonical shipped `NewReplay -> newReplayApp` did not register the endpoint.

**Root cause.** Route composition existed in more than one construction path and the shipped replay constructor lagged behind the canonical handler registry.

**Failure scenario.** Even with a valid provider injected, the real shipped surface could return 404 because the route itself was absent.

**Impact.** Feature 3D could appear implemented in unit/provider tests while remaining unreachable in the production composition.

**Original design.** Existing replay construction without provider-aware Corporate Actions wiring.

**Why it was insufficient.** Testing a provider or synthetic router does not prove the canonical application constructor exposes the already-approved endpoint.

**Second-order scenario.** Registering the route only when provider configuration is enabled would make disabled runtime return 404, conflating `route does not exist` with the established `source unavailable` state.

**Final remediation.** Preserve legacy constructors, add provider-aware replay/development constructors, always register the Corporate Actions route, inject a nil provider when disabled, and preserve existing fail-closed 503 `CORPORATE_ACTIONS_SOURCE_UNAVAILABLE` semantics.

**Why selected.** It fixes the actual composition root without changing the public URL/OpenAPI contract or forcing runtime activation.

**Regression coverage.** `TestReplayProductionConstructorRegistersCorporateActionsRouteWhenProviderIsDisabled` proves disabled provider -> 503, not 404; `TestReplayProductionConstructorInjectsCorporateActionsProvider` proves injected provider -> successful route.

**Evidence.** Included in PR #164, reviewed Internal/External, CI #499/#501 green.

**Residual limitations.** Provider availability still depends on explicit runtime configuration; the route being present is not evidence of source activation.

## 7. Controls that were correct without invented rejection history

Not every final control passed through a rejected design. Where the candidate was already correct, this closure records the threat/control directly rather than fabricating remediation chronology.

- Redirects are disabled because forwarding the Bearer token to an unexpected redirect target would create credential-leak risk. Regression coverage proves Authorization is not forwarded.
- Response bodies are capped at 256 KiB to bound memory exposure and oversized provider data fails closed.
- No automatic retry is used, avoiding multiplicative traffic and ambiguous provider-pressure behavior.
- Unknown mapped instruments fail before network I/O; no discovery method is called.
- Provider errors omit token/raw response payloads.
- No provider payload is persisted, cached cross-request, archived, logged as raw data, or sent to background workers.
- Provider event evidence never automatically mutates ledger truth.

## 8. Residual limitations and future gates

Feature 3D remains deliberately narrow:

1. Only `SBER`, `GAZP`, and `SU26238RMFS4` have reviewed static provider mappings.
2. Unsupported/future dividend types fail closed until semantics are reviewed.
3. Absent coupon amount remains unknown; no undocumented zero/sentinel meaning is inferred for future floating/variable coupons.
4. Rate/concurrency admission is provider-instance/process-local, not distributed across replicas.
5. No provider schedule date proves actual settlement; `PAID`/`CONFIRMED` are not inferred.
6. Provider evidence does not create ledger transactions, cash, realized income, or tax truth.
7. Runtime activation remains a separate operational/governance action requiring a valid read-only token and exact compliance with the existing `CONDITIONAL-GO` row.
8. Broader T-Invest methods, public-contract expansion, persistence, polling, caching, scraping/fallback, account/broker synchronization, trading, or price use require new review.
9. The public machine identifier remains `T_INVEST_API`; a future generic presentation/display-name layer may render a friendlier label without changing provider identity, but that is not part of this closure.
10. Production traffic/secret-store state is not asserted because this closure did not inspect production telemetry or deployed secrets.

## 9. Merge chronology accuracy

GitHub records `AsifAbbasov` as the actor for the Ready-for-review event at `2026-09-09T13:27:45Z` and for the merge event at `2026-09-09T13:54:40Z`. The assistant did not execute the merge mutation. A later chat message saying `разрешаю` occurred after GitHub had already merged PR #164, so this closure does **not** use that later message as retroactive pre-merge authorization evidence.

This chronology note exists to preserve factual governance history, not to reopen the already merged implementation.

## 10. Documentation synchronization set

This post-merge closure path is docs-only and targets exactly these repository surfaces:

1. `README.md`
2. `docs/SOURCE_OF_TRUTH.md`
3. `docs/ROADMAP.md`
4. `docs/DOCUMENT_INDEX.md`
5. `docs/IMPLEMENTATION_LOG.md`
6. `docs/VERSION_MATRIX.md`
7. `docs/OPEN_QUESTIONS.md`
8. `docs/CHANGELOG.md`
9. `docs/registries/DATA_SOURCE_REGISTRY.md`
10. `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_CLOSURE.md`

Historical Feature 3D implementation/evidence/errata files are not rewritten.

## 11. Closure acceptance conditions

This document becomes canonical only after the docs-only closure PR proves all of the following:

```text
changed paths = exactly the 10 documentation paths listed above
runtime/source/test/OpenAPI/config/workflow changes = 0
TINVEST_CORPORATE_ACTIONS_CONSTRAINED rights expansion = 0
runtime activation = NO
live token introduction = NO
Stage 3.78 start = NO
required exact-head CI = GREEN
Governance/Closure review = APPROVED
human merge authorization = separate future gate
```

Until that merge occurs, Feature 3D implementation itself remains canonical through PR #164, while this post-merge documentation closure remains a candidate.