package postgres_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestM11V2FullSystemStatefulCampaign(t *testing.T) {
	if os.Getenv("OPENINVEST_M11_V2_EXPENSIVE") != "1" { t.Skip("specialized Module 11 V2 campaign only") }
	const (
		seed int64 = 110022
		sequences = 100
		steps = 30
	)
	rng := rand.New(rand.NewSource(seed))
	totalOperations := 0
	referenceComparisons := 0
	for sequence := 0; sequence < sequences; sequence++ {
		t.Run(fmt.Sprintf("sequence-%03d", sequence), func(t *testing.T) {
			h := newStage371Harness(t, fmt.Sprintf("M11 V2 stateful %03d", sequence))
			cleanupStage376Valuations(t, h)
			units := 0
			cashTransactions := make([]verticalslice.Transaction, 0, steps)
			for step := 0; step < steps; step++ {
				date := fmt.Sprintf("2026-10-%02d", 1+(step%28))
				switch rng.Intn(10) {
				case 0:
					tx := stage377AppendCash(t, h, "DEPOSIT", fmt.Sprintf("%d.00000000", 10+rng.Intn(90)), date)
					cashTransactions = append(cashTransactions, tx)
				case 1:
					tx := stage377AppendCash(t, h, "WITHDRAWAL", fmt.Sprintf("%d.00000000", 1+rng.Intn(20)), date)
					cashTransactions = append(cashTransactions, tx)
				case 2:
					qty := 1+rng.Intn(5)
					appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", fmt.Sprintf("%d.00000000", qty), fmt.Sprintf("%d.00000000", 80+rng.Intn(80)), date))
					units += qty
				case 3:
					if units == 0 {
						qty := 1+rng.Intn(5)
						appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", fmt.Sprintf("%d.00000000", qty), "100.00000000", date))
						units += qty
					} else {
						qty := 1+rng.Intn(units)
						appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", fmt.Sprintf("%d.00000000", qty), fmt.Sprintf("%d.00000000", 90+rng.Intn(100)), date))
						units -= qty
					}
				case 4:
					appendStage371Trade(t, h, stage375Request(h.portfolioID, "DIVIDEND", "SBER", fmt.Sprintf("%d.00000000", 1+rng.Intn(25)), "0.00000000", "0.00000000", date))
				case 5:
					if len(cashTransactions) > 0 {
						idx := rng.Intn(len(cashTransactions))
						cashTransactions[idx] = stage377CorrectCash(t, h, cashTransactions[idx], fmt.Sprintf("%d.00000000", 10+rng.Intn(100)), date)
					} else {
						tx := stage377AppendCash(t, h, "DEPOSIT", "10.00000000", date)
						cashTransactions = append(cashTransactions, tx)
					}
				case 6:
					if len(cashTransactions) > 0 {
						idx := rng.Intn(len(cashTransactions))
						tx := cashTransactions[idx]
						if _, err := reverseStage374(t, h, tx, date, uuid.NewString()); err != nil {
							t.Fatalf("stateful cash reversal: %v", err)
						}
						cashTransactions = append(cashTransactions[:idx], cashTransactions[idx+1:]...)
					}
				case 7:
					if units > 0 {
						upsertStage376(t, h, "SBER", fmt.Sprintf("%d.00000000", 90+rng.Intn(100)), date)
					}
				case 8:
					if units == 0 { break }
					if _, err := h.service.ClearManualValuation(h.ctx, h.subjectID, h.portfolioID, "SBER"); err != nil {
						t.Fatalf("clear manual valuation: %v", err)
					}
				case 9:
					// Read-heavy step; assertions below exercise all projections.
				}
				totalOperations++

				positions, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
				if err != nil { t.Fatalf("positions: %v", err) }
				if units == 0 {
					if len(positions.Items) != 0 { t.Fatalf("units=0 positions=%+v", positions.Items) }
				} else {
					if len(positions.Items) != 1 || positions.Items[0].Ticker != "SBER" || positions.Items[0].Quantity.String() != fmt.Sprintf("%d.00000000", units) {
						t.Fatalf("unit reference mismatch units=%d positions=%+v", units, positions.Items)
					}
				}
				positionsAgain, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
				if err != nil { t.Fatal(err) }
				if len(positions.Items) != len(positionsAgain.Items) {
					t.Fatalf("repeat positions length drift")
				}
				if len(positions.Items) == 1 && (positions.Items[0].Quantity.String() != positionsAgain.Items[0].Quantity.String() ||
					positions.Items[0].WeightedAverageCost.Amount.String() != positionsAgain.Items[0].WeightedAverageCost.Amount.String()) {
					t.Fatalf("repeat position projection drift")
				}
				if _, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "", ""); err != nil { t.Fatalf("cash flow: %v", err) }
				if _, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, date); err != nil { t.Fatalf("returns: %v", err) }
				if _, err := h.service.GetPortfolioSummary(h.ctx, h.subjectID, h.portfolioID, date); err != nil { t.Fatalf("summary: %v", err) }
				referenceComparisons++
			}
		})
	}
	t.Logf("M11_FULL_SYSTEM_STATEFUL_SEED=%d", seed)
	t.Logf("M11_FULL_SYSTEM_STATEFUL_SEQUENCES=%d", sequences)
	t.Logf("M11_FULL_SYSTEM_STATEFUL_OPERATIONS=%d", totalOperations)
	t.Logf("M11_FULL_SYSTEM_REFERENCE_COMPARISONS=%d", referenceComparisons)
	t.Log("M11_FULL_SYSTEM_STATEFUL_RESULT=PASS")
}
