-- name: UpsertRevenueSummary :exec
INSERT INTO revenue_summary (shop_id, date, total, order_count)
VALUES ($1, $2, $3, 1)
ON CONFLICT (shop_id, date)
DO UPDATE SET
  total = revenue_summary.total + EXCLUDED.total,
  order_count = revenue_summary.order_count + 1;

-- name: GetRevenueSummary :many
SELECT date, total, order_count
FROM revenue_summary
WHERE shop_id = $1
ORDER BY date DESC;