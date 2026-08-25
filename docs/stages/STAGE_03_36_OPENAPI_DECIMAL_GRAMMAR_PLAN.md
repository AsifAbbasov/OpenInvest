# Stage 3.36 - P3-03 OpenAPI Decimal Grammar Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only |
| Date | 2026-08-24 |
| Canonical base | `develop` at `876ce64c3992fa579174766b97301f6eb0a193d6` |
| Finding | P3-03 |
| Scope | Contract-to-parser lexical parity for the public Decimal value only |
| Runtime implementation authorized here | No |

## 1. Finding / symptom

P3-03: OpenAPI describes Decimal with a strict lexical grammar, but the Go Decimal parser accepts
additional spellings.

The published schema accepts an optional minus sign, either zero or an integer without leading zeroes,
and an optional fractional part of one through eight ASCII digits. It rejects a plus sign, leading
zeroes on nonzero integers, surrounding whitespace, and a decimal point with no following digit.
`backend-go/internal/decimal.FromString`, however, trims whitespace, accepts `+`, permits arbitrary
leading zeroes, and accepts `1.`. The parser also currently allows an unbounded run of leading zeroes
because its precision check removes them before measuring the input.

Examples currently accepted by the runtime but prohibited by the OpenAPI Decimal schema include:

- `+1`
- `001.25`
- ` 1.25 `
- `1.`
- a long input such as `0000000000000000000000000000000001`

This is a contract/runtime discrepancy. It does not indicate binary floating-point use, an arithmetic
rounding defect, a PostgreSQL storage overflow, or corruption of existing financial entries.

## 2. Existing guarantees and risk

The Stage 2 canonical model requires a base-10 JSON string with no more than eight fractional digits,
and PostgreSQL stores the value as `NUMERIC(28,8)`. The OpenAPI Decimal schema already represents the
numeric capacity as at most 20 integer digits plus at most eight fractional digits.

The mismatch makes generated clients and direct API callers disagree about which request bodies are
valid. It also leaves parser work proportional to a potentially unbounded leading-zero prefix for an
otherwise small monetary value. Silently broadening the published schema to reproduce the parser's
permissive spellings would weaken the public contract and fail to fix that admission bound.

This finding is P3 because the existing value type remains fixed-scale and exact, and no financial
corruption from the alternate spellings is demonstrated. It is not P3 because the spellings are already
rejected: direct HTTP input currently reaches the permissive parser, and CSV does so after its separate
field-edge normalization. Nevertheless, all financial command input must have one explicit,
reproducible admissible language.

## 3. Chosen lexical contract

The future implementation must make the Go parser and the public OpenAPI Decimal schema accept the
same complete UTF-8 string language:

```text
^-?(0|[1-9][0-9]{0,19})(\\.[0-9]{1,8})?$
```

The grammar is ASCII-only and whole-string anchored. It permits `0`, `-0`, `1`, `-1`, `0.5`, and up to
20 integer plus eight fractional digits. It rejects a leading plus, whitespace, scientific notation,
locale separators, signs separated from digits, an empty fraction, leading zeroes other than the
single zero, a decimal point with no digits, and any value outside `NUMERIC(28,8)` lexical capacity.

The internal Decimal value remains fixed at eight fractional digits. Thus a valid input such as `1.2`
continues to serialize as `"1.20000000"`; output formatting is not a request for callers to submit only
eight fractional digits. The existing treatment of negative zero is retained: `-0` is an admissible
input spelling and normalizes to the zero value. No new canonical-sign or display policy is introduced.

This stage does not change semantic range validation such as positive/nonnegative command fields. Those
rules remain the responsibility of the existing OpenAPI composed schemas and application invariants.

## 4. Compatibility boundary

The contract has already prohibited the newly rejected spellings, so no conforming public API request
loses validity. Requests using the permissive spellings were never contract-conforming and must fail as
ordinary client validation errors after remediation rather than being normalized silently.

Persisted financial values are decoded from canonical database text produced by the application and
must remain readable. The implementation must prove that every database-returned Decimal string and
every fixture/catalog Decimal string in the repository is already in the chosen grammar before parser
strictness is changed. It must not rewrite historical ledger rows, snapshots, imports, or audit events.

