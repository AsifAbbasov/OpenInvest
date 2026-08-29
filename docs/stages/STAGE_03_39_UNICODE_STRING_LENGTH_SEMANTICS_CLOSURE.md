# Stage 3.39 — P3-04 Unicode and OpenAPI String-Length Semantics Closure

| Field | Value |
| --- | --- |
| Status | MERGE-ACTIVATED CLOSURE RECORD — before protected activation this document is the closure candidate and P3-04 remains OPEN; once this record is present on protected `develop`, it is the canonical closure record and P3-04 is CLOSED |
| Date | 2026-08-28 |
| Finding | Original audit P3-04 — general Unicode / OpenAPI `minLength` / `maxLength` semantics |
| Planning gate | PR #96 merged into `develop` at `32b198ee9d349f119ed374fd86d47622e27bcd73` |
| Runtime / forensic PR | PR #97 — `fix: align Unicode string length semantics (Stage 3.39)` |
| Final published PR head | `26f5ca18ca5772db569d22ce2eff64d5a7850b1b` |
| Implementation squash merge | `abbd9f9f61574621e206f2e196b1fb8f056dc194` |
| Exact-head final CI | CI #279 / run `33121609429`, 10/10 required jobs successful on `26f5ca18ca5772db569d22ce2eff64d5a7850b1b` |
| Final Internal published-head verification | `APPROVED`; P0/P1/P2/P3 none on exact head `26f5ca18ca5772db569d22ce2eff64d5a7850b1b` |
| Final External published-head verification | `APPROVED`; P0/P1/P2/P3 none on exact head `26f5ca18ca5772db569d22ce2eff64d5a7850b1b` |
| Published Stage-report blob before merge | `2b772d27b174ec0c196262b57935a6632191f1fa` |
| Closure runtime scope | None — documentation/governance synchronization only |
| Closure activation rule | Before this record is merged into protected `develop`, P3-04 remains OPEN. Once this record and its synchronized canonical surfaces are present on protected `develop`, P3-04 is CLOSED. |
| Post-closure original audit backlog | P0=0 / P1=0 / P2=0 / P3=5: P3-06, P3-07, P3-08, P3-09, P3-10 |

## 1. Closure basis

Stage 3.39 remediated original audit P3-04 by aligning implemented bounded human-readable strings on
Unicode code-point semantics across OpenAPI, Go and Web while keeping the import CSV 2 MiB limit
explicitly byte-based.

The implementation and its complete forensic/evidence history were squash-merged through PR #97 into
protected `develop` at:

`abbd9f9f61574621e206f2e196b1fb8f056dc194`

The implementation merge is an established repository fact. This closure change does not modify that
runtime and does not create a second implementation event.

## 2. Exact implementation evidence

The final published PR #97 head was:

`26f5ca18ca5772db569d22ce2eff64d5a7850b1b`

Fresh exact-head CI #279 / run `33121609429` completed successfully with all ten required jobs:

- Go tests;
- Python tests;
- Frontend build and typecheck;
- OpenAPI contract;
- Docker Compose config;
- PostgreSQL migration validation;
- Go vet;
- Go race tests;
- Go vulnerability scan;
- Dependency security scan.

The final published Stage-report blob before merge was:

`2b772d27b174ec0c196262b57935a6632191f1fa`.

## 3. Final reviewer evidence

Both final published-head verification roles independently returned `APPROVED` on exact head
`26f5ca18ca5772db569d22ce2eff64d5a7850b1b` with P0/P1/P2/P3 = None.

The final Internal verification confirmed exact pre-commit-to-published equivalence, 765/765 Stage-report
lines reviewed, CI #279 10/10, no stale lifecycle statement, and no runtime impact.

The final External verification independently confirmed the published identity, CI #279, the complete
forensic history, all prior documentation/evidence findings resolved, and no runtime/governance blocker.

