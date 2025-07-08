CREATE IF NOT EXISTS TABLE doctors (
    doctor_id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    specialization VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);