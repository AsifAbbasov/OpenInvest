// @ts-nocheck
// Behavioral test mounts the real PortfolioListSlice. Only API/auth/dependencies are test doubles.
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";

const dom = new JSDOM('<!doctype html><html><body><div id="root"></div></body></html>', {
  url: "https://localhost/#portfolios",
});
for (const [key, value] of Object.entries({
  window: dom.window, document: dom.window.document, navigator: dom.window.navigator,
  HTMLElement: dom.window.HTMLElement, Event: dom.window.Event,
})) Object.defineProperty(globalThis, key, { configurable: true, writable: true, value });
Object.defineProperty(globalThis, "IS_REACT_ACT_ENVIRONMENT", { configurable: true, writable: true, value: true });

const React = await import("react");
const { createRoot } = await import("react-dom/client");
const { PortfolioListSlice } = await import("../src/features/portfolio/components/PortfolioListSlice.tsx");

test("empty portfolio response retains a focusable #portfolios target", async () => {
  const container = dom.window.document.getElementById("root");
  const root = createRoot(container);
  try {
    await React.act(async () => {
      root.render(React.createElement(PortfolioListSlice));
      await Promise.resolve();
    });
    await React.act(async () => { await Promise.resolve(); });
    assert.ok(container.querySelector('[data-testid="empty-portfolio-create"]'),
      "fixture must demonstrate the empty portfolio path");
    assert.match(container.textContent ?? "", /Asset discovery/);
    const anchor = container.querySelector('#portfolios');
    assert.ok(anchor, "fragment target must exist with zero portfolios");
    assert.equal(container.querySelectorAll('#portfolios').length, 1);
    anchor.focus();
    assert.equal(dom.window.document.activeElement, anchor);
  } finally {
    await React.act(async () => root.unmount());
  }
});

test.after(() => dom.window.close());