These reviewer verdicts authorize no repository mutation by themselves; the separate human Ready/merge
gate was subsequently exercised and PR #97 was actually squash-merged as `abbd9f9f61574621e206f2e196b1fb8f056dc194`.

## 4. Technical contract closed by P3-04

The merged implementation establishes:

- valid UTF-8 plus Unicode code-point counting for affected bounded human-readable Go surfaces;
- raw-before-trim admission where the public contract requires it;
- narrow read-only historical replay compatibility only for the evidenced portfolio-create population;
- no analogous source-account-label historical replay branch;
- malformed importer CSV notes fail closed before code-point counting;
- explicit Web well-formed-Unicode/code-point validation instead of UTF-16 `maxlength` semantics;
- OpenAPI character bounds that describe code points;
- a separate 2 MiB UTF-8 byte/resource contract for `csvPayload`;
- no normalization, grapheme-cluster, case-folding, schema, migration, financial arithmetic, Decimal,
  authentication, privacy, or infrastructure expansion.

## 5. Review-driven history preserved

The Stage 3.39 implementation dossier permanently preserves the material review chain, including:

- `STAGE-03-39-P2-01` — cross-version portfolio exact-replay risk;
- `STAGE-03-39-P3-02` — stale Stage 3.38 canonical lifecycle state;
- `STAGE-03-39-P3-03` — unsupported source-label replay branch;
- planning disclosure and literal-boundary-vector findings;
- `STAGE-03-39-P3-06` — malformed CSV-note UTF-8 admission;
- `EXT-PR97-P3-01` — stale published Stage report;
- `INT-DOCS-P3-01` — unsupported authorization assertion;
- `INT-FORENSIC-P3-01` — inaccurate historical replay example;
- `INT-FORENSIC-P3-02` — volatile current-head wording;
- `EXT-PUBLISHED-P3-01` — stale live “remaining gates” wording;
- `INT-DOCS-P3-02` — historical wording implying an unsupported separately evidenced authorization event.

No failed review is erased by this closure record.

## 6. Why closure is documentation-only

PR #97 already placed the approved runtime and forensic/evidence implementation into protected
`develop`. Closure therefore changes only canonical governance state:

- `docs/SOURCE_OF_TRUTH.md`;
- `docs/ROADMAP.md`;
- `docs/stages/STAGE_03_39_UNICODE_STRING_LENGTH_SEMANTICS_IMPLEMENTATION.md`;
- this closure dossier.

No Go, TypeScript, OpenAPI executable contract, SQL, migration, dependency, workflow, security control,
replay authority, importer behavior, or financial calculation is changed.

## 7. Publication-stable closure activation

This record deliberately does not predict a future closure-PR head, CI number or squash-merge SHA.

The activation rule is structural:

1. while this closure candidate is not part of protected `develop`, original audit P3-04 remains OPEN;
2. if a review finds a blocker, it is remediated before merge and affected gates repeat;
3. once this exact closure record and synchronized canonical surfaces are squash-merged into protected
   `develop`, P3-04 is CLOSED;
4. the canonical post-closure original audit backlog is then P0=0 / P1=0 / P2=0 / P3=5:
   P3-06, P3-07, P3-08, P3-09, P3-10.

Because the rule is conditional on presence in protected `develop`, publishing a Draft closure PR does
not prematurely claim closure, and the record does not become stale merely because its own PR head
changes during governed remediation.

## 8. Residual limitations / explicitly unaddressed scope

Stage 3.39 does not address:

- original audit P3-06 `httpapi/api.go` decomposition;
- original audit P3-07 transaction-form fixture/default semantics;
- original audit P3-08 migration-validator policy hardening;
- original audit P3-09 Next.js maintenance;
- original audit P3-10 Fiber maintenance;
- Unicode normalization/case folding/grapheme policy;
- generic raw-JSON transport rewriting;
- unrelated privacy Stage 3.25 work.

Those items remain separately governed.

## 9. Closure decision

The implementation evidence is complete and merged.

