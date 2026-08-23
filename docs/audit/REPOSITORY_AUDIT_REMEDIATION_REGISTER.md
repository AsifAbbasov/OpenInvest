# Repository Audit Remediation Register

| Field | Value |
| --- | --- |
| Document ID | REG-AUDIT-REM-001 |
| Status | Current audit-remediation index |
| Scope | Original 32-finding OpenInvest repository audit |
| Canonical evidence baseline | Stage 3.33 closure on `develop` at `71a1faeb97d33d05f2936111b53f1285edddabe9` |
| Documentation standard | `REMEDIATION_DOCUMENTATION_STANDARD.md` |
| Owner | Principal Architect |

## Purpose

This register is the single cross-stage engineering index for the original OpenInvest repository audit. It preserves what was wrong, why it mattered, how the problem could affect the project, how it was remediated when closed, why the selected remediation was preferred, where regression evidence lives, and what remains open.

It does not replace the detailed stage dossiers. Detailed causal analysis, review iterations, rejected alternatives, regression fixtures, CI evidence, residual risks, and deployment consequences remain in the linked `docs/stages/` documents.

Beginning with Stage 3.34, every remediation dossier must satisfy `docs/audit/REMEDIATION_DOCUMENTATION_STANDARD.md` before canonical closure.

## Audit totals

Original audit:

- P0: 0
- P1: 5
- P2: 17
- P3: 10
- Total: 32

After canonical Stage 3.33 closure:

- P0 open: 0
- P1 open: 0
- P2 open: 2
- P3 open: 10
- Total open: 12
- Closed: 20 / 32 = 62.5%
- P1 + P2 closed: 20 / 22 = 90.9%

The only remaining P2 findings are **P2-16** and **P2-17**.

## Executive register

