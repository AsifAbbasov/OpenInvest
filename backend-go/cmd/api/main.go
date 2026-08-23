package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const developmentImportReviewTokenSecret = "openinvest-development-import-review-token-secret"

func newApp() *fiber.App {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if err := validateRuntimeSafety(databaseURL); err != nil {
		log.Fatal(err)
	}
	if databaseURL == "" {
		store := unavailableStore{}
		return httpapi.NewDevelopmentReplay(verticalslice.NewService(store, verticalslice.SystemClock{}))
	}
	store, err := openPostgresStore(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:               []byte(os.Getenv("OPENINVEST_ACCESS_TOKEN_SECRET")),
		RefreshCookieSecure:             !envBool("OPENINVEST_REFRESH_COOKIE_INSECURE"),
		AllowDevelopmentBypass:          envBool("OPENINVEST_DEV_AUTH_BYPASS"),
		AllowEphemeralAccessTokenSecret: envBool("OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET") || envBool("OPENINVEST_DEV_AUTH_BYPASS"),
	})
	if err != nil {
		log.Fatal(err)
	}
	app, err := httpapi.NewReplay(
		verticalslice.NewService(store, verticalslice.SystemClock{}),
		authService,
		configuredImportReviewTokenSecret(),
	)
	if err != nil {
		log.Fatal(err)
	}
	return app
}

func openPostgresStore(databaseURL string) (*postgres.Store, error) {
	if isExplicitDevelopmentEnvironment() {
		// Local development may use the schema owner for migration convenience. Staging and
		// production must prove the dedicated append-only runtime privilege boundary at startup.
		return postgres.Open(databaseURL)
	}
	return postgres.OpenRuntime(databaseURL)
}

func validateRuntimeSafety(databaseURL string) error {
	if strings.TrimSpace(databaseURL) == "" {
		if isExplicitDevelopmentEnvironment() {
			return nil
		}
		return errors.New("DATABASE_URL is required unless OPENINVEST_ENV=development or local")
	}
	if !(envBool("OPENINVEST_DEV_AUTH_BYPASS") ||
		envBool("OPENINVEST_REFRESH_COOKIE_INSECURE") ||
		envBool("OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET")) {
		return nil
	}
	if isExplicitDevelopmentEnvironment() {
		return nil
	}
	return errors.New("unsafe development auth settings require OPENINVEST_ENV=development or local")
}

func configuredImportReviewTokenSecret() []byte {
	if configured := strings.TrimSpace(os.Getenv("OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET")); configured != "" {
		return []byte(configured)
	}
	if isExplicitDevelopmentEnvironment() {
		return []byte(developmentImportReviewTokenSecret)
	}
	return nil
}

func isExplicitDevelopmentEnvironment() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("OPENINVEST_ENV"))) {
	case "development", "local":
		return true
	default:
		return false
	}
}

func main() {
	if err := newApp().Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

func envBool(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	default:
		return false
	}
}

type unavailableStore struct{}

func (unavailableStore) Ping(context.Context) error {
	return errors.New("database url is not configured")
}

func (unavailableStore) SearchAssets(context.Context, verticalslice.AssetSearchFilter) ([]verticalslice.AssetSummary, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) ListPortfolios(context.Context, string, verticalslice.PortfolioFilter) ([]verticalslice.Portfolio, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) CreatePortfolio(context.Context, verticalslice.CommandContext, verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, errors.New("database url is not configured")
}

func (unavailableStore) GetPortfolio(context.Context, string, string) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, errors.New("database url is not configured")
}

func (unavailableStore) ListTransactions(context.Context, string, string, verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) ListImportReviewTransactions(context.Context, string, string, verticalslice.ImportReviewHistoryFilter) ([]verticalslice.Transaction, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) AppendTransaction(context.Context, verticalslice.CommandContext, verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	return verticalslice.Transaction{}, errors.New("database url is not configured")
}

func (unavailableStore) AppendImportedTransactions(context.Context, verticalslice.CommandContext, verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) GetPortfolioSummary(context.Context, string, string, string) (verticalslice.PortfolioSummary, error) {
	return verticalslice.PortfolioSummary{
		TotalValue:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		CashValue:         verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		StockValue:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		BondValue:         verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		InvestedCapital:   verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		DividendsReceived: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		CouponsReceived:   verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
	}, errors.New("database url is not configured")
}

func (unavailableStore) RegisterUser(context.Context, auth.RegistrationRecord) (auth.StoredUser, error) {
	return auth.StoredUser{}, errors.New("database url is not configured")
}

func (unavailableStore) FindUserByEmail(context.Context, string) (auth.StoredUser, string, error) {
	return auth.StoredUser{}, "", errors.New("database url is not configured")
}

func (unavailableStore) CreateSession(context.Context, auth.SessionRecord) error {
	return errors.New("database url is not configured")
}

func (unavailableStore) RotateSession(context.Context, string, string, auth.SessionRecord, time.Time) (auth.StoredUser, error) {
	return auth.StoredUser{}, errors.New("database url is not configured")
}

func (unavailableStore) RevokeSession(context.Context, string, string, bool, time.Time) (bool, error) {
	return false, errors.New("database url is not configured")
}

func (unavailableStore) RecordAuthEvent(context.Context, auth.AuthAuditRecord) error {
	return errors.New("database url is not configured")
}
