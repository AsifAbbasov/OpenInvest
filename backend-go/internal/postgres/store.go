package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrIdempotencyConflict  = errors.New("idempotency conflict")
	ErrIdempotencyInFlight  = errors.New("idempotency in flight")
	ErrUnsupportedDuplicate = errors.New("duplicate response is not available")
)

type assetFixture struct {
	ID        string
	Ticker    string
	AssetType string
	Name      string
	ISIN      *string
	LotSize   string
}

var approvedAssetFixtures = map[string]assetFixture{
	"SBER": {
		ID:        "00000000-0000-4000-8000-00000000a001",
		Ticker:    "SBER",
		AssetType: "stock",
		Name:      "Sberbank ordinary shares",
		ISIN:      stringPtr("RU0009029540"),
		LotSize:   "10.00000000",
	},
	"GAZP": {
		ID:        "00000000-0000-4000-8000-00000000a002",
		Ticker:    "GAZP",
		AssetType: "stock",
		Name:      "Gazprom ordinary shares",
		ISIN:      stringPtr("RU0007661625"),
		LotSize:   "10.00000000",
	},
	"SU26238RMFS4": {
		ID:        "00000000-0000-4000-8000-00000000b001",
		Ticker:    "SU26238RMFS4",
		AssetType: "bond",
		Name:      "OFZ 26238",
		ISIN:      stringPtr("RU000A1038V6"),
		LotSize:   "1.00000000",
	},
}

type Store struct {
	db *sql.DB
}

