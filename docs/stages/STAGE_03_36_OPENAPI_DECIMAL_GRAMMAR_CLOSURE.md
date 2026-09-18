# Stage 3.36 — P3-03 OpenAPI Decimal Grammar Closure

Canonical record: PR #87, PR #88; commit(s) `251296e0831cbb0b81c7799cc82cbdf3b451ae6e`, `131f1bf963e9d232b9e23273edd54caf54c10ffb`, `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa`.

## 1. Finding / symptom

P3-03 identified a contract/runtime discrepancy in the public financial Decimal language. The
published OpenAPI schema already required the complete ASCII grammar
`^-?(0|[1-9][0-9]{0,19})(\.[0-9]{1,8})?$`, while the Go parser accepted additional spellings such
as a leading plus, leading zeroes, surrounding whitespace, an empty fractional part, and an
unbounded leading-zero prefix.

The finding was not a binary-floating-point defect and no stored financial corruption was shown.
The risk was inconsistent request admission across generated clients, direct JSON callers, CSV
normalization, and the Go runtime, plus unnecessarily unbounded lexical work before fixed-scale
conversion.

## 2. Root cause

`backend-go/internal/decimal.FromString` historically normalized a broader input language than the
OpenAPI contract. It trimmed whitespace, admitted `+`, tolerated non-canonical leading zeroes and
`1.`, and measured precision after removing leading zeroes. The OpenAPI contract was stricter and
already represented the intended financial boundary.

## 3. Failure scenario

A caller could submit a Decimal such as `+1`, `001.25`, ` 1.25 `, `1.`, or an arbitrarily long
leading-zero spelling. The runtime could accept or normalize it even though an OpenAPI-conforming
client would reject it. For imports, changing parser semantics also risked invalidating the exact
read-only replay of a command that had completed under parser version 1.

## 4. Impact

The mismatch weakened API predictability and made financial input acceptance depend on the entry
path. It also allowed bounded numeric values to carry unbounded lexical prefixes. The impact stayed
P3 because fixed-scale Decimal arithmetic and PostgreSQL `NUMERIC(28,8)` storage remained exact and
no monetary corruption was demonstrated.

## 5. Severity rationale

P3 is retained. The defect affected contract parity, validation determinism, and parser resource
bounds, not the numerical value model, financial arithmetic, database precision, or persisted ledger
integrity.

## 6. Existing guarantees violated

The Stage 2 financial contract requires Decimal values to be base-10 JSON strings, forbids binary
floating point, limits public lexical capacity to PostgreSQL `NUMERIC(28,8)`, and expects OpenAPI and
runtime admission to describe one reproducible request language. P3-03 violated the last guarantee.

## 7. Considered solutions


## 8. Chosen remediation

PR #88 implemented the narrow approved remediation:


No migration, Decimal arithmetic change, float conversion, stored-data rewrite, snapshot change, or
other P3 remediation was included.

## 9. Why this solution

The OpenAPI grammar was already the approved public contract. Narrowing the runtime parser preserves
all conforming clients, removes ambiguous lexical normalization, bounds parser work before expensive
conversion, and avoids weakening financial transport semantics. The parser-version transition keeps
fresh-write authorization fail closed while preserving exact replay of immutable completed commands.

## 10. Rejected alternatives


## 11. Trade-offs

Previously tolerated but non-contract request spellings are now rejected. A small compatibility path
exists for reconstructing completed parser-v1 import commands, increasing replay code complexity, but
that path cannot authorize a fresh write and is covered by negative replay tests. Existing persisted
financial values require no rewrite.

## 12. Regression tests and verification

The final runtime candidate includes evidence for:

- accepted canonical Decimal vectors, including precision boundaries and negative zero;
- rejection of plus signs, leading zeroes, whitespace, empty fractions, scientific notation, locale
  separators, Unicode digits, excess fraction digits, 21 integer digits, and oversized lexemes;
- strict HTTP validation before store append;
- CSV review/append behavior with field-edge normalization retained;
- parser-v1 token rejection for fresh writes after the v2 transition;
- exact read-only replay of completed parser-v1 commands, including formerly permissive Decimal
  spellings;
