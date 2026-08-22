package importer

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	SourceKindUserUploadedFile = "USER_UPLOADED_FILE"

	ReviewStatusAppendable = "APPENDABLE"
	ReviewStatusDuplicate  = "DUPLICATE"
	ReviewStatusConflict   = "CONFLICT"
	ReviewStatusInvalid    = "INVALID"

	DecisionApprove = "APPROVE"
	DecisionIgnore  = "IGNORE"
	DecisionReject  = "REJECT"
)

var (
	ErrInvalidImport = errors.New("invalid import")
	ErrUnsafeAppend  = errors.New("unsafe append")

	tickerPattern = regexp.MustCompile(`^[A-Z0-9]{1,32}$`)
)

type ReviewRequest struct {
	SubjectID          string
	PortfolioID        string
	SourceKind         string
	SourceAccountLabel string
	FileHash           string
	Existing           []verticalslice.Transaction
	Reader             io.Reader
}

type Review struct {
	SubjectID          string      `json:"subjectId"`
	PortfolioID        string      `json:"portfolioId"`
	SourceKind         string      `json:"sourceKind"`
	SourceAccountLabel string      `json:"sourceAccountLabel"`
	FileHash           string      `json:"fileHash"`
	Summary            Summary     `json:"summary"`
	Rows               []RowReview `json:"rows"`
}

type Summary struct {
	TotalRows      int `json:"totalRows"`
	AppendableRows int `json:"appendableRows"`
	DuplicateRows  int `json:"duplicateRows"`
	ConflictRows   int `json:"conflictRows"`
	InvalidRows    int `json:"invalidRows"`
}

type RowReview struct {
	RowNumber          int        `json:"rowNumber"`
	RowHash            string     `json:"rowHash"`
	Status             string     `json:"status"`
	ReasonCodes        []string   `json:"reasonCodes"`
	Candidate          *Candidate `json:"candidate,omitempty"`
	Fingerprint        string     `json:"fingerprint,omitempty"`
	BrokerOperationID  string     `json:"brokerOperationId,omitempty"`
	sourceFingerprint  string
	brokerOperationKey string
}

type Candidate struct {
	TransactionType string               `json:"transactionType"`
	Ticker          *string              `json:"ticker,omitempty"`
	Quantity        *decimal.Decimal     `json:"quantity,omitempty"`
	UnitPrice       *verticalslice.Money `json:"unitPrice,omitempty"`
	GrossAmount     verticalslice.Money  `json:"grossAmount"`
	Commission      verticalslice.Money  `json:"commission"`
	Tax             verticalslice.Money  `json:"tax"`
	TradeDate       string               `json:"tradeDate"`
	SettlementDate  *string              `json:"settlementDate,omitempty"`
	SafeNote        *string              `json:"safeNote,omitempty"`
}

type Decision struct {
	RowNumber int
	RowHash   string
	Action    string
}

type DecisionIdentity struct {
	RowNumber int
	RowHash   string
}

