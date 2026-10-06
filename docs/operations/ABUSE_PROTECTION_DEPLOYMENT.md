# Abuse Protection Deployment Ownership

OpenInvest deliberately separates three security layers.

1. **Per-process safety limits** stay inside the Go API and fail fast before expensive work. These
   include auth/IP and credential buckets, Argon2 admission, import execution/fresh admission and
   heavy-import capacity, expensive portfolio-read admission, the dividend fresh-command limiter,
   and the T-Invest provider's local concurrency/request budget.
2. **Deployment-global abuse limits** are owned by the shared edge/gateway in staging and production.
   The application does not claim that process-local maps/channels become global when replicas scale.
3. **Provider-global budget** for T-Invest is separately deployment-owned. Provider activation remains
   fail closed unless that shared provider-budget owner has been explicitly verified.

Staging/production startup requires:

```text
OPENINVEST_DEPLOYMENT_GLOBAL_ABUSE_CONTROL=verified-edge-v1
```

Set this value only after the serving edge/gateway has shared state across all API replicas and
enforces deployment-level admission for auth, imports, anonymous dividend calculations and expensive
portfolio reads. Local process limits remain defense in depth and must not be used as evidence that
the deployment-global limit exists.

The edge contract must preserve canonical client identity. It must not trust caller-supplied
forwarding headers directly, and its client-IP policy must be consistent with the application's
validated trusted-proxy boundary. At minimum the deployment control must be no weaker than these
repository process ceilings for the corresponding traffic classes:

```text
auth aggregate attempts             <= 2000/min deployment-wide
login/register emergency IP budget  <= 100/min per canonical client IP and route
login/register expensive auth work  <= 2 concurrently deployment-wide
refresh/logout IP budget            <= 20/min per canonical client IP and route
import execution                    <= 120/min deployment-wide
fresh import commands               <= 60/min deployment-wide
heavy import work                   <= 2 concurrently deployment-wide
anonymous fresh dividend commands   <= 20/min per canonical client IP and <= 1200/min deployment-wide
expensive portfolio reads           <= 8 concurrently deployment-wide

The repository process-local expensive-read admission allows up to 5 concurrent reads for one
subject because the current Portfolio detail UI has a verified automatic fan-out wave of five
(summary + current positions + cash flow + historical positions + returns). The sixth concurrent
same-subject expensive read fails fast locally. Repeated manual reload clicks are separate user
actions and remain subject to the bounded admission policy.
```

The Go layer additionally enforces a normalized-credential login/register bucket. The edge need not
parse passwords and must never log request bodies or credentials; repository tests prove the local
credential layer remains independent of account existence and that IP emergency protection still
bounds rotating identities.

T-Invest Corporate Actions stays disabled unless explicitly enabled. When enabled, startup also
requires:

```text
OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER=verified-shared-provider-budget-v1
```

That value may be set only when a shared control external to any single API process owns the
provider-wide request budget across replicas. The shared provider control must bound the aggregate
deployment to at most 60 provider requests/minute and 4 concurrent provider requests. The provider's
local 60 requests/minute and concurrency 4 controls remain safety limits, not deployment-global claims.

Development/local mode does not require these deployment ownership acknowledgements because it is
not a horizontally scaled production security boundary. Production fail-closed configuration tests
must remain green whenever these ownership gates change.
