# Stage 3.39 — P3-04 Unicode and OpenAPI String-Length Semantics Implementation

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION MERGED — PR #97 squash-merged into protected `develop` at `abbd9f9f61574621e206f2e196b1fb8f056dc194`; closure status is MERGE-ACTIVATED: P3-04 is OPEN before the approved closure record is present on protected `develop` and CLOSED once that record is present |
| Date | 2026-08-27 |
| Finding | Original audit P3-04 — general Unicode / OpenAPI `minLength` / `maxLength` semantics |
| Canonical runtime base | `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` |
Canonical record: PR #96; commit(s) `32b198ee9d349f119ed374fd86d47622e27bcd73`.
| Runtime branch | `fix/stage-03-39-p3-04-runtime` |
| Runtime implementation commit | `740f5c313baac75d0406e9146a7bbc74a94d48c1` — `fix: align Unicode string length semantics` |
| Runtime-head CI | CI #274 / run `33049818896` — exact head `740f5c313baac75d0406e9146a7bbc74a94d48c1`, 10/10 required jobs successful |
Canonical record: commit(s) `de083b11f791c26c18ef635fc91d1322c281601b`.
| Pre-forensic publication baseline | Draft PR #97 at `de083b11f791c26c18ef635fc91d1322c281601b`; 2 commits; 17 changed files versus base before this forensic-history candidate |
| Baseline CI before forensic publication | CI #276 / run `33072520619` — exact baseline head `de083b11f791c26c18ef635fc91d1322c281601b`, 10/10 required jobs successful |
| Finding status | MERGE-ACTIVATED — OPEN before the approved closure record is present on protected `develop`; CLOSED once that record is present on protected `develop` |

> **Finding-ID note.** IDs such as `STAGE-03-39-P2-01`, `STAGE-03-39-P3-02`,
> `STAGE-03-39-P3-03`, `STAGE-03-39-P3-04`, `STAGE-03-39-P3-05`, and
> `STAGE-03-39-P3-06` below are **review-local IDs** from Stage 3.39 review cycles. They are not the
> original repository-audit P3-06/P3-05 findings. The original finding being remediated by this Stage
> is P3-04.

## 1. Problem

P3-04 tracks cross-layer disagreement over what a bounded human-readable string length means.
OpenAPI 3.1 / JSON Schema 2020-12 publishes character bounds, Go historically mixed
`len(string)` byte counts with rune counts, the browser used UTF-16 `maxlength`, and the import CSV
2 MiB resource limit was incorrectly represented as JSON Schema `maxLength`.

## 2. Root cause

The repository had no single explicit length unit for general human text. ASCII-only examples masked
three different units:

- UTF-8 bytes in some Go guards;
- Unicode code points in the OpenAPI/JSON Schema contract;
- UTF-16 code units in browser-native `maxlength`.

The defect also crossed existing compatibility boundaries. Portfolio create historically trimmed the
name before its 100-byte check and before request hashing, while import source-account labels were
already rejected by a raw byte guard before importer trim. Treating those two surfaces as equivalent
would either break exact replay or create unnecessary replay complexity.

## 3. Primary failure scenarios

1. A 100-code-point Cyrillic portfolio name can exceed 100 UTF-8 bytes and be rejected despite satisfying
   the OpenAPI contract.
2. A 100-emoji value occupies 200 UTF-16 code units and can be blocked by native browser `maxlength`.
3. A raw portfolio name longer than 100 code points can become 100 after trimming and bypass a raw
   contract limit if admission happens after trim.
4. A pre-Stage-3.39 completed portfolio command can lose exact replay after deployment if new raw
   validation executes before historical replay authority is checked.
5. A synthetic source-label compatibility branch can unnecessarily modify Stage 3.32 token/replay flow
   even though the old public raw byte guard makes the claimed historical population impossible.
6. `csvPayload.maxLength: 2097152` describes characters while runtime intentionally enforces 2 MiB of
   UTF-8 bytes.
7. A malformed UTF-8 CSV note can pass `encoding/csv.Reader`, reach importer code-point counting, and be
   classified differently from the frozen fail-closed importer contract unless validity is checked first.

## 4. Impact / project risk

The original P3-04 debt was low-severity contract drift, but review exposed two materially stronger
correctness concerns that had to be solved before closure:

