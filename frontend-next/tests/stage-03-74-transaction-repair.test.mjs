import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const api = fs.readFileSync(new URL("../src/common/api/openinvest.ts", import.meta.url), "utf8");
const detail = fs.readFileSync(new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url), "utf8");
const controls = fs.readFileSync(new URL("../src/features/portfolio/components/TransactionRepairControls.tsx", import.meta.url), "utf8");

test("Stage 3.74 uses the frozen PATCH and DELETE transaction commands", () => {
  assert.match(api, /method: "PATCH"/);
  assert.match(api, /method: "DELETE"/);
  assert.match(api, /expectedRevision/);
  assert.match(api, /effectiveDate/);
  assert.match(api, /IdempotentAuthenticatedRequest/);
});

test("Stage 3.74 exposes explicit Edit and Reverse UX without delete-permanently language", () => {
  assert.match(detail, /<th>Actions<\/th>/);
  assert.match(controls, />\s*Edit\s*</);
  assert.match(controls, />\s*Reverse\s*</);
  assert.match(controls, /History will be preserved/);
  assert.doesNotMatch(controls, /Delete permanently/i);
  assert.match(controls, /Reason for correction/);
  assert.match(controls, /Reason for reversal/);
  assert.match(controls, /aria-live="polite"/);
});

test("Stage 3.74 preserves browser retry identity for ambiguous repair requests", () => {
  assert.match(controls, /idempotencyIntentForBrowser/);
  assert.match(controls, /principalScopedIdempotencyScope/);
  assert.match(controls, /transaction-correct:/);
  assert.match(controls, /transaction-reverse:/);
  assert.match(controls, /clearBrowserIdempotencyIntent/);
  assert.doesNotMatch(controls, /idempotencyKey:\s*crypto\.randomUUID\(\)/);
});
