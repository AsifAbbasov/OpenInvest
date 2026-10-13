# AUTH-ADV-02 email ownership remediation candidate V1

## Authority and scope

Implementation baseline: protected develop `8d9998590fba6c605a64ddcc8429f51df82b56c3`,
tree `bed3b0b3466de160c6bb954b41b373e5ed7ea1ba`, ten required CI contexts.
The original evidence remains unchanged in PR #249 and the independently tested
strengthened branch at `e02c6d28721277c74df486476de0dcb8c1e6473c`, tree
`8946e083c0ddec22c76b2826af5f53c06371be8d` (normal run 38389674999 10/10;
specialized run 38389674988 6/6). Those runs are not remediation-head evidence.

Conditional authorization permits isolated implementation only. ADR-010 remains
Proposed. Architecture/contract/security acceptance, merge, deployment and external
email activation are not authorized. The candidate is not a finding-closure claim.

## Implementation boundaries

- `auth/email_ownership.go`: email-only intake; 256-bit crypto/rand secrets;
  canonical Base64URL encoding; SHA-256 digest; owner-selected password completion;
  independent bounded database/delivery contexts; sanitized delivery failures.
- `postgres/email_ownership_store.go`: shared capacity of 10000 pending intents;
  database-clock 24-hour intent and 30-minute token expiry; intent generation;
  parent-first row locking; single-use activation; no session insertion; one audit
  success event; existing user conflict refuses completion without replacing data.
- Migration 000015 adds registration intents, token digests, proof records and
  leased delivery jobs. Unique indexes establish row/email/token identity and
  foreign keys bind related rows. No old migration or legacy user is rewritten.
  Timestamp/state cross-field rules are enforced by transactional store operations;
  this candidate does not claim additional unsupported paired-SQL CHECK grammar.
- Bounded delivery: at most four attempts per intent generation, one-minute resend
  suppression, five-minute retry after failed delivery, one-minute lease. Each
  ambiguous retry issues another secret, retaining prior valid tokens until expiry.
  Tokens are never persisted as plaintext/ciphertext. Failed delivery cannot activate.
- `verificationmail/smtp.go`: standard-library SMTP adapter with mandatory STARTTLS,
  TLS 1.2 minimum, finite deadlines, configured HTTPS completion URL and sanitized
  errors. External delivery is disabled unless the separate worker is explicitly
  started with `-enable-external-delivery`. No real SMTP was called in testing.
- Public register/resend: generic 202, no cookie/session or identity identifier.
  Verify: POST only, ACCOUNT_READY, no automatic login. Existing login then creates
  ordinary session roots; refresh/CSRF/session-family semantics are unchanged.
- OpenAPI operations, request/response examples, validator route/security inventory
  and password/timezone contract tests follow the moved credential-creation boundary.
- Frontend never authenticates on the registration acknowledgment. Verification
  consumes a fragment token in memory and removes it from the address bar; no
  localStorage or analytics persistence. Password/preferences are owner-selected.
  The completion page is presentation-only, no business Route Handler or DB access.
- Existing auth tests needing an account use an explicit legacy fixture followed by
  real HTTP login. They do not hide a runtime fallback. Cookie/CSRF/replay assertions
  remain. CORS cardinality rejection assertions are unchanged; normal cookie
  compatibility now exercises login because registration deliberately issues none.

## Runtime/database permissions and retention

Only new identity relations receive the privileges necessary for intents, leasing,
consumption and bounded cleanup. Existing users/credentials remain SELECT/INSERT;
immutable ledger and append-only audit grants are unchanged. Runtime deletion is
limited to the new parent intent relation; cascading expiry cleanup deletes at most
100 expired intents older than seven days per intake transaction. The intake shared
advisory lock coordinates capacity across processes. Cleanup is not expiry authority.
No automatic proof backfill exists: missing proof rows represent legacy unverified
email evidence. Existing account states, sessions, passwords and financial links
are unchanged. Possession of a new verification secret never transfers an existing
account's financial records. Legacy recovery/revocation decisions remain approval
blockers; this candidate implements no anonymous password-reset operation.

