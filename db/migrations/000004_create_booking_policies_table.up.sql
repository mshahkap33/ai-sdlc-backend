CREATE TABLE booking_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_category_id UUID NOT NULL REFERENCES vehicle_categories (id),
    booking_mode TEXT NOT NULL CHECK (booking_mode IN ('vehicle_level', 'category_level', 'both')),
    effective_from TIMESTAMP WITH TIME ZONE NOT NULL,
    effective_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_booking_policies_vehicle_category_id ON booking_policies (vehicle_category_id);
CREATE INDEX idx_booking_policies_booking_mode ON booking_policies (booking_mode);
