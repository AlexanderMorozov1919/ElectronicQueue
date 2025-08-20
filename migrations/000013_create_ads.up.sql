CREATE TABLE IF NOT EXISTS ads (
    id SERIAL PRIMARY KEY,
    picture BYTEA,
    video BYTEA,
    duration_sec INTEGER, -- Убрали NOT NULL и DEFAULT
    repeat_count INTEGER, -- Убрали NOT NULL и DEFAULT
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    reception_on BOOLEAN NOT NULL DEFAULT TRUE,
    schedule_on BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_media_fields CHECK (
        (picture IS NOT NULL AND video IS NULL AND duration_sec IS NOT NULL AND repeat_count IS NULL) OR
        (video IS NOT NULL AND picture IS NULL AND repeat_count IS NOT NULL AND duration_sec IS NULL)
    )
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