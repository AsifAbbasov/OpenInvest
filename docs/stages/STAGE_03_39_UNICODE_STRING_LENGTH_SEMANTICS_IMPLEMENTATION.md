# Stage 3.39 — P3-04 Unicode and OpenAPI String-Length Semantics Implementation

| Field | Value |
| --- | --- |
| Status | LOCAL RUNTIME CANDIDATE — not committed, not pushed, no PR |
| Date | 2026-08-27 |
| Finding | P3-04 — general Unicode / OpenAPI `minLength` / `maxLength` semantics |
| Canonical runtime base | `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` |
| Planning gate | PR #96 squash-merged into `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` after exact-head CI #273 / run `33018628673` and fresh independent `APPROVED` review |
| Runtime branch | local-only `fix/stage-03-39-p3-04-runtime` |
| Runtime commit / push | NOT AUTHORIZED / not performed |
| Runtime PR / merge | NOT AUTHORIZED / not performed |
| Finding status | OPEN — runtime and later closure governance are still required |

## 1. Problem

P3-04 tracks cross-layer disagreement over what a bounded human-readable string length means.
OpenAPI 3.1 / JSON Schema 2020-12 publishes character bounds, Go historically mixed
`len(string)` byte counts with rune counts, the browser used UTF-16 `maxlength`, and the import CSV
2 MiB resource limit was incorrectly represented as JSON Schema `maxLength`.

## 2. Root cause

The repository had no single explicit length unit for general human text. This made ASCII behavior look
correct while Cyrillic and supplementary-plane Unicode could be rejected by Go byte counts or browser
UTF-16 code-unit counts before reaching the canonical API rule.

Portfolio create also had a deployment-compatibility wrinkle: pre-Stage-3.39 runtime trimmed the name
before its 100-byte check and before request hashing. Moving fresh admission to raw code-point
validation can therefore make a previously successful raw request newly invalid on retry.

## 3. Failure scenarios

1. A 100-code-point Cyrillic portfolio name can exceed 100 UTF-8 bytes and be rejected despite
   satisfying the OpenAPI contract.
2. A 100-emoji value occupies 200 UTF-16 code units and can be blocked by native browser `maxlength`.
3. A raw portfolio name longer than 100 code points can become 100 after trimming and bypass a raw
   contract limit.
4. A pre-Stage-3.39 completed portfolio command using that historic trim-first identity can lose exact
   replay before its Stage 3.38 authority expires.
5. `csvPayload.maxLength: 2097152` describes character count while runtime intentionally enforces
   2 MiB UTF-8 bytes.

## 4. Chosen implementation

The candidate applies the merged Stage 3.39 plan without expanding scope:

- bounded public human text uses valid UTF-8 plus Unicode code-point counting;
- fresh raw portfolio and asset-search admission occurs before trim;
- note and source-account-label application defenses use code points and reject malformed internal
  UTF-8;
- source-account-label trim/token/command/broker/persistence identity remains unchanged;
- the import parser version remains unchanged;
- the CSV limit remains exactly `2 * 1024 * 1024` UTF-8 bytes;
- OpenAPI removes CSV character `maxLength` and publishes
  `x-openinvest-max-utf8-bytes: 2097152`;
- Web validation uses one small explicit well-formed/code-point helper and no longer relies on native
  UTF-16 `maxlength` for the four affected surfaces;
- no Unicode normalization, grapheme-cluster semantics, case folding, transliteration, dependency,
  migration, or new service is introduced.

## 5. Portfolio cross-version replay

Only a raw portfolio name that is newly over the Stage 3.39 100-code-point limit and could have passed
the exact historic trim-first 100-byte rule is eligible for compatibility lookup.

The service validates the idempotency key, rejects malformed UTF-8, reconstructs the historic trimmed
request only for that newly-over-limit case, calls the existing `ReplayLookupStore` read-only lookup,
returns an exact unexpired completed artifact when found, propagates fail-closed lookup errors, and
falls back to fresh Stage 3.39 rejection when no authority exists.

The compatibility path never reserves, reclaims, writes, creates a generation, extends TTL, changes
request-hash format, or invokes the replay response builder.

`backend-go/internal/postgres/replay_lookup.go` is intentionally unchanged because its generic lookup
already implements no-row/expiry as no authority, hash conflict, in-flight and unsupported fail-closed
states, artifact-integrity checks, and exact response recovery without writes.

## 6. Import compatibility boundary

No source-label historical replay exception is added. The old public HTTP guard checked raw
`len(sourceAccountLabel) <= 120` before importer trim. For valid UTF-8, code-point count cannot exceed
byte count, so moving the public guard to `<=120` code points expands multilingual admission but does
not create a previously admitted population that becomes invalid.

`ReviewParserVersion` remains `2`. `import_replay_recovery.go` and import replay ordering are untouched.

## 7. PostgreSQL / migration decision

No migration is added. Existing PostgreSQL character checks for transaction notes and source account
labels already use character semantics compatible with the approved code-point model. Portfolio name
remains an application write-boundary rule; no active bypass was established by planning review.

## 8. Regression proof carried by this candidate

Focused tests cover portfolio 100/101 ASCII, Cyrillic, and supplementary-plane boundaries; malformed
UTF-8; raw-before-trim admission; exact historical replay and no second effect; no-authority rejection;
asset raw bounds; note 500/501; source-label 120/121 and trim identity; unchanged parser version; exact
2 MiB CSV byte admission; OpenAPI byte-vs-character publication; Web code-point/surrogate behavior; and
the unchanged browser `File.size` byte guard.

The one-click package also runs focused Go tests, the full local Go suite, Go vet, the OpenAPI
validator, frontend tests, frontend typecheck, and frontend build. PostgreSQL-backed authoritative CI
remains a later published-head gate and is not claimed by this local candidate.

## 9. Scope

This candidate does not include P3-06, P3-07, P3-08, P3-09, P3-10, Stage 3.25 privacy work,
authentication redesign, database migration, dependency upgrade, generic import replay redesign,
Unicode normalization, or future update/reversal endpoint implementation.

## 10. Internal review evidence

Mandatory independent pre-commit Internal Review Agent verdict is **PENDING**.

The Builder must not commit or push until local evidence is complete, the exact full diff is reviewed
line-by-line, the Internal Review Agent returns `APPROVED`, and the human separately authorizes
feature-branch commit/push.

P3-04 remains **OPEN**.
