import assert from "node:assert/strict";
import test from "node:test";

const api = await import("../src/common/api/openinvest.ts");

function errorResponse(message, status, headers = {}) {
  return new Response(
    JSON.stringify({
      error: { code: "RATE_LIMITED", message, details: [] },
      meta: {
        requestId: "11111111-1111-4111-8111-111111111111",
        traceId: "22222222222222222222222222222222",
        generatedAt: "2026-10-08T12:00:00Z",
      },
    }),
    { status, headers: { "Content-Type": "application/json", ...headers } },
  );
}

test("M11 frontend preserves backend enforcement but does not expose Retry-After to callers", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return errorResponse("Too many corporate actions requests", 429, { "Retry-After": "60" });
  };

  const result = await api.getCorporateActionProjection({
    instrumentIds: ["SBER"],
    from: "2026-01-01",
    to: "2026-12-31",
  });

  assert.equal(result.ok, false);
  assert.equal(result.status, 429);
  assert.equal(Object.hasOwn(result, "retryAfter"), false);
  assert.equal(calls.length, 1);
  assert.equal(calls[0].init.credentials, "omit");
  assert.equal(calls[0].init.headers.Authorization, undefined);
  assert.equal(calls[0].init.headers["X-CSRF-Token"], undefined);
  console.log("M11_FRONTEND_429_RETRY_AFTER_EXPOSED=NO");
  console.log("M11_FRONTEND_429_AUTOMATIC_RETRY=NO");
});

test("M11 financial mutations keep bearer and idempotency boundaries separated from cookies", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return new Response(
      JSON.stringify({
        data: {
          id: "transaction-id",
          portfolioId: "portfolio-id",
          transactionType: "DEPOSIT",
          status: "ACTIVE",
          ticker: null,
          quantity: null,
          unitPrice: null,
          grossAmount: { amount: "100.00000000", currency: "RUB" },
          commission: { amount: "0.00000000", currency: "RUB" },
          tax: { amount: "0.00000000", currency: "RUB" },
          tradeDate: "2026-10-08",
          settlementDate: null,
          revision: 1,
          createdAt: "2026-10-08T12:00:00Z",
          updatedAt: "2026-10-08T12:00:00Z",
        },
        meta: {
          requestId: "11111111-1111-4111-8111-111111111111",
          traceId: "22222222222222222222222222222222",
          generatedAt: "2026-10-08T12:00:00Z",
        },
      }),
      { status: 201, headers: { "Content-Type": "application/json" } },
    );
  };

  const result = await api.appendTransaction(
    "portfolio-id",
    {
      transactionType: "DEPOSIT",
      ticker: null,
      quantity: null,
      unitPrice: null,
      grossAmount: { amount: "100.00000000", currency: "RUB" },
      commission: { amount: "0.00000000", currency: "RUB" },
      tax: { amount: "0.00000000", currency: "RUB" },
      tradeDate: "2026-10-08",
      settlementDate: null,
    },
    { accessToken: "bearer-token", idempotencyKey: "m11-frontend-key-0001" },
  );

  assert.equal(result.ok, true);
  assert.equal(calls.length, 1);
  assert.equal(calls[0].init.headers.Authorization, "Bearer bearer-token");
  assert.equal(calls[0].init.headers["Idempotency-Key"], "m11-frontend-key-0001");
  assert.equal(calls[0].init.credentials, undefined);
  assert.equal(calls[0].init.headers["X-CSRF-Token"], undefined);
  console.log("M11_FRONTEND_BUSINESS_AUTH_COOKIE_CROSSOVER=NONE");
});
