-- name: CreateProductSize :one
INSERT INTO "product_size" (
  size_value
) VALUES (
  $1
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
size_value
)
SELECT sqlc.arg(size_value) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProductSize :one
SELECT * FROM "product_size"
WHERE id = $1 LIMIT 1;

-- name: ListProductSizes :many
SELECT * FROM "product_size"
ORDER BY id;

-- name: UpdateProductSize :one
UPDATE "product_size"
SET 
size_value = COALESCE(sqlc.narg(size_value),size_value)
WHERE id = sqlc.arg(id)
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
size_value = COALESCE(sqlc.narg(size_value),size_value)
WHERE "product_size".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: DeleteProductSize :exec
DELETE FROM "product_size"
WHERE id = $1;