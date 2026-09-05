import assert from "node:assert/strict";
import test from "node:test";

import { calculateDividend } from "../src/common/api/openinvest.ts";

function successResponse() {
  return new Response(JSON.stringify({
    data: {
      ticker: "SBER",
      quantity: "1000.00000000",
      dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
      grossDividend: { amount: "34840.00000000", currency: "RUB" },
      positionCost: { amount: "280000.00000000", currency: "RUB" },
      grossYield: "0.12442857",
      taxIncluded: false,
      methodologyVersion: "dividend-calculator-v1",
    },
    meta: {
      requestId: "10000000-0000-4000-8000-000000000001",
      traceId: "10000000000000000000000000000001",
      generatedAt: "2026-09-05T00:00:00Z",
    },
  }), { status: 200, headers: { "Content-Type": "application/json" } });
}

test("calculateDividend uses the shared OpenInvest client and sends exact input with explicit idempotency", async (t) => {
  const originalFetch = globalThis.fetch;
  let capturedURL = "";
  let capturedInit;
  globalThis.fetch = async (input, init) => {
    capturedURL = String(input);
    capturedInit = init;
    return successResponse();
  };
  t.after(() => { globalThis.fetch = originalFetch; });

  const result = await calculateDividend({
    ticker: "SBER",
    quantity: "1000.00000000",
    dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
    positionCost: { amount: "280000.00000000", currency: "RUB" },
  }, {
    idempotencyKey: "stage-03-68-key-00000001",
  });

  assert.equal(result.ok, true);
  assert.match(capturedURL, /\/api\/v1\/dividends\/calculate$/);
  assert.equal(capturedInit.method, "POST");
  assert.equal(capturedInit.credentials, "omit");
  assert.equal(capturedInit.headers["Idempotency-Key"], "stage-03-68-key-00000001");
  assert.deepEqual(JSON.parse(capturedInit.body), {
    ticker: "SBER",
    quantity: "1000.00000000",
    dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
    positionCost: { amount: "280000.00000000", currency: "RUB" },
  });
});

test("calculateDividend exposes explicit HTTP status but leaves transport uncertainty status-less", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });

  globalThis.fetch = async () => new Response(JSON.stringify({
    error: { code: "VALIDATION_ERROR", message: "quantity must be greater than zero" },
  }), { status: 400, headers: { "Content-Type": "application/json" } });
  const rejected = await calculateDividend({
    ticker: "SBER",
    quantity: "0.00000000",
    dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
  }, { idempotencyKey: "stage-03-68-key-00000002" });
  assert.equal(rejected.ok, false);
  assert.equal(rejected.status, 400);

  globalThis.fetch = async () => { throw new TypeError("network down"); };
  const uncertain = await calculateDividend({
    ticker: "SBER",
    quantity: "1.00000000",
    dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
  }, { idempotencyKey: "stage-03-68-key-00000003" });
  assert.equal(uncertain.ok, false);
  assert.equal(uncertain.status, undefined);
});
