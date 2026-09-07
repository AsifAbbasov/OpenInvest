from pathlib import Path
import re
from textwrap import dedent


def write(path: str, content: str) -> None:
    p = Path(path)
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(dedent(content).lstrip(), encoding="utf-8")


def replace_once(path: str, old: str, new: str) -> None:
    p = Path(path)
    text = p.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"{path}: expected exactly one match, found {count}: {old[:80]!r}")
    p.write_text(text.replace(old, new, 1), encoding="utf-8")


def regex_once(path: str, pattern: str, replacement: str) -> None:
    p = Path(path)
    text = p.read_text(encoding="utf-8")
    updated, count = re.subn(pattern, replacement, text, count=1, flags=re.S)
    if count != 1:
        raise SystemExit(f"{path}: regex expected exactly one match, found {count}: {pattern[:80]!r}")
    p.write_text(updated, encoding="utf-8")


write("backend-go/internal/verticalslice/stage_03_74_transactions.go", r'''
package verticalslice

import (
    "context"
    "fmt"
    "strings"
    "time"
)

var ErrTransactionConflict = fmt.Errorf("transaction revision conflict")

type CorrectTransactionRequest struct {
    PortfolioID     string
    TransactionID   string
    ExpectedRevision int
    Reason          string
    Corrected       AppendTransactionRequest
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
''')

write("backend-go/internal/verticalslice/stage_03_74_store_adapter.go", r'''
package verticalslice

import "context"

func (adapter *stage371StoreAdapter) CorrectTransactionWithReplayStage374(
    ctx context.Context,
    command CommandContext,
    request CorrectTransactionRequest,
    build TransactionReplayBuilder,
) (Transaction, CommandReplayArtifact, error) {
    store, ok := adapter.Store.(Stage374TransactionRepairReplayStore)
    if !ok {
        return Transaction{}, CommandReplayArtifact{}, ErrReplayUnavailable
    }
    return store.CorrectTransactionWithReplayStage374(ctx, command, request, build)
}

func (adapter *stage371StoreAdapter) ReverseTransactionWithReplayStage374(
    ctx context.Context,
    command CommandContext,
    request ReverseTransactionRequest,
    build TransactionReversalReplayBuilder,
) (TransactionReversal, CommandReplayArtifact, error) {
    store, ok := adapter.Store.(Stage374TransactionRepairReplayStore)
    if !ok {
        return TransactionReversal{}, CommandReplayArtifact{}, ErrReplayUnavailable
    }
    return store.ReverseTransactionWithReplayStage374(ctx, command, request, build)
}
''')

write("backend-go/internal/postgres/effective_ledger.go", r'''
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type effectiveLedgerRow struct {
    EntryID          string
    TransactionID    string
    AssetID          *string
    Ticker           *string
    AssetType        *string
    TransactionType  string
    Quantity         *decimal.Decimal
    UnitPrice        *decimal.Decimal
    GrossAmount      decimal.Decimal
    Commission       decimal.Decimal
    Tax              decimal.Decimal
    TradeDate        string
    LedgerSequence   int64
    Revision         int
    UpdatedAt        time.Time
}

func validateEffectiveLedgerShapeTx(ctx context.Context, tx *sql.Tx, portfolioID string) error {
    if err := assertPortfolioLedgerSequenceCompleteTx(ctx, tx, portfolioID); err != nil {
        return err
    }
    var malformedRevisions int64
    if err := tx.QueryRowContext(ctx, `
        WITH ordered AS (
            SELECT
                transaction_id,
                revision,
                prior_entry_id,
                lag(entry_id) OVER (PARTITION BY transaction_id ORDER BY revision) AS previous_entry_id,
                row_number() OVER (PARTITION BY transaction_id ORDER BY revision) AS expected_revision
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NULL
        )
        SELECT COUNT(*)
        FROM ordered
        WHERE revision <> expected_revision
           OR (revision = 1 AND prior_entry_id IS NOT NULL)
           OR (revision > 1 AND prior_entry_id IS DISTINCT FROM previous_entry_id)
    `, portfolioID).Scan(&malformedRevisions); err != nil {
        return err
    }

    var malformedReversals int64
    if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM investment.transaction_entries reversal
        WHERE reversal.portfolio_id = $1
          AND reversal.reverses_transaction_id IS NOT NULL
          AND (
              reversal.revision <> 1
              OR reversal.prior_entry_id IS NOT NULL
              OR NOT EXISTS (
                  SELECT 1
                  FROM investment.transaction_entries target
                  WHERE target.portfolio_id = reversal.portfolio_id
                    AND target.transaction_id = reversal.reverses_transaction_id
                    AND target.reverses_transaction_id IS NULL
              )
          )
    `, portfolioID).Scan(&malformedReversals); err != nil {
        return err
    }

    var duplicateReversals int64
    if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM (
            SELECT reverses_transaction_id
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NOT NULL
            GROUP BY reverses_transaction_id
            HAVING COUNT(*) > 1
        ) duplicate_reversals
    `, portfolioID).Scan(&duplicateReversals); err != nil {
        return err
    }
    if malformedRevisions != 0 || malformedReversals != 0 || duplicateReversals != 0 {
        return ErrUnsupportedPositionLedger
    }
    return nil
}

func effectiveLedgerRowsTx(ctx context.Context, tx *sql.Tx, portfolioID string, asOfDate string) ([]effectiveLedgerRow, error) {
    if err := validateEffectiveLedgerShapeTx(ctx, tx, portfolioID); err != nil {
        return nil, err
    }
    rows, err := tx.QueryContext(ctx, `
        WITH revisioned AS (
            SELECT
                te.*,
                min(te.ledger_sequence) OVER (PARTITION BY te.transaction_id) AS logical_sequence,
                row_number() OVER (PARTITION BY te.transaction_id ORDER BY te.revision DESC) AS latest_rank
            FROM investment.transaction_entries te
            WHERE te.portfolio_id = $1
              AND te.reverses_transaction_id IS NULL
        ), reversals AS (
            SELECT reverses_transaction_id, min(trade_date) AS effective_date
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NOT NULL
            GROUP BY reverses_transaction_id
        )
        SELECT
            te.entry_id::text,
            te.transaction_id::text,
            te.asset_id::text,
            a.ticker,
            a.asset_type,
            te.transaction_type,
            te.quantity::text,
            te.unit_price_amount::text,
            te.gross_amount::text,
            te.commission_amount::text,
            te.tax_amount::text,
            te.trade_date::text,
            te.logical_sequence,
            te.revision,
            te.created_at
        FROM revisioned te
        LEFT JOIN reversals reversal ON reversal.reverses_transaction_id = te.transaction_id
        LEFT JOIN investment.assets a ON a.id = te.asset_id
        WHERE te.latest_rank = 1
          AND ($2::text = '' OR te.trade_date <= NULLIF($2::text, '')::date)
          AND (
              reversal.reverses_transaction_id IS NULL
              OR ($2::text <> '' AND reversal.effective_date > NULLIF($2::text, '')::date)
          )
        ORDER BY te.trade_date ASC, te.logical_sequence ASC
    `, portfolioID, asOfDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := []effectiveLedgerRow{}
    for rows.Next() {
        var row effectiveLedgerRow
        var assetID, ticker, assetType sql.NullString
        var quantity, unitPrice sql.NullString
        var gross, commission, tax string
        if err := rows.Scan(
            &row.EntryID, &row.TransactionID, &assetID, &ticker, &assetType,
            &row.TransactionType, &quantity, &unitPrice, &gross, &commission, &tax,
            &row.TradeDate, &row.LedgerSequence, &row.Revision, &row.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        if row.LedgerSequence <= 0 {
            return nil, ErrLedgerSequenceUnavailable
        }
        if assetID.Valid {
            value := assetID.String
            row.AssetID = &value
        }
        if ticker.Valid {
            value := ticker.String
            row.Ticker = &value
        }
        if assetType.Valid {
            value := assetType.String
            row.AssetType = &value
        }
        if quantity.Valid {
            parsed, err := decimal.FromString(quantity.String)
            if err != nil { return nil, err }
            row.Quantity = &parsed
        }
        if unitPrice.Valid {
            parsed, err := decimal.FromString(unitPrice.String)
            if err != nil { return nil, err }
            row.UnitPrice = &parsed
        }
        var parseErr error
        row.GrossAmount, parseErr = decimal.FromString(gross)
        if parseErr != nil { return nil, parseErr }
        row.Commission, parseErr = decimal.FromString(commission)
        if parseErr != nil { return nil, parseErr }
        row.Tax, parseErr = decimal.FromString(tax)
        if parseErr != nil { return nil, parseErr }
        result = append(result, row)
    }
    return result, rows.Err()
}

func effectiveSnapshotCashTx(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string) (decimal.Decimal, decimal.Decimal, string, error) {
    rows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, snapshotDate)
    if err != nil {
        return decimal.Zero(), decimal.Zero(), "", err
    }
    cash := decimal.Zero()
    invested := decimal.Zero()
    for _, row := range rows {
        switch row.TransactionType {
        case "DEPOSIT":
            cash = cash.Add(row.GrossAmount)
        case "WITHDRAWAL":
            cash = cash.Sub(row.GrossAmount)
        case "BUY":
            outflow := row.GrossAmount.Add(row.Commission).Add(row.Tax)
            cash = cash.Sub(outflow)
            invested = invested.Add(outflow)
        case "SELL":
            inflow := row.GrossAmount.Sub(row.Commission).Sub(row.Tax)
            cash = cash.Add(inflow)
        }
        if !cash.FitsStorage() || !invested.FitsStorage() {
            return decimal.Zero(), decimal.Zero(), "", fmt.Errorf("snapshot financial values exceed NUMERIC(28,8)")
        }
    }
    var watermark string
    if err := tx.QueryRowContext(ctx, `
        SELECT COALESCE(MAX(created_at)::text, 'empty')
        FROM investment.transaction_entries
        WHERE portfolio_id = $1 AND trade_date <= $2::date
    `, portfolioID, snapshotDate).Scan(&watermark); err != nil {
        return decimal.Zero(), decimal.Zero(), "", err
    }
    return cash, invested, watermark, nil
}
''')

