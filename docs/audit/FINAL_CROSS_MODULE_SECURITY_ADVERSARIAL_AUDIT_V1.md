# Final Cross-Module Security Adversarial Audit V1

## Scope and immutable baseline

- Audit module: **Module #11 — Final Cross-Module Security Audit**.
- Detection/verification only; no product remediation is included.
- Protected baseline: `develop@cf5d1b0f151699d8498b4215df79fdd8e7a3cca2`.
- Baseline tree: `34c59ffa46496ae197b8a43eb888c23ae47cdeea`.
- Audit branch: `audit/final-cross-module-security-v1`.
- Product/runtime files changed by this audit: **none**.
- Audit-only additions: dedicated workflow, Go integration/property/fuzz tests, frontend contract test, and this evidence report.

## Audit harness chronology

| Commit / run | Result | Meaning |
| --- | --- | --- |
| `1b773716ca21a38332fb518f193bb98d1264562d` / run `37801022520` | FAILURE | Initial harness exposed two test defects: replay typed-result expectation was wrong and valuation cleanup omitted an FK-owned row. Product behavior was not established vulnerable. |
| `ba5924a8c7e516012fb795bbd044414f342d77d9` / run `37801402232` | FAILURE | Harness corrections introduced a compile-only defect (`declared and not used: first`), affecting PostgreSQL groups before execution. |
| `c92f254ae1eda91ef7ff7a6fa3f4e4c88694bb15` / run `37801655030` | SUCCESS | Corrected six-group adversarial harness completed successfully. |

Harness defects are classified separately as TD-M11-01..03 and are not product vulnerabilities.

## Executed attack campaigns

### A — authentication / ownership / idempotency

PostgreSQL-backed tests exercised foreign-principal reads across portfolio, transaction list, positions, summary, cash-flow and returns; same-key replay under the owner; same key under a foreign principal; changed-body same-key conflict; same key on a different portfolio path; and residual replay-row checks.

Observed:
- `M11_CM01_FOREIGN_SUBJECT_READ_MATRIX=PASS`
- `M11_CM02_NO_CROSS_PRINCIPAL_MUTATION=PASS`
- `M11_CM03_IDEMPOTENCY_SCOPE_AND_EXACT_REPLAY=PASS`
- Completed replay returns the persisted replay artifact as designed.
- Refresh rotation and logout revoke refresh-session state, but an already issued short-lived access JWT remains accepted until its configured 15-minute expiry.

The last item is **HARDENING_ONLY M11-H01**, not a demonstrated contract bypass: the current architecture explicitly uses short-lived signed access tokens and does not claim per-request server-side access-token revocation.

### B — ledger / import / position / WAC / cash-flow / XIRR / concurrency

Real PostgreSQL transactions and barriers exercised:
- import vs SELL;
- SELL vs SELL;
- correction vs reversal;
- correction-history vs clean-reference rebuild;
- XIRR read concurrent with correction.

Observed:
- `M11_IMPORT_VS_SELL_SERIALIZATION=PASS`
- `M11_SELL_VS_SELL_SERIALIZATION=PASS`
- `M11_CORRECTION_VS_REVERSAL_SERIALIZATION=PASS`
- `M11_CM04_EFFECTIVE_LEDGER_SINGLE_SOURCE=PASS`
- `M11_CM05_POSITION_WAC_REBUILD_EQUIVALENCE=PASS`
- `M11_CM06_CASH_FLOW_XIRR_REBUILD_EQUIVALENCE=PASS`
- `M11_XIRR_READ_VS_MUTATION_COMMITTED_STATES_ONLY=PASS`

The reference-history case included corrected DEPOSIT/WITHDRAWAL dates, reversed DIVIDEND, corrected/backdated BUY, SELL and reversed ephemeral BUY. Position quantity/WAC/acquisition basis, cash totals, XIRR inputs/result and summary were compared to a clean effective-history reference.

### C — cancellation / database transaction / runtime role

A portfolio row lock forced cancellation of an idempotent financial append before commit. Evidence verified zero ledger rows and zero replay reservations after cancellation, followed by successful retry with the same key.

Observed:
- `M11_CM07_CANCELLATION_BEFORE_MUTATION_COMMIT=PASS`
- `M11_RETRY_AFTER_CANCELLATION=PASS`

Existing runtime-role integration attacks were executed under a real restricted PostgreSQL login and passed:
- append through permitted application path;
- reject privileged session masked by SET ROLE;
- reject latent SET ROLE mutation path;
- reject latent ADMIN OPTION mutation path.

No privilege escalation was reproduced.

### D — HTTP parser / rate-limit / provider-budget composition

Executed:
- unknown JSON fields;
- trailing JSON;
- duplicate JSON key observation;
- invalid date payload flood followed by valid expensive requests;
- duplicate query parameter request;
- uniform mapped 404 responses;
- canonical provider request-key fuzzing.

