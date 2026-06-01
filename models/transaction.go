package models

import "time"

// TransactionRecord is the domain entity persisted to DynamoDB.
//
// The DynamoDB table is keyed by `InventoryId` (partition key) and `Id`
// (sort key), matching the Java service's `TransactionRecord` entity.
type TransactionRecord struct {
	InventoryId       string    `dynamodbav:"inventoryId"        json:"inventoryId"`
	Id                string    `dynamodbav:"id"                 json:"id"`
	Type              string    `dynamodbav:"type"               json:"type"`
	Source            string    `dynamodbav:"source"             json:"source"`
	Destination       string    `dynamodbav:"destination"        json:"destination"`
	InventoryMeasure  string    `dynamodbav:"inventoryMeasure"   json:"inventoryMeasure"`
	InventoryCategory string    `dynamodbav:"inventoryCategory"  json:"inventoryCategory"`
	InventoryUnit     string    `dynamodbav:"inventoryUnit"      json:"inventoryUnit"`
	Quantity          int32     `dynamodbav:"quantity"           json:"quantity"`
	Status            string    `dynamodbav:"status"             json:"status"`
	CreatedAt         time.Time `dynamodbav:"createdAt"          json:"createdAt"`
}
