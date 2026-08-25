# Stage 3.36 - P3-03 OpenAPI Decimal Grammar Implementation

| Field | Value |
| --- | --- |
| Status | Runtime implementation canonical through PR #88 / closure governance pending |
| Date | 2026-08-25 |
| Planning gate | PR #87 squash-merged at `251296e0831cbb0b81c7799cc82cbdf3b451ae6` |
| Finding | P3-03 |
| Frozen runtime head | `131f1bf963e9d232b9e23273edd54caf54c10ffb` |
| Runtime merge | `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa` |
| Runtime merge authorized here | No |

## Objective

Implement the approved Decimal lexical contract without changing arithmetic, PostgreSQL storage,
the public OpenAPI grammar, migrations, or any P3 finding other than P3-03.

## Implementation

- `decimal.FromString` now admits only the complete published ASCII Decimal grammar:
  `^-?(0|[1-9][0-9]{0,19})(\.[0-9]{1,8})?$`. In the double-quoted YAML source, the
  literal decimal point is written as `\\.`; after YAML parsing, the regex token is `\.`.
- Admission rejects an input longer than the grammar-derived 30-byte maximum before fixed-scale
  `big.Int` conversion. Existing valid values retain eight-place output and half-even arithmetic;
  `-0` still normalizes to zero.
- CSV retains its existing field-edge `TrimSpace` normalization, then uses the same strict parser.
  It does not retain a CSV-only permissive grammar.
- The import review parser version is bumped from `1` to `2`. A parser-v1 review token cannot
  authorize a new append after deployment.
- The completed-command recovery path verifies the token HMAC before selecting historic parser
  semantics. It then rechecks the complete historical proof at its issuance instant before any
  lookup: token family/lifetime shape, parser version, subject, portfolio, source context, file
  hash, semantic digest, row identities, appendable rows, and submitted decisions. Only then may it
  perform an exact, read-only artifact lookup scoped by principal, canonical path, idempotency key,
  and canonical request hash.
- The pre-v2 Decimal parser is available only inside that authenticated completed-replay
  reconstruction. It is not reachable from review or fresh append admission.
- A dedicated OpenAPI validator test compares the literal Decimal schema pattern and the runtime
  parser over accepted and rejected corpus vectors.

## Regression Evidence

- Decimal corpus: documented forms accepted; leading plus/zeroes, whitespace, empty fraction,
  scientific notation, separators, Unicode digits, and oversized lexical input rejected.
- HTTP transaction corpus: prohibited JSON Decimal spellings return the established `400`
  validation result before the store append method is called.
- CSV corpus: prohibited normalized Decimal spellings cannot become appendable; field-edge
  whitespace still normalizes to a valid Decimal before admission.
- Parser-version transition: a v1 token cannot create a new write. A v1 completed import is replayed
  exactly after the v2 deployment for both `001.25` and a contract-conforming Decimal, preserving
  status, body, request ID, trace ID, and zero new financial writes.
- Recovery rejects a new key or principal, bad HMAC, changed source context, changed raw CSV payload
  (even with a recomputed file hash), changed row identity, changed decision, or a different
  canonical path without returning the historic artifact or creating a new write.
- PostgreSQL round-trip: `TestStoreVerticalSlice` passed against a fresh disposable database after
  applying migrations `000001` through `000006`, exercising transaction, summary, and list paths.

## Verification

The final merge-candidate head `131f1bf963e9d232b9e23273edd54caf54c10ffb` passed GitHub Actions
CI #257 / run `32822925542` with all 10 required checks successful. The implementation also records
focused Decimal/import/HTTP/OpenAPI PASS, full `go test ./...` PASS, full `pnpm run verify` PASS, and
`git diff --check` PASS.

## Non-Scope

No migration, stored-data rewrite, Decimal arithmetic or rounding change, float conversion, JSON
numeric coercion, OpenAPI grammar change, snapshot change, cleanup lifecycle work, timezone work,
Unicode/maxLength work, HTTP decomposition, dependency maintenance, or Stage 3.25 privacy work is
included.

## Canonical Status

The Stage 3.36 runtime implementation is canonical through PR #88 at
`ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa`. The runtime merge alone does not close the audit
finding. P3-03 remains canonically **OPEN** until separately reviewed closure governance receives
fresh independent `APPROVED`, exact-head green CI, explicit human closure-merge authorization, and
is merged into `develop`.
