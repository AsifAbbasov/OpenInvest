# Stage 3.37 - P3-02 True IANA Timezone Semantics Plan

| Field | Value |
| --- | --- |
| Status | Planning/review gate only |
| Date | 2026-08-25 |
| Canonical base | `develop` at `9c83b68e28bbb8bc971620d3e00be5e177ce0820` |
| Finding | P3-02 |
| Scope | Registration-timezone admission parity with the canonical IANA timezone contract |
| Runtime implementation authorized here | No |

## 1. Finding / symptom

P3-02: the public/canonical user timezone contract says the persisted preference is an IANA timezone
identifier, while the current Go registration boundary proves only that the submitted string is not
blank after `TrimSpace` and is no longer than 64 bytes. `Register` then persists the original,
untrimmed value.

Consequently values such as `Not/AZone`, `Mars/Olympus`, `Local`, and surrounding-whitespace forms can
pass current registration validation even though they are not acceptable persisted user timezone
preferences under the canonical contract.

This is a preference-contract/runtime mismatch. It is not evidence that financial BusinessDate values
are currently shifted by user timezone: the canonical model defines user timezone as display/
notification-only and explicitly excludes it from financial-date math.

## 2. Root cause and severity rationale

The registration validator treats timezone as a bounded nonblank string rather than a timezone-database
identifier. PostgreSQL stores the preference as `TEXT NOT NULL DEFAULT 'UTC'` and relies on the
application write boundary for semantic admission.

The finding remains P3 because the defect affects preference correctness and future display/
notification behavior, while no demonstrated financial corruption, privilege escalation, or data-loss
path follows from the current implementation. It nevertheless requires remediation because callers can
persist values the API and canonical model describe more narrowly.

## 3. Chosen authority model

The future implementation must validate the exact submitted timezone string using Go's documented
`time.LoadLocation` semantics.

The plan does not claim that imported `time/tzdata` is the exclusive timezone-membership authority.
`time.LoadLocation` may resolve timezone data through `ZONEINFO`, the host system timezone database,
`$GOROOT/lib/time/zoneinfo.zip`, and finally the standard-library `time/tzdata` fallback when imported.
The implementation should import `_ "time/tzdata"` so minimal deployment images still contain a
standard-library timezone database fallback.

Deployment-specific replacement of higher-precedence timezone sources is operational configuration and
must remain governed. Stage 3.37 does not introduce a custom embedded-only resolver or a third-party
timezone dependency.

## 4. Exact admission contract

Timezone validation must operate on the exact decoded JSON string.

1. Reject empty input before calling `time.LoadLocation`.
2. Preserve the existing bounded admission requirement. The general Unicode/OpenAPI `maxLength`
   semantics remain P3-04 and are not solved here.
3. Do not trim, lowercase, uppercase, alias-rewrite, or otherwise normalize the identifier.
4. Explicitly reject `Local` before calling `time.LoadLocation`, because Go special-cases it as the
   host-local location and therefore its persisted meaning would depend on deployment configuration.
5. Continue accepting `UTC` for compatibility with the current database default and Web registration
   default.
6. For every other exact string, accept only when `time.LoadLocation(name)` succeeds.
7. Do not maintain a handcrafted regex or private IANA-name list in parallel with the resolver.

Representative valid values:

- `UTC`
- `Asia/Baku`
- `Europe/Berlin`
- `America/New_York`
- `Etc/GMT+4` when `time.LoadLocation("Etc/GMT+4")` succeeds

Representative invalid values:

- empty string
- whitespace-only string
- `Local`
- `Not/AZone`
- `Mars/Olympus`
- raw offset spelling `+04:00`
- raw offset spelling `UTC+04:00`
- ` Asia/Baku`
- `Asia/Baku `

A valid timezone-database identifier must not be rejected merely because its effective rules are a
fixed offset. In particular, the implementation must not add a rule that rejects loadable `Etc/GMT*`
identifiers.

## 5. OpenAPI clarification boundary

