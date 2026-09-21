package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestOINew04AUD02SnapshotBounds(t *testing.T) {
	if replaySnapshotDateBound != replayBoundRawRows {
		t.Fatalf("snapshot date bound=%d want replay raw-row bound=%d", replaySnapshotDateBound, replayBoundRawRows)
	}
	if replaySnapshotSQLStatementBound != 1 {
		t.Fatalf("snapshot SQL statement bound=%d want=1", replaySnapshotSQLStatementBound)
	}
}

func oiNew04RunSnapshotFanoutRegression(
	t *testing.T,
	h oiNew04ActiveHarness,
	activeService *verticalslice.Service,
	activationGeneration int64,
) {
	t.Helper()

	// OI-NEW-04-AUD-02 distinct-trade-date proof: exercise the maximum public import batch through
	// the real ACTIVE-R2 service path with 100 different trade dates. This is intentionally separate
	// from the D=5,000 snapshot-cardinality fixture below: one proves realistic distinct command dates,
	// the other proves the independent high-D persistence bound.
	distinctSubject := uuid.NewString()
	distinctPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, distinctSubject, "OI-NEW-04 AUD-02 distinct trade dates")
	const distinctTradeDates = 100
	distinctRows := make([]verticalslice.AppendTransactionRequest, 0, distinctTradeDates)
	for i := 0; i < distinctTradeDates; i++ {
		amount := fmt.Sprintf("%d.00000000", 2000+i)
		distinctRows = append(distinctRows, oiNew04CashRequest(
			distinctPortfolio.ID,
			"DEPOSIT",
			amount,
			oiNew04DateOffset("2024-01-01", i),
		))
	}
	distinctInstrumentation := &replayInstrumentation{}
	_, _, err := activeService.AppendImportedTransactionsWithReplay(
		withReplayInstrumentation(h.ctx, distinctInstrumentation),
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		distinctSubject,
		uuid.NewString(),
		"/api/v1/portfolios/"+distinctPortfolio.ID+"/imports/append",
		verticalslice.AppendImportBatchRequest{
			PortfolioID:        distinctPortfolio.ID,
			Transactions:       distinctRows,
			SourceKind:         "USER_UPLOADED_FILE",
			SourceAccountLabel: "OI-NEW-04-AUD-02",
			SourceFileHash:     fmt.Sprintf("%064x", 42),
		},
		func(items []verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       []byte(fmt.Sprintf("%d", len(items))),
				RequestID:  uuid.NewString(),
				TraceID:    "oi-new-04-aud-02-distinct-dates",
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("ACTIVE-R2 distinct-trade-date import: %v", err)
	}
	distinctCounters := replayInstrumentationSnapshotOf(distinctInstrumentation)
	if distinctCounters.SnapshotWrites != distinctTradeDates || distinctCounters.SnapshotSQLStatements != 1 ||
		distinctCounters.FullHistoryLoads != 0 || distinctCounters.BoundedEngineExecutions != 1 ||
		distinctCounters.LegacyEngineExecutions != 0 || distinctCounters.EpochWrites < 1 {
		t.Fatalf("distinct-trade-date instrumentation=%+v", distinctCounters)
	}
	if got := oiNew04DistinctLedgerTradeDateCount(t, h.owner.db, distinctPortfolio.ID); got != distinctTradeDates {
		t.Fatalf("distinct ledger trade dates=%d want=%d", got, distinctTradeDates)
	}
	if got := oiNew04DistinctSnapshotDateCount(t, h.owner.db, distinctPortfolio.ID); got != distinctTradeDates {
		t.Fatalf("distinct-trade-date snapshots=%d want=%d", got, distinctTradeDates)
	}
	t.Logf("DISTINCT_TRADE_DATES=PASS D=%d", distinctTradeDates)

	// D == B remains admitted. Existing snapshot semantics still rebuild every distinct affected
	// date, but ACTIVE C' must persist all versions through one set-based PostgreSQL statement.
	highDSubject := uuid.NewString()
	highDPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, highDSubject, "OI-NEW-04 AUD-02 high D")
	oiNew04SeedSnapshotDates(t, h.owner.db, highDPortfolio.ID, "2000-01-01", replaySnapshotDateBound, h.clock.now.Add(-time.Hour))
	highDInstrumentation := &replayInstrumentation{}
	highDCtx := withReplayInstrumentation(h.ctx, highDInstrumentation)
	oiNew04AppendCashReplay(t, highDCtx, activeService, highDSubject, highDPortfolio.ID, "DEPOSIT", "101.00000000", "2000-01-01", uuid.NewString())
	highD := replayInstrumentationSnapshotOf(highDInstrumentation)
	if highD.SnapshotWrites != int64(replaySnapshotDateBound) {
		t.Fatalf("high-D snapshot writes=%d want=%d", highD.SnapshotWrites, replaySnapshotDateBound)
	}
	if highD.SnapshotSQLStatements > replaySnapshotSQLStatementBound || highD.SnapshotSQLStatements != 1 {
		t.Fatalf("high-D snapshot SQL statements=%d bound=%d", highD.SnapshotSQLStatements, replaySnapshotSQLStatementBound)
	}
	if highD.FullHistoryLoads != 0 || highD.BoundedEngineExecutions != 1 || highD.LegacyEngineExecutions != 0 || highD.EpochWrites < 1 {
		t.Fatalf("high-D bounded instrumentation=%+v", highD)
	}
	// One raw row is applied once for the snapshot timeline and once for the final epoch state.
	// The count must remain constant as D grows to 5,000; it must not become D*H.
	if highD.LedgerRowsApplied > 2 {
		t.Fatalf("high-D replay multiplied ledger work by snapshot dates: %+v", highD)
	}
	if got := oiNew04SnapshotVersionCount(t, h.owner.db, highDPortfolio.ID, 2); got != replaySnapshotDateBound {
		t.Fatalf("high-D rebuilt snapshot versions=%d want=%d", got, replaySnapshotDateBound)
	}
	if got := oiNew04DistinctSnapshotDateCount(t, h.owner.db, highDPortfolio.ID); got != replaySnapshotDateBound {
		t.Fatalf("high-D distinct snapshot dates=%d want=%d", got, replaySnapshotDateBound)
	}
	t.Logf("HIGH_D_DISTINCT_DATES=PASS D=%d", replaySnapshotDateBound)
	t.Logf("SNAPSHOT_DATE_BOUND=PASS allowed=%d", replaySnapshotDateBound)
	t.Logf("SNAPSHOT_SQL_STATEMENT_BOUND=PASS statements=%d bound=%d", highD.SnapshotSQLStatements, replaySnapshotSQLStatementBound)
	t.Logf("NO_D_TIMES_H_REPLAY=PASS LedgerRowsApplied=%d FullHistoryLoads=%d", highD.LedgerRowsApplied, highD.FullHistoryLoads)

	// D == bound+1 fails after the command enters its transaction but before any snapshot SQL or
	// replay-epoch write. The outer replay command transaction must roll back the reserved command and
	// already-inserted ledger row as well as every other financial/audit effect.
	tooManySubject := uuid.NewString()
	tooManyPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, tooManySubject, "OI-NEW-04 AUD-02 D+1")
	oiNew04SeedSnapshotDates(t, h.owner.db, tooManyPortfolio.ID, "2000-01-01", replaySnapshotDateBound+1, h.clock.now.Add(-time.Hour))
	tooManyKey := uuid.NewString()
	tooManyRequestID := uuid.NewString()
	tooManyBefore := oiNew04StateCounts(t, h.owner.db, tooManyPortfolio.ID, tooManyKey)
	tooManyAuditBefore := oiNew04AuditCountByRequestID(t, h.owner.db, tooManyRequestID)
	tooManyInstrumentation := &replayInstrumentation{}
	_, _, err = activeService.AppendTransactionWithReplay(
		withReplayInstrumentation(h.ctx, tooManyInstrumentation),
		verticalslice.RequestContext{RequestID: tooManyRequestID},
		tooManySubject,
		tooManyKey,
		"/api/v1/portfolios/"+tooManyPortfolio.ID+"/transactions",
		oiNew04CashRequest(tooManyPortfolio.ID, "DEPOSIT", "102.00000000", "2000-01-01"),
		oiNew04TransactionArtifact,
	)
	if !errors.Is(err, ErrRetroactiveReplayWindowExceeded) {
		t.Fatalf("D+1 error=%v want=%v", err, ErrRetroactiveReplayWindowExceeded)
	}
	tooManyAfter := oiNew04StateCounts(t, h.owner.db, tooManyPortfolio.ID, tooManyKey)
	if tooManyBefore != tooManyAfter {
		t.Fatalf("D+1 rejection leaked state before=%+v after=%+v", tooManyBefore, tooManyAfter)
	}
	if afterAudit := oiNew04AuditCountByRequestID(t, h.owner.db, tooManyRequestID); afterAudit != tooManyAuditBefore {
		t.Fatalf("D+1 rejection leaked audit effect before=%d after=%d", tooManyAuditBefore, afterAudit)
	}
	tooManyCounters := replayInstrumentationSnapshotOf(tooManyInstrumentation)
	if tooManyCounters.SnapshotWrites != 0 || tooManyCounters.SnapshotSQLStatements != 0 {
		t.Fatalf("D+1 rejection reached snapshot persistence: %+v", tooManyCounters)
	}
	t.Log("MAX_PLUS_ONE_FAIL_CLOSED=PASS")

	// Cancel immediately after the one high-D set-based snapshot INSERT returns. The rows exist only
	// inside the command transaction at that point; cancellation must roll back snapshots, ledger,
	// idempotency reservation and replay state together.
	cancelSubject := uuid.NewString()
	cancelPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, cancelSubject, "OI-NEW-04 AUD-02 cancellation")
	const cancellationDates = 257
	oiNew04SeedSnapshotDates(t, h.owner.db, cancelPortfolio.ID, "2001-01-01", cancellationDates, h.clock.now.Add(-time.Hour))
	cancelKey := uuid.NewString()
	cancelRequestID := uuid.NewString()
	cancelBefore := oiNew04StateCounts(t, h.owner.db, cancelPortfolio.ID, cancelKey)
	cancelAuditBefore := oiNew04AuditCountByRequestID(t, h.owner.db, cancelRequestID)
	baseCancelCtx, cancel := context.WithCancel(h.ctx)
	defer cancel()
	cancelInstrumentation := &replayInstrumentation{}
	cancelCtx := withReplayInstrumentation(baseCancelCtx, cancelInstrumentation)
	cancelCtx = withReplayTestFault(cancelCtx, &replayTestFault{cancelAfterSnapshotSQLStatements: 1, cancel: cancel})
	_, _, err = activeService.AppendTransactionWithReplay(
		cancelCtx,
		verticalslice.RequestContext{RequestID: cancelRequestID},
		cancelSubject,
		cancelKey,
		"/api/v1/portfolios/"+cancelPortfolio.ID+"/transactions",
		oiNew04CashRequest(cancelPortfolio.ID, "DEPOSIT", "103.00000000", "2001-01-01"),
		oiNew04TransactionArtifact,
	)
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("high-D cancellation error=%v", err)
	}
	cancelCounters := replayInstrumentationSnapshotOf(cancelInstrumentation)
	if cancelCounters.SnapshotSQLStatements != 1 || cancelCounters.SnapshotWrites != cancellationDates {
		t.Fatalf("high-D cancellation did not reach exactly one snapshot statement before rollback: %+v", cancelCounters)
	}
	cancelAfter := oiNew04StateCounts(t, h.owner.db, cancelPortfolio.ID, cancelKey)
	if cancelBefore != cancelAfter {
		t.Fatalf("high-D cancellation leaked state before=%+v after=%+v", cancelBefore, cancelAfter)
	}
	if afterAudit := oiNew04AuditCountByRequestID(t, h.owner.db, cancelRequestID); afterAudit != cancelAuditBefore {
		t.Fatalf("high-D cancellation leaked audit effect before=%d after=%d", cancelAuditBefore, afterAudit)
	}
	t.Log("HIGH_D_CANCELLATION_ROLLBACK=PASS")

	// Same-portfolio high-D commands must still serialize on PostgreSQL ownership. Each admitted
	// command gets one snapshot statement, epochs remain contiguous, snapshot versions remain unique,
	// and the latest epoch watermark must equal the canonical ledger watermark.
	concurrentSubject := uuid.NewString()
	concurrentPortfolio := oiNew04CreatePortfolio(t, h.ctx, activeService, concurrentSubject, "OI-NEW-04 AUD-02 concurrency")
	const concurrencyDates = 128
	const concurrencyFirstDate = "2002-01-01"
	oiNew04SeedSnapshotDates(t, h.owner.db, concurrentPortfolio.ID, concurrencyFirstDate, concurrencyDates, h.clock.now.Add(-time.Hour))

	type concurrentResult struct {
		err      error
		counters replayInstrumentationSnapshot
	}
	start := make(chan struct{})
	results := make(chan concurrentResult, 2)
	var wg sync.WaitGroup
	for i, amount := range []string{"104.00000000", "105.00000000"} {
		wg.Add(1)
		go func(i int, amount string) {
			defer wg.Done()
			<-start
			instrumentation := &replayInstrumentation{}
			_, _, err := activeService.AppendTransactionWithReplay(
				withReplayInstrumentation(h.ctx, instrumentation),
				verticalslice.RequestContext{RequestID: uuid.NewString()},
				concurrentSubject,
				uuid.NewString(),
				"/api/v1/portfolios/"+concurrentPortfolio.ID+"/transactions",
				oiNew04CashRequest(concurrentPortfolio.ID, "DEPOSIT", amount, concurrencyFirstDate),
				oiNew04TransactionArtifact,
			)
			results <- concurrentResult{err: err, counters: replayInstrumentationSnapshotOf(instrumentation)}
		}(i, amount)
	}
	close(start)
	wg.Wait()
	close(results)
	for result := range results {
		if result.err != nil {
			t.Fatalf("high-D same-portfolio command: %v", result.err)
		}
		if result.counters.SnapshotSQLStatements != 1 || result.counters.SnapshotWrites != concurrencyDates ||
			result.counters.EpochWrites < 1 || result.counters.FullHistoryLoads != 0 ||
			result.counters.BoundedEngineExecutions != 1 || result.counters.LegacyEngineExecutions != 0 {
			t.Fatalf("high-D concurrency instrumentation=%+v", result.counters)
		}
	}
	oiNew04AssertEpochSequenceContiguous(t, h.owner.db, concurrentPortfolio.ID)
	oiNew04AssertSnapshotVersionsUnique(t, h.owner.db, concurrentPortfolio.ID, concurrencyFirstDate)
	lastDate := oiNew04DateOffset(concurrencyFirstDate, concurrencyDates-1)
	oiNew04AssertSnapshotVersionsUnique(t, h.owner.db, concurrentPortfolio.ID, lastDate)
	if latest := oiNew04LatestEpochWatermark(t, h.owner.db, concurrentPortfolio.ID, activationGeneration); latest != oiNew04MaxLedgerSequence(t, h.owner.db, concurrentPortfolio.ID) {
		t.Fatalf("high-D concurrency replay watermark=%d ledger watermark=%d", latest, oiNew04MaxLedgerSequence(t, h.owner.db, concurrentPortfolio.ID))
	}
	t.Log("HIGH_D_SAME_PORTFOLIO_CONCURRENCY=PASS")
}

