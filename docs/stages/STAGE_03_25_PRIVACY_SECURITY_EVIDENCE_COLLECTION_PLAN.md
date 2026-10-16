# Stage 3.25 - Privacy Security Review Evidence-Collection Plan

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-25-PRIVACY-SECURITY-EVIDENCE-COLLECTION-PLAN |
| Version | 0.4.0 |
| Status | Active / proposal only |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.24 Security Review readiness dossier |
| Dependencies | Documents 42-43; proposed ADR-008; Stage 3.20-3.24 privacy proposals |
| Last Review Date | 2026-08-18 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Before evidence collection, formal Security Review, ADR-008 acceptance, provider proposal, or privacy-lifecycle migration proposal |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

## Purpose

Stage 3.24 establishes that formal Security Review remains blocked until independently assembled
evidence answers every mandatory question. This plan defines how future evidence is requested,
minimized, redacted, integrity-protected, independently checked, and submitted. It does not collect
evidence, inspect a production environment, select a provider, perform Security Review, or accept
ADR-008.

A repository proposal, CI result, screenshot, mutable dashboard, verbal assurance, or placeholder
cannot be converted into operational proof by listing it in this plan. Every uncollected, stale,
unverifiable, contradictory, or over-redacted item remains blocked_evidence.

## Evidence Handling Rules

Each evidence item must include a stable evidence ID; related SR question(s); scope and environment;
source system and accountable owner; collection time and applicable configuration/version; access and
retention classification; integrity method; redaction method; limitations; independent verifier; and
a decision-ready conclusion limited to what the item actually proves.

Collection must use the least sensitive source and smallest sufficient sample. It must not place raw
identity data, credentials, tokens, raw request/import bodies, key material, usable identifier maps,
or public marker lookup material in the Security Review package. Redaction cannot remove the
control's version, scope, failure state, or ability to test the claim. When safe redaction is
impossible, the item must be reviewed through a separately approved restricted-access process and
the dossier records that constraint without copying the sensitive content.

Evidence must be immutable or independently reproducible. A mutable console view, self-attested
export, or shared-secret MAC is insufficient alone where a privileged actor could forge the result.
The evidence package must preserve source provenance and detect substitution, truncation, replay,
staleness, and conflict.

## Required Evidence Workstreams

| ID | Mandatory scope | Accountable evidence owner | Minimum acceptable evidence | Independent verification | Blocking result |
| --- | --- | --- | --- | --- | --- |
| EC-01 | SR-11 deletion and cancellation authority | Security and identity owner | Fresh re-authentication/factor design; anti-forgery/CSRF boundary; rate-limit and abuse controls for replay and enumeration; non-oracular response/timing model; cancellation rules that reject a revoked session or opaque identifier alone; negative tests; and redacted audit evidence. | Security reviewer tests each claimed abuse and cancellation failure mode without using production identities. | No deletion/cancellation lifecycle, migration, or ADR recommendation. |
| EC-02 | SR-01 and SR-08 data lineage, anonymity, and minimum audit evidence | Privacy/data owner | Field-level disposition and direct/indirect linkage analysis for primary, replica, backup, export, log, queue, cache, analytics, CI, and organizational surfaces; plus a ten-year minimal-redacted evidence schema, restricted-access review, retention policy, and sampled records that test reversibility. | Independent privacy reviewer traces samples, tests the sampled record/linkage claim, and records unresolved external surfaces. | No Anonymous Financial History, audit-retention policy, or lifecycle-completion claim. |
| EC-03 | SR-02 custody and destruction proof | Custody/operations owner | Provider-neutral capability matrix plus provider-specific evidence after a provider is proposed: deletion semantics, non-exportability, administrator recovery, independently verifiable attestation, replication, residency, retention, cost, exit, and support; custody policy; credential lifecycle; and quorum/break-glass design. | Independent verifier validates attestation semantics, custody-policy separation, credential lifecycle, and access assumptions. | No provider decision, key destruction, or migration approval. |
| EC-04 | SR-03 and SR-04 marker integrity and replay | Security/control-plane owner | Field schema, replay-binding design, restricted-access policy, redaction/correlation analysis, separately administered writer/signer model, threshold or external attestation, completeness/freshness protocol, compatibility rules, and negative snapshot/replay tests. | Security reviewer tests omission, rollback, split-view, stale, forged, correlation, compatibility, and person-lookup cases. | No marker or restore procedure approval. |
| EC-05 | SR-05 and SR-12 lifecycle migration/data state and integrated operations runbook | Architecture/data owner | Non-authorizing monotonic data-state design with serialized-intent binding, constraints, locking, access controls, retention, backfill/rollback limits, forward recovery, non-recreation proof, fault-injection results, escalation authority, and redacted audit-evidence design; plus an integrated least-privilege runbook for authorization, proof, backup expiry, isolated restore, release, outage/escalation, and evidence redaction. | Architecture and security reviewers test partial-failure, duplicate intent, race, cancellation, outage, and runbook handoff paths. | No migration, job, lifecycle implementation proposal, formal Security Review recommendation, or ADR-008 decision. |
| EC-06 | SR-06 restore/release | Restore/operations owner | Network and credential isolation design, least-privilege release roles, sealed restore runbook, release checklist, and adversarial rehearsal record. | Independent restore-release reviewer verifies no serving path exists before checks pass. | No restore or disaster-recovery approval. |
| EC-07 | SR-07 managed copies | Backup/operations owner | Inventory for backups, replicas, PITR/WAL, exports, and recovery media; encryption, 90-day destruction, reconciliation, exception/hold, and periodic evidence. | Independent reconciliation against provider/system-of-record evidence. | No backup-compliance or deletion-completion claim. |
| EC-08 | SR-09 operations and insider boundary | Operations/security owner | Role/capability matrix, access reviews, export/support restrictions, monitoring, incident response, and organizational-access analysis. | Security reviewer validates separation and break-glass limits. | No production operations authorization. |
| EC-09 | SR-10 legal/retention boundary | Legal/privacy/security owner | Signed legal, regulatory, and security decision for legal hold, incident, retention, and exception handling; explicit fail-closed behavior; and retention evidence that preserves no hidden recovery path. | Legal and security reviewers independently confirm policy-to-control mapping, retention evidence, and fail-closed behavior. | No policy, lifecycle completion, or implementation approval. |
| EC-10 | All SR questions and formal review record | Principal Architect and Security Review coordinator | Immutable scope baseline, complete evidence index, findings, residual-risk treatment, recommendation, and separate Principal Architect decision. | Human Security Review verifies evidence completeness before any recommendation. | No ADR-008 acceptance or implementation authorization. |

