-- name: CreateProductItem :one
INSERT INTO "product_item" (
  product_id,
  size_id,
  image_id,
  color_id,
  product_sku,
  qty_in_stock,
  -- product_image,
  price,
  active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetProductItem :one
SELECT * FROM "product_item"
WHERE id = $1 LIMIT 1;

-- name: GetProductItemForUpdate :one
SELECT * FROM "product_item"
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListProductItemsByIDs :many
SELECT pi.id, p.name, pi.product_id, 
pi.price, pi.active, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
WHERE pi.id = ANY(sqlc.arg(products_ids)::bigint[]);

-- name: ListProductItems :many
SELECT pi.*, p.*, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value,
COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
ORDER BY pi.id
LIMIT $1
OFFSET $2;

-- name: ListProductItemsV2 :many
WITH t1 AS (
SELECT COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
WHERE pi.active = TRUE
LIMIT 1
)
SELECT pi.*, p.name, p.description, p.category_id, p.active as parent_product_active, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value,
(SELECT total_count FROM t1)
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
WHERE pi.active = TRUE AND p.active = TRUE
ORDER BY pi.id DESC
LIMIT $1;

-- name: ListProductItemsNextPage :many
WITH t1 AS (
SELECT COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
WHERE pi.active = TRUE
LIMIT 1
)
SELECT pi.*, p.name, p.description, p.category_id, p.active as parent_product_active, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value,
(SELECT total_count FROM t1)
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
WHERE pi.id < $2 AND pi.active = TRUE AND p.active = TRUE
ORDER BY pi.id DESC
LIMIT $1;

-- name: SearchProductItems :many
WITH t1 AS (
SELECT COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
WHERE pi.active = TRUE
LIMIT 1
)
SELECT pi.*, p.name, p.description, p.category_id, p.active as parent_product_active, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value,
(SELECT total_count FROM t1)
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
WHERE pi.active = TRUE AND p.active = TRUE AND p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
ORDER BY pi.id DESC, ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1;

-- name: SearchProductItemsNextPage :many
WITH t1 AS (
SELECT COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
WHERE pi.active = TRUE
LIMIT 1
)
SELECT pi.*, p.name, p.description, p.category_id, p.active as parent_product_active, ps.size_value, pimg.product_image_1,
pimg.product_image_2, pimg.product_image_3, pclr.color_value,
(SELECT total_count FROM t1)
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
WHERE pi.active = TRUE AND p.active = TRUE AND p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
AND pi.id < $2
ORDER BY pi.id DESC, ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1;

-- name: UpdateProductItem :one
UPDATE "product_item"
SET
product_sku = COALESCE(sqlc.narg(product_sku),product_sku),
qty_in_stock = COALESCE(sqlc.narg(qty_in_stock),qty_in_stock),
size_id = COALESCE(sqlc.narg(size_id),size_id),
image_id = COALESCE(sqlc.narg(image_id),image_id),
color_id = COALESCE(sqlc.narg(color_id),color_id),
price = COALESCE(sqlc.narg(price),price),
active = COALESCE(sqlc.narg(active),active),
updated_at = now()
WHERE id = sqlc.arg(id)
AND product_id = sqlc.arg(product_id)
RETURNING *;

-- name: DeleteProductItem :exec
DELETE FROM "product_item"
WHERE id = $1;