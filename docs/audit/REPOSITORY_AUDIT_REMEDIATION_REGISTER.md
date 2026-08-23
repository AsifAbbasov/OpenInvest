# Repository Audit Remediation Register

| Field | Value |
| --- | --- |
| Status | Current remediation index |
| Original audit | 32 findings: P0=0, P1=5, P2=17, P3=10 |
| Canonical baseline for this register | `develop` at `71a1faeb97d33d05f2936111b53f1285edddabe9` after Stage 3.33 closure |
| Current remaining backlog | P0=0, P1=0, P2=2, P3=10; total=12 |
| Documentation standard | `docs/audit/AUDIT_FINDING_DOCUMENTATION_STANDARD.md` |

## Purpose

This file is the single cross-finding index for the original repository audit. It is not a substitute for the detailed stage dossiers. Each row points to the remediation stage or current planned stage where the full root cause, impact, design rationale, alternatives, tests, review history, residual risk, and canonical evidence belong.

Statuses in this register are conservative. A finding is CLOSED only after canonical merge and closure governance. Planning or an approved implementation candidate does not by itself close a finding.

## P1 findings

| ID | Finding summary | Root cause / project impact summary | Remediation / rationale summary | Detailed evidence | Canonical status |
| --- | --- | --- | --- | --- | --- |
| P1-01 | Refresh replay did not invalidate the full session family | A replayed refresh credential could leave related sessions usable, weakening token-family compromise containment | Family-wide replay invalidation and adversarial session tests; selected to fail closed on credential replay rather than preserve partially trustworthy family state | `docs/stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md` | CLOSED — Stage 3.28 closure PR #60, `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` |
| P1-02 | `broker_operation_id` was lost before canonical ledger persistence | Imported broker-operation identity was discarded at the append boundary, weakening durable deduplication/reconciliation identity | Versioned persisted import identity carried through append boundary; preserves immutable provenance rather than relying on transient parser state | `docs/stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md` | CLOSED — Stage 3.27 closure PR #58, `6f50c39...` closure line with implementation `6e8c806de857f844954f1db513487357dfe90187` |
| P1-03 | Cash-flow near-match could conflict different amounts on same day | Reconciliation key omitted amount for cash operations, allowing economically distinct events to collide | Amount-aware cash reconciliation; selected because economic identity must include the cash amount | `docs/stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md` | CLOSED — Stage 3.27 |
| P1-04 | Deposit/withdrawal commission/tax accepted but snapshot semantics ignored them | API accepted financial fields whose economic effect was undefined, creating silent ledger/projection divergence | Fail-closed zero-fee cash-flow semantics until an explicit economic model exists; safer than accepting unsupported money fields | `docs/stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md` | CLOSED — Stage 3.27 |
| P1-05 | Public Argon2 admission allowed dangerous memory bursts | Expensive password hashing could be admitted at a rate that multiplied memory usage under burst load | Bounded authentication admission around Argon2 work; chosen to control resource amplification before expensive work starts | `docs/stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md` | CLOSED — Stage 3.28 |

## P2 findings

