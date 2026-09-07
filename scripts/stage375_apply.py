from pathlib import Path
import re


def read(path):
    return Path(path).read_text()


def write(path, text):
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    Path(path).write_text(text)


def replace_once(path, old, new, label):
    text = read(path)
    if old not in text:
        if new in text:
            return
        raise SystemExit(f"missing anchor: {label}")
    write(path, text.replace(old, new, 1))


# Manual append now accepts the already-frozen income/expense transaction shapes.
path = "backend-go/internal/verticalslice/service.go"
text = read(path)
old = '''\tcase "DEPOSIT", "WITHDRAWAL":
\t\tif request.Ticker != nil || request.Quantity != nil || request.UnitPrice != nil {
\t\t\treturn fmt.Errorf("%w: cash flows must not include ticker, quantity, or unitPrice", ErrInvalidInput)
\t\t}
\t\tif !request.Commission.Amount.IsZero() || !request.Tax.Amount.IsZero() {
\t\t\treturn fmt.Errorf("%w: cash flow commission and tax are unsupported and must be zero", ErrInvalidInput)
\t\t}
\tdefault:
\t\treturn fmt.Errorf("%w: transactionType is outside Stage 3.2 scope", ErrInvalidInput)
'''
new = '''\tcase "DIVIDEND", "COUPON":
\t\tif request.Ticker == nil || !tickerPattern.MatchString(*request.Ticker) {
\t\t\treturn fmt.Errorf("%w: ticker is required for income transactions", ErrInvalidInput)
\t\t}
\t\tif request.UnitPrice != nil {
\t\t\treturn fmt.Errorf("%w: income transactions must not include unitPrice", ErrInvalidInput)
\t\t}
\t\tif request.Quantity != nil && (!request.Quantity.IsPositive() || !request.Quantity.FitsStorage()) {
\t\t\treturn fmt.Errorf("%w: income quantity must be positive and storage-safe when supplied", ErrInvalidInput)
\t\t}
\tcase "FEE", "TAX":
\t\tif request.Quantity != nil || request.UnitPrice != nil {
\t\t\treturn fmt.Errorf("%w: expense transactions must not include quantity or unitPrice", ErrInvalidInput)
\t\t}
\t\tif request.Ticker != nil && !tickerPattern.MatchString(*request.Ticker) {
\t\t\treturn fmt.Errorf("%w: expense ticker is invalid", ErrInvalidInput)
\t\t}
\tcase "DEPOSIT", "WITHDRAWAL":
\t\tif request.Ticker != nil || request.Quantity != nil || request.UnitPrice != nil {
\t\t\treturn fmt.Errorf("%w: cash flows must not include ticker, quantity, or unitPrice", ErrInvalidInput)
\t\t}
\t\tif !request.Commission.Amount.IsZero() || !request.Tax.Amount.IsZero() {
\t\t\treturn fmt.Errorf("%w: cash flow commission and tax are unsupported and must be zero", ErrInvalidInput)
\t\t}
\tdefault:
\t\treturn fmt.Errorf("%w: transactionType is invalid", ErrInvalidInput)
'''
if old not in text:
    raise SystemExit("missing append transaction switch anchor")
write(path, text.replace(old, new, 1))

# Stage 3.71 adapter must expose the new optional read-model interface.
path = "backend-go/internal/verticalslice/stage_03_71_store_adapter.go"
text = read(path)
anchor = '''func (adapter *stage371StoreAdapter) GetPortfolioPositions(
\tctx context.Context,
\tsubjectID string,
\tportfolioID string,
\tasOfDate string,
) (PortfolioPositionsProjection, error) {
\tpositionsStore, ok := adapter.Store.(PortfolioPositionsStore)
\tif !ok {
\t\treturn PortfolioPositionsProjection{}, ErrPositionProjectionUnavailable
\t}
\treturn positionsStore.GetPortfolioPositions(ctx, subjectID, portfolioID, asOfDate)
}
'''
addition = anchor + '''
func (adapter *stage371StoreAdapter) GetPortfolioCashFlow(
\tctx context.Context,
\tsubjectID string,
\tportfolioID string,
\tfromDate string,
\ttoDate string,
) (PortfolioCashFlowProjection, error) {
\tcashFlowStore, ok := adapter.Store.(PortfolioCashFlowStore)
\tif !ok {
\t\treturn PortfolioCashFlowProjection{}, ErrCashFlowProjectionUnavailable
\t}
\treturn cashFlowStore.GetPortfolioCashFlow(ctx, subjectID, portfolioID, fromDate, toDate)
}
'''
if "func (adapter *stage371StoreAdapter) GetPortfolioCashFlow(" not in text:
    if anchor not in text:
        raise SystemExit("missing stage371 adapter positions anchor")
    write(path, text.replace(anchor, addition, 1))

