CREATE TABLE maintenance_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    maintenance_work_order_id UUID REFERENCES maintenance_work_orders (id),
    compliance_document_id UUID REFERENCES vehicle_compliance_documents (id),
    alert_type TEXT NOT NULL
        CHECK (alert_type IN ('overdue_service', 'expiring_document', 'recall', 'safety_defect')),
    severity TEXT NOT NULL CHECK (severity IN ('normal', 'urgent')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved')),
    raised_at TIMESTAMP WITH TIME ZONE NOT NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_maintenance_alerts_vehicle_id ON maintenance_alerts (vehicle_id);
CREATE INDEX idx_maintenance_alerts_maintenance_work_order_id ON maintenance_alerts (maintenance_work_order_id);
CREATE INDEX idx_maintenance_alerts_compliance_document_id ON maintenance_alerts (compliance_document_id);
CREATE INDEX idx_maintenance_alerts_alert_type ON maintenance_alerts (alert_type);
CREATE INDEX idx_maintenance_alerts_severity ON maintenance_alerts (severity);
CREATE INDEX idx_maintenance_alerts_status ON maintenance_alerts (status);
CREATE INDEX idx_maintenance_alerts_raised_at ON maintenance_alerts (raised_at);
