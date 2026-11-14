import assert from "node:assert/strict";
import { afterEach, test } from "node:test";

// Place this file under frontend-next/tests in an isolated worktree after applying the patch.
const api = await import("../src/common/api/openinvest.ts");
const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

const REQUEST_ID = "11111111-1111-4111-8111-111111111111";
const TRACE_ID = "22222222222222222222222222222222";
const META = { requestId: REQUEST_ID, traceId: TRACE_ID, generatedAt: "2026-11-11T00:00:00Z" };
const KEY = "33333333-3333-4333-8333-333333333333";
const portfolioPayload = { name: "Test portfolio", baseCurrency: "RUB" };
const auth = { accessToken: "test-token", idempotencyKey: KEY };

function wire(value, status = 200, headers = {}) {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json", ...headers },
  });
}
function success() { return wire({ data: { id: "portfolio-test-id" }, meta: META }, 201); }
function failure(status, code, retryAfter = null) {
  return wire(
    { error: { code, message: `Expected ${code}`, details: [] }, meta: META }, status,
    retryAfter === null ? {} : { "Retry-After": String(retryAfter) },
  );
}
const fetchGet = () => api.listPortfolios({ accessToken: "test-token" });
const fetchWrite = () => api.createPortfolio(portfolioPayload, auth);

test("valid success requires and preserves OpenAPI response meta", async () => {
  globalThis.fetch = async () => success();
  const result = await fetchWrite();
  assert.equal(result.ok, true);
  assert.equal(result.requestId, REQUEST_ID);
  assert.equal(result.traceId, TRACE_ID);
});

test("explicit HTTP statuses retain stable error code and trace identifiers", async () => {
  const entries = [
    [401, "UNAUTHORIZED", "AUTHENTICATION"],
    [403, "FORBIDDEN", "AUTHORIZATION"],
    [409, "IDEMPOTENCY_CONFLICT", "CONFLICT"],
    [422, "VALIDATION_ERROR", "VALIDATION"],
    [429, "RATE_LIMITED", "RATE_LIMITED"],
    [503, "SERVICE_UNAVAILABLE", "SERVER"],
  ];
  for (const [status, code, kind] of entries) {
    globalThis.fetch = async () => failure(status, code, status === 429 ? 13 : null);
    const result = await fetchWrite();
    assert.equal(result.ok, false);
    assert.equal(result.status, status);
    assert.equal(result.code, code);
    assert.equal(result.failureKind, kind);
    assert.equal(result.requestId, REQUEST_ID);
    assert.equal(result.traceId, TRACE_ID);
    assert.equal(result.retryAfterSeconds, status === 429 ? 13 : undefined);
    assert.equal(result.completion, status >= 500 ? "UNKNOWN" : "REJECTED");
  }
});

test("malformed success meta is a contract failure, not a network failure", async () => {
  const cases = [
    { data: { id: "ok" } },
    { data: { id: "ok" }, meta: { requestId: REQUEST_ID } },
    { data: { id: "ok" }, meta: { ...META, traceId: "bad" } },
    { meta: META },
    null,
    [],
  ];
  for (const body of cases) {
    globalThis.fetch = async () => wire(body);
    const result = await fetchGet();
    assert.equal(result.ok, false);
    assert.equal(result.failureKind, "INVALID_RESPONSE");
    assert.equal(result.status, 200);
  }
});

test("invalid JSON, non-JSON and empty replies preserve received status", async () => {
  const responses = [
    new Response("{broken json", {status: 503, headers: {"Content-Type": "application/json"}}),
    new Response("gateway down", {status: 503, headers: {"Content-Type": "text/plain", "X-Request-ID": REQUEST_ID}}),
    new Response(null, {status: 204}),
  ];
  for (const response of responses) {
    globalThis.fetch = async () => response;
    const result = await fetchGet();
    assert.equal(result.ok, false);
    assert.equal(result.failureKind, "INVALID_RESPONSE");
    assert.equal(result.status, response.status);
  }
});

test("invalid error envelope is not accepted as a validated server error", async () => {
  globalThis.fetch = async () => wire({ error: {code: "CONFLICT", message: "Missing details"}, meta: META }, 409);
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.failureKind, "INVALID_RESPONSE");
  assert.equal(result.status, 409);
});

