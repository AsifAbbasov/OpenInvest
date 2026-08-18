# Stage 3.21 - Privacy Data Inventory Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-21-PRIVACY-DATA-INVENTORY-PROPOSAL |
| Version | 0.1.2 |
| Status | Complete / merged through PR #50 at `207325e0497cc2608b99366f7f840472d270b6ed` |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.20 privacy threat-model proposal |
| Dependencies | Documents 42-43; ADR-005; ADR-006; proposed ADR-008; Stage 2 ER model and migration strategy; Stage 3.17-3.20 privacy proposals |
| Last Review Date | 2026-08-18 |
| Next Review Date | Historical proposal closed; successor Stage 3.22 |

## Purpose

This repository-derived inventory is the next required privacy-lifecycle evidence package. It maps
observed durable fields, code/browser surfaces, and unverified external surfaces to a future
deletion or retention disposition. It makes every known direct identifier, reversible link,
correlation field, free-text field, and opaque payload visible before anyone claims that retained
financial history is anonymous.

This is not a production data discovery, Security Review, legal assessment, acceptance of ADR-008,
or authorization to delete, anonymize, migrate, encrypt, back up, restore, or change runtime
behavior. The repository cannot prove what a deployed platform, operator, browser extension,
support system, observability vendor, replica, backup, or export currently retains.

## Authority and Method

Documents 42-43 require complete deletion of identity data, irreversible destruction of the
person-to-financial link, Anonymous Financial History, no reasonable technical or organizational
reidentification path, ten-year audit retention, and encrypted-backup destruction within 90 days.
Those requirements take priority over this proposal. A row state called deleted, a foreign-key
cascade, an opaque UUID, a hash, or encrypted media is not by itself proof of this outcome.

This inventory has three evidence states:

| Evidence state | Meaning |
| --- | --- |
| Observed in repository | Current migration or source code exposes the field or surface. It is not proof of production deployment or retention. |
| Logical-only | A Stage 2 model or future proposal mentions it, but no current migration establishes it. |
| External evidence required | The repository has no sufficient evidence. Absence from source code is not proof of absence in an operated environment. |

The classifications used below are intentionally conservative:

| Classification | Meaning |
| --- | --- |
| Direct personal | Directly identifies or contacts a person. |
| Secret or authentication material | Raw credential/token or a durable verifier that can authorize or assist an attack. |
| Direct or reversible bridge | Connects a person to a subject, portfolio, session, or financial history. |
| Correlated or indirect | Can link records, devices, requests, imports, or behavior and may enable reidentification when combined with other material. |
| Free-form or opaque | Content is not constrained enough to classify safely. |
| Reference data | Shared product/reference information; verify that it is not tenant- or person-specific before retaining. |

The future dispositions are requirements for a later approved design, not current behavior:

| Disposition | Required result |
| --- | --- |
| Delete | Remove the record and all derived copies after the approved lifecycle. |
| Revoke then delete | Revoke its ability to authorize first, then remove durable material. |
| Sever or transform, then retain | Preserve only after the person/link/reidentification risk is irreversibly removed and proven. |
| Retain only after proof | A retention duty may apply, but a field-level non-reidentification and access-control proof is still required. |
| Unknown blocks completion | No anonymity, backup, restore, or deletion-completion claim is allowed until disposition and evidence exist. |

Repository evidence examined for this proposal includes the repository migration chain/schema-as-code
000001_stage_03_01_vertical_slice, 000002_stage_03_11_auth_privacy, and
000003_stage_03_16_transaction_source_provenance; the PostgreSQL stores; Go HTTP API; and Next.js
auth/import code. Application of this schema to an actual production database is external evidence
required. This proposal does not inspect a production database, storage account, platform logs,
provider configuration, or a human operator's local material.

## Observed PostgreSQL Inventory

### Logical-only future surfaces

The Stage 2 logical model and the Stage 3.17-3.20 proposals mention a deletion request/control
plane, durable deletion marker, per-subject erasure-key hierarchy, and restore-reconciliation
evidence. None of these is established by the repository migration chain. They are logical-only
future surfaces, not current records that can be inventoried, relied on, or claimed as controls.
Any later design must give each field, key reference, marker, audit record, queue message, and
backup projection an explicit disposition before implementation.

### Identity and access

