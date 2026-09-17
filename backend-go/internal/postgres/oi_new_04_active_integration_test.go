package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type oiNew04FixedClock struct{ now time.Time }

func (c oiNew04FixedClock) Now() time.Time { return c.now }

type oiNew04ActiveHarness struct {
	ctx        context.Context
	owner      *Store
	runtime    *Store
	ownerURL   string
	runtimeURL string
	clock      oiNew04FixedClock
}

func TestOINew04ActiveCPrimeIndependentReviewSuite(t *testing.T) {
	if os.Getenv("OPENINVEST_OI_NEW_04_ACTIVE_R2_TESTS") != "1" {
		t.Skip("OPENINVEST_OI_NEW_04_ACTIVE_R2_TESTS is not enabled")
	}
	h := newOINew04ActiveHarness(t)

	legacyService := verticalslice.NewService(h.owner, h.clock)

	// Build H >= 10,000 through the canonical pre-activation R0 service path. The import API is
	// intentionally used in 100-row batches (its frozen limit) so the history is real command/ledger
	// state rather than a SQL-only synthetic fixture.
	bigSubject := uuid.NewString()
	bigPortfolio := oiNew04CreatePortfolio(t, h.ctx, legacyService, bigSubject, "OI-NEW-04 H>=10000")
	oiNew04AppendDepositHistory(t, h.ctx, legacyService, bigSubject, bigPortfolio.ID, 10000)
	if got := oiNew04LedgerCount(t, h.owner.db, bigPortfolio.ID); got < 10000 {
		t.Fatalf("large-history fixture rows=%d want >=10000", got)
	}
	t.Log("H_GE_10000_FIXTURE=PASS")

	// Build a complete legacy oracle before activation. The same command program is replayed through
	// ACTIVE C' after activation and latest snapshot versions are compared with exact Decimal text.
	oracleSubject := uuid.NewString()
	oraclePortfolio := oiNew04CreatePortfolio(t, h.ctx, legacyService, oracleSubject, "OI-NEW-04 legacy oracle")
	oracleResult := oiNew04RunFinancialScenario(t, h.ctx, legacyService, oracleSubject, oraclePortfolio.ID)
	oracleFingerprints := oiNew04SnapshotFingerprints(t, h.owner.db, oraclePortfolio.ID)
	if len(oracleFingerprints) == 0 {
		t.Fatal("legacy oracle produced no snapshots")
	}

	// PREACTIVE must not become serving merely because the generation row exists.
	generationN, err := h.owner.AllocateReplayGeneration(h.ctx, h.clock.now)
	if err != nil {
		t.Fatalf("allocate generation N: %v", err)
	}
	state := oiNew04CurrentPolicy(t, h.owner, h.ctx)
	if state.Active || state.BlockedAfterInvalidation {
		t.Fatalf("PREACTIVE generation unexpectedly serving: %+v", state)
	}
	t.Log("PREACTIVE_NOT_ACTIVE=PASS")

	// Provisional then final backfill: finalization must supersede provisional immutable evidence.
	provisional, err := h.owner.ProvisionalReplayBackfill(h.ctx, generationN.ActivationGeneration, 4999, h.clock.now)
	if err != nil {
		t.Fatalf("provisional backfill N: %v", err)
	}
	if len(provisional.EpochIDs) == 0 {
		t.Fatal("provisional backfill produced no epochs")
	}
	beforeFinalMax := oiNew04MaxEpochGeneration(t, h.owner.db, bigPortfolio.ID)
	_, manifestN, err := h.owner.FinalizeReplayGeneration(h.ctx, generationN.ActivationGeneration, 4999, h.clock.now)
	if err != nil {
		t.Fatalf("finalize generation N: %v", err)
	}
	afterFinalMax := oiNew04MaxEpochGeneration(t, h.owner.db, bigPortfolio.ID)
	if afterFinalMax <= beforeFinalMax {
		t.Fatalf("final epoch did not supersede provisional epoch: before=%d after=%d", beforeFinalMax, afterFinalMax)
	}
	t.Log("PROVISIONAL_SUPERSEDED_BY_FINAL=PASS")
	if err := h.owner.ActivateReplayGeneration(h.ctx, generationN.ActivationGeneration, manifestN.SHA256, h.clock.now); err != nil {
		t.Fatalf("activate generation N: %v", err)
	}
	state = oiNew04CurrentPolicy(t, h.owner, h.ctx)
	if !state.Active || state.ActivationGeneration != generationN.ActivationGeneration {
		t.Fatalf("generation N not current after activation: %+v", state)
	}
	t.Log("ACTIVATION_N=PASS")

	// R2 application store is opened only after the actual database grant profile is R2.
	runtime, err := OpenRuntimeWithCapability(h.runtimeURL, RuntimeCapabilityR2)
	if err != nil {
		t.Fatalf("open real R2 runtime: %v", err)
	}
	h.runtime = runtime
	t.Cleanup(func() { _ = runtime.Close() })
	activeService := verticalslice.NewService(runtime, h.clock)
	if err := activeService.ValidateRuntimeIntegrity(h.ctx); err != nil {
		t.Fatalf("R2 active startup integrity: %v", err)
	}

	// H>=10000 + 4999 mutable suffix: first active append reaches exactly B=5000 without any H load.
	firstBoundInstrumentation := &replayInstrumentation{}
	firstBoundCtx := withReplayInstrumentation(h.ctx, firstBoundInstrumentation)
	oiNew04AppendCashReplay(t, firstBoundCtx, activeService, bigSubject, bigPortfolio.ID, "DEPOSIT", "10001.00000000", "2025-01-02", uuid.NewString())
	firstBound := replayInstrumentationSnapshotOf(firstBoundInstrumentation)
	if firstBound.FullHistoryLoads != 0 || firstBound.BoundedEngineExecutions != 1 || firstBound.LegacyEngineExecutions != 0 {
		t.Fatalf("4999->5000 instrumentation=%+v", firstBound)
	}
	if suffix := oiNew04CurrentMutableSuffixCount(t, h.owner.db, bigPortfolio.ID, generationN.ActivationGeneration); suffix > replayBoundRawRows {
		t.Fatalf("active suffix exceeded B after 4999 case: %d", suffix)
	}
	t.Logf("BOUND_4999=PASS FullHistoryLoads=%d HistoryRowsLoaded=%d HistoryRowsExamined=%d LedgerRowsApplied=%d SnapshotWrites=%d EpochWrites=%d",
		firstBound.FullHistoryLoads, firstBound.HistoryRowsLoaded, firstBound.HistoryRowsExamined, firstBound.LedgerRowsApplied, firstBound.SnapshotWrites, firstBound.EpochWrites)

	// Existing suffix is now 5000. A semantically valid append must compact/freeze a closed family
	// and remain bounded, never falling back to whole history.
	secondBoundInstrumentation := &replayInstrumentation{}
	secondBoundCtx := withReplayInstrumentation(h.ctx, secondBoundInstrumentation)
	oiNew04AppendCashReplay(t, secondBoundCtx, activeService, bigSubject, bigPortfolio.ID, "DEPOSIT", "10002.00000000", "2025-01-03", uuid.NewString())
	secondBound := replayInstrumentationSnapshotOf(secondBoundInstrumentation)
	if secondBound.FullHistoryLoads != 0 || secondBound.BoundedEngineExecutions != 1 || secondBound.LegacyEngineExecutions != 0 {
		t.Fatalf("5000 compaction instrumentation=%+v", secondBound)
	}
	if suffix := oiNew04CurrentMutableSuffixCount(t, h.owner.db, bigPortfolio.ID, generationN.ActivationGeneration); suffix > replayBoundRawRows {
		t.Fatalf("active suffix exceeded B after 5000 case: %d", suffix)
	}
	t.Log("BOUND_5000=PASS")
	t.Log("H_GE_10000_ACTIVE_BOUNDED=PASS")

	// A projected 5001-row command is rejected by the bounded admission planner. It may inspect only
	// the existing <=B suffix; it must never reconstruct H.
	tooLargeInstrumentation := &replayInstrumentation{}
	tooLargeCtx := withReplayInstrumentation(h.ctx, tooLargeInstrumentation)
	tx, err := runtime.db.BeginTx(tooLargeCtx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin 5001 admission tx: %v", err)
	}
	if err := lockPortfolioTx(tooLargeCtx, tx, bigSubject, bigPortfolio.ID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("lock 5001 portfolio: %v", err)
	}
	_, err = runtime.prepareFinancialReplayPlanTx(tooLargeCtx, tx, bigPortfolio.ID, replayMutationImpact{Kind: "import", RawRows: replayBoundRawRows + 1}, h.clock.now)
	_ = tx.Rollback()
	if !errors.Is(err, ErrRetroactiveReplayWindowExceeded) {
		t.Fatalf("5001 admission error=%v want %v", err, ErrRetroactiveReplayWindowExceeded)
	}
	tooLarge := replayInstrumentationSnapshotOf(tooLargeInstrumentation)
	if tooLarge.FullHistoryLoads != 0 {
		t.Fatalf("5001 rejection loaded full history: %+v", tooLarge)
	}
	t.Log("BOUND_5001_FAIL_CLOSED=PASS")

	// All four public financial mutation paths must execute the bounded engine and never legacy.
	pathsSubject := uuid.NewString()
	pathsPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, pathsSubject, "OI-NEW-04 four paths")
	if oiNew04CurrentGenerationEpochCount(t, h.owner.db, pathsPortfolio.ID, generationN.ActivationGeneration) != 1 {
		t.Fatal("new post-activation portfolio did not receive exactly one genesis epoch")
	}
	t.Log("NEW_POST_ACTIVATION_PORTFOLIO_GENESIS=PASS")
	appendInstrumentation := &replayInstrumentation{}
	buy := oiNew04AppendTradeReplay(t, withReplayInstrumentation(h.ctx, appendInstrumentation), activeService, pathsSubject, pathsPortfolio.ID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01", uuid.NewString())
	oiNew04AssertBoundedOnly(t, "append", appendInstrumentation)
	t.Log("ACTIVE_APPEND_C_PRIME=PASS")

	importInstrumentation := &replayInstrumentation{}
	oiNew04ImportDepositsReplay(t, withReplayInstrumentation(h.ctx, importInstrumentation), activeService, pathsSubject, pathsPortfolio.ID, 1, 700000)
	oiNew04AssertBoundedOnly(t, "import", importInstrumentation)
	t.Log("ACTIVE_IMPORT_C_PRIME=PASS")

	correctInstrumentation := &replayInstrumentation{}
	corrected := oiNew04CorrectTradeReplay(t, withReplayInstrumentation(h.ctx, correctInstrumentation), activeService, pathsSubject, pathsPortfolio.ID, buy, "10.00000000", "110.00000000", "2026-01-01", uuid.NewString())
	if corrected.Revision != 2 {
		t.Fatalf("active correction revision=%d want2", corrected.Revision)
	}
	oiNew04AssertBoundedOnly(t, "correct", correctInstrumentation)
	t.Log("ACTIVE_CORRECT_C_PRIME=PASS")

	reverseInstrumentation := &replayInstrumentation{}
	oiNew04ReverseReplay(t, withReplayInstrumentation(h.ctx, reverseInstrumentation), activeService, pathsSubject, pathsPortfolio.ID, corrected, "2026-01-02", uuid.NewString())
	oiNew04AssertBoundedOnly(t, "reverse", reverseInstrumentation)
	t.Log("ACTIVE_REVERSE_C_PRIME=PASS")
	t.Log("ACTIVE_LEGACY_ENGINE_CALLS=0")

	// Exact legacy oracle: same command program, same deterministic clock, different portfolio id.
	activeOracleSubject := uuid.NewString()
	activeOraclePortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, activeOracleSubject, "OI-NEW-04 active oracle")
	activeResult := oiNew04RunFinancialScenario(t, h.ctx, activeService, activeOracleSubject, activeOraclePortfolio.ID)
	activeFingerprints := oiNew04SnapshotFingerprints(t, h.owner.db, activeOraclePortfolio.ID)
	if strings.Join(oracleFingerprints, "\n") != strings.Join(activeFingerprints, "\n") {
		t.Fatalf("legacy/active snapshot fingerprints differ\nlegacy:\n%s\nactive:\n%s", strings.Join(oracleFingerprints, "\n"), strings.Join(activeFingerprints, "\n"))
	}
	if oracleResult.Rev2SameDate.Revision != 2 || oracleResult.Rev3SameDate.Revision != 3 || oracleResult.Rev2MovedLater.Revision != 2 || oracleResult.Rev2MovedEarly.Revision != 2 ||
		activeResult.Rev2SameDate.Revision != 2 || activeResult.Rev3SameDate.Revision != 3 || activeResult.Rev2MovedLater.Revision != 2 || activeResult.Rev2MovedEarly.Revision != 2 ||
		oracleResult.Reversal.Status != "REVERSED" || activeResult.Reversal.Status != "REVERSED" {
		t.Fatal("revision/reversal equivalence program did not execute expected chain")
	}
	t.Log("REAL_REV2_EQUIVALENCE=PASS")
	t.Log("REAL_REV3_EQUIVALENCE=PASS")
	t.Log("REAL_CORRECTION_EARLIER_EQUIVALENCE=PASS")
	t.Log("REAL_CORRECTION_LATER_EQUIVALENCE=PASS")
	t.Log("REAL_REVERSAL_EQUIVALENCE=PASS")

	// Mandatory non-invertible WAC proof through the real production reversal path.
	wacSubject := uuid.NewString()
	wacPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, wacSubject, "OI-NEW-04 WAC")
	firstBuy := oiNew04AppendTradeReplay(t, h.ctx, activeService, wacSubject, wacPortfolio.ID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-02-01", uuid.NewString())
	oiNew04AppendTradeReplay(t, h.ctx, activeService, wacSubject, wacPortfolio.ID, "BUY", "SBER", "10.00000000", "200.00000000", "2026-02-02", uuid.NewString())
	oiNew04AppendTradeReplay(t, h.ctx, activeService, wacSubject, wacPortfolio.ID, "SELL", "SBER", "5.00000000", "999.00000000", "2026-02-03", uuid.NewString())
	oiNew04ReverseReplay(t, h.ctx, activeService, wacSubject, wacPortfolio.ID, firstBuy, "2026-02-04", uuid.NewString())
	projection, err := activeService.GetPortfolioPositions(h.ctx, wacSubject, wacPortfolio.ID, "2026-02-04")
	if err != nil {
		t.Fatalf("non-invertible WAC projection: %v", err)
	}
	if len(projection.Items) != 1 {
		t.Fatalf("non-invertible WAC projection items=%d want1", len(projection.Items))
	}
	item := projection.Items[0]
	if item.Ticker != "SBER" || item.Quantity.String() != "5.00000000" || item.WeightedAverageCost.Amount.String() != "200.00000000" || item.AcquisitionBasis.Amount.String() != "1000.00000000" {
		t.Fatalf("non-invertible WAC projection ticker=%s qty=%s wac=%s basis=%s want SBER/5/200/1000", item.Ticker, item.Quantity.String(), item.WeightedAverageCost.Amount.String(), item.AcquisitionBasis.Amount.String())
	}
	if stock := oiNew04LatestSnapshotStock(t, h.owner.db, wacPortfolio.ID, "2026-02-04"); stock != "1000.00000000" {
		t.Fatalf("non-invertible WAC acquisition basis snapshot=%s want1000", stock)
	}
	t.Log("NON_INVERTIBLE_WAC_CASE=PASS")

	// source_state_sha256 must verify identically under different PostgreSQL session TimeZone values.
	utcDigest := oiNew04RecomputeLatestDigestUnderTimeZone(t, h.owner, wacPortfolio.ID, generationN.ActivationGeneration, "UTC")
	bakuDigest := oiNew04RecomputeLatestDigestUnderTimeZone(t, h.owner, wacPortfolio.ID, generationN.ActivationGeneration, "Asia/Baku")
	if utcDigest != bakuDigest {
		t.Fatalf("timezone-dependent source digest utc=%s baku=%s", utcDigest, bakuDigest)
	}
	beforeMutationDigest := utcDigest
	oiNew04AppendCashReplay(t, h.ctx, activeService, wacSubject, wacPortfolio.ID, "DEPOSIT", "123.00000000", "2026-02-05", uuid.NewString())
	afterMutationDigest := oiNew04LatestSourceDigest(t, h.owner.db, wacPortfolio.ID, generationN.ActivationGeneration)
	if beforeMutationDigest == afterMutationDigest {
		t.Fatal("financial mutation did not change source_state_sha256")
	}
	t.Log("SOURCE_STATE_DIGEST_TIMEZONE_INDEPENDENT=YES")

	// Finalization freeze uses the same exclusive advisory lock as Finalize/Activate. A new
	// financial writer must fail closed immediately while S5_FINALIZING holds that lock.
	freezeSubject := uuid.NewString()
	freezePortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, freezeSubject, "OI-NEW-04 freeze")
	freezeBefore := oiNew04LedgerCount(t, h.owner.db, freezePortfolio.ID)
	freezeTx, err := h.owner.db.BeginTx(h.ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := lockReplayFinalizationTx(h.ctx, freezeTx); err != nil {
		_ = freezeTx.Rollback()
		t.Fatalf("hold finalization lock: %v", err)
	}
	_, _, freezeErr := activeService.AppendTransactionWithReplay(h.ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, freezeSubject, uuid.NewString(),
		"/api/v1/portfolios/"+freezePortfolio.ID+"/transactions", oiNew04CashRequest(freezePortfolio.ID, "DEPOSIT", "222.00000000", "2026-02-06"), oiNew04TransactionArtifact)
	_ = freezeTx.Rollback()
	if !errors.Is(freezeErr, ErrReplayStateStale) || oiNew04LedgerCount(t, h.owner.db, freezePortfolio.ID) != freezeBefore {
		t.Fatalf("finalization freeze writer result err=%v ledgerBefore=%d ledgerAfter=%d", freezeErr, freezeBefore, oiNew04LedgerCount(t, h.owner.db, freezePortfolio.ID))
	}
	t.Log("FINALIZATION_FREEZE=PASS")

	// Completed ACTIVE command replay must return persisted artifact before either engine executes.
	idemSubject := uuid.NewString()
	idemPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, idemSubject, "OI-NEW-04 idempotency")
	idemReq := oiNew04CashRequest(idemPortfolio.ID, "DEPOSIT", "321.00000000", "2026-03-01")
	idemKey := uuid.NewString()
	idemPath := "/api/v1/portfolios/" + idemPortfolio.ID + "/transactions"
	firstArtifactBody := []byte("oi-new-04-persisted-artifact")
	_, firstArtifact, err := activeService.AppendTransactionWithReplay(h.ctx, verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "oi-new-04-idem-1"}, idemSubject, idemKey, idemPath, idemReq,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{StatusCode: 201, Body: firstArtifactBody, RequestID: uuid.NewString(), TraceID: "persisted-trace"}, nil
		})
	if err != nil {
		t.Fatalf("first active idempotent command: %v", err)
	}
	beforeCounts := oiNew04StateCounts(t, h.owner.db, idemPortfolio.ID, idemKey)
	duplicateInstrumentation := &replayInstrumentation{}
	_, duplicateArtifact, err := activeService.AppendTransactionWithReplay(withReplayInstrumentation(h.ctx, duplicateInstrumentation), verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "oi-new-04-idem-2"}, idemSubject, idemKey, idemPath, idemReq,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{}, errors.New("duplicate must not rebuild artifact")
		})
	if err != nil {
		t.Fatalf("duplicate active idempotent command: %v", err)
	}
	afterCounts := oiNew04StateCounts(t, h.owner.db, idemPortfolio.ID, idemKey)
	if beforeCounts != afterCounts || string(duplicateArtifact.Body) != string(firstArtifact.Body) {
		t.Fatalf("active duplicate changed state before=%+v after=%+v artifact=%q/%q", beforeCounts, afterCounts, duplicateArtifact.Body, firstArtifact.Body)
	}
	duplicateCounters := replayInstrumentationSnapshotOf(duplicateInstrumentation)
	if duplicateCounters.BoundedEngineExecutions != 0 || duplicateCounters.LegacyEngineExecutions != 0 || duplicateCounters.EpochWrites != 0 || duplicateCounters.SnapshotWrites != 0 {
		t.Fatalf("duplicate executed financial engine: %+v", duplicateCounters)
	}
	t.Log("ACTIVE_IDEMPOTENCY_ZERO_NEW_STATE=PASS")

	// Full-command atomic rollback after an epoch and its child state have been INSERTed in the actual
	// command transaction. The context-local seam then returns an injected error; a separate owner
	// connection must observe zero partial state.
	atomicSubject := uuid.NewString()
	atomicPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, atomicSubject, "OI-NEW-04 rollback")
	atomicKey := uuid.NewString()
	atomicBefore := oiNew04StateCounts(t, h.owner.db, atomicPortfolio.ID, atomicKey)
	atomicInstr := &replayInstrumentation{}
	faultErr := errors.New("oi-new-04 deterministic post-epoch failure")
	faultCtx := withReplayInstrumentation(h.ctx, atomicInstr)
	faultCtx = withReplayTestFault(faultCtx, &replayTestFault{failAfterEpochWrites: 1, err: faultErr})
	_, _, err = activeService.AppendTransactionWithReplay(faultCtx, verticalslice.RequestContext{RequestID: uuid.NewString()}, atomicSubject, atomicKey,
		"/api/v1/portfolios/"+atomicPortfolio.ID+"/transactions", oiNew04CashRequest(atomicPortfolio.ID, "DEPOSIT", "777.00000000", "2026-03-02"), oiNew04TransactionArtifact)
	if !errors.Is(err, faultErr) {
		t.Fatalf("full-command injected failure=%v", err)
	}
	if replayInstrumentationSnapshotOf(atomicInstr).EpochWrites < 1 {
		t.Fatal("atomicity seam did not fire after an epoch write")
	}
	atomicAfter := oiNew04StateCounts(t, h.owner.db, atomicPortfolio.ID, atomicKey)
	if atomicBefore != atomicAfter {
		t.Fatalf("full-command rollback leaked state before=%+v after=%+v", atomicBefore, atomicAfter)
	}
	t.Log("FULL_COMMAND_ROLLBACK=PASS")

	// Cancellation uses the same post-epoch point but cancels the actual command context.
	cancelSubject := uuid.NewString()
	cancelPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, cancelSubject, "OI-NEW-04 cancel")
	cancelKey := uuid.NewString()
	cancelBefore := oiNew04StateCounts(t, h.owner.db, cancelPortfolio.ID, cancelKey)
	baseCancelCtx, cancel := context.WithCancel(h.ctx)
	cancelInstr := &replayInstrumentation{}
	cancelCtx := withReplayInstrumentation(baseCancelCtx, cancelInstr)
	cancelCtx = withReplayTestFault(cancelCtx, &replayTestFault{cancelAfterEpochWrites: 1, cancel: cancel})
	_, _, err = activeService.AppendTransactionWithReplay(cancelCtx, verticalslice.RequestContext{RequestID: uuid.NewString()}, cancelSubject, cancelKey,
		"/api/v1/portfolios/"+cancelPortfolio.ID+"/transactions", oiNew04CashRequest(cancelPortfolio.ID, "DEPOSIT", "888.00000000", "2026-03-03"), oiNew04TransactionArtifact)
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("full-command cancellation error=%v", err)
	}
	if replayInstrumentationSnapshotOf(cancelInstr).EpochWrites < 1 {
		t.Fatal("cancellation seam did not fire after an epoch write")
	}
	cancelAfter := oiNew04StateCounts(t, h.owner.db, cancelPortfolio.ID, cancelKey)
	if cancelBefore != cancelAfter {
		t.Fatalf("full-command cancellation leaked state before=%+v after=%+v", cancelBefore, cancelAfter)
	}
	t.Log("FULL_COMMAND_CANCELLATION=PASS")

	// Same-portfolio ACTIVE R2 concurrency: both commands serialize, epochs stay contiguous/unique,
	// and snapshot versions on the shared date cannot collide.
	concurrentSubject := uuid.NewString()
	concurrentPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, concurrentSubject, "OI-NEW-04 concurrency")
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i, amount := range []string{"901.00000000", "902.00000000"} {
		wg.Add(1)
		go func(i int, amount string) {
			defer wg.Done()
			<-start
			_, _, err := activeService.AppendTransactionWithReplay(h.ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, concurrentSubject, uuid.NewString(),
				"/api/v1/portfolios/"+concurrentPortfolio.ID+"/transactions", oiNew04CashRequest(concurrentPortfolio.ID, "DEPOSIT", amount, "2026-03-04"), oiNew04TransactionArtifact)
			errs <- err
		}(i, amount)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("same-portfolio active command: %v", err)
		}
	}
	oiNew04AssertEpochSequenceContiguous(t, h.owner.db, concurrentPortfolio.ID)
	oiNew04AssertSnapshotVersionsUnique(t, h.owner.db, concurrentPortfolio.ID, "2026-03-04")
	t.Log("ACTIVE_SAME_PORTFOLIO_CONCURRENCY=PASS")

	// Keep two portfolios for N -> INVALIDATED -> N+1 edge cases.
	genesisSubject := uuid.NewString()
	genesisCarry := oiNew04CreatePortfolio(t, h.ctx, activeService, genesisSubject, "OI-NEW-04 genesis carry")
	oldGenerationMax := oiNew04MaxEpochGeneration(t, h.owner.db, genesisCarry.ID)
	missingSubject := uuid.NewString()
	missingNonempty := oiNew04CreatePortfolio(t, h.ctx, activeService, missingSubject, "OI-NEW-04 missing nonempty")
	oiNew04AppendCashReplay(t, h.ctx, activeService, missingSubject, missingNonempty.ID, "DEPOSIT", "10.00000000", "2026-03-05", uuid.NewString())

	if err := h.owner.InvalidateReplayGeneration(h.ctx, generationN.ActivationGeneration, "independent-review lifecycle test", h.clock.now); err != nil {
		t.Fatalf("invalidate N: %v", err)
	}
	state = oiNew04CurrentPolicy(t, h.owner, h.ctx)
	if state.Active || !state.BlockedAfterInvalidation || state.ActivationGeneration != generationN.ActivationGeneration {
		t.Fatalf("invalidated N state=%+v", state)
	}
	t.Log("INVALIDATION_NO_FALLBACK=PASS")

	generationN1, err := h.owner.AllocateReplayGeneration(h.ctx, h.clock.now.Add(time.Minute))
	if err != nil {
		t.Fatalf("allocate N+1: %v", err)
	}
	if generationN1.ActivationGeneration != generationN.ActivationGeneration+1 {
		t.Fatalf("N+1 generation=%d want%d", generationN1.ActivationGeneration, generationN.ActivationGeneration+1)
	}
	state = oiNew04CurrentPolicy(t, h.owner, h.ctx)
	if state.Active || !state.BlockedAfterInvalidation {
		t.Fatalf("PREACTIVE N+1 incorrectly bypassed invalidated N: %+v", state)
	}

	_, err = h.owner.ProvisionalReplayBackfill(h.ctx, generationN1.ActivationGeneration, 4999, h.clock.now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("provisional N+1: %v", err)
	}
	provisionalN1Max := oiNew04MaxEpochGeneration(t, h.owner.db, genesisCarry.ID)
	_, staleManifest, err := h.owner.FinalizeReplayGeneration(h.ctx, generationN1.ActivationGeneration, 4999, h.clock.now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("finalize N+1: %v", err)
	}
	finalN1Max := oiNew04MaxEpochGeneration(t, h.owner.db, genesisCarry.ID)
	if finalN1Max <= provisionalN1Max {
		t.Fatalf("N+1 final did not supersede provisional: provisional=%d final=%d", provisionalN1Max, finalN1Max)
	}

	// Simulate an R0 writer after finalization but before activation. Use a forward-dated row here so
	// the old manifest is invalidated without deliberately making the PREACTIVE generation impossible
	// to re-finalize. The separate post-activation backdated writer case below proves stale-boundary
	// detection and fail-closed behavior.
	staleWriterStore, err := OpenOwnerWithApplicationCapability(h.ownerURL, RuntimeCapabilityR0)
	if err != nil {
		t.Fatalf("open stale R0 writer: %v", err)
	}
	staleWriterService := verticalslice.NewService(staleWriterStore, h.clock)
	oiNew04AppendCashReplay(t, h.ctx, staleWriterService, bigSubject, bigPortfolio.ID, "DEPOSIT", "99999.00000000", "2026-12-31", uuid.NewString())
	_ = staleWriterStore.Close()
	if err := h.owner.ActivateReplayGeneration(h.ctx, generationN1.ActivationGeneration, staleManifest.SHA256, h.clock.now.Add(4*time.Minute)); err == nil {
		t.Fatal("activation unexpectedly accepted stale finalization manifest")
	}
	t.Log("FINALIZATION_MANIFEST_MISMATCH=PASS")
	_, manifestN1, err := h.owner.FinalizeReplayGeneration(h.ctx, generationN1.ActivationGeneration, 4999, h.clock.now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("re-finalize N+1: %v", err)
	}
	if err := h.owner.ActivateReplayGeneration(h.ctx, generationN1.ActivationGeneration, manifestN1.SHA256, h.clock.now.Add(6*time.Minute)); err != nil {
		t.Fatalf("activate N+1: %v", err)
	}
	state = oiNew04CurrentPolicy(t, h.owner, h.ctx)
	if !state.Active || state.ActivationGeneration != generationN1.ActivationGeneration {
		t.Fatalf("N+1 not current: %+v", state)
	}
	t.Log("REACTIVATION_N_PLUS_1=PASS")

	// Defensive genesis: erase only current-generation test fixture state, leaving the old immutable N
	// epoch. The runtime command must allocate MAX(portfolio epoch_generation)+1 and not collide.
	oiNew04DeleteGenerationEpochFixture(t, h.owner.db, genesisCarry.ID, generationN1.ActivationGeneration)
	oiNew04AppendCashReplay(t, h.ctx, activeService, genesisSubject, genesisCarry.ID, "DEPOSIT", "42.00000000", "2026-04-01", uuid.NewString())
	newMin, newMax := oiNew04GenerationEpochRange(t, h.owner.db, genesisCarry.ID, generationN1.ActivationGeneration)
	if newMin <= oldGenerationMax || newMax < newMin {
		t.Fatalf("defensive genesis not portfolio-global monotonic oldMax=%d newRange=%d..%d", oldGenerationMax, newMin, newMax)
	}
	t.Log("GENESIS_EPOCH_GENERATION_FIXED=YES")

	// Missing current-generation epoch with non-empty ledger must fail immediately without H replay.
	oiNew04DeleteGenerationEpochFixture(t, h.owner.db, missingNonempty.ID, generationN1.ActivationGeneration)
	missingInstr := &replayInstrumentation{}
	_, _, err = activeService.AppendTransactionWithReplay(withReplayInstrumentation(h.ctx, missingInstr), verticalslice.RequestContext{RequestID: uuid.NewString()}, missingSubject, uuid.NewString(),
		"/api/v1/portfolios/"+missingNonempty.ID+"/transactions", oiNew04CashRequest(missingNonempty.ID, "DEPOSIT", "11.00000000", "2026-04-02"), oiNew04TransactionArtifact)
	if !errors.Is(err, ErrReplayEpochMissing) {
		t.Fatalf("missing nonempty epoch error=%v", err)
	}
	if replayInstrumentationSnapshotOf(missingInstr).FullHistoryLoads != 0 {
		t.Fatal("missing epoch path performed whole-history reconstruction")
	}
	t.Log("REPLAY_EPOCH_MISSING_NONEMPTY_ZERO_FULL_HISTORY=PASS")

	// Create a genuinely stale legacy write *after* N+1 activation. It crosses the frozen boundary,
	// so the next R2 command must reject it rather than silently absorb a stale R0 writer.
	postActivationStaleStore, err := OpenOwnerWithApplicationCapability(h.ownerURL, RuntimeCapabilityR0)
	if err != nil {
		t.Fatalf("open post-activation stale R0 writer: %v", err)
	}
	postActivationStaleService := verticalslice.NewService(postActivationStaleStore, h.clock)
	oiNew04AppendCashReplay(t, h.ctx, postActivationStaleService, bigSubject, bigPortfolio.ID, "DEPOSIT", "100004.00000000", "2019-01-01", uuid.NewString())
	_ = postActivationStaleStore.Close()

	staleInstr := &replayInstrumentation{}
	_, _, err = activeService.AppendTransactionWithReplay(withReplayInstrumentation(h.ctx, staleInstr), verticalslice.RequestContext{RequestID: uuid.NewString()}, bigSubject, uuid.NewString(),
		"/api/v1/portfolios/"+bigPortfolio.ID+"/transactions", oiNew04CashRequest(bigPortfolio.ID, "DEPOSIT", "100003.00000000", "2026-04-03"), oiNew04TransactionArtifact)
	if !errors.Is(err, ErrReplayStateStale) {
		t.Fatalf("stale legacy writer error=%v want %v", err, ErrReplayStateStale)
	}
	if replayInstrumentationSnapshotOf(staleInstr).FullHistoryLoads != 0 {
		t.Fatal("stale writer detection used full history")
	}
	t.Log("STALE_WRITER_DETECTION=PASS")

	// Close existing runtime before changing grants underneath the session and exercise actual
	// OpenRuntimeWithCapability validation (positive + mismatch + extra/missing privilege cases).
	if err := runtime.Close(); err != nil {
		t.Fatalf("close R2 before profile matrix: %v", err)
	}
	h.runtime = nil
	oiNew04ExerciseRuntimeProfiles(t, h.owner.db, h.runtimeURL)
	t.Log("R0_OPEN_RUNTIME=PASS")
	t.Log("R1_OPEN_RUNTIME=PASS")
	t.Log("R2_OPEN_RUNTIME=PASS")

	t.Log("NO_D_TIMES_H_REPLAY=PASS")
	t.Log("NO_R_TIMES_H_REPLAY=PASS")
}

