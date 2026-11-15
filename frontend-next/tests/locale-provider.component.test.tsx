// @ts-nocheck
import assert from "node:assert/strict";
import test from "node:test";
import { JSDOM } from "jsdom";
const dom = new JSDOM("<!doctype html><html lang='ru'><body><div id='root'></div></body></html>", { url: "http://localhost:3000" });
for (const key of ["window", "document", "navigator", "HTMLElement", "HTMLInputElement", "Event", "MouseEvent"]) {
  Object.defineProperty(globalThis, key, { value: dom.window[key], configurable: true, writable: true });
}
globalThis.IS_REACT_ACT_ENVIRONMENT = true;
const { act } = await import("react");
const { createRoot } = await import("react-dom/client");
const { LanguageSwitch, LocaleProvider, useLocale, translations } = await import("../src/features/locale/LocaleProvider.tsx");

function Probe() {
  const { language, t } = useLocale();
  return <>
    <LanguageSwitch />
    <span data-testid="language">{language}</span>
    <strong data-testid="translation">{t("Dashboard")}</strong>
    <label>Portfolio <input aria-label="portfolio-name" /></label>
  </>;
}
function element(selector: string): HTMLElement { const value=dom.window.document.querySelector(selector); assert.ok(value, selector); return value; }
async function click(label: string) {
  await act(async () => { const button = [...dom.window.document.querySelectorAll('button')].find(el => el.textContent === label); assert.ok(button, label); button.dispatchEvent(new dom.window.MouseEvent("click", { bubbles: true })); });
}
async function mount() {
  dom.window.document.body.innerHTML = '<div id="root"></div>';
  dom.window.document.documentElement.lang = "ru";
  const root=createRoot(element("#root"));
  await act(async () => root.render(<LocaleProvider><Probe /></LocaleProvider>));
  return root;
}
test("RU default; RU/EN/RU toggle keeps form input intact and maintains html lang", { concurrency: false }, async () => {
  dom.window.localStorage.clear();
  const root=await mount();
  try {
    assert.equal(element('[data-testid="language"]').textContent, "ru");
    assert.equal(element('[data-testid="translation"]').textContent, "Обзор");
    const input=element('input[aria-label="portfolio-name"]') as HTMLInputElement;
    input.value="Long-term portfolio";
    await click("EN");
    assert.equal(element('[data-testid="language"]').textContent,"en");
    assert.equal(element('[data-testid="translation"]').textContent,"Dashboard");
    assert.equal(dom.window.document.documentElement.lang,"en");
    assert.equal(input.value,"Long-term portfolio");
    assert.equal(dom.window.localStorage.getItem("openinvest.ui.language"),"en");
    await click("RU");
    assert.equal(element('[data-testid="language"]').textContent,"ru");
    assert.equal(dom.window.document.documentElement.lang,"ru");
    assert.equal(input.value,"Long-term portfolio");
  } finally { await act(async () => root.unmount()); }
});
test("saved EN survives remount; no RU default overwrites storage", { concurrency: false }, async () => {
  dom.window.localStorage.setItem("openinvest.ui.language","en");
  const root=await mount();
  try {
    assert.equal(element('[data-testid="translation"]').textContent,"Dashboard");
    assert.equal(dom.window.document.documentElement.lang,"en");
    assert.equal(dom.window.localStorage.getItem("openinvest.ui.language"),"en");
  } finally { await act(async () => root.unmount()); }
});
test("translation dictionary is nonempty and contains finance/navigation keys", () => {
  for (const key of ["Dashboard", "BUY", "SELL", "As of", "positions valued", "Gross dividend"]) {
    assert.ok(Object.hasOwn(translations,key),key);
    assert.ok(translations[key].trim().length > 0,key);
  }
});


test("real DividendCalculator keeps entered values across RU/EN switching and starts empty", { concurrency: false }, async () => {
  const { DividendCalculator } = await import("../src/features/dividends/components/DividendCalculator.tsx");
  dom.window.localStorage.clear();
  dom.window.document.body.innerHTML = '<div id="root"></div>';
  const root = createRoot(element("#root"));
  await act(async () => root.render(<LocaleProvider><LanguageSwitch /><DividendCalculator /></LocaleProvider>));
  try {
    const inputs = Array.from(dom.window.document.querySelectorAll("input")) as HTMLInputElement[];
    assert.ok(inputs.length >= 4);
    assert.deepEqual(inputs.slice(0, 4).map(input => input.value), ["", "", "", ""]);
    const setter = Object.getOwnPropertyDescriptor(dom.window.HTMLInputElement.prototype, "value")!.set!;
    await act(async () => {
      setter.call(inputs[0], "GAZP");
      inputs[0].dispatchEvent(new dom.window.Event("input", { bubbles: true }));
      setter.call(inputs[1], "12.5");
      inputs[1].dispatchEvent(new dom.window.Event("input", { bubbles: true }));
    });
    await click("EN");
    assert.equal(inputs[0].value, "GAZP");
    assert.equal(inputs[1].value, "12.5");
    await click("RU");
    assert.equal(inputs[0].value, "GAZP");
    assert.equal(inputs[1].value, "12.5");
  } finally { await act(async () => root.unmount()); }
});

test("server markup hydrates without mismatch and restores stored English", { concurrency: false }, async () => {
  const { renderToString } = await import("react-dom/server");
  const { hydrateRoot } = await import("react-dom/client");
  dom.window.localStorage.setItem("openinvest.ui.language", "en");
  const markup = renderToString(<LocaleProvider><Probe /></LocaleProvider>);
  assert.match(markup, /Обзор/);
  dom.window.document.body.innerHTML = `<div id="hydrate-root">${markup}</div>`;
  dom.window.document.documentElement.lang = "ru";
  const errors: string[] = [];
  const root = await act(async () => hydrateRoot(element("#hydrate-root"), <LocaleProvider><Probe /></LocaleProvider>, {
    onRecoverableError(error) { errors.push(String(error)); },
  }));
  try {
    assert.deepEqual(errors, [], "Hydration recovery must not occur");
    assert.equal(element('[data-testid="translation"]').textContent, "Dashboard");
    assert.equal(dom.window.document.documentElement.lang, "en");
    assert.equal(dom.window.localStorage.getItem("openinvest.ui.language"), "en");
  } finally { await act(async () => root.unmount()); }
});

test("typed dictionary has nonempty Russian text for every key", () => {
  for (const [key, value] of Object.entries(translations)) {
    assert.ok(key.trim(), "Empty translation key");
    assert.equal(typeof value, "string", key);
    assert.ok(value.trim().length > 0, key);
  }
});


test("portfolio detail summary and history have typed RU and EN strings", async () => {
  const { readFileSync } = await import("node:fs");
  const source = readFileSync(new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url), "utf8");
  const required = [
    "Review portfolio value, income, positions and money-weighted return. Time-weighted, nominal, real and inflation-adjusted return views are not available yet.",
    "Historical positions scope",
    "Time Machine reconstructs positions for the selected date. Historical market value, cash and performance are not shown when the required data is unavailable.",
  ];
  for (const english of required) {
    assert.ok(Object.hasOwn(translations, english), `Missing translation: ${english}`);
    assert.ok(translations[english].trim().length > 0, `Empty Russian translation: ${english}`);
    assert.notEqual(translations[english], english, `RU copy not translated: ${english}`);
    assert.ok(source.includes(`t(${JSON.stringify(english)})`), `Detail screen does not use translated key: ${english}`);
  }
  assert.ok(source.includes('aria-label={t("Historical positions scope")}'));
  assert.ok(source.includes('positionViewMode === "historical"'), "History scope must remain conditional");
});
