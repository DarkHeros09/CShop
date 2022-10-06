-- name: CreateVariation :one
INSERT INTO "variation" (
  category_id,
  name
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetVariation :one
SELECT * FROM "variation"
WHERE id = $1 LIMIT 1;

-- name: ListVariations :many
SELECT * FROM "variation"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateVariation :one
UPDATE "variation"
SET
name = COALESCE(sqlc.narg(name),name),
category_id = COALESCE(sqlc.narg(category_id),category_id)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteVariation :exec
DELETE FROM "variation"
WHERE id = $1;