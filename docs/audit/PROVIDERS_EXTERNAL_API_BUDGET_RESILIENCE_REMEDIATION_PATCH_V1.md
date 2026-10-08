# Module #10 — Providers / External API / Budget Abuse / Resilience — Remediation Patch V1

## 1. Identity and authorization

```text
TASK=PROVIDERS_EXTERNAL_API_BUDGET_RESILIENCE_REMEDIATION_PATCH_V1
MODULE=10
AUTHORIZED_FINDING=M10-P2-01
AUTHORIZED_SEVERITY=P2
AUTHORIZED_CLASS=PUBLIC_PROVIDER_BUDGET_EXHAUSTION
M10_P3_01=HARDENING_ONLY
M10_P3_01_SECURITY_FINDING=NO
REMEDIATION_AUTHORIZED=YES
APPROVE_FOR_COMMIT=NO
APPROVE_FOR_MERGE=NO
FORMAL_MODULE_CLOSED=NO
```

This patch remediates only the repository-side anonymous provider-budget exhaustion path. It does not claim deployment-global enforcement, does not remediate the duplicate-key hardening item, and does not begin Module #11 or Module #12.

## 2. Frozen baseline

```text
PROTECTED_BRANCH=develop
BASE_SHA=af16e498fff348c5c70938f127cd164a84a227aa
BASE_TREE=e927460d8d5d668fed4255e8d3a8dfd519df5053
BASELINE_DRIFT=NO
BRANCH_PROTECTED=YES

REMEDIATION_BRANCH=remediation/module10-provider-budget-abuse-v1
AUTHORITATIVE_RUNTIME_HEAD=94170dd6f9350cab89bbc4559eee4b022a4b4bde
AUTHORITATIVE_RUNTIME_TREE=b0faf648064b5d9a5e6a785958d5601ce61a0660
AUTHORITATIVE_RUN_ID=37629789116
AUTHORITATIVE_RUN_RESULT=SUCCESS
```

The authoritative runtime head contains the production patch, tests, OpenAPI/operations updates, and a temporary evidence workflow. The canonical review branch removes that temporary workflow after the successful run and adds this report; the workflow remains preserved in Git history at the authoritative runtime head.

## 3. Repository-side design

The public endpoint remains intentionally anonymous:

```text
GET /api/v1/corporate-actions/projection
AUTH_REQUIRED=NO
CSRF_REQUIRED=NO
```

The remediation inserts abuse control after canonical query validation and before provider work:

```text
PER_CLIENT_CANONICAL_IP_LIMIT=12/minute
PROCESS_GLOBAL_ENDPOINT_CEILING=48/minute
LIMITER_WINDOW=60_seconds
MAX_TRACKED_CLIENT_KEYS=128
REQUEST_COALESCING=YES
COALESCER_MAX_INFLIGHT_KEYS=48
CACHE_ENABLED=NO
PROVIDER_PROCESS_LOCAL_BUDGET=60/minute
PROVIDER_PROCESS_LOCAL_CONCURRENCY=4
```

The client key uses the repository's existing validated peer/proxy trust model through normalizedClientIP. Direct mode ignores arbitrary forwarding headers. Trusted-proxy behavior remains governed by HTTPNetworkConfig.

The coalescing key is a fixed-size SHA-256 digest of the canonical request identity:

```text
sorted canonical instrument IDs
+
from date
+
to date
```

This prevents query-order and equivalent canonical request variants from creating separate in-flight provider work. There is no response cache, so no stale Corporate Actions payload is introduced.

Shared in-flight provider work uses reference-counted cancellation: one waiter leaving does not cancel work still needed by other waiters; when all waiters cancel, the shared provider context is cancelled. Completed/error calls are removed from the bounded map.

## 4. Old attack vs patched routed endpoint

Canonical old result:

```text
ANONYMOUS_REQUESTS=60
BASELINE_PROVIDER_CALLS=60
LEGITIMATE_REQUEST_AFTER_EXHAUSTION=503
```

Patched exact routed result:

```text
ANONYMOUS_REQUESTS=60
HTTP_200=12
HTTP_429=48
HTTP_503=0
PROVIDER_CALLS_CREATED=12
```

The old one-for-one relationship is broken:

```text
UNAUTHENTICATED_CLIENT_CANNOT_ONE_FOR_ONE_EXHAUST_PROVIDER_BUDGET=YES
```

The 429 responses are safe abuse rejections and include Retry-After: 60.

Immediately after the attack, a legitimate request from a different canonical peer produced:

```text
LEGITIMATE_REQUEST_AFTER_ATTACK_STATUS=200
LEGITIMATE_PROVIDER_CALL_AFTER_ATTACK=YES
```