write("backend-go/internal/postgres/current_transaction_projection.go", r'''
package postgres

func currentTransactionSelectSQL() string {
    return `
        WITH revisioned AS (
            SELECT
                source.*,
                row_number() OVER (PARTITION BY source.transaction_id ORDER BY source.revision DESC) AS latest_rank
            FROM investment.transaction_entries source
            WHERE source.reverses_transaction_id IS NULL
        ), te AS (
            SELECT * FROM revisioned WHERE latest_rank = 1
        )
        SELECT
            te.entry_id, te.transaction_id, te.portfolio_id, te.transaction_type,
            CASE
                WHEN EXISTS (
                    SELECT 1 FROM investment.transaction_entries reversal
                    WHERE reversal.portfolio_id = te.portfolio_id
                      AND reversal.reverses_transaction_id = te.transaction_id
                ) THEN 'REVERSED'
                WHEN te.revision > 1 THEN 'CORRECTED'
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
            (SELECT min(first_entry.created_at) FROM investment.transaction_entries first_entry WHERE first_entry.transaction_id = te.transaction_id AND first_entry.reverses_transaction_id IS NULL) AS created_at,
            te.created_at AS updated_at
        FROM te
        LEFT JOIN investment.assets a ON a.id = te.asset_id
    `
}
''')

