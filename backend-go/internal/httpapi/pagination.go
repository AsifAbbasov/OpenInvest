package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"strings"
)

type paginationCursorPayload struct {
	Version         int    `json:"v"`
	Resource        string `json:"resource"`
	SubjectID       string `json:"subjectId"`
	PortfolioID     string `json:"portfolioId,omitempty"`
	AssetType       string `json:"assetType,omitempty"`
	QueryHash       string `json:"queryHash,omitempty"`
	Ticker          string `json:"ticker,omitempty"`
	TransactionType string `json:"transactionType,omitempty"`
	FromDate        string `json:"fromDate,omitempty"`
	ToDate          string `json:"toDate,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
	TradeDate       string `json:"tradeDate,omitempty"`
	EntryID         string `json:"entryId"`
}

func derivePaginationCursorSecret(importReviewSecret []byte) []byte {
	mac := hmac.New(sha256.New, importReviewSecret)
	_, _ = mac.Write([]byte("openinvest/pagination-cursor/v1"))
	return mac.Sum(nil)
}

func (api *API) signPaginationCursor(payload paginationCursorPayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, api.paginationCursorSecret)
	_, _ = mac.Write([]byte(bodyPart))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return bodyPart + "." + signature, nil
}

func (api *API) verifyPaginationCursor(raw string) (paginationCursorPayload, error) {
	if len(raw) == 0 || len(raw) > maxPaginationCursorBytes {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	mac := hmac.New(sha256.New, api.paginationCursorSecret)
	_, _ = mac.Write([]byte(parts[0]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	var payload paginationCursorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	return payload, nil
}

func invalidPaginationCursor() error {
	return fmt.Errorf("%w: cursor is invalid", verticalslice.ErrInvalidInput)
}
