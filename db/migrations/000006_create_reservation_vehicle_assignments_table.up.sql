CREATE TABLE reservation_vehicle_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL REFERENCES reservations (id),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL,
    assignment_type TEXT NOT NULL CHECK (assignment_type IN ('customer_selected', 'system_assigned_at_fulfillment')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT uq_reservation_vehicle_assignments_reservation_id UNIQUE (reservation_id)
);

CREATE INDEX idx_reservation_vehicle_assignments_vehicle_id ON reservation_vehicle_assignments (vehicle_id);
