# Stage 3.38 — P3-05 Idempotency and Session Retention/Cleanup Implementation

| Field | Value |
| --- | --- |
| Status | Runtime merged / separate closure governance pending |
| Date | 2026-08-26 |
| Finding | P3-05 — idempotency/session cleanup |
| Canonical runtime base | `develop` at `a944f1e5d5ee7d84db5393e8760eda254d732edd` |
| Planning gate | PR #93 squash-merged at `a944f1e5d5ee7d84db5393e8760eda254d732edd` |
| Runtime branch | `fix/stage-03-38-p3-05-retention-runtime`; published as PR #94 |
| Runtime commit / push | Exact published head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` |
| Runtime merge | PR #94 squash-merged at `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` |
| Final finding status | OPEN; runtime merge alone will not close P3-05 |

## 1. Finding / symptom

P3-05 tracks operational-retention debt on `investment.command_deduplication` and
`identity.sessions`. Both surfaces already had `expires_at`, but runtime authority and bounded
physical cleanup did not consistently honor it.

## 2. Root cause

Command reservation persisted a 24-hour expiry but conflict resolution and read-only replay did not
consult expiry. Authentication used a timestamp captured before PostgreSQL serialization, handled
revoked-session replay before expiry in refresh, and had no equivalent pre-effect expiry gate in
logout. Neither table had a bounded production cleanup path or expiry-leading global index.

## 3. Failure scenarios

1. A command key could remain replay/conflict-authoritative indefinitely while the browser retry
   journal expired after 24 hours.
2. A request could start before command/session expiry, block on database serialization, and cross the
   boundary while still using stale pre-lock time.
3. An expired revoked refresh token could retain family/user containment authority.
4. Sustained relevant write/auth traffic could grow both tables without bounded online cleanup.

## 4. Impact

The defect affects API lifecycle consistency, stale-session security authority, operational table
growth, and primary-database technical-retention hygiene. P3-05 does not identify a financial
calculation, Decimal, BusinessDate, ledger, or snapshot arithmetic defect.

## 5. Severity rationale

P3 remains appropriate because the debt is lifecycle/operational rather than an immediate
authentication bypass or demonstrated financial-integrity failure. The security-relevant part is
bounded stale revocation authority after credential expiry, not creation of a valid authenticated
session from an expired credential.

## 6. Existing guarantees violated

The old runtime left incomplete the 24-hour browser/server retry alignment, exact replay retention
boundary, "expired credentials have zero current authority" rule, bounded maintenance requirement,
and retention of revoked-but-unexpired Stage 3.28 evidence without indefinite post-expiry authority.

## 7. Considered solutions

Considered approaches included unbounded deletion, cleanup-only correctness, permanent replay,
immediate deletion of revoked sessions, permanent revoked-session retention, a new worker/cron
service, and expiry-aware logical semantics plus bounded opportunistic cleanup. For exact-key command
serialization, planning allowed a proven conflict/retry loop, expiry-aware upsert, transaction-scoped
exact-scope advisory lock plus uniqueness, or an equivalent mechanism.

## 8. Chosen remediation

The runtime candidate uses:

- transaction-scoped exact-scope advisory serialization for command admission/reclamation;
- row locking plus a fresh `clock_timestamp()` only after serialization;
- a new UUID/hash/timestamps/empty artifact state for each post-expiry command generation;
- fresh database wall clock for read-only replay expiry without writes;
- post-user-lock/post-session-lock expiry checks before refresh replay containment, rotation, logout
  revocation, or all-session revocation;
- fixed 128-row cleanup batches ordered by `(expires_at, id)` with `FOR UPDATE SKIP LOCKED`;
- opportunistic cleanup in command reservation and auth session create/rotate/revoke transactions;
- additive `(expires_at, id)` indexes in migration `000007_stage_03_38_operational_retention`;
- OpenAPI publication of the 24-hour server idempotency window.

## 9. Why this solution

Correctness no longer depends on cleanup scheduling. Exact command authority is decided after one
explicit database serialization point, and session authority is decided after the pre-existing auth
serialization locks. Cleanup remains bounded and inside the modular monolith, so no new deployable
component is required.

## 10. Rejected alternatives

- Unbounded `DELETE`: excessive lock/latency risk.
- Cleanup-only correctness: request behavior would depend on cleanup timing.
- Permanent replay: contradicts the approved 24-hour lifecycle.
- Immediate revoked-session deletion: destroys Stage 3.28 replay evidence before token expiry.
- Permanent revoked-session retention: gives expired evidence indefinite operational lifetime.
- New worker/cron/Redis path: unnecessary architecture expansion.
- Service `Now()`, transaction-start `now()`, or a timestamp sampled before blocking: stale across
  expiry-straddling waits.

## 11. Trade-offs

Each idempotent reservation takes one advisory lock and one bounded cleanup pass. Relevant auth
mutations take one bounded session cleanup pass. Advisory-hash collisions can serialize unrelated
scopes but cannot cross-authorize data because unique/predicate identity remains authoritative.
Cleanup provides growth control, not a strict deletion SLA for an idle database.

## 12. Regression tests

The candidate adds proof for:

- `expires_at == decision_time` being expired;
- pre-expiry exact replay;
- expired read-only replay returning not found;
- same-key fresh generation after expiry with new ID and exactly 24-hour server window;
- two concurrent post-expiry requests converging on one new business effect/artifact;
- a deliberately blocked command request starting pre-expiry but deciding post-expiry, with no
  old/new hash/artifact mixing;
- refresh blocked across expiry causing zero containment/rotation authority;
- logout blocked across expiry causing zero user-revocation authority;
- fixed 128-row command/session cleanup batches;
- unexpired rows surviving cleanup;
- both expiry-leading indexes;
- OpenAPI lexical contract preservation plus 24-hour lifecycle publication.

All existing Stage 3.28 and Stage 3.32 regressions remain mandatory.

## 13. Adversarial review findings

Planning review history:

1. first independent planning review: `REQUEST CHANGES`, one P3 stale-clock/serialization blocker;
2. revised v2: post-serialization command time, fresh replay lookup time, post-lock session time, and
   boundary-straddling regressions frozen;
3. renewed pre-commit planning review: `APPROVED`, P0/P1/P2/P3 None;
4. fresh published-head planning review of PR #93 exact head
   `7a4ef7115b5fbab4c9017c6032112f028825c959`: `APPROVED`, P0/P1/P2/P3 None.

Independent pre-commit runtime review history is preserved below. The v2 and renewed-v3 reviews produced substantive runtime `REQUEST CHANGES`; the subsequent exact-v4 and DOCFIX2 reviews produced documentation/evidence-only `REQUEST CHANGES`. The current DOCFIX4 candidate awaits renewed review. Any runtime `REQUEST CHANGES` must be appended here
with exact evidence rather than overwritten.

### First independent pre-commit runtime review — `REQUEST CHANGES`

The first independent pre-commit runtime review of the 11-file v2 candidate returned
`REQUEST CHANGES`, with P0/P1/P2 = None and two P3 blockers.

#### P3 runtime review blocker 1 — global-cleanup lock inversion

The reviewer identified a concrete lock-order cycle in the v2 ordering.

For commands, v2 acquired the exact advisory lock, then took unrelated expired-row locks in global
cleanup, and only afterwards acquired the exact command row `FOR UPDATE`. Two transactions on distinct
expired scopes could therefore each hold the other's exact row through cleanup and then wait on the
other transaction's exact row.

The auth path had the analogous problem because rotate/logout cleanup ran before the established
user-advisory-lock -> presented-session-row-lock sequence. `SKIP LOCKED` prevents cleanup-to-cleanup
waiting, but it cannot prevent a later exact-row wait from completing a cycle.

The reviewer classified this as P3 availability/operational debt: PostgreSQL deadlock victim rollback
would preserve consistency, but valid writes could fail under sustained expired-row traffic.

#### P3 runtime review blocker 2 — in-place reclamation branch was not forced

The v2 tests were green, but because global command cleanup ran before exact row acquisition, an
expired target row could be physically deleted before reservation reached the in-place reclamation
`UPDATE`. Therefore the tests did not deterministically prove the previously corrected reclamation
SQL path, including the reset of every old replay-artifact field.

#### Narrow v3 remediation

The v3 candidate keeps the Stage 3.38 architecture unchanged and applies only the requested narrow
correction:

- command order is exact advisory lock -> exact row select/reclaim/insert + authoritative decision ->
  bounded global cleanup -> return/continue;
- refresh/logout order is owner resolution -> user advisory lock -> presented session `FOR UPDATE` ->
  post-lock expiry/security decision -> bounded global cleanup -> security/business side effects;
- registration and ordinary session creation retain cleanup because they do not subsequently wait on a
  presented session row;
- batch size 128, `(expires_at,id)` order, `SKIP LOCKED`, same-transaction cleanup and fail-closed
  cleanup errors remain unchanged;
- a deterministic reclamation regression inspects the replacement row before completion and proves
  new UUID/hash, null terminal/artifact fields, fresh database admission time, and exact 24-hour
  expiry;
- deterministic PostgreSQL lock-order tests hold the exact command row or auth user lock and prove an
  unrelated expired exact row remains lockable while the mutation is blocked before its canonical
  exact acquisition point;
- a concurrent two-scope expired-command regression proves both reservations complete without a
  database deadlock error.

This review history is preserved; the first runtime `REQUEST CHANGES` is not replaced by the later
candidate.

### Runtime lifecycle normalization after renewed v3 review

- Runtime iteration 1 = v2 local candidate. It included the SQLSTATE 42P08 correction and then received the first independent `REQUEST CHANGES`.
- Runtime iteration 2 = v3 local remediation. It closed the original cleanup/exact-lock inversion and deterministic reclamation evidence gap, then received the renewed independent v3 `REQUEST CHANGES`.
- Runtime iteration 3 = current v4 local remediation. It addresses the mixed-version unique-conflict admission timestamp race, auth cleanup ordering after broader family/user mutations, and the contradictory documentation labels.

No runtime commit, push, PR, CI, or merge evidence exists yet for this candidate. P3-05 remains **OPEN**.

### Renewed independent pre-commit runtime review of v3 — `REQUEST CHANGES`

The renewed independent review confirmed that the two blockers from the first runtime review were genuinely remediated, but returned a new `REQUEST CHANGES` verdict with P0/P1/P2 = None and three P3 blockers:

1. the no-row `ON CONFLICT DO NOTHING` path could persist a pre-wait admission timestamp when a mixed-version/non-cooperating writer held the same unique key;
2. auth cleanup still preceded broader family/user-wide revocation updates and could recreate a cross-user cleanup-to-containment deadlock;
3. the implementation record contained contradictory current-review and runtime-iteration wording.

V4 addresses those findings without changing the frozen Stage 3.38 scope. No runtime GitHub PR/head/CI/merge evidence is asserted, and P3-05 remains **OPEN**.

### Exact v4 pre-commit review — `REQUEST CHANGES` (documentation only)

The exact v4 review package was independently hash-verified and reviewed. The reviewer found no
remaining P0/P1/P2/P3 runtime blocker in the technical Stage 3.38 implementation. The three substantive
blockers from the valid renewed v3 review were confirmed remediated:

- post-UNIQUE-conflict admission time is finalized only after the potentially blocking INSERT returns
  as winner;
- auth cleanup runs after broader family/user/allSessions mutation work;
- the new mixed-version and auth lock-order regressions are meaningful and pass.

The review nevertheless returned `REQUEST CHANGES` for one P3 governance/documentation blocker:
the durable implementation record still contained the stale statement
`Runtime iteration 1 is this local candidate`, contradicting the normalized lifecycle where iteration 1
is v2, iteration 2 is v3, and iteration 3 is current v4.

This documentation-only remediation replaces that stale current-candidate assertion with historically
unambiguous v2 wording and strengthens local evidence checks so the obsolete sentence must be absent.
No runtime code, migration, OpenAPI behavior, privilege, or retention semantics are changed by this
iteration.

No runtime commit, push, PR, GitHub CI, or merge evidence exists yet. P3-05 remains **OPEN** until the
full runtime publication and separately governed closure lifecycle complete.

### Exact v4 DOCFIX2 pre-commit review — `REQUEST CHANGES` (historical quotation only)

The exact DOCFIX2 package preserved the runtime v4 implementation unchanged and passed the full local
verification suite. The reviewer confirmed that the stale active assertion had been removed and that
the normalized lifecycle correctly states iteration 1 = v2, iteration 2 = v3, and iteration 3 =
current v4.

The review nevertheless returned one P3 governance/evidence blocker: the newly added historical record
for the prior exact-v4 documentation review accidentally quoted the corrected sentence
`Runtime iteration 1 was the v2 local candidate` as though it were the rejected sentence. The actual
rejected wording had been `Runtime iteration 1 is this local candidate`.

DOCFIX3 changes only that durable historical quotation and adds a negative-control consistency proof:
the checker must reject a temporary document containing an active stale current-candidate assertion,
while accepting the real candidate where the stale wording appears only inside preserved historical
`REQUEST CHANGES` evidence.

No runtime code, migration, OpenAPI behavior, privilege, retention semantic, commit, push, PR, GitHub
CI, or merge evidence is changed or fabricated. P3-05 remains **OPEN**.

### Exact v4 DOCFIX3 pre-commit review — `REQUEST CHANGES` (stale review-count only)

The exact DOCFIX3 package was independently hash-verified. The reviewer confirmed that the technical
v4 runtime blockers remained closed, the DOCFIX2 historical-quotation defect was closed, and the
semantic negative-control documentation checks were meaningful.

The review nevertheless returned one P3 governance/evidence blocker in section 13: the active
introductory prose still said `Independent runtime adversarial review has occurred twice`, while the
durable record already preserved four runtime-candidate review cycles: the first v2 review, renewed-v3
review, exact-v4 documentation-only review, and DOCFIX2 quotation-only review.

DOCFIX4 removes the hard-coded stale count from active prose and replaces it with role-based history:
v2 and renewed-v3 were substantive runtime `REQUEST CHANGES`; exact-v4 and DOCFIX2 were
documentation/evidence-only `REQUEST CHANGES`; the current DOCFIX4 candidate awaits renewed review.

The semantic checker is strengthened so the stale hard-coded count is allowed only as preserved
historical `REQUEST CHANGES` evidence, never as active lifecycle prose. A second negative-control proof
injects the stale count into active section-13 prose and requires the checker to reject it.

No runtime Go code, tests, migration, OpenAPI behavior, privilege, retention semantic, commit, push, PR,
GitHub CI, or merge evidence is changed or fabricated. P3-05 remains **OPEN**.

## 14. Remediation iterations

Planning iteration 1 defined retention/batching but used pre-serialization time.
Planning iteration 2 corrected the authoritative clock boundary.
Runtime iteration 1 was the v2 local candidate: advisory serialization, post-lock clocks, bounded cleanup,
migration/indexes, OpenAPI wording, and focused concurrency regressions.

### Runtime verification iteration 1 — local failure and correction

The first local runtime verification run did **not** pass and is preserved as engineering evidence.
Static/OpenAPI/migration validation and `go vet` passed, as did the focused session-expiry and bounded
cleanup tests, but the first three command-retention PostgreSQL tests failed before exercising their
assertions with PostgreSQL `SQLSTATE 42P08`:

`inconsistent types deduced for parameter $7`

Root cause: the fresh-generation INSERT reused one untyped bind parameter both as a `TIMESTAMPTZ`
column value and as the left operand of interval arithmetic. PostgreSQL could not infer one consistent
parameter type. The same ambiguity also existed in the expired-generation UPDATE for its admission
timestamp.

Correction: the runtime candidate now explicitly casts the authoritative post-serialization admission
parameter to `timestamptz` at both SQL use sites. This is a narrow SQL typing correction; it does not
change the approved clock source, serialization point, retention duration, or business semantics.

Because the failed verification exposed the candidate before commit/push, no GitHub runtime evidence
was invalidated. All focused and full verification is rerun from scratch after this correction.

The correction pass also strengthens mandatory planning evidence with explicit tests for concurrent
`SKIP LOCKED` cleanup, existing runtime-role cleanup privileges plus protected ledger/audit DELETE
denial, preservation of `audit.events`, and the cross-layer rule that an expired import proof cannot
authorize a fresh write when retained replay is no longer found.

### Runtime verification iteration 2 — v3 lock-order/reclamation correction

After the first independent runtime review returned `REQUEST CHANGES`, the local v3 remediation moved
opportunistic cleanup after canonical exact acquisition/authority decisions and added deterministic
lock-order plus in-place reclamation proof.

The v3 local verification reran Stage 3.38 focused PostgreSQL tests, the Stage 3.38 HTTP retention
proof, Stage 3.28 auth containment/concurrency regressions, Stage 3.32 PostgreSQL replay regressions,
the full Go suite, `go vet`, OpenAPI/migration validators, race-enabled Stage 3.38 PostgreSQL tests,
and migration down/reapply evidence.

No commit, push, runtime PR, or merge was created by that verification. A renewed independent
pre-commit runtime review is still required for the exact v3 patch.

### Runtime verification iteration 3 — v4 mixed-version/auth-order correction

After the renewed v3 review returned `REQUEST CHANGES`, v4 added post-unique-wait admission finalization,
moved auth cleanup after broader family/user session mutations, repaired the evidence record, and added
deterministic mixed-version unique-conflict plus user-wide auth ordering regressions.

The v4 export script records PASS only after focused Stage 3.38 PostgreSQL/HTTP/OpenAPI checks, Stage
3.28 and Stage 3.32 regression suites, full Go tests, `go vet`, race-enabled Stage 3.38 PostgreSQL tests,
migration down/reapply proof, and focused documentation consistency checks all succeed.

No commit, push, runtime PR, CI run, or merge is created by this local verification.

## 15. Residual risk / limitations

- Opportunistic cleanup has no idle-database deletion deadline.
- Backups, replicas, provider retention, cryptographic erasure, account deletion, and Stage 3.25
  privacy evidence remain separate.
- Mixed-version rollout can cause temporary extra serialization or fail-closed retry behavior; it must
  not be treated as proof that old instances implement Stage 3.38 semantics.
- Newly issued session TTL calculation remains service-owned as before; Stage 3.38 changes only the
  authority decision for the presented existing row.
- This stage does not change the browser retry journal implementation.

## 16. Operational / deployment consequences

Apply migration `000007_stage_03_38_operational_retention` before the new runtime. It performs no data
rewrite. Relevant mutating requests may delete at most 128 expired technical rows per table and
transaction. `SKIP LOCKED` permits multi-instance cleanup progress. No new secret, env var, service,
queue, Redis dependency, or runtime privilege is required. Immutable ledger and audit-event tables
remain outside cleanup.

## 17. Exact evidence

Planning evidence:

- planning canonical base before PR #93: `305a53bb07136b274717ff48778a5e93d7b1607c`;
- Stage 3.38 planning PR: #93;
- exact planning head: `7a4ef7115b5fbab4c9017c6032112f028825c959`;
- planning squash merge / runtime base: `a944f1e5d5ee7d84db5393e8760eda254d732edd`.

Runtime evidence:

- independently approved local DOCFIX4 patch SHA256:
  `7c114a0ec845505bc9a3dabf9ee8d491243db058ab9b2394ebe3ce12dc168eb3`;
- runtime branch: `fix/stage-03-38-p3-05-retention-runtime`;
- exact runtime commit / published head: `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`;
- exact published tree: `4e3083517677eb75f0f2b6822e8c59cac208b03d`;
- runtime PR: #94;
- exact-head GitHub Actions CI: #268 / run `32913862780`;
- required CI result: 10/10 successful;
- fresh published-head independent review: `APPROVED`, P0/P1/P2/P3 = None;
- explicit human Ready + squash-merge authorization applied only to exact head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`;
- canonical runtime squash merge: `2df9946d77ee044a191a0422c8cccbbfe02dc7c9`;
- canonical read-back: `develop` pointed exactly at `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` after PR #94 merge.

No native GitHub review object is asserted for the external independent ChatGPT reviews; the evidence
claim is the preserved external independent verdict plus exact repository/PR/CI identifiers.

The runtime merge is canonical but does not itself close P3-05.

## 18. Final canonical status

P3-05 remains **OPEN** at this runtime-merged base.

The runtime remediation is canonical in `develop`, but closure requires a separately governed closure
package that synchronizes `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, this implementation record, and a
Stage 3.38 closure record.

P3-05 becomes CLOSED only after that exact closure package receives independent pre-commit closure
`APPROVED`, receives separate human commit/push authorization, is published through a closure PR,
passes exact-head closure CI, receives fresh independent published-head closure `APPROVED`, receives
separate explicit human squash-merge authorization, and is merged into `develop`.

Until then, the canonical original audit backlog remains P0=0 / P1=0 / P2=0 / P3=7:
P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. After canonical P3-05 closure, it becomes
P0=0 / P1=0 / P2=0 / P3=6: P3-04, P3-06, P3-07, P3-08, P3-09, and P3-10.
