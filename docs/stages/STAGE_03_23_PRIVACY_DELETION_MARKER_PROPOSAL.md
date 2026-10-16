# Stage 3.23 - Privacy Deletion-Marker Control-Plane Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-23-PRIVACY-DELETION-MARKER-PROPOSAL |
| Version | 0.3.0 |
| Status | Complete / merged through PR #52 at `f7f23bce33038f259c976db6375079c68209a7aa` |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.22 key-custody proposal |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.19-3.22 privacy proposals; Stage 3.21 inventory |
| Last Review Date | 2026-08-18 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Historical proposal closed; successor Stage 3.24 Security Review readiness dossier |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/stages/STAGE_03_23_PRIVACY_DELETION_MARKER_PROPOSAL.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

## Purpose

Stage 3.22 requires a durable, non-identifying deletion marker to prevent a restored copy from
silently reviving a destroyed identity-to-financial link. This proposal defines the minimum
provider-neutral control-plane properties for that marker. It is deliberately narrower than an
implementation: it defines evidence, lifecycle, restore gating, and unresolved design proof, not
an API, database schema, provider configuration, or runbook.

The current repository has no deletion-marker control plane. The reversible
`identity.user_investment_links` relationship, existing audit fields, request/trace identifiers,
hashes, and external persistence gaps identified by Stage 3.21 cannot be treated as a marker or as
proof of anonymization. A later implementation must establish every field-level disposition and
external evidence obligation before it may claim completion.

This proposal does not accept ADR-008, perform Security Review, select a provider, establish a
legal retention period, change OpenAPI or database contracts, alter backups, or authorize runtime
work. Destruction of erasure material remains governed by Stage 3.22; a marker never substitutes
for independently verifiable custody proof.

## Security Boundary

A marker must survive every managed recoverable copy of application data, but it must not become a
new identity map. It is restricted correlation material, not anonymous public data. The future
design must analyze linkage attacks across marker values, lifecycle timing, audit records, restore
jobs, snapshots, and external logs. It must reject a design that makes a marker guessable,
queryable by ordinary application identities, or reversible to a person, account, subject,
portfolio, request, import, device, or raw foreign key.

The marker plane is separate from the application data path, key custody, and restore-release
authority. No credential, service, or workflow may both decide deletion is complete and override
the evidence needed to release a restored environment. A missing, stale, unverifiable, conflicting,
or unavailable marker state fails closed.

## Minimum Record and Prohibited Content

| Class | Minimum property |
| --- | --- |
| Marker identifier | Random, unguessable, non-semantic value; never derived from identity data. |
| Replay binding | Separate random restore-only reference, scoped to marker generation and restore evidence. |
| Lifecycle state | Prepared, active, disputed/unavailable, retirement-eligible, or evidence-only. |
| Integrity metadata | Approved algorithm/signer reference, signed canonical snapshot watermark, and verification result. |
| Proof commitments | References or commitments binding custody, unlinking, and backup evidence without copying personal content. |
| Retention/access metadata | Policy version, timestamps, authorized purpose, and safe failure code. |

It must never contain an email, name, user/subject/portfolio identifier, raw foreign key,
credential, raw request or import payload, raw key, raw key handle, predictable identity hash,
trace identifier that joins to personal logs, or public lookup token. Any proposed commitment,
hash, identifier, or timing field needs a documented correlation analysis; "hashed" is not an
anonymization claim.

## Required Separation of Authorities

| Role | May do | Must not do |
| --- | --- | --- |
| Lifecycle coordinator | Submit idempotent transition intent after separately approved deletion conditions. | Publish completion or release a restore. |
| Marker publisher | Create a prepared record and advance only after required proof verifies. | Recover identities, key material, bypass custody evidence, or hold the snapshot-signing/attestation credential. |
| Integrity authority | Independently sign or validate canonical marker snapshots. | Publish marker lifecycle state, share administration/credentials with a marker publisher, or alter lifecycle meaning to release its own restore. |
| Custody verifier | Verify Stage 3.22 destruction proof. | Treat marker presence as proof of destruction. |
| Restore replayer | Apply active marker only in an isolated restore. | Use production ingress, credentials, or release authority. |
| Restore-release authority | Authorize serving only after independently recorded checks pass. | Manufacture or amend marker/custody evidence. |
| Audit/evidence authority | Retain minimally redacted completion evidence. | Retain a usable identity map or a recoverable backup beyond policy. |

## Monotonic Lifecycle

The control plane must be append-only and idempotent, with a stable transition intent binding.
Reusing an idempotency key for a different serialized intent is a rejection, not a retry. Every
transition needs an attributable, integrity-protected result and a safe failure outcome.

