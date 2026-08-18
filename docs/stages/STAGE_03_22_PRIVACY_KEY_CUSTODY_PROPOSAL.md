# Stage 3.22 - Privacy Key-Custody and Destruction-Proof Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-22-PRIVACY-KEY-CUSTODY-PROPOSAL |
| Version | 0.1.2 |
| Status | Complete / merged through PR #51 at `5f42d32db1e045c23fb99a5af8f136b7a49e3bc2` |
| Owner | Principal Architect |
| Supersedes | None; follows the merged Stage 3.21 privacy data-inventory proposal |
| Dependencies | Documents 42-43; ADR-005; ADR-006; proposed ADR-008; Stage 2 ER model and migration strategy; Stage 3.17-3.21 privacy proposals |
| Last Review Date | 2026-08-18 |
| Next Review Date | Historical proposal closed; successor Stage 3.23 deletion-marker control-plane proposal |

## Purpose

Stage 3.21 showed that the repository has no per-subject erasure-key hierarchy, custody boundary,
destruction evidence, or restore gate. This proposal specifies the minimum provider-neutral custody
and destruction-proof properties a later design must satisfy. It narrows one evidence gap from the
threat model; it does not claim that destroying a key alone makes existing financial data anonymous.

The proposal has four outcomes:

1. define the authorities that must stay separate from the normal application path;
2. define a one-way logical lifecycle for future per-subject erasure material;
3. define the minimum content and verification properties of non-identifying destruction proof; and
4. state the adversarial tests, provider evidence, and operational decisions that remain mandatory
   before any migration, API, provider, backup, or runtime work can be authorized.

This is not a provider selection, KMS/HSM/Vault configuration, cryptographic implementation,
database schema, account-deletion API, operational runbook, or Security Review verdict. It does not
accept ADR-008, establish a legal retention policy, or prove production backup or restore behavior.

## Authoritative Boundary

Documents 42-43 require complete identity deletion, irreversible destruction of the
person-to-financial link, Anonymous Financial History, ten-year audit retention, and destruction of
every encrypted backup within 90 days. ADR-008 is still proposed and non-normative. It requires a
future per-subject erasure boundary, independently verifiable deletion markers, least privilege, and
fail-closed restore.

The current repository does not have those controls. `identity.user_investment_links` is a plain
reversible database relationship, while audit actors, import values, free-form fields, request/trace
identifiers, file hashes, outbox payloads, and browser state can retain direct or indirect links.
The Stage 3.21 inventory therefore remains controlling evidence: no key-custody design may call the
retained records anonymous until all reasonable technical and organizational reidentification paths
are independently addressed.

No proposal can assume an atomic distributed transaction between PostgreSQL, a custody service, a
marker/control plane, and backup providers. Safety must instead come from durable, monotonic,
observable steps that fail closed when proof is missing, inconsistent, or unverifiable.

## Terms and Non-Goals

| Term | Meaning in this proposal |
| --- | --- |
| Per-subject erasure material | Future non-exportable cryptographic material whose irreversible destruction is required to make a protected identity-to-financial bridge unusable. It is not a current repository object. |
| Custody plane | A separately controlled future service or provider boundary that creates, protects, authorizes destruction of, and attests to erasure-material state. |
| Custody handle | A random, provider-independent lifecycle reference. It is not derived from email, user ID, subject ID, portfolio ID, filename, or other personal data. It is correlation material and is not safe to expose broadly. |
| Destruction proof | Integrity-protected evidence that a specific custody handle reached an irreversible destruction state under the approved custody policy. |
| Deletion marker | A separately durable, non-identifying control-plane record used to gate restore. It is not a replacement for custody proof and is not designed in this stage. |
| Completion | The later lifecycle outcome only after every required identity, linkage, key, marker, backup, restore, and evidence condition is verified. It is never synonymous with a provider accepting a destruction request. |

Out of scope: key algorithms, envelope-encryption format, provider APIs, HSM choice, key alias naming,
credential technology, exact quorum size, subject/key table design, marker schema, deletion UX,
grace-period mechanics, data migrations, backups, restore execution, and production access grants.

## Required Separation of Authorities

The following are logical roles, not existing repository identities. One person, workload identity,
credential, or service account must not hold a combination that defeats the stated separation.

