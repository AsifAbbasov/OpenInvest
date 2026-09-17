package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
)

func TestOINew04SourceStateDigestIsDeterministicAndSensitive(t *testing.T) {
	epoch := replayEpoch{
		PortfolioID:             "00000000-0000-4000-8000-000000000001",
		PolicyVersion:           replayPolicyVersion,
		ActivationGeneration:    3,
		EpochGeneration:         7,
		BoundaryLogicalSequence: 41,
		BuildRawLedgerWatermark: 47,
		PositionMethodology:     replayPositionMethodology,
		SnapshotMethodology:     replaySnapshotMethodology,
		Positions:               map[string]replayPosition{},
		Financial:               zeroReplayFinancialState(),
	}
	positionA := replayPosition{
		AssetID: "00000000-0000-4000-8000-00000000a001", AssetType: "stock", PositionGeneration: 2,
		State: mustReplayPositionState(t, "3.00000000", "101.25000000"),
	}
	positionB := replayPosition{
		AssetID: "00000000-0000-4000-8000-00000000b001", AssetType: "bond", PositionGeneration: 9,
		State: mustReplayPositionState(t, "2.00000000", "1000.00000000"),
	}
	first := cloneReplayWorkingState(epoch)
	first.Positions[positionA.AssetID] = positionA
	first.Positions[positionB.AssetID] = positionB
	second := cloneReplayWorkingState(epoch)
	second.Positions[positionB.AssetID] = positionB
	second.Positions[positionA.AssetID] = positionA

	digestA := replaySourceStateDigest(epoch, first)
	digestB := replaySourceStateDigest(epoch, second)
	if digestA != digestB {
		t.Fatalf("digest depends on map insertion order: %s != %s", digestA, digestB)
	}

	changed := cloneReplayState(first)
	changed.Financial.InvestedCapital = decimal.Must("1.00000000")
	if digestA == replaySourceStateDigest(epoch, changed) {
		t.Fatal("one-field financial mutation did not change source-state digest")
	}
}

func TestOINew04PartitionBoundaryKeepsRevisionFamilyTogether(t *testing.T) {
	raw := []boundedLedgerRow{
		testReplayRaw("a1", "tx-a", 1, 1, nil, nil, "2026-01-01"),
		testReplayRaw("b1", "tx-b", 2, 1, nil, nil, "2026-01-01"),
		testReplayRaw("a2", "tx-a", 3, 2, oiNew04StringPointer("a1"), nil, "2026-01-01"),
	}
	prefix, mutable, boundary, _, err := partitionReplayBoundary(raw, 1)
	if err != nil {
		t.Fatalf("partition: %v", err)
	}
	if len(prefix) != 0 || len(mutable) != 3 || boundary != 0 {
		t.Fatalf("family split was not pulled back: prefix=%d mutable=%d boundary=%d", len(prefix), len(mutable), boundary)
	}
}

func TestOINew04SameDaySuffixAfterBoundaryIsApplied(t *testing.T) {
	boundaryDate := "2026-01-10"
	epoch := replayEpoch{
		PolicyVersion: replayPolicyVersion, ActivationGeneration: 1, EpochGeneration: 1,
		BoundaryTradeDate: &boundaryDate, BoundaryLogicalSequence: 1, BuildRawLedgerWatermark: 1,
		PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
		Positions: map[string]replayPosition{}, Financial: zeroReplayFinancialState(),
	}
	row := testReplayRaw("d2", "tx-deposit", 2, 1, nil, nil, boundaryDate)
	row.TransactionType = "DEPOSIT"
	row.GrossAmount = decimal.Must("25.00000000")
	canonical, err := canonicalizeBoundedLedger(context.Background(), []boundedLedgerRow{row})
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	state, err := replayStateAtDate(context.Background(), epoch, canonical, []boundedLedgerRow{row}, boundaryDate)
	if err != nil {
		t.Fatalf("replay same-day suffix: %v", err)
	}
	if got := state.Financial.Amounts.Deposits.String(); got != "25.00000000" {
		t.Fatalf("same-day row after boundary was skipped: deposits=%s", got)
	}
}

