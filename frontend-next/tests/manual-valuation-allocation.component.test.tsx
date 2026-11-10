import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { renderToStaticMarkup } from "react-dom/server";

import type {
  ApiResult,
  MarketValuation,
  PortfolioPositionProjection,
  PortfolioPositionsProjection,
} from "../src/common/api/openinvest.ts";
import { AcquisitionBasisAllocation } from "../src/features/portfolio/components/AcquisitionBasisAllocation";
import { ManualValuationAllocation } from "../src/features/portfolio/components/ManualValuationAllocation";
import { PositionsBlock } from "../src/features/portfolio/components/PositionsBlock";

const rub = (amount: string) => ({ amount, currency: "RUB" as const });
const unavailable: MarketValuation = {
  status: "UNAVAILABLE",
  reason: "NO_APPROVED_MARKET_PRICE_SOURCE",
  source: null,
  marketPrice: null,
  marketValue: null,
  unrealizedGain: null,
  unrealizedReturn: null,
  marketWeight: null,
  provider: null,
  asOf: null,
};

function manuallyValued(amount: string, weight: string | null, asOf: string): MarketValuation {
  return {
    status: "AVAILABLE",
    reason: null,
    source: "USER_SUPPLIED",
    marketPrice: rub("10.00000000"),
    marketValue: rub(amount),
    unrealizedGain: rub("0.00000000"),
    unrealizedReturn: "0.00000000",
    marketWeight: weight,
    provider: null,
    asOf,
  };
}

function position(
  ticker: string,
  assetType: "STOCK" | "BOND",
  valuation: MarketValuation,
  acquisitionBasisWeight = "0.50000000",
): PortfolioPositionProjection {
  return {
    ticker,
    assetType,
    quantity: "10.00000000",
    weightedAverageCost: rub("50.00000000"),
    acquisitionBasis: rub("500.00000000"),
    acquisitionBasisWeight,
    marketValuation: valuation,
  };
}

const first = () => position("FIRST", "STOCK", manuallyValued("750.00000000", "0.75000000", "2026-10-28"));
const second = () => position("SECOND", "BOND", manuallyValued("250.00000000", "0.25000000", "2026-10-30"));
const third = () => position("NO-PRICE", "STOCK", unavailable);

function projection(
  items: PortfolioPositionProjection[],
  status: "COMPLETE" | "PARTIAL",
  valuedPositions: number,
  valuedPositionsMarketValue = "1000.00000000",
): ApiResult<PortfolioPositionsProjection> {
  return {
    ok: true,
    requestId: "00000000-0000-4000-8000-000000000001",
    data: {
      items,
      totalAcquisitionBasis: rub("1000.00000000"),
      valuationSummary: {
        status,
        valuedPositions,
        totalOpenPositions: items.length,
        valuedPositionsMarketValue: rub(valuedPositionsMarketValue),
        valuedPositionsAcquisitionBasis: rub("1000.00000000"),
        totalAcquisitionBasis: rub("1000.00000000"),
        unrealizedGain: rub("0.00000000"),
        cashValue: rub("99999.99000000"),
        currentPortfolioValue: status === "COMPLETE" ? rub("123456.78000000") : null,
      },
      calculation: { methodologyVersion: "portfolio-position-projection-v2-manual-valuation", inputsAsOf: "2026-10-30" },
    },
  };
}

function render(result: ApiResult<PortfolioPositionsProjection> | null, viewMode: "current" | "historical" = "current") {
  return renderToStaticMarkup(<ManualValuationAllocation result={result} viewMode={viewMode} />);
}

test("COMPLETE coverage uses valued-securities denominator, server weights and individual manual dates", () => {
  const html = render(projection([first(), second()], "COMPLETE", 2));
  assert.match(html, /aria-label="Manual-valued position allocation"/);
  assert.match(html, /Coverage: <strong>2 \/ 2<\/strong> open positions valued · COMPLETE/);
  assert.match(html, /Denominator — manually valued securities only, excluding cash: <strong>1 000\.00 ₽<\/strong>/);
  assert.match(html, /FIRST[\s\S]*?750\.00 ₽ · 75\.0%[\s\S]*?2026-10-28[\s\S]*?style="width:75\.0%"/);
  assert.match(html, /SECOND[\s\S]*?250\.00 ₽ · 25\.0%[\s\S]*?2026-10-30[\s\S]*?style="width:25\.0%"/);
  assert.match(html, /Source: USER_SUPPLIED/);
  assert.match(html, /USER_SUPPLIED manual prices, not live, broker, or provider-confirmed quotes/);
  assert.match(html, /dates may differ/);
  assert.match(html, /aria-hidden="true"/);
  assert.doesNotMatch(html, /99 999\.99 ₽|123 456\.78 ₽|currentPortfolioValue|portfolio market allocation/);
});

