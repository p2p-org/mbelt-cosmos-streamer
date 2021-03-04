DROP TRIGGER IF EXISTS trg_consistency_sink_trim_before_upsert ON cosmos.consistency;
DROP FUNCTION IF EXISTS cosmos.sink_trim_consistency_before_insert;
DROP FUNCTION IF EXISTS cosmos.set_consistency;
DROP TABLE IF EXISTS cosmos.consistency;
