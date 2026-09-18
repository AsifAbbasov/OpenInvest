# Stage 3.44 — P3-10 Fiber Maintenance Implementation

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION MERGED — PR #104 squash-merged into protected `develop` at `c980a21f16b30449ec7fb7b07decc386d77bc27d`; P3-10 closure is merge-activated under Stage 3.45 |
| Date | 2026-08-30 |
| Exact implementation base | `develop@eaac5a5deb64196b263464e0d85e622065520b0e` |
| Finding | Original audit `P3-10 — Fiber maintenance` |
| Planning authority | Merged Stage 3.43 plan blob `37a32856692ac58f408f2dd50335bb65019d9983` |
| Final evidence-publication head | `749446930a730c3f7c5b3402618577953d53d3f4` |
| Implementation squash merge | `c980a21f16b30449ec7fb7b07decc386d77bc27d` |
| Final exact-head CI | CI #297 / run `33312723075`, 10/10 required jobs successful |
| P3-10 lifecycle | Before the approved Stage 3.45 closure record and synchronized canonical surfaces are present on protected `develop`, P3-10 is OPEN; once they are present, P3-10 is CLOSED |
| Closure state at implementation publication | P3-10 OPEN |
| Audit state at implementation publication | 28 / 32 = 87.5%; remaining P3-06, P3-07, P3-08, P3-10 |

## 1. Problem and root cause

The repository was pinned to Fiber `v3.3.0`. Stage 3.43 established the narrow maintenance target
`v3.5.0`, including the already-known Fiber-attributable shared-direct selected-module movement
`golang.org/x/crypto v0.51.0 -> v0.54.0`.

OpenInvest directly uses `golang.org/x/crypto/argon2`, so the cryptographic module movement is part
of the reviewed compatibility surface rather than an incidental transitive update.

## 2. Exact baseline

- protected base: `eaac5a5deb64196b263464e0d85e622065520b0e`;
- `backend-go/go.mod` baseline blob: `9342530d300392173c9ebeca0ce42a8ec9d86459`;
- `backend-go/go.sum` baseline blob: `1a206517eae6e65a91fabbfa8c4ba440cf8b526a`;
- `backend-go/internal/auth/password.go` unchanged baseline blob: `9f6f4b69a6ff057bb9ba167ee6727dac3706ade0`;
- Fiber selected baseline: `v3.3.0`;
- x/crypto selected baseline: `v0.51.0`;
- repository Go directive: `1.25.14`;
- reproducible gate toolchain: `go1.25.14`.

## 3. Implemented change

- Fiber selected version: `v3.3.0 -> v3.5.0`;
- x/crypto selected version: `v0.51.0 -> v0.54.0`;
- root x/crypto declaration after resolution: `v0.54.0`;
- Go directive remains: `1.25.14`;
- unrelated direct selections remain:
  - `github.com/google/uuid v1.6.0`;
  - `github.com/jackc/pgx/v5 v5.9.2`;
  - `gopkg.in/yaml.v3 v3.0.1`;
- production application source was not changed;
- one test-only compatibility regression was added:
  `backend-go/internal/auth/password_dependency_compat_test.go`.

## 4. Resolved module graph drift


The comparator is keyed by module path and treats an unversioned repository root module as
unchanged when that same unversioned root appears in both snapshots.

| Module | Before | After |
| --- | --- | --- |
| `github.com/andybalholm/brotli` | `v1.2.1` | `v1.2.2` |
| `github.com/gofiber/fiber/v3` | `v3.3.0` | `v3.5.0` |
| `github.com/gofiber/schema` | `v1.7.1` | `v1.8.3` |
| `github.com/gofiber/utils/v2` | `v2.0.6` | `v2.4.1` |
| `github.com/klauspost/compress` | `v1.18.6` | `v1.19.2` |
| `github.com/mattn/go-colorable` | `v0.1.14` | `v0.1.15` |
| `github.com/mattn/go-isatty` | `v0.0.22` | `v0.0.24` |
| `github.com/shamaton/msgpack/v3` | `v3.1.2` | `v3.2.0` |
| `github.com/valyala/fasthttp` | `v1.71.0` | `v1.73.0` |
| `golang.org/x/crypto` | `v0.51.0` | `v0.54.0` |
| `golang.org/x/net` | `v0.55.0` | `v0.57.0` |
| `golang.org/x/sync` | `v0.21.0` | `v0.22.0` |
| `golang.org/x/sys` | `v0.45.0` | `v0.47.0` |
| `golang.org/x/term` | `v0.43.0` | `v0.45.0` |
| `golang.org/x/text` | `v0.39.0` | `v0.40.0` |