The future runtime remediation may update `openapi/components/schemas.yaml` so both
`RegisterRequest.timezone` and `User.timezone` clearly describe the same resolver-based contract.
Suitable wording is conceptually:

> IANA timezone database identifier accepted by the server timezone database resolution. `UTC` is
> supported. The host-local pseudo-zone `Local` and raw UTC-offset strings such as `+04:00` or
> `UTC+04:00` are not accepted.

No regex should attempt to enumerate the IANA database. The OpenAPI change must not claim that imported
`time/tzdata` overrides higher-precedence `LoadLocation` sources or that all fixed-offset IANA zones are
forbidden.

## 6. Persistence and compatibility boundary

No PostgreSQL migration is planned. `identity.users.timezone` may remain `TEXT NOT NULL DEFAULT 'UTC'`;
application admission is the current semantic write boundary.

Existing persisted timezone strings must not be silently rewritten or guessed. The historical runtime
could have accepted arbitrary nonblank values within the existing bound, and the repository contains no
geographical fact from which the user's intended preference can be reconstructed. Historical invalid
values therefore remain a documented compatibility/operational risk unless future product scope adds an
explicit user-confirmed correction path.

Any future user-preference update write path must reuse the same timezone-admission rule rather than
creating a second timezone contract.

## 7. Future implementation boundary

After this planning gate is independently reviewed, explicitly authorized, and canonically merged, the
runtime remediation may be limited to:

- `backend-go/internal/auth/service.go` or a narrowly extracted auth-timezone helper implementing the
  exact admission contract;
- the standard-library `_ "time/tzdata"` import required for fallback availability;
- focused auth service tests proving valid/invalid resolver cases and proving invalid registration never
  reaches `Store.RegisterUser`;
- focused `backend-go/internal/httpapi` tests proving an invalid timezone maps to the established client
  validation response and cannot become a successful registration;
- `openapi/components/schemas.yaml` description-only clarification needed to publish the exact same
  semantics; and
- Stage 3.37 implementation and later closure-governance evidence.

Frontend changes are not required merely because the existing registration UI uses a free-text timezone
field. The Go API remains the authoritative write boundary. A frontend change may enter the runtime PR
only if implementation evidence finds an actual contract defect that cannot be fixed at the API boundary;
that would require renewed review before mutation.

## 8. Required regression proof

The implementation must prove at least:

1. `UTC`, `Asia/Baku`, `Europe/Berlin`, `America/New_York`, and a loadable `Etc/GMT+4` are accepted.
2. `""`, whitespace-only input, `Local`, `Not/AZone`, `Mars/Olympus`, `+04:00`, `UTC+04:00`, leading
   whitespace, and trailing whitespace are rejected.
3. Accepted values are persisted exactly as submitted; no transport or service trimming/normalization
   occurs.
4. Invalid values produce no successful `Store.RegisterUser` call.
5. HTTP registration with an invalid timezone follows the established deterministic client-validation
   path and performs no successful user persistence.
6. The application binary includes the standard-library timezone-data fallback without tests falsely
   asserting that embedded `time/tzdata` is always the source selected by `LoadLocation`.
7. No financial `BusinessDate`, SQL `DATE`, dividend/trade/settlement/snapshot economic-date rule, or
   UTC `SystemTimestamp` behavior changes.
8. The runtime change does not claim to close P3-04 general Unicode/maxLength semantics merely because
   the timezone field retains its current bound.

## 9. Alternatives rejected

| Alternative | Rejection rationale |
| --- | --- |
| Keep nonblank/length-only validation | Continues allowing values that violate the published IANA contract. |
| Trim or normalize timezone strings before validation | Hides caller mistakes and changes the exact persisted preference identity. |
| Hand-maintain an IANA regex/list | Duplicates timezone-database authority and will drift. |
| Reject all fixed-offset zones | Incorrectly rejects valid loadable identifiers such as `Etc/GMT+4`. |
| Treat `_ "time/tzdata"` as exclusive authority | Does not match documented `time.LoadLocation` source precedence. |
| Add PostgreSQL timezone-membership enforcement | Adds environment/version coupling without a demonstrated need; application admission is sufficient for this finding. |
| Backfill historic invalid values automatically | Would invent user preference data. |
| Use user timezone in financial-date math | Violates the canonical BusinessDate contract and creates a financial regression. |

