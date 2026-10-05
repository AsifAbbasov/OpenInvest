import type { NextConfig } from "next";

const documentSecurityHeaders = [
  {
    key: "Content-Security-Policy",
    value: "frame-ancestors 'none'",
  },
  {
    key: "X-Frame-Options",
    value: "DENY",
  },
];

const nextConfig: NextConfig = {
  poweredByHeader: false,
  reactStrictMode: true,
  async headers() {
    return [
      {
        // Keep document-only anti-framing policy off immutable Next static assets.
        source: "/((?!_next/static|_next/image|favicon.ico).*)",
        headers: documentSecurityHeaders,
      },
    ];
  },
  turbopack: {
    root: process.cwd(),
  },
};

export default nextConfig;
