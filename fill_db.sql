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
    administrators,
    reception_logs,
    patients
RESTART IDENTITY CASCADE;

-- -----------------------------------------------------------------
-- --                        1. УСЛУГИ                            --
-- -----------------------------------------------------------------
INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Получить результаты', 'D');

-- -----------------------------------------------------------------
-- --                      2. РЕГИСТРАТОРЫ                        --
-- -----------------------------------------------------------------
-- Пароли: 'pass1' - 'pass7'
INSERT INTO registrars (window_number, login, password_hash) VALUES
(1, 'admin1', '$2a$10$g3C8q/gOeSAT2uVgFXz8M.Xs4OELKH7gD24P3nciXTfbm4RePxWqG'),
(2, 'admin2', '$2a$10$fnLMINfO.s4.zMr6MooIHuoxiLy1CcFCGajzmH8VlxGw0BvnMB75C'),
(3, 'admin3', '$2a$10$LFWIFiiMooFIOXX2IsW4vuQwvNE3vtUqPLDyvKA8cdO/1YSjKzAsu'),
(4, 'admin4', '$2a$10$NOnwKKyVVyhAZGPgtOUF9uQzbWYSmtVZSHUosreJfbxj/vYL/XmC2');

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
-- --                      4. АДМИНИСТРАТОРЫ                      --
-- -----------------------------------------------------------------
-- Пароль: 'superpass'
INSERT INTO administrators (full_name, login, password_hash) VALUES
('Главный Администратор', 'root', '$2a$10$hpeOn4.Jv/K6M3lvFpJ9NO.PjLi4U3vpFfWWZ6cPS/nUc0lwDSuGu');

-- -----------------------------------------------------------------
-- --                        5. ПАЦИЕНТЫ                          --
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
-- --        6. РАСПИСАНИЕ ВРАЧЕЙ (СЕГОДНЯ + 6 ДНЕЙ ВПЕРЕД)       --
-- -----------------------------------------------------------------
INSERT INTO schedules (doctor_id, cabinet, date, start_time, end_time)
SELECT
    d.doctor_id,
    (100 + d.doctor_id) AS cabinet,
    d.day::date,
    s.start_time::time,
    (s.start_time + '30 minutes'::interval)::time AS end_time
FROM 
    (SELECT doctor_id, generate_series(CURRENT_DATE, CURRENT_DATE + interval '6 days', '1 day') as day FROM doctors) d
CROSS JOIN generate_series(
    (CURRENT_DATE + '06:00'::time)::timestamp,
    (CURRENT_DATE + '21:30'::time)::timestamp,
    '30 minutes'::interval
) AS s(start_time)
WHERE 
    (
        (d.doctor_id = 1 AND extract(isodow from d.day) <= 5 AND s.start_time::time >= '08:00' AND s.start_time::time < '18:00')
        OR
        (d.doctor_id = 2 AND s.start_time::time >= '09:00' AND s.start_time::time < '20:00')
        OR
        (d.doctor_id = 3 AND extract(isodow from d.day) IN (1, 3, 5) AND s.start_time::time >= '07:00' AND s.start_time::time < '16:00')
        OR
        (d.doctor_id = 4 AND extract(isodow from d.day) IN (2, 4, 6) AND s.start_time::time >= '14:00' AND s.start_time::time < '22:00')
        OR
        (d.doctor_id = 5 AND extract(isodow from d.day) IN (1, 2, 3) AND s.start_time::time >= '10:00' AND s.start_time::time < '15:00')
        OR
        (d.doctor_id = 6 AND extract(isodow from d.day) <= 6 AND s.start_time::time >= '06:00' AND s.start_time::time < '16:00')
        OR
        (d.doctor_id = 7 AND s.start_time::time >= '12:00' AND s.start_time::time < '21:00')
    );

