# Stage 4 - Commercial Readiness Assurance Plan

| Field | Value |
| --- | --- |
| Status | Planning-only candidate; no implementation authorization |
| Date | 2026-09-04 |
| Planning base | `develop@983104267221706c3c2ebd8d9be358e3921334b5` |
| Preconditions | Original repository audit closed by Stage 3.56: 32/32, with P0=0 / P1=0 / P2=0 / P3=0 |
| Purpose | Define an evidence-honest route from locally testable assurance work to later commercial-operation gates |
| Runtime, API, schema, migration, dependency, CI, infrastructure change | None |
| Production data, credentials, providers, external services, or paid engagement | None |
| Relation to Stage 3.25 | Separate; this plan neither replaces nor authorizes privacy Security Review evidence collection |

## 1. Objective

The original repository audit is closed. This plan does not reopen it or certify the product for
commercial production. It identifies the next assurance work that is valuable for a financial
application and separates what can be implemented and verified locally from work that requires an
authorized production environment, independent specialists, or budget.

Each item below is a future, separately reviewed scope. A later implementation must state its
invariants, input domain, resource limits, evidence, rollback behavior, and acceptance criteria
before changing code or configuration.

## 2. Current evidence and boundaries

The current repository has deterministic tests around financial transactions, imports, exact
idempotency replay, decimal parsing, and snapshot rebuilds. It does not currently contain a
property-testing framework or native fuzz targets, a declared load-test workload and threshold,
production telemetry, a production backup/restore rehearsal, or an independent penetration-test
engagement. Absence of those artifacts is not evidence that the corresponding capability has been
performed.

No work in this plan may:

- use real user, broker, restricted, or production data;
- claim availability, latency, security, backup, restore, or incident results that were not measured;
- add a hosted service, cloud provider, credential, paid testing engagement, or production deployment;
- turn a test harness into a public API, financial calculation policy, or ledger migration without a
  separately approved scope; or
- treat a local test pass as a substitute for independent security or operational evidence.

## 3. Assurance sequence

| Order | Future scope | Primary value | Required boundary before approval |
| --- | --- | --- | --- |
| 4.1 | Property-based testing pilot | Generate broad valid and invalid combinations around high-risk invariants | Select a small invariant set and deterministic shrinking/reproduction format; do not introduce float arithmetic or change financial semantics |
| 4.2 | Native fuzzing pilot | Exercise parser and boundary handling against malformed or unexpected inputs | Define corpus, time/resource budget, crash/reproduction handling, and inputs that are safe to retain |
| 4.3 | Load and benchmark baseline | Establish reproducible, declared local/staging workloads before performance claims | Define topology, data shape, concurrency, warmup, duration, metrics, pass/fail thresholds, and non-production data policy |
| 4.4 | Operational readiness design | Specify telemetry, alert, backup/restore, release, and incident evidence that can later be collected honestly | Select an environment and owners; Stage 3.25 privacy evidence remains independently governed |
| 4.5 | Independent security assessment | Obtain an authorized outside attack-oriented and architectural assessment | Approve budget, scope, rules of engagement, environment, data handling, remediation ownership, and disclosure path |
| 4.6 | Production resilience and operational evidence | Validate behavior during real operations and later controlled fault scenarios | Require an established production topology, observability, recovery procedures, change control, and explicit risk approval |

Formal verification or model checking is not a prerequisite for every scope. It may be proposed
later for a bounded state machine such as command idempotency, refresh/session rotation, or a
concurrency-sensitive ledger/import transition, with a model and properties small enough to review.

## 4. Candidate invariant inventory

The first two scopes should use existing behavior rather than inventing new financial rules. Candidate
properties, each subject to a separate implementation decision, are:

| Area | Candidate property | Existing boundary |
| --- | --- | --- |
| Decimal handling | Accepted decimal strings preserve the documented scale and rejected grammar never reaches financial persistence as a float | `backend-go/internal/decimal` and import review parsing |
| Command idempotency | A completed request retried with the same principal, key, path, and serialized intent replays one result; a conflicting intent does not create another financial command | `backend-go/internal/verticalslice`, HTTP replay, and PostgreSQL deduplication |
| Import transition | The parse/review/approve/append path binds the approved rows and does not append a different intent after retry or parser-version recovery | `backend-go/internal/importer` and `backend-go/internal/httpapi` |
| Snapshot rebuild | The same append-only transaction history and as-of dates produce the same resulting snapshot projection | vertical-slice and PostgreSQL snapshot behavior |

These are candidate testing properties, not new product guarantees. A future test must be checked
against the current API, persistence, and historical replay contracts at the implementation base.

## 5. Fuzzing and load-test guardrails

Native Go fuzzing is preferred for a first parser-boundary pilot because it keeps the initial tool
surface small. A fuzz scope must avoid network calls, external credentials, non-deterministic shared
databases, unbounded corpus growth, and retention of sensitive input. A discovered failure must yield
a minimized reproducible test before it is called remediated.

No load number is meaningful until its environment and workload are declared. A later load or
benchmark scope must identify the API path or worker boundary, request mix, fixture source, database
state, concurrency, duration, host limits, collection method, and pass/fail interpretation. It must
report measured conditions and variance, not generalize a local result into a production capacity
claim.

## 6. Operational and external gates

Operational readiness cannot be proved by this repository alone. Before a commercial launch decision,
the responsible owners need independently reviewable evidence for the selected deployment topology,
telemetry and alert handling, least-privilege access, backup and restore rehearsal, vulnerability and
patch process, incident response, capacity planning, and privacy controls.

An independent penetration test and security audit require an authorized target environment, written
rules of engagement, budget, safe test data, escalation contacts, and a remediation/retest process.
They are not replaced by model review, static checks, or local fuzzing.

Chaos testing is deferred until the product has a controlled production-like topology, backup and
restore practice, telemetry, and an approved blast-radius policy. It must not be performed merely to
claim resilience.

## 7. Acceptance conditions for this planning stage

This stage is complete as a planning candidate only when review confirms that it:

1. records the closed original audit state without reopening a historical finding;
2. separates local engineering checks from production, independent-security, and paid activities;
3. names no fabricated measurements, vendor selection, compliance certification, or operational proof;
4. keeps Stage 3.25 privacy evidence collection separate; and
5. authorizes no runtime, schema, API, dependency, infrastructure, or external-service change.

Approval of this document does not approve any item in the assurance sequence. The next implementation
scope must be selected explicitly and reviewed under the effective delivery workflow.
