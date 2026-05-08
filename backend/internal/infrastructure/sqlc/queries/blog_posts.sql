-- name: GetBlogPostByID :one
SELECT id, title, slug, content, excerpt, cover_image, status, author_id, created_at, published_at, metadata
FROM blog_posts WHERE id = $1;

-- name: GetBlogPostBySlug :one
SELECT id, title, slug, content, excerpt, cover_image, status, author_id, created_at, published_at, metadata
FROM blog_posts WHERE slug = $1;

-- name: ListBlogPosts :many
SELECT id, title, slug, content, excerpt, cover_image, status, author_id, created_at, published_at, metadata
FROM blog_posts
WHERE ($1::text = '' OR status = $1::text)
ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountBlogPosts :one
SELECT COUNT(*) FROM blog_posts WHERE ($1::text = '' OR status = $1::text);

-- name: CreateBlogPost :one
INSERT INTO blog_posts (title, slug, content, excerpt, cover_image, status, author_id, published_at, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at;

-- name: UpdateBlogPost :exec
UPDATE blog_posts
SET title = $1, slug = $2, content = $3, excerpt = $4, cover_image = $5,
    status = $6, author_id = $7, published_at = $8, metadata = $9
WHERE id = $10;

-- name: DeleteBlogPost :exec
DELETE FROM blog_posts WHERE id = $1;
