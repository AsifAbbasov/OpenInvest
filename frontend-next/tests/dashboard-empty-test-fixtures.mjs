// Test-only replacements. Never imported by production code.
import React from "react";
export function useAuth() {
  return { accessToken: "fake-test-access-token", principalId: "00000000-0000-4000-8000-000000000001" };
}
export async function listPortfolios() {
  return { ok: true, requestId: "test", data: { items: [], pagination: { hasMore: false, nextCursor: null } } };
}
export function startPortfolioLoad(state, accessToken) {
  const generation = state.generation + 1;
  return { state: { generation, accessToken }, attempt: { generation, accessToken } };
}
export function shouldCommitPortfolioLoad() { return true; }
export function CreatePortfolioForm() {
  return React.createElement("div", { "data-testid": "empty-portfolio-create" }, "Create portfolio");
}
export default function Link({ href, children, ...rest }) {
  return React.createElement("a", { href, ...rest }, children);
}
