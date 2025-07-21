-- =================================================================
-- ==          ПРОВЕРКА РАСПИСАНИЯ ВРАЧА НА СЕГОДНЯШНИЙ ДЕНЬ        ==
-- =================================================================
--
-- Назначение:
-- Показывает все слоты для конкретного врача (ID = 1) на текущую дату.
-- Если слот занят, выводится ФИО пациента.
--
-- P.S. Замените `s.doctor_id = 1` на ID нужного врача.
-- psql -h localhost -p 5432 -U postgres -d el_queue -f scripts/check_doctor_schedule_today.sql
--

SELECT
    s.schedule_id,
    s.start_time,
    s.end_time,
    s.is_available,
    p.full_name AS patient_name
FROM
    schedules s
-- Используем LEFT JOIN, чтобы показать и свободные слоты (у которых нет записи)
LEFT JOIN
    appointments a ON s.schedule_id = a.schedule_id
LEFT JOIN
    patients p ON a.patient_id = p.patient_id
WHERE
    s.doctor_id = 1  -- <-- УКАЖИТЕ ID НУЖНОГО ВРАЧА
    AND s.date = CURRENT_DATE -- CURRENT_DATE всегда возвращает сегодняшний день
ORDER BY
    s.start_time;