- a **P2 cross-version exact-replay regression** could turn a previously successful ambiguous portfolio
  request into a post-deployment `400`, violating Stage 3.32 exact-response replay and Stage 3.38 finite
  authority guarantees;
- a malformed internal CSV note could violate the Stage 3.39 importer-entry fail-closed invariant and
  make importer review semantics diverge from application admission semantics.

unsupported source-label replay branch, incomplete PR disclosure, insufficiently literal boundary
vectors, and inaccurate permanent review-evidence statements. None of those were permitted to be
silently erased from history.


## 5.1 Review-local `STAGE-03-39-P2-01` — portfolio cross-version replay

### Problem

The initial planning contract required fresh raw portfolio admission at `<=100` Unicode code points
before trim while also claiming idempotency/replay behavior remained unchanged.

### Why it happened

The plan separated raw/trimmed/persisted string semantics but did not initially separate **fresh write
admission** from **recovery of an already completed command**.

### Failure / attack scenario

1. Before Stage 3.39, a client sends `POST /portfolios` with one leading space plus 100 ASCII (single-byte UTF-8) characters.
2. Old runtime trims first, obtains a normalized name of exactly 100 UTF-8 bytes, accepts it, creates the portfolio, and stores
   the exact replay artifact under the normalized command hash.
3. The client loses the HTTP response.
4. Stage 3.39 deploys while the command is still within the 24-hour Stage 3.38 authority window.
5. The client retries the same raw request with the same idempotency key.
6. A naive new raw-before-trim guard sees 101 code points and returns `400` before replay lookup.
7. The client receives validation failure instead of the exact original success artifact.

### What it threatened

Had that initial planning design been deployed unchanged, it would have violated the previously closed Stage 3.32 P2 exact-replay guarantee. The direct retry would not
itself create a second portfolio, but a client that interprets `400` as no prior success could initiate a
new operation/key and create a duplicate business object.

### Initial solution

Raw-before-trim validation was applied uniformly before the replay-aware create path, with no explicit
cross-version exception.

### Why review rejected it

The requirements were internally contradictory: fresh raw-over-limit input had to be rejected, while an
otherwise authentic matching unexpired completed command had to retain exact-response authority.

### Revised solution

The planning contract froze one narrow compatibility path:

`historical trim-first identity → generic read-only replay lookup → exact completed artifact`

Only the confirmed historical portfolio-create population is eligible. Missing/expired authority falls
back to the new raw validation and cannot create a fresh over-limit write. Same-key/different-hash stays
conflict; in-flight/unsupported/corrupt states remain fail closed.

### Why this solution was chosen

It reuses the existing generic `ReplayLookupStore`, preserves Stage 3.32 exact-response semantics and
Stage 3.38 finite authority, and does not redesign idempotency, change hash format, reserve/reclaim a key,
extend TTL, or create a second business effect.

### Required / implemented regression proof

- historical raw 101 → trimmed 100 exact retry returns stored status/body/request-id/trace-id;
- business effect remains one;
- fresh key with the same raw 101 is rejected;
- expired/missing historical authority cannot authorize a new generation;
- lookup errors fail closed;
- Stage 3.32 replay and Stage 3.38 expiry/reclamation invariants remain green.

### Residual limitation

Compatibility exists only for the evidenced historical portfolio-create case and only while old replay
authority remains valid. It is not a generic bypass around new validation.

## 5.2 Review-local `STAGE-03-39-P3-02` — stale Stage 3.38 canonical state

### Problem / cause

An intermediate planning remediation updated `SOURCE_OF_TRUTH.md` and `ROADMAP.md` to P3-05 CLOSED,
but left active top-level state in the Stage 3.38 implementation and closure dossiers saying P3-05 was
OPEN / PR #95 was not merged.

### Failure scenario / impact


### Revised solution

Synchronize all four active current-state surfaces while preserving historical failed-review chronology:

- `SOURCE_OF_TRUTH.md`;
- `ROADMAP.md`;
- Stage 3.38 implementation dossier active metadata/final status;
- Stage 3.38 closure dossier active metadata/final status.

Historical `changes required`, old CI heads, and prior OPEN statements remain only as explicitly
historical evidence.

### Evidence

The next technical reassessment confirmed the four-surface synchronization and closed this review-local P3.

## 5.3 Review-local `STAGE-03-39-P3-03` — unsupported source-label replay branch

### Problem

