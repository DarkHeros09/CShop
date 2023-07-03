-- name: CreateProductColor :one
INSERT INTO "product_color" (
  color_value
) VALUES (
  $1
)
RETURNING *;

-- name: GetProductColor :one
SELECT * FROM "product_color"
WHERE id = $1 LIMIT 1;

-- name: UpdateProductColor :one
UPDATE "product_color"
SET 
color_value = COALESCE(sqlc.narg(color_value),color_value)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductColor :exec
DELETE FROM "product_color"
WHERE id = $1;