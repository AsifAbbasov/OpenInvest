# Stage 3.37 — P3-02 True IANA Timezone Semantics Closure

| Field | Value |
| --- | --- |
| Status | Closure candidate / independent closure review pending |
| Date | 2026-08-25 |
| Finding | P3-02 — True IANA Timezone Semantics |
| Planning gate | PR #90 squash-merged at `46f74528dcc19424ad087d30d4f2f778e2079b87` |
| Runtime PR | PR #91 — `fix: enforce Stage 3.37 P3-02 IANA timezone admission` |
| Initial published runtime head | `465a7f0ddfe5a7bf892ec8a735915688cdaf59ad` |
| Frozen final runtime head | `1a2f89a0fa5095b3cca790521afa484bdc61e8a6` |
| Runtime merge | `cb6d9b28cd47b1cd283b5861b916e0be627d0ac2` |
| Exact-head CI | GitHub Actions CI #265 / run `32869754524`, 10/10 required jobs successful |
| Closure merge authorized here | No |

## 1. Finding / symptom

P3-02 identified a contract/runtime mismatch at registration. The canonical model and public contract
described the user's timezone as an IANA timezone-database identifier, but the Go registration
boundary historically proved only that the string was nonblank after `TrimSpace` and within the
existing 64-byte bound. The original untrimmed value was then persisted.

That allowed arbitrary bounded strings such as `Not/AZone`, `Mars/Olympus`, `Local`, and
surrounding-whitespace forms to pass the registration boundary despite violating the intended
preference contract.

The finding did not demonstrate a financial-date defect. Canonical `BusinessDate` remains a calendar
date without timezone, and user timezone remains display/notification-only.

## 2. Root cause

Timezone admission was implemented as generic bounded-string validation instead of resolver-backed
timezone-database validation. PostgreSQL stores `identity.users.timezone` as text and relies on the
application write boundary for semantic admission.

## 3. Failure scenario

A caller could register with a value that was neither a stable timezone-database identifier nor a
portable user preference. In particular, Go's `Local` pseudo-location could resolve to host-local
configuration, and an operator-controlled higher-precedence `ZONEINFO` directory could make otherwise
invalid-looking names resolve.

The first runtime candidate exposed an important adversarial variant: relying only on
`time.LoadLocation` was insufficient to guarantee rejection of surrounding-whitespace and raw-offset
spellings because a deliberately constructed `ZONEINFO` source can contain valid TZif files under
those exact names.

## 4. Impact and severity

The impact remains P3. The defect affected preference correctness, API/runtime parity, and future
display/notification behavior. No financial arithmetic corruption, privilege escalation, data loss,
or demonstrated mutation of financial `BusinessDate` semantics was found.

## 5. Existing guarantees violated

The canonical Stage 2 model defines user timezone as an IANA timezone preference used for display and
notifications only. Runtime registration did not enforce that published semantic boundary and could
persist arbitrary bounded strings.

## 6. Chosen remediation

PR #91 implemented the approved narrow remediation:

- validate the exact submitted timezone string without trim/case/alias normalization;
- reject empty input, the existing >64-byte overflow, and exact `Local` before resolver admission;
- reject surrounding-whitespace forms through comparison-only `strings.TrimSpace(name) != name`;
- reject raw ASCII `±HH:MM` and `UTC±HH:MM` spellings before resolver admission;
- explicitly retain `UTC`;
- delegate all other exact identifiers to `time.LoadLocation(name)`;
- retain valid loadable tzdb identifiers such as `Etc/GMT+4`;
- import `_ "time/tzdata"` only for standard-library fallback availability;
- persist an accepted timezone exactly as submitted;
- reject invalid admission before `Store.RegisterUser`;
- align `User.timezone` and `RegisterRequest.timezone` OpenAPI descriptions with the runtime rule.

No migration, historical preference rewrite, frontend change, financial-date change, or unrelated P3
remediation was included.

## 7. Why this solution

Using Go's resolver avoids a private IANA list that can drift while preserving the platform's existing
timezone-data model. The narrow pre-resolver guard closes contract requirements that must remain true
even when a higher-precedence timezone source deliberately contains unusual file names.