func oiNew04DistinctLedgerTradeDateCount(t *testing.T, db *sql.DB, portfolioID string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(DISTINCT trade_date)
		FROM investment.transaction_entries
		WHERE portfolio_id=$1::uuid
	`, portfolioID).Scan(&count); err != nil {
		t.Fatalf("distinct ledger trade-date count: %v", err)
	}
	return count
}

func oiNew04SeedSnapshotDates(t *testing.T, db *sql.DB, portfolioID, startDate string, count int, calculatedAt time.Time) {
	t.Helper()
	if count < 1 {
		t.Fatalf("invalid snapshot fixture count=%d", count)
	}
	result, err := db.Exec(`
		INSERT INTO analytics.portfolio_snapshots (
			id, portfolio_id, snapshot_date,
			total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			snapshot_version, methodology_version, input_watermark, calculated_at
		)
		SELECT
			md5($1::text || ':oi-new-04-aud-02:' || gs::text)::uuid,
			$1::uuid,
			$2::date + (gs - 1)::integer,
			0, 0, 0, 0,
			0, 0, 0,
			1, $3, 'oi-new-04-aud-02-fixture', $4
		FROM generate_series(1, $5::integer) AS gs
		ORDER BY gs
	`, portfolioID, startDate, stage371SnapshotMethodology, calculatedAt, count)
	if err != nil {
		t.Fatalf("seed %d snapshot dates: %v", count, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("seed snapshot RowsAffected: %v", err)
	}
	if rows != int64(count) {
		t.Fatalf("seeded snapshot rows=%d want=%d", rows, count)
	}
}

func oiNew04SnapshotVersionCount(t *testing.T, db *sql.DB, portfolioID string, version int) int {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id=$1::uuid
		  AND methodology_version=$2
		  AND snapshot_version=$3
	`, portfolioID, stage371SnapshotMethodology, version).Scan(&count); err != nil {
		t.Fatalf("snapshot version count: %v", err)
	}
	return count
}

