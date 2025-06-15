-- name: CreateAddress :one
INSERT INTO "address" (
  name,
  user_id,
  telephone,
  address_line,
  region,
  city
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetAddress :one
SELECT * FROM "address"
WHERE id = $1 
LIMIT 1;

-- name: GetAddressByCity :one
SELECT * FROM "address"
WHERE city = $1 
LIMIT 1;

-- name: ListAddressesByID :many
SELECT * FROM "address"
WHERE id = ANY(sqlc.arg(addresses_ids)::bigint[]);

-- name: ListAddressesByCity :many
SELECT * FROM "address"
WHERE city = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateAddress :one
UPDATE "address"
SET 
name = COALESCE(sqlc.narg(name),name),
telephone = COALESCE(sqlc.narg(telephone),telephone),
address_line = COALESCE(sqlc.narg(address_line),address_line),
region = COALESCE(sqlc.narg(region),region),
city = COALESCE(sqlc.narg(city),city),
updated_at = now()
WHERE id = sqlc.arg(id)
AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteAddress :exec
DELETE FROM "address"
WHERE id = $1;

-- name: GetUserAddress :one
SELECT * FROM "address" AS ad
JOIN "user" AS u ON u.id = ad.user_id
WHERE user_id = $1
AND ad.id = $2
LIMIT 1;

-- name: ListAddressesByUserID :many
SELECT u.default_address_id, ad.* FROM "address" AS ad
JOIN "user" AS u ON u.id = ad.user_id
WHERE u.id = $1
ORDER BY ad.id;

-- name: DeleteUserAddress :one
DELETE FROM "address" AS ad
WHERE user_id = $1
AND ad.id = $2
RETURNING *;