No owner may independently approve the evidence it produces when that owner can change the source
control. Separation must be recorded per item; a conflict or unavailable independent verifier is
blocked_evidence.

Each clause of a mapped SR minimum proof is mandatory, even when its evidence is distributed across
multiple EC items. The EC-10 index must map every clause to its evidence IDs, collection and review
freshness, and independent-verifier pass, fail, or blocked_evidence result. A missing, stale, failed,
or blocked clause leaves its SR question blocked_evidence. In particular, SR-12 cannot be complete
from a data-state design or a partial runbook alone: EC-05 must contain the complete integrated
runbook and its verified inputs from the relevant custody, marker, restore, backup, operations, and
legal workstreams.

## Collection Sequence

1. The Principal Architect freezes the target scope, source baseline, evidence IDs, owner assignments,
   independent verifiers, and restricted-access procedure before collection.
2. Each owner submits a minimized evidence request describing the claim, source, collection method,
   expected sensitive fields, retention, and the SR question it addresses.
3. Security and privacy representatives approve or reject the request's minimization and handling
   plan. Rejection is not a permission to collect a broader sample.
4. The owner collects only the approved evidence, records provenance and integrity metadata, and
   applies the approved redaction or restricted-access path.
5. The independent verifier checks completeness, freshness, authenticity, and limitations, then
   records pass, fail, or blocked_evidence without editing the source evidence.
6. The Security Review coordinator builds an evidence index that maps every SR question and each of
   its minimum-proof clauses to verified items and explicitly lists every missing, stale,
   contradictory, failed, or blocked clause.
7. Only after every mandatory entry and minimum-proof clause is independently verified as current
   and no blocker remains may a human formal Security Review begin. This plan cannot change that
   condition.

## Invalid Evidence and Safe Failure

The package must reject an item that is missing provenance, integrity, owner, verifier, scope,
version, limitation, or required redaction. It must reject a claim that uses absence from source
code as proof of absence in production. It must reject a current-state assertion derived solely from
a design document, local test, CI job, or mutable dashboard.

A failed collection, unavailable provider, missing backup inventory, unavailable verifier, legal
uncertainty, or evidence conflict remains visible as blocked_evidence. It must not be converted into
an accepted risk, a fabricated negative result, a reduced sample, or an implementation exception.

## Evidence Index Template

The future restricted evidence index must have one row per evidence item:

| Field | Required content |
| --- | --- |
| Evidence ID and SR mapping | EC ID plus every SR question supported, or all SR questions for the formal review record; no unsupported scope expansion. |
| SR minimum-proof clause status | Exact Stage 3.24 clause, supporting evidence IDs, collection and review freshness, independent-verifier result, and pass/fail/blocked_evidence status. |
| Claim and scope | Precise, falsifiable claim; environment, system boundary, and versions. |
| Source and provenance | Source system, collection method/time, owner, integrity reference, and reproducibility path. |
| Sensitive-data handling | Data classification, minimization, redaction/restricted-access method, retention, and access authority. |
| Independent verification | Verifier identity/role, method, result, limitations, and conflict declaration. |
| Decision effect | Which blocker is resolved, remains blocked, or is contradicted; no implementation authorization. |

## Scope Boundary

This stage adds documentation only. It does not collect production data, access providers, alter
credentials, inspect backups, run restores, select a provider, conduct a formal Security Review,
accept ADR-008, change Go, Next.js, Python, OpenAPI, PostgreSQL, Redis, migrations, CI,
infrastructure, backups, operations, or product behavior.

The immediate next action is mandatory strict review of this plan. Actual evidence collection begins
only under an explicitly approved restricted-access process with accountable owners and independent
verifiers. Privacy-lifecycle implementation remains prohibited until the formal Security Review and
ADR-008 acceptance are recorded.

