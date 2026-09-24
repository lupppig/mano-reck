CREATE TABLE platform.outbox_events (
    id uuid PRIMARY KEY,
    tenant_id uuid,
    event_type text NOT NULL,
    event_version integer NOT NULL CHECK (event_version > 0),
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    subject text NOT NULL,
    occurred_at timestamptz NOT NULL,
    correlation_id uuid NOT NULL,
    causation_id uuid NOT NULL,
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'publishing', 'published', 'dead_letter')),
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_owner uuid,
    lease_expires_at timestamptz,
    published_at timestamptz,
    dead_lettered_at timestamptz,
    last_error_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT outbox_event_type_format
        CHECK (event_type ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$'),
    CONSTRAINT outbox_aggregate_type_format
        CHECK (aggregate_type ~ '^[a-z][a-z0-9_]*$'),
    CONSTRAINT outbox_subject_format
        CHECK (subject ~ '^manoreck\.events\.[a-z][a-z0-9_.]*\.v[1-9][0-9]*$'),
    CONSTRAINT outbox_payload_identity
        CHECK (
            payload ->> 'id' = id::text
            AND payload ->> 'type' = event_type
            AND (payload ->> 'version')::integer = event_version
        ),
    CONSTRAINT outbox_lease_state
        CHECK (
            (status = 'publishing' AND lease_owner IS NOT NULL AND lease_expires_at IS NOT NULL)
            OR (status <> 'publishing' AND lease_owner IS NULL AND lease_expires_at IS NULL)
        ),
    CONSTRAINT outbox_terminal_state
        CHECK (
            (status = 'published' AND published_at IS NOT NULL AND dead_lettered_at IS NULL)
            OR (status = 'dead_letter' AND dead_lettered_at IS NOT NULL AND published_at IS NULL)
            OR (status IN ('pending', 'publishing') AND published_at IS NULL AND dead_lettered_at IS NULL)
        )
);

CREATE INDEX outbox_events_pending_idx
    ON platform.outbox_events (next_attempt_at, created_at)
    WHERE status = 'pending';

CREATE INDEX outbox_events_expired_lease_idx
    ON platform.outbox_events (lease_expires_at, created_at)
    WHERE status = 'publishing';

CREATE INDEX outbox_events_tenant_created_idx
    ON platform.outbox_events (tenant_id, created_at DESC)
    WHERE tenant_id IS NOT NULL;

CREATE TABLE platform.event_consumptions (
    consumer_name text NOT NULL,
    event_id uuid NOT NULL,
    event_type text NOT NULL,
    delivery_count integer NOT NULL CHECK (delivery_count > 0),
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer_name, event_id),
    CONSTRAINT event_consumption_consumer_name
        CHECK (consumer_name ~ '^[a-z][a-z0-9_-]*-v[1-9][0-9]*$'),
    CONSTRAINT event_consumption_type_format
        CHECK (event_type ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$')
);

COMMENT ON TABLE platform.outbox_events IS
    'Committed integration events awaiting at-least-once JetStream publication';
COMMENT ON TABLE platform.event_consumptions IS
    'Durable per-consumer event idempotency records';
