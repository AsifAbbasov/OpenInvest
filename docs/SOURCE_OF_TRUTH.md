# OpenInvest Source of Truth

| Field | Value |
| --- | --- |
| Document ID | SOT-001 |
| Version | 1.4.76 |
| Status | Approved / Architecture Freeze Active |
| Owner | Principal Architect |
| Supersedes | Disconnected source-of-truth declarations in legacy documents |
| Dependencies | Documents 42–43 and accepted ADRs |
| Last Review Date | 2026-08-30 |
| Next Review Date | Before Stage 3.25 privacy evidence-collection plan review or the next separately reviewed audit-remediation scope |

## Architecture status

**Architecture Freeze v1.2: ACTIVE**
**Documentation Freeze: ACTIVE**
**Last completed implementation stage: Stage 3.41 — P3-09 Next.js security maintenance implementation**
**Last completed privacy planning gate: Stage 3.24 — Privacy Security Review Readiness Dossier**
**Last completed audit-remediation planning gate: Stage 3.40 — P3-09 Next.js security maintenance planning**
**Last completed architecture amendment: Next.js Web Presentation Amendment**
**Stage 3.39 implementation merge baseline: PR #97 squash merge `abbd9f9f61574621e206f2e196b1fb8f056dc194`; the moving protected `develop` HEAD is intentionally not hard-coded by this merge-activated closure record**
**Stage 3.41 implementation merge baseline: PR #101 squash merge `a2cfeaa5ca68fdd951e2a99f69c96aec362fc416`; the moving protected `develop` HEAD is intentionally not hard-coded by the Stage 3.42 merge-activated closure record**
**Current privacy-planning work item: Stage 3.25 privacy Security Review evidence-collection plan; it remains documentation-only and does not authorize privacy-lifecycle implementation**
**Stage 3.27 remediation: CLOSED for P1-02, P1-03, and P1-04; implementation PR #55 was squash-merged into `develop` at `6e8c806de857f844954f1db513487357dfe90187` after exact-head CI #90, renewed independent `APPROVED` review on `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c`, explicit human squash-merge authorization, and closure governance through PR #58**
**Stage 3.28 remediation: CLOSED for P1-01 and P1-05; implementation PR #59 was squash-merged into `develop` at `dc83f5f3a11da164e6809593861d96ccf47b29ca` after exact-head CI #114, renewed independent `APPROVED` review on `92edab5d3e93dafe2fcc6247644e38e878a4202f`, explicit human squash-merge authorization, and closure governance squash-merged through PR #60 at `0ddc618a3450ea81fd4befb3b10c959b3cb82a25`; Stage 3.25 and P2/P3 remain separate**
**Stage 3.29 remediation: CLOSED for P2-05/P2-06/P2-07/P2-08/P2-15; implementation PR #61 was squash-merged into `develop` at `7331d3f34783baec3997497d1a79b78eaa558bd4` after exact-head CI #124, first independent `REQUEST CHANGES`, blocker remediation on `f9e70e70956c76edbc2ab02c52d45124b2dea525`, renewed independent `APPROVED`, explicit human squash-merge authorization, and closure governance squash-merged through PR #62 at `0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc`; Stage 3.25 and remaining P2/P3 stay separate**
**Stage 3.30 remediation: CLOSED for P2-02/P2-03/P2-04; implementation PR #63 was squash-merged into `develop` at `8f68dd18800918e6a9882e995e13dba2723dc929` after exact-head CI #128, independent final `APPROVED`, explicit human merge authorization, and closure governance squash-merged through PR #64 at `ae6497050692798795efb85678af64db97cc5f53`; Stage 3.25 and remaining P2/P3 stay separate**
**Stage 3.31 remediation: CLOSED for P2-01/P2-14; implementation PR #65 was squash-merged into `develop` at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897` after exact-head CI #133, independent final `APPROVED` review on `82557c55c0772a66707088b858ec9eafc2073119`, explicit human squash-merge authorization, and closure governance squash-merged through PR #66 at `ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746`; Stage 3.25 remains separate**
**Stage 3.32 remediation: CLOSED for P2-09/P2-13; implementation PR #67 was squash-merged into `develop` at `0623d5ef326cd783b7dc0417dbcb02f18c506171` after exact-head CI #181, first independent `REQUEST CHANGES` for the P2-13 cross-principal browser retry-slot collision, remediation, repeat independent `APPROVED` review on `02aa2417a3caca79e2afc4e7b598b92055de96b7`, explicit human squash-merge authorization, and closure governance squash-merged through PR #68 at `a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be`; 5 P2 plus 10 P3 findings remained after closure; Stage 3.25 remains separate**
**Stage 3.33 remediation: CLOSED for P2-10/P2-11/P2-12; implementation PR #69 was squash-merged into `develop` at `87a7c38e16062a5f3fcef3727f60c0c6741eb805` after exact-head CI #199, two independent `REQUEST CHANGES` cycles for P2-12 credential-graph escalation gaps, both remediations, final independent `APPROVED` review on `88ec8f739f7bcc96267c25f41560e1960d4d48d5`, explicit human squash-merge authorization, and closure governance squash-merged through PR #70 at `71a1faeb97d33d05f2936111b53f1285edddabe9`; exactly 2 P2 plus 10 P3 findings remained after closure**
**Stage 3.34 remediation: P2-17 implementation is canonical through PR #80 at `c686a6721df51063ccf62a0303bb759d2215d60e` after exact-head CI #230, independent implementation `APPROVED`, and explicit human squash-merge authorization; P2-16 is mechanically enforced through public-repository `develop` protection and squash-only merge policy; Stage 3.34 remediation closure is canonical through PR #82 at `ae5a152114cc163867a363953f8a3202396b1f6c`. Stage 3.35 P3-01 runtime implementation is canonical through PR #84 at `a47df19ccc7edff73f39f4e76aec47580c168c46`; governance closure is completed through PR #85.**
**Stage 3.36 remediation: planning is canonical through PR #87 at `251296e0831cbb0b81c7799cc82cbdf3b451ae6e`; runtime implementation is canonical through PR #88 at `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa`, from frozen runtime head `131f1bf963e9d232b9e23273edd54caf54c10ffb`, after exact-head CI #257 / run `32822925542` with all 10 required jobs successful; closure governance is canonical through PR #89 at `9c83b68e28bbb8bc971620d3e00be5e177ce0820`. P3-03 is CLOSED; the original audit backlog immediately after Stage 3.36 closure was P0=0 / P1=0 / P2=0 / P3=8: P3-02, P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 remains separate.**
**Stage 3.37 remediation: planning is canonical through PR #90 at `46f74528dcc19424ad087d30d4f2f778e2079b87`; runtime implementation is canonical through PR #91 at `cb6d9b28cd47b1cd283b5861b916e0be627d0ac2`, from frozen final runtime head `1a2f89a0fa5095b3cca790521afa484bdc61e8a6`, after exact-head CI #265 / run `32869754524` with all 10 required jobs successful and fresh published-head independent `APPROVED`; closure governance is canonical through PR #92 at `305a53bb07136b274717ff48778a5e93d7b1607c` after exact-head closure CI #266 / run `32880893193`, fresh independent closure `APPROVED`, and separate explicit human squash-merge authorization. P3-02 is CLOSED. The current original audit backlog is P0=0 / P1=0 / P2=0 / P3=7: P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 remains separate.**
**Stage 3.38 remediation: planning is canonical through PR #93 at `a944f1e5d5ee7d84db5393e8760eda254d732edd`; runtime implementation is canonical through PR #94 at `2df9946d77ee044a191a0422c8cccbbfe02dc7c9`, from exact published runtime head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`, after exact-head CI #268 / run `32913862780` with all 10 required jobs successful and fresh published-head independent `APPROVED`. Closure PR #95 was ultimately published at final head `25eb3b9c3c153672f22a6718a7815a5d3c527f44`, passed exact-head CI #271 / run `32961508562`, and was actually squash-merged into `develop` at `c5962fa09b6d7d145dda203dbdf90069de7b1fcc`. P3-05 is CLOSED. The current original audit backlog is P0=0 / P1=0 / P2=0 / P3=6: P3-04, P3-06, P3-07, P3-08, P3-09, and P3-10. Historical Stage 3.38 failed review/remediation evidence remains preserved in its stage documents, while their active top-level current-state metadata is synchronized by the companion Stage 3.38 proposed patches. Stage 3.25 remains separate.**
**Stage 3.39 remediation: planning is canonical through PR #96 at `32b198ee9d349f119ed374fd86d47622e27bcd73`; implementation plus governed forensic/evidence history is canonical through PR #97 squash-merged into protected `develop` at `abbd9f9f61574621e206f2e196b1fb8f056dc194` from final published head `26f5ca18ca5772db569d22ce2eff64d5a7850b1b`, after exact-head CI #279 / run `33121609429` with all 10 required jobs successful and separate final Internal + External published-head `APPROVED` verdicts. Closure PR #99 reached final head `4c2439f3fdc213fd38d2669233d993cc3dac043b` and was actually squash-merged into protected `develop` at `41e35b672d166cf74c3f0c3ee248330193ae51c1`. P3-04 is CLOSED. The current pre-Stage-3.42 original audit backlog is P0=0 / P1=0 / P2=0 / P3=5: P3-06, P3-07, P3-08, P3-09, P3-10. Stage 3.25 remains separate.**
**Stage 3.40/3.41 P3-09 remediation: planning is canonical through PR #100 squash-merged at `559b57d0951cdc67125c2f72fc1fcfb34399e90e` with approved plan blob `33cb076fa529d3efcdfbea9a95d111aec30ccbad`. Implementation PR #101 reached final evidence head `d88be3c90231f374d7e6b7d94f4cd89e6788f700`, passed CI #291 / run `33277717164` with all 10 required jobs successful, preserved package blob `d6d605620e1bff426998d8bda716b7c2eda0613d` and lockfile blob `b3d656e792bdd28b16dea553b378f15f553b3074`, resolved `EXT-STAGE-03-41-P3-01`, received final External `APPROVED` and final evidence-publication verification `APPROVED`, and was actually squash-merged into protected `develop` at `a2cfeaa5ca68fdd951e2a99f69c96aec362fc416`. This SHA is the immutable Stage 3.41 implementation merge baseline; the moving protected `develop` HEAD is intentionally not hard-coded as permanently equal to it.**
**Stage 3.42 P3-09 closure activation: before the approved Stage 3.42 closure record and synchronized canonical surfaces are present on protected `develop`, P3-09 remains OPEN and the original audit backlog remains P3=5: P3-06, P3-07, P3-08, P3-09, P3-10. Once those surfaces are present on protected `develop`, P3-09 is CLOSED and the original audit becomes 28/32 closed (87.5%), with P0=0 / P1=0 / P2=0 / P3=4: P3-06, P3-07, P3-08, P3-10.**
**Stage 2 status: Closed / merged into `develop`; ADR-006 accepted**
**Web presentation amendment status: Closed / merged into `develop`; ADR-007 accepted**