func newOINew04ActiveHarness(t *testing.T) oiNew04ActiveHarness {
	t.Helper()
	ownerURL := os.Getenv("OPENINVEST_OI_NEW_04_ACTIVE_OWNER_URL")
	runtimeURL := os.Getenv("OPENINVEST_OI_NEW_04_ACTIVE_RUNTIME_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("active OI-NEW-04 database URLs are not set")
	}
	owner, err := OpenOwnerWithApplicationCapability(ownerURL, RuntimeCapabilityR0)
	if err != nil {
		t.Fatalf("open owner: %v", err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	return oiNew04ActiveHarness{
		ctx:        context.Background(),
		owner:      owner,
		ownerURL:   ownerURL,
		runtimeURL: runtimeURL,
		clock:      oiNew04FixedClock{now: time.Date(2026, 9, 16, 8, 0, 0, 123456789, time.UTC)},
	}
}

func oiNew04CreatePortfolio(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, name string) verticalslice.Portfolio {
	t.Helper()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, uuid.NewString(), "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{Name: name, BaseCurrency: verticalslice.RUB})
	if err != nil {
		t.Fatalf("create portfolio %q: %v", name, err)
	}
	return portfolio
}

func oiNew04CashRequest(portfolioID, transactionType, amount, tradeDate string) verticalslice.AppendTransactionRequest {
	money := verticalslice.Money{Amount: decimal.Must(amount), Currency: verticalslice.RUB}
	return verticalslice.AppendTransactionRequest{
		PortfolioID: portfolioID, TransactionType: transactionType, GrossAmount: &money,
		Commission: verticalslice.ZeroMoney(), Tax: verticalslice.ZeroMoney(), TradeDate: tradeDate,
	}
}

func oiNew04TradeRequest(portfolioID, transactionType, ticker, quantity, price, tradeDate string) verticalslice.AppendTransactionRequest {
	q := decimal.Must(quantity)
	p := verticalslice.Money{Amount: decimal.Must(price), Currency: verticalslice.RUB}
	return verticalslice.AppendTransactionRequest{
		PortfolioID: portfolioID, TransactionType: transactionType, Ticker: &ticker, Quantity: &q, UnitPrice: &p,
		Commission: verticalslice.ZeroMoney(), Tax: verticalslice.ZeroMoney(), TradeDate: tradeDate,
	}
}

func oiNew04IncomeRequest(portfolioID, transactionType, ticker, amount, tradeDate string) verticalslice.AppendTransactionRequest {
	m := verticalslice.Money{Amount: decimal.Must(amount), Currency: verticalslice.RUB}
	return verticalslice.AppendTransactionRequest{
		PortfolioID: portfolioID, TransactionType: transactionType, Ticker: &ticker, GrossAmount: &m,
		Commission: verticalslice.ZeroMoney(), Tax: verticalslice.ZeroMoney(), TradeDate: tradeDate,
	}
}

func oiNew04TransactionArtifact(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
	return verticalslice.CommandReplayArtifact{StatusCode: 201, Body: []byte(tx.ID), RequestID: uuid.NewString(), TraceID: "oi-new-04-trace"}, nil
}

func oiNew04AppendCashReplay(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID, typ, amount, date, key string) verticalslice.Transaction {
	t.Helper()
	tx, _, err := service.AppendTransactionWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, key,
		"/api/v1/portfolios/"+portfolioID+"/transactions", oiNew04CashRequest(portfolioID, typ, amount, date), oiNew04TransactionArtifact)
	if err != nil {
		t.Fatalf("append %s %s: %v", typ, amount, err)
	}
	return tx
}

func oiNew04AppendTradeReplay(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID, typ, ticker, quantity, price, date, key string) verticalslice.Transaction {
	t.Helper()
	tx, _, err := service.AppendTransactionWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, key,
		"/api/v1/portfolios/"+portfolioID+"/transactions", oiNew04TradeRequest(portfolioID, typ, ticker, quantity, price, date), oiNew04TransactionArtifact)
	if err != nil {
		t.Fatalf("append %s %s: %v", typ, ticker, err)
	}
	return tx
}

