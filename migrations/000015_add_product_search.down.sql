DROP TRIGGER IF EXISTS trg_product_search_vector ON products;

DROP FUNCTION IF EXISTS update_product_search_vector;

DROP INDEX IF EXISTS idx_products_search;

DROP INDEX IF EXISTS idx_products_category_id;

DROP INDEX IF EXISTS idx_products_price;

ALTER TABLE
  products DROP COLUMN IF EXISTS search_vector;

ALTER TABLE
  products DROP COLUMN IF EXISTS category_id;