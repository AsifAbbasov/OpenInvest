import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { renderToStaticMarkup } from "react-dom/server";

import type { ApiResult, PortfolioPositionProjection, PortfolioPositionsProjection } from "../src/common/api/openinvest.ts";
import { AcquisitionBasisAllocation } from "../src/features/portfolio/components/AcquisitionBasisAllocation";
import { PositionsBlock } from "../src/features/portfolio/components/PositionsBlock";

const rub = (amount: string) => ({ amount, currency: "RUB" as const });
const unavailable = {
  status: "UNAVAILABLE" as const,
  reason: "NO_APPROVED_MARKET_PRICE_SOURCE" as const,
  source: null,
  marketPrice: null,
  marketValue: null,
  unrealizedGain: null,
  unrealizedReturn: null,
  marketWeight: null,
  provider: null,
  asOf: null,
};

function position(ticker: string, assetType: "STOCK" | "BOND", basis: string, weight: string | null): PortfolioPositionProjection {
  return {
    ticker,
    assetType,
    quantity: "1.00000000",
    weightedAverageCost: rub(basis),
    acquisitionBasis: rub(basis),
    acquisitionBasisWeight: weight,
    marketValuation: unavailable,
  };
}

function projection(
  items: PortfolioPositionProjection[] = [
    position("STOCK-ONE", "STOCK", "750.00000000", "0.75000000"),
    position("BOND-TWO", "BOND", "250.00000000", "0.25000000"),
  ],
  totalAcquisitionBasis = "1000.00000000",
): ApiResult<PortfolioPositionsProjection> {
  return {
    ok: true,
    requestId: "00000000-0000-4000-8000-000000000001",
    data: {
      items,
      totalAcquisitionBasis: rub(totalAcquisitionBasis),
      valuationSummary: {
        status: "PARTIAL",
        valuedPositions: 0,
        totalOpenPositions: items.length,
        valuedPositionsMarketValue: rub("0.00000000"),
        valuedPositionsAcquisitionBasis: rub("0.00000000"),
        totalAcquisitionBasis: rub(totalAcquisitionBasis),
        unrealizedGain: rub("0.00000000"),
        cashValue: rub("99999.99000000"),
        currentPortfolioValue: null,
      },
      calculation: { methodologyVersion: "portfolio-position-projection-v2-manual-valuation", inputsAsOf: "2026-10-30" },
    },
  };
}

function render(result: ApiResult<PortfolioPositionsProjection> | null, viewMode: "current" | "historical" = "current") {
  return renderToStaticMarkup(<AcquisitionBasisAllocation result={result} viewMode={viewMode} />);
}

test("allocation uses server acquisition basis and weight, excludes cash and disclaims market allocation", () => {
  const html = render(projection());
  assert.match(html, /aria-label="Acquisition-basis allocation"/);
  assert.match(html, /Cash is excluded/);
  assert.match(html, /not market-value weights/);
  assert.match(html, /Total position acquisition basis: <strong>1 000\.00 ₽<\/strong>/);
  assert.match(html, /STOCK-ONE/);
  assert.match(html, /BOND-TWO/);
  assert.match(html, /750\.00 ₽ · 75\.0%/);
  assert.match(html, /250\.00 ₽ · 25\.0%/);
  assert.match(html, /style="width:75\.0%"/);
  assert.match(html, /style="width:25\.0%"/);
  assert.match(html, /aria-hidden="true"/);
  assert.doesNotMatch(html, /99 999\.99 ₽|marketWeight|market-price allocation/i);
});

test("negative control: changing only backend weight strings changes bars without rebasing money", () => {
  const original = render(projection());
  const swapped = render(projection([
    position("STOCK-ONE", "STOCK", "750.00000000", "0.25000000"),
    position("BOND-TWO", "BOND", "250.00000000", "0.75000000"),
  ]));
  assert.match(original, /STOCK-ONE[\s\S]*?750\.00 ₽ · 75\.0%[\s\S]*?style="width:75\.0%"/);
  assert.match(swapped, /STOCK-ONE[\s\S]*?750\.00 ₽ · 25\.0%[\s\S]*?style="width:25\.0%"/);
  assert.match(swapped, /BOND-TWO[\s\S]*?250\.00 ₽ · 75\.0%[\s\S]*?style="width:75\.0%"/);
  assert.notEqual(original, swapped);
});

