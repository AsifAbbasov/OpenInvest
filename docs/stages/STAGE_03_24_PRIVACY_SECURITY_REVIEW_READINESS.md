# Stage 3.24 - Privacy Security Review Readiness Dossier

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-24-PRIVACY-SECURITY-REVIEW-READINESS |
| Version | 0.5.0 |
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

## Architecture remediation candidate v1

Scope: protected `develop@cffd2091168fa28817a49464438db01804b4ee48`, tree
`c5ada81772bb6179931636a52f88b1ae43281b91`; PR #275 is merged. This prospective design adds no
implementation or acceptance. Historical review dates above remain historically recorded / UNVERIFIED.
Earlier SR acceptance blockers remain; the matrix below adds **design maturity**, not closed SRs.

**Identifier reconciliation:** preserve the original SR table as authority. In the new task, “SR-04
Restore Safety” expands original SR-04 replay targeting plus SR-06 isolation; “SR-11 Deletion
Confirmation” includes a receipt but does not replace original SR-11 fresh-factor/cancellation authority.
All original SR-01–SR-12 proof clauses and EC-01–EC-10 remain mandatory.

### Reusable source mechanisms and current-state index

All paths below are relative to the exact baseline above; code/schema inspection is repository evidence,
not production discovery. Documents 42/43 and accepted ADR-005/006 remain byte-for-byte unchanged.

| Ref | Exact repository source / mechanism | Safe reuse and limit |
| --- | --- | --- |
| R1 | `backend-go/internal/auth/{service,tokens,password}.go`; `backend-go/internal/httpapi/{auth_handlers,security_headers}.go` | Existing signature/identity/expiry, password verification, CSRF/cookie and no-store boundaries; access auth currently returns signed subject without database lifecycle check. Preserve JWT-01/02/03 policies; deletion adds an independent current-state gate. |
| R2 | `backend-go/internal/postgres/auth_store.go`: lockUserRefreshes, lockRefreshSession, RotateSession, RevokeSession, recordAuthAudit | Reuse transaction/advisory locking, family revocation and deduplicated rejection evidence; lock order and raw audit user IDs need the new privacy boundary. |
| R3 | `infrastructure/postgres/migrations/000001_stage_03_01_vertical_slice.up.sql`, 000002 auth/privacy, 000013 audit deduplication, 000015 email ownership | Actual user/subject/link/session/audit/registration rows and dependencies. No lifecycle request, key, independent marker or restore-release schema exists. |
| R4 | `backend-go/internal/postgres/{store,stage_03_74_transactions,replay_store,replay_lookup,retention}.go` | Reuse portfolio owner checks, row locks, command reservation/replay and bounded cleanup; no common deletion fence currently covers every path. |
| R5 | Migrations 000003–000004 provenance, 000006–000007 replay/retention, 000010–000012 valuation/replay | Field-level personal/correlation and derived-state scope; financial immutability/deduplication must survive any proposed projection change. |
| R6 | `infrastructure/postgres/runtime/openinvest_runtime_role.sql`; migration `policy_manifest.json`; `docs/database/MIGRATION_STRATEGY_STAGE_02.md` | Reuse exact capability profiles and Expand→Populate→Switch→Validate→Contract; API role cannot delete identity/ledger/audit. New capability and manifest proposal require separate review. |
| R7 | `backend-go/internal/httpapi/import_security.go`; `frontend-next/src/features/auth/session.ts`, `src/app/auth/verify/page.tsx`; `backend-go/internal/sharedbudget/redis.go` | Existing token expiry, generation-aware UI and Redis atomic budgets are reusable mechanisms, not account erasure. Browser/Redis persistence is separately inventoried. |
| R8 | `docker-compose.yml`; `docs/database/ER_MODEL_STAGE_02.md`; Stage 3.21 external inventory | Local PostgreSQL/Redis volumes, canonical 30-day grace, backup/ledger constraints. No deployed backup/KMS inventory or sealed restore enforcement is proved. |

### Lifecycle, authorization and receipt contract

