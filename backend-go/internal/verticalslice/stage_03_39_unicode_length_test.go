package verticalslice

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type stage39ReplayLookupStore struct {
	recordingStore
	lookupArtifact    CommandReplayArtifact
	lookupFound       bool
	lookupErr         error
	lookupCommand     CommandContext
	lookupMethod      string
	lookupCalls       int
	createReplayCalls int
}

func (store *stage39ReplayLookupStore) LookupReplayArtifact(
	_ context.Context,
	command CommandContext,
	method string,
) (CommandReplayArtifact, bool, error) {
	store.lookupCalls++
	store.lookupCommand = command
	store.lookupMethod = method
	return store.lookupArtifact, store.lookupFound, store.lookupErr
}

func (store *stage39ReplayLookupStore) CreatePortfolioWithReplay(
	_ context.Context,
	_ CommandContext,
	request CreatePortfolioRequest,
	build PortfolioReplayBuilder,
) (Portfolio, CommandReplayArtifact, error) {
	store.createReplayCalls++
	portfolio := Portfolio{Name: request.Name, BaseCurrency: request.BaseCurrency, Version: 1}
	artifact, err := build(portfolio)
	return portfolio, artifact, err
}

func (store *stage39ReplayLookupStore) AppendTransactionWithReplay(
	context.Context,
	CommandContext,
	AppendTransactionRequest,
	TransactionReplayBuilder,
) (Transaction, CommandReplayArtifact, error) {
	return Transaction{}, CommandReplayArtifact{}, errors.New("not used")
}

func (store *stage39ReplayLookupStore) AppendImportedTransactionsWithReplay(
	context.Context,
	CommandContext,
	AppendImportBatchRequest,
	ImportedTransactionsReplayBuilder,
) ([]Transaction, CommandReplayArtifact, error) {
	return nil, CommandReplayArtifact{}, errors.New("not used")
}

func TestStage339CreatePortfolioUsesLiteralUnicodeCodePointBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{name: "100 ASCII", value: strings.Repeat("A", 100)},
		{name: "101 ASCII", value: strings.Repeat("A", 101), wantError: true},
		{name: "100 Cyrillic", value: strings.Repeat("Ж", 100)},
		{name: "101 Cyrillic", value: strings.Repeat("Ж", 101), wantError: true},
		{name: "100 supplementary", value: strings.Repeat("😀", 100)},
		{name: "101 supplementary", value: strings.Repeat("😀", 101), wantError: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(&recordingStore{}, fixedClock{})
			_, err := service.CreatePortfolio(
				context.Background(),
				RequestContext{},
				"subject",
				"stage-03-39-portfolio-key-0001",
				"/api/v1/portfolios",
				CreatePortfolioRequest{Name: tc.value, BaseCurrency: RUB},
			)
			if tc.wantError {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected invalid input, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected value to be accepted, got %v", err)
			}
		})
	}
}

func TestStage339CreatePortfolioFreshAdmissionChecksRawBeforeTrim(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	_, err := service.CreatePortfolio(
		context.Background(),
		RequestContext{},
		"subject",
		"stage-03-39-portfolio-key-0002",
		"/api/v1/portfolios",
		CreatePortfolioRequest{Name: " " + strings.Repeat("A", 100), BaseCurrency: RUB},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected raw 101-code-point name to fail before trim, got %v", err)
	}
}

func TestStage339CreatePortfolioRejectsMalformedInternalUTF8(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	_, err := service.CreatePortfolio(
		context.Background(),
		RequestContext{},
		"subject",
		"stage-03-39-portfolio-key-0003",
		"/api/v1/portfolios",
		CreatePortfolioRequest{Name: string([]byte{0xff, 'A'}), BaseCurrency: RUB},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected malformed UTF-8 to fail closed, got %v", err)
	}
}

