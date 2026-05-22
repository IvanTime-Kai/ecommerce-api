-- name: CreateOutboxEvent :exec
INSERT INTO outbox (id, event_type, payload)
VALUES ($1, $2, $3);

-- name: GetPendingOutboxEvents :many
SELECT * FROM outbox
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT 10;

-- name: MarkOutboxEventProcessed :exec
UPDATE outbox
SET status = 'processed', processed_at = NOW()
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox
SET status = 'failed' WHERE id=$1;

