import assert from "node:assert/strict";
import test from "node:test";

const session = await import("../src/features/auth/session.ts");

const user = {
  id: "user-id",
  email: "investor@example.com",
  language: "en",
  theme: "system",
  timezone: "UTC",
  privacyMode: true,
  createdAt: "2026-07-11T00:00:00Z",
};

const firstSession = {
  accessToken: "first-access-token",
  accessTokenExpiresAt: "2026-07-11T00:15:00Z",
  csrfToken: "first-csrf-token",
};

const rotatedSession = {
  accessToken: "rotated-access-token",
  accessTokenExpiresAt: "2026-07-11T00:30:00Z",
  csrfToken: "rotated-csrf-token",
};

function makeRuntime(initialState = session.authenticatedState({ user, session: firstSession })) {
  return {
    state: initialState,
    operation: { generation: 0, pending: null },
    pending: null,
    getState() { return this.state; },
    getOperation() { return this.operation; },
    setOperation(operation) { this.operation = operation; },
    setPendingOperation(operation) { this.pending = operation; },
    setState(nextState) {
      this.state = typeof nextState === "function" ? nextState(this.state) : nextState;
    },
  };
}

test("auth result creates authenticated in-memory state without refresh token fields", () => {
  const state = session.applyAuthResult({
    ok: true,
    requestId: "request-id",
    data: { user, session: firstSession },
  });
  assert.equal(state.status, "authenticated");
  assert.equal(state.session.accessToken, "first-access-token");
  assert.equal("refreshToken" in state.session, false);
});

test("refresh success rotates only the active in-memory session", () => {
  const state = session.authenticatedState({ user, session: firstSession });
  const refreshed = session.applyRefreshResult(state, {
    ok: true,
    requestId: "request-id",
    data: rotatedSession,
  });
  assert.equal(refreshed.status, "authenticated");
  assert.equal(refreshed.session.accessToken, "rotated-access-token");
  assert.equal(refreshed.session.csrfToken, "rotated-csrf-token");
});

test("refresh 401 clears authenticated state as a true auth rejection", () => {
  const state = session.authenticatedState({ user, session: firstSession });
  const refreshed = session.applyRefreshResult(state, {
    ok: false,
    status: 401,
    message: "Invalid or expired session",
  });
  assert.equal(refreshed.status, "anonymous");
  assert.equal(refreshed.message, "Session expired. Sign in again.");
});

for (const [label, failure] of [
  ["500", { ok: false, status: 500, message: "Internal server error" }],
  ["503", { ok: false, status: 503, message: "Service unavailable" }],
  ["429", { ok: false, status: 429, message: "Too many authentication attempts" }],
  ["network failure", { ok: false, message: "OpenInvest is temporarily unavailable. Please try again." }],
]) {
  test(`refresh ${label} preserves authenticated state and current tokens`, () => {
    const state = session.authenticatedState({ user, session: firstSession });
    const refreshed = session.applyRefreshResult(state, failure);
    assert.equal(refreshed.status, "authenticated");
    assert.equal(refreshed.user.id, user.id);
    assert.equal(refreshed.session.accessToken, firstSession.accessToken);
    assert.equal(refreshed.session.csrfToken, firstSession.csrfToken);
    assert.equal(refreshed.message, failure.message);
  });
}

test("refresh can be retried after a transient failure", async () => {
  const runtime = makeRuntime();
  let calls = 0;
  const client = {
    async refreshSession(csrfToken) {
      calls += 1;
      assert.equal(csrfToken, firstSession.csrfToken);
      if (calls === 1) return { ok: false, status: 503, message: "Service unavailable" };
      return { ok: true, requestId: "request-id", data: rotatedSession };
    },
  };
  const failed = await session.refreshActiveSession(runtime, client);
  assert.equal(failed.ok, false);
  assert.equal(runtime.state.status, "authenticated");
  assert.equal(runtime.state.session.accessToken, firstSession.accessToken);
  assert.equal(runtime.pending, null);

  const retried = await session.refreshActiveSession(runtime, client);
  assert.equal(retried.ok, true);
  assert.equal(runtime.state.status, "authenticated");
  assert.equal(runtime.state.session.accessToken, rotatedSession.accessToken);
  assert.equal(runtime.state.session.csrfToken, rotatedSession.csrfToken);
  assert.equal(runtime.pending, null);
  assert.equal(calls, 2);
});

