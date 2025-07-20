-- =================================================================
-- ==     ПРОВЕРКА ПОСЛЕДНИХ ЗАПИСЕЙ НА ПРИЕМ С ДЕТАЛИЗАЦИЕЙ       ==
-- =================================================================
--
-- Назначение:
-- Этот скрипт показывает 10 последних созданных записей на прием,
-- объединяя данные из четырех таблиц (appointments, patients,
-- schedules, doctors) для вывода полной и понятной информации.
--
-- Как использовать из терминала:
-- psql -h localhost -p 5432 -U postgres -d el_queue -f scripts\check_appointments.sql
--

SELECT
    a.appointment_id,                     -- ID самой записи
    p.full_name AS patient_name,          -- ФИО пациента
    d.full_name AS doctor_name,           -- ФИО врача
    s.date AS appointment_date,           -- Дата приема
    s.start_time AS appointment_time,     -- Время начала приема
    t.ticket_number                       -- Номер талона, по которому пришел пациент
FROM
    appointments a
    -- Присоединяем таблицу пациентов, чтобы получить ФИО по patient_id
JOIN
    patients p ON a.patient_id = p.patient_id
    -- Присоединяем таблицу расписаний, чтобы получить дату, время и ID врача
JOIN
    schedules s ON a.schedule_id = s.schedule_id
    -- Присоединяем таблицу врачей, чтобы получить ФИО по doctor_id из расписания
JOIN
    doctors d ON s.doctor_id = d.doctor_id
    -- Присоединяем таблицу талонов, чтобы получить номер талона
LEFT JOIN
    tickets t ON a.ticket_id = t.ticket_id
ORDER BY
    a.appointment_id DESC                 -- Сортируем по ID записи в обратном порядке (самые новые сверху)
LIMIT 10;                                 -- Ограничиваем вывод последними 10 записями