CREATE TABLE tickets (
    id integer NOT NULL,
    queue_number character varying(50) NOT NULL,
    creation_time timestamp with time zone,
    status character varying(50) NOT NULL,
    queue_id integer NOT NULL,
    patient_card_id integer,
    doctor_id integer
);
