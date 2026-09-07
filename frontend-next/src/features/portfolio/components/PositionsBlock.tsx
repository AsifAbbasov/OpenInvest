import type { ApiResult, PortfolioPositionsProjection } from "@/common/api/openinvest";
import { formatMoney, formatQuantityForDisplay, formatRatioAsPercent } from "@/common/presentation/format";
import { ManualValuationCell } from "@/features/portfolio/components/ManualValuationCell";

type PositionsBlockProps = {
  result: ApiResult<PortfolioPositionsProjection> | null;
  viewMode: "current" | "historical";
  historicalDate: string;
  accessToken: string | null;
  portfolioId: string;
  onShowCurrent: () => void;
  onShowHistorical: () => void;
  onHistoricalDateChange: (value: string) => void;
  onValuationChanged: () => Promise<void>;
};

export function PositionsBlock({
  result,
  viewMode,
  historicalDate,
  accessToken,
  portfolioId,
  onShowCurrent,
  onShowHistorical,
  onHistoricalDateChange,
  onValuationChanged,
}: PositionsBlockProps) {
  const historicalSelectionPending = viewMode === "historical" && historicalDate === "";
  const heading = viewMode === "historical" && historicalDate !== "" ? `Portfolio on ${historicalDate}` : "Positions";

  return (
    <section className="panel" aria-label="Portfolio positions">
      <div className="section-heading positions-heading">
        <div>
          <p className="eyebrow">{viewMode === "historical" ? "Portfolio time machine" : "Cost basis and manual valuation"}</p>
          <h2>{heading}</h2>
        </div>
        {result?.ok ? (
          <div className="positions-meta">
            <span>Total acquisition basis</span>
            <strong>{formatMoney(result.data.totalAcquisitionBasis)}</strong>
            {result.data.calculation.inputsAsOf ? <small>Inputs through {result.data.calculation.inputsAsOf}</small> : null}
          </div>
        ) : null}
      </div>

      <div className="segmented-control" aria-label="Portfolio position date mode">
        <button type="button" className={viewMode === "current" ? "active" : undefined} onClick={onShowCurrent}>
          Current
        </button>
        <button type="button" className={viewMode === "historical" ? "active" : undefined} onClick={onShowHistorical}>
          Historical
        </button>
      </div>

      {viewMode === "historical" ? (
        <label>
          Historical date
          <input
            type="date"
            value={historicalDate}
            onChange={(event) => onHistoricalDateChange(event.target.value)}
            aria-describedby="portfolio-time-machine-status"
          />
        </label>
      ) : null}

      <p id="portfolio-time-machine-status" className="muted">
        {viewMode === "current"
          ? "Viewing the latest accepted ledger projection. Current is not a wall-clock market valuation. Manual prices are explicit user-supplied valuation inputs, not live quotes."
          : historicalDate === ""
            ? "Choose a historical BusinessDate to reconstruct positions from the immutable ledger."
            : `Viewing portfolio as of ${historicalDate}. Manual valuation is shown only for an exact matching price date and position lifecycle.`}
      </p>

      {result?.ok && viewMode === "current" ? (
        <div className="panel" aria-label="Manual valuation coverage">
          <p className="eyebrow">Manual valuation coverage</p>
          <strong>
            {result.data.valuationSummary.valuedPositions} / {result.data.valuationSummary.totalOpenPositions} positions valued · {result.data.valuationSummary.status}
          </strong>
          <p>Cash: {formatMoney(result.data.valuationSummary.cashValue)}</p>
          <p>
            Current portfolio value: {result.data.valuationSummary.currentPortfolioValue
              ? formatMoney(result.data.valuationSummary.currentPortfolioValue)
              : "Unavailable until every open position has a manual valuation"}
          </p>
        </div>
      ) : null}

      {result === null && !historicalSelectionPending ? (
        <p className="muted">
          {viewMode === "historical"
            ? `Loading portfolio as of ${historicalDate} from the canonical ledger projection…`
            : "Loading positions from the canonical ledger projection…"}
        </p>
      ) : null}

      {result?.ok === false ? (
        <div className="warning position-warning" role="status">
          <strong>Positions unavailable</strong>
          <p>{result.message}</p>
          <p>No values are inferred from portfolio summary or transaction rows.</p>
        </div>
      ) : null}

      {result?.ok && result.data.items.length === 0 ? (
        <p className="muted">
          {viewMode === "historical" && historicalDate !== ""
            ? `No open stock or bond positions on ${historicalDate}.`
            : "No open stock or bond positions."}
        </p>
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
                      {viewMode === "historical" && item.marketValuation.status === "UNAVAILABLE" ? (
                        <span className="market-unavailable">Market valuation unavailable</span>
                      ) : (
                        <ManualValuationCell
                          accessToken={accessToken}
                          portfolioId={portfolioId}
                          item={item}
                          viewMode={viewMode}
                          onChanged={onValuationChanged}
                        />
                      )}
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
                {viewMode === "historical" && item.marketValuation.status === "UNAVAILABLE" ? (
                  <span className="market-unavailable">Market valuation unavailable</span>
                ) : (
                  <ManualValuationCell
                    accessToken={accessToken}
                    portfolioId={portfolioId}
                    item={item}
                    viewMode={viewMode}
                    onChanged={onValuationChanged}
                  />
                )}
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
