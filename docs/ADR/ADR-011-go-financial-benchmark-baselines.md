# ADR-011 — Package-local Go performance baselines for financial computations

- **Status / owner / date:** PROPOSED documentation record; implementation already merged; Asif Abbasov; 2026-10-30. Supersedes: none.
- **Context:** WAC, XIRR, snapshots and ledger replay had financial regression, fuzz and race coverage, but no reproducible Go performance baseline.
- **Root cause / problem:** correctness tests did not quantify `ns/op`, `B/op` and `allocs/op`; allocation or latency drift could remain unmeasured.
- **Risk:** unmeasured performance cost or treating a synthetic benchmark as a production capacity guarantee.
- **Donor disposition:** concurrency/regression practices = CONFIRM EXISTING; finance benchmarks = ADAPT/IMPLEMENTED; OpenTelemetry tracing = FUTURE. No donor source code was copied.
- **Alternatives:** reuse donor benchmark code or introduce an external harness versus measure current code with built-in `testing.B`; choose the package-local standard-library option.
- **Decision:** use deterministic fixtures outside timed loops for WAC rebuild, XIRR cash-flow aggregation, bounded snapshot timeline and snapshot-state digest; report allocations and gate large workloads with `OPENINVEST_BENCH_LARGE=1`.
- **Rationale / trade-offs:** no new dependency, financial semantic change or CI timing threshold; synthetic inputs are less representative than end-to-end production workloads.
- **Bounds:** snapshot timeline uses DEPOSIT-only rows and the existing 5,000-row replay limit; XIRR has 28 distinct dates; 100k digest positions are a synthetic fingerprint stress case, not an allowed 100k ledger replay.
- **Implementation:** [PR #289](https://github.com/AsifAbbasov/OpenInvest/pull/289), three benchmark-only files, squash commit `6c8cbffdb4435ca46277aa81bffe39acfc495d56`; [PR #290](https://github.com/AsifAbbasov/OpenInvest/pull/290) separately updated security dependencies and Go 1.26.10.
- **Verification:** [exact-head CI #41047943799](https://github.com/AsifAbbasov/OpenInvest/actions/runs/41047943799) passed 10/10; Go 1.26.10 Linux/AMD EPYC produced 45/45 metric rows. Earlier Go 1.26.9/Apple M1 Pro results are historical and not a valid direct speed comparison.
- **Consequences / residuals:** compare future runs only on controlled, like-for-like environments; no optimization, PostgreSQL performance measurement, production-capacity claim or tracing implementation is established.
- **Authority:** documentation of an implemented engineering choice only; Documents 42–43, Architecture Freeze v1.2, financial code, databases and deployment gates remain unchanged.
