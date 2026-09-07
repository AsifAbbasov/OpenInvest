import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const root = new URL("../", import.meta.url);
const read = (path) => readFile(new URL(path, root), "utf8");

test("Stage 3.75 exposes all canonical manual transaction types without frontend cash-flow arithmetic", async () => {
  const form = await read("src/features/portfolio/components/AddTransactionForm.tsx");
  for (const type of ["BUY", "SELL", "DIVIDEND", "COUPON", "FEE", "TAX", "DEPOSIT", "WITHDRAWAL"]) {
    assert.match(form, new RegExp(`\\"${type}\\"`));
  }
  assert.match(form, /isAssetIncome/);
  assert.match(form, /isExpense/);
  assert.match(form, /grossAmount:/);
  assert.match(form, /quantity: isTrade \?/);
  assert.doesNotMatch(form, /netInvestmentIncome\s*=/);
  assert.doesNotMatch(form, /netCashFlow\s*=/);
});

test("Stage 3.75 uses a typed bounded backend aggregate and renders server values only", async () => {
  const api = await read("src/common/api/openinvest.ts");
  const block = await read("src/features/portfolio/components/CashFlowIncomeBlock.tsx");
  const detail = await read("src/features/portfolio/components/PortfolioDetailSlice.tsx");

  assert.match(api, /getPortfolioCashFlow/);
  assert.match(api, /\/cash-flow/);
  assert.match(api, /PortfolioCashFlowProjection/);
  assert.match(block, /projection\.periods\.map/);
  assert.match(block, /formatMoney\(totals\.buyOutflows\)/);
  assert.match(block, /formatMoney\(totals\.sellInflows\)/);
  assert.match(block, /formatMoney\(totals\.netInvestmentIncome\)/);
  assert.match(block, /period\.totals\.buyOutflows/);
  assert.match(block, /period\.totals\.sellInflows/);
  assert.match(block, /their own recorded commission\/tax/);
  assert.match(block, /accepted future-dated rows/);
  assert.match(block, /not portfolio performance/);
  assert.doesNotMatch(block, /reduce\(/);
  assert.doesNotMatch(block, /parseFloat|Number\(/);
  assert.match(detail, /getPortfolioCashFlow/);
  assert.match(detail, /CashFlowIncomeBlock/);
  assert.doesNotMatch(detail, /Nominal return rate/);
  assert.doesNotMatch(detail, />XIRR</);
  assert.doesNotMatch(detail, /Real gain/);
});