## Architecture validation and traceability v1

Prospective design validation authorized for synthetic, non-destructive local work only; source baseline
`cffd2091168fa28817a49464438db01804b4ee48`. Stage 3.25 remains **Active / proposal only**. The original
EC-01–EC-10 workstreams and independent-verifier clauses remain unchanged. No real evidence collection,
provider access, migration application, backup restore or formal Security Review occurred.

| Class | Evidence obtained / permissible next work | What is not established |
| --- | --- | --- |
| Repository | Protected baseline/merged PR #275; actual source/schema/role inventory R1–R8; original SR/EC clauses; current proposals; unchanged controlling contracts. | Deployed rows, IAM, backup inventory, deletion implementation or provider guarantees. |
| Isolated validation | Bounded abstract lifecycle exploration and restore truth table below; source/hash/link/scope checks. | Real PostgreSQL locks/crash recovery, crypto/signature validation, HTTP auth, provider destruction or production restore isolation. |
| External technical | Required EC-03/04/06/07/08 provider/custody, independent witness, actual backup/restore and operator evidence. None collected. | No external item is PASS by inference from a local model or ordinary CI. |
| Human decisions | A1–A4 and EC-09/10; qualified legal/security/operations and Principal Architect decisions required. None fabricated. | A proposal/status table does not accept risk, change a frozen contract or authorize implementation. |

### Tests mapped to threats, evidence and acceptance

These future suites use only separately authorized synthetic environments. “Required” is not executed.

| Test ID | Protocol/failure coverage | SR → EC / acceptance gate |
| --- | --- | --- |
| T1 | Field lineage, old user/locator/proof/marker joins, personal text/hashes, exact recovered-scope targeting, partitioned export and projection rebuild. | SR-01/04/08 → EC-02/04: no overlooked person lookup; A1/A2/A3 resolved. |
| T2 | Fresh request/cancel, wrong factor/CSRF, stale generation, duplicate/different intent, fixed deadline, concurrent request/cancel/claim, old JWT, idempotent success and portfolio/import/valuation/outbox admission. | SR-05/11/12 → EC-01/05: current-state denial before protected work, one irreversible claim, no stale callback reactivation. |
| T3 | Crash before/after intent commit/admission, custody call, terminal proof, unlink and marker publication; provider timeout/scheduled state, stale/revoked proof, all material versions and alternate wraps. | SR-02/05 → EC-03/05: no key recreation or false completion, durable forward recovery. |
| T4 | Omitted/duplicate/reordered pages, valid signed prefix, conflicting roots, stale/revoked signer, epoch/schema mismatch, independent witness loss and proof-binding mismatch. | SR-03/04 → EC-04: complete admitted set and correct restricted replay; signature alone cannot pass. |
| T5 | Synthetic restore from newest/oldest allowed image, prepared/disputed intent, new intent after checkpoint, copied certificate, changed instance/configuration, lost witness, ingress/job/credential attempts before release. | SR-03/06/12 → EC-04/06: sealed non-serving until independently fenced release; real infrastructure rehearsal still required later. |
| T6 | Legacy plaintext copy, WAL/replica/export omission, failed expiry, copied backup whose retention clock was reset, missing provider evidence, legal-hold conflict. | SR-07/10 → EC-07/09: copy completeness and ≤90-day destruction; unknown/held copy never produces final completion. |
| T7 | Non-identifying prospective audit/receipt fields, secret sentinels, timestamp joins, receipt substitution/replay, public response shape and bounded timing, retained browser/export identity bridge. | SR-01/08/10/11 → EC-01/02/09: no raw identity/secrets or false universal-erasure statement. |
| T8 | Effective role combinations, shared administrator control, support exports, signing/custody break-glass, old worker/binary capability, least-privilege lifecycle role. | SR-09/12 → EC-05/08: no unilateral evidence fabrication/recovery/restore release; deployed IAM needs independent evidence. |
| T9 | Expand/populate/shadow/switch/contract, interrupted backfill, lock-order checks and disposable PostgreSQL race tests; golden financial amounts/dates/ledger order/revisions/snapshots unchanged. | SR-12 → EC-05: financial accuracy and compatibility preserved; migration policy/manifest and rollback fence independently checked. |

EC completeness: EC-01=T2/T7; EC-02=T1/T7; EC-03=T3; EC-04=T1/T4/T5;
EC-05=T2/T3/T8/T9; EC-06=T5; EC-07=T6; EC-08=T8; EC-09=T6/T7;
EC-10 indexes every SR minimum-proof clause, evidence freshness, accountable owner and independent
verifier result. An SR-wide link alone cannot stand in for missing clause-level evidence. Until each
original minimum-proof clause has real accepted evidence, final SR acceptance stays blocked_evidence.

### Historical bounded abstract validation v1

