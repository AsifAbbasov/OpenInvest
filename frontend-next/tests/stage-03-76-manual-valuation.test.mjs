import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const api = await readFile(new URL("../src/common/api/openinvest.ts", import.meta.url), "utf8");
const positions = await readFile(
  new URL("../src/features/portfolio/components/PositionsBlock.tsx", import.meta.url),
  "utf8",
);
const valuation = await readFile(
  new URL("../src/features/portfolio/components/ManualValuationCell.tsx", import.meta.url),
  "utf8",
);
const detail = await readFile(
  new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url),
  "utf8",
);

test("Stage 3.76 API keeps manual valuation explicitly user supplied and retry-safe", () => {
  assert.match(api, /status: "AVAILABLE"/);
  assert.match(api, /source: "USER_SUPPLIED"/);
  assert.match(api, /methodologyVersion: "portfolio-position-projection-v2-manual-valuation"/);
  assert.match(api, /method: "PUT"/);
  assert.match(api, /method: "DELETE"/);
  assert.match(api, /\/valuations\/\$\{encodeURIComponent\(ticker\)\}/);
  assert.doesNotMatch(api, /MOEX.*upsertManualValuation|broker.*upsertManualValuation/i);
});

test("Stage 3.76 UI renders backend-derived valuation and performs no financial arithmetic", () => {
  assert.match(valuation, /Source: Manual/);
  assert.match(valuation, /Market value:/);
  assert.match(valuation, /Unrealized P\/L:/);
  assert.match(valuation, /Unrealized return:/);
  assert.match(valuation, /Valued-position allocation:/);
  assert.match(valuation, /Add manual price/);
  assert.match(valuation, /Edit manual price/);
  assert.match(valuation, /Clear manual price/);
  assert.doesNotMatch(valuation, /\.quantity\s*[+*\/-]|\.acquisitionBasis\s*[+*\/-]|\.marketPrice\s*[+*\/-]|reduce\(/);
});

test("Stage 3.76 exposes partial versus complete portfolio valuation honestly", () => {
  assert.match(positions, /Manual valuation coverage/);
  assert.match(positions, /currentPortfolioValue/);
  assert.match(positions, /Unavailable until every open position has a manual valuation/);
  assert.match(positions, /Current is not a wall-clock market valuation/);
  assert.match(positions, /exact matching price date and position lifecycle/);
});

test("Stage 3.76 mutations refresh existing guarded current and historical projections", () => {
  assert.match(detail, /onValuationChanged=\{refreshAfterLedgerMutation\}/);
  assert.match(detail, /const refreshAfterLedgerMutation = useCallback/);
  assert.match(detail, /Promise\.all\(\[load\(\), loadHistoricalPositions\(\)\]\)/);
  assert.match(valuation, /mutationGeneration/);
  assert.match(valuation, /mutationGeneration\.current !== generation/);
});
