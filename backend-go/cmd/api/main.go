package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	developmentImportReviewTokenSecret = "openinvest-development-import-review-token-secret"
	runtimeIntegrityStartupTimeout     = 30 * time.Second
	gracefulShutdownTimeout            = 10 * time.Second
	apiListenAddress                   = ":8080"
)

type applicationRuntime struct {
	app   *fiber.App
	close func() error
}

func (runtime *applicationRuntime) Close() error {
	if runtime == nil || runtime.close == nil {
		return nil
	}
	return runtime.close()
}

func newRuntime() (runtime *applicationRuntime, err error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if err := validateRuntimeSafety(databaseURL); err != nil {
		return nil, fmt.Errorf("validate runtime safety: %w", err)
	}

	httpNetworkConfig, err := configuredHTTPNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("configure HTTP network boundary: %w", err)
	}

	corporateActionProvider, err := configuredTInvestCorporateActionProvider()
	if err != nil {
		return nil, fmt.Errorf("configure corporate action provider: %w", err)
	}

	if databaseURL == "" {
		store := unavailableStore{}
		return &applicationRuntime{
			app: httpapi.NewDevelopmentReplayWithCorporateActionProviderAndHTTPNetworkConfig(
				verticalslice.NewService(store, verticalslice.SystemClock{}),
				corporateActionProvider,
				httpNetworkConfig,
			),
		}, nil
	}

	runtimeCapability, err := postgres.ParseRuntimeCapabilityProfile(
		os.Getenv("OPENINVEST_RUNTIME_CAPABILITY_PROFILE"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse runtime capability profile: %w", err)
	}

	store, err := openPostgresStore(databaseURL, runtimeCapability)
	if err != nil {
		return nil, fmt.Errorf("open postgres store: %w", err)
	}

	defer func() {
		if err == nil {
			return
		}
		if closeErr := store.Close(); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("close postgres store after initialization failure: %w", closeErr),
			)
		}
	}()

	service, err := newValidatedRuntimeService(store)
	if err != nil {
		return nil, err
	}

	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:               []byte(os.Getenv("OPENINVEST_ACCESS_TOKEN_SECRET")),
		RefreshCookieSecure:             !envBool("OPENINVEST_REFRESH_COOKIE_INSECURE"),
		AllowDevelopmentBypass:          envBool("OPENINVEST_DEV_AUTH_BYPASS"),
		AllowEphemeralAccessTokenSecret: envBool("OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET") || envBool("OPENINVEST_DEV_AUTH_BYPASS"),
	})
	if err != nil {
		return nil, fmt.Errorf("initialize auth service: %w", err)
	}

	app, err := httpapi.NewReplayWithCorporateActionProviderAndHTTPNetworkConfig(
		service,
		authService,
		configuredImportReviewTokenSecret(),
		corporateActionProvider,
		httpNetworkConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize HTTP application: %w", err)
	}

	return &applicationRuntime{
		app:   app,
		close: store.Close,
	}, nil
}

type shutdownServer interface {
	ShutdownWithContext(context.Context) error
}

type httpLifecycle interface {
	shutdownServer
	Listener(net.Listener, ...fiber.ListenConfig) error
}

type readinessListener struct {
	net.Listener
	once  sync.Once
	ready chan struct{}
}

func (listener *readinessListener) Accept() (net.Conn, error) {
	listener.once.Do(func() {
		close(listener.ready)
	})
	return listener.Listener.Accept()
}

func shutdownHTTP(server shutdownServer, timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}

