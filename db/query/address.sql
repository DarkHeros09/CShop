-- name: CreateAddress :one
INSERT INTO "address" (
  name,
  telephone,
  address_line,
  region,
  city
) VALUES (
  $1, $2, $3, $4, $5
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
RETURNING *;

-- name: DeleteAddress :exec
DELETE FROM "address"
WHERE id = $1;