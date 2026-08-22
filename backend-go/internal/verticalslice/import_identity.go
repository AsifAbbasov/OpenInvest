package verticalslice

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const ImportIdentityVersion = 1

// BrokerOperationKey returns a privacy-minimized deterministic identity key for a
// broker-provided operation identifier. The raw broker identifier is intentionally
// not persisted in the ledger.
func BrokerOperationKey(raw string) string {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// NormalizedTransactionFingerprint returns the canonical economic fingerprint used
// when a broker operation identifier is unavailable and to bind broker identities
// to the reviewed normalized transaction semantics. Note text is intentionally not
// part of the economic identity.
func NormalizedTransactionFingerprint(request AppendTransactionRequest) (string, error) {
	gross, err := GrossFor(request)
	if err != nil {
		return "", err
	}

	ticker := ""
	if request.Ticker != nil {
		ticker = strings.TrimSpace(*request.Ticker)
	}
	quantity := ""
	if request.Quantity != nil {
		quantity = request.Quantity.String()
	}
	unitPrice := ""
	if request.UnitPrice != nil {
		unitPrice = request.UnitPrice.Amount.String()
	}
	settlementDate := ""
	if request.SettlementDate != nil {
		settlementDate = strings.TrimSpace(*request.SettlementDate)
	}

	fields := []string{
		strings.TrimSpace(request.PortfolioID),
		strings.TrimSpace(request.TransactionType),
		ticker,
		quantity,
		unitPrice,
		gross.Amount.String(),
		request.Commission.Amount.String(),
		request.Tax.Amount.String(),
		strings.TrimSpace(request.TradeDate),
		settlementDate,
	}
	hash := sha256.Sum256([]byte(strings.Join(fields, "|")))
	return hex.EncodeToString(hash[:]), nil
}