| Role | Permitted future authority | Must not have authority |
| --- | --- | --- |
| Application data path | Request an already-authorized lifecycle transition; use only scoped, non-exportable data-protection operations required for active accounts. | Export erasure material; create a replacement for destroyed material; approve or execute irreversible destruction; release a restored environment. |
| Lifecycle authority | Validate the separately approved deletion state and submit an idempotent destruction intent. | Treat intent acceptance as destruction proof; access raw key material; override a missing or invalid proof. |
| Custody authorization | Apply independently authenticated policy/quorum checks to destruction requests. | Map a custody handle to a person through application data; silently authorize its own request; release restore traffic. |
| Custody executor | Perform provider-native irreversible destruction or enforce its terminal state and emit attestation. | Return raw material, recreate a destroyed version, or issue an unsigned/unauditable completion signal. |
| Marker authority | Persist and serve integrity-protected non-identifying marker evidence required for restore gating. | Store a usable person-to-subject map; substitute marker presence for custody verification; recover key material. |
| Restore authority | Operate an isolated, non-serving restore and verify marker/proof evidence. | Use production credentials, bypass failed verification, or release traffic unilaterally. |
| Restore-release authority | Independently authorize network/credential release only after recorded evidence verifies. | Modify or manufacture custody/marker evidence to satisfy its own release decision. |
| Audit/evidence authority | Verify signed evidence and retain the minimum non-identifying audit record. | Access raw key material or reconstruct a person-to-financial map for convenience. |

The exact human, provider, service-account, and break-glass design remains a Security Review and
operations decision. Emergency access cannot include a recovery path for already destroyed erasure
material. A break-glass procedure may escalate an incomplete deletion, but it cannot mark it complete
or restore normal access without the same independent evidence.

## Provider-Neutral Custody Requirements

A later provider evaluation and implementation design must prove all of the following:

1. Erasure material is generated inside the custody boundary and is non-exportable to the
   application, database, operators, logs, diagnostics, backups, or support tooling.
2. Every custody handle is generated randomly, has no identity-bearing alias, label, tag, request
   payload, or derivation, and is treated as restricted correlation material.
3. The provider distinguishes a reversible disable/schedule state from irreversible destruction.
   A completion claim requires proof of the latter, including all relevant versions, replicas,
   caches, and recovery windows.
4. A normal application credential cannot authorize destruction by itself and cannot recreate a
   destroyed key, point the old handle at new material, or obtain plaintext material through another
   API.
5. The custody plane exposes an integrity-verifiable destruction result with a stable verification
   key or independently auditable root of trust. A mutable dashboard status or a success log line is
   insufficient evidence.
6. Custody audit events are minimized and redacted: no email, user/subject/portfolio ID, session or
   token value, raw request, filename, source label, or financial payload may appear in key labels,
   tags, event data, or alert text.
7. A provider or control-plane outage, policy conflict, stale attestation key, unknown key version,
   duplicate request, or proof-verification failure leaves the lifecycle in a non-complete,
   fail-closed state and triggers non-identifying escalation.
8. The provider's deletion semantics, retention, multi-region replication, audit-log retention,
   administrative recovery, tenant isolation, credential recovery, destructive-operation approvals,
   attestation format, availability objective, cost, and exit path are recorded before selection.

The future mapping that locates erasure material is itself sensitive. It must either be protected by
the erasure boundary or be provably unable to reconnect retained data to a person after completion.
Keeping an ordinary database table from user or subject ID to a live custody handle is not a
sufficient design. This proposal deliberately does not choose the mapping mechanism.

## Logical Erasure-Material State Machine

The following state names are a future design constraint, not a schema or API. A later design may
refine the names but may not weaken the one-way evidence properties.

