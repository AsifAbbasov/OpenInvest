"use strict";

// External REAL-browser harness: install Playwright outside this repository,
// provision two verified accounts in a disposable local stack, then run it.
const assert = require("node:assert/strict");
const { randomUUID } = require("node:crypto");
const FRONTEND = process.env.M12_FRONTEND_URL || "http://localhost:3000";
const ALT = process.env.M12_FRONTEND_ALT_URL || "http://127.0.0.1:3000";
const BACKEND = process.env.M12_BACKEND_URL || "http://localhost:18080";
const accounts = [
  { email: process.env.M12_VERIFIED_EMAIL_A, password: process.env.M12_VERIFIED_PASSWORD_A },
  { email: process.env.M12_VERIFIED_EMAIL_B, password: process.env.M12_VERIFIED_PASSWORD_B },
];
const results = [];

function requireFixture() {
  if (process.env.M12_DISPOSABLE_FIXTURE !== "1" || accounts.some((a) => !a.email || !a.password) || accounts[0].email === accounts[1].email) {
    throw Error("EXTERNAL_EVIDENCE_REQUIRED: two distinct pre-verified disposable accounts required");
  }
  for (const endpoint of [FRONTEND, ALT, BACKEND]) {
    const url = new URL(endpoint);
    if (url.protocol !== "http:" || !["localhost", "127.0.0.1", "[::1]"].includes(url.hostname)) {
      throw Error("EXTERNAL_EVIDENCE_REQUIRED: only loopback disposable hosts are permitted");
    }
  }
  if (FRONTEND === ALT) throw Error("EXTERNAL_EVIDENCE_REQUIRED: separate disallowed origin required");
}

function record(name, ok) {
  results.push({ name, ok });
  // Never log fixture credentials, cookies, browser storage or response bodies.
  process.stdout.write(JSON.stringify({ name, ok }) + "\n");
}
async function check(name, fn) {
  try { await fn(); record(name, true); } catch { record(name, false); }
}
async function request(page, path, init = {}) {
  return page.evaluate(async ({ url, init }) => {
    try {
      const response = await fetch(url, { credentials: "include", ...init });
      const bodyText = await response.text();
      let body = null;
      try { body = JSON.parse(bodyText); } catch { /* body is not JSON */ }
      return { transport: true, status: response.status, body };
    } catch { return { transport: false, status: 0, body: null }; }
  }, { url: BACKEND + path, init });
}
const post = (page, path, body, headers = {}) => request(page, path, {
  method: "POST", headers: { "Content-Type": "application/json", ...headers },
  body: body === undefined ? undefined : JSON.stringify(body),
});
const refresh = (page, csrf) => post(page, "/api/v1/auth/refresh", undefined, csrf ? { "X-CSRF-Token": csrf } : {});
const logout = (page, csrf) => post(page, "/api/v1/auth/logout", { allSessions: false }, { "X-CSRF-Token": csrf });
const list = (page, token) => request(page, "/api/v1/portfolios", { headers: { Authorization: `Bearer ${token}` } });
const create = (page, token, name) => post(page, "/api/v1/portfolios", { name, baseCurrency: "RUB" }, {
  Authorization: `Bearer ${token}`, "Idempotency-Key": randomUUID(),
});
const state = (result) => result.body?.data?.session;
const cookie = async (context) => (await context.cookies(BACKEND + "/api/v1/auth/refresh")).find((c) => c.name === "oi_refresh");

async function probeCopiedCookie(browser, stale, csrf) {
  const context = await browser.newContext();
  try {
    await context.addCookies([stale]);
    const page = await context.newPage();
    await page.goto(FRONTEND);
    return refresh(page, csrf);
  } finally { await context.close(); }
}

