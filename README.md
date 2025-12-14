# Event Sourcing E-Commerce Application

This is an **Event Sourcing** e-commerce application built in Go, implementing a Cart aggregate for shopping cart management. It demonstrates Domain-Driven Design (DDD) principles, CQRS pattern, and Event Sourcing architecture with MySQL as both Event Store and Read Model storage, plus Kafka for messaging.

---

## Architecture

The application follows **Clean Architecture** principles with Event Sourcing:

- **Domain Layer**: Cart aggregate, Value Objects (Price, Quantity), Commands, Events
- **UseCase Layer**: CQRS command/query handlers, business logic
- **Infrastructure Layer**: Event store, Read model projectors, HTTP handlers, Kafka messaging
- **Messaging**: Outbox pattern for reliable event delivery between Event Store and Kafka

### Event Sourcing Components

- **Cart Aggregate**: Manages shopping cart state through events (CartCreated, ItemAddedToCart, CartPurchased)
- **Event Store**: MySQL-based event persistence with optimistic locking
- **Read Models**: Separate cart views for querying (carts and cart_items tables)
- **Outbox Pattern**: Ensures reliable event publishing to Kafka
- **KRaft Mode**: Modern Kafka without ZooKeeper dependency

---

## Quick Start

### Prerequisites

- `direnv` - Environment variable management
- `docker` and `docker-compose` - Container orchestration
- `go` 1.24+ - Go programming language
- `task` - Task automation tool

### Installation

1. **Install Task (if not already installed):**

   ```bash
   # macOS
   brew install go-task/tap/go-task

   # Linux/Others
   sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d
   ```

2. **Start infrastructure (Kafka + MySQL):**

   ```bash
   task docker:up
   ```

3. **Run database migrations:**

   ```bash
   task migrate:all:up
   ```

4. **Verify setup:**

   ```bash
   task test
   ```

5. **Start the application:**
   ```bash
   task run
   ```

---

## Available Tasks

### Development

- `task test` - Run all tests
- `task lint` - Run code linting
- `task build` - Build application
- `task run` - Run the application

### Docker

- `task docker:up` - Start Kafka + MySQL containers
- `task docker:down` - Stop containers and remove volumes

### Database Migrations

#### Event Store Migrations

- `task migrate:eventstore:up` - Run event store migrations
- `task migrate:eventstore:down` - Rollback event store migrations
- `task migrate:eventstore:status` - Check migration status
- `task migrate:eventstore:create -- migration_name` - Create new migration

#### Read Model Migrations

- `task migrate:readmodel:up` - Run read model migrations
- `task migrate:readmodel:down` - Rollback read model migrations

#### Test Database Migrations

- `task migrate:test:eventstore:up` - Run event store migrations for test database
- `task migrate:test:eventstore:down` - Rollback event store migrations for test database
- `task migrate:test:readmodel:up` - Run read model migrations for test database
- `task migrate:test:readmodel:down` - Rollback read model migrations for test database

#### Combined Commands

- `task migrate:all:up` - Run all migrations (eventstore + readmodel for both main and test databases)
- `task migrate:all:down` - Rollback all migrations (eventstore + readmodel for both main and test databases)

---

## API Endpoints

The application provides RESTful APIs for cart management:

### Add Item to Cart

```bash
POST /carts/{aggregate_id}/items
```

**Request body:**

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174001",
  "item_id": "123e4567-e89b-12d3-a456-426614174002",
  "name": "Test Product",
  "price": 29.99,
  "tenant_id": "123e4567-e89b-12d3-a456-426614174003"
}
```

**Example:**

```bash
curl -X POST "http://localhost:8080/carts/550e8400-e29b-41d4-a716-446655440000/items" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174001",
    "item_id": "123e4567-e89b-12d3-a456-426614174002",
    "name": "Test Product",
    "price": 29.99,
    "tenant_id": "123e4567-e89b-12d3-a456-426614174003"
  }'
