# Module #10 — Providers / External API / Budget Abuse / Resilience — Adversarial Audit V1

## 1. Audit identity

```text
TASK=PROVIDERS_EXTERNAL_API_BUDGET_ABUSE_RESILIENCE_HOSTILE_AUDIT_V1
MODULE=10
CYBERSECURITY=YES
AUDIT_MODE=HOSTILE
FORMAL_MODULE_CLOSED=NO
REMEDIATION_AUTHORIZED=NO
```

This report records hostile repository/runtime evidence only. It does not remediate confirmed findings, does not create a remediation PR, and does not close Module #10.

## 2. Frozen protected baseline

```text
PROTECTED_BRANCH=develop
BASE_SHA=af16e498fff348c5c70938f127cd164a84a227aa
BASE_TREE=e927460d8d5d668fed4255e8d3a8dfd519df5053
BRANCH_PROTECTED=YES

AUDIT_BRANCH=audit/providers-external-api-budget-resilience-v1
AUTHORITATIVE_RUNTIME_AUDIT_HEAD=f995d3f97d8c5e203aef6254e7e19162da6ae2b6
AUTHORITATIVE_RUNTIME_AUDIT_TREE=c19a86c88eb8ad81957427fbe9d41f9e20b22dd0
AUTHORITATIVE_RUNTIME_RUN=37613291011
AUTHORITATIVE_RUNTIME_RUN_RESULT=SUCCESS
```

The authoritative hostile run executed on the isolated audit branch. Before this report commit, the audit branch differed from the protected baseline only by audit-only test and workflow files:

```text
.github/workflows/module10-providers-hostile-audit.yml
backend-go/internal/httpapi/module10_provider_endpoint_audit_test.go
backend-go/internal/provider/moexiss/module10_adversarial_audit_test.go
backend-go/internal/provider/tinvest/module10_adversarial_audit_test.go
```

No production implementation file was modified.

## 3. Provider and externally reachable surface inventory

### 3.1 T-Invest Corporate Actions

Active-capable adapter:

```text
PROVIDER=T_INVEST_API
IMPLEMENTATION=backend-go/internal/provider/tinvest/provider.go
PRODUCTION_BASE=https://invest-public-api.tbank.ru/rest
METHODS=GetDividends,GetBondCoupons
TOTAL_HTTP_TIMEOUT=5s
RESPONSE_BODY_CAP=256KiB
PROCESS_LOCAL_REQUEST_BUDGET=60/minute
PROCESS_LOCAL_MAX_CONCURRENCY=4
AUTOMATIC_RETRY=NO
REDIRECT_FOLLOWING=NO
COOKIE_JAR=DISABLED
```

Runtime activation is controlled by:

```text
OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED
OPENINVEST_TINVEST_READONLY_TOKEN
OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER
```

Production/staging activation fails closed unless the deployment ownership acknowledgement equals:

```text
verified-shared-provider-budget-v1
```

That repository-side contract is not proof that a real shared external limiter exists in a deployed multi-replica environment.

### 3.2 Public provider-backed HTTP endpoint

```text
GET /api/v1/corporate-actions/projection
AUTHENTICATION_REQUIRED=NO
FRONTEND_CREDENTIALS=omit
HANDLER_CACHE=NONE
HANDLER_REQUEST_COALESCING=NONE
HANDLER_DEDUP=NONE
```

The handler validates the query before provider invocation and then performs request-lifetime provider work. No Corporate Actions provider persistence path was found.

### 3.3 MOEX ISS quote adapter

```text
PROVIDER=MOEX_ISS
IMPLEMENTATION=backend-go/internal/provider/moexiss/provider.go
PRODUCTION_BASE=https://iss.moex.com
TOTAL_HTTP_TIMEOUT=5s
RESPONSE_BODY_CAP=64KiB
REDIRECT_FOLLOWING=NO
SHIPPED_RUNTIME_ACTIVATION=NO
STATUS=DORMANT
```

Repository composition remains provider-free for MOEX; no user-triggered shipped MOEX endpoint was discovered.

## 4. Authoritative hostile execution model

Workflow:

```text
.github/workflows/module10-providers-hostile-audit.yml
```

Execution properties:

