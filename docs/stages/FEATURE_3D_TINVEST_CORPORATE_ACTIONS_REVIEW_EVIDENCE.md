# Feature 3D — T-Invest Corporate Actions Review Evidence Publication

| Field | Value |
| --- | --- |
| Status | Evidence-only follow-up publication after External published-head verdict |
| Date | 2026-09-09 |
| Repository | `AsifAbbasov/OpenInvest` |
| Branch | `feature/tinvest-corporate-actions` |
| Canonical base | `develop@e4e676f8a168bf1a2618185ee345cab245d1225d` |
| Implementation head reviewed by External phase | `7c021e74db3a66488c4c6d87412729f524d9023f` |
| Implementation tree | `9ae3ea97a7bc1d1ce07879cee493dfe721d84eb4` |
| Draft PR | `#164` |
| Exact-head implementation CI | Run `#499` / `34353668492` — completed successfully |
| Runtime activation | NO |
| Live T-Invest token | NOT USED |
| Stage 3.78 | NOT STARTED |

## 1. Purpose

This file is the mandatory post-External publication of the Internal review evidence that was intentionally withheld from the repository and Draft PR before the fresh External published-head verdict, as required by `docs/REVIEW_WORKFLOW.md` v1.4.0.

The original implementation dossier remains the historical prepublication freeze record. This follow-up does not rewrite that earlier temporal state. It appends the review chronology after the External phase has completed.

This commit is evidence-only. It must not change runtime code, tests, configuration, OpenAPI, dependencies, migrations, CI workflows, source-rights registry rows, frontend behavior, or provider semantics.

## 2. Reviewed implementation subject

The Internal phase ultimately reviewed the frozen 10-file Feature 3D candidate that became implementation commit:

```text
7c021e74db3a66488c4c6d87412729f524d9023f
```

with tree:

```text
9ae3ea97a7bc1d1ce07879cee493dfe721d84eb4
```

and direct parent / canonical base:

```text
e4e676f8a168bf1a2618185ee345cab245d1225d
```

The complete reviewed file set was:

1. `.env.example`
2. `backend-go/cmd/api/main.go`
3. `backend-go/cmd/api/tinvest_runtime.go`
4. `backend-go/cmd/api/tinvest_runtime_test.go`
5. `backend-go/internal/httpapi/replay_app.go`
6. `backend-go/internal/httpapi/replay_app_corporateactions_test.go`
7. `backend-go/internal/provider/tinvest/provider.go`
8. `backend-go/internal/provider/tinvest/provider_test.go`
9. `backend-go/internal/provider/tinvest/types.go`
10. `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_IMPLEMENTATION.md`

## 3. Internal review independence and edit boundary

The Internal review phase was read-only. The reviewer did not edit files, run auto-fixes, stage changes, create commits, push branches, or silently resolve findings.

When the Internal phase returned blocking findings, Builder work resumed separately, applied the remediation, reran affected local checks, and returned the complete revised candidate for another read-only review cycle.

The final Internal verdict was produced only after the candidate was frozen and no further edits were made during that final review.

## 4. Internal findings and Builder remediation chronology

### INT-3D-001 — dividend lifecycle/type mapping was too permissive

**Severity:** Blocking

**Finding:** the first candidate ignored `Dividend.dividend_type`, so provider rows such as `Cancelled`, `Return of Capital`, or `Daily Accrual` could be emitted as ordinary `ANNOUNCED` cash dividends.

**Risk:** provider semantics could be converted into false canonical financial/event truth.

**Builder remediation:** added explicit dividend-type handling:

```text
Regular Cash → DIVIDEND / ANNOUNCED
Cancelled    → DIVIDEND / CANCELLED
other/unknown provider types → provider-data-invalid / fail closed
```

Dividend type was also included in deterministic event identity. Regression tests cover explicit cancellation and unsupported types.

**Resolution:** RESOLVED.

### INT-3D-002 — retrieval timestamp was sampled before source retrieval completed

**Severity:** Blocking

**Finding:** `RetrievedAt` could be sampled before the outbound provider request/response completed.

**Risk:** freshness/audit evidence could state an observation time earlier than the actual retrieval.

**Builder remediation:** retrieval time is sampled after response decode for a non-empty provider event batch. A controlled regression test proves the clock is not sampled while the provider response is still blocked.

