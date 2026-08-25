# Stage 3.38 — P3-05 Idempotency and Session Retention/Cleanup Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only; runtime implementation not authorized |
| Date | 2026-08-25 |
| Canonical planning base | `develop` at `305a53bb07136b274717ff48778a5e93d7b1607c` |
| Finding | P3-05 — idempotency/session cleanup |
| Prior closure | Stage 3.37 / P3-02 closure merged through PR #92 at `305a53bb07136b274717ff48778a5e93d7b1607c` |
| Runtime implementation authorized here | No |
| Commit / push authorized here | No |
| Merge authorized here | No |

## 1. Objective

Close only the design gap behind original audit finding P3-05 by freezing deterministic retention
boundaries and a bounded PostgreSQL cleanup mechanism for:

1. `investment.command_deduplication`, including completed exact-response replay artifacts; and
2. `identity.sessions`, including active-expired, revoked-expired, and legacy null-family session rows.

This planning gate does **not** itself implement cleanup and does **not** close P3-05.

## 2. Canonical evidence and current defect

### 2.1 Command idempotency/replay

Current PostgreSQL reservation code gives every command-deduplication row:

`expires_at = command.Now + 24 hours`

but current duplicate resolution does not consult `expires_at`. The current read-only replay lookup also
does not consult `expires_at`. Therefore, while the row remains present, the unique scope

`(principal_id, method, canonical_path, idempotency_key)`

can remain authoritative indefinitely.

The Web retry journal independently uses a 24-hour TTL and discards/replaces an expired browser retry
key after that boundary. The server-side row currently has the matching 24-hour timestamp but no
matching semantic or physical lifecycle.

Stage 3.32 deliberately requires exact completed-response replay before mutable business state is
consulted and permits read-only recovery of an already-completed import after its short-lived review
token expires. P3-05 must preserve those guarantees **inside the retained idempotency window**.

### 2.2 Refresh sessions

Every refresh-session row carries an application-owned `expires_at`. The default refresh-token TTL is
30 days, but configuration may change it; therefore the row's persisted `expires_at`, not a hard-coded
30-day cleanup rule, is authoritative.

Current rotation/logout logic can still locate expired rows. In particular, a revoked row is currently
handled as replay before the expiration check in refresh rotation, and logout has no equivalent expiry
gate before revocation behavior. Because session rows are not deleted, revoked-token replay containment
can therefore remain effective indefinitely rather than only for the token's valid lifetime.

There is no production cleanup path for expired session rows.

### 2.3 Physical cleanup

No bounded production cleanup path was found for either table. Both tables can therefore grow without
bound under sustained write/auth traffic even though each relevant row already carries an expiration
timestamp.

This is operational/retention debt. It is not evidence of a current financial-calculation error and it
does not reopen already-closed replay-containment findings.

## 3. Frozen retention contract

### 3.1 Authoritative time semantics

Stage 3.38 distinguishes request/service timestamps from the time that authoritatively decides an
expiry-sensitive state transition.

For P3-05, the authoritative PostgreSQL decision times are:

- **command mutation decision time** — a database wall-clock instant sampled only **after** the request
  has acquired the exact idempotency-key serialization point needed to decide whether the current
  generation is still authoritative or may be reclaimed;
- **session mutation decision time** — a database wall-clock instant sampled only **after** the
  existing per-user/session serialization locks required by refresh/logout have been acquired;
- **read-only replay lookup time** — a fresh database wall-clock instant evaluated for that lookup,
  not a timestamp captured earlier by the HTTP/service layer;
- **cleanup decision time** — a fresh database wall-clock instant used by the bounded cleanup
  candidate-selection statement.

A stale `CommandContext.Now`, auth-service `now`, transaction-start `now()` / `current_timestamp`, or
statement-start timestamp that was captured before a blocking serialization wait must **not** decide
the exact P3-05 expiry boundary.

The runtime implementation must use PostgreSQL wall-clock semantics that continue advancing across
lock waits (for example `clock_timestamp()` sampled/evaluated after the required serialization point)
or an equivalent mechanism with tests proving the same behavior.

