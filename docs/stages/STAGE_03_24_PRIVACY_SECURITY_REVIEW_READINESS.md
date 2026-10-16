# Stage 3.24 - Privacy Security Review Readiness Dossier

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-24-PRIVACY-SECURITY-REVIEW-READINESS |
| Version | 0.2.0 |
| Status | Complete / merged through PR #53 at `544ad8cc7371caf93913ea7716f3feb68be0ea44` |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.23 deletion-marker control-plane proposal |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.23 privacy proposals; Stage 3.21 inventory |
| Last Review Date | 2026-08-18 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Historical dossier closed; successor Stage 3.25 evidence-collection plan |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

## Purpose

Documents 42-43 and proposed ADR-008 require an explicit Security Review and Principal Architect
acceptance before any privacy-lifecycle implementation may start. Stages 3.20 through 3.23 provide
design evidence, but none is a deployed control, production discovery, provider attestation,
operational proof, Security Review verdict, or ADR acceptance.

This dossier makes that boundary testable. It identifies the evidence that exists only as a
repository-derived proposal, the evidence that must be independently produced, the questions a
formal Security Review must resolve, and the conditions under which it must remain blocked. It does
not conduct Security Review, select a provider, approve a migration, or create an implementation
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
6. ADR-008 remains proposed and non-normative until strict review, formal Security Review, and
   explicit Principal Architect acceptance.

No field removal, cascade, soft-delete state, hash, UUID, encrypted media, provider console, log
line, successful request, or verbal assurance is sufficient evidence by itself.

## Evidence Status Register

| Area | Current repository-supported evidence | What it does not prove | Security Review entry evidence still required |
| --- | --- | --- | --- |
| Threat boundary | Stage 3.20 maps assets, adversaries, failure paths, and residual risks. | Production threat coverage or risk acceptance. | Named threat-model owner, reviewed threat assumptions, residual-risk disposition, and attacker/failure test plan. |
| Deletion authority | Stage 3.20 requires a future fresh-factor and anti-forgery lifecycle boundary. | Current session handling, a revoked session, an opaque identifier, or a successful request prove deletion/cancellation authority or status privacy. | Fresh re-auth/factor design, anti-forgery and rate-limit controls, non-oracular response model, cancellation authority rules, abuse tests, and audit/redaction evidence. |
| Data disposition | Stage 3.21 inventories observed schema/code/browser surfaces and marks external gaps. | Production data, backups, replicas, exports, logs, vendor systems, or organizational access. | Field-level disposition for every real store/copy, sampled discovery evidence, lineage, access path, and linkage analysis. |
| Migration and operations | Stages 3.17-3.23 describe future lifecycle, custody, marker, restore, and evidence properties. | A migration data-state design, constrained forward recovery, production runbook, or least-privilege operations control. | Non-authorizing migration/data-state design; locking/constraint/retention/non-recreation analysis; least-privilege deletion, proof, backup, restore, release, escalation, and evidence-redaction runbook. |
| Key custody | Stage 3.22 defines provider-neutral custody, destruction proof, and separation requirements. | A provider's deletion semantics, durability, access controls, or attestation. | Capability matrix; provider-specific threat/access/residency/cost/exit assessment; custody policy; signer/key rotation; independently verifiable attestation evidence. |
| Marker control plane | Stage 3.23 defines non-identifying marker content, lifecycle, independent integrity, and restore gate. | A durable implementation, complete marker snapshot, availability target, or correlation safety. | Field schema/redaction review; correlation analysis; signed snapshot protocol; pagination/completeness proof; quorum/replication/incident design; abuse tests. |
| Restore and backups | The proposals specify isolated restore and 90-day managed-copy rules. | Actual backup inventory, deletion, restore isolation, credential isolation, or provider behavior. | Backup/PITR/replica/export inventory, retention and destruction evidence, reconciliation, sealed restore runbook, and rehearsal results. |
| Governance | Strict documentation reviews and PR evidence exist for Stages 3.20-3.23. | Security approval, legal approval, production compliance, or deployment authorization. | Dated Security Review record with named accountable reviewers, scope, evidence links, findings, risk decisions, and Principal Architect decision. |

