-- name: CreateProductBrand :one
INSERT INTO "product_brand" (
  brand_name,
  brand_image
) VALUES (
  $1, $2
)
ON CONFLICT(brand_name) DO UPDATE SET 
brand_name = EXCLUDED.brand_name,
brand_image = EXCLUDED.brand_image
RETURNING *;

-- name: GetProductBrand :one
SELECT * FROM "product_brand"
WHERE id = $1 LIMIT 1;

-- name: ListProductBrands :many
SELECT * FROM "product_brand"
ORDER BY id;
-- LIMIT $1
-- OFFSET $2;

-- name: UpdateProductBrand :one
UPDATE "product_brand"
SET brand_name = sqlc.arg(brand_name)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductBrand :exec
DELETE FROM "product_brand"
WHERE id = sqlc.arg(id);
