const { chromium } = require("playwright");

const FRONTEND = process.env.M12_FRONTEND_URL || "http://localhost:3000";
const FRONTEND_ALT = process.env.M12_FRONTEND_ALT_URL || "http://127.0.0.1:3000";
const BACKEND = process.env.M12_BACKEND_URL || "http://localhost:18080";

const results = [];
let failures = 0;

function record(name, ok, evidence = {}) {
  results.push({ name, ok, evidence });
  if (!ok) failures++;
  process.stdout.write(JSON.stringify({ name, ok, evidence }) + "\n");
}

async function browserFetch(page, path, init = {}) {
  return await page.evaluate(async ({ backend, path, init }) => {
    try {
      const response = await fetch(backend + path, init);
      const text = await response.text();
      let body = null;
      try { body = JSON.parse(text); } catch { body = text; }
      return {
        ok: true,
        status: response.status,
        headers: Object.fromEntries(response.headers.entries()),
        body,
      };
    } catch (error) {
      return { ok: false, error: String(error) };
    }
  }, { backend: BACKEND, path, init });
}

function sessionFrom(body) {
  return body && body.data && body.data.session ? body.data.session : null;
}

function userFrom(body) {
  return body && body.data && body.data.user ? body.data.user : null;
}

async function register(page, email) {
  return await browserFetch(page, "/api/v1/auth/register", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      email,
      password: "correct horse battery staple",
      language: "en",
      theme: "system",
      timezone: "UTC",
    }),
  });
}

async function refresh(page, csrf) {
  return await browserFetch(page, "/api/v1/auth/refresh", {
    method: "POST",
    credentials: "include",
    headers: { "X-CSRF-Token": csrf, "Content-Type": "application/json" },
  });
}

async function logout(page, csrf) {
  return await browserFetch(page, "/api/v1/auth/logout", {
    method: "POST",
    credentials: "include",
    headers: { "X-CSRF-Token": csrf, "Content-Type": "application/json" },
    body: JSON.stringify({ allSessions: false }),
  });
}

async function createPortfolio(page, token, name) {
  return await browserFetch(page, "/api/v1/portfolios", {
    method: "POST",
    credentials: "include",
    headers: {
      "Authorization": "Bearer " + token,
      "Content-Type": "application/json",
      "Idempotency-Key": crypto.randomUUID(),
    },
    body: JSON.stringify({ name, baseCurrency: "RUB" }),
  });
}

async function listPortfolios(page, token) {
  return await browserFetch(page, "/api/v1/portfolios", {
    method: "GET",
    credentials: "include",
    headers: { "Authorization": "Bearer " + token, "Content-Type": "application/json" },
  });
}

