-- +goose Up
ALTER TABLE users ADD COLUMN fullname TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    summary TEXT NOT NULL,
    image_url TEXT NOT NULL,
    igdb_id INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_games (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, game_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_games;
DROP TABLE IF EXISTS games;
ALTER TABLE users DROP COLUMN fullname;
