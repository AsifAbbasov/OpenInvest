# Architecture Decision Records — compact ADR / MADR
Accepted ADRs are immutable; a changed choice needs a separately approved superseding decision, not historical rewriting. Documents 42–43 retain their hierarchy.
ADR-001–007: accepted; [ADR-008](ADR-008-privacy-lifecycle-erasure-and-restore.md): proposed and non-normative; [ADR-009](ADR-009-deterministic-portfolio-ledger-ordering-and-wac.md): accepted after approved Issue #136 / protected PR #137.
[Original ADR-010](ADR-010-email-ownership-before-account-activation.md) is migration-hash-pinned and historically proposed; the [approved addendum](ADR-010-email-ownership-approval-candidate.md) is separate. No in-place modification to accepted/pinned originals.
**Template for NEW decisions:** target 10, maximum 15 physical lines including title; combine fields when needed, do not invent alternatives.
`# ADR-NNN — Decision title`
`Status / Version / Owner / Date / Supersedes:` distinguish proposed, accepted and obsolete.
`Context:` relevant conditions, constraints and existing authority.
`Root Cause:` why the defect or decision pressure arose.
`Problem:` observable failure or unmet requirement.
`Risk:` impact if unchanged.
`Alternatives:` only genuinely considered options and their costs.
`Decision / Rationale:` what was selected and why alternatives lost.
`Consequences / Trade-offs:` benefits, compromises and accepted risks.
`Implementation / Verification / Residuals:` code/PR, exact-head tests/CI, limitations and separate approval gates.