| Finding | Problem summary | Impact if triggered | Remediation / direction | Evidence | Status |
| --- | --- | --- | --- | --- | --- |
| P1-01 | Refresh-token replay did not reliably invalidate the whole session family. | A stolen/replayed refresh token could leave sibling sessions usable and weaken replay containment. | Persisted session-family identity, serialized refresh/logout mutation, conservative legacy containment, family revocation on replay. | Stage 3.28 / PR #59 / closure PR #60 | CLOSED |
| P1-02 | `broker_operation_id` could be lost before the canonical ledger boundary. | Strong broker identity could disappear, weakening duplicate/conflict guarantees and allowing economically identical operations to coexist incorrectly. | Versioned persisted import identity plus fallback/strong identity conflict prevention. | Stage 3.27 / PR #55 / closure PR #58 | CLOSED |
| P1-03 | Cash-flow near-match classification omitted amount. | Different cash movements on the same date could be classified as duplicates/conflicts. | Amount-aware cash reconciliation while preserving broker identity precedence. | Stage 3.27 | CLOSED |
| P1-04 | DEPOSIT/WITHDRAWAL accepted fee/tax values that snapshot economics ignored. | API could accept economic values that silently had no financial effect, producing misleading ledger/snapshot semantics. | Fail-closed zero-fee semantics for cash flows until fee economics are explicitly modeled. | Stage 3.27 | CLOSED |
| P1-05 | Public Argon2 work had no process-wide admission bound. | Login/register bursts could create multi-gigabyte memory pressure and availability failure. | Keep approved Argon2id cost; add fail-fast shared capacity gate and reject over-budget stored encodings. | Stage 3.28 | CLOSED |
| P2-01 | Logout rejection path could generate unbounded database audit writes. | Abuse could convert cheap requests into persistent DB write amplification. | Route logout through auth admission before rejected-auth persistence. | Stage 3.31 / PR #65 / closure #66 | CLOSED |
| P2-02 | Import review token did not freeze normalized financial meaning. | A reviewed file could be appended under semantically changed parser/normalization behavior. | Versioned signed review token bound to normalized parser/review semantics and exact payload context. | Stage 3.30 / PR #63 / closure #64 | CLOSED |
| P2-03 | 100-row limit was enforced only after full parsing. | Oversized input could consume parsing/memory work before rejection. | Parser-owned fail-fast admission on the 101st data record. | Stage 3.30 | CLOSED |
| P2-04 | Import review inspected only the latest 100 existing transactions. | Older relevant duplicates/conflicts could be missed. | Targeted full-history PostgreSQL query over relevant dates and privacy-minimized identity keys. | Stage 3.30 | CLOSED |
| P2-05 | Invalid decimal input could escape validation and produce HTTP 500. | User input error became server-failure semantics and could expose inconsistent contract handling. | Deterministic decimal validation/error mapping at command boundaries. | Stage 3.29 / PR #61 / closure #62 | CLOSED |
| P2-06 | Note length could exceed the stored contract and fail at persistence. | Valid-looking API input could become a database/runtime error; Unicode semantics could diverge. | Application/import boundary enforcement of the 500-character stored-note contract. | Stage 3.29 | CLOSED |
| P2-07 | Accepted decimal magnitudes/derived snapshot values could exceed PostgreSQL `NUMERIC(28,8)`. | Financial commands could fail late or overflow derived persistence after partial work if not guarded atomically. | Align ingress magnitude with persistence bounds and preflight all persistence-bound snapshot metrics in the same transaction. | Stage 3.29; first review REQUEST CHANGES then remediation | CLOSED |
| P2-08 | Financial JSON commands did not uniformly reject unknown properties. | Misspelled or stale client fields could be silently ignored, making financial intent ambiguous. | Strict JSON decoding / `additionalProperties:false` behavior for financial writes. | Stage 3.29 | CLOSED |
| P2-09 | Idempotent duplicate handling did not replay the exact original HTTP response. | Retry could observe mutable current resource state instead of the response produced by the original committed command. | Persist exact versioned HTTP artifact atomically with financial mutation; replay stored status/body/request/trace identity. | Stage 3.32 / PR #67 / closure #68 | CLOSED |
| P2-10 | `snapshotDatesRebuilt` was derived from input dates rather than actual database rebuild work. | API could report an incomplete set of affected snapshots after backdated imports. | PostgreSQL owns deterministic affected-date union and returns exact rebuild outcome into replay artifact. | Stage 3.33 / PR #69 / closure #70 | CLOSED |
| P2-11 | Backdated batch imports repeatedly cascaded rebuild of the same later snapshot. | Unnecessary repeated projection work and inflated snapshot versions within one command. | Compute one sorted unique affected-date plan and rebuild every date exactly once. | Stage 3.33 | CLOSED |
| P2-12 | Append-only ledger immutability was not enforced by the authenticated PostgreSQL runtime privilege graph. | Runtime credentials could potentially mutate protected ledger/audit data or regain mutation authority through role escalation. | Dedicated least-privilege role plus same-connection session/current identity validation, SET-reachable graph checks, owner/schema/mutation checks, and ADMIN OPTION rejection. | Stage 3.33; two REQUEST CHANGES cycles; final APPROVED | CLOSED |
| P2-13 | Browser idempotency retry identity was lost across reload/remount and initially not principal-isolated. | A failed command retry could get a new idempotency key; later design could collide between users in one tab. | Short-lived sessionStorage retry journal keyed by stable-principal-scoped hashed slot; no raw financial/auth payload persistence. | Stage 3.32; first review REQUEST CHANGES then remediation | CLOSED |
| P2-14 | Authentication limiter key map had unbounded lifecycle/cardinality. | Rotating attacker keys could grow memory indefinitely and bypass intended bounded-resource behavior. | Finite per-key attempts, global downstream budget, bucket cardinality cap, expired-bucket reclamation. | Stage 3.31 | CLOSED |
| P2-15 | Duplicate CSV headers were not rejected after normalization. | Ambiguous columns could overwrite/alias financial fields and change import meaning. | Fail closed on duplicate normalized headers before row normalization. | Stage 3.29 | CLOSED |
| P2-16 | `develop` has no mechanical GitHub governance enforcement; merge/rebase/squash are all enabled. | Direct/bypass paths can evade PR review, required checks, conversation resolution, linear history, or force-push/deletion controls. | Stage 3.34 plan: enforce protected `develop`, required checks/PRs, admin/owner coverage, no force/delete, linear history, squash-only. If account/plan cannot enforce owner/admin path, finding remains OPEN. | Stage 3.34 planning PR #71 approved; not yet merged or implemented | OPEN |
| P2-17 | CI lacks the remaining race/static/vulnerability/dependency/scheduled-security class. | Concurrency defects, known vulnerable reachable Go code, dependency risk, or security regressions can merge without dedicated gates. | Stage 3.34 plan: add vet, PostgreSQL-backed race suite, pinned govulncheck, Go/pnpm/Python dependency security coverage, nightly checks; preserve existing concurrency. | Stage 3.34 planning PR #71 approved; implementation not started | OPEN |
| P3-01 | Password length semantics are byte-oriented rather than an explicit user-facing character/grapheme contract. | Unicode passwords can hit limits differently from user expectation and contract wording. | Not yet remediated. Dedicated P3 planning must select and document canonical password-length semantics without weakening entropy/security limits. | Future P3 stage | OPEN |
| P3-02 | Timezone validation is not true IANA timezone validation. | Invalid or unsupported timezone identifiers can pass shallow validation and fail later or produce inconsistent date presentation. | Not yet remediated. Expected direction: canonical IANA lookup/validation at the owning boundary, subject to dedicated stage review. | Future P3 stage | OPEN |
| P3-03 | OpenAPI Decimal grammar and Go decimal parser acceptance are not perfectly aligned. | Client-generated contract and server parser can disagree on values considered valid. | Not yet remediated. Must choose one canonical grammar and prove OpenAPI/parser equivalence with contract tests. | Future P3 stage | OPEN |
| P3-04 | Some `maxLength` checks use byte semantics where contracts read as character limits. | Unicode input can be rejected/accepted inconsistently across API, Go, UI, and persistence boundaries. | Not yet remediated. Must define Unicode unit per field and align validators/tests. | Future P3 stage | OPEN |
| P3-05 | Idempotency/session persistence lacks a complete cleanup/retention lifecycle. | Technical records can accumulate indefinitely, creating storage/operational debt and unclear lifecycle guarantees. | Not yet remediated. Future stage must define retention, safe cleanup ownership, race behavior, and evidence without breaking exact replay/session guarantees. | Future P3 stage | OPEN |
| P3-06 | `httpapi/api.go` has grown into a god-module. | High coupling increases change risk, review difficulty, and merge-conflict/maintenance cost. | Not yet remediated. Future refactor must preserve API behavior while decomposing by bounded responsibility; no behavior change should be hidden in structural cleanup. | Future P3 stage | OPEN |
| P3-07 | Transaction form contains fixture-like defaults. | Demo-looking values can be mistaken for real user intent and weaken production UX trust. | Not yet remediated. Future UI cleanup must remove misleading defaults while preserving explicit validation and accessibility. | Future P3 stage | OPEN |
| P3-08 | Migration validator is weaker than the migration policy it claims to enforce. | Unsafe/nonconforming migration shapes may pass tooling despite governance documentation promising stronger checks. | Not yet remediated. Future stage must reconcile stated policy with executable validator coverage and mutation tests. | Future P3 stage | OPEN |
| P3-09 | Next.js dependency is behind the desired maintained version. | Missed bug/security/performance fixes and increasing future upgrade gap. | Not yet remediated. Upgrade must be separately reviewed with lockfile/build/test evidence and no architecture drift. | Future dependency-maintenance stage | OPEN |
| P3-10 | Fiber dependency is behind the desired maintained version; the originally checked CVE was not active for the pinned code path/version. | Maintenance/security debt grows even without a currently proven exploitable CVE. | Not yet remediated. Upgrade must be evidence-driven with Go tests/OpenAPI/runtime regression verification rather than urgency claims unsupported by an active vulnerability. | Future dependency-maintenance stage | OPEN |

