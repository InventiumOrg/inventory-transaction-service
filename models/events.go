package models

// InventoryImportEvent is the Avro-serialized payload published to the
// `inventory.transaction.status.updated` topic. Field names match the registered
// Avro schema (lowerCamelCase to mirror Java producer behavior).
type InventoryImportEvent struct {
	InventoryID       string `json:"inventoryId"`
	InventoryMeasure  string `json:"inventoryMeasure"`
	InventoryCategory string `json:"inventoryCategory"`
	InventoryUnit     string `json:"inventoryUnit"`
	Quantity          int    `json:"quantity"`
	Type              string `json:"type"`
}
