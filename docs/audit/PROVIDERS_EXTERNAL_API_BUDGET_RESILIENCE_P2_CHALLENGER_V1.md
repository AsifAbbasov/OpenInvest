# Module #10 — Providers / External API / Budget Abuse / Resilience — P2 Challenger V1

## 1. Challenger identity

```text
TASK=PROVIDERS_EXTERNAL_API_BUDGET_RESILIENCE_P2_CHALLENGER_V1
MODULE=10
MODE=INDEPENDENT_CHALLENGER_FALSE_POSITIVE_RECONCILIATION
REMEDIATION_AUTHORIZED=NO
FORMAL_MODULE_CLOSED=NO
```

This challenger attempts to disprove, narrow, downgrade, or reclassify the two provisional Hostile Audit V1 findings. It does not modify production implementation, create a remediation PR, merge, or begin Module #11 or Module #12.

## 2. Frozen baseline and isolation

```text
PROTECTED_BRANCH=develop
BASE_SHA=af16e498fff348c5c70938f127cd164a84a227aa
BASE_TREE=e927460d8d5d668fed4255e8d3a8dfd519df5053
BASELINE_DRIFT=NO
BRANCH_PROTECTED=YES

CHALLENGER_BRANCH=audit/providers-external-api-budget-resilience-p2-challenger-v1
CHALLENGER_RUNTIME_HEAD=b7e1b45dd1ec56bff99569550112880c82f8082f
CHALLENGER_RUNTIME_TREE=8448d4185914d9c234e509dc638a3573974220c6

AUTHORITATIVE_RUN_ID=37619830891
AUTHORITATIVE_RUN_HEAD_SHA=b7e1b45dd1ec56bff99569550112880c82f8082f
AUTHORITATIVE_RUN_RESULT=SUCCESS
```

At the authoritative runtime head, the challenger differs from the frozen baseline only by:

```text
.github/workflows/module10-providers-p2-challenger.yml
backend-go/internal/provider/tinvest/module10_p2_challenger_test.go
```

The canonical challenger branch additionally contains this report. Production Go implementation, frontend implementation, migrations, deployment implementation, and normal production CI remain unchanged.

## 3. Actual shipped request path

Production routing registers:

```text
GET /api/v1/corporate-actions/projection
```

through the replay application router. The route is part of the public API operation set and the frontend calls it with credentials omitted.

Static composition plus routed runtime evidence establishes:

```text
AUTH_REQUIRED=NO
CSRF_REQUIRED=NO
OUTER_RATE_LIMITER_PRESENT=NO
OUTER_RATE_LIMITER_KEY=NOT_APPLICABLE
OUTER_RATE_LIMIT=NONE
PROVIDER_RATE_LIMIT=60_PER_MINUTE_PER_PROVIDER_PROCESS_OBJECT
PROVIDER_CONCURRENCY_LIMIT=4
ANY_CACHE=NO
ANY_SINGLEFLIGHT=NO
ANY_DEDUP=NO
ANY_EDGE_CONTRACT=YES_REPOSITORY_OWNERSHIP_ACKNOWLEDGEMENT_ONLY
```

The shipped replay middleware stack applies request lifecycle, sensitive-response cache policy, and local-development CORS where applicable. No endpoint-level or outer Fiber rate limiter is installed around the Corporate Actions projection route.

## 4. M10-P2-01 routed reproducer

The challenger used the production-equivalent shipped router constructor with the real T-Invest provider implementation pointed at a deterministic fake upstream.

Sequential anonymous campaign:

```text
VALID_ANONYMOUS_REQUESTS=60
HTTP_200=60
PROVIDER_CALLS_CREATED=60
AUTH_REQUIRED=NO
CSRF_REQUIRED=NO
OUTER_RATE_LIMITER=NO
```

Immediately after consuming the process-local provider budget:

```text
LEGITIMATE_REQUEST_AFTER_EXHAUSTION_STATUS=503
LEGITIMATE_REQUEST_AFTER_EXHAUSTION_PROVIDER_CALL=NO
```

The provider budget rejects the request before another upstream call is made, but the legitimate routed request is unavailable.

Using the same production requestBudget implementation with a controllable audit clock, the next request after the one-minute rolling window succeeded:

```text
RECOVERY_STATUS=200
RECOVERY_TIME=60_SECONDS_BUDGET_WINDOW
RECOVERY_REQUIRES_PROCESS_RESTART=NO
```

Concurrent routed campaign:

```text
CLIENTS=10
HTTP_200=4
HTTP_503=6
UPSTREAM_REQUESTS=4
PROVIDER_CONCURRENCY_LIMIT=4
```