## 10. Explicit non-goals

Stage 3.37 P3-02 does not change:

- `BusinessDate` semantics or SQL `DATE` columns;
- UTC `SystemTimestamp` semantics;
- transaction, dividend, settlement, snapshot, or tax-year economic dates;
- Decimal arithmetic or P3-03 closure;
- P3-04 Unicode/maxLength semantics;
- P3-05 idempotency/session cleanup lifecycle;
- P3-06 `httpapi/api.go` decomposition;
- P3-07 transaction fixture/default cleanup;
- P3-08 migration validator policy;
- P3-09 Next.js maintenance;
- P3-10 Fiber maintenance; or
- Stage 3.25 privacy evidence planning.

## 11. Governance bookkeeping carried by the planning PR

The planning PR may also synchronize the canonical current-state registries with already-established
repository truth from PR #89:

- PR #89 is merged;
- canonical merge SHA is `9c83b68e28bbb8bc971620d3e00be5e177ce0820`;
- P3-03 is canonically CLOSED;
- the original audit backlog is now P0=0 / P1=0 / P2=0 / P3=8;
- the remaining findings are P3-02, P3-04, P3-05, P3-06, P3-07, P3-08, P3-09, and P3-10.

This synchronization is factual bookkeeping only. It must not be represented as a second Stage 3.36
remediation or change the status/scope of any other finding.

## 12. Adversarial review requirements

The implementation reviewer must challenge:

- accidental acceptance of `Local` through the Go special case;
- trimming or normalization before resolver admission;
- a second custom timezone regex/list that can drift from `LoadLocation`;
- false claims that `time/tzdata` is the exclusive selected source;
- rejection of valid `Etc/GMT*` names merely because they are fixed-offset zones;
- environment-specific `ZONEINFO` replacement and whether deployment governance still makes the
  accepted contract operationally reasonable;
- another timezone write path that bypasses the auth validator;
- any use of user timezone in financial BusinessDate calculations;
- accidental absorption of P3-04 or any other remaining P3 finding; and
- any migration/backfill that guesses user preference data.

## 13. Operational and deployment consequences

The planned runtime change requires no new service, worker, database object, migration, provider, or
third-party package. Importing standard-library `time/tzdata` increases the Go binary by the timezone
database payload but makes tzdb availability independent of a minimal host image lacking system zoneinfo.
Higher-precedence `ZONEINFO`/system/GOROOT timezone data can still be selected by `time.LoadLocation` and
must be treated as governed deployment configuration.

No existing persisted preference is rewritten by the plan or runtime remediation.

## 14. Exact planning-review history

The first independent planning review returned `REQUEST CHANGES` because the proposed authority model
incorrectly described embedded `time/tzdata` as exclusive and used wording that could prohibit valid
loadable `Etc/GMT*` identifiers.

The revised plan corrected both points: `time.LoadLocation` semantics are authoritative, `time/tzdata`
is fallback availability, `Local` is explicitly rejected, raw offset spellings are distinguished from
loadable timezone identifiers, and valid `Etc/GMT*` names remain admissible. The repeat independent
planning review returned `APPROVED` with no P0/P1/P2/P3 planning findings.

## 15. Closure rule

This document authorizes no runtime change by itself. P3-02 remains OPEN until all of the following
occur:

1. this planning candidate passes exact-head CI and independent review evidence;
2. the user explicitly authorizes its squash merge and the planning PR is canonically merged;
3. a separately authorized runtime implementation is created from that canonical planning baseline;
4. the runtime candidate passes focused tests, full exact-head CI, and fresh independent implementation
   review after every material change;
5. the user separately authorizes the runtime squash merge; and
6. separately reviewed closure governance records the exact runtime head, CI, review, merge, and
   remaining audit state.

Until then P3-02 remains OPEN. Stage 3.25 privacy evidence planning and every other remaining P3 finding
remain separate.
