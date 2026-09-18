CREATE TABLE IF NOT EXISTS vehicle_categories (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    active BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOL NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vehicle_categories_name ON vehicle_categories (name);
