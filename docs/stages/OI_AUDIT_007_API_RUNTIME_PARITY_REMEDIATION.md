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
| Commit / push / PR | NOT AUTHORIZED / NONE |

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

Changed-file count: **7**, inside the <=25 review budget.

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

## Backward compatibility

- `POST /api/v1/dividends/calculate`: restorative/additive relative to current broken production router;
  it restores the already canonical Stage 3.68 contract.
- The five `planned` operations keep their frozen path/method/schema definitions.
- No database/schema/migration/provider/source/dependency change.
- No Stage 3.78B, Stage 3.79 or Stage 3.80 activation.

## Rollback

Revert this remediation candidate. No migration or business-data backfill exists. Reverting would
reintroduce the current route-loss/contract-status defect and is therefore technically simple but not
desirable.

## Governance state

```text
IMPLEMENTATION_CANDIDATE=LOCAL_ONLY
LOCAL_COMMIT=NONE
PUSH=NONE
PR=NONE
INTERNAL_REVIEW=REQUIRED
SOURCE_ACTIVATION=NONE
STAGE_3_78B=NOT_STARTED
```
