# GitHub configuration and verification

## Existing configuration

- [Normal CI](workflows/ci.yml) defines ten jobs: Go tests, Python tests, frontend build/typecheck, OpenAPI, Docker Compose, PostgreSQL migrations, Go vet, Go race, Go vulnerability scan and dependency security scan.
- CI runs for pull requests targeting `develop`/`main`, on the scheduled default branch, or through `workflow_dispatch`; a merge alone is not a push-triggered run.
- [Pull request template](PULL_REQUEST_TEMPLATE.md) and [all committed workflows](workflows/) remain the source for contribution/verification behavior; historical specialized workflows are not blanket permission to start campaigns.

## Exact-head evidence

For protected-branch post-merge verification, use Actions → CI → Run workflow → the intended branch. Confirm the run's event, branch and `head_sha`; a previous pull-request run is not evidence for a new merge SHA. Check every required job and security-scan result rather than only the aggregate badge.

The normal schedule is `17 3 * * *` (03:17 UTC / 07:17 Asia/Baku) on the repository default branch. GitHub can delay scheduled execution; verify the actual SHA and completed result, not the nominal time. Running workflows does not authorize merging, deploying or changing branch protection.

## Documentation

- [Project onboarding](../README.md) and [documentation index](../docs/DOCUMENT_INDEX.md).
- [Module #1 technical record](../docs/audit/MODULE1_TECHNICAL_CLOSURE.md), including exact post-merge CI evidence.
- [Documentation standard](../docs/governance/DOCUMENTATION_STANDARD.md): ten lines for standalone explanatory summaries; navigation, onboarding and full technical contracts retain necessary detail.
