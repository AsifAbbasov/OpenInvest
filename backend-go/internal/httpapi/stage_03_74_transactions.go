package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type correctTransactionRequestDTO struct {
	ExpectedRevision int                         `json:"expectedRevision"`
	Reason           string                      `json:"reason"`
	Corrected        appendTransactionRequestDTO `json:"corrected"`
}

type reverseTransactionRequestDTO struct {
	ExpectedRevision int    `json:"expectedRevision"`
	Reason           string `json:"reason"`
	EffectiveDate    string `json:"effectiveDate"`
}

type transactionReversalDTO struct {
	TransactionID         string `json:"transactionId"`
	ReversalTransactionID string `json:"reversalTransactionId"`
	Status                string `json:"status"`
	EffectiveDate         string `json:"effectiveDate"`
}

func (api *API) correctTransactionReplay(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request correctTransactionRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if !nestedJSONFieldPresent(c.Request().Body(), "corrected", "settlementDate") {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "corrected.settlementDate is required")
	}
	corrected, err := request.Corrected.toApp(c.Params("portfolioId"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	appRequest := verticalslice.CorrectTransactionRequest{
		PortfolioID:      c.Params("portfolioId"),
		TransactionID:    c.Params("transactionId"),
		ExpectedRevision: request.ExpectedRevision,
		Reason:           request.Reason,
		Corrected:        corrected,
	}
	_, artifact, err := api.service.CorrectTransactionWithReplay(
		c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), appRequest,
		func(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(meta, http.StatusOK, mapTransaction(transaction))
		},
	)
	if err != nil {
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func (api *API) reverseTransactionReplay(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request reverseTransactionRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest := verticalslice.ReverseTransactionRequest{
		PortfolioID:      c.Params("portfolioId"),
		TransactionID:    c.Params("transactionId"),
		ExpectedRevision: request.ExpectedRevision,
		Reason:           request.Reason,
		EffectiveDate:    request.EffectiveDate,
	}
	_, artifact, err := api.service.ReverseTransactionWithReplay(
		c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), appRequest,
		func(result verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(meta, http.StatusOK, transactionReversalDTO{
				TransactionID:         result.TransactionID,
				ReversalTransactionID: result.ReversalTransactionID,
				Status:                result.Status,
				EffectiveDate:         result.EffectiveDate,
			})
		},
	)
	if err != nil {
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func nestedJSONFieldPresent(body []byte, objectField string, field string) bool {
	var raw map[string]json.RawMessage
	if json.Unmarshal(body, &raw) != nil {
		return false
	}
	nestedRaw, ok := raw[objectField]
	if !ok {
		return false
	}
	var nested map[string]json.RawMessage
	if json.Unmarshal(nestedRaw, &nested) != nil {
		return false
	}
	_, ok = nested[field]
	return ok
}
