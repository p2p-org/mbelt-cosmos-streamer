DROP TRIGGER IF EXISTS trg_blocks_sink_trim_after_upsert ON cosmos._blocks;
DROP TRIGGER IF EXISTS trg_blocks_sink_upsert ON cosmos._blocks;
DROP FUNCTION IF EXISTS cosmos.sink_blocks_insert;
DROP FUNCTION IF EXISTS cosmos.sink_trim_blocks_after_insert;
DROP TABLE IF EXISTS cosmos._blocks;
DROP TABLE IF EXISTS cosmos.blocks;
DROP SCHEMA IF EXISTS cosmos;
DROP TYPE IF EXISTS status_enum;
