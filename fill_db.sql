INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться к врачу', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Другой вопрос', 'D')
ON CONFLICT (service_id) DO NOTHING;

INSERT INTO tickets (ticket_number, status, window_number, created_at)
VALUES
  ('A001', 'ожидает', 1, CURRENT_TIMESTAMP),
  ('A002', 'приглашен', 2, CURRENT_TIMESTAMP),
  ('B003', 'на_приеме', 3, CURRENT_TIMESTAMP),
  ('B004', 'завершен', 1, CURRENT_TIMESTAMP),
  ('D005', 'зарегистрирован', 3, CURRENT_TIMESTAMP)
ON CONFLICT (ticket_number) DO NOTHING;

-- Добавляем данные о враче
-- Предполагаем, что этот врач будет иметь ID=1
INSERT INTO doctors (doctor_id, full_name, specialization, is_active) VALUES
(1, 'Иванов Иван Иванович', 'Терапевт', TRUE)
ON CONFLICT (doctor_id) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  specialization = EXCLUDED.specialization,
  is_active = EXCLUDED.is_active;