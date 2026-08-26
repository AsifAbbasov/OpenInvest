# Stage 3.39 — P3-04 Unicode and OpenAPI String-Length Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning/review candidate only; runtime implementation not authorized |
| Date | 2026-08-26 |
| Canonical planning base | `develop` at `c5962fa09b6d7d145dda203dbdf90069de7b1fcc` |
| Finding | P3-04 — general Unicode / OpenAPI `minLength` / `maxLength` semantics |
| Prior closure | Stage 3.38 / P3-05 closure actually merged through PR #95 at `c5962fa09b6d7d145dda203dbdf90069de7b1fcc` |
| Runtime implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / merge authorized here | No |

## 1. Objective

Freeze one deterministic length model for the **implemented** bounded-string surfaces covered by P3-04
without absorbing P3-01 password semantics, P3-02 timezone semantics, P3-06 `httpapi/api.go`
decomposition, P3-07 fixture/default work, P3-08 migration-validator policy, P3-09/P3-10 dependency
maintenance, or Stage 3.25 privacy evidence planning.

This planning gate does not change runtime behavior and does not close P3-04.

The intended correction is contract parity, not Unicode normalization. Public OpenAPI string-length
keywords, Go admission, PostgreSQL character checks, and Web presentation validation must stop using
different units for the same field.

## 2. Normative authority and length units

The repository publishes OpenAPI 3.1.0 with JSON Schema Draft 2020-12 as its schema dialect. Under that
contract, ordinary JSON Schema string length is character length over Unicode code points.

The relevant platform facts are intentionally not treated as interchangeable:

- JSON Schema `minLength` / `maxLength`: Unicode code-point character count.
- Go `len(string)`: UTF-8 byte count.
- Go `utf8.RuneCountInString`: Unicode code-point count for valid UTF-8 strings.
- JavaScript `String.length`: UTF-16 code-unit count.
- JavaScript `Array.from(value).length`: Unicode code-point iteration for well-formed strings.
- PostgreSQL `length(text)` / `char_length(text)`: character count, not UTF-8 byte count.

Therefore a native Web `maxlength`, Go `len(string)`, and OpenAPI `maxLength` are not equivalent for
general Unicode text.

No grapheme-cluster policy is introduced. No NFC, NFD, NFKC, case folding, transliteration, or other
normalization is authorized.

## 3. Current cross-layer inventory

| Surface | Published contract | Current Go / importer behavior | Current PostgreSQL behavior | Current Web behavior | P3-04 disposition |
| --- | --- | --- | --- | --- | --- |
| Asset search `query` | `1..100` via OpenAPI | Handler trims; service uses `utf8.RuneCountInString` after trim | Not persisted | Native `maxLength={100}` counts UTF-16 code units; submit path trims | Web mismatch; transport raw-bound admission also needs exact parity |
| Portfolio `name` create | `1..100` via OpenAPI | Both create paths trim and use `len(name) > 100` | `TEXT`; nonblank CHECK only | Native `maxLength={100}` then `trim()` before submit | Confirmed byte/code-unit drift |
| Transaction `note` | `maxLength: 500` | Service uses rune count; importer uses rune count after safe-note transformation | `length(note) <= 500` | Native `maxLength={500}` then trimmed payload | Backend/database aligned; Web supplementary-plane mismatch |
| Import `sourceAccountLabel` | `maxLength: 120` | HTTP DTO uses `len`; importer trims; review-history service uses rune count; append-batch service uses `len` | `char_length(source_account_label) <= 120` | Native `maxLength={120}` then trimmed payload | Confirmed transport/service/Web drift |
| Import `csvPayload` | currently `maxLength: 2097152` | HTTP intentionally enforces `<= 2 MiB` **UTF-8 bytes** | Raw payload not persisted | `File.size` enforces `<= 2 MiB` bytes before `File.text()` | OpenAPI keyword expresses wrong unit; runtime byte budget must remain bytes |
| Password | Stage 3.35 explicit code-point/exact-secret contract | Closed P3-01 implementation | Password hash only | Explicit password policy helper | Out of P3-04 |
| Registration timezone | Stage 3.37 resolver-based IANA contract | Closed P3-02 implementation | `TEXT` preference | Existing free-text field | Out of P3-04 |
| Email | OpenAPI `format: email`, max 254 | Existing mail parser + byte bound | Identity storage | Browser email field | No P3-04 change; conforming current mailbox contract is ASCII-domain |
| Tickers, hashes, idempotency keys, row hashes, trace IDs | ASCII patterns / fixed encodings | Byte length is equivalent on admitted domain | ASCII/fixed technical identity | Generated or ASCII-restricted | Out of P3-04 |
| Update/reverse reasons and other future DTOs | Bounded OpenAPI strings | No active router implementation for the corresponding future mutation flow | N/A | N/A | Do not implement future functionality as part of P3-04 |

