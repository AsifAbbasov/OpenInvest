# Stage 3.35 — P3-01 Password Character Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only |
| Date | 2026-08-24 |
| Canonical base | `develop` at `ae5a152114cc163867a363953f8a3202396b1f6c` |
| Finding | P3-01 |
| Scope | Password length and exact-secret admission semantics only |
| Runtime implementation authorized here | No |

## 1. Finding / symptom

P3-01: password length semantics use bytes rather than the intended user-facing character semantics.

`backend-go/internal/auth/service.go` says passwords must be 12..256 characters but uses Go `len(string)`, which counts UTF-8 bytes. `openapi/components/schemas.yaml` expresses the same rule with `minLength`/`maxLength`, while the Web form relies on browser `minLength`. Non-ASCII input can therefore be counted differently across layers.

A related exact-secret inconsistency exists in login admission: registration does not trim a password, but login currently rejects `strings.TrimSpace(request.Password) == ""`. A password consisting entirely of whitespace can therefore be accepted at registration and later rejected before exact hash verification.

## 2. Root cause

No canonical password-length unit was defined. Go byte length, OpenAPI character length, and browser/JavaScript string behavior were treated as equivalent when they are not.

The login path also uses whitespace normalization as an emptiness test even though password identity is otherwise exact-byte based.

Argon2 hashing itself is not the defect; it should continue receiving the exact submitted password bytes.

## 3. Failure scenarios

- A password with fewer than 12 Cyrillic characters can still occupy at least 12 UTF-8 bytes and pass the current backend minimum.
- A 256-code-point password can exceed 256 UTF-8 bytes and be rejected although the contract permits 256 characters.
- Astral Unicode can produce Web/backend disagreement.
- Enforcing the corrected registration minimum during login could lock out credentials historically accepted under the old byte-based rule.
- A 12-space password can be accepted at registration but cannot currently authenticate because login treats its trimmed value as empty.
- If login gains a code-point upper bound without also validating UTF-8, internal/non-HTTP callers could still bypass the contract's Unicode-string model with malformed Go strings.

## 4. Impact

Contract/runtime inconsistency, multilingual UX defects, misleading password policy enforcement, and possible credential lockout during remediation or for exact whitespace-only secrets already accepted by registration. No plaintext-password exposure or Argon2 cryptographic defect is claimed.

## 5. Severity rationale

P3 is appropriate: the defect is real and security-adjacent but is not known to create privilege escalation, credential disclosure, financial corruption, or systemic availability failure.

## 6. Existing guarantees violated

- OpenAPI and runtime validation must agree.
- User-facing validation must be deterministic across supported clients and backend.
- Maintenance must not invalidate previously accepted credentials.
- Password hashing and authentication must preserve the exact secret without hidden trimming or normalization.

## 7. Considered solutions

A. Redefine the public policy as UTF-8 bytes.
B. Define registration length as Unicode code points.
C. Define length as grapheme clusters.
D. Normalize Unicode before validation/hash.
E. Keep login whitespace trimming as a special emptiness rule.

## 8. Chosen remediation

Choose **B: Unicode code points** for new-registration length and exact submitted bytes for credential identity.

Canonical semantics:

- Registration: 12..256 Unicode code points inclusive.
- Hash/verify: exact UTF-8 bytes of the submitted password; no trimming, case folding, NFC/NFKC, or other normalization.
- Registration from internal/non-HTTP callers must fail closed on malformed UTF-8.
- Login must not apply the new 12-code-point creation minimum to existing credentials.
- Login emptiness means exactly `password == ""`; whitespace-only passwords must not be rewritten or rejected merely because trimming would make them empty.
- Login accepts only valid UTF-8, non-empty passwords of at most 256 Unicode code points. Invalid/oversized login input must fail as invalid credentials before Argon2, preserving generic authentication failure semantics.
- Login public contract should therefore use 1..256 code points; creation policy belongs to registration.

Expected implementation primitives:

- Go registration: `unicode/utf8.ValidString` + `utf8.RuneCountInString`.
- Go login admission: `utf8.ValidString`, exact empty-string check, and `utf8.RuneCountInString <= 256`; no `strings.TrimSpace` password check.
- Web: explicit code-point count such as `Array.from(password).length`; do not rely on HTML `minLength` as the canonical rule.
- OpenAPI RegisterRequest: retain `minLength: 12`, `maxLength: 256`, document code-point/no-normalization semantics.
- OpenAPI LoginRequest: `minLength: 1`, `maxLength: 256`, document legacy-compatible exact-secret authentication semantics.

Any credential accepted through the historical public JSON registration path was valid Unicode and, because the old registration ceiling was 256 UTF-8 bytes, necessarily had no more than 256 Unicode code points. The login validity/upper-bound rule therefore does not lock out historically valid public credentials.

## 9. Why this solution

It aligns Go, OpenAPI, and Web with one deterministic unit, supports multilingual secrets, preserves exact bytes for Argon2, eliminates the whitespace-only registration/login contradiction, closes the internal Unicode boundary, avoids hidden normalization, and prevents legacy lockout without expanding scope.

