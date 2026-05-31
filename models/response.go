package models

import "time"

// TransactionRecordResponse is what gets returned from the API.
type TransactionRecordResponse struct {
	Id                string    `json:"id"`
	Type              string    `json:"type"`
	Source            string    `json:"source"`
	Destination       string    `json:"destination"`
	CreatedAt         time.Time `json:"createdAt"`
	InventoryId       string    `json:"inventoryId"`
	InventoryMeasure  string    `json:"inventoryMeasure"`
	InventoryCategory string    `json:"inventoryCategory"`
	InventoryUnit     string    `json:"inventoryUnit"`
	Quantity          int32     `json:"quantity"`
	Status            string    `json:"status"`
}

// ToResponse converts a TransactionRecord into its API response form.
func ToResponse(t TransactionRecord) TransactionRecordResponse {
	return TransactionRecordResponse{
		Id:                t.Id,
		Type:              t.Type,
		Source:            t.Source,
		Destination:       t.Destination,
		CreatedAt:         t.CreatedAt,
		InventoryId:       t.InventoryId,
		InventoryMeasure:  t.InventoryMeasure,
		InventoryCategory: t.InventoryCategory,
		InventoryUnit:     t.InventoryUnit,
		Quantity:          t.Quantity,
		Status:            t.Status,
	}
}

// ErrorResponse mirrors the Java GlobalExceptionHandler response shape.
type ErrorResponse struct {
	Timestamp        time.Time         `json:"timestamp"`
	Status           int               `json:"status"`
	Error            string            `json:"error"`
	Message          string            `json:"message"`
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
}
