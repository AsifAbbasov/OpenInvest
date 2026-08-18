# Stage 3.23 - Privacy Deletion-Marker Control-Plane Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-23-PRIVACY-DELETION-MARKER-PROPOSAL |
| Version | 0.1.1 |
| Status | Complete / merged through PR #52 at `f7f23bce33038f259c976db6375079c68209a7aa` |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.22 key-custody proposal |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.19-3.22 privacy proposals; Stage 3.21 inventory |
| Last Review Date | 2026-08-18 |
| Next Review Date | Historical proposal closed; successor Stage 3.24 Security Review readiness dossier |

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
