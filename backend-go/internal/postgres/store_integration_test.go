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
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const importCSVHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

type assetRowSnapshot struct {
	Exists          bool
	ID              string
	AssetType       string
	Name            string
	Currency        string
	Market          string
	LifecycleStatus string
	ISIN            sql.NullString
	LotSize         string
	UpdatedAt       string
}

func closeDBOnCleanup(t *testing.T, db *sql.DB, label string) {
	t.Helper()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close %s db: %v", label, err)
		}
	})
}

func snapshotAssetRow(t *testing.T, ctx context.Context, db *sql.DB, ticker string) assetRowSnapshot {
	t.Helper()
	var snapshot assetRowSnapshot
	err := db.QueryRowContext(ctx, `
		SELECT id, asset_type, name, currency, market, lifecycle_status, isin, lot_size::text, updated_at::text
		FROM investment.assets
		WHERE ticker = $1
	`, ticker).Scan(
		&snapshot.ID,
		&snapshot.AssetType,
		&snapshot.Name,
		&snapshot.Currency,
		&snapshot.Market,
		&snapshot.LifecycleStatus,
		&snapshot.ISIN,
		&snapshot.LotSize,
		&snapshot.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return snapshot
	}
	if err != nil {
		t.Fatalf("snapshot existing asset %s: %v", ticker, err)
	}
	snapshot.Exists = true
	return snapshot
}

