CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS peers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name text NOT NULL,
    public_key text NOT NULL,
    ip_address inet NOT NULL,
    allowed_ips inet[] DEFAULT '{}',
    endpoint text,
    persistent_keepalive smallint,
    last_handshake timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW())
