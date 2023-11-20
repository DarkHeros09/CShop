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

-- name: ListProductItemsV2Old :many
-- WITH t1 (total_count) AS (
-- SELECT COUNT(*) OVER() AS total_count
-- FROM "product_item" AS pi
-- WHERE pi.active = TRUE
-- LIMIT 1
-- )
SELECT pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active as parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value, COUNT(*) OVER() AS total_count,
 cpromo.id as category_promo_id, cpromo.name as category_promo_name, cpromo.description as category_promo_description,
 cpromo.discount_rate as category_promo_discount_rate, COALESCE(cpromo.active, false) as category_promo_active,
 cpromo.start_date as category_promo_start_date, cpromo.end_date as category_promo_end_date,
 bpromo.id as brand_promo_id, bpromo.name as brand_promo_name, bpromo.description as brand_promo_description,
 bpromo.discount_rate as brand_promo_discount_rate, COALESCE(bpromo.active, false) as brand_promo_active,
 bpromo.start_date as brand_promo_start_date, bpromo.end_date as brand_promo_end_date,
 ppromo.id as product_promo_id, ppromo.name as product_promo_name, ppromo.description as product_promo_description,
 ppromo.discount_rate as product_promo_discount_rate, COALESCE(ppromo.active, false) as product_promo_active,
 ppromo.start_date as product_promo_start_date, ppromo.end_date as product_promo_end_date
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id AND p.active = TRUE
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id AND cp.active = true
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id AND cpromo.active =true AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id AND bp.active = true
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id AND bpromo.active =true AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id AND pp.active = true
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id AND ppromo.active = true AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE

WHERE pi.active = TRUE

AND CASE
WHEN COALESCE(sqlc.narg(is_promoted), false) = TRUE
THEN (cpromo.active = true OR
	bpromo.active = true OR
	ppromo.active = true)
ELSE 1=1
END
AND CASE
WHEN COALESCE(sqlc.narg(is_new), false) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY pi.id DESC
LIMIT $1;

-- name: ListProductItemsNextPageOld :many
-- WITH t1 AS (
-- SELECT COUNT(*) OVER() AS total_count
-- FROM "product_item" AS pi
-- WHERE pi.active = TRUE
-- LIMIT 1
-- )
SELECT pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active as parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value, COUNT(*) OVER() AS total_count,
 cpromo.id as category_promo_id, cpromo.name as category_promo_name, cpromo.description as category_promo_description,
 cpromo.discount_rate as category_promo_discount_rate, COALESCE(cpromo.active, false) as category_promo_active,
 cpromo.start_date as category_promo_start_date, cpromo.end_date as category_promo_end_date,
 bpromo.id as brand_promo_id, bpromo.name as brand_promo_name, bpromo.description as brand_promo_description,
 bpromo.discount_rate as brand_promo_discount_rate, COALESCE(bpromo.active, false) as brand_promo_active,
 bpromo.start_date as brand_promo_start_date, bpromo.end_date as brand_promo_end_date,
 ppromo.id as product_promo_id, ppromo.name as product_promo_name, ppromo.description as product_promo_description,
 ppromo.discount_rate as product_promo_discount_rate, COALESCE(ppromo.active, false) as product_promo_active,
 ppromo.start_date as product_promo_start_date, ppromo.end_date as product_promo_end_date
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id AND p.active = TRUE
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id AND cp.active = true
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id AND cpromo.active =true AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id AND bp.active = true
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id AND bpromo.active =true AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id AND pp.active = true
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id AND ppromo.active = true AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE

WHERE pi.id < $2
AND pi.active = TRUE

AND CASE
WHEN COALESCE(sqlc.narg(is_promoted), false) = TRUE
THEN (cpromo.active = true OR
	bpromo.active = true OR
	ppromo.active = true)
ELSE 1=1
END
AND CASE
WHEN COALESCE(sqlc.narg(is_new), false) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY pi.id DESC
LIMIT $1;

