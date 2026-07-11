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
        generatedAt: "2026-07-11T00:00:00Z",
      },
    }),
    { status, headers: { "Content-Type": "application/json" } },
  );
}

test("register and login use the Go API directly with credentialed refresh-cookie boundary", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return jsonResponse({
      user: {
        id: "user-id",
        email: "investor@example.com",
        language: "en",
        theme: "system",
        timezone: "UTC",
        privacyMode: true,
        createdAt: "2026-07-11T00:00:00Z",
      },
      session: {
        accessToken: "access-token",
        accessTokenExpiresAt: "2026-07-11T00:15:00Z",
        csrfToken: "csrf-token",
      },
    });
  };

  const result = await api.register({
    email: "investor@example.com",
    password: "correct horse battery staple",
    language: "en",
    theme: "system",
    timezone: "UTC",
  });
  const loginResult = await api.login({
    email: "investor@example.com",
    password: "correct horse battery staple",
  });

  assert.equal(result.ok, true);
  assert.equal(loginResult.ok, true);
  assert.equal(calls[0].url, "http://localhost:8080/api/v1/auth/register");
  assert.equal(calls[0].init.method, "POST");
  assert.equal(calls[0].init.credentials, "include");
  assert.equal(calls[0].init.headers.Authorization, undefined);
  assert.deepEqual(JSON.parse(calls[0].init.body), {
    email: "investor@example.com",
    password: "correct horse battery staple",
    language: "en",
    theme: "system",
    timezone: "UTC",
  });
  assert.equal(calls[1].url, "http://localhost:8080/api/v1/auth/login");
  assert.equal(calls[1].init.method, "POST");
  assert.equal(calls[1].init.credentials, "include");
  assert.equal(calls[1].init.headers.Authorization, undefined);
  assert.deepEqual(JSON.parse(calls[1].init.body), {
    email: "investor@example.com",
    password: "correct horse battery staple",
  });
});

test("refresh and logout send CSRF with credentialed cookie requests", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    if (String(url).endsWith("/refresh")) {
      return jsonResponse({
        accessToken: "rotated-access-token",
        accessTokenExpiresAt: "2026-07-11T00:30:00Z",
        csrfToken: "rotated-csrf-token",
      });
    }
    return jsonResponse({ revoked: true });
  };

  await api.refreshSession("csrf-token");
  await api.logout({ allSessions: false }, "rotated-csrf-token");

  assert.equal(calls[0].url, "http://localhost:8080/api/v1/auth/refresh");
  assert.equal(calls[0].init.method, "POST");
  assert.equal(calls[0].init.credentials, "include");
  assert.equal(calls[0].init.headers["X-CSRF-Token"], "csrf-token");
  assert.equal(calls[0].init.headers.Authorization, undefined);
  assert.equal(calls[0].init.body, undefined);
  assert.equal(calls[1].url, "http://localhost:8080/api/v1/auth/logout");
  assert.equal(calls[1].init.method, "POST");
  assert.equal(calls[1].init.credentials, "include");
  assert.equal(calls[1].init.headers["X-CSRF-Token"], "rotated-csrf-token");
  assert.equal(calls[1].init.headers.Authorization, undefined);
  assert.deepEqual(JSON.parse(calls[1].init.body), { allSessions: false });
});