Under `docs/REVIEW_WORKFLOW.md`, before protected activation this closure-governance change must pass
Internal Review, human publication authorization, Draft PR publication, exact-head CI, independent
External Review, and human squash-merge authorization. After protected activation, this sentence is a
stable statement of the required pre-activation path rather than a claim that those gates are still
pending.

Until this closure-governance record is actually present on protected `develop`, **P3-04 remains OPEN**.

When it is present on protected `develop`, **P3-04 is CLOSED** and the original 32-finding audit becomes:

- closed: 27 / 32;
- completion: 84.375%;
- remaining: 5 / 32;
- remaining findings: P3-06, P3-07, P3-08, P3-09, P3-10.

## 10. Closure Internal Review evidence — published after independent External verdict

The current closure Internal Review evidence was intentionally withheld from the clean Draft PR until
the independent External Review recorded its own verdict, as required by `docs/REVIEW_WORKFLOW.md`.

That independent External verdict has now been recorded out-of-band in the independent review record
for Draft PR #99; no GitHub-native PR review/comment artifact is claimed. The previously withheld
Internal evidence and the blindness-remediation chain are therefore published here as permanent
forensic evidence. Publishing this history does not authorize Ready/merge and does not close P3-04.

### 10.1 First complete closure-candidate Internal Review

The first complete four-file closure candidate received `REQUEST CHANGES` with two documentation /
governance P3 findings.

#### `INT-CLOSURE-P3-01` — active lifecycle fields would self-stale at protected activation

**Problem.** The candidate used hard-coded pre-activation values such as `CLOSURE CANDIDATE`,
`Finding status | OPEN`, and present-tense wording that review/publication gates were still future work.

**Why it happened.** The decision rule was merge-activated, but several metadata fields still encoded
one temporary lifecycle phase as an always-current fact.

**Failure scenario.** If that candidate were merged into protected `develop`, the canonical closure
record would still call itself a candidate and expose an active OPEN state even though its own
merge-activated rule would say P3-04 was CLOSED.

**Impact.** Canonical governance surfaces would contradict one another at closure activation.
Runtime behavior was unaffected.

**Rejected approach.** Keep merge-activated semantics only in prose while leaving temporary metadata
hard-coded.

**Remediation.** Lifecycle fields became phase-neutral and merge-activated. Before the approved closure
record is present on protected `develop`, P3-04 is OPEN; once present, P3-04 is CLOSED.

**Result.** RESOLVED before clean blind-review publication.

#### `INT-CLOSURE-P3-02` — moving `develop` HEAD was frozen as a permanent current baseline

**Problem.** The first candidate described `abbd9f9f61574621e206f2e196b1fb8f056dc194` as the current canonical `develop` baseline.

**Why it happened.** That SHA was genuinely the protected `develop` HEAD during preparation, but the
wording confused a moving branch pointer with an immutable implementation fact.

**Failure scenario.** The closure squash merge must advance `develop`, so the newly merged Source of
Truth would immediately become stale while still calling the implementation SHA the current branch head.

**Impact.** Canonical branch-state evidence would become false at closure activation. Runtime behavior
was unaffected.

**Rejected approach.** Hard-code the known implementation merge SHA as permanent current branch state.

**Remediation.** `abbd9f9f61574621e206f2e196b1fb8f056dc194` is recorded only as the immutable **Stage 3.39 implementation merge
baseline**. The moving protected `develop` HEAD is not hard-coded and no future closure SHA is predicted.

**Result.** RESOLVED before clean blind-review publication.

### 10.2 Closure remediation Internal re-review

After both findings above were remediated, the complete revised closure candidate received fresh
Internal Closure Re-Review:

- exact reviewed four-file precommit patch SHA256:
  `847369c094112e60793b330581cda7c317ca12bd62aa37574efd4745c7f0616a`;
