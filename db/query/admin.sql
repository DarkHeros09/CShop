-- name: CreateAdmin :one
INSERT INTO "admin" (
  username,
  email,
  password,
  type_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetAdmin :one
SELECT * FROM "admin"
WHERE id = $1 LIMIT 1;

-- name: GetAdminByEmail :one
SELECT * FROM "admin"
WHERE email = $1 LIMIT 1;

-- name: ListAdmins :many
SELECT * FROM "admin"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAdmin :one
UPDATE "admin"
SET active = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAdmin :exec
DELETE FROM "admin"
WHERE id = $1;