## 5. Argon2 historical compatibility

The exact fixed Argon2id fixture is executed against both:

1. a temporary exact-base tree using baseline x/crypto `v0.51.0`;
2. the target candidate using selected x/crypto `v0.54.0`.

Both v5 reproductions passed.

The regression also proves:

- the historical hash remains verifiable;
- a wrong password remains rejected;
- memory remains 65536 KiB;
- time cost remains 3;
- parallelism remains 1;
- salt length remains 16 bytes;
- key length remains 32 bytes;
- encoded format remains `argon2id$v=19$m=65536,t=3,p=1$...`;
- newly generated hashes still round-trip.

This is compatibility evidence for dependency maintenance, not evidence of a pre-existing auth defect.

## 6. Tooling chronology

Runner v1 exited non-zero with code 23 because it incorrectly equated the ambient developer-machine
Go version with the repository-pinned toolchain policy. No dependency mutation had started.

Runner v2 successfully downloaded/executed pinned `go1.25.14`, but its parser merged stderr/stdout,
produced a duplicated version value, and exited non-zero with code 23 before dependency mutation.

Runner v3 completed the substantive dependency and test gates, but artifact generation was invalid:
an unquoted heredoc interpreted Markdown backticks as shell command substitutions. The console emitted
tooling errors after the substantive gates. v3 nevertheless printed `CANDIDATE SUCCESS`; that overall

Runner v4 attempted to repair v3 but exited non-zero with code 16 before fetch/candidate inspection
because it incorrectly required the linked worktree's `.git` marker to be a directory. In a Git
linked worktree `.git` is normally a file. v4 therefore produced no candidate mutation.

Runner v5 corrects that precondition by asking Git itself whether both paths are valid worktrees,
then preserves the corrupted v3 dossier externally before replacement, independently reconstructs
baseline evidence, repeats target gates, and generates this dossier using quoted templates plus
explicit placeholder substitution.


evidence-integrity defects. Neither changed runtime/dependency/test behavior.

   unversioned root module `github.com/openinvest/openinvest/backend-go` as both `<removed>`
   and `<new>`, even though exact baseline/target `go list -m all` snapshots contain the same
   root line. v6 replaced the comparator with module-path-keyed comparison and removed those
   two false rows.
   chronology but did not carry self-contained operator-console provenance for those runs.
   v6 adds explicitly sourced operator-console provenance excerpts for v1-v4 alongside the derived notes.

P3-10 runtime defects.

## 7. Local quality evidence

v5 PASS gates:

- independently reconstructed baseline historical Argon2 compatibility;
- target `go mod tidy` idempotence;
- exact target module selections;
- `go mod verify`;
- independently reproduced target historical Argon2 compatibility;
- `go test -race ./internal/auth`;
- `go test ./...`;
- `go vet ./...`;
- pinned `govulncheck@v1.7.0 ./...`;
- `git diff --check`;
- exact four-file scope accounting including untracked files;
- generated-dossier content validation.


## 8. Scope isolation

Candidate surface is exactly four files:

1. `backend-go/go.mod`;
2. `backend-go/go.sum`;
3. `backend-go/internal/auth/password_dependency_compat_test.go`;
4. `docs/stages/STAGE_03_44_FIBER_MAINTENANCE_IMPLEMENTATION.md`.

P3-06, P3-07 and P3-08 remain outside scope. No `httpapi/api.go` decomposition, transaction-form
fixture/default work, migration-validator hardening, frontend change, OpenAPI change, or database
schema change is included.

## 9. Review and publication lifecycle

`PLAN-STAGE-03-43-P3-01` remains append-only history and its required shared-direct dependency/
Argon2 compatibility controls are satisfied by this candidate evidence.

and Draft PR publication.

merge, or closure.


candidate.

Reviewed candidate identity at the Internal phase:

- `backend-go/go.mod` blob
  `a6be6f2266a428b52206ee3890c7d5b199dab97b`;
- `backend-go/go.sum` blob
  `bdd8e6edfd45a2725fb1f8dc0831d45ae9f39cd0`;
