# Module 1 Authentication / Session / Privacy — Independent Adversarial Audit V2

## A. Executive Security Verdict

This audit used the protected baseline required by the audit directive and added only audit tests/workflow evidence on a separate audit branch. No production implementation was changed, PR #249 was not modified, protected `develop` was not modified, no pull request was merged, the five-hour campaign was not started, and no final repository-wide assault was started.

Repository-controlled authentication/session properties were challenged with deterministic tests, raw HTTP ambiguity fixtures, PostgreSQL concurrency/state oracles, frontend auth-state race tests, focused Go fuzzing and negative controls.

### High-level result

- No P0, P1 or P2 authentication/session/privacy bypass was reproduced under the tested conditions.
- Historical `AUTH-ADV-02` remains confirmed P3: registration exposes direct new-vs-existing email differentiation, creates an immediately active account without email ownership verification, and reserves the address namespace.
- Three new P3 hardening/security findings were reproduced:
  1. noncanonical byte-equivalent JWT signature encoding is accepted;
  2. conflicting duplicate Authorization / refresh-cookie representations are order-sensitive and can authorize when the valid value is first;
  3. duplicate JSON keys in authentication registration input are accepted instead of being rejected as ambiguous.
- Signed JWTs containing duplicate critical JSON claims or an unknown `crit` header are also accepted. The behavior is verified, but no untrusted signing path exists in the tested architecture; this is recorded as hardening-only rather than counted as an additional exploitable finding.
- Byte-changing JWT mutation, malformed JWT structure, session fixation, cross-principal reads/mutations, CSRF/session mixing, refresh/logout concurrency, replay/family containment, frontend transient-failure handling and stale-operation generation handling did not reproduce a security violation under the tested conditions.
- Focused fuzzing completed 5,061,153 executions over four targets / 120 seconds with no fuzz-discovered invariant failure.

The specialized workflow is expected to remain red while confirmed adversarial assertions remain unremediated. A red specialized job is not automatically a harness defect; in the final pre-report execution the red jobs are the deterministic JWT and raw-HTTP ambiguity groups because their product invariants are intentionally violated.

## B. Baseline Integrity

```text
REPOSITORY=AsifAbbasov/OpenInvest

EXPECTED_DEVELOP_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3
OBSERVED_DEVELOP_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3

EXPECTED_DEVELOP_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba
OBSERVED_DEVELOP_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba

DEVELOP_PROTECTED=YES
REQUIRED_CONTEXTS=10

PR_249_STATE=OPEN_DRAFT
PR_249_MERGED=NO
PR_249_HEAD_SHA=89acc03102dd745a77c5c86391f40d22d06ab577
PR_249_HEAD_TREE=e1a2f34f0c947a2612ac850d873593c76cdeb336

BASELINE_CI_RUN_ID=38181761847
BASELINE_CI_RESULT=SUCCESS_10_OF_10
```

The ten baseline CI jobs were independently re-fetched and all were successful: Docker Compose config, OpenAPI contract, Go vulnerability scan, Go vet, Python tests, Frontend build and typecheck, PostgreSQL migration validation, Dependency security scan, Go tests and Go race tests.

Audit work was performed on:

```text
AUDIT_BRANCH=audit/module1-auth-session-privacy-adversarial-v2
PRE_REPORT_TESTED_HEAD_SHA=17fbe6b2f9959b5976de606abaf2375291a2d62d
PRE_REPORT_TESTED_HEAD_TREE=22c29b6af86626027bdf4853a7981143c249ae7c
SPECIALIZED_RUN_ID=38322758278
```

A report-only commit changes the audit tree and therefore requires another exact-head specialized run before final Gatekeeper handoff. The final external STOP report must use that post-report run and must not substitute the pre-report run above.

## C. Attack Surface and Threat Model

Tested repository-owned boundaries:

