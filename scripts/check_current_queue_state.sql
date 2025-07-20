-- =================================================================
-- ==              ПРОВЕРКА ТЕКУЩЕГО СОСТОЯНИЯ ОЧЕРЕДИ             ==
-- =================================================================
--
-- Назначение:
-- Показывает все активные талоны (кроме "завершенных"),
-- чтобы быстро оценить, кто и как долго ждет, и кто куда вызван.
--
-- psql -h localhost -p 5432 -U postgres -d el_queue -f scripts/check_current_queue_state.sql
--

SELECT
    ticket_number,
    status,
    window_number,
    service_type,
    created_at,
    -- Вычисляем время ожидания для наглядности
    (now() - created_at) AS waiting_time
FROM
    tickets
WHERE
    status <> 'завершен'
ORDER BY
    created_at ASC; -- Сортируем так, чтобы самые "старые" талоны были вверху