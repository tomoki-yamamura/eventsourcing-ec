package container

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/bus"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/client"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore/deserializer"
	kafkaEventStore "github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/eventstore"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/transaction"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/projector/todo"
	commandUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore"
	queryUseCase "github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query"
)

type Container struct {
	// Config
	Cfg *config.Config

	// Repository layer
	Transaction  repository.Transaction
	EventStore   repository.EventStore
	Deserializer repository.EventDeserializer

	// Gateway implementation
	EventBus      gateway.EventBus
	TodoProjector gateway.Projector
	TodoViewRepo  readmodelstore.TodoListStore

	// Use case layer (CQRS)
	TodoListCreateCommand commandUseCase.TodoListCreateCommandInterface
	TodoAddItemCommand    commandUseCase.TodoAddItemCommandInterface
	QueryUseCase          queryUseCase.TodoListQueryInterface
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

	// Repository layer
	c.Transaction = transaction.NewTransaction(databaseClient.GetDB())
	c.Deserializer = deserializer.NewEventDeserializer()
	
	// Use Kafka Event Store instead of MySQL Event Store
	router := kafkaEventStore.NewStaticMapRouter()
	eventStore, err := kafkaEventStore.NewKafkaEventStore(cfg.KafkaConfig.Brokers, router, c.Deserializer)
	if err != nil {
		return err
	}
	c.EventStore = eventStore

	// Event Bus and Projector
	c.EventBus = bus.NewInMemoryEventBus()
	viewRepo := todo.NewInMemoryTodoListViewRepository()
	c.TodoViewRepo = viewRepo
	c.TodoProjector = todo.NewTodoProjector(viewRepo)

	// Use case layer (CQRS)
	c.TodoListCreateCommand = commandUseCase.NewTodoListCreateCommand(c.Transaction, c.EventStore, c.EventBus)
	c.TodoAddItemCommand = commandUseCase.NewTodoAddItemCommand(c.Transaction, c.EventStore, c.EventBus)
	c.QueryUseCase = queryUseCase.NewTodoListQuery(c.TodoViewRepo)

	return nil
}

func (c *Container) RestoreReadModels(ctx context.Context) error {
	// TODO: Implement GetAllEvents for Kafka Event Store or use different approach
	// Currently disabled as GetAllEvents is not implemented for Kafka Event Store
	return nil
}
