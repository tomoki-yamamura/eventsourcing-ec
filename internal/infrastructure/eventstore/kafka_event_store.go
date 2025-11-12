package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	appErrors "github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

const (
	eventTopicPrefix = "events"
)

type kafkaEventStore struct {
	producer     sarama.SyncProducer
	consumer     sarama.Consumer
	deserializer repository.EventDeserializer
	brokers      []string
}

type StoredEvent struct {
	EventID      uuid.UUID   `json:"event_id"`
	EventType    string      `json:"event_type"`
	EventData    any         `json:"event_data"`
	Version      int         `json:"version"`
	AggregateID  uuid.UUID   `json:"aggregate_id"`
	CreatedAt    time.Time   `json:"created_at"`
}

func NewKafkaEventStore(brokers []string, deserializer repository.EventDeserializer) (repository.EventStore, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, appErrors.Unknown.Wrap(err, "failed to create kafka producer")
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, appErrors.Unknown.Wrap(err, "failed to create kafka consumer")
	}

	return &kafkaEventStore{
		producer:     producer,
		consumer:     consumer,
		deserializer: deserializer,
		brokers:      brokers,
	}, nil
}

func (k *kafkaEventStore) SaveEvents(ctx context.Context, aggregateID uuid.UUID, events []event.Event) error {
	topic := k.getTopicName(aggregateID)

	for _, evt := range events {
		storedEvent := StoredEvent{
			EventID:     evt.GetEventID(),
			EventType:   evt.GetEventType(),
			EventData:   evt,
			Version:     evt.GetVersion(),
			AggregateID: aggregateID,
			CreatedAt:   time.Now(),
		}

		eventBytes, err := json.Marshal(storedEvent)
		if err != nil {
			return appErrors.Unknown.Wrap(err, "failed to serialize event")
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(aggregateID.String()),
			Value: sarama.ByteEncoder(eventBytes),
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("event-type"),
					Value: []byte(evt.GetEventType()),
				},
				{
					Key:   []byte("version"),
					Value: []byte(strconv.Itoa(evt.GetVersion())),
				},
			},
		}

		partition, offset, err := k.producer.SendMessage(msg)
		if err != nil {
			return appErrors.RepositoryError.Wrap(err, "failed to send event to kafka")
		}

		log.Printf("Event saved to kafka: topic=%s, partition=%d, offset=%d", topic, partition, offset)
	}

	return nil
}

func (k *kafkaEventStore) LoadEvents(ctx context.Context, aggregateID uuid.UUID) ([]event.Event, error) {
	topic := k.getTopicName(aggregateID)

	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		if err == sarama.ErrUnknownTopicOrPartition {
			return nil, appErrors.NotFound.New("aggregate not found")
		}
		return nil, appErrors.QueryError.Wrap(err, "failed to create partition consumer")
	}
	defer partitionConsumer.Close()

	var events []event.Event
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var storedEvent StoredEvent
			if err := json.Unmarshal(msg.Value, &storedEvent); err != nil {
				return nil, appErrors.QueryError.Wrap(err, "failed to unmarshal stored event")
			}

			// 指定されたaggregateIDのイベントのみを取得
			if storedEvent.AggregateID != aggregateID {
				continue
			}

			eventData, err := json.Marshal(storedEvent.EventData)
			if err != nil {
				return nil, appErrors.QueryError.Wrap(err, "failed to marshal event data")
			}

			evt, err := k.deserializer.Deserialize(storedEvent.EventType, eventData)
			if err != nil {
				return nil, appErrors.QueryError.Wrap(err, fmt.Sprintf("failed to deserialize event %s", storedEvent.EventType))
			}

			events = append(events, evt)

		case <-timeout.C:
			if len(events) == 0 {
				return nil, appErrors.NotFound.New("no events found for aggregate")
			}
			return events, nil

		case err := <-partitionConsumer.Errors():
			return nil, appErrors.QueryError.Wrap(err, "consumer error")
		}
	}
}

func (k *kafkaEventStore) GetAllEvents(ctx context.Context) ([]event.Event, error) {
	return nil, appErrors.Unknown.New("GetAllEvents is not implemented for kafka event store")
}

func (k *kafkaEventStore) getTopicName(aggregateID uuid.UUID) string {
	return fmt.Sprintf("%s-%s", eventTopicPrefix, aggregateID.String())
}

func (k *kafkaEventStore) Close() error {
	if err := k.producer.Close(); err != nil {
		return err
	}
	return k.consumer.Close()
}
