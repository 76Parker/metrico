CREATE TABLE IF NOT EXISTS metric.metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE,
    type text NOT NULL, -- не создаю собственный тип тк в будущем может понадобиться расширить список типов
    value DOUBLE PRECISION,
    delta BIGINT,
    hash text,
    created_at TIMESTAMPz NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPz NOT NULL DEFAULT NOW()
);
