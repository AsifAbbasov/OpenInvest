# Stage 3.60 — MOEX ISS Runtime / Source Activation Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-60-MOEX-ISS-RUNTIME-SOURCE-ACTIVATION-PLAN |
| Version | 0.1.0-candidate |
| Status | Planning / source-governance decision only; NO shipped-runtime activation authorized |
| Owner | Principal Architect / Market Data |
| Canonical planning base | `develop@f55ad38c1f5ea52ba4502904fefa51c164c45006` |
| Protected-base tree | `42596285fa8de467108ee1553b1575718761c7f2` |
| Dependencies | Stage 3.57 provider boundary; Stage 3.59 planning + implementation; `docs/registries/DATA_SOURCE_REGISTRY.md`; `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Date | 2026-09-04 |

## 1. Purpose

Stage 3.60 decides whether the already-implemented `MOEX_ISS` quote adapter may be activated in the shipped OpenInvest runtime under the project's current constraints.

This stage does not implement or wire runtime behavior. It is a fail-closed source/terms decision.

The decision is:

```text
MOEX_ISS adapter implementation   MERGED
shipped/runtime activation        NO-GO
public API/UI display             NO-GO
background/non-display use        NO-GO
reason                            current MOEX usage terms + OpenInvest zero-budget constraint
```

The adapter remains dormant behind the application-owned `QuoteProvider` boundary.

## 2. Current protected state

At the planning base:

- protected `develop` is `f55ad38c1f5ea52ba4502904fefa51c164c45006`;
- protected tree is `42596285fa8de467108ee1553b1575718761c7f2`;
- Stage 3.59 planning and implementation are merged;
- `backend-go/internal/provider/moexiss` contains the real delayed TQBR adapter;
- `backend-go/cmd/api/main.go` still constructs `verticalslice.NewService(...)` without a quote provider;
- therefore the shipped application performs no MOEX quote request and exposes no MOEX quote to users;
- no public quote endpoint, asset-search quote enrichment, MOEX cache, MOEX persistence, polling worker, or background ingestion exists.

That dormant state is the required baseline for this decision.

## 3. Official MOEX terms evidence

The decision relies only on current official Moscow Exchange materials reviewed on 2026-09-04.

### 3.1 ISS delayed access is technically available without subscription

Official ISS materials state that market data may be provided in real time by subscription or with delay without subscription/authorization.

Reference:

- `https://www.moex.com/a2193`
- `https://www.moex.com/a8531`

Technical accessibility does not itself grant product-use, redistribution, display, automated-processing, or commercial rights.

### 3.2 ISS information without a contract is limited to familiarisation

The official MOEX ISS page states that information received from ISS is available only for familiarisation and cannot be used for profit or other actions beyond familiarisation, including providing that information or materials/data/products/services based on it to third parties. Any use beyond familiarisation requires an agreement with PJSC Moscow Exchange.

Reference:

- `https://www.moex.com/a2193`

This is directly incompatible with treating free delayed ISS access as an automatically approved production data licence for OpenInvest.

### 3.3 Public web/app placement requires an information agreement

MOEX's public-data page states that public placement of exchange information on websites or applications requires an information agreement with Moscow Exchange.

The reviewed tariff page lists, for non-issuers:

- real-time public market data: `300,000 RUB/month`;
- 15-minute delayed public information: `25,500 RUB/month`.

References:

- `https://www.moex.com/ru/products/publicdata`
- `https://www.moex.com/s1147`

OpenInvest's current project constraint is zero-budget/free-tier infrastructure and data access. Stage 3.60 therefore does not accept a paid MOEX market-data contract.

### 3.4 Non-display and derived-data use are separately governed

The Moscow Exchange Market Data Policy explicitly governs purposes including further distribution, Non-display systems, and Derived Data and requires the appropriate agreement for the relevant use mode.

References:

- `https://www.moex.com/en/datapolicy/`
- `https://www.moex.com/ru/datapolicy/`
- `https://www.moex.com/s3503`

Therefore hiding MOEX data from the UI does not create a loophole that automatically permits production automated processing, portfolio calculations, alerts, derived analytics, or background ingestion.

## 4. Activation decision

Under the current source evidence and project constraints:

```text
DECISION = NO-GO FOR SHIPPED RUNTIME
```

Stage 3.60 does not authorize any of the following:

- `cmd/api` composition-root wiring to `moexiss.NewQuoteProvider`;
- user-triggered shipped MOEX requests;
- public quote or asset-price API responses sourced from MOEX;
- Web/mobile display of MOEX quote data;
- background polling or scheduled MOEX retrieval;
- cache or database persistence of MOEX quote data for product operation;
- historical market-data ingestion;
- use of MOEX market data for alerts, portfolio valuation, deterministic insights, derived indicators, or other product calculations;
- onward distribution or redistribution;
- real-time subscription;
- treating delayed unauthenticated ISS access as an OpenInvest production licence.