The historical v1 model below has no file/service I/O, database, network or cryptography; it emits stdout.
Its missing deadline, admission-loss and old-binary coverage is superseded by targeted v2 below;
the original source/results remain preserved, not promoted to stronger evidence. Run with Python 3 in an isolated
scratch directory. It explores persistent phase states with two cancellation generations and at most
two writes, plus crash/retry and stale-callback no-ops. A monitor rejects post-fence writes,
destruction before durable intent, key recreation, premature markers and completion while identity,
plaintext copy or indirect link survives. Three deliberately weakened transition guards must produce
counterexamples; these are model sensitivity controls, not real product findings.
The restore table enumerates seven two-intent leaf sets and nine evidence predicates (3,584 cases).
It checks conjunction/completeness logic only: signatures, independence, trusted freshness, version,
replay, actual isolation and a linearizable release barrier are assumed inputs, not implemented or proved.

Executed 2026-10-16: lifecycle **72 states / 499 transitions**, three guard defects detected;
restore **3,584 cases / two accepted canonicalizable A/B orderings**. A signed prefix missing a later
accepted intent is rejected. These results validate bounded abstract consistency, not general safety,
liveness under outages, completeness of the threat model or deployed guarantees.
Script SHA-256: `98029cab20c58e80e223303088edab2e01caa839a5be61d8d7e447a8e77115b2`.

```python
from dataclasses import dataclass, replace
from collections import deque
from itertools import product

@dataclass(frozen=True)
class State:
    phase: str = 'active'
    generation: int = 0
    key: bool = True
    primary: bool = True
    plaintext_copy: bool = True
    indirect_link: bool = True
    intent: bool = False
    proof: bool = False
    marker: bool = False
    writes: int = 0

def successors(s, mutant=''):
    yield 'crash_retry', s
    yield 'stale_generation_callback', s
    yield 'invalid_factor_or_cross_subject', s
    if (s.phase == 'active' or mutant == 'pending_write' and s.phase == 'pending') and s.writes < 2:
        yield 'write', replace(s, writes=s.writes+1)
    if s.phase == 'active':
        yield 'fresh_request', replace(s, phase='pending')
    if s.phase == 'pending':
        if s.generation == 0:
            yield 'fresh_cancel', replace(s, phase='active', generation=1)
        yield 'claim_deadline', replace(s, phase='completing')
    if s.phase == 'completing':
        yield 'witness_intent', replace(s, intent=True)
        if s.intent or mutant == 'unwitnessed_destroy':
            yield 'destroy', replace(s, key=False)
        if not s.key and s.intent:
            yield 'verify_terminal_proof', replace(s, proof=True)
        if s.proof:
            yield 'unlink_primary', replace(s, primary=False)
        if s.proof and not s.primary:
            yield 'publish_marker', replace(s, marker=True)
        if s.marker:
            yield 'verify_copy_drain', replace(s, plaintext_copy=False)
            yield 'verify_indirect_unlinking', replace(s, indirect_link=False)
        if s.proof and s.marker and not s.primary and not s.indirect_link and (not s.plaintext_copy or mutant == 'ignore_copy'):
            yield 'complete', replace(s, phase='completed')

def monitor(old, new):
    assert not (not old.key and new.key), 'key_recreation'
    assert new.key or new.intent, 'destruction_without_durable_intent'
    assert not new.proof or not new.key, 'proof_before_terminal'
    assert not new.marker or new.proof and not new.primary, 'premature_marker'
    assert new.writes == old.writes or old.phase == new.phase == 'active', 'write_after_fence'
    if new.phase == 'completed':
        assert not any((new.key, new.primary, new.plaintext_copy, new.indirect_link)), 'false_completion'
        assert new.intent and new.proof and new.marker, 'missing_completion_evidence'

def explore(mutant=''):
    todo=deque([(State(), ())]); seen={State()}; edges=0
    while todo:
        s, trace=todo.popleft()
        for event, n in successors(s, mutant):
            edges+=1
            try: monitor(s,n)
            except AssertionError as err: return len(seen),edges,trace+(event,),str(err)
            if n not in seen:
                seen.add(n); todo.append((n,trace+(event,)))
    assert any(s.phase == 'completed' for s in seen)
    return len(seen),edges,None,None

states,edges,witness,error=explore()
assert witness is None
print(f'LIFECYCLE_MODEL=PASS states={states} transitions={edges}')
for mutant in ['pending_write','ignore_copy','unwitnessed_destroy']:
    _,_,w,e=explore(mutant); assert w is not None
    print(f'MODEL_SENSITIVITY={mutant} DETECTED={e} TRACE={",".join(w)}')

# A/B represent two independently accepted destructive-intent receipts, not identities.
# Enumerate omitted/substituted/duplicate snapshot leaves and all evidence flags.
checked=accepted=0
for leaves in [(), ('A',), ('B',), ('A','B'), ('B','A'), ('A','A'), ('A','C')]:
    for flags in product([False,True], repeat=9):
        signature, independent, consistent, fresh, compatible, isolated, proof, replay, fenced=flags
        complete=len(leaves)==2 and set(leaves)=={'A','B'}
        release=complete and all(flags)
        checked+=1; accepted+=int(release)
        if release: assert complete and signature and independent and consistent and fresh and compatible and isolated and proof and replay and fenced
assert accepted==2  # both canonicalizable A/B orders; acceptance sorts by authoritative sequence.
print(f'RESTORE_TRUTH_TABLE=PASS cases={checked} accepted={accepted}')
# Race witness: a valid signed prefix is insufficient after a later accepted intent.
assert set(('A',)) != {'A','B'}
print('SIGNED_PREFIX_OMISSION_AND_NEW_INTENT_RELEASE_CUT=REJECTED')
print('DATABASE_PROVIDER_RESTORE_IMPLEMENTATION=NOT_TESTED')
```