An intermediate plan treated `sourceAccountLabel` like portfolio name and proposed a new public
cross-version historical replay exception plus changes to Stage 3.32 import replay recovery ordering.

### Why review rejected it

The existing public HTTP boundary already checked raw `len(sourceAccountLabel) <= 120` **before**
importer trim. For valid UTF-8:

`Unicode code points <= UTF-8 bytes`.

Therefore any historically admitted public label with `<=120` bytes already had `<=120` code points.
The byte→code-point change expands multilingual admission; it does not create a newly-invalid historical
public population.

### Second attack / scope scenario

A planned regression tried to seed a historical public append whose raw label exceeded 120 only because
of trim-removable whitespace. That command cannot be produced through the conforming old public path:
the raw byte guard rejects it before importer trim/token/replay processing.

### Revised solution

Remove the synthetic source-label compatibility branch; do not change `import_replay_recovery.go` or
import-handler replay ordering for P3-04. Keep:


`ReviewParserVersion` remains `2`, and existing Stage 3.32 recovery regressions stay unchanged/green.

### Why this solution was chosen

It is the smaller KISS/YAGNI solution and avoids unnecessary changes to security/correctness-sensitive
signed-token and replay control flow.

### Evidence

Subsequent technical reassessment confirmed the unsupported replay scope was removed and closed this finding.


The first published planning head `f07b305cedb6532763ef0ae2ac1ad288f8e437c0` passed CI #272 but the
exact-head CI evidence were not all explicitly recorded.

**Failure scenario:** a green documentation PR could reach human merge without explicit governance

**Remediation:** update only the PR description with all mandatory disclosures and exact CI evidence.
No code/doc commit was required for this finding by itself.


showed that this could allow an implementation to prove `100 Cyrillic PASS` and only `101 ASCII REJECT`
while leaving a multibyte reject-boundary bug undetected.

The plan was tightened to literal mandatory vectors:

- portfolio `100 ASCII PASS / 101 ASCII REJECT`;
- `100 Cyrillic PASS / 101 Cyrillic REJECT`;
- supplementary-plane `100 PASS / 101 REJECT`;
- source label `120 Cyrillic PASS / 121 Cyrillic REJECT`;
- source label `120 supplementary PASS / 121 supplementary REJECT`;
- relevant Stage 3.29/3.32/3.35/3.37/3.38 regression suites remain green.

The remediation was published as planning head `2c453291ce5e41183d930874cba48382245573ea`.
returned `APPROVED`, and PR #96 was squash-merged as
`32b198ee9d349f119ed374fd86d47622e27bcd73`.


The final runtime contract was deliberately narrow:

- bounded public human text uses valid UTF-8 plus Unicode code-point counting;
- fresh raw portfolio and asset-search admission occurs before trim;
- transaction note and source-account-label application defenses use code points and reject malformed
  internal UTF-8;
- source-account-label trim/token/command/broker/persistence identity remains unchanged;
- no source-label historical replay branch is added;
- `ReviewParserVersion` remains `2`;
- the CSV limit remains exactly `2 * 1024 * 1024` UTF-8 bytes;
- OpenAPI removes CSV character `maxLength` and publishes
  `x-openinvest-max-utf8-bytes: 2097152`;
- Web uses explicit well-formed-Unicode/code-point helpers rather than native UTF-16 `maxlength`;
- no NFC/NFD/NFKC normalization, case folding, transliteration, grapheme-cluster policy, dependency,
  migration, service, worker or generic raw-JSON rewrite is introduced.


## 7.1 Review-local `STAGE-03-39-P3-06` — initial runtime candidate rejected

### Problem

The initial complete runtime candidate correctly changed general text length checks to code points, but
`backend-go/internal/importer/importer.go::normalizeCandidate(...)` still did:

```go
note := neutralizeSpreadsheetText(...)
if utf8.RuneCountInString(note) > 500 {
    ...
}
```

with no `utf8.ValidString(note)` first.

### Why it happened

The implementation hardened the application transaction-note boundary and importer source-label
boundary, but implicitly assumed the CSV parser had already produced well-formed Unicode note text.
That assumption is false: Go's `encoding/csv.Reader` parses CSV structure, not UTF-8 validity.

### Failure / attack scenario

A CSV row contains a note with raw byte `0xff`. `encoding/csv.Reader` returns the field as a Go string
without a UTF-8 validation error. Spreadsheet neutralization does not make that byte sequence valid.
The candidate then executes `RuneCountInString` and can form an importer review candidate instead of
failing closed at the importer entry point.