```text
TINVEST_SCENARIOS=MATRIX
MATRIX_FAIL_FAST=FALSE
INDIVIDUAL_JOB_TIMEOUT=4m
GO_TEST_TIMEOUT=3m
ARTIFACT_UPLOAD=if:always()
FINAL_ENFORCEMENT=AFTER_ALL_JOBS
SUBTESTS_SKIPPED_DUE_TO_EARLY_FAILURE=0
```

Authoritative run:

```text
RUN_ID=37613291011
EVENT=push
BRANCH=audit/providers-external-api-budget-resilience-v1
HEAD_SHA=f995d3f97d8c5e203aef6254e7e19162da6ae2b6
RESULT=SUCCESS

TINVEST_RESULT=success
ENDPOINT_RESULT=success
MOEX_RESULT=success
RUNTIME_RESULT=success
FINAL_ENFORCEMENT=success
```

A green hostile run means the challenger executed as designed. It does not mean that every measured product behavior is acceptable; confirmed findings below are intentionally asserted by successful tests that prove the vulnerable or residual behavior deterministically.

## 5. Slow / hanging upstream evidence

Covered runtime cases:

```text
upstream accepts request and never sends headers
headers sent but body never completes
very slow chunked response
connection reset mid-body
TCP refusal
deterministic DNS-like resolution failure
HTTP 429
HTTP 500
HTTP 502
HTTP 503
HTTP 504
malformed JSON
truncated JSON
oversized JSON
valid JSON with adversarial field sizes
```

Measured results:

```text
SLOW_PROVIDER_HANG_TEST=PASS
CALLER_DEADLINE_HONORED=YES
UPSTREAM_REQUEST_CANCELLED=YES
CONNECTION_RELEASED=YES
PROVIDER_CONCURRENCY_SLOT_RELEASED=YES
NO_GOROUTINE_LEAK=YES
RECOVERY_PASS=YES
```

Raw TCP cancellation witnesses independently proved connection closure for both no-header and partial-body cases at the caller deadline:

```text
MODULE10_RAW_TCP case=never_headers caller_deadline=true upstream_connection_closed=true
MODULE10_RAW_TCP case=partial_body caller_deadline=true upstream_connection_closed=true
```

The ordinary slow-provider challenger additionally proved that request permits are released for no-header, incomplete-body, slow-chunked, connection-reset, TCP-refusal, and DNS-like failures.

## 6. Distributed provider budget

The provider request gate is process-local. A deterministic shared fake upstream measured aggregate admission:

```text
EXPECTED_PROVIDER_BUDGET_PER_PROCESS=60/minute
OBSERVED_1_INSTANCE=60
OBSERVED_2_INSTANCES=120
OBSERVED_4_INSTANCES=240
BUDGET_MULTIPLICATION=PROCESS_LOCAL_MULTIPLICATION
```

Classification:

```text
REPOSITORY_SIDE_CONTRACT=VERIFIED
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES
```

The runtime activation gate correctly rejects production-like activation without the exact shared-budget-owner acknowledgement. That does not prove an actual shared edge/gateway or distributed limiter exists.

## 7. CONFIRMED FINDINGS

### M10-P2-01 — Public provider endpoint can deterministically exhaust provider budget

```text
SEVERITY=P2
CLASS=CONFIRMED_FINDING
AFFECTED_ENDPOINT=GET /api/v1/corporate-actions/projection
ATTACKER_AUTHENTICATION_REQUIRED=NO
DETERMINISTIC_REPRODUCER=YES
REALISTIC_PRECONDITION=public access plus syntactically valid supported corporate-action query
```

Runtime evidence:

```text
VALID_PUBLIC_REQUESTS=30
PROVIDER_CALLS_CREATED=30
ENDPOINT_CACHE=NONE
ENDPOINT_DEDUP=NONE

IDENTICAL_CONCURRENT_CLIENTS=10
UPSTREAM_CALLS_CREATED=4

IDENTICAL_CONCURRENT_CLIENTS=50
UPSTREAM_CALLS_CREATED=4

IDENTICAL_CONCURRENT_CLIENTS=100
UPSTREAM_CALLS_CREATED=4

DEAD_PROVIDER_INCOMING_REQUESTS=100
UPSTREAM_ATTEMPTS=60

PROCESS_LOCAL_BUDGET_1_INSTANCE=60
PROCESS_LOCAL_BUDGET_2_INSTANCES=120
PROCESS_LOCAL_BUDGET_4_INSTANCES=240
```