An implementation must not merely place `clock_timestamp()` syntactically in a statement if that
expression can be evaluated before a blocking unique-index/lock wait. The implementation and tests
must prove that the authoritative value used for the expiry decision is obtained after serialization
has actually been acquired.

This Stage 3.38 clock rule is narrow:

- it governs idempotency-generation admission/reclamation, read-only replay expiry, refresh/logout
  expiry authority, and cleanup eligibility;
- it does **not** redefine BusinessDate, SQL `DATE`, SystemTimestamp, financial timestamps, or the
  existing business/audit use of `CommandContext.Now`.

### 3.2 Idempotency window

The authoritative server idempotency/replay window for one command generation is:

`created_at <= command_mutation_decision_time < expires_at`

for mutation/reclamation decisions, and:

`created_at <= replay_lookup_time < expires_at`

for read-only exact replay lookup.

A fresh command generation must persist:

`expires_at = created_at + 24 hours`.

For a **fresh** generation, `created_at` must be anchored to the same authoritative server-admission
decision that wins the exact-key serialization. It must not inherit a handler/service timestamp that
predates a serialization wait. The resulting server window is therefore truly 24 hours from fresh
server command admission.

The boundary is exact:

- authoritative decision time `< expires_at` → that generation remains authoritative;
- authoritative decision time `>= expires_at` → that generation/replay artifact is expired and no
  longer authoritative.

No grace period is added.

The existing Web retry journal's 24-hour TTL remains aligned with this server window.

### 3.3 Semantics before command expiry

If the request obtains the exact-key serialization point while the existing generation is still
unexpired according to the authoritative command mutation decision time, all Stage 3.32 guarantees
remain unchanged:

- same principal/method/path/key + same canonical request hash → exact original completed response;
- same principal/method/path/key + different canonical request hash → idempotency conflict;
- a committed incomplete/unsupported record retains the existing fail-closed duplicate behavior;
- exact replay is resolved before mutable portfolio/business state;
- an authentic historical import proof whose current review token is expired may recover only an
  already-completed exact response;
- expired review-token recovery remains read-only and cannot authorize a fresh write.

A request has **no entitlement to an old generation merely because its HTTP/service handler began
before expiry**. If it waits and does not acquire the exact-key serialization point until after the old
generation expires, the post-serialization decision governs.

### 3.4 Semantics at or after command expiry

If the authoritative command mutation decision time is `>= expires_at`:

- the old command-deduplication generation is no longer authoritative;
- the old exact-response replay artifact is no longer available through read-only replay lookup once
  the fresh lookup-time boundary is also at/after expiry;
- the same idempotency key may be admitted as a **new command generation** under the same
  principal/method/path;
- exactly one concurrent request may establish that fresh generation;
- the fresh generation receives a new command identity, new request hash, authoritative fresh
  `created_at`, `expires_at = created_at + 24 hours`, and empty replay-terminal fields before its new
  business effect is processed;
- old request hash/artifact state must never leak into the new generation;
- a stale/expired import review token still cannot authorize a new write. If its completed replay
  artifact is outside the idempotency window, recovery returns “not found” and normal current
  verification remains authoritative.

A request that started before the old generation expired but reached exact-key serialization only after
another request legitimately established the post-expiry generation must observe the **new** generation
under the ordinary duplicate/in-flight/conflict rules. It cannot demand the old artifact based on its
earlier handler timestamp.

Outside the 24-hour window, the API no longer guarantees suppression of an old client retry. This is
an explicit bounded-idempotency contract, not a claim that the prior business effect disappears.

### 3.5 Session retention boundary

A session row remains security-authoritative only when the authoritative **post-lock session mutation
decision time** is:

`session_mutation_decision_time < session.expires_at`.

Before that boundary:

- active sessions remain eligible for normal refresh/logout processing;
- revoked-but-unexpired session rows remain retained;
- replay of a revoked-but-unexpired refresh token must preserve Stage 3.28 family containment;
- legacy null-family revoked rows preserve their conservative “revoke active user sessions” behavior.

