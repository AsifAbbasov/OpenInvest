# Stage 3.38 — P3-05 Idempotency and Session Retention/Cleanup Closure

| Field | Value |
| --- | --- |
| Status | Closure candidate / independent closure review pending |
| Date | 2026-08-26 |
| Finding | P3-05 — idempotency/session retention and cleanup |
| Planning gate | PR #93 squash-merged at `a944f1e5d5ee7d84db5393e8760eda254d732edd` |
| Runtime PR | PR #94 — `fix: implement Stage 3.38 P3-05 retention cleanup` |
| Frozen published runtime head | `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` |
| Published runtime tree | `4e3083517677eb75f0f2b6822e8c59cac208b03d` |
| Runtime merge | `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` |
| Exact-head runtime CI | GitHub Actions CI #268 / run `32913862780`, 10/10 required jobs successful |
| Pre-commit approved patch SHA256 | `7c114a0ec845505bc9a3dabf9ee8d491243db058ab9b2394ebe3ce12dc168eb3` |
| Closure merge authorized here | No |

## 1. Finding / symptom

P3-05 tracked operational-retention debt across `investment.command_deduplication` and
`identity.sessions`. Both surfaces persisted expiry metadata, but command replay/conflict authority,
session containment/revocation authority, and bounded physical cleanup did not consistently honor the
retention boundary.

The defect was lifecycle/operational. It did not demonstrate a financial arithmetic, BusinessDate,
Decimal, ledger, or snapshot correctness failure.

## 2. Root cause

The pre-remediation command path persisted a nominal 24-hour expiry without making expiry authoritative
for conflict/replay and exact-key reuse. Read-only replay did not reject expired artifacts. Auth
refresh/logout decisions could rely on timestamps sampled before PostgreSQL serialization, allowing a
request to cross expiry while still using stale authority. No bounded production cleanup path or
expiry-leading global indexes existed for the two mutable technical tables.

## 3. Failure scenario

The reviewed failure scenarios were:

1. an idempotency key remaining replay/conflict-authoritative beyond the browser retry horizon;
2. a command or session request starting before expiry, blocking on database serialization, crossing
   expiry, and still using stale pre-lock time;
3. an expired revoked refresh token retaining family/user containment authority;
4. opportunistic cleanup acquiring rows in an order that could deadlock with exact-key or broader
   user/family session mutation locks;
5. a mixed-version/non-cooperating UNIQUE writer delaying fresh command admission while a provisional
   timestamp incorrectly shortened the new 24-hour generation;
6. indefinite growth of command/session technical rows under sustained traffic.

## 4. Impact

Impact was bounded to API lifecycle consistency, stale-session security authority, operational table
growth, and primary-database technical-retention hygiene.

No P0/P1/P2 impact was established. The final runtime reviews found no remaining P3 runtime blocker.

## 5. Severity rationale

P3 remained appropriate because the issue was lifecycle/operational rather than a demonstrated
authentication bypass, privilege escalation, financial corruption, or data-loss defect. The
security-sensitive portion was stale revocation/containment authority after credential expiry.

## 6. Existing guarantees violated

The old runtime violated or left incomplete:

- the 24-hour browser/server retry alignment;
- exact replay retention boundary;
- "expired credentials have zero current authority";
- deterministic post-expiry exact-key reclamation;
- bounded maintenance of mutable technical retention rows;
- preservation of revoked-but-unexpired Stage 3.28 containment without indefinite post-expiry
  authority.

## 7. Considered solutions

Considered alternatives included unbounded deletion, cleanup-only correctness, permanent replay,
immediate deletion of revoked sessions, permanent revoked-session retention, a worker/cron service,
and expiry-aware logical semantics plus bounded opportunistic cleanup.

For command serialization the design considered a conflict/retry loop, expiry-aware upsert,
transaction-scoped exact-scope advisory serialization plus uniqueness, or equivalent deterministic
serialization.

## 8. Chosen remediation

The canonical runtime implements:

- exact-scope transaction advisory serialization for command admission/reclamation;
- fresh PostgreSQL wall-clock authority after required serialization;
- inclusive expiry: `decision_time >= expires_at` means expired;
- fresh UUID/hash/timestamps and cleared replay/terminal fields for post-expiry command generations;
- no-row admission timestamp finalization only after a potentially blocking UNIQUE-conflict INSERT
  actually wins;