# Snapshot cash is now the same canonical cash-flow arithmetic used by the read model.
path = "backend-go/internal/postgres/effective_ledger.go"
text = read(path)
pattern = re.compile(r'func effectiveSnapshotCashTx\(ctx context\.Context, tx \*sql\.Tx, portfolioID string, snapshotDate string\) \(decimal\.Decimal, decimal\.Decimal, string, error\) \{.*?\n\}', re.S)
replacement = '''func effectiveSnapshotCashTx(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string) (decimal.Decimal, decimal.Decimal, string, error) {
\trows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, snapshotDate)
\tif err != nil {
\t\treturn decimal.Zero(), decimal.Zero(), "", err
\t}
\tamounts := zeroCashFlowAmounts()
\tinvested := decimal.Zero()
\tfor _, row := range rows {
\t\tif err := amounts.apply(row); err != nil {
\t\t\treturn decimal.Zero(), decimal.Zero(), "", err
\t\t}
\t\tif row.TransactionType == "BUY" {
\t\t\toutflow := row.GrossAmount.Add(row.Commission).Add(row.Tax)
\t\t\tinvested = invested.Add(outflow)
\t\t\tif !invested.FitsStorage() {
\t\t\t\treturn decimal.Zero(), decimal.Zero(), "", fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8)", verticalslice.ErrInvalidInput)
\t\t\t}
\t\t}
\t}
\tcash := amounts.netCashFlow()
\tif !cash.FitsStorage() {
\t\treturn decimal.Zero(), decimal.Zero(), "", fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8)", verticalslice.ErrInvalidInput)
\t}
\tvar watermark string
\tif err := tx.QueryRowContext(ctx, `
        SELECT COALESCE(MAX(created_at)::text, 'empty')
        FROM investment.transaction_entries
        WHERE portfolio_id = $1 AND trade_date <= $2::date
    `, portfolioID, snapshotDate).Scan(&watermark); err != nil {
\t\treturn decimal.Zero(), decimal.Zero(), "", err
\t}
\treturn cash, invested, watermark, nil
}'''
new_text, count = pattern.subn(replacement, text, count=1)
if count != 1:
    raise SystemExit(f"effectiveSnapshotCashTx replacement count={count}")
write(path, new_text)

# Summary reads snapshot + cash-flow truth in one repeatable-read transaction.
path = "backend-go/internal/postgres/stage_03_71_summary.go"
text = read(path)
old_method = '''func (s *Store) GetPortfolioSummaryStage371(
\tctx context.Context,
\tsubjectID string,
\tportfolioID string,
\tasOfDate string,
) (verticalslice.PortfolioSummary, error) {
\tif _, err := s.GetPortfolio(ctx, subjectID, portfolioID); err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\treturn getPortfolioSummaryStage371(ctx, s.db, portfolioID, asOfDate)
}
'''
new_method = '''func (s *Store) GetPortfolioSummaryStage371(
\tctx context.Context,
\tsubjectID string,
\tportfolioID string,
\tasOfDate string,
) (verticalslice.PortfolioSummary, error) {
\ttx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
\tif err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\tdefer rollback(tx)
\tif _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\tsummary, err := getPortfolioSummaryStage371(ctx, tx, portfolioID, asOfDate)
\tif err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\tcashFlow, err := portfolioCashFlowProjectionTx(ctx, tx, portfolioID, "", summary.AsOfDate)
\tif err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\tsummary.DividendsReceived = cashFlow.Totals.DividendsGross
\tsummary.CouponsReceived = cashFlow.Totals.CouponsGross
\tif err := tx.Commit(); err != nil {
\t\treturn verticalslice.PortfolioSummary{}, err
\t}
\treturn summary, nil
}
'''
if old_method not in text:
    raise SystemExit("missing Stage 3.71 summary method anchor")
text = text.replace(old_method, new_method, 1)
text = text.replace("func getPortfolioSummaryStage371(\n\tctx context.Context,\n\tdb *sql.DB,", "func getPortfolioSummaryStage371(\n\tctx context.Context,\n\tdb queryer,", 1)
write(path, text)

# Wire the additive read endpoint in both app constructors.
for path in ["backend-go/internal/httpapi/routes.go", "backend-go/internal/httpapi/replay_app.go"]:
    text = read(path)
    anchor = '\tapp.Get("/api/v1/portfolios/:portfolioId/positions", api.getPortfolioPositions)\n'
    addition = anchor + '\tapp.Get("/api/v1/portfolios/:portfolioId/cash-flow", api.getPortfolioCashFlow)\n'
    if 'cash-flow", api.getPortfolioCashFlow' not in text:
        if anchor not in text:
            raise SystemExit(f"missing positions route anchor in {path}")
        text = text.replace(anchor, addition, 1)
        write(path, text)

