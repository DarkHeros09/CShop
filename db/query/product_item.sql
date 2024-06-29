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

-- name: AdminCreateProductItem :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
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
)
SELECT sqlc.arg(product_id), sqlc.arg(size_id), sqlc.arg(image_id), sqlc.arg(color_id), sqlc.arg(product_sku), sqlc.arg(qty_in_stock), sqlc.arg(price), sqlc.arg(active) FROM t1
WHERE is_admin=1
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
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id 
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id  
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
WHERE 
pi.active = TRUE AND
p.active =TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
((pp.active = TRUE
AND ppromo.active = TRUE
AND ppromo.start_date <= CURRENT_DATE 
AND ppromo.end_date >= CURRENT_DATE)
OR (cp.active = TRUE
AND cpromo.active =TRUE 
AND cpromo.start_date <= CURRENT_DATE 
AND cpromo.end_date >= CURRENT_DATE)
OR (bp.active = TRUE
AND bpromo.active =TRUE 
AND bpromo.start_date <= CURRENT_DATE 
AND bpromo.end_date >= CURRENT_DATE))=TRUE 
ELSE TRUE
END 
AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
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
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC,
CASE
    WHEN sqlc.narg(order_by_low_price) IS NOT NULL
    THEN pi.id END ASC,
CASE
	WHEN (sqlc.narg(order_by_low_price),sqlc.narg(category_id),sqlc.narg(brand_id)) IS NOT NULL
	THEN pi.product_id
    END ASC,
