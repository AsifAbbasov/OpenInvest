import type { ApiResult, PortfolioPositionsProjection } from "@/common/api/openinvest";
import { formatMoney, formatRatioAsPercent } from "@/common/presentation/format";

type ManualValuationAllocationProps = {
  result: ApiResult<PortfolioPositionsProjection> | null;
  viewMode: "current" | "historical";
};

const validMarketWeight = /^(?:0(?:\.\d{1,8})?|1(?:\.0{1,8})?)$/;

// This is a visual representation of server-provided USER_SUPPLIED values.
// It intentionally performs no market-value, denominator, or weight arithmetic.
export function ManualValuationAllocation({ result, viewMode }: ManualValuationAllocationProps) {
  if (viewMode !== "current" || !result?.ok || result.data.items.length === 0) {
    // PositionsBlock owns loading, error, historical and empty-portfolio states.
    return null;
  }

  const { items, valuationSummary } = result.data;
  const valuedItems = items.filter(
    (item) => item.marketValuation.status === "AVAILABLE" && item.marketValuation.source === "USER_SUPPLIED",
  );
  const invalidCoverage =
    valuationSummary.valuedPositions !== valuedItems.length ||
    valuationSummary.totalOpenPositions !== items.length ||
    (valuationSummary.status === "COMPLETE" && valuationSummary.valuedPositions !== valuationSummary.totalOpenPositions) ||
    (valuationSummary.status === "PARTIAL" && valuationSummary.valuedPositions === valuationSummary.totalOpenPositions);
  const zeroValuedTotal = /^0+(?:\.0+)?$/.test(valuationSummary.valuedPositionsMarketValue.amount);
  const invalidWeights = valuedItems.some(
    (item) => item.marketValuation.status === "AVAILABLE" &&
      (item.marketValuation.marketWeight === null || !validMarketWeight.test(item.marketValuation.marketWeight)),
  );

  return (
    <section className="acquisition-allocation manual-valuation-allocation" aria-label="Manual-valued position allocation">
      <h3>
        {valuationSummary.status === "PARTIAL"
          ? "Manual valuation allocation — among valued positions only"
          : "Manual valuation allocation — valued securities"}
      </h3>
      <p className="muted">
        USER_SUPPLIED manual prices, not live, broker, or provider-confirmed quotes. Each security has its own price date;
        dates may differ. Cash and positions without a manual price are excluded from these weights.
      </p>
      <p className="acquisition-allocation-total">
        Coverage: <strong>{valuationSummary.valuedPositions} / {valuationSummary.totalOpenPositions}</strong> open positions valued
        {` · ${valuationSummary.status}`}
      </p>
      <p className="acquisition-allocation-total">
        Denominator — manually valued securities only, excluding cash: <strong>{formatMoney(valuationSummary.valuedPositionsMarketValue)}</strong>
      </p>

      {invalidCoverage ? (
        <p className="muted" role="status">Allocation unavailable: server valuation coverage is inconsistent.</p>
      ) : valuedItems.length === 0 ? (
        <p className="muted" role="status">Allocation unavailable: no open positions have eligible manual prices.</p>
      ) : zeroValuedTotal ? (
        <p className="muted" role="status">Allocation unavailable: total manually valued securities amount is zero.</p>
      ) : invalidWeights ? (
        <p className="muted" role="status">Allocation unavailable: one or more server-provided market weights are missing or invalid.</p>
      ) : (
        <ul className="acquisition-allocation-list">
          {valuedItems.map((item) => {
            const valuation = item.marketValuation;
            if (valuation.status !== "AVAILABLE") {
              return null;
            }
            // Presentation scaling only: the backend already calculated marketWeight.
            const percentage = formatRatioAsPercent(valuation.marketWeight);
            return (
              <li className="acquisition-allocation-item" key={item.ticker}>
                <div className="acquisition-allocation-labels">
                  <strong className="acquisition-allocation-ticker">{item.ticker}</strong>
                  <span className="muted">{item.assetType}</span>
                  <span className="acquisition-allocation-values">{formatMoney(valuation.marketValue)} · {percentage}</span>
                </div>
                <small className="manual-valuation-allocation-source">
                  Source: USER_SUPPLIED · Price as of <time dateTime={valuation.asOf}>{valuation.asOf}</time>
                </small>
                <div className="acquisition-allocation-track" aria-hidden="true">
                  <span
                    className={`acquisition-allocation-fill acquisition-allocation-${item.assetType.toLowerCase()}`}
                    style={{ width: percentage }}
                  />
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
