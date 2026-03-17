CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    email_hash  TEXT        NOT NULL UNIQUE,
    settings    JSONB       NOT NULL DEFAULT '{"currency":"USD","timezone":"UTC"}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email_hash ON users(email_hash);
