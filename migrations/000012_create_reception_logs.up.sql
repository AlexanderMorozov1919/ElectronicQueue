CREATE TABLE IF NOT EXISTS reception_logs (
    log_id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL,
    registrar_id INTEGER,
    window_number INTEGER NOT NULL,
    called_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration INTERVAL,
    CONSTRAINT fk_ticket
        FOREIGN KEY(ticket_id) 
        REFERENCES tickets(ticket_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_registrar
        FOREIGN KEY(registrar_id)
        REFERENCES registrars(registrar_id)
        ON DELETE SET NULL
);

-- Индекс для быстрого поиска активных логов по талону
CREATE INDEX IF NOT EXISTS idx_reception_logs_active ON reception_logs (ticket_id) WHERE completed_at IS NULL;