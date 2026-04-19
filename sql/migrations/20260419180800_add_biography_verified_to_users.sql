-- +goose Up
ALTER TABLE users ADD COLUMN biography TEXT;
ALTER TABLE users ADD COLUMN verified BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN biography;
ALTER TABLE users DROP COLUMN verified;