test("network rejection of a financial write stays explicitly uncertain", async () => {
  let sent = 0;
  globalThis.fetch = async (_url, init) => {
    sent += 1;
    assert.equal(init.headers["Idempotency-Key"], KEY);
    throw new TypeError("request failed");
  };
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.failureKind, "NETWORK");
  assert.equal(result.completion, "UNKNOWN");
  assert.equal(sent, 1, "financial mutations must not automatically retry");
});

test("explicit cancellation remains distinct from network failure", async () => {
  globalThis.fetch = async () => { throw new DOMException("aborted", "AbortError"); };
  const result = await fetchGet();
  assert.equal(result.ok, false);
  assert.equal(result.failureKind, "CANCELLED");
});

test("explicit timeout remains distinct from network failure", async () => {
  globalThis.fetch = async () => { throw new DOMException("timeout", "TimeoutError"); };
  const result = await fetchGet();
  assert.equal(result.ok, false);
  assert.equal(result.failureKind, "TIMEOUT");
});

test("missing idempotency key refuses write before any network call", async () => {
  let sent = 0;
  globalThis.fetch = async () => { sent += 1; return success(); };
  const missing = await api.createPortfolio(portfolioPayload, {accessToken: "test-token"});
  assert.equal(missing.ok, false);
  assert.equal(missing.failureKind, "REQUEST_SETUP");
  assert.equal(missing.code, "CLIENT_IDEMPOTENCY_KEY_REQUIRED");
  assert.equal(sent, 0);
});

test("lost response after a committed write recovers with the same explicit key", async () => {
  const ledger = new Map();
  const seen = [];
  globalThis.fetch = async (_url, init) => {
    const key = init.headers["Idempotency-Key"];
    seen.push(key);
    if (!ledger.has(key)) {
      ledger.set(key, success());
      throw new TypeError("connection lost after server commit");
    }
    return success();
  };
  const first = await fetchWrite();
  assert.equal(first.ok, false);
  assert.equal(first.completion, "UNKNOWN");
  const second = await fetchWrite();
  assert.equal(second.ok, true);
  assert.deepEqual(seen, [KEY, KEY]);
  assert.equal(ledger.size, 1);
});


test("abort during response body read is not mistaken for malformed JSON", async () => {
  globalThis.fetch = async () => ({
    status: 200,
    ok: true,
    headers: new Headers({"Content-Type": "application/json", "X-Request-ID": REQUEST_ID}),
    text: async () => { throw new DOMException("aborted", "AbortError"); },
  });
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.failureKind, "CANCELLED");
  assert.equal(result.completion, "UNKNOWN");
  assert.equal(result.status, 200);
  assert.equal(result.requestId, REQUEST_ID);
});

test("409 IDEMPOTENCY_IN_FLIGHT keeps the original financial command uncertain", async () => {
  let calls = 0;
  globalThis.fetch = async (_url, init) => {
    calls += 1;
    assert.equal(init.headers["Idempotency-Key"], KEY);
    return failure(409, "IDEMPOTENCY_IN_FLIGHT");
  };
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.status, 409);
  assert.equal(result.code, "IDEMPOTENCY_IN_FLIGHT");
  assert.equal(result.failureKind, "CONFLICT");
  assert.equal(result.completion, "UNKNOWN");
  assert.equal(calls, 1, "financial mutation must never automatically retry");
});

test("timeout during response body read is classified separately from invalid JSON", async () => {
  globalThis.fetch = async () => ({
    status: 200,
    ok: true,
    headers: new Headers({"Content-Type": "application/json", "X-Request-ID": REQUEST_ID, "X-Trace-ID": TRACE_ID}),
    text: async () => { throw new DOMException("timeout", "TimeoutError"); },
  });
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.status, 200);
  assert.equal(result.requestId, REQUEST_ID);
  assert.equal(result.traceId, TRACE_ID);
  assert.equal(result.failureKind, "TIMEOUT");
  assert.equal(result.completion, "UNKNOWN");
});

test("read interruption after HTTP headers preserves uncertain write outcome", async () => {
  globalThis.fetch = async () => ({
    status: 201,
    ok: true,
    headers: new Headers({"Content-Type": "application/json", "X-Request-ID": REQUEST_ID, "X-Trace-ID": TRACE_ID}),
    text: async () => { throw new TypeError("stream interrupted"); },
  });
  const result = await fetchWrite();
  assert.equal(result.ok, false);
  assert.equal(result.status, 201);
  assert.equal(result.failureKind, "NETWORK");
  assert.equal(result.completion, "UNKNOWN");
});
