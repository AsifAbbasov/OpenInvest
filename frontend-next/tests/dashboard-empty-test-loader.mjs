import { pathToFileURL } from "node:url";
const fixture = pathToFileURL(new URL("./dashboard-empty-test-fixtures.mjs", import.meta.url).pathname).href;
const targets = [
  /(?:^|\/)common\/api\/openinvest(?:\.ts)?$/,
  /(?:^|\/)features\/auth\/components\/AuthShell(?:\.tsx)?$/,
  /(?:^|\/)features\/portfolio\/components\/CreatePortfolioForm(?:\.tsx)?$/,
  /(?:^|\/)features\/portfolio\/loadGuard(?:\.ts)?$/,
];
export async function resolve(specifier, context, nextResolve) {
  if (specifier === "next/link" || targets.some(pattern => pattern.test(specifier))) {
    return { url: fixture, shortCircuit: true };
  }
  return nextResolve(specifier, context);
}
