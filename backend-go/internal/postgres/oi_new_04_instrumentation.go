package postgres

import (
	"context"
	"errors"
	"sync/atomic"
)

// replayInstrumentation is attached to a context by OI-NEW-04 structural/integration tests.
// Production behavior is unchanged when no instrumentation is attached. Atomic counters make the
// hook safe under the race detector and same-portfolio concurrency tests.
type replayInstrumentation struct {
	fullHistoryLoads        atomic.Int64
	ledgerShapeValidations  atomic.Int64
	historyRowsLoaded       atomic.Int64
	historyRowsExamined     atomic.Int64
	ledgerRowsApplied       atomic.Int64
	snapshotWrites          atomic.Int64
	epochWrites             atomic.Int64
	boundedEngineExecutions atomic.Int64
	legacyEngineExecutions  atomic.Int64
}

type replayInstrumentationSnapshot struct {
	FullHistoryLoads        int64
	LedgerShapeValidations  int64
	HistoryRowsLoaded       int64
	HistoryRowsExamined     int64
	LedgerRowsApplied       int64
	SnapshotWrites          int64
	EpochWrites             int64
	BoundedEngineExecutions int64
	LegacyEngineExecutions  int64
}

type replayInstrumentationContextKey struct{}

type replayTestFaultContextKey struct{}

type replayTestFault struct {
	failAfterEpochWrites   int64
	cancelAfterEpochWrites int64
	cancel                 context.CancelFunc
	err                    error
	fired                  atomic.Bool
}

func withReplayInstrumentation(ctx context.Context, instrumentation *replayInstrumentation) context.Context {
	if instrumentation == nil {
		return ctx
	}
	return context.WithValue(ctx, replayInstrumentationContextKey{}, instrumentation)
}

// withReplayTestFault is deliberately unexported and context-scoped. It is a deterministic,
// non-global candidate-only seam used to prove full-command transaction rollback/cancellation.
func withReplayTestFault(ctx context.Context, fault *replayTestFault) context.Context {
	if fault == nil {
		return ctx
	}
	return context.WithValue(ctx, replayTestFaultContextKey{}, fault)
}

func replayInstrumentationFromContext(ctx context.Context) *replayInstrumentation {
	if ctx == nil {
		return nil
	}
	instrumentation, _ := ctx.Value(replayInstrumentationContextKey{}).(*replayInstrumentation)
	return instrumentation
}

func replayInstrumentationSnapshotOf(instrumentation *replayInstrumentation) replayInstrumentationSnapshot {
	if instrumentation == nil {
		return replayInstrumentationSnapshot{}
	}
	return replayInstrumentationSnapshot{
		FullHistoryLoads:        instrumentation.fullHistoryLoads.Load(),
		LedgerShapeValidations:  instrumentation.ledgerShapeValidations.Load(),
		HistoryRowsLoaded:       instrumentation.historyRowsLoaded.Load(),
		HistoryRowsExamined:     instrumentation.historyRowsExamined.Load(),
		LedgerRowsApplied:       instrumentation.ledgerRowsApplied.Load(),
		SnapshotWrites:          instrumentation.snapshotWrites.Load(),
		EpochWrites:             instrumentation.epochWrites.Load(),
		BoundedEngineExecutions: instrumentation.boundedEngineExecutions.Load(),
		LegacyEngineExecutions:  instrumentation.legacyEngineExecutions.Load(),
	}
}

func countReplayFullHistoryLoad(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.fullHistoryLoads.Add(1)
	}
}

func countReplayLedgerShapeValidation(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.ledgerShapeValidations.Add(1)
	}
}

func countReplayHistoryRowsLoaded(ctx context.Context, rows int) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.historyRowsLoaded.Add(int64(rows))
	}
}

func countReplayHistoryRowsExamined(ctx context.Context, rows int) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.historyRowsExamined.Add(int64(rows))
	}
}

func countReplayLedgerRowApplied(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.ledgerRowsApplied.Add(1)
	}
}

func countReplaySnapshotWrite(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.snapshotWrites.Add(1)
	}
}

func countReplayEpochWrite(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.epochWrites.Add(1)
	}
}

func countReplayBoundedEngineExecution(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.boundedEngineExecutions.Add(1)
	}
}

func countReplayLegacyEngineExecution(ctx context.Context) {
	if instrumentation := replayInstrumentationFromContext(ctx); instrumentation != nil {
		instrumentation.legacyEngineExecutions.Add(1)
	}
}

func maybeReplayTestFaultAfterEpochWrite(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	fault, _ := ctx.Value(replayTestFaultContextKey{}).(*replayTestFault)
	if fault == nil || fault.fired.Load() {
		return nil
	}
	instrumentation := replayInstrumentationFromContext(ctx)
	var epochWrites int64
	if instrumentation != nil {
		epochWrites = instrumentation.epochWrites.Load()
	}
	if fault.cancelAfterEpochWrites > 0 && epochWrites >= fault.cancelAfterEpochWrites && fault.fired.CompareAndSwap(false, true) {
		if fault.cancel != nil {
			fault.cancel()
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return context.Canceled
	}
	if fault.failAfterEpochWrites > 0 && epochWrites >= fault.failAfterEpochWrites && fault.fired.CompareAndSwap(false, true) {
		if fault.err != nil {
			return fault.err
		}
		return errors.New("oi-new-04 injected post-epoch failure")
	}
	return nil
}
