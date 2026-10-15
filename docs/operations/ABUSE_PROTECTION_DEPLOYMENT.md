# Abuse Protection Deployment Ownership

## Ownership and deployment prerequisite

Local Go limits fail fast for auth/IP/credentials, Argon2, import execution/fresh/heavy work,
expensive portfolio reads, fresh dividends, Corporate Actions caller budgets/coalescing and provider work.
They do not become deployment-global when replicas scale. Staging/production requires a verified
shared edge/gateway for auth, imports, anonymous dividend calculations and expensive portfolio reads:

```text
OPENINVEST_DEPLOYMENT_GLOBAL_ABUSE_CONTROL=verified-edge-v1
```

Set this acknowledgment only after verifying shared admission across all API replicas.
The edge must preserve canonical identity under the application's trusted-proxy policy; never trust
caller forwarding headers directly. Minimum deployment ceilings:

```text
auth aggregate attempts             <= 2000/min deployment-wide
login/register emergency IP budget  <= 100/min per canonical client IP and route
login/register expensive auth work  <= 2 concurrently deployment-wide
refresh/logout IP budget            <= 20/min per canonical client IP and route
import execution                    <= 120/min deployment-wide
fresh import commands               <= 60/min deployment-wide
heavy import work                   <= 2 concurrently deployment-wide
anonymous fresh dividend commands   <= 20/min per canonical client IP and <= 1200/min deployment-wide
corporate-actions projection         <= 12/min per canonical client IP and <= 48/min deployment-wide
expensive portfolio reads           <= 8 concurrently deployment-wide
```

Local expensive-read admission permits five concurrent reads per subject for the verified UI wave
(summary, current positions, cash flow, historical positions, returns); the sixth fails fast.
Manual reloads remain separate bounded user actions.
Normalized-credential login/register buckets are independent of account existence and retain IP
emergency limits for rotating identities. The edge need not parse passwords and must not log bodies/credentials.

## Shared Corporate Actions and provider budgets

Corporate Actions additionally uses atomic, TTL-bounded Redis admission: ≤12/min per canonical client
and ≤48/min across replicas. Redis outage fails closed, never reverting to process-local counters.
The edge acknowledgment neither enforces this Redis boundary nor proves the edge is deployed.
T-Invest remains disabled unless explicitly enabled. Outside development/local mode, activation
fails closed without reachable shared Redis and this separate operator acknowledgment:

```text
OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER=verified-shared-provider-budget-v1
OPENINVEST_SHARED_BUDGET_REDIS_URL=<operator-supplied-shared-security-budget-endpoint>
```

The acknowledgment is not enforcement. Redis owns ≤60 provider requests in any rolling 60 seconds
across replicas and conservative `X-RateLimit-Remaining` / `X-RateLimit-Reset` observations.
Endpoint/provider journals use atomic Redis-server-time sorted sets with bounded TTLs.
A legacy string counter becomes a full-window reservation, potentially denying one transition window;
no window resets early. Operators must verify clock integrity and shared namespace/topology.

## Local operation lifetime and residuals

Four local provider operations maximum remains process-local resource defense. Caller cancellation
returns promptly without cancelling the independently owned HTTP operation. Its separate five-second
deadline cancels the transport; the slot remains owned until operation/response-body closure completes,
not until an elapsed timer. Arbitrary remote server computation may survive connection closure;
external-provider server-side termination remains NOT_VERIFIED.
Development/local mode does not require deployment acknowledgments; it is not a horizontally scaled
production boundary. Keep production fail-closed configuration tests green when ownership gates change.
