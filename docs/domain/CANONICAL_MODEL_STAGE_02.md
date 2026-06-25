# Stage 2 Canonical Model

| Field | Value |
| --- | --- |
| Document ID | DOMAIN-STAGE-02 |
| Version | 1.0.0 |
| Status | Proposed / Awaiting Review |
| Owner | Principal Architect |
| Supersedes | Conflicting DTO shapes in legacy documents |
| Dependencies | Documents 42–43; ADR-003; ADR-004; ADR-006 |
| Last Review Date | 2026-06-20 |
| Next Review Date | At Stage 2 approval |

## Purpose

Define one vocabulary shared by the API contract, future Go domain layer, Python analytics,
PostgreSQL design, and clients. This is a semantic freeze, not generated code and not a promise
that persistence rows will mirror public DTOs one-for-one.

## Context boundaries

| Context | Owns | Must not own |
| --- | --- | --- |
| Identity | account, credential, session, privacy preferences | portfolio ledger and calculations |
| Investment | portfolio identity, immutable transactions, canonical asset/security master | passwords, provider collectors, analytics formulas |
| Analytics | snapshots, calculated summaries, methodology versions | canonical transaction writes |
| External Data Gateway | provider collection and normalization before canonical ingestion | user portfolios, canonical ownership, client APIs |
| Audit | append-only security/business action evidence | secrets or financial document contents |
| Tax (future) | isolated tax-profile/export data behind feature flag | MVP portfolio behavior |

Cross-context API DTOs are composed read models. They do not transfer ownership of the source
record to the consuming context.

## Primitive value objects

### Decimal

Base-10 value serialized as a JSON string. Binary `float`/`double` is forbidden in API,
domain, persistence, tests, snapshots, and financial events.

- API accepts at most 8 fractional digits.
- Internal financial arithmetic uses exactly 8 fractional digits.
- Public persisted/calculated output is quantized with half-even rounding.
- Monetary display uses 2 decimals and is a presentation concern; stored/API precision remains 8.
- Scientific notation, NaN, Infinity, locale separators, and currency symbols are invalid.

### Money

```text
Money = Decimal amount + Currency currency
```

`currency` is `RUB` in MVP. Arithmetic between different currencies is impossible by type and
validation, not silently converted. Negative money is valid where economic meaning requires it;
commands constrain sign according to transaction type.

### BusinessDate

Calendar date without time or timezone. Maps to SQL `DATE`. Used for trade, registry, payment,
settlement, dividend, snapshot, valuation, and tax-year dates. Market-day interpretation uses
the MOEX calendar. UTC conversion must never shift a BusinessDate.

### SystemTimestamp

UTC instant serialized as RFC 3339 ending in `Z`; maps to `TIMESTAMP WITH TIME ZONE`. Used only
for creation/update, audit, event, trace, worker, source-observation, and calculation timestamps.

## Identity DTOs

### User

| Field | Type | Rule |
| --- | --- | --- |
| `id` | UUID | Public opaque identity ID |
| `email` | email | Required identity attribute |
| `language` | `ru \| en` | Required |
| `theme` | `light \| dark \| system` | Required |
| `timezone` | IANA timezone | Display/notification only; never financial-date math |
| `privacyMode` | boolean | `true` for every new account |
| `createdAt` | SystemTimestamp | Audit timestamp |

Passwords and refresh tokens are command-only secrets and never canonical response DTO fields.

## Asset model

### AssetType

`STOCK | BOND`. ETF, fund, ZPIF, foreign security, and currency instruments are excluded from MVP.

### Asset

Discriminated union keyed by `assetType`:

- common: ticker, name, asset type, RUB currency, MOEX market, lot size, lifecycle status,
  normalized last cash price, price timestamp, registered source reference;
- stock: sector and ISIN;
- bond: ISIN, face value, maturity date, coupon type, nullable coupon rate.

Lot size and bond face value are strictly positive. Normalized market price and coupon rate are
non-negative when present. Provider payloads containing impossible negative magnitudes are rejected
before canonical ingestion; signed Money is reserved for fields whose meaning permits loss/debt.

