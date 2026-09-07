package verticalslice

import "context"

type stage371StoreAdapter struct {
	Store
}

func adaptStage371Store(store Store) Store {
	if store == nil {
		return store
	}
	return &stage371StoreAdapter{Store: store}
}

func (adapter *stage371StoreAdapter) Ping(ctx context.Context) error {
	if err := adapter.Store.Ping(ctx); err != nil {
		return err
	}
	if stageStore, ok := adapter.Store.(Stage371ReadyStore); ok {
		return stageStore.Stage371Ready(ctx)
	}
	return nil
}

func (adapter *stage371StoreAdapter) GetPortfolioSummary(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (PortfolioSummary, error) {
	if stageStore, ok := adapter.Store.(Stage371SummaryStore); ok {
		return stageStore.GetPortfolioSummaryStage371(ctx, subjectID, portfolioID, asOfDate)
	}
	return adapter.Store.GetPortfolioSummary(ctx, subjectID, portfolioID, asOfDate)
}

func (adapter *stage371StoreAdapter) GetPortfolioPositions(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (PortfolioPositionsProjection, error) {
	positionsStore, ok := adapter.Store.(PortfolioPositionsStore)
	if !ok {
		return PortfolioPositionsProjection{}, ErrPositionProjectionUnavailable
	}
	return positionsStore.GetPortfolioPositions(ctx, subjectID, portfolioID, asOfDate)
}

func (adapter *stage371StoreAdapter) AppendTransaction(
	ctx context.Context,
	command CommandContext,
	request AppendTransactionRequest,
) (Transaction, error) {
	if stageStore, ok := adapter.Store.(Stage371TransactionStore); ok {
		return stageStore.AppendTransactionStage371(ctx, command, request)
	}
	return adapter.Store.AppendTransaction(ctx, command, request)
}

func (adapter *stage371StoreAdapter) AppendImportedTransactions(
	ctx context.Context,
	command CommandContext,
	request AppendImportBatchRequest,
) ([]Transaction, error) {
	if stageStore, ok := adapter.Store.(Stage371ImportStore); ok {
		return stageStore.AppendImportedTransactionsStage371(ctx, command, request)
	}
	return adapter.Store.AppendImportedTransactions(ctx, command, request)
}

func (adapter *stage371StoreAdapter) CreatePortfolioWithReplay(
	ctx context.Context,
	command CommandContext,
	request CreatePortfolioRequest,
	build PortfolioReplayBuilder,
) (Portfolio, CommandReplayArtifact, error) {
	replayStore, ok := adapter.Store.(ReplayStore)
	if !ok {
		return Portfolio{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.CreatePortfolioWithReplay(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) AppendTransactionWithReplay(
	ctx context.Context,
	command CommandContext,
	request AppendTransactionRequest,
	build TransactionReplayBuilder,
) (Transaction, CommandReplayArtifact, error) {
	if stageStore, ok := adapter.Store.(Stage371TransactionReplayStore); ok {
		return stageStore.AppendTransactionWithReplayStage371(ctx, command, request, build)
	}
	replayStore, ok := adapter.Store.(ReplayStore)
	if !ok {
		return Transaction{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.AppendTransactionWithReplay(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) AppendImportedTransactionsWithReplay(
	ctx context.Context,
	command CommandContext,
	request AppendImportBatchRequest,
	build ImportedTransactionsReplayBuilder,
) ([]Transaction, CommandReplayArtifact, error) {
	if stageStore, ok := adapter.Store.(Stage371ImportReplayStore); ok {
		return stageStore.AppendImportedTransactionsWithReplayStage371(ctx, command, request, build)
	}
	replayStore, ok := adapter.Store.(ReplayStore)
	if !ok {
		return nil, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.AppendImportedTransactionsWithReplay(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) AppendImportedTransactionsWithOutcome(
	ctx context.Context,
	command CommandContext,
	request AppendImportBatchRequest,
) (ImportAppendOutcome, error) {
	if stageStore, ok := adapter.Store.(Stage371ImportOutcomeStore); ok {
		return stageStore.AppendImportedTransactionsWithOutcomeStage371(ctx, command, request)
	}
	outcomeStore, ok := adapter.Store.(ImportAppendOutcomeStore)
	if !ok {
		return ImportAppendOutcome{}, ErrImportOutcomeUnavailable
	}
	return outcomeStore.AppendImportedTransactionsWithOutcome(ctx, command, request)
}

func (adapter *stage371StoreAdapter) AppendImportedTransactionsWithOutcomeReplay(
	ctx context.Context,
	command CommandContext,
	request AppendImportBatchRequest,
	build ImportedTransactionsOutcomeReplayBuilder,
) ([]Transaction, CommandReplayArtifact, error) {
	if stageStore, ok := adapter.Store.(Stage371ImportOutcomeReplayStore); ok {
		return stageStore.AppendImportedTransactionsWithOutcomeReplayStage371(ctx, command, request, build)
	}
	replayStore, ok := adapter.Store.(ImportOutcomeReplayStore)
	if !ok {
		return nil, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.AppendImportedTransactionsWithOutcomeReplay(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) LookupReplayArtifact(
	ctx context.Context,
	command CommandContext,
	method string,
) (CommandReplayArtifact, bool, error) {
	lookupStore, ok := adapter.Store.(ReplayLookupStore)
	if !ok {
		return CommandReplayArtifact{}, false, ErrReplayUnavailable
	}
	return lookupStore.LookupReplayArtifact(ctx, command, method)
}

func (adapter *stage371StoreAdapter) CalculateDividendWithReplay(
	ctx context.Context,
	command CommandContext,
	calculation DividendCalculation,
	build DividendReplayBuilder,
) (DividendCalculation, CommandReplayArtifact, error) {
	replayStore, ok := adapter.Store.(DividendReplayStore)
	if !ok {
		return DividendCalculation{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.CalculateDividendWithReplay(ctx, command, calculation, build)
}