func restoreAssetRow(t *testing.T, ctx context.Context, db *sql.DB, ticker string, snapshot assetRowSnapshot) {
	t.Helper()

	if !snapshot.Exists {
		if _, err := db.ExecContext(ctx, `DELETE FROM investment.assets WHERE ticker = $1`, ticker); err != nil {
			t.Fatalf("delete restored absent asset %s: %v", ticker, err)
		}
		restored := snapshotAssetRow(t, ctx, db, ticker)
		if restored.Exists {
			t.Fatalf("expected restored asset %s to be absent, got %+v", ticker, restored)
		}
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::numeric, $10::timestamptz)
		ON CONFLICT (ticker) DO UPDATE SET
			id = EXCLUDED.id,
			asset_type = EXCLUDED.asset_type,
			name = EXCLUDED.name,
			currency = EXCLUDED.currency,
			market = EXCLUDED.market,
			lifecycle_status = EXCLUDED.lifecycle_status,
			isin = EXCLUDED.isin,
			lot_size = EXCLUDED.lot_size,
			updated_at = EXCLUDED.updated_at
	`, snapshot.ID, ticker, snapshot.AssetType, snapshot.Name, snapshot.Currency, snapshot.Market,
		snapshot.LifecycleStatus, snapshot.ISIN, snapshot.LotSize, snapshot.UpdatedAt); err != nil {
		t.Fatalf("restore asset %s: %v", ticker, err)
	}

	restored := snapshotAssetRow(t, ctx, db, ticker)
	if restored.ID != snapshot.ID ||
		restored.AssetType != snapshot.AssetType ||
		restored.Name != snapshot.Name ||
		restored.Currency != snapshot.Currency ||
		restored.Market != snapshot.Market ||
		restored.LifecycleStatus != snapshot.LifecycleStatus ||
		restored.ISIN != snapshot.ISIN ||
		restored.LotSize != snapshot.LotSize ||
		restored.UpdatedAt != snapshot.UpdatedAt {
		t.Fatalf("asset %s restore mismatch: got %+v want %+v", ticker, restored, snapshot)
	}
}

func cleanupPortfolioRows(t *testing.T, ctx context.Context, db *sql.DB, portfolioID string) {
	t.Helper()

	statements := []struct {
		name  string
		query string
	}{
		{
			name: "snapshot positions",
			query: `
				DELETE FROM analytics.snapshot_positions
				WHERE snapshot_id IN (
					SELECT id FROM analytics.portfolio_snapshots WHERE portfolio_id = $1
				)
			`,
		},
		{name: "portfolio snapshots", query: `DELETE FROM analytics.portfolio_snapshots WHERE portfolio_id = $1`},
		{name: "calculation runs", query: `DELETE FROM analytics.calculation_runs WHERE portfolio_id = $1`},
		{name: "transaction entries", query: `DELETE FROM investment.transaction_entries WHERE portfolio_id = $1`},
		{name: "portfolio", query: `DELETE FROM investment.portfolios WHERE id = $1`},
	}

	for _, statement := range statements {
		result, err := db.ExecContext(ctx, statement.query, portfolioID)
		if err != nil {
			t.Fatalf("cleanup %s for portfolio %s: %v", statement.name, portfolioID, err)
		}
		if statement.name == "portfolio" {
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				t.Fatalf("read cleanup portfolio rows affected for %s: %v", portfolioID, err)
			}
			if rowsAffected != 1 {
				t.Fatalf("expected to cleanup 1 portfolio row for %s, cleaned %d", portfolioID, rowsAffected)
			}
		}
	}
}

func TestStoreVerticalSlice(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-000001", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Long-term capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	ticker := "SBER"
	quantity := decimal.Must("100.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("280.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000001", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.Money{Amount: decimal.Must("28.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-10",
	})
	if err != nil {
		t.Fatalf("append transaction: %v", err)
	}

	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000002", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.Money{Amount: decimal.Must("28.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-10",
	})
	if err != nil {
		t.Fatalf("append second same-day transaction: %v", err)
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "")
	if err != nil {
		t.Fatalf("get portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "56000.00000000" {
		t.Fatalf("expected stock value 56000.00000000, got %s", got)
	}

	_, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-09")
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("expected no summary before first trade business date, got %v", err)
	}

	backdatedQuantity := decimal.Must("10.00000000")
	backdatedPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000003", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &backdatedQuantity,
		UnitPrice:       &backdatedPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-05",
	})
	if err != nil {
		t.Fatalf("append backdated transaction: %v", err)
	}

	summary, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-05")
	if err != nil {
		t.Fatalf("get backdated portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "1000.00000000" {
		t.Fatalf("expected stock value 1000.00000000 for 2026-01-05, got %s", got)
	}

	summary, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-10")
	if err != nil {
		t.Fatalf("get portfolio summary after backdated transaction: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "57000.00000000" {
		t.Fatalf("expected stock value 57000.00000000 after backdated rebuild, got %s", got)
	}

	filtered, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{
		TransactionType: "BUY",
		FromDate:        "2026-01-06",
		ToDate:          "2026-01-10",
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("list filtered transactions: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered transactions, got %d", len(filtered))
	}
}

func TestStoreAppendTransactionRejectsUnsupportedTicker(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-catalog-001", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Catalog validation",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	ticker := "UNKNOWN1"
	quantity := decimal.Must("1.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-catalog-001", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-01",
	})
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected unsupported ticker rejection, got %v", err)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected unsupported ticker append to leave no transactions, got %d", len(listed))
	}
}

func TestStoreAppendTransactionSeedsApprovedBondFixture(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-catalog-002", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Bond catalog validation",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	ticker := "SU26238RMFS4"
	quantity := decimal.Must("1.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("700.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-catalog-002", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-02",
	})
	if err != nil {
		t.Fatalf("append approved bond transaction: %v", err)
	}
	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-07-02")
	if err != nil {
		t.Fatalf("get bond summary: %v", err)
	}
	if got := summary.BondValue.Amount.String(); got != "700.00000000" {
		t.Fatalf("expected bond value 700.00000000, got %s", got)
	}
	if got := summary.StockValue.Amount.String(); got != "0.00000000" {
		t.Fatalf("expected stock value 0.00000000 for bond transaction, got %s", got)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open catalog check db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	var assetID, assetType, name, currency, market, lifecycleStatus string
	var isin sql.NullString
	var lotSize string
	if err := db.QueryRowContext(ctx, `
		SELECT id, asset_type, name, currency, market, lifecycle_status, isin, lot_size::text
		FROM investment.assets
		WHERE ticker = $1
	`, ticker).Scan(&assetID, &assetType, &name, &currency, &market, &lifecycleStatus, &isin, &lotSize); err != nil {
		t.Fatalf("query approved bond fixture: %v", err)
	}
	if assetID != "00000000-0000-4000-8000-00000000b001" ||
		assetType != "bond" ||
		name != "OFZ 26238" ||
		currency != "RUB" ||
		market != "MOEX" ||
		lifecycleStatus != "active" ||
		!isin.Valid ||
		isin.String != "RU000A1038V6" ||
		lotSize != "1.00000000" {
		t.Fatalf("expected approved bond fixture metadata, got id=%s type=%s name=%q currency=%s market=%s lifecycle=%s isin=%v lotSize=%s",
			assetID, assetType, name, currency, market, lifecycleStatus, isin, lotSize)
	}
}

func TestStoreSearchAssetsReturnsOnlyActiveCanonicalFixtures(t *testing.T) {
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
		t.Fatalf("open catalog setup db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	ctx := context.Background()
	tickers := []string{"SBER", "GAZP", "SU26238RMFS4"}
	previousRows := make(map[string]assetRowSnapshot, len(tickers))
	for _, ticker := range tickers {
		previousRows[ticker] = snapshotAssetRow(t, ctx, db, ticker)
	}
	t.Cleanup(func() {
		for _, ticker := range tickers {
			restoreAssetRow(t, ctx, db, ticker, previousRows[ticker])
		}
	})

	sberID := "00000000-0000-4000-8000-00000000a001"
	if previousRows["SBER"].Exists {
		sberID = previousRows["SBER"].ID
	}
	gazpID := "00000000-0000-4000-8000-00000000a002"
	if previousRows["GAZP"].Exists {
		gazpID = previousRows["GAZP"].ID
	}
	bondID := "00000000-0000-4000-8000-00000000b001"
	if previousRows["SU26238RMFS4"].Exists {
		bondID = previousRows["SU26238RMFS4"].ID
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES
			($1, 'SBER', 'stock', 'Sberbank ordinary shares', 'RUB', 'MOEX', 'active', 'RU0009029540', 10::numeric, now()),
			($2, 'SU26238RMFS4', 'bond', 'OFZ 26238', 'RUB', 'MOEX', 'active', 'RU000A1038V6', 1::numeric, now()),
			($3, 'GAZP', 'stock', 'Gazprom ordinary shares', 'RUB', 'MOEX', 'active', 'RU0000000000', 10::numeric, now())
		ON CONFLICT (ticker) DO UPDATE SET
			asset_type = EXCLUDED.asset_type,
			name = EXCLUDED.name,
			currency = EXCLUDED.currency,
			market = EXCLUDED.market,
			lifecycle_status = EXCLUDED.lifecycle_status,
			isin = EXCLUDED.isin,
			lot_size = EXCLUDED.lot_size,
			updated_at = EXCLUDED.updated_at
	`, sberID, bondID, gazpID)
	if err != nil {
		t.Fatalf("seed search catalog rows: %v", err)
	}

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	assets, err := service.SearchAssets(ctx, verticalslice.AssetSearchFilter{Query: "S", Limit: 20})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}
	if len(assets.Items) != 2 {
		t.Fatalf("expected SBER and SU26238RMFS4, got %+v", assets)
	}
	if assets.Items[0].Ticker != "SBER" || assets.Items[0].AssetType != "STOCK" || assets.Items[0].LastPrice != nil {
		t.Fatalf("unexpected first asset summary: %+v", assets.Items[0])
	}
	if assets.Items[1].Ticker != "SU26238RMFS4" || assets.Items[1].AssetType != "BOND" || assets.Items[1].LastPrice != nil {
		t.Fatalf("unexpected second asset summary: %+v", assets.Items[1])
	}

	stocks, err := service.SearchAssets(ctx, verticalslice.AssetSearchFilter{Query: "S", AssetType: "STOCK", Limit: 20})
	if err != nil {
		t.Fatalf("search stock assets: %v", err)
	}
	if len(stocks.Items) != 1 || stocks.Items[0].Ticker != "SBER" {
		t.Fatalf("expected only active SBER stock, got %+v", stocks)
	}

	tickerContainsOnly, err := service.SearchAssets(ctx, verticalslice.AssetSearchFilter{Query: "RMFS", Limit: 20})
	if err != nil {
		t.Fatalf("search ticker contains-only fragment: %v", err)
	}
	if len(tickerContainsOnly.Items) != 0 {
		t.Fatalf("expected ticker contains-only fragment not to match ticker without name fragment, got %+v", tickerContainsOnly)
	}

	firstPage, err := service.SearchAssets(ctx, verticalslice.AssetSearchFilter{Query: "S", Limit: 1})
	if err != nil {
		t.Fatalf("search first page: %v", err)
	}
	if !firstPage.HasMore || firstPage.NextCursor == nil || len(firstPage.Items) != 1 || firstPage.Items[0].Ticker != "SBER" {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}
	secondPage, err := service.SearchAssets(ctx, verticalslice.AssetSearchFilter{Query: "S", Cursor: *firstPage.NextCursor, Limit: 1})
	if err != nil {
		t.Fatalf("search second page: %v", err)
	}
	if secondPage.HasMore || secondPage.NextCursor != nil || len(secondPage.Items) != 1 || secondPage.Items[0].Ticker != "SU26238RMFS4" {
		t.Fatalf("unexpected second page: %+v", secondPage)
	}
}

