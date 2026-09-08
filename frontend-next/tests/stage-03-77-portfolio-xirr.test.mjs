import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const api = await readFile(new URL("../src/common/api/openinvest.ts", import.meta.url), "utf8");
const performance = await readFile(
  new URL("../src/features/portfolio/components/PerformanceBlock.tsx", import.meta.url),
  "utf8",
);
const detail = await readFile(
  new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url),
  "utf8",
);

test("Stage 3.77 API exposes a typed backend-owned exact-date projection", () => {
  assert.match(api, /type PortfolioReturnAvailable/);
  assert.match(api, /type PortfolioReturnUnavailable/);
  assert.match(api, /PortfolioReturnUnavailableReason/);
  assert.match(api, /methodologyVersion: "portfolio-xirr-v1"/);
  assert.match(api, /dayCountConvention: "ACT\/365"/);
  assert.match(api, /PortfolioReturnRequestOptions = \{\s*asOfDate: string;/s);
  assert.match(api, /\/returns\?\$\{searchParams\.toString\(\)\}/);
  assert.match(api, /new URLSearchParams\(\{ asOfDate: options\.asOfDate \}\)/);
});

test("Stage 3.77 UI renders AVAILABLE and UNAVAILABLE backend states distinctly", () => {
  assert.match(performance, /status === "AVAILABLE"/);
  assert.match(performance, /status === "UNAVAILABLE"/);
  assert.match(performance, /Money-weighted return \(XIRR\)/);
  assert.match(performance, /XIRR unavailable/);
  assert.match(performance, /ACT\/365/);
  assert.match(performance, /explicit manual valuation/);
  assert.match(performance, /Reason: \{result\.data\.reason\}/);
  assert.doesNotMatch(performance, /unavailable[^\n]*0%|0%[^\n]*unavailable/i);
});

test("Stage 3.77 frontend does not solve XIRR or fabricate terminal valuation", () => {
  assert.doesNotMatch(performance, /newton|bisection|npv|math\.pow|math\.exp|parsefloat|number\(result\.data\.xirr/i);
  assert.doesNotMatch(detail, /newton|bisection|npv|math\.pow|math\.exp|parsefloat/i);
  assert.doesNotMatch(performance, /provider|live price|market feed|quote/i);
  assert.match(performance, /result\.data\.terminalPortfolioValue\.amount/);
});

test("Stage 3.77 uses an explicit date and guarded refresh after financial mutations", () => {
  assert.match(detail, /const \[performanceDate, setPerformanceDate\] = useState\(""\)/);
  assert.match(detail, /performanceLoadGuard/);
  assert.match(detail, /performanceLoadIdentity/);
  assert.match(detail, /asOfDate: asOfDateAtLoad/);
  assert.match(detail, /shouldCommitPortfolioLoad\(performanceLoadGuard\.current, attempt\)/);
  assert.match(detail, /Promise\.all\(\[load\(\), loadHistoricalPositions\(\)\]\)/);
  assert.match(detail, /performanceIdentityIsCurrent/);
  assert.match(detail, /performanceLoadIdentity\.current\.asOfDate === performanceDate/);
  assert.match(detail, /if \(performanceIdentityIsCurrent\) \{\s*await loadPerformance\(\);/s);
  assert.match(detail, /onValuationChanged=\{refreshAfterLedgerMutation\}/);
  assert.match(detail, /onSaved=\{refreshAfterLedgerMutation\}/);
  assert.match(detail, /onMutated=\{refreshAfterLedgerMutation\}/);
});
