## 17. Published review and evidence chronology

This section is evidence-only and is intended for publication only after the fresh External published-head verdict, as required by `docs/REVIEW_WORKFLOW.md` v1.4.0. Sections 1–16 preserve the current Stage 3.70 / ADR-009 financial and architectural decision. This follow-up changes no WAC methodology, ledger-ordering rule, API-planning boundary, runtime, test, OpenAPI file, SQL/migration, dependency, provider, Feature 3D, or Stage 3.71 authorization.

### 17.1 Prepublication Internal review evidence

The mandatory prepublication Internal review was read-only and made no repository edits. Its evidence was withheld from the Draft PR/repository until the External published-head phase completed.

Internal review report SHA-256:

```text
092045012ef45b7c338ea205d4a447abb13358ad493d3495533b68a6ac811bb3
```

Historical prepublication Internal verdict:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The Internal review covered the then-frozen seven-file candidate against `develop@d2258134433fe214695db43db7de3b6bf003e9cf`, tree `9b73746a298a4d36c6dcb0e1037ba45b17dfb0e6`, including Issue #136 admission, the historical `ledgerSequence` backfill tuple, future import-range ordering, `investedCapital` separation, registry synchronization, and merge-activated lifecycle wording.

This Internal verdict is preserved as chronological process evidence, not as proof that the first published semantic head was flawless. The later fresh External review found three material problems in that published candidate and therefore superseded the affected Internal semantic conclusions before merge: the scale-8 WAC state contradiction, the POST transaction 409/429 contract mismatch, and stale SOT version metadata in `DOCUMENT_INDEX.md`.

### 17.2 Initial published head and CI #356

Draft PR #137 initially published:

```text
head:
a5c7a507a586271b7117580585c88191693144e3

tree:
0070b29fae263a7a836abcff433cf9b72cfb6568
```

CI #356 / run `33998907479` completed on that exact head with all ten required jobs `SUCCESS`.

### 17.3 First External published-head verdict

The designated review chat performed a fresh External review of the complete initial published diff and supporting repository evidence without relying on the earlier Internal verdict as proof.

Verdict on `a5c7a507a586271b7117580585c88191693144e3`:

```text
P0 = 0
P1 = 1 blocking
P2 = 1 blocking
P3 = 1
VERDICT = REQUEST CHANGES
```

Findings:

1. **P1 — WAC scale-8 self-contradiction.** The first candidate simultaneously treated `quantity + acquisitionBasis` as canonical state, re-derived WAC from rounded basis/quantity, and required partial SELL to preserve WAC. Fractional quantities make those rules mutually inconsistent after scale-8 rounding.
2. **P2 — POST transaction conflict-contract mismatch.** The frozen POST transaction contract used `409 IdempotencyConflict`, not a generic business-conflict response, and did not list `429`; the candidate incorrectly described stronger existing surfaces.
3. **P3 — stale canonical version registry.** `DOCUMENT_INDEX.md` still identified SOT-001 as `1.4.63` while the Stage 3.70 SOT candidate was `1.4.90`.

No Ready/merge action was taken after this verdict.

### 17.4 Builder remediation and local remediation review

Builder remediation changed exactly three documentation files relative to the first published head:

1. `docs/ADR/ADR-009-deterministic-portfolio-ledger-ordering-and-wac.md`;
2. `docs/stages/STAGE_03_70_PORTFOLIO_POSITION_COST_BASIS_PLANNING.md`;
3. `docs/DOCUMENT_INDEX.md`.

The remediation:

- made `quantity + weightedAverageCost` the authoritative open-position state;
- made acquisition basis the derived `Round8HalfEven(quantity × weightedAverageCost)` amount;
- added the mandatory fractional-quantity non-invertibility regression vector and prohibited inverse repricing of WAC from rounded basis after SELL;
- explicitly planned one narrow Stage 3.71 POST transaction `409` OpenAPI change preserving `IDEMPOTENCY_CONFLICT` while adding `INSUFFICIENT_POSITION_QUANTITY`;
- removed the false claim that POST transaction already exposes `429`;
- synchronized the SOT-001 version in `DOCUMENT_INDEX.md` to `1.4.90`.

Local remediation review SHA-256:

```text
56b85bd03579e0a775ee971fb38a4c1094e1a7e2d2aec7dbdf4fdfb05ed965b3
```

Local remediation verdict:

```text
P0 = 0
P1 blocking = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

That local verdict authorized only the remediation publication request and did not authorize Ready, merge, ADR acceptance, evidence publication, or Stage 3.71.

### 17.5 Remediated semantic head and CI #357

After explicit human remediation commit/push authorization, the branch advanced by one non-force commit:

```text
head:
2af826392e081eb07d3234be1c94833bae768aa7

tree:
8ab46969085f0ad65c8f5ae88b2e572c8e7a5389
```

The semantic-head delta from `a5c7a507...` changed exactly the three remediation documentation files and no runtime/OpenAPI/SQL/dependency file.

CI #357 / run `34000772097` completed on exact head `2af826392e081eb07d3234be1c94833bae768aa7` with all ten required jobs `SUCCESS`:

```text
Go tests                         PASS
Go race tests                    PASS
Go vet                           PASS
Go vulnerability scan           PASS
Python tests                     PASS
Frontend build and typecheck     PASS
OpenAPI contract                 PASS
Docker Compose config            PASS
PostgreSQL migration validation  PASS
Dependency security scan         PASS
```

### 17.6 Fresh External re-review

The designated review chat freshly re-reviewed the complete seven-file PR diff on exact head `2af826392e081eb07d3234be1c94833bae768aa7`, including the remediated WAC rules, current frozen POST transaction contract, Issue #136, ordering/backfill/import boundaries, current snapshot cash/asset semantics, canonical registries, provider exclusions, and CI #357.

Final pre-evidence External verdict:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

The previous P1/P2/P3 findings were verified resolved. No new blocking finding remained, runtime/API-file/SQL/dependency/provider drift was `NONE`, and PR #137 had zero unresolved review threads at the verdict.

### 17.7 Evidence-only publication rule and next gate

This follow-up publishes only the previously withheld Internal evidence plus the already-observed publication/remediation/CI/External chronology. It MUST NOT change Sections 1–16 or any other Stage 3.70/ADR-009 semantic surface.

The evidence head must pass the same required GitHub CI. The designated review chat must then verify that the semantic-head → evidence-head transition is documentation/evidence-only, exact, complete, and introduces no semantic/runtime drift.

Only after that verification may the Principal Architect separately:

1. explicitly accept ADR-009 / Stage 3.70;
2. authorize Ready + squash merge of PR #137.

Stage 3.71 runtime implementation remains separately blocked until Stage 3.70 / ADR-009 is canonical on protected `develop` and receives a new explicit implementation authorization.
