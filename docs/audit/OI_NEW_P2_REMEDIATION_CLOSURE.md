# OI-NEW P2 Remediation Closure

## 1. Purpose

This record permanently captures the completed OI-NEW P2 remediation set as technical repository history. It records the original defects, technical corrections, preserved invariants, regression coverage, canonical pull requests, canonical merge commits, and final arithmetic.

## 2. Original P2 audit scope

- OI-NEW-01 — PostgreSQL runtime privilege boundary.
- OI-NEW-02 — authentication/storage error distinction.
- OI-NEW-03 — readiness/startup integrity-check separation.
- OI-NEW-04 — bounded retroactive replay with exact weighted-average cost.
- OI-NEW-05 — HTTP server timeout hardening.
- OI-NEW-06 — trusted proxy and canonical client-IP semantics.
- OI-NEW-07 — frontend transient-auth state preservation.
- OI-NEW-08 — Portfolio Summary `asOfDate` contract/runtime alignment.

## 3. Final status matrix

| Finding | Status | PR | Canonical merge |
| --- | --- | ---: | --- |
| OI-NEW-01 | CLOSED | #184 | `e72dd4d11c3143e2fc4db08c5d02cb6d3c80bcdf` |
| OI-NEW-02 | CLOSED | #183 | `8ad6261ae61d057669f6e1a554e029ecc086394f` |
| OI-NEW-03 | CLOSED | #186 | `a043c2143cfb7b08062f2375cd4e548e16ddcb90` |
| OI-NEW-04 | CLOSED | #187 | `eeb0554fe1a16d4668acdb5db37df8002e8434e0` |
| OI-NEW-05 | CLOSED | #185 | `bc72f4be91a1bbdb17132271fac023fd212def36` |
| OI-NEW-06 | CLOSED | #185 | `bc72f4be91a1bbdb17132271fac023fd212def36` |
| OI-NEW-07 | CLOSED | #183 | `8ad6261ae61d057669f6e1a554e029ecc086394f` |
| OI-NEW-08 | CLOSED | #188 | `d9f6c263dd3ff4d22fc978372014381880104489` |

## 4. OI-NEW-01

**Original problem.** PostgreSQL runtime permissions exceeded the intended application capability boundary.

**Why it mattered.** Excess privileges could bypass append-only and least-privilege assumptions.

**Root cause.** Runtime access was not expressed as one exact table/column capability contract with fail-closed startup enforcement.

**Technical remediation.** Runtime privilege state is reconstructed atomically to the expected table/column capability set. Append-only constraints, role membership, ownership escape paths, forbidden `CREATE`, grant-option, and security-definer capabilities are checked explicitly. Startup validates the boundary and fails closed.

**Why this remediation.** Database authorization is enforced where persisted financial truth lives rather than relying only on application conventions.

**Preserved invariants.** Ownership isolation, append-only financial history, and financial methodology remain unchanged.

**Regression/tests.** Privilege-contract tests cover permitted and forbidden capabilities and startup validation.

**Canonical record.** PR #184; canonical merge `e72dd4d11c3143e2fc4db08c5d02cb6d3c80bcdf`; required CI 10/10 PASS.

**Status.** `OI_NEW_01=CLOSED`.

## 5. OI-NEW-02

**Original problem.** Storage/infrastructure failures on authentication paths could be represented as credential or session rejection.

**Why it mattered.** Operational failures became indistinguishable from authentication failures.

**Root cause.** Error classification collapsed invalid-auth cases and store failures into the same outward mapping.

**Technical remediation.** Invalid credentials remain authentication failures; invalid sessions remain session failures; storage/infrastructure faults map to sanitized `INTERNAL_ERROR`. The dummy password path remains intact.

**Why this remediation.** Authentication truth and infrastructure health are separate failure domains.

**Preserved invariants.** Password verification behavior, enumeration resistance, and sanitized client errors remain intact.

**Regression/tests.** Auth-path tests cover invalid credentials, invalid sessions, store failures, and dummy verification.

**Canonical record.** PR #183; canonical merge `8ad6261ae61d057669f6e1a554e029ecc086394f`; required CI 10/10 PASS.

**Status.** `OI_NEW_02=CLOSED`.

## 6. OI-NEW-03

**Original problem.** Public readiness could repeatedly execute deep integrity work.

**Why it mattered.** A health probe could amplify database load and make availability checks expensive.

**Root cause.** Deep integrity validation and cheap readiness evidence were not separated by lifecycle.

**Technical remediation.** Public `/ready` performs a cheap bounded database `Ping`. Deep integrity checks execute at controlled startup, which fails closed on error.

**Why this remediation.** Expensive invariants belong at a controlled lifecycle boundary; public probes must remain cheap and bounded.

