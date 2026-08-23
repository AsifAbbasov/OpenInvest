# Audit Remediation Documentation Standard

| Field | Value |
| --- | --- |
| Document ID | GOV-AUDIT-REM-001 |
| Status | Mandatory governance standard for repository-audit remediation |
| Applies to | Every remaining audit finding and every future remediation finding accepted into the canonical backlog |
| Owner | Principal Architect |
| Effective baseline | Stage 3.33 canonical closure baseline; applies prospectively to Stage 3.34 and all P3 remediation |

## Purpose

OpenInvest remediation work must preserve the engineering reasoning behind a fix, not only the final code change. A future engineer must be able to determine what failed, why it mattered, how the failure could manifest in the product, which alternatives were considered, why the selected remediation was preferred, how regression is prevented, and which residual risks remain.

README is an entry point and summary only. The detailed evidence belongs in the relevant stage/remediation dossier under `docs/stages/`, with the global index maintained in `docs/audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md`.

A finding is not documentation-complete merely because code and tests exist.

## Mandatory 18-part finding record

Every remediation stage that closes, partially closes, or materially changes the disposition of an audit finding MUST contain the following information for each finding in scope.

### 1. Finding / symptom

State exactly what the review discovered. Identify the affected contract, component, boundary, workflow, data path, or operational control. Avoid vague statements such as “security issue fixed.”

### 2. Root cause

Explain why the defect existed. Distinguish the immediate bug from the design or control failure that allowed it to exist.

Examples of root-cause classes include:

- ownership of a result assigned to the wrong layer;
- incomplete identity model;
- missing persistence constraint;
- boundedness enforced only at an outer adapter;
- mutable state consulted during idempotent replay;
- privilege validation proving only the current state rather than the reachable credential graph;
- contract semantics accepted by one boundary but ignored by another.

### 3. Failure scenario

Give at least one concrete scenario that demonstrates the defect. For concurrency, security, replay, financial, persistence, or lifecycle defects, the scenario must include the state transition or sequence needed to trigger the failure.

### 4. Project/product impact

Explain how the defect would affect OpenInvest if triggered. Relevant impact classes include:

- incorrect financial result;
- duplicate or missing ledger effect;
- inconsistent snapshot/analytics output;
- security boundary escape;
- privacy exposure;
- availability/resource exhaustion;
- misleading API/UI result;
- unrecoverable operational ambiguity;
- governance bypass;
- maintenance or upgrade risk.

### 5. Severity rationale

Explain why the finding is P0/P1/P2/P3. The explanation must connect likelihood/reachability and impact to the assigned severity; it must not rely only on the original label.

### 6. Existing guarantee or invariant violated

Identify the architecture, API, persistence, security, privacy, operational, or governance guarantee that the defect violated. Cite the relevant Source of Truth, ADR, architecture document, contract, or established invariant when applicable.

### 7. Candidate solutions considered

Record the materially different approaches considered before implementation. Do not invent alternatives after the fact; record only approaches actually evaluated or rejected during the remediation/review cycle.

### 8. Chosen remediation

Describe the implemented or approved fix precisely enough that a reviewer can understand which layer owns the new guarantee and where enforcement occurs.

### 9. Why this remediation was chosen

Explain why the selected fix is preferable for correctness, security, privacy, maintainability, operational safety, scope discipline, or compatibility with the frozen architecture.

### 10. Rejected alternatives

Record rejected material alternatives and why they were rejected. Typical reasons include weaker correctness, hidden race windows, contract drift, architecture expansion, unsupported infrastructure, excessive state, inability to prove atomicity, or inability to fail closed.

### 11. Trade-offs and cost

Record the cost introduced by the solution: extra query, lock, memory, latency, CI time, operational step, stricter validation, compatibility limitation, or maintenance burden. A remediation must not be described as cost-free when it is not.

### 12. Regression tests / proof

Identify the automated test, integration test, migration proof, adversarial fixture, CI job, or API contract verification that would fail if the defect returned. For security or concurrency findings, prefer adversarial state construction over happy-path-only assertions.

### 13. Adversarial / independent review findings

Preserve every material `REQUEST CHANGES` finding produced during independent review. Do not overwrite or erase earlier rejected designs after a later solution passes review.

### 14. Remediation iterations

When a first fix is insufficient, describe each iteration in order:

1. previous candidate;
2. reviewer-discovered gap;
3. new failure path;
4. corrective design;
5. new proof.

The historical path is part of the engineering record.

### 15. Residual risks and limitations

State what remains unaddressed, deferred, operationally dependent, plan-dependent, provider-dependent, compatibility-only, or outside the finding scope. If a reviewer explicitly judged a residual issue non-blocking, record that distinction.

### 16. Operational / deployment consequences

When applicable, document migrations, role provisioning, environment variables, provider controls, branch settings, rollout order, backfill, cleanup, recovery, monitoring, or manual governance actions required to make the fix effective outside source code.

### 17. Exact evidence

Record exact evidence where available:

- implementation PR;
- final reviewed head SHA;
- exact-head CI run and conclusion;
- review verdict;
- explicit human merge authorization;
- squash-merge commit;
- closure PR and closure commit.

Do not cite a CI run from an older head as proof for a later head.

### 18. Final canonical status

State `OPEN`, `PARTIALLY REMEDIATED`, `IMPLEMENTED / NOT YET CANONICAL`, or `CLOSED` and identify the canonical commit at which that status became true. Reviewer approval alone does not make a finding canonically closed.

## Minimum stage structure

A remediation stage SHOULD organize its dossier in this order:

1. scope and source findings;
2. per-finding 18-part records or a clearly equivalent structure;
3. cross-finding atomicity/concurrency/security interactions;
4. regression and CI evidence;
5. independent review history;
6. residual risk and out-of-scope items;
7. deployment/operational requirements;
8. canonical closure evidence.

A stage may combine repeated sections when several findings share one root cause or one implementation boundary, but every mandatory information class must remain discoverable for each finding.

## README rule

README MUST remain a concise navigation layer. For each completed audit-remediation stage it should contain only:

- findings closed;
- one-sentence problem summary;
- one-sentence remediation summary;
- implementation/closure status;
- link to the detailed stage dossier and/or audit register.

README must not become the primary forensic engineering record.

## Global audit register rule

`REPOSITORY_AUDIT_REMEDIATION_REGISTER.md` is the single cross-stage index for the original 32 findings. It must record at minimum:

- finding ID and severity;
- problem/root-cause summary;
- impact summary;
- remediation strategy or planned remediation;
- rationale summary;
- stage/PR/canonical evidence;
- current status;
- link to the detailed dossier.

The register is an index and executive engineering summary. The stage dossier remains the detailed evidence source.

## Review and closure gate

Beginning with Stage 3.34, independent review MUST check documentation completeness against this standard before a finding is declared canonically closed.

If implementation is correct but the required reasoning/evidence record is materially incomplete, closure governance must remain incomplete until the record is repaired.

## Historical findings

Stages 3.27–3.33 predate this formal standard but already preserve most of these information classes. The audit register may summarize and link their existing evidence without rewriting history. Where a historical dossier lacks a final status line because closure was recorded separately, governance synchronization may add a non-destructive final canonical-status note and link to the closure record.
