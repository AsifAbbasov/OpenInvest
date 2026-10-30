"use strict";

// External REAL-browser harness: install Playwright outside this repository,
// provision two verified accounts in a disposable local stack, then run it.
const assert = require("node:assert/strict");
const { randomUUID } = require("node:crypto");
const { readFileSync } = require("node:fs");
const http = require("node:http");
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
  if (new URL(FRONTEND).origin === new URL(BACKEND).origin ||
      new URL(ALT).origin === new URL(BACKEND).origin) {
    throw Error("EXTERNAL_EVIDENCE_REQUIRED: API and browser test origins must differ");
  }
  // Operator-supplied evidence is a prerequisite, not proof created by this harness.
  // The operator must independently verify the database identities before execution.
  let proof;
  try { proof = JSON.parse(readFileSync(process.env.M12_DISPOSABLE_INFRA_PROOF_PATH, "utf8")); }
  catch { throw Error("EXTERNAL_EVIDENCE_REQUIRED: verified disposable infrastructure attestation missing"); }
  const db = proof?.postgres;
  if (proof?.purpose !== "OPENINVEST_P11_DISPOSABLE" ||
      proof.backendUrl !== BACKEND || proof.frontendUrl !== FRONTEND ||
      proof.approvedAlternateOrigin !== ALT || proof.playwrightVerified !== true ||
      !proof.operatorVerifiedAt || !proof.operatorIdentity ||
      db?.isolated !== true || db?.sourceClusterExcluded !== true ||
      !db?.clusterId || !db?.database || !db?.host || !db?.role ||
      proof.accounts?.a?.email !== accounts[0].email ||
      proof.accounts?.b?.email !== accounts[1].email ||
      !proof.accounts.a.id || !proof.accounts.b.id ||
      proof.accounts.a.id === proof.accounts.b.id) {
    throw Error("EXTERNAL_EVIDENCE_REQUIRED: incomplete independently verified fixture identities");
  }
  return proof;
}

function requireDelivery(probe) {
  assert.equal(probe?.delivery?.inContext, true, "copied cookie absent from destination context");
  assert.equal(probe?.delivery?.requestObserved, true, "no outgoing auth request witnessed");
  assert.equal(probe?.delivery?.cookieMatched, true, "original cookie not attached to outgoing request");
  assert.equal(probe?.response?.transport, true, "browser request failed before HTTP response");
}

function requireRefreshRace(pair) {
  assert.equal(pair.length, 2);
  pair.forEach(requireDelivery);
  assert.deepEqual(pair.map((p) => p.response.status).sort(), [200, 401],
    "exactly one rotation must commit and the replay must be rejected");
  const winner = pair.find((p) => p.response.status === 200);
  assert.ok(winner.response.body?.data?.csrfToken && winner.replacementCookie?.value,
    "successful rotation must return replacement session credentials");
  return winner;
}

function requireLogoutRace(logoutProbe, refreshProbe) {
  requireDelivery(logoutProbe); requireDelivery(refreshProbe);
  const logoutStatus = logoutProbe.response.status;
  const refreshStatus = refreshProbe.response.status;
  assert.ok([200, 401].includes(logoutStatus) && [200, 401].includes(refreshStatus));
  assert.ok(logoutStatus === 200 || refreshStatus === 200,
    "401 + 401 cannot establish a successful logout");
  if (logoutStatus === 200) assert.equal(logoutProbe.response.body?.data?.revoked, true);
  if (refreshStatus === 200) {
    assert.ok(refreshProbe.response.body?.data?.csrfToken && refreshProbe.replacementCookie?.value);
  }
  return logoutStatus === 200 ? "logout_committed" : "requires_replacement_logout";
}

function requireCommittedLogout(probe) {
  assert.equal(probe.response.status, 200, "original logout failed its committed-revocation contract");
  assert.equal(probe.response.body?.data?.revoked, true);
}

function requireCorsPolicyProof({ control, result, origin, hits, diagnostic, cspViolation }) {
  assert.equal(control?.reachable, true, "allowed-Origin browser/backend control failed");
  assert.equal(result?.transport, false, "disallowed Origin could read an API response");
  assert.equal(cspViolation, false, "browser CSP, not CORS, blocked the request");
  assert.equal(diagnostic, true, "Chromium did not diagnose an Origin/CORS policy rejection");
  assert.ok(hits.some((h) => h.origin === origin && h.method === "OPTIONS" &&
    h.upstreamReached && h.status >= 100 && h.status < 600 &&
    h.acao !== origin), "no backend preflight witness with denied Origin");
  assert.ok(hits.every((h) => !h.networkError && h.acao !== origin),
    "upstream network error or permissive CORS response");
}

