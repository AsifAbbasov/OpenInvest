# Stage 3.35 — P3-01 Password Character Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only |
| Date | 2026-08-24 |
| Canonical base | `develop` at `ae5a152114cc163867a363953f8a3202396b1f6c` |
| Finding | P3-01 |
| Scope | Password length, exact-secret admission, and auth HTTP decoding semantics only |
| Runtime implementation authorized here | No |

## 1. Finding / symptom

P3-01: password length semantics use bytes rather than the intended user-facing character semantics.

`backend-go/internal/auth/service.go` says passwords must be 12..256 characters but uses Go `len(string)`, which counts UTF-8 bytes. `openapi/components/schemas.yaml` expresses the same rule with `minLength`/`maxLength`, while the Web form relies on browser `minLength`. Non-ASCII input can therefore be counted differently across layers.

A related exact-secret inconsistency exists in login admission: registration does not trim a password, but login currently rejects `strings.TrimSpace(request.Password) == ""`. A password consisting entirely of whitespace can therefore be accepted at registration and later rejected before exact hash verification.

A transport-boundary inconsistency also exists. The current auth handlers decode request bodies through Go `encoding/json`. That decoder replaces malformed UTF-8 bytes and invalid UTF-16 surrogate escapes in JSON strings with `U+FFFD` instead of necessarily rejecting the request. A later `utf8.ValidString` check in `auth.Service` cannot detect that lossy replacement because the service receives an already-valid Go string.

## 2. Root cause

No canonical password-length unit was defined. Go byte length, OpenAPI character length, and browser/JavaScript string behavior were treated as equivalent when they are not.

The login path also uses whitespace normalization as an emptiness test even though password identity is otherwise exact-byte based.

The HTTP boundary currently relies on permissive `encoding/json` string decoding without an auth-specific rule forbidding lossy replacement of malformed Unicode input before the password reaches the service.

Argon2 hashing itself is not the defect; it should continue receiving the exact UTF-8 bytes of the decoded password string value.

## 3. Failure scenarios

- A password with fewer than 12 Cyrillic characters can still occupy at least 12 UTF-8 bytes and pass the current backend minimum.
- A 256-code-point password can exceed 256 UTF-8 bytes and be rejected although the contract permits 256 characters.
- Astral Unicode can produce Web/backend disagreement.
- Enforcing the corrected registration minimum during login could lock out credentials historically accepted under the old byte-based rule.
- A 12-space password can be accepted at registration but cannot currently authenticate because login treats its trimmed value as empty.
- If login gains a code-point upper bound without also validating UTF-8, internal/non-HTTP callers could still bypass the contract's Unicode-string model with malformed Go strings.
- A malformed raw UTF-8 byte sequence inside an auth password JSON string can be silently replaced by `U+FFFD` during decoding and then pass a service-level `utf8.ValidString` check.
- An invalid/unpaired JSON surrogate escape such as `\uD800` can likewise be decoded as `U+FFFD`, collapsing malformed transport input into the same service value as a legitimately submitted replacement character.
- A Web/JavaScript password can itself contain an unpaired UTF-16 surrogate. `Array.from(password).length` alone does not reject that ill-formed value, and `JSON.stringify` serializes it as a surrogate escape that the strict backend must reject.

## 4. Impact

Contract/runtime inconsistency, multilingual UX defects, misleading password policy enforcement, possible credential lockout during remediation or for exact whitespace-only secrets already accepted by registration, and hidden transport-time password transformation for malformed Unicode input. No plaintext-password exposure or Argon2 cryptographic defect is claimed.

## 5. Severity rationale

P3 is appropriate: the defect is real and security-adjacent but is not known to create privilege escalation, credential disclosure, financial corruption, or systemic availability failure.

## 6. Existing guarantees violated

- OpenAPI and runtime validation must agree.
- User-facing validation must be deterministic across supported clients and backend, including rejection of ill-formed surrogate sequences before submission.
- Maintenance must not invalidate previously accepted conforming credentials.
- Password hashing and authentication must preserve the exact decoded secret without hidden trimming or normalization.
- The auth HTTP boundary must not silently rewrite malformed password input into a different valid Unicode password value before service validation.

## 7. Considered solutions