## Document priority

`Document 43 → Document 42 → accepted ADRs → Documents 1–41 → README → comments`

No lower-priority document, code, or comment may override a higher-priority decision. An architecture change requires an ADR, review, approval, and update to this file.

## Product definition

OpenInvest is a **Personal Capital Operating System**. It is not a broker, bank, trading terminal, investment adviser, or tax service.

The first public MVP targets investors with real portfolio-accounting pain: long-term, dividend,
FIRE, and multi-account investors who need independent, explainable real-return analytics. It is not
optimized for casual brokerage-app users who only need a simple green/red return badge.

## Frozen stack and architecture

- Monorepo: `OpenInvest/`
- Backend: Go 1.24+ and Fiber
- Analytics: Python managed with uv
- Database: PostgreSQL; schemas `identity`, `investment`, `analytics`, `tax`, `audit`
- Cache: Redis plus process-local RAM cache
- Web frontend: Next.js App Router, TypeScript, and pnpm; presentation layer only under ADR-007
- Current client implementation scope: Web MVP only
- Mobile future: iOS SwiftUI and Android Jetpack Compose; no current mobile implementation
- Style: API First, DDD, Clean Architecture, Event Driven
- Data: canonical database, immutable transactions, rebuildable versioned snapshots
- Events: at-least-once delivery and idempotent business processing through outbox/inbox
- Security: Zero Trust and Privacy by Design
- External sources: official/permitted sources registered before use and accessed only through backend collectors
- Client: no external market-data calls and no LocalStorage for business data
- Boundary: Browser/Next.js → OpenAPI-defined Go API → PostgreSQL/Redis/future Python workers;
  Next.js never replaces the Go business API or accesses data stores directly
- Delivery: no automatic commit or push without user review and approval
- Review: mandatory feature branch → local checks → read-only Internal Review phase in one designated
  review chat → human commit/push permission → Draft PR → exact-head green CI → fresh External phase
  in the same designated chat without using the Internal verdict/findings as supporting evidence →
  evidence-only publication of the previously withheld current Internal chronology after the External
  verdict → CI and same-chat exact verification → explicit human merge authorization → squash merge;
  see `REVIEW_WORKFLOW.md`.