write("backend-go/internal/postgres/stage_03_74_transactions.go", r'''
package postgres

import (
    "context"
    "database/sql"
    "errors"
    "strings"

    "github.com/google/uuid"

    "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage374LogicalEntry struct {
    EntryID         string
    TransactionID   string
    Revision        int
    TransactionType string
    AssetID         sql.NullString
    Quantity        sql.NullString
    UnitPrice       sql.NullString
    GrossAmount     string
    Commission      string
    Tax             string
    TradeDate       string
    SettlementDate  sql.NullString
    Note            sql.NullString
}

func (s *Store) CorrectTransactionWithReplayStage374(
    ctx context.Context,
    command verticalslice.CommandContext,
    request verticalslice.CorrectTransactionRequest,
    build verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    defer rollback(tx)

    reservation, err := reserveReplayCommand(ctx, tx, command, "PATCH")
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    if reservation.Duplicate {
        if err := tx.Commit(); err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
        return verticalslice.Transaction{}, reservation.Artifact, nil
    }
    if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
    }
    if err := validateEffectiveLedgerShapeTx(ctx, tx, request.PortfolioID); err != nil {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
    }
    current, err := latestLogicalEntryTx(ctx, tx, request.PortfolioID, request.TransactionID)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    reversed, err := logicalTransactionReversedTx(ctx, tx, request.PortfolioID, request.TransactionID)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    if reversed || current.Revision != request.ExpectedRevision {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, verticalslice.ErrTransactionConflict
    }
    gross, err := verticalslice.GrossFor(request.Corrected)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    assetID, err := ensureAsset(ctx, tx, request.Corrected)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    ledgerSequence, err := nextLedgerSequenceTx(ctx, tx, request.PortfolioID)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }

    _, err = tx.ExecContext(ctx, `
        INSERT INTO investment.transaction_entries (
            entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
            quantity, unit_price_amount, unit_price_currency,
            gross_amount, gross_currency, commission_amount, commission_currency,
            tax_amount, tax_currency, trade_date, settlement_date, note,
            correction_reason, prior_entry_id, reverses_transaction_id,
            source_kind, source_file_hash, source_account_label,
            source_broker_operation_key, source_fingerprint, source_identity_version,
            created_at, request_id, trace_id, ledger_sequence
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9,
            $10, 'RUB', $11, 'RUB',
            $12, 'RUB', $13, $14, $15,
            $16, $17, NULL,
            'MANUAL', NULL, '', NULL, NULL, NULL,
            $18, NULLIF($19, '')::uuid, NULLIF($20, ''), $21
        )
    `, reservation.ID, request.TransactionID, request.PortfolioID, assetID, current.Revision+1, request.Corrected.TransactionType,
        decimalString(request.Corrected.Quantity), moneyAmount(request.Corrected.UnitPrice), moneyCurrency(request.Corrected.UnitPrice),
        gross.Amount.String(), request.Corrected.Commission.Amount.String(), request.Corrected.Tax.Amount.String(),
        request.Corrected.TradeDate, request.Corrected.SettlementDate, request.Corrected.Note,
        request.Reason, current.EntryID, command.Now, command.RequestID, command.TraceID, ledgerSequence)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }

    if _, _, err := rebuildPortfolioPositionsTx(ctx, tx, request.PortfolioID, ""); err != nil {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
    }
    affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, []string{current.TradeDate, request.Corrected.TradeDate})
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
    }
    transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, reservation.ID)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    artifact, err := build(transaction)
    if err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
        return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
    }
    if err := tx.Commit(); err != nil { return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err }
    return transaction, artifact, nil
}

func (s *Store) ReverseTransactionWithReplayStage374(
    ctx context.Context,
    command verticalslice.CommandContext,
    request verticalslice.ReverseTransactionRequest,
    build verticalslice.TransactionReversalReplayBuilder,
) (verticalslice.TransactionReversal, verticalslice.CommandReplayArtifact, error) {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    defer rollback(tx)

    reservation, err := reserveReplayCommand(ctx, tx, command, "DELETE")
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    if reservation.Duplicate {
        if err := tx.Commit(); err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
        return verticalslice.TransactionReversal{}, reservation.Artifact, nil
    }
    if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
    }
    if err := validateEffectiveLedgerShapeTx(ctx, tx, request.PortfolioID); err != nil {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
    }
    current, err := latestLogicalEntryTx(ctx, tx, request.PortfolioID, request.TransactionID)
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    reversed, err := logicalTransactionReversedTx(ctx, tx, request.PortfolioID, request.TransactionID)
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    if reversed || current.Revision != request.ExpectedRevision {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, verticalslice.ErrTransactionConflict
    }
    ledgerSequence, err := nextLedgerSequenceTx(ctx, tx, request.PortfolioID)
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    reversalTransactionID := uuid.NewString()
    _, err = tx.ExecContext(ctx, `
        INSERT INTO investment.transaction_entries (
            entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
            quantity, unit_price_amount, unit_price_currency,
            gross_amount, gross_currency, commission_amount, commission_currency,
            tax_amount, tax_currency, trade_date, settlement_date, note,
            correction_reason, prior_entry_id, reverses_transaction_id,
            source_kind, source_file_hash, source_account_label,
            source_broker_operation_key, source_fingerprint, source_identity_version,
            created_at, request_id, trace_id, ledger_sequence
        ) VALUES (
            $1, $2, $3, $4, 1, $5,
            $6::numeric, $7::numeric, CASE WHEN $7::numeric IS NULL THEN NULL ELSE 'RUB' END,
            $8::numeric, 'RUB', $9::numeric, 'RUB', $10::numeric, 'RUB',
            $11::date, $12::date, $13,
            $14, NULL, $15,
            'MANUAL', NULL, '', NULL, NULL, NULL,
            $16, NULLIF($17, '')::uuid, NULLIF($18, ''), $19
        )
    `, reservation.ID, reversalTransactionID, request.PortfolioID, nullStringValue(current.AssetID), current.TransactionType,
        nullStringValue(current.Quantity), nullStringValue(current.UnitPrice), current.GrossAmount, current.Commission, current.Tax,
        request.EffectiveDate, nullStringValue(current.SettlementDate), nullStringValue(current.Note), request.Reason, request.TransactionID,
        command.Now, command.RequestID, command.TraceID, ledgerSequence)
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }

    if _, _, err := rebuildPortfolioPositionsTx(ctx, tx, request.PortfolioID, ""); err != nil {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
    }
    affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, []string{request.EffectiveDate})
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
    }
    result := verticalslice.TransactionReversal{
        TransactionID: request.TransactionID,
        ReversalTransactionID: reversalTransactionID,
        Status: "REVERSED",
        EffectiveDate: request.EffectiveDate,
    }
    artifact, err := build(result)
    if err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
        return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
    }
    if err := tx.Commit(); err != nil { return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err }
    return result, artifact, nil
}

func latestLogicalEntryTx(ctx context.Context, tx *sql.Tx, portfolioID string, transactionID string) (stage374LogicalEntry, error) {
    var entry stage374LogicalEntry
    err := tx.QueryRowContext(ctx, `
        SELECT entry_id::text, transaction_id::text, revision, transaction_type,
               asset_id::text, quantity::text, unit_price_amount::text,
               gross_amount::text, commission_amount::text, tax_amount::text,
               trade_date::text, settlement_date::text, note
        FROM investment.transaction_entries
        WHERE portfolio_id = $1
          AND transaction_id = $2
          AND reverses_transaction_id IS NULL
        ORDER BY revision DESC
        LIMIT 1
    `, portfolioID, transactionID).Scan(
        &entry.EntryID, &entry.TransactionID, &entry.Revision, &entry.TransactionType,
        &entry.AssetID, &entry.Quantity, &entry.UnitPrice,
        &entry.GrossAmount, &entry.Commission, &entry.Tax,
        &entry.TradeDate, &entry.SettlementDate, &entry.Note,
    )
    if errors.Is(err, sql.ErrNoRows) { return stage374LogicalEntry{}, ErrNotFound }
    return entry, err
}

func logicalTransactionReversedTx(ctx context.Context, tx *sql.Tx, portfolioID string, transactionID string) (bool, error) {
    var count int
    if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM investment.transaction_entries
        WHERE portfolio_id = $1 AND reverses_transaction_id = $2
    `, portfolioID, transactionID).Scan(&count); err != nil {
        return false, err
    }
    if count > 1 { return false, ErrUnsupportedPositionLedger }
    return count == 1, nil
}

func nullStringValue(value sql.NullString) any {
    if !value.Valid { return nil }
    return strings.TrimSpace(value.String)
}
''')

write("backend-go/internal/httpapi/stage_03_74_transactions.go", r'''
package httpapi

import (
    "encoding/json"
    "net/http"

    "github.com/gofiber/fiber/v3"

    "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type correctTransactionRequestDTO struct {
    ExpectedRevision int                         `json:"expectedRevision"`
    Reason           string                      `json:"reason"`
    Corrected        appendTransactionRequestDTO `json:"corrected"`
}

type reverseTransactionRequestDTO struct {
    ExpectedRevision int    `json:"expectedRevision"`
    Reason           string `json:"reason"`
    EffectiveDate    string `json:"effectiveDate"`
}

type transactionReversalDTO struct {
    TransactionID         string `json:"transactionId"`
    ReversalTransactionID string `json:"reversalTransactionId"`
    Status                string `json:"status"`
    EffectiveDate         string `json:"effectiveDate"`
}

func (api *API) correctTransactionReplay(c fiber.Ctx) error {
    meta := requestMeta(c)
    subjectID, err := api.subjectID(c)
    if err != nil { return writeMappedErrorWithMeta(c, meta, err) }
    var request correctTransactionRequestDTO
    if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
        return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
    }
    if !nestedJSONFieldPresent(c.Request().Body(), "corrected", "settlementDate") {
        return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "corrected.settlementDate is required")
    }
    corrected, err := request.Corrected.toApp(c.Params("portfolioId"))
    if err != nil { return writeMappedErrorWithMeta(c, meta, err) }
    appRequest := verticalslice.CorrectTransactionRequest{
        PortfolioID: c.Params("portfolioId"),
        TransactionID: c.Params("transactionId"),
        ExpectedRevision: request.ExpectedRevision,
        Reason: request.Reason,
        Corrected: corrected,
    }
    _, artifact, err := api.service.CorrectTransactionWithReplay(
        c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), appRequest,
        func(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
            return buildCommandReplayArtifact(meta, http.StatusOK, mapTransaction(transaction))
        },
    )
    if err != nil { return writeReplayAwareError(c, meta, err) }
    return writeCommandReplayArtifact(c, artifact)
}

func (api *API) reverseTransactionReplay(c fiber.Ctx) error {
    meta := requestMeta(c)
    subjectID, err := api.subjectID(c)
    if err != nil { return writeMappedErrorWithMeta(c, meta, err) }
    var request reverseTransactionRequestDTO
    if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
        return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
    }
    appRequest := verticalslice.ReverseTransactionRequest{
        PortfolioID: c.Params("portfolioId"),
        TransactionID: c.Params("transactionId"),
        ExpectedRevision: request.ExpectedRevision,
        Reason: request.Reason,
        EffectiveDate: request.EffectiveDate,
    }
    _, artifact, err := api.service.ReverseTransactionWithReplay(
        c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), appRequest,
        func(result verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
            return buildCommandReplayArtifact(meta, http.StatusOK, transactionReversalDTO{
                TransactionID: result.TransactionID,
                ReversalTransactionID: result.ReversalTransactionID,
                Status: result.Status,
                EffectiveDate: result.EffectiveDate,
            })
        },
    )
    if err != nil { return writeReplayAwareError(c, meta, err) }
    return writeCommandReplayArtifact(c, artifact)
}

func nestedJSONFieldPresent(body []byte, objectField string, field string) bool {
    var raw map[string]json.RawMessage
    if json.Unmarshal(body, &raw) != nil { return false }
    nestedRaw, ok := raw[objectField]
    if !ok { return false }
    var nested map[string]json.RawMessage
    if json.Unmarshal(nestedRaw, &nested) != nil { return false }
    _, ok = nested[field]
    return ok
}
''')