- `backend-go/internal/auth/password_dependency_compat_test.go` blob
  `ef1452d56fa8666ce3a260d4ba0b2bdc29236856`;
- Stage 3.44 dossier blob at that phase
  `475133892755d564be2b792520eaf8bbc289c10e`.

The Internal report concluded:

- line-by-line coverage: `COMPLETE`;
- exact protected base: `MATCHES`;
- candidate subject: `COMPLETE`;
- Fiber target: `VERIFIED`;
- x/crypto shared-direct movement: `VERIFIED`;
- resolved module graph: `PASS`;
- unversioned root-module handling: `PASS`;
- historical Argon2 compatibility: `PASS`;
- Argon2 parameters/format: `UNCHANGED`;
- unrelated direct dependency drift: `NONE`;
- runtime-source change: `NONE`;
- local gates: `PASS`;
- scope accounting: `PASS`;
- Runner v1 chronology: `CORRECT`;
- Runner v2 chronology: `CORRECT`;
- Runner v3 chronology: `CORRECT`;
- Runner v4 chronology: `CORRECT`;
- `PLAN-STAGE-03-43-P3-01`: `SATISFIED`;
- workflow/governance: `PASS`;
- publication stability: `PASS`;
- audit lifecycle/arithmetic: `PASS`;
- P0 = 0;
- P1 = 0;
- P2 = 0;
- P3 = 0;
- new material finding: `NO`;
- verdict: `APPROVED`.

The later External remediation changed only the Stage 3.44 dossier. The three non-dossier
implementation blobs above remained byte-identical and were re-verified on the final External
remediation head.

## 10. Closure semantics

Implementation merge does not itself activate P3-10 closure. Under the canonical workflow, P3-10
`develop`.

Given the 28 / 32 audit baseline recorded by this implementation stage, a closure activation with no
concurrent finding changes corresponds to 29 / 32 = 90.625%, leaving exactly P3-06, P3-07 and P3-08.

No future PR number, CI run, published head or merge SHA is predicted here.

## 11. Governance / closure activation

Stage 3.44 development-path work is complete and merged.

Immutable post-evidence history:

1. final evidence-publication head
   `749446930a730c3f7c5b3402618577953d53d3f4`;
2. final implementation dossier blob before squash merge
   `04ffc9c059871147a15d3eb6627f05268c686ed2`;
3. CI #297 / run `33312723075` — `10/10 SUCCESS`;
4. single-context exact evidence-publication verification:
   - evidence publication `COMPLETE AND ACCURATE`;
   - evidence-only scope `CONFIRMED`;
   - runtime/dependency semantic drift `NONE`;
   - exact protected base `MATCHES`;
   - exact evidence-publication head `MATCHES`;
   - published dossier blob `MATCHES`;
   - frozen runtime/dependency/test blobs `MATCH`;
   - Internal evidence chronology `CORRECT`;
   - External history `CORRECT`;
   - PR metadata `CONSISTENT`;
   - exact-head CI #297 `10/10 SUCCESS`;
   - publication stability `PASS`;
   - audit lifecycle/arithmetic `PASS`;
   - P0/P1/P2/P3 = 0;
   - new material finding `NO`;
   - verdict `APPROVED`;
5. explicit merge gate for the Ready transition and squash merge of exact head
   `749446930a730c3f7c5b3402618577953d53d3f4`;
6. PR #104 was in Ready state before merge;
7. actual PR #104 squash merge
   `c980a21f16b30449ec7fb7b07decc386d77bc27d`;
8. protected `develop` was read back at that exact implementation merge SHA.

The implementation merge is an immutable repository fact. The moving protected `develop` HEAD is not
represented as permanently equal to that implementation SHA.

Original audit P3-10 closure is intentionally separate from the implementation merge. Stage 3.45 is
documentation/governance-only closure activation:

1. before the approved Stage 3.45 closure record and synchronized canonical surfaces are present on
   protected `develop`, P3-10 remains OPEN;
2. once they are present on protected `develop`, P3-10 is CLOSED;
3. Stage 3.45 must not change the accepted runtime/dependency/test identities;
4. no future Stage 3.45 PR number, published head, CI run, or squash-merge SHA is predicted here.

classified `NO MATERIAL PROJECT FINDING` are not erased or softened.

No statement in this dossier authorizes Stage 3.45 commit, push, Draft PR creation, Ready, merge,
branch deletion, or protected-branch mutation.
