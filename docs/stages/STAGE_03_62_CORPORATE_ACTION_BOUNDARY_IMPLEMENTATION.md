# Stage 3.62 — Corporate Action Boundary Implementation

| Field | Value |
| --- | --- |
| Status | Local implementation candidate; not committed/pushed; no PR/Ready/merge/external ingestion/API/UI/persistence authorization implied |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@0e449b4d729e4388081be4990c034c67c5e5019a` |
| Protected-base tree | `f7c5daaffc0a37350e8ac7e21cd83dd30b3c8f0a` |
| Planning authority | `docs/stages/STAGE_03_61_CORPORATE_ACTIONS_CALENDAR_PLANNING.md` |
| Feature | Feature 3A — Corporate Action Boundary |
| External source activation | None |

## 1. Purpose

Implement the smallest provider-neutral corporate-actions boundary required by canonical Stage 3.61 without coupling OpenInvest to Interfax, NSD, MOEX, issuer HTML, a database schema, or a public API.

The candidate provides:

```text
future approved adapter
        ↓
CorporateActionProvider
        ↓
FetchCorporateActions
        ↓
CorporateActionEvent validation
        ↓
future pure Calendar / Heatmap projections
```

No external HTTP or production provider exists in Feature 3A.

## 2. Candidate files

1. `backend-go/internal/verticalslice/corporateactions.go`
2. `backend-go/internal/verticalslice/corporateactions_test.go`
3. `docs/stages/STAGE_03_62_CORPORATE_ACTION_BOUNDARY_IMPLEMENTATION.md`

No existing runtime composition, OpenAPI, SQL/migrations, frontend/mobile, dependency, worker, cache, source registry, or CI workflow file changes.

## 3. Canonical model

Feature 3A defines provider-neutral concepts:

- `CorporateActionKind`: `DIVIDEND`, `COUPON`;
- `CorporateActionStatus`: `ANNOUNCED`, `CONFIRMED`, `PAID`, `CANCELLED`;
- stable application `EventID`;
- canonical `InstrumentID` using the existing OpenInvest ticker-shaped identity contract;
- optional `RecordDate`;
- optional `PaymentDate`;
- optional exact-decimal `AmountPerUnit`;
- optional `SupersedesEventID` for correction/cancellation evidence linkage;
- source `AsOf`;
- OpenInvest `RetrievedAt`;
- mandatory `CorporateActionProvenance` with canonical provider identifier and opaque source event identity.

`SupersedesEventID` is intentionally minimal. It allows a new immutable evidence event to point to earlier evidence without deleting or overwriting that earlier event. Persistence, revision sequencing, graph closure, replay and supersession uniqueness remain deferred as Stage 3.61 requires.

## 4. Unknown and money semantics

Unknown amount, record date and payment date remain absent (`nil`). The boundary never converts unknown values to zero or fabricates a date.

When amount exists:

- canonical `decimal.Decimal` is reused;
- no `float64` path exists;
- amount must be non-negative and fit canonical NUMERIC(28,8) storage bounds;
- currency must be a canonical three-letter uppercase code.

Feature 3A does not perform FX conversion, yield calculations, tax, portfolio income projection, or monetary heatmap aggregation.

## 5. Identity and provenance

Application `EventID` is strict and application-owned. Provider source IDs are deliberately not forced into the same ASCII schema: `SourceEventID` is treated as an opaque non-empty, trimmed value with a bounded size and no control characters.

This avoids leaking or constraining future provider-specific identity formats while still rejecting malformed provenance.

Provider identifiers remain canonical uppercase tokens. No provider HTML, JSON shape, URL, NSD code, Interfax message type, or scraper selector enters the canonical model.

## 6. Time and date semantics

Business dates use canonical `YYYY-MM-DD` strings and reject empty, padded or impossible dates.

`AsOf` and `RetrievedAt` are mandatory UTC timestamps. Feature 3A does not invent an `AsOf <= RetrievedAt` invariant because Stage 3.61 does not authorize one and provider evidence can have different timing semantics.

The query requires a non-empty unique instrument set and an inclusive canonical `from`/`to` window with `to >= from`.

## 7. Provider boundary and fail-closed behavior

`CorporateActionProvider` is narrow:

```go
CorporateActions(ctx context.Context, query CorporateActionQuery) ([]CorporateActionEvent, error)
```

`FetchCorporateActions` owns application-side boundary enforcement:

1. validate query before provider invocation;
2. reject nil and typed-nil providers as `ErrCorporateActionsProviderUnavailable`;
3. copy the instrument slice before handing the query to a provider so a provider cannot mutate caller-owned query storage;
4. preserve caller context cancellation/deadline and provider-neutral `ProviderUnavailable` / `ProviderData` errors;
5. normalize any unclassified provider-specific error to `ErrCorporateActionsProviderUnavailable` without leaking provider details;
6. validate every returned event;
7. reject duplicate application event IDs;
8. reject events for instruments outside the requested set;
9. accept an empty result without fabricating events.

No generic plugin framework, provider registry, Kafka/event bus, worker, retry policy, polling, cache or persistence layer is added.

## 8. Correction and cancellation representation

Feature 3A does not define a database revision model. It only ensures the canonical event can represent immutable evidence chains:

```text
Event A
  ↓ superseded by
