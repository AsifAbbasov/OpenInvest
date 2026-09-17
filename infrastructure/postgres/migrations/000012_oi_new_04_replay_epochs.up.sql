BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';

-- OI-NEW-04 C' additive replay-state schema.
-- No backfill, activation, grants, ledger rewrite, or snapshot rewrite occurs here.
-- Unique NOT-NULL indexes implement the design-key identities while keeping paired-SQL v1
-- impact classes explicit; FKs/CHECKs remain first-class PostgreSQL constraints.

CREATE TABLE analytics.replay_policy_generations (
    policy_version TEXT NOT NULL,
    activation_generation BIGINT NOT NULL,
    replay_bound_raw_rows INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE UNIQUE INDEX replay_policy_generations_uidx
    ON analytics.replay_policy_generations (policy_version, activation_generation);
ALTER TABLE analytics.replay_policy_generations
    ADD CONSTRAINT replay_policy_generations_generation_positive
    CHECK (activation_generation > 0);
ALTER TABLE analytics.replay_policy_generations
    ADD CONSTRAINT replay_policy_generations_bound_exact
    CHECK (replay_bound_raw_rows = 5000);

CREATE TABLE analytics.replay_policy_events (
    event_id UUID NOT NULL,
    policy_version TEXT NOT NULL,
    activation_generation BIGINT NOT NULL,
    event_type TEXT NOT NULL,
    manifest_sha256 CHAR(64),
    portfolio_count BIGINT,
    reason_code TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE UNIQUE INDEX replay_policy_events_event_uidx
    ON analytics.replay_policy_events (event_id);
CREATE UNIQUE INDEX replay_policy_events_generation_event_uidx
    ON analytics.replay_policy_events (policy_version, activation_generation, event_type);
ALTER TABLE analytics.replay_policy_events
    ADD CONSTRAINT replay_policy_events_generation_fk
    FOREIGN KEY (policy_version, activation_generation)
    REFERENCES analytics.replay_policy_generations (policy_version, activation_generation)
    ON DELETE RESTRICT;
ALTER TABLE analytics.replay_policy_events
    ADD CONSTRAINT replay_policy_events_type_check
    CHECK (event_type IN ('ACTIVATED', 'INVALIDATED'));

CREATE TABLE analytics.portfolio_replay_epochs (
    epoch_id UUID NOT NULL,
    portfolio_id UUID NOT NULL,
    policy_version TEXT NOT NULL,
    activation_generation BIGINT NOT NULL,
    epoch_generation BIGINT NOT NULL,
    boundary_trade_date DATE,
    boundary_logical_sequence BIGINT NOT NULL,
    build_raw_ledger_watermark BIGINT NOT NULL,
    position_methodology_version TEXT NOT NULL,
    snapshot_methodology_version TEXT NOT NULL,
    latest_position_trade_date DATE,
    source_state_sha256 CHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE UNIQUE INDEX portfolio_replay_epochs_epoch_uidx
    ON analytics.portfolio_replay_epochs (epoch_id);
CREATE UNIQUE INDEX portfolio_replay_epochs_generation_uidx
    ON analytics.portfolio_replay_epochs (portfolio_id, epoch_generation);
CREATE INDEX portfolio_replay_epochs_current_idx
    ON analytics.portfolio_replay_epochs (
        portfolio_id,
        policy_version,
        activation_generation,
        epoch_generation DESC
    );
ALTER TABLE analytics.portfolio_replay_epochs
    ADD CONSTRAINT portfolio_replay_epochs_portfolio_fk
    FOREIGN KEY (portfolio_id)
    REFERENCES investment.portfolios (id)
    ON DELETE CASCADE;
ALTER TABLE analytics.portfolio_replay_epochs
    ADD CONSTRAINT portfolio_replay_epochs_policy_fk
    FOREIGN KEY (policy_version, activation_generation)
    REFERENCES analytics.replay_policy_generations (policy_version, activation_generation)
    ON DELETE RESTRICT;
ALTER TABLE analytics.portfolio_replay_epochs
    ADD CONSTRAINT portfolio_replay_epochs_generation_positive
    CHECK (epoch_generation > 0);
ALTER TABLE analytics.portfolio_replay_epochs
    ADD CONSTRAINT portfolio_replay_epochs_boundary_nonnegative
    CHECK (boundary_logical_sequence >= 0);
ALTER TABLE analytics.portfolio_replay_epochs
    ADD CONSTRAINT portfolio_replay_epochs_watermark_nonnegative
    CHECK (build_raw_ledger_watermark >= 0);

CREATE TABLE analytics.portfolio_replay_positions (
    epoch_id UUID NOT NULL,
    asset_id UUID NOT NULL,
    asset_type TEXT NOT NULL,
    quantity NUMERIC(28, 8) NOT NULL,
    weighted_average_cost_amount NUMERIC(28, 8) NOT NULL,
    position_generation BIGINT NOT NULL
);
CREATE UNIQUE INDEX portfolio_replay_positions_uidx
    ON analytics.portfolio_replay_positions (epoch_id, asset_id);
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_epoch_fk
    FOREIGN KEY (epoch_id)
    REFERENCES analytics.portfolio_replay_epochs (epoch_id)
    ON DELETE CASCADE;
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_asset_fk
    FOREIGN KEY (asset_id)
    REFERENCES investment.assets (id)
    ON DELETE RESTRICT;
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_asset_type_check
    CHECK (asset_type IN ('stock', 'bond'));
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_quantity_positive
    CHECK (quantity > 0);
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_wac_positive
    CHECK (weighted_average_cost_amount > 0);
ALTER TABLE analytics.portfolio_replay_positions
    ADD CONSTRAINT portfolio_replay_positions_generation_positive
    CHECK (position_generation > 0);

CREATE TABLE analytics.portfolio_replay_financial_state (
    epoch_id UUID NOT NULL,
    deposits_amount NUMERIC(28, 8) NOT NULL,
    withdrawals_amount NUMERIC(28, 8) NOT NULL,
    buy_outflows_amount NUMERIC(28, 8) NOT NULL,
    sell_inflows_amount NUMERIC(28, 8) NOT NULL,
    dividends_gross_amount NUMERIC(28, 8) NOT NULL,
    coupons_gross_amount NUMERIC(28, 8) NOT NULL,
    fees_amount NUMERIC(28, 8) NOT NULL,
    taxes_amount NUMERIC(28, 8) NOT NULL,
    net_investment_income_amount NUMERIC(28, 8) NOT NULL,
    invested_capital_amount NUMERIC(28, 8) NOT NULL,
    input_watermark_at TIMESTAMP WITH TIME ZONE,
    input_watermark TEXT NOT NULL
);
CREATE UNIQUE INDEX portfolio_replay_financial_state_uidx
    ON analytics.portfolio_replay_financial_state (epoch_id);
ALTER TABLE analytics.portfolio_replay_financial_state
    ADD CONSTRAINT portfolio_replay_financial_state_epoch_fk
    FOREIGN KEY (epoch_id)
    REFERENCES analytics.portfolio_replay_epochs (epoch_id)
    ON DELETE CASCADE;

COMMIT;
