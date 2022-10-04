-- name: CreateUserAddress :one
INSERT INTO "user_address" (
  user_id,
  address_id,
  is_default
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserAddress :one
SELECT * FROM "user_address"
WHERE user_id = $1
And address_id = $2
LIMIT 1;

-- name: ListUserAddresses :many
SELECT * FROM "user_address"
WHERE user_id = $1
ORDER BY address_id
LIMIT $2
OFFSET $3;

-- name: UpdateUserAddress :one
UPDATE "user_address"
SET 
is_default = $1
WHERE user_id = $2
And address_id = $3
RETURNING *;

-- name: DeleteUserAddress :exec
DELETE FROM "user_address"
WHERE user_id = $1
And address_id = $2;