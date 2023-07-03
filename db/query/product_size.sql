-- name: CreateProductSize :one
INSERT INTO "product_size" (
  size_value
) VALUES (
  $1
)
RETURNING *;

-- name: GetProductSize :one
SELECT * FROM "product_size"
WHERE id = $1 LIMIT 1;

-- name: UpdateProductSize :one
UPDATE "product_size"
SET 
size_value = COALESCE(sqlc.narg(size_value),size_value)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductSize :exec
DELETE FROM "product_size"
WHERE id = $1;