A. Redefine the public policy as UTF-8 bytes.
B. Define registration length as Unicode code points.
C. Define length as grapheme clusters.
D. Normalize Unicode before validation/hash.
E. Keep login whitespace trimming as a special emptiness rule.
F. Rely on `encoding/json` replacement behavior and validate only after decoding.
G. Add narrow auth-specific fail-closed decoding for malformed Unicode while preserving ordinary JSON escape semantics for valid strings.

## 8. Chosen remediation

Choose **B + G: Unicode code points for password length, exact decoded-string bytes for credential identity, and fail-closed auth HTTP decoding for malformed Unicode input**.

Canonical semantics:

- Registration: 12..256 Unicode code points inclusive.
- Hash/verify: exact UTF-8 bytes of the decoded password string value; no trimming, case folding, NFC/NFKC, or other normalization.
- JSON escape syntax is transport encoding, not part of credential identity. Valid representations that decode to the same Unicode string value represent the same password.
- Auth HTTP register/login decoding must reject malformed raw UTF-8 and invalid/unpaired surrogate escape sequences rather than allow lossy substitution with `U+FFFD` before constructing the password string.
- A legitimately submitted `U+FFFD` code point remains an ordinary valid password character subject to the normal length rules; it must not be conflated with malformed transport input that would otherwise be replaced with `U+FFFD`.
- The supported password character domain is valid Unicode scalar values represented as well-formed UTF-8/UTF-16; unpaired surrogate code units are not admissible password characters. Length is counted by Unicode scalar/code-point iteration after well-formedness is established.
- Registration from internal/non-HTTP callers must fail closed on malformed UTF-8.
- Login must not apply the new 12-code-point creation minimum to existing credentials.
- Login emptiness means exactly `password == ""`; whitespace-only passwords must not be rewritten or rejected merely because trimming would make them empty.
- Login service admission accepts only valid UTF-8, non-empty passwords of at most 256 Unicode code points. Invalid/oversized service input must fail as invalid credentials before Argon2, preserving generic authentication failure semantics.
- Malformed auth HTTP JSON/Unicode input is a request-validation failure and must be rejected before service/Argon2 execution; it is not required to masquerade as a credential mismatch.
- Login public contract should therefore use 1..256 code points; creation policy belongs to registration.

Expected implementation primitives:

- Go registration service: `unicode/utf8.ValidString` + `utf8.RuneCountInString`.
- Go login service admission: `utf8.ValidString`, exact empty-string check, and `utf8.RuneCountInString <= 256`; no `strings.TrimSpace` password check.
- Auth HTTP boundary: a narrow lossless password-decoding guard/DTO path that rejects malformed raw UTF-8 and invalid surrogate escape sequences before they can be replaced with `U+FFFD`. Do not broaden global `decodeStrictJSON` behavior across unrelated endpoints unless implementation review separately proves that change safe and in-scope.
- Web: reject ill-formed JavaScript UTF-16 strings containing unpaired surrogates, then use explicit code-point counting such as `Array.from(password).length`; do not rely on HTML `minLength` as the canonical rule. A built-in `String.prototype.isWellFormed()` check may be used only if supported by the project's configured TypeScript/browser target; otherwise use a small local surrogate-pair validator.
- OpenAPI RegisterRequest: retain `minLength: 12`, `maxLength: 256`, document code-point/no-normalization semantics.
- OpenAPI LoginRequest: `minLength: 1`, `maxLength: 256`, document legacy-compatible exact-secret authentication semantics.

Historically conforming credentials accepted through the public JSON registration contract were Unicode strings and, because the old registration ceiling was 256 UTF-8 bytes, necessarily had no more than 256 Unicode code points. The login validity/upper-bound rule therefore does not lock out those historically conforming public credentials.

Malformed raw UTF-8 or invalid-surrogate auth requests were never conforming public credentials. The current Go decoder may nevertheless have replaced such malformed transport input with `U+FFFD` before hashing. Stage 3.35 does not promise byte-for-byte compatibility for those non-conforming raw request representations; it records that limitation explicitly and prevents any new lossy replacement path. A stored hash produced from a replacement character remains a hash of the decoded Unicode value, not of the original malformed raw bytes.

## 9. Why this solution

It aligns Go, OpenAPI, and Web with one deterministic length unit, supports multilingual secrets, preserves exact decoded-string bytes for Argon2, eliminates the whitespace-only registration/login contradiction, closes both the service Unicode boundary and the auth transport replacement boundary, avoids hidden normalization, and prevents lockout of historically conforming credentials without expanding into general Unicode remediation.

