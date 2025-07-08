CREATE OR REPLACE FUNCTION notify_ticket_update() RETURNS TRIGGER AS $$
BEGIN
    -- NEW содержит новую версию строки для операций INSERT или UPDATE
    PERFORM pg_notify('ticket_update', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер, который вызывает функцию после каждого INSERT или UPDATE в таблице tickets
CREATE TRIGGER tickets_update_trigger
AFTER INSERT OR UPDATE ON tickets
FOR EACH ROW EXECUTE FUNCTION notify_ticket_update();