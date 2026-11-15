// @ts-nocheck
// Mount the production navigation component in a real React DOM tree.
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";
import { ApplicationNavigation, selectedSection } from "../src/features/navigation/components/ApplicationNavigation";

const dom = new JSDOM('<!doctype html><html><body><div id="app"></div></body></html>', {
  url: "https://localhost/",
});
for (const [key, value] of Object.entries({
  window: dom.window, document: dom.window.document, navigator: dom.window.navigator,
  HTMLElement: dom.window.HTMLElement, Event: dom.window.Event,
  MouseEvent: dom.window.MouseEvent, KeyboardEvent: dom.window.KeyboardEvent,
})) Object.defineProperty(globalThis, key, { configurable: true, writable: true, value });
Object.defineProperty(globalThis, "IS_REACT_ACT_ENVIRONMENT", { configurable: true, writable: true, value: true });

function TestLink({ href, className, children, "aria-current": current }) {
  return <a href={href} className={className} aria-current={current}>{children}</a>;
}

const React = await import("react");
const { createRoot } = await import("react-dom/client");

const rootNode = dom.window.document.getElementById("app");
const root = createRoot(rootNode);
async function render(pathname, hash, label = "Primary navigation") {
  await React.act(async () => {
    root.render(<ApplicationNavigation
      pathname={pathname} hash={hash} label={label}
      className="navigation" linkClassName="link" activeClassName="active"
      LinkComponent={TestLink}
    />);
  });
  return rootNode.querySelector("nav");
}
function selected(nav) { return [...nav.querySelectorAll('a[aria-current]')]; }

test("real React navigation updates current route and removes stale aria-current", async () => {
  for (const [path, hash, expected] of [
    ["/", "", "Dashboard"],
    ["/assets", "", "Assets"],
    ["/dividends/calculator", "", "Dividend calculator"],
    ["/corporate-actions", "", "Corporate actions"],
    ["/", "#portfolios", "Portfolios"],
    ["/portfolios/11111111", "", "Portfolios"],
    ["/", "", "Dashboard"],
  ]) {
    const nav = await render(path, hash);
    const current = selected(nav);
    assert.equal(current.length, 1, `exactly one current item for ${path}${hash}`);
    assert.equal(current[0].textContent, expected);
    assert.ok(current[0].classList.contains("active"));
    assert.equal(current[0].getAttribute("aria-current"), expected === "Portfolios" ? "location" : "page");
  }
  assert.equal(selectedSection("/unknown", ""), null);
  assert.equal(selected(await render("/unknown", "")).length, 0);
});

test("portfolio fragment is keyboard reachable in both navigation variants", async () => {
  for (const label of ["Primary navigation", "Mobile navigation"]) {
    const nav = await render("/", "#portfolios", label);
    const anchor = [...nav.querySelectorAll("a")].find(a => a.textContent === "Portfolios");
    assert.equal(anchor?.getAttribute("href"), "/#portfolios");
    assert.equal(nav.getAttribute("aria-label"), label);
    anchor.focus();
    assert.equal(dom.window.document.activeElement, anchor);
    assert.equal(anchor.getAttribute("tabindex"), null, "anchor retains native keyboard activation");
  }
});

test("portfolio anchor is in unconditional Dashboard markup", async () => {
  const fs = await import("node:fs/promises");
  const url = new URL("../src/features/portfolio/components/DashboardSlice.tsx", import.meta.url);
  const source = await fs.readFile(url, "utf8");
  assert.equal((source.match(/id="portfolios"/g) ?? []).length, 1);
  assert.match(source, /<div id="portfolios" tabIndex=\{-1\}/);
  assert.ok(source.indexOf('id="portfolios"') < source.indexOf('result === null ?'),
    "anchor must exist in loading, empty and populated states");
  assert.doesNotMatch(source, /<section className="panel" id="portfolios">/);
});

test.after(async () => { await React.act(async () => root.unmount()); dom.window.close(); });