## 10. Rejected alternatives

- UTF-8 bytes: implementation detail and poor user-facing semantics.
- Grapheme clusters: unnecessary segmentation complexity for this P3.
- Unicode normalization: changes credential identity and risks compatibility.
- Registration minimum on login: risks locking out historically accepted multibyte credentials.
- Trim-based login emptiness: changes admission based on password content even though registration/hash identity is exact; a valid stored whitespace-only secret could become unusable.
- Service-only `utf8.ValidString`: insufficient for HTTP requests because permissive JSON decoding can replace malformed input before the service sees it.
- Global JSON-decoder reform as part of P3-01: broader than the password-specific finding and risks silently absorbing P3-04/general input-contract work.

## 11. Trade-offs

A 256-code-point password can use up to 1024 UTF-8 bytes, versus the old effective 256-byte registration ceiling. This input increase is negligible relative to the existing Argon2 64 MiB memory cost and does not authorize any Argon2 parameter/concurrency change.

Code points are intentionally not grapheme clusters; combining sequences can count as multiple characters under this contract.

Registration and login intentionally differ at the lower bound: registration enforces password-creation policy, while login must verify historically accepted exact credentials.

The HTTP boundary gains password-specific decoding/validation responsibility. This is intentional because service-only validation cannot reconstruct whether a valid `U+FFFD` arrived legitimately or was synthesized by a permissive JSON decoder. The implementation must remain narrow and must not become an unrelated `httpapi` refactor.

## 12. Required regression tests

Backend service:

1. 11 ASCII code points rejected; 12 accepted.
2. 11 Cyrillic code points whose byte count is >=12 rejected; 12 accepted.
3. 256 emoji/supplementary code points accepted despite >256 UTF-8 bytes; 257 rejected.
4. malformed internal UTF-8 registration input rejected.
5. canonically equivalent but byte-distinct Unicode sequences remain distinct secrets; no normalization introduced.
6. legacy stored multibyte password shorter than 12 code points still logs in, while new registration with it is rejected.
7. a valid 12-space password can register and subsequently authenticate unchanged.
8. login rejects only the truly empty password for emptiness and does not trim before verification.
9. malformed internal UTF-8 login input and 257-code-point login input are rejected as invalid credentials before Argon2.

HTTP auth boundary:

10. registration request containing malformed raw UTF-8 in the password JSON string is rejected as `400 VALIDATION_ERROR` before service/Argon2 execution.
11. login request containing malformed raw UTF-8 in the password JSON string is rejected as `400 VALIDATION_ERROR` before service/Argon2 execution.
12. registration and login reject invalid/unpaired surrogate escape input such as `\uD800` instead of decoding it to `U+FFFD`.
13. a valid supplementary-plane surrogate pair such as `\uD83D\uDE00` decodes normally and participates in code-point counting as one code point.
14. a legitimately submitted `U+FFFD` character remains valid input and is not rejected merely because malformed input would otherwise have been replaced by the same character.

OpenAPI:

15. Register password remains 12..256 with explicit code-point/exact-decoded-secret semantics.
16. Login becomes 1..256 with compatibility semantics.
17. contract validation stays green.

Web:

18. registration explicitly counts code points.
19. six emoji do not pass merely because UTF-16 code units reach twelve.
20. an unpaired surrogate such as `\uD800` is rejected client-side and is not submitted as a password.
21. a valid surrogate pair / supplementary-plane character counts as one code point.
22. 12 code points pass; 256 pass; 257 fail.
23. login can submit the legacy multibyte compatibility vector.
24. login does not trim or block a non-empty whitespace-only exact secret before submission.

## 13. Adversarial review requirements

Reviewer must challenge cross-layer counting parity, malformed-UTF-8 handling, invalid-surrogate handling on both Web and HTTP boundaries, accidental UTF-16 dependence, transport-time `U+FFFD` replacement, normalization/trimming, whitespace-only credential handling, legacy lockout, generic login-failure semantics, unintended weakening of registration policy, P3-04 scope creep, and any unauthorized Argon2 changes.

## 14. Remediation iterations

The planning phase recorded four hardening iterations before approval.

First, inspection found that the existing login `strings.TrimSpace` emptiness guard conflicts with exact-secret semantics and registration behavior, so the plan added exact-empty login admission and whitespace-only regression coverage.

