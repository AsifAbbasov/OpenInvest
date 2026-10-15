# Module #1 — architecture acceptance and integration readiness

Current owner disposition: [approved closeout decisions](MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md). Earlier proposed/pending statements below record the pre-approval package and are superseded only as to these three decisions; historical evidence and residuals remain unchanged.

Status: OWNER_DECISION_PACKAGE; MODULE1_CLOSED=NO. Prepared 2026-10-15.
No production implementation, test or workflow change is made by this package.
The original ADR010 is preserved byte-identically as migration 000015 authority;
the approval-ready revision is a separate prospective addendum candidate.

## Frozen implementation and protection

| Object | SHA | Tree |
| --- | --- | --- |
| Protected develop | 8d9998590fba6c605a64ddcc8429f51df82b56c3 | bed3b0b3466de160c6bb954b41b373e5ed7ea1ba |
| PR251 implementation | 0ec890e469efa87bb552f02b7acd66fc2fee0abc | b445bc2fb515ca18effdb4d167bc86682b5f44ef |
| PR266 accepted JWT01 implementation | e982d5a68e083a9e65b7f80b902f746b1b408bcb | 23cb0831d37c2a46dcdeaa96e21d065015d53bca |
| Independent source PR260 | bcfbc2687b60c68b33a8c4b0fbe969952c649d7d | f0d56fb63076a8eaf32bc1f96fe0cdc4928bea96 |

Develop remains protected with ten required contexts. PR251 and PR260–266 are
OPEN/DRAFT/UNMERGED at preparation. Their heads are not modified. The owner-provided
closeout directive records ACCEPT_JWT01_REMEDIATION; this package does not invent
an additional GitHub review event or repeat that accepted review.

## Documentation package and contract reconciliation

- [ADR010 candidate](../ADR/ADR-010-email-ownership-approval-candidate.md):
  Proposed, approval-ready; exact API, transaction, generation/lease, legacy and release boundaries.
- [Approved JWT01 rule](../api/ACCESS_JWT_IDENTITY_POLICY_V1.md): owner-approved policy,
  without marking the broader email-ownership ADR accepted.
- [JWT02/JWT03 decision record](MODULE1_JWT02_JWT03_POLICY_DECISION.md): recommended
  Option A pending owner, Option B explicitly separate, accepted informational parser scope.
- [PR265 provenance](MODULE1_PR265_EVIDENCE_PROVENANCE.md): separate post-deadline
  authorization, exact run outcomes, pending procedural disposition.
- [This integration-readiness record](MODULE1_ARCHITECTURE_ACCEPTANCE_AND_INTEGRATION_READINESS.md):
  consolidated ancestry/inventory, acceptance matrix, owner decisions and release gates.

The following inconsistencies/ambiguities are resolved in the candidate documentation,
without rewriting old evidence or modifying accepted ADR006:

| Prior ambiguity | Precise candidate meaning |
| --- | --- |
| Editing ADR010 would invalidate pinned migration authority | Preserve original bytes/manifest and propose a separate revision/addendum; first documentation CI 38718304601 correctly rejected the in-place edit with MIG016_AUTHORITY_REFERENCE |
| Stage2/ADR006 registration session issuance vs branch OpenAPI 202 | Proposed ADR010 supersedes only new public registration issuance after owner acceptance; protected baseline remains historical |
| Anonymous register taking password/preferences | Email-only intake; credentials/preferences selected during proof completion |
| ACCOUNT_READY mistaken for login | No session/cookie; explicit later normal login |
| Proof 'issued after persistence' | Raw proof generated transiently; only digest persisted before delivery |
| All retries invalidate prior proof | Same generation's valid proofs remain until expiry/consumption; generation renewal invalidates old generation |
| Fourth crash/old-generation lease stranded | REL01 terminalization and REL02 atomic generation fencing already implemented; no new code here |
| 'No session before proof' interpreted as disabling legacy users | New public activation only; explicit legacy fixture/internal method and existing login continuity preserved |
| 'Bounded retention' interpreted as idle cleanup guarantee | At most 100 old intents cleaned per intake; no idle retention SLA |
| Database errors interpreted as proof of no commit | Atomic storage does not guarantee the caller observed the outcome; lost ACK remains ambiguous |
| 'SMTP disabled' interpreted as absent transport implementation | Adapter exists; opt-in worker and production sender/TLS/operations approval absent |
| JWT01 acceptance interpreted as all JWT tests green | JWT02 historical RED remains; JWT03 unchanged; separate policy decisions |
| Historical CLOSED_WITH_P3_RESIDUAL interpreted as current formal closure | Historical report retained; current acceptance/architecture/integration gates remain open |

