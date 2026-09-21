package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	replayBoundRawRows             = 5000
	replayPolicyVersion            = "oi-new-04-c-prime-v1"
	replayPositionMethodology      = "adr-009-wac-v1"
	replaySnapshotMethodology      = stage371SnapshotMethodology
	replaySourceStateDigestVersion = "oi-new-04-source-state-v1"
)

var (
	ErrRetroactiveReplayWindowExceeded = errors.New("retroactive replay window exceeded")
	ErrReplayStateStale                = errors.New("replay state stale")
	ErrReplayEpochMissing              = errors.New("replay epoch missing")
)

type replayPolicyState struct {
	Active                   bool
	BlockedAfterInvalidation bool
	PolicyVersion            string
	ActivationGeneration     int64
	ReplayBoundRawRows       int
	PositionMethodology      string
	SnapshotMethodology      string
}

type replayEpoch struct {
	ID                      string
	PortfolioID             string
	PolicyVersion           string
	ActivationGeneration    int64
	EpochGeneration         int64
	BoundaryTradeDate       *string
	BoundaryLogicalSequence int64
	BuildRawLedgerWatermark int64
	PositionMethodology     string
	SnapshotMethodology     string
	LatestPositionTradeDate *string
	SourceStateSHA256       string
	Positions               map[string]replayPosition
	Financial               replayFinancialState
}

type replayPosition struct {
	AssetID            string
	AssetType          string
	PositionGeneration int64
	State              position.State
}

type replayFinancialState struct {
	Amounts          cashFlowAmounts
	InvestedCapital  decimal.Decimal
	InputWatermarkAt *time.Time
	InputWatermark   string
}

type boundedLedgerRow struct {
	EntryID               string
	TransactionID         string
	AssetID               *string
	AssetType             *string
	TransactionType       string
	Quantity              *decimal.Decimal
	UnitPrice             *decimal.Decimal
	GrossAmount           decimal.Decimal
	Commission            decimal.Decimal
	Tax                   decimal.Decimal
	TradeDate             string
	LedgerSequence        int64
	Revision              int
	PriorEntryID          *string
	ReversesTransactionID *string
	CreatedAt             time.Time
	CreatedAtText         string
}

type canonicalReplayRow struct {
	boundedLedgerRow
	LogicalSequence int64
	ReversalDate    *string
}

type replayMutationImpact struct {
	Kind               string
	RawRows            int
	TradeDates         []string
	TransactionID      string
	CurrentTradeDate   string
	CorrectedTradeDate string
	ReversalEffective  string
}

type financialReplayPlan struct {
	Policy            replayPolicyState
	Epoch             replayEpoch
	ExistingRaw       []boundedLedgerRow
	ObservedWatermark int64
	Legacy            bool
}

type replayWorkingState struct {
	Positions               map[string]replayPosition
	Financial               replayFinancialState
	LatestPositionTradeDate *string
}

func currentReplayPolicyTx(ctx context.Context, tx *sql.Tx) (replayPolicyState, error) {
	var state replayPolicyState
	var invalidated bool
	err := tx.QueryRowContext(ctx, `
		WITH latest_activated AS (
			SELECT g.policy_version, g.activation_generation, g.replay_bound_raw_rows
			FROM analytics.replay_policy_generations g
			JOIN analytics.replay_policy_events activated
			  ON activated.policy_version = g.policy_version
			 AND activated.activation_generation = g.activation_generation
			 AND activated.event_type = 'ACTIVATED'
			WHERE g.policy_version = $1
			ORDER BY g.activation_generation DESC
			LIMIT 1
		)
		SELECT
			latest.policy_version,
			latest.activation_generation,
			latest.replay_bound_raw_rows,
			EXISTS (
				SELECT 1 FROM analytics.replay_policy_events invalidated
				WHERE invalidated.policy_version = latest.policy_version
				  AND invalidated.activation_generation = latest.activation_generation
				  AND invalidated.event_type = 'INVALIDATED'
			)
		FROM latest_activated latest
	`, replayPolicyVersion).Scan(
		&state.PolicyVersion,
		&state.ActivationGeneration,
		&state.ReplayBoundRawRows,
		&invalidated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return replayPolicyState{}, nil
	}
	if err != nil {
		return replayPolicyState{}, err
	}
	if state.PolicyVersion != replayPolicyVersion || state.ReplayBoundRawRows != replayBoundRawRows {
		return replayPolicyState{}, fmt.Errorf("%w: unsupported replay policy metadata", ErrReplayStateStale)
	}
	state.PositionMethodology = replayPositionMethodology
	state.SnapshotMethodology = replaySnapshotMethodology
	if invalidated {
		// Never fall back N-1 and never silently resume legacy under R1/R2. Emergency legacy
		// requires the approved drain/revoke to R0; N+1 becomes serving only after ACTIVATED(N+1).
		state.BlockedAfterInvalidation = true
		return state, nil
	}
	state.Active = true
	return state, nil
}

func (s *Store) prepareFinancialReplayPlanTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	impact replayMutationImpact,
	now time.Time,
) (financialReplayPlan, error) {
	if s == nil || !s.runtimeCapabilityProfile.valid() {
		return financialReplayPlan{}, fmt.Errorf("%w: invalid application capability profile", ErrReplayStateStale)
	}
	if err := s.assertReplayWriteWindowTx(ctx, tx); err != nil {
		return financialReplayPlan{}, err
	}
	// R0 intentionally has no replay-table SELECT capability. It is the pre-activation compatibility
	// profile and therefore stays on the canonical legacy path without consulting replay state.
	// Activation requires a full drain of R0/A0 before the owner appends ACTIVATED(N).
	if s.runtimeCapabilityProfile == RuntimeCapabilityR0 {
		return financialReplayPlan{Legacy: true}, nil
	}

	policy, err := currentReplayPolicyTx(ctx, tx)
	if err != nil {
		return financialReplayPlan{}, err
	}
	if policy.BlockedAfterInvalidation {
		return financialReplayPlan{}, ErrReplayStateStale
	}
	if !policy.Active {
		return financialReplayPlan{Legacy: true}, nil
	}
	if s.runtimeCapabilityProfile != RuntimeCapabilityR2 {
		return financialReplayPlan{}, fmt.Errorf("%w: active replay policy requires R2 application capability", ErrReplayStateStale)
	}

	epoch, err := loadLatestReplayEpochTx(ctx, tx, portfolioID, policy)
	if errors.Is(err, sql.ErrNoRows) {
		empty, emptyErr := portfolioLedgerEmptyTx(ctx, tx, portfolioID)
		if emptyErr != nil {
			return financialReplayPlan{}, emptyErr
		}
		if !empty {
			return financialReplayPlan{}, ErrReplayEpochMissing
		}
		epoch, err = appendGenesisReplayEpochTx(ctx, tx, portfolioID, policy, now)
	}
	if err != nil {
		return financialReplayPlan{}, err
	}

	staleRows, err := loadRawLedgerRangeTx(
		ctx, tx, portfolioID,
		epoch.BuildRawLedgerWatermark,
		0,
		replayBoundRawRows+1,
	)
	if err != nil {
		return financialReplayPlan{}, err
	}
	if len(staleRows) > replayBoundRawRows {
		return financialReplayPlan{}, ErrReplayStateStale
	}
	if err := detectStaleRowsAgainstEpoch(ctx, tx, epoch, staleRows); err != nil {
		return financialReplayPlan{}, err
	}

	mutableRows, err := loadRawLedgerRangeTx(
		ctx, tx, portfolioID,
		epoch.BoundaryLogicalSequence,
		epoch.BuildRawLedgerWatermark,
		replayBoundRawRows+1,
	)
	if err != nil {
		return financialReplayPlan{}, err
	}
	if len(mutableRows) > replayBoundRawRows {
		return financialReplayPlan{}, ErrReplayStateStale
	}
	if _, err := canonicalizeBoundedLedger(ctx, mutableRows); err != nil {
		return financialReplayPlan{}, fmt.Errorf("%w: %v", ErrReplayStateStale, err)
	}
	if err := verifyEpochSourceState(ctx, epoch, mutableRows); err != nil {
		return financialReplayPlan{}, err
	}

	existingRaw := make([]boundedLedgerRow, 0, len(mutableRows)+len(staleRows))
	existingRaw = append(existingRaw, mutableRows...)
	existingRaw = append(existingRaw, staleRows...)
	// A stale writer is never healed by scanning beyond B. If the state presented by the
	// latest persisted epoch already exceeds the hard suffix bound, fail closed.
	if len(existingRaw) > replayBoundRawRows {
		return financialReplayPlan{}, ErrReplayStateStale
	}
	if _, err := canonicalizeBoundedLedger(ctx, existingRaw); err != nil {
		return financialReplayPlan{}, fmt.Errorf("%w: %v", ErrReplayStateStale, err)
	}
	advancedEpoch, admittedRaw, err := compactReplayPlanForAdmission(ctx, epoch, existingRaw, impact.RawRows)
	if err != nil {
		return financialReplayPlan{}, err
	}
	if err := admitReplayMutationTx(ctx, tx, advancedEpoch, impact, len(admittedRaw)); err != nil {
		return financialReplayPlan{}, err
	}
	epoch = advancedEpoch
	existingRaw = admittedRaw
	observedWatermark := epoch.BuildRawLedgerWatermark
	if len(staleRows) > 0 {
		observedWatermark = staleRows[len(staleRows)-1].LedgerSequence
	}
	return financialReplayPlan{
		Policy: policy, Epoch: epoch, ExistingRaw: existingRaw, ObservedWatermark: observedWatermark,
	}, nil
}

func compactReplayPlanForAdmission(
	ctx context.Context,
	epoch replayEpoch,
	existingRaw []boundedLedgerRow,
	projectedNewRows int,
) (replayEpoch, []boundedLedgerRow, error) {
	if projectedNewRows < 0 || projectedNewRows > replayBoundRawRows {
		return replayEpoch{}, nil, ErrRetroactiveReplayWindowExceeded
	}
	if len(existingRaw) > replayBoundRawRows {
		return replayEpoch{}, nil, ErrReplayStateStale
	}
	if len(existingRaw)+projectedNewRows <= replayBoundRawRows {
		return epoch, append([]boundedLedgerRow(nil), existingRaw...), nil
	}

	keepRows := replayBoundRawRows - projectedNewRows
	prefix, mutable, boundarySequence, boundaryDate, err := partitionReplayBoundary(existingRaw, keepRows)
	if err != nil {
		return replayEpoch{}, nil, err
	}
	if len(mutable)+projectedNewRows > replayBoundRawRows {
		return replayEpoch{}, nil, ErrRetroactiveReplayWindowExceeded
	}
	if len(prefix) == 0 {
		return replayEpoch{}, nil, ErrRetroactiveReplayWindowExceeded
	}

	canonicalPrefix, err := canonicalizeBoundedLedger(ctx, prefix)
	if err != nil {
		return replayEpoch{}, nil, fmt.Errorf("%w: compact replay prefix: %v", ErrReplayStateStale, err)
	}
	if boundaryDate == nil {
		return replayEpoch{}, nil, ErrReplayStateStale
	}
	boundaryState, err := replayStateAtDate(ctx, epoch, canonicalPrefix, prefix, *boundaryDate)
	if err != nil {
		return replayEpoch{}, nil, err
	}

	advanced := epoch
	advanced.BoundaryLogicalSequence = boundarySequence
	advanced.BoundaryTradeDate = boundaryDate
	advanced.Positions = boundaryState.Positions
	advanced.Financial = boundaryState.Financial
	advanced.LatestPositionTradeDate = boundaryState.LatestPositionTradeDate
	return advanced, mutable, nil
}

