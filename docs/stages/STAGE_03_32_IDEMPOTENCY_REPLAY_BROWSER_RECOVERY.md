# Stage 3.32 — Exact Idempotency Replay and Browser Retry Recovery

Canonical record: PR #67; commit(s) `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746`, `0623d5ef326cd783b7dc0417dbcb02f18c506171`, `02aa2417a3caca79e2afc4e7b598b92055de96b7`, `57fcc25e949277a0e933f290998e41d0f7476b5c`.

## Purpose

Stage 3.32 addresses the gap between exactly-once business effect and exact observable HTTP replay.
The pre-stage implementation prevented many duplicate financial writes, but a duplicate command was
answered by rereading mutable database state rather than by returning the original HTTP result. The
browser also retained an unresolved idempotency key only in component memory, so reload/remount could
turn an ambiguous successful write into a new command.

The remediation keeps PostgreSQL as the authoritative transaction boundary and keeps Next.js
presentation-only. It introduces no Redis dependency, queue, worker, microservice, or product feature.

## P2-09 — completed commands did not preserve the original HTTP response

### Observed defect

`investment.command_deduplication` stored request identity and terminal state but did not store the
original HTTP response. Duplicate portfolio/transaction/import writes reread current resource state.
That protected business effects in common cases but violated the frozen API contract: an identical
replay must return the original status and body.

Because response metadata includes request ID, trace ID, and generation time, reconstructing a fresh
response could not be byte-for-byte identical even when the underlying financial object had not
changed.

### Remediation

Migration `000006_stage_03_32_idempotency_replay` adds a versioned exact response artifact to the
existing command-deduplication row:

- HTTP status;
- serialized response body bytes;
- original request ID;
- original trace ID;
- SHA-256 of the stored response body;
- artifact version.

The migration is additive and performs no data rewrite in the up direction. Legacy completed rows
remain distinguishable because their replay version is null. They fail closed rather than pretending
that a mutable resource reread is the original response.

For every currently implemented idempotent financial write, the transaction is now:

1. reserve or resolve the scoped idempotency command;
2. if already completed, return the stored artifact before consulting mutable business state;
3. for a new command, perform the financial write and derived/audit work;
4. construct the canonical HTTP response while the database transaction is still open;
5. persist the exact response artifact and its body hash;
6. commit the financial effect and replay artifact together.

If response serialization/artifact completion fails, the command reservation and financial write
roll back together. A ledger effect cannot commit while its exact successful response is absent.

### Mutable-state ordering

Exact replay is resolved before `ensureSubject` or portfolio locking. This is required because a
completed command must remain replayable even when the portfolio has subsequently changed state. A
new command still performs the normal subject/portfolio checks and locks inside the same transaction;
its reservation rolls back with any later failure.

### Concurrency and durability

The existing unique command scope remains principal + method + canonical path + idempotency key.
PostgreSQL conflict serialization ensures concurrent identical callers converge on one completed
command. Integration coverage proves that two concurrent identical portfolio creates produce one

Replay persistence is database-backed rather than process memory. A regression closes the first
Store connection completely, opens a new Store against the same PostgreSQL database, and verifies

### Import-review token expiry

Stage 3.30 intentionally made signed import-review tokens short-lived. A successful import can still
have an ambiguous client outcome after the token expires, while the idempotency command remains the
same completed financial operation.

Stage 3.32 preserves both properties:

- a fresh write still requires the normal signed review-token verification before financial append;
- only an otherwise fully authentic/context/semantic-valid token that failed because its lifetime
  elapsed may enter read-only completed-command recovery;
- recovery revalidates the token at its original issuance instant using the existing HMAC/context/
  parser-digest/row/decision verifier, without mutating the live request clock;
- signature, context, parser, row, or decision tampering cannot reach replay recovery;
- recovery never authorizes a new financial write; if no exact completed artifact is found, the
  original proof failure remains the response.

This is intentionally narrower than treating every invalid token as a replay candidate.

### Stored artifact integrity

A replay artifact is accepted only when its version, status, body size, request ID, trace ID, and
stored SHA-256 are structurally valid. Corrupt artifacts fail closed. The HTTP boundary sends the
stored bytes directly and restores the original request/trace response headers, so the response body
and technical identity are not regenerated from the retry request.

Canonical record: commit(s) `57fcc25e949277a0e933f290998e41d0f7476b5c`.
`02aa2417a3caca79e2afc4e7b598b92055de96b7` reconfirmed P2-09 CLOSED.

## P2-13 — browser retry identity was lost on reload/remount

### Observed defect

The browser kept the current idempotency intent only in React `useRef`. Same-mount retries reused the
key, but reload/remount lost it. After an ambiguous server success, the next submission could receive
a new UUID and become a second command.

### Initial remediation

The Web layer added a short-lived technical retry journal in `sessionStorage` containing only:

- version;
- opaque idempotency key;
- expiry timestamp.

