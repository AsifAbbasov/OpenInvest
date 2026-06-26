# Web Frontend Architecture Amendment — Next.js Presentation Layer

| Field | Value |
| --- | --- |
| Document ID | WEB-ARCH-001 |
| Version | 1.0.0 |
| Status | Implemented locally / Awaiting Review |
| Owner | Builder Engineer |
| Supersedes | Current Web implementation target only; see ADR-007 |
| Dependencies | Documents 42–43; ADR-003; ADR-005; accepted ADR-007 |
| Last Review Date | 2026-06-20 |
| Next Review Date | At architecture-amendment approval |

## Goal

Replace the non-business Vite skeleton with a Next.js App Router + TypeScript + pnpm Web skeleton
without weakening the Go API, OpenAPI First, security, privacy, financial, or MVP boundaries.

## Decision and publication order

The human architecture request accepted the Next.js presentation-only decision. No local edit-order
claim is used as audit evidence, and no commit is authorized by this report. If the amendment is
approved, repository history must preserve decision-before-implementation ordering with two commits
on the dedicated `feature/nextjs-web-presentation` branch:

1. ADR-007, Source of Truth, document registries, changelog, and historical-freeze supersession;
2. Vite skeleton removal, Next.js skeleton, CI, README, ignore rules, and implementation log.

The branch must not be pushed until both commits and the complete branch diff pass the mandatory
review workflow and the human explicitly approves push. No direct push to `develop` or `main` is
allowed.

Stage 2 is merged into `develop`, ADR-006 is accepted, and PR #4 is based on the updated
`develop` baseline. PR #4 targets `develop` and is isolated to the Next.js Web Presentation
Amendment. After rebasing onto the Stage 2 baseline, checks were rerun for the Web amendment
scope before requesting review.

## Implemented scope

- accepted ADR-007 with presentation-only responsibilities and prohibitions;
- Source of Truth, Document Index, Version Matrix, Changelog, and historical freeze annotation;
- `frontend-next/` with one `src/app` App Router root;
- one static architecture-status page and root metadata/layout;
- pnpm 11.8.0 lockfile and explicit sharp-only dependency build permission;
- CI frontend working directory changed to `frontend-next` with frozen install, typecheck, build,
  and disabled Next.js telemetry;
- Vite skeleton removed from tracked source;
- current README and implementation log updated.

## Explicitly excluded

- business/domain logic and business Route Handlers;
- Server Actions;
- direct PostgreSQL or Redis access;
- portfolio, dividend, financial, inflation, purchasing-power, or tax calculations;
- MOEX, CBR, Rosstat, broker, or other external-provider integration;
- repositories, SQL migrations, workers, generated clients, and OpenAPI changes;
- frontend business screens and LocalStorage business persistence;
- SwiftUI, Jetpack Compose, or any mobile implementation;
- Stage 3.

## Review-size exception

This amendment exceeds the default 25-file review count because one atomic, reversible Web-runtime
replacement deletes eleven tracked Vite files, introduces the corresponding Next.js source/config
and lockfile, and updates mandatory architecture registries. These are one responsibility and one
rollback outcome; no business behavior is bundled. The lockfile and mechanical legacy deletions are
reported separately from hand-written logic.

The auditable Principal Architect / Human Reviewer approval evidence for the PR #4 34-file
review-size exception is recorded as PR comment:
<https://github.com/AsifAbbasov/OpenInvest/pull/4#issuecomment-4810564738>.
This approval covers the review-size exception only. It is not merge approval, not Stage 3
approval, and not approval for future PR-size exceptions.

The previously requested stage-handoff `AGENTS.md`/workflow change is intentionally excluded from
this branch and must be restored in a separate governance change.

## Verification

- `go test ./...`: passed.
- `uv sync --extra dev --locked && uv run pytest`: passed; one known upstream warning.
- `pnpm install --frozen-lockfile`: passed with pnpm 11.8.0.
- `pnpm run typecheck`: passed with Next.js 16.2.9.
- `pnpm run build`: passed; only static `/` and `/_not-found` routes exist.
- `docker compose config --quiet`: passed.
- OpenAPI validator: passed, 22 operations, 2,501 resolved references, 11 documents.
- YAML, JSON, Markdown-link, forbidden-boundary, and whitespace checks: passed.

## Risks and follow-up

- Next.js introduces a Node Web runtime and framework upgrade surface.
- Full lint and presentation tests are deferred until before the first business screen.
- Local developer orchestration remains a separate developer-experience change.

## Rollback

Revert the future implementation commit and the preceding ADR-007 decision commit. The Vite
skeleton is restored; Go/Python/OpenAPI/database/user data are unaffected.

## Stop condition

After checks and Internal Review Agent evidence, stop without commit, push, merge, or Stage 3 and
wait for explicit human approval.
