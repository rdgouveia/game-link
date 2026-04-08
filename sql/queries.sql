-- name: CreateUser :one
INSERT INTO users (username, email, password, fullname)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, password, fullname;

-- name: GetUserByEmail :one
SELECT id, username, email, password, fullname FROM users WHERE email = $1;

-- name: CreateGame :one
INSERT INTO games (name, summary, image_url, igdb_id)
VALUES ($1, $2, $3, $4)
RETURNING id, name, summary, image_url, igdb_id;

-- name: GetGameByIgdbID :one
SELECT id, name, summary, image_url, igdb_id FROM games WHERE igdb_id = $1;

-- name: AddGameToUser :exec
INSERT INTO user_games (user_id, game_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveGameFromUser :exec
DELETE FROM user_games WHERE user_id = $1 AND game_id = $2;

-- name: GetGamesByUser :many
SELECT g.id, g.name, g.summary, g.image_url, g.igdb_id
FROM games g
JOIN user_games ug ON ug.game_id = g.id
WHERE ug.user_id = $1;

