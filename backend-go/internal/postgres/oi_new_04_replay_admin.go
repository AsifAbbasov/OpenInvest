package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type ReplayGeneration struct {
	PolicyVersion        string
	ActivationGeneration int64
}

type ReplayBackfillResult struct {
	PolicyVersion        string
	ActivationGeneration int64
	PortfolioIDs         []string
	EpochIDs             []string
}

type ReplayActivationManifest struct {
	CanonicalBytes []byte
	SHA256         string
	PortfolioCount int
}

func (s *Store) AllocateReplayGeneration(ctx context.Context, now time.Time) (ReplayGeneration, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ReplayGeneration{}, err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return ReplayGeneration{}, err
	}

	var next int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(activation_generation), 0) + 1
		FROM analytics.replay_policy_generations
		WHERE policy_version = $1
	`, replayPolicyVersion).Scan(&next); err != nil {
		return ReplayGeneration{}, err
	}
	generation := ReplayGeneration{PolicyVersion: replayPolicyVersion, ActivationGeneration: next}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.replay_policy_generations (
			policy_version, activation_generation, replay_bound_raw_rows, created_at
		) VALUES ($1, $2, $3, $4)
	`, generation.PolicyVersion, generation.ActivationGeneration, replayBoundRawRows, now); err != nil {
		return ReplayGeneration{}, err
	}
	if err := tx.Commit(); err != nil {
		return ReplayGeneration{}, err
	}
	return generation, nil
}

func (s *Store) ProvisionalReplayBackfill(
	ctx context.Context,
	activationGeneration int64,
	targetMutableRows int,
	now time.Time,
) (ReplayBackfillResult, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ReplayBackfillResult{}, err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return ReplayBackfillResult{}, err
	}
	generation, err := replayGenerationTx(ctx, tx, activationGeneration)
	if err != nil {
		return ReplayBackfillResult{}, err
	}
	if err := assertReplayGenerationPreactiveTx(ctx, tx, generation); err != nil {
		return ReplayBackfillResult{}, err
	}
	result, err := replayBackfillTx(ctx, tx, generation, targetMutableRows, now)
	if err != nil {
		return ReplayBackfillResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ReplayBackfillResult{}, err
	}
	return result, nil
}

func (s *Store) FinalizeReplayGeneration(
	ctx context.Context,
	activationGeneration int64,
	targetMutableRows int,
	now time.Time,
) (ReplayBackfillResult, ReplayActivationManifest, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	if err := lockReplayFinalizationTx(ctx, tx); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE investment.portfolios IN SHARE MODE`); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE investment.transaction_entries IN SHARE MODE`); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	generation, err := replayGenerationTx(ctx, tx, activationGeneration)
	if err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	if err := assertReplayGenerationPreactiveTx(ctx, tx, generation); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	result, err := replayBackfillTx(ctx, tx, generation, targetMutableRows, now)
	if err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	manifest, err := buildReplayActivationManifestTx(ctx, tx, generation)
	if err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	if err := tx.Commit(); err != nil {
		return ReplayBackfillResult{}, ReplayActivationManifest{}, err
	}
	return result, manifest, nil
}

func replayBackfillTx(
	ctx context.Context,
	tx *sql.Tx,
	generation replayPolicyState,
	targetMutableRows int,
	now time.Time,
) (ReplayBackfillResult, error) {
	if targetMutableRows < 0 || targetMutableRows > replayBoundRawRows {
		return ReplayBackfillResult{}, fmt.Errorf("target mutable rows must be between 0 and %d", replayBoundRawRows)
	}
	portfolioIDs, err := activePortfolioIDsTx(ctx, tx)
	if err != nil {
		return ReplayBackfillResult{}, err
	}
	result := ReplayBackfillResult{
		PolicyVersion:        generation.PolicyVersion,
		ActivationGeneration: generation.ActivationGeneration,
		PortfolioIDs:         append([]string(nil), portfolioIDs...),
	}
	for _, portfolioID := range portfolioIDs {
		if err := ctx.Err(); err != nil {
			return ReplayBackfillResult{}, err
		}
		epoch, err := buildAdminReplayEpochTx(ctx, tx, generation, portfolioID, targetMutableRows, now)
		if err != nil {
			return ReplayBackfillResult{}, fmt.Errorf("portfolio %s: %w", portfolioID, err)
		}
		if err := insertReplayEpochTx(ctx, tx, epoch, now); err != nil {
			return ReplayBackfillResult{}, fmt.Errorf("portfolio %s: %w", portfolioID, err)
		}
		result.EpochIDs = append(result.EpochIDs, epoch.ID)
	}
	return result, nil
}

