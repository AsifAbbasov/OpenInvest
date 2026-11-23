import assert from "node:assert/strict";
import test from "node:test";
import {
  announceAuthChange,
  isSessionChangeMessage,
  logoutWithCurrentProof,
  refreshWithCurrentProof,
  subscribeAuthChanges,
  withAuthLock,
} from "../src/features/auth/session-coordination.ts";

const user = { id: "user-1" };
const ok = (csrf) => ({
  ok: true,
  requestId: "req",
  data: { user, session: { accessToken: "access", csrfToken: csrf, accessTokenExpiresAt: "2030-01-01T00:00:00Z" } },
});

// Two simulated tabs share a server-side cookie and a Web Locks queue.
function installLocks(t) {
  const before = Object.getOwnPropertyDescriptor(globalThis, "navigator");
  let held = Promise.resolve();
  const invocations = [];
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: { locks: {
      request(name, options, callback) {
        assert.equal(name, "openinvest-auth-session-v1");
        assert.equal(options.mode, "exclusive");
        const work = held.then(async () => {
          invocations.push("lock-enter");
          try { return await callback(); }
          finally { invocations.push("lock-exit"); }
        });
        held = work.then(() => {}, () => {});
        return work;
      },
    } },
  });
  t.after(() => {
    if (before) Object.defineProperty(globalThis, "navigator", before);
    else delete globalThis.navigator;
  });
  return invocations;
}

test("Web Locks serialize two concurrent bootstrap/refresh operations across tabs", async (t) => {
  const order = installLocks(t);
  let serverCsrf = "csrf-initial";
  let cookieVersion = 0;
  async function bootstrap() {
    serverCsrf = `csrf-bootstrap-${++cookieVersion}`;
    order.push(`bootstrap-${cookieVersion}`);
    return ok(serverCsrf);
  }
  async function refresh(csrf) {
    assert.equal(csrf, serverCsrf, "must use current server CSRF");
    order.push(`refresh-${cookieVersion}`);
    await Promise.resolve();
    serverCsrf = `csrf-rotated-${cookieVersion}`;
    return { ok: true, requestId: "req", data: ok(serverCsrf).data.session };
  }
  const [a, b] = await Promise.all([
    withAuthLock(() => refreshWithCurrentProof(user.id, bootstrap, refresh)),
    withAuthLock(() => refreshWithCurrentProof(user.id, bootstrap, refresh)),
  ]);
  assert.equal(a.ok, true);
  assert.equal(b.ok, true);
  assert.deepEqual(order, [
    "lock-enter", "bootstrap-1", "refresh-1", "lock-exit",
    "lock-enter", "bootstrap-2", "refresh-2", "lock-exit",
  ]);
});

test("logout following a concurrent refresh obtains the latest cookie/CSRF proof", async (t) => {
  installLocks(t);
  let csrf = "initial";
  let loggedOut = false;
  const bootstrap = async () => {
    if (loggedOut) return { ok: false, status: 401, message: "Invalid session" };
    csrf = `${csrf}:bootstrap`;
    return ok(csrf);
  };
  const refresh = async (proof) => {
    assert.equal(proof, csrf);
    csrf = "refreshed";
    return { ok: true, requestId: "req", data: ok(csrf).data.session };
  };
  const logout = async (payload, proof) => {
    assert.deepEqual(payload, { allSessions: false });
    assert.equal(proof, csrf);
    loggedOut = true;
    return { ok: true, requestId: "req", data: { revoked: true } };
  };
  const [refreshed, revoked] = await Promise.all([
    withAuthLock(() => refreshWithCurrentProof(user.id, bootstrap, refresh)),
    withAuthLock(() => logoutWithCurrentProof(user.id, { allSessions: false }, bootstrap, logout)),
  ]);
  assert.equal(refreshed.ok, true);
  assert.equal(revoked.ok, true);
  assert.equal(loggedOut, true);
  assert.equal((await bootstrap()).status, 401);
});

test("stale principal blocks refresh and logout before their protected mutation", async () => {
  const other = { ...ok("fresh"), data: { ...ok("fresh").data, user: { id: "other-user" } } };
  let mutations = 0;
  const refresh = await refreshWithCurrentProof("user-1", async () => other, async () => {
    mutations++;
    throw new Error("unexpected refresh");
  });
  const logout = await logoutWithCurrentProof("user-1", { allSessions: false }, async () => other, async () => {
    mutations++;
    throw new Error("unexpected logout");
  });
  assert.equal(refresh.status, 401);
  assert.equal(logout.status, 401);
  assert.equal(mutations, 0);
});

