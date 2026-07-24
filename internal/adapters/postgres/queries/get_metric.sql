-- name: GetByName :one
SELECT * FROM metric.metrics WHERE name = $1;