func Open(databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) SearchAssets(ctx context.Context, filter verticalslice.AssetSearchFilter) ([]verticalslice.AssetSummary, error) {
	approvedTickers := make([]string, 0, len(approvedAssetFixtures))
	for ticker := range approvedAssetFixtures {
		approvedTickers = append(approvedTickers, ticker)
	}
	sort.Strings(approvedTickers)

	args := []any{likePrefixPattern(filter.Query), likeContainsPattern(filter.Query)}
	conditions := []string{
		"lifecycle_status = 'active'",
		"currency = 'RUB'",
		"market = 'MOEX'",
		"(ticker ILIKE $1 ESCAPE '\\' OR name ILIKE $2 ESCAPE '\\')",
	}
	if filter.AssetType != "" {
		args = append(args, strings.ToLower(filter.AssetType))
		conditions = append(conditions, "asset_type = $"+strconv.Itoa(len(args)))
	}
	if filter.AfterTicker != "" {
		args = append(args, filter.AfterTicker)
		conditions = append(conditions, "ticker > $"+strconv.Itoa(len(args)))
	}
	canonicalFixtureConditions := make([]string, 0, len(approvedTickers))
	for _, ticker := range approvedTickers {
		fixture := approvedAssetFixtures[ticker]
		args = append(args, fixture.Ticker)
		tickerPlaceholder := "$" + strconv.Itoa(len(args))
		args = append(args, fixture.AssetType)
		assetTypePlaceholder := "$" + strconv.Itoa(len(args))
		args = append(args, fixture.Name)
		namePlaceholder := "$" + strconv.Itoa(len(args))
		args = append(args, fixture.ISIN)
		isinPlaceholder := "$" + strconv.Itoa(len(args))
		args = append(args, fixture.LotSize)
		lotSizePlaceholder := "$" + strconv.Itoa(len(args))
		canonicalFixtureConditions = append(canonicalFixtureConditions, "("+strings.Join([]string{
			"ticker = " + tickerPlaceholder,
			"asset_type = " + assetTypePlaceholder,
			"name = " + namePlaceholder,
			"((" + isinPlaceholder + "::text IS NULL AND isin IS NULL) OR isin = " + isinPlaceholder + "::text)",
			"lot_size = " + lotSizePlaceholder + "::numeric",
		}, " AND ")+")")
	}
	conditions = append(conditions, "("+strings.Join(canonicalFixtureConditions, " OR ")+")")
	args = append(args, filter.Limit)
	limitPlaceholder := "$" + strconv.Itoa(len(args))

	rows, err := s.db.QueryContext(ctx, `
		SELECT ticker, name, asset_type, currency, isin, lot_size::text
		FROM investment.assets
		WHERE `+strings.Join(conditions, " AND ")+`
		ORDER BY ticker ASC
		LIMIT `+limitPlaceholder+`
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := []verticalslice.AssetSummary{}
	for rows.Next() {
		asset, ok, err := scanApprovedAssetSummary(rows)
		if err != nil {
			return nil, err
		}
		if ok {
			assets = append(assets, asset)
		}
	}
	return assets, rows.Err()
}

func (s *Store) ListPortfolios(ctx context.Context, subjectID string, filter verticalslice.PortfolioFilter) ([]verticalslice.Portfolio, error) {
	args := []any{subjectID}
	conditions := []string{"subject_id = $1", "portfolio_state = 'active'"}
	if filter.BeforeUpdatedAt != nil {
		args = append(args, *filter.BeforeUpdatedAt, filter.BeforeID)
		conditions = append(conditions, "(updated_at, id) < ($"+strconv.Itoa(len(args)-1)+"::timestamptz, $"+strconv.Itoa(len(args))+"::uuid)")
	}
	args = append(args, filter.Limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, base_currency, version, created_at, updated_at
		FROM investment.portfolios
		WHERE `+strings.Join(conditions, " AND ")+`
		ORDER BY updated_at DESC, id DESC
		LIMIT $`+strconv.Itoa(len(args))+`
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []verticalslice.Portfolio
	for rows.Next() {
		portfolio, err := scanPortfolio(rows)
		if err != nil {
			return nil, err
		}
		portfolios = append(portfolios, portfolio)
	}
	return portfolios, rows.Err()
}

func (s *Store) CreatePortfolio(ctx context.Context, command verticalslice.CommandContext, request verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Portfolio{}, err
	}
	defer rollback(tx)

	if err := ensureSubject(ctx, tx, command.SubjectID); err != nil {
		return verticalslice.Portfolio{}, err
	}

	portfolioID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.Portfolio{}, err
	}
	if duplicate {
		portfolio, err := getPortfolioTx(ctx, tx, command.SubjectID, portfolioID)
		if err != nil {
			return verticalslice.Portfolio{}, err
		}
		return portfolio, tx.Commit()
	}

	now := command.Now
	_, err = tx.ExecContext(ctx, `
		INSERT INTO investment.portfolios
			(id, subject_id, name, base_currency, portfolio_state, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', 1, $5, $5)
	`, portfolioID, command.SubjectID, request.Name, request.BaseCurrency, now)
	if err != nil {
		return verticalslice.Portfolio{}, err
	}

	portfolio, err := getPortfolioTx(ctx, tx, command.SubjectID, portfolioID)
	if err != nil {
		return verticalslice.Portfolio{}, err
	}
	if err := completeCommand(ctx, tx, portfolioID); err != nil {
		return verticalslice.Portfolio{}, err
	}
	return portfolio, tx.Commit()
}

func (s *Store) GetPortfolio(ctx context.Context, subjectID string, portfolioID string) (verticalslice.Portfolio, error) {
	return getPortfolioQuery(ctx, s.db, subjectID, portfolioID)
}

func (s *Store) ListTransactions(ctx context.Context, subjectID string, portfolioID string, filter verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	if _, err := s.GetPortfolio(ctx, subjectID, portfolioID); err != nil {
		return nil, err
	}
	args := []any{portfolioID}
	conditions := []string{"te.portfolio_id = $1"}
	if filter.TransactionType != "" {
		args = append(args, filter.TransactionType)
		conditions = append(conditions, "te.transaction_type = $"+strconv.Itoa(len(args)))
	}
	if filter.FromDate != "" {
		args = append(args, filter.FromDate)
		conditions = append(conditions, "te.trade_date >= $"+strconv.Itoa(len(args))+"::date")
	}
	if filter.ToDate != "" {
		args = append(args, filter.ToDate)
		conditions = append(conditions, "te.trade_date <= $"+strconv.Itoa(len(args))+"::date")
	}
	if filter.BeforeTradeDate != "" {
		args = append(args, filter.BeforeTradeDate, filter.BeforeEntryID)
		conditions = append(conditions, "(te.trade_date, te.entry_id) < ($"+strconv.Itoa(len(args)-1)+"::date, $"+strconv.Itoa(len(args))+"::uuid)")
	}
	args = append(args, filter.Limit)
	rows, err := s.db.QueryContext(ctx, transactionSelectSQL()+`
			WHERE `+strings.Join(conditions, " AND ")+`
			ORDER BY te.trade_date DESC, te.entry_id DESC
			LIMIT $`+strconv.Itoa(len(args))+`
		`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []verticalslice.Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, rows.Err()
}

func (s *Store) AppendTransaction(ctx context.Context, command verticalslice.CommandContext, request verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	defer rollback(tx)

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return verticalslice.Transaction{}, err
	}

	entryID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if duplicate {
		transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
		if err != nil {
			return verticalslice.Transaction{}, err
		}
		return transaction, tx.Commit()
	}

	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	equivalentDuplicate, err := equivalentTransactionExists(ctx, tx, request, true)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if equivalentDuplicate {
		return verticalslice.Transaction{}, verticalslice.ErrInvalidInput
	}
	assetID, err := ensureAsset(ctx, tx, request)
	if err != nil {
		return verticalslice.Transaction{}, err
	}

	transactionID := uuid.NewString()
	var ticker *string
	if request.Ticker != nil {
		value := *request.Ticker
		ticker = &value
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
			quantity, unit_price_amount, unit_price_currency,
			gross_amount, gross_currency, commission_amount, commission_currency,
			tax_amount, tax_currency, trade_date, settlement_date, note,
			source_kind, source_file_hash, created_at, request_id, trace_id
		)
		VALUES (
			$1, $2, $3, $4, 1, $5,
			$6, $7, $8,
			$9, 'RUB', $10, 'RUB',
			$11, 'RUB', $12, $13, $14,
			'MANUAL', NULL, $15, NULLIF($16, '')::uuid, NULLIF($17, '')
		)
	`, entryID, transactionID, request.PortfolioID, assetID, request.TransactionType,
		decimalString(request.Quantity), moneyAmount(request.UnitPrice), moneyCurrency(request.UnitPrice),
		gross.Amount.String(), request.Commission.Amount.String(),
		request.Tax.Amount.String(), request.TradeDate, request.SettlementDate, request.Note,
		command.Now, command.RequestID, command.TraceID)
	if err != nil {
		return verticalslice.Transaction{}, err
	}

	if err := rebuildAffectedSnapshots(ctx, tx, request.PortfolioID, request.TradeDate, command.Now); err != nil {
		return verticalslice.Transaction{}, err
	}
	if err := completeCommand(ctx, tx, entryID); err != nil {
		return verticalslice.Transaction{}, err
	}

	transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	_ = ticker
	return transaction, tx.Commit()
}

func (s *Store) AppendImportedTransactions(ctx context.Context, command verticalslice.CommandContext, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return nil, err
	}

	commandID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return nil, err
	}
	if duplicate {
		transactions, err := getImportedTransactionsByCommandTx(ctx, tx, commandID, request)
		if err != nil {
			return nil, err
		}
		return transactions, tx.Commit()
	}

	for _, transactionRequest := range request.Transactions {
		duplicate, err := importIdentityExists(ctx, tx, transactionRequest, request.SourceAccountLabel)
		if err != nil {
			return nil, err
		}
		if duplicate {
			return nil, verticalslice.ErrInvalidInput
		}
		legacyOrManualDuplicate, err := legacyOrManualEquivalentTransactionExists(ctx, tx, transactionRequest)
		if err != nil {
			return nil, err
		}
		if legacyOrManualDuplicate {
			return nil, verticalslice.ErrInvalidInput
		}
		conflict, err := nearConflictTransactionExists(ctx, tx, transactionRequest)
		if err != nil {
			return nil, err
		}
		if conflict {
			return nil, verticalslice.ErrInvalidInput
		}
	}

	if err := ensureAssets(ctx, tx, request.Transactions); err != nil {
		return nil, err
	}

	entryIDs := make([]string, 0, len(request.Transactions))
	snapshotDateSet := map[string]struct{}{}
	for index, transactionRequest := range request.Transactions {
		entryID, err := importEntryID(commandID, index)
		if err != nil {
			return nil, err
		}
		if err := insertTransactionEntryWithID(ctx, tx, command, transactionRequest, entryID, request.SourceKind, request.SourceFileHash, request.SourceAccountLabel); err != nil {
			return nil, err
		}
		entryIDs = append(entryIDs, entryID)
		snapshotDateSet[transactionRequest.TradeDate] = struct{}{}
	}

	snapshotDates := make([]string, 0, len(snapshotDateSet))
	for snapshotDate := range snapshotDateSet {
		snapshotDates = append(snapshotDates, snapshotDate)
	}
	sort.Strings(snapshotDates)
	for _, snapshotDate := range snapshotDates {
		if err := rebuildAffectedSnapshots(ctx, tx, request.PortfolioID, snapshotDate, command.Now); err != nil {
			return nil, err
		}
	}
	if err := recordImportAppendAudit(ctx, tx, command, request.PortfolioID); err != nil {
		return nil, err
	}
	if err := completeCommand(ctx, tx, commandID); err != nil {
		return nil, err
	}

	transactions := make([]verticalslice.Transaction, 0, len(entryIDs))
	for _, entryID := range entryIDs {
		transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, tx.Commit()
}

