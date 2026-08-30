# Stage 3.44 — P3-10 Fiber Maintenance Implementation

| Field | Value |
| --- | --- |
| Date | 2026-08-30 |
| Exact implementation base | `develop@eaac5a5deb64196b263464e0d85e622065520b0e` |
| Finding | Original audit `P3-10 — Fiber maintenance` |
| Planning authority | Merged Stage 3.43 plan blob `37a32856692ac58f408f2dd50335bb65019d9983` |
| Runtime closure state | P3-10 remains OPEN |
| Current audit state | 28 / 32 = 87.5%; remaining P3-06, P3-07, P3-08, P3-10 |
| Internal Review Evidence | WITHHELD — external published-head phase pending |

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

The Internal Review package carries exact `go list -m all` baseline/target evidence.

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
success banner is explicitly rejected and its Internal Review package must not be used.

Runner v4 attempted to repair v3 but exited non-zero with code 16 before fetch/candidate inspection
because it incorrectly required the linked worktree's `.git` marker to be a directory. In a Git
linked worktree `.git` is normally a file. v4 therefore produced no candidate mutation.

Runner v5 corrects that precondition by asking Git itself whether both paths are valid worktrees,
then preserves the corrupted v3 dossier externally before replacement, independently reconstructs
baseline evidence, repeats target gates, and generates this dossier using quoted templates plus
explicit placeholder substitution.

### Builder anti-regression preflight after v5 package

Before the candidate was sent to Internal review, Builder package inspection found two
evidence-integrity defects. Neither changed runtime/dependency/test behavior.

1. `BUILDER-PREFLIGHT-STAGE-03-44-01` — the v5 derived module comparator treated the
   unversioned root module `github.com/openinvest/openinvest/backend-go` as both `<removed>`
   and `<new>`, even though exact baseline/target `go list -m all` snapshots contain the same
   root line. v6 replaced the comparator with module-path-keyed comparison and removed those
   two false rows.
2. `BUILDER-PREFLIGHT-STAGE-03-44-02` — v5 required Internal review to verify v1/v2 runner
   chronology but did not carry self-contained operator-console provenance for those runs.
   v6 adds explicitly sourced operator-console provenance excerpts for v1-v4 alongside the derived notes.

Both defects were repaired before Internal review. They are evidence/governance defects, not
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

All command logs and exit codes are carried in the Internal Review package.

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

This candidate requires complete Internal line-by-line READ-ONLY review before any commit/push
authorization.

Internal evidence remains withheld from the Draft PR/repository evidence surface until the External
verdict, per canonical `REVIEW_WORKFLOW.md` v1.3.0.

No review verdict by itself authorizes commit, push, remote branch publication, PR mutation, Ready,
merge, or closure.

## 10. Closure semantics

P3-10 remains OPEN after this implementation candidate and would remain OPEN even after a future
implementation merge.

Only a separate reviewed docs-only closure-governance activation on protected `develop` may close
P3-10.

If no other finding changes concurrently, eventual closure would produce 29 / 32 = 90.625%, leaving
exactly P3-06, P3-07 and P3-08.

No future PR number, CI run, published head or merge SHA is predicted here.