-- -----------------------------------------------------------------
-- --                7. ТАЛОНЫ И ЗАПИСИ НА ПРИЕМ                  --
-- -----------------------------------------------------------------
-- Сценарий: Середина рабочего дня, примерно 12:00
-- -- 7.1 Талоны в статусе "завершен"
-- INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at, called_at, completed_at) VALUES
-- ('A001', 'завершен', 'make_appointment', 1, NOW() - INTERVAL '3 hour', NOW() - INTERVAL '2 hour 50 minutes', NOW() - INTERVAL '2 hour 45 minutes'),
-- ('C001', 'завершен', 'lab_tests', 2, NOW() - INTERVAL '2 hour 30 minutes', NOW() - INTERVAL '2 hour 20 minutes', NOW() - INTERVAL '2 hour 10 minutes'),
-- ('D001', 'завершен', 'documents', 1, NOW() - INTERVAL '2 hour', NOW() - INTERVAL '1 hour 55 minutes', NOW() - INTERVAL '1 hour 40 minutes');

-- -- 7.2 Талоны в статусе "приглашен"
-- INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at, called_at) VALUES
-- ('A002', 'приглашен', 'make_appointment', 1, NOW() - INTERVAL '1 hour 30 minutes', NOW() - INTERVAL '1 minute'),
-- ('B001', 'приглашен', 'confirm_appointment', 3, NOW() - INTERVAL '1 hour 25 minutes', NOW() - INTERVAL '30 seconds');

-- -- 7.3 Талоны в статусе "ожидает"
-- INSERT INTO tickets (ticket_number, status, service_type, created_at) VALUES
-- ('C002', 'ожидает', 'lab_tests', NOW() - INTERVAL '1 hour'),
-- ('D002', 'ожидает', 'documents', NOW() - INTERVAL '55 minutes'),
-- ('A003', 'ожидает', 'make_appointment', NOW() - INTERVAL '50 minutes'),
-- ('B002', 'ожидает', 'confirm_appointment', NOW() - INTERVAL '40 minutes'),
-- ('C003', 'ожидает', 'lab_tests', NOW() - INTERVAL '30 minutes'),
-- ('A004', 'ожидает', 'make_appointment', NOW() - INTERVAL '20 minutes'),
-- ('D003', 'ожидает', 'documents', NOW() - INTERVAL '10 minutes'),
-- ('B003', 'ожидает', 'confirm_appointment', NOW() - INTERVAL '5 minutes');

-- -- 7.4 Один пациент НА ПРИЕМЕ (для демонстрации)
-- DO $$
-- DECLARE
--     v_schedule_id INT;
--     v_ticket_id INT;
-- BEGIN
--     SELECT schedule_id INTO v_schedule_id FROM schedules WHERE doctor_id = 1 AND date = CURRENT_DATE AND start_time >= '11:00:00' AND is_available = TRUE ORDER BY start_time LIMIT 1;
--     IF v_schedule_id IS NOT NULL THEN
--         INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at, started_at) VALUES ('B010', 'на_приеме', 'confirm_appointment', 4, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '5 minutes') RETURNING ticket_id INTO v_ticket_id;
--         INSERT INTO appointments (schedule_id, patient_id, ticket_id) VALUES (v_schedule_id, 1, v_ticket_id);
--         UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;
--     END IF;
-- END $$;

-- -- 7.5 Создаем по 4 ЗАПИСАННЫХ пациента для КАЖДОГО врача на СЕГОДНЯ
-- DO $$
-- DECLARE
--     d_id INT;
--     p_id_start INT := 2;
--     v_schedule_id INT;
--     v_ticket_id INT;
--     ticket_num INT := 11;
-- BEGIN
--     FOR d_id IN 1..7 LOOP
--         FOR i IN 1..4 LOOP
--             SELECT schedule_id INTO v_schedule_id FROM schedules
--             WHERE doctor_id = d_id AND date = CURRENT_DATE AND is_available = TRUE
--             ORDER BY start_time
--             LIMIT 1;

--             IF v_schedule_id IS NOT NULL THEN
--                 INSERT INTO tickets (ticket_number, status, service_type, window_number, created_at) 
--                 VALUES ('B0' || ticket_num::text, 'зарегистрирован', 'confirm_appointment', floor(random() * 7 + 1)::INT, NOW() - (random() * 60 + 5) * INTERVAL '1 minute') 
--                 RETURNING ticket_id INTO v_ticket_id;
                
