-- name: CreateUserSession :one
INSERT INTO "user_session" (
  id,
  user_id,
  refresh_token,
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM "user_session"
WHERE id = $1 LIMIT 1;