# Position history validation now materializes the supported effective ledger instead of rejecting every correction/reversal.
regex_once(
    "backend-go/internal/postgres/position_rebuild.go",
    r"func validatePositionHistoryTx\(ctx context\.Context, tx \*sql\.Tx, portfolioID string, assetID \*string\) error \{.*?\n\}\n\nfunc portfolioAcquisitionValuesTx",
    '''func validatePositionHistoryTx(ctx context.Context, tx *sql.Tx, portfolioID string, assetID *string) error {
\tif assetID == nil {
\t\treturn nil
\t}
\t_, _, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, "")
\treturn err
}

func portfolioAcquisitionValuesTx''',
)

write("backend-go/internal/postgres/position_projection_rebuild.go", r'''
package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/openinvest/openinvest/backend-go/internal/position"
    "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type rebuiltPortfolioPosition struct {
    AssetID   string
    Ticker    string
    AssetType string
    State     position.State
}

func rebuildPortfolioPositionsTx(
    ctx context.Context,
    tx *sql.Tx,
    portfolioID string,
    asOfDate string,
) ([]rebuiltPortfolioPosition, *string, error) {
    ledgerRows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, asOfDate)
    if err != nil { return nil, nil, err }

    positions := make([]rebuiltPortfolioPosition, 0)
    positionIndex := map[string]int{}
    var latestIncludedTradeDate *string
    for _, row := range ledgerRows {
        if row.TransactionType != "BUY" && row.TransactionType != "SELL" { continue }
        if row.AssetID == nil || row.Ticker == nil || row.AssetType == nil || row.Quantity == nil || row.UnitPrice == nil {
            return nil, nil, ErrUnsupportedPositionLedger
        }
        index, ok := positionIndex[*row.AssetID]
        if !ok {
            index = len(positions)
            positionIndex[*row.AssetID] = index
            positions = append(positions, rebuiltPortfolioPosition{
                AssetID: *row.AssetID,
                Ticker: *row.Ticker,
                AssetType: *row.AssetType,
                State: position.Empty(),
            })
        }
        next, err := position.Apply(positions[index].State, position.Trade{
            Type: row.TransactionType,
            Quantity: *row.Quantity,
            UnitPrice: *row.UnitPrice,
        })
        switch {
        case errors.Is(err, position.ErrInsufficientQuantity):
            return nil, nil, verticalslice.ErrInsufficientPositionQuantity
        case errors.Is(err, position.ErrDerivedOverflow), errors.Is(err, position.ErrInvalidTrade):
            return nil, nil, fmt.Errorf("%w: position rebuild exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
        case err != nil:
            return nil, nil, err
        }
        positions[index].State = next
        if latestIncludedTradeDate == nil || row.TradeDate > *latestIncludedTradeDate {
            value := row.TradeDate
            latestIncludedTradeDate = &value
        }
    }
    return positions, latestIncludedTradeDate, nil
}
''')

write("backend-go/internal/postgres/stage_03_71_snapshot.go", r'''
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"

    "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const stage371SnapshotMethodology = "stage-03-71-position-cost-snapshot-v1"

func rebuildSnapshotStage371(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string, now time.Time) error {
    positionValues, err := portfolioAcquisitionValuesTx(ctx, tx, portfolioID, snapshotDate)
    if err != nil { return err }
    cashValue, investedCapital, watermark, err := effectiveSnapshotCashTx(ctx, tx, portfolioID, snapshotDate)
    if err != nil { return err }

    snapshotID := uuid.NewString()
    result, err := tx.ExecContext(ctx, `
        INSERT INTO analytics.portfolio_snapshots (
            id, portfolio_id, snapshot_date,
            total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
            invested_capital_amount, nominal_return_rate, real_return_rate,
            snapshot_version, methodology_version, input_watermark, calculated_at
        )
        WITH computed AS (
            SELECT
                $5::numeric AS cash_value,
                $6::numeric AS stock_value,
                $7::numeric AS bond_value,
                $8::numeric AS invested_capital,
                $5::numeric + $6::numeric + $7::numeric AS total_value,
                CASE WHEN $8::numeric > 0
                    THEN (($5::numeric + $6::numeric + $7::numeric) - $8::numeric) / $8::numeric
                    ELSE 0
                END AS nominal_return_rate
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
                  AND methodology_version = $9
            ), 1),
            $9, $10, $4
        FROM computed
        WHERE
            round(total_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(cash_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(stock_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(bond_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(invested_capital, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(nominal_return_rate, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
    `, snapshotID, portfolioID, snapshotDate, now,
        cashValue.String(), positionValues.Stock.String(), positionValues.Bond.String(), investedCapital.String(),
        stage371SnapshotMethodology, watermark)
    if err != nil { return err }
    rowsAffected, err := result.RowsAffected()
    if err != nil { return err }
    if rowsAffected != 1 {
        return fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8) storage precision", verticalslice.ErrInvalidInput)
    }
    return nil
}
''')

# Current transaction list becomes one row per logical transaction and hides reversal command rows.
replace_once(
    "backend-go/internal/postgres/store.go",
    "rows, err := s.db.QueryContext(ctx, transactionSelectSQL()+`\n\t\t\tWHERE `+strings.Join(conditions, \" AND \")+`",
    "rows, err := s.db.QueryContext(ctx, currentTransactionSelectSQL()+`\n\t\t\tWHERE `+strings.Join(conditions, \" AND \")+`",
)
replace_once(
    "backend-go/internal/postgres/store.go",
    "\t\t\tte.revision,\n\t\t\tte.created_at,\n\t\t\tte.created_at AS updated_at",
    "\t\t\tte.revision,\n\t\t\t(SELECT min(first_entry.created_at) FROM investment.transaction_entries first_entry WHERE first_entry.transaction_id = te.transaction_id AND first_entry.reverses_transaction_id IS NULL) AS created_at,\n\t\t\tte.created_at AS updated_at",
)

