# Stage 3.37 - P3-02 True IANA Timezone Semantics Implementation

| Field | Value |
| --- | --- |
| Status | Local runtime candidate / pre-commit independent review pending |
| Date | 2026-08-25 |
| Planning gate | PR #90 squash-merged at `46f74528dcc19424ad087d30d4f2f778e2079b87` |
| Canonical runtime base | `develop` at `46f74528dcc19424ad087d30d4f2f778e2079b87` |
| Local candidate branch | `codex/stage-03-37-p3-02-timezone-runtime` |
| Finding | P3-02 |
| Commit / push | Not performed; separate authorization required |
| Runtime merge authorized here | No |

## Objective

Implement the approved P3-02 resolver-based IANA timezone admission contract without changing
financial BusinessDate semantics, SQL DATE behavior, database schema, historic persisted preferences,
frontend behavior, or any other remaining P3 finding.

## Candidate implementation

- Registration validates the exact submitted timezone string; no trimming, case conversion, alias
  rewriting, or other normalization is introduced.
- Empty input and the existing 64-byte admission overflow are rejected before resolver admission.
  The byte/code-point parity question remains P3-04 and is not claimed as closed here.
- `Local` is explicitly rejected before `time.LoadLocation` because Go resolves it to host-local
  configuration.
- `UTC` remains explicitly accepted for compatibility.
- Every other exact value is accepted only when `time.LoadLocation(name)` succeeds.
- `_ "time/tzdata"` is imported as standard-library fallback availability. The candidate does not
  claim it overrides `ZONEINFO`, system zoneinfo, or `$GOROOT/lib/time/zoneinfo.zip`.
- Loadable tzdb identifiers such as `Etc/GMT+4` remain accepted; surrounding-whitespace forms and
  raw ASCII UTC-offset spellings in the `±HH:MM` / `UTC±HH:MM` families are rejected by a narrow
  application pre-resolver syntax guard, without rewriting the submitted value.
- Case sensitivity is intentionally not hard-coded. The exact string is passed unchanged to
  `time.LoadLocation`; whether a case-changed path resolves can depend on the active timezone-data
  source/filesystem. No application normalization occurs.
- OpenAPI `User.timezone` and `RegisterRequest.timezone` descriptions publish the same semantics
  without a handcrafted IANA regex/list.
- No PostgreSQL migration or automatic historical-value rewrite is introduced.

## Regression evidence in the local candidate

- Auth service corpus covers `UTC`, `Asia/Baku`, `Europe/Berlin`, `America/New_York`,
  `Etc/GMT+4`, empty/whitespace input, `Local`, unknown/invented zones, raw offsets, surrounding
  whitespace, path traversal/absolute-path variants, and the existing 64-byte bound.
- A pure pre-resolver syntax-guard test proves that surrounding whitespace and the ASCII
  `±HH:MM` / `UTC±HH:MM` raw-offset families are rejected independently of `time.LoadLocation`,
  while `UTC`, `Asia/Baku`, and `Etc/GMT+4` are not rejected by that lexical guard.
- The test suite deliberately does not assert that `asia/baku` must be rejected, because the approved
  contract is resolver-based and `LoadLocation` source behavior can differ by deployment filesystem.
  The application still performs no case normalization.
- Service proof verifies an accepted timezone is persisted exactly as submitted and invalid timezone
  admission never reaches `Store.RegisterUser`.
- HTTP proof sends a leading-whitespace timezone and verifies the established 400 validation path plus
  zero registration-store calls, proving the transport does not silently trim it.
- OpenAPI contract proof verifies both timezone descriptions, preserves RegisterRequest 1..64 bounds,
  and rejects introduction of a handcrafted timezone pattern.

## Local verification

The candidate passed all of the following before this implementation record was finalized:

- focused Stage 3.37 auth tests: PASS
- focused Stage 3.37 HTTP test: PASS
- focused Stage 3.37 OpenAPI contract test: PASS
- `go run ./cmd/validate-openapi`: PASS
- full `go test ./...`: PASS
- full `go vet ./...`: PASS
- Go toolchain: `go version go1.26.2 darwin/arm64`

`git diff --check` and the exact changed-file allowlist are verified by the finalizer script after this
document is written. No commit or push is performed.

## Adversarial test correction

The first local candidate test incorrectly assumed that the case-changed identifier `asia/baku` must
always fail. On the current macOS environment, `time.LoadLocation("asia/baku")` resolves successfully.
That is compatible with the approved resolver-based contract because `LoadLocation` may use
environment-specific higher-precedence timezone data and filesystems. The runtime implementation was
not tightened to impose a custom case rule; instead, the over-constrained test was removed.

## Independent pre-commit review blocker remediation

The first independent pre-commit runtime review returned `REQUEST CHANGES` with one P3 blocker:
surrounding-whitespace and raw-offset invalidity was not invariant under the approved `LoadLocation`
source model because a deliberately configured higher-precedence `ZONEINFO` directory can contain
valid TZif files under those exact names.

The revised candidate adds only a narrow pre-resolver lexical guard:

- compare `strings.TrimSpace(name)` with the original value and reject on inequality without trimming
  or persisting a rewritten value;
- reject the ASCII raw-offset families `±HH:MM` and `UTC±HH:MM`;
- continue accepting `UTC`;
- continue delegating all other exact identifiers to `time.LoadLocation`;
- do not reject `Etc/GMT+4` or other loadable tzdb identifiers merely because they represent fixed
  offsets; and
- retain resolver-dependent case behavior without application case normalization.

Focused and full verification below is rerun after this material change. A fresh independent
pre-commit review is required; the earlier `REQUEST CHANGES` cannot be reused as approval.

## Residual risk

Historical users may already contain invalid timezone strings admitted before P3-02 remediation. This
candidate does not guess or rewrite those preferences because the repository has no authoritative fact
from which to infer the intended user timezone. Any future preference-update path must reuse the same
resolver-based admission rule.

Deployment-specific higher-precedence timezone data can still affect `time.LoadLocation` resolution,
including case behavior. That behavior is the approved planning model and remains governed operational
configuration rather than an embedded-only tzdb guarantee.

## Non-scope

No BusinessDate/SQL DATE conversion, UTC SystemTimestamp change, transaction/dividend/settlement/
snapshot economic-date change, Decimal work, P3-04 general Unicode/maxLength remediation, cleanup
lifecycle work, `api.go` decomposition, fixture cleanup, migration-policy work, Next.js/Fiber
maintenance, database migration, frontend change, provider work, or Stage 3.25 privacy work is included.

## Canonical status

This is an uncommitted local runtime candidate only. P3-02 remains OPEN. The candidate requires fresh
independent pre-commit review, separate explicit commit/push authorization, exact-head CI after
publication, fresh published-head independent review, separate explicit runtime squash-merge
authorization, canonical merge, and separately governed closure evidence before P3-02 may be CLOSED.
