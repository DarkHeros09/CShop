-- name: CreateUserAddress :one
INSERT INTO "user_address" (
  user_id,
  address_id,
  default_address
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: CreateUserAddressWithAddress :one
WITH t1 AS (
  INSERT INTO "address" as a (
    address_line,
    region,
    city
    ) VALUES (
    $1, $2, $3
    )
   RETURNING a.id, a.address_line, a.region, a.city
  ),

  t2 AS ( 
  INSERT INTO "user_address" (
    user_id,
    address_id,
    default_address
    ) VALUES 
    ( $4,
    (SELECT id FROM t1),
      $5
    ) 
  RETURNING *
  )

SELECT 
user_id,
address_id,
default_address,
address_line,
region,
city 
FROM t1, t2; 

-- name: GetUserAddress :one
SELECT * FROM "user_address"
WHERE user_id = $1
AND address_id = $2
LIMIT 1;

-- name: GetUserAddressWithAddress :one
SELECT ua.user_id, ua.address_id, ua.default_address, "address".address_line,  "address".region,  "address".city
FROM "user_address" AS ua 
JOIN "user" ON ua.user_id = "user".id 
JOIN "address" ON ua.address_id = "address".id
WHERE "user".id = sqlc.arg(user_id)
AND "address".id = sqlc.arg(address_id);

-- name: ListUserAddresses :many
SELECT * FROM "user_address"
WHERE user_id = $1
ORDER BY address_id
LIMIT $2
OFFSET $3;

-- name: UpdateUserAddress :one
UPDATE "user_address"
SET 
default_address = $1
WHERE user_id = $2
AND address_id = $3
RETURNING *;

-- -- name: UpdateUserAddressWithAddress :one
-- WITH t1 AS (
--     UPDATE "address" as a
--     SET
--     address_line = COALESCE(sqlc.narg(address_line),address_line), 
--     region = COALESCE(sqlc.narg(region),region), 
--     city= COALESCE(sqlc.narg(city),city)
--     WHERE id = COALESCE(sqlc.arg(id),id)
--     RETURNING a.id, a.address_line, a.region, a.city
--    ),
   
--     t2 AS (
--     UPDATE "user_address"
--     SET
--     default_address = COALESCE(sqlc.narg(default_address),default_address)
--     WHERE
--     user_id = COALESCE(sqlc.arg(user_id),user_id)
--     AND address_id = COALESCE(sqlc.arg(address_id),address_id)
--     RETURNING user_id, address_id, default_address
-- 	)
	
-- SELECT 
-- user_id,
-- address_id,
-- default_address,
-- address_line,
-- region,
-- city From t1,t2;

-- name: DeleteUserAddress :one
DELETE FROM "user_address"
WHERE user_id = $1
AND address_id = $2
RETURNING *;