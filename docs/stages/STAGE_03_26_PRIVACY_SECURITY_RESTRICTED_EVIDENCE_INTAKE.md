# Stage 3.26 - Privacy Security Restricted-Access Evidence Intake

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-26-PRIVACY-SECURITY-RESTRICTED-EVIDENCE-INTAKE |
| Version | 0.1.0 |
| Status | Active / intake authorization only |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.25 evidence-collection plan |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.25 privacy proposals |
| Last Review Date | 2026-08-18 |
| Next Review Date | Before any evidence receipt, access grant, source inspection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

## Purpose

The repository user authorized initiation of the restricted-access intake required by Stage 3.25.
This authorization starts a controlled process for requesting real evidence; it does not grant
production, provider, backup, legal, identity, or credential access. It also does not designate an
accountable organizational owner, approve a source inspection, or authorize collection until every
item below is satisfied.

This stage turns the Stage 3.25 process into an auditable intake boundary. It records the canonical
source baseline, requires named accountable owners and independent verifiers, limits collection to
approved minimized requests, and preserves `blocked_evidence` when any prerequisite is absent. It
does not collect, inspect, store, or validate operational evidence; perform a Security Review;
accept ADR-008; select a provider; or authorize privacy-lifecycle implementation.

## Frozen Intake Baseline

| Item | Value |
| --- | --- |
| Repository baseline | `develop` at `213d1d9b4369a5e046b26c3a08990aa571603eaa` after Stage 3.25 PR #54 |
| Controlling requirements | Documents 42-43, proposed ADR-008, Stage 3.20 threat model, Stage 3.21 inventory, Stage 3.22 custody proposal, Stage 3.23 marker proposal, Stage 3.24 readiness dossier, and Stage 3.25 evidence plan |
| Repository authorization | The user approval recorded in the Codex task on 2026-08-18 permits intake preparation only. |
| Authorized evidence state | No evidence received, no source inspected, no operational claim made. |
| Access state | No production, provider, backup, legal, identity, or credential access grant is recorded in this repository. |

A change to this baseline, an evidence source, a collection tool, a data classification, or a
minimum-proof clause requires a new intake request. It may not be silently inherited by a later
collection event.

## Intake Preconditions

No EC item may move out of `blocked_evidence` until its request records all of the following:

1. a named accountable owner with authority over the actual source, not just a role label;
2. a named independent verifier who cannot alter the source or approve their own evidence;
3. the exact Stage 3.24 SR minimum-proof clause(s) and Stage 3.25 EC item(s) to be tested;
4. the source system, environment, version/configuration scope, approved collection method, and
   smallest sufficient sample;
5. data classification, permitted fields, redaction method, restricted evidence location, access
   approver, retention/expiry, and destruction method;
6. a source-owner-approved least-privilege, time-bounded access path with no shared credentials;
7. an integrity and provenance method that detects substitution, truncation, replay, staleness, and
   conflict without creating a public person or marker lookup; and
8. a safe failure and escalation route if minimization, redaction, independent verification, or
   source access cannot be completed.

Repository user approval cannot satisfy a missing source-owner approval, organizational authority,
independent verifier, or restricted evidence location. Any absent, expired, conflicting, or
unverifiable precondition remains `blocked_evidence`.

The names, email addresses, account identifiers, and the mapping from any authorization record to a
person belong only in the approved restricted location. A repository-visible document may use only a
non-personal role/capability reference and a non-resolvable authorization-record reference.

## Evidence-Intake Register

| EC item | Current state | Required named parties before access | Source state | Allowed next action |
| --- | --- | --- | --- | --- |
| EC-01 SR-11 deletion/cancellation authority | `blocked_evidence` | Security/identity source owner; independent security verifier | Not designated or inspected | Submit minimized authority-abuse-test request. |
| EC-02 SR-01/SR-08 lineage and minimum audit evidence | `blocked_evidence` | Privacy/data source owner; independent privacy verifier | Not designated or inspected | Submit field-level lineage and redacted-audit request. |
| EC-03 SR-02 custody/destruction proof | `blocked_evidence` | Custody/operations source owner; independent custody verifier | Provider and custody source not designated | Submit provider-neutral capability request only. |
| EC-04 SR-03/SR-04 marker integrity/replay | `blocked_evidence` | Control-plane source owner; independent security verifier | Marker source not designated or inspected | Submit restricted marker-design/evidence request. |
| EC-05 SR-05/SR-12 lifecycle data state/runbook | `blocked_evidence` | Architecture/data source owner; independent architecture and security verifiers | No lifecycle system or runbook authorized | Submit non-authorizing design/runbook request. |
| EC-06 SR-06 restore/release | `blocked_evidence` | Restore/operations source owner; independent restore-release verifier | Restore source not designated or inspected | Submit isolated restore-control request. |
| EC-07 SR-07 managed copies | `blocked_evidence` | Backup/operations source owner; independent reconciliation verifier | Backup/PITR/export inventory not obtained | Submit managed-copy inventory request. |
| EC-08 SR-09 operations/insider boundary | `blocked_evidence` | Operations/security source owner; independent security verifier | Operations source not designated or inspected | Submit role/access-control request. |
| EC-09 SR-10 legal/retention boundary | `blocked_evidence` | Legal/privacy/security decision owner; independent legal and security verifiers | Legal/regulatory source not obtained | Submit restricted policy-decision request. |
| EC-10 formal Security Review record | `blocked_evidence` | Security Review coordinator; Principal Architect decision owner; independent human Security Review | No complete evidence package exists | Maintain clause-level register only; do not start review. |

