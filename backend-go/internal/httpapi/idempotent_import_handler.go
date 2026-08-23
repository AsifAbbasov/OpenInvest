package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// appendImportReplaySafe preserves the normal signed-review validation path for fresh writes.
// Only an otherwise-valid token whose lifetime has expired may enter read-only replay recovery.
// Signature, context, parser semantics, row identities, and decisions remain mandatory for recovery.
func (api *API) appendImportReplaySafe(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := verticalslice.ValidateIdempotencyKey(c.Get("Idempotency-Key")); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importAppendRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)
	if request.SourceFileHash != fileHash {
		return writeMappedErrorWithMeta(c, meta, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend))
	}

	preflightReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(preflightReview.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	decisions := request.toAppDecisions()
	verifyErr := api.verifyImportReviewToken(
		request.ReviewToken,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		request.SourceAccountLabel,
		fileHash,
		preflightReview,
		decisions,
	)
	if verifyErr != nil {
		// A failed proof must never authorize a new write. Recovery is allowed only when the token
		// is authentic, context/semantics-valid, and failed solely because its lifetime elapsed.
		if api.expiredImportReviewTokenCanRecover(
			request.ReviewToken,
			subjectID,
			portfolioID,
			importer.SourceKindUserUploadedFile,
			request.SourceAccountLabel,
			fileHash,
			preflightReview,
			decisions,
		) {
			if appendRequests, buildErr := importer.BuildAppendRequests(preflightReview, decisions); buildErr == nil && len(appendRequests) > 0 {
				artifact, found, lookupErr := api.service.LookupImportedTransactionsReplay(
					c.Context(),
					meta.toApp(),
					subjectID,
					c.Get("Idempotency-Key"),
					c.Path(),
					verticalslice.AppendImportBatchRequest{
						PortfolioID:        portfolioID,
						Transactions:       appendRequests,
						SourceKind:         preflightReview.SourceKind,
						SourceAccountLabel: preflightReview.SourceAccountLabel,
						SourceFileHash:     preflightReview.FileHash,
						Decisions:          stage0332ImportDecisions(decisions),
					},
				)
				if lookupErr == nil && found {
					return writeCommandReplayArtifact(c, artifact)
				}
			}
		}
		// Preserve the original proof failure unless an exact completed response was recovered.
		return writeMappedErrorWithMeta(c, meta, verifyErr)
	}

	if err := importer.VerifyDecisionIdentities(preflightReview, request.toAppDecisionIdentities()); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	_, artifact, err := importflow.ReviewAndAppendWithReplay(
		c.Context(),
		api.service,
		importflow.Request{
			RequestContext:     meta.toApp(),
			SubjectID:          subjectID,
			PortfolioID:        portfolioID,
			IdempotencyKey:     c.Get("Idempotency-Key"),
			RequestPath:        c.Path(),
			SourceKind:         importer.SourceKindUserUploadedFile,
			SourceAccountLabel: request.SourceAccountLabel,
			SourceFileHash:     fileHash,
			Existing:           nil,
			Reader:             strings.NewReader(request.CSVPayload),
			Decisions:          decisions,
		},
		func(result importflow.Result) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(
				meta,
				http.StatusCreated,
				mapImportAppendResult(portfolioID, importer.SourceKindUserUploadedFile, fileHash, result),
			)
		},
	)
	if err != nil {
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func stage0332ImportDecisions(decisions []importer.Decision) []verticalslice.AppendImportDecision {
	mapped := make([]verticalslice.AppendImportDecision, 0, len(decisions))
	for _, decision := range decisions {
		mapped = append(mapped, verticalslice.AppendImportDecision{
			RowNumber: decision.RowNumber,
			Action:    decision.Action,
		})
	}
	return mapped
}