- access-token parsing and authentication;
- malformed/duplicate Authorization representation;
- refresh cookie and CSRF binding;
- refresh rotation and replay;
- logout revocation;
- PostgreSQL session-family state;
- high-contention refresh races;
- cross-principal portfolio read/mutation isolation;
- idempotency-key principal binding;
- private authenticated cache policy;
- registration enumeration and namespace squatting;
- email normalization;
- frontend refresh/logout operation generation;
- sanitized auth/storage error boundary;
- tested response/logger secret sentinels.

Not claimed:

- production reverse-proxy/CDN normalization;
- third-party log aggregation;
- real stolen user credentials;
- production sessions;
- production database/network behavior;
- external identity provider behavior.

## D. Security Property Matrix

| Property | Attack / oracle | Result |
| --- | --- | --- |
| Anonymous protected-route access | 15 protected route method/path combinations; non-2xx required | PASS; 15/15 returned 401 |
| Authorization ambiguity | duplicate equal/conflicting raw fields | FAIL for conflicting valid-first; order-sensitive first-value selection |
| Session fixation | attacker-selected pre-auth refresh cookie before register/login | PASS; fresh server-generated cookie issued |
| Refresh-cookie ambiguity | duplicate equal and conflicting cookies + CSRF | FAIL for valid-first/invalid-last; invalid-first/valid-last rejected |
| CSRF/session binding | A refresh cookie + B CSRF, DB active-session count | PASS; 401 and DB state unchanged |
| Refresh replay | service and PostgreSQL family state | PASS |
| Logout revocation | service + PostgreSQL race state | PASS |
| Refresh/logout concurrency | PostgreSQL races + persistent state | PASS |
| High-contention refresh | 2/4/8/16/32 contenders × 25 repetitions | PASS; one winner per fixture, then containment |
| Cross-principal read | B token → A portfolio | PASS; no 2xx / no A body returned |
| Cross-principal mutation | B token → A transaction + DB row count | PASS; no mutation persisted |
| Idempotency principal binding | same idempotency key under A and B | PASS; independent resources |
| Privacy cache | authenticated private response | PASS; no-store |
| JWT malformed structure | deterministic malformed matrix + fuzz | PASS |
| JWT byte-changing tamper | HMAC verification + 1,505,040 fuzz executions | PASS |
| JWT canonical textual encoding | byte-equivalent signature alias | FAIL; noncanonical alias accepted |
| JWT duplicate critical JSON claims | validly signed duplicate alg/uid/exp | HARDENING FAIL; no untrusted signing path established |
| JWT unknown critical header | validly signed unknown crit/kid | HARDENING FAIL; no untrusted signing path established |
| JWT expiration boundary | injected controlled clock at exp boundary | PASS |
| Email normalization | deterministic policy + 527,775 fuzz executions | PASS |
| Registration enumeration | new vs duplicate synthetic email response | CONFIRMED historical P3 |
| Email ownership | immediate active account and login | ABSENT |
| Pre-registration squatting | registration reserves synthetic email namespace | CONFIRMED |
| Timing enumeration | 20 samples/class | MEASURED_NOT_CLASSIFIED; distributions overlap |
| Frontend 429/5xx/network handling | state-transition oracle | PASS |
| Frontend concurrent refresh | single-flight oracle | PASS |
| Frontend stale refresh/logout after principal transition | real AuthShell generation contract | PASS |
| Auth error leakage | synthetic response and standard-logger sentinels | PASS for tested paths |
| Auth/storage error classification | historical OI-NEW-02 regression ×20 | PASS |

## E. Adversarial Test Execution

### New audit files

```text
backend-go/internal/auth/module1_adversarial_jwt_test.go
backend-go/internal/auth/module1_session_state_machine_test.go
backend-go/internal/httpapi/module1_adversarial_auth_http_test.go
backend-go/internal/postgres/module1_adversarial_session_integration_test.go
backend-go/internal/postgres/module1_cross_principal_integration_test.go
frontend-next/tests/module1-auth-state-adversarial.test.mjs
.github/workflows/module1-auth-adversarial-v2.yml
```

### Focused fuzz campaigns

