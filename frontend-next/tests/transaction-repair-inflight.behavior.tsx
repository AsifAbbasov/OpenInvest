// @ts-nocheck
// Behavioral test: mounts the REAL TransactionRepairControls in React and executes
// correction/reversal through the REAL API client and idempotency journal.
// Only useAuth is replaced with a fixed authenticated principal through the
// adjacent test-only Node loader. There is no backend or production network.
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";

const dom = new JSDOM('<!doctype html><html><body><div id="root"></div></body></html>', {
  url: "https://localhost/portfolio/test",
});
for (const [name, value] of Object.entries({
  window: dom.window,
  document: dom.window.document,
  navigator: dom.window.navigator,
  HTMLElement: dom.window.HTMLElement,
  HTMLInputElement: dom.window.HTMLInputElement,
  Event: dom.window.Event,
  MouseEvent: dom.window.MouseEvent,
})) {
  Object.defineProperty(globalThis, name, {configurable:true,writable:true,value});
}
Object.defineProperty(globalThis, "IS_REACT_ACT_ENVIRONMENT", {
  configurable:true,writable:true,value:true,
});

const React = await import("react");
const {createRoot} = await import("react-dom/client");
const {TransactionRepairControls} = await import("../src/features/portfolio/components/TransactionRepairControls.tsx");

const meta = {
  requestId: "11111111-1111-4111-8111-111111111111",
  traceId: "22222222222222222222222222222222",
  generatedAt: "2026-11-11T00:00:00Z",
};
const transaction = {
  id: "33333333-3333-4333-8333-333333333333",
  portfolioId: "44444444-4444-4444-8444-444444444444",
  transactionType: "DEPOSIT",
  status: "ACTIVE",
  ticker: null, quantity: null, unitPrice: null,
  grossAmount: {amount: "100.00000000", currency: "RUB"},
  commission: {amount: "0.00000000", currency: "RUB"},
  tax: {amount: "0.00000000", currency: "RUB"},
  tradeDate: "2026-11-09", settlementDate: null,
  revision: 1,
  createdAt: "2026-11-09T00:00:00Z",
  updatedAt: "2026-11-09T00:00:00Z",
};
function apiResponse(value, status) {
  return new Response(JSON.stringify(value), {status, headers: {"Content-Type":"application/json"}});
}
function pending() {
  return apiResponse({error:{code:"IDEMPOTENCY_IN_FLIGHT",message:"Idempotency-Key is currently being processed",details:[]},meta},409);
}
function completed(operation) {
  return apiResponse({data: operation === "correction" ? {...transaction,revision:2} : {effectiveDate:"2026-11-09"},meta},200);
}
async function click(node) {
  assert.ok(node,"expected button missing");
  await React.act(async()=>{
    node.dispatchEvent(new dom.window.MouseEvent("click",{bubbles:true,cancelable:true}));
    await Promise.resolve();
  });
}
function button(root,text) {
  return [...root.querySelectorAll("button")].find(x=>x.textContent?.trim()===text);
}
async function reasonInput(root,value) {
  const input = [...root.querySelectorAll("input")].find(x=>x.closest("label")?.textContent?.includes("Reason"));
  assert.ok(input,"reason input missing");
  const setter=Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype,"value").set;
  await React.act(async()=>{
    setter.call(input,value);
    input.dispatchEvent(new dom.window.Event("input",{bubbles:true}));input.dispatchEvent(new dom.window.Event("change",{bubbles:true}));
    await Promise.resolve();
  });
}


function observeNextRequest() {
  let resolve;
  const promise = new Promise(r => { resolve = r; });
  return { promise, resolve };
}