-- name: ListProductItemsV2 :many
WITH t1 AS (
SELECT pc.id AS product_category_id, pc.parent_category_id ,pc.category_name, pc.category_image, cp.category_promotion_image, cp.active AS category_promotion_active, 
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_category" AS pc
INNER JOIN
"category_promotion" AS cp ON cp.category_id = pc.id
INNER JOIN
"promotion" AS cpromo ON cpromo.id = cp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN cp.active = TRUE
AND cpromo.active =TRUE AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END 
)
, t2 AS(
SELECT pb.id AS product_brand_id, pb.brand_name, pb.brand_image, bp.brand_promotion_image, bp.active AS brand_promotion_active, 
bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_brand" AS pb
INNER JOIN
"brand_promotion" AS bp ON bp.brand_id = pb.id
INNER JOIN
"promotion" AS bpromo ON bpromo.id = bp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN bp.active = TRUE
AND bpromo.active = TRUE AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END 
)
, t3 AS(
SELECT p.id AS pid ,p.name, p.description, p.category_id, p.brand_id, p.active AS parent_product_active, 
pp.product_id AS promotion_product_id , pp.product_promotion_image, pp.active AS product_promotion_active, 
ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date,
t1.*, t2.*
 FROM "product" AS p
 LEFT JOIN
 "product_promotion" AS pp ON pp.product_id = p.id
LEFT JOIN
"promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN t1 ON t1.product_category_id = p.category_id
LEFT JOIN t2 ON t2.product_brand_id = p.brand_id
WHERE 
p.active = TRUE
AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
pp.active = TRUE AND 
(
    pp.product_id IS NOT NULL OR
    t1.product_category_id IS NOT NULL OR
    t2.product_brand_id IS NOT NULL
)
AND
ppromo.active = TRUE 
AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE 
ELSE 1=1
END 
)
, t4 AS (
SELECT *, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
INNER JOIN t3 ON t3.pid = pi.product_id
WHERE 
pi.active = TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pi.color_id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN pi.size_id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY 
	CASE
	WHEN COALESCE(sqlc.narg(order_by_id), TRUE) = TRUE
	THEN pi.id
	ELSE 0
END DESC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC
)
, t5 AS (
SELECT t4.*, ps.size_value FROM t4
LEFT JOIN "product_size" AS ps ON ps.id = t4.size_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
)
, t6 AS (
SELECT t5.*, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM t5
LEFT JOIN "product_image" AS pimg ON pimg.id = t5.image_id
)
, t7 AS (
SELECT t6.*, pclr.color_value FROM t6
LEFT JOIN "product_color" AS pclr ON pclr.id = t6.color_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
)

SELECT
t7.id, t7.product_id, t7.size_id, t7.image_id, t7.color_id, t7.product_sku, t7.qty_in_stock, t7.price,
t7.active, t7.created_at, t7.updated_at, t7.name, t7.description, t7.category_id, t7.category_name, t7.parent_category_id,
 t7.category_image, t7.brand_id, t7.brand_name, t7.brand_image, t7.parent_product_active, t7.size_value,
 t7.product_image_1, t7.product_image_2, t7.product_image_3, t7.color_value, 
 t7.total_count,
 t7.category_promo_id, t7.category_promo_name, t7.category_promo_description,
 t7.category_promo_discount_rate, t7.category_promo_active,
 t7.category_promo_start_date, t7.category_promo_end_date,
 t7.brand_promo_id, t7.brand_promo_name, t7.brand_promo_description,
 t7.brand_promo_discount_rate, t7.brand_promo_active,
 t7.brand_promo_start_date, t7.brand_promo_end_date,
 t7.product_promo_id, t7.product_promo_name, t7.product_promo_description,
 t7.product_promo_discount_rate, t7.product_promo_active,
 t7.product_promo_start_date, t7.product_promo_end_date
FROM t7
LIMIT $1;