func TestOINew04CompactionAdmitsBoundedCommandWithoutExaminingBeyondB(t *testing.T) {
	existing := make([]boundedLedgerRow, replayBoundRawRows)
	for i := range existing {
		existing[i] = testReplayRaw(
			fmt.Sprintf("entry-%d", i+1), fmt.Sprintf("tx-%d", i+1), int64(i+1), 1, nil, nil, "2026-01-01",
		)
		existing[i].TransactionType = "DEPOSIT"
		existing[i].GrossAmount = decimal.Must("1.00000000")
	}
	epoch := replayEpoch{
		PolicyVersion: replayPolicyVersion, ActivationGeneration: 1, EpochGeneration: 1,
		PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
		Positions: map[string]replayPosition{}, Financial: zeroReplayFinancialState(),
	}
	advanced, mutable, err := compactReplayPlanForAdmission(context.Background(), epoch, existing, 1)
	if err != nil {
		t.Fatalf("compact 5000 + 1: %v", err)
	}
	if len(mutable) != replayBoundRawRows-1 {
		t.Fatalf("mutable suffix=%d want=%d", len(mutable), replayBoundRawRows-1)
	}
	if advanced.BoundaryLogicalSequence != 1 {
		t.Fatalf("advanced boundary=%d want=1", advanced.BoundaryLogicalSequence)
	}
	if _, _, err := compactReplayPlanForAdmission(context.Background(), epoch, existing, replayBoundRawRows+1); !errors.Is(err, ErrRetroactiveReplayWindowExceeded) {
		t.Fatalf("projected command > B error=%v", err)
	}
}

func TestOINew04SnapshotTimelineIsLinearWithoutReversals(t *testing.T) {
	ctx := context.Background()
	instrumentation := &replayInstrumentation{}
	ctx = withReplayInstrumentation(ctx, instrumentation)
	epoch := replayEpoch{
		PolicyVersion: replayPolicyVersion, ActivationGeneration: 1, EpochGeneration: 1,
		PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology,
		Positions: map[string]replayPosition{}, Financial: zeroReplayFinancialState(),
	}
	raw := make([]boundedLedgerRow, 100)
	for i := range raw {
		date := fmt.Sprintf("2026-01-%02d", (i%28)+1)
		raw[i] = testReplayRaw(fmt.Sprintf("e-%d", i), fmt.Sprintf("t-%d", i), int64(i+1), 1, nil, nil, date)
		raw[i].TransactionType = "DEPOSIT"
		raw[i].GrossAmount = decimal.Must("1.00000000")
	}
	canonical, err := canonicalizeBoundedLedger(ctx, raw)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	dates := make([]string, 0, 28)
	for day := 1; day <= 28; day++ {
		dates = append(dates, fmt.Sprintf("2026-01-%02d", day))
	}
	states, err := replayStatesForDates(ctx, epoch, canonical, raw, dates)
	if err != nil {
		t.Fatalf("timeline: %v", err)
	}
	if len(states) != len(dates) {
		t.Fatalf("states=%d dates=%d", len(states), len(dates))
	}
	metrics := replayInstrumentationSnapshotOf(instrumentation)
	if metrics.LedgerRowsApplied != int64(len(canonical)) {
		t.Fatalf("rows applied=%d want=%d; snapshot dates must not multiply replay work", metrics.LedgerRowsApplied, len(canonical))
	}
}

func TestOINew04ReplayHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := canonicalizeBoundedLedger(ctx, []boundedLedgerRow{testReplayRaw("e", "t", 1, 1, nil, nil, "2026-01-01")})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled canonicalization error=%v", err)
	}
}

func zeroReplayFinancialState() replayFinancialState {
	return replayFinancialState{
		Amounts: zeroCashFlowAmounts(), InvestedCapital: decimal.Zero(), InputWatermark: "empty",
	}
}

func mustReplayPositionState(t *testing.T, quantity, price string) position.State {
	t.Helper()
	state, err := position.Apply(position.Empty(), position.Trade{
		Type: "BUY", Quantity: decimal.Must(quantity), UnitPrice: decimal.Must(price),
	})
	if err != nil {
		t.Fatalf("position state: %v", err)
	}
	return state
}

func testReplayRaw(entryID, transactionID string, sequence int64, revision int, prior, reverses *string, tradeDate string) boundedLedgerRow {
	created := time.Date(2026, 1, 1, 0, 0, int(sequence%60), 0, time.UTC)
	return boundedLedgerRow{
		EntryID: entryID, TransactionID: transactionID, TransactionType: "DEPOSIT",
		GrossAmount: decimal.Zero(), Commission: decimal.Zero(), Tax: decimal.Zero(),
		TradeDate: tradeDate, LedgerSequence: sequence, Revision: revision,
		PriorEntryID: prior, ReversesTransactionID: reverses, CreatedAt: created, CreatedAtText: created.Format("2006-01-02 15:04:05+00"),
	}
}

func oiNew04StringPointer(value string) *string { return &value }