func TestStage339PortfolioHistoricalReplayUsesLegacyTrimFirstHashWithoutSecondEffect(t *testing.T) {
	expectedArtifact := CommandReplayArtifact{
		StatusCode: 201,
		Body:       []byte(`{"data":{"id":"historical"}}`),
		RequestID:  "11111111-1111-4111-8111-111111111111",
		TraceID:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	store := &stage39ReplayLookupStore{lookupArtifact: expectedArtifact, lookupFound: true}
	service := NewService(store, fixedClock{})
	requestContext := RequestContext{
		RequestID: "22222222-2222-4222-8222-222222222222",
		TraceID:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	key := "stage-03-39-historical-key-0001"
	path := "/api/v1/portfolios"
	normalized := CreatePortfolioRequest{Name: strings.Repeat("A", 100), BaseCurrency: RUB}
	expectedCommand, err := service.command(requestContext, "subject", key, path, normalized)
	if err != nil {
		t.Fatalf("build expected historical command: %v", err)
	}

	builderCalls := 0
	_, artifact, err := service.CreatePortfolioWithReplay(
		context.Background(),
		requestContext,
		"subject",
		key,
		path,
		CreatePortfolioRequest{Name: " " + normalized.Name, BaseCurrency: RUB},
		func(Portfolio) (CommandReplayArtifact, error) {
			builderCalls++
			return CommandReplayArtifact{}, nil
		},
	)
	if err != nil {
		t.Fatalf("historical replay: %v", err)
	}
	if store.lookupCalls != 1 || store.lookupMethod != "POST" {
		t.Fatalf("expected one POST read-only lookup, calls=%d method=%q", store.lookupCalls, store.lookupMethod)
	}
	if store.lookupCommand.RequestHash != expectedCommand.RequestHash {
		t.Fatalf("historical request hash changed: got %s want %s", store.lookupCommand.RequestHash, expectedCommand.RequestHash)
	}
	if artifact.StatusCode != expectedArtifact.StatusCode ||
		string(artifact.Body) != string(expectedArtifact.Body) ||
		artifact.RequestID != expectedArtifact.RequestID ||
		artifact.TraceID != expectedArtifact.TraceID {
		t.Fatalf("exact historical artifact changed: got %+v want %+v", artifact, expectedArtifact)
	}
	if builderCalls != 0 || store.createReplayCalls != 0 {
		t.Fatalf("historical replay must create no second effect: builder=%d createReplay=%d", builderCalls, store.createReplayCalls)
	}
}

func TestStage339PortfolioExpiredOrMissingHistoricalAuthorityCannotCreateNewGeneration(t *testing.T) {
	store := &stage39ReplayLookupStore{lookupFound: false}
	service := NewService(store, fixedClock{})
	builderCalls := 0
	_, _, err := service.CreatePortfolioWithReplay(
		context.Background(),
		RequestContext{},
		"subject",
		"stage-03-39-expired-key-0001",
		"/api/v1/portfolios",
		CreatePortfolioRequest{Name: " " + strings.Repeat("A", 100), BaseCurrency: RUB},
		func(Portfolio) (CommandReplayArtifact, error) {
			builderCalls++
			return CommandReplayArtifact{}, nil
		},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected fresh Stage 3.39 validation after no historical authority, got %v", err)
	}
	if store.lookupCalls != 1 || store.createReplayCalls != 0 || builderCalls != 0 {
		t.Fatalf("expired/missing authority must not reserve or create: lookup=%d createReplay=%d builder=%d", store.lookupCalls, store.createReplayCalls, builderCalls)
	}
}

func TestStage339PortfolioHistoricalLookupErrorsFailClosed(t *testing.T) {
	sentinel := errors.New("lookup authority failure")
	store := &stage39ReplayLookupStore{lookupErr: sentinel}
	service := NewService(store, fixedClock{})
	_, _, err := service.CreatePortfolioWithReplay(
		context.Background(),
		RequestContext{},
		"subject",
		"stage-03-39-failclosed-key-1",
		"/api/v1/portfolios",
		CreatePortfolioRequest{Name: " " + strings.Repeat("A", 100), BaseCurrency: RUB},
		func(Portfolio) (CommandReplayArtifact, error) { return CommandReplayArtifact{}, nil },
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected lookup failure to propagate, got %v", err)
	}
	if store.createReplayCalls != 0 {
		t.Fatalf("lookup failure must stop before any write-capable replay store call")
	}
}

func TestStage339SearchAssetsChecksRawCodePointsBeforeTrimAndRejectsMalformedUTF8(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	if _, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: strings.Repeat("😀", 100)}); err != nil {
		t.Fatalf("100 supplementary query code points must be accepted: %v", err)
	}
	if _, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: strings.Repeat("😀", 101)}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("101 supplementary query code points must be rejected, got %v", err)
	}
	if _, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: " " + strings.Repeat("A", 100)}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("raw over-limit query must fail before trim, got %v", err)
	}
	if _, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: string([]byte{0xff})}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("malformed query must fail closed, got %v", err)
	}
}

