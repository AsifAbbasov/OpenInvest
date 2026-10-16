# Stage 3.25 - Privacy Security Review Evidence-Collection Plan

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-25-PRIVACY-SECURITY-EVIDENCE-COLLECTION-PLAN |
| Version | 0.1.0 |
| Status | Active / proposal only |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.24 Security Review readiness dossier |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.24 privacy proposals |
| Last Review Date | 2026-08-18 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Before evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

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


## SR-to-EC execution trace and independent test catalogue — 2026-10-16

This is a **proposed evidence request**. No production evidence collection, provider access,
data deletion, Security Review, legal opinion, hosted restore, or destructive test has occurred.
The catalogue expands—not substitutes for—the complete EC-01–EC-10 minimum proof clauses
above. Architecture owners must freeze the system/source version and appoint a distinct verifier
before collecting each restricted item. Test IDs below are **NOT_EXECUTED**, except the
read-only source inspection and baseline verification expressly called out as `REPO_OBSERVED`.

### Threat → source → candidate control → test → external evidence → gate

| Requirement → EC | Source anchor / threat | Candidate control | Independent test and evidence route | Non-waivable acceptance gate |
| --- | --- | --- | --- | --- |
| SR-01 → EC-02, EC-10 | migration 000001 link; 000003–000004 labels/hashes; audit/exports and rare holdings reconstruct identity | Per-subject protected join + field-level lineage, erase secondary correlations; restricted redaction/approval for immutable ledger/audit conflict | **TV-01** synthetic adversarial join against old/new SQL, unique holding patterns, exported CSV, raw notes, trace and broker-file hashes. Repo: columns observed. Isolated: synthetic fixtures. External: sampled provider/operator/export lineage and privacy review. | No reasonable reidentification path in any retained copy; indeterminate ⇒ BLOCKED. |
| SR-02 → EC-03, EC-10 | No KMS boundary; alternative wrap, recovery or admin escrow defeats erasure | Non-exportable subject-scoped material, independent destruction proof and dual control | **TV-02** key-version/recovery/replica/alias/escrow challenge; duplicate delete; forged/scheduled-vs-terminal proof. Repo: absent. Isolated: simulated custody driver only. External: real provider attestation, IAM, deletion semantics and exit assessment. | Independent verifier authenticates terminal destruction of *all* effective versions/copies. |
| SR-03 → EC-04, EC-10 | No append-only marker log; privileged writer can omit or reorder states | Externally witnessed sequence/watermark + independently signed canonical checkpoint | **TV-03** corrupt leaf, omit page, replay older signed root, split view across two restores, forge signer and conflicting generation. Repo: absent. Isolated: pure verifier model. External: signer/WORM/quorum topology and replay rehearsal. | Completeness and non-equivocation independently proven, not only signature validity. |
| SR-04 → EC-04, EC-06 | Historic copy of plaintext user↔subject bridge | Restore-only random replay binding in sealed zone, reconcile before serving | **TV-04** real test fixture representing pre-migration backup, missing mapping, twice-replayed marker, wrong subject binding and stale checkpoint. External: actual backup media and credential/firewall test. | No resurrected mapping; no false foreign deletion; no serving until verified. |
| SR-05 → EC-05 | Current user states, but no deletion API/transactions; stateless access JWT | Serialized lifecycle generation, revocation plus protected-path/commit fence, idempotent saga | **TV-05** 32 bounded simulated adversarial operations: login/refresh/financial write/deletion intent/cancel/worker, injected COMMIT ambiguity and custody outage. External: reviewer-approved grace, incident/repair on-call and IAM. | No post-fence authorized work, single irreversible winner, no false completion. |
| SR-06 → EC-06, EC-10 | docker-compose local resources, absent production release IAM | Sealed subnet/account, disabled jobs/SMTP, restore-scoped credentials, independent release | **TV-06** deny ingress/egress, production credential use, background outbox execution, DNS/webhook and false release attempt in restore rehearsal. External: deployed rules and sealed release log. | No network/credential/job before every restore predicate passes. |
| SR-07 → EC-07, EC-09 | Local volumes/AOF; no provider copy chain or verified 90-day destruction | Copy lineage register incl. WAL/replicas, encrypted copy TTL and destruction attestation | **TV-07** missing snapshot/export/replica, delayed provider purge, forged receipt, 90-day boundary, hold conflict. External: provider logs/copy inventory and independent reconciliation. | Each recoverable managed copy accounted/destroyed by contractual maximum; any unknown ⇒ BLOCKED. |
| SR-08 → EC-02, EC-09 | audit actor=user UUID, target=session UUID, dedupe session_id, request/trace IDs | Minimally identifying event v2 + append-only integrity, separately approved redaction overlay | **TV-08** 10-year audit reconstructibility challenge from an old partition; prove event integrity and actor/target nonlinkability. External: legal field minima, operational logging inventory and privacy sample. | Audit proof remains complete while no identity reconstruction path survives. |
| SR-09 → EC-08, EC-03 | No custody/marker/release IAM; support export potential | Role domains with distinct credentials, quorum, break-glass time limits and monitoring | **TV-09** malicious app/support/DB admin, signer impersonation, shared parent IAM, escrow recovery and release bypass. External: real IAM principals, reviews, incident workflow. | No single actor can both cause deletion and forge success/recover linkage. |
| SR-10 → EC-09, EC-10 | Doc43 retention and deletion terms vs unknown jurisdiction/hold requirements | Signed legal conflict matrix and fail-closed state; no runtime exception until approved | **TV-10** approved legal-state decision-table walkthrough; retention boundary/resolution of contradictory obligation. External: counsel and privacy officer. | Written legally qualified non-conflicting decision; otherwise **BLOCKED**. |
| SR-11 → EC-01, EC-05 | Current login and CSRF; no fresh delete/cancel authority | Reauth factor tied to one-use action, CSRF, generation and restricted status receipt | **TV-11** proof replay, CSRF swap, stolen lifecycle ID, cross-account cancellation, revoked token, clock/grace boundary, enumeration/timing oracle. External: product security and privacy decision. | Only fresh authorized subject can initiate/cancel and receipts do not misrepresent completion. |
| SR-12 → EC-05, EC-10 | Stage2 migration lifecycle; old FK, older backups, old workers | Expand/Populate/Switch/Validate/Contract with pre-irreversibility rollback and post-irreversibility forward recovery | **TV-12** old binary writer, partial backfill, stale JWT, backup restore predating cutover, async workers; compare Decimal(28,8), BusinessDate, reversal history and snapshots. External: signed migration/runbook rehearsal, cutover/release decision. | No old bridge reconstructed and canonical immutable financial mathematics unchanged. |

