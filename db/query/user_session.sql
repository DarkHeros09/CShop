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

-- name: UpdateUserSession :one
UPDATE "user_session"
SET 
is_blocked = COALESCE(sqlc.narg(is_blocked),is_blocked),
updated_at = now()
WHERE id = sqlc.arg(id)
AND user_id = sqlc.arg(user_id)
AND refresh_token = sqlc.arg(refresh_token)
RETURNING *;