The inventory is intentionally implementation-aware. An OpenAPI schema existing for a future endpoint
does not authorize implementation of that endpoint.

## 4. Root cause

P3-04 is not one bad `len()` call. The repository historically used the word “characters” while
allowing each layer to select its native length primitive.

That created four distinct failure modes:

1. Go byte counting can reject valid multilingual input below the OpenAPI code-point ceiling.
2. Browser `maxlength` can reject supplementary-plane characters earlier than the API contract.
3. Application trimming before validation can cause a raw request longer than the published
   `maxLength` to be accepted after normalization.
4. `csvPayload` uses an intentional byte resource budget but publishes it through a character-count
   JSON Schema keyword.

The source-account-label path is especially sensitive because the importer trims the label and the
normalized value participates in review history, signed review-token context, command hashing, broker
identity scope, and persistence. P3-04 must not change that identity model while fixing the length unit.

## 5. Failure scenarios

Representative current defects include:

- 100 Cyrillic portfolio-name code points satisfy OpenAPI but exceed 100 UTF-8 bytes and are rejected
  by the current Go create path.
- A supplementary-plane character consumes two JavaScript UTF-16 code units, so a native
  `maxLength={100}` can block a 100-code-point value long before the OpenAPI ceiling.
- A 120-code-point Cyrillic import source label is permitted by OpenAPI and PostgreSQL
  `char_length(...) <= 120` but is rejected by the current HTTP `len(...) > 120` guard.
- A raw portfolio name exceeding 100 code points can become <=100 only after trimming; validating only
  the trimmed value does not implement raw OpenAPI `maxLength`.
- A multi-byte `csvPayload` can contain fewer than 2,097,152 code points but exceed the intended 2 MiB
  UTF-8 resource budget. OpenAPI currently describes the wrong constraint even though runtime correctly
  measures bytes.

No financial arithmetic corruption is claimed from these scenarios.

## 6. Severity rationale

P3 remains appropriate. The defect is real cross-layer contract drift and creates multilingual UX and
validation inconsistency, but no demonstrated authentication bypass, cross-account access, canonical
financial corruption, migration destruction, or systemic availability failure follows from it.

The import label must nevertheless be handled conservatively because it participates in previously
hardened import identity and review-token guarantees.

## 7. Frozen general text contract

For the P3-04 surfaces listed as in scope:

1. Public JSON/OpenAPI `minLength` / `maxLength` semantics are Unicode code points.
2. Go service/application admission for general text must use valid UTF-8 plus code-point counting,
   not byte length.
3. Web presentation code must not use native `maxlength` as the authoritative Unicode limit for these
   fields. It must validate the exact value it will send by well-formed-Unicode checking plus explicit
   code-point counting.
4. Existing field-specific trimming behavior is preserved. P3-04 does not introduce normalization.
5. Where OpenAPI applies `maxLength` to the submitted string, **fresh command/request admission** must
   reject a raw value above that code-point bound even if later trimming could reduce it.
