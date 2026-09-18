
| Field | Value |
| --- | --- |
| Version | 0.1.1 |
| Status | Complete / merged through PR #53 at `544ad8cc7371caf93913ea7716f3feb68be0ea44` |
| Supersedes | None; follows merged Stage 3.23 deletion-marker control-plane proposal |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.23 privacy proposals; Stage 3.21 inventory |

## Purpose

acceptance before any privacy-lifecycle implementation may start. Stages 3.20 through 3.23 provide
design evidence, but none is a deployed control, production discovery, provider attestation,

This dossier makes that boundary testable. It identifies the evidence that exists only as a
repository-derived proposal, the evidence that must be independently produced, the questions a
authorization.

## Controlling Decision Boundary

The following conditions are non-negotiable:

1. identity and personal data must be deleted completely;
2. the person-to-financial link must be irreversibly destroyed before retained financial history can
   be called Anonymous Financial History;
3. managed recoverable backups must be encrypted and destroyed within 90 days;
4. minimum redacted audit/verification evidence is retained for ten years without forming an identity
   map or extending backup retention;
5. a restore stays isolated and non-serving until independent marker evidence and destroyed-key
   unavailability are verified; and
   explicit Principal Architect acceptance.

No field removal, cascade, soft-delete state, hash, UUID, encrypted media, provider console, log
line, successful request, or verbal assurance is sufficient evidence by itself.

## Evidence Status Register


A missing entry is a blocking evidence gap, not an assumption that the corresponding data or control is
absent, safe, or compliant.


The formal review must answer each question with evidence, an accountable owner, a decision, and
residual-risk treatment:

| ID | Question | Minimum acceptable proof | Blocker if unresolved |
| --- | --- | --- | --- |
| SR-01 | Can any current or recoverable artifact reconnect a deleted identity to retained financial history? | Field-level lineage and direct/indirect correlation analysis covering primary, replica, backup, export, log, queue, cache, analytics, CI, and operator surfaces. | No Anonymous Financial History or completion claim. |
| SR-02 | Does the erasure boundary prevent export, recreation, or administrator recovery of destroyed material? | Provider capabilities, custody policy, credential lifecycle, quorum/break-glass design, and independently verifiable destruction semantics. | No key-custody selection, destruction, or migration design approval. |
| SR-03 | Can a marker writer fabricate, omit, roll back, or split-view a releaseable marker snapshot? | Separately administered writer/signer credentials, threshold or external attestation, complete watermark protocol, compatibility/freshness rules, and negative tests. | No restore release or marker implementation authorization. |
| SR-04 | Does replay target only the required recovered material without becoming a person lookup? | Field schema, replay-binding design, restricted access policy, and correlation/redaction analysis. | No marker schema or restore procedure approval. |
| SR-05 | Can partial failure, duplicate intent, race, cancellation, or outage convert an incomplete deletion to completion? | Monotonic state machine, serialized-intent binding, fault-injection results, escalation authority, and audit evidence design. | No lifecycle API, migration, or job authorization. |
| SR-06 | Can any restore serve traffic, run jobs, use production credentials, or release early before proof? | Network/credential isolation, restore/release role separation, runbook, release checklist, and adversarial rehearsal. | No restore or disaster-recovery approval. |
| SR-07 | Are all managed recoverable copies encrypted, inventoried, bounded to 90 days, and evidenced as destroyed? | Provider inventory, retention/deletion evidence, reconciliation, exception/hold policy, and periodic controls. | No backup-compliance or deletion-completion claim. |
| SR-09 | Are privacy effects resilient to malicious or mistaken insiders and support/operations access? | Role/capability matrix, access reviews, support/export controls, monitoring, incident response, and organizational-access analysis. | No production operations authorization. |
| SR-10 | Are legal hold, incident, and regulatory requirements reconciled without preserving a hidden recovery path? | Approved legal/security decision, explicit exception model, retention evidence, and fail-closed behavior. | No policy or implementation approval. |
| SR-11 | Can only a freshly authorized subject initiate or cancel a deletion lifecycle without anti-forgery, replay, enumeration, or status-oracle abuse? | Fresh re-authentication/factor design; anti-forgery/CSRF boundary; rate-limit and abuse controls; non-oracular response and timing model; cancellation rules that reject a revoked session or opaque identifier alone; negative tests and redacted audit evidence. | No deletion/cancellation API, lifecycle migration, or ADR-008 recommendation. |

## Evidence Package Rules

Every evidence item submitted to the formal review must identify its source, collection time,
environment, owner, retention/access classification, integrity protection, limitations, and the
review question it answers. Evidence must be independently reproducible or externally auditable
where a single application or operator could otherwise forge it.

The review must reject evidence that contains raw personal data, credentials, raw request/import
payloads, a usable identity map, or a marker lookup capability. Redaction must not conceal whether
a required control exists, its version, its scope, or a verification failure.

A provider may be evaluated but is not selected by this dossier. Any provider proposal needs a
separate impact record covering deletion semantics, regional copies, durability, support, incident
handling, audit privacy, cost, exit, credential recovery, administrative recovery, and contractual
evidence access.

## Readiness Outcomes

This stage permits only three honest readiness outcomes:

| Outcome | Meaning | Allowed next action |
| --- | --- | --- |
| blocked_evidence | A required item is absent, unverifiable, stale, or conflicts with a higher-priority source. | Obtain the missing evidence through separately approved design, provider, legal, or operations work; do not start implementation. |

accepted_for_decision. It is unavailable while any mandatory predecessor input, including SR-11 or
provider marketing claim, or local test run cannot do so.

## Review Record Requirements


1. the immutable scope and source baseline;
2. named accountable security, architecture, privacy, operations, and legal participants;
3. each SR question, evidence link, evidence limitations, finding, severity, owner, and due date;
4. explicit treatment for every residual risk: accepted by whom, reduced, transferred, or blocked;
5. provider, backup, restore, and organizational-access assumptions;
6. an adversarial restore rehearsal result, including stale, missing, conflicting, forged,
   reordered, and unavailable marker/proof states;
7. an ADR-008 recommendation distinct from the Principal Architect acceptance decision; and
8. a clear statement that no deployment, migration, provider selection, or runtime change is
   authorized until the required separate proposal and merge gate exist.
9. the SR-11 authority, anti-forgery, rate-limit, cancellation, and status-oracle evidence; and
10. the SR-12 migration/data-state and operations-runbook evidence.

## Scope Boundary

This stage adds documentation only. It does not change Go, Next.js, Python, OpenAPI, PostgreSQL,
Redis, migrations, CI, infrastructure, credentials, backups, providers, operational access, or
provider trial, backup deletion, restore, key operation, account deletion, anonymization, or
ADR-008 acceptance.

Canonical record: PR #53.
`544ad8cc7371caf93913ea7716f3feb68be0ea44`. Its successor, Stage 3.25, defines the evidence