func TestStoreAppendTransactionAcceptsCanonicalFixtureWithLegacyID(t *testing.T) {
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
		t.Fatalf("open catalog setup db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	ctx := context.Background()
	ticker := "GAZP"
	previousAsset := snapshotAssetRow(t, ctx, db, ticker)
	t.Cleanup(func() {
		restoreAssetRow(t, ctx, db, ticker, previousAsset)
	})

	legacyAssetID := uuid.NewString()
	_, err = db.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, $2, 'stock', 'Gazprom ordinary shares', 'RUB', 'MOEX', 'active', 'RU0007661625', 10::numeric, now())
		ON CONFLICT (ticker) DO UPDATE SET
			id = EXCLUDED.id,
			asset_type = EXCLUDED.asset_type,
			name = EXCLUDED.name,
			currency = EXCLUDED.currency,
			market = EXCLUDED.market,
			lifecycle_status = EXCLUDED.lifecycle_status,
			isin = EXCLUDED.isin,
			lot_size = EXCLUDED.lot_size,
			updated_at = EXCLUDED.updated_at
	`, legacyAssetID, ticker)
	if err != nil {
		t.Fatalf("seed legacy-id canonical asset: %v", err)
	}

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-catalog-legacy-id", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Legacy asset identity catalog validation",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() {
		cleanupPortfolioRows(t, ctx, db, portfolio.ID)
	})

	quantity := decimal.Must("1.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("200.00000000"), Currency: verticalslice.RUB}
	transaction, err := service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-catalog-legacy-id", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-02",
	})
	if err != nil {
		t.Fatalf("append canonical legacy-id asset transaction: %v", err)
	}

	var storedAssetID string
	if err := db.QueryRowContext(ctx, `
		SELECT asset_id::text
		FROM investment.transaction_entries
		WHERE transaction_id = $1
	`, transaction.ID).Scan(&storedAssetID); err != nil {
		t.Fatalf("query transaction asset id: %v", err)
	}
	if storedAssetID != legacyAssetID {
		t.Fatalf("expected existing canonical legacy asset id %s, got %s", legacyAssetID, storedAssetID)
	}
}

func TestStoreAppendTransactionDoesNotReactivateInactiveFixture(t *testing.T) {
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
		t.Fatalf("open catalog setup db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	ctx := context.Background()
	ticker := "GAZP"
	previousAsset := snapshotAssetRow(t, ctx, db, ticker)
	t.Cleanup(func() {
		restoreAssetRow(t, ctx, db, ticker, previousAsset)
	})

	assetID := "00000000-0000-4000-8000-00000000a002"
	_, err = db.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, $2, 'stock', 'Inactive Gazprom fixture', 'RUB', 'MOEX', 'inactive', 'RU0007661625', 10::numeric, now())
		ON CONFLICT (ticker) DO UPDATE SET
			lifecycle_status = 'inactive',
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`, assetID, ticker)
	if err != nil {
		t.Fatalf("seed inactive asset: %v", err)
	}

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-catalog-003", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Inactive catalog validation",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() {
		cleanupPortfolioRows(t, ctx, db, portfolio.ID)
	})

	quantity := decimal.Must("1.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-catalog-003", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-03",
	})
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected inactive fixture to be rejected, got %v", err)
	}

	var lifecycleStatus string
	if err := db.QueryRowContext(ctx, `
		SELECT lifecycle_status
		FROM investment.assets
		WHERE ticker = $1
	`, ticker).Scan(&lifecycleStatus); err != nil {
		t.Fatalf("query inactive asset lifecycle: %v", err)
	}
	if lifecycleStatus != "inactive" {
		t.Fatalf("expected inactive asset to remain inactive, got %s", lifecycleStatus)
	}
}