```text
FuzzModule1JWTMalformedStructureNeverAuthenticates
DURATION=30s
EXECUTIONS=2,076,359
RESULT=PASS

FuzzModule1JWTByteChangingPayloadMutationNeverAuthenticates
DURATION=30s
EXECUTIONS=1,505,040
RESULT=PASS

FuzzModule1EmailNormalizationIdempotent
DURATION=30s
EXECUTIONS=527,775
RESULT=PASS

FuzzModule1RefreshLogoutStateMachine
DURATION=30s
EXECUTIONS=951,979
RESULT=PASS

FUZZ_EXECUTIONS_TOTAL=5,061,153
FUZZ_DURATION_TOTAL=120s
```

### PostgreSQL contention

```text
CONTENDERS=2,4,8,16,32
REPETITIONS_PER_LEVEL=25
TOTAL_REFRESH_CALLS=1550

CONTENDERS=2   SUCCESS=25 AUTH_FAILURE=25
CONTENDERS=4   SUCCESS=25 AUTH_FAILURE=75
CONTENDERS=8   SUCCESS=25 AUTH_FAILURE=175
CONTENDERS=16  SUCCESS=25 AUTH_FAILURE=375
CONTENDERS=32  SUCCESS=25 AUTH_FAILURE=775

ACTIVE_SESSIONS_AFTER_EACH_FIXTURE=0
GOROUTINES_BASELINE=4
GOROUTINES_RECOVERY=5
MAX_OBSERVED_DB_CONNECTIONS=6
```

Existing PostgreSQL tests also passed:

```text
TestLogoutRacingRefreshCannotLeaveDescendantActive
TestReplayRacingDescendantRefreshLeavesNoActiveFamilyMember
TestConcurrentRefreshReplayRevokesWinningSessionFamily
```

### Registration timing evidence

```text
SAMPLES_PER_CLASS=20

NEW_P50=approximately 264ms
NEW_P95=approximately 295ms

DUPLICATE_P50=approximately 261ms
DUPLICATE_P95=approximately 311ms

AUTH_ADV_02_TIMING_ENUMERATION=MEASURED_NOT_CLASSIFIED
```

Direct semantic enumeration does not depend on timing and remains confirmed.

## F. Confirmed Findings

### AUTH-ADV-02 — P3 — Historical finding revalidated

**Root cause:** registration creates an active account without email ownership verification and returns an explicit conflict for a duplicate normalized email.

**Observed result:**

```text
NEW_EMAIL=201
DUPLICATE_EMAIL=409
DUPLICATE_MESSAGE="Email is already registered"
EMAIL_OWNERSHIP_VERIFICATION=ABSENT
ACCOUNT_STATE=active
IMMEDIATE_LOGIN=SUCCESS
PRE_REGISTRATION_SQUATTING=CONFIRMED
```

**Impact:** direct account enumeration and reservation of a registration namespace without proving control of the email address.

**Remediation:** independent product decision; not authorized in this audit.

### M1-P3-01 — P3 — Noncanonical JWT textual alias accepted

**Root cause:** `verifyAccessToken` decodes the signature with RawURLEncoding and compares decoded signature bytes, but does not require the encoded signature segment to equal its canonical re-encoding.

**Attack prerequisite:** possession of an already valid access token.

**Observed result:** a byte-equivalent noncanonical Base64URL signature spelling is accepted.

**Impact:** no signature forgery or cross-principal access was reproduced. The weakness permits multiple textual identities for the same bearer token and can undermine string-identity-based logging, caching or deny-list controls if such controls are introduced/deployed.

**Regression requirement:** reject noncanonical Base64URL spellings before authentication.

### M1-P3-02 — P3 — Conflicting duplicate authentication representations are order-sensitive

**Root cause:** HTTP auth code obtains one value via `c.Get("Authorization")` and refresh/logout obtain one named cookie via `c.Cookies(...)`; conflicting physical duplicates are not rejected as ambiguous before security decisions.

