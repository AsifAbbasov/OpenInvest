# Documentation Changelog

| Field | Value |
| --- | --- |
| Document ID | REG-CHG-001 |
| Version | 1.2.4 |
| Status | Active |
| Supersedes | None |
| Dependencies | `SOURCE_OF_TRUTH.md` |

> **Document role — dated change history**
>
> This file records what changed over time; historical entries remain evidence of what was true when recorded.
> It does not define current runtime truth, future sequencing, executable HTTP authority, provider/source-use rights, or architectural decision authority.
> Use [`SOURCE_OF_TRUTH.md`](SOURCE_OF_TRUTH.md), [`ROADMAP.md`](ROADMAP.md), [`../openapi/openapi.yaml`](../openapi/openapi.yaml), [`registries/DATA_SOURCE_REGISTRY.md`](registries/DATA_SOURCE_REGISTRY.md), and accepted [`ADR/`](ADR/) for those respective roles.

## 2026-09-10 — Historical forensic reconciliation canonicalized after protected merge

Canonical record: PR #167; commit(s) `5da679f0d15f8742239659ff9cebc451a213894d`, `e742d1084fdcc5751acf896738ba9065a4b0f741`.

## 2026-09-10 — Historical feature forensic documentation reconciliation

- The candidate records 47 material forensic records: 33 with full forensic coverage and 14 partial records, with 60 unsupported historical forensic fields explicitly marked `NOT RECORDED IN CONTEMPORANEOUS EVIDENCE` rather than reconstructed.
- The reconciliation changes no Go runtime, SQL/migrations, OpenAPI behavior, frontend runtime, dependency, provider/source-use decision, provider activation state or Stage 3.78 authorization.

## 2026-09-09 — Feature 3D constrained T-Invest Corporate Actions implementation merged; documentation closure synchronized

- PR #164 was squash-merged into protected `develop` at `247081a95a7daf33c0077c88c5f41cb2e8161865` from final evidence head `ddc5b5be36f0b5ae127c6ee430918a9dba1b453e`; merged tree `ec7bc9152210913b1a6ef742bddd599abee50bc7` matches the final evidence tree.
- Exact implementation-head CI #499 / run `34353668492` and final evidence-head CI #501 / run `34356672643` both completed all 10 protected jobs successfully.
- Feature 3D implements only registry mode `TINVEST_CORPORATE_ACTIONS_CONSTRAINED`: `GetDividends` -> `DIVIDEND` and `GetBondCoupons` -> `COUPON`, with static approved instruments, exact Decimal normalization, bounded REST transport, fail-fast concurrency, no retry/polling/persistence, and provider-neutral public contracts.
- GitHub event history records `AsifAbbasov` as the actor for the Ready transition and the squash merge; the assistant performed no merge mutation. A later chat message saying `разрешаю` occurred after the merge and is not treated as retroactive pre-merge evidence.
- This documentation closure starts no Stage 3.78 work and grants no broader T-Invest rights or operational activation.

## 2026-09-07 — Stage 3.74 transaction correction/reversal canonical and documentation synchronized

- PR #150 was squash-merged into protected `develop` at `0580bf7e98c532202f84bbf9ceacd97aedbe4140` from exact final head `ff6d3efb8bb0d63f8cab55ac82fdf1052946aa02`.
- CI #435 / run `34117662576` completed all 10 required jobs successfully before protected merge.
- Stage 3.74 makes correction/reversal append-only and auditable, projects deterministic effective-ledger truth into positions/Time Machine, rejects historical oversell atomically, and exposes retry-safe Edit/Reverse UX.
- No database migration, market provider activation, paid infrastructure or production rollout was introduced.
- Post-merge documentation synchronization advances the next runtime gate to Stage 3.75+.

## 2026-09-07 — Stage 3.73 Portfolio Time Machine lifecycle closure

Canonical record: PR #148; commit(s) `683f9c4647f888bb3dbdfb9dd365b84b95137b46`, `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7`.

## 2026-09-07 — Repository documentation reconciliation through Stage 3.72

