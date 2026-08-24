import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const policy = await import("../src/features/auth/passwordPolicy.ts");

test("registration counts Unicode code points rather than UTF-16 code units", () => {
  assert.notEqual(policy.passwordValidationMessage("😀".repeat(6), "register"), null);
  assert.equal(policy.passwordValidationMessage("😀".repeat(12), "register"), null);
  assert.equal(policy.passwordValidationMessage("😀".repeat(256), "register"), null);
  assert.notEqual(policy.passwordValidationMessage("😀".repeat(257), "register"), null);
});

test("login preserves historical multibyte and whitespace-only exact secrets", () => {
  assert.equal(policy.passwordValidationMessage("абвгде", "login"), null);
  assert.notEqual(policy.passwordValidationMessage("абвгде", "register"), null);
  assert.equal(policy.passwordValidationMessage(" ".repeat(12), "login"), null);
  assert.notEqual(policy.passwordValidationMessage("", "login"), null);
});

test("well-formed Unicode rejects lone surrogates and accepts valid pairs", () => {
  assert.equal(policy.isWellFormedPasswordUnicode("\ud800"), false);
  assert.equal(policy.isWellFormedPasswordUnicode("\udc00"), false);
  assert.equal(policy.isWellFormedPasswordUnicode("😀"), true);
  assert.equal(policy.passwordCodePointCount("😀"), 1);
});

test("AuthForm delegates password admission to the explicit policy", async () => {
  const source = await readFile(new URL("../src/features/auth/components/AuthForm.tsx", import.meta.url), "utf8");
  assert.match(source, /passwordValidationMessage\(password, mode\)/);
  assert.doesNotMatch(source, /minLength=\{12\}/);
  assert.doesNotMatch(source, /password\.trim\(/);
});