## Detailed closed-finding rationale

### P1-01 — refresh replay family containment

**Root cause.** Session records did not provide a sufficiently strong persisted family boundary for replay response, especially for legacy rows and concurrent refresh/logout behavior.

**Consequence.** Token rotation alone is insufficient if replay of an older refresh token does not revoke every token descended from the same authentication family. A stolen refresh credential could therefore retain useful sibling access.

**Chosen method.** Stage 3.28 introduced persisted family identity for new sessions, conservative handling of legacy sessions, serialization of refresh/logout mutations, replay-triggered family revocation, and PostgreSQL defense in depth.

**Why selected.** It contains the security event at the canonical persistence boundary instead of relying on process memory, browser behavior, or timing assumptions. It preserves the existing auth architecture and does not require Redis or a new service.

**Detailed dossier.** `../stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md`.

### P1-02 / P1-03 / P1-04 — import financial identity and cash semantics

**Root cause.** The import pipeline had inconsistent ownership of broker identity, cash duplicate identity, and fee/tax semantics across parse/review/append/snapshot layers.

**Consequence.** Strong source identity could be dropped; economically different same-date cash movements could conflict; cash-flow fee fields could be accepted but ignored by financial calculations.

**Chosen method.** Persist versioned import identity, include amount in cash reconciliation, reject mixed weak/strong identity ambiguity, and fail closed to zero fees/taxes for deposit/withdrawal until a complete economic model exists.