Candidate persisted `privacy_lifecycle` has random lifecycle ID, identity/subject links **only until
approved unlinking**, generation, optimistic version, phase, fixed grace deadline, intent digest,
external-operation reference/commitments and retry/checkpoint state. Identity-bearing fields must have
an approved erasure disposition; no ten-year retained tombstone with raw user or subject IDs. Unique
one non-final request per identity; immutable deadline; compare-and-set generation/version; phase/proof
checks. Exact SQL types, constraints, role grants and OpenAPI are future proposals, not DDL here.

| State | Admission / transition | Safety and failure result |
| --- | --- | --- |
| active | Freshly authorize REQUEST; obtain G1 PREPARE receipt before local `admission_pending`; revoke all sessions under G2, then witness activation and persist its fixed deadline before acknowledgement. | Duplicate same intent replays; different idempotency body rejects. Unknown transaction result reconciles; never creates a second grace window. |
| admission_pending | Independent PREPARE exists; local denied state committed but activation/deadline may be unknown. | No normal work, no acknowledgement of accepted deletion and no irreversible claim; reconcile same intent or witnessed abort. |
| pending | Deny normal login/refresh/protected reads and writes; permit only scoped fresh cancellation/status ceremony before deadline. | Same lock serializes cancellation with completion claim. Cancellation increments authorization generation and requires a new normal login, not session resurrection. |
| completing: claimed / witnessed | After deadline, independently witness the claim, then CAS its matching database receipt; an unknown service outcome cannot authorize destruction. | Cancellation is permanently closed. Outbox/witness outage remains non-serving and non-complete; retries cannot change intent/generation. |
| completing: key_proven / primary_unlinked / marker_active | Reconcile exact destruction operation/proof, remove approved live personal scope, then activate independently verified marker. | Crash at any boundary resumes forward; stale callback rejects; no compensating key creation or reactivation. |
| completing: waiting_evidence | Await every copy/indirect-link/legal/proof obligation, including legacy plaintext copies. | Retry exhaustion escalates without dropping obligation. No completion receipt while any condition is unknown. |
| completed | Independent conjunction: terminal key proof + live unlinking + authentic complete marker + reconciled copy scope + non-reidentification + allowed retention evidence. | Terminal; no cancellation/reactivation/rollback. Keep only approved non-identifying evidence. |

**Fresh confirmation candidate:** reuse password verification in a purpose-restricted identity port,
not regular Login (which must reject pending accounts). One-use challenge digest binds action,
principal, current generation, nonce, CSRF context and expiry; proposed maximum five minutes,
subject to security acceptance. REQUEST needs active authentication plus fresh password. CANCEL uses
only the restricted pending-identity ceremony, fresh password and CSRF/challenge proof; it never issues
normal bearer/refresh credentials. A stale/revoked session, request ID, receipt or ticket alone is not
authority. Challenge responses for unknown/ineligible identities must have the same public shape and
bounded work; re-use dummy password verification and existing IP/credential admission, with reviewed
limits and failure semantics. No account recovery, email-based cancellation or new factor is assumed.
Raw passwords/challenges never enter audit, logs, idempotency hashes or durable evidence.
Challenge consumption and intent reservation must commit atomically. After pending, a same-intent retry
uses fresh purpose-restricted authorization to replay the existing receipt/state; it does not require
or restore a normal active session. A consumed challenge never authorizes another effect. An ambiguous
commit reconciles the same intent; failed authorization cannot inspect another principal's replay key.