Observed:
- unknown fields rejected;
- trailing JSON rejected;
- duplicate JSON keys decode with last-value-wins semantics, but no cross-module security/correctness bypass was demonstrated;
- invalid payloads reached the provider **0** times and did not consume the per-client expensive-provider allowance;
- valid requests were admitted up to 12/60s and then returned 429 with `Retry-After: 60`;
- one duplicate-query HTTP request caused at most one provider call;
- mapped not-found responses were uniform and did not leak SQL/PostgreSQL details;
- `M11_CM08_INVALID_REQUESTS_DO_NOT_AMPLIFY_PROVIDER_BUDGET=PASS`;
- `M11_CM12_VALIDATION_RATE_LIMIT_PROVIDER_ORDERING=PASS`;
- `M11_CM09_NOT_FOUND_ORACLE_UNIFORM=PASS`;
- `M11_CM10_INTERNAL_ERROR_DETAIL_LEAK=NONE`.

The inherited Module #10 duplicate-JSON-key state remains hardening-only/non-blocking; this audit found no new impact.

### E — frontend / backend security contract

Node contract tests verified:
- public Corporate Actions request uses `credentials: omit`, no Authorization and no CSRF header;
- frontend performs no automatic retry on 429;
- financial mutation carries Bearer authorization and Idempotency-Key and does not cross over to refresh-cookie/CSRF semantics.

Observation:
- frontend result objects currently do **not** expose backend `Retry-After` to callers. This is **HARDENING_ONLY M11-H02** because the frontend also performs no automatic retry and therefore does not bypass or amplify the backend budget control.

### F — stateful / property / fuzz

Stateful position campaign:
- deterministic seed `110011`;
- **1,000 sequences**;
- **50 operations per sequence**;
- **50,000 operations total**;
- incremental state compared to full rebuild after every operation;
- result: `M11_CM05_INCREMENTAL_REBUILD_EQUIVALENCE=PASS`.

Six fuzz targets executed successfully:
1. `FuzzM11StrictJSONDecoder`
2. `FuzzM11ProviderCanonicalRequestKey`
3. `FuzzM11CSVReviewSemantics`
4. `FuzzM11PositionPipeline`
5. `FuzzM11CommandIdentity`
6. `FuzzM11DecimalCashXIRRPipeline`

Recorded wrapper durations on successful run: **45s + 5s + 7s + 7s + 7s + 5s = 76s**. Wrapper duration includes compile/corpus setup. Representative fuzz execution counts included 2,083 strict-JSON executions, 29,059 provider-key executions, 105,864 CSV executions, 73,878 position executions, 13,416 command-identity executions and 60,014 decimal/XIRR executions.

## Finding classification

### P0/P1/P2/P3 product findings

No P0, P1, P2 or P3 product vulnerability was demonstrated by the corrected Module #11 harness.

### HARDENING_ONLY

**M11-H01 — issued access JWT remains valid until TTL after refresh/logout**

- Reproduced on corrected PostgreSQL/auth integration harness.
- Access token contract is short-lived stateless bearer (15-minute default).
- Refresh-session revocation does not invalidate already-issued bearer tokens.
- A bearer captured before logout can continue authorizing mutations until expiry.
- This is a residual security-hardening tradeoff, not a bypass of the documented current token model.

**M11-H02 — frontend does not expose Retry-After**

- Backend 429 enforcement remains effective.
- Frontend performs no automatic retry.
- Missing propagation can reduce caller UX/backoff visibility but did not amplify provider traffic in this audit.

### Test defects

- **TD-M11-01:** initial test incorrectly expected a typed transaction value on completed idempotent replay instead of the persisted replay artifact.
- **TD-M11-02:** initial audit cleanup omitted manual valuation rows, causing an FK cleanup failure after otherwise successful equivalence assertions.
- **TD-M11-03:** intermediate harness correction left an unused variable and failed compilation before PostgreSQL groups could execute.
- All three were corrected on the audit branch; corrected run `37801655030` passed.

## Non-verified / partially verified boundaries