Therefore one anonymous client no longer consumes the complete provider allowance and no longer forces the next legitimate routed request into provider-budget 503.

## 5. Identical / distinct / canonical-equivalent attacks

Identical concurrent routed requests:

```text
IDENTICAL_CLIENTS=10  PROVIDER_CALLS=1  HTTP_200=10 HTTP_429=0  MAX_ACTIVE_PROVIDER_CALLS=1
IDENTICAL_CLIENTS=50  PROVIDER_CALLS=1  HTTP_200=12 HTTP_429=38 MAX_ACTIVE_PROVIDER_CALLS=1
IDENTICAL_CLIENTS=100 PROVIDER_CALLS=1  HTTP_200=12 HTTP_429=88 MAX_ACTIVE_PROVIDER_CALLS=1
```

Required amplification property:

```text
IDENTICAL_CONCURRENT_PROVIDER_AMPLIFICATION<=1=PASS
```

Distinct valid query attack:

```text
REQUESTS=60
PROVIDER_CALLS=12
HTTP_200=12
HTTP_429=48
RESULT=PASS
```

Canonical-equivalent query-order attack:

```text
CLIENTS=10
REQUEST_VARIANTS=SBER,GAZP|GAZP,SBER
PROVIDER_CALLS=1
RESULT=PASS
```

Fresh request IDs, idempotency keys, or arbitrary headers are not part of limiter or coalescer identity.

## 6. Multi-instance repository-side retest

For the same abusive canonical client IP across separate routed application/provider objects:

```text
ONE_INSTANCE_PROVIDER_CALLS=12
TWO_INSTANCE_PROVIDER_CALLS=24
FOUR_INSTANCE_PROVIDER_CALLS=48
```

This is a repository-side improvement over the hostile/challenger 60/120/240 behavior for one anonymous client, but the endpoint limiter and provider limiter are still process-local.

A separate many-IP process-global campaign proved the emergency ceiling:

```text
REQUESTS=60
DISTINCT_CLIENT_IPS=60
PROVIDER_CALLS=48
HTTP_200=48
HTTP_429=12
PROCESS_GLOBAL_ENDPOINT_CEILING=48
```

The repository does not claim this process-global ceiling is shared across replicas.

## 7. Deployment-global residual

The existing startup ownership contract remains distinct from runtime proof:

```text
REPOSITORY_SIDE_PROVIDER_BUDGET_CONTRACT=VERIFIED
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES
```

A real shared edge/provider-budget owner across replicas is still a Module #12 verification item. No Redis or distributed dependency was added to manufacture a false closure.

## 8. NAT / proxy / IPv4 / IPv6 fairness

Runtime tests proved:

```text
SHARED_NAT_NORMAL_REQUESTS_BEFORE_LIMIT=12
FALSE_POSITIVE_429_WITHIN_NORMAL_TEST_FLOW=NO
DIRECT_MODE_X_FORWARDED_FOR_ROTATION_BYPASS=NO
IPV4=PASS
IPV6=PASS
```

The direct-mode test changed X-Forwarded-For repeatedly while the real TCP peer remained fixed; the thirteenth request still received 429, proving arbitrary forwarded headers cannot create fresh buckets.

The actual frontend Corporate Actions form issues one abortable request per explicit submit and aborts the previous request before starting a replacement. The 12/minute per-client limit is above this normal request pattern.

## 9. Cardinality and memory bounds

A bounded 10,000-key attack against the limiter recorded:

```text
ATTEMPTS=10000
ACTIVE_KEYS_BEFORE_EXPIRY=48
CONFIGURED_MAX_KEYS=128
ACTIVE_KEYS_AFTER_WINDOW_EXPIRY_AND_SWEEP=1
BOUNDED_ENTRY_COUNT=YES
EXPIRATION_OR_EVICTION=YES
NO_PERMANENT_ATTACKER_CONTROLLED_STATE=YES
MEMORY_GROWTH_BOUND=YES_BY_FIXED_LIMITS
```

The admitted-key count is additionally bounded by the 48/minute process-global ceiling. Coalescer in-flight keys are independently capped at 48 and are removed on success, error, or all-waiter cancellation.

## 10. Cancellation and provider resilience regression

The authoritative run materialized and reran the prior Module #10 provider hostile tests from:

```text
HOSTILE_RUNTIME_HEAD=f995d3f97d8c5e203aef6254e7e19162da6ae2b6
```

Confirmed again:

```text
CALLER_DEADLINE_HONORED=YES
UPSTREAM_CONNECTION_CLOSED=YES
PROVIDER_SLOT_RELEASED=YES
RECOVERY_PASS=YES
NO_RETRY_STORM=YES
```