Application `AuthorizeSubjectOperation` is defense in depth, not the compatibility boundary. G2 below
must independently enforce denial even if that call is omitted. All compatible application access
must also go through authoritative `AuthorizeSubjectOperation` inside the relevant
transaction, **before** owner lookup, returning idempotent success or external side effects. Candidate
lock order: subject lifecycle fence → identity refresh advisory lock → session/family → sorted portfolio
IDs → command scope; ordinary operations take a shared fence and deletion/cancellation takes exclusive.
Maintain one canonical lock key/row, not two independent user/subject locks that can diverge. Reads
hold their fence through result authorization; writes hold it through commit. Re-check state/generation
with the lock acquired on the primary before using a repeatable-read financial snapshot. A stale
snapshot/cache cannot supply lifecycle authority. If this cannot be guaranteed in a transaction mode,
redesign that path before rollout. Long-running streams need explicit revocation/termination and bounded
admission; do not claim retracting already transmitted bytes. Worker/outbox admission uses the same
scope/generation; an already submitted email/provider request cannot be unsent and stays inventoried.

This adds a future lifecycle check without changing financial authorization's validated
investmentSubjectId, token signature/claim policy or current refresh semantics. It requires separate
approval for the new access-denial contract; no change to the accepted future-iat risk is implied.

**User confirmation receipt:** acceptance returns non-final status only over authenticated TLS/no-store;
completion can issue a signed statement with random receipt nonce, policy/schema, time and final
non-identifying evidence commitment. No user/subject/portfolio/key/marker locator or public status
lookup. A receipt proves only the attested scoped result under trusted issuer semantics, not universal
erasure. Before identity deletion, optionally bind a user-held verifier capability whose hash is stored
only in an approved restricted transition projection; possession permits receipt retrieval, never
cancellation or financial access. Capability retention, loss and correlation require A3 review. If no
safe retrieval binding is accepted, use a separately approved delivery mechanism or report receipt
retrieval unsupported; never retain an identity map for convenience. No receipt is returned as final
while completing. A replayed receipt is not evidence for another generation.

### Required migration and operations proposal

1. **Resolve A1/A2/A3/A4 first.** Inventory actual stored/copy fields; no schema change cures unknown
   provenance, personal immutable text or missing provider proof by assertion.
2. **Expand:** propose lifecycle/fence, protected identity-resolution, scoped material versions,
   transactional intent outbox, proof/checkpoint projection and restore bindings with no public lookup.
   Reuse explicit migration manifest, bounded locks/batches and owner/runtime separation. No generic
   DELETE/UPDATE grant to API, no destructive DOWN that reconstructs erased material.
3. **Populate and validate:** owner-only bounded migration inventories every active subject; protect
   permitted identity/annotation mappings; compare decrypted resolution and all financial results in
   shadow mode. Do not rewrite amounts/dates/rounding/ledger sequence/revisions/source identity or old
   audit events. If A1/A2 need a contract amendment, stop that part until approved.
4. **Switch by the mandatory G2 cutover below:** no deletion operation is enabled until database
   policies, credential/role revocation, connection/worker drain and independent artifact admission
   are verified. A readiness claim or an application-only generation check is not enforcement.
   Dual-read plaintext fallback is forbidden after the switch/cut. Capabilities/grants follow exact
   explicit profiles, never PUBLIC or schema-owner credentials in normal runtime.
5. **Validate custody/journal/restore before destructive activation:** independently test synthetic
   erasure, all versions and recovery paths, witness omission, release-cut races, role isolation and
   copy expiry; establish provider/operations evidence and pause/escalation authority.
6. **Contract only with authorization:** remove obsolete personal plaintext paths after validation,
   reconcile all legacy plaintext backups/WAL/exports/indexes and applied copies; preserve financial
   ledger/snapshot meaning. Unknown copies keep lifecycle completing. The 90-day bound is not reset
   by copying a backup. Historical audit treatment follows A2, not a privilege shortcut.

Before irreversible activity, rollback may stop rollout and leave additive structures unused under
the approved protected state; it cannot silently re-enable plaintext access. After key destruction or
personal erasure, only forward repair is permitted. Old database restore stays sealed and applies the
independent markers; rolling back a binary cannot bypass the fence. Required operational handoffs:
authorizer admits → independent witness acknowledges → custody authorizes/executes → verifier checks →
lifecycle unlinks → marker authority publishes → backup owner reconciles → restore verifier rehearses →
independent releaser signs. Each failed handoff has a durable non-identifying obligation and an assigned
role, not an invented person/date. Provider outages, unknown clocks or split views stop progress.