## Mandatory quality gates

- Builder Agent cannot approve its own work.
- Every changed line is reviewed internally before commit permission is requested.
- Internal Review Agent produces findings only and cannot edit, stage, commit, or push.
- Draft PR cannot merge without exact-head green CI, completed designated-chat review phases,
  publication and verification of required Internal evidence, and explicit human approval.
- Development-path External review is a fresh sequential phase in the same designated review chat.
  It must not use the current Internal verdict/findings as supporting evidence.
- Current Internal review evidence is withheld from repository/PR evidence until the External verdict;
  after that verdict it is published in an evidence-only follow-up, followed by required CI and
  same-chat exact verification before human merge authorization.
- Every fifth completed stage requires a full repository line-by-line audit before proceeding.
- A review gate may add evidence and reject scope; it cannot silently change frozen architecture.

## MVP scope

Included: registration; portfolio; transactions; stock card; bond card; dividend calculator; dividend calendar; snapshots; weighted-average cost; XIRR; real return; inflation-adjusted return; purchasing-power card; dashboard.

Excluded: AI Assistant; scenarios; premium analytics; Tax XML export; email automation; forecasts; family accounts; public API; foreign securities; mobile applications.

Public-MVP readiness additionally requires an approved import/reconciliation path so users are not
forced to enter large transaction histories manually. The preferred first import path is user-supplied
broker files with explicit review, not credential scraping or direct broker API synchronization.

Purchasing Power remains an MVP differentiator, but it is a secondary explanatory insight. Real
return, capital, dividends/coupons, and inflation-adjusted performance stay above consumer-good
equivalents in dashboard priority.

Tax export remains outside MVP. Any future tax calculation core must be deterministic and covered by
financial/legal test vectors; AI may explain or assist review, but must never be the source of tax
truth.

