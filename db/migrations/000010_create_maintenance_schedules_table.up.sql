CREATE TABLE maintenance_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_category_id UUID REFERENCES vehicle_categories (id),
    vehicle_id UUID REFERENCES vehicles (id),
    trigger_type TEXT NOT NULL
        CHECK (trigger_type IN ('date_interval', 'mileage_interval', 'manufacturer_schedule', 'telematics_condition')),
    interval_days INTEGER CHECK (interval_days > 0),
    interval_mileage DECIMAL CHECK (interval_mileage > 0),
    manufacturer_reference TEXT,
    telematics_condition TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_maintenance_schedules_vehicle_category_id ON maintenance_schedules (vehicle_category_id);
CREATE INDEX idx_maintenance_schedules_vehicle_id ON maintenance_schedules (vehicle_id);
CREATE INDEX idx_maintenance_schedules_trigger_type ON maintenance_schedules (trigger_type);
