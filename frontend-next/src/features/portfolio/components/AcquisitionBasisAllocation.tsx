import type { ApiResult, PortfolioPositionsProjection } from "@/common/api/openinvest";
import { formatMoney, formatRatioAsPercent } from "@/common/presentation/format";

type AcquisitionBasisAllocationProps = {
  result: ApiResult<PortfolioPositionsProjection> | null;
  viewMode: "current" | "historical";
};

// Strictly a presentation of the positions projection. The Go API owns both
// the position bases and their proportional weights; no browser sums or ratios.
export function AcquisitionBasisAllocation({ result, viewMode }: AcquisitionBasisAllocationProps) {
  if (viewMode !== "current" || !result?.ok || result.data.items.length === 0) {
    // PositionsBlock already renders loading, error and empty states.
    return null;
  }

  const { items, totalAcquisitionBasis } = result.data;
  const zeroTotal = /^0+(?:\.0+)?$/.test(totalAcquisitionBasis.amount);
  const weightsUnavailable = items.some(
    (item) => item.acquisitionBasisWeight === null || !/^(?:0(?:\.\d{1,8})?|1(?:\.0{1,8})?)$/.test(item.acquisitionBasisWeight),
  );

  return (
    <section className="acquisition-allocation" aria-label="Acquisition-basis allocation">
      <h3>Acquisition-basis allocation</h3>
      <p className="muted">
        Share of remaining acquisition cost for open stocks and bonds. Cash is excluded. These are not market-value weights.
      </p>
      <p className="acquisition-allocation-total">
        Total position acquisition basis: <strong>{formatMoney(totalAcquisitionBasis)}</strong>
      </p>

      {zeroTotal ? (
        <p className="muted" role="status">
          Allocation unavailable: total acquisition basis is zero, so the position weights are undefined.
        </p>
      ) : weightsUnavailable ? (
        <p className="muted" role="status">
          Allocation unavailable: one or more server-provided position weights are missing or invalid.
        </p>
      ) : (
        <ul className="acquisition-allocation-list">
          {items.map((item) => {
            // The existing formatter changes the ratio's display scale only.
            const percent = formatRatioAsPercent(item.acquisitionBasisWeight);
            return (
              <li className="acquisition-allocation-item" key={item.ticker}>
                <div className="acquisition-allocation-labels">
                  <strong className="acquisition-allocation-ticker">{item.ticker}</strong>
                  <span className="muted">{item.assetType}</span>
                  <span className="acquisition-allocation-values">{formatMoney(item.acquisitionBasis)} · {percent}</span>
                </div>
                <div className="acquisition-allocation-track" aria-hidden="true">
                  <span
                    className={`acquisition-allocation-fill acquisition-allocation-${item.assetType.toLowerCase()}`}
                    style={{ width: percent }}
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