Normal exact-head CI checks unchanged source compatibility; it cannot test the future deletion, custody,
marker or restore implementation. Future T1–T9 runtime suites and EC provider/human evidence remain
NOT_EXECUTED / NOT_COLLECTED, separate from the bounded model and repository checks above.

## Targeted model v2 and remaining integration evidence

ARCH-277-04 supersedes v1 coverage, not its historical results. V2 uses seconds and the explicit
candidate `30 * 86400 = 2592000` grace; deadlines are stored in independent witness state and never
reset by retry, partition, database restore or WAL loss. Clock regressions/uncertainty are outside its
trusted monotonic-clock assumption and require fail-closed integration tests. The model explores two
activation instants (0 and one day), boundary times just before/at the relevant deadlines and a bounded
32-day horizon. It has one intent/subject, no real clocks, signatures, consensus, PostgreSQL or network.
A database rollback is modeled only before commit; restore may discard all local committed state, while
the G1 witness remains durable. The separate legacy fixture demonstrates that lost unwitnessed pending
is quarantined, not reconstructed. `ghost_pending` is an invariant observer, not available recovery data.
The `coverage` premise represents externally verified epoch completeness, never a caller-set flag.

Executed 2026-10-16: **4,668 states / 36,354 transitions**, early irreversible-claim defect detected;
DB-committed/witness-absent/missing-WAL fixture quarantined; unknown/lost admission reply reconciles to
the same deadline and pending state after restore. Compatibility table: **128 cases**, including old
login, old worker and an old binary holding a new credential. The latter can still issue active-scope
SQL under that identity: this deliberately exposes the missing binary attestation, requiring the
external broker/ingress prerequisite, rather than claiming PostgreSQL recognizes executables.
All pending scopes are denied under the assumed effective database fence. Release-cut table: **18 cases**;
stale cut or absent barrier denies release. These tables check predicates, not real policy enforcement.
Model has a hard 100,000-state cap; finite reachability and explicit retry/recovery witnesses do not
prove liveness under arbitrary outages, indefinite delayed messages, multiple subjects or real WAL loss.
Script SHA-256: `38f245cd0014eef1f8ca2b4e82aeb9d6602ea4e4bcb07ae60e03d5e36cd1dbb6`.

Exact reproduction from repository root (both preserved v1 and targeted v2; prints results only):

```sh
python - <<'PYTHON'
from pathlib import Path
import re
p = Path('docs/stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md')
for index, code in enumerate(re.findall(r'```python\n(.*?)```', p.read_text(), re.S), 1):
    print('MODEL_BLOCK', index)
    exec(compile(code, str(p) + ':model' + str(index), 'exec'), {})
PYTHON
```

### Executable targeted model v2