1. **Absent:** no lifecycle request is accepted as completed.
2. **Prepared:** an intent is durable, but destructive replay is prohibited. A prepared marker in a
   restore is a non-serving escalation condition.
3. **Active:** allowed only after independent custody destruction proof and primary
   linkage-removal evidence verify against the same lifecycle generation.
4. **Replay-attested:** for a specific isolated restore, the active marker was applied idempotently
   and bound to the canonical snapshot and restore operation; this does not authorize serving.
5. **Retirement-eligible:** reached only after evidence covers every managed recoverable copy under
   the maximum 90-day backup policy. Unknown external copies, logs, exports, replicas, or providers
   block completion; their absence must never be assumed.
6. **Evidence-only:** operational marker access is retired while minimum redacted audit and
   verification evidence remains retained for ten years. This does not extend backup retention or
   preserve a recoverable identity map.

`disputed`, `unavailable`, signature failure, snapshot gap, incompatible version, duplicate
generation, and out-of-order evidence are fail-closed conditions until independently authorized
investigation produces a new auditable decision. They must not be auto-coerced to active or
evidence-only.

The marker publisher and integrity authority must use separately administered credentials and
independent authorization paths. A publisher must not be able to issue, amend, or select the
trusted completeness watermark for its own state writes. If an independent signer is unavailable,
the only permitted alternatives are an approved threshold scheme or externally auditable attestation
in which the publisher cannot unilaterally satisfy the release verifier.

## Snapshot Integrity and Availability

Before an environment can be released after restore, it must obtain a complete canonical marker
snapshot with a signed or externally auditable integrity proof. A shared-secret MAC controlled only
by the restoring path is insufficient as the sole authenticity mechanism. The verifier must prove
schema compatibility, explicit completeness/watermark semantics, bounded freshness, signature or
external-attestation validity, and no truncation, rollback, split-view, or omitted state. It must
account for every marker state, including prepared and unavailable ones.

A restore-release verifier must accept a completeness watermark only when the signer or attester is
independent of every credential that published the included marker lifecycle state. It must reject
a proof that a publisher can create, amend, or select unilaterally, even if its signature is
cryptographically valid.

The normal application path cannot use marker lookup as a user-facing authorization service. When
the plane cannot establish an acceptable snapshot, recovery remains isolated and non-serving. The
future operations design must specify quorum, pagination, replication, monitoring, disaster
recovery, and reconciliation without weakening this gate.

## Restore Gate

Restores occur in a sealed, non-serving environment with no production ingress, credentials,
background jobs, webhooks, or release route. The procedure must:

1. validate the canonical marker snapshot before data is made serviceable;
2. resolve every marker state and stop on unknown, prepared, disputed, unavailable, stale, or
   unverified state;
3. replay every active marker idempotently using the replay binding tied to that snapshot and restore
   operation, without resolving an identity;
4. verify corresponding destroyed erasure material remains unavailable under Stage 3.22;
5. record redacted evidence; and
6. require a separate release decision before networks, credentials, or jobs are enabled.

Recovery success means only that a non-serving copy is safe enough for independent release review.
It is not proof every external record was deleted and cannot replace restoration testing or provider
evidence.

## Retention and Required Evidence

Managed recoverable backups are limited to 90 days by Documents 42-43. Active marker evidence must
survive until every managed recoverable copy is proven replayed or destroyed. Afterwards, the plane
may keep only minimum redacted, non-reversible completion and verification evidence for the ten-year
audit obligation. Ten-year audit evidence does not authorize ten-year backups, a broad marker store,
raw linkage data, or reidentification capability.

Retention, legal hold, incident investigation, and external-provider evidence require separate
human and legal/security decisions. None may silently prevent destruction proof, make an old backup
recoverable, or convert a failed deletion into a completed lifecycle state.

A later migration or implementation proposal must provide field-level schema and redaction
disposition; non-correlation analysis; serialized-intent and replay-binding design; role/capability
matrix; signed snapshot/completeness protocol; provider-specific durability and deletion evidence;
custody-to-marker transition protocol; migration/backfill and rollback limits; backup, replica,
export, log, and CI evidence; retention/hold policy; monitoring and incident response; and a
rehearsed isolated restore with stale, missing, conflicting, forged, reordered, and unavailable
marker states injected.

This proposal was strictly reviewed and squash-merged through PR #52 at
`f7f23bce33038f259c976db6375079c68209a7aa`. Its successor, Stage 3.24, prepares the mandatory
Security Review evidence and decision boundary. No field-level migration or runtime implementation
is authorized by this historical closure.

## Target marker and restore protocol v1

