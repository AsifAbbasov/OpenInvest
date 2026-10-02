# Stage 3.78 - Legacy Snapshot Return Unavailability

## Scope

OI-XIRR-001 removes the false implication that persisted legacy snapshot ROI is an approved economic return. OI-XIRR-002 clarifies the intentionally different local-cost summary and terminal-market-value XIRR measures.

## Decision

No approved nominal-return, real-return, inflation, TWR, or simple-return methodology exists for the current snapshot model. In particular, `invested_capital_amount` is cumulative BUY acquisition outflow, while `total_value_amount` includes all remaining external cash. Dividing their difference by BUY outflow therefore mixes capital populations and can manufacture an extreme apparent return.

The remediation does not introduce a substitute formula. Migration `000014_valuation_xirr_legacy_return_unavailable` adds the database-enforced `legacy_return_status = UNAVAILABLE` marker. It applies by constant default to historical rows, so their prior numeric columns are no longer authoritative return data. New writers store only a scale-8 zero placeholder under that marker; the placeholder is not an economic zero return and must be exposed as unavailable.

## Methodology versions

`stage-03-71-position-cost-snapshot-v1` remains historical and unchanged. New normal and ACTIVE-R2 snapshots use `stage-03-71-position-cost-snapshot-v2`; v2 changes only the legacy-return availability semantics. Summary selection ranks v2 above v1 when both exist for a BusinessDate. Existing v1 replay epochs remain readable as compatibility checkpoints, and the next persisted epoch records v2.

## Public contract

`PortfolioSummary.nominalReturnRate` and `PortfolioSummary.realReturn` remain null. XIRR remains the sole activated public return metric. Planned snapshot serialization must expose `legacyReturnStatus: UNAVAILABLE` and null nominal/real rates until a separately approved methodology exists.

`PortfolioSummary.totalValue` is canonical cash plus local remaining acquisition-cost basis. `PortfolioReturnProjection.terminalPortfolioValue` is the exact-date market terminal value used for XIRR when complete user-supplied valuation coverage exists. Their divergence is intentional: with a deposit of 100, a BUY of 1 at 100, and a manual valuation of 125, summary total value is 100 while XIRR terminal value is 125.

## Rollout

Apply `000014` before the application revision. The migration is additive and contains no DML: old rows receive the constant `UNAVAILABLE` default and new writes are constrained to that status. The paired DOWN is permitted only in a disposable environment before accepted runtime reliance; production rollback leaves the additive marker in place and rolls the application forward.
