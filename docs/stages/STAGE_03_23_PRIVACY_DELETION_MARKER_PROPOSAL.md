# Stage 3.23 - Privacy Deletion-Marker Control-Plane Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-23-PRIVACY-DELETION-MARKER-PROPOSAL |
| Version | 0.1.1 |
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


## Signed checkpoint, completeness and sealed replay protocol — 2026-10-16 (proposal)

The existing marker security properties above are unchanged. The following is an
**illustrative verifier interface**, not a provider, schema, signed production snapshot
or implementation authorization.

### Two independent journal witnesses and a concrete proof boundary

The **lifecycle DB fence** durably increments an internal monotonically increasing
`lifecycle_generation` and commits a witnessed intent commitment before custody
destruction can start. A separate publisher appends a marker to a restricted independent
journal. The signer's completeness witness observes **all intents** through a separately
authenticated change feed/commit acknowledgement and tracks holes: simply signing
the publisher's selected page of rows cannot detect an omitted deletion. The witness
must support a durable prefix commitment that survives application and database restore
and cannot be rolled back by the restoring administrator.

Candidate canonical checkpoint fields:
```text
checkpoint_version, log_epoch, first_sequence, last_sequence,
prior_checkpoint_digest, canonical_leaf_root, intent_witness_root,
leaf_count, outstanding_prepared_count, signer_policy_version,
issued_at, valid_until, anti_rollback_reference, signature
```

Canonical marker leaf commitment:
```text
marker_schema_version, random_marker_id, generation, randomized_replay_binding,
non_personal_state, intent_digest, custody_proof_digest,
primary_unlink_proof_digest, state_timestamp, previous_state_digest
```

**Sensitive design warning:** `randomized_replay_binding` is a restricted secret-adjacent
handle; the public checkpoint should contain only its reviewed commitment. A restore-only
mapping to recovered encrypted artifacts must reside in sealed privileged storage;
including any raw `user_id`, `investment_subject_id`, source-account hash, request/trace
ID or an unsalted predictable digest would recreate precisely the forbidden bridge.

**Verifier decision algorithm** (candidate, not executed):

1. Boot restored environment with production ingress, DNS egress, workers, job queues,
   support tools, KMS destructive privileges and normal credentials **disabled**.
2. Pin witness/signer trust roots out of the recovered database fault domain; check
   key/version rotation continuity and compare at least two independently administered
   witness views to detect split view.
3. Fetch **every** bounded page from the canonical first sequence to the latest trusted
   watermark, proving the `leaf_count`, consecutive sequence, previous-state digest,
   complete Merkle root, pending/disputed leaves and absence of omission. Explicitly
   prove the witness captured all committed lifecycle fences in the selected scope.
4. If a leaf is `prepared`, `disputed`, unknown, stale, missing or incompatible, quarantine
   the restored environment; do not fall back to older or smaller valid snapshots.
5. Use ephemeral sealed binding to reconcile **each** recovered sensitive artifact with
   the corresponding active marker, independently verify its terminal custody proof,
   and eliminate the resurrected identity/link and all authorized sensitive projections.
   Repeating the same marker must be idempotent; applying another marker's binding must
   fail without destroying an unrelated subject.
6. Scan/reconcile the protected SQL, WAL/PITR reference, retained audit joins, Redis
   copies, pending jobs and replay/import state; independently record precise failures.
   Destroy all transient plaintext and restore-only binding material.
7. A distinct release quorum validates **all** proof results and the 90-day managed-copy
   inventory, signs a one-use environment-specific permit, and only then allows the
   ingress/credential/job activation sequence. No “emergency bypass” may override
   an unverifiable deletion marker.

### Failure and adversarial acceptance matrix

| Fault injection / negative control | Mandatory behavior |
| --- | --- |
| Signed but old complete snapshot | Reject by external monotonic watermark/freshness. |
| Publisher omits one prepared/deleted intent | Signer cannot attest complete prefix; block release. |
| Two valid but conflicting signed roots | Independent witnesses detect split-view; quarantine. |
| Valid proof with wrong intent/generation/version | Reject marker activation and replay. |
| Valid marker but missing restricted restore binding | Remain sealed; never guess identity from UUID. |
| Same marker replayed twice | Same final result; no extra financial event/reversal. |
| Marker applied to another subject | Reject without foreign mutation. |
| Missing custody attestation, weak purge semantics or rewrapped key | Reject release regardless of signed marker. |
| Compromised app DB and signer still using same admin root | Trust separation failed; block decision. |
| Legacy plaintext backup contains no compatible reconciliation metadata | Quarantine old copy; do not serve or silently discard financial history. |

**Independent observation requirements:** authenticated checkpoint source, witness
anti-rollback result, page count and hole search, Merkle verification, custody proof check,
signed release decision, no hidden correlation path, and unchanged immutable ledger
quantities/dates/revisions. Local mock verifier tests may establish parser logic but
do **not** prove deployed witness independence or historical backup safety.

Map this design to EC-04/EC-06/EC-10 and TV-03/TV-04 from
[Stage 3.25](STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md#sr-to-ec-execution-trace-and-independent-test-catalogue-2026-10-16).
**Current status: PARTIALLY_DESIGNED / implementation absent; restore-release BLOCKED.**
