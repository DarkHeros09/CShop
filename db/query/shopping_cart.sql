-- name: CreateShoppingCart :one
INSERT INTO "shopping_cart" (
  user_id
) VALUES (
  $1
)
RETURNING *;

-- name: GetShoppingCart :one
SELECT * FROM "shopping_cart"
WHERE id = $1 LIMIT 1;

-- name: ListShoppingCarts :many
SELECT * FROM "shopping_cart"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateShoppingCart :one
UPDATE "shopping_cart"
SET 
user_id = COALESCE(sqlc.narg(user_id),user_id)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShoppingCart :exec
DELETE FROM "shopping_cart"
WHERE id = $1;