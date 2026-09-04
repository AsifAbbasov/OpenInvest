# Stage 3.63 — Corporate Action Calendar + Heatmap Projection Implementation

| Field | Value |
| --- | --- |
| Status | Local implementation candidate; not committed/pushed; no PR/Ready/merge/API/UI/source activation authorization implied |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@fbbca6aeee7c12300a37eb5748628275aac427e4` |
| Protected-base tree | `ff311bd8f03524f83724af024afb038c4a5a246c` |
| Planning authority | `docs/stages/STAGE_03_61_CORPORATE_ACTIONS_CALENDAR_PLANNING.md` |
| Contract dependency | Stage 3.62 / Feature 3A Corporate Action Boundary |
| Feature | Feature 3B — Calendar + Heatmap Projection |
| External source activation | None |

## 1. Purpose

Implement pure deterministic dated projections over canonical `CorporateActionEvent` values without adding an external source, persistence, API, or UI.

The stage adds:

```text
[]CorporateActionEvent
        ↓
current supersession resolution
        ↓
Calendar projection
        ↓
count/density Heatmap projection
```

## 2. Candidate files

1. `backend-go/internal/verticalslice/corporateactionprojection.go`
2. `backend-go/internal/verticalslice/corporateactionprojection_test.go`
3. `docs/stages/STAGE_03_63_CORPORATE_ACTION_CALENDAR_HEATMAP_PROJECTION_IMPLEMENTATION.md`

No Feature 3A contract, provider, runtime composition, SQL/migration, OpenAPI, frontend/mobile, dependency, worker, cache, source registry, or CI workflow file is changed.

## 3. Calendar contract

`ProjectCorporateActionCalendar` first reuses canonical Feature 3A event validation, then validates projection-specific supersession integrity.

Effective-date rules are exactly Stage 3.61:

- dividend: `RecordDate` when present, otherwise `PaymentDate`;
- coupon: `PaymentDate` when present;
- no usable date: retain source evidence outside this pure function, but omit the event from dated projection.

Output is sorted deterministically by:

1. effective date;
2. instrument ID;
3. action kind;
4. event ID.

Calendar output clones pointer-owned event fields so caller mutation of projected dates, amount pointer, or supersession pointer cannot mutate the input slice.

## 4. Current evidence / supersession semantics

The projection is a current dated view, not a history store.

When event B declares `SupersedesEventID = A`, A remains evidence in the caller-owned input but is excluded from the current dated projection. B remains eligible for projection under the normal effective-date rule.

If B is `CANCELLED` and has no usable date, A disappears from the dated projection and B remains undated evidence outside this projection. This prevents a cancelled historical date from remaining visible as a current scheduled action.

A predecessor may be absent from the supplied batch because the batch may represent a time window or partial evidence history.

When the predecessor is present, the superseding event must match its `InstrumentID` and `Kind`. Cross-instrument or cross-kind supersession is rejected fail-closed.

Ambiguous forks (two events superseding the same predecessor in the supplied batch) and supersession cycles are rejected as `ErrInvalidCorporateActionProjection`.

No ordering invariant based on `AsOf`, `RetrievedAt`, provider identity, or persistence revision is invented because Stage 3.61/3.62 do not authorize one.

## 5. Heatmap contract

`ProjectCorporateActionHeatmap` derives only from the current calendar projection and aggregates by effective date.

Each bucket exposes count/density dimensions only:

- `TotalCount`;
- `DividendCount`;
- `CouponCount`;
- `AnnouncedCount`;
- `ConfirmedCount`;
- `PaidCount`;
- `CancelledCount`.

No money, yield, FX, tax, portfolio holdings, expected income, or guarantee metric is aggregated.

Status dimensions are explicit so `ANNOUNCED` events are never silently presented as confirmed or guaranteed income.

Heatmap buckets are sorted ascending by date.

## 6. Failure semantics

- malformed canonical event → preserve `ErrInvalidCorporateAction` from Feature 3A;
- duplicate event ID → preserve canonical rejection;
- cross-instrument/kind supersession → `ErrInvalidCorporateActionProjection`;
- supersession fork → `ErrInvalidCorporateActionProjection`;
- supersession cycle → `ErrInvalidCorporateActionProjection`;
- missing historical predecessor → allowed;
- empty input → valid empty calendar/heatmap;
- undated current event → omitted from dated outputs, never assigned a fabricated date.

## 7. Tests

Focused tests cover:

- dividend record-date preference;
- dividend payment-date fallback;
- coupon payment-date rule even if record date exists;
- undated omission;
- deterministic ordering and final event-ID tie-break;
- current unsuperseded projection;
- cancellation removing obsolete dated projection;
- missing predecessor outside current batch;
- cross-instrument supersession rejection;
- cross-kind supersession rejection;
- fork rejection;
- cycle rejection;
- canonical validation propagation;
- pointer-alias isolation;
- heatmap counts by kind/status;
- heatmap supersession behavior;
- heatmap error propagation;
- empty input.

Focused statement coverage for `corporateactionprojection.go`: `100.0%`.

## 8. Local deterministic gates

A narrow harness compiles the candidate against the canonical Stage 3.62 corporate-action contract and existing OpenInvest Money/Decimal/ticker seam.

- `gofmt`: PASS;
- `go test ./...`: PASS;
- `go test -race ./...`: PASS;
- `go vet ./...`: PASS;
- production projection file statement coverage: 100.0%;
- external HTTP/network code: NONE;
- direct `time.Now()`: NONE;
- `float64`: NONE;
- SQL/DB/OpenAPI/frontend/runtime composition changes: NONE.

Repository-wide evidence still requires exact-head GitHub CI after separate publication authorization.

## 9. Internal review finding resolved before freeze

Internal review found one demonstrated P2 design gap in the first local candidate: a present predecessor was not required to share instrument and action kind with its superseder, allowing an erroneous cross-instrument or cross-kind link to remove an unrelated dated event.

Remediation added fail-closed same-`InstrumentID` and same-`Kind` checks plus regression tests. No broader source/persistence/revision semantics were introduced.

After remediation: P0=0, P1=0, P2 blocking=0, P3 blocking=0.

## 10. Architectural consequences

Feature 3B creates a source-neutral pure projection layer that can be tested with fixtures today and reused unchanged by a future approved adapter/API/UI.

It does not improve external data coverage and does not make OpenInvest equivalent to commercial data products. It only ensures that once canonical events exist, calendar and density views are deterministic and do not fabricate dates, income, or lifecycle certainty.

## 11. Explicit exclusions

No external provider/HTTP/scraping, source activation, DB/migration, persistence/replay, OpenAPI/API, frontend/mobile, notifications, workers, caching, monetary heatmap, yield, FX, tax, portfolio forecast, amortization/redemption, Feature 3C, or Feature 3D.

## 12. Governance state

This is a local development-path candidate only. Internal review follows:

```text
contract → implementation → failure cases → tests → CI expectations → architectural consequences
```

Only demonstrated defects receive severity. UNKNOWN remains UNKNOWN.

Commit/push/Draft PR require separate explicit human authorization. After publication: exact-head CI → fresh External review → remediation if demonstrated → evidence publication/verification → separate Ready/squash-merge authorization.