- `INT-CLOSURE-P3-01` — RESOLVED;
- `INT-CLOSURE-P3-02` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`.

That approval still did not permit disclosure of the current closure Internal result inside the blind
External Review target. The workflow required the evidence to remain out-of-band until the independent
External verdict.

## 11. Blind-review remediation history

### 11.1 Disclosure defect discovered before External Review

The first published closure Draft PR (#98) exposed the current closure Internal findings/verdict in its
PR body and committed closure dossier before an independent External verdict existed.

Because `docs/REVIEW_WORKFLOW.md` requires External Review **without internal verdict disclosure**,
PR #98 could not serve as the independent blind-review target.

A first proposed remediation attempted to replace the committed Internal evidence with
`WITHHELD — blind external review pending` and to rewrite the same PR's branch/body. Fresh Internal
Blindness Remediation Review returned `REQUEST CHANGES` with two blocking governance P3 findings.

#### `INT-BLIND-P3-01` — blind-safe PR body omitted mandatory Required PR disclosure

**Problem.** The blind-safe body correctly removed current Internal verdict/findings but became too
minimal and omitted mandatory disclosure categories.

**Missing categories.** ADR; DDD/bounded contexts; database/schema/migrations; mathematical
calculations; performance/cost; security/privacy; external data sources; backward compatibility;
rollback; review-size budget; explicit out-of-scope.

**Impact.** The remediation would have restored blind wording while violating a different mandatory
governance rule.

**Remediation.** The clean PR body was expanded to explicitly cover every Required PR disclosure item
while keeping current closure Internal evidence withheld.

**Result.** RESOLVED.

#### `INT-BLIND-P3-02` — an already contaminated PR could not be made genuinely blind by editing it

**Problem.** PR #98 had already publicly carried the current Internal findings/verdict.

**Failure scenario.** Editing the body and rewriting the feature branch would clean the active diff,
but the same PR object could retain discoverable edit/activity/history paths that exposed the previous
Internal result.

**Impact.** Reusing PR #98 would not provide a robust independent blind-review object and could anchor
the External reviewer.

**Rejected approach.** Rewrite branch + edit the same PR #98.

**Remediation.** PR #98 remained only a failed/contaminated forensic attempt. A new clean branch was
created directly from `abbd9f9f61574621e206f2e196b1fb8f056dc194`, containing one clean four-file closure commit with current Internal
evidence withheld, and a new Draft PR (#99) became the sole blind External Review target.

**Result.** RESOLVED.

### 11.2 Fresh Internal Clean Blind-Closure Review

The corrected clean candidate received a fresh separate Internal Review:

- exact clean four-file patch SHA256:
  `d55dcc3aa8daac04aebd7c398a7e97d178add4c87f2c9f14dbb5106d6db92030`;
- proposed closure dossier blob:
  `0d919cfe6a637f5eea8687a374101ff6926d1ba9`;
- required base: `abbd9f9f61574621e206f2e196b1fb8f056dc194`;
- `INT-BLIND-P3-01` — RESOLVED;
- `INT-BLIND-P3-02` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`.

The reviewed candidate was published as one clean commit:

`618927a4a336691059621b120696d0895afd136d` — `docs: prepare blind Stage 3.39 P3-04 closure`

with exact parent `abbd9f9f61574621e206f2e196b1fb8f056dc194`.

Draft PR #99 was created from that clean head. CI #281 / run `33157801196` completed on exact
head `618927a4a336691059621b120696d0895afd136d` with all 10 required jobs successful.

## 12. Independent External blind review and PR-body remediation

The independent External reviewer evaluated Draft PR #99 without access to the current closure
Internal findings/verdict.

### 12.1 First External verdict — `EXT-PR99-P3-01`

The first External review confirmed exact clean identity/scope, CI #281 10/10, correct closure
semantics, no runtime defect, and correct post-closure backlog. It returned `REQUEST CHANGES` for one
PR-metadata P3 only.

#### `EXT-PR99-P3-01` — live PR body retained pre-publication identity/CI wording

