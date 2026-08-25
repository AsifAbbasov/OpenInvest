# Stage 3.37 - P3-02 True IANA Timezone Semantics Implementation

| Field | Value |
| --- | --- |
| Status | Runtime implementation canonical through PR #91 / closure governance pending |
| Date | 2026-08-25 |
| Planning gate | PR #90 squash-merged at `46f74528dcc19424ad087d30d4f2f778e2079b87` |
| Canonical runtime base | `develop` at `46f74528dcc19424ad087d30d4f2f778e2079b87` |
| Runtime PR | PR #91 |
| Candidate branch | `codex/stage-03-37-p3-02-timezone-runtime` |
| Initial published runtime head | `465a7f0ddfe5a7bf892ec8a735915688cdaf59ad` |
| Frozen final runtime head | `1a2f89a0fa5095b3cca790521afa484bdc61e8a6` |
| Runtime merge | `cb6d9b28cd47b1cd283b5861b916e0be627d0ac2` |
| Exact-head CI | CI #265 / run `32869754524`, 10/10 required jobs successful |
| Finding | P3-02 |
| Commit / push | Published through PR #91 and canonically squash-merged |
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

## Regression evidence

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
- HTTP proof covers both a leading-whitespace timezone (`" Asia/Baku"`) and a raw UTC-offset timezone
  (`"UTC+04:00"`), verifying the established 400 validation path and zero registration-store calls.
- OpenAPI contract proof verifies both timezone descriptions, preserves RegisterRequest 1..64 bounds,
  and rejects introduction of a handcrafted timezone pattern.

## Pre-publication local verification

Before publication, the revised local candidate passed all of the following:

- focused Stage 3.37 auth tests: PASS
- focused Stage 3.37 HTTP test: PASS
- focused Stage 3.37 OpenAPI contract test: PASS
- `go run ./cmd/validate-openapi`: PASS
- full `go test ./...`: PASS
- full `go vet ./...`: PASS
- Go toolchain: `go version go1.26.2 darwin/arm64`
- `git diff --check`: PASS
- exact six-file changed-file allowlist: PASS

At local finalization no commit or push had yet occurred. After renewed independent pre-commit
`APPROVED`, the exact reviewed runtime candidate was committed and pushed to PR #91.

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

The revised candidate added only a narrow pre-resolver lexical guard:

- compare `strings.TrimSpace(name)` with the original value and reject on inequality without trimming
  or persisting a rewritten value;
- reject the ASCII raw-offset families `±HH:MM` and `UTC±HH:MM`;
- continue accepting `UTC`;
- continue delegating all other exact identifiers to `time.LoadLocation`;
- do not reject `Etc/GMT+4` or other loadable tzdb identifiers merely because they represent fixed
  offsets; and
- retain resolver-dependent case behavior without application case normalization.

Focused and full verification was rerun after that material change. A renewed independent pre-commit
review then returned `APPROVED` with P0/P1/P2/P3 = None, after which the exact reviewed runtime
candidate was published in PR #91.

## First published-head review correction

The first published-head review of PR #91 at initial runtime head
`465a7f0ddfe5a7bf892ec8a735915688cdaf59ad` independently verified the runtime implementation and
exact-head CI #264 / run `32867005056`, but returned `REQUEST CHANGES` with one P3 documentation
blocker: this implementation record still described the already-published PR as an uncommitted local
candidate awaiting pre-commit review.

No runtime-code defect was reported in that published-head review. This documentation-only correction
updates the lifecycle state and HTTP evidence without changing runtime code, OpenAPI, tests, or finding
scope. Because the correction advances the PR head, CI #264 and the published-head review on
`465a7f0ddfe5a7bf892ec8a735915688cdaf59ad` are historical evidence only and cannot be reused as
exact-head approval for the corrected head.

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

The Stage 3.37 runtime implementation is canonical through PR #91 at
`cb6d9b28cd47b1cd283b5861b916e0be627d0ac2`, from frozen final runtime head
`1a2f89a0fa5095b3cca790521afa484bdc61e8a6` after exact-head CI #265 / run `32869754524`
and fresh published-head independent `APPROVED`.

Runtime merge alone does not close the audit finding. P3-02 remains canonically **OPEN** until
separately reviewed closure governance receives fresh independent `APPROVED`, exact-head green closure
CI after publication, explicit human closure-merge authorization, and is merged into `develop`.