```

### Get Cart

```bash
GET /carts/{aggregate_id}
```

**Example:**

```bash
curl -X GET "http://localhost:8080/carts/550e8400-e29b-41d4-a716-446655440000"
```

### Create Tenant Cart Abandonment Policy

```bash
POST /tenants/{aggregate_id}/cart-abandoned-policies
```

**Request body:**

```json
{
  "title": "Standard Cart Abandonment Policy",
  "abandoned_minutes": 30,
  "quiet_time_from": "22:00:00",
  "quiet_time_to": "08:00:00"
}
```

### Update Tenant Cart Abandonment Policy

```bash
PUT /tenants/{aggregate_id}/cart-abandoned-policies
```

### Get Tenant Cart Abandonment Policy

```bash
GET /tenants/{aggregate_id}/cart-abandoned-policies
```

---

## Directory Structure

```
.
├── README.md               # Project documentation
├── Taskfile.yaml          # Task automation configuration
├── main.go                # Application entry point
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── docker-compose.yaml    # Docker services configuration
├── init-topics.sh         # Kafka topics initialization
├── .envrc                 # Environment variables (direnv)
├── .envrc.example         # Environment variables template
├── docker/                # Docker configurations
├── db/                    # Database migration files
├── container/             # Container-related files
└── internal/              # Application source code
    ├── config/            # Configuration management
    ├── errors/            # Custom error types
    ├── domain/            # Domain layer (DDD)
    │   ├── aggregate/     # Domain aggregates (Cart, Tenant)
    │   ├── command/       # Domain commands
    │   ├── entity/        # Domain entities
    │   ├── event/         # Domain events
    │   ├── repository/    # Domain repository interfaces
    │   └── value/         # Value objects (Price, Quantity, etc.)
    ├── usecase/           # Application layer (CQRS)
    │   ├── command/       # Command handlers and inputs
    │   ├── query/         # Query handlers, inputs, outputs
    │   └── ports/         # Interface definitions
    │       ├── gateway/   # External service interfaces
    │       ├── messaging/ # Messaging interfaces and DTOs
    │       ├── presenter/ # Presentation interfaces
    │       ├── readmodelstore/ # Read model storage interfaces
    │       └── view/      # View interfaces
    └── infrastructure/    # Infrastructure layer
        ├── database/      # Database implementations
        │   ├── client/    # Database clients
        │   ├── eventstore/ # Event Store implementation
        │   │   ├── deserializer/ # Event deserialization
        │   │   └── migration/    # Event store migrations
        │   ├── outbox/    # Outbox pattern implementation
        │   ├── readmodel/ # Read model implementations
        │   │   ├── cart/  # Cart read model
        │   │   ├── tenant/ # Tenant read model
        │   │   └── migrations/ # Read model migrations
        │   ├── testutil/  # Database testing utilities
        │   └── transaction/ # Transaction management
        ├── dto/           # Data transfer objects
        ├── handler/       # HTTP handlers
        │   ├── command/   # Command endpoint handlers
        │   ├── query/     # Query endpoint handlers
        │   ├── request/   # HTTP request models
        │   └── response/  # HTTP response models
        ├── messaging/     # Messaging implementations
        │   ├── kafka/     # Kafka integration
        │   │   └── ksql/  # KSQL configurations
        │   └── outbox/    # Outbox messaging
        ├── presenter/     # Presentation layer
        │   └── viewmodel/ # View models
        ├── projector/     # Event projectors
        │   ├── cart/      # Cart projector
        │   ├── service/   # Projector services
        │   └── tenant/    # Tenant projector
        ├── delayqueue/    # Delay queue implementation
        ├── register/      # Dependency injection
        ├── router/        # HTTP routing
        ├── subscriber/    # Event subscribers
        │   └── service/   # Subscriber services
        └── view/          # View implementations
```

### Key Directories

- **`internal/domain/`**: Core business logic following DDD principles
- **`internal/usecase/`**: Application logic implementing CQRS pattern
- **`internal/infrastructure/`**: External concerns (database, HTTP, messaging)
- **`internal/infrastructure/database/eventstore/`**: Event sourcing implementation
- **`internal/infrastructure/messaging/`**: Kafka integration and outbox pattern
- **`internal/infrastructure/projector/`**: Read model projection logic

---

## Database Schema

### Event Store Tables

**events** - Stores all domain events

```sql
CREATE TABLE events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_id CHAR(36) NOT NULL UNIQUE,
    aggregate_id CHAR(36) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSON NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_aggregate_version (aggregate_id, version)
);
```

**outbox** - Outbox pattern for reliable messaging

```sql
CREATE TABLE outbox (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id CHAR(36) NOT NULL,
    aggregate_id CHAR(36) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSON NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP NULL,
    status ENUM('PENDING', 'PUBLISHED', 'FAILED') DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    error_message TEXT NULL
);
```

## Testing

Run the test suite:

```bash
# Run all tests
task test

# Run specific test packages
go test ./internal/domain/...
go test ./internal/usecase/...
go test ./internal/infrastructure/...
```

The application includes:

- **Unit tests** for domain logic, value objects, aggregates
- **Integration tests** for event store, database operations
- **Test database** setup with dedicated migration support

---
