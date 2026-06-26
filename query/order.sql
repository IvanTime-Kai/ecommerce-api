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

-- name: ConfirmOrder :one
UPDATE orders SET status = 'confirmed', updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: ShipOrder :one
UPDATE orders SET status = 'shipping', updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: DeliverOrder :one
UPDATE orders SET status = 'delivered', updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: CancelOrder :one
UPDATE orders SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items WHERE order_id = $1;

-- name: CreateIdempotencyKey :exec
INSERT INTO idempotency_keys (key, response, expires_at)
VALUES ($1, $2, $3);

-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys
WHERE key = $1 AND expires_at > NOW();

-- name: GetOrdersByUserIDWithCursor :many
SELECT * FROM orders
WHERE user_id = $1
  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR id < $2)
ORDER BY id DESC
LIMIT $3;

-- name: GetOrderItemsByOrderIDs :many
SELECT * FROM order_items 
WHERE order_id = ANY(@order_ids::uuid[]);