# Wire canonical PATCH/DELETE routes in both constructors.
replace_once(
    "backend-go/internal/httpapi/routes.go",
    '\tapp.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransaction)\n',
    '\tapp.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransaction)\n\tapp.Patch("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.correctTransactionReplay)\n\tapp.Delete("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.reverseTransactionReplay)\n',
)
replace_once(
    "backend-go/internal/httpapi/replay_app.go",
    '\tapp.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransactionReplay)\n',
    '\tapp.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransactionReplay)\n\tapp.Patch("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.correctTransactionReplay)\n\tapp.Delete("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.reverseTransactionReplay)\n',
)
replace_once(
    "backend-go/internal/httpapi/transport_helpers.go",
    '\tcase errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):\n\t\treturn writeErrorWithMeta(c, meta, http.StatusConflict, "INSUFFICIENT_POSITION_QUANTITY", "Transaction would make the position quantity negative")\n',
    '\tcase errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):\n\t\treturn writeErrorWithMeta(c, meta, http.StatusConflict, "INSUFFICIENT_POSITION_QUANTITY", "Transaction would make the position quantity negative")\n\tcase errors.Is(err, verticalslice.ErrTransactionConflict):\n\t\treturn writeErrorWithMeta(c, meta, http.StatusConflict, "CONFLICT", "Transaction revision is stale or the transaction is already reversed")\n',
)

# Frontend typed API.
replace_once(
    "frontend-next/src/common/api/openinvest.ts",
    "export type CreateTransactionPayload = {\n",
    "export type CorrectTransactionPayload = {\n  expectedRevision: number;\n  reason: string;\n  corrected: CreateTransactionPayload;\n};\n\nexport type ReverseTransactionPayload = {\n  expectedRevision: number;\n  reason: string;\n  effectiveDate: string;\n};\n\nexport type TransactionReversal = {\n  transactionId: string;\n  reversalTransactionId: string;\n  status: \"REVERSED\";\n  effectiveDate: string;\n};\n\nexport type CreateTransactionPayload = {\n",
)
replace_once(
    "frontend-next/src/common/api/openinvest.ts",
    "export async function reviewPortfolioImport(\n",
    '''export async function correctTransaction(
  portfolioId: string,
  transactionId: string,
  payload: CorrectTransactionPayload,
  auth: IdempotentAuthenticatedRequest,
): Promise<ApiResult<Transaction>> {
  return request<Transaction>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions/${encodeURIComponent(transactionId)}`, {
    method: "PATCH",
    headers: { ...idempotentHeaders(auth.idempotencyKey), ...bearerHeaders(auth.accessToken) },
    body: JSON.stringify(payload),
  });
}

export async function reverseTransaction(
  portfolioId: string,
  transactionId: string,
  payload: ReverseTransactionPayload,
  auth: IdempotentAuthenticatedRequest,
): Promise<ApiResult<TransactionReversal>> {
  return request<TransactionReversal>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions/${encodeURIComponent(transactionId)}`, {
    method: "DELETE",
    headers: { ...idempotentHeaders(auth.idempotencyKey), ...bearerHeaders(auth.accessToken) },
    body: JSON.stringify(payload),
  });
}

export async function reviewPortfolioImport(
''',
)

write("frontend-next/src/features/portfolio/components/TransactionRepairControls.tsx", r'''
"use client";

import { useState } from "react";

import {
  correctTransaction,
  reverseTransaction,
  type CreateTransactionPayload,
  type Transaction,
} from "@/common/api/openinvest";

type Props = {
  accessToken: string;
  portfolioId: string;
  transaction: Transaction;
  onMutated: () => Promise<void>;
};

export function TransactionRepairControls({ accessToken, portfolioId, transaction, onMutated }: Props) {
  const [mode, setMode] = useState<"edit" | "reverse" | null>(null);
  const [quantity, setQuantity] = useState(transaction.quantity ?? "");
  const [unitPrice, setUnitPrice] = useState(transaction.unitPrice?.amount ?? "");
  const [grossAmount, setGrossAmount] = useState(transaction.grossAmount.amount);
  const [tradeDate, setTradeDate] = useState(transaction.tradeDate);
  const [reason, setReason] = useState("");
  const [effectiveDate, setEffectiveDate] = useState(transaction.tradeDate);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  if (transaction.status === "REVERSED") {
    return <span className="muted">Reversed</span>;
  }

  async function saveCorrection() {
    if (reason.trim() === "") {
      setMessage("Reason for correction is required.");
      return;
    }
    setBusy(true);
    setMessage(null);
    const corrected = correctedPayload(transaction, quantity, unitPrice, grossAmount, tradeDate);
    const result = await correctTransaction(
      portfolioId,
      transaction.id,
      { expectedRevision: transaction.revision, reason: reason.trim(), corrected },
      { accessToken, idempotencyKey: crypto.randomUUID() },
    );
    setBusy(false);
    if (!result.ok) {
      setMessage(result.message);
      return;
    }
    setMode(null);
    setReason("");
    setMessage(`Saved revision ${result.data.revision}.`);
    await onMutated();
  }

  async function saveReversal() {
    if (reason.trim() === "" || effectiveDate === "") {
      setMessage("Reason and effective date are required.");
      return;
    }
    setBusy(true);
    setMessage(null);
    const result = await reverseTransaction(
      portfolioId,
      transaction.id,
      { expectedRevision: transaction.revision, reason: reason.trim(), effectiveDate },
      { accessToken, idempotencyKey: crypto.randomUUID() },
    );
    setBusy(false);
    if (!result.ok) {
      setMessage(result.message);
      return;
    }
    setMode(null);
    setReason("");
    setMessage(`Reversed effective ${result.data.effectiveDate}.`);
    await onMutated();
  }

  return (
    <div>
      <div className="button-row" aria-label={`Actions for ${transaction.transactionType} ${transaction.ticker ?? "cash"}`}>
        <button type="button" className="secondary-button" disabled={busy} onClick={() => { setMessage(null); setMode("edit"); }}>
          Edit
        </button>
        <button type="button" className="secondary-button" disabled={busy} aria-describedby={`reverse-help-${transaction.id}`} onClick={() => { setMessage(null); setMode("reverse"); }}>
          Reverse
        </button>
      </div>

      {mode === "edit" ? (
        <fieldset disabled={busy}>
          <legend>Edit transaction · revision {transaction.revision}</legend>
          {transaction.transactionType === "BUY" || transaction.transactionType === "SELL" ? (
            <>
              <label>Quantity<input value={quantity} onChange={(event) => setQuantity(event.target.value)} inputMode="decimal" /></label>
              <label>Unit price (RUB)<input value={unitPrice} onChange={(event) => setUnitPrice(event.target.value)} inputMode="decimal" /></label>
            </>
          ) : (
            <label>Gross amount (RUB)<input value={grossAmount} onChange={(event) => setGrossAmount(event.target.value)} inputMode="decimal" /></label>
          )}
          <label>Trade date<input type="date" value={tradeDate} onChange={(event) => setTradeDate(event.target.value)} /></label>
          <label>Reason for correction<input value={reason} maxLength={300} required onChange={(event) => setReason(event.target.value)} /></label>
          <div className="button-row">
            <button type="button" className="primary-button" onClick={() => void saveCorrection()}>{busy ? "Saving…" : "Save correction"}</button>
            <button type="button" className="secondary-button" onClick={() => setMode(null)}>Cancel</button>
          </div>
        </fieldset>
      ) : null}

      {mode === "reverse" ? (
        <fieldset disabled={busy}>
          <legend>Reverse transaction</legend>
          <p id={`reverse-help-${transaction.id}`} className="muted">History will be preserved; the operation is cancelled by a separate immutable ledger entry.</p>
          <label>Effective date<input type="date" value={effectiveDate} onChange={(event) => setEffectiveDate(event.target.value)} /></label>
          <label>Reason for reversal<input value={reason} maxLength={300} required onChange={(event) => setReason(event.target.value)} /></label>
          <div className="button-row">
            <button type="button" className="primary-button" onClick={() => void saveReversal()}>{busy ? "Reversing…" : "Confirm reversal"}</button>
            <button type="button" className="secondary-button" onClick={() => setMode(null)}>Cancel</button>
          </div>
        </fieldset>
      ) : null}
      {message ? <p className="muted" role="status" aria-live="polite">{message}</p> : null}
    </div>
  );
}

function correctedPayload(
  transaction: Transaction,
  quantity: string,
  unitPrice: string,
  grossAmount: string,
  tradeDate: string,
): CreateTransactionPayload {
  const isTrade = transaction.transactionType === "BUY" || transaction.transactionType === "SELL";
  return {
    transactionType: transaction.transactionType,
    ticker: transaction.ticker,
    quantity: isTrade ? quantity : transaction.quantity,
    unitPrice: isTrade ? { amount: unitPrice, currency: "RUB" } : null,
    grossAmount: isTrade ? undefined : { amount: grossAmount, currency: "RUB" },
    commission: transaction.commission,
    tax: transaction.tax,
    tradeDate,
    settlementDate: transaction.settlementDate,
    note: transaction.note ?? null,
  };
}
''')