Invalid requests tested by the endpoint challenger were rejected before provider invocation, which is good, but syntactically valid anonymous requests consume provider budget one-for-one until the provider-level controls intervene.

Impact:

- an unauthenticated caller can consume the provider request allowance without authenticated portfolio/user work;
- identical requests are not coalesced;
- the process-local budget multiplies across replicas;
- legitimate users can be forced into provider-unavailable/rate-limited behavior after the available allowance is spent;
- exact monetary/provider-account cost was not measured and is not claimed.

Existing mitigations reduce blast radius but do not eliminate the issue:

- maximum four active provider calls per process;
- 60 requests/minute process-local gate;
- external shared-budget ownership is required by deployment contract;
- malformed/invalid requests are rejected before provider execution.

No remediation was performed during this hostile phase.

### M10-P3-01 — Duplicate JSON object keys are accepted by provider response decoding

```text
SEVERITY=P3
CLASS=CONFIRMED_FINDING
DETERMINISTIC_REPRODUCER=YES
PERSISTENT_FINANCIAL_WRITE=NO
BODY_SIZE_BOUND=YES
```

The hostile provider response:

```json
{"dividends":[],"dividends":[]}
```

is accepted by the current JSON decoder. This creates avoidable ambiguity when upstream evidence contains duplicate keys.

Severity remains P3 because:

- response size is bounded;
- canonical event validation remains fail-closed for invalid dates, money, enums, and malformed structures;
- events are deduplicated after canonicalization;
- this provider path produces request-lifetime projections and does not persist provider-owned financial state.

Unknown additive fields were also tolerated in bounded 1 KiB and 100 KiB cases. That behavior is recorded as compatibility behavior, not independently classified as a vulnerability.

## 8. Provider endpoint validation and amplification

Endpoint challenger evidence:

```text
REQUESTS_SENT_VALID=30
UPSTREAM_PROVIDER_CALLS_CREATED=30
VALIDATION_REJECTIONS_TESTED=4
VALIDATION_REJECTION_BEFORE_PROVIDER_CALL=YES
AUTH_REJECTION_BEFORE_PROVIDER_CALL=NOT_APPLICABLE_ENDPOINT_IS_PUBLIC
CACHE_HIT=0
DEDUP_HIT=0
PROVIDER_COST_AMPLIFICATION=ONE_PROVIDER_CALL_PER_VALID_SEQUENTIAL_REQUEST
```

Fresh request identifiers/headers do not create a durable-state issue because the endpoint is a GET and does not use idempotency persistence. Query variation can, however, continue to consume provider budget as long as the query remains valid.

## 9. Concurrency exhaustion

Tested:

```text
C=1
C=4
C=5
C=8
C=40
CONFIGURED_LIMIT=4
```

Measured maximum active upstream calls never exceeded four.

```text
MAX_ACTIVE_UPSTREAM_REQUESTS<=CONFIGURED_LIMIT=PASS
NO_UNBOUNDED_WAIT_QUEUE=PASS
FAIL_FAST_OR_BOUNDED_WAIT=PASS
ALL_PERMITS_RELEASE_AFTER_TIMEOUT=PASS
ALL_PERMITS_RELEASE_AFTER_CANCEL=PASS
ALL_PERMITS_RELEASE_AFTER_ERROR=PASS
RECOVERY_TO_FULL_CAPACITY=PASS
```

## 10. Retry / backoff

The provider has no automatic retry layer.

Runtime status tests for 429/500/502/503/504 each observed exactly one upstream attempt.

```text
RETRY_COUNT_BOUNDED=YES
AUTOMATIC_RETRY=NO
NO_RETRY_STORM=YES
NO_RETRY_AFTER_CONTEXT_CANCEL=YES
NO_MULTIPLICATIVE_RETRIES_ACROSS_LAYERS=YES
BACKOFF=NOT_APPLICABLE_NO_RETRY
JITTER=NOT_APPLICABLE_NO_RETRY
```

For current constrained provider semantics, absence of automatic retries is acceptable and reduces quota-amplification risk.