**Preserved invariants.** Integrity checks still gate successful startup and the public readiness contract remains truthful.

**Regression/tests.** Startup and readiness tests prove the separation and fail-closed behavior.

**Canonical record.** PR #186; canonical merge `a043c2143cfb7b08062f2375cd4e548e16ddcb90`; required CI 10/10 PASS.

**Status.** `OI_NEW_03=CLOSED`.

## 7. OI-NEW-04

**Original problem.** Retroactive changes could force replay across unbounded immutable transaction history.

**Why it mattered.** Reversal-heavy and backdated histories could make snapshot reconstruction scale with full ledger age.

**Root cause.** Exact path-dependent weighted-average cost required ordered replay, but no immutable replay boundary bounded the mutable horizon.

**Technical remediation.** An immutable replay epoch and bounded retroactive replay window preserve the existing exact weighted-average-cost algorithm. The hard bound is `B=5000` raw immutable `transaction_entries` rows. Once activated, replay uses frozen boundary state plus the mutable suffix and does not silently fall back to full-history replay.

**Why this remediation.** Algebraic subtraction is unsafe for path-dependent weighted-average cost; freezing a proven prefix preserves ordered semantics while bounding future work.

**Preserved invariants.** Decimal behavior, exact weighted-average cost, correction/reversal truth, immutable ledger history, and snapshot semantics remain authoritative.

**Regression/tests.** Coverage includes backdated operations, reversal-heavy sequences, epoch generation lifecycle, bound enforcement, and exact result parity.

**Canonical record.** PR #187; canonical merge `eeb0554fe1a16d4668acdb5db37df8002e8434e0`; required CI 10/10 PASS.

**Operational reference.** [`../operations/OI_NEW_04_REPLAY_POLICY_OPERATIONS.md`](../operations/OI_NEW_04_REPLAY_POLICY_OPERATIONS.md).

**Status.** `OI_NEW_04=CLOSED`.

## 8. OI-NEW-05

**Original problem.** HTTP server timeout boundaries were not explicit.

**Why it mattered.** Slow or stalled clients could consume server resources without the intended time bounds.

**Root cause.** Runtime server configuration omitted the required timeout policy.

**Technical remediation.** `ReadTimeout=15s`, `WriteTimeout=30s`, `IdleTimeout=60s`.

**Why this remediation.** Explicit server timeouts provide predictable resource bounds without changing endpoint semantics.

**Preserved invariants.** API contracts and financial behavior are unchanged.

**Regression/tests.** Runtime configuration coverage asserts the timeout values.

**Canonical record.** PR #185; canonical merge `bc72f4be91a1bbdb17132271fac023fd212def36`; required CI 10/10 PASS.

**Status.** `OI_NEW_05=CLOSED`.

## 9. OI-NEW-06

**Original problem.** Client-IP derivation required a strict boundary between direct traffic and explicitly trusted proxy traffic.

**Why it mattered.** Blindly accepting forwarding headers would allow spoofed client identity.

**Root cause.** Proxy trust and forwarded-header interpretation were not represented as one explicit validated policy.

**Technical remediation.** Direct mode ignores forwarding headers. Trusted-proxy mode is explicit and accepts only validated CIDR/IP allowlists; trust-all ranges are rejected. `X-Forwarded-For` is interpreted only in trusted mode using right-to-left trusted-hop resolution, and the resulting client IP is canonicalized.

**Why this remediation.** Forwarded identity is trustworthy only when every accepted hop is governed by an explicit trust boundary.

**Preserved invariants.** Direct deployments continue to use the TCP peer identity; unrelated auth and financial behavior are unchanged.

**Regression/tests.** Coverage includes spoof attempts, trusted/untrusted hop chains, invalid trust ranges, and canonical representation.

**Canonical record.** PR #185; canonical merge `bc72f4be91a1bbdb17132271fac023fd212def36`; required CI 10/10 PASS.

**Status.** `OI_NEW_06=CLOSED`.

## 10. OI-NEW-07

**Original problem.** Transient refresh/logout failures could incorrectly destroy authentication state, while stale refresh completion could race with successful logout.

**Why it mattered.** Network/5xx/429 failures are not proof that a session is invalid, and stale asynchronous completion must not resurrect cleared auth state.

**Root cause.** Frontend state transitions did not fully separate authoritative auth rejection from transient transport/server failure and logout lifecycle ordering.

**Technical remediation.** Refresh `401` may clear auth state. `429`, `5xx`, and network failures preserve auth state. Failed logout remains retryable; successful logout clears state; stale refresh completion cannot restore state after successful logout.

**Why this remediation.** Local state now follows authoritative session evidence instead of treating transient availability failures as authentication truth.

**Preserved invariants.** Successful auth rejection still clears invalid state and successful logout remains final.

