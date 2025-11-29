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
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/delayqueue"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/kafka"
	outboxPublisher "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/messaging/outbox"
	cartReadModel "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/readmodel/cart"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/subscriber"
	cartAbandonmentService "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/subscriber/service"
	cartProjector "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/projector/cart"
	projectorService "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/projector/service"
	commandUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
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

	// Messaging
	MessageProducer messaging.MessageProducer
	TopicRouter     messaging.TopicRouter
	DelayQueue      messaging.DelayQueue
	OutboxPublisher messaging.OutboxPublisher

	// Read model
	CartStore readmodelstore.CartStore

	// Subscribers
	CartAbandonmentSubscriber messaging.Subscriber
	CartProjector            gateway.Projector

	// Consumer Groups
	CartAbandonmentConsumer messaging.ConsumerGroup
	ProjectorConsumer       messaging.ConsumerGroup

	// Use case layer
	CartAddItemCommand                      commandUseCase.CartAddItemCommandInterface
	CreateTenantCartAbandonedPolicyCommand  commandUseCase.CreateTenantCartAbandonedPolicyCommandInterface
	UpdateTenantCartAbandonedPolicyCommand  commandUseCase.UpdateTenantCartAbandonedPolicyCommandInterface
	GetCartQuery                           queryUseCase.GetCartQueryInterface

	// Services
	CartAbandonmentService gateway.CartAbandonmentService
	ProjectorService       gateway.ProjectorService
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

	// Messaging infrastructure
	c.MessageProducer, err = kafka.NewProducer(cfg.KafkaConfig.Brokers)
	if err != nil {
		return err
	}
	c.TopicRouter = kafka.NewStaticTopicRouter()
	c.DelayQueue = delayqueue.NewMemoryDelayQueue()
	c.OutboxPublisher = outboxPublisher.NewOutboxPublisher(
		c.Transaction,
		c.OutboxRepo,
		c.MessageProducer,
		c.TopicRouter,
	)

	c.CartAddItemCommand = commandUseCase.NewCartAddItemCommand(c.Transaction, c.EventStore, c.OutboxRepo)
	c.CreateTenantCartAbandonedPolicyCommand = commandUseCase.NewCreateTenantCartAbandonedPolicyCommand(c.Transaction, c.EventStore, c.OutboxRepo)
	c.UpdateTenantCartAbandonedPolicyCommand = commandUseCase.NewUpdateTenantCartAbandonedPolicyCommand(c.Transaction, c.EventStore, c.OutboxRepo)

	// Read model and queries
	c.CartStore = cartReadModel.NewCartReadModel(c.Transaction)
	c.GetCartQuery = queryUseCase.NewGetCartQuery(c.CartStore)

	// Subscribers
	c.CartAbandonmentSubscriber = subscriber.NewCartAbandonmentSubscriber(
		c.Transaction,
		c.EventStore,
		c.DelayQueue,
	)
	c.CartProjector = cartProjector.NewCartProjector(c.CartStore)

	// Consumer Groups
	topics := []string{"ec.cart-events"}
	c.CartAbandonmentConsumer, err = kafka.NewConsumerGroup(cfg.KafkaConfig.Brokers, "cart-abandonment-group", topics, c.Deserializer)
	if err != nil {
		return err
	}
	c.ProjectorConsumer, err = kafka.NewConsumerGroup(cfg.KafkaConfig.Brokers, "cart-projector-group", topics, c.Deserializer)
	if err != nil {
		return err
	}

	// Services
	c.CartAbandonmentService = cartAbandonmentService.NewCartAbandonmentService(
		c.Deserializer,
		c.CartAbandonmentSubscriber,
		c.CartAbandonmentConsumer,
		c.DelayQueue,
	)
	c.ProjectorService = projectorService.NewProjectorService(
		c.Transaction,
		c.Deserializer,
		c.CartProjector,
		c.ProjectorConsumer,
	)

	return nil
}
