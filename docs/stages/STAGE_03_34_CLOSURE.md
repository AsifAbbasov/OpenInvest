# Stage 3.34 — GitHub Governance and CI/Security Hardening Closure

| Field | Value |
| --- | --- |
| Status | Closure candidate; exact-head CI, independent governance review, and explicit human squash-merge authorization pending |
| Canonical baseline | `develop` at `c686a6721df51063ccf62a0303bb759d2215d60e` |
| Planning PR | #71 |
| Planning merge | `b4299bcdc28202c27388642dc7b426b159bb315c` |
| Implementation PR | #80 |
| Reviewed implementation head | `fd3a72a159161ec0bdf8018fdbf6e0a3da361885` |
| Implementation merge | `c686a6721df51063ccf62a0303bb759d2215d60e` |
| Findings | P2-16, P2-17 |
| Scope | Documentation/governance closure only; no runtime, migration, OpenAPI, dependency, architecture, product, or privacy-lifecycle changes |

This record is the final Stage 3.34 closure candidate. It supplements the implementation dossier in
`STAGE_03_34_CI_SECURITY_HARDENING.md`. Where that pre-merge dossier still says that P2-16/P2-17 are
OPEN or that final implementation evidence is pending, this closure record supplies the later evidence
and supersedes those status statements once this closure PR is canonical on `develop`.

## P2-16 — GitHub governance enforcement

### 1. Finding / symptom

The repository did not mechanically enforce the agreed delivery workflow on `develop`. A privileged
repository owner could rely on process discipline instead of GitHub-enforced controls for pull-request
entry, required checks, conversation resolution, linear history, force-push/deletion prevention, and
admin bypass containment.

### 2. Root cause

At the time of the original audit the repository was private on a GitHub Free account. GitHub exposed
branch-protection configuration UI but explicitly warned that the rules would not be enforced for that
private repository on the current plan. Repository merge methods also initially allowed merge commits
and rebases in addition to squash merges.

### 3. Failure scenario

Without mechanical enforcement, an administrator/owner could accidentally or intentionally bypass the
review path, merge before mandatory CI completed, directly alter protected history, force-push or delete
the default branch, or create non-linear history inconsistent with the canonical squash-only workflow.
A policy document alone would detect none of those actions before they affected the authoritative branch.

### 4. Impact

The defect was repository-wide governance debt. It weakened the credibility of every remediation claim
because code and documentation could reach the canonical branch without the same controls used to
produce the audit evidence.

### 5. Severity rationale

P2 is retained because this is a systemic prevention/control weakness rather than a single local code
defect. It does not itself change financial calculations or expose customer data, but it can undermine
the integrity of any change that does.

### 6. Guarantees violated

The accepted workflow requires pull-request delivery, required CI, resolved review conversations,
squash-only integration, protected branch history, and an explicit human merge gate. Stage 3.34 planning
also required normal administrator/owner bypass to be disabled and required the finding to remain OPEN
if that could not be mechanically enforced.

### 7. Considered solutions

1. Keep the repository private and create non-enforced/decorative branch rules.
2. Keep the repository private and rely on documentation/manual discipline.
3. Purchase or move to an account tier that enforces private-repository protections.
4. Make the repository public after a full-history public-release audit, then use GitHub Free enforced
   classic branch protection.
5. Add detection-only CI/hooks without branch protection.

### 8. Chosen remediation

The repository was first prepared for safe public visibility with a disposable full-history secret scan.
The audit checked 318 commits across fetched repository refs with Gitleaks v8.30.1 using a pinned image
digest. Nine findings were reviewed and classified as synthetic/test/example values; no real production
credential, API key, password, or private key was identified. One non-noreply historical Git email was
identified and explicitly accepted by the repository owner as public metadata.

The repository was then changed from private to public. GitHub Free enforcement became available and the
`develop` classic branch-protection rule was configured to require:

- a pull request before merging;
- all ten final Stage 3.34 CI checks;
- conversation resolution before merging;
- linear history;
- protections to apply to administrators / no normal bypass;
- force pushes disabled; and
- branch deletion disabled.

Repository merge policy was separately reduced to squash-only: squash merge enabled; merge commits and
rebase merge disabled.

### 9. Why this solution was chosen

The user explicitly had no budget for a paid GitHub plan and explicitly approved making OpenInvest public
after the public-release audit. A public repository provides enforceable GitHub Free branch protection
without weakening the accepted control set or pretending that decorative private rules are sufficient.

### 10. Rejected alternatives

**Decorative private branch protection** was rejected because GitHub explicitly stated it would not be
enforced and therefore could not close P2-16.

**Manual-only governance** was rejected because the finding requires prevention, not only policy.

**Paid Team/Enterprise migration** was not selected because it was unnecessary once the user approved the
public-repository route and had no budget for a paid plan.

**Detection-only Actions/hooks** were rejected as substitutes because they can report a violation only
after an unauthorized history change has happened.

### 11. Trade-offs

The source code and Git history are now publicly readable. The owner explicitly accepted exposure of the
historical non-noreply Git email. Public visibility also increases external scrutiny and forkability.
Those costs are accepted in exchange for enforceable no-cost repository governance.

### 12. Regression / verification coverage

Post-save GitHub API evidence reports:

- repository visibility `public`;
- default branch `develop`;
- `develop` `protected=true` with protection enabled;
- required status-check enforcement level `everyone`;
- exactly ten required GitHub Actions contexts:
  - `Go tests`
  - `Python tests`
  - `Frontend build and typecheck`
  - `OpenAPI contract`
  - `Docker Compose config`
  - `PostgreSQL migration validation`
  - `Go vet`
  - `Go race tests`
  - `Go vulnerability scan`
  - `Dependency security scan`;
