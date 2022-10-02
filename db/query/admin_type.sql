-- name: CreateAdminType :one
INSERT INTO "admin_type" (
  admin_type
) VALUES ( $1 )
RETURNING *;

-- name: GetAdminType :one
SELECT * FROM "admin_type"
WHERE id = $1 LIMIT 1;

-- name: ListAdminTypes :many
SELECT * FROM "admin_type"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAdminType :one
UPDATE "admin_type"
SET admin_type = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAdminTypeByID :exec
DELETE FROM "admin_type"
WHERE id = $1;

-- name: DeleteAdminTypeByType :exec
DELETE FROM "admin_type"
WHERE admin_type = $1;