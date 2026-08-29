# Stage 3.40 — P3-09 Next.js Security Maintenance Plan

| Field | Value |
| --- | --- |
| Status | Planning/review candidate only; dependency/runtime implementation not authorized |
| Date | 2026-08-29 |
| Canonical planning base | `develop@41e35b672d166cf74c3f0c3ee248330193ae51c1` |
| Finding | Original audit `P3-09 — Next.js maintenance` |
| Prior closure | P3-04 CLOSED through PR #99 squash merge `41e35b672d166cf74c3f0c3ee248330193ae51c1` |
| Current audit state | 27 / 32 closed = 84.375%; remaining P3-06, P3-07, P3-08, P3-09, P3-10 |
| Runtime/dependency implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / Ready / merge authorized here | No |

## 1. Objective

Close the maintenance gap represented by original audit finding P3-09 by moving the OpenInvest
Next.js dependency from the currently affected `16.3.2` release to the official patched stable
Active-LTS release `16.3.3`, while preserving current application behavior and avoiding unrelated
framework, React, architecture, feature, or deployment changes.

This planning artifact does not itself modify dependencies and does not close P3-09.

## 2. Current repository baseline

At the planning base:

- `frontend-next/package.json` pins `next` to `16.3.2`;
- `react` is `19.2.7`;
- `react-dom` is `19.2.7`;
- package manager is `pnpm@11.8.0`;
- Node engine is `>=22.22.2`;
- `next.config.ts` contains only:
  - `poweredByHeader: false`;
  - `reactStrictMode: true`;
  - Turbopack root configuration;
- repository search found no `cacheComponents` usage;
- repository search found no `next/image` import.

No evidence in the current repository establishes that OpenInvest is hosted on Windows or that it
actively optimizes attacker-controlled AVIF images. Therefore exploitability of each upstream issue
must not be overstated. Nevertheless, the installed Next.js version is inside the upstream affected
version ranges and is below the patched stable release.

## 3. Upstream security basis

On 2026-08-25 the official Next.js security release instructed users on the 16.x Active-LTS line to
upgrade to `16.3.3` to address two Critical vulnerabilities.

### GHSA-p293-qw3h-jr36

Unauthenticated remote code execution on Windows-hosted Next.js servers.

Affected:
- `>=16.0 <16.3.3`

Patched:
- `16.3.3`

OpenInvest currently pins `16.3.2`, so dependency-version applicability is confirmed.
OpenInvest-specific runtime exploitability is not claimed because the repository does not establish a
Windows production host.

### GHSA-2xp9-vwfh-vxw4

Unauthenticated remote code execution in the Image Optimization API when AVIF files are used.

Affected:
- `<16.3.3` on the current 16.x line

Patched:
- `16.3.3`

OpenInvest currently pins `16.3.2`, so dependency-version applicability is confirmed.
Repository search found no `next/image` import and current `next.config.ts` does not opt into an AVIF
image format, so an active OpenInvest exploit path is not asserted from repository evidence alone.

## 4. Severity and priority

The original audit severity remains **P3**. This planning stage does not retroactively reclassify the
audit finding.

Operational priority is elevated because the maintained upstream line now contains two Critical
security fixes that are absent from the repository's pinned `16.3.2`.

If implementation or security review finds a demonstrated OpenInvest-reachable exploit path, any
severity reclassification must be explicit and separately evidenced rather than silently rewriting the
original audit severity.

## 5. Frozen target

The approved implementation target SHALL be:

`next: 16.3.2 -> 16.3.3`

Rules:

1. Use the stable `16.3.3` release, not a canary/preview build.
2. Keep `react` at `19.2.7`.
3. Keep `react-dom` at `19.2.7`.
4. Keep current `@types/react`, `@types/react-dom`, TypeScript, pnpm, and Node policy unless the exact
   Next.js update proves an unavoidable compatibility requirement.
5. If a React/ReactDOM/type/toolchain bump becomes necessary, stop implementation and return to review
   before expanding scope.
6. Do not combine P3-09 with Fiber P3-10 or `httpapi/api.go` decomposition P3-06.

## 6. Expected implementation surface

Expected mandatory dependency files:

- `frontend-next/package.json`
- `frontend-next/pnpm-lock.yaml`

Expected stage evidence/report file may be added under `docs/stages/`.

Application source changes are **not expected**.

If `16.3.3` requires an application-code compatibility change, it must be:
- minimal;
- directly attributable to the version bump;
- behavior-preserving;
- covered by targeted regression tests;
- explicitly called out in the implementation review.

No OpenAPI, Go, SQL, migration, financial, privacy, authentication protocol, or infrastructure change
is authorized.