Exact-value validation also avoids silently correcting caller mistakes and preserves the persisted
preference identity.

## 8. Rejected alternatives

- Nonblank/length-only validation — leaves the original contract mismatch open.
- Trim or normalize then persist — silently changes caller input and stored identity.
- Handcrafted IANA regex/private list — duplicates timezone-database authority and can drift.
- Reject all fixed-offset timezone-database identifiers — incorrectly rejects loadable names such as
  `Etc/GMT+4`.
- Treat embedded `time/tzdata` as exclusive authority — contradicts `time.LoadLocation` source
  precedence.
- PostgreSQL timezone-membership enforcement — unnecessary environment/version coupling for this
  finding.
- Automatic historical backfill — would guess user preference data.
- Apply user timezone to financial `BusinessDate` math — violates the canonical financial-date model.

## 9. Runtime verification evidence

The final runtime candidate proves:

- `UTC`, `Asia/Baku`, `Europe/Berlin`, `America/New_York`, and loadable `Etc/GMT+4` admission;
- rejection of empty/whitespace input, `Local`, unknown/invented zones, path-like invalid values,
  raw offsets, surrounding whitespace, and existing byte-bound overflow;
- positive and negative raw-offset families:
  `+04:00`, `-04:00`, `UTC+04:00`, `UTC-04:00`;
- surrounding-whitespace and raw-offset rejection in the application syntax guard independently of
  resolver behavior;
- exact accepted `Etc/GMT+4` persistence;
- invalid service-level admission never reaching `Store.RegisterUser`;
- HTTP 400 plus zero registration-store calls for `" Asia/Baku"` and `"UTC+04:00"`;
- OpenAPI descriptions and 1..64 registration bounds without a handcrafted timezone pattern.

## 10. Exact-head CI

The initial published runtime head `465a7f0ddfe5a7bf892ec8a735915688cdaf59ad` passed historical
CI #264 / run `32867005056`, but that evidence ceased to be exact-head evidence after the
documentation correction advanced the PR.

The frozen final runtime head `1a2f89a0fa5095b3cca790521afa484bdc61e8a6` passed GitHub Actions
CI #265 / run `32869754524` with all 10 required checks successful:

- Go tests;
- Python tests;
- Frontend build and typecheck;
- OpenAPI contract;
- Docker Compose config;
- PostgreSQL migration validation;
- Go vet;
- Go race tests;
- Go vulnerability scan;
- Dependency security scan.

## 11. Adversarial review history

The remediation required two independent `REQUEST CHANGES` cycles at different lifecycle gates.

First local pre-commit review:
- `REQUEST CHANGES` because whitespace/raw-offset invalidity was resolver-dependent under a custom
  higher-precedence `ZONEINFO` source.
- Remediation added only the narrow pre-resolver whitespace/raw-offset syntax guard.
- Renewed local pre-commit review returned `APPROVED` with P0/P1/P2/P3 = None.

First published-head review on `465a7f0ddfe5a7bf892ec8a735915688cdaf59ad`:
- runtime implementation was accepted;
- `REQUEST CHANGES` was issued only because the implementation record still falsely described the
  already-published candidate as uncommitted/local.
- A one-file documentation correction was independently pre-commit reviewed and `APPROVED`.

Final published-head review on `1a2f89a0fa5095b3cca790521afa484bdc61e8a6`:
- independently verified base/head identity;
- independently verified CI #265 / run `32869754524` and all 10 jobs;
- returned `APPROVED` with P0/P1/P2/P3 = None;
- explicitly found the exact head suitable for separate Ready + squash-merge authorization.

## 12. Runtime merge evidence

The user separately and explicitly authorized Ready + squash merge of exact head
`1a2f89a0fa5095b3cca790521afa484bdc61e8a6`.

PR #91 was moved from Draft to Ready and squash-merged. GitHub returned canonical merge SHA:

`cb6d9b28cd47b1cd283b5861b916e0be627d0ac2`

