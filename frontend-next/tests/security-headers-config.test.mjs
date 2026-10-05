import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const nextConfig = await readFile(
  new URL("../next.config.ts", import.meta.url),
  "utf8",
);

test("document responses have a centralized anti-framing policy", () => {
  assert.match(nextConfig, /Content-Security-Policy/);
  assert.match(nextConfig, /frame-ancestors 'none'/);
  assert.match(nextConfig, /X-Frame-Options/);
  assert.match(nextConfig, /value: "DENY"/);
});

test("anti-framing policy excludes immutable Next static assets", () => {
  assert.match(nextConfig, /_next\/static/);
  assert.match(nextConfig, /_next\/image/);
});
