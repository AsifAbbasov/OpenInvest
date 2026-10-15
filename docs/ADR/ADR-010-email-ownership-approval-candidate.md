# ADR-010: Email ownership before new account activation

Current owner disposition: [approved closeout decisions](../audit/MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md). Earlier proposed/pending statements below record the pre-approval package and are superseded only as to these three decisions; historical evidence and residuals remain unchanged.

| Field | Value |
| --- | --- |
| Document ID | ADR-010 |
| Version | 1.1.0 approval candidate |
| Status | Approved by project owner, 2026-10-15; original pinned ADR unchanged |
| Supersedes | Registration/session issuance portion of ADR-006 only after explicit architecture approval |
| Dependencies | ADR-006; ADR-007; OpenAPI; ACCESS_JWT_IDENTITY_POLICY_V1 |
| Implementation snapshot | PR #266, e982d5a68e083a9e65b7f80b902f746b1b408bcb |
| Authority | Documentation preparation authorized; architecture acceptance, merge and deployment are separate gates |
| Historical authority | Original ADR010 file remains byte-identical because migration 000015 pins its hash |

## Historical authority binding

This is a prospective revision/addendum candidate, not an in-place replacement of
[the original ADR010](ADR-010-email-ownership-before-account-activation.md). Migration
000015 pins that historical document hash in policy_manifest.json. The original
file, migration metadata and validator remain untouched. Owner approval of this
candidate may establish the later architecture decision without rewriting migration
evidence; any future rebinding/replacement requires its own governed change.

## Context and scope

Historical AUTH-ADV-02 covered registration membership disclosure, pre-registration
identity claiming and missing email ownership evidence. The implementation candidate
replaces public immediate account/session issuance with email proof before activation.
ADR-006 remains Accepted. This Proposed revision describes the actual candidate and
its intended narrow supersession; it does not rewrite the historical closure report.
The owner-approved [JWT identity contract](../api/ACCESS_JWT_IDENTITY_POLICY_V1.md)
is a separate approved rule, not implicit approval of this ADR.

## Decision proposed for owner approval

1. Anonymous registration/resend creates or schedules only a registration intent.
   It creates no user, password, financial subject or authenticated session.
2. The public request accepts email only. A valid eligible/new/existing/pending
   address receives the same generic acknowledgment, not membership disclosure.
   No constant-time or email-deliverability guarantee is asserted.
3. Only a valid single-use proof bound to intent, generation and purpose can create
   a new account. The address comes from the locked intent, not completion input.
4. Completion atomically creates the user, credentials, financial subject/link,
   privacy defaults, verified-email evidence and success audit event; consumes
   outstanding intent proofs and cancels delivery. It creates no session.
5. ACCOUNT_READY means activation, not authentication. The owner subsequently logs
   in through the existing login/session flow. Refresh rotation and CSRF remain.
6. Existing non-deleted accounts cannot be reclaimed through this registration
   flow. Email possession alone cannot replace their password, sessions, financial
   identity or records. Existing verified ownership and legacy ownership remain.
7. Legacy accounts are not backfilled as verified, disabled or migrated silently.
   Missing proof evidence means legacy-unverified. Legacy login continuity remains;
   anonymous recovery, email changes, financial transfer and bearer revocation need
   separate policy authorization. No password-reset endpoint is introduced.
8. Bounded workers use expiring fenced leases and at most four attempts per intent
   generation. New generation renewal cannot be stranded behind an old lease.
9. SMTP is disabled by default and requires separate deployment approval. Neither
   accepting this ADR nor merging code enables a sender, relay or paid provider.
10. Financial authorization remains the validated signed investmentSubjectId.
    This ADR does not collapse user identity and financial identity.

## Actual API and frontend contract

| Operation | Request / response | Identity/session effect |
| --- | --- | --- |
| POST /api/v1/auth/register | Email only; 202 REQUEST_ACCEPTED with generic message | Intent/scheduling only; no cookie/session |
| POST /api/v1/auth/email/resend | Same email-only intake; same 202 | Bounded scheduling; does not reset an active generation's budget |
| POST /api/v1/auth/email/verify | Token plus owner-selected password/language/theme/timezone; 200 ACCOUNT_READY | New account only; no automatic login/cookie |
| POST /api/v1/auth/login | Existing credential contract | Ordinary access/refresh/CSRF session |
| POST /api/v1/auth/refresh and /logout | Existing cookie + session-bound CSRF contract | Existing rotation/revocation behavior |

Malformed inputs, capacity/rate gates and unavailable dependencies retain the
OpenAPI 400/429/503 contracts. Generic acknowledgment is not proof of delivery.
The Next.js registration form submits no password and does not authenticate after
202. The verification page captures a fragment proof in memory, strips it while
preserving query/history state, fences superseded responses, and presents normal
login after success. Mounted polling for silent history changes is 250 ms; submit
and response handling also check navigation synchronously. No token is persisted
to localStorage/sessionStorage. This is presentation code under ADR-007, not a
second business API. No universal deployed analytics/logging guarantee is claimed.