func oiNew04CorrectTradeReplay(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID string, current verticalslice.Transaction, quantity, price, date, key string) verticalslice.Transaction {
	t.Helper()
	request := oiNew04TradeRequest(portfolioID, current.TransactionType, *current.Ticker, quantity, price, date)
	corrected, _, err := service.CorrectTransactionWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, key,
		"/api/v1/portfolios/"+portfolioID+"/transactions/"+current.ID,
		verticalslice.CorrectTransactionRequest{PortfolioID: portfolioID, TransactionID: current.ID, ExpectedRevision: current.Revision, Reason: "OI-NEW-04 equivalence correction", Corrected: request},
		oiNew04TransactionArtifact)
	if err != nil {
		t.Fatalf("correct transaction %s: %v", current.ID, err)
	}
	return corrected
}

func oiNew04ReverseReplay(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID string, current verticalslice.Transaction, effectiveDate, key string) verticalslice.TransactionReversal {
	t.Helper()
	result, _, err := service.ReverseTransactionWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, key,
		"/api/v1/portfolios/"+portfolioID+"/transactions/"+current.ID,
		verticalslice.ReverseTransactionRequest{PortfolioID: portfolioID, TransactionID: current.ID, ExpectedRevision: current.Revision, Reason: "OI-NEW-04 equivalence reversal", EffectiveDate: effectiveDate},
		func(r verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{StatusCode: 200, Body: []byte(r.ReversalTransactionID), RequestID: uuid.NewString(), TraceID: "oi-new-04-trace"}, nil
		})
	if err != nil {
		t.Fatalf("reverse transaction %s: %v", current.ID, err)
	}
	return result
}

