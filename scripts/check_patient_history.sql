-- =================================================================
-- ==                 ПОЛНАЯ ИСТОРИЯ ВИЗИТОВ ПАЦИЕНТА             ==
-- =================================================================
--
-- Назначение:
-- Находит пациента по части ФИО и показывает все его записи на прием.
--
-- P.S. Замените '%Андреев%' на ФИО нужного пациента.
-- psql -h localhost -p 5432 -U postgres -d el_queue -f scripts/check_patient_history.sql
--

SELECT
    p.full_name AS patient,
    d.full_name AS doctor,
    d.specialization,
    s.date,
    s.start_time,
    t.ticket_number
FROM
    patients p
JOIN
    appointments a ON p.patient_id = a.patient_id
JOIN
    schedules s ON a.schedule_id = s.schedule_id
JOIN
    doctors d ON s.doctor_id = d.doctor_id
LEFT JOIN
    tickets t ON a.ticket_id = t.ticket_id
WHERE
    p.full_name ILIKE '%Андреев%' -- <-- ЧАСТЬ ФИО ПАЦИЕНТА
ORDER BY
    s.date DESC, s.start_time DESC;