**Why selected.** The solution removes ambiguity instead of guessing economics. It makes identity durable at the canonical ledger boundary and keeps unmodeled financial effects impossible rather than silently ignored.

**Detailed dossier.** `../stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md`.

### P1-05 — Argon2 resource admission

**Root cause.** Password hashing cost was individually correct but unconstrained at the process admission layer.

**Consequence.** Concurrent unauthenticated work could multiply the approved ~64 MiB Argon2 memory cost into an availability event.

**Chosen method.** Preserve approved Argon2id parameters and bound expensive hash/verify/dummy operations behind one shared fail-fast capacity gate; reject stored encodings that exceed approved work before executing them.

**Why selected.** Lowering Argon2 cost would weaken password security. Admission control addresses availability without weakening the cryptographic baseline.

**Detailed dossier.** `../stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md`.

### P2-01 / P2-14 — authentication operational hardening

**Root cause.** Rejected logout persistence happened before bounded admission, and limiter state lacked bounded global lifecycle/cardinality.

**Consequence.** Attackers could amplify requests into persistent writes or memory growth even when individual auth limits existed.

**Chosen method.** Put logout behind admission before persistence and bound both attempt flow and stored limiter buckets with reclamation.

**Why selected.** It closes resource-amplification paths at the existing process-local boundary without pretending to provide distributed rate limiting or introducing Redis scope.

**Detailed dossier.** `../stages/STAGE_03_31_AUTH_OPERATIONAL_HARDENING.md`.

### P2-02 / P2-03 / P2-04 — import review integrity

**Root cause.** Review proof did not bind normalized meaning; row bounds were adapter-level; reconciliation used an arbitrary latest-page shortcut.

**Consequence.** Review-to-append semantic drift, oversized parser work, and missed old duplicates were possible.

**Chosen method.** Bind a versioned review token to normalized parser/review semantics, enforce row limit in the parser, and query the complete relevant history using bounded financial identity/date predicates.

**Why selected.** Each guarantee is moved to the layer that owns it: parser owns parsing limits, signed review proof owns reviewed semantics, PostgreSQL owns existing ledger truth.

**Detailed dossier.** `../stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md`.

### P2-05 / P2-06 / P2-07 / P2-08 / P2-15 — input/contract hardening

**Root cause.** API/application validation and database/storage limits were not fully aligned, and ambiguous JSON/CSV shapes were tolerated.

**Consequence.** User errors could become 500s, Unicode/storage limits could diverge, valid-looking financial values could overflow persistence, and unknown/duplicate fields could change financial intent.

**Chosen method.** Deterministic validation errors, explicit note bounds, `NUMERIC(28,8)`-aligned financial admission including derived snapshot values, strict JSON writes, and duplicate normalized-header rejection.

**Why selected.** Fail-fast contract validation is cheaper and safer than late persistence errors. Snapshot range protection remains inside the same transaction, preserving atomic financial behavior.

**Review iteration.** Initial Stage 3.29 review returned REQUEST CHANGES because ingress bounds alone did not prove aggregate/derived snapshot values fit persistence. The final solution added pre-persistence aggregate/derived admission and PostgreSQL rollback proof.

**Detailed dossier.** `../stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md`.

### P2-09 / P2-13 — exact replay and browser retry continuity

**Root cause.** Server idempotency stored business completion but not the exact original HTTP artifact; browser retry identity was ephemeral and initially not principal-isolated.

**Consequence.** Duplicate retries could produce a response based on later mutable state, and reload/remount could generate a new command identity. The first browser fix also risked cross-principal retry-slot collision.

**Chosen method.** Persist the exact response artifact atomically with the financial mutation; maintain only opaque short-lived retry metadata in sessionStorage using a stable-principal-scoped hashed slot.

**Why selected.** Exact replay must be independent of mutable resource state. Browser continuity needs technical command identity but not persisted financial payloads or tokens. Principal scoping preserves account isolation.

**Review iteration.** First independent review kept P2-13 open until stable-principal scoping and an A→B→A regression were added.

**Detailed dossier.** `../stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md`.

### P2-10 / P2-11 — snapshot rebuild reporting and work planning

**Root cause.** `importflow` guessed rebuilt dates from input, while PostgreSQL performed broader cascading rebuild work once per imported trade date.

**Consequence.** API reporting could omit actually rebuilt later snapshots and one command could rebuild the same later snapshot multiple times.

