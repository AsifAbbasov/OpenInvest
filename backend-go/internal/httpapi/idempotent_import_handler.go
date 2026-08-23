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

// appendImportReplaySafe performs an exact completed-command lookup before rechecking the signed
// review token lifetime. The lookup is read-only and is bound to the authenticated principal,
// canonical path, idempotency key, and canonical import request hash. If no exact completion exists,
// the normal signed-review verification and atomic append path remains mandatory.
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
		if lookupErr != nil {
			return writeReplayAwareError(c, meta, lookupErr)
		}
		if found {
			return writeCommandReplayArtifact(c, artifact)
		}
	}

	if err := api.verifyImportReviewToken(
		request.ReviewToken,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		request.SourceAccountLabel,
		fileHash,
		preflightReview,
		decisions,
	); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
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
