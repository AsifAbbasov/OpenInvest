package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type cashFlowAmounts struct {
	Deposits            decimal.Decimal
	Withdrawals         decimal.Decimal
	BuyOutflows         decimal.Decimal
	SellInflows         decimal.Decimal
	DividendsGross      decimal.Decimal
	CouponsGross        decimal.Decimal
	Fees                decimal.Decimal
	Taxes               decimal.Decimal
	NetInvestmentIncome decimal.Decimal
}

func zeroCashFlowAmounts() cashFlowAmounts {
	return cashFlowAmounts{
		Deposits:            decimal.Zero(),
		Withdrawals:         decimal.Zero(),
		BuyOutflows:         decimal.Zero(),
		SellInflows:         decimal.Zero(),
		DividendsGross:      decimal.Zero(),
		CouponsGross:        decimal.Zero(),
		Fees:                decimal.Zero(),
		Taxes:               decimal.Zero(),
		NetInvestmentIncome: decimal.Zero(),
	}
}

func (amounts *cashFlowAmounts) apply(row effectiveLedgerRow) error {
	amounts.Fees = amounts.Fees.Add(row.Commission)
	amounts.Taxes = amounts.Taxes.Add(row.Tax)

	switch row.TransactionType {
	case "DEPOSIT":
		amounts.Deposits = amounts.Deposits.Add(row.GrossAmount)
	case "WITHDRAWAL":
		amounts.Withdrawals = amounts.Withdrawals.Add(row.GrossAmount)
	case "BUY":
		amounts.BuyOutflows = amounts.BuyOutflows.Add(row.GrossAmount)
	case "SELL":
		amounts.SellInflows = amounts.SellInflows.Add(row.GrossAmount)
	case "DIVIDEND":
		amounts.DividendsGross = amounts.DividendsGross.Add(row.GrossAmount)
		amounts.NetInvestmentIncome = amounts.NetInvestmentIncome.Add(row.GrossAmount.Sub(row.Commission).Sub(row.Tax))
	case "COUPON":
		amounts.CouponsGross = amounts.CouponsGross.Add(row.GrossAmount)
		amounts.NetInvestmentIncome = amounts.NetInvestmentIncome.Add(row.GrossAmount.Sub(row.Commission).Sub(row.Tax))
	case "FEE":
		amounts.Fees = amounts.Fees.Add(row.GrossAmount)
	case "TAX":
		amounts.Taxes = amounts.Taxes.Add(row.GrossAmount)
	default:
		return fmt.Errorf("%w: unsupported transaction type %q in cash-flow projection", verticalslice.ErrInvalidInput, row.TransactionType)
	}
	return amounts.validateStorage()
}

func (amounts cashFlowAmounts) validateStorage() error {
	values := []decimal.Decimal{
		amounts.Deposits,
		amounts.Withdrawals,
		amounts.BuyOutflows,
		amounts.SellInflows,
		amounts.DividendsGross,
		amounts.CouponsGross,
		amounts.Fees,
		amounts.Taxes,
		amounts.netExternalFlow(),
		amounts.netInvestmentIncome(),
		amounts.netCashFlow(),
	}
	for _, value := range values {
		if !value.FitsStorage() {
			return fmt.Errorf("%w: cash-flow aggregate exceeds NUMERIC(28,8) storage precision", verticalslice.ErrInvalidInput)
		}
	}
	return nil
}

func (amounts cashFlowAmounts) netExternalFlow() decimal.Decimal {
	return amounts.Deposits.Sub(amounts.Withdrawals)
}

func (amounts cashFlowAmounts) netInvestmentIncome() decimal.Decimal {
	return amounts.NetInvestmentIncome
}

func (amounts cashFlowAmounts) netCashFlow() decimal.Decimal {
	return amounts.Deposits.
		Sub(amounts.Withdrawals).
		Sub(amounts.BuyOutflows).
		Add(amounts.SellInflows).
		Add(amounts.DividendsGross).
		Add(amounts.CouponsGross).
		Sub(amounts.Fees).
		Sub(amounts.Taxes)
}

