-- name: CreateProductImage :one
INSERT INTO "product_image" (
  product_image_1,
  product_image_2,
  product_image_3
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetProductImage :one
SELECT * FROM "product_image"
WHERE id = $1 LIMIT 1;

-- name: UpdateProductImage :one
UPDATE "product_image"
SET 
product_image_1 = COALESCE(sqlc.narg(product_image_1),product_image_1),
product_image_2 = COALESCE(sqlc.narg(product_image_2),product_image_2),
product_image_3 = COALESCE(sqlc.narg(product_image_3),product_image_3)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductImage :exec
DELETE FROM "product_image"
WHERE id = $1;