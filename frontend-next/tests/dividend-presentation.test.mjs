import assert from "node:assert/strict";
import test from "node:test";

import { formatDividendMoney, formatGrossYield } from "../src/features/dividends/presentation.ts";

test("gross yield formatting preserves exact backend ratio digits", () => {
  assert.equal(formatGrossYield("0.12442857"), "12.442857%");
  assert.equal(formatGrossYield("1.00000000"), "100%");
  assert.equal(formatGrossYield("0.00000001"), "0.000001%");
  assert.equal(formatGrossYield(null), "—");
});

test("money display rounds exact decimal strings to two places without binary floating point", () => {
  assert.equal(formatDividendMoney("34840.00000000"), "34840 ₽");
  assert.equal(formatDividendMoney("34.84000000"), "34.84 ₽");
  assert.equal(formatDividendMoney("0.00000001"), "0 ₽");
  assert.equal(formatDividendMoney("1.00500000"), "1 ₽");
  assert.equal(formatDividendMoney("1.01500000"), "1.02 ₽");
  assert.equal(formatDividendMoney("1.00500001"), "1.01 ₽");
  assert.equal(formatDividendMoney("999.99900000"), "1000 ₽");
});
