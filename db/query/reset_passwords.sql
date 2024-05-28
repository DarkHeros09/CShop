-- name: CreateResetPassword :one
INSERT INTO "reset_passwords" (
    user_id,
    email,
    secret_code
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetResetPasswordUserIDByID :one
SELECT user_id FROM "reset_passwords"
WHERE
    id = @id
    AND secret_code = @secret_code
    AND is_used = FALSE
    AND expired_at > now()
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