Product risk refinement is closed and merged into `develop` at
`65bdf6537b44ed57e1c00bf68d2dacd70aa09702`. Stage 3.3 is closed and merged into `develop` at
`11805cc298bba13f09f7f7af8b1e1178dc351209`, with closure documentation merged at
`fe402030359459f909c156a1e993f18ceed257bf`. Stage 3.4 is closed and merged into `develop` at
`86582efaa420b2c38465a5d0da041814149392c7`; it added end-to-end local verification and root
developer commands only. Stage 3.5 is closed and merged into `develop` at
`072d38d94b529221d6467502f82f03a674a7d805`; it approved the design guardrails for user-supplied
broker-file import. Stage 3.6 implementation is closed and merged into `develop` at
`e2b05650a4422b97d4bd924254367106b6a4686b`, with closure governance merged at
`fb651632036fabaa31ec92e9d28b5782ca0f92e5`; it added an internal CSV parse/normalize/review/
append-plan slice only. It does not authorize public import endpoints, broker API integration,
upload UI, SQL migrations, workers, or automatic ledger append. Stage 3.7 import append planning is
merged into `develop` at `36d86c7ff2a9c75478de155d4f60b979b8da9376`. Stage 3.7 implementation is
closed and merged into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`; it added internal
atomic append of user-approved import rows with duplicate revalidation, idempotency protection,
minimal audit evidence, and deterministic snapshot rebuilds. Public import endpoints, upload UI,
import-session persistence, broker/provider integrations, workers, tax, mobile, and AI remain out of
scope.

Stage 3.8 import review append flow planning is merged into `develop` at
`a35af2f5207bd564647d2a3fc032f4f940e62ddd`. Stage 3.8 implementation is closed and merged into
`develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`; it added internal orchestration between
Stage 3.6 review output and Stage 3.7 atomic append. Public import endpoints, OpenAPI changes,
upload UI, SQL import-session persistence, raw file persistence, workers, broker/provider
integrations, tax, mobile, AI, and automatic append without explicit approved decisions remain out
of scope.

Stage 3.8 closure governance is merged into `develop` at
`cb9a392eb90ede954d9cc68b247bada13a1540d9`. Stage 3.9 import API boundary planning is merged into
`develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`. Stage 3.9 implementation is closed and
merged into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`, with closure governance
merged at `682ffd856395a6e3e988817551a512898fda2d38`; it added only the public Go API boundary,
OpenAPI contract, DTOs, tests, and documentation for user-supplied CSV import review/append.
Stage 3.10 import upload/review UI planning is closed and merged into `develop` at
`27480d6ff22e2929e33aeac352aef8a1b01bb448`. Stage 3.10 implementation is closed and merged into
`develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71`; it added only the presentation-only
Next.js import upload/review UI over the existing Go API boundary. SQL import-session persistence,
raw file persistence, backend contract changes, workers, broker/provider integrations, tax, mobile,
and AI remain out of scope. Stage 3.11 authentication and privacy-boundary planning is closed and
merged into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125`. Stage 3.11 implementation is
closed and merged into `develop` at `5c49173ac858995929f266c2de991282dd194dec`; it implemented only
the approved Go API auth, privacy-default, session, CSRF, audit, and persistence boundary. It does
not authorize frontend auth UI, email verification, OAuth/passkeys/2FA, tax, mobile, AI, workers,
provider integrations, or portfolio feature expansion. Unsafe local auth flags with a configured
`DATABASE_URL` require `OPENINVEST_ENV=development` or `local`; production authorization must use
signed access tokens. Stage 3.11 closure governance is merged into `develop` at
`2febb6f49224ec6252368d2195a4e3054ea24278`. Stage 3.12 Web authentication UI planning is closed
and merged into `develop` at `25be13ce84844562e0381b79f4b81cbfed7eb44d`. Stage 3.12 Web
authentication UI implementation is closed and merged into `develop` at
`b4840b60346109e3cd54a07d9e1e131fc0cfad23`. Stage 3.12 closure governance is merged into
`321eaf4f75df83d85fd356a8d6a454e49bbc4db4`. Stage 3.13 instrument catalog planning is
merged into `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92`. Stage 3.13 instrument
catalog implementation is closed and merged into `develop` at
`b9c05fb14d0ee03e6de4dfc04ff67c16da33040b`; it is limited to backend-owned approved
asset-catalog fixture resolution, supported ticker validation, existing asset table usage,
snapshot stock/bond bucket classification from backend-owned asset type, tests, and documentation.
It does not authorize OpenAPI changes, SQL migrations, Go handler changes, frontend work, provider
integrations, workers, market-data ingestion, stock/bond cards, dividend/coupon scope, tax, mobile,
or AI. Stage 3.13 closure governance is merged into `develop` at
`45a298e3ba36dbe711fa27b8d044d80a77cfd74a`. Stage 3.14 asset search/card API
boundary planning is closed and merged into `develop` at
`2c4f7853599a455bb0cc04114b338a1145baf39c`. Stage 3.14 implementation is closed and
merged into `develop` at `57a9404952cb65693614109dd4a14d41fa5c4295`; it is limited to the Go API
asset search/detail boundary over the approved local catalog. Runtime asset search uses
`lastPrice: null` until an approved market-data source exists, and asset detail remains deferred
unless every mandatory source/detail field can be populated without fabricated data. Stage 3.14
closure governance is merged into `develop` at `f5289eb604b8ba31aa422d0d09950da02e0f48b3`.
Stage 3.15 Web asset discovery UI planning is closed and merged into `develop` at
`dfeab109b2825fe0e0317e87a7abf2e706a29ea6`. Stage 3.15 implementation is closed and merged into
`develop` at `22bede651a646d0e8b06568bda457d0626891e63`; it added only the reviewed Next.js
presentation-only asset discovery boundary over the existing Go API and does not authorize market
data, stock/bond card calculations, provider integrations, workers, tax, mobile, or AI. Stage 3.15
closure governance is merged into `develop` at `9eec98c36d7aeffb21dc2d7e7e0eb1681106901d`. Stage
3.16 repository audit planning is closed and merged into `develop` at
`74eebe9ec8231764f21ce384c4690d073d0273da`; the mandatory audit returned `REQUEST CHANGES`.
Its in-scope blocking findings were resolved by the Stage 3.16 audit fixes, which passed read-only
review and GitHub CI before PR #44 was squash-merged into `develop` at
`9e6b8a753bf73ef020ce40461df25a5878344d92`. The original audit report retains its historical
`REQUEST CHANGES` verdict. Neither the closure nor prior fixes authorize a subsequent
implementation stage. Stage 3.17 privacy-lifecycle planning was squash-merged through PR #46 at
`1e8c240`; it remains documentation-only evidence and does not authorize implementation. Stage 3.18
was squash-merged through PR #47 at `4680e9c1b7b916169972c84ad8c3879955c7f509`; it preserved the
contract/security boundary without authorizing implementation. Stage 3.19 was squash-merged through
PR #48 at `fdf74c16446e7623f76882aa7add64554141abc6`; it remains a provider-neutral security/ADR
proposal for cryptographic erasure, deletion-marker replay, restore, and separation of duties. Stage
3.20 was squash-merged through PR #49 at `849d934906f878a6d79ba89e940e5ba470e64c09`; it recorded
the residual threat boundary without claiming a Security Review or implementation. Stage 3.21 was
squash-merged through PR #50 at `207325e0497cc2608b99366f7f840472d270b6ed`; it inventories
field-level disposition and external evidence gaps. Stage 3.22 was squash-merged through PR #51 at
`5f42d32db1e045c23fb99a5af8f136b7a49e3bc2`; it defines provider-neutral key custody and
destruction proof. Stage 3.23 was squash-merged through PR #52 at
`f7f23bce33038f259c976db6375079c68209a7aa`; it defines the non-identifying deletion-marker
control plane and restore gate. Stage 3.24 was squash-merged through PR #53 at
`544ad8cc7371caf93913ea7716f3feb68be0ea44`; it defines Security Review readiness. Stage 3.25
is the separate evidence-collection plan. Proposed ADR-008 is non-normative until formal Security
Review and explicit human acceptance; no implementation, OpenAPI change, migration, provider
selection, or operational configuration is authorized.

## Stage 3.27 audit remediation status

Stage 3.27 is a separately authorized, narrowly scoped remediation of repository-audit findings
P1-02, P1-03, and P1-04. It does not expand product scope or authorize the privacy-lifecycle,
market-data, tax, mobile, AI, or provider work that remains governed separately.

The remediation candidate is based on `develop` at
`213d1d9b4369a5e046b26c3a08990aa571603eaa`. Local Go 1.25.14/PostgreSQL 18 verification,
migration apply/rollback/reapply, OpenAPI validation, full Go tests, `go vet`, direct database
constraint regression coverage, and independent pre-commit review have passed. The implementation
candidate was committed and pushed as `19a8abbb0c07ded7441839bfa99b538739e21fbc`, Draft PR #55
was opened against `develop`, and GitHub Actions CI run #83 passed on that implementation head.

Stage 3.27 is not canonical and must not be described as closed until the final PR head has green CI,
the required PR review is approved, explicit human merge approval is given, and the PR is
squash-merged into `develop`.

### Stage 3.27 final-review correction

Final independent review of PR #55 at pre-correction head `c6c3a4c91a108426448a2bc230873ab9e479a335` identified a blocking P1-02 order-dependence: a fallback fingerprint row could be persisted first and then coexist with a later strong broker-keyed row carrying the same scoped financial fingerprint. The corrective candidate rejects mixed fallback/strong identity in importer review, batch validation, PostgreSQL store lookup, and a concurrency-serialized database guard, while still allowing distinct non-null broker identities with identical economics. Temporary CI run #85 passed on the corrective code candidate before final governance synchronization; it remains historical, head-specific evidence. Current CI state is intentionally not pinned to a run number in tracked governance: the authoritative gate is the required PR checks attached to the exact merge-candidate head.


### Stage 3.27 closure governance

Implementation PR #55 was squash-merged into `develop` at `6e8c806de857f844954f1db513487357dfe90187`.
Its exact merge-candidate head `b281d5bdc1c28ca4f4ac6d913ca9683859209e4c` passed GitHub Actions CI #90
and renewed independent external review returned `APPROVED` after correction
of the earlier order-dependent fallback-to-strong identity blocker.

Explicit human authorization approved the squash merge.

Closure governance is recorded through PR #58. Stage 3.27 is closed for P1-02,
P1-03, and P1-04. The then-remaining P1-01/P1-05 findings were subsequently
addressed by the separately governed Stage 3.28 remediation. Stage 3.25 privacy
evidence planning remains separate.

## Stage 3.28 audit remediation closure governance

Stage 3.28 is the separately authorized remediation of repository-audit findings P1-01 and P1-05.
It does not reopen Stage 3.27, authorize Stage 3.25 privacy implementation, or close any P2/P3 item.

Implementation PR #59 was squash-merged into `develop` at `dc83f5f3a11da164e6809593861d96ccf47b29ca`. Its exact
merge-candidate head `92edab5d3e93dafe2fcc6247644e38e878a4202f` passed GitHub Actions CI #114. The first independent review
returned `REQUEST CHANGES` for one governance-status mismatch only and reported no additional
blocking P1-01/P1-05 security or correctness issue. After the one-line governance correction,
renewed independent review returned `APPROVED` on the exact final head. Explicit human
squash-merge authorization was then given.

P1-01 is remediated through persisted session-family identity for new sessions, conservative
legacy-session containment, serialized refresh/logout mutation paths, replay-triggered family
revocation, and PostgreSQL defense in depth.

P1-05 is remediated without weakening Argon2id parameters: the approved 64 MiB / t=3 / p=1
cost remains, while hash/verify/dummy work shares a process-wide fail-fast two-operation capacity
gate and over-budget stored encodings are rejected before expensive work.

The independent reviewer treated the current generic HTTP 500 mapping for `ErrAuthCapacity` as
non-blocking for the P1-05 resource-exhaustion finding. Dedicated `503 + Retry-After` semantics
remain optional HTTP-contract hardening.

Closure governance PR #60 was squash-merged into `develop` at
`0ddc618a3450ea81fd4befb3b10c959b3cb82a25` after exact-head CI #122, independent closure
`APPROVED` review, and fresh explicit human squash-merge authorization.

Stage 3.28 is CLOSED for P1-01 and P1-05. Stage 3.25 privacy evidence planning and the then-remaining
P2/P3 findings stayed separate.

## Financial standard

- Decimal only; binary float forbidden for financial values.
- Half-even rounding.
- Eight decimal places internally; two for monetary display.
- SQL `DATE` for financial business dates.
- UTC `TIMESTAMP WITH TIME ZONE` for system timestamps.
- MOEX calendar for MVP market events.
- Canonical financial test vectors are mandatory before production algorithms.

## Privacy definitions

- **Personal Data:** information that identifies a person directly or indirectly.
- **Pseudonymized Data:** data linked through a reversible identifier.
- **Anonymous Data:** data that cannot be linked back to an individual by any reasonable technical or organizational means.

Deleting a user removes identity data and irreversibly destroys its link to the financial ledger. The detached ledger becomes **Anonymous Financial History**. OpenInvest retains no re-identification mechanism.

## Retention

- Identity/personal data: deleted completely after the approved deletion lifecycle.
- Audit: 10 years.
- Anonymous transactions and snapshots: no fixed expiration.
- Backups: maximum 90 days, then automatic destruction.

## Feature matrix

| Capability | MVP | State |
| --- | --- | --- |
| Registration and privacy defaults | Yes | Stage 3.12 closed; Go API auth plus presentation-only Web auth UI |
| Portfolio and transactions | Yes | Stage 3.4 verification closed |
| MOEX shares and bonds | Yes | Stage 3.15 Web asset discovery UI slice closed over the Stage 3.14 asset API boundary |
| Dashboard and snapshots | Yes | Stage 3.4 verification closed |
| WAC, XIRR, real/inflation returns | Yes | Planned |
| Dividend calculator/calendar | Yes | Planned |
| Broker file import and reconciliation | Public-MVP readiness candidate | Stage 3.10 upload/review UI slice closed; no import-session persistence |
| Purchasing power | Yes | Planned as secondary insight |
| Tax export | No | Experimental; feature flag off |
| Foreign securities | No | Backlog v2.0 |
| AI, mobile, premium, public API | No | Backlog v2.0 |

## ADR registry

| ADR | Decision | Status |
| --- | --- | --- |
| ADR-001 | Go and Fiber backend | Accepted |
| ADR-002 | PostgreSQL canonical database | Accepted |
| ADR-003 | OpenAPI-first contracts | Accepted |
| ADR-004 | Versioned rebuildable snapshots | Accepted |
| ADR-005 | Privacy by Design | Accepted; interpreted with Document 43 anonymization terminology |
| ADR-006 | Stage 2 MVP contract and canonical model freeze | Accepted |
| ADR-007 | Next.js App Router for Web presentation only | Accepted |

## Version matrix

The complete version and legacy-document matrix is maintained in `VERSION_MATRIX.md`. `DOCUMENT_INDEX.md` is the navigation registry.

## Open questions

No unresolved architecture questions exist at Freeze v1.2 activation. See `OPEN_QUESTIONS.md` for the controlled process.


## Stage 3.29 audit remediation closure governance

Stage 3.29 is a narrow repository-audit remediation for P2-05, P2-06, P2-07, P2-08, and P2-15 only.

Implementation PR #61 was squash-merged into `develop` at
`7331d3f34783baec3997497d1a79b78eaa558bd4`. The reviewed implementation head was
`f9e70e70956c76edbc2ab02c52d45124b2dea525`; exact-head GitHub Actions CI #124 completed
`SUCCESS`. The first independent final review returned `REQUEST CHANGES` for one P2-07 gap:
otherwise-valid financial inputs could still overflow `analytics.portfolio_snapshots`
`NUMERIC(28,8)` aggregate/derived values. The remediation kept snapshot methodology SQL-owned,
added same-transaction pre-persistence range admission for all persistence-bound snapshot metrics,
and added PostgreSQL integration proof for cumulative deposit overflow, BUY component-sum overflow,
and rollback atomicity. Renewed independent final review on the exact head returned `APPROVED`.
Explicit human authorization was received before the squash merge.

P2-05 maps malformed decimal command input to deterministic validation semantics. P2-06 enforces
the 500-character stored-note contract at application/import boundaries. P2-07 aligns canonical
Decimal/OpenAPI magnitude with `NUMERIC(28,8)` and guards persistence-bound derived and aggregate
snapshot values. P2-08 makes portfolio/transaction JSON writes fail closed on unknown fields.
P2-15 rejects duplicate normalized CSV headers before row normalization.

Closure governance PR #62 passed exact-head CI #127 on
`d70cf8322bae8713e6a6808624fa1493a46ed0ad`, received independent closure `APPROVED` review and
fresh explicit human squash-merge authorization, and was squash-merged into `develop` at
`0bfb3ea9f8e4cc7337a92caef5c7a73f9a8921bc`.

Stage 3.29 is CLOSED for P2-05/P2-06/P2-07/P2-08/P2-15. At Stage 3.29 closure, the original audit
backlog contained 12 P2 and 10 P3 findings. Later remediation does not alter that historical closure
scope. Stage 3.25 privacy Security Review evidence planning remains separate and is not superseded.


## Stage 3.30 audit remediation closure governance

Stage 3.30 is a narrow repository-audit remediation for P2-02, P2-03, and P2-04 only.

Implementation PR #63 was squash-merged into `develop` at `8f68dd18800918e6a9882e995e13dba2723dc929`. The independently reviewed
exact implementation head was `2f788e0811d78c9def0502676a74bee2f9922bf5`; exact-head GitHub Actions CI #128 completed `SUCCESS`.

P2-02 now binds a short-lived signed review token to explicit token/parser versions, subject and
portfolio context, source label/file hash, every row identity, deterministic normalized parser
semantics, final review semantics, and the set of rows that were APPENDABLE at review time. Append
reparses the exact payload and fails closed on expiry or normalized semantic drift. Mutable
post-review ledger races still reach the locked store, preserving the store as the final duplicate
and identity authority. This does not close P2-09 or P2-13.

P2-03 moves the 100-row computational admission limit into `ReviewCSV`: the parser fails when the
101st data record is read, so HTTP and non-HTTP callers share the same bound.

P2-04 removes the latest-100 reconciliation cutoff. Review derives bounded relevant trade dates and
privacy-minimized import identity keys, then PostgreSQL queries all matching historical rows without
an arbitrary recency limit and without loading the entire portfolio ledger. CI #128 ran the Go suite
with PostgreSQL configured, including the regression proving an old relevant row outside the public
latest-100 page is still returned.

Closure governance PR #64 passed exact-head CI #132 on
`7d97f5f967074f98311adcd4b8f7962e0584c719`, received independent closure `APPROVED` review and
fresh explicit human squash-merge authorization, and was squash-merged into `develop` at
`ae6497050692798795efb85678af64db97cc5f53`.

Stage 3.30 is CLOSED for P2-02/P2-03/P2-04. At Stage 3.30 closure, the original audit backlog
contained 9 P2 and 10 P3 findings. Later remediation does not alter that historical closure scope.
Stage 3.25 privacy Security Review evidence planning remains separate and is not superseded.


## Stage 3.31 audit remediation closure governance

Stage 3.31 is a narrow repository-audit remediation for P2-01 and P2-14 only.

Implementation PR #65 was squash-merged into `develop` at `9bf4d1d31597918eacf0c3358bf6caa2aa9db897`. The independently reviewed
exact implementation head was `82557c55c0772a66707088b858ec9eafc2073119`; exact-head GitHub Actions CI #133 completed `SUCCESS`.

P2-01 is remediated by routing logout through authentication admission before request decoding,
auth-service work, and rejected-auth persistence. Rate-limited requests therefore cannot create
another rejected logout audit event. Logout now exposes the existing HTTP 429 `RateLimited` contract.

P2-14 is remediated with finite per-key attempt count, finite global downstream-attempt budget per
window, finite active key-bucket cardinality, and cross-key expired-bucket reclamation. Capacity
exhaustion fails closed. The limiter remains process-local and no Redis/distributed limiter is claimed.

Closure governance PR #66 was squash-merged into `develop` at
`ebc8222d2fdd03b6e3cbdb185bd3db6d0a6b4746`. P2-01/P2-14 are CLOSED. The original audit backlog
then contained 7 P2 and 10 P3 findings. Stage 3.25 privacy Security Review evidence planning remains
separate and is not superseded.


## Stage 3.32 audit remediation closure governance

Stage 3.32 is a narrow repository-audit remediation for P2-09 and P2-13 only.

Implementation PR #67 was squash-merged into `develop` at
`0623d5ef326cd783b7dc0417dbcb02f18c506171`. The exact independently reviewed implementation head
was `02aa2417a3caca79e2afc4e7b598b92055de96b7`; exact-head GitHub Actions CI #181 completed
`SUCCESS` across all six jobs.

P2-09 now persists the original versioned HTTP response artifact in the same PostgreSQL transaction
as the financial mutation, replays stored bytes/status/request/trace identity without rereading mutable
resource state, validates response-body integrity, serializes duplicate command resolution, survives a
new Store connection, and preserves import replay only for otherwise-valid authentic expired review
proofs. Legacy completed rows without an exact artifact fail closed.

P2-13 now persists only short-lived opaque technical retry metadata in `sessionStorage`, derives the
storage slot from SHA-256 of stable authenticated principal + operation + optional portfolio scope,
preserves unresolved retry identity across reload/remount, rotates changed intent, clears confirmed
success/conflict state, and isolates User A/User B retry journals in the same browser tab without
persisting raw principal, portfolio, financial payload, CSV, review token, or auth tokens.

The first independent review returned `REQUEST CHANGES` because the browser retry namespace was not
principal-scoped. The remediation added stable-principal scoping and an explicit A→B→A regression.
Repeat independent review on the final exact head returned `APPROVED`, marking P2-09 and P2-13
CLOSED with no new blocking P1/P2 regression. Explicit human squash-merge authorization was then
received before PR #67 merged.

Stage 3.32 is CLOSED for P2-09/P2-13. Closure governance PR #68 was squash-merged into `develop` at
`a73b7f8c008d2f903e22e9b8a85b7c6248d6d3be`. The remaining original repository-audit backlog after
Stage 3.32 closure was 5 P2 and 10 P3 findings: P2-10/P2-11/P2-12/P2-16/P2-17 plus all P3 findings.
Stage 3.25 privacy Security Review evidence planning remains separate and is not superseded.


## Stage 3.33 audit remediation closure governance

Stage 3.33 is a narrow repository-audit remediation for P2-10, P2-11, and P2-12 only.

Implementation PR #69 was squash-merged into `develop` at
`87a7c38e16062a5f3fcef3727f60c0c6741eb805`. The exact independently reviewed implementation head
was `88ec8f739f7bcc96267c25f41560e1960d4d48d5`; exact-head GitHub Actions CI #199 completed
`SUCCESS` across all six jobs.

P2-10 moves ownership of the exact `snapshotDatesRebuilt` result into PostgreSQL. The database builds
the deterministic union of imported trade dates and applicable existing later snapshots, and the
DB-owned outcome is carried into the exact replay artifact before commit.

P2-11 replaces per-trade-date cascading snapshot rebuilds on the canonical import path with one sorted,
deduplicated affected-date plan rebuilt exactly once while the same-portfolio lock remains held.
Version-based PostgreSQL regression coverage proves pre-existing later snapshots advance exactly once.

P2-12 enforces the append-only runtime boundary with a dedicated least-privilege PostgreSQL capability
role and fail-closed startup validation of the authenticated credential graph. Validation rejects
privileged or masked sessions, protected-schema CREATE, protected-table owner/mutation capability,
SET-reachable escalation roles, and `MEMBER WITH ADMIN OPTION` role-administration paths. PostgreSQL
integration coverage proves legitimate append succeeds while direct mutation, masked-session,
latent SET-role, and latent ADMIN OPTION escalation scenarios are rejected.

The first independent review closed P2-10/P2-11 but returned `REQUEST CHANGES` for P2-12 because
only `current_user` was initially trusted. The first remediation added same-connection
`session_user`/`current_user` validation and SET-reachable credential-graph checks. The second
independent review again returned `REQUEST CHANGES` because `ADMIN TRUE, INHERIT FALSE, SET FALSE`
remained a latent escalation path. The second remediation rejected ADMIN OPTION capability as a class
for the authenticated principal and all SET-reachable roles. Final repeat independent review returned
P2-10 CLOSED, P2-11 CLOSED, P2-12 CLOSED, no new blocking regressions, and `APPROVED`. Explicit human
squash-merge authorization was received before PR #69 merged.

Stage 3.33 is CLOSED for P2-10/P2-11/P2-12. Closure governance PR #70 was squash-merged into `develop`
at `71a1faeb97d33d05f2936111b53f1285edddabe9`. The remaining original repository-audit backlog after
Stage 3.33 closure was exactly 2 P2 and 10 P3 findings: P2-16/P2-17 plus all P3 findings. Stage 3.25
privacy Security Review evidence planning remains separate and is not superseded.


## Stage 3.34 audit remediation closure governance

Stage 3.34 is the separately governed final P2 remediation for P2-16 and P2-17 only. Planning PR #71
was squash-merged into `develop` at `b4299bcdc28202c27388642dc7b426b159bb315c` after exact-head
CI #205 and repeat independent planning `APPROVED` following correction of the initial admin-bypass
planning defect.

P2-17 implementation PR #80 was squash-merged into `develop` at
`c686a6721df51063ccf62a0303bb759d2215d60e`. Its frozen independently reviewed head
`fd3a72a159161ec0bdf8018fdbf6e0a3da361885` passed primary exact-head CI #230 and secondary
exact-head CI #229 with all ten jobs successful. The implementation preserved the original six jobs
and added Go vet, PostgreSQL-backed Go race tests, pinned govulncheck, dependency security scanning,
and nightly/manual verification. Independent implementation review returned `APPROVED`, P2-17
CLOSED CANDIDATE, and explicitly left P2-16 OPEN until repository settings were mechanically enforced.
Explicit human squash-merge authorization was received before PR #80 merged.

P2-16 is now mechanically enforced by repository settings rather than policy alone. The repository is
public after the separately authorized full-history public-release audit; `develop` is protected with
pull-request entry, the ten final Stage 3.34 GitHub Actions checks, conversation resolution, linear
history, no normal administrator bypass, force-push disabled, and deletion disabled. Repository merge
policy is squash-only: squash enabled, merge commits disabled, rebase merge disabled.

Stage 3.34 closure is tracked through PR #82. The first independent closure review on head
`b208ddfdca9fcc3ddb3589fb7d47c698fa9f0dfa` returned `REQUEST CHANGES` for canonical-status drift:
`SOURCE_OF_TRUTH.md` and `ROADMAP.md` still represented Stage 3.33 / two remaining P2 findings as the
current state. This revision synchronizes those higher-level governance documents. Because the head
changed, the prior closure review does not carry forward; PR #82 requires fresh exact-head green CI,
fresh independent `APPROVED`, and explicit human squash-merge authorization before merge.

Stage 3.34 closure through PR #82 is canonical on `develop`. P2-16 and P2-17 are CLOSED and the original 32-finding audit backlog is exactly P0=0, P1=0, P2=0, P3=10 before Stage 3.35 P3-01 closure. After Stage 3.35 P3-01 closure, the remaining findings are P3-02 through P3-10. P3-09
Next.js maintenance and P3-10 Fiber maintenance remain separately governed. Stage 3.25 privacy
Security Review evidence planning remains separate and is not superseded.

## Stage 3.36 audit remediation closure governance

Stage 3.36 is the separately governed P3-03 OpenAPI Decimal Grammar remediation. Planning PR #87
was squash-merged into `develop` at `251296e0831cbb0b81c7799cc82cbdf3b451ae6e`. Runtime PR #88
was squash-merged into `develop` at `ebbc1c17b905e60d9e82337fc4a1ecd6cf9bccaa`; the frozen final
runtime head was `131f1bf963e9d232b9e23273edd54caf54c10ffb` and exact-head GitHub Actions CI #257 /
run `32822925542` completed successfully across all 10 required jobs.

The runtime narrows `decimal.FromString` to the published ASCII Decimal grammar, rejects oversized
lexemes before `big.Int` conversion, moves import review semantics to parser version 2, prevents a
parser-v1 review token from authorizing a fresh financial write, and retains only authenticated exact
read-only recovery of an already-completed parser-v1 import artifact. It does not alter Decimal
arithmetic, rounding, PostgreSQL `NUMERIC(28,8)`, migrations, stored financial values, or any other
P3 finding.

GitHub-native submitted review records and PR discussion comments for PR #88 are empty in the
repository API. This Source of Truth therefore does not invent a hosted external-review artifact.
Closure governance PR #89 re-reviewed the merged runtime evidence together with the documentation-only
closure diff, passed exact-head closure CI #258 / run `32827433078` on head
`7fd4e549e396cd9d99dff5c6d1678d0a4099fb70`, received independent `APPROVED`, received separate
explicit human squash-merge authorization, and was squash-merged into `develop` at
`9c83b68e28bbb8bc971620d3e00be5e177ce0820`.

P3-03 is canonically CLOSED. The current original audit backlog is P0=0 / P1=0 / P2=0 / P3=8:
P3-02, P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 privacy Security Review
evidence planning remains separate and is not superseded.

## Stage 3.37 audit remediation closure governance

Stage 3.37 is the separately governed P3-02 True IANA Timezone Semantics remediation. Planning PR #90
was squash-merged into `develop` at `46f74528dcc19424ad087d30d4f2f778e2079b87`. Runtime PR #91
was squash-merged into `develop` at `cb6d9b28cd47b1cd283b5861b916e0be627d0ac2`; the frozen final
runtime head was `1a2f89a0fa5095b3cca790521afa484bdc61e8a6` and exact-head GitHub Actions
CI #265 / run `32869754524` completed successfully across all 10 required jobs.

The runtime validates the exact submitted timezone identity, rejects `Local`, surrounding-whitespace
forms, and raw ASCII `±HH:MM` / `UTC±HH:MM` before resolver admission, explicitly retains `UTC`,
delegates other exact identifiers to `time.LoadLocation`, retains `_ "time/tzdata"` only as fallback
availability, preserves loadable identifiers such as `Etc/GMT+4`, and persists accepted identity
unchanged. It does not alter BusinessDate, SQL `DATE`, SystemTimestamp, financial calculations,
migrations, historical preference rows, or any other P3 finding.

The first independent pre-commit runtime review returned `REQUEST CHANGES` for a custom-`ZONEINFO`
whitespace/raw-offset bypass; the revised candidate closed that bypass and received renewed
`APPROVED`. The first published-head review returned `REQUEST CHANGES` only for stale lifecycle
wording in the implementation record. The independently pre-commit-approved documentation correction
advanced the final head to `1a2f89a0fa5095b3cca790521afa484bdc61e8a6`; CI #265 succeeded on that
exact head and the fresh final published-head review returned `APPROVED` with no P0/P1/P2/P3 finding.

Closure governance PR #92 was published from exact closure head
`dc206d9d9cf68b00c182ce1a247e5dcc289e16e8`, passed exact-head GitHub Actions CI #266 / run
`32880893193` with all 10 required jobs successful, received fresh independent closure `APPROVED`,
received separate explicit human Ready + squash-merge authorization, and was squash-merged into
`develop` at `305a53bb07136b274717ff48778a5e93d7b1607c`.

P3-02 is canonically CLOSED. The current original audit backlog is P0=0 / P1=0 / P2=0 / P3=7:
P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 privacy Security Review evidence
planning remains separate and is not superseded.

## Stage 3.38 audit remediation planning status

Stage 3.38 is the separately governed planning/review candidate for original finding P3-05 —
idempotency/session retention and cleanup. It is planning-only at this point. It does not authorize
runtime implementation, it does not close P3-05, and it does not alter the current P3=7 backlog.

The planning candidate freezes a bounded 24-hour idempotency/replay authority window, persisted
per-session expiry authority, post-serialization/post-lock expiry decisions, concurrency-safe
exact-key reclamation, bounded 128-row opportunistic PostgreSQL cleanup, dedicated expiry-leading
indexes, and mandatory expiry-boundary race regressions. The first independent pre-commit planning
review returned `REQUEST CHANGES` for a stale-clock/serialization race; the revised v2 candidate
requires authoritative expiry time only after the relevant serialization point and received renewed
independent pre-commit `APPROVED`.

Stage 3.38 planning is not canonical merely because this candidate exists. It becomes canonical only
through its own separately reviewed, exact-head-CI-green, explicitly authorized planning PR merge.
P3-04, P3-06, P3-07, P3-08, P3-09, P3-10, and Stage 3.25 remain separately governed.

## Stage 3.38 audit remediation closure governance

Stage 3.38 is the separately governed P3-05 idempotency/session retention and cleanup remediation.
Planning PR #93 was squash-merged into `develop` at `a944f1e5d5ee7d84db5393e8760eda254d732edd`.
Runtime PR #94 published exact head `5ea8c6f4eddd735ea834dc4a27ecb70da7f81508`, which passed GitHub Actions
CI #268 / run `32913862780` with all 10 required jobs successful and received fresh
independent published-head `APPROVED` with P0/P1/P2/P3 = None. The user then separately authorized
Ready + squash merge of that exact head. PR #94 was squash-merged into `develop` at
`2df9946d77ee044a191a0422c8cccbbfe02dc7c9`, and `develop` was read back at that exact SHA.

The runtime makes command/replay authority expire at the frozen 24-hour boundary, uses fresh PostgreSQL
wall clock after required serialization for mutation authority, permits exact-key fresh-generation
reclamation after expiry, removes expired-session containment/revocation authority, preserves
revoked-but-unexpired Stage 3.28 containment, and performs bounded 128-row opportunistic cleanup with
`ORDER BY expires_at,id` and `FOR UPDATE SKIP LOCKED` inside triggering mutation transactions.

The runtime review history is intentionally preserved: initial local SQLSTATE `42P08`; first v2
`REQUEST CHANGES`; renewed-v3 `REQUEST CHANGES`; exact-v4 documentation-only `REQUEST CHANGES`;
DOCFIX2 historical-quotation `REQUEST CHANGES`; DOCFIX3 stale-review-count `REQUEST CHANGES`; final
DOCFIX4 pre-commit `APPROVED`; exact-head CI; and fresh published-head `APPROVED`.

This closure candidate changes governance documentation only. It does not modify runtime code, SQL,
OpenAPI, privileges, retention semantics, BusinessDate/SQL `DATE`, SystemTimestamp, Decimal or
financial arithmetic, immutable ledger/snapshot rules, account deletion, Stage 3.25 privacy work, or
P3-04/P3-06/P3-07/P3-08/P3-09/P3-10.

Until the closure PR itself receives fresh independent closure `APPROVED`, exact-head green CI after
publication, separate explicit human squash-merge authorization, and is merged into `develop`, P3-05
remains canonically OPEN and the original audit backlog stays P0=0 / P1=0 / P2=0 / P3=7. After that
closure merge, P3-05 becomes CLOSED and the remaining backlog is P3=6: P3-04, P3-06, P3-07, P3-08,
P3-09, and P3-10.

## Stage 3.42 P3-09 closure governance

Stage 3.40 planning is canonical through PR #100 at
`559b57d0951cdc67125c2f72fc1fcfb34399e90e`. Stage 3.41 implementation is canonical through
PR #101 final evidence head `d88be3c90231f374d7e6b7d94f4cd89e6788f700`, CI #291 / run
`33277717164` with all 10 required jobs successful, and actual squash merge
`a2cfeaa5ca68fdd951e2a99f69c96aec362fc416`.

Stage 3.42 is documentation/governance-only closure activation and changes no runtime/dependency,
OpenAPI, database, migration, security/session, financial, privacy, or architecture behavior.

Before the approved Stage 3.42 closure record and synchronized canonical surfaces are present on
protected `develop`, P3-09 remains OPEN. The original audit state is 27/32 closed (84.375%) with
exactly five remaining findings: P3-06, P3-07, P3-08, P3-09, P3-10.

Once the approved Stage 3.42 closure record and synchronized canonical surfaces are present on
protected `develop`, P3-09 is CLOSED. The original audit state becomes 28/32 closed (87.5%) with
exactly four remaining findings: P3-06, P3-07, P3-08, P3-10.

No future Stage 3.42 PR number, published head, CI run, or squash-merge SHA is predicted here.
