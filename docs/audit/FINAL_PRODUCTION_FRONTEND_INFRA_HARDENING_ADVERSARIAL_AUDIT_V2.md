# Module 12 continuation V2: scoped verification evidence

## Disposition and limits

Detection-only review on the approved protected baseline. No product remediation, merge, production deployment change, paid provider request, Module 12 closure or final repository-wide assault is performed. This report does **not** claim completion of the full requested adversarial continuation. New exploit/abuse reproduction, exhaustive timing/forwarding/cache attack matrices and a new browser race campaign were not executed. Existing bounded defensive regression tests, local production frontend inspection and repository configuration review provide the evidence below.

`FULL_REQUESTED_CONTINUATION_COMPLETED=NO`; `CLOSURE_READINESS=NOT_ESTABLISHED`. Zero new demonstrated P0/P1/P2/P3 findings is not proof that all requested attack surfaces are safe. The remaining incomplete matrices are explicitly preserved.

## Frozen baseline and chronology

| Item | Evidence |
| --- | --- |
| V1 actual base | `7631e9fe62a8be2a527d12c608c8539771dd3277` |
| V1 PR | #244, OPEN, DRAFT, UNMERGED; head `29ad44cadfdd8f072b65da0235de540b34afb462` |
| V1 result | STOPPED_ON_2_P2 |
| Remediation | PR #245, merged |
| Remediation merge / V2 base | `eba69ae32dbe810fb3436fe2f6691777d7cbfe7d` |
| V2 base tree | `c4e8addd305489b4faf09420dd87e86207767e91` |
| Base branch | `develop`, independently fetched, protected |
| Exact post-merge CI | [38073089322](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38073089322), workflow_dispatch, SUCCESS, 10/10 |
| V2 branch | `audit/final-production-frontend-infra-hardening-v2` |

The supplied chronology contains `V1_BASE=7631e9fe62a8be810fb3436fe2f6691777d7cbfe7d`. That value is not the verified V1 base. The actual V1 base above is independently evidenced by the original baseline and the remediation merge parent. The original supplied string is preserved here without adopting it as a valid commit.

V1 recorded two test/harness defects. Runs 37959604010 and 37960141249 hit the earlier cancellation-fixture timeout; 37960770615 had workflow composition failure; 37960928767 was cancelled. Corrected specialized run 37961449087 and normal run 37961448982 were recorded by V1. These are historical events, not new product findings and not substitutes for V2 CI.

## Verification method and collect-all behavior

The added `scripts/audit/module12_v2_verify.py` executes independent existing test groups, records each return code and log, continues after a group failure, writes a JSON summary and fails after collection. It does not introduce new offensive fixtures. The specialized workflow checks out the exact PR head, runs backend and frontend independently, and uploads their evidence with `always()` even after failures. The normal protected CI remains unchanged.

Local backend: Go 1.26.9, Redis 7.0.15 on loopback. Specialized CI uses Redis 8 Alpine. Local frontend: Node 24.19.0, pinned pnpm 11.8.0 and Next 16.3.8. Specialized CI pins Node 22.22.2. Redis is a real server, not a process-local limiter mock. Local full Go and race suites pass; database-dependent tests skipped locally because PostgreSQL was unavailable. Local skips are not credited as PostgreSQL verification. The normal CI provides its configured real PostgreSQL environment.

No external production origin, provider credentials, CDN account, container runtime, production log aggregator or production database topology was available. No production-only guarantee is derived from local tests.

The final audit head, specialized run identifiers and normal exact-head CI result are returned in the final stop report after publication. Historical post-merge CI is only baseline evidence.

## A: shared budgets — PASS for repository-owned topology

Existing shared HTTP, shared provider and rolling-window groups all passed under the race detector.

| Property | Result |
| --- | --- |
| HTTP global, 1 / 2 / 4 replicas | 48 / 48 / 48 |
| Same client across replicas | 12 |
| Provider, 1 / 2 / 4 instances | 60 / 60 / 60 |
| Restart resets shared budget | NO |
| Shared backend unavailable | FAIL_CLOSED; provider calls 0 |
| Rolling global / client / provider | PASS / PASS / PASS |
| Boundary burst global / client / provider | BLOCKED_TO_ROLLING_LIMIT |

