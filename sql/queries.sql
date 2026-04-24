-- name: InsertMessage :one
INSERT INTO messages (sender_id, receiver_id, content)
VALUES ($1, $2, $3)
RETURNING id, sender_id, receiver_id, content, read_at, created_at;

-- name: GetConversation :many
SELECT id, sender_id, receiver_id, content, read_at, created_at
FROM messages
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: MarkMessagesAsRead :exec
UPDATE messages SET read_at = NOW()
WHERE receiver_id = $1 AND sender_id = $2 AND read_at IS NULL;

-- name: GetConversations :many
WITH last_messages AS (
    SELECT
        CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS other_id,
        id, sender_id, receiver_id, content, read_at, created_at,
        ROW_NUMBER() OVER (
            PARTITION BY CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END
            ORDER BY created_at DESC
        ) AS rn
    FROM messages
    WHERE sender_id = $1 OR receiver_id = $1
)
SELECT
    lm.other_id,
    lm.id,
    lm.sender_id,
    lm.receiver_id,
    lm.content,
    lm.read_at,
    lm.created_at,
    u.username,
    u.fullname,
    u.avatar_url,
    (SELECT COUNT(*)::int FROM messages m2
     WHERE m2.receiver_id = $1 AND m2.sender_id = lm.other_id AND m2.read_at IS NULL) AS unread_count
FROM last_messages lm
JOIN users u ON u.id = lm.other_id
WHERE lm.rn = 1
ORDER BY lm.created_at DESC;

-- name: CreateUser :one
INSERT INTO users (username, email, password, fullname)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, password, fullname, biography, verified;

-- name: GetUserByEmail :one
SELECT id, username, email, password, fullname, biography, verified FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, username, email, fullname, biography, verified, avatar_url FROM users WHERE id = $1;

-- name: UpdateUserProfile :exec
UPDATE users SET fullname = $1, biography = $2 WHERE id = $3;

-- name: UpdateUserAvatar :exec
UPDATE users SET avatar_url = $1 WHERE id = $2;

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

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: SearchUsersByFullname :many
SELECT u.id, u.username, u.fullname, u.biography, u.verified, u.avatar_url,
       ul.status         AS link_status,
       ul.requester_id   AS link_requester_id
FROM users u
LEFT JOIN user_links ul ON (
    (ul.requester_id = $2 AND ul.recipient_id = u.id) OR
    (ul.recipient_id = $2 AND ul.requester_id = u.id)
)
WHERE u.fullname ILIKE '%' || $1 || '%'
  AND u.id != $2
LIMIT 20;

-- name: GetUserByUsername :one
SELECT u.id, u.username, u.fullname, u.biography, u.verified, u.avatar_url,
       ul.status         AS link_status,
       ul.requester_id   AS link_requester_id
FROM users u
LEFT JOIN user_links ul ON (
    (ul.requester_id = $2 AND ul.recipient_id = u.id) OR
    (ul.recipient_id = $2 AND ul.requester_id = u.id)
)
WHERE u.username = $1;

-- name: GetMatchCandidates :many
SELECT
    u.id,
    u.username,
    u.fullname,
    u.biography,
    u.verified,
    u.avatar_url,
    COUNT(ug2.game_id)::int                                          AS common_games,
    (SELECT COUNT(*)::int FROM user_games ug3 WHERE ug3.user_id = u.id) AS total_games
FROM users u
JOIN user_games ug2 ON ug2.user_id = u.id
    AND ug2.game_id IN (SELECT ug1.game_id FROM user_games ug1 WHERE ug1.user_id = $1)
WHERE u.id != $1
  AND NOT EXISTS (
      SELECT 1 FROM user_links ul
      WHERE (ul.requester_id = $1 AND ul.recipient_id = u.id)
         OR (ul.recipient_id = $1 AND ul.requester_id = u.id)
  )
GROUP BY u.id
HAVING COUNT(ug2.game_id) > 0
ORDER BY common_games DESC
LIMIT 300;

-- name: GetUserAvailability :many
SELECT day_of_week, start_time, end_time
FROM user_availability
WHERE user_id = $1
ORDER BY day_of_week, start_time;

-- name: GetAvailabilityForUsers :many
SELECT user_id, day_of_week, start_time, end_time
FROM user_availability
WHERE user_id = ANY($1::uuid[])
ORDER BY user_id, day_of_week, start_time;

-- name: DeleteUserAvailability :exec
DELETE FROM user_availability WHERE user_id = $1;

-- name: InsertAvailabilitySlot :exec
INSERT INTO user_availability (user_id, day_of_week, start_time, end_time)
VALUES ($1, $2, $3, $4);

-- name: CreateLink :exec
INSERT INTO user_links (requester_id, recipient_id)
VALUES ($1, $2);

-- name: AcceptLink :exec
UPDATE user_links SET status = 'accepted'
WHERE requester_id = $1 AND recipient_id = $2;

-- name: DeleteLink :exec
DELETE FROM user_links
WHERE (requester_id = $1 AND recipient_id = $2)
   OR (requester_id = $2 AND recipient_id = $1);

-- name: GetLinkBetweenUsers :one
SELECT status FROM user_links
WHERE (requester_id = $1 AND recipient_id = $2)
   OR (requester_id = $2 AND recipient_id = $1)
LIMIT 1;

-- name: GetPendingLinksReceived :many
SELECT ul.requester_id, ul.created_at, u.username, u.fullname, u.verified, u.avatar_url
FROM user_links ul
JOIN users u ON u.id = ul.requester_id
WHERE ul.recipient_id = $1 AND ul.status = 'pending';

-- name: GetAcceptedLinks :many
SELECT
    ul.requester_id,
    ur.username  AS requester_username,
    ur.fullname  AS requester_fullname,
    ur.verified  AS requester_verified,
    ur.avatar_url AS requester_avatar_url,
    ul.recipient_id,
    rc.username  AS recipient_username,
    rc.fullname  AS recipient_fullname,
    rc.verified  AS recipient_verified,
    rc.avatar_url AS recipient_avatar_url
FROM user_links ul
JOIN users ur ON ur.id = ul.requester_id
JOIN users rc ON rc.id = ul.recipient_id
WHERE (ul.requester_id = $1 OR ul.recipient_id = $1)
  AND ul.status = 'accepted';

-- name: GetGamesByUser :many
SELECT g.id, g.name, g.summary, g.image_url, g.igdb_id
FROM games g
JOIN user_games ug ON ug.game_id = g.id
WHERE ug.user_id = $1;
