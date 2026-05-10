-- name: GetProductByID :one
SELECT id, name, description, price, stock, is_pre_sale, pre_sale_available_at, images, category_id, is_active, created_at, updated_at
FROM products WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price, stock, is_pre_sale, pre_sale_available_at, images, category_id, is_active, created_at, updated_at
FROM products
WHERE (sqlc.narg(category_id)::uuid IS NULL OR category_id = sqlc.narg(category_id))
  AND (sqlc.arg(active_only)::boolean = false OR is_active = true)
ORDER BY created_at DESC LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE (sqlc.narg(category_id)::uuid IS NULL OR category_id = sqlc.narg(category_id))
  AND (sqlc.arg(active_only)::boolean = false OR is_active = true);

-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, is_pre_sale, pre_sale_available_at, images, category_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at, updated_at;

-- name: UpdateProduct :exec
UPDATE products
SET name = $1, description = $2, price = $3, stock = $4, is_pre_sale = $5,
    pre_sale_available_at = $6, images = $7, category_id = $8, is_active = $9
WHERE id = $10;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;