### What it threatened

No authentication bypass or proven financial corruption was demonstrated, so severity remained P3.
However, the frozen Stage 3.39 contract explicitly required importer entry points to reject malformed
internal UTF-8 **before** code-point counting. Without the guard:

- importer review and later application admission could disagree on the same malformed note;
- the regression suite did not prove the required importer-entry guarantee;
- malformed internal text could be classified as appendable rather than deterministic invalid input.

### Initial solution

Post-neutralization note length was counted with `utf8.RuneCountInString`, relying on the existing CSV
pipeline for input shape.

### Second attack scenario / required proof

The correction had to be tested through the real path:

`ReviewCSV → encoding/csv → normalizeCandidate`

not by calling a helper directly. The row also had to become non-appendable with a deterministic,
non-sensitive reason while preserving post-neutralization ordering and parser compatibility.

### Revised solution

The v7 remediation changed the sequence to:

```go
note := neutralizeSpreadsheetText(...)
if !utf8.ValidString(note) {
    return ..., []string{"NOTE_INVALID_UTF8"}
}
if utf8.RuneCountInString(note) > 500 {
    return ..., []string{"NOTE_TOO_LONG"}
}
```

The row becomes `INVALID`, `Candidate == nil`, and the reason code contains no raw user data.
`ReviewParserVersion` remains `2`.

### Why this solution was chosen

It is the narrowest fail-closed boundary that satisfies the frozen contract. Validation remains on the
actual post-neutralization value that would become the candidate note. It avoids a generic transport
rewrite and avoids an unjustified parser-version bump.

### Regression test

`TestStage339ImporterRejectsMalformedCSVNoteBeforeCodePointCounting` constructs a real CSV payload
containing `string([]byte{0xff})`, runs it through `ReviewCSV`, and proves:

- one reviewed row;
- status `INVALID`;
- no candidate;
- reason `NOTE_INVALID_UTF8`;
- `ReviewParserVersion == 2`.

`encoding/csv`.

### Residual limitation

P3-04 does not add a generic lossless raw-JSON Unicode scanner. Standard JSON transport behavior and
the password-specific Stage 3.35 decoder remain separate contracts. The importer guard covers the
internal CSV reachability that review proved.

# 8. Final runtime behavior and compatibility boundaries

## 8.1 Portfolio create

Fresh requests validate raw valid UTF-8/code points before trim. Only a newly-over-limit raw request
that could have passed the exact old trim-first 100-byte rule is eligible for historical read-only
lookup. A matching unexpired completed artifact is replayed exactly; missing/expired authority grants
no write capability.

The compatibility path never reserves, reclaims, writes, creates a generation, extends TTL, changes

## 8.2 Asset search

Raw query validity/code-point bound is enforced before trim so the API contract cannot be bypassed by
trim-removable characters. Cursor/filter identity continues to use the existing trimmed query.

## 8.3 Transaction note


## 8.4 Source account label


## 8.5 CSV payload

The 2 MiB limit is a byte/resource ceiling. Runtime and browser `File.size` remain byte-based. OpenAPI
publishes `x-openinvest-max-utf8-bytes: 2097152` and no longer misuses JSON Schema character
`maxLength: 2097152`.

# 9. Regression evidence carried by the runtime

Focused tests prove:

- portfolio 100/101 ASCII;
- portfolio 100/101 Cyrillic;
- portfolio 100/101 supplementary-plane;
- malformed portfolio UTF-8;
- raw-before-trim portfolio rejection;
- exact historical portfolio replay / no second business effect;
- missing/expired authority cannot create a new generation;
- replay lookup failure is fail closed;
- asset 100/101 supplementary-plane and raw-before-trim behavior;
- transaction note 500/501 and malformed UTF-8;
- source label 120/121 Cyrillic;
- source label 120/121 supplementary-plane;
- source-label trim identity;
- malformed source label;
- malformed CSV note through real `ReviewCSV` flow;
- `ReviewParserVersion == 2`;
- exact 2 MiB CSV byte admission;
- OpenAPI Unicode semantics and CSV byte extension;
- Web code-point counting and unpaired-surrogate rejection;
- removal of affected native `maxlength`;
- continued browser `File.size` byte check.

