package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

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
