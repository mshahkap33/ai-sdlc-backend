CREATE TABLE vehicle_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    previous_status TEXT
        CHECK (previous_status IN ('available', 'reserved', 'rented', 'cleaning', 'maintenance', 'damaged', 'retired')),
    new_status TEXT NOT NULL
        CHECK (new_status IN ('available', 'reserved', 'rented', 'cleaning', 'maintenance', 'damaged', 'retired')),
    trigger_event TEXT NOT NULL,
    reason TEXT,
    effective_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_vehicle_status_history_vehicle_id ON vehicle_status_history (vehicle_id);
CREATE INDEX idx_vehicle_status_history_new_status ON vehicle_status_history (new_status);
CREATE INDEX idx_vehicle_status_history_trigger_event ON vehicle_status_history (trigger_event);
CREATE INDEX idx_vehicle_status_history_effective_at ON vehicle_status_history (effective_at);