**Problem.** After PR #99 and CI #281 already existed, the PR body still used pre-publication
phrasing and stated only that exact-head CI would be required after publication.

**Impact.** Repository files were correct, but the live PR disclosure was stale and did not record the
actual exact-head CI evidence required by the PR disclosure contract.

**Remediation.** PR-body-only metadata was updated, with no repository file or commit change, to record:

- current PR `#99`;
- current base `abbd9f9f61574621e206f2e196b1fb8f056dc194`;
- current exact head `618927a4a336691059621b120696d0895afd136d`;
- current head branch `docs/stage-03-39-p3-04-closure-blind`;
- CI #281 / run `33157801196` — 10/10 required jobs successful;
- only the future evidence-only follow-up head and final closure squash-merge SHA remain intentionally
  unknown.

The body continued to keep current closure Internal findings/verdict withheld until the External verdict.

### 12.2 External re-review

Fresh live PR-body re-review confirmed:

- Draft PR #99 remained OPEN / Draft / unmerged;
- exact repository head remained `618927a4a336691059621b120696d0895afd136d`;
- one commit / four files / +214 / -14 remained unchanged;
- CI #281 remained 10/10 successful on the exact head;
- `EXT-PR99-P3-01` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`.

The independent External verdict therefore exists before publication of the current closure Internal
evidence in this section, satisfying the anti-anchoring order required by `docs/REVIEW_WORKFLOW.md`.

## 13. Evidence-publication Internal Review remediation

Fresh Internal review of the first evidence-publication candidate returned `REQUEST CHANGES` for one
documentation/evidence-integrity P3 only:

### `INT-EVIDENCE-P3-01` — External verdict provenance wording overstated its hosting location

**Problem.** The first evidence-publication candidate said that the independent External verdict had
been “recorded on Draft PR #99”.

**Why it happened.** The wording correctly associated the External verdict with PR #99 as its review
target, but collapsed two distinct facts: an independent review record exists for PR #99, while the
live GitHub PR discussion does not itself contain a GitHub-native review/comment artifact carrying that
verdict.

**Failure scenario.** A future reader could interpret the canonical dossier as proof that PR #99
itself hosts the External approval record even though live GitHub discussion evidence does not establish
such a hosted artifact.

**Impact.** Provenance wording would be stronger than the available evidence. Runtime behavior,
closure semantics, audit arithmetic, CI identity, and branch state were unaffected.

**Rejected approach.** Keep the shorter “recorded on Draft PR #99” wording and rely on context to
distinguish review target from evidence-hosting location.

**Remediation.** The dossier now states that the verdict was recorded **out-of-band in the independent
review record for Draft PR #99** and explicitly states that no GitHub-native PR review/comment artifact
is claimed.

**Result.** Provenance was narrowed to the evidence actually available. The subsequent fresh
Internal Evidence-Publication Re-Review confirmed:

- `INT-EVIDENCE-P3-01` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`;
- exact approved evidence-only patch SHA256:
  `4cb3ad27e7709722fdd137c741b372c5dca7791fc67f21a3df77fd7aa0b88981`;
- resulting published closure dossier blob:
  `fd7679f926e33897a4ad5377e9b14ded99116479`.

That approval occurred before publication of evidence-only commit
`1a7155eaf7c8195f11547e1e7f921dd32a896662` (`docs: publish Stage 3.39 closure review evidence`).

## 14. Final exact-head External verification remediation

Final External exact-head verification was performed against Draft PR #99 at published head
`1a7155eaf7c8195f11547e1e7f921dd32a896662` after CI #282 / run `33177434968` completed 10/10 successfully.

The reviewer confirmed live PR identity, the two-commit / four-file documentation-only scope,
protected `develop` at `abbd9f9f61574621e206f2e196b1fb8f056dc194`, current dossier blob `fd7679f926e33897a4ad5377e9b14ded99116479`, PR-body synchronization,
closure activation semantics, backlog arithmetic, and no runtime impact.