function mustFailControl(name, fn) {
  assert.throws(fn, assert.AssertionError);
  process.stdout.write(JSON.stringify({ negativeControl: name, result: "REJECTED_AS_REQUIRED" }) + "\n");
}

function runNegativeControls() {
  const delivered = { delivery: { inContext: true, requestObserved: true, cookieMatched: true }, response: { transport: true, status: 401 } };
  mustFailControl("B01_refresh_401_401", () => requireRefreshRace([delivered, delivered]));
  mustFailControl("B02_logout_refresh_401_401", () => requireLogoutRace(delivered, delivered));
  mustFailControl("B02_failed_original_logout_not_success", () => requireCommittedLogout({
    response: { status: 401, body: { data: { revoked: false } } },
  }));
  const goodControl = { reachable: true };
  const missing = { transport: false };
  mustFailControl("B03_unreachable_backend", () => requireCorsPolicyProof({
    control: { reachable: false }, result: missing, origin: "null", hits: [], diagnostic: false, cspViolation: false,
  }));
  mustFailControl("B03_generic_fetch_exception", () => requireCorsPolicyProof({
    control: goodControl, result: missing, origin: "null", hits: [], diagnostic: false, cspViolation: false,
  }));
  mustFailControl("B04_cookie_not_sent", () => requireDelivery({
    delivery: { inContext: true, requestObserved: true, cookieMatched: false },
    response: { transport: true, status: 401 },
  }));
  process.stdout.write("NEGATIVE_CONTROLS=6/6_PASS\n");
}

function record(name, ok) {
  results.push({ name, ok });
  // Never log fixture credentials, cookies, browser storage or response bodies.
  process.stdout.write(JSON.stringify({ name, ok }) + "\n");
}
async function check(name, fn) {
  try { await fn(); record(name, true); } catch { record(name, false); }
}
async function request(page, path, init = {}, base = BACKEND) {
  return page.evaluate(async ({ url, init }) => {
    try {
      const response = await fetch(url, { credentials: "include", ...init });
      const bodyText = await response.text();
      let body = null;
      try { body = JSON.parse(bodyText); } catch { /* body is not JSON */ }
      return { transport: true, status: response.status, body };
    } catch { return { transport: false, status: 0, body: null }; }
  }, { url: base + path, init });
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

async function openNeutralOriginPage(context, origin) {
  const page = await context.newPage();
  // Suppress application boot: navigation owns the frontend Origin but does not
  // run its auth-refresh effect or emit unrelated CORS/CSP console errors.
  await page.route(new URL(origin).origin + "/**", async (route) => {
    if (route.request().isNavigationRequest()) {
      await route.fulfill({ status: 200, contentType: "text/html",
        body: "<!doctype html><title>P11 isolated browser probe</title>" });
    } else {
      await route.continue();
    }
  });
  const navigation = await page.goto(new URL("/", origin).href);
  assert.equal(navigation?.ok(), true, "isolated Origin test navigation failed");
  return page;
}

async function probeCopiedCookie(browser, stale, csrf, action = "refresh") {
  assert.ok(stale?.value, "source cookie is missing");
  const context = await browser.newContext();
  try {
    await context.addCookies([stale]);
    const inContext = (await cookie(context))?.value === stale.value;
    let requestObserved = false;
    let cookieMatched = false;
    await context.route("**/api/v1/auth/*", async (route) => {
      const outgoing = route.request();
      const expectedPath = action === "logout" ? "/api/v1/auth/logout" : "/api/v1/auth/refresh";
      if (outgoing.url() === BACKEND + expectedPath && outgoing.method() === "POST") {
        requestObserved = true;
        const headers = await outgoing.allHeaders();
        // Secret comparison stays in memory; never print Cookie or credentials.
        cookieMatched = String(headers.cookie || "").split(";")
          .some((part) => part.trim() === "oi_refresh=" + stale.value);
      }
      await route.continue();
    });
    const page = await openNeutralOriginPage(context, FRONTEND);
    const response = action === "logout" ? await logout(page, csrf) : await refresh(page, csrf);
    return { response, replacementCookie: await cookie(context),
      delivery: { inContext, requestObserved, cookieMatched } };
  } finally { await context.close(); }
}

async function startWitnessProxy() {
  const hits = [];
  const server = http.createServer((incoming, outgoing) => {
    const destination = new URL(incoming.url, BACKEND);
    const upstream = http.request(destination, {
      method: incoming.method, headers: { ...incoming.headers, host: destination.host },
    }, (response) => {
      hits.push({
        method: incoming.method, path: destination.pathname,
        origin: incoming.headers.origin || "", status: response.statusCode,
        acao: response.headers["access-control-allow-origin"] || "",
        upstreamReached: true, networkError: false,
      });
      outgoing.writeHead(response.statusCode, response.headers);
      response.pipe(outgoing);
    });
    upstream.on("error", () => {
      hits.push({ origin: incoming.headers.origin || "", method: incoming.method,
        path: destination.pathname, upstreamReached: false, networkError: true });
      if (!outgoing.headersSent) outgoing.writeHead(502);
      outgoing.end();
    });
    incoming.pipe(upstream);
  });
  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });
  return {
    url: "http://127.0.0.1:" + server.address().port, hits,
    close: () => new Promise((resolve, reject) => server.close((err) => err ? reject(err) : resolve())),
  };
}