- Synchronized README, Implementation Log, Version Matrix, Open Questions and Document Index with the canonical Stage 3.72 runtime/documentation baseline.
- Recorded the original Stage 3.16 repository audit as fully remediated: 32/32 findings CLOSED, P0/P1/P2/P3 = 0/0/0/0, through Stage 3.56.
- Registered the Stage 3.57–3.60 market-data lifecycle: provider-neutral boundary and delayed MOEX ISS adapter implemented, production/public activation NO-GO, adapter dormant.
- Registered Corporate Actions Stages 3.61–3.67, including Calendar/Heatmap API/UI and request cancellation; real-source Feature 3D remains separately gated.
- Registered the Stage 3.68/3.69 Dividend Calculator lifecycle.
- Registered ADR-009 and Stages 3.70–3.72: deterministic manual SELL/WAC/acquisition-basis accounting and the dedicated portfolio positions API/UI with explicit market valuation unavailable semantics.
- Added a canonical cross-finding audit register and a repository-documentation reconciliation record.
- No Go, SQL/migrations, OpenAPI runtime contract, frontend runtime behavior, dependencies, provider configuration or production deployment changed.

## 2026-06-19 — Architecture Freeze v1.2

- Approved Documents 42 and 43 as the two highest-priority architecture sources.
- Resolved business-date versus UTC timestamp semantics.
- Replaced exactly-once transport language with at-least-once delivery and idempotent business processing.
- Corrected privacy terminology from pseudonymization to anonymization when re-identification is impossible.
- Froze MVP scope, asset scope, financial precision, retention, SLO boundaries, data schemas, and document precedence.
- Consolidated source documents into the repository and activated Documentation Freeze.

## 2026-06-21 — Stage 2 governance hardening

- Registered proposed ADR-006 and all Stage 2 contract artifacts without approving the ADR.
- Added the repository-owned OpenAPI validator to the pull-request CI gate.
- Reserved explicit `EXAMPLE_*` source identifiers so contract examples cannot be mistaken for
  approved MOEX, Rosstat, CBR, or other production sources.
- Synchronized the Stage 2 status across governance registries and the implementation log.

## 2026-06-25 — Stage 2 final review blockers

- Required explicit reversal `effectiveDate` BusinessDate so immutable-ledger reversals and
  snapshot rebuilds do not depend on system timestamps.
- Changed economically non-negative aggregate values from signed `Money` to `NonNegativeMoney`.
- Tightened `traceparent` validation to reject W3C-invalid version `ff`, all-zero trace IDs, and
  all-zero parent IDs.
- Documented repository OpenAPI validator limitations and added focused mutation guards instead of
  claiming complete JSON Schema 2020-12 compliance.
- Documented that the initial repository validator was a temporary tooling risk to be removed from
  the approved stack.

## 2026-06-25 — Stage 2 closure and ADR-006 acceptance

- Squash-merged PR #2 into `develop` at `bfde623552ebea6eac7bdaabf0d1a2263883de12`.
- Marked Stage 2 Contract and Canonical Model Freeze as closed.
- Declared `develop` at the Stage 2 merge commit as the canonical Stage 2 baseline.

## 2026-06-25 — ADR-007 Web frontend amendment

- Replaced the current Web MVP React + Vite SPA target with Next.js App Router + TypeScript + pnpm.
- Restricted Next.js to presentation, routing, rendering, metadata, and Go API orchestration.
- Prohibited business APIs, database access, financial calculations, and external-source integration
  in Next.js.
- Kept Go as the canonical business API and Python as the future analytics/collector worker layer.
- Confirmed SwiftUI and Jetpack Compose as future-only clients with no current mobile scope.

## 2026-06-26 — Next.js Web presentation amendment closure

- Squash-merged PR #4 into `develop` at `6a7748cc24fc852d42b90b0e0cb843b6020f3973`.
- Declared Next.js App Router + TypeScript + pnpm as the current Web presentation baseline.
- Confirmed Stage 3 remains not started; the next approved work item is Stage 3 planning.

## 2026-06-26 — Stage 3 first vertical slice planning started

- Added the Stage 3 planning document for the first portfolio/transaction/snapshot/API/Web slice.
- Defined small implementation PR boundaries to avoid scope creep.
- Reconfirmed that, tax export, mobile, external providers, broker import, and Stage 3 business
  expansion beyond the first slice remain out of scope.

## 2026-06-27 — Stage 3.1 database foundation started

- Squash-merged PR #6 into `develop` at `03908905b74da5c35d2fee71c2ed4956e4c06464`.
- Started the local PostgreSQL foundation for the first vertical slice.
- Added plain SQL migration pairs instead of selecting a migration library.
- Added migration validation to CI.
- Updated the local PostgreSQL 18 Docker volume mount to support live migration verification.
- Kept Go API, Next.js presentation, Python workers, and external provider integrations out of
  Stage 3.1 scope.

## 2026-06-27 — Stage 3.1 closed and Stage 3.2 started