| State | Required condition | Allowed transition | Prohibited result |
| --- | --- | --- | --- |
| `absent` | No erasure material is provisioned for the future protected boundary. | `provisioned` through approved custody creation. | Assuming the current plain database link is protected. |
| `provisioned` | Material exists only inside custody; its restricted handle and policy version can be verified. | `destruction_requested`. | Exporting material or exposing a person-derived key alias. |
| `destruction_requested` | A durable idempotent intent references the restricted handle and the approved lifecycle evidence. Normal auth/protected writes are blocked by the broader lifecycle design. | `destruction_authorized`, or a recorded non-complete failure/escalation. | Reporting completion, cancelling through a stale request, or creating replacement material. |
| `destruction_authorized` | Independent policy/quorum evidence is valid for the same intent and handle. | `destroyed_proven`, or non-complete failure/escalation. | Treating authorization as provider destruction or permitting data-plane recovery. |
| `destroyed_proven` | A verifier independently validates terminal irreversible-destruction proof under the approved policy/version. | Only into the broader deletion verification/marker workflow. | Recreating material under the same handle, reactivating the subject, or serving restore traffic solely from this state. |
| `proof_disputed` / `proof_unavailable` | Proof is missing, stale, conflicting, unverifiable, or incompatible. | Escalation and evidence repair that does not recreate material. | `completed`, normal reactivation, or restore release. |

Creation and destruction requests must be idempotent only with respect to the exact serialized
intent, custody handle, policy version, and lifecycle generation. Reusing an idempotency key for a
different subject, handle, policy, or request payload must fail. A retry may observe the same terminal
proof; it must never silently issue replacement material or overwrite disputed evidence.

There is no rollback from `destroyed_proven`. A future system may recover forward by deleting
remaining live links, reconciling markers and backups, or escalating an incomplete workflow. It may
not restore deleted identity data, recreate equivalent erasure material, or claim that a newly
provisioned key repairs the prior security boundary.

## Minimum Destruction-Proof Contract

The future proof format may be provider-native only when a separate verifier can validate the
provider's authenticity, integrity, policy/version semantics, and terminal-destruction meaning. If a
portable wrapper is used, it must carry or bind the underlying provider evidence rather than merely
repeat a status string.

At minimum, a proof must bind all of these fields:

| Required proof element | Requirement |
| --- | --- |
| Proof format and schema version | Canonical, versioned, and reject-unknown by default. |
| Random lifecycle and custody handles | Generated independently of direct identifiers; access-restricted; never presented as public lookup keys. |
| Exact destruction-intent digest | Binds the proof to the approved serialized intent, policy version, and lifecycle generation without persisting the raw request or a usable map. |
| Custody policy and material version | Identifies the evaluated policy and every relevant material/version/replica scope without personal labels. |
| Terminal outcome and trusted timestamps | States irreversible destruction and the trusted observation time; scheduled deletion, disabled state, or accepted request is insufficient. |
| Provider or custody attestation | Independently verifiable digital signature or externally auditable append-only equivalent, including signer/key identity and algorithm/version. A shared-secret MAC is insufficient because a verifier that holds the secret can fabricate the same evidence. |
| Verifier record | Records verifier identity/class, verification time, validation result, and non-personal failure code. |
| Retention and access metadata | The minimal redacted audit/verification record defined below is retained for the mandatory ten-year audit period under restricted access. Any additional proof retention requires separate approval and must not quietly become a permanent correlation store. The 90-day limit applies to managed backup copies, not this audit-retention obligation. |

For the mandatory ten-year period, the minimal redacted audit/verification record must preserve the
proof schema version, restricted random lifecycle/custody handles or an equally verifiable binding,
exact destruction-intent digest, custody policy/material scope, terminal result/time, attestation
signer/key/algorithm version, and verifier result/time. It must include enough immutable signed or
externally auditable evidence for a later verifier to validate the recorded result; a mutable status
or an unsigned summary is insufficient. Access to this record is limited to the approved audit and
evidence authorities and is itself subject to the linkage analysis required by Stage 3.21.

The proof, marker, audit evidence, traces, metrics, alerts, and support exports must exclude raw key
material, emails, user/subject/portfolio IDs, credentials, session values, raw deletion requests,
file names, import data, trace values that can be joined to personal logs, and any reversible
person-to-financial map. A random handle does not become anonymous merely because it is random; the
future data inventory and access model must show that it cannot reasonably be joined back to a
person or retained financial history.

Proof validation must fail closed for missing fields, unknown versions/algorithms, stale or revoked
attestation keys, invalid signatures, mismatched intent/policy/material versions, conflicting
terminal states, untrusted clocks, unapproved retention, or unavailable evidence. The validation
result itself must not disclose protected lifecycle status to an unauthenticated caller.

## Restore and Backup Boundary