**Observed result:**

```text
AUTHORIZATION_VALID_FIRST_INVALID_LAST=200
AUTHORIZATION_INVALID_FIRST_VALID_LAST=401

REFRESH_COOKIE_VALID_FIRST_INVALID_LAST=200
REFRESH_COOKIE_INVALID_FIRST_VALID_LAST=401
```

Equal duplicate values are accepted and are not classified as a finding.

**Impact:** direct credential theft was not demonstrated; the caller still needs a valid credential (and valid CSRF for refresh). The security risk is parser/intermediary differential interpretation at the authentication boundary.

**Regression requirement:** fail closed on conflicting duplicate credential representations or prove a single canonical upstream representation.

### M1-P3-03 — P3 — Duplicate auth JSON keys are accepted

**Root cause:** strict JSON handling rejects unknown fields/trailing data but the underlying JSON decoder accepts duplicate object member names.

**Observed result:** registration with two `email` keys returned 201 rather than a validation error.

**Impact:** no privilege escalation was reproduced because the attacker controls both values. This is a security-sensitive parser ambiguity that can create differential interpretation between application, validation, WAF/audit tooling or future intermediaries.

**Regression requirement:** reject duplicate object keys for auth/security-sensitive JSON.

### M1-HARD-01 — Hardening-only — Signed JWT duplicate claims / unknown crit accepted

Validly signed synthetic JWTs with conflicting duplicate `alg`, `uid` or `exp` claims and a JWT containing an unsupported `crit` header were accepted.

The current architecture provides no untrusted path to mint a valid HMAC signature over these altered header/payload bytes. Therefore this behavior is not counted as an additional attacker-exploitable P3 in this report. It remains a standards/parser-hardening gap and should receive a strict-parser regression if remediation is later authorized.

## G. Historical Findings Revalidation

```text
AUTH-ADV-02=CONFIRMED_P3_DEFERRED
OI-NEW-02=PASS_REVALIDATED
OI-NEW-07=PASS_REVALIDATED
```

OI-NEW-07 had an initial false alarm in the new audit harness. The first synthetic race test changed principal state without reproducing the real `AuthShell.handleAuth()` generation increment. Static review of `AuthShell` showed that real login/register increments generation. After correcting the oracle to model that controller transition, both late refresh and late logout tests passed, and a negative control proved the stale-operation oracle detects an omitted generation transition.

## H. Harness Defects

The following harness defects were found and corrected; none is classified as a product vulnerability:

1. **HD-01 — Frontend generation omission.** Synthetic principal transition did not increment operation generation as real AuthShell does.
2. **HD-02 — Registration timing limiter contamination.** Repeated timing requests saturated auth admission and produced 429; timing fixtures were isolated.
3. **HD-03 — Invalid PostgreSQL password-hash fixture.** Direct session fixture violated the real Argon2id database constraint; replaced with a valid Argon2id-compatible synthetic hash.
4. **HD-04 — Overbroad test selector.** `^TestModule1` also matched `TestModule10...`; selector corrected.
5. **HD-05 — Workflow string-replacement corruption.** A JavaScript replacement token corrupted YAML during an audit workflow edit; workflow was fully rewritten from a clean specification.
6. **HD-06 — State-machine compile type mismatch.** `testing.TB` was passed to a helper accepting `*testing.T`; corrected before final evidence.
7. **HD-07 — Equal-duplicate oracle too strict.** Two identical credentials were initially treated as conflicting ambiguity. Final oracle allows equal duplicates and rejects only conflicting duplicates.

## I. Coverage Gaps

The following remain unverified or only partially verified:

- Real CDN/reverse-proxy normalization of duplicate Authorization/Cookie fields.
- Production cookie-domain topology and browser cookie-tossing behavior.
- Third-party/production log aggregation.
- Real production email verification/delivery infrastructure does not exist in the tested scope; no email was sent.
- Registration timing was measured but deliberately not classified as a timing side channel from the bounded sample.
- No external stolen credentials or production sessions were used.
- The separate five-hour campaign was not started as explicitly required.
- No final repository-wide assault was started.
- JWT duplicate-claim/crit behavior is proven at the verifier boundary, but attacker exploitability is not established because an untrusted signer is absent.

