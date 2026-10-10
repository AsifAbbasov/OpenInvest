# Final Production / Frontend / Infrastructure Hardening — Adversarial Audit V1

## Scope and immutable baseline

- Module: **12 — Final Production / Frontend / Infrastructure Hardening**.
- Mode: detection / verification only; no product remediation is included.
- Frozen protected baseline: `develop@7631e9fe62a8be2a527d12c608c8539771dd3277`.
- Frozen baseline tree: `b97fe7bd2e03b6a9251e7cb581450366b69c66ef`.
- Audit branch: `audit/final-production-frontend-infra-hardening-v1`.
- Product/runtime source changed by this audit: **none**.
- Audit-only additions: dedicated workflow/tests and this evidence report.
- The later repository-wide fuzz/property/security assault was not started.

## Run chronology

| Run | Head | Result | Classification |
| --- | --- | --- | --- |
| `37959604010` | `a224fb30d4734bc5e406c27fbe34e1f353de9916` | FAILURE | Provider real-socket audit timed out after 10m while harness waited for upstream cancellation. |
| `37960141249` | `74d52d3de3bd69d98d13e7ee22325950a844b11b` | FAILURE | Same unbounded provider-cancellation harness wait; timeout after 10m. |
| `37960770615` | `abce2b912868e33366da68590361bcb69960229e` | FAILURE | Workflow-composition defect; corrected audit workflow only. |
| `37960928767` | `ae57cf2d50114765c4f77e8727cc8d24a60603b5` | FAILURE | Provider cancellation harness still waited unbounded; timeout after 8m. |
| `37961449087` | `bc830613f055d6f567f4a0f206f36b3b0412dd20` | SUCCESS | Corrected bounded seven-job Module #12 specialized audit. |
| `37961448982` | `bc830613f055d6f567f4a0f206f36b3b0412dd20` | SUCCESS | Normal repository CI, 10/10 required jobs. |

The failed chronology is preserved. Harness-only defects are TD-M12-01/02 below; they are not product findings.

## Exact tested topology

- Real TCP listeners for Fiber HTTP boundary on loopback.
- Independent in-process application replicas, each with its own limiter/coalescer/provider state.
- Replica counts: **1, 2, 4**.
- Real `net/http` sockets to an `httptest` upstream for provider cancellation/reset/partial-response/timeout behavior.
- Real PostgreSQL 18 service container for pool exhaustion/recovery.
- Production Next.js build + `next start` runtime on loopback.
- Repository Docker Compose configuration only; services present: PostgreSQL and Redis.
- No repository-owned application container runtime manifest.
- No repository-owned external CDN/load-balancer/TLS edge manifest.

These tests prove repository behavior under the tested topology. They do **not** prove the actual deployed production topology.

## P2 findings

### M12-P2-01 — MULTI_REPLICA_PROVIDER_BUDGET_MULTIPLICATION

- Severity: **P2**.
- Boundary: application replicas / expensive Corporate Actions provider budget.
- Reproduction: real HTTP boundary with independent replicas.
  - 1 replica: 60 HTTP requests → 48 HTTP 2xx, 12 HTTP 429, 48 provider calls.
  - 2 replicas: 120 HTTP requests → 96 HTTP 2xx, 24 HTTP 429, 96 provider calls.
  - 4 replicas: 240 HTTP requests → 192 HTTP 2xx, 48 HTTP 429, 192 provider calls.
  - Process restart resets the process-local endpoint ceiling; first request to the fresh process returns 200.
  - Provider-level process-local limit independently scales 60 → 120 → 240 upstream calls for 1/2/4 Provider instances.
- Attacker prerequisites: deployment with multiple independently stateful application replicas and sufficient requests/routing to exercise more than one replica.
- Impact: horizontally scaled deployment can multiply expensive external-provider traffic and provider-account quota/cost pressure by replica count; restart/rolling restart also resets local budget state.
- Evidence: specialized run `37961449087`, HTTP markers `M12_MULTI_INSTANCE`, `M12_PROCESS_RESTART_RESETS_LOCAL_BUDGET=YES`, provider marker `M12_PROVIDER_INSTANCE_BUDGET`.
- Production exploitation: **NOT demonstrated**; repository-side reachability and absence of shared enforcement are demonstrated.
- Recommendation: use a deployment-global/shared budget owner or external provider-account enforcement with independently verified distributed semantics; do not represent the existing process-local ceilings as production-global.