func oiNew04DistinctSnapshotDateCount(t *testing.T, db *sql.DB, portfolioID string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(DISTINCT snapshot_date)
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id=$1::uuid
	`, portfolioID).Scan(&count); err != nil {
		t.Fatalf("distinct snapshot date count: %v", err)
	}
	return count
}

func oiNew04AuditCountByRequestID(t *testing.T, db *sql.DB, requestID string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit.events WHERE request_id=$1::uuid`, requestID).Scan(&count); err != nil {
		t.Fatalf("audit request count: %v", err)
	}
	return count
}

func oiNew04LatestEpochWatermark(t *testing.T, db *sql.DB, portfolioID string, activationGeneration int64) int64 {
	t.Helper()
	var watermark int64
	if err := db.QueryRow(`
		SELECT build_raw_ledger_watermark
		FROM analytics.portfolio_replay_epochs
		WHERE portfolio_id=$1::uuid
		  AND policy_version=$2
		  AND activation_generation=$3
		ORDER BY epoch_generation DESC
		LIMIT 1
	`, portfolioID, replayPolicyVersion, activationGeneration).Scan(&watermark); err != nil {
		t.Fatalf("latest epoch watermark: %v", err)
	}
	return watermark
}

func oiNew04MaxLedgerSequence(t *testing.T, db *sql.DB, portfolioID string) int64 {
	t.Helper()
	var watermark int64
	if err := db.QueryRow(`
		SELECT COALESCE(MAX(ledger_sequence), 0)
		FROM investment.transaction_entries
		WHERE portfolio_id=$1::uuid
	`, portfolioID).Scan(&watermark); err != nil {
		t.Fatalf("max ledger sequence: %v", err)
	}
	return watermark
}

func oiNew04DateOffset(startDate string, days int) string {
	parsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		panic(fmt.Sprintf("invalid test date %q: %v", startDate, err))
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}