replace_once(
    "frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx",
    'import { PositionsBlock } from "@/features/portfolio/components/PositionsBlock";\n',
    'import { PositionsBlock } from "@/features/portfolio/components/PositionsBlock";\nimport { TransactionRepairControls } from "@/features/portfolio/components/TransactionRepairControls";\n',
)
replace_once(
    "frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx",
    "                  <th>Status</th>\n",
    "                  <th>Status</th>\n                  <th>Actions</th>\n",
)
replace_once(
    "frontend-next/src/features/portfolio/components/PortfolioDetailSlice.tsx",
    "                    <td>{transaction.status}</td>\n",
    '''                    <td>{transaction.status}{transaction.revision > 1 ? ` · revision ${transaction.revision}` : ""}</td>
                    <td>
                      <TransactionRepairControls
                        accessToken={accessToken}
                        portfolioId={portfolioId}
                        transaction={transaction}
                        onMutated={refreshAfterLedgerMutation}
                      />
                    </td>
''',
)

write("frontend-next/tests/stage-03-74-transaction-repair.test.mjs", r'''
import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const api = fs.readFileSync(new URL("../src/common/api/openinvest.ts", import.meta.url), "utf8");
const detail = fs.readFileSync(new URL("../src/features/portfolio/components/PortfolioDetailSlice.tsx", import.meta.url), "utf8");
const controls = fs.readFileSync(new URL("../src/features/portfolio/components/TransactionRepairControls.tsx", import.meta.url), "utf8");

test("Stage 3.74 uses the frozen PATCH and DELETE transaction commands", () => {
  assert.match(api, /method: "PATCH"/);
  assert.match(api, /method: "DELETE"/);
  assert.match(api, /expectedRevision/);
  assert.match(api, /effectiveDate/);
  assert.match(api, /IdempotentAuthenticatedRequest/);
});

test("Stage 3.74 exposes explicit Edit and Reverse UX without delete-permanently language", () => {
  assert.match(detail, /<th>Actions<\/th>/);
  assert.match(controls, />Edit</);
  assert.match(controls, />Reverse</);
  assert.match(controls, /History will be preserved/);
  assert.doesNotMatch(controls, /Delete permanently/i);
  assert.match(controls, /Reason for correction/);
  assert.match(controls, /Reason for reversal/);
  assert.match(controls, /aria-live="polite"/);
});
''')

