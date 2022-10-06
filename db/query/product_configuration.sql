-- name: CreateProductConfiguration :one
INSERT INTO "product_configuration" (
  product_item_id,
  variation_option_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetProductConfiguration :one
SELECT * FROM "product_configuration"
WHERE product_item_id = $1 LIMIT 1;

-- name: ListProductConfigurations :many
SELECT * FROM "product_configuration"
ORDER BY product_item_id
LIMIT $1
OFFSET $2;

-- name: UpdateProductConfiguration :one
UPDATE "product_configuration"
SET
variation_option_id = COALESCE(sqlc.narg(variation_option_id),variation_option_id)
WHERE product_item_id = sqlc.arg(product_item_id)
RETURNING *;

-- name: DeleteProductConfiguration :exec
DELETE FROM "product_configuration"
WHERE product_item_id = $1;