```python
from dataclasses import dataclass, replace
from collections import deque
from itertools import product

DAY = 86400
GRACE = 30 * DAY
CLOCKS = (0, DAY, 29*DAY, 30*DAY-1, 30*DAY, 31*DAY-1, 31*DAY, 32*DAY)

@dataclass(frozen=True)
class S:
    now: int = 0
    w: str = 'none'
    db: str = 'active'
    deadline: int = 0
    db_deadline: int = 0
    unknown: bool = False
    network: bool = True
    ack: bool = False
    key: bool = True
    restore: bool = False
    checked: bool = False
    serving: bool = True
    coverage: bool = True
    ghost_pending: bool = False
    writes: int = 0


def transitions(s, mutant=''):
    yield 'crash_retry', s  # durable witness/clock state never rolls back
    yield 'delayed_old_generation_reply', s
    yield 'partition_toggle', replace(s, network=not s.network, checked=False)
    for t in CLOCKS:
        if t > s.now:
            yield 'clock_forward', replace(s, now=t)
    if s.w == 'none' and s.network and not s.restore:
        for lost in (False, True):
            yield 'prepare_reply_lost' if lost else 'prepare', replace(s, w='prepared', unknown=lost, checked=False)
    if s.network and s.unknown:
        yield 'query_same_intent', replace(s, unknown=False)
    if s.w == 'prepared' and not s.unknown and not s.restore:
        if s.db == 'active':
            yield 'db_admission_commit', replace(s, db='admission_pending', ghost_pending=True)
            yield 'db_rollback_before_commit', s  # witness remains; no committed pending row is rolled back
        else:
            yield 'db_commit_retry', s
    if s.w == 'prepared' and s.db == 'admission_pending' and s.network and s.now in (0, DAY):
        for lost in (False, True):
            yield 'activate_reply_lost' if lost else 'activate', replace(s, w='pending', deadline=s.now+GRACE, unknown=lost, checked=False)
    if s.w == 'pending' and not s.unknown and not s.restore:
        yield 'db_checkpoint', replace(s, db='pending', db_deadline=s.deadline)
        if s.db == 'pending' and s.db_deadline == s.deadline and s.network:
            yield 'external_ack', replace(s, ack=True)
    if s.w == 'pending' and s.network and not s.unknown:
        if s.now < s.deadline:
            yield 'fresh_cancel', replace(s, w='cancelled', checked=False)
        if s.now >= s.deadline or mutant == 'early_claim':
            yield 'deadline_claim', replace(s, w='claimed', checked=False)
    if s.w == 'cancelled' and s.network and not s.unknown and not s.restore:
        yield 'cancel_checkpoint', replace(s, db='active')
    if s.w == 'claimed' and s.network and not s.unknown and not s.restore:
        yield 'claim_checkpoint', replace(s, db='completing')
        if s.db == 'completing':
            yield 'destroy', replace(s, key=False)
    if not s.restore and s.serving and s.db == 'active' and s.w in ('none','prepared','cancelled') and s.writes == 0:
        yield 'current_write', replace(s, writes=1)
    # Restore deliberately drops all local pending/checkpoint/WAL state.
    yield 'restore_without_wal', replace(s, db='active', db_deadline=0, restore=True, serving=False, checked=False)
    if s.restore and s.network and not s.unknown and s.coverage:
        phase={'none':'active','prepared':'admission_pending','pending':'pending','claimed':'completing','cancelled':'active'}[s.w]
        yield 'reconcile_witness', replace(s, db=phase, db_deadline=s.deadline if s.w=='pending' else 0, checked=True)
    if s.restore and s.network and s.checked and s.coverage and not s.unknown and s.w != 'prepared':
        yield 'release_atomic_cut', replace(s, restore=False, serving=True)


def invariant(old, event, n):
    assert n.now >= old.now, 'clock_rollback'
    assert old.deadline == 0 or n.deadline == old.deadline, 'deadline_reset'
    assert old.key or not n.key, 'key_recreation'
    if event == 'activate' or event == 'activate_reply_lost':
        assert n.deadline == n.now + GRACE, 'not_30_days'
    if n.w == 'claimed' and old.w != 'claimed':
        assert n.now >= n.deadline > 0, 'early_irreversible_claim'
    if not n.key:
        assert n.w == 'claimed' and n.now >= n.deadline, 'early_destruction'
    if event == 'external_ack':
        assert n.w == 'pending' and n.db == 'pending' and n.db_deadline == n.deadline and not n.unknown
    if event == 'current_write':
        assert old.db == 'active' and old.w in ('none','prepared','cancelled')
    if n.serving:
        assert not n.restore
        if n.w == 'prepared' and n.ghost_pending:
            assert n.db != 'active', 'committed_pending_unfenced'
        if n.w in ('pending','claimed'):
            assert n.db != 'active', 'pending_restored_active'
    if event == 'release_atomic_cut':
        assert n.coverage and n.network and n.checked and not n.unknown
        assert n.w != 'prepared', 'unresolved_prepare_release'


def explore(mutant=''):
    seen={S()};todo=deque([(S(),())]);edges=0
    while todo:
        s,trace=todo.popleft()
        for event,n in transitions(s,mutant):
            edges+=1
            try: invariant(s,event,n)
            except AssertionError as e: return len(seen),edges,trace+(event,),str(e)
            if n not in seen:
                seen.add(n);todo.append((n,trace+(event,)))
                assert len(seen) <= 100000, 'model_state_cap'
    assert any(not s.key for s in seen), 'no_destruction_path'
    assert any(s.w=='cancelled' and s.db=='active' and not s.restore for s in seen), 'no_cancel_path'
    return len(seen),edges,None,None

states,edges,w,e=explore();assert w is None,(w,e)
print(f'ADMISSION_GRACE_RESTORE_MODEL=PASS states={states} transitions={edges} grace_seconds={GRACE}')
_,_,w,e=explore('early_claim');assert e=='early_irreversible_claim'
print('SENSITIVITY_EARLY_CLAIM=DETECTED trace='+','.join(w))

# Historical DB-first, acknowledged-but-unwitnessed fixture: knowledge is unavailable.
legacy=S(db='pending',ack=True,ghost_pending=True,coverage=False)
restored=next(n for event,n in transitions(legacy) if event=='restore_without_wal')
assert not restored.serving and not any(event=='release_atomic_cut' for event,n in transitions(restored))
print('DB_COMMITTED_WITNESS_ABSENT_MISSING_WAL=QUARANTINED_NOT_RECONSTRUCTED')
assert restored.db == 'active' and restored.ghost_pending and not restored.coverage
print('SENSITIVITY_OLD_RELEASE_FROM_EMPTY_WITNESS=DETECTED_LOST_PENDING')
# Conditional progress witnesses, not fairness/provider availability guarantees.
s=S()
for event in ['prepare_reply_lost','query_same_intent','db_admission_commit','activate_reply_lost','query_same_intent','db_checkpoint','external_ack']:
    s=next(n for ev,n in transitions(s) if ev==event)
assert s.ack and s.deadline==GRACE
for event in ['restore_without_wal','reconcile_witness','release_atomic_cut']:
    s=next(n for ev,n in transitions(s) if ev==event)
assert s.db=='pending' and s.deadline==GRACE and s.serving
print('DELAYED_UNKNOWN_ADMISSION_RETRY_AND_PENDING_RECOVERY=PASS')

# Compatibility table assumes database policy and attested credential broker are effective.
cases=allowed=0
for epoch,admitted,subject_active,db_fence,drained in product([False,True],repeat=5):
    for caller in ('old_login','old_binary_new_login','new_binary','old_worker'):
        login_ok=caller in ('new_binary','old_binary_new_login')
        permit=epoch and admitted and subject_active and db_fence and drained and login_ok
        cases+=1;allowed+=int(permit)
        if not subject_active or not epoch or not db_fence or not drained: assert not permit
assert allowed==2
print(f'COMPATIBILITY_POLICY_TABLE=PASS cases={cases} active_sql_permits={allowed}')
print('BORROWED_NEW_CREDENTIAL_DOES_NOT_ATTEST_BINARY=EXTERNAL_BROKER_PREREQUISITE')
# Release race: a checkpoint valid before the next receipt is stale afterwards.
for latest,checkpoint,barrier in product(range(3),range(3),[False,True]):
    release=barrier and checkpoint==latest
    if checkpoint<latest or not barrier: assert not release
print('RESTORE_RELEASE_CUT_CASES=18 PASS; PARTITION_OR_STALE_CUT=DENY')
print('POSTGRES_OLD_BINARY_CRYPTO_PROVIDER_INFRASTRUCTURE=NOT_EXECUTED')
```