### SR remediation and acceptance matrix

DESIGN_CANDIDATE_DOCUMENTED means a proposal is written, not accepted, implemented or independently
verified. PARTIALLY_DESIGNED means a technical binding is unresolved. EXTERNAL_DEPENDENCY identifies
necessary provider/platform/organizational proof. BLOCKED means a mandatory prerequisite or decision
is unresolved. These labels never mean production authorization. Every row's final acceptance remains blocked until its original minimum
proof clauses pass independent review.

| SR / design status | Current source → control / threat | Alternatives, dependency and sequencing | Test → EC evidence → explicit acceptance / missing proof |
| --- | --- | --- | --- |
| SR-01 BLOCKED | R3/R5/R7 → field disposition + protected identity port + copy/indirect-link completion barrier; prevent a surviving note/hash/JWT/audit bridge. | Cascade/UUID insufficient; immutable personal-field rewrite/generalization needs A1/A2/A4 before population. | T1/T7 → EC-02/10: every field/copy and reasonable join resolved; no anonymity claim while personal immutable facts or external correlation remain. Real discovery/approval missing. |
| SR-02 PARTIALLY_DESIGNED | R3/R6 → subject material isolation and independent terminal proof in Stage 3.22; prevent alternate-wrap/admin recovery. | Non-exportable provider vs independently operated hardware; A3 before selection/creation. | T3 → EC-03/10: all versions/copies/non-recreation and proof trust accepted; simulated negative decrypt does not establish provider destruction. |
| SR-03 BLOCKED | R3/R8 → independent admitted-intent journal/checkpoints and receipt reconciliation in Stage 3.23; reject signed omissions/split views. | Single writer-signer rejected; independent witnesses/threshold needs A3/G1; PREPARE before local pending, complete admitted-intent coverage before release. | T4/T5 → EC-04/10: all admitted generations through release cut included, replay/rollback rejected; real independent durability/admin evidence absent. |
| SR-04 PARTIALLY_DESIGNED | R3/R5 → restore-only binding + protected scope replay; prevent marker becoming a person lookup or skipping recovered material. | Raw user/subject IDs rejected; exact proof/lookup/retention join graph needs A3 before schema. | T1/T4 → EC-04/10: complete targeting without any usable residual map, including old locator/proof joins. Exact binding and provider fixtures missing. |
| SR-05 BLOCKED | R1/R2/R4 → fenced lifecycle saga and generation-bound callbacks; prevent write/cancel/completion races and false success. | Long distributed transaction rejected; G1/G2 required before enabling the monotonic saga. | T2/T3 → EC-05/10: every crash cut/retry/race preserves denial and no false completion; PostgreSQL/coordinator implementation tests pending. |
| SR-06 EXTERNAL_DEPENDENCY | R8 → sealed instance and independent release certificate tied to configuration/cut; prevent early ingress/jobs or copied release evidence. | Manual operator assertion rejected; platform-enforced deny before separately authorized release. | T5 → EC-06/10: rehearsal proves every serving/credential/job route denied before valid release; topology enforcement evidence missing. |
| SR-07 EXTERNAL_DEPENDENCY | R8 → per-copy lineage/expiry/destruction reconciliation plus legacy-copy barrier; prevent hidden/re-aged recoverable copy. | Timer alone rejected; provider evidence and independent inventory; A4 for policy conflicts. | T6 → EC-07/10: all managed copies encrypted and destroyed within 90 days with proof; actual inventory/destruction missing. |
| SR-08 BLOCKED | R2/R3 → minimized future privacy evidence with restricted commitments; prevent user/session/request/time joins over ten-year audit. | Appending a redacted projection leaves original IDs intact; historical mutation/encryption requires A2 and legal review first. | T1/T7 → EC-02/10: ten-year integrity/retention without prohibited map; original-row disposition and sampled linkage proof missing. |
| SR-09 EXTERNAL_DEPENDENCY | R6 → role separation across application, lifecycle, custody authorizer/executor, marker writer/witness, restore/releaser and evidence verifier. | Shared-admin “separate credentials” insufficient; A3 organizational boundary before provider integration. | T8 → EC-08/10: no unilateral routine role combination can forge release/recover keys; actual IAM/support/access-review evidence missing. |
| SR-10 BLOCKED | R8 + Documents 42/43 → explicit exception register and no false completion; prevent hidden hold/recovery path or extended backup retention. | No assumed regulatory retention right; qualified legal/privacy and Principal Architect resolve A4 before any conflicting operation. | T6/T7 → EC-09/10: signed policy decision reconciles duties without silent contract change; legal decisions absent. |
| SR-11 PARTIALLY_DESIGNED | R1/R2/R7 → fresh purpose-bound password/challenge, restricted cancellation, non-oracular status and non-identifying receipt; prevent stolen-session/replayed receipt authority. | Opaque ID alone rejected; exact receipt capability/privacy binding and recovery/loss UX need A3/owner decision. | T2/T7 → EC-01/10: original fresh-factor/CSRF/abuse/cancellation requirements plus no final receipt before evidence; HTTP/UX/receipt implementation tests pending. |
| SR-12 BLOCKED | R3/R4/R6 → staged rollout/compatibility fence, owner-only backfill, immutable ledger and forward-only irreversible cut; prevent old binaries/plaintext restore rollback. | Destructive one-shot migration rejected; A1–A4 plus G1/G2 resolve before activation. | T2/T5/T8/T9 → EC-05/10: full migration/data-state/runbook clauses independently verified, no lost financial facts, old copies drained; runnable migration/rehearsal missing. |

