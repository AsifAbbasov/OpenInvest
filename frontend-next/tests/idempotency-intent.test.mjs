import assert from "node:assert/strict";
import test from "node:test";

const idempotency = await import("../src/common/api/idempotency.ts");

test("same write intent retains its key after an ambiguous retry", () => {
  const generated = ["intent-key-1", "intent-key-2"];
  let state = idempotency.emptyIdempotencyIntent;

  state = idempotency.idempotencyIntentFor(state, '{"name":"RUB"}', () => generated.shift());
  const retry = idempotency.idempotencyIntentFor(state, '{"name":"RUB"}', () => generated.shift());

  assert.equal(state.key, "intent-key-1");
  assert.equal(retry.key, "intent-key-1");
  assert.equal(generated.length, 1);
});

test("changed write intent receives a new idempotency key", () => {
  const generated = ["intent-key-1", "intent-key-2"];
  const first = idempotency.idempotencyIntentFor(
    idempotency.emptyIdempotencyIntent,
    '{"grossAmount":"100"}',
    () => generated.shift(),
  );
  const changed = idempotency.idempotencyIntentFor(first, '{"grossAmount":"200"}', () => generated.shift());

  assert.equal(first.key, "intent-key-1");
  assert.equal(changed.key, "intent-key-2");
});
