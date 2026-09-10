# OI-AUDIT-007 — API / Runtime Parity Remediation Candidate

| Field | Value |
|---|---|
| Scope | Post-Stage-3.77 audit remediation |
| Finding | `OI-AUDIT-007` |
| Severity | P2 |
| Canonical base | `develop@e6ccfef2984ac17075952fbe9b5d62ff834e8b69` |
| Canonical tree | `c4287cff527a02f25c3144c8aca2cf55e30a419e` |
| Path | Development path |
| ADR | None proposed |
| Published runtime head | `fffa4ae516e21cd46e555f78b2646c3a9b5fd119` |
| Draft PR | `#171` |

## Revalidation of the original finding

The audit correctly identified a mismatch between the Stage-2 frozen OpenAPI surface and the shipped
`NewReplayWithCorporateActionProvider -> newReplayApp` router, but remediation-level review refines the
root cause into two different classes.

### Current regression that must be fixed

`POST /api/v1/dividends/calculate` is not merely a future contract reservation.

Stage 3.68 explicitly implemented that frozen operation end-to-end and declared its lifecycle complete.
The Stage 3.68 implementation added `dividendLimiter` to the legacy `New` / `NewDevelopment`
constructors and registered the route in `registerRoutes`, but it did not add the same limiter or route
to the replay-safe constructor used by `backend-go/cmd/api/main.go`.

Therefore the shipped production composition currently loses an implemented feature and would also
lose the Stage 3.68 fresh-command limiter if the route were restored without constructor remediation.

### Frozen future contract operations

These five operations are different:

- `PATCH /api/v1/portfolios/{portfolioId}`
- `DELETE /api/v1/portfolios/{portfolioId}`
- `GET /api/v1/portfolios/{portfolioId}/snapshots`
- `GET /api/v1/dividends/calendar`
- `GET /api/v1/dashboard`

Stage 2 intentionally froze the complete authorized MVP contract before runtime implementation and
explicitly stated that Go handlers/services were not implemented in that stage. No current handlers
were found for these five operations.

Deleting them from OpenAPI would therefore rewrite the frozen MVP contract merely to make today's
router match. This candidate does not do that.

## Remediation decision

1. Restore `POST /api/v1/dividends/calculate` to the replay-safe production router.
2. Configure `newDividendCalculatorRateLimiter()` in both replay-safe constructors.
3. Keep the five authorized-but-unimplemented Stage-2 operations in OpenAPI.
4. Mark only those operations with:
   `x-openinvest-runtime-status: planned`.
5. Make the repository OpenAPI validator derive the expected shipped route set from the OpenAPI:
   unmarked operation = implemented/current runtime; `planned` = frozen contract reservation.
6. Hard-code the exact five currently approved Stage-2 runtime reservations as an allowlist. Marking any
   other operation `planned` is itself a validator error, so lifecycle metadata cannot be used to hide a
   new missing-route regression without a reviewed validator-code change.
7. Parse exactly one canonical `func newReplayApp(api *API) *fiber.App` declaration and inspect only
   its body for production route evidence. Missing, duplicate, unsupported-signature, non-literal-path,
   duplicate-route, missing-current, and shipped-planned cases fail closed. Dead/helper functions and
   nested function literals cannot satisfy production parity.

This preserves the API-first frozen MVP contract while making runtime readiness explicit and
machine-verifiable.

## Why no ADR

This candidate does not change financial, privacy, security, tax, snapshot, event, provider, database,
framework, protocol, cost or rollback semantics. It restores an already-approved Stage 3.68 operation
and adds executable lifecycle metadata/validation to distinguish implemented runtime from the already
frozen future contract. No architecture decision is replaced.

## Exact proposed repository files

1. `backend-go/internal/httpapi/replay_app.go`
2. `backend-go/internal/httpapi/replay_app_dividends_test.go`
3. `backend-go/cmd/validate-openapi/main.go`
4. `backend-go/cmd/validate-openapi/runtime_route_parity.go`
5. `backend-go/cmd/validate-openapi/runtime_route_parity_test.go`
6. `openapi/openapi.yaml`
7. `docs/stages/OI_AUDIT_007_API_RUNTIME_PARITY_REMEDIATION.md`

Changed-file count for the reviewed runtime candidate: **7**, inside the <=25 review budget.

## Required regression evidence

- production `NewReplay` accepts a valid dividend-calculator request and returns 200, not router 404;
- calculator uses the replay-safe service path;
- both replay constructors initialize the Stage 3.68 fresh-command limiter;
- runtime route parser normalizes Fiber `:param` syntax to OpenAPI `{param}`;
- route evidence is accepted only from the exact `newReplayApp(api *API) *fiber.App` body;
- dead/helper routes, missing/duplicate `newReplayApp`, unsupported signature, and duplicate routes fail closed;
- `go run ./cmd/validate-openapi` fails if any unmarked operation is absent from `newReplayApp`;
- validator fails if a `planned` operation is accidentally shipped without lifecycle promotion;
- validator rejects `planned` on any operation outside the exact five-operation Stage-2 allowlist;
- replay constructor source-contract test requires both Stage 3.68 dividend limiter initializers;
- all existing Stage 3.68 calculator/replay tests remain green;
- all ten protected CI jobs remain required after publication.

## Internal Review evidence

The mandatory development-path Internal review was performed read-only against the complete seven-file
candidate. The reviewer accounted for every changed repository file; sampling and Builder self-review
were not accepted as independent evidence.