A missing entry is a blocking evidence gap, not an assumption that the corresponding data or control is
absent, safe, or compliant.

## Required Security Review Questions

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
| SR-08 | Does ten-year audit evidence stay minimal, redacted, restricted, and non-reversible? | Evidence schema, access review, retention policy, sampled records, and linkage analysis. | No retention policy or lifecycle completion approval. |
| SR-09 | Are privacy effects resilient to malicious or mistaken insiders and support/operations access? | Role/capability matrix, access reviews, support/export controls, monitoring, incident response, and organizational-access analysis. | No production operations authorization. |
| SR-10 | Are legal hold, incident, and regulatory requirements reconciled without preserving a hidden recovery path? | Approved legal/security decision, explicit exception model, retention evidence, and fail-closed behavior. | No policy or implementation approval. |
| SR-11 | Can only a freshly authorized subject initiate or cancel a deletion lifecycle without anti-forgery, replay, enumeration, or status-oracle abuse? | Fresh re-authentication/factor design; anti-forgery/CSRF boundary; rate-limit and abuse controls; non-oracular response and timing model; cancellation rules that reject a revoked session or opaque identifier alone; negative tests and redacted audit evidence. | No deletion/cancellation API, lifecycle migration, or ADR-008 recommendation. |
| SR-12 | Do the non-authorizing migration/data-state design and least-privilege operations runbook preserve fail-closed forward recovery across independent systems? | Data states, constraints, locking, access controls, retention, backfill/rollback limits, forward recovery, non-recreation rule, and a runbook for authorization, proof, backup expiry, isolated restore, release, outage/escalation, and evidence redaction. | No formal Security Review recommendation, ADR-008 decision, migration, or lifecycle implementation proposal. |

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
| reviewable_with_findings | The evidence package is sufficient for a formal Security Review to record findings, but no acceptance is implied. | Conduct the human Security Review and record its independent verdict. |
| accepted_for_decision | Every mandatory predecessor input and SR question is verified, no blocked_evidence finding remains, the formal Security Review records its recommendation, and the Principal Architect explicitly accepts ADR-008. | Consider only a separately reviewed and explicitly authorized implementation proposal; no runtime work is implied. |

Only an authorized human Security Review and explicit Principal Architect decision may create
accepted_for_decision. It is unavailable while any mandatory predecessor input, including SR-11 or
SR-12, is absent, stale, unverifiable, or blocked. A documentation review, CI success, PR approval,
provider marketing claim, or local test run cannot do so.

## Review Record Requirements

A future formal Security Review record must include:

1. the immutable scope and source baseline;
2. named accountable security, architecture, privacy, operations, and legal participants;
3. each SR question, evidence link, evidence limitations, finding, severity, owner, and due date;
4. explicit treatment for every residual risk: accepted by whom, reduced, transferred, or blocked;
5. provider, backup, restore, and organizational-access assumptions;
6. an adversarial restore rehearsal result, including stale, missing, conflicting, forged,
   reordered, and unavailable marker/proof states;
7. an ADR-008 recommendation distinct from the Principal Architect acceptance decision; and
8. a clear statement that no deployment, migration, provider selection, or runtime change is
   authorized until the required separate proposal and human authorization exist.
9. the SR-11 authority, anti-forgery, rate-limit, cancellation, and status-oracle evidence; and
10. the SR-12 migration/data-state and operations-runbook evidence.

## Scope Boundary

This stage adds documentation only. It does not change Go, Next.js, Python, OpenAPI, PostgreSQL,
Redis, migrations, CI, infrastructure, credentials, backups, providers, operational access, or
product behavior. It does not perform a Security Review, legal review, production discovery,
provider trial, backup deletion, restore, key operation, account deletion, anonymization, or
ADR-008 acceptance.

This dossier was strictly reviewed and squash-merged through PR #53 at
`544ad8cc7371caf93913ea7716f3feb68be0ea44`. Its successor, Stage 3.25, defines the evidence
collection process without collecting evidence or conducting Security Review. Privacy-lifecycle
implementation remains prohibited until formal Security Review and ADR-008 acceptance are recorded.

