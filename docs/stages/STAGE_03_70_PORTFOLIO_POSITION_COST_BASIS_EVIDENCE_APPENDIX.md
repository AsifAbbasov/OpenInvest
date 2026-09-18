

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


Verdict on `a5c7a507a586271b7117580585c88191693144e3`:

```text
P0 = 0
P1 = 1 blocking
P2 = 1 blocking
P3 = 1
VERDICT = changes required
```

Findings:

1. **P1 — WAC scale-8 self-contradiction.** The first candidate simultaneously treated `quantity + acquisitionBasis` as canonical state, re-derived WAC from rounded basis/quantity, and required partial SELL to preserve WAC. Fractional quantities make those rules mutually inconsistent after scale-8 rounding.
2. **P2 — POST transaction conflict-contract mismatch.** The frozen POST transaction contract used `409 IdempotencyConflict`, not a generic business-conflict response, and did not list `429`; the candidate incorrectly described stronger existing surfaces.
3. **P3 — stale canonical version registry.** `DOCUMENT_INDEX.md` still identified SOT-001 as `1.4.63` while the Stage 3.70 SOT candidate was `1.4.90`.

No Ready/merge action was taken after this verdict.


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

### 17.6 Fresh external validation

Canonical record: commit(s) `2af826392e081eb07d3234be1c94833bae768aa7`.

Final pre-evidence External verdict:

```text
P0 = 0
P1 = 0
P2 blocking = 0
P3 blocking = 0
VERDICT = APPROVED
```

Canonical record: PR #137.

### 17.7 Evidence-only publication rule and next gate

This follow-up publishes only the previously withheld Internal evidence plus the already-observed publication/remediation/CI/External chronology. It MUST NOT change Sections 1–16 or any other Stage 3.70/ADR-009 semantic surface.


Only after that verification may the Principal Architect separately:

1. explicitly accept ADR-009 / Stage 3.70;
2. authorize Ready + squash merge of PR #137.

Stage 3.71 runtime implementation remains separately blocked until Stage 3.70 / ADR-009 is canonical on protected `develop` and receives a new explicit implementation authorization.