## Complete acceptance matrix for this closeout scope

PASS below means the cited bounded test scope, not all possible executions or current
production deployment. Historical independent branches are not relabeled as fresh
runs on PR266. No comprehensive campaign is rerun by this documentation task.

| Boundary | Disposition | Evidence and limit |
| --- | --- | --- |
| Anonymous registration creates no user/session | Tested; core candidate supported | PR251 ordinary ownership tests; PR264 real browser + PostgreSQL 38622757789 |
| Single-use digest proof, expiry/generation/atomic activation | Tested bounded scope | PR251 integration; PR260 38554843076; PR266 normal Go coverage |
| Registration renewal/abandoned lease and retry exhaustion | Targeted fixes retained | PR251 CI 38517528739; no new generation reset within one budget |
| Stale frontend proof/response, StrictMode, scrub | Targeted mounted regressions retained | PR251 tests inherited by PR266; not every real-router schedule certified |
| Actual browser activation + subsequent login, no auto-session | Tested local production runtime | PR264 38622757789 and PR265 38636806873; external SMTP/edge excluded |
| Existing account and financial ownership preservation | Tested/source-supported | PR263 active-owner checks 38623135638; current completion refuses non-deleted membership |
| Normal login, issuance, refresh financial identity | PASS on accepted candidate | PR266 38691771383 job 116279366055, real PostgreSQL handlers |
| Refresh replay, logout, cookie/CSRF, session/principal isolation | Historical tested scope + normal regressions | PR263 38623135638; PR265 CSRF/race jobs in 38636806790; not exhaustive schedules |
| Financial subject scoping and owner/outsider separation | PASS on candidate, historical supplementary evidence | PR266 bidirectional HTTP/PG check; PR263 finance bearer boundary |
| JWT01 mandatory sub/uid consistency | GATEKEEPER_ACCEPTED | PR266 original independent assertions PASS, exact-head ordinary CI 10/10 |
| JWT02 future iat | OWNER_DECISION_PENDING; unchanged RED | Recommend documented Option A; no external forgery demonstrated |
| JWT03 duplicate claim decoding | DOCUMENTED_ACCEPTED_BEHAVIOR / INFORMATIONAL | PR263 observation; parser unchanged in PR266 |
| Raw auth/parser variations and JWT mutations | Historical tested scope | PR263/PR265; mutation counts are not continuous fuzz-framework execution counts |
| Privacy cache and sanitized error/log boundaries | Bounded repository evidence only | Existing normal auth/HTTP tests; prior audit limitations retained; no external aggregator claim |
| Real crash/ACK and durable identity/financial recovery | Historical tested scope | PR260 38554843076/38554842556; PR262 38609351696; no internal WAL-critical-section guarantee |
| Lost-ACK two-row financial import | Tested store/service subset | PR265 job 116101937726; full public CSV review/append pipeline under ACK loss remains unverified |
| Runtime role separation and additive migration | PASS normal candidate scope | PR266 migration/race/Go CI; deployed DB/network/TLS policy unverified |
| Delivery attempts/leases and no plaintext proof persistence | Implemented and bounded test evidence | Ownership tests; SMTP transport cannot forcibly end remote work |
| Real provider email, edge, global topology, external logs | NOT_VERIFIED_PRODUCTION_ONLY | Deployment security verification required, never inferred from localhost/CI |
| Campaign-window/procedural acceptance | PENDING_OWNER | PR265 runs outside initial window; scoped instruction retrieved before execution |
| ADR010, recovery continuity and bearer lifetime boundary | PENDING_OWNER_ARCHITECTURE | No implicit new recovery or immediate bearer revocation feature |
| Formal Module1 closure | NOT_GRANTED | Requires owner decisions, final consolidated Gatekeeper and integration/post-merge gates |

No new product defect or missing core AUTH-ADV-02 production component is demonstrated
by this documentary/source review. That is not a new complete P0–P3 inventory or proof
that every historical evidence gap is resolved. Real production mail activation,
retention operations and complete pipeline/deployment residuals remain explicit.

## Consolidated integration candidate and exact cumulative changes

PR266 is one commit directly after PR251; PR251 is an ancestor. Against protected
develop it contains eight commits and 46 changed files. It can serve as the single
consolidated implementation candidate after approval. Do not independently merge
PR251 and then blindly replay its cumulative PR266 patch as a separate change.