### Approval register and next independent scope

A1: Principal Architect/data/privacy decision for personal fields embedded in immutable financial facts,
including deduplication/source-rights effects. A2: Principal Architect/security/legal decision for legacy
audit IDs and ten-year evidence. A3: exact custody/proof/replay/receipt binding and independent control
plane, provider/cost/operations selection and assurance. A4: qualified legal/privacy retention/hold and
managed/external-copy threat scope. No current requirement is weakened or silently replaced.

Independent Gatekeeper should review these four decisions and protocol counterexamples, original
SR/EC clause coverage, code traceability and model limitations. Ask for specific implementation/provider
rehearsals only in separately authorized phases. Documentation merge must not be read as ADR acceptance,
legal approval, key provisioning, deletion API implementation or production readiness.

### Mandatory compatibility enforcement G2

**ARCH-277-02: design candidate only / enforcement gate BLOCKED.** The current runtime role grants
raw SELECT/INSERT on financial tables and no privacy RLS/trigger gate. Old binaries retain that path
unless privileges and running sessions are actually removed. A PostgreSQL login identifies credentials,
not the executable holding them; a borrowed new credential cannot prove a new binary. G2 therefore
requires both database policy and independently controlled workload/credential/ingress admission.

| Required layer | Proposed mandatory contract / limitation |
| --- | --- |
| Trusted database epoch | Owner-controlled compatibility manifest binds database/cluster generation, admitted login roles, policy/schema hashes and enabled operation classes. Derive runtime identity from server-authenticated `session_user`, not headers, `application_name`, caller-set GUC or a submitted version string. Runtime cannot change this manifest or disable enforcement. |
| Versioned credentials | Distinct short-lived LOGIN identities for API/worker deployment generations, no shared legacy login, no SET ROLE path to obsolete/bypass capabilities. A trusted workload broker admits only approved immutable artifact/configuration measurements and rotates/revokes identities. Artifact/attestation and broker enforcement need actual infrastructure evidence; database SQL alone cannot provide it. |
| Database lifecycle fence | Future forced RLS USING/WITH CHECK plus mutation guards cover every subject-bearing base table/partition, owner lookup, auth/session issuance/rotation, replay/derived/worker path and identity resolution. Policies traverse portfolio/subject scope and reject absent/pending/completing/erased states even when old SQL omits the application helper. Shared reference tables have separately approved non-personal policy. |
| Transaction serialization | A reviewed invoker gate takes a conflicting shared lifecycle-row lock and checks current state before SELECT visibility or mutation; completion takes exclusive lock. Repeatable-read conflict aborts/retries from fresh authority, never accepts an old active snapshot. Mere RLS snapshot lookup, advisory locks taken only by cooperative code, or a trigger covering only writes is insufficient. Exact SQL/function privileges and planner/snapshot behavior require disposable PostgreSQL proof before approval. |
| Constrained capabilities | No owner, SUPERUSER, BYPASSRLS, TRUNCATE, policy/trigger disable, replication, SECURITY DEFINER escape, database/schema CREATE or session-authorization escalation for runtime. Preserve ledger/audit INSERT-only financial protections; do not grant UPDATE/DELETE on financial facts to implement a fence. Any lock-helper privilege is minimal and cannot update lifecycle authority; otherwise choose a separately approved trusted database mediation boundary. |
| Egress and caches | Drain old API/worker processes, pools, queued messages and cached responses. An old process serving in-memory data or holding SMTP/provider credentials bypasses database policy; ingress and egress brokers must reject obsolete workloads and scope/generation leases. Already accepted remote work cannot be recalled and stays inventoried. Development auth bypass is disabled in admissible deployment configuration; its omission of API auth still cannot skip database subject denial. |