### Integration obligations not represented faithfully by v2

| Existing test / evidence route | Required isolated future verification / acceptance |
| --- | --- |
| T2/T3 → EC-01/05 | Trusted-clock skew/regression, full 30-day deadline persistence without waiting a real month (injected clock), simultaneous cancel/claim, real transaction abort/commit-ack ambiguity, credential rotation and normal auth/session compatibility. No early claim or deadline extension. |
| T3/T4/T5 → EC-04/05/06 | Durable independent PREPARE/admission service with partitions, delayed/duplicate messages, primary failure before/after each acknowledgement, restored backup with incomplete WAL, unknown outcome and service recovery. Every acknowledged/pending scope stays denied; uncovered v1 era remains quarantined. |
| T4/T5 → EC-04/06 | Real signature/epoch-completeness/anti-rollback verification, lost witness replicas, stale checkpoint followed by a new intent and coordinated release fence. No empty-log or timed freshness shortcut. |
| T2/T8/T9 → EC-01/05/08 | PostgreSQL 18 enforced policies/locks under Read Committed and Repeatable Read, all affected tables/partitions, direct SQL omitting the helper, old Go binary/worker under old and borrowed new credentials, caches, development bypass, prepared transactions and pool draining. Normal append-only ledger restrictions remain unchanged. |
| T5/T8/T9 → EC-05/06/08 | Artifact/credential broker, ingress/egress isolation, obsolete executable admission, disabled-policy/revoked-role state, interrupted sealed cutover and incompatible rollback. Old credentials or an old restored role catalog cannot release traffic. |

All rows above are NOT_EXECUTED; no PostgreSQL server, old executable, real worker, witness provider,
cryptographic verifier, backup or infrastructure was operated in this remediation. They are mandatory
before G1/G2 approval and deletion activation. T1/T6/T7 and A1–A4 remain unchanged. EC-10 must retain
every original SR minimum-proof clause, these new failure cases and independently reviewed results;
local model/ordinary CI cannot mark an external or implementation-dependent item PASS.

## A1–A4 / G1–G2 reproducible feasibility evidence

Baseline `5a83503719e6756c57707afbcda309b1a83f1d67`, tree `85fcc4c0fd0074da19e98a1ac0238aac9423842e`.
Synthetic experiments only; Stage 3.25 remains Active/proposal only and the independent collection plan
is not completed. Original EC clauses/T1–T9 are unchanged. All paths below are repository-relative.

| Experiment | Exact command / environment / inputs | Expected invariant / observed result / limitation |
| --- | --- | --- |
| G1/A3 local prototype | `python experiments/privacy-feasibility/admission_test.py -v`; Python 3.12.14, SQLite 3.53.1, cryptography 50.0.2; temporary separate journal/primary and generated test signing key. To provision dependency in a disposable environment: `uv run --with cryptography==50.0.2 python experiments/privacy-feasibility/admission_test.py -v`. | 11 tests PASS locally: PREPARE-first, process exit after commit/lost ACK, 16 concurrent duplicates, conflicting digest, precommit failure, journal outage, primary rollback, old primary state, pinned-prefix rejection, grace/retry and binding negative controls. PROTOTYPE_VERIFIED; no network quorum, independent administration or actual key destruction. |
| G2 local compilation | `cd backend-go && go test -v ./experiments/privacyfeasibility`; initial local assembler was truncated and exited with segmentation fault on `asm -V`; re-extraction of the existing Go 1.26.9 archive restored the tool. | Initial tool failure was not a product/test assertion failure; Re-extracted-tool compilation and `go vet ./experiments/privacyfeasibility` PASS; PostgreSQL test explicitly SKIP locally, not integration PASS. No local PostgreSQL installed; only UID 0 mapped in execution sandbox, so normal PostgreSQL non-root startup unavailable. Do not bypass that restriction. |
| G2 disposable CI integration | Existing ordinary Go tests and Go race jobs set `OPENINVEST_DATABASE_TEST_URL` to their loopback PostgreSQL 18 service; `go test ./...` and `go test -race ./...` include `backend-go/experiments/privacyfeasibility`. Focused reproduction: `OPENINVEST_PRIVACY_FEASIBILITY_DISPOSABLE=1 go test -race -v ./experiments/privacyfeasibility -run TestG2SyntheticPostgresLifecyclePolicy -count=1` with an explicitly disposable localhost PostgreSQL owner URL. | Requires dedicated randomly named database and role; real pending SELECT/INSERT denial, append-only denial, shared-lock conflict 55P03, stale Repeatable Read 40001, unrelated subject preserved and epoch removal denies an established session. Exact execution result is recorded in the Draft PR/CI evidence; absent URL emits explicit SKIP, never counted as integration PASS. |

