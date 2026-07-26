import assert from "node:assert/strict";
import test from "node:test";

const state = await import("../src/features/assets/assetSearchState.ts");

const sber = {
  ticker: "SBER",
  name: "Sberbank ordinary shares",
  assetType: "STOCK",
  currency: "RUB",
  lotSize: "10.00000000",
  lastPrice: null,
};

const bond = {
  ticker: "SU26238RMFS4",
  name: "OFZ 26238",
  assetType: "BOND",
  currency: "RUB",
  lotSize: "1.00000000",
  lastPrice: null,
};

test("query or asset-type change resets cursor results selection and pending generation", () => {
  let current = state.initialAssetSearchState();
  let started = state.beginAssetSearch(current, { query: "S", assetType: "" }, null);
  current = state.acceptAssetSearchResult(started.state, started.attempt, [sber], {
    nextCursor: "cursor-1",
    hasMore: true,
    limit: 20,
  });
  current = state.selectAsset(current, "SBER");

  const reset = state.resetAssetSearchForKeyChange(current, { query: "S", assetType: "BOND" });

  assert.equal(reset.generation, current.generation + 1);
  assert.equal(reset.query, "S");
  assert.equal(reset.assetType, "BOND");
  assert.deepEqual(reset.items, []);
  assert.equal(reset.pagination.nextCursor, null);
  assert.equal(reset.selectedTicker, null);

  const stale = state.acceptAssetSearchResult(reset, started.attempt, [sber], {
    nextCursor: null,
    hasMore: false,
    limit: 20,
  });
  assert.equal(stale, reset);
  assert.deepEqual(stale.items, []);
});

test("pagination appends only for the active accepted cursor chain", () => {
  let current = state.initialAssetSearchState();
  let first = state.beginAssetSearch(current, { query: "S", assetType: "" }, null);
  current = state.acceptAssetSearchResult(first.state, first.attempt, [sber], {
    nextCursor: "cursor-1",
    hasMore: true,
    limit: 1,
  });

  const wrongCursor = { ...first.attempt, generation: current.generation, cursor: "cursor-other" };
  const rejected = state.acceptAssetSearchResult(current, wrongCursor, [bond], {
    nextCursor: null,
    hasMore: false,
    limit: 1,
  });
  assert.equal(rejected, current);
  assert.deepEqual(rejected.items.map((item) => item.ticker), ["SBER"]);

  const next = state.beginAssetSearch(current, { query: "S", assetType: "" }, "cursor-1");
  const accepted = state.acceptAssetSearchResult(next.state, next.attempt, [bond], {
    nextCursor: null,
    hasMore: false,
    limit: 1,
  });

  assert.deepEqual(accepted.items.map((item) => item.ticker), ["SBER", "SU26238RMFS4"]);
  assert.equal(accepted.pagination.nextCursor, null);
});
