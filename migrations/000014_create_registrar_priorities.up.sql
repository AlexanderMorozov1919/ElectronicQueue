CREATE TABLE IF NOT EXISTS registrar_category_priorities (
    registrar_id INTEGER NOT NULL REFERENCES registrars(registrar_id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    PRIMARY KEY (registrar_id, service_id)
);