write("backend-go/internal/postgres/stage_03_74_transactions_integration_test.go", r'''
package postgres_test

import (
    "errors"
    "sync"
    "testing"

    "github.com/google/uuid"

    "github.com/openinvest/openinvest/backend-go/internal/decimal"
    "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage374Artifact(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
    return verticalslice.CommandReplayArtifact{StatusCode: 200, Body: []byte(transaction.ID)}, nil
}

func correctStage374(t *testing.T, h stage371Harness, transaction verticalslice.Transaction, quantity, price, tradeDate, key string) (verticalslice.Transaction, error) {
    t.Helper()
    q := decimal.Must(quantity)
    p := verticalslice.Money{Amount: decimal.Must(price), Currency: verticalslice.RUB}
    corrected := verticalslice.AppendTransactionRequest{
        PortfolioID: h.portfolioID,
        TransactionType: transaction.TransactionType,
        Ticker: transaction.Ticker,
        Quantity: &q,
        UnitPrice: &p,
        Commission: transaction.Commission,
        Tax: transaction.Tax,
        TradeDate: tradeDate,
        SettlementDate: transaction.SettlementDate,
        Note: transaction.Note,
    }
    result, _, err := h.service.CorrectTransactionWithReplay(
        h.ctx, verticalslice.RequestContext{}, h.subjectID, key,
        "/api/v1/portfolios/"+h.portfolioID+"/transactions/"+transaction.ID,
        verticalslice.CorrectTransactionRequest{
            PortfolioID: h.portfolioID,
            TransactionID: transaction.ID,
            ExpectedRevision: transaction.Revision,
            Reason: "Incorrect broker value",
            Corrected: corrected,
        },
        stage374Artifact,
    )
    return result, err
}

func reverseStage374(t *testing.T, h stage371Harness, transaction verticalslice.Transaction, effectiveDate, key string) (verticalslice.TransactionReversal, error) {
    t.Helper()
    result, _, err := h.service.ReverseTransactionWithReplay(
        h.ctx, verticalslice.RequestContext{}, h.subjectID, key,
        "/api/v1/portfolios/"+h.portfolioID+"/transactions/"+transaction.ID,
        verticalslice.ReverseTransactionRequest{
            PortfolioID: h.portfolioID,
            TransactionID: transaction.ID,
            ExpectedRevision: transaction.Revision,
            Reason: "Broker cancelled transaction",
            EffectiveDate: effectiveDate,
        },
        func(result verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
            return verticalslice.CommandReplayArtifact{StatusCode: 200, Body: []byte(result.ReversalTransactionID)}, nil
        },
    )
    return result, err
}

func TestStage374CorrectionRebuildsCurrentAndHistoricalPositions(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 correction")
    original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "300.00000000", "2026-03-01"))
    corrected, err := correctStage374(t, h, original, "100.00000000", "280.00000000", "2026-03-01", uuid.NewString())
    if err != nil { t.Fatalf("correct transaction: %v", err) }
    if corrected.ID != original.ID || corrected.Revision != 2 || corrected.Status != "CORRECTED" {
        t.Fatalf("correction projection mismatch: %+v", corrected)
    }
    current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
    if err != nil { t.Fatalf("current positions: %v", err) }
    if len(current.Items) != 1 || current.Items[0].WeightedAverageCost.Amount.String() != "280.00000000" || current.Items[0].AcquisitionBasis.Amount.String() != "28000.00000000" {
        t.Fatalf("corrected current position mismatch: %+v", current.Items)
    }
    before, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-02-01")
    if err != nil { t.Fatalf("historical before trade: %v", err) }
    if len(before.Items) != 0 { t.Fatalf("position must not exist before corrected trade date: %+v", before.Items) }
    after, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-03-15")
    if err != nil { t.Fatalf("historical after trade: %v", err) }
    if len(after.Items) != 1 || after.Items[0].WeightedAverageCost.Amount.String() != "280.00000000" {
        t.Fatalf("historical correction not reflected: %+v", after.Items)
    }
    assertContiguousLedgerSequence(t, h, 2)
}

func TestStage374SecondCorrectionAndStaleRevisionConflict(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 revisions")
    original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
    revision2, err := correctStage374(t, h, original, "10.00000000", "110.00000000", "2026-01-01", uuid.NewString())
    if err != nil { t.Fatalf("first correction: %v", err) }
    revision3, err := correctStage374(t, h, revision2, "10.00000000", "120.00000000", "2026-01-01", uuid.NewString())
    if err != nil { t.Fatalf("second correction: %v", err) }
    if revision3.Revision != 3 { t.Fatalf("revision got %d want 3", revision3.Revision) }
    _, err = correctStage374(t, h, original, "10.00000000", "130.00000000", "2026-01-01", uuid.NewString())
    if !errors.Is(err, verticalslice.ErrTransactionConflict) { t.Fatalf("stale revision error = %v", err) }
}

func TestStage374CorrectionOversellRollsBackAtomically(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 oversell correction")
    buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "100.00000000", "2026-01-01"))
    appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "80.00000000", "150.00000000", "2026-02-01"))
    _, err := correctStage374(t, h, buy, "50.00000000", "100.00000000", "2026-01-01", uuid.NewString())
    if !errors.Is(err, verticalslice.ErrInsufficientPositionQuantity) { t.Fatalf("oversell correction error = %v", err) }
    var revisions int
    if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_id=$2`, h.portfolioID, buy.ID).Scan(&revisions); err != nil { t.Fatal(err) }
    if revisions != 1 { t.Fatalf("invalid correction must rollback, rows=%d", revisions) }
}

func TestStage374ReversalUsesEffectiveDateForTimeMachine(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 reversal time machine")
    buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-10"))
    reversal, err := reverseStage374(t, h, buy, "2026-03-01", uuid.NewString())
    if err != nil { t.Fatalf("reverse transaction: %v", err) }
    if reversal.Status != "REVERSED" || reversal.TransactionID != buy.ID { t.Fatalf("reversal mismatch: %+v", reversal) }
    before, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-02-15")
    if err != nil { t.Fatalf("before reversal: %v", err) }
    if len(before.Items) != 1 || before.Items[0].Quantity.String() != "10.00000000" { t.Fatalf("historically active before reversal: %+v", before.Items) }
    after, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-03-01")
    if err != nil { t.Fatalf("after reversal: %v", err) }
    if len(after.Items) != 0 { t.Fatalf("reversal must be effective on selected date: %+v", after.Items) }
    current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
    if err != nil { t.Fatalf("current after reversal: %v", err) }
    if len(current.Items) != 0 { t.Fatalf("current must exclude reversed transaction: %+v", current.Items) }
}

func TestStage374ReversingBuyThatCreatesOversellRollsBack(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 oversell reversal")
    buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "100.00000000", "2026-01-01"))
    appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "80.00000000", "150.00000000", "2026-02-01"))
    _, err := reverseStage374(t, h, buy, "2026-01-15", uuid.NewString())
    if !errors.Is(err, verticalslice.ErrInsufficientPositionQuantity) { t.Fatalf("oversell reversal error = %v", err) }
    var reversals int
    if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND reverses_transaction_id=$2`, h.portfolioID, buy.ID).Scan(&reversals); err != nil { t.Fatal(err) }
    if reversals != 0 { t.Fatalf("invalid reversal must rollback, rows=%d", reversals) }
}

func TestStage374ConcurrentCorrectionsOnlyOneWins(t *testing.T) {
    h := newStage371Harness(t, "Stage 3.74 concurrent correction")
    original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
    start := make(chan struct{})
    errorsOut := make(chan error, 2)
    var wg sync.WaitGroup
    for _, price := range []string{"110.00000000", "120.00000000"} {
        wg.Add(1)
        go func(price string) {
            defer wg.Done()
            <-start
            _, err := correctStage374(t, h, original, "10.00000000", price, "2026-01-01", uuid.NewString())
            errorsOut <- err
        }(price)
    }
    close(start)
    wg.Wait()
    close(errorsOut)
    success, conflicts := 0, 0
    for err := range errorsOut {
        if err == nil { success++ } else if errors.Is(err, verticalslice.ErrTransactionConflict) { conflicts++ } else { t.Fatalf("unexpected concurrent error: %v", err) }
    }
    if success != 1 || conflicts != 1 { t.Fatalf("concurrent results success=%d conflicts=%d", success, conflicts) }
}
''')

write("docs/stages/STAGE_03_74_TRANSACTION_CORRECTION_REVERSAL_IMPLEMENTATION.md", r'''
# Stage 3.74 — Transaction Correction & Reversal / Ledger Repair UX

| Field | Value |
| --- | --- |
| Stage | 3.74 |
| Status | IMPLEMENTATION CANDIDATE — canonical only after reviewed protected merge |
| Canonical base | `develop@18fff1150ac8d1d3c4eafbf7065d38d0bbb47f55` |
| Runtime budget | 0 ₽ |

## Purpose

Stage 3.74 closes the user-facing ledger-repair gap without weakening the immutable financial ledger. A correction appends the next revision of the same logical transaction. A reversal appends a separate immutable command row with an explicit BusinessDate. No existing financial row is updated or deleted.

## Frozen contract reused

The implementation wires the already-frozen `correctTransaction` PATCH and `reverseTransaction` DELETE operations. `expectedRevision`, mandatory reasons, reversal `effectiveDate`, `Idempotency-Key`, the standard error envelope and the existing transaction/reversal response schemas remain authoritative. DELETE is command semantics only.

## Effective ledger materialization

Position and snapshot consumers no longer interpret raw immutable rows as independent economic events. The supported pipeline is:

`raw immutable entries → validated revision chains → latest logical transaction → reversal effective-date filter → deterministic effective ledger → Stage 3.71 position.Apply/Rebuild`.

Corrections are retroactive truth: the latest valid revision supplies the transaction fields for historical reconstruction. Reversal becomes effective on its explicit `effectiveDate`; before that BusinessDate the logical transaction remains visible in Time Machine, on/after it the transaction is excluded. Same-BusinessDate ordering uses the logical transaction's first ledger sequence, so a correction does not accidentally move an existing logical transaction merely because its revision was appended later.

Malformed revision chains, malformed reversal targets or duplicate reversals fail closed with `ErrUnsupportedPositionLedger`.

## Atomicity and concurrency

All correction/reversal mutations reuse the existing portfolio row `FOR UPDATE` serialization. `UNIQUE(transaction_id, revision)` remains the database backstop against duplicate revision numbers. Under the portfolio lock the runtime checks `expectedRevision` and reversal state, appends exactly one candidate entry, rebuilds the effective position ledger, rejects historical oversell, rebuilds affected snapshots, persists the exact replay artifact and commits atomically.

A migration is not required. The existing schema already contains `revision`, `prior_entry_id`, `correction_reason`, `reverses_transaction_id` and the unique revision constraint. The runtime role remains `SELECT, INSERT` only on `investment.transaction_entries`; UPDATE/DELETE/TRUNCATE stay revoked.

## Oversell safety

A backdated correction or reversal is validated against the full effective portfolio history before commit. If changing/reversing a BUY makes a later SELL impossible, the command returns the existing insufficient-position conflict and the new ledger row plus snapshot rebuilds roll back together.

## Snapshots and Time Machine

Stage 3.71 snapshot cash/invested-capital inputs and Stage 3.72/3.73 positions now consume the same effective ledger semantics. Correction rebuild planning starts from both the old and corrected trade dates; reversal rebuild planning starts from the reversal effective date. Historical positions therefore remain active before a later reversal and disappear on/after the reversal date.

## Transaction list UX

The main transaction list projects one row per logical transaction. Raw revisions and reversal command rows are not presented as independent trades. The current row exposes `ACTIVE`, `CORRECTED` with current revision, or `REVERSED`. Edit and Reverse actions use the frozen API. The reversal confirmation explicitly states that history is preserved.

The Stage 3.74 UI intentionally edits canonical financial fields needed for ledger repair without adding a forensic revision-history endpoint. A future dedicated audit-detail view would require separately reviewed contract scope.

## Idempotency

PATCH and DELETE are financial commands and use the existing exact-response replay system. Same key + same payload replays the original success without another revision/reversal. Same key + different payload conflicts. Replay resolution occurs before mutable business-state checks so a completed command remains exactly replayable after later ledger changes.

## Security

Portfolio ownership is checked by the existing subject-scoped portfolio lock before target transaction state is read or mutated. A transaction in another subject/portfolio therefore preserves the existing not-found/anti-enumeration boundary.

## Verification vectors

The implementation adds PostgreSQL witnesses for correction WAC/basis, revision 1→2→3, stale revision conflict, correction-induced historical oversell rollback, reversal effective-date Time Machine behavior, reversal-induced oversell rollback and concurrent correction races. Existing Stage 3.71–3.73 tests continue to cover close/reopen, same-date ordering, position projection, market-unavailable semantics and rapid historical-load protection. Frontend contract tests cover PATCH/DELETE, Edit/Reverse copy, mandatory reasons and destructive-action accessibility text.

## Non-scope

No market provider, market value, P/L, XIRR, inflation, corporate-action source activation, broker sync, imported SELL expansion, tax-basis methodology, bond NKD/YTM, AI, notifications, Redis, Kafka, workers, new SaaS or paid dependency is introduced.

## Expected result

A user can correct a mistaken price/quantity/date with a reason, see the current logical revision and recalculated position/WAC/basis, and reverse an operation with reason/effectiveDate while retaining full immutable history. Time Machine reflects the corrected/reversed economic truth, and no accepted repair can create an oversold portfolio.
''')