func ReviewCSV(request ReviewRequest) (Review, error) {
	if strings.TrimSpace(request.SubjectID) == "" {
		return Review{}, fmt.Errorf("%w: subjectId is required", ErrInvalidImport)
	}
	if strings.TrimSpace(request.PortfolioID) == "" {
		return Review{}, fmt.Errorf("%w: portfolioId is required", ErrInvalidImport)
	}
	sourceKind := strings.TrimSpace(request.SourceKind)
	if sourceKind == "" {
		sourceKind = SourceKindUserUploadedFile
	}
	if sourceKind != SourceKindUserUploadedFile {
		return Review{}, fmt.Errorf("%w: sourceKind must be USER_UPLOADED_FILE", ErrInvalidImport)
	}

	reader := csv.NewReader(request.Reader)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return Review{}, fmt.Errorf("%w: CSV header is required", ErrInvalidImport)
	}
	columns, err := mapColumns(header)
	if err != nil {
		return Review{}, err
	}

	review := Review{
		SubjectID:          request.SubjectID,
		PortfolioID:        request.PortfolioID,
		SourceKind:         sourceKind,
		SourceAccountLabel: strings.TrimSpace(request.SourceAccountLabel),
		FileHash:           strings.TrimSpace(request.FileHash),
		Rows:               []RowReview{},
	}
	seenFingerprints := map[string]int{}
	type brokerIdentitySeen struct {
		rowNumber         int
		sourceFingerprint string
	}
	seenBrokerOperationIDs := map[string]brokerIdentitySeen{}
	seenFallbackFingerprints := map[string]int{}
	seenStrongFingerprints := map[string]int{}
	seenNearMatchCandidates := map[string]RowReview{}

	rowNumber := 1
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rowNumber++
		if err != nil {
			review.Rows = append(review.Rows, invalidRow(rowNumber, nil, "MALFORMED_CSV_ROW"))
			continue
		}
		row := reviewRow(request.PortfolioID, columns, record, rowNumber)
		if row.Candidate != nil {
			duplicate, identityConflict := existingIdentityMatch(row, request.PortfolioID, review.SourceAccountLabel, request.Existing)
			if duplicate {
				row.Status = ReviewStatusDuplicate
				row.ReasonCodes = appendReason(row.ReasonCodes, "EXACT_NORMALIZED_FINGERPRINT_MATCH")
			} else if identityConflict {
				row.Status = ReviewStatusConflict
				row.ReasonCodes = appendReason(row.ReasonCodes, "BROKER_OPERATION_IDENTITY_CONFLICT")
			}
			if firstRow, ok := seenFingerprints[row.Fingerprint]; ok {
				row.Status = ReviewStatusDuplicate
				row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("DUPLICATE_IMPORTED_ROW_%d", firstRow))
			} else if row.Fingerprint != "" {
				seenFingerprints[row.Fingerprint] = row.RowNumber
			}
			if row.sourceFingerprint != "" {
				if row.brokerOperationKey == "" {
					if firstRow, ok := seenStrongFingerprints[row.sourceFingerprint]; ok {
						row.Status = ReviewStatusConflict
						row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("MIXED_IMPORT_IDENTITY_STRENGTH_ROW_%d", firstRow))
					}
					if _, ok := seenFallbackFingerprints[row.sourceFingerprint]; !ok {
						seenFallbackFingerprints[row.sourceFingerprint] = row.RowNumber
					}
				} else {
					if firstRow, ok := seenFallbackFingerprints[row.sourceFingerprint]; ok {
						row.Status = ReviewStatusConflict
						row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("MIXED_IMPORT_IDENTITY_STRENGTH_ROW_%d", firstRow))
					}
					if _, ok := seenStrongFingerprints[row.sourceFingerprint]; !ok {
						seenStrongFingerprints[row.sourceFingerprint] = row.RowNumber
					}
				}
			}
			if row.BrokerOperationID != "" {
				brokerScopeKey := request.SubjectID + "|" + request.PortfolioID + "|" + review.SourceAccountLabel + "|" + sourceKind + "|" + row.brokerOperationKey
				if first, ok := seenBrokerOperationIDs[brokerScopeKey]; ok {
					if first.sourceFingerprint == row.sourceFingerprint {
						row.Status = ReviewStatusDuplicate
						row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("DUPLICATE_BROKER_OPERATION_ID_ROW_%d", first.rowNumber))
					} else {
						row.Status = ReviewStatusConflict
						row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("BROKER_OPERATION_IDENTITY_CONFLICT_ROW_%d", first.rowNumber))
					}
				} else {
					seenBrokerOperationIDs[brokerScopeKey] = brokerIdentitySeen{
						rowNumber:         row.RowNumber,
						sourceFingerprint: row.sourceFingerprint,
					}
				}
			}
			if row.Status == ReviewStatusAppendable && nearMatch(row.Candidate, request.Existing) {
				row.Status = ReviewStatusConflict
				row.ReasonCodes = appendReason(row.ReasonCodes, "NEAR_DUPLICATE_REQUIRES_REVIEW")
			}
			if row.Status == ReviewStatusAppendable {
				nearKey := nearMatchKey(row.Candidate)
				if firstRow, ok := seenNearMatchCandidates[nearKey]; ok && nearMatchImported(row.Candidate, firstRow.Candidate) {
					row.Status = ReviewStatusConflict
					row.ReasonCodes = appendReason(row.ReasonCodes, fmt.Sprintf("NEAR_DUPLICATE_IMPORTED_ROW_%d_REQUIRES_REVIEW", firstRow.RowNumber))
				} else if nearKey != "" {
					seenNearMatchCandidates[nearKey] = row
				}
			}
		}
		review.Rows = append(review.Rows, row)
	}
	review.Summary = summarize(review.Rows)
	return review, nil
}

