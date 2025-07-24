-- =================================================================
-- ==       СКРИПТ ЗАПОЛНЕНИЯ БАЗЫ ДАННЫХ ТЕСТОВЫМИ ДАННЫМИ       ==
-- ==                 (СЦЕНАРИЙ "РАБОЧИЙ ДЕНЬ")                  ==
-- =================================================================

-- -----------------------------------------------------------------
-- --                     0. ОЧИСТКА ДАННЫХ                       --
-- -----------------------------------------------------------------
TRUNCATE TABLE 
    appointments, 
    tickets, 
    schedules, 
    services, 
    doctors, 
    registrars, 
    patients 
RESTART IDENTITY CASCADE;

-- -----------------------------------------------------------------
-- --                        1. УСЛУГИ                            --
-- -----------------------------------------------------------------
INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться к врачу', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Другой вопрос', 'D');

-- -----------------------------------------------------------------
-- --                      2. РЕГИСТРАТОРЫ                        --
-- -----------------------------------------------------------------
-- Пароли: 'pass1' - 'pass7'
INSERT INTO registrars (window_number, login, password_hash) VALUES
(1, 'admin1', '$2a$10$g3C8q/gOeSAT2uVgFXz8M.Xs4OELKH7gD24P3nciXTfbm4RePxWqG'),
(2, 'admin2', '$2a$10$fnLMINfO.s4.zMr6MooIHuoxiLy1CcFCGajzmH8VlxGw0BvnMB75C'),
(3, 'admin3', '$2a$10$LFWIFiiMooFIOXX2IsW4vuQwvNE3vtUqPLDyvKA8cdO/1YSjKzAsu'),
(4, 'admin4', '$2a$10$NOnwKKyVVyhAZGPgtOUF9uQzbWYSmtVZSHUosreJfbxj/vYL/XmC2'),
(5, 'admin5', '$2a$10$sSIr0.6WFNw2jyVXE2OEbusHqE.nPCbytzskyx2HJ/0TKyWJBqKeO'),
(6, 'admin6', '$2a$10$Ud9Hwrm4vawrae6WpIsRWe1A1wAWkAqdlX65/R5LuFkRJ1w17Qxri'),
(7, 'admin7', '$2a$10$hllZlVYZ0R.kEvq0il3e2eXyctV/3X0li0OT7DeKfJrY9QTQZwTbO');

-- -----------------------------------------------------------------
-- --                         3. ВРАЧИ                            --
-- -----------------------------------------------------------------
-- Пароли: 'pass1' - 'pass7'
INSERT INTO doctors (full_name, specialization, login, password_hash, status) VALUES
('Иванов Иван Иванович', 'Терапевт', 'doctor1', '$2a$10$9S2D6Vr.2Cv2wSest1EwPe2x/wZKW0raBzZ4CyX906iq7vB2cJ8Za', 'активен'),
('Петров Петр Петрович', 'Хирург', 'doctor2', '$2a$10$80G/wVQ/dwtI8TZpHwhKfOMX36bL3y5dPBbLcBdeLEiDNA8Ogg9FC', 'активен'),
('Смирнова Мария Викторовна', 'Кардиолог', 'doctor3', '$2a$10$Otd/PbC3Dhvxo7rVsyHrHerP460E.t4XiWHMkHvStcU4ijG5A6Ap.', 'перерыв'),
('Кузнецова Ольга Дмитриевна', 'Невролог', 'doctor4', '$2a$10$1yAN3/hB8O93vSZjPw/B4O0R0NgWuddnAPy.tDiCTNxWW6rQyzOqW', 'активен'),
('Михайлов Михаил Михайлович', 'Офтальмолог', 'doctor5', '$2a$10$n68/soTxF/YVkR1olmR16u3FwyFoHLxj5IDrscjy.DeGl7pK9w1x.', 'неактивен'),
('Васильева Елена Сергеевна', 'Педиатр', 'doctor6', '$2a$10$ZmSQHlwqr/25oZdSg3Zod.hvvSdcLg0M.8K0b.D5hZBK9BqXLga..', 'активен'),
('Соколов Сергей Александрович', 'ЛОР', 'doctor7', '$2a$10$9dSK.8zXoR0lCfatQ4mBn.2l./3g.JYNbZCUEMZauwD.nFrJ115he', 'активен');

