CREATE TABLE reviews (
  id UUID PRIMARY KEY,
  order_id UUID NOT NULL REFERENCES orders(id),
  product_id UUID NOT NULL REFERENCES products(id),
  user_id UUID NOT NULL REFERENCES users(id),
  rating INT NOT NULL CHECK (
    rating >= 1
    AND rating <= 5
  ),
  comment TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(order_id, product_id)
);

CREATE INDEX idx_reviews_product_id ON reviews (product_id);

CREATE INDEX idx_reviews_user_id ON reviews (user_id);