- Squash-merged PR #7 into `develop` at `b1a3f23`.
- Started Stage 3.2 Go API Vertical-Slice Backend.
- Added the first Go API path for portfolio creation, transaction append, local snapshot rebuild,
  and summary read without changing the frozen OpenAPI contract.
- Kept frontend screens, mobile, tax, dividends, external providers, workers, and Redis out of
  Stage 3.2 scope.

## 2026-06-27 — Stage 3.2 closed and MVP product-risk refinement added

- Squash-merged PR #8 into `develop` at `8971918c8046fb9a2d6bf9f97897432cf08fbde1`.
- Added `product/MVP_PRODUCT_RISK_REFINEMENT.md` to convert PRD criticism into controlled MVP
  risk governance.
- Sharpened the initial ICP toward long-term, dividend, FIRE, and multi-account investors with
  real portfolio-accounting pain.
- Moved broker file import and reconciliation into near-term public-MVP readiness consideration
  while keeping direct broker API synchronization and credential scraping out of current scope.
- Clarified that Tax  cannot be a calculation source; any future tax core must be deterministic
  and test-vector driven.
- Repositioned Purchasing Power as a secondary explanatory insight below capital, real return,
  dividends/coupons, and inflation-adjusted performance.

## 2026-06-27 — Product-risk refinement closed and Stage 3.3 started

- Squash-merged PR #9 into `develop` at `65bdf6537b44ed57e1c00bf68d2dacd70aa09702`.
- Started Stage 3.3 Next.js Presentation Slice.
- Added the first Web presentation path for portfolio list/detail, add-transaction form, summary,
  and transaction history using only the Go API.
- Kept Next.js out of business calculations, database access, provider integration, Route Handlers,
  Server Actions with business behavior, authentication implementation, and mobile scope.

## 2026-07-01 — Ruby tooling removed from active project stack

- Removed the temporary Ruby OpenAPI and migration validators from `scripts/`.
- Replaced them with Go CLI validators under `backend-go/cmd` so validation tooling stays inside
  the approved Go backend stack.
- Updated CI and local-check documentation to run the Go validators instead of Ruby scripts.

## 2026-07-01 — Stage 3.3 closed and Stage 3.4 selected as next work

- Squash-merged PR #10 into `develop` at `11805cc298bba13f09f7f7af8b1e1178dc351209`.
  merge approval.
- Added the first Web path for portfolio list/detail, create portfolio, add transaction, summary,
  and transaction history through the Go API boundary.
- Added local Go API CORS/OPTIONS support for explicit local Web origins and limited the transaction
  form to the transaction types currently accepted by the Go vertical slice.
- Selected Stage 3.4 end-to-end verification as the next implementation focus.

## 2026-07-01 — Stage 3.4 end-to-end verification started

- Added root pnpm commands for infrastructure, local API/Web startup, checks, and Stage 3.4 smoke
  verification.
- Added `scripts/stage-03-04-smoke.sh` to prove the local PostgreSQL → Go API → immutable
  transaction append → snapshot rebuild → summary response path.
- Added the Stage 3.4 report and synchronized governance registries.
- Kept Stage 3.4 limited to verification and onboarding; no new business logic, SQL migrations,
  provider integrations, workers, frontend feature expansion, or mobile code were added.

## 2026-07-01 — Stage 3.4 closed

- Squash-merged PR #13 into `develop` at `86582efaa420b2c38465a5d0da041814149392c7`.
- Added root verification commands and a controlled local smoke path for PostgreSQL, Go API,
  immutable transaction append, snapshot rebuild, and summary response.

## 2026-07-02 — Stage 3.5 broker-file import design started


## 2026-07-02 — Stage 3.5 closed and Stage 3.6 started

Canonical record: PR #14; commit(s) `072d38d94b529221d6467502f82f03a674a7d805`.

## 2026-07-02 — Stage 3.6 broker-file import reconciliation slice closed

Canonical record: PR #15; commit(s) `e2b05650a4422b97d4bd924254367106b6a4686b`.

## 2026-07-02 — Stage 3.7 import append planning started

- Added a documentation-only Stage 3.7 planning document for the future atomic import append scope.
- Defined the proposed append path as explicit approved import decisions → atomic database append →
  immutable ledger entries → snapshot rebuild → audit evidence.
- Kept implementation, public import endpoints, frontend upload UI, broker/provider integrations,

## 2026-07-02 — Stage 3.7 import append slice started

- Added the active implementation report for the internal atomic import append slice.
- Scoped the implementation to Go service/store internals and PostgreSQL transaction behavior only.
- Kept public import API, upload UI, SQL import-session persistence, workers, broker/provider
  integrations, tax logic, mobile, and  out of scope.