The storage slot name is derived from SHA-256 of a technical scope. It does not persist the financial
payload, transaction amounts, CSV, source-account label, portfolio data, review token, access token,
or CSRF token. Within the same mounted interaction, a changed intent rotates to a new key. Across
reload/remount, an unresolved technical key is recovered and sent again. Confirmed success or a proven
idempotency conflict clears the applicable journal entry.

### Post-review remediation

The browser retry namespace now includes the stable authenticated `user.id` before hashing:

`principal + operation + optional portfolio scope → SHA-256 storage slot`

`AuthShell` exposes the already-present stable `AuthUser.id` to authenticated child presentation
components. The access token is deliberately not used as the retry owner because token refresh would
change it and break continuity. Portfolio creation, manual transaction append, and import append all
construct their retry scopes through the same principal-scoped helper.

The raw principal ID and raw portfolio/operation scope are never written to `sessionStorage`; only the
SHA-256-derived slot name and `{version, opaque idempotency key, expiresAt}` value persist. A logout
therefore does not destroy an unresolved key, User B uses a different slot, and if User A later signs
back in within the retry TTL the original User A key remains recoverable.

### Post-review regression

A dedicated regression models the exact review scenario:

1. User A persists an unresolved portfolio-create key;
2. User B signs in in the same logical browser-tab storage and receives a different key/slot;
3. User B clears the successful B slot;
4. User A returns and recovers the original unresolved A key;
5. serialized storage contains neither principal ID nor business payload text.

Existing reload, changed-intent, TTL, malformed-state, scope-minimization, and success-clearing
regressions remain active. Import component fixtures were also updated to supply the stable principal
without changing their stale-response/token-rotation semantics.

`02aa2417a3caca79e2afc4e7b598b92055de96b7` marked P2-13 CLOSED and reported no new blocking
P1/P2 regression.

## Regression evidence

The Stage 3.32 test set includes:

- exact portfolio HTTP status/body/request-ID/trace-ID replay;
- exact transaction replay without a second ledger entry;
- same key with a different canonical payload returns idempotency conflict;
- concurrent identical commands produce one business effect and one response artifact;
- exact replay survives a completely new PostgreSQL Store connection;
- completed transaction replay remains valid after the portfolio changes out of active state;
- completed import replays after its authentic signed review token expires;
- an expired token with a tampered signature cannot enter replay recovery;
- migration validator accepts the Stage 3.32 migration;
- migration apply, every-migration rollback rehearsal, and full reapply succeed on PostgreSQL 18;
- browser retry key survives reload/remount through session storage;
- changed same-mount intent rotates the key;
- expired browser retry state is discarded;
- browser storage exposes neither raw intent/payload, raw principal identity, nor raw technical scope;
- User A/User B browser retry journals remain independent in the same tab and clearing B does not clear A;
- successful writes clear the applicable principal-scoped retry journal;
- full frontend typecheck, tests, and production build remain green.

Exact final implementation head `02aa2417a3caca79e2afc4e7b598b92055de96b7` passed GitHub Actions CI #181
with all six jobs successful: Go tests with PostgreSQL/migrations, PostgreSQL migration validation,
frontend build/typecheck/tests, Python tests, OpenAPI contract validation, and Docker Compose
configuration validation.


than waived:


The first independent verdict was `changes required`; it remains historical evidence. The repeat
Canonical record: commit(s) `02aa2417a3caca79e2afc4e7b598b92055de96b7`.
`APPROVED`, with P2-09 and P2-13 both marked CLOSED and no new blocking regression. Explicit human
squash-merge authorization was then received, and PR #67 was squash-merged into `develop` at
`0623d5ef326cd783b7dc0417dbcb02f18c506171`.

## Privacy and retention boundary

The replay artifact can contain the same financial response data that was returned to the authenticated
user. Stage 3.21 already classifies `investment.command_deduplication` as a principal/correlation
surface whose future disposition is deletion after the authorization/retry window, with cleanup,
replica, backup, and replay evidence still required.

Stage 3.32 does not claim that privacy lifecycle is implemented or approved. It does not introduce a
cleanup worker, provider retention policy, backup purge, anonymization mechanism, or Stage 3.25
operational retry metadata; physical database/provider lifecycle remains a separate privacy track.
This residual boundary must not be cited as privacy closure.

The browser remediation stores no raw principal identifier. The stable authenticated user ID is used
only as transient input to the SHA-256 technical slot derivation together with operation/portfolio
scope.

## Closure governance

Implementation PR #67 was squash-merged into `develop` at
`0623d5ef326cd783b7dc0417dbcb02f18c506171` after exact-head CI #181, repeat independent
Canonical record: commit(s) `02aa2417a3caca79e2afc4e7b598b92055de96b7`.
squash-merge authorization.

When this closure record is canonical on `develop`, Stage 3.32 is CLOSED for P2-09 and P2-13. The
remaining original repository-audit backlog is 5 P2 and 10 P3 findings: P2-10/P2-11/P2-12/P2-16/P2-17
superseded. No architecture or product-scope expansion is introduced by closure governance.
