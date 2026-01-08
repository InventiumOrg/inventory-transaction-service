# Project Structure

## Gradle Standard Directory Layout
```
src/
├── main/
│   ├── java/com/inventium/inventory_transaction_service/
│   │   ├── ImportServiceApplication.java     # Main Spring Boot application
│   │   ├── controller/                       # REST controllers (@RestController)
│   │   │   └── ImportRecordController.java
│   │   ├── service/                          # Business logic (@Service)
│   │   │   ├── ImportRecordService.java
│   │   │   └── KafkaProducerService.java
│   │   ├── repository/                       # Data access (@Repository)
│   │   │   └── ImportRecordRepository.java
│   │   ├── entity/                           # DynamoDB entities
│   │   │   └── ImportRecord.java
│   │   ├── dto/                              # Data Transfer Objects
│   │   │   ├── ImportRecordRequest.java
│   │   │   └── ImportRecordResponse.java
│   │   └── config/                           # Configuration classes
│   │       ├── DynamoDbConfig.java
│   │       └── KafkaConfig.java
│   └── resources/
│       └── application.yml                   # Application configuration
└── test/
    └── java/com/inventium/import_service/
        └── ImportServiceApplicationTests.java
```

## Package Organization

### Controller Layer (`controller/`)
- REST endpoints following RESTful principles
- Request/response handling and validation
- HTTP status code management

### Service Layer (`service/`)
- Business logic implementation
- Transaction management
- Integration with external services (Kafka)

### Repository Layer (`repository/`)
- Data access operations
- DynamoDB interactions using Enhanced Client
- Query implementations

### Entity Layer (`entity/`)
- DynamoDB table mappings
- Domain model representations
- Database annotations

### DTO Layer (`dto/`)
- Request/Response objects for API
- Data validation annotations
- Separation from internal entities

### Configuration (`config/`)
- Spring configuration classes
- Bean definitions for AWS and Kafka clients
- Environment-specific settings

## Naming Conventions
- Classes: PascalCase (e.g., `ImportRecordService, ExportRecordService`)
- Methods: camelCase (e.g., `createImportRecord`)
- Constants: UPPER_SNAKE_CASE (e.g., `IMPORT_TOPIC`)
- Packages: lowercase (e.g., `com.inventium.inventory_transaction_service`)