func BuildAppendRequests(review Review, decisions []Decision) ([]verticalslice.AppendTransactionRequest, error) {
	if err := VerifyDecisionIdentities(review, decisionIdentities(decisions)); err != nil {
		return nil, err
	}
	rows := map[int]RowReview{}
	for _, row := range review.Rows {
		rows[row.RowNumber] = row
	}
	decidedRows := map[int]struct{}{}
	appendRequests := []verticalslice.AppendTransactionRequest{}
	for _, decision := range decisions {
		if _, ok := decidedRows[decision.RowNumber]; ok {
			return nil, fmt.Errorf("%w: row %d has duplicate decisions", ErrUnsafeAppend, decision.RowNumber)
		}
		decidedRows[decision.RowNumber] = struct{}{}
		switch decision.Action {
		case DecisionIgnore, DecisionReject:
			continue
		case DecisionApprove:
			row, ok := rows[decision.RowNumber]
			if !ok {
				return nil, fmt.Errorf("%w: row %d is not in review", ErrUnsafeAppend, decision.RowNumber)
			}
			if row.Status != ReviewStatusAppendable || row.Candidate == nil {
				return nil, fmt.Errorf("%w: row %d is not appendable", ErrUnsafeAppend, decision.RowNumber)
			}
			request := row.Candidate.toAppendRequest(review.PortfolioID)
			request.ImportProvenance = &verticalslice.ImportProvenance{
				IdentityVersion:    verticalslice.ImportIdentityVersion,
				BrokerOperationKey: row.brokerOperationKey,
				SourceFingerprint:  row.sourceFingerprint,
			}
			appendRequests = append(appendRequests, request)
		default:
			return nil, fmt.Errorf("%w: decision action is invalid", ErrUnsafeAppend)
		}
	}
	return appendRequests, nil
}

func VerifyDecisionIdentities(review Review, identities []DecisionIdentity) error {
	rows := map[int]RowReview{}
	for _, row := range review.Rows {
		rows[row.RowNumber] = row
	}
	for _, identity := range identities {
		row, ok := rows[identity.RowNumber]
		if !ok {
			return fmt.Errorf("%w: row %d is not in review", ErrUnsafeAppend, identity.RowNumber)
		}
		if strings.TrimSpace(identity.RowHash) == "" {
			return fmt.Errorf("%w: row %d hash is required", ErrUnsafeAppend, identity.RowNumber)
		}
		if identity.RowHash != row.RowHash {
			return fmt.Errorf("%w: row %d hash does not match reviewed row", ErrUnsafeAppend, identity.RowNumber)
		}
	}
	return nil
}

func decisionIdentities(decisions []Decision) []DecisionIdentity {
	identities := make([]DecisionIdentity, 0, len(decisions))
	for _, decision := range decisions {
		identities = append(identities, DecisionIdentity{RowNumber: decision.RowNumber, RowHash: decision.RowHash})
	}
	return identities
}

func reviewRow(portfolioID string, columns map[string]int, record []string, rowNumber int) RowReview {
	rowHash := hashRecord(record)
	candidate, brokerOperationID, brokerOperationKey, reasons := normalizeCandidate(columns, record)
	if len(reasons) > 0 && candidate == nil {
		return RowReview{RowNumber: rowNumber, RowHash: rowHash, Status: ReviewStatusInvalid, ReasonCodes: reasons}
	}
	status := ReviewStatusAppendable
	if len(reasons) > 0 {
		status = ReviewStatusConflict
	}
	fingerprint := ""
	sourceFingerprint := ""
	if candidate != nil {
		appendRequest := candidate.toAppendRequest(portfolioID)
		if normalized, err := verticalslice.NormalizedTransactionFingerprint(appendRequest); err == nil {
			sourceFingerprint = normalized
			fingerprint = reviewFingerprintFor(normalized, brokerOperationKey)
		} else {
			reasons = appendReason(reasons, "NORMALIZED_FINGERPRINT_INVALID")
			status = ReviewStatusConflict
		}
	}
	return RowReview{
		RowNumber:          rowNumber,
		RowHash:            rowHash,
		Status:             status,
		ReasonCodes:        reasons,
		Candidate:          candidate,
		Fingerprint:        fingerprint,
		BrokerOperationID:  brokerOperationID,
		sourceFingerprint:  sourceFingerprint,
		brokerOperationKey: brokerOperationKey,
	}
}

