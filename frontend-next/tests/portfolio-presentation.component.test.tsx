import assert from "node:assert/strict";
import test from "node:test";
import { renderToStaticMarkup } from "react-dom/server";

import type { PortfolioSummary } from "../src/common/api/openinvest.ts";
import { CapitalSummary } from "../src/features/portfolio/components/PortfolioDetailSlice";

const rub = (amount: string) => ({ amount, currency: "RUB" as const });

function summaryFixture(): PortfolioSummary {
  return {
    portfolioId: "00000000-0000-4000-8000-000000000001",
    asOfDate: "2026-10-30",
    totalValue: rub("1220.00000000"),
    cashValue: rub("-80.00000000"),
    stockValue: rub("900.00000000"),
    bondValue: rub("400.00000000"),
    investedCapital: rub("1500.00000000"),
    dividendsReceived: rub("12.34500000"),
    couponsReceived: rub("5.33500000"),
    nominalReturnRate: null,
    xirr: null,
    realReturn: null,
    purchasingPower: {
      portfolioValue: rub("1220.00000000"),
      asOfDate: "2026-10-30",
      equivalents: [],
    },
    positions: [],
    calculation: {
      methodologyVersion: "portfolio-summary-v1",
      calculatedAt: "2026-10-30T12:00:00Z",
      inputsAsOf: "2026-10-30",
    },
  };
}

function renderSummary(summary: PortfolioSummary) {
  const html = renderToStaticMarkup(<CapitalSummary summary={summary} />);
  assert.match(html, /<section class="panel capital-summary" aria-label="Current portfolio summary">/);
  return html;
}

test("capital summary displays the server as-of date and all seven server values", () => {
  const html = renderSummary(summaryFixture());
  assert.match(html, /<time dateTime="2026-10-30">2026-10-30<\/time>/i);
  for (const value of ["1 220.00 ₽", "-80.00 ₽", "900.00 ₽", "400.00 ₽", "1 500.00 ₽", "12.34 ₽", "5.34 ₽"]) {
    assert.ok(html.includes(value), `missing server-provided value ${value}`);
  }
  assert.match(html, /Bonds \(cost basis\)/);
  assert.match(html, /Gross dividends recorded/);
  assert.match(html, /Gross coupons recorded/);
});

test("capital and income metrics carry precise scope and accessible group labels", () => {
  const html = renderSummary(summaryFixture());
  assert.match(html, /role="group" aria-label="Capital components"/);
  assert.match(html, /role="group" aria-label="Invested capital and recorded gross income"/);
  assert.match(html, /not a current market valuation/);
  assert.match(html, /not external contributions/);
  assert.match(html, /before fees and taxes/);
  assert.doesNotMatch(html, /market return|live quote|portfolio contributors/i);
});

test("negative control: each displayed value comes from its own server field", () => {
  const original = summaryFixture();
  const mutated = {
    ...original,
    totalValue: rub("999.00000000"),
    bondValue: rub("321.00000000"),
    asOfDate: "2026-10-29",
  };
  const first = renderSummary(original);
  const second = renderSummary(mutated);
  assert.match(second, /999\.00 ₽/);
  assert.match(second, /321\.00 ₽/);
  assert.doesNotMatch(second, /1 220\.00 ₽|400\.00 ₽/);
  assert.ok(second.includes("900.00 ₽") && second.includes("-80.00 ₽"));
  assert.match(second, /<time dateTime="2026-10-29">2026-10-29<\/time>/i);
  assert.notEqual(first, second);
});

test("zero gross income remains a recorded zero, not a fabricated return", () => {
  const summary = summaryFixture();
  summary.dividendsReceived = rub("0.00000000");
  summary.couponsReceived = rub("0.00000000");
  const html = renderSummary(summary);
  for (const label of ["Gross dividends recorded", "Gross coupons recorded"]) {
    assert.ok(html.includes(`<span>${label}</span><strong>0.00 ₽</strong>`), `${label} must show zero`);
  }
  assert.doesNotMatch(html, /0%|annual return|guaranteed/i);
});