| Surface and fields | Classification and risk | Future disposition | Required evidence |
| --- | --- | --- | --- |
| identity.users: id, email_normalized, account/language/theme/timezone, timestamps, deletion timestamps | email_normalized is direct personal data. id is a direct bridge used by current joins and audit insertion. Preferences and timestamps can be indirect identifiers. A deleted account state does not prove deletion. | Delete identity and direct fields; remove all person-linking uses of id. | Transactional deletion proof across dependent tables, post-delete query proof, and restore rehearsal. |
| identity.credentials: user_id, password_hash, timestamps | Authentication material plus a direct bridge. Argon2id storage is a security control, not anonymous data. | Revoke then delete. | Revocation and deletion proof in primary and restored copies. |
| identity.privacy_settings: user_id, privacy/tax/notification/analytics fields, timestamps | Direct bridge plus behavior/preferences that may be personal. | Delete. | Cascade/worker proof and restored-copy search. |
| identity.sessions: id, user_id, refresh/CSRF hashes, state, expiry, timestamps | Authentication material plus user/session correlation. Raw refresh and CSRF tokens are not stored here, but hashes remain sensitive durable verifiers. | Revoke then delete. | All-session invalidation, primary/backup evidence, and replay-negative tests. |
| identity.user_investment_links: user_id, investment_subject_id, created_at | The live, plain reversible identity-to-financial bridge. | Delete or irreversibly sever before retaining financial history. | No residual relationship in primary, replica, restore, export, audit, or operator material. |
| investment.subjects: id, state, anonymous timestamp, timestamps | Opaque subject ID is still a bridge while the link or correlated material exists. The anonymous state alone is not proof. | Sever or transform, then retain only after proof. | Linkage analysis covering every referencing table and external copy. |

### Financial, import, and derived records

| Surface and fields | Classification and risk | Future disposition | Required evidence |
| --- | --- | --- | --- |
| investment.portfolios: id, subject_id, name, state/version/timestamps | subject_id is an indirect bridge. User-entered name is free-form and can contain a person or account reference. | Sever/transform link; classify and remove/transform name where necessary; retain only after proof. | Field-level content policy, sampled/redacted discovery method, and no-link proof. |
| investment.transaction_entries: entry/transaction/portfolio/asset IDs, dates, amounts, revision/reversal links | Financial history is indirectly identifying through portfolio_id, transaction graph, dates, and unique holdings. It cannot be declared anonymous merely because user_id is absent. | Sever or transform bridges, then retain only after proof. | Reidentification analysis against all retained tables, exports, and external systems. |
| investment.transaction_entries.note, correction_reason | Free-form user or operator text; current SQL constraints do not prevent personal data. | Unknown blocks completion until content rule and removal/transform design are proved. | Field inventory, bounded input policy or classification/remediation evidence, and restore test. |
| investment.transaction_entries.request_id, trace_id, source_kind, source_file_hash | Request/trace values and a user-uploaded-file hash are correlation/provenance material. A SHA-256 file hash can link a retained ledger entry to another copy of a file. | Unknown blocks completion unless transformed or shown non-reidentifying. | Cross-system trace/log/import inventory and linkage analysis. |
| investment.command_deduplication: principal_id, method/path, idempotency key, request/response hashes, timestamps/expiry | Direct principal bridge and request correlation. expires_at is not a demonstrated deletion/backup lifecycle. | Delete after authorization/retry window; do not retain as anonymous history. | Cleanup execution, replicas/backups, and idempotency replay evidence. |
| investment.outbox_events: aggregate IDs, payload JSONB, error/timestamps | Aggregate IDs can bridge to financial data; unbounded JSONB and error content are opaque/free-form. | Unknown blocks completion until every event version has an approved payload field inventory and lifecycle. | Event schema registry/allowlist, consumer inventory, dead-letter/export evidence. |
| analytics.portfolio_snapshots, snapshot_positions, calculation_runs | Portfolio/snapshot IDs, values, dates, watermarks, and positions remain indirectly identifying while any bridge persists. | Sever or transform, then retain only after proof. | Derived-data rebuild/deletion design and cross-store linkage analysis. |
| analytics.inbox_messages: event ID, consumer, business version, status/error/timestamps | Event correlation and potentially operator/error information. It has no payload in the observed schema, but can link to outbox/consumer copies. | Delete or transform according to the event inventory; unknown blocks completion. | Consumer/dead-letter/provider inventory and purge proof. |
| investment.assets: ticker/name/ISIN/market/lot size | Candidate shared reference data, not per-person data in the observed migration. | Retain only after proof that sources and use remain non-personal. | Catalog provenance and tenant-separation check. |

### Audit records

