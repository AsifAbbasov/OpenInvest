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