At:

`session_mutation_decision_time >= session.expires_at`

the presented refresh token is invalid regardless of whether its physical row has already been deleted.

Critically, the expiry comparison must occur **after** the required user/session serialization locks
and **before** any replay/family/user containment, rotation, or logout-revocation side effect.

Therefore:

- an expired revoked token must not revoke a currently active family or currently active user sessions;
- an expired active token must not rotate, refresh, or authorize logout revocation effects;
- a request that began before expiry but waited on the existing serialization lock until after expiry
  is treated as expired;
- the expired row may be physically deleted by bounded cleanup.

The persisted per-row `expires_at` remains authoritative. No separate hard-coded 30-day cleanup
threshold is introduced.

This planning stage does not redefine how a newly issued session's existing persisted `expires_at` is
calculated; it freezes only the post-lock authority check for using an already-presented session row.

## 4. Logical expiry must not depend on cleanup scheduling

Physical cleanup is an operational lifecycle mechanism. Correct request semantics must not depend on
whether a cleanup batch happened to run.

Runtime implementation must therefore make all logical boundaries explicit:

1. read-only command replay lookup must refuse a row when a **fresh database lookup-time** is at/after
   `expires_at`; it must not use a stale `command.Now` captured before the lookup;
2. exact command reservation/reclamation must acquire its exact-key serialization point first and only
   then make the expiry decision from an authoritative advancing database wall clock;
3. a request that reaches exact-key serialization after the old generation's expiry has no claim on
   the old generation even if the request started earlier;
4. the fresh command generation's `created_at` and `expires_at` must be anchored consistently to the
   winning fresh-admission decision, with `expires_at = created_at + 24 hours`;
5. refresh/logout must acquire the existing required serialization locks first, sample/evaluate the
   authoritative post-lock decision time, and reject the row if expired **before** any
   replay-containment, family/user revocation, rotation, or logout effect;
6. bounded physical cleanup uses its own fresh database cleanup-time and is never a prerequisite for
   the above logical decisions.

PostgreSQL transaction-start `now()` / `current_timestamp` is not sufficient for a transaction that
may wait across the boundary. A statement-start timestamp is likewise not sufficient if the statement
can block before acquiring the serialization point. The implementation must use an advancing
wall-clock value obtained after the relevant wait/lock, or prove equivalent semantics.

This guarantees deterministic behavior even if an expired row has not yet been physically reaped.

## 5. Exact-key command reclamation and concurrency

The runtime must not implement expiry as an unsafe “delete then insert and hope” sequence.

For the exact unique command scope, reservation must provide one serializable logical decision with
these outcomes:

- no authoritative row → exactly one fresh reservation wins;
- existing row + authoritative post-serialization decision time `< expires_at` → current
  duplicate/conflict/replay semantics;
- existing row + authoritative post-serialization decision time `>= expires_at` → exactly one request
  establishes the fresh generation;
- concurrent same-key requests that straddle or follow expiry cannot create two fresh business effects;
- a loser observes the winning **new** unexpired generation and follows the existing
  duplicate/in-flight/conflict rules;
- no request can mix the old generation's request hash/replay artifact with the new generation's
  identity/timestamps.

The exact-key **serialization point** is part of the contract. The implementation may use a
conflict-aware retry loop with row locking, an expiry-aware PostgreSQL upsert whose timing semantics are
proven, a transaction-scoped exact-scope advisory lock plus the existing uniqueness constraint, or an
equivalent transactionally proven mechanism.

Whatever mechanism is chosen:

1. the expiry decision must be made after the request actually owns the serialization point;
2. the fresh generation's `created_at` must use the authoritative fresh-admission time obtained at that
   point;
3. `expires_at` must be exactly 24 hours after that persisted `created_at`;
4. a pre-wait HTTP/service timestamp must not be reused as the expiry/admission authority;
5. the implementation must not rely on the global cleanup batch selecting the exact row.

A dedicated PostgreSQL concurrency regression must deliberately block one request so that its handler
starts before expiry but its exact-key decision occurs after expiry, while another request establishes
the new generation first. The result must prove:

- the delayed request has no access to the old generation based on its start time;
- old/new artifact or hash state never mixes;
- at most one post-expiry generation/business effect wins.

## 6. Bounded physical cleanup

### 6.1 Batch size

The fixed runtime cleanup batch size for Stage 3.38 is:

`128 rows`

per relevant table per triggering mutating transaction.

It is intentionally a code constant, not a user-controlled environment variable, to keep lock and
query cost bounded and to avoid an unreviewed operator setting turning cleanup into an unbounded
request-path operation.

Changing this batch size later is an operational tuning change and must be separately reviewed.

### 6.2 PostgreSQL selection pattern

Each cleanup batch must:

- obtain/use a fresh database cleanup-time for the candidate-selection statement;
- select only rows with `expires_at <= cleanup_decision_time`;
- order candidates by `(expires_at, id)`;
- limit candidate selection to 128;
- lock candidates with `FOR UPDATE SKIP LOCKED`;
- delete only the selected candidate IDs;
- execute in the same database transaction as the triggering mutation.

The intended shape is a bounded candidate CTE plus `DELETE ... USING`, or an equivalent proven query.
A single fresh `clock_timestamp()`-based cleanup instant may be captured inside that non-waiting
`SKIP LOCKED` selection statement so one pass uses a coherent eligibility boundary.

`SKIP LOCKED` is required so multiple application instances can make cleanup progress without waiting
on the same expired rows.

### 6.3 Trigger model

No new microservice, external worker, queue, Redis lifecycle, cron provider, or unbounded background
goroutine is introduced for P3-05.

Cleanup is application-owned and opportunistic:

- an idempotent command reservation performs one bounded expired-command cleanup batch;
- auth transactions that create, rotate, or revoke refresh sessions perform one bounded expired-session
  cleanup batch;
- registration counts as a session-creating auth transaction;
- read-only exact replay lookup remains read-only and must not perform cleanup;
- read-only user/business queries do not become hidden maintenance operations.

This is sufficient for storage-growth control: if mutating traffic stops, these tables stop growing;
under sustained relevant mutating traffic, each operation can reap up to 128 expired rows while
creating at most the normal small number of new rows.

This is **not** claimed to provide a strict wall-clock deletion SLA for an idle database. Privacy
deletion/backup-destruction SLAs remain outside P3-05.

### 6.4 Cleanup failure semantics

Cleanup runs inside the triggering mutation's transaction.

If the bounded cleanup query fails:

- the transaction fails and rolls back;
- no partial command/session mutation is committed;
- no cleanup failure is silently converted into a successful business write;
- the next request may retry normally.

The two dedicated expiry indexes plus the 128-row bound and `SKIP LOCKED` are required to keep this
fail-closed choice operationally narrow.

## 7. Database migration contract

Existing migrations remain immutable.

Runtime implementation is expected to add one new reversible migration, provisionally:

`000007_stage_03_38_operational_retention`

containing only additive cleanup-support indexes:

- `investment.command_deduplication (expires_at, id)`;
- `identity.sessions (expires_at, id)`.

The down migration drops only those Stage 3.38 indexes.

No existing row requires a data backfill: both tables already have non-null `expires_at`.

No schema owner/runtime privilege expansion is required. The existing runtime role already permits
DELETE on these mutable tables. `investment.transaction_entries` and `audit.events` remain protected
by their existing append/read-only runtime policy.

Adding this new migration does **not** close or absorb P3-08. The known migration-validator-vs-policy
gap remains a separate finding.

## 8. Public API contract

P3-05 runtime must clarify the OpenAPI `Idempotency-Key` contract so clients can see that:

- exact idempotency/replay is guaranteed for 24 hours from initial server command admission;
- before expiry, reuse follows existing replay/conflict rules;
- at/after expiry, the same key may be admitted as a new command.

This is a retention clarification of the existing 24-hour server/browser behavior.

No new endpoint is added.

No auth API response shape changes.

