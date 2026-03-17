CREATE TABLE game_events (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID        NOT NULL REFERENCES transactions(id),
    user_id        UUID        NOT NULL REFERENCES users(id),
    attribute      TEXT        NOT NULL,
    delta          INT         NOT NULL,
    reason         TEXT        NOT NULL DEFAULT '',
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_game_events_user_id ON game_events(user_id);
CREATE INDEX idx_game_events_tx_id   ON game_events(transaction_id);
