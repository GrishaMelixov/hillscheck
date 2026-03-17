CREATE TABLE game_profiles (
    user_id    UUID    PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    level      INT     NOT NULL DEFAULT 1,
    xp         BIGINT  NOT NULL DEFAULT 0,
    hp         INT     NOT NULL DEFAULT 100 CHECK (hp >= 0 AND hp <= 1000),
    mana       INT     NOT NULL DEFAULT 50  CHECK (mana >= 0 AND mana <= 500),
    strength   INT     NOT NULL DEFAULT 10,
    intellect  INT     NOT NULL DEFAULT 10,
    luck       INT     NOT NULL DEFAULT 10,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