- **NV-M11-01:** full cross-session CSRF attack matrix at the actual HTTP cookie boundary was not executed end-to-end in this campaign.
- **NV-M11-02:** authorization/mutation matrix was not exhaustively repeated for every CSV/correction/reversal/Corporate-Actions HTTP route combination.
- **NV-M11-03:** full cancellation phase matrix was not reproduced at every requested phase (after tx start, during query, after mutation-before-commit, provider request, response serialization). One blocked-before-commit DB case is verified.
- **NV-M11-04:** parser/routing matrix is incomplete for percent-encoding, path normalization, duplicate headers, mixed-case headers and all Content-Type variants.
- **NV-M11-05:** concurrency matrix is incomplete for BUY-vs-SELL, correction-vs-correction, reversal-vs-reversal, snapshot/rebuild-vs-mutation and provider-projection-vs-cancellation.
- **NV-M11-06:** CSV/import campaign did not compare every accepted adversarial import sequence to a clean full-system reference after every operation.
- **NV-M11-07:** 1,000×50 stateful coverage applies to the position BUY/SELL engine; equally deep generated correction/reversal/cash/XIRR full-system histories were not executed.
- **NV-M11-08:** dedicated route-UUID+authorization and correction/reversal-state-machine fuzz targets were assessed as useful but were not executed; the six executed targets cover adjacent boundaries only.
- **NV-M11-09:** horizontally shared production provider-budget enforcement remains NOT VERIFIED and belongs to Module #12; repository-local composition tests do not prove distributed enforcement.

## Cross-module invariant table

| Invariant | Result | Evidence |
| --- | --- | --- |
| CM-01 NO_CROSS_PRINCIPAL_DATA_ACCESS | PASS | Foreign subject matrix across portfolio/transactions/positions/summary/cash-flow/returns returned not-found. |
| CM-02 NO_CROSS_PRINCIPAL_MUTATION | PARTIALLY_VERIFIED | Foreign principal same-key ledger mutation rejected with no replay residue; not every mutation route was repeated. |
| CM-03 EXACTLY_ONCE_BUSINESS_EFFECT_UNDER_RETRY | PASS | Exact replay artifact, changed-body conflict, scoped same-key behavior, cancellation retry recovery. |
| CM-04 EFFECTIVE_LEDGER_SINGLE_SOURCE_OF_TRUTH | PASS | Mutated history matched clean reference projections. |
| CM-05 POSITION_WAC_REBUILD_EQUIVALENCE | PASS | PostgreSQL reference case plus 1,000×50 stateful position campaign. |
| CM-06 CASH_FLOW_XIRR_REBUILD_EQUIVALENCE | PASS | Corrected/reversed history matched clean reference cash flows, XIRR and summary. |
| CM-07 CANCELLATION_ATOMICITY | PARTIALLY_VERIFIED | Blocked pre-commit cancellation verified; exhaustive phase matrix not executed. |
| CM-08 RATE_LIMIT_PROVIDER_BUDGET_COMPOSITION | PARTIALLY_VERIFIED | Invalid-before-budget ordering, per-client limit and provider-call bound verified; distributed/shared budget remains Module #12. |
| CM-09 NO_AUTH_OR_EXISTENCE_ORACLE | PARTIALLY_VERIFIED | Mapped not-found semantics uniform; full timing/error-class matrix not executed. |
| CM-10 NO_SECRET_OR_INTERNAL_ERROR_LEAKAGE | PARTIALLY_VERIFIED | No SQL/Postgres detail in tested mapped errors; exhaustive provider/token/stack-trace matrix not executed. |
| CM-11 RUNTIME_DB_ROLE_FAIL_CLOSED | PASS | Restricted runtime-role attack suite passed; no escalation reproduced. |
| CM-12 SECURITY_CONTROL_ORDERING_SAFE | PARTIALLY_VERIFIED | Validation-before-budget and ownership/idempotency ordering verified; full HTTP parser/CSRF matrix incomplete. |
| CM-13 CONCURRENT_COMMAND_SERIALIZATION_SAFE | PARTIALLY_VERIFIED | Import-vs-SELL, SELL-vs-SELL, correction-vs-reversal and XIRR-read-vs-correction passed; full matrix incomplete. |
| CM-14 FRONTEND_BACKEND_SECURITY_CONTRACT_SAFE | PARTIALLY_VERIFIED | Auth/idempotency/cookie separation and no automatic 429 retry verified; Retry-After not surfaced and broader logout/refresh UI races not exhaustively exercised. |

## Limitations and residual boundary

This audit demonstrates strong repository-side composition behavior under the executed matrices. It does **not** prove horizontally shared production provider budgets, actual multi-instance deployment behavior, exhaustive real-socket failure behavior, or every requested HTTP/cancellation/concurrency permutation.

Module #12 was not started.

## Evidence summary

- Baseline SHA/tree: `cf5d1b0f151699d8498b4215df79fdd8e7a3cca2` / `34c59ffa46496ae197b8a43eb888c23ae47cdeea`.
- Corrected cross-module audit run: `37801655030` — SUCCESS.
- Stateful sequences: **1,000**.
- Stateful operations: **50,000**.
- Fuzz targets executed: **6**.
- Recorded fuzz wrapper duration: **76 seconds**.
- Product findings: P0=0, P1=0, P2=0, P3=0.
- Hardening-only observations: 2.
- Test defects: 3.
- Not-verified items: 9.
- Remediation authorized: **NO**.
- Module #11 formally closed: **NO**.
- Module #12 started: **NO**.
