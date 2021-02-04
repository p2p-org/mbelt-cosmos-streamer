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
    "time" BIGINT,
   "tx_index" BIGINT,
   "logs" STRING,
   "events" STRING,
   "msgs" STRING,
   "fee" STRING,
   "signatures" STRING,
   "memo" STRING,
   "status" STRING,
   "external_info" STRING
) WITH (kafka_topic='transactions_stream', value_format='JSON');

CREATE STREAM TRANSACTIONS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM TRANSACTIONS_STREAM EMIT CHANGES;

CREATE STREAM MESSAGES_STREAM (
   "block_height" BIGINT,
  "tx_hash" VARCHAR,
  "tx_index" BIGINT,
  "msg_index" BIGINT,
  "msg_type" VARCHAR,
  "msg_info" STRING,
  "logs" STRING,
  "events" STRING,
  "external_info" STRING
) WITH (kafka_topic='messages_stream', value_format='JSON');

CREATE STREAM MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS SELECT *
FROM MESSAGES_STREAM EMIT CHANGES;
