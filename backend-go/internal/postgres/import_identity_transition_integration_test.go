package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStoreRejectsMixedIdentityStrengthInBothOrders(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	tests := []struct {
		name        string
		firstStrong bool
	}{
		{name: "fallback then strong", firstStrong: false},
		{name: "strong then fallback", firstStrong: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			store, err := postgres.Open(databaseURL)
			if err != nil {
				t.Fatalf("open postgres store: %v", err)
			}
			defer store.Close()

			db, err := sql.Open("pgx", databaseURL)
			if err != nil {
				t.Fatalf("open identity transition database: %v", err)
			}
			closeDBOnCleanup(t, db, "identity transition")

			service := verticalslice.NewService(store, verticalslice.SystemClock{})
			ctx := context.Background()
			subjectID := uuid.NewString()
			portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "identity-transition-portfolio-key", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
				Name:         "Identity transition",
				BaseCurrency: verticalslice.RUB,
			})
			if err != nil {
				t.Fatalf("create portfolio: %v", err)
			}
			t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })

			fallback, strong := transitionRequests(t, portfolio.ID)
			first := fallback
			second := strong
			if testCase.firstStrong {
				first = strong
				second = fallback
			}

			_, err = service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "identity-transition-key-01", "/internal/imports/append", verticalslice.AppendImportBatchRequest{
				PortfolioID:        portfolio.ID,
				Transactions:       []verticalslice.AppendTransactionRequest{first},
				SourceKind:         "USER_UPLOADED_FILE",
				SourceAccountLabel: "broker-account-a",
				SourceFileHash:     strings.Repeat("a", 64),
			})
			if err != nil {
				t.Fatalf("append first identity: %v", err)
			}

			_, err = service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "identity-transition-key-02", "/internal/imports/append", verticalslice.AppendImportBatchRequest{
				PortfolioID:        portfolio.ID,
				Transactions:       []verticalslice.AppendTransactionRequest{second},
				SourceKind:         "USER_UPLOADED_FILE",
				SourceAccountLabel: "broker-account-a",
				SourceFileHash:     strings.Repeat("b", 64),
			})
			if !errors.Is(err, verticalslice.ErrInvalidInput) {
				t.Fatalf("expected mixed identity transition to fail closed, got %v", err)
			}

			listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
			if err != nil {
				t.Fatalf("list transactions: %v", err)
			}
			if len(listed) != 1 {
				t.Fatalf("expected exactly one canonical operation after rejected transition, got %d", len(listed))
			}
		})
	}
}

func TestStoreAllowsDistinctStrongIdentitiesWithSameFingerprint(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open strong identity database: %v", err)
	}
	closeDBOnCleanup(t, db, "strong identity")

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "strong-pair-portfolio-key", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Strong identity pair",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })

	_, strongA := transitionRequests(t, portfolio.ID)
	strongB := strongA
	copyProvenance := *strongA.ImportProvenance
	copyProvenance.BrokerOperationKey = verticalslice.BrokerOperationKey("broker-operation-b")
	strongB.ImportProvenance = &copyProvenance

	for index, request := range []verticalslice.AppendTransactionRequest{strongA, strongB} {
		_, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "strong-pair-import-key-0"+string(rune('1'+index)), "/internal/imports/append", verticalslice.AppendImportBatchRequest{
			PortfolioID:        portfolio.ID,
			Transactions:       []verticalslice.AppendTransactionRequest{request},
			SourceKind:         "USER_UPLOADED_FILE",
			SourceAccountLabel: "broker-account-a",
			SourceFileHash:     strings.Repeat(string(rune('c'+index)), 64),
		})
		if err != nil {
			t.Fatalf("append distinct strong identity %d: %v", index+1, err)
		}
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected two independently identified broker operations, got %d", len(listed))
	}
}