## Repository revalidation — 2026-10-16 (proposal, blocked_evidence)

**Scope:** protected `develop@769d57ed73dece68b93622307aee61f06d2259d8`, tree
`9cf2189b2508e7ca838509b523adcc27f590cada`; ten required contexts. Documents 43 → 42 → accepted
ADR-005/006 remain controlling. ADR-008 is Proposed. Stages 3.17–3.24 historically completed their
respective planning/documentation activities.
Stage 3.25 remains `Active / proposal only`: publication of its evidence-collection plan does not
prove completed evidence collection. Formal Security Review, ADR-008 acceptance and implementation
remain blocked; no runtime control or Security Review acceptance is established.
No deployed data, provider account, keys, backups or operator credentials were accessed.

### Integrity defects and exact restoration provenance

Commit `8bcb2d1bba3454de1262ddc4e045fddf299a9564` removed substantive requirements from ADR-008
and the nine Stage 3.17–3.25 records. Its parent retains their exact pre-loss text. This change restores
that content in place, retaining existing statuses/versions and source-linked historical owner/review-date
fields; it does not amend frozen commits, accepted ADRs, migrations or historical audit evidence.
ADR-008's seven decision clauses, Stage 3.24 SR-08/SR-12 and readiness outcomes, Stage 3.21 audit
inventory, Stage 3.22 evidence checklist, and the emptied Stage 3.25 collection plan are restored.
The original ADR source is `fdf74c16446e7623f76882aa7add64554141abc6`; the dossier source is
`544ad8cc7371caf93913ea7716f3feb68be0ea44`; the collection-plan source is
`213d1d9b4369a5e046b26c3a08990aa571603eaa`. Restoration is not a new approval.
Stage 3.25's historical Active/proposal wording is not authorization to collect external evidence now.

### Current evidence and approval effect

The following are repository observations, not deployed guarantees. Exact requirement/proof clauses
remain in the original SR table above; every clause, including SR-12, remains mandatory.