CSV is a user-supplied financial input boundary that ultimately invokes the same Decimal parser. Its
existing record normalization trims field-edge whitespace before decimal admission, which is separate
from the public JSON Decimal grammar and remains unchanged. After that explicit normalization, CSV
values using a newly rejected lexical spelling must fail as row validation errors before an append plan
or ledger mutation. It must not add a CSV-only permissive escape hatch for plus signs, leading zeroes,
empty fractions, scientific notation, or excess digits.

## 5. Future implementation boundary

After this plan receives the required review and approval, the implementation may be limited to:

- `backend-go/internal/decimal/decimal.go` and focused unit tests, using a bounded whole-string
  lexical check before allocation or fixed-scale conversion;
- `backend-go/internal/httpapi` focused tests proving invalid Decimal spellings map to the existing
  deterministic HTTP 400 validation path and do not call storage;
- `backend-go/internal/importer` focused tests proving prohibited CSV Decimal spellings produce
  review errors and cannot reach append. Because this changes normalized candidate/status semantics,
  the implementation must bump `importer.ReviewParserVersion` and preserve the signed review-token
  invalidation contract;
- `backend-go/internal/postgres` focused disposable-PostgreSQL integration tests proving canonical
  persisted Decimal text remains readable through the affected transaction and summary paths;
- `backend-go/cmd/validate-openapi` contract-parity validation and tests, if needed to prevent future
  drift between the literal schema pattern and the Go acceptance corpus;
- `openapi/components/schemas.yaml` only if a clarification is necessary. The grammar itself should
  remain unchanged unless the implementation proves a standards-level ambiguity in its use; and
- the required Stage 3.36 implementation and closure-governance evidence.

The parser must not use a regular expression as the only protection against excessively long input if
the implementation would still allocate or scan an unbounded string first. It must impose a small,
explicit maximum input length derived from the grammar before expensive conversion. Error text remains
an internal detail; HTTP behavior must continue to expose only the existing validation taxonomy.

The implementation must not loosen the schema, accept Unicode digits, use floating point, add numeric
coercion from JSON numbers, alter Decimal arithmetic or rounding, alter PostgreSQL precision/scale,
make a migration, change snapshots, or split unrelated HTTP code.

## 6. Required regression proof

### Parser and contract parity

1. A conformance corpus proves that the parser accepts every representative published form: zero,
   negative zero, signed and unsigned nonzero integers, one through eight fraction digits, and each
   integer/fraction precision boundary.
2. The same corpus proves rejection of `+1`, `001`, `01.0`, surrounding or embedded whitespace,
   `1.`, `.1`, `1.000000000`, scientific notation, separators, Unicode digits, bare signs, and a
   long leading-zero prefix.
3. Boundary vectors prove acceptance of 20 integer plus eight fractional digits and rejection of 21
   integer digits, including when leading zeroes attempt to conceal the width.
4. A parser/OpenAPI parity test fails if an admissible/rejected corpus vector has different outcomes
   under the contract schema and the runtime parser. It must exercise whole-string matching, not a
   substring match that would hide a newline or suffix discrepancy.
5. Valid forms retain fixed-eight-place decimal serialization and exact arithmetic. Negative zero
   keeps its current normalized zero behavior.

### HTTP and import boundaries

6. Table-driven strict JSON command tests prove `+1`, leading zeroes, surrounding or embedded
   whitespace, `1.`, a long leading-zero prefix, and every other divergent corpus entry return the
   established HTTP 400 validation error before any application/store write.
7. Valid Decimal forms still reach the existing transaction command path without behavior changes.
8. Table-driven CSV review tests prove that `+1`, leading zeroes, `1.`, a long leading-zero prefix,
   and every other prohibited form after field-edge normalization return a row validation error, have
   no approved append decision, and cannot affect the ledger or snapshots. A field-edge-whitespace
   vector must instead prove the established CSV behavior: it normalizes to the valid scalar `1.25`
   before Decimal admission and is not treated as a raw-JSON grammar violation.
