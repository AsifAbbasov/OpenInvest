-- OpenInvest Stage 3.1 database foundation.
--
-- Scope:
-- - local first vertical slice only;
-- - additive schema/table creation;
-- - no tax tables;
-- - no external-provider tables;
-- - no destructive statements.
--
-- Financial rules:
-- - decimal values use NUMERIC(28, 8);
-- - financial dates use DATE;
-- - system timestamps use TIMESTAMPTZ and must be supplied/used as UTC by the application.

BEGIN;

CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS investment;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE identity.users (
    id UUID PRIMARY KEY,
    email_normalized TEXT NOT NULL,
    account_state TEXT NOT NULL DEFAULT 'active'
        CHECK (account_state IN ('active', 'pending_deletion', 'deleted')),
    language TEXT NOT NULL DEFAULT 'en',
    theme TEXT NOT NULL DEFAULT 'system',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deletion_requested_at TIMESTAMPTZ,
    deletion_completed_at TIMESTAMPTZ,
    CONSTRAINT users_email_normalized_not_blank CHECK (length(trim(email_normalized)) > 0),
    CONSTRAINT users_deletion_order CHECK (
        deletion_completed_at IS NULL
        OR deletion_requested_at IS NOT NULL
    )
);

CREATE UNIQUE INDEX users_active_email_normalized_uidx
    ON identity.users (email_normalized)
    WHERE account_state <> 'deleted';

CREATE TABLE investment.subjects (
    id UUID PRIMARY KEY,
    subject_state TEXT NOT NULL DEFAULT 'active'
        CHECK (subject_state IN ('active', 'anonymous')),
    anonymous_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subjects_anonymous_timestamp CHECK (
        (subject_state = 'anonymous' AND anonymous_at IS NOT NULL)
        OR (subject_state = 'active' AND anonymous_at IS NULL)
    )
);

