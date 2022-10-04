-- name: CreateProductCategory :one
INSERT INTO "product_category" (
  parent_category_id,
  category_name
) VALUES (
  $1, $2
)
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
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListProductCategoriesByParent :many
SELECT * FROM "product_category"
WHERE parent_category_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateProductCategory :one
UPDATE "product_category"
SET category_name = sqlc.arg(category_name)
WHERE id = sqlc.arg(id)
And
( parent_category_id is NULL OR parent_category_id = sqlc.arg(parent_category_id) )
RETURNING *;

-- name: DeleteProductCategory :exec
DELETE FROM "product_category"
WHERE id = sqlc.arg(id)
AND ( parent_category_id is NULL OR parent_category_id = sqlc.arg(parent_category_id) );