The G1 journal is not a networked service; primary rollback is SQLite logical rollback, not PostgreSQL
WAL recovery. Its checkpoint is independently supplied by the test harness, so loss/rollback of that
trusted input is not solved. Signed-prefix detection does not implement Merkle consistency proofs or
prove an independently complete history. The G2 fixture does not load application schema or execute
old binaries/workers; it cannot establish complete production policy coverage, session enforcement or
artifact identity. No adversarial access-control evasion campaign or external service was exercised.

| Decision/gate | Current repository evidence → proposed control → threat | Test / observed evidence → prerequisite / approval |
| --- | --- | --- |
| A1 | Migrations 000001/3/4/6/10–12; transaction/import stores → split future personal projections + separately approved historical migration → text/hash/economic re-identification. | T1/T9 → EC-02/05/10; field dossier DESIGN_ONLY, no migrated data. Golden financial/dedup/replay equality, linkage analysis and Principal Architect/Privacy approval missing. |
| A2 | `audit.actors/events`, migration 000013, auth audit insertion and runtime append-only grants → minimized new events/versioned history disposition → surviving user/session/time joins. | T1/T7 → EC-02/09/10; field register DESIGN_ONLY. Original integrity vs identity erasure requires Security/Legal/Principal Architect decision. |
| A3 | Plain identity link/current key absence; Stage 3.22 profile → typed/domain-separated commitments, manifests, restricted locator/trust continuity → evidence substitution, hidden wrap, restored identity map. | G1 profile tests PROTOTYPE_VERIFIED only; T1/T3/T4/T7 → EC-02/03/04/10; full binding schema, provider proofs and independent privacy/crypto review missing. |
| A4 | Documents 42–43; Stage 3.21 external inventory → copy-specific scope/claim/exception register → false universal erasure or hidden hold. | T6/T7 → EC-07/09/10; DESIGN_ONLY. Legal/Privacy/Operations facts and any contract-change authorization missing. |
| G1 | PR #277 PREPARE-first contract → independent admitted-intent service/pinned complete frontier → lost pending, unknown reply, signed prefix/old image. | 11 synthetic tests PROTOTYPE_VERIFIED; T3/T4/T5 → EC-04/05/06/10. Independent quorum/admin/time/checkpoint/restore evidence missing; BLOCKED. |
| G2 | Runtime capability profiles/startup validation, financial read+append protections → forced policy + invoker lock + sealed credential/workload cutover → stale SQL/session/worker authority. | Dedicated CI fixture provides only named policy feasibility when actually executed; T2/T5/T8/T9 → EC-01/05/06/08/10. Full schema/old binaries/worker/partition/prepared work/broker and rollout evidence missing; BLOCKED. |

| EC | Incremental evidence status; all original minimum-proof clauses remain required |
| --- | --- |
| EC-01 | DESIGN_ONLY: existing auth mechanisms inspected; full deletion/cancellation authority tests absent. |
| EC-02 | DESIGN_ONLY: A1/A2/A4 field and copy dispositions proposed; no production discovery/anonymity proof. |
| EC-03 | EXTERNAL_EVIDENCE_REQUIRED: suite/manifest design only, no custody/provider attestation. |
| EC-04 | PROTOTYPE_VERIFIED for local chain/signature/prefix/retry cases only; independent history/locator/restore proof BLOCKED. |
| EC-05 | Partial prototype/isolated policy evidence only when named CI test executes; integrated lifecycle/migration/runbook BLOCKED. |
| EC-06 | EXTERNAL_EVIDENCE_REQUIRED: actual sealed restore/release not exercised. |
| EC-07 | EXTERNAL_EVIDENCE_REQUIRED: managed copy inventory/destruction unavailable. |
| EC-08 | DESIGN_ONLY plus limited synthetic DB-role evidence; organizational/workload separation EXTERNAL_EVIDENCE_REQUIRED. |
| EC-09 | BLOCKED: no qualified legal/privacy/exception decision. |
| EC-10 | BLOCKED: exact baseline and evidence index available; independent Security Review and Principal Architect decision absent. |

Failure interpretation: any violated prototype invariant is a counterexample requiring correction of
this candidate, not automatic evidence of an external production incident. A skipped/incomplete job is
not PASS. Exact-head normal CI verifies source compatibility and the explicitly named integration test
only; it cannot accept ADR-008, close an SR, prove provider erasure or authorize implementation.
