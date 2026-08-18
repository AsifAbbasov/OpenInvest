# Stage 3.25 - Privacy Security Review Evidence-Collection Plan

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-25-PRIVACY-SECURITY-EVIDENCE-COLLECTION-PLAN |
| Version | 0.1.0 |
| Status | Active / proposal only |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.24 Security Review readiness dossier |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.24 privacy proposals |
| Last Review Date | 2026-08-18 |
| Next Review Date | Before evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

## Purpose

Stage 3.24 establishes that formal Security Review remains blocked until independently assembled
evidence answers every mandatory question. This plan defines how future evidence is requested,
minimized, redacted, integrity-protected, independently checked, and submitted. It does not collect
evidence, inspect a production environment, select a provider, perform Security Review, or accept
ADR-008.

A repository proposal, CI result, screenshot, mutable dashboard, verbal assurance, or placeholder
cannot be converted into operational proof by listing it in this plan. Every uncollected, stale,
unverifiable, contradictory, or over-redacted item remains blocked_evidence.

## Evidence Handling Rules

Each evidence item must include a stable evidence ID; related SR question(s); scope and environment;
source system and accountable owner; collection time and applicable configuration/version; access and
retention classification; integrity method; redaction method; limitations; independent verifier; and
a decision-ready conclusion limited to what the item actually proves.

Collection must use the least sensitive source and smallest sufficient sample. It must not place raw
identity data, credentials, tokens, raw request/import bodies, key material, usable identifier maps,
or public marker lookup material in the Security Review package. Redaction cannot remove the
control's version, scope, failure state, or ability to test the claim. When safe redaction is
impossible, the item must be reviewed through a separately approved restricted-access process and
the dossier records that constraint without copying the sensitive content.

Evidence must be immutable or independently reproducible. A mutable console view, self-attested
export, or shared-secret MAC is insufficient alone where a privileged actor could forge the result.
The evidence package must preserve source provenance and detect substitution, truncation, replay,
staleness, and conflict.

## Required Evidence Workstreams

| ID | Mandatory scope | Accountable evidence owner | Minimum acceptable evidence | Independent verification | Blocking result |
| --- | --- | --- | --- | --- | --- |
| EC-01 | SR-11 deletion and cancellation authority | Security and identity owner | Fresh re-authentication/factor design; anti-forgery/CSRF boundary; rate-limit and abuse controls for replay and enumeration; non-oracular response/timing model; cancellation rules that reject a revoked session or opaque identifier alone; negative tests; and redacted audit evidence. | Security reviewer tests each claimed abuse and cancellation failure mode without using production identities. | No deletion/cancellation lifecycle, migration, or ADR recommendation. |
| EC-02 | SR-01 and SR-08 data lineage, anonymity, and minimum audit evidence | Privacy/data owner | Field-level disposition and direct/indirect linkage analysis for primary, replica, backup, export, log, queue, cache, analytics, CI, and organizational surfaces; plus a ten-year minimal-redacted evidence schema, restricted-access review, retention policy, and sampled records that test reversibility. | Independent privacy reviewer traces samples, tests the sampled record/linkage claim, and records unresolved external surfaces. | No Anonymous Financial History, audit-retention policy, or lifecycle-completion claim. |
| EC-03 | SR-02 custody and destruction proof | Custody/operations owner | Provider-neutral capability matrix plus provider-specific evidence after a provider is proposed: deletion semantics, non-exportability, administrator recovery, independently verifiable attestation, replication, residency, retention, cost, exit, and support; custody policy; credential lifecycle; and quorum/break-glass design. | Independent verifier validates attestation semantics, custody-policy separation, credential lifecycle, and access assumptions. | No provider decision, key destruction, or migration approval. |
| EC-04 | SR-03 and SR-04 marker integrity and replay | Security/control-plane owner | Field schema, replay-binding design, restricted-access policy, redaction/correlation analysis, separately administered writer/signer model, threshold or external attestation, completeness/freshness protocol, compatibility rules, and negative snapshot/replay tests. | Security reviewer tests omission, rollback, split-view, stale, forged, correlation, compatibility, and person-lookup cases. | No marker or restore procedure approval. |
| EC-05 | SR-05 and SR-12 lifecycle migration/data state and integrated operations runbook | Architecture/data owner | Non-authorizing monotonic data-state design with serialized-intent binding, constraints, locking, access controls, retention, backfill/rollback limits, forward recovery, non-recreation proof, fault-injection results, escalation authority, and redacted audit-evidence design; plus an integrated least-privilege runbook for authorization, proof, backup expiry, isolated restore, release, outage/escalation, and evidence redaction. | Architecture and security reviewers test partial-failure, duplicate intent, race, cancellation, outage, and runbook handoff paths. | No migration, job, lifecycle implementation proposal, formal Security Review recommendation, or ADR-008 decision. |
| EC-06 | SR-06 restore/release | Restore/operations owner | Network and credential isolation design, least-privilege release roles, sealed restore runbook, release checklist, and adversarial rehearsal record. | Independent restore-release reviewer verifies no serving path exists before checks pass. | No restore or disaster-recovery approval. |
| EC-07 | SR-07 managed copies | Backup/operations owner | Inventory for backups, replicas, PITR/WAL, exports, and recovery media; encryption, 90-day destruction, reconciliation, exception/hold, and periodic evidence. | Independent reconciliation against provider/system-of-record evidence. | No backup-compliance or deletion-completion claim. |
| EC-08 | SR-09 operations and insider boundary | Operations/security owner | Role/capability matrix, access reviews, export/support restrictions, monitoring, incident response, and organizational-access analysis. | Security reviewer validates separation and break-glass limits. | No production operations authorization. |
| EC-09 | SR-10 legal/retention boundary | Legal/privacy/security owner | Signed legal, regulatory, and security decision for legal hold, incident, retention, and exception handling; explicit fail-closed behavior; and retention evidence that preserves no hidden recovery path. | Legal and security reviewers independently confirm policy-to-control mapping, retention evidence, and fail-closed behavior. | No policy, lifecycle completion, or implementation approval. |
| EC-10 | All SR questions and formal review record | Principal Architect and Security Review coordinator | Immutable scope baseline, complete evidence index, findings, residual-risk treatment, recommendation, and separate Principal Architect decision. | Human Security Review verifies evidence completeness before any recommendation. | No ADR-008 acceptance or implementation authorization. |