| Requirement | Verified repository evidence / failure scenario | Missing proof and required next decision | Disposition |
| --- | --- | --- | --- |
| SR-01 — no reidentification | Migrations 000001–000015 retain plain user/subject links, user audit IDs, text and correlation fields; see the updated Stage 3.21 inventory. Removing a user row leaves joinable and indirectly identifying material. | Field-level transform/protection policy for every retained field and real-copy lineage; residual financial-pattern anonymity analysis. | Architecture/implementation approval BLOCKED; deployment evidence also missing. |
| SR-02 — key non-recovery | No subject custody adapter, erasure key or destruction verifier exists. Shared backup encryption cannot revoke one subject. | Non-exportable subject boundary, all versions/wrappings/copies, independent attestation trust, recovery/escrow limits and provider-specific capabilities. | Architecture decision BLOCKED; provider/deployed proof later mandatory, not waived. |
| SR-03 — marker integrity/completeness | Stage 3.23 is a proposal; no independent acceptance log, signer or completeness verifier exists. A valid signature over an incomplete snapshot would still omit a deletion. | Independent durable intent receipt, sequence/epoch and signed completeness watermark, pagination/count/root reconciliation, anti-rollback and release-cut fencing. | Marker design approval BLOCKED; implementation/rehearsal pending. |
| SR-04 — restore targeting without identity lookup | No replay-binding schema exists; arbitrary random handles alone neither locate recovered rows nor prove unlinking. | Restore-only random binding carried by protected material; demonstrate targeting and prove that binding/commitments cannot join a retained person map. | Architecture decision BLOCKED. |
| SR-05 — monotonic completion | Only active/pending_deletion/deleted user checks exist; no deletion request/coordinator. A database commit cannot atomically destroy external keys. | Generation-fenced state/lock and proof protocol below; persistent partial-failure tests. | Lifecycle implementation approval BLOCKED. |
| SR-06 — isolated restore release | No repository restore gate/runbook enforces network/job/credential quarantine or a separate releaser. Restoring a database alone could revive access. | Sealed restore entry, independent evidence verifier and release authority, topology/credential checks and rehearsal. | Architecture/runbook approval BLOCKED; physical isolation remains deployment-only. |
| SR-07 — all copies ≤90 days | Compose volumes are development configuration; no complete deployed backup/WAL/PITR/export inventory or destruction evidence exists. Legacy plaintext copies cannot be erased by a newly introduced key. | Per-copy birth/expiry/replication register and reconciled destruction; legacy-copy drain barrier; legal exception treatment. | No completion/compliance claim; provider/operations BLOCKED. |
| SR-08 — ten-year non-identifying audit | `recordAuthAudit` writes user UUID as actor ID; migration 000013 preserves session IDs independently of session rows. Audit events are append-only to runtime. | Prospective non-identifying evidence schema and separately approved historical personal-field treatment; no silent ledger/audit mutation. | Retention/anonymity architecture BLOCKED. |
| SR-09 — insider/support boundary | Runtime role denies ledger mutation and identity deletion; it does not define custody, marker signer, restore or support privileges. An administrator with copies can defeat an unproven boundary. | Capability/organizational separation matrix, independent root of trust, export/support controls, break-glass and incident evidence. | Design unresolved; production authorization BLOCKED. |
| SR-10 — legal/retention conflicts | Document 43 requires full identity deletion, ten-year audit, indefinite anonymous financial history, backup maximum 90 days. No approved legal-hold exception model is evidenced. | Human legal/privacy decision; a hold cannot silently keep recoverable identity while reporting completed deletion. Conflicting duty stays blocked. | Policy/architecture approval BLOCKED; no legal conclusion here. |
| SR-11 — fresh deletion/cancellation authority | Login/refresh/logout and CSRF controls exist, but no fresh deletion factor or cancellation/status route exists. Stateless access JWT remains valid until expiry without account-state lookup. | Fresh-factor ceremony, one-use action binding, CSRF, bounded attempts, uniform non-oracular replies and cancellation rules; all protected-path denial design. | API/ADR recommendation BLOCKED; existing auth tests are not deletion proof. |
| SR-12 — migration/operations safety | Expand→Populate→Switch→Validate→Contract and immutable financial facts are established; current role cannot perform planned erasure. No integrated privacy runbook exists. | Reviewed schema/locking/backfill/role proposal, rollback fence and integrated operations/evidence runbook below. | Implementation and formal acceptance BLOCKED. |

No SR is marked PASS. Repository inspection closes uncertainty about current code/schema and restores
requirements; it cannot replace missing implemented, provider, legal or operational proof. Under the
existing dossier, these missing inputs block ADR acceptance, not merely eventual deployment.

### Provider-neutral design candidate and alternatives

**Proposed hierarchy:** an independently administered custody root protects non-exportable per-subject
erasure keys; the subject boundary protects every reversible identity/link/annotation reference, not
merely whole backup media. Separate data-encryption keys may be wrapped only under that subject's
boundary; every alternate wrap, key version, escrow, cache, diagnostic and recovery copy must be
accounted for. A surviving wrap under a shared master key defeats subject erasure. Plaintext already
exported to an administrator or legacy backup is not revoked by destroying a future key.

