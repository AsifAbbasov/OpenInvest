# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.9 |
| Status | Active / no current questions |
| Owner | Principal Architect |
| Supersedes | Informal TODOs for architecture decisions |
| Dependencies | Document 43 |
| Last Review Date | 2026-09-12 |
| Next Review Date | Before Feature 3D runtime activation or broader T-Invest source/use rights, any new architecture decision affecting provider/public activation, external market valuation, privacy lifecycle, tax basis, imported SELL semantics, or Stage 3.78+ scope, or 2026-12-19, whichever comes first |

> **Document role — decision-question register**
>
> This register preserves unresolved and resolved questions, ADR admission evidence, and historical decision context.
> An open question is not current product state; a resolved question is not runtime authority; a proposal is not approval; closing a question does not itself implement runtime behavior.
> Current human-readable product/runtime truth lives in [`SOURCE_OF_TRUTH.md`](SOURCE_OF_TRUTH.md).

## Current questions

None.

Current product/runtime state is maintained in [`SOURCE_OF_TRUTH.md`](SOURCE_OF_TRUTH.md). Provider/source-use decisions are maintained in [`registries/DATA_SOURCE_REGISTRY.md`](registries/DATA_SOURCE_REGISTRY.md). Implementation chronology is maintained in [`IMPLEMENTATION_LOG.md`](IMPLEMENTATION_LOG.md), and future sequencing in [`ROADMAP.md`](ROADMAP.md).

## Resolved architecture questions

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering

- Owner: Principal Architect / Portfolio & Analytics
- Due date: 2026-09-06 for the Stage 3.70 planning decision
- Affected decisions: immutable transaction-ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL and oversell atomicity; backdated rebuild semantics; SELL-safe snapshot methodology boundary; future broker-import SELL ordering boundary
- ADR: `ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics`
- Lifecycle: `RESOLVED / CLOSED COMPLETED` — Stage 3.70 / ADR-009 became canonical through protected squash merge PR #137 at `b772e52221fbb694b3116bd1b579db99d4e56302`; GitHub Issue #136 is closed with `state_reason=completed`.
- Follow-on runtime: Stage 3.71 was separately authorized and later completed through PR #138; resolution of Issue #136 does not retroactively claim that the Issue itself authorized runtime implementation.
- GitHub: `https://github.com/AsifAbbasov/OpenInvest/issues/136`

The original Issue → ADR → review → explicit human approval → architecture update sequence remains preserved. Stage 3.71 runtime required and received a separate implementation authorization after Stage 3.70 activation.

## Admission process

A new unresolved architecture question must have an Issue, owner, due date, affected decisions, and proposed ADR. The sequence is Issue → ADR → Review → Approval → Architecture Update. Production-code TODOs may not substitute for this register.