test("temporary bootstrap failure never performs a refresh or logout", async () => {
  let mutations = 0;
  const unavailable = async () => ({ ok: false, status: 503, message: "Unavailable" });
  const a = await refreshWithCurrentProof("user-1", unavailable, async () => { mutations++; });
  const b = await logoutWithCurrentProof("user-1", { allSessions: false }, unavailable, async () => { mutations++; });
  assert.equal(a.status, 503);
  assert.equal(b.status, 503);
  assert.equal(mutations, 0);
});

test("BroadcastChannel messages contain only a version-neutral invalidation signal", (t) => {
  const previous = Object.getOwnPropertyDescriptor(globalThis, "BroadcastChannel");
  const sent = [];
  const channels = [];
  class FakeBroadcastChannel {
    constructor(name) { this.name = name; channels.push(this); }
    postMessage(data) { sent.push(data); }
    close() { this.closed = true; }
  }
  Object.defineProperty(globalThis, "BroadcastChannel", { configurable: true, value: FakeBroadcastChannel });
  t.after(() => {
    if (previous) Object.defineProperty(globalThis, "BroadcastChannel", previous);
    else delete globalThis.BroadcastChannel;
  });
  let received = 0;
  const unsubscribe = subscribeAuthChanges(() => received++);
  announceAuthChange();
  assert.deepEqual(sent, [{ type: "session-changed" }]);
  assert.equal(channels[0].name, "openinvest-auth-state-v1");
  assert.equal(channels[1].name, "openinvest-auth-state-v1");
  assert.equal(channels[1].closed, true);
  channels[0].onmessage({ data: sent[0] });
  channels[0].onmessage({ data: { type: "other", accessToken: "secret" } });
  assert.equal(received, 1);
  assert.equal(isSessionChangeMessage(null), false);
  assert.equal(isSessionChangeMessage({ type: "session-changed" }), true);
  unsubscribe();
  assert.equal(channels[0].closed, true);
});

test("logout before refresh prevents refresh-token resurrection", async (t) => {
  installLocks(t);
  let active = true;
  let refreshCalls = 0;
  const bootstrap = async () => active ? ok("fresh-csrf") : ({ ok: false, status: 401, message: "Invalid session" });
  const [logoutResult, refreshResult] = await Promise.all([
    withAuthLock(() => logoutWithCurrentProof(user.id, { allSessions: false }, bootstrap, async () => {
      active = false;
      return { ok: true, requestId: "req", data: { revoked: true } };
    })),
    withAuthLock(() => refreshWithCurrentProof(user.id, bootstrap, async () => {
      refreshCalls++;
      return { ok: true, requestId: "req", data: ok("rotated").data.session };
    })),
  ]);
  assert.equal(logoutResult.ok, true);
  assert.equal(refreshResult.ok, false);
  assert.equal(refreshResult.status, 401);
  assert.equal(refreshCalls, 0);
});

test("fallback serializes operations within one tab when Web Locks are absent", async (t) => {
  const before = Object.getOwnPropertyDescriptor(globalThis, "navigator");
  Object.defineProperty(globalThis, "navigator", { configurable: true, value: { locks: undefined } });
  t.after(() => {
    if (before) Object.defineProperty(globalThis, "navigator", before);
    else delete globalThis.navigator;
  });
  const history = [];
  await Promise.all([
    withAuthLock(async () => { history.push("first-start"); await Promise.resolve(); history.push("first-end"); }),
    withAuthLock(async () => { history.push("second-start"); await Promise.resolve(); history.push("second-end"); }),
  ]);
  assert.deepEqual(history, ["first-start", "first-end", "second-start", "second-end"]);
});

test("AuthShell wires Web Locks, CSRF re-bootstrap and cross-tab invalidation", async () => {
  const { readFile } = await import("node:fs/promises");
  const content = await readFile(new URL("../src/features/auth/components/AuthShell.tsx", import.meta.url), "utf8");
  assert.match(content, /bootstrapRequestRef\.current \?\? withAuthLock\(bootstrapSession\)/);
  assert.match(content, /subscribeAuthChanges\(/);
  assert.match(content, /refreshWithCurrentProof\(/);
  assert.match(content, /logoutWithCurrentProof\(/);
  assert.match(content, /announceAuthChange\(/);
});

test("OpenAPI documents cookie-backed bootstrap and its non-rotating response", async () => {
  const { readFile } = await import("node:fs/promises");
  const path = await readFile(new URL("../../openapi/openapi.yaml", import.meta.url), "utf8");
  const responses = await readFile(new URL("../../openapi/components/responses.yaml", import.meta.url), "utf8");
  assert.match(path, /\/api\/v1\/auth\/bootstrap:/);
  assert.match(path, /operationId: bootstrapSession/);
  assert.match(path, /name: X-OpenInvest-Bootstrap/);
  assert.match(path, /responses\/Bootstrap/);
  assert.match(responses, /Bootstrap:\n[\s\S]*?No refresh cookie is issued or rotated/);
  assert.match(responses, /Cache-Control:/);
});
