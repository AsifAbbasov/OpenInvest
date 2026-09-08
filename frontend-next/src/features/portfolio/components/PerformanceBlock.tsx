"use client";

import type { ApiResult, PortfolioReturnProjection, PortfolioReturnUnavailableReason } from "@/common/api/openinvest";

type PerformanceBlockProps = {
  result: ApiResult<PortfolioReturnProjection> | null;
  asOfDate: string;
  onAsOfDateChange: (value: string) => void;
};

const unavailableReasonText: Record<PortfolioReturnUnavailableReason, string> = {
  INCOMPLETE_VALUATION: "Every open position needs an explicit manual valuation for this exact date.",
  NO_EXTERNAL_CONTRIBUTIONS: "No effective external contribution is available for this date range.",
  INSUFFICIENT_DATE_SPAN: "The effective cash-flow series does not span more than one BusinessDate.",
  NO_SIGN_CHANGE: "The constructed cash-flow series has no mathematically admissible XIRR root.",
  NON_POSITIVE_TERMINAL_VALUE: "The exact-date terminal portfolio value is not positive.",
  AMBIGUOUS_MULTIPLE_ROOTS: "The cash-flow series has more than one admissible XIRR root, so no single return is published.",
  NUMERICAL_SOLUTION_FAILED: "The backend could not publish a unique numerically reliable XIRR for this series.",
};

export function PerformanceBlock({ result, asOfDate, onAsOfDateChange }: PerformanceBlockProps) {
  return (
    <section className="panel" aria-label="Money-weighted return">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Performance</p>
          <h2>Money-weighted return (XIRR)</h2>
        </div>
        <label>
          BusinessDate
          <input
            aria-label="XIRR BusinessDate"
            type="date"
            value={asOfDate}
            onChange={(event) => onAsOfDateChange(event.target.value)}
          />
        </label>
      </div>

      {asOfDate === "" ? (
        <p className="muted">Choose an explicit BusinessDate to request the backend-owned return projection.</p>
      ) : null}

      {asOfDate !== "" && result === null ? <p className="muted">Loading return projection…</p> : null}

      {result?.ok === false ? (
        <div className="warning-text">
          <strong>Return request failed.</strong> {result.message}
        </div>
      ) : null}

      {result?.ok && result.data.status === "AVAILABLE" ? (
        <div>
          <p className="metric-label">Annual XIRR</p>
          <p className="metric-value">{percentFromDecimalString(result.data.xirr)}</p>
          <p className="muted">
            Backend result {result.data.xirr} · ACT/365 · external contributions/withdrawals plus the terminal portfolio value.
          </p>
          <p className="muted">
            Terminal value: {result.data.terminalPortfolioValue.amount} RUB for {result.data.asOfDate}. Open positions require explicit manual valuation for that exact date.
          </p>
        </div>
      ) : null}

      {result?.ok && result.data.status === "UNAVAILABLE" ? (
        <div className="warning-text">
          <strong>XIRR unavailable.</strong> {unavailableReasonText[result.data.reason]}
          <div className="muted">Reason: {result.data.reason} · ACT/365 · as of {result.data.asOfDate}</div>
        </div>
      ) : null}
    </section>
  );
}

function percentFromDecimalString(value: string) {
  const negative = value.startsWith("-");
  const unsigned = negative ? value.slice(1) : value;
  const [whole, fraction = ""] = unsigned.split(".");
  const digits = `${whole}${fraction.padEnd(2, "0")}`.replace(/^0+(?=\d)/, "");
  const decimalPlaces = fraction.length > 2 ? fraction.length - 2 : 0;
  const padded = digits.padStart(decimalPlaces + 1, "0");
  const split = padded.length - decimalPlaces;
  const percent = decimalPlaces === 0 ? padded : `${padded.slice(0, split)}.${padded.slice(split)}`;
  return `${negative ? "-" : ""}${percent}%`;
}