func oiNew04ImportDepositsReplay(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID string, count, offset int) {
	t.Helper()
	if count < 1 || count > 100 {
		t.Fatalf("invalid import count %d", count)
	}
	rows := make([]verticalslice.AppendTransactionRequest, 0, count)
	for i := 0; i < count; i++ {
		amount := fmt.Sprintf("%d.00000000", offset+i+1)
		rows = append(rows, oiNew04CashRequest(portfolioID, "DEPOSIT", amount, "2025-01-01"))
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("oi-new-04-import-%s-%d-%d", portfolioID, offset, count)))
	request := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolioID, Transactions: rows, SourceKind: "USER_UPLOADED_FILE",
		SourceAccountLabel: "OI-NEW-04", SourceFileHash: hex.EncodeToString(hash[:]),
	}
	_, _, err := service.AppendImportedTransactionsWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, uuid.NewString(),
		"/api/v1/portfolios/"+portfolioID+"/imports/append", request,
		func(items []verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{StatusCode: 201, Body: []byte(fmt.Sprintf("%d", len(items))), RequestID: uuid.NewString(), TraceID: "oi-new-04-import"}, nil
		})
	if err != nil {
		t.Fatalf("import deposits offset=%d count=%d: %v", offset, count, err)
	}
}

func oiNew04AppendDepositHistory(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID string, count int) {
	t.Helper()
	for offset := 0; offset < count; offset += 100 {
		n := count - offset
		if n > 100 {
			n = 100
		}
		oiNew04ImportDepositsReplay(t, ctx, service, subjectID, portfolioID, n, offset)
	}
}

