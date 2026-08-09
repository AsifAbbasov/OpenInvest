// @ts-nocheck

import assert from "node:assert/strict";
import test from "node:test";

import { act } from "react";
import { createRoot } from "react-dom/client";
import { JSDOM } from "jsdom";

import { ImportUploadReviewPanel } from "../src/features/portfolio/components/ImportUploadReviewPanel.tsx";

globalThis.IS_REACT_ACT_ENVIRONMENT = true;

const importCSV = "transaction_type,gross_amount,commission,tax,trade_date,currency\nDEPOSIT,100.00000000,0,0,2026-08-08,RUB\n";

function apiResponse(data: unknown) {
  return new Response(JSON.stringify({
    data,
    meta: {
      requestId: "10000000-0000-4000-8000-000000000010",
      traceId: "10000000000000000000000000000010",
      generatedAt: "2026-08-08T00:00:00Z",
    },
  }), { status: 200, headers: { "Content-Type": "application/json" } });
}

function deferred<T>() {
  let resolve: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve: resolve! };
}

function reviewData(portfolioId: string) {
  return {
    portfolioId,
    sourceKind: "USER_UPLOADED_FILE",
    sourceAccountLabel: "Manual CSV import",
    sourceFileHash: "a".repeat(64),
    reviewToken: "review-token-32-characters-minimum-value",
    retentionPolicy: "TRANSIENT_NOT_STORED",
    reviewGuarantee: "SIGNED_REVIEW_TOKEN_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS",
    summary: { totalRows: 1, appendableRows: 1, duplicateRows: 0, conflictRows: 0, invalidRows: 0 },
    rows: [{
      rowNumber: 2,
      rowHash: "b".repeat(64),
      status: "APPENDABLE",
      reasonCodes: [],
      candidate: {
        transactionType: "DEPOSIT",
        grossAmount: { amount: "100.00000000", currency: "RUB" },
        commission: { amount: "0.00000000", currency: "RUB" },
        tax: { amount: "0.00000000", currency: "RUB" },
        tradeDate: "2026-08-08",
      },
    }],
  };
}

function appendData(portfolioId: string) {
  return {
    portfolioId,
    sourceKind: "USER_UPLOADED_FILE",
    sourceFileHash: "a".repeat(64),
    parsedRowCount: 1,
    acceptedRowCount: 1,
    nonAppendedRowCount: 0,
    appendedTransactionIds: ["10000000-0000-4000-8000-000000000011"],
    snapshotDatesRebuilt: ["2026-08-08"],
    auditActionCode: "IMPORT_APPEND_BATCH",
    nonSensitiveWarnings: [],
    appendValidationPolicy: "REVIEW_RERUN_AND_ATOMIC_STORE_REVALIDATION",
    rawPayloadRetentionRule: "RAW_CSV_NOT_STORED",
  };
}

function mountPanel() {
  const dom = new JSDOM("<!doctype html><html><body><div id=\"root\"></div></body></html>", { url: "http://localhost" });
  for (const [name, value] of Object.entries({
    window: dom.window,
    document: dom.window.document,
    navigator: dom.window.navigator,
    HTMLElement: dom.window.HTMLElement,
    HTMLInputElement: dom.window.HTMLInputElement,
    Event: dom.window.Event,
    MouseEvent: dom.window.MouseEvent,
  })) {
    Object.defineProperty(globalThis, name, { configurable: true, writable: true, value });
  }
  const container = dom.window.document.getElementById("root")!;
  return { dom, container, root: createRoot(container) };
}

async function selectCSV(container: HTMLElement) {
  const input = container.querySelector("input[type=file]")!;
  Object.defineProperty(input, "files", {
    configurable: true,
    value: [{ name: "broker.csv", size: importCSV.length, text: async () => importCSV }],
  });
  await act(async () => {
    input.dispatchEvent(new Event("change", { bubbles: true }));
    await Promise.resolve();
  });
}

async function submitReview(container: HTMLElement) {
  const form = container.querySelector("form")!;
  await act(async () => {
    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await Promise.resolve();
  });
}

test("mounted panel ignores a stale review response after a portfolio or token rerender", async (t) => {
  const originalFetch = globalThis.fetch;
  const review = deferred<Response>();
  let imported = 0;
  globalThis.fetch = async () => review.promise;
  const { dom, container, root } = mountPanel();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<ImportUploadReviewPanel accessToken="token-a" portfolioId="portfolio-a" onImported={() => { imported += 1; }} />));
  await selectCSV(container);
  await submitReview(container);
  await act(async () => root.render(<ImportUploadReviewPanel accessToken="token-b" portfolioId="portfolio-b" onImported={() => { imported += 1; }} />));
  await act(async () => {
    review.resolve(apiResponse(reviewData("portfolio-a")));
    await review.promise;
    await Promise.resolve();
  });

  assert.equal(imported, 0);
  assert.equal(container.querySelector(".import-review"), null);
  assert.doesNotMatch(container.textContent ?? "", /Review received from the Go API/);
});

test("mounted panel does not invoke onImported for a stale append response", async (t) => {
  const originalFetch = globalThis.fetch;
  const append = deferred<Response>();
  let imported = 0;
  globalThis.fetch = async (input: RequestInfo | URL) => {
    if (String(input).endsWith("/imports/review")) {
      return apiResponse(reviewData("portfolio-a"));
    }
    if (String(input).endsWith("/imports/append")) {
      return append.promise;
    }
    throw new Error(`unexpected request: ${String(input)}`);
  };
  const { dom, container, root } = mountPanel();
  t.after(async () => {
    await act(async () => root.unmount());
    dom.window.close();
    globalThis.fetch = originalFetch;
  });

  await act(async () => root.render(<ImportUploadReviewPanel accessToken="token-a" portfolioId="portfolio-a" onImported={() => { imported += 1; }} />));
  await selectCSV(container);
  await submitReview(container);
  assert.ok(container.querySelector(".import-review"));
  await act(async () => {
    (container.querySelector('input[aria-label="Approve row 2"]') as HTMLInputElement).click();
  });
  await act(async () => {
    [...container.querySelectorAll("button")].find((button) => button.textContent?.startsWith("Append"))!.click();
    await Promise.resolve();
  });
  await act(async () => root.render(<ImportUploadReviewPanel accessToken="token-b" portfolioId="portfolio-b" onImported={() => { imported += 1; }} />));
  await act(async () => {
    append.resolve(apiResponse(appendData("portfolio-a")));
    await append.promise;
    await Promise.resolve();
  });

  assert.equal(imported, 0);
  assert.equal(container.querySelector(".success-panel"), null);
});