Candidate control plane only. Preserve the original separation rules and prohibited fields above.
The log uses authenticated append-only ordering and independent checkpoints; [RFC 9162](https://www.rfc-editor.org/rfc/rfc9162.html)
sections 2.1.3–2.1.4 inform inclusion/consistency checks. This is an adaptation of an Experimental
certificate-transparency protocol, not adoption of a public log, a deletion standard or proof that
all real-world deletions were submitted. Official source inspected 2026-10-16.

| Record | Required binding / validation |
| --- | --- |
| Intent admission receipt | Random lifecycle ID, generation, policy version, domain-separated digest of minimized canonical intent, authoritative sequence/epoch and independent witness signature. No raw confirmation factor or identity map. |
| Marker transition | Same lifecycle/generation/digest, random restore-only binding, phase, proof commitments, previous-transition commitment and schema version. Prepared cannot authorize deletion completion or serving. |
| Checkpoint | Log identity/epoch/size/root, previous trusted checkpoint, policy/schema/signer versions, trusted observation time and independent witness endorsements. Completeness includes every admitted intent and its resolved or unresolved state. |
| Restore release certificate | Restore instance/configuration digest, checkpoint/cut, replay result, custody/copy/evidence commitments, independent release signer and finite validity. A copied certificate for another instance/cut is invalid. |

**Admission v2 — ARCH-277-01:** the earlier database-first candidate is superseded. Prerequisite
**G1** is an independently administered, linearizable intent service whose acknowledged records survive
the entire assessed PostgreSQL/backup failure domain. Its recovery point for acknowledged admissions is
zero loss; a merely asynchronous outbox or periodically exported log is insufficient. Quorum, replica,
anti-rollback trust and the privacy-safe restore binding remain A3 decisions, not available controls.
If G1 or the ability to reconcile every subject is unproved, admission/restore-release stays BLOCKED.

1. Freshly authorize and assign one immutable intent ID/digest/generation. Before any local pending
   commit, independently persist **PREPARE** with a restricted restorable scope binding. A signed
   durable receipt is required. Do not acknowledge pending from a local transaction alone.
2. Under the database-enforced compatibility/lifecycle fence, commit `admission_pending`, revoke
   normal sessions and deny normal work; bind the PREPARE receipt and transactional checkpoint. An
   aborted transaction does not remove PREPARE; a committed pending state is never “rolled back”.
3. Independently **ACTIVATE_PENDING** the same intent after verifying the database checkpoint.
   The service fixes `accepted_at` and `deadline = accepted_at + 30 * 86400 seconds` using trusted UTC.
   This is a proposed exact-duration interpretation of the canonical 30-day grace, requiring owner
   confirmation before implementation; it does not invent a historical deadline. PREPARE has no
   erasure deadline. Activation retries return the original times, never a new window.
4. Persist the activation receipt/deadline locally, then return an externally acknowledged pending
   response. A lost/delayed response does not move the durable acceptance time; restricted fresh
   status/retry reports the existing state/deadline, not another accepted request. No irreversible
   claim before that deadline; unknown/regressing/untrusted time freezes the claim.
5. Cancellation and deadline claim serialize at the authoritative service as well as the database:
   cancellation requires `now < deadline` and a fresh purpose-bound factor; claim requires
   `now >= deadline`. Only a witnessed cancellation may restore active state. Once claimed, no
   cancellation or key recreation is possible. No database transaction spans the service call.

| Observed condition | Required recovery / acknowledgement / quarantine |
| --- | --- |
| DB committed, witness absent | Forbidden in v2. Treat legacy/inconsistent state as unresolved pending; no active fallback or external success. Recover only from independently authenticated intent/WAL evidence; absent evidence quarantines the entire affected restore scope. |
| PREPARE/activation admitted, DB checkpoint absent | Recover the minimal denied state and original generation/deadline from authenticated service evidence and approved binding. Resume forward; no identity recreation. Unknown binding or unresolved PREPARE prevents release, not silent omission. |
| Admission outcome unknown / acknowledgement lost | Query/retry the same immutable intent until a durable terminal admission result is verified. Timeout/partition/empty local cache is neither rejection nor success. No new generation or deadline. |
| Primary loss before witness acknowledgement | Local pending commit is not permitted without PREPARE receipt. If a v1 DB-first operation might have existed, missing primary/WAL means absence cannot be proved; quarantine. No claim to reconstruct missing WAL. |
| Backup lacks complete WAL | Independent admitted-intent coverage must account for PREPARE, pending, cancellation and irreversible claims, not only completed markers. Reapply denied states before access; unknown historical coverage blocks global release. |
| Delayed/duplicate request, cancellation or receipt | Exact digest/generation/epoch must match. Authoritative aborted/cancelled outcomes stay durable; delayed PREPARE cannot reactivate a cancelled generation. Retry never compensates irreversible destruction. |
| Independent service unavailable/lost/rolled back | Deny admission acknowledgement and restore release. A live denied scope remains denied. Loss outside the approved G1 failure model is an incident, never inferred safe from an empty marker snapshot. |

Release therefore needs independently authenticated **epoch coverage** from the first possible local
privacy admission through the release cut, including proof that a legacy DB-first era had no uncovered
operations. With an older backup, missing WAL and no independent record, no algorithm can distinguish
“no request” from “lost pending request”. There is no negative-evidence shortcut: keep the whole
uncertain environment non-serving until independent evidence resolves it. A proposal/migration flag
alone cannot prove zero historical operations. No current production deletion endpoint exists, but a
future rollout must independently establish its own zero-operation boundary.

PREPARE/unknown outcomes are visible denied obligations, never destructive replay instructions.
An uncommitted attempt may be explicitly aborted only by an authorized, witnessed terminal transition
that fences delayed messages; database rollback or lease timeout alone cannot erase it. Liveness is
conditional on restored service quorum, trusted time, binding availability and authorized reconciliation;
permanent partitions may prevent progress indefinitely without weakening safety.

**Completeness:** verify every receipt through cut N against the G1 independent register including PREPARE/unknown outcomes,
canonical sequence/entry count/root, all pages, inclusion proofs and consistency with a previously
pinned checkpoint. Reject duplicate sequences/generations, missing pages, mismatched intent digests,
unknown schemas/epochs/signers, rollback and conflicting roots at the same size. Compare checkpoints
across separately administered witnesses; one writer and its own signer are not independent. The
witness register must itself have monitored replicated durability, retention and disaster recovery.
Signatures prove authenticity of a view, not that unsubmitted or lost intents do not exist; unresolved
journal/data reconciliation prevents release. No numeric quorum or provider is self-approved here.

**Release cut:** require G1 epoch coverage and the G2 database/credential cutover from Stage 3.24;
block new admission to the recovered environment and establish a coordinated epoch
barrier with the authoritative intent service; drain all pre-barrier admissions, obtain checkpoint N,
reconcile recovery data and replay every required marker, then obtain independent release authority.
Intent admission after N belongs to the next epoch and cannot use the recovered instance until its
continuous lifecycle gate is synchronized. A standalone freshness timeout is insufficient. If a shared
admission barrier cannot be established, keep recovery non-serving; do not globally pause an active
production environment without separately approved operations authority.

**Replay:** use only reviewed restore binding/protected scope membership (A3), not user-facing lookup.
Apply idempotently by (restore instance, lifecycle generation, checkpoint); verify each affected record
and independent destroyed-key unavailability. Missing primary record can be an idempotent outcome
only after scope completeness is established; it does not prove the whole subject was erased.
Resolve prepared/disputed states from durable intent and independent proof; never blindly delete from
an unverified/prepared marker. Partial replay resumes from verified checkpoints, with release denied.

**Sealed runbook:** restore operator creates an isolated instance with no public ingress, production
signing keys, application credentials, mail/provider credentials, scheduled work or outbound exports.
Use separate short-lived restore-verifier credentials; allow only assessed custody/journal evidence
paths. Verify database image/schema and replay tools; load trusted checkpoint roots independently;
replay/verify; reconcile copies and scopes; obtain a separate release certificate; enable ingress,
credentials and jobs only for that exact instance/configuration. Restart/configuration change or
expired/invalid certificate returns the candidate to non-serving. Platform enforcement is unverified.

Backups/copy register: random copy ID, origin/lineage, creation time, expiry no later than creation +
90 days, encryption/material class, replica/export/recovery locations, deletion operation, independent
provider evidence and reconciliation result. Copying/rewrapping must not reset retention of the same
recoverable data. Unknown locations or failed expiry trigger incident escalation and block final
completion. Markers remain operational until all managed copies are replayed or destroyed; then keep
only non-identifying ten-year evidence. Physical replicas and exports with other approved purposes
need explicit disposition; they are not silently exempt recovery copies.

Legacy plaintext backups, logical exports, WAL, diagnostic copies and lookup indexes are a mandatory
separate barrier. A new subject key cannot sanitize them. Final completion waits for independently
verified destruction/expiry and linkage resolution; “crypto-erased now, backups later” is not completed.
Legal holds cannot silently extend the 90-day contract or reintroduce keys: unresolved conflict remains
blocked for human decision. Actual provider inventory/destruction and restore-isolation evidence are
not supplied by this candidate.
