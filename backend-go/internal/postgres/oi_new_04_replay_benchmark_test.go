package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
)

var (
	benchmarkSnapshotStates []replayWorkingState
	benchmarkSnapshotDigest string
)

// BenchmarkSnapshotTimelineBounded measures the existing pure replay/snapshot
// calculation. The 5k cap is a production invariant, never bypassed here.
func BenchmarkSnapshotTimelineBounded(b *testing.B) {
	for _, count := range []int{1_000, replayBoundRawRows} {
		b.Run(fmt.Sprintf("ledger-rows=%d", count), func(b *testing.B) {
			if count > 1_000 && os.Getenv("OPENINVEST_BENCH_LARGE") != "1" {
				b.Skip("set OPENINVEST_BENCH_LARGE=1 for the 5k admitted replay bound")
			}
			ctx := context.Background()
			epoch := replayEpoch{
				PolicyVersion: replayPolicyVersion, ActivationGeneration: 1, EpochGeneration: 1,
				PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
				Positions: map[string]replayPosition{}, Financial: zeroReplayFinancialState(),
			}
			dates := make([]string, 28)
			for i := range dates {
				dates[i] = fmt.Sprintf("2026-01-%02d", i+1)
			}
			raw := make([]boundedLedgerRow, count)
			one := decimal.Must("1.00000000")
			for i := range raw {
				raw[i] = testReplayRaw(fmt.Sprintf("entry-%06d", i), fmt.Sprintf("tx-%06d", i), int64(i+1), 1, nil, nil, dates[i%len(dates)])
				raw[i].TransactionType = "DEPOSIT"
				raw[i].GrossAmount = one
			}
			canonical, err := canonicalizeBoundedLedger(ctx, raw)
			if err != nil {
				b.Fatal(err)
			}
			if states, err := replayStatesForDates(ctx, epoch, canonical, raw, dates); err != nil || len(states) != len(dates) {
				b.Fatalf("invalid snapshot fixture: states=%d err=%v", len(states), err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				states, err := replayStatesForDates(ctx, epoch, canonical, raw, dates)
				if err != nil || len(states) != len(dates) {
					b.Fatalf("snapshot replay failed: states=%d err=%v", len(states), err)
				}
				benchmarkSnapshotStates = states
			}
		})
	}
}

// BenchmarkReplaySnapshotStateDigest is an isolated stress measurement of the
// canonical deterministic snapshot fingerprint; 10k/100k positions are NOT
// an admitted ledger replay operation and make no production capacity claim.
func BenchmarkReplaySnapshotStateDigest(b *testing.B) {
	for _, count := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("synthetic-positions=%d", count), func(b *testing.B) {
			if count > 1_000 && os.Getenv("OPENINVEST_BENCH_LARGE") != "1" {
				b.Skip("set OPENINVEST_BENCH_LARGE=1 for 10k/100k synthetic stress")
			}
			one := decimal.Must("1.00000000")
			hundred := decimal.Must("100.00000000")
			epoch := replayEpoch{
				PolicyVersion: replayPolicyVersion, ActivationGeneration: 1, EpochGeneration: 1,
				PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
			}
			state := replayWorkingState{Positions: make(map[string]replayPosition, count), Financial: zeroReplayFinancialState()}
			for i := 0; i < count; i++ {
				asset := fmt.Sprintf("synthetic-asset-%06d", i)
				state.Positions[asset] = replayPosition{
					AssetID: asset, AssetType: "stock", PositionGeneration: 1,
					State: position.State{Quantity: one, WeightedAverageCost: hundred, AcquisitionBasis: hundred, Open: true},
				}
			}
			if digest := replaySourceStateDigest(epoch, state); len(digest) != 64 {
				b.Fatalf("invalid SHA-256 fingerprint: length=%d", len(digest))
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkSnapshotDigest = replaySourceStateDigest(epoch, state)
			}
		})
	}
}