- rejection of bad HMAC, changed principal, source context, raw CSV/file identity, row identity,
  decisions, and canonical path;
- disposable-PostgreSQL transaction/summary/list round-trip after migrations `000001` through
  `000006`;
- literal OpenAPI Decimal-pattern pinning against the runtime acceptance corpus.

The exact final runtime head `131f1bf963e9d232b9e23273edd54caf54c10ffb` passed GitHub Actions
CI #257, run `32822925542`, with all 10 required Stage 3.34 checks successful.

## 13. Adversarial review findings

The runtime candidate required a second hardening commit after the initial implementation commit
`5ef7ea1f523344929938f5d570f791b574a6578c`. The final head
`131f1bf963e9d232b9e23273edd54caf54c10ffb` strengthened contract evidence by pinning the literal
published Decimal pattern rather than relying on Go regular-expression behavior for the parity
oracle, added carriage-return/line-ending adversarial vectors, added canonical-path replay isolation,
and corrected the implementation record.

Canonical record: PR #88.

## 14. Remediation iterations

1. Planning PR #87 established the exact contract, compatibility boundary, required adversarial
   vectors, and non-scope.
2. Runtime commit `5ef7ea1f523344929938f5d570f791b574a6578c` implemented the strict parser,
   parser-version transition, replay compatibility, and regression suite.
3. Runtime hardening commit `131f1bf963e9d232b9e23273edd54caf54c10ffb` corrected the Decimal
   contract test oracle and added further replay/line-ending evidence.
4. PR #88 was squash-merged to `develop` as
   `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa` after exact-head green CI.
5. This closure package synchronizes governance state without changing runtime code.

## 15. Residual risk / limitations

P3-03 does not close or absorb:

- P3-02 IANA timezone semantics;
- P3-04 general Unicode / OpenAPI `maxLength` semantics;
- P3-05 idempotency/session cleanup lifecycle;
- P3-06 `httpapi/api.go` decomposition;
- P3-07 transaction-form fixture/default cleanup;
- P3-08 migration-validator policy hardening;
- P3-09 Next.js maintenance;
- P3-10 Fiber maintenance;

The historic parser-v1 replay path remains only as compatibility code for immutable completed-command
recovery and must not be generalized into fresh-write admission.

## 16. Operational / deployment consequences

No migration, backfill, reindex, worker, scheduler, dependency change, provider change, credential
change, or manual data rewrite is required. Deployment narrows admission of non-contract Decimal
spellings. Completed parser-v1 import commands retain exact authenticated read-only replay; fresh
writes require the current parser/review version.

## 17. Exact evidence

- Planning merge: PR #87 → `251296e0831cbb0b81c7799cc82cbdf3b451ae6e`.
- Initial runtime commit: `5ef7ea1f523344929938f5d570f791b574a6578c`.
- Frozen final runtime head: `131f1bf963e9d232b9e23273edd54caf54c10ffb`.
- Runtime PR: #88, `fix: enforce OpenAPI Decimal grammar`.
- Exact-head CI: CI #257 / Actions run `32822925542`; 10/10 required jobs successful.
- Canonical runtime squash merge: `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa`.
- GitHub-native PR #88 submitted reviews: none returned by the repository API.
- GitHub-native PR #88 discussion comments: none returned by the repository API.
- Closure package: this document plus synchronized `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, and the Stage
  3.36 implementation record.

## 18. Final canonical status rule

At the base of this closure candidate, the canonical repository audit count remains:

- P0: 0
- P1: 0
- P2: 0
- P3: 9

P3-03 runtime remediation is canonical in `develop`, but the finding is not declared canonically
CLOSED by this unmerged branch. P3-03 becomes CLOSED only when this closure package receives a fresh
and explicitly authorizes the closure squash merge, and the closure PR is merged into `develop`.

The resulting post-closure backlog is P0=0 / P1=0 / P2=0 / P3=8, consisting of P3-02, P3-04,
P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 remains separate.
