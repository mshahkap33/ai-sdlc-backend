CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_category_id UUID NOT NULL REFERENCES vehicle_categories (id),
    requested_vehicle_id UUID REFERENCES vehicles (id),
    booking_mode TEXT NOT NULL CHECK (booking_mode IN ('vehicle_level', 'category_level')),
    start_datetime TIMESTAMP WITH TIME ZONE NOT NULL,
    end_datetime TIMESTAMP WITH TIME ZONE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled', 'fulfilled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_reservations_vehicle_category_id ON reservations (vehicle_category_id);
CREATE INDEX idx_reservations_requested_vehicle_id ON reservations (requested_vehicle_id);
CREATE INDEX idx_reservations_start_datetime ON reservations (start_datetime);
CREATE INDEX idx_reservations_end_datetime ON reservations (end_datetime);
CREATE INDEX idx_reservations_status ON reservations (status);