**Regression/tests.** Frontend state-machine coverage exercises refresh error classes, retryable logout, successful logout, and stale refresh races.

**Canonical record.** PR #183; canonical merge `8ad6261ae61d057669f6e1a554e029ecc086394f`; required CI 10/10 PASS.

**Status.** `OI_NEW_07=CLOSED`.

## 11. OI-NEW-08

**Original problem.** Portfolio Summary OpenAPI `asOfDate` semantics did not match runtime behavior.

**Why it mattered.** A client could infer a MOEX-calendar default that the endpoint did not use, and an explicitly empty query could collapse into omitted mode.

**Root cause.** Portfolio Summary reused a shared query-parameter description whose default semantics belonged elsewhere, while the HTTP boundary used a raw query accessor.

**Technical remediation.** Portfolio Summary owns an endpoint-local optional `BusinessDate` contract. Explicit `asOfDate` selects the latest calculated snapshot with `snapshot_date <= requested date`; `data.asOfDate` reports the actual selected snapshot date. Omitted `asOfDate` selects the latest available accepted calculated snapshot with no hidden wall-clock cutoff and no implicit MOEX business-calendar cutoff. Explicit empty and invalid values return `400`; a date before the earliest eligible snapshot remains `404`.

**Why this remediation.** The public contract states existing runtime truth instead of introducing new time/calendar behavior.

**Preserved invariants.** Financial methodology is unchanged. No migration was added. Dashboard retains its shared parameter contract.

**Regression/tests.** HTTP, service, PostgreSQL, and OpenAPI coverage prove explicit selection, actual selected date reporting, future-dated omission behavior, invalid input handling, before-earliest behavior, and Dashboard non-drift.

**Canonical record.** PR #188; canonical merge `d9f6c263dd3ff4d22fc978372014381880104489`; required CI 10/10 PASS.

**Status.** `OI_NEW_08=CLOSED`.

## 12. Cross-cutting invariants

- append-only financial history and correction/reversal truth;
- exact Decimal and weighted-average-cost behavior;
- backend ownership of financial methodology;
- ownership isolation and sanitized failure surfaces;
- explicit infrastructure trust boundaries;
- bounded public health checks;
- endpoint contracts that state runtime truth;
- no implicit activation of external market-data behavior.

## 13. Canonical PR / commit / CI ledger

| PR | Findings | Canonical merge | Required CI |
| ---: | --- | --- | --- |
| #183 | OI-NEW-02, OI-NEW-07 | `8ad6261ae61d057669f6e1a554e029ecc086394f` | 10/10 PASS |
| #184 | OI-NEW-01 | `e72dd4d11c3143e2fc4db08c5d02cb6d3c80bcdf` | 10/10 PASS |
| #185 | OI-NEW-05, OI-NEW-06 | `bc72f4be91a1bbdb17132271fac023fd212def36` | 10/10 PASS |
| #186 | OI-NEW-03 | `a043c2143cfb7b08062f2375cd4e548e16ddcb90` | 10/10 PASS |
| #187 | OI-NEW-04 | `eeb0554fe1a16d4668acdb5db37df8002e8434e0` | 10/10 PASS |
| #188 | OI-NEW-08 | `d9f6c263dd3ff4d22fc978372014381880104489` | 10/10 PASS |

## 14. Current canonical state

- `develop@d9f6c263dd3ff4d22fc978372014381880104489`
- tree `1c93cc51a25da3fdc8207da8e94aa1c0d1ab0d7a`
- OI-NEW P2 remediation is closed `8/8`.
- no P2 finding remains open.

## 15. Residual non-P2 scope

This record does not authorize Stage 3.78+, P3, provider activation, or additional runtime scope.

## 16. Final closure statement

```text
P2_TOTAL=8
P2_CLOSED=8
P2_REMAINING=0

P2_PHASE_CLOSED=YES
```

The OI-NEW P2 technical remediation set is complete at `develop@d9f6c263dd3ff4d22fc978372014381880104489`.

## Migration-validator documentation binding reconciliation

The Stage 3.54 machine-contract registries remain structurally unchanged while project-development attribution is removed from the human-readable plan.

- historical frozen plan SHA-256: `c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba`;
- current cleaned plan SHA-256: `f2b598c396bdcdda4168fca9b7ef51db0f3b154b3acfc2abc06501cffba897b1`;
- validator binding files: `backend-go/cmd/validate-migrations/policy.go` and `backend-go/cmd/validate-migrations/stage355_contract_test.go`;
- each binding file changes only the exact pinned plan SHA-256 literal;
- migration SQL, manifest semantics, parser behavior, validation rules, runtime behavior, database schema, financial methodology, OpenAPI and CI workflows are unchanged.

`MIGRATION_VALIDATOR_DOCUMENTATION_BINDING_CHANGED=YES`.