## 7. Required lockfile invariants

Implementation must prove:

- direct `next` dependency resolves to exactly `16.3.3`;
- no active `next@16.3.2` resolution remains in `frontend-next/pnpm-lock.yaml`;
- lockfile changes are reproducible from the declared package manifest;
- no unrelated dependency drift is accepted merely because the lockfile was regenerated;
- transitive changes caused by the exact Next.js patch update are enumerated in the stage evidence.

## 8. Required local verification

At minimum:

```text
cd frontend-next
pnpm install --frozen-lockfile
pnpm test
pnpm typecheck
pnpm build
```

Dependency/security verification must also include the repository's existing dependency security scan
or its exact local equivalent where available.

The implementation evidence must record exact command outcomes rather than a generic "tests passed".

## 9. Required behavioral regression coverage

The patch update must preserve existing OpenInvest Web behavior, particularly:

- application routes render/build exactly under the existing App Router structure;
- browser-to-Go API ownership remains unchanged;
- auth/session client behavior is unchanged;
- cookies/headers and request credentials behavior are unchanged;
- import UI request construction remains unchanged;
- portfolio/transaction UI request construction remains unchanged;
- asset search/navigation remains unchanged;
- no server-side financial/business logic is introduced into Next.js;
- no new external data call from the browser is introduced.

Existing tests are authoritative where they cover these surfaces. A source change required by the
framework update must receive targeted regression coverage.

## 10. Upstream 16.3.x regression note

A public upstream report filed after `16.3.3` describes a memory-retention problem on 16.3.x when
`cacheComponents` / `use cache` are involved.

Current OpenInvest repository search found no `cacheComponents` usage and current `next.config.ts`
does not enable it.

Therefore:

- this report is recorded as an upstream watch item;
- it is not evidence that OpenInvest currently has the reported leak;
- it does not justify moving OpenInvest to an unstable 16.4 canary;
- implementation review must fail closed if new evidence shows OpenInvest actually enables the affected
  mode.

## 11. Security acceptance evidence

Before implementation can receive final approval, evidence must show:

1. `next` resolves to patched `16.3.3`;
2. the two 2026-08-25 Critical advisories no longer match the installed Next.js version;
3. required frontend tests/typecheck/build pass;
4. dependency security scan passes;
5. exact-head GitHub CI passes all required jobs after publication;
6. no unrelated dependency/framework drift occurred;
7. no new material security finding remains unresolved.

## 12. Rollback policy

A normal rollback from `16.3.3` back to affected `16.3.2` is **not** an acceptable long-lived security
remediation.

If `16.3.3` causes a blocking regression:

- do not merge;
- remediate forward on the feature branch if scope remains narrow; or
- stop and return to planning if the required change materially expands scope.

After protected merge, an emergency revert may be used only as an explicit incident action with the
security exposure acknowledged; the preferred remediation remains a supported patched release.

## 13. Explicit out-of-scope

This stage does not close or modify:

- P3-06 — `httpapi/api.go` decomposition;
- P3-07 — transaction-form fixture/default semantics;
- P3-08 — migration-validator policy hardening;
- P3-10 — Fiber maintenance;
- Stage 3.25 privacy work;
- React major/minor maintenance;
- Node or pnpm maintenance;
- new Next.js features;
- Cache Components adoption;
- image-pipeline redesign;
- routing redesign;
- API proxy redesign;
- Web architecture amendment changes;
- Go/API behavior;
- database/schema/migrations;
- financial calculations.

## 14. Implementation review gates

After this plan is accepted, implementation remains on the mandatory **development path**:

1. feature branch from the exact then-current protected `develop`;
2. exact scoped dependency change and evidence;
3. local quality gates;
4. Internal line-by-line review in the designated review chat;
5. remediation of all findings;
6. explicit human commit/push authorization;
7. Draft PR to `develop`;
8. exact-head required CI;
9. External published-head review in the same designated review chat;
10. evidence publication/verification required by the effective workflow;
11. explicit human squash-merge authorization;
12. separate closure-governance activation.

No review verdict by itself authorizes a protected action.

## 15. Closure semantics

P3-09 remains OPEN throughout planning and implementation.

It becomes CLOSED only after:
- the approved implementation is actually merged into protected `develop`; and
- the required closure-governance record/canonical audit surfaces are subsequently activated according
  to the effective repository workflow.

After P3-09 closure, if no other audit finding changes concurrently:

- closed: 28 / 32;
- completion: 87.5%;
- remaining: 4 / 32;
- remaining findings: P3-06, P3-07, P3-08, P3-10.