Pre-publication local gates also covered `git diff --check`, focused and full Go tests, Go vet, OpenAPI
validation, frontend tests/typecheck/build, and dependency/migration scope checks.

# 11. Documentation review history after runtime publication

## 11.1 `EXT-PR97-P3-01` — published Stage report was stale

### Problem

The implementation report committed with the runtime still said:

- local runtime candidate;
- not committed / not pushed / no PR;
- branch local-only;
- authoritative CI still future;

### Why it happened

The report accurately described the pre-publication local candidate when authored, but it had not been

### Failure scenario / threat

If merged unchanged, the permanent Stage 3.39 record would falsely say the runtime had never been
corruption even though runtime behavior is correct.

### Initial docs remediation

The first docs-only candidate updated publication facts, recorded CI #274, published the previously

### External-review disposition

The external runtime review explicitly found **no blocking technical defect**; its only blocker was this
stale durable report. The substantive stale-state defect was fixed by the docs candidate.

## 11.2 `INT-DOCS-P3-01` — first docs remediation made an unsupported governance assertion

### Problem

The first docs-only remediation additionally stated that the runtime had been published after an
explicit separate human commit/push authorization and repeated that the human had separately authorized
the exact candidate.

### Second attack scenario


### Revised docs v2.1 solution

Remove only the two unsupported authorization assertions. Preserve the evidenced sequence:


without claiming evidence that was not present in the package.

### Why chosen

needs to change.

### Regression / evidence

The complete revised docs-only v2.1 patch SHA256 was
`f48e6bd6cd880ecd16921698edb1b293c3c68958fb1801b8bf10da6d20656b8b` with proposed blob
`b2441a42108756654a5816c1387f44c597e31de4`.

Fresh docs-only internal validation:


The approved evidence-only document was published as commit
`de083b11f791c26c18ef635fc91d1322c281601b`. Fresh exact-head CI #276 / run `33072520619` then
completed `success` with all 10 required jobs green.

The Draft PR #97 description was subsequently synchronized to the then-current baseline head `de083b11...`, CI #276,

# 13. Residual limitations / explicitly unaddressed behavior

Stage 3.39 deliberately does **not** define or change:

- NFC/NFD/NFKC normalization;
- locale-aware case folding;
- grapheme-cluster/user-perceived-character limits;
- transliteration;
- generic raw-JSON lossless malformed-Unicode handling;
- password Stage 3.35 transport semantics;
- timezone Stage 3.37 semantics;
- a source-label historical replay exception;
- request-hash/token/cursor format versions;
- database schema/migrations;
- financial arithmetic/Decimal/tax/snapshot methodology;
- P3-06 `httpapi/api.go` decomposition;
- P3-07 transaction form fixture/default cleanup;
- P3-08 migration-validator hardening;
- P3-09 Next.js/React maintenance;
- P3-10 Fiber maintenance;
- future mutation/update/reversal endpoints.

Malformed UTF-8 handling is intentionally fail-closed at the affected Go/importer entry points proven by
this Stage; the Stage does not claim that all possible transports preserve malformed byte sequences
unchanged before those boundaries.

# 14. Closure state and publication-independent gate model

## 14.1 Historical sequence recorded at forensic-v2 candidate preparation

At the time forensic v2 was still a pre-commit candidate, the governed sequence was:

2. publish the approved forensic-history document as a docs-only commit;
3. run fresh CI on the resulting exact PR head;
6. remediate any blocking finding and rerun the affected exact-head gates;
7. obtain explicit human Ready/merge authorization;
8. squash-merge PR #97 into protected `develop`;
9. record the actual merge SHA and only then mark original audit P3-04 CLOSED.

This list is intentionally historical. It describes the sequence as it existed at forensic-v2 candidate
preparation and is **not** a live “remaining gates” checklist.

`APPROVED`, was published as docs-only commit
`6d419f5931ad868b7280623ad447550b0befc6a9`, and exact-head CI #277 / run `33077556962` completed
10/10 successful jobs.

## 14.2 `EXT-PUBLISHED-P3-01` — stale “remaining gates” wording on the published forensic head

### Problem

External published-head verification of
`6d419f5931ad868b7280623ad447550b0befc6a9` found that this section still introduced the pre-publication
sequence with **“Required remaining gates are”**, even though pre-commit approval, forensic publication,
and exact-head CI had already completed.

### Root cause

