# Module 1 Authentication / Session / Privacy — Strengthened Adversarial Evidence Report

## Executive verdict

This report records the authorized bounded continuation of the OpenInvest Module #1 authentication/session/privacy audit.

The work was performed only on the isolated audit branch `audit/module1-auth-session-privacy-strengthened-v2`. Protected `develop` was not modified, PR #249 was not merged, and no product remediation was performed.

The strengthened campaign revalidated the historical P3 `AUTH-ADV-02` (registration membership enumeration / missing email ownership verification / pre-registration namespace squatting). No new P0, P1, or P2 product security finding was demonstrated by the executed bounded tests.

Two major initial failures were confirmed as invalid test-oracle assumptions rather than product vulnerabilities:
1. noncanonical but byte-equivalent JWT Base64URL signature spelling;
2. mandatory duplicate-JSON rejection where the current contract intentionally does not define duplicate-member rejection.

The strengthened matrix also added cross-principal credential-collision evidence, machine-derived protected-route coverage, bounded high-contention PostgreSQL session testing, idempotency security-context testing, auth parser fuzzing, JWT algorithm/required-claim tests, malformed cookie/content-type tests, bounded resource snapshots, and synthetic auth-secret log sentinels.

## Baseline integrity

```text
REPOSITORY=AsifAbbasov/OpenInvest
DEVELOP_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3
DEVELOP_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba
DEVELOP_PROTECTED=YES
REQUIRED_CONTEXTS=10

INITIAL_AUDIT_SHA=48b088d4aab4887a4e8428c28085940b1142e604
INITIAL_AUDIT_TREE=63ea86c1c24d0e2b2fc466ae76ad30ac10ceb01b

PR249_STATE=OPEN_DRAFT
PR249_MERGED=NO

AUDIT_BRANCH=audit/module1-auth-session-privacy-strengthened-v2
PRE_REPORT_EVIDENCE_SHA=3bcd1a71842e797de8a2becb86c683e6f04bbf03
PRE_REPORT_EVIDENCE_TREE=631584f5ffa216fe63cc0799841c13f68ccd0fb0
```

The final report commit SHA/tree and its exact-head workflow IDs are intentionally emitted in the external Gatekeeper STOP report after this file is committed and the resulting workflows finish; embedding a commit's own SHA or its future workflow IDs inside itself would require another commit and invalidate exact-head provenance.

## Threat model and security boundaries examined

The executed tests challenge:
- bearer access-token parsing and principal resolution;
- JWT structural integrity, HMAC integrity, required claims, algorithm/type validation, expiry behavior and textual representation;
- refresh-cookie parsing and physical-header collisions;
- refresh rotation, replay, logout and family-state persistence;
- CSRF/session binding;
- session fixation;
- cross-principal authorization and portfolio isolation;
- idempotency principal/payload/operation/resource scope;
- PostgreSQL concurrent session transitions;
- registration identity normalization, enumeration and namespace reservation;
- frontend stale-response and transient-error state transitions;
- auth error sanitization, privacy cache behavior and repository-owned log/error leakage;
- anonymous access to machine-derived protected runtime routes.

No production accounts, real user credentials, third-party systems, production infrastructure, or five-hour campaign were used.

## Initial failure analysis

### HF-01 — JWT canonical textual encoding

```text
ORIGINAL_JOB=jwt-structural-fuzz
ORIGINAL_ASSERTION=noncanonical byte-equivalent signature spelling must be rejected
ROOT_CAUSE=test oracle conflated textual spelling with decoded signature bytes
CLASSIFICATION=HARNESS_ORACLE_DEFECT
PRODUCT_FINDING=NO
CORRECT_ORACLE=byte-changing mutation must fail; byte-equivalent textual alias is governed only by an explicit canonical-text policy
CURRENT_POLICY=JWT_CANONICALIZATION_POLICY_NOT_SPECIFIED
```

The corrected test independently decodes the canonical and alias signature strings and confirms byte equality before classification. Acceptance is allowed only when the resolved principal remains unchanged.

### HF-02 — duplicate JSON object members