All twelve test IDs are **proposed test specifications**, not evidence of a run; a
`TV-05` target of 32 is a bounded design suggestion, **not** an executed count or a guarantee.
Every trace row retains original EC clause-level obligations.

### Mandatory EC proof registry — explicit classifications

| EC | Repository evidence now | Feasible isolated verification later | External/provider/production dependency | Legal/operations/Principal Architect dependency |
| --- | --- | --- | --- | --- |
| EC-01 | HTTP refresh/logout/CSRF code and lack of deletion contract **REPO_OBSERVED** | TV-11 synthetic HTTP, negative authorization | Real anti-abuse/observability configuration | Fresh-factor, grace, cancellation UX/contract decision **BLOCKED** |
| EC-02 | Stage3.21 inventory + migration/audit fields **REPO_OBSERVED** | TV-01/08 synthetic join and audit transform | Primary/replica/log/export/CV/data-subject lineage **MISSING** | Independent privacy/anonymity/legal reviewer **MISSING** |
| EC-03 | Key hierarchy not implemented **REPO_OBSERVED** | TV-02 mock-only semantic faults | Provider terminal destruction, non-export/copies/IAM **MISSING** | Custody policy, provider and approver decision **MISSING** |
| EC-04 | No marker schema **REPO_OBSERVED** | TV-03/04 completeness+replay model | Durable signer/witness/quorum and checkpoint **MISSING** | Separate integrity authority/security review **MISSING** |
| EC-05 | User states, migrations, retention source **REPO_OBSERVED** | TV-05/12 SQL fault injections in disposable PG | Operational runbook, real worker/copy integration **MISSING** | Lifecycle schema/locking, rollout/forward-recovery review **MISSING** |
| EC-06 | Local Compose only **REPO_OBSERVED** | TV-06 staging ingress/job denial | Real sealed restore + credential test **MISSING** | Restore release owners/quorum **MISSING** |
| EC-07 | No deployed managed-copy ledger **REPO_OBSERVED** | TV-07 synthetic retention copy graph | Provider backup/PITR/WAL/export expiry proof **MISSING** | Backup operator/legal retention review **MISSING** |
| EC-08 | Current audit rows and code **REPO_OBSERVED** | TV-09 simulated IAM denial | Actual IAM/access/support exports **MISSING** | Named segregation and break-glass authority **MISSING** |
| EC-09 | Document43 retention rules **REPO_OBSERVED** | TV-10 policy state-machine table | Applicable legal evidence **MISSING** | Qualified legal/privacy approval **MISSING** |
| EC-10 | Frozen source/ADR/SR index **REPO_OBSERVED** | Cross-matrix/negative-test QA | Independent evidence-item provenance/review **MISSING** | Formal Security Review + Principal Architect acceptance **MISSING** |