### M12-P2-02 — DISCONNECT_CANCELLATION_DOES_NOT_BOUND_UPSTREAM_PROVIDER_WORK

- Severity: **P2**.
- Boundary: downstream disconnect / request context / provider transport / concurrency guard.
- Reproduction:
  - Real downstream TCP client disconnect during provider work produced `M12_DOWNSTREAM_DISCONNECT_DURING_PROVIDER_PROPAGATES_CONTEXT=false`.
  - Direct provider context-cancellation campaign ran 5 rounds × 4 concurrent requests = **20 cancellations**.
  - After 750ms: upstream active requests = **20**, upstream context cancellations = **0**, while local provider semaphore slots in use = **0**.
  - `provider_max_active=20` despite configured local concurrency protection; subsequent legitimate request still succeeded.
  - Partial upstream response, connection reset and 5s timeout failed closed and recovered.
- Attacker prerequisites: ability to initiate provider-backed requests and disconnect/cancel before provider completion.
- Impact: abandoned requests can continue consuming upstream sockets/provider work after local concurrency slots are released, permitting cancellation-driven upstream concurrency amplification until timeout/resource cleanup.
- Evidence: specialized run `37961449087`, markers `M12_DOWNSTREAM_DISCONNECT_DURING_PROVIDER_PROPAGATES_CONTEXT=false` and `M12_PROVIDER_UPSTREAM_CANCEL_PROPAGATED=false`.
- Production exploitation: **NOT demonstrated against the real external provider**; repository-side real-socket reachability is demonstrated.
- Recommendation: preserve downstream cancellation through the real HTTP boundary and ensure upstream transport work is terminated before releasing local concurrency accounting; verify with real provider/deployment observability after remediation authorization.

## HARDENING_ONLY observations

### M12-H01 — PARTIAL_BROWSER_HEADER_SET_AT_NEXT_RUNTIME

Production `next start` emitted:
- `Content-Security-Policy: frame-ancestors 'none'`
- `X-Frame-Options: DENY`

It did not emit `Strict-Transport-Security`, `X-Content-Type-Options`, `Referrer-Policy`, or `Permissions-Policy`. Repository operations documentation explicitly delegates TLS/HSTS to an external edge, so HSTS absence is not upgraded to a production exploit finding here. No exploit from the other missing headers was demonstrated.

### M12-H02 — PRODUCTION_CORS_MISSING_ALLOWLIST_DEFAULTS_TO_LOCALHOST

With `OPENINVEST_ALLOWED_WEB_ORIGINS` unset, the tested application CORS configuration allowed `http://localhost:3000`. Exact configured-origin allowlisting and disallowed-origin credential denial passed. Refresh cookie is `Secure`, `HttpOnly`, `SameSite=Strict`, path `/api/v1/auth`, with `Cache-Control: no-store`. No cross-principal browser exploit was demonstrated.

### M12-H03 — ARBITRARY_HOST_ACCEPTED_WITHOUT_REFLECTION

The backend accepted arbitrary `Host` / forwarded-host material at the tested direct boundary, but did not reflect it in response headers and did not influence redirects. An external edge host allowlist is not represented in repository configuration.

## Browser / secret / cache evidence

- Frontend production static source-map count: **0**.
- Provider/access/import-review secret sentinels in browser static artifacts: **none**.
- Secret sentinels in rendered runtime body: **none**.
- Source scan found no access/refresh/CSRF token storage references in `localStorage`, `sessionStorage`, or JS-written `document.cookie`.
- Frontend root runtime response used public `Cache-Control: s-maxage=31536000`.
- Auth/session response under test used `Cache-Control: no-store`.
- Public Corporate Actions handler owns `Cache-Control: no-store`.
- Exact configured-origin CORS allowlist passed; disallowed origin received no credential authorization.
- HTTP→HTTPS redirect and HSTS at the real production edge were not repository-owned/testable.

