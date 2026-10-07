import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const detail = await readFile(
  new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url),
  "utf8",
);

function section(start, end) {
  const from = detail.indexOf(start);
  const to = detail.indexOf(end, from + start.length);
  assert.notEqual(from, -1, `missing section start: ${start}`);
  assert.notEqual(to, -1, `missing section end: ${end}`);
  return detail.slice(from, to);
}

test("normal portfolio UI expensive-read fanout remains five or fewer per render wave", () => {
  const mainLoad = section("const load = useCallback", "const loadHistoricalPositions = useCallback");
  const historicalLoad = section("const loadHistoricalPositions = useCallback", "const loadPerformance = useCallback");
  const performanceLoad = section("const loadPerformance = useCallback", "function invalidateHistoricalLoad");
  const mutationRefresh = section("const refreshAfterLedgerMutation = useCallback", "function showCurrentPositions");

  const mainExpensiveCalls = [
    "getPortfolioSummary(",
    "getPortfolioPositions(",
    "getPortfolioCashFlow(",
  ].filter((call) => mainLoad.includes(call));
  assert.deepEqual(mainExpensiveCalls, [
    "getPortfolioSummary(",
    "getPortfolioPositions(",
    "getPortfolioCashFlow(",
  ]);
  assert.match(mainLoad, /await Promise\.all\(\[/);

  assert.equal((historicalLoad.match(/getPortfolioPositions\(/g) ?? []).length, 1);
  assert.equal((performanceLoad.match(/getPortfolioReturns\(/g) ?? []).length, 1);

  // On a portfolio/principal transition, React runs passive effects from the committed render.
  // The current load (3), historical positions effect (1 when the previous historical date is
  // populated) and returns effect (1 when the previous performance date is populated) can all
  // be in flight before the reset effect's state update is committed: maximum automatic wave = 5.
  const normalFrontendMaxExpensiveReadFanout = mainExpensiveCalls.length + 1 + 1;
  assert.equal(normalFrontendMaxExpensiveReadFanout, 5);

  // A ledger mutation itself starts load() + historical positions together (4 expensive reads)
  // and only starts the performance refresh after that Promise.all completes.
  assert.match(mutationRefresh, /await Promise\.all\(\[load\(\), loadHistoricalPositions\(\)\]\)/);
  assert.match(mutationRefresh, /await loadPerformance\(\)/);
});

test("manual reload remains one main-load wave, not an extra hidden endpoint bundle", () => {
  assert.match(
    detail,
    /onClick=\{\(\) => void load\(\)\}/,
  );
});
