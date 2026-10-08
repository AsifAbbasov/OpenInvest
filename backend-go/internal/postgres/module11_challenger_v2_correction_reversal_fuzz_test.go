package postgres_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func FuzzM11CorrectionReversalStateMachine(f *testing.F) {
	f.Add([]byte{0, 2, 1})
	f.Add([]byte{1, 0, 1})
	f.Add([]byte{2, 0, 2, 1})
	f.Add([]byte{0, 0, 0, 1})

	f.Fuzz(func(t *testing.T, ops []byte) {
		if len(ops) == 0 || len(ops) > 8 {
			t.Skip()
		}
		h := newStage371Harness(t, "M11 V2 correction reversal fuzz")
		original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
		current := original
		currentPrice := "100.00000000"
		reversed := false

		for index, op := range ops {
			switch op % 3 {
			case 0:
				price := 100 + int(op%50)
				priceText := fmt.Sprintf("%d.00000000", price)
				if reversed {
					_, err := correctStage374(t, h, current, "10.00000000", priceText, "2026-01-01", uuid.NewString())
					if !errors.Is(err, verticalslice.ErrTransactionConflict) {
						t.Fatalf("step=%d correction after reversal err=%v", index, err)
					}
				} else {
					next, err := correctStage374(t, h, current, "10.00000000", priceText, "2026-01-01", uuid.NewString())
					if err != nil {
						t.Fatalf("step=%d correction: %v", index, err)
					}
					current = next
					currentPrice = priceText
				}
			case 1:
				_, err := reverseStage374(t, h, current, "2026-01-02", uuid.NewString())
				if reversed {
					if !errors.Is(err, verticalslice.ErrTransactionConflict) {
						t.Fatalf("step=%d repeated reversal err=%v", index, err)
					}
				} else {
					if err != nil {
						t.Fatalf("step=%d reversal: %v", index, err)
					}
					reversed = true
				}
			case 2:
				if !reversed && current.Revision == original.Revision {
					next, err := correctStage374(t, h, current, "10.00000000", "111.00000000", "2026-01-01", uuid.NewString())
					if err != nil {
						t.Fatalf("step=%d establish revision: %v", index, err)
					}
					current = next
					currentPrice = "111.00000000"
				}
				_, err := correctStage374(t, h, original, "10.00000000", "123.00000000", "2026-01-01", uuid.NewString())
				if !errors.Is(err, verticalslice.ErrTransactionConflict) {
					t.Fatalf("step=%d stale correction err=%v", index, err)
				}
			}

			projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
			if err != nil {
				t.Fatalf("step=%d positions: %v", index, err)
			}
			if reversed {
				if len(projection.Items) != 0 {
					t.Fatalf("step=%d reversed transaction resurrected position=%+v", index, projection.Items)
				}
			} else {
				if len(projection.Items) != 1 ||
					projection.Items[0].Quantity.String() != "10.00000000" ||
					projection.Items[0].WeightedAverageCost.Amount.String() != currentPrice {
					t.Fatalf("step=%d active state drift current=%+v position=%+v", index, current, projection.Items)
				}
			}
		}
	})
}