- read-only replay expiry using fresh DB wall clock and no hidden cleanup;
- session expiry authority only after existing user/session serialization locks;
- zero containment/revocation authority for expired presented sessions;
- preservation of revoked-but-unexpired Stage 3.28 containment;
- bounded cleanup batches of exactly 128 ordered by `(expires_at,id)` using
  `FOR UPDATE SKIP LOCKED`;
- cleanup inside the triggering mutation transaction, with cleanup failure rolling back that mutation;
- additive expiry-leading indexes through migration `000007_stage_03_38_operational_retention`;
- OpenAPI publication of the 24-hour server command/replay retention guarantee.

## 9. Why this solution

Correctness no longer depends on cleanup scheduling. Logical authority is decided independently of
physical row deletion. The chosen ordering preserves one lock direction for command/auth mutation and
maintenance, while `SKIP LOCKED` prevents waiting on already locked global cleanup candidates.

The remediation remains inside the existing modular monolith and requires no worker, cron, queue,
Redis path, Kubernetes component, provider dependency, or privilege expansion.

## 10. Rejected alternatives

- Unbounded `DELETE` — excessive lock/latency risk.
- Cleanup-only correctness — request semantics would depend on cleanup timing.
- Permanent replay — contradicts the approved 24-hour lifecycle.
- Immediate revoked-session deletion — destroys replay evidence before credential expiry.
- Permanent revoked-session retention — gives expired evidence indefinite operational lifetime.
- Timestamp sampled before blocking — stale across expiry-straddling waits.
- New worker/cron/queue — unnecessary architecture expansion for P3-05.
- Privilege expansion — unnecessary; existing runtime role is sufficient on the two mutable technical
  tables and remains denied forbidden ledger/audit deletion.

## 11. Trade-offs

Each relevant command reservation or auth mutation may perform one bounded cleanup pass. Cleanup can
extend lock hold time, but batch size is fixed, expiry-leading indexes exist, and `SKIP LOCKED`
prevents maintenance from waiting on locked global candidates.

Cleanup has no idle-database deletion SLA. Advisory-hash collisions may serialize unrelated command
scopes but cannot cross-authorize data because exact unique/predicate identity remains authoritative.

## 12. Regression tests

The merged runtime proves:

- equality-at-expiry is expired;
- pre-expiry replay/conflict behavior remains;
- expired replay lookup returns not found without writes;
- exact-key fresh generation after expiry;
- deterministic clearing of old generation response state;
- concurrent post-expiry retry converges on one new effect/artifact;
- command race straddling expiry cannot mix generations;
- refresh/logout blocked across expiry has zero stale authority;
- cleanup batch/index contract;
- concurrent cleanup `SKIP LOCKED`;
- runtime-role cleanup without privilege expansion and preservation of audit;
- command cleanup after exact-row acquisition;
- distinct expired reservations complete without deadlock;
- auth cleanup after presented-session serialization;
- mixed-version UNIQUE-conflict admission timestamp is post-wait;
- auth cleanup occurs after broader user-wide row locks;
- two-user allSessions mutations complete without cleanup deadlock;
- expired import proof plus missing replay cannot authorize a second financial append;
- Stage 3.28 and Stage 3.32 regressions remain green;
- race-enabled Stage 3.38 suite remains green.

## 13. Adversarial review findings

The full review history is deliberately preserved.

Local runtime iteration 1 exposed PostgreSQL `SQLSTATE 42P08` parameter-type ambiguity before any
commit/push. That was corrected with explicit `timestamptz` typing.

First independent v2 pre-commit review returned `REQUEST CHANGES` for:
- global cleanup before exact command acquisition, creating a lock-order inversion/deadlock risk;
- insufficient deterministic proof of in-place expired-generation reclamation.

The v3 remediation closed those items, but renewed independent v3 review returned `REQUEST CHANGES`
for:
- no-row admission timestamp sampled before a potentially blocking mixed-version UNIQUE conflict;
- auth cleanup before broader family/user updates, leaving a cross-user deadlock cycle;
- contradictory durable review/iteration documentation.