test("zero total basis leaves ratios undefined rather than manufacturing allocation", () => {
  const html = render(projection([position("ZERO", "STOCK", "0.00000000", null)], "0.00000000"));
  assert.match(html, /total acquisition basis is zero/);
  assert.match(html, /role="status"/);
  assert.doesNotMatch(html, /acquisition-allocation-fill|style="width:/);
});

test("missing or invalid backend weights suppress all allocation bars", () => {
  for (const weight of [null, "", "-0.10000000", "1.10000000"]) {
    const html = render(projection([
      position("STOCK-ONE", "STOCK", "750.00000000", "0.75000000"),
      position("BOND-TWO", "BOND", "250.00000000", weight),
    ]));
    assert.match(html, /server-provided position weights are missing or invalid/);
    assert.doesNotMatch(html, /acquisition-allocation-list|style="width:/);
  }
});

test("empty, loading, error and historical modes reuse parent status without displaying allocation", () => {
  assert.equal(render(projection([], "0.00000000")), "");
  assert.equal(render(null), "");
  assert.equal(render({ ok: false, message: "Source unavailable", status: 503 }), "");
  assert.equal(render(projection(), "historical"), "");
});

test("PositionsBlock mounts allocation only for successful current positions and retains parent states", () => {
  const shared = {
    historicalDate: "2026-10-29",
    accessToken: null,
    portfolioId: "00000000-0000-4000-8000-000000000001",
    onShowCurrent: () => undefined,
    onShowHistorical: () => undefined,
    onHistoricalDateChange: (_date: string) => undefined,
    onValuationChanged: async () => undefined,
  };
  const current = renderToStaticMarkup(<PositionsBlock {...shared} result={projection()} viewMode="current" />);
  assert.match(current, /aria-label="Acquisition-basis allocation"/);
  assert.match(current, /aria-label="Portfolio positions"/);
  const historical = renderToStaticMarkup(<PositionsBlock {...shared} result={projection()} viewMode="historical" />);
  assert.doesNotMatch(historical, /aria-label="Acquisition-basis allocation"/);
  const loading = renderToStaticMarkup(<PositionsBlock {...shared} result={null} viewMode="current" />);
  assert.match(loading, /Loading positions/);
  assert.doesNotMatch(loading, /aria-label="Acquisition-basis allocation"/);
  const error = renderToStaticMarkup(
    <PositionsBlock {...shared} result={{ ok: false, message: "Source unavailable", status: 503 }} viewMode="current" />,
  );
  assert.match(error, /Positions unavailable/);
  assert.doesNotMatch(error, /aria-label="Acquisition-basis allocation"/);
  const empty = renderToStaticMarkup(<PositionsBlock {...shared} result={projection([], "0.00000000")} viewMode="current" />);
  assert.match(empty, /No open stock or bond positions/);
  assert.doesNotMatch(empty, /aria-label="Acquisition-basis allocation"/);
});

test("large exact money strings and long tickers remain present; responsive CSS allows wrapping", () => {
  const ticker = "LONG_STOCK_TICKER_WITH_NO_BREAKS_12345678901234567890";
  const large = "999999999999999999999.99000000";
  const html = render(projection([position(ticker, "STOCK", large, "1.00000000")], large));
  assert.ok(html.includes(ticker));
  assert.ok(html.includes("999 999 999 999 999 999 999.99 ₽"));
  assert.match(html, /style="width:100\.0%"/);
  const css = readFileSync(new URL("../src/app/styles.css", import.meta.url), "utf8");
  assert.match(css, /\.acquisition-allocation-ticker\s*\{\s*overflow-wrap: anywhere;/);
  assert.match(css, /\.acquisition-allocation-values\s*\{[\s\S]*?overflow-wrap: anywhere;/);
  assert.match(css, /\.acquisition-allocation-labels\s*\{[\s\S]*?flex-wrap: wrap;/);
});
