import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const file = (name) => readFile(new URL(`../src/${name}`, import.meta.url), "utf8");

test("Dashboard and portfolio list are distinct routes and components", async () => {
  const [home, list, dashboard, portfolios, nav] = await Promise.all([
    file("app/page.tsx"), file("app/portfolios/page.tsx"),
    file("features/portfolio/components/DashboardSlice.tsx"),
    file("features/portfolio/components/PortfolioListSlice.tsx"),
    file("features/navigation/components/ApplicationNavigation.tsx"),
  ]);
  assert.match(home, /DashboardSlice/);
  assert.match(list, /PortfolioListSlice/);
  assert.match(nav, /href: "\/portfolios"/);
  assert.doesNotMatch(nav, /href: "\/#portfolios"/);
  assert.match(dashboard, /href="\/portfolios"/);
  assert.doesNotMatch(dashboard, /listPortfolios\(/);
  assert.match(portfolios, /listPortfolios\(/);
  assert.match(portfolios, /shouldCommitPortfolioLoad/);
  assert.match(portfolios, /CreatePortfolioForm/);
  assert.match(portfolios, /href={`\/portfolios\/\$\{portfolio.id\}`}/);
});
