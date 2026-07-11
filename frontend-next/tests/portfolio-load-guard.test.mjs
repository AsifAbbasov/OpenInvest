import assert from "node:assert/strict";
import test from "node:test";

const guard = await import("../src/features/portfolio/loadGuard.ts");

test("portfolio load guard rejects old-token response after refreshed-token load starts", () => {
  let state = { generation: 0, accessToken: "old-access-token" };
  const oldLoad = guard.startPortfolioLoad(state, state.accessToken);
  state = oldLoad.state;

  state = { ...state, accessToken: "new-access-token" };
  const refreshedLoad = guard.startPortfolioLoad(state, state.accessToken);
  state = refreshedLoad.state;

  assert.equal(guard.shouldCommitPortfolioLoad(state, oldLoad.attempt), false);
  assert.equal(guard.shouldCommitPortfolioLoad(state, refreshedLoad.attempt), true);
});

test("portfolio detail token rotation starts a replacement load before old response can commit", () => {
  let state = { generation: 0, accessToken: "old-access-token" };
  const outstandingDetailLoad = guard.startPortfolioLoad(state, state.accessToken);
  state = outstandingDetailLoad.state;

  state = { ...state, accessToken: "new-access-token" };
  const replacementDetailLoad = guard.startPortfolioLoad(state, state.accessToken);
  state = replacementDetailLoad.state;

  assert.equal(outstandingDetailLoad.attempt.accessToken, "old-access-token");
  assert.equal(replacementDetailLoad.attempt.accessToken, "new-access-token");
  assert.equal(guard.shouldCommitPortfolioLoad(state, outstandingDetailLoad.attempt), false);
  assert.equal(guard.shouldCommitPortfolioLoad(state, replacementDetailLoad.attempt), true);
});

test("portfolio load callbacks read the latest token instead of a stale closure token", () => {
  let state = { generation: 4, accessToken: "old-access-token" };
  state = { ...state, accessToken: "new-access-token" };

  const callbackLoad = guard.startPortfolioLoad(state, state.accessToken);

  assert.equal(callbackLoad.attempt.accessToken, "new-access-token");
  assert.equal(callbackLoad.attempt.generation, 5);
  assert.equal(guard.shouldCommitPortfolioLoad(callbackLoad.state, callbackLoad.attempt), true);
});
