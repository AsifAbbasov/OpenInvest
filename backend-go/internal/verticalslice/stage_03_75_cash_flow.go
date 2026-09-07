package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const PortfolioCashFlowMethodologyVersion = "portfolio-cash-flow-income-v1"

var ErrCashFlowProjectionUnavailable = errors.New("portfolio cash-flow projection is unavailable")

type PortfolioCashFlowStore interface {
	GetPortfolioCashFlow(
		ctx context.Context,
		subjectID string,
		portfolioID string,
		fromDate string,
		toDate string,
	) (PortfolioCashFlowProjection, error)
}

type PortfolioCashFlowTotals struct {
	Deposits            Money
	Withdrawals         Money
	BuyOutflows         Money
	SellInflows         Money
	DividendsGross      Money
	CouponsGross        Money
	Fees                Money
	Taxes               Money
	NetExternalFlow     Money
	NetInvestmentIncome Money
	NetCashFlow         Money
}

type PortfolioCashFlowPeriod struct {
	Month  string
	Totals PortfolioCashFlowTotals
}

type PortfolioCashFlowProjection struct {
	PortfolioID       string
	FromDate          *string
	ToDate            *string
	Totals            PortfolioCashFlowTotals
	Periods           []PortfolioCashFlowPeriod
	InputsAsOf        *string
	MethodologyVersion string
}

func (s *Service) GetPortfolioCashFlow(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	fromDate string,
	toDate string,
) (PortfolioCashFlowProjection, error) {
	fromDate = strings.TrimSpace(fromDate)
	toDate = strings.TrimSpace(toDate)
	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err != nil {
			return PortfolioCashFlowProjection{}, fmt.Errorf("%w: fromDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err != nil {
			return PortfolioCashFlowProjection{}, fmt.Errorf("%w: toDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	if fromDate != "" && toDate != "" && fromDate > toDate {
		return PortfolioCashFlowProjection{}, fmt.Errorf("%w: fromDate must be before or equal to toDate", ErrInvalidInput)
	}
	store, ok := s.store.(PortfolioCashFlowStore)
	if !ok {
		return PortfolioCashFlowProjection{}, ErrCashFlowProjectionUnavailable
	}
	return store.GetPortfolioCashFlow(ctx, subjectID, portfolioID, fromDate, toDate)
}
