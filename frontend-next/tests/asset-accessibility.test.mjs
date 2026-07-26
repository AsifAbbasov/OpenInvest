import assert from "node:assert/strict";
import test from "node:test";

const accessibility = await import("../src/features/assets/assetAccessibility.ts");

test("asset live-region text announces async search outcomes", () => {
  assert.equal(accessibility.assetSearchStatusMessage("loading", 0), "Loading supported assets.");
  assert.equal(accessibility.assetSearchStatusMessage("loadingMore", 1), "Loading more supported assets.");
  assert.equal(accessibility.assetSearchStatusMessage("empty", 0), "No supported assets matched this search.");
  assert.equal(accessibility.assetSearchStatusMessage("error", 0), "");
  assert.equal(accessibility.assetSearchStatusMessage("ready", 1), "1 supported asset shown.");
  assert.equal(accessibility.assetSearchStatusMessage("ready", 2), "2 supported assets shown.");
});

test("asset detail live-region text announces observable detail outcomes", () => {
  assert.equal(accessibility.assetDetailStatusMessage("idle"), "");
  assert.equal(accessibility.assetDetailStatusMessage("loading", "SBER"), "Checking detail availability for SBER.");
  assert.equal(accessibility.assetDetailStatusMessage("available", "SBER"), "Detail is available for SBER.");
  assert.equal(accessibility.assetDetailStatusMessage("deferred", "SBER"), "Detail is unavailable for SBER.");
  assert.equal(accessibility.assetDetailStatusMessage("error", "SBER"), "");
});

test("asset keyboard contract closes detail only on Escape while detail is open", () => {
  assert.equal(accessibility.shouldCloseAssetDetailForKey("Escape", true), true);
  assert.equal(accessibility.shouldCloseAssetDetailForKey("Enter", true), false);
  assert.equal(accessibility.shouldCloseAssetDetailForKey("Escape", false), false);
});

test("asset focus restoration prefers origin result and falls back to search input", () => {
  assert.deepEqual(accessibility.assetDetailFocusRestoreTarget("SBER", ["SBER", "SU26238RMFS4"]), {
    kind: "result",
    ticker: "SBER",
  });
  assert.deepEqual(accessibility.assetDetailFocusRestoreTarget("SBER", ["SU26238RMFS4"]), {
    kind: "searchInput",
  });
  assert.deepEqual(accessibility.assetDetailFocusRestoreTarget(null, ["SBER"]), {
    kind: "searchInput",
  });
});

test("asset detail focus moves on entry or replacement without stealing focus on async completion", () => {
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion({ status: "idle" }, { status: "loading", ticker: "SBER" }),
    true,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion(
      { status: "loading", ticker: "SBER" },
      { status: "available", ticker: "SBER" },
    ),
    false,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion(
      { status: "loading", ticker: "SBER" },
      { status: "deferred", ticker: "SBER" },
    ),
    false,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion(
      { status: "available", ticker: "SBER" },
      { status: "loading", ticker: "SBER" },
    ),
    true,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion(
      { status: "deferred", ticker: "SBER" },
      { status: "loading", ticker: "SBER" },
    ),
    true,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion(
      { status: "available", ticker: "SBER" },
      { status: "loading", ticker: "GAZP" },
    ),
    true,
  );
  assert.equal(
    accessibility.shouldFocusAssetDetailRegion({ status: "available", ticker: "SBER" }, { status: "idle" }),
    false,
  );
});

test("asset detail generation rejects stale responses after cancellation", () => {
  const first = accessibility.nextAssetDetailAttempt(0, "SBER");
  assert.deepEqual(first, { generation: 1, ticker: "SBER" });
  assert.equal(accessibility.shouldAcceptAssetDetailResult(first.generation, first), true);

  const cancelledGeneration = accessibility.cancelAssetDetailGeneration(first.generation);
  assert.equal(accessibility.shouldAcceptAssetDetailResult(cancelledGeneration, first), false);

  const second = accessibility.nextAssetDetailAttempt(cancelledGeneration, "GAZP");
  assert.equal(accessibility.shouldAcceptAssetDetailResult(second.generation, second), true);
  assert.equal(accessibility.shouldAcceptAssetDetailResult(second.generation, first), false);
});