[Implementation vs develop](https://github.com/AsifAbbasov/OpenInvest/compare/8d9998590fba6c605a64ddcc8429f51df82b56c3...e982d5a68e083a9e65b7f80b902f746b1b408bcb)
contains exactly:

- `.github/workflows/ci.yml`
- `backend-go/cmd/validate-openapi/auth_password_contract_test.go`
- `backend-go/cmd/validate-openapi/auth_timezone_contract_test.go`
- `backend-go/cmd/validate-openapi/main.go`
- `backend-go/cmd/verification-mail-worker/main.go`
- `backend-go/internal/auth/email_ownership.go`
- `backend-go/internal/auth/email_ownership_test.go`
- `backend-go/internal/auth/jwt_identity_contract_test.go`
- `backend-go/internal/auth/testdata/jwt01/pr261.go.txt`
- `backend-go/internal/auth/testdata/jwt01/pr265.go.txt`
- `backend-go/internal/auth/tokens.go`
- `backend-go/internal/httpapi/auth_handlers.go`
- `backend-go/internal/httpapi/auth_password_boundary_test.go`
- `backend-go/internal/httpapi/auth_test.go`
- `backend-go/internal/httpapi/cors_origin_test.go`
- `backend-go/internal/httpapi/email_ownership_test.go`
- `backend-go/internal/httpapi/legacy_auth_fixture_test.go`
- `backend-go/internal/httpapi/replay_app.go`
- `backend-go/internal/httpapi/routes.go`
- `backend-go/internal/httpapi/security_headers_test.go`
- `backend-go/internal/httpapi/transport_helpers.go`
- `backend-go/internal/postgres/email_delivery_recovery_test.go`
- `backend-go/internal/postgres/email_delivery_renewal_test.go`
- `backend-go/internal/postgres/email_ownership_integration_test.go`
- `backend-go/internal/postgres/email_ownership_store.go`
- `backend-go/internal/postgres/jwt_identity_http_integration_test.go`
- `backend-go/internal/postgres/runtime_privileges.go`
- `backend-go/internal/verificationmail/smtp.go`
- `docs/ADR/ADR-010-email-ownership-before-account-activation.md`
- `docs/audit/AUTH_ADV_02_EMAIL_OWNERSHIP_REMEDIATION_V1.md`
- `frontend-next/src/app/auth/verify/layout.tsx`
- `frontend-next/src/app/auth/verify/page.tsx`
- `frontend-next/src/common/api/openinvest.ts`
- `frontend-next/src/features/auth/components/AuthForm.tsx`
- `frontend-next/src/features/auth/components/AuthShell.tsx`
- `frontend-next/tests/email-ownership-api.test.mjs`
- `frontend-next/tests/email-verification-context.component.test.tsx`
- `frontend-next/tests/email-verification.component.test.tsx`
- `infrastructure/postgres/migrations/000015_auth_email_ownership.down.sql`
- `infrastructure/postgres/migrations/000015_auth_email_ownership.up.sql`
- `infrastructure/postgres/migrations/policy_manifest.json`
- `infrastructure/postgres/runtime/openinvest_runtime_role.sql`
- `openapi/components/responses.yaml`
- `openapi/components/schemas.yaml`
- `openapi/examples/auth.json`
- `openapi/openapi.yaml`

The new documentation branch starts exactly at PR266. Its direct comparison must
contain only the five documentation paths listed in the package section, not new
runtime/test/workflow modifications. Its develop-targeted PR includes inherited
implementation files; they are not new documentation-task changes.

A stacked PR to fix/module1-jwt01-identity-contract was preferred, but existing CI
filters pull_request bases to develop/main and the available connector has no workflow
dispatch operation. To obtain applicable exact-head CI without editing workflows or
requiring a new browser authorization, the documentation PR targets develop. Review
its direct comparison against e982d5a68e083a9e65b7f80b902f746b1b408bcb for scope.

## Gates and intended sequence — no merge authorization

1. Owner reviews the documentation candidate and explicitly decides ADR010, JWT02
   option, and PR265 procedural disposition. Confirm legacy continuity and current
   stateless bearer lifetime boundary, without authorizing recovery/revocation redesign.
2. Independent final Module1 Gatekeeper reviews the consolidated package/residuals;
   accepted JWT01 does not need its completed review/campaign repeated.
3. Before any authorized integration, re-fetch protected develop and candidate
   SHAs/trees, branch protection and ten required contexts. Changed baseline/head
   invalidates exact-snapshot integration assumptions; stop for re-evaluation.
4. Use one reviewed consolidated implementation lineage. If the documentation
   candidate is chosen as the eventual single integration candidate, it includes
   PR266 plus these five docs. Its exact final SHA/tree needs independent acceptance
   and normal CI 10/10; do not merge PR251/PR266/docs independently by assumption.
5. Merge requires separate owner authorization and normal protection, with exact
   resulting tree/parent verification. No merge or ready-for-review conversion is
   performed by this task. Post-merge normal CI must verify the exact protected head;
   no prior PR run substitutes for it. Formal closure is a separate reviewed action.
6. Deployment is separately gated: coordinated schema/runtime grants/API/frontend,
   restore/maintenance procedure, approved sender/relay/HTTPS verification URL,
   delivery monitoring and retention operations. Code existence is not delivery readiness.

Normal CI required contexts: Go tests; Python tests; Frontend build and typecheck;
OpenAPI contract; Docker Compose config; PostgreSQL migration validation; Go vet;
Go race tests; Go vulnerability scan; Dependency security scan. PR266 run
[38691771383](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38691771383) passed
10/10 at its exact head. Reachable govulncheck=0, pnpm/pip audit clean, race passed;
GO-2026-5932 remains a non-reachable module advisory, not suppressed. This old run is
implementation evidence, not CI for the new documentation head. New PR CI is recorded
in the delivery report after the final documentation commit; no recursive self-SHA
or invented future workflow result is embedded here.

## Rollback and deployment boundaries

Retain expanded identity/proof tables and evidence after any activation. Do not run
DOWN destructively or restore the old unverified public-registration endpoint. Old
runtime guards reject added grants, so blind binary rollback is unsafe. Prepare a
reviewed forward fix or approved maintenance admission closure; no dedicated feature
flag/zero-downtime rollout is assumed. Do not transfer financial ownership, silently
verify legacy accounts, reset passwords or invalidate sessions through this package.
No SMTP provider/domain/credentials, infrastructure, routing or production database
is configured. Production deployment controls remain post-deployment verification.

## Exact owner decisions requested

The following text is a proposal for the owner to adopt, not Builder approval:

> ADR010_APPROVED=YES — I approve the Proposed ADR010 revision in this documentation
> candidate as the email-ownership/new-account activation architecture and API change,
> including legacy login continuity, no email-only legacy takeover, current stateless
> access-token boundary and separately gated recovery/revocation changes. SMTP and
> operational deployment approval remain separate. Status may be recorded as Accepted
> only after this explicit approval is retained in the project record.
>
> JWT02_OPTION_A_APPROVED=YES — I select the exact Option A residual described in the
> JWT02/JWT03 record, retain mandatory exp validation and preserve historical RED
> expectations without claiming they passed. I do not authorize Option B code changes.
>
> PR265_PROCEDURAL_ACCEPTANCE=YES — I confirm the 20:32:21Z separate scoped instruction
> and accept final 21:00:23Z results as supplemental post-deadline evidence, not an
> extension/reset of the original campaign. Technical limitations and REDs remain.
>
> MERGE_AUTHORIZED=NO; DEPLOY_AUTHORIZED=NO; MODULE1_CLOSED=NO.

Owner may instead select the separately specified JWT02 Option B, which requires a
new narrowly authorized implementation and Gatekeeper gate; it is not silently bundled.
The next action is owner architecture/procedural decisions on this package, followed
by final consolidated readiness review. Production-only residuals are not relabeled PASS.

## State at package preparation

ADR010_FORMAL_STATUS=PROPOSED; JWT01_REMEDIATION=GATEKEEPER_ACCEPTED;
JWT02_OWNER_APPROVAL=PENDING; PR265_PROCEDURAL_ACCEPTANCE=PENDING;
READY_FOR_OWNER_ARCHITECTURE_DECISION=YES;
READY_FOR_FINAL_MODULE1_GATEKEEPER=NO_PENDING_OWNER_DECISIONS;
MODULE1_CLOSED=NO; MERGE_PERFORMED=NO; DEPLOY_PERFORMED=NO.

## Current closeout review gate

The three owner decisions are approved in the linked nine-line record. Owner-decision blockers above are resolved; the preparation snapshot is retained as history. Next technical review is Main Chat Review; automatic Gatekeeper delegation is not authorized. Consolidated integration review, exact candidate CI, separate merge authorization and post-merge verification remain required. SMTP/recovery/immediate JWT revocation/deployment stay separately gated; Module #1 is not formally closed.