func (s *Store) GetPortfolioSummary(ctx context.Context, subjectID string, portfolioID string, asOfDate string) (verticalslice.PortfolioSummary, error) {
	if _, err := s.GetPortfolio(ctx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioSummary{}, err
	}
	return getPortfolioSummary(ctx, s.db, portfolioID, asOfDate)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanApprovedAssetSummary(row scanner) (verticalslice.AssetSummary, bool, error) {
	var ticker string
	var name string
	var assetType string
	var currency string
	var isin sql.NullString
	var lotSizeText string
	if err := row.Scan(&ticker, &name, &assetType, &currency, &isin, &lotSizeText); err != nil {
		return verticalslice.AssetSummary{}, false, err
	}
	fixture, ok := approvedAssetFixture(ticker)
	if !ok || fixture.Name != name || fixture.AssetType != assetType || currency != verticalslice.RUB {
		return verticalslice.AssetSummary{}, false, nil
	}
	if (fixture.ISIN == nil && isin.Valid) || (fixture.ISIN != nil && (!isin.Valid || isin.String != *fixture.ISIN)) {
		return verticalslice.AssetSummary{}, false, nil
	}
	lotSize, err := decimal.FromString(lotSizeText)
	if err != nil {
		return verticalslice.AssetSummary{}, false, err
	}
	fixtureLotSize := decimal.Must(fixture.LotSize)
	if !lotSize.Equal(fixtureLotSize) {
		return verticalslice.AssetSummary{}, false, nil
	}
	return verticalslice.AssetSummary{
		Ticker:    ticker,
		Name:      name,
		AssetType: strings.ToUpper(assetType),
		Currency:  currency,
		LotSize:   fixtureLotSize,
		LastPrice: nil,
	}, true, nil
}

func likeContainsPattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(value) + "%"
}

func likePrefixPattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value) + "%"
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func ensureSubject(ctx context.Context, tx *sql.Tx, subjectID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO investment.subjects (id, subject_state)
		VALUES ($1, 'active')
		ON CONFLICT (id) DO NOTHING
	`, subjectID)
	return err
}

func equivalentTransactionExists(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest, importedOnly bool) (bool, error) {
	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return false, err
	}
	var ticker any
	if request.Ticker != nil {
		ticker = *request.Ticker
	}
	var quantity any
	if request.Quantity != nil {
		quantity = request.Quantity.String()
	}
	var unitPrice any
	if request.UnitPrice != nil {
		unitPrice = request.UnitPrice.Amount.String()
	}
	var settlementDate any
	if request.SettlementDate != nil {
		settlementDate = *request.SettlementDate
	}
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM investment.transaction_entries te
			LEFT JOIN investment.assets a ON a.id = te.asset_id
			WHERE te.portfolio_id = $1
				AND te.transaction_type = $2
				AND te.gross_amount = $3::numeric
				AND te.commission_amount = $4::numeric
				AND te.tax_amount = $5::numeric
				AND te.trade_date = $6::date
				AND (($7::date IS NULL AND te.settlement_date IS NULL) OR te.settlement_date = $7::date)
				AND (($8::text IS NULL AND a.ticker IS NULL) OR a.ticker = $8::text)
				AND (($9::numeric IS NULL AND te.quantity IS NULL) OR te.quantity = $9::numeric)
				AND (($10::numeric IS NULL AND te.unit_price_amount IS NULL) OR te.unit_price_amount = $10::numeric)
				AND (NOT $11::boolean OR te.source_kind = 'USER_UPLOADED_FILE')
			)
	`, request.PortfolioID, request.TransactionType, gross.Amount.String(),
		request.Commission.Amount.String(), request.Tax.Amount.String(), request.TradeDate,
		settlementDate, ticker, quantity, unitPrice, importedOnly).Scan(&exists)
	return exists, err
}

func importIdentityExists(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest, sourceAccountLabel string) (bool, error) {
	provenance := request.ImportProvenance
	if provenance == nil || provenance.IdentityVersion != verticalslice.ImportIdentityVersion {
		return false, verticalslice.ErrInvalidInput
	}

	var exists bool
	if provenance.BrokerOperationKey != "" {
		err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM investment.transaction_entries te
				WHERE te.portfolio_id = $1
					AND te.source_kind = 'USER_UPLOADED_FILE'
					AND te.source_identity_version = $2
					AND te.source_account_label = $3
					AND (
						te.source_broker_operation_key = $4
						OR (
							te.source_broker_operation_key IS NULL
							AND te.source_fingerprint = $5
						)
					)
			)
		`, request.PortfolioID, provenance.IdentityVersion, strings.TrimSpace(sourceAccountLabel), provenance.BrokerOperationKey, provenance.SourceFingerprint).Scan(&exists)
		return exists, err
	}

	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM investment.transaction_entries te
			WHERE te.portfolio_id = $1
				AND te.source_kind = 'USER_UPLOADED_FILE'
				AND te.source_identity_version = $2
				AND te.source_account_label = $3
				AND te.source_fingerprint = $4
		)
	`, request.PortfolioID, provenance.IdentityVersion, strings.TrimSpace(sourceAccountLabel), provenance.SourceFingerprint).Scan(&exists)
	return exists, err
}

func legacyOrManualEquivalentTransactionExists(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest) (bool, error) {
	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return false, err
	}
	var ticker any
	if request.Ticker != nil {
		ticker = *request.Ticker
	}
	var quantity any
	if request.Quantity != nil {
		quantity = request.Quantity.String()
	}
	var unitPrice any
	if request.UnitPrice != nil {
		unitPrice = request.UnitPrice.Amount.String()
	}
	var settlementDate any
	if request.SettlementDate != nil {
		settlementDate = *request.SettlementDate
	}
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM investment.transaction_entries te
			LEFT JOIN investment.assets a ON a.id = te.asset_id
			WHERE te.portfolio_id = $1
				AND te.transaction_type = $2
				AND te.gross_amount = $3::numeric
				AND te.commission_amount = $4::numeric
				AND te.tax_amount = $5::numeric
				AND te.trade_date = $6::date
				AND (($7::date IS NULL AND te.settlement_date IS NULL) OR te.settlement_date = $7::date)
				AND (($8::text IS NULL AND a.ticker IS NULL) OR a.ticker = $8::text)
				AND (($9::numeric IS NULL AND te.quantity IS NULL) OR te.quantity = $9::numeric)
				AND (($10::numeric IS NULL AND te.unit_price_amount IS NULL) OR te.unit_price_amount = $10::numeric)
				AND (
					te.source_kind = 'MANUAL'
					OR (te.source_kind = 'USER_UPLOADED_FILE' AND te.source_identity_version IS NULL)
				)
			)
	`, request.PortfolioID, request.TransactionType, gross.Amount.String(),
		request.Commission.Amount.String(), request.Tax.Amount.String(), request.TradeDate,
		settlementDate, ticker, quantity, unitPrice).Scan(&exists)
	return exists, err
}

func nearConflictTransactionExists(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest) (bool, error) {
	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return false, err
	}
	var ticker any
	if request.Ticker != nil {
		ticker = *request.Ticker
	}
	var quantity any
	if request.Quantity != nil {
		quantity = request.Quantity.String()
	}
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM investment.transaction_entries te
			LEFT JOIN investment.assets a ON a.id = te.asset_id
			WHERE te.portfolio_id = $1
				AND te.transaction_type = $2
				AND te.trade_date = $3::date
				AND (($4::text IS NULL AND a.ticker IS NULL) OR a.ticker = $4::text)
				AND (($5::numeric IS NULL AND te.quantity IS NULL) OR te.quantity = $5::numeric)
				AND (te.transaction_type NOT IN ('DEPOSIT', 'WITHDRAWAL') OR te.gross_amount = $6::numeric)
				AND (
					te.gross_amount <> $6::numeric
					OR te.commission_amount <> $7::numeric
					OR te.tax_amount <> $8::numeric
				)
		)
	`, request.PortfolioID, request.TransactionType, request.TradeDate, ticker, quantity,
		gross.Amount.String(), request.Commission.Amount.String(), request.Tax.Amount.String()).Scan(&exists)
	return exists, err
}