func serveHTTP(
	ctx context.Context,
	server httpLifecycle,
	listener net.Listener,
	shutdownTimeout time.Duration,
) error {
	if listener == nil {
		return errors.New("HTTP listener is required")
	}
	if shutdownTimeout <= 0 {
		_ = listener.Close()
		return errors.New("shutdown timeout must be positive")
	}
	if ctx.Err() != nil {
		if err := listener.Close(); err != nil {
			return fmt.Errorf("close listener before serve: %w", err)
		}
		return nil
	}

	ready := make(chan struct{})
	readyListener := &readinessListener{
		Listener: listener,
		ready:    ready,
	}
	serveErrCh := make(chan error, 1)

	go func() {
		serveErrCh <- server.Listener(readyListener)
	}()

	select {
	case err := <-serveErrCh:
		_ = listener.Close()
		if err != nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ready:
	}

	select {
	case err := <-serveErrCh:
		if err != nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	shutdownErr := shutdownHTTP(server, shutdownTimeout)
	if shutdownErr != nil {
		select {
		case serveErr := <-serveErrCh:
			if serveErr != nil {
				return errors.Join(
					shutdownErr,
					fmt.Errorf("serve HTTP after shutdown: %w", serveErr),
				)
			}
		default:
		}
		return shutdownErr
	}

	serveErr := <-serveErrCh
	if serveErr != nil {
		return fmt.Errorf("serve HTTP after shutdown: %w", serveErr)
	}
	return nil
}

func serveAPI(ctx context.Context, app *fiber.App) error {
	if ctx.Err() != nil {
		return nil
	}

	listener, err := net.Listen("tcp4", apiListenAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", apiListenAddress, err)
	}

	return serveHTTP(ctx, app, listener, gracefulShutdownTimeout)
}

type runtimeBuilder func() (*applicationRuntime, error)
type runtimeServer func(context.Context, *fiber.App) error

func runApplication(
	ctx context.Context,
	build runtimeBuilder,
	serve runtimeServer,
) (err error) {
	runtime, err := build()
	if err != nil {
		return fmt.Errorf("initialize runtime: %w", err)
	}
	if runtime == nil || runtime.app == nil {
		return errors.New("initialize runtime: nil application")
	}

	defer func() {
		if closeErr := runtime.Close(); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("close runtime resources: %w", closeErr),
			)
		}
	}()

	if err := serve(ctx, runtime.app); err != nil {
		return fmt.Errorf("run HTTP server: %w", err)
	}
	return nil
}

func openPostgresStore(databaseURL string, runtimeCapability postgres.RuntimeCapabilityProfile) (*postgres.Store, error) {
	if isExplicitDevelopmentEnvironment() {
		// Local development may use the schema owner for migration convenience. Staging and
		// production must prove the dedicated append-only runtime privilege boundary at startup.
		return postgres.OpenOwnerWithApplicationCapability(databaseURL, runtimeCapability)
	}
	return postgres.OpenRuntimeWithCapability(databaseURL, runtimeCapability)
}

func newValidatedRuntimeService(store verticalslice.Store) (*verticalslice.Service, error) {
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx, cancel := context.WithTimeout(context.Background(), runtimeIntegrityStartupTimeout)
	defer cancel()
	if err := service.ValidateRuntimeIntegrity(ctx); err != nil {
		return nil, fmt.Errorf("validate runtime integrity: %w", err)
	}
	return service, nil
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

func configuredHTTPNetworkConfig() (httpapi.HTTPNetworkConfig, error) {
	trustProxy, err := strictEnvBool("OPENINVEST_TRUST_PROXY")
	if err != nil {
		return httpapi.HTTPNetworkConfig{}, err
	}
	rawAllowlist := strings.TrimSpace(os.Getenv("OPENINVEST_TRUSTED_PROXY_CIDRS"))
	var allowlist []string
	if rawAllowlist != "" {
		allowlist = strings.Split(rawAllowlist, ",")
	}
	return httpapi.NewHTTPNetworkConfig(trustProxy, allowlist)
}

func strictEnvBool(name string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "", "0", "false", "no":
		return false, nil
	case "1", "true", "yes":
		return true, nil
	default:
		return false, fmt.Errorf("%s must be an explicit boolean", name)
	}
}

func runMain() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	return runApplication(ctx, newRuntime, serveAPI)
}

func main() {
	if err := runMain(); err != nil {
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
