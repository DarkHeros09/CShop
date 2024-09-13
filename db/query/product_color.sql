-- name: CreateProductColor :one
INSERT INTO "product_color" (
  color_value
) VALUES (
  $1
)
RETURNING *;

-- name: AdminCreateProductColor :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_color" (
color_value
)
SELECT sqlc.arg(color_value) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProductColor :one
SELECT * FROM "product_color"
WHERE id = $1 LIMIT 1;

-- name: ListProductColors :many
SELECT * FROM "product_color"
ORDER BY id;

-- name: UpdateProductColor :one
UPDATE "product_color"
SET 
color_value = COALESCE(sqlc.narg(color_value),color_value)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdminUpdateProductColor :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "product_color"
SET 
color_value = COALESCE(sqlc.narg(color_value),color_value)
WHERE "product_color".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: DeleteProductColor :exec
DELETE FROM "product_color"
WHERE id = $1;