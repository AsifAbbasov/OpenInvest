// @ts-nocheck

import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";

const dom = new JSDOM("<!doctype html><html><body></body></html>", { url: "http://localhost" });
for (const [name, value] of Object.entries({
  self: dom.window,
  window: dom.window,
  document: dom.window.document,
  navigator: dom.window.navigator,
  HTMLElement: dom.window.HTMLElement,
  HTMLInputElement: dom.window.HTMLInputElement,
  Event: dom.window.Event,
  DOMException: dom.window.DOMException,
})) {
  Object.defineProperty(globalThis, name, { configurable: true, writable: true, value });
}
Object.defineProperty(globalThis, "crypto", {
  configurable: true,
  writable: true,
  value: { randomUUID: (() => { let n = 0; return () => `10000000-0000-4000-8000-${String(++n).padStart(12, "0")}`; })() },
});
globalThis.IS_REACT_ACT_ENVIRONMENT = true;

const { act } = await import("react");
const { createRoot } = await import("react-dom/client");
const { DividendCalculator } = await import("../src/features/dividends/components/DividendCalculator.tsx");

function mount() {
  dom.window.document.body.innerHTML = '<div id="root"></div>';
  const container = dom.window.document.getElementById("root")!;
  return { container, root: createRoot(container) };
}

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
      requestId: "10000000-0000-4000-8000-000000000010",
      traceId: "10000000000000000000000000000010",
      generatedAt: "2026-09-05T00:00:00Z",
    },
  }), { status: 200, headers: { "Content-Type": "application/json" } });
}

async function submit(container: HTMLElement) {
  await act(async () => {
    container.querySelector("form")!.dispatchEvent(new dom.window.Event("submit", { bubbles: true, cancelable: true }));
    await Promise.resolve();
    await Promise.resolve();
  });
}

function setInput(input: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype, "value")!.set!;
  setter.call(input, value);
  input.dispatchEvent(new dom.window.Event("input", { bubbles: true }));
  input.dispatchEvent(new dom.window.Event("change", { bubbles: true }));
}

test("calculator renders backend result and never computes gross/yield locally", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  let requestBody;
  globalThis.fetch = async (_input, init) => {
    requestBody = JSON.parse(init.body);
    return successResponse();
  };
  const { container, root } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<DividendCalculator />));
  await submit(container);

  assert.deepEqual(requestBody, {
    ticker: "SBER",
    quantity: "1000.00000000",
    dividendPerUnit: { amount: "34.84000000", currency: "RUB" },
    positionCost: { amount: "280000.00000000", currency: "RUB" },
  });
  assert.match(container.textContent ?? "", /34840 ₽/);
  assert.match(container.textContent ?? "", /12\.442857%/);
  assert.match(container.textContent ?? "", /Gross · tax not included/);
});

test("uncertain transport retry reuses key; explicit input change releases it", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  const keys: string[] = [];
  let attempt = 0;
  globalThis.fetch = async (_input, init) => {
    keys.push(init.headers["Idempotency-Key"]);
    attempt += 1;
    if (attempt === 1) throw new TypeError("uncertain transport");
    return successResponse();
  };
  const { container, root } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<DividendCalculator />));
  await submit(container);
  await submit(container);
  assert.equal(keys[0], keys[1]);

  // The second attempt succeeded, so the next unchanged submission starts a new completed intent.
  await submit(container);
  assert.notEqual(keys[1], keys[2]);

  const quantityInput = [...container.querySelectorAll("input")][1] as HTMLInputElement;
  await act(async () => setInput(quantityInput, "999.00000000"));
  await submit(container);
  assert.notEqual(keys[2], keys[3]);
});

test("explicit HTTP failure also retains key for unchanged-payload retry", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  const keys: string[] = [];
  let attempt = 0;
  globalThis.fetch = async (_input, init) => {
    keys.push(init.headers["Idempotency-Key"]);
    attempt += 1;
    if (attempt === 1) {
      return new Response(JSON.stringify({
        error: { code: "IDEMPOTENCY_IN_FLIGHT", message: "Idempotency-Key is currently being processed" },
      }), { status: 409, headers: { "Content-Type": "application/json" } });
    }
    return successResponse();
  };
  const { container, root } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<DividendCalculator />));
  await submit(container);
  await submit(container);
  assert.equal(keys[0], keys[1]);
});

test("input change aborts obsolete request and stale completion cannot commit", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  let resolveFirst;
  let firstSignal: AbortSignal | undefined;
  globalThis.fetch = (_input, init) => new Promise((resolve) => {
    resolveFirst = resolve;
    firstSignal = init.signal;
    // Deliberately ignore abort so the request can complete stale. The component
    // must still reject the completion through its generation guard.
  });
  const { container, root } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<DividendCalculator />));
  await act(async () => {
    container.querySelector("form")!.dispatchEvent(new dom.window.Event("submit", { bubbles: true, cancelable: true }));
    await Promise.resolve();
  });
  const quantityInput = [...container.querySelectorAll("input")][1] as HTMLInputElement;
  await act(async () => setInput(quantityInput, "2000.00000000"));
  assert.equal(firstSignal?.aborted, true);
  await act(async () => {
    resolveFirst?.(successResponse());
    await Promise.resolve();
    await Promise.resolve();
  });
  assert.doesNotMatch(container.textContent ?? "", /34840 ₽/);
});

test("unmount aborts an active calculator request", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  let activeSignal: AbortSignal | undefined;
  globalThis.fetch = (_input, init) => new Promise(() => {
    activeSignal = init.signal;
  });
  const { container, root } = mount();
  t.after(() => {
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<DividendCalculator />));
  await act(async () => {
    container.querySelector("form")!.dispatchEvent(new dom.window.Event("submit", { bubbles: true, cancelable: true }));
    await Promise.resolve();
  });
  assert.equal(activeSignal?.aborted, false);

  await act(async () => root.unmount());
  assert.equal(activeSignal?.aborted, true);
});

