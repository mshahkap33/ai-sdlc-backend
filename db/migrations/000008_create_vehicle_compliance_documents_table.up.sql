CREATE TABLE vehicle_compliance_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id),
    document_type TEXT NOT NULL CHECK (document_type IN ('registration', 'insurance', 'inspection', 'recall')),
    status TEXT NOT NULL CHECK (status IN ('valid', 'expired', 'open', 'resolved')),
    expires_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    reference_number TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_vehicle_compliance_documents_vehicle_id ON vehicle_compliance_documents (vehicle_id);
CREATE INDEX idx_vehicle_compliance_documents_document_type ON vehicle_compliance_documents (document_type);
CREATE INDEX idx_vehicle_compliance_documents_status ON vehicle_compliance_documents (status);
CREATE INDEX idx_vehicle_compliance_documents_expires_at ON vehicle_compliance_documents (expires_at);