async function main() {
  const browser = await chromium.launch({ headless: true });
  try {
    const contextA = await browser.newContext();
    const contextB = await browser.newContext();
    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();

    await pageA.goto(FRONTEND, { waitUntil: "domcontentloaded" });
    await pageB.goto(FRONTEND, { waitUntil: "domcontentloaded" });

    const stamp = Date.now();
    const regA = await register(pageA, "m12-a-" + stamp + "@example.test");
    const regB = await register(pageB, "m12-b-" + stamp + "@example.test");
    record("A_login_register", regA.ok && regA.status === 201, { status: regA.status, transport: regA.ok });
    record("B_login_register", regB.ok && regB.status === 201, { status: regB.status, transport: regB.ok });

    const sessionA = sessionFrom(regA.body);
    const sessionB = sessionFrom(regB.body);
    const userA = userFrom(regA.body);
    const userB = userFrom(regB.body);
    if (!sessionA || !sessionB || !userA || !userB) {
      record("session_material_present", false, {});
      process.exitCode = 1;
      return;
    }
    record("distinct_principals", userA.id !== userB.id, { a: userA.id, b: userB.id });

    const pa = await createPortfolio(pageA, sessionA.accessToken, "A-" + stamp);
    const pb = await createPortfolio(pageB, sessionB.accessToken, "B-" + stamp);
    record("A_protected_mutation", pa.ok && pa.status === 201, { status: pa.status });
    record("B_protected_mutation", pb.ok && pb.status === 201, { status: pb.status });

    const crossA = await listPortfolios(pageB, sessionA.accessToken);
    const crossB = await listPortfolios(pageA, sessionB.accessToken);
    const aText = JSON.stringify(crossA.body);
    const bText = JSON.stringify(crossB.body);
    record("A_access_token_with_B_cookie", crossA.ok && crossA.status === 200 && aText.includes("A-" + stamp) && !aText.includes("B-" + stamp), { status: crossA.status });
    record("B_access_token_with_A_cookie", crossB.ok && crossB.status === 200 && bText.includes("B-" + stamp) && !bText.includes("A-" + stamp), { status: crossB.status });

    const raceContext = await browser.newContext();
    const racePage = await raceContext.newPage();
    await racePage.goto(FRONTEND, { waitUntil: "domcontentloaded" });
    const raceReg = await register(racePage, "m12-race-" + stamp + "@example.test");
    const raceSession = sessionFrom(raceReg.body);
    if (raceSession) {
      const rr = await Promise.all([refresh(racePage, raceSession.csrfToken), refresh(racePage, raceSession.csrfToken)]);
      const successCount = rr.filter(x => x.ok && x.status >= 200 && x.status < 300).length;
      record("A_refresh_parallel_A_refresh", successCount <= 1, { statuses: rr.map(x => x.status), successes: successCount });
    } else {
      record("A_refresh_parallel_A_refresh", false, { reason: "fixture registration failed" });
    }

    const logoutContext = await browser.newContext();
    const logoutPage = await logoutContext.newPage();
    await logoutPage.goto(FRONTEND, { waitUntil: "domcontentloaded" });
    const logoutReg = await register(logoutPage, "m12-logout-" + stamp + "@example.test");
    const logoutSession = sessionFrom(logoutReg.body);
    if (logoutSession) {
      const lr = await Promise.all([refresh(logoutPage, logoutSession.csrfToken), logout(logoutPage, logoutSession.csrfToken)]);
      record("A_refresh_parallel_A_logout", lr.every(x => x.ok) && lr.some(x => x.status >= 400), { statuses: lr.map(x => x.status) });
      const stale = await refresh(logoutPage, logoutSession.csrfToken);
      record("LOGOUT_DOES_NOT_RESTORE_STALE_SESSION", stale.ok && stale.status === 401, { status: stale.status });
    } else {
      record("A_refresh_parallel_A_logout", false, { reason: "fixture registration failed" });
    }

    const cookiesA = await contextA.cookies(BACKEND);
    const refreshCookieA = cookiesA.find(c => /refresh/i.test(c.name));
    if (refreshCookieA) {
      const crossContext = await browser.newContext();
      await crossContext.addCookies([refreshCookieA]);
      const crossPage = await crossContext.newPage();
      await crossPage.goto(FRONTEND, { waitUntil: "domcontentloaded" });
      const mixed = await refresh(crossPage, sessionB.csrfToken);
      record("A_refresh_cookie_in_B_context", mixed.ok && mixed.status === 401, { status: mixed.status });
      await crossContext.close();
    } else {
      record("A_refresh_cookie_in_B_context", false, { reason: "refresh cookie not observed by browser context" });
    }

    const disallowedContext = await browser.newContext();
    const disallowedPage = await disallowedContext.newPage();
    await disallowedPage.goto(FRONTEND_ALT, { waitUntil: "domcontentloaded" });
    const disallowed = await browserFetch(disallowedPage, "/api/v1/auth/register", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "m12-disallowed-" + stamp + "@example.test",
        password: "correct horse battery staple",
        language: "en",
        theme: "system",
        timezone: "UTC",
      }),
    });
    record("disallowed_origin_browser_fetch", !disallowed.ok || !disallowed.headers || !disallowed.headers["access-control-allow-origin"], { transport: disallowed.ok, status: disallowed.status });

    const nullPage = await contextA.newPage();
    await nullPage.goto("data:text/html,<html><body>m12</body></html>");
    const nullOrigin = await browserFetch(nullPage, "/api/v1/auth/register", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    });
    record("Origin_null", !nullOrigin.ok || !nullOrigin.headers || !nullOrigin.headers["access-control-allow-origin"], { transport: nullOrigin.ok, status: nullOrigin.status });

    const allowedPreflight = await contextA.request.fetch(BACKEND + "/api/v1/auth/register", {
      method: "OPTIONS",
      headers: {
        "Origin": FRONTEND,
        "Access-Control-Request-Method": "POST",
        "Access-Control-Request-Headers": "content-type,x-csrf-token",
      },
    });
    record("valid_preflight", allowedPreflight.status() === 204 && allowedPreflight.headers()["access-control-allow-origin"] === FRONTEND, { status: allowedPreflight.status() });

    const deniedPreflight = await contextA.request.fetch(BACKEND + "/api/v1/auth/register", {
      method: "OPTIONS",
      headers: {
        "Origin": "https://evil.example.test",
        "Access-Control-Request-Method": "POST",
      },
    });
    record("disallowed_preflight", deniedPreflight.status() === 204 && !deniedPreflight.headers()["access-control-allow-origin"], { status: deniedPreflight.status() });

    const contentTypes = [
      ["application/json", JSON.stringify({ name: "csrf-json", baseCurrency: "RUB" })],
      ["text/plain", "name=csrf-plain"],
      ["application/x-www-form-urlencoded", "name=csrf-form&baseCurrency=RUB"],
      ["multipart/form-data; boundary=m12", "--m12\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\ncsrf\r\n--m12--\r\n"],
    ];
    for (const [contentType, body] of contentTypes) {
      const res = await browserFetch(pageA, "/api/v1/portfolios", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": contentType },
        body,
      });
      record("unsafe_mutation_without_bearer_" + contentType.split(";")[0], res.ok && res.status === 401, { status: res.status });
    }

    await contextA.close();
    await contextB.close();
    await raceContext.close();
    await logoutContext.close();
    await disallowedContext.close();
  } finally {
    await browser.close();
  }

  process.stdout.write(JSON.stringify({
    summary: {
      cases: results.length,
      failures,
      NO_CROSS_PRINCIPAL_SESSION_CONFUSION: failures === 0 ? "YES" : "CHECK_RESULTS",
      NO_CROSS_ORIGIN_CREDENTIAL_BYPASS: failures === 0 ? "YES" : "CHECK_RESULTS",
      NO_UNSAFE_MUTATION_CSRF: failures === 0 ? "YES" : "CHECK_RESULTS",
    }
  }) + "\n");
  if (failures > 0) process.exitCode = 1;
}

main().catch(error => {
  console.error(error);
  process.exit(1);
});
