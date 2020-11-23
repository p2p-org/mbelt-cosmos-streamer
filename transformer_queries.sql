CREATE STREAM BLOCKS_STREAM (
  "hash" VARCHAR,
  "chain_id" VARCHAR,
  "height" BIGINT,
  "time" BIGINT,
  "num_tx" BIGINT,
  "total_txs" BIGINT,
  "last_block_hash" VARCHAR,
  "validator" VARCHAR,
  "txs_hash" STRING,
  "status" VARCHAR
) WITH (kafka_topic='blocks_stream', value_format='JSON');

CREATE STREAM BLOCKS_STREAM_AVRO WITH(PARTITIONS=1, VALUE_FORMAT='AVRO', REPLICAS=1) AS SELECT *
FROM BLOCKS_STREAM EMIT CHANGES;

CREATE STREAM TRANSACTIONS_STREAM (
  "tx_hash" VARCHAR,
  "chain_id" VARCHAR,
   "block_height" BIGINT,
   "block_hash" VARCHAR,
   "time" BIGINT,
   "tx_index" BIGINT,
   "logs" TEXT,
   "events" TEXT,
   "msgs" TEXT,
   "fee" TEXT,
   "signatures" TEXT,
   "memo" TEXT,
   "status" STRING,
   "external_info" TEXT
) WITH (kafka_topic='transactions_stream', value_format='JSON');

CREATE STREAM TRANSACTIONS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM TRANSACTIONS_STREAM EMIT CHANGES;


CREATE STREAM MESSAGES_STREAM (
  "cid" VARCHAR,
  "block_cid" VARCHAR,
  "method" INTEGER,
  "from" VARCHAR,
  "to" VARCHAR,
  "value" BIGINT,
  "gas" VARCHAR,
  "params" STRING,
  "data" VARCHAR,
  "block_time" BIGINT
) WITH (kafka_topic='messages_stream', value_format='JSON');

CREATE STREAM MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM MESSAGES_STREAM EMIT CHANGES;
