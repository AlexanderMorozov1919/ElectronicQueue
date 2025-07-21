CREATE TABLE IF NOT EXISTS doctors (
    doctor_id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    login VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255),
    specialization VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);