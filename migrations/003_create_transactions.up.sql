CREATE TYPE tx_status AS ENUM ('pending', 'processed', 'failed');

CREATE TABLE transactions (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id          TEXT        NOT NULL,
    account_id           UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount               BIGINT      NOT NULL,
    mcc                  SMALLINT,
    original_description TEXT        NOT NULL DEFAULT '',
    clean_category       TEXT,
    status               tx_status   NOT NULL DEFAULT 'pending',
    occurred_at          TIMESTAMPTZ NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Idempotency key: one external_id per account
CREATE UNIQUE INDEX idx_tx_external_id_account ON transactions(account_id, external_id);
CREATE INDEX idx_tx_account_id  ON transactions(account_id);
CREATE INDEX idx_tx_status      ON transactions(status);
CREATE INDEX idx_tx_occurred_at ON transactions(occurred_at DESC);
