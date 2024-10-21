-- name: CreateAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "app_policy" 
( "policy" )
SELECT sqlc.arg(policy) FROM t1
WHERE EXISTS (SELECT 1 FROM t1)
RETURNING *;

-- name: GetAppPolicy :one
SELECT * FROM "app_policy"
LIMIT 1;

-- name: UpdateAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "app_policy"
SET 
policy = COALESCE(sqlc.narg(policy),policy),
updated_at = NOW()
WHERE "app_policy".id = sqlc.arg(id)
AND EXISTS (SELECT 1 FROM t1)
RETURNING *;

-- name: DeleteAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "app_policy"
WHERE "app_policy".id = sqlc.arg(id)
AND EXISTS (SELECT 1 FROM t1)
RETURNING *;