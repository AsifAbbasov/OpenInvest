package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) GetPortfolioSummaryStage371(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioSummary, error) {
	if _, err := s.GetPortfolio(ctx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioSummary{}, err
	}
	return getPortfolioSummaryStage371(ctx, s.db, portfolioID, asOfDate)
}

func getPortfolioSummaryStage371(
	ctx context.Context,
	db *sql.DB,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioSummary, error) {
	var summary verticalslice.PortfolioSummary
	var totalValue string
	var cashValue string
	var stockValue string
	var bondValue string
	var investedCapital string
	var nominalReturn string
	var calculatedAt time.Time
	args := []any{portfolioID}
	dateFilter := ""
	if asOfDate != "" {
		dateFilter = "AND snapshot_date <= $2::date"
		args = append(args, asOfDate)
	}

	err := db.QueryRowContext(ctx, `
		SELECT portfolio_id, snapshot_date::text, total_value_amount::text, cash_value_amount::text,
			stock_value_amount::text, bond_value_amount::text, invested_capital_amount::text,
			nominal_return_rate::text, methodology_version, calculated_at
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_status = 'calculated'
			`+dateFilter+`
		ORDER BY
			snapshot_date DESC,
			CASE methodology_version
				WHEN 'stage-03-71-position-cost-snapshot-v1' THEN 2
				WHEN 'stage-03-02-local-cost-snapshot-v1' THEN 1
				ELSE 0
			END DESC,
			snapshot_version DESC,
			calculated_at DESC,
			id DESC
		LIMIT 1
	`, args...).Scan(
		&summary.PortfolioID,
		&summary.AsOfDate,
		&totalValue,
		&cashValue,
		&stockValue,
		&bondValue,
		&investedCapital,
		&nominalReturn,
		&summary.MethodologyVersion,
		&calculatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return verticalslice.PortfolioSummary{}, ErrNotFound
	}
	if err != nil {
		return verticalslice.PortfolioSummary{}, err
	}

	summary.TotalValue = verticalslice.Money{Amount: decimal.Must(totalValue), Currency: verticalslice.RUB}
	summary.CashValue = verticalslice.Money{Amount: decimal.Must(cashValue), Currency: verticalslice.RUB}
	summary.StockValue = verticalslice.Money{Amount: decimal.Must(stockValue), Currency: verticalslice.RUB}
	summary.BondValue = verticalslice.Money{Amount: decimal.Must(bondValue), Currency: verticalslice.RUB}
	summary.InvestedCapital = verticalslice.Money{Amount: decimal.Must(investedCapital), Currency: verticalslice.RUB}
	summary.DividendsReceived = verticalslice.ZeroMoney()
	summary.CouponsReceived = verticalslice.ZeroMoney()
	summary.NominalReturnRate = decimal.Must(nominalReturn)
	summary.RealReturn = verticalslice.RealReturn{
		NominalReturnRate: summary.NominalReturnRate,
		InflationRate:     decimal.Zero(),
		RealReturnRate:    summary.NominalReturnRate,
		NominalGain:       summary.TotalValue.Sub(summary.InvestedCapital),
		RealGain:          summary.TotalValue.Sub(summary.InvestedCapital),
		FromDate:          summary.AsOfDate,
		ToDate:            summary.AsOfDate,
		Methodology:       "stage-03-02-no-inflation-placeholder-v1",
	}
	summary.PurchasingPower = verticalslice.PurchasingPower{
		PortfolioValue: summary.TotalValue,
		AsOfDate:       summary.AsOfDate,
		Equivalents:    []verticalslice.PurchasingPowerEquivalent{},
	}
	summary.Positions = []verticalslice.PortfolioPosition{}
	summary.CalculatedAt = calculatedAt
	return summary, nil
}