async function observeCorsRejection(page, probe) {
  let corsDiagnostic = false;
  await page.evaluate(() => {
    window.__m12CspViolation = false;
    document.addEventListener("securitypolicyviolation", () => { window.__m12CspViolation = true; }, { once: true });
  });
  const listener = (message) => {
    // Only a boolean reaches the evidence; console text may contain URLs.
    if (message.type() === "error" && /CORS policy|CORS header|cross-origin resource sharing/i.test(message.text())) {
      corsDiagnostic = true;
    }
  };
  page.on("console", listener);
  try {
    const result = await probe();
    return { result, diagnostic: corsDiagnostic,
      cspViolation: await page.evaluate(() => window.__m12CspViolation === true) };
  } finally { page.off("console", listener); }
}

async function main() {
  const fixtureProof = requireFixture();
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
      assert.equal(logins[0].body.data.user.id, fixtureProof.accounts.a.id,
        "verified A principal differs from isolated database attestation");
      assert.equal(logins[1].body.data.user.id, fixtureProof.accounts.b.id,
        "verified B principal differs from isolated database attestation");
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
    await check("bearer_cross_browser_context_keeps_principal", async () => {
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
      const a = await probeCopiedCookie(browser, cookies[0], state(logins[1]).csrfToken);
      const b = await probeCopiedCookie(browser, cookies[1], state(logins[0]).csrfToken);
      requireDelivery(a); requireDelivery(b);
      assert.equal(a.response.status, 401); assert.equal(b.response.status, 401);
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
      const replay = await probeCopiedCookie(browser, stale, csrf);
      requireDelivery(replay); assert.equal(replay.response.status, 401);
      assert.equal(await cookie(ctx[1]), undefined);
    });
    await check("rotated_cookie_replay_rejected", async () => {
      const stale = await cookie(ctx[0]); const csrf = state(logins[0]).csrfToken;
      assert.equal((await refresh(pages[0], csrf)).status, 200);
      const replay = await probeCopiedCookie(browser, stale, csrf);
      requireDelivery(replay); assert.equal(replay.response.status, 401);
    });
    await check("concurrent_refresh_same_original_cookie_one_success_family_revoked", async () => {
      const again = await post(pages[0], "/api/v1/auth/login", accounts[0]);
      assert.equal(again.status, 200);
      const original = await cookie(ctx[0]);
      assert.ok(original?.value);
      // Separate cookie jars freeze the same original token for BOTH requests.
      const pair = await Promise.all([
        probeCopiedCookie(browser, original, state(again).csrfToken),
        probeCopiedCookie(browser, original, state(again).csrfToken),
      ]);
      const winner = requireRefreshRace(pair);
      assert.notEqual(winner.replacementCookie.value, original.value);
      // A replay of the revoked original revokes the entire active family.
      const replacement = await probeCopiedCookie(browser, winner.replacementCookie, winner.response.body.data.csrfToken);
      requireDelivery(replacement);
      assert.equal(replacement.response.status, 401, "replayed family descendant remained usable");
    });
    await check("logout_refresh_race_explicitly_terminates_entire_session_family", async () => {
      const again = await post(pages[1], "/api/v1/auth/login", accounts[1]);
      assert.equal(again.status, 200);
      const original = await cookie(ctx[1]); const csrf = state(again).csrfToken;
      assert.ok(original?.value);
      const [logoutProbe, refreshProbe] = await Promise.all([
        probeCopiedCookie(browser, original, csrf, "logout"),
        probeCopiedCookie(browser, original, csrf),
      ]);
      const outcome = requireLogoutRace(logoutProbe, refreshProbe);
      if (outcome === "requires_replacement_logout") {
        const termination = await probeCopiedCookie(browser, refreshProbe.replacementCookie,
          refreshProbe.response.body.data.csrfToken, "logout");
        requireDelivery(termination);
        assert.equal(termination.response.status, 200, "fallback logout did not commit");
        assert.equal(termination.response.body?.data?.revoked, true);
      }
      const originalReplay = await probeCopiedCookie(browser, original, csrf);
      requireDelivery(originalReplay);
      assert.equal(originalReplay.response.status, 401);
      if (refreshProbe.response.status === 200) {
        const descendant = await probeCopiedCookie(browser, refreshProbe.replacementCookie,
          refreshProbe.response.body.data.csrfToken);
        requireDelivery(descendant);
        assert.equal(descendant.response.status, 401, "logout failed to revoke rotated descendant");
      }
      // Cleanup above is mandatory, but never turns a failed original logout into PASS.
      requireCommittedLogout(logoutProbe);
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
    const witness = await startWitnessProxy();
    try {
      const allowedOrigin = new URL(FRONTEND).origin;
      const control = { reachable: false };
      const allowedContext = await browser.newContext();
      await check("allowed_origin_browser_api_and_backend_witness_control", async () => {
        const allowedPage = await openNeutralOriginPage(allowedContext, FRONTEND);
        const r = await request(allowedPage, "/api/v1/auth/register", {
          method: "POST", headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: `p11-allowed-${marker}@example.test` }),
        }, witness.url);
        assert.equal(r.transport, true); assert.equal(r.status, 202);
        assert.ok(witness.hits.some((h) => h.origin === allowedOrigin &&
          h.method === "POST" && h.upstreamReached && h.status === 202 && h.acao === allowedOrigin));
        control.reachable = true;
      });
      await allowedContext.close();
      if (!control.reachable) throw Error("EXTERNAL_EVIDENCE_REQUIRED: allowed-Origin positive control failed");
      const denied = await browser.newContext();
      try {
        const page = await openNeutralOriginPage(denied, ALT);
        await check("disallowed_origin_rejected_by_cors_with_backend_witness", async () => {
          const origin = new URL(ALT).origin;
          const from = witness.hits.length;
          const observation = await observeCorsRejection(page, () => request(page, "/api/v1/auth/register", {
            method: "POST", headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email: `p11-alt-${marker}@example.test` }),
          }, witness.url));
          requireCorsPolicyProof({ control, ...observation, origin,
            hits: witness.hits.slice(from).filter((h) => h.path === "/api/v1/auth/register") });
        });
      } finally { await denied.close(); }
      const nullPage = await ctx[0].newPage();
      await nullPage.goto("data:text/html,<title>P11</title>");
      await check("null_origin_rejected_by_cors_with_backend_witness", async () => {
        const from = witness.hits.length;
        const observation = await observeCorsRejection(nullPage, () => request(nullPage, "/api/v1/auth/register", {
          method: "POST", headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: `p11-null-${marker}@example.test` }),
        }, witness.url));
        requireCorsPolicyProof({ control, ...observation, origin: "null",
          hits: witness.hits.slice(from).filter((h) => h.path === "/api/v1/auth/register") });
      });
    } finally { await witness.close(); }
    await Promise.all(ctx.map((c) => c.close()));
  } finally { await browser.close(); }
  const failed = results.filter((r) => !r.ok).length;
  process.stdout.write(JSON.stringify({ summary: { checked: results.length, failed, browser: "Chromium", fixture: "DISPOSABLE" } }) + "\n");
  if (failed) process.exitCode = 1;
}

if (process.argv.length === 3 && process.argv[2] === "--negative-controls") {
  runNegativeControls();
} else {
  main().catch((error) => {
    console.error(error.message.startsWith("EXTERNAL_EVIDENCE_REQUIRED:") ? error.message : "P11_BROWSER_HARNESS_FAILED");
    process.exitCode = 2;
  });
}
