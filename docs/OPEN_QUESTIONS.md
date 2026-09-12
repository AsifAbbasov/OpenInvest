# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.8 |
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

None. Stage 3.77 remains the last numbered implementation stage and is complete/canonical through PR #157 squash merge `63d916b447e91c4de54efee5c66b27cf7727be92` from final evidence head `b3bc17e710e306f9d1a4d7be95dfe56b3f7202ba` / tree `e91e28e96416d9510962a66c1399459844574bba` after protected exact-head CI run `34303895226` completed 10/10 SUCCESS. Separately governed Feature 3D is now implemented/canonical through PR #164 squash merge `247081a95a7daf33c0077c88c5f41cb2e8161865` from final evidence head `ddc5b5be36f0b5ae127c6ee430918a9dba1b453e`, tree `ec7bc9152210913b1a6ef742bddd599abee50bc7`; implementation CI #499 / run `34353668492` and evidence-head CI #501 / run `34356672643` both completed 10/10 SUCCESS, final Internal and fresh External reviews were `APPROVED`, and evidence/no-semantic-drift verification was `APPROVED`. This closes the Feature 3D implementation question only: `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` source/use remains exact-scope `CONDITIONAL-GO`, runtime activation remains NO, no live T-Invest token was used in implementation/review, no production provider traffic is authorized or claimed by Feature 3D, and broader methods, persistence/cache/polling, provider-specific public-contract expansion, ledger coupling, or operational activation require a new separately admitted decision. The existing Stage 3.77 XIRR semantics and exclusions remain unchanged. Stage 3.78+ remains not started/unauthorized.

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