func (s *Store) BuildReplayActivationManifest(ctx context.Context, activationGeneration int64) (ReplayActivationManifest, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return ReplayActivationManifest{}, err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return ReplayActivationManifest{}, err
	}
	generation, err := replayGenerationTx(ctx, tx, activationGeneration)
	if err != nil {
		return ReplayActivationManifest{}, err
	}
	manifest, err := buildReplayActivationManifestTx(ctx, tx, generation)
	if err != nil {
		return ReplayActivationManifest{}, err
	}
	if err := tx.Commit(); err != nil {
		return ReplayActivationManifest{}, err
	}
	return manifest, nil
}

func (s *Store) ActivateReplayGeneration(
	ctx context.Context,
	activationGeneration int64,
	expectedManifestSHA256 string,
	now time.Time,
) error {
	if len(expectedManifestSHA256) != 64 || strings.ToLower(expectedManifestSHA256) != expectedManifestSHA256 {
		return errors.New("activation manifest SHA256 must be 64 lowercase hexadecimal characters")
	}
	if _, err := hex.DecodeString(expectedManifestSHA256); err != nil {
		return errors.New("activation manifest SHA256 must be lowercase hexadecimal")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return err
	}
	if err := lockReplayFinalizationTx(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE investment.portfolios IN SHARE MODE`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE investment.transaction_entries IN SHARE MODE`); err != nil {
		return err
	}
	generation, err := replayGenerationTx(ctx, tx, activationGeneration)
	if err != nil {
		return err
	}
	if err := assertReplayGenerationPreactiveTx(ctx, tx, generation); err != nil {
		return err
	}
	manifest, err := buildReplayActivationManifestTx(ctx, tx, generation)
	if err != nil {
		return err
	}
	if manifest.SHA256 != expectedManifestSHA256 {
		return fmt.Errorf("activation manifest changed: expected %s got %s", expectedManifestSHA256, manifest.SHA256)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.replay_policy_events (
			event_id, policy_version, activation_generation, event_type,
			manifest_sha256, portfolio_count, reason_code, created_at
		) VALUES ($1::uuid, $2, $3, 'ACTIVATED', $4, $5, NULL, $6)
	`, uuid.NewString(), generation.PolicyVersion, generation.ActivationGeneration,
		manifest.SHA256, manifest.PortfolioCount, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) InvalidateReplayGeneration(ctx context.Context, activationGeneration int64, reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 1000 {
		return errors.New("invalidation reason_code must be 1..1000 bytes")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := assertReplayAdminOwnerTx(ctx, tx); err != nil {
		return err
	}
	if err := lockReplayFinalizationTx(ctx, tx); err != nil {
		return err
	}
	generation, err := replayGenerationTx(ctx, tx, activationGeneration)
	if err != nil {
		return err
	}
	var activated, invalidated bool
	if err := tx.QueryRowContext(ctx, `
		SELECT
			EXISTS (
				SELECT 1 FROM analytics.replay_policy_events
				WHERE policy_version = $1 AND activation_generation = $2 AND event_type = 'ACTIVATED'
			),
			EXISTS (
				SELECT 1 FROM analytics.replay_policy_events
				WHERE policy_version = $1 AND activation_generation = $2 AND event_type = 'INVALIDATED'
			)
	`, generation.PolicyVersion, generation.ActivationGeneration).Scan(&activated, &invalidated); err != nil {
		return err
	}
	if !activated || invalidated {
		return errors.New("only an active non-invalidated generation can be invalidated")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.replay_policy_events (
			event_id, policy_version, activation_generation, event_type,
			manifest_sha256, portfolio_count, reason_code, created_at
		) VALUES ($1::uuid, $2, $3, 'INVALIDATED', NULL, NULL, $4, $5)
	`, uuid.NewString(), generation.PolicyVersion, generation.ActivationGeneration, reason, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) VerifyReplayGeneration(ctx context.Context, activationGeneration int64) (ReplayActivationManifest, error) {
	return s.BuildReplayActivationManifest(ctx, activationGeneration)
}

func replayGenerationTx(ctx context.Context, tx *sql.Tx, activationGeneration int64) (replayPolicyState, error) {
	var generation replayPolicyState
	if err := tx.QueryRowContext(ctx, `
		SELECT policy_version, activation_generation, replay_bound_raw_rows
		FROM analytics.replay_policy_generations
		WHERE policy_version = $1 AND activation_generation = $2
	`, replayPolicyVersion, activationGeneration).Scan(
		&generation.PolicyVersion, &generation.ActivationGeneration, &generation.ReplayBoundRawRows,
	); err != nil {
		return replayPolicyState{}, err
	}
	generation.PositionMethodology = replayPositionMethodology
	generation.SnapshotMethodology = replaySnapshotMethodology
	if generation.PolicyVersion != replayPolicyVersion || generation.ReplayBoundRawRows != replayBoundRawRows {
		return replayPolicyState{}, ErrReplayStateStale
	}
	return generation, nil
}

func assertReplayGenerationPreactiveTx(ctx context.Context, tx *sql.Tx, generation replayPolicyState) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM analytics.replay_policy_events
			WHERE policy_version = $1 AND activation_generation = $2
		)
	`, generation.PolicyVersion, generation.ActivationGeneration).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("replay generation is no longer PREACTIVE")
	}
	return nil
}

func activePortfolioIDsTx(ctx context.Context, tx *sql.Tx) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text
		FROM investment.portfolios
		WHERE portfolio_state = 'active'
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func buildAdminReplayEpochTx(
	ctx context.Context,
	tx *sql.Tx,
	generation replayPolicyState,
	portfolioID string,
	targetMutableRows int,
	now time.Time,
) (replayEpoch, error) {
	raw, err := loadAllRawLedgerTx(ctx, tx, portfolioID)
	if err != nil {
		return replayEpoch{}, err
	}
	prefix, mutable, boundarySequence, boundaryDate, err := partitionReplayBoundary(raw, targetMutableRows)
	if err != nil {
		return replayEpoch{}, err
	}

	base := replayEpoch{
		PortfolioID: portfolioID, PolicyVersion: generation.PolicyVersion,
		ActivationGeneration: generation.ActivationGeneration, BoundaryTradeDate: boundaryDate,
		BoundaryLogicalSequence: boundarySequence, PositionMethodology: replayPositionMethodology,
		SnapshotMethodology: replaySnapshotMethodology, Positions: map[string]replayPosition{},
		Financial: replayFinancialState{Amounts: zeroCashFlowAmounts(), InvestedCapital: decimal.Zero(), InputWatermark: "empty"},
	}
	boundaryState := cloneReplayWorkingState(base)
	if len(prefix) > 0 {
		canonicalPrefix, err := canonicalizeBoundedLedger(ctx, prefix)
		if err != nil {
			return replayEpoch{}, err
		}
		boundaryState, err = replayStateAtDate(ctx, base, canonicalPrefix, prefix, *boundaryDate)
		if err != nil {
			return replayEpoch{}, err
		}
	}

	epochGeneration, err := nextReplayEpochGenerationTx(ctx, tx, portfolioID)
	if err != nil {
		return replayEpoch{}, err
	}
	buildWatermark := int64(0)
	if len(raw) > 0 {
		buildWatermark = raw[len(raw)-1].LedgerSequence
	}
	epoch := base
	epoch.ID = uuid.NewString()
	epoch.EpochGeneration = epochGeneration
	epoch.BuildRawLedgerWatermark = buildWatermark
	epoch.Positions = boundaryState.Positions
	epoch.Financial = boundaryState.Financial
	epoch.LatestPositionTradeDate = boundaryState.LatestPositionTradeDate

	canonicalMutable, err := canonicalizeBoundedLedger(ctx, mutable)
	if err != nil {
		return replayEpoch{}, err
	}
	finalState := cloneReplayWorkingState(epoch)
	if maxDate := maxReplayTradeDate(canonicalMutable, epoch.BoundaryTradeDate); maxDate != "" {
		finalState, err = replayStateAtDate(ctx, epoch, canonicalMutable, mutable, maxDate)
		if err != nil {
			return replayEpoch{}, err
		}
	}
	epoch.SourceStateSHA256 = replaySourceStateDigest(epoch, finalState)
	return epoch, nil
}

func partitionReplayBoundary(
	raw []boundedLedgerRow,
	targetMutableRows int,
) (prefix []boundedLedgerRow, mutable []boundedLedgerRow, boundarySequence int64, boundaryDate *string, err error) {
	if targetMutableRows < 0 || targetMutableRows > replayBoundRawRows {
		return nil, nil, 0, nil, errors.New("invalid target mutable row count")
	}
	if len(raw) == 0 {
		return []boundedLedgerRow{}, []boundedLedgerRow{}, 0, nil, nil
	}

	firstByFamily := map[string]int{}
	for index, row := range raw {
		if row.ReversesTransactionID == nil {
			if _, exists := firstByFamily[row.TransactionID]; !exists {
				firstByFamily[row.TransactionID] = index
			}
		}
	}
	frozenCount := len(raw) - targetMutableRows
	if frozenCount < 0 {
		frozenCount = 0
	}
	for {
		moved := false
		for index := frozenCount; index < len(raw); index++ {
			row := raw[index]
			target := row.TransactionID
			if row.ReversesTransactionID != nil {
				target = *row.ReversesTransactionID
			}
			first, exists := firstByFamily[target]
			if !exists {
				return nil, nil, 0, nil, ErrUnsupportedPositionLedger
			}
			if first < frozenCount {
				frozenCount = first
				moved = true
				break
			}
		}
		if !moved {
			break
		}
	}
	if len(raw)-frozenCount > replayBoundRawRows {
		return nil, nil, 0, nil, fmt.Errorf("%w: no family-closed boundary fits %d raw rows", ErrReplayStateStale, replayBoundRawRows)
	}
	prefix = append([]boundedLedgerRow(nil), raw[:frozenCount]...)
	mutable = append([]boundedLedgerRow(nil), raw[frozenCount:]...)
	if len(prefix) == 0 {
		return prefix, mutable, 0, nil, nil
	}
	boundarySequence = prefix[len(prefix)-1].LedgerSequence
	maxFrozenDate := prefix[0].TradeDate
	for _, row := range prefix[1:] {
		if row.TradeDate > maxFrozenDate {
			maxFrozenDate = row.TradeDate
		}
	}
	for _, row := range mutable {
		if row.TradeDate < maxFrozenDate {
			return nil, nil, 0, nil, fmt.Errorf("%w: raw/effective time ordering crosses proposed immutable boundary", ErrReplayStateStale)
		}
	}
	boundaryDate = &maxFrozenDate
	return prefix, mutable, boundarySequence, boundaryDate, nil
}

func loadAllRawLedgerTx(ctx context.Context, tx *sql.Tx, portfolioID string) ([]boundedLedgerRow, error) {
	countReplayFullHistoryLoad(ctx)
	rows, err := tx.QueryContext(ctx, `
		SELECT
			te.entry_id::text, te.transaction_id::text, te.asset_id::text, a.asset_type,
			te.transaction_type, te.quantity::text, te.unit_price_amount::text,
			te.gross_amount::text, te.commission_amount::text, te.tax_amount::text,
			te.trade_date::text, te.ledger_sequence, te.revision,
			te.prior_entry_id::text, te.reverses_transaction_id::text,
			te.created_at, te.created_at::text
		FROM investment.transaction_entries te
		LEFT JOIN investment.assets a ON a.id = te.asset_id
		WHERE te.portfolio_id = $1::uuid
		ORDER BY te.ledger_sequence ASC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := scanRawLedgerRows(rows)
	if err == nil {
		countReplayHistoryRowsLoaded(ctx, len(result))
	}
	return result, err
}

