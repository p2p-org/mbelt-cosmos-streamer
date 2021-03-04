DROP TRIGGER IF EXISTS trg_messages_sink_upsert ON cosmos._messages;
DROP FUNCTION IF EXISTS cosmos.sink_messages_insert;
DROP TABLE IF EXISTS cosmos._messages;
DROP TABLE IF EXISTS cosmos.messages;