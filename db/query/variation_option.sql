-- name: CreateVariationOption :one
INSERT INTO "variation_option" (
  variation_id,
  value
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetVariationOption :one
SELECT * FROM "variation_option"
WHERE id = $1 LIMIT 1;

-- name: ListVariationOptions :many
SELECT * FROM "variation_option"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateVariationOption :one
UPDATE "variation_option"
SET
variation_id = COALESCE(sqlc.narg(variation_id),variation_id),
value = COALESCE(sqlc.narg(value),value)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteVariationOption :exec
DELETE FROM "variation_option"
WHERE id = $1;