# Typed frontend API read model.
path = "frontend-next/src/common/api/openinvest.ts"
text = read(path)
type_anchor = '''export type PortfolioPositionsRequestOptions = {
  asOfDate?: string;
  signal?: AbortSignal;
};
'''
type_addition = type_anchor + '''
export type PortfolioCashFlowTotals = {
  deposits: Money;
  withdrawals: Money;
  buyOutflows: Money;
  sellInflows: Money;
  dividendsGross: Money;
  couponsGross: Money;
  fees: Money;
  taxes: Money;
  netExternalFlow: Money;
  netInvestmentIncome: Money;
  netCashFlow: Money;
};

export type PortfolioCashFlowPeriod = {
  month: string;
  totals: PortfolioCashFlowTotals;
};

export type PortfolioCashFlowProjection = {
  portfolioId: string;
  fromDate: string | null;
  toDate: string | null;
  totals: PortfolioCashFlowTotals;
  periods: PortfolioCashFlowPeriod[];
  calculation: {
    methodologyVersion: "portfolio-cash-flow-income-v1";
    inputsAsOf: string | null;
  };
};

export type PortfolioCashFlowRequestOptions = {
  fromDate?: string;
  toDate?: string;
  signal?: AbortSignal;
};
'''
if "export type PortfolioCashFlowProjection" not in text:
    if type_anchor not in text:
        raise SystemExit("missing frontend positions request type anchor")
    text = text.replace(type_anchor, type_addition, 1)
function_anchor = '''export async function listTransactions(
  portfolioId: string,
  auth: AuthenticatedRequest,
  page: ListPageRequest = {},
): Promise<ApiResult<ListData<Transaction>>> {
'''
function_addition = '''export async function getPortfolioCashFlow(
  portfolioId: string,
  auth: AuthenticatedRequest,
  options: PortfolioCashFlowRequestOptions = {},
): Promise<ApiResult<PortfolioCashFlowProjection>> {
  const searchParams = new URLSearchParams();
  if (options.fromDate) searchParams.set("fromDate", options.fromDate);
  if (options.toDate) searchParams.set("toDate", options.toDate);
  const query = searchParams.toString();
  return request<PortfolioCashFlowProjection>(
    `/api/v1/portfolios/${encodeURIComponent(portfolioId)}/cash-flow${query === "" ? "" : `?${query}`}`,
    { headers: bearerHeaders(auth.accessToken), signal: options.signal },
  );
}

''' + function_anchor
if "export async function getPortfolioCashFlow(" not in text:
    if function_anchor not in text:
        raise SystemExit("missing frontend listTransactions anchor")
    text = text.replace(function_anchor, function_addition, 1)
write(path, text)

# Existing transaction form exposes all canonical types and exact shape/null semantics.
path = "frontend-next/src/features/portfolio/components/AddTransactionForm.tsx"
text = read(path)
text = text.replace(
    'const transactionTypes: TransactionType[] = ["BUY", "SELL", "DEPOSIT", "WITHDRAWAL"];',
    'const transactionTypes: TransactionType[] = ["BUY", "SELL", "DIVIDEND", "COUPON", "FEE", "TAX", "DEPOSIT", "WITHDRAWAL"];',
    1,
)
text = text.replace(
    '  const isAssetIncome = transactionType === "DIVIDEND" || transactionType === "COUPON";\n  const isCashFlow = transactionType === "DEPOSIT" || transactionType === "WITHDRAWAL";',
    '  const isAssetIncome = transactionType === "DIVIDEND" || transactionType === "COUPON";\n  const isExpense = transactionType === "FEE" || transactionType === "TAX";\n  const isCashFlow = transactionType === "DEPOSIT" || transactionType === "WITHDRAWAL";',
    1,
)
# Replace build-payload shape block by semantic role.
old = '''      ticker: isCashFlow ? null : ticker.trim().toUpperCase(),
      quantity: isTrade || isAssetIncome ? quantity.trim() : null,
      unitPrice: isTrade ? { amount: unitPrice.trim(), currency: "RUB" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount.trim(), currency: "RUB" },
      commission: { amount: commission.trim(), currency: "RUB" },
      tax: { amount: tax.trim(), currency: "RUB" },
'''
new = '''      ticker: isTrade || isAssetIncome ? ticker.trim().toUpperCase() : null,
      quantity: isTrade ? quantity.trim() : isAssetIncome && quantity.trim() !== "" ? quantity.trim() : null,
      unitPrice: isTrade ? { amount: unitPrice.trim(), currency: "RUB" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount.trim(), currency: "RUB" },
      commission: { amount: isExpense || isCashFlow ? "0.00000000" : commission.trim(), currency: "RUB" },
      tax: { amount: isExpense || isCashFlow ? "0.00000000" : tax.trim(), currency: "RUB" },
'''
if old not in text:
    raise SystemExit("missing AddTransactionForm payload anchor")
