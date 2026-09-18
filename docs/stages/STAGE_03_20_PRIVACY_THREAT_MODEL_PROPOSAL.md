# Stage 3.20 - Privacy Lifecycle Threat Model Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-20-PRIVACY-THREAT-MODEL-PROPOSAL |
| Version | 0.1.2 |
| Status | Complete / merged through PR #49 at `849d934906f878a6d79ba89e940e5ba470e64c09` |
| Supersedes | None; follows merged Stage 3.19 privacy security/ADR proposal |
| Dependencies | Documents 42-43; ADR-005; ADR-006; proposed ADR-008; Stage 2 ER model and migration strategy; Stage 3.17-3.19 privacy proposals |

## Purpose

Stage 3.19 defined a provider-neutral future control model for privacy-lifecycle erasure. This stage
turns its required threat-model evidence into a reviewable proposal. It identifies the assets,
trust boundaries, adversaries, failure paths, security properties, residual risks, and evidence that

implementation authorization. It adds no API, schema, key, backup, operational procedure, or
runtime behavior.

## Authoritative Boundary

Documents 42-43 require complete identity deletion, irreversible removal of the person-to-financial
link, Anonymous Financial History, no reasonable technical or organizational reidentification path,
10-year audit retention, and encrypted-backup destruction within 90 days. Those accepted rules have
priority over this proposal.

The current deployment does not implement the proposed controls. `identity.user_investment_links`
is a plain reversible foreign-key relationship. No per-subject erasure-key hierarchy, durable
deletion-marker control plane, deletion-completion state machine, backup-destruction evidence, or
restore serving gate exists. A present-day cascade, deleted row, or encrypted backup must not be
claimed as cryptographic erasure or Anonymous Financial History.

## Scope and Non-Goals

In scope:

- future privacy-lifecycle assets, trust boundaries, adversaries, abuse cases, and fault paths;
- privacy-specific security invariants and evidence gates; and

Out of scope:

- acceptance of ADR-008 or a statement of security, legal, or production compliance;
- an account-deletion or cancellation API, authentication-factor policy, user experience, or data
  model;
- PostgreSQL migrations, RLS, retention jobs, queues, workers, provider integration, KMS/Vault
  selection, backup configuration, restore execution, or operational access grants;
- physical deletion of immutable financial transactions or snapshots; and
- market data, financial calculations, tax, mobile, premium, email, or public API work.

## Assets and Required Invariants

| Asset or property | Required invariant after a completed future deletion |
| --- | --- |
| Identity, credentials, sessions, and direct contact data | Live records are deleted or revoked and cannot authenticate or authorize protected access. |
| Person-to-financial connection | No live or recoverable mapping, secret, key, log field, export, queue payload, cache, replica, or audit actor field can reasonably reconnect retained history to the person. |
| Retained financial history | Immutable transactions and snapshots remain only as Anonymous Financial History; the inventory must assess indirect and organizational reidentification, not only a database foreign key. |
| Per-subject erasure material | Future key material is non-exportable, cannot be recreated by the normal application path, and has independently verifiable destruction evidence. |
| Deletion marker and evidence | The future marker is durable beyond every recoverable backup, integrity-protected, availability-monitored, non-identifying, and never a usable person-to-subject map. |
| Backups and restore media | Every managed copy remains encrypted, expires within 90 days, has destruction evidence, and cannot serve restored data before marker replay and key-unavailability verification. |
| Audit and incident evidence | Audit retention remains 10 years without retaining credentials, raw request bodies, direct identifiers, or reversible mappings in lifecycle evidence. |

## Trust Boundaries and Authorities

