# Valuation / Cash Flow / XIRR Adversarial Audit Closure
## Module
```text
MODULE=VALUATION_CASH_FLOW_XIRR
CLOSURE_CANDIDATE=CLOSED_WITH_P3_RESIDUAL
FORMAL_MODULE_CLOSED=NO
```
This is the formal closure candidate for the Valuation / Cash Flow / XIRR
module. Formal closure becomes `YES` only after this document is merged to
protected `develop` and independent post-merge verification confirms its exact
content and the preserved remediation state.
## Audit Baseline
The original audited protected baseline was:
```text
SHA=f4c1f4e251165a2fc1b7f81a262a1647141c97a7
TREE=edbde827ffd85ea955f0b80931bc9a9ee1048be5
```
The remediation merge baseline was:
```text
SHA=d67566c73c90481329b51812a0c62d67d12d5c1e
TREE=f0bf1534f0b84b8bd3ce72a635645fa20ddeff94
```
## Audit Scope
The adversarial audit covered:
- portfolio cash-flow truth and DEPOSIT/WITHDRAWAL signs;
- BUY/SELL internal-flow classification;
- dividends, coupons, and manual income-expense semantics;
- correction/reversal interaction, as-of truth, and future leakage;
- `USER_SUPPLIED` valuation, valuation generation identity, and close/reopen safety;
- terminal portfolio value, XIRR, ACT/365, multiple roots, and non-convergence;
- NaN/Inf, numerical-failure, Decimal/float-boundary, SQL, and input safety;
- cross-user isolation, concurrency, summary/returns, OpenAPI/frontend parity,
  and resource amplification.
## Audit Results
```text
P0_CONFIRMED_FINAL=0
P1_CONFIRMED_FINAL=0
P2_CONFIRMED_FINAL=0
```
### OI-XIRR-001 - P3 closed
Legacy snapshot ROI persistence could fabricate a return because total snapshot
value included idle cash while `investedCapital` represented cumulative BUY
outflow.
Remediation did not invent a replacement nominal/real-return methodology.
Legacy ROI is `UNAVAILABLE`; historical numeric residue is retained but
non-authoritative; new snapshots use methodology v2; public
`nominalReturnRate` and `realReturn` remain `null`; XIRR remains the activated
return metric.
```text
OI_XIRR_001=CLOSED
```
### OI-XIRR-002 - P3 closed
The public contract did not clearly distinguish local acquisition-cost snapshot
value from the exact-date market terminal value used for XIRR.
`PortfolioSummary.totalValue` is now explicitly documented as a local
cost-basis snapshot value, while `terminalPortfolioValue` is documented as the
exact-date market terminal value. No financial calculation changed for this
finding.
```text
OI_XIRR_002=CLOSED
```
## Rejected and Non-Findings
```text
OI_XIRR_003=REJECTED_INTENTIONAL_CONTRACT
Reason=future explicit BusinessDate acceptance was not demonstrated to violate
canonical contract and no future-row leakage beyond asOfDate was demonstrated.

OI_XIRR_004=REJECTED_INTENTIONAL_CONTRACT
Reason=corrections are intentionally retroactive truth while reversals become
effective on their explicit effectiveDate under canonical Stage 3.74 semantics.

OI_XIRR_005=REJECTED_FALSE_POSITIVE
Reason=the real HTTP/Service boundary rejects non-canonical manual valuation
date forms before persistence.

OI_XIRR_006=REJECTED_INTENTIONAL_CONTRACT
Reason=bounded scale-8 Half-Even WAC behaviour is canonical ADR-009 methodology.
```

## Residual

```text
OI_XIRR_007=DUPLICATE_EXISTING_RESIDUAL
OWNER=OI_PWEL_001
SEVERITY=P3
DISPOSITION=DEFER_TO_REPLAY_SNAPSHOT_RESOURCE_AUDIT
```

