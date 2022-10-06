-- name: CreateProductItem :one
INSERT INTO "product_item" (
  product_id,
  product_sku,
  qty_in_stock,
  product_image,
  price,
  active
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetProductItem :one
SELECT * FROM "product_item"
WHERE id = $1 LIMIT 1;

-- name: ListProductItems :many
SELECT * FROM "product_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateProductItem :one
UPDATE "product_item"
SET
product_id = COALESCE(sqlc.narg(product_id),product_id),
product_sku = COALESCE(sqlc.narg(product_sku),product_sku),
qty_in_stock = COALESCE(sqlc.narg(qty_in_stock),qty_in_stock),
product_image = COALESCE(sqlc.narg(product_image),product_image),
price = COALESCE(sqlc.narg(price),price),
active = COALESCE(sqlc.narg(active),active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductItem :exec
DELETE FROM "product_item"
WHERE id = $1;