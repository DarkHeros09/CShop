-- name: CreateHomePageTextBanner :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "home_page_text_banner" (
  name,
  description,
  active
) 
SELECT sqlc.arg(name),
 sqlc.arg(description),
 sqlc.arg(active)
FROM t1
WHERE EXISTS (SELECT 1 FROM t1)
RETURNING *;

-- name: GetHomePageTextBanner :one
SELECT * FROM "home_page_text_banner"
WHERE id = $1 LIMIT 1;

-- name: ListHomePageTextBanners :many
SELECT * FROM "home_page_text_banner"
WHERE active = TRUE
ORDER BY created_at DESC, updated_at DESC
LIMIT 5;

-- name: UpdateHomePageTextBanner :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "home_page_text_banner"
SET 
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description),
active = COALESCE(sqlc.narg(active),active)
WHERE "home_page_text_banner".id = sqlc.arg(id)
AND EXISTS (SELECT 1 FROM t1)
RETURNING *;

-- name: DeleteHomePageTextBanner :exec
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "home_page_text_banner"
WHERE "home_page_text_banner".id = sqlc.arg(id)
AND EXISTS (SELECT 1 FROM t1);