--                 INSERT INTO appointments (schedule_id, patient_id, ticket_id) VALUES (v_schedule_id, p_id_start, v_ticket_id);
                
--                 UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;

--                 p_id_start := p_id_start + 1;
--                 ticket_num := ticket_num + 1;
--                 IF p_id_start > 15 THEN p_id_start := 2; END IF;
--             END IF;
--         END LOOP;
--     END LOOP;
-- END $$;

-- -- 7.6 Создание будущих записей на прием (без талонов) на 6 дней вперед
-- DO $$
-- DECLARE
--     d_id INT;
--     p_id INT;
--     s_id INT;
--     day_offset INT;
-- BEGIN
--     FOR d_id IN 1..7 LOOP
--         FOR day_offset IN 1..6 LOOP
--             FOR i IN 1..4 LOOP
--                 p_id := floor(random() * 15 + 1)::INT;
                
--                 SELECT schedule_id INTO s_id FROM schedules
--                 WHERE doctor_id = d_id AND date = (CURRENT_DATE + day_offset * INTERVAL '1 day')
--                 AND is_available = TRUE
--                 ORDER BY random()
--                 LIMIT 1;
                
--                 IF s_id IS NOT NULL THEN
--                     INSERT INTO appointments (schedule_id, patient_id) VALUES (s_id, p_id);
--                     UPDATE schedules SET is_available = FALSE WHERE schedule_id = s_id;
--                 END IF;
--             END LOOP;
--         END LOOP;
--     END LOOP;
-- END $$;

-- -- 7.7 Добавляем "заочные" записи (без талонов) на сегодня и завтра
-- DO $$
-- DECLARE
--     d_id INT;
--     p_id INT;
--     s_id INT;
--     day_offset INT;
-- BEGIN
--     FOR d_id IN 1..7 LOOP 
--         FOR day_offset IN 0..1 LOOP 
--             FOR i IN 1..3 LOOP 
--                 p_id := floor(random() * 15 + 1)::INT;
                
--                 SELECT schedule_id INTO s_id FROM schedules
--                 WHERE doctor_id = d_id AND date = (CURRENT_DATE + day_offset * INTERVAL '1 day')
--                 AND is_available = TRUE
--                 AND (CASE WHEN day_offset = 0 THEN start_time >= '14:00:00' ELSE TRUE END)
--                 ORDER BY random()
--                 LIMIT 1;
                
--                 IF s_id IS NOT NULL THEN
--                     INSERT INTO appointments (schedule_id, patient_id) VALUES (s_id, p_id);
--                     UPDATE schedules SET is_available = FALSE WHERE schedule_id = s_id;
--                 END IF;
--             END LOOP;
--         END LOOP;
--     END LOOP;
-- END $$;

-- -- 7.8 Добавляем ОПОЗДАВШЕГО пациента (запись на 9 утра) и связанный с ним талон
-- DO $$
-- DECLARE
--     v_schedule_id INT;
--     v_ticket_id INT;
--     v_patient_id INT := 3; -- Берем 3-го пациента для примера
-- BEGIN
--     -- Находим слот на 9:00 сегодня у первого врача
--     SELECT schedule_id INTO v_schedule_id 
--     FROM schedules 
--     WHERE doctor_id = 1 AND date = CURRENT_DATE AND start_time = '09:00:00'
--     LIMIT 1;
    
--     IF v_schedule_id IS NOT NULL THEN
--         -- Создаем талон "В ожидании" для этой записи
--         INSERT INTO tickets (ticket_number, status, service_type, created_at) 
--         VALUES ('B099', 'ожидает', 'confirm_appointment', NOW() - INTERVAL '15 minutes') 
--         RETURNING ticket_id INTO v_ticket_id;
        
--         -- Создаем саму запись, связывая пациента, расписание и талон
--         INSERT INTO appointments (schedule_id, patient_id, ticket_id) 
--         VALUES (v_schedule_id, v_patient_id, v_ticket_id);
        
--         -- Помечаем слот как занятый
--         UPDATE schedules SET is_available = FALSE WHERE schedule_id = v_schedule_id;
--     END IF;
-- END $$;