```text
ORIGINAL_JOB=http-auth-ambiguity
ORIGINAL_CASES=duplicate_email;duplicate_password
ORIGINAL_EXPECTATION=HTTP_400_AND_ZERO_PERSISTENCE_CALLS
OBSERVED_RUNTIME=LAST_MEMBER_WINS
ROOT_CAUSE=test asserted an undocumented duplicate-member rejection policy
CLASSIFICATION=HARNESS_ORACLE_DEFECT
PRODUCT_FINDING=NO
REPOSITORY_AUTHORITY=M10_P3_01_HARDENING_ONLY_DEFERRED_NON_BLOCKING
```

Strict rejection remains required and verified for trailing documents, wrong top-level type and unknown fields.

## Security property matrix

| Property | Strengthened oracle / attack | Result |
|---|---|---|
| Authentication bypass | malformed/missing bearer, parser fuzz, anonymous protected-route inventory | PASS for executed matrix |
| Protected-route guard coverage | machine-derived runtime route inventory | 15/15 tested; 0 without auth guard |
| Authorization ambiguity | duplicate/equal/conflicting/multi-value Authorization with principal oracle | first-value semantics observed; no wrong-principal resolution in tested cases |
| Cross-principal ambiguous Authorization | valid A + valid B in both orders | PASS; first principal preserved |
| Session fixation | attacker-controlled pre-auth refresh-like cookie | PASS for tested vector |
| Refresh cookie ambiguity | same-principal and cross-principal physical duplicate Cookie fields | first-value semantics observed; no tested principal confusion |
| CSRF binding | refresh A + CSRF B | PASS |
| Refresh rotation/replay | existing replay regressions + state-machine fuzz + PostgreSQL races | PASS for executed paths |
| Logout/session family | refresh/logout race and replay/descendant race | PASS |
| PostgreSQL session invariants | real PostgreSQL 18 integration | PASS |
| High contention | 2/4/8/16/32 contenders, 25 repetitions/level | PASS; no excess valid lineage |
| Cross-principal resource isolation | A/B portfolio access | PASS |
| Idempotency principal scope | A/B same key | PASS |
| Idempotency payload/operation/resource scope | same key across changed security context | PASS |
| JWT signature integrity | byte-changing fuzz + fixed tampering | PASS |
| JWT parser robustness | structural fuzz | PASS |
| JWT alg/type/required claims | alg none/HS512/empty/missing; typ; uid; investmentSubjectId; exp | PASS |
| JWT textual canonicalization | byte-equivalent alias | policy not specified; principal unchanged |
| JWT future iat | valid signed future iat | accepted; hardening observation, no attacker signing path demonstrated |
| Duplicate signed claim | duplicate investmentSubjectId | last member accepted; observation only, signing secret required |
| Registration enumeration | new vs duplicate synthetic address | CONFIRMED historical AUTH-ADV-02 |
| Email ownership verification | active registration without ownership proof | ABSENT / historical AUTH-ADV-02 |
| Pre-registration squatting | namespace reserved before ownership proof | CONFIRMED historical AUTH-ADV-02 |
| Registration timing | bounded 6 samples/class | evidence collected; not classified from small sample |
| Email normalization | fixed matrix + fuzz | case/trim stable; plus/dot distinct by current policy |
| Frontend stale responses | logout/refresh and principal transition matrix | PASS |
| Auth DB error classification | existing sanitized infrastructure failure regression | PASS for tested paths |
| Privacy cache | existing auth lifecycle cache tests | PASS for tested paths |
| Auth log leakage | synthetic raw/encoded sentinels in repository-owned log/error paths | PASS for tested paths; external aggregator NOT_VERIFIED |
| Resource recovery | goroutine/FD/heap snapshots + post-attack legitimate request | service recovery PASS; FD delta retained as observation |

## New adversarial execution

### Fixed and race matrices

```text
TEST_ID=M1-JWT-FIXED
COMMAND=go test -race -v ./internal/auth -run '^TestModule1JWT' -count=1
RESULT=PASS

TEST_ID=M1-HTTP-AUTH
COMMAND=go test -race -v ./internal/httpapi -run '^TestModule1' -count=1 -timeout=5m
RESULT=PASS

TEST_ID=M1-PG-RACES
ENVIRONMENT=PostgreSQL_18_service_container
CASES=logout_refresh_race;replay_descendant_race;session_family_guard;registration_enumeration;cross_principal;idempotency_scope;bounded_high_contention
RESULT=PASS
COLLECT_ALL_FAILURES=0

TEST_ID=M1-FRONTEND
COMMAND=node --experimental-strip-types --test tests/auth-session.test.mjs tests/module1-auth-adversarial.test.mjs
CASES=20
RESULT=PASS_20/20
```

