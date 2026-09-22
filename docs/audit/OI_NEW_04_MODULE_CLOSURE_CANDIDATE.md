# OI-NEW-04 - Deep adversarial review closure record

## Status

```text
MODULE=OI-NEW-04
STATUS=CLOSED
TECHNICAL_DISPOSITION=COMPLETE
REPOSITORY_WIDE_AUDIT=ONGOING
```

This is the verified module-level closure record for OI-NEW-04. It does not
declare the OpenInvest repository-wide audit complete. Independent final
consolidation and post-merge identity verification have completed.

## Problem / audit scope

OI-NEW-04 protects the financial write path when a portfolio requires
retroactive replay. Its assurance boundary covers bounded replay work,
generation-scoped replay epochs and watermarks, append-only snapshots,
same-portfolio concurrency, cancellation and rollback, finalization freezes,
idempotency, and deterministic Decimal/WAC financial results.

The module must fail closed rather than reconstruct unbounded ledger history
when an operation exceeds its admitted replay or snapshot-date limits.

## Findings lifecycle

### OI-NEW-04-AUD-01

```text
STATUS=CLOSED
CLASS=ACTIVE-R2 protected CI evidence gap
```

The issue was that protected normal and race Go jobs could complete without
running the strongest opt-in ACTIVE-R2 suite. The remediation explicitly wires
the suite into both protected Go paths, with canonical R2 runtime setup and
separate owner/runtime database identities. The detailed closure history is in
[OI_NEW_04_AUD_01_CLOSURE.md](OI_NEW_04_AUD_01_CLOSURE.md).

### OI-NEW-04-AUD-02

```text
STATUS=CLOSED
CLASS=snapshot persistence SQL fan-out and transaction-duration amplification
```

The remediation bounds affected snapshot dates at `D <= 5000`, uses one
set-based snapshot persistence statement, retains append-only snapshots, and
preserves replay, Decimal, and WAC semantics. The detailed closure record is in
[OI_NEW_04_AUD_02_CLOSURE.md](OI_NEW_04_AUD_02_CLOSURE.md).

## Final adversarial verification

```text
BASE_SHA=a48ffc140b7b43a2cbfcedcbfe6767396a58fa91
BASELINE_DRIFT=NO
AUD_01_STATUS=CLOSED
AUD_02_STATUS=CLOSED
STALE_FAMILY_QUERY_LEAD=P3_RESIDUAL_ACCEPTED
NEW_P2_PLUS_FINDING=NONE
```

| Check | Result |
| --- | --- |
| ACTIVE-R2 normal suite | PASS |
| ACTIVE-R2 race suite | PASS |
| Existing normal markers | 17/17 |
| Existing race markers | 17/17 |
| Data-race warning | NO |
| `go test ./...` | PASS |
| `go test -race ./...` | BASELINE_CONTROLLED_FAILURE |
| `go vet ./...` | PASS |
| Full-history fallback | ABSENT |
| B=5000 bounded admission | PASS |
| B=5001 fail closed | PASS |
| D=5000 snapshot bound | PASS |
| D=5001 fail closed | PASS |
| Cancellation atomicity | PASS |
| Same-portfolio concurrency | PASS |
| Epoch/watermark invariants | PASS |
| Finalization freeze | PASS |
| Provisional superseded by final | PASS |
| Financial semantics | PASS |
| Decimal/WAC semantics | PASS |
| R0/R1/R2 capability boundary | PASS |
| Supported production bypass write path | NO |

The production composition root creates `NewReplayRuntime`, whose implemented
financial mutation routes use the replay-safe handlers. Legacy direct-store
paths remain in non-canonical code but were not reachable through the supported
current ACTIVE-R2 runtime; no production bypass write path was demonstrated.

The source tree identical to the reviewed baseline had already passed protected
GitHub PR CI #563, including successful `Go tests`, `Go race tests`, and the
explicit ACTIVE-R2 normal and race suite steps. No separate workflow run exists
for the final baseline squash SHA, and this document does not claim one.

## Final closure evidence

```text
CLOSURE_PR=201
CLOSURE_FINAL_DEVELOP_SHA=7eede7a36c9df05bd7486a3cc738462544bc60f7
CLOSURE_PARENT_SHA=a48ffc140b7b43a2cbfcedcbfe6767396a58fa91
SOURCE_TREE_EQUALS_FINAL_TREE=YES
FILE_BLOB_IDENTITY=1/1
POST_MERGE_DRIFT=NO
FINAL_POST_MERGE_WORKFLOW_RUNS=0
POST_MERGE_CI_CLAIM=NONE
```

## Residual limitations

```text
detectStaleRowsAgainstEpoch / logicalFamilyFrozenTx=P3_RESIDUAL_ACCEPTED
```

The stale/mixed legacy-writer condition can cause N+1 logical-family lookup
amplification. A duplicate-family case can approach roughly 5,000 indexed SQL
lookups. Each query is inexpensive with the current index, but the total work
and portfolio-lock duration remain O(N).

This is non-blocking because clean, supported ACTIVE-R2 could not construct the
high-N stale state; no financial corruption was demonstrated; and cancellation
and rollback remained atomic. It remains a future hardening opportunity and is
not marked fixed by this closure record.

## Race evidence qualification

Local full `go test -race ./...` reproduced the previously known one-second
HTTP auth timeout behavior under race and Argon2 instrumentation. The same
failure is baseline-controlled. No data-race warning was observed, and the
dedicated ACTIVE-R2 race suite passed. Protected GitHub Go race evidence for
the current reviewed baseline had already passed where applicable.

```text
CANDIDATE_FULL_RACE=HTTPAPI_TIMEOUT_ONLY
BASELINE_FULL_RACE=HTTPAPI_TIMEOUT_ONLY
CANDIDATE_DATA_RACE=NO
BASELINE_DATA_RACE=NO
```

## Module disposition

```text
OI_NEW_04_MODULE_DISPOSITION=COMPLETE
TECHNICAL_DISPOSITION=COMPLETE
NO_UNRESOLVED_P2_PLUS=YES
POST_MERGE_CLOSURE_VERIFIED=YES
OI_NEW_04_STATUS=CLOSED
```

No new P2+ finding was demonstrated by the final consolidation. Independent
post-merge verification completed with exact tree, blob, and SHA256 identity;
OI-NEW-04 is `CLOSED`.

## Broader audit boundary

OpenInvest repository-wide audit remains `ONGOING`. This closure record covers only
OI-NEW-04 and neither closes nor weakens any unrelated audit area.
