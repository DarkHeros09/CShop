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

-- name: AdminCreateProductBrand :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_brand" (
  brand_name,
  brand_image
) 
SELECT sqlc.arg(brand_name), sqlc.arg(brand_image) FROM t1
WHERE is_admin=1
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
