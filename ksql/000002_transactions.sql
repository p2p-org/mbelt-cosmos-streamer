CREATE STREAM {APP_PREFIX}_TRANSACTIONS_STREAM (
  "tx_hash" VARCHAR,
  "chain_id" VARCHAR,
   "block_height" BIGINT,
    "time" BIGINT,
   "tx_index" BIGINT,
    "count_messages" BIGINT,
   "logs" STRING,
   "events" STRING,
   "msgs" STRING,
   "fee" STRING,
   "signatures" STRING,
   "memo" STRING,
   "status" STRING,
   "external_info" STRING
) WITH (kafka_topic=' {APP_PREFIX}_TRANSACTIONS_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_TRANSACTIONS_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS
SELECT * FROM {APP_PREFIX}_TRANSACTIONS_STREAM EMIT CHANGES;