These are repository-controlled instances sharing the test Redis authority, not an observed multi-host production deployment or provider-account quota. No Redis implementation changes were made.

## B: cancellation and transport — PASS for existing bounded regressions

The following existing regressions all passed independently: `TestNoncooperativeUpstreamCannotRecycleCapacityOnCallerCancellation`, `TestDetachedOperationRetainsRealSocketCapacity`, `TestDetachedOperationHardTimeoutStillTerminatesLocalSocket`, `TestRawDownstreamDisconnectRetainsLocalOperationOwnership`, `TestCancelledProviderWorkRemainsBoundedByConcurrency`, and `TestProviderPartialResetTimeoutStillFailClosedAndRecover`.

| Hostile real HTTP fixture measurement | Result |
| --- | --- |
| First remote handlers / cancelled callers | 4 / 4 |
| Callers return promptly | YES |
| Observation delay | 500 ms, before independent five-second hard timeout |
| Semaphore / local operations after caller cancellation | 4 / 4 |
| Replacement attempts / reaching remote | 4 / 0 |
| Remote total started / maximum active | 4 / 4 |
| Semaphore / local operations after explicit release | 0 / 0 |
| Subsequent legitimate request | PASS |
| Goroutines baseline / recovery | 3 / 3 |
| File descriptors baseline / recovery | 8 / 8 |

The detached cooperative real-socket campaign records 40 caller cancellations, zero replacements before local terminal, maximum local transports 4, maximum semaphore 4, goroutines 3 / 27 / 3 and file descriptors 8 / 16 / 8 (baseline / peak / recovery). The custom-lifetime campaign records 40 cancellations, maximum local transports 4, semaphore 4, goroutines 2 / 10 / 2 and file descriptors 7 / 7 / 7. These are separate campaigns; their counters are not merged into one fabricated measurement.

The raw TCP regression records four disconnects, occupied semaphore 4, upstream handlers 4, subsequent occupied request 503 and post-terminal recovery 200. `DOWNSTREAM_DISCONNECT_CONTEXT_PROPAGATION=NO` for the Fiber/fasthttp boundary. Partial response, connection reset and upstream timeout recovery all PASS.

`M12_P2_02_CALLER_CANCELLATION_AMPLIFICATION=REMEDIATED_REPOSITORY_SIDE`; `CALLER_CANCELLATION_CANNOT_RECYCLE_PROVIDER_CAPACITY_EARLY=YES`; `LOCAL_PROVIDER_TRANSPORT_CONCURRENCY_BOUNDED=YES` for these fixtures. `REAL_EXTERNAL_PROVIDER_SERVER_SIDE_CANCELLATION=NOT_VERIFIED`; `REMOTE_PROVIDER_SERVER_SIDE_TERMINATION=NOT_VERIFIED`. A remote server may continue computation after the client's independent hard timeout. This is not a claim of arbitrary remote termination.

## C–L: repository and local runtime review