func oiNew04AssertBoundedOnly(t *testing.T, name string, instrumentation *replayInstrumentation) {
	t.Helper()
	s := replayInstrumentationSnapshotOf(instrumentation)
	if s.BoundedEngineExecutions != 1 || s.LegacyEngineExecutions != 0 || s.FullHistoryLoads != 0 || s.EpochWrites != 1 {
		t.Fatalf("%s did not execute exactly one bounded C' engine: %+v", name, s)
	}
}

type oiNew04ScenarioResult struct {
	Rev2SameDate   verticalslice.Transaction
	Rev3SameDate   verticalslice.Transaction
	Rev2MovedLater verticalslice.Transaction
	Rev2MovedEarly verticalslice.Transaction
	Reversal       verticalslice.TransactionReversal
}

func oiNew04RunFinancialScenario(t *testing.T, ctx context.Context, service *verticalslice.Service, subjectID, portfolioID string) oiNew04ScenarioResult {
	t.Helper()
	oiNew04AppendCashReplay(t, ctx, service, subjectID, portfolioID, "DEPOSIT", "100000.00000000", "2026-01-01", uuid.NewString())
	a := oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-02", uuid.NewString())
	a2 := oiNew04CorrectTradeReplay(t, ctx, service, subjectID, portfolioID, a, "10.00000000", "110.00000000", "2026-01-02", uuid.NewString())
	a3 := oiNew04CorrectTradeReplay(t, ctx, service, subjectID, portfolioID, a2, "10.00000000", "120.00000000", "2026-01-02", uuid.NewString())
	b := oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "BUY", "SBER", "10.00000000", "200.00000000", "2026-01-03", uuid.NewString())
	b2 := oiNew04CorrectTradeReplay(t, ctx, service, subjectID, portfolioID, b, "10.00000000", "210.00000000", "2026-01-04", uuid.NewString())
	bond := oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "BUY", "SU26238RMFS4", "2.00000000", "1000.00000000", "2026-01-03", uuid.NewString())
	bond2 := oiNew04CorrectTradeReplay(t, ctx, service, subjectID, portfolioID, bond, "2.00000000", "1000.00000000", "2026-01-01", uuid.NewString())
	_ = oiNew04CorrectTradeReplay(t, ctx, service, subjectID, portfolioID, bond2, "2.00000000", "1000.00000000", "2026-01-04", uuid.NewString())
	oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "SELL", "SBER", "5.00000000", "300.00000000", "2026-01-05", uuid.NewString())
	income := []verticalslice.AppendTransactionRequest{
		oiNew04IncomeRequest(portfolioID, "DIVIDEND", "SBER", "100.00000000", "2026-01-05"),
		oiNew04IncomeRequest(portfolioID, "COUPON", "SU26238RMFS4", "50.00000000", "2026-01-05"),
		oiNew04CashRequest(portfolioID, "FEE", "10.00000000", "2026-01-05"),
		oiNew04CashRequest(portfolioID, "TAX", "5.00000000", "2026-01-05"),
	}
	for _, req := range income {
		_, _, err := service.AppendTransactionWithReplay(ctx, verticalslice.RequestContext{RequestID: uuid.NewString()}, subjectID, uuid.NewString(), "/api/v1/portfolios/"+portfolioID+"/transactions", req, oiNew04TransactionArtifact)
		if err != nil {
			t.Fatalf("scenario %s: %v", req.TransactionType, err)
		}
	}
	reversal := oiNew04ReverseReplay(t, ctx, service, subjectID, portfolioID, a3, "2026-01-06", uuid.NewString())
	oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "SELL", "SBER", "5.00000000", "300.00000000", "2026-01-07", uuid.NewString())
	oiNew04AppendTradeReplay(t, ctx, service, subjectID, portfolioID, "BUY", "SBER", "2.00000000", "333.33333333", "2026-01-08", uuid.NewString())
	oiNew04AppendCashReplay(t, ctx, service, subjectID, portfolioID, "WITHDRAWAL", "500.00000000", "2026-01-09", uuid.NewString())
	return oiNew04ScenarioResult{Rev2SameDate: a2, Rev3SameDate: a3, Rev2MovedLater: b2, Rev2MovedEarly: bond2, Reversal: reversal}
}

