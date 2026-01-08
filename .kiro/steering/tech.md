# Technology Stack

## Framework & Language
- **Java 25** - Primary programming language (Amazon Correcto)
- **Spring Boot 4.0.1** - Application framework
- **Gradle** - Build system and dependency management

## Database & Storage
- **AWS DynamoDB** - NoSQL database for import records
- **DynamoDB Enhanced Client** - AWS SDK for DynamoDB operations

## Messaging & Streaming
- **Apache Kafka** - Event streaming platform
- **Spring Kafka** - Kafka integration for Spring Boot

## Key Dependencies
- Spring Boot Starter Web - REST API endpoints
- Spring Boot Starter Validation - Request validation
- AWS SDK DynamoDB - Database operations
- Spring Kafka - Message streaming

## Common Commands

### Build & Run (Gradle)
```bash
# Build the project
./gradlew build

# Run tests
./gradlew test

# Clean and build
./gradlew clean build

# Run application
./gradlew bootRun

# Run with specific profile
./gradlew bootRun --args='--spring.profiles.active=dev'
```

### Podman (if needed)
```bash
# Build Docker image
podman build -t inventory-transaction-service:1.0.0 .

# Run container
podman run -p 14330:14330 inventory-transaction-service
```

## Environment Variables
- `KAFKA_BOOTSTRAP_SERVERS` - Kafka broker addresses
- `AWS_REGION` - AWS region for DynamoDB
- `DYNAMODB_ENDPOINT` - DynamoDB endpoint (for local development)