9. Valid CSV Decimal values continue through parse, review, explicit approval, and append under the
   existing financial-identity and reconciliation rules.
10. A token/review issued with the prior `ReviewParserVersion` cannot authorize append after the
    semantic change and requires a fresh review. A newly issued token under the bumped version still
    authorizes an otherwise valid review; the regression must exercise the existing signed-token
    verification path rather than a test-only shortcut.

### Persistence and bounded admission

11. A disposable-PostgreSQL integration test proves canonical stored Decimal text remains readable
    through all existing summary/transaction paths.
12. The parser rejects an oversized lexical input before big-integer conversion. The test must make
    the bound observable without relying on wall-clock timing alone.
13. No test, error, trace, or audit entry records raw rejected financial payloads beyond the existing
    safe error contract.

## 7. Alternatives rejected

| Alternative | Rejection rationale |
| --- | --- |
| Broaden OpenAPI to allow all current parser spellings | Weakens the published API and preserves an unbounded leading-zero admission path. |
| Normalize trim/plus/leading-zero forms before parsing | Hides client mistakes, changes signed lexical identity, and keeps contract drift. |
| Require exactly eight fractional digits for inputs | Breaks already conforming forms such as `1` and `1.2` without a correctness benefit. |
| Let PostgreSQL reject malformed or oversized values | Validation would occur too late and may become a generic persistence error. |
| Permit JSON numeric values | Reintroduces binary floating-point ambiguity and violates the frozen Decimal contract. |
| Change Decimal arithmetic while fixing lexical admission | Expands a narrow P3 remediation into unreviewed financial-calculation work. |
| Add a permissive CSV exception | Creates another client/runtime contract mismatch at the ledger boundary. |

## 8. Adversarial review requirements

The implementation review must challenge:

- a regex or validator that matches a valid substring rather than the complete string;
- divergence between JSON Schema/OpenAPI regex behavior and Go validation, especially anchors and
  final line terminators;
- acceptance of leading plus, leading zeroes, blank fraction, Unicode digits, or excessively long
  strings by any HTTP path or by CSV after its existing field-edge whitespace normalization;
- an over-correction that rejects documented valid forms or changes `-0` normalization;
- hidden float conversion, rounding-policy change, schema/migration change, or changed persisted
  financial values;
- whether an input limit is imposed before big-integer allocation/conversion;
- error mapping that converts client validation into 500 or leaks raw financial values; and
- accidental scope absorption from P3-04 Unicode/maxLength, P3-05 cleanup, P3-06 HTTP decomposition,
  or the separate Stage 3.25 privacy work.

## 9. Operational, data, and security consequences

No migration, backfill, reindex, worker, scheduler, provider integration, credential change, or
external data retrieval is planned. No production data is opened or modified by this planning stage.

The remediation narrows non-conforming request admission and bounds parser work. It does not claim a
privacy-lifecycle outcome, change retention, or close any finding other than P3-03 after a separately
reviewed implementation and closure process.

## 10. Exact evidence and closure rule

This document authorizes no runtime change. P3-03 remains OPEN until all of the following occur:

1. this planning gate is independently reviewed and canonically merged;
2. a separately scoped implementation passes its parser, HTTP, importer, OpenAPI, PostgreSQL, and
   full exact-head CI evidence;
3. a fresh independent review reaches `APPROVED` after every implementation change;
4. the user explicitly authorizes the canonical merge; and
5. separately reviewed closure governance records the exact head, CI, review, merge, and remaining
   audit state.

Until then the original audit state remains P0=0, P1=0, P2=0, P3=9. Stage 3.25 privacy evidence
planning remains active and separate.

## 11. Explicit non-scope

No implementation begins under this planning document. This stage does not authorize changes to
financial arithmetic, rounding, database schemas, migrations, stored ledger/history, snapshots,
import identity, idempotency, session cleanup, account deletion/anonymization, audit retention,
OpenAPI endpoints, frontend product behavior, global Unicode/maxLength policy, HTTP decomposition,
dependencies, infrastructure, tax, AI, mobile, broker integrations, or any P3 finding other than
P3-03.
