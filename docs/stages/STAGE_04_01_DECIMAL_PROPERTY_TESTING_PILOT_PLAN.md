# Stage 4.1 - Decimal Property-Testing Pilot Plan

| Field | Value |
| --- | --- |
| Status | Planning-only candidate; no implementation authorization |
| Date | 2026-09-04 |
| Parent candidate | Stage 4 Commercial Readiness Assurance Plan at `875ca89834b895b0faf27d80940fe59136bfdfd5` |
| Protected planning baseline | `origin/develop@983104267221706c3c2ebd8d9be358e3921334b5` |
| Current branch base | `875ca89834b895b0faf27d80940fe59136bfdfd5`; stacked until the parent candidate is merged or this scope is rebased |
| Target | `backend-go/internal/decimal` ingress grammar, canonical string form, and JSON-string transport |
| Runtime, API, schema, migration, dependency, CI, infrastructure change | None in this plan |
| Out of scope | Decimal arithmetic, imports, idempotency, storage, snapshots, fuzzing, load testing, production data, external services, and commercial-readiness claims |

## 1. Objective

Add a small, reproducible property-testing pilot for the current Decimal boundary without changing
the financial domain. The implementation candidate must generate valid and invalid decimal spellings
from the published grammar, validate canonical parse/string/JSON round trips, and leave every existing
runtime contract unchanged.

This is a test-only pilot, not a claim that all financial invariants have property coverage. It does
not replace the existing table-driven boundary tests; it complements them with a broad, bounded
generated input set.

## 2. Frozen implementation boundary

The eventual implementation may add a test file in `backend-go/internal/decimal` and test-only
helpers in that package. It must not modify:

- `decimal.Decimal`, `FromString`, `FromLegacyStringForReplay`, arithmetic, or storage behavior;
- OpenAPI Decimal schemas, CSV normalization, import parser versions, replay-token semantics, or
  PostgreSQL NUMERIC definitions;
- Go production dependencies, module versions, CI workflow, database fixtures, or public APIs.

The pilot uses Go's standard `testing/quick` and `math/rand` packages. It adds no Go module
dependency. The implementation must set `quick.Config.MaxCount` explicitly to 10,000 for each
generated property and set a fixed `math/rand` source; it must not rely on the moving
`-quickchecks` default or a time-derived seed.

## 3. Generated input contract

Valid generated inputs must independently construct only the current fresh-write grammar:

```text
-?(0|[1-9][0-9]{0,19})(\.[0-9]{1,8})?
```

The generator must cover all of these deliberately rather than relying on an unconstrained string
generator:

| Family | Required cases |
| --- | --- |
| Sign | no sign, negative non-zero, and negative zero |
| Integer | zero, one digit, multi-digit non-zero prefix, and exactly 20 digits |
| Fraction | absent, one digit, eight digits, all zeroes, and non-zero suffix |
| Boundary | `99999999999999999999.99999999`, `0.00000000`, and `-0` |
| Composition | deterministic generated combinations plus the explicit boundary cases |

Invalid generated mutations must start from a generated valid spelling or an explicit boundary case
and include at least: plus sign, leading zero on a multi-digit whole part, ASCII edge whitespace,
embedded whitespace, trailing dot, extra fractional digit, second dot, exponent marker, comma,
non-ASCII decimal digit, empty string, and a 21-digit whole part. Each mutation is required to be
rejected by fresh `FromString`.

The legacy replay parser is explicitly excluded. A property of fresh grammar must never be interpreted
as permission to reject an already-completed historical replay.

## 4. Properties and independent checks

| ID | Property | Required assertion |
| --- | --- | --- |
| DEC-PROP-01 | Valid fresh grammar is accepted | Every generated valid spelling returns a Decimal without error |
| DEC-PROP-02 | Canonical representation is stable | `FromString(input).String()` itself matches the canonical fresh grammar with exactly eight fractional digits; `-0` may canonically render as non-negative zero |
| DEC-PROP-03 | Parse/string round trip preserves value | Parsing the canonical string succeeds and is `Equal` to the first parsed Decimal |
| DEC-PROP-04 | JSON remains a string boundary | `MarshalJSON` decodes as a JSON string exactly equal to `Decimal.String()`; parsing that string preserves value |
| DEC-PROP-05 | Selected non-contract forms fail closed | Each generated invalid mutation returns an error from fresh `FromString` |

The implementation must construct the valid grammar independently in test code. It may use a compact
test-only regular expression to verify canonical output, but it must not call unexported production
parsing helpers or reproduce `FromString` by calling it as the oracle.

The following are intentionally not properties in this pilot:

- multiplication, division, banker rounding, overflow, or arithmetic identities, because they require
  a separately reviewed independent oracle and storage-bound policy;
- CSV trim behavior, parser-version replay, or review tokens, because those belong to the import
  boundary;
- persisted idempotency, concurrent append, or snapshot determinism, because they need transaction
  and PostgreSQL evidence.

## 5. Reproduction and counterexample policy

The generator seed, property ID, input spelling, and mutation name must be included in every failure.
A failing command is reproduced with:

```sh
cd backend-go
go test ./internal/decimal -run 'TestPropertyDecimal' -count=1
```

`testing/quick` provides deterministic case generation with the supplied seed but no structural
shrinker. For this bounded pilot, the exact quoted input is the immediate reproducer. Before closing
a defect found by the pilot, remediation must add a smallest practical table-driven regression case
and retain the original seed/input in the issue or review evidence. Introducing an automatic shrinking
framework is a separate dependency/tooling decision, not an implicit addition in this scope.

## 6. Execution and acceptance

The implementation must run:

```sh
cd backend-go
go test ./internal/decimal -count=1
go test ./...
```

It must also pass the repository's normal verification command. A result is accepted only if:

1. every property has exactly the fixed seed and 10,000 generated cases;
2. explicit boundary cases and every invalid-mutation family are present;
3. no runtime, schema, OpenAPI, dependency, import, replay, or CI behavior changes;
4. failure output is sufficient to reproduce an exact case without a production input;
5. the existing table-driven Decimal tests remain and pass; and
6. review confirms that generated coverage is not represented as production, security, fuzzing, or
   commercial-readiness evidence.

## 7. Delivery and rebasing rule

This plan is stacked on the unmerged Stage 4 planning candidate. It does not authorize a PR or merge.
Before implementation work is treated as current governance authority, the parent plan must either be
merged into protected `develop` or the implementation branch must be explicitly rebased onto the
then-current protected base and its plan references revalidated.

Any later implementation follows the effective review workflow. Per the current Principal Architect
directive, this task uses internal review only and does not initiate an external review.
