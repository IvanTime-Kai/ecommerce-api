CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE INDEX idx_orders_shop_id ON orders(shop_id);

CREATE INDEX idx_orders_shop_active ON orders(shop_id, status)
WHERE
  status IN ('pending', 'confirmed', 'shipping');