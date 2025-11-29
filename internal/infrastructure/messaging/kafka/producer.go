package kafka

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

type Producer struct {
	producer sarama.SyncProducer
	brokers  []string
}

func NewProducer(brokers []string) (messaging.MessageProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, appErrors.Unknown.Wrap(err, "failed to create kafka producer")
	}

	return &Producer{
		producer: producer,
		brokers:  brokers,
	}, nil
}

func (p *Producer) PublishMessage(topic string, key string, message *messaging.Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return appErrors.Unknown.Wrap(err, "failed to serialize message")
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(messageBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("message-type"),
				Value: []byte(message.Type),
			},
			{
				Key:   []byte("version"),
				Value: []byte(strconv.Itoa(message.Version)),
			},
		},
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return appErrors.Unknown.Wrap(err, "failed to send message to kafka")
	}

	return nil
}

func (p *Producer) PublishDelayedMessage(topic, key string, message *messaging.Message, delay time.Duration) error {
	return appErrors.Unknown.New("delayed message publishing not supported by Kafka producer")
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
