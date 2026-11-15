import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const source = readFileSync(new URL("../src/features/dividends/components/DividendCalculator.tsx", import.meta.url), "utf8");

test("dividend calculator starts with empty user-entered financial values", () => {
  assert.match(source, /const \[ticker, setTicker\] = useState\(""\)/);
  assert.match(source, /const \[quantity, setQuantity\] = useState\(""\)/);
  assert.match(source, /const \[dividendPerUnit, setDividendPerUnit\] = useState\(""\)/);
  assert.match(source, /const \[positionCost, setPositionCost\] = useState\(""\)/);
  assert.doesNotMatch(source, /useState\("(?:SBER|34\.84000000|280000\.00000000|1000\.00000000)"\)/);
});
