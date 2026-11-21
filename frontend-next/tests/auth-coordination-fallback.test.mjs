import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

// Different module instances deliberately emulate independent tab-local
// queues. Both see the same shared browser-level Web Locks availability.
const base = new URL("../src/features/auth/session-coordination.ts", import.meta.url);
const tabA = await import(`${base.href}?independent-tab=A`);
const tabB = await import(`${base.href}?independent-tab=B`);

function mockNavigator(t, value) {
  const before = Object.getOwnPropertyDescriptor(globalThis, "navigator");
  Object.defineProperty(globalThis, "navigator", { configurable: true, value });
  t.after(() => {
    if (before) Object.defineProperty(globalThis, "navigator", before);
    else delete globalThis.navigator;
  });
}

function mockBroadcastChannel(t, value) {
  const before = Object.getOwnPropertyDescriptor(globalThis, "BroadcastChannel");
  if (value === undefined) delete globalThis.BroadcastChannel;
  else Object.defineProperty(globalThis, "BroadcastChannel", { configurable: true, value });
  t.after(() => {
    if (before) Object.defineProperty(globalThis, "BroadcastChannel", before);
    else delete globalThis.BroadcastChannel;
  });
}

function sharedLocks() {
  let tail = Promise.resolve();
  let inCritical = 0;
  return {
    async request(name, options, operation) {
      assert.equal(name, "openinvest-auth-session-v1");
      assert.equal(options.mode, "exclusive");
      const task = tail.then(async () => {
        inCritical++;
        assert.equal(inCritical, 1, "overlapping protected operations");
        try { return await operation(); }
        finally { inCritical--; }
      });
      tail = task.then(() => {}, () => {});
      return task;
    },
  };
}

const userId = "user-1";
const response = (csrf) => ({
  ok: true,
  requestId: "synthetic",
  data: {
    user: { id: userId },
    session: { accessToken: "synthetic-access", csrfToken: csrf, accessTokenExpiresAt: "2030-01-01T00:00:00Z" },
  },
});

function fakeServer() {
  let active = true;
  let csrf = "csrf-start";
  let epoch = 0;
  let calls = 0;
  return {
    get mutations() { return calls; },
    get active() { return active; },
    async bootstrap() {
      if (!active) return { ok: false, status: 401, message: "Session invalid" };
      csrf = `csrf-bootstrap-${++epoch}`;
      return response(csrf);
    },
    async refresh(proof) {
      assert.equal(proof, csrf, "stale CSRF proof");
      calls++;
      csrf = `csrf-refresh-${++epoch}`;
      return { ok: true, requestId: "synthetic", data: response(csrf).data.session };
    },
    async logout(_payload, proof) {
      assert.equal(proof, csrf, "stale logout CSRF proof");
      calls++;
      active = false;
      return { ok: true, requestId: "synthetic", data: { revoked: true } };
    },
  };
}

test("A2-01: two independent tab modules without Web Locks never start simultaneous refresh", async (t) => {
  mockNavigator(t, { locks: undefined });
  const server = fakeServer();
  const [a, b] = await Promise.all([
    tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh)),
    tabB.withProtectedAuthLock(() => tabB.refreshWithCurrentProof(userId, server.bootstrap, server.refresh)),
  ]);
  assert.equal(a.ok, false);
  assert.equal(b.ok, false);
  assert.equal(a.failureKind, "REQUEST_SETUP");
  assert.equal(b.failureKind, "REQUEST_SETUP");
  assert.equal(server.mutations, 0);
  assert.equal(server.active, true);
});

test("A2-01: refresh versus logout without Web Locks cannot invoke either mutation", async (t) => {
  mockNavigator(t, { locks: undefined });
  const server = fakeServer();
  const [a, b] = await Promise.all([
    tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh)),
    tabB.withProtectedAuthLock(() => tabB.logoutWithCurrentProof(userId, { allSessions: false }, server.bootstrap, server.logout)),
  ]);
  assert.equal(a.ok, false);
  assert.equal(b.ok, false);
  assert.equal(server.mutations, 0);
});

test("A2-01: missing BroadcastChannel cannot re-enable uncoordinated protected writes", async (t) => {
  mockNavigator(t, { locks: undefined });
  mockBroadcastChannel(t, undefined);
  const server = fakeServer();
  tabA.announceAuthChange();
  const a = await tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  const b = await tabB.withProtectedAuthLock(() => tabB.logoutWithCurrentProof(userId, { allSessions: false }, server.bootstrap, server.logout));
  assert.equal(a.ok, false);
  assert.equal(b.ok, false);
  assert.equal(server.mutations, 0);
});

