package verticalslice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrMissingIdempotency = errors.New("missing idempotency key")
	ErrNotFound           = errors.New("not found")
)

var tickerPattern = regexp.MustCompile(`^[A-Z0-9]{1,32}$`)
var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{16,128}$`)

type Service struct {
	store Store
	clock Clock
}

func NewService(store Store, clock Clock) *Service {
	return &Service{store: store, clock: clock}
}

func (s *Service) Ready(ctx context.Context) error {
	return s.store.Ping(ctx)
}

func (s *Service) SearchAssets(ctx context.Context, filter AssetSearchFilter) (AssetSearchResult, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	if filter.Query == "" {
		return AssetSearchResult{}, fmt.Errorf("%w: query is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(filter.Query) > 100 {
		return AssetSearchResult{}, fmt.Errorf("%w: query must be at most 100 characters", ErrInvalidInput)
	}
	if filter.AssetType != "" {
		switch filter.AssetType {
		case "STOCK", "BOND":
		default:
			return AssetSearchResult{}, fmt.Errorf("%w: assetType is invalid", ErrInvalidInput)
		}
	}
	limit := normalizeLimit(filter.Limit, 20, 100)
	if filter.AfterTicker != "" && !tickerPattern.MatchString(filter.AfterTicker) {
		return AssetSearchResult{}, fmt.Errorf("%w: asset keyset anchor is invalid", ErrInvalidInput)
	}
	filter.Limit = limit + 1
	items, err := s.store.SearchAssets(ctx, filter)
	if err != nil {
		return AssetSearchResult{}, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextTicker *string
	if hasMore && len(items) > 0 {
		value := items[len(items)-1].Ticker
		nextTicker = &value
	}
	return AssetSearchResult{Items: items, NextTicker: nextTicker, HasMore: hasMore, Limit: limit}, nil
}

func (s *Service) GetAsset(ctx context.Context, ticker string) (AssetSummary, error) {
	ticker = strings.TrimSpace(ticker)
	if !tickerPattern.MatchString(ticker) {
		return AssetSummary{}, ErrNotFound
	}
	// Stage 3.14 intentionally defers the asset-card detail response until a registered runtime
	// source and all mandatory stock/bond detail fields are available without fabricated data.
	// The HTTP route exists to preserve the frozen API boundary, but it must not emit incomplete
	// detail DTOs or EXAMPLE_* source identifiers.
	return AssetSummary{}, ErrNotFound
}

func (s *Service) ListPortfolios(ctx context.Context, subjectID string, filter PortfolioFilter) ([]Portfolio, error) {
	filter.Limit = normalizeLimit(filter.Limit, 20, 101)
	filter.BeforeID = strings.TrimSpace(filter.BeforeID)
	if (filter.BeforeUpdatedAt == nil) != (filter.BeforeID == "") {
		return nil, fmt.Errorf("%w: portfolio cursor anchor is invalid", ErrInvalidInput)
	}
	return s.store.ListPortfolios(ctx, subjectID, filter)
}

func (s *Service) CreatePortfolio(ctx context.Context, requestContext RequestContext, subjectID string, idempotencyKey string, requestPath string, request CreatePortfolioRequest) (Portfolio, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return Portfolio{}, err
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || len(request.Name) > 100 {
		return Portfolio{}, fmt.Errorf("%w: portfolio name must be 1..100 characters", ErrInvalidInput)
	}
	if request.BaseCurrency != RUB {
		return Portfolio{}, fmt.Errorf("%w: baseCurrency must be RUB", ErrInvalidInput)
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return Portfolio{}, err
	}
	return s.store.CreatePortfolio(ctx, command, request)
}

func (s *Service) GetPortfolio(ctx context.Context, subjectID string, portfolioID string) (Portfolio, error) {
	return s.store.GetPortfolio(ctx, subjectID, portfolioID)
}

func (s *Service) ListTransactions(ctx context.Context, subjectID string, portfolioID string, filter TransactionFilter) ([]Transaction, error) {
	filter.Limit = normalizeLimit(filter.Limit, 50, 101)
	filter.TransactionType = strings.TrimSpace(filter.TransactionType)
	filter.FromDate = strings.TrimSpace(filter.FromDate)
	filter.ToDate = strings.TrimSpace(filter.ToDate)
	filter.BeforeTradeDate = strings.TrimSpace(filter.BeforeTradeDate)
	filter.BeforeEntryID = strings.TrimSpace(filter.BeforeEntryID)
	if filter.TransactionType != "" {
		switch filter.TransactionType {
		case "BUY", "SELL", "DIVIDEND", "COUPON", "DEPOSIT", "WITHDRAWAL", "FEE", "TAX":
		default:
			return nil, fmt.Errorf("%w: transactionType is invalid", ErrInvalidInput)
		}
	}
	if strings.TrimSpace(filter.FromDate) != "" {
		if _, err := time.Parse("2006-01-02", filter.FromDate); err != nil {
			return nil, fmt.Errorf("%w: fromDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	if strings.TrimSpace(filter.ToDate) != "" {
		if _, err := time.Parse("2006-01-02", filter.ToDate); err != nil {
			return nil, fmt.Errorf("%w: toDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	if filter.FromDate != "" && filter.ToDate != "" && filter.FromDate > filter.ToDate {
		return nil, fmt.Errorf("%w: fromDate must be before or equal to toDate", ErrInvalidInput)
	}
	if (filter.BeforeTradeDate == "") != (filter.BeforeEntryID == "") {
		return nil, fmt.Errorf("%w: transaction cursor anchor is invalid", ErrInvalidInput)
	}
	if filter.BeforeTradeDate != "" {
		if _, err := time.Parse("2006-01-02", filter.BeforeTradeDate); err != nil {
			return nil, fmt.Errorf("%w: transaction cursor anchor is invalid", ErrInvalidInput)
		}
	}
	return s.store.ListTransactions(ctx, subjectID, portfolioID, filter)
}

func (s *Service) ListImportReviewTransactions(ctx context.Context, subjectID string, portfolioID string, filter ImportReviewHistoryFilter) ([]Transaction, error) {
	filter.SourceAccountLabel = strings.TrimSpace(filter.SourceAccountLabel)
	if utf8.RuneCountInString(filter.SourceAccountLabel) > 120 {
		return nil, fmt.Errorf("%w: sourceAccountLabel must be at most 120 characters", ErrInvalidInput)
	}
	var err error
	filter.TradeDates, err = normalizeImportReviewDates(filter.TradeDates)
	if err != nil {
		return nil, err
	}
	filter.BrokerOperationKeys, err = normalizeImportReviewHashes(filter.BrokerOperationKeys, "brokerOperationKeys")
	if err != nil {
		return nil, err
	}
	filter.SourceFingerprints, err = normalizeImportReviewHashes(filter.SourceFingerprints, "sourceFingerprints")
	if err != nil {
		return nil, err
	}
	if len(filter.TradeDates) == 0 && len(filter.BrokerOperationKeys) == 0 && len(filter.SourceFingerprints) == 0 {
		return []Transaction{}, nil
	}
	return s.store.ListImportReviewTransactions(ctx, subjectID, portfolioID, filter)
}

func normalizeImportReviewDates(values []string) ([]string, error) {
	if len(values) > 100 {
		return nil, fmt.Errorf("%w: import review date filter exceeds 100 keys", ErrInvalidInput)
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return nil, fmt.Errorf("%w: import review trade date is invalid", ErrInvalidInput)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func normalizeImportReviewHashes(values []string, field string) ([]string, error) {
	if len(values) > 100 {
		return nil, fmt.Errorf("%w: import review %s exceeds 100 keys", ErrInvalidInput, field)
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
			return nil, fmt.Errorf("%w: import review %s contains invalid SHA-256 key", ErrInvalidInput, field)
		}
		decoded, err := hex.DecodeString(value)
		if err != nil || len(decoded) != sha256.Size {
			return nil, fmt.Errorf("%w: import review %s contains invalid SHA-256 key", ErrInvalidInput, field)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func (s *Service) AppendTransaction(ctx context.Context, requestContext RequestContext, subjectID string, idempotencyKey string, requestPath string, request AppendTransactionRequest) (Transaction, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return Transaction{}, err
	}
	if err := validateAppendTransaction(request); err != nil {
		return Transaction{}, err
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return Transaction{}, err
	}
	return s.store.AppendTransaction(ctx, command, request)
}

func (s *Service) AppendImportedTransactions(ctx context.Context, requestContext RequestContext, subjectID string, idempotencyKey string, requestPath string, request AppendImportBatchRequest) ([]Transaction, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}
	prepared, err := prepareAppendImportBatch(request)
	if err != nil {
		return nil, err
	}
	request = prepared
	if err := validateAppendImportBatch(request); err != nil {
		return nil, err
	}

	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return nil, err
	}
	return s.store.AppendImportedTransactions(ctx, command, request)
}

func (s *Service) GetPortfolioSummary(ctx context.Context, subjectID string, portfolioID string, asOfDate string) (PortfolioSummary, error) {
	if strings.TrimSpace(asOfDate) != "" {
		if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
			return PortfolioSummary{}, fmt.Errorf("%w: asOfDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	return s.store.GetPortfolioSummary(ctx, subjectID, portfolioID, asOfDate)
}

func (s *Service) command(requestContext RequestContext, subjectID string, idempotencyKey string, requestPath string, payload any) (CommandContext, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return CommandContext{}, err
	}
	hash := sha256.Sum256(encoded)
	now := s.clock.Now()
	return CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: strings.TrimSpace(idempotencyKey),
		RequestHash:    hex.EncodeToString(hash[:]),
		RequestPath:    requestPath,
		RequestID:      requestContext.RequestID,
		TraceID:        requestContext.TraceID,
		Now:            now,
	}, nil
}

func validateAppendImportBatch(request AppendImportBatchRequest) error {
	if strings.TrimSpace(request.PortfolioID) == "" {
		return fmt.Errorf("%w: portfolioId is required", ErrInvalidInput)
	}
	if request.SourceKind != "USER_UPLOADED_FILE" {
		return fmt.Errorf("%w: sourceKind must be USER_UPLOADED_FILE", ErrInvalidInput)
	}
	if len(request.SourceAccountLabel) > 120 {
		return fmt.Errorf("%w: sourceAccountLabel must be at most 120 characters", ErrInvalidInput)
	}
	sourceFileHash := strings.TrimSpace(request.SourceFileHash)
	if sourceFileHash == "" {
		return fmt.Errorf("%w: sourceFileHash is required", ErrInvalidInput)
	}
	if len(sourceFileHash) != hex.EncodedLen(sha256.Size) {
		return fmt.Errorf("%w: sourceFileHash must be a SHA-256 hex digest", ErrInvalidInput)
	}
	if _, err := hex.DecodeString(sourceFileHash); err != nil {
		return fmt.Errorf("%w: sourceFileHash must be a SHA-256 hex digest", ErrInvalidInput)
	}
	if len(request.Transactions) == 0 {
		return fmt.Errorf("%w: at least one imported transaction is required", ErrInvalidInput)
	}
	if len(request.Transactions) > 100 {
		return fmt.Errorf("%w: imported transaction batch must contain at most 100 rows", ErrInvalidInput)
	}
	seen := map[string]struct{}{}
	seenIdentityStrength := map[string]int{}
	for index, transaction := range request.Transactions {
		if transaction.PortfolioID != request.PortfolioID {
			return fmt.Errorf("%w: imported transaction %d portfolioId does not match batch portfolioId", ErrInvalidInput, index+1)
		}
		if err := validateAppendTransaction(transaction); err != nil {
			return fmt.Errorf("%w: imported transaction %d is invalid", err, index+1)
		}
		if err := validateImportProvenance(transaction.ImportProvenance); err != nil {
			return fmt.Errorf("%w: imported transaction %d has invalid provenance", err, index+1)
		}
		provenance := transaction.ImportProvenance
		identityStrength := 1
		if provenance.BrokerOperationKey != "" {
			identityStrength = 2
		}
		if previousStrength, ok := seenIdentityStrength[provenance.SourceFingerprint]; ok && previousStrength != identityStrength {
			return fmt.Errorf("%w: imported transaction %d mixes fallback and broker identity for the same financial fingerprint", ErrInvalidInput, index+1)
		}
		seenIdentityStrength[provenance.SourceFingerprint] = identityStrength
		key, err := appendRequestBusinessKey(transaction)
		if err != nil {
			return err
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: imported transaction %d duplicates another approved row", ErrInvalidInput, index+1)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateAppendTransaction(request AppendTransactionRequest) error {
	switch request.TransactionType {
	case "BUY":
		if request.Ticker == nil || !tickerPattern.MatchString(*request.Ticker) {
			return fmt.Errorf("%w: ticker is required for trades", ErrInvalidInput)
		}
		if request.Quantity == nil || !request.Quantity.IsPositive() {
			return fmt.Errorf("%w: positive quantity is required for trades", ErrInvalidInput)
		}
		if !request.Quantity.FitsStorage() {
			return fmt.Errorf("%w: quantity exceeds NUMERIC(28,8) storage precision", ErrInvalidInput)
		}
		if request.UnitPrice == nil || !request.UnitPrice.Amount.IsPositive() || request.UnitPrice.Currency != RUB {
			return fmt.Errorf("%w: positive RUB unitPrice is required for trades", ErrInvalidInput)
		}
		if !request.UnitPrice.Amount.FitsStorage() {
			return fmt.Errorf("%w: unitPrice exceeds NUMERIC(28,8) storage precision", ErrInvalidInput)
		}
	case "SELL":
		return fmt.Errorf("%w: SELL is outside Stage 3.2 scope until cost-basis position rebuild is implemented", ErrInvalidInput)
	case "DEPOSIT", "WITHDRAWAL":
		if request.Ticker != nil || request.Quantity != nil || request.UnitPrice != nil {
			return fmt.Errorf("%w: cash flows must not include ticker, quantity, or unitPrice", ErrInvalidInput)
		}
		if !request.Commission.Amount.IsZero() || !request.Tax.Amount.IsZero() {
			return fmt.Errorf("%w: cash flow commission and tax are unsupported and must be zero", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: transactionType is outside Stage 3.2 scope", ErrInvalidInput)
	}

	if request.Commission.Currency != RUB || request.Tax.Currency != RUB {
		return fmt.Errorf("%w: commission and tax must be RUB", ErrInvalidInput)
	}
	if request.Commission.Amount.IsNegative() || request.Tax.Amount.IsNegative() {
		return fmt.Errorf("%w: commission and tax must be non-negative", ErrInvalidInput)
	}
	if !request.Commission.Amount.FitsStorage() || !request.Tax.Amount.FitsStorage() {
		return fmt.Errorf("%w: commission and tax must fit NUMERIC(28,8) storage precision", ErrInvalidInput)
	}
	if request.Note != nil && utf8.RuneCountInString(*request.Note) > 500 {
		return fmt.Errorf("%w: note must be at most 500 characters", ErrInvalidInput)
	}
	if strings.TrimSpace(request.TradeDate) == "" {
		return fmt.Errorf("%w: tradeDate is required", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", request.TradeDate); err != nil {
		return fmt.Errorf("%w: tradeDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	if request.SettlementDate != nil {
		if _, err := time.Parse("2006-01-02", *request.SettlementDate); err != nil {
			return fmt.Errorf("%w: settlementDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	if _, err := GrossFor(request); err != nil {
		return err
	}
	return nil
}

func appendRequestBusinessKey(request AppendTransactionRequest) (string, error) {
	if request.ImportProvenance != nil && request.ImportProvenance.IdentityVersion == ImportIdentityVersion {
		if request.ImportProvenance.BrokerOperationKey != "" {
			return "broker|" + request.ImportProvenance.BrokerOperationKey, nil
		}
		return "fingerprint|" + request.ImportProvenance.SourceFingerprint, nil
	}

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
		settlementDate = *request.SettlementDate
	}
	parts := []string{
		request.PortfolioID,
		request.TransactionType,
		ticker,
		quantity,
		unitPrice,
		gross.Amount.String(),
		request.Commission.Amount.String(),
		request.Tax.Amount.String(),
		request.TradeDate,
		settlementDate,
	}
	return strings.Join(parts, "|"), nil
}

func prepareAppendImportBatch(request AppendImportBatchRequest) (AppendImportBatchRequest, error) {
	request.SourceAccountLabel = strings.TrimSpace(request.SourceAccountLabel)
	prepared := make([]AppendTransactionRequest, len(request.Transactions))
	for index, transaction := range request.Transactions {
		fingerprint, err := NormalizedTransactionFingerprint(transaction)
		if err != nil {
			return AppendImportBatchRequest{}, err
		}

		provenance := transaction.ImportProvenance
		if provenance == nil {
			provenance = &ImportProvenance{
				IdentityVersion:   ImportIdentityVersion,
				SourceFingerprint: fingerprint,
			}
		} else {
			copyProvenance := *provenance
			copyProvenance.BrokerOperationKey = strings.TrimSpace(copyProvenance.BrokerOperationKey)
			copyProvenance.SourceFingerprint = strings.TrimSpace(copyProvenance.SourceFingerprint)
			provenance = &copyProvenance
			if provenance.SourceFingerprint != fingerprint {
				return AppendImportBatchRequest{}, fmt.Errorf("%w: import source fingerprint does not match normalized transaction", ErrInvalidInput)
			}
		}

		transaction.ImportProvenance = provenance
		prepared[index] = transaction
	}
	request.Transactions = prepared
	return request, nil
}

func validateImportProvenance(provenance *ImportProvenance) error {
	if provenance == nil {
		return fmt.Errorf("%w: import provenance is required", ErrInvalidInput)
	}
	if provenance.IdentityVersion != ImportIdentityVersion {
		return fmt.Errorf("%w: unsupported import identity version", ErrInvalidInput)
	}
	if !isSHA256Hex(provenance.SourceFingerprint) {
		return fmt.Errorf("%w: source fingerprint must be a SHA-256 hex digest", ErrInvalidInput)
	}
	if provenance.BrokerOperationKey != "" && !isSHA256Hex(provenance.BrokerOperationKey) {
		return fmt.Errorf("%w: broker operation key must be a SHA-256 hex digest", ErrInvalidInput)
	}
	return nil
}

func isSHA256Hex(value string) bool {
	if len(value) != hex.EncodedLen(sha256.Size) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func ValidateIdempotencyKey(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrMissingIdempotency
	}
	if !idempotencyKeyPattern.MatchString(value) {
		return fmt.Errorf("%w: Idempotency-Key must be 16..128 characters and match ^[A-Za-z0-9._:-]+$", ErrInvalidInput)
	}
	return nil
}

func normalizeLimit(value int, fallback int, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func GrossFor(request AppendTransactionRequest) (Money, error) {
	if request.Quantity == nil || request.UnitPrice == nil {
		if request.GrossAmount != nil {
			if request.GrossAmount.Currency != RUB || request.GrossAmount.Amount.IsNegative() {
				return Money{}, fmt.Errorf("%w: grossAmount must be non-negative RUB", ErrInvalidInput)
			}
			if !request.GrossAmount.Amount.FitsStorage() {
				return Money{}, fmt.Errorf("%w: grossAmount exceeds NUMERIC(28,8) storage precision", ErrInvalidInput)
			}
			return *request.GrossAmount, nil
		}
		return Money{}, fmt.Errorf("%w: grossAmount is required for cash flows", ErrInvalidInput)
	}
	derived := Money{Amount: request.Quantity.Mul(request.UnitPrice.Amount), Currency: RUB}
	if !derived.Amount.FitsStorage() {
		return Money{}, fmt.Errorf("%w: derived grossAmount exceeds NUMERIC(28,8) storage precision", ErrInvalidInput)
	}
	if request.GrossAmount != nil {
		if !request.GrossAmount.Amount.FitsStorage() {
			return Money{}, fmt.Errorf("%w: grossAmount exceeds NUMERIC(28,8) storage precision", ErrInvalidInput)
		}
		if request.GrossAmount.Currency != RUB || !request.GrossAmount.Amount.Equal(derived.Amount) {
			return Money{}, fmt.Errorf("%w: grossAmount must equal quantity multiplied by unitPrice", ErrInvalidInput)
		}
	}
	return derived, nil
}

func ZeroMoney() Money {
	return Money{Amount: decimal.Zero(), Currency: RUB}
}
