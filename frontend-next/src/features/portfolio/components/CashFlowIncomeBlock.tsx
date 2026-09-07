import { formatMoney } from "@/common/presentation/format";
import type { ApiResult, PortfolioCashFlowProjection } from "@/common/api/openinvest";

type CashFlowIncomeBlockProps = {
  result: ApiResult<PortfolioCashFlowProjection> | null;
};

export function CashFlowIncomeBlock({ result }: CashFlowIncomeBlockProps) {
  if (result === null) {
    return (
      <section className="panel" aria-label="Cash Flow & Income">
        <p className="eyebrow">Ledger-derived cash truth</p>
        <h2>Cash Flow &amp; Income</h2>
        <p className="muted">Loading cash-flow aggregates from the canonical effective ledger…</p>
      </section>
    );
  }

  if (!result.ok) {
    return (
      <section className="panel warning" aria-label="Cash Flow & Income">
        <p className="eyebrow">Ledger-derived cash truth</p>
        <h2>Cash Flow &amp; Income unavailable</h2>
        <p>{result.message}</p>
      </section>
    );
  }

  const projection = result.data;
  const totals = projection.totals;

  return (
    <section className="panel" aria-label="Cash Flow & Income">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Ledger-derived cash truth</p>
          <h2>Cash Flow &amp; Income</h2>
        </div>
        <span className="muted">{projection.calculation.methodologyVersion}</span>
      </div>
      <p className="muted">
        Recorded RUB cash movements from the immutable effective ledger. These figures are not portfolio performance,
        market return, or tax advice.
      </p>

      <div className="metric-grid">
        <CashMetric label="Deposits" value={formatMoney(totals.deposits)} />
        <CashMetric label="Withdrawals" value={formatMoney(totals.withdrawals)} />
        <CashMetric label="Dividends" value={formatMoney(totals.dividendsGross)} />
        <CashMetric label="Coupons" value={formatMoney(totals.couponsGross)} />
        <CashMetric label="Fees" value={formatMoney(totals.fees)} />
        <CashMetric label="Taxes" value={formatMoney(totals.taxes)} />
        <CashMetric label="Net external flow" value={formatMoney(totals.netExternalFlow)} />
        <CashMetric label="Net investment income" value={formatMoney(totals.netInvestmentIncome)} />
        <CashMetric label="Net cash movement" value={formatMoney(totals.netCashFlow)} />
      </div>

      {projection.periods.length === 0 ? (
        <p className="muted">No effective ledger cash flows in this range.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Month</th>
                <th>Deposits</th>
                <th>Withdrawals</th>
                <th>Dividends</th>
                <th>Coupons</th>
                <th>Fees</th>
                <th>Taxes</th>
                <th>Net cash</th>
              </tr>
            </thead>
            <tbody>
              {projection.periods.map((period) => (
                <tr key={period.month}>
                  <td>{period.month}</td>
                  <td>{formatMoney(period.totals.deposits)}</td>
                  <td>{formatMoney(period.totals.withdrawals)}</td>
                  <td>{formatMoney(period.totals.dividendsGross)}</td>
                  <td>{formatMoney(period.totals.couponsGross)}</td>
                  <td>{formatMoney(period.totals.fees)}</td>
                  <td>{formatMoney(period.totals.taxes)}</td>
                  <td>{formatMoney(period.totals.netCashFlow)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function CashMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
