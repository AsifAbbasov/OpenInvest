# Stage 3.16 — Repository Audit Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-16-REPOSITORY-AUDIT-PLAN |
| Version | 0.1.1 |
| Status | Active / planning |
| Owner | Builder Engineer |
| Supersedes | Informal next-step discussion after Stage 3.15 |
| Dependencies | `SOURCE_OF_TRUTH.md`; `REVIEW_WORKFLOW.md`; Stage 3.15 closure governance |
| Last Review Date | 2026-07-27 |
| Next Review Date | Before repository audit execution |

## Purpose

Stage 3.16 plans the mandatory full repository audit required before the next implementation stage.

The audit exists to catch architecture, scope, documentation, dependency, test, and boundary drift
after the completed Stage 3 increments and before financial algorithms such as WAC, XIRR, real
return, inflation-adjusted return, dividends, or purchasing-power work begins.

## Trigger

`REVIEW_WORKFLOW.md` requires every fifth completed stage to receive a full repository
line-by-line audit covering architecture, DDD, SOLID, API, security, privacy, performance,
dependencies, tests, documentation, cost, and ADR consistency before the next stage proceeds.

Stage 3.15 implementation and closure governance are now merged into `develop`. The next MVP gaps
include financial calculations and source-backed read models, which are higher-risk than the recent
presentation and API-boundary slices. A repository audit is therefore the next safe gate.

## Audit Scope

The future audit should inspect:

- architecture freeze and accepted ADR consistency;
- Source of Truth, roadmap, version matrix, document index, changelog, and implementation log
  consistency;
- DDD boundaries, SOLID design pressure, and layering drift in Go, TypeScript, and Python code;
- OpenAPI contract, examples, and generated/validated schema references;
- Go API, service, store, auth, import, asset, snapshot, and audit boundaries;
- Next.js presentation-only boundary under ADR-007;
- Python worker skeleton and dependency boundaries;
- PostgreSQL migrations, schema ownership, and migration validation;
- package, Go module, Python lock, CI, Docker Compose, and infrastructure configuration;
- test coverage, local verification commands, and CI evidence;
- privacy, security, retention, anonymization, token/session, CSRF, and audit-event handling;
- dependency, runtime, CI, storage, and provider cost exposure;
- financial-calculation readiness and missing canonical vectors before any production algorithms.

## Explicit Exclusions

Stage 3.16 planning does not authorize:

- code changes;
- OpenAPI changes;
- SQL migrations;
- dependency changes;
- financial algorithms or calculation formulas;
- market data, provider integrations, workers, or scheduled collectors;
- stock-card or bond-card calculations;
- dividend/coupon ingestion or calendars;
- purchasing-power equivalents;
- tax, mobile, AI, premium, public API, or email automation;
- direct pushes or merges outside the reviewed PR workflow.

## Audit Method

The future audit should run as its own reviewed stage and produce a durable report. It should:

- start only after this planning PR is reviewed, merged, and the main checkout is fast-forwarded to
  `develop`;
- record the immutable audit target SHA as the full 40-character `git rev-parse HEAD` value after
  the planning merge;
- generate a tracked-file coverage manifest from that exact SHA before review starts;
- mark every `git ls-files` path in the manifest as `audited` or `excluded`;
- justify every exclusion by file or narrow class; acceptable exclusions are limited to generated,
  vendored, binary, or archival materials whose contents are not active product/runtime authority;
- treat active source, tests, migrations, OpenAPI, configuration, lockfiles, scripts, and governance
  documents as in scope unless the audit report gives a specific, reviewed reason otherwise;
- provide line-by-line evidence for every active audited file, or a structured-parser equivalent for
  machine-generated/lockfile content where literal line review would be misleading;
- separate blocking findings from non-blocking observations;
- identify each finding with file/line evidence where possible;
- map findings to architecture, DDD, SOLID, API, privacy/security, test, performance, dependency,
  documentation, cost, and ADR-consistency risk;
- require fixes or explicit human risk acceptance before the next implementation stage.

## Acceptance Criteria

- The audit plan is reviewed and merged before audit execution starts.
- The future audit report names one immutable post-planning audit target SHA.
- The future audit report includes a file-level coverage manifest for every tracked path at that
  SHA.
- The future audit report covers every required category in `REVIEW_WORKFLOW.md`: architecture,
  DDD, SOLID, API, security, privacy, performance, dependencies, tests, documentation, cost, and ADR
  consistency.
- Every manifest entry is either audited with evidence or excluded with a narrow reviewed reason.
- The future audit report does not silently authorize implementation work.
- Any financial-algorithm stage remains blocked until audit findings are resolved or explicitly
  accepted.
- No market-data, provider, tax, mobile, AI, or worker scope enters through the audit gate.

## Recommended Next Step

After this planning document is reviewed and merged, execute the Stage 3.16 repository audit as a
separate read-only audit stage. Only after that audit is approved, and any blocking findings are
resolved or explicitly accepted, select the next implementation planning gate.