### Fuzz campaigns

All campaigns were bounded with hard outer timeouts and used four workers in GitHub Actions.

| Target | Seeds | Duration | Executions | Result |
|---|---:|---:|---:|---|
| JWT byte-changing mutation | 3 | 30s | 3,341,557 | PASS |
| JWT structural parser | 5 | 30s | 4,159,677 | PASS |
| Session lifecycle state machine | 3 | 30s | 1,488,950 | PASS |
| Email normalization properties | 5 | 30s | 351,102 | PASS |
| Bearer parser properties | 7 | 30s | 493,719 | PASS |
| Strict auth JSON decoder | 7 | 30s | 275,629 | PASS |

```text
FUZZ_TARGETS=6
TOTAL_FUZZ_EXECUTIONS=10110634
BOUNDED_FUZZ_TIME=180s_requested
FUZZ_FAILURES=0
MINIMIZED_FAILURES=0
```

### Cross-principal ambiguous credentials

```text
AUTH_A_THEN_B=HTTP_200_FIRST_PRINCIPAL_PRESERVED
AUTH_B_THEN_A=HTTP_200_FIRST_PRINCIPAL_PRESERVED
COOKIE_C_THEN_D=HTTP_200_FIRST_PRINCIPAL_PRESERVED
COOKIE_F_THEN_E=HTTP_200_FIRST_PRINCIPAL_PRESERVED

CROSS_PRINCIPAL_AMBIGUOUS_AUTHORIZATION_CONFUSION=NO_FOR_TESTED_FIRST_VALUE_SEMANTICS
CROSS_PRINCIPAL_AMBIGUOUS_REFRESH_COOKIE_CONFUSION=NO_FOR_TESTED_FIRST_VALUE_SEMANTICS
```

This is repository/runtime evidence only. A deployment intermediary with different duplicate-header normalization remains outside this isolated audit.

### Protected route inventory

```text
PROTECTED_ROUTES_DISCOVERED=15
PROTECTED_ROUTES_TESTED=15
PROTECTED_ROUTES_WITHOUT_AUTH_GUARD=0
```

The machine-derived runtime inventory covered GET/POST portfolio routes, valuation mutations, portfolio projections, transaction read/write/correction/reversal, and import review/append routes.

### Idempotency security context

```text
IDEMPOTENCY_EXACT_RETRY_STABLE=YES
IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_PAYLOAD=NO
IDEMPOTENCY_RESULT_CROSSES_PRINCIPAL=NO
IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_OPERATION=NO
IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_RESOURCE=NO
```

### Bounded high-contention campaign

```text
CONTENTION_LEVELS=2,4,8,16,32
REPETITIONS_PER_LEVEL=25
TOTAL_ATTEMPTS=1550
TOTAL_SUCCESSES=125
ACTUAL_DURATION=40.09s

GOROUTINES_BASELINE=4
GOROUTINES_PEAK_OBSERVED=38
GOROUTINES_AFTER=5

FD_BASELINE=9
FD_AFTER=15

HEAP_ALLOC_BASELINE=1440672
HEAP_ALLOC_AFTER_GC=1149640

RACE_CREATES_MORE_VALID_SESSION_LINEAGES_THAN_SERIAL_POLICY_ALLOWS=NO
POST_ATTACK_SERVICE_RECOVERY=PASS
```

The FD increase is recorded, not hidden. This snapshot alone is insufficient to classify a descriptor leak because the process retains expected PostgreSQL/runtime descriptors and no per-level monotonic leak oracle was established.

### Registration / AUTH-ADV-02

```text
AUTH_ADV_02_DIRECT_RESPONSE_ENUMERATION=CONFIRMED
EMAIL_OWNERSHIP_VERIFICATION=ABSENT_BY_RUNTIME_REGISTRATION_FLOW
PRE_REGISTRATION_SQUATTING=CONFIRMED
ACCOUNT_STATE_AFTER_REGISTRATION=active
NORMALIZED_IDENTITY_ROWS=1

AUTH_ADV_02_TIMING_SAMPLES_PER_CLASS=6
NEW_MEDIAN≈267.8ms
DUPLICATE_MEDIAN≈263.2ms
AUTH_ADV_02_TIMING_ENUMERATION=EVIDENCE_COLLECTED_NOT_CLASSIFIED_FROM_SMALL_SAMPLE
```

