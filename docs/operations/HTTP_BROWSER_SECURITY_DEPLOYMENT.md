# HTTP and Browser Security Deployment

OpenInvest owns application-level anti-framing and sensitive API cache policy. The Next.js
application sends `Content-Security-Policy: frame-ancestors 'none'` and `X-Frame-Options: DENY`
for document responses. The API sends `Cache-Control: no-store` for auth/session and portfolio
responses; public reference-data routes keep handler-owned cache policy.

TLS termination and HSTS are not owned by the repository because no production edge provider is
selected here. The verified HTTPS edge owner must configure them and must not strip application
security headers. Until that boundary is selected, HSTS remains intentionally deferred.

The API should normally be private behind that edge. `OPENINVEST_API_LISTEN_ADDRESS` is optional
only for `OPENINVEST_ENV=development` or `local`, where it defaults to `127.0.0.1:8080`.
Staging and production fail closed unless an operator supplies an explicit valid host:port.
Wildcard `0.0.0.0` or `[::]` binding is permitted only through that explicit setting when a
container or platform requires it; network firewall and TLS ownership remain operator duties.

`OPENINVEST_TRUST_PROXY=true` is independent from listener selection. It requires
`OPENINVEST_TRUSTED_PROXY_CIDRS` with only verified proxy ranges; direct mode remains the default
and ignores forwarding headers. Do not use trust-all CIDRs or infer proxy trust from private,
loopback, or link-local addresses.