func normalizeCandidate(columns map[string]int, record []string) (*Candidate, string, string, []string) {
	reasons := []string{}
	transactionType := strings.ToUpper(strings.TrimSpace(value(record, columns, "transaction_type")))
	currency := strings.ToUpper(strings.TrimSpace(value(record, columns, "currency")))
	if currency != verticalslice.RUB {
		reasons = append(reasons, "NON_RUB_CURRENCY")
	}
	if transactionType == "" {
		return nil, "", "", []string{"MISSING_TRANSACTION_TYPE"}
	}

	ticker := strings.ToUpper(strings.TrimSpace(value(record, columns, "ticker")))
	quantityText := value(record, columns, "quantity")
	unitPriceText := value(record, columns, "unit_price")
	grossText := value(record, columns, "gross_amount")
	commissionText := value(record, columns, "commission")
	taxText := value(record, columns, "tax")
	tradeDate := strings.TrimSpace(value(record, columns, "trade_date"))
	settlementDateText := strings.TrimSpace(value(record, columns, "settlement_date"))
	brokerOperationID := strings.TrimSpace(value(record, columns, "broker_operation_id"))
	safeBrokerOperationID := neutralizeSpreadsheetText(brokerOperationID)
	brokerOperationKey := verticalslice.BrokerOperationKey(brokerOperationID)
	note := neutralizeSpreadsheetText(value(record, columns, "note"))

	gross, err := parseMoney(grossText, "GROSS_AMOUNT")
	grossParsed := err == nil
	if err != nil {
		reasons = append(reasons, err.Error())
	}
	commission, err := parseOptionalMoney(commissionText, "COMMISSION")
	if err != nil {
		reasons = append(reasons, err.Error())
	}
	tax, err := parseOptionalMoney(taxText, "TAX")
	if err != nil {
		reasons = append(reasons, err.Error())
	}
	if !validBusinessDate(tradeDate) {
		reasons = append(reasons, "INVALID_TRADE_DATE")
	}
	var settlementDate *string
	if settlementDateText != "" {
		if !validBusinessDate(settlementDateText) {
			reasons = append(reasons, "INVALID_SETTLEMENT_DATE")
		} else {
			settlementDate = &settlementDateText
			if tradeDate != "" && settlementDateText < tradeDate {
				reasons = append(reasons, "SETTLEMENT_BEFORE_TRADE_DATE")
			}
		}
	}

	candidate := Candidate{
		TransactionType: transactionType,
		GrossAmount:     gross,
		Commission:      commission,
		Tax:             tax,
		TradeDate:       tradeDate,
		SettlementDate:  settlementDate,
	}
	if note != "" {
		candidate.SafeNote = &note
	}

	switch transactionType {
	case "BUY":
		if !tickerPattern.MatchString(ticker) {
			reasons = append(reasons, "INVALID_TICKER")
		} else {
			candidate.Ticker = &ticker
		}
		quantity, err := parsePositiveDecimal(quantityText, "QUANTITY")
		if err != nil {
			reasons = append(reasons, err.Error())
		} else {
			candidate.Quantity = &quantity
		}
		unitPrice, err := parseMoney(unitPriceText, "UNIT_PRICE")
		if err != nil {
			reasons = append(reasons, err.Error())
		} else if !unitPrice.Amount.IsPositive() {
			reasons = append(reasons, "UNIT_PRICE_NOT_POSITIVE")
		} else {
			candidate.UnitPrice = &unitPrice
		}
		if grossParsed && candidate.Quantity != nil && candidate.UnitPrice != nil {
			expectedGross := candidate.Quantity.Mul(candidate.UnitPrice.Amount)
			if !expectedGross.Equal(candidate.GrossAmount.Amount) {
				reasons = append(reasons, "GROSS_AMOUNT_MISMATCH")
			}
		}
	case "DEPOSIT", "WITHDRAWAL":
		if ticker != "" || strings.TrimSpace(quantityText) != "" || strings.TrimSpace(unitPriceText) != "" {
			reasons = append(reasons, "CASH_FLOW_HAS_ASSET_FIELDS")
		}
		if !commission.Amount.IsZero() || !tax.Amount.IsZero() {
			reasons = append(reasons, "CASH_FLOW_FEES_UNSUPPORTED")
		}
	case "SELL", "DIVIDEND", "COUPON", "FEE", "TAX":
		reasons = append(reasons, "TRANSACTION_TYPE_REQUIRES_FUTURE_REVIEW")
	default:
		reasons = append(reasons, "UNKNOWN_TRANSACTION_TYPE")
	}

	return &candidate, safeBrokerOperationID, brokerOperationKey, uniqueReasons(reasons)
}

