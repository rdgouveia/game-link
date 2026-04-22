CREATE TABLE IF NOT EXISTS users (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username  TEXT NOT NULL,
    email     TEXT NOT NULL UNIQUE,
    password  TEXT NOT NULL,
    fullname  TEXT NOT NULL,
    biography  TEXT,
    verified   BOOLEAN NOT NULL DEFAULT false,
    avatar_url TEXT
);

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

CREATE TABLE IF NOT EXISTS user_availability (
    user_id     UUID     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time  TIME     NOT NULL,
    end_time    TIME     NOT NULL,
    PRIMARY KEY (user_id, day_of_week, start_time),
    CHECK (end_time > start_time)
);

CREATE TABLE IF NOT EXISTS user_links (
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (requester_id, recipient_id),
    CHECK (requester_id <> recipient_id)
);
