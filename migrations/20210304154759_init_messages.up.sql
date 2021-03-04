CREATE TABLE IF NOT EXISTS cosmos.messages
(
    id            BIGSERIAL PRIMARY KEY,
    block_height  bigint      NOT NULL,
    tx_hash       varchar(64) NOT NULL,
    tx_index      int,
    msg_index     int,
    msg_type      varchar(64) NOT NULL,
    msg_info      jsonb,
    logs          text,
    events        jsonb,
    external_info jsonb -- TODO Is it necessary at all?
    );

CREATE INDEX messages_block_height_idx ON cosmos.messages (block_height);
CREATE INDEX messages_tx_hash_idx ON cosmos.messages (tx_hash);
CREATE INDEX messages_tx_index_idx ON cosmos.messages (tx_index);

CREATE TABLE IF NOT EXISTS cosmos._messages
(
    id            BIGSERIAL PRIMARY KEY,
    block_height  bigint      NOT NULL,
    tx_hash       varchar(64) NOT NULL,
    tx_index      int,
    msg_index     int,
    msg_type      varchar(64) NOT NULL,
    msg_info      text,
    logs          text,
    events        text,
    external_info text -- TODO Is it necessary at all?
    );
-- TODO проверять по параметрам tx_hash, block_hash  msg_index

CREATE OR REPLACE FUNCTION cosmos.sink_messages_insert()
    RETURNS trigger AS
$$
BEGIN
INSERT INTO cosmos.messages("block_height",
                            "tx_hash",
                            "tx_index",
                            "msg_index",
                            "msg_type",
                            "msg_info",
                            "logs",
                            "events",
                            "external_info")
VALUES (NEW."block_height",
        NEW."tx_hash",
        NEW."tx_index",
        NEW."msg_index",
        NEW."msg_type",
        NEW."msg_info"::jsonb,
        NEW."logs"::text,
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