# ACCESS_JWT_IDENTITY_POLICY_V1

Current owner disposition: [approved closeout decisions](../audit/MODULE1_OWNER_CLOSEOUT_APPROVALS_V1.md). Earlier proposed/pending statements below record the pre-approval package and are superseded only as to these three decisions; historical evidence and residuals remain unchanged.

| Field | Value |
| --- | --- |
| Policy status | Approved by project owner on 2026-10-15 |
| Approval evidence | Explicit owner message headed OPENINVEST — FORMAL JWT CONTRACT APPROVAL in the implementation conversation |
| Implementation | PR #266, e982d5a68e083a9e65b7f80b902f746b1b408bcb |
| Security review | ACCEPT_JWT01_REMEDIATION, as supplied by the owner in the architecture-closeout directive |
| Broader ADR status | ADR-006 Accepted; ADR-010 addendum owner-approved; original pinned ADR unchanged |

## Approved rule

1. sub must be present and non-empty.
2. uid must be present and non-empty.
3. investmentSubjectId must be present and non-empty.
4. sub MUST equal uid exactly, without normalization or substitution.
5. Missing, empty or inconsistent mandatory identity claims MUST reject the token.
6. Financial authorization continues using the validated signed investmentSubjectId.
7. Existing HS256 signature, expected algorithm/type and expiration checks remain.
8. No legacy compatibility fallback is permitted.

The rule does not add an iat restriction, change duplicate JSON parsing, introduce a
new issuer/key/library, or require a database lookup for each access token. It does
not protect against compromise of the signing key. The only reviewed issuer already
sets sub=uid=user.ID; clients treat the access token as opaque. No alternate authorized
issuer or different historical issuer format was found in repository scope.
A deployment's untracked external issuer is not certified by this repository review.

## Trace and error compatibility

PostgreSQL user/link -> StoredUser -> Service.result -> signAccessToken ->
verifyAccessToken -> AuthenticateAccessToken -> API.subjectID -> subject-scoped SQL.
The identity consistency check is not a replacement for financial subject scoping.
Rejected access tokens retain ErrInvalidSession and the existing HTTP 401
UNAUTHORIZED / Invalid or expired session contract.

Sources: [tokens.go](../../backend-go/internal/auth/tokens.go),
[service](../../backend-go/internal/auth/service.go),
[HTTP handlers](../../backend-go/internal/httpapi/auth_handlers.go),
[financial store](../../backend-go/internal/postgres/store.go).

## Immutable evidence

- [PR #266](https://github.com/AsifAbbasov/OpenInvest/pull/266), tree
  23cb0831d37c2a46dcdeaa96e21d065015d53bca; exact parent PR251
  0ec890e469efa87bb552f02b7acd66fc2fee0abc.
- Original [PR261 RED](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38591817479/job/115948357047)
  and [PR265 RED](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38636806790/job/116101937341).
- Full source blobs dfe275965e8702b1d2abec96ff5579373875a57b and
  21eac313f4b3d05cb55472ff627c83ecc66408e5 remain byte-identical in
  [testdata](../../backend-go/internal/auth/testdata/jwt01/).
- Both original JWT-01 assertions, positive/negative regression controls and real
  PostgreSQL HTTP login/refresh/bidirectional financial isolation passed in
  [job 116279366055](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38691771383/job/116279366055).
  [Exact-head ordinary CI](https://github.com/AsifAbbasov/OpenInvest/actions/runs/38691771383)
  succeeded 10/10, including Go race and vet. No timestamp/parser acceptance is inferred.

## Boundaries still requiring decisions

See [JWT02/JWT03 decision record](../audit/MODULE1_JWT02_JWT03_POLICY_DECISION.md).
The prospective [ADR010 approval candidate](../ADR/ADR-010-email-ownership-approval-candidate.md)
preserves the hash-pinned historical ADR file. Approval of this identity policy does not accept ADR-010, authorize merge/deploy,
close Module #1, or approve the PR265 campaign's procedural provenance.
