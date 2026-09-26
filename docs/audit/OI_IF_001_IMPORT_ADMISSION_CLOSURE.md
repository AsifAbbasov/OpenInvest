# OI-IF-001 — Import Admission / Resource-Exhaustion Closure

Status: CLOSURE_CANDIDATE
Finding: OI_IF_001
Severity: P2
Original evidence state: CONFIRMED

## 1. Finding identity

Before remediation, authenticated import review and append processing had no
dedicated import-specific server-side active-concurrency bound plus finite
execution and fresh-command admission. Repeated import workloads could
therefore consume parser, application, and database work without the dedicated
resource boundary introduced later by PR #210. No production outage was
demonstrated.

## 2. Original audit scope

The Import Flow deep audit reviewed 65 / 65 files and 17,410 / 17,410
reviewable lines (100.0000%). It produced two P2 findings: `OI_IF_001` and
`OI_IF_002`. `OI_IF_002` is independently CLOSED; this record neither rewrites
its history nor reopens it.

## 3. Root cause

The original import boundary lacked the combined protection of:

1. active processing capacity;
2. an execution-attempt budget; and
3. a fresh-command budget.

Replay semantics make ordinary rate limiting insufficient. An exact completed
replay must remain retrievable, conflict and in-flight state remain
authoritative, and historic V1/V2 completed replays remain recoverable. A
denied fresh command must not reserve transactional replay state, while invalid
work still requires bounded execution admission before expensive processing.

## 4. Review history

An earlier design/remediation was rejected because invalid pre-identity work
could bypass fresh-rate protection while occupying expensive processing
capacity, and because source-file-hash ordering could allow a completed replay
artifact to mask a tampered `sourceFileHash`.

The accepted R3 design separates active capacity, execution admission, and
fresh-command admission. It restores exact source-file-hash fail-closed
ordering before replay recovery. The result is a bounded admission policy that
preserves authoritative replay behavior rather than treating every request as a
new command.

## 5. Final remediation

The merged policy is:

```text
ACTIVE_IMPORT_CAPACITY=2

EXECUTION_PER_SUBJECT=12
EXECUTION_GLOBAL=120
EXECUTION_WINDOW=60s
EXECUTION_MAX_TRACKED_SUBJECTS=2048

FRESH_PER_SUBJECT=6
FRESH_GLOBAL=60
FRESH_WINDOW=60s
FRESH_MAX_TRACKED_SUBJECTS=2048
```

HTTP behavior is:

```text
capacity exhaustion:
503 IMPORT_CAPACITY_EXHAUSTED
Retry-After: 1

execution exhaustion:
503 IMPORT_ADMISSION_EXHAUSTED
Retry-After: 60

fresh-command exhaustion:
429 RATE_LIMITED
Retry-After: 60
```

## 6. Ordering and replay invariants

The verified append path is:

```text
authentication
cheap idempotency-key validation
active capacity
execution admission
JSON / DTO processing
sourceFileHash exact fail-closed validation
current parser preparation
signed review proof / decisions
read-only replay lookup
completed/conflict/in-flight return if present
fresh admission only when there is no replay state
one read-only replay race recheck if fresh admission is denied
existing transactional append/replay boundary
```

The final implementation and regression coverage verify that:

- capacity denial consumes neither execution nor fresh budget;
- execution denial consumes no fresh budget;
- current completed replay consumes execution admission but not fresh admission;
- historical V1/V2 completed replay consumes execution admission but not fresh admission;
- conflict and in-flight outcomes consume no fresh admission;
- historical parser semantics can recover an existing completed artifact but cannot authorize a fresh write;
- `sourceFileHash` mismatch fails closed before replay recovery;
- fresh-rate denial performs exactly one read-only replay recheck; and
- the import append route is bound to `appendImportReplaySafe`.

## 7. CSV and OI-IF-002 interaction

