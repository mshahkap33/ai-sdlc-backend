CREATE TABLE vehicle_telematics_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    reading_type TEXT NOT NULL CHECK (reading_type IN ('mileage', 'diagnostic_code', 'condition_alert')),
    mileage_value DECIMAL CHECK (mileage_value >= 0),
    condition_code TEXT,
    reported_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_vehicle_telematics_readings_vehicle_id ON vehicle_telematics_readings (vehicle_id);
CREATE INDEX idx_vehicle_telematics_readings_reading_type ON vehicle_telematics_readings (reading_type);
CREATE INDEX idx_vehicle_telematics_readings_reported_at ON vehicle_telematics_readings (reported_at);