func admitReplayMutationTx(
	ctx context.Context,
	tx *sql.Tx,
	epoch replayEpoch,
	impact replayMutationImpact,
	currentMutableRows int,
) error {
	if impact.RawRows < 0 || currentMutableRows+impact.RawRows > replayBoundRawRows {
		return ErrRetroactiveReplayWindowExceeded
	}
	for _, tradeDate := range impact.TradeDates {
		if crossesFrozenTradeBoundary(epoch, tradeDate) {
			return ErrRetroactiveReplayWindowExceeded
		}
	}

	switch impact.Kind {
	case "append", "import":
		return nil
	case "correct":
		frozen, err := logicalFamilyFrozenTx(ctx, tx, impact.TransactionID, epoch.BoundaryLogicalSequence)
		if err != nil {
			return err
		}
		if frozen || crossesFrozenTradeBoundary(epoch, impact.CurrentTradeDate) || crossesFrozenTradeBoundary(epoch, impact.CorrectedTradeDate) {
			return ErrRetroactiveReplayWindowExceeded
		}
		return nil
	case "reverse":
		frozen, err := logicalFamilyFrozenTx(ctx, tx, impact.TransactionID, epoch.BoundaryLogicalSequence)
		if err != nil {
			return err
		}
		if frozen || crossesFrozenTradeBoundary(epoch, impact.ReversalEffective) {
			return ErrRetroactiveReplayWindowExceeded
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown financial mutation kind %q", ErrReplayStateStale, impact.Kind)
	}
}

func crossesFrozenTradeBoundary(epoch replayEpoch, tradeDate string) bool {
	if epoch.BoundaryTradeDate == nil || strings.TrimSpace(tradeDate) == "" {
		return false
	}
	return tradeDate < *epoch.BoundaryTradeDate
}

func logicalFamilyFrozenTx(ctx context.Context, tx *sql.Tx, transactionID string, boundarySequence int64) (bool, error) {
	var firstSequence int64
	err := tx.QueryRowContext(ctx, `
		SELECT MIN(ledger_sequence)
		FROM investment.transaction_entries
		WHERE transaction_id = $1::uuid
		  AND reverses_transaction_id IS NULL
	`, transactionID).Scan(&firstSequence)
	if err != nil {
		return false, err
	}
	return firstSequence <= boundarySequence, nil
}

func detectStaleRowsAgainstEpoch(ctx context.Context, tx *sql.Tx, epoch replayEpoch, rows []boundedLedgerRow) error {
	for _, row := range rows {
		if crossesFrozenTradeBoundary(epoch, row.TradeDate) {
			return ErrReplayStateStale
		}
		if row.ReversesTransactionID != nil {
			frozen, err := logicalFamilyFrozenTx(ctx, tx, *row.ReversesTransactionID, epoch.BoundaryLogicalSequence)
			if err != nil {
				return err
			}
			if frozen {
				return ErrReplayStateStale
			}
			continue
		}
		if row.Revision > 1 {
			frozen, err := logicalFamilyFrozenTx(ctx, tx, row.TransactionID, epoch.BoundaryLogicalSequence)
			if err != nil {
				return err
			}
			if frozen {
				return ErrReplayStateStale
			}
		}
	}
	return nil
}

func loadLatestReplayEpochTx(ctx context.Context, tx *sql.Tx, portfolioID string, policy replayPolicyState) (replayEpoch, error) {
	var epoch replayEpoch
	var boundaryDate, latestPositionDate sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT
			e.epoch_id::text,
			e.portfolio_id::text,
			e.policy_version,
			e.activation_generation,
			e.epoch_generation,
			e.boundary_trade_date::text,
			e.boundary_logical_sequence,
			e.build_raw_ledger_watermark,
			e.position_methodology_version,
			e.snapshot_methodology_version,
			e.latest_position_trade_date::text,
			e.source_state_sha256
		FROM analytics.portfolio_replay_epochs e
		WHERE e.portfolio_id = $1::uuid
		  AND e.policy_version = $2
		  AND e.activation_generation = $3
		ORDER BY e.epoch_generation DESC
		LIMIT 1
	`, portfolioID, policy.PolicyVersion, policy.ActivationGeneration).Scan(
		&epoch.ID,
		&epoch.PortfolioID,
		&epoch.PolicyVersion,
		&epoch.ActivationGeneration,
		&epoch.EpochGeneration,
		&boundaryDate,
		&epoch.BoundaryLogicalSequence,
		&epoch.BuildRawLedgerWatermark,
		&epoch.PositionMethodology,
		&epoch.SnapshotMethodology,
		&latestPositionDate,
		&epoch.SourceStateSHA256,
	)
	if err != nil {
		return replayEpoch{}, err
	}
	if boundaryDate.Valid {
		value := boundaryDate.String
		epoch.BoundaryTradeDate = &value
	}
	if latestPositionDate.Valid {
		value := latestPositionDate.String
		epoch.LatestPositionTradeDate = &value
	}
	if epoch.PolicyVersion != policy.PolicyVersion ||
		epoch.ActivationGeneration != policy.ActivationGeneration ||
		epoch.PositionMethodology != replayPositionMethodology ||
		epoch.SnapshotMethodology != replaySnapshotMethodology ||
		epoch.BoundaryLogicalSequence > epoch.BuildRawLedgerWatermark ||
		epoch.EpochGeneration <= 0 || len(epoch.SourceStateSHA256) != 64 {
		return replayEpoch{}, ErrReplayStateStale
	}

	positions, err := loadReplayPositionsTx(ctx, tx, epoch.ID)
	if err != nil {
		return replayEpoch{}, err
	}
	financial, err := loadReplayFinancialStateTx(ctx, tx, epoch.ID)
	if err != nil {
		return replayEpoch{}, err
	}
	epoch.Positions = positions
	epoch.Financial = financial
	return epoch, nil
}

func loadReplayPositionsTx(ctx context.Context, tx *sql.Tx, epochID string) (map[string]replayPosition, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT asset_id::text, asset_type, quantity::text, weighted_average_cost_amount::text, position_generation
		FROM analytics.portfolio_replay_positions
		WHERE epoch_id = $1::uuid
		ORDER BY asset_id
	`, epochID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := map[string]replayPosition{}
	for rows.Next() {
		var assetID, assetType, quantityText, wacText string
		var generation int64
		if err := rows.Scan(&assetID, &assetType, &quantityText, &wacText, &generation); err != nil {
			return nil, err
		}
		quantity, err := decimal.FromString(quantityText)
		if err != nil {
			return nil, ErrReplayStateStale
		}
		wac, err := decimal.FromString(wacText)
		if err != nil {
			return nil, ErrReplayStateStale
		}
		state, err := position.Apply(position.Empty(), position.Trade{Type: "BUY", Quantity: quantity, UnitPrice: wac})
		if err != nil {
			return nil, ErrReplayStateStale
		}
		positions[assetID] = replayPosition{AssetID: assetID, AssetType: assetType, PositionGeneration: generation, State: state}
	}
	return positions, rows.Err()
}

func loadReplayFinancialStateTx(ctx context.Context, tx *sql.Tx, epochID string) (replayFinancialState, error) {
	var values [10]string
	var watermarkAt sql.NullTime
	var watermark string
	err := tx.QueryRowContext(ctx, `
		SELECT
			deposits_amount::text,
			withdrawals_amount::text,
			buy_outflows_amount::text,
			sell_inflows_amount::text,
			dividends_gross_amount::text,
			coupons_gross_amount::text,
			fees_amount::text,
			taxes_amount::text,
			net_investment_income_amount::text,
			invested_capital_amount::text,
			input_watermark_at,
			input_watermark
		FROM analytics.portfolio_replay_financial_state
		WHERE epoch_id = $1::uuid
	`, epochID).Scan(
		&values[0], &values[1], &values[2], &values[3], &values[4],
		&values[5], &values[6], &values[7], &values[8], &values[9], &watermarkAt, &watermark,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return replayFinancialState{}, ErrReplayStateStale
		}
		return replayFinancialState{}, err
	}
	parsed := make([]decimal.Decimal, len(values))
	for i, raw := range values {
		value, err := decimal.FromString(raw)
		if err != nil || !value.FitsStorage() {
			return replayFinancialState{}, ErrReplayStateStale
		}
		parsed[i] = value
	}
	state := replayFinancialState{
		Amounts: cashFlowAmounts{
			Deposits: parsed[0], Withdrawals: parsed[1], BuyOutflows: parsed[2], SellInflows: parsed[3],
			DividendsGross: parsed[4], CouponsGross: parsed[5], Fees: parsed[6], Taxes: parsed[7],
			NetInvestmentIncome: parsed[8],
		},
		InvestedCapital: parsed[9],
		InputWatermark:  watermark,
	}
	if watermarkAt.Valid {
		value := watermarkAt.Time
		state.InputWatermarkAt = &value
	}
	if err := state.Amounts.validateStorage(); err != nil || !state.InvestedCapital.FitsStorage() || state.InputWatermark == "" {
		return replayFinancialState{}, ErrReplayStateStale
	}
	if state.InputWatermark == "empty" && state.InputWatermarkAt != nil {
		return replayFinancialState{}, ErrReplayStateStale
	}
	if state.InputWatermark != "empty" && state.InputWatermarkAt == nil {
		return replayFinancialState{}, ErrReplayStateStale
	}
	return state, nil
}

