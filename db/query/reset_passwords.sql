-- name: CreateResetPassword :one
INSERT INTO "reset_passwords" (
    user_id,
    -- email,
    secret_code
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetResetPassword :one
SELECT * FROM "reset_passwords"
WHERE id = $1 LIMIT 1;

-- name: GetResetPasswordUserIDByID :one
SELECT user_id FROM "reset_passwords"
WHERE
    id = @id
    AND secret_code = @secret_code
    AND is_used = FALSE
    AND expired_at > now()
LIMIT 1;

-- name: GetResetPasswordsByEmail :one
SELECT u.email, u.username, u.is_blocked AS is_blocked_user, u.is_email_verified, 
rp.* FROM "reset_passwords" AS rp
JOIN "user" AS u ON u.id = rp.user_id
WHERE u.email = $1
ORDER BY rp.created_at DESC
LIMIT 1;

-- name: UpdateResetPassword :one
UPDATE "reset_passwords"
SET
    is_used = TRUE
WHERE
    id = @id
    AND secret_code = @secret_code
    AND is_used = FALSE
    AND expired_at > now()
RETURNING *;

-- name: GetLastUsedResetPassword :one
SELECT rp.* FROM "reset_passwords" AS rp
JOIN "user" AS u ON u.id = rp.user_id
WHERE u.email = $1
-- AND secret_code = $2
AND is_used = TRUE
AND expired_at > now()
ORDER BY rp.updated_at DESC, rp.created_at DESC
LIMIT 1;