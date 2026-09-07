package verticalslice

import "context"

func (adapter *stage371StoreAdapter) CorrectTransactionWithReplayStage374(
	ctx context.Context,
	command CommandContext,
	request CorrectTransactionRequest,
	build TransactionReplayBuilder,
) (Transaction, CommandReplayArtifact, error) {
	store, ok := adapter.Store.(Stage374TransactionRepairReplayStore)
	if !ok {
		return Transaction{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return store.CorrectTransactionWithReplayStage374(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) ReverseTransactionWithReplayStage374(
	ctx context.Context,
	command CommandContext,
	request ReverseTransactionRequest,
	build TransactionReversalReplayBuilder,
) (TransactionReversal, CommandReplayArtifact, error) {
	store, ok := adapter.Store.(Stage374TransactionRepairReplayStore)
	if !ok {
		return TransactionReversal{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return store.ReverseTransactionWithReplayStage374(ctx, command, request, build)
}