6. That new fresh-admission rule must not invalidate an already-completed, still-unexpired exact
   idempotency command **only where repository evidence proves that a value accepted before Stage 3.39
   can become newly invalid after Stage 3.39**. The confirmed case is replay-aware portfolio creation,
   because the old runtime trimmed the name before validation and command hashing. For that path runtime
   implementation must reconstruct the historical normalized command identity and perform a
   **read-only exact replay lookup before fresh-admission rejection**. A matching completed artifact may
   be replayed; the compatibility path must never reserve/reclaim a command, authorize a fresh write,
   or extend the 24-hour Stage 3.38 authority window. Existing same-key
   conflict/in-flight/unsupported/corrupt semantics remain authoritative when the read-only lookup
   observes those states.
7. If that read-only lookup finds no authoritative portfolio artifact because no command exists or the
   old generation is expired, normal Stage 3.39 fresh admission applies and an over-limit raw portfolio
   name is rejected. Expiry therefore never grants a historical over-limit request new write authority.
   This exception is not generalized to another field without concrete historical-admission evidence.
8. A persistence-bound normalized value must also remain inside the same semantic bound where a
   database/application invariant already requires that bound.
9. Canonically equivalent but code-point-distinct Unicode strings remain distinct values unless an
   already-established field rule says otherwise.
10. P3-04 does not redefine password identity, timezone identity, ASCII technical identifiers, or
    financial Decimal grammar.

## 8. Exact field contracts

### 8.1 Portfolio name

Fresh public create admission:

- decoded name must be valid Unicode;
- raw submitted name must contain at most 100 Unicode code points;
- existing trimming is then applied;
- trimmed name must be non-empty;
- trimmed persisted name must remain at most 100 Unicode code points;
- no Unicode normalization or case folding occurs.

The normal non-replay create path applies that rule directly. The replay-aware create path has one
additional **cross-version compatibility gate** because pre-Stage-3.39 runtime trimmed the name before
building `CommandContext.RequestHash` and before entering `CreatePortfolioWithReplay`. Therefore a raw
name that is now over 100 code points, but that the previous runtime accepted only after trimming, may
correspond to an already-completed command inside the 24-hour Stage 3.38 authority window.

For such a replay-aware request, and only before rejecting the new raw bound:

1. preserve the raw name for the new Stage 3.39 admission decision;
2. reconstruct the historical command identity using the **same trim-first normalized request** that the
   pre-Stage-3.39 runtime hashed;
3. perform the existing generic replay lookup as a **read-only** operation;
4. if the same principal/method/path/key and historical normalized request hash resolve to a valid,
   unexpired completed artifact, return that exact stored status/body/request-id/trace-id;
5. if the authoritative row has a different request hash, preserve the existing idempotency conflict;
6. if the row is unexpired but in-flight/unsupported/corrupt, preserve the existing fail-closed duplicate
   semantics rather than converting it into a fresh write; and
7. if no authoritative artifact exists or the old generation is expired, apply normal fresh admission,
   so raw `>100` is `400 VALIDATION_ERROR` and cannot create a new portfolio.

This compatibility path may require a narrow portfolio-specific service lookup helper that reuses the
existing `ReplayLookupStore`; it does **not** authorize an idempotency redesign or a write-capable bypass.
It must never reserve/reclaim a key, create a portfolio, extend expiry, or make an expired historical
request admissible.

No PostgreSQL migration is planned. The current portfolio table remains `TEXT` with its nonblank CHECK;
the application write boundary remains authoritative for the 100-code-point public limit.

### 8.2 Asset search query

Public query admission:

- raw query must be valid Unicode;
- raw query must contain at most 100 Unicode code points;
- existing trimming is then applied;
- trimmed query must be non-empty;
- the normalized trimmed query remains the value used for search and cursor query-hash identity.

The existing service rune-count guard remains defense in depth and must add valid-UTF-8 protection for
non-HTTP/internal callers.

No catalog/persistence semantics change is authorized.

### 8.3 Transaction note

For a non-null note:

- valid UTF-8 is required at the Go application boundary;
- the value admitted by the API may contain at most 500 Unicode code points;
- the backend does not gain a new trim/normalization rule;
- the existing PostgreSQL `length(note) <= 500` remains defense in depth;
- importer safe-note validation continues to count the **post-neutralization value that would be
  persisted**.

The current Web form may continue sending its existing trimmed note, but it must validate that exact
outgoing value explicitly by code points instead of relying on native UTF-16 `maxlength`.

### 8.4 Import source-account label

Public review/append admission:

- the raw decoded label must be valid Unicode;
- the raw submitted label may contain at most 120 Unicode code points;
- importer normalization remains `strings.TrimSpace` exactly as today;
- the normalized label is the identity used by parser review, review-history lookup, signed review
  token, append context, command hashing, broker-operation identity scope, and persistence;
- no case folding or Unicode normalization is introduced;
- the normalized label must still satisfy the 120-code-point defense-in-depth check before canonical
  append.

The implementation must not bump `ReviewParserVersion` merely for this correction because accepted
conforming parser row semantics and normalized source-label identity do not change. If implementation
review proves that a parser-version change is actually required, that is a blocker requiring a revised
plan rather than an unreviewed version bump.

No Stage 3.39 import-replay compatibility branch is planned for this length correction. The current
public review/append HTTP boundary already rejects `sourceAccountLabel` when
`len(request.SourceAccountLabel) > 120` **before** importer trimming. For valid UTF-8,
Unicode code-point count is never greater than UTF-8 byte count. Therefore every label accepted by the
pre-Stage-3.39 public HTTP path already had at most 120 Unicode code points. Changing the public rule
from bytes to code points can admit additional multibyte labels, but it cannot make a previously
accepted public label newly invalid.

The existing Stage 3.32 read-only import replay recovery, token verification, parser-version semantics,
decision binding, normalized-label identity, source-file hash, and command-hash behavior must remain
unchanged and green. `import_replay_recovery.go` or import-handler replay ordering must not be changed
for P3-04 unless implementation review first proves a concrete supported historical caller that could
have created a completed artifact outside the public HTTP admission rule. Such evidence would require a
plan revision before code changes.

### 8.5 Import CSV payload

The existing 2 MiB limit is a **UTF-8 byte/resource bound**, not a human character limit.

Therefore:

- `maxHTTPImportPayloadBytes = 2 * 1024 * 1024` remains authoritative at runtime;
- the Web `File.size` byte check remains aligned with that resource bound;
- OpenAPI must stop representing this byte limit as `maxLength: 2097152`;
- the schema should publish the byte rule in description text and an OpenAPI extension such as
  `x-openinvest-max-utf8-bytes: 2097152`;
- focused contract tests must ensure future edits do not reintroduce a character-count keyword as a
  substitute for the byte budget.

P3-04 does not change raw-file retention or authorize a new upload protocol.

## 9. Transport and malformed-Unicode boundary

P3-04 is a **length-unit parity** remediation. It does not authorize a repository-wide replacement of
Go `encoding/json` or a global transport rewrite.

Stage 3.35 already added password-specific lossless Unicode decoding because exact credential identity
made permissive replacement materially security-sensitive. That closed P3-01 and must not be rewritten
as part of P3-04.

For P3-04 general text:

- Go application/importer entry points must fail closed on malformed internal UTF-8 strings before
  code-point counting.
- Web helpers must reject ill-formed JavaScript UTF-16 strings with unpaired surrogates before sending
  affected user-entered text.
- A generic raw-JSON surrogate/UTF-8 scanner is **not** authorized by this plan because it would change
  all strict-JSON routes, including idempotent financial writes and authentication, and could alter
  replay/compatibility semantics beyond the original P3-04 finding.
- If implementation review establishes that raw transport replacement is required to close P3-04,
  that is a planning blocker and this document must be revised before runtime mutation.

This boundary is deliberate scope control, not a claim that lossy decoding is desirable.