-- name: ListProductItemsNextPage :many
WITH t1 AS (
SELECT pc.id AS product_category_id, pc.parent_category_id ,pc.category_name, pc.category_image, cp.category_promotion_image, cp.active AS category_promotion_active, 
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_category" AS pc
INNER JOIN
"category_promotion" AS cp ON cp.category_id = pc.id
INNER JOIN
"promotion" AS cpromo ON cpromo.id = cp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN cp.active = TRUE
AND cpromo.active =TRUE AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END 
)
, t2 AS(
SELECT pb.id AS product_brand_id, pb.brand_name, pb.brand_image, bp.brand_promotion_image, bp.active AS brand_promotion_active, 
bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_brand" AS pb
INNER JOIN
"brand_promotion" AS bp ON bp.brand_id = pb.id
INNER JOIN
"promotion" AS bpromo ON bpromo.id = bp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN bp.active = TRUE
AND bpromo.active = TRUE AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END 
)
, t3 AS(
SELECT p.id AS pid ,p.name, p.description, p.category_id, p.brand_id, p.active AS parent_product_active, 
pp.product_id AS promotion_product_id , pp.product_promotion_image, pp.active AS product_promotion_active, 
ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date,
t1.*, t2.*
 FROM "product" AS p
 LEFT JOIN
 "product_promotion" AS pp ON pp.product_id = p.id
LEFT JOIN
"promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN t1 ON t1.product_category_id = p.category_id
LEFT JOIN t2 ON t2.product_brand_id = p.brand_id
WHERE 
p.id < sqlc.arg(product_id) AND
p.active = TRUE
AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
pp.active = TRUE AND 
(
    pp.product_id IS NOT NULL OR
    t1.product_category_id IS NOT NULL OR
    t2.product_brand_id IS NOT NULL
)
AND
ppromo.active = TRUE 
AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE 
ELSE 1=1
END 
)
, t4 AS (
SELECT *, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
INNER JOIN t3 ON t3.pid = pi.product_id
WHERE 
pi.id < sqlc.arg(product_item_id) AND
pi.active = TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pi.color_id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN pi.size_id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY 
	CASE
	WHEN COALESCE(sqlc.narg(order_by_id), TRUE) = TRUE
	THEN pi.id
	ELSE 0
END DESC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC
)
, t5 AS (
SELECT t4.*, ps.size_value FROM t4
LEFT JOIN "product_size" AS ps ON ps.id = t4.size_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
)
, t6 AS (
SELECT t5.*, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM t5
LEFT JOIN "product_image" AS pimg ON pimg.id = t5.image_id
)
, t7 AS (
SELECT t6.*, pclr.color_value FROM t6
LEFT JOIN "product_color" AS pclr ON pclr.id = t6.color_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
)

SELECT
t7.id, t7.product_id, t7.size_id, t7.image_id, t7.color_id, t7.product_sku, t7.qty_in_stock, t7.price,
t7.active, t7.created_at, t7.updated_at, t7.name, t7.description, t7.category_id, t7.category_name, t7.parent_category_id,
 t7.category_image, t7.brand_id, t7.brand_name, t7.brand_image, t7.parent_product_active, t7.size_value,
 t7.product_image_1, t7.product_image_2, t7.product_image_3, t7.color_value, 
 t7.total_count,
 t7.category_promo_id, t7.category_promo_name, t7.category_promo_description,
 t7.category_promo_discount_rate, t7.category_promo_active,
 t7.category_promo_start_date, t7.category_promo_end_date,
 t7.brand_promo_id, t7.brand_promo_name, t7.brand_promo_description,
 t7.brand_promo_discount_rate, t7.brand_promo_active,
 t7.brand_promo_start_date, t7.brand_promo_end_date,
 t7.product_promo_id, t7.product_promo_name, t7.product_promo_description,
 t7.product_promo_discount_rate, t7.product_promo_active,
 t7.product_promo_start_date, t7.product_promo_end_date
FROM t7
LIMIT $1;

