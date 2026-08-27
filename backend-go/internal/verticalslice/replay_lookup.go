package verticalslice

import "context"

// ReplayLookupStore exposes a read-only exact replay lookup used when a completed command must be
// replayed before time-sensitive request proofs (for example an import review token) are rechecked.
// The lookup never reserves a key and never creates a business effect.
type ReplayLookupStore interface {
	LookupReplayArtifact(ctx context.Context, command CommandContext, method string) (CommandReplayArtifact, bool, error)
}

// lookupPortfolioReplay reconstructs the pre-Stage-3.39 trim-first portfolio command identity
// and performs only the generic read-only completed-command lookup. It never reserves, reclaims,
// writes, creates a business effect, or extends replay authority.
func (s *Service) lookupPortfolioReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request CreatePortfolioRequest,
) (CommandReplayArtifact, bool, error) {
	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return CommandReplayArtifact{}, false, err
	}
	lookupStore, ok := s.store.(ReplayLookupStore)
	if !ok {
		return CommandReplayArtifact{}, false, ErrReplayUnavailable
	}
	return lookupStore.LookupReplayArtifact(ctx, command, "POST")
}

// LookupImportedTransactionsReplay reconstructs the same canonical import command hash used by
// AppendImportedTransactionsWithReplay and performs a read-only lookup for an already completed
// exact response artifact. A missing artifact returns found=false and leaves normal validation and
// append processing to the caller.
func (s *Service) LookupImportedTransactionsReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request AppendImportBatchRequest,
) (CommandReplayArtifact, bool, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return CommandReplayArtifact{}, false, err
	}
	prepared, err := prepareAppendImportBatch(request)
	if err != nil {
		return CommandReplayArtifact{}, false, err
	}
	request = prepared
	if err := validateAppendImportBatch(request); err != nil {
		return CommandReplayArtifact{}, false, err
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return CommandReplayArtifact{}, false, err
	}
	lookupStore, ok := s.store.(ReplayLookupStore)
	if !ok {
		return CommandReplayArtifact{}, false, ErrReplayUnavailable
	}
	return lookupStore.LookupReplayArtifact(ctx, command, "POST")
}
