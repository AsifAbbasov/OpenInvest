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


Both final published-head verification roles independently returned `APPROVED` on exact head
`26f5ca18ca5772db569d22ce2eff64d5a7850b1b` with P0/P1/P2/P3 = None.

The final Internal verification confirmed exact pre-commit-to-published equivalence, 765/765 Stage-report

The final External verification independently confirmed the published identity, CI #279, the complete
forensic history, all prior documentation/evidence findings resolved, and no runtime/governance blocker.

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


evidence and are not retroactively reclassified.

Until this closure-governance record is actually present on protected `develop`, **P3-04 remains OPEN**.

When it is present on protected `develop`, **P3-04 is CLOSED** and the original 32-finding audit becomes:

- closed: 27 / 32;
- completion: 84.375%;
- remaining: 5 / 32;
- remaining findings: P3-06, P3-07, P3-08, P3-09, P3-10.

## 11. Blind-review remediation history

### 11.2 Fresh Internal Clean Blind-Closure Review


- exact clean four-file patch SHA256:
  `d55dcc3aa8daac04aebd7c398a7e97d178add4c87f2c9f14dbb5106d6db92030`;
- proposed closure dossier blob:
  `0d919cfe6a637f5eea8687a374101ff6926d1ba9`;
- required base: `abbd9f9f61574621e206f2e196b1fb8f056dc194`;
- `INT-BLIND-P3-01` — RESOLVED;
- `INT-BLIND-P3-02` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`.


`618927a4a336691059621b120696d0895afd136d` — `docs: prepare blind Stage 3.39 P3-04 closure`

with exact parent `abbd9f9f61574621e206f2e196b1fb8f056dc194`.

Draft PR #99 was created from that clean head. CI #281 / run `33157801196` completed on exact
head `618927a4a336691059621b120696d0895afd136d` with all 10 required jobs successful.

## 14. Final exact-head External verification remediation

Final External exact-head verification was performed against Draft PR #99 at published head
`1a7155eaf7c8195f11547e1e7f921dd32a896662` after CI #282 / run `33177434968` completed 10/10 successfully.

protected `develop` at `abbd9f9f61574621e206f2e196b1fb8f056dc194`, current dossier blob `fd7679f926e33897a4ad5377e9b14ded99116479`, PR-body synchronization,
closure activation semantics, backlog arithmetic, and no runtime impact.

The verdict was `changes required` for one documentation / forensic-completeness P3 only:

### `EXT-FINAL-P3-01` — permanent dossier omitted the completed Evidence-Publication internal validation

**Problem.** Section 13 recorded the `INT-EVIDENCE-P3-01` remediation but still ended with
“Fresh internal validation is required...” and did not preserve the already-completed re-review that
authorized publication of the evidence-only follow-up.


**Impact.** Forensic chronology would be incomplete and one active sentence would be stale. Runtime,
CI identity, PR identity, closure semantics, audit arithmetic, and remaining-P3 scope were unaffected.


**Remediation.** Section 13 now permanently records the completed Evidence-Publication Internal
approved patch `4cb3ad27e7709722fdd137c741b372c5dca7791fc67f21a3df77fd7aa0b88981`, resulting blob `fd7679f926e33897a4ad5377e9b14ded99116479`, and the fact that the approval
preceded evidence-only publication commit `1a7155eaf7c8195f11547e1e7f921dd32a896662`.

### Review-evidence boundary for this remediation

This dossier permanently records **material findings, their remediation, and completed review results


The governed boundary is:

2. after publication, its exact new head must pass required CI;
3. the live PR body must be synchronized to that actual head and CI;
5. any **new material finding** must be remediated and preserved before merge;
6. a no-new-finding `APPROVED` verification does not require another repository commit merely to embed
   its own verdict in the dossier;
7. separate explicit human Ready/merge authorization is still required.

This boundary preserves material forensic history without requiring separate review chats and without


### `GOV-SYNC-P3-01` — stale §9 dual-review requirement


**Remediation.** §9 is now synchronized to the one-chat governance/closure path while preserving the
earlier Internal/External events as historical evidence.

**Result.** RESOLVED.

### `GOV-SYNC-P3-02` — proposed workflow appeared to authorize its own adoption

**Problem.** An earlier candidate relied on the proposed v1.3 text itself to justify a changed review
topology before v1.3 was canonical.

**Remediation.** The current transition is grounded in the explicit Principal Architect process
Ready, merge, protected-branch mutation, or finding suppression. The synchronized workflow still
becomes canonical only after protected merge.

**Result.** RESOLVED in this candidate.

### `INT-TRANSITION-P3-01` — in-flight PR transition timing ambiguity

**Problem.** An earlier workflow candidate did not state which policy governs review actions for PRs
already open but unmerged when a new workflow version becomes effective.

**Remediation.** Review actions performed after a new workflow version becomes effective use that
effective version according to the PR's actual current scope, including already-open but unmerged PRs.
Completed earlier review events retain their historical meaning and are not retroactively reclassified.

**Result.** RESOLVED in this candidate.

### `GOV-CHAT-P3-01` — one-chat rewrite weakened substantive development evidence controls


**Why it happened.** Chat topology and substantive evidence controls were changed in the same rewrite,
and single-context phase independence was treated as if it were automatically equivalent to the full
existing evidence chain.

**Failure scenario.** A future development PR could expose its Internal verdict/findings before the
External conclusion, allow that earlier conclusion to act as supporting evidence, and merge without

**Impact.** Governance/evidence integrity only. Runtime, API, schema, migrations, dependencies,
workflows, security/privacy runtime, financial semantics, and Stage 3.39 runtime behavior are
unaffected.

**Rejected approach.** Keep only the instruction that the External phase “reviews from scratch” while
dropping the explicit withholding/publication/verification controls.

**Remediation.** The one-chat model now preserves equivalent substantive controls for development:
Internal evidence is withheld from the Draft PR/repository evidence surface until the External verdict;
the required Internal evidence in an evidence-only follow-up; CI runs on that head; and the same
no-new-finding evidence-verification verdict does not create recursive commit churn.

**Additional Principal Architect resolution.** The remaining anti-anchoring mismatch identified on
re-review is accepted and made explicit: single-context visibility is a deliberate replacement for strict
blind information-isolation. The dossier and repository workflow no longer claim substantive equivalence
on that specific property. Evidentiary non-reliance, PR/repository withholding, post-External Internal
mandatory.


- `GOV-SYNC-P3-01` — RESOLVED;
- `GOV-SYNC-P3-02` — RESOLVED;
- `INT-TRANSITION-P3-01` — RESOLVED;
- `GOV-CHAT-P3-01` — RESOLVED;
- P0/P1/P2/P3 = 0;
- verdict: `APPROVED`;
- exact approved complete patch SHA256:
  `8fb1ca4dc6b6b360c071630104234a4bf594a5a6fcd50d1847655b72672670c7`;
- exact proposed closure dossier blob:
  `cdb74707d1135ddf2ae9cc9c69d844230148ac71`.

That exact approved candidate was subsequently published as commit
`e14a21e05daa4f7e6d4fdc477dff34a51b0fc51e` (`docs: synchronize one-chat Stage 3.39 governance`).
The prepublication approval did not authorize Ready/merge and did not close P3-04.


**Problem.** Final exact-published-head Governance / Closure verification of
`e14a21e05daa4f7e6d4fdc477dff34a51b0fc51e` found that the published dossier still said fresh


**Impact.** Governance and forensic chronology would be internally contradictory. Runtime behavior,
API contracts, schema/migrations, dependencies, CI workflow implementation, security/privacy runtime,
financial semantics, Stage 3.39 Unicode behavior, closure activation semantics, and audit arithmetic
were unaffected.

**Rejected approach.** Treat the stale sentence as harmless because the live PR body and review package
contain the correct chronology. The closure dossier itself is intended to be the durable forensic
record, so contradictory active wording cannot be delegated to a volatile evidence surface.

Canonical record: commit(s) `e14a21e05daa4f7e6d4fdc477dff34a51b0fc51e`.


## 16. Evidence-only follow-up publication invariant

This evidence publication is documentation-only. It changes no Stage 3.39 runtime behavior, OpenAPI
executable contract, schema, migration, dependency, workflow, authentication/security/privacy runtime,
replay/idempotency behavior, importer/parser behavior, financial arithmetic, or infrastructure.

No documentation/evidence follow-up is merge-authoritative merely because its text is committed.
Before any human Ready/merge authorization, the current published remediation head must pass required
CI, the live PR body must be synchronized to the actual head and CI evidence, and the designated
remediated and the affected gates repeated.

A no-new-finding final verification is intentionally retained as live PR evidence rather than forcing
This prevents recursive evidence churn while preserving every material finding and remediation.

predicted here. That keeps this permanent record publication-stable.

Until the approved closure record is actually present on protected `develop`, **P3-04 remains OPEN**.
Only protected closure activation changes it to CLOSED.
