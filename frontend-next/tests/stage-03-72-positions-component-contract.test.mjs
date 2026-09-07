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
const styles = await readFile(new URL("../src/app/styles.css", import.meta.url), "utf8");

test("Stage 3.72 current positions remain in the existing guarded portfolio load", () => {
  assert.match(detail, /getPortfolioPositions\(portfolioId, \{ accessToken: attempt\.accessToken \}\)/);
  assert.match(detail, /const \[portfolio, summary, positions, transactions\] = await Promise\.all/);
  assert.match(detail, /shouldCommitPortfolioLoad\(loadGuard\.current, attempt\)/);
  assert.match(detail, /loadIdentity\.current\.principalId === principalAtLoad/);
  assert.match(detail, /loadIdentity\.current\.portfolioId === portfolioAtLoad/);
  assert.match(detail, /\[accessToken, principalId, portfolioId\]/);
  assert.match(detail, /setState\(\{ portfolio, summary, positions, transactions \}\)/);
  assert.match(detail, /positionViewMode === "historical" \? historicalPositions : state\?\.positions \?\? null/);
  assert.match(detail, /result=\{visiblePositions\}/);
});

test("Stage 3.72 positions render explicit loading, failure, empty and unavailable-market states without fallback math", () => {
  assert.match(positions, /Loading positions from the canonical ledger projection/);
  assert.match(positions, /Positions unavailable/);
  assert.match(positions, /No values are inferred from portfolio summary or transaction rows/);
  assert.match(positions, /No open stock or bond positions/);
  assert.match(positions, /Market valuation unavailable/);
  assert.match(positions, /formatQuantityForDisplay\(item\.quantity\)/);
  assert.doesNotMatch(positions, /formatDecimalForDisplay\(item\.quantity\)/);
  assert.doesNotMatch(positions, /marketPrice|marketValue|unrealizedGain|marketWeight/);
  assert.doesNotMatch(positions, /summary\.positions|listTransactions|reduce\(/);
  assert.doesNotMatch(positions, /acquisitionBasis\s*[+*\/-]|[+*\/-]\s*item\.acquisitionBasis/);
  assert.doesNotMatch(positions, /weightedAverageCost\s*[+*\/-]|[+*\/-]\s*item\.weightedAverageCost/);
});

test("Stage 3.72 accepted transaction and import flows still refresh the canonical projections", () => {
  assert.match(detail, /const refreshAfterLedgerMutation = useCallback/);
  assert.match(detail, /Promise\.all\(\[load\(\), loadHistoricalPositions\(\)\]\)/);
  assert.match(detail, /<AddTransactionForm[\s\S]*onSaved=\{refreshAfterLedgerMutation\}/);
  assert.match(detail, /<ImportUploadReviewPanel[\s\S]*onImported=\{refreshAfterLedgerMutation\}/);
});

test("Stage 3.72 positions provide distinct desktop table and mobile card surfaces", () => {
  assert.match(positions, /className="table-wrap positions-table"/);
  assert.match(positions, /className="position-cards"/);
  assert.match(positions, /Average acquisition price/);
  assert.match(positions, /Acquisition-basis weight/);
  assert.match(styles, /\.position-cards/);
  assert.match(styles, /\.positions-table/);
  assert.match(styles, /@media \(max-width: 720px\)/);
});
