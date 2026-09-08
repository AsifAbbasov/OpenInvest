package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) GetPortfolioReturns(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioReturnProjection, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}
	defer rollback(tx)
	if _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}
	projection, err := portfolioReturnProjectionTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}
	return projection, nil
}

func portfolioReturnProjectionTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioReturnProjection, error) {
	ledgerRows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}
	externalCashFlows := make([]verticalslice.PortfolioReturnCashFlow, 0, len(ledgerRows))
	for _, row := range ledgerRows {
		switch row.TransactionType {
		case "DEPOSIT":
			externalCashFlows = append(externalCashFlows, verticalslice.PortfolioReturnCashFlow{
				Date:   row.TradeDate,
				Amount: decimal.Zero().Sub(row.GrossAmount),
			})
		case "WITHDRAWAL":
			externalCashFlows = append(externalCashFlows, verticalslice.PortfolioReturnCashFlow{
				Date:   row.TradeDate,
				Amount: row.GrossAmount,
			})
		case "BUY", "SELL", "DIVIDEND", "COUPON", "FEE", "TAX":
			// Internal portfolio activity affects canonical cash/terminal value but is not
			// an external investor cash flow for money-weighted return.
			continue
		default:
			return verticalslice.PortfolioReturnProjection{}, fmt.Errorf(
				"%w: unsupported transaction type %q in effective ledger",
				verticalslice.ErrInvalidInput,
				row.TransactionType,
			)
		}
	}

	positions, err := portfolioPositionsProjectionTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioReturnProjection{}, err
	}

	return verticalslice.BuildPortfolioReturnProjection(
		portfolioID,
		asOfDate,
		externalCashFlows,
		positions.ValuationSummary.CurrentPortfolioValue,
		positions.ValuationSummary.Status == verticalslice.ValuationCoverageCompleteStatus,
	)
}