The concurrency ceiling correctly bounds instantaneous provider work. It does not prevent a low-cost sequential caller from consuming the full minute budget.

## 5. Independent multi-instance reproduction

Against one shared deterministic fake provider, separate routed application/provider instances admitted:

```text
ONE_INSTANCE_ADMITTED=60
TWO_INSTANCE_ADMITTED=120
FOUR_INSTANCE_ADMITTED=240
PROCESS_LOCAL_MULTIPLICATION=YES
```

The test used a separate production T-Invest provider object per routed application instance. Each object owns its own requestBudget.

Repository deployment contract:

```text
OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER=verified-shared-provider-budget-v1
REPOSITORY_SIDE_PROVIDER_BUDGET_CONTRACT=VERIFIED
```

This is an ownership acknowledgement and startup gate only. It is not runtime proof of a real shared limiter.

```text
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES
```

## 6. M10-P2-01 severity reconciliation

The challenger actively attempted to downgrade or reject the original P2.

Evidence against downgrade:

- the endpoint is anonymously reachable through the shipped routing stack;
- no outer application rate limiter isolates an anonymous caller before provider budget consumption;
- 60 syntactically valid anonymous requests consumed all 60 process-local provider admissions;
- the next legitimate provider-backed request received HTTP 503;
- the attack requires no authenticated account, no CSRF token, no durable-state setup, and no expensive client-side work;
- the process-local budget independently multiplies 60 -> 120 -> 240 across 1/2/4 instances;
- the repository contract does not prove actual deployment-global enforcement;
- sustained traffic can repeatedly recreate the denial window.

Evidence limiting severity:

- no financial mutation or durable corruption occurs;
- concurrency is bounded to four active upstream requests per process;
- over-budget requests fail fast rather than queue indefinitely;
- the local budget recovers automatically after the rolling one-minute window;
- exact production monetary/provider-account cost was not measured.

Independent result:

```text
M10_P2_01_REPRODUCED=YES
M10_P2_01_FINAL_SEVERITY=P2
M10_P2_01_FINAL_STATUS=CONFIRMED
```

The finding remains P2 because the routed attack is low-cost, anonymous, deterministic, and produces a real legitimate-user availability denial. The absence of durable corruption prevents escalation above P2.

## 7. M10-P3-01 conflicting duplicate-key matrix

Hostile Audit V1 only established that identical duplicate keys were accepted. The challenger used conflicting values and both key orders.

### 7.1 Different valid values

```text
FIRST=amount_1_date_A
SECOND=amount_2_date_B
RESULT=amount_2_date_B

FIRST=amount_2_date_B
SECOND=amount_1_date_A
RESULT=amount_1_date_A
```

Observed decoder behavior:

```text
DUPLICATE_KEY_DECODER_SEMANTICS=LAST_VALUE_WINS
```

Changing duplicate order changed the canonical event and routed projection output.

### 7.2 Valid / invalid order

```text
FIRST_VALID_SECOND_INVALID_STATUS=502
FIRST_INVALID_SECOND_VALID_STATUS=200
```

The order therefore changes the validation result.

### 7.3 Large / small order

Using one duplicate value containing 100 distinct valid dividends and one containing a single valid dividend:

```text
LARGE_FIRST_SMALL_SECOND_ACCEPTED_EVENTS=1
SMALL_FIRST_LARGE_SECOND_ACCEPTED_EVENTS=100
```

The order therefore changes accepted event count and projection semantics.

### 7.4 Persistence boundary

The Corporate Actions endpoint remains a request-lifetime read-only projection path:

```text
NO_PROVIDER_PERSISTENCE=YES
NO_FINANCIAL_DB_MUTATION=YES
PERSISTENT_FINANCIAL_WRITE_FROM_DUPLICATE_KEYS=NO
```

## 8. M10-P3-01 severity reconciliation

Behavioral reproduction is real:

```text
DUPLICATE_KEYS_ACCEPTED=YES
CONFLICTING_DUPLICATE_KEYS_TESTED=YES
DUPLICATE_KEY_ORDER_CHANGES_ACCEPTED_RESULT=YES
DUPLICATE_KEY_ORDER_CHANGES_PROJECTION=YES
M10_P3_01_REPRODUCED=YES
```

However, the challenger does not retain it as a P3 security vulnerability.

Reasons:

- Go decoder semantics are deterministic: the last duplicate value wins;
- the duplicate values originate from the configured external provider trust boundary, not from an unauthenticated API caller;
- a compromised or malicious provider already has the ability to send an arbitrary single valid value, so duplicate-key ordering does not grant a new authority beyond that existing provider trust;
- no provider payload is persisted;
- no financial database mutation occurs;
- the effect is confined to the current request's projection result;
- canonical validation still rejects the final effective invalid value.