## 2026-07-02 — Stage 3.7 import append slice closed

- Squash-merged PR #18 into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`.
- Added internal atomic append of user-approved import rows with idempotency protection, duplicate
  revalidation, minimal audit evidence, and deterministic snapshot rebuilds.
- Added live PostgreSQL coverage for concurrent duplicate-batch serialization.
- Kept public import API, upload UI, SQL import-session persistence, workers, provider
  integrations, tax, mobile, and Stage 3.8 implementation out of scope.

## 2026-07-03 — Stage 3.8 import review append flow planning started

- Added a documentation-only planning scope for the future internal import review → append flow.
- Defined the proposed orchestration as broker CSV bytes → parse/normalize → review candidates →
  explicit accepted decisions → atomic append → snapshot rebuild → deterministic result.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, broker/provider integrations, tax, mobile, and implementation out of
  scope.

## 2026-07-03 — Stage 3.8 import review append flow slice started

- Added the active implementation report for the internal import review → append flow slice.
- Scoped implementation to an internal Go orchestration package and tests only.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, and Stage 3.9 out of scope.

## 2026-07-03 — Stage 3.8 import review append flow slice closed

- Squash-merged PR #21 into `develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`.
- Added internal import review → append orchestration with explicit approved decisions and
  non-sensitive result metadata.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, and Stage 3.9 out of scope.

## 2026-07-08 — Stage 3.9 import API boundary planning started


## 2026-07-08 — Stage 3.9 import API boundary planning closed and implementation started

- Squash-merged PR #23 into `develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`.
- Started the Stage 3.9 implementation slice for public Go API import review/append endpoints.
- Added a stateless API-boundary decision: review results are transient, append receives the same
  CSV payload plus explicit row decisions, and append reruns review before atomic store mutation.
- Kept raw CSV persistence, import-session tables, frontend upload UI, direct broker APIs, workers,
  tax, mobile, and Stage 3.10 outside the implementation scope.

## 2026-07-08 — Stage 3.9 import API boundary slice closed

- Squash-merged PR #24 into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`.
- Added public Go API endpoints for transient user-supplied CSV import review and explicit append.
- Preserved the stateless boundary: no review IDs, no import-session table, no raw CSV persistence,
  and append reruns review before atomic store mutation.
  replay, full import append request hashing, and append-payload example validation.
- Kept frontend upload UI, SQL import-session persistence, workers, broker/provider integrations,
  tax, mobile, and Stage 3.10 implementation out of scope.

## 2026-07-08 — Stage 3.10 import upload/review UI planning started

Canonical record: commit(s) `682ffd856395a6e3e988817551a512898fda2d38`.

## 2026-07-09 — Stage 3.10 import upload/review UI slice started

- Squash-merged Stage 3.10 planning into `develop` at
  `27480d6ff22e2929e33aeac352aef8a1b01bb448`.
- Preserved the Go API as the only business authority; no backend contract, SQL, provider, worker,
  tax, mobile, or  scope entered the slice.

## 2026-07-09 — Stage 3.10 closed and Stage 3.11 planning started

- Squash-merged PR #27 into `develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71`.
  approval.
- Started Stage 3.11 as documentation-only planning for the future authentication, session, CSRF,
  and privacy-default boundary.
- Kept auth implementation, schema migrations, password hashing, token issuance, frontend session
  code, business logic, workers, tax, mobile, and provider integrations out of scope.

## 2026-07-09 — Stage 3.11 authentication and privacy-boundary slice started

- Squash-merged PR #28 into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125`.
- Closed the Stage 3.11 planning scope and started the implementation slice on
  `feature/stage-03-11-auth-privacy-slice`.
- Scoped implementation to Go API auth handlers, Argon2id password hashing, rotating refresh
  sessions, CSRF enforcement, privacy-default persistence, additive PostgreSQL migration, and tests.
- Kept frontend auth UI, business logic in Next.js, email verification, OAuth/passkeys/2FA, workers,
  provider integrations, tax, mobile, and Stage 3.12 out of scope.

## 2026-07-09 — Stage 3.11 authentication and privacy-boundary slice closed

- Squash-merged PR #29 into `develop` at `5c49173ac858995929f266c2de991282dd194dec`.
- Marked Stage 3.11 implementation as complete in the roadmap, Source of Truth, document index,
  version matrix, implementation log, and stage report.
- Confirmed the closed slice remains Go API/auth persistence only and does not authorize frontend
  auth UI, provider integrations, workers, tax, mobile, or Stage 3.12 implementation.

