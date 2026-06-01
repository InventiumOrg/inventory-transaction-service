package models

// TransactionRecordRequest is the API payload for creating/updating a
// transaction record. Validation tags mirror the Java DTO's javax constraints.
type TransactionRecordRequest struct {
	Type              string `json:"type"              binding:"required"`
	Source            string `json:"source"            binding:"required"`
	Destination       string `json:"destination"       binding:"required"`
	InventoryId       string `json:"inventoryId"       binding:"required"`
	InventoryMeasure  string `json:"inventoryMeasure"  binding:"required"`
	InventoryCategory string `json:"inventoryCategory" binding:"required"`
	InventoryUnit     string `json:"inventoryUnit"     binding:"required"`
	Quantity          int32  `json:"quantity"          binding:"required,gt=0"`
	Status            string `json:"status"`
}
