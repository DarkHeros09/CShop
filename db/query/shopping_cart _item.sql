-- name: CreateShoppingCartItem :one
INSERT INTO "shopping_cart_item" (
  shopping_cart_id,
  product_item_id,
  qty
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetShoppingCartItem :one
SELECT * FROM "shopping_cart_item"
WHERE id = $1 LIMIT 1;

-- name: ListShoppingCartItems :many
SELECT * FROM "shopping_cart_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListShoppingCartItemsByCartID :many
SELECT * FROM "shopping_cart_item"
WHERE shopping_cart_id = $1
ORDER BY id;

-- name: UpdateShoppingCartItem :one
UPDATE "shopping_cart_item"
SET 
shopping_cart_id = COALESCE(sqlc.narg(shopping_cart_id),shopping_cart_id),
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id),
qty = COALESCE(sqlc.narg(qty),qty)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShoppingCartItem :exec
DELETE FROM "shopping_cart_item"
WHERE id = $1;