## 10. Rejected alternatives

- UTF-8 bytes: implementation detail and poor user-facing semantics.
- Grapheme clusters: unnecessary segmentation complexity for this P3.
- Unicode normalization: changes credential identity and risks compatibility.
- Registration minimum on login: risks locking out historically accepted multibyte credentials.
- Trim-based login emptiness: changes admission based on password content even though registration/hash identity is exact; a valid stored whitespace-only secret could become unusable.

## 11. Trade-offs

A 256-code-point password can use up to 1024 UTF-8 bytes, versus the old effective 256-byte registration ceiling. This input increase is negligible relative to the existing Argon2 64 MiB memory cost and does not authorize any Argon2 parameter/concurrency change.

Code points are intentionally not grapheme clusters; combining sequences can count as multiple characters under this contract.

Registration and login intentionally differ at the lower bound: registration enforces password-creation policy, while login must verify historically accepted exact credentials.

## 12. Required regression tests

Backend:

1. 11 ASCII code points rejected; 12 accepted.
2. 11 Cyrillic code points whose byte count is >=12 rejected; 12 accepted.
3. 256 emoji/supplementary code points accepted despite >256 UTF-8 bytes; 257 rejected.
4. malformed internal UTF-8 registration input rejected.
5. canonically equivalent but byte-distinct Unicode sequences remain distinct secrets; no normalization introduced.
6. legacy stored multibyte password shorter than 12 code points still logs in, while new registration with it is rejected.
7. a valid 12-space password can register and subsequently authenticate unchanged.
8. login rejects only the truly empty password for emptiness and does not trim before verification.
9. malformed internal UTF-8 login input and 257-code-point login input are rejected as invalid credentials before Argon2.

OpenAPI:

10. Register password remains 12..256 with explicit code-point/exact-secret semantics.
11. Login becomes 1..256 with compatibility semantics.
12. contract validation stays green.

Web:

13. registration explicitly counts code points.
14. six emoji do not pass merely because UTF-16 code units reach twelve.
15. 12 code points pass; 256 pass; 257 fail.
16. login can submit the legacy multibyte compatibility vector.
17. login does not trim or block a non-empty whitespace-only exact secret before submission.

## 13. Adversarial review requirements

Reviewer must challenge cross-layer counting parity, malformed-UTF-8 handling, accidental UTF-16 dependence, normalization/trimming, whitespace-only credential handling, legacy lockout, generic login-failure semantics, unintended weakening of registration policy, P3-04 scope creep, and any unauthorized Argon2 changes.

## 14. Remediation iterations

The planning phase recorded two pre-review hardening iterations. First, inspection found that the existing login `strings.TrimSpace` emptiness guard conflicts with exact-secret semantics and registration behavior, so the plan added exact-empty login admission and whitespace-only regression coverage. Second, the login boundary was tightened to valid UTF-8 plus a 256-code-point maximum so the service and OpenAPI share the same Unicode-string model without affecting historical public credentials.

Any later implementation-review blocker must be preserved in the Stage 3.35 dossier. Every changed implementation head requires fresh exact-head CI and fresh independent review.

## 15. Residual risk / limitations

P3-01 closure will not add composition rules, breached-password checks, password reset, passkeys/OAuth/2FA, grapheme-cluster semantics, or general Unicode length fixes. P3-02, P3-04, P3-05, P3-06, P3-09, and P3-10 remain separate.

## 16. Operational / deployment consequences

No migration, rehash campaign, database-column change, Argon2 setting change, or environment change is planned. Existing hashes remain valid because verification continues using the exact submitted password bytes.

## 17. Exact evidence required for closure

Closure must record planning PR/head/CI/review/human authorization/merge; implementation PR/final head/10-check CI/review/remediation iterations/human authorization/merge; and closure-governance PR/head/CI/review/human authorization/merge plus audit-state synchronization.

## 18. Final canonical status rule

P3-01 remains **OPEN** during planning and during any unmerged implementation candidate. It becomes **CLOSED** only after approved planning, separately reviewed implementation, exact-head 10/10 CI, independent `APPROVED`, explicit human squash-merge authorization, canonical implementation merge, and separately approved closure governance.

Until then the original-audit state remains P0=0 / P1=0 / P2=0 / P3=10.

## Proposed implementation boundary

Expected files only:

- `backend-go/internal/auth/service.go`
- focused auth tests
- `openapi/components/schemas.yaml`
- focused OpenAPI contract tests if needed
- `frontend-next/src/features/auth/components/AuthForm.tsx`
- focused Web auth tests
- Stage 3.35 implementation/closure documentation required by governance

`backend-go/internal/auth/password.go` should remain unchanged unless testing exposes an actual exact-byte hashing defect.

## Explicit non-scope

No runtime change is authorized by this planning PR. No password normalization, Argon2 changes, password recovery, OAuth/passkeys/2FA, general P3-04 remediation, timezone work, cleanup-lifecycle work, HTTP decomposition, dependency maintenance, migration, architecture amendment, privacy-lifecycle implementation, or product-scope expansion is authorized.
