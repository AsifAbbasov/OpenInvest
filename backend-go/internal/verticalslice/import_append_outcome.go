package verticalslice

import (
	"context"
	"errors"
	"fmt"
)

var ErrImportOutcomeUnavailable = errors.New("import append outcome persistence unavailable")

type ImportAppendOutcome struct {
	Transactions         []Transaction
	SnapshotDatesRebuilt []string
}

type ImportAppendOutcomeStore interface {
	AppendImportedTransactionsWithOutcome(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
	) (ImportAppendOutcome, error)
}

type ImportedTransactionsOutcomeReplayBuilder func(ImportAppendOutcome) (CommandReplayArtifact, error)

type ImportOutcomeReplayStore interface {
	AppendImportedTransactionsWithOutcomeReplay(
		ctx context.Context,
		command CommandContext,
		request AppendImportBatchRequest,
		build ImportedTransactionsOutcomeReplayBuilder,
	) ([]Transaction, CommandReplayArtifact, error)
}

func (s *Service) AppendImportedTransactionsWithOutcome(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request AppendImportBatchRequest,
) (ImportAppendOutcome, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return ImportAppendOutcome{}, err
	}
	prepared, err := prepareAppendImportBatch(request)
	if err != nil {
		return ImportAppendOutcome{}, err
	}
	request = prepared
	if err := validateAppendImportBatch(request); err != nil {
		return ImportAppendOutcome{}, err
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return ImportAppendOutcome{}, err
	}
	outcomeStore, ok := s.store.(ImportAppendOutcomeStore)
	if !ok {
		return ImportAppendOutcome{}, ErrImportOutcomeUnavailable
	}
	return outcomeStore.AppendImportedTransactionsWithOutcome(ctx, command, request)
}

func (s *Service) AppendImportedTransactionsWithOutcomeReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request AppendImportBatchRequest,
	build ImportedTransactionsOutcomeReplayBuilder,
) ([]Transaction, CommandReplayArtifact, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return nil, CommandReplayArtifact{}, err
	}
	prepared, err := prepareAppendImportBatch(request)
	if err != nil {
		return nil, CommandReplayArtifact{}, err
	}
	request = prepared
	if err := validateAppendImportBatch(request); err != nil {
		return nil, CommandReplayArtifact{}, err
	}
	if build == nil {
		return nil, CommandReplayArtifact{}, fmt.Errorf("%w: import outcome replay builder is required", ErrReplayUnavailable)
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return nil, CommandReplayArtifact{}, err
	}
	replayStore, ok := s.store.(ImportOutcomeReplayStore)
	if !ok {
		return nil, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return replayStore.AppendImportedTransactionsWithOutcomeReplay(ctx, command, request, build)
}