## Executed local evidence

- Normal Go package suite: passed before the final test additions; PostgreSQL/Redis
  environment-dependent cases skipped locally and are not credited as service proof.
- Focused race ownership service/HTTP tests: PASS (four top-level tests), including
  no session on activation, normal login afterward and unavailable-backend refusal.
- OpenAPI validator: PASS, 32 operations, 4158 resolved references, 19 documents.
- Migration validator: PASS, 15 pairs; paired-SQL metadata/fingerprints preserved.
- Frontend typecheck and production build: PASS for the implemented verification
  route; existing frontend suite PASS and two new request/response tests PASS.
- Final-head normal CI and full local race results are reported externally after
  completion. No old run is substituted. Local PostgreSQL executable/service was
  unavailable; new integration cases must execute in normal CI's PostgreSQL 18.

PostgreSQL integration coverage includes no identity before proof, digest-only token
storage, activation without session, subsequent login, consumed-token refusal,
concurrent single-winner completion and database-controlled expiry refusal.

## Explicit acceptance gaps

The existing Module 1 specialized adversarial campaign was not imported, weakened,
or executed by this executor. Its required remediation-head result is NOT_VERIFIED.
The full requested delivery fault/crash matrix, token substitution/purpose matrix,
legacy takeover matrix, bounded timing classification, mounted verification-screen
state matrix, secret-sentinel logging campaign and production worker telemetry are
not established by the limited ordinary regression suite. No partial result is
promoted to complete adversarial proof. The newly added registration contract needs
independent gatekeeping and the acceptance gaps must be resolved before closure.

Production email configuration/deliverability, sender-domain setup, edge behavior,
external log aggregation and legacy recovery policy remain unverified/unapproved.
The four-attempt-per-intent delivery budget is conservative and may delay recovery
when messages fail or mailbox traffic is unwanted; product/abuse policy review is
required. No constant-time response claim is made.

## Stop gate

```
AUTH_ADV_02_BEFORE=CONFIRMED_P3
AUTH_ADV_02_AFTER=IMPLEMENTED_CANDIDATE_FULL_REVALIDATION_PENDING
ARCHITECTURE_DECISION_STATUS=PROPOSED
PRODUCTION_EMAIL_DELIVERY_CONFIGURED=NO
SPECIALIZED_MODULE1_CI=NOT_EXECUTED_EXECUTOR_POLICY
AUTH_ADV_02_REMEDIATED_AND_REVALIDATED=NOT_YET_PROVEN
PRODUCTION_DEPLOYMENT_PERFORMED=NO
MERGE_PERFORMED=NO
MODULE_1_CLOSED=NO
FIVE_HOUR_CAMPAIGN_STARTED=NO
FINAL_REPOSITORY_WIDE_ASSAULT_STARTED=NO
IMPLEMENTATION_READY_FOR_GATEKEEPER_ACCEPTANCE=NO_FULL_EVIDENCE_PENDING
STOP_FOR_INDEPENDENT_GATEKEEPER_REVIEW=YES
```

## First exact-head CI failure and justified correction

Run 38411406391 on e93e3174ffa19573bc8f994c381369e6431ba8c0 passed eight
jobs, including migration validation, but Go tests/race failed. Complete job logs
115344943310 and 115344943148 identify an implementation integration defect:
OpenRuntime correctly rejected grants on the four new unknown identity relations.
The correction adds exactly their reviewed SELECT/INSERT/UPDATE/DELETE envelopes
to runtime_privileges.go, retaining rejection of unknown relations, dangerous
capabilities, ownership paths, DDL rights and audit/ledger mutations. A real runtime
role activation test was added. This is not a relaxed oracle or suppressed failure.
Fresh exact-head normal CI is required after this correction. Deployment must
coordinate new code/schema/runtime grants; old startup guards cannot accept newly
introduced capabilities. No rolling-deployment compatibility is claimed here.
