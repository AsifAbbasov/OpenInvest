# ADR-008: Privacy-Lifecycle Erasure and Restore Controls

| Field | Value |
| --- | --- |
| Document ID | ADR-008 |
| Version | 0.2.0 |
| Status | Proposed |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | Documents 42-43; ADR-005; ADR-006; Stage 2 ER model and migration strategy; Stage 3.17 and 3.18 privacy proposals |
| Last Review Date | 2026-08-09 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Security Review and explicit human decision before any privacy-lifecycle implementation |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/ADR/ADR-008-privacy-lifecycle-erasure-and-restore.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

## Context

Documents 42-43 require complete identity deletion, irreversible destruction of the person-to-financial
link, Anonymous Financial History, 10-year audit retention, and encrypted backup destruction within
90 days. The Stage 2 migration strategy requires revocable per-subject key material and deletion-ledger
replay before a restored environment accepts traffic.

The deployed schema does not yet meet that boundary. `identity.user_investment_links` is a plain
foreign-key relationship, and no key hierarchy, deletion ledger, backup-destruction evidence, or
restore serving gate exists. A database delete, a user state transition, or encrypted backup media by
itself is pseudonymization when an older restore, an operator, or residual data can recreate the
identity link.

The deletion flow also crosses a database and a future key-custody system. There is no distributed
transaction that can make database deletion and irreversible key destruction atomic. The design must
therefore fail closed under partial failure and never report completion before both independently
verifiable effects are complete.

## Proposed Decision

This ADR proposes the following provider-neutral control model. It is not accepted, implemented, or
authorization to select a KMS, Vault, backup provider, database role model, API route, or migration.

1. A future migration must replace the plain reversible identity-to-subject protection boundary with
   a per-subject erasure boundary. All material that can reconnect an identity to retained financial
   history must be protected by non-exportable, revocable per-subject key material. A raw surviving
   foreign-key mapping is insufficient.
2. Completion must create a durable, non-identifying deletion marker before any restored environment
   can serve traffic. The marker contains only random lifecycle/erasure handles, irreversible key
   destruction evidence, lifecycle timestamps, integrity/version metadata, and non-personal failure
   codes. It contains no email, user ID, subject ID, password, session token, raw request body, or
   usable identity map.
3. The marker/control plane must survive every recoverable data backup. It must have independent
   integrity protection and availability monitoring; losing, conflicting with, or failing to verify
   it blocks restore serving rather than weakening deletion.
4. Key custody must enforce separation of duties. The normal application path may request an
   authorized lifecycle transition but cannot export, recreate, or unilaterally restore destroyed
   subject key material. The exact roles, approval threshold, credential mechanism, and provider
   controls remain Security Review decisions.
5. Completion follows one-way, fail-closed progress: durable erasure intent, protected-write and
   authentication block, key-destruction request/proof, deletion of live identity/link data,
   anonymization of the retained financial subject, and verification. A request remains
   `completing`, never `completed`, while any step is missing or unverifiable.
6. Restore begins network-isolated and non-serving. It verifies the independent marker set, applies
   markers idempotently to remove/revoke recovered reidentification material, proves destroyed keys
   remain unavailable, records non-identifying evidence, and only then permits traffic.
7. Backups must stay encrypted, expire no later than 90 days, and produce destruction evidence for
   every managed copy. The final storage topology and proof source require provider-specific review.

The future OpenAPI, PostgreSQL schema, operational runbook, and implementation tests must use this
control model only after this ADR is accepted through the required review and human approval.

## Security Properties

| Property | Proposed control |
| --- | --- |
| Irreversible unlinking | Destroy the per-subject erasure key and every live reversible mapping; do not treat a UUID or removed user row as anonymous by itself. |
| No restore reidentification | Restore replays independently protected deletion markers before serving and fails closed on missing or unverifiable evidence. |
| Least privilege | Application, key-custody, restore, and audit authorities remain separate; no normal business API may recover erased material. |
| Partial-failure safety | An incomplete deletion denies normal auth and protected writes, remains observable, and cannot be reported as completed. |
| Non-identifying evidence | Audit and marker records retain outcome/timing/integrity evidence without credentials, tokens, personal fields, or usable maps. |
| Immutable history | Financial transactions and snapshots remain; retained content must separately pass the field-level anonymization inventory. |
| Backup bound | Every encrypted copy is destroyed by the approved 90-day maximum, with durable evidence and escalation on failure. |