Forensic v2 correctly removed the self-invalidating claim that `de083b11...` would remain the permanent
current head, but the closure section still treated a candidate-preparation checklist as live mutable
state. The SHA wording was publication-stable; the **completion-state label was not**.

### Failure scenario / impact

Immediately after forensic v2 publication and CI #277, the durable Stage report and the live PR body
could disagree about which gates were actually outstanding. That recreates the same audit-truth class
that earlier caused `EXT-PR97-P3-01`: a technically correct runtime accompanied by a factually stale
permanent lifecycle record.

The finding is documentation/governance P3 only. It does not change or invalidate runtime behavior,
OpenAPI, replay authority, migrations, security/privacy behavior, or financial semantics.

### Initial solution and why review rejected it

The initial forensic solution renamed the section
“publication-independent remaining gates” and avoided predicting a future SHA. External verification
showed that this was insufficient because the numbered list still called already-completed steps
“remaining”.

### Revised solution

The candidate-preparation sequence above is now explicitly **historical evidence**. Live mutable state is
not encoded as a permanent “current/remaining” checklist.

Instead, the durable publication-independent obligations are:

- any later head-changing documentation/evidence remediation must be verified on the exact resulting
  published head before merge consideration;
- required CI must be green on that exact resulting head;
- both Internal and External reviewers must verify the latest published evidence/documentation state;
- any blocking finding must be remediated and the affected gates repeated;
- Ready/merge requires explicit merge gate;
- merge into `develop` remains squash-only;
- original audit P3-04 becomes CLOSED only after the actual protected-branch merge and closure record.

These are workflow invariants, not assertions that a particular step is currently incomplete. The
document intentionally does not predict the SHA of this remediation if it is later published.

### Residual limitation

The live mutable status of a Draft PR — exact current head, current CI run, or which verification is in
progress — belongs to GitHub PR/CI evidence. This durable Stage report records historical completed facts
and invariant closure requirements so that publication does not make its own statements false.

## 14.3 `INT-DOCS-P3-02` — historical completion wording implied an unsupported authorization event

### Problem

The first `EXT-PUBLISHED-P3-01` remediation correctly converted the stale live-gate checklist into a
historical sequence, but historical step 2 said the forensic-history document was published
**“under a separately authorized docs-only commit/push gate”** and the next paragraph said items 1–3
were subsequently completed.

### Root cause

The remediation mixed two different things:

- an evidenced historical fact: the docs-only forensic commit exists and was published;
- a normative governance requirement: protected publication actions require separate merge gate.

The review package proved the first fact but did not itself contain direct evidence of the distinct
authorization event. Treating completion of the publication step as proof of the authorization event
recreated the same evidence-integrity class previously caught by `INT-DOCS-P3-01`.

### Failure scenario / impact

A future auditor could read the permanent Stage report and conclude that the repository evidence proved
publication state.

This is documentation/governance P3 only. It has no runtime, security, financial, replay, database,
OpenAPI, importer, or frontend impact.

### Initial solution and why review rejected it

The first remediation tried to preserve the full governed sequence verbatim, including the phrase
“separately authorized docs-only commit/push gate”, while also saying those historical steps were
completed.

preceding separate authorization event.

### Revised solution

The historical sequence now records only the evidenced event:

> publish the approved forensic-history document as a docs-only commit

The separate human-authorization requirement remains in the publication-independent workflow
invariants, where it belongs as a normative rule rather than as a claimed historically evidenced event.

### Why this solution was chosen

It is the smallest correction that keeps both kinds of truth intact:

- historical evidence says what the available artifacts prove;
- governance invariants say what the process requires.

No runtime or repository-behavior claim needs to change.

### Residual limitation


## 14.4 Closure invariant

PR #97 was actually squash-merged into protected `develop` at
`abbd9f9f61574621e206f2e196b1fb8f056dc194` from exact final published head
`26f5ca18ca5772db569d22ce2eff64d5a7850b1b` after CI #279 / run `33121609429` completed

The implementation merge satisfies the runtime merge gate but does not by itself close original audit
P3-04. P3-04 remains **OPEN** until the separate Stage 3.39 closure-governance record and synchronized
canonical surfaces are present on protected `develop`.

Once that closure-governance record is present on protected `develop`, P3-04 is **CLOSED**. The
merge-activated rule deliberately does not predict the future closure-PR head, CI number or merge SHA.

for the separate closure-governance record.