text = text.replace(old, new, 1)
# Replace ticker/quantity field visibility and required semantics.
text = text.replace('{!isCashFlow ? (\n        <label>', '{isTrade || isAssetIncome ? (\n        <label>', 1)
text = text.replace('            required\n            value={quantity}', '            required={isTrade}\n            value={quantity}', 1)
# Hide nested commission/tax for pure expense/cash rows and let frozen zeros be submitted.
commission_block = '''      <label>
        Commission
        <input
          inputMode="decimal"
          maxLength={29}
          placeholder="0.00000000"
          required
          value={commission}
          onChange={(event) => setCommission(event.target.value)}
        />
      </label>
      <label>
        Tax
        <input
          inputMode="decimal"
          maxLength={29}
          placeholder="0.00000000"
          required
          value={tax}
          onChange={(event) => setTax(event.target.value)}
        />
      </label>
'''
commission_new = '''      {!isExpense && !isCashFlow ? (
        <>
          <label>
            Commission
            <input
              inputMode="decimal"
              maxLength={29}
              placeholder="0.00000000"
              required
              value={commission}
              onChange={(event) => setCommission(event.target.value)}
            />
          </label>
          <label>
            Tax
            <input
              inputMode="decimal"
              maxLength={29}
              placeholder="0.00000000"
              required
              value={tax}
              onChange={(event) => setTax(event.target.value)}
            />
          </label>
        </>
      ) : null}
'''
if commission_block not in text:
    raise SystemExit("missing commission/tax UI anchor")
text = text.replace(commission_block, commission_new, 1)
text = text.replace('Append immutable transaction', 'Append immutable transaction / income / expense', 1)
write(path, text)

# Portfolio page loads cash-flow projection with the same guarded current load and refreshes it after every ledger mutation.
path = "frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx"
text = read(path)
text = text.replace('  getPortfolio,\n  getPortfolioPositions,', '  getPortfolio,\n  getPortfolioCashFlow,\n  getPortfolioPositions,', 1)
text = text.replace('  type Portfolio,\n  type PortfolioPositionsProjection,', '  type Portfolio,\n  type PortfolioCashFlowProjection,\n  type PortfolioPositionsProjection,', 1)
text = text.replace('import { AddTransactionForm } from "@/features/portfolio/components/AddTransactionForm";', 'import { AddTransactionForm } from "@/features/portfolio/components/AddTransactionForm";\nimport { CashFlowIncomeBlock } from "@/features/portfolio/components/CashFlowIncomeBlock";', 1)
text = text.replace('  positions: ApiResult<PortfolioPositionsProjection>;\n  transactions:', '  positions: ApiResult<PortfolioPositionsProjection>;\n  cashFlow: ApiResult<PortfolioCashFlowProjection>;\n  transactions:', 1)
old = '''    const [portfolio, summary, positions, transactions] = await Promise.all([
      getPortfolio(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioSummary(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioPositions(portfolioId, { accessToken: attempt.accessToken }),
      listTransactions(portfolioId, { accessToken: attempt.accessToken }),
    ]);'''
new = '''    const [portfolio, summary, positions, cashFlow, transactions] = await Promise.all([
      getPortfolio(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioSummary(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioPositions(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioCashFlow(portfolioId, { accessToken: attempt.accessToken }),
      listTransactions(portfolioId, { accessToken: attempt.accessToken }),
    ]);'''
if old not in text:
    raise SystemExit("missing PortfolioDetail Promise.all anchor")
text = text.replace(old, new, 1)
text = text.replace('      setState({ portfolio, summary, positions, transactions });', '      setState({ portfolio, summary, positions, cashFlow, transactions });', 1)
# Remove placeholder return cards; preserve only ledger/cost-basis facts.
old_metrics = '''          <Metric label="Invested capital" value={formatMoney(summary.investedCapital)} />
          <Metric label="Nominal return rate" value={formatNullableDecimal(summary.nominalReturnRate)} />
          <Metric label="XIRR" value={formatNullableDecimal(summary.xirr)} />
          <Metric label="Real gain" value={summary.realReturn ? formatMoney(summary.realReturn.realGain) : "Unavailable"} />
          <Metric label="Purchasing power basis" value={formatMoney(summary.purchasingPower.portfolioValue)} />'''
new_metrics = '''          <Metric label="Invested capital" value={formatMoney(summary.investedCapital)} />
          <Metric label="Dividends received" value={formatMoney(summary.dividendsReceived)} />
          <Metric label="Coupons received" value={formatMoney(summary.couponsReceived)} />'''