| Boundary | Trust assumption that must be tested | Authority that must remain unavailable |
| --- | --- | --- |
| Browser and session | A future fresh-factor and anti-forgery ceremony can distinguish deletion and cancellation authority from a stolen or replayed session. | A normal session, CSRF bypass, opaque request ID, or stale browser response cannot silently create, cancel, or reverse a completed deletion. |
| Application lifecycle | The service can coordinate idempotent, monotonic state without possessing reversible erasure material. | Application code cannot export, recreate, or restore a destroyed per-subject key. |
| PostgreSQL and support access | Database access may expose retained data and must not be sufficient to reconnect it to a deleted person. | A DBA, support query, backup query, or surviving foreign key cannot reconstruct the identity link. |
| Key custody | Key-destruction proof is authentic and independently verifiable. | A single application or ordinary operator identity cannot unilaterally recover erased material. |
| Privacy control plane | Marker integrity, completeness, compatibility, and availability are independently verifiable. | A marker cannot contain direct identifiers, credentials, raw requests, predictable identity hashes, or a usable reversible map. |
| Backup and restore | Restore can begin isolated, without production traffic or credentials, and can prove replay before release. | A restored environment cannot accept traffic while marker or key evidence is missing, stale, conflicting, or unverifiable. |
| Audit and monitoring | Evidence can prove outcome and escalation without becoming a shadow identity store. | Logs, traces, alerts, and reports cannot preserve a reidentification path. |

These are proposed future controls, not present-day assurances. Every assumption remains untrusted
until a later design, implementation, and adversarial rehearsal provide evidence.

## Adversary and Failure Model


- unauthenticated callers probing deletion or cancellation state;
- an authenticated attacker with a stolen browser session, CSRF capability, replayed request, or
  stale client response;
- compromised application code, application credentials, or support tooling;
- privileged database, replica, cache, queue, export, or audit access;
- a malicious, mistaken, or unavailable key-custody operator or service;
- a malicious, mistaken, unavailable, or out-of-date marker/control-plane operator or service;
- an old encrypted backup, accidental disaster restore, or malicious/incorrect restore procedure;
- partial completion, timeout, duplicate delivery, race, or operator retry across independent
  database, key-custody, marker, and backup systems; and
- indirect reidentification through retained financial attributes combined with external or
  organizational knowledge.

This proposal does not assume a distributed transaction across PostgreSQL and a future key-custody
system. Safety must come from one-way, observable, fail-closed progress and independently verifiable
evidence, not an atomicity claim that the architecture cannot make.

## Threat Register


## Security Requirements Derived From the Threat Model

Before ADR-008 can be accepted, the future design must demonstrate all of these requirements:

1. Lifecycle state is durable, idempotent, serialized where required, monotonic, and never reports
   `completed` while any independent proof is absent or unverifiable.
2. A completion claim blocks normal authentication and protected writes. No retry, cancellation, or
   support shortcut can recreate key material or return the subject to a normal active state.
3. The marker control plane survives every recoverable backup and fails restore closed when its
   evidence is missing, stale, conflicting, unavailable, unauthentic, or incompatible.
4. The marker, audit evidence, and operational reports contain no direct identity fields,
   credentials, raw request payloads, predictable identity hashes, or usable mapping. A later design
   must prove how replay targets recovered material without violating this rule.
5. Future key custody separates application, key-custody, restore-release, and audit authorities.
   The normal request path cannot export, recreate, or recover destroyed material.
6. Restore remains network-isolated and non-serving until independent marker verification,
   idempotent replay, key-unavailability proof, evidence recording, and separately authorized
   release all succeed.
7. Every managed backup copy is encrypted, retained for no more than 90 days, reconciled against an
   inventory, and accompanied by destruction evidence or a fail-closed escalation.
8. The field-level inventory covers identity, links, credentials, sessions, settings, portfolio
   metadata, imports, projections, audit/log payloads, exports, queues, caches, replicas, indexes,
   and backup media. It must address indirect as well as direct reidentification.

## Required Evidence and Review Gates

later implementation proposal:


## Residual Risks

This proposal intentionally leaves the following risks unresolved rather than concealing them:


## Proposal Acceptance Criteria

- The threat model accurately distinguishes deployed gaps from future controls.
- It preserves Documents 42-43, immutable financial history, 10-year audit retention, and the
  90-day encrypted-backup limit without claiming current compliance.
- It treats marker correlation, privileged access, restore release, partial failure, and indirect
  reidentification as first-class security concerns.
- It requires evidence and explicit approval rather than filling provider, API, schema, or runbook
  gaps with assumptions.
- No runtime, OpenAPI, SQL, infrastructure, dependency, secret, provider, backup, or product-scope
  change appears in this stage.

## Recommended Next Step
