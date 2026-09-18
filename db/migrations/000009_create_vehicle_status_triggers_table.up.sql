CREATE TABLE vehicle_status_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    resulting_status TEXT NOT NULL
        CHECK (resulting_status IN ('available', 'reserved', 'rented', 'cleaning', 'maintenance', 'damaged', 'retired')),
    requires_reason BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT uq_vehicle_status_triggers_event_type UNIQUE (event_type)
);

CREATE INDEX idx_vehicle_status_triggers_event_type ON vehicle_status_triggers (event_type);