### First Internal review

Verdict: `REQUEST CHANGES`.

Blocking findings:

- `IR-OI007-01` — P2: the initial parity scanner inspected the whole `replay_app.go` file, so an
  unrelated dead/helper `app.<Method>(...)` call could falsely satisfy production-route parity.
  Required remediation: scope route extraction to exactly one canonical
  `func newReplayApp(api *API) *fiber.App`, fail closed on missing/duplicate/unsupported declarations,
  and add dead/helper-route regression coverage.
- `IR-OI007-02` — evidence blocker: the exact candidate did not yet have complete development-path local
  quality-gate evidence on the repository-approved Go toolchain. Required remediation: rerun the exact
  candidate gates and report the race result without converting an unavailable/failing gate into PASS.

### Internal re-review — Candidate R3

Exact candidate identity reviewed:

```text
Candidate ZIP SHA256 = 81a7f515c6596180c15770f5112bca153e0e2b1ad6c326db8586b0d095a01220
Manifest SHA256      = 40047534b8adf83040df496c4b10a89cba94033291a6cdff25c70998c3aa8948
Canonical base       = e6ccfef2984ac17075952fbe9b5d62ff834e8b69
Canonical tree       = c4287cff527a02f25c3144c8aca2cf55e30a419e
Changed files        = 7
Additions            = 756
Deletions            = 1
```

Reviewed files, all in full:

1. `backend-go/internal/httpapi/replay_app.go`
2. `backend-go/internal/httpapi/replay_app_dividends_test.go`
3. `backend-go/cmd/validate-openapi/main.go`
4. `backend-go/cmd/validate-openapi/runtime_route_parity.go`
5. `backend-go/cmd/validate-openapi/runtime_route_parity_test.go`
6. `openapi/openapi.yaml`
7. `docs/stages/OI_AUDIT_007_API_RUNTIME_PARITY_REMEDIATION.md`

Re-review disposition:

- `IR-OI007-01` — `RESOLVED`: production route extraction is scoped to exactly one canonical
  `newReplayApp(api *API) *fiber.App`; dead/helper and nested-function routes cannot satisfy parity;
  missing/duplicate/unsupported declarations fail closed; regression tests cover the failure modes.
- `IR-OI007-02` — `RESOLVED` for candidate-scoped Internal review: exact-candidate deterministic gates
  were rerun on Go 1.25.14. The overall local `go test -race ./internal/httpapi` result was explicitly
  **not** claimed as PASS because a pre-existing auth HTTP timeout was reproduced on the clean exact
  baseline. Candidate-specific replay/dividend/limiter race tests passed, and no new OI-AUDIT-007 race
  regression was demonstrated.

Internal re-review verdict: `APPROVED`.

Reviewer mutation confirmation:

```text
REVIEW_MODE = READ_ONLY
FILES_EDITED = NONE
AUTO_FIXES = NONE
STAGING = NONE
COMMIT = NONE
PUSH = NONE
PR = NONE
MERGE = NONE
```

## Published-head CI and External sequence

The reviewed runtime candidate was published as:

```text
PR = #171
RUNTIME_HEAD = fffa4ae516e21cd46e555f78b2646c3a9b5fd119
RUNTIME_TREE = 1a55192ceccc305bafcb68af15a89efc8dc930b1
BASE = develop@e6ccfef2984ac17075952fbe9b5d62ff834e8b69
CI_RUN = #519 / 34538621427
CI_RESULT = 10/10 SUCCESS
GO_RACE_TESTS = SUCCESS
```

A fresh External published-head review was then performed against that exact runtime head and returned
`APPROVED`. The External phase did not use the Internal verdict/findings as supporting evidence for its
technical conclusion. Publication of the Internal evidence in this section occurs only after that
External verdict, as required by `docs/REVIEW_WORKFLOW.md` v1.4.0.

This follow-up is evidence-only. It does not change runtime code, executable API contracts, tests,
dependencies, migrations, financial semantics, provider/source activation, or the Stage 3.78/3.79/3.80
runtime state. Required CI must run again on the evidence-follow-up PR head before human merge
authorization.

## Backward compatibility

- `POST /api/v1/dividends/calculate`: restorative/additive relative to the broken production router;
  it restores the already canonical Stage 3.68 contract.
- The five `planned` operations keep their frozen path/method/schema definitions.
- No database/schema/migration/provider/source/dependency change.
- No Stage 3.78B, Stage 3.79 or Stage 3.80 activation.

## Rollback

The runtime remediation is independently reversible and has no migration or business-data backfill.
The evidence-only follow-up can also be reverted independently without changing runtime behavior.
Reverting the runtime remediation would reintroduce the route-loss/contract-status defect.

## Governance state

```text
PUBLISHED_RUNTIME_HEAD=fffa4ae516e21cd46e555f78b2646c3a9b5fd119
DRAFT_PR=171
INTERNAL_REVIEW=APPROVED
EXTERNAL_PUBLISHED_HEAD_REVIEW=APPROVED
INTERNAL_REVIEW_EVIDENCE=PUBLISHED_IN_EVIDENCE_ONLY_FOLLOWUP
EVIDENCE_FOLLOWUP_CI=REQUIRED
READY=NOT_AUTHORIZED
MERGE=NOT_AUTHORIZED
SOURCE_ACTIVATION=NONE
STAGE_3_78B=NOT_STARTED
STAGE_3_79=NOT_STARTED
STAGE_3_80=NOT_STARTED
```