test("logout success clears authenticated state", async () => {
  const runtime = makeRuntime();
  const result = await session.logoutActiveSession(runtime, {
    async logout() {
      return { ok: true, requestId: "request-id", data: { revoked: true } };
    },
  });
  assert.equal(result.ok, true);
  assert.equal(runtime.state.status, "anonymous");
  assert.equal(runtime.state.message, "Signed out.");
});

for (const [label, failure] of [
  ["401", { ok: false, status: 401, message: "Invalid or expired session" }],
  ["429", { ok: false, status: 429, message: "Too many authentication attempts" }],
  ["500", { ok: false, status: 500, message: "Internal server error" }],
  ["network failure", { ok: false, message: "OpenInvest is temporarily unavailable. Please try again." }],
]) {
  test(`logout ${label} preserves authenticated state and does not claim sign-out`, async () => {
    const runtime = makeRuntime();
    const result = await session.logoutActiveSession(runtime, {
      async logout(payload, csrfToken) {
        assert.deepEqual(payload, { allSessions: false });
        assert.equal(csrfToken, firstSession.csrfToken);
        return failure;
      },
    });
    assert.equal(result.ok, false);
    assert.equal(runtime.state.status, "authenticated");
    assert.equal(runtime.state.session.accessToken, firstSession.accessToken);
    assert.equal(runtime.state.session.csrfToken, firstSession.csrfToken);
    assert.equal(runtime.state.message, failure.message);
    assert.notEqual(runtime.state.message, "Signed out.");
    assert.equal(runtime.pending, null);
  });
}

test("logout generation supersedes older refresh responses", () => {
  const initial = { generation: 0, pending: null };
  const refreshOperation = session.startSessionOperation(initial, "refresh");
  const logoutOperation = session.startSessionOperation(refreshOperation, "logout");
  assert.equal(session.isCurrentOperation(logoutOperation, refreshOperation.generation), false);
  assert.equal(session.isCurrentOperation(logoutOperation, logoutOperation.generation), true);
  const afterOldRefresh = session.finishSessionOperation(logoutOperation, refreshOperation.generation);
  assert.deepEqual(afterOldRefresh, logoutOperation);
});

test("current operation completion clears pending state", () => {
  const current = session.startSessionOperation({ generation: 4, pending: null }, "logout");
  const completed = session.finishSessionOperation(current, current.generation);
  assert.equal(completed.generation, current.generation);
  assert.equal(completed.pending, null);
});

test("auth shell session controller does not restore auth when refresh resolves after logout", async () => {
  let resolveRefresh;
  const clientCalls = [];
  const runtime = makeRuntime();
  const client = {
    async refreshSession(csrfToken) {
      clientCalls.push({ operation: "refresh", csrfToken });
      return await new Promise((resolve) => {
        resolveRefresh = () => resolve({
          ok: true,
          requestId: "11111111-1111-4111-8111-111111111111",
          data: rotatedSession,
        });
      });
    },
    async logout(payload, csrfToken) {
      clientCalls.push({ operation: "logout", payload, csrfToken });
      return { ok: true, requestId: "33333333-3333-4333-8333-333333333333", data: { revoked: true } };
    },
  };

  const refreshPromise = session.refreshActiveSession(runtime, client);
  assert.equal(runtime.pending, "refresh");

  const logoutResult = await session.logoutActiveSession(runtime, client);
  assert.equal(logoutResult.ok, true);
  assert.equal(runtime.state.status, "anonymous");
  assert.equal(runtime.pending, null);

  resolveRefresh();
  const refreshResult = await refreshPromise;
  assert.equal(refreshResult.ok, true);
  assert.equal(runtime.state.status, "anonymous");
  assert.equal(runtime.pending, null);
  assert.equal(runtime.operation.generation, 2);
  assert.deepEqual(clientCalls, [
    { operation: "refresh", csrfToken: "first-csrf-token" },
    { operation: "logout", payload: { allSessions: false }, csrfToken: "first-csrf-token" },
  ]);
});