func mapColumns(header []string) (map[string]int, error) {
	required := []string{
		"transaction_type", "ticker", "quantity", "unit_price", "gross_amount", "commission",
		"tax", "trade_date", "settlement_date", "currency", "broker_operation_id", "note",
	}
	columns := map[string]int{}
	for index, value := range header {
		columns[strings.ToLower(strings.TrimSpace(value))] = index
	}
	for _, column := range required {
		if _, ok := columns[column]; !ok {
			return nil, fmt.Errorf("%w: missing CSV column %s", ErrInvalidImport, column)
		}
	}
	return columns, nil
}

func value(record []string, columns map[string]int, column string) string {
	index := columns[column]
	if index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func parseMoney(input string, code string) (verticalslice.Money, error) {
	amount, err := decimal.FromString(input)
	if err != nil {
		return verticalslice.ZeroMoney(), fmt.Errorf("%s_INVALID", code)
	}
	if amount.IsNegative() {
		return verticalslice.ZeroMoney(), fmt.Errorf("%s_NEGATIVE", code)
	}
	return verticalslice.Money{Amount: amount, Currency: verticalslice.RUB}, nil
}

func parseOptionalMoney(input string, code string) (verticalslice.Money, error) {
	if strings.TrimSpace(input) == "" {
		return verticalslice.ZeroMoney(), nil
	}
	return parseMoney(input, code)
}

func parsePositiveDecimal(input string, code string) (decimal.Decimal, error) {
	value, err := decimal.FromString(input)
	if err != nil {
		return decimal.Zero(), fmt.Errorf("%s_INVALID", code)
	}
	if !value.IsPositive() {
		return decimal.Zero(), fmt.Errorf("%s_NOT_POSITIVE", code)
	}
	return value, nil
}

func validBusinessDate(input string) bool {
	if strings.TrimSpace(input) == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", input)
	return err == nil
}

func hashRecord(record []string) string {
	hash := sha256.Sum256([]byte(strings.Join(record, "\x1f")))
	return hex.EncodeToString(hash[:])
}

func reviewFingerprintFor(sourceFingerprint string, brokerOperationKey string) string {
	hash := sha256.Sum256([]byte(sourceFingerprint + "|" + brokerOperationKey))
	return hex.EncodeToString(hash[:])
}

func existingIdentityMatch(row RowReview, portfolioID string, sourceAccountLabel string, transactions []verticalslice.Transaction) (duplicate bool, conflict bool) {
	if row.Candidate == nil || row.sourceFingerprint == "" {
		return false, false
	}
	sourceAccountLabel = strings.TrimSpace(sourceAccountLabel)
	for _, transaction := range transactions {
		if transaction.SourceKind == SourceKindUserUploadedFile && transaction.SourceIdentityVersion == verticalslice.ImportIdentityVersion {
			if transaction.SourceAccountLabel != sourceAccountLabel {
				continue
			}
			if row.brokerOperationKey != "" {
				if transaction.SourceBrokerOperationKey == row.brokerOperationKey {
					if transaction.SourceFingerprint == row.sourceFingerprint {
						return true, false
					}
					return false, true
				}
				if transaction.SourceBrokerOperationKey == "" && transaction.SourceFingerprint == row.sourceFingerprint {
					return false, true
				}
				continue
			}
			if transaction.SourceFingerprint == row.sourceFingerprint {
				return true, false
			}
			continue
		}

		if existingTransactionFingerprint(portfolioID, transaction) == row.sourceFingerprint {
			return true, false
		}
	}
	return false, false
}

func existingTransactionFingerprint(portfolioID string, transaction verticalslice.Transaction) string {
	gross := transaction.GrossAmount
	request := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: transaction.TransactionType,
		Ticker:          transaction.Ticker,
		Quantity:        transaction.Quantity,
		UnitPrice:       transaction.UnitPrice,
		GrossAmount:     &gross,
		Commission:      transaction.Commission,
		Tax:             transaction.Tax,
		TradeDate:       transaction.TradeDate,
		SettlementDate:  transaction.SettlementDate,
	}
	fingerprint, err := verticalslice.NormalizedTransactionFingerprint(request)
	if err != nil {
		return ""
	}
	return fingerprint
}