The v4 runtime remediation closed all three substantive issues and added deterministic concurrency
regressions. Subsequent exact-v4/DOCFIX2/DOCFIX3 reviews found documentation/evidence-only defects:
a stale iteration assertion, a misquoted rejected sentence, and a stale hard-coded review count.
DOCFIX4 closed those evidence defects and received final independent pre-commit `APPROVED` with
P0/P1/P2/P3 = None.

After publication, the exact GitHub head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` received fresh independent published-head
`APPROVED` after direct PR diff, blob identity, base/head, and CI verification. No new P0/P1/P2/P3
blocker was found.

No native GitHub review object is claimed for these external independent ChatGPT reviews.

## 14. Remediation iterations

1. Planning PR #93 defined retention, batching, logical expiry, and concurrency boundaries.
2. Planning review exposed pre-serialization clock authority; the plan was corrected before merge.
3. Runtime v2 initially failed locally with SQLSTATE `42P08`; SQL typing was corrected.
4. v2 independent review exposed command cleanup lock order and reclamation-proof gaps.
5. v3 remediated those and passed full local verification.
6. Renewed v3 review exposed mixed-version admission timing, auth cleanup ordering, and evidence drift.
7. v4 remediated runtime timing/ordering and added deterministic regressions.
8. Exact-v4/DOCFIX2/DOCFIX3 documentation-only review cycles repaired durable evidence without changing
   runtime behavior.
9. DOCFIX4 passed full local verification and independent pre-commit `APPROVED`.
10. The exact approved candidate was committed/pushed as `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`.
11. Draft PR #94 passed exact-head CI #268 / `32913862780` 10/10.
12. Fresh published-head independent review returned `APPROVED`.
13. The user separately authorized Ready + squash merge.
14. PR #94 was squash-merged as `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` and `develop` was read back at that SHA.
15. This closure package synchronizes canonical governance state without changing runtime behavior.
16. The first local closure semantic verification failed before commit/push because the generated closure record did not contain an explicit active `P3-05 remains OPEN` sentence, even though the conditional closure rule was otherwise present. This documentation-only correction adds that unambiguous OPEN statement and reruns the complete closure semantic verification.
17. The second local closure semantic verification also failed before commit/push, this time because the checker compared the post-closure remaining-P3 set as one literal single-line string. ROADMAP already contained the correct set, but normal Markdown line wrapping split `P3-09` and `and P3-10` across a newline. The candidate content was not semantically wrong; the verifier was. The checker is corrected to compare normalized whitespace while still requiring the exact same six finding IDs.
18. The third local closure semantic verification failed before commit/push because the checker looked for the current-backlog sequence `P3-05, P3-06, ...` anywhere in the whole ROADMAP and mistook the intentionally preserved current P3=7 backlog for a bad post-closure forecast. The document was correct: current state still includes P3-05, while only the post-closure forecast removes it. The verifier is corrected to validate current-state and forecast clauses separately.
19. The fourth local closure semantic verification failed before commit/push because the closure record stated the current count as P3=7 but did not enumerate the exact current seven-finding set. ROADMAP and SOURCE_OF_TRUTH already preserved that set, but the closure record itself was less explicit than the verifier required. This is treated as a documentation-completeness gap rather than hidden by weakening the check: the exact current set is now stated in section 18 and the verification is rerun.
20. The fifth local closure semantic verification failed before commit/push because the checker required the literal token `P3=7` in ROADMAP, while ROADMAP expresses the canonical count in its existing bullet form `P3: 7`. The exact current seven-finding set was already present and correct. This was a verifier false negative, not a repository-state defect. The checker is corrected to validate the count semantically (`P3=7` or `P3: 7`) and to require the exact current seven-finding set independently.

## 15. Residual risk / limitations

- Opportunistic cleanup has no idle-database deletion deadline.
- Backups, replicas, provider retention, cryptographic erasure, and account deletion remain separate.
- Mixed-version rollout can cause temporary extra serialization/fail-closed retry behavior; old
  instances are not thereby proven to implement Stage 3.38 semantics.
- Newly issued session TTL calculation remains service-owned as before; Stage 3.38 changes authority
  for the presented existing row.
- The browser retry journal implementation is unchanged.
- P3-08 migration-policy hardening remains separate.

## 16. Operational / deployment consequences

Migration `000007_stage_03_38_operational_retention` adds only the two `(expires_at,id)` indexes and
performs no backfill/data rewrite. Relevant mutating requests may delete at most 128 expired technical
rows per table and transaction.

No new secret, env var, service, worker, scheduler, queue, Redis dependency, Kubernetes resource,
provider integration, or runtime privilege is required. Immutable ledger, snapshots, and audit events
remain outside cleanup.

## 17. Exact evidence

- Planning canonical base: `305a53bb07136b274717ff48778a5e93d7b1607c`.
- Planning PR: #93.
- Planning exact head: `7a4ef7115b5fbab4c9017c6032112f028825c959`.
- Planning squash merge: `a944f1e5d5ee7d84db5393e8760eda254d732edd`.
- Independently approved final local runtime patch SHA256: `7c114a0ec845505bc9a3dabf9ee8d491243db058ab9b2394ebe3ce12dc168eb3`.
- Runtime branch: `fix/stage-03-38-p3-05-retention-runtime`.
- Exact runtime commit / published head: `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`.
- Exact published runtime tree: `4e3083517677eb75f0f2b6822e8c59cac208b03d`.
- Runtime PR: #94.
- Exact-head CI: #268 / run `32913862780`; 10/10 required jobs successful.
- Independent final local pre-commit review: `APPROVED`, P0/P1/P2/P3 = None.
- Fresh exact published-head review: `APPROVED`, P0/P1/P2/P3 = None.
- Separate explicit human Ready + squash-merge authorization: yes, exact head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` only.
- Canonical runtime squash merge: `2df9946d77ee044a191a0422c8cccbbfe02dc7c9`.
- Canonical branch read-back: `develop` pointed exactly at `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` after PR #94.
- Closure package: this document plus synchronized `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, and the Stage
  3.38 implementation record.
- Initial closure-candidate local semantic verification: FAILED before commit/push because the draft lacked the explicit active sentence `P3-05 remains OPEN`; corrected in-place and reverified.
- Second closure-candidate local semantic verification: FAILED before commit/push because the verification script required the remaining P3 set as one literal line; ROADMAP had the correct six-item set split by Markdown line wrapping. The checker was corrected to normalize whitespace.
- Third closure-candidate local semantic verification: FAILED before commit/push because the checker globally rejected the valid current-backlog `P3-05, P3-06, ...` sequence instead of scoping the check to post-closure forecast prose. The verifier was corrected to validate current and forecast states separately.
- Fourth closure-candidate local semantic verification: FAILED before commit/push because the closure record gave the current count P3=7 without enumerating the exact current seven-finding set. Section 18 now states that set explicitly and the verification is rerun.
- Fifth closure-candidate local semantic verification: FAILED before commit/push because the checker required literal `P3=7` in ROADMAP even though ROADMAP correctly used its canonical bullet form `P3: 7`. The verifier now accepts the repository's existing count notation while still requiring the exact current seven-finding set.
- Closure PR/head/CI/merge: not asserted yet; this is a local uncommitted closure candidate.

## 18. Final canonical status rule

P3-05 remains **OPEN** in this unmerged closure candidate.

The exact current P3=7 finding set is: P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10.

At the base of this closure candidate, the canonical original audit count remains:

- P0: 0
- P1: 0
- P2: 0
- P3: 7

P3-05 runtime remediation is canonical in `develop`, but this unmerged closure package does not itself
declare the finding canonically CLOSED.

P3-05 becomes CLOSED only when this exact closure package receives fresh independent pre-commit
closure `APPROVED`, receives separate human commit/push authorization, is published in a closure PR,
passes exact-head green closure CI, receives fresh independent published-head closure `APPROVED`,
receives separate explicit human squash-merge authorization, and the closure PR is merged into
`develop`.

The resulting post-closure backlog is P0=0 / P1=0 / P2=0 / P3=6, consisting of P3-04, P3-06, P3-07,
P3-08, P3-09, and P3-10. Stage 3.25 remains separate.
