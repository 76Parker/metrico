-- name: Upsert :exec
INSERT INTO metric.metrics (name, type, value, delta, hash)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (name)
DO UPDATE SET
value = EXCLUDED.value, delta = EXCLUDED.delta, updated_at = NOW();
