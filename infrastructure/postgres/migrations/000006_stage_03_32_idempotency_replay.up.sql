BEGIN;

ALTER TABLE investment.command_deduplication
    ADD COLUMN response_version SMALLINT,
    ADD COLUMN response_status SMALLINT,
    ADD COLUMN response_body BYTEA,
    ADD COLUMN response_request_id UUID,
    ADD COLUMN response_trace_id TEXT;

-- Pre-Stage-3.32 rows used response_hash=request_hash as a placeholder rather than a hash of
-- the observable response. Clear only that known placeholder so new rows can distinguish a real
-- replay artifact from legacy completed commands that cannot be reconstructed exactly.
UPDATE investment.command_deduplication
SET response_hash = NULL
WHERE terminal_status = 'success'
  AND response_hash = request_hash;

ALTER TABLE investment.command_deduplication
    ADD CONSTRAINT command_deduplication_response_artifact_shape CHECK (
        (
            response_version IS NULL
            AND response_status IS NULL
            AND response_body IS NULL
            AND response_request_id IS NULL
            AND response_trace_id IS NULL
        )
        OR (
            response_version = 1
            AND response_status BETWEEN 100 AND 599
            AND response_body IS NOT NULL
            AND octet_length(response_body) BETWEEN 1 AND 262144
            AND response_request_id IS NOT NULL
            AND response_trace_id IS NOT NULL
            AND length(response_trace_id) BETWEEN 1 AND 128
            AND response_hash IS NOT NULL
            AND length(response_hash) = 64
        )
    );

COMMIT;
