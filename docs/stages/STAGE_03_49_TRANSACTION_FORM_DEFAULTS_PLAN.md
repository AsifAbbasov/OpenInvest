# Stage 3.49 — P3-07 Transaction Form Fixture / Default Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning-only governance scope; this artifact grants no implementation authorization |
| Date | 2026-09-01 |
| Canonical planning base | `develop@f24d8df2e3aa7fd44560f5fee5b2ef9ccdd8bca1` |
| Protected-base tree | `b635f878f0b1198028ea8f88257173c441dc2e61` |
| Finding | Original audit `P3-07 — transaction form fixture defaults` / `transaction-form fixture/default semantics` |
| Planning-base audit state | 30 / 32 closed = 93.75%; remaining P3-07, P3-08 |
| Implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / Ready / merge authorized here | No |

## 1. Objective

Close original audit finding P3-07 by removing Stage 3.3 demonstration/fixture business values from
the initial state of the production `AddTransactionForm` while preserving the existing
presentation-only architecture, Go API authority, idempotency behavior, transaction payload contract,
and transaction-type field applicability.

This stage is planning-only. It changes no production source, tests, dependencies, OpenAPI, backend,
database, migration validator, workflow, or protected branch.

## 2. Exact planning baseline

At protected `develop@f24d8df2e3aa7fd44560f5fee5b2ef9ccdd8bca1`:

- protected-base tree: `b635f878f0b1198028ea8f88257173c441dc2e61`;
- `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx` blob:
  `0c5e423ee1a02044b0500d10d03bb0f407b62a96`;
- `frontend-next/src/common/api/openinvest.ts` blob:
  `20d92d90f0d99c0049e9d16d5ca5bbc125a1d759`;
- `frontend-next/package.json` blob:
  `d6d605620e1bff426998d8bda716b7c2eda0613d`;
- `frontend-next/AGENTS.md` blob:
  `f930da94e0a124b7e7a90ae1d2efaf73631fbffb`;
- `docs/REVIEW_WORKFLOW.md` blob:
  `06f9cabd04e6791be1892ae1f0eae8d915fddc02`;
- Stage 3.48 closure blob:
  `3a720c20d5515d41d45b86a9c52094f5758361ec`.

The Stage 3.48 protected activation closed P3-06. At this planning base the original audit is
30/32 closed (93.75%) and the only remaining original findings are P3-07 and P3-08.

## 3. Planning-base defect evidence

At the immutable Stage 3.49 planning base, the production `AddTransactionForm` initializes
user-editable transaction business fields with demonstration values inherited from the early Web
presentation slice:

- ticker: `SBER`;
- quantity: `10.00000000`;
- unit price: `280.00000000`;
- gross amount: `2800.00000000`;
- commission: `2.80000000`;
- tax: `0.00000000`;
- trade date: `2026-01-10`;
- settlement date: `2026-01-13`;
- note: `Stage 3.3 Web presentation slice`.

The form also initializes the transaction-type selector to `BUY` and always serializes currency as
`RUB`.

P3-07 concerns the **fixture/business-value defaults**, not every initial UI control value and not
the already-frozen RUB contract.

## 4. Failure scenario and impact

A production data-entry form that opens with plausible security, amount, fee, date and note values can
make demonstration data look like user data. A user can overlook one or more prefilled values and
submit an immutable ledger transaction containing stale example facts.

The fixed historical dates are additionally guaranteed to become increasingly misleading as time
passes. Replacing them with the current date would still manufacture a business fact that the user did
not provide, so that is not an acceptable remediation.

This is P3 correctness/UX debt. The plan does not claim a demonstrated security exploit or backend
ledger-integrity bypass: the Go API remains authoritative and validates submitted requests.

## 5. Frozen default taxonomy

Implementation SHALL distinguish control/contract defaults from user business facts.

### 5.1 Business-value fields — MUST start empty

On initial render, all of the following form state values MUST be empty strings:

- `ticker`;
- `quantity`;
- `unitPrice`;
- `grossAmount`;
- `commission`;
- `tax`;
- `tradeDate`;
- `settlementDate`;
- `note`.

No Stage 3.x example value, security ticker, quantity, price, amount, fee, tax, trade date, settlement
date or note may be pre-populated in these fields.

No automatic “today”, “T+N”, browser-local date, portfolio-derived amount, last-used value, asset
search result, or other inferred business fact may be introduced as a replacement default in P3-07.

### 5.2 Control default — MAY remain `BUY`

`transactionType = "BUY"` may remain the initial selector value because the select needs an initial
branch to determine which fields are rendered. This is a form-control choice, not evidence that a BUY
transaction occurred.

P3-07 does not expand the transaction types exposed by the Stage 3.3 Web form. The existing visible
set remains frozen for this remediation unless a separate reviewed scope changes it.

### 5.3 Contract invariant — RUB remains fixed

The payload's `currency: "RUB"` behavior is an existing API/product contract invariant, not fixture
business data. P3-07 must not turn currency into a new editable field and must not alter backend or
OpenAPI currency semantics.

### 5.4 Placeholder text

Human-readable placeholders MAY be added for entry guidance only if they are not form values, are not
serialized into the payload, and cannot be confused with persisted defaults. Placeholder addition is
optional and is not required to close P3-07.

## 6. Payload and field-applicability invariants

The implementation must preserve the current payload-shape rules:

- cash flows (`DEPOSIT`, `WITHDRAWAL`) submit `ticker = null`;
- cash flows submit `quantity = null`;
- cash flows submit `unitPrice = null`;
- trades use quantity and unit price;
- asset income types, if ever admitted by the existing control surface, retain their current
  quantity/gross-amount applicability;
- non-trade gross amount behavior remains unchanged;
- empty settlement date serializes as `null`;
- empty normalized note serializes as `null`;
- note Unicode validation remains unchanged;
- ticker normalization remains trim + uppercase where applicable;
- idempotency intent identity remains derived from the actual payload;
- an idempotency conflict retains the current retry-slot clearing behavior.

Changing transaction arithmetic, introducing frontend calculations, deriving gross amount from price ×
quantity, or changing backend validation is outside P3-07.

## 7. Type-switch stale-value isolation

When a user enters a value and changes transaction type, a hidden field must not leak into a payload
for a type where that field is inapplicable.

The existing payload builder already gates fields by transaction type. P3-07 implementation must
preserve that property and add regression evidence sufficient to catch stale hidden-value leakage.

The remediation does not require destructive clearing of every hidden React state value on each type
change if payload nullability/gating deterministically prevents leakage.

## 8. Planned implementation surface

The default implementation scope is intentionally small:

1. `frontend-next/src/features/portfolio/components/AddTransactionForm.tsx`;
2. one or more focused frontend tests under `frontend-next/tests/`.

A small private presentation helper under the same `frontend-next/src/features/portfolio/` boundary may
be introduced only if it materially improves deterministic testing without adding business logic or a
new architecture layer.

Default expectation: no more than 3 changed implementation files.

The implementation must comply with `frontend-next/AGENTS.md`: presentation-only, no database access,
financial calculation, external data source, business persistence, or backend business logic.

## 9. Explicitly out of scope

P3-07 MUST NOT include:

- P3-08 migration-validator policy hardening;
- backend Go changes;
- database/schema/migration changes;
- executable OpenAPI changes;
- transaction-domain or financial-arithmetic changes;
- new transaction types in the Web UI;
- currency model changes;
- asset-search integration into the form;
- automatic price lookup;
- automatic date or settlement-date derivation;
- post-submit form-reset redesign unless required by a discovered P3-07 correctness blocker;
- dependency or lockfile changes;
- Next.js/React maintenance;
- CSS/design-system redesign;
- Stage 3.25 privacy work.

