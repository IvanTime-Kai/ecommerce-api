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
WHERE shop_id = @shop_id
  AND (@from_date::date = '0001-01-01'::date OR date >= @from_date::date)
  AND (@to_date::date = '0001-01-01'::date OR date <= @to_date::date)
ORDER BY date DESC;
