# Module #1 — JWT-02 and JWT-03 policy decision record

Current owner disposition: [approved closeout decisions](MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md). Earlier proposed/pending statements below record the pre-approval package and are superseded only as to these three decisions; historical evidence and residuals remain unchanged.

Status: OWNER_APPROVED_OPTION_A; JWT02_OWNER_APPROVAL=APPROVED.
Implementation frozen at PR266 e982d5a68e083a9e65b7f80b902f746b1b408bcb.
No verifier/parser change is part of this documentation PR.

## JWT-02 — future issued-at

Gatekeeper disposition supplied by owner: DEFENSE_IN_DEPTH_GAP / LOW; no external
HMAC forgery or account takeover demonstrated. The issuer normally uses current time;
the verifier checks exp but does not require/compare numeric iat. No approved skew
contract was found in ADR-006/OpenAPI. The original negative assertions remain RED:
[PR261](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38591817479/job/115948357047),
[PR265](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38636806790/job/116101936984).
PR266 changed only identity consistency. Its green CI is not a green future-iat suite.
Historical source fixtures preserve the full RED assertions rather than weaken them.

### Option A — approved for current closeout

Accept the existing lack of future-iat comparison as a documented residual. Keep
exp <= now rejection unchanged; introduce no iat restriction or compatibility fallback.
A correctly signed token with future iat can remain acceptable until exp; iat does
not postpone activation and is not a revocation oracle. This is a bounded statement
about this verifier, not universal issuer safety or acceptance of malformed JSON/types.
The verifier does not enforce a separate maximum remaining lifetime; the ordinary
issuer defaults to 15 minutes. Compromised signing authority and clock faults remain operational risks. Track the
historical rejection expectation as unapproved stronger policy, not a passing test.
Revisit on a new issuer, deployment time contract or concrete inconsistent consumer.

Exact proposed owner text:

> I approve JWT02_OPTION_A for the current Module #1 closeout. Future-iat comparison
> is not required by the current access-token contract. Existing expiration validation
> remains mandatory. I accept the documented signed-future-iat residual and preserve
> historical RED tests without calling them PASS. This does not authorize merge,
> deployment, or weakening signature/identity/expiry validation.

### Option B — alternative requiring separate authorization

Propose MAX_FUTURE_IAT_SKEW=60 seconds, relative to the verifier's UTC clock. Require
an explicitly present integer iat; reject iat > now+60s, iat > exp, or malformed/overflow
values; retain existing exp <= now rejection and identity/signature rules. An old iat
alone is not rejection if exp remains valid. Exactly +60s is allowed, +61s rejected.
Implement overflow-safe comparisons; numeric timestamp policy needs explicit owner
approval before code changes. No new issuer/audience/maximum-age rule is implied.

Compatibility: normal repository issuance remains valid, but externally minted tokens
with missing iat or excessive skew would be rejected. Require time synchronization,
review of all replicas/issuers and new exact-head CI. Narrow tests: normal issuance,
missing/null/string/fraction/overflow iat, +59/+60/+61 seconds, +24h, iat>exp, expired
exp and clock movement; login/refresh and financial isolation must remain unchanged.
No code or test implementation of Option B is authorized by this record.

Exact alternative owner text:

> I approve JWT02_OPTION_B with maximum future-iat skew 60 seconds and the timestamp
> semantics in this record, and authorize a separate minimal remediation with
> compatibility review and independent Gatekeeper acceptance. Option A is not selected.

## JWT-03 — duplicate claim names

Gatekeeper disposition supplied by owner: DOCUMENTED_ACCEPTED_BEHAVIOR / INFORMATIONAL,
no mandatory fix. Go encoding/json currently applies last-value decoding for repeated
recognized string claims. Effective decoded sub/uid must still satisfy JWT-01 after
PR266; decoded financial subject is still signature-covered. This does not prove
cross-parser equivalence or permit external signing. Historical observation:
[PR263 JWT job](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38623135638/job/116055467143).
No parser change, new task or universal production-safety claim is made. New evidence
of a contract violation would require a separately scoped decision.

## Other architecture decisions are distinct

Access-token authentication remains stateless; logout revokes refresh/session material
but does not make already-issued access JWTs immediately invalid. Default access TTL
is 15 minutes. Any denylist/session lookup or different lifetime needs separate approval.
Legacy account recovery/email change and ADR-010 acceptance remain separate gates.
