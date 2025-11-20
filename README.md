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
   task migrate:up
   task migrate:test:up
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

#### Read Model (Projector) Migrations

- `task migrate:projector:up` - Run read model migrations
- `task migrate:projector:down` - Rollback read model migrations
- `task migrate:projector:status` - Check migration status
- `task migrate:projector:create -- migration_name` - Create new migration

#### Combined Commands

- `task migrate:up` - Run all migrations (eventstore + projector)
- `task migrate:down` - Rollback all migrations
- `task migrate:test:up` - Run all migrations for test database

---

## API Endpoints

The application provides RESTful APIs for cart management:

### Add Item to Cart

```bash
POST /additem/{aggregate_id}
```

**Request body:**

```json
{
  "user_id": "user-123",
  "item_id": "item-456",
  "quantity": 2,
  "price": 29.99
}
```

**Example:**

```bash
curl -X POST "http://localhost:8080/additem/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "item_id": "item-456",
    "quantity": 2,
    "price": 29.99
  }'
```

**Response:**

```json
{
  "aggregate_id": "550e8400-e29b-41d4-a716-446655440000",
  "version": 2,
  "events": [
    {
      "type": "ItemAddedToCart",
      "version": 2,
      "occurred_at": "2025-11-20T16:30:00Z"
    }
  ],
  "status": "success",
  "executed_at": "2025-11-20T16:30:00Z"
}
```

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
