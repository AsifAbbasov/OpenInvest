"use client";

import { useEffect, useRef, useState } from "react";

import {
  clearManualValuation,
  upsertManualValuation,
  type PortfolioPositionProjection,
} from "@/common/api/openinvest";
import { formatMoney, formatRatioAsPercent } from "@/common/presentation/format";

type ManualValuationCellProps = {
  accessToken: string | null;
  portfolioId: string;
  item: PortfolioPositionProjection;
  viewMode: "current" | "historical";
  onChanged: () => Promise<void>;
};

export function ManualValuationCell({
  accessToken,
  portfolioId,
  item,
  viewMode,
  onChanged,
}: ManualValuationCellProps) {
  const [editing, setEditing] = useState(false);
  const [price, setPrice] = useState("");
  const [asOfDate, setAsOfDate] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const mutationGeneration = useRef(0);
  const available = item.marketValuation.status === "AVAILABLE";

  useEffect(() => {
    mutationGeneration.current += 1;
    setEditing(false);
    setIsSaving(false);
    setError(null);
    if (item.marketValuation.status === "AVAILABLE") {
      setPrice(item.marketValuation.marketPrice.amount);
      setAsOfDate(item.marketValuation.asOf);
    } else {
      setPrice("");
      setAsOfDate("");
    }
    return () => {
      mutationGeneration.current += 1;
    };
  }, [item.marketValuation, item.ticker, portfolioId, viewMode]);

  function beginEdit() {
    setError(null);
    if (item.marketValuation.status === "AVAILABLE") {
      setPrice(item.marketValuation.marketPrice.amount);
      setAsOfDate(item.marketValuation.asOf);
    } else {
      setPrice("");
      setAsOfDate("");
    }
    setEditing(true);
  }

  function cancelEdit() {
    mutationGeneration.current += 1;
    setEditing(false);
    setIsSaving(false);
    setError(null);
  }

  async function save() {
    if (!accessToken || price === "" || asOfDate === "" || isSaving) {
      return;
    }
    const generation = mutationGeneration.current + 1;
    mutationGeneration.current = generation;
    setIsSaving(true);
    setError(null);
    const result = await upsertManualValuation(
      portfolioId,
      item.ticker,
      { marketPrice: { amount: price, currency: "RUB" }, asOfDate },
      { accessToken },
    );
    if (mutationGeneration.current !== generation) {
      return;
    }
    setIsSaving(false);
    if (!result.ok) {
      setError(result.message);
      return;
    }
    setEditing(false);
    await onChanged();
  }

  async function clear() {
    if (!accessToken || isSaving) {
      return;
    }
    const generation = mutationGeneration.current + 1;
    mutationGeneration.current = generation;
    setIsSaving(true);
    setError(null);
    const result = await clearManualValuation(portfolioId, item.ticker, { accessToken });
    if (mutationGeneration.current !== generation) {
      return;
    }
    setIsSaving(false);
    if (!result.ok) {
      setError(result.message);
      return;
    }
    setEditing(false);
    await onChanged();
  }

  if (viewMode === "historical") {
    return available ? <ValuationReadout item={item} /> : <span className="market-unavailable">Market valuation unavailable</span>;
  }

  if (editing) {
    return (
      <div>
        <label>
          Manual price (RUB)
          <input
            inputMode="decimal"
            value={price}
            onChange={(event) => setPrice(event.target.value)}
            placeholder="318.50000000"
          />
        </label>
        <label>
          Price date
          <input type="date" value={asOfDate} onChange={(event) => setAsOfDate(event.target.value)} />
        </label>
        <div>
          <button type="button" disabled={!accessToken || price === "" || asOfDate === "" || isSaving} onClick={() => void save()}>
            {isSaving ? "Saving…" : "Save manual price"}
          </button>
          <button type="button" className="secondary-button" disabled={isSaving} onClick={cancelEdit}>
            Cancel
          </button>
        </div>
        {error ? <p className="warning-text" role="status">{error}</p> : null}
      </div>
    );
  }

  return (
    <div>
      {available ? <ValuationReadout item={item} /> : <span className="market-unavailable">Market valuation unavailable</span>}
      <div>
        <button type="button" className="secondary-button" disabled={!accessToken || isSaving} onClick={beginEdit}>
          {available ? "Edit manual price" : "Add manual price"}
        </button>
        {available ? (
          <button type="button" className="secondary-button" disabled={!accessToken || isSaving} onClick={() => void clear()}>
            {isSaving ? "Clearing…" : "Clear manual price"}
          </button>
        ) : null}
      </div>
      {error ? <p className="warning-text" role="status">{error}</p> : null}
    </div>
  );
}

function ValuationReadout({ item }: { item: PortfolioPositionProjection }) {
  if (item.marketValuation.status !== "AVAILABLE") {
    return null;
  }
  const valuation = item.marketValuation;
  return (
    <div>
      <small>Source: Manual · Price as of {valuation.asOf}</small>
      <div>Manual price: <strong>{formatMoney(valuation.marketPrice)}</strong></div>
      <div>Market value: <strong>{formatMoney(valuation.marketValue)}</strong></div>
      <div>Unrealized P/L: <strong>{formatMoney(valuation.unrealizedGain)}</strong></div>
      <div>Unrealized return: <strong>{formatRatioAsPercent(valuation.unrealizedReturn)}</strong></div>
      <div>Valued-position allocation: <strong>{formatRatioAsPercent(valuation.marketWeight)}</strong></div>
    </div>
  );
}