func getImportedTransactionsByCommandTx(ctx context.Context, tx *sql.Tx, commandID string, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	transactions := make([]verticalslice.Transaction, 0, len(request.Transactions))
	for index := range request.Transactions {
		entryID, err := importEntryID(commandID, index)
		if err != nil {
			return nil, err
		}
		transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func importEntryID(commandID string, index int) (string, error) {
	namespace, err := uuid.Parse(commandID)
	if err != nil {
		return "", err
	}
	return uuid.NewSHA1(namespace, []byte("import-entry-"+strconv.Itoa(index+1))).String(), nil
}

func insertTransactionEntry(ctx context.Context, tx *sql.Tx, command verticalslice.CommandContext, request verticalslice.AppendTransactionRequest) (string, error) {
	entryID := uuid.NewString()
	if _, err := ensureAsset(ctx, tx, request); err != nil {
		return "", err
	}
	if err := insertTransactionEntryWithID(ctx, tx, command, request, entryID, "MANUAL", "", ""); err != nil {
		return "", err
	}
	return entryID, nil
}

func insertTransactionEntryWithID(ctx context.Context, tx *sql.Tx, command verticalslice.CommandContext, request verticalslice.AppendTransactionRequest, entryID string, sourceKind string, sourceFileHash string, sourceAccountLabel string) error {
	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return err
	}
	assetID, err := activeAssetID(ctx, tx, request)
	if err != nil {
		return err
	}

	transactionID := uuid.NewString()
	brokerOperationKey := ""
	sourceFingerprint := ""
	var sourceIdentityVersion any
	if request.ImportProvenance != nil {
		brokerOperationKey = request.ImportProvenance.BrokerOperationKey
		sourceFingerprint = request.ImportProvenance.SourceFingerprint
		sourceIdentityVersion = request.ImportProvenance.IdentityVersion
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
			quantity, unit_price_amount, unit_price_currency,
			gross_amount, gross_currency, commission_amount, commission_currency,
			tax_amount, tax_currency, trade_date, settlement_date, note,
			source_kind, source_file_hash, source_account_label,
			source_broker_operation_key, source_fingerprint, source_identity_version,
			created_at, request_id, trace_id
		)
		VALUES (
			$1, $2, $3, $4, 1, $5,
			$6, $7, $8,
			$9, 'RUB', $10, 'RUB',
			$11, 'RUB', $12, $13, $14,
			$15, NULLIF($16, ''), $17,
			NULLIF($18, ''), NULLIF($19, ''), $20,
			$21, NULLIF($22, '')::uuid, NULLIF($23, '')
		)
	`, entryID, transactionID, request.PortfolioID, assetID, request.TransactionType,
		decimalString(request.Quantity), moneyAmount(request.UnitPrice), moneyCurrency(request.UnitPrice),
		gross.Amount.String(), request.Commission.Amount.String(),
		request.Tax.Amount.String(), request.TradeDate, request.SettlementDate, request.Note,
		sourceKind, sourceFileHash, strings.TrimSpace(sourceAccountLabel),
		brokerOperationKey, sourceFingerprint, sourceIdentityVersion,
		command.Now, command.RequestID, command.TraceID)
	if err != nil {
		return err
	}
	return nil
}

func recordImportAppendAudit(ctx context.Context, tx *sql.Tx, command verticalslice.CommandContext, portfolioID string) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit.actors (id, actor_kind, created_at)
		VALUES ($1, 'user', $2)
		ON CONFLICT (id) DO NOTHING
	`, command.SubjectID, command.Now); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO audit.events (
			id, actor_id, action_code, target_kind, target_id, outcome,
			request_id, trace_id, occurred_at, schema_version
		)
		VALUES (
			$1, $2, 'IMPORT_APPEND_BATCH', 'portfolio', $3, 'success',
			NULLIF($4, '')::uuid, NULLIF($5, ''), $6, 1
		)
	`, uuid.NewString(), command.SubjectID, portfolioID, command.RequestID, command.TraceID, command.Now)
	return err
}

func reserveCommand(ctx context.Context, tx *sql.Tx, command verticalslice.CommandContext, method string) (string, bool, error) {
	commandID := uuid.NewString()
	var existingID string
	var existingHash string
	var terminalStatus sql.NullString
	err := tx.QueryRowContext(ctx, `
		INSERT INTO investment.command_deduplication
			(id, principal_id, method, canonical_path, idempotency_key, request_hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::timestamptz, $7::timestamptz + interval '24 hours')
		ON CONFLICT (principal_id, method, canonical_path, idempotency_key)
		DO UPDATE SET idempotency_key = investment.command_deduplication.idempotency_key
		RETURNING id, request_hash, terminal_status
	`, commandID, command.SubjectID, method, command.RequestPath, command.IdempotencyKey, command.RequestHash, command.Now).
		Scan(&existingID, &existingHash, &terminalStatus)
	if err != nil {
		return "", false, err
	}
	if existingID == commandID {
		return commandID, false, nil
	}
	if existingHash != command.RequestHash {
		return "", false, ErrIdempotencyConflict
	}
	if !terminalStatus.Valid {
		return "", false, ErrIdempotencyInFlight
	}
	return existingID, true, nil
}

func completeCommand(ctx context.Context, tx *sql.Tx, commandID string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE investment.command_deduplication
		SET terminal_status = 'success', response_hash = request_hash
		WHERE id = $1
	`, commandID)
	return err
}

