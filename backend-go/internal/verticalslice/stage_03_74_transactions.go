package verticalslice

import (
	"context"
	"fmt"
	"strings"
	"time"
)

var ErrTransactionConflict = fmt.Errorf("transaction revision conflict")

type CorrectTransactionRequest struct {
	PortfolioID      string
	TransactionID    string
	ExpectedRevision int
	Reason           string
	Corrected        AppendTransactionRequest
}

type ReverseTransactionRequest struct {
	PortfolioID      string
	TransactionID    string
	ExpectedRevision int
	Reason           string
	EffectiveDate    string
}

type TransactionReversal struct {
	TransactionID         string
	ReversalTransactionID string
	Status                string
	EffectiveDate         string
}

type TransactionReversalReplayBuilder func(TransactionReversal) (CommandReplayArtifact, error)

type Stage374TransactionRepairReplayStore interface {
	CorrectTransactionWithReplayStage374(
		ctx context.Context,
		command CommandContext,
		request CorrectTransactionRequest,
		build TransactionReplayBuilder,
	) (Transaction, CommandReplayArtifact, error)
	ReverseTransactionWithReplayStage374(
		ctx context.Context,
		command CommandContext,
		request ReverseTransactionRequest,
		build TransactionReversalReplayBuilder,
	) (TransactionReversal, CommandReplayArtifact, error)
}

