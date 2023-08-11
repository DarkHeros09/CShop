-- name: CreateHomePageTextBanner :one
INSERT INTO "home_page_text_banner" (
  name,
  description
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetHomePageTextBanner :one
SELECT * FROM "home_page_text_banner"
WHERE id = $1 LIMIT 1;

-- name: ListHomePageTextBanners :many
SELECT * FROM "home_page_text_banner";

-- name: UpdateHomePageTextBanner :one
UPDATE "home_page_text_banner"
SET 
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteHomePageTextBanner :exec
DELETE FROM "home_page_text_banner"
WHERE id = $1;