test("PARTIAL coverage labels allocation only among valued positions, excluding cash and unavailable securities", () => {
  const html = render(projection([first(), second(), third()], "PARTIAL", 2));
  assert.match(html, /Manual valuation allocation — among valued positions only/);
  assert.match(html, /Coverage: <strong>2 \/ 3<\/strong> open positions valued · PARTIAL/);
  assert.match(html, /750\.00 ₽ · 75\.0%/);
  assert.match(html, /250\.00 ₽ · 25\.0%/);
  assert.doesNotMatch(html, /NO-PRICE|99 999\.99 ₽|123 456\.78 ₽|full portfolio market value/i);
});

test("no eligible manual valuations and all unavailable positions display no bars", () => {
  const html = render(projection([third(), position("ALSO-EMPTY", "BOND", unavailable)], "PARTIAL", 0, "0.00000000"));
  assert.match(html, /Coverage: <strong>0 \/ 2<\/strong>/);
  assert.match(html, /no open positions have eligible manual prices/);
  assert.doesNotMatch(html, /acquisition-allocation-fill|style="width:/);
});

test("zero total valued-security amount suppresses bars even with a valued position", () => {
  const html = render(projection([position("ZERO", "STOCK", manuallyValued("0.00000000", null, "2026-10-31"))], "COMPLETE", 1, "0.00000000"));
  assert.match(html, /total manually valued securities amount is zero/);
  assert.doesNotMatch(html, /acquisition-allocation-fill|style="width:/);
});

test("a zero server marketWeight remains an explicitly displayed zero, not a missing weight", () => {
  const html = render(projection([
    position("TINY", "STOCK", manuallyValued("0.00000000", "0.00000000", "2026-10-27")),
    position("OTHER", "BOND", manuallyValued("1000.00000000", "1.00000000", "2026-10-30")),
  ], "COMPLETE", 2));
  assert.match(html, /TINY[\s\S]*?0\.00 ₽ · 0\.0%[\s\S]*?style="width:0\.0%"/);
  assert.match(html, /OTHER[\s\S]*?1 000\.00 ₽ · 100\.0%[\s\S]*?style="width:100\.0%"/);
  assert.doesNotMatch(html, /Allocation unavailable/);
});

test("negative control: swapping only server marketWeight changes bar widths but not monetary values", () => {
  const original = render(projection([first(), second()], "COMPLETE", 2));
  const swapped = render(projection([
    position("FIRST", "STOCK", manuallyValued("750.00000000", "0.25000000", "2026-10-28")),
    position("SECOND", "BOND", manuallyValued("250.00000000", "0.75000000", "2026-10-30")),
  ], "COMPLETE", 2));
  assert.match(original, /FIRST[\s\S]*?750\.00 ₽ · 75\.0%[\s\S]*?style="width:75\.0%"/);
  assert.match(swapped, /FIRST[\s\S]*?750\.00 ₽ · 25\.0%[\s\S]*?style="width:25\.0%"/);
  assert.match(swapped, /SECOND[\s\S]*?250\.00 ₽ · 75\.0%[\s\S]*?style="width:75\.0%"/);
  assert.notEqual(original, swapped);
});

test("negative control: changing acquisitionBasisWeight only leaves manual market bars identical", () => {
  const original = render(projection([first(), second()], "COMPLETE", 2));
  const altered = render(projection([
    position("FIRST", "STOCK", manuallyValued("750.00000000", "0.75000000", "2026-10-28"), "0.10000000"),
    position("SECOND", "BOND", manuallyValued("250.00000000", "0.25000000", "2026-10-30"), "0.90000000"),
  ], "COMPLETE", 2));
  assert.equal(altered, original);
});

test("any null, malformed, negative or greater-than-one manual weight suppresses the entire chart", () => {
  for (const invalid of [null, "", "abc", "-0.01000000", "1.10000000", "0.123456789"]) {
    const html = render(projection([
      first(),
      position("SECOND", "BOND", manuallyValued("250.00000000", invalid, "2026-10-30")),
    ], "COMPLETE", 2));
    assert.match(html, /server-provided market weights are missing or invalid/);
    assert.doesNotMatch(html, /acquisition-allocation-list|style="width:/);
  }
});

test("mismatched server coverage fails closed rather than presenting a misleading allocation", () => {
  const mismatched = render(projection([first(), third()], "COMPLETE", 1));
  assert.match(mismatched, /server valuation coverage is inconsistent/);
  assert.doesNotMatch(mismatched, /style="width:/);
});

test("large exact RUB values and long labels stay present with shared responsive CSS", () => {
  const ticker = "LONG_TICKER_NO_BREAKS_123456789012345678901234567890";
  const large = "999999999999999999999.99000000";
  const html = render(projection([position(ticker, "STOCK", manuallyValued(large, "1.00000000", "2026-10-29"))], "COMPLETE", 1, large));
  assert.ok(html.includes(ticker));
  assert.ok(html.includes("999 999 999 999 999 999 999.99 ₽"));
  assert.match(html, /style="width:100\.0%"/);
  const css = readFileSync(new URL("../src/app/styles.css", import.meta.url), "utf8");
  assert.match(css, /\.acquisition-allocation-ticker\s*\{\s*overflow-wrap: anywhere;/);
  assert.match(css, /\.manual-valuation-allocation-source\s*\{[\s\S]*?overflow-wrap: anywhere;/);
  assert.match(css, /\.acquisition-allocation-labels\s*\{[\s\S]*?flex-wrap: wrap;/);
});

test("empty, loading, error and historical view do not render manual allocation", () => {
  assert.equal(render(projection([], "COMPLETE", 0, "0.00000000")), "");
  assert.equal(render(null), "");
  assert.equal(render({ ok: false, message: "Source unavailable", status: 503 }), "");
  assert.equal(render(projection([first(), second()], "COMPLETE", 2), "historical"), "");
});

test("PositionsBlock integrates B2 beside B1 while preserving current, empty, loading, error and historical states", () => {
  const props = {
    historicalDate: "2026-10-29",
    accessToken: null,
    portfolioId: "00000000-0000-4000-8000-000000000001",
    onShowCurrent: () => undefined,
    onShowHistorical: () => undefined,
    onHistoricalDateChange: (_date: string) => undefined,
    onValuationChanged: async () => undefined,
  };
  const populated = projection([first(), second(), third()], "PARTIAL", 2);
  const current = renderToStaticMarkup(<PositionsBlock {...props} result={populated} viewMode="current" />);
  assert.match(current, /aria-label="Acquisition-basis allocation"/);
  assert.match(current, /aria-label="Manual-valued position allocation"/);
  assert.match(current, /Manual valuation coverage/);
  assert.match(current, /Add manual price/);
  assert.match(current, /Edit manual price/);
  assert.match(current, /Portfolio positions mobile view/);
  const acquisition = renderToStaticMarkup(<AcquisitionBasisAllocation result={populated} viewMode="current" />);
  assert.match(acquisition, /Share of remaining acquisition cost/);
  const historical = renderToStaticMarkup(<PositionsBlock {...props} result={populated} viewMode="historical" />);
  assert.doesNotMatch(historical, /aria-label="Manual-valued position allocation"|aria-label="Acquisition-basis allocation"/);
  const loading = renderToStaticMarkup(<PositionsBlock {...props} result={null} viewMode="current" />);
  assert.match(loading, /Loading positions/);
  assert.doesNotMatch(loading, /aria-label="Manual-valued position allocation"/);
  const error = renderToStaticMarkup(<PositionsBlock {...props} result={{ ok: false, message: "Source unavailable", status: 503 }} viewMode="current" />);
  assert.match(error, /Positions unavailable/);
  assert.doesNotMatch(error, /aria-label="Manual-valued position allocation"/);
  const empty = renderToStaticMarkup(<PositionsBlock {...props} result={projection([], "COMPLETE", 0, "0.00000000")} viewMode="current" />);
  assert.match(empty, /No open stock or bond positions/);
  assert.doesNotMatch(empty, /aria-label="Manual-valued position allocation"/);
});