- squash merge enabled;
- merge commits disabled; and
- rebase merge disabled.

The saved branch-protection form records pull-request entry, conversation resolution, linear history,
no normal administrator bypass, force-push disabled, and deletion disabled. Destructive direct-push or
force-push probes were intentionally not used because a failed control would itself mutate canonical
history.

### 13. Adversarial review findings

Stage 3.34 planning initially received `REQUEST CHANGES` because the first plan allowed administrator /
owner bypass to remain merely disclosed. The plan was corrected to require mechanical enforcement and to
fail closed if the account could not provide it. Repeat independent planning review returned `APPROVED`.

The implementation reviewer later approved P2-17 but explicitly required P2-16 to remain OPEN until
repository settings were mechanically enforced. P2-16 is therefore closed only by the later public
visibility + enforced branch-protection evidence, not by the CI implementation merge itself.

### 14. Remediation iterations

1. Merge methods were reduced to squash-only and API-verified.
2. Rulesets were inspected while the repository was private; GitHub warned they would not be enforced.
3. Classic branch protection was inspected while private; GitHub gave the same enforcement limitation.
4. A disposable full-history public-release audit PR #81 was created, executed, and closed without merge.
5. After the owner explicitly accepted public visibility and historical email exposure, the repository
   was changed to public.
6. Classic protection was created for `develop`, then edited to add the final control set and all ten
   stable CI check names.
7. Post-save repository/branch API evidence confirmed public visibility, protected `develop`,
   `enforcement_level=everyone`, and all ten required check contexts.

### 15. Residual risk

Classic branch protection is a GitHub-hosted control and remains subject to GitHub platform semantics and
future configuration changes. Repository administrators can still edit repository settings through the
GitHub control plane; such administrative reconfiguration is outside the branch rule itself and must be
treated as a governed change. Public visibility is irreversible in the sense that already-cloned public
history cannot be made private again by later changing repository visibility.

### 16. Operational / deployment consequences

No application runtime or deployment behavior changed. Repository contribution/merge operations now have
stronger admission rules, and all future pull requests to `develop` must satisfy the configured required
checks before merge.

### 17. Exact evidence

- Stage 3.34 planning PR #71, approved planning head
  `0583fda8da92cbc15efd7e0497bd36027956c87e`, CI #205 SUCCESS, independent planning `APPROVED`,
  squash merge `b4299bcdc28202c27388642dc7b426b159bb315c`.
- Repository merge-policy API after configuration: `allow_squash_merge=true`,
  `allow_merge_commit=false`, `allow_rebase_merge=false`.
- Disposable public-release audit PR #81: closed, not merged; full-history scan inspected 318 commits.
- Repository API after visibility change: `visibility=public`, default branch `develop`.
- Branch API after final save: `develop` protected, protection enabled, required status checks enforced
  for `everyone`, with the exact ten final GitHub Actions contexts listed above.
- Canonical `develop` remained at `c686a6721df51063ccf62a0303bb759d2215d60e` while the repository
  settings were changed; no governance-setting operation rewrote application history.

### 18. Final canonical status

P2-16 is **CLOSED CANDIDATE**. It becomes canonically **CLOSED** when this closure PR passes exact-head CI,
independent governance review returns `APPROVED`, explicit human squash-merge authorization is given, and
the closure PR is squash-merged into `develop`.

## P2-17 — final closure evidence

The full 18-part engineering dossier remains
`STAGE_03_34_CI_SECURITY_HARDENING.md`. Final evidence that was pending in that implementation-time record
is now fixed as follows:

- final implementation PR: #80;
- frozen independently reviewed head: `fd3a72a159161ec0bdf8018fdbf6e0a3da361885`;
- exact-head CI #230 / run `32671862989`: **SUCCESS, 10/10 jobs**;
- secondary exact-head CI #229: **SUCCESS, 10/10 jobs**;
- independent implementation review: **APPROVED**;
- reviewer disposition: P2-17 CLOSED CANDIDATE, no new blocking P1/P2 regression, P2-16 explicitly left
  OPEN until repository governance enforcement;
- explicit human squash-merge authorization was received before merge; and
- PR #80 squash merge: `c686a6721df51063ccf62a0303bb759d2215d60e`.

Therefore P2-17 is **CLOSED CANDIDATE** for formal Stage 3.34 closure. This record does not silently close
P3-09 or P3-10: Next.js and Fiber maintenance remain separately governed P3 findings.

## Stage 3.34 closure semantics

When this closure record is squash-merged into `develop` after exact-head green CI, independent
`APPROVED` governance review, and explicit human authorization:

- P2-16 becomes canonically **CLOSED**;
- P2-17 becomes canonically **CLOSED**;
- the original 32-finding audit has no remaining P0, P1, or P2 findings;
- exactly 10 P3 findings remain:
  - P3-01 password byte semantics;
  - P3-02 true IANA timezone;
  - P3-03 OpenAPI Decimal grammar;
  - P3-04 Unicode/maxLength semantics;
  - P3-05 idempotency/session cleanup;
  - P3-06 split `httpapi/api.go`;
  - P3-07 transaction form fixture defaults;
  - P3-08 migration validator weaker than policy;
  - P3-09 Next.js maintenance; and
  - P3-10 Fiber maintenance.

Stage 3.25 privacy Security Review evidence planning remains separate and unchanged.

No P3 implementation begins as part of this closure PR. The next audit-remediation implementation scope
must be separately reviewed and must preserve the repository's mandatory human merge gate.