No frontend runtime behavior change is required because the existing browser retry journal already
expires after 24 hours. Existing frontend expiry tests remain evidence and may be strengthened only if
needed for cross-layer contract proof.

## 9. Required runtime implementation scope

A later separately authorized Stage 3.38 runtime candidate may touch only what is necessary to implement
this plan, expected to include:

- PostgreSQL replay reservation/lookup lifecycle code;
- PostgreSQL auth-session resolution/cleanup lifecycle code;
- focused unit/integration/concurrency tests;
- the new additive retention-index migration and down migration;
- OpenAPI `Idempotency-Key` wording plus focused contract validation;
- Stage 3.38 implementation evidence documentation;
- narrowly necessary registry bookkeeping.

No runtime implementation is authorized by this planning document.

## 10. Mandatory regression evidence

Runtime approval must include at least the following tests.

### 10.1 Idempotency/replay

1. same key + same payload before expiry at the authoritative serialized decision point → exact
   original replay;
2. same key + different payload before expiry at that point → conflict;
3. exact read-only import recovery before command expiry still works after review-token expiry;
4. fresh read-only lookup at the exact boundary `lookup_time == expires_at` → replay is not found;
5. after expiry, the same key can establish a fresh reservation with no old artifact/hash fields;
6. the fresh generation persists `expires_at = created_at + 24 hours`, with `created_at` anchored to
   fresh server admission rather than an earlier handler timestamp;
7. after expiry, an expired review token cannot convert missing replay into a fresh import write;
8. concurrent same-key requests after expiry produce at most one new business effect;
9. **straddling command race:** request A begins before expiry and is deliberately blocked before the
   exact-key serialization point; request B reaches the post-expiry serialization point first and
   establishes the new generation; when A continues it must not receive/use the old generation based
   on its captured start time, must observe only the new generation's state, and old/new artifact/hash
   fields must never mix;
10. an unexpired row cannot be removed by cleanup;
11. one cleanup pass deletes at most 128 expired command rows;
12. repeated cleanup passes can drain a backlog larger than 128;
13. concurrent cleanup passes do not block each other on the same candidates and do not delete
    unexpired rows.

### 10.2 Sessions

14. revoked-but-unexpired refresh replay still revokes the active session family;
15. legacy null-family revoked-but-unexpired replay retains conservative user-session containment;
16. authoritative post-lock decision time exactly equal to `expires_at` is invalid before any
    family/user revocation effect;
17. revoked-and-expired replay does not revoke a currently active descendant/family;
18. expired active sessions cannot refresh, rotate, or cause logout revocation effects;
19. **straddling session race:** a refresh request and a logout request each have a regression where
    the handler begins before expiry but is deliberately blocked on the existing user/session
    serialization lock until after expiry; after the lock is acquired, the request must cause zero
    family/user revocation, zero rotation, and zero logout-session mutation effect other than permitted
    rejection/audit evidence;
20. one cleanup pass deletes at most 128 expired session rows and leaves unexpired rows untouched;
21. repeated/concurrent cleanup passes safely drain an expired-session backlog;
22. auth audit evidence remains present; P3-05 cleanup never deletes `audit.events`.

### 10.3 Migration/privilege/contract

23. migration up/down validation passes;
24. required `(expires_at, id)` indexes exist after up and disappear after down;
25. production runtime-role verification still passes without privilege expansion;
26. OpenAPI validation passes and the 24-hour idempotency window is asserted by a focused contract test;
27. tests prove expiry-sensitive mutation decisions do not rely on stale HTTP/service timestamps or
    PostgreSQL transaction-start `now()` across lock waits;
28. full Go test, Go vet, race-test, frontend, migration, OpenAPI and security CI remain green.

## 11. Historical rows

Historical command/session rows are not guessed, rewritten, or assigned fabricated timestamps.

They already contain `expires_at`.

Once the Stage 3.38 runtime is canonical:

- rows whose persisted `expires_at` is at/before the fresh database cleanup decision time become eligible for bounded primary-database cleanup;
- unexpired rows retain their exact current semantics until their own boundary;
- cleanup does not modify prior financial effects;
- cleanup does not remove audit evidence.