Raw TCP witnesses again observed connection closure for never-headers and partial-body hangs. Slow chunked, connection reset, TCP refusal/DNS-like failure, 429, 500, 502, 503, 504, malformed/truncated/oversized provider data, concurrency exhaustion, failure isolation, token handling, and dormant MOEX bounds all passed.

The existing provider-object hostile stampede test still sees the provider's direct concurrency bound of four because singleflight is intentionally implemented at the routed endpoint layer, not inside the provider adapter.

## 11. Singleflight failure/cancellation state

Targeted tests proved:

```text
NO_SINGLEFLIGHT_STUCK_ENTRY=YES
ERROR_RECOVERY=PASS
ALL_WAITER_CANCELLATION_REMOVES_ACTIVE_CALL=YES
NO_CACHE_POISONING=NOT_APPLICABLE_NO_CACHE
```

Failed provider results are never cached because no response cache was introduced.

## 12. OpenAPI and frontend contract

OpenAPI now explicitly documents normal 429 behavior for:

```text
GET /api/v1/corporate-actions/projection
```

The existing reusable RateLimited response provides:

```text
HTTP_429
Retry-After
ErrorResponse
X-Request-ID
X-Trace-ID
```

Runtime and OpenAPI are therefore aligned.

No frontend implementation change was required. Frontend typecheck, tests, and production build all passed.

## 13. Duplicate JSON key item

Independent challenger classification remains unchanged:

```text
M10_P3_01=HARDENING_ONLY
M10_P3_01_SECURITY_FINDING=NO
M10_P3_01_CLOSURE_BLOCKER=NO
DUPLICATE_JSON_KEY_HARDENING=DEFERRED_NON_BLOCKING
```

The P2 remediation does not depend on parser changes.

## 14. Test infrastructure defect control

The first WIP evidence run used fiber.App.Test with net/http.Request.RemoteAddr to simulate distinct client IPs. Fiber's test adapter did not preserve that net/http peer address into the fasthttp request context, causing the "legitimate different client" to remain in the attacker's bucket.

```text
TEST_DEFECT=TD-M10-02
CLASS=TEST_HARNESS_DEFECT
PRODUCT_FINDING_FROM_TEST_DEFECT=NO
CORRECTION=explicit_fasthttp_RequestCtx_with_real_TCP_peer
```

After correction, the routed legitimate-survival test passed and all other evidence was rerun on the corrected exact head.

## 15. Authoritative validation

```text
RUN_ID=37629789116
HEAD_SHA=94170dd6f9350cab89bbc4559eee4b022a4b4bde
HEAD_TREE=b0faf648064b5d9a5e6a785958d5601ce61a0660
RESULT=SUCCESS

TARGETED_MODULE_10_REMEDIATION=SUCCESS
PRIOR_HOSTILE_PROVIDER_REGRESSION=SUCCESS
GO_TEST_ALL=SUCCESS
GO_RACE_ALL=SUCCESS
DATA_RACE=NONE
GO_VET=SUCCESS
OPENAPI_CONTRACT=SUCCESS
FRONTEND_TYPECHECK=SUCCESS
FRONTEND_TESTS=SUCCESS
FRONTEND_BUILD=SUCCESS
GO_VULNERABILITY_SCAN=SUCCESS
PNPM_AUDIT=SUCCESS
FINAL_EVIDENCE_ENFORCEMENT=SUCCESS
```

Dependency files were not changed. Security scans were nevertheless rerun and were green.

## 16. Patch status

Repository-side finding state:

```text
M10_P2_01_BEFORE_STATUS=CONFIRMED_P2
M10_P2_01_AFTER_STATUS=REMEDIATED_PENDING_INDEPENDENT_PATCH_REVIEW
NEW_CONFIRMED_FINDINGS=0
PATCH_BLOCKERS=0
```

Residual state:

```text
EXTERNAL_SHARED_PROVIDER_BUDGET=NOT_VERIFIED
CARRY_FORWARD_MODULE_12=YES
NV_M10_02=FULL_API_PROCESS_MULTI_INSTANCE_CHAOS_NOT_VERIFIED
NV_M10_03=EXHAUSTIVE_REAL_SOCKET_RESOURCE_MATRIX_NOT_VERIFIED
```

No final remediation commit, PR, or merge has been authorized.

```text
FINAL_REMEDIATION_COMMIT=NO
PR_CREATED=NO
MERGE_PERFORMED=NO
READY_FOR_INDEPENDENT_PATCH_REVIEW=YES
APPROVE_FOR_COMMIT=NO
APPROVE_FOR_MERGE=NO
FORMAL_MODULE_CLOSED=NO
REPOSITORY_WIDE_AUDIT=ONGOING
NEXT_ACTION=STOP_FOR_INDEPENDENT_PATCH_REVIEW
```
