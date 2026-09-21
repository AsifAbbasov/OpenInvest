# OI-NEW-04-AUD-01 — Independent audit CI-gate closure record

## Status

| Field | Value |
| --- | --- |
| Finding | `OI-NEW-04-AUD-01` |
| Original severity | `P2` |
| Original class | protected financial/concurrency CI release-gate gap |
| Current status | `CLOSED` |
| Machine-readable status | `OI_NEW_04_AUD_01_STATUS=CLOSED` |
| Closure runtime scope | None — documentation/governance only |
| Repository-wide audit status | Ongoing; this record is finding-specific only |

This document records the final finding-specific closure of the later independent-audit
finding `OI-NEW-04-AUD-01`. It does not rewrite or supersede the historical
OI-NEW-04 technical-remediation record.

Chronology is preserved as:

historical OI-NEW-04 remediation
→ later independent adversarial audit
→ `OI-NEW-04-AUD-01`
→ protected-CI gate remediation
→ PR #195
→ runtime post-merge verification
→ independent closure review
→ closure PR #196
→ closure post-merge verification
→ final `CLOSED` disposition.

The historical OI-NEW-04 status `CLOSED` in
`OI_NEW_P2_REMEDIATION_CLOSURE.md` remains historical truth about the original
OI-NEW-04 remediation. This record addresses only the later release-gate
evidence defect.

## Concise closure summary

- **Problem:** protected CI could pass while the strongest ACTIVE-R2 replay/concurrency suite was skipped.
- **Root cause:** the suite was opt-in, but required Go CI jobs did not wire its opt-in and owner/runtime database URLs into the protected path.
- **Failure scenario:** a PR could receive green required CI without executing the ACTIVE-R2 evidence that covers bounded replay, rollback/cancellation, concurrency, finalization freeze and stale-writer detection.
- **Impact:** release-gate evidence for OI-NEW-04 was incomplete; this was a CI assurance defect, not a confirmed financial-calculation defect.
- **Initial remediation:** PR #195 wired canonical R2 setup plus explicit normal and `-race` execution of `TestOINew04ActiveCPrimeIndependentReviewSuite`.
- **Review rejection:** N/A — the first remediation passed independent review.
- **Second attack scenario:** N/A — no second remediation design was required.
- **Why this solution:** it preserves generic R0 coverage and production behavior while making ACTIVE-R2 evidence mechanically mandatory in both protected Go gates.
- **Regression evidence:** both targeted suites run and pass, all nine mandatory markers are required, and the race gate reports no data race.
- **CI/review evidence:** PR #195 / CI #558 and closure PR #196 / CI #559 both passed all 10 protected checks; both merges were independently post-merge verified.
- **Residual limitations:** this closes only the CI-gate evidence defect. The broader OI-NEW-04 deep adversarial reassessment and repository-wide audit continue.
- **Disposition:** `OI_NEW_04_AUD_01_STATUS=CLOSED`.

## Finding

**Original defect.** The comprehensive ACTIVE-R2 replay/concurrency integration
suite already existed, but the protected `Go tests` and `Go race tests` jobs did
not configure its opt-in and owner/runtime database URLs. Required CI could
therefore complete successfully while
`TestOINew04ActiveCPrimeIndependentReviewSuite` was skipped.

**Why it mattered.** The strongest OI-NEW-04 evidence for bounded replay,
rollback/cancellation safety, same-portfolio concurrency, finalization freeze,
stale-writer detection and replay-generation behavior was not mechanically
required by the protected release gate.

**Root cause.** The ACTIVE-R2 suite was opt-in, but the protected CI jobs did not
wire that opt-in into their required execution path.

## Remediation

PR #195 changed only:

`.github/workflows/ci.yml`

The protected job names remain:

- `Go tests`
- `Go race tests`

The existing generic R0 commands remain:

- `go test ./...`
- `go test -race ./...`

After the generic R0 phase, both protected jobs now configure canonical R2
through:

`infrastructure/postgres/runtime/openinvest_runtime_role.sql`

with:

`runtime_capability_profile=R2`

Both jobs explicitly execute:

`TestOINew04ActiveCPrimeIndependentReviewSuite`

in normal and `-race` modes.

The ACTIVE-R2 opt-in is explicit:

`OPENINVEST_OI_NEW_04_ACTIVE_R2_TESTS=1`

Owner/runtime database identities remain separate:

- owner: `openinvest`
- runtime: `openinvest_runtime_ci`

Both targeted gates mechanically require the suite PASS line:

`--- PASS: TestOINew04ActiveCPrimeIndependentReviewSuite`

and all nine mandatory evidence markers.

## Immutable remediation evidence

```text
SOURCE_REVIEWED_SHA=8c83b0696e805abec81f6716464b9efdf974a130
SOURCE_REVIEWED_TREE=c978b58519dc143b51b5403cc6c4ae0e788c571f
SOURCE_DIFF_SHA256=fc477b9e5f9fbdf2c3a5b80fa1cd639b43d0560255aeb20583a121e0b5653619

PR=#195
PR_MERGED=YES

FINAL_DEVELOP_SHA=0e867207effc7bade5d36fb395bceebe9884a038
FINAL_PARENT_SHA=de1514adba086dc9ca2224986f6b06238715453d
FINAL_TREE=c978b58519dc143b51b5403cc6c4ae0e788c571f

SOURCE_TREE_EQUALS_FINAL_TREE=YES

FINAL_CI_YML_BLOB=3d6ca701c899d55d97fac4e706389da0eecd8611

CHANGED_FILES=1
EXACT_FILE=.github/workflows/ci.yml
ADDITIONS=68
DELETIONS=0
```