No owner may independently approve the evidence it produces when that owner can change the source
control. Separation must be recorded per item; a conflict or unavailable independent verifier is
blocked_evidence.

Each clause of a mapped SR minimum proof is mandatory, even when its evidence is distributed across
multiple EC items. The EC-10 index must map every clause to its evidence IDs, collection and review
freshness, and independent-verifier pass, fail, or blocked_evidence result. A missing, stale, failed,
or blocked clause leaves its SR question blocked_evidence. In particular, SR-12 cannot be complete
from a data-state design or a partial runbook alone: EC-05 must contain the complete integrated
runbook and its verified inputs from the relevant custody, marker, restore, backup, operations, and
legal workstreams.

## Collection Sequence

1. The Principal Architect freezes the target scope, source baseline, evidence IDs, owner assignments,
   independent verifiers, and restricted-access procedure before collection.
2. Each owner submits a minimized evidence request describing the claim, source, collection method,
   expected sensitive fields, retention, and the SR question it addresses.
3. Security and privacy representatives approve or reject the request's minimization and handling
   plan. Rejection is not a permission to collect a broader sample.
4. The owner collects only the approved evidence, records provenance and integrity metadata, and
   applies the approved redaction or restricted-access path.
5. The independent verifier checks completeness, freshness, authenticity, and limitations, then
   records pass, fail, or blocked_evidence without editing the source evidence.
6. The Security Review coordinator builds an evidence index that maps every SR question and each of
   its minimum-proof clauses to verified items and explicitly lists every missing, stale,
   contradictory, failed, or blocked clause.
7. Only after every mandatory entry and minimum-proof clause is independently verified as current
   and no blocker remains may a human formal Security Review begin. This plan cannot change that
   condition.

## Invalid Evidence and Safe Failure

The package must reject an item that is missing provenance, integrity, owner, verifier, scope,
version, limitation, or required redaction. It must reject a claim that uses absence from source
code as proof of absence in production. It must reject a current-state assertion derived solely from
a design document, local test, CI job, or mutable dashboard.

A failed collection, unavailable provider, missing backup inventory, unavailable verifier, legal
uncertainty, or evidence conflict remains visible as blocked_evidence. It must not be converted into
an accepted risk, a fabricated negative result, a reduced sample, or an implementation exception.

## Evidence Index Template

The future restricted evidence index must have one row per evidence item:

| Field | Required content |
| --- | --- |
| Evidence ID and SR mapping | EC ID plus every SR question supported, or all SR questions for the formal review record; no unsupported scope expansion. |
| SR minimum-proof clause status | Exact Stage 3.24 clause, supporting evidence IDs, collection and review freshness, independent-verifier result, and pass/fail/blocked_evidence status. |
| Claim and scope | Precise, falsifiable claim; environment, system boundary, and versions. |
| Source and provenance | Source system, collection method/time, owner, integrity reference, and reproducibility path. |
| Sensitive-data handling | Data classification, minimization, redaction/restricted-access method, retention, and access authority. |
| Independent verification | Verifier identity/role, method, result, limitations, and conflict declaration. |
| Decision effect | Which blocker is resolved, remains blocked, or is contradicted; no implementation authorization. |

## Scope Boundary

This stage adds documentation only. It does not collect production data, access providers, alter
credentials, inspect backups, run restores, select a provider, conduct a formal Security Review,
accept ADR-008, change Go, Next.js, Python, OpenAPI, PostgreSQL, Redis, migrations, CI,
infrastructure, backups, operations, or product behavior.

The immediate next action is mandatory strict review of this plan. Actual evidence collection begins
only under an explicitly approved restricted-access process with accountable owners and independent
verifiers. Privacy-lifecycle implementation remains prohibited until the formal Security Review and
ADR-008 acceptance are recorded.
