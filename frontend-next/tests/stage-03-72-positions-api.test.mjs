import assert from "node:assert/strict";
import test from "node:test";

const api = await import("../src/common/api/openinvest.ts");

function jsonResponse(data, status = 200) {
  return new Response(
    JSON.stringify({
      data,
      meta: {
        requestId: "11111111-1111-4111-8111-111111111111",
        traceId: "22222222222222222222222222222222",
        generatedAt: "2026-09-07T00:00:00Z",
      },
    }),
    { status, headers: { "Content-Type": "application/json" } },
  );
}

test("Stage 3.72 typed client requests positions with bearer auth and optional asOfDate", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url: String(url), init });
    return jsonResponse({
      items: [{
        ticker: "SBER",
        assetType: "STOCK",
        quantity: "150.00000000",
        weightedAverageCost: { amount: "275.00000000", currency: "RUB" },
        acquisitionBasis: { amount: "41250.00000000", currency: "RUB" },
        acquisitionBasisWeight: "1.00000000",
        marketValuation: {
          status: "UNAVAILABLE",
          reason: "NO_APPROVED_MARKET_PRICE_SOURCE",
          marketPrice: null,
          marketValue: null,
          unrealizedGain: null,
          marketWeight: null,
          provider: null,
          asOf: null,
        },
      }],
      totalAcquisitionBasis: { amount: "41250.00000000", currency: "RUB" },
      calculation: {
        methodologyVersion: "portfolio-position-projection-v1",
        inputsAsOf: "2026-09-03",
      },
    });
  };

  const result = await api.getPortfolioPositions(
    "portfolio/id",
    { accessToken: "access-token" },
    { asOfDate: "2026-09-03" },
  );

  assert.equal(result.ok, true);
  assert.equal(calls.length, 1);
  assert.equal(
    calls[0].url,
    "http://localhost:8080/api/v1/portfolios/portfolio%2Fid/positions?asOfDate=2026-09-03",
  );
  assert.equal(calls[0].init.headers.Authorization, "Bearer access-token");
  assert.equal(calls[0].init.credentials, undefined);
  assert.equal(result.data.items[0].marketValuation.marketPrice, null);
  assert.equal(result.data.items[0].acquisitionBasisWeight, "1.00000000");
});

test("Stage 3.72 typed client omits the query string when latest accepted ledger projection is requested", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url: String(url), init });
    return jsonResponse({
      items: [],
      totalAcquisitionBasis: { amount: "0.00000000", currency: "RUB" },
      calculation: { methodologyVersion: "portfolio-position-projection-v1", inputsAsOf: null },
    });
  };

  const result = await api.getPortfolioPositions("portfolio-id", { accessToken: "access-token" });
  assert.equal(result.ok, true);
  assert.equal(calls[0].url, "http://localhost:8080/api/v1/portfolios/portfolio-id/positions");
  assert.equal(result.data.calculation.inputsAsOf, null);
});