func TestStoreAppendTransactionRejectsConflictingActiveFixture(t *testing.T) {
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
		t.Fatalf("open catalog setup db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	ctx := context.Background()
	ticker := "SU26238RMFS4"
	previousAsset := snapshotAssetRow(t, ctx, db, ticker)
	t.Cleanup(func() {
		restoreAssetRow(t, ctx, db, ticker, previousAsset)
	})

	conflictingAssetID := uuid.NewString()
	if previousAsset.Exists {
		conflictingAssetID = previousAsset.ID
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, $2, 'stock', 'Legacy generic bond-as-stock row', 'RUB', 'MOEX', 'active', 'RU000A1038V6', 1::numeric, now())
		ON CONFLICT (ticker) DO UPDATE SET
			id = EXCLUDED.id,
			asset_type = EXCLUDED.asset_type,
			name = EXCLUDED.name,
			currency = EXCLUDED.currency,
			market = EXCLUDED.market,
			lifecycle_status = EXCLUDED.lifecycle_status,
			isin = EXCLUDED.isin,
			lot_size = EXCLUDED.lot_size,
			updated_at = EXCLUDED.updated_at
	`, conflictingAssetID, ticker)
	if err != nil {
		t.Fatalf("seed conflicting active asset: %v", err)
	}

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-catalog-004", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Conflicting catalog validation",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() {
		cleanupPortfolioRows(t, ctx, db, portfolio.ID)
	})

	quantity := decimal.Must("1.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("700.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-catalog-004", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-07-04",
	})
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected conflicting active fixture to be rejected, got %v", err)
	}
}

func TestStoreAppendImportedTransactionsIsAtomicAndIdempotent(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-01", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Imported capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	depositGross := verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB}
	ticker := "SBER"
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	batch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount:     &depositGross,
				Commission:      verticalslice.ZeroMoney(),
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-19",
			},
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "BUY",
				Ticker:          &ticker,
				Quantity:        &quantity,
				UnitPrice:       &unitPrice,
				Commission:      verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB},
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-20",
			},
		},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash",
	}

	transactions, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-001", "/internal/imports/append", batch)
	if err != nil {
		t.Fatalf("append imported transactions: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 imported transactions, got %d", len(transactions))
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-06-20")
	if err != nil {
		t.Fatalf("get portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "200.00000000" {
		t.Fatalf("expected stock value 200.00000000, got %s", got)
	}
	if got := summary.CashValue.Amount.String(); got != "799.00000000" {
		t.Fatalf("expected cash value 799.00000000, got %s", got)
	}

	replayed, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-001", "/internal/imports/append", batch)
	if err != nil {
		t.Fatalf("replay imported transactions: %v", err)
	}
	if len(replayed) != 2 {
		t.Fatalf("expected 2 replayed imported transactions, got %d", len(replayed))
	}
	if replayed[0].ID != transactions[0].ID || replayed[1].ID != transactions[1].ID {
		t.Fatalf("expected replay to return original transactions, got %v want %v", replayed, transactions)
	}

	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "manual-equivalent-after-import", "/api/v1/portfolios/"+portfolio.ID+"/transactions", batch.Transactions[1])
	if err != nil {
		t.Fatalf("append later equivalent manual transaction: %v", err)
	}

	replayedAfterEquivalent, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-001", "/internal/imports/append", batch)
	if err != nil {
		t.Fatalf("replay imported transactions after later equivalent: %v", err)
	}
	if replayedAfterEquivalent[0].ID != transactions[0].ID || replayedAfterEquivalent[1].ID != transactions[1].ID {
		t.Fatalf("expected replay after later equivalent to return original transactions, got %v want %v", replayedAfterEquivalent, transactions)
	}

	secondDeposit := verticalslice.Money{Amount: decimal.Must("500.00000000"), Currency: verticalslice.RUB}
	conflictingBatch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount:     &secondDeposit,
				Commission:      verticalslice.ZeroMoney(),
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-21",
			},
			batch.Transactions[1],
		},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash-2",
	}

	_, err = service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-002", "/internal/imports/append", conflictingBatch)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected duplicate conflict for import batch, got %v", err)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("expected rollback to keep 3 transactions, got %d", len(listed))
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open audit check db: %v", err)
	}
	closeDBOnCleanup(t, db, "test")

	var auditCount int
	err = db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM audit.events
		WHERE target_id = $1 AND action_code = 'IMPORT_APPEND_BATCH' AND outcome = 'success'
	`, portfolio.ID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit evidence: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one import append audit event, got %d", auditCount)
	}
}

