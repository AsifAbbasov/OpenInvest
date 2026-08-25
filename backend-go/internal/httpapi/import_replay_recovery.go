package httpapi

import (
	"context"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// recoverCompletedImportReplay reconstructs the command under the parser semantics signed into an
// authentic token and performs only a read-only lookup. It cannot authorize a fresh append: callers
// use its artifact only after current verification failed, and a missing artifact always falls back
// to that failure. The supplied parser version is never accepted for a new write.
func (api *API) recoverCompletedImportReplay(
	ctx context.Context,
	requestContext verticalslice.RequestContext,
	subjectID string,
	portfolioID string,
	idempotencyKey string,
	requestPath string,
	sourceAccountLabel string,
	sourceFileHash string,
	token string,
	csvPayload string,
	decisions []importer.Decision,
) (verticalslice.CommandReplayArtifact, bool) {
	payload, err := api.decodeImportReviewToken(token)
	if err != nil || payload.IssuedAt <= 0 || payload.ExpiresAt <= payload.IssuedAt {
		return verticalslice.CommandReplayArtifact{}, false
	}
	parserReview, err := importer.ReviewCSVForParserVersion(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: sourceAccountLabel,
		FileHash:           sourceFileHash,
		Reader:             strings.NewReader(csvPayload),
	}, payload.ParserVersion)
	if err != nil || validateImportRowCount(parserReview.Summary.TotalRows) != nil {
		return verticalslice.CommandReplayArtifact{}, false
	}

	// Verify the complete historic proof at issuance. A copied API avoids mutating
	// the live request clock and is safe for concurrent requests.
	recoveryVerifier := *api
	issuedAt := time.Unix(payload.IssuedAt, 0).UTC()
	recoveryVerifier.now = func() time.Time { return issuedAt }
	if err := recoveryVerifier.verifyImportReviewTokenForParserVersion(
		token,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		sourceAccountLabel,
		sourceFileHash,
		parserReview,
		decisions,
		payload.ParserVersion,
	); err != nil {
		return verticalslice.CommandReplayArtifact{}, false
	}
	appendRequests, err := importer.BuildAppendRequests(parserReview, decisions)
	if err != nil || len(appendRequests) == 0 {
		return verticalslice.CommandReplayArtifact{}, false
	}
	artifact, found, err := api.service.LookupImportedTransactionsReplay(
		ctx,
		requestContext,
		subjectID,
		idempotencyKey,
		requestPath,
		verticalslice.AppendImportBatchRequest{
			PortfolioID:        portfolioID,
			Transactions:       appendRequests,
			SourceKind:         parserReview.SourceKind,
			SourceAccountLabel: parserReview.SourceAccountLabel,
			SourceFileHash:     parserReview.FileHash,
			Decisions:          stage0332ImportDecisions(decisions),
		},
	)
	if err != nil || !found {
		return verticalslice.CommandReplayArtifact{}, false
	}
	return artifact, true
}
