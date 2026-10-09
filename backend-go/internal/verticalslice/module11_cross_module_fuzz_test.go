package verticalslice

import (
	"fmt"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type m11FixedClock struct{}

func (m11FixedClock) Now() time.Time {
	return time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
}

func FuzzM11CommandIdentity(f *testing.F) {
	f.Add("principal-a", "/api/v1/portfolios/a/transactions", "m11-command-key-0000001", "Portfolio")
	f.Add("principal-b", "/api/v1/portfolios/b/transactions", "m11-command-key-0000002", "Other")
	f.Fuzz(func(t *testing.T, principal, path, key, name string) {
		if len(principal)+len(path)+len(key)+len(name) > 4096 {
			t.Skip()
		}
		service := NewService(nil, m11FixedClock{})
		payload := CreatePortfolioRequest{Name: name, BaseCurrency: RUB}
		first, err := service.command(RequestContext{RequestID: "r1", TraceID: "t1"}, principal, key, path, payload)
		if err != nil {
			t.Fatalf("first command: %v", err)
		}
		second, err := service.command(RequestContext{RequestID: "r2", TraceID: "t2"}, principal, key, path, payload)
		if err != nil {
			t.Fatalf("second command: %v", err)
		}
		if first.RequestHash != second.RequestHash || first.SubjectID != second.SubjectID ||
			first.RequestPath != second.RequestPath || first.IdempotencyKey != second.IdempotencyKey {
			t.Fatalf("identical command identity drift: first=%+v second=%+v", first, second)
		}
		if first.RequestID == second.RequestID || first.TraceID == second.TraceID {
			t.Fatalf("request/trace metadata unexpectedly collapsed")
		}
	})
}

func FuzzM11DecimalCashXIRRPipeline(f *testing.F) {
	f.Add(uint16(100), uint16(10), uint16(365))
	f.Add(uint16(1000), uint16(100), uint16(180))
	f.Add(uint16(1), uint16(1), uint16(1))
	f.Fuzz(func(t *testing.T, contribution, gain, dayGap uint16) {
		contribution = contribution%10000 + 1
		gain = gain%5000 + 1
		dayGap = dayGap%1000 + 1
		start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, int(dayGap))
		contributionAmount := decimal.Must(fmt.Sprintf("-%d.00000000", contribution))
		terminal := Money{
			Amount:   decimal.Must(fmt.Sprintf("%d.00000000", int(contribution)+int(gain))),
			Currency: RUB,
		}
		flows := []PortfolioReturnCashFlow{{
			Date:   start.Format("2006-01-02"),
			Amount: contributionAmount,
		}}
		first, err := BuildPortfolioReturnProjection("m11-fuzz", end.Format("2006-01-02"), flows, &terminal, true)
		if err != nil {
			t.Fatalf("first projection: %v", err)
		}
		second, err := BuildPortfolioReturnProjection("m11-fuzz", end.Format("2006-01-02"), flows, &terminal, true)
		if err != nil {
			t.Fatalf("second projection: %v", err)
		}
		if first.Status != second.Status || first.Reason != second.Reason ||
			(first.XIRR == nil) != (second.XIRR == nil) {
			t.Fatalf("projection nondeterminism: first=%+v second=%+v", first, second)
		}
		if first.XIRR != nil && first.XIRR.String() != second.XIRR.String() {
			t.Fatalf("XIRR nondeterminism: %s != %s", first.XIRR.String(), second.XIRR.String())
		}
	})
}
