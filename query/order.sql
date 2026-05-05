-- name: CreateOrder :one
INSERT INTO
  orders (
    id,
    user_id,
    shop_id,
    total_amount,
    shipping_full_name,
    shipping_phone,
    shipping_province,
    shipping_district,
    shipping_ward,
    shipping_street
  )
VALUES
  (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
  ) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO
  order_items (
    id,
    order_id,
    product_id,
    product_name,
    quantity,
    price
  )
VALUES
  ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetOrderByID :one
SELECT
  *
FROM
  orders
WHERE
  id = $1;

-- name: GetOrdersByUserID :many
SELECT
  *
FROM
  orders
WHERE
  user_id = $1;

-- name: GetOrdersByShopID :many
SELECT
  *
FROM
  orders
WHERE
  shop_id = $1;

-- name: UpdateOrderStatus :one
UPDATE
  orders
SET
  status = $2,
  updated_at = NOW()
WHERE
  id = $1 RETURNING *;