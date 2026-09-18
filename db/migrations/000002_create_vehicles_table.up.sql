CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY,
    vin TEXT NOT NULL,
    plate_number TEXT NOT NULL,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER NOT NULL,
    category_id UUID NOT NULL REFERENCES vehicle_categories (id),
    mileage DECIMAL NOT NULL CHECK (mileage >= 0),
    fuel_level DECIMAL NOT NULL CHECK (fuel_level BETWEEN 0 AND 100),
    location TEXT NOT NULL,
    condition_notes TEXT,
    status TEXT NOT NULL DEFAULT 'available'
        CHECK (status IN ('available', 'reserved', 'rented', 'cleaning', 'maintenance', 'damaged', 'retired')),
    status_since TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOL NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vehicles_vin ON vehicles (vin);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vehicles_plate_number ON vehicles (plate_number);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles (status);
CREATE INDEX IF NOT EXISTS idx_vehicles_category_id ON vehicles (category_id);
