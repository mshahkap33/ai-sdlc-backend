CREATE TABLE maintenance_work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    maintenance_schedule_id UUID REFERENCES maintenance_schedules (id),
    trigger_source TEXT NOT NULL
        CHECK (trigger_source IN ('date_interval', 'mileage_interval', 'manufacturer_schedule', 'telematics_condition', 'recall', 'safety_defect')),
    priority TEXT NOT NULL CHECK (priority IN ('routine', 'urgent')),
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'scheduled', 'in_progress', 'completed', 'cancelled')),
    due_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_maintenance_work_orders_vehicle_id ON maintenance_work_orders (vehicle_id);
CREATE INDEX idx_maintenance_work_orders_maintenance_schedule_id ON maintenance_work_orders (maintenance_schedule_id);
CREATE INDEX idx_maintenance_work_orders_trigger_source ON maintenance_work_orders (trigger_source);
CREATE INDEX idx_maintenance_work_orders_priority ON maintenance_work_orders (priority);
CREATE INDEX idx_maintenance_work_orders_status ON maintenance_work_orders (status);
CREATE INDEX idx_maintenance_work_orders_due_at ON maintenance_work_orders (due_at);
