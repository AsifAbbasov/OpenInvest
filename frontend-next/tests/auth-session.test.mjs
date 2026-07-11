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

test("refresh failure clears authenticated state", () => {
  const state = session.authenticatedState({ user, session: firstSession });
  const refreshed = session.applyRefreshResult(state, {
    ok: false,
    status: 401,
    message: "Invalid or expired session",
  });

  assert.equal(refreshed.status, "anonymous");
  assert.equal(refreshed.message, "Session expired. Sign in again.");
});

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
  const runtime = {
    state: session.authenticatedState({ user, session: firstSession }),
    operation: { generation: 0, pending: null },
    pending: null,
    getState() {
      return this.state;
    },
    getOperation() {
      return this.operation;
    },
    setOperation(operation) {
      this.operation = operation;
    },
    setPendingOperation(operation) {
      this.pending = operation;
    },
    setState(nextState) {
      this.state = typeof nextState === "function" ? nextState(this.state) : nextState;
    },
  };

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