| Surface and fields | Classification and risk | Future disposition | Required evidence |
| --- | --- | --- | --- |
| audit.actors: id, actor_kind, timestamp | Current auth code inserts the user UUID as a user actor ID. It is therefore a direct stable identity key despite no SQL foreign key to identity.users. | Transform/sever before ten-year retention, or retain only after proof. | Field-level actor replacement design, privileged-access controls, and primary/restore linkage-negative tests. |
| audit.events: actor_id, target kind/ID, action/outcome, request/trace IDs, timestamps | actor_id, target IDs, and request/trace values can retain identity, session, or financial correlation. Ten-year retention does not make them anonymous. | Retain only after proof; remove or transform identifying/correlating fields while preserving the minimal audit purpose. | Audit schema contract, retention/minimization review, restored-copy search, and access evidence. |

## Code, Browser, and Transient Surfaces

| Surface | Observed repository fact | Classification and required disposition |
| --- | --- | --- |
| Registration/login requests | Email and password pass through the Go API request path. The repository establishes database hashes, not a complete platform request-log policy. | Direct personal/secret while in transit; external logs are unknown and block completion until inventoried. |
| Browser auth state | The Web auth shell keeps the active access token and CSRF token in application memory; the refresh token is an HTTP-only cookie. No business-data localStorage or sessionStorage use was found in the inspected frontend source. | Authentication material. Revoke/expire and delete according to browser/session evidence; browser caches, extensions, crash reports, and device backups require external evidence. |
| Import-review request | The browser submits raw CSV and an optional user-supplied source-account label for review. No SQL import-session/raw-file table exists in the repository migration chain. | Raw CSV can contain financial records and free text; the label can identify an account or person. They are transient only in observed application code, not proven absent from browser memory, request/proxy logs, telemetry, caches, support tools, or backups. |
| Import-review response | The API returns portfolio ID, source kind/label/file hash, review token, row number/hash/fingerprint, reason codes, and candidate transaction data including optional safeNote. | Direct/indirect identifiers, financial data, hashes, and optional free text are held in a browser response/state. Delete from client/server transient handling after the approved operation; unknown external observability/caching copies block completion. |
| Signed import-review token | The token is base64url-encoded JSON plus HMAC. Its payload contains subject ID, portfolio ID, source kind/label, file hash, and row number/hash pairs. HMAC provides integrity, not confidentiality; the observed payload has no expiry field. | Correlation material and direct/indirect bridges. Do not treat it as opaque or anonymous. Require a bounded issuance/revocation/expiry lifecycle and evidence for server memory, browser, telemetry, proxy, cache, and backup copies before deletion completion. |
| Browser import state | While the import component is mounted, React state holds filename, raw CSV, reviewed CSV, source label, selected row numbers, review response/token, and append result. The observed code clears key fields after scope change and successful append, but no independent browser-memory lifecycle proof exists. | Filename/label, raw/reviewed CSV, response candidates, fingerprints, safeNote, hashes, and token are personal, financial, free-form, or correlated material. Browser caches, extensions, crash reports, and device backups require external evidence. |
| Application logging/errors | Source inspection found startup error logging and no complete request/field-redaction system. This cannot prove a deployment has no request, SQL, proxy, or tracing logs. | Unknown blocks completion until each logging/telemetry sink, retention period, and redaction rule is inventoried. |
| Test data, fixtures, CI artifacts | Repository tests use sample identity/auth/import values. Their retention and access outside the working tree are not established here. | Treat as non-production examples only; require CI artifact/cache and secret/PII handling evidence before completion. |

## External and Out-of-Repository Inventory

The following surfaces cannot be marked absent or compliant from repository inspection. Each has an
accountable future owner and must produce evidence before a deletion can be marked complete.

| Surface | Current evidence state | Required owner and evidence |
| --- | --- | --- |
| PostgreSQL physical backups, PITR/WAL, replicas, snapshots, restore media | External evidence required. No provider-specific backup lifecycle is configured here. | Platform/DB owner: inventory, encryption/key custody, <=90-day destruction proof, restore rehearsal, and deletion-marker reconciliation. |
| Redis persistence | Local docker-compose declares Redis with AOF storage; inspected app source does not establish a production Redis persistence integration. | Platform/application owner: deployed topology, keyspace, TTL, persistence/backup/export inventory, and purge proof. |
| Queue, outbox consumers, dead-letter stores, webhooks | Schema supports outbox/inbox, but provider/consumer durable copies are not proved by this repository. | Event owner: payload allowlist, consumer register, DLQ/retry/export retention, and deletion propagation proof. |
| Reverse proxy, CDN, load balancer, API gateway, WAF, observability, APM, error reporting | External evidence required. | Platform/security owner: request/header/body trace policy, redaction, retention, access, exports, and deletion/expiry evidence. |
| Browser/device backups, extensions, crash reporting | External evidence required. | Product/security owner: client-data policy, token/cookie behavior, and user/device residual-risk assessment. |
| Support tickets, operator runbooks, incident evidence, ad hoc exports | External evidence required. | Operations/privacy owner: data-handling policy, access register, redaction/deletion process, and audit trail. |
| CI caches, test reports, artifact storage, source mirrors | External evidence required. | Engineering/CI owner: retention, access, secret/PII handling, and purge evidence. |