A proposed invoker locking helper must be shown safe without exposing new SECURITY DEFINER paths,
which current runtime readiness rejects. If safe column-level lock privileges, RLS table coverage or
snapshot correctness cannot be demonstrated, G2 remains BLOCKED and deletion is not enabled. RLS is
not superuser/owner protection: privileged maintenance stays separately authorized, isolated and
independently audited, with no routine serving credentials. Colluding or compromised infrastructure
administration is not solved by a PostgreSQL policy and remains A3/SR-09 evidence.

**Safe cutover (future runbook, not executed):**
1. Keep privacy admission/destruction disabled. Expand the policy/epoch structures and test compatible
   binaries/workers, raw SQL, pools, replicas and direct paths under synthetic data; establish G1.
2. Quarantine ingress and outbound work; drain every old worker/API, transaction and connection pool.
   Deny old LOGIN/CONNECT, revoke old memberships/table/function paths and invalidate cached secrets.
   NOLOGIN/CONNECT revocation alone does not end existing sessions: independently enumerate and
   terminate old backends, confirm none remain, and resolve prepared transactions under owner review.
   Do not assume prepared transactions are gone when their client connection closes.
3. While traffic stays sealed, atomically commit the new database epoch/admitted-role manifest,
   explicit grants and enforced policy set. Reconstruct maximum capabilities without broad defaults;
   verify every base-table, partition, function and inherited/PUBLIC grant. Failure rolls back that
   database transaction but leaves the environment sealed; no distributed atomic deployment is claimed.
4. Admit only broker-verified compatible workloads with new identities; negative-test old credentials,
   old binaries using old/new credential paths, old workers, direct SQL, stale access/refresh sessions,
   development bypass and cached responses. A credential password change is not a completed cutover.
5. Independently verify policy hashes, effective privileges, zero obsolete connections/prepared work,
   witness coverage, worker/ingress/egress admission and denial for pending subjects. Only then may a
   separately authorized release enable privacy admission. Unknown policy/broker/epoch state denies.

Rollback can return to a tested compatible artifact at the same enforced policy/epoch; it cannot
regrant the legacy role, disable RLS, reuse stale sessions or install an incompatible binary. After
irreversible activity, forward-only reconciliation applies. At recovery, old credentials restored in
a database image are not release evidence: G2 is re-established while sealed before serving anything.
A1–A4 remain unresolved; this design creates no credential, policy, provider selection or migration.

