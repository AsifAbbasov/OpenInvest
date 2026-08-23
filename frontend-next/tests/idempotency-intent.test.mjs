import assert from "node:assert/strict";
import test from "node:test";

const idempotency = await import("../src/common/api/idempotency.ts");

test("same write intent retains its key after an ambiguous retry", () => {
  const generated = ["intent-key-000001", "intent-key-000002"];
  let state = idempotency.emptyIdempotencyIntent;

  state = idempotency.idempotencyIntentFor(state, '{"name":"RUB"}', () => generated.shift());
  const retry = idempotency.idempotencyIntentFor(state, '{"name":"RUB"}', () => generated.shift());

  assert.equal(state.key, "intent-key-000001");
  assert.equal(retry.key, "intent-key-000001");
  assert.equal(generated.length, 1);
});

test("changed write intent receives a new idempotency key", () => {
  const generated = ["intent-key-000001", "intent-key-000002"];
  const first = idempotency.idempotencyIntentFor(
    idempotency.emptyIdempotencyIntent,
    '{"grossAmount":"100"}',
    () => generated.shift(),
  );
  const changed = idempotency.idempotencyIntentFor(first, '{"grossAmount":"200"}', () => generated.shift());

  assert.equal(first.key, "intent-key-000001");
  assert.equal(changed.key, "intent-key-000002");
});

test("browser retry intent survives reload without persisting principal or business scope", async () => {
  const storage = fakeStorage();
  const firstKey = "11111111-1111-4111-8111-111111111111";
  const principalId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
  const portfolioId = "portfolio-secret-123";
  const scope = idempotency.principalScopedIdempotencyScope(principalId, `transaction-append:${portfolioId}`);
  const payload = '{"grossAmount":{"amount":"2800.00000000","currency":"RUB"}}';
  const now = 1_800_000_000_000;

  const first = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    payload,
    scope,
    () => firstKey,
    storage,
    now,
  );
  const afterReload = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    payload,
    scope,
    () => "22222222-2222-4222-8222-222222222222",
    storage,
    now + 1_000,
  );

  assert.equal(first.key, firstKey);
  assert.equal(afterReload.key, firstKey);
  assert.equal(storage.entries().length, 1);
  const serializedStorage = JSON.stringify(storage.entries());
  assert.equal(serializedStorage.includes("2800.00000000"), false);
  assert.equal(serializedStorage.includes(portfolioId), false);
  assert.equal(serializedStorage.includes(principalId), false);
  assert.equal(serializedStorage.includes("grossAmount"), false);
});

test("changed intent in the same mounted interaction rotates and persists the key", async () => {
  const storage = fakeStorage();
  const scope = idempotency.principalScopedIdempotencyScope(
    "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "transaction-append:portfolio-a",
  );
  const first = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    '{"grossAmount":"100"}',
    scope,
    () => "11111111-1111-4111-8111-111111111111",
    storage,
    1_800_000_000_000,
  );
  const changed = await idempotency.idempotencyIntentForBrowser(
    first,
    '{"grossAmount":"200"}',
    scope,
    () => "22222222-2222-4222-8222-222222222222",
    storage,
    1_800_000_001_000,
  );
  const afterReload = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    '{"grossAmount":"200"}',
    scope,
    () => "33333333-3333-4333-8333-333333333333",
    storage,
    1_800_000_002_000,
  );

  assert.equal(changed.key, "22222222-2222-4222-8222-222222222222");
  assert.equal(afterReload.key, changed.key);
});

test("expired browser retry intent is discarded and replaced", async () => {
  const storage = fakeStorage();
  const scope = idempotency.principalScopedIdempotencyScope(
    "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "portfolio-create",
  );
  const first = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    '{"name":"A"}',
    scope,
    () => "11111111-1111-4111-8111-111111111111",
    storage,
    1_800_000_000_000,
  );
  const afterExpiry = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    '{"name":"A"}',
    scope,
    () => "22222222-2222-4222-8222-222222222222",
    storage,
    1_800_086_400_001,
  );

  assert.notEqual(first.key, afterExpiry.key);
  assert.equal(afterExpiry.key, "22222222-2222-4222-8222-222222222222");
});

test("confirmed write can clear the browser retry intent", async () => {
  const storage = fakeStorage();
  const scope = idempotency.principalScopedIdempotencyScope(
    "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "portfolio-create",
  );
  await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    '{"name":"A"}',
    scope,
    () => "11111111-1111-4111-8111-111111111111",
    storage,
    1_800_000_000_000,
  );
  assert.equal(storage.entries().length, 1);

  await idempotency.clearBrowserIdempotencyIntent(scope, storage);
  assert.equal(storage.entries().length, 0);
});

test("retry journal is isolated across authenticated principals in the same browser tab", async () => {
  const storage = fakeStorage();
  const userA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
  const userB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
  const scopeA = idempotency.principalScopedIdempotencyScope(userA, "portfolio-create");
  const scopeB = idempotency.principalScopedIdempotencyScope(userB, "portfolio-create");
  const payloadA = '{"name":"User A unresolved portfolio"}';
  const payloadB = '{"name":"User B portfolio"}';
  const keyA = "aaaaaaaa-1111-4111-8111-111111111111";
  const keyB = "bbbbbbbb-2222-4222-8222-222222222222";
  const now = 1_800_000_000_000;

  const firstA = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    payloadA,
    scopeA,
    () => keyA,
    storage,
    now,
  );

  const firstB = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    payloadB,
    scopeB,
    () => keyB,
    storage,
    now + 1_000,
  );

  assert.equal(firstA.key, keyA);
  assert.equal(firstB.key, keyB);
  assert.equal(storage.entries().length, 2);

  // User B's confirmed success clears only B's technical retry slot.
  await idempotency.clearBrowserIdempotencyIntent(scopeB, storage);
  assert.equal(storage.entries().length, 1);

  // User A later signs back in in the same tab and recovers the unresolved original key.
  const restoredA = await idempotency.idempotencyIntentForBrowser(
    idempotency.emptyIdempotencyIntent,
    payloadA,
    scopeA,
    () => "cccccccc-3333-4333-8333-333333333333",
    storage,
    now + 2_000,
  );
  assert.equal(restoredA.key, keyA);

  const serializedStorage = JSON.stringify(storage.entries());
  assert.equal(serializedStorage.includes(userA), false);
  assert.equal(serializedStorage.includes(userB), false);
  assert.equal(serializedStorage.includes("User A unresolved portfolio"), false);
  assert.equal(serializedStorage.includes("User B portfolio"), false);
});

test("principal and operation scopes are both required", () => {
  assert.throws(() => idempotency.principalScopedIdempotencyScope("", "portfolio-create"));
  assert.throws(() => idempotency.principalScopedIdempotencyScope("user-a", ""));
});

function fakeStorage() {
  const values = new Map();
  return {
    getItem(key) {
      return values.has(key) ? values.get(key) : null;
    },
    setItem(key, value) {
      values.set(key, value);
    },
    removeItem(key) {
      values.delete(key);
    },
    entries() {
      return [...values.entries()];
    },
  };
}