test("A2-01: restoration of Web Locks enables protected operations after blocked attempts", async (t) => {
  mockNavigator(t, { locks: undefined });
  const server = fakeServer();
  const blocked = await tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  assert.equal(blocked.ok, false);
  assert.equal(server.mutations, 0);
  Object.defineProperty(globalThis, "navigator", { configurable: true, value: { locks: sharedLocks() } });
  const recovered = await tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  assert.equal(recovered.ok, true);
  assert.equal(server.mutations, 1);
});

test("A2-01: lost BroadcastChannel message does not cause replay when Web Locks are available", async (t) => {
  mockNavigator(t, { locks: sharedLocks() });
  mockBroadcastChannel(t, undefined);
  const server = fakeServer();
  const first = await tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  // Intentionally do not broadcast. A different module/tab has stale React state.
  const second = await tabB.withProtectedAuthLock(() => tabB.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  assert.equal(first.ok, true);
  assert.equal(second.ok, true);
  assert.equal(server.mutations, 2);
  assert.equal(server.active, true);
});

test("A2-01: failed bootstrap releases lock; a later independent tab can recover", async (t) => {
  mockNavigator(t, { locks: sharedLocks() });
  const server = fakeServer();
  let attempts = 0;
  const flakyBootstrap = async () => ++attempts === 1
    ? { ok: false, status: 503, message: "Backend unavailable" }
    : server.bootstrap();
  const [first, second] = await Promise.all([
    tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, flakyBootstrap, server.refresh)),
    tabB.withProtectedAuthLock(() => tabB.refreshWithCurrentProof(userId, flakyBootstrap, server.refresh)),
  ]);
  assert.equal(first.status, 503);
  assert.equal(second.ok, true);
  assert.equal(server.mutations, 1);
});

test("A2-01: logout before refresh keeps the session revoked across tabs", async (t) => {
  mockNavigator(t, { locks: sharedLocks() });
  const server = fakeServer();
  const [logout, refresh] = await Promise.all([
    tabA.withProtectedAuthLock(() => tabA.logoutWithCurrentProof(userId, { allSessions: false }, server.bootstrap, server.logout)),
    tabB.withProtectedAuthLock(() => tabB.refreshWithCurrentProof(userId, server.bootstrap, server.refresh)),
  ]);
  assert.equal(logout.ok, true);
  assert.equal(refresh.status, 401);
  assert.equal(server.mutations, 1);
  assert.equal(server.active, false);
});

test("A2-01: AuthShell uses the fail-closed coordinator for both protected mutations", async () => {
  const shell = await readFile(new URL("../src/features/auth/components/AuthShell.tsx", import.meta.url), "utf8");
  assert.match(shell, /return withProtectedAuthLock\(async \(\) => \{\s*const result = await refreshWithCurrentProof\(/);
  assert.match(shell, /return withProtectedAuthLock\(async \(\) => \{\s*const result = await logoutWithCurrentProof\(/);
});

test("A2-02: OpenAPI includes explicit Bootstrap 403 and 500 response mappings", async () => {
  const spec = await readFile(new URL("../../openapi/openapi.yaml", import.meta.url), "utf8");
  const responses = await readFile(new URL("../../openapi/components/responses.yaml", import.meta.url), "utf8");
  const section = spec.split("  /api/v1/auth/bootstrap:\n")[1]?.split("  /api/v1/auth/refresh:\n")[0];
  assert.ok(section, "bootstrap OpenAPI path missing");
  assert.match(section, /"403": \{ \$ref: "\.\/components\/responses\.yaml#\/responses\/Forbidden" \}/);
  assert.match(section, /"500": \{ \$ref: "\.\/components\/responses\.yaml#\/responses\/InternalError" \}/);
  assert.match(responses, /^  Forbidden:$/m);
  assert.match(responses, /^  InternalError:$/m);
});


test("A2-01: rejected Web Locks acquisition reports uncertainty without replay", async (t) => {
  mockNavigator(t, { locks: { async request() { throw new Error("Lock refused"); } } });
  const server = fakeServer();
  const result = await tabA.withProtectedAuthLock(() => tabA.refreshWithCurrentProof(userId, server.bootstrap, server.refresh));
  assert.equal(result.ok, false);
  assert.equal(result.completion, "UNKNOWN");
  assert.equal(server.mutations, 0);
});
