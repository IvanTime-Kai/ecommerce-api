-- name: CreateReview :one
INSERT INTO reviews (id, order_id, product_id, user_id, rating, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetReviewsByProductID :many
SELECT r.*, u.full_name as user_name
FROM reviews r
JOIN users u ON r.user_id = u.id
WHERE r.product_id = $1
  AND ($2::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR r.id < $2)
ORDER BY r.id DESC
LIMIT $3;

-- name: GetReviewByOrderAndProduct :one
SELECT * FROM reviews
WHERE order_id = $1 AND product_id = $2;