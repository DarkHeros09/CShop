-- name: CreateProductSize :one
INSERT INTO "product_size" (
  product_item_id,
  size_value,
  qty
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: AdminCreateProductSize :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_size" (
 product_item_id,
  size_value,
  qty
)
SELECT 
sqlc.arg(product_item_id), 
sqlc.arg(size_value), 
sqlc.arg(qty) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProductSize :one
SELECT * FROM "product_size"
WHERE id = $1 LIMIT 1;

-- name: GetProductItemSizeForUpdate :one
SELECT * FROM "product_size"
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListProductSizes :many
SELECT * FROM "product_size"
ORDER BY id;

-- name: UpdateProductSize :one
UPDATE "product_size"
SET 
size_value = COALESCE(sqlc.narg(size_value),size_value),
qty = COALESCE(sqlc.narg(qty),qty)
WHERE id = sqlc.arg(id)
AND product_item_id = sqlc.arg(product_item_id)
RETURNING *;

-- name: AdminUpdateProductSize :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "product_size"
SET 
size_value = COALESCE(sqlc.narg(size_value),size_value),
qty = COALESCE(sqlc.narg(qty),qty)
WHERE "product_size".id = sqlc.arg(id)
AND product_item_id = sqlc.arg(product_item_id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: DeleteProductSize :exec
DELETE FROM "product_size"
WHERE id = $1;