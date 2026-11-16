import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const shell = await readFile(new URL("../src/features/navigation/components/ApplicationShell.tsx", import.meta.url), "utf8");
const auth = await readFile(new URL("../src/features/auth/components/AuthShell.tsx", import.meta.url), "utf8");
const dashboard = await readFile(new URL("../src/features/portfolio/components/DashboardSlice.tsx", import.meta.url), "utf8");
const nav = await readFile(new URL("../src/features/navigation/components/ApplicationNavigation.tsx", import.meta.url), "utf8");
const css = await readFile(new URL("../src/features/navigation/components/ApplicationShell.module.css", import.meta.url), "utf8");

test("shell contains independent Dashboard and Portfolios navigation targets", () => {
  for (const target of ['"/"', '"/portfolios"', '"/assets"', '"/dividends/calculator"', '"/corporate-actions"']) {
    assert.ok(nav.includes(target), `missing ${target}`);
  }
  assert.match(dashboard, /id="portfolios"/);
  assert.match(nav, /aria-current=/);
});

test("shell preserves AuthShell session ownership and excludes anonymous login", () => {
  assert.match(auth, /if \(state\.status !== "authenticated"\)/);
  assert.match(auth, /return <AuthForm/);
  assert.match(auth, /<ApplicationShell/);
  assert.match(shell, /Skip to main content/);
  assert.match(shell, /aria-label="Mobile navigation"|"Mobile navigation"/);
});

test("responsive shell uses keyboard focus outlines and scoped CSS", () => {
  assert.match(css, /@media \(max-width: 1023px\)/);
  assert.match(css, /@media \(max-width: 600px\)/);
  assert.match(css, /:focus-visible/);
  assert.doesNotMatch(shell, /fetch\(|localStorage|sessionStorage|calculateDividend/);
});
