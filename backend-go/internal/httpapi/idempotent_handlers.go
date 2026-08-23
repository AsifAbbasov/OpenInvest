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

func (api *API) createPortfolioReplay(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request createPortfolioRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}

	_, artifact, err := api.service.CreatePortfolioWithReplay(
		c.Context(),
		meta.toApp(),
		subjectID,
		c.Get("Idempotency-Key"),
		c.Path(),
		verticalslice.CreatePortfolioRequest{Name: request.Name, BaseCurrency: request.BaseCurrency},
		func(portfolio verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(meta, http.StatusCreated, mapPortfolio(portfolio))
		},
	)
	if err != nil {
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func (api *API) appendTransactionReplay(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if !jsonFieldPresent(c.Request().Body(), "settlementDate") {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "settlementDate is required")
	}
	var request appendTransactionRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest, err := request.toApp(c.Params("portfolioId"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	_, artifact, err := api.service.AppendTransactionWithReplay(
		c.Context(),
		meta.toApp(),
		subjectID,
		c.Get("Idempotency-Key"),
		c.Path(),
		appRequest,
		func(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(meta, http.StatusCreated, mapTransaction(transaction))
		},
	)
	if err != nil {
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func (api *API) appendImportReplay(c fiber.Ctx) error {
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
	if err := api.verifyImportReviewToken(
		request.ReviewToken,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		request.SourceAccountLabel,
		fileHash,
		preflightReview,
		request.toAppDecisions(),
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
			Decisions:          request.toAppDecisions(),
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