Second, the login service boundary was tightened to valid UTF-8 plus a 256-code-point maximum so the service and OpenAPI share the same Unicode-string model without affecting historically conforming public credentials.

Third, independent adversarial planning review found that service-level `utf8.ValidString` is insufficient for HTTP auth input because Go `encoding/json` can replace malformed raw UTF-8 and invalid surrogate escapes with `U+FFFD` before the service sees the password. The plan therefore added a narrow fail-closed auth HTTP decoding requirement, transport regression vectors, corrected the historical-compatibility claim to cover conforming public credentials only, and explicitly separated malformed HTTP validation failure from generic service-level credential failure.

Fourth, local adversarial review of the revised candidate found that `Array.from(password).length` alone still accepts a JavaScript string containing an unpaired surrogate even though strict backend decoding must reject the serialized surrogate escape. The plan therefore added Web well-formed-Unicode validation and regression coverage so the client and backend share the same admissible character domain.

Any later implementation-review blocker must be preserved in the Stage 3.35 dossier. Every changed implementation head requires fresh exact-head CI and fresh independent review.

## 15. Residual risk / limitations

P3-01 closure will not add composition rules, breached-password checks, password reset, passkeys/OAuth/2FA, grapheme-cluster semantics, or general Unicode length fixes. P3-02, P3-04, P3-05, P3-06, P3-09, and P3-10 remain separate.

Stage 3.35 does not guarantee compatibility for historically non-conforming malformed raw JSON password representations that may previously have been lossy-decoded to `U+FFFD`. Such requests did not preserve their original raw-byte credential identity under the existing decoder. The remediation instead makes future behavior fail-closed and deterministic.

## 16. Operational / deployment consequences

No migration, rehash campaign, database-column change, Argon2 setting change, or environment change is planned. Existing hashes remain valid because verification continues using the exact decoded password string bytes.

The auth HTTP path gains strict malformed-Unicode rejection before service/Argon2 execution. No general API Unicode-policy change or global HTTP decomposition is authorized.

## 17. Exact evidence required for closure

Closure must record planning PR/head/CI/review/human authorization/merge; implementation PR/final head/10-check CI/review/remediation iterations/human authorization/merge; and closure-governance PR/head/CI/review/human authorization/merge plus audit-state synchronization.

Implementation evidence for the third planning-remediation iteration must include focused auth HTTP tests proving rejection of malformed raw UTF-8 and invalid surrogate escapes without `U+FFFD` substitution, plus service-level proof for malformed internal strings and the existing cross-layer code-point vectors.

## 18. Final canonical status rule

P3-01 remains **OPEN** during planning and during any unmerged implementation candidate. It becomes **CLOSED** only after approved planning, separately reviewed implementation, exact-head 10/10 CI, independent `APPROVED`, explicit human squash-merge authorization, canonical implementation merge, and separately approved closure governance.

Until then the original-audit state remains P0=0 / P1=0 / P2=0 / P3=10.

## Proposed implementation boundary

Expected files only:

- `backend-go/internal/auth/service.go`
- focused auth service tests
- `backend-go/internal/httpapi/api.go` and/or one narrowly scoped auth password-decoding helper inside `backend-go/internal/httpapi`
- focused auth HTTP tests, expected in `backend-go/internal/httpapi/auth_test.go` or a Stage 3.35-focused auth test file
- `openapi/components/schemas.yaml`
- focused OpenAPI contract tests if needed
- `frontend-next/src/features/auth/components/AuthForm.tsx`
- focused Web auth tests
- Stage 3.35 implementation/closure documentation required by governance

`backend-go/internal/auth/password.go` should remain unchanged unless testing exposes an actual exact-byte hashing defect.

A global rewrite of `decodeStrictJSON`, broad HTTP decomposition, or generalized Unicode validation across unrelated request DTOs is not authorized by this plan.

## Explicit non-scope

No runtime change is authorized by this planning PR. No password normalization, Argon2 changes, password recovery, OAuth/passkeys/2FA, general P3-04 remediation, general JSON/Unicode-policy reform outside auth, timezone work, cleanup-lifecycle work, HTTP decomposition, dependency maintenance, migration, architecture amendment, privacy-lifecycle implementation, or product-scope expansion is authorized.