func oiNew04SnapshotFingerprints(t *testing.T, db *sql.DB, portfolioID string) []string {
	t.Helper()
	rows, err := db.Query(`
		SELECT DISTINCT ON (snapshot_date)
			snapshot_date::text, cash_value_amount::text, stock_value_amount::text, bond_value_amount::text,
			invested_capital_amount::text, input_watermark, total_value_amount::text,
			nominal_return_rate::text, real_return_rate::text, methodology_version, snapshot_version
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id=$1::uuid
		ORDER BY snapshot_date, snapshot_version DESC
	`, portfolioID)
	if err != nil {
		t.Fatalf("snapshot fingerprints: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var date, cash, stock, bond, invested, watermark, total, nominal, real, methodology string
		var version int64
		if err := rows.Scan(&date, &cash, &stock, &bond, &invested, &watermark, &total, &nominal, &real, &methodology, &version); err != nil {
			t.Fatalf("scan snapshot fingerprint: %v", err)
		}
		out = append(out, strings.Join([]string{date, cash, stock, bond, invested, watermark, total, nominal, real, methodology, fmt.Sprint(version)}, "|"))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("snapshot fingerprint rows: %v", err)
	}
	return out
}

func oiNew04CurrentPolicy(t *testing.T, store *Store, ctx context.Context) replayPolicyState {
	t.Helper()
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer rollback(tx)
	state, err := currentReplayPolicyTx(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func oiNew04LedgerCount(t *testing.T, db *sql.DB, portfolioID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1::uuid`, portfolioID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func oiNew04MaxEpochGeneration(t *testing.T, db *sql.DB, portfolioID string) int64 {
	t.Helper()
	var n int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(epoch_generation),0) FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid`, portfolioID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func oiNew04CurrentGenerationEpochCount(t *testing.T, db *sql.DB, portfolioID string, generation int64) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid AND policy_version=$2 AND activation_generation=$3`, portfolioID, replayPolicyVersion, generation).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func oiNew04CurrentMutableSuffixCount(t *testing.T, db *sql.DB, portfolioID string, generation int64) int {
	t.Helper()
	var boundary int64
	if err := db.QueryRow(`
		SELECT boundary_logical_sequence FROM analytics.portfolio_replay_epochs
		WHERE portfolio_id=$1::uuid AND policy_version=$2 AND activation_generation=$3
		ORDER BY epoch_generation DESC LIMIT 1
	`, portfolioID, replayPolicyVersion, generation).Scan(&boundary); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1::uuid AND ledger_sequence>$2`, portfolioID, boundary).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func oiNew04LatestReplayPosition(t *testing.T, db *sql.DB, portfolioID string, generation int64, ticker string) (string, string) {
	t.Helper()
	var quantity, wac string
	if err := db.QueryRow(`
		SELECT p.quantity::text, p.weighted_average_cost_amount::text
		FROM analytics.portfolio_replay_epochs e
		JOIN analytics.portfolio_replay_positions p ON p.epoch_id=e.epoch_id
		JOIN investment.assets a ON a.id=p.asset_id
		WHERE e.portfolio_id=$1::uuid AND e.policy_version=$2 AND e.activation_generation=$3 AND a.ticker=$4
		ORDER BY e.epoch_generation DESC LIMIT 1
	`, portfolioID, replayPolicyVersion, generation, ticker).Scan(&quantity, &wac); err != nil {
		t.Fatal(err)
	}
	return quantity, wac
}

func oiNew04LatestSnapshotStock(t *testing.T, db *sql.DB, portfolioID, date string) string {
	t.Helper()
	var value string
	if err := db.QueryRow(`
		SELECT stock_value_amount::text FROM analytics.portfolio_snapshots
		WHERE portfolio_id=$1::uuid AND snapshot_date=$2::date AND methodology_version=$3
		ORDER BY snapshot_version DESC LIMIT 1
	`, portfolioID, date, stage371SnapshotMethodology).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func oiNew04LatestSourceDigest(t *testing.T, db *sql.DB, portfolioID string, generation int64) string {
	t.Helper()
	var digest string
	if err := db.QueryRow(`
		SELECT source_state_sha256 FROM analytics.portfolio_replay_epochs
		WHERE portfolio_id=$1::uuid AND policy_version=$2 AND activation_generation=$3
		ORDER BY epoch_generation DESC LIMIT 1
	`, portfolioID, replayPolicyVersion, generation).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	return digest
}

func oiNew04RecomputeLatestDigestUnderTimeZone(t *testing.T, owner *Store, portfolioID string, generation int64, zone string) string {
	t.Helper()
	ctx := context.Background()
	tx, err := owner.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, `SET LOCAL TIME ZONE '`+strings.ReplaceAll(zone, `'`, `''`)+`'`); err != nil {
		t.Fatal(err)
	}
	policy := replayPolicyState{Active: true, PolicyVersion: replayPolicyVersion, ActivationGeneration: generation, ReplayBoundRawRows: replayBoundRawRows, PositionMethodology: replayPositionMethodology, SnapshotMethodology: replaySnapshotMethodology}
	epoch, err := loadLatestReplayEpochTx(ctx, tx, portfolioID, policy)
	if err != nil {
		t.Fatal(err)
	}
	mutable, err := loadRawLedgerRangeTx(ctx, tx, portfolioID, epoch.BoundaryLogicalSequence, epoch.BuildRawLedgerWatermark, replayBoundRawRows+1)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalizeBoundedLedger(ctx, mutable)
	if err != nil {
		t.Fatal(err)
	}
	state := cloneReplayWorkingState(epoch)
	if maxDate := maxReplayTradeDate(canonical, epoch.BoundaryTradeDate); maxDate != "" {
		state, err = replayStateAtDate(ctx, epoch, canonical, mutable, maxDate)
		if err != nil {
			t.Fatal(err)
		}
	}
	digest := replaySourceStateDigest(epoch, state)
	if digest != epoch.SourceStateSHA256 {
		t.Fatalf("recomputed digest in timezone %s=%s stored=%s", zone, digest, epoch.SourceStateSHA256)
	}
	return digest
}

type oiNew04Counts struct {
	LedgerRows, SnapshotRows, EpochRows, PositionRows, FinancialRows, CommandRows int
	LatestDigest                                                                  string
}

func oiNew04StateCounts(t *testing.T, db *sql.DB, portfolioID, idempotencyKey string) oiNew04Counts {
	t.Helper()
	var c oiNew04Counts
	queries := []struct {
		q string
		p *int
	}{
		{`SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1::uuid`, &c.LedgerRows},
		{`SELECT COUNT(*) FROM analytics.portfolio_snapshots WHERE portfolio_id=$1::uuid`, &c.SnapshotRows},
		{`SELECT COUNT(*) FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid`, &c.EpochRows},
		{`SELECT COUNT(*) FROM analytics.portfolio_replay_positions WHERE epoch_id IN (SELECT epoch_id FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid)`, &c.PositionRows},
		{`SELECT COUNT(*) FROM analytics.portfolio_replay_financial_state WHERE epoch_id IN (SELECT epoch_id FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid)`, &c.FinancialRows},
	}
	for _, item := range queries {
		if err := db.QueryRow(item.q, portfolioID).Scan(item.p); err != nil {
			t.Fatal(err)
		}
	}
	if idempotencyKey != "" {
		if err := db.QueryRow(`SELECT COUNT(*) FROM investment.command_deduplication WHERE idempotency_key=$1`, idempotencyKey).Scan(&c.CommandRows); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.QueryRow(`SELECT COALESCE((SELECT source_state_sha256 FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid ORDER BY epoch_generation DESC LIMIT 1),'')`, portfolioID).Scan(&c.LatestDigest)
	return c
}

func oiNew04AssertEpochSequenceContiguous(t *testing.T, db *sql.DB, portfolioID string) {
	t.Helper()
	rows, err := db.Query(`SELECT epoch_generation FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid ORDER BY epoch_generation`, portfolioID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var values []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		values = append(values, v)
	}
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			t.Fatalf("epoch generation gap/duplicate: %v", values)
		}
	}
}

func oiNew04AssertSnapshotVersionsUnique(t *testing.T, db *sql.DB, portfolioID, date string) {
	t.Helper()
	var count, distinct int
	if err := db.QueryRow(`SELECT COUNT(*), COUNT(DISTINCT snapshot_version) FROM analytics.portfolio_snapshots WHERE portfolio_id=$1::uuid AND snapshot_date=$2::date AND methodology_version=$3`, portfolioID, date, stage371SnapshotMethodology).Scan(&count, &distinct); err != nil {
		t.Fatal(err)
	}
	if count != distinct {
		t.Fatalf("snapshot version collision count=%d distinct=%d", count, distinct)
	}
}

func oiNew04DeleteGenerationEpochFixture(t *testing.T, db *sql.DB, portfolioID string, generation int64) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid AND policy_version=$2 AND activation_generation=$3`, portfolioID, replayPolicyVersion, generation); err != nil {
		t.Fatal(err)
	}
}

func oiNew04GenerationEpochRange(t *testing.T, db *sql.DB, portfolioID string, generation int64) (int64, int64) {
	t.Helper()
	var min, max sql.NullInt64
	if err := db.QueryRow(`SELECT MIN(epoch_generation), MAX(epoch_generation) FROM analytics.portfolio_replay_epochs WHERE portfolio_id=$1::uuid AND policy_version=$2 AND activation_generation=$3`, portfolioID, replayPolicyVersion, generation).Scan(&min, &max); err != nil {
		t.Fatal(err)
	}
	if !min.Valid || !max.Valid {
		t.Fatal("generation epoch range is empty")
	}
	return min.Int64, max.Int64
}

func oiNew04ExerciseRuntimeProfiles(t *testing.T, owner *sql.DB, runtimeURL string) {
	t.Helper()
	set := func(profile RuntimeCapabilityProfile) {
		t.Helper()
		tables := []string{"replay_policy_generations", "replay_policy_events", "portfolio_replay_epochs", "portfolio_replay_positions", "portfolio_replay_financial_state"}
		for _, table := range tables {
			if _, err := owner.Exec(`REVOKE ALL PRIVILEGES ON analytics.` + table + ` FROM openinvest_runtime`); err != nil {
				t.Fatalf("revoke %s: %v", table, err)
			}
		}
		if profile == RuntimeCapabilityR1 || profile == RuntimeCapabilityR2 {
			for _, table := range tables {
				if _, err := owner.Exec(`GRANT SELECT ON analytics.` + table + ` TO openinvest_runtime`); err != nil {
					t.Fatalf("grant SELECT %s: %v", table, err)
				}
			}
		}
		if profile == RuntimeCapabilityR2 {
			for _, table := range []string{"portfolio_replay_epochs", "portfolio_replay_positions", "portfolio_replay_financial_state"} {
				if _, err := owner.Exec(`GRANT INSERT ON analytics.` + table + ` TO openinvest_runtime`); err != nil {
					t.Fatalf("grant INSERT %s: %v", table, err)
				}
			}
		}
	}
	openPass := func(profile RuntimeCapabilityProfile) {
		t.Helper()
		store, err := OpenRuntimeWithCapability(runtimeURL, profile)
		if err != nil {
			t.Fatalf("OpenRuntimeWithCapability(%s): %v", profile, err)
		}
		_ = store.Close()
	}
	openFail := func(profile RuntimeCapabilityProfile) {
		t.Helper()
		store, err := OpenRuntimeWithCapability(runtimeURL, profile)
		if store != nil {
			_ = store.Close()
		}
		if err == nil {
			t.Fatalf("OpenRuntimeWithCapability(%s) unexpectedly passed", profile)
		}
	}

	set(RuntimeCapabilityR0)
	openPass(RuntimeCapabilityR0)
	openFail(RuntimeCapabilityR1)
	set(RuntimeCapabilityR1)
	openPass(RuntimeCapabilityR1)
	openFail(RuntimeCapabilityR2)
	set(RuntimeCapabilityR2)
	openPass(RuntimeCapabilityR2)
	openFail(RuntimeCapabilityR1)

	if _, err := owner.Exec(`GRANT UPDATE ON analytics.portfolio_replay_epochs TO openinvest_runtime`); err != nil {
		t.Fatal(err)
	}
	openFail(RuntimeCapabilityR2)
	if _, err := owner.Exec(`REVOKE UPDATE ON analytics.portfolio_replay_epochs FROM openinvest_runtime`); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.Exec(`REVOKE SELECT ON analytics.replay_policy_events FROM openinvest_runtime`); err != nil {
		t.Fatal(err)
	}
	openFail(RuntimeCapabilityR2)
	set(RuntimeCapabilityR2)
	openPass(RuntimeCapabilityR2)
	set(RuntimeCapabilityR0)
}

// Keep deterministic source order stable in any future evidence extension.
func oiNew04Sorted(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
