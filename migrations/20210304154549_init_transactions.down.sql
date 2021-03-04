DROP TRIGGER IF EXISTS trg_transactions_sink_trim_after_upsert ON cosmos._transactions;
DROP FUNCTION IF EXISTS cosmos.sink_trim_transactions_after_insert;
DROP TRIGGER IF EXISTS trg_transactions_sink_upsert ON cosmos._transactions;
DROP FUNCTION IF EXISTS cosmos.sink_transactions_insert;
DROP FUNCTION IF EXISTS cosmos.block_status_change;
DROP TABLE IF EXISTS cosmos._transactions;
DROP TABLE IF EXISTS cosmos.transactions;