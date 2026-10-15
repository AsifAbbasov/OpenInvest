# OpenInvest Source of Truth
Document ID: SOT-001; Version: 1.5.4; architecture authority remains Documents 42–43 and accepted ADRs; this edit consolidates documentation without changing policy or implementation.

## Authority and verified state
Document priority: `Document 43 → Document 42 → accepted ADRs → Documents 1–41 → README → comments`.

[OpenAPI](../openapi/openapi.yaml) owns the executable HTTP contract; [accepted ADRs](ADR/) own architectural rationale; [Data Source Registry](registries/DATA_SOURCE_REGISTRY.md) alone governs source use. [Roadmap](ROADMAP.md) owns sequencing; [Implementation Log](IMPLEMENTATION_LOG.md) and [stage dossiers](stages/) retain chronology; [Document Index](DOCUMENT_INDEX.md) routes readers.
Architecture Freeze v1.2 and documentation freeze remain active. The numbered implemented frontier is Stage 3.77; this cleanup authorizes no Stage 3.78+ work, deployment or new audit campaign.
[Module #1 technical acceptance/integration](audit/MODULE1_TECHNICAL_CLOSURE.md) is PASS through PR #268 and exact post-merge CI #38758661473. This is not a new formal module-closure decision. Historical JWT RED evidence and accepted JWT02 residual remain unchanged.
[Module #12 scoped remediation/revalidation](audit/FINAL_PRODUCTION_FRONTEND_INFRA_HARDENING_ADVERSARIAL_AUDIT_CLOSURE.md) covers shared Redis provider budgets, local operation ownership and ambiguous Origin rejection; its embedded candidate flags are dated evidence, not present-day deployment approval. Actual edge/CDN behavior, account quotas, multi-host topology, TLS/HSTS, external logs and deployed PostgreSQL/network policy remain deployment evidence gaps.

## Product and architecture boundaries

OpenInvest is a privacy-first Personal Capital Operating System, not a broker, bank, trading terminal, investment adviser or tax service. PostgreSQL owns canonical financial data; clients do not invent unavailable market facts.
The first public MVP targets long-term, dividend, FIRE and multi-account investors needing independent, explainable real-return accounting; it is not optimized for casual green/red brokerage-return badges.

- Monorepo; Go/Fiber API (toolchain pinned by [go.mod](../backend-go/go.mod)); Python analytics managed with uv; PostgreSQL logical domains `identity`, `investment`, `analytics`, `tax`, `audit`; Redis and process-local cache only where approved.
- Next.js App Router/TypeScript/pnpm is the Web presentation layer under ADR-007: Browser → OpenAPI Go API → PostgreSQL/Redis/future workers; no direct client datastore or external market-data access, no LocalStorage business data. iOS SwiftUI/Android Jetpack Compose remain future scope.
- API First, DDD, Clean Architecture, Event Driven; immutable ledger, versioned rebuildable snapshots; approved outbox/inbox processing uses at-least-once delivery with idempotent business handling; Zero Trust and Privacy by Design.
- Planned MVP scope includes accounts, portfolios, transactions, stock/bond cards, dividend calculator/calendar, snapshots, WAC, XIRR, real/inflation-adjusted return, purchasing-power insight and dashboard. Scope membership is not a claim every planned feature is implemented.
- [Backlog exclusions](BACKLOG_V2.md) and [product guardrails](product/MVP_PRODUCT_RISK_REFINEMENT.md) remain binding: no unapproved assistant/scenario/premium/Tax XML/email/forecast/family/public API/foreign-market/mobile expansion. Purchasing Power is secondary to capital/return/dividends; tax truth needs deterministic financial/legal test vectors.
- Every fifth completed stage requires the established full repository line-by-line audit gate. Independent review, exact published-head CI and separate merge authority remain distinct; this document does not waive the review workflow.

## Active financial and API invariants

- [ADR-009](ADR/ADR-009-deterministic-portfolio-ledger-ordering-and-wac.md): deterministic portfolio-local ledger order; manual SELL/WAC/acquisition basis; reject oversell atomically; backdated rebuild/snapshot methodology does not authorize imported SELL expansion or tax/bond-NKD semantics.
- Open STOCK/BOND positions expose quantity, authoritative WAC, remaining/total acquisition basis and independent Half-Even basis weights; deterministic ticker order, endpoint-local BusinessDate cutoff, ownership isolation and repeatable-read replay. Omit closed positions; inactive catalog status does not hide owned holdings; existing summary/snapshot-position contracts remain distinct.
- Current/historical views reuse the positions `asOfDate` contract. Corrections/reversals append revisions, preserve original rows and derive effective-ledger truth; historical oversell fails closed.
- Cash/income uses effective-ledger cash flows, gross recorded dividend/coupon income, fees/taxes and monthly breakdown; expected provider events are not realized ledger income.
- Manual valuation accepts only authenticated `USER_SUPPLIED` RUB prices, honest coverage, lifecycle-safe position generations and exact-date historical eligibility. Provider-derived/live valuation remains unavailable; no inferred prices or market weights when required data are absent.
- XIRR uses DEPOSIT/WITHDRAWAL external flows, exact-date COMPLETE manual terminal valuation, ACT/365 and scale-8 Half-Even output; multiple roots/numerical failure fail closed. TWR, inflation/real-return implementation, FX and provider activation need separate gates.
- Input boundaries retain exact ASCII Decimal grammar before big.Int work, `NUMERIC(28,8)` bounds including aggregate/snapshot metrics, 500-character notes, unknown-JSON-field rejection and Unicode-code-point length semantics; binary-float financial truth is forbidden.
- Timezones preserve exact resolver-backed identifiers; reject `Local`, surrounding whitespace and raw `±HH:MM`/`UTC±HH:MM`, retain UTC and loadable identifiers such as `Etc/GMT+4`; tzdata is fallback availability, not normalization. BusinessDate and financial calculations remain independent.
- CSV review rejects duplicate normalized headers before normalization and stops on the 101st data row. User-approved append is atomic; raw-file/import-session persistence and automatic append are not implied.
- Signed review proofs bind versions, subject, portfolio, source/file hash, every row identity, normalized/final semantics and the reviewed APPENDABLE set; reparse exact input and fail closed on expiry/drift. Locked storage remains final duplicate/identity authority. Parser-v1 proofs cannot authorize fresh writes; authentic completed legacy artifacts permit only scoped read-only recovery.
- Persist original versioned response bytes/status/request/trace identity atomically with financial mutation; validate body integrity, serialize duplicate command resolution and survive new Store connections. Missing legacy artifacts fail closed; completed replay does not reread mutable resources.
- Browser retry journals contain only short-lived opaque metadata in sessionStorage, keyed by SHA-256 of stable principal/operation/optional portfolio. Retain unresolved retry identity across reload, rotate changed intent, clear confirmed success/conflict; never store raw identity, financial payload, CSV, review/auth tokens.
- PostgreSQL owns exact `snapshotDatesRebuilt`; compute one sorted deduplicated affected-date union and rebuild each once under the same-portfolio lock, including applicable later snapshots.
- Command/replay authority expires after 24 hours, with fresh DB clock after serialization; exact-key generation reclamation and session-expiry authority are preserved. Revoked-but-unexpired refresh containment remains; bounded cleanup uses at most 128 rows, `ORDER BY expires_at,id`, `FOR UPDATE SKIP LOCKED` in triggering transactions.

## Financial standard

- Decimal only; binary float forbidden for financial values.
- Half-even rounding.
- Eight decimal places internally; two for monetary display.
- SQL `DATE` for financial business dates.
- UTC `TIMESTAMP WITH TIME ZONE` for system timestamps.
- MOEX calendar for MVP market events.
- Canonical financial test vectors are mandatory before production algorithms.

## Privacy definitions — approved target, separately gated lifecycle

- **Personal Data:** information that identifies a person directly or indirectly.
- **Pseudonymized Data:** data linked through a reversible identifier.
- **Anonymous Data:** data that cannot be linked back to an individual by any reasonable technical or organizational means.

Deleting a user removes identity data and irreversibly destroys its link to the financial ledger. The detached ledger becomes **Anonymous Financial History**. OpenInvest retains no re-identification mechanism.

## Retention — approved policy, not a deployment verification

- Identity/personal data: deleted completely after the approved deletion lifecycle.
- Audit: 10 years.
- Anonymous transactions and snapshots: no fixed expiration.
- Backups: maximum 90 days, then automatic destruction.

Privacy lifecycle design is not proof that deletion, cryptographic erasure, key custody or restore controls are deployed; [ADR-008](ADR/ADR-008-privacy-lifecycle-erasure-and-restore.md) and Stage 3.25 evidence retain their separate approval/implementation boundaries.

## Authentication, runtime permissions and delivery

- Public register/resend accepts email only with generic acknowledgment; proof completion creates a new verified account atomically without automatic login. [Approved email-ownership addendum](ADR/ADR-010-email-ownership-approval-candidate.md) and current OpenAPI retain full proof, generation/lease, attempt/timeout, legacy, frontend, grant, retention and rollback requirements; the original migration-pinned ADR-010 is unchanged.
- [JWT identity policy](api/ACCESS_JWT_IDENTITY_POLICY_V1.md): mandatory nonempty `sub`, `uid`, `investmentSubjectId`; exact `sub=uid`, no fallback; financial authorization uses the validated signed investment subject. Existing HS256/algorithm/type/expiration checks remain; JWT02 Option A accepts the documented future-iat residual, JWT03 remains informational; access JWTs are stateless, with no new immediate revocation feature.
- Refresh-family identity, serialized rotation/logout, replay containment, conservative legacy handling and PostgreSQL defense remain. Auth admission precedes logout decode/service/rejected-event persistence; capacity exhaustion fails closed with the existing 429 contract.
- Argon2id stays 64 MiB/t=3/p=1; hash/verify/dummy work shares fail-fast two-operation capacity and rejects over-budget stored encodings before expensive work. Auth limiter bounds per-key/global attempts and active cardinality with expired-bucket reclamation; provider shared Redis budgets are a separate boundary.
- Runtime capability/credential-graph checks reject privileged/masked identities, protected-schema CREATE, protected-table ownership/mutation, SET-reachable escalation and ADMIN OPTION paths. Current additive email-ownership grants retain their explicit envelopes; no broad ledger/audit mutation rights are inferred.
- Verification proof is 32 crypto-random bytes/canonical unpadded Base64URL, digest-only SHA-256 in PostgreSQL, purpose `new_registration`; intent lasts 24h, proof expires at min(30min, intent expiry). Same-generation valid proofs survive resend; renewal increments generation, replaces old job ownership and resets attempts only for the new generation; stale/expired callbacks cannot complete a new lease.
- At most four pre-counted delivery attempts per generation; one-minute lease/resend suppression, five-minute failed retry, terminalize an exhausted expired lease without a fifth proof. Parent intent locks precede jobs/tokens; shared intake admission caps active pending intents at 10000.
- Activation transaction binds intent email/proof/generation and creates user, credential, financial link, privacy defaults, verification and audit atomically; consumes outstanding proofs/cancels delivery, never logs in. Existing nondeleted accounts cannot be reclaimed; legacy users are not silently reverified, migrated or granted new ownership.
- DB lock/statement limits are 2s/4s with 5s intake/completion DB context; bounded Argon2 work precedes that context. Delivery context is 30s, detached finalization 5s; SMTP requires STARTTLS/TLS≥1.2, finite connection deadlines and HTTPS proof URLs, opt-in worker and separately approved sender/DB configuration. Remote SMTP work is not forcibly terminated; delivery is at-least-once.
- Verification UI keeps proofs in memory, scrubs fragments, fences token-context/stale responses and survives StrictMode replay; 250ms polling covers silent history changes with synchronous submission/response checks. No localStorage/sessionStorage secret persistence or automatic authentication.
- Cleanup removes at most 100 intents expired over seven days per intake, not an idle retention SLA. Coordinated frontend/API/schema/runtime-grant rollout is required; grants are explicit/table-level where defined. Rollback closes admission through approved routing/maintenance, retains additive schema/evidence, never restores unverified activation or blindly rolls back to incompatible privilege guards; no dedicated registration-disable flag is claimed.
- Module #1 does not authorize SMTP activation, account recovery, immediate access-token revocation or deployment. Local development bypass/insecure-cookie flags require development/local mode; production authorization uses signed access tokens.

## Providers and public projections

- QuoteProvider → canonical MarketQuote → application service preserves exact RUB Decimal, UTC `AsOf`/`RetrievedAt`, provenance and deterministic FRESH/STALE/UNKNOWN validation. The delayed MOEX TQBR adapter is implemented but dormant; asset search keeps `lastPrice: null` without an approved runtime source and detail waits for complete provenance.
- CorporateActionProvider/CorporateActionEvent supports validated provenance, correction/cancellation linkage, deterministic dated Calendar and count/density-only Heatmap. Source unavailable differs from a valid empty result; private SourceEventID is not public, supersession integrity fails closed, responses are no-store, and stale/unmounted/replaced frontend requests are bounded/cancelled.
- Feature 3D implementation is complete only for `TINVEST_CORPORATE_ACTIONS_CONSTRAINED`; source/use CONDITIONAL-GO is not runtime activation. Only GetDividends/DIVIDEND and GetBondCoupons/COUPON; server-side read-only token, no raw redistribution, persistence, cross-request provider cache, background polling or automatic ledger mutation; retain FAQ/support conflict and exact registry limits.
- Repository shared rolling limits retain endpoint global 48/60s, client 12/60s and provider 60/60s with at most four local operations; Redis failure fails closed. Caller cancellation cannot recycle capacity before independent local operation termination; arbitrary remote server computation termination is not guaranteed.
- The public dividend calculator uses user-supplied ticker/quantity/dividend/unit-cost and backend exact Decimal gross/yield; no provider, tax calculation or portfolio mutation is implied.

## Governance and existing detailed records

The original 32-finding audit and OI-NEW P2 set have their bounded closure records in [original audit register](audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md) and [OI-NEW closure](audit/OI_NEW_P2_REMEDIATION_CLOSURE.md); no new global finding inventory is asserted here. Protected develop requires ten CI contexts, PR entry, conversation resolution, linear history, no normal admin bypass/force-push/deletion and squash-only merge policy; this edit changes no settings.
Accepted ADR-001–007 and ADR-009 retain Documents 42–43 precedence; approved ADR-010 addendum is separate from its historical hash-pinned original. [Version Matrix](VERSION_MATRIX.md), [Open Questions](OPEN_QUESTIONS.md), [stage dossiers](stages/) and [Implementation Log](IMPLEMENTATION_LOG.md) retain full decisions, exact commits and test outcomes. Duplicated prose is removed, not replaced by Git-only contracts.
The cleaned Stage 3.54 plan hash remains `f2b598c396bdcdda4168fca9b7ef51db0f3b154b3acfc2abc06501cffba897b1`; earlier `c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba` is historical, with unchanged machine-contract cardinalities/semantics.

## Historical evidence and protected snapshots

Detailed outcomes remain in the current [implementation log](IMPLEMENTATION_LOG.md), [original audit register](audit/REPOSITORY_AUDIT_REMEDIATION_REGISTER.md) and stage dossiers: [import identity](stages/STAGE_03_27_IMPORT_FINANCIAL_IDENTITY_REMEDIATION.md), [auth/session](stages/STAGE_03_28_AUTH_SECURITY_REMEDIATION.md), [CI/governance](stages/STAGE_03_34_CLOSURE.md), [timezones](stages/STAGE_03_37_IANA_TIMEZONE_SEMANTICS_CLOSURE.md), [retention](stages/STAGE_03_38_IDEMPOTENCY_SESSION_RETENTION_CLOSURE.md), [Next.js](stages/STAGE_03_42_NEXTJS_SECURITY_MAINTENANCE_CLOSURE.md), [Fiber](stages/STAGE_03_45_FIBER_MAINTENANCE_CLOSURE.md), [HTTP decomposition](stages/STAGE_03_48_HTTPAPI_DECOMPOSITION_CLOSURE.md), [validator](stages/STAGE_03_56_P3_08_MIGRATION_VALIDATOR_CLOSURE.md) and [provider boundary](stages/STAGE_03_58_MARKET_DATA_PROVIDER_BOUNDARY_CLOSURE.md). Their historical failures and review iterations are not relabeled PASS.
The earlier Stage 3.34 closure draft `b208ddfdca9fcc3ddb3589fb7d47c698fa9f0dfa` received `changes required` for canonical-status drift; the linked closure dossier records the corrected final disposition.

| Historical verification | Exact source | Original result / current dossier |
| --- | --- | --- |
| Stage 3.37 closure | `dc206d9d9cf68b00c182ce1a247e5dcc289e16e8`; [CI 32880893193](https://github.com/AsifAbbasov/OpenInvest/actions/runs/32880893193) | SUCCESS 10/10; [timezone closure](stages/STAGE_03_37_IANA_TIMEZONE_SEMANTICS_CLOSURE.md) |
| Stage 3.45 closure | `e91faace81b76b14a51de4dea3a4c18d697998af`; [CI 33315057112](https://github.com/AsifAbbasov/OpenInvest/actions/runs/33315057112) | SUCCESS 10/10; [Fiber closure](stages/STAGE_03_45_FIBER_MAINTENANCE_CLOSURE.md) |
| Stage 3.59 implementation | `8d3c23cdb257ea2134d9f0ebc7b11fa264144611`; [CI 33900806691](https://github.com/AsifAbbasov/OpenInvest/actions/runs/33900806691) | SUCCESS 10/10; [MOEX adapter](stages/STAGE_03_59_MOEX_ISS_QUOTE_PROVIDER_IMPLEMENTATION.md), production activation remains separate |
| Stage 3.72 implementation | `b6cd7766a6a584440dfdd7d00c35c34c8553a9ff`; [CI 34091588982](https://github.com/AsifAbbasov/OpenInvest/actions/runs/34091588982) | SUCCESS 10/10; [position projection](stages/STAGE_03_72_PORTFOLIO_POSITION_PROJECTION_IMPLEMENTATION_CLOSURE.md) |

Earlier documentation commit identities remain recorded for traceability: Stage 3.6 closure `fb651632036fabaa31ec92e9d28b5782ca0f92e5`, Stage 3.8 closure `cb9a392eb90ede954d9cc68b247bada13a1540d9`, and Stage 3.36 governance `7fd4e549e396cd9d99dff5c6d1678d0a4099fb70`. These references do not create new approval or runtime authority.
The activation snapshots below retain their original bytes and historical outcomes. Embedded OPEN/CLOSED/candidate values apply to the cited activation point, not a new current release decision.

<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_52_BEGIN CURRENT_AUTHORITY=NO -->
<!-- OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1
CANONICAL_WORKFLOW_BEFORE_ACTIVATION=1.3.0
PROPOSED_WORKFLOW=1.4.0
AMENDMENT_STATUS=CANONICAL_ACTIVATED
ACTIVATED_BY_PR=111
ACTIVATED_MERGE_SHA=93e59cbf4821fc51aba5bdb9815b52a73fbc67a0
ACTIVATED_TREE=3686ff3606d7c5f4fe97060abc12dffd0ccd3477
EFFECTIVE_WORKFLOW=1.4.0
ADOPTION_PATH=V1_3_POST_DEVELOPMENT_GOVERNANCE
SELF_BOOTSTRAP_NEW_RULES=FORBIDDEN
P2_GOV_01=UNRESOLVED_BLOCKER
P2_GOV_02_TO_05=REMEDIATED_IN_STAGE_03_51_V6_REVIEW
P3_07_STATE=OPEN
STAGE_03_51_PUBLICATION_ELIGIBILITY=BLOCKED
P3_08_STATE=OPEN_UNAFFECTED
CURRENT_AUDIT_CLOSED=30/32
CURRENT_AUDIT_PERCENT=93.75%
NEXT_AFTER_AMENDMENT=SEPARATE_P2_GOV_01_DISPOSITION
DISPOSITION_STAGE=3.53
THEN=REVISE_AND_REREVIEW_STAGE_03_51
<!-- OPENINVEST_STAGE_03_52_WORKFLOW_AMENDMENT_STATE_V1_END -->
<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_52_END -->

<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_53_BEGIN CURRENT_AUTHORITY=NO -->
<!-- OPENINVEST_STAGE_03_53_P2_GOV_01_DISPOSITION_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_53_P2_GOV_01_DISPOSITION_STATE_V1
CANONICAL_WORKFLOW=1.4.0
WORKFLOW_ACTIVATION_PR=111
WORKFLOW_ACTIVATION_MERGE_SHA=93e59cbf4821fc51aba5bdb9815b52a73fbc67a0
DEVIATION_ID=P2-GOV-01
DEVIATION_CLASS=IRREVERSIBLE_HISTORICAL_GOVERNANCE_LIFECYCLE
AFFECTED_STAGE=3.50
AFFECTED_PR=110
AFFECTED_PUBLISHED_HEAD=be774b3a8423ffba98633b257983856b2c990b95
AFFECTED_MERGE_SHA=915d42f614121959fface9846a07cc1b412febe2
VIOLATED_WORKFLOW=1.3.0
HISTORICAL_COMPLIANCE=NONCOMPLIANT_PRESERVED
PRE_MERGE_BLOCKER_STATE=UNRESOLVED_BLOCKER
POST_MERGE_EFFECTIVE_STATUS=DISPOSITIONED_HISTORICAL_NONCOMPLIANCE_PRESERVED_RESIDUAL_GOVERNANCE_RISK_ACCEPTED
RISK_ACCEPTANCE_REQUIRED_BEFORE_MERGE=TRUE
RISK_ACCEPTANCE_RECORD=DESIGNATED_REVIEW_CHAT_BOUND_TO_EXACT_PUBLISHED_HEAD
MERGE_AUTHORIZATION_REQUIRED_SEPARATELY=TRUE
RECORD_SELF_ASSERTS_RISK_ACCEPTANCE_OR_MERGE_AUTH=NO
ACTIVATION_CONDITION=THIS_EXACT_DISPOSITION_RECORD_ON_PROTECTED_DEVELOP_AFTER_REQUIRED_GATES
P3_07_STATE=OPEN
STAGE_03_51_PRE_DISPOSITION=BLOCKED
STAGE_03_51_POST_DISPOSITION=ELIGIBLE_FOR_REVISION_AND_REREVIEW_NOT_CLOSED
CURRENT_AUDIT_CLOSED=30/32
CURRENT_AUDIT_PERCENT=93.75%
DISPOSITION_CHANGES_ORIGINAL_AUDIT_ARITHMETIC=NO
P3_08_STATE=OPEN_UNAFFECTED
<!-- OPENINVEST_STAGE_03_53_P2_GOV_01_DISPOSITION_STATE_V1_END -->
<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_53_END -->

<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_51_BEGIN CURRENT_AUTHORITY=NO -->
<!-- OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1
CANONICAL_WORKFLOW=1.4.0
CLOSURE_BASE=ea1f204eab47bf16566096722d6390557b8141af
CLOSURE_BASE_TREE=488cce09e87d42b3e5e03441336197bd10228c51
STAGE_03_49_PLAN_PR=109
STAGE_03_49_PLAN_MERGE_SHA=cfcc384a97327cc8b74aa05567b9629abf40a5fb
STAGE_03_50_IMPLEMENTATION_PR=110
STAGE_03_50_IMPLEMENTATION_HEAD=be774b3a8423ffba98633b257983856b2c990b95
STAGE_03_50_IMPLEMENTATION_MERGE_SHA=915d42f614121959fface9846a07cc1b412febe2
P2_GOV_01_STATUS=DISPOSITIONED_HISTORICAL_NONCOMPLIANCE_PRESERVED_RESIDUAL_GOVERNANCE_RISK_ACCEPTED
P2_GOV_01_DISPOSITION_PR=112
P2_GOV_01_DISPOSITION_MERGE_SHA=ea1f204eab47bf16566096722d6390557b8141af
HISTORICAL_COMPLIANCE=NONCOMPLIANT_PRESERVED
DISPOSITION_DOES_NOT_RETROACTIVELY_COMPLY=TRUE
P2_GOV_02_TO_05=REMEDIATED_AND_RETAINED
P2_GOV_06=REMEDIATED_AND_VERIFIED
CLOSURE_PR=113
CLOSURE_PUBLISHED_HEAD=1868749965fa1c113875592668afaa8f1f1ca35e
CLOSURE_PUBLISHED_TREE=c83120dcfdf1cab281fea10818860827eb699b64
CLOSURE_CI_RUN_NUMBER=309
CLOSURE_CI_RUN_ID=33540615693
CLOSURE_CI=10/10_SUCCESS
EXACT_PUBLISHED_HEAD_REVIEW=APPROVED
CLOSURE_MERGE_SHA=072350205b2746bdcd83f20718eb59efcd0478ef
CLOSURE_MERGE_TREE=c83120dcfdf1cab281fea10818860827eb699b64
P3_07_STATE=CLOSED
AUDIT_CLOSED=31/32
AUDIT_PERCENT=96.875%
REMAINING_ORIGINAL_FINDING=P3-08
P3_08_STATE=OPEN_UNAFFECTED
RUNTIME_CHANGE=NONE
FORENSIC_NARRATIVE=STAGE_03_51_DOCUMENT
BRANCH_DELETION_AUTHORIZED=NO
<!-- OPENINVEST_STAGE_03_51_P3_07_POST_MERGE_STATE_V1_END -->
<!-- OPENINVEST_HISTORICAL_ACTIVATION_TIME_SNAPSHOT_STAGE_03_51_END -->

<!-- OPENINVEST_STAGE_03_56_P3_08_CLOSURE_STATE_V1_BEGIN -->
SCHEMA=OPENINVEST_STAGE_03_56_P3_08_CLOSURE_STATE_V1
CANONICAL_WORKFLOW=1.4.0
CLOSURE_BASE=6a443969aef944bde0946d36c79f67ddb87c28fe
CLOSURE_BASE_TREE=6d894f710329332f2f64b7d280a9b27a94be86d9
STAGE_03_54_PLAN_PR=115
STAGE_03_54_PLAN_MERGE_SHA=b79a9d3c43621e56e598901bdf472771e8b68ef8
STAGE_03_54_PLAN_BLOB=90fa563b9256b19055e2c14e52909596b392f221
STAGE_03_54_PLAN_SHA256=c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba
STAGE_03_55_IMPLEMENTATION_PR=116
STAGE_03_55_PUBLISHED_HEAD=9df58319b59a1bd3ab9817d07d59c3b3c36a1b1a
STAGE_03_55_PUBLISHED_TREE=6d894f710329332f2f64b7d280a9b27a94be86d9
STAGE_03_55_MERGE_SHA=6a443969aef944bde0946d36c79f67ddb87c28fe
STAGE_03_55_MERGE_TREE=6d894f710329332f2f64b7d280a9b27a94be86d9
STAGE_03_55_CI_RUN_NUMBER=315
STAGE_03_55_CI_RUN_ID=33799997370
STAGE_03_55_CI=10/10_SUCCESS
STAGE_03_55_PUBLISHED_EXACT_HEAD_REVIEW=APPROVED
HISTORICAL_SQL=14/14_BYTE_IDENTICAL
PLAN_REGISTER=18/18_PASS
H_REGISTER=269/269_PASS
SOURCE_ANCHORS=82/82_EXACT
S2_REGISTER=168/168_EXACT
P3D_REGISTER=27/27_EXACT
FINITE_DOMAINS=19/19_EXACT
RULE_REGISTER=75/75_EXACT
TC_EXECUTION_LEDGER=631/631_EXACT
TC_MISSING=0
TC_DUPLICATE=0
MPROP_REGISTER=631/631_EXACT_TC_BIJECTION
PROPERTY_MUTATIONS=1895/1895_KILLED
EXTRA_RED_TEAM=31/31_KILLED
TAXONOMY_AUTHORITY=TA-01_SINGLE_SOURCE
OBSERVER_CLOSURE=89+79=168_EXACT
ALLOWED_BRANCHES=89_EXACT_ZERO_UNMAPPED
PRE_MERGE_P3_08_STATE=OPEN
POST_MERGE_P3_08_STATE=CLOSED
CLOSURE_ACTIVATION=THIS_EXACT_STAGE_03_56_RECORD_AND_SYNCHRONIZED_STATE_ON_PROTECTED_DEVELOP_AFTER_REQUIRED_GATES
PRE_MERGE_AUDIT_CLOSED=31/32
PRE_MERGE_AUDIT_PERCENT=96.875%
POST_MERGE_AUDIT_CLOSED=32/32
POST_MERGE_AUDIT_PERCENT=100%
POST_MERGE_REMAINING_ORIGINAL_FINDING=NONE
POST_MERGE_P0=0
POST_MERGE_P1=0
POST_MERGE_P2=0
POST_MERGE_P3=0
RUNTIME_CHANGE=NONE
SCHEMA_DATA_SQL_OPENAPI_FRONTEND_DEPENDENCY_CHANGE=NONE
REMOTE_MUTATION_AUTHORIZED_BY_RECORD=NO
BRANCH_DELETION_AUTHORIZED=NO
<!-- OPENINVEST_STAGE_03_56_P3_08_CLOSURE_STATE_V1_END -->