if old_metrics not in text:
    raise SystemExit("missing legacy summary metric block")
text = text.replace(old_metrics, new_metrics, 1)
text = text.replace('import { formatMoney, formatNullableDecimal } from "@/common/presentation/format";', 'import { formatMoney } from "@/common/presentation/format";', 1)
positions_anchor = '''      <PositionsBlock
        result={visiblePositions}
        viewMode={positionViewMode}
        historicalDate={historicalDate}
        onShowCurrent={showCurrentPositions}
        onShowHistorical={showHistoricalPositions}
        onHistoricalDateChange={changeHistoricalDate}
      />
'''
positions_add = positions_anchor + '''
      {positionViewMode === "current" ? <CashFlowIncomeBlock result={state?.cashFlow ?? null} /> : null}
'''
if "<CashFlowIncomeBlock" not in text:
    if positions_anchor not in text:
        raise SystemExit("missing PositionsBlock render anchor")
    text = text.replace(positions_anchor, positions_add, 1)
write(path, text)

# OpenAPI additive endpoint.
path = "openapi/openapi.yaml"
text = read(path)
path_anchor = '''  /api/v1/portfolios/{portfolioId}/snapshots:
'''
endpoint = '''  /api/v1/portfolios/{portfolioId}/cash-flow:
    parameters:
      - $ref: "#/components/parameters/RequestId"
      - $ref: "#/components/parameters/Traceparent"
      - $ref: "#/components/parameters/PortfolioId"
    get:
      operationId: getPortfolioCashFlow
      summary: Get deterministic effective-ledger cash-flow and income projection
      description: >-
        Aggregates the authenticated portfolio's Stage 3.74 effective ledger by canonical tradeDate.
        Fees include standalone FEE gross amounts plus recorded transaction commission fields; taxes
        include standalone TAX gross amounts plus recorded transaction tax fields. The response is
        recorded cash-flow truth only and is not market performance, investment return, or tax advice.
      tags: [Portfolios]
      parameters:
        - $ref: "#/components/parameters/FromDate"
        - $ref: "#/components/parameters/ToDate"
      responses:
        "200": { $ref: "./components/responses.yaml#/responses/PortfolioCashFlow" }
        "400": { $ref: "./components/responses.yaml#/responses/BadRequest" }
        "401": { $ref: "./components/responses.yaml#/responses/Unauthorized" }
        "404": { $ref: "./components/responses.yaml#/responses/NotFound" }

'''
if "/api/v1/portfolios/{portfolioId}/cash-flow:" not in text:
    if path_anchor not in text:
        raise SystemExit("missing OpenAPI snapshots path anchor")
    text = text.replace(path_anchor, endpoint + path_anchor, 1)
schema_alias_anchor = '    PortfolioPositionsResponse: { $ref: "./components/schemas.yaml#/PortfolioPositionsResponse" }\n'
schema_aliases = schema_alias_anchor + '    PortfolioCashFlowTotals: { $ref: "./components/schemas.yaml#/PortfolioCashFlowTotals" }\n    PortfolioCashFlowPeriod: { $ref: "./components/schemas.yaml#/PortfolioCashFlowPeriod" }\n    PortfolioCashFlowCalculation: { $ref: "./components/schemas.yaml#/PortfolioCashFlowCalculation" }\n    PortfolioCashFlowProjection: { $ref: "./components/schemas.yaml#/PortfolioCashFlowProjection" }\n    PortfolioCashFlowResponse: { $ref: "./components/schemas.yaml#/PortfolioCashFlowResponse" }\n'
if "PortfolioCashFlowResponse:" not in text:
    if schema_alias_anchor not in text:
        raise SystemExit("missing OpenAPI positions schema alias anchor")
    text = text.replace(schema_alias_anchor, schema_aliases, 1)
write(path, text)

