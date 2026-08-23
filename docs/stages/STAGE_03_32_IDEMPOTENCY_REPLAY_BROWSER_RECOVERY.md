# Stage 3.32 — Exact Idempotency Replay and Browser Retry Recovery

| Field | Value |
| --- | --- |
| Status | Implementation candidate; first independent review requested changes for P2-13; remediation and exact-head verification green; repeat independent review and human merge authorization pending |
| Owner | Principal Architect |
| Baseline | `develop` at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` |
| Branch | `fix/stage-03-32-idempotency-recovery` |
| Implementation PR | #67 |
| Implementation merge | Pending |
| Final remediation head before this record | `13dbf3ad06ed35bd643c6810e383713ea2463baa` |
| Final implementation/documentation head | `b5be85b48ccb22c482368bb007d15a5197ca58a8` |
| Exact-head CI | GitHub Actions #174 — SUCCESS, all six jobs passed |
| First independent review | `REQUEST CHANGES` on `57fcc25e949277a0e933f290998e41d0f7476b5c`: P2-09 CLOSED; P2-13 NOT CLOSED because browser retry slots were not principal-scoped |
| Repeat independent review | Pending on exact head `b5be85b48ccb22c482368bb007d15a5197ca58a8` |
| Human implementation merge authorization | Pending |
| Trigger | Repository-audit P2-09 and P2-13 |
| Scope | Exact original-response idempotent replay, atomic replay-artifact persistence, import retry recovery across review-token expiry, principal-isolated short-lived browser retry identity, regression and migration coverage |
| Out of scope | P2-10/P2-11/P2-12/P2-16/P2-17, all P3 findings, Stage 3.25 privacy Security Review evidence work, provider/backup retention, product-scope expansion |

## Purpose

Stage 3.32 addresses the gap between exactly-once business effect and exact observable HTTP replay.