-- name: SearchProductItemsOld :many
-- WITH t1 AS (
-- SELECT COUNT(*) OVER() AS total_count
-- FROM "product_item" AS pi
-- WHERE pi.active = TRUE
-- LIMIT 1
-- )
SELECT pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active as parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value, COUNT(*) OVER() AS total_count,
 cpromo.id as category_promo_id, cpromo.name as category_promo_name, cpromo.description as category_promo_description,
 cpromo.discount_rate as category_promo_discount_rate, COALESCE(cpromo.active, false) as category_promo_active,
 cpromo.start_date as category_promo_start_date, cpromo.end_date as category_promo_end_date,
 bpromo.id as brand_promo_id, bpromo.name as brand_promo_name, bpromo.description as brand_promo_description,
 bpromo.discount_rate as brand_promo_discount_rate, COALESCE(bpromo.active, false) as brand_promo_active,
 bpromo.start_date as brand_promo_start_date, bpromo.end_date as brand_promo_end_date,
 ppromo.id as product_promo_id, ppromo.name as product_promo_name, ppromo.description as product_promo_description,
 ppromo.discount_rate as product_promo_discount_rate, COALESCE(ppromo.active, false) as product_promo_active,
 ppromo.start_date as product_promo_start_date, ppromo.end_date as product_promo_end_date
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id AND p.active = TRUE
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id AND cp.active = true
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id AND cpromo.active =true AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id AND bp.active = true
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id AND bpromo.active =true AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id AND pp.active = true
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id AND ppromo.active = true AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE

WHERE pi.active = TRUE AND p.search @@ 
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