# Schemas for bounded backend aggregate response.
path = "openapi/components/schemas.yaml"
text = read(path)
anchor = '''PortfolioPositionsResponse:
  allOf:
    - $ref: "#/BaseResponse"
    - type: object
'''
if "PortfolioCashFlowTotals:" not in text:
    idx = text.find(anchor)
    if idx < 0:
        raise SystemExit("missing PortfolioPositionsResponse schema anchor")
    # Insert after the complete positions response block, immediately before PortfolioSnapshot.
    marker = "\nPortfolioSnapshot:\n"
    pos = text.find(marker, idx)
    if pos < 0:
        raise SystemExit("missing PortfolioSnapshot marker after positions response")
    block = '''
PortfolioCashFlowTotals:
  type: object
  additionalProperties: false
  required:
    [deposits, withdrawals, buyOutflows, sellInflows, dividendsGross, couponsGross, fees, taxes, netExternalFlow, netInvestmentIncome, netCashFlow]
  properties:
    deposits: { $ref: "#/Money" }
    withdrawals: { $ref: "#/Money" }
    buyOutflows: { $ref: "#/Money" }
    sellInflows: { $ref: "#/Money" }
    dividendsGross: { $ref: "#/Money" }
    couponsGross: { $ref: "#/Money" }
    fees: { $ref: "#/Money" }
    taxes: { $ref: "#/Money" }
    netExternalFlow: { $ref: "#/Money" }
    netInvestmentIncome: { $ref: "#/Money" }
    netCashFlow: { $ref: "#/Money" }

PortfolioCashFlowPeriod:
  type: object
  additionalProperties: false
  required: [month, totals]
  properties:
    month:
      type: string
      pattern: "^[0-9]{4}-(0[1-9]|1[0-2])$"
    totals: { $ref: "#/PortfolioCashFlowTotals" }

PortfolioCashFlowCalculation:
  type: object
  additionalProperties: false
  required: [methodologyVersion, inputsAsOf]
  properties:
    methodologyVersion:
      type: string
      const: portfolio-cash-flow-income-v1
    inputsAsOf:
      oneOf:
        - $ref: "#/BusinessDate"
        - type: "null"

PortfolioCashFlowProjection:
  type: object
  additionalProperties: false
  required: [portfolioId, fromDate, toDate, totals, periods, calculation]
  properties:
    portfolioId:
      type: string
      format: uuid
    fromDate:
      oneOf:
        - $ref: "#/BusinessDate"
        - type: "null"
    toDate:
      oneOf:
        - $ref: "#/BusinessDate"
        - type: "null"
    totals: { $ref: "#/PortfolioCashFlowTotals" }
    periods:
      type: array
      items: { $ref: "#/PortfolioCashFlowPeriod" }
    calculation: { $ref: "#/PortfolioCashFlowCalculation" }

PortfolioCashFlowResponse:
  allOf:
    - $ref: "#/BaseResponse"
    - type: object
      required: [data]
      properties:
        data: { $ref: "#/PortfolioCashFlowProjection" }

'''
    text = text[:pos] + "\n" + block + text[pos:]
write(path, text)

# Response registry.
path = "openapi/components/responses.yaml"
text = read(path)
anchor = '''  SnapshotList:
'''
block = '''  PortfolioCashFlow:
    description: Deterministic effective-ledger cash-flow and income aggregate.
    headers:
      X-Request-ID: { $ref: "#/headers/RequestId" }
      X-Trace-ID: { $ref: "#/headers/TraceId" }
    content:
      application/json:
        schema: { $ref: "./schemas.yaml#/PortfolioCashFlowResponse" }
        examples:
          default: { $ref: "../examples/cash-flow.json#/cashFlowResponse" }

'''
if "  PortfolioCashFlow:\n" not in text:
    if anchor not in text:
        raise SystemExit("missing responses SnapshotList anchor")
    text = text.replace(anchor, block + anchor, 1)
write(path, text)

# Strict OpenAPI validator registration.
path = "backend-go/cmd/validate-openapi/main.go"
text = read(path)
operation_anchor = '\t"GET /api/v1/portfolios/{portfolioId}/positions":                       "getPortfolioPositions",\n'
if 'cash-flow":' not in text:
    if operation_anchor not in text:
        raise SystemExit("missing validator positions operation anchor")
    text = text.replace(operation_anchor, operation_anchor + '\t"GET /api/v1/portfolios/{portfolioId}/cash-flow":                       "getPortfolioCashFlow",\n', 1)
schema_anchor = '\t"PortfolioPositionsCalculation", "PortfolioPositionsProjection", "PortfolioPositionsResponse", "ImportReviewResult", "ImportAppendResult",\n'
if '"PortfolioCashFlowTotals"' not in text:
    if schema_anchor not in text:
        raise SystemExit("missing validator required schema anchor")
    text = text.replace(schema_anchor, '\t"PortfolioPositionsCalculation", "PortfolioPositionsProjection", "PortfolioPositionsResponse",\n\t"PortfolioCashFlowTotals", "PortfolioCashFlowPeriod", "PortfolioCashFlowCalculation", "PortfolioCashFlowProjection", "PortfolioCashFlowResponse",\n\t"ImportReviewResult", "ImportAppendResult",\n', 1)
write(path, text)

