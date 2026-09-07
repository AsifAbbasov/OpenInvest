import assert from "node:assert/strict";
import test from "node:test";

const { formatQuantityForDisplay, formatRatioAsPercent } = await import("../src/common/presentation/format.ts");

test("Stage 3.72 quantity formatting preserves fractional open positions without financial rounding", () => {
  assert.equal(formatQuantityForDisplay("150.00000000"), "150");
  assert.equal(formatQuantityForDisplay("0.00000001"), "0.00000001");
  assert.equal(formatQuantityForDisplay("12345.12000000"), "12 345.12");
});

test("Stage 3.72 ratio formatting is presentation-only and deterministic", () => {
  assert.equal(formatRatioAsPercent("0.34210000"), "34.2%");
  assert.equal(formatRatioAsPercent("0.33333333"), "33.3%");
  assert.equal(formatRatioAsPercent("0.34250000"), "34.2%");
  assert.equal(formatRatioAsPercent("0.34350000"), "34.4%");
  assert.equal(formatRatioAsPercent("1.00000000"), "100.0%");
  assert.equal(formatRatioAsPercent("0.00000000"), "0.0%");
  assert.equal(formatRatioAsPercent(null), "Undefined");
});
