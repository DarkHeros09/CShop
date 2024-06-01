-- name: CreateAdminSession :one
INSERT INTO "admin_session" (
  id,
  admin_id,
  refresh_token,
  admin_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetAdminSession :one
SELECT * FROM "admin_session"
WHERE id = $1 LIMIT 1;

-- name: UpdateAdminSession :one
UPDATE "admin_session"
SET 
is_blocked = COALESCE(sqlc.narg(is_blocked),is_blocked),
updated_at = now()
WHERE id = sqlc.arg(id)
AND admin_id = sqlc.arg(admin_id)
AND refresh_token = sqlc.arg(refresh_token)
RETURNING *;