## 2026-07-09 — Stage 3.12 Web authentication UI planning started

- Added the planning-only Stage 3.12 document for a future Next.js presentation auth/session UI.
- Registered Stage 3.12 planning in the roadmap, Source of Truth, document index, version matrix,
  implementation log, and Stage 3 plan.
- Kept implementation, backend changes, OpenAPI changes, SQL migrations, token-storage changes,
  provider integrations, workers, tax, mobile, and Stage 3.13 out of scope.

## 2026-07-11 — Stage 3.12 Web authentication UI slice started

- Squash-merged PR #31 into `develop` at `25be13ce84844562e0381b79f4b81cbfed7eb44d`.
- Started the implementation slice for a Next.js presentation-only registration, login,
  authenticated shell, refresh, and logout UI over the existing Stage 3.11 Go API auth boundary.
- Scoped implementation to typed frontend auth API calls, in-memory access-token handling, CSRF
  wiring for refresh/logout, route gating, presentation states, tests, and local CORS credentials
  support for the HttpOnly refresh cookie.
- Kept business logic in Next.js, Route Handlers, Server Actions, OpenAPI contract changes, SQL
  migrations, refresh-token JavaScript storage, email verification, OAuth/passkeys/2FA, provider
  integrations, workers, tax, mobile, and Stage 3.13 out of scope.

## 2026-07-11 — Stage 3.12 Web authentication UI slice closed

- Squash-merged PR #32 into `develop` at `b4840b60346109e3cd54a07d9e1e131fc0cfad23`.
- Closed the Stage 3.12 presentation-only Web authentication UI slice after green CI and strict
- Updated governance registries and the Source of Truth to mark Stage 3.12 complete and advance the
- Kept Route Handlers, Server Actions, OpenAPI changes, SQL migrations, provider integrations,
  workers, tax, mobile, and Stage 3.13 implementation out of scope.

## 2026-07-11 — Stage 3.13 instrument catalog planning started

- Squash-merged Stage 3.12 closure governance into `develop` at
  `321eaf4f75df83d85fd356a8d6a454e49bbc4db4`.
- Added the planning-only Stage 3.13 document for a future backend-owned MVP instrument catalog
  boundary for MOEX shares and bonds.
- Registered Stage 3.13 planning in the roadmap, Source of Truth, document index, version matrix,
  implementation log, and Stage 3 plan.
- Kept implementation, provider integrations, workers, market-data ingestion, financial
  calculations, tax, mobile, and frontend business authority out of scope.

## 2026-07-12 — Stage 3.13 instrument catalog slice started

Canonical record: PR #34.
  `ca16af9adba249fc8c32c9b246b5f92f7e290b92`.
- Started the backend-only instrument catalog implementation on
  `feature/stage-03-13-instrument-catalog`.
- Added an implementation report for approved local asset fixture resolution, unsupported ticker
  rejection, existing `investment.assets` usage, and stock/bond bucket preservation.
- Kept OpenAPI changes, SQL migrations, Go handler changes, frontend work, provider integrations,
  workers, market-data ingestion, stock/bond cards, dividend/coupon scope, tax, mobile, and  out
  of scope.

## 2026-07-12 — Stage 3.13 implementation hardening updated


## 2026-07-13 — Stage 3.13 instrument catalog slice closed

Canonical record: PR #35; commit(s) `b9c05fb14d0ee03e6de4dfc04ff67c16da33040b`.

## 2026-07-13 — Stage 3.14 asset search/card API boundary planning started

- Squash-merged Stage 3.13 closure governance into `develop` at
  `45a298e3ba36dbe711fa27b8d044d80a77cfd74a`.
- Started documentation-only planning for a future Go API asset search/detail boundary over the
  Stage 3.13 backend-owned catalog.
- Kept implementation, frontend stock/bond cards, OpenAPI changes, SQL migrations, external
  provider integrations, workers, market-data ingestion, financial calculations, tax, mobile, and
  out of scope.

## 2026-07-13 — Stage 3.14 technical reassessment fixes


## 2026-07-13 — Stage 3.14 asset API boundary slice started

- Squash-merged Stage 3.14 planning PR #37 into `develop` at
  `2c4f7853599a455bb0cc04114b338a1145baf39c`.
- Started the backend-only implementation slice on `feature/stage-03-14-asset-api-boundary`.
- Added the implementation report for the public Go API asset search/detail boundary over the
  approved Stage 3.13 local catalog.