## Alternatives Considered

### Delete identity rows and rely on cascades

Rejected. Cascades remove live rows but do not prevent an older backup, log, audit mapping, or operator
with retained material from reconnecting a person to financial history.

### Encrypt only whole backups

Rejected. Backup encryption without subject-scoped revocation leaves an authorized restore/key holder
able to recover identity mappings. It is not cryptographic erasure.

### Keep a reversible deletion map for operational convenience

Rejected. A reversible map makes the resulting data pseudonymized and conflicts with Documents 42-43.

### Let the application service destroy and recreate keys

Rejected. It collapses key custody into the normal request path and creates a unilateral
reidentification/recovery authority.

### Mark deletion complete after one subsystem succeeds

Rejected. Distributed partial failure can leave either live reidentification material or recoverable
key material. Completion requires evidence for every irreversible effect.

## Consequences

### Positive

- Establishes a reviewable no-reidentification and restore boundary before destructive code exists.
- Separates provider choice from durable security requirements.
- Gives future migration, API, operations, and test work a shared failure model.
- Preserves immutable financial history without mislabelling pseudonymization as anonymity.

### Costs and Constraints

- Requires a future key-custody service/control plane, backup evidence, restore rehearsal, and
  least-privilege operational roles; those choices may incur provider and operating cost.
- The current plain identity link cannot remain the sole protection of retained financial history.
- Completion may take longer than a database transaction and must be observable as an asynchronous,
  fail-closed lifecycle rather than an instant delete response.
- This decision does not settle deletion-request API authentication, concrete data types, database
  locking, provider selection, ledger retention duration, or legal interpretation.

## Compatibility and Rollback

This proposed ADR changes no deployed system. It can be withdrawn by reverting this documentation
commit before acceptance. After implementation, rollback cannot recreate destroyed key material or a
deleted identity link; future implementation planning must specify only forward recovery and a
non-serving restore procedure.

## Security and Privacy Impact

The proposal strengthens the intended privacy model but creates no current data collection, key,
backup, provider, or operational behavior. Security Review must validate threat assumptions, trust
boundaries, key custody, marker integrity/availability, destruction proof, restore isolation,
incident handling, and the data inventory before acceptance.

## Approval Outcome

Proposed only. It has no authority until strict review, Security Review, and explicit Principal
Architect acceptance. Stage 3.19 does not authorize an account-deletion API, schema migration,
runtime job, provider integration, or production claim.

## Revalidation candidate — 2026-10-16