Event B (corrected or cancelled evidence)
```

A self-reference or malformed supersession identity is rejected. The target is not required to be present in the same returned batch because it may be historical evidence from an earlier retrieval. Cross-batch persistence and replay rules remain explicitly deferred.

## 9. Deterministic fixture provider and tests

The fixture provider is test-only and performs no network I/O.

Tests cover:

- all four lifecycle statuses;
- dividend and coupon kinds;
- unknown amount/date preservation;
- correction/cancellation linkage without deleting earlier evidence;
- exact decimal amount validation;
- nil/non-canonical decimal rejection;
- canonical currency validation;
- canonical and invalid business dates;
- required UTC `AsOf`/`RetrievedAt`;
- strict application event identity;
- opaque provider source identity;
- provenance validation;
- duplicate event IDs;
- valid/invalid/duplicate instrument query inputs;
- invalid and reversed date windows;
- validation before provider invocation;
- nil and typed-nil provider fail-closed behavior;
- provider-neutral unavailable/data error passthrough;
- caller context cancellation/deadline preservation;
- unclassified provider-error normalization without provider-detail leakage;
- malformed provider output rejection;
- returned-instrument scope enforcement;
- provider query-slice mutation isolation;
- empty result without fabrication.

Focused harness statement coverage for `corporateactions.go`: `100.0%`.

## 10. Local deterministic gates

The execution sandbox cannot clone the repository and uses Go 1.23 while canonical `go.mod` requires the repository toolchain. A narrow harness compiled the candidate unchanged against the current OpenInvest ticker/Money/decimal seam.

Focused evidence:

- `gofmt`: PASS;
- `go test ./...`: PASS;
- `go test -race ./...`: PASS;
- `go vet ./...`: PASS;
- `go test -cover ./internal/verticalslice`: PASS, 100.0% statements for the candidate boundary harness;
- external HTTP/network code: NONE;
- direct `time.Now()` in candidate production file: NONE;
- `float64` in candidate production file: NONE;
- SQL/DB/OpenAPI/frontend/runtime composition changes: NONE.

Canonical repository-wide evidence still requires exact-head GitHub CI after separately authorized publication.

## 11. Architectural consequences

The boundary keeps the product independent from the currently unavailable paid corporate-actions feeds. A future approved adapter can map provider evidence into the canonical event contract without changing Calendar/Heatmap projection logic.

This implementation intentionally does **not** make OpenInvest competitive with paid market-data products. It only creates a correct seam so future licensed data can be integrated without redesigning the domain layer.

## 12. Explicit exclusions

No:

- Interfax/NSD/MOEX/issuer adapter;
- scraping;
- external HTTP;
- persistence or migration;
- API/OpenAPI;
- Calendar projection;
- Heatmap projection;
- frontend/mobile;
- notifications;
- portfolio forecast;
- FX/yield/tax;
- amortization/redemption;
- source registry activation;
- Feature 3B/3C/3D implementation.

## 13. Governance state

This is a local development-path candidate only. Mandatory Internal read-only review must assess:

```text
contract → implementation → failure cases → tests → CI expectations → architectural consequences
```

Only demonstrated defects receive P0/P1/P2/P3 findings. UNKNOWN remains UNKNOWN.

If Internal review approves the frozen candidate, commit/push/Draft PR still require separate explicit human authorization. Publication then requires exact-head CI, fresh External review, remediation if needed, evidence publication/verification, and a separate human Ready/squash-merge gate under `docs/REVIEW_WORKFLOW.md` v1.4.0.

## 14. Published review and evidence chronology

This section is evidence-only and records the development-path publication history after the fresh External published-head phase. Sections 1–13 remain the historical prepublication implementation record.

### 14.1 Frozen prepublication subject

- canonical base: `develop@0e449b4d729e4388081be4990c034c67c5e5019a`;
- base tree: `f7c5daaffc0a37350e8ac7e21cd83dd30b3c8f0a`;
- frozen three-file manifest SHA-256: `d5885cb9ea3878b3936d6a5c0bd3457bfaebe2fc1587037dbadcd038de4e75fe`;
- Internal review SHA-256: `5d4e398214bc2bcb181c710cc1a1460987dda994bad0d4eda818c0a992016b60`;
- Internal verdict: `APPROVED`, blocking findings none.

### 14.2 Initial published subject and External findings

Initial Draft PR #126 subject:

- head `28aa90a85ac5363405dc38c8ebf77c894c1d42ea`;
- tree `3d5cb9b990a543e0de16c9531be65c0e094aeb9f`;
- exactly three changed files;
- frozen Git blobs matched local `git hash-object` identities.

Fresh External review COMMENT `5116906103` returned `REQUEST CHANGES` with two demonstrated P3 findings and no P0/P1/P2 findings:

1. opaque `SourceEventID` validation rejected C0/DEL controls but did not reject Unicode C1 controls despite the documented no-control-character contract;
2. Section 7 named `ErrCorporateActionsUnavailable` instead of the actual `ErrCorporateActionsProviderUnavailable` sentinel.

### 14.3 Remediation and corrected semantic head

Remediation remained within the same three-file feature scope:

- `validateOpaqueCorporateActionID` now uses `unicode.IsControl`;
- a regression test rejects a C1 control code point;
- Section 7 uses the exact canonical sentinel name.

Corrected semantic publication:

- head `97089869e8f3c5ccaf14a9aecd4927d8e2c2eb85`;
- tree `67c566b5c902ff33c0285a0cc73ac618fe6fd1ff`;
- CI #337 / run `33909927892`: all ten required jobs `SUCCESS`;
- fresh External re-review COMMENT `5116978636`: `APPROVED`, P0=0, P1=0, P2=0, P3 blocking=0.

No HTTP/source activation, persistence, migration, OpenAPI/API, frontend/mobile, Calendar/Heatmap projection, worker/cache, or Feature 3B/3C/3D surface was added by remediation.

### 14.4 Evidence-publication rule

This section changes documentation/evidence only. It does not authorize source activation, external ingestion, persistence, API/UI, Calendar/Heatmap projection, or later Feature 3 stages.

The evidence-publication head itself must pass all ten required CI jobs. Exact evidence verification must confirm that the transition from corrected semantic head changes only this implementation record and introduces no runtime/test semantic drift. Ready/squash merge remains a separate human authorization gate.
