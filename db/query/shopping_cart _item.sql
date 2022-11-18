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

-- name: GetShoppingCartItemByUserIDCartID :one
SELECT sci.*, sc.user_id
FROM "shopping_cart_item" AS sci
LEFT JOIN "shopping_cart" AS sc ON sc.id = sci.shopping_cart_id
WHERE sc.user_id = $1
AND sci.shopping_cart_id = $2
LIMIT 1;

-- name: ListShoppingCartItems :many
SELECT * FROM "shopping_cart_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListShoppingCartItemsByCartID :many
SELECT * FROM "shopping_cart_item"
WHERE shopping_cart_id = $1
ORDER BY id;

-- name: ListShoppingCartItemsByUserID :many
SELECT sc.user_id, sci.*
FROM "shopping_cart" AS sc
LEFT JOIN "shopping_cart_item" AS sci ON sci.shopping_cart_id = sc.id
WHERE sc.user_id = $1;

-- name: UpdateShoppingCartItem :one
WITH t1 AS (
  SELECT user_id FROM "shopping_cart" AS sc
  WHERE sc.id = sqlc.arg(shopping_cart_id)
)

UPDATE "shopping_cart_item" AS sci
SET 
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id),
qty = COALESCE(sqlc.narg(qty),qty)
WHERE sci.id = sqlc.arg(id)
RETURNING *, (SELECT user_id FROM t1);

-- name: DeleteShoppingCartItem :exec
WITH t1 AS (
  SELECT id FROM "shopping_cart" AS sc
  WHERE sc.user_id = sqlc.arg(user_id)
)
DELETE FROM "shopping_cart_item" AS sci
WHERE sci.id = sqlc.arg(id)
AND sci.shopping_cart_id = (SELECT id FROM t1); 

-- name: DeleteShoppingCartItemAllByUser :many
WITH t1 AS(
  SELECT id FROM "shopping_cart" WHERE user_id = $1
)
DELETE FROM "shopping_cart_item"
WHERE shopping_cart_id = (SELECT id FROM t1)
RETURNING *;