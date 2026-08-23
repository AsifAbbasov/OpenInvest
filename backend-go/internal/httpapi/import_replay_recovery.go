package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
)

// expiredImportReviewTokenCanRecover returns true only when the supplied token is currently expired
// but would otherwise pass the complete existing signature, lifetime-shape, subject, portfolio,
// source, parser-digest, row-identity, and decision verification. Untrusted token timestamps are used
// only to choose a historical verification instant; they never bypass the normal verifier.
func (api *API) expiredImportReviewTokenCanRecover(
	token string,
	subjectID string,
	portfolioID string,
	sourceKind string,
	sourceAccountLabel string,
	sourceFileHash string,
	parserReview importer.Review,
	decisions []importer.Decision,
) bool {
	token = strings.TrimSpace(token)
	if len(token) == 0 || len(token) > maxImportReviewTokenBytes {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var payload importReviewTokenPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	if payload.IssuedAt <= 0 || payload.ExpiresAt <= payload.IssuedAt {
		return false
	}
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if api.nowUTC().Before(expiresAt) {
		return false
	}

	// Verify the entire token at the instant it was issued. A copied API avoids mutating the live
	// request clock and makes this safe for concurrent requests.
	recoveryVerifier := *api
	issuedAt := time.Unix(payload.IssuedAt, 0).UTC()
	recoveryVerifier.now = func() time.Time { return issuedAt }
	return recoveryVerifier.verifyImportReviewToken(
		token,
		subjectID,
		portfolioID,
		sourceKind,
		sourceAccountLabel,
		sourceFileHash,
		parserReview,
		decisions,
	) == nil
}
