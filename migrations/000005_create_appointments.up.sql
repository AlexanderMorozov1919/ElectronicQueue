CREATE TABLE IF NOT EXISTS appointments (
    appointment_id SERIAL PRIMARY KEY,
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
    ticket_id INTEGER REFERENCES tickets(ticket_id) ON DELETE SET NULL,
    patient_id INTEGER REFERENCES patients(patient_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);