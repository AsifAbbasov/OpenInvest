import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { translations } from "../src/features/locale/LocaleProvider.tsx";

const list = await readFile(
  new URL("../src/features/portfolio/components/PortfolioListSlice.tsx", import.meta.url),
  "utf8",
);
const dashboard = await readFile(
  new URL("../src/features/portfolio/components/DashboardSlice.tsx", import.meta.url),
  "utf8",
);

test("portfolio errors have distinct non-English Russian translations", () => {
  const cases = [
    "Your access has expired. Refresh your session or sign in again.",
    "You do not have permission to view these portfolios.",
    "Connection to OpenInvest failed. Try again.",
  ];
  for (const english of cases) {
    const russian = translations[english];
    assert.equal(typeof russian, "string", `missing Russian message for ${english}`);
    assert.ok(russian.trim().length > 0);
    assert.notEqual(russian, english);
    assert.ok(list.includes(`t("${english}")`), `portfolio UI does not use ${english}`);
  }
});

test("portfolio failure branches select 401, 403 and transport errors separately", () => {
  assert.match(list, /result\.status === 401[\s\S]*?t\("Your access has expired\./);
  assert.match(list, /result\.status === 403[\s\S]*?t\("You do not have permission/);
  assert.match(list, /result\.failureKind === "NETWORK" \|\| result\.failureKind === "TIMEOUT"/);
  assert.doesNotMatch(list, /<p>\{result\.message\}<\/p>/, "raw API error must not be rendered");
});

test("portfolio version is localized without modifying the version value", () => {
  assert.equal(translations["Version"], "Версия");
  assert.match(list, /\{t\("Version"\)\} \{portfolio\.version\}/);
  assert.doesNotMatch(list, /· version \{portfolio\.version\}/);
});

test("Dashboard portfolio links share button styling and keep canonical route", () => {
  const links = [...dashboard.matchAll(/<Link href="\/portfolios"[^>]*>/g)].map(([tag]) => tag);
  assert.equal(links.length, 2);
  for (const link of links) assert.match(link, /className="secondary-button"/);
  assert.doesNotMatch(dashboard, /href="\/#portfolios"/);
});