func nearMatch(candidate *Candidate, existing []verticalslice.Transaction) bool {
	if candidate == nil {
		return false
	}
	for _, transaction := range existing {
		if nearMatchTransactionKey(transaction) == nearMatchKey(candidate) {
			if transaction.Commission.Amount.String() != candidate.Commission.Amount.String() ||
				transaction.Tax.Amount.String() != candidate.Tax.Amount.String() ||
				transaction.GrossAmount.Amount.String() != candidate.GrossAmount.Amount.String() {
				return true
			}
		}
	}
	return false
}

func nearMatchImported(candidate *Candidate, existing *Candidate) bool {
	if candidate == nil || existing == nil {
		return false
	}
	if nearMatchKey(candidate) != nearMatchKey(existing) {
		return false
	}
	return candidate.Commission.Amount.String() != existing.Commission.Amount.String() ||
		candidate.Tax.Amount.String() != existing.Tax.Amount.String() ||
		candidate.GrossAmount.Amount.String() != existing.GrossAmount.Amount.String()
}

func nearMatchKey(candidate *Candidate) string {
	if candidate == nil {
		return ""
	}
	parts := []string{
		candidate.TransactionType,
		candidate.TradeDate,
		ptr(candidate.Ticker),
		decimalPtr(candidate.Quantity),
	}
	if candidate.TransactionType == "DEPOSIT" || candidate.TransactionType == "WITHDRAWAL" {
		parts = append(parts, candidate.GrossAmount.Amount.String())
	}
	return strings.Join(parts, "|")
}

func nearMatchTransactionKey(transaction verticalslice.Transaction) string {
	parts := []string{
		transaction.TransactionType,
		transaction.TradeDate,
		ptr(transaction.Ticker),
		decimalPtr(transaction.Quantity),
	}
	if transaction.TransactionType == "DEPOSIT" || transaction.TransactionType == "WITHDRAWAL" {
		parts = append(parts, transaction.GrossAmount.Amount.String())
	}
	return strings.Join(parts, "|")
}

func summarize(rows []RowReview) Summary {
	summary := Summary{TotalRows: len(rows)}
	for _, row := range rows {
		switch row.Status {
		case ReviewStatusAppendable:
			summary.AppendableRows++
		case ReviewStatusDuplicate:
			summary.DuplicateRows++
		case ReviewStatusConflict:
			summary.ConflictRows++
		case ReviewStatusInvalid:
			summary.InvalidRows++
		}
	}
	return summary
}

func invalidRow(rowNumber int, record []string, reason string) RowReview {
	return RowReview{RowNumber: rowNumber, RowHash: hashRecord(record), Status: ReviewStatusInvalid, ReasonCodes: []string{reason}}
}

func appendReason(reasons []string, reason string) []string {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}

func uniqueReasons(reasons []string) []string {
	if len(reasons) == 0 {
		return reasons
	}
	seen := map[string]struct{}{}
	unique := []string{}
	for _, reason := range reasons {
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		unique = append(unique, reason)
	}
	sort.Strings(unique)
	return unique
}

func neutralizeSpreadsheetText(input string) string {
	value := strings.TrimSpace(input)
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + value
	default:
		return value
	}
}

func ptr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func decimalPtr(value *decimal.Decimal) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func moneyPtr(value *verticalslice.Money) string {
	if value == nil {
		return ""
	}
	return value.Amount.String()
}

func (candidate Candidate) toAppendRequest(portfolioID string) verticalslice.AppendTransactionRequest {
	return verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: candidate.TransactionType,
		Ticker:          candidate.Ticker,
		Quantity:        candidate.Quantity,
		UnitPrice:       candidate.UnitPrice,
		GrossAmount:     &candidate.GrossAmount,
		Commission:      candidate.Commission,
		Tax:             candidate.Tax,
		TradeDate:       candidate.TradeDate,
		SettlementDate:  candidate.SettlementDate,
		Note:            candidate.SafeNote,
	}
}
