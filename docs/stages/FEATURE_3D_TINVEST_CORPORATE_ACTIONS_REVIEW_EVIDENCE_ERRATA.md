# Feature 3D — Review Evidence Publication Errata

| Field | Value |
| --- | --- |
| Status | Evidence-only correction discovered during post-publication verification |
| Date | 2026-09-09 |
| Repository | `AsifAbbasov/OpenInvest` |
| Branch | `feature/tinvest-corporate-actions` |
| Implementation head | `7c021e74db3a66488c4c6d87412729f524d9023f` |
| First evidence commit | `d8577b924e4227625e2150594fd9c9d978c7ba90` |
| Runtime semantic change | NONE |

## 1. Purpose

During the mandatory exactness/no-semantic-drift verification of the first evidence-only follow-up, the review chat identified two documentation-evidence statements that required correction. This errata preserves the original evidence publication append-only rather than silently rewriting historical evidence.

No runtime code, test, configuration, OpenAPI, dependency, migration, CI workflow, source-rights registry, frontend, provider behavior, or activation state is changed by this correction.

## 2. Correction to INT-3D-010 wording

The first review-evidence publication stated that the frozen implementation dossier had been fully synchronized to include both:

- explicit observation-time `AsOf = RetrievedAt` semantics; and
- deterministic application-generated `SourceEventID` semantics.

That statement was too broad.

The frozen implementation dossier correctly states that `EventID` and internal `SourceEventID` are deterministically derived from normalized provider facts, but one later sentence still calls `SourceEventID` "provider-owned". The adapter code does **not** use a native T-Invest corporate-action event identifier; it generates an application-owned deterministic SHA-256 digest and stores it in the internal provenance `SourceEventID` field.

Correct semantics:

```text
SourceEventID = application-generated deterministic digest of normalized provider evidence
native T-Invest event ID claim = NO
public API exposure of SourceEventID = NO
```

The frozen dossier also does not state the `AsOf` interpretation as explicitly as the review record intended. The selected T-Invest methods do not provide one common dataset-level provider snapshot timestamp suitable for the canonical `AsOf` field. The adapter therefore uses OpenInvest observation time for both mandatory canonical timestamps:

```text
AsOf        = time OpenInvest observed/normalized the provider response
RetrievedAt = same OpenInvest observation time
```

This is **not** a claim that T-Invest updated the source fact at that instant and is **not** a provider freshness/SLA assertion.

These clarifications describe the already-reviewed runtime behavior at implementation SHA `7c021e74db3a66488c4c6d87412729f524d9023f`; they do not change it.

Accordingly, the `INT-3D-010` remediation statement in the first evidence file should be read narrowly: review-size arithmetic, optional-field wording, mixed-sign money wording, and rate-header wording were synchronized in the frozen dossier; the two provenance/time clarifications above are completed by this append-only evidence errata.

## 3. Correction to evidence-only changed-file expectation

The first evidence record stated the expected evidence-follow-up diff as:

```text
changed documentation/evidence files = this file only
```

Because this errata is itself required to correct evidence accuracy, the final cumulative evidence-only diff from implementation head `7c021e74db3a66488c4c6d87412729f524d9023f` is expected to contain exactly these two documentation paths:

1. `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE.md`
2. `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_REVIEW_EVIDENCE_ERRATA.md`

The required invariant remains:

```text
changed implementation/runtime files = 0
changed tests/config/OpenAPI/dependencies/migrations/CI/frontend = 0
provider/runtime semantics = unchanged from 7c021e74db3a66488c4c6d87412729f524d9023f
```

## 4. Review and verdict impact

This errata does not reopen a runtime/security/financial/API finding. It corrects the review-evidence description of behavior already present in the External-reviewed implementation head.

The External published-head verdict on `7c021e74db3a66488c4c6d87412729f524d9023f` remains:

```text
APPROVED
```

The evidence-publication gate remains incomplete until GitHub CI is green on the final evidence head and the designated review chat verifies the cumulative diff as documentation/evidence-only with no semantic drift.

## 5. Activation and authority

This errata authorizes nothing beyond evidence correction.

```text
RUNTIME ACTIVATION = NO
LIVE T-INVEST TOKEN = NOT USED
READY = NOT AUTHORIZED
MERGE = NOT AUTHORIZED
STAGE 3.78 = NOT STARTED
```