## Completion Blockers and Required Evidence

The following observations are blockers, not implementation tasks for this stage:

1. audit.actors.id currently receives the user UUID for authentication events; a ten-year audit
   retention design must remove or transform this durable direct identity key without destroying the
   minimal evidentiary purpose.
2. Portfolio names, transaction notes, correction reasons, outbox JSONB, errors, and operator-facing
   material do not have enough field constraints to classify them as anonymous.
3. Request IDs, trace IDs, idempotency records, signed import-review tokens, imported-file hashes,
   filenames, source labels, raw/reviewed CSV, review responses, subject/portfolio graphs, and event
   IDs can bridge records even after the visible user row is removed.
4. The repository has no evidence for deployed backup/PITR/replica, provider queue, proxy/log,
   observability, CI artifact, export, or support-system behavior.
5. No current migration provides the future deletion-marker, per-subject erasure-key, restore gate,
   or cross-system completion reconciliation required by the earlier proposals.

A later design may mark a deletion complete only after it has a signed, redacted evidence record that
maps every field and external surface above to a disposition; proves direct and indirect links are
unavailable in primary and restored copies; proves revocation before deletion; and records the
accountable owners for delayed backup or provider destruction. It must fail closed when an unknown
surface cannot be checked.

## Scope Boundary

This stage adds no API, OpenAPI, schema, migration, retention job, queue, worker, encryption key,
provider, backup configuration, restoration procedure, test fixture, or runtime behavior. It does
not approve a data-processing inventory as legally complete, authorize production discovery, or
change the accepted requirement to retain only truly anonymous financial history and necessary
minimized audit evidence.

## Proposal Acceptance Criteria

- The document distinguishes observed repository facts, logical-only model claims, and unknown
  external surfaces.
- Every observed durable table/field group and code/browser surface has a conservative
  classification, future disposition, and evidence expectation.
- It explicitly identifies free-form, JSONB, hash, ID, trace, audit, and import correlation risks.
- It preserves the Document 43 audit, anonymous-financial-history, and backup requirements without
  claiming that the current system satisfies them.
- It authorizes no runtime, OpenAPI, schema, provider, key-management, backup, or operational work.

## Review Evidence

Internal-review evidence was withheld from PR #50 until the blind external reviewer reached an
independent conclusion. Publication of this section records review evidence only. It does not
accept ADR-008, constitute Security Review approval, authorize implementation, or establish
production compliance.

| Gate | Evidence | Verdict |
| --- | --- | --- |
| Internal review | Existing dedicated read-only internal-review task reviewed the complete pre-commit Stage 3.21 diff, governing privacy sources, migrations, Go API/import-token construction, React import state, governance registers, and `git diff --check`. It reviewed the corrective revision after its initial findings. Builder and public-CI verification are recorded separately. | `APPROVED` |
| External review | Dedicated blind external-review task independently reviewed published Draft PR #50, its complete 10-file diff, public CI, and governing sources without receiving the current internal verdict or findings before its conclusion. | `APPROVED` |

The review evidence is not operational proof. Field-level production discovery, key custody,
deletion-marker control, migration design, provider/backup evidence, operations runbook, and an
adversarial restore rehearsal remain future gates.

## Historical Closure

Stage 3.21 was squash-merged through PR #50 at `207325e0497cc2608b99366f7f840472d270b6ed` after
the recorded internal and blind external `APPROVED` verdicts. The merged proposal remains evidence
only: it does not accept ADR-008, constitute Security Review approval, or authorize implementation.

The successor Stage 3.22 separately addresses the provider-neutral key-custody and destruction-proof
evidence gap. The deletion-marker/control-plane, migration, operations, and restore-rehearsal gaps
remain separate future gates.
