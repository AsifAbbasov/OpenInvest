# Stage 3.38 — P3-05 Idempotency and Session Retention/Cleanup Closure

| Field | Value |
| --- | --- |
| Status | CLOSED / PR #95 closure governance squash-merged into `develop`; P3-05 CLOSED |
| Date | 2026-08-26 |
| Finding | P3-05 — idempotency/session retention and cleanup |
| Planning gate | PR #93 squash-merged at `a944f1e5d5ee7d84db5393e8760eda254d732edd` |
| Runtime PR | PR #94 — `fix: implement Stage 3.38 P3-05 retention cleanup` |
| Frozen published runtime head | `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` |
| Published runtime tree | `4e3083517677eb75f0f2b6822e8c59cac208b03d` |
| Runtime merge | `2df9946d77ee044a191a0422c8cccbbfe02dc7c9` |
| Exact-head runtime CI | GitHub Actions CI #268 / run `32913862780`, 10/10 required jobs successful |
| Pre-commit approved patch SHA256 | `7c114a0ec845505bc9a3dabf9ee8d491243db058ab9b2394ebe3ce12dc168eb3` |
| Closure pre-commit approved patch SHA256 | `02c1b7a6dc7d6b8fa05be1f80af67a981737a20e50943096b6aef6e24fdb655b` |
| Published closure PR/head | PR #95 final published head `25eb3b9c3c153672f22a6718a7815a5d3c527f44`; squash-merged into `develop` at `c5962fa09b6d7d145dda203dbdf90069de7b1fcc` |
| Exact-head closure CI | Final PR #95 head `25eb3b9c3c153672f22a6718a7815a5d3c527f44`: GitHub Actions CI #271 / run `32961508562`, 10/10 required jobs successful |
Canonical record: commit(s) `8f5d10a3e7d138b69f59531f6e8875add6c7e766`, `25eb3b9c3c153672f22a6718a7815a5d3c527f44`, `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`.
| Closure merge authorized here | No — authorization remained a separate explicit human gate; that gate was subsequently satisfied and the actual PR #95 squash merge is recorded above |
| Final finding status | P3-05 CLOSED through actual PR #95 squash merge `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`; remaining original audit backlog P3=6: P3-04, P3-06, P3-07, P3-08, P3-09, P3-10 |

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

- global cleanup before exact command acquisition, creating a lock-order inversion/deadlock risk;
- insufficient deterministic proof of in-place expired-generation reclamation.

for:
- no-row admission timestamp sampled before a potentially blocking mixed-version UNIQUE conflict;
- auth cleanup before broader family/user updates, leaving a cross-user deadlock cycle;
- contradictory durable review/iteration documentation.


After publication, the exact GitHub head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508` received fresh independent published-head
`APPROVED` after direct PR diff, blob identity, base/head, and CI verification. No new P0/P1/P2/P3
blocker was found.

Canonical record: PR #95; commit(s) `8f5d10a3e7d138b69f59531f6e8875add6c7e766`.

the original published-vs-local contradiction was fixed, but found a new publication-stability defect:
active section-17/status wording said the remediation was `uncommitted` / `remediation pending`, which
would become false immediately if the exact candidate were committed/pushed. This is P3 governance/evidence
integrity only. The correction replaces ephemeral state assertions with immutable lifecycle events and rules
that remain truthful before and after publication; runtime/code/config scope remains unchanged.


## 14. Remediation iterations

Canonical record: PR #93, PR #94, PR #95; commit(s) `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`, `2df9946d77ee044a191a0422c8cccbbfe02dc7c9`, `8f5d10a3e7d138b69f59531f6e8875add6c7e766`.

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
- Closure pre-commit candidate patch SHA256: `02c1b7a6dc7d6b8fa05be1f80af67a981737a20e50943096b6aef6e24fdb655b`; independently `APPROVED`, P0/P1/P2/P3 = None.
- Closure commit / first published head: `8f5d10a3e7d138b69f59531f6e8875add6c7e766` with parent `2df9946d77ee044a191a0422c8cccbbfe02dc7c9`.
- Historical first published closure PR state: #95 was Draft/OPEN and not merged at first head `8f5d10a3e7d138b69f59531f6e8875add6c7e766`, base `develop`.
- Historical initial closure exact-head CI: #270 / run `32950023896` on `8f5d10a3e7d138b69f59531f6e8875add6c7e766`; 10/10 required jobs successful.
  governance/evidence-integrity blocker for stale active local/uncommitted lifecycle wording.
- Historical authorization state for the first published closure head: Ready authorization not granted; squash-merge authorization not granted; closure merge not performed at that point.
Canonical record: PR #95.
  integrity blocker for self-invalidating active `uncommitted` / `remediation pending` wording.
- Failed first remediation candidate identity: incremental patch SHA256
  `9cb5887a09508282244eabd0f2329fdc0befce251144d9f9aa7900737db35eff`; prospective full PR patch SHA256
  `7956ac8939eb09c3a655c086997109e0f8ae51938e334a840e8d90e64ffebce1`; verification report SHA256
  `8cbd200b0822650071f0735ee0b2e57ca4e867ef2d1550db60a7f2b9a7ede96a`.
- First publication-stable local semantic verification: FAILED before commit/push with `ERROR: publication-stable section-17 rule missing`. The rule was present but split by Markdown line wrapping; the checker incorrectly required a contiguous single-line literal. This is preserved as a verifier false negative. The corrected verifier compares normalized whitespace and still rejects the forbidden active self-invalidating phrases.

- Final closure remediation head: `25eb3b9c3c153672f22a6718a7815a5d3c527f44`.
- Final exact-head closure CI: #271 / run `32961508562`; 10/10 required jobs successful.
- Separate explicit human Ready + squash-merge authorization: satisfied for the final closure head.
- Actual PR #95 squash merge / canonical post-closure base: `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`.

## 18. Final canonical status rule

P3-05 is **CLOSED**.

Canonical record: PR #95; commit(s) `25eb3b9c3c153672f22a6718a7815a5d3c527f44`, `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`.

The canonical original audit count after closure is:

- P0: 0
- P1: 0
- P2: 0
- P3: 6

The remaining findings are P3-04, P3-06, P3-07, P3-08, P3-09, and P3-10. Historical references above to P3-05 being OPEN, PR #95 being unmerged, CI #270, or earlier authorization states are preserved only as chronology of prior candidates/reviews and do not override this current status. Stage 3.25 remains separate.
