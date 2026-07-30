-- name: BatchUpsert :exec
-- BatchUpsert через jsonb_to_recordset
-- Смотрите инструкцию для параметра `BatchUpsertParam` в `internal/adapters/postgres/pgen/custom_params.go`
/*
 * BATCH-операция использует подход JSONB-based batch upsert с использованием `jsonb_to_recordset`.
 * JSON сохраняет nil как null, который PostgreSQL преобразует в SQL NULL, и это позволяет сохранить NULL для полей value или delta.

 * - Почему не UNNEST-based batch upsert подход?
 * При подходе UNNEST-based batch upsert sqlc генерирует массивы как []float64 и []int64,
 * и это не позволяют представить NULL для полей value и delta, а иметь zero-value вместо NULL не будет корретной логикой.

 * - Почему вручную не написать UNNEST batch upsert вручную в коде?
 * В проекте принят единый подход к работе с SQL через sqlc
 * Написание ручного запроса через `pgx` напрямую в коде нарушило бы единую подход к работе с SQL-запросами
 */
INSERT INTO metric.metrics (name, type, value, delta)
SELECT
    input.name,
    input.type,
    input.value,
    input.delta
FROM jsonb_to_recordset(sqlc.arg(metrics)::jsonb) AS input(
    name  text,
    type  text,
    value double precision,
    delta bigint
)
ON CONFLICT (name)
DO UPDATE SET
    value      = EXCLUDED.value,
    delta      = COALESCE(EXCLUDED.delta, 0) + COALESCE(metric.metrics.delta, 0),
    updated_at = NOW();