-- name: SearchProductItems :many
WITH t1 AS (
SELECT pc.id AS product_category_id, pc.parent_category_id ,pc.category_name, pc.category_image, cp.category_promotion_image, cp.active AS category_promotion_active, 
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_category" AS pc
INNER JOIN
"category_promotion" AS cp ON cp.category_id = pc.id
INNER JOIN
"promotion" AS cpromo ON cpromo.id = cp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN cp.active = TRUE
AND cpromo.active =TRUE AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END 
)
, t2 AS(
SELECT pb.id AS product_brand_id, pb.brand_name, pb.brand_image, bp.brand_promotion_image, bp.active AS brand_promotion_active, 
bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_brand" AS pb
INNER JOIN
"brand_promotion" AS bp ON bp.brand_id = pb.id
INNER JOIN
"promotion" AS bpromo ON bpromo.id = bp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN bp.active = TRUE
AND bpromo.active = TRUE AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END 
)
, t3 AS(
SELECT p.id AS pid ,p.name, p.description, p.category_id, p.brand_id, p.active AS parent_product_active, p.search,
pp.product_id AS promotion_product_id , pp.product_promotion_image, pp.active AS product_promotion_active, 
ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date,
t1.*, t2.*
 FROM "product" AS p
 LEFT JOIN
 "product_promotion" AS pp ON pp.product_id = p.id
LEFT JOIN
"promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN t1 ON t1.product_category_id = p.category_id
LEFT JOIN t2 ON t2.product_brand_id = p.brand_id
WHERE 
p.active = TRUE AND p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
pp.active = TRUE AND 
(
    pp.product_id IS NOT NULL OR
    t1.product_category_id IS NOT NULL OR
    t2.product_brand_id IS NOT NULL
)
AND
ppromo.active = TRUE 
AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE 
ELSE 1=1
END 
)
, t4 AS (
SELECT *, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
INNER JOIN t3 ON t3.pid = pi.product_id
WHERE 
pi.active = TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pi.color_id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN pi.size_id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY 
	CASE
	WHEN COALESCE(sqlc.narg(order_by_id), TRUE) = TRUE
	THEN pi.id
	ELSE 0
END DESC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC , ts_rank(t3.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
)
, t5 AS (
SELECT t4.*, ps.size_value FROM t4
LEFT JOIN "product_size" AS ps ON ps.id = t4.size_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
)
, t6 AS (
SELECT t5.*, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM t5
LEFT JOIN "product_image" AS pimg ON pimg.id = t5.image_id
)
, t7 AS (
SELECT t6.*, pclr.color_value FROM t6
LEFT JOIN "product_color" AS pclr ON pclr.id = t6.color_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
)

SELECT
t7.id, t7.product_id, t7.size_id, t7.image_id, t7.color_id, t7.product_sku, t7.qty_in_stock, t7.price,
t7.active, t7.created_at, t7.updated_at, t7.name, t7.description, t7.category_id, t7.category_name, t7.parent_category_id,
 t7.category_image, t7.brand_id, t7.brand_name, t7.brand_image, t7.parent_product_active, t7.size_value,
 t7.product_image_1, t7.product_image_2, t7.product_image_3, t7.color_value, 
 t7.total_count,
 t7.category_promo_id, t7.category_promo_name, t7.category_promo_description,
 t7.category_promo_discount_rate, t7.category_promo_active,
 t7.category_promo_start_date, t7.category_promo_end_date,
 t7.brand_promo_id, t7.brand_promo_name, t7.brand_promo_description,
 t7.brand_promo_discount_rate, t7.brand_promo_active,
 t7.brand_promo_start_date, t7.brand_promo_end_date,
 t7.product_promo_id, t7.product_promo_name, t7.product_promo_description,
 t7.product_promo_discount_rate, t7.product_promo_active,
 t7.product_promo_start_date, t7.product_promo_end_date
FROM t7
LIMIT $1;

-- name: SearchProductItemsNextPageOld :many
-- WITH t1 AS (
-- SELECT COUNT(*) OVER() AS total_count
-- FROM "product_item" AS pi
-- WHERE pi.active = TRUE
-- LIMIT 1
-- )
SELECT pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active as parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value, COUNT(*) OVER() AS total_count,
 cpromo.id as category_promo_id, cpromo.name as category_promo_name, cpromo.description as category_promo_description,
 cpromo.discount_rate as category_promo_discount_rate, COALESCE(cpromo.active, false) as category_promo_active,
 cpromo.start_date as category_promo_start_date, cpromo.end_date as category_promo_end_date,
 bpromo.id as brand_promo_id, bpromo.name as brand_promo_name, bpromo.description as brand_promo_description,
 bpromo.discount_rate as brand_promo_discount_rate, COALESCE(bpromo.active, false) as brand_promo_active,
 bpromo.start_date as brand_promo_start_date, bpromo.end_date as brand_promo_end_date,
 ppromo.id as product_promo_id, ppromo.name as product_promo_name, ppromo.description as product_promo_description,
 ppromo.discount_rate as product_promo_discount_rate, COALESCE(ppromo.active, false) as product_promo_active,
 ppromo.start_date as product_promo_start_date, ppromo.end_date as product_promo_end_date
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id AND p.active = TRUE
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id AND cp.active = true
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id AND cpromo.active =true AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id AND bp.active = true
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id AND bpromo.active =true AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id AND pp.active = true
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id AND ppromo.active = true AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE
WHERE pi.active = TRUE AND p.search @@  
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

-- name: SearchProductItemsNextPage :many
WITH t1 AS (
SELECT pc.id AS product_category_id, pc.parent_category_id ,pc.category_name, pc.category_image, cp.category_promotion_image, cp.active AS category_promotion_active, 
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_category" AS pc
INNER JOIN
"category_promotion" AS cp ON cp.category_id = pc.id
INNER JOIN
"promotion" AS cpromo ON cpromo.id = cp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN cp.active = TRUE
AND cpromo.active =TRUE AND cpromo.start_date <= CURRENT_DATE AND cpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(category_id), 0) > 0 
    THEN pc.id = sqlc.narg(category_id)
    ELSE 1=1
END 
)
, t2 AS(
SELECT pb.id AS product_brand_id, pb.brand_name, pb.brand_image, bp.brand_promotion_image, bp.active AS brand_promotion_active, 
bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_brand" AS pb
INNER JOIN
"brand_promotion" AS bp ON bp.brand_id = pb.id
INNER JOIN
"promotion" AS bpromo ON bpromo.id = bp.promotion_id
WHERE CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN bp.active = TRUE
AND bpromo.active = TRUE AND bpromo.start_date <= CURRENT_DATE AND bpromo.end_date >= CURRENT_DATE
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(brand_id), 0) > 0 
    THEN pb.id = sqlc.narg(brand_id)
    ELSE 1=1
