package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type manualValuationRecord struct {
	AssetID            string
	PositionGeneration int64
	Price              decimal.Decimal
	Currency           string
	AsOfDate           string
	Source             string
}

func (s *Store) Stage376Ready(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
        SELECT portfolio_id, asset_id, position_opened_ledger_sequence,
               price_amount, price_currency, as_of_date, source, updated_at
        FROM investment.portfolio_manual_valuations
        WHERE false
    `)
	if err != nil {
		return err
	}
	return rows.Close()
}

func (s *Store) GetPortfolioPositions(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioPositionsProjection, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	defer rollback(tx)

	if _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}

	result, err := portfolioPositionsProjectionTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	return result, nil
}

// portfolioPositionsProjectionTx is the canonical Stage 3.76 transaction-level valuation
// projection. Callers that already own a consistent database snapshot reuse this helper
// rather than reimplementing position rebuild, manual valuation, and cash orchestration.
func portfolioPositionsProjectionTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioPositionsProjection, error) {
	rebuilt, latestIncludedTradeDate, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	manualValuations, err := loadManualValuationsTx(ctx, tx, portfolioID)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	cashValue, err := effectiveProjectionCashTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}

	inputsAsOf := latestIncludedTradeDate
	if asOfDate != "" {
		value := asOfDate
		inputsAsOf = &value
	}
	return buildPortfolioPositionsProjectionWithValuations(
		rebuilt,
		inputsAsOf,
		asOfDate,
		manualValuations,
		cashValue,
	)
}

// Retained for Stage 3.72 unit tests and compatibility witnesses. Production reads use the
// Stage 3.76 enrichment path above.
func buildPortfolioPositionsProjection(
	rebuilt []rebuiltPortfolioPosition,
	inputsAsOf *string,
) (verticalslice.PortfolioPositionsProjection, error) {
	return buildPortfolioPositionsProjectionWithValuations(rebuilt, inputsAsOf, "", nil, decimal.Zero())
}

func buildPortfolioPositionsProjectionWithValuations(
	rebuilt []rebuiltPortfolioPosition,
	inputsAsOf *string,
	requestedAsOfDate string,
	manualValuations map[string]manualValuationRecord,
	cashValue decimal.Decimal,
) (verticalslice.PortfolioPositionsProjection, error) {
	sort.SliceStable(rebuilt, func(left, right int) bool {
		if rebuilt[left].Ticker == rebuilt[right].Ticker {
			return rebuilt[left].AssetID < rebuilt[right].AssetID
		}
		return rebuilt[left].Ticker < rebuilt[right].Ticker
	})

	items := make([]verticalslice.PortfolioPositionProjection, 0, len(rebuilt))
	totalAcquisitionBasis := decimal.Zero()
	valuedMarketValue := decimal.Zero()
	valuedAcquisitionBasis := decimal.Zero()
	valuedUnrealizedGain := decimal.Zero()
	valuedIndexes := make([]int, 0, len(rebuilt))

	for _, rebuiltPosition := range rebuilt {
		assetType, err := publicAssetType(rebuiltPosition.AssetType)
		if err != nil {
			return verticalslice.PortfolioPositionsProjection{}, err
		}
		if !rebuiltPosition.State.Open {
			continue
		}

		totalAcquisitionBasis = totalAcquisitionBasis.Add(rebuiltPosition.State.AcquisitionBasis)
		if !totalAcquisitionBasis.FitsStorage() {
			return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: total acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
		}

		item := verticalslice.PortfolioPositionProjection{
			Ticker:    rebuiltPosition.Ticker,
			AssetType: assetType,
			Quantity:  rebuiltPosition.State.Quantity,
			WeightedAverageCost: verticalslice.Money{
				Amount: rebuiltPosition.State.WeightedAverageCost, Currency: verticalslice.RUB,
			},
			AcquisitionBasis: verticalslice.Money{
				Amount: rebuiltPosition.State.AcquisitionBasis, Currency: verticalslice.RUB,
			},
			MarketValuation: verticalslice.MarketValuationUnavailable{
				Status: verticalslice.MarketValuationUnavailableStatus,
				Reason: verticalslice.MarketValuationUnavailableReason,
			},
		}

		record, hasManual := manualValuations[rebuiltPosition.AssetID]
		eligibleDate := requestedAsOfDate == "" || record.AsOfDate == requestedAsOfDate
		if hasManual && eligibleDate && record.PositionGeneration == rebuiltPosition.PositionGeneration {
			if record.Currency != verticalslice.RUB || record.Source != verticalslice.ManualValuationSourceUserSupplied {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: stored manual valuation provenance is invalid", verticalslice.ErrInvalidInput)
			}
			marketValue := rebuiltPosition.State.Quantity.Mul(record.Price)
			unrealizedGain := marketValue.Sub(rebuiltPosition.State.AcquisitionBasis)
			if !marketValue.FitsStorage() || !unrealizedGain.FitsStorage() {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: derived market valuation exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
			unrealizedReturn, err := verticalsliceManualReturn(unrealizedGain, rebuiltPosition.State.AcquisitionBasis)
			if err != nil {
				return verticalslice.PortfolioPositionsProjection{}, err
			}
			priceMoney := verticalslice.Money{Amount: record.Price, Currency: verticalslice.RUB}
			valueMoney := verticalslice.Money{Amount: marketValue, Currency: verticalslice.RUB}
			gainMoney := verticalslice.Money{Amount: unrealizedGain, Currency: verticalslice.RUB}
			asOf := record.AsOfDate
			item.MarketValuation = verticalslice.MarketValuationUnavailable{
				Status:           verticalslice.MarketValuationAvailableStatus,
				Source:           verticalslice.ManualValuationSourceUserSupplied,
				MarketPrice:      &priceMoney,
				MarketValue:      &valueMoney,
				UnrealizedGain:   &gainMoney,
				UnrealizedReturn: unrealizedReturn,
				AsOf:             &asOf,
			}
			valuedMarketValue = valuedMarketValue.Add(marketValue)
			valuedAcquisitionBasis = valuedAcquisitionBasis.Add(rebuiltPosition.State.AcquisitionBasis)
			valuedUnrealizedGain = valuedUnrealizedGain.Add(unrealizedGain)
			if !valuedMarketValue.FitsStorage() || !valuedAcquisitionBasis.FitsStorage() || !valuedUnrealizedGain.FitsStorage() {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: portfolio valuation aggregate exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
			valuedIndexes = append(valuedIndexes, len(items))
		}
		items = append(items, item)
	}

	if totalAcquisitionBasis.IsPositive() {
		for index := range items {
			weight, err := items[index].AcquisitionBasis.Amount.Div(totalAcquisitionBasis)
			if err != nil || !weight.FitsStorage() {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: acquisition basis weight exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
			}
			items[index].AcquisitionBasisWeight = decimalPointer(weight)
		}
	}
	if valuedMarketValue.IsPositive() {
		for _, index := range valuedIndexes {
			weight, err := items[index].MarketValuation.MarketValue.Amount.Div(valuedMarketValue)
			if err != nil || !weight.FitsStorage() {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: market weight exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
			}
			items[index].MarketValuation.MarketWeight = decimalPointer(weight)
		}
	}

	coverageStatus := verticalslice.ValuationCoveragePartialStatus
	var currentPortfolioValue *verticalslice.Money
	if len(valuedIndexes) == len(items) {
		coverageStatus = verticalslice.ValuationCoverageCompleteStatus
		currentValue := cashValue.Add(valuedMarketValue)
		if !currentValue.FitsStorage() {
			return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: current portfolio value exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
		}
		value := verticalslice.Money{Amount: currentValue, Currency: verticalslice.RUB}
		currentPortfolioValue = &value
	}

	return verticalslice.PortfolioPositionsProjection{
		Items:                 items,
		TotalAcquisitionBasis: verticalslice.Money{Amount: totalAcquisitionBasis, Currency: verticalslice.RUB},
		ValuationSummary: verticalslice.PortfolioValuationSummary{
			Status:                          coverageStatus,
			ValuedPositions:                 len(valuedIndexes),
			TotalOpenPositions:              len(items),
			ValuedPositionsMarketValue:      verticalslice.Money{Amount: valuedMarketValue, Currency: verticalslice.RUB},
			ValuedPositionsAcquisitionBasis: verticalslice.Money{Amount: valuedAcquisitionBasis, Currency: verticalslice.RUB},
			TotalAcquisitionBasis:           verticalslice.Money{Amount: totalAcquisitionBasis, Currency: verticalslice.RUB},
			UnrealizedGain:                  verticalslice.Money{Amount: valuedUnrealizedGain, Currency: verticalslice.RUB},
			CashValue:                       verticalslice.Money{Amount: cashValue, Currency: verticalslice.RUB},
			CurrentPortfolioValue:           currentPortfolioValue,
		},
		InputsAsOf:         copyStringPointer(inputsAsOf),
		MethodologyVersion: verticalslice.PortfolioPositionProjectionMethodologyVersion,
	}, nil
}

func (s *Store) UpsertManualValuation(
	ctx context.Context,
	subjectID string,
	request verticalslice.ManualValuationRequest,
) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := getPortfolioTx(ctx, tx, subjectID, request.PortfolioID); err != nil {
		return err
	}
	target, err := currentOpenPositionByTickerTx(ctx, tx, request.PortfolioID, request.Ticker)
	if err != nil {
		return err
	}
	if target.PositionGeneration <= 0 {
		return ErrLedgerSequenceUnavailable
	}
	if _, err := tx.ExecContext(ctx, `
        INSERT INTO investment.portfolio_manual_valuations (
            portfolio_id, asset_id, position_opened_ledger_sequence,
            price_amount, price_currency, as_of_date, source, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6::date, $7, now())
        ON CONFLICT (portfolio_id, asset_id) DO UPDATE SET
            position_opened_ledger_sequence = EXCLUDED.position_opened_ledger_sequence,
            price_amount = EXCLUDED.price_amount,
            price_currency = EXCLUDED.price_currency,
            as_of_date = EXCLUDED.as_of_date,
            source = EXCLUDED.source,
            updated_at = now()
    `, request.PortfolioID, target.AssetID, target.PositionGeneration,
		request.Price.Amount.String(), request.Price.Currency, request.AsOfDate,
		verticalslice.ManualValuationSourceUserSupplied); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ClearManualValuation(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	ticker string,
) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
        DELETE FROM investment.portfolio_manual_valuations valuation
        USING investment.assets asset
        WHERE valuation.portfolio_id = $1
          AND valuation.asset_id = asset.id
          AND asset.ticker = $2
    `, portfolioID, ticker); err != nil {
		return err
	}
	return tx.Commit()
}

func currentOpenPositionByTickerTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	ticker string,
) (rebuiltPortfolioPosition, error) {
	rebuilt, _, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, "")
	if err != nil {
		return rebuiltPortfolioPosition{}, err
	}
	for _, item := range rebuilt {
		if item.State.Open && item.Ticker == ticker {
			return item, nil
		}
	}
	return rebuiltPortfolioPosition{}, verticalslice.ErrNotFound
}

func loadManualValuationsTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
) (map[string]manualValuationRecord, error) {
	rows, err := tx.QueryContext(ctx, `
        SELECT
            asset_id::text,
            position_opened_ledger_sequence,
            price_amount::text,
            price_currency,
            as_of_date::text,
            source
        FROM investment.portfolio_manual_valuations
        WHERE portfolio_id = $1
    `, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]manualValuationRecord{}
	for rows.Next() {
		var record manualValuationRecord
		var price string
		if err := rows.Scan(
			&record.AssetID,
			&record.PositionGeneration,
			&price,
			&record.Currency,
			&record.AsOfDate,
			&record.Source,
		); err != nil {
			return nil, err
		}
		parsed, err := decimal.FromString(price)
		if err != nil || !parsed.IsPositive() || !parsed.FitsStorage() ||
			record.PositionGeneration <= 0 || record.Currency != verticalslice.RUB ||
			record.Source != verticalslice.ManualValuationSourceUserSupplied {
			return nil, fmt.Errorf("%w: stored manual valuation is invalid", verticalslice.ErrInvalidInput)
		}
		if _, err := time.Parse("2006-01-02", record.AsOfDate); err != nil {
			return nil, fmt.Errorf("%w: stored manual valuation date is invalid", verticalslice.ErrInvalidInput)
		}
		record.Price = parsed
		result[record.AssetID] = record
	}
	return result, rows.Err()
}

func effectiveProjectionCashTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	asOfDate string,
) (decimal.Decimal, error) {
	rows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return decimal.Zero(), err
	}
	amounts := zeroCashFlowAmounts()
	for _, row := range rows {
		if err := amounts.apply(row); err != nil {
			return decimal.Zero(), err
		}
	}
	cash := amounts.netCashFlow()
	if !cash.FitsStorage() {
		return decimal.Zero(), fmt.Errorf("%w: cash value exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
	}
	return cash, nil
}

func verticalsliceManualReturn(gain decimal.Decimal, basis decimal.Decimal) (*decimal.Decimal, error) {
	if basis.IsZero() {
		return nil, nil
	}
	value, err := gain.Div(basis)
	if err != nil || !value.FitsStorage() {
		return nil, fmt.Errorf("%w: unrealized return exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
	}
	return decimalPointer(value), nil
}

func publicAssetType(assetType string) (string, error) {
	switch assetType {
	case "stock":
		return "STOCK", nil
	case "bond":
		return "BOND", nil
	default:
		return "", fmt.Errorf("%w: unsupported asset type %q in position projection", verticalslice.ErrInvalidInput, assetType)
	}
}

func decimalPointer(value decimal.Decimal) *decimal.Decimal {
	copy := value
	return &copy
}

func copyStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