func getPortfolioQuery(ctx context.Context, db queryer, subjectID string, portfolioID string) (verticalslice.Portfolio, error) {
	return scanPortfolio(db.QueryRowContext(ctx, `
		SELECT id, name, base_currency, version, created_at, updated_at
		FROM investment.portfolios
		WHERE id = $1 AND subject_id = $2 AND portfolio_state = 'active'
	`, portfolioID, subjectID))
}

func getPortfolioTx(ctx context.Context, tx *sql.Tx, subjectID string, portfolioID string) (verticalslice.Portfolio, error) {
	return getPortfolioQuery(ctx, tx, subjectID, portfolioID)
}

func lockPortfolioTx(ctx context.Context, tx *sql.Tx, subjectID string, portfolioID string) error {
	var id string
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM investment.portfolios
		WHERE id = $1 AND subject_id = $2 AND portfolio_state = 'active'
		FOR UPDATE
	`, portfolioID, subjectID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func scanPortfolio(row scanner) (verticalslice.Portfolio, error) {
	var portfolio verticalslice.Portfolio
	err := row.Scan(&portfolio.ID, &portfolio.Name, &portfolio.BaseCurrency, &portfolio.Version, &portfolio.CreatedAt, &portfolio.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return verticalslice.Portfolio{}, ErrNotFound
	}
	return portfolio, err
}

func transactionSelectSQL() string {
	return `
		SELECT
			te.entry_id, te.transaction_id, te.portfolio_id, te.transaction_type,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM investment.transaction_entries later
					WHERE later.transaction_id = te.transaction_id AND later.revision > te.revision
				) THEN 'CORRECTED'
				WHEN EXISTS (
					SELECT 1 FROM investment.transaction_entries reversal
					WHERE reversal.reverses_transaction_id = te.transaction_id
				) THEN 'REVERSED'
				ELSE 'ACTIVE'
			END AS status,
			a.ticker,
			te.quantity::text,
			te.unit_price_amount::text,
			te.unit_price_currency,
			te.gross_amount::text,
			te.commission_amount::text,
			te.tax_amount::text,
			te.trade_date::text,
			te.settlement_date::text,
			te.note,
			te.source_kind,
			te.source_account_label,
			COALESCE(te.source_broker_operation_key, ''),
			COALESCE(te.source_fingerprint, ''),
			COALESCE(te.source_identity_version, 0),
			te.revision,
			te.created_at,
			te.created_at AS updated_at
		FROM investment.transaction_entries te
		LEFT JOIN investment.assets a ON a.id = te.asset_id
	`
}

func getTransactionByEntryTx(ctx context.Context, tx *sql.Tx, portfolioID string, entryID string) (verticalslice.Transaction, error) {
	return scanTransaction(tx.QueryRowContext(ctx, transactionSelectSQL()+`
		WHERE te.portfolio_id = $1 AND te.entry_id = $2
	`, portfolioID, entryID))
}

func scanTransaction(row scanner) (verticalslice.Transaction, error) {
	var transaction verticalslice.Transaction
	var ticker sql.NullString
	var quantity sql.NullString
	var unitPriceAmount sql.NullString
	var unitPriceCurrency sql.NullString
	var grossAmount string
	var commissionAmount string
	var taxAmount string
	var settlementDate sql.NullString
	var note sql.NullString
	err := row.Scan(
		&transaction.EntryID, &transaction.ID, &transaction.PortfolioID, &transaction.TransactionType, &transaction.Status,
		&ticker, &quantity, &unitPriceAmount, &unitPriceCurrency,
		&grossAmount, &commissionAmount, &taxAmount,
		&transaction.TradeDate, &settlementDate, &note,
		&transaction.SourceKind, &transaction.SourceAccountLabel,
		&transaction.SourceBrokerOperationKey, &transaction.SourceFingerprint, &transaction.SourceIdentityVersion,
		&transaction.Revision, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return verticalslice.Transaction{}, ErrNotFound
	}
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if ticker.Valid {
		transaction.Ticker = &ticker.String
	}
	if quantity.Valid {
		parsed := decimal.Must(quantity.String)
		transaction.Quantity = &parsed
	}
	if unitPriceAmount.Valid {
		parsed := verticalslice.Money{Amount: decimal.Must(unitPriceAmount.String), Currency: unitPriceCurrency.String}
		transaction.UnitPrice = &parsed
	}
	transaction.GrossAmount = verticalslice.Money{Amount: decimal.Must(grossAmount), Currency: verticalslice.RUB}
	transaction.Commission = verticalslice.Money{Amount: decimal.Must(commissionAmount), Currency: verticalslice.RUB}
	transaction.Tax = verticalslice.Money{Amount: decimal.Must(taxAmount), Currency: verticalslice.RUB}
	if settlementDate.Valid {
		transaction.SettlementDate = &settlementDate.String
	}
	if note.Valid {
		transaction.Note = &note.String
	}
	return transaction, nil
}

func ensureAsset(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest) (*string, error) {
	if request.Ticker == nil {
		return nil, nil
	}
	fixture, ok := approvedAssetFixture(*request.Ticker)
	if !ok {
		return nil, verticalslice.ErrInvalidInput
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, $2, $3, $4, 'RUB', 'MOEX', 'active', $5, $6::numeric, now())
		ON CONFLICT (ticker) DO NOTHING
	`, fixture.ID, fixture.Ticker, fixture.AssetType, fixture.Name, fixture.ISIN, fixture.LotSize)
	if err != nil {
		return nil, err
	}
	return activeAssetID(ctx, tx, request)
}

