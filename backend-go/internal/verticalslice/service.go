package verticalslice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrMissingIdempotency = errors.New("missing idempotency key")
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

func (s *Service) ListPortfolios(ctx context.Context, subjectID string, limit int) ([]Portfolio, error) {
	return s.store.ListPortfolios(ctx, subjectID, normalizeLimit(limit, 20, 100))
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
	filter.Limit = normalizeLimit(filter.Limit, 50, 100)
	if strings.TrimSpace(filter.TransactionType) != "" {
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
	return s.store.ListTransactions(ctx, subjectID, portfolioID, filter)
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
	if len(request.Transactions) == 0 {
		return fmt.Errorf("%w: at least one imported transaction is required", ErrInvalidInput)
	}
	if len(request.Transactions) > 100 {
		return fmt.Errorf("%w: imported transaction batch must contain at most 100 rows", ErrInvalidInput)
	}
	seen := map[string]struct{}{}
	for index, transaction := range request.Transactions {
		if transaction.PortfolioID != request.PortfolioID {
			return fmt.Errorf("%w: imported transaction %d portfolioId does not match batch portfolioId", ErrInvalidInput, index+1)
		}
		if err := validateAppendTransaction(transaction); err != nil {
			return fmt.Errorf("%w: imported transaction %d is invalid", err, index+1)
		}
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
		if request.UnitPrice == nil || !request.UnitPrice.Amount.IsPositive() || request.UnitPrice.Currency != RUB {
			return fmt.Errorf("%w: positive RUB unitPrice is required for trades", ErrInvalidInput)
		}
	case "SELL":
		return fmt.Errorf("%w: SELL is outside Stage 3.2 scope until cost-basis position rebuild is implemented", ErrInvalidInput)
	case "DEPOSIT", "WITHDRAWAL":
		if request.Ticker != nil || request.Quantity != nil || request.UnitPrice != nil {
			return fmt.Errorf("%w: cash flows must not include ticker, quantity, or unitPrice", ErrInvalidInput)
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
	return nil
}

func appendRequestBusinessKey(request AppendTransactionRequest) (string, error) {
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
			return *request.GrossAmount, nil
		}
		return Money{}, fmt.Errorf("%w: grossAmount is required for cash flows", ErrInvalidInput)
	}
	derived := Money{Amount: request.Quantity.Mul(request.UnitPrice.Amount), Currency: RUB}
	if request.GrossAmount != nil {
		if request.GrossAmount.Currency != RUB || !request.GrossAmount.Amount.Equal(derived.Amount) {
			return Money{}, fmt.Errorf("%w: grossAmount must equal quantity multiplied by unitPrice", ErrInvalidInput)
		}
	}
	return derived, nil
}

func ZeroMoney() Money {
	return Money{Amount: decimal.Zero(), Currency: RUB}
}
