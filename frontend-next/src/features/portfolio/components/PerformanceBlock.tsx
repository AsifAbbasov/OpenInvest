"use client";

import type { ApiResult, PortfolioReturnProjection, PortfolioReturnUnavailableReason } from "@/common/api/openinvest";

type PerformanceBlockProps = {
  result: ApiResult<PortfolioReturnProjection> | null;
  asOfDate: string;
  onAsOfDateChange: (value: string) => void;
};

const unavailableReasonText: Record<PortfolioReturnUnavailableReason, string> = {
  INCOMPLETE_VALUATION: "Add a manual valuation for every open position on this date.",
  NO_EXTERNAL_CONTRIBUTIONS: "No qualifying external cash flows are available for this date range.",
  INSUFFICIENT_DATE_SPAN: "The available cash flows do not cover more than one date.",
  NO_SIGN_CHANGE: "The cash flows do not produce a valid XIRR for this period.",
  NON_POSITIVE_TERMINAL_VALUE: "The portfolio value on this date must be positive to calculate XIRR.",
  AMBIGUOUS_MULTIPLE_ROOTS: "More than one valid XIRR is possible for these cash flows, so no single return is shown.",
  NUMERICAL_SOLUTION_FAILED: "A single reliable XIRR could not be calculated for these cash flows.",
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
          Date
          <input
            aria-label="XIRR date"
            type="date"
            value={asOfDate}
            onChange={(event) => onAsOfDateChange(event.target.value)}
          />
        </label>
      </div>

      {asOfDate === "" ? (
        <p className="muted">Choose a date to calculate money-weighted return.</p>
      ) : null}

      {asOfDate !== "" && result === null ? <p className="muted">Loading return…</p> : null}

      {result?.ok === false ? (
        <div className="warning-text">
          <strong>Return unavailable.</strong> {result.message}
        </div>
      ) : null}

      {result?.ok && result.data.status === "AVAILABLE" ? (
        <div>
          <p className="metric-label">Annual XIRR</p>
          <p className="metric-value">{percentFromDecimalString(result.data.xirr)}</p>
          <p className="muted">
            ACT/365 · based on deposits, withdrawals and the portfolio value on the selected date.
          </p>
          <p className="muted">
            Terminal value: {result.data.terminalPortfolioValue.amount} RUB for {result.data.asOfDate}. Open positions require explicit manual valuation for that exact date.
          </p>
        </div>
      ) : null}

      {result?.ok && result.data.status === "UNAVAILABLE" ? (
        <div className="warning-text">
          <strong>XIRR unavailable.</strong> {unavailableReasonText[result.data.reason]}
          <div className="muted">ACT/365 · as of {result.data.asOfDate}</div>
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
