CREATE TABLE IF NOT EXISTS tickets (
    ticket_id SERIAL PRIMARY KEY,
    ticket_number VARCHAR(20) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL CHECK (status IN (
        'ожидает',
        'приглашен',
        'на_приеме',
        'завершен',
        'зарегистрирован'
    )),
    service_type VARCHAR(50),
    window_number INTEGER,
    qr_code BYTEA,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    called_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);