No runtime code change is necessary to enforce this decision because the merged composition root remains provider-free.

## 5. What remains allowed

The following remains allowed inside the already-approved engineering/test boundary:

- retaining the merged adapter source code;
- deterministic unit and `httptest` coverage;
- CI compilation/testing/security analysis;
- static code review;
- provider-contract documentation;
- optional human-run technical smoke evidence used only to verify adapter compatibility, provided it is not shipped as a product feature, automated as a production collector, cached/persisted for product use, or redistributed.

These allowances do not convert adapter existence into runtime-source approval.

## 6. Fail-closed composition rule

Until a later source-activation stage is approved and protected-merged, the shipped composition root must remain logically equivalent to:

```text
cmd/api
  ↓
NewService(...)
  ↓
quoteProvider = nil
```

It must not become:

```text
cmd/api
  ↓
NewServiceWithQuoteProvider(..., moexiss.Provider)
```

A future accidental provider wiring before source approval is a governance defect and must fail review.

## 7. No product-level workaround

Stage 3.60 rejects attempts to bypass the source decision by changing transport or presentation shape.

The following do not make the use automatically acceptable:

- calling MOEX only when a user opens a page;
- calling MOEX server-side instead of from the browser;
- suppressing provider attribution;
- displaying delayed rather than real-time data;
- returning derived calculations instead of raw prices;
- storing only the latest value;
- using MOEX only for portfolio valuation or notifications;
- calling the feature development/demo/beta while it is shipped to third parties.

Any of those use modes requires a separately reviewed rights/terms basis.

## 8. Future activation prerequisites

A future Stage may reconsider activation only after all of the following are established with evidence:

1. exact intended use mode is defined: display, redistribution, non-display, derived-data, internal use, or another MOEX-defined category;
2. Moscow Exchange contractual/usage rights for that exact mode are documented;
3. required attribution/display/audit obligations are understood;
4. monetary cost is known and explicitly accepted by the Principal Architect;
5. rate-limit/traffic expectations are documented;
6. caching/retention/persistence rights are documented if any such behavior is planned;
7. the Data Source Registry is updated from NO-GO to an exact approved production use mode;
8. runtime composition wiring receives a separately reviewed implementation plan;
9. public API/UI contract changes, if any, receive their own review;
10. exact-head CI, fresh External review, evidence publication, and explicit human merge authorization pass.

Absence of evidence for any required right remains UNKNOWN and therefore cannot be converted into production permission.

## 9. Zero-budget consequence

The current OpenInvest constraint is MacBook development plus free services/APIs/free tiers.

The reviewed official MOEX public delayed-data tariff is not zero-cost. Stage 3.60 therefore does not recommend or authorize purchasing that service.

The correct engineering consequence is to keep the high-quality MOEX adapter dormant rather than weaken source governance.

## 10. Roadmap consequence

Stage 3.60 does not block further product development.

The next product-facing work item should move to:

```text
Dividend / Coupon Calendar + Heatmap
```

but its planning must begin with **corporate-actions source discovery and Data Source Registry approval**, not by assuming MOEX price-data rights also authorize corporate-action use.

The next stage must independently determine whether a zero-cost source exists whose terms permit the required OpenInvest use mode.

## 11. Explicit non-scope

Stage 3.60 does not include:

- Go runtime changes;
- `cmd/api` changes;
- HTTP/API handlers;
- OpenAPI changes;
- frontend/mobile changes;
- SQL/migrations;
- quote persistence;
- Redis/cache;
- workers/schedulers;
- MOEX retries/batching/polling;
- a paid MOEX contract;
- corporate-action provider implementation;
- dividend/coupon calendar implementation;
- Feature 3 implementation.

## 12. Acceptance criteria

This planning decision is acceptable only when review confirms:

- official source evidence supports the distinction between technical ISS access and permitted product use;
- no claim treats free delayed access as production rights;
- public display and non-display/derived-data concerns are not conflated;
- zero-budget constraint is respected;
- source registry remains fail-closed;
- shipped composition remains provider-free;
- no runtime/API/DB/frontend/dependency drift is introduced;
- future activation has explicit evidence prerequisites;
- the next calendar work cannot inherit unapproved MOEX rights by implication.

## 13. Review focus

Planning review must challenge:

- whether NO-GO is actually supported by official MOEX materials;
- whether any text accidentally authorizes a hidden non-display production path;
- whether optional smoke testing is scoped narrowly enough;
- whether a dormant adapter is safe to retain;
- whether the zero-budget consequence is explicit;
- whether future activation requires a new registry decision rather than silently reusing this stage;
- whether Feature 3 remains separate.

## 14. Next governed action

This document and the synchronized Data Source Registry candidate must receive mandatory read-only Internal planning/governance review.

Only after `APPROVED` and a separate human commit/push/Draft-PR authorization may this two-file documentation/governance candidate be published.

Protected merge of this planning decision does not authorize MOEX runtime activation. It makes the NO-GO source decision canonical.
