package verticalslice

import "context"

type Stage371ReadyStore interface {
	Stage371Ready(ctx context.Context) error
}

type Stage371TransactionStore interface {
	AppendTransactionStage371(
		ctx context.Context,
		command CommandContext,
		request AppendTransactionRequest,
	) (Transaction, error)
}

type Stage371TransactionReplayStore interface {
	AppendTransactionWithReplayStage371(
		ctx context.Context,
		command CommandContext,
		request AppendTransactionRequest,
		build TransactionReplayBuilder,
	) (Transaction, CommandReplayArtifact, error)
}

type Stage371ImportStore interface {
	AppendImportedTransactionsStage371(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
	) ([]Transaction, error)
}

type Stage371ImportReplayStore interface {
	AppendImportedTransactionsWithReplayStage371(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
		build ImportedTransactionsReplayBuilder,
	) ([]Transaction, CommandReplayArtifact, error)
}

type Stage371ImportOutcomeStore interface {
	AppendImportedTransactionsWithOutcomeStage371(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
	) (ImportAppendOutcome, error)
}

type Stage371ImportOutcomeReplayStore interface {
	AppendImportedTransactionsWithOutcomeReplayStage371(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
		build ImportedTransactionsOutcomeReplayBuilder,
	) ([]Transaction, CommandReplayArtifact, error)
}