func TestStage339TransactionNoteUsesCodePointsAndRejectsMalformedUTF8(t *testing.T) {
	tests := []struct {
		name      string
		note      string
		wantError bool
	}{
		{name: "500 supplementary", note: strings.Repeat("😀", 500)},
		{name: "501 supplementary", note: strings.Repeat("😀", 501), wantError: true},
		{name: "malformed", note: string([]byte{0xff}), wantError: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingStore{}
			service := NewService(store, fixedClock{})
			gross := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}
			note := tc.note
			_, err := service.AppendTransaction(
				context.Background(),
				RequestContext{},
				"subject",
				"stage-03-39-note-key-00001",
				"/api/v1/portfolios/p/transactions",
				AppendTransactionRequest{
					PortfolioID:     "portfolio-id",
					TransactionType: "DEPOSIT",
					GrossAmount:     &gross,
					Commission:      ZeroMoney(),
					Tax:             ZeroMoney(),
					TradeDate:       "2026-08-27",
					Note:            &note,
				},
			)
			if tc.wantError {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected invalid input, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected note to be accepted, got %v", err)
			}
		})
	}
}

func TestStage339AppendSourceLabelUsesCodePointsAndPreservesTrimmedIdentity(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantLabel string
		wantError bool
	}{
		{name: "120 Cyrillic", label: strings.Repeat("Ж", 120), wantLabel: strings.Repeat("Ж", 120)},
		{name: "121 Cyrillic", label: strings.Repeat("Ж", 121), wantError: true},
		{name: "120 supplementary", label: strings.Repeat("😀", 120), wantLabel: strings.Repeat("😀", 120)},
		{name: "121 supplementary", label: strings.Repeat("😀", 121), wantError: true},
		{name: "trim identity", label: " Broker A ", wantLabel: "Broker A"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingStore{}
			service := NewService(store, fixedClock{})
			request := stage39ValidImportBatch(t, tc.label)
			_, err := service.AppendImportedTransactions(
				context.Background(),
				RequestContext{},
				"subject",
				"stage-03-39-import-key-0001",
				"/internal/imports/append",
				request,
			)
			if tc.wantError {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected invalid input, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("append imported transaction: %v", err)
			}
			if store.importRequest.SourceAccountLabel != tc.wantLabel {
				t.Fatalf("source label identity changed: got %q want %q", store.importRequest.SourceAccountLabel, tc.wantLabel)
			}
		})
	}
}

func stage39ValidImportBatch(t *testing.T, label string) AppendImportBatchRequest {
	t.Helper()
	gross := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}
	transaction := AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-08-27",
	}
	fingerprint, err := NormalizedTransactionFingerprint(transaction)
	if err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	transaction.ImportProvenance = &ImportProvenance{
		IdentityVersion:   ImportIdentityVersion,
		SourceFingerprint: fingerprint,
	}
	return AppendImportBatchRequest{
		PortfolioID:        "portfolio-id",
		Transactions:       []AppendTransactionRequest{transaction},
		SourceKind:         "USER_UPLOADED_FILE",
		SourceAccountLabel: label,
		SourceFileHash:     strings.Repeat("a", 64),
	}
}