## Resource / socket evidence

### Backend HTTP boundary

- Disconnect before a complete HTTP request reached provider: **0 provider calls**.
- Slow incomplete request bounded by configured 15s read timeout.
- Goroutines: baseline 11, peak 15, recovery 9.
- File descriptors: baseline 9, peak 12, recovery 11.
- Subsequent app lifecycle/coalescer recovery completed.

### Provider upstream boundary

- Concurrent cancellations attempted: **20**.
- Local provider semaphore after cancellation: **0 slots in use**.
- Upstream active after 750ms: **20**.
- Upstream observed context cancellations: **0**.
- Partial response: fail closed, recovery PASS.
- Connection reset: fail closed, recovery PASS.
- Upstream timeout: ~5.005s, recovery PASS.
- Goroutines: baseline 3, peak 23, recovery 4.
- File descriptors: baseline 10, peak 30, recovery 11.
- Provider max active upstream requests observed: **20**.

### PostgreSQL pool

- Max open: 10.
- Peak open/in-use: 10/10.
- 11th connection blocked until context deadline (~401ms).
- Wait count: 1.
- After releasing capacity: connection acquisition succeeded; recovery in-use = 0, idle = 5.

## Shared-provider-budget verdict

```text
IS_PROVIDER_BUDGET_SHARED_ACROSS_REPLICAS=NO
CAN_N_REPLICAS_MULTIPLY_PROVIDER_CALL_BUDGET=YES
WHERE_IS_GLOBAL_BUDGET_STATE_STORED=APPLICATION_PROCESS_MEMORY
IS_ENFORCEMENT_AT_APP_PROCESS_ONLY=YES
IS_ENFORCEMENT_AT_PROVIDER_ACCOUNT_LEVEL_VERIFIED=NO
REAL_EXTERNAL_PROVIDER_BUDGET_TEST_EXECUTED=NO
```

Startup acknowledgements for deployment-global abuse control and provider-budget ownership are configuration acknowledgements only, not external/shared-state proof.

## Infrastructure / deployment boundary

Repository evidence shows:
- Docker Compose services: PostgreSQL and Redis.
- Compose host bindings for those services are loopback-only.
- No application Dockerfile/runtime manifest is present.
- No external edge/CDN/load-balancer manifest is present.
- No repository TLS/HSTS edge configuration is present.
- Production wildcard API listen address can be explicitly configured.
- Trusted proxy mode requires explicit trusted CIDRs; direct mode ignores forwarding headers.
- Rotating/duplicate X-Forwarded-For did not bypass direct-mode per-client limiter.
- Actual cloud/container capabilities, seccomp/AppArmor, read-only filesystem, non-root UID, orchestration network policy and real production PostgreSQL connection policy cannot be proven from the repository artifacts tested.

## NOT_VERIFIED table

| ID | State | Boundary |
| --- | --- | --- |
| NV-M12-01 | NOT_VERIFIED | Actual production CDN/load balancer/reverse-proxy topology and trusted CIDRs. |
| NV-M12-02 | NOT_VERIFIED | Real production TLS termination, HTTP→HTTPS redirect and HSTS emission. |
| NV-M12-03 | NOT_VERIFIED | Real external provider-account global quota/budget enforcement. |
| NV-M12-04 | NOT_VERIFIED | Real external-provider cancellation/connection termination behavior and account-side resource recovery. |
| NV-M12-05 | NOT_VERIFIED | Actual multi-host/container replica routing, shared state and rolling-restart behavior in production orchestration. |
| NV-M12-06 | NOT_VERIFIED | Production container runtime privileges/capabilities, seccomp/AppArmor, filesystem and network policy. |
| NV-M12-07 | NOT_VERIFIED | Production log aggregation/redaction behavior for tokens, cookies, provider credentials and financial payloads. |
| NV-M12-08 | PARTIALLY_VERIFIED | Full duplicate/conflicting Forwarded/XFF/CF-Connecting-IP matrix behind the real edge; repository direct/trusted cases only. |
| NV-M12-09 | PARTIALLY_VERIFIED | Full browser CSRF/auth race at actual deployed origins; repository cookie/CORS controls verified, real deployed origin boundary not exercised. |
| NV-M12-10 | PARTIALLY_VERIFIED | Exhaustive slow-client/half-open/disconnect timing matrix; representative real-socket cases executed. |
| NV-M12-11 | NOT_VERIFIED | Production database TLS/network/firewall/runtime-role topology outside repository CI. |
| NV-M12-12 | PARTIALLY_VERIFIED | Cache/CDN poisoning at actual CDN edge; repository response cache controls inspected/tested, no real CDN exists in evidence. |