async function awaitObserved(promise, description) {
  let timer;
  try {
    await Promise.race([
      promise,
      new Promise((_, reject) => {
        timer = setTimeout(
          () => reject(new Error(`Timed out awaiting ${description}`)),
          5000
        );
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}

async function waitForStatus(container, pattern) {
  if (pattern.test(container.textContent ?? "")) return;

  await new Promise((resolve, reject) => {
    let timer;
    const observer = new dom.window.MutationObserver(() => {
      if (pattern.test(container.textContent ?? "")) {
        clearTimeout(timer);
        observer.disconnect();
        resolve();
      }
    });

    observer.observe(container, {
      childList: true,
      characterData: true,
      subtree: true,
    });

    timer = setTimeout(() => {
      observer.disconnect();
      reject(new Error(
        `Expected status ${pattern}, received: ${container.textContent}`
      ));
    }, 5000);
  });
}

function storedRetryKeys() {
  return Array.from(
    { length: dom.window.sessionStorage.length },
    (_, index) => dom.window.sessionStorage.key(index)
  ).filter(key => key?.startsWith("oi:idempotency-retry:v1:"));
}

for (const operation of ["correction", "reversal"]) {
  test(`real React ${operation} preserves one key after 409 IDEMPOTENCY_IN_FLIGHT and retries with it`, async () => {
    dom.window.sessionStorage.clear();

    const current = dom.window.document.getElementById("root");
    current.replaceChildren();

    const root = createRoot(current);
    const requests = [];

    let mutated = 0;
    let nextRequest = null;
    let mutationCompleted = null;

    const originalFetch = globalThis.fetch;

    globalThis.fetch = async (url, options) => {
      requests.push({
        url: String(url),
        method: options?.method,
        key: new Headers(options?.headers).get("Idempotency-Key"),
      });

      nextRequest?.resolve();

      return requests.length === 1
        ? pending()
        : completed(operation);
    };

    async function performCommand() {
      const observed = observeNextRequest();
      nextRequest = observed;

      await click(button(current,
        operation === "correction"
          ? "Save correction"
          : "Confirm reversal"
      ));

      await awaitObserved(observed.promise, "HTTP request");

      nextRequest = null;

      await React.act(async () => {
        await Promise.resolve();
      });
    }

    async function openAndFill() {
      await click(button(current,
        operation === "correction" ? "Edit" : "Reverse"
      ));

      await reasonInput(current, "Synthetic controlled retry");

      const reason = [...current.querySelectorAll("input")]
        .find(input => input.closest("label")?.textContent?.includes("Reason"));

      assert.equal(reason?.value, "Synthetic controlled retry");
    }

    try {
      await React.act(async () => {
        root.render(React.createElement(TransactionRepairControls, {
          accessToken: "synthetic-token",
          portfolioId: transaction.portfolioId,
          transaction,
          onMutated: async () => {
            mutated++;
            mutationCompleted?.resolve();
          },
        }));
      });

      await openAndFill();
      await performCommand();

      assert.equal(requests.length, 1, "one initial command required");
      assert.match(requests[0].key ?? "", /^[A-Za-z0-9._:-]{16,128}$/);
      assert.equal(mutated, 0);

      await waitForStatus(current, /currently being processed/);

      assert.equal(
        storedRetryKeys().length,
        1,
        "409 IN_FLIGHT must preserve the browser retry journal"
      );

      const saved = JSON.parse(
        dom.window.sessionStorage.getItem(storedRetryKeys()[0])
      );

      assert.equal(saved.key, requests[0].key);

      const firstSuccess = observeNextRequest();
      mutationCompleted = firstSuccess;

      await performCommand();

      await awaitObserved(firstSuccess.promise, "confirmed mutation");

      assert.equal(requests.length, 2);
      assert.equal(
        requests[1].key,
        requests[0].key,
        "explicit retry must preserve command identity"
      );

      assert.equal(
        requests[0].method,
        operation === "correction" ? "PATCH" : "DELETE"
      );

      assert.equal(mutated, 1);
      assert.equal(
        storedRetryKeys().length,
        0,
        "confirmed success must clear the browser retry journal"
      );

      // A separate user action, even with an identical payload,
      // must receive a fresh financial command identity.
      await openAndFill();

      const secondSuccess = observeNextRequest();
      mutationCompleted = secondSuccess;

      await performCommand();
      await awaitObserved(secondSuccess.promise, "second mutation");

      assert.equal(requests.length, 3);
      assert.notEqual(
        requests[2].key,
        requests[0].key,
        "new user command must receive a new Idempotency-Key"
      );
      assert.equal(mutated, 2);
      assert.equal(storedRetryKeys().length, 0);
    } finally {
      globalThis.fetch = originalFetch;
      await React.act(async () => root.unmount());
    }
  });
}
