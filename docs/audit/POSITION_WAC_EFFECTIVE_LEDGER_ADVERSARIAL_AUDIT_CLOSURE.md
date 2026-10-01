# Position / WAC / Effective Ledger Adversarial Audit Closure

Status: CLOSURE_CANDIDATE_PENDING_INDEPENDENT_REVIEW

## Scope and audit baseline

This record is the governance closure candidate for the Position / WAC /
Effective Ledger adversarial-audit module. It covers deterministic portfolio
position reconstruction, Weighted Average Cost (WAC), remaining acquisition
basis, effective-ledger selection, correction and reversal interaction,
historical/as-of truth, portfolio-local ordering, position close/reopen
semantics, position generation identity, position authorization/isolation, and
position-dependent replay correctness.

The completed audit baseline was
`f82d0e8491e8e0bbec44e23bea9bd35d82ea7709`, tree
`6a613b7dd39c341e4c32a4dc64e8772a70b5c20c`.

## Coverage

```text
HUMAN_MAINTAINED_FILES_REVIEWED=105/105
REVIEWABLE_LINES_REVIEWED=34494/34494
REVIEW_COVERAGE_PERCENT=100
```

This is completed audit-scope evidence, not a claim that future code changes
or unreviewed attack variations are impossible to introduce.

## Final findings

### OI_PWEL_001 - P3 residual

```text
SEVERITY=P3
DISPOSITION=P3_CONFIRMED_DEFERRED_RESOURCE_HARDENING
OWNER_MODULE=REPLAY_SNAPSHOT_RESOURCE_AUDIT
```

Current position projection can perform unbounded effective-ledger work for
large historical or legacy ledgers and invokes the effective-ledger selector
more than once during projection. This audit did not demonstrate financial
corruption, data-isolation failure, IDOR, a mutation path, or an
unauthenticated exploit. OI_PWEL_001 is not closed; it is the sole retained
module residual and is delegated to the future Replay / Snapshot / Resource
audit.

### OI_PWEL_002 - closed

```text
SEVERITY=P3
STATUS=CLOSED
```

Malformed portfolio UUID route parameters previously reached PostgreSQL and
mapped to HTTP 500. The remediation validates `portfolioId` with the existing
`transactionRouteUUID` helper before downstream service, store, or database
work in all affected handlers:

- `GET /api/v1/portfolios/{portfolioId}/positions`
- `PUT /api/v1/portfolios/{portfolioId}/valuations/{ticker}`
- `DELETE /api/v1/portfolios/{portfolioId}/valuations/{ticker}`

Malformed UUID input now returns `400 VALIDATION_ERROR` with zero downstream
service/store calls. The DELETE manual-valuation OpenAPI operation contains the
existing standard `400 BadRequest` contract. No new error code or schema was
introduced.

## OI_PWEL_002 remediation evidence

Technical remediation PR #218 squash-merged as
`7945221c4308e4bd11e863ff5176ac018840734c`, with final tree
`f466b7d18a5f41de5c8b1218010948755d30d77e`. Its authoritative protected PR CI
was run #592 (ID `36766334816`), with 10 of 10 required jobs successful.
Post-merge protected application workflow was `NONE_EXPECTED`. Independent
post-merge verification passed.

PR #216 is `CLOSED_SUPERSEDED_UNMERGED`: its OI_PWEL_002 patch was valid but
its CI was blocked by the newly published Next.js advisory. Replacement PR
#218 carried the byte-identical approved PWEL patch on the security-fixed
`develop` baseline. This record does not rewrite or delete PR #216 history.

## Rejected adversarial hypotheses

The following outcomes were NOT_DEMONSTRATED by this module audit. This is not
a claim of impossibility.

```text
CROSS_USER_IDOR=NOT_DEMONSTRATED
ACCEPTED_OVERSELL=NOT_DEMONSTRATED
PARTIAL_SELL_WAC_REPRICING=NOT_DEMONSTRATED
WAC_DIVERGENCE=NOT_DEMONSTRATED
UNSTABLE_POSITION_ORDERING=NOT_DEMONSTRATED
REVERSAL_EFFECTIVE_DATE_OFF_BY_ONE=NOT_DEMONSTRATED
IMPORTED_SELL_BYPASS=NOT_DEMONSTRATED
STALE_VALUATION_AFTER_REOPEN=NOT_DEMONSTRATED
DECIMAL_FRONTEND_FLOAT_FINANCIAL_DIVERGENCE=NOT_DEMONSTRATED
RACE_FINANCIAL_CORRUPTION=NOT_DEMONSTRATED
```

## Dependency-security publication detour

During OI_PWEL_002 publication, protected CI surfaced
GHSA-vcvr-r3jv-pc5j in Next.js 16.3.4. It was independently remediated in PR
#217, squash `0621e8c90711593248091d4f3b0eef59c2a1db38`. The resulting
dependency baseline is `next=16.3.6`, `react=19.2.7`, and
`react-dom=19.2.7`; the dependency-security blocker is closed. This detour is
not a PWEL finding and this record does not claim a demonstrated OpenInvest
RCE.

## Final module disposition

```text
P0_OPEN_IN_MODULE=0
P1_OPEN_IN_MODULE=0
P2_OPEN_IN_MODULE=0
P3_CLOSED_IN_MODULE=1
P3_RESIDUAL_IN_MODULE=1

OI_PWEL_002=CLOSED
OI_PWEL_001=P3_RESIDUAL

POSITION_WAC_EFFECTIVE_LEDGER_MODULE=CLOSURE_CANDIDATE_PENDING_INDEPENDENT_REVIEW
POSITION_WAC_EFFECTIVE_LEDGER_MODULE_AFTER_REQUIRED_GATES=CLOSED_WITH_P3_RESIDUAL
RESIDUAL=OI_PWEL_001
RESIDUAL_OWNER_MODULE=REPLAY_SNAPSHOT_RESOURCE_AUDIT
FORMAL_MODULE_CLOSED=NO
REPOSITORY_WIDE_AUDIT=ONGOING
```

This candidate closes neither the repository-wide audit nor the future Replay
/ Snapshot / Resource audit. Formal module closure requires this record to
merge and to pass independent post-merge verification.
