-- Заполнение таблицы услуг
INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться к врачу', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Другой вопрос', 'D')
ON CONFLICT (service_id) DO NOTHING;

-- Добавляем данные о враче
INSERT INTO doctors (doctor_id, full_name, specialization, is_active) VALUES
(1, 'Иванов Иван Иванович', 'Терапевт', TRUE)
ON CONFLICT (doctor_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  specialization = EXCLUDED.specialization,
  is_active = EXCLUDED.is_active;

-- Добавляем двух тестовых регистраторов.
-- Пароль для обоих: 'admin1' и 'admin2'
INSERT INTO registrars (registrar_id, full_name, login, password_hash, is_active) VALUES
(1, 'Петрова Анна Сергеевна', 'admin1', '$2a$10$Bpqg4mfUfFNLe09MC6QvveFNcY80VAiLSjpOOSBqAnV7avNJo5eEi', TRUE),
(2, 'Сидорова Елена Игоревна', 'admin2', '$2a$10$l3Lt5ogEKuQ1.PoSilqWQ.12ymyxTcpWTOBBBKJE6grMJ2emeFPcy', TRUE)
ON CONFLICT (registrar_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  login = EXCLUDED.login,
  password_hash = EXCLUDED.password_hash,
  is_active = EXCLUDED.is_active;

/*
INSERT INTO tickets (ticket_number, status, window_number, created_at)
VALUES
  ('A001', 'ожидает', 1, CURRENT_TIMESTAMP),
  ('B001', 'приглашен', 2, CURRENT_TIMESTAMP),
  ('C001', 'на_приеме', 3, CURRENT_TIMESTAMP),
  ('D001', 'завершен', 1, CURRENT_TIMESTAMP),
  ('A002', 'зарегистрирован', 3, CURRENT_TIMESTAMP)
ON CONFLICT (ticket_number) DO NOTHING;
*/