| Surface | Evidence and remaining limit |
| --- | --- |
| Trusted proxy / forwarding | Existing full Go suites exercise direct/trusted proxy configuration. `http_server_config.go` validates trusted CIDRs, rejects empty/full-family trust, and enables IP validation with explicit trust. Direct mode does not configure proxy-header identity. Full new duplicate/conflicting/mixed-family spoofing matrix NOT_EXECUTED; matrix result PARTIALLY_VERIFIED. |
| Browser auth and CSRF | Existing backend auth/CSRF and frontend tests pass. Refresh cookies use the auth path, HttpOnly, Strict SameSite and configured Secure; no Domain is set. Frontend tests are not a deployed browser-origin race campaign. New real-browser concurrent refresh/logout and isolated-principal matrix NOT_EXECUTED; PARTIALLY_VERIFIED. |
| Real sockets | Existing bounded tests above PASS. New exhaustive idle/slow-drip/half-close/disconnect-position matrix NOT_EXECUTED; PARTIALLY_VERIFIED. |
| Logging and redaction | Source review does not establish fault-path redaction. Build-input sentinels below are not a logging test. No success/failure log-sentinel campaign or production aggregator test was executed. RAW_ACCESS_TOKEN / RAW_REFRESH_TOKEN / RAW_PROVIDER_TOKEN / RAW_DATABASE_PASSWORD / RAW_REDIS_PASSWORD remain NOT_VERIFIED. |
| Cache controls | Repository sensitive response policy sets no-store for auth and portfolios; corporate-action response policy also uses no-store. Existing HTTP tests pass. Local public Next root returns s-maxage=31536000, no Set-Cookie and no redirect; 404 returns private/no-cache/no-store. New variation/poisoning campaign and CDN behavior NOT_EXECUTED; PARTIALLY_VERIFIED. |
| Infrastructure | Repository inventory found docker-compose.yml, no application production Dockerfile/ingress/Kubernetes/ECS/Terraform manifest. Actual container runtime policy NOT_VERIFIED. |
| PostgreSQL | Store pool MaxOpen=10, MaxIdle=5, MaxLifetime=30 minutes. DSN is supplied; code review does not establish mandatory production TLS. Repository role/migration tests are part of normal CI, local PostgreSQL tests skipped. Actual production role/firewall/TLS topology NOT_VERIFIED. |
| CORS | cors.go matches configured allowed origins exactly and enables credentials only for an allowed origin. An unset origin configuration defaults to localhost:3000 and 127.0.0.1:3000. Existing CORS tests pass. A new duplicate-Origin/Host runtime matrix NOT_EXECUTED. |
| Host | Historical V1 arbitrary-Host observation remains historical evidence. This continuation does not claim a newly executed comprehensive Host/forwarded-host matrix. Production routing NOT_VERIFIED. |
| Error / health | Existing Go error/auth/health suites pass; health remains a liveness boundary. Local frontend root/404 observed. The new complete 400/401/403/404/409/429/500/502/503 fault-injection response-and-log matrix NOT_EXECUTED. No blanket secret-leak absence claim. |

### Local production frontend headers

Observed on normal `next start` over loopback, root 200 and missing-page 404:

| Header | Root and 404 |
| --- | --- |
| Content-Security-Policy | `frame-ancestors 'none'` |
| X-Frame-Options | DENY |
| Strict-Transport-Security | ABSENT |
| X-Content-Type-Options | ABSENT |
| Referrer-Policy | ABSENT |
| Permissions-Policy | ABSENT |
| X-Powered-By | ABSENT |

This revalidates carried HARDENING_ONLY H01. It does not establish external TLS termination or edge HSTS. H02 (unset allowed-origin configuration falls back to development origins) is still present by source review. H03 (Host acceptance/routing boundary) retains the historical classification without converting incomplete new runtime coverage to PASS. No new severity classification is invented for these already recorded observations.

### Frontend public artifact exposure

Production typecheck, frontend tests, pnpm audit and build PASS. Twenty public static files scanned; `SOURCE_MAP_COUNT=0`; synthetic private build-input sentinel matches in public static files=0 and root/404 response bodies=0. Eight distinct synthetic inputs cover token, cookie, bearer, provider, database-password, Redis-password, import-review and financial-payload labels. They are synthetic build inputs, not real logged HTTP credentials or live financial data.

Source search found zero localStorage token references, zero sessionStorage token references and zero document.cookie token references in frontend source. sessionStorage is used for idempotency keys; the claim is about token references, not absence of all browser storage. No provider/account credential was used.

### Container and production boundaries

| Requested field | Repository evidence |
| --- | --- |
| APPLICATION_PRODUCTION_CONTAINER_MANIFEST | ABSENT |
| PRIVILEGED | Not configured in Compose; actual runtime NOT_VERIFIED |
| HOST_NETWORK | Not configured in Compose; actual runtime NOT_VERIFIED |
| RUNTIME_USER | Application production container ABSENT / NOT_VERIFIED |
| READ_ONLY_FILESYSTEM | Application policy ABSENT / NOT_VERIFIED |
| CAPABILITIES | Application policy ABSENT / NOT_VERIFIED |
| SECCOMP_APPARMOR | Application policy ABSENT / NOT_VERIFIED |
| PORT_BINDINGS | Compose PostgreSQL and Redis bind loopback; deployed bindings NOT_VERIFIED |
| SECRET_BUILD_ARGS | No application production build manifest found; actual build environment NOT_VERIFIED |