**Evidence labels:** `REPO_OBSERVED` = proof limited to repository source; `MISSING`
= not supplied, not evidence of absence; `NOT_EXECUTED` = no test run; `BLOCKED` =
no authority to implement/accept. Local design documentation cannot substitute for a provider
destruction attestation or an independent human sign-off.

### Evidence item integrity and minimal handoff record

For every proposed test/proof, an accountable future owner must record:

```text
evidence_id; sr_id; ec_id; clause_id; source_commit_tree; component_schema_version;
environment_class; scope_copy_ids_redacted; test_seed; test_case_hash;
command_or_provider_operation_digest; expected_invariant; observed_status;
artifact_digest_sha256; verifier_identity_role; reviewer_independence;
collected_at_utc; valid_until_utc; redaction_scope; missing_external_dependency;
decision=PASS|FAIL|BLOCKED_EVIDENCE
```

Never put email, UUID bridges, raw keys/proofs, import bodies, financial payloads,
reversible marker lookups or real production identifiers in these unrestricted artifacts.
A signature/HMAC alone does not prove the evidence source truthful; independent source,
verifier and anti-rollback checks remain required. Evidence retention itself must be
classified under the ten-year audit obligation and the distinct 90-day backup bound.

### Review sequencing / decision points

1. Principal Architect freezes SR/EC inventory, proposed custody/marker schema and source SHA;
   names independent privacy, security, custody, DBA, restore, legal and release owners.
2. Legal/privacy reviewer resolves SR-10 and financial-pattern vs immutable-ledger conflict;
   if impossible under Document 43, open separate **superseding approval request**.
3. Independently assess custody candidates, maker/checker separation, marker snapshot
   completeness and a trusted anti-rollback witness. Do not procure a provider here.
4. Approve field-level lineage disposition (identity, audit, imports, derived, logs, legacy
   backups) before allowing backfill of any live user.
5. Approve and run isolated test suite TV-01–TV-12, preserving RED evidence; then rehearse
   complete sealed restore+release protocol against a known pre-erasure backup.
6. Independent reviewers complete every EC clause; formal Security Review can commence
   only if evidence index has zero missing/blocked clauses. Principal Architect separately
   accepts or rejects ADR-008 and all future migrations/operations in a later authorized PR.

**Current recommendation: EC-01–EC-10 have uncollected mandatory proof clauses;
ADR-008 remains Proposed / BLOCKED. No test or production action is authorized by this plan.**
