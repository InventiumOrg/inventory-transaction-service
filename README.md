# Inventory Transaction Service

A Go microservice for managing inventory import/export transaction records,
backed by DynamoDB and producing Avro-encoded events to Kafka via Confluent
Schema Registry.

The project follows the same layout as
[`InventiumOrg/inventory-service`](https://github.com/InventiumOrg/inventory-service)
(Gin + OTEL + Prometheus) but uses DynamoDB instead of Postgres+sqlc since this
service owns the transaction table.

## Features

- Create / read / update inventory transaction records
- DynamoDB storage (auto-creates the table on first start)
- Kafka producer publishing Avro-encoded `InventoryImportEvent`s on update
- OpenTelemetry (OTLP/HTTP) tracing + metrics
- Prometheus `/metrics` endpoint
- `/health`, `/healthz`, `/readyz` health endpoints
- Multi-stage distroless Docker image
- Graceful shutdown

## Tech stack

- **Go 1.25**
- **Gin** HTTP framework
- **AWS SDK for Go v2** (DynamoDB)
- **segmentio/kafka-go** + **linkedin/goavro/v2** for Avro Kafka publishing
- **spf13/viper** for configuration
- **OpenTelemetry SDK** + **Prometheus client_golang** for observability

## Project layout

```
inventory-transaction-service/
├── api/                  # Gin server wiring (OTEL, Prometheus, CORS, routes)
├── config/               # Viper-based env config
├── handlers/             # HTTP handlers (transaction CRUD, health)
├── middlewares/          # Cross-cutting middleware (panic recovery)
├── models/               # Domain entity, DTOs, Kafka event payload
├── observability/        # OTEL + Prometheus setup
├── processors/           # Kafka Avro producer + Schema Registry client
├── repository/           # DynamoDB client + transaction repository
├── routes/               # Route group definitions
└── main.go               # Composition root
```

## API

| Method | Path                                          | Description                       |
|-------:|-----------------------------------------------|-----------------------------------|
|   POST | `/api/v1/transactions`                        | Create a transaction record       |
|    GET | `/api/v1/transactions/:inventoryId/:id`       | Get a transaction record by key   |
|    PUT | `/api/v1/transactions/:inventoryId/:id`       | Update a transaction record       |
|    GET | `/health`, `/healthz`, `/readyz`              | Health/liveness/readiness         |
|    GET | `/metrics`                                    | Prometheus metrics                |

### Create transaction example

```bash
curl -X POST http://localhost:14330/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "IMPORT",
    "source": "warehouse-001",
    "destination": "pos-system-001",
    "inventoryId": "item-12345",
    "inventoryMeasure": "weight",
    "inventoryCategory": "electronics",
    "inventoryUnit": "kg",
    "quantity": 10
  }'
```

## Configuration

Configuration is loaded from environment variables (and an optional
`app.env` file in the working directory). See `app.env.example` for the
complete list.

| Variable                          | Default                                      | Description                       |
|-----------------------------------|----------------------------------------------|-----------------------------------|
| `SERVER_PORT`                     | `14330`                                      | HTTP listen port                  |
| `AWS_REGION`                      | `ap-southeast-1`                             | AWS region                        |
| `DYNAMODB_ENDPOINT`               | _(empty)_                                    | Override for DynamoDB Local       |
| `DYNAMODB_TABLE_NAME`             | `inventory-transaction-service`              | Table name                        |
| `KAFKA_BOOTSTRAP_SERVERS`         | `localhost:9092`                             | Kafka brokers (comma-separated)   |
| `KAFKA_TOPIC_NAME`                | `inventory.transaction.status.updated`       | Outbound Kafka topic              |
| `KAFKA_CA_PEM_LOCATION`           | _(empty)_                                    | CA PEM (Aiven Kafka TLS)          |
| `KAFKA_SVC_PEM_LOCATION`          | _(empty)_                                    | Service cert+key PEM              |
| `SCHEMA_REGISTRY_URL`             | _(empty)_                                    | Confluent Schema Registry URL     |
| `SCHEMA_REGISTRY_USERNAME`        | _(empty)_                                    | Basic-auth username               |
| `SCHEMA_REGISTRY_PASSWORD`        | _(empty)_                                    | Basic-auth password               |
| `KAFKA_INVENTORY_SCHEMA_SUBJECT`  | `inventory.transaction.status.updated`       | Avro subject                      |
| `KAFKA_AUTO_REGISTER_SCHEMAS`     | `true`                                       | Auto-register fallback schema     |
| `OTEL_EXPORTER_OTLP_ENDPOINT`     | _(empty)_                                    | OTLP/HTTP collector endpoint      |
| `TRACING_SAMPLING_PROBABILITY`    | `0.1`                                        | Trace sample ratio (0-1)          |

## Running locally

### Prerequisites

- Go 1.25+
- Docker (for DynamoDB Local + Kafka)

### Quick start

```bash
cp app.env.example app.env
make tidy
docker compose up -d dynamodb-local kafka zookeeper
make run
```

### Tests

```bash
make test
```

### Docker

```bash
make docker
make docker-run
```

Or run everything:

```bash
docker compose up --build
```

## Observability

- **Tracing**: OTLP/HTTP to `OTEL_EXPORTER_OTLP_ENDPOINT` when set.
- **Metrics**:
  - Prometheus pull at `/metrics` (HTTP request, DB, Kafka, business metrics).
  - OTEL counters/histograms in parallel.

## Notes on conversion

This service was converted from a Java Spring Boot codebase. Key mappings:

| Java                                       | Go                                          |
|--------------------------------------------|---------------------------------------------|
| `TransactionRecordController`              | `handlers/transaction.go`                   |
| `TransactionRecordService`                 | inlined in `handlers/transaction.go`        |
| `TransactionRecordRepository` (DynamoDb)   | `repository/transaction.go`                 |
| `DynamoDBConfig` (table auto-create)       | `repository.EnsureTable`                    |
| `InventoryProducer` (Avro + Kafka)         | `processors/inventory_producer.go` (+ Avro) |
| `SchemaRegistryService`                    | `processors/schema_registry_client.go`      |
| `GlobalExceptionHandler`                   | inline error helpers + `middlewares.Recovery` |
| Spring `application.properties`            | `config/config.go` (Viper) + `app.env`      |
| Micrometer + OTLP                          | `observability/otel.go` + `prometheus.go`   |
