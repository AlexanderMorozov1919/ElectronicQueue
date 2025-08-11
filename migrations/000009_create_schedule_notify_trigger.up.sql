-- Подавляем вывод NOTICE-сообщений, например, при удалении несуществующего триггера
SET client_min_messages TO warning;

CREATE OR REPLACE FUNCTION notify_schedule_change() RETURNS TRIGGER AS $$
DECLARE
    payload JSONB;
    data_row RECORD;
    doctor_info RECORD;
    operation_text TEXT;
BEGIN
    operation_text := TG_OP;

    -- Определяем, какую строку использовать: старую (при удалении) или новую
    IF (operation_text = 'DELETE') THEN
        data_row := OLD;
    ELSE
        data_row := NEW;
    END IF;

    -- Получаем информацию о враче
    SELECT doctor_id, full_name, specialization
    INTO doctor_info
    FROM doctors
    WHERE doctor_id = data_row.doctor_id;

    -- Если врач не найден, ничего не делаем
    IF NOT FOUND THEN
        IF (operation_text = 'DELETE') THEN
            RETURN OLD;
        ELSE
            RETURN NEW;
        END IF;
    END IF;

    -- Формируем сложный JSON объект, который ожидает фронтенд
    payload := jsonb_build_object(
        'operation', lower(operation_text),
        'data', jsonb_build_object(
            'date', to_char(data_row.date, 'YYYY-MM-DD'),
            'doctors', jsonb_build_array(
                jsonb_build_object(
                    'id', doctor_info.doctor_id,
                    'full_name', doctor_info.full_name,
                    'specialization', doctor_info.specialization,
                    'slots', jsonb_build_array(
                        jsonb_build_object(
                            'start_time', to_char(data_row.start_time, 'HH24:MI:SS'),
                            'end_time', to_char(data_row.end_time, 'HH24:MI:SS'),
                            'is_available', data_row.is_available,
                            'cabinet', data_row.cabinet
                        )
                    )
                )
            )
        )
    );

    -- Отправляем уведомление на канал 'schedule_update'
    PERFORM pg_notify('schedule_update', payload::text);

    IF (operation_text = 'DELETE') THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Удаляем старый триггер, если он существует
DROP TRIGGER IF EXISTS schedules_change_trigger ON schedules;

-- Создаем новый триггер, который будет срабатывать после вставки, обновления или удаления
CREATE TRIGGER schedules_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON schedules
FOR EACH ROW EXECUTE FUNCTION notify_schedule_change();

-- Возвращаем уровень сообщений по умолчанию
RESET client_min_messages;