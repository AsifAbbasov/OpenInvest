// @ts-nocheck

import assert from "node:assert/strict";
import test from "node:test";

import { act } from "react";
import { createRoot } from "react-dom/client";
import { JSDOM } from "jsdom";

import { CorporateActionsSlice } from "../src/features/corporate-actions/components/CorporateActionsSlice.tsx";

globalThis.IS_REACT_ACT_ENVIRONMENT = true;

function mountSlice() {
  const dom = new JSDOM("<!doctype html><html><body><div id=\"root\"></div></body></html>", { url: "http://localhost" });
  for (const [name, value] of Object.entries({
    window: dom.window,
    document: dom.window.document,
    navigator: dom.window.navigator,
    HTMLElement: dom.window.HTMLElement,
    HTMLInputElement: dom.window.HTMLInputElement,
    Event: dom.window.Event,
  })) {
    Object.defineProperty(globalThis, name, { configurable: true, writable: true, value });
  }
  const container = dom.window.document.getElementById("root")!;
  return { dom, container, root: createRoot(container) };
}

function setInput(container: HTMLElement, selector: string, value: string) {
  const input = container.querySelector(selector) as HTMLInputElement;
  const setter = Object.getOwnPropertyDescriptor(input.ownerDocument.defaultView!.HTMLInputElement.prototype, "value")!.set!;
  setter.call(input, value);
  input.dispatchEvent(new Event("input", { bubbles: true }));
  input.dispatchEvent(new Event("change", { bubbles: true }));
}

async function submit(container: HTMLElement) {
  await act(async () => {
    container.querySelector("form")!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await Promise.resolve();
    await Promise.resolve();
  });
}

function errorResponse(status: number, code: string, message: string) {
  return new Response(JSON.stringify({
    error: { code, message, details: [] },
    meta: {
      requestId: "10000000-0000-4000-8000-000000000001",
      traceId: "10000000000000000000000000000001",
      generatedAt: "2026-09-05T00:00:00Z",
    },
  }), { status, headers: { "Content-Type": "application/json" } });
}


function deferred<T>() {
  let resolve: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve: resolve! };
}

function projectionResponse(calendar = [], heatmap = []) {
  return new Response(JSON.stringify({
    data: {
      calendar,
      heatmap,
      coverage: { inputMode: "PROVIDER", instrumentIds: ["SBER"], from: "2026-01-01", to: "2026-12-31" },
    },
    meta: {
      requestId: "10000000-0000-4000-8000-000000000002",
      traceId: "10000000000000000000000000000002",
      generatedAt: "2026-09-05T00:00:00Z",
    },
  }), { status: 200, headers: { "Content-Type": "application/json" } });
}

async function prepareForm(container: HTMLElement) {
  await act(async () => {
    setInput(container, 'input[placeholder="SBER, GAZP"]', "SBER");
    const dates = [...container.querySelectorAll('input[type="date"]')];
    setInput(container, 'input[type="date"]:first-of-type', "2026-01-01");
    const second = dates[1] as HTMLInputElement;
    const setter = Object.getOwnPropertyDescriptor(second.ownerDocument.defaultView!.HTMLInputElement.prototype, "value")!.set!;
    setter.call(second, "2026-12-31");
    second.dispatchEvent(new Event("input", { bubbles: true }));
    second.dispatchEvent(new Event("change", { bubbles: true }));
  });
}

test("503 source unavailable is not rendered as zero events", async (t) => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => errorResponse(503, "CORPORATE_ACTIONS_SOURCE_UNAVAILABLE", "Corporate actions source is unavailable");
  const { dom, container, root } = mountSlice();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<CorporateActionsSlice />));
  await prepareForm(container);
  await submit(container);

  assert.match(container.textContent ?? "", /Corporate actions source unavailable/);
  assert.match(container.textContent ?? "", /not shown as “zero events”/);
  assert.doesNotMatch(container.textContent ?? "", /legitimate empty projection/);
});