# Frontend contract test: all types are manually available and cash-flow math remains backend-owned.
write("frontend-next/tests/stage-03-75-cash-flow-income.test.mjs", r'''import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const root = new URL("../", import.meta.url);
const read = (path) => readFile(new URL(path, root), "utf8");

test("Stage 3.75 exposes all canonical manual transaction types without frontend cash-flow arithmetic", async () => {
  const form = await read("src/features/portfolio/components/AddTransactionForm.tsx");
  for (const type of ["BUY", "SELL", "DIVIDEND", "COUPON", "FEE", "TAX", "DEPOSIT", "WITHDRAWAL"]) {
    assert.match(form, new RegExp(`\\"${type}\\"`));
  }
  assert.match(form, /isAssetIncome/);
  assert.match(form, /isExpense/);
  assert.match(form, /grossAmount:/);
  assert.match(form, /quantity: isTrade \?/);
  assert.doesNotMatch(form, /netInvestmentIncome\s*=/);
  assert.doesNotMatch(form, /netCashFlow\s*=/);
});

test("Stage 3.75 uses a typed bounded backend aggregate and renders server values only", async () => {
  const api = await read("src/common/api/openinvest.ts");
  const block = await read("src/features/portfolio/components/CashFlowIncomeBlock.tsx");
  const detail = await read("src/features/portfolio/components/PortfolioDetailSlice.tsx");

  assert.match(api, /getPortfolioCashFlow/);
  assert.match(api, /\/cash-flow/);
  assert.match(api, /PortfolioCashFlowProjection/);
  assert.match(block, /projection\.periods\.map/);
  assert.match(block, /formatMoney\(totals\.netInvestmentIncome\)/);
  assert.match(block, /not portfolio performance/);
  assert.doesNotMatch(block, /reduce\(/);
  assert.doesNotMatch(block, /parseFloat|Number\(/);
  assert.match(detail, /getPortfolioCashFlow/);
  assert.match(detail, /CashFlowIncomeBlock/);
  assert.doesNotMatch(detail, /Nominal return rate/);
  assert.doesNotMatch(detail, />XIRR</);
  assert.doesNotMatch(detail, /Real gain/);
});
''')

# Dossier and minimal active-registry synchronization. Merge activates canonical status later.
write("docs/stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md", '''# Stage 3.75 — Portfolio Cash Flow & Income Truth

| Field | Value |
| --- | --- |
| Status | IMPLEMENTATION CANDIDATE — canonical only after reviewed protected merge |
| Canonical base | `develop@7043871e44eeb25e516a40b059f89d3cb25e03c1` |
| Budget | `0 RUB` |
| Runtime providers | none |
| New database migration | none |

## Purpose

Stage 3.75 closes the financial gap between the immutable transaction ledger and user-visible cash/income truth. It reuses Stage 3.74 effective-ledger correction/reversal semantics and adds one backend-owned, rebuildable aggregate projection for deposits, withdrawals, trades, income, fees and recorded taxes.

## Canonical methodology

For each effective ledger row, transaction-level `commission` is counted once in `fees` and transaction-level `tax` is counted once in `taxes`. Then the row gross amount is categorized exactly once:

- `DEPOSIT` -> `deposits`
- `WITHDRAWAL` -> `withdrawals`
- `BUY` -> `buyOutflows`
- `SELL` -> `sellInflows`
- `DIVIDEND` -> `dividendsGross`
- `COUPON` -> `couponsGross`
- standalone `FEE` gross -> `fees`
- standalone `TAX` gross -> `taxes`

Derived identities are:

```text
netExternalFlow = deposits - withdrawals
netInvestmentIncome = dividendsGross + couponsGross - fees - taxes
netCashFlow = deposits - withdrawals - buyOutflows + sellInflows
              + dividendsGross + couponsGross - fees - taxes
```

`netInvestmentIncome` is recorded ledger income after all recorded fees/taxes in the selected range. It is not investment return, market performance, or tax advice.

## Date and correction semantics

- grouping uses canonical `tradeDate`, never `created_at`;
- periods are backend-produced `YYYY-MM` buckets in ascending order;
- `fromDate` and `toDate` are optional BusinessDate filters;
- explicit `toDate` uses Stage 3.74 historical effective-ledger semantics, including reversal `effectiveDate`;
- corrections replace earlier revisions economically instead of being added to them;
- reversals remove the target economic effect when effective under the requested cutoff;
- omitted `toDate` preserves Stage 3.74 current effective-ledger semantics.

## API

```http
GET /api/v1/portfolios/{portfolioId}/cash-flow?fromDate=YYYY-MM-DD&toDate=YYYY-MM-DD
```

The response is bounded aggregate data: totals plus monthly periods. The browser does not fetch the full ledger to calculate financial truth.

## Summary and snapshot integration

The same cash-flow arithmetic now drives snapshot cash value. Stage 3.71 invested-capital semantics remain unchanged: BUY gross + BUY commission + BUY tax. `PortfolioSummary.dividendsReceived` and `couponsReceived` are populated from effective-ledger truth instead of hardcoded zero placeholders. Summary reads snapshot and income truth under one repeatable-read transaction.

## Manual transaction entry

The existing Web form now exposes the already-frozen `DIVIDEND`, `COUPON`, `FEE`, and `TAX` types. No second form or new transaction type was created. Income follows the frozen ticker / optional quantity / gross amount contract. Pure standalone FEE/TAX UI writes their gross amount with null ticker/quantity/unitPrice and zero nested commission/tax; nested commission/tax remain available on trade/income rows.

## UI boundary

The portfolio page renders backend totals and monthly buckets only. Legacy nominal-return/XIRR/real-gain/purchasing-power cards are not presented as investment performance in this stage because approved market valuation and return methodology remain unavailable.

## Non-scope

No market prices/value, unrealized P/L, XIRR/return methodology, inflation, provider activation, broker sync, tax advice/declaration, bond YTM/duration/NKD, notification, AI, Redis, Kafka, worker, paid service, new transaction table, aggregate cache table, or production rollout is authorized.

## Verification

The implementation includes backend integration vectors for mixed ledgers, standalone and nested fees/taxes, monthly/range aggregation, summary cash/dividend/coupon truth, correction, reversal effective dates, backdated income, empty periods and subject isolation; HTTP contract witnesses; frontend manual-type and backend-owned-math contract tests; and the repository's full protected CI matrix.
''')