Custody proof is necessary but not sufficient for restore safety. A restored environment remains
network-isolated, non-serving, and without production credentials until the later marker design and
restore procedure independently establish all of the following:

1. the complete marker set is authentic, available, compatible, and replayed idempotently;
2. every recovered protected record is reconciled against the marker and proof set;
3. each claimed destroyed custody handle is independently verified as unavailable for recovery;
4. no restored job, migration, support utility, or application credential can recreate material,
   re-establish a link, or issue normal traffic before release; and
5. a separate restore-release authority records a non-identifying decision after all evidence is
   verified.

Missing custody proof, a marker conflict, unavailable custody plane, uncertain provider deletion
semantics, or an unverifiable historical attestation blocks release. The workflow may preserve the
isolated environment for investigation under restricted access, but it must not turn uncertainty
into availability by serving traffic.

Every managed backup, replica, PITR/WAL segment, provider export, and recovery medium still requires
the Stage 3.21 inventory, encryption, 90-day destruction evidence, and reconciliation design. Key
destruction does not excuse a missing backup inventory or remove residual reidentification risk in
unencrypted or independently copied material.

## Required Evidence Before Security Review or Implementation

Before a Security Review can recommend ADR-008 acceptance or a later implementation proposal can
begin, the accountable owners must provide at least:

1. a field-level mapping design showing how erasure material protects every live and recoverable
   person-to-financial bridge without retaining a raw fallback mapping;
2. a provider-neutral capability matrix covering non-exportability, versioning, irreversible
   deletion, recovery/admin paths, attestation, deletion semantics, availability, regional copies,
   audit privacy, retention, cost, and exit;
3. a provider-specific threat, access, cost, residency, support, incident, and procurement record
   after a provider is proposed, with no provider accepted merely by this document;
4. an exact custody policy and separation-of-duties design, including identity/credential lifecycle,
   approval quorum, break-glass limits, administrative recovery limits, and access reviews;
5. a canonical destruction-proof schema and verifier design with redaction review, replay/idempotency
   behavior, key/attestation rotation, compatibility rules, evidence retention, and failure codes;
6. a non-identifying marker/control-plane design that proves restore targeting without forming a
   usable person-to-subject map;
7. a migration and data-state proposal for identity, links, credentials, sessions, audit actors,
   free-form fields, import artifacts, derived data, queues, logs, and backup copies;
8. an operations runbook for authorization, destruction, proof collection, outage/escalation,
   audit access, isolated restore, marker replay, backup expiry, evidence retention, and release;
9. adversarial tests for unauthorized export/destruction/recreation, duplicate intent, race,
   partial failure, false proof, stale verifier, provider outage, stale restore, backup expiry, and
   early traffic release; and
10. an isolated restore rehearsal proving a completed identity cannot authenticate, write protected
    data, or be reconnected to retained financial history.

No missing item may be replaced with a database cascade, a soft-delete state, a successful HTTP
response, a mutable provider console, an unverified log, or a verbal operational assurance.

## Proposal Acceptance Criteria

- It separates normal application, custody authorization/execution, marker, restore, release, and
  evidence authorities without claiming current implementation.
- It distinguishes reversible disable/schedule actions from irreversibly proven destruction.
- It requires a restricted, non-identifying, integrity-verifiable proof tied to the exact approved
  intent and custody-policy/material scope.
- It makes failure, conflict, stale evidence, unavailable custody, and uncertain semantics block
  completion and restore release.
- It preserves the Stage 3.21 conclusion that no key design alone proves Anonymous Financial History.
- It selects no provider and adds no runtime, OpenAPI, migration, database, backup, credential,
  operational, dependency, or product-scope change.

## Scope Boundary

This stage adds documentation only. It does not modify Go, Next.js, Python, OpenAPI, PostgreSQL,
Redis, migrations, test fixtures, CI, infrastructure, providers, credentials, backup settings, or
operational access. It does not authorize deletion, anonymization, encryption, key provisioning,
key destruction, a provider trial, a production discovery, a restore, or an implementation stage.

This proposal was strictly reviewed and squash-merged through PR #51 at
`5f42d32db1e045c23fb99a5af8f136b7a49e3bc2`. Its successor, Stage 3.23, designs only the
non-identifying deletion-marker control plane and restore gate. No field-level migration or runtime
implementation is authorized by this historical closure.