## 11. Circuit breaker

No circuit breaker exists.

A deterministic dead-provider campaign measured:

```text
INCOMING_REQUESTS=100
UPSTREAM_ATTEMPTS=60
PROCESS_LOCAL_REQUEST_BUDGET=60/minute
CONCURRENCY_LIMIT=4
```

Classification:

```text
CIRCUIT_BREAKER=CIRCUIT_BREAKER_NOT_REQUIRED
```

The existing bounded local budget and fail-fast concurrency control prevent unbounded dead-provider amplification. A breaker could be an operational optimization, but the hostile evidence does not justify classifying its absence as a security defect.

## 12. Single-flight / cache stampede

Identical concurrent requests were not coalesced:

```text
CLIENT_REQUESTS=10  UPSTREAM_REQUESTS=4   COALESCED_REQUESTS=0
CLIENT_REQUESTS=50  UPSTREAM_REQUESTS=4   COALESCED_REQUESTS=0
CLIENT_REQUESTS=100 UPSTREAM_REQUESTS=4   COALESCED_REQUESTS=0
```

The concurrency gate caps instantaneous amplification, but absence of request coalescing contributes directly to M10-P2-01 because identical anonymous requests can spend multiple provider-budget units.

Classification:

```text
CACHE_STAMPEDE=P2_PROVIDER_BUDGET_AMPLIFICATION
FINDING=M10-P2-01
```

## 13. Response-size / parser abuse

Runtime cases included:

```text
1 KiB bounded additive unknown field
100 KiB bounded additive unknown field
1 MiB response
10 MiB response
malformed JSON
truncated JSON
invalid UTF-8 in canonical enum data
very large enum field
large hostile array
deep hostile unknown nesting
duplicate JSON fields
negative money
invalid date
```

Results:

```text
BODY_SIZE_BOUND=PASS
PARSER_MEMORY_BOUND=PASS_WITH_TRANSPORT_CAP
NO_PROCESS_OOM=PASS
NO_UNBOUNDED_ALLOCATION=PASS_WITH_TRANSPORT_CAP
NO_PANIC=PASS
OVERSIZED_1MiB_REJECTED=YES
OVERSIZED_10MiB_REJECTED=YES
```

Duplicate-key ambiguity is separately recorded as M10-P3-01.

## 14. Provider data integrity

Confirmed protections:

- invalid dates are rejected;
- negative money is rejected;
- invalid enum values are rejected;
- invalid UTF-8 that corrupts canonical enum data is rejected;
- duplicate canonical corporate-action events are deduplicated;
- canonical Event IDs are derived by the adapter rather than trusting arbitrary provider-owned IDs;
- malformed and truncated provider evidence fails closed;
- projection processing does not persist provider payloads or financial mutations.

Future-dated events are not automatically invalid because the Corporate Actions calendar legitimately represents announced future events.

```text
PROVIDER_DATA_INTEGRITY=PASS_WITH_M10_P3_01_DUPLICATE_KEY_AMBIGUITY
NO_PARTIAL_DB_WRITE=YES_BY_ARCHITECTURE
```

## 15. Partial failure / transaction boundary

No provider persistence path exists for Corporate Actions. Provider data is fetched, canonicalized, validated, projected, and returned within the request.

```text
NO_PARTIAL_FINANCIAL_MUTATION=YES
NO_STUCK_DEDUP_STATE=NOT_APPLICABLE_NO_PROVIDER_DEDUP_PERSISTENCE
NO_ORPHAN_PROVIDER_STATE=NOT_APPLICABLE_NO_PROVIDER_PERSISTENCE
SAFE_RETRY=YES_FOR_READ_ONLY_PROJECTION
PARTIAL_FAILURE_ATOMICITY=NOT_APPLICABLE_NO_PROVIDER_PERSISTENCE
```

## 16. Secret / token handling

Evidence:

```text
TOKEN_SERVER_SIDE_ONLY=YES
TOKEN_RETURNED_TO_CLIENT=NO_EVIDENCE
TOKEN_STORED_IN_FRONTEND=NO
EMPTY_TOKEN_REJECTED=YES
LEADING_WHITESPACE_TOKEN_REJECTED=YES
TRAILING_WHITESPACE_TOKEN_REJECTED=YES
CONTROL_CHARACTER_TOKEN_REJECTED=YES
PROVIDER_DISABLED_BY_DEFAULT=YES
PRODUCTION_ACTIVATION_FAIL_CLOSED=YES
TRANSPORT_ERROR_SECRET_REDACTION=PASS
```

