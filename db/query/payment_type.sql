-- name: CreatePaymentType :one
INSERT INTO "payment_type" (
  value
) VALUES (
  $1
) 
ON CONFLICT(value) DO UPDATE SET value = $1
RETURNING *;

-- name: AdminCreatePaymentType :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "payment_type" (
  value,
  is_active
)
SELECT sqlc.arg(value), sqlc.arg(is_active) FROM t1
WHERE is_admin=1
ON CONFLICT(value) DO UPDATE SET value = sqlc.arg(value)
RETURNING *;

-- name: GetPaymentType :one
SELECT * FROM "payment_type"
WHERE id = $1 LIMIT 1;

-- name: ListPaymentTypes :many
SELECT * FROM "payment_type";
-- ORDER BY id
-- LIMIT $1
-- OFFSET $2;

-- name: AdminListPaymentTypes :many
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT * FROM "payment_type"
WHERE (SELECT is_admin FROM t1) = 1;

-- name: AdminUpdatePaymentType :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "payment_type"
SET 
value = COALESCE(sqlc.narg(value),value),
is_active = COALESCE(sqlc.narg(is_active),is_active)
WHERE "payment_type".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: AdminDeletePaymentType :exec
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "payment_type"
WHERE "payment_type".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1;

-- name: UpdatePaymentType :one
UPDATE "payment_type"
SET 
value = COALESCE(sqlc.narg(value),value),
is_active = COALESCE(sqlc.narg(is_active),is_active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePaymentType :exec
DELETE FROM "payment_type"
WHERE id = $1;