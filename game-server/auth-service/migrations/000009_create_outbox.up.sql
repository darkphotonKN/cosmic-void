CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Core routing and event metadata
    routing_key VARCHAR(255) NOT NULL,
    exchange VARCHAR(255) NOT NULL,
    payload BYTEA NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- State of published
    -- null = pending
    -- not null (existing timestamp) = published
    -- whole row not present = no issues
    published_at TIMESTAMPTZ NULL
);

-- Partial index for the worker's hot path: scanning pending rows in FIFO order.
CREATE INDEX idx_outbox_pending ON outbox (created_at)
    WHERE published_at IS NULL;

-- Partial index for periodic cleanup of already-published rows.
CREATE INDEX idx_outbox_cleanup ON outbox (published_at)
    WHERE published_at IS NOT NULL;