func portfolioLedgerEmptyTx(ctx context.Context, tx *sql.Tx, portfolioID string) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM investment.transaction_entries
			WHERE portfolio_id = $1::uuid
			LIMIT 1
		)
	`, portfolioID).Scan(&exists); err != nil {
		return false, err
	}
	return !exists, nil
}

func loadRawLedgerRangeTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	afterSequence int64,
	throughSequence int64,
	limit int,
) ([]boundedLedgerRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			te.entry_id::text,
			te.transaction_id::text,
			te.asset_id::text,
			a.asset_type,
			te.transaction_type,
			te.quantity::text,
			te.unit_price_amount::text,
			te.gross_amount::text,
			te.commission_amount::text,
			te.tax_amount::text,
			te.trade_date::text,
			te.ledger_sequence,
			te.revision,
			te.prior_entry_id::text,
			te.reverses_transaction_id::text,
			te.created_at,
			te.created_at::text
		FROM investment.transaction_entries te
		LEFT JOIN investment.assets a ON a.id = te.asset_id
		WHERE te.portfolio_id = $1::uuid
		  AND te.ledger_sequence > $2
		  AND ($3::bigint = 0 OR te.ledger_sequence <= $3)
		ORDER BY te.ledger_sequence ASC
		LIMIT $4
	`, portfolioID, afterSequence, throughSequence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]boundedLedgerRow, 0)
	for rows.Next() {
		var row boundedLedgerRow
		var assetID, assetType, quantity, unitPrice, priorID, reversesID sql.NullString
		var gross, commission, tax string
		if err := rows.Scan(
			&row.EntryID, &row.TransactionID, &assetID, &assetType, &row.TransactionType,
			&quantity, &unitPrice, &gross, &commission, &tax,
			&row.TradeDate, &row.LedgerSequence, &row.Revision, &priorID, &reversesID,
			&row.CreatedAt, &row.CreatedAtText,
		); err != nil {
			return nil, err
		}
		if assetID.Valid {
			value := assetID.String
			row.AssetID = &value
		}
		if assetType.Valid {
			value := assetType.String
			row.AssetType = &value
		}
		if quantity.Valid {
			value, err := decimal.FromString(quantity.String)
			if err != nil {
				return nil, ErrReplayStateStale
			}
			row.Quantity = &value
		}
		if unitPrice.Valid {
			value, err := decimal.FromString(unitPrice.String)
			if err != nil {
				return nil, ErrReplayStateStale
			}
			row.UnitPrice = &value
		}
		var err error
		if row.GrossAmount, err = decimal.FromString(gross); err != nil {
			return nil, ErrReplayStateStale
		}
		if row.Commission, err = decimal.FromString(commission); err != nil {
			return nil, ErrReplayStateStale
		}
		if row.Tax, err = decimal.FromString(tax); err != nil {
			return nil, ErrReplayStateStale
		}
		if priorID.Valid {
			value := priorID.String
			row.PriorEntryID = &value
		}
		if reversesID.Valid {
			value := reversesID.String
			row.ReversesTransactionID = &value
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	countReplayHistoryRowsLoaded(ctx, len(result))
	return result, nil
}

func canonicalizeBoundedLedger(ctx context.Context, raw []boundedLedgerRow) ([]canonicalReplayRow, error) {
	countReplayLedgerShapeValidation(ctx)
	countReplayHistoryRowsExamined(ctx, len(raw))
	type family struct {
		rows []boundedLedgerRow
	}
	families := map[string]*family{}
	reversals := map[string]string{}
	for _, row := range raw {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if row.LedgerSequence <= 0 {
			return nil, ErrLedgerSequenceUnavailable
		}
		if row.ReversesTransactionID != nil {
			if row.Revision != 1 || row.PriorEntryID != nil {
				return nil, ErrUnsupportedPositionLedger
			}
			if _, duplicate := reversals[*row.ReversesTransactionID]; duplicate {
				return nil, ErrUnsupportedPositionLedger
			}
			reversals[*row.ReversesTransactionID] = row.TradeDate
			continue
		}
		f := families[row.TransactionID]
		if f == nil {
			f = &family{}
			families[row.TransactionID] = f
		}
		f.rows = append(f.rows, row)
	}

	result := make([]canonicalReplayRow, 0, len(families))
	for transactionID, family := range families {
		sort.Slice(family.rows, func(i, j int) bool { return family.rows[i].Revision < family.rows[j].Revision })
		for i, row := range family.rows {
			expectedRevision := i + 1
			if row.Revision != expectedRevision {
				return nil, ErrUnsupportedPositionLedger
			}
			if i == 0 {
				if row.PriorEntryID != nil {
					return nil, ErrUnsupportedPositionLedger
				}
			} else if row.PriorEntryID == nil || *row.PriorEntryID != family.rows[i-1].EntryID {
				return nil, ErrUnsupportedPositionLedger
			}
		}
		latest := family.rows[len(family.rows)-1]
		logicalSequence := family.rows[0].LedgerSequence
		canonical := canonicalReplayRow{boundedLedgerRow: latest, LogicalSequence: logicalSequence}
		if reversalDate, ok := reversals[transactionID]; ok {
			value := reversalDate
			canonical.ReversalDate = &value
			delete(reversals, transactionID)
		}
		result = append(result, canonical)
	}
	if len(reversals) != 0 {
		// A reversal that targets a family outside the mutable suffix would make the frozen epoch invalid.
		return nil, ErrReplayStateStale
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TradeDate != result[j].TradeDate {
			return result[i].TradeDate < result[j].TradeDate
		}
		return result[i].LogicalSequence < result[j].LogicalSequence
	})
	return result, nil
}