Normative candidate artifacts: [OpenAPI](../../openapi/openapi.yaml),
[schemas](../../openapi/components/schemas.yaml),
[frontend verification](../../frontend-next/src/app/auth/verify/page.tsx).

## Proof, transaction and worker semantics

- Secrets: crypto/rand 32 bytes, canonical unpadded Base64URL; PostgreSQL stores
  only SHA-256 digests. Plaintext exists transiently in application/transport memory.
  Purpose is new_registration; no recovery purpose is enabled.
- PostgreSQL clock: intent lifetime 24 hours; proof expiry is the earlier of
  30 minutes and intent expiry. Consumption/generation checks are authoritative.
- Parent intent locks precede jobs/tokens. Shared intake advisory locking and
  a maximum 10000 active pending intents bound repository-side scheduling.
- Repeating unexpired intake does not invalidate still-valid proofs. Renewing an
  expired unconsumed intent increments generation and atomically replaces old job
  ownership, clearing lease fields and resetting attempts only for a new generation.
- Attempts increment before delivery; four per generation, one-minute lease,
  one-minute next-delivery suppression after leasing, five-minute failed retry.
  A worker pass can terminalize one exhausted expired lease without a fifth token.
  Stale/expired lease callbacks are no-ops. Database fencing cannot forcibly stop
  an already-started external SMTP operation; at-least-once delivery is intentional.
- Exhausted same-generation delivery is not restarted by resend; renewal follows
  expired-intent generation rules. A mailbox owner may have to wait for renewal.
- Old unconsumed proofs can remain valid until generation change, expiry or intent
  consumption. This is not exactly-once email delivery or stored plaintext replay.
- Database store transactions set lock_timeout=2s and statement_timeout=4s;
  intake/completion database contexts are bounded to 5s. Password hashing precedes
  the completion DB context and retains existing bounded Argon2 admission.
- Delivery context is 30s, finalization has a detached bounded 5s context. SMTP
  requires STARTTLS (TLS >=1.2), finite connection deadlines and an HTTPS proof URL.
- SQL transactions are atomic. A lost COMMIT acknowledgment does not prove rollback;
  an account may have committed despite a client error. No anonymous account recovery
  fallback is inferred from such an error. Dedicated historical ACK tests are evidence
  for tested operations, not every possible network/database failure.

Source: [ownership service](../../backend-go/internal/auth/email_ownership.go),
[transactional store](../../backend-go/internal/postgres/email_ownership_store.go),
[worker](../../backend-go/cmd/verification-mail-worker/main.go),
[SMTP adapter](../../backend-go/internal/verificationmail/smtp.go).

## Database permissions and privacy lifecycle

Migration 000015 is additive: intents, proof digests, verification evidence and
leased jobs. Runtime capability envelopes are explicit; new relation UPDATE rights
are table-level where defined, not claimed column-only. Existing ledger/audit grants
are unchanged. Parent intent deletion cascades only its new dependent rows.
Intake cleanup deletes at most 100 intents expired more than seven days ago per
transaction. This is activity-driven cleanup, not a guaranteed idle-system retention
SLA. Deployment must approve operational retention/monitoring; no new cleanup worker
or production log aggregation is assumed.

## Compatibility, integration and rollback

Public registration changes from immediate issuance to email-only 202; credential
selection moves to completion. Frontend/API/OpenAPI/schema/runtime-grant rollout must
be coordinated. Old startup privilege guards reject newly introduced capabilities;
rolling binary compatibility and zero-downtime migration are not established.
The exported internal Service.Register remains for legacy test fixtures; reviewed
public handlers do not call it. Future callers must not reintroduce public fallback.

Rollback must close new registration/verification/resend admission rather than restore
unverified public activation. There is no dedicated registration-disable flag claimed
by this candidate: an approved maintenance/routing procedure is required. Retain
additive schema/evidence after activation; do not blindly run destructive DOWN or
roll back to an old privilege guard. Prefer a reviewed forward correction. Preserve
established accounts, sessions and financial records; never compensate by transferring
ownership. No rollback, routing, schema or production change is performed here.

## Evidence, alternatives and costs

The [readiness matrix](../audit/MODULE1_ARCHITECTURE_ACCEPTANCE_AND_INTEGRATION_READINESS.md)
links original findings, targeted fixes, real PostgreSQL/browser evidence and bounded
limitations. Alternatives rejected for this candidate: immediate unverified activation;
raw/ciphertext proof persistence; automatic login on activation; email-only legacy
password/financial takeover; unbounded delivery retries. Operational costs are delayed
registration when mail is unavailable, pending PII/queue retention, relay/TLS monitoring,
and coordinated release. External delivery and production edge remain unverified.

## Approval outcome — not yet granted

ADR010_FORMAL_STATUS=PROPOSED. Proposed owner decision: approve this revision's
email-proof-before-new-account contract, API change, legacy continuity and explicit
operational boundaries; leave SMTP activation, recovery redesign and new revocation
semantics separately gated. Architecture approval does not authorize merge/deploy.
The precise approval text and remaining gates are in the readiness report.
