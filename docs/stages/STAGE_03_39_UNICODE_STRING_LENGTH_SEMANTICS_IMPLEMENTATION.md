# Stage 3.39 — P3-04 Unicode and OpenAPI String-Length Semantics Implementation

| Field | Value |
| --- | --- |
| Status | PUBLISHED DRAFT PR — documentation/evidence remediation after blind external review |
| Date | 2026-08-27 |
| Finding | P3-04 — general Unicode / OpenAPI `minLength` / `maxLength` semantics |
| Canonical runtime base | `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` |
| Planning gate | PR #96 squash-merged into `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` after exact-head CI #273 / run `33018628673` and fresh independent `APPROVED` review |
| Runtime branch | published `fix/stage-03-39-p3-04-runtime` |
| Runtime implementation commit | `740f5c313baac75d0406e9146a7bbc74a94d48c1` — published after approved pre-commit Internal Review |
| Runtime PR / merge | Draft PR #97 is open; Ready/merge are NOT AUTHORIZED and have not been performed |
| Published runtime-head CI | CI #274 / run `33049818896` completed `success` on exact runtime head `740f5c313baac75d0406e9146a7bbc74a94d48c1` with all 10 required jobs green |
| Blind external review | exact published runtime head reviewed independently; verdict `REQUEST CHANGES` for one documentation/governance P3 only, with no blocking runtime technical defect |
| Finding status | OPEN — documentation evidence follow-up, exact-head CI/review, human approval, merge, and closure governance remain |

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

The published runtime implementation applies the merged Stage 3.39 plan without expanding scope:

- bounded public human text uses valid UTF-8 plus Unicode code-point counting;
- fresh raw portfolio and asset-search admission occurs before trim;
- note and source-account-label application defenses use code points and reject malformed internal UTF-8;
- importer CSV note validates the post-spreadsheet-neutralization value as UTF-8 before code-point counting;
- source-account-label trim/token/command/broker/persistence identity remains unchanged;
- `ReviewParserVersion` remains `2`;
- the CSV limit remains exactly `2 * 1024 * 1024` UTF-8 bytes;
- OpenAPI removes CSV character `maxLength` and publishes `x-openinvest-max-utf8-bytes: 2097152`;
- Web validation uses one small explicit well-formed/code-point helper and no longer relies on native UTF-16 `maxlength` for the four affected surfaces;
- no Unicode normalization, grapheme-cluster semantics, case folding, transliteration, dependency, migration, or new service is introduced.

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

CSV transaction-note review keeps the existing spreadsheet-neutralization ordering and now rejects
malformed UTF-8 on the post-neutralization value before counting its 500-code-point limit.

## 7. PostgreSQL / migration decision

No migration is added. Existing PostgreSQL character checks for transaction notes and source account
labels already use character semantics compatible with the approved code-point model. Portfolio name
remains an application write-boundary rule; no active bypass was established by planning review.

## 8. Regression and CI evidence

The published candidate includes focused coverage for portfolio 100/101 ASCII, Cyrillic, and
supplementary-plane boundaries; malformed UTF-8; raw-before-trim admission; exact historical replay
and no second effect; no-authority rejection; asset raw bounds; transaction note 500/501; source-label
120/121 and trim identity; malformed source label; malformed CSV note through the real `ReviewCSV`
path; unchanged parser version; exact 2 MiB CSV byte admission; OpenAPI byte-vs-character publication;
Web code-point/surrogate behavior; and the unchanged browser `File.size` byte guard.

Before publication, the exact candidate also passed local focused Go tests, the full local Go suite,
Go vet, the OpenAPI validator, frontend tests, frontend typecheck, frontend build, `git diff --check`,
and dependency/migration scope checks.

After publication, authoritative GitHub CI #274 / run `33049818896` ran on exact runtime head
`740f5c313baac75d0406e9146a7bbc74a94d48c1` and completed `success` with all 10 required jobs green:
Go tests, Python tests, Frontend build and typecheck, OpenAPI contract, Docker Compose config,
PostgreSQL migration validation, Go vet, Go race tests, Go vulnerability scan, and Dependency security scan.

Any later evidence-only documentation commit advances the PR head and therefore requires fresh
exact-head CI before Ready/merge consideration.

## 9. Scope

The runtime implementation does not include P3-06, P3-07, P3-08, P3-09, P3-10, Stage 3.25 privacy
work, authentication redesign, database migration, dependency upgrade, generic import replay redesign,
Unicode normalization, or future update/reversal endpoint implementation.

The external-review remediation recorded in this document is documentation/evidence-only and does not
change runtime code, OpenAPI behavior, migrations, dependencies, replay authority, parser version, or financial semantics.

## 10. Internal Review Evidence

Per `docs/REVIEW_WORKFLOW.md`, this evidence was intentionally withheld from the Draft PR during the
blind external review and is published only after the independent external verdict was recorded.

The mandatory independent pre-commit Internal Review Agent:

- reviewed the complete revised 17/17-file candidate line-by-line;
- reviewed complete pre-commit diff SHA256 `4bff9f65393ecfbe45868bc6f7dabaf12f8e6b300e6978d76746f7ffcce250c8`;
- verified canonical base `32b198ee9d349f119ed374fd86d47622e27bcd73`;
- returned final verdict **`APPROVED`**;
- reported **P0: none, P1: none, P2: none, P3: none** on the revised candidate;
- confirmed the earlier malformed importer CSV-note gap was resolved before approval;
- confirmed all 17 changed files were reviewed;
- confirmed the reviewer made no edits, commits, pushes, PR mutations, Ready transition, or merge.

The approved runtime candidate was subsequently published as runtime implementation commit
`740f5c313baac75d0406e9146a7bbc74a94d48c1`.

## 11. Blind External Review and Evidence-Only Remediation

The independent External Review Agent reviewed Draft PR #97 on exact published runtime head
`740f5c313baac75d0406e9146a7bbc74a94d48c1`, verified exact-head CI #274, reviewed all 17/17 changed
files, and did not use Internal Review evidence before reaching its verdict.

The external review found:

- no P0;
- no P1;
- no P2;
- no blocking runtime technical defect;
- one blocking documentation/governance P3, `EXT-PR97-P3-01`, because this Stage report still described
  the implementation as local-only/uncommitted/unpushed/no-PR and the Internal Review as pending.

External verdict on that runtime head: **`REQUEST CHANGES`**.

This follow-up documentation remediation addresses `EXT-PR97-P3-01` by recording publication-stable
facts and publishing the previously withheld Internal Review evidence after the blind verdict, exactly
as required by `docs/REVIEW_WORKFLOW.md`.

After this evidence-only documentation change is published, the resulting new PR head requires fresh
exact-head CI and both the Internal Review Agent and External Review Agent must verify the evidence-only
change. Ready/merge remain human-authorized gates only.

P3-04 remains **OPEN** until the remaining governed publication/review/merge/closure sequence completes.