## J. Independent Gatekeeper Handoff

The final Gatekeeper block must be refreshed after the report-only commit and its exact-head specialized workflow run. Current pre-report evidence:

```text
REPOSITORY=AsifAbbasov/OpenInvest

DEVELOP_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3
DEVELOP_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba
DEVELOP_PROTECTED=YES

AUDIT_BRANCH=audit/module1-auth-session-privacy-adversarial-v2
AUDIT_HEAD_SHA=POST_REPORT_PENDING
AUDIT_HEAD_TREE=POST_REPORT_PENDING

PR_249_STATE=OPEN_DRAFT
PR_249_MERGED=NO

PRODUCTION_CODE_MODIFIED=NO
PROTECTED_DEVELOP_MODIFIED=NO

AUTH_BYPASS=NOT_REPRODUCED_WITHOUT_VALID_CREDENTIALS
SESSION_FIXATION=PASS
SESSION_HIJACKING_RESISTANCE=PASS_TESTED_SCOPE
REFRESH_ROTATION=PASS
REFRESH_REPLAY=PASS
LOGOUT_REVOCATION=PASS
CSRF=PASS_TESTED_SCOPE
COOKIE_POLICY=PASS_EXISTING_POLICY_TESTS
SESSION_ISOLATION=PASS
AUTH_CONCURRENCY=PASS
POSTGRES_SESSION_RACES=PASS
RAW_HTTP_AMBIGUITY=FAIL_P3
JWT_SIGNATURE_INTEGRITY=PASS
JWT_TAMPERING=PASS
JWT_CANONICALIZATION=FAIL_P3
REGISTRATION_ENUMERATION=CONFIRMED_P3
EMAIL_OWNERSHIP_VERIFICATION=ABSENT
PRE_REGISTRATION_SQUATTING=CONFIRMED
PRIVACY_CACHE=PASS_TESTED_SCOPE
AUTH_LOG_LEAKAGE=PASS_TESTED_PATHS
FRONTEND_REFRESH_FAILURE_HANDLING=PASS
AUTH_DATABASE_ERROR_CLASSIFICATION=PASS

FUZZ_TARGETS=4
FUZZ_EXECUTIONS=5061153
FUZZ_DURATION=120s
CONCURRENCY_TESTS=HIGH_CONTENTION_5_LEVELS_PLUS_3_POSTGRES_RACES_PLUS_FRONTEND_STATE_RACES
FAILED_TESTS=CONFIRMED_SECURITY_ASSERTIONS_ONLY_AFTER_HARNESS_REPAIR
SKIPPED_TESTS=0_IN_FINAL_SPECIALIZED_RUN
HARNESS_DEFECTS=7_CORRECTED

NORMAL_CI_RESULT=BASELINE_SUCCESS_10_OF_10_RUN_38181761847
SPECIALIZED_CI_RESULT=PRE_REPORT_COMPLETED_WITH_CONFIRMED_FINDINGS
SPECIALIZED_CI_RUN_ID=38322758278

CONFIRMED_P0=0
CONFIRMED_P1=0
CONFIRMED_P2=0
CONFIRMED_P3=4

NEW_FINDINGS=3_P3_PLUS_1_HARDENING_ONLY
UNVERIFIED_PROPERTIES=SEE_COVERAGE_GAPS
REMEDIATION_REQUIRED=YES_FOR_CONFIRMED_P3_FINDINGS

FIVE_HOUR_CAMPAIGN_STARTED=NO
MODULE_1_CLOSURE_READY=NO
MODULE_1_CLOSED=NO

MERGE_AUTHORIZED=NO
REMEDIATION_AUTHORIZED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO

STOP_FOR_INDEPENDENT_GATEKEEPER_REVIEW=YES
```