func scanRawLedgerRows(rows *sql.Rows) ([]boundedLedgerRow, error) {
	result := make([]boundedLedgerRow, 0)
	for rows.Next() {
		var row boundedLedgerRow
		var assetID, assetType, quantity, unitPrice, priorID, reversesID sql.NullString
		var gross, commission, tax string
		if err := rows.Scan(&row.EntryID, &row.TransactionID, &assetID, &assetType, &row.TransactionType,
			&quantity, &unitPrice, &gross, &commission, &tax, &row.TradeDate, &row.LedgerSequence,
			&row.Revision, &priorID, &reversesID, &row.CreatedAt, &row.CreatedAtText); err != nil {
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
	return result, rows.Err()
}

func buildReplayActivationManifestTx(
	ctx context.Context,
	tx *sql.Tx,
	generation replayPolicyState,
) (ReplayActivationManifest, error) {
	portfolioIDs, err := activePortfolioIDsTx(ctx, tx)
	if err != nil {
		return ReplayActivationManifest{}, err
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "oi-new-04-activation-manifest-v1\npolicy_version=%s\nactivation_generation=%d\nreplay_bound_raw_rows=%d\n",
		generation.PolicyVersion, generation.ActivationGeneration, generation.ReplayBoundRawRows)
	for _, portfolioID := range portfolioIDs {
		var epochID, sourceSHA string
		var epochGeneration, buildWatermark int64
		err := tx.QueryRowContext(ctx, `
			SELECT epoch_id::text, epoch_generation, build_raw_ledger_watermark, source_state_sha256
			FROM analytics.portfolio_replay_epochs
			WHERE portfolio_id = $1::uuid
			  AND policy_version = $2
			  AND activation_generation = $3
			ORDER BY epoch_generation DESC
			LIMIT 1
		`, portfolioID, generation.PolicyVersion, generation.ActivationGeneration).Scan(
			&epochID, &epochGeneration, &buildWatermark, &sourceSHA,
		)
		if err != nil {
			return ReplayActivationManifest{}, fmt.Errorf("portfolio %s latest epoch: %w", portfolioID, err)
		}
		latest, err := latestLedgerSequenceTx(ctx, tx, portfolioID)
		if err != nil {
			return ReplayActivationManifest{}, err
		}
		if latest != buildWatermark {
			return ReplayActivationManifest{}, fmt.Errorf("portfolio %s ledger watermark changed after finalization", portfolioID)
		}
		epoch, err := loadLatestReplayEpochTx(ctx, tx, portfolioID, generation)
		if err != nil {
			return ReplayActivationManifest{}, fmt.Errorf("portfolio %s replay epoch: %w", portfolioID, err)
		}
		mutableRows, err := loadRawLedgerRangeTx(ctx, tx, portfolioID, epoch.BoundaryLogicalSequence, epoch.BuildRawLedgerWatermark, replayBoundRawRows+1)
		if err != nil {
			return ReplayActivationManifest{}, err
		}
		if len(mutableRows) > replayBoundRawRows {
			return ReplayActivationManifest{}, fmt.Errorf("portfolio %s replay suffix exceeds hard bound", portfolioID)
		}
		if err := verifyEpochSourceState(ctx, epoch, mutableRows); err != nil {
			return ReplayActivationManifest{}, fmt.Errorf("portfolio %s replay state witness: %w", portfolioID, err)
		}
		fmt.Fprintf(&builder, "portfolio=%s|epoch=%s|epoch_generation=%d|watermark=%d|source_state_sha256=%s\n",
			portfolioID, epochID, epochGeneration, buildWatermark, sourceSHA)
	}
	canonical := []byte(builder.String())
	digest := sha256.Sum256(canonical)
	return ReplayActivationManifest{
		CanonicalBytes: canonical,
		SHA256:         hex.EncodeToString(digest[:]),
		PortfolioCount: len(portfolioIDs),
	}, nil
}

func assertReplayAdminOwnerTx(ctx context.Context, tx *sql.Tx) error {
	var currentUser, owner string
	if err := tx.QueryRowContext(ctx, `
		SELECT current_user::text, pg_get_userbyid(c.relowner)::text
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'analytics' AND c.relname = 'replay_policy_generations'
	`).Scan(&currentUser, &owner); err != nil {
		return err
	}
	if strings.TrimSpace(currentUser) == "" || currentUser != owner {
		return errors.New("replay admin command requires the replay-schema owner connection")
	}
	return nil
}
