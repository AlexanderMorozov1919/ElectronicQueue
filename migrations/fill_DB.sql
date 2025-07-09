INSERT INTO services (service_id, name, letter) VALUES
  ('make_appointment', 'Записаться к врачу', 'A'),
  ('confirm_appointment', 'Прием по записи', 'B'),
  ('lab_tests', 'Сдать анализы', 'C'),
  ('documents', 'Другой вопрос', 'D');

-- Примеры талонов
INSERT INTO tickets (ticket_number, status, window_number, created_at)
VALUES
  ('A001', 'ожидает', 1, CURRENT_TIMESTAMP),
  ('A002', 'приглашен', 2, CURRENT_TIMESTAMP),
  ('B003', 'на_приеме', 3, CURRENT_TIMESTAMP),
  ('B004', 'завершен', 1, CURRENT_TIMESTAMP),
  ('C005', 'подойти_к_окну', 2, CURRENT_TIMESTAMP),
  ('D006', 'зарегистрирован', 3, CURRENT_TIMESTAMP);