**Chosen method.** PostgreSQL computes one deterministic affected-date union and returns it as the command outcome; one sorted unique plan is rebuilt exactly once.

**Why selected.** The database is the only layer that knows both imported ledger changes and existing snapshot state under the same lock/transaction, so it must own both plan and reported outcome.

**Detailed dossier.** `../stages/STAGE_03_33_SNAPSHOT_REBUILD_POSTGRES_IMMUTABILITY.md`.

### P2-12 — PostgreSQL append-only runtime boundary

**Root cause.** Runtime immutability was an application/design intention rather than a complete proof of the authenticated PostgreSQL credential graph.

**Consequence.** A credential could potentially mutate protected tables directly or recover mutation authority through masked session identity, SET ROLE reachability, or ADMIN OPTION role administration.

**Chosen method.** Dedicated least-privilege runtime capability role plus fail-closed startup validation on one physical connection. Validation covers `session_user == current_user`, dangerous role attributes, schema CREATE, protected ownership/mutation rights, SET-reachable roles, and role-administration capability.

**Why selected.** Append-only integrity must survive an application bug. Database ACLs are the final persistence boundary; startup validation ensures the actual configured credentials satisfy that boundary rather than trusting deployment documentation.

**Review iterations.** First review exposed `current_user`/SET-role masking; second exposed `ADMIN TRUE / INHERIT FALSE / SET FALSE`. Both attack paths are retained as PostgreSQL regressions and in the stage dossier.

**Detailed dossier.** `../stages/STAGE_03_33_SNAPSHOT_REBUILD_POSTGRES_IMMUTABILITY.md`; canonical closure record `../stages/STAGE_03_33_CLOSURE.md`.

## Remaining P2 remediation

### P2-16 — GitHub governance enforcement

**Confirmed problem.** Canonical `develop` is currently unprotected and repository settings allow merge commit, rebase merge, and squash merge.

**Project effect.** The documented delivery process can be bypassed mechanically, including by direct/administrative paths, so green CI and review are convention rather than an enforced repository invariant.

**Approved planning direction.** Stage 3.34 requires protected `develop`, required PR/check enforcement, conversation resolution, blocked force push/deletion, linear history, squash-only repository methods, and protection applied to the administrator/owner path. If GitHub account/repository configuration cannot mechanically enforce the owner/admin path, P2-16 remains OPEN; disclosure is insufficient.

**Why this direction.** Governance findings cannot be closed by documentation. The repository hosting control must make the forbidden path unavailable.

**Current state.** Planning PR #71 received repeat independent `APPROVED`, but is not canonical until explicit human squash-merge authorization and merge. Implementation/settings enforcement has not started.

### P2-17 — CI/security hardening

**Confirmed problem.** Existing CI has six useful PR jobs and bounded concurrency, but lacks dedicated vet, race, vulnerability, dependency-security, and scheduled security verification.

**Project effect.** Certain race defects, static-analysis failures, reachable known vulnerabilities, dependency risks, or vulnerabilities discovered when no PR is open may not block or surface through the mandatory pipeline.

**Approved planning direction.** Preserve current jobs/concurrency and add `go vet`, PostgreSQL-backed `go test -race`, pinned `govulncheck`, supported Go/pnpm/Python dependency-security checks, and scheduled/nightly security verification with least-privilege workflow permissions.

**Why this direction.** These checks close distinct evidence gaps without changing application architecture or pretending GitHub security products are available when the private-repository entitlement is absent.

**Current state.** Planning PR #71 is independently approved but not merged; Track A implementation has not started.

## Remaining P3 register

The following findings remain OPEN. Their detailed remediation method is deliberately not marked “chosen” until a separately reviewed P3 stage evaluates alternatives and records the 18-part evidence required by `REMEDIATION_DOCUMENTATION_STANDARD.md`.

### P3-01 — password length semantics

- **Cause:** byte-length implementation and user-facing length semantics are not explicitly aligned for Unicode input.
- **Possible effect:** surprising rejection/acceptance for multibyte passwords and contract inconsistency.
- **Closure requirement:** define the canonical unit and align API, Go, UI, tests, and security limits without weakening password policy.

### P3-02 — true IANA timezone validation

- **Cause:** validation checks shape/format rather than authoritative timezone database membership.
- **Possible effect:** invalid zones can persist and fail later at conversion/presentation boundaries.
- **Closure requirement:** one canonical IANA validation path with positive/negative contract tests.

### P3-03 — Decimal grammar parity

- **Cause:** OpenAPI decimal lexical grammar and Go parser acceptance are not mathematically identical.
- **Possible effect:** generated clients and server disagree on valid financial strings.
- **Closure requirement:** one canonical lexical grammar plus mutation/parity tests across OpenAPI and Go.

### P3-04 — Unicode `maxLength` semantics

- **Cause:** some validators count bytes while contracts imply characters.
- **Possible effect:** non-ASCII user input behaves inconsistently between layers.
- **Closure requirement:** per-field Unicode unit definition and end-to-end validator parity.

### P3-05 — idempotency/session cleanup lifecycle

- **Cause:** exact-replay/session persistence gained correctness guarantees before a complete retention/cleanup lifecycle was defined.
- **Possible effect:** indefinite growth and unclear operational retention guarantees.
- **Closure requirement:** safe retention/cleanup design that cannot race active replay/session semantics, plus cleanup regression and operational evidence.

### P3-06 — `httpapi/api.go` god-module

- **Cause:** incremental vertical slices accumulated composition/handler responsibilities in one module.
- **Possible effect:** higher review/change risk, coupling, and maintenance cost.
- **Closure requirement:** behavior-preserving decomposition with unchanged OpenAPI/runtime semantics and full regression coverage.

### P3-07 — transaction form fixture-like defaults

- **Cause:** development/demo convenience values remain in user-facing form state.
- **Possible effect:** misleading UX and accidental submission risk.
- **Closure requirement:** neutral production defaults/placeholders with preserved validation and accessibility.

### P3-08 — migration validator policy gap

- **Cause:** governance describes stronger migration rules than the executable validator proves.
- **Possible effect:** a migration can satisfy tooling while violating the intended policy.
- **Closure requirement:** either strengthen executable checks or narrow the stated policy, with mutation tests proving the final contract.

### P3-09 — Next.js maintenance update

- **Cause:** dependency maintenance lag.
- **Possible effect:** growing upgrade delta and delayed bug/security fixes.
- **Closure requirement:** reviewed version selection, lockfile-only dependency delta where possible, typecheck/tests/build, and no presentation-boundary drift.

### P3-10 — Fiber maintenance update

- **Cause:** dependency maintenance lag; original audit did not establish an active exploitable CVE in the pinned project path.
- **Possible effect:** growing maintenance/security delta even absent a currently proven exploit.
- **Closure requirement:** evidence-driven upgrade with full Go/PostgreSQL/OpenAPI regressions and no unsupported vulnerability claim.

## Canonical closure history

| Stage | Findings | Implementation | Closure | Final state produced |
| --- | --- | --- | --- | --- |
| 3.27 | P1-02, P1-03, P1-04 | PR #55 → `6e8c806de857f844954f1db513487357dfe90187` | PR #58 | 2 P1 + 17 P2 + 10 P3 remained |
| 3.28 | P1-01, P1-05 | PR #59 → `dc83f5f3a11da164e6809593861d96ccf47b29ca` | PR #60 → `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` | P1=0; 17 P2 + 10 P3 remained |
| 3.29 | P2-05/06/07/08/15 | PR #61 → `7331d3f34783baec3997497d1a79b78eaa558bd4` | PR #62 → `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` | 12 P2 + 10 P3 remained |
| 3.30 | P2-02/03/04 | PR #63 → `8f68dd18800918e6a9882e995e13dba2723dc929` | PR #64 → `ae6497050692798795efb85678af64db97cc5f53` | 9 P2 + 10 P3 remained |
| 3.31 | P2-01/14 | PR #65 → `9bf4d1d31597918eacf0c3358bf6caa2aa9db897` | PR #66 → `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` | 7 P2 + 10 P3 remained |
| 3.32 | P2-09/13 | PR #67 → `0623d5ef326cd783b7dc0417dbcb02f18c506171` | PR #68 → `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` | 5 P2 + 10 P3 remained |
| 3.33 | P2-10/11/12 | PR #69 → `87a7c38e16062a5f3fcef3727f60c0c6741eb805` | PR #70 → `71a1faeb97d33d05f2936111b53f1285edddabe9` | 2 P2 + 10 P3 remain |

## Governance rule for future updates

When a finding changes state, this register must be updated in the same closure-governance flow that makes the new state canonical. Historical `REQUEST CHANGES`, rejected alternatives, and prior insufficient remediation attempts must not be deleted merely because a later design succeeds.