-- -----------------------------------------------------------------
-- --                        4. ПАЦИЕНТЫ                          --
-- -----------------------------------------------------------------
INSERT INTO patients (passport_series, passport_number, oms_number, full_name, birth_date, phone) VALUES
('4510', '123456', '1111111111111111', 'Андреев Андрей Андреевич', '1980-05-15', '+79112223344'),
('4511', '654321', '2222222222222222', 'Борисова Борислава Борисовна', '1992-11-20', '+79213334455'),
('4512', '789012', '3333333333333333', 'Васильев Василий Васильевич', '1975-02-10', '+79314445566'),
('4513', '210987', '4444444444444444', 'Григорьева Галина Григорьевна', '2001-08-30', '+79515556677'),
('4514', '345678', '5555555555555555', 'Дмитриев Дмитрий Дмитриевич', '1988-12-01', '+79616667788'),
('4515', '112233', '6666666666666666', 'Егорова Елизавета Егоровна', '1995-03-25', '+79011112233'),
('4516', '445566', '7777777777777777', 'Железнов Ждан Жанович', '1963-07-12', '+79022223344'),
('4517', '778899', '8888888888888888', 'Зайцева Зинаида Захаровна', '1982-01-18', '+79033334455'),
('4518', '101112', '9999999999999999', 'Константинов Константин Константинович', '1999-09-09', '+79044445566'),
('4519', '131415', '1010101010101010', 'Лебедева Любовь Львовна', '1978-04-04', '+79055556677'),
('4520', '161718', '1212121212121212', 'Морозов Максим Максимович', '2003-06-21', '+79066667788'),
('4521', '192021', '1313131313131313', 'Николаева Надежда Николаевна', '1985-10-11', '+79088889900'),
('4522', '222324', '1414141414141414', 'Орлов Олег Олегович', '1991-05-14', '+79099990011'),
('4523', '252627', '1515151515151515', 'Романова Раиса Романовна', '1968-02-28', '+79811112233'),
('4524', '282930', '1616161616161616', 'Сергеев Станислав Сергеевич', '1977-11-07', '+79822223344');

-- -----------------------------------------------------------------
-- --        5. РАСПИСАНИЕ ВРАЧЕЙ (СЕГОДНЯ + 6 ДНЕЙ ВПЕРЕД)       --
-- -----------------------------------------------------------------
-- Индивидуальные временные границы для каждого врача
WITH doctor_times AS (
    SELECT 1 AS doctor_id, '08:00'::time AS start_time, '19:30'::time AS end_time UNION ALL
    SELECT 2, '09:00', '18:00' UNION ALL
    SELECT 3, '10:00', '17:00' UNION ALL
    SELECT 4, '08:30', '16:30' UNION ALL
    SELECT 5, '12:00', '19:30' UNION ALL
    SELECT 6, '08:00', '15:00' UNION ALL
    SELECT 7, '11:00', '19:00'
)
INSERT INTO schedules (doctor_id, cabinet, date, start_time, end_time)
SELECT
    d.doctor_id,
    (100 + d.doctor_id) AS cabinet,
    d.day::date,
    s.start_time::time,
    (s.start_time + '30 minutes'::interval)::time AS end_time
FROM 
    (SELECT doctor_id, generate_series(CURRENT_DATE, CURRENT_DATE + interval '6 days', '1 day') as day FROM doctors) d
    JOIN doctor_times dt ON d.doctor_id = dt.doctor_id
    CROSS JOIN LATERAL generate_series(
        (CURRENT_DATE + dt.start_time::time),
        (CURRENT_DATE + dt.end_time::time),
        '30 minutes'::interval
    ) AS s(start_time);

-- -----------------------------------------------------------------
-- --                6. ТАЛОНЫ И ЗАПИСИ НА ПРИЕМ                  --
-- -----------------------------------------------------------------
-- Сценарий: Середина рабочего дня, примерно 12:00
-- 6.1 Талоны в статусе "завершен"
INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at, called_at, completed_at) VALUES
('A001', 'завершен', 'make_appointment', 1, NOW() - INTERVAL '3 hour', NOW() - INTERVAL '2 hour 50 minutes', NOW() - INTERVAL '2 hour 45 minutes'),
('C001', 'завершен', 'lab_tests', 2, NOW() - INTERVAL '2 hour 30 minutes', NOW() - INTERVAL '2 hour 20 minutes', NOW() - INTERVAL '2 hour 10 minutes'),
('D001', 'завершен', 'documents', 1, NOW() - INTERVAL '2 hour', NOW() - INTERVAL '1 hour 55 minutes', NOW() - INTERVAL '1 hour 40 minutes');

-- 6.2 Талоны в статусе "приглашен"
INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at, called_at) VALUES
('A002', 'приглашен', 'make_appointment', 1, NOW() - INTERVAL '1 hour 30 minutes', NOW() - INTERVAL '1 minute'),
('B001', 'приглашен', 'confirm_appointment', 3, NOW() - INTERVAL '1 hour 25 minutes', NOW() - INTERVAL '30 seconds');