## 10. Web implementation boundary

The future runtime change may add one small presentation-only Unicode helper outside the auth feature,
for example under `frontend-next/src/common`, with:

- a well-formed UTF-16/surrogate-pair check; and
- explicit code-point count based on Unicode iteration.

It may be used by the currently implemented UI surfaces:

- `CreatePortfolioForm`;
- `AssetDiscoverySlice`;
- `AddTransactionForm`;
- `ImportUploadReviewPanel`.

Each component must validate the exact normalized value it sends. Native `maxLength` must not remain the
only or stricter authoritative rule for supplementary-plane input.

The closed Stage 3.35 password helper remains separately governed and need not be refactored merely to
deduplicate a few lines.

## 11. Backend implementation boundary

After an approved and merged planning gate, the runtime remediation may be limited to:

- `backend-go/internal/verticalslice/service.go` for portfolio, search, note, and import-batch
  application validation plus the narrow replay-aware portfolio admission ordering required above;
- `backend-go/internal/verticalslice/replay_lookup.go` only if needed to expose a portfolio-specific
  read-only lookup helper over the already-existing `ReplayLookupStore`;
- `backend-go/internal/postgres/replay_lookup.go` only if implementation evidence shows the existing
  generic read-only lookup cannot be reused unchanged; no write semantics may be added there;
- `backend-go/internal/importer/importer.go` for source-account-label Unicode/code-point defense in
  depth while preserving the existing trim/identity normalization;
- the narrow portions of `backend-go/internal/httpapi/api.go` needed to replace the public
  source-label byte guard with valid-Unicode/code-point admission and enforce raw asset-query/OpenAPI
  bounds without changing route ownership;
- focused Go tests in the existing verticalslice/importer/httpapi test packages;
- `openapi/openapi.yaml` and `openapi/components/schemas.yaml` only where contract text/byte-budget
  expression requires correction;
- focused OpenAPI validator tests;
- the four affected Web components plus one small common Unicode helper and focused component tests;
- Stage 3.39 implementation/closure evidence required by governance.

No `httpapi/api.go` decomposition enters this work. P3-06 stays separate.

## 12. PostgreSQL and migration boundary

No PostgreSQL schema change or migration is planned.

Existing persistence evidence is intentionally reused:

- transaction note: `length(note) <= 500`;
- import source-account label: `char_length(source_account_label) <= 120`;
- portfolio name: existing nonblank `TEXT` CHECK, with the public 100-code-point rule enforced by the
  application write boundary.

Adding a new portfolio-name migration is rejected for this P3 unless implementation evidence proves an
active canonical write path can bypass application validation. No such bypass is currently established.

## 13. Required regression matrix

### Go service/importer

1. Portfolio name: 100 ASCII accepted; 100 Cyrillic accepted; 100 supplementary-plane code points
   accepted; 101 code points rejected for **fresh** admission.
2. Fresh portfolio name raw value above 100 code points is rejected even if trimming would reduce it to
   100.
3. Portfolio name with malformed internal UTF-8 is rejected.
4. Cross-version portfolio replay: seed a pre-Stage-3.39 completed command whose original raw name was
   101 code points but whose trim-first normalized request was 100; the same principal/method/path/key
   and historical normalized request after Stage 3.39 returns the exact stored status/body and produces
   no second business effect.
5. The same historical raw 101 request with a **fresh key** receives `400 VALIDATION_ERROR`.
6. The same raw 101 request after the old command has expired receives `400 VALIDATION_ERROR` and cannot
   reclaim/create a new generation.
7. Same key plus a different historical normalized request hash remains idempotency conflict; unexpired
   in-flight/unsupported/corrupt duplicate states remain fail-closed.
8. Asset query: 100 supplementary-plane code points accepted; 101 rejected; whitespace-only remains
   rejected; normalized cursor/search identity remains trimmed as today.
9. Transaction note: 500 supplementary-plane code points accepted; 501 rejected; malformed internal
   UTF-8 rejected.