END 
)
, t3 AS(
SELECT p.id AS pid ,p.name, p.description, p.category_id, p.brand_id, p.active AS parent_product_active, p.search,
pp.product_id AS promotion_product_id , pp.product_promotion_image, pp.active AS product_promotion_active, 
ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date,
t1.*, t2.*
 FROM "product" AS p
 LEFT JOIN
 "product_promotion" AS pp ON pp.product_id = p.id
LEFT JOIN
"promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN t1 ON t1.product_category_id = p.category_id
LEFT JOIN t2 ON t2.product_brand_id = p.brand_id
WHERE 
p.id < sqlc.arg(product_id) AND
p.active = TRUE AND p.search @@  
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
pp.active = TRUE AND 
(
    pp.product_id IS NOT NULL OR
    t1.product_category_id IS NOT NULL OR
    t2.product_brand_id IS NOT NULL
)
AND
ppromo.active = TRUE 
AND ppromo.start_date <= CURRENT_DATE AND ppromo.end_date >= CURRENT_DATE 
ELSE 1=1
END 
)
, t4 AS (
SELECT *, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
INNER JOIN t3 ON t3.pid = pi.product_id
WHERE 
pi.id < sqlc.arg(product_item_id) AND
pi.active = TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
THEN pi.created_at >= CURRENT_DATE - INTERVAL '5 days'
ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pi.color_id = sqlc.narg(color_id)
    ELSE 1=1
END
AND CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN pi.size_id = sqlc.narg(size_id)
    ELSE 1=1
END
ORDER BY 
	CASE
	WHEN COALESCE(sqlc.narg(order_by_id), TRUE) = TRUE
	THEN pi.id
	ELSE 0
END DESC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC, ts_rank(t3.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
)
, t5 AS (
SELECT t4.*, ps.size_value FROM t4
LEFT JOIN "product_size" AS ps ON ps.id = t4.size_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(size_id), 0) > 0 
    THEN ps.id = sqlc.narg(size_id)
    ELSE 1=1
END
)
, t6 AS (
SELECT t5.*, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM t5
LEFT JOIN "product_image" AS pimg ON pimg.id = t5.image_id
)
, t7 AS (
SELECT t6.*, pclr.color_value FROM t6
LEFT JOIN "product_color" AS pclr ON pclr.id = t6.color_id
WHERE CASE
    WHEN COALESCE(sqlc.narg(color_id), 0) > 0 
    THEN pclr.id = sqlc.narg(color_id)
    ELSE 1=1
END
)

SELECT
t7.id, t7.product_id, t7.size_id, t7.image_id, t7.color_id, t7.product_sku, t7.qty_in_stock, t7.price,
t7.active, t7.created_at, t7.updated_at, t7.name, t7.description, t7.category_id, t7.category_name, t7.parent_category_id,
 t7.category_image, t7.brand_id, t7.brand_name, t7.brand_image, t7.parent_product_active, t7.size_value,
 t7.product_image_1, t7.product_image_2, t7.product_image_3, t7.color_value, 
 t7.total_count,
 t7.category_promo_id, t7.category_promo_name, t7.category_promo_description,
 t7.category_promo_discount_rate, t7.category_promo_active,
 t7.category_promo_start_date, t7.category_promo_end_date,
 t7.brand_promo_id, t7.brand_promo_name, t7.brand_promo_description,
 t7.brand_promo_discount_rate, t7.brand_promo_active,
 t7.brand_promo_start_date, t7.brand_promo_end_date,
 t7.product_promo_id, t7.product_promo_name, t7.product_promo_description,
 t7.product_promo_discount_rate, t7.product_promo_active,
 t7.product_promo_start_date, t7.product_promo_end_date
FROM t7
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