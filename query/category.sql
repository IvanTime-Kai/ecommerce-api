-- name: GetCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL
ORDER BY name;

-- name: GetSubcategories :many
SELECT * FROM categories
WHERE parent_id = $1
ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (id, parent_id, name, slug)
VALUES ($1, $2, $3, $4)
RETURNING *;
