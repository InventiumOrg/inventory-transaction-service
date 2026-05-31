package models

import (
	"testing"
	"time"
)

func TestToResponse(t *testing.T) {
	now := time.Now().UTC()
	in := TransactionRecord{
		Id:                "id-1",
		InventoryId:       "inv-1",
		Type:              "IMPORT",
		Source:            "warehouse",
		Destination:       "pos",
		InventoryMeasure:  "weight",
		InventoryCategory: "electronics",
		InventoryUnit:     "kg",
		Quantity:          10,
		Status:            "In Progress",
		CreatedAt:         now,
	}

	got := ToResponse(in)

	if got.Id != in.Id || got.InventoryId != in.InventoryId ||
		got.Type != in.Type || got.Quantity != in.Quantity ||
		got.Status != in.Status || !got.CreatedAt.Equal(now) {
		t.Fatalf("ToResponse field mismatch: got %#v, want %#v", got, in)
	}
}
