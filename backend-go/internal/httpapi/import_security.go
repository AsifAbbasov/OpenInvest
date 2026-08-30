package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"strings"
	"time"
)

func importPayloadHash(payload string) string {
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func normalizedImportReviewSecret(configured []byte) ([]byte, error) {
	secret := bytes.TrimSpace(configured)
	if len(secret) < minImportReviewTokenSecretBytes {
		return nil, fmt.Errorf("import review token secret must be at least %d bytes", minImportReviewTokenSecretBytes)
	}
	hash := sha256.Sum256(secret)
	return hash[:], nil
}

func (api *API) signImportReviewToken(subjectID string, parserReview importer.Review, finalReview importer.Review) (string, error) {
	if parserReview.PortfolioID != finalReview.PortfolioID ||
		parserReview.SourceKind != finalReview.SourceKind ||
		parserReview.SourceAccountLabel != finalReview.SourceAccountLabel ||
		!strings.EqualFold(parserReview.FileHash, finalReview.FileHash) ||
		len(parserReview.Rows) != len(finalReview.Rows) {
		return "", fmt.Errorf("import review phases do not share the same source context")
	}

	parserDigest, err := importer.ReviewSemanticDigest(parserReview)
	if err != nil {
		return "", err
	}
	finalDigest, err := importer.ReviewSemanticDigest(finalReview)
	if err != nil {
		return "", err
	}

	rows := make([]importReviewTokenRowIdentity, 0, len(parserReview.Rows))
	appendableRows := []int{}
	for index, parserRow := range parserReview.Rows {
		finalRow := finalReview.Rows[index]
		if parserRow.RowNumber != finalRow.RowNumber || parserRow.RowHash != finalRow.RowHash {
			return "", fmt.Errorf("import review phases do not share row identity")
		}
		rows = append(rows, importReviewTokenRowIdentity{
			RowNumber: parserRow.RowNumber,
			RowHash:   parserRow.RowHash,
		})
		if finalRow.Status == importer.ReviewStatusAppendable {
			appendableRows = append(appendableRows, finalRow.RowNumber)
		}
	}

	issuedAt := api.nowUTC()
	payload := importReviewTokenPayload{
		Version:            importReviewTokenVersion,
		ParserVersion:      importer.ReviewParserVersion,
		IssuedAt:           issuedAt.Unix(),
		ExpiresAt:          issuedAt.Add(importReviewTokenTTL).Unix(),
		SubjectID:          subjectID,
		PortfolioID:        finalReview.PortfolioID,
		SourceKind:         finalReview.SourceKind,
		SourceAccountLabel: finalReview.SourceAccountLabel,
		SourceFileHash:     strings.ToLower(finalReview.FileHash),
		ParserReviewDigest: parserDigest,
		FinalReviewDigest:  finalDigest,
		AppendableRows:     appendableRows,
		Rows:               rows,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(body)
	signature := api.signImportReviewTokenPart(bodyPart)
	token := bodyPart + "." + signature
	if len(token) > maxImportReviewTokenBytes {
		return "", fmt.Errorf("import review token exceeds %d bytes", maxImportReviewTokenBytes)
	}
	return token, nil
}

func (api *API) verifyImportReviewToken(
	token string,
	subjectID string,
	portfolioID string,
	sourceKind string,
	sourceAccountLabel string,
	sourceFileHash string,
	parserReview importer.Review,
	decisions []importer.Decision,
) error {
	return api.verifyImportReviewTokenForParserVersion(
		token,
		subjectID,
		portfolioID,
		sourceKind,
		sourceAccountLabel,
		sourceFileHash,
		parserReview,
		decisions,
		importer.ReviewParserVersion,
	)
}

func (api *API) verifyImportReviewTokenForParserVersion(
	token string,
	subjectID string,
	portfolioID string,
	sourceKind string,
	sourceAccountLabel string,
	sourceFileHash string,
	parserReview importer.Review,
	decisions []importer.Decision,
	parserVersion int,
) error {
	payload, err := api.decodeImportReviewToken(token)
	if err != nil {
		return err
	}
	if payload.Version != importReviewTokenVersion ||
		payload.ParserVersion != parserVersion ||
		payload.ExpiresAt-payload.IssuedAt != int64(importReviewTokenTTL/time.Second) ||
		payload.IssuedAt <= 0 {
		return fmt.Errorf("%w: reviewToken version or lifetime is invalid", importer.ErrUnsafeAppend)
	}
	now := api.nowUTC()
	issuedAt := time.Unix(payload.IssuedAt, 0).UTC()
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if now.Before(issuedAt.Add(-time.Minute)) || !now.Before(expiresAt) {
		return fmt.Errorf("%w: reviewToken is expired or not yet valid", importer.ErrUnsafeAppend)
	}

	if payload.SubjectID != subjectID ||
		payload.PortfolioID != portfolioID ||
		payload.SourceKind != sourceKind ||
		payload.SourceAccountLabel != strings.TrimSpace(sourceAccountLabel) ||
		payload.SourceFileHash != sourceFileHash {
		return fmt.Errorf("%w: reviewToken does not match append context", importer.ErrUnsafeAppend)
	}
	if !validSHA256Hex(payload.ParserReviewDigest) || !validSHA256Hex(payload.FinalReviewDigest) {
		return fmt.Errorf("%w: reviewToken semantic digest is invalid", importer.ErrUnsafeAppend)
	}

	parserDigest, err := importer.ReviewSemanticDigestForParserVersion(parserReview, parserVersion)
	if err != nil {
		return err
	}
	if parserDigest != payload.ParserReviewDigest {
		return fmt.Errorf("%w: normalized import semantics changed; review again", importer.ErrUnsafeAppend)
	}
	if len(payload.Rows) != len(parserReview.Rows) {
		return fmt.Errorf("%w: reviewToken row set does not match parser review", importer.ErrUnsafeAppend)
	}

	tokenRows := map[int]string{}
	for _, row := range payload.Rows {
		if row.RowNumber < 2 || !validSHA256Hex(row.RowHash) {
			return fmt.Errorf("%w: reviewToken row identity is invalid", importer.ErrUnsafeAppend)
		}
		if _, exists := tokenRows[row.RowNumber]; exists {
			return fmt.Errorf("%w: reviewToken contains duplicate row identity", importer.ErrUnsafeAppend)
		}
		tokenRows[row.RowNumber] = row.RowHash
	}
	for _, row := range parserReview.Rows {
		if tokenRows[row.RowNumber] != row.RowHash {
			return fmt.Errorf("%w: reviewToken row set does not match parser review", importer.ErrUnsafeAppend)
		}
	}

	appendableRows := map[int]struct{}{}
	for _, rowNumber := range payload.AppendableRows {
		if _, exists := appendableRows[rowNumber]; exists {
			return fmt.Errorf("%w: reviewToken contains duplicate appendable row", importer.ErrUnsafeAppend)
		}
		if _, exists := tokenRows[rowNumber]; !exists {
			return fmt.Errorf("%w: reviewToken appendable row is not in reviewed rows", importer.ErrUnsafeAppend)
		}
		appendableRows[rowNumber] = struct{}{}
	}

	for _, decision := range decisions {
		if tokenRows[decision.RowNumber] != decision.RowHash {
			return fmt.Errorf("%w: reviewToken row identity does not match decision", importer.ErrUnsafeAppend)
		}
		if decision.Action == importer.DecisionApprove {
			if _, ok := appendableRows[decision.RowNumber]; !ok {
				return fmt.Errorf("%w: row %d was not appendable in the approved review", importer.ErrUnsafeAppend, decision.RowNumber)
			}
		}
	}
	return nil
}

func (api *API) decodeImportReviewToken(token string) (importReviewTokenPayload, error) {
	token = strings.TrimSpace(token)
	if len(token) == 0 || len(token) > maxImportReviewTokenBytes {
		return importReviewTokenPayload{}, fmt.Errorf("%w: reviewToken is invalid", verticalslice.ErrInvalidInput)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return importReviewTokenPayload{}, fmt.Errorf("%w: reviewToken is invalid", verticalslice.ErrInvalidInput)
	}
	expectedSignature := api.signImportReviewTokenPart(parts[0])
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return importReviewTokenPayload{}, fmt.Errorf("%w: reviewToken signature is invalid", importer.ErrUnsafeAppend)
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return importReviewTokenPayload{}, fmt.Errorf("%w: reviewToken payload is invalid", verticalslice.ErrInvalidInput)
	}
	var payload importReviewTokenPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return importReviewTokenPayload{}, fmt.Errorf("%w: reviewToken payload is invalid", verticalslice.ErrInvalidInput)
	}
	return payload, nil
}

func (api *API) signImportReviewTokenPart(bodyPart string) string {
	mac := hmac.New(sha256.New, api.importReviewSecret)
	_, _ = mac.Write([]byte(bodyPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func validateImportRowCount(totalRows int) error {
	if totalRows > maxHTTPImportRows {
		return fmt.Errorf("%w: import CSV must contain at most %d data rows", verticalslice.ErrInvalidInput, maxHTTPImportRows)
	}
	return nil
}