10. Import source label: 120 Cyrillic and 120 supplementary-plane code points accepted through review;
    121 rejected for fresh admission.
11. Source-label trimming produces the same signed/persisted identity as before; review and append with
    the same normalized label still verify.
12. Fresh source-label variants that exceed the raw public bound are rejected before they can authorize
    a new review/append as a shorter normalized public value.
13. The existing Stage 3.32 import token-expiry/parser-version/read-only replay recovery suites remain
    unchanged and green; P3-04 adds no source-label-specific replay branch or replay-ordering change.
14. Import append defense in depth uses code points, not bytes, while the normalized label used for
    review/token/command/broker/persistence identity remains exactly the existing trimmed value.
15. Repository evidence or focused tests demonstrate the compatibility premise explicitly: every
    valid-UTF-8 `sourceAccountLabel` admitted by the old public `len(...) <= 120` HTTP guard necessarily
    had `<=120` Unicode code points, so no pre-Stage-3.39 public newly-invalid replay population exists.
16. P3-04 may change **portfolio validation/replay ordering only for the confirmed Stage 3.32/3.38
    deployment-compatibility case**; it must not weaken or broaden idempotency authority, TTL,
    request-hash, conflict, or business-effect semantics, and it must not reorder import replay for an
    unsupported synthetic scenario.

### HTTP/OpenAPI

17. Portfolio Unicode boundary failures map to the established 400 validation path when no
    authoritative historical portfolio replay outcome supersedes fresh admission; source-label Unicode
    boundary failures use the normal 400 validation path with no new replay exception.
18. Asset search enforces the raw 100-code-point public bound before existing trim semantics.
19. OpenAPI retains the human-text code-point bounds.
20. `csvPayload` no longer uses JSON Schema `maxLength` to express the 2 MiB byte resource limit.
21. Both import review and append publish the same UTF-8 byte extension/description.
22. Runtime rejects a payload over 2 MiB UTF-8 bytes even when its Unicode code-point count is below
    2,097,152.
23. Runtime accepts an otherwise-valid payload at the byte boundary when all other import limits pass.
24. The full Stage 3.32 exact replay suite and Stage 3.38 expiry/reclamation suite remain green, including
    replay-before-current-state ordering and post-expiry zero old-generation authority.

### Web

25. 100 supplementary-plane portfolio-name code points are not blocked merely because JavaScript uses
    200 UTF-16 code units; 101 are rejected before request submission.
26. Asset query, transaction note, and source-account label use the same explicit code-point helper.
27. Unpaired surrogate input is rejected client-side for these affected text fields.
28. The import file size check remains byte-based and unchanged.
29. The exact trimmed values currently sent by the UI remain the values validated and transmitted.
30. Existing auth/password tests remain unchanged and green.

### PostgreSQL/integration

31. Existing transaction-note and source-label character constraints continue to accept multibyte values
    at their code-point boundary.
32. No migration is added and migration validation remains green.

## 14. Alternatives rejected

| Alternative | Rejection rationale |
| --- | --- |
| Redefine all public limits as UTF-8 bytes | Contradicts OpenAPI/JSON Schema string-length semantics and existing PostgreSQL character constraints. |
| Keep native HTML `maxlength` as canonical | Counts UTF-16 code units and rejects supplementary-plane text too early. |
| Use grapheme clusters | Adds segmentation policy and dependency complexity not required by the original finding. |
| Normalize Unicode | Changes value identity and can break hashes/review context without any product requirement. |
| Trim before enforcing the public raw `maxLength` | Allows an instance that violates the published raw schema to be accepted after transformation. |
| Convert the 2 MiB CSV limit to code points | Weakens the resource bound and misstates actual memory/request cost. |
| Express CSV bytes with JSON Schema `maxLength` | `maxLength` counts string characters, not UTF-8 bytes. |
| Add a generic raw-JSON Unicode scanner now | Broad transport/replay/auth change exceeds the demonstrated P3-04 length finding. |
| Add a portfolio-name database migration now | No active bypass has been established; unnecessary schema/backfill risk for a low-severity contract fix. |
| Refactor `httpapi/api.go` while touching it | Would silently absorb P3-06. |
| Refactor Stage 3.35 password policy into the new Web helper | Unnecessary coupling to a separately closed auth/security finding. |

