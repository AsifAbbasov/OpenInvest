import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const policy = await import("../src/common/presentation/unicode.ts");

test("Stage 3.39 counts Unicode code points rather than UTF-16 code units", () => {
  assert.equal("😀".repeat(100).length, 200);
  assert.equal(policy.unicodeCodePointCount("😀".repeat(100)), 100);
  assert.equal(policy.unicodeTextValidationError("😀".repeat(100), 100), null);
  assert.equal(policy.unicodeTextValidationError("😀".repeat(101), 100), "TOO_LONG");
  assert.equal(policy.unicodeTextValidationError("Ж".repeat(100), 100), null);
  assert.equal(policy.unicodeTextValidationError("Ж".repeat(101), 100), "TOO_LONG");

  // Import source-account-label boundary required by the Stage 3.39 plan.
  assert.equal(policy.unicodeTextValidationError("Ж".repeat(120), 120), null);
  assert.equal(policy.unicodeTextValidationError("Ж".repeat(121), 120), "TOO_LONG");
  assert.equal(policy.unicodeTextValidationError("😀".repeat(120), 120), null);
  assert.equal(policy.unicodeTextValidationError("😀".repeat(121), 120), "TOO_LONG");
});

test("Stage 3.39 rejects unpaired surrogates and accepts valid pairs", () => {
  assert.equal(policy.isWellFormedUnicode("\ud800"), false);
  assert.equal(policy.isWellFormedUnicode("\udc00"), false);
  assert.equal(policy.isWellFormedUnicode("😀"), true);
  assert.equal(policy.unicodeTextValidationError("\ud800", 100), "ILL_FORMED");
});

test("affected Web surfaces validate normalized values and do not rely on native maxlength", async () => {
  const files = {
    portfolio: await readFile(new URL("../src/features/portfolio/components/CreatePortfolioForm.tsx", import.meta.url), "utf8"),
    asset: await readFile(new URL("../src/features/assets/components/AssetDiscoverySlice.tsx", import.meta.url), "utf8"),
    transaction: await readFile(new URL("../src/features/portfolio/components/AddTransactionForm.tsx", import.meta.url), "utf8"),
    importPanel: await readFile(new URL("../src/features/portfolio/components/ImportUploadReviewPanel.tsx", import.meta.url), "utf8"),
  };

  assert.match(files.portfolio, /const normalizedName = name\.trim\(\)/);
  assert.match(files.portfolio, /unicodeTextValidationError\(normalizedName, 100\)/);
  assert.doesNotMatch(files.portfolio, /maxLength=\{100\}/);

  assert.match(files.asset, /const trimmedQuery = query\.trim\(\)/);
  assert.match(files.asset, /unicodeTextValidationError\(trimmedQuery, 100\)/);
  assert.doesNotMatch(files.asset, /maxLength=\{100\}/);

  assert.match(files.transaction, /const normalizedNote = note\.trim\(\)/);
  assert.match(files.transaction, /unicodeTextValidationError\(normalizedNote, 500\)/);
  assert.doesNotMatch(files.transaction, /maxLength=\{500\}/);

  assert.match(files.importPanel, /const normalizedSourceAccountLabel = sourceAccountLabel\.trim\(\)/);
  assert.match(files.importPanel, /unicodeTextValidationError\(normalizedSourceAccountLabel, 120\)/);
  assert.doesNotMatch(files.importPanel, /maxLength=\{120\}/);
  assert.match(files.importPanel, /file\.size > maxCsvPayloadBytes/);
});