- Scoped implementation to backend asset search summaries with `lastPrice: null` and a wired but
  deferred asset-detail boundary until registered runtime provenance and mandatory detail fields are
  available.
- Kept OpenAPI changes, SQL migrations, external provider integrations, market-data ingestion,
  frontend stock/bond cards, workers, financial calculations, tax, mobile, and Stage 3.15 out
  of scope.

## 2026-07-14 — Stage 3.14 asset API boundary slice closed

- Squash-merged PR #38 into `develop` at `57a9404952cb65693614109dd4a14d41fa5c4295`.
- Closed the public Go API asset search boundary over active canonical approved catalog rows.
- Preserved `lastPrice: null` and deferred asset-card detail to avoid fabricated market data,
  source provenance, sector, face value, maturity date, or coupon-type facts.
- Advanced the governance baseline to Stage 3.14 implementation completion; Stage 3.15 remains not
  started.

## 2026-07-26 — Stage 3.15 Web asset discovery UI planning started

- Squash-merged Stage 3.14 closure governance PR #39 into `develop` at
  `f5289eb604b8ba31aa422d0d09950da02e0f48b3`.
- Started documentation-only planning for a future Next.js presentation-only asset search entry and
  deferred asset-card state over the existing Go asset API.
- Kept implementation, OpenAPI changes, SQL migrations, Route Handlers, Server Actions, direct
  database access, market data, provider integrations, workers, financial calculations, tax, mobile,
  and  out of scope.

## 2026-07-26 — Stage 3.15 technical reassessment fixes

- Tightened the future Web asset discovery UI state contract for query/type changes, cursor reset,
  accepted cursor chains, and stale-response protection.
- Required public asset API calls to use `credentials: "omit"` and omit bearer tokens, cookies, CSRF
  headers, and browser storage usage.
- Expanded accessibility criteria beyond labels to define testable keyboard navigation, focus
  destination/restoration, and live-region/error announcements for asynchronous search states.

## 2026-07-26 — Stage 3.15 Web asset discovery UI slice started

- Squash-merged Stage 3.15 planning PR #40 into `develop` at
  `dfeab109b2825fe0e0317e87a7abf2e706a29ea6`.
- Kept backend API, OpenAPI, SQL, Route Handler, Server Action, datastore, market-data, provider,
  worker, calculation, tax, mobile, and  scope out of the slice.


- Tightened detail-request invalidation so stale detail responses cannot restore removed selection
  state after search reset.
- Moved focus into the detail region during loading and separated polite status announcements from
  assertive error alerts.
- Aligned successful asset-detail typing with the frozen `Asset` contract and expanded frontend
  tests for detail generation, accessibility helpers, and component accessibility wiring.
- Aligned follow-up detail typing with frozen `SourceReference`, `AssetStatus`, and optional bond
  coupon-rate fields; made successful detail copy distinct from deferred detail; and hardened focus
  tests against async outcome focus stealing.
- Preserved Escape/focus behavior for same-ticker detail retries while keeping async detail outcomes
  from stealing focus.

## 2026-07-27 — Stage 3.15 Web asset discovery UI slice closed

Canonical record: PR #41; commit(s) `22bede651a646d0e8b06568bda457d0626891e63`.

## 2026-07-27 — Stage 3.16 repository audit planning started

Canonical record: PR #42; commit(s) `9eec98c36d7aeffb21dc2d7e7e0eb1681106901d`.

## 2026-07-27 — Stage 3.16 technical reassessment findings fixed

- Added a required immutable post-planning audit target SHA and tracked-file coverage manifest for
  the future repository audit stage.
- Required every tracked path to be audited or narrowly excluded with a reviewed generated,
  vendored, binary, or archival rationale.
- Made SOLID, cost, and ADR-consistency review explicit in audit scope, evidence mapping, and
  acceptance criteria.

## 2026-08-04 — Stage 3.16 repository audit fixes started

Canonical record: PR #43; commit(s) `74eebe9ec8231764f21ce384c4690d073d0273da`.

## 2026-08-08 — Stage 3.16 audit-fix hardening completed pending review

- Removed the import review-token static fallback and require a configured 32-byte-or-longer secret
  outside explicitly named local/development environments.
- Added source-aware transaction provenance, strict SHA-256 import source validation, and coverage
  for manual/import duplicate policy and shared append serialization.
- Replaced client-controlled list offsets with signed opaque keyset cursors scoped to the
  authenticated query and deterministic database anchors; frontend forwards them without parsing.
- Made a missing database URL fail closed unless `OPENINVEST_ENV` explicitly declares development or
  local, and aligned runtime import-token/hash validation with OpenAPI bounds and lowercase digests.
