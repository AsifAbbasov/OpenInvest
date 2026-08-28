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

## 10. Closure Internal Review remediation history

### 10.1 `INT-CLOSURE-P3-01` — active lifecycle fields would self-stale at closure activation

**Problem.** The first complete four-file closure candidate used active top-level values such as
`CLOSURE CANDIDATE`, `Finding status | OPEN`, and present-tense wording that the closure change was
`eligible for` future review/publication gates.

**Why it happened.** The first closure design correctly made the *decision rule* merge-activated, but
some metadata fields still encoded the pre-activation phase as a hard-coded current value.

**Failure scenario.** If that exact candidate were squash-merged into protected `develop`, the closure
record would immediately be canonical while still calling itself a candidate, and the implementation
dossier would still expose an active `Finding status | OPEN` even though the merge-activated rule said
P3-04 was CLOSED.

**Impact.** The four canonical surfaces would contradict one another at the exact moment intended to
make closure canonical. This is governance/evidence-integrity P3, not a runtime defect.

**Initial solution rejected by review.** Keep the merge-activated rule only in prose while leaving
pre-activation current-state metadata hard-coded.

**Revised solution.** Top-level lifecycle fields are phase-neutral and merge-activated:
- before the approved closure record is present on protected `develop` -> P3-04 OPEN;
- once present -> P3-04 CLOSED.
Section 9 now describes the required pre-activation workflow conditionally so the same text remains
true after activation.

**Why chosen.** It requires no future SHA/CI prediction and remains truthful across local preparation,
commit, push, Draft PR, CI, review, and protected merge.

**Residual limitation.** Protected activation is still a future governed action at candidate-review
time. This document does not claim that action has occurred until the record is actually present on
protected `develop`.

### 10.2 `INT-CLOSURE-P3-02` — moving `develop` HEAD was frozen as an active current baseline

**Problem.** The first closure candidate changed `SOURCE_OF_TRUTH.md` to
`Current canonical implementation baseline: develop at abbd9f9f61574621e206f2e196b1fb8f056dc194`.

**Why it happened.** `abbd9f9f61574621e206f2e196b1fb8f056dc194` is the real Stage 3.39 implementation squash merge and was the live
protected `develop` HEAD while the closure candidate was prepared. The wording incorrectly treated that
moving branch pointer as a permanent current-state fact.

**Failure scenario.** The future closure squash merge must advance `develop` to a new SHA. The newly
merged Source of Truth would therefore become stale immediately while still calling `abbd9f9f61574621e206f2e196b1fb8f056dc194` the
current `develop` baseline.

**Impact.** Canonical branch-state evidence would be false immediately after the closure action. This is
governance/evidence-integrity P3 only.

**Initial solution rejected by review.** Record the known implementation merge SHA as the current
protected-branch HEAD.

**Revised solution.** Record `abbd9f9f61574621e206f2e196b1fb8f056dc194` as the immutable **Stage 3.39 implementation merge baseline** and
explicitly state that the moving protected `develop` HEAD is not hard-coded by this merge-activated
closure record.

**Why chosen.** The implementation merge SHA is an immutable evidenced fact; the future closure merge
SHA does not yet exist and must not be predicted or represented by a placeholder.

**Regression/review evidence.** The revised complete four-file candidate must regenerate from exact
base `abbd9f9f61574621e206f2e196b1fb8f056dc194`, preserve exactly four documentation/governance paths, pass `git diff --check`, and receive
fresh Internal Closure Re-Review before any commit/push authorization request.

**Residual limitation.** The final closure squash SHA can only be recorded as a historical immutable
fact after it actually exists; no current candidate fabricates it.
