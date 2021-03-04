package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaPartition    = 0
	TopicBlocks       = "BLOCKS_STREAM"
	TopicMessages     = "MESSAGES_STREAM"
	TopicTransactions = "TRANSACTIONS_STREAM"
)

type KafkaDatastore struct {
	config       *config.Config
	kafkaWriters map[string]*kafka.Writer
	// ack   chan kafka.Event
	// pushChan chan interface{}
}

func Init(config *config.Config) (*KafkaDatastore, error) {
	ds := &KafkaDatastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
		// pushChan:     make(chan interface{}),
	}
	// , TopicMessages
	for _, topic := range []string{TopicBlocks, TopicTransactions, TopicMessages} {
		topicWithPrefix := strings.ToUpper(config.KafkaPrefix) + "_" + topic
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{ds.config.KafkaHost},
			Topic:    topicWithPrefix,
			Balancer: &kafka.LeastBytes{},
		})
		if writer == nil {
			return nil, errors.New("cannot create kafka writer")
		}

		ds.kafkaWriters[topicWithPrefix] = writer
	}

	return ds, nil
}

func (ds *KafkaDatastore) Push(topic string, m map[string]interface{}) (err error) {
	topicWithPrefix := strings.ToUpper(ds.config.KafkaPrefix) + "_" + topic
	var (
		kMsgs []kafka.Message
	)
	// log.Println("[KafkaDatastore][push][Debug] Push data to kafka")

	if ds.kafkaWriters == nil {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writers not initialized")
		return errors.New("cannot push")
	}

	if _, ok := ds.kafkaWriters[topicWithPrefix]; !ok {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writer not initialized for topic", topicWithPrefix)
	}

	for key, value := range m {
		data, err := json.Marshal(value)
		if err != nil {
			log.Println("[KafkaDatastore][Error][push]", "Cannot marshal push data", err)
			return errors.New("cannot push")
		}
		kMsgs = append(kMsgs, kafka.Message{
			Key:   []byte(key),
			Value: data,
		})
	}

	if len(kMsgs) == 0 {
		return nil
	}

	err = ds.kafkaWriters[topicWithPrefix].WriteMessages(context.TODO(), kMsgs...)

	if err != nil {
		log.Println("[KafkaDatastore][Error][push]", "Cannot produce data", err)
	}
	return err
}
