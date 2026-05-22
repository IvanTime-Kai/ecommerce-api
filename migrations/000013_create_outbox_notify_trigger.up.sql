CREATE OR REPLACE FUNCTION notify_outbox()
RETURNS TRIGGER AS $$
BEGIN 
  PERFORM pg_notify('outbox_channel', NEW.id::text);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER outbox_insert_trigger
AFTER INSERT ON outbox
FOR EACH ROW EXECUTE FUNCTION notify_outbox();
