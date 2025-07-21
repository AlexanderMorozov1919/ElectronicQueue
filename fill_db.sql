-- =================================================================
-- ==       СКРИПТ ЗАПОЛНЕНИЯ БАЗЫ ДАННЫХ ТЕСТОВЫМИ ДАННЫМИ       ==
-- =================================================================

-- -----------------------------------------------------------------
-- --                        1. УСЛУГИ                            --
-- -----------------------------------------------------------------
INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться к врачу', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Другой вопрос', 'D')
ON CONFLICT (service_id) DO NOTHING;


-- -----------------------------------------------------------------
-- --                         2. ВРАЧИ                            --
-- -----------------------------------------------------------------
INSERT INTO doctors (doctor_id, full_name, specialization, is_active) VALUES
(1, 'Иванов Иван Иванович', 'Терапевт', TRUE),
(2, 'Петров Петр Петрович', 'Хирург', TRUE),
(3, 'Смирнова Мария Викторовна', 'Кардиолог', TRUE),
(4, 'Кузнецова Ольга Дмитриевна', 'Невролог', TRUE)
ON CONFLICT (doctor_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  specialization = EXCLUDED.specialization,
  is_active = EXCLUDED.is_active;


-- -----------------------------------------------------------------
-- --                      3. РЕГИСТРАТОРЫ                        --
-- -----------------------------------------------------------------
-- Пароли: 'admin1' и 'admin2'
INSERT INTO registrars (registrar_id, full_name, login, password_hash, is_active) VALUES
(1, 'Петрова Анна Сергеевна', 'admin1', '$2a$10$Bpqg4mfUfFNLe09MC6QvveFNcY80VAiLSjpOOSBqAnV7avNJo5eEi', TRUE),
(2, 'Сидорова Елена Игоревна', 'admin2', '$2a$10$l3Lt5ogEKuQ1.PoSilqWQ.12ymyxTcpWTOBBBKJE6grMJ2emeFPcy', TRUE)
ON CONFLICT (registrar_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  login = EXCLUDED.login,
  password_hash = EXCLUDED.password_hash,
  is_active = EXCLUDED.is_active;


-- -----------------------------------------------------------------
-- --                        4. ПАЦИЕНТЫ                          --
-- -----------------------------------------------------------------
INSERT INTO patients (patient_id, passport_series, passport_number, oms_number, full_name, birth_date, phone) VALUES
(1, '4510', '123456', '1234567890123456', 'Андреев Андрей Андреевич', '1980-05-15', '+79112223344'),
(2, '4511', '654321', '2345678901234567', 'Борисова Борислава Борисовна', '1992-11-20', '+79213334455'),
(3, '4512', '789012', '3456789012345678', 'Васильев Василий Васильевич', '1975-02-10', '+79314445566'),
(4, '4513', '210987', '4567890123456789', 'Григорьева Галина Григорьевна', '2001-08-30', '+79515556677'),
(5, '4514', '345678', '5678901234567890', 'Дмитриев Дмитрий Дмитриевич', '1988-12-01', '+79616667788')
ON CONFLICT (patient_id) DO UPDATE SET
  passport_series = EXCLUDED.passport_series,
  passport_number = EXCLUDED.passport_number,
  oms_number = EXCLUDED.oms_number,
  full_name = EXCLUDED.full_name,
  birth_date = EXCLUDED.birth_date,
  phone = EXCLUDED.phone;


-- =================================================================
-- ==                   5. РАСПИСАНИЯ ВРАЧЕЙ (июль 2025)          ==
-- =================================================================
DELETE FROM schedules WHERE doctor_id IN (1, 2, 3, 4) AND date_trunc('month', date) = '2025-07-01';

-- Расписание для Врача №1: Иванов Иван Иванович (Терапевт)
INSERT INTO schedules (doctor_id, date, start_time, end_time, is_available)
SELECT
    1 AS doctor_id, day::date, ts::time AS start_time, (ts + '1 hour'::interval)::time AS end_time, true AS is_available
FROM generate_series('2025-07-01'::timestamp, '2025-07-31'::timestamp, '1 day'::interval) AS day
CROSS JOIN generate_series('2025-01-01 08:00'::timestamp, '2025-01-01 15:00'::timestamp, '1 hour'::interval) AS ts
WHERE EXTRACT(isodow FROM day) BETWEEN 1 AND 5;

-- Расписание для Врача №2: Петров Петр Петрович (Хирург)
INSERT INTO schedules (doctor_id, date, start_time, end_time, is_available)
SELECT
    2 AS doctor_id, day::date, ts::time AS start_time, (ts + '1 hour'::interval)::time AS end_time, true AS is_available
FROM (SELECT day, ROW_NUMBER() OVER (ORDER BY day) as rn FROM generate_series('2025-07-01'::timestamp, '2025-07-31'::timestamp, '1 day'::interval) AS day) AS numbered_days
CROSS JOIN generate_series('2025-01-01 09:00'::timestamp, '2025-01-01 17:00'::timestamp, '1 hour'::interval) AS ts
WHERE floor((rn - 1) / 2)::int % 2 = 0;

-- Расписание для Врача №3: Смирнова Мария Викторовна (Кардиолог)
INSERT INTO schedules (doctor_id, date, start_time, end_time, is_available)
SELECT
    3 AS doctor_id, day::date, ts::time AS start_time, (ts + '30 minutes'::interval)::time AS end_time, true AS is_available
FROM generate_series('2025-07-01'::timestamp, '2025-07-31'::timestamp, '1 day'::interval) AS day
CROSS JOIN generate_series('2025-01-01 09:00'::timestamp, '2025-01-01 12:30'::timestamp, '30 minutes'::interval) AS ts
WHERE EXTRACT(isodow FROM day) IN (1, 3, 5);

-- Расписание для Врача №4: Кузнецова Ольга Дмитриевна (Невролог)
INSERT INTO schedules (doctor_id, date, start_time, end_time, is_available)
SELECT
    4 AS doctor_id, day::date, ts::time AS start_time, (ts + '1 hour'::interval)::time AS end_time, true AS is_available
FROM generate_series('2025-07-01'::timestamp, '2025-07-31'::timestamp, '1 day'::interval) AS day
CROSS JOIN generate_series('2025-01-01 10:00'::timestamp, '2025-01-01 16:00'::timestamp, '1 hour'::interval) AS ts
WHERE EXTRACT(isodow FROM day) IN (2, 4);

INSERT INTO schedules (doctor_id, date, start_time, end_time, is_available)
SELECT
    4 AS doctor_id, day::date, ts::time AS start_time, (ts + '1 hour'::interval)::time AS end_time, true AS is_available
FROM generate_series('2025-07-01'::timestamp, '2025-07-31'::timestamp, '1 day'::interval) AS day
CROSS JOIN generate_series('2025-01-01 09:00'::timestamp, '2025-01-01 11:00'::timestamp, '1 hour'::interval) AS ts
WHERE EXTRACT(isodow FROM day) = 6;