func (s *Service) CorrectTransactionWithReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request CorrectTransactionRequest,
	build TransactionReplayBuilder,
) (Transaction, CommandReplayArtifact, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return Transaction{}, CommandReplayArtifact{}, err
	}
	if strings.TrimSpace(request.PortfolioID) == "" || strings.TrimSpace(request.TransactionID) == "" {
		return Transaction{}, CommandReplayArtifact{}, fmt.Errorf("%w: portfolioId and transactionId are required", ErrInvalidInput)
	}
	if request.ExpectedRevision < 1 {
		return Transaction{}, CommandReplayArtifact{}, fmt.Errorf("%w: expectedRevision must be at least 1", ErrInvalidInput)
	}
	request.Reason = strings.TrimSpace(request.Reason)
	if request.Reason == "" {
		return Transaction{}, CommandReplayArtifact{}, fmt.Errorf("%w: correction reason is required", ErrInvalidInput)
	}
	if err := validateUnicodeCodePointMax(request.Reason, 300, "reason"); err != nil {
		return Transaction{}, CommandReplayArtifact{}, err
	}
	request.Corrected.PortfolioID = request.PortfolioID
	request.Corrected.ImportProvenance = nil
	if err := validateCorrectedTransaction(request.Corrected); err != nil {
		return Transaction{}, CommandReplayArtifact{}, err
	}
	if build == nil {
		return Transaction{}, CommandReplayArtifact{}, fmt.Errorf("%w: correction replay builder is required", ErrReplayUnavailable)
	}
	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return Transaction{}, CommandReplayArtifact{}, err
	}
	store, ok := s.store.(Stage374TransactionRepairReplayStore)
	if !ok {
		return Transaction{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return store.CorrectTransactionWithReplayStage374(ctx, command, request, build)
}

func (s *Service) ReverseTransactionWithReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request ReverseTransactionRequest,
	build TransactionReversalReplayBuilder,
) (TransactionReversal, CommandReplayArtifact, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return TransactionReversal{}, CommandReplayArtifact{}, err
	}
	if strings.TrimSpace(request.PortfolioID) == "" || strings.TrimSpace(request.TransactionID) == "" {
		return TransactionReversal{}, CommandReplayArtifact{}, fmt.Errorf("%w: portfolioId and transactionId are required", ErrInvalidInput)
	}
	if request.ExpectedRevision < 1 {
		return TransactionReversal{}, CommandReplayArtifact{}, fmt.Errorf("%w: expectedRevision must be at least 1", ErrInvalidInput)
	}
	request.Reason = strings.TrimSpace(request.Reason)
	if request.Reason == "" {
		return TransactionReversal{}, CommandReplayArtifact{}, fmt.Errorf("%w: reversal reason is required", ErrInvalidInput)
	}
	if err := validateUnicodeCodePointMax(request.Reason, 300, "reason"); err != nil {
		return TransactionReversal{}, CommandReplayArtifact{}, err
	}
	request.EffectiveDate = strings.TrimSpace(request.EffectiveDate)
	if _, err := time.Parse("2006-01-02", request.EffectiveDate); err != nil {
		return TransactionReversal{}, CommandReplayArtifact{}, fmt.Errorf("%w: effectiveDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	if build == nil {
		return TransactionReversal{}, CommandReplayArtifact{}, fmt.Errorf("%w: reversal replay builder is required", ErrReplayUnavailable)
	}
	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, request)
	if err != nil {
		return TransactionReversal{}, CommandReplayArtifact{}, err
	}
	store, ok := s.store.(Stage374TransactionRepairReplayStore)
	if !ok {
		return TransactionReversal{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	return store.ReverseTransactionWithReplayStage374(ctx, command, request, build)
}

func validateCorrectedTransaction(request AppendTransactionRequest) error {
	switch request.TransactionType {
	case "BUY", "SELL":
		if request.Ticker == nil || !tickerPattern.MatchString(*request.Ticker) {
			return fmt.Errorf("%w: ticker is required for trades", ErrInvalidInput)
		}
		if request.Quantity == nil || !request.Quantity.IsPositive() || !request.Quantity.FitsStorage() {
			return fmt.Errorf("%w: positive storage-safe quantity is required for trades", ErrInvalidInput)
		}
		if request.UnitPrice == nil || !request.UnitPrice.Amount.IsPositive() || request.UnitPrice.Currency != RUB || !request.UnitPrice.Amount.FitsStorage() {
			return fmt.Errorf("%w: positive storage-safe RUB unitPrice is required for trades", ErrInvalidInput)
		}
	case "DIVIDEND", "COUPON":
		if request.Ticker == nil || !tickerPattern.MatchString(*request.Ticker) {
			return fmt.Errorf("%w: ticker is required for income transactions", ErrInvalidInput)
		}
		if request.UnitPrice != nil {
			return fmt.Errorf("%w: income transactions must not include unitPrice", ErrInvalidInput)
		}
		if request.Quantity != nil && (!request.Quantity.IsPositive() || !request.Quantity.FitsStorage()) {
			return fmt.Errorf("%w: income quantity must be positive and storage-safe when supplied", ErrInvalidInput)
		}
	case "FEE", "TAX":
		if request.Quantity != nil || request.UnitPrice != nil {
			return fmt.Errorf("%w: expense transactions must not include quantity or unitPrice", ErrInvalidInput)
		}
		if request.Ticker != nil && !tickerPattern.MatchString(*request.Ticker) {
			return fmt.Errorf("%w: expense ticker is invalid", ErrInvalidInput)
		}
	case "DEPOSIT", "WITHDRAWAL":
		if request.Ticker != nil || request.Quantity != nil || request.UnitPrice != nil {
			return fmt.Errorf("%w: cash flows must not include ticker, quantity, or unitPrice", ErrInvalidInput)
		}
		if !request.Commission.Amount.IsZero() || !request.Tax.Amount.IsZero() {
			return fmt.Errorf("%w: cash flow commission and tax are unsupported and must be zero", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: transactionType is invalid", ErrInvalidInput)
	}
	if request.Commission.Currency != RUB || request.Tax.Currency != RUB || request.Commission.Amount.IsNegative() || request.Tax.Amount.IsNegative() {
		return fmt.Errorf("%w: commission and tax must be non-negative RUB", ErrInvalidInput)
	}
	if !request.Commission.Amount.FitsStorage() || !request.Tax.Amount.FitsStorage() {
		return fmt.Errorf("%w: commission and tax must fit NUMERIC(28,8)", ErrInvalidInput)
	}
	if request.Note != nil {
		if err := validateUnicodeCodePointMax(*request.Note, 500, "note"); err != nil {
			return err
		}
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(request.TradeDate)); err != nil {
		return fmt.Errorf("%w: tradeDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	if request.SettlementDate != nil {
		if _, err := time.Parse("2006-01-02", *request.SettlementDate); err != nil {
			return fmt.Errorf("%w: settlementDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	_, err := GrossFor(request)
	return err
}
