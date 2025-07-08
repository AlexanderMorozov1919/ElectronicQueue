CREATE TABLE tickets (
    ticket_id SERIAL PRIMARY KEY,
    ticket_number VARCHAR(20) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL CHECK (status IN (
        'ожидает',         -- Создан, ждет вызова
        'приглашен',       -- Вызван регистратором
        'на_приеме',       -- Врач начал прием
        'завершен',        -- Прием окончен
        'подойти_к_окну',  -- Подойти к окну (регистратор)
        'зарегистрирован'  -- Зарегистрирован (отправлен к врачу)
    )),
    window_number INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    called_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