Repository contracts require a read-only token. The actual privileges of a real externally issued production credential were not exercised in this repository-only hostile audit.

## 17. SSRF / URL control

Exported production constructors use fixed provider base URLs. User-controlled request fields do not control provider scheme, host, or base URL.

```text
SSRF_ATTACK_SURFACE=NOT_USER_CONTROLLED
SSRF=NOT_APPLICABLE_WITH_REASON
```

Private constructors accept test base URLs only inside provider packages and do not expose a shipped user-controlled URL surface.

## 18. Redirect handling

Both T-Invest and dormant MOEX provider-owned clients reject redirects.

T-Invest cross-target redirect testing confirmed:

```text
REDIRECT_TARGET_REQUESTS=0
AUTHORIZATION_HEADER_LEAK=NO
```

MOEX redirect testing also confirmed zero target requests.

```text
REDIRECT_HANDLING=PASS
```

## 19. Connection / file-descriptor / goroutine evidence

A bounded 10,000-request synthetic transport-failure campaign recorded:

```text
FAILURE_REQUESTS=10000
GOROUTINES_BEFORE=2
GOROUTINES_AFTER=2
OPEN_FDS_BEFORE=9
OPEN_FDS_AFTER=9
PROVIDER_PERMITS_AFTER=0
```

Real-socket spot cases also covered slow no-header, partial-body, slow chunked, connection reset, caller cancellation, and TCP refusal.

However, the exact requested full cross-product of 100/1000/10000 **real-socket** requests under each of timeout, reset, malformed response, and cancellation was not executed.

Classification:

```text
NO_MONOTONIC_RESOURCE_LEAK=SUPPORTED_BY_CURRENT_EVIDENCE
EXHAUSTIVE_100_1000_10000_REAL_SOCKET_MATRIX=NOT_VERIFIED
```

This is an audit-evidence gap, not a confirmed product vulnerability.

## 20. Failure isolation

Sequence exercised:

```text
success
failure
failure
timeout
malformed response
success
```

Final success was recovered, permits returned to zero, and the 10,000-failure campaign did not poison subsequent capacity.

```text
FAILURE_ISOLATION=PASS
```

## 21. Multi-instance / replica chaos

Provider-replica harness with a shared deterministic upstream exercised:

```text
kill/cancel replica A active request
construct/restart replica A
pause replica B upstream request
continue traffic through replica A
resume replica B
final success through replica B
```

Measured:

```text
GLOBAL_DEADLOCK=NO
PERMANENT_BUDGET_LOSS=NO_EVIDENCE_IN_PROVIDER_HARNESS
STUCK_PERMITS=NO
SAFE_RECOVERY=YES
```

This proves provider-object replica isolation against a shared upstream, not full operating-system/container API-instance lifecycle.

```text
MULTI_INSTANCE_CHAOS=PARTIAL_PASS_PROVIDER_REPLICA_HARNESS
FULL_API_PROCESS_MULTI_INSTANCE_CHAOS=NOT_VERIFIED
```

## 22. Test infrastructure defect discovered and corrected

### TD-M10-01 — Passive httptest server-context observation was invalid as an upstream-cancellation witness

Initial audit tests blocked an `httptest.Server` handler waiting only on its own request context and expected the handler to observe client close immediately. That produced audit timeouts and could have been misclassified as a product cancellation failure.

The challenger was corrected without changing production code:

- connection cancellation is now proved with a raw TCP witness;
- slow `httptest.Server` handlers use explicit audit cleanup signals;
- T-Invest scenarios run as independent matrix jobs with `fail-fast: false`;
- final authoritative run is green.

Raw TCP proved that the client connection does close at the caller deadline for both no-header and partial-body upstreams.

```text
TEST_DEFECTS=1
TD_M10_01=CORRECTED
PRODUCT_FINDING_FROM_TD_M10_01=NO
```

## 23. Rejected hypotheses

The following hostile hypotheses were rejected by runtime/static evidence:

1. caller deadline is ignored by T-Invest provider;
2. provider concurrency permits leak on timeout/error;
3. provider can exceed four active upstream calls within one process;
4. automatic retries amplify 429/5xx failures;
5. redirects can forward the T-Invest bearer token to another host;
6. oversized 1 MiB / 10 MiB provider responses are accepted;
7. repeated transport failures monotonically leak goroutines/file descriptors in the measured campaign;
8. one failed provider request permanently poisons future provider capacity;
9. user input controls the production provider base URL;
10. dormant MOEX adapter is currently wired into shipped application runtime.

```text
REJECTED_HYPOTHESES=10
```

## 24. Not verified / carry-forward items

### NV-M10-01 — Real external shared provider budget owner

```text
REPOSITORY_SIDE_CONTRACT=VERIFIED
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES
```

### NV-M10-02 — Full API-process multi-instance chaos

Provider replicas were exercised, but two real application processes/containers were not killed/restarted/paused as part of this repository-only campaign.

```text
FULL_API_PROCESS_MULTI_INSTANCE_CHAOS=NOT_VERIFIED
```

### NV-M10-03 — Exhaustive real-socket FD/leak matrix

10,000 synthetic transport failures plus real-socket representative cases were executed, but the complete 100/1000/10000 real-socket cross-product for every required failure mode was not.

```text
EXHAUSTIVE_REAL_SOCKET_RESOURCE_MATRIX=NOT_VERIFIED
```

These items must not be silently converted into PASS.

## 25. Category status matrix

| # | Category | Status | Evidence / classification |
|---:|---|---|---|
| 1 | Slow/hanging provider | PASS | deadline, raw TCP cancellation, connection/permit release, recovery |
| 2 | Distributed provider budget | PASS_WITH_CARRY_FORWARD | 60 -> 120 -> 240 multiplication measured; external shared control not verified |
| 3 | Provider endpoint budget abuse | FAIL | M10-P2-01 |
| 4 | Concurrency exhaustion | PASS | max active <= 4; fail-fast; recovery |
| 5 | Retry / backoff | PASS | no automatic retries; one attempt for 429/5xx |
| 6 | Circuit breaker | PASS | CIRCUIT_BREAKER_NOT_REQUIRED based on bounded pressure |
| 7 | Single-flight / cache stampede | FAIL | contributes to M10-P2-01 |
| 8 | Response size / parser abuse | FAIL | size bounds pass; duplicate-key ambiguity M10-P3-01 |
| 9 | Provider data integrity | PASS | fail-closed canonical validation, with M10-P3-01 noted |
| 10 | Partial failure / transaction boundary | NOT_APPLICABLE_WITH_REASON | no provider persistence/financial DB mutation |
| 11 | Secret / token handling | PASS | server-side, fail-closed activation, redaction |
| 12 | SSRF / URL control | NOT_APPLICABLE_WITH_REASON | production base URL not user-controlled |
| 13 | Redirect handling | PASS | redirect targets receive zero requests |
| 14 | Connection pool / FD leak | BLOCKED_WITH_REASON | strong evidence, but exhaustive requested real-socket matrix not completed |
| 15 | Failure isolation | PASS | failure sequence recovers |
| 16 | Multi-instance chaos | BLOCKED_WITH_REASON | provider-replica chaos passes; full API-process chaos not verified |

```text
SUBTESTS_PLANNED=16
SUBTESTS_EXECUTED=16
SUBTESTS_PASS=9
SUBTESTS_FAIL=3
SUBTESTS_BLOCKED=2
SUBTESTS_NOT_APPLICABLE=2
SUBTESTS_SKIPPED_DUE_TO_EARLY_FAILURE=0
```

## 26. Severity summary

```text
CONFIRMED_P0=0
CONFIRMED_P1=0
CONFIRMED_P2=1
CONFIRMED_P3=1

CONFIRMED_FINDINGS=2
REJECTED_HYPOTHESES=10
TEST_DEFECTS=1
NOT_VERIFIED_ITEMS=3
```

Confirmed findings:

```text
M10-P2-01 PUBLIC_PROVIDER_BUDGET_EXHAUSTION
M10-P3-01 DUPLICATE_JSON_KEY_AMBIGUITY
```

## 27. Canonical hostile-audit state

