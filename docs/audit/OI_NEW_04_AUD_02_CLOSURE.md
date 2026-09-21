# OI-NEW-04-AUD-02 — Snapshot persistence fan-out closure candidate

## Status

| Field | Value |
| --- | --- |
| Finding | `OI-NEW-04-AUD-02` |
| Original severity | `P2` |
| Original class | ACTIVE-R2 snapshot persistence SQL fan-out / transaction-duration amplification |
| Current status | `CLOSURE CANDIDATE` |
| Machine-readable status | `OI_NEW_04_AUD_02_STATUS=AWAITING_INDEPENDENT_CLOSURE_REVIEW` |
| Closure runtime scope | None — documentation/governance only |
| Repository-wide audit status | Ongoing; this record is finding-specific only |

This document is a finding-specific closure candidate for `OI-NEW-04-AUD-02`.
It does not rewrite the remediation candidate record in
`OI_NEW_04_AUD_02_REMEDIATION.md`, and it does not claim that the broader
OI-NEW-04 adversarial reassessment or repository-wide audit is complete.

Chronology:

historical OI-NEW-04 remediation
→ later independent adversarial audit
→ `OI-NEW-04-AUD-02`
→ bounded/set-based snapshot-persistence remediation
→ independent pre-commit approval
→ PR #198
→ protected PR CI #561
→ independent PR/CI merge approval
→ squash merge
→ independent post-merge verification
→ this closure candidate.

## Concise closure summary

- **Problem:** ACTIVE-R2 persisted each affected snapshot date through a separate PostgreSQL `INSERT` / `ExecContext`, causing O(D) database statement fan-out while the portfolio mutation transaction and serialization lock remained held.
- **Root cause:** replay states were computed correctly in Go, but persistence still looped over affected dates and performed one database statement per date.
- **Failure scenario:** a mutation affecting many historical snapshot dates could keep bounded ledger replay while still amplifying PostgreSQL round trips and transaction duration independently of the raw-ledger work bound.
- **Impact:** database statement count and lock/transaction duration could grow with affected snapshot-date cardinality. No incorrect financial result was independently demonstrated by this finding.
- **Remediation:** ACTIVE-R2 now enforces a defensive affected-date ceiling of 5,000 and persists all admitted snapshot versions with one set-based PostgreSQL `INSERT`, while retaining the existing deterministic replay-state calculation.
- **Why this solution:** the date ceiling reuses the already-established C′ bounded-work limit and the independent SQL-statement bound removes the former one-round-trip-per-date amplification.
- **Financial semantics:** exact Decimal state calculation, WAC behavior, immutable ledger behavior, methodology versioning, input watermarking and append-only snapshot/version semantics are preserved.
- **Security/capability boundary:** no R0/R1/R2 runtime privilege change, migration, auth/session change, provider change, frontend production change or OpenAPI change was introduced.
- **Regression evidence:** canonical ACTIVE-R2 import with 100 distinct trade dates, D=5,000 affected snapshot dates, D=5,001 fail-closed rollback, cancellation rollback, same-portfolio concurrency, independent row/statement instrumentation, normal ACTIVE-R2 execution and race execution all passed.
- **Residual limitation:** `detectStaleRowsAgainstEpoch` / `logicalFamilyFrozenTx` query amplification remains a separate hardening lead and is not closed by this finding.
- **Broader scope:** the OpenInvest Independent Adversarial Repository Audit remains ongoing.

## Remediation source

```text
PRE_REMEDIATION_DEVELOP_SHA=285a4f544aa66778a7ee7b21de1ec026571f3cdf

SOURCE_BRANCH=fix/oi-new-04-snapshot-fanout-bound
SOURCE_COMMIT_SHA=eb0c0eb037a325757326feb60c7938e6ec907dce
SOURCE_PARENT_SHA=285a4f544aa66778a7ee7b21de1ec026571f3cdf
SOURCE_TREE_SHA=e9505742de8cc0759856dbde818611d288ce57ac
SOURCE_COMMIT_MESSAGE=fix(oi-new-04): bound snapshot persistence fan-out

APPROVED_PATCH_SHA256=e99a20ede7b8dc0786b118a60a85fbd74578995c0d8670131f5ea3bbfdcd1a77
```

The source commit changed exactly six approved files:

```text
backend-go/internal/postgres/oi_new_04_active_integration_test.go
backend-go/internal/postgres/oi_new_04_instrumentation.go
backend-go/internal/postgres/oi_new_04_replay_epochs.go
backend-go/internal/postgres/oi_new_04_snapshot_batch.go
backend-go/internal/postgres/oi_new_04_snapshot_fanout_integration_test.go
docs/audit/OI_NEW_04_AUD_02_REMEDIATION.md
```

No unrelated source file entered the reviewed source commit.

## Protected PR CI evidence

```text
PR_NUMBER=198
PR_HEAD_SHA=eb0c0eb037a325757326feb60c7938e6ec907dce

CI_RUN_NUMBER=561
CI_RUN_ID=35665192621
CI_EVENT=pull_request
CI_HEAD_SHA=eb0c0eb037a325757326feb60c7938e6ec907dce
CI_STATUS=completed
CI_CONCLUSION=success

REQUIRED_CHECKS=10/10 SUCCESS
```

Required contexts:

```text
Go tests=SUCCESS
Python tests=SUCCESS
Frontend build and typecheck=SUCCESS
OpenAPI contract=SUCCESS
Docker Compose config=SUCCESS
PostgreSQL migration validation=SUCCESS
Go vet=SUCCESS
Go race tests=SUCCESS
Go vulnerability scan=SUCCESS
Dependency security scan=SUCCESS
```