**Status remains Proposed / acceptance BLOCKED.** The original decision clauses were restored from
pre-loss Git evidence; this is not acceptance or runtime implementation. The prospective
[Stage 3.24 assessment](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#repository-revalidation--2026-10-16-proposal-blocked_evidence)
contains current-baseline SR-01–SR-12 findings, alternatives, lifecycle/authorization fencing,
legacy-backup barriers, independent marker completeness/release protocol and future implementation tests.
The [Stage 3.21 delta](../stages/STAGE_03_21_PRIVACY_DATA_INVENTORY_PROPOSAL.md#current-baseline-inventory-delta--2026-10-16-repository-only)
adds registration, audit deduplication, replay/valuation, JWT, import-token and Redis surfaces.
A key-erasure proof cannot sanitize old plaintext copies or establish anonymity of financial patterns;
all current blockers require independent resolution under Documents 42–43 before approval.


## Proposed implementation-ready amendment — 2026-10-16 (NOT ACCEPTED)

**Decision authority:** This amendment specifies candidate contracts only. Documents 43 and 42,
accepted ADR-005/ADR-006, and the Stage 2 migration strategy remain authoritative. Its purpose is
to make [Stage 3.24's SR-01–SR-12 remediation design](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#sr-01sr-12-implementation-design-and-threat-to-evidence-trace-2026-10-16)
reviewable, not to approve deletion, providers, migrations, new public endpoints, or retrospective
claims of anonymity. Actual implementation is **absent**.

### Proposed trust boundaries and retained financial facts

```text
[Browser / user] --fresh-factor+CSRF--> [Go lifecycle coordinator, untrusted for release]
                                            |
                                            v
                         [PG identity: private transition/fence + auth epoch]
                                            |
                   +------------------------+--------------------------+
                   v                        v                          v
       [KMS custody operator]    [Independent marker publisher]   [Financial plane]
       no export/restore key     no marker-signing key             append-only facts
                   |                        |                          |
     [Independent verifier]    [Independent snapshot signer]       [derived views]
                   +------------------------+
                                            v
                        [Sealed recovery/replay environment]
                                            |
                                  [Separate release quorum]
                                            |
                                       serving gate
```

**Non-equivalence:** encrypting an ordinary FK, destroying a wrapping key, marking a subject
`anonymous`, and removing the email are distinct operations. Existing immutable transaction
rows (amount, date, revision, reversal, asset and portfolio graph) must remain financially exact,
yet their unusual patterns may identify a real person when joined against broker statements,
public transactions, logs or exported files. **If a documented attacker can still correlate
these facts, classification as Anonymous Financial History remains BLOCKED**; the approved
retention contract cannot be silently weakened by reporting pseudonymization as anonymity.
Any irreconcilable conflict requires a new higher-priority architectural decision.

**Suggested future erasure envelope:** a separately encrypted identity↔investment link reference
with a random subject-lifecycle handle and subject-scoped non-exportable erasure material.
Plaintext `identity.user_investment_links`, JWT `uid`/`investmentSubjectId`, audit actor and
target IDs, source-account labels, notes, traces, import fingerprints, and snapshots remain
**unprotected historical paths** until individually inventoried and addressed. Current source
evidence: migration 000001; migrations 000003–000004, 000006–000015; `auth/tokens.go`;
`postgres/auth_store.go`, `postgres/email_ownership_store.go`, `postgres/retention.go`;
`httpapi/import_security.go`. No runtime erasure fence exists.

### Proposed deletion authorization and state transition protocol (SR-05, SR-11)

Future state records are *candidate* migration objects, not present tables:

| Phase | Atomic / independently witnessed invariant | Allowed response |
| --- | --- | --- |
| `active` | One current auth generation; no deletion fence. | Standard operations. |
| `pending` | One stable intent for user+generation; fresh user proof; idempotency hash; approved grace deadline; **same SQL serialization authority** as login, refresh and financial subject ownership; revoke sessions and fence all protected reads/writes before acknowledging. | Generic receipt only; cancellation through a separately fresh proof. |
| `completing` | Grace expired and cancellation loses compare-and-set; append independently witnessed erasure intent and freeze immutable generation; reject new sessions, proof issuance, imports, jobs, ledger, exports and authorization. | No completion claim. Retry reconciliation only. |
| `custody_unavailable` / `proof_disputed` | Material destruction is unknown or verifier disagrees. | Non-serving, no cancellation/recreation; escalate. |
| `destroyed_proven` | Non-exportable subject material terminally destroyed; independent verifier authenticates all versions/replicas; evidence tied to exact intent. | No rollback; continue forward. |
| `unlinking` | Delete all live identity/correlation bridges and authorized sensitive projections; preserve immutable financial facts; check every copy and event/worker sink. | No completion claim. |
| `verifying` | Positive custody proof, primary unlink proof, field/anonymity proof, independently signed active marker snapshot and copy inventory all reconcile to same generation. | Still fail closed until all gates prove. |
| `completed` | All applicable proof clauses independently pass and are durably committed; restricted non-identifying receipt issued. | Completion status may be asserted **only to the authorized requester**. |
| `cancelled` | Only `pending` and before deadline/destruction fence; fresh factor, exact generation lock; produce **new** auth epoch; never revive previously revoked sessions. | Generic cancellation receipt. |

The 30-day grace mentioned in Stage 3.24 is a **design candidate, not an approved deadline**.
Require a separate Principal Architect/privacy decision before writing policy defaults or API contracts.
Grace cannot be restarted with a new idempotency key. Authentication on every protected route must
validate both JWT signature **and** current subject-state / authorization generation. Ordinary JWT
expiry (currently 15 minutes) and refresh-cookie revocation do **not** enforce immediate
account-deletion fencing. Race resolution: lock lifecycle+identity ownership first, then scoped
financial work; fail closed on stale epoch or state, including transaction commit-time validation.
Enforce server-side fencing for outbox/inbox consumers, materializers, scheduled jobs, import and
retry workers, not merely Fiber HTTP middleware or browser state. Document canonical lock
ordering, timeout, deadlock retry budget and transaction uncertainty handling in the future migration.
Before any irreversible custody action an independent durable marker **prepared** event and
recovery checkpoint must exist; **active** markers require both custody and live-unlink proof.

### Proposed restore and marker ordering (SR-03, SR-04, SR-06)

A sealed restore receives a **signed complete checkpoint**, not an untrusted list of tombstones:
`{schema_version, epoch, sequence_high, prior_root, canonical_merkle_root,
source_log_roots, min_readable_generation, issued_at, valid_until,
signer_id, signer_policy_version, signature}` (illustrative only). It must independently
verify monotonic sequence continuity, proof of *no omissions*, anti-rollback monotonic witness
outside the backup fault domain, signer-authority separation, quorum/split-view detection,
freshness and retention scope. A signature over a filtered subset alone does not prove
completeness. Where the publisher alone can determine the supposedly canonical high
watermark, the signer must refuse to sign and restore must stay sealed.

The replay service only obtains a **restricted restore-scoped binding** between independent
random marker handles and protected recovered erasure artifacts; it cannot look up a marker by
email, `user_id`, portfolio, financial subject, ordinary request ID or a deterministic hash.
The temporary binding and credentials are confined to the sealed environment and destroyed
before release, with independently verified evidence. A historic plaintext backup **cannot**
be made anonymous by decrypting it in an operator-accessible staging environment and subsequently
deleting a row: use isolated access, mandatory replay before serving, erase all transient plaintext,
and prove all managed copies are within the 90-day retention boundary. Missing binding/marker,
legacy unclassified copy, stale signer checkpoint, incomplete page, unsigned item, changed
generation, offline custody or anti-rollback witness => **RELEASE DENIED**.

No restored runtime may start user ingress, background jobs, CDN refresh, email/notifications,
external webhooks or production credentials until a *separate* release authority validates the
full predicate `complete_log && replay_proofs && custody_unrecoverable &&
identity_unlinked && no_reidentification_path && backup_inventory && independent_approval`.
Failing any conjunct means `sealed`, never a partial serving mode.

### Migration and rollback barrier (SR-12)

- **Expand:** proposed isolated lifecycle/fence tables, custody-handle references, proof registry,
  restricted restore-only binding store, append-only marker feed and audit-minimum v2. No DDL here.
  Database role privileges prohibit routine API deletion, marker signing, release and key recovery.
- **Populate:** bounded, restartable **synthetic-first** mapping for old plaintext
  `identity.user_investment_links`; enumerate every present token, audit join, imported free text,
  historical backup, WAL, AOF, export and support copy. Legacy plaintext must not be copied into
  attestation tags, logs or migration reports. Compare per-scope counts, missing bridge references,
  generation races and financial projections without exposing identities to reviewers.
- **Switch:** dual-read only inside restricted boundary with one unambiguous authority; fail closed
  if a pre-fence process can still create plaintext FK/audit/user↔subject correlation.
  Old app binaries, expired offline clients, workers, retries and restores must be fenced by
  database/security policy before dropping the legacy path. No indefinite dual-write plaintext.
- **Validate:** independent negative-linkage tests across all copies, denied stale JWT,
  cancellation/destruction races, old-backup restore, signer split-view and custody failure;
  reconcile immutable ledger count/Decimal string values, BusinessDate, revisions, reversals,
  snapshot rebuilds, and replay idempotency.
- **Contract:** only after approved reviewer proof and maximum recoverable copy expiry with
  verified provider evidence may legacy reversible material be retired. **Rollback is possible
  only before irreversible first destruction**, and only if it would not reintroduce a forbidden
  bridge. After that point recover **forward**, never re-wrap, resurrect or serve old plaintext.

### Explicit unresolved approval decisions

`PA-01` independent custody provider/semantics, non-exportability and proof verifier;
`PA-02` signing/root-of-trust and independent completeness/anti-rollback witness;
`PA-03` policy on financial-pattern identifiability, audit immutable-event transformation
and restricted evidence retention; `PA-04` fresh factor, grace deadline, cancellation and
safe response contract; `PA-05` copy inventory, deletion SLA and residency; `PA-06`
legal exceptions/retention/hold conflicts; `PA-07` dedicated production roles/approvers
and restricted restore-release runbook; `PA-08` potential contract update for irreversible
anonymization versus immutable uniquely identifying history. None is approved by this amendment.

**Review status: Proposed / BLOCKED.** Section-specific controls and acceptance tests are
tracked in the [SR matrix](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#sr-01sr-12-implementation-design-and-threat-to-evidence-trace-2026-10-16)
and [EC plan](../stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md#sr-to-ec-execution-trace-and-independent-test-catalogue-2026-10-16).
