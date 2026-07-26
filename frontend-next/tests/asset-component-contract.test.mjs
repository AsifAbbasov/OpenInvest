import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const component = await readFile(
  new URL("../src/features/assets/components/AssetDiscoverySlice.tsx", import.meta.url),
  "utf8",
);

test("asset discovery component wires observable accessibility contracts", () => {
  assert.match(component, /aria-live="polite"/);
  assert.match(component, /role="alert"/);
  assert.match(component, /tabIndex=\{-1\}/);
  assert.match(component, /onKeyDown=\{handleEscape\}/);
  assert.match(component, /shouldFocusAssetDetailRegion\(previousDetailRef\.current, detail\)/);
  assert.match(component, /assetSearchStatusMessage/);
  assert.match(component, /assetDetailStatusMessage/);
});

test("asset discovery component wires detail invalidation helpers", () => {
  assert.match(component, /cancelDetailState\(\)/);
  assert.match(component, /cancelAssetDetailGeneration/);
  assert.match(component, /nextAssetDetailAttempt/);
  assert.match(component, /shouldAcceptAssetDetailResult/);
});

test("asset discovery component renders distinct successful and deferred detail states", () => {
  assert.match(component, /detail\.status === "available"/);
  assert.match(component, /Asset detail is available from the Go API/);
  assert.match(component, /detail\.status === "deferred"/);
  assert.match(component, /Asset detail is unavailable for this selection/);
  assert.match(component, /detail\.status === "deferred" \? "Deferred asset detail" : "Asset detail"/);
});