test("successful empty provider response renders a legitimate empty state", async (t) => {
  const originalFetch = globalThis.fetch;
  let requestedURL = "";
  globalThis.fetch = async (input: RequestInfo | URL) => {
    requestedURL = String(input);
    return projectionResponse();
  };
  const { dom, container, root } = mountSlice();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<CorporateActionsSlice />));
  await prepareForm(container);
  await act(async () => {
    setInput(container, 'input[placeholder="SBER, GAZP"]', "SBER GAZP SBER");
    await Promise.resolve();
  });
  await submit(container);

  assert.equal(new URL(requestedURL).searchParams.get("instrumentId"), "SBER,GAZP");
  assert.match(container.textContent ?? "", /No current dated events/);
  assert.match(container.textContent ?? "", /legitimate empty projection/);
});

test("more than fifty unique instruments is rejected before fetch", async (t) => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    return projectionResponse();
  };
  const { dom, container, root } = mountSlice();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<CorporateActionsSlice />));
  await prepareForm(container);
  await act(async () => {
    setInput(container, 'input[placeholder="SBER, GAZP"]', Array.from({ length: 51 }, (_, i) => `I${String(i).padStart(2, "0")}`).join(" "));
    await Promise.resolve();
  });
  await submit(container);

  assert.equal(calls, 0);
  assert.match(container.textContent ?? "", /Use at most 50 instruments/);
});

test("populated projection preserves lifecycle status and count-only heatmap", async (t) => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => projectionResponse(
    [{
      effectiveDate: "2026-10-10",
      event: {
        eventId: "evt-1",
        instrumentId: "SBER",
        kind: "DIVIDEND",
        status: "ANNOUNCED",
        recordDate: "2026-10-10",
        paymentDate: "2026-10-20",
        amountPerUnit: { amount: "12.34000000", currency: "RUB" },
        supersedesEventId: null,
        asOf: "2026-09-05T00:00:00Z",
        retrievedAt: "2026-09-05T00:01:00Z",
        provenance: { provider: "FIXTURE" },
      },
    }],
    [{ date: "2026-10-10", totalCount: 1, dividendCount: 1, couponCount: 0, announcedCount: 1, confirmedCount: 0, paidCount: 0, cancelledCount: 0 }],
  );
  const { dom, container, root } = mountSlice();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<CorporateActionsSlice />));
  await prepareForm(container);
  await submit(container);

  assert.match(container.textContent ?? "", /ANNOUNCED/);
  assert.match(container.textContent ?? "", /not guaranteed income/);
  assert.match(container.textContent ?? "", /Counts only/);
  assert.match(container.textContent ?? "", /Source: FIXTURE/);
  assert.match(container.textContent ?? "", /Retrieved: 2026-09-05T00:01:00Z/);
  assert.equal(container.querySelector('[data-density-level="4"]')?.textContent, "1");
});


test("stale request completion cannot overwrite a newer corporate-action query", async (t) => {
  const originalFetch = globalThis.fetch;
  const first = deferred<Response>();
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    if (calls === 1) return first.promise;
    return projectionResponse([], []);
  };
  const { dom, container, root } = mountSlice();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<CorporateActionsSlice />));
  await prepareForm(container);
  await submit(container);

  await act(async () => {
    setInput(container, 'input[placeholder="SBER, GAZP"]', "GAZP");
    await Promise.resolve();
  });
  await submit(container);
  assert.match(container.textContent ?? "", /legitimate empty projection/);

  await act(async () => {
    first.resolve(errorResponse(503, "CORPORATE_ACTIONS_SOURCE_UNAVAILABLE", "Corporate actions source is unavailable"));
    await first.promise;
    await Promise.resolve();
  });

  assert.match(container.textContent ?? "", /legitimate empty projection/);
  assert.doesNotMatch(container.textContent ?? "", /Corporate actions source unavailable/);
});
