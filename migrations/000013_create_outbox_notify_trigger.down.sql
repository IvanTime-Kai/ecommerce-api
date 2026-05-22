DROP TRIGGER IF EXISTS outbox_insert_trigger ON outbox;

DROP FUNCTION IF EXISTS notify_outbox();