func cloneReplayWorkingState(epoch replayEpoch) replayWorkingState {
	positions := make(map[string]replayPosition, len(epoch.Positions))
	for key, value := range epoch.Positions {
		positions[key] = value
	}
	return replayWorkingState{
		Positions:               positions,
		Financial:               epoch.Financial,
		LatestPositionTradeDate: copyStringPointer(epoch.LatestPositionTradeDate),
	}
}

func replayStateAtDate(
	ctx context.Context,
	epoch replayEpoch,
	rows []canonicalReplayRow,
	rawRows []boundedLedgerRow,
	snapshotDate string,
) (replayWorkingState, error) {
	state := cloneReplayWorkingState(epoch)
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return replayWorkingState{}, err
		}
		if row.TradeDate > snapshotDate {
			break
		}
		if row.ReversalDate != nil && *row.ReversalDate <= snapshotDate {
			continue
		}
		if err := applyCanonicalReplayRow(ctx, &state, row); err != nil {
			return replayWorkingState{}, err
		}
	}
	state.Financial.InputWatermark, state.Financial.InputWatermarkAt = replayInputWatermarkRaw(epoch.Financial.InputWatermark, epoch.Financial.InputWatermarkAt, rawRows, snapshotDate)
	return state, nil
}

// replayStatesForDates derives an ordered snapshot timeline without replaying the bounded suffix
// from the beginning for every snapshot date. Ordinary intervals are advanced in one forward pass.
// A reversal can invalidate non-invertible WAC state, so crossing a distinct reversal effective date
// causes one exact bounded in-memory restart from the immutable epoch boundary. The database suffix
// is still loaded/validated once, no pre-epoch history is read, and snapshot count D does not multiply
// bounded-history replay work. This intentionally trades a bounded R*B in-memory worst case for exact
// WAC semantics; it never regresses to R*H or D*H.
func replayStatesForDates(
	ctx context.Context,
	epoch replayEpoch,
	rows []canonicalReplayRow,
	rawRows []boundedLedgerRow,
	dates []string,
) ([]replayWorkingState, error) {
	if len(dates) == 0 {
		return []replayWorkingState{}, nil
	}
	orderedDates := append([]string(nil), dates...)
	for i := 1; i < len(orderedDates); i++ {
		if orderedDates[i] <= orderedDates[i-1] {
			return nil, fmt.Errorf("%w: snapshot dates must be strictly increasing", ErrReplayStateStale)
		}
	}

	reversalDates := make([]string, 0)
	seenReversalDate := map[string]struct{}{}
	for _, row := range rows {
		if row.ReversalDate == nil {
			continue
		}
		if _, seen := seenReversalDate[*row.ReversalDate]; seen {
			continue
		}
		seenReversalDate[*row.ReversalDate] = struct{}{}
		reversalDates = append(reversalDates, *row.ReversalDate)
	}
	sort.Strings(reversalDates)

	// Watermark visibility advances once across raw immutable rows ordered by effective date.
	// This avoids a hidden D*B full-prefix scan merely to reproduce MAX(created_at).
	watermarkRows := append([]boundedLedgerRow(nil), rawRows...)
	sort.SliceStable(watermarkRows, func(i, j int) bool {
		if watermarkRows[i].TradeDate != watermarkRows[j].TradeDate {
			return watermarkRows[i].TradeDate < watermarkRows[j].TradeDate
		}
		return watermarkRows[i].LedgerSequence < watermarkRows[j].LedgerSequence
	})
	watermark := epoch.Financial.InputWatermark
	var watermarkAt *time.Time
	if epoch.Financial.InputWatermarkAt != nil {
		value := *epoch.Financial.InputWatermarkAt
		watermarkAt = &value
	}
	watermarkIndex := 0

	state := cloneReplayWorkingState(epoch)
	currentDate := ""
	if epoch.BoundaryTradeDate != nil {
		currentDate = *epoch.BoundaryTradeDate
	}
	// rows is already the raw suffix strictly after boundary_logical_sequence, so same-day
	// suffix rows must start at index zero rather than being skipped by date alone.
	rowIndex := 0
	reversalIndex := firstStringAfter(reversalDates, currentDate)
	states := make([]replayWorkingState, 0, len(orderedDates))

	for _, snapshotDate := range orderedDates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := ctxDateOrder(epoch, currentDate, snapshotDate); err != nil {
			return nil, err
		}

		for watermarkIndex < len(watermarkRows) && watermarkRows[watermarkIndex].TradeDate <= snapshotDate {
			row := watermarkRows[watermarkIndex]
			if watermarkAt == nil || row.CreatedAt.After(*watermarkAt) {
				value := row.CreatedAt
				watermarkAt = &value
				watermark = row.CreatedAtText
			}
			watermarkIndex++
		}
		if watermark == "" {
			watermark = "empty"
		}

		crossedReversal := false
		for reversalIndex < len(reversalDates) && reversalDates[reversalIndex] <= snapshotDate {
			if reversalDates[reversalIndex] > currentDate {
				crossedReversal = true
			}
			reversalIndex++
		}
		if crossedReversal {
			var err error
			state, err = replayStateAtDate(ctx, epoch, rows, rawRows, snapshotDate)
			if err != nil {
				return nil, err
			}
			rowIndex = firstCanonicalRowAfter(rows, snapshotDate)
		} else {
			for rowIndex < len(rows) && rows[rowIndex].TradeDate <= snapshotDate {
				row := rows[rowIndex]
				if row.ReversalDate == nil || *row.ReversalDate > snapshotDate {
					if err := applyCanonicalReplayRow(ctx, &state, row); err != nil {
						return nil, err
					}
				}
				rowIndex++
			}
		}
		state.Financial.InputWatermark = watermark
		if watermarkAt == nil {
			state.Financial.InputWatermarkAt = nil
		} else {
			value := *watermarkAt
			state.Financial.InputWatermarkAt = &value
		}
		states = append(states, cloneReplayState(state))
		currentDate = snapshotDate
	}
	return states, nil
}

