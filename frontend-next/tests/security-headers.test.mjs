import assert from "node:assert/strict";
import test from "node:test";

const { default: nextConfig } = await import("../next.config.ts");

test("browser security headers are globally configured and fail closed", async () => {
  assert.equal(typeof nextConfig.headers, "function", "Next config must provide global security headers");

  const rules = await nextConfig.headers();
  assert.ok(rules.length >= 1, "at least one global header rule is required");

  const applicableRules = rules.filter((rule) =>
    typeof rule.source === "string" && rule.source.includes("(.*)")
  );
  assert.ok(applicableRules.length >= 1, "security headers must apply to application documents");

  const merged = new Map();
  for (const rule of applicableRules) {
    for (const { key, value } of rule.headers ?? []) {
      merged.set(key.toLowerCase(), value);
    }
  }

  const csp = merged.get("content-security-policy");
  assert.ok(csp, "Content-Security-Policy is required");
  assert.ok(csp.includes("default-src 'self'"), "CSP must default to self");
  assert.ok(csp.includes("object-src 'none'"), "CSP must disable object embedding");
  assert.ok(csp.includes("frame-ancestors 'none'"), "CSP must deny framing");
  assert.ok(csp.includes("base-uri 'self'"), "CSP must constrain base URI");

  assert.equal(
    merged.get("strict-transport-security"),
    "max-age=63072000; includeSubDomains; preload",
    "HSTS must be long-lived and preload-compatible",
  );
  assert.equal(merged.get("x-content-type-options"), "nosniff");
  assert.equal(merged.get("x-frame-options"), "DENY");
  assert.equal(merged.get("referrer-policy"), "strict-origin-when-cross-origin");

  const permissions = merged.get("permissions-policy");
  assert.ok(permissions, "Permissions-Policy is required");
  for (const directive of ["camera=()", "microphone=()", "geolocation=()"]) {
    assert.ok(permissions.includes(directive), `Permissions-Policy missing ${directive}`);
  }
});
