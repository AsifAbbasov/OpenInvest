import assert from "node:assert/strict";
import test from "node:test";

const session = await import("../src/features/auth/session.ts");

const userA = {
  id: "user-a",
  email: "a@example.test",
  language: "en",
  theme: "system",
  timezone: "UTC",
  privacyMode: true,
  createdAt: "2026-10-13T00:00:00Z",
};
const userB = { ...userA, id: "user-b", email: "b@example.test" };

const sessionA = {
  accessToken: "access-a",
  accessTokenExpiresAt: "2026-10-13T00:15:00Z",
  csrfToken: "csrf-a",
};
const rotatedA = {
  accessToken: "access-a-rotated",
  accessTokenExpiresAt: "2026-10-13T00:30:00Z",
  csrfToken: "csrf-a-rotated",
};
const sessionB = {
  accessToken: "access-b",
  accessTokenExpiresAt: "2026-10-13T00:20:00Z",
  csrfToken: "csrf-b",
};

function runtime(initial = session.authenticatedState({ user: userA, session: sessionA })) {
  return {
    state: initial,
    operation: { generation: 0, pending: null },
    pending: null,
    getState() { return this.state; },
    getOperation() { return this.operation; },
    setOperation(value) { this.operation = value; },
    setPendingOperation(value) { this.pending = value; },
    setState(value) { this.state = typeof value === "function" ? value(this.state) : value; },
  };
}

test("Module1 transient refresh errors preserve authenticated principal and token", () => {
  for (const failure of [
    { ok: false, status: 429, message: "rate limited" },
    { ok: false, status: 500, message: "internal" },
    { ok: false, status: 503, message: "unavailable" },
    { ok: false, message: "network" },
  ]) {
    const before = session.authenticatedState({ user: userA, session: sessionA });
    const after = session.applyRefreshResult(before, failure);
    assert.equal(after.status, "authenticated");
    assert.equal(after.user.id, userA.id);
    assert.equal(after.session.accessToken, sessionA.accessToken);
  }
});

test("Module1 concurrent refresh admission remains single-flight", async () => {
  const rt = runtime();
  let resolveFirst;
  let calls = 0;
  const client = {
    async refreshSession() {
      calls++;
      return await new Promise((resolve) => { resolveFirst = resolve; });
    },
  };
  const first = session.refreshActiveSession(rt, client);
  const second = await session.refreshActiveSession(rt, client);
  assert.equal(second.ok, false);
  assert.equal(second.message, "Session operation already in progress.");
  assert.equal(calls, 1);
  resolveFirst({ ok: true, requestId: "r1", data: rotatedA });
  const result = await first;
  assert.equal(result.ok, true);
  assert.equal(rt.state.status, "authenticated");
  assert.equal(rt.state.session.accessToken, rotatedA.accessToken);
});

test("Module1 late refresh for principal A cannot overwrite a later principal B login", async () => {
  const rt = runtime();
  let resolveRefresh;
  const refreshPromise = session.refreshActiveSession(rt, {
    async refreshSession(csrfToken) {
      assert.equal(csrfToken, sessionA.csrfToken);
      return await new Promise((resolve) => { resolveRefresh = resolve; });
    },
  });
  assert.equal(rt.pending, "refresh");

  rt.setOperation({ generation: rt.getOperation().generation + 1, pending: null });
  rt.setPendingOperation(null);
  rt.setState(session.authenticatedState({ user: userB, session: sessionB }));
  resolveRefresh({ ok: true, requestId: "late-a", data: rotatedA });
  await refreshPromise;

  assert.equal(rt.state.status, "authenticated");
  assert.equal(rt.state.user.id, userB.id, "OLD_PRINCIPAL_RESPONSE_OVERWRITES_NEW_PRINCIPAL");
  assert.equal(rt.state.session.accessToken, sessionB.accessToken, "STALE_ACCESS_TOKEN_RESTORED");
  assert.equal(rt.state.session.csrfToken, sessionB.csrfToken, "STALE_CSRF_TOKEN_RESTORED");
});

test("Module1 late logout for principal A cannot clear a later principal B login", async () => {
  const rt = runtime();
  let resolveLogout;
  const logoutPromise = session.logoutActiveSession(rt, {
    async logout(payload, csrfToken) {
      assert.deepEqual(payload, { allSessions: false });
      assert.equal(csrfToken, sessionA.csrfToken);
      return await new Promise((resolve) => { resolveLogout = resolve; });
    },
  });
  assert.equal(rt.pending, "logout");

  rt.setOperation({ generation: rt.getOperation().generation + 1, pending: null });
  rt.setPendingOperation(null);
  rt.setState(session.authenticatedState({ user: userB, session: sessionB }));
  resolveLogout({ ok: true, requestId: "late-logout-a", data: { revoked: true } });
  await logoutPromise;

  assert.equal(rt.state.status, "authenticated", "STALE_LOGOUT_CLEARED_NEW_PRINCIPAL");
  assert.equal(rt.state.user.id, userB.id);
  assert.equal(rt.state.session.accessToken, sessionB.accessToken);
});

test("Module1 stale refresh after successful logout never restores authentication", async () => {
  const rt = runtime();
  let resolveRefresh;
  const refreshPromise = session.refreshActiveSession(rt, {
    async refreshSession() {
      return await new Promise((resolve) => { resolveRefresh = resolve; });
    },
  });
  const logoutResult = await session.logoutActiveSession(rt, {
    async logout() {
      return { ok: true, requestId: "logout", data: { revoked: true } };
    },
  });
  assert.equal(logoutResult.ok, true);
  assert.equal(rt.state.status, "anonymous");

  resolveRefresh({ ok: true, requestId: "late-refresh", data: rotatedA });
  await refreshPromise;
  assert.equal(rt.state.status, "anonymous", "LOGOUT_UNDONE_BY_LATE_RESPONSE");
});


test("Module1 negative control proves stale-operation oracle detects missing generation transition", () => {
  const current = { generation: 1, pending: "refresh" };
  assert.equal(session.isCurrentOperation(current, 1), true);
  const principalTransition = { generation: 2, pending: null };
  assert.equal(session.isCurrentOperation(principalTransition, 1), false, "NEGATIVE_CONTROL_DETECTED");
});