CASE WHEN sqlc.narg(order_by_low_price) IS NULL
	THEN pi.id END DESC,
    CASE
	WHEN sqlc.narg(order_by_low_price) IS NULL AND (sqlc.narg(category_id),sqlc.narg(brand_id)) IS NOT NULL
	THEN pi.product_id
    END DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsNextPage :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id 
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id  
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
WHERE
(
    CASE
    WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
	THEN (
        (pi.price < sqlc.narg(price)) OR
         (pi.price = sqlc.narg(price) AND  pi.id < sqlc.arg(product_item_id)) OR
         CASE WHEN (sqlc.narg(category_id)::INTEGER,sqlc.narg(brand_id)::INTEGER) IS NOT NULL
         THEN (pi.price = sqlc.narg(price) AND  pi.id = sqlc.arg(product_item_id) AND pi.product_id < sqlc.arg(product_id))
         ELSE 1=1
         END
         )
    WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
	THEN (
        (pi.price > sqlc.narg(price)) OR
         (pi.price = sqlc.narg(price) AND  pi.id > sqlc.arg(product_item_id)) OR
         CASE WHEN (sqlc.narg(category_id)::INTEGER,sqlc.narg(brand_id)::INTEGER) IS NOT NULL
         THEN (pi.price = sqlc.narg(price) AND  pi.id = sqlc.arg(product_item_id) AND pi.product_id > sqlc.arg(product_id))
         ELSE 1=1
         END
         )
	ELSE pi.id < sqlc.arg(product_item_id) AND
        CASE WHEN (sqlc.narg(category_id)::INTEGER,sqlc.narg(brand_id)::INTEGER) IS NOT NULL
        THEN pi.product_id < sqlc.arg(product_id)
        ELSE 1=1
        END
END ) AND
pi.active = TRUE AND
p.active =TRUE AND
CASE
WHEN COALESCE(sqlc.narg(is_promoted), FALSE) = TRUE
THEN 
((pp.active = TRUE
AND ppromo.active = TRUE
AND ppromo.start_date <= CURRENT_DATE 
AND ppromo.end_date >= CURRENT_DATE)
OR (cp.active = TRUE
AND cpromo.active =TRUE 
AND cpromo.start_date <= CURRENT_DATE 
AND cpromo.end_date >= CURRENT_DATE)
OR (bp.active = TRUE
AND bpromo.active =TRUE 
AND bpromo.start_date <= CURRENT_DATE 
AND bpromo.end_date >= CURRENT_DATE))=TRUE 
ELSE TRUE
END 
AND
CASE
WHEN COALESCE(sqlc.narg(is_new), FALSE) = TRUE
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
	WHEN COALESCE(sqlc.narg(order_by_low_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END ASC,
CASE
	WHEN COALESCE(sqlc.narg(order_by_high_price), FALSE) = TRUE
		THEN pi.price
	ELSE ''
END DESC,
CASE
    WHEN sqlc.narg(order_by_low_price)::BOOLEAN IS NOT NULL
    THEN pi.id END ASC,
CASE
	WHEN (sqlc.narg(order_by_low_price)::BOOLEAN,sqlc.narg(category_id)::INTEGER,sqlc.narg(brand_id)::INTEGER) IS NOT NULL
	THEN pi.product_id
    END ASC,
CASE WHEN sqlc.narg(order_by_low_price)::BOOLEAN IS NULL
	THEN pi.id END DESC,
    CASE
	WHEN sqlc.narg(order_by_low_price)::BOOLEAN IS NULL AND (sqlc.narg(category_id)::INTEGER,sqlc.narg(brand_id)::INTEGER) IS NOT NULL
	THEN pi.product_id
    END DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: SearchProductItemsOld :many
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
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id 
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id  
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
WHERE 
pi.active = TRUE AND
p.active =TRUE AND
p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
ORDER BY 
pi.id DESC,
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: SearchProductItemsNextPageOld :many
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
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_promotion" AS pp ON pp.product_id = p.id 
LEFT JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id  
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
LEFT JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
LEFT JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
LEFT JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
WHERE 
pi.id < sqlc.arg(product_item_id) AND
p.id < sqlc.arg(product_id) AND
pi.active = TRUE AND
p.active =TRUE AND
p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
ORDER BY 
pi.id DESC,
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
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

-- name: ListProductItemsWithPromotions :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
INNER JOIN "product_promotion" AS pp ON pp.product_id = p.id 
INNER JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
 
WHERE 
p.id = sqlc.arg(product_id) AND
pi.active = TRUE AND
p.active = TRUE AND
((pp.active = TRUE
AND ppromo.active = TRUE
AND ppromo.start_date <= CURRENT_DATE 
AND ppromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsWithPromotionsNextPage :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 ppromo.id AS product_promo_id, ppromo.name AS product_promo_name, ppromo.description AS product_promo_description,
 ppromo.discount_rate AS product_promo_discount_rate, COALESCE(ppromo.active, FALSE) AS product_promo_active,
 ppromo.start_date AS product_promo_start_date, ppromo.end_date AS product_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
INNER JOIN "product_promotion" AS pp ON pp.product_id = p.id 
INNER JOIN "promotion" AS ppromo ON ppromo.id = pp.promotion_id 
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
 
WHERE 
p.id = sqlc.arg(product_id) AND
pi.id < sqlc.arg(product_item_id) AND
pi.active = TRUE AND
p.active =TRUE AND
((pp.active = TRUE
AND ppromo.active = TRUE
AND ppromo.start_date <= CURRENT_DATE 
AND ppromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsWithBrandPromotions :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
INNER JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
INNER JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
 
WHERE 
pb.id = sqlc.arg(brand_id) AND
pi.active = TRUE AND
p.active =TRUE AND
((bp.active = TRUE
AND bpromo.active = TRUE
AND bpromo.start_date <= CURRENT_DATE 
AND bpromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsWithBrandPromotionsNextPage :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
 bpromo.id AS brand_promo_id, bpromo.name AS brand_promo_name, bpromo.description AS brand_promo_description,
 bpromo.discount_rate AS brand_promo_discount_rate, COALESCE(bpromo.active, FALSE) AS brand_promo_active,
 bpromo.start_date AS brand_promo_start_date, bpromo.end_date AS brand_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
INNER JOIN "brand_promotion" AS bp ON bp.brand_id = p.brand_id 
INNER JOIN "promotion" AS bpromo ON bpromo.id = bp.promotion_id  
 
WHERE 
pb.id = sqlc.arg(brand_id) AND
pi.id < sqlc.arg(product_item_id) AND
p.id < sqlc.arg(product_id) AND
pi.active = TRUE AND
p.active =TRUE AND
((bp.active = TRUE
AND bpromo.active = TRUE
AND bpromo.start_date <= CURRENT_DATE 
AND bpromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsWithCategoryPromotions :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
INNER JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
INNER JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id 
 
WHERE 
pc.id = sqlc.arg(category_id) AND
pi.active = TRUE AND
p.active =TRUE AND
((cp.active = TRUE
AND cpromo.active = TRUE
AND cpromo.start_date <= CURRENT_DATE 
AND cpromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductItemsWithCategoryPromotionsNextPage :many
WITH t1 AS(
SELECT 
 pi.*, p.name, p.description, p.category_id, p.brand_id, pc.category_name, pc.parent_category_id,
 pc.category_image, pb.brand_name, pb.brand_image, p.active AS parent_product_active, ps.size_value,
 pimg.product_image_1, pimg.product_image_2, pimg.product_image_3, pclr.color_value,
cpromo.id AS category_promo_id, cpromo.name AS category_promo_name, cpromo.description AS category_promo_description,
 cpromo.discount_rate AS category_promo_discount_rate, COALESCE(cpromo.active, FALSE) AS category_promo_active,
 cpromo.start_date AS category_promo_start_date, cpromo.end_date AS category_promo_end_date
FROM "product_item" AS pi
INNER JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "product_size" AS ps ON ps.id = pi.size_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_color" AS pclr ON pclr.id = pi.color_id
LEFT JOIN "product_category" AS pc ON pc.id = p.category_id
INNER JOIN "category_promotion" AS cp ON cp.category_id = p.category_id 
INNER JOIN "promotion" AS cpromo ON cpromo.id = cp.promotion_id  
LEFT JOIN "product_brand" AS pb ON pb.id = p.brand_id
 
WHERE 
pc.id = sqlc.arg(category_id) AND
pi.id < sqlc.arg(product_item_id) AND
p.id < sqlc.arg(product_id) AND
pi.active = TRUE AND
p.active =TRUE AND
((cp.active = TRUE
AND cpromo.active = TRUE
AND cpromo.start_date <= CURRENT_DATE 
AND cpromo.end_date >= CURRENT_DATE))
ORDER BY 
pi.id DESC,
p.id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: GetActiveProductItems :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(pi.id) FROM product_item AS pi
JOIN product AS p ON p.id = pi.product_id
WHERE p.active = TRUE AND pi.active = TRUE
AND EXISTS(SELECT is_admin FROM t1);

-- name: GetTotalProductItems :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(pi.id) FROM product_item AS pi
JOIN product AS p ON p.id = pi.product_id
WHERE EXISTS(SELECT is_admin FROM t1);