-- 6.3 Талоны в статусе "ожидает"
INSERT INTO tickets (ticket_number, status, service_type, created_at) VALUES
('C002', 'ожидает', 'lab_tests', NOW() - INTERVAL '1 hour'),
('D002', 'ожидает', 'documents', NOW() - INTERVAL '55 minutes'),
('A003', 'ожидает', 'make_appointment', NOW() - INTERVAL '50 minutes'),
('B002', 'ожидает', 'confirm_appointment', NOW() - INTERVAL '40 minutes'),
('C003', 'ожидает', 'lab_tests', NOW() - INTERVAL '30 minutes'),
('A004', 'ожидает', 'make_appointment', NOW() - INTERVAL '20 minutes'),
('D003', 'ожидает', 'documents', NOW() - INTERVAL '10 minutes'),
('B003', 'ожидает', 'confirm_appointment', NOW() - INTERVAL '5 minutes');

-- 6.4 Талоны и записи, связанные с врачами (ТОЛЬКО НА СЕГОДНЯ)
DO $$
DECLARE
    v_schedule_id INT;
    v_ticket_id INT;
BEGIN
    -- --- ЗАПИСЬ №1: НА ПРИЕМЕ ---
    SELECT schedule_id INTO v_schedule_id FROM schedules WHERE doctor_id = 1 AND date = CURRENT_DATE AND start_time = '11:30:00' LIMIT 1;
    INSERT INTO tickets (ticket_number, status, service_type, created_at, started_at) VALUES ('B010', 'на_приеме', 'confirm_appointment', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '5 minutes') RETURNING ticket_id INTO v_ticket_id;
    INSERT INTO appointments (schedule_id, patient_id, ticket_id) VALUES (v_schedule_id, 1, v_ticket_id);
    UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;
    
    -- --- ЗАПИСЬ №2: ЗАРЕГИСТРИРОВАН (СЛЕДУЮЩИЙ) ---
    SELECT schedule_id INTO v_schedule_id FROM schedules WHERE doctor_id = 2 AND date = CURRENT_DATE AND start_time = '12:00:00' LIMIT 1;
    INSERT INTO tickets (ticket_number, status, service_type, created_at) VALUES ('B011', 'зарегистрирован', 'confirm_appointment', NOW() - INTERVAL '45 minutes') RETURNING ticket_id INTO v_ticket_id;
    INSERT INTO appointments (schedule_id, patient_id, ticket_id) VALUES (v_schedule_id, 2, v_ticket_id);
    UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;
    
    -- --- ЗАПИСЬ №3: ЗАРЕГИСТРИРОВАН (ВРАЧ НА ПЕРЕРЫВЕ) ---
    SELECT schedule_id INTO v_schedule_id FROM schedules WHERE doctor_id = 3 AND date = CURRENT_DATE AND start_time = '12:30:00' LIMIT 1;
    INSERT INTO tickets (ticket_number, status, service_type, created_at) VALUES ('B012', 'зарегистрирован', 'confirm_appointment', NOW() - INTERVAL '35 minutes') RETURNING ticket_id INTO v_ticket_id;
    INSERT INTO appointments (schedule_id, patient_id, ticket_id) VALUES (v_schedule_id, 3, v_ticket_id);
    UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;
END $$;

-- 6.5 Создание будущих записей на прием (без талонов) на 7 дней вперед
DO $$
DECLARE
    d_id INT;
    p_id INT;
    s_id INT;
    day_offset INT;
BEGIN
    FOR d_id IN 1..7 LOOP -- Для каждого врача
        FOR day_offset IN 0..6 LOOP -- На каждый из 7 дней
            FOR i IN 1..4 LOOP -- Создаем по 4 случайные записи на день
                -- Выбираем случайного пациента
                p_id := floor(random() * 15 + 1)::INT;
                -- Выбираем случайный временной слот, который еще не занят
                SELECT schedule_id INTO s_id FROM schedules
                WHERE doctor_id = d_id AND date = (CURRENT_DATE + day_offset * INTERVAL '1 day')
                AND is_available = TRUE
                ORDER BY random()
                LIMIT 1;
                
                -- Если свободный слот найден, создаем запись
                IF s_id IS NOT NULL THEN
                    INSERT INTO appointments (schedule_id, patient_id) VALUES (s_id, p_id);
                    UPDATE schedules SET is_available = FALSE WHERE schedule_id = s_id;
                END IF;
            END LOOP;
        END LOOP;
    END LOOP;
END $$;
