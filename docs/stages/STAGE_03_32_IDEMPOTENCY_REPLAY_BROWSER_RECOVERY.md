# Stage 3.32 — Exact Idempotency Replay and Browser Retry Recovery

| Field | Value |
| --- | --- |
| Status | Implementation candidate; code verification green; independent final review and human merge authorization pending |
| Owner | Principal Architect |
| Baseline | `develop` at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746` |
| Branch | `fix/stage-03-32-idempotency-recovery` |
| Implementation PR | #67 |
| Implementation merge | Pending |
| Code-verification head | `52d329bf0567c5223038aaa24e2f531df1e9e4a9` |
| Code-verification CI | GitHub Actions #163 — SUCCESS, all six jobs passed |
| Independent final review | Pending |
| Human implementation merge authorization | Pending |
| Trigger | Repository-audit P2-09 and P2-13 |
| Scope | Exact original-response idempotent replay, atomic replay-artifact persistence, import retry recovery across review-token expiry, short-lived browser retry identity, regression and migration coverage |
| Out of scope | P2-10/P2-11/P2-12/P2-16/P2-17, all P3 findings, Stage 3.25 privacy Security Review evidence work, provider/backup retention, product-scope expansion |

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
business effect, execute the response builder once, and return the same artifact.

Replay persistence is database-backed rather than process memory. A regression closes the first
Store connection completely, opens a new Store against the same PostgreSQL database, and verifies
that the original response is recovered without executing the response builder or financial write.

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

## P2-13 — browser retry identity was lost on reload/remount

### Observed defect

The browser kept the current idempotency intent only in React `useRef`. Same-mount retries reused the
key, but reload/remount lost it. After an ambiguous server success, the next submission could receive
a new UUID and become a second command.

### Remediation

The Web layer now maintains a short-lived technical retry journal in `sessionStorage`.

The durable browser record contains only:

- version;
- opaque idempotency key;
- expiry timestamp.

The storage slot name is derived from SHA-256 of a technical scope. It does not persist the financial
payload, transaction amounts, CSV, source-account label, portfolio data, review token, access token,
or CSRF token. The journal therefore remains a retry-control record, not browser business-data
storage.

Within the same mounted interaction, a changed intent rotates to a new key. Across reload/remount,
an unresolved technical key is recovered and sent again. The backend remains authoritative: a
matching command replays the original response, while a proven idempotency conflict allows the
browser to discard the stale retry key. Confirmed success also clears the journal.

The journal is applied to portfolio create, manual transaction append, and import append.

## Regression evidence

The Stage 3.32 test set includes:

- exact portfolio HTTP status/body/request-ID/trace-ID replay;
- exact transaction replay without a second ledger entry;
- response-builder failure rolls back both business write and command reservation;
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
- browser storage exposes neither raw intent/payload nor raw technical scope;
- successful writes clear the retry journal;
- full frontend typecheck, tests, and production build remain green.

Code-verification head `52d329bf0567c5223038aaa24e2f531df1e9e4a9` passed GitHub Actions CI
#163 with all six workflow jobs successful: Go tests with PostgreSQL/migrations, PostgreSQL migration
validation, frontend build/typecheck/tests, Python tests, OpenAPI contract validation, and Docker
Compose configuration validation.

## Adversarial implementation review findings resolved before independent review

During the implementation pass, the following defects were found and corrected rather than waived:

- the first migration draft used `UPDATE` in an up migration and was rejected by the repository
  migration validator; the final up migration is additive with no data rewrite;
- the first import-recovery ordering could make every fresh import pay for a replay lookup; recovery
  now occurs only after review proof expiry;
- the first recovery rule was too broad and could have considered other invalid tokens; recovery is
  now restricted to otherwise-valid expired tokens, with a negative tampered-signature regression;
- financial replay initially consulted mutable portfolio state before resolving the completed command;
  replay resolution now occurs first;
- early HTTP unit fakes retained Fiber request-buffer-backed strings rather than owning persisted
  bytes; the fakes now clone artifacts so they model PostgreSQL persistence while keeping the strict
  exact-header assertions;
- a concurrency regression initially referenced the wrong helper name; it was corrected before the
  successful code-verification run.

No unresolved code blocker is recorded by this implementation pass. This is not an independent final
review verdict.

## Privacy and retention boundary

The replay artifact can contain the same financial response data that was returned to the authenticated
user. Stage 3.21 already classifies `investment.command_deduplication` as a principal/correlation
surface whose future disposition is deletion after the authorization/retry window, with cleanup,
replica, backup, and replay evidence still required.

Stage 3.32 does not claim that privacy lifecycle is implemented or approved. It does not introduce a
cleanup worker, provider retention policy, backup purge, anonymization mechanism, or Stage 3.25
Security Review evidence. The pre-existing `expires_at` field and the browser retry TTL remain
operational retry metadata; physical database/provider lifecycle remains a separate privacy track.
This residual boundary must not be cited as privacy closure.

## Scope boundary and next gate

Stage 3.32 is not CLOSED by this implementation record. P2-09/P2-13 may be marked closed only after:

1. the final implementation head passes exact-head CI;
2. an independent final review returns `APPROVED` with no unresolved blocker;
3. explicit human implementation merge authorization is given;
4. PR #67 is squash-merged into `develop`;
5. separate closure governance records the canonical merge and updates the repository status.

P2-10/P2-11/P2-12/P2-16/P2-17 and all P3 findings remain outside this stage. Stage 3.25 privacy
Security Review evidence planning remains separate and is not superseded.