Repeated effective-ledger/projection materialization can amplify resource work;
reconciliation profiling strengthened evidence for that debt. It is not a
separate Valuation/XIRR finding, remains governed by `OI_PWEL_001`, and no
financial-correctness or cross-user-integrity failure was demonstrated.

The closure target is therefore `CLOSED_WITH_P3_RESIDUAL`.

## XIRR Conclusion

No confirmed numerical XIRR defect was reproduced within the executed
adversarial scope. This is not a mathematical proof for every possible
cash-flow vector.

```text
XIRR_BASIC_CORRECTNESS=PASS
XIRR_MULTIPLE_ROOT_POLICY=PASS
XIRR_NONCONVERGENCE_FAIL_CLOSED=PASS
XIRR_NAN_INF_SAFETY=PASS
XIRR_OVERFLOW_SAFETY=PASS
XIRR_ACT_365=PASS
XIRR_CURRENT_BASELINE_ORACLE=PASS
```

## Other Final Invariants

```text
CASH_FLOW_SIGN_CORRECTNESS=PASS
CORRECTION_REVERSAL_DOUBLE_COUNTING=NO_CONFIRMED_DEFECT
AS_OF_FUTURE_DATA_LEAKAGE=NO_CONFIRMED_DEFECT
VALUATION_GENERATION_ISOLATION=PASS
STALE_VALUATION_AFTER_REOPEN=NO_CONFIRMED_DEFECT
CROSS_USER_ISOLATION=PASS
DECIMAL_AUTHORITATIVE_PATH=PASS
SUMMARY_RETURNS_FINANCIAL_SEMANTICS=PASS_AFTER_CONTRACT_CLARIFICATION
```

## Remediation Record

```text
PR=220
PR_HEAD=a09796de60f91c072bf302608cb2b08cf6de46c2
MERGE_SHA=d67566c73c90481329b51812a0c62d67d12d5c1e
MERGE_TREE=f0bf1534f0b84b8bd3ce72a635645fa20ddeff94
MIGRATION=000014_valuation_xirr_legacy_return_unavailable
SNAPSHOT_METHODOLOGY_CURRENT=stage-03-71-position-cost-snapshot-v2
SNAPSHOT_METHODOLOGY_LEGACY=stage-03-71-position-cost-snapshot-v1
LEGACY_RETURN_STATUS=UNAVAILABLE
```

## Verification Record

Approved evidence: complete Go tests, Go vet, Go race CI, migration validation,
PostgreSQL migration apply/down/reapply, OpenAPI validation, frontend tests
115/115, frontend typecheck/build, dependency security scan, Go vulnerability
scan, Python tests, ACTIVE-R2 integration, and ACTIVE-R2 race all passed.

```text
RUN_ID=37046545148
REQUIRED_JOBS=10
SUCCESS=10
FAILED=0
```

## Post-Merge Workflow

Post-merge workflow: `NONE_EXPECTED`.

The CI workflow has `pull_request`, `schedule`, and `workflow_dispatch`
triggers, with no push trigger for `develop`. This record does not imply that
post-merge CI executed.

## Known Test-Harness Limitation

Earlier reconciliation broad local `internal/httpapi` race execution encountered
existing one-second auth-test timeout sensitivity. This was not classified as a
data race. Protected PR CI subsequently completed the full repository Go race
job successfully for the remediation head.

## Final Candidate State

```text
OI_XIRR_001=CLOSED
OI_XIRR_002=CLOSED

OI_XIRR_007=P3_RESIDUAL_OWNED_BY_OI_PWEL_001

MODULE_CLOSURE_CANDIDATE=CLOSED_WITH_P3_RESIDUAL
FORMAL_MODULE_CLOSED=NO
REPOSITORY_WIDE_AUDIT=ONGOING
NEXT_MODULE_AFTER_FORMAL_CLOSURE=REPLAY_SNAPSHOT_RESOURCE
```