# Minimal registry synchronization inside the implementation PR.
replace_once(
    "docs/ROADMAP.md",
    "| 3.73 — Portfolio Time Machine / Historical Position View | Turn the existing Stage 3.72 endpoint-local `asOfDate` semantics into user-visible Current ↔ Historical position navigation without a second financial engine, new endpoint, table, cache or provider | Complete / canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; Stage 3.73 lifecycle/documentation closed by the protected merge and this post-merge registry synchronization |",
    "| 3.73 — Portfolio Time Machine / Historical Position View | Turn the existing Stage 3.72 endpoint-local `asOfDate` semantics into user-visible Current ↔ Historical position navigation without a second financial engine, new endpoint, table, cache or provider | Complete / canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; Stage 3.73 lifecycle/documentation closed by the protected merge and this post-merge registry synchronization |\n| 3.74 — Transaction Correction & Reversal / Ledger Repair UX | Wire the frozen correction/reversal commands, preserve append-only auditability, materialize deterministic effective ledger truth, reject historical oversell and expose Edit/Reverse UX | Implementation candidate; canonical only after reviewed protected merge; no new API/provider/paid infrastructure |",
)
replace_once(
    "docs/IMPLEMENTATION_LOG.md",
    "| 3.73 — Portfolio Time Machine / Historical Position View | Expose exact historical holdings by BusinessDate using the existing Stage 3.72 `asOfDate` projection and canonical Stage 3.71 position engine | Complete / canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; no new API, schema, provider or financial algorithm | [Stage 3.73 report](stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md) |",
    "| 3.73 — Portfolio Time Machine / Historical Position View | Expose exact historical holdings by BusinessDate using the existing Stage 3.72 `asOfDate` projection and canonical Stage 3.71 position engine | Complete / canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; no new API, schema, provider or financial algorithm | [Stage 3.73 report](stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md) |\n| 3.74 — Transaction Correction & Reversal / Ledger Repair UX | Append correction revisions and explicit-date reversals, project one current logical transaction, rebuild effective positions/snapshots and reject historical oversell | Implementation candidate; protected merge required for canonical status | [Stage 3.74 report](stages/STAGE_03_74_TRANSACTION_CORRECTION_REVERSAL_IMPLEMENTATION.md) |",
)
replace_once(
    "docs/DOCUMENT_INDEX.md",
    "| Stage 3.73 Portfolio Time Machine / Historical Position View | Complete/canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; Current ↔ Historical UI, stale-result protection and historical financial witnesses over the existing Stage 3.72 `asOfDate` contract; no new endpoint/schema/provider | `stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md` |",
    "| Stage 3.73 Portfolio Time Machine / Historical Position View | Complete/canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` from exact final head `a74fd85a46843ccdfe8192f2ac8687169a5a2ac7` after CI #424 / run `34109389368` 10/10 SUCCESS; Current ↔ Historical UI, stale-result protection and historical financial witnesses over the existing Stage 3.72 `asOfDate` contract; no new endpoint/schema/provider | `stages/STAGE_03_73_PORTFOLIO_TIME_MACHINE_IMPLEMENTATION.md` |\n| Stage 3.74 Transaction Correction & Reversal / Ledger Repair UX | Implementation candidate over frozen PATCH/DELETE commands; append-only revisions/reversals, effective-ledger materialization, oversell rollback, exact replay and Edit/Reverse UX | `stages/STAGE_03_74_TRANSACTION_CORRECTION_REVERSAL_IMPLEMENTATION.md` |",
)
replace_once(
    "README.md",
    "Stage 3 is now implemented through **Stage 3.73 — Portfolio Time Machine / Historical Position View**.",
    "Stage 3 is canonically implemented through **Stage 3.73 — Portfolio Time Machine / Historical Position View**; Stage 3.74 Transaction Correction & Reversal / Ledger Repair UX is the current reviewed implementation candidate.",
)

# Source of Truth gets a candidate frontier note without pre-declaring canonical completion.
needle = "**Implementation frontier: Stage 3.73 — Portfolio Time Machine / Historical Position View is COMPLETE / CANONICAL"
p = Path("docs/SOURCE_OF_TRUTH.md")
s = p.read_text(encoding="utf-8")
idx = s.find(needle)
if idx < 0:
    raise SystemExit("SOURCE_OF_TRUTH Stage 3.73 frontier anchor missing")
line_end = s.find("\n", idx)
if line_end < 0:
    line_end = len(s)
addition = "\n**Current candidate frontier: Stage 3.74 — Transaction Correction & Reversal / Ledger Repair UX implements the already-frozen PATCH/DELETE commands, immutable revision/reversal semantics and effective-ledger materialization. It is non-canonical until its implementation PR passes review/CI and is protected-merged; it activates no provider or new paid infrastructure.**"
s = s[:line_end] + addition + s[line_end:]
p.write_text(s, encoding="utf-8")