## 12. Privacy and audit boundary

P3-05 provides ordinary primary-database operational retention for two already-expiring technical
surfaces.

It does **not** claim:

- account-deletion completion;
- anonymization completion;
- backup or replica destruction;
- provider-specific retention proof;
- cryptographic erasure;
- deletion-marker completion;
- Stage 3.25 Privacy Security Review completion.

`audit.events` is explicitly out of cleanup scope and remains append/read-only for the runtime role.

The Stage 3.21/3.25 privacy work may later consume P3-05 primary-database cleanup as one evidence input,
but P3-05 cannot close those privacy gates.

## 13. Explicit non-scope

This plan does not absorb or close:

- P3-04 — Unicode / general `maxLength` semantics;
- P3-06 — split `backend-go/internal/httpapi/api.go`;
- P3-07 — transaction-form fixture/default cleanup;
- P3-08 — migration validator weaker than migration policy;
- P3-09 — Next.js maintenance;
- P3-10 — Fiber maintenance;
- Stage 3.25 privacy Security Review evidence work;
- P2-14 auth rate-limiter lifecycle, already separately closed.

Also out of scope:

- BusinessDate / SQL `DATE` semantics;
- SystemTimestamp semantics;
- financial arithmetic/Decimal behavior;
- transaction ledger deletion;
- snapshot deletion;
- audit-event deletion;
- user-account deletion;
- new worker/service/Redis/Kubernetes/provider dependencies;
- dependency upgrades.

## 14. Rejected alternatives

### Delete every old row with an unbounded statement

Rejected because it creates uncontrolled lock/latency behavior and does not solve exact-key expiry
semantics independently of cleanup timing.

### Cleanup only, without making lookup/reservation/session logic expiry-aware

Rejected because correctness would depend on whether a cleanup batch happened to run.

### Keep command replay forever

Rejected because it contradicts the existing 24-hour expiration metadata/browser retry window and
leaves the unique key namespace/storage unbounded.

### Delete revoked sessions immediately

Rejected because revoked-but-unexpired rows are required for Stage 3.28 refresh replay containment.

### Keep revoked sessions forever for replay detection

Rejected because replay containment must be bounded by the refresh token's valid lifetime; an expired
token must not retain indefinite authority to revoke current sessions.

### Dedicated external cleanup worker or cron provider

Rejected for P3-05 because bounded opportunistic cleanup makes progress under the same traffic that
creates the rows, requires no new deployable component, and preserves the current modular-monolith
scope.

### Mutating read-only replay recovery

Rejected because Stage 3.32 explicitly requires that recovery path to remain read-only.

## 15. Planning review history

The first independent pre-commit Stage 3.38 planning review returned `REQUEST CHANGES` with one P3
blocker and no P0/P1/P2 findings.

The blocker was a stale-clock race: the first candidate defined exact expiry using timestamps captured
by the service before PostgreSQL serialization/lock waits. A request could begin before expiry, wait
past expiry, and still claim old command/session authority even though another request had already
legitimately crossed the boundary.

This revised candidate remediates that blocker narrowly by freezing:

- post-serialization authoritative command mutation time;
- fresh lookup-time for read-only replay;
- post-lock authoritative refresh/logout time before any containment side effect;
- fresh-generation `created_at` / `expires_at` anchored to fresh server admission;
- explicit command/session regressions that straddle the expiry boundary.

No retention architecture, 128-row cleanup batch, index design, privilege model, deployable topology,
or finding scope was broadened by that remediation.

## 16. Acceptance gate

P3-05 remains OPEN after this planning document.

Runtime implementation may begin only after:

1. independent planning review returns `APPROVED`;
2. separate human commit/push authorization;
3. planning PR exact-head CI is green;
4. fresh independent published-head planning review returns `APPROVED`;
5. separate human squash-merge authorization;
6. planning PR is merged into `develop`.

Runtime implementation then requires its own full review/CI/merge cycle, followed by separately governed
closure evidence before P3-05 may be canonically CLOSED.