Carried-forward residuals:
- `NV-M11-09=HORIZONTALLY_SHARED_PRODUCTION_PROVIDER_BUDGET_NOT_VERIFIED`.
- `REAL_EXTERNAL_SHARED_PROVIDER_BUDGET_NOT_VERIFIED`.
- `FULL_API_PROCESS_MULTI_INSTANCE_CHAOS_NOT_VERIFIED`: repository multi-instance HTTP behavior is now reproduced for 1/2/4 independent replicas, but actual production multi-process/container/load-balancer chaos remains not verified.
- `EXHAUSTIVE_REAL_SOCKET_RESOURCE_MATRIX_NOT_VERIFIED`: representative real-socket cases are now executed, but the exhaustive phase/timing matrix remains not verified.

## Test defects

- **TD-M12-01 — UNBOUNDED_PROVIDER_CANCELLATION_HARNESS_WAIT.** Specialized runs `37959604010`, `37960141249`, and `37960928767` timed out because the harness waited for upstream cancellation that the product did not deliver. The harness was bounded without changing product behavior; the corrected run preserved and measured the non-cancellation result.
- **TD-M12-02 — WORKFLOW_COMPOSITION_DEFECT.** Run `37960770615` failed at workflow composition; the audit workflow was corrected without product changes.

## Security scanner / normal CI evidence

Normal CI run `37961448982` on audit harness head `bc830613f055d6f567f4a0f206f36b3b0412dd20` completed **SUCCESS**, 10/10 required jobs:
- Go tests: SUCCESS.
- Python tests: SUCCESS.
- Frontend build/typecheck/tests: SUCCESS.
- OpenAPI contract: SUCCESS.
- Docker Compose config: SUCCESS.
- PostgreSQL migration validation: SUCCESS.
- Go vet: SUCCESS.
- Go race tests: SUCCESS; no `DATA RACE`.
- Go vulnerability scan: `No vulnerabilities found`; affected by 0 vulnerabilities.
- Dependency security scan: pnpm audit and pip-audit both reported no known vulnerabilities.

## Finding summary

| ID | Severity | Production exploitation demonstrated? | Disposition |
| --- | --- | --- | --- |
| M12-P2-01 | P2 | NO — repository multi-replica reachability demonstrated | STOP for independent classification/remediation authorization |
| M12-P2-02 | P2 | NO — repository real-socket reachability demonstrated | STOP for independent classification/remediation authorization |
| M12-H01 | HARDENING_ONLY | NO | Browser header posture |
| M12-H02 | HARDENING_ONLY | NO | CORS deployment hardening |
| M12-H03 | HARDENING_ONLY | NO | Edge host allowlist hardening |
| TD-M12-01 | TEST_DEFECT | N/A | Harness corrected |
| TD-M12-02 | TEST_DEFECT | N/A | Workflow corrected |

## Final disposition

- P0: **0**
- P1: **0**
- P2: **2**
- P3: **0**
- HARDENING_ONLY: **3**
- TEST_DEFECTS: **2**
- NOT_VERIFIED entries: **12**
- Product remediation authorized: **NO**
- Formal Module #12 closure: **NO**
- Final repository-wide fuzz/property/security assault started: **NO**

The audit does not claim production-wide provider-budget enforcement, actual edge/TLS guarantees, real external-provider quota behavior, or exhaustive socket/resource safety from repository-only tests.
