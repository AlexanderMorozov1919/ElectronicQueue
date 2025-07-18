CREATE TABLE IF NOT EXISTS patients (
    patient_id SERIAL PRIMARY KEY,
    passport_series VARCHAR(4) NOT NULL,
    passport_number VARCHAR(6) NOT NULL,
    oms_number VARCHAR(16) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    birth_date DATE,
    phone VARCHAR(20),
    UNIQUE (passport_series, passport_number)
);
