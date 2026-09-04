import assert from "node:assert/strict";
import test from "node:test";

import {
  corporateActionEffectiveDateLabel,
  corporateActionHeatmapLevel,
  corporateActionHeatmapMaximum,
  MAX_CORPORATE_ACTION_INSTRUMENTS,
  parseCorporateActionInstrumentInput,
} from "../src/features/corporate-actions/corporateActionsModel.ts";

test("instrument input parsing preserves canonical identity text while normalizing separators", () => {
  assert.deepEqual(parseCorporateActionInstrumentInput("SBER,  GAZP\nRU000A SBER"), ["SBER", "GAZP", "RU000A"]);
  assert.deepEqual(parseCorporateActionInstrumentInput("  "), []);
  assert.equal(MAX_CORPORATE_ACTION_INSTRUMENTS, 50);
});

test("effective-date label preserves Stage 3.61 dividend/coupon rules", () => {
  assert.equal(corporateActionEffectiveDateLabel({ kind: "DIVIDEND", recordDate: "2026-10-10" }, "2026-10-10"), "Record date");
  assert.equal(corporateActionEffectiveDateLabel({ kind: "DIVIDEND", recordDate: null }, "2026-10-20"), "Payment date");
  assert.equal(corporateActionEffectiveDateLabel({ kind: "COUPON", recordDate: "2026-10-01" }, "2026-10-20"), "Payment date");
});

test("heatmap levels are deterministic and based only on event density", () => {
  assert.equal(corporateActionHeatmapLevel(0, 4), 0);
  assert.equal(corporateActionHeatmapLevel(1, 4), 1);
  assert.equal(corporateActionHeatmapLevel(2, 4), 2);
  assert.equal(corporateActionHeatmapLevel(3, 4), 3);
  assert.equal(corporateActionHeatmapLevel(4, 4), 4);
  assert.equal(corporateActionHeatmapMaximum([{ totalCount: 1 }, { totalCount: 4 }, { totalCount: 2 }]), 4);
});
