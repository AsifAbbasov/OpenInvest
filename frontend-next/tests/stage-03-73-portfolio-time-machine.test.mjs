import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const detail = await readFile(
  new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url),
  "utf8",
);
const positions = await readFile(
  new URL("../src/features/portfolio/components/PositionsBlock.tsx", import.meta.url),
  "utf8",
);
const api = await readFile(new URL("../src/common/api/openinvest.ts", import.meta.url), "utf8");

test("Stage 3.73 reuses the Stage 3.72 positions endpoint instead of creating frontend financial state", () => {
  assert.match(detail, /getPortfolioPositions\([\s\S]*\{ asOfDate: historicalDate \}/);
  assert.match(api, /export type PortfolioPositionsRequestOptions = \{[\s\S]*asOfDate\?: string/);
  assert.match(api, /\/positions\$\{query === "" \? "" : `\?\$\{query\}`\}/);
  assert.doesNotMatch(detail, /fetch\(|axios|\/time-machine|\/historical-positions/);
  assert.doesNotMatch(detail, /weightedAverageCost\s*[+*\/-]|acquisitionBasis\s*[+*\/-]|reduce\(/);
});

test("Stage 3.73 exposes explicit Current and Historical states without calling Current Today", () => {
  assert.match(positions, />Current</);
  assert.match(positions, />Historical</);
  assert.match(positions, /type="date"/);
  assert.match(positions, /Viewing the latest accepted ledger projection/);
  assert.match(positions, /Current is not a wall-clock market valuation/);
  assert.match(positions, /Viewing portfolio as of \$\{historicalDate\}/);
  assert.doesNotMatch(positions, />Today</);
  assert.doesNotMatch(positions, /max=\{/);
});

test("Stage 3.73 guards rapid historical switching by generation, principal, portfolio and requested date", () => {
  assert.match(detail, /historicalLoadGuard = useRef<PortfolioLoadGuardState>/);
  assert.match(detail, /startPortfolioLoad\([\s\S]*historicalLoadGuard\.current/);
  assert.match(detail, /historicalLoadIdentity\.current\.principalId === principalAtLoad/);
  assert.match(detail, /historicalLoadIdentity\.current\.portfolioId === portfolioAtLoad/);
  assert.match(detail, /historicalLoadIdentity\.current\.viewKey === viewKeyAtLoad/);
  assert.match(detail, /shouldCommitPortfolioLoad\(historicalLoadGuard\.current, attempt\)/);
  assert.match(detail, /AS_OF:\$\{historicalDate\}/);
});

test("Stage 3.73 has honest loading, date-selection, empty and error states", () => {
  assert.match(positions, /Choose a historical BusinessDate/);
  assert.match(positions, /Loading portfolio as of \$\{historicalDate\}/);
  assert.match(positions, /No open stock or bond positions on \$\{historicalDate\}/);
  assert.match(positions, /Positions unavailable/);
  assert.match(positions, /No values are inferred from portfolio summary or transaction rows/);
});

test("Stage 3.73 preserves market-unavailable semantics in historical mode", () => {
  assert.match(positions, /Market valuation unavailable/);
  assert.doesNotMatch(positions, /marketPrice|marketValue|unrealizedGain|marketWeight/);
  assert.doesNotMatch(positions, /Portfolio value over time|market return|XIRR/);
});

test("Stage 3.73 refreshes selected historical truth after accepted ledger mutation", () => {
  assert.match(detail, /const refreshAfterLedgerMutation = useCallback/);
  assert.match(detail, /Promise\.all\(\[load\(\), loadHistoricalPositions\(\)\]\)/);
  assert.match(detail, /onSaved=\{refreshAfterLedgerMutation\}/);
  assert.match(detail, /onImported=\{refreshAfterLedgerMutation\}/);
});
