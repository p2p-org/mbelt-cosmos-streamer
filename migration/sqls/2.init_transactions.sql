CREATE TABLE IF NOT EXISTS cosmos.transactions
(
    tx_hash       varchar(64) NOT NULL PRIMARY KEY,
    chain_id      varchar(64) NOT NULL,
    block_height  bigint      NOT NULL,
    tx_index      int,
    logs          text,
    events        jsonb,
    msgs          jsonb,
    fee           jsonb,
    signatures    jsonb,
    memo          varchar(1024),
    status        status_enum,
    external_info jsonb -- TODO Is it necessary at all?
);

CREATE TABLE IF NOT EXISTS cosmos._transactions
(
    tx_hash       varchar(64) NOT NULL PRIMARY KEY,
    chain_id      varchar(64) NOT NULL,
    block_height  bigint      NOT NULL,
    tx_index      int,
    logs          text,
    events        text,
    msgs          text,
    fee           text,
    signatures    text,
    memo          varchar(1024),
    status        varchar(64),
    external_info text -- TODO Is it necessary at all?
);
--
-- CREATE TABLE IF NOT EXISTS cosmos.transactions_0
-- (
--     UNIQUE (block_hash, tx_hash),
--     CHECK (block_height BETWEEN 0 AND 50000 )
-- ) INHERITS (cosmos.transactions);
--
-- CREATE INDEX transactions_0_block_height_idx ON cosmos.transactions_0 (block_height);
-- CREATE INDEX transactions_0_chain_id_idx ON cosmos.transactions_0 (chain_id);
-- CREATE INDEX transactions_0_status_idx ON cosmos.transactions_0 (status);
--
-- CREATE INDEX transactions_0_events_params_idx ON cosmos.transactions_0 USING GIN ((events -> 'Params') jsonb_path_ops);
-- CREATE INDEX transactions_0_msgs_idx ON cosmos.transactions_0 USING GIN (msgs jsonb_path_ops);
--
-- CREATE OR REPLACE FUNCTION transactions_insert_trigger()
--     RETURNS TRIGGER AS
-- $$
-- BEGIN
--     INSERT INTO cosmos.transactions_0 VALUES (NEW.*) ON CONFLICT DO NOTHING;
--     RETURN NULL;
-- END;
-- $$
--     LANGUAGE plpgsql;
--
-- CREATE TRIGGER insert_transactions_trigger
--     BEFORE INSERT
--     ON cosmos.transactions
--     FOR EACH ROW
-- EXECUTE FUNCTION transactions_insert_trigger();


CREATE OR REPLACE FUNCTION cosmos.block_status_change()
    RETURNS trigger AS
$$
BEGIN
    UPDATE cosmos.blocks
    SET status = 'confirmed'::status_enum
    where height = NEW.block_height
      AND num_tx = (
        select count(*) from cosmos.transactions where block_height = NEW.block_height
    )
      AND cosmos.blocks.chain_id = NEW.chain_id;
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_block_status_change
    AFTER INSERT
    ON cosmos.transactions
    FOR EACH ROW
EXECUTE PROCEDURE cosmos.block_status_change();


CREATE OR REPLACE FUNCTION cosmos.sink_transactions_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO cosmos.transactions("tx_hash",
                                    "chain_id",
                                    "block_height",
                                    "tx_index",
                                    "logs",
                                    "events",
                                    "msgs",
                                    "fee",
                                    "signatures",
                                    "memo",
                                    "status",
                                    "external_info")
    VALUES (NEW."tx_hash",
            NEW."chain_id",
            NEW."block_height",
            NEW."tx_index",
            NEW."logs",
            NEW."events"::jsonb,
            NEW."msgs"::jsonb,
            NEW."fee"::jsonb,
            NEW."signatures"::jsonb,
            NEW."memo",
            NEW."status"::status_enum,
            NEW."external_info"::jsonb)
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_transactions_sink_upsert
    BEFORE INSERT
    ON cosmos._transactions
    FOR EACH ROW
EXECUTE PROCEDURE cosmos.sink_transactions_insert();



CREATE OR REPLACE FUNCTION cosmos.sink_trim_transactions_after_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE FROM cosmos._transactions WHERE "tx_hash" = NEW."tx_hash" AND "block_height" = NEW."block_height";
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_transactions_sink_trim_after_upsert
    AFTER INSERT
    ON cosmos._transactions
    FOR EACH ROW
EXECUTE PROCEDURE cosmos.sink_trim_transactions_after_insert();