The verdict was `REQUEST CHANGES` for one documentation / forensic-completeness P3 only:

### `EXT-FINAL-P3-01` — permanent dossier omitted the completed Evidence-Publication Internal Re-Review

**Problem.** Section 13 recorded the `INT-EVIDENCE-P3-01` remediation but still ended with
“Fresh Internal re-review is required...” and did not preserve the already-completed re-review that
authorized publication of the evidence-only follow-up.

**Why it happened.** The evidence-publication candidate was reviewed before commit, but the permanent
forensic text captured the pre-review gate wording and was not updated to the resulting historical
approval before final exact-head External verification.

**Failure scenario.** If the dossier were merged unchanged, a future reader would see a canonical
closure record that claims to preserve the full material review/remediation chain while stopping the
evidence-publication chronology immediately before the approval that actually permitted publication.

**Impact.** Forensic chronology would be incomplete and one active sentence would be stale. Runtime,
CI identity, PR identity, closure semantics, audit arithmetic, and remaining-P3 scope were unaffected.

**Rejected approach.** Leave the approval only in the live PR body and rely on that volatile surface
while the permanent dossier claims complete material review history.

**Remediation.** Section 13 now permanently records the completed Evidence-Publication Internal
Re-Review, including `INT-EVIDENCE-P3-01 — RESOLVED`, P0/P1/P2/P3 = 0, verdict `APPROVED`, exact
approved patch `4cb3ad27e7709722fdd137c741b372c5dca7791fc67f21a3df77fd7aa0b88981`, resulting blob `fd7679f926e33897a4ad5377e9b14ded99116479`, and the fact that the approval
preceded evidence-only publication commit `1a7155eaf7c8195f11547e1e7f921dd32a896662`.

### Review-evidence boundary for this remediation

This dossier permanently records **material findings, their remediation, and completed review results
that authorized already-published repository evidence**. It deliberately does not recursively embed
the future verdict that will review this exact `EXT-FINAL-P3-01` remediation candidate, because doing
so would require a new repository commit whose only purpose is to record the review of the preceding
recording commit and would create unbounded self-reference.

The governed boundary is:

1. this remediation candidate must receive fresh Internal Review before publication;
2. after publication, its exact new head must pass required CI;
3. the live PR body must be synchronized to that actual head and CI;
4. final Internal and External exact-head verification results for that published remediation are
   recorded on the live PR evidence surface;
5. any **new material finding** from those verifications must be remediated and preserved before merge;
6. a no-new-finding `APPROVED` verification does not require another repository commit merely to embed
   its own verdict in the dossier;
7. separate explicit human Ready/merge authorization is still required.

This boundary preserves material forensic history without predicting a future commit SHA, CI run,
review verdict, or final squash-merge SHA.

## 15. Evidence-only follow-up publication invariant

This evidence publication is documentation-only. It changes no Stage 3.39 runtime behavior, OpenAPI
executable contract, schema, migration, dependency, workflow, authentication/security/privacy runtime,
replay/idempotency behavior, importer/parser behavior, financial arithmetic, or infrastructure.

No documentation/evidence follow-up is merge-authoritative merely because its text is committed.
Before any human Ready/merge authorization, the current published remediation head must pass required
CI, the live PR body must be synchronized to the actual head and CI evidence, and both Internal and
External reviewers must verify the exact published head. Any new material blocker must be remediated
and the affected gates repeated.

A no-new-finding final verification is intentionally retained as live PR evidence rather than forcing
another repository commit solely to record the verdict that reviewed the previous recording commit.
This prevents recursive evidence churn while preserving every material finding and remediation.

No future remediation commit SHA, CI run number, review verdict, or final closure squash-merge SHA is
predicted here. That keeps this permanent record publication-stable.

Until the approved closure record is actually present on protected `develop`, **P3-04 remains OPEN**.
Only protected closure activation changes it to CLOSED.