**Resolution:** RESOLVED.

### INT-3D-003 — provider rate-limit response evidence was ignored

**Severity:** Blocking

**Finding:** the first candidate enforced only the local OpenInvest request budget and did not fold T-Invest `x-ratelimit-remaining` / `x-ratelimit-reset` evidence into subsequent admission.

**Risk:** OpenInvest could continue sending requests after the provider had reported that the current allowance was exhausted.

**Builder remediation:** the request budget now observes provider remaining/reset headers conservatively, never increases the stricter local 60/minute limit, and does not allow stale/out-of-order higher remaining values to raise a known lower allowance. No wait/retry loop was introduced.

**Resolution:** RESOLVED.

### INT-3D-004 — mixed-sign MoneyValue components could be arithmetically normalized

**Severity:** Blocking

**Finding:** positive `units` with negative `nano`, or negative `units` with positive `nano`, could be arithmetically combined instead of rejected.

**Risk:** invalid provider money encoding could be silently transformed into a different canonical amount.

**Builder remediation:** contradictory component signs now fail closed as provider-data-invalid before exact decimal conversion. Regression tests cover both sign directions.

**Resolution:** RESOLVED.

### INT-3D-005 — optional provider fields were incorrectly treated as required

**Severity:** Blocking

**Finding:** absent dividend/coupon dates or amounts could become provider errors rather than canonical unknown values.

**Risk:** valid partial provider evidence could be rejected, and the canonical Stage 3.61/3.62 rule that unknown values remain absent could be violated.

**Builder remediation:** optional source timestamps and `MoneyValue` now normalize to canonical `nil`; malformed present values still fail closed. Undated evidence remains valid at the event boundary and is omitted only by dated projections according to existing projection rules.

**Resolution:** RESOLVED.

### INT-3D-006 — coupon identity accepted non-positive couponNumber

**Severity:** Blocking

**Finding:** missing/zero/negative `couponNumber` could participate in event identity.

**Risk:** malformed provider rows could produce unstable or ambiguous canonical event identity.

**Builder remediation:** coupon number must be strictly positive. Tests cover zero and negative values.

**Resolution:** RESOLVED.

### INT-3D-007 — concurrency semaphore allowed an unbounded waiter population

**Severity:** Blocking

**Finding:** callers could block waiting for a provider semaphore slot, bounding active network calls but not bounding waiting goroutines under load.

**Risk:** resource exhaustion / back-pressure failure during concurrent traffic.

**Builder remediation:** semaphore admission is now fail-fast when four provider calls are active. A regression test holds four active requests and proves additional callers fail provider-unavailable instead of queueing.

**Resolution:** RESOLVED.

### INT-3D-008 — empty protobuf/JSON repeated fields compatibility

**Severity:** Blocking

**Finding:** a legitimate provider `200` response such as `{}` with an omitted repeated field could be interpreted as invalid rather than as an empty result.

**Risk:** protobuf/JSON-compatible empty source responses could incorrectly become 502 provider-data failures.

**Builder remediation:** omitted or `null` repeated `dividends` / `events` fields are accepted as empty results; malformed top-level JSON, arrays, trailing values, and oversized bodies remain rejected. Regression tests cover all of these cases.

**Resolution:** RESOLVED.

### INT-3D-009 — concurrent batch error class depended on goroutine completion order

**Severity:** Blocking

**Finding:** when different authorized instrument calls concurrently returned provider-data and provider-unavailable failures, the public class could depend on which goroutine completed first.

**Risk:** identical provider facts could yield nondeterministic 502/503 behavior.

**Builder remediation:** the bounded approved request set is collected and classified with fixed precedence independent of completion order: caller context cancellation first; otherwise provider-data-invalid takes precedence over provider-unavailable. Regression tests reverse completion order and require the same result.

**Resolution:** RESOLVED.

### INT-3D-010 — implementation-evidence wording required precision

**Severity:** Blocking documentation/evidence accuracy

**Finding:** several dossier statements required tighter wording so evidence did not overclaim implementation behavior: optional fields, mixed-sign money, provider rate headers, review-size arithmetic, `AsOf` semantics, and generated `SourceEventID` semantics.