# Minimal active docs: candidate wording only until protected merge.
for path, anchor, line in [
    ("docs/ROADMAP.md", "| 3.74 — Transaction Correction & Reversal / Ledger Repair UX |", "| 3.75 — Portfolio Cash Flow & Income Truth | Reuse Stage 3.74 effective ledger for backend-owned cash/income totals, monthly aggregation, manual income/expense entry and truthful summary cash/income | Implementation candidate; canonical only after reviewed protected merge; zero-budget, no provider/migration/cache-table |\n"),
    ("docs/IMPLEMENTATION_LOG.md", "| 3.74 — Transaction Correction & Reversal / Ledger Repair UX |", "| 3.75 — Portfolio Cash Flow & Income Truth | Add effective-ledger cash/income projection, additive API, manual DIVIDEND/COUPON/FEE/TAX entry and truthful cash/dividend/coupon summary | Implementation candidate; protected merge required for canonical status | [Stage 3.75 report](stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md) |\n"),
]:
    text = read(path)
    if "3.75 — Portfolio Cash Flow & Income Truth" not in text:
        pos = text.find(anchor)
        if pos < 0:
            raise SystemExit(f"missing docs anchor in {path}")
        end = text.find("\n", pos)
        text = text[:end+1] + line + text[end+1:]
        write(path, text)

path = "docs/DOCUMENT_INDEX.md"
text = read(path)
if "Stage 3.75 Portfolio Cash Flow & Income Truth" not in text:
    marker = "| Stage 3.74 Transaction Correction & Reversal / Ledger Repair UX |"
    pos = text.find(marker)
    if pos < 0:
        raise SystemExit("missing document index Stage 3.74 anchor")
    end = text.find("\n", pos)
    line = "| Stage 3.75 Portfolio Cash Flow & Income Truth | Implementation candidate: effective-ledger cash/income projection, additive API, manual income/expense UX and summary truth | `stages/STAGE_03_75_PORTFOLIO_CASH_FLOW_INCOME_IMPLEMENTATION.md` |\n"
    text = text[:end+1] + line + text[end+1:]
    write(path, text)

path = "docs/SOURCE_OF_TRUTH.md"
text = read(path)
old = "**Current candidate frontier: none. No Stage 3.75+ product/runtime scope is authorized by the Stage 3.74 merge or this documentation closure.**"
new = "**Current candidate frontier: Stage 3.75 — Portfolio Cash Flow & Income Truth is the separately authorized implementation candidate on top of canonical Stage 3.74; it remains non-canonical until its implementation PR passes review/CI and is protected-merged. No Stage 3.76+ product/runtime scope is authorized.**"
if old not in text:
    if new not in text:
        raise SystemExit("missing Source of Truth candidate frontier anchor")
else:
    write(path, text.replace(old, new, 1))

path = "README.md"
text = read(path)
anchor = "Stage 3 is canonically implemented through **Stage 3.74 — Transaction Correction & Reversal / Ledger Repair UX**."
new = anchor + " Stage 3.75 Portfolio Cash Flow & Income Truth is the current reviewed implementation candidate and is not canonical until protected merge."
if "Stage 3.75 Portfolio Cash Flow & Income Truth is the current reviewed implementation candidate" not in text:
    if anchor not in text:
        raise SystemExit("missing README Stage 3 frontier anchor")
    write(path, text.replace(anchor, new, 1))
