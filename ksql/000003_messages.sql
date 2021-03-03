CREATE STREAM {APP_PREFIX}_MESSAGES_STREAM (
   "block_height" BIGINT,
  "tx_hash" VARCHAR,
  "tx_index" BIGINT,
  "msg_index" BIGINT,
  "msg_type" VARCHAR,
  "msg_info" STRING,
  "logs" STRING,
  "events" STRING,
  "external_info" STRING
) WITH (kafka_topic='{APP_PREFIX}_MESSAGES_STREAM', value_format='JSON');

CREATE STREAM {APP_PREFIX}_MESSAGES_STREAM_AVRO WITH(PARTITIONS=1, REPLICAS=1, VALUE_FORMAT='AVRO') AS
SELECT * FROM {APP_PREFIX}_MESSAGES_STREAM EMIT CHANGES;