-- name: AdminCreateFeaturedProductItem :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "featured_product_item" (
  product_item_id,
  active,
  start_date,
  end_date,
  priority
)
SELECT sqlc.arg(product_item_id), sqlc.arg(active), sqlc.arg(start_date), sqlc.arg(end_date), sqlc.arg(priority) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetFeaturedProductItem :one
SELECT * FROM "featured_product_item"
WHERE product_item_id = $1
LIMIT 1;

-- name: ListFeaturedProductItems :many
SELECT * FROM "featured_product_item"
ORDER BY product_item_id
LIMIT $1
OFFSET $2;

-- name: AdminListFeaturedProductItems :many
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT 
fp.*, pi.*, p.name AS product_name, 
p.description FROM "featured_product_item" AS fp
LEFT JOIN "product_item" AS pi ON pi.id = fp.product_item_id
LEFT JOIN "product" AS p ON p.id = pi.product_id
WHERE (SELECT is_admin FROM t1) = 1
ORDER BY product_item_id;

-- name: AdminUpdateFeaturedProductItem :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "featured_product_item"
SET
active = COALESCE(sqlc.narg(active),active),
start_date = COALESCE(sqlc.narg(start_date),start_date),
end_date = COALESCE(sqlc.narg(end_date),end_date),
priority = COALESCE(sqlc.narg(priority),priority)
WHERE product_item_id = sqlc.arg(product_item_id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;


-- name: DeleteFeaturedProductItem :exec
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "featured_product_item"
WHERE product_item_id = $1
AND (SELECT is_admin FROM t1) = 1
RETURNING *;