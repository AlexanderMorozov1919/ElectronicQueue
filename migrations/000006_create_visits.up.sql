CREATE TABLE visits (
    id integer NOT NULL,
    patient_card_id integer NOT NULL,
    doctor_id integer NOT NULL,
    visit_date timestamp with time zone NOT NULL,
    notes text,
    ticket_id integer
);