Any discovered requirement for one of these surfaces is a fail-closed scope-expansion event and returns
to planning/review before implementation proceeds.

## 10. Regression and adversarial verification requirements

The implementation must add deterministic evidence for at least the following:

1. initial business-value inputs render empty;
2. the transaction-type selector retains the planned control default;
3. the production form no longer contains the Stage 3.3 fixture values listed in section 3 as initial
   values;
4. empty optional settlement date maps to `null`;
5. empty normalized note maps to `null`;
6. type-specific hidden values do not leak into inapplicable payload fields;
7. user-entered values, not placeholders, are what reach the request payload;
8. the existing idempotency path remains intact;
9. existing Unicode note validation remains intact.

Tests may use the existing Node/tsx/jsdom test infrastructure. No new test dependency is authorized.

In addition to focused tests, implementation validation must run:

- `pnpm test`;
- `pnpm run typecheck`;
- `pnpm run build`;

and the published implementation head must pass all repository-mandatory GitHub CI checks.

## 11. Review-size budget

The canonical default review limits apply:

- no more than 25 changed files;
- no more than 800 changed lines of hand-written business logic.

Stage 3.50 is expected to remain far below both limits. If either limit would be exceeded, stop before
Internal review unless the canonical Principal Architect exception is explicitly documented and
approved before review begins.

Do not silently split P3-07 across multiple implementation PRs merely to satisfy the budget. A
multi-PR implementation strategy requires a separately reviewed lifecycle/evidence plan first.

## 12. Development review path for the later implementation

Because Stage 3.50 will change production Web source, it follows the mandatory development path:

1. exact approved planning base/plan;
2. dedicated feature branch from the then-approved implementation base;
3. implementation + focused tests;
4. local gates;
5. complete read-only Internal review in the designated review chat;
6. remediation and gate rerun when required;
7. explicit human commit/push authorization;
8. Draft PR;
9. exact-head mandatory GitHub CI;
10. fresh External published-head review in the same designated review chat;
11. evidence-only follow-up only if required by the workflow;
12. exact evidence-head CI / publication verification when applicable;
13. separate human Ready authorization;
14. separate human squash-merge authorization;
15. implementation merge;
16. separately governed closure activation before P3-07 becomes CLOSED.

Internal approval does not authorize publication. External approval does not authorize Ready or merge.
Implementation merge alone does not close P3-07.

## 13. Planning publication lifecycle

Stage 3.49 itself is a one-document planning change.

The permanent planning artifact must remain publication-stable:

- it predicts no future PR number, published head, CI run number or squash SHA;
- reviewer verdicts are evidence bound to exact candidate identities, not mutable self-authored
  approval fields inside this document;
- mutable baseline facts are anchored to the immutable Stage 3.49 planning base;
- P3-07 remains OPEN throughout planning review/publication/merge;
- P3-08 remains independently OPEN and unaffected.

Planning publication requires its own review and explicit human authorizations. Nothing in this
document grants commit, push, PR, Ready, merge, implementation or closure authority.

## 14. Audit arithmetic

At the immutable Stage 3.49 planning base:

- closed: 30 / 32;
- completion: 93.75%;
- P0 = 0;
- P1 = 0;
- P2 = 0;
- P3 = 2;
- remaining: P3-07, P3-08.

Planning merge does not change that arithmetic.

After a future separately governed P3-07 implementation and protected closure activation:

- closed would become 31 / 32;
- completion would become 96.875%;
- remaining original finding would be P3-08.

That future conditional arithmetic is not a claim that P3-07 is already closed.

## 15. Planning decision

Proceed with a narrow frontend-only P3-07 remediation after this plan is independently approved and
canonically merged.

The planned remediation removes fixture-derived business facts from the initial production transaction
form state, preserves existing transaction payload applicability/idempotency/Unicode behavior, adds
focused regression coverage, and leaves P3-08 untouched.

P3-07 remains OPEN until the later implementation is complete and a separately governed closure is
activated on protected `develop`.
