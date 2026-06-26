# Web Frontend Architecture Amendment — Next.js Presentation Layer

| Field | Value |
| --- | --- |
| Document ID | WEB-ARCH-001 |
| Version | 1.0.1 |
| Status | Closed / Merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Current Web implementation target only; see ADR-007 |
| Dependencies | Documents 42–43; ADR-003; ADR-005; accepted ADR-007 |
| Last Review Date | 2026-06-26 |
| Next Review Date | 2026-12-26 |

## Goal

Replace the non-business Vite skeleton with a Next.js App Router + TypeScript + pnpm Web skeleton
without weakening the Go API, OpenAPI First, security, privacy, financial, or MVP boundaries.

## Decision and merge record

The human architecture request accepted the Next.js presentation-only decision and ADR-007 records
the architecture boundary. The amendment was implemented on the dedicated
`feature/nextjs-web-presentation` branch, reviewed, approved, and squash-merged into `develop`.

- PR: <https://github.com/AsifAbbasov/OpenInvest/pull/4>
- Merge commit / canonical Web baseline:
  `6a7748cc24fc852d42b90b0e0cb843b6020f3973`
- Merge date: 2026-06-26
- Merge method: squash merge into `develop`

Stage 2 is merged into `develop`, ADR-006 is accepted, and PR #4 was based on the updated
`develop` baseline. PR #4 targeted `develop` and was isolated to the Next.js Web Presentation
Amendment. After rebasing onto the Stage 2 baseline, checks were rerun for the Web amendment
scope before final review and human merge approval.

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

Revert merge commit `6a7748cc24fc852d42b90b0e0cb843b6020f3973` and, if architecture rollback is
approved through ADR, supersede ADR-007. The Vite skeleton is restored; Go/Python/OpenAPI/database
and user data are unaffected.

## Closure

The amendment is closed. Stage 3 remains not started and requires its own planning document,
review, and approval before implementation.
