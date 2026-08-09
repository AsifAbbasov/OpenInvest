import assert from "node:assert/strict";
import test from "node:test";

const guard = await import("../src/features/portfolio/importOperationGuard.ts");

test("starting a new review invalidates an older append response", () => {
	let state = { scope: "portfolio-a\u0000token-a", reviewGeneration: 0, appendGeneration: 0 };
	const firstReview = guard.startImportReview(state, "portfolio-a\u0000token-a");
	state = firstReview.state;
	const append = guard.startImportAppend(state, "portfolio-a\u0000token-a");
	state = append.state;

	const replacementReview = guard.startImportReview(state, "portfolio-a\u0000token-a");
  state = replacementReview.state;

  assert.equal(guard.shouldCommitImportAppend(state, append.attempt), false);
  assert.equal(guard.shouldCommitImportReview(state, replacementReview.attempt), true);
});

test("file or label invalidation supersedes stale review and append operations", () => {
	let state = { scope: "portfolio-a\u0000token-a", reviewGeneration: 3, appendGeneration: 7 };
	const review = guard.startImportReview(state, "portfolio-a\u0000token-a");
	state = review.state;
	const append = guard.startImportAppend(state, "portfolio-a\u0000token-a");
	state = append.state;

	const invalidation = guard.startImportReview(state, "portfolio-a\u0000token-a");
  state = invalidation.state;

  assert.equal(guard.shouldCommitImportReview(state, review.attempt), false);
	assert.equal(guard.shouldCommitImportAppend(state, append.attempt), false);
});

test("portfolio or token change invalidates an in-flight import operation", () => {
	let state = { scope: "portfolio-a\u0000token-a", reviewGeneration: 0, appendGeneration: 0 };
	const review = guard.startImportReview(state, "portfolio-a\u0000token-a");
	state = review.state;
	const append = guard.startImportAppend(state, "portfolio-a\u0000token-a");
	state = append.state;

	const replacement = guard.startImportReview(state, "portfolio-b\u0000token-b");
	state = replacement.state;

	assert.equal(guard.shouldCommitImportReview(state, review.attempt), false);
	assert.equal(guard.shouldCommitImportAppend(state, append.attempt), false);
  assert.equal(guard.shouldCommitImportReview(state, replacement.attempt), true);
});

test("scope synchronization rejects an old response before passive effects run", () => {
  let state = { scope: "portfolio-a\u0000token-a", reviewGeneration: 4, appendGeneration: 7 };
  const review = guard.startImportReview(state, state.scope);
  state = review.state;
  const append = guard.startImportAppend(state, state.scope);
  state = append.state;

  state = guard.synchronizeImportScope(state, "portfolio-b\u0000token-b");

  assert.equal(guard.shouldCommitImportReview(state, review.attempt), false);
  assert.equal(guard.shouldCommitImportAppend(state, append.attempt), false);
  assert.equal(state.scope, "portfolio-b\u0000token-b");
});
