CREATE TABLE IF NOT EXISTS business_processes (
    process_name VARCHAR(50) PRIMARY KEY,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

-- Заполняем таблицу всеми известными процессами
INSERT INTO business_processes (process_name, is_enabled) VALUES
('terminal', TRUE),
('reception', TRUE),
('registry', TRUE),
('doctor', TRUE),
('queue_doctor', TRUE),
('schedule', TRUE),
('database', TRUE),
('appointment', TRUE)
ON CONFLICT (process_name) DO NOTHING;