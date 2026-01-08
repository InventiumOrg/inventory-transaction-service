# Product Overview

## Inventory Transaction Service

A microservice responsible for creating and managing import/export records when inventory items are transferred into warehouses or point-of-sale systems. This service integrates with separate inventory, warehouse, and POS microservices to track item movements and maintain audit trails.

## Key Features
- Create import/export records for inventory transfers
- Track source and destination of imports/exports
- Integration with Kafka for event streaming
- DynamoDB storage for scalable record keeping
- REST API for external service integration