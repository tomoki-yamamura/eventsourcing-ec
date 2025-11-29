package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore/deserializer"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
	cartProjector "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/projector/cart"
	cartReadModel "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/readmodel/cart"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
)

type ProjectorService struct {
	transaction   repository.Transaction
	deserializer  repository.EventDeserializer
	cartProjector gateway.Projector
	kafkaConsumer *kafka.ConsumerGroup
}

func NewProjectorService(cfg *config.Config) (*ProjectorService, error) {
	// Database client
	databaseClient, err := client.NewClient(cfg.DatabaseConfig)
	if err != nil {
		return nil, err
	}

	// Transaction
	transaction := transaction.NewTransaction(databaseClient.GetDB())

	// Deserializer
	deserializer := deserializer.NewEventDeserializer()

	// Read model and projector
	cartStore := cartReadModel.NewCartReadModel(transaction)
	cartProjector := cartProjector.NewCartProjector(cartStore)

	// Kafka consumer
	topics := []string{"ec.cart-events"}
	kafkaConsumer, err := kafka.NewConsumerGroup(cfg.KafkaConfig.Brokers, "cart-projector-group", topics, deserializer)
	if err != nil {
		return nil, err
	}

	service := &ProjectorService{
		transaction:   transaction,
		deserializer:  deserializer,
		cartProjector: cartProjector,
		kafkaConsumer: kafkaConsumer,
	}

	kafkaConsumer.AddHandler(service.handleMessage)

	return service, nil
}

func (s *ProjectorService) handleMessage(ctx context.Context, msg *messaging.Message) error {
	// Marshal message data for deserializer
	eventData, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	// Deserialize event
	event, err := s.deserializer.Deserialize(msg.Type, eventData)
	if err != nil {
		return err
	}

	// Handle with projector
	return s.cartProjector.Handle(ctx, event)
}

func (s *ProjectorService) Start(ctx context.Context) error {
	log.Println("Starting Projector Service...")
	return s.kafkaConsumer.Start(ctx)
}