MOEX ticker is the public lookup key. The Investment context owns the canonical security master and
normalized current Asset facts. The External Data Gateway owns provider adapters and normalization
before those facts cross the Investment ingestion port; it owns no canonical table or client API.
Provider-specific IDs may exist inside adapters but cannot leak into the client contract. Bond
prices are normalized to RUB cash price per unit; provider quote percentages are collector inputs,
not canonical API values.

## Portfolio aggregate

### Portfolio

| Field | Type | Rule |
| --- | --- | --- |
| `id` | UUID | Aggregate identity |
| `name` | string | Mutable metadata |
| `baseCurrency` | `RUB` | Frozen for MVP |
| `version` | positive integer | Optimistic concurrency for metadata |
| `createdAt` | SystemTimestamp | Immutable |
| `updatedAt` | SystemTimestamp | Changes with metadata lifecycle |

Portfolio does not embed transactions or snapshots. They are independently paginated resources.
Deleting portfolio metadata cannot physically delete financial history.

## Transaction ledger

### TransactionType

`BUY | SELL | DIVIDEND | COUPON | FEE | TAX | DEPOSIT | WITHDRAWAL`

### Transaction

| Field | Type | Rule |
| --- | --- | --- |
| `id` | UUID | Stable logical transaction identity |
| `portfolioId` | UUID | Owning portfolio reference |
| `transactionType` | TransactionType | Economic meaning |
| `status` | `ACTIVE \| CORRECTED \| REVERSED` | Current projection state |
| `ticker` | ticker/null | Required for asset events; null for pure RUB cash |
| `quantity` | Decimal/null | Required for BUY/SELL |
| `unitPrice` | Money/null | Per-unit cash price when applicable |
| `grossAmount` | Money | Canonical gross cash amount |
| `commission` | Money | Explicit; zero is represented, never omitted |
| `tax` | Money | Recorded amount only; not tax advice/export |
| `tradeDate` | BusinessDate | Economic date |
| `settlementDate` | BusinessDate/null | Separate from trade date |
| `note` | string/null | User annotation, length-limited |
| `revision` | positive integer | Current logical revision |
| `createdAt` | SystemTimestamp | Original record time |
| `updatedAt` | SystemTimestamp | Latest correction projection time |

#### Immutability invariant

The physical ledger is append-only. API PATCH appends a correction revision. API DELETE appends
a reversal transaction. No operation updates/deletes prior financial facts in place. A current
projection may be materialized, but audit can reproduce every revision and reversal.

Every reversal command carries an explicit `effectiveDate` BusinessDate. That date is the economic
date used for ledger projections, derived holdings, and deterministic snapshot rebuilds. System
timestamps such as request receipt, worker execution, `createdAt`, or audit time must never be used
to derive reversal economics.

#### Command validation matrix

| Type | Ticker/quantity/unit price | Gross amount |
| --- | --- | --- |
| BUY, SELL | required | server-derived and optionally verified; correction may replace it |
| DIVIDEND, COUPON | ticker required; quantity may be recorded | required |
| FEE, TAX | optional ticker | required |
| DEPOSIT, WITHDRAWAL | null ticker/quantity/unit price | required |

Create commands are a discriminated union whose required/null fields follow this matrix. A
correction contains a complete corrected command plus expected revision and audit reason, so the
same type-specific constraints apply and gross amount is explicitly correctable.

Transaction command magnitudes (`quantity`, `unitPrice`, `grossAmount`, `commission`, `tax`) are
non-negative. `transactionType` determines economic direction; storage and calculations must not
infer type from sign. Calculated gain/loss Money and cash-flow vectors may be negative where their
DTO/methodology explicitly permits it. These rules require financial test vectors before code.
The canonical Transaction response enforces the same type matrix: trade responses require asset,
positive quantity, and positive unit price; income responses require an asset and no unit price;
fee/tax responses have no quantity or unit price; pure RUB cash responses have no asset fields.

## PortfolioSummary

Backend-owned read model answering the primary product questions for one BusinessDate:

- total, cash, stock, and bond value;
- invested capital;
- dividends and coupons received;
- nominal return rate and nullable XIRR;
- RealReturn;
- PurchasingPower;
- positions with quantity, weighted-average cost, market price/value, unrealized gain, weight;
- CalculationMeta with methodology version, calculation timestamp, and input date.

`xirr=null` is a valid, explainable result when cash-flow history has no mathematical solution.
The client must not substitute zero.