CREATE TABLE identity.user_investment_links (
    user_id UUID PRIMARY KEY REFERENCES identity.users (id) ON DELETE CASCADE,
    investment_subject_id UUID NOT NULL REFERENCES investment.subjects (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_investment_links_subject_uidx UNIQUE (investment_subject_id)
);

CREATE TABLE investment.assets (
    id UUID PRIMARY KEY,
    ticker TEXT NOT NULL,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('stock', 'bond')),
    name TEXT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (currency = 'RUB'),
    market TEXT NOT NULL DEFAULT 'MOEX' CHECK (market = 'MOEX'),
    lifecycle_status TEXT NOT NULL DEFAULT 'active'
        CHECK (lifecycle_status IN ('active', 'inactive')),
    isin TEXT,
    lot_size NUMERIC(28, 8) NOT NULL DEFAULT 1 CHECK (lot_size > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT assets_ticker_format CHECK (ticker ~ '^[A-Z0-9]{1,32}$'),
    CONSTRAINT assets_name_not_blank CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX assets_ticker_uidx ON investment.assets (ticker);
CREATE INDEX assets_type_status_ticker_idx ON investment.assets (asset_type, lifecycle_status, ticker);

CREATE TABLE investment.portfolios (
    id UUID PRIMARY KEY,
    subject_id UUID NOT NULL REFERENCES investment.subjects (id),
    name TEXT NOT NULL,
    base_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (base_currency = 'RUB'),
    portfolio_state TEXT NOT NULL DEFAULT 'active'
        CHECK (portfolio_state IN ('active', 'removed_from_active_use')),
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    removed_at TIMESTAMPTZ,
    CONSTRAINT portfolios_name_not_blank CHECK (length(trim(name)) > 0),
    CONSTRAINT portfolios_removed_timestamp CHECK (
        (portfolio_state = 'removed_from_active_use' AND removed_at IS NOT NULL)
        OR (portfolio_state = 'active' AND removed_at IS NULL)
    )
);

CREATE INDEX portfolios_subject_state_updated_idx
    ON investment.portfolios (subject_id, portfolio_state, updated_at DESC);

CREATE TABLE investment.transaction_entries (
    entry_id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    portfolio_id UUID NOT NULL REFERENCES investment.portfolios (id),
    asset_id UUID REFERENCES investment.assets (id),
    revision INTEGER NOT NULL CHECK (revision > 0),
    transaction_type TEXT NOT NULL
        CHECK (transaction_type IN ('BUY', 'SELL', 'DIVIDEND', 'COUPON', 'FEE', 'TAX', 'DEPOSIT', 'WITHDRAWAL')),
    quantity NUMERIC(28, 8) CHECK (quantity IS NULL OR quantity >= 0),
    unit_price_amount NUMERIC(28, 8) CHECK (unit_price_amount IS NULL OR unit_price_amount >= 0),
    unit_price_currency CHAR(3) CHECK (unit_price_currency IS NULL OR unit_price_currency = 'RUB'),
    gross_amount NUMERIC(28, 8) NOT NULL CHECK (gross_amount >= 0),
    gross_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (gross_currency = 'RUB'),
    commission_amount NUMERIC(28, 8) NOT NULL DEFAULT 0 CHECK (commission_amount >= 0),
    commission_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (commission_currency = 'RUB'),
    tax_amount NUMERIC(28, 8) NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    tax_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (tax_currency = 'RUB'),
    trade_date DATE NOT NULL,
    settlement_date DATE,
    note TEXT CHECK (note IS NULL OR length(note) <= 500),
    correction_reason TEXT,
    prior_entry_id UUID REFERENCES investment.transaction_entries (entry_id),
    reverses_transaction_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    request_id UUID,
    trace_id TEXT,
    CONSTRAINT transaction_entries_revision_uidx UNIQUE (transaction_id, revision),
    CONSTRAINT transaction_entries_trade_asset_shape CHECK (
        (
            transaction_type IN ('BUY', 'SELL')
            AND asset_id IS NOT NULL
            AND quantity IS NOT NULL
            AND quantity > 0
            AND unit_price_amount IS NOT NULL
            AND unit_price_amount > 0
            AND unit_price_currency = 'RUB'
        )
        OR (
            transaction_type IN ('DIVIDEND', 'COUPON')
            AND asset_id IS NOT NULL
            AND unit_price_amount IS NULL
            AND unit_price_currency IS NULL
        )
        OR (
            transaction_type IN ('FEE', 'TAX')
            AND quantity IS NULL
            AND unit_price_amount IS NULL
            AND unit_price_currency IS NULL
        )
        OR (
            transaction_type IN ('DEPOSIT', 'WITHDRAWAL')
            AND asset_id IS NULL
            AND quantity IS NULL
            AND unit_price_amount IS NULL
            AND unit_price_currency IS NULL
        )
    ),
    CONSTRAINT transaction_entries_no_self_reversal CHECK (
        reverses_transaction_id IS NULL
        OR reverses_transaction_id <> transaction_id
    )
);

CREATE INDEX transaction_entries_portfolio_trade_idx
    ON investment.transaction_entries (portfolio_id, trade_date DESC, entry_id DESC);
CREATE INDEX transaction_entries_portfolio_asset_trade_idx
    ON investment.transaction_entries (portfolio_id, asset_id, trade_date DESC, entry_id DESC);
CREATE INDEX transaction_entries_reversal_idx
    ON investment.transaction_entries (reverses_transaction_id)
    WHERE reverses_transaction_id IS NOT NULL;

CREATE TABLE investment.command_deduplication (
    id UUID PRIMARY KEY,
    principal_id UUID NOT NULL,
    method TEXT NOT NULL,
    canonical_path TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    terminal_status TEXT,
    response_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT command_deduplication_key_uidx
        UNIQUE (principal_id, method, canonical_path, idempotency_key),
    CONSTRAINT command_deduplication_method_not_blank CHECK (length(trim(method)) > 0),
    CONSTRAINT command_deduplication_path_not_blank CHECK (length(trim(canonical_path)) > 0),
    CONSTRAINT command_deduplication_key_not_blank CHECK (length(trim(idempotency_key)) > 0)
);

CREATE TABLE investment.outbox_events (
    event_id UUID PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_version BIGINT NOT NULL CHECK (aggregate_version > 0),
    event_type TEXT NOT NULL,
    event_version INTEGER NOT NULL CHECK (event_version > 0),
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    last_error_code TEXT,
    CONSTRAINT outbox_aggregate_not_blank CHECK (length(trim(aggregate_type)) > 0),
    CONSTRAINT outbox_event_type_not_blank CHECK (length(trim(event_type)) > 0)
);

CREATE INDEX outbox_unpublished_created_idx
    ON investment.outbox_events (created_at, event_id)
    WHERE published_at IS NULL;

CREATE TABLE analytics.portfolio_snapshots (
    id UUID PRIMARY KEY,
    portfolio_id UUID NOT NULL REFERENCES investment.portfolios (id),
    snapshot_date DATE NOT NULL,
    total_value_amount NUMERIC(28, 8) NOT NULL,
    total_value_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (total_value_currency = 'RUB'),
    cash_value_amount NUMERIC(28, 8) NOT NULL,
    cash_value_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (cash_value_currency = 'RUB'),
    stock_value_amount NUMERIC(28, 8) NOT NULL CHECK (stock_value_amount >= 0),
    stock_value_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (stock_value_currency = 'RUB'),
    bond_value_amount NUMERIC(28, 8) NOT NULL CHECK (bond_value_amount >= 0),
    bond_value_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (bond_value_currency = 'RUB'),
    invested_capital_amount NUMERIC(28, 8) NOT NULL CHECK (invested_capital_amount >= 0),
    invested_capital_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (invested_capital_currency = 'RUB'),
    nominal_return_rate NUMERIC(28, 8) NOT NULL,
    real_return_rate NUMERIC(28, 8) NOT NULL,
    snapshot_version INTEGER NOT NULL CHECK (snapshot_version > 0),
    methodology_version TEXT NOT NULL,
    input_watermark TEXT NOT NULL,
    snapshot_status TEXT NOT NULL DEFAULT 'calculated'
        CHECK (snapshot_status IN ('calculated', 'superseded', 'failed')),
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT portfolio_snapshots_methodology_not_blank CHECK (length(trim(methodology_version)) > 0),
    CONSTRAINT portfolio_snapshots_watermark_not_blank CHECK (length(trim(input_watermark)) > 0),
    CONSTRAINT portfolio_snapshots_identity_uidx
        UNIQUE (portfolio_id, snapshot_date, snapshot_version, methodology_version)
);

CREATE INDEX portfolio_snapshots_portfolio_date_idx
    ON analytics.portfolio_snapshots (portfolio_id, snapshot_date DESC);

CREATE TABLE analytics.snapshot_positions (
    id UUID PRIMARY KEY,
    snapshot_id UUID NOT NULL REFERENCES analytics.portfolio_snapshots (id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES investment.assets (id),
    quantity NUMERIC(28, 8) NOT NULL CHECK (quantity >= 0),
    weighted_average_cost_amount NUMERIC(28, 8) NOT NULL CHECK (weighted_average_cost_amount >= 0),
    weighted_average_cost_currency CHAR(3) NOT NULL DEFAULT 'RUB'
        CHECK (weighted_average_cost_currency = 'RUB'),
    market_price_amount NUMERIC(28, 8) NOT NULL CHECK (market_price_amount >= 0),
    market_price_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (market_price_currency = 'RUB'),
    market_value_amount NUMERIC(28, 8) NOT NULL CHECK (market_value_amount >= 0),
    market_value_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (market_value_currency = 'RUB'),
    unrealized_gain_amount NUMERIC(28, 8) NOT NULL,
    unrealized_gain_currency CHAR(3) NOT NULL DEFAULT 'RUB' CHECK (unrealized_gain_currency = 'RUB'),
    portfolio_weight NUMERIC(28, 8) NOT NULL CHECK (portfolio_weight >= 0),
    CONSTRAINT snapshot_positions_asset_once_uidx UNIQUE (snapshot_id, asset_id)
);

CREATE TABLE analytics.calculation_runs (
    id UUID PRIMARY KEY,
    portfolio_id UUID NOT NULL REFERENCES investment.portfolios (id),
    methodology_version TEXT NOT NULL,
    input_date DATE NOT NULL,
    ledger_watermark TEXT NOT NULL,
    run_status TEXT NOT NULL CHECK (run_status IN ('started', 'completed', 'failed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    safe_error_code TEXT,
    CONSTRAINT calculation_runs_methodology_not_blank CHECK (length(trim(methodology_version)) > 0),
    CONSTRAINT calculation_runs_watermark_not_blank CHECK (length(trim(ledger_watermark)) > 0)
);

CREATE TABLE analytics.inbox_messages (
    id UUID PRIMARY KEY,
    consumer_name TEXT NOT NULL,
    event_id UUID NOT NULL,
    event_version INTEGER NOT NULL CHECK (event_version > 0),
    processing_state TEXT NOT NULL DEFAULT 'pending'
        CHECK (processing_state IN ('pending', 'processing', 'processed', 'retry', 'dead_letter')),
    business_version BIGINT CHECK (business_version IS NULL OR business_version > 0),
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    last_error_code TEXT,
    CONSTRAINT inbox_messages_consumer_event_uidx UNIQUE (consumer_name, event_id),
    CONSTRAINT inbox_messages_consumer_not_blank CHECK (length(trim(consumer_name)) > 0)
);

CREATE INDEX inbox_messages_pending_idx
    ON analytics.inbox_messages (processing_state, received_at);

CREATE TABLE audit.actors (
    id UUID PRIMARY KEY,
    actor_kind TEXT NOT NULL CHECK (actor_kind IN ('user', 'system', 'anonymous')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit.events (
    id UUID PRIMARY KEY,
    actor_id UUID REFERENCES audit.actors (id),
    action_code TEXT NOT NULL,
    target_kind TEXT NOT NULL,
    target_id UUID,
    outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure')),
    request_id UUID,
    trace_id TEXT,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    schema_version INTEGER NOT NULL DEFAULT 1 CHECK (schema_version > 0),
    CONSTRAINT audit_events_action_not_blank CHECK (length(trim(action_code)) > 0),
    CONSTRAINT audit_events_target_not_blank CHECK (length(trim(target_kind)) > 0)
);

CREATE INDEX audit_events_actor_time_idx ON audit.events (actor_id, occurred_at DESC);
CREATE INDEX audit_events_target_time_idx ON audit.events (target_kind, target_id, occurred_at DESC);
CREATE INDEX audit_events_request_idx ON audit.events (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX audit_events_trace_idx ON audit.events (trace_id) WHERE trace_id IS NOT NULL;

COMMIT;
