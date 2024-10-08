-- name: CreateProductImage :one
INSERT INTO "product_image" (
  product_image_1,
  product_image_2,
  product_image_3
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: AdminCreateProductImages :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_image" (
  product_image_1,
  product_image_2,
  product_image_3
)
SELECT sqlc.arg(product_image_1), sqlc.arg(product_image_2), sqlc.arg(product_image_3) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProductImage :one
SELECT * FROM "product_image"
WHERE id = $1 LIMIT 1;

-- name: ListProductImagesV2 :many
WITH t1 AS(
SELECT 
 pimg.id, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM "product_image" AS pimg
ORDER BY id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductImagesNextPage :many
WITH t1 AS(
SELECT 
 pimg.id, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM "product_image" AS pimg
WHERE
 pimg.id < sqlc.arg(id) 
ORDER BY id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: UpdateProductImage :one
UPDATE "product_image"
SET 
product_image_1 = COALESCE(sqlc.narg(product_image_1),product_image_1),
product_image_2 = COALESCE(sqlc.narg(product_image_2),product_image_2),
product_image_3 = COALESCE(sqlc.narg(product_image_3),product_image_3)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdminUpdateProductImage :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "product_image"
SET 
product_image_1 = COALESCE(sqlc.narg(product_image_1),product_image_1),
product_image_2 = COALESCE(sqlc.narg(product_image_2),product_image_2),
product_image_3 = COALESCE(sqlc.narg(product_image_3),product_image_3)
WHERE "product_image".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: DeleteProductImage :exec
DELETE FROM "product_image"
WHERE id = $1;