## 15. Compatibility, security, performance, and cost

Compatibility:

- conforming ASCII input remains unchanged;
- conforming multilingual input up to the published code-point limits becomes consistently accepted;
- for **fresh admission**, raw input that exceeds the published max but would previously become valid
  only after trimming becomes a deterministic validation failure;
- a matching pre-Stage-3.39 completed **portfolio-create** command remains replayable only while its
  existing Stage 3.38 generation is authoritative, using the historical trim-first normalized request
  identity; this is a deployment-compatibility preservation of Stage 3.32 exact replay, not permission
  for a new over-limit write;
- no analogous public source-account-label compatibility branch is introduced because the old raw
  120-byte HTTP guard already implied a maximum of 120 Unicode code points for every accepted valid-UTF8
  request;
- after expiry, or when no matching authoritative completed artifact exists, the new raw bound applies
  normally and the historical over-limit form has zero write authority;
- no persisted data is rewritten;
- no parser-version, review-token version, cursor version, command-hash format, database schema, or
  financial methodology version changes are planned.

Security/privacy:

- no new secret handling, session policy, authorization scope, external provider, retention rule, or
  privacy lifecycle is introduced;
- source-account-label identity binding must remain exact after the existing trim normalization.

Performance/cost:

- code-point counting is linear in bounded small human-text fields;
- the 2 MiB import byte bound is preserved;
- no new service, dependency, worker, queue, cache, database object, or paid infrastructure is added.

## 16. Governance bookkeeping carried by the planning increment

The documentation-only planning increment may synchronize already-established post-PR-95 truth:

- PR #95 is actually merged;
- canonical Stage 3.38 closure merge SHA is
  `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`;
- P3-05 is CLOSED;
- the original 32-finding audit backlog is P0=0 / P1=0 / P2=0 / P3=6;
- the exact remaining original findings are P3-04, P3-06, P3-07, P3-08, P3-09, and P3-10;
- Stage 3.25 remains a separate proposal-only privacy evidence-collection work item;
- P3-04 remains OPEN throughout planning.