func TestDatabaseRejectsMixedIdentityStrengthInBothOrders(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open direct identity database: %v", err)
	}
	closeDBOnCleanup(t, db, "direct identity")
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()

	for _, firstStrong := range []bool{false, true} {
		name := "fallback then strong"
		if firstStrong {
			name = "strong then fallback"
		}
		t.Run(name, func(t *testing.T) {
			subjectID := uuid.NewString()
			portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "direct-identity-portfolio-"+uuid.NewString(), "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
				Name:         "Direct identity guard",
				BaseCurrency: verticalslice.RUB,
			})
			if err != nil {
				t.Fatalf("create portfolio: %v", err)
			}
			t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })
			_, strong := transitionRequests(t, portfolio.ID)
			fingerprint := strong.ImportProvenance.SourceFingerprint
			brokerKey := strong.ImportProvenance.BrokerOperationKey

			var firstKey any
			var secondKey any = brokerKey
			if firstStrong {
				firstKey = brokerKey
				secondKey = nil
			}
			if err := insertDirectIdentityRow(ctx, db, portfolio.ID, fingerprint, firstKey); err != nil {
				t.Fatalf("insert first identity: %v", err)
			}
			err = insertDirectIdentityRow(ctx, db, portfolio.ID, fingerprint, secondKey)
			if err == nil {
				t.Fatal("expected database to reject mixed fallback/strong identity")
			}
			if !strings.Contains(err.Error(), "transaction_entries_import_identity_strength_conflict") &&
				!strings.Contains(err.Error(), "mixed fallback/strong import identity") {
				t.Fatalf("expected Stage 3.27 identity-strength guard, got %v", err)
			}
		})
	}
}

func TestDatabaseSerializesConcurrentMixedIdentityStrength(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open concurrent identity database: %v", err)
	}
	closeDBOnCleanup(t, db, "concurrent identity")
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "concurrent-identity-portfolio", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Concurrent identity guard",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })
	_, strong := transitionRequests(t, portfolio.ID)
	fingerprint := strong.ImportProvenance.SourceFingerprint
	brokerKey := strong.ImportProvenance.BrokerOperationKey

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		errs <- insertDirectIdentityRow(ctx, db, portfolio.ID, fingerprint, nil)
	}()
	go func() {
		defer wg.Done()
		<-start
		errs <- insertDirectIdentityRow(ctx, db, portfolio.ID, fingerprint, brokerKey)
	}()
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	rejections := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if strings.Contains(err.Error(), "transaction_entries_import_identity_strength_conflict") ||
			strings.Contains(err.Error(), "mixed fallback/strong import identity") {
			rejections++
			continue
		}
		t.Fatalf("unexpected concurrent identity error: %v", err)
	}
	if successes != 1 || rejections != 1 {
		t.Fatalf("expected advisory guard to serialize one success and one rejection, got successes=%d rejections=%d", successes, rejections)
	}
}

func transitionRequests(t *testing.T, portfolioID string) (verticalslice.AppendTransactionRequest, verticalslice.AppendTransactionRequest) {
	t.Helper()
	gross := verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB}
	base := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-05",
	}
	fingerprint, err := verticalslice.NormalizedTransactionFingerprint(base)
	if err != nil {
		t.Fatalf("fingerprint identity transition: %v", err)
	}
	fallback := base
	fallback.ImportProvenance = &verticalslice.ImportProvenance{
		IdentityVersion:   verticalslice.ImportIdentityVersion,
		SourceFingerprint: fingerprint,
	}
	strong := base
	strong.ImportProvenance = &verticalslice.ImportProvenance{
		IdentityVersion:    verticalslice.ImportIdentityVersion,
		BrokerOperationKey: verticalslice.BrokerOperationKey("broker-operation-a"),
		SourceFingerprint:  fingerprint,
	}
	return fallback, strong
}

func insertDirectIdentityRow(ctx context.Context, db *sql.DB, portfolioID string, fingerprint string, brokerKey any) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, revision, transaction_type,
			gross_amount, commission_amount, tax_amount, trade_date,
			source_kind, source_file_hash, source_account_label,
			source_broker_operation_key, source_fingerprint, source_identity_version
		)
		VALUES (
			$1, $2, $3, 1, 'DEPOSIT',
			1000.00000000, 0.00000000, 0.00000000, '2026-07-05',
			'USER_UPLOADED_FILE', $4, 'broker-account-a',
			$5, $6, 1
		)
	`, uuid.NewString(), uuid.NewString(), portfolioID, strings.Repeat("f", 64), brokerKey, fingerprint)
	return err
}
