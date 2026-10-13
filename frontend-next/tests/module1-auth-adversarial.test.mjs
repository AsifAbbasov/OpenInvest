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

const userB = {
  ...userA,
  id: "user-b",
  email: "b@example.test",
};

const sessionA = {
  accessToken: "access-a",
  accessTokenExpiresAt: "2026-10-13T00:15:00Z",
  csrfToken: "csrf-a",
};

const sessionARotated = {
  accessToken: "access-a-rotated",
  accessTokenExpiresAt: "2026-10-13T00:30:00Z",
  csrfToken: "csrf-a-rotated",
};

const sessionB = {
  accessToken: "access-b",
  accessTokenExpiresAt: "2026-10-13T00:20:00Z",
  csrfToken: "csrf-b",
};

function makeRuntime(initialState) {
  return {
    state: initialState,
    operation: { generation: 0, pending: null },
    pending: null,
    getState() { return this.state; },
    getOperation() { return this.operation; },
    setOperation(value) { this.operation = value; },
    setPendingOperation(value) { this.pending = value; },
    setState(next) { this.state = typeof next === "function" ? next(this.state) : next; },
  };
}

test("Module1 stale refresh from principal A must not overwrite principal B session", async () => {
  const runtime = makeRuntime(session.authenticatedState({ user: userA, session: sessionA }));
  let resolveRefresh;
  const refreshPromise = session.refreshActiveSession(runtime, {
    async refreshSession(csrfToken) {
      assert.equal(csrfToken, "csrf-a");
      return await new Promise((resolve) => {
        resolveRefresh = () => resolve({
          ok: true,
          requestId: "11111111-1111-4111-8111-111111111111",
          data: sessionARotated,
        });
      });
    },
  });

  assert.equal(runtime.pending, "refresh");

  // Model the real AuthShell.handleAuth transition: it increments operation generation
  // before replacing the authenticated principal, invalidating the older refresh.
  runtime.setOperation({
    generation: runtime.getOperation().generation + 1,
    pending: null,
  });
  runtime.setPendingOperation(null);
  runtime.setState(session.authenticatedState({ user: userB, session: sessionB }));
  assert.equal(runtime.state.user.id, "user-b");
  assert.equal(runtime.state.session.accessToken, "access-b");

  resolveRefresh();
  const result = await refreshPromise;
  assert.equal(result.ok, true);

  assert.equal(runtime.state.status, "authenticated");
  assert.equal(runtime.state.user.id, "user-b");
  assert.equal(
    runtime.state.session.accessToken,
    "access-b",
    "late refresh from principal A must not replace principal B session",
  );
  assert.equal(runtime.state.session.csrfToken, "csrf-b");
});

test("Module1 stale refresh after explicit anonymous transition cannot restore auth", async () => {
  const runtime = makeRuntime(session.authenticatedState({ user: userA, session: sessionA }));
  let resolveRefresh;
  const promise = session.refreshActiveSession(runtime, {
    async refreshSession() {
      return await new Promise((resolve) => {
        resolveRefresh = () => resolve({
          ok: true,
          requestId: "22222222-2222-4222-8222-222222222222",
          data: sessionARotated,
        });
      });
    },
  });

  runtime.setState(session.anonymousState("Signed out."));
  resolveRefresh();
  await promise;
  assert.equal(runtime.state.status, "anonymous");
});

test("Module1 transient refresh failures preserve current principal and tokens", () => {
  for (const failure of [
    { ok: false, status: 429, message: "Too many authentication attempts" },
    { ok: false, status: 500, message: "Internal server error" },
    { ok: false, status: 503, message: "Service unavailable" },
    { ok: false, message: "Network unavailable" },
  ]) {
    const current = session.authenticatedState({ user: userA, session: sessionA });
    const next = session.applyRefreshResult(current, failure);
    assert.equal(next.status, "authenticated");
    assert.equal(next.user.id, "user-a");
    assert.equal(next.session.accessToken, "access-a");
    assert.equal(next.session.csrfToken, "csrf-a");
  }
});

test("Module1 frontend oracle negative control detects a wrong principal expectation", () => {
  const state = session.authenticatedState({ user: userA, session: sessionA });
  assert.throws(() => assert.equal(state.user.id, "user-b"));
});
