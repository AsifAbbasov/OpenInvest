BEGIN;

ALTER TABLE investment.command_deduplication
    DROP CONSTRAINT IF EXISTS command_deduplication_response_artifact_shape;

ALTER TABLE investment.command_deduplication
    DROP COLUMN IF EXISTS response_trace_id,
    DROP COLUMN IF EXISTS response_request_id,
    DROP COLUMN IF EXISTS response_body,
    DROP COLUMN IF EXISTS response_status,
    DROP COLUMN IF EXISTS response_version;

-- Restore the pre-Stage-3.32 placeholder behavior for full rollback compatibility.
UPDATE investment.command_deduplication
SET response_hash = request_hash
WHERE terminal_status = 'success';

COMMIT;
