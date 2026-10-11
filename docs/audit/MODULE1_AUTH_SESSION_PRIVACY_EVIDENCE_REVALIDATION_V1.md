# Module 1 Authentication / Session / Privacy Evidence Revalidation V1

```text
MODULE=1_AUTHENTICATION_SESSION_PRIVACY
MODE=DETECTION_AND_EVIDENCE_ONLY
BASE_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3
BASE_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba
BRANCH=develop
PROTECTED=YES
REQUIRED_CONTEXTS=10
AUDIT_BRANCH=audit/module1-auth-session-privacy-evidence-revalidation
PRODUCT_FILES_CHANGED=NONE
TEST_FILES_CHANGED=NONE
WORKFLOW_FILES_CHANGED=NONE
MODULE_1_CLOSURE_READY=NO
FORMAL_MODULE_1_CLOSED=NO
```

## Historical reconstruction and contradictory evidence

The frozen protected baseline and all ten required contexts were independently
read from GitHub before creating the isolated audit branch. The canonical
[Auth closure](AUTH_SESSION_PRIVACY_ADVERSARIAL_AUDIT_CLOSURE.md), merged through
[PR #205](https://github.com/AsifAbbasov/OpenInvest/pull/205), says
`CLOSED_WITH_P3_RESIDUAL`, not `REOPENED_FOR_EVIDENCE_REVALIDATION`.
PR #205's body independently retains that disposition. Its issue-comment
collection is empty. The checked local git history contains the merged closure
`013ff7f`; no reopening record was located in the reviewed repository documents.

[PR #203](https://github.com/AsifAbbasov/OpenInvest/pull/203) remediated
AUTH-ADV-01 P2 durable authentication-audit storage amplification. The technical
squash is `40010f4c7513dab99a31ab85c879242a77228b7a`; PR #204 recorded its
closure. PostgreSQL deduplication bounds known-session security evidence by
action/session, while unknown/anonymous rejection paths write no durable audit
rows. These are historical accepted results, not new long-running proof.

AUTH-ADV-02 remains documented as P3 deferred product-security hardening:
registration membership enumeration, namespace/account squatting and missing
email ownership verification. Stage 3.11 explicitly excluded SMTP/email
verification. This audit does not erase or reclassify that residual, and does
not assert P3=0. Its present independent dynamic classification remains pending.

The supplied historical statement `CONFIRMED_PRODUCTION_BYPASS=NONE` is not
proof of universal absence. No new bypass was demonstrated in this safe review;
the requested broader bypass revalidation was not executed.

## Historical long-running harness reconstruction

Initial repository-only searches did not identify the original harnesses.
A read-only historical Actions search for October 5–7 recovered
[failed run 37463802252](https://github.com/AsifAbbasov/OpenInvest/actions/runs/37463802252)
at `de968102eb754bbda678988c7a28c67b9e1560de`, and
[successful later run 37506260387](https://github.com/AsifAbbasov/OpenInvest/actions/runs/37506260387)
at `922f520fb0c817784ad62c69fd40f9df40391ecd`.
The workflow is `.github/workflows/auth-session-theft-audit-24h.yml`,
on the historical `chore/native-go-fuzz-targets` branch. It is absent from
the frozen develop tree. These are historical runs, never new exact-head proof.

### Harness 1 — cookie/replay HTTP soak

```text
HARNESS_ID=cookie-csrf-theft-soak
ORIGINAL_PURPOSE=HTTP_COOKIE_CSRF_FIXATION_AND_LOGOUT_SOAK
FAILED_RUN=37463802252
FAILED_JOB=112269649064
FAILED_TEST=TestAuthTheftRotatedAndLoggedOutCookiesCannotBeReused/rotation_replay
OBSERVED_FAILURE=stolen pre-rotation pair replay 19 accepted: status=429
HARNESS_DEFECT_OR_PRODUCT_DEFECT=HARNESS_STATUS_ORACLE_DEFECT
CORRECTED_OR_REPLACED=HISTORICALLY_CORRECTED_NOT_IMPORTED_HERE
LATER_RUN=37506260387
LATER_JOB=112415580129
FINAL_HISTORICAL_EXECUTION_RESULT=SUCCESS
CURRENT_BASELINE_SPECIALIZED_RESULT=NOT_EXECUTED
```

The original test required every repeated rejection to be 401. Rate limiting
returned 429, which the diagnostic incorrectly called acceptance. The later
fixture accepts 401/429 during its bounded series, advances a controlled limiter
clock, then independently requires 401 after reset for both rotated and logged-out
credentials. This additional assertion prevents 429 from masking resurrection.
Historical successful output reached iteration 85; each workflow iteration ran
the HTTP test family under race detection with count=100. It is a five-hour job
inside a workflow named “24h”, not 24 hours of sequential coverage.

A further run `37503131810`, job `112404935796`, failed compilation:
`auth_theft_adversarial_test.go:148:17: undefined: verticalslice`.
This is another harness failure, not a product bypass. Later source includes
the dependency; no fixture repair was performed or imported by this audit.

### Harness 2 — signed-bearer mutation fuzz

```text
HARNESS_ID=jwt-tamper-fuzz
ORIGINAL_PURPOSE=REJECT_MUTATED_SIGNED_BEARER_AND_ALGORITHM_CONFUSION
FAILED_RUN=37463802252
FAILED_JOB=112269648983
FAILED_FUZZ=FuzzAuthTheftMutatedSignedBearerNeverAuthenticates
OBSERVED_FAILURE=mutated signed bearer authenticated: index=328 xor=103
CORPUS_ID=559e27f9a93bb196
HARNESS_DEFECT_OR_PRODUCT_DEFECT=ORACLE_SEMANTICS_REVISED_CANONICALIZATION_RESIDUAL
CORRECTED_OR_REPLACED=HISTORICALLY_REVISED_NOT_IMPORTED_HERE
LATER_RUN=37506260387
LATER_JOB=112415579998
FINAL_HISTORICAL_EXECUTION_RESULT=SUCCESS_FOR_REVISED_PROPERTY
CURRENT_BASELINE_SPECIALIZED_RESULT=NOT_EXECUTED
```

The original oracle rejected every textual change. Later source permits a token
only when its decoded segments remain identical; any authenticated semantic
change still fails. The later five-hour fuzz logged 916077214 executions and PASS.
This proves the revised cryptographic-distinction property for that historical
run, not the original stronger textual-canonicality property. A separate
`TestSecurityBreakerJWTRejectsNonCanonicalBase64Aliases` exists in the later
source but skips unless OPENINVEST_SECURITY_BREAKER_TESTS=1. The reviewed
workflow does not set that flag, so its success does not establish canonical
alias rejection. This residual needs independent classification; no new dynamic
reproduction or product finding is asserted here.

The historical workflow also had two successful PostgreSQL jobs for session
race and family containment. Four of four success is verified from job metadata;
none is credited to the present audit head. Resilience Chaos Audit is a separate
workflow, not one of these two auth failures.

```text
HISTORICAL_BROKEN_HARNESSES_REPORTED=2
HISTORICAL_BROKEN_HARNESSES_FOUND=2_AUTH_FAILURES_IDENTIFIED
HISTORICAL_BROKEN_HARNESSES_HISTORICALLY_CORRECTED=2_WITH_ORACLE_SCOPE_CAVEAT
HISTORICAL_BROKEN_HARNESSES_RESOLVED_ON_CURRENT_BASELINE=0
BROKEN_REQUIRED_HARNESSES=2_CURRENT_BASELINE_REVALIDATION_PENDING
FORMAL_CLOSURE_PENDING_EVIDENCE_REVALIDATION=YES
```

The supplied reopening decision is still not a canonical repository closure
amendment located by this review. Historical success and fixture corrections
resolve the immediate old job failures; the unexecuted current-baseline property
matrix and JWT canonicalization disposition continue to block closure readiness.

## Safe local runtime verification

Existing unmodified Go tests ran on the frozen baseline with Go 1.26.9 and
`go test -race -v`, count=1, bounded package timeouts of 90/60 seconds:

- TestServiceRegisterCreatesPrivacyFirstSession
- TestServiceRequiresExplicitAccessTokenSecretOrEphemeralFlag
- TestServiceLoginWrongPasswordRemainsInvalidCredentials
- TestServiceLoginPropagatesInfrastructureFailure
- TestAuthRegisterSetsHttpOnlyRefreshCookieWithoutBodyLeak
- TestAuthLogoutClearsRefreshCookie
- TestAuthInfrastructureFailuresUseSanitizedInternalError
- TestCORSOriginCompatibility
- TestSensitiveResponseCachePolicyCoversAuthLifecycleAndErrors

All selected tests passed; no data race was observed. HTTP tests exercise existing
Fiber handlers through the repository's normal local test fixtures. They use
synthetic credentials and an in-memory auth store, not deployed PostgreSQL.
Compatibility checks observed Secure=true, HttpOnly=true, SameSite=Strict,
registration 201 and Cache-Control=no-store with ordinary Host/forwarded inputs.
They are not an adversarial forwarding-identity matrix.

## Property inventory and proof limits

| Requested boundary | Evidence / present status |
| --- | --- |
| Login/register/session issuance | Ordinary registration and invalid-login regressions PASS_IN_TESTED_SCOPE |
| Authentication bypass / fixation | Static configuration and token generation reviewed; comprehensive dynamic revalidation NOT_EXECUTED |
| Refresh lifecycle / rotation / replay | Existing service and PostgreSQL regression sources inventoried; specialized repeated-sequence revalidation NOT_EXECUTED |
| Logout/revocation | Cookie clearing PASS_IN_TESTED_SCOPE; comprehensive persistence/family containment not independently specialized here |
| CSRF | Existing service/HTTP checks and stored CSRF hashes inventoried; full cross-context/browser matrix NOT_EXECUTED |
| Cookie semantics | Secure/HttpOnly/Strict and ordinary compatibility PASS_IN_TESTED_SCOPE; production configuration remains external |
| Session/account and horizontal/vertical isolation | Relevant subject/store boundaries inventoried; cross-principal authorization matrix NOT_EXECUTED |
| Concurrent refresh/session state | Existing PostgreSQL races inventoried; historical failures located; current-baseline specialized proof pending |
| Enumeration | Historical AUTH-ADV-02 P3 deferred disposition preserved; not silently converted to PASS |
| Password reset/recovery | No such route in registerRoutes; NOT_PRESENT_IN_REVIEWED_ROUTE_TABLE, not a claim about external services |
| Brute-force/rate interaction | Existing admission tests inventoried; new load/abuse campaign NOT_EXECUTED |
| Malformed/duplicate/conflicting auth material | Parser/getter source reviewed; raw ambiguity matrix NOT_EXECUTED |
| Proxy/Forwarded/Host | Ordinary cookie/no-store compatibility PASS; comprehensive identity variation NOT_EXECUTED |
| Privacy-sensitive cache | Existing auth lifecycle/error no-store regression PASS_IN_TESTED_SCOPE |
| Error/log leakage | Sanitized infrastructure error body regression PASS; historical AUTH_23 log review preserved, no new complete diagnostic-sink proof |
| PostgreSQL persistence | Normal CI supplies ordinary service-container regression execution; historical specialized proof reviewed; current-baseline specialized proof pending |

New autonomous bypass, cross-session credential misuse, malicious replay/race,
raw authentication ambiguity and abuse reproduction matrices were not executed
under this executor's policy. Ordinary permanent regression results do not
substitute for those missing campaign results. No tests/assertions were changed,
no production service was targeted and no product remediation was attempted.

## Exact-head CI and independent stop gate

This report-only audit commit opens a draft PR to develop. The existing normal
CI must run on its exact head; run IDs and completed job results are returned
separately after completion. No prior run is substituted. Normal CI success
cannot establish execution of the missing specialized harnesses.

```text
SPECIALIZED_RESULT=HISTORICAL_SUCCESS_CURRENT_HEAD_NOT_EXECUTED_EXECUTOR_POLICY
REQUIRED_AUTH_SESSION_PRIVACY_PROPERTIES_REVALIDATED=NO
NEW_FINDINGS=NONE_DEMONSTRATED_IN_EXECUTED_SCOPE
JWT_CANONICALIZATION_RESIDUAL=REQUIRES_INDEPENDENT_DISPOSITION
P0=NOT_FULLY_REVALIDATED
P1=NOT_FULLY_REVALIDATED
P2=NOT_FULLY_REVALIDATED
P3=HISTORICAL_AUTH_ADV_02_DEFERRED_REQUIRES_DISPOSITION
MODULE_1_CLOSURE_READY=NO
FORMAL_MODULE_1_CLOSED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO
REMEDIATION_AUTHORIZED=NO
APPROVE_FOR_MERGE=NO
STOP_FOR_INDEPENDENT_REVIEW=YES
```