Destruction means terminal irreversible destruction of all effective material, not disable, scheduled
deletion, a renamed alias, HTTP success or a console screenshot. Evidence binds random restricted
handle, lifecycle generation, exact intent digest, material scope, policy, terminal time and independent
attestation/verifier. Reuse of a handle with new material is forbidden. Attestation verifies a claim
under an assessed trust boundary; it is not mathematical proof that a malicious provider kept no copy.
Use [NIST SP 800-88 Rev. 2](https://doi.org/10.6028/NIST.SP.800-88r2) as a sanitization-assurance
reference, not as certification of this design or a provider. No algorithm/provider is selected here.

| Option | Trade-off / decision required |
| --- | --- |
| Delete rows/cascade; keep UUIDs | Simple, but old copies, audit and financial patterns remain identifying; insufficient for Document 43. |
| Shared database/backup encryption | Operationally simpler; cannot establish per-subject destruction while shared recovery keys survive. |
| Subject custody boundary plus independently protected marker plane | Preferred candidate; isolates revocation but adds availability/cost, migration, privilege and proof dependencies; key destruction alone cannot anonymize financial patterns. |
| Application-owned exportable keys | Cheaper integration but ordinary application/backup compromise retains recovery capability; fails the proposed separation requirement. |
| Delete financial history or coarsen immutable facts | May reduce correlation but violates existing retention/immutability decisions; not authorized. Any alternative needs an explicit higher-authority decision, not an implementation shortcut. |

### Deletion state, authorization and concurrency candidate

1. **Initiate:** authenticated active subject plus separately approved fresh factor; bind an expiring,
   single-use confirmation to subject, action, lifecycle generation and CSRF context. Never store/log
   the factor. Idempotency binds normalized non-secret intent; same request replays one result, a new
   key cannot reset the 30-day grace deadline. Errors/status must not disclose another account.
2. **Pending:** serialize on the same account/lifecycle authority used by refresh/login and all
   portfolio/import/valuation writes; revoke sessions and deny normal protected work immediately,
   as Stage 3.17 requires. A cancellation-only capability cannot authorize business APIs. A revoked
   session or opaque request ID alone cannot cancel. Fresh-factor/recovery ceremony remains a human
   decision; recovery availability is not assumed from the existing login feature.
3. **Cancel:** only before the recorded deadline and irreversible completion claim; serialize with
   expiry on the same generation/version. Require fresh authorization and issue new sessions only
   after a verified active transition. Never resurrect revoked sessions. This is the sole explicit
   pre-irreversible reversal; “monotonic” means no rollback after the completion fence.
4. **Claim completing:** compare-and-set the generation/deadline and persist an independently
   witnessed erasure intent before destruction. All authorizers must reject pending/completing and
   erased states; expired caches or a valid 15-minute access JWT must not reopen access. A future
   subject-state/authorization-generation check on every protected path is necessary; current JWT
   signature validation does not implement it. No immediate-JWT-revocation change occurs here.
5. **Reconcile:** bounded idempotent key-destruction operation → independently verify terminal proof
   → remove live identity/reversible links and approved sensitive projections → publish verified
   active marker → verify all completion predicates. Persist each phase; never hold a database
   transaction open across provider calls. Failure stays completing/unavailable with safe escalation.
6. **Complete:** only after no live or recoverable identity/link/key remains under the approved scope,
   all copy and correlation obligations have proof, and independently verifiable evidence is durable.
   A surviving legacy plaintext backup, free-text identifier or unknown copy blocks completion.
   Marker activation/live unlinking is not synonymous with this stronger final completion state.

Future lock design: one per-identity/subject lifecycle fence before session/family, portfolio and
command locks, with one documented total order and row/version checks at commit. Existing advisory
user-refresh locks and portfolio locks alone do not share this fence. Review against PostgreSQL 18
[concurrency rules](https://www.postgresql.org/docs/18/explicit-locking.html); include imports,
idempotency replays, outbox consumers and late worker callbacks. A precomputed success response must
not bypass current authorization. Do not authorize destructive work merely by using SERIALIZABLE.

A crash after irreversible destruction but before database acknowledgement must reconcile forward
from the same durable intent/proof; a retry cannot provision a substitute key. Stale worker responses
must fail generation/version checks. Use bounded backoff and escalation without making deletion
silently disappear after retry exhaustion. Cancellation after claim, missing proof and legal conflict
remain non-complete; no compensating action may restore access or destroyed material.

### Marker, backup and restore candidate

Keep the marker plane outside the application backup/administrator failure domain. A writer and signer
with nominally different credentials but the same unilateral administrator do not provide independence.
The signer must reconcile accepted intents against an independent durable journal, not just sign whatever
the writer submitted. Define signed epoch, complete sequence range, count/root and pagination, policy
compatibility, trusted freshness, anti-rollback anchor and conflicting-view detection. A correct signature
alone proves neither omission-freedom nor that there are no later deletion intents.

A restore-specific release fence must cover all accepted intents through the release cut and prevent a
newly accepted deletion from falling between snapshot verification and traffic enablement. No serving,
background jobs, outbound delivery, provider credentials or support exports before independent release.
Unknown/prepared/disputed markers, missing pages, stale cut, incompatible signer/policy, forged evidence,
conflicting generations or unavailable journal/custody keep the environment isolated. Replay binding must
locate only affected protected records, without an identity lookup; its schema/correlation proof remains
an architecture blocker. Recheck destroyed-key unavailability independently after replay.

Inventory backups, PITR/WAL, replicas, snapshots, exports, key-service recovery media and copied operator
artifacts. Each managed recoverable backup copy is encrypted and expires within 90 days; an unknown
copy prevents proof. Old plaintext-link copies must be verifiably destroyed/aged out before final
completion, even after the new boundary is deployed. Minimum redacted completion evidence survives ten
years; it must not become a ten-year identity map or backup. Evidence retirement needs completeness
proof for all recoverable copies, not a wall-clock timer alone. Provider outages do not relax release.

### Future implementation boundary and verification sequence (not authorization)

| Order | Separately reviewed work | Required exit proof |
| --- | --- | --- |
| 1 | Human policy/security decisions, field-level inventory including new registration data; choose viable custody and marker trust models. | Every SR clause mapped; legal conflict and immutable-text treatment resolved; no assumed provider guarantees. |
| 2 | Additive schema proposal: lifecycle generation/version/deadline and phase constraints, unique active intent, scoped key reference, restricted replay binding, proof commitments and retry status. | No raw secret/personal marker payload; one total lock order; distinct migration/runtime/lifecycle capabilities; no runtime DELETE privilege expansion by convenience. |
| 3 | Expand/populate protected mappings; reconcile every existing identity/subject and all reversible copies; shadow-validate. | Zero unprotected active mapping at switch; controlled plaintext rollback copies inventoried; no destructive backfill or ledger UPDATE/DELETE. |
| 4 | Switch authorization fences, fresh-factor/cancellation API and generation-fenced coordinator behind reviewed contracts. | Pending denial across login/refresh/old JWT/portfolio/import and replay paths; no source-schema state mistaken for an enforced gate. |
| 5 | Independent custody/proof/marker integration, backup expiry and sealed restore runbook. | Provider capability and role evidence, loss/recovery exercises, complete release-cut protocol; no purchases/keys/production actions in this PR. |
| 6 | Validate then contract obsolete personal projections only with explicit approval; drain old backups. | Non-reidentification and copy-completeness evidence; irreversible cut forbids rollback to plaintext or key recreation; only forward recovery thereafter. |

No rollback restores destroyed keys or deleted identity data. Before irreversible work, rollback must
remain within the approved protected data state; after the cut, disable serving and recover forward.
The operations runbook must cover independent authorization, proof failure, marker outage, retry
exhaustion, backup expiry, sealed restore, release, key/signing rotation, incident escalation and
redacted evidence access. No current operator schedule/authority is invented here.

Future tests must cover: valid/invalid fresh factor and CSRF; idempotent duplicate/conflicting intent;
request/cancel/deadline races; write-before/after fence and replayed success; cross-account isolation;
crash before/after every external side effect and acknowledgement; stale worker generation; provider
schedule versus irreversible proof; all key versions/alternate wraps; lost/invalid/stale/revoked proof;
marker omission/reordering/split view/truncated pages; release-cut race; actual isolated restore from
oldest/most recent backup; legacy plaintext-copy barrier; failed expiry; insider privilege separation;
retained text/hash/timing linkage and secret-sentinel redaction; unchanged financial ledger/snapshots.
Use disposable databases/synthetic data and race detection; simulated provider success cannot replace
provider attestation or deployed backup/restore rehearsal. These lifecycle tests are NOT_EXECUTED:
there is no authorized implementation to test. Existing normal CI checks compatibility only.

**Verdict: BLOCKED for ADR-008 acceptance and implementation.** This prospective assessment is ready
for independent review of findings, not a formal Security Review verdict or Principal Architect decision.
Repository-verifiable analysis is complete for this baseline; implementation-dependent tests, exact
provider/custody proof, legal policy and real deployed-copy/restore evidence remain missing.
