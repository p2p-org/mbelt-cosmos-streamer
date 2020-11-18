CREATE TABLE IF NOT EXISTS cosmos.messages
(
    id            BIGSERIAL PRIMARY KEY,
--     block_height  bigint                   NOT NULL,
    block_hash    varchar(64) NOT NULL,
    tx_hash       varchar(64) NOT NULL,
    tx_index      int,
    msg_index     int,
    msg_type      varchar(64) NOT NULL,
    msg_info      jsonb,
    logs          text,
    events        jsonb,
    external_info jsonb, -- TODO Is it necessary at all?
    unique (tx_hash, block_hash)
);

CREATE TABLE IF NOT EXISTS cosmos._messages
(
    id            BIGSERIAL PRIMARY KEY,
    block_hash    varchar(64) NOT NULL,
    tx_hash       varchar(64) NOT NULL,
    tx_index      int,
    msg_index     int,
    msg_type      varchar(64) NOT NULL,
    msg_info      jsonb,
    logs          text,
    events        jsonb,
    external_info jsonb -- TODO Is it necessary at all?
);
-- TODO проверять по параметрам tx_hash, block_hash  msg_index
-- CREATE TABLE IF NOT EXISTS cosmos.messages_0
-- (
--     UNIQUE (block_hash, tx_hash),
--     CHECK (id BETWEEN 0 AND 10000 )
-- ) INHERITS (cosmos.messages);
--
-- CREATE INDEX messages_0_block_height_idx ON cosmos.messages_0 (block_height);
-- CREATE INDEX messages_0_chain_id_idx ON cosmos.messages_0 (chain_id);
-- CREATE INDEX messages_0_status_idx ON cosmos.messages_0 (status);
--
-- CREATE INDEX messages_0_events_params_idx ON cosmos.messages_0 USING GIN ((events -> 'Params') jsonb_path_ops);
-- CREATE INDEX messages_0_msg_idx ON cosmos.messages_0 USING GIN (msg jsonb_path_ops);

-- CREATE OR REPLACE FUNCTION messages_insert_trigger()
--     RETURNS TRIGGER AS
-- $$
-- BEGIN
--     INSERT INTO cosmos.messages_0 VALUES (NEW.*) ON CONFLICT DO NOTHING;
--     RETURN NULL;
-- END;
-- $$
--     LANGUAGE plpgsql;
--
-- CREATE TRIGGER insert_messages_trigger
--     BEFORE INSERT
--     ON cosmos.messages
--     FOR EACH ROW
-- EXECUTE FUNCTION messages_insert_trigger();



CREATE OR REPLACE FUNCTION cosmos.sink_messages_insert()
    RETURNS trigger AS
$$
BEGIN
    INSERT INTO cosmos.messages("block_hash",
                                "tx_hash",
                                "tx_index",
                                "msg_index",
                                "msg_type",
                                "msg_info",
                                "logs",
                                "events",
                                "external_info")
    VALUES (NEW."block_hash",
            NEW."tx_hash",
            NEW."tx_index",
            NEW."msg_index",
            NEW."msg_type",
            NEW."msg_info",
            NEW."logs",
            NEW."events"::jsonb,
            NEW."external_info"::jsonb)
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_messages_sink_upsert
    BEFORE INSERT
    ON cosmos._messages
    FOR EACH ROW
EXECUTE PROCEDURE cosmos.sink_messages_insert();