```text
TASK=PROVIDERS_EXTERNAL_API_BUDGET_ABUSE_RESILIENCE_HOSTILE_AUDIT_V1

BASE_SHA=af16e498fff348c5c70938f127cd164a84a227aa
BASE_TREE=e927460d8d5d668fed4255e8d3a8dfd519df5053
BASELINE_DRIFT=NO

AUDIT_BRANCH=audit/providers-external-api-budget-resilience-v1

PROVIDERS_DISCOVERED=T_INVEST_API,MOEX_ISS_DORMANT
EXTERNAL_ENDPOINTS_DISCOVERED=GET_/api/v1/corporate-actions/projection

SLOW_PROVIDER_HANG_TEST=PASS
CALLER_DEADLINE_HONORED=YES
UPSTREAM_REQUEST_CANCELLED=YES
CONNECTION_RELEASED=YES
PROVIDER_SLOT_RELEASED=YES
GOROUTINE_LEAK=NO
RECOVERY_PASS=YES

DISTRIBUTED_PROVIDER_BUDGET=PROCESS_LOCAL_MULTIPLICATION_WITH_REPOSITORY_CONTRACT
ONE_INSTANCE_ADMITTED=60
TWO_INSTANCE_ADMITTED=120
FOUR_INSTANCE_ADMITTED=240
BUDGET_MULTIPLICATION=PROCESS_LOCAL_MULTIPLICATION
REPOSITORY_SIDE_CONTRACT=VERIFIED
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES

PROVIDER_ENDPOINT_BUDGET_ABUSE=CONFIRMED_P2
CONCURRENCY_EXHAUSTION=PASS
RETRY_BACKOFF=PASS_NO_AUTOMATIC_RETRIES
CIRCUIT_BREAKER=CIRCUIT_BREAKER_NOT_REQUIRED
CACHE_STAMPEDE=P2_PROVIDER_BUDGET_AMPLIFICATION
RESPONSE_SIZE_BOUND=PASS
PROVIDER_DATA_INTEGRITY=P3_DUPLICATE_JSON_KEY_AMBIGUITY
PARTIAL_FAILURE_ATOMICITY=NOT_APPLICABLE_NO_PROVIDER_PERSISTENCE
TOKEN_SECRET_HANDLING=PASS
SSRF=NOT_APPLICABLE_WITH_REASON_SSRF_ATTACK_SURFACE_NOT_USER_CONTROLLED
REDIRECT_HANDLING=PASS
CONNECTION_FD_LEAK=BLOCKED_WITH_REASON_EXHAUSTIVE_100_1000_10000_REAL_SOCKET_MATRIX_NOT_EXECUTED
FAILURE_ISOLATION=PASS
MULTI_INSTANCE_CHAOS=PARTIAL_PASS_PROVIDER_REPLICA_HARNESS_FULL_API_PROCESS_NOT_VERIFIED

SUBTESTS_PLANNED=16
SUBTESTS_EXECUTED=16
SUBTESTS_PASS=9
SUBTESTS_FAIL=3
SUBTESTS_BLOCKED=2
SUBTESTS_NOT_APPLICABLE=2
SUBTESTS_SKIPPED_DUE_TO_EARLY_FAILURE=0

CONFIRMED_P0=0
CONFIRMED_P1=0
CONFIRMED_P2=1
CONFIRMED_P3=1

CONFIRMED_FINDINGS=2
REJECTED_HYPOTHESES=10
TEST_DEFECTS=1
NOT_VERIFIED_ITEMS=3

PRODUCTION_CODE_CHANGED=NO
FINAL_REMEDIATION_COMMIT=NO
PR_CREATED=NO
MERGE_PERFORMED=NO

READY_FOR_INDEPENDENT_REVIEW=YES
FORMAL_MODULE_CLOSED=NO
REPOSITORY_WIDE_AUDIT=ONGOING

FINAL_STATUS=READY_FOR_INDEPENDENT_REVIEW
NEXT_ACTION=STOP_FOR_INDEPENDENT_REVIEW
```

## 28. Stop gate

No Module #10 remediation was performed.

No remediation PR was created.

No merge was performed.

Module #11 and Module #12 were not started.

The next action is independent review of this hostile audit and its findings.