func ctxDateOrder(epoch replayEpoch, currentDate, snapshotDate string) error {
	if snapshotDate == "" {
		return fmt.Errorf("%w: empty snapshot date", ErrReplayStateStale)
	}
	if epoch.BoundaryTradeDate != nil && snapshotDate < *epoch.BoundaryTradeDate {
		return ErrRetroactiveReplayWindowExceeded
	}
	if currentDate != "" && snapshotDate < currentDate {
		return fmt.Errorf("%w: non-monotonic snapshot timeline", ErrReplayStateStale)
	}
	return nil
}

func firstCanonicalRowAfter(rows []canonicalReplayRow, date string) int {
	if date == "" {
		return 0
	}
	return sort.Search(len(rows), func(i int) bool { return rows[i].TradeDate > date })
}

func firstStringAfter(values []string, value string) int {
	if value == "" {
		return 0
	}
	return sort.Search(len(values), func(i int) bool { return values[i] > value })
}

func cloneReplayState(state replayWorkingState) replayWorkingState {
	positions := make(map[string]replayPosition, len(state.Positions))
	for key, value := range state.Positions {
		positions[key] = value
	}
	return replayWorkingState{
		Positions:               positions,
		Financial:               state.Financial,
		LatestPositionTradeDate: copyStringPointer(state.LatestPositionTradeDate),
	}
}

func applyCanonicalReplayRow(ctx context.Context, state *replayWorkingState, row canonicalReplayRow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	countReplayLedgerRowApplied(ctx)
	if err := state.Financial.Amounts.apply(row.effectiveLedgerRow()); err != nil {
		return err
	}
	if row.TransactionType == "BUY" {
		outflow := row.GrossAmount.Add(row.Commission).Add(row.Tax)
		state.Financial.InvestedCapital = state.Financial.InvestedCapital.Add(outflow)
		if !state.Financial.InvestedCapital.FitsStorage() {
			return fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8)", verticalslice.ErrInvalidInput)
		}
	}
	if row.TransactionType != "BUY" && row.TransactionType != "SELL" {
		return nil
	}
	if row.AssetID == nil || row.AssetType == nil || row.Quantity == nil || row.UnitPrice == nil {
		return ErrUnsupportedPositionLedger
	}

	current, exists := state.Positions[*row.AssetID]
	positionState := position.Empty()
	positionGeneration := int64(0)
	if exists {
		positionState = current.State
		positionGeneration = current.PositionGeneration
	}
	wasOpen := positionState.Open
	next, err := position.Apply(positionState, position.Trade{Type: row.TransactionType, Quantity: *row.Quantity, UnitPrice: *row.UnitPrice})
	switch {
	case errors.Is(err, position.ErrInsufficientQuantity):
		return verticalslice.ErrInsufficientPositionQuantity
	case errors.Is(err, position.ErrDerivedOverflow), errors.Is(err, position.ErrInvalidTrade):
		return fmt.Errorf("%w: position rebuild exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
	case err != nil:
		return err
	}
	if row.TransactionType == "BUY" && !wasOpen {
		positionGeneration = row.LogicalSequence
	}
	if !next.Open {
		delete(state.Positions, *row.AssetID)
	} else {
		state.Positions[*row.AssetID] = replayPosition{
			AssetID: *row.AssetID, AssetType: *row.AssetType, PositionGeneration: positionGeneration, State: next,
		}
	}
	if state.LatestPositionTradeDate == nil || row.TradeDate > *state.LatestPositionTradeDate {
		value := row.TradeDate
		state.LatestPositionTradeDate = &value
	}
	return nil
}

func (row canonicalReplayRow) effectiveLedgerRow() effectiveLedgerRow {
	return effectiveLedgerRow{
		EntryID: row.EntryID, TransactionID: row.TransactionID, AssetID: row.AssetID, AssetType: row.AssetType,
		TransactionType: row.TransactionType, Quantity: row.Quantity, UnitPrice: row.UnitPrice,
		GrossAmount: row.GrossAmount, Commission: row.Commission, Tax: row.Tax,
		TradeDate: row.TradeDate, LedgerSequence: row.LogicalSequence, Revision: row.Revision, UpdatedAt: row.CreatedAt,
	}
}

func replayInputWatermarkRaw(base string, baseAt *time.Time, rows []boundedLedgerRow, snapshotDate string) (string, *time.Time) {
	watermark := base
	var latest time.Time
	if baseAt != nil {
		latest = *baseAt
	}
	for _, row := range rows {
		// Legacy snapshot input_watermark is MAX(created_at) over raw immutable ledger rows
		// whose trade/effective date is visible at the snapshot date. It intentionally includes
		// superseded revisions and reversal rows, not only canonical effective transactions.
		if row.TradeDate > snapshotDate {
			continue
		}
		if latest.IsZero() || row.CreatedAt.After(latest) {
			latest = row.CreatedAt
			watermark = row.CreatedAtText
		}
	}
	if watermark == "" {
		watermark = "empty"
	}
	if latest.IsZero() {
		return watermark, nil
	}
	value := latest
	return watermark, &value
}

func replayAcquisitionTotals(state replayWorkingState) (acquisitionValueTotals, error) {
	totals := acquisitionValueTotals{Stock: decimal.Zero(), Bond: decimal.Zero()}
	for _, item := range state.Positions {
		if !item.State.Open {
			continue
		}
		switch item.AssetType {
		case "stock":
			totals.Stock = totals.Stock.Add(item.State.AcquisitionBasis)
		case "bond":
			totals.Bond = totals.Bond.Add(item.State.AcquisitionBasis)
		default:
			return acquisitionValueTotals{}, ErrReplayStateStale
		}
	}
	if !totals.Stock.FitsStorage() || !totals.Bond.FitsStorage() {
		return acquisitionValueTotals{}, verticalslice.ErrInvalidInput
	}
	return totals, nil
}

func (s *Store) executeFinancialReplayPlanTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	plan financialReplayPlan,
	affectedDates []string,
	now time.Time,
) error {
	if plan.Legacy {
		countReplayLegacyEngineExecution(ctx)
		return rebuildSnapshotPlan(ctx, tx, portfolioID, affectedDates, now)
	}
	countReplayBoundedEngineExecution(ctx)
	if s.runtimeCapabilityProfile != RuntimeCapabilityR2 {
		return ErrReplayStateStale
	}
	// Snapshot-date work has its own explicit defensive ceiling. It intentionally matches the
	// approved C' mutable-history bound; persistence is independently set-based below, so D does
	// not become one PostgreSQL round trip per date.
	if len(affectedDates) > replaySnapshotDateBound {
		return ErrRetroactiveReplayWindowExceeded
	}

	newRows, err := loadRawLedgerRangeTx(
		ctx, tx, portfolioID,
		plan.ObservedWatermark,
		0,
		replayBoundRawRows-len(plan.ExistingRaw)+1,
	)
	if err != nil {
		return err
	}
	if len(plan.ExistingRaw)+len(newRows) > replayBoundRawRows {
		return ErrRetroactiveReplayWindowExceeded
	}
	combined := make([]boundedLedgerRow, 0, len(plan.ExistingRaw)+len(newRows))
	combined = append(combined, plan.ExistingRaw...)
	combined = append(combined, newRows...)
	canonicalRows, err := canonicalizeBoundedLedger(ctx, combined)
	if err != nil {
		return err
	}

	states, err := replayStatesForDates(ctx, plan.Epoch, canonicalRows, combined, affectedDates)
	if err != nil {
		return err
	}
	if err := insertReplaySnapshotsBatchTx(ctx, tx, portfolioID, affectedDates, now, states); err != nil {
		return err
	}

	maxDate := maxReplayTradeDate(canonicalRows, plan.Epoch.BoundaryTradeDate)
	finalState := cloneReplayWorkingState(plan.Epoch)
	if maxDate != "" {
		finalState, err = replayStateAtDate(ctx, plan.Epoch, canonicalRows, combined, maxDate)
		if err != nil {
			return err
		}
	}
	newWatermark, err := latestLedgerSequenceTx(ctx, tx, portfolioID)
	if err != nil {
		return err
	}
	return appendRuntimeReplayEpochTx(ctx, tx, plan.Epoch, newWatermark, finalState, now)
}

func maxReplayTradeDate(rows []canonicalReplayRow, boundary *string) string {
	maxDate := ""
	if boundary != nil {
		maxDate = *boundary
	}
	for _, row := range rows {
		if row.TradeDate > maxDate {
			maxDate = row.TradeDate
		}
		if row.ReversalDate != nil && *row.ReversalDate > maxDate {
			maxDate = *row.ReversalDate
		}
	}
	return maxDate
}

func latestLedgerSequenceTx(ctx context.Context, tx *sql.Tx, portfolioID string) (int64, error) {
	var sequence int64
	err := tx.QueryRowContext(ctx, `
		SELECT ledger_sequence
		FROM investment.transaction_entries
		WHERE portfolio_id = $1::uuid
		ORDER BY ledger_sequence DESC
		LIMIT 1
	`, portfolioID).Scan(&sequence)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return sequence, err
}