func TestStoreAppendImportedTransactionsSerializesConcurrentDuplicateBatches(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-02", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Concurrent imported capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	gross := verticalslice.Money{Amount: decimal.Must("750.00000000"), Currency: verticalslice.RUB}
	batch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{{
			PortfolioID:     portfolio.ID,
			TransactionType: "DEPOSIT",
			GrossAmount:     &gross,
			Commission:      verticalslice.ZeroMoney(),
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       "2026-06-23",
		}},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash-concurrent",
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"import-batch-key-101", "import-batch-key-102"} {
		wg.Add(1)
		go func(idempotencyKey string) {
			defer wg.Done()
			<-start
			_, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, idempotencyKey, "/internal/imports/append", batch)
			errs <- err
		}(key)
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	duplicates := 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, verticalslice.ErrInvalidInput):
			duplicates++
		default:
			t.Fatalf("unexpected concurrent append error: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("expected one success and one duplicate rejection, got successes=%d duplicates=%d", successes, duplicates)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected exactly one imported transaction after concurrent batches, got %d", len(listed))
	}
}

func TestImportReviewAppendFlowAppendsApprovedRows(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-flow-01", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Review append flow",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	result, err := importflow.ReviewAndAppend(ctx, service, importflow.Request{
		SubjectID:          subjectID,
		PortfolioID:        portfolio.ID,
		IdempotencyKey:     "import-flow-key-1001",
		RequestPath:        "/internal/imports/review-append",
		SourceAccountLabel: "manual-broker-label",
		Reader: strings.NewReader(importCSVHeader +
			"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-flow-1,cash in\n" +
			"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,2026-06-21,RUB,op-flow-2,buy\n"),
		Decisions: []importer.Decision{
			{RowNumber: 2, Action: importer.DecisionApprove},
			{RowNumber: 3, Action: importer.DecisionApprove},
		},
	})
	if err != nil {
		t.Fatalf("review and append flow: %v", err)
	}
	if result.ParsedRowCount != 2 || result.AcceptedRowCount != 2 || result.NonAppendedRowCount != 0 {
		t.Fatalf("unexpected result counts: %+v", result)
	}
	if len(result.AppendedTransactionIDs) != 2 {
		t.Fatalf("expected 2 appended transaction ids, got %v", result.AppendedTransactionIDs)
	}
	if got := result.SnapshotDatesRebuilt; len(got) != 2 || got[0] != "2026-06-19" || got[1] != "2026-06-20" {
		t.Fatalf("unexpected snapshot rebuild dates: %v", got)
	}
	if result.AuditActionCode != "IMPORT_APPEND_BATCH" {
		t.Fatalf("unexpected audit action code: %s", result.AuditActionCode)
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-06-20")
	if err != nil {
		t.Fatalf("get portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "200.00000000" {
		t.Fatalf("expected stock value 200.00000000, got %s", got)
	}
	if got := summary.CashValue.Amount.String(); got != "799.00000000" {
		t.Fatalf("expected cash value 799.00000000, got %s", got)
	}
}

func TestImportReviewAppendFlowRejectsStaleDuplicateWithoutPartialAppend(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-flow-02", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Stale duplicate flow",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	duplicateGross := verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-import-flow-dup", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "DEPOSIT",
		GrossAmount:     &duplicateGross,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-06-19",
	})
	if err != nil {
		t.Fatalf("seed duplicate transaction: %v", err)
	}

	_, err = importflow.ReviewAndAppend(ctx, service, importflow.Request{
		SubjectID:      subjectID,
		PortfolioID:    portfolio.ID,
		IdempotencyKey: "import-flow-key-1002",
		RequestPath:    "/internal/imports/review-append",
		Existing:       nil,
		Reader: strings.NewReader(importCSVHeader +
			"DEPOSIT,,,,500.00000000,0.00000000,0.00000000,2026-06-18,,RUB,op-flow-safe,new cash in\n" +
			"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-flow-dup,stale duplicate\n"),
		Decisions: []importer.Decision{
			{RowNumber: 2, Action: importer.DecisionApprove},
			{RowNumber: 3, Action: importer.DecisionApprove},
		},
	})
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected stale duplicate rejection, got %v", err)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected rollback to keep only seeded duplicate transaction, got %d", len(listed))
	}
}