## Raw ACTIVE-R2 evidence

The protected `Go tests` and `Go race tests` raw job logs were inspected directly.
`TestOINew04ActiveCPrimeIndependentReviewSuite` ran and passed in both jobs and
was not skipped.

New `OI-NEW-04-AUD-02` markers were present in both normal and race logs:

```text
DISTINCT_TRADE_DATES=PASS
HIGH_D_DISTINCT_DATES=PASS
SNAPSHOT_DATE_BOUND=PASS
SNAPSHOT_SQL_STATEMENT_BOUND=PASS
NO_D_TIMES_H_REPLAY=PASS
MAX_PLUS_ONE_FAIL_CLOSED=PASS
HIGH_D_CANCELLATION_ROLLBACK=PASS
HIGH_D_SAME_PORTFOLIO_CONCURRENCY=PASS
```

All nine previously mandatory OI-NEW-04 markers also remained present in both
normal and race logs:

```text
H_GE_10000_ACTIVE_BOUNDED=PASS
BOUND_5001_FAIL_CLOSED=PASS
FULL_COMMAND_ROLLBACK=PASS
FULL_COMMAND_CANCELLATION=PASS
ACTIVE_SAME_PORTFOLIO_CONCURRENCY=PASS
FINALIZATION_FREEZE=PASS
STALE_WRITER_DETECTION=PASS
REPLAY_EPOCH_MISSING_NONEMPTY_ZERO_FULL_HISTORY=PASS
PROVISIONAL_SUPERSEDED_BY_FINAL=PASS
```

Race evidence:

```text
WARNING_DATA_RACE=ABSENT
DATA_RACE=NO
```

## Local race baseline-control evidence

Before the remote batch, local `go test -race -count=1 ./...` on the candidate
failed only in `internal/httpapi` due existing one-second HTTP test timeouts.
A pristine detached worktree at the exact pre-remediation base reproduced the
same timeout-only failure with no data race.

```text
CANDIDATE_FULL_RACE=HTTPAPI_TIMEOUT_ONLY
BASELINE_FULL_RACE=HTTPAPI_TIMEOUT_ONLY
CANDIDATE_DATA_RACE=NO
BASELINE_DATA_RACE=NO
BASELINE_REPRODUCES_LOCAL_HTTPAPI_TIMEOUT=YES
LOCAL_FULL_RACE_GATE=PASS_WITH_BASELINE_CONTROL
```

Protected GitHub race CI for PR #198 subsequently passed.

## Independent review chronology

```text
PRE_COMMIT=APPROVE_FOR_COMMIT
PR_CI=APPROVE_FOR_MERGE
POST_MERGE=POST_MERGE_VERIFIED
```

The independent PR/CI review additionally reported:

```text
PR_HEAD_EXACT=PASS
ONE_COMMIT=PASS
SCOPE=PASS
CI_10_OF_10=PASS
ACTIVE_R2_NORMAL_RAW_LOG=PASS
ACTIVE_R2_RACE_RAW_LOG=PASS
NEW_MARKERS_NORMAL=8/8
NEW_MARKERS_RACE=8/8
EXISTING_MARKERS_NORMAL=9/9
EXISTING_MARKERS_RACE=9/9
DATA_RACE=NO
POST_PUSH_DRIFT=NO
NEW_P2_PLUS_BLOCKER=NO
SAFE_TO_MERGE=YES
```

## Merge and post-merge identity evidence

PR #198 was squash-merged only after the independent `APPROVE_FOR_MERGE`
verdict.

```text
FINAL_DEVELOP_SHA=1c9d33bc353bf1d34b52d6c9e7286e41175647fb
FINAL_PARENT_SHA=285a4f544aa66778a7ee7b21de1ec026571f3cdf
FINAL_TREE_SHA=e9505742de8cc0759856dbde818611d288ce57ac
SOURCE_TREE_SHA=e9505742de8cc0759856dbde818611d288ce57ac

SOURCE_TREE_EQUALS_FINAL_TREE=YES
FINAL_FILE_BLOB_IDENTITY=6/6
POST_MERGE_DRIFT=NO
```

The independent post-merge review returned:

```text
POST_MERGE_VERIFIED

DEVELOP_SHA=PASS
PARENT_SHA=PASS
TREE_IDENTITY=PASS
FILE_BLOB_IDENTITY=6/6
POST_MERGE_DRIFT=NO
NEW_P2_PLUS_BLOCKER=NO
READY_FOR_FINDING_CLOSURE=YES
```

## Post-merge CI statement

No separate workflow run was present for the squash-merge SHA
`1c9d33bc353bf1d34b52d6c9e7286e41175647fb`.

This closure candidate therefore does **not** claim a post-merge CI success.
Content identity is instead supported by exact source-tree/final-tree equality
and 6/6 final file-blob equality, while the identical source tree had already
passed the protected PR CI #561.

## Scope boundary

This closure candidate closes only the finding-specific question of
`OI-NEW-04-AUD-02` after independent review of the remediation, protected PR CI,
raw ACTIVE-R2 evidence and post-merge content identity.

It does not assert that:

- all possible OI-NEW-04 performance/query-amplification leads are closed;
- the broader OI-NEW-04 adversarial reassessment is complete;
- the repository-wide OpenInvest audit is complete.

## Closure activation rule

Until this exact closure candidate is independently reviewed and the approved
documentation is merged through the governed documentation path, the finding
must not be recorded as finally closed.

Current candidate state:

```text
OI_NEW_04_AUD_02_STATUS=AWAITING_INDEPENDENT_CLOSURE_REVIEW
```

This document intentionally does not declare `OI-NEW-04-AUD-02` finally
`CLOSED` at pre-commit stage.
