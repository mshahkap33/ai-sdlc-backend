-- Enable pgcrypto so gen_random_uuid() is available for UUID primary keys.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