- Extended migration CI to validate down migrations and rehearse every rollback, full reapply, and
  schema assertions in disposable PostgreSQL.
- Fixed import retry replay after ledger mutation, anchored transaction keysets on internal entry IDs,
  restored an explicit development-only token-secret fallback, applied the provenance migration in
  smoke, and scope-bound Web import async responses to the active portfolio/session.
- Published the durable 200-path Stage 3.16 audit coverage manifest and synchronized governance
- Signed asset-search continuations with query/type-bound HMAC keysets and synchronously invalidate
  stale import operations before passive effects run after session or portfolio changes.
- Hardened migration-rehearsal traversal and schema evidence, and added a mounted React lifecycle test
  for stale review and append promises after a portfolio or session change.

## 2026-08-09 — Stage 3.16 audit-fix closure

Canonical record: PR #44; commit(s) `9e6b8a753bf73ef020ce40461df25a5878344d92`.

## 2026-08-09 — Stage 3.17 privacy lifecycle planning started

- Added the documentation-only planning gate for account deletion, irreversible anonymization,
  backup destruction, and retention execution left open by the Stage 3.16 audit.
  operational evidence before any future implementation can begin.
- Kept all runtime, OpenAPI, SQL, infrastructure, market-data, financial-calculation, tax, mobile,
  and  changes out of scope.

## 2026-08-09 — Stage 3.18 privacy contract and security proposal started

Canonical record: PR #46.

## 2026-08-09 — Stage 3.19 privacy security and ADR proposal started

- Recorded Stage 3.18 as squash-merged through PR #47 at
  `4680e9c1b7b916169972c84ad8c3879955c7f509`.
- Added proposed ADR-008 and a documentation-only security dossier for provider-neutral
  cryptographic erasure, deletion markers, restore isolation, separation of duties, backup evidence,
  and partial-failure handling.
- Did not accept ADR-008 or change runtime code, OpenAPI, SQL schema, dependencies, providers,
  key-management configuration, backup operations, or product scope.

## 2026-08-09 — Stage 3.20 privacy threat-model proposal started

- Recorded Stage 3.19 as squash-merged through PR #48 at
  `fdf74c16446e7623f76882aa7add64554141abc6`.
- Added a documentation-only threat-model proposal covering browser/session abuse, application and
  privileged-data access, key custody, marker correlation/availability, backup/restore, partial
  failure, evidence redaction, and indirect reidentification.
  providers, key-management configuration, backup operations, or product scope.

## 2026-08-09 — Stage 3.21 privacy data-inventory proposal started

- Recorded Stage 3.20 as squash-merged through PR #49 at
  `849d934906f878a6d79ba89e940e5ba470e64c09`.
- Added a conservative repository-derived inventory for identity, access, financial, import,
  analytics, audit, browser, code, and external privacy surfaces, including field-level disposition
  expectations and evidence owners.
- Identified stable audit actor IDs, free-form content, opaque event payloads, hashes, request/trace
  correlation, and unknown external persistence as completion blockers.
  key-management, backup operations, or product scope.

## 2026-08-18 — Stage 3.22 privacy key-custody and destruction-proof proposal started

- Recorded Stage 3.21 as squash-merged through PR #50 at
  `207325e0497cc2608b99366f7f840472d270b6ed`.
- Added a documentation-only, provider-neutral custody proposal covering separated authorities,
  one-way per-subject erasure-material states, non-identifying destruction proof, proof validation,
  restore gating, provider evaluation, and adversarial evidence requirements.
  schema, migrations, credentials, dependencies, backups, CI, operations, or product scope.

## 2026-08-18 — Stage 3.23 privacy deletion-marker control-plane proposal started

- Recorded Stage 3.22 as squash-merged through PR #51 at
Canonical record: commit(s) `5f42d32db1e045c23fb99a5af8f136b7a49e3bc2`.
  green GitHub Actions.
- Added a documentation-only control-plane proposal for a restricted, non-identifying marker,
  monotonic state transitions, signed restore snapshots, fail-closed replay, authority separation,
  and distinct 90-day backup versus ten-year minimum-audit evidence boundaries.
  schema, migrations, credentials, dependencies, backups, CI, operations, or product scope.


Canonical record: PR #52; commit(s) `f7f23bce33038f259c976db6375079c68209a7aa`.

## 2026-08-22 — Stage 3.27 import financial-identity remediation verified

Canonical record: PR #55; commit(s) `19a8abbb0c07ded7441839bfa99b538739e21fbc`.