Independent reconciliation:

```text
M10_P3_01_FINAL_SEVERITY=HARDENING_ONLY
M10_P3_01_FINAL_STATUS=DOWNGRADED_NOT_COUNTED_AS_SECURITY_FINDING
```

Rejecting duplicate JSON keys remains reasonable parser hardening, especially to avoid cross-parser ambiguity if architecture changes later, but current evidence does not support counting this behavior as a P3 security finding.

## 9. Slow / hanging provider independent reconfirmation

The challenger independently exercised:

```text
never sends headers
partial body then hangs
explicit caller cancellation
```

For every case a raw TCP witness, independent from the client return value, observed connection closure.

```text
CALLER_DEADLINE_HONORED=YES
UPSTREAM_CONNECTION_CLOSED=YES
PROVIDER_SLOT_RELEASED=YES
RECOVERY_PASS=YES
```

The same provider object successfully completed a subsequent request after each cancellation case.

```text
SLOW_PROVIDER_HANG=REJECTED_AS_CURRENT_PRODUCT_FINDING
```

## 10. Not-verified items preserved

No missing evidence was silently converted to PASS.

```text
NV_M10_01=EXTERNAL_SHARED_PROVIDER_BUDGET_NOT_VERIFIED_CARRY_FORWARD_MODULE_12
NV_M10_02=FULL_API_PROCESS_MULTI_INSTANCE_CHAOS_NOT_VERIFIED
NV_M10_03=EXHAUSTIVE_REAL_SOCKET_RESOURCE_MATRIX_NOT_VERIFIED
```

Provider-object multi-instance tests do not substitute for real API-process/container lifecycle chaos. The compact raw-TCP cancellation reconfirmation does not substitute for the exhaustive 100/1000/10000 x timeout/reset/malformed/cancellation resource matrix.

## 11. Test defect control

Hostile Audit V1 identified TD-M10-01, where passive httptest handler-context observation was not a valid sole cancellation oracle. This challenger does not reuse that defect.

```text
TD_M10_01=PREVIOUSLY_CORRECTED
NEW_TEST_DEFECTS=0
PRODUCT_FINDING_FROM_TEST_DEFECT=NO
```

Raw TCP closure is the independent cancellation witness.

## 12. Authoritative workflow execution

```text
RUN_ID=37619830891
HEAD_SHA=b7e1b45dd1ec56bff99569550112880c82f8082f
RESULT=SUCCESS
```

Independent jobs:

```text
Challenger / routed-budget-exhaustion=SUCCESS
Challenger / routed-distributed-budget=SUCCESS
Challenger / conflicting-duplicate-keys=SUCCESS
Challenger / slow-provider-raw-tcp=SUCCESS
Final independent challenger enforcement=SUCCESS
```

The matrix uses fail-fast false and evidence upload on every challenger job.

## 13. Production immutability

At the authoritative runtime head, exact diff from frozen baseline contains only:

```text
.github/workflows/module10-providers-p2-challenger.yml
backend-go/internal/provider/tinvest/module10_p2_challenger_test.go
```

The canonical branch adds only this report.

```text
PRODUCTION_CODE_CHANGED=NO
FINAL_REMEDIATION_COMMIT=NO
PR_CREATED=NO
MERGE_PERFORMED=NO
```

## 14. Canonical reconciliation

```text
M10_P2_01=CONFIRMED_P2_PUBLIC_PROVIDER_BUDGET_EXHAUSTION
M10_P3_01=DOWNGRADED_TO_HARDENING_ONLY_DUPLICATE_JSON_KEY_ACCEPTANCE

FINAL_CONFIRMED_P0=0
FINAL_CONFIRMED_P1=0
FINAL_CONFIRMED_P2=1
FINAL_CONFIRMED_P3=0

FALSE_POSITIVES_REJECTED=0
FINDINGS_DOWNGRADED=1
FINDINGS_UPGRADED=0
NEW_CONFIRMED_FINDINGS=0

READY_FOR_INDEPENDENT_CHALLENGER_REVIEW=YES
FORMAL_MODULE_CLOSED=NO
REPOSITORY_WIDE_AUDIT=ONGOING
NEXT_ACTION=STOP_FOR_INDEPENDENT_CHALLENGER_REVIEW
```

## 15. Stop gate

No remediation was performed.

No remediation PR was created.

Nothing was merged.

Module #11 and Module #12 were not started.
