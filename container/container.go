package container

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore/deserializer"
	outboxRepo "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/outbox"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
	outboxPublisher "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/outbox"
	cartReadModel "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/readmodel/cart"
	commandUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	queryUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
)

type Container struct {
	// Config
	Cfg *config.Config

	// Repository layer
	Transaction  repository.Transaction
	EventStore   repository.EventStore
	OutboxRepo   repository.OutboxRepository
	Deserializer repository.EventDeserializer

	// Messaging (interfaces only)
	OutboxPublisher messaging.OutboxPublisher

	// Read model
	CartStore readmodelstore.CartStore

	// Use case layer (CQRS)
	CartAddItemCommand commandUseCase.CartAddItemCommandInterface
	GetCartQuery       queryUseCase.GetCartQueryInterface
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Inject(ctx context.Context, cfg *config.Config) error {
	c.Cfg = cfg

	databaseClient, err := client.NewClient(cfg.DatabaseConfig)
	if err != nil {
		return err
	}

	c.Transaction = transaction.NewTransaction(databaseClient.GetDB())
	c.Deserializer = deserializer.NewEventDeserializer()
	c.EventStore = eventstore.NewEventStore(c.Deserializer)
	c.OutboxRepo = outboxRepo.NewOutboxRepository()

	messageProducer, err := kafka.NewProducer(cfg.KafkaConfig.Brokers)
	if err != nil {
		return err
	}
	topicRouter := kafka.NewStaticTopicRouter()
	c.OutboxPublisher = outboxPublisher.NewOutboxPublisher(
		c.Transaction,
		c.OutboxRepo,
		messageProducer,
		topicRouter,
	)

	c.CartAddItemCommand = commandUseCase.NewCartAddItemCommand(c.Transaction, c.EventStore, c.OutboxRepo)

	// Read model and queries
	c.CartStore = cartReadModel.NewCartReadModel(c.Transaction)
	c.GetCartQuery = queryUseCase.NewGetCartQuery(c.CartStore)

	return nil
}
