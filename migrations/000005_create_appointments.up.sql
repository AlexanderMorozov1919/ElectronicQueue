CREATE TABLE IF NOT EXISTS appointments (
    appointment_id SERIAL PRIMARY KEY,
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
    patient_id INTEGER NOT NULL REFERENCES patients(patient_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
