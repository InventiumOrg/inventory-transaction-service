# Inventory Transaction Service

A Spring Boot microservice for managing inventory import/export transaction records with DynamoDB storage and Kafka event streaming.

## Features

- Create and retrieve inventory transaction records
- Support for both import and export transactions
- DynamoDB storage with Enhanced Client
- Kafka event publishing for transaction events
- REST API with validation and error handling
- Health check endpoints
- Docker containerization support

## Technology Stack

- **Java 25** (Amazon Corretto)
- **Spring Boot 4.0.1**
- **AWS DynamoDB** with Enhanced Client
- **Apache Kafka** with Spring Kafka
- **Gradle** build system
- **Docker** containerization

## API Endpoints

### Transaction Management
- `POST /api/v1/transactions` - Create a new transaction record
- `GET /api/v1/transactions/{id}` - Retrieve a transaction record by ID

### Health Check
- `GET /health` - Service health status

## Quick Start

### Prerequisites
- Java 25
- Docker and Docker Compose (for local development)

### Running Locally

1. **Clone the repository**
```bash
git clone <repository-url>
cd inventory-transaction-service
```

2. **Build the application**
```bash
./gradlew build
```

3. **Start dependencies with Docker Compose**
```bash
docker-compose up -d kafka zookeeper dynamodb-local
```

4. **Run the application**
```bash
./gradlew bootRun
```

The service will be available at `http://localhost:14330`

### Running with Docker

1. **Build and run everything**
```bash
docker-compose up --build
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka broker addresses | `localhost:9092` |
| `AWS_REGION` | AWS region for DynamoDB | `ap-southeast-1` |
| `DYNAMODB_ENDPOINT` | DynamoDB endpoint (for local dev) | _(empty)_ |

### Application Properties

Key configuration in `src/main/resources/application.properties`:

```properties
server.port=14330
spring.kafka.bootstrap-servers=${KAFKA_BOOTSTRAP_SERVERS:localhost:9092}
kafka.topic.transaction-events=transaction-events
aws.region=ap-southeast-1
```

## API Usage Examples

### Create Transaction Record

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

### Get Transaction Record

```bash
curl http://localhost:14330/api/v1/transactions/{transaction-id}
```

## Development

### Build Commands

```bash
# Build the project
./gradlew build

# Run tests
./gradlew test

# Clean and build
./gradlew clean build

# Run application
./gradlew bootRun
```

### Testing

The service includes comprehensive error handling and validation:

- Input validation with detailed error messages
- Custom exception handling for not found scenarios
- Global exception handler for consistent error responses

## Architecture

The service follows a layered architecture:

- **Controller Layer**: REST endpoints and request/response handling
- **Service Layer**: Business logic and Kafka integration
- **Repository Layer**: DynamoDB data access
- **Configuration Layer**: Spring and AWS client configuration

## Kafka Integration

Transaction events are automatically published to the `transaction-events` topic when records are created, enabling downstream services to react to inventory changes.

## Monitoring

- Health check endpoint at `/health`
- Spring Boot Actuator endpoints available
- Structured logging with configurable levels