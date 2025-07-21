ALTER TABLE appointments
ADD COLUMN ticket_id INTEGER REFERENCES tickets(ticket_id) ON DELETE SET NULL;