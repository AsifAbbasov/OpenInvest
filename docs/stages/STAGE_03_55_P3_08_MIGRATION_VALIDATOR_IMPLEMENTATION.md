# Stage 3.55 — P3-08 Migration Validator Implementation

| Field | Value |
| --- | --- |
| Status | Implementation candidate v5 / frozen Internal Review subject |
| Date | 2026-09-03 |
| Canonical implementation base | `develop@b79a9d3c43621e56e598901bdf472771e8b68ef8` |
| Approved planning authority | `docs/stages/STAGE_03_54_P3_08_MIGRATION_VALIDATOR_PLAN.md` / Git blob `90fa563b9256b19055e2c14e52909596b392f221` / SHA-256 `c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba` |
| Original audit state | `P3-08=OPEN`; `31/32 = 96.875%` |
| Pre-review size exception | User/Principal-Architect authorization: up to ~1,850 production Go lines in `cmd/validate-migrations`, <=10 changed files, one P3-08 responsibility, stdlib-only; all semantic/review/CI/governance gates remain mandatory |
| Commit / push / PR | Not authorized at this record state |
| Runtime/schema/OpenAPI/frontend/dependency change | None authorized or introduced |

## 1. Scope

Stage 3.55 implements the already-approved Stage 3.54 v20 machine-enforceable migration-policy contract. It does not change the Stage 3.54 semantics and does not close P3-08 by itself.

Candidate implementation surfaces are limited to:

- `backend-go/cmd/validate-migrations/` — stdlib-only validator, strict manifest decoder, narrow SQL scanner/parser, policy constants and focused tests;
- `infrastructure/postgres/migrations/policy_manifest.json` — frozen identity-only baseline for `000001`–`000007`, plus append-only future enforced entries;
- `.github/workflows/ci.yml` — migration validation dominance for the existing `migrations`, `go`, and `go-race` SQL-executing jobs;
- this implementation record.

No existing SQL migration is edited. No migration framework, SQL parser dependency, parser generator, runtime migration service, schema/data migration, OpenAPI/frontend/business-logic change, or dependency/lockfile change is part of this candidate.

## 2. Implemented validator boundaries

The candidate implements:

- exact `.sql` discovery, canonical filename/pairing/identity rules, non-regular/symlink rejection, and future manifest bijection;
- strict JSON duplicate/unknown/missing/null/type/trailing-token checks and canonical integer decoding;
- exact frozen legacy baseline identity without retrospective policy claims;
- `local`, `repository`, and fail-closed `pr --base-sha=<40-hex>` modes;
- PR-base ancestry plus path/byte immutability for every base-existing migration SQL and append-only existing manifest entries;
- paired-SQL v1 `expand`-only enforcement;
- direction-specific execution metadata with independent UP/DOWN values and exact timeout-to-SQL binding;
- narrow CREATE TABLE / ADD COLUMN / CHECK / FOREIGN KEY / INDEX grammars and exact DOWN inverse forms, with stable direct rule classification for procedural/client/control and multi-predicate CHECK rejections;
- raw statement SHA-256 binding, duplicate-effect rejection, DDL impact bijection, finite enums/bounds, derived minimum risk, owner/touched-schema equality, observability, monitoring, rollout, rollback/roll-forward declarations, and typed authority references;
- canonical authority path grammar, `(kind,path)` uniqueness, regular-file/symlink-component checks for newly introduced evidence;
- frozen PostgreSQL 18.6 ColId disallowed projection used by governed unquoted identifiers;
- CI exact-SHA dominance: `migrations` validates first and publishes `validated_sha`; `go` and `go-race` require `migrations` and assert `validated_sha == GITHUB_SHA` before applying SQL.

## 3. Historical SQL immutability

All fourteen historical SQL files (`000001`–`000007`, UP+DOWN) are copied byte-for-byte from the protected Stage 3.54 evidence/protected base. Candidate comparison reports 14/14 exact equality. The sidecar baseline records their exact SHA-256 identities only.

## 4. Candidate size / review budget

Current production implementation consists of `main.go`, `policy.go`, `manifest.go`, `validator.go`, and `scanner.go`: exactly **1,847 physical Go source lines** after `gofmt`, within the separately authorized ~1,850-line production-Go exception. Candidate repository diff is exactly ten files and one responsibility: P3-08 migration validator hardening.

## 5. Local evidence state before Internal Review

Available sandbox evidence:

- `gofmt` on validator package: PASS;
- focused `go test -count=1 ./cmd/validate-migrations`: PASS in the sandbox harness;
- focused `go test -race -count=1 ./cmd/validate-migrations`: PASS in the sandbox harness;
- focused `go vet ./cmd/validate-migrations`: PASS in the sandbox harness;
- `go run ./cmd/validate-migrations --mode=local`: PASS against seven legacy pairs;
- canonical Stage 3.54 plan SHA binding and exact registry/edge/allowed-branch meta-tests: PASS;
- executable canonical TC ledger: `TC-001…TC-631` = 631/631 executed Go subtests; exact TC↔R polarity/owner edges and POS↔ALLOWED mapping revalidated; `MISSING_REQUIRED_TESTS=0`, `DUPLICATE_REQUIRED_TESTS=0`;
- frozen Stage 3.54 v20 proof replay: `V20_ALL_PROOFS=PASS`, property mutations `1895/1895`, extra red-team `31/31`;
- historical SQL comparison against protected Stage 3.54 evidence: 14/14 exact.

Environment limitations are not converted into false PASS claims: this sandbox does not contain the complete canonical repository worktree, its installed Go toolchain is `go1.23.2` rather than canonical `go 1.25.14`, and Docker/PostgreSQL are unavailable. Focused tests/race/vet therefore prove implementation logic under the available stdlib-compatible harness but do not substitute for repository-wide canonical-toolchain or PostgreSQL evidence. Repository-wide `go test ./...`, repository-wide `go vet ./...`, `docker compose config --quiet`, and disposable PostgreSQL apply/down/baseline/reapply remain mandatory exact-head GitHub CI evidence.

## 6. Governance state

This implementation record does not authorize commit, push, PR, Ready, merge, or P3-08 closure. Stage 3.55 must complete read-only Internal Review, Builder-only fixes, rerun required gates, receive separate human publication authorization, pass Draft-PR exact-head ten-check CI, complete fresh External published-head review and publication evidence, and receive later Ready/merge authorization. P3-08 remains OPEN until separately governed Stage 3.56 closure after protected implementation evidence.
