CREATE TABLE IF NOT EXISTS administrators (
    administrator_id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    login VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255)
);