`develop` was then independently read back and confirmed to point exactly at that SHA.

Runtime merge does not itself close P3-02.

## 13. Remediation iterations

1. Planning PR #90 established the resolver authority model, exact identity rule, non-scope, and
   historical-data boundary.
2. Initial runtime candidate added resolver-backed validation, OpenAPI wording, tests, and fallback
   `time/tzdata`.
3. First pre-commit `REQUEST CHANGES` exposed the custom-`ZONEINFO` bypass.
4. Revised runtime candidate added the narrow pre-resolver syntax guard and received renewed
   pre-commit `APPROVED`.
5. Initial published head `465a7f0d...` passed CI #264.
6. First published-head review found only stale implementation-record lifecycle wording.
7. The independently approved one-file documentation correction advanced the head to `1a2f89a0...`.
8. Final head passed CI #265 and fresh published-head `APPROVED`.
9. PR #91 was explicitly authorized and squash-merged as `cb6d9b28...`.
10. This closure package synchronizes governance state without changing runtime behavior.

## 14. Residual risk / limitations

P3-02 closure does not guess or rewrite timezone preferences stored before remediation. Historical
rows may therefore contain values that would no longer pass new-registration admission. The repository
contains no authoritative fact from which to infer the user's intended timezone.

Higher-precedence timezone data can still affect resolver admission and case behavior. This is part of
the approved operational authority model rather than an embedded-only tzdb guarantee.

Any future timezone-preference update write path must reuse the same admission semantics.

## 15. Explicit non-scope

P3-02 does not close or absorb:

- P3-04 general Unicode / OpenAPI `maxLength` semantics;
- P3-05 idempotency/session cleanup lifecycle;
- P3-06 `httpapi/api.go` decomposition;
- P3-07 transaction-form fixture/default cleanup;
- P3-08 migration-validator policy hardening;
- P3-09 Next.js maintenance;
- P3-10 Fiber maintenance;
- Stage 3.25 privacy Security Review evidence collection.

It also changes no BusinessDate/SQL `DATE` rule, UTC SystemTimestamp rule, financial calculation,
migration, database schema, frontend behavior, dependency, provider integration, or historic row.

## 16. Operational / deployment consequences

No migration, backfill, reindex, worker, scheduler, provider, credential change, or manual data rewrite
is required. The Go binary includes standard-library timezone-data fallback availability; configured
higher-precedence timezone sources remain operationally governed.

## 17. Exact evidence

- Planning PR: #90.
- Planning merge: `46f74528dcc19424ad087d30d4f2f778e2079b87`.
- Initial published runtime head: `465a7f0ddfe5a7bf892ec8a735915688cdaf59ad`.
- Historical initial-head CI: #264 / run `32867005056`.
- Frozen final runtime head: `1a2f89a0fa5095b3cca790521afa484bdc61e8a6`.
- Final exact-head CI: #265 / run `32869754524`; 10/10 required jobs successful.
- Runtime PR: #91.
- Canonical runtime squash merge: `cb6d9b28cd47b1cd283b5861b916e0be627d0ac2`.
- Canonical branch read-back: `develop` pointed exactly at the runtime squash merge after PR #91.
- Closure package: this document plus synchronized `SOURCE_OF_TRUTH.md`, `ROADMAP.md`, and the
  Stage 3.37 implementation record.

## 18. Final canonical status rule

At the base of this closure candidate, the canonical repository audit count remains:

- P0: 0
- P1: 0
- P2: 0
- P3: 8

P3-02 runtime remediation is canonical in `develop`, but this unmerged closure package does not itself
declare the finding canonically CLOSED.

P3-02 becomes CLOSED only when this exact closure package receives fresh independent `APPROVED`
review, receives exact-head green closure CI after publication, the user separately and explicitly
authorizes the closure squash merge, and the closure PR is merged into `develop`.

The resulting post-closure backlog is P0=0 / P1=0 / P2=0 / P3=7, consisting of P3-04, P3-05,
P3-06, P3-07, P3-08, P3-09, and P3-10. Stage 3.25 remains separate.
