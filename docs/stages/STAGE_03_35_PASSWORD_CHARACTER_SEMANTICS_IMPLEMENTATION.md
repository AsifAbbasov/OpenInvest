# Stage 3.35 — P3-01 Password Character Semantics Implementation

| Field | Value |
| --- | --- |
| Status | Runtime implementation completed / canonical through PR #84; governance closure through PR #85 |
| Date | 2026-08-24 |
| Canonical base | `develop` at `2ac48b1333e121e2f9fa6722f6962202728082e7` |
| Planning gate | PR #83 squash-merged |
| Finding | P3-01 |
| Runtime merge authorized here | No |

## Objective

Implement the approved Stage 3.35 password semantic model without changing Argon2 parameters, global JSON decoding, migrations, architecture, or the broader P3-04 Unicode policy.

## Invariant

New registration accepts only well-formed Unicode passwords of 12..256 Unicode code points. Login accepts well-formed Unicode passwords of 1..256 code points to preserve historically conforming multibyte credentials. Credential identity remains the exact UTF-8 bytes of the decoded password string, with no trimming or Unicode normalization. Auth HTTP transport must reject malformed UTF-8 and unpaired surrogate escapes before Go `encoding/json` can replace them with `U+FFFD`.

## Implementation boundary

Production changes are limited to:

- `backend-go/internal/auth/service.go`
- `backend-go/internal/httpapi/api.go`
- `backend-go/internal/httpapi/auth_password_json.go`
- `openapi/components/schemas.yaml`
- `frontend-next/src/features/auth/components/AuthForm.tsx`
- `frontend-next/src/features/auth/passwordPolicy.ts`

Focused regression evidence is added under the existing auth/httpapi/OpenAPI/Web test surfaces.

## Implementation summary

- Registration uses `utf8.ValidString` and `utf8.RuneCountInString` for the 12..256 policy.
- Login removes password `TrimSpace` admission and instead requires valid UTF-8, non-empty exact input, and at most 256 code points before store lookup/Argon2.
- `password.go` and all Argon2 parameters/concurrency remain unchanged.
- Register/login DTO password fields use an auth-specific JSON scalar type whose `UnmarshalJSON` validates raw UTF-8 and UTF-16 surrogate escapes before ordinary string decoding.
- The global `decodeStrictJSON` function is unchanged.
- OpenAPI keeps registration at 12..256 and changes login to 1..256 with exact-secret/no-normalization documentation.
- Web rejects unpaired UTF-16 surrogates and counts code points explicitly with `Array.from` before submission; shared browser `minLength={12}` is removed.

## Migrations / ADR / external data / mathematics

- Database or schema migration: None.
- ADR change: None; this implements the separately approved Stage 3.35 plan.
- External data source change: None.
- Financial or mathematical calculation change: None.

## Security / privacy / compatibility

The change prevents malformed transport input from being silently rewritten into a different valid credential. It does not expose plaintext or change password hashing. Historically conforming multibyte credentials remain eligible for login even when they contain fewer than 12 code points. Whitespace-only exact secrets are no longer rejected merely because trimming would make them empty.

## Local validation evidence

Focused helper validation completed on the frozen local candidate:

- Go auth password JSON Unicode vectors: PASS.
- Node Web password-policy vectors: PASS, 4/4.
- Remediation iteration: an initial Web well-formedness implementation incorrectly accepted a terminal high surrogate because `charCodeAt` returned `NaN` out of range; the candidate now checks the index boundary explicitly and the regression suite passes.

The following are **NOT_VERIFIED** locally in this environment because a full repository checkout was not available:

- `gofmt` across the complete repository diff beyond the standalone Go candidate files;
- `go test ./...` against the complete repository;
- `go vet ./...`;
- race-enabled repository tests;
- `pnpm run typecheck`;
- complete `pnpm test`;
- `pnpm run build`;
- repository OpenAPI validator.

These items must not be represented as PASS before authoritative exact-head CI or an equivalent complete local checkout executes them.

## Internal Review Evidence

`WITHHELD — blind external review pending`.

This is a governance placeholder only. It does not assert that the pre-commit candidate has already received an internal `APPROVED` verdict. The actual internal findings/verdict remain out of the future PR until the independent external review gate is complete, as required by `docs/REVIEW_WORKFLOW.md`.

## Rollback

The implementation is code/contract-only with no migration or stored-data rewrite. Rollback is the inverse code change while preserving existing password hashes. No rehash campaign is required.

## Out of scope

- Argon2 parameters or concurrency.
- Password reset, breached-password checks, passkeys, OAuth, or MFA.
- Unicode normalization or grapheme-cluster semantics.
- Global JSON decoder reform.
- General P3-04 string/maxLength remediation.
- P3-02, P3-05, P3-06, P3-07, P3-08, P3-09, P3-10.

## Canonical status

P3-01 runtime implementation is **CLOSED** through PR #84 at `a47df19ccc7edff73f39f4e76aec47580c168c46`. Governance closure is completed separately through PR #85. The finding is not considered canonically CLOSED until the governance closure change is merged.