The reviewed source tree and merged `develop` tree are identical. The squash
merge therefore preserved the reviewed repository tree while producing a new
commit identity.

## Protected CI evidence

```text
WORKFLOW=CI
PR_CI_RUN_NUMBER=558
PR_CI_RUN_ID=35589000574
PR_CI_EVENT=pull_request
PR_CI_HEAD_SHA=8c83b0696e805abec81f6716464b9efdf974a130
PR_CI_CONCLUSION=SUCCESS
REQUIRED_CHECKS=10/10 SUCCESS
```

Required jobs:

1. `Go tests` — SUCCESS
2. `Python tests` — SUCCESS
3. `Frontend build and typecheck` — SUCCESS
4. `OpenAPI contract` — SUCCESS
5. `Docker Compose config` — SUCCESS
6. `PostgreSQL migration validation` — SUCCESS
7. `Go vet` — SUCCESS
8. `Go race tests` — SUCCESS
9. `Go vulnerability scan` — SUCCESS
10. `Dependency security scan` — SUCCESS

Independent review additionally inspected raw GitHub Actions evidence and
confirmed that `TestOINew04ActiveCPrimeIndependentReviewSuite` actually ran and
passed in both protected `Go tests` and protected `Go race tests`; it was not
silently skipped.

The following mandatory markers were observed in both gates:

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

## Independent review chronology

```text
RUNTIME_PRE_COMMIT=APPROVE_FOR_COMMIT
RUNTIME_EXACT_COMMIT=APPROVE_FOR_PUSH
RUNTIME_POST_MERGE=POST_MERGE_VERIFIED

CLOSURE_PRE_COMMIT=APPROVE_FOR_CLOSURE_COMMIT
CLOSURE_PR_CI=APPROVE_FOR_CLOSURE_MERGE
CLOSURE_POST_MERGE=POST_MERGE_CLOSURE_VERIFIED
```

The runtime-remediation history does not invent an `APPROVE_FOR_MERGE` verdict
for PR #195. The later documentation closure PR #196 separately received
`APPROVE_FOR_CLOSURE_MERGE`.

## Scope statement

This remediation changed CI enforcement only.

It did not change:

- financial methodology;
- canonical transaction ledger;
- weighted-average-cost algorithm;
- snapshots or replay implementation;
- Go production code;
- database migrations;
- database schema;
- OpenAPI;
- frontend;
- Python;
- provider integrations;
- authentication or session behavior.

The remediation addresses a protected release-gate evidence defect. It is not a
closure of a discovered financial-calculation defect.

## Repository-wide boundary

This closure record is specific to `OI-NEW-04-AUD-01`.

It does **not** assert that:

- the OpenInvest independent adversarial repository audit is complete;
- the repository is generally approved;
- all findings are closed;
- the application is production-ready.

The broader OpenInvest Independent Adversarial Repository Audit remains ongoing.

## Final closure evidence

```text
CLOSURE_PR=#196
CLOSURE_SOURCE_SHA=45530ecf5aa5608f89da074c0a3f3a5ea0a8fd0b
CLOSURE_SOURCE_TREE=bbc251c3e661fbdf7f52a07224b7ad1dd41b4ad1

CLOSURE_CI_RUN_NUMBER=559
CLOSURE_CI_RUN_ID=35593962558
CLOSURE_CI_RESULT=10/10 SUCCESS

FINAL_CLOSURE_DEVELOP_SHA=917828a9cf6b20c7c0e4a956de3ebdd60e8b9850
FINAL_CLOSURE_PARENT_SHA=0e867207effc7bade5d36fb395bceebe9884a038
FINAL_CLOSURE_TREE=bbc251c3e661fbdf7f52a07224b7ad1dd41b4ad1
FINAL_CLOSURE_FILE_BLOB=eb0e0ceb64e1e5afb73a8fa2dd3a27c0a9392583
FINAL_CLOSURE_FILE_SHA256=37d2e751ee7fac5780df41a2d7556c0aabd84c1ff9420aa8b985e0b89bd6c061

SOURCE_TREE_EQUALS_FINAL_CLOSURE_TREE=YES
ACTIVE_R2_NORMAL_RAW_LOG=PASS
ACTIVE_R2_RACE_RAW_LOG=PASS
MANDATORY_MARKERS_NORMAL=9/9
MANDATORY_MARKERS_RACE=9/9
DATA_RACE=NO
POST_MERGE_DRIFT=NO
NEW_P2_PLUS_BLOCKER=NO
POST_MERGE_CLOSURE_VERIFIED=YES
```

## Final disposition

```text
FINDING=OI-NEW-04-AUD-01
DISPOSITION=CLOSED
CLOSURE_BASIS=POST_MERGE_CLOSURE_VERIFIED
OI_NEW_04_AUD_01_STATUS=CLOSED
```

This is a finding-specific closure only. The historical OI-NEW-04 remediation
remains historical truth, the OI-NEW-04 deep adversarial reassessment continues,
and the broader OpenInvest Independent Adversarial Repository Audit remains
ongoing.
