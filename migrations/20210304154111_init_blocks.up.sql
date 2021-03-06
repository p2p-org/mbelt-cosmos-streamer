CREATE SCHEMA IF NOT EXISTS cosmos;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_enum') THEN
            CREATE TYPE status_enum AS ENUM ('pending', 'confirmed', 'rejected', 'onfork');
        END IF;
END$$;

CREATE TABLE IF NOT EXISTS cosmos.blocks
(
    "hash"            varchar(64)              NOT NULL PRIMARY KEY,
    "chain_id"        varchar(64)              NOT NULL,
    "height"          bigint                   NOT NULL,
    "time"            timestamp WITH TIME ZONE NOT NULL,
                                    "num_tx"          bigint                   NOT NULL,
                                    "total_txs"       bigint,
                                    "last_block_hash" varchar(64),
    "validator"       varchar(64),
    "txs_hash"        varchar(64)[],
    "status"          status_enum,
    unique (hash, height)
    );
CREATE INDEX IF NOT EXISTS blocks_height_index ON cosmos.blocks (height);
CREATE INDEX IF NOT EXISTS blocks_chain_id_index ON cosmos.blocks (chain_id);


-- Fix for unquoting varchar json
-- CREATE OR REPLACE FUNCTION varchar_to_jsonb(varchar) RETURNS jsonb AS
-- $$
-- SELECT to_jsonb($1)
--            $$ LANGUAGE SQL;
--
-- CREATE CAST (varchar as jsonb) WITH FUNCTION varchar_to_jsonb(varchar) AS IMPLICIT;

-- Internal tables

CREATE TABLE IF NOT EXISTS cosmos._blocks
(
    "hash"            varchar(64)              NOT NULL PRIMARY KEY,
    "chain_id"        varchar(64)              NOT NULL,
    "height"          bigint                   NOT NULL,
    "time"            BIGINT NOT NULL,
    "num_tx"          bigint                   NOT NULL,
    "total_txs"       bigint,
    "last_block_hash" varchar(64),
    "validator"       varchar(64),
    "txs_hash"        TEXT,
    "status"          varchar(64)
    );

-- Blocks
CREATE OR REPLACE FUNCTION cosmos.sink_blocks_insert()
    RETURNS trigger AS
$$
BEGIN
INSERT INTO cosmos.blocks("hash",
                          "chain_id",
                          "height",
                          "time",
                          "num_tx",
                          "total_txs",
                          "last_block_hash",
                          "validator",
                          "txs_hash",
                          "status")
VALUES (NEW."hash",
        NEW."chain_id",
        NEW."height",
        to_timestamp(NEW."time"),
        NEW."num_tx",
        NEW."total_txs",
        NEW."last_block_hash",
        NEW."validator",
        NEW."txs_hash"::varchar(256)[],
        NEW."status"::status_enum)
    ON CONFLICT DO NOTHING;

RETURN NEW;
END ;

$$
LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_sink_upsert
    BEFORE INSERT
    ON cosmos._blocks
    FOR EACH ROW
    EXECUTE PROCEDURE cosmos.sink_blocks_insert();


-- Blocks

CREATE OR REPLACE FUNCTION cosmos.sink_trim_blocks_after_insert()
    RETURNS trigger AS
$$
BEGIN
DELETE FROM cosmos._blocks WHERE "hash" = NEW."hash" AND "chain_id" = NEW."chain_id" AND "height" = NEW."height";
RETURN NEW;
END ;

$$
LANGUAGE 'plpgsql';

CREATE TRIGGER trg_blocks_sink_trim_after_upsert
    AFTER INSERT
    ON cosmos._blocks
    FOR EACH ROW
    EXECUTE PROCEDURE cosmos.sink_trim_blocks_after_insert();

