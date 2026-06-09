-- name: CreateProduct :one
INSERT INTO
  products (id, shop_id, name, description, price, stock)
VALUES
  ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetProductByID :one
SELECT
  *
FROM
  products
WHERE
  id = $1;

-- name: GetProductsByShopID :many
SELECT
  *
FROM
  products
WHERE
  shop_id = $1
  AND status = 'active';

-- name: UpdateProduct :one
UPDATE
  products
SET
  name = $2,
  description = $3,
  status = $4,
  updated_at = NOW()
WHERE
  id = $1 RETURNING *;

-- name: DeleteProduct :exec
UPDATE
  products
SET
  status = 'deleted',
  updated_at = NOW()
WHERE
  id = $1;

-- name: GetProductByIDAndShopOwner :one
SELECT
  p.*
FROM
  products p
  JOIN shops s ON s.id = p.shop_id
WHERE
  p.id = $1
  AND s.owner_id = $2;

-- name: GetProductsForOrder :many
SELECT
  id,
  name,
  price,
  stock,
  shop_id
FROM
  products
WHERE
  id = ANY($1::uuid[])
  AND shop_id = $2
  AND status = 'active'
ORDER BY
  id FOR
UPDATE
;

-- name: DeductProductStock :one
UPDATE
  products
SET
  stock = stock - $2,
  updated_at = NOW()
WHERE
  id = $1
  AND stock >= $2 RETURNING *;

-- name: GetAllProducts :many
SELECT * FROM products;

-- name: SearchProducts :many
SELECT
  p.id,
  p.shop_id,
  p.category_id,
  p.name,
  p.description,
  p.price,
  p.stock,
  p.status,
  p.created_at,
  p.updated_at,
  s.name AS shop_name,
  c.name AS category_name
FROM products p
JOIN shops s ON p.shop_id = s.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.status = 'active'
  AND ($1::text = '' OR p.search_vector @@ plainto_tsquery('english', $1))
  AND ($2::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR p.category_id = $2)
  AND ($3::numeric = 0 OR p.price >= $3)
  AND ($4::numeric = 0 OR p.price <= $4)
  AND ($5::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR p.id < $5)
ORDER BY p.id DESC
LIMIT $6;
