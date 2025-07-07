CREATE IF NOT EXISTS TABLE appointments (
    appointment_id SERIAL PRIMARY KEY,
    schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
    patient_id INTEGER NOT NULL REFERENCES patients(patient_id),
    ticket_id INTEGER UNIQUE REFERENCES tickets(ticket_id), 
    diagnosis TEXT,
    recommendations TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