`OI_IF_002` remains independently CLOSED. The final `OI_IF_001` remediation
preserves parser-v3 structural-width policy, the exact 12-field active data-row
bound, the 128-byte header-field bound, the 100-data-row parser cutoff, and
historical V1/V2 completed-replay compatibility. It does not reopen OI_IF_002.

## 8. Test and CI evidence

Technical PR: #210

- Reviewed runtime/security commit: `d9d95b818efef909dbb121ba62c785ac297dd049`
- Approved test-only follow-up: `0c0859ab11890d52bd6ee65d6245ff65b43e5784`
- Technical PR final head: `0c0859ab11890d52bd6ee65d6245ff65b43e5784`
- Protected PR workflow: `36231304293`, run 580, attempt 1
- Required jobs: 10
- Result: 10 / 10 SUCCESS

The successful jobs were Go tests, Python tests, Frontend build and typecheck,
OpenAPI contract, Docker Compose config, PostgreSQL migration validation, Go
vet, Go race tests, Go vulnerability scan, and Dependency security scan.

An earlier CI run failed because a newly added test diagnostic formatted an
`importRate` containing `sync.Mutex` with `%+v`. The independently reviewed
one-line test-only correction replaced that diagnostic with scalar fields, and
fresh final CI passed. This was not a runtime vulnerability.

## 9. Technical merge

Technical PR #210 was squash-merged at `2026-09-26T12:00:00Z`.

```text
FINAL_SQUASH=4c7068c795d97186e5706179c04722e8485dc8c8
PARENT=0d0a1fa100e82d745fc34818c5e2a49ab960ee4e
FINAL_TREE=01ed03718f763112c0ce72ac8eb2cd6bf82405c5
POST_MERGE_DIFF=18 files, +1560, -74
```

Independent post-merge verification confirmed that protected `develop` equals
the final squash, its tree equals the reviewed PR tree, the approved file set
is exact, no unexpected scope was introduced, and all reviewed security
invariants remain present.

## 10. Post-merge CI statement

```text
POST_MERGE_WORKFLOW=NONE
```

The current workflow triggers on `pull_request`, schedule, and manual dispatch,
not push-to-`develop`. The authoritative protected technical CI was workflow
#580 on the exact reviewed PR head, followed by independent post-merge
tree/diff verification. This record does not claim a separate CI workflow on
the final squash SHA.

## 11. Residual limitations

Admission is process-local, not a distributed or edge DoS shield. If the API
scales horizontally, effective limits multiply per replica and deployment must
revisit shared/gateway admission. This is a future scale-out residual, not an
open OI_IF_001 P2 under the current single-process assumption.

The capacity-2, execution-12/120, and fresh-6/60 values are governance
defaults, not production-load-derived optimal thresholds. Future telemetry and
load testing must tune them.

A coordinated set of authenticated subjects can consume the finite global
execution budget and cause bounded fail-closed 503 responses for other import
users. This is a bounded P3 availability/fairness residual, not the original
unbounded P2 resource-amplification path. Transport-layer request buffering and
edge admission are outside the narrow OI_IF_001 finding; PR #210 does not claim
to solve edge or network DoS.

## 12. Scope exclusions

This closure does not close the Import Flow module, the repository-wide audit,
OI-NEW-11 security-header work, OI-NEW-12 future distributed-limiter work,
AUTH-ADV-02, provider/security-header/JWT/product-security residuals, or the
frontend/production final security audit. It starts none of those workstreams.

## 13. Closure activation

```text
PRE_MERGE_STATUS=CLOSURE_CANDIDATE
POST_MERGE_STATUS=CLOSED
CANONICAL_ACTIVATION=THIS_EXACT_CLOSURE_RECORD_ON_PROTECTED_DEVELOP_AFTER_REQUIRED_GATES
```

Before this closure-record PR is merged and independently post-merge verified,
OI_IF_001 remains `REMEDIATED_PENDING_CLOSURE_RECORD`. After that process
succeeds, `OI_IF_001=CLOSED`.
