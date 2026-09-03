# Stage 3.54 — P3-08 Migration Validator / Migration Policy Hardening Plan

| Field | Value |
| --- | --- |
| Status | Planning-only candidate v20 after independent v19 `REQUEST CHANGES`; grants no implementation authorization |
| Date | 2026-09-03 |
| Canonical planning base | `develop@35c4413c84c3442989e76469742a2fd06155f484` |
| Protected-base tree | `4e6b51ead36bcf84af768fb890901003269631b0` |
| Finding | Original audit `P3-08 — migration validator weaker than migration policy` / `migration-validator policy hardening` |
| Planning-base audit state | 31 / 32 closed = 96.875%; P3-08 is the only remaining original finding |
| First planning review | `REQUEST CHANGES`; `P3-08-PLAN-01` derived from failed manifest/adversarial areas and remediated in v3 |
| v3 planning re-review | `REQUEST CHANGES`; four numbered findings: hidden Populate, procedural/dynamic SQL, incomplete policy coverage, incomplete positive acceptance coverage |
| v4 planning re-review | `REQUEST CHANGES`; PLAN-02/03 confirmed, PLAN-04/05 incomplete, new PLAN-06/07 added |
| v7 planning re-review | `REQUEST CHANGES`; PLAN-07/10/13/14/15 confirmed PASS; residual PLAN-04/11/16 plus finite-domain root cause remain |
| v8 planning re-review | `REQUEST CHANGES`; v7 blockers fixed, but residual PLAN-04 source purity/fidelity, residual PLAN-11 TC contradiction, and new rollout-reference semantic ambiguity remained |
| v9 planning re-review | `REQUEST CHANGES`; PLAN-11 and PLAN-18 confirmed PASS; residual PLAN-04 source completeness/provenance, PLAN-07/16 manifest primitives, and PLAN-10 literal CREATE TABLE/ADD COLUMN grammar remained |
| v10 planning re-review | `REQUEST CHANGES`; source-anchor completeness/provenance and v9 path/trim/CREATE TABLE fixes confirmed, but residual PLAN-04 evidence bindings, PLAN-16 aggregate proof, PLAN-10 nested SQL grammars and PLAN-12 CI discovery dominance remained; no PLAN-19 |
| v11 planning re-review | `REQUEST CHANGES`; all five v10 residuals were confirmed fixed, but residual PLAN-04 source-authority/evidence-scope defects and PLAN-16 aggregate/P3D-evidence/FK-negative-proof defects remained; no PLAN-19 |
| v12 planning re-review | `REQUEST CHANGES`; all five v11 blockers were confirmed fixed; residual PLAN-04 complete-machine observability, PLAN-16 CREATE TABLE max-boundary proof, and PLAN-10 keyword/identifier lexical overlap remained; no PLAN-19 |
| v13 planning re-review | `REQUEST CHANGES`; all three v12 blockers were confirmed fixed; preventive layer itself exposed wrong PostgreSQL ColId authority projection, three omitted bounded estimate fields, and incomplete migration-subject sibling discovery; no PLAN-19/GUARD-07 |
| v14 planning re-review | `REQUEST CHANGES`; all three v13 blockers were confirmed fixed in current semantics, but bound discovery was still name-hard-coded, ColId proof was satisfied by a duplicate summary instead of the primary authority, semantic-atom ownership lacked a real atom universe, and `S2-101→R018` was semantically unrelated; no PLAN-19/GUARD-07 |
| v15 planning re-review | `REQUEST CHANGES`; all v14 current-semantic blockers were confirmed fixed, but global single-source enforcement, occurrence-level ATOM bijection, stale evidence-partition summary, four semantically unrelated partial S2→R edges, and missing UP/DOWN-independent positive acceptance remained; no PLAN-19/GUARD-07 |
| v16 planning re-review | `REQUEST CHANGES`; v15 blockers were largely closed; residuals were incomplete structural-cardinality universe, coarse owner-level ATOM semantics, and the remaining prerequisite-only `S2-107→R028` edge; no PLAN-19/GUARD-07 |
| v17 planning re-review | `REQUEST CHANGES`; S2-107 and all 36 remaining partial S2→R bindings were accepted, but an active observer summary said 90 instead of derived 89, sibling child semantics still drifted beyond PROP-001…014 after evidence regeneration, BND/CARD scope definitions contradicted each other, and the prompt-required extra-red-team runner was missing from the frozen ZIP; no PLAN-19/GUARD-07 |
| v18 planning re-review | `REQUEST CHANGES`; package/reproducibility proof passed, but `TC/MPROP-591` and `TC/MPROP-597` froze stale S2 partition facts, four machine acceptance branches remained prose-only, and an active checklist mislabeled `SEMANTIC_PROPERTY_MANIFEST` as semantic authority; no PLAN-19/GUARD-07 |
| v19 planning re-review | `REQUEST CHANGES`; all v18 blockers confirmed closed; residuals were four-field UP/DOWN execution-metadata independence narrowed to timeout-only proof plus stale unvalidated `CANDIDATE_SIZE_BYTES` in the frozen review-package contract; no PLAN-19/GUARD-07 |
| Current planning revision | v20 — `P3-08-PLAN-01…18` and `P3-08-GUARD-01…06` retained; v19 residuals remediated narrowly by parameterizing `TC-598/599` across all four direction-specific execution metadata fields and binding the frozen package candidate-size fact to actual bytes; pending independent re-review |
| Implementation authorized here | No |
| Commit / push authorized here | No |
| Pull Request / Ready / merge authorized here | No |
| Runtime/database schema change authorized here | No |

## 1. Objective

Close original audit finding P3-08 by making repository migration validation materially match the
**machine-enforceable** part of the canonical Stage 2 migration strategy, while preserving a strict
boundary around controls that require human, operational, security/privacy, backup/restore, production
volume, or deployment evidence.

The remediation must not:

- rewrite already-merged SQL migration history;
- manufacture retrospective policy compliance;
- reduce existing PostgreSQL apply/rollback/reapply evidence;
- treat a manifest string as proof that an ADR/review/rehearsal happened;
- introduce a migration framework or external parser dependency;
- silently weaken the canonical migration strategy.

Stage 3.54 is planning-only. P3-08 remains OPEN throughout this stage.

## 2. Exact planning baseline

At protected `develop@35c4413c84c3442989e76469742a2fd06155f484`:

- protected-base tree: `4e6b51ead36bcf84af768fb890901003269631b0`;
- migration strategy blob: `d4656e2bb124fe6ff0783e619eaf608ed1082297`;
- current migration validator blob: `d601accc49306983320c9ba61f1a91f85a7495e7`;
- current CI workflow blob: `8437eba4f9f33ea3c331da839296e244719be6f8`;
- Stage 3.1 database-foundation record blob: `e10f1df9874d0e5aa78f7f4c0e70a76fcb23db61`;
- canonical review workflow blob: `3d0dd80e9d3825858c52b7dc0043010e549f720a`;
- current Source of Truth blob: `7af217c527cfddb4bb1379af92e765e351af64f8`.

The migration directory contains exactly seven merged SQL migration pairs / fourteen SQL files,
IDs `000001` through `000007`.

Those fourteen SQL files are immutable historical migration artifacts for P3-08.

Original audit state at the planning base:

```text
P0=0
P1=0
P2=0
P3=1
CLOSED=31/32
PERCENT=96.875%
REMAINING=P3-08
```

## 3. Planning-review finding ledger

Planning-review defects are permanent forensic records inside Stage 3.54. They are **not** added to the
original 32-finding audit. P3-08 remains the sole open original finding.

### P3-08-PLAN-01 — timeout requirement not closed through schema + tests

| Field | Value |
| --- | --- |
| Source | First planning review |
| Status | REMEDIATED in v3 / reviewer-confirmed in v3 re-review / RETAINED |
| Root cause | timeout policy existed in prose but not in typed schema and deterministic tests |
| Permanent regression | typed per-direction timeout metadata + exact SQL timeout-application checks + negative/positive tests |

### P3-08-PLAN-02 — hidden Populate can masquerade as `phase=expand`

| Field | Value |
| --- | --- |
| Source | v3 re-review Finding 1 |
| Status | REMEDIATED in v4 / reviewer-confirmed PASS in v4 re-review / RETAINED |
| Root cause | blocklist did not define a phase-specific executable statement contract |
| Permanent regression | narrow up-statement allowlist; DML/materialization/unknown statements rejected |

### P3-08-PLAN-03 — procedural/dynamic SQL can bypass lexical guards

| Field | Value |
| --- | --- |
| Source | v3 re-review Finding 2 |
| Status | REMEDIATED in v4 / reviewer-confirmed PASS in v4 re-review / RETAINED |
| Root cause | top-level lexical scanning did not close later execution of SQL hidden inside procedural/string bodies |
| Permanent regression | procedural/dynamic SQL and psql execution surfaces rejected in enforced v1 |

### P3-08-PLAN-04 — Stage 2 policy coverage registry was not canonical/exhaustive

| Field | Value |
| --- | --- |
| Source | v3 re-review Finding 3; v4 re-review Finding 1; v5 re-review Findings 4–5; v10 re-review Finding 1 |
| Status | REMEDIATED in v17 / v17 reviewer confirmed `S2-107` demotion and all 36 remaining partial S2→R bindings as defensible direct subsets / RETAINED in v19 |
| v4 residual defect A | registry declared six legal dispositions but used non-enum variants |
| v4 residual defect B | normative controls were still compressed/omitted |
| v4 residual defect C | dependency **graph validity** was incorrectly conflated with dependency **semantic completeness** |
| Impact | false `MIG021_POLICY_COVERAGE` completeness / machine-proof overclaim |
| v5 remediation | stable `S2-*` IDs and six-value enum improved coverage but v5 reviewer still found three omitted normative controls plus two machine-proof overclaims |
| v6 remediation | registry expanded and evidence overclaims were corrected, but v6 mixed a derived CI-hardening control into `S2-*` and omitted atomic Low/Destructive risk-example anchors |
| v7 residual | source/derived namespaces and risk anchors were fixed, but `S2-137` copied only `logs` and silently dropped canonical `metrics` privacy wording |
| v8 remediation | `S2-137` preserves the exact combined metrics+logs sensitive-content prohibition; machine proof remains limited to controlled validator/CI surfaces and total production observability hygiene remains Security/Operations evidence |
| v8 residual | `S2-025/027/028` and risk-example rows still fused canonical source wording with P3-08-derived strengthenings; Snapshot Expand was missing as an atomic control; `market/inflation`, `encryption-key`, and `cryptographically destroyed` qualifiers were reduced |
| v9 remediation | source-only `S2-*` wording is restored; derived pre-existing-index, executable-timeout-binding and operation-risk-floor strengthenings move to `P3D-*`; missing Snapshot Expand is appended as stable `S2-165`; lost source qualifiers are restored |
| v9 residual | canonical `never manually`, `require diagnosis`, `canonical ledger`, exact `key hierarchy`/`Vault policy` fidelity and source/derived dependency/per-direction provenance were still incomplete; no byte-bound source-anchor registry existed |
| v10 remediation | adds exact `SA-001…SA-082` source anchors bound to Stage 2 line ranges and fragment SHA-256; every active `S2-*` is mapped exactly once; adds `S2-166` (`never manually`) and `S2-167` (`require diagnosis`); restores lossy qualifiers; moves dependency-graph/semantic-adequacy/per-direction strengthenings to `P3D-015…017`; R045 fails on missing/hash-drift/unknown/duplicate source accountability |
| v10 residual | `S2-021` was incorrectly marked fully MACHINE even though index grammar/concurrency cannot prove old-application behavioral compatibility; `S2-093` cited dependency-graph rule R018 even though graph validity cannot prove an approved ongoing Populate responsibility |
| v11 remediation | `S2-021` becomes `STRUCTURE_PLUS_HUMAN_ADEQUACY` with R008/R029 explicitly limited to index structure/concurrency and Compatibility review owning old-version behavior; `S2-093` has no machine rule and requires operational/architecture approval evidence; §23.1a freezes complete-machine, paired-SQL-scope-rejected, partial-machine and no-machine S2 evidence scopes and P3D-021/R049/TC-469…472 reject semantically unrelated or overclaiming S2→R bindings |
| v11 residual | `SA-001` hashed only the lifecycle sequence and excluded the source byte carrying `mandatory`; `S2-109` and `S2-118` were still classified full-machine even though R010/R028 observe only governed migration SQL/disposable DOWN, not every production rollback/snapshot mechanism |
| v12 remediation | `SA-001` now binds lines 16–20 including the Purpose scope plus `mandatory`; line-accounting excludes only genuinely non-normative/context bytes; `S2-109`/`S2-118` move to partial machine evidence with explicit external rollback/snapshot remainders; R045/R049 mutation proofs now cover authority-qualifier drift and external-behavior overclaim |
| v12 residual | `S2-005` still treated a global **no migration rewrites ledger history** invariant as complete-machine even though separately governed operational migrations (including Populate) lie outside paired-SQL observation |
| v13 remediation | re-run observer-universe analysis across every globally quantified migration/versioning S2 control; only 7 rows remain complete-machine, 17 global rows move to partial evidence, and GUARD-01/R056 requires requirement-subject ⊆ machine-observer before any complete-machine classification |

### P3-08-PLAN-05 — positive acceptance coverage was still incomplete

| Field | Value |
| --- | --- |
| Source | v3 re-review Finding 4; v4 re-review Finding 2 |
| Status | REMEDIATED in v16 / v16 reviewer-confirmed PASS for explicit UP≠DOWN positive branch / RETAINED |
| v4 residual examples | valid high risk, sensitive/mixed classification, measured observability, permitted N/A observability and mixed profile had no explicit positive |
| Impact | validator could reject reviewed-valid branches while all negative tests stayed green |
| v5 remediation | complete `ALLOWED-*` registry; every allowed enum/cross-field/statement/rollback branch maps to at least one explicit positive `TC-*`; missing map is `MIG021_POLICY_COVERAGE` |

### P3-08-PLAN-06 — up/down statement safety was internally contradictory

| Field | Value |
| --- | --- |
| Source | v4 re-review Finding 3 |
| Status | REMEDIATED in v5 / reviewer-confirmed PASS in v5 re-review / RETAINED |
| Root cause | v4 applied lifecycle destructive-DDL rules to both `up.sql` and `down.sql` while also requiring a paired safe rollback |
| Concrete failure | safe Expand `CREATE TABLE` required down `DROP TABLE`, which v4 could reject as destructive outside Contract; exempting all down DROP would permit unrelated destructive rollback |
| Impact | either valid additive migrations could never pass or down safety would be weakened by an improvised exception |
| v5 remediation | direction-specific contracts: future paired-SQL v1 supports only **Expand** in `up.sql`; `down.sql` is explicitly a **disposable CI/rehearsal inverse, never automatic production rollback authority**; scanner derives exact up effects and permits only exact scoped inverses in down; DML/CASCADE/IF EXISTS/unrelated targets/partial or extra inverses fail closed |
| Additional simplification | Populate, Switch, Validate and Contract are rejected by paired-SQL v1 and require separately governed mechanisms; this removes fake reversibility for forward-only Validate/Contract operations |

### P3-08-PLAN-07 — advertised test contract was not durably self-contained

| Field | Value |
| --- | --- |
| Source | v4 re-review Finding 4; v5 re-review Finding 6 |
| Status | REMEDIATED in v10 / reviewer-confirmed PASS for authority-path and ASCII-trim primitives in v10 re-review / RETAINED |
| Root cause | v4 referred to cases 1–119 defined only in rejected/unmerged v3; v5 then left several allegedly frozen enums/bounds/type/operator/hash semantics unspecified |
| Impact | future protected-repository audit could not reconstruct the meaning of more than half the claimed permanent test contract |
| v5 remediation | all test definitions became physically self-contained |
| v6 remediation | lock modes, replication enum, major numeric bounds, type allowlist, FK defaults and raw-statement hashing were frozen, but numeric/string/CHECK literal grammar remained implementation-defined |
| Numbering rule | v5 is the first prospective self-contained canonical test registry; earlier rejected-candidate numeric labels have no canonical authority |
| v9 Builder residual | aggregate/open-domain manifest fields still allowed Stage 3.55 to choose element/reference semantics or string normalization even though finite enums were closed |
| v9 remediation | §7.11 freezes aggregate field JSON types/requiredness, binds `rollout.metrics[]` and `monitoring.signals[]` to exact measured observability category keys, and freezes required open-text validation/preservation semantics; H181…H185 + P3D-014 + R044 + TC-435…TC-441 retain regression coverage |
| v9 residual | `safe normalized repository-relative path` and `ASCII-edge trimming` still allowed materially different Stage 3.55 accepted languages |
| v10 remediation | §7.9 freezes an ASCII canonical path grammar with no implicit normalization and exact `(kind,path)` uniqueness; §7.11 freezes `ASCII_TRIM_BYTES={09,0A,0B,0C,0D,20}` and preserves non-ASCII whitespace/content; P3D-019 + R046/R047 + TC-447…TC-458 retain boundary coverage |


### P3-08-PLAN-08 — timeout metadata was not necessarily bound to actual PostgreSQL execution

| Field | Value |
| --- | --- |
| Source | Builder v5 pre-review closure-design audit after v4 `REQUEST CHANGES`; v5 re-review Finding 2 |
| Status | REMEDIATED in v6 / reviewer-confirmed PASS in v7 re-review / RETAINED |
| Root cause | earlier planning could validate timeout metadata as structure without requiring the migration SQL execution path to actually apply the declared PostgreSQL timeout values before governed DDL |
| Concrete failure | manifest declares `lock_timeout_ms=5000` and `statement_timeout_ms=30000`, while SQL executes DDL without matching `SET` / `SET LOCAL`, applies different values, applies them after DDL, or weakens/resets them before DDL |
| Impact | validator could report timeout-policy compliance while PostgreSQL executes without the declared protection; this is false machine evidence and leaves the validator weaker than policy |
| v5 remediation | §9 defines direction-specific execution contracts: transactional directions require exact `SET LOCAL lock_timeout` + `SET LOCAL statement_timeout` before protected DDL; supported non-transactional concurrent-index directions require exact session `SET`; values must equal manifest milliseconds; missing/mismatched/late/duplicate/conflicting/reset GUC controls reject |
| Permanent regression | full `direction × transaction_mode × GUC` matrix under R011/R012: existing TC-081…090 plus v6 TC-332/333 for non-transactional DOWN missing lock/statement GUC; positive TC-278…281 retained |
| Evidence boundary | machine validation proves the declared timeout values are actually applied on the validated SQL path; it does **not** claim the chosen values are sufficient for unknown production cardinality/load |


### P3-08-PLAN-09 — mixed transactional/non-transactional UP effects were not explicitly closed

| Field | Value |
| --- | --- |
| Source | Builder v5 pre-review closure-design audit |
| Status | REMEDIATED in v5 / reviewer-confirmed PASS in v5 re-review / RETAINED |
| Root cause | one `up_transaction_mode` existed, but the earlier contract did not explicitly reject a direction mixing ordinary transactional DDL with `CREATE INDEX CONCURRENTLY` |
| Concrete failure | an up file could contain `CREATE TABLE ...` plus `CREATE INDEX CONCURRENTLY ...`; treating the file transactional breaks PostgreSQL concurrent-index rules, while treating it non-transactional silently broadens the reviewed non-transactional surface |
| Impact | Stage 3.55 could invent unreviewed execution semantics or accept a migration whose framing cannot satisfy the reviewed contract |
| v5 remediation | §15/§16 freeze direction homogeneity: transactional up may contain only transactional allowlisted effects; non-transactional up may contain only supported concurrent-index effects and timeout session controls; mixed directions reject/rescope |
| Permanent regression | H101/H102 + `TC-307…TC-309` under R032 |

### P3-08-PLAN-10 — future DDL subgrammar still allowed masking or hidden semantic side effects

| Field | Value |
| --- | --- |
| Source | Builder v5 pre-review closure-design audit; v10 re-review Findings 3–4 |
| Status | REMEDIATED in v16 / v15 reviewer confirmed current PostgreSQL ColId projection but found section-local rather than global single-source enforcement / pending re-verification |
| Root cause | "safe CREATE TABLE / CHECK / FK" wording was narrower than v4 but not yet exact enough to prevent drift-masking (`IF NOT EXISTS`) and hidden behavior through generated/serial/custom types, function-bearing CHECK/default expressions, referential actions, or unsupported table storage/lifecycle options |
| Concrete failure | `CREATE TABLE IF NOT EXISTS` can silently reuse a pre-existing object and the paired down can later drop it; a CHECK can call a function; FK `ON DELETE CASCADE` changes future write/delete semantics; `SERIAL`/identity/generated/custom types introduce hidden objects/dependencies |
| Impact | validator could accept behavior outside the reviewed additive Expand subset, or a disposable down could destroy an object that was not actually created by the migration |
| v5 remediation | §§15–17 narrowed DDL substantially; v6 further removed embedded CREATE TABLE constraints, but `frozen simple column-key grammar` and identifier truncation semantics remained incomplete |
| v9 residual | CREATE TABLE/ADD COLUMN defined allowed/forbidden clause families but not exact clause order, multiplicity, empty/trailing-comma/duplicate-column or mandatory `COLUMN` behavior |
| v10 remediation | §15.3 introduces one exact token grammar for `create_table`, `column_def`, and scalar types; §15.4 binds ADD COLUMN to the same `column_def`, mandates `COLUMN`, freezes clause order/multiplicity, 1..64 table columns, duplicate-name rejection and fail-closed extra-token behavior |
| v10 residual | value bounds existed for `numeric(p,s)` / `varchar(n)` parameters but their token spellings were not frozen; CHECK froze only the predicate and DOWN froze effect classes rather than complete accepted statement productions |
| v11 remediation | §15.3 freezes canonical decimal-token productions for `precision`, `scale`, and `varchar_length`; §15.5 freezes the complete CHECK statement envelope; §17.2 freezes one literal production for every accepted DOWN class; P3D-023…025 + R051…R053 + TC-480…497 recursively close the nested SQL language |
| v11 residual | FOREIGN KEY local/reference list-cardinality equality was normative but lacked deterministic negative regression, so an implementation could omit the equality check and rely on PostgreSQL to reject |
| v12 remediation | explicit two-way FK cardinality-mismatch negatives plus a matching multi-column positive bind the equality invariant to R033 before PostgreSQL execution; the FK grammar owner is also included in the aggregate semantic freeze |
| v12 residual | keyword terminals and unquoted identifier regex overlap, but Stage 3.54 did not freeze reserved-keyword exclusion/context precedence; two scanners could accept different SQL languages |
| v13 remediation | §15.3a froze an exact 78-member `RESERVED_KEYWORD` extraction, but v13 review proved that category alone is not PostgreSQL `ColId` admissibility |
| v13 residual | PostgreSQL `TYPE_FUNC_NAME_KEYWORD` words such as `collation`, `concurrently`, and `cross` are not legal `ColId` table/column names, yet v13 accepted them because it rejected only `RESERVED_KEYWORD` |
| v14 remediation | §15.3a freezes the actual REL_18_6 `ColId := IDENT | unreserved_keyword | col_name_keyword` grammar and derives the exact 101-member project disallowed set as `TYPE_FUNC_NAME_KEYWORD ∪ RESERVED_KEYWORD`; both `kwlist.h` and `gram.y` blob identities, category sets, union, hashes and contextual witnesses are frozen; authority-projection mutation is mandatory |
| Permanent regression | H103…H109 + H194…H203 + `TC-310…TC-319` + `TC-459…TC-497` under R033/R048/R051/R052/R053 |

### P3-08-PLAN-11 — author-declared risk could underclassify canonical medium-risk operation classes

| Field | Value |
| --- | --- |
| Source | Builder v5 pre-review Stage 2 classification cross-check; v5 re-review Finding 3 |
| Status | REMEDIATED in v9 / reviewer-confirmed PASS in v9 re-review / RETAINED |
| Root cause | v5 had strong gates after a risk value was chosen and still permitted constraint syntax inside CREATE TABLE, allowing machine-obvious medium effects to hide under a low CREATE TABLE minimum |
| Concrete failure | a pre-existing-table index or new FK/CHECK constraint could declare `risk=low`, avoiding staged-rollout structural requirements while all low-risk validations pass |
| Impact | risk-specific policy gates can be bypassed by under-declaring risk; validator remains weaker than the classification policy |
| v5 remediation | explicit index/ADD CONSTRAINT minimums were added but did not cover embedded CREATE TABLE constraints |
| v6 remediation | CREATE TABLE removed CHECK/FK/PK/UNIQUE/EXCLUDE, but still allowed column-level `NOT NULL`; v7 reviewer correctly identified PostgreSQL NOT NULL as a constraint that could remain under the low ADD COLUMN floor |
| v8 remediation | paired-SQL v1 rejects `NOT NULL` in both CREATE TABLE and ADD COLUMN; the only low column/table forms are nullable; any future NOT NULL support is explicit scope expansion with a reviewed medium-or-higher risk contract |
| v8 residual | `TC-331` still said generic nullable ADD COLUMN may remain low, contradicting the frozen `ADD COLUMN + literal DEFAULT => minimum medium` rule and TC-424 |
| v9 remediation | `TC-331` is narrowed to nullable ADD COLUMN **without DEFAULT**; TC-265 remains the medium positive path for literal DEFAULT and TC-424 remains the laundering rejection |
| Permanent regression | H110/H111/H162/H171/H172/H176 + `TC-320…TC-324`,`TC-413`,`TC-414`,`TC-424` under R033/R034; Stage 2 registry explicitly includes classification-example dispositions |

### P3-08-PLAN-12 — a CI path could execute migration SQL before validator success

| Field | Value |
| --- | --- |
| Source | Builder v5 pre-review inspection of current `.github/workflows/ci.yml`; v5 re-review Finding 1; v10 re-review Finding 5 |
| Status | REMEDIATED in v11 / reviewer-confirmed PASS in v11 re-review / RETAINED |
| Root cause | v5 modeled only `migrations` + `go`, while frozen CI also has `go-race` with its own PostgreSQL migration apply path |
| Concrete failure | PR adds a migration containing policy-forbidden SQL; GitHub schedules `Go tests` first/parallel and its `Apply PostgreSQL migrations` step executes exact candidate bytes before `PostgreSQL migration validation` returns failure |
| Impact | validator is a merge gate but not an execution-before-use gate across CI; unsafe/rejected SQL can execute in CI under the PostgreSQL service superuser before validation |
| v5 remediation | dominance was added for `go` only and was incomplete |
| v6 remediation | exact current SQL-executing inventory is frozen as job IDs `migrations`, `go`, `go-race`; both dependent Go jobs require `needs: migrations` plus exact `validated_sha`; conservative workflow inventory test fails on migration-application markers outside the frozen inventory |
| v10 residual | CI executes every `*.up.sql`, while validator planning said only an invalid `.sql` file that "resembles a migration" rejects; a refactor could therefore ignore `evil.up.sql` while CI still executes it |
| v11 remediation | §10 defines `VALIDATOR_SQL_DISCOVERY_SET` as every migration-directory basename ending exactly `.sql`; every member must match the canonical UP/DOWN filename grammar or validation fails, never ignores; current `*.up.sql` CI execution set is mechanically proven a subset of validator discovery/approval through P3D-026/R054/TC-498…502 |
| Permanent regression | H112…H115 + H204 + R035/R054 + `TC-325…TC-327`,`TC-337…TC-341`,`TC-498…TC-502`; exact published-head CI is required evidence |

### P3-08-PLAN-13 — PostgreSQL identifier truncation was not frozen

| Field | Value |
| --- | --- |
| Source | v6 re-review Finding 4 |
| Status | REMEDIATED in v7 / reviewer-confirmed PASS in subsequent re-reviews / RETAINED |
| Root cause | future SQL identifiers were lowercase/unquoted but had no byte-length invariant, while PostgreSQL 18 truncates identifiers beyond its default 63-byte maximum |
| Concrete failure | validator derives effect identity from a >63-byte source identifier while PostgreSQL creates/catalogs the truncated name; two source names sharing the first 63 bytes can collide |
| Impact | validator identity and database identity can diverge, weakening exact inverse and ownership proofs |
| v7 remediation | future v1 SQL identifiers use exact ASCII regex `^[a-z][a-z0-9_]{0,62}$`; every identifier component is therefore 1–63 bytes; overlong/quoted/uppercase/Unicode/leading-underscore forms reject before SQL execution |
| Permanent regression | H144/H145/H152 + R039 + 63-byte PASS, 64-byte REJECT and same-prefix collision cases |

### P3-08-PLAN-14 — TC→R and R→TC semantic mappings diverged

| Field | Value |
| --- | --- |
| Source | v6 re-review Finding 5 |
| Status | REMEDIATED in v7 / reviewer-confirmed PASS in v7 re-review / RETAINED |
| Root cause | meta-audit checked ID existence and polarity but did not assert exact equality of the two independently written edge sets |
| Concrete failure | a TC row can name a rule absent from that rule's coverage list, or a rule can name a test that does not name the rule |
| Impact | two conforming-looking coverage reports can produce different semantic graphs while all headline counts remain green |
| v7 remediation | TC rows are the canonical edge declaration; the rule coverage table is an exact generated projection; package meta-audit requires bidirectional edge-set equality |
| Permanent regression | H146/H147 + R037 + asymmetric-edge negative fixtures and exact-graph positive fixture |

### P3-08-PLAN-15 — stale registry summary contradicted exact rule range

| Field | Value |
| --- | --- |
| Source | v6 re-review Finding 6; v15 re-review stale `60` vs derived `59`; v17 re-review V17-F01 (`90` vs derived `89`) |
| Status | REMEDIATED again in v19 / all active observer/partition numeric summaries are derived from the canonical S2 partition and no stale numeric literal is an independent authority / pending independent re-verification |
| Root cause | prose summary retained a stale bare count (`35-rule`) after the exact registry became `R001…R036` |
| Concrete failure | implementation/closure readers can derive different expected totals from summary vs normative registry |
| Impact | semantic-freeze document contains conflicting authorities |
| v7 remediation | bare registry counts are forbidden as normative authority; summaries always state exact ID ranges and package meta-audit verifies every declared range/count pair |
| v15 residual | one active source-evidence summary retained `60 partial` after the derived registry became `59`; NLA froze the stale line but did not prove arithmetic consistency |
| v17 residual | after S2-107 demotion the canonical partition derived `7+46+36=89` observer-bearing, while one active GUARD-01 summary and the v17 meta-audit still required `90` |
| v19 remediation | one canonical S2 partition is parsed first; observer-bearing is derived as complete+scope-rejected+partial; every active numeric observer/partition representation is checked against the derived tuple; a `89→90` mutation fails |
| Permanent regression | H148/H263 + R038/R056/R063 + stale-summary/observer mutation fixtures |

### P3-08-PLAN-16 — Builder meta-audit proved structure but not semantic-freeze completeness

| Field | Value |
| --- | --- |
| Source | Builder root-cause analysis after v6 `REQUEST CHANGES`; v10 re-review Finding 2; v16 re-review V16-F01/F02; v17 re-review V17-F02/F03 |
| Status | REMEDIATED architecturally in v19 / v17 reviewer confirmed byte manifests and known PROP fixtures but found sibling child-semantic survivors and contradictory BND/CARD scope / pending independent re-verification |
| Root cause | prior meta-audits proved contiguous IDs, valid references, polarity and positive mappings, but did not prove source/derived namespace purity, literal symbol resolution, PostgreSQL identifier semantics, exact grammar closure, bidirectional coverage-edge equality **or exhaustive finite-domain field definitions** |
| Concrete failure | a plan can be internally count-consistent yet still leave implementation-selected semantics or two inconsistent semantic mappings |
| Impact | reviewer repeatedly finds defects outside the syntactic registry checks; headline `META_AUDIT=PASS` was weaker than users reasonably inferred |
| v7 remediation | added source purity, grammar, edge equality and placeholder lint, but still allowed a required field to have `unknown value → reject` without a literal exhaustive allowed set |
| v8 remediation | adds an exact finite-domain registry for every finite manifest field, scanner-derived `statement_class` mapping, exact five-value authority-kind enum, and meta-gates requiring domain presence, normative equality and no delegation to Stage 3.55/code/tests |
| v9 Builder residual | finite-domain exactness still did not prove semantic completeness of open-domain arrays/text; this is a residual of the existing semantic-meta-audit class, not a new PLAN ID |
| v9 remediation | package proof audit now inventories formal manifest aggregate/open fields and checks their exact type/reference/normalization contracts through §7.11, P3D-014, R044 and TC-435…TC-441 |
| v10 remediation | Builder meta-audit now verifies exact `SA-*` line-range/hash/S2 accountability against the canonical Stage 2 bytes, exact path grammar markers, exact trim-byte set, literal CREATE TABLE/ADD COLUMN productions and their deterministic boundary tests; structural PASS cannot bypass these semantic checks |
| v10 residual | `P3D-008` claimed no implementation-selected alias/enum/bound/grammar/normalization but was mapped only to R040/R041, leaving path, trim, finite-domain, CREATE TABLE/ADD COLUMN and other frozen semantic surfaces outside the aggregate proof |
| v11 remediation | §23.2a defines one exact `SEMANTIC_FREEZE_RULE_SET`; P3D-008 points only to aggregate R050; R050 fails if any required semantic rule is omitted/added and mutation cases exercise enum, path, trim, CREATE TABLE/ADD COLUMN, CREATE INDEX, scalar/value/type-parameter, CHECK, DOWN and validator-discovery semantics; Builder meta-audit checks the exact set rather than known needles |
| v11 residual | aggregate set omitted existing FK grammar owner R033; P3D-003 still claimed all literal/operator/type bounds while citing only R040; FK list-cardinality rejection lacked deterministic negative proof |
| v12 remediation | semantic-freeze scope is derived from an exact semantic-owner inventory instead of a hand-picked family list; R033 is in-scope; P3D-003 is narrowed to the R040 scalar-data/CHECK-predicate surface while P3D-023 exclusively owns type-parameter bounds; P3D-027/R055 adds machine-scope accounting for every P3D binding; FK cardinality mismatch is covered by explicit negatives |
| v12 residual | the explicit CREATE TABLE `1..64` project cap had no exact `64 PASS / 65 REJECT` witnesses, so the bound could disappear while every registered test stayed green |
| v13 remediation | GUARD-02/R057 inventories finite bounds globally and requires exact boundary/adjacent-invalid witnesses plus mutation-kill coverage; CREATE TABLE gets TC-521/522 and the guard scans sibling cardinality/length/range bounds rather than only this fixture |
| v13 residual | the BND registry was still a hand-authored expected set and omitted three formal-manifest integer fields (`affected_rows_estimate`, `disk_impact_bytes_estimate`, `wal_impact_bytes_estimate`) with exact `0..INT64_MAX` bounds |
| v14 remediation | BND-17…19 add independent lower/upper witnesses for those fields; more importantly, Builder discovery parses bounded integer manifest fields from §§7.4/7.5/7.11 and requires discovered semantic keys to be represented in BND, while a synthetic bounded-field mutation must fail without relying on an expected BND count |
| v16 residual | BND occurrence enforcement was sound, but the declared boundary universe omitted structural cardinalities; the owner-level ATOM registry hashed broad R bodies rather than individual machine properties |
| v17 remediation/residual | SEM/CARD all-line manifests and `PROP-001…014` closed the known v16 mutations, but independent sibling JSON/dependency/filename properties could still drift after legitimate evidence regeneration; BND/CARD scope also had two conflicting definitions |
| v19 remediation | machine acceptance/rejection is exactly the canonical `TC↔MPROP` bijection; every TC condition/outcome/owner set has a direct atomic MPROP witness, mutation obligations are registry-derived for every MPROP, and exactly one `TA-01` authority defines BND vs structural-cardinality/MPROP scope. SEM/CARD/NLA are byte-accountability only and cannot masquerade as semantic completeness |
| Permanent regression | H149…H175 + H181…H185 + H200 + H249…H269 + R037…R050/R057/R063/R072…R075 plus registry-derived property mutation and package-contract gates |

### P3-08-PLAN-17 — finite-domain field completeness was not mechanically audited

| Field | Value |
| --- | --- |
| Source | Builder root-cause analysis after v7 `REQUEST CHANGES` Findings 3–4 |
| Status | REMEDIATED in v8 / reviewer-confirmed PASS in v9 re-review / RETAINED |
| Root cause | semantic lint looked for known placeholder phrases but did not inventory every required finite-domain manifest field and prove that its exhaustive value set was frozen inside Stage 3.54 |
| Concrete failure | `statement_class` could reject an "unknown" value even though the valid set was undefined; authority `kind` said values "include" five examples and delegated the exact enum to future code/tests |
| Impact | two Stage 3.55 implementations could produce incompatible canonical manifests while both claiming semantic-freeze compliance |
| v8 remediation | §7.10 is the single finite-domain registry; every finite field has an exact domain, recognized-vocabulary versus paired-v1-supported subset where relevant, and a normative authority section; package meta-audit compares the exact field set and exact member sets |
| Permanent regression | H163…H175 + P3D-009 + R042 + TC-416…TC-423 |

### P3-08-PLAN-18 — cross-object rollout/evidence reference semantics were not frozen

| Field | Value |
| --- | --- |
| Source | v8 re-review Finding 4 |
| Status | REMEDIATED in v9 / reviewer-confirmed PASS in v9 re-review / RETAINED |
| Root cause | `rollout.plan_ref` and `authority_refs[kind=staged_rollout]` represented the same rollout-plan identity through two independent fields without an exact equality/reference rule |
| Concrete failure | a manifest could bind a validated immutable staged-rollout authority ref to plan A while `rollout.plan_ref` named plan B, or two implementations could interpret `plan_ref` as path/hash/index/string differently |
| Impact | medium/high risk evidence binding was implementation-defined; two compliant implementations could accept incompatible manifests |
| v9 remediation | remove `rollout.plan_ref` from the manifest entirely; `authority_refs[kind=staged_rollout]` becomes the sole normative rollout-plan identity; `rollout.mode=staged` requires exactly one such reference, while `mode=standard` requires zero |
| Proof obligations | PO-03, PO-05, PO-10, PO-16, PO-19 |
| Permanent regression | H177…H180 + P3D-013 + R043 + TC-425…TC-427 |

### Finding-ledger invariants

1. `P3-08-PLAN-01…18` remain permanently visible in every later Stage 3.54 revision.
2. No reviewer-confirmed finding may be silently reopened, renamed, merged away or erased.
3. A newly discovered material planning defect receives the next `P3-08-PLAN-NN`; a residual of an existing finding stays under that existing ID with a new residual entry.
4. Every remediation must add or strengthen a permanent prevention rule and deterministic regression.
5. `REMEDIATED` means Builder design exists; `REVIEWER_CONFIRMED` requires explicit designated-reviewer confirmation.
6. Original audit arithmetic remains exactly 31/32 until separately governed protected P3-08 closure.

### Preventive failure-class guard ledger — v13

The immutable `P3-08-PLAN-01…18` ledger remains forensic history. v13 adds a **separate preventive guard registry** so a recurring root cause is machine-gated even when a concrete Reviewer finding is correctly classified as a residual of an old PLAN. A GUARD is not a replacement/renumbering of PLAN findings and does not manufacture `PLAN-19`.

| Guard | Generic failure class | Mandatory pre-review prevention |
| --- | --- | --- |
| `P3-08-GUARD-01` | **evidence-observability / quantified-subject escape** — a rule proves a subset of an `all/every/never/no` requirement but the row is promoted to complete-machine | freeze requirement subject universe + observer universe; complete-machine requires `subject_universe ⊆ observer_universe`; scan all sibling global controls |
| `P3-08-GUARD-02` | **finite-boundary proof omission** — a frozen min/max/cardinality/length/range exists in prose/grammar without exact boundary witness and adjacent-invalid rejection | global finite-bound inventory; boundary PASS/NEG pair or explicit mathematical N/A; semantic mutation must kill Builder gate |
| `P3-08-GUARD-03` | **lexical token-class overlap** — two token classes overlap and precedence/context/exclusion is implementation-selected | inventory every non-empty token-class intersection; freeze owner/precedence/context/exclusion set and positive/negative witnesses |
| `P3-08-GUARD-04` | **latent semantic owner omission** — a normative semantic atom exists in prose but no exact R owner represents it | semantic-atom registry/inventory; each atom maps exactly once to an R owner or explicit reviewer-only evidence class |
| `P3-08-GUARD-05` | **mutation survivor** — a materially changed frozen semantic constant can leave all canonical gates green | one mutation-kill proof per semantic-atom class; changing bound/set/owner/normalization/evidence scope must fail |
| `P3-08-GUARD-06` | **fixture-only remediation** — Builder fixes named Reviewer examples but does not scan the complete sibling domain | every remediation records generic predicate, sibling domain, exhaustive scan result and permanent regression before Builder PASS |

Guard invariants:

1. `P3-08-GUARD-01…06` are exact and append-only for this prevention layer.
2. Every future Reviewer blocker must be classified twice: forensic PLAN class **and** preventive GUARD class/predicate (or explicit reason no Guard applies).
3. A known example passing is insufficient; the complete sibling domain must be scanned.
4. `BUILDER_PROOF_OBLIGATIONS_PASS` requires `PO-01…20` plus `GP-01…06`.
5. A Guard may be strengthened without changing the forensic PLAN ID of a residual finding.

## 4. Broader fail-closed bypass/error register

The finding ledger records defects actually found in review. The H-register records bypass/error
classes the design must proactively defend against. H entries are prevention requirements, not
additional original-audit findings.

| ID | Bypass/error class | Required fail-closed control |
| --- | --- | --- |
| P3-08-H01 | old SQL edited + candidate checksum updated | compare exact base SQL blobs/paths against protected PR base |
| P3-08-H02 | historical manifest metadata rewritten without SQL edit | base-existing manifest entries immutable/append-only |
| P3-08-H03 | duplicate JSON key changes meaning under last-key-wins parsing | reject duplicate keys at every object depth |
| P3-08-H04 | unknown JSON field silently ignored | reject unknown fields |
| P3-08-H05 | trailing second JSON document/payload | reject trailing non-whitespace JSON tokens |
| P3-08-H06 | symlink/non-regular migration file escapes byte identity | reject symlinks and non-regular migration SQL/manifest files |
| P3-08-H07 | case-only filename collision differs macOS/Linux | lowercase ASCII grammar + case-insensitive collision rejection |
| P3-08-H08 | malformed/ambiguous ID/name grammar | strict six-digit ID + ASCII stem grammar |
| P3-08-H09 | comment-obfuscated forbidden token | lexical normalization before statement classification |
| P3-08-H10 | keywords in comments/strings cause false positives | lexical scanner separates inert comments/literals from executable statement syntax |
| P3-08-H11 | PR mode without real base commit | explicit PR mode + mandatory base SHA |
| P3-08-H12 | shallow checkout silently disables base comparison | CI makes base object available; unresolved base is fatal |
| P3-08-H13 | supplied base is not supported ancestor | resolve commit + ancestry validation |
| P3-08-H14 | future migration hides inside legacy baseline | frozen separate legacy structure capped at `000007` |
| P3-08-H15 | retrospective metadata invented for legacy | legacy stores identity/hashes only + explicit non-retroactive status |
| P3-08-H16 | evidence reference string treated as approval | typed reference structure only; substantive adequacy external |
| P3-08-H17 | final CI/review evidence required before CI can run | final evidence stays outside pre-CI manifest |
| P3-08-H18 | one transaction mode cannot model up/down safely | separate up/down transaction modes |
| P3-08-H19 | non-positive/inverted timeout satisfies field presence | numeric bounds + cross-field consistency |
| P3-08-H20 | manifest schema silently weakened | schema version frozen; redesign requires reviewed scope |
| P3-08-H21 | hard-coded Stage 3.1/3.11 guards disappear | preserve or replace with equal/stronger assertions |
| P3-08-H22 | first/reapplied schemas diverge after rollback rehearsal | deterministic catalog fingerprint/invariant equivalence |
| P3-08-H23 | rollback leaves managed objects behind | explicit rollback-baseline postcondition |
| P3-08-H24 | validator error too ambiguous for durable tests | stable typed validation errors |
| P3-08-H25 | `INSERT ... SELECT` hides Populate under Expand | reject all DML/data-moving classes in enforced paired-SQL v1 |
| P3-08-H26 | CTAS / `SELECT INTO` materializes a backfill | classify and reject data-producing DDL under Expand |
| P3-08-H27 | `COPY`, `MERGE`, materialized-view refresh or equivalent moves data | reject explicit classes and unknown/unclassified mutation classes |
| P3-08-H28 | new unsupported statement class bypasses blocklist | phase-specific allowlist; unknown statement class fails closed |
| P3-08-H29 | `DO` + dynamic `EXECUTE` hides forbidden SQL | procedural/dynamic SQL forbidden in enforced v1 |
| P3-08-H30 | concatenated dynamic SQL defeats literal scanning | dynamic execution surface rejected categorically |
| P3-08-H31 | function/procedure body carries hidden mutation | new function/procedure/DO/CALL/PREPARE execution surfaces rejected |
| P3-08-H32 | malformed/unterminated comment/string/dollar quote confuses scanner | lexical error is fatal, never "best effort" |
| P3-08-H33 | per-DDL impact requirement omitted from manifest | statement-bound typed `ddl_impact` declaration |
| P3-08-H34 | canonical observability categories omitted behind generic `signals[]` | each required category gets measured/N-A disposition with reason |
| P3-08-H35 | policy row absent from coverage matrix | exhaustive policy inventory + completeness test/record |
| P3-08-H36 | allowed branch has no positive test | bidirectional rule-to-test mapping |
| P3-08-H37 | statement impact metadata does not bind to actual SQL | normalized statement SHA-256 + exact statement-count/bijection |
| P3-08-H38 | fake precise estimate passes structure | adequacy explicitly human-reviewed; machine checks type/range/source/rationale only |
| P3-08-H39 | `SELECT` in Validate invokes side-effecting function | arbitrary SELECT is not in v1 Validate allowlist |
| P3-08-H40 | security-sensitive `COPY PROGRAM` / file IO | all COPY rejected in enforced v1 |
| P3-08-H41 | prepared statement hides dynamic execution | PREPARE/EXECUTE rejected in enforced v1 |
| P3-08-H42 | a rule is tested only negatively or only positively | coverage inventory requires both when rule has an allowed branch |
| P3-08-H43 | policy disposition says "outside scope" without owner/future gate | every non-machine row includes responsible evidence owner and closure/deferred gate |
| P3-08-H44 | N/A observability used as blanket bypass | N/A requires category-specific non-empty rationale; risk/phase rules can forbid N/A |
| P3-08-H45 | multiple SQL statements collapse into one impact declaration | one normalized executable DDL statement ↔ one impact entry |
| P3-08-H46 | semicolon/text tricks split scanner incorrectly | deterministic lexical statement splitter; malformed structure fails |
| P3-08-H47 | psql meta-command executes/includes SQL or shell outside scanner (`\\i`, `\\gexec`, `\\copy`, `\\!`) | reject psql meta-commands in enforced migration files |
| P3-08-H48 | psql variable substitution changes candidate SQL at execution | reject psql variable-substitution constructs in enforced migrations |
| P3-08-H49 | rewrite/side-effect hides inside DDL expression/default | allow only frozen safe DDL subforms; unsupported expressions/function-bearing defaults/constraints fail closed |
| P3-08-H50 | alternate PostgreSQL string syntax confuses scanner | handle standard/prefixed/dollar-quoted lexical forms deterministically or fail |
| P3-08-H51 | transaction-control variant escapes exact framing | only frozen BEGIN/COMMIT framing allowed; SAVEPOINT/CHAIN/other control rejected unless explicitly supported |
| P3-08-H52 | index online/concurrent policy omitted behind generic impact metadata | per-index online strategy disposition; non-concurrent path requires explicit structural rationale + human adequacy review |
| P3-08-H53 | policy disposition enum drift | every `S2-*` row uses exactly one frozen six-value primary disposition; secondary notes are separate fields |
| P3-08-H54 | normative Stage 2 control omitted from registry | self-contained exhaustive `S2-*` inventory + exact control-ID set test |
| P3-08-H55 | graph-valid dependency list omits a semantically required dependency | graph validity MACHINE; semantic completeness STRUCTURE_PLUS_HUMAN_ADEQUACY and explicitly reviewed |
| P3-08-H56 | allowed enum/cross-field branch has no positive test | `ALLOWED-*` registry requires at least one positive `TC-*` per allowed branch |
| P3-08-H57 | up lifecycle safety policy reused blindly for down | separate up contract and disposable-down inverse contract |
| P3-08-H58 | down drops unrelated pre-existing object | down inverse target must derive exactly from an up effect in the same migration |
| P3-08-H59 | safe inverse is falsely rejected as generic destructive DDL | scoped inverse allowlist permits only derived inverse classes |
| P3-08-H60 | canonical test ID has no canonical definition | every required `TC-*` is fully defined in this Stage 3.54 document |
| P3-08-H61 | test contract depends on rejected/unmerged artifact | no normative reference to rejected candidate blobs/archives/chat definitions |
| P3-08-H62 | "all allowed branches covered" is asserted without proof | deterministic allowed-branch registry ↔ positive-test mapping completeness |
| P3-08-H63 | authority/evidence path is later reused or content changes | new refs bind `path + content_sha256`; historical entry validation does not reinterpret current path content |
| P3-08-H64 | unqualified/quoted/search-path-dependent identifiers break object identity | enforced v1 uses schema-qualified lowercase ASCII object identifiers; search-path mutation/quoted object identity unsupported |
| P3-08-H65 | timeout metadata exists but PostgreSQL never applies it | exact timeout `SET`/`SET LOCAL` statements must precede DDL and match per-direction manifest values |
| P3-08-H66 | down execution has different timeout/lock/impact profile but only up is modeled | separate `up_execution`, `down_execution`, `up_ddl_impact`, `down_ddl_impact` |
| P3-08-H67 | disposable down file is mistaken for approved production rollback | manifest/workflow explicitly label down as disposable rehearsal inverse only; production rollback is a separate reviewed plan |
| P3-08-H68 | `IF EXISTS` masks missing/wrong rollback object | `IF EXISTS` is rejected in future enforced down inverses |
| P3-08-H69 | `CASCADE` expands rollback beyond intended effect | `CASCADE` rejected in enforced down |
| P3-08-H70 | multi-effect up has partial rollback | every reversible up effect must have exactly one down inverse |
| P3-08-H71 | down contains duplicate/extra inverse | down-effect bijection rejects duplicates/orphans |
| P3-08-H72 | down inverse order invalid | deterministic reverse dependency order required for derived effects |
| P3-08-H73 | high-risk allowed branch has no complete authority-gate positive | explicit high-risk positive with required refs/rollout/observability |
| P3-08-H74 | observability N/A policy drifts between plan and code | exact category-by-supported-phase required/N-A matrix frozen |
| P3-08-H75 | medium/high rollout gate is omitted or treated as prose | typed rollout structure + risk-specific rules + tests |
| P3-08-H76 | high-risk golden-vector/restore/security gates are claimed from strings | validator checks structured refs only; substantive evidence remains human/operational |
| P3-08-H77 | non-concurrent index targets pre-existing table without online safety | pre-existing-table index must use CONCURRENTLY; non-concurrent permitted only for table created in same up |
| P3-08-H78 | unsafe ADD COLUMN default/expression causes side effects or rewrite ambiguity | frozen literal-only safe default grammar; function/expression/generated/identity subforms rejected in v1 |
| P3-08-H79 | ADD CONSTRAINT performs blocking validation during Expand | supported CHECK/FK additions must be `NOT VALID`; other constraint subforms rejected |
| P3-08-H80 | manifest owner declaration does not match schemas touched | touched-schema set and declared owner set must match frozen current schema owners |
| P3-08-H81 | historical authority ref is revalidated against mutable current path and breaks later CI | reachability/content validation applies when entry is introduced; base-existing immutable refs are treated as historical digests |
| P3-08-H82 | unsupported manifest enum becomes accepted through data-only edit | enum sets are code/test-frozen and existing manifest entries append-only |
| P3-08-H83 | invalid UTF-8/NUL/BOM creates scanner/psql disagreement | future enforced SQL must be valid UTF-8, no NUL/BOM; byte hash remains exact |
| P3-08-H84 | Unicode homoglyph/control whitespace disguises executable grammar | executable SQL grammar/identifiers are ASCII-only outside inert literal/comment content |
| P3-08-H85 | missing final semicolon or empty-statement noise changes splitter behavior | enforced files require canonical terminated statement sequence; empty executable statements rejected |
| P3-08-H86 | timeout `SET` occurs after DDL | timeout-control ordering is machine-checked before first protected DDL |
| P3-08-H87 | timeout SQL value differs from manifest | parsed GUC values must exactly equal direction-specific manifest milliseconds |
| P3-08-H88 | arbitrary `SET`/`RESET` changes execution semantics | only frozen timeout GUC controls are allowed in enforced files |
| P3-08-H89 | risk gate does not require staged rollout for medium/high | risk-specific rollout rules are frozen and mapped to positive/negative tests |
| P3-08-H90 | Stage 2 medium/high gate evidence is silently treated as machine proof | structural refs are machine-validated; rehearsal/golden/restore/security adequacy remains named external gate |
| P3-08-H91 | Contract/Validate forward-only operations are forced into fake paired reversibility | paired-SQL v1 accepts only Expand; other lifecycle phases are explicitly rejected/deferred |
| P3-08-H92 | CREATE INDEX online decision depends on unknown table size | v1 uses deterministic stronger rule: pre-existing table => CONCURRENTLY; same-up new table may be non-concurrent |
| P3-08-H93 | one policy row mixes several primary dispositions | each `S2-*` control is atomic enough for exactly one primary disposition |
| P3-08-H94 | observability profile permits arbitrary N/A | only `rows_or_batches` and `validation_mismatches` may be N/A for supported Expand; all other categories measured |
| P3-08-H95 | risk/classification enum values are allowed but untested positively | every allowed risk/classification enum value maps to positive tests |
| P3-08-H96 | evidence ref path traversal/symlink/content substitution | safe relative regular-file path + content SHA-256 validation on introduction |
| P3-08-H97 | existing manifest evidence ref is silently altered | base-existing entry byte immutability covers refs and digests |
| P3-08-H98 | rollback inverse targets right name in wrong schema | inverse object identity includes schema + object + subobject kind/name |
| P3-08-H99 | down DML/procedural/psql execution is treated as rollback | down allowlist excludes DML, procedural/dynamic SQL and psql commands |
| P3-08-H100 | completeness is inferred from headline counts | exact set equality for control IDs, rule IDs, allowed-branch IDs and test IDs; counts are diagnostic only |
| P3-08-H101 | transactional and non-transactional UP effect classes mixed in one direction | up direction must be homogeneous; mixed ordinary DDL + concurrent-index effects reject/rescope |
| P3-08-H102 | non-transactional mode becomes a generic escape for ordinary DDL | non-transactional up is frozen to concurrent-index effects plus exact timeout controls only |
| P3-08-H103 | `IF NOT EXISTS` masks drift and makes down ownership ambiguous | reject `IF NOT EXISTS` in all future enforced up DDL |
| P3-08-H104 | CREATE TABLE special/storage/lifecycle form expands semantics | reject TEMP/TEMPORARY/UNLOGGED/ON COMMIT/TABLESPACE/partition/inheritance/LIKE/AS and unsupported options |
| P3-08-H105 | hidden object/dependency created through SERIAL/identity/generated/custom/domain/array type | frozen built-in type grammar only; generated/identity/serial/custom/domain/array forms reject |
| P3-08-H106 | function/subquery/cast hides behavior inside default or CHECK | frozen literal/default and simple CHECK grammar; functions/subqueries/casts/dynamic expressions reject |
| P3-08-H107 | FK referential action changes runtime semantics during Expand | only frozen default NO ACTION semantics; CASCADE/SET NULL/SET DEFAULT and unsupported update/delete actions reject |
| P3-08-H108 | FK deferrability/match options create unreviewed semantics | frozen NOT DEFERRABLE / default MATCH semantics only; unsupported options reject |
| P3-08-H109 | constraint/table grammar remains "safe" only by prose | exact subgrammar is rule/test frozen; unknown subform fails closed |
| P3-08-H110 | author labels constraint/index low risk to bypass medium gates | operation-derived minimum risk floor: index/constraint => at least medium |
| P3-08-H111 | machine-obvious medium-risk operation has no risk-floor regression | exact operation→minimum-risk map and negative/positive tests |
| P3-08-H112 | Go-test CI job executes migration SQL before migration validator job finishes | Go job depends on successful migrations job for exact same SHA before database apply |
| P3-08-H113 | CI job dependency exists but exact validated commit is ambiguous | migrations job publishes `validated_sha`; dependent migration-executing job asserts equality with `GITHUB_SHA` |
| P3-08-H114 | `go-race` executes candidate migration SQL before exact-SHA validator success | `go-race` is included in validation dominance with `needs: migrations` + exact `validated_sha` |
| P3-08-H115 | future/current CI SQL-execution path omitted from dominance inventory | conservative workflow inventory guard rejects migration-application markers outside frozen job IDs |
| P3-08-H116 | non-transactional DOWN omits session `lock_timeout` | full direction×mode×GUC negative matrix |
| P3-08-H117 | non-transactional DOWN omits session `statement_timeout` | full direction×mode×GUC negative matrix |
| P3-08-H118 | inline CREATE TABLE CHECK launders medium constraint risk as low table risk | CREATE TABLE is constraint-free in paired-SQL v1 |
| P3-08-H119 | inline CREATE TABLE FK/REFERENCES launders medium constraint risk | CREATE TABLE is constraint-free; FK only via ADD CONSTRAINT NOT VALID |
| P3-08-H120 | inline PRIMARY KEY/UNIQUE/EXCLUDE creates hidden index/constraint effects under low risk | those forms reject; uniqueness only via separately governed CREATE UNIQUE INDEX |
| P3-08-H121 | canonical priority rule omitted from Stage 2 registry | atomic S2 control with explicit disposition |
| P3-08-H122 | Populate active-read-path isolation omitted from Stage 2 registry | atomic rejected/deferred Populate control |
| P3-08-H123 | Validate row-count-only prohibition omitted from Stage 2 registry | atomic rejected/deferred Validate control |
| P3-08-H124 | declared metric names are overclaimed as proof that real metrics exist/usefully emit | metrics structure separated from human/operational adequacy |
| P3-08-H125 | validator log-redaction fixture is overclaimed as proof of whole production migration logging hygiene | total logging hygiene classified operational/closure; machine fixture is partial support only |
| P3-08-H126 | implementation chooses a different lock-mode enum than reviewed plan | exact eight-value enum frozen in Stage 3.54 |
| P3-08-H127 | implementation chooses a different replication-impact enum | exact four-value enum frozen |
| P3-08-H128 | implementation chooses arbitrary timeout/duration/impact integer bounds | exact numeric bounds frozen |
| P3-08-H129 | implementation broadens/narrows CREATE/ADD type allowlist without planning review | exact built-in type forms frozen |
| P3-08-H130 | implementation invents a broader CHECK/FK comparison grammar | exact atomic CHECK + FK syntax frozen |
| P3-08-H131 | implementations compute `statement_sha256` from different normalization algorithms | exact raw-statement byte-slice hashing algorithm frozen |
| P3-08-H132 | canonical Low-risk `new empty table` example omitted from source registry | atomic source row + reviewed disposition |
| P3-08-H133 | canonical Low-risk `additive nullable column` example omitted from source registry | atomic source row + reviewed disposition |
| P3-08-H134 | destructive risk examples compressed/omitted as classification anchors | atomic DROP / irreversible-conversion / history-rewrite source rows |
| P3-08-H135 | P3-08-derived hardening is mislabeled as literal Stage 2 source policy | separate `P3D-*` derived-control namespace |
| P3-08-H136 | numeric literal plus-sign acceptance differs by implementation | exact numeric lexical grammar forbids leading `+` |
| P3-08-H137 | numeric leading-zero/dot spelling differs by implementation | exact canonical integer/numeric grammar |
| P3-08-H138 | integer overflow is delegated to PostgreSQL instead of validator contract | exact smallint/integer/bigint bounds |
| P3-08-H139 | numeric(p,s) precision/scale compatibility is implementation-defined | exact precision/scale algorithm; no server rounding reliance |
| P3-08-H140 | negative-zero spelling produces multiple canonical literal identities | negative zero rejected |
| P3-08-H141 | string escaping/control-character policy is implementation-defined | exact ordinary ASCII single-quoted grammar with doubled quote only |
| P3-08-H142 | CHECK `<safe-literal>` has no literal definition/type mapping | exact per-type CHECK literal/operator grammar |
| P3-08-H143 | CHECK on a pre-existing column requires unknown static type inference | v1 CHECK limited to column added earlier in same UP with known frozen type |
| P3-08-H144 | PostgreSQL truncates >63-byte object identifier | exact 1–63 ASCII-byte identifier grammar |
| P3-08-H145 | two overlong source identifiers collide after PostgreSQL truncation | reject overlong identifier before execution + collision regression |
| P3-08-H146 | TC→R edge exists but corresponding R→TC edge is absent | exact bidirectional edge-set equality |
| P3-08-H147 | R→TC edge exists but corresponding TC→R edge is absent | exact bidirectional edge-set equality |
| P3-08-H148 | summary prose contains stale registry count/range | range/count assertions derived and linted from exact registries |
| P3-08-H149 | source and derived policy namespaces are mixed | `S2-*` source-only + `P3D-*` derived-only registry invariant |
| P3-08-H150 | derived control is missing/duplicate/unknown | exact `P3D-*` ID-set equality |
| P3-08-H151 | source extraction keeps contiguous IDs but omits a risk-table cell | atomic risk-table source anchors + designated reviewer source comparison |
| P3-08-H152 | index name omitted and PostgreSQL auto-generates identity | exact index grammar requires explicit name |
| P3-08-H153 | expression/partial/include/method/opclass/collation/sort/null index semantics leak through "simple" grammar | literal index EBNF; all extra clause families reject |
| P3-08-H154 | index key list is empty/oversized/contains duplicates | exact 1–32 distinct simple-column keys |
| P3-08-H155 | explicit index method/storage/tablespace/ONLY changes semantics | reject all non-default index clauses in v1 |
| P3-08-H156 | future implementation treats quoted/Unicode/leading-underscore identifier as equivalent | exact identifier regex; no aliases |
| P3-08-H157 | varchar default exceeds declared length | decoded payload length must be <= n |
| P3-08-H158 | CHECK operator is accepted for a type whose frozen v1 grammar does not permit it | exact operator-by-type matrix |
| P3-08-H159 | a placeholder phrase (`safe-literal`, `simple grammar`, `frozen later`) survives semantic freeze | semantic-lint forbidden-placeholder set |
| P3-08-H160 | structural meta-audit passes while literal grammar/hash/registry summaries remain unresolved | semantic-freeze meta-gate must pass before reviewer package |

| P3-08-H161 | compound canonical control is partially copied and one subject (`metrics` vs `logs`) disappears | source extraction preserves every subject/qualifier of the atomic canonical sentence |
| P3-08-H162 | `NOT NULL` constraint is syntactically hidden inside low-risk table/column form | paired-SQL v1 rejects NOT NULL; future support requires reviewed scope expansion |
| P3-08-H163 | required finite-domain field rejects unknown values but has no exhaustive valid set | every finite field must exist in §7.10 exact domain registry |
| P3-08-H164 | authority enum wording says `include` rather than `exactly` | authority kinds are exactly five; no alias/additional kind |
| P3-08-H165 | enum/domain is delegated to code/tests | semantic freeze forbids future authority direction; plan is authoritative |
| P3-08-H166 | `statement_class` manifest value differs from scanner-derived SQL class | exact statement-form→class mapping; mismatch rejects |
| P3-08-H167 | statement-class case/alias/synonym creates incompatible manifests | exact lowercase class strings only |
| P3-08-H168 | finite-domain registry omits a manifest field | exact field-name set equality meta-gate |
| P3-08-H169 | finite-domain registry member set disagrees with normative section | exact member-set equality meta-gate |
| P3-08-H170 | recognized vocabulary is confused with paired-v1 supported subset | §7.10 separates vocabulary from supported subset for phase/risk |
| P3-08-H171 | CREATE TABLE `NOT NULL` launders a constraint into Low `new empty table` | NOT NULL rejected in v1 CREATE TABLE |
| P3-08-H172 | ADD COLUMN `NOT NULL DEFAULT` launders a constraint into Low additive column | NOT NULL rejected in v1 ADD COLUMN |
| P3-08-H173 | future implementation silently adds enum alias/value | every finite field is closed-world; unlisted value rejects |
| P3-08-H174 | semantic meta-audit checks domain names but not exact members | package audit compares exact member sets |
| P3-08-H175 | test covers unknown enum rejection without testing every valid enum value | parameterized positive coverage for statement_class and authority kind plus domain-registry exactness |

| P3-08-H176 | pre-existing-table ADD COLUMN with literal DEFAULT inherits generic low floor despite stronger execution/lock semantics | nullable no-default ADD COLUMN is low; nullable literal-default ADD COLUMN has medium minimum |
| P3-08-H177 | `rollout.plan_ref` creates a second independently interpreted rollout-plan identity | field removed; typed `staged_rollout` authority ref is the sole plan identity |
| P3-08-H178 | staged rollout identity disagrees with separately validated authority path/hash | no second plan-reference field exists; strict JSON rejects `rollout.plan_ref` |
| P3-08-H179 | multiple distinct `staged_rollout` authority refs make the active rollout plan ambiguous | `rollout.mode=staged` requires exactly one `staged_rollout` authority ref |
| P3-08-H180 | standard rollout carries staged-rollout authority evidence and creates contradictory state | `rollout.mode=standard` requires zero `staged_rollout` authority refs |
| P3-08-H181 | open-domain array field has no frozen element/reference type | §7.11 exact aggregate field model and category-key reference grammar |
| P3-08-H182 | rollout metric list uses arbitrary names, duplicates or a category marked N/A | `rollout.metrics[]` is a unique non-empty subset of measured §7.6 category keys |
| P3-08-H183 | monitoring signal list drifts from canonical observability categories | `monitoring.signals[]` is a unique non-empty subset of measured §7.6 category keys |
| P3-08-H184 | procedure/verification/reason field accepts null/object/whitespace-only value | required open text is a non-null JSON string and non-empty after exact `ASCII_TRIM_BYTES` |
| P3-08-H185 | implementation applies hidden case/Unicode/content normalization to open text | validation uses trimming only for emptiness; accepted string content otherwise preserves decoded Unicode scalar sequence |
| P3-08-H186 | canonical Stage 2 normative block omitted while S2 IDs remain contiguous | byte-bound SA registry must contain every exact expected source anchor and map every active S2 exactly once |
| P3-08-H187 | source anchor line range exists but fragment bytes/hash drift | R045 recomputes SHA-256 over exact frozen Stage 2 line bytes and rejects any mismatch |
| P3-08-H188 | P3-08-derived dependency/per-direction strengthening is relabeled as canonical Stage 2 authority | source-faithful S2 wording only; graph validity, semantic adequacy and per-direction expansion live in P3D-015…017 |
| P3-08-H189 | `./docs/x.md` or `docs/./x.md` is silently normalized into an accepted authority path | authority path must already be canonical; `.` segments reject and no path cleaning may convert invalid input |
| P3-08-H190 | `docs//x.md` or trailing slash creates empty authority-path segment | empty segments, repeated separators and leading/trailing slash reject |
| P3-08-H191 | backslash or URL-style separator changes path meaning across operating systems | `/` is the only separator; backslash and URL/percent-decoding semantics reject |
| P3-08-H192 | Unicode/space/control path spelling creates cross-platform evidence identity drift | authority path grammar is ASCII `[A-Za-z0-9._-]` segments only with exact length bounds |
| P3-08-H193 | duplicate `(kind,path)` authority refs carry conflicting hashes | `(kind,path)` is unique; duplicate or conflicting-hash entries reject |
| P3-08-H194 | library Unicode TrimSpace treats U+00A0 or other Unicode space as empty | exact trim bytes are only 09/0A/0B/0C/0D/20; non-ASCII whitespace is preserved and counts as content |
| P3-08-H195 | CREATE TABLE accepts `DEFAULT literal NULL` or repeated NULL/DEFAULT due clause-order ambiguity | exact `column_def := identifier scalar_type [NULL] [DEFAULT safe_literal]` order; each optional clause at most once |
| P3-08-H196 | CREATE TABLE accepts empty list, trailing comma or duplicate column identifier | exact 1..64 `column_def` list; no trailing comma; pairwise-distinct column names |
| P3-08-H197 | ADD COLUMN omits `COLUMN`, reorders clauses or accepts trailing extra token | exact `ALTER TABLE qualified_table ADD COLUMN column_def ;` production; anything outside it rejects |
| P3-08-H198 | an S2 row is marked MACHINE although mapped rules prove only a structural prerequisite, not the stated operational/compatibility fact | exact four-way S2 evidence taxonomy plus explicit partial-evidence binding registry; paired-SQL scope rejection/non-machine remainder cannot be implied away by rule IDs |
| P3-08-H199 | an unrelated machine rule is attached to an S2 requirement merely to keep the machine column non-empty | S2→R semantic-scope registry rejects unrelated bindings; empty machine evidence is allowed when the requirement is genuinely operational/human |
| P3-08-H200 | a global semantic-freeze headline remains green while one newer enum/bound/grammar/normalization rule is outside its proof surface | exact `SEMANTIC_FREEZE_RULE_SET` equality + category mutation fixtures |
| P3-08-H201 | `numeric(01,0)`, `numeric(+1,0)`, `varchar(0005)` or equivalent spelling passes because implementation validates only numeric value bounds | exact canonical decimal-token productions for precision/scale/varchar length; signs/leading zeros/underscores/exponents reject |
| P3-08-H202 | CHECK predicate is narrow but outer `ALTER TABLE ... ADD CONSTRAINT ... CHECK (...) NOT VALID` envelope accepts extra parentheses/options/tokens | exact full CHECK statement token production; anything outside the production rejects |
| P3-08-H203 | DOWN effect identity is correct but scanner accepts alternate/multi-target/`RESTRICT`/extra-token spellings | exact literal production per supported DOWN inverse class; all optional/alternate tokens outside the production reject |
| P3-08-H204 | validator ignores a noncanonical `evil.up.sql` while frozen CI `*.up.sql` glob executes it | validator discovery includes every exact `.sql` suffix entry and rejects every noncanonical filename; current CI execution set must be a subset of validator-approved subjects |
| P3-08-H205 | pre-publication package review requires a remote branch that workflow intentionally forbids publishing before local approval | review stage is explicit: PRE_PUBLICATION_PACKAGE binds exact bytes/base/scope and remote branch is NOT_APPLICABLE; PUBLISHED_EXACT_HEAD review later requires remote ref/head/CI and cannot inherit local approval |
| P3-08-H206 | source anchor preserves a value/sequence but excludes the nearby byte that gives it mandatory/forbidden/required authority | qualifier-bearing source anchors; mutation of obligation-strength source bytes must invalidate R045 proof |
| P3-08-H207 | clean governed migration SQL is treated as proof that every external production rollback mechanism preserves financial facts | S2-109 partial evidence: SQL/DOWN subset machine-proven, actual production rollback procedure remains Operations/Architecture evidence |
| P3-08-H208 | clean migration SQL is treated as proof that separate snapshot rebuild/runtime code never mutates historical transactions | S2-118 partial evidence: governed SQL subset machine-proven, snapshot/runtime immutability remains separate evidence |
| P3-08-H209 | an existing SQL grammar owner is absent from the aggregate semantic-freeze set while the headline proof remains green | exact semantic-owner inventory partitions every R-rule and derives aggregate set mechanically; no unclassified rule |
| P3-08-H210 | a broad P3D English requirement cites a narrower rule and silently overclaims its proof scope | P3D evidence-scope registry + R055; broad claims must bind complete owner set or be narrowed |
| P3-08-H211 | FK parser validates tokens/options but omits local/reference column-count equality and relies on PostgreSQL to reject | deterministic 2→1 and 1→2 mismatch negatives plus 2→2 positive, all pre-execution under R033 |
| P3-08-H212 | FK column list admits duplicate local or referenced identifiers and defers duplicate-column rejection to PostgreSQL | exact pairwise-distinct local and referenced lists under R033 |
| P3-08-H213 | FK column-list cardinality has no project cap so scanner accepts a form the reviewed v1 language did not bound | exact `1..32` project-policy cap for each list; 33-column form rejects before PostgreSQL |
| P3-08-H214 | FK scanner/catalog responsibilities are conflated, causing either hidden server-oracle syntax or impossible lexical proof of catalog state | scanner owns exact token/list/cardinality grammar; disposable PostgreSQL rehearsal owns table/column existence, type compatibility and referenced uniqueness after lexical acceptance |
| P3-08-H215 | paired-SQL scope rejection is mislabeled as full machine proof of an external Stage 2 lifecycle/operational requirement | four-way S2 evidence taxonomy separates MACHINE_COMPLETE from PAIRED_SQL_SCOPE_REJECTED; scope rejection never claims the external requirement itself is machine-proven |
| P3-08-H216 | a P3D meta/provenance/evidence requirement is marked fully MACHINE even though its final semantic-subset/source-classification judgment is reviewer-owned | exact P3D machine-vs-structure-plus-human partition + P3D evidence-scope registry; R055 freezes structure while independent reviewer owns semantic adequacy |
| P3-08-H217 | a globally quantified migration invariant is marked complete-machine although operational/Populate migration mechanisms exist outside paired-SQL observation | GUARD-01 + R056: exact subject/observer universes; global rows cannot be complete unless machine observation covers all migration mechanisms |
| P3-08-H218 | a finite project cap is frozen in grammar but has no exact maximum PASS and adjacent-invalid NEG | GUARD-02 + R057 boundary inventory and mutation-kill proof |
| P3-08-H219 | SQL keyword and identifier lexical classes overlap and implementations choose different precedence/context behavior | GUARD-03 + R058 exact PostgreSQL 18.6 reserved set and contextual identifier policy |
| P3-08-H220 | a later finite bound/cardinality/length is added without entering the global boundary-proof inventory | GUARD-02 requires every finite semantic atom to enter the inventory before Builder PASS |
| P3-08-H221 | a material semantic rule exists only in English prose and therefore escapes the registered semantic-owner complement | GUARD-04/R059 exact semantic-atom ownership + GUARD-05 mutation survivor rejection |
| P3-08-H222 | remediation closes the Reviewer-named fixture while a sibling instance of the same generic predicate remains | GUARD-06/R061 requires generic predicate + complete sibling scan + permanent regression |
| P3-08-H223 | exact upstream bytes/hash are correct but extraction criterion proves a different semantic property than the project claims | GUARD-03/04 + R062 authority-projection proof: parser production + keyword categories + project context must compose to the claimed language |
| P3-08-H224 | a completeness checker validates a hand-authored expected set and therefore self-validates omissions | GUARD-01/02/06 + R063: discovered candidate universe must be computed from normative controls/field schema and equal the declared registry |
| P3-08-H225 | bounded integer manifest field exists but has no BND atom because expected BND count was not updated | bounded-field discovery from §§7.4/7.5/7.11 + BND exact semantic-key equality + synthetic new-field mutation |
| P3-08-H226 | migration-subject control exists but is outside the sibling registry because the expected sibling set omitted it | all-168 S2 discovery predicate + explicit discovered/deferred partition + synthetic new-global-control mutation |
| P3-08-H227 | PostgreSQL keyword category is mistaken for object-identifier admissibility without consulting `ColId` grammar | exact REL_18_6 `gram.y` ColId production plus kwlist category projection; `TYPE_FUNC_NAME_KEYWORD` and `RESERVED_KEYWORD` both reject in project ColId positions |
| P3-08-H228 | authority category is added/removed upstream while frozen project projection/hash remains internally self-consistent | exact two-source authority identity, category-set hashes, union hash and projection mutation test |
| P3-08-H229 | a new semantic candidate is introduced but all existing mutation fixtures touch only already-registered atoms | GUARD-05 candidate-injection mutations must create an unregistered bounded field / migration-subject / lexical category and force Builder failure |
| P3-08-H230 | positive completeness fixture asserts `all exact` without independently deriving the candidate universe | positive guard proof must compare discovery output to registry; expected counts are diagnostic only |
| P3-08-H231 | bounded-field discovery filters candidates through pre-known field names and therefore misses a new bounded field | parse every formal-manifest integer field generically; every non-singleton integer field must resolve to one machine-readable BOUND-SPEC |
| P3-08-H232 | normative bound prose and BND/field schema carry independently editable numeric copies | one authoritative BOUND_SPEC_REGISTRY; prose references it; BND witnesses must equal its exact lower/upper values |
| P3-08-H233 | a correct ColId production is duplicated and the checker finds the stale copy after the primary authority mutates | one anchored COLID_POLICY_SPEC in §15.3a; later sections may cite only its digest/ID, never repeat the production |
| P3-08-H234 | registered R complement is exact while a new normative semantic sentence has no R owner or TC | frozen semantic-candidate line accountability plus exact ATOM↔semantic-R bijection; new normative candidate line fails without evidence regeneration/re-review |
| P3-08-H235 | semantic atom body mutates while owner/counts stay fixed | ATOM digest is computed from the exact authoritative R semantic body and must match frozen atom evidence |
| P3-08-H236 | an S2 partial evidence edge references a valid but semantically unrelated R-rule | no-machine/reviewer-only is preferred when no genuine machine subset exists; `S2-101→R018` is removed and regression-forbidden |
| P3-08-H237 | mutation suite is described in prose but never executed against the frozen checker | packaged v20 property-mutation and extra-red-team runners execute from the frozen ZIP; property obligations are registry-derived and exact totals are independently recounted |
| P3-08-H238 | changing a normative candidate line and its nearby summary leaves structural registries green | external frozen NORMATIVE_LINE_ACCOUNTABILITY evidence is compared independently from plan counts/IDs and must mismatch |
| P3-08-H239 | a second machine-readable BOUND-SPEC authority appears outside the canonical registry and section-local proof still reports single-source | global physical occurrence scan: every `BOUND-SPEC|` row in the entire candidate must belong to the one anchored BOUND_SPEC_REGISTRY; identical or conflicting duplicates reject |
| P3-08-H240 | a second machine-readable COLID_POLICY production appears outside §15.3a while the anchored policy remains correct | global physical occurrence scan requires exactly one `COLID_POLICY|` row in the entire candidate and it must be the anchored authority |
| P3-08-H241 | duplicate ATOM declarations collapse through dictionary assignment and preserve the apparent unique-ID set | parse physical ATOM occurrences first; exact declaration count, ID multiplicity, owner multiplicity and digest validity are checked before any map construction |
| P3-08-H242 | a stale prose partition cardinality is frozen by NLA and contradicts the derived evidence registry | one machine-readable S2 evidence-partition authority is derived from canonical rows; every active numeric partition summary is compared to it, never trusted independently |
| P3-08-H243 | a valid R rule is attached to an S2 control because it is a prerequisite/mechanism rather than a logical subset | conservative evidence policy is applied to every partial edge: confirmed plus proactively identified prerequisite-only behavioral/reporting/compatibility/classification/policy rows are reviewer/external-evidence only; the full remaining partial-edge set is frozen for independent semantic re-review |
| P3-08-H244 | UP and DOWN execution metadata are each valid in isolation but implementation silently requires equality | explicit accepted cross-field branch with one migration using different valid UP/DOWN lock+statement timeout values; equality coupling must fail the positive contract |
| P3-08-H245 | a specialized Guard appears green only because NLA or candidate identity catches the same mutation first | specialized mutation fixtures regenerate NLA and isolate identity where appropriate, proving the intended global-uniqueness/bijection/acceptance guard itself fails closed |
| P3-08-H246 | duplicate authority/atom mutation keeps unique IDs and headline counts unchanged | occurrence-level counters and globally anchored authority locations are proof inputs; count-preserving duplicate+missing substitutions reject |
| P3-08-H247 | machine evidence proves only a prerequisite/declaration for an S2 behavioral or policy proposition | full partial-edge semantic pass applies the logical-subset rule uniformly; behavior/reporting/compatibility/classification/policy rows without a direct machine-observable conjunct are reviewer/external-evidence only |
| P3-08-H248 | a future edit reattaches one of the conservatively demoted prerequisite-only S2 rows under a different valid R ID | exact reviewer-only non-subset set plus family mutation regressions fail any machine-rule reattachment regardless of whether the R ID itself exists and is well-tested |



| P3-08-H249 | finite-bound proof covered registered numeric BND atoms but not all structural multiplicity/cardinality semantics | v20 retains the v19 single taxonomy: BND is quantitative-only; structural cardinality acceptance belongs to TC↔MPROP, with PROP-SPEC as compact mirrors and CARD-PROP byte-accountability only |
| P3-08-H250 | owner-level ATOM bijection hashes broad R descriptions and allows child normative property drift after NLA regeneration | v20 retains the v19 atomic TC↔MPROP machine contract; R/ATOM remain grouping/index only and broad owner closure is never a direct child-property witness |
| P3-08-H251 | exact-one/non-empty/max-one cardinality can mutate while all registered BND rows remain green | structural-cardinality discovery covers exact-one, max-one, min-one/non-empty, list min/max and 1↔1 multiplicity; regenerated NLA cannot substitute for CARD authority |
| P3-08-H252 | `NULL` / `DEFAULT` multiplicity can drift from max1 to max2 under unchanged coarse R048 title | property-level `PROP-SPEC` rows bind exact `max=1` values to R048 and explicit positive/negative witnesses |
| P3-08-H253 | CHECK one-atomic-predicate grammar can drift to two predicates under unchanged coarse R040/R033 title | property-level `PROP-SPEC` binds exact predicate cardinality `1` plus direct positive/negative witnesses |
| P3-08-H254 | touched-schema owner equality can weaken to subset while broad owner-consistency R title is unchanged | property-level `PROP-SPEC` freezes set relation `EQUAL`, owner R019 and acceptance/rejection witnesses |
| P3-08-H255 | pre-existing-table index CONCURRENTLY requirement can become optional under unchanged R029 title | property-level boolean `PROP-SPEC` freezes `required=true` for pre-existing-table index concurrency and direct witnesses |
| P3-08-H256 | exact UP↔DOWN inverse multiplicity can drift from 1↔1 to one-to-many | structural-cardinality + property-level specs freeze both directions at exactly one and bind duplicate/missing inverse witnesses |
| P3-08-H257 | `S2-107→R028` treats structural exact inverse as evidence that a schema down is demonstrably safe | S2-107 demoted to reviewer/Operations no-machine; R028 remains mandatory validator hardening but is not canonical safety evidence |
| P3-08-H258 | semantic-accountability evidence can be regenerated after a prose mutation, bypassing byte-only NLA | v20 retains v19 treatment of NLA/SEM/CARD as byte evidence only; machine behavior cannot change without an explicit canonical TC+MPROP contract change that is itself re-reviewed |
| P3-08-H259 | property registry could become another hand-picked expected list | v19 MPROP is mechanically one-to-one with the entire canonical TC set; unknown TC or orphan MPROP fails exact-set equality |
| P3-08-H260 | coarse semantic owner and atomic normative property were conflated | v19 explicitly distinguishes owner-level ATOM/R grouping from atomic TC↔MPROP machine properties; one R may own many MPROPs |
| P3-08-H261 | lexical semantic discovery can omit new obligation/relation vocabulary | v20 retains the v19 rule that machine semantics are not discovered from English tokens; the machine behavior universe is the exact canonical TC↔MPROP set |
| P3-08-H262 | regenerated byte manifests can self-authorize prose drift if they are mistaken for machine semantic authority | v19 removes that claim: byte manifests may regenerate only as evidence; any intended machine behavior change requires an explicit TC+MPROP change and a new independent review |
| P3-08-H263 | derived observer count is hand-written and contradicts canonical partition (`90` vs derived `89`) | compute observer count only from canonical S2 partition and reject every active count-bearing observer summary that differs |
| P3-08-H264 | byte-complete SEM/CARD manifests are mistaken for semantic machine-property completeness | SEM/CARD are byte-accountability only; canonical machine behavior is exactly TC↔MPROP |
| P3-08-H265 | sibling child semantics drift behind broad R ownership after evidence regeneration | every canonical TC has exactly one atomic MPROP with exact condition/outcome/owners/direct witness; broad R closure is grouping only |
| P3-08-H266 | BND is simultaneously defined as all-cardinality and quantitative-only | one global TAXONOMY-AUTHORITY defines BND/structural-cardinality/MPROP scope; competing definitions reject |
| P3-08-H267 | REVIEW_PROMPT requires a runner absent from frozen package | package-contract verifier parses every referenced v19 runner, requires presence/checksum, and clean-unzip executes all mandatory runners |
| P3-08-H268 | unknown machine behavior is introduced in prose without a canonical test/property | prose cannot create acceptance behavior; new machine behavior requires a new/changed TC plus exactly-one MPROP and a new review |
| P3-08-H269 | mutation suite is exemplar-selected and misses the next sibling | mutation obligations are generated one-to-one from every MPROP, including drift/polarity/owner/orphan attacks |

## 5. Canonical policy model and enforcement boundary

The Stage 2 migration strategy requires:

- `Expand → Populate → Switch → Validate → Contract`;
- no destructive one-step migration;
- staged authority for destructive rename/type/drop/semantic reuse;
- immutable migration history;
- compatibility during transitions;
- financial ledger history never rewritten for cleanup;
- observable/rehearsed changes;
- lock/statement timeout discipline for risky DDL;
- no binary floating-point conversion for financial values;
- domain-aware validation;
- risk classification and risk-appropriate gates;
- monotonic immutable migration IDs;
- declaration of owner, phase, dependency, reversibility, expected duration, lock risk,
  data classification, monitoring, rollback and roll-forward procedure.

The validator SHALL enforce only what repository artifacts can deterministically prove.

### Machine-enforceable

Examples:

- filename and ID grammar;
- pair/bijection rules;
- hashes;
- base-relative immutability;
- strict JSON schema;
- declared dependency graph validity;
- typed metadata presence/enum/cross-field rules;
- selected SQL safety guards;
- transaction framing;
- timeout field presence/range relationships;
- required typed pre-authority reference categories;
- exact current SQL hashes;
- executable disposable PostgreSQL rehearsal.

### Not machine-proven merely by the manifest

Examples:

- substantive ADR approval;
- substantive Security/Privacy Review quality;
- real production cardinality;
- old/new application coexistence in production;
- a fresh usable production backup;
- point-in-time restore success;
- representative SLO/performance evidence;
- observation-window completion;
- human review occurrence merely because a string says "review";
- adequacy of a chosen timeout for unknown production load.

Documentation and validator output MUST use language such as `structure validated`,
`reference declared`, or `rehearsal executed`, not `policy fully proven` when evidence cannot support
that stronger claim.

## 6. Historical SQL immutability boundary

P3-08 SHALL NOT edit any existing `000001`–`000007` `.up.sql` or `.down.sql` file.

The implementation must prove their exact byte identity using:

1. candidate SHA-256;
2. protected-base Git blob/path comparison in PR mode;
3. a frozen legacy-baseline manifest identity.

Changing old SQL and changing its checksum in the same candidate must fail.

Renaming or deleting an old SQL file must fail.

The validator must compare base-existing paths and bytes directly; it must not rely on Git rename
heuristics.

## 7. Sidecar manifest architecture — v12

Implementation SHALL introduce:

```text
infrastructure/postgres/migrations/policy_manifest.json
```

using Go standard library plus small in-repository strict-decoding/scanning helpers only.

### 7.1 Top-level structure

```json
{
  "schema_version": 1,
  "legacy_baseline": {
    "max_id": "000007",
    "policy_metadata_status": "not_retroactively_asserted",
    "entries": []
  },
  "enforced_migrations": []
}
```

Historical baseline identity and future fully enforced declarations are structurally separate.

### 7.2 Legacy baseline entry

Each legacy entry contains deterministic identity only:

```text
id
name
up_sha256
down_sha256
```

Rules:

- exactly seven entries: `000001`–`000007`;
- exact current names/hashes;
- sorted and unique;
- no future migration may enter this structure;
- no retrospective owner/duration/monitoring/rollback/rehearsal claim;
- legacy SQL is not retroactively subjected to v5 lexical/statement restrictions.

### 7.3 Future enforced migration entry

Every ID above the frozen baseline declares:

```text
id
name
up_sha256
down_sha256

owners
phase
dependencies
risk
data_classification

reversibility
up_transaction_mode
down_transaction_mode

up_execution
down_execution

up_ddl_impact
down_ddl_impact

observability
monitoring
rollout
production_rollback
roll_forward
authority_refs
```

No unknown fields are permitted.

`phase` retains the canonical enum:

```text
expand | populate | switch | validate | contract
```

but paired-SQL validator v1 accepts only `expand`. The other four values are valid vocabulary but
fail this mechanism with a stable unsupported-phase rule and require separately governed delivery
mechanisms.

### 7.4 Direction-specific execution metadata

Both `up_execution` and `down_execution` contain:

```text
expected_duration_seconds
lock_risk
lock_timeout_ms
statement_timeout_ms
```

All four values are mandatory for future enforced migrations.

The **v16 numeric authority is single-source**. The exact lower/upper values for all three integer fields are defined only by `BOUND_SPEC_REGISTRY` in §23.2c. This section does not carry an independently editable numeric copy. `statement_timeout_ms >= lock_timeout_ms` remains the separate cross-field invariant `BND-15`. `lock_risk` remains exactly `low | medium | high`. JSON nulls, floats, scientific-notation numbers, string-encoded numbers, and values outside the authoritative BOUND-SPEC reject.

All four direction-specific execution metadata fields are independent across UP and DOWN: for each field, any individually valid UP value may differ from the individually valid DOWN value while all other constraints remain satisfied. No implementation may require equality between UP and DOWN for `expected_duration_seconds`, `lock_risk`, `lock_timeout_ms`, or `statement_timeout_ms`.

Manifest timeout metadata is not evidence by itself. Section 9 binds the exact manifest millisecond
values to executable PostgreSQL `SET LOCAL` / session `SET` controls before governed DDL in every
supported direction/mode quadrant.

### 7.5 Direction-specific statement-bound DDL impact

For every executable DDL statement classified by the v1 scanner, the corresponding direction contains
exactly one impact entry:

```text
statement_sha256
statement_class
estimated_lock_mode
affected_rows_estimate
disk_impact_bytes_estimate
wal_impact_bytes_estimate
replication_impact
online_strategy
abort_condition
estimate_basis
```

`statement_class` is a closed v12 enum. It is **exactly one of**:

```text
create_table
add_column
add_check_constraint
add_foreign_key
create_index
create_unique_index
create_index_concurrently
create_unique_index_concurrently
drop_table
drop_column
drop_constraint
drop_index
drop_index_concurrently
```

The scanner derives the class from executable SQL; the manifest does not choose it freely.

| Exact executable form | Exact `statement_class` |
| --- | --- |
| supported `CREATE TABLE` | `create_table` |
| supported `ALTER TABLE ... ADD COLUMN` | `add_column` |
| supported `ALTER TABLE ... ADD CONSTRAINT ... CHECK ... NOT VALID` | `add_check_constraint` |
| supported `ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY ... NOT VALID` | `add_foreign_key` |
| supported `CREATE INDEX` | `create_index` |
| supported `CREATE UNIQUE INDEX` | `create_unique_index` |
| supported `CREATE INDEX CONCURRENTLY` | `create_index_concurrently` |
| supported `CREATE UNIQUE INDEX CONCURRENTLY` | `create_unique_index_concurrently` |
| exact disposable `DROP TABLE` inverse | `drop_table` |
| exact disposable `ALTER TABLE ... DROP COLUMN` inverse | `drop_column` |
| exact disposable `ALTER TABLE ... DROP CONSTRAINT` inverse | `drop_constraint` |
| exact disposable `DROP INDEX` inverse | `drop_index` |
| exact disposable `DROP INDEX CONCURRENTLY` inverse | `drop_index_concurrently` |

Any alias, case variant, broader/narrower class name, or manifest value that does not equal the scanner-derived
class rejects. No class may be added in Stage 3.55 without a separately reviewed Stage 3.54 scope change.

#### Frozen statement-hash algorithm

`statement_sha256` does **not** hash an implementation-defined "normalized" statement.

The v12 algorithm is:

1. decode the file as valid UTF-8 after the separate no-BOM/no-NUL checks;
2. segment top-level SQL statements at a semicolon that is outside comments, quoted strings,
   dollar-quoted strings and quoted identifiers;
3. for a DDL statement, select the exact original UTF-8 byte slice beginning at the first byte of its
   first executable token and ending **including** its terminating semicolon;
4. preserve every byte inside that slice exactly, including internal whitespace, comments, keyword
   casing and literal spelling;
5. hash those exact bytes with SHA-256;
6. encode the digest as exactly 64 lowercase hexadecimal characters.

Leading whitespace/comments before the first executable token are outside the slice. No keyword
lowercasing, comment removal, whitespace folding or numeric/string canonicalization is permitted for
the digest.

#### Frozen enums and integer bounds

`estimated_lock_mode` is exactly one of:

```text
access_share
row_share
row_exclusive
share_update_exclusive
share
share_row_exclusive
exclusive
access_exclusive
```

`replication_impact` is exactly one of:

```text
none
low
medium
high
```

The three quantitative estimates are JSON integers whose exact lower/upper values are defined only by `BOUND_SPEC_REGISTRY` in §23.2c (`BND-17…19`). This section intentionally contains no second numeric copy. Machine validation does not prove estimate realism, actual lock mode under production concurrency, actual WAL volume, or actual replication pressure.

Other rules:

- impact entries form a bijection with executable DDL statements in that direction;
- no duplicate/orphan entry;
- `abort_condition` and `estimate_basis` are non-empty UTF-8 strings after ASCII-edge trimming;
- `online_strategy` is exactly `concurrent | same_migration_new_table | not_applicable`;
- index on a table that pre-exists the same up migration requires `concurrent`;
- non-concurrent index is permitted only when the target table is created in the same up migration;
- non-index DDL uses `not_applicable`.

Machine validation proves schema, exact statement binding and deterministic cross-field rules only.

### 7.6 Exact observability profile for paired-SQL v1 Expand

Every future enforced migration contains all categories:

```text
rows_or_batches
lock_wait
statement_duration
replication_lag
wal_growth
disk_growth
validation_mismatches
retry_pause_abort_reason
change_deployment_correlation
```

Each category is:

```text
mode = measured | not_applicable
signal_or_method
reason
```

For the **only supported phase, Expand**, v12 freezes the exact mapping:

| Category | Required mode |
| --- | --- |
| `rows_or_batches` | `not_applicable` with non-empty reason |
| `lock_wait` | `measured` |
| `statement_duration` | `measured` |
| `replication_lag` | `measured` |
| `wal_growth` | `measured` |
| `disk_growth` | `measured` |
| `validation_mismatches` | `not_applicable` with non-empty reason |
| `retry_pause_abort_reason` | `measured` |
| `change_deployment_correlation` | `measured` |

`measured` requires non-empty `signal_or_method` and empty/absent N/A reason.
`not_applicable` requires an empty/absent signal and a category-specific non-empty reason.

There is no free-form risk/phase interpretation left for Stage 3.55 to invent.

### 7.7 Rollout structure

```text
rollout.mode
rollout.metrics[]
```

Rules:

- low risk: `mode=standard`; metrics remain non-empty because observability is mandatory; there MUST be zero `authority_refs[]` entries with `kind=staged_rollout`;
- medium/high risk: `mode=staged`, non-empty metrics and **exactly one** `authority_refs[]` entry with `kind=staged_rollout` are mandatory;
- `rollout.metrics[]` is a non-empty unique subset of the exact §7.6 category keys and every selected category must be `mode=measured`; arbitrary metric names/objects reject;
- `rollout.plan_ref` does not exist in v12 and is rejected by strict unknown-field handling; the typed `staged_rollout` authority object is the sole normative rollout-plan identity and its `path + content_sha256` bind content identity;
- high risk additionally requires the other authority/evidence structures frozen in §18.

### 7.8 Reversibility and production rollback semantics

`reversibility` is the frozen value:

```text
disposable_down_exact_inverse
```

for paired-SQL v1.

This does **not** mean the down file is approved for production rollback.

`production_rollback` separately declares:

```text
strategy
procedure
verification
```

Allowed production strategies for the supported additive Expand scope are:

```text
application_or_config_rollback
leave_additive_structure_unused
```

A production schema-destructive rollback is not authorized by the presence of `.down.sql`.

The down file exists for disposable rehearsal and must be an exact scoped structural inverse under §17.

### 7.9 Authority/evidence references

Each typed reference contains:

```text
kind
path
content_sha256
```

Rules for a **newly introduced** enforced entry:

- `path` is an already-canonical repository-relative ASCII path under the exact grammar below; the validator never cleans/normalizes an invalid spelling into a valid one;
- target is a regular non-symlink file resolved from the candidate repository root;
- exact SHA-256 matches candidate bytes;
- `kind` is exactly the five-value enum frozen in §18.2 and mirrored by tests; no other value or alias is permitted.

Exact authority-path lexical grammar:

```text
authority_path := segment ("/" segment)*
segment        := segment_char{1,255}
segment_char   := ASCII ALPHA | DIGIT | "." | "_" | "-"
total path     := 1..1024 ASCII bytes
```

Additional closed-world rules:

- `/` is the **only** separator;
- the path has no leading or trailing `/`;
- empty segments and repeated `//` reject;
- the exact segments `.` and `..` reject;
- backslash, colon, percent sign, spaces, ASCII control bytes and every non-ASCII code point reject because they are outside `segment_char`;
- URL/URI syntax, percent decoding, tilde expansion, environment expansion and platform-specific separator rewriting do not exist in this grammar;
- no `path.Clean`, `filepath.Clean`, Unicode normalization, case folding or dot-segment resolution may turn an invalid manifest string into an accepted identity;
- path identity is byte-for-byte, case-sensitive ASCII after JSON decoding;
- `authority_refs[]` uniqueness key is the exact pair `(kind,path)`: a repeated pair rejects whether the hash is identical or conflicting; the same path under a different `kind` is a distinct typed reference and is allowed only if all independent kind/risk gates pass.

After lexical acceptance, repository-root join/containment and regular-file/non-symlink checks are safety checks only; they do not rewrite the manifest path identity.

After an entry becomes base-existing, the complete entry including path/hash is immutable. Future
validator runs do not reinterpret the historical digest as a requirement that the same mutable path
still has identical current content; protected history preserves the evidence identity.

For rollout-plan identity, cardinality is closed-world:

- `rollout.mode=standard` => exactly zero `authority_refs[kind=staged_rollout]`;
- `rollout.mode=staged` => exactly one `authority_refs[kind=staged_rollout]`;
- two or more staged-rollout refs reject even when each path/hash is individually valid;
- `rollout.plan_ref` is not a field and therefore rejects as unknown JSON.

This makes the typed authority object the single source of truth for rollout-plan identity and avoids both evidence substitution and false future failures caused by legitimate later documentation evolution.

### 7.10 Exact finite-domain registry — closes P3-08-PLAN-17

This table is the closed-world inventory of every manifest **finite vocabulary/discriminator/version/status
field** plus the exact observability category-key set. Identity strings, repository paths, hashes, free-form
procedures/signals, numeric estimates and dependency IDs are governed by their own grammars and are not
finite vocabularies.

A required finite field absent from this table is a semantic-freeze failure. The `Vocabulary` column is
the exact recognized set; `Paired-SQL v1 supported subset` distinguishes canonical-but-unsupported
vocabulary without conflating it with an unknown value.

| Domain ID | Field/key set | Exact vocabulary/domain | Paired-SQL v1 supported subset | Normative authority |
| --- | --- | --- | --- | --- |
| `FD-001` | `schema_version` | `{1}` | `{1}` | §7.1 |
| `FD-002` | `legacy_baseline.max_id` | `{000007}` | same | §7.1–7.2 |
| `FD-003` | `legacy_baseline.policy_metadata_status` | `{not_retroactively_asserted}` | same | §7.1–7.2 |
| `FD-004` | `phase` | `{expand,populate,switch,validate,contract}` | `{expand}` | §7.3 / §15.1 |
| `FD-005` | `risk` | `{low,medium,high,destructive}` | `{low,medium,high}` | §18.1 |
| `FD-006` | `data_classification` | `{schema_only,financial,identity_personal,sensitive,mixed}` | same subject to cross-field refs | §18.3 |
| `FD-007` | `reversibility` | `{disposable_down_exact_inverse}` | same | §7.8 / §17 |
| `FD-008` | `up_transaction_mode` / `down_transaction_mode` | `{transactional,non_transactional}` | same subject to direction-specific homogeneity | §15.2 / §17.4 |
| `FD-009` | `up_execution.lock_risk` / `down_execution.lock_risk` | `{low,medium,high}` | same | §7.4 |
| `FD-010` | `up_ddl_impact[].statement_class` / `down_ddl_impact[].statement_class` | `{create_table,add_column,add_check_constraint,add_foreign_key,create_index,create_unique_index,create_index_concurrently,create_unique_index_concurrently,drop_table,drop_column,drop_constraint,drop_index,drop_index_concurrently}` | scanner-derived subset valid for that direction | §7.5 / §§15–17 |
| `FD-011` | `*.ddl_impact[].estimated_lock_mode` | `{access_share,row_share,row_exclusive,share_update_exclusive,share,share_row_exclusive,exclusive,access_exclusive}` | same | §7.5 |
| `FD-012` | `*.ddl_impact[].replication_impact` | `{none,low,medium,high}` | same | §7.5 |
| `FD-013` | `*.ddl_impact[].online_strategy` | `{concurrent,same_migration_new_table,not_applicable}` | cross-field subset by statement class | §7.5 / §15.6 |
| `FD-014` | `observability.<category>.mode` | `{measured,not_applicable}` | exact category-specific mapping in §7.6 | §7.6 |
| `FD-015` | `rollout.mode` | `{standard,staged}` | `standard` for low; `staged` for medium/high | §7.7 / §18.1 |
| `FD-016` | `production_rollback.strategy` | `{application_or_config_rollback,leave_additive_structure_unused}` | same | §7.8 / §19.3 |
| `FD-017` | `authority_refs[].kind` | `{adr,security_privacy_review,golden_vectors,restore_rehearsal,staged_rollout}` | same subject to risk/classification gates | §18.2 |
| `FD-018` | `owners[]` | `{identity,investment,analytics,audit}` | exact touched-schema set | §18.4 |
| `FD-019` | observability category-key set | `{rows_or_batches,lock_wait,statement_duration,replication_lag,wal_growth,disk_growth,validation_mismatches,retry_pause_abort_reason,change_deployment_correlation}` | exact Expand profile | §7.6 |

`FD-001…FD-019` is the exact v16 finite-domain set. No additional finite manifest domain exists outside
this registry. If Stage 3.55 needs another finite field/domain/value, implementation stops and returns to a
reviewed Stage 3.54 scope change.

Meta-validation MUST maintain exact equality between this registry and the finite-domain inventory
implemented by the strict decoder. It compares both the exact `FD-*` ID set and each exact member set
against the cited normative section. Phrases such as `includes`, `frozen in tests`, `frozen in code`,
`future value`, `implementation defined`, or any equivalent delegation are forbidden for finite domains.

### 7.11 Formal manifest field/type and open-domain model — closes PLAN-07/PLAN-16 residual

The strict decoder contract is defined by this table together with the cited detailed sections.
`required` means the JSON key must be present and non-null unless the cited mode-specific rule explicitly
permits absence. `open string` means the value is not an enum, but its JSON type and emptiness/normalization
policy are still frozen.

| Field/path | JSON type | Requiredness / cardinality | Domain / grammar authority |
| --- | --- | --- | --- |
| `schema_version` | integer | required exactly once | FD-001 |
| `legacy_baseline` | object | required exactly once | §7.2 |
| `legacy_baseline.max_id` | string | required | FD-002 |
| `legacy_baseline.policy_metadata_status` | string | required | FD-003 |
| `legacy_baseline.entries` | array<object> | required; exactly seven unique entries | §7.2 / §10 |
| `legacy_baseline.entries[].id/name/up_sha256/down_sha256` | string | all required | §10 identity/name grammar; SHA fields exactly 64 lowercase hex over exact bytes |
| `enforced_migrations` | array<object> | required; zero or more, unique IDs, strictly increasing by ID | §§7.3,10,14 |
| `enforced_migrations[].id/name/up_sha256/down_sha256` | string | all required | §10; SHA fields exactly 64 lowercase hex |
| `owners[]` | array<string> | required; non-empty set | FD-018 / §18.4 |
| `phase` | string | required | FD-004 |
| `dependencies[]` | array<string> | required; may be empty; unique migration IDs | §§10–11 |
| `risk` | string | required | FD-005 / §18.1 |
| `data_classification` | string | required | FD-006 / §18.3 |
| `reversibility` | string | required | FD-007 / §§7.8,17 |
| `up_transaction_mode`,`down_transaction_mode` | string | required | FD-008 / §§15.2,17.4 |
| `up_execution`,`down_execution` | object | required | §7.4; four exact required fields only |
| `*.expected_duration_seconds`,`*.lock_timeout_ms`,`*.statement_timeout_ms` | integer | required | §7.4 bounds/cross-field rules |
| `*.lock_risk` | string | required | FD-009 |
| `up_ddl_impact[]`,`down_ddl_impact[]` | array<object> | required; exact statement bijection | §7.5 |
| `*.ddl_impact[].statement_sha256` | string | required | §7.5 exact statement-hash algorithm |
| `*.ddl_impact[].statement_class` | string | required | FD-010 |
| `*.ddl_impact[].estimated_lock_mode` | string | required | FD-011 |
| `*.ddl_impact[].affected_rows_estimate`,`disk_impact_bytes_estimate`,`wal_impact_bytes_estimate` | integer | required | §7.5 exact bounds |
| `*.ddl_impact[].replication_impact` | string | required | FD-012 |
| `*.ddl_impact[].online_strategy` | string | required | FD-013 |
| `*.ddl_impact[].abort_condition`,`estimate_basis` | open string | required | non-empty after exact `ASCII_TRIM_BYTES`; otherwise no case/Unicode/content normalization |
| `observability` | object | required; exact nine category keys | FD-019 / §7.6 |
| `observability.<category>.mode` | string | required | FD-014 |
| `observability.<category>.signal_or_method` | open string | mode-dependent | measured=>required non-whitespace; N/A=>absent or empty string; no null |
| `observability.<category>.reason` | open string | mode-dependent | N/A=>required non-whitespace; measured=>absent or empty string; no null |
| `monitoring` | object | required | §19.1 |
| `monitoring.signals[]` | array<string> | required; non-empty unique set | each element MUST be an exact FD-019 category key whose §7.6 mode is `measured`; arbitrary names reject |
| `monitoring.success_condition`,`monitoring.abort_condition` | open string | required | non-empty after exact `ASCII_TRIM_BYTES`; otherwise no normalization |
| `rollout` | object | required | §§7.7,18.1 |
| `rollout.mode` | string | required | FD-015 |
| `rollout.metrics[]` | array<string> | required; non-empty unique set | each element MUST be an exact FD-019 category key whose §7.6 mode is `measured`; arbitrary names reject |
| `production_rollback` | object | required | §§7.8,19.3 |
| `production_rollback.strategy` | string | required | FD-016 |
| `production_rollback.procedure`,`production_rollback.verification` | open string | required | non-empty after exact `ASCII_TRIM_BYTES`; otherwise no normalization |
| `roll_forward` | object | required | §19.4 |
| `roll_forward.procedure`,`roll_forward.verification` | open string | required | non-empty after exact `ASCII_TRIM_BYTES`; otherwise no normalization |
| `authority_refs[]` | array<object> | required; may be empty subject to risk/classification/rollout cardinality | §§7.9,18.1–18.3 |
| `authority_refs[].kind` | string | required | FD-017 |
| `authority_refs[].path` | string | required | §7.9 exact canonical authority-path grammar |
| `authority_refs[].content_sha256` | string | required | exactly 64 lowercase hex matching introduced candidate bytes |

For all `open string` fields: JSON `null`, arrays, objects, numbers and booleans reject.

The exact emptiness-only trim set is:

```text
ASCII_TRIM_BYTES = {0x09,0x0A,0x0B,0x0C,0x0D,0x20}
                 = {TAB, LF, VT, FF, CR, SPACE}
```

The validator constructs a temporary emptiness view by removing only those six ASCII bytes from the
leading and trailing edges. If that temporary view is empty, a required open string rejects. The accepted
decoded string itself is preserved byte-for-byte as UTF-8: no lowercasing, Unicode normalization,
whitespace folding or content rewriting occurs. Non-ASCII whitespace such as U+00A0 is **not** trimmed
and therefore counts as content. A generic Unicode helper such as Go `strings.TrimSpace` is forbidden for
this decision because it would implement a broader language than the frozen contract.

`monitoring.signals[]` and `rollout.metrics[]` are references to the already-frozen observability category
keys, not assertions that an external production metric actually exists or emits correctly. Operational
existence/use/adequacy remains reviewer/Operations evidence.

## 8. Strict JSON integrity rules

The decoder SHALL:

- reject duplicate keys at every object nesting level;
- reject unknown fields;
- reject missing required fields;
- reject wrong JSON types;
- reject trailing JSON documents/tokens;
- use integer-safe decoding for numeric fields;
- reject `null` except where schema explicitly permits it;
- reject empty/whitespace-only required strings;
- reject duplicate values where set semantics are intended.

A simple `json.Unmarshal` into a struct without duplicate-key detection is insufficient.

`policy_manifest.json` must be a regular file, not a symlink.

## 9. Timeout execution contract — retains P3-08-PLAN-01 and remediates P3-08-PLAN-08

Timeouts are direction-specific and mandatory for every future enforced migration.

This section is the executable remediation for `P3-08-PLAN-08`: timeout metadata is invalid unless the
validated SQL direction actually applies the exact declared PostgreSQL timeout values before governed DDL.

Manifest:

```text
up_execution.lock_timeout_ms
up_execution.statement_timeout_ms
down_execution.lock_timeout_ms
down_execution.statement_timeout_ms
```

SQL must actually apply them before protected DDL.

### 9.1 Transactional direction

Canonical executable order:

```text
BEGIN;
SET LOCAL lock_timeout = '<manifest-up-or-down-lock-timeout-ms>ms';
SET LOCAL statement_timeout = '<manifest-up-or-down-statement-timeout-ms>ms';
<allowlisted DDL...>
COMMIT;
```

No protected DDL may precede the timeout controls.

### 9.2 Non-transactional direction

Used only where PostgreSQL requires it, currently the frozen concurrent-index path.

Canonical executable order:

```text
SET lock_timeout = '<manifest-direction-lock-timeout-ms>ms';
SET statement_timeout = '<manifest-direction-statement-timeout-ms>ms';
<allowlisted non-transactional DDL...>;
```

Because each migration direction is executed by a fresh `psql` process in the repository CI model,
session-scoped timeout settings do not leak to another migration invocation.

### 9.3 Exact matching rules

Validator rejects:

- missing timeout control;
- wrong GUC;
- wrong unit/value;
- timeout SQL value different from manifest;
- timeout control after first protected DDL;
- duplicate/conflicting timeout control;
- arbitrary additional `SET`/`RESET`;
- non-positive/out-of-range manifest timeout;
- statement timeout below lock timeout.

The validator proves the declared timeout values are actually applied to the SQL execution path it
validates. It still does not claim those values are operationally sufficient for production volume.

## 10. Migration identity / filename grammar

Every migration SQL filename MUST match the exact lowercase ASCII grammar:

```text
NNNNNN_name.up.sql
NNNNNN_name.down.sql
```

where:

- `NNNNNN` is exactly six ASCII digits;
- `000000` is forbidden;
- name uses lowercase ASCII alphanumeric segments separated by single underscores;
- no leading/trailing/repeated underscore;
- no spaces;
- no Unicode;
- no path separator;
- suffix casing is exact.

Validator discovery is exact and intentionally broader than the current CI execution selector:

```text
VALIDATOR_SQL_DISCOVERY_SET :=
  every direct migration-directory entry whose basename ends exactly ".sql"
```

Every member of `VALIDATOR_SQL_DISCOVERY_SET` is a governed SQL candidate. It MUST match exactly one canonical
filename production:

```text
NNNNNN_name.up.sql
NNNNNN_name.down.sql
```

or validation fails. A `.sql` entry is **never silently ignored** because it lacks the six-digit prefix or otherwise
does not "look migration-like".

Validator rejects:

- duplicate numeric IDs under different names;
- duplicate names under different IDs where identity would be ambiguous;
- case-insensitive filename collisions;
- orphan up/down files;
- every `.sql` basename outside the exact canonical UP/DOWN filename grammar, including `evil.up.sql`,
  `notes.down.sql`, `000008-bad-name.up.sql` and `123_bad.up.sql`;
- symlink/non-regular migration SQL;
- manifest/file disagreement.

The frozen current CI execution subject is:

```text
CURRENT_CI_UP_EXECUTION_SET :=
  every migration-directory entry selected by the current exact `*.up.sql` selectors in jobs
  migrations, go and go-race
```

For the frozen CI blob, `CURRENT_CI_UP_EXECUTION_SET` MUST be a subset of the validator-discovered set and every
member must have passed canonical filename/pair/manifest/SQL validation before dependent SQL execution.
`R054` checks this subject relationship against the exact CI evidence. Job ordering alone is insufficient.

IDs are strictly increasing in manifest order. Gaps are allowed because monotonic does not mean
contiguous.

## 11. Dependency graph — syntax validity vs semantic completeness

For every enforced migration, the machine validator proves **declared graph validity**:

- dependencies are unique IDs;
- self-dependency rejected;
- missing dependency rejected;
- dependency on a later ID rejected;
- dependency cycles rejected;
- dependency may reference a frozen legacy ID or an earlier enforced ID.

Gaps are allowed and dependency on the immediately preceding numeric ID is not mandatory.

### 11.1 What MACHINE proves

```text
the dependency list that was declared is structurally valid
```

### 11.2 What MACHINE does not prove

The validator does **not** claim it can prove that the author declared every semantically necessary
cross-schema/application dependency.

That completeness is:

```text
STRUCTURE_PLUS_HUMAN_ADEQUACY
```

because an omitted dependency may not be inferable from the narrow SQL forms alone.

For allowlisted DDL, the scanner MAY surface directly visible object references as review evidence, but
v5 does not turn partial reference inference into a false completeness claim.

The policy registry therefore separates:

```text
S2 dependency graph validity        → MACHINE
S2 dependency semantic completeness → STRUCTURE_PLUS_HUMAN_ADEQUACY
```

## 12. Base-relative immutability mode

The CLI SHALL have explicit modes rather than silently inferring whether base validation happened.

Preferred contract:

```text
--mode=local
--mode=pr --base-sha=<40-hex>
--mode=repository
```

### PR mode

`--mode=pr` requires:

- full 40-hex base SHA;
- SHA resolves to a Git commit;
- base commit is available locally;
- base is an ancestor of the checked candidate/PR test commit in the supported CI topology;
- base migration tree can be enumerated.

Failure of any condition is fatal.

For each base-existing migration:

- same path must remain;
- exact blob bytes must remain;
- no deletion;
- no rename;
- no modification.

If the protected base already contains a policy manifest, every base-existing manifest entry is
immutable. Candidate changes are append-only for new migrations.

For the first Stage 3.55 introduction, the base has no manifest; validator builds the permitted legacy
baseline from exact protected-base SQL identities and requires candidate legacy entries to match those
identities exactly.

### Local/repository mode

All candidate self-consistency checks run.

Output must state base-relative history comparison was **not applicable**, not "immutability proven".

## 13. CI base availability

The migration CI job SHALL make the actual PR base commit available.

The implementation may use the already-pinned `actions/checkout` with sufficient history
(e.g. `fetch-depth: 0`) or an equivalently deterministic explicit fetch. No new Action is authorized.

PR CI invokes validator with the actual pull-request base SHA.

If CI wiring omits the base or the object cannot be resolved, PR-mode validation fails closed.

Scheduled/manual execution uses documented non-PR mode and must not fabricate a PR-base comparison.

## 14. Manifest append-only semantics

Once `policy_manifest.json` is merged:

- existing legacy entries are immutable;
- existing enforced entries are immutable;
- new migration entries append only;
- `schema_version=1` is frozen for this implementation.

Changing an existing entry in a future PR fails base-relative validation even if SQL is unchanged.

A future schema-version redesign requires a separate reviewed governance/development scope; P3-08 v1
does not create a self-authorizing schema-upgrade escape hatch.

## 15. Paired-SQL v1 phase and UP statement contract

v12 deliberately narrows the future paired-SQL mechanism instead of pretending one lightweight scanner
can safely implement the whole lifecycle.

### 15.1 Supported lifecycle surface

| Canonical phase | Paired-SQL v1 |
| --- | --- |
| `expand` | SUPPORTED — narrow allowlist |
| `populate` | REJECTED — requires separately governed operational migration mechanism |
| `switch` | REJECTED — separately deployable application/config change |
| `validate` | REJECTED — validation execution/reporting is a separately governed mechanism |
| `contract` | REJECTED — destructive/forward-only removal requires separate mechanism and evidence |

This does **not** weaken the Stage 2 lifecycle. It prevents unsupported phases from being mislabeled as
ordinary paired SQL.

### 15.2 UP direction transaction-class homogeneity — closes P3-08-PLAN-09

One `up.sql` direction has one execution class.

- `up_transaction_mode=transactional`: permits only ordinary transactional allowlisted Expand DDL plus
  canonical transaction/timeout controls; `CONCURRENTLY` is forbidden.
- `up_transaction_mode=non_transactional`: permits only one or more supported
  `CREATE [UNIQUE] INDEX CONCURRENTLY` statements plus exact session timeout controls; CREATE TABLE,
  ALTER TABLE, non-concurrent index DDL and every other statement class are forbidden.
- a direction containing both classes is invalid and must be split/rescoped before implementation.

No fallback runner may silently split a single migration file into multiple transaction modes.

### 15.3 Future enforced UP allowlist

The first canonical v1 allowlist is intentionally small:

1. `CREATE TABLE` under the **constraint-free frozen grammar** below;
2. `ALTER TABLE <qualified-table> ADD COLUMN` with the frozen safe column grammar;
3. `ALTER TABLE <qualified-table> ADD CONSTRAINT` only for the separately frozen CHECK/FOREIGN KEY
   forms with `NOT VALID`;
4. `CREATE INDEX` / `CREATE UNIQUE INDEX` using the exact §15.6 simple-column index grammar;
5. `CREATE INDEX CONCURRENTLY` / `CREATE UNIQUE INDEX CONCURRENTLY` using the same exact §15.6 grammar.

Everything else fails closed until separately reviewed.

`IF NOT EXISTS` is forbidden for every future enforced CREATE/ALTER/INDEX form. A migration must fail
on unexpected pre-existing state rather than silently claim ownership of an object it did not create.

#### Exact v12 CREATE TABLE / column-definition grammar

The grammar below is a **token grammar**. SQL keyword terminals are matched ASCII-case-insensitively by
the lexical classifier; identifiers remain lowercase-only under the project identifier grammar. Ordinary
whitespace/comments may occur only between tokens as already handled by §16.1; they cannot split one token.
Statement hashing still preserves the exact original bytes and keyword casing under §7.5.

A v1 `CREATE TABLE` must match exactly:

```text
create_table :=
  CREATE TABLE qualified_table
  "(" column_def ("," column_def){0,63} ")"
  ";"

qualified_table :=
  identifier "." identifier

column_def :=
  identifier scalar_type [NULL] [DEFAULT safe_literal]

scalar_type :=
    BOOLEAN
  | SMALLINT
  | INTEGER
  | BIGINT
  | NUMERIC "(" precision "," scale ")"
  | TEXT
  | VARCHAR "(" varchar_length ")"
  | UUID
  | DATE
  | TIMESTAMP WITH TIME ZONE
```

`safe_literal` uses the exact lexical/value bounds in §15.4.

The scalar type-parameter tokens are separately frozen here and are **not** parsed by "convert to integer then
check the value" semantics:

```text
decimal_nonzero :=
  [1-9][0-9]*

decimal_zero_or_nonzero :=
  "0" | [1-9][0-9]*

precision :=
  decimal_nonzero

scale :=
  decimal_zero_or_nonzero

varchar_length :=
  decimal_nonzero
```

After lexical acceptance, the existing value bounds apply:

```text
1 <= precision <= 38
0 <= scale <= precision
1 <= varchar_length <= 10_485_760
```

Therefore:

- leading `+` or `-` rejects;
- leading zeros reject except the single token `0` for scale;
- `_`, decimal points, exponent notation and any other numeric-token decoration reject;
- empty parameters reject;
- whitespace/comments may occur only **between grammar tokens** under §16.1 and never inside one decimal token;
- `numeric(01,0)`, `numeric(+1,0)`, `numeric(1,00)`, `varchar(0005)` and `varchar(+5)` reject before PostgreSQL execution.

The CREATE TABLE column list therefore has exactly `1..64` column definitions. `GUARD-02` requires the exact boundary pair: 64 otherwise-valid pairwise-distinct columns MUST pass (TC-521) and 65 MUST reject before PostgreSQL execution (TC-522); omission of either witness invalidates the frozen bound. Column identifiers are pairwise
distinct by exact identifier bytes. Empty lists, trailing commas and duplicate column identifiers reject.

Within one `column_def`:

- `NULL` may appear at most once and only immediately after `scalar_type`;
- `DEFAULT` may appear at most once and only after the optional `NULL`;
- `DEFAULT safe_literal NULL`, repeated `NULL`, repeated `DEFAULT`, or any extra token rejects;
- absence of explicit `NULL` still means the column is nullable under PostgreSQL semantics;
- `NOT NULL` is never part of the v1 production.

The grammar is closed-world: after the final `)` only `;` is accepted for this statement. No unlisted
table/column clause is parser-deferred to PostgreSQL.

The following are forbidden inside CREATE TABLE:

```text
PRIMARY KEY
UNIQUE
CHECK
REFERENCES / FOREIGN KEY
EXCLUDE
CONSTRAINT
```

`NOT NULL` is also rejected because PostgreSQL treats it as a constraint and Stage 2 classifies a new constraint as Medium while a new empty table is a Low example.

This is deliberate. It prevents syntactic risk laundering: every CHECK/FK is introduced through the
separate `ADD CONSTRAINT ... NOT VALID` path, and every uniqueness/index effect goes through the
separate index path where the `medium` minimum-risk rule is machine-visible.

The exact allowed column type forms are:

```text
boolean
smallint
integer
bigint
numeric(p,s)  where p/scale tokens obey the exact productions above, 1 <= p <= 38 and 0 <= s <= p
text
varchar(n)    where n obeys the exact decimal_nonzero production above, 1 <= n <= 10_485_760
uuid
date
timestamp with time zone
```

No aliases are accepted (`int`, `int4`, `int8`, `decimal`, `timestamptz`, etc. reject in v1 even if
PostgreSQL treats them as aliases). Arrays, domains, enums, custom types, `money`, binary floats,
`json/jsonb`, network/geometric/range/composite types, SERIAL pseudo-types, identity/generated columns
and type modifiers outside the literal forms above reject/rescope.

CREATE TABLE also rejects:

- `TEMP` / `TEMPORARY`;
- `UNLOGGED`;
- `ON COMMIT`;
- `TABLESPACE`;
- partition/inheritance/`LIKE`/`AS`;
- storage parameters;
- procedural/function expressions;
- unsupported table options.

The optional `DEFAULT safe_literal` branch uses only the exact §15.4 literal grammar and target-type compatibility rules.

Future enforced SQL object identifiers use one exact grammar:

```text
identifier = [a-z][a-z0-9_]{0,62}
```

Therefore every identifier component is:

- project-policy maximum 63 bytes even if a custom PostgreSQL build advertises a larger `max_identifier_length`;
- ASCII only;
- lowercase and unquoted;
- 1–63 bytes exactly (ASCII bytes = characters);
- not permitted to start with `_`;
- free of `$`, Unicode, quotes or server-side truncation.

Every schema/table/column/index/constraint component is validated independently. A 64-byte identifier
rejects **before** SQL execution. Two source identifiers that would share the same first 63 PostgreSQL
bytes are both rejected by the individual length rule, so PostgreSQL truncation can never define
validator object identity.

#### 15.3a Keyword ↔ identifier lexical intersection — v16 global-single-source ColId authority

`COLID_POLICY_SPEC` below is the **only normative machine-readable ColId policy representation in the entire candidate**. Exactly one physical line beginning `COLID_POLICY|` may exist globally, and it must be this anchored §15.3a authority line. Any later Guard, summary, test, or evidence section may cite only `COLID_POLICY_SHA256`; it MUST NOT restate an independently editable production.

```text
COLID_POLICY|ref=REL_18_6|kwlist_blob=a4af3f717a1118e4b3561786c9f642c2ca5772d5|gram_blob=03c80eaaf22a74fa2a4a6b977e394d3bc34ffb46|production=IDENT,unreserved_keyword,col_name_keyword|disallowed_categories=TYPE_FUNC_NAME_KEYWORD,RESERVED_KEYWORD|type_func_count=23|reserved_count=78|disallowed_count=101|type_func_sha256=847a5d59765ccb0f3bc47a0642e6b5fa74cff47502a0059750686b5b84af953f|reserved_sha256=870e73576d611d33b419d2281b308c5b7e9592dba32d1a95180995d8b8938295|disallowed_sha256=3a9027604ec759856e3f9fdbaadaccc4588c00b213328ab5ca0018231448e0d6
```

`COLID_POLICY_SHA256 = 4782d6b70e84ec183297874379abfca4ff29f06a40987c32f15f81cd7d520c8a`.

The policy is derived from PostgreSQL `REL_18_6` exact authorities `src/backend/parser/gram.y` (Git blob `03c80eaaf22a74fa2a4a6b977e394d3bc34ffb46`) and `src/include/parser/kwlist.h` (Git blob `a4af3f717a1118e4b3561786c9f642c2ca5772d5`). The production admits lexer `IDENT`, `UNRESERVED_KEYWORD`, and `COL_NAME_KEYWORD`; therefore `TYPE_FUNC_NAME_KEYWORD` and `RESERVED_KEYWORD` are disallowed in governed unquoted table/column-style object-name positions.

The exact frozen category/union members remain:

```text
PG18_6_TYPE_FUNC_NAME_WORDS =
authorization
binary
collation
concurrently
cross
current_schema
freeze
full
ilike
inner
is
isnull
join
left
like
natural
notnull
outer
overlaps
right
similar
tablesample
verbose
```

```text
PG18_6_COLID_DISALLOWED_WORDS =
all
analyse
analyze
and
any
array
as
asc
asymmetric
authorization
binary
both
case
cast
check
collate
collation
column
concurrently
constraint
create
cross
current_catalog
current_date
current_role
current_schema
current_time
current_timestamp
current_user
default
deferrable
desc
distinct
do
else
end
except
false
fetch
for
foreign
freeze
from
full
grant
group
having
ilike
in
initially
inner
intersect
into
is
isnull
join
lateral
leading
left
like
limit
localtime
localtimestamp
natural
not
notnull
null
offset
on
only
or
order
outer
overlaps
placing
primary
references
returning
right
select
session_user
similar
some
symmetric
system_user
table
tablesample
then
to
trailing
true
union
unique
user
using
variadic
verbose
when
where
window
with
```

`update` remains an accepted contextual non-reserved witness; `table`, `select`, `user`, `collation`, `concurrently`, and `cross` remain deterministic rejection witnesses. The packaged PostgreSQL evidence must reproduce the policy hash and all category hashes.

Qualified object references are explicit where identity requires them and never depend on `search_path`.
Index names themselves are unqualified in `CREATE INDEX`; the index schema is deterministically the
parent table schema. DOWN references the derived `schema.index_name`.

Future schema creation is not supported by v1. The four currently governed schemas already exist;
adding another schema requires reviewed scope expansion.

### 15.4 Exact ADD COLUMN production and safe literal grammar

The only accepted ADD COLUMN statement production is:

```text
add_column :=
  ALTER TABLE qualified_table ADD COLUMN column_def ";"
```

where `qualified_table`, `column_def`, `scalar_type`, identifier rules and clause order are **exactly**
those frozen in §15.3.

Therefore:

- the `COLUMN` keyword is mandatory;
- exactly one column is added per ADD COLUMN statement;
- nullable column without default is accepted;
- nullable column with a frozen scalar literal server-side default is accepted;
- explicit `NULL` is optional but, when present, appears before `DEFAULT`;
- `DEFAULT ... NULL`, duplicate `NULL`, duplicate `DEFAULT`, missing `COLUMN`, comma-separated multiple
  columns, trailing tokens and every unlisted clause reject;
- `NOT NULL` rejects in paired-SQL v1 even with a literal default.

PostgreSQL 18 treats NOT NULL as a constraint, while canonical Stage 2 names only an additive **nullable**
column as the Low classification example. Future NOT NULL support therefore requires separately reviewed
scope with an explicit medium-or-higher operation-risk contract.

The default-literal grammar is literal and implementation-independent.

#### Boolean literal

```text
TRUE | FALSE
```

No lowercase/mixed-case alias is canonical in v1.

#### Integer literal

Lexically:

```text
0
-?[1-9][0-9]*
```

with these additional rules:

- leading `+` is forbidden;
- leading zeroes are forbidden except the single literal `0`;
- `-0` is forbidden;
- no decimal point, exponent, underscore, hex/octal/binary form or cast is allowed;
- parsed exact integer must be within the target PostgreSQL type:
  - `smallint`: `-32768..32767`;
  - `integer`: `-2147483648..2147483647`;
  - `bigint`: `-9223372036854775808..9223372036854775807`.

#### `numeric(p,s)` literal

Lexically:

```text
0
-?(?:[1-9][0-9]*|0)\.[0-9]+
-?[1-9][0-9]*
```

The scanner does not use floating point. It parses sign, integer digits and fraction digits as decimal
text. Canonical rules:

- leading `+` forbidden;
- integer part has no leading zero unless it is exactly `0`;
- `.5`, `5.`, exponent notation and digit separators reject;
- a negative spelling whose mathematical magnitude is zero rejects (`-0`, `-0.0`, etc.);
- when `s=0`, no decimal point is allowed;
- fractional digit count must be `<= s`;
- integer significant digit count is `0` when integer part is `0`, otherwise its digit count;
- integer significant digit count must be `<= p-s`;
- no server-side rounding/truncation is relied upon; a literal outside the declared precision/scale
  rejects before PostgreSQL execution.

#### Ordinary string literal

Only this v1 form is accepted:

```text
'payload'
```

where decoded payload consists only of printable ASCII bytes `0x20..0x7E`; backslash is forbidden; an
embedded single quote is represented only by doubled `''`. Newline, carriage return, tab, NUL,
backslash escapes, `E''`, `U&''`, dollar-quoted forms and Unicode payload bytes reject.

For `varchar(n)`, decoded payload character count must be `<= n`. `text` has no additional payload-length
policy in v1.

#### Type/default compatibility

```text
boolean                    -> boolean literal only
smallint/integer/bigint    -> integer literal only
numeric(p,s)               -> numeric literal only
text/varchar(n)            -> ordinary string literal only
uuid/date/timestamp with time zone -> no default in v1
```

Rejected additionally:

- function/expression defaults;
- casts;
- generated columns;
- identity/sequence semantics;
- subqueries;
- any `NOT NULL` in v1 CREATE TABLE or ADD COLUMN.

Machine validation proves the exact lexical/type/range contract only; business compatibility and lock
adequacy remain human reviewed with the impact declaration.

### 15.5 Safe ADD CONSTRAINT grammar

CREATE TABLE itself contains no CHECK/FK/PK/UNIQUE/EXCLUDE constraints in v1.

For a pre-existing table, v1 supports only these two `ALTER TABLE ... ADD CONSTRAINT` families and
requires `NOT VALID`.

#### CHECK

The complete accepted CHECK statement token language is:

```text
add_check_constraint :=
  ALTER TABLE qualified_table
  ADD CONSTRAINT identifier
  CHECK "(" check_predicate ")"
  NOT VALID
  ";"

check_predicate :=
    identifier IS NULL
  | identifier IS NOT NULL
  | identifier check_operator typed_check_literal
```

This is the **entire** outer envelope. `ALTER TABLE ONLY`, extra parenthesis layers, omitted/reordered `NOT VALID`,
`NO INHERIT`, `DEFERRABLE`-style additions, trailing clauses/tokens and every other PostgreSQL CHECK spelling
outside this production reject/rescope. Ordinary whitespace/comments are allowed only between tokens under §16.1.

To avoid hidden base-schema type inference, v1 CHECK may reference **only one column added earlier in
the same UP direction**. The scanner therefore knows the exact frozen column type before validating the
constraint. CHECK on a pre-existing column rejects/rescopes.

`check_predicate` is exactly one atomic predicate:

`<column>` uses the exact 1–63-byte identifier grammar and must resolve to the earlier same-UP
`ADD COLUMN` effect on the same table.

The comparison/operator matrix is exact:

```text
boolean:
  operator -> = | <>
  literal  -> TRUE | FALSE

smallint/integer/bigint:
  operator -> = | <> | < | <= | > | >=
  literal  -> exact integer literal from §15.4, including target-type range

numeric(p,s):
  operator -> = | <> | < | <= | > | >=
  literal  -> exact numeric(p,s) literal from §15.4

text/varchar(n):
  operator -> = | <>
  literal  -> exact ordinary string literal from §15.4
             and varchar decoded length <= n

uuid/date/timestamp with time zone:
  comparison CHECK unsupported in paired-SQL v1
```

`IS NULL` / `IS NOT NULL` is allowed for any same-UP frozen type.

No `AND`, `OR`, unary `NOT`, `IN`, `BETWEEN`, `LIKE`, regex, arithmetic, casts, function calls,
subqueries, qualified column references, multi-column expressions, collations or implicit
implementation-selected literal classes are accepted.

#### FOREIGN KEY

The entire accepted FOREIGN KEY statement language is:

```text
add_foreign_key :=
  ALTER TABLE qualified_table
  ADD CONSTRAINT identifier
  FOREIGN KEY "(" fk_column_list ")"
  REFERENCES qualified_table "(" fk_column_list ")"
  NOT VALID
  ";"

fk_column_list :=
  identifier ("," identifier){0,31}
```

Frozen list semantics:

- each local and referenced list therefore has exact cardinality `1..32`;
- identifiers use the exact 1–63-byte lowercase ASCII/unquoted grammar;
- local-column identifiers are pairwise distinct;
- referenced-column identifiers are pairwise distinct;
- `count(local_columns) == count(referenced_columns)` is mandatory;
- empty lists, 33+ columns, duplicate identifiers, mismatched list counts and every token outside the
  production reject **before PostgreSQL execution**.

No explicit `MATCH`, `ON DELETE`, `ON UPDATE`, `DEFERRABLE`, `INITIALLY`, `NOT DEFERRABLE`,
`ONLY`, extra parenthesis/envelope form or other option is accepted. The frozen omitted-option semantics are
PostgreSQL defaults: `MATCH SIMPLE`, `ON DELETE NO ACTION`, `ON UPDATE NO ACTION`, not deferrable.

The scanner owns the complete lexical/token/list/cardinality contract above. Catalog-dependent facts are a
separate evidence boundary: referenced table/column existence, local↔referenced type compatibility and
referenced UNIQUE/PRIMARY KEY eligibility are proven by the disposable PostgreSQL rehearsal after scanner
acceptance. PostgreSQL may reject those catalog-invalid statements, but it is never used to decide whether a
different token grammar/list shape was accepted by the scanner.

PRIMARY KEY, UNIQUE and EXCLUDE constraints are not supported through ADD CONSTRAINT in paired-SQL v1.
Uniqueness is represented only through the separately governed CREATE UNIQUE INDEX path.

Any other constraint syntax fails closed.

### 15.6 Exact v1 index grammar and online rule

The token grammar, ignoring only ordinary whitespace/comments already handled by the lexical scanner,
is exactly:

```text
CREATE [UNIQUE] INDEX [CONCURRENTLY] index_name
ON schema_name.table_name
(column_name [, column_name ...])
;
```

Frozen rules:

- `index_name`, schema/table names and every key column use the exact 1–63-byte identifier grammar;
- index name is mandatory and unqualified in CREATE; PostgreSQL places it in the parent table schema;
- key count is `1..32`; this is a project-policy cap aligned with the PostgreSQL default `max_index_keys`,
  not a license to widen automatically on a custom server build;
- key column names are simple, unqualified and pairwise distinct;
- no expression/function key;
- no `WHERE` partial predicate;
- no `INCLUDE`;
- no explicit `USING` method (v1 relies on PostgreSQL default B-tree);
- no `COLLATE`;
- no operator class or operator-class parameters;
- no `ASC` / `DESC`;
- no `NULLS FIRST` / `NULLS LAST`;
- no `NULLS DISTINCT` / `NULLS NOT DISTINCT`;
- no storage `WITH (...)`;
- no `TABLESPACE`;
- no `ONLY`;
- no omitted/auto-generated index name;
- no additional/unknown clause.

For an index whose target table existed before the current UP:

```text
CONCURRENTLY is mandatory
```

For an index on a table created earlier in the same UP:

```text
non-concurrent is permitted
```

When target table/columns were created earlier in the same UP, the effect graph checks their names
statically. For a pre-existing table/column surface, catalog existence/type correctness is proven by
the disposable PostgreSQL apply gate; the lexical scanner does not fabricate base-schema knowledge.

This exact subset is intentionally much narrower than PostgreSQL's full CREATE INDEX grammar. Any
additional index semantic requires a reviewed planning expansion.

### 15.7 Hidden data movement remains forbidden

UP rejects, at minimum:

```text
INSERT
UPDATE
DELETE
MERGE
COPY
CREATE TABLE ... AS SELECT
SELECT ... INTO
CREATE MATERIALIZED VIEW ... AS SELECT
REFRESH MATERIALIZED VIEW
```

plus all unknown/unclassified executable statement classes.

No "small seed" exception exists in v1.

### 15.8 Transaction/control statements

Outside the canonical framing/timeout controls in §§9 and 16, arbitrary transaction/session controls
are rejected.

## 16. SQL lexical, client and executable-surface safety

### 16.1 Scanner role

The implementation uses a deterministic narrow PostgreSQL lexical scanner/statement classifier, not a
full parser.

Future enforced migration SQL must be:

- valid UTF-8;
- no UTF-8 BOM;
- no NUL byte;
- executable grammar/identifiers ASCII-only outside inert string/comment content;
- canonically semicolon-terminated.

Scanner handles or fail-closes on:

- whitespace/newlines;
- `--` comments;
- nested block comments;
- ordinary and supported PostgreSQL-prefixed string forms;
- quoted lexical regions;
- dollar-quoted regions/tags;
- malformed/unterminated lexical forms;
- statement boundaries.

Historical `000001`–`000007` are legacy identities and are not retroactively subjected to these v1
grammar restrictions.

### 16.2 Procedural/dynamic/client execution boundary — retains P3-08-PLAN-03

New enforced SQL rejects:

```text
DO
CREATE FUNCTION
CREATE PROCEDURE
CALL
PREPARE
EXECUTE
```

and any frozen equivalent procedural/dynamic execution entry point.

Because repository CI executes files through `psql`, enforced SQL also rejects client-side execution or
interpretation features:

```text
\i
\ir
\gexec
\copy
\!
all other psql backslash meta-commands
psql variable substitution
```

A future legitimate need returns to separately reviewed scope.

### 16.3 Allowed control statements

Only these non-DDL executable controls are supported:

- canonical `BEGIN` / `COMMIT` where the direction is transactional;
- exact timeout `SET LOCAL` statements in transactional mode;
- exact timeout `SET` statements in the frozen non-transactional concurrent-index mode.

Arbitrary `SET`, `RESET`, `SAVEPOINT`, `ROLLBACK TO`, `COMMIT AND CHAIN`, prepared execution or
session mutation is rejected.

### 16.4 UP global safety

For future enforced up SQL:

- binary floating SQL types are rejected;
- all DML/data-moving classes are rejected;
- ledger/audit history mutation is impossible under the allowlist and independently guarded;
- local-timezone conversion constructs are rejected;
- unsupported/unknown statement or subform is rejected;
- procedural/dynamic/psql surfaces are rejected;
- timeout controls must match §9;
- phase must be `expand`.

### 16.5 DOWN global safety

Down is **not** checked with the up lifecycle allowlist. It follows the exact-inverse contract in §17.

Nevertheless down always rejects:

- DML/data movement;
- procedural/dynamic SQL;
- psql commands/substitution;
- `CASCADE`;
- `IF EXISTS`;
- arbitrary/unknown destructive statements;
- unrelated object targets;
- binary-float/timezone/data-history mutation constructs;
- unsupported transaction/session controls.

### 16.6 False-positive boundary

Forbidden words appearing only in inert comments or inert permitted literal content do not by
themselves trigger a statement-class finding.

That does not create a dynamic-SQL loophole because all dynamic/procedural execution surfaces are
independently rejected.

### 16.7 Existing structural guards

Stage 3.1 and Stage 3.11 deterministic structural assertions remain covered or are replaced by
equal/stronger generalized assertions. Silent deletion is a blocking regression.

## 17. Direction-specific DOWN inverse contract — closes P3-08-PLAN-06

The `.down.sql` file exists to support **disposable CI/rehearsal reversal**. It is not automatic
production rollback authority.

### 17.1 Derived UP effects

For every allowlisted up DDL, the scanner derives a normalized effect identity.

Supported effect classes:

```text
CREATE_TABLE(schema.table)
ADD_COLUMN(schema.table.column)
ADD_CONSTRAINT(schema.table.constraint)
CREATE_INDEX(schema.index on schema.table)
```

The normalized identity includes schema/object/subobject where applicable.

### 17.2 Allowed DOWN inverse classes and exact token productions

Only exact inverses of effects created by the same up migration are supported, and each inverse class has exactly
one accepted token production:

```text
drop_table :=
  DROP TABLE qualified_table ";"

drop_column :=
  ALTER TABLE qualified_table DROP COLUMN identifier ";"

drop_constraint :=
  ALTER TABLE qualified_table DROP CONSTRAINT identifier ";"

drop_index :=
  DROP INDEX qualified_index ";"

drop_index_concurrently :=
  DROP INDEX CONCURRENTLY qualified_index ";"

qualified_index :=
  identifier "." identifier
```

Effect mapping remains:

```text
CREATE_TABLE      → drop_table with exact_same_table
ADD_COLUMN        → drop_column with exact_same_table / exact_same_column
ADD_CONSTRAINT    → drop_constraint with exact_same_table / exact_same_constraint
CREATE_INDEX      → drop_index with exact_same_index
CREATE_INDEX_CONCURRENTLY → drop_index_concurrently with exact_same_index
```

The production is closed-world:

- `IF EXISTS` rejects;
- `CASCADE` rejects;
- explicit `RESTRICT` rejects;
- multi-target DROP rejects;
- `ONLY`, schema wildcards/broad removal and unrelated targets reject;
- any trailing/unrecognized token or alternate PostgreSQL spelling rejects;
- index identity is schema-qualified exactly through `qualified_index`;
- ordinary whitespace/comments may occur only between tokens under §16.1.

PostgreSQL execution is never used as a parser oracle to decide whether a different DOWN spelling is "equivalent".

### 17.3 Bijection and completeness

Machine validation requires:

- every derived reversible up effect has exactly one down inverse;
- every down inverse maps to exactly one up effect;
- no duplicate inverse;
- no orphan/extra inverse;
- object identity matches exactly;
- required reverse dependency order is valid.

For multi-effect migrations, partial rollback fails.

### 17.4 Down transaction modes

- ordinary inverse DDL is transactional and uses canonical BEGIN → timeout SET LOCAL → inverse DDL →
  COMMIT;
- `DROP INDEX CONCURRENTLY` is non-transactional and uses the exact timeout session-control form.

A down file with mixed transactional/non-transactional inverse requirements in one direction is
unsupported in v1. The corresponding up design must be split/rescoped before implementation rather
than inventing a complex runner.

### 17.5 Production rollback is separate

`production_rollback` describes the reviewed production strategy. For additive Expand, expected
strategies are application/config rollback and/or leaving the additive structure unused.

The existence and successful CI execution of `.down.sql` means only:

```text
the disposable structural inverse is syntactically/executably rehearsable
```

It does not mean:

```text
production data-loss rollback is approved
```

This distinction is mandatory in validator output, implementation documentation and Stage 3.56 closure
evidence.

## 18. Risk, classification, rollout and authority-reference structure

### 18.1 Risk enum

Allowed future Expand risk values:

```text
low
medium
high
```

`destructive` remains canonical vocabulary but is rejected by paired-SQL v1 because Contract is not
supported.

Risk gates:

| Risk | Machine-enforced structure | External/human adequacy |
| --- | --- | --- |
| low | review/CI path, rollback structure, observability declaration, timeout/impact metadata | ordinary review |
| medium | `rollout.mode=staged`, non-empty declared rollout metric references, exactly one typed `authority_refs[kind=staged_rollout]` | real rehearsal, real metrics existence/use and staged-rollout adequacy |
| high | all medium structures + ADR ref + security/privacy ref + golden-vectors ref + restore-rehearsal ref | actual approval/rehearsal/vector adequacy |

The validator checks only structured presence/type/path/hash and operation-derived minimum risk. It
never claims that a referenced dashboard/metric exists in production, emits correctly, is adequate, or
was actually used.

#### Operation-derived minimum risk — closes P3-08-PLAN-11

Author-declared risk may be **higher** than the machine-derived minimum, never lower.

Frozen v1 minimums:

| Machine-derived up effect | Minimum risk |
| --- | --- |
| constraint-free nullable `CREATE TABLE` | `low` |
| safe additive nullable `ADD COLUMN` **without DEFAULT** | `low` |
| safe additive nullable `ADD COLUMN` **with frozen literal DEFAULT** | `medium` |
| any supported `ADD CONSTRAINT` CHECK/FK | `medium` |
| any supported `CREATE INDEX` / `CREATE UNIQUE INDEX`, concurrent or same-up new-table | `medium` |

Because §§15.3–15.4 forbid NOT NULL/CHECK/FK/PK/UNIQUE/EXCLUDE in low table/column forms, a machine-obvious constraint/index cannot be hidden inside a syntactically low-risk CREATE TABLE or ADD COLUMN effect. A pre-existing-table ADD COLUMN with even a frozen literal DEFAULT is conservatively `medium`; only the nullable no-default form inherits the canonical Low additive-column example.

Populate/Switch/Validate/Contract are rejected by paired-SQL v1 and therefore cannot use this table to
downgrade their canonical risk.

Semantic high-risk triggers such as financial-representation redesign, identity-link semantics,
encryption or event ordering are not safely inferable from syntax alone; their classification remains
human/architecture/domain adequacy. Machine validation must never claim absence of a high-risk semantic
trigger merely because the syntactic minimum is low/medium.

### 18.2 Authority-reference kinds

For v12 paired-SQL v1, `kind` is **exactly one of**:

```text
adr
security_privacy_review
golden_vectors
restore_rehearsal
staged_rollout
```

No other value, alias, case variant or future implementation-defined kind is permitted. Stage 3.55
mirrors this enum in code/tests; it does not define or extend it.

Each new reference contains:

```text
kind
path
content_sha256
```

and follows §7.9.

### 18.3 Data classification

Frozen enum:

```text
schema_only
financial
identity_personal
sensitive
mixed
```

Minimum structural gates:

- `sensitive` and `mixed` require a security/privacy reference;
- `identity_personal` requires a security/privacy reference;
- `financial` remains subject to financial/domain review and high-risk rules when risk is high;
- `schema_only` must not be used when the touched schema/object semantics obviously require a stronger
  classification under the reviewed change.

Machine validation checks enum and required structural refs. Classification correctness remains
`STRUCTURE_PLUS_HUMAN_ADEQUACY`.

### 18.4 Owner/schema consistency

`owners` is a non-empty set drawn from the current schema-owner enum:

```text
identity
investment
analytics
audit
```

For allowlisted future SQL, the machine-derived touched-schema set must equal the declared owners set.

This proves declaration consistency, not organizational accountability beyond the repository's current
schema ownership convention.

### 18.5 No circular final evidence

Final PR CI results, Internal/External verdicts, Ready authorization and merge authorization are not
embedded as pre-CI manifest requirements.

They remain governed by `REVIEW_WORKFLOW.md` and later closure evidence.

## 19. Monitoring, observability, production rollback and post-execution evidence

### 19.1 Monitoring plan

Each enforced migration declares:

```text
monitoring.signals[]
monitoring.success_condition
monitoring.abort_condition
```

`monitoring.signals[]` must be a non-empty unique subset of the exact §7.6 category-key set and every selected category must be `mode=measured`. `monitoring.success_condition` and `monitoring.abort_condition` follow the open-string contract in §7.11.

### 19.2 Observability

The exact paired-SQL v1 Expand mapping is frozen in §7.6.

Stage 3.55 is not allowed to invent additional `not_applicable` categories.

### 19.3 Production rollback

```text
production_rollback.strategy
production_rollback.procedure
production_rollback.verification
```

The plan is distinct from disposable `.down.sql`.

### 19.4 Roll-forward

```text
roll_forward.procedure
roll_forward.verification
```

### 19.5 Signed validation report

The canonical signed validation-report requirement is post-execution/closure evidence, not a
self-authored pre-execution manifest assertion.

Where applicable it identifies:

- dataset/watermark;
- queries/tool versions;
- mismatches;
- accepted risk;
- reviewer/sign-off evidence.

An unexplained financial mismatch blocks the later Contract mechanism. Paired-SQL v1 does not implement
Contract.

### 19.6 Production-only evidence boundaries

The following are never declared machine-proven by the migration validator:

- actual production backup/PITR freshness/restorability;
- representative production volume;
- old/new application coexistence;
- observation-window completion;
- actual staged-rollout execution;
- actual Security/Privacy approval;
- golden-vector adequacy;
- actual restore rehearsal;
- least-privilege production principal identity.

Their `S2-*` rows name the required external owner/gate.

## 20. Data-classification semantics

The exact v1 enum and structural gates are defined in §18.3.

Additional rules:

- classification is mandatory for every future enforced migration;
- unknown value rejected;
- value cannot be empty or null;
- classification is immutable after merge with the manifest entry;
- touched-schema/declared-owner consistency is machine-checked;
- reviewer verifies whether the chosen classification is semantically strong enough.

The validator never downgrades a classification automatically to make a migration pass.

## 21. Stable validation error system

Stable categories:

```text
MIG001_DISCOVERY
MIG002_FILENAME
MIG003_PAIRING
MIG004_MANIFEST_JSON
MIG005_MANIFEST_BIJECTION
MIG006_HASH
MIG007_BASE_CONTEXT
MIG008_BASE_IMMUTABILITY
MIG009_METADATA
MIG010_DEPENDENCY
MIG011_SQL_LEXICAL
MIG012_STATEMENT_CLASS
MIG013_SQL_SAFETY
MIG014_TRANSACTION
MIG015_TIMEOUT
MIG016_AUTHORITY_REFERENCE
MIG017_LEGACY_BASELINE
MIG018_DDL_IMPACT
MIG019_OBSERVABILITY
MIG020_REHEARSAL
MIG021_POLICY_COVERAGE
MIG022_DOWN_INVERSE
MIG023_ROLLOUT
MIG024_OWNER_CLASSIFICATION
MIG025_TEST_CONTRACT
MIG026_SEMANTIC_FREEZE
MIG027_MAPPING_INTEGRITY
```

Errors carry where applicable:

```text
code
rule
migration_id
direction
path
statement_index
effect_id
control_id
detail
```

No SQL row content, credentials, tokens, personal data or financial-document content is logged.

Tests assert stable code/rule/context, not brittle full prose.

## 22. Existing CI evidence to preserve and strengthen

Current required `PostgreSQL migration validation` evidence remains:

1. validator execution;
2. full up apply on disposable PostgreSQL 18;
3. reverse down rehearsal;
4. full reapply;
5. selected schema/provenance invariants;
6. append-only runtime privilege validation.

Stage 3.55 strengthens this without creating an eleventh required check.

### 22.1 Validation dominance across every current CI SQL-execution path — closes P3-08-PLAN-12

The frozen base workflow has **three** jobs that execute repository migration SQL:

```text
job id       display name                         migration SQL behavior
migrations   PostgreSQL migration validation      validates, then applies/down/reapplies
go           Go tests                             applies every *.up.sql
go-race      Go race tests                        applies every *.up.sql
```

`go` and `go-race` can currently run in parallel with `migrations`; both must therefore be dominated.

Stage 3.55 SHALL change the existing job graph without adding an eleventh required check:

1. `migrations` remains the authoritative migration gate.
2. Its validator step runs before **any** SQL apply in that job.
3. After exact-SHA validator success, `migrations` publishes `validated_sha == GITHUB_SHA`.
4. Both `go` **and** `go-race` declare `needs: migrations`.
5. Before each job's own `Apply PostgreSQL migrations` step, it asserts
   `needs.migrations.outputs.validated_sha == GITHUB_SHA`.
6. If `migrations` fails, is skipped, validates another SHA, or does not publish the output, neither
   `go` nor `go-race` may execute repository migration SQL.

#### Frozen CI SQL-execution inventory guard

Stage 3.55 tests also freeze the current SQL-executing job inventory as:

```text
migrations
go
go-race
```

The CI-policy test inspects `.github/workflows/ci.yml` and fails if:

- any of these three job IDs disappears/renames without a reviewed inventory update;
- `go` or `go-race` loses `needs: migrations`;
- either dependent job lacks the exact-SHA assertion before its migration-apply step;
- the workflow contains a migration-application marker outside the frozen inventory.

For this v1 guard, a **migration-application marker** is any workflow run-script line containing the
repository path token `infrastructure/postgres/migrations` together with either `.up.sql`, `.down.sql`
or a `psql` invocation in that same job. The guard is intentionally conservative and may fail closed
on ambiguous workflow edits.

This static inventory rule proves the reviewed current workflow shapes and catches obvious future
migration-apply paths. It does **not** claim semantic detection of arbitrary future shell/download/code
that somehow executes migrations without those markers; workflow review remains responsible for such
scope changes.

This is an execution-before-use guarantee for the three current repository CI SQL paths, not
production deployment authorization.

### 22.2 PR-mode base context

For pull requests:

- actual PR base SHA is passed to validator;
- base Git object/history required for comparison is fetched;
- missing/unresolvable/non-ancestor base is fatal.

### 22.3 Exact timeout execution

Future enforced migrations must contain timeout SQL matching manifest (§9). CI executes the exact bytes,
so the timeout controls validated are the controls PostgreSQL receives.

### 22.4 Rollback/reapply postconditions

CI additionally establishes:

- first full-apply catalog fingerprint/invariant set;
- reverse down execution to expected disposable baseline;
- explicit managed-object baseline assertion;
- full reapply;
- equality of first/reapplied deterministic catalog fingerprint/invariant set.

If text-based schema fingerprinting is unstable, use normalized catalog queries/invariants rather than
claiming `pg_dump` text equality.

### 22.5 Down-role wording

CI output and Stage 3.55 documentation must call `.down.sql`:

```text
disposable rollback/rehearsal inverse
```

and must not call successful down execution proof of approved production rollback.

### 22.6 Required check inventory

The protected repository remains at exactly ten required checks. P3-08 strengthens the existing
`PostgreSQL migration validation` job; it does not create a new required-check name.

## 23. Exhaustive canonical Stage 2 control registry — closes P3-08-PLAN-04

v12 retains the stable atomic source IDs and the v11 reviewer-confirmed source-wording/source-derived separation, while tightening source-authority bytes and evidence scope.

Exactly six primary dispositions are legal:

```text
MACHINE
STRUCTURE_PLUS_HUMAN_ADEQUACY
OPERATIONAL_OR_CLOSURE_EVIDENCE
REJECTED_BY_PAIRED_SQL_V1
DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE
HISTORICAL_ENTRY_CRITERION_ONLY
```

A row has **one and only one** primary disposition. Clarifying notes, machine-rule mappings and
external evidence gates are separate columns and can never be parsed as alternate dispositions.

Dependency graph validity and semantic completeness are deliberately separate controls.

### 23.0 Byte-bound canonical Stage 2 source-anchor registry — closes PLAN-04/PLAN-16 proof gap

Source authority for this registry is exactly:

```text
docs/database/MIGRATION_STRATEGY_STAGE_02.md
blob d4656e2bb124fe6ff0783e619eaf608ed1082297
UTF-8/LF bytes exactly as stored in the protected base
```

An `SA-*` fragment hash is SHA-256 over the exact UTF-8 bytes of the inclusive canonical source line
range, including each stored terminal LF. The line ranges below are the complete reviewed set of
normative Stage 2 semantic blocks for P3-08. Headings, blank lines, Markdown fences/table headers and
document metadata are intentionally not normative anchors. To prevent that exclusion boundary from
becoming another hand-waved completeness gap, the exact non-blank canonical line set that is **not**
covered by an `SA-*` range is frozen as:

```text
SOURCE_NON_NORMATIVE_NONBLANK_LINES =
{1,3,4,5,6,7,8,9,10,11,12,14,21,25,38,53,68,81,98,113,115,116,122,132,142,155,168,182,189}
```

Those 29 lines are limited to the document title/metadata table, section headings, Markdown fence/table
syntax, and non-normative Purpose context. The metadata status `Proposed Strategy / No Migrations Authorized` is
document-state context, not a future migration-validator control. The obligation-strength bytes `The mandatory lifecycle is:`
and the lifecycle sequence are one authority unit and are byte-bound together by `SA-001` across lines 16–20. Any
future non-blank canonical source line outside the union of the `SA-*` coverage and that exact exclusion
set fails source-accountability proof. The Builder gate therefore proves all **149/149 non-blank source
lines are accounted** as either normative anchor coverage or an explicitly reviewed non-normative line; **120**
non-blank lines are anchor-covered and exactly **29** are explicitly excluded.

Every `SA-*` row MUST satisfy all of these:

1. exact line range equals the frozen v12 anchor declaration;
2. exact fragment SHA-256 recomputes from the frozen Stage 2 bytes;
3. every mapped `S2-*` exists;
4. every active `S2-001…S2-168` appears in exactly one source anchor;
5. no source anchor has zero accountable `S2-*` controls;
6. `S2-*` requirement wording may split an anchor into atomic conjuncts but may not add semantics absent
   from that anchor; any stronger P3-08 rule belongs in `P3D-*`;
7. changing the canonical source blob, anchor range, fragment hash or accountability mapping requires a
   separately reviewed Stage 3.54 source-extraction update; Stage 3.55 may not infer a replacement.

| Source anchor | Canonical source lines | Exact fragment SHA-256 | Accountable source controls | Semantic unit |
| --- | ---: | --- | --- | --- |
| `SA-001` | `16-20` | `af71bfb60ef731fbd65561164f9b9b44de6e4d617b5fa1fae5a2401bf6b40ecf` | `S2-001` | purpose/scope + mandatory lifecycle authority + sequence |
| `SA-002` | `23-23` | `b634a0e2137e1fca1bad6a744196013dfb1918df46dca14be544d89bf66498e3` | `S2-162` | global priority |
| `SA-003` | `27-27` | `c1de8ef73c576a04ed9a1796ccebd5a47a259349f6bf18b2f758f13348b0ac59` | `S2-002` | non-negotiable 1 |
| `SA-004` | `28-28` | `20d51d7ff9c255d8aa8209218baf9b3c84c8e68a94f2c85503b06b3a92099061` | `S2-003` | non-negotiable 2 |
| `SA-005` | `29-29` | `ab4c42bea65ac3ae6c5f63ee9720ced155e41dfe48b130c7983230e7ed83a6ad` | `S2-004` | non-negotiable 3 |
| `SA-006` | `30-30` | `4e9ad570fdd58d9e49cb66eb630f2db7d89591cf4d827f055dd5121d4362ea0a` | `S2-005` | non-negotiable 4 |
| `SA-007` | `31-31` | `312c2d20466837aded4342f9fbfe4fb3f396b34ba26152c2e6ba4a3bcd87e792` | `S2-006` | non-negotiable 5 |
| `SA-008` | `32-32` | `78bc06e7c7b7ee8a421a29a6e2722be3277885a5781c133969300c923fa2cbab` | `S2-007,S2-008,S2-009,S2-010,S2-011` | non-negotiable 6 conjuncts |
| `SA-009` | `33-33` | `d321f6c6b32ed9e9982610a39a635c27d718239891826493fd572782a5c7ec8f` | `S2-012,S2-013,S2-166` | non-negotiable 7 conjuncts |
| `SA-010` | `34-35` | `8461d71fbb9b51384cd7fea97d1004f5f509cc912cd13a646f2c36a0806a5f1e` | `S2-014,S2-015` | non-negotiable 8 conjuncts |
| `SA-011` | `36-36` | `3e8f5e48953247375cec88c5d2802d9e0536dc947c7e050863873e8a3124cba0` | `S2-016` | non-negotiable 9 |
| `SA-012` | `40-40` | `49fda574768e9dd3ad64ee8f55d173e20cad15b89661522791aa71e386648be9` | `S2-017` | Expand backward-compatible structures |
| `SA-013` | `42-42` | `1de882d1584b3475e2710ddea5a5f7ba2709f9e395167391741a749d0174c262` | `S2-018` | nullable/default column |
| `SA-014` | `43-43` | `2187ba84de03d7cec86545bf8dd20865f0c318c1c2feb07aeaaed7278108502f` | `S2-019,S2-020,S2-021` | table/schema/index ignored by old versions |
| `SA-015` | `44-44` | `2cfe3d19c41a26ea47518c84d45f054ddd1b1075a9ae61a7f715e4afafdfd562` | `S2-022` | enum/check compatibility |
| `SA-016` | `45-45` | `c66398552b7fcda8b1b2577198801384ef171c349df995635c6ca327b39ef1d8` | `S2-023` | read model/event version coexistence |
| `SA-017` | `46-46` | `ebc0764903a06b898590e93ed5d265f2454029d349f635d0bb42b4d76738313a` | `S2-024` | additive API unknown-field tolerance |
| `SA-018` | `48-49` | `12a6d95e92ecbb22214cd4655d7fbb6758c1b7fd0fb25fc4b3dac9275abeef87` | `S2-025,S2-026,S2-027,S2-028` | large-index online/transaction + explicit timeouts |
| `SA-019` | `50-51` | `8cde3a2cdf07174c18548112599751ca7c6d776618562ae98582508208de15b1` | `S2-029,S2-030,S2-031,S2-032,S2-033` | DDL impact declarations |
| `SA-020` | `55-55` | `fb35c248b68f443d3ef6917a1c53159df08d625064f62d9f6cc22f23705d52dd` | `S2-163` | backfill no active-read-path change |
| `SA-021` | `57-57` | `a637c6fc8fbaa99a9592e4cbb8929a5e493e8e13b2d6a6e6a9ca6df7a47ef1b4` | `S2-034` | resumable idempotent stable-PK batches |
| `SA-022` | `58-58` | `f9589063fed6157591ccce75f1f4116bd9f398fc50c007fc4f468567007cf9db` | `S2-035` | bounded tx/rate limits |
| `SA-023` | `59-59` | `090cf66e657df7fc07d9e16bcb876f11f2b5959f32b346e9c02bd7c73d6f5fa1` | `S2-036` | separate progress/watermark |
| `SA-024` | `60-60` | `ee5f96ecfe7c41f231616a66ee14c3e625974889550c1de35ffc7d8f7582767b` | `S2-037` | continuous checksums/counts/invariants |
| `SA-025` | `61-61` | `2f6b46e98c185883344211a9034479f85504def1af2778b71617058d117aed40` | `S2-038` | pause/resume no duplicate effects |
| `SA-026` | `62-62` | `c90c897cdd02423cdf832b26a7f6c3644ea78f5f66db8fd60f8fa439e5c91584` | `S2-039` | no binary float |
| `SA-027` | `63-63` | `75f08bb5297f80e848bfaf66a9b6698245373209b7195cd73d202c0de322efe1` | `S2-040` | no local timezone conversion |
| `SA-028` | `65-66` | `303346cae81e2f92849cf5822ab301c39b949659191fb9113ce3c4f74e16455c` | `S2-041,S2-042,S2-093` | operational jobs/business worker/removal/approved ongoing exception |
| `SA-029` | `70-70` | `4e1ab3d584202c85039e77958271a5e6cf8f097f707245f87187f1fe6b8c970c` | `S2-043` | separately deployable switch app change |
| `SA-030` | `72-72` | `167b82bb5e8ce0c1f4800de8cf8c89771f86803afc33831a9bbf5a57d0068ef1` | `S2-044` | read both reps |
| `SA-031` | `73-73` | `2cdec9abefe70f765c2564137ac35c90fca88d7d457f2c544d1003cc7029069a` | `S2-045` | shadow reads where privacy permits |
| `SA-032` | `74-74` | `fbb4f3eb8e42972a54cb0395c490095d387aa02ec6bb3330b3e37c6eed858e7c` | `S2-046` | narrow feature/config switch |
| `SA-033` | `75-75` | `956a6ccb75d5abd0280859931ac9bc8b8f16cccb2c75a21509f384a7c3f8c667` | `S2-047` | retain old write/read rollback path |
| `SA-034` | `76-76` | `ebcf7c8e68e1cbe87c9c2c208b325497e1afe6147f3346651fa4c1374b78106e` | `S2-048,S2-049` | avoid dual writes; if unavoidable define ordering/recovery/source |
| `SA-035` | `78-78` | `7bde3b9fdc898dc163d7530d74846465e0f5ac27d6cf3861fa54e1c911482044` | `S2-050,S2-051` | canonical source explicit + cache invalidation |
| `SA-036` | `79-79` | `b0aea811bbe60c1b288e76bc2c671f52d77d35b28a567f33711a9fd8bc97d90e` | `S2-052` | OpenAPI breaking switch versioned/not DB migration |
| `SA-037` | `83-83` | `c6d41f63026b460f0c41465310173f981da94afb6ea99e64b11220067485c2cc` | `S2-164` | domain-aware not row-count-only |
| `SA-038` | `85-85` | `62e723f6512e5bfef6198b4437ffec8feaf85851981223c8ccad3852fdd1b33c` | `S2-053` | referential/uniqueness |
| `SA-039` | `86-86` | `152dcd80999f368329a690f3f17ce571ed985768a74b5c5bd40687c960420aa5` | `S2-054` | decimal scale/half-even |
| `SA-040` | `87-87` | `68e7133936ec1f7a2f2efec1f9b617dc8ea93f9360ae18acbd150348fc16fbba` | `S2-055` | revision/reversal |
| `SA-041` | `88-88` | `d7b6f953e2abde38e73ce6ae5d14805dd9fa9cc7495d8c57179e687d5448a13c` | `S2-056` | BusinessDate no UTC drift |
| `SA-042` | `89-89` | `49d86860cb16f20a98be4a50e39a6c497c4cef1a3997fec8fec4e7a7f33d22ad` | `S2-057` | snapshot equality canonical ledger + methodology version |
| `SA-043` | `90-90` | `7f2853c6488ce10034fa124dfb6189b08432a2a328ae809e2595a36fa1463fd4` | `S2-058` | identity/investment separation + absence personal data |
| `SA-044` | `91-91` | `3da5f081faa33388ae7c726e97382e058c13073e7dd3e70195ee68fba75e595b` | `S2-059` | outbox/inbox + business version |
| `SA-045` | `92-92` | `7a1daebc88f04717e3fdeb4ab2c847dc2a668edeb4e6db2a1f554c7f2ec5a9db` | `S2-060` | query plans/SLO representative volume |
| `SA-046` | `93-93` | `cfd68b853caa72bbfb65bbc84f9f09cc547dbf235320cd1a892f082230010ae1` | `S2-061` | restore+replay nonprod |
| `SA-047` | `95-96` | `e90e11680b64f997790986f462c076a119a7767890964163810ac944ba7d0683` | `S2-062,S2-063,S2-064,S2-065,S2-066,S2-067` | signed validation report + unexplained mismatch |
| `SA-048` | `100-108` | `cdb3713dc9ff02c69f060f215183a0c722126457220a4df2be8f70c369842131` | `S2-068,S2-069,S2-070,S2-071,S2-072,S2-073,S2-074` | Contract preconditions |
| `SA-049` | `110-111` | `102e41efdde81ccbde551e9b3b342de958e40dd2969f03ef750979c3057449fb` | `S2-075,S2-076` | own PR/deployment + no bundling + DROP not default |
| `SA-050` | `117-117` | `baf529a017ed01878fd8be6e7180f8dc2727b9b9b04879540da8e6d9c356e37b` | `S2-077,S2-078,S2-079,S2-149,S2-150` | Low examples/gates |
| `SA-051` | `118-118` | `97cdc6c2e4f2ec40ce3cf26d478c911065983c3025186f0257cb73d2b7564d52` | `S2-080,S2-081,S2-082,S2-151,S2-152,S2-153,S2-154` | Medium examples/gates |
| `SA-052` | `119-119` | `197d5d6c0c6ea4fa4a315bd4bcf045f2e2a2aeafaef3c431697d09c917d3eec7` | `S2-083,S2-084,S2-085,S2-086,S2-155,S2-156,S2-157,S2-158` | High examples/gates |
| `SA-053` | `120-120` | `5c5b658f1c0afef7c4b9506fa166372da5ab3d17452c34294689d27d2a6d0aaf` | `S2-087,S2-159,S2-160,S2-161` | Destructive examples/gate |
| `SA-054` | `124-124` | `0377fdda4e2d3d95f2f3c3fb672a8c0d8d633defacd749abbde0dd13a6679ece` | `S2-088,S2-089` | IDs monotonic/immutable |
| `SA-055` | `125-126` | `47cb10af9e12bc33585b550d52281f3ea534e6af9a89a7a8519f744403d47296` | `S2-090,S2-091,S2-092,S2-094,S2-095,S2-096,S2-097,S2-098,S2-099,S2-100` | declarations |
| `SA-056` | `127-127` | `e0eac76f0d6060fd885e846e47ebdb8bdbc9e08d71dcbe6a61fd01f48c17f144` | `S2-101,S2-102` | independent contexts + cross-schema deps explicit |
| `SA-057` | `128-129` | `bca98295e875998dbde58e31b0adedf6dc841e2829aef7f476a4ee5c163e1115` | `S2-103` | compatible range not latest |
| `SA-058` | `130-130` | `b0f012a114f82d421293359e02d531efd7252ca13a804b64ed823c03bacd337e` | `S2-104,S2-105,S2-167` | failed migration stop/diagnosis/not manually successful |
| `SA-059` | `134-134` | `9054108e441da0eb1feecee468ae24dc0c794ce10d8e52c5c5ca509b8d659ddf` | `S2-106` | prefer app/config rollback |
| `SA-060` | `134-135` | `9ef7b647ce108e4df0982d472596c448a5516ac1fd426754f9362af3854ab3eb` | `S2-107` | down only demonstrably safe |
| `SA-061` | `135-136` | `93c94ffabb75290c2e0cda9163b941e80592590e0cc9b4d60a05c3b68014b9c0` | `S2-108,S2-109` | populated data may remain + no delete financial facts |
| `SA-062` | `138-140` | `a041c5f1be34e142a982a158206b9e4dd2400518eddf3f3d0890579775698375` | `S2-110,S2-111` | external effects compensation + at least-once protections |
| `SA-063` | `144-144` | `b752094004c96c77031c070aba59f7d74770d71fe8a460a8b4ce705062e91fbe` | `S2-112` | snapshot change new version |
| `SA-064` | `146-146` | `c47f420e34a504e3cba3b132b58558f5f765af5155e9a67a9fa7270124103b09` | `S2-165` | expand new representation/version |
| `SA-065` | `147-147` | `ff208156c4db822b5b407c44fe4d0e507ebad6cabd3cc86bcde0393424aa10b4` | `S2-113` | rebuild immutable tx + market/inflation |
| `SA-066` | `148-148` | `beaae69fa69cea9811ec0055bff0b944f3a04d74b49d5a132939b7606b9e9cd8` | `S2-114` | golden vectors/prior version explained deltas |
| `SA-067` | `149-149` | `98da21c35e396fc8d1ffbd5b12e73942c2aa2e18f1f5e4c3ec97ba95a57f1fd7` | `S2-115` | switch methodology/version |
| `SA-068` | `150-150` | `785f09d3e3e2385c0c1aecdf593647b21fc16b2868133e1983b0fbfa19946f6e` | `S2-116` | retain prior rollback window |
| `SA-069` | `151-151` | `fd57ca76f20f6f95764f8d1a7fc103e1f3ae286212c455eeb06321782ff5c7ba` | `S2-117` | contract retention/audit |
| `SA-070` | `153-153` | `e59533bc8ba2002020d451d9e66eb2ea5f071e6f56a17fa8a8966c9003d41da6` | `S2-118` | never mutate history |
| `SA-071` | `157-157` | `8be554a00500df14acdd98c589ffd40d4df9ff5f287382850af35b9a09bba102` | `S2-119` | irreversible after older backup restore |
| `SA-072` | `158-158` | `b98f887d2540d6a2da09cc7a551bf61906dac51c5f54c2cf8701482fd990871a` | `S2-120` | revocable per-subject encryption-key |
| `SA-073` | `159-160` | `87049dd46b1260a99a0c650d025116bb2bed7632307006e7a81933e287300db5` | `S2-121` | delete live rows/mapping + cryptographically destroy key |
| `SA-074` | `160-161` | `61d0de68a6debf510942290ab364dffc79434d273351746d190e09d7d25ea8dc` | `S2-122` | restored backup cannot recreate link |
| `SA-075` | `161-162` | `8ec351a8c489c2663bbe7e7c7dc7861b0584ce27f987e2ec09c702101a390d89` | `S2-123,S2-124,S2-125` | replay deletion ledger + backups encrypted/expire 90d |
| `SA-076` | `164-165` | `a528f92550b54a36bcd545f4d2dcd1388a92f0bc59d4b8248eafe4b7bf0f8962` | `S2-126` | exact key hierarchy/Vault/deletion ledger/runbook Security Review |
| `SA-077` | `165-166` | `a91716f65b4a7b54dd8df4955fd49432260b90f5f3f3e732386f18306924cc8a` | `S2-127` | operator reconstruction = pseudonymization/fails anonymization |
| `SA-078` | `170-177` | `c4790cd00dd14ec9cb2ae58f037a4957f98b65199085fe07ce4ba5c45d00c3b6` | `S2-128,S2-129,S2-130,S2-131,S2-132,S2-133,S2-134,S2-135,S2-136` | production migration reports fields |
| `SA-079` | `179-180` | `7310a6de43dd13e645077423383dda97f6938b53556e1d3520bdf933944cec27` | `S2-137` | metrics and logs sensitive-content prohibition |
| `SA-080` | `184-186` | `a5e03f55c2f195bf06c3dac2db9c4d104b2b068583d2aa11cff7efb661570135` | `S2-168,S2-138,S2-139` | no library choice Stage2; selection criteria |
| `SA-081` | `186-187` | `b65d2ea28a3f92fbce050a0a00e3bab46f1049294932171502df2b7779b414db` | `S2-140,S2-141` | ADR for architecture/lock-in; no weakening |
| `SA-082` | `191-199` | `d56b74037cdc510ae09463569ad24e382bb1047b2be3e141ea41cae619c56f8a` | `S2-142,S2-143,S2-144,S2-145,S2-146,S2-147,S2-148` | Stage4 entry criteria before first SQL migration |

`SA-001…SA-082` is the exact source-anchor set for the current candidate. `R045` mechanically verifies exact SA ID/range/hash
and exact-once S2 accountability. This does **not** make human source-fidelity judgment unnecessary:
Reviewer still checks that each atomic S2 wording preserves the subject, qualifier, conjunction and scope
of its mapped source bytes. It does make silent source-unit omission or known-needle-only proof impossible
without failing the Builder gate.

### 23.1 Canonical control rows

| Control ID | Stage 2 source | Atomic normative control | Primary disposition | Machine rule IDs | External owner/gate |
| --- | --- | --- | --- | --- | --- |
| `S2-001` | Purpose | Expand→Populate→Switch→Validate→Contract is the mandatory lifecycle | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Principal Architect / lifecycle review; lifecycle vocabulary is prerequisite, not proof of mandatory sequencing |
| `S2-002` | Non-negotiable 1 | No destructive one-step migration | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R007,R008 | separately governed operational migrations and lifecycle mechanisms must independently avoid destructive one-step migration |
| `S2-003` | Non-negotiable 2 | DROP/destructive type conversion/rename/semantic reuse requires staged ADR authority | `REJECTED_BY_PAIRED_SQL_V1` | R007,R008 | Separate destructive-migration scope |
| `S2-004` | Non-negotiable 3 | Old/new application versions coexist during transition window | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Release/closure evidence |
| `S2-005` | Non-negotiable 4 | Financial ledger rows are never UPDATE/DELETE rewritten for cleanup | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R010,R008,R009 | every separately governed operational migration, including Populate tooling, must independently preserve financial-ledger history |
| `S2-006` | Non-negotiable 5 | Snapshots may be rebuilt; transactions may not be rewritten | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R010 | Domain reviewer |
| `S2-007` | Non-negotiable 6 | Every migration is versioned | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R002,R005 | separately governed operational migrations must carry independently governed immutable/versioned identity where Stage 2 calls them migrations |
| `S2-008` | Non-negotiable 6 | Every merged migration is immutable | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R004,R006 | immutability of separately governed operational migration artifacts remains outside paired-SQL validator observation |
| `S2-009` | Non-negotiable 6 | Every migration is reviewed | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | REVIEW_WORKFLOW exact-head evidence |
| `S2-010` | Non-negotiable 6 | Every migration is observable | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Operations/reviewer; declaration profile is prerequisite, not proof that production migration is observable |
| `S2-011` | Non-negotiable 6 | Every migration is rehearsed | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R023 | Risk-specific operational evidence |
| `S2-012` | Non-negotiable 7 | Production schema changes execute through CI/CD | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Deployment evidence |
| `S2-013` | Non-negotiable 7 | Production schema changes use least-privilege credentials | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Deployment/security evidence |
| `S2-014` | Non-negotiable 8 | Backup/PITR verified before high-risk changes | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | High-risk closure gate |
| `S2-015` | Non-negotiable 8 | Backup existence is not a substitute for rollback design | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R026 | Reviewer |
| `S2-016` | Non-negotiable 9 | No production-code TODO may defer a known migration risk; use Issue→ADR→approval | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Review/governance gate |
| `S2-017` | Phase 1 | Expand structures are backward-compatible | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Compatibility reviewer; additive grammar/impact metadata are prerequisites, not proof of backward compatibility |
| `S2-018` | Phase 1 | New column is nullable or uses safe server-side default | `MACHINE` | R008 | — |
| `S2-019` | Phase 1 | New table is ignored by old application versions | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Compatibility reviewer |
| `S2-020` | Phase 1 | New schema is ignored by old application versions | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Separate reviewed schema-extension scope |
| `S2-021` | Phase 1 | New index is ignored by old application versions | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Compatibility reviewer; grammar/concurrency are prerequisites, not evidence of old-version ignore behavior |
| `S2-022` | Phase 1 | New enum/check behavior only after old readers tolerate it | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Compatibility reviewer |
| `S2-023` | Phase 1 | New read model/event version coexists with old representation | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Separate application/event lifecycle scope |
| `S2-024` | Phase 1 | Additive API field only when clients tolerate unknown fields | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | API/OpenAPI review |
| `S2-025` | Phase 1 | Large-table indexes use PostgreSQL online/concurrent mechanisms where supported | `MACHINE` | R029 | — |
| `S2-026` | Phase 1 | Large-table indexes are kept out of transactions when PostgreSQL requires it | `MACHINE` | R011,R029 | — |
| `S2-027` | Phase 1 | Lock timeout is explicit | `MACHINE` | R012 | — |
| `S2-028` | Phase 1 | Statement timeout is explicit | `MACHINE` | R012 | — |
| `S2-029` | Phase 1 | Every DDL statement has an estimated lock mode | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R013 | Reviewer adequacy |
| `S2-030` | Phase 1 | Every DDL statement has an affected row count | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R013 | Reviewer adequacy |
| `S2-031` | Phase 1 | Every DDL statement has a disk impact | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R013 | Reviewer adequacy |
| `S2-032` | Phase 1 | Every DDL statement has a replication/WAL impact | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R013 | Reviewer adequacy |
| `S2-033` | Phase 1 | Every DDL statement has an abort condition | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R013 | Reviewer adequacy |
| `S2-034` | Phase 2 | Populate uses resumable/idempotent batches ordered by stable primary key | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-035` | Phase 2 | Populate uses bounded transactions and rate limits | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-036` | Phase 2 | Populate persists progress/watermark separately from business data | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-037` | Phase 2 | Populate continuously measures checksums/counts/domain invariants | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-038` | Phase 2 | Populate pauses/resumes without duplicate business effects | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-039` | Phase 2 | Populate performs no binary-float conversion for financial values | `REJECTED_BY_PAIRED_SQL_V1` | R007,R030 | Operational-migration scope |
| `S2-040` | Phase 2 | Populate performs no date conversion through local timezones | `REJECTED_BY_PAIRED_SQL_V1` | R007,R030 | Operational-migration scope |
| `S2-041` | Phase 2 | Populate jobs are operational migrations, not permanent business workers | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration scope |
| `S2-042` | Phase 2 | Populate jobs are removed after validation and Contract | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Operational-migration/closure scope |
| `S2-043` | Phase 3 | Switch is a separately deployable application change | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application/config scope |
| `S2-044` | Phase 3 | Switch-capable code can read both representations | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application scope |
| `S2-045` | Phase 3 | Shadow reads/comparison are enabled only where privacy permits | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application + privacy scope |
| `S2-046` | Phase 3 | Traffic switch uses narrowly scoped feature/config control | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application/config scope |
| `S2-047` | Phase 3 | Old read/write path is retained for rollback during observation | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Release scope |
| `S2-048` | Phase 3 | Indefinite dual writes are avoided | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application scope |
| `S2-049` | Phase 3 | If dual writes unavoidable, ordering/failure recovery/source of truth are explicit | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Architecture review |
| `S2-050` | Phase 3 | Canonical source remains explicit throughout Switch | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Architecture/release review |
| `S2-051` | Phase 3 | Cache invalidation is part of Switch plan | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Application/release review |
| `S2-052` | Phase 3 | OpenAPI-breaking switch uses versioned contract and is not hidden in DB migration | `REJECTED_BY_PAIRED_SQL_V1` | R007 | API review |
| `S2-053` | Phase 4 | Validate checks referential/uniqueness invariants | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-054` | Phase 4 | Validate checks decimal scale and half-even expected results | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-055` | Phase 4 | Validate checks transaction revision continuity/reversal integrity | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-056` | Phase 4 | Validate checks BusinessDate equality without UTC drift | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-057` | Phase 4 | Validate checks snapshot rebuild equality against canonical ledger and methodology version | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-058` | Phase 4 | Validate checks identity/investment separation and absence of personal data in financial schemas | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation + privacy review |
| `S2-059` | Phase 4 | Validate checks outbox/inbox deduplication and business-version monotonicity | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate validation mechanism |
| `S2-060` | Phase 4 | Validate checks query plans and SLO evidence on representative volume | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Performance/validation scope |
| `S2-061` | Phase 4 | Validate includes backup restore plus migration replay in non-production | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Restore/validation scope |
| `S2-062` | Phase 4 | Signed validation report identifies dataset/watermark | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Validation/closure evidence |
| `S2-063` | Phase 4 | Signed validation report identifies queries/tool versions | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Validation/closure evidence |
| `S2-064` | Phase 4 | Signed validation report records mismatches | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Validation/closure evidence |
| `S2-065` | Phase 4 | Signed validation report records accepted risk | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Validation/closure evidence |
| `S2-066` | Phase 4 | Signed validation report identifies reviewer | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Validation/closure evidence |
| `S2-067` | Phase 4 | Unexplained financial mismatch blocks Contract | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Principal Architect / closure |
| `S2-068` | Phase 5 | Contract only after all production traffic uses new path | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-069` | Phase 5 | Contract only after rollback/observation window elapsed | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-070` | Phase 5 | Contract only when no supported app version depends on obsolete representation | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-071` | Phase 5 | Contract only after validation is green | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-072` | Phase 5 | Contract only when retention/legal/privacy permits removal | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Legal/privacy/Contract scope |
| `S2-073` | Phase 5 | A staged ADR explicitly authorizes destructive removal before Contract | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-074` | Phase 5 | A fresh backup/restore rehearsal proves the final path before Contract | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-075` | Phase 5 | Contract is its own PR/deployment and not bundled with Expand/Switch | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-076` | Phase 5 | DROP is not the default end state; retaining a deprecated structure temporarily is safer than premature loss | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive/Contract scope |
| `S2-077` | Risk table | Low risk minimum gate: review | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | REVIEW_WORKFLOW |
| `S2-078` | Risk table | Low risk minimum gate: CI | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Required CI |
| `S2-079` | Risk table | Low risk minimum gate: rollback statement | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R026 | Reviewer |
| `S2-080` | Risk table | Medium risk minimum gate: rehearsal | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R023 | Operational/review evidence |
| `S2-081` | Risk table | Medium risk minimum gate: metrics | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R014,R025 | Machine proves declared metric/observability/rollout structure only; real production metrics existence/emission/use and adequacy require reviewer/operations evidence |
| `S2-082` | Risk table | Medium risk minimum gate: staged rollout | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R025 | Release/reviewer adequacy |
| `S2-083` | Risk table | High risk minimum gate: ADR | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R015,R016 | Principal Architect |
| `S2-084` | Risk table | High risk minimum gate: security/privacy review | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R015,R016,R017 | Security/Privacy reviewer |
| `S2-085` | Risk table | High risk minimum gate: golden vectors | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R015,R016 | Domain/math reviewer |
| `S2-086` | Risk table | High risk minimum gate: restore rehearsal | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | High-risk operational gate |
| `S2-087` | Risk table | Destructive risk requires separate staged ADR and is normally forbidden | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Destructive scope |
| `S2-088` | Versioning | Migration IDs are monotonic | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R002,R005 | IDs/ordering for any separately governed operational migrations must satisfy the Stage 2 versioning model through their own governance |
| `S2-089` | Versioning | Migration IDs are immutable | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R006 | immutability of separately governed operational-migration identity remains external evidence |
| `S2-090` | Versioning | Each future migration declares schema owner | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R019 | schema-owner declaration for separately governed operational migrations remains external governance evidence |
| `S2-091` | Versioning | Each future migration declares lifecycle phase | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R007 | phase declaration for separately governed operational migrations remains external governance evidence |
| `S2-092` | Versioning | Each future migration declares dependency | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R018 | dependency declaration for separately governed operational migrations remains external governance evidence |
| `S2-093` | Phase 2 | A Populate job may remain after validation/Contract only when an approved ongoing responsibility remains | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Approved Architecture/reviewer evidence for the continuing responsibility |
| `S2-094` | Versioning | Each future migration declares reversibility | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R026,R028 | reversibility semantics for separately governed operational migrations remain external governance evidence |
| `S2-095` | Versioning | Each future migration declares expected duration | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R012 | expected-duration declaration for separately governed operational migrations remains external governance evidence |
| `S2-096` | Versioning | Each future migration declares lock risk | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R012 | lock-risk declaration for separately governed operational migrations remains external governance evidence |
| `S2-097` | Versioning | Each future migration declares data classification | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R017 | classification for separately governed operational migrations remains external governance evidence |
| `S2-098` | Versioning | Each future migration declares monitoring | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R014 | monitoring declaration/adequacy for separately governed operational migrations remains external evidence |
| `S2-099` | Versioning | Each future migration declares rollback procedure | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R026 | rollback procedure for separately governed operational migrations remains external governance evidence |
| `S2-100` | Versioning | Each future migration declares roll-forward procedure | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R026 | roll-forward procedure for separately governed operational migrations remains external governance evidence |
| `S2-101` | Versioning | Separate context migrations may deploy independently | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Architecture review |
| `S2-102` | Versioning | Cross-schema dependencies are explicit | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R018 | Architecture review |
| `S2-103` | Versioning | Production runtime verifies compatible migration range, not merely latest | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Runtime/migration-tool scope |
| `S2-104` | Versioning | Failed migration stops pipeline | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R023 | failure propagation/stoppage for separately governed operational migration mechanisms remains external pipeline evidence |
| `S2-105` | Versioning | Failed migration is not manually marked successful | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Operations/process evidence |
| `S2-106` | Rollback | Rollback prefers application/config rollback while expanded structures remain additive | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Reviewer/operations; rollback-strategy declaration is prerequisite, not proof of operational preference |
| `S2-107` | Rollback | Schema down migration is used only when demonstrably safe | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Rollback/Operations/Reviewer; exact inverse is a prerequisite, not evidence that production execution is demonstrably safe |
| `S2-108` | Rollback | Populated data may remain unused after rollback | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Operational migration scope |
| `S2-109` | Rollback | Financial facts are never deleted to simulate rollback | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R010,R028 | Operations/Architecture evidence that actual production rollback mechanisms never delete financial facts |
| `S2-110` | Rollback | If a migration created external side effects or emitted events, rollback is compensating and idempotent | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Operational/event migration scope |
| `S2-111` | Rollback | Transport remains at least once; consumers protect business effects with inbox keys/business versions | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Runtime/event scope |
| `S2-112` | Snapshot | Snapshot schema/methodology change creates a new version | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-113` | Snapshot | Snapshot rebuilds from immutable transactions and registered market/inflation inputs | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-114` | Snapshot | Snapshot compares golden vectors and prior version with explained deltas | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-115` | Snapshot | Snapshot switches reads by methodology/version | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-116` | Snapshot | Snapshot retains prior version through rollback window | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-117` | Snapshot | Snapshot contracts only when retention/audit permits | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope |
| `S2-118` | Snapshot | Historical transactions are never mutated to make snapshot match | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R010 | Snapshot/runtime reviewer evidence that rebuild/runtime mechanisms preserve historical transaction immutability |
| `S2-119` | Identity deletion | Identity deletion remains irreversible after restoring older encrypted backup | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/privacy deletion architecture |
| `S2-120` | Identity deletion | Sensitive identity/link material is protected with revocable per-subject encryption-key material | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/privacy deletion architecture |
| `S2-121` | Identity deletion | Deletion removes live rows/reversible mapping and cryptographically destroys the corresponding key material | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/privacy deletion architecture |
| `S2-122` | Identity deletion | Restored backup cannot recreate deleted identity link without destroyed key | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/privacy deletion architecture |
| `S2-123` | Identity deletion | Restore replays deletion ledger before serving traffic | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/privacy restore runbook |
| `S2-124` | Identity deletion | Backups remain encrypted | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Security/operations evidence |
| `S2-125` | Identity deletion | Backups expire within 90 days | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Security/operations evidence |
| `S2-126` | Identity deletion | The exact key hierarchy, Vault policy, deletion ledger and restore runbook require Security Review before implementation | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Security Review |
| `S2-127` | Identity deletion | Operator-reconstructable deleted identity link is pseudonymization and fails the approved anonymization requirement | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Security/privacy review |
| `S2-128` | Observability | Production migration reports version/phase/owner/start/end/status | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-129` | Observability | Production migration reports rows/batches processed without row content | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-130` | Observability | Production migration reports lock wait | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-131` | Observability | Production migration reports statement duration | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-132` | Observability | Production migration reports replication lag | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-133` | Observability | Production migration reports WAL/disk growth | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-134` | Observability | Production migration reports validation mismatch counts | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-135` | Observability | Production migration reports retry/pause/abort reason | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-136` | Observability | Production migration reports request/change ticket and deployment correlation | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Execution evidence; report schema declaration is prerequisite, not proof of runtime emission |
| `S2-137` | Observability | Metrics and logs contain no passwords/tokens/passport/INN/raw XML/PDF/financial document content | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | R020,R031 | R020/R031 prove only validator/controlled-CI output hygiene; total PostgreSQL/psql/deployment logging surface requires Security/Operations closure evidence |
| `S2-138` | Tooling | Selecting a migration library is a Stage 4 dependency decision | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Migration-tool selection scope |
| `S2-139` | Tooling | Tool selection considers Go compatibility, checksum/locking behavior, transactional support, observability, maintenance, license and rollback workflow | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Migration-tool selection scope |
| `S2-140` | Tooling | A library choice that affects architecture or creates lock-in requires an ADR | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Principal Architect |
| `S2-141` | Tooling | No tool may weaken Stage 2 strategy | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Architecture/reviewer |
| `S2-142` | Stage 4 entry | Stage 4 entry required ER model and ADR-006 approval | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-143` | Stage 4 entry | Stage 4 entry required exact PostgreSQL types/constraints/role grants reviewed | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-144` | Stage 4 entry | Stage 4 entry required migration tool approved | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-145` | Stage 4 entry | Stage 4 entry required local/CI disposable database tests | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-146` | Stage 4 entry | Stage 4 entry required upgrade/rollback rehearsals | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-147` | Stage 4 entry | Stage 4 entry required anonymization/key-destruction threat model approved | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-148` | Stage 4 entry | Stage 4 entry required no unresolved canonical-model question | `HISTORICAL_ENTRY_CRITERION_ONLY` | — | Historical entry criterion |
| `S2-149` | Risk table examples | `new empty table` is a canonical Low-risk example | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Reviewer/domain risk adequacy; minimum floor is prerequisite, not proof of exact Low classification |
| `S2-150` | Risk table examples | `additive nullable column` is a canonical Low-risk example | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Reviewer/domain risk adequacy; minimum floor is prerequisite, not proof of exact Low classification |
| `S2-151` | Risk table examples | `backfill` is a canonical Medium-risk example | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate Populate mechanism |
| `S2-152` | Risk table examples | `new constraint` is a canonical Medium-risk example | `MACHINE` | R034 | — |
| `S2-153` | Risk table examples | `new index` is a canonical Medium-risk example | `MACHINE` | R034 | — |
| `S2-154` | Risk table examples | `read-path switch` is a canonical Medium-risk example | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate Switch/application mechanism |
| `S2-155` | Risk table examples | `financial representation` is a canonical High-risk classification anchor | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Domain/Principal Architect; structural high-risk gates are prerequisite, not proof of classification-anchor semantics |
| `S2-156` | Risk table examples | `identity link` is a canonical High-risk classification anchor | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Security/Privacy reviewer; structural high-risk gates are prerequisite, not proof of classification-anchor semantics |
| `S2-157` | Risk table examples | `encryption` is a canonical High-risk classification anchor | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Security/Privacy review |
| `S2-158` | Risk table examples | `event ordering` is a canonical High-risk classification anchor | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Architecture/domain review |
| `S2-159` | Risk table examples | `DROP` is a canonical Destructive-risk example | `REJECTED_BY_PAIRED_SQL_V1` | R008,R010 | Separate staged ADR / outside Expand-only paired SQL |
| `S2-160` | Risk table examples | `irreversible conversion` is a canonical Destructive-risk example | `REJECTED_BY_PAIRED_SQL_V1` | R008 | Separate staged ADR / outside Expand-only paired SQL |
| `S2-161` | Risk table examples | `history rewrite` is a canonical Destructive-risk example | `REJECTED_BY_PAIRED_SQL_V1` | R010 | Financial history rewrite remains forbidden |
| `S2-162` | Global priority | Correctness, security, privacy, rollback and availability take priority over delivery speed | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Architecture/reviewer decision rule; registry presence is prerequisite, not proof that operational priority conflicts are resolved by the policy |
| `S2-163` | Phase 2 Populate | Populate/backfill must not change the active read path while new representation is being filled | `REJECTED_BY_PAIRED_SQL_V1` | R007,R008 | Separate governed Populate mechanism; paired-SQL v1 rejects data movement |
| `S2-164` | Phase 4 Validate | Validation must be domain-aware and must not rely only on row counts | `REJECTED_BY_PAIRED_SQL_V1` | R007 | Separate governed Validate mechanism/closure evidence |
| `S2-165` | Snapshot | Snapshot lifecycle expands a new snapshot representation/version before rebuild/switch | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Snapshot lifecycle scope; appended in v9 to preserve stable earlier S2 IDs |
| `S2-166` | Non-negotiable 7 | Production schema changes are never executed manually | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Deployment/process evidence |
| `S2-167` | Versioning | Failed migrations require diagnosis | `OPERATIONAL_OR_CLOSURE_EVIDENCE` | — | Operations/process evidence |
| `S2-168` | Tooling | Stage 2 does not choose a migration library | `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` | — | Migration-tool selection scope |

### 23.1a Machine-evidence scope registry — closes PLAN-04 evidence-binding overclaim

The `Machine rule(s)` column in §23.1 is **not** automatically proof of the entire English Requirement.

Exactly four evidence modes exist for S2 rows:

1. **complete machine proof** — only rows whose disposition is `MACHINE`; the mapped rule set proves the entire S2 Requirement inside the validator/governed-SQL observable boundary;
2. **paired-SQL scope rejection** — rows whose disposition is `REJECTED_BY_PAIRED_SQL_V1`; machine rules prove only that the paired-SQL v1 surface cannot carry the external Populate/Switch/Validate/Contract/destructive behavior. They do **not** prove that the external Stage 2 lifecycle/operational requirement has been fulfilled; the `Required human / external evidence` column remains authoritative for that remainder;
3. **partial machine proof** — a non-MACHINE row may cite rules only when the table below states the exact machine-proven subset and the remaining semantic/operational claim is owned by the stated non-machine evidence;
4. **no machine proof** — machine-rule column is `—`; the requirement is historical, deferred, operational or closure evidence and no unrelated rule may be attached merely to make the registry appear machine-complete.

The exact **complete-machine S2 set** for the current candidate is:

```text
S2_MACHINE_COMPLETE_SET = {
  S2-018,S2-025,S2-026,S2-027,S2-028,S2-152,S2-153
}
```

This set has exactly **7** rows and every member has disposition `MACHINE`. `GUARD-01` additionally requires a per-row subject-universe/observer-universe proof before membership in this set; count/set equality alone is not sufficient.

The exact **paired-SQL scope-rejected set** is:

```text
S2_PAIRED_SQL_SCOPE_REJECTED_SET = {
  S2-003,S2-034,S2-035,S2-036,S2-037,S2-038,S2-039,S2-040,S2-041,
  S2-042,S2-043,S2-044,S2-045,S2-046,S2-047,S2-048,S2-049,S2-050,
  S2-051,S2-052,S2-053,S2-054,S2-055,S2-056,S2-057,S2-058,S2-059,
  S2-060,S2-061,S2-068,S2-069,S2-070,S2-071,S2-072,S2-073,S2-074,
  S2-075,S2-076,S2-087,S2-151,S2-154,S2-159,S2-160,S2-161,S2-163,S2-164
}
```

This set has exactly **46** rows and every member has disposition `REJECTED_BY_PAIRED_SQL_V1`. For these rows, the machine evidence establishes only **scope exclusion / rejection on the paired-SQL v1 surface**. It is never represented as complete proof that the external lifecycle, application, operational, privacy, validation, Contract or destructive-governance requirement itself occurred or is adequate.

`S2-002`, `S2-005`, `S2-007`, `S2-008`, `S2-088…S2-100`, `S2-104`, `S2-109` and `S2-118` are intentionally outside both sets because the machine observes only a subset of their global/operational subject. `S2-001`, `S2-010`, `S2-017`, `S2-019`, `S2-021`, `S2-022`, `S2-093`, `S2-101`, `S2-106`, `S2-107`, `S2-128`, `S2-129`, `S2-130`, `S2-131`, `S2-132`, `S2-133`, `S2-134`, `S2-135`, `S2-136`, `S2-141`, `S2-149`, `S2-150`, `S2-155`, `S2-156` and `S2-162` intentionally have no machine rule and remain reviewer/external-evidence owned because no direct machine-observable logical subset is claimed.

Every other S2 row carrying machine-rule IDs is in the exact partial-evidence registry below.

| S2 ID | Machine-proven subset only | Required non-machine remainder |
| --- | --- | --- |
| `S2-002` | R007/R008 prove paired-SQL v1 excludes destructive one-step forms | separately governed operational migrations and lifecycle mechanisms must independently avoid destructive one-step migration |
| `S2-005` | R010/R008/R009 prove governed paired-SQL and disposable inverse surfaces cannot UPDATE/DELETE ledger history for cleanup | every separately governed operational migration, including Populate tooling, must independently preserve financial-ledger history |
| `S2-007` | R002/R005 prove every paired-SQL migration pair uses the frozen versioned filename/manifest identity | separately governed operational migrations must carry independently governed immutable/versioned identity where Stage 2 calls them migrations |
| `S2-008` | R004/R006 prove merged paired-SQL migration files are immutable against the protected-base model | immutability of separately governed operational migration artifacts remains outside paired-SQL validator observation |
| `S2-088` | R002/R005 prove paired-SQL migration IDs are monotonic within the governed SQL sequence | IDs/ordering for any separately governed operational migrations must satisfy the Stage 2 versioning model through their own governance |
| `S2-089` | R006 proves paired-SQL migration IDs/files are immutable after merge | immutability of separately governed operational-migration identity remains external evidence |
| `S2-090` | R019 proves schema owner is declared for every paired-SQL migration manifest | schema-owner declaration for separately governed operational migrations remains external governance evidence |
| `S2-091` | R007 proves lifecycle phase is declared for every paired-SQL migration manifest | phase declaration for separately governed operational migrations remains external governance evidence |
| `S2-092` | R018 proves declared dependency structure for paired-SQL migrations | dependency declaration for separately governed operational migrations remains external governance evidence |
| `S2-094` | R026/R028 prove reversibility/rollback declaration and exact disposable inverse for paired-SQL migrations | reversibility semantics for separately governed operational migrations remain external governance evidence |
| `S2-095` | R012 proves expected duration metadata for paired-SQL directions | expected-duration declaration for separately governed operational migrations remains external governance evidence |
| `S2-096` | R012 proves lock-risk metadata for paired-SQL directions | lock-risk declaration for separately governed operational migrations remains external governance evidence |
| `S2-097` | R017 proves data classification is declared for paired-SQL migration manifests | classification for separately governed operational migrations remains external governance evidence |
| `S2-098` | R014 proves monitoring declaration structure for paired-SQL migrations | monitoring declaration/adequacy for separately governed operational migrations remains external evidence |
| `S2-099` | R026 proves rollback procedure is declared for paired-SQL migration manifests | rollback procedure for separately governed operational migrations remains external governance evidence |
| `S2-100` | R026 proves roll-forward procedure is declared for paired-SQL migration manifests | roll-forward procedure for separately governed operational migrations remains external governance evidence |
| `S2-104` | R023 proves any paired-SQL apply/DOWN/baseline/reapply failure fails the validator/CI pipeline | failure propagation/stoppage for separately governed operational migration mechanisms remains external pipeline evidence |
| `S2-006` | R010 proves governed SQL does not rewrite financial facts through prohibited DML | snapshot rebuild correctness and domain equivalence |
| `S2-011` | R023 proves the disposable rehearsal sequence executes | freshness, representativeness and operational adequacy |
| `S2-015` | R026 proves rollback/roll-forward structure is declared | backup and rollback design adequacy remain reviewer-owned |
| `S2-029` | R013 proves an estimated lock-mode field exists in the frozen domain | correctness/adequacy of the estimate for production |
| `S2-030` | R013 proves an affected-row estimate field exists with frozen type/bounds | correctness/adequacy of the production estimate |
| `S2-031` | R013 proves a disk-impact declaration exists with frozen type/bounds | correctness/adequacy of the production estimate |
| `S2-032` | R013 proves a replication/WAL-impact declaration exists with frozen domain | correctness/adequacy of the production estimate |
| `S2-033` | R013 proves an abort-condition declaration exists structurally | operational adequacy of the chosen abort condition |
| `S2-079` | R026 proves rollback/roll-forward structure is present | low-risk rollback strategy adequacy |
| `S2-080` | R023 proves disposable rehearsal execution | medium-risk rehearsal freshness/representativeness |
| `S2-081` | R014/R025 prove metrics/observability/rollout structure only | real metrics emission/use and operational adequacy |
| `S2-082` | R025 proves staged-rollout fields and typed evidence structure | staged rollout plan adequacy and release approval |
| `S2-083` | R015/R016 prove required typed ADR authority-reference/risk structure | substantive ADR approval and architectural adequacy |
| `S2-084` | R015/R016/R017 prove required security/privacy authority-reference/classification structure | substantive Security/Privacy Review occurrence and adequacy |
| `S2-085` | R015/R016 prove required golden-vector authority-reference/risk structure | domain/math correctness and review adequacy of vectors |
| `S2-102` | R018 proves explicit declared dependency edges are valid | semantic completeness of cross-schema dependencies |
| `S2-109` | R010/R028 prove governed migration SQL and disposable same-UP DOWN cannot delete financial facts through prohibited SQL/effects | actual production rollback procedure/application/tooling never deletes financial facts to simulate rollback |
| `S2-118` | R010 proves governed migration SQL contains no historical-transaction mutation surface | separate snapshot rebuild/runtime/application tooling never mutates historical transaction rows to force snapshot equality |
| `S2-137` | R020/R031 prove only validator and controlled-CI log-safety surfaces | total PostgreSQL/psql/deployment metrics+logs privacy closure |

The exact evidence cardinalities are derived from canonical S2 rows and frozen once in the machine-readable authority below; prose MUST NOT carry an independently editable count:

```text
S2-EVIDENCE-PARTITION|complete=7|scope_rejected=46|partial=36|none=79
```

No other row outside `S2_MACHINE_COMPLETE_SET`, `S2_PAIRED_SQL_SCOPE_REJECTED_SET` and this partial registry may carry a machine-rule ID.

`S2-101` has no machine-proven subset in the current candidate: dependency-graph validity is not a logical subset of independent deployability, so its machine-rule cell is `—` and Architecture review owns the control. Reintroducing R018 there is a mandatory mutation failure (TC-571).

A machine binding is valid only if its machine-proven subset is a logical subset of the S2 requirement. For any Requirement containing the semantic quantifiers `all`, `every`, `never`, or `no` over **migration(s)**, `GUARD-01` additionally requires an explicit `REQUIREMENT_SUBJECT_UNIVERSE` and `MACHINE_OBSERVER_UNIVERSE`; complete `MACHINE` is legal only when `REQUIREMENT_SUBJECT_UNIVERSE ⊆ MACHINE_OBSERVER_UNIVERSE`. The sibling scan covers every such S2 row, not only previously reported examples.
The existence of a valid rule ID is never sufficient. Builder and reviewer mutation fixtures must fail when:

- any machine edge is reattached to reviewer-only `S2-019`, `S2-021`, `S2-022`, or `S2-141`;
- `S2-093` is given R018 or any unrelated machine rule;
- `S2-109` or `S2-118` is promoted to complete-machine while external rollback/snapshot behavior remains outside validator observation;
- a partial S2 binding lacks an explicit machine subset or non-machine remainder;
- a non-complete-machine, non-scope-rejected S2 row carries a machine rule but is absent from this registry;
- a `REJECTED_BY_PAIRED_SQL_V1` row is described as complete machine proof of its external Stage 2 requirement rather than scope rejection only.

`R049` enforces the exact structural sets/registry identities. The designated independent reviewer still owns the semantic-subset judgment. v20 carries forward the full v19 semantic pass over every prior partial edge: four reviewer-confirmed prerequisite bindings plus eighteen additional prerequisite-only behavioral/reporting/compatibility/classification/policy bindings are conservatively reviewer/external-evidence owned. The exact remaining 36 partial-edge IDs are frozen. Builder rejects any unreviewed partial-set drift; independent review must still inspect all 36 semantic edges rather than infer relevance from R existence.

### 23.2 P3-08-derived hardening registry

`S2-*` is reserved **only** for controls extracted from the canonical Stage 2 source blob.
P3-08-specific strengthenings are typed separately:

| Derived ID | Requirement | Disposition | Machine rules |
| --- | --- | --- | --- |
| `P3D-001` | every current migration-SQL CI path (`migrations`,`go`,`go-race`) is exact-SHA validation-dominated before SQL execution | `MACHINE` | R035 |
| `P3D-002` | source-policy and derived-hardening namespaces never mix; exact S2/P3D sets are separately validated | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R038 |
| `P3D-003` | every scalar data-literal/default/CHECK-predicate operator, lexical form, type-compatibility rule and value bound owned by R040 is frozen in Stage 3.54, not chosen in Stage 3.55 | `MACHINE` | R040 |
| `P3D-004` | future SQL object identifiers cannot depend on PostgreSQL truncation; exact 1–63-byte grammar applies | `MACHINE` | R039 |
| `P3D-005` | supported CREATE INDEX syntax is the exact literal §15.6 subset and all omitted PostgreSQL clause families fail closed | `MACHINE` | R041 |
| `P3D-006` | TC→R and R→TC semantic edge sets are exactly equal | `MACHINE` | R037 |
| `P3D-007` | every registry summary uses exact ID range plus computed count; stale bare counts fail semantic lint | `MACHINE` | R038 |
| `P3D-008` | implementation may not introduce an alias, enum value, bound, grammar form or normalization rule absent from the frozen plan | `MACHINE` | R050 |
| `P3D-009` | every required finite-domain manifest field has an exhaustive in-plan value set; scanner-derived enums equal manifest declarations; no domain is delegated to Stage 3.55/code/tests | `MACHINE` | R042 |
| `P3D-010` | every supported index on a pre-existing table uses the reviewed concurrent form even when the table is not independently proven large | `MACHINE` | R029 |
| `P3D-011` | declared lock/statement timeout metadata is bound to the actual validated PostgreSQL execution controls before governed DDL | `MACHINE` | R012 |
| `P3D-012` | machine-obvious supported SQL effects impose the frozen §18.1 minimum-risk floor; author declaration may raise but never lower it | `MACHINE` | R034 |
| `P3D-013` | rollout-plan identity has exactly one source of truth: staged mode has exactly one typed staged-rollout authority ref, standard mode has zero, and `rollout.plan_ref` is forbidden | `MACHINE` | R043 |
| `P3D-014` | manifest aggregate/open-domain field types and reference grammars are frozen; rollout/monitoring arrays reference measured canonical observability keys and required free text has one exact string policy | `MACHINE` | R044 |
| `P3D-015` | dependency graph validity is a P3-08 machine strengthening distinct from the canonical Stage 2 requirement merely to declare dependency | `MACHINE` | R018 |
| `P3D-016` | dependency semantic completeness/adequacy is a human-review strengthening distinct from structural dependency declaration | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R018 |
| `P3D-017` | expected-duration and lock-risk manifest/execution modeling is per direction even though Stage 2 requires only migration-level declaration | `MACHINE` | R012 |
| `P3D-018` | canonical Stage 2 source accountability is byte-bound through `SA-*` anchors rather than inferred from S2 continuity or known-needle checks; normative classification/fidelity remains independently reviewed | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R045 |
| `P3D-019` | authority-path identity and open-string emptiness use exact lexical/byte grammars with no implementation-selected normalization | `MACHINE` | R046,R047 |
| `P3D-020` | CREATE TABLE and ADD COLUMN accepted token languages are the exact literal productions in §§15.3–15.4 | `MACHINE` | R048 |
| `P3D-021` | the S2 evidence-binding registry is structurally closed: complete-machine/scope-rejected/partial/no-machine sets are exact, partial rows state machine subset + external remainder, and semantic subset adequacy remains independently reviewed | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R049 |
| `P3D-022` | the global no-implementation-invention invariant is proven against one exact aggregate semantic-freeze rule set, not a hand-picked subset | `MACHINE` | R050 |
| `P3D-023` | `numeric(p,s)` precision/scale and `varchar(n)` length parameters use exact canonical decimal-token languages plus frozen value bounds | `MACHINE` | R051 |
| `P3D-024` | accepted CHECK constraints use one exact complete `ALTER TABLE ... ADD CONSTRAINT ... CHECK (...) NOT VALID ;` token production | `MACHINE` | R052 |
| `P3D-025` | every supported DOWN inverse class has one exact literal token production; effect identity alone never broadens accepted syntax | `MACHINE` | R053 |
| `P3D-026` | validator SQL-file discovery covers every migration-directory `.sql` subject and therefore strictly dominates the frozen CI `*.up.sql` execution set; malformed SQL filenames reject rather than disappear | `MACHINE` | R054 |
| `P3D-027` | the P3D evidence-binding registry is structurally closed: exact dispositions/rule mappings and broad-vs-leaf scope declarations are frozen, while semantic proof-subset adequacy remains independently reviewed | `STRUCTURE_PLUS_HUMAN_ADEQUACY` | R055 |

### 23.2a Exact aggregate semantic-freeze proof surface — closes PLAN-16

`P3D-008` is intentionally broad, but v12 does not maintain its rule set by hand-picked examples.
Every machine rule is first classified by whether it owns **candidate acceptance/rejection or implementation
semantics** versus being only proof/meta orchestration.

The exact meta/proof-only exclusion is:

```text
SEMANTIC_META_RULE_SET = {
  R037,  # TC↔R projection equality
  R038,  # registry/source-vs-derived structural meta-lint
  R045,  # canonical source-anchor proof
  R049,  # S2 evidence-scope meta-proof
  R050,  # aggregate semantic-set proof itself
  R055,  # P3D evidence-scope meta-proof
  R056,  # observer-universe meta-proof
  R057,  # finite-boundary proof inventory
  R059,  # semantic-atom ownership meta-proof
  R060,  # mutation-kill meta-proof
  R061,  # remediation-generalization meta-proof
  R062,  # external-authority semantic-projection meta-proof
  R063,  # discovered-universe equality / anti-circular-expected-set meta-proof
  R064,  # semantic-atom bijection/body-digest meta-proof
  R065,  # normative-line accountability meta-proof
  R066,  # single-source ColId policy meta-proof
  R067,  # S2-101 evidence-edge honesty regression/meta-boundary
  R068,  # global physical authority occurrence meta-proof
  R069,  # occurrence-level ATOM bijection meta-proof
  R070,  # derived S2 partition-summary / evidence-edge closure meta-proof
  R071,  # UP/DOWN independent positive-acceptance meta-proof
  R072,  # property-level semantic manifest meta-proof
  R073,  # structural-cardinality discovery/equality meta-proof
  R074,  # exact child-property PROP-SPEC binding meta-proof
  R075   # conservative S2-107 evidence-relevance meta-proof
}
```

The one authoritative semantic-owner set is mechanically derived as:

```text
SEMANTIC_FREEZE_RULE_SET =
  {R001…R075} - SEMANTIC_META_RULE_SET
```

Therefore the exact current semantic-owner set has **50** rules and necessarily includes, among others:

```text
R008  # future UP allowlist / accepted statement families
R009  # DOWN inverse semantic contract
R011  # transaction framing/control statements
R027  # encoding/client lexical surface
R029  # concurrent-index contextual acceptance
R030  # binary-float/local-timezone prohibition
R032  # homogeneous transaction-class semantics
R033  # FOREIGN KEY + CREATE/ALTER/constraint/type grammar owner
R036  # bounds/type/FK-default semantics
R039,R040,R041,R042,R043,R044,R046,R047,R048,R051,R052,R053,R054
```

`R050` proves all of the following, not merely equality with a manually typed subset:

1. every `R001…R075` rule is classified exactly once as semantic-owner or meta/proof-only;
2. the two sets are disjoint and their union is exactly `R001…R075`;
3. `SEMANTIC_FREEZE_RULE_SET` is exactly the complement of `SEMANTIC_META_RULE_SET`;
4. R033 is necessarily in the semantic-owner set because it owns FOREIGN KEY grammar/cardinality;
5. adding any future R-rule requires an explicit Stage 3.54 classification before the aggregate proof can pass;
6. removing any semantic owner or relabeling it meta/proof-only without reviewed rationale fails.

The mutation suite must independently demonstrate failure for at least:

- filesystem/filename and strict-JSON input language;
- lifecycle/UP/DOWN/DML accepted statement families;
- transaction/timeout/execution-control grammar;
- risk/classification/dependency/owner cross-field semantics;
- finite enum/domain and SQL identifier/bound semantics;
- authority-path normalization and open-text trim;
- scalar literal/value and scalar type-parameter grammar;
- CREATE TABLE/ADD COLUMN;
- **FOREIGN KEY grammar/cardinality via R033**;
- CHECK envelope;
- CREATE INDEX;
- DOWN literal grammar;
- validator-discovery/CI execution-subject grammar.

Thus no already-registered semantic owner can sit outside the aggregate simply because the Builder forgot to
name its family. `P3D-008 = MACHINE | R050` means the complete reviewed non-meta rule surface is frozen.

### 23.2b P3D evidence-scope registry — closes PLAN-16 derived-binding overclaim

P3D evidence uses the same conservative rule as S2 evidence: a valid R-rule ID is not by itself proof of every English word in a derived control. Exactly two P3D evidence modes exist in the current candidate:

1. **complete machine proof** — the entire P3D Requirement is inside the named rule(s)' mechanically observable/frozen property;
2. **structure plus human semantic adequacy** — machine rules freeze the structural subset, while independent review owns semantic provenance/subset/equality that cannot be established merely by ID/count/text matching.

The exact P3D structure-plus-human set is:

```text
P3D_STRUCTURE_PLUS_HUMAN_SET = {
  P3D-002,P3D-016,P3D-018,P3D-021,P3D-027
}
```

The remaining **22/27** P3D rows are `MACHINE`. No P3D row is allowed to hide a human semantic remainder while declaring complete machine proof.

| P3D ID | Machine-proven subset only | Independent reviewer-owned semantic remainder |
| --- | --- | --- |
| `P3D-002` | R038 proves exact S2/P3D registry separation, ID sets and source/derived structural mapping | whether an S2 wording is genuinely source-faithful without P3-08 strengthening, and whether a P3D item is genuinely derived rather than canonical |
| `P3D-016` | R018 proves declared dependency graph validity | whether all real semantic dependencies were actually declared |
| `P3D-018` | R045 proves exact Stage 2 blob, SA ranges/hashes, exact-once S2 accountability and 149/149 line accounting | whether excluded source bytes are truly non-normative and whether each anchor preserves the complete source authority/qualifiers it claims |
| `P3D-021` | R049 proves the exact four-way S2 evidence partition, exact partial registry and known invalid binding mutations | whether each S2→R machine-proven subset is genuinely a logical subset of the English S2 requirement |
| `P3D-027` | R055 proves exact P3D ID→rule/disposition registry, broad-vs-leaf structural split and known rebroadening mutations | whether each P3D→R binding semantically proves the stated property rather than merely referencing a structurally valid rule |

For the remaining machine-complete P3D rows:

- each row names at least one R-rule;
- broad aggregate claims cite aggregate rules (for example P3D-008/P3D-022 → R050), not one narrow leaf rule;
- P3D-003 is intentionally narrow and owns only the R040 scalar-data/default/CHECK-predicate surface;
- P3D-023 exclusively owns the separate R051 `numeric(p,s)` / `varchar(n)` parameter grammar and bounds;
- changing only an R051 type-parameter bound MUST NOT be represented as a failure of P3D-003; it MUST fail P3D-023 and the aggregate P3D-008/P3D-022 surface.

`R055` mechanically freezes the exact P3D evidence-mode partition, P3D ID→rule mapping and scope declarations. It MUST fail if any of P3D-002/016/018/021/027 is promoted to complete `MACHINE` without eliminating its reviewer-owned semantic remainder. The independent reviewer still owns the semantic-subset/source-classification judgment; this boundary is intentional and normative.

### 23.2c Preventive Guard registries — v20 TC-atomic / global-occurrence / package-accountability surfaces

#### GUARD-01 universal machine-observer closure — disposition-derived, no semantic sibling allowlist

v13 proved that a hand-written semantic sibling set can self-validate an omission. v20 retains the v19 removal of any hand-authored migration-subject list from the approval proof. The authoritative observer universe is derived directly from the canonical `S2-001…S2-168` table and its primary evidence disposition, not from English keyword matching.

The exact derived partition is:

```text
S2_MACHINE_COMPLETE_SET               = 7
S2_PAIRED_SQL_SCOPE_REJECTED_SET      = 46
S2_PARTIAL_MACHINE_SET                = 36
S2_NO_MACHINE_SET                     = 79
TOTAL                                 = 168

OBSERVER_BEARING_S2_SET =
  S2_MACHINE_COMPLETE_SET
  ∪ S2_PAIRED_SQL_SCOPE_REJECTED_SET
  ∪ S2_PARTIAL_MACHINE_SET
  = 89
```

This is the complete machine-observer domain because every canonical S2 row is classified exactly once and no `S2_NO_MACHINE_SET` row may carry an R-rule reference. Every one of the **36** partial rows is already represented in the §23.1a partial-evidence registry with an explicit machine-proven subset plus external/reviewer remainder. Every one of the **46** scope-rejected rows proves only rejection/exclusion on paired-SQL v1, never fulfillment of the external lifecycle requirement. Every one of the **7** complete-machine rows must have its complete Requirement inside the machine-observable surface.

Consequently any current or future row that actually carries partial machine evidence is automatically inside the universal observer closure regardless of wording. `S2-011` is a current example. `S2-010` is intentionally **not** observer-bearing in the current candidate because a monitoring declaration is only a prerequisite for the behavioral requirement "Every migration is observable"; it is reviewer/Operations-owned with no machine R edge.

A lexical/semantic migration-subject scan may still be used as a **diagnostic adversarial cross-check**, but it is not an authority for completeness and its expected count is never an approval condition. R056/R063 instead prove equality between the canonical evidence dispositions, the machine-rule-bearing rows, and the complete/scope-rejected/partial/no-machine registries.

Candidate-injection mutation: add a synthetic S2 row with `STRUCTURE_PLUS_HUMAN_ADEQUACY` plus an R-rule but omit it from the partial evidence-scope registry; universal observer closure MUST fail without relying on any word such as `migration`, `every`, `never`, or a hand-maintained expected sibling count. Removing `S2-011` or any other canonical partial-machine row from the partial evidence registry while leaving its canonical S2 row unchanged MUST fail for the same reason. TC-559/560 exercise these failures; TC-561 proves exact `89 observer-bearing + 79 no-machine = 168` closure.

#### GUARD-02 finite-boundary registry — discovered bounded fields plus exact witnesses

`BOUNDARY_ATOM_REGISTRY` contains only frozen **quantitative** machine bounds: numeric ranges, byte-length ranges, bounded integer fields and explicitly numeric list-size ranges. It does **not** own structural exact-one/non-empty/at-most-one/set-relation/bijection semantics. Those semantics belong to the canonical TC↔MPROP machine-property authority below. This split is authoritative and single-source. v20 retains v19 field-name-free discovery. Builder parses **every** formal-manifest field whose JSON type is `integer`; each discovered integer field must resolve exactly once either to a singleton finite-domain authority (`schema_version → FD-001`) or to one row in the authoritative `BOUND_SPEC_REGISTRY` below. Field names are data, never code constants in the discovery function. Expected counts are diagnostic only.

| Boundary atom | Frozen bound | Valid boundary proof | Invalid boundary proof | Owner |
| --- | --- | --- | --- | --- |
| `BND-01` | future SQL identifier bytes `1..63` | TC-394 = 63 PASS | TC-391 = 64 REJECT | R039/R057 |
| `BND-02` | CREATE TABLE column count `1..64` | TC-521 = 64 PASS | TC-522 = 65 REJECT | R048/R057 |
| `BND-03` | FOREIGN KEY local/ref list count `1..32` | TC-515 = 32↔32 PASS | TC-514 = 33↔33 REJECT; TC-516 = zero REJECT | R033/R057 |
| `BND-04` | CREATE INDEX key count `1..32` | TC-390 = 32 PASS | TC-387 = 33+ REJECT; TC-386 = zero REJECT | R041/R057 |
| `BND-05` | numeric precision `1..38` and scale `0..p` | TC-480 includes p=1,p=38,s=0,s=p PASS | TC-483 includes p=0,p=39,s>p REJECT | R051/R057 |
| `BND-06` | varchar type length `1..10_485_760` | TC-480 includes n=1,n=10_485_760 PASS | TC-483 includes n=0,n=10_485_761 REJECT | R051/R057 |
| `BND-07` | signed smallint/integer/bigint literal ranges | TC-409 exact minima/maxima PASS | TC-359/360/361 outside range REJECT | R040/R057 |
| `BND-08` | authority path segment `1..255` ASCII bytes | TC-535 = 255 PASS | TC-536 = 256 REJECT | R046/R057 |
| `BND-09` | authority path total `1..1024` ASCII bytes | TC-537 = 1024 PASS | TC-538 = 1025 REJECT | R046/R057 |
| `BND-10` | SHA-256 lexical length exactly 64 lowercase hex | TC-539 = exact 64 PASS | TC-540 = 63/65 or non-hex REJECT | R005/R057 |
| `BND-11` | decoded varchar(n) payload length `<= n` | TC-368 = exactly n PASS | TC-366 = n+1 REJECT | R040/R057 |
| `BND-12` | expected_duration_seconds `[1,86_400]` | TC-541 = 1 and 86_400 PASS | TC-065/066 <=0; TC-080 >86_400 REJECT | R012/R036/R057 |
| `BND-13` | lock_timeout_ms `[1,86_400_000]` | TC-542 = 1 and 86_400_000 PASS | TC-073/075 <=0; TC-080 >86_400_000 REJECT | R012/R036/R057 |
| `BND-14` | statement_timeout_ms `[1,86_400_000]` | TC-543 = 1 and 86_400_000 PASS | TC-074/076 <=0; TC-080 >86_400_000 REJECT | R012/R036/R057 |
| `BND-15` | statement_timeout_ms `>= lock_timeout_ms` | TC-544 equality PASS at min/max | TC-077/078 statement < lock REJECT | R012/R057 |
| `BND-16` | future migration ID `000001..999999` | TC-545 = 999999 PASS | TC-012 = 000000; TC-011 malformed grammar REJECT | R002/R057 |
| `BND-17` | `affected_rows_estimate` `0..9_223_372_036_854_775_807` | TC-546 = 0 and INT64_MAX PASS | TC-547 = -1 and INT64_MAX+1 REJECT | R036/R057 |
| `BND-18` | `disk_impact_bytes_estimate` `0..9_223_372_036_854_775_807` | TC-548 = 0 and INT64_MAX PASS | TC-549 = -1 and INT64_MAX+1 REJECT | R036/R057 |
| `BND-19` | `wal_impact_bytes_estimate` `0..9_223_372_036_854_775_807` | TC-550 = 0 and INT64_MAX PASS | TC-551 = -1 and INT64_MAX+1 REJECT | R036/R057 |

`BOUND_SPEC_REGISTRY` is the **only normative numeric source** for bounded integer manifest fields. Global single-source means the entire normative candidate contains exactly these six physical `BOUND-SPEC|...` occurrences, all inside this anchored registry; an identical or conflicting occurrence anywhere else rejects:

```text
BOUND-SPEC|field=expected_duration_seconds|lower=1|upper=86400|bnd=BND-12
BOUND-SPEC|field=lock_timeout_ms|lower=1|upper=86400000|bnd=BND-13
BOUND-SPEC|field=statement_timeout_ms|lower=1|upper=86400000|bnd=BND-14
BOUND-SPEC|field=affected_rows_estimate|lower=0|upper=9223372036854775807|bnd=BND-17
BOUND-SPEC|field=disk_impact_bytes_estimate|lower=0|upper=9223372036854775807|bnd=BND-18
BOUND-SPEC|field=wal_impact_bytes_estimate|lower=0|upper=9223372036854775807|bnd=BND-19
```

Every numeric prose/table witness outside this block is derived evidence and MUST equal these values. No second `BOUND-SPEC|...` production is permitted anywhere in the candidate. Builder performs a global occurrence scan before parsing the registry and checks exact location, physical multiplicity, field uniqueness and BND ownership. Builder proof compares exact lower/upper values, not only field membership. A new integer field in §7.11 with no FD singleton and no BOUND-SPEC fails. Removing or changing only a BOUND-SPEC lower/upper value while keeping BND IDs/counts unchanged also fails the frozen mutation suite.


Formal-manifest integer discovery is generic: the current document happens to yield seven integer fields, of which `schema_version` resolves to FD-001 and the other six resolve from BOUND-SPEC rows. No field name is embedded in discovery code. TC-552/563 inject a new integer field without a bound/domain authority and MUST fail; TC-562 mutates only an existing frozen upper bound and MUST fail exact value binding; TC-564 proves full generic integer-field closure.

#### GUARD-03/04 lexical intersections, authority projection and semantic ownership

```text
LEXICAL_INTERSECTION-01 = SQL_KEYWORD_SPELLING ∩ UNQUOTED_IDENTIFIER_REGEX
semantic owner = R058
authority-projection meta owner = R062
kwlist blob = a4af3f717a1118e4b3561786c9f642c2ca5772d5
gram.y blob = 03c80eaaf22a74fa2a4a6b977e394d3bc34ffb46
ColId policy authority = COLID_POLICY_SHA256 4782d6b70e84ec183297874379abfca4ff29f06a40987c32f15f81cd7d520c8a
TYPE_FUNC_NAME_KEYWORD members = 23 / sha256 847a5d59765ccb0f3bc47a0642e6b5fa74cff47502a0059750686b5b84af953f
RESERVED_KEYWORD members = 78 / sha256 870e73576d611d33b419d2281b308c5b7e9592dba32d1a95180995d8b8938295
ColId-disallowed union members = 101 / sha256 3a9027604ec759856e3f9fdbaadaccc4588c00b213328ab5ca0018231448e0d6
witnesses = TC-523,TC-524,TC-525,TC-526,TC-554…TC-558
```

`SEMANTIC_ATOM_OWNERSHIP` requires every normative enum/bound/grammar/normalization/lexical-intersection/evidence-scope acceptance property to name exactly one semantic owner R-rule or explicit reviewer-only evidence class. R059 owns this meta-registry. **Authority extraction is not semantic proof by itself**: R062 additionally requires the upstream parser production and category mapping to compose to the exact project property claimed.

##### v20 owner-level semantic R index (retained; not machine-property completeness authority)

The plan no longer treats “all registered R rules are classified” as proof that no normative candidate was added elsewhere. Two independent mechanisms are mandatory:

1. `SEMANTIC_OWNER_ATOM_REGISTRY` is a bijection over semantic-owner R rules; each owner-index digest is SHA-256 of the exact authoritative R semantic body plus LF. This registry proves owner-level integrity only. SEM/CARD manifests prove byte-accountability only; TC↔MPROP is the canonical machine-property authority.
2. `NORMATIVE_LINE_ACCOUNTABILITY` is an external frozen evidence file generated from a deterministic semantic-candidate lexical predicate over the exact plan bytes. Builder validates it but does not regenerate it. Adding/changing a normative candidate line such as a new SQL `MUST` rule therefore fails even when R/TC/ATOM counts remain unchanged. Semantic adequacy of prose→R ownership remains an explicit independent Reviewer judgment; machine proof does not pretend to understand arbitrary English.

| Semantic atom | Exact owner | Body SHA-256 | Authority |
| --- | --- | --- | --- |
| `ATOM-001` | `R001` | `5b6c9d88031f48114d96de66788569ef5f4bdb6e74a79af467e57d3883fa8280` | exact R001 semantic body |
| `ATOM-002` | `R002` | `be0995597f1d8fe87a3c59875300012ee93e38cdaff34ab6c04d342c5d4952a9` | exact R002 semantic body |
| `ATOM-003` | `R003` | `aef06b9445a20942895e6932e994c159e4c0236e0e18f396b83d777bf237eb36` | exact R003 semantic body |
| `ATOM-004` | `R004` | `e0ab2612b65ce89f01338e116a96a722cf84fd1b17d4d2bea09f8ea03184da08` | exact R004 semantic body |
| `ATOM-005` | `R005` | `bc7ff99af58fb463972e9aed200c750457112ba70d883b2caaab19922d51d572` | exact R005 semantic body |
| `ATOM-006` | `R006` | `6257667e86f4052b710b1eae8403d6b8e77d4e4fe9d9c776e4f5606be3aa5912` | exact R006 semantic body |
| `ATOM-007` | `R007` | `c9ac1961ad9597b6c9249b6694432748267b62a29c4e490dbf39ac0a31e3d422` | exact R007 semantic body |
| `ATOM-008` | `R008` | `77546b70d1b8af27c23194bff72d32e357f3d5afdd1cbbcd181dae62c3753fb5` | exact R008 semantic body |
| `ATOM-009` | `R009` | `2680861f075268ac6c601ee380d01937cd95dc48085f5173301356f19d61105e` | exact R009 semantic body |
| `ATOM-010` | `R010` | `02dad1a264ccc7fbeddbc25610cb46ea15c6e15fa81ed90d2f261b074a86e5c9` | exact R010 semantic body |
| `ATOM-011` | `R011` | `52e1c5b6f9abad2bcb0ee0df5c54030b19b6c8f2005a54c0b11a1f0a45bb01fd` | exact R011 semantic body |
| `ATOM-012` | `R012` | `048acc3db99ba0cf3363d0e74761c834ebbe63c8857af06995ecbfc297e280f3` | exact R012 semantic body |
| `ATOM-013` | `R013` | `1aa9caacb2885a7dbf7e5fb506a1155af8d5947b59fc8c8c0f18ea06b5bdd89d` | exact R013 semantic body |
| `ATOM-014` | `R014` | `2adc861d69ded206e0a89441ebda9346b9b00fc42f0fe1c0b86bc54e53ecd979` | exact R014 semantic body |
| `ATOM-015` | `R015` | `900a43b1815ae2526bca2b4c74ad768788e51ff4c33aae55485a7320a5db7720` | exact R015 semantic body |
| `ATOM-016` | `R016` | `f3cf0a3c1f0f1becbfe0a8b825168932d63f1575e3b83797937ffe39627bd011` | exact R016 semantic body |
| `ATOM-017` | `R017` | `d9b26cd36d0273b00357c981d53f93fbbcff5f3523dc0bed3b4ff8bb926b443f` | exact R017 semantic body |
| `ATOM-018` | `R018` | `746dbb1531bd98ab3c0d11cfdb9ed68e4fde151a6956dbc802377a3268c980c4` | exact R018 semantic body |
| `ATOM-019` | `R019` | `a350973aaa9baf05cac68d87f688a04e7334e486e73a7bc75ff5b7048d12f9b9` | exact R019 semantic body |
| `ATOM-020` | `R020` | `162504403a1017d36cb608ef4d44adfa7fcdff757c923338c69608db7907f197` | exact R020 semantic body |
| `ATOM-021` | `R021` | `4ae5e9b545f6d4ddfa611507ce0df8ebf307ecdfe1a6e738370637d41c4f9168` | exact R021 semantic body |
| `ATOM-022` | `R022` | `73974f1742459e5cac19fcc1d2cc5144e1b7c70ce3dd01d5e26d0a8e4b40f9d3` | exact R022 semantic body |
| `ATOM-023` | `R023` | `2b92c66ad1c517ffebc1066af2d7c81fbc42a22d1eff033b2e76abaacaffd673` | exact R023 semantic body |
| `ATOM-024` | `R024` | `369bb4eddaae377376d3753a1bae2f70b4f6c8eb9733cacba0b32ef144a8bf46` | exact R024 semantic body |
| `ATOM-025` | `R025` | `46d315dc514069c652685bdb69c0ec88fc401aed8943f6e0b9d2fec927ff4ab2` | exact R025 semantic body |
| `ATOM-026` | `R026` | `cc5e4f1639303e400f8a684634b712163c727c7a8a57a83ad7bfbf2b1957a354` | exact R026 semantic body |
| `ATOM-027` | `R027` | `4d6e82bfb885ff19331190ba83e4b4e249d13c50b951c02baf1f3d470f4589c2` | exact R027 semantic body |
| `ATOM-028` | `R028` | `8eb287f7ad78be756ef8a19fa85adcd59ed5e9be250e95466213632830a8f6ea` | exact R028 semantic body |
| `ATOM-029` | `R029` | `43b14ebbe871ddd986b0da76c15b01f48ab6590a493e9e216d9a3bb539f52825` | exact R029 semantic body |
| `ATOM-030` | `R030` | `890bba3256be37c77ccf9bec6721571c30996912e065e90205367c9e5cbb54e9` | exact R030 semantic body |
| `ATOM-031` | `R031` | `e3509c5f0cf83589f150bd5ba3afae6166a4de72e185b985e7e86bde8d782dcb` | exact R031 semantic body |
| `ATOM-032` | `R032` | `c58601abb1d30ef6da6acec16312126aa1a20e5cea804b81aaaf71096a495e34` | exact R032 semantic body |
| `ATOM-033` | `R033` | `e56a5d04deb78006864cddc4b390bc3f9925cacd53a53475859d91ab0c6b4d5a` | exact R033 semantic body |
| `ATOM-034` | `R034` | `dedf09e52eaab9a28203e902effd55d70660209971632f40dac7747fecadfaf5` | exact R034 semantic body |
| `ATOM-035` | `R035` | `ad0ace0c1d129385f37fff35b46f9db38a14d02dddd8b7dcadba41e7c368fced` | exact R035 semantic body |
| `ATOM-036` | `R036` | `4ea7c6df29567488cbbac81879b68f933f1e0b928cb0f79856c58e93fa4b433e` | exact R036 semantic body |
| `ATOM-037` | `R039` | `0da43f61ef081bad57704c5ea76a252489cf23d10e96e35602bc6c86e70301ad` | exact R039 semantic body |
| `ATOM-038` | `R040` | `d993d77d7596fab15dfd4484ffbd145d531ca7bd183098b97806472b7bc08bf6` | exact R040 semantic body |
| `ATOM-039` | `R041` | `52f4da3b9e1d650140c4edafa5be192502f1ed84a9df4d7cf69455aef7ad0647` | exact R041 semantic body |
| `ATOM-040` | `R042` | `917b4ae53792559edcb4dfaa6c9c5cfc6c2bada8af1eebe6c3a0f0fa2f980f56` | exact R042 semantic body |
| `ATOM-041` | `R043` | `4a209e74b35153db6b17ddbb4627995e8a1c6530f42d5e255bcf87008bf18b11` | exact R043 semantic body |
| `ATOM-042` | `R044` | `988bc64fb67cdd98a3ac44007a6a74d0935781275d5e9625fa6d963c6983a07c` | exact R044 semantic body |
| `ATOM-043` | `R046` | `2f795b288ca574ac90d385c1d10ef9dab67c257a44674ae8164e2eb2d288c17f` | exact R046 semantic body |
| `ATOM-044` | `R047` | `c6ad320bdee122131f6851a0faca8018242ee838b711871903074047259f3a4d` | exact R047 semantic body |
| `ATOM-045` | `R048` | `7b7b966e2841046ecc4fed1382c94feeff64341e0f32bd95f8d63f2bfce4fd2d` | exact R048 semantic body |
| `ATOM-046` | `R051` | `b210149d5508c9ca4f73fe473edad8337c1bcff1054a43f93823319a6c5215c1` | exact R051 semantic body |
| `ATOM-047` | `R052` | `fc219c3acf6e5607efb4c299c5e881438b9be62c3c9053ea4c38b36e8dce2335` | exact R052 semantic body |
| `ATOM-048` | `R053` | `5a58936ebf63382447dd2345f3e73ee561f3ad27d0fbd70d16c4e7947cd527e1` | exact R053 semantic body |
| `ATOM-049` | `R054` | `0536656a0584a04e29350a0dd67678c231f1f37a1a38b184aa81d8b474feb8c1` | exact R054 semantic body |
| `ATOM-050` | `R058` | `40ac0d5499e280902ecc6952a300ed82ee97bda7cc84d0e6c4ee990c92f5e3bc` | exact R058 semantic body |

Exact owner-index atom set: `ATOM-001…ATOM-050`. Exact owner set: the 50-rule `SEMANTIC_FREEZE_RULE_SET`. The registry contains exactly 50 **physical declarations**. Every ATOM ID occurs exactly once and every semantic owner occurs exactly once; duplicate identical rows, duplicate IDs with different owners, duplicate owners under different IDs, missing+duplicate substitutions and extra physical rows all reject before any dictionary/map is constructed. `R064`/`R069` validate occurrence-level bijection; `R065` validates frozen semantic-candidate line accountability.


##### v20 TC-atomic machine-property authority — retains v19 closures and binds the residual four-field independence contract

v20 retains `ATOM-001…050` only as an **R-owner index**. `SEMANTIC_PROPERTY_MANIFEST` and `CARDINALITY_PROPERTY_MANIFEST` are retained only as byte-accountability/index evidence over the active physical surface; they are **not** semantic-completeness authorities and broad owner TC closure is never a direct child-property witness.

The canonical validator acceptance/rejection contract is instead one-to-one with the canonical test registry. Every `TC-001…TC-631` row has exactly one `MPROP-001…MPROP-631` record containing the exact TC condition/assertion bytes, expected outcome (`PASS`/`REJECT`), exact owner-R set and direct witness polarity. A NEG TC is its own direct rejection witness; a POS TC is its own direct acceptance witness. This makes the machine property atomic at the same granularity as the executable contract rather than at coarse R-owner granularity.

v20 retains v19 **truth binding beyond TC↔MPROP mirroring**: every numeric S2 partition fact embedded in a canonical TC/MPROP condition is checked against an independent derivation from the canonical `S2-001…S2-168` rows and the exact partial-edge registry. A TC and MPROP that are edited consistently with each other but disagree with that derived source fact MUST fail. This closes the v18 stale-mirror failure without introducing a second handwritten partition authority.

No prose sentence, R title, SEM-PROP/CARD-PROP record, BND row, FD row or legacy `PROP-SPEC` may independently introduce or change validator acceptance behavior. They may explain, group or constrain already-authorized properties. A new machine behavior is authorized only by adding/changing/removing a canonical TC together with its exactly-one MPROP; an unknown machine child property without TC+MPROP is outside the reviewed Stage 3.54 contract and fails closed/re-enters planning review. If explanatory prose conflicts with TC↔MPROP, the candidate is review-invalid, but Stage 3.55 MUST implement the TC↔MPROP authority and MUST NOT infer a different acceptance behavior from prose.

Current accepted-language sibling closure is represented directly by `TC/MPROP-627…631`: typed authority identity accepts same-path/different-kind, SQL keyword terminals are ASCII-case-insensitive, terminated line comments between tokens are accepted, terminated nested block comments between tokens are accepted, and forbidden words confined to inert comments do not become executable syntax. Removing or weakening any of these five direct positive witnesses is a semantic-freeze failure.

Machine-readable authority bindings:

```text
SEMANTIC-PROPERTY-MANIFEST|format=P3-08-SEM-PROP-v1|count=2232|sha256=f1e5b121826cd2e89bcf2f693d2819f27e01134e43dfd72461d70b8794b2a278
CARDINALITY-PROPERTY-MANIFEST|format=P3-08-CARD-PROP-v1|count=2232|sha256=2d630edd547074f1973d4761bb597288e24d558fd78de07636cc2c459481a8c0
MACHINE-PROPERTY-MANIFEST|format=P3-08-MPROP-v1|count=631|sha256=dfb52bdf9a32be13dab19a31193d15eb97ec3bbc1a50aecb8a40353d0cb68044
MUTATION-OBLIGATION-MANIFEST|format=P3-08-MUT-OBL-v1|count=631|sha256=ef63c223d81f5d16778a8e488a3b56f092f6f030479174a05a4cfc2fdf5fe2b1
TAXONOMY-AUTHORITY|id=TA-01|BND=quantitative_numeric_byte_and_explicit_numeric_list_bounds|STRUCTURAL_CARDINALITY=MPROP_or_PROP_SPEC|MPROP=all_machine_accept_reject_properties_one_to_one_with_TC|ATOM=R_owner_index_only|SEM_CARD_MANIFEST=byte_accountability_only|NLA=byte_accountability_only
```

`TAXONOMY-AUTHORITY|TA-01` is the **only authoritative registry-scope definition**. Active prose may reference this split but may not redefine it. The forbidden former claim that BND contains all structural cardinalities is a regression failure.

`GUARD-02` is therefore split without overlap: BND owns quantitative bounds; structural cardinality/relation acceptance is represented by TC↔MPROP and, where an additional compact numeric/set relation is useful, the existing `PROP-SPEC` mirror. SEM/CARD manifests remain conservative physical-accountability surfaces only.

Mutation obligations are derived from the MPROP registry itself. For every MPROP the proof suite must exercise at least: TC-condition drift with frozen MPROP, polarity drift, owner-set drift, TC-without-MPROP injection and MPROP-without-TC injection. Builder is not allowed to hand-select a smaller sibling list.

The following high-risk cardinality/acceptance properties additionally have canonical machine-readable `PROP-SPEC` values. These are not examples; they are direct authorities for the exact child semantics that v16 allowed to drift:

```text
PROP-SPEC|id=PROP-001|kind=cardinality|subject=ddl_impact_per_executable_ddl|relation=exact|value=1|owner=R013|positive=TC-282,TC-283|negative=TC-144,TC-145,TC-146,TC-147,TC-149,TC-158|authority_sha256=57081c6df40cc0c2d8d91ee2fc722b86bd040b6a6761f3be0ad9177aa7ec600f
PROP-SPEC|id=PROP-002|kind=cardinality|subject=owners_count|relation=min|value=1|owner=R019|positive=TC-284,TC-285|negative=TC-609|authority_sha256=57041510dc2e55247d958d768c05131299e0f38dedb6a86f4c144f4ace0ae6d1
PROP-SPEC|id=PROP-003|kind=cardinality|subject=monitoring_signals_count|relation=min|value=1|owner=R044|positive=TC-439|negative=TC-610|authority_sha256=6807fbf22a2f91b37ec27aa2a9b469e30d303a7c924a3878d0f56fc04483ecff
PROP-SPEC|id=PROP-004|kind=cardinality|subject=rollout_metrics_count|relation=min|value=1|owner=R044|positive=TC-437|negative=TC-611|authority_sha256=49e3d36c96d4974bbbaca20c81dcfe976b7997fa14f5af9b91bd63c02b256700
PROP-SPEC|id=PROP-005|kind=cardinality|subject=add_column_column_count|relation=exact|value=1|owner=R048|positive=TC-465,TC-466|negative=TC-467|authority_sha256=b019e6bfbfa39beca78a7958c40a955bd80c6545bb39fc3c82f1db91455d9e93
PROP-SPEC|id=PROP-006|kind=cardinality|subject=column_def_null_clause_count|relation=max|value=1|owner=R048|positive=TC-460,TC-466|negative=TC-464|authority_sha256=bcdd555c696c9749015fc705c2883a4b1a7ce303149b001d71b766e59626a043
PROP-SPEC|id=PROP-007|kind=cardinality|subject=column_def_default_clause_count|relation=max|value=1|owner=R048|positive=TC-460,TC-466|negative=TC-464|authority_sha256=042d2a8d7bcddecf70b8ade55df9cc3ba3d430239c9ee98ef7726fc0683f5719
PROP-SPEC|id=PROP-008|kind=cardinality|subject=check_atomic_predicate_count|relation=exact|value=1|owner=R040|positive=TC-373,TC-374|negative=TC-612|authority_sha256=9e4d535998fcd5596101391e4bfa853e77ae0c2302b2091cee260964bed04116
PROP-SPEC|id=PROP-009|kind=cardinality|subject=up_effect_down_inverse_count|relation=exact|value=1|owner=R028|positive=TC-277,TC-302|negative=TC-206,TC-207,TC-613|authority_sha256=e497033592b611ed90653d3727877385bff68d3085687af949c8cbac882a18a7
PROP-SPEC|id=PROP-010|kind=cardinality|subject=down_inverse_up_effect_count|relation=exact|value=1|owner=R028|positive=TC-277,TC-302|negative=TC-207,TC-613|authority_sha256=4b1d71aff6c5a5a4a71834526447cb7caf2a1bf169f8fcb02d852159000a1f26
PROP-SPEC|id=PROP-011|kind=set_relation|subject=touched_schema_vs_declared_owners|relation=equal|value=true|owner=R019|positive=TC-284,TC-285|negative=TC-614|authority_sha256=6637c96d0532017da45783c405a5b0fcffbf3d7f9ad0e894e65f01a6e8e6d184
PROP-SPEC|id=PROP-012|kind=boolean|subject=preexisting_table_index_requires_concurrently|relation=required|value=true|owner=R029|positive=TC-269,TC-270,TC-390|negative=TC-138,TC-160,TC-615|authority_sha256=8913ea35f8b78a06cd4b119347717b7526342b10c45e9b590c0dc54a5a8d79f0
PROP-SPEC|id=PROP-013|kind=cardinality|subject=staged_rollout_ref_count_when_staged|relation=exact|value=1|owner=R043|positive=TC-427|negative=TC-616|authority_sha256=ccb86b3569b39f422c5a617b48e2a74a8039ea7a7b148ffcb3d72f7622315083
PROP-SPEC|id=PROP-014|kind=cardinality|subject=staged_rollout_ref_count_when_standard|relation=exact|value=0|owner=R043|positive=TC-617|negative=TC-618|authority_sha256=7ab1023236b27be344cc8880eb715e0cf6a987c5a086776a581bd3597f14250e
```

All physical `PROP-SPEC|` occurrences are globally inventoried and may exist only in this anchored block. `R072` owns property-manifest exactness, `R073` owns structural-cardinality discovery/equality, `R074` owns exact `PROP-SPEC` value/witness binding, and `R075` owns the current semantic-evidence conservative demotion set including S2-107.

#### GUARD-05/06 mutation-kill, new-candidate injection and remediation generalization

A carried residual is closed only when the named fixture **and** the discovered sibling domain are protected. v20 requires candidate-injection mutations, not only mutations of already registered atoms:

- add a synthetic bounded integer manifest field without BND → R063/R057 fail;
- add a synthetic observer-bearing S2 row (`STRUCTURE_PLUS_HUMAN_ADEQUACY` + R-rule) without partial evidence-scope registration → R063/R056 fail;
- drop `TYPE_FUNC_NAME_KEYWORD` from the authority projection while keeping the exact 78-member reserved set/hash → R062/R060 fail;
- remove `S2-011` or any other canonical partial-machine row from the partial evidence-scope registry while its canonical observer-bearing row remains → R063/R061 fail.

`REMEDIATION_SIBLING_SCAN` records derived universes rather than fixed semantic expected sets. v20 retains the complete MPROP set as the mutation source: every property generates mandatory condition/polarity/owner/orphan obligations; no hand-picked subset is an approval authority.

`REMEDIATION_SIBLING_SCAN` records derived universes rather than fixed semantic expected sets. v20 requires three executable layers from the frozen package: `builder_v20_property_mutation_runner.py` generated from every MPROP, `builder_v20_extra_red_team_runner.py` for independent cross-family attacks including the v17 sibling failures, and `builder_v20_independent_crosscheck.py` which does not import the main meta-audit. Exact totals are package-derived and must reproduce after clean unzip; no earlier headline is exhaustive authority.

### 23.3 Registry completeness invariants

The namespace contract is semantic, not merely numeric:

- every `S2-*` Requirement preserves the canonical source subject, qualifiers, conjunctions and scope without adding P3-08-derived strengthening;
- every `S2-*` appears in exactly one byte-bound `SA-*` row; source line/range/hash mapping is normative provenance, not a Builder note;
- every P3-08-specific strengthening is declared under `P3D-*`;
- stable `S2-*` IDs are never renumbered to hide a later extraction omission; a newly discovered atomic source control appends the next ID;
- source completeness is judged against exact `SA-001…SA-082` anchors recomputed from the canonical Stage 2 blob, not against internal continuity alone.

For canonical source controls:

- exact required source control-ID set: `S2-001…S2-168`;
- `S2-*` may contain only atomic controls traceable to canonical Stage 2 blob `d4656e2bb124fe6ff0783e619eaf608ed1082297`;
- no P3-08-derived CI/meta/semantic-freeze hardening may appear under `S2-*`;
- risk-table examples are atomic: Low `new empty table`, Low `additive nullable column`, Medium `backfill`,
  Medium `new constraint`, Medium `new index`, Medium `read-path switch`, High `financial representation`,
  High `identity link`, High `encryption`, High `event ordering`, Destructive `DROP`,
  Destructive `irreversible conversion`, Destructive `history rewrite`;
- no gaps, duplicates, aliases or unknown active source IDs; every active source ID appears exactly once in `SA-*`;
- every row has exactly one primary disposition from the six-value enum;
- every source obligation-strength qualifier used by an S2 requirement must be inside its accountable SA byte range, not imported from an excluded source line;
- S2 complete-machine classification is permitted only when the complete requirement lies inside the validator/governed-SQL observable boundary; paired-SQL rejection is a separate scope-only category and never full proof of an external requirement; otherwise the row is partial/no-machine;
- a `MACHINE` row has at least one registered machine rule;
- a `STRUCTURE_PLUS_HUMAN_ADEQUACY` row identifies structural rule(s) where applicable plus a human adequacy gate;
- an `OPERATIONAL_OR_CLOSURE_EVIDENCE` row identifies the later owner/gate;
- a `REJECTED_BY_PAIRED_SQL_V1` row maps to stable unsupported/safety rejection;
- a `DEFERRED_TO_SEPARATE_RUNTIME_TOOLING_SCOPE` row identifies future responsible scope;
- a `HISTORICAL_ENTRY_CRITERION_ONLY` row is explicitly historical;
- no normative Stage 2 sentence or risk-table cell is considered covered merely because a broader row sounds similar.

For derived hardening:

- exact derived set: `P3D-001…P3D-027`;
- derived IDs are not counted as canonical Stage 2 source extraction;
- every P3D row is explicitly labeled derived and maps to machine rules;
- unknown/missing/duplicate P3D IDs fail.

Evidence boundary:

- R021 can prove the **approved frozen S2 registry set** after publication, but cannot machine-prove that
  the initial human extraction from prose was semantically exhaustive;
- Stage 3.54 designated review owns the source→S2 extraction-completeness judgment before publication;
- Stage 3.55 may not add/drop/reword an S2 or P3D control, change disposition, or move a control between
  namespaces without returning to reviewed planning scope.

## 24. Self-contained adversarial + acceptance contract — closes P3-08-PLAN-05 and PLAN-07

This section is the complete prospective canonical test contract. It depends on no rejected Stage 3.54
candidate, archive, chat message or unmerged blob.

Earlier rejected candidate numeric labels were never canonical. v20 contains the complete prospective canonical test registry and freezes every semantic value referenced by the tests:

```text
TC-001…TC-631
```

Every row below is normative. Parameterization in Go is allowed only when execution evidence preserves
the exact permanent IDs and semantics.

### 24.1 Complete test-case registry

| Test ID | Polarity | Category | Required behavior | Machine rules |
| --- | --- | --- | --- | --- |
| `TC-001` | `NEG` | discovery/filesystem | missing migrations directory → reject | R001,R027 |
| `TC-002` | `NEG` | discovery/filesystem | no up migrations → reject | R001,R027 |
| `TC-003` | `NEG` | discovery/filesystem | symlink migration SQL → reject | R001,R027 |
| `TC-004` | `NEG` | discovery/filesystem | non-regular migration path → reject | R001,R027 |
| `TC-005` | `NEG` | discovery/filesystem | symlink policy manifest → reject | R001,R027 |
| `TC-006` | `NEG` | discovery/filesystem | case-insensitive filename collision → reject | R001,R027 |
| `TC-007` | `NEG` | discovery/filesystem | unexpected migration-like .sql filename → reject | R001,R027 |
| `TC-008` | `NEG` | discovery/filesystem | invalid UTF-8 in future enforced SQL → reject | R001,R027 |
| `TC-009` | `NEG` | discovery/filesystem | UTF-8 BOM in future enforced SQL → reject | R001,R027 |
| `TC-010` | `NEG` | discovery/filesystem | NUL byte in future enforced SQL → reject | R001,R027 |
| `TC-011` | `NEG` | identity/pairing | invalid six-digit ID grammar → reject | R002 |
| `TC-012` | `NEG` | identity/pairing | 000000 ID → reject | R002 |
| `TC-013` | `NEG` | identity/pairing | uppercase/Unicode/space/repeated-underscore stem → reject | R002 |
| `TC-014` | `NEG` | identity/pairing | missing down pair → reject | R002 |
| `TC-015` | `NEG` | identity/pairing | orphan down pair → reject | R002 |
| `TC-016` | `NEG` | identity/pairing | up/down stem mismatch → reject | R002 |
| `TC-017` | `NEG` | identity/pairing | duplicate numeric ID under different names → reject | R002 |
| `TC-018` | `NEG` | identity/pairing | duplicate name under different IDs → reject | R002 |
| `TC-019` | `NEG` | identity/pairing | manifest/file ID-name disagreement → reject | R002 |
| `TC-020` | `NEG` | identity/pairing | enforced IDs not strictly increasing → reject | R002 |
| `TC-021` | `NEG` | strict-json | malformed JSON → reject | R003 |
| `TC-022` | `NEG` | strict-json | duplicate top-level key → reject | R003 |
| `TC-023` | `NEG` | strict-json | parameterized over every non-top-level object path in the closed manifest schema: duplicate key at that nesting path → reject | R003 |
| `TC-024` | `NEG` | strict-json | unknown top-level field → reject | R003 |
| `TC-025` | `NEG` | strict-json | parameterized over every non-top-level object path in the closed manifest schema, including migration and nested child objects: inject one unknown field → reject | R003 |
| `TC-026` | `NEG` | strict-json | trailing second JSON document → reject | R003 |
| `TC-027` | `NEG` | strict-json | parameterized over every typed field path in the closed manifest schema: substitute a JSON value of any non-allowed type → reject | R003 |
| `TC-028` | `NEG` | strict-json | parameterized over every required field path in the closed manifest schema: remove that required member → reject | R003 |
| `TC-029` | `NEG` | strict-json | parameterized over every field path whose schema does not explicitly permit null: substitute JSON null → reject | R003 |
| `TC-030` | `NEG` | strict-json | parameterized over every array/list field with set semantics: duplicate one otherwise-valid member value → reject | R003 |
| `TC-031` | `NEG` | strict-json | out-of-range integer → reject | R003 |
| `TC-032` | `NEG` | legacy-baseline | legacy baseline entry count not exactly seven → reject | R004 |
| `TC-033` | `NEG` | legacy-baseline | legacy max ID not 000007 → reject | R004 |
| `TC-034` | `NEG` | legacy-baseline | missing legacy entry → reject | R004 |
| `TC-035` | `NEG` | legacy-baseline | extra/future legacy entry → reject | R004 |
| `TC-036` | `NEG` | legacy-baseline | wrong legacy name → reject | R004 |
| `TC-037` | `NEG` | legacy-baseline | legacy up checksum mismatch → reject | R004 |
| `TC-038` | `NEG` | legacy-baseline | legacy down checksum mismatch → reject | R004 |
| `TC-039` | `NEG` | legacy-baseline | retrospective legacy metadata field attempt → reject | R004 |
| `TC-040` | `NEG` | legacy-baseline | legacy entries not strictly sorted → reject | R004 |
| `TC-041` | `NEG` | manifest-bijection/hash | SQL pair missing manifest entry → reject | R005 |
| `TC-042` | `NEG` | manifest-bijection/hash | manifest entry missing SQL pair → reject | R005 |
| `TC-043` | `NEG` | manifest-bijection/hash | invalid SHA-256 grammar → reject | R005 |
| `TC-044` | `NEG` | manifest-bijection/hash | up hash mismatch → reject | R005 |
| `TC-045` | `NEG` | manifest-bijection/hash | down hash mismatch → reject | R005 |
| `TC-046` | `NEG` | manifest-bijection/hash | future enforced ID <= 000007 → reject | R005 |
| `TC-047` | `NEG` | dependency-graph | duplicate dependency → reject | R018 |
| `TC-048` | `NEG` | dependency-graph | self dependency → reject | R018 |
| `TC-049` | `NEG` | dependency-graph | missing declared dependency ID → reject | R018 |
| `TC-050` | `NEG` | dependency-graph | dependency on later ID → reject | R018 |
| `TC-051` | `NEG` | dependency-graph | dependency cycle → reject | R018 |
| `TC-052` | `NEG` | metadata/phase | missing owners → reject | R007,R009,R016,R017,R026 |
| `TC-053` | `NEG` | metadata/phase | unknown phase enum → reject | R007,R009,R016,R017,R026 |
| `TC-054` | `NEG` | metadata/phase | phase=populate in paired-SQL v1 → reject | R007,R009,R016,R017,R026 |
| `TC-055` | `NEG` | metadata/phase | phase=switch in paired-SQL v1 → reject | R007,R009,R016,R017,R026 |
| `TC-056` | `NEG` | metadata/phase | phase=validate in paired-SQL v1 → reject | R007,R009,R016,R017,R026 |
| `TC-057` | `NEG` | metadata/phase | phase=contract in paired-SQL v1 → reject | R007,R009,R016,R017,R026 |
| `TC-058` | `NEG` | metadata/phase | unknown risk enum → reject | R007,R009,R016,R017,R026 |
| `TC-059` | `NEG` | metadata/phase | risk=destructive in paired-SQL v1 → reject | R007,R009,R016,R017,R026 |
| `TC-060` | `NEG` | metadata/phase | unknown classification enum → reject | R007,R009,R016,R017,R026 |
| `TC-061` | `NEG` | metadata/phase | missing monitoring → reject | R007,R009,R016,R017,R026 |
| `TC-062` | `NEG` | metadata/phase | missing production rollback → reject | R007,R009,R016,R017,R026 |
| `TC-063` | `NEG` | metadata/phase | missing roll-forward → reject | R007,R009,R016,R017,R026 |
| `TC-064` | `NEG` | metadata/phase | unsupported reversibility value → reject | R007,R009,R016,R017,R026 |
| `TC-065` | `NEG` | execution/timeouts | up expected duration <=0 → reject | R012,R011 |
| `TC-066` | `NEG` | execution/timeouts | down expected duration <=0 → reject | R012,R011 |
| `TC-067` | `NEG` | execution/timeouts | unknown up lock risk → reject | R012,R011 |
| `TC-068` | `NEG` | execution/timeouts | unknown down lock risk → reject | R012,R011 |
| `TC-069` | `NEG` | execution/timeouts | missing up lock timeout → reject | R012,R011 |
| `TC-070` | `NEG` | execution/timeouts | missing up statement timeout → reject | R012,R011 |
| `TC-071` | `NEG` | execution/timeouts | missing down lock timeout → reject | R012,R011 |
| `TC-072` | `NEG` | execution/timeouts | missing down statement timeout → reject | R012,R011 |
| `TC-073` | `NEG` | execution/timeouts | up lock timeout <=0 → reject | R012,R011 |
| `TC-074` | `NEG` | execution/timeouts | up statement timeout <=0 → reject | R012,R011 |
| `TC-075` | `NEG` | execution/timeouts | down lock timeout <=0 → reject | R012,R011 |
| `TC-076` | `NEG` | execution/timeouts | down statement timeout <=0 → reject | R012,R011 |
| `TC-077` | `NEG` | execution/timeouts | up statement timeout < lock timeout → reject | R012,R011 |
| `TC-078` | `NEG` | execution/timeouts | down statement timeout < lock timeout → reject | R012,R011 |
| `TC-079` | `NEG` | execution/timeouts | timeout value encoded as string/float in manifest → reject | R012,R011 |
| `TC-080` | `NEG` | execution/timeouts | timeout > 86_400_000 ms or duration > 86_400 seconds → reject | R012,R011,R036 |
| `TC-081` | `NEG` | execution/timeouts | transactional up missing SET LOCAL lock_timeout before DDL → reject | R012,R011 |
| `TC-082` | `NEG` | execution/timeouts | transactional up missing SET LOCAL statement_timeout before DDL → reject | R012,R011 |
| `TC-083` | `NEG` | execution/timeouts | transactional down missing SET LOCAL lock_timeout before inverse DDL → reject | R012,R011 |
| `TC-084` | `NEG` | execution/timeouts | transactional down missing SET LOCAL statement_timeout before inverse DDL → reject | R012,R011 |
| `TC-085` | `NEG` | execution/timeouts | non-transactional concurrent-index up missing session SET lock_timeout → reject | R012,R011 |
| `TC-086` | `NEG` | execution/timeouts | non-transactional concurrent-index up missing session SET statement_timeout → reject | R012,R011 |
| `TC-087` | `NEG` | execution/timeouts | timeout SQL value differs from manifest → reject | R012,R011 |
| `TC-088` | `NEG` | execution/timeouts | timeout SET appears after protected DDL → reject | R012,R011 |
| `TC-089` | `NEG` | execution/timeouts | duplicate/conflicting timeout SET → reject | R012,R011 |
| `TC-090` | `NEG` | execution/timeouts | arbitrary SET/RESET GUC → reject | R012,R011 |
| `TC-091` | `NEG` | lexical/dynamic/client | unterminated single-quoted string → reject | R027,R010,R011 |
| `TC-092` | `NEG` | lexical/dynamic/client | unterminated quoted identifier → reject | R027,R010,R011 |
| `TC-093` | `NEG` | lexical/dynamic/client | unterminated block comment → reject | R027,R010,R011 |
| `TC-094` | `NEG` | lexical/dynamic/client | unterminated dollar quote → reject | R027,R010,R011 |
| `TC-095` | `NEG` | lexical/dynamic/client | empty executable statement/semicolon noise violating canonical sequence → reject | R027,R010,R011 |
| `TC-096` | `NEG` | lexical/dynamic/client | missing required final semicolon → reject | R027,R010,R011 |
| `TC-097` | `NEG` | lexical/dynamic/client | DO block → reject | R027,R010,R011 |
| `TC-098` | `NEG` | lexical/dynamic/client | CREATE FUNCTION → reject | R027,R010,R011 |
| `TC-099` | `NEG` | lexical/dynamic/client | CREATE PROCEDURE → reject | R027,R010,R011 |
| `TC-100` | `NEG` | lexical/dynamic/client | CALL → reject | R027,R010,R011 |
| `TC-101` | `NEG` | lexical/dynamic/client | PREPARE → reject | R027,R010,R011 |
| `TC-102` | `NEG` | lexical/dynamic/client | EXECUTE prepared/dynamic statement → reject | R027,R010,R011 |
| `TC-103` | `NEG` | lexical/dynamic/client | concatenated dynamic SQL execution surface → reject | R027,R010,R011 |
| `TC-104` | `NEG` | lexical/dynamic/client | \i psql include → reject | R027,R010,R011 |
| `TC-105` | `NEG` | lexical/dynamic/client | \ir psql include-relative → reject | R027,R010,R011 |
| `TC-106` | `NEG` | lexical/dynamic/client | \gexec → reject | R027,R010,R011 |
| `TC-107` | `NEG` | lexical/dynamic/client | \copy → reject | R027,R010,R011 |
| `TC-108` | `NEG` | lexical/dynamic/client | \! shell command → reject | R027,R010,R011 |
| `TC-109` | `NEG` | lexical/dynamic/client | other psql backslash meta-command → reject | R027,R010,R011 |
| `TC-110` | `NEG` | lexical/dynamic/client | psql variable substitution → reject | R027,R010,R011 |
| `TC-111` | `NEG` | lexical/dynamic/client | SAVEPOINT → reject | R027,R010,R011 |
| `TC-112` | `NEG` | lexical/dynamic/client | COMMIT AND CHAIN → reject | R027,R010,R011 |
| `TC-113` | `NEG` | lexical/dynamic/client | ROLLBACK TO SAVEPOINT → reject | R027,R010,R011 |
| `TC-114` | `NEG` | up-allowlist/safety | INSERT VALUES in enforced up → reject | R008,R010,R029,R030,R019 |
| `TC-115` | `NEG` | up-allowlist/safety | INSERT SELECT → reject | R008,R010,R029,R030,R019 |
| `TC-116` | `NEG` | up-allowlist/safety | UPDATE → reject | R008,R010,R029,R030,R019 |
| `TC-117` | `NEG` | up-allowlist/safety | DELETE → reject | R008,R010,R029,R030,R019 |
| `TC-118` | `NEG` | up-allowlist/safety | MERGE → reject | R008,R010,R029,R030,R019 |
| `TC-119` | `NEG` | up-allowlist/safety | COPY FROM → reject | R008,R010,R029,R030,R019 |
| `TC-120` | `NEG` | up-allowlist/safety | COPY TO/PROGRAM → reject | R008,R010,R029,R030,R019 |
| `TC-121` | `NEG` | up-allowlist/safety | CREATE TABLE AS SELECT → reject | R008,R010,R029,R030,R019 |
| `TC-122` | `NEG` | up-allowlist/safety | SELECT INTO → reject | R008,R010,R029,R030,R019 |
| `TC-123` | `NEG` | up-allowlist/safety | CREATE MATERIALIZED VIEW AS SELECT → reject | R008,R010,R029,R030,R019 |
| `TC-124` | `NEG` | up-allowlist/safety | REFRESH MATERIALIZED VIEW → reject | R008,R010,R029,R030,R019 |
| `TC-125` | `NEG` | up-allowlist/safety | unknown executable statement class → reject | R008,R010,R029,R030,R019 |
| `TC-126` | `NEG` | up-allowlist/safety | CREATE SCHEMA in paired-SQL v1 → reject | R008,R010,R029,R030,R019 |
| `TC-127` | `NEG` | up-allowlist/safety | CREATE TABLE LIKE → reject | R008,R010,R029,R030,R019 |
| `TC-128` | `NEG` | up-allowlist/safety | CREATE TABLE inheritance/partition attachment unsupported form → reject | R008,R010,R029,R030,R019 |
| `TC-129` | `NEG` | up-allowlist/safety | ADD COLUMN function/expression default → reject | R008,R010,R029,R030,R019 |
| `TC-130` | `NEG` | up-allowlist/safety | ADD COLUMN generated/identity semantics → reject | R008,R010,R029,R030,R019 |
| `TC-131` | `NEG` | up-allowlist/safety | ADD COLUMN NOT NULL without safe literal default on pre-existing table → reject | R008,R010,R029,R030,R019 |
| `TC-132` | `NEG` | up-allowlist/safety | ADD CONSTRAINT CHECK without NOT VALID on pre-existing table → reject | R008,R010,R029,R030,R019 |
| `TC-133` | `NEG` | up-allowlist/safety | ADD FOREIGN KEY without NOT VALID on pre-existing table → reject | R008,R010,R029,R030,R019 |
| `TC-134` | `NEG` | up-allowlist/safety | unsupported ADD CONSTRAINT form → reject | R008,R010,R029,R030,R019 |
| `TC-135` | `NEG` | up-allowlist/safety | unsupported ALTER TABLE subcommand → reject | R008,R010,R029,R030,R019 |
| `TC-136` | `NEG` | up-allowlist/safety | expression index → reject | R008,R010,R029,R030,R019 |
| `TC-137` | `NEG` | up-allowlist/safety | partial index predicate → reject | R008,R010,R029,R030,R019 |
| `TC-138` | `NEG` | up-allowlist/safety | index on pre-existing table without CONCURRENTLY → reject | R008,R010,R029,R030,R019 |
| `TC-139` | `NEG` | up-allowlist/safety | quoted object identifier in enforced DDL → reject | R008,R010,R029,R030,R019 |
| `TC-140` | `NEG` | up-allowlist/safety | unqualified object identity where schema qualification required → reject | R008,R010,R029,R030,R019 |
| `TC-141` | `NEG` | up-allowlist/safety | SET search_path → reject | R008,R010,R029,R030,R019 |
| `TC-142` | `NEG` | up-allowlist/safety | binary FLOAT/REAL/DOUBLE PRECISION type → reject | R008,R010,R029,R030,R019 |
| `TC-143` | `NEG` | up-allowlist/safety | local-timezone conversion construct → reject | R008,R010,R029,R030,R019 |
| `TC-144` | `NEG` | ddl-impact | up DDL missing impact entry → reject | R013,R029 |
| `TC-145` | `NEG` | ddl-impact | down inverse DDL missing impact entry → reject | R013,R029 |
| `TC-146` | `NEG` | ddl-impact | orphan impact entry → reject | R013,R029 |
| `TC-147` | `NEG` | ddl-impact | duplicate impact entry → reject | R013,R029 |
| `TC-148` | `NEG` | ddl-impact | statement_sha256 mismatch → reject | R013,R029 |
| `TC-149` | `NEG` | ddl-impact | impact count differs from DDL count → reject | R013,R029 |
| `TC-150` | `NEG` | ddl-impact | unknown statement_class in impact entry → reject | R013,R029 |
| `TC-151` | `NEG` | ddl-impact | estimated_lock_mode outside current frozen exact eight-value enum → reject | R013,R029,R036 |
| `TC-152` | `NEG` | ddl-impact | negative affected_rows_estimate → reject | R013,R029 |
| `TC-153` | `NEG` | ddl-impact | negative disk estimate → reject | R013,R029 |
| `TC-154` | `NEG` | ddl-impact | negative WAL estimate → reject | R013,R029 |
| `TC-155` | `NEG` | ddl-impact | replication_impact outside exact none/low/medium/high enum → reject | R013,R029,R036 |
| `TC-156` | `NEG` | ddl-impact | blank abort_condition → reject | R013,R029 |
| `TC-157` | `NEG` | ddl-impact | blank estimate_basis → reject | R013,R029 |
| `TC-158` | `NEG` | ddl-impact | two DDL statements share one impact entry → reject | R013,R029 |
| `TC-159` | `NEG` | ddl-impact | index online_strategy missing → reject | R013,R029 |
| `TC-160` | `NEG` | ddl-impact | pre-existing-table index declares non-concurrent strategy → reject | R013,R029 |
| `TC-161` | `NEG` | ddl-impact | non-index DDL declares index online strategy → reject | R013,R029 |
| `TC-162` | `NEG` | observability | missing observability category → reject | R014 |
| `TC-163` | `NEG` | observability | unknown observability category → reject | R014 |
| `TC-164` | `NEG` | observability | duplicate observability category → reject | R014 |
| `TC-165` | `NEG` | observability | measured category without signal_or_method → reject | R014 |
| `TC-166` | `NEG` | observability | measured category with N/A reason set → reject | R014 |
| `TC-167` | `NEG` | observability | N/A category without reason → reject | R014 |
| `TC-168` | `NEG` | observability | N/A category with signal_or_method set → reject | R014 |
| `TC-169` | `NEG` | observability | lock_wait marked N/A for Expand → reject | R014 |
| `TC-170` | `NEG` | observability | statement_duration marked N/A for Expand → reject | R014 |
| `TC-171` | `NEG` | observability | replication_lag marked N/A for Expand → reject | R014 |
| `TC-172` | `NEG` | observability | wal_growth marked N/A for Expand → reject | R014 |
| `TC-173` | `NEG` | observability | disk_growth marked N/A for Expand → reject | R014 |
| `TC-174` | `NEG` | observability | retry_pause_abort_reason marked N/A for Expand → reject | R014 |
| `TC-175` | `NEG` | observability | change_deployment_correlation marked N/A for Expand → reject | R014 |
| `TC-176` | `NEG` | observability | generic monitoring signals without canonical observability profile → reject | R014 |
| `TC-177` | `NEG` | risk/rollout/authority/classification | medium risk with rollout.mode!=staged → reject | R015,R016,R017,R019,R025 |
| `TC-178` | `NEG` | risk/rollout/authority/classification | medium risk missing rollout metrics → reject | R015,R016,R017,R019,R025 |
| `TC-179` | `NEG` | risk/rollout/authority/classification | medium risk missing staged_rollout ref → reject | R015,R016,R017,R019,R025 |
| `TC-180` | `NEG` | risk/rollout/authority/classification | high risk with rollout.mode!=staged → reject | R015,R016,R017,R019,R025 |
| `TC-181` | `NEG` | risk/rollout/authority/classification | high risk missing ADR ref → reject | R015,R016,R017,R019,R025 |
| `TC-182` | `NEG` | risk/rollout/authority/classification | high risk missing security/privacy ref → reject | R015,R016,R017,R019,R025 |
| `TC-183` | `NEG` | risk/rollout/authority/classification | high risk missing golden-vectors ref → reject | R015,R016,R017,R019,R025 |
| `TC-184` | `NEG` | risk/rollout/authority/classification | high risk missing restore-rehearsal ref → reject | R015,R016,R017,R019,R025 |
| `TC-185` | `NEG` | risk/rollout/authority/classification | identity_personal classification missing security/privacy ref → reject | R015,R016,R017,R019,R025 |
| `TC-186` | `NEG` | risk/rollout/authority/classification | sensitive classification missing security/privacy ref → reject | R015,R016,R017,R019,R025 |
| `TC-187` | `NEG` | risk/rollout/authority/classification | mixed classification missing security/privacy ref → reject | R015,R016,R017,R019,R025 |
| `TC-188` | `NEG` | risk/rollout/authority/classification | unknown authority ref kind → reject | R015,R016,R017,R019,R025 |
| `TC-189` | `NEG` | risk/rollout/authority/classification | authority ref absolute path → reject | R015,R016,R017,R019,R025 |
| `TC-190` | `NEG` | risk/rollout/authority/classification | authority ref .. traversal → reject | R015,R016,R017,R019,R025 |
| `TC-191` | `NEG` | risk/rollout/authority/classification | authority ref target symlink → reject | R015,R016,R017,R019,R025 |
| `TC-192` | `NEG` | risk/rollout/authority/classification | authority ref content hash mismatch on introduction → reject | R015,R016,R017,R019,R025 |
| `TC-193` | `NEG` | risk/rollout/authority/classification | duplicate exact authority identity `(kind,path)` → reject | R015,R016,R017,R019,R025 |
| `TC-194` | `NEG` | risk/rollout/authority/classification | owners omit touched schema → reject | R015,R016,R017,R019,R025 |
| `TC-195` | `NEG` | risk/rollout/authority/classification | owners contain untouched schema → reject | R015,R016,R017,R019,R025 |
| `TC-196` | `NEG` | down-inverse | down DROP TABLE targets unrelated pre-existing table → reject | R009,R028,R010,R011 |
| `TC-197` | `NEG` | down-inverse | down DROP COLUMN targets column not added by same up → reject | R009,R028,R010,R011 |
| `TC-198` | `NEG` | down-inverse | down DROP CONSTRAINT targets constraint not added by same up → reject | R009,R028,R010,R011 |
| `TC-199` | `NEG` | down-inverse | down DROP INDEX targets index not created by same up → reject | R009,R028,R010,R011 |
| `TC-200` | `NEG` | down-inverse | down target has correct name but wrong schema → reject | R009,R028,R010,R011 |
| `TC-201` | `NEG` | down-inverse | down uses CASCADE → reject | R009,R028,R010,R011 |
| `TC-202` | `NEG` | down-inverse | down uses IF EXISTS → reject | R009,R028,R010,R011 |
| `TC-203` | `NEG` | down-inverse | down contains DML → reject | R009,R028,R010,R011 |
| `TC-204` | `NEG` | down-inverse | down contains procedural/dynamic SQL → reject | R009,R028,R010,R011 |
| `TC-205` | `NEG` | down-inverse | down contains psql command/substitution → reject | R009,R028,R010,R011 |
| `TC-206` | `NEG` | down-inverse | multi-effect up missing one inverse → reject | R009,R028,R010,R011 |
| `TC-207` | `NEG` | down-inverse | down contains duplicate inverse → reject | R009,R028,R010,R011 |
| `TC-208` | `NEG` | down-inverse | down contains extra/orphan inverse → reject | R009,R028,R010,R011 |
| `TC-209` | `NEG` | down-inverse | down inverse order violates derived dependencies → reject | R009,R028,R010,R011 |
| `TC-210` | `NEG` | down-inverse | transactional/non-transactional inverse classes mixed in one down direction → reject | R009,R028,R010,R011 |
| `TC-211` | `NEG` | pr-base/immutability | PR mode without base SHA → reject | R006,R004 |
| `TC-212` | `NEG` | pr-base/immutability | malformed base SHA → reject | R006,R004 |
| `TC-213` | `NEG` | pr-base/immutability | base SHA not resolvable → reject | R006,R004 |
| `TC-214` | `NEG` | pr-base/immutability | base not supported ancestor → reject | R006,R004 |
| `TC-215` | `NEG` | pr-base/immutability | modified existing SQL → reject | R006,R004 |
| `TC-216` | `NEG` | pr-base/immutability | deleted existing SQL → reject | R006,R004 |
| `TC-217` | `NEG` | pr-base/immutability | renamed existing SQL → reject | R006,R004 |
| `TC-218` | `NEG` | pr-base/immutability | old SQL modified + candidate checksum updated → reject | R006,R004 |
| `TC-219` | `NEG` | pr-base/immutability | base-existing manifest entry modified with SQL unchanged → reject | R006,R004 |
| `TC-220` | `NEG` | pr-base/immutability | base-existing manifest entry removed → reject | R006,R004 |
| `TC-221` | `NEG` | pr-base/immutability | base-existing legacy baseline expanded → reject | R006,R004 |
| `TC-222` | `NEG` | pr-base/immutability | base-existing authority ref digest modified → reject | R006,R004 |
| `TC-223` | `NEG` | policy/test-contract | required S2 control ID missing → reject | R021,R022 |
| `TC-224` | `NEG` | policy/test-contract | duplicate S2 control ID → reject | R021,R022 |
| `TC-225` | `NEG` | policy/test-contract | unknown S2 control ID → reject | R021,R022 |
| `TC-226` | `NEG` | policy/test-contract | policy row uses non-enum primary disposition → reject | R021,R022 |
| `TC-227` | `NEG` | policy/test-contract | policy row has multiple primary dispositions → reject | R021,R022 |
| `TC-228` | `NEG` | policy/test-contract | MACHINE row missing machine rule mapping → reject | R021,R022 |
| `TC-229` | `NEG` | policy/test-contract | STRUCTURE_PLUS_HUMAN_ADEQUACY row missing structural rule or human gate → reject | R021,R022 |
| `TC-230` | `NEG` | policy/test-contract | OPERATIONAL_OR_CLOSURE_EVIDENCE row missing owner/gate → reject | R021,R022 |
| `TC-231` | `NEG` | policy/test-contract | REJECTED_BY_PAIRED_SQL_V1 row missing stable rejection rule → reject | R021,R022 |
| `TC-232` | `NEG` | policy/test-contract | DEFERRED row missing future scope owner/gate → reject | R021,R022 |
| `TC-233` | `NEG` | policy/test-contract | rule ID missing from rule registry → reject | R021,R022 |
| `TC-234` | `NEG` | policy/test-contract | required TC ID missing → reject | R021,R022 |
| `TC-235` | `NEG` | policy/test-contract | duplicate TC ID → reject | R021,R022 |
| `TC-236` | `NEG` | policy/test-contract | unknown/stale TC ID in mapping → reject | R021,R022 |
| `TC-237` | `NEG` | policy/test-contract | allowed branch missing positive case mapping → reject | R021,R022 |
| `TC-238` | `NEG` | policy/test-contract | machine rule missing negative coverage when it has rejection path → reject | R021,R022 |
| `TC-239` | `NEG` | ci-rehearsal/regression | first full apply fails → CI fail | R023,R024 |
| `TC-240` | `NEG` | ci-rehearsal/regression | reverse down rehearsal fails → CI fail | R023,R024 |
| `TC-241` | `NEG` | ci-rehearsal/regression | rollback baseline contains unexpected managed object → CI fail | R023,R024 |
| `TC-242` | `NEG` | ci-rehearsal/regression | full reapply fails → CI fail | R023,R024 |
| `TC-243` | `NEG` | ci-rehearsal/regression | first/reapply catalog invariant mismatch → CI fail | R023,R024 |
| `TC-244` | `NEG` | ci-rehearsal/regression | runtime append-only privilege regression → CI fail | R023,R024 |
| `TC-245` | `NEG` | ci-rehearsal/regression | Stage 3.1 structural guard disappears without equivalent replacement → reject | R023,R024 |
| `TC-246` | `NEG` | ci-rehearsal/regression | Stage 3.11 structural guard disappears without equivalent replacement → reject | R023,R024 |
| `TC-247` | `POS` | positive-acceptance | valid low-risk Expand using only safe empty-table/additive-column effects with required observability/timeouts/impact → pass | R016,R025,R034 |
| `TC-248` | `POS` | positive-acceptance | valid medium-risk Expand with staged rollout/ref/metrics → pass | R016,R025 |
| `TC-249` | `POS` | positive-acceptance | valid high-risk Expand with staged rollout + ADR/security/golden/restore refs → pass structural validation | R015,R016,R025 |
| `TC-250` | `POS` | positive-acceptance | valid schema_only classification on appropriate structural change → pass | R017 |
| `TC-251` | `POS` | positive-acceptance | valid financial classification with applicable review structure → pass | R017 |
| `TC-252` | `POS` | positive-acceptance | valid identity_personal classification + security/privacy ref → pass | R017,R015 |
| `TC-253` | `POS` | positive-acceptance | valid sensitive classification + security/privacy ref → pass | R017,R015 |
| `TC-254` | `POS` | positive-acceptance | valid mixed classification + security/privacy ref → pass | R017,R015 |
| `TC-255` | `POS` | positive-acceptance | valid Expand profile with measured required categories and N/A only for rows/batches + validation mismatches → pass | R014 |
| `TC-256` | `POS` | positive-acceptance | valid measured observability object → pass | R014 |
| `TC-257` | `POS` | positive-acceptance | valid rows_or_batches N/A with category-specific reason → pass | R014 |
| `TC-258` | `POS` | positive-acceptance | valid validation_mismatches N/A with category-specific reason → pass | R014 |
| `TC-259` | `POS` | positive-acceptance | valid migration with no declared dependency where reviewer confirms none required → pass graph validation | R018 |
| `TC-260` | `POS` | positive-acceptance | valid dependency on frozen legacy ID → pass | R018 |
| `TC-261` | `POS` | positive-acceptance | valid dependency on earlier enforced non-immediate ID → pass | R018 |
| `TC-262` | `POS` | positive-acceptance | valid monotonic migration ID gap → pass | R002 |
| `TC-263` | `POS` | positive-acceptance | valid CREATE TABLE without unsupported subforms → pass | R008 |
| `TC-264` | `POS` | positive-acceptance | valid ADD COLUMN nullable/no default → pass | R008 |
| `TC-265` | `POS` | positive-acceptance | valid nullable ADD COLUMN with frozen safe literal default, risk=medium and complete medium gates → pass | R008,R034,R016,R025 |
| `TC-266` | `POS` | positive-acceptance | valid ADD CHECK ... NOT VALID using exactly one frozen atomic IS NULL/IS NOT NULL/comparison predicate, risk=medium and complete medium gates → pass | R008,R033,R034,R016,R025,R036 |
| `TC-267` | `POS` | positive-acceptance | valid ADD FOREIGN KEY ... NOT VALID with frozen safe FK semantics, risk=medium and complete medium gates → pass | R008,R033,R034,R016,R025 |
| `TC-268` | `POS` | positive-acceptance | valid non-concurrent index on table created earlier in same up with risk=medium and complete medium gates → pass | R008,R029,R034,R016,R025 |
| `TC-269` | `POS` | positive-acceptance | valid CREATE INDEX CONCURRENTLY on pre-existing table with risk=medium and complete medium gates → pass | R008,R029,R011,R034,R016,R025 |
| `TC-270` | `POS` | positive-acceptance | valid CREATE UNIQUE INDEX CONCURRENTLY on pre-existing table with risk=medium and complete medium gates → pass | R008,R029,R011,R034,R016,R025 |
| `TC-271` | `POS` | positive-acceptance | valid multi-effect Expand with all allowlisted statements → pass | R008,R028 |
| `TC-272` | `POS` | positive-acceptance | matching disposable DROP TABLE inverse → pass | R009,R028 |
| `TC-273` | `POS` | positive-acceptance | matching disposable DROP COLUMN inverse → pass | R009,R028 |
| `TC-274` | `POS` | positive-acceptance | matching disposable DROP CONSTRAINT inverse → pass | R009,R028 |
| `TC-275` | `POS` | positive-acceptance | matching transactional DROP INDEX inverse → pass | R009,R028 |
| `TC-276` | `POS` | positive-acceptance | matching non-transactional DROP INDEX CONCURRENTLY inverse → pass | R009,R028,R011 |
| `TC-277` | `POS` | positive-acceptance | complete exact multi-effect down bijection in reverse dependency order → pass | R009,R028 |
| `TC-278` | `POS` | positive-acceptance | transactional up exact SET LOCAL timeout controls matching manifest before DDL → pass | R012,R011 |
| `TC-279` | `POS` | positive-acceptance | transactional down exact SET LOCAL timeout controls matching manifest before inverse DDL → pass | R012,R011 |
| `TC-280` | `POS` | positive-acceptance | concurrent-index up exact session SET timeouts matching manifest → pass | R012,R011 |
| `TC-281` | `POS` | positive-acceptance | concurrent-index down exact session SET timeouts matching manifest → pass | R012,R011 |
| `TC-282` | `POS` | positive-acceptance | valid statement-bound up_ddl_impact bijection using exact raw-statement byte-slice SHA-256 algorithm → pass | R013,R036 |
| `TC-283` | `POS` | positive-acceptance | valid statement-bound down_ddl_impact bijection → pass | R013 |
| `TC-284` | `POS` | positive-acceptance | declared single schema owner exactly matches touched-schema set → pass | R019 |
| `TC-285` | `POS` | positive-acceptance | declared multiple owners exactly match multi-schema touched set → pass | R019 |
| `TC-286` | `POS` | positive-acceptance | new typed authority ref with exact canonical authority path and matching content hash → pass | R015 |
| `TC-287` | `POS` | positive-acceptance | exact seven real legacy pairs + identity-only baseline → pass | R004,R005 |
| `TC-288` | `POS` | positive-acceptance | valid PR mode with resolvable ancestor base and unchanged historical SQL/manifest → pass | R006 |
| `TC-289` | `POS` | positive-acceptance | valid local mode performs self-consistency and explicitly reports base comparison N/A → pass | R005 |
| `TC-290` | `POS` | positive-acceptance | real current apply→reverse-down→baseline→reapply→equivalence→runtime privilege chain → pass | R023,R024 |
| `TC-291` | `POS` | positive-acceptance | phase=expand with otherwise valid entry → pass | R007 |
| `TC-292` | `POS` | positive-acceptance | safe numeric/decimal literal ADD COLUMN default → pass | R008 |
| `TC-293` | `POS` | positive-acceptance | safe boolean literal ADD COLUMN default → pass | R008 |
| `TC-294` | `POS` | positive-acceptance | safe ordinary quoted-string literal ADD COLUMN default → pass | R008 |
| `TC-295` | `POS` | positive-acceptance | valid CREATE UNIQUE INDEX non-concurrent on table created in same up with risk=medium and complete medium gates → pass | R008,R029,R034,R016,R025 |
| `TC-296` | `POS` | positive-acceptance | production_rollback.strategy=application_or_config_rollback with complete procedure/verification → pass | R026 |
| `TC-297` | `POS` | positive-acceptance | production_rollback.strategy=leave_additive_structure_unused with complete procedure/verification → pass | R026 |
| `TC-298` | `POS` | positive-acceptance | parameterized acceptance of exactly access_share,row_share,row_exclusive,share_update_exclusive,share,share_row_exclusive,exclusive,access_exclusive → pass | R013,R036 |
| `TC-299` | `POS` | positive-acceptance | parameterized acceptance of exactly none,low,medium,high replication_impact values → pass | R013,R036 |
| `TC-300` | `POS` | positive-acceptance | parameterized acceptance of identity/investment/analytics/audit owner when it exactly matches touched schema → pass | R019 |
| `TC-301` | `POS` | positive-acceptance | exact canonical source registry `S2-001…S2-168` is present and exercises all six frozen primary disposition values → pass | R021 |
| `TC-302` | `POS` | positive-acceptance | reversibility=disposable_down_exact_inverse with valid exact inverse pair → pass | R009,R028 |
| `TC-303` | `NEG` | error/log-safety | validator error emitted without stable code/rule context → focused test fail | R031 |
| `TC-304` | `NEG` | error/log-safety | validator logs SQL row/personal/financial-document content in synthetic redaction fixture → reject regression | R020,R031 |
| `TC-305` | `NEG` | error/log-safety | two rule failures collapse to ambiguous generic success/failure without category identity → reject regression | R031 |
| `TC-306` | `POS` | positive-acceptance | parameterized acceptance of every frozen up/down execution lock_risk enum value low/medium/high in otherwise valid fixtures → pass | R012 |


| `TC-307` | `NEG` | up-transaction-class | transactional up mixes ordinary DDL with CREATE INDEX CONCURRENTLY → reject | R032,R011,R008 |
| `TC-308` | `NEG` | up-transaction-class | non-transactional up contains CREATE TABLE/ALTER TABLE/non-concurrent index DDL → reject | R032,R011,R008 |
| `TC-309` | `NEG` | up-transaction-class | one up direction requires both transactional and non-transactional effect classes → reject/rescope | R032,R011 |
| `TC-310` | `NEG` | ddl-subgrammar | CREATE TABLE IF NOT EXISTS → reject | R033,R008 |
| `TC-311` | `NEG` | ddl-subgrammar | ALTER TABLE ADD COLUMN IF NOT EXISTS → reject | R033,R008 |
| `TC-312` | `NEG` | ddl-subgrammar | CREATE INDEX IF NOT EXISTS → reject | R033,R008 |
| `TC-313` | `NEG` | ddl-subgrammar | CREATE TABLE TEMP/UNLOGGED/ON COMMIT/TABLESPACE/partition/inheritance/LIKE/AS unsupported family → reject | R033,R008 |
| `TC-314` | `NEG` | ddl-subgrammar | CREATE TABLE column uses function/expression default → reject | R033,R008 |
| `TC-315` | `NEG` | ddl-subgrammar | CREATE TABLE column uses SERIAL/identity/generated/custom/domain/array type semantics → reject | R033,R008 |
| `TC-316` | `NEG` | ddl-subgrammar | CREATE TABLE CHECK uses function/subquery/cast/unsupported expression → reject | R033,R008 |
| `TC-317` | `NEG` | ddl-subgrammar | ALTER ADD CHECK uses function/subquery/cast/unsupported expression → reject | R033,R008 |
| `TC-318` | `NEG` | ddl-subgrammar | CREATE/ADD FK uses CASCADE/SET NULL/SET DEFAULT or unsupported referential action → reject | R033,R008 |
| `TC-319` | `NEG` | ddl-subgrammar | CREATE/ADD FK uses DEFERRABLE or unsupported MATCH/options → reject | R033,R008 |
| `TC-320` | `NEG` | risk-floor | supported ADD CONSTRAINT declared risk=low → reject | R034,R016,R025 |
| `TC-321` | `NEG` | risk-floor | supported CREATE INDEX/UNIQUE INDEX declared risk=low → reject | R034,R016,R025 |
| `TC-322` | `NEG` | risk-floor | author-declared risk below machine-derived operation minimum → reject | R034,R016,R025 |
| `TC-323` | `POS` | positive-acceptance | safe ADD CONSTRAINT NOT VALID with risk=medium and complete medium gates → pass | R034,R016,R025,R008 |
| `TC-324` | `POS` | positive-acceptance | safe supported index with risk=medium and complete medium gates → pass | R034,R016,R025,R008,R029 |
| `TC-325` | `NEG` | ci-validation-dominance | migration-executing Go job has no successful migrations-job dependency/validated SHA → execution gate fails | R035 |
| `TC-326` | `NEG` | ci-validation-dominance | validated_sha differs from GITHUB_SHA before dependent migration apply → execution gate fails | R035 |
| `TC-327` | `POS` | positive-acceptance | migrations job succeeds on exact SHA, publishes validated_sha, dependent Go job exact-SHA assertion passes before SQL apply | R035 |
| `TC-328` | `POS` | positive-acceptance | homogeneous transactional up with only ordinary allowlisted DDL → pass | R032,R011,R008 |
| `TC-329` | `POS` | positive-acceptance | homogeneous non-transactional up with only concurrent-index effects and exact session timeouts → pass | R032,R011,R008,R029,R012 |
| `TC-330` | `POS` | positive-acceptance | parameterized constraint-free nullable CREATE TABLE acceptance across every current frozen built-in type form and permitted literal-default combination → pass | R033,R008,R036 |
| `TC-331` | `POS` | positive-acceptance | constraint-free nullable CREATE TABLE or nullable ADD COLUMN **without DEFAULT** may remain risk=low with complete low-risk controls → pass | R034,R016 |
| `TC-332` | `NEG` | execution/timeouts | non-transactional concurrent-index DOWN missing session SET lock_timeout before inverse DDL → reject | R012,R011 |
| `TC-333` | `NEG` | execution/timeouts | non-transactional concurrent-index DOWN missing session SET statement_timeout before inverse DDL → reject | R012,R011 |
| `TC-334` | `NEG` | ddl-subgrammar/risk-floor | CREATE TABLE contains inline/table CHECK constraint → reject; constraint must use ADD CONSTRAINT NOT VALID | R033,R034,R008 |
| `TC-335` | `NEG` | ddl-subgrammar/risk-floor | CREATE TABLE contains REFERENCES/FOREIGN KEY constraint → reject; FK must use ADD CONSTRAINT NOT VALID | R033,R034,R008 |
| `TC-336` | `NEG` | ddl-subgrammar/risk-floor | CREATE TABLE contains PRIMARY KEY/UNIQUE/EXCLUDE/CONSTRAINT effect → reject | R033,R034,R008 |
| `TC-337` | `NEG` | ci-validation-dominance | `go-race` migration apply has no successful `needs: migrations` exact-SHA dependency → execution gate fails | R035 |
| `TC-338` | `NEG` | ci-validation-dominance | `go-race` validated_sha differs from GITHUB_SHA before migration apply → execution gate fails | R035 |
| `TC-339` | `POS` | positive-acceptance | `go-race` depends on successful migrations job and exact validated_sha assertion passes before its SQL apply → pass | R035 |
| `TC-340` | `NEG` | ci-validation-dominance | workflow contains migration-application marker in a job outside frozen `{migrations,go,go-race}` inventory → CI-policy test fails closed | R035 |
| `TC-341` | `POS` | positive-acceptance | exact current migration-SQL inventory is `{migrations,go,go-race}`; `migrations` validates first and both Go jobs are exact-SHA dominated → pass | R035 |
| `TC-342` | `NEG` | frozen-semantics | implementation accepts timeout/duration beyond current frozen BOUND-SPEC numeric bounds → contract test fails | R036,R012 |
| `TC-343` | `NEG` | frozen-semantics | implementation accepts a CREATE/ADD type form outside current frozen built-in allowlist or an alias not listed → reject | R036,R033,R008 |
| `TC-344` | `NEG` | frozen-semantics | implementation accepts CHECK syntax outside exact atomic predicate/operator grammar or FK options outside exact default semantics → reject | R036,R033,R008 |
| `TC-345` | `POS` | positive-acceptance | current frozen semantic constants/hash vectors (bounds, type allowlist, typed literal/CHECK/FK grammar, raw statement SHA algorithm) are accepted exactly and no broader variant is implied → pass | R036,R012,R013,R033,R040 |
| `TC-346` | `NEG` | policy-source-extraction | canonical Low-risk `new empty table` source row missing → policy coverage fail | R021,R038 |
| `TC-347` | `NEG` | policy-source-extraction | canonical Low-risk `additive nullable column` source row missing → policy coverage fail | R021,R038 |
| `TC-348` | `NEG` | policy-source-extraction | canonical Destructive `DROP` classification anchor missing → policy coverage fail | R021,R038 |
| `TC-349` | `NEG` | policy-source-extraction | canonical Destructive `irreversible conversion` classification anchor missing → policy coverage fail | R021,R038 |
| `TC-350` | `NEG` | policy-source-extraction | canonical Destructive `history rewrite` classification anchor missing → policy coverage fail | R021,R038 |
| `TC-351` | `NEG` | policy-source-extraction | P3-08-derived hardening is placed under `S2-*` source namespace → reject registry | R038 |
| `TC-352` | `NEG` | policy-source-extraction | `P3D-*` derived set missing/duplicate/unknown ID → reject registry | R038 |
| `TC-353` | `POS` | positive-acceptance | exact source registry `S2-001…S2-168` plus separate derived registry `P3D-001…P3D-027` with no namespace mixing → pass | R021,R038 |
| `TC-354` | `NEG` | literal-grammar | numeric/integer default uses leading `+` → reject | R040,R033 |
| `TC-355` | `NEG` | literal-grammar | numeric/integer default uses forbidden leading zero such as `01` → reject | R040,R033 |
| `TC-356` | `NEG` | literal-grammar | numeric default uses `.5` without canonical leading zero → reject | R040,R033 |
| `TC-357` | `NEG` | literal-grammar | numeric default uses trailing decimal point `5.` → reject | R040,R033 |
| `TC-358` | `NEG` | literal-grammar | integer/numeric literal is negative zero (`-0`, `-0.0`, equivalent all-zero magnitude) → reject | R040,R033 |
| `TC-359` | `NEG` | literal-grammar | smallint literal outside `-32768..32767` → reject | R040 |
| `TC-360` | `NEG` | literal-grammar | integer literal outside `-2147483648..2147483647` → reject | R040 |
| `TC-361` | `NEG` | literal-grammar | bigint literal outside signed 64-bit PostgreSQL range → reject | R040 |
| `TC-362` | `NEG` | literal-grammar | numeric(p,s) integer significant digits exceed `p-s` → reject before PostgreSQL rounding/error | R040 |
| `TC-363` | `NEG` | literal-grammar | numeric(p,s) fractional digits exceed `s` → reject before PostgreSQL rounding | R040 |
| `TC-364` | `NEG` | literal-grammar | ordinary string default contains backslash escape surface → reject | R040 |
| `TC-365` | `NEG` | literal-grammar | ordinary string default contains control/newline/tab/NUL or non-ASCII payload → reject | R040 |
| `TC-366` | `NEG` | literal-grammar | decoded varchar(n) default payload length exceeds n → reject | R040 |
| `TC-367` | `POS` | positive-acceptance | exact canonical integer/numeric default boundary vectors within target type/precision/scale → pass | R040,R033 |
| `TC-368` | `POS` | positive-acceptance | ordinary ASCII string with doubled embedded quote and varchar decoded length exactly n → pass | R040,R033 |
| `TC-369` | `NEG` | check-grammar | CHECK targets pre-existing/untyped column rather than same-UP known-type ADD COLUMN → reject/rescope | R040,R033 |
| `TC-370` | `NEG` | check-grammar | boolean CHECK uses ordering operator `<`, `<=`, `>`, or `>=` → reject | R040,R033 |
| `TC-371` | `NEG` | check-grammar | text/varchar CHECK uses ordering operator rather than `=`/`<>` → reject | R040,R033 |
| `TC-372` | `NEG` | check-grammar | uuid/date/timestamp comparison CHECK is attempted in paired-SQL v1 → reject/rescope | R040,R033 |
| `TC-373` | `POS` | positive-acceptance | same-UP integer/numeric column with exact typed atomic CHECK literal/operator → pass | R040,R033,R008 |
| `TC-374` | `POS` | positive-acceptance | same-UP boolean/text/varchar column with exact supported atomic CHECK literal/operator → pass | R040,R033,R008 |
| `TC-375` | `NEG` | index-grammar | expression/function index key → reject | R041,R033,R008 |
| `TC-376` | `NEG` | index-grammar | partial index `WHERE` predicate → reject | R041,R033,R008 |
| `TC-377` | `NEG` | index-grammar | index `INCLUDE` clause → reject | R041,R033,R008 |
| `TC-378` | `NEG` | index-grammar | explicit `USING` method → reject | R041,R033,R008 |
| `TC-379` | `NEG` | index-grammar | index COLLATE/operator-class/operator-class-options → reject | R041,R033,R008 |
| `TC-380` | `NEG` | index-grammar | index key ASC/DESC sort option → reject | R041,R033,R008 |
| `TC-381` | `NEG` | index-grammar | NULLS FIRST/LAST or NULLS [NOT] DISTINCT index option → reject | R041,R033,R008 |
| `TC-382` | `NEG` | index-grammar | index storage `WITH (...)` clause → reject | R041,R033,R008 |
| `TC-383` | `NEG` | index-grammar | index TABLESPACE clause → reject | R041,R033,R008 |
| `TC-384` | `NEG` | index-grammar | `ON ONLY` / ONLY index target → reject | R041,R033,R008 |
| `TC-385` | `NEG` | index-grammar | CREATE INDEX omits explicit index name → reject | R041,R039,R033 |
| `TC-386` | `NEG` | index-grammar | index has zero key columns → reject | R041 |
| `TC-387` | `NEG` | index-grammar | index has more than 32 key columns → reject | R041 |
| `TC-388` | `NEG` | index-grammar | index repeats the same key column → reject | R041 |
| `TC-389` | `POS` | positive-acceptance | exact named non-concurrent one-key index on same-UP new table with medium risk → pass | R041,R039,R029,R034,R008 |
| `TC-390` | `POS` | positive-acceptance | exact named UNIQUE INDEX CONCURRENTLY on pre-existing table with 32 distinct simple keys and medium risk → pass | R041,R039,R029,R034,R032,R012 |
| `TC-391` | `NEG` | identifier-grammar | any future SQL object identifier component is 64 ASCII bytes → reject before SQL execution | R039,R033 |
| `TC-392` | `NEG` | identifier-grammar | two overlong source identifiers share first 63 bytes and would collide after PostgreSQL truncation → both reject before execution | R039,R028 |
| `TC-393` | `NEG` | identifier-grammar | quoted, Unicode, uppercase, dollar-containing or leading-underscore future identifier → reject | R039,R033 |
| `TC-394` | `POS` | positive-acceptance | 63-byte lowercase ASCII identifier matching `[a-z][a-z0-9_]{0,62}` → pass | R039 |
| `TC-395` | `NEG` | mapping-integrity | TC→R semantic edge exists but corresponding R→TC projection edge is absent → meta-gate fail | R037 |
| `TC-396` | `NEG` | mapping-integrity | R→TC semantic edge exists but corresponding TC→R edge is absent → meta-gate fail | R037 |
| `TC-397` | `POS` | positive-acceptance | exact TC→R and generated R→TC edge sets are equal → pass | R037 |
| `TC-398` | `NEG` | semantic-meta-lint | summary says stale bare rule/test/control count inconsistent with exact ID range → meta-gate fail | R038 |
| `TC-399` | `POS` | positive-acceptance | all normative registry summaries publish exact ID range and computed count consistently → pass | R038 |
| `TC-400` | `NEG` | semantic-meta-lint | current normative section contains unresolved placeholder such as `<safe-literal>`, `frozen simple grammar`, or `freeze later` → meta-gate fail | R038,R040,R041 |
| `TC-401` | `POS` | positive-acceptance | semantic-freeze placeholder scan is clean and every referenced grammar/enum/bound resolves in-plan → pass | R038,R040,R041 |
| `TC-402` | `NEG` | semantic-meta-lint | implementation introduces unlisted literal alias, enum value, bound, identifier form or index clause → contract test fail | R040,R041,R039 |
| `TC-403` | `NEG` | policy-source-extraction | source row uses P3D identity or derived row masquerades as S2 identity → reject registry | R038 |
| `TC-404` | `POS` | positive-acceptance | implementation/closure report emits source and derived control sets separately with exact equality → pass | R038 |
| `TC-405` | `NEG` | identifier/index-grammar | index key identifier exceeds 63 bytes → reject | R039,R041 |
| `TC-406` | `NEG` | identifier/index-grammar | explicit index name exceeds 63 bytes → reject | R039,R041 |
| `TC-407` | `NEG` | identifier/index-grammar | schema/table identifier component exceeds 63 bytes → reject | R039,R041 |
| `TC-408` | `POS` | positive-acceptance | numeric(p,s) literal exactly at integer/fraction precision-scale boundary → pass | R040 |
| `TC-409` | `POS` | positive-acceptance | parameterized smallint/integer/bigint exact minimum and maximum literals → pass | R040 |
| `TC-410` | `NEG` | literal-grammar | any negative-zero numeric spelling accepted as canonical → reject regression | R040 |
| `TC-411` | `NEG` | literal-grammar | embedded single quote is not doubled exactly inside ordinary string literal → reject | R040 |

| `TC-412` | `NEG` | policy-source-extraction | S2-137 drops either `metrics` or `logs` from canonical sensitive-content prohibition → source-registry gate fails | R021,R038 |
| `TC-413` | `NEG` | ddl-subgrammar/risk-floor | CREATE TABLE contains column-level `NOT NULL` → reject; paired-SQL v1 low empty-table form is nullable only | R008,R033,R034,R042 |
| `TC-414` | `NEG` | ddl-subgrammar/risk-floor | ADD COLUMN contains `NOT NULL` with or without literal default → reject; future NOT NULL support requires reviewed scope expansion | R008,R033,R034,R042 |
| `TC-415` | `POS` | positive-acceptance | parameterized scanner-derived `statement_class` acceptance covers exactly all 13 frozen values in their valid UP/DOWN forms | R013,R042 |
| `TC-416` | `NEG` | finite-domain/ddl-impact | manifest statement_class differs from scanner-derived exact class → reject | R013,R042 |
| `TC-417` | `NEG` | finite-domain/ddl-impact | statement_class uses unlisted alias/case/broader/narrower value → reject | R013,R042 |
| `TC-418` | `NEG` | finite-domain-meta | required finite-domain manifest field/domain is absent from exact `FD-001…FD-019` §7.10 registry → semantic meta-gate fails | R038,R042 |
| `TC-419` | `POS` | positive-acceptance | parameterized authority reference accepts exactly all five frozen `kind` values in structurally valid contexts | R015,R042 |
| `TC-420` | `NEG` | finite-domain/authority | authority `kind` uses any additional alias/case/future value → reject | R015,R042 |
| `TC-421` | `NEG` | finite-domain-meta | normative finite-domain text delegates exact values to code/tests/Stage 3.55 or uses non-exhaustive `include` wording → semantic meta-gate fails | R038,R042 |
| `TC-422` | `NEG` | finite-domain-meta | §7.10 domain member set differs from cited normative section or decoder domain → semantic meta-gate fails | R038,R042 |
| `TC-423` | `POS` | positive-acceptance | exact `FD-001…FD-019` set and every exact domain-member set equal §7.10, normative authorities and decoder contract → pass | R038,R042 |
| `TC-424` | `NEG` | risk-floor | pre-existing-table nullable ADD COLUMN with literal DEFAULT declares risk=low → reject; frozen minimum is medium | R008,R034,R016,R025 |
| `TC-425` | `NEG` | rollout-reference-binding | manifest contains removed `rollout.plan_ref` field → reject as unknown; staged-rollout authority ref is the sole plan identity | R043 |
| `TC-426` | `NEG` | rollout-reference-binding | staged rollout contains two distinct valid `authority_refs[kind=staged_rollout]` → reject as ambiguous | R043 |
| `TC-427` | `POS` | positive-acceptance | medium/high staged rollout with non-empty metrics, exactly one valid `staged_rollout` authority ref and no `rollout.plan_ref` → pass rollout-reference binding | R043 |
| `TC-428` | `NEG` | policy-source-extraction | `S2-025` source Requirement broadens canonical `large-table` to derived `pre-existing-table` policy → source-purity gate fails | R021,R038 |
| `TC-429` | `NEG` | policy-source-extraction | `S2-027` or `S2-028` source Requirement adds executable `applied` semantics beyond canonical explicit-timeout wording → source-purity gate fails | R021,R038 |
| `TC-430` | `NEG` | policy-source-extraction | atomic canonical Snapshot step `expand new snapshot representation/version` is absent → source-completeness gate fails | R021,R038 |
| `TC-431` | `NEG` | policy-source-extraction | Snapshot rebuild source control drops `market/inflation` qualifier from registered inputs → source-fidelity gate fails | R021,R038 |
| `TC-432` | `NEG` | policy-source-extraction | identity-deletion source controls drop either `encryption-key` or `cryptographically destroyed` qualifier → source-fidelity gate fails | R021,R038 |
| `TC-433` | `NEG` | policy-source-extraction | canonical risk-example Requirement text contains P3-08-derived operation-floor semantics rather than source-only classification text → source-purity gate fails | R021,R038 |
| `TC-434` | `POS` | positive-acceptance | exact current source-fidelity anchors for S2-012/013/025/027/028/057/092/095/096/113/120/121/126/149/150/152/153/165/166/167 plus derived P3D-010…018 separation → pass | R021,R038 |
| `TC-435` | `NEG` | manifest-open-domain | `rollout.metrics[]` contains arbitrary name/object/non-string instead of exact observability category key → reject | R044 |
| `TC-436` | `NEG` | manifest-open-domain | `rollout.metrics[]` contains duplicate key or category whose §7.6 mode is `not_applicable` → reject | R044 |
| `TC-437` | `POS` | positive-acceptance | `rollout.metrics[]` is a non-empty unique set of exact measured §7.6 category keys → pass | R044 |
| `TC-438` | `NEG` | manifest-open-domain | `monitoring.signals[]` contains arbitrary/non-string/duplicate/N-A category → reject | R044 |
| `TC-439` | `POS` | positive-acceptance | `monitoring.signals[]` is a non-empty unique set of exact measured §7.6 category keys → pass | R044 |
| `TC-440` | `NEG` | manifest-open-domain | required open-text field is null, non-string or empty after exact `ASCII_TRIM_BYTES` → reject | R044 |
| `TC-441` | `POS` | positive-acceptance | parameterized required open-text fields accept content non-empty under exact `ASCII_TRIM_BYTES` and preserve decoded content without case/Unicode/content normalization → pass | R044 |
| `TC-442` | `POS` | positive-acceptance | exact `SA-001…SA-082` line-range/hash registry recomputes against frozen Stage 2 and maps every `S2-001…S2-168` exactly once → pass | R045 |
| `TC-443` | `NEG` | source-anchor | delete one required SA row or leave one canonical normative source block unaccounted → reject | R045 |
| `TC-444` | `NEG` | source-anchor | SA line range or fragment SHA-256 differs from exact frozen Stage 2 bytes → reject | R045 |
| `TC-445` | `NEG` | source-anchor | SA maps unknown S2, maps one S2 twice, or leaves an active S2 unmapped → reject | R045 |
| `TC-446` | `NEG` | source-provenance | source-faithful S2 fixture is replaced by P3-08-derived graph/per-direction strengthening under S2 namespace → semantic meta-audit fail | R045 |
| `TC-447` | `POS` | positive-acceptance | authority path `docs/rehearsal.md` matches exact ASCII canonical path grammar and candidate hash → pass | R046 |
| `TC-448` | `NEG` | authority-path | authority path begins `./` → reject; no cleaning/normalization | R046 |
| `TC-449` | `NEG` | authority-path | authority path contains exact `.` or `..` segment → reject | R046 |
| `TC-450` | `NEG` | authority-path | authority path contains repeated `//`, empty segment, leading slash or trailing slash → reject | R046 |
| `TC-451` | `NEG` | authority-path | authority path contains backslash, colon, percent/URL syntax or platform separator variant → reject | R046 |
| `TC-452` | `NEG` | authority-path | authority path contains space, ASCII control byte, non-ASCII code point, segment >255 bytes or total path >1024 bytes → reject | R046 |
| `TC-453` | `NEG` | authority-path | two authority refs repeat exact `(kind,path)` with same or conflicting SHA-256 → reject | R046 |
| `TC-454` | `NEG` | authority-path | implementation accepts invalid path only after `path.Clean`/`filepath.Clean`, case fold, Unicode normalization or dot-segment resolution → regression fail | R046 |
| `TC-455` | `NEG` | manifest-open-domain | parameterized required open-text value consists only of any combination of exact trim bytes 09/0A/0B/0C/0D/20 → reject as empty | R047 |
| `TC-456` | `POS` | positive-acceptance | required open-text value consisting of U+00A0 is non-empty, accepted and preserved because Unicode whitespace is outside `ASCII_TRIM_BYTES` → pass | R047 |
| `TC-457` | `POS` | positive-acceptance | required open-text value with ASCII trim bytes around non-empty content is accepted while the original decoded UTF-8 content remains byte-for-byte unchanged → pass | R047 |
| `TC-458` | `NEG` | manifest-open-domain | implementation substitutes Unicode-aware TrimSpace or rewrites accepted text after emptiness check → regression fail | R047 |
| `TC-459` | `POS` | positive-acceptance | CREATE TABLE exactly matches `CREATE TABLE schema.table (col integer);` under current frozen literal production → pass | R048 |
| `TC-460` | `POS` | positive-acceptance | CREATE TABLE with multiple pairwise-distinct columns and canonical `NULL DEFAULT safe_literal` order matches current frozen literal production → pass | R048 |
| `TC-461` | `NEG` | ddl-token-grammar | CREATE TABLE has empty column list or trailing comma → reject | R048 |
| `TC-462` | `NEG` | ddl-token-grammar | CREATE TABLE repeats a column identifier → reject before PostgreSQL execution | R048 |
| `TC-463` | `NEG` | ddl-token-grammar | column definition uses `DEFAULT safe_literal NULL` or another reversed optional-clause order → reject | R048 |
| `TC-464` | `NEG` | ddl-token-grammar | column definition repeats `NULL`, repeats `DEFAULT`, or includes extra token after allowed branch → reject | R048 |
| `TC-465` | `POS` | positive-acceptance | `ALTER TABLE schema.table ADD COLUMN col integer;` matches exact ADD COLUMN production → pass | R048 |
| `TC-466` | `POS` | positive-acceptance | ADD COLUMN with canonical optional `NULL DEFAULT safe_literal` order matches exact production → pass | R048 |
| `TC-467` | `NEG` | ddl-token-grammar | ADD COLUMN omits mandatory `COLUMN`, adds multiple comma-separated columns, or reverses optional clause order → reject | R048 |
| `TC-468` | `NEG` | ddl-token-grammar | ADD COLUMN contains trailing/unrecognized clause/token outside exact production → reject | R048 |
| `TC-469` | `NEG` | evidence-binding | `S2-021` carries any machine R edge or loses reviewer-only Compatibility ownership → semantic evidence gate fails | R049 |
| `TC-470` | `NEG` | evidence-binding | `S2-093` cites R018 or any machine rule as proof of approved ongoing Populate responsibility → reject binding as unrelated evidence | R049 |
| `TC-471` | `NEG` | evidence-binding | any S2 row outside complete-machine/scope-rejected sets carries machine rules but lacks an exact partial machine-subset + non-machine-remainder registry entry → meta-gate fail | R049 |
| `TC-472` | `POS` | positive-acceptance | exact current frozen four-way S2 evidence taxonomy plus 36-row partial-evidence binding registry is complete; the exact conservative prerequisite-only reviewer/external-evidence set is machine-edge-free, paired-SQL scope rejection and no-machine approval binding are accepted → pass | R049 |
| `TC-473` | `NEG` | semantic-meta-mutation | remove semantic owner R042 from derived `SEMANTIC_FREEZE_RULE_SET` or relabel it meta/proof-only → aggregate P3D-008 proof fails | R050 |
| `TC-474` | `NEG` | semantic-meta-mutation | remove semantic owner R046 or R047 from complement-derived aggregate set → fail | R050 |
| `TC-475` | `NEG` | semantic-meta-mutation | remove semantic owner R048, R041 or R033 from aggregate set → fail | R050 |
| `TC-476` | `NEG` | semantic-meta-mutation | remove semantic owner R040 or R051 from aggregate set → fail | R050 |
| `TC-477` | `NEG` | semantic-meta-mutation | remove semantic owner R052 or R053 from aggregate set → fail | R050 |
| `TC-478` | `NEG` | semantic-meta-mutation | remove semantic owner R036/R039/R054, or add any new R-rule without exact semantic-vs-meta classification → fail | R050 |
| `TC-479` | `POS` | positive-acceptance | exact aggregate semantic-owner/meta partition covers R001…R075 exactly once and derived `SEMANTIC_FREEZE_RULE_SET` equals the 50-rule non-meta complement → pass | R050 |
| `TC-480` | `POS` | positive-acceptance | canonical type parameters `numeric(1,0)`, `numeric(38,38)`, `varchar(1)` and `varchar(10485760)` match exact decimal-token grammar and bounds → pass | R051 |
| `TC-481` | `NEG` | type-parameter-grammar | `numeric(01,0)`, `numeric(+1,0)`, `numeric(1,00)`, `varchar(0005)` or `varchar(+5)` → reject before PostgreSQL | R051 |
| `TC-482` | `NEG` | type-parameter-grammar | precision/scale/varchar parameter contains sign, underscore, decimal point, exponent, empty token or other numeric decoration outside exact production → reject | R051 |
| `TC-483` | `NEG` | type-parameter-grammar | lexically canonical parameter violates p/s/n value bounds (`p=0`, `p=39`, `s>p`, `n=0`, `n=10485761`) → reject | R051 |
| `TC-484` | `POS` | positive-acceptance | exact CHECK envelope `ALTER TABLE s.t ADD CONSTRAINT c CHECK (col = 1) NOT VALID;` with valid same-UP typed column → pass | R052 |
| `TC-485` | `NEG` | check-envelope-grammar | CHECK uses extra parenthesis layer or alternate outer parenthesization not present in exact production → reject | R052 |
| `TC-486` | `NEG` | check-envelope-grammar | CHECK omits, duplicates or reorders `NOT VALID` → reject | R052 |
| `TC-487` | `NEG` | check-envelope-grammar | CHECK adds `NO INHERIT`, `ONLY`, trailing clause/token or any unsupported envelope option → reject | R052 |
| `TC-488` | `NEG` | check-envelope-grammar | CHECK outer token order differs from exact `ALTER TABLE qualified_table ADD CONSTRAINT identifier CHECK (...) NOT VALID ;` production → reject | R052 |
| `TC-489` | `POS` | positive-acceptance | exact `DROP TABLE schema.table;` inverse of same-UP CREATE TABLE → pass | R053 |
| `TC-490` | `POS` | positive-acceptance | exact `ALTER TABLE schema.table DROP COLUMN col;` inverse of same-UP ADD COLUMN → pass | R053 |
| `TC-491` | `POS` | positive-acceptance | exact `ALTER TABLE schema.table DROP CONSTRAINT c;` inverse of same-UP ADD CONSTRAINT → pass | R053 |
| `TC-492` | `POS` | positive-acceptance | exact `DROP INDEX schema.idx;` inverse of same-UP non-concurrent CREATE INDEX → pass | R053 |
| `TC-493` | `POS` | positive-acceptance | exact `DROP INDEX CONCURRENTLY schema.idx;` inverse of same-UP concurrent CREATE INDEX → pass | R053 |
| `TC-494` | `NEG` | down-token-grammar | DOWN uses multi-target DROP TABLE/INDEX or comma-separated target list → reject | R053 |
| `TC-495` | `NEG` | down-token-grammar | DOWN adds `IF EXISTS` → reject | R053 |
| `TC-496` | `NEG` | down-token-grammar | DOWN adds `CASCADE`, explicit `RESTRICT`, `ONLY` or any unsupported option → reject | R053 |
| `TC-497` | `NEG` | down-token-grammar | DOWN contains trailing/unrecognized token, wrong token order or alternate spelling outside exact class production → reject | R053 |
| `TC-498` | `NEG` | ci-discovery-domain | migration directory contains `evil.up.sql` or `notes.down.sql`; exact `.sql` discovery sees it and canonical filename validation rejects instead of ignoring → fail | R054 |
| `TC-499` | `NEG` | ci-discovery-domain | migration directory contains `000008-bad-name.up.sql` or `123_bad.up.sql`; `.sql` discovery includes it but canonical filename grammar rejects → fail | R054 |
| `TC-500` | `NEG` | ci-discovery-domain | proposed validator discovery predicate is narrower than every direct migration-directory `.sql` basename or silently ignores malformed `.up.sql/.down.sql` → meta-gate fail | R054 |
| `TC-501` | `POS` | positive-acceptance | every canonical paired `.up.sql/.down.sql` member is discovered, validated and current frozen CI `*.up.sql` execution subjects are a subset of validator-approved files → pass | R054 |
| `TC-502` | `NEG` | ci-discovery-domain | frozen CI selector/inventory contains an executable `*.up.sql` subject that can exist outside validator discovery/approval or execute before exact-SHA validation → fail | R054 |
| `TC-503` | `NEG` | source-authority-binding | canonical source changes/removes only `mandatory` in line 17 while lifecycle sequence line 20 stays byte-identical → SA/source-authority proof for S2-001 must fail | R045 |
| `TC-504` | `NEG` | evidence-binding | S2-109 is classified complete-machine from clean governed SQL/disposable DOWN while external production rollback tooling can still delete financial facts → fail evidence-scope gate | R049 |
| `TC-505` | `NEG` | evidence-binding | S2-118 is classified complete-machine from clean governed SQL while separate snapshot rebuild/runtime tooling can mutate historical transaction rows → fail evidence-scope gate | R049 |
| `TC-506` | `NEG` | semantic-meta-mutation | R033 is removed from semantic-owner set, relabeled meta/proof-only, or any new R-rule is left unclassified → aggregate semantic-freeze proof fails | R050 |
| `TC-507` | `NEG` | derived-evidence-binding | P3D-003 is broadened to claim all SQL type bounds while citing only R040, or a frozen P3D disposition/rule mapping is structurally rebroadended beyond its declared scope registry → fail | R055 |
| `TC-508` | `POS` | positive-acceptance | exact P3D evidence-mode registry is structurally scope-closed; P3D-003 owns only R040 scalar-data/CHECK-predicate semantics, P3D-023 owns R051 type-parameter semantics, aggregate claims bind R050, and reviewer-owned semantic remainders remain explicit → pass | R055 |
| `TC-509` | `NEG` | fk-cardinality | ADD FOREIGN KEY has two local columns but one referenced column → reject before PostgreSQL execution | R033 |
| `TC-510` | `NEG` | fk-cardinality | ADD FOREIGN KEY has one local column but two referenced columns → reject before PostgreSQL execution | R033 |
| `TC-511` | `POS` | positive-acceptance | ADD FOREIGN KEY with two distinct local columns and two distinct referenced columns satisfies exact non-zero equal-cardinality rule → pass | R033 |
| `TC-512` | `NEG` | fk-list-grammar | ADD FOREIGN KEY repeats a local-column identifier inside the local list → reject before PostgreSQL | R033 |
| `TC-513` | `NEG` | fk-list-grammar | ADD FOREIGN KEY repeats a referenced-column identifier inside the referenced list → reject before PostgreSQL | R033 |
| `TC-514` | `NEG` | fk-list-grammar | ADD FOREIGN KEY uses 33 local and 33 referenced columns even though counts match → reject project-policy cap before PostgreSQL | R033 |
| `TC-515` | `POS` | positive-acceptance | ADD FOREIGN KEY uses 32 pairwise-distinct local and 32 pairwise-distinct referenced columns with equal cardinality → pass scanner list grammar | R033 |
| `TC-516` | `NEG` | fk-list-grammar | ADD FOREIGN KEY contains an empty local or referenced column list → reject before PostgreSQL | R033 |
| `TC-517` | `NEG` | evidence-binding | a `REJECTED_BY_PAIRED_SQL_V1` S2 row is counted as complete machine proof of its external lifecycle/operational requirement merely because paired SQL v1 rejects that operation → fail evidence-scope gate | R049 |
| `TC-518` | `POS` | positive-acceptance | exact four-way S2 evidence taxonomy is 7 complete-machine + 46 paired-SQL-scope-rejected + 36 partial + 79 no-machine = 168, with scope rejection never represented as external-requirement proof → pass | R049 |
| `TC-519` | `NEG` | evidence-binding | any of P3D-002/P3D-016/P3D-018/P3D-021/P3D-027 is promoted to complete MACHINE while its source-classification/semantic-subset reviewer remainder still exists → fail P3D evidence-scope gate | R055 |
| `TC-520` | `POS` | positive-acceptance | exact P3D evidence taxonomy is 22 complete-machine + 5 structure-plus-human rows with explicit machine subset and reviewer-owned semantic remainder → pass | R055 |
| `TC-521` | `POS` | positive-acceptance | CREATE TABLE contains exactly 64 pairwise-distinct otherwise-valid column definitions → pass exact project max boundary | R048,R057 |
| `TC-522` | `NEG` | ddl-boundary | CREATE TABLE contains 65 pairwise-distinct otherwise-valid column definitions → reject before PostgreSQL execution | R048,R057 |
| `TC-523` | `NEG` | lexical-overlap | identifier position uses an exact PostgreSQL 18.6 reserved member, including `table`, `select` or `user` → reject before PostgreSQL | R058 |
| `TC-524` | `POS` | positive-acceptance | identifier position uses PostgreSQL 18.6 non-reserved keyword `update` and project identifier grammar matches → contextual identifier passes | R058 |
| `TC-525` | `POS` | positive-acceptance | ordinary non-keyword identifier `portfolio` matches project identifier grammar → pass | R058 |
| `TC-526` | `NEG` | lexical-overlap | reserved member/count/hash/upstream tag-or-blob identity drifts, runtime keyword discovery is substituted, or precedence differs from §15.3a → fail lexical-freeze gate | R058,R060 |
| `TC-527` | `NEG` | evidence-binding | any globally quantified all/every/never/no-migration S2 requirement is `MACHINE_COMPLETE` while a separately governed migration mechanism lies outside its observer universe → fail | R049,R056 |
| `TC-528` | `POS` | positive-acceptance | exact S2 evidence partition is 7 complete + 46 paired-SQL-scope-rejected + 36 partial + 79 none and every global migration/versioning sibling is explicitly classified: observer-bearing rows have a machine subset/remainder and no-machine rows name reviewer/external ownership → pass | R049,R056 |
| `TC-529` | `NEG` | boundary-proof | a finite normative min/max/cardinality/length/range exists without required exact boundary witness and adjacent-invalid rejection → fail global bound inventory | R057 |
| `TC-530` | `POS` | positive-acceptance | every finite normative bound in the semantic-atom inventory has boundary witnesses or explicit mathematical non-applicability rationale → pass | R057 |
| `TC-531` | `NEG` | lexical-overlap | any non-empty lexical token-class intersection lacks exact precedence/context/exclusion owner and deterministic witness pair → fail | R058,R059 |
| `TC-532` | `POS` | positive-acceptance | all registered lexical intersections, including keyword↔identifier, have one deterministic owner and witness pair → pass | R058,R059 |
| `TC-533` | `NEG` | semantic-meta-mutation | a normative semantic atom is ownerless/multiply authoritative, its material mutation survives all gates, or a remediation scans only the named fixture and leaves a sibling escape → fail | R059,R060,R061 |
| `TC-534` | `POS` | positive-acceptance | GUARD-01…06 generic predicates, complete sibling scans, semantic-atom ownership and mutation-kill requirements are all present and exact → pass | R056,R057,R058,R059,R060,R061 |
| `TC-535` | `POS` | positive-acceptance | canonical authority path contains one segment of exactly 255 allowed ASCII bytes and total path remains <=1024 → pass segment max boundary | R046,R057 |
| `TC-536` | `NEG` | authority-path-boundary | authority path contains one otherwise-valid segment of 256 ASCII bytes → reject before path/hash resolution | R046,R057 |
| `TC-537` | `POS` | positive-acceptance | canonical multi-segment authority path totals exactly 1024 allowed ASCII bytes with every segment <=255 → pass total max boundary | R046,R057 |
| `TC-538` | `NEG` | authority-path-boundary | canonical-looking authority path totals 1025 ASCII bytes → reject before file resolution | R046,R057 |
| `TC-539` | `POS` | positive-acceptance | SHA-256 field is exactly 64 lowercase hexadecimal characters and matches exact candidate bytes → pass lexical-length boundary | R005,R057 |
| `TC-540` | `NEG` | hash-boundary | SHA-256 field is 63 or 65 characters, uppercase, or contains non-hex character → reject before byte-identity comparison | R005,R057 |
| `TC-541` | `POS` | positive-acceptance | expected_duration_seconds accepts exact lower 1 and upper 86_400 for both directions → pass frozen boundary | R012,R036,R057 |
| `TC-542` | `POS` | positive-acceptance | lock_timeout_ms accepts exact lower 1 and upper 86_400_000 with compatible statement timeout → pass frozen boundary | R012,R036,R057 |
| `TC-543` | `POS` | positive-acceptance | statement_timeout_ms accepts exact lower 1 and upper 86_400_000 when cross-field relation is satisfied → pass frozen boundary | R012,R036,R057 |
| `TC-544` | `POS` | positive-acceptance | statement_timeout_ms equals lock_timeout_ms at exact lower and upper boundaries → pass `statement>=lock` equality edge | R012,R057 |
| `TC-545` | `POS` | positive-acceptance | future migration ID `999999` satisfies exact six-digit nonzero lexical/ordering domain → pass upper lexical boundary | R002,R057 |
| `TC-546` | `POS` | positive-acceptance | `affected_rows_estimate` accepts exact 0 and 9_223_372_036_854_775_807 → pass independent lower/upper field boundaries | R036,R057 |
| `TC-547` | `NEG` | manifest-boundary | `affected_rows_estimate` is -1 or 9_223_372_036_854_775_808 → reject before SQL execution | R036,R057 |
| `TC-548` | `POS` | positive-acceptance | `disk_impact_bytes_estimate` accepts exact 0 and 9_223_372_036_854_775_807 → pass independent lower/upper field boundaries | R036,R057 |
| `TC-549` | `NEG` | manifest-boundary | `disk_impact_bytes_estimate` is -1 or 9_223_372_036_854_775_808 → reject before SQL execution | R036,R057 |
| `TC-550` | `POS` | positive-acceptance | `wal_impact_bytes_estimate` accepts exact 0 and 9_223_372_036_854_775_807 → pass independent lower/upper field boundaries | R036,R057 |
| `TC-551` | `NEG` | manifest-boundary | `wal_impact_bytes_estimate` is -1 or 9_223_372_036_854_775_808 → reject before SQL execution | R036,R057 |
| `TC-552` | `NEG` | discovered-boundary-universe | synthetic formal-manifest integer field with a finite bound is added while no BND atom represents its semantic key → discovered-vs-registered equality fails even if old BND count/set remains unchanged | R057,R060,R063 |
| `TC-553` | `POS` | positive-acceptance | all six bounded integer manifest semantic fields are discovered and each resolves to registered BND ownership, including independent BND-17/18/19 estimate atoms → pass | R057,R063 |
| `TC-554` | `NEG` | lexical-overlap | governed ColId uses PostgreSQL REL_18_6 `collation` (`TYPE_FUNC_NAME_KEYWORD`) → reject before PostgreSQL | R058,R062 |
| `TC-555` | `NEG` | lexical-overlap | governed ColId uses PostgreSQL REL_18_6 `concurrently` (`TYPE_FUNC_NAME_KEYWORD`) → reject before PostgreSQL | R058,R062 |
| `TC-556` | `NEG` | lexical-overlap | governed ColId uses PostgreSQL REL_18_6 `cross` (`TYPE_FUNC_NAME_KEYWORD`) → reject before PostgreSQL | R058,R062 |
| `TC-557` | `NEG` | authority-projection-mutation | exact reserved 78-member set/hash remains unchanged but `TYPE_FUNC_NAME_KEYWORD` is omitted from ColId-disallowed projection, or gram.y ColId production/blob is not bound → fail authority-projection proof | R058,R060,R062 |
| `TC-558` | `POS` | positive-acceptance | exact REL_18_6 ColId projection is `IDENT|unreserved|col_name`; 23 type-func + 78 reserved = 101 disallowed unique words, `update` remains allowed and ordinary identifier remains allowed → pass | R058,R062 |
| `TC-559` | `NEG` | observer-universe-closure | `S2-011` or any other canonical partial-machine row retains machine evidence but is removed from the partial evidence-scope registry → disposition-derived observer closure fails | R056,R061,R063 |
| `TC-560` | `NEG` | observer-universe-closure | synthetic canonical S2 row with `STRUCTURE_PLUS_HUMAN_ADEQUACY` plus an R-rule is introduced without partial evidence-scope registration → fail regardless of wording and without any hand-authored sibling expected set | R056,R060,R063 |
| `TC-561` | `POS` | positive-acceptance | disposition-derived observer closure is exact: 7 complete + 46 scope-rejected + 36 partial = 89 observer-bearing, 79 no-machine, total 168; every partial row has evidence-scope registration and no no-machine row carries R evidence | R056,R063 |
| `TC-562` | `NEG` | single-source-bound | mutate only `affected_rows_estimate` authoritative upper bound while BND/test IDs remain unchanged → exact BOUND-SPEC/value/accountability proof fails | R057,R060,R065 |
| `TC-563` | `NEG` | generic-integer-discovery | add a new formal-manifest integer field while keeping formal row count stable and adding no FD/BOUND-SPEC → generic integer-field closure fails without field-name allowlist | R057,R060,R065 |
| `TC-564` | `POS` | positive-acceptance | every formal-manifest integer field is generically discovered and resolves exactly once to FD-001 singleton or one BOUND-SPEC with exact lower/upper and BND owner → pass | R057,R065 |
| `TC-565` | `NEG` | colid-single-source | mutate only anchored §15.3a COLID_POLICY_SPEC production/category projection while later summaries/evidence remain old → policy digest/evidence equality fails | R058,R060,R066 |
| `TC-566` | `POS` | positive-acceptance | exactly one anchored COLID_POLICY_SPEC exists; its digest, upstream blobs, category sets, 101-union and packaged evidence all agree → pass | R058,R062,R066 |
| `TC-567` | `NEG` | unregistered-semantic-candidate | insert a new normative SQL sentence `every index name MUST end with _idx` without R/TC/ATOM/accountability update → frozen normative-line evidence fails | R059,R060,R064,R065 |
| `TC-568` | `POS` | positive-acceptance | deterministic semantic-candidate extraction equals frozen NORMATIVE_LINE_ACCOUNTABILITY evidence and all 50 semantic R owners have exactly one ATOM → pass | R059,R064,R065 |
| `TC-569` | `NEG` | semantic-atom-integrity | change an ATOM owner or body digest while R/TC counts remain unchanged → exact ATOM↔R bijection/digest proof fails | R059,R060,R064 |
| `TC-570` | `POS` | positive-acceptance | ATOM-001…050 exactly equals the 50-rule semantic-owner complement and every atom digest equals its authoritative R semantic body → pass | R050,R059,R064 |
| `TC-571` | `NEG` | evidence-binding | reintroduce `S2-101 → R018` as machine evidence for independent deployability → explicit unrelated-edge regression fails even though R018 exists and dependency tests pass | R049,R060,R067 |
| `TC-572` | `POS` | positive-acceptance | S2-101 is reviewer/Architecture-owned with no machine R edge; R018 remains attached only to dependency-structure controls where relevant → pass | R049,R067 |
| `TC-573` | `NEG` | atom-universe-growth | add a new semantic R-rule without a corresponding ATOM or add an ATOM without exactly one semantic R owner → atom-universe equality fails | R050,R060,R064 |
| `TC-574` | `NEG` | normative-accountability-drift | change/delete one frozen semantic-candidate plan line while all IDs/counts remain intact → packaged NORMATIVE_LINE_ACCOUNTABILITY mismatch fails | R059,R064,R065 |
| `TC-575` | `POS` | positive-acceptance | historical v15 packaged mutation audit killed its eight registered bypass mutations; the current contract does not treat that closed set as exhaustive → pass | R060,R061,R063,R064,R065,R066,R067 |
| `TC-576` | `NEG` | global-bound-authority | duplicate identical `BOUND-SPEC|...` outside the anchored registry while NLA is regenerated → global physical-authority scan fails | R057,R068 |
| `TC-577` | `NEG` | global-bound-authority | conflicting `BOUND-SPEC|field=affected_rows_estimate|lower=1|...` outside the anchored registry while NLA is regenerated → fail | R057,R068 |
| `TC-578` | `NEG` | global-bound-authority | duplicate bounded field under a different BND ID anywhere in candidate → global field/authority uniqueness fails | R057,R068 |
| `TC-579` | `NEG` | global-bound-authority | second complete machine-readable BOUND_SPEC block is introduced → global occurrence/location scan fails | R057,R068 |
| `TC-580` | `POS` | positive-acceptance | globally exactly six physical `BOUND-SPEC|...` rows exist, all and only inside anchored BOUND_SPEC_REGISTRY, with unique fields and exact BND/value binding → pass | R057,R068 |
| `TC-581` | `NEG` | global-colid-authority | second identical `COLID_POLICY|...` line outside §15.3a while NLA is regenerated → global policy occurrence scan fails | R058,R066,R068 |
| `TC-582` | `NEG` | global-colid-authority | second conflicting `COLID_POLICY|...` line outside §15.3a → fail | R058,R066,R068 |
| `TC-583` | `NEG` | global-colid-authority | second ColId policy with same declared digest but different physical location → fail before semantic projection comparison | R058,R066,R068 |
| `TC-584` | `POS` | positive-acceptance | exactly one physical `COLID_POLICY|...` exists globally and it is the anchored §15.3a authority bound to PostgreSQL REL_18_6 evidence → pass | R058,R066,R068 |
| `TC-585` | `NEG` | atom-occurrence-bijection | duplicate exact `ATOM-001→R001` physical declaration while NLA is regenerated → occurrence multiplicity fails | R064,R069 |
| `TC-586` | `NEG` | atom-occurrence-bijection | duplicate ATOM ID with a different owner → ATOM-ID multiplicity fails | R064,R069 |
| `TC-587` | `NEG` | atom-occurrence-bijection | duplicate semantic owner under a different ATOM ID → owner multiplicity fails | R064,R069 |
| `TC-588` | `NEG` | atom-occurrence-bijection | remove one atom and duplicate another so physical count remains 50 → exact ID/owner multiplicity fails | R064,R069 |
| `TC-589` | `NEG` | atom-occurrence-bijection | add an extra duplicate physical ATOM row while unique ID set remains 50 → physical declaration count fails | R064,R069 |
| `TC-590` | `POS` | positive-acceptance | exactly 50 physical ATOM declarations exist; IDs 001…050 and semantic owners are each exact-once and every occurrence digest matches its R body → pass | R050,R064,R069 |
| `TC-591` | `NEG` | partition-summary-consistency | any active evidence-partition numeric summary differs from the canonical-row-derived `7/46/36/79` authority while registries stay unchanged → fail | R049,R070 |
| `TC-592` | `POS` | positive-acceptance | canonical S2 rows derive exactly `7 complete + 46 scope-rejected + 36 partial + 79 none`; the sole partition authority and every active numeric summary agree → pass | R049,R070 |
| `TC-593` | `NEG` | evidence-binding | reattach R008 to reviewer-only S2-019 old-version ignore behavior → fail as prerequisite-not-logical-subset | R049,R070 |
| `TC-594` | `NEG` | evidence-binding | reattach R008/R029 to reviewer-only S2-021 old-version ignore behavior → fail | R049,R070 |
| `TC-595` | `NEG` | evidence-binding | reattach R008 to reviewer-only S2-022 reader-tolerance timing behavior → fail | R049,R070 |
| `TC-596` | `NEG` | evidence-binding | reattach R021 to reviewer-only S2-141 tool-nonweakening behavior → fail | R049,R070 |
| `TC-597` | `POS` | positive-acceptance | exact 36-row remaining partial registry is structurally complete under the current frozen strict semantic pass; every confirmed/proactively identified prerequisite-only row is reviewer/external-evidence owned and no such edge remains → pass | R049,R070 |
| `TC-598` | `POS` | positive-acceptance | parameterized four-field direction-independence: for each field independently, an otherwise valid migration holds the other execution metadata at valid compatible values and accepts `expected_duration_seconds: UP=10 DOWN=20`, `lock_risk: UP=low DOWN=high`, `lock_timeout_ms: UP=1000 DOWN=3000`, and `statement_timeout_ms: UP=2000 DOWN=4000`; every one-field UP≠DOWN witness → pass | R012,R071 |
| `TC-599` | `NEG` | positive-universe-regression | for each field in `{expected_duration_seconds,lock_risk,lock_timeout_ms,statement_timeout_ms}`, introduce a normative or implementation requirement `UP[field] == DOWN[field]` → the corresponding parameterized TC-598 acceptance witness fails; all four equality-coupling mutants must be killed independently | R012,R071 |
| `TC-600` | `POS` | positive-acceptance | packaged current mutation and extra red-team suites kill all mandatory v15 reviewer survivors plus adjacent global-duplicate, occurrence, partition, evidence-edge and UP/DOWN-coupling attacks → pass | R060,R061,R068,R069,R070,R071 |
| `TC-601` | `NEG` | evidence-binding | reattach R007 to reviewer-only S2-001 mandatory lifecycle sequencing → fail because lifecycle vocabulary is prerequisite, not a logical subset of mandatory ordering | R070 |
| `TC-602` | `NEG` | evidence-binding | reattach R014 to reviewer-only S2-010 production observability → fail because declaration profile is prerequisite, not runtime observability evidence | R070 |
| `TC-603` | `NEG` | evidence-binding | reattach R008/R013 to reviewer-only S2-017 backward compatibility → fail because additive grammar/impact metadata do not prove old/new app compatibility | R070 |
| `TC-604` | `NEG` | evidence-binding | reattach R026 to reviewer-only S2-106 rollback preference → fail because strategy declaration does not prove operational preference/execution | R070 |
| `TC-605` | `NEG` | evidence-binding | reattach R014 to any reviewer-only S2-128…S2-136 production-reporting row → fail because observability schema declaration does not prove runtime report emission | R070 |
| `TC-606` | `NEG` | evidence-binding | reattach R034/R016/R017 to reviewer-only S2-149/S2-150/S2-155/S2-156 exact risk-classification examples/anchors → fail because floors/gates do not prove exact semantic classification | R070 |
| `TC-607` | `NEG` | evidence-binding | reattach R021 to reviewer-only S2-162 priority policy → fail because registry presence does not prove operational conflict resolution by that priority | R070 |
| `TC-608` | `POS` | positive-acceptance | exact conservative reviewer-only non-subset set is machine-edge-free; the remaining 36 partial bindings are the only rows for which a direct observer-scoped logical subset is claimed after the full semantic pass → pass | R070 |



| `TC-609` | `NEG` | cardinality-property | `owners[]` is empty → reject; PROP-002 min-one cardinality must remain exact | R019,R073,R074 |
| `TC-610` | `NEG` | cardinality-property | `monitoring.signals[]` is empty → reject; PROP-003 min-one cardinality must remain exact | R044,R073,R074 |
| `TC-611` | `NEG` | cardinality-property | `rollout.metrics[]` is empty → reject; PROP-004 min-one cardinality must remain exact | R044,R073,R074 |
| `TC-612` | `NEG` | property-level-grammar | CHECK contains two atomic predicates joined by AND/OR → reject; PROP-008 exact-one predicate must remain exact | R040,R033,R074 |
| `TC-613` | `NEG` | property-level-inverse | one UP effect maps to two DOWN inverses or one DOWN inverse maps to multiple UP effects → reject exact 1↔1 multiplicity | R028,R009,R073,R074 |
| `TC-614` | `NEG` | property-level-owner | touched-schema set is only a subset/superset of declared owners instead of exact equality → reject | R019,R074 |
| `TC-615` | `NEG` | property-level-index | pre-existing-table index omits CONCURRENTLY while all other syntax is valid → reject; required=true cannot become optional | R029,R041,R074 |
| `TC-616` | `NEG` | cardinality-property | staged rollout has zero or two `staged_rollout` authority refs → reject exact-one | R043,R073,R074 |
| `TC-617` | `POS` | positive-acceptance | standard rollout has zero `staged_rollout` authority refs → pass exact-zero branch | R043,R074 |
| `TC-618` | `NEG` | cardinality-property | standard rollout has one or more `staged_rollout` authority refs → reject exact-zero branch | R043,R073,R074 |
| `TC-619` | `NEG` | semantic-property-manifest | mutate a child normative property, regenerate NLA, leave SEMANTIC_PROPERTY_MANIFEST frozen → property manifest mismatch rejects | R072,R074 |
| `TC-620` | `NEG` | semantic-property-manifest | regenerate semantic-property evidence after child mutation but leave frozen plan manifest digest/count unchanged → digest/count binding rejects | R072,R074 |
| `TC-621` | `NEG` | property-spec | change NULL/DEFAULT max1, CHECK exact1, owner equality, concurrency required, or inverse exact1 while preserving coarse R body → PROP-SPEC/value binding rejects | R074 |
| `TC-622` | `NEG` | cardinality-discovery | add/change an active structural cardinality line with regenerated NLA but without matching cardinality manifest/property authority → fail | R073,R074 |
| `TC-623` | `NEG` | evidence-binding | reattach any machine rule to reviewer-only `S2-107` → semantic evidence gate fails | R049,R075 |
| `TC-624` | `POS` | positive-acceptance | exact current evidence partition is 7 complete + 46 scope-rejected + 36 partial + 79 no-machine; S2-107 has no machine edge | R049,R075 |
| `TC-625` | `POS` | positive-acceptance | SEM/CARD manifests exactly byte-account every active normative line, remain explicitly non-semantic, and the single TA-01 taxonomy assigns machine acceptance/rejection only to exact TC↔MPROP properties while BND remains quantitative-only | R072,R073,R074 |
| `TC-626` | `NEG` | semantic-property-manifest | add a new normative property and regenerate NLA + SEM/CARD manifests + plan manifest bindings while checker frozen surface digest remains unchanged → reject | R072,R074 |
| `TC-627` | `POS` | authority-identity | two independently valid `authority_refs[]` entries use the same exact `path` but different valid `kind` values → pass; uniqueness is exact `(kind,path)`, never path-only | R015 |
| `TC-628` | `POS` | sql-lexical-case | an otherwise valid paired-SQL v1 statement uses lowercase or mixed-case SQL keyword terminals while project identifiers remain valid lowercase identifiers → pass because keyword matching is ASCII-case-insensitive | R027 |
| `TC-629` | `POS` | sql-line-comment | an otherwise valid paired-SQL v1 statement contains a terminated line comment only between grammar tokens, with the following token beginning after the required newline → pass; a legal line comment is a lexical separator, not an automatic rejection | R027 |
| `TC-630` | `POS` | sql-block-comment | an otherwise valid paired-SQL v1 statement contains a terminated nested block comment only between grammar tokens → pass; a legal block comment is a lexical separator, not an automatic rejection | R027 |
| `TC-631` | `POS` | sql-inert-comment | an otherwise valid paired-SQL v1 statement contains a forbidden executable keyword such as `DROP` solely inside a terminated inert comment → the comment content remains inert and does not create a statement-class rejection | R027 |

### 24.2 Machine-rule registry and exact generated test projection

TC rows are the **canonical semantic edge declarations**. The R-table below is a required exact
projection of those TC→R references, split by polarity. Stage 3.54 package meta-audit regenerates the
edge sets and requires equality in both directions; hand-maintained divergence is a planning failure.

| Rule | Requirement | Negative tests (exact projection) | Positive tests (exact projection) |
| --- | --- | --- | --- |
| `R001` | migration-directory and filesystem discovery integrity | `TC-001,TC-002,TC-003,TC-004,TC-005,TC-006,TC-007,TC-008,TC-009,TC-010` | — |
| `R002` | filename/ID grammar, pairing, monotonic identity | `TC-011,TC-012,TC-013,TC-014,TC-015,TC-016,TC-017,TC-018,TC-019,TC-020` | `TC-262,TC-545` |
| `R003` | strict JSON semantics | `TC-021,TC-022,TC-023,TC-024,TC-025,TC-026,TC-027,TC-028,TC-029,TC-030,TC-031` | — |
| `R004` | frozen non-retroactive legacy baseline | `TC-032,TC-033,TC-034,TC-035,TC-036,TC-037,TC-038,TC-039,TC-040,TC-211,TC-212,TC-213,TC-214,TC-215,TC-216,TC-217,TC-218,TC-219,TC-220,TC-221,TC-222` | `TC-287` |
| `R005` | manifest↔file bijection and exact candidate hashes | `TC-041,TC-042,TC-043,TC-044,TC-045,TC-046,TC-540` | `TC-287,TC-289,TC-539` |
| `R006` | PR-base Git immutability and base-context validity | `TC-211,TC-212,TC-213,TC-214,TC-215,TC-216,TC-217,TC-218,TC-219,TC-220,TC-221,TC-222` | `TC-288` |
| `R007` | paired-SQL v1 lifecycle phase support/rejection | `TC-052,TC-053,TC-054,TC-055,TC-056,TC-057,TC-058,TC-059,TC-060,TC-061,TC-062,TC-063,TC-064` | `TC-291` |
| `R008` | future UP allowlist and safe subform grammar | `TC-114,TC-115,TC-116,TC-117,TC-118,TC-119,TC-120,TC-121,TC-122,TC-123,TC-124,TC-125,TC-126,TC-127,TC-128,TC-129,TC-130,TC-131,TC-132,TC-133,TC-134,TC-135,TC-136,TC-137,TC-138,TC-139,TC-140,TC-141,TC-142,TC-143,TC-307,TC-308,TC-310,TC-311,TC-312,TC-313,TC-314,TC-315,TC-316,TC-317,TC-318,TC-319,TC-334,TC-335,TC-336,TC-343,TC-344,TC-375,TC-376,TC-377,TC-378,TC-379,TC-380,TC-381,TC-382,TC-383,TC-384,TC-413,TC-414,TC-424` | `TC-263,TC-264,TC-265,TC-266,TC-267,TC-268,TC-269,TC-270,TC-271,TC-292,TC-293,TC-294,TC-295,TC-323,TC-324,TC-328,TC-329,TC-330,TC-373,TC-374,TC-389` |
| `R009` | direction-specific DOWN inverse contract | `TC-052,TC-053,TC-054,TC-055,TC-056,TC-057,TC-058,TC-059,TC-060,TC-061,TC-062,TC-063,TC-064,TC-196,TC-197,TC-198,TC-199,TC-200,TC-201,TC-202,TC-203,TC-204,TC-205,TC-206,TC-207,TC-208,TC-209,TC-210,TC-613` | `TC-272,TC-273,TC-274,TC-275,TC-276,TC-277,TC-302` |
| `R010` | no DML/data-history/procedural/client mutation surface | `TC-091,TC-092,TC-093,TC-094,TC-095,TC-096,TC-097,TC-098,TC-099,TC-100,TC-101,TC-102,TC-103,TC-104,TC-105,TC-106,TC-107,TC-108,TC-109,TC-110,TC-111,TC-112,TC-113,TC-114,TC-115,TC-116,TC-117,TC-118,TC-119,TC-120,TC-121,TC-122,TC-123,TC-124,TC-125,TC-126,TC-127,TC-128,TC-129,TC-130,TC-131,TC-132,TC-133,TC-134,TC-135,TC-136,TC-137,TC-138,TC-139,TC-140,TC-141,TC-142,TC-143,TC-196,TC-197,TC-198,TC-199,TC-200,TC-201,TC-202,TC-203,TC-204,TC-205,TC-206,TC-207,TC-208,TC-209,TC-210` | — |
| `R011` | transaction mode, framing and supported control statements | `TC-065,TC-066,TC-067,TC-068,TC-069,TC-070,TC-071,TC-072,TC-073,TC-074,TC-075,TC-076,TC-077,TC-078,TC-079,TC-080,TC-081,TC-082,TC-083,TC-084,TC-085,TC-086,TC-087,TC-088,TC-089,TC-090,TC-091,TC-092,TC-093,TC-094,TC-095,TC-096,TC-097,TC-098,TC-099,TC-100,TC-101,TC-102,TC-103,TC-104,TC-105,TC-106,TC-107,TC-108,TC-109,TC-110,TC-111,TC-112,TC-113,TC-196,TC-197,TC-198,TC-199,TC-200,TC-201,TC-202,TC-203,TC-204,TC-205,TC-206,TC-207,TC-208,TC-209,TC-210,TC-307,TC-308,TC-309,TC-332,TC-333` | `TC-269,TC-270,TC-276,TC-278,TC-279,TC-280,TC-281,TC-328,TC-329` |
| `R012` | direction-specific execution metadata and actual timeout application | `TC-065,TC-066,TC-067,TC-068,TC-069,TC-070,TC-071,TC-072,TC-073,TC-074,TC-075,TC-076,TC-077,TC-078,TC-079,TC-080,TC-081,TC-082,TC-083,TC-084,TC-085,TC-086,TC-087,TC-088,TC-089,TC-090,TC-332,TC-333,TC-342,TC-599` | `TC-278,TC-279,TC-280,TC-281,TC-306,TC-329,TC-345,TC-390,TC-541,TC-542,TC-543,TC-544,TC-598` |
| `R013` | statement-bound direction-specific DDL impact | `TC-144,TC-145,TC-146,TC-147,TC-148,TC-149,TC-150,TC-151,TC-152,TC-153,TC-154,TC-155,TC-156,TC-157,TC-158,TC-159,TC-160,TC-161,TC-416,TC-417` | `TC-282,TC-283,TC-298,TC-299,TC-345,TC-415` |
| `R014` | canonical observability profile | `TC-162,TC-163,TC-164,TC-165,TC-166,TC-167,TC-168,TC-169,TC-170,TC-171,TC-172,TC-173,TC-174,TC-175,TC-176` | `TC-255,TC-256,TC-257,TC-258` |
| `R015` | typed immutable authority/evidence reference structure | `TC-177,TC-178,TC-179,TC-180,TC-181,TC-182,TC-183,TC-184,TC-185,TC-186,TC-187,TC-188,TC-189,TC-190,TC-191,TC-192,TC-193,TC-194,TC-195,TC-420` | `TC-249,TC-252,TC-253,TC-254,TC-286,TC-419,TC-627` |
| `R016` | risk-specific structural gates | `TC-052,TC-053,TC-054,TC-055,TC-056,TC-057,TC-058,TC-059,TC-060,TC-061,TC-062,TC-063,TC-064,TC-177,TC-178,TC-179,TC-180,TC-181,TC-182,TC-183,TC-184,TC-185,TC-186,TC-187,TC-188,TC-189,TC-190,TC-191,TC-192,TC-193,TC-194,TC-195,TC-320,TC-321,TC-322,TC-424` | `TC-247,TC-248,TC-249,TC-265,TC-266,TC-267,TC-268,TC-269,TC-270,TC-295,TC-323,TC-324,TC-331` |
| `R017` | data-classification structural gates | `TC-052,TC-053,TC-054,TC-055,TC-056,TC-057,TC-058,TC-059,TC-060,TC-061,TC-062,TC-063,TC-064,TC-177,TC-178,TC-179,TC-180,TC-181,TC-182,TC-183,TC-184,TC-185,TC-186,TC-187,TC-188,TC-189,TC-190,TC-191,TC-192,TC-193,TC-194,TC-195` | `TC-250,TC-251,TC-252,TC-253,TC-254` |
| `R018` | declared dependency graph validity | `TC-047,TC-048,TC-049,TC-050,TC-051` | `TC-259,TC-260,TC-261` |
| `R019` | owner↔touched-schema consistency | `TC-114,TC-115,TC-116,TC-117,TC-118,TC-119,TC-120,TC-121,TC-122,TC-123,TC-124,TC-125,TC-126,TC-127,TC-128,TC-129,TC-130,TC-131,TC-132,TC-133,TC-134,TC-135,TC-136,TC-137,TC-138,TC-139,TC-140,TC-141,TC-142,TC-143,TC-177,TC-178,TC-179,TC-180,TC-181,TC-182,TC-183,TC-184,TC-185,TC-186,TC-187,TC-188,TC-189,TC-190,TC-191,TC-192,TC-193,TC-194,TC-195,TC-609,TC-614` | `TC-284,TC-285,TC-300` |
| `R020` | no sensitive content in validator/CI logs | `TC-304` | — |
| `R021` | exhaustive source-only canonical Stage 2 control registry and six-value disposition integrity | `TC-223,TC-224,TC-225,TC-226,TC-227,TC-228,TC-229,TC-230,TC-231,TC-232,TC-233,TC-234,TC-235,TC-236,TC-237,TC-238,TC-346,TC-347,TC-348,TC-349,TC-350,TC-412,TC-428,TC-429,TC-430,TC-431,TC-432,TC-433` | `TC-301,TC-353,TC-434` |
| `R022` | self-contained test registry and allowed-branch mapping | `TC-223,TC-224,TC-225,TC-226,TC-227,TC-228,TC-229,TC-230,TC-231,TC-232,TC-233,TC-234,TC-235,TC-236,TC-237,TC-238` | — |
| `R023` | disposable PostgreSQL apply/down/baseline/reapply rehearsal with fail-closed propagation: any apply/DOWN/baseline/reapply failure fails the validator/CI pipeline | `TC-239,TC-240,TC-241,TC-242,TC-243,TC-244,TC-245,TC-246` | `TC-290` |
| `R024` | Stage 3.1/3.11 and runtime privilege regression preservation | `TC-239,TC-240,TC-241,TC-242,TC-243,TC-244,TC-245,TC-246` | `TC-290` |
| `R025` | risk-specific rollout structure | `TC-177,TC-178,TC-179,TC-180,TC-181,TC-182,TC-183,TC-184,TC-185,TC-186,TC-187,TC-188,TC-189,TC-190,TC-191,TC-192,TC-193,TC-194,TC-195,TC-320,TC-321,TC-322,TC-424` | `TC-247,TC-248,TC-249,TC-265,TC-266,TC-267,TC-268,TC-269,TC-270,TC-295,TC-323,TC-324` |
| `R026` | production rollback/roll-forward structural declaration | `TC-052,TC-053,TC-054,TC-055,TC-056,TC-057,TC-058,TC-059,TC-060,TC-061,TC-062,TC-063,TC-064` | `TC-296,TC-297` |
| `R027` | lexical/encoding/client-surface integrity | `TC-001,TC-002,TC-003,TC-004,TC-005,TC-006,TC-007,TC-008,TC-009,TC-010,TC-091,TC-092,TC-093,TC-094,TC-095,TC-096,TC-097,TC-098,TC-099,TC-100,TC-101,TC-102,TC-103,TC-104,TC-105,TC-106,TC-107,TC-108,TC-109,TC-110,TC-111,TC-112,TC-113` | `TC-628,TC-629,TC-630,TC-631` |
| `R028` | derived effect inventory and exact down bijection | `TC-196,TC-197,TC-198,TC-199,TC-200,TC-201,TC-202,TC-203,TC-204,TC-205,TC-206,TC-207,TC-208,TC-209,TC-210,TC-392,TC-613` | `TC-271,TC-272,TC-273,TC-274,TC-275,TC-276,TC-277,TC-302` |
| `R029` | deterministic concurrent-index online rule | `TC-114,TC-115,TC-116,TC-117,TC-118,TC-119,TC-120,TC-121,TC-122,TC-123,TC-124,TC-125,TC-126,TC-127,TC-128,TC-129,TC-130,TC-131,TC-132,TC-133,TC-134,TC-135,TC-136,TC-137,TC-138,TC-139,TC-140,TC-141,TC-142,TC-143,TC-144,TC-145,TC-146,TC-147,TC-148,TC-149,TC-150,TC-151,TC-152,TC-153,TC-154,TC-155,TC-156,TC-157,TC-158,TC-159,TC-160,TC-161,TC-615` | `TC-268,TC-269,TC-270,TC-295,TC-324,TC-329,TC-389,TC-390` |
| `R030` | binary-float and local-timezone migration prohibition | `TC-114,TC-115,TC-116,TC-117,TC-118,TC-119,TC-120,TC-121,TC-122,TC-123,TC-124,TC-125,TC-126,TC-127,TC-128,TC-129,TC-130,TC-131,TC-132,TC-133,TC-134,TC-135,TC-136,TC-137,TC-138,TC-139,TC-140,TC-141,TC-142,TC-143` | — |
| `R031` | stable typed validation error taxonomy and log-safety context | `TC-303,TC-304,TC-305` | — |
| `R032` | homogeneous UP transaction class; non-transactional mode limited to concurrent-index effects | `TC-307,TC-308,TC-309` | `TC-328,TC-329,TC-390` |
| `R033` | deterministic safe CREATE/ALTER/constraint/type grammar, FOREIGN KEY non-zero equal list cardinality, nullable constraint-free CREATE TABLE/ADD COLUMN and no `IF NOT EXISTS` masking | `TC-310,TC-311,TC-312,TC-313,TC-314,TC-315,TC-316,TC-317,TC-318,TC-319,TC-334,TC-335,TC-336,TC-343,TC-344,TC-354,TC-355,TC-356,TC-357,TC-358,TC-369,TC-370,TC-371,TC-372,TC-375,TC-376,TC-377,TC-378,TC-379,TC-380,TC-381,TC-382,TC-383,TC-384,TC-385,TC-391,TC-393,TC-413,TC-414,TC-509,TC-510,TC-512,TC-513,TC-514,TC-516,TC-612` | `TC-266,TC-267,TC-330,TC-345,TC-367,TC-368,TC-373,TC-374,TC-511,TC-515` |
| `R034` | machine-derived minimum risk for supported operation classes and no NOT NULL/embedded-constraint laundering | `TC-320,TC-321,TC-322,TC-334,TC-335,TC-336,TC-413,TC-414,TC-424` | `TC-247,TC-265,TC-266,TC-267,TC-268,TC-269,TC-270,TC-295,TC-323,TC-324,TC-331,TC-389,TC-390` |
| `R035` | CI validation-dominance exact-SHA handshake plus conservative migration-SQL job inventory guard | `TC-325,TC-326,TC-337,TC-338,TC-340` | `TC-327,TC-339,TC-341` |
| `R036` | frozen execution/impact numeric bounds, lock/replication enums, built-in type forms, FK default semantics and raw-statement SHA-256 algorithm | `TC-080,TC-151,TC-155,TC-342,TC-343,TC-344,TC-547,TC-549,TC-551` | `TC-266,TC-282,TC-298,TC-299,TC-330,TC-345,TC-541,TC-542,TC-543,TC-546,TC-548,TC-550` |
| `R037` | exact bidirectional equality of TC→R and R→TC semantic coverage edges | `TC-395,TC-396` | `TC-397` |
| `R038` | canonical Stage 2 source-registry purity/completeness, separate P3D derived registry, and registry-summary semantic consistency | `TC-346,TC-347,TC-348,TC-349,TC-350,TC-351,TC-352,TC-398,TC-400,TC-403,TC-412,TC-418,TC-421,TC-422,TC-428,TC-429,TC-430,TC-431,TC-432,TC-433` | `TC-353,TC-399,TC-401,TC-404,TC-423,TC-434` |
| `R039` | exact future SQL identifier grammar and 63-byte no-truncation identity | `TC-385,TC-391,TC-392,TC-393,TC-402,TC-405,TC-406,TC-407` | `TC-389,TC-390,TC-394` |
| `R040` | exact scalar literal/default/CHECK lexical, type, range and precision-scale semantics | `TC-354,TC-355,TC-356,TC-357,TC-358,TC-359,TC-360,TC-361,TC-362,TC-363,TC-364,TC-365,TC-366,TC-369,TC-370,TC-371,TC-372,TC-400,TC-402,TC-410,TC-411,TC-612` | `TC-345,TC-367,TC-368,TC-373,TC-374,TC-401,TC-408,TC-409` |
| `R041` | exact named simple-column CREATE INDEX grammar and clause-family rejection | `TC-375,TC-376,TC-377,TC-378,TC-379,TC-380,TC-381,TC-382,TC-383,TC-384,TC-385,TC-386,TC-387,TC-388,TC-400,TC-402,TC-405,TC-406,TC-407,TC-615` | `TC-389,TC-390,TC-401` |
| `R042` | exact finite-domain registry, scanner-derived statement_class equality, and authority-kind closed-world semantics | `TC-413,TC-414,TC-416,TC-417,TC-418,TC-420,TC-421,TC-422` | `TC-415,TC-419,TC-423` |
| `R043` | single-source staged-rollout evidence binding and authority cardinality; `rollout.plan_ref` forbidden | `TC-425,TC-426,TC-616,TC-618` | `TC-427,TC-617` |
| `R044` | formal manifest aggregate/open-domain field types, observability-key references and non-normalizing required-text policy | `TC-435,TC-436,TC-438,TC-440,TC-610,TC-611` | `TC-437,TC-439,TC-441` |
| `R045` | byte-bound Stage 2 source-anchor completeness, obligation-strength qualifier binding, exact fragment hashes, exact-once S2 accountability and source/derived provenance | `TC-443,TC-444,TC-445,TC-446,TC-503` | `TC-442` |
| `R046` | exact canonical authority-path lexical identity, length bounds, no normalization and `(kind,path)` uniqueness | `TC-448,TC-449,TC-450,TC-451,TC-452,TC-453,TC-454,TC-536,TC-538` | `TC-447,TC-535,TC-537` |
| `R047` | exact six-byte ASCII open-text trim set and preservation of accepted decoded UTF-8 content | `TC-455,TC-458` | `TC-456,TC-457` |
| `R048` | literal CREATE TABLE/ADD COLUMN token productions, clause order/multiplicity and structural boundary rejection | `TC-461,TC-462,TC-463,TC-464,TC-467,TC-468,TC-522` | `TC-459,TC-460,TC-465,TC-466,TC-521` |
| `R049` | exact four-way S2 evidence scope: complete-machine / paired-SQL-scope-rejected / partial / no-machine partition, explicit external-behavior remainders, quantified-subject observer closure, and no unrelated/full-proof overclaim | `TC-469,TC-470,TC-471,TC-504,TC-505,TC-517,TC-527,TC-571,TC-591,TC-593,TC-594,TC-595,TC-596,TC-623` | `TC-472,TC-518,TC-528,TC-572,TC-592,TC-597,TC-624` |
| `R050` | aggregate P3D-008 semantic-owner/meta partition, complement-set equality and cross-family mutation coverage | `TC-473,TC-474,TC-475,TC-476,TC-477,TC-478,TC-506,TC-573` | `TC-479,TC-570,TC-590` |
| `R051` | exact scalar type-parameter decimal-token grammar plus precision/scale/varchar bounds | `TC-481,TC-482,TC-483` | `TC-480` |
| `R052` | exact complete CHECK statement envelope/token production | `TC-485,TC-486,TC-487,TC-488` | `TC-484` |
| `R053` | exact literal DOWN token productions for each supported inverse class | `TC-494,TC-495,TC-496,TC-497` | `TC-489,TC-490,TC-491,TC-492,TC-493` |
| `R054` | exact validator `.sql` discovery domain and frozen CI `*.up.sql` execution-subject dominance | `TC-498,TC-499,TC-500,TC-502` | `TC-501` |
| `R055` | exact P3D evidence-mode partition, ID→rule scope registry and broad-claim/leaf-rule structural separation | `TC-507,TC-519` | `TC-508,TC-520` |
| `R056` | GUARD-01 universal evidence observability: observer-bearing universe is derived from canonical S2 dispositions/rule presence; every partial row has explicit machine-subset/remainder, scope-rejected rows prove exclusion only, and complete rows must be fully observable | `TC-527,TC-559,TC-560` | `TC-528,TC-534,TC-561` |
| `R057` | GUARD-02 finite-bound inventory with generic all-integer manifest discovery, single-source BOUND-SPEC exact values, boundary PASS/NEG pairs, and new-unregistered-bound mutation rejection | `TC-522,TC-529,TC-536,TC-538,TC-540,TC-547,TC-549,TC-551,TC-552,TC-562,TC-563,TC-576,TC-577,TC-578,TC-579` | `TC-521,TC-530,TC-534,TC-535,TC-537,TC-539,TC-541,TC-542,TC-543,TC-544,TC-545,TC-546,TC-548,TC-550,TC-553,TC-564,TC-580` |
| `R058` | GUARD-03 exact single-source ColId lexical ownership; anchored COLID_POLICY_SPEC binds PostgreSQL REL_18_6 category projection, contextual acceptance and deterministic witnesses | `TC-523,TC-526,TC-531,TC-554,TC-555,TC-556,TC-557,TC-565,TC-581,TC-582,TC-583` | `TC-524,TC-525,TC-532,TC-534,TC-558,TC-566,TC-584` |
| `R059` | GUARD-04 semantic-atom ownership plus frozen normative-line candidate accountability; no latent ownerless/multiply authoritative semantic R and no silent normative-candidate insertion | `TC-531,TC-533,TC-567,TC-569,TC-574` | `TC-532,TC-534,TC-568,TC-570` |
| `R060` | GUARD-05 mutation-kill sufficiency for frozen semantic atoms plus new-unregistered-candidate injections in bounds, authority projections and discovered sibling universes | `TC-526,TC-533,TC-552,TC-557,TC-560,TC-562,TC-563,TC-565,TC-567,TC-569,TC-571,TC-573` | `TC-534,TC-575,TC-600` |
| `R061` | GUARD-06 remediation generalization: Reviewer finding → generic predicate + deterministic discovery rule + derived sibling domain + permanent regression, never named-fixture/expected-set-only closure | `TC-533,TC-559` | `TC-534,TC-575,TC-600` |
| `R062` | GUARD-03/04 external-authority projection: exact PostgreSQL REL_18_6 kwlist category sets + gram.y ColId production compose to the project ColId-disallowed semantic set; exact bytes/hash alone cannot prove a different property | `TC-554,TC-555,TC-556,TC-557` | `TC-558,TC-566` |
| `R063` | anti-circular completeness: S2 observer-bearing universe is disposition-derived and bounded-integer candidates are schema-derived; both must exactly equal their proof registries, while hand-authored semantic sibling counts are never authority | `TC-552,TC-559,TC-560` | `TC-553,TC-561,TC-575` |
| `R064` | exact ATOM-001…050 ↔ semantic-owner-R bijection and exact R-body digest binding; new/missing/duplicate/mutated atom fails | `TC-567,TC-569,TC-573,TC-574,TC-585,TC-586,TC-587,TC-588,TC-589` | `TC-568,TC-570,TC-575,TC-590` |
| `R065` | frozen normative-line accountability: deterministic semantic-candidate line extraction must exactly equal packaged evidence; Builder validates but never regenerates evidence | `TC-562,TC-563,TC-567,TC-574` | `TC-564,TC-568,TC-575` |
| `R066` | single-source ColId authority: anchored §15.3a COLID_POLICY_SPEC exact digest must equal packaged PostgreSQL projection evidence; no duplicate production may satisfy proof | `TC-565,TC-581,TC-582,TC-583` | `TC-566,TC-575,TC-584` |
| `R067` | S2 evidence-edge honesty regression: S2-101 has no machine subset and therefore carries no R018 edge; semantic subset adequacy for remaining partial edges is Reviewer-owned | `TC-571` | `TC-572,TC-575` |
| `R068` | global physical single-source authority: all `BOUND-SPEC|` and `COLID_POLICY|` occurrences are inventoried over the entire candidate and may exist only in their one anchored authority block; duplicate identical/conflicting/relocated copies fail | `TC-576,TC-577,TC-578,TC-579,TC-581,TC-582,TC-583` | `TC-580,TC-584,TC-600` |
| `R069` | occurrence-level ATOM bijection: physical declaration count, ATOM-ID multiplicity, semantic-owner multiplicity and per-occurrence R-body digest are checked before map construction | `TC-585,TC-586,TC-587,TC-588,TC-589` | `TC-590,TC-600` |
| `R070` | derived S2 partition-summary and evidence-edge closure: canonical rows derive the sole active partition counts; every confirmed or proactively identified prerequisite-not-logical-subset S2 row is reviewer/external-evidence only and the exact remaining partial set is frozen for semantic review | `TC-591,TC-593,TC-594,TC-595,TC-596,TC-601,TC-602,TC-603,TC-604,TC-605,TC-606,TC-607` | `TC-592,TC-597,TC-600,TC-608` |
| `R071` | positive acceptance for all four direction-independent execution metadata fields: valid UP and DOWN `expected_duration_seconds`, `lock_risk`, `lock_timeout_ms`, and `statement_timeout_ms` values may differ independently; equality coupling for any field is forbidden | `TC-599` | `TC-598,TC-600` |

| `R072` | SEM/CARD active-line byte-accountability: every active physical line is exact-accounted; these manifests are explicitly non-semantic and cannot authorize machine behavior | `TC-619,TC-620,TC-626` | `TC-625` |
| `R073` | structural-cardinality accountability/index: BND is quantitative-only; structural cardinality semantics are owned by TC↔MPROP and optional compact PROP-SPEC mirrors, while CARD-PROP is byte-accountability only | `TC-609,TC-610,TC-611,TC-613,TC-616,TC-618,TC-622` | `TC-625` |
| `R074` | exact compact PROP-SPEC mirror binding for selected high-risk quantitative/set relations; canonical acceptance/rejection authority remains the exact direct TC↔MPROP property | `TC-609,TC-610,TC-611,TC-612,TC-613,TC-614,TC-615,TC-616,TC-618,TC-619,TC-620,TC-621,TC-622,TC-626` | `TC-617,TC-625` |
| `R075` | current conservative Stage-2 evidence relevance: prerequisite-only S2 rows including S2-107 remain machine-edge-free; exact inverse hardening is not evidence of production-down safety | `TC-623` | `TC-624` |

A rule with a rejection path must have negative coverage. A rule needs positive coverage only where
it intentionally exposes an allowed branch. Pure prohibitions do not get fake positive cases.

`R037` additionally requires exact equality of the complete edge sets:

```text
TC_EDGE_SET = {(test_id, rule_id) declared by every TC row}
R_EDGE_SET  = {(test_id, rule_id) represented by the generated R projection}

TC_EDGE_SET == R_EDGE_SET
```

No report may choose one representation and ignore the other.

### 24.3 Allowed-branch registry — closes P3-08-PLAN-05

Every intentionally accepted enum value, cross-field exception, statement class, inverse class and
mode has a stable branch ID mapped to one or more positive tests.

| Allowed branch ID | Positive test IDs |
| --- | --- |
| `ALLOWED-AUTH-REF` | `TC-286`, `TC-447`, `TC-535`, `TC-537`, `TC-539` |
| `ALLOWED-CI-REHEARSAL` | `TC-290` |
| `ALLOWED-CLASS-FINANCIAL` | `TC-251` |
| `ALLOWED-CLASS-IDENTITY` | `TC-252` |
| `ALLOWED-CLASS-MIXED` | `TC-254` |
| `ALLOWED-CLASS-SCHEMA` | `TC-250` |
| `ALLOWED-CLASS-SENSITIVE` | `TC-253` |
| `ALLOWED-DEFAULT-BOOLEAN` | `TC-293` |
| `ALLOWED-DEFAULT-NUMERIC` | `TC-292`, `TC-367`, `TC-408`, `TC-409` |
| `ALLOWED-DEFAULT-STRING` | `TC-294`, `TC-368` |
| `ALLOWED-DEP-EARLIER` | `TC-261` |
| `ALLOWED-DEP-LEGACY` | `TC-260` |
| `ALLOWED-DEP-NONE` | `TC-259` |
| `ALLOWED-DISPOSITION-ENUM` | `TC-301` |
| `ALLOWED-DOWN-DROP-COLUMN` | `TC-273`, `TC-490` |
| `ALLOWED-DOWN-DROP-CONSTRAINT` | `TC-274`, `TC-491` |
| `ALLOWED-DOWN-DROP-INDEX` | `TC-275`, `TC-492` |
| `ALLOWED-DOWN-DROP-INDEX-CONCURRENT` | `TC-276`, `TC-493` |
| `ALLOWED-DOWN-DROP-TABLE` | `TC-272`, `TC-489` |
| `ALLOWED-DOWN-MULTI-EFFECT` | `TC-277` |
| `ALLOWED-EXEC-LOCK-RISK-ENUM` | `TC-306` |
| `ALLOWED-ID-GAP` | `TC-262` |
| `ALLOWED-IMPACT-DOWN` | `TC-283` |
| `ALLOWED-IMPACT-UP` | `TC-282` |
| `ALLOWED-INDEX-UNIQUE-NEW-TABLE` | `TC-295` |
| `ALLOWED-LEGACY` | `TC-287` |
| `ALLOWED-LOCAL-MODE` | `TC-289` |
| `ALLOWED-LOCK-MODE-ENUM` | `TC-298` |
| `ALLOWED-OBS-MEASURED` | `TC-256` |
| `ALLOWED-OBS-MIXED-PROFILE` | `TC-255` |
| `ALLOWED-OBS-NA-MISMATCH` | `TC-258` |
| `ALLOWED-OBS-NA-ROWS` | `TC-257` |
| `ALLOWED-OWNER-ENUM` | `TC-300` |
| `ALLOWED-OWNER-MULTI` | `TC-285` |
| `ALLOWED-OWNER-SINGLE` | `TC-284` |
| `ALLOWED-PHASE-EXPAND` | `TC-291` |
| `ALLOWED-PR-BASE` | `TC-288` |
| `ALLOWED-PROD-ROLLBACK-APP` | `TC-296` |
| `ALLOWED-PROD-ROLLBACK-LEAVE` | `TC-297` |
| `ALLOWED-REPLICATION-IMPACT-ENUM` | `TC-299` |
| `ALLOWED-REVERSIBILITY` | `TC-302` |
| `ALLOWED-RISK-HIGH` | `TC-249` |
| `ALLOWED-RISK-LOW` | `TC-247` |
| `ALLOWED-RISK-MEDIUM` | `TC-248` |
| `ALLOWED-TIMEOUT-DOWN-NONTXN` | `TC-281` |
| `ALLOWED-TIMEOUT-DOWN-TXN` | `TC-279` |
| `ALLOWED-TIMEOUT-UP-NONTXN` | `TC-280` |
| `ALLOWED-TIMEOUT-UP-TXN` | `TC-278` |
| `ALLOWED-UP-DOWN-EXECUTION-METADATA-INDEPENDENT` | `TC-598` |
| `ALLOWED-UP-ADD-CHECK-NOT-VALID` | `TC-266`, `TC-373`, `TC-374`, `TC-484` |
| `ALLOWED-UP-ADD-COLUMN-LITERAL-DEFAULT` | `TC-265`, `TC-466` |
| `ALLOWED-UP-ADD-COLUMN-NULLABLE` | `TC-264`, `TC-465` |
| `ALLOWED-UP-ADD-FK-NOT-VALID` | `TC-267`, `TC-511`, `TC-515` |
| `ALLOWED-UP-CREATE-TABLE` | `TC-263` |
| `ALLOWED-UP-INDEX-NEW-TABLE` | `TC-268`, `TC-389` |
| `ALLOWED-UP-INDEX-PREEXISTING` | `TC-269` |
| `ALLOWED-UP-MULTI-EFFECT` | `TC-271` |
| `ALLOWED-UP-UNIQUE-INDEX-PREEXISTING` | `TC-270`, `TC-390` |
| `ALLOWED-CI-VALIDATION-DOMINANCE` | `TC-327` |
| `ALLOWED-CI-GO-RACE-DOMINANCE` | `TC-339` |
| `ALLOWED-CI-SQL-EXECUTION-INVENTORY` | `TC-341`, `TC-501` |
| `ALLOWED-UP-HOMOGENEOUS-TXN` | `TC-328` |
| `ALLOWED-UP-HOMOGENEOUS-NONTXN` | `TC-329` |
| `ALLOWED-UP-CREATE-TABLE-FROZEN-GRAMMAR` | `TC-330`, `TC-459`, `TC-460`, `TC-480`, `TC-521`, `TC-524`, `TC-525` |
| `ALLOWED-FROZEN-SEMANTIC-CONSTANTS` | `TC-345`, `TC-479`, `TC-530`, `TC-532`, `TC-541`, `TC-542`, `TC-543`, `TC-544`, `TC-545`, `TC-546`, `TC-548`, `TC-550`, `TC-553`, `TC-558`, `TC-561`, `TC-564`, `TC-566`, `TC-568`, `TC-570`, `TC-572`, `TC-575`, `TC-580`, `TC-584`, `TC-590`, `TC-592`, `TC-597`, `TC-600`, `TC-608` |
| `ALLOWED-V19-S2-EVIDENCE-PARTITION` | `TC-624` |
| `ALLOWED-V19-PROPERTY-ACCOUNTABILITY` | `TC-625` |
| `ALLOWED-AUTHORITY-SAME-PATH-DIFFERENT-KIND` | `TC-627` |
| `ALLOWED-SQL-KEYWORD-ASCII-CASE-INSENSITIVE` | `TC-628` |
| `ALLOWED-SQL-LINE-COMMENT-BETWEEN-TOKENS` | `TC-629` |
| `ALLOWED-SQL-BLOCK-COMMENT-BETWEEN-TOKENS` | `TC-630` |
| `ALLOWED-SQL-FORBIDDEN-WORD-INERT-COMMENT` | `TC-631` |
| `ALLOWED-RISK-LOW-OPERATION-FLOOR` | `TC-331` |
| `ALLOWED-RISK-MEDIUM-CONSTRAINT` | `TC-323` |
| `ALLOWED-RISK-MEDIUM-INDEX` | `TC-324` |
| `ALLOWED-SOURCE-DERIVED-REGISTRY` | `TC-353`, `TC-404`, `TC-442`, `TC-472`, `TC-508`, `TC-518`, `TC-520`, `TC-528`, `TC-534` |
| `ALLOWED-IDENTIFIER-63-BYTE` | `TC-394` |
| `ALLOWED-RULE-TEST-EDGE-EQUALITY` | `TC-397` |
| `ALLOWED-REGISTRY-SUMMARY-CONSISTENT` | `TC-399` |
| `ALLOWED-SEMANTIC-LINT-CLEAN` | `TC-401` |
| `ALLOWED-STATEMENT-CLASS-ENUM` | `TC-415` |
| `ALLOWED-AUTHORITY-KIND-ENUM` | `TC-419` |
| `ALLOWED-FINITE-DOMAIN-REGISTRY` | `TC-423` |
| `ALLOWED-STAGED-ROLLOUT-SOLE-AUTHORITY` | `TC-427` |
| `ALLOWED-STANDARD-ROLLOUT-ZERO-STAGED-REF` | `TC-617` |
| `ALLOWED-SOURCE-FIDELITY-V10` | `TC-434` |
| `ALLOWED-ROLLOUT-METRIC-KEY-REFS` | `TC-437` |
| `ALLOWED-MONITORING-SIGNAL-KEY-REFS` | `TC-439` |
| `ALLOWED-OPEN-TEXT-PRESERVED` | `TC-441`, `TC-456`, `TC-457` |

### 24.4 Exact completeness and semantic-freeze checks

Stage 3.55 focused tests must assert exact set equality for:

```text
source anchors:    SA-001…SA-082
source controls:   S2-001…S2-168
derived controls:  P3D-001…P3D-027
finite domains:    FD-001…FD-019
machine rules:     R001…R075
test cases:        TC-001…TC-631
semantic edges:    1404 exact TC↔R edges
S2 evidence modes: 7 complete-machine + 46 paired-SQL-scope-rejected + 36 partial-machine + 79 no-machine = 168 exact
P3D evidence modes: 22 complete-machine + 5 structure-plus-human = 27 exact
semantic freeze:    exact 50-rule derived `SEMANTIC_FREEZE_RULE_SET`
semantic owner index: ATOM-001…ATOM-050 exact owner/digest binding; property byte-accountability: SEMANTIC_PROPERTY_MANIFEST exact; machine acceptance/rejection semantics: TC↔MPROP exact
normative lines:    frozen `NORMATIVE_LINE_ACCOUNTABILITY` exact against candidate bytes
bound authority:    one `BOUND_SPEC_REGISTRY` row per non-singleton formal integer field, exact values + BND owner
ColId authority:    one anchored `COLID_POLICY_SPEC`, exact REL_18_6 projection and packaged authority evidence
allowed branches:  exact ALLOWED-* set in §24.3
```

The implementation/closure report publishes these separately:

```text
required_source_control_ids
implemented_source_control_ids
required_derived_control_ids
implemented_derived_control_ids
required_finite_domain_ids
implemented_finite_domain_ids
finite_domain_member_set_mismatches
required_rule_ids
implemented_rule_ids
required_test_ids
executed_test_ids
missing_test_ids
duplicate_test_ids
tc_to_rule_edges
rule_to_test_edges
rule_test_edge_symmetric_difference
required_allowed_branch_ids
allowed_branch_to_positive_test_mapping
unmapped_positive_test_ids
multiply_mapped_positive_test_ids
unmapped_allowed_branch_ids
semantic_freeze_placeholder_failures
registry_summary_range_count_mismatches
```

Hard invariants:

```text
TC_EDGE_SET == R_EDGE_SET
positive_test_mapping covers every POS TC exactly once
every allowed branch maps >=1 POS TC
source and derived control sets are disjoint by namespace
finite-domain ID set = `FD-001…FD-019` exactly
finite-domain fields are closed-world and defined only by §7.10
every summary count is computed from an exact ID range/set
semantic-freeze unresolved-placeholder count = 0
semantic rule classification covers R001…R075 exactly once
SEMANTIC_FREEZE_RULE_SET = all R rules minus exact meta/proof-only set
ATOM owner set == SEMANTIC_FREEZE_RULE_SET exactly, with exact R-body digests
NORMATIVE_LINE_ACCOUNTABILITY candidate-line identity == current planning candidate exactly
every formal-manifest integer field resolves exactly once to FD-001 singleton or BOUND_SPEC_REGISTRY; every BOUND_SPEC value == its BND authority
COLID_POLICY_SPEC occurs exactly once in its anchored normative block and equals packaged REL_18_6 authority projection
S2-101 has no machine R edge; independent deployability remains Architecture/reviewer-owned
P3D machine-binding scope registry = EXACT
FOREIGN KEY list proof = MISMATCH_BOTH_DIRECTIONS + EMPTY + DUPLICATE_LOCAL + DUPLICATE_REF + 33_REJECT + 32_PASS
finite-domain field set and every domain-member set = EXACT
```

Semantic-freeze lint scope is current-version normative §§7–23 plus closure/decision summaries §§30/33.
It excludes the forensic finding/H ledgers and negative-test descriptions, where historical bad phrases
may be quoted intentionally. In lint scope, unresolved meta-phrases such as "safe-literal", "simple
column-key grammar", "freeze later", "implementation-defined enum/bound", or equivalent delegation to
Stage 3.55 are forbidden. For finite domains, non-exhaustive language such as `include`, `for example`, or `frozen in code/tests` is also forbidden.


Any missing/duplicate/unknown ID, namespace mix, asymmetric rule/test edge, stale summary, unresolved
semantic placeholder, unmapped positive or unmapped allowed branch fails `MIG021_POLICY_COVERAGE`,
`MIG025_TEST_CONTRACT` or the stable meta-validation error assigned in Stage 3.55.

Headline counts are diagnostic only. Exact sets, exact semantic edges and literal grammar are authoritative.

### 24.4a Preventive Guard proof obligations — v20 TC-atomic machine-property level

These obligations strengthen pre-review prevention without renumbering historical `PO-01…20`:

| Guard proof | Required result |
| --- | --- |
| `GP-01 OBSERVER_UNIVERSE_CLOSURE` | migration-subject candidates are discovered from all S2 rows; declared registry equals discovery; every observer-bearing candidate has explicit subject/observer subset + remainder; no hand-typed expected set may be completeness authority |
| `GP-02 FINITE_BOUNDARY_PAIRS` | all formal integer fields are discovered without name allowlists and resolve to FD singleton or exact BOUND-SPEC; numeric values, BND witnesses and mutation tests must agree |
| `GP-03 LEXICAL_INTERSECTION_CLOSURE` | one anchored COLID_POLICY_SPEC is the sole normative lexical authority; its digest and upstream projection evidence must agree and duplicate prose cannot satisfy proof |
| `GP-04 SEMANTIC_ATOM_OWNERSHIP` | owner-level ATOM index remains exact; SEM-PROP is byte-accountability/index evidence binding active physical lines to exact digest/owner attribution only; machine acceptance/rejection completeness is owned by exact TC↔MPROP; one R may own many MPROPs; semantic adequacy remains independently reviewed |
| `GP-05 MUTATION_KILL` | the packaged executable mutation suite kills all required existing-value, new-candidate, authority, atom and evidence-edge mutations with frozen checker/evidence |
| `GP-06 REMEDIATION_GENERALIZATION` | each residual remediation records generic predicate + discovery rule + resulting sibling domain; registry equality is checked against discovery, so named fixture or expected count alone cannot close it |

`BUILDER_PROOF_OBLIGATIONS_PASS` requires exact PASS of `PO-01…20` **and** `GP-01…06`.

### 24.5 Builder proof obligations — v20 TC-atomic / single-taxonomy proof architecture

The Builder gate is no longer satisfied by registry continuity/counts alone. Before packaging any later
Stage 3.54 candidate, the Builder must evaluate these permanent proof obligations against the exact
candidate and exact canonical evidence. A `PASS` is scoped to what can actually be mechanically or
documentarily established; human/operational adequacy remains outside machine proof.

| Proof obligation | Required proof |
| --- | --- |
| `PO-01 SOURCE COMPLETENESS` | every canonical normative Stage 2 source unit, including obligation-strength bytes such as mandatory/never/required qualifiers, is inside accountable `SA-*` bytes and has an `S2-*` disposition; internal S2 continuity or line accounting alone is insufficient |
| `PO-02 SOURCE FIDELITY` | no canonical source subject, obligation-strength qualifier, conjunction or scope is silently lost, imported from excluded bytes, or strengthened inside `S2-*` |
| `PO-03 MANIFEST COMPLETENESS` | every manifest field is defined by a normative section with exact type/requiredness/open-vs-closed semantics |
| `PO-04 DOMAIN COMPLETENESS` | every finite field is present in `FD-*` with exact exhaustive members and supported subset |
| `PO-05 CROSS-FIELD COMPLETENESS` | every documented field relationship/cardinality is an explicit invariant/rule, including rollout↔authority binding |
| `PO-06 SQL CLOSED-WORLD` | every accepted SQL form belongs to the exact reviewed grammar; everything else rejects/rescopes |
| `PO-07 RISK MONOTONICITY` | no supported syntax can lower effective risk below the canonical/derived floor |
| `PO-08 EXECUTION BINDING` | declared timeout/transaction semantics equal the actual validated PostgreSQL execution controls |
| `PO-09 DOWN INVERSE SAFETY` | every allowed destructive DOWN effect is the exact safe inverse of a corresponding UP effect |
| `PO-10 COVERAGE SINGLE-SOURCE` | duplicate normative representations of the same semantic identity are forbidden or generated from one canonical declaration |
| `PO-11 POSITIVE COMPLETENESS` | every intentionally allowed branch has positive acceptance evidence |
| `PO-12 NEGATIVE COMPLETENESS` | every reviewed invariant violation has deterministic negative regression evidence, including both directions of any list-cardinality/equality boundary plus zero/duplicate/max-boundary cases such as FOREIGN KEY local↔referenced column lists |
| `PO-13 CI VALIDATION DOMINANCE` | no repository CI migration SQL execution precedes exact-SHA validator success **and every executed SQL file belongs to the exact validator discovery/approval subject set** |
| `PO-14 HISTORICAL IMMUTABILITY` | protected historical migration path+bytes cannot be changed/replaced/reclassified |
| `PO-15 EVIDENCE-BOUNDARY HONESTY` | every S2→R and P3D→R machine binding proves only a property within the English requirement; external/runtime/operational remainder is explicit, broad claims bind complete owners, and unrelated/narrow rules cannot masquerade as full proof |
| `PO-16 NO IMPLEMENTATION-DEFINED SEMANTICS` | Stage 3.55 cannot invent enum values, aliases, bounds, grammar, defaults, reference identity or interpretation; every non-meta machine rule is classified into the aggregate semantic-owner complement and no semantic owner is omitted |
| `PO-17 REGISTRY CONSISTENCY` | exact IDs/ranges/references/polarity/projections are internally consistent |
| `PO-18 LEGACY BOUNDARY` | `000001…000007` exemptions remain identity-frozen and cannot expand to future migrations |
| `PO-19 FAILURE CLOSED-WORLD` | unknown, malformed, ambiguous or unsupported input rejects rather than guessing |
| `PO-20 YAGNI / IMPLEMENTABILITY` | the approved design remains implementable with the intended narrow Go/std-lib approach and does not silently require a full PostgreSQL parser/framework |

Builder success marker:

```text
BUILDER_PROOF_OBLIGATIONS_PASS
```

means only that the above obligations pass for the exact packaged candidate under the inspected evidence.
It is **not** reviewer approval and is **not** a claim that no undiscovered defect can exist.

New-review classification rule:

1. identify the failed/missing `PO-*` first;
2. determine whether the defect is a residual/regression of `P3-08-PLAN-01…18`;
3. append the next unused `P3-08-PLAN-NN` only for a genuinely new failure class;
4. add/strengthen the invariant and deterministic regression, not merely another threat example.

## 25. Stage 3.55 implementation architecture / YAGNI

The larger policy/test contract does **not** authorize a large framework.

Preferred surface remains one command package:

```text
backend-go/cmd/validate-migrations/
  main.go
  validator.go
  manifest.go
  scanner.go
  policy.go
  *_test.go
```

plus:

```text
infrastructure/postgres/migrations/policy_manifest.json
.github/workflows/ci.yml
docs/stages/STAGE_03_55_P3_08_MIGRATION_VALIDATOR_IMPLEMENTATION.md
```

Tests may be split by concern for reviewability.

Constraints:

- Go standard library only;
- no external SQL parser;
- no migration framework;
- no parser generator;
- no runtime service/library;
- no new SQL migration;
- no schema/data change;
- no frontend/OpenAPI/business-logic change;
- no dependency/lockfile change.

### Complexity stop rules

Stop and rescope before Internal review if:

- safe classification requires general PostgreSQL semantic parsing;
- dynamic/procedural SQL support would be needed;
- unsupported DDL forms are being added merely to make synthetic tests pass;
- production rollback semantics start depending on disposable down files;
- policy registry cannot remain deterministic;
- production validator logic exceeds the canonical hand-written business-logic review budget without a
  pre-review Principal Architect exception.

The preferred solution to unsupported syntax is fail-closed rejection, not parser expansion.

## 26. Explicitly out of scope

P3-08 MUST NOT include:

- a new SQL migration;
- modification of `000001`–`000007` migration SQL;
- production schema/data mutation;
- application runtime migration execution;
- support for Populate/Switch/Validate/Contract inside paired-SQL v1;
- a third-party migration library or SQL parser;
- Stage 4/future migration-tool adoption;
- Vault/key deletion/backup implementation;
- tax schema;
- OpenAPI/frontend/business-logic changes;
- financial arithmetic changes;
- auth/session behavior changes;
- P3-07 reopening;
- Stage 3.25 Security Review implementation;
- new dependency/lockfile;
- reduction of the protected ten-check CI inventory.

Any requirement for these surfaces is a fail-closed scope-expansion event.

## 27. Local implementation gates

Before Internal review, Stage 3.55 must run at least:

```text
cd backend-go && go test ./cmd/validate-migrations
cd backend-go && go test ./...
cd backend-go && go vet ./...
cd backend-go && go run ./cmd/validate-migrations --mode=local
docker compose config --quiet
```

Focused validator output must also prove:

```text
CONTROL_ID_SET=EXACT
RULE_ID_SET=EXACT
TEST_ID_SET=EXACT
ALLOWED_BRANCH_SET=EXACT
UNMAPPED_ALLOWED_BRANCHES=0
MISSING_REQUIRED_TESTS=0
DUPLICATE_REQUIRED_TESTS=0
```

plus a disposable PostgreSQL apply → exact disposable down inverse → baseline assertion → reapply →
equivalence gate where Docker is available.

The published head must pass all ten required GitHub CI checks.

## 28. Review-size / complexity stop rules

Canonical limits remain:

- <= 25 changed files;
- <= 800 changed lines of hand-written business logic unless a documented pre-review exception exists.

Additional P3-08 stop rules:

- if strict SQL guardrails require a real PostgreSQL parser, stop and rescope;
- if manifest validation requires a third-party schema library, stop and justify/rescope;
- if PR-base verification cannot be deterministic with standard Git + Go tooling, stop rather than
  silently degrade to candidate checksums;
- if schema fingerprinting is formatting-brittle, use catalog invariants instead of adding fragile
  textual equality.

## 29. Development review lifecycle

Stage 3.55 changes validator code and CI, so REVIEW_WORKFLOW v1.4.0 development path is mandatory:

1. approved/merged Stage 3.54 plan;
2. feature branch;
3. implementation/tests/record;
4. local gates;
5. complete read-only Internal review;
6. Builder-only fixes + rerun;
7. separate human commit/push authorization;
8. Draft PR;
9. exact-head ten-check CI;
10. fresh same-chat External published-head review independent of Internal verdict;
11. fixes/CI rerun if needed;
12. External verdict;
13. required Internal evidence publication only after External verdict;
14. CI on evidence-only head where required;
15. exact publication verification;
16. separate human Ready authorization;
17. separate human squash-merge authorization;
18. implementation merge;
19. separately governed closure before P3-08 changes state.

## 30. Closure model

A later Stage 3.56 governance/closure record may set P3-08 CLOSED only after protected Stage 3.55
implementation and complete evidence.

Closure must verify at minimum:

- new validator canonical on protected `develop`;
- strict manifest and base-relative immutability active;
- exact seven legacy pairs pass without retroactive metadata claims;
- fourteen historical SQL files byte-immutable;
- `P3-08-PLAN-01…18` all reviewer-confirmed remediated;
- exact H-register `P3-08-H01…H269` retained/dispositioned;
- exact `SA-001…SA-082` source-anchor ranges/hashes/accountability recompute against the canonical Stage 2 blob;
- exact canonical Stage 2 source control-ID set `S2-001…S2-168` plus derived set `P3D-001…P3D-027` implemented with reviewed dispositions;
- exact finite-domain set `FD-001…FD-019` and member sets implemented;
- exact machine-rule set `R001…R075` implemented, including R055 P3D evidence-scope enforcement;
- exact self-contained test set `TC-001…TC-631` executed with no missing/duplicate IDs and FK cardinality/list boundary rejects (2→1,1→2,empty,duplicate,33) plus 2→2 and 32→32 passes;
- exact one-to-one `MPROP-001…MPROP-631` machine-property registry equals the canonical TC set and binds exact condition/outcome/owners/direct witness polarity;
- registry-derived mutation obligations cover every MPROP with no Builder-selected omission;
- exactly one TAXONOMY-AUTHORITY defines BND/structural-cardinality/MPROP/ATOM/SEM-CARD/NLA scope;
- every active observer-bearing numeric summary derives to 89 from 7+46+36 and closes with 79 no-machine to 168;
- exact allowed-branch registry has zero unmapped positive branches;
- paired-SQL v1 supports only the reviewed Expand surface;
- hidden Populate/Switch/Validate/Contract cannot masquerade as supported paired SQL;
- procedural/dynamic/psql execution surfaces fail closed;
- direction-specific timeout values are actually applied to PostgreSQL SQL bytes;
- direction-specific DDL impact is complete;
- disposable down inverses are exact scoped effects and are not represented as production rollback approval;
- canonical observability mapping is exact;
- risk/classification/rollout/authority structural gates are active without human-evidence overclaim;
- exact S2 evidence-binding partition is `7 complete-machine + 46 paired-SQL-scope-rejected + 36 partial-machine + 79 no-machine`, with S2-109/S2-118 partial and the conservative reviewer/external-evidence no-machine set covering lifecycle-sequencing, compatibility, runtime-reporting, exact-risk-classification and priority-policy rows where machine rules would prove only prerequisites;
- P3D-008 aggregate semantic-freeze partition covers `R001…R075` exactly once, with exact meta/proof-only exclusions and a derived 50-rule semantic-owner complement including R033 and R058;
- scalar type-parameter, FOREIGN KEY exact list grammar/cardinality, CHECK-envelope and every supported DOWN token language are recursively literal/closed with deterministic positive+negative proof;
- validator `.sql` discovery/approval set covers every frozen CI `*.up.sql` execution subject;
- disposable PostgreSQL apply→down→baseline→reapply evidence is green;
- ten required CI checks are green;
- no runtime/schema/data/API/frontend/dependency change was smuggled into P3-08;
- no unresolved material finding remains.

Only then may original audit arithmetic become:

```text
32/32 = 100%
```

## 31. Planning publication lifecycle and review-subject binding

Stage 3.54 remains exactly one documentation change.

Two review stages are intentionally distinct and approval is **not transferable** between them.

### 31.1 `PRE_PUBLICATION_PACKAGE`

Before any commit/push/PR authorization, the independent planning review subject is byte-bound by:

- exact repository identity;
- exact protected base commit/tree;
- exact previous reviewed candidate blob;
- exact current candidate Git blob + SHA-256;
- exact changed-path set (`EXACT 1 DOC`);
- exact canonical/evidence blobs and package checksums.

The branch label is an intended local worktree label, **not a required remote ref** at this stage. Because this workflow
explicitly forbids commit/push before local approval, absence of the remote branch/PR is expected and is
`NOT_APPLICABLE`, not `NOT_VERIFIED`.

A pre-publication reviewer MUST still return `REQUEST CHANGES` if candidate bytes, base identity, scope, authority
or supplied approval-critical evidence cannot be verified. The reviewer may not silently bind to `develop` or any
mutable remote branch.

### 31.2 `PUBLISHED_EXACT_HEAD`

Only after a clean pre-publication `APPROVED` verdict **and separate explicit human authorization** for commit/push
may the planning candidate be published. The fresh published-head review then requires:

- exact remote branch/ref;
- exact published HEAD commit/tree;
- exact candidate blob at that HEAD;
- exact PR head/base identity where a PR exists;
- exact-head CI/evidence required by governance.

Failure to resolve those published identities is then a `VERIFICATION BLOCKER`.

The pre-publication verdict does not authorize publication and does not substitute for the later exact published-head
review. The later review starts from the published subject independently.

This planning artifact:

- contains no self-authored approval;
- predicts no future PR/CI/head/merge identity;
- grants no implementation/commit/push/Ready/merge authority;
- keeps P3-08 OPEN;
- keeps audit 31/32 = 96.875%;
- follows the documentation-only planning publication path.

## 32. Review output quality requirement

A future strict review verdict is exactly `APPROVED` or `REQUEST CHANGES`. If approval-critical evidence cannot be established, the reviewer records a `VERIFICATION BLOCKER` / `NOT_VERIFIED` basis and returns `REQUEST CHANGES`; that result must not consist only of PASS/FAIL labels.

For every FAIL, the reviewer should provide a numbered finding with:

- affected section;
- violated requirement or loophole;
- concrete failure/bypass scenario;
- impact;
- minimal remediation.

This is an evidence-quality requirement for efficient remediation, not authority for the reviewer to
mutate files.

## 33. Planning decision

Proceed to Stage 3.55 only after this **v20 TC-atomic machine-property / single-taxonomy semantic-freeze plan** receives a clean independent
`APPROVED` review and Stage 3.54 itself is published/merged under the required governance gates.

The intended closure design after the retained Stage 3.54 review/remediation history through v18 plus the v20 Builder TC-atomic machine-property / single-taxonomy / registry-derived-mutation proof audit is:

```text
immutable historical SQL
+ identity-only non-retroactive legacy baseline
+ append-only strict manifest
+ per-direction execution/timeouts actually applied in SQL
+ paired-SQL v1 limited to narrow Expand only
+ deterministic UP allowlist
+ exact disposable DOWN inverse model
+ down file explicitly not production rollback authority
+ statement-bound UP/DOWN DDL impact
+ deterministic concurrent-index rule
+ exact Expand observability profile
+ typed risk/classification/rollout/authority structures
+ strict JSON/filesystem/encoding/client-surface rules
+ exact canonical authority-path grammar and six-byte ASCII open-text trim semantics
+ literal CREATE TABLE / ADD COLUMN token productions
+ canonical scalar type-parameter token grammar
+ literal CHECK envelope and literal DOWN inverse productions
+ qualifier-bearing source anchors for obligation-strength bytes
+ conservative four-way S2 complete/scope-rejected/partial/no-machine evidence partition
+ complement-derived all-rule semantic-owner freeze with R033 included
+ exact P3D machine-evidence scope registry
+ deterministic FOREIGN KEY exact list grammar with 1..32 cap, pairwise distinctness, equal cardinality and catalog-boundary separation
+ exact S2 machine-evidence scope partition and partial-binding registry
+ aggregate semantic-freeze rule-set equality/mutation coverage
+ validator-discovery / CI-execution subject dominance
+ stage-correct PRE_PUBLICATION_PACKAGE versus PUBLISHED_EXACT_HEAD review binding
+ declared dependency graph validity without semantic-completeness overclaim
+ PR-base Git immutability
+ byte-bound canonical source-anchor registry `SA-001…SA-082`
+ exact source-only Stage 2 registry `S2-001…S2-168`
+ exact derived hardening registry `P3D-001…P3D-027`
+ exact finite-domain registry `FD-001…FD-019`
+ exact 75-rule machine registry (`R001…R075`)
+ exact `ATOM-001…ATOM-050` semantic-owner registry with R-body digest binding
+ frozen normative-line accountability over the exact v20 planning candidate
+ single-source `BOUND_SPEC_REGISTRY` with generic formal-integer discovery and exact bound-value binding
+ single-source PostgreSQL REL_18_6 `COLID_POLICY_SPEC` anchored to the normative grammar block
+ `S2-101` independent-deployability evidence explicitly reviewer/Architecture-owned with no unrelated R018 edge
+ self-contained 631-case adversarial/acceptance contract (`TC-001…TC-631`)
+ exact one-to-one 631-property machine contract (`MPROP-001…MPROP-631`) generated from TC rows
+ registry-derived property mutation obligations rather than exemplar-selected mutations
+ single authoritative BND/structural-cardinality/MPROP taxonomy
+ prompt↔package executable contract verified after clean unzip
+ exact 89-branch positive acceptance registry with every one of 155 POS tests mapped exactly once
+ permanent `PO-01…PO-20` Builder proof obligations and residual-vs-new finding classification
+ stable typed errors
+ preserved and strengthened disposable PostgreSQL rehearsal
+ explicit machine-vs-human/operational evidence boundary
```

Counts are not accepted as proof by themselves. Exact ID-set equality and mapping completeness are the
authoritative closure conditions.

P3-08 remains OPEN until the later protected Stage 3.56 closure activation.

