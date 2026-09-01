// @ts-nocheck

import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { JSDOM } from "jsdom";

// React DOM's change-event feature detection runs when react-dom/client is imported.
// Install one stable JSDOM environment first so controlled input onChange events behave
// the same way for every test in this file.
const dom = new JSDOM("<!doctype html><html><body></body></html>", { url: "http://localhost" });
for (const [name, value] of Object.entries({
  window: dom.window,
  document: dom.window.document,
  navigator: dom.window.navigator,
  HTMLElement: dom.window.HTMLElement,
  HTMLInputElement: dom.window.HTMLInputElement,
  HTMLSelectElement: dom.window.HTMLSelectElement,
  HTMLTextAreaElement: dom.window.HTMLTextAreaElement,
  Event: dom.window.Event,
  InputEvent: dom.window.InputEvent,
  MouseEvent: dom.window.MouseEvent,
})) {
  Object.defineProperty(globalThis, name, { configurable: true, writable: true, value });
}

globalThis.IS_REACT_ACT_ENVIRONMENT = true;

const { act } = await import("react");
const { createRoot } = await import("react-dom/client");
const { AddTransactionForm } = await import(
  "../src/features/portfolio/components/AddTransactionForm.tsx"
);

const principalId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const portfolioId = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";

function apiResponse(data: unknown, status = 200) {
  return new Response(JSON.stringify({
    data,
    meta: {
      requestId: "10000000-0000-4000-8000-000000000010",
      traceId: "10000000000000000000000000000010",
      generatedAt: "2026-09-01T00:00:00Z",
    },
  }), { status, headers: { "Content-Type": "application/json" } });
}

function apiError(message: string, status = 500) {
  return new Response(JSON.stringify({
    error: {
      code: "TEST_ERROR",
      message,
    },
  }), { status, headers: { "Content-Type": "application/json" } });
}

function mountForm() {
  dom.window.sessionStorage.clear();
  dom.window.document.body.innerHTML = '<div id="root"></div>';
  const container = dom.window.document.getElementById("root")!;
  const root = createRoot(container);
  return { container, root };
}

function controlFor(container: HTMLElement, labelText: string) {
  const label = [...container.querySelectorAll("label")].find((candidate) =>
    (candidate.textContent ?? "").trim().startsWith(labelText),
  );
  assert.ok(label, `missing label: ${labelText}`);
  const control = label.querySelector("input, select, textarea");
  assert.ok(control, `missing control for: ${labelText}`);
  return control as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
}

async function setControl(container: HTMLElement, labelText: string, value: string) {
  const control = controlFor(container, labelText);
  const setter = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(control), "value")?.set;
  assert.ok(setter, `missing native value setter for: ${labelText}`);

  await act(async () => {
    setter.call(control, value);
    control.dispatchEvent(new dom.window.Event("input", { bubbles: true }));
    control.dispatchEvent(new dom.window.Event("change", { bubbles: true }));
    await new Promise((resolve) => setTimeout(resolve, 0));
  });

  assert.equal(control.value, value, `${labelText} DOM value did not update`);
}

async function renderForm(root: ReturnType<typeof createRoot>, onSaved = () => {}) {
  await act(async () => {
    root.render(
      <AddTransactionForm
        accessToken="access-token"
        principalId={principalId}
        portfolioId={portfolioId}
        onSaved={onSaved}
      />,
    );
  });
}

async function submit(container: HTMLElement) {
  const previousStatus = container.querySelector(".form-status")?.textContent ?? null;

  await act(async () => {
    container.querySelector("form")!.dispatchEvent(
      new dom.window.Event("submit", { bubbles: true, cancelable: true }),
    );
  });

  for (let attempt = 0; attempt < 200; attempt += 1) {
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 5));
    });
    const currentStatus = container.querySelector(".form-status")?.textContent ?? null;
    if (currentStatus !== null && currentStatus !== previousStatus) {
      return;
    }
  }

  assert.fail("form submit did not settle with a new status");
}

function requestBody(init?: RequestInit) {
  assert.equal(typeof init?.body, "string");
  return JSON.parse(init!.body as string);
}

function idempotencyKey(init?: RequestInit) {
  return new Headers(init?.headers).get("Idempotency-Key");
}

test("initial production form has no fixture-derived business values", { concurrency: false }, async (t) => {
  const { container, root } = mountForm();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.document.body.innerHTML = "";
  });

  await renderForm(root);

  assert.equal((controlFor(container, "Transaction type") as HTMLSelectElement).value, "BUY");
  for (const label of ["Ticker", "Quantity", "Unit price", "Commission", "Tax", "Trade date", "Settlement date", "Note"]) {
    assert.equal(controlFor(container, label).value, "", `${label} must start empty`);
  }

  await setControl(container, "Transaction type", "DEPOSIT");
  assert.equal(controlFor(container, "Gross amount").value, "", "Gross amount must start empty");

  const source = await readFile(
    new URL("../src/features/portfolio/components/AddTransactionForm.tsx", import.meta.url),
    "utf8",
  );
  for (const fixture of [
    "\"SBER\"",
    "\"10.00000000\"",
    "\"280.00000000\"",
    "\"2800.00000000\"",
    "\"2.80000000\"",
    "\"0.00000000\"",
    "\"2026-01-10\"",
    "\"2026-01-13\"",
    "\"Stage 3.3 Web presentation slice\"",
  ]) {
    assert.equal(source.includes(fixture), false, `production source retains fixture literal ${fixture}`);
  }
  assert.match(source, /useState<TransactionType>\("BUY"\)/);
  assert.match(source, /currency: "RUB"/);
});

