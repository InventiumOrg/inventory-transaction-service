# Spring Boot Development Guidelines
# This is a repository for an inventory transaction service written in Java that will do the following:
- Create import records everytime an inventory item is being taken into a warehouse or point of sales
(inventory, warehouse, p.o.s are different microservices)
# A typical import entity should have the following field:
- id
- source
- destination
- createdAt
- inventoryID
- amount
- status
- type

## Component Structure
- Use @RestController for REST endpoints
- Use @Service for business logic
- Use @Repository for data access
- Use @Component for other beans

## Database usage
- DynamoDB from AWS

## Dependency Injection
- Prefer constructor injection over field injection
- Use final fields for injected dependencies
- Avoid circular dependencies

## API Design
- Follow REST principles for endpoint design
- Use appropriate HTTP methods (GET, POST, PUT, DELETE)
- Return appropriate HTTP status codes
- Use DTOs for request/response objects

## Data Streaming
- Kafka