## PortfolioSnapshot

Versioned, deterministic, rebuildable projection keyed conceptually by portfolio, snapshot date,
snapshot version, and methodology version. It records values and returns necessary for fast
historical rendering. Transactions remain the source of truth. Rebuild creates/replaces a
projection version; it never rewrites transaction history.

## DividendEvent

| Field | Rule |
| --- | --- |
| `ticker` | MOEX asset key |
| `status` | announced, approved, paid, cancelled only |
| `amountPerUnit` | RUB Money |
| `registryDate` | required BusinessDate |
| `paymentDate` | nullable separate BusinessDate |
| `settlementDate` | nullable separate BusinessDate |
| `source` | registered source code + observation timestamp |

Forecast status is deliberately absent. Coupon events may use a future dedicated model; they are
not disguised as dividends.

## DividendCalculation

Pure gross calculation DTO: ticker, quantity, dividend per unit, gross dividend, optional position
cost, nullable gross yield, `taxIncluded=false`, and methodology version. It does not promise a
payment, forecast an issuer decision, or calculate/export tax.

Quantity is strictly positive. Dividend per unit and calculated gross dividend are non-negative.
When supplied, position cost is strictly positive; zero or negative cost is rejected rather than
silently producing an undefined yield. `grossYield` is non-negative when position cost is present
and is `null` only when position cost is omitted. These calculator constraints do not define signed
portfolio cash-flow vectors.

## RealReturn

Contains nominal return rate, inflation rate, real return rate, nominal gain, real gain, start/end
BusinessDates, and methodology version. The formula and source inputs must be explainable. Rates
are Decimal strings, not percentages encoded as whole numbers: `0.06452885` means approximately
6.45%.

## PurchasingPower

Contains portfolio Money, BusinessDate, and equivalents. Every equivalent includes a stable
category, human label, RUB unit price, Decimal quantity, and registered source reference.
Equivalent unit price is strictly positive and its resulting quantity is non-negative.
Categories in MVP: iPhone, MacBook, average salary, food basket, utilities, apartment rent, car,
and square meter. Missing/stale source data removes or marks an equivalent at read-model creation;
the API must not fabricate a value.

Portfolio position quantity, weighted-average cost, market price, market value, and weight are
non-negative. Unrealized gain remains signed because it represents either a gain or a loss.
Dividend amount per unit is non-negative; cancelled/zero declarations may carry zero, but never a
negative amount.

## Pagination

`nextCursor`, `hasMore`, and validated `limit`. Cursor content is opaque and is not a domain ID.
Lists require deterministic ordering with a unique tie-breaker.

## Error

Stable code, safe message, and zero or more field details. Error DTOs carry no stack, SQL,
provider response, token, credential, passport, INN, or raw financial document content.

## Canonical naming rules

- JSON: lower camel case.
- Enum wire values: upper snake case.
- IDs: UUID strings, opaque to clients.
- Tickers: uppercase MOEX ticker, 1–32 alphanumeric characters.
- `createdAt`/`updatedAt`: SystemTimestamp only.
- financial/event dates: explicit BusinessDate names; never a generic `date` when semantics differ.
- `amount`: Decimal string within Money; never an unqualified numeric.
- absent is not zero; nullable is used only where the domain explicitly permits no value.

## Mapping policy

Public DTOs are anti-corruption boundaries. Future persistence models may use normalized tables,
internal surrogate keys, encrypted columns, and projection tables. Future Go types may separate
commands, entities, and responses. None may change the canonical meaning or wire representation
without contract review and an ADR when breaking.

## Assumptions resolved by this freeze

- MVP is web-first; refresh authentication uses an HttpOnly cookie and CSRF token.
- Public assets use ticker lookup; internal provider IDs remain private.
- Portfolio deletion means the portfolio is removed from active use while immutable financial
  history is retained.
- Transaction PATCH/DELETE are correction/reversal commands, not physical mutation.
- Dividend calculator is gross and excludes tax behavior.
- Dividend calendar contains official events only.
- All successful/error bodies use the canonical envelope.

## Open questions

None. New uncertainty must enter `OPEN_QUESTIONS.md` through Issue → ADR → review → approval;
it cannot be left as a production TODO or silently decided during backend implementation.