func ensureAssets(ctx context.Context, tx *sql.Tx, requests []verticalslice.AppendTransactionRequest) error {
	tickerSet := map[string]verticalslice.AppendTransactionRequest{}
	for _, request := range requests {
		if request.Ticker != nil {
			tickerSet[*request.Ticker] = request
		}
	}
	tickers := make([]string, 0, len(tickerSet))
	for ticker := range tickerSet {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)
	for _, ticker := range tickers {
		if _, err := ensureAsset(ctx, tx, tickerSet[ticker]); err != nil {
			return err
		}
	}
	return nil
}

func activeAssetID(ctx context.Context, tx *sql.Tx, request verticalslice.AppendTransactionRequest) (*string, error) {
	if request.Ticker == nil {
		return nil, nil
	}
	fixture, ok := approvedAssetFixture(*request.Ticker)
	if !ok {
		return nil, verticalslice.ErrInvalidInput
	}
	var existingID string
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM investment.assets
		WHERE ticker = $1
			AND asset_type = $2
			AND name = $3
			AND currency = 'RUB'
			AND market = 'MOEX'
			AND lifecycle_status = 'active'
			AND (($4::text IS NULL AND isin IS NULL) OR isin = $4::text)
			AND lot_size = $5::numeric
	`, fixture.Ticker, fixture.AssetType, fixture.Name, fixture.ISIN, fixture.LotSize).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, verticalslice.ErrInvalidInput
		}
		return nil, err
	}
	return &existingID, nil
}

func approvedAssetFixture(ticker string) (assetFixture, bool) {
	fixture, ok := approvedAssetFixtures[ticker]
	return fixture, ok
}

func stringPtr(value string) *string {
	return &value
}

func rebuildAffectedSnapshots(ctx context.Context, tx *sql.Tx, portfolioID string, transactionTradeDate string, now time.Time) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT snapshot_date::text
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_date >= $2::date
			AND methodology_version = 'stage-03-02-local-cost-snapshot-v1'
		UNION
		SELECT $2::date::text
		ORDER BY 1
	`, portfolioID, transactionTradeDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	var snapshotDates []string
	for rows.Next() {
		var snapshotDate string
		if err := rows.Scan(&snapshotDate); err != nil {
			return err
		}
		snapshotDates = append(snapshotDates, snapshotDate)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, snapshotDate := range snapshotDates {
		if err := rebuildSnapshot(ctx, tx, portfolioID, snapshotDate, now); err != nil {
			return err
		}
	}
	return nil
}

func rebuildSnapshot(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string, now time.Time) error {
	snapshotID := uuid.NewString()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.portfolio_snapshots (
			id, portfolio_id, snapshot_date,
			total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			snapshot_version, methodology_version, input_watermark, calculated_at
		)
		WITH ledger AS (
			SELECT
				COALESCE(SUM(CASE
					WHEN te.transaction_type = 'DEPOSIT' THEN te.gross_amount
					WHEN te.transaction_type = 'WITHDRAWAL' THEN -te.gross_amount
					WHEN te.transaction_type = 'BUY' THEN -(te.gross_amount + te.commission_amount + te.tax_amount)
					WHEN te.transaction_type = 'SELL' THEN te.gross_amount - te.commission_amount - te.tax_amount
					ELSE 0
				END), 0) AS cash_value,
				COALESCE(SUM(CASE
					WHEN te.transaction_type = 'BUY' AND a.asset_type = 'stock' THEN te.gross_amount
					WHEN te.transaction_type = 'SELL' AND a.asset_type = 'stock' THEN -te.gross_amount
					ELSE 0
				END), 0) AS stock_value,
				COALESCE(SUM(CASE
					WHEN te.transaction_type = 'BUY' AND a.asset_type = 'bond' THEN te.gross_amount
					WHEN te.transaction_type = 'SELL' AND a.asset_type = 'bond' THEN -te.gross_amount
					ELSE 0
				END), 0) AS bond_value,
				COALESCE(SUM(CASE WHEN te.transaction_type = 'BUY' THEN te.gross_amount + te.commission_amount + te.tax_amount ELSE 0 END), 0) AS invested_capital,
					COALESCE(MAX(te.created_at)::text, 'empty') AS watermark
				FROM investment.transaction_entries te
				LEFT JOIN investment.assets a ON a.id = te.asset_id
				WHERE te.portfolio_id = $2
					AND te.trade_date <= $3::date
			), computed AS (
			SELECT
				cash_value,
				stock_value,
				bond_value,
				invested_capital,
				cash_value + stock_value + bond_value AS total_value,
				CASE WHEN invested_capital > 0
					THEN ((cash_value + stock_value + bond_value) - invested_capital) / invested_capital
					ELSE 0
				END AS nominal_return_rate,
				watermark
			FROM ledger
		)
		SELECT
			$1, $2, $3::date,
			total_value, cash_value, stock_value, bond_value,
			invested_capital, nominal_return_rate, nominal_return_rate,
				COALESCE((
					SELECT MAX(snapshot_version) + 1
					FROM analytics.portfolio_snapshots
					WHERE portfolio_id = $2
						AND snapshot_date = $3::date
						AND methodology_version = 'stage-03-02-local-cost-snapshot-v1'
				), 1),
				'stage-03-02-local-cost-snapshot-v1', watermark, $4
		FROM computed
		WHERE
			round(total_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(cash_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(stock_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(bond_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(invested_capital, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(nominal_return_rate, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
	`, snapshotID, portfolioID, snapshotDate, now)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8) storage precision", verticalslice.ErrInvalidInput)
	}
	return nil
}

func getPortfolioSummary(ctx context.Context, db *sql.DB, portfolioID string, asOfDate string) (verticalslice.PortfolioSummary, error) {
	var summary verticalslice.PortfolioSummary
	var totalValue string
	var cashValue string
	var stockValue string
	var bondValue string
	var investedCapital string
	var nominalReturn string
	var calculatedAt time.Time
	args := []any{portfolioID}
	dateFilter := ""
	if asOfDate != "" {
		dateFilter = "AND snapshot_date <= $2::date"
		args = append(args, asOfDate)
	}
	err := db.QueryRowContext(ctx, `
		SELECT portfolio_id, snapshot_date::text, total_value_amount::text, cash_value_amount::text,
			stock_value_amount::text, bond_value_amount::text, invested_capital_amount::text,
			nominal_return_rate::text, methodology_version, calculated_at
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1 AND snapshot_status = 'calculated'
		`+dateFilter+`
		ORDER BY snapshot_date DESC, calculated_at DESC
		LIMIT 1
	`, args...).Scan(
		&summary.PortfolioID, &summary.AsOfDate, &totalValue, &cashValue,
		&stockValue, &bondValue, &investedCapital, &nominalReturn,
		&summary.MethodologyVersion, &calculatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return verticalslice.PortfolioSummary{}, ErrNotFound
	}
	if err != nil {
		return verticalslice.PortfolioSummary{}, err
	}
	summary.TotalValue = verticalslice.Money{Amount: decimal.Must(totalValue), Currency: verticalslice.RUB}
	summary.CashValue = verticalslice.Money{Amount: decimal.Must(cashValue), Currency: verticalslice.RUB}
	summary.StockValue = verticalslice.Money{Amount: decimal.Must(stockValue), Currency: verticalslice.RUB}
	summary.BondValue = verticalslice.Money{Amount: decimal.Must(bondValue), Currency: verticalslice.RUB}
	summary.InvestedCapital = verticalslice.Money{Amount: decimal.Must(investedCapital), Currency: verticalslice.RUB}
	summary.DividendsReceived = verticalslice.ZeroMoney()
	summary.CouponsReceived = verticalslice.ZeroMoney()
	summary.NominalReturnRate = decimal.Must(nominalReturn)
	summary.RealReturn = verticalslice.RealReturn{
		NominalReturnRate: summary.NominalReturnRate,
		InflationRate:     decimal.Zero(),
		RealReturnRate:    summary.NominalReturnRate,
		NominalGain:       summary.TotalValue.Sub(summary.InvestedCapital),
		RealGain:          summary.TotalValue.Sub(summary.InvestedCapital),
		FromDate:          summary.AsOfDate,
		ToDate:            summary.AsOfDate,
		Methodology:       "stage-03-02-no-inflation-placeholder-v1",
	}
	summary.PurchasingPower = verticalslice.PurchasingPower{
		PortfolioValue: summary.TotalValue,
		AsOfDate:       summary.AsOfDate,
		Equivalents:    []verticalslice.PurchasingPowerEquivalent{},
	}
	summary.Positions = []verticalslice.PortfolioPosition{}
	summary.CalculatedAt = calculatedAt
	return summary, nil
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func decimalString(value *decimal.Decimal) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func moneyAmount(value *verticalslice.Money) any {
	if value == nil {
		return nil
	}
	return value.Amount.String()
}

func moneyCurrency(value *verticalslice.Money) any {
	if value == nil {
		return nil
	}
	return value.Currency
}
