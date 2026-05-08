-- name: GetUserByID :one
SELECT id, email, password_hash, name, role, email_verified, email_verified_at, stripe_customer_id, created_at, updated_at
FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, role, email_verified, email_verified_at, stripe_customer_id, created_at, updated_at
FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, role)
VALUES ($1, $2, $3, $4)
RETURNING id, email_verified, email_verified_at, created_at, updated_at;

-- name: UpdateUser :exec
UPDATE users SET email = $1, password_hash = $2, name = $3, role = $4, updated_at = NOW()
WHERE id = $5;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, password_hash, name, role, email_verified, email_verified_at, stripe_customer_id, created_at, updated_at
FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: VerifyEmail :exec
UPDATE users SET email_verified = true, email_verified_at = NOW() WHERE id = $1;

-- name: UpdateStripeCustomerID :exec
UPDATE users SET stripe_customer_id = $1, updated_at = NOW() WHERE id = $2;