### Auth log/error secret sentinels

Synthetic sentinels were used for access/refresh/CSRF/password-like and infrastructure-secret-like values.

```text
AUTH_LOG_RAW_SECRET_LEAK=NO_FOR_TESTED_REPOSITORY_OWNED_PATHS
AUTH_ERROR_RESPONSE_RAW_SECRET_LEAK=NO_FOR_TESTED_PATHS
EXTERNAL_LOG_AGGREGATOR=NOT_VERIFIED
```

## Historical findings revalidation

### AUTH-ADV-02

```text
STATUS=CONFIRMED
SEVERITY=P3
DIRECT_RESPONSE_ENUMERATION=CONFIRMED
EMAIL_OWNERSHIP_VERIFICATION=ABSENT
PRE_REGISTRATION_SQUATTING=CONFIRMED
TIMING_ENUMERATION=NOT_CONFIRMED_FROM_BOUNDED_SMALL_SAMPLE
```

No product remediation is authorized in this phase.

### OI-NEW-02

```text
STATUS=PASS_FOR_EXECUTED_AUTH_DATABASE_ERROR_MAPPING
LIMITATION=not exhaustive fault injection for every possible PostgreSQL error class
```

### OI-NEW-07

```text
STATUS=PASS_FOR_EXECUTED_FRONTEND_STATE_MATRIX
STALE_REFRESH_AFTER_LOGOUT_REAUTHENTICATES=NO
OLD_PRINCIPAL_RESPONSE_OVERWRITES_NEW_PRINCIPAL=NO
TRANSIENT_REFRESH_FAILURE_FORCES_LOGOUT=NO
```

## Confirmed product findings

### AUTH-ADV-02 — P3

**Root cause:** registration creates an active account without email ownership proof and exposes distinguishable duplicate-vs-new registration behavior.

**Attack preconditions:** attacker can submit a synthetic registration request for an address whose ownership has not been demonstrated.

**Observed consequence:** address namespace can be reserved before ownership proof; duplicate/new membership is distinguishable through the registration result.

**Reproduction:** isolated PostgreSQL registration integration test `TestModule1RegistrationEnumerationAndPreRegistrationSquattingEvidence`.

**Recommended remediation (not implemented here):** introduce ownership verification / pending-registration policy and remove or normalize externally observable membership distinctions according to product requirements.

**Regression requirement:** retain direct-response, namespace-squatting and bounded timing tests after any separately authorized remediation.

## Harness defects

```text
HD-01=JWT_BYTE_EQUIVALENT_TEXT_ALIAS_WAS_INCORRECTLY_TREATED_AS_BYTE_TAMPERING
HD-02=DUPLICATE_JSON_REJECTION_WAS_ASSERTED_WITHOUT_CONTRACT_AUTHORITY
```

Several intermediate workflow-only commits repaired audit workflow command/syntax defects. They changed no product runtime behavior and final evidence was rerun after those corrections.

## Coverage gaps / NOT_VERIFIED

The following are intentionally not promoted to PASS:

1. **Production intermediary duplicate-header normalization.** Tests prove repository/Fiber first-value behavior in the isolated environment, not arbitrary CDN/LB/proxy normalization.
2. **External log aggregation.** Repository-owned log/error paths were scanned; external production collectors were not accessed.
3. **Full production resource telemetry.** The bounded CI campaign recorded goroutine/FD/heap snapshots, but did not inspect production host metrics or Redis because this Module #1 campaign used isolated PostgreSQL and repository-local processes.
4. **Timing side-channel classification.** Six samples/class are deliberately insufficient for a timing-vulnerability conclusion.
5. **JWT future-iat policy.** Current verifier does not enforce future `iat`; no attacker-controlled signing path was demonstrated, so this remains a hardening/policy observation rather than a confirmed auth bypass.
6. **Duplicate signed JWT claims.** Validly signed duplicate claims follow JSON last-member behavior; no attacker signing capability was demonstrated.
7. **Five-hour campaign.** Explicitly prohibited and not started.
8. **Final repository-wide assault.** Explicitly prohibited and not started.

No feasible authorized test from the updated Gatekeeper matrix was silently skipped. Properties that require deployment or external-system evidence remain NOT_VERIFIED rather than PASS.

## Audit-only change history

The strengthened branch advanced from the authorized initial audit SHA through the following audit-only commits:

```text
e01705b37ac2dd8180649bbe9de60f9a451fdc0b  test(audit): correct auth ambiguity security oracles
5f76a582f43774537ab0dd827878ed0581173565  test(audit): validate JWT alias oracle by decoded bytes
b0773eca573090d7f65ce775096a79fb146a2413  ci(audit): run strengthened Module 1 branch verification
8d4f899ce5a096dc563fa9ff63697bd851f4af6c  test(audit): expand raw auth and cookie ambiguity matrix
6e54b95853ced006112bb48bfbc2eebba834bc06  test(audit): expand JWT structural boundary evidence
06f832a0ded8fa1269fa90391b6efca8fce5804d  test(audit): add registration canonicalization policy matrix
39977bb0d267299c3573a991bb4b230f5a861591  test(audit): account for HTTP leading OWS normalization
50bb159a2aa8aee0ef3838439302dbef7abf1737  test(audit): add auth normalization fuzz properties
67bcf8b55d971eb030713302f1f423cc23e2db2a  test(audit): add auth parser fuzz properties
c2355a027b04e4f6e80a4bfa5ff00705b55b7ad8  test(audit): add cross-principal credential collision evidence
d7aa0af889b84b92a65e46c4a381ac508efead95  ci(audit): run strengthened auth parser fuzz campaigns
695b9bdb75096685051435e1e0473815678f904e  test(audit): add bounded session contention campaign
73cd70102554278c5e88a911f5ab8613b7952ba1  ci(audit): execute bounded session contention campaign
a074ce9a88c8cd9cd69ac1b13bc0f13e4da9f5bb  ci(audit): fix bounded contention workflow command
badf9dce1e2642f074984af87c48b83e11a0360b  ci(audit): repair bounded contention job syntax
75a10d59e3fdeb23c40d263e049b994f68fb51b7  ci(audit): restore verified workflow and add contention case
29e91465112f7a0a8232a90f9c8c1259847575b7  test(audit): cover JWT algorithm and required claims
618c0df5bcd9435501788aa6bacc377db4c9f5ed  test(audit): cover malformed cookies and content types
d5569a9b2e4ff6a9cc92dd9aedbb53d3b6e      test(audit): derive protected route auth inventory
263b4bca953412c2ba88376299d97b3cc27dda99  test(audit): expand idempotency security context matrix
b783db86e811cbbfaf65dbc5dd44e43aa9c0c015  ci(audit): execute idempotency scope matrix
a4b9917c55e571e07db966eab5981385603001d5  test(audit): record bounded resource recovery snapshots
0ba475725e4251790de41dc21995f4ced5c79985  ci(audit): repair workflow and run idempotency scope matrix
114f1b33890ac73ce8dd8bb1873dec2409ddf0a9  test(audit): add auth secret log sentinels
3bcd1a71842e797de8a2becb86c683e6f04bbf03  test(audit): fix single-read auth sentinel fixture
```

Changed paths are restricted to Module #1 audit tests and the specialized workflow. No runtime/product implementation file was modified.

## Pre-report workflow evidence

```text
PRE_REPORT_NORMAL_CI_RUN_ID=38374607396
PRE_REPORT_NORMAL_CI=SUCCESS_10/10

PRE_REPORT_SPECIALIZED_RUN_ID=38374607607
PRE_REPORT_SPECIALIZED=SUCCESS_6/6

SPECIALIZED_JOBS=
frontend-auth-state
auth-parser-fuzz
existing-auth-regressions
jwt-structural-fuzz
postgres-session-races
http-auth-ambiguity
```

A fresh exact-head rerun is required and must be referenced externally after this report commit.

## Independent Gatekeeper handoff

```text
REPOSITORY=AsifAbbasov/OpenInvest

DEVELOP_SHA=8d9998590fba6c605a64ddcc8429f51df82b56c3
DEVELOP_TREE=bed3b0b3466de160c6bb954b41b373e5ed7ea1ba
DEVELOP_PROTECTED=YES

INITIAL_AUDIT_SHA=48b088d4aab4887a4e8428c28085940b1142e604
INITIAL_AUDIT_TREE=63ea86c1c24d0e2b2fc466ae76ad30ac10ceb01b

AUDIT_BRANCH=audit/module1-auth-session-privacy-strengthened-v2
PR249_STATE=OPEN_DRAFT
PR249_MERGED=NO

PRODUCTION_CODE_MODIFIED=NO
PROTECTED_DEVELOP_MODIFIED=NO

AUTH_BYPASS=PASS_FOR_EXECUTED_BOUNDED_MATRIX
SESSION_FIXATION=PASS_FOR_TESTED_VECTOR
SESSION_HIJACKING_RESISTANCE=PASS_FOR_EXECUTED_SYNTHETIC_BOUNDARIES
REFRESH_ROTATION=PASS
REFRESH_REPLAY=PASS_FOR_EXECUTED_MATRIX
LOGOUT_REVOCATION=PASS_FOR_EXECUTED_RACES
CSRF=PASS_FOR_TESTED_CROSS_PRINCIPAL_BINDING
COOKIE_POLICY=PASS_FOR_TESTED_AUTH_LIFECYCLE
SESSION_ISOLATION=PASS_FOR_EXECUTED_BOUNDARIES
AUTH_CONCURRENCY=PASS_FOR_EXECUTED_DETERMINISTIC_AND_CONTENTION_TESTS
POSTGRES_SESSION_RACES=PASS
RAW_HTTP_AMBIGUITY=PARTIALLY_VERIFIED_FIRST_VALUE_SEMANTICS_DEPLOYMENT_INTERMEDIARY_NOT_VERIFIED
JWT_SIGNATURE_INTEGRITY=PASS
JWT_TAMPERING=PASS
JWT_CANONICALIZATION=NO_REQUIRED_TEXTUAL_POLICY_PRINCIPAL_UNCHANGED
REGISTRATION_ENUMERATION=CONFIRMED_P3_AUTH_ADV_02
EMAIL_OWNERSHIP_VERIFICATION=ABSENT_CONFIRMED_P3_AUTH_ADV_02
PRE_REGISTRATION_SQUATTING=CONFIRMED_P3_AUTH_ADV_02
PRIVACY_CACHE=PASS_FOR_TESTED_PATHS
AUTH_LOG_LEAKAGE=NO_RAW_SECRET_LEAK_FOR_TESTED_REPOSITORY_PATHS
FRONTEND_REFRESH_FAILURE_HANDLING=PASS
AUTH_DATABASE_ERROR_CLASSIFICATION=PASS_FOR_TESTED_PATHS

FUZZ_TARGETS=6
FUZZ_EXECUTIONS=10110634
FUZZ_DURATION=180s_requested
CONCURRENCY_TESTS=POSTGRES_RACES_PLUS_1550_ATTEMPT_CONTENTION_CAMPAIGN
FAILED_TESTS=0_ON_PRE_REPORT_EXACT_HEAD
SKIPPED_TESTS=0_FEASIBLE_AUTHORIZED_MATRIX;EXTERNAL_OR_DEPLOYMENT_ONLY_PROPERTIES_NOT_VERIFIED
HARNESS_DEFECTS=2_MAJOR_ORACLE_DEFECTS_PLUS_INTERMEDIATE_WORKFLOW_SYNTAX_DEFECTS

CONFIRMED_P0=0
CONFIRMED_P1=0
CONFIRMED_P2=0
CONFIRMED_P3=1_HISTORICAL_AUTH_ADV_02_REVALIDATED

NEW_FINDINGS=NONE
UNVERIFIED_PROPERTIES=PRODUCTION_INTERMEDIARY_HEADER_NORMALIZATION;EXTERNAL_LOG_AGGREGATOR;PRODUCTION_RESOURCE_TELEMETRY;TIMING_SIDE_CHANNEL_CLASSIFICATION
REMEDIATION_REQUIRED=AUTH_ADV_02_REMAINS_OPEN_FOR_SEPARATELY_AUTHORIZED_PRODUCT_REMEDIATION

FIVE_HOUR_CAMPAIGN_STARTED=NO
MODULE_1_CLOSURE_READY=NO_GATEKEEPER_DECISION_REQUIRED_AND_AUTH_ADV_02_REMAINS_OPEN
MODULE_1_CLOSED=NO

MERGE_AUTHORIZED=NO
REMEDIATION_AUTHORIZED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO

STOP_FOR_INDEPENDENT_GATEKEEPER_REVIEW=YES
```
