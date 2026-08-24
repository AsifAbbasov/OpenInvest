# Stage 3.35 — P3-01 Password Character Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only |
| Date | 2026-08-24 |
| Canonical base | `develop` at `ae5a152114cc163867a363953f8a3202396b1f6c` |
| Finding | P3-01 |
| Scope | Password length semantics only |
| Runtime implementation authorized here | No |

## 1. Finding / symptom

P3-01: password length semantics use bytes rather than the intended user-facing character semantics.

`backend-go/internal/auth/service.go` says passwords must be 12..256 characters but uses Go `len(string)`, which counts UTF-8 bytes. `openapi/components/schemas.yaml` expresses the same rule with `minLength`/`maxLength`, while the Web form relies on browser `minLength`. Non-ASCII input can therefore be counted differently across layers.

## 2. Root cause

No canonical password-length unit was defined. Go byte length, OpenAPI character length, and browser/JavaScript string behavior were treated as equivalent when they are not.

Argon2 hashing itself is not the defect; it should continue receiving the exact submitted password bytes.

## 3. Failure scenarios

- A password with fewer than 12 Cyrillic characters can still occupy at least 12 UTF-8 bytes and pass the current backend minimum.
- A 256-code-point password can exceed 256 UTF-8 bytes and be rejected although the contract permits 256 characters.
- Astral Unicode can produce Web/backend disagreement.
- Enforcing the corrected registration minimum during login could lock out credentials historically accepted under the old byte-based rule.

## 4. Impact

Contract/runtime inconsistency, multilingual UX defects, misleading password policy enforcement, and possible legacy-login lockout during remediation. No plaintext-password exposure or Argon2 cryptographic defect is claimed.

## 5. Severity rationale

P3 is appropriate: the defect is real and security-adjacent but is not known to create privilege escalation, credential disclosure, financial corruption, or systemic availability failure.

## 6. Existing guarantees violated

- OpenAPI and runtime validation must agree.
- User-facing validation must be deterministic across supported clients and backend.
- Maintenance must not invalidate previously accepted credentials.
- Password hashing must preserve the exact secret without hidden normalization.

## 7. Considered solutions

A. Redefine the public policy as UTF-8 bytes.
B. Define registration length as Unicode code points.
C. Define length as grapheme clusters.
D. Normalize Unicode before validation/hash.

## 8. Chosen remediation

Choose **B: Unicode code points** for new-registration length.

Canonical semantics:

- Registration: 12..256 Unicode code points inclusive.
- Hash/verify: exact UTF-8 bytes of the submitted password; no trimming, case folding, NFC/NFKC, or other normalization.
- Registration from internal/non-HTTP callers must fail closed on malformed UTF-8.
- Login must not apply the new 12-code-point creation minimum to existing credentials.
- Login public contract should accept non-empty passwords up to 256 code points; creation policy belongs to registration.

Expected implementation primitives:

- Go: `unicode/utf8.ValidString` + `utf8.RuneCountInString`.
- Web: explicit code-point count such as `Array.from(password).length`; do not rely on HTML `minLength` as the canonical rule.
- OpenAPI RegisterRequest: retain `minLength: 12`, `maxLength: 256`, document code-point/no-normalization semantics.
- OpenAPI LoginRequest: `minLength: 1`, `maxLength: 256`, document legacy-compatible authentication semantics.

## 9. Why this solution

It aligns Go, OpenAPI, and Web with one deterministic unit, supports multilingual secrets, preserves exact bytes for Argon2, avoids hidden normalization, and prevents legacy lockout without expanding scope.

## 10. Rejected alternatives

- UTF-8 bytes: implementation detail and poor user-facing semantics.
- Grapheme clusters: unnecessary segmentation complexity for this P3.
- Unicode normalization: changes credential identity and risks compatibility.
- Registration minimum on login: risks locking out historically accepted multibyte credentials.

## 11. Trade-offs

A 256-code-point password can use up to 1024 UTF-8 bytes, versus the old effective 256-byte registration ceiling. This input increase is negligible relative to the existing Argon2 64 MiB memory cost and does not authorize any Argon2 parameter/concurrency change.

Code points are intentionally not grapheme clusters; combining sequences can count as multiple characters under this contract.

## 12. Required regression tests

Backend:

1. 11 ASCII code points rejected; 12 accepted.
2. 11 Cyrillic code points whose byte count is >=12 rejected; 12 accepted.
3. 256 emoji/supplementary code points accepted despite >256 UTF-8 bytes; 257 rejected.
4. malformed internal UTF-8 registration input rejected.
5. no trimming/normalization introduced.
6. legacy stored multibyte password shorter than 12 code points still logs in, while new registration with it is rejected.

OpenAPI:

7. Register password remains 12..256 with explicit semantics.
8. Login becomes 1..256 with compatibility semantics.
9. contract validation stays green.

Web:

10. registration explicitly counts code points.
11. six emoji do not pass merely because UTF-16 length reaches twelve.
12. 12 code points pass; 256 pass; 257 fail.
13. login can submit the legacy compatibility vector.

## 13. Adversarial review requirements

Reviewer must challenge cross-layer counting parity, accidental UTF-16 dependence, normalization/trimming, legacy lockout, unintended weakening of registration policy, P3-04 scope creep, and any unauthorized Argon2 changes.

## 14. Remediation iterations

Any implementation-review blocker must be preserved in the Stage 3.35 dossier. Every changed implementation head requires fresh exact-head CI and fresh independent review.

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
