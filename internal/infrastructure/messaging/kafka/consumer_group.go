package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
)

type ConsumerGroup struct {
	brokers       []string
	groupID       string
	topics        []string
	handlers      []messaging.MessageHandler
	deserializer  repository.EventDeserializer
	consumerGroup sarama.ConsumerGroup
}

func NewConsumerGroup(brokers []string, groupID string, topics []string, deserializer repository.EventDeserializer) (messaging.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &ConsumerGroup{
		brokers:       brokers,
		groupID:       groupID,
		topics:        topics,
		handlers:      make([]messaging.MessageHandler, 0),
		deserializer:  deserializer,
		consumerGroup: consumerGroup,
	}, nil
}

func (c *ConsumerGroup) AddHandler(handler messaging.MessageHandler) {
	c.handlers = append(c.handlers, handler)
}

func (c *ConsumerGroup) Start(ctx context.Context) error {
	consumer := &groupConsumer{
		consumerGroup: c,
	}

	go func() {
		for {
			if err := c.consumerGroup.Consume(ctx, c.topics, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Printf("Error from consumer group: %v", err)
		}
	}()

	<-ctx.Done()
	return c.consumerGroup.Close()
}

func (c *ConsumerGroup) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var msg dto.Message
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return nil
	}

	for _, handler := range c.handlers {
		if err := handler(ctx, &msg); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}

	return nil
}

func (c *ConsumerGroup) Close() error {
	return c.consumerGroup.Close()
}

type groupConsumer struct {
	consumerGroup *ConsumerGroup
}

func (gc *groupConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (gc *groupConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (gc *groupConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			err := gc.consumerGroup.handleMessage(session.Context(), message)
			if err != nil {
				log.Printf("Error handling message: %v", err)
			}
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