## 2026-08-22 — Stage 3.27 final-review identity-transition correction

Canonical record: PR #55; commit(s) `c6c3a4c91a108426448a2bc230873ab9e479a335`.

## 2026-08-22 — Stage 3.27 closure governance

- Recorded exact-head CI #90 success on `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c`.
  P1-02 fallback/strong identity defect.
- Recorded explicit merge gate.
- Recorded PR #55 squash merge into `develop` at `6e8c806de857f844954f1db513487357dfe90187`.
- PR #58 records closure governance for Stage 3.27; once this record is canonical
  on `develop`, P1-02/P1-03/P1-04 are closed.
- Stage 3.27 scope remains P1-02/P1-03/P1-04 only.
- P1-01 and P1-05 remain separate future remediation.
- Stage 3.25 remains separate.

## 2026-08-23 — Stage 3.28 authentication security remediation closure

- Squash-merged implementation PR #59 into `develop` at `dc83f5f3a11da164e6809593861d96ccf47b29ca` for repository-audit P1-01 and P1-05.
- Exact implementation head `92edab5d3e93dafe2fcc6247644e38e878a4202f` passed GitHub Actions CI #114.
- Closed refresh-token replay/session-family containment with persisted family identity, serialized mutation paths, replay containment, legacy fail-closed behavior, and PostgreSQL insertion defense.
- Closed the Argon2 resource-exhaustion finding without weakening 64 MiB / t=3 / p=1 by adding a process-wide fail-fast two-operation admission gate and strict stored-encoding budget checks.
- Dedicated `ErrAuthCapacity` HTTP `503 + Retry-After` semantics remain optional non-blocking contract hardening.
- Closure governance is recorded through PR #60 when canonical on `develop`.
- Stage 3.25 and all P2/P3 findings remain separate.

## 2026-08-23 — Stage 3.29 input and contract hardening

- Squash-merged implementation PR #61 into `develop` at `7331d3f34783baec3997497d1a79b78eaa558bd4`.
- Remediated P2-05 malformed-decimal HTTP semantics, P2-06 note-length boundary drift, P2-07
  `NUMERIC(28,8)` ingress/derived/aggregate persistence bounds, P2-08 permissive financial JSON
  commands, and P2-15 duplicate normalized CSV headers.
- Exact-head CI #124 completed successfully on
  `f9e70e70956c76edbc2ab02c52d45124b2dea525`.
  failing closed; the final implementation added same-transaction PostgreSQL range admission and
  atomic rollback integration coverage, then received renewed independent `APPROVED`.
- Detailed engineering rationale and rejected alternatives remain in the Stage 3.29 report.
- Closure governance is tracked through PR #62. Once canonical, these five P2 findings are closed;
  12 P2 and 10 P3 findings remain. Stage 3.25 privacy work remains separate.

## 2026-08-23 — Stage 3.30 import review integrity

Canonical record: PR #63, PR #64; commit(s) `8f68dd18800918e6a9882e995e13dba2723dc929`, `2f788e0811d78c9def0502676a74bee2f9922bf5`.

## 2026-08-23 — Stage 3.31 authentication operational hardening

- Squash-merged implementation PR #65 into `develop` at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`.
- P2-01 now routes logout through bounded auth admission before rejected-auth audit persistence.
- P2-14 now bounds active limiter keys, per-key attempts, and global downstream auth attempts per window,
  with expired-bucket reclamation.
- Logout OpenAPI now advertises the existing 429 `RateLimited` response.
Canonical record: commit(s) `82557c55c0772a66707088b858ec9eafc2073119`.
- Closure governance was squash-merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746`; P2-01/P2-14 are closed and 7 P2 plus 10 P3 findings remained. Stage 3.25 remains separate.

## 2026-08-23 — Stage 3.32 exact idempotency replay and browser retry recovery

Canonical record: PR #67; commit(s) `0623d5ef326cd783b7dc0417dbcb02f18c506171`, `02aa2417a3caca79e2afc4e7b598b92055de96b7`.

## OI-NEW P2 remediation closure — 8/8

Canonical state: `develop@d9f6c263dd3ff4d22fc978372014381880104489`.

Completed technical corrections:

- PostgreSQL runtime least privilege;
- auth/storage error distinction;
- readiness/startup integrity-check separation;
- HTTP server timeouts;
- trusted proxy and canonical client-IP semantics;
- bounded retroactive replay with exact weighted-average cost;
- frontend transient-auth state preservation;
- Portfolio Summary `asOfDate` contract/runtime alignment.

P2 is closed `8/8`.