Primary references inspected 2026-10-16: PostgreSQL 18
[row security](https://www.postgresql.org/docs/18/ddl-rowsecurity.html),
[row locking](https://www.postgresql.org/docs/18/explicit-locking.html),
[role changes](https://www.postgresql.org/docs/18/sql-alterrole.html),
[backend termination](https://www.postgresql.org/docs/18/functions-admin.html) and
[prepared transactions](https://www.postgresql.org/docs/18/view-pg-prepared-xacts.html).
These references establish mechanisms/limitations, not that this repository implements G2.

### Recovery invariants and targeted finding disposition

G1 safety: no locally committed privacy pending state without independent PREPARE; no external
pending acknowledgement without witnessed activation plus matching database checkpoint; one immutable
30-day deadline per accepted generation; no irreversible claim before it; unknown outcome is not
rejection; incomplete epoch coverage or unresolvable binding quarantines recovery; authenticated
pending/claimed state cannot restore as unrestricted active. Only a witnessed, freshly authorized
cancellation before the deadline can undo pending. No key or erased identity is recreated to reconcile.
Conditional liveness requires service quorum, trusted time, complete replay binding and authorized
operators; permanent partitions or missing evidence may block forever. Safety never assumes recovery
of missing WAL or a vanished independent witness.

| Finding | Targeted correction / evidence route | Remaining acceptance gate |
| --- | --- | --- |
| ARCH-277-01 | Stage 3.23 admission v2, G1 epoch coverage, explicit DB/witness/unknown recovery cases; v2 model and T3/T4/T5 → EC-04/05/06/10. | BLOCKED until independently durable admission/binding and recovery proof exist. |
| ARCH-277-02 | G2 database + workload credential/ingress gate and sealed cutover; policy table and T2/T8/T9 → EC-01/05/08/10. | BLOCKED until actual PostgreSQL and stale-binary/worker/drain evidence exists. |
| ARCH-277-03 | Matrix now uses candidate/partial/external/blocked terminology; SR-03/05/12 blocked, SR-06/07/09 external. | Classification correction is repository-verifiable; no SR accepted/implemented/production-authorized. |
| ARCH-277-04 | Executable v2 covers timed grace, unknown admission, pending restore, missing WAL quarantine and compatibility/release predicates. | Abstract consistency only; T2/T3/T4/T5/T8/T9 still require real integration evidence. |

## A1–A4 / G1–G2 feasibility disposition — no acceptance

Baseline `5a83503719e6756c57707afbcda309b1a83f1d67`, tree `85fcc4c0fd0074da19e98a1ac0238aac9423842e`;
PR #277 is merged. Original SR clauses, historical provenance, A1–A4, Documents 42–43 and accepted
ADRs remain authoritative. Date of source/prototype inspection: 2026-10-16 UTC, not a formal review date.

Evidence vocabulary: **DESIGN_ONLY** = specified proposal; **PROTOTYPE_VERIFIED** = named synthetic
invariant executed; **INTEGRATION_VERIFIED** = named integration and environment actually executed;
**EXTERNAL_EVIDENCE_REQUIRED** = provider/deployed/operator proof absent; **BLOCKED** = prerequisite
or authorization missing; **ACCEPTED** requires a separate authorized decision (none here).
These evidence labels do not replace SR design-maturity classifications or grant production authority.

### G2 tested mechanism and mandatory remaining boundary

The isolated [Go experiment](../../backend-go/experiments/privacyfeasibility/g2_test.go) creates its own
randomly named database and LOGIN role on loopback PostgreSQL 18, populates two synthetic subjects,
and removes them on completion. It applies no application migration or policy. It uses a SECURITY
INVOKER VOLATILE helper, server `session_user` epoch membership, forced row-level policy, SELECT FOR
SHARE on a lifecycle row, and a constrained `UPDATE(lock_token)` grant (CHECK token=0) solely to allow
row locks. Runtime cannot update lifecycle state or ledger facts. A completion state update conflicts
with the shared row lock; a stale Repeatable Read snapshot must abort with SQLSTATE 40001.

This candidate mechanism is useful only if complete table/function/query coverage, lock privileges,
snapshot semantics, planner behavior, revocation and absence of owner/BYPASSRLS paths are proven.
The experiment is not application coverage: only its one synthetic ledger is tested. The role can
see synthetic lifecycle rows; production lifecycle metadata privacy remains an explicit minimization
and policy design task. The helper does not attest executable identity, user principal or real session.
A workload supplying allowed credentials is indistinguishable at SQL identity level; DB lifecycle denial
can restrict pending subjects but independent credential/artifact/ingress/egress admission is still required.

Current runtime sources `runtime_privileges.go` and `runtime/openinvest_runtime_role.sql` already reject
privileged roles/SECURITY DEFINER paths and preserve financial/audit read+append permissions. Reuse
that validation pattern; changing the exact capability profile or policy manifest needs separate review.
An experiment with a synthetic role does not authorize broadening existing application capabilities.
Do not generalize the successful lock privilege to every partition, auth/session, replay or worker path.

Retain G2 sealed cutover: disable deletion; install/test candidate policy; isolate ingress/egress; drain
old workers/pools/caches; revoke legacy LOGIN/CONNECT/memberships/grants, terminate existing backends,
reconcile prepared transactions; commit epoch/policy/grant manifest; admit only independently verified
artifacts; verify effective permissions and all negative paths before separate release. Restored old
credentials are never current release evidence. Rollback stays in the compatible epoch or stays sealed.
Superuser/BYPASSRLS, unforced owners, replication and external caches are outside the RLS guarantee;
TRUNCATE/REFERENCES and privileged functions need explicit capability denial. No old Go binary/worker,
development-auth execution path, artifact broker, prepared-transaction draining or real session-revocation
campaign was run here; those stay required T2/T5/T8/T9 integration evidence, not PASS.

### Revised SR maturity and evidence

| SR | Design maturity | New evidence / unresolved gate |
| --- | --- | --- |
| SR-01 | BLOCKED | A1 field/alternative dossier DESIGN_ONLY; immutable personal fields and economic correlation still require approved disposition. |
| SR-02 | PARTIALLY_DESIGNED | A3 suite/manifest candidate DESIGN_ONLY; no provider destruction/non-recreation evidence. |
| SR-03 | BLOCKED | G1 local journal properties PROTOTYPE_VERIFIED; independent quorum, complete epoch and binding remain EXTERNAL_EVIDENCE_REQUIRED. |
| SR-04 | PARTIALLY_DESIGNED | Restricted restore capability/locator lifetime proposed; real complete-scope mapping and privacy proof missing. |
| SR-05 | BLOCKED | Synthetic PREPARE/grace/recovery and candidate SQL policy evidence only; full application lifecycle/auth/DB integration absent. |
| SR-06 | EXTERNAL_DEPENDENCY | Release/barrier design exists; actual network, credential and workload isolation unverified. |
| SR-07 | EXTERNAL_DEPENDENCY | A4 managed-copy boundary inventoried; no provider inventory or <=90-day destruction evidence. |
| SR-08 | BLOCKED | A2 field disposition and migration alternatives DESIGN_ONLY; append-only historic identity conflicts unresolved. |
| SR-09 | EXTERNAL_DEPENDENCY | Required role split specified; actual organizational separation, workload broker and provider access unverified. |
| SR-10 | BLOCKED | A4 exact decision questions provided; no approved legal exceptions. |
| SR-11 | PARTIALLY_DESIGNED | Receipt binding specified; full fresh-factor/cancel/CSRF/privacy and user-claim tests absent. |
| SR-12 | BLOCKED | Isolated lock/privilege feasibility is not accepted rollout or app migration; A1–A4/G1/G2 must resolve first. |

No SR is implemented, independently accepted or closed by these documents. Recommended focused
Gatekeeper scope: audit source-field completeness and A1/A2 contract conflicts; challenge G1 durability,
coverage/release assumptions; rerun G2 on PostgreSQL 18 and inspect invoker privilege/snapshot boundaries;
validate A3 privacy/retention binding and A4 authority questions. Actual provider, Legal/Privacy,
Operations/Security and Principal Architect decisions remain mandatory before implementation approval.
