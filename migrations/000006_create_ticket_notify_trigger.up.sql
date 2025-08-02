-- Подавляем вывод NOTICE-сообщений, например, при удалении несуществующего триггера
SET client_min_messages TO warning;

CREATE OR REPLACE FUNCTION notify_ticket_change() RETURNS TRIGGER AS $$
DECLARE
    payload JSON;
    action TEXT;
    channel_name TEXT := 'ticket_update';
    data_row RECORD;
BEGIN
    action := TG_OP;

    IF (TG_OP = 'DELETE') THEN
        data_row := OLD;
    ELSE
        data_row := NEW;
    END IF;

    payload := json_build_object(
        'action', lower(action),
        'data', json_build_object(
            'ticket_id', data_row.ticket_id,
            'ticket_number', data_row.ticket_number,
            'status', data_row.status,
            'service_type', data_row.service_type,
            'window_number', data_row.window_number,
            
            'qr_code', encode(data_row.qr_code, 'base64'), 
            
            'created_at', to_char(data_row.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
            'called_at', to_char(data_row.called_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
            'started_at', to_char(data_row.started_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
            'completed_at', to_char(data_row.completed_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"')
        )
    );

    PERFORM pg_notify(channel_name, payload::text);

    RETURN data_row;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tickets_update_trigger ON tickets;
DROP TRIGGER IF EXISTS tickets_change_trigger ON tickets;

CREATE TRIGGER tickets_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON tickets
FOR EACH ROW EXECUTE FUNCTION notify_ticket_change();

-- Возвращаем уровень сообщений по умолчанию
RESET client_min_messages;