| ID | Finding summary | Root cause / project impact summary | Remediation / rationale summary | Detailed evidence | Canonical status |
| --- | --- | --- | --- | --- | --- |
| P2-01 | Logout could generate unlimited database audit writes | Logout path admitted rejected/auth-related work without sufficient bounded admission | Logout placed behind auth admission and persistence bounded by limiter policy | `docs/stages/STAGE_03_31_AUTH_OPERATIONAL_HARDENING.md` | CLOSED — PR #66, `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` |
| P2-02 | Import review token did not freeze normalized financial meaning | Approval token was not bound strongly enough to the exact reviewed parser semantics | Versioned short-lived review token bound to normalized semantics and APPENDABLE rows | `docs/stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md` | CLOSED — PR #64, `ae6497050692798795efb85678af64db97cc5f53` |
| P2-03 | 100-row limit was applied after full parse | Admission bound occurred too late, allowing unnecessary parsing/resource work | Parser-owned fail on the 101st CSV data row | `docs/stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md` | CLOSED — Stage 3.30 |
| P2-04 | Import review considered only the latest 100 existing transactions | Truncated history could miss relevant prior transactions and produce incorrect reconciliation | PostgreSQL targeted full-history query constrained to reconciliation-relevant dates/privacy-minimized identity | `docs/stages/STAGE_03_30_IMPORT_REVIEW_INTEGRITY.md` | CLOSED — Stage 3.30 |
| P2-05 | Invalid decimal input could become HTTP 500 | Parser/validation errors were not uniformly mapped to client-domain errors | Strict decimal ingress validation and explicit error semantics | `docs/stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` | CLOSED — PR #62, `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc` |
| P2-06 | Note length >500 could become HTTP 500 | Length validation and storage semantics were inconsistent | Explicit Unicode-aware/contract-aligned input validation for the audited boundary | `docs/stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` | CLOSED — Stage 3.29 |
| P2-07 | Decimal magnitude could exceed PostgreSQL `NUMERIC` bounds | Application ingress did not enforce database representable bounds before persistence/aggregation | `NUMERIC(28,8)`-aligned ingress and snapshot aggregate bounds | `docs/stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` | CLOSED — Stage 3.29 |
| P2-08 | `additionalProperties:false` was not uniform on financial commands | OpenAPI strictness was inconsistent, allowing unreviewed fields on sensitive commands | Strict JSON command schemas for financial mutations | `docs/stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` | CLOSED — Stage 3.29 |
| P2-09 | Idempotent duplicate did not replay exact original HTTP result | Idempotency stored business completion without preserving the exact response artifact | Persist opaque exact HTTP artifact atomically with mutation and replay it verbatim | `docs/stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md` | CLOSED — closure PR #68, `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be` |
| P2-10 | `snapshotDatesRebuilt` omitted dates actually rebuilt | Response derived dates from input instead of the database-owned rebuild plan | PostgreSQL owns deterministic affected-date union and returns exact outcome | `docs/stages/STAGE_03_33_SNAPSHOT_REBUILD_POSTGRES_IMMUTABILITY.md`; `docs/stages/STAGE_03_33_CLOSURE.md` | CLOSED — PR #70, `71a1faeb97d33d05f2936111b53f1285edddabe9` |
| P2-11 | Backdated batch repeatedly rebuilt later snapshots | Rebuild cascade was invoked once per imported trade date, repeating work on overlapping suffixes | Compute one sorted unique rebuild plan and rebuild each affected date exactly once | Stage 3.33 dossier and closure | CLOSED — Stage 3.33 |
| P2-12 | PostgreSQL runtime lacked declared append-only privilege boundary | Application runtime credentials could retain mutation/escalation authority inconsistent with immutable ledger architecture | Least-privilege runtime role plus same-connection credential-graph validation of session/current user, SET-reachable roles, ownership/mutation capabilities, and ADMIN OPTION; multiple adversarial DB tests | Stage 3.33 dossier and closure | CLOSED — Stage 3.33 |
| P2-13 | Browser retry/idempotency identity was lost on reload/remount | Retry key lived only in transient component state and later needed principal isolation | Principal-scoped sessionStorage retry journal without raw financial payload/token persistence | `docs/stages/STAGE_03_32_IDEMPOTENCY_REPLAY_BROWSER_RECOVERY.md` | CLOSED — Stage 3.32 |
| P2-14 | Authentication limiter key map was unbounded | Per-key limiter state had no global cardinality/reclamation bound | Bound attempts, total downstream auth work, active bucket cardinality, and expired bucket reclamation | `docs/stages/STAGE_03_31_AUTH_OPERATIONAL_HARDENING.md` | CLOSED — Stage 3.31 |
| P2-15 | Duplicate CSV headers were not rejected | Parser accepted ambiguous column identity | Fail-closed duplicate-header rejection | `docs/stages/STAGE_03_29_INPUT_CONTRACT_HARDENING.md` | CLOSED — Stage 3.29 |
| P2-16 | GitHub governance is not mechanically enforced | `develop` is unprotected and repository merge settings do not enforce the frozen PR/check/squash-only policy | Stage 3.34 Track B: protected `develop`, required PR/checks/conversation resolution, no force-push/deletion, linear history, squash-only, and no owner/admin bypass; if unavailable, finding remains open | Stage 3.34 plan in PR #71 | OPEN |
| P2-17 | CI lacks complete security/concurrency verification class | Existing functional CI does not include vet/race/vulnerability/dependency/nightly security gates; concurrency itself is already present | Stage 3.34 Track A: `go vet`, PostgreSQL-backed `go test -race`, pinned `govulncheck`, dependency-security coverage for Go/pnpm/Python, scheduled security verification, least privilege | Stage 3.34 plan in PR #71 | OPEN |

## P3 findings

The P3 findings remain OPEN and must each receive a detailed dossier satisfying `AUDIT_FINDING_DOCUMENTATION_STANDARD.md` before canonical closure.

| ID | Finding summary | Risk / intended remediation direction | Status |
| --- | --- | --- | --- |
| P3-01 | Password length semantics use bytes rather than the intended user-facing character semantics | Align password validation contract and implementation semantics without weakening minimum security requirements | OPEN |
| P3-02 | Timezone validation is not true IANA timezone validation | Replace superficial validation with canonical IANA location validation and contract tests | OPEN |
| P3-03 | OpenAPI Decimal grammar differs from the Go decimal parser | Establish one canonical decimal grammar and prove parser/contract parity | OPEN |
| P3-04 | Some `maxLength` semantics are byte-based rather than Unicode/code-point aligned | Make API contract, validation, and persistence semantics explicit and consistent | OPEN |
| P3-05 | Idempotency/session cleanup lifecycle is absent | Define bounded retention/cleanup semantics that preserve replay/security guarantees | OPEN |
| P3-06 | `httpapi/api.go` is a god-module | Decompose HTTP composition without changing contracts or auth/financial boundaries | OPEN |
| P3-07 | Transaction form contains fixture-like defaults | Remove misleading demo defaults from production-facing UI state | OPEN |
| P3-08 | Migration validator is weaker than the documented migration policy | Make automated validation enforce the actual declared migration invariants | OPEN |
| P3-09 | Next.js dependency requires planned update | Upgrade within reviewed compatibility/security scope with lockfile/build/regression evidence | OPEN |
| P3-10 | Fiber dependency requires planned update | Upgrade within reviewed compatibility/security scope; original audit noted no active CVE on the then-current version | OPEN |

## Canonical progress

After Stage 3.33 canonical closure:

- original findings: 32;
- closed: 20;
- remaining: 12;
- completion by finding count: 62.5%;
- P1+P2 closed: 20/22 = 90.9%;
- remaining P2: P2-16 and P2-17;
- remaining P3: P3-01 through P3-10.

If Stage 3.34 later closes both P2 findings canonically, the remaining audit backlog becomes P0=0, P1=0, P2=0, P3=10, total=10. This does not imply production readiness and does not close separate Stage 3.25 privacy Security Review evidence work.

## Maintenance rule

Every future remediation PR that changes the canonical status of an audit finding must update this register in its closure-governance step. A finding must never be marked CLOSED here before the corresponding canonical merge and required closure evidence exist.