**Risk:** governance evidence could be technically misleading even when runtime behavior was correct.

**Builder remediation:** the frozen implementation dossier was synchronized to the actual candidate semantics, including exact `754` non-blank/non-comment provider business-logic lines (`provider.go = 687`, `types.go = 67`), observation-time `AsOf = RetrievedAt` policy for this adapter, and deterministic application-generated source-event digest semantics.

**Resolution:** RESOLVED.

## 5. Final Internal review record

After all Builder remediation and affected local checks, the complete 10-file candidate was frozen and reread in a final read-only Internal phase.

Final blocking state:

```text
P0 = 0
P1 = 0
P2 blocking = 0
```

Remaining blocking findings: **NONE**.

Final Internal verdict:

```text
APPROVED
```

The reviewer made no edits during the final review.

## 6. Prepublication local evidence

The environment could not honestly execute the complete repository under the canonical Go 1.25.14 toolchain before publication, so full-repository success was not claimed at the precommit gate.

Available focused evidence passed for the frozen candidate, including:

```text
provider: go test -race -count=1 ./internal/provider/tinvest   PASS
provider: go vet ./internal/provider/tinvest                  PASS
runtime:  go test -race -count=1 ./cmd/api                    PASS
runtime:  go vet ./cmd/api                                    PASS
gofmt / syntax checks                                         PASS
forbidden T-Invest production-surface scan                    PASS
persistence / polling / retry mechanism scans                 PASS
```

## 7. Published implementation and authoritative CI evidence

The approved frozen candidate was published as one implementation commit:

```text
7c021e74db3a66488c4c6d87412729f524d9023f
feat: add constrained T-Invest corporate actions provider
```

Draft PR `#164` targeted `develop@e4e676f8a168bf1a2618185ee345cab245d1225d`.

GitHub Actions CI run `#499` (`34353668492`) executed on exact head `7c021e74db3a66488c4c6d87412729f524d9023f` and completed successfully. All ten required protected checks passed:

1. Go tests
2. Python tests
3. Frontend build and typecheck
4. OpenAPI contract
5. Docker Compose config
6. PostgreSQL migration validation
7. Go vet
8. Go race tests
9. Go vulnerability scan
10. Dependency security scan

This exact-head CI is authoritative repository evidence and supersedes the earlier prepublication limitation for the implementation head.

## 8. External published-head review handoff

After exact-head CI was green, the designated review chat performed a fresh read-only External published-head review of PR `#164` at implementation SHA `7c021e74db3a66488c4c6d87412729f524d9023f`.

That phase independently re-evaluated the published code, canonical repository contracts, source-rights boundary, provider transport/security behavior, financial semantics, performance/resilience controls, API compatibility, scope, documentation, and CI evidence.

External blocking findings:

```text
0
```

External verdict:

```text
APPROVED
```

One non-blocking presentation note remained: the canonical machine provider identifier `T_INVEST_API` may eventually receive a generic UI display-name mapping such as `T-Invest API`; this is not an OpenAPI/runtime correctness blocker.

## 9. Evidence-only publication boundary

This file is the post-External publication required by the development workflow. It records the previously withheld Internal evidence without changing the reviewed implementation semantics.

The follow-up commit must be verified as documentation/evidence-only against implementation head `7c021e74db3a66488c4c6d87412729f524d9023f`.

Required verification after publication:

```text
changed implementation/runtime files = 0
changed documentation/evidence files  = this file only
provider/runtime tree semantics        = unchanged
required CI on evidence-follow-up head = GREEN
```

Until those conditions are verified, this evidence publication is not the final merge gate.

## 10. Remaining governance gates

After this evidence-only commit is published:

```text
required CI on evidence-follow-up head
→ exact diff / no-semantic-drift verification
→ human review
→ separate explicit Ready authorization if required
→ separate explicit squash-merge authorization
```

This evidence record does **not** authorize:

- Ready transition;
- merge or auto-merge;
- direct mutation of `develop` or `main`;
- runtime activation;
- creation/use of a live T-Invest token;
- Stage 3.78;
- broader T-Invest methods or source/use rights.

Current activation state remains:

```text
RUNTIME ACTIVATION = NO
LIVE T-INVEST TOKEN = NOT USED
STAGE 3.78 = NOT STARTED
```