func nextReplayEpochGenerationTx(ctx context.Context, tx *sql.Tx, portfolioID string) (int64, error) {
	var next int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(epoch_generation), 0) + 1
		FROM analytics.portfolio_replay_epochs
		WHERE portfolio_id = $1::uuid
	`, portfolioID).Scan(&next); err != nil {
		return 0, err
	}
	if next <= 0 {
		return 0, ErrReplayStateStale
	}
	return next, nil
}

func appendGenesisReplayEpochTx(ctx context.Context, tx *sql.Tx, portfolioID string, policy replayPolicyState, now time.Time) (replayEpoch, error) {
	epochGeneration, err := nextReplayEpochGenerationTx(ctx, tx, portfolioID)
	if err != nil {
		return replayEpoch{}, err
	}
	epoch := replayEpoch{
		ID: uuid.NewString(), PortfolioID: portfolioID,
		PolicyVersion: policy.PolicyVersion, ActivationGeneration: policy.ActivationGeneration,
		EpochGeneration: epochGeneration, BoundaryLogicalSequence: 0, BuildRawLedgerWatermark: 0,
		PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
		Positions: map[string]replayPosition{},
		Financial: replayFinancialState{Amounts: zeroCashFlowAmounts(), InvestedCapital: decimal.Zero(), InputWatermark: "empty"},
	}
	state := cloneReplayWorkingState(epoch)
	epoch.SourceStateSHA256 = replaySourceStateDigest(epoch, state)
	if err := insertReplayEpochTx(ctx, tx, epoch, now); err != nil {
		return replayEpoch{}, err
	}
	return epoch, nil
}

func appendRuntimeReplayEpochTx(ctx context.Context, tx *sql.Tx, previous replayEpoch, buildWatermark int64, finalState replayWorkingState, now time.Time) error {
	epoch := previous
	epoch.ID = uuid.NewString()
	epoch.EpochGeneration++
	epoch.BuildRawLedgerWatermark = buildWatermark
	epoch.SourceStateSHA256 = replaySourceStateDigest(epoch, finalState)
	return insertReplayEpochTx(ctx, tx, epoch, now)
}

func insertReplayEpochTx(ctx context.Context, tx *sql.Tx, epoch replayEpoch, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.portfolio_replay_epochs (
			epoch_id, portfolio_id, policy_version, activation_generation,
			epoch_generation, boundary_trade_date, boundary_logical_sequence,
			build_raw_ledger_watermark, position_methodology_version, snapshot_methodology_version,
			latest_position_trade_date, source_state_sha256, created_at
		) VALUES (
			$1::uuid, $2::uuid, $3, $4, $5,
			NULLIF($6, '')::date, $7, $8, $9, $10, NULLIF($11, '')::date, $12, $13
		)
	`, epoch.ID, epoch.PortfolioID, epoch.PolicyVersion, epoch.ActivationGeneration,
		epoch.EpochGeneration, pointerString(epoch.BoundaryTradeDate), epoch.BoundaryLogicalSequence,
		epoch.BuildRawLedgerWatermark, epoch.PositionMethodology, epoch.SnapshotMethodology,
		pointerString(epoch.LatestPositionTradeDate), epoch.SourceStateSHA256, now)
	if err != nil {
		return err
	}

	assetIDs := make([]string, 0, len(epoch.Positions))
	for assetID := range epoch.Positions {
		assetIDs = append(assetIDs, assetID)
	}
	sort.Strings(assetIDs)
	for _, assetID := range assetIDs {
		item := epoch.Positions[assetID]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO analytics.portfolio_replay_positions (
				epoch_id, asset_id, asset_type, quantity, weighted_average_cost_amount, position_generation
			) VALUES ($1::uuid, $2::uuid, $3, $4::numeric, $5::numeric, $6)
		`, epoch.ID, item.AssetID, item.AssetType, item.State.Quantity.String(), item.State.WeightedAverageCost.String(), item.PositionGeneration); err != nil {
			return err
		}
	}
	amounts := epoch.Financial.Amounts
	_, err = tx.ExecContext(ctx, `
		INSERT INTO analytics.portfolio_replay_financial_state (
			epoch_id, deposits_amount, withdrawals_amount, buy_outflows_amount, sell_inflows_amount,
			dividends_gross_amount, coupons_gross_amount, fees_amount, taxes_amount,
			net_investment_income_amount, invested_capital_amount, input_watermark_at, input_watermark
		) VALUES (
			$1::uuid, $2::numeric, $3::numeric, $4::numeric, $5::numeric,
			$6::numeric, $7::numeric, $8::numeric, $9::numeric, $10::numeric, $11::numeric, $12, $13
		)
	`, epoch.ID,
		amounts.Deposits.String(), amounts.Withdrawals.String(), amounts.BuyOutflows.String(), amounts.SellInflows.String(),
		amounts.DividendsGross.String(), amounts.CouponsGross.String(), amounts.Fees.String(), amounts.Taxes.String(),
		amounts.NetInvestmentIncome.String(), epoch.Financial.InvestedCapital.String(), epoch.Financial.InputWatermarkAt, epoch.Financial.InputWatermark)
	if err != nil {
		return err
	}
	countReplayEpochWrite(ctx)
	if err := maybeReplayTestFaultAfterEpochWrite(ctx); err != nil {
		return err
	}
	return nil
}

func verifyEpochSourceState(ctx context.Context, epoch replayEpoch, mutableRaw []boundedLedgerRow) error {
	canonicalRows, err := canonicalizeBoundedLedger(ctx, mutableRaw)
	if err != nil {
		return ErrReplayStateStale
	}
	maxDate := maxReplayTradeDate(canonicalRows, epoch.BoundaryTradeDate)
	state := cloneReplayWorkingState(epoch)
	if maxDate != "" {
		state, err = replayStateAtDate(ctx, epoch, canonicalRows, mutableRaw, maxDate)
		if err != nil {
			return ErrReplayStateStale
		}
	}
	if replaySourceStateDigest(epoch, state) != epoch.SourceStateSHA256 {
		return ErrReplayStateStale
	}
	return nil
}

func replaySourceStateDigest(epoch replayEpoch, state replayWorkingState) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s\npolicy=%s\nactivation_generation=%d\nboundary_date=%s\nboundary_sequence=%d\nbuild_watermark=%d\nposition_methodology=%s\nsnapshot_methodology=%s\n",
		replaySourceStateDigestVersion, epoch.PolicyVersion, epoch.ActivationGeneration, pointerString(epoch.BoundaryTradeDate),
		epoch.BoundaryLogicalSequence, epoch.BuildRawLedgerWatermark, epoch.PositionMethodology, epoch.SnapshotMethodology)
	amounts := state.Financial.Amounts
	// The legacy/public input_watermark string is intentionally excluded from replay identity: it is
	// populated from PostgreSQL's session rendering of TIMESTAMPTZ for compatibility with the frozen
	// snapshot contract. Canonical replay identity instead hashes only the scanned instant normalized
	// to UTC RFC3339Nano, making the same database state independent of session TimeZone.
	fmt.Fprintf(&builder, "deposits=%s\nwithdrawals=%s\nbuy_outflows=%s\nsell_inflows=%s\ndividends=%s\ncoupons=%s\nfees=%s\ntaxes=%s\nnet_investment_income=%s\ninvested_capital=%s\ncanonical_input_watermark_at=%s\nlatest_position_trade_date=%s\n",
		amounts.Deposits.String(), amounts.Withdrawals.String(), amounts.BuyOutflows.String(), amounts.SellInflows.String(),
		amounts.DividendsGross.String(), amounts.CouponsGross.String(), amounts.Fees.String(), amounts.Taxes.String(),
		amounts.NetInvestmentIncome.String(), state.Financial.InvestedCapital.String(),
		pointerTimeString(state.Financial.InputWatermarkAt), pointerString(state.LatestPositionTradeDate))
	assetIDs := make([]string, 0, len(state.Positions))
	for assetID := range state.Positions {
		assetIDs = append(assetIDs, assetID)
	}
	sort.Strings(assetIDs)
	for _, assetID := range assetIDs {
		item := state.Positions[assetID]
		fmt.Fprintf(&builder, "position=%s|%s|%s|%s|%d\n", item.AssetID, item.AssetType,
			item.State.Quantity.String(), item.State.WeightedAverageCost.String(), item.PositionGeneration)
	}
	digest := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(digest[:])
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func pointerTimeString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func (s *Store) appendGenesisIfActiveTx(ctx context.Context, tx *sql.Tx, portfolioID string, now time.Time) error {
	if s == nil || !s.runtimeCapabilityProfile.valid() {
		return ErrReplayStateStale
	}
	if err := s.assertReplayWriteWindowTx(ctx, tx); err != nil {
		return err
	}
	if s.runtimeCapabilityProfile == RuntimeCapabilityR0 {
		return nil
	}
	policy, err := currentReplayPolicyTx(ctx, tx)
	if err != nil {
		return err
	}
	if policy.BlockedAfterInvalidation {
		return ErrReplayStateStale
	}
	if !policy.Active {
		return nil
	}
	if s.runtimeCapabilityProfile != RuntimeCapabilityR2 {
		return ErrReplayStateStale
	}
	_, err = loadLatestReplayEpochTx(ctx, tx, portfolioID, policy)
	if err == nil {
		return ErrReplayStateStale
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	empty, err := portfolioLedgerEmptyTx(ctx, tx, portfolioID)
	if err != nil {
		return err
	}
	if !empty {
		return ErrReplayEpochMissing
	}
	_, err = appendGenesisReplayEpochTx(ctx, tx, portfolioID, policy, now)
	return err
}
