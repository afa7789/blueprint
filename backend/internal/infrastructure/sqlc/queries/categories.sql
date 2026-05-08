-- name: ListCategories :many
SELECT id, name, slug, parent_id FROM categories ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (name, slug, parent_id) VALUES ($1, $2, $3) RETURNING id;

-- name: UpdateCategory :exec
UPDATE categories SET name = $1, slug = $2, parent_id = $3 WHERE id = $4;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;
