import type { ApiResult, PortfolioPositionsProjection } from "@/common/api/openinvest";
import { formatMoney, formatQuantityForDisplay, formatRatioAsPercent } from "@/common/presentation/format";

type PositionsBlockProps = {
  result: ApiResult<PortfolioPositionsProjection> | null;
};

export function PositionsBlock({ result }: PositionsBlockProps) {
  return (
    <section className="panel" aria-label="Portfolio positions">
      <div className="section-heading positions-heading">
        <div>
          <p className="eyebrow">Cost basis view</p>
          <h2>Positions</h2>
        </div>
        {result?.ok ? (
          <div className="positions-meta">
            <span>Total acquisition basis</span>
            <strong>{formatMoney(result.data.totalAcquisitionBasis)}</strong>
            {result.data.calculation.inputsAsOf ? <small>Inputs through {result.data.calculation.inputsAsOf}</small> : null}
          </div>
        ) : null}
      </div>

      {result === null ? <p className="muted">Loading positions from the canonical ledger projection…</p> : null}

      {result?.ok === false ? (
        <div className="warning position-warning" role="status">
          <strong>Positions unavailable</strong>
          <p>{result.message}</p>
          <p>No values are inferred from portfolio summary or transaction rows.</p>
        </div>
      ) : null}

      {result?.ok && result.data.items.length === 0 ? (
        <p className="muted">No open stock or bond positions.</p>
      ) : null}

      {result?.ok && result.data.items.length > 0 ? (
        <>
          <div className="table-wrap positions-table">
            <table>
              <thead>
                <tr>
                  <th>Asset</th>
                  <th>Quantity</th>
                  <th>Average acquisition price</th>
                  <th>Acquisition basis</th>
                  <th>Acquisition-basis weight</th>
                  <th>Market valuation</th>
                </tr>
              </thead>
              <tbody>
                {result.data.items.map((item) => (
                  <tr key={item.ticker}>
                    <td>
                      <strong>{item.ticker}</strong>
                      <small className="position-asset-type">{item.assetType}</small>
                    </td>
                    <td>{formatQuantityForDisplay(item.quantity)}</td>
                    <td>{formatMoney(item.weightedAverageCost)}</td>
                    <td>{formatMoney(item.acquisitionBasis)}</td>
                    <td>{formatRatioAsPercent(item.acquisitionBasisWeight)}</td>
                    <td>
                      <span className="market-unavailable">Market valuation unavailable</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="position-cards" aria-label="Portfolio positions mobile view">
            {result.data.items.map((item) => (
              <article className="position-card" key={item.ticker}>
                <div className="position-card-title">
                  <strong>{item.ticker}</strong>
                  <span>{item.assetType}</span>
                </div>
                <PositionCardRow label="Quantity" value={formatQuantityForDisplay(item.quantity)} />
                <PositionCardRow label="Average acquisition price" value={formatMoney(item.weightedAverageCost)} />
                <PositionCardRow label="Acquisition basis" value={formatMoney(item.acquisitionBasis)} />
                <PositionCardRow label="Acquisition-basis weight" value={formatRatioAsPercent(item.acquisitionBasisWeight)} />
                <span className="market-unavailable">Market valuation unavailable</span>
              </article>
            ))}
          </div>
        </>
      ) : null}
    </section>
  );
}

function PositionCardRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="position-card-row">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
