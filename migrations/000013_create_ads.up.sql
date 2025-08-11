CREATE TABLE IF NOT EXISTS ads (
    id SERIAL PRIMARY KEY,
    picture BYTEA, -- Поле для изображений, теперь может быть NULL
    video BYTEA,   -- Новое поле для видео в формате MP4
    duration_sec INTEGER NOT NULL DEFAULT 5, -- Длительность показа для изображений
    repeat_count INTEGER NOT NULL DEFAULT 1, -- Количество повторов для видео
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    reception_on BOOLEAN NOT NULL DEFAULT TRUE,
    schedule_on BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON ads
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();