The register is an intake ledger, not evidence. It must never be changed to pass, verified, or
reviewable merely because a request was submitted, an owner was invited, CI passed, a document was
approved, or an access path exists.

## Restricted Collection Protocol

1. The accountable source owner submits one request per evidence claim, with the preconditions and
   exact SR clauses listed above.
2. The privacy and security approvers approve minimization, handling, and restricted location before
   any source access. They reject instead of broadening a request that cannot be safely minimized.
3. The source owner grants only the approved, time-bounded access to the designated collector. No
   shared, standing, break-glass, copied, or repository-stored credential is permitted.
4. The collector obtains only the approved evidence and stores raw or restricted material outside
   this repository in the approved location. The repository may contain only a non-sensitive intake
   receipt that does not function as an identity map, secret, raw payload, or marker lookup.
5. The independent verifier checks the clause-level claim, provenance, integrity, freshness,
   minimization, redaction, and limitations without editing the collected evidence.
6. The coordinator records the actual verifier result as pass, fail, or `blocked_evidence` for every
   clause only in the restricted evidence index. A repository-visible intake receipt records no
   verifier result. A failure, stale receipt, unavailable source, retention conflict, or disagreement
   blocks the related SR question and is escalated through the approved route.
7. A formal Security Review may begin only under the separate Stage 3.25 condition that every
   required clause is current, independently verified, and free of blockers.

## Intake Receipt Contract

Every repository-visible receipt must include only the following non-sensitive metadata:

| Field | Required content |
| --- | --- |
| Receipt label and baseline | Repository-local randomly generated, non-resolvable receipt label; frozen baseline; EC item; exact SR clause(s); claim; and scope. The label must not be copied from, resolve to, or support status lookup in a source, ticket, authorization, or evidence system. |
| Authority and separation | Non-personal role/capability references for source owner, access approver, collector, and independent verifier; a non-resolvable restricted authorization-record reference; conflict declaration; and expiry. |
| Source and method | Source-system class, environment/configuration scope, collection time, minimized method, and restricted-location reference without a secret or lookup capability. |
| Handling | Classification, permitted fields, redaction/restricted-access method, retention/expiry, and destruction responsibility. |
| Integrity and freshness | Integrity/provenance method identifier, collection time, freshness rule, and receipt expiry only. |
| Receipt classification | Fixed literal `intake_metadata_only`; no outcome, limitation, failure detail, escalation identifier, status reference, or implementation authorization. |

The receipt must not include raw identity data, names, email addresses, account identifiers,
credentials, tokens, request/import bodies, provider exports, backup contents, key material,
reusable signatures, direct source URLs, usable identifier maps, or public marker lookup material.
A receipt that cannot satisfy this contract is itself `blocked_evidence` and must use the separately
approved restricted-access path without copying the sensitive material into Git.

Verification times, check results, pass/fail/blocked outcomes, verifier decisions, and failure
details belong only in the restricted evidence index. They must never appear in a repository-visible
receipt or be encoded in its label, metadata, status, file name, path, or commit message.

A request, repository-visible receipt, or CI result is intake/process metadata only. It does not
communicate intake success/failure, `blocked_evidence`, a limitation, or an escalation. It does not
prove the existence or effectiveness of a control, close an SR clause, change an EC item to verified
or reviewable, or replace independently verified restricted evidence.

## Safe Failure and Scope Boundary

No evidence is received by this stage. No source has been inspected, no access has been granted, and
no organizational owner or independent verifier has yet been recorded. Therefore every EC item
remains `blocked_evidence`, no formal Security Review may start, ADR-008 remains proposed, and no
privacy-lifecycle implementation is authorized.

This stage adds documentation only. It does not access production or providers; collect data; inspect
backups; run restores; create credentials; change Go, Next.js, Python, OpenAPI, PostgreSQL, Redis,
migrations, CI, infrastructure, operations, or product behavior. The immediate next action is
mandatory strict review of this intake boundary. A real evidence request can be considered only after
the required named parties, source-owner approval, restricted storage, and minimization plan exist.