test("business API requests include bearer access token without browser storage", async () => {
  const calls = [];
  const originalLocalStorage = globalThis.localStorage;
  const originalSessionStorage = globalThis.sessionStorage;
  const originalIndexedDB = globalThis.indexedDB;
  const originalDocument = globalThis.document;

  Object.defineProperty(globalThis, "localStorage", {
    configurable: true,
    get() {
      throw new Error("localStorage must not be used by the auth API client");
    },
  });
  Object.defineProperty(globalThis, "sessionStorage", {
    configurable: true,
    get() {
      throw new Error("sessionStorage must not be used by the auth API client");
    },
  });
  Object.defineProperty(globalThis, "indexedDB", {
    configurable: true,
    get() {
      throw new Error("indexedDB must not be used by the auth API client");
    },
  });
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    get() {
      throw new Error("document.cookie must not be used by the auth API client");
    },
  });

  try {
    globalThis.fetch = async (url, init) => {
      calls.push({ url, init });
      if (String(url).endsWith("/portfolios/portfolio-id/summary")) {
        return jsonResponse({
          portfolioId: "portfolio-id",
          asOfDate: "2026-07-11",
          totalValue: { amount: "100.00000000", currency: "RUB" },
          cashValue: { amount: "100.00000000", currency: "RUB" },
          stockValue: { amount: "0.00000000", currency: "RUB" },
          bondValue: { amount: "0.00000000", currency: "RUB" },
          investedCapital: { amount: "100.00000000", currency: "RUB" },
          dividendsReceived: { amount: "0.00000000", currency: "RUB" },
          couponsReceived: { amount: "0.00000000", currency: "RUB" },
          nominalReturnRate: "0.00000000",
          xirr: null,
          realReturn: {
            nominalReturnRate: "0.00000000",
            inflationRate: "0.00000000",
            realReturnRate: "0.00000000",
            nominalGain: { amount: "0.00000000", currency: "RUB" },
            realGain: { amount: "0.00000000", currency: "RUB" },
            fromDate: "2026-07-11",
            toDate: "2026-07-11",
            methodologyVersion: "test",
          },
          purchasingPower: {
            portfolioValue: { amount: "100.00000000", currency: "RUB" },
            asOfDate: "2026-07-11",
            equivalents: [],
          },
          positions: [],
          calculation: {
            methodologyVersion: "test",
            calculatedAt: "2026-07-11T00:00:00Z",
            inputsAsOf: "2026-07-11",
          },
        });
      }
      if (String(url).endsWith("/portfolios/portfolio-id/transactions") && init.method !== "POST") {
        return jsonResponse({ items: [], pagination: { nextCursor: null, hasMore: false, limit: 20 } });
      }
      if (String(url).endsWith("/portfolios/portfolio-id") && init.method !== "POST") {
        return jsonResponse({
          id: "portfolio-id",
          name: "Long-term RUB portfolio",
          baseCurrency: "RUB",
          version: 1,
          createdAt: "2026-07-11T00:00:00Z",
          updatedAt: "2026-07-11T00:00:00Z",
        });
      }
      if (String(url).endsWith("/portfolios") && init.method === "POST") {
        return jsonResponse({
          id: "portfolio-id",
          name: "Long-term RUB portfolio",
          baseCurrency: "RUB",
          version: 1,
          createdAt: "2026-07-11T00:00:00Z",
          updatedAt: "2026-07-11T00:00:00Z",
        }, 201);
      }
      if (String(url).endsWith("/transactions")) {
        return jsonResponse({
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
          tradeDate: "2026-07-11",
          settlementDate: null,
          revision: 1,
          createdAt: "2026-07-11T00:00:00Z",
          updatedAt: "2026-07-11T00:00:00Z",
        }, 201);
      }
      if (String(url).endsWith("/imports/review")) {
        return jsonResponse({
          portfolioId: "portfolio-id",
          sourceKind: "USER_UPLOADED_FILE",
          sourceAccountLabel: "Manual CSV import",
          sourceFileHash: "hash",
          retentionPolicy: "TRANSIENT_NOT_STORED",
          reviewGuarantee: "PREFLIGHT_ONLY_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS",
          summary: { totalRows: 0, appendableRows: 0, duplicateRows: 0, conflictRows: 0, invalidRows: 0 },
          rows: [],
        });
      }
      if (String(url).endsWith("/imports/append")) {
        return jsonResponse({
          portfolioId: "portfolio-id",
          sourceKind: "USER_UPLOADED_FILE",
          sourceFileHash: "hash",
          parsedRowCount: 1,
          acceptedRowCount: 1,
          nonAppendedRowCount: 0,
          appendedTransactionIds: ["transaction-id"],
          snapshotDatesRebuilt: ["2026-07-11"],
          auditActionCode: "IMPORT_APPEND_BATCH",
          nonSensitiveWarnings: [],
          appendValidationPolicy: "REVIEW_RERUN_AND_ATOMIC_STORE_REVALIDATION",
          rawPayloadRetentionRule: "RAW_CSV_NOT_STORED",
        });
      }
      return jsonResponse({ items: [], pagination: { nextCursor: null, hasMore: false, limit: 20 } });
    };

    await api.listPortfolios({ accessToken: "access-token" });
    await api.createPortfolio({ name: "Long-term RUB portfolio", baseCurrency: "RUB" }, { accessToken: "access-token" });
    await api.getPortfolio("portfolio-id", { accessToken: "access-token" });
    await api.getPortfolioSummary("portfolio-id", { accessToken: "access-token" });
    await api.listTransactions("portfolio-id", { accessToken: "access-token" });
    await api.appendTransaction("portfolio-id", {
      transactionType: "DEPOSIT",
      ticker: null,
      quantity: null,
      unitPrice: null,
      grossAmount: { amount: "100.00000000", currency: "RUB" },
      commission: { amount: "0.00000000", currency: "RUB" },
      tax: { amount: "0.00000000", currency: "RUB" },
      tradeDate: "2026-07-11",
      settlementDate: null,
    }, { accessToken: "access-token" });
    await api.reviewPortfolioImport("portfolio-id", { csvPayload: "type,amount\n" }, { accessToken: "access-token" });
    await api.appendReviewedPortfolioImport("portfolio-id", {
      csvPayload: "type,amount\n",
      decisions: [{ rowNumber: 1, action: "APPROVE" }],
    }, { accessToken: "access-token" });

    assert.equal(calls.length, 8);
    for (const call of calls) {
      assert.equal(call.init.headers.Authorization, "Bearer access-token");
      assert.equal(call.init.credentials, undefined);
    }
    assert.deepEqual(calls.map((call) => [call.init.method ?? "GET", call.url]), [
      ["GET", "http://localhost:8080/api/v1/portfolios"],
      ["POST", "http://localhost:8080/api/v1/portfolios"],
      ["GET", "http://localhost:8080/api/v1/portfolios/portfolio-id"],
      ["GET", "http://localhost:8080/api/v1/portfolios/portfolio-id/summary"],
      ["GET", "http://localhost:8080/api/v1/portfolios/portfolio-id/transactions"],
      ["POST", "http://localhost:8080/api/v1/portfolios/portfolio-id/transactions"],
      ["POST", "http://localhost:8080/api/v1/portfolios/portfolio-id/imports/review"],
      ["POST", "http://localhost:8080/api/v1/portfolios/portfolio-id/imports/append"],
    ]);
    assert.deepEqual(JSON.parse(calls[1].init.body), {
      name: "Long-term RUB portfolio",
      baseCurrency: "RUB",
    });
    assert.deepEqual(JSON.parse(calls[5].init.body), {
      transactionType: "DEPOSIT",
      ticker: null,
      quantity: null,
      unitPrice: null,
      grossAmount: { amount: "100.00000000", currency: "RUB" },
      commission: { amount: "0.00000000", currency: "RUB" },
      tax: { amount: "0.00000000", currency: "RUB" },
      tradeDate: "2026-07-11",
      settlementDate: null,
    });
    assert.deepEqual(JSON.parse(calls[6].init.body), { csvPayload: "type,amount\n" });
    assert.deepEqual(JSON.parse(calls[7].init.body), {
      csvPayload: "type,amount\n",
      decisions: [{ rowNumber: 1, action: "APPROVE" }],
    });
  } finally {
    Object.defineProperty(globalThis, "localStorage", { configurable: true, value: originalLocalStorage });
    Object.defineProperty(globalThis, "sessionStorage", { configurable: true, value: originalSessionStorage });
    Object.defineProperty(globalThis, "indexedDB", { configurable: true, value: originalIndexedDB });
    Object.defineProperty(globalThis, "document", { configurable: true, value: originalDocument });
  }
});
