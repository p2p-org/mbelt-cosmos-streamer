CREATE TABLE IF NOT EXISTS cosmos.consistency
(
    id               BIGSERIAL PRIMARY KEY,
    block_height     bigint    NOT NULL,
    max_block_height bigint    NOT NULL,
    count_blocks     bigint    NOT NULL,
    count_txs        bigint    NOT NULL,
    count_messages   bigint    NOT NULL,
    created_at       timestamp NOT NULL,
    unique (block_height)
);

CREATE INDEX consistency_block_height_idx ON cosmos.consistency (block_height);

CREATE OR REPLACE FUNCTION cosmos.sink_trim_consistency_before_insert()
    RETURNS trigger AS
$$
BEGIN
    DELETE
    FROM cosmos.consistency
    where id in (select id from cosmos.consistency order by id desc offset 100 limit 1);
    RETURN NEW;
END ;

$$
    LANGUAGE 'plpgsql';

CREATE TRIGGER trg_consistency_sink_trim_before_upsert
    BEFORE INSERT
    ON cosmos.consistency
    FOR EACH ROW
EXECUTE PROCEDURE cosmos.sink_trim_consistency_before_insert();


CREATE OR REPLACE FUNCTION cosmos.set_consistency(h bigint) RETURNS SETOF void AS
$BODY$
DECLARE
    cBlocks        bigint;
    maxBlockHeight bigint;
    cTxs           bigint;
    cMsgs          bigint;
BEGIN
    select max(height) into maxBlockHeight from cosmos.blocks;
    select count(*) into cBlocks from cosmos.blocks;
    select count(*) into cTxs from cosmos.transactions;
    select count(*) into cMsgs from cosmos.messages;
    Insert into cosmos.consistency (block_height, max_block_height, count_blocks, count_txs, count_messages, created_at)
    VALUES (h, maxBlockHeight, cBlocks, cTxs, cMsgs, now())
    ON CONFLICT DO NOTHING;
END
$BODY$
    LANGUAGE plpgsql;
