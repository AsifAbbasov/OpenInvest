# ADR-010: Email ownership before new account activation

- Status: Proposed
- Supersedes: Registration issuance portion of ADR-006 only after independent architecture acceptance
- Authority: conditional implementation authorization; no merge/deployment authorization

Use registration intents separate from users. Anonymous email entry creates no user,
password, financial subject or session. A single-use 256-bit random email token, stored
only as a SHA-256 digest, permits its owner to choose a password and atomically create
an active account. Activation returns ACCOUNT_READY, never an authenticated session.
New/duplicate/pending requests use generic acknowledgments. SMTP is provider-neutral,
disabled absent explicit configuration; delivery failure never enables legacy registration.
No email-provider purchase or production activation is authorized.

Legacy users remain unchanged and unverified unless evidence is explicitly collected.
Email possession alone never replaces their password or grants their financial data.
Recovery and bearer revocation changes remain separate approval blockers.

The new contract requires coordinated frontend/API rollout. Pending PII and delivery
metadata expire and require bounded cleanup. Raw verification secrets never persist.
Token generation, purpose, expiry and consumption are bound transactionally.
Runtime permissions remain least privilege, immutable ledger/audit protections unchanged.

Merge requires Security Gatekeeper, architecture and contract approval. Delivery activation
requires an approved sender/relay configuration and operational verification. Rollback
must close registration, never restore immediate unverified activation.

## Implementation defaults and operational handoff

Token TTL 30 minutes; intent generation TTL 24 hours. At most four delivery attempts
per generation, one-minute resend suppression and five-minute failure retry. Queue
leases expire after one minute; attempts are counted before sending. Tokens are
issued into memory after digest persistence, so crash recovery may send another
bounded secret; old valid secrets remain usable until intent consumption/expiry.
Parent intent row locks precede jobs/tokens; database statements are hard bounded.
New identities require proof; legacy proof is absent and no legacy policy changes.

The standalone worker is disabled by default. Deployment must explicitly approve
SMTP host/port/from/user/password, HTTPS verification URL and database configuration.
No sender-domain, relay or paid provider is provisioned by this implementation.
This implementation ADR is not accepted or authority to enable real delivery.