func moneyFromDecimal(value decimal.Decimal) verticalslice.Money {
	return verticalslice.Money{Amount: value, Currency: verticalslice.RUB}
}

func (amounts cashFlowAmounts) totals() verticalslice.PortfolioCashFlowTotals {
	return verticalslice.PortfolioCashFlowTotals{
		Deposits:            moneyFromDecimal(amounts.Deposits),
		Withdrawals:         moneyFromDecimal(amounts.Withdrawals),
		BuyOutflows:         moneyFromDecimal(amounts.BuyOutflows),
		SellInflows:         moneyFromDecimal(amounts.SellInflows),
		DividendsGross:      moneyFromDecimal(amounts.DividendsGross),
		CouponsGross:        moneyFromDecimal(amounts.CouponsGross),
		Fees:                moneyFromDecimal(amounts.Fees),
		Taxes:               moneyFromDecimal(amounts.Taxes),
		NetExternalFlow:     moneyFromDecimal(amounts.netExternalFlow()),
		NetInvestmentIncome: moneyFromDecimal(amounts.netInvestmentIncome()),
		NetCashFlow:         moneyFromDecimal(amounts.netCashFlow()),
	}
}

func (s *Store) GetPortfolioCashFlow(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	fromDate string,
	toDate string,
) (verticalslice.PortfolioCashFlowProjection, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return verticalslice.PortfolioCashFlowProjection{}, err
	}
	defer rollback(tx)

	if _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioCashFlowProjection{}, err
	}
	projection, err := portfolioCashFlowProjectionTx(ctx, tx, portfolioID, fromDate, toDate)
	if err != nil {
		return verticalslice.PortfolioCashFlowProjection{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.PortfolioCashFlowProjection{}, err
	}
	return projection, nil
}

func portfolioCashFlowProjectionTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	fromDate string,
	toDate string,
) (verticalslice.PortfolioCashFlowProjection, error) {
	rows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, toDate)
	if err != nil {
		return verticalslice.PortfolioCashFlowProjection{}, err
	}

	totals := zeroCashFlowAmounts()
	monthly := map[string]cashFlowAmounts{}
	var latestIncludedTradeDate *string
	for _, row := range rows {
		if fromDate != "" && row.TradeDate < fromDate {
			continue
		}
		if err := totals.apply(row); err != nil {
			return verticalslice.PortfolioCashFlowProjection{}, err
		}
		month := row.TradeDate[:7]
		period, exists := monthly[month]
		if !exists {
			period = zeroCashFlowAmounts()
		}
		if err := period.apply(row); err != nil {
			return verticalslice.PortfolioCashFlowProjection{}, err
		}
		monthly[month] = period
		tradeDate := row.TradeDate
		if latestIncludedTradeDate == nil || tradeDate > *latestIncludedTradeDate {
			latestIncludedTradeDate = &tradeDate
		}
	}

	months := make([]string, 0, len(monthly))
	for month := range monthly {
		months = append(months, month)
	}
	sort.Strings(months)
	periods := make([]verticalslice.PortfolioCashFlowPeriod, 0, len(months))
	for _, month := range months {
		periods = append(periods, verticalslice.PortfolioCashFlowPeriod{Month: month, Totals: monthly[month].totals()})
	}

	inputsAsOf := latestIncludedTradeDate
	if toDate != "" {
		value := toDate
		inputsAsOf = &value
	}
	return verticalslice.PortfolioCashFlowProjection{
		PortfolioID:        portfolioID,
		FromDate:           optionalStringPointer(fromDate),
		ToDate:             optionalStringPointer(toDate),
		Totals:             totals.totals(),
		Periods:            periods,
		InputsAsOf:         copyStringPointer(inputsAsOf),
		MethodologyVersion: verticalslice.PortfolioCashFlowMethodologyVersion,
	}, nil
}

func optionalStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}