async function main() {
  requireFixture();
  let chromium;
  try { ({ chromium } = require("playwright")); }
  catch { throw Error("EXTERNAL_EVIDENCE_REQUIRED: Playwright must be provisioned outside the repository"); }
  const browser = await chromium.launch({ headless: true });
  try {
    const ctx = await Promise.all([browser.newContext(), browser.newContext()]);
    const pages = await Promise.all(ctx.map((c) => c.newPage()));
    await Promise.all(pages.map((p) => p.goto(FRONTEND)));
    const marker = randomUUID().slice(0, 12);
    const anon = await browser.newContext();
    const anonPage = await anon.newPage();
    await anonPage.goto(FRONTEND);
    const email = `m12-p11-${marker}@example.test`;
    await check("register_202_without_auto_login", async () => {
      const r = await post(anonPage, "/api/v1/auth/register", { email });
      assert.equal(r.status, 202);
      assert.equal(r.body?.data?.status, "REQUEST_ACCEPTED");
      assert.equal(r.body?.data?.session, undefined);
      assert.equal(r.body?.data?.user, undefined);
      assert.equal(await cookie(anon), undefined);
    });
    await check("parallel_registration_202_without_cookies", async () => {
      const pair = await Promise.all([post(anonPage, "/api/v1/auth/register", { email }), post(anonPage, "/api/v1/auth/register", { email })]);
      assert.ok(pair.every((r) => r.status === 202 && !r.body?.data?.session));
      assert.equal(await cookie(anon), undefined);
    });
    await anon.close();

    const logins = await Promise.all(pages.map((p, i) => post(p, "/api/v1/auth/login", accounts[i])));
    await check("verified_logins_have_distinct_principals", async () => {
      assert.ok(logins.every((r) => r.status === 200 && state(r)?.accessToken && state(r)?.csrfToken));
      assert.notEqual(logins[0].body.data.user.id, logins[1].body.data.user.id);
    });
    if (!results.at(-1).ok) throw Error("EXTERNAL_EVIDENCE_REQUIRED: verified fixture logins failed");
    const cookies = await Promise.all(ctx.map(cookie));
    await check("secure_httponly_strict_isolated_refresh_cookies", async () => {
      for (const c of cookies) {
        assert.ok(c?.value); assert.equal(c.httpOnly, true); assert.equal(c.secure, true);
        assert.equal(c.sameSite, "Strict"); assert.equal(c.path, "/api/v1/auth");
      }
      assert.notEqual(cookies[0].value, cookies[1].value);
      assert.equal(await pages[0].evaluate(() => document.cookie.includes("oi_refresh=")), false);
    });
    const names = [`P11-A-${marker}`, `P11-B-${marker}`];
    await check("two_principals_mutate_only_their_portfolios", async () => {
      const r = await Promise.all(pages.map((p, i) => create(p, state(logins[i]).accessToken, names[i])));
      assert.ok(r.every((x) => x.status === 201));
    });
    await check("bearer_across_other_browser_cookie_keeps_principal", async () => {
      const a = await list(pages[1], state(logins[0]).accessToken);
      const b = await list(pages[0], state(logins[1]).accessToken);
      assert.equal(a.status, 200); assert.equal(b.status, 200);
      assert.ok(JSON.stringify(a.body).includes(names[0]) && !JSON.stringify(a.body).includes(names[1]));
      assert.ok(JSON.stringify(b.body).includes(names[1]) && !JSON.stringify(b.body).includes(names[0]));
    });
    await check("missing_and_wrong_csrf_rejected", async () => {
      assert.equal((await refresh(pages[0])).status, 401);
      assert.equal((await refresh(pages[0], state(logins[1]).csrfToken)).status, 401);
    });
    await check("mixed_cookie_and_other_principal_csrf_rejected", async () => {
      assert.equal((await probeCopiedCookie(browser, cookies[0], state(logins[1]).csrfToken)).status, 401);
      assert.equal((await probeCopiedCookie(browser, cookies[1], state(logins[0]).csrfToken)).status, 401);
    });
    await check("parallel_refresh_preserves_isolated_principals", async () => {
      const rotated = await Promise.all(pages.map((p, i) => refresh(p, state(logins[i]).csrfToken)));
      assert.ok(rotated.every((r) => r.status === 200 && r.body?.data?.csrfToken));
      const own = await Promise.all(pages.map((p, i) => list(p, rotated[i].body.data.accessToken)));
      for (let i = 0; i < 2; i++) {
        assert.equal(own[i].status, 200);
        assert.ok(JSON.stringify(own[i].body).includes(names[i]));
        assert.ok(!JSON.stringify(own[i].body).includes(names[1 - i]));
        state(logins[i]).csrfToken = rotated[i].body.data.csrfToken;
      }
    });
    await check("logout_invalidates_stale_browser_cookie", async () => {
      const stale = await cookie(ctx[1]); const csrf = state(logins[1]).csrfToken;
      assert.equal((await logout(pages[1], csrf)).status, 200);
      assert.equal((await probeCopiedCookie(browser, stale, csrf)).status, 401);
      assert.equal(await cookie(ctx[1]), undefined);
    });
    await check("rotated_cookie_replay_rejected", async () => {
      const stale = await cookie(ctx[0]); const csrf = state(logins[0]).csrfToken;
      assert.equal((await refresh(pages[0], csrf)).status, 200);
      assert.equal((await probeCopiedCookie(browser, stale, csrf)).status, 401);
    });
    await check("concurrent_refresh_single_session_at_most_one_success", async () => {
      const again = await post(pages[0], "/api/v1/auth/login", accounts[0]);
      assert.equal(again.status, 200);
      const pair = await Promise.all([refresh(pages[0], state(again).csrfToken), refresh(pages[0], state(again).csrfToken)]);
      assert.ok(pair.filter((r) => r.status === 200).length <= 1);
    });
    await check("logout_racing_refresh_cannot_restore_stale_session", async () => {
      const again = await post(pages[1], "/api/v1/auth/login", accounts[1]);
      assert.equal(again.status, 200);
      const stale = await cookie(ctx[1]); const csrf = state(again).csrfToken;
      const pair = await Promise.all([logout(pages[1], csrf), refresh(pages[1], csrf)]);
      assert.ok(pair.every((r) => r.transport && r.status >= 200 && r.status < 500));
      assert.equal((await probeCopiedCookie(browser, stale, csrf)).status, 401);
    });
    await check("mutation_without_bearer_denied_for_four_content_types", async () => {
      for (const [type, body] of [
        ["application/json", "{}"], ["text/plain", "x"],
        ["application/x-www-form-urlencoded", "name=x"],
        ["multipart/form-data; boundary=m12", "--m12--\r\n"],
      ]) {
        assert.equal((await request(pages[0], "/api/v1/portfolios", { method: "POST", headers: { "Content-Type": type }, body })).status, 401);
      }
    });
    const denied = await browser.newContext();
    try {
      const page = await denied.newPage(); await page.goto(ALT);
      await check("disallowed_origin_cannot_read_registration", async () => {
        const r = await post(page, "/api/v1/auth/register", { email: `p11-alt-${marker}@example.test` });
        assert.equal(r.transport, false);
      });
    } finally { await denied.close(); }
    const nullPage = await ctx[0].newPage();
    await nullPage.goto("data:text/html,<title>P11</title>");
    await check("null_origin_cannot_read_registration", async () => {
      const r = await post(nullPage, "/api/v1/auth/register", { email: `p11-null-${marker}@example.test` });
      assert.equal(r.transport, false);
    });
    await Promise.all(ctx.map((c) => c.close()));
  } finally { await browser.close(); }
  const failed = results.filter((r) => !r.ok).length;
  process.stdout.write(JSON.stringify({ summary: { checked: results.length, failed, browser: "Chromium", fixture: "DISPOSABLE" } }) + "\n");
  if (failed) process.exitCode = 1;
}

main().catch((error) => {
  console.error(error.message.startsWith("EXTERNAL_EVIDENCE_REQUIRED:") ? error.message : "P11_BROWSER_HARNESS_FAILED");
  process.exitCode = 2;
});