No clearly documented deployed public application origin or repository-owned production CDN/LB/ingress/TLS termination configuration was found in the reviewed deployment inventory. `REAL_PRODUCTION_EDGE_CONFIGURATION=NOT_PRESENT_IN_REPOSITORY` for that inventory. No DNS/CDN/firewall changes or paid external requests occurred.

## Complete V1 residual disposition

| Residual | Final status | Evidence collected | What remains unproven |
| --- | --- | --- | --- |
| NV-M12-01 production CDN/LB/proxy topology and trusted CIDRs | NOT_VERIFIED | Repository proxy settings reviewed; existing tests pass | Actual edge topology and deployed CIDRs |
| NV-M12-02 TLS termination / redirect / HSTS | NOT_VERIFIED | Local Next headers recorded | Production TLS, HTTP→HTTPS and edge HSTS |
| NV-M12-03 provider-account quota | NOT_VERIFIED | Repository shared budgets and rolling windows PASS | Account-wide external provider authority |
| NV-M12-04 provider cancellation/recovery | NOT_VERIFIED | Local independent operation ownership and recovery PASS | External server-side termination and account-side recovery |
| NV-M12-05 multi-host routing / rolling restart | NOT_VERIFIED | Controlled 1/2/4 instances and preserved shared state PASS | Actual multi-host/container routing and rolling deployment |
| NV-M12-06 container security policy | NOT_VERIFIED | Compose and artifact inventory reviewed | Actual privileges, capabilities, seccomp, filesystem and network policy |
| NV-M12-07 production logging/redaction | NOT_VERIFIED | Limited source review | Fault-path sentinel redaction and production aggregation |
| NV-M12-08 forwarding-header matrix | PARTIALLY_VERIFIED | Existing proxy/config tests PASS | Requested exhaustive duplicate/conflicting spoofing matrix |
| NV-M12-09 browser auth/CSRF race | PARTIALLY_VERIFIED | Existing backend and frontend tests PASS | Real browser/deployed origins and full concurrent-principal matrix |
| NV-M12-10 socket timing matrix | PARTIALLY_VERIFIED | Existing bounded real-socket regressions PASS | Exhaustive downstream/upstream timing combinations |
| NV-M12-11 production DB topology | NOT_VERIFIED | Pool bounds and repository role scripts reviewed; CI tests retained | Deployed TLS/firewall/network/role settings |
| NV-M12-12 cache/CDN poisoning | PARTIALLY_VERIFIED | Sensitive no-store policy and local Next responses inspected | New poisoning campaign and actual CDN behavior |

Counts: eight NOT_VERIFIED and four PARTIALLY_VERIFIED residuals; no residual silently omitted. These counts concern the twelve-entry residual table, not a count of every possible untested case.

Module 10 external shared provider budget, full API multi-instance chaos and exhaustive real-socket residuals remain unverified where external/full-topology evidence is required. Module 11 RLSA_01 real edge/gateway production authority remains unverified. Repository Redis evidence does not retroactively prove those production boundaries.

## Scanners and stop gate

Local `go test ./...`, `go test -race ./...` and `go vet ./...` pass with the stated local PostgreSQL skips. Local frontend tests/typecheck/build/audit pass. Exact final-head normal CI must independently supply its ten job outcomes: Go tests, Python tests, frontend build/typecheck, OpenAPI, Compose, PostgreSQL migrations, vet, race, Go vulnerability scan and dependency security scan. No historical run is credited as final V2-head CI.

Known non-reachable advisory is retained, not suppressed: GO-2026-5932, golang.org/x/crypto@v0.57.0, package golang.org/x/crypto/openpgp. Reachability is determined by the exact-head govulncheck log, not assumed from the module advisory. Final run IDs and scanner outcomes are reported separately after those jobs finish.

No new product P0/P1/P2/P3 finding was demonstrated within this scoped verification. Full requested continuation remains incomplete; unresolved severity totals across unexecuted matrices are NOT_ESTABLISHED. New harness defects must be reported separately from product findings if CI reveals one. Two V1 test defects remain historical, not new V2 defects.

`REMEDIATION_AUTHORIZED=NO`; `APPROVE_FOR_MERGE=NO`; `FORMAL_MODULE_12_CLOSED=NO`; `FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO`.