The Stage 3.38 historical failed review/remediation chronology must not be rewritten or erased.
However, its **active top-level/current-state metadata** is not historical evidence and must be
synchronized to the already-completed PR #95 merge. The Stage 3.39 planning package therefore carries
narrow proposed patches for `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, the Stage 3.38 implementation record,
and the Stage 3.38 closure record. Historical failed-review sections, evidence hashes, and chronology
remain unchanged in meaning.

## 17. Adversarial review requirements

The planning and later implementation reviewers must challenge at least:

- raw-versus-trimmed OpenAPI admission order;
- accidental byte counting in Go;
- accidental UTF-16 code-unit counting in Web;
- supplementary-plane boundary vectors;
- malformed internal UTF-8 and Web unpaired-surrogate behavior;
- source-account-label trim identity across parser review, review history, signed token, append,
  idempotency hashing, broker identity, and PostgreSQL, including proof that no unsupported
  cross-version public replay exception is added;
- pre-Stage-3.39 portfolio command replay when new raw admission would otherwise reject before lookup;
- same-key/different-hash, in-flight, unsupported, corrupt, absent, and expired replay states under the
  new compatibility ordering;
- whether any source-label correction changes `ReviewParserVersion` semantics;
- whether `csvPayload` byte budget is weakened or mislabeled;
- accidental migration/schema expansion;
- accidental global JSON-decoder/auth/idempotency redesign beyond the narrow portfolio read-only compatibility ordering;
- accidental P3-06 `httpapi/api.go` decomposition;
- accidental future endpoint implementation;
- any normalization that changes canonical text identity;
- regression of Stage 3.29 note semantics;
- regression of Stage 3.30 import review integrity;
- regression of Stage 3.32 exact replay or Stage 3.38 24-hour expiry/reclamation authority;
- regression of Stage 3.35 password semantics;
- regression of Stage 3.37 timezone semantics.

Any material contradiction requires `REQUEST CHANGES`; it must not be hidden as a documentation note.

## 18. Planning validation and closure rule

### Planning change classification

`DOCUMENTATION_ONLY` + planning/governance bookkeeping.

No runtime code, OpenAPI executable contract, migration, dependency, CI workflow, or installer behavior
changes in the planning increment itself.

### Planning review history

Initial planning bundle SHA256 `65e0a6da1814da48cde5e676171b38d24718c79b1b6d72bc6a3a52e66aad0200`
received independent `REQUEST CHANGES` with one confirmed P2 planning blocker,
`STAGE-03-39-P2-01`. The reviewer demonstrated that the original plan's new raw portfolio-name bound
would reject a request before the existing replay store could return a pre-Stage-3.39 completed exact
artifact, violating the Stage 3.32 exact-response guarantee inside the Stage 3.38 24-hour authority
window. The failed review is preserved in the remediation review bundle and is not silently discarded.

Remediation v2 separated fresh portfolio admission from historical completed replay and added a narrow
read-only compatibility rule. Fresh independent review of v2 classified `STAGE-03-39-P2-01` as
**CLOSED BY REMEDIATION**, but returned `REQUEST CHANGES` with two new P3 planning blockers:

- `STAGE-03-39-P3-02`: active Stage 3.38 implementation/closure status metadata was not included in the
  post-PR-95 governance synchronization; and
- `STAGE-03-39-P3-03`: v2 incorrectly generalized the portfolio replay premise to
  `sourceAccountLabel`, even though the old public raw 120-byte guard makes such a newly-invalid
  historical public label impossible for valid UTF-8.

This v3 candidate addresses those findings by adding narrow current-state patch instructions for the
two Stage 3.38 documents while preserving their historical review chronology, and by removing the
unsupported source-label replay branch/import-handler ordering change. The existing Stage 3.32 import
recovery path is preserved unchanged. Neither new P3 finding is considered closed for governance
purposes until a fresh independent review of this exact v3 candidate returns `APPROVED`.

### Required planning evidence

Before the planning increment may be committed/pushed:

- exact remote baseline identified as `develop@c5962fa09b6d7d145dda203dbdf90069de7b1fcc`;
- complete proposed change instructions available for all five documentation surfaces;
- Markdown/scope consistency checked;
- no unrelated files in the candidate;
- mandatory independent read-only Internal Review Agent line-by-line review returns `APPROVED`;
- human explicitly authorizes the exact feature-branch commit/push action.

After publication:

- Draft PR targets `develop`;
- exact-head required CI is green;
- fresh independent external review returns `APPROVED`;
- human separately authorizes Ready/squash merge;
- actual planning merge is read back from GitHub.

P3-04 remains **OPEN** after planning merge. It becomes CLOSED only after a separately governed runtime
implementation and later closure-governance increment satisfy their own exact-head CI, independent
review, explicit human authorization, actual merge, and audit-state synchronization.

The planned runtime change categories are expected to be:

- `BACKEND_AFFECTED`;
- `FRONTEND_AFFECTED`;
- `OPENAPI_AFFECTED`;
- `IMPORT_FINANCIAL_AFFECTED`.

`POSTGRESQL_AFFECTED` is expected to be `NOT_APPLICABLE` because no migration/schema change is planned.
`AUTH_SECURITY_AFFECTED` is expected to be `NOT_APPLICABLE` only if implementation preserves the
Stage 3.35 auth path and does not broaden generic JSON decoding. If that assumption changes, the plan
must be revised and auth/security validation gates become mandatory.
