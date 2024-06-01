-- name: CreateProductCategory :one
INSERT INTO "product_category" (
  parent_category_id,
  category_name,
  category_image
) VALUES (
  $1, $2, $3
)
ON CONFLICT(category_name) DO UPDATE SET 
category_name = EXCLUDED.category_name,
category_image = EXCLUDED.category_image
RETURNING *;

-- name: AdminCreateProductCategory :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_category" (
  parent_category_id,
  category_name,
  category_image
)
SELECT sqlc.arg(parent_category_id), sqlc.arg(category_name), sqlc.arg(category_image) FROM t1
WHERE is_admin=1
ON CONFLICT(category_name) DO UPDATE SET 
category_name = EXCLUDED.category_name,
category_image = EXCLUDED.category_image
RETURNING *;

-- name: GetProductCategory :one
SELECT * FROM "product_category"
WHERE id = $1 LIMIT 1;

-- name: GetProductCategoryByParent :one
SELECT * FROM "product_category"
WHERE id = $1
And parent_category_id = $2
LIMIT 1;

-- name: ListProductCategories :many
SELECT * FROM "product_category"
ORDER BY id;
-- LIMIT $1
-- OFFSET $2;

-- name: ListProductCategoriesByParent :many
SELECT * FROM "product_category"
WHERE parent_category_id = $1
ORDER BY id;
-- LIMIT $2
-- OFFSET $3;

-- name: UpdateProductCategory :one
UPDATE "product_category"
SET category_name = sqlc.arg(category_name)
WHERE id = sqlc.arg(id)
AND
( parent_category_id is NULL OR parent_category_id = sqlc.arg(parent_category_id) )
RETURNING *;

-- name: DeleteProductCategory :exec
DELETE FROM "product_category"
WHERE id = sqlc.arg(id)
AND ( parent_category_id is NULL OR parent_category_id = sqlc.arg(parent_category_id) );