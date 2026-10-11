# Module 12 V2 CORS cardinality remediation V1

```text
FINDING_ID=M12-V2-CORS-01
ALIAS=M12V2-P3-01
SEVERITY=P3
CLASSIFICATION=PRODUCT_SECURITY_FINDING
ORIGINAL_AUDIT_PR=246
ORIGINAL_AUDIT_HEAD=086407d98fd3edd6a931b72eff763cc8dda478a3
ORIGINAL_SPECIALIZED_RUN=38159497398
BASE_SHA=eba69ae32dbe810fb3436fe2f6691777d7cbfe7d
BASE_TREE=c4e8addd305489b4faf09420dd87e86207767e91
BASE_BRANCH=develop
BASE_PROTECTED=YES
REMEDIATION_BRANCH=fix/module12-cors-duplicate-origin-hardening
```

## Root cause and remediation

ROOT_CAUSE=The CORS middleware authorized a lossy single-value `c.Get("Origin")` lookup without checking physical field cardinality. The independently accepted historical audit evidence supplies the pre-fix result; this remediation did not reproduce it.

REMEDIATION=Inspect every request Origin value before authorization. Reject multiple fields or a comma-combined value with HTTP 400 before the application handler. Only zero or one unambiguous field reaches existing exact allow-list and OPTIONS handling. The generic rejection does not reflect Origin and emits no CORS permission headers.

```text
ORIGIN_CARDINALITY_POLICY=EXACTLY_ZERO_OR_ONE_PHYSICAL_ORIGIN
MULTIPLE_PHYSICAL_ORIGIN_FIELDS=FAIL_CLOSED
CONFLICTING_DUPLICATE_ORIGIN=FAIL_CLOSED
EQUAL_DUPLICATE_ORIGIN=FAIL_CLOSED
COMMA_COMBINED_MULTI_ORIGIN=FAIL_CLOSED
AMBIGUOUS_ORIGIN_HTTP_STATUS=400
AMBIGUOUS_ORIGIN_ACAO=ABSENT
AMBIGUOUS_ORIGIN_ACAC=ABSENT
AMBIGUOUS_ORIGIN_CREDENTIAL_CORS_PERMISSION=NO
AMBIGUOUS_ORIGIN_HANDLER_EXECUTION=NO
```

API verification used the repository-pinned Fiber v3.5.0 and fasthttp v1.73.0 sources. Fiber `DefaultCtx.Request()` exposes `*fasthttp.Request`; `RequestHeader.PeekAll("Origin")` enumerates all stored matching field values using normalized header names. Origin follows the ordinary-header all-values path. Returned borrowed values are consumed synchronously before any subsequent header lookup. The real listener regressions additionally observed two and three separate fields, including mixed-case field names. Separate one-field regressions cover an intermediary-combined comma representation.

Source references: [Fiber v3.5.0 ctx.go](https://github.com/gofiber/fiber/blob/v3.5.0/ctx.go), [fasthttp v1.73.0 header.go](https://github.com/valyala/fasthttp/blob/v1.73.0/header.go).

## Scope

```text
FILES_CHANGED=backend-go/internal/httpapi/cors.go;backend-go/internal/httpapi/cors_origin_test.go;docs/audit/MODULE12_V2_CORS_01_REMEDIATION_V1.md
PRODUCTION_FILES_CHANGED=backend-go/internal/httpapi/cors.go
TEST_FILES_CHANGED=backend-go/internal/httpapi/cors_origin_test.go
```

No provider, Redis budget, authentication/session, PostgreSQL, frontend or workflow implementation changes are included. The new branch starts from protected develop, not from either historical audit PR. PR #246 and PR #244 are preserved.

## Permanent ordinary-suite coverage

`TestCORSOriginCardinality` sends synthetic requests through a real ephemeral loopback TCP listener into Fiber and the production CORS middleware. It observes raw field cardinality at the application boundary, status, permission headers, body non-reflection and handler invocation. Each case runs both POST and OPTIONS. A counter verifies that no ambiguous mutation request reaches the downstream handler. These permanent tests do not require the audit branch.

```text
DUPLICATE_EQUAL_ORIGIN=PASS_FAIL_CLOSED
DUPLICATE_CONFLICTING_ALLOWED_FIRST=PASS_FAIL_CLOSED
DUPLICATE_CONFLICTING_DISALLOWED_FIRST=PASS_FAIL_CLOSED
THREE_ORIGIN_FIELDS=PASS_FAIL_CLOSED
COMMA_COMBINED_ORIGIN=PASS_FAIL_CLOSED_EQUAL_AND_CONFLICTING
SINGLE_ALLOWED_ORIGIN=PASS_EXISTING_CREDENTIAL_CORS_PERMISSION
SINGLE_DISALLOWED_ORIGIN=PASS_NO_CORS_PERMISSION
ORIGIN_NULL=PASS_NO_CORS_PERMISSION
MISSING_ORIGIN=PASS_NO_CORS_PERMISSION
ALLOWED_PREFLIGHT=PASS_204_WITH_EXISTING_PERMISSIONS
DISALLOWED_PREFLIGHT=PASS_204_WITHOUT_PERMISSIONS
AMBIGUOUS_PREFLIGHT=PASS_400_WITHOUT_PERMISSIONS
```

`TestCORSOriginCompatibility` exercises actual production authentication registration with the existing synthetic auth store and secure-cookie configuration. The five variants cover ordinary Host, an alternative Host, X-Forwarded-Host, and Forwarded host with HTTP or HTTPS proto. All retain status 201, credentialed single-Origin CORS, Secure/HttpOnly/SameSite=Strict host-only refresh cookies, no-store, and no unexpected Location. These are compatibility observations, not assertions about deployment proxy trust. Existing auth lifecycle/cache-policy regressions remain unchanged and are included in normal CI.

## Direct verification

Executed from backend-go with Go 1.26.9:

```sh
go test -v ./internal/httpapi -run '^TestCORSOrigin(Cardinality|Compatibility)$' -count=1
go test -race -v ./internal/httpapi -run '^TestCORSOrigin(Cardinality|Compatibility)$' -count=1
```

```text
FOCUSED_TEST_RESULT=PASS
FOCUSED_RACE_RESULT=PASS
CARDINALITY_CASES_PER_RUN=20
COMPATIBILITY_CASES_PER_RUN=5
DATA_RACE=NONE_OBSERVED
```

The normal unchanged CI runs `go test ./...` and `go test -race ./...`, which include these tests. Exact-head CI run identity and all ten job conclusions must be verified from GitHub after opening the remediation PR and returned in the final stop report; local focused results do not substitute for protected CI.

### FOCUSED_TEST exact output markers

```text
M12_CORS_ORIGIN_CASE=single_allowed_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="https://allowed.example.test" ACAC="true" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_allowed_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="https://allowed.example.test" ACAC="true" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_disallowed_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_disallowed_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=origin_null METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=origin_null METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=missing_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=0 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=missing_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=0 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_equal_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_equal_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_allowed_first METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_allowed_first METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_disallowed_first METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_disallowed_first METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=three_origin_fields METHOD=POST PHYSICAL_ORIGIN_FIELDS=3 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=three_origin_fields METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=3 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_allowed_and_disallowed METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_allowed_and_disallowed METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_equal_origins METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_equal_origins METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=ordinary AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=host AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_host AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_http AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_https AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
PASS
```

### FOCUSED_RACE exact output markers

```text
M12_CORS_ORIGIN_CASE=single_allowed_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="https://allowed.example.test" ACAC="true" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_allowed_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="https://allowed.example.test" ACAC="true" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_disallowed_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=single_disallowed_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=origin_null METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=origin_null METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=missing_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=0 STATUS=200 ACAO="" ACAC="" HANDLER_EXECUTED=1 RESULT=PASS
M12_CORS_ORIGIN_CASE=missing_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=0 STATUS=204 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_equal_origin METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_equal_origin METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_allowed_first METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_allowed_first METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_disallowed_first METHOD=POST PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=duplicate_conflicting_origin_disallowed_first METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=2 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=three_origin_fields METHOD=POST PHYSICAL_ORIGIN_FIELDS=3 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=three_origin_fields METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=3 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_allowed_and_disallowed METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_allowed_and_disallowed METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_equal_origins METHOD=POST PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_ORIGIN_CASE=comma_combined_equal_origins METHOD=OPTIONS PHYSICAL_ORIGIN_FIELDS=1 STATUS=400 ACAO="" ACAC="" HANDLER_EXECUTED=0 RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=ordinary AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=host AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_host AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_http AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
M12_CORS_COMPATIBILITY_CASE=forwarded_https AUTH_REGISTRATION_STATUS=201 COOKIE_SECURE=true COOKIE_HTTPONLY=true COOKIE_SAMESITE=3 CACHE_CONTROL="no-store" RESULT=PASS
PASS
```

## Verification boundaries and stop gate

This proves repository-side rejection at the tested Fiber/fasthttp application boundary, including both physically repeated fields and a single combined comma value. It does not prove production edge/CDN normalization. No external targets or production credentials were used. Independent review, authorized merge, exact post-merge verification and audit PR #246 revalidation remain outstanding.

```text
PRODUCTION_EDGE_CDN_NORMALIZATION=NOT_VERIFIED
APPROVE_REMEDIATION_FOR_MERGE=NO
FORMAL_MODULE_12_CLOSED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO
```
