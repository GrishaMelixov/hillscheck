CREATE TYPE account_type AS ENUM ('debit', 'credit', 'cash');

CREATE TABLE accounts (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT          NOT NULL,
    type        account_type  NOT NULL,
    balance     BIGINT        NOT NULL DEFAULT 0,
    currency    CHAR(3)       NOT NULL DEFAULT 'USD',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