test("BUY submission uses only user-entered values and preserves null semantics", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ input: RequestInfo | URL; init?: RequestInit }> = [];
  let saved = 0;
  globalThis.fetch = async (input, init) => {
    requests.push({ input, init });
    return apiResponse({ id: "10000000-0000-4000-8000-000000000011" });
  };
  const { container, root } = mountForm();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.document.body.innerHTML = "";
    globalThis.fetch = originalFetch;
  });

  await renderForm(root, () => { saved += 1; });
  await setControl(container, "Ticker", "sber");
  await setControl(container, "Quantity", "3.00000000");
  await setControl(container, "Unit price", "301.25000000");
  await setControl(container, "Commission", "1.25000000");
  await setControl(container, "Tax", "0.00000000");
  await setControl(container, "Trade date", "2026-09-01");
  await setControl(container, "Settlement date", "");
  await setControl(container, "Note", "  user supplied note  ");
  await submit(container);

  assert.equal(requests.length, 1);
  assert.equal(saved, 1);
  const payload = requestBody(requests[0].init);
  assert.deepEqual(payload, {
    transactionType: "BUY",
    ticker: "SBER",
    quantity: "3.00000000",
    unitPrice: { amount: "301.25000000", currency: "RUB" },
    grossAmount: null,
    commission: { amount: "1.25000000", currency: "RUB" },
    tax: { amount: "0.00000000", currency: "RUB" },
    tradeDate: "2026-09-01",
    settlementDate: null,
    note: "user supplied note",
  });
  assert.match(idempotencyKey(requests[0].init) ?? "", /^[A-Za-z0-9._:-]{16,128}$/);
});

test("type switch prevents stale trade values from leaking into cash-flow payload", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ input: RequestInfo | URL; init?: RequestInit }> = [];
  globalThis.fetch = async (input, init) => {
    requests.push({ input, init });
    return apiResponse({ id: "10000000-0000-4000-8000-000000000012" });
  };
  const { container, root } = mountForm();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.document.body.innerHTML = "";
    globalThis.fetch = originalFetch;
  });

  await renderForm(root);
  await setControl(container, "Ticker", "GAZP");
  await setControl(container, "Quantity", "7.00000000");
  await setControl(container, "Unit price", "150.00000000");

  await setControl(container, "Transaction type", "DEPOSIT");
  await setControl(container, "Gross amount", "1000.00000000");
  await setControl(container, "Commission", "0.00000000");
  await setControl(container, "Tax", "0.00000000");
  await setControl(container, "Trade date", "2026-09-01");
  await setControl(container, "Settlement date", "");
  await setControl(container, "Note", "");
  await submit(container);

  assert.equal(requests.length, 1);
  const payload = requestBody(requests[0].init);
  assert.equal(payload.transactionType, "DEPOSIT");
  assert.equal(payload.ticker, null);
  assert.equal(payload.quantity, null);
  assert.equal(payload.unitPrice, null);
  assert.deepEqual(payload.grossAmount, { amount: "1000.00000000", currency: "RUB" });
  assert.equal(payload.settlementDate, null);
  assert.equal(payload.note, null);
});

test("same failed intent reuses the same Idempotency-Key on retry", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ input: RequestInfo | URL; init?: RequestInit }> = [];
  let call = 0;
  globalThis.fetch = async (input, init) => {
    requests.push({ input, init });
    call += 1;
    if (call === 1) {
      return apiError("temporary failure", 500);
    }
    return apiResponse({ id: "10000000-0000-4000-8000-000000000013" });
  };
  const { container, root } = mountForm();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.document.body.innerHTML = "";
    globalThis.fetch = originalFetch;
  });

  await renderForm(root);
  await setControl(container, "Ticker", "LKOH");
  await setControl(container, "Quantity", "1.00000000");
  await setControl(container, "Unit price", "7000.00000000");
  await setControl(container, "Commission", "5.00000000");
  await setControl(container, "Tax", "0.00000000");
  await setControl(container, "Trade date", "2026-09-01");

  await submit(container);
  await submit(container);

  assert.equal(requests.length, 2);
  const firstKey = idempotencyKey(requests[0].init);
  const secondKey = idempotencyKey(requests[1].init);
  assert.match(firstKey ?? "", /^[A-Za-z0-9._:-]{16,128}$/);
  assert.equal(secondKey, firstKey);
});

test("ill-formed Unicode note is rejected before any request", { concurrency: false }, async (t) => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    return apiResponse({});
  };
  const { container, root } = mountForm();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.document.body.innerHTML = "";
    globalThis.fetch = originalFetch;
  });

  await renderForm(root);
  await setControl(container, "Ticker", "SBER");
  await setControl(container, "Quantity", "1.00000000");
  await setControl(container, "Unit price", "300.00000000");
  await setControl(container, "Commission", "0.00000000");
  await setControl(container, "Tax", "0.00000000");
  await setControl(container, "Trade date", "2026-09-01");
  await setControl(container, "Note", "\ud800");
  await submit(container);

  assert.equal(calls, 0);
  assert.match(container.textContent ?? "", /Transaction note contains invalid Unicode/);
});
