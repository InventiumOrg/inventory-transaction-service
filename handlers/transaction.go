package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"inventory-transaction-service/models"
	"inventory-transaction-service/processors"
	"inventory-transaction-service/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

const (
	statusInProgress = "In Progress"
	statusCompleted  = "Completed"
)

// CreateTransaction handles `POST /api/v1/transactions`.
func (h *Handlers) CreateTransaction(c *gin.Context) {
	ctx, span := h.Tracer.Start(c.Request.Context(), "CreateTransaction")
	defer span.End()

	var req models.TransactionRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	record := models.TransactionRecord{
		Id:                uuid.NewString(),
		InventoryId:       req.InventoryId,
		Type:              req.Type,
		Source:            req.Source,
		Destination:       req.Destination,
		InventoryMeasure:  req.InventoryMeasure,
		InventoryCategory: req.InventoryCategory,
		InventoryUnit:     req.InventoryUnit,
		Quantity:          req.Quantity,
		Status:            statusInProgress,
		CreatedAt:         time.Now().UTC(),
	}

	span.SetAttributes(
		attribute.String("transaction.id", record.Id),
		attribute.String("inventory.id", record.InventoryId),
		attribute.String("transaction.type", record.Type),
	)

	saved, err := h.saveRecord(ctx, "create", record)
	if err != nil {
		writeInternalError(c, "Failed to create transaction record", err)
		return
	}

	if h.Metrics != nil {
		h.Metrics.RecordTransactionOperation("create", saved.Type)
	}
	slog.Info("Transaction record created",
		slog.String("id", saved.Id),
		slog.String("inventoryId", saved.InventoryId))

	c.JSON(http.StatusCreated, models.ToResponse(saved))
}

// ListTransactions handles `GET /api/v1/transactions/:inventoryId`.
//
// Query params:
//   - limit  — page size (default 20, max 100)
//   - cursor — opaque pagination token returned by the previous response
func (h *Handlers) ListTransactions(c *gin.Context) {
	ctx, span := h.Tracer.Start(c.Request.Context(), "ListTransactions")
	defer span.End()

	inventoryId := c.Param("inventoryId")
	span.SetAttributes(attribute.String("inventory.id", inventoryId))

	limit := int32(20)
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = int32(n)
		}
	}
	cursor := c.Query("cursor")

	start := time.Now()
	records, nextCursor, err := h.Repo.FindByInventoryId(ctx, inventoryId, limit, cursor)
	if h.Metrics != nil {
		h.Metrics.RecordDBOperation("list", "transaction", time.Since(start), err)
	}
	if err != nil {
		writeInternalError(c, "Failed to list transaction records", err)
		return
	}

	data := make([]models.TransactionRecordResponse, len(records))
	for i, r := range records {
		data[i] = models.ToResponse(r)
	}

	c.JSON(http.StatusOK, models.ListTransactionResponse{
		Data:       data,
		Count:      len(data),
		NextCursor: nextCursor,
	})
}

// GetTransaction handles `GET /api/v1/transactions/:inventoryId/:id`.
func (h *Handlers) GetTransaction(c *gin.Context) {
	ctx, span := h.Tracer.Start(c.Request.Context(), "GetTransaction")
	defer span.End()

	inventoryId := c.Param("inventoryId")
	id := c.Param("id")
	span.SetAttributes(
		attribute.String("inventory.id", inventoryId),
		attribute.String("transaction.id", id),
	)

	start := time.Now()
	record, err := h.Repo.FindById(ctx, inventoryId, id)
	if h.Metrics != nil {
		h.Metrics.RecordDBOperation("get", "transaction", time.Since(start), err)
	}

	if err != nil {
		if errors.Is(err, repository.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Timestamp: time.Now().UTC(),
				Status:    http.StatusNotFound,
				Error:     "Not Found",
				Message:   "Transaction record not found with inventoryId: " + inventoryId + " and ID: " + id,
			})
			return
		}
		writeInternalError(c, "Failed to retrieve transaction record", err)
		return
	}

	c.JSON(http.StatusOK, models.ToResponse(record))
}

// UpdateTransaction handles `PUT /api/v1/transactions/:inventoryId/:id`.
//
// Mirrors the Java service: full-replace semantics, and best-effort Kafka
// publish of the resulting state. A Kafka failure does not fail the HTTP
// request.
func (h *Handlers) UpdateTransaction(c *gin.Context) {
	ctx, span := h.Tracer.Start(c.Request.Context(), "UpdateTransaction")
	defer span.End()

	inventoryId := c.Param("inventoryId")
	id := c.Param("id")

	var req models.TransactionRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	record := models.TransactionRecord{
		Id:                id,
		InventoryId:       req.InventoryId,
		Type:              req.Type,
		Source:            req.Source,
		Destination:       req.Destination,
		InventoryMeasure:  req.InventoryMeasure,
		InventoryCategory: req.InventoryCategory,
		InventoryUnit:     req.InventoryUnit,
		Quantity:          req.Quantity,
		Status:            req.Status,
		CreatedAt:         time.Now().UTC(),
	}
	span.SetAttributes(
		attribute.String("inventory.id", inventoryId),
		attribute.String("transaction.id", id),
		attribute.String("transaction.status", record.Status),
	)

	saved, err := h.saveRecord(ctx, "update", record)
	if err != nil {
		writeInternalError(c, "Failed to update transaction record", err)
		return
	}
	if h.Metrics != nil {
		h.Metrics.RecordTransactionOperation("update", saved.Type)
	}

	// Best-effort publish to Kafka only when the transaction reaches Completed.
	// Mirrors Java service: a Kafka failure does not fail the HTTP request.
	if h.Producer != nil && saved.Status == statusCompleted {
		event := processors.CreateInventoryImportEvent(saved)
		if err := h.Producer.Send(ctx, saved.InventoryId, event); err != nil {
			slog.Error("Failed to publish inventory status update",
				slog.String("inventoryId", saved.InventoryId),
				slog.Any("error", err))
		}
	}

	c.JSON(http.StatusOK, models.ToResponse(saved))
}

// saveRecord wraps repository persistence with metric instrumentation.
func (h *Handlers) saveRecord(ctx context.Context, op string, record models.TransactionRecord) (models.TransactionRecord, error) {
	start := time.Now()
	saved, err := h.Repo.Save(ctx, record)
	if h.Metrics != nil {
		h.Metrics.RecordDBOperation(op, "transaction", time.Since(start), err)
	}
	return saved, err
}

func writeValidationError(c *gin.Context, err error) {
	resp := models.ErrorResponse{
		Timestamp: time.Now().UTC(),
		Status:    http.StatusBadRequest,
		Error:     "Validation Failed",
		Message:   "Invalid input parameters",
	}

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = fe.Tag() + ": " + fe.Field()
		}
		resp.ValidationErrors = fields
	} else {
		resp.Message = err.Error()
	}

	c.JSON(http.StatusBadRequest, resp)
}

func writeInternalError(c *gin.Context, message string, err error) {
	slog.Error(message, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Timestamp: time.Now().UTC(),
		Status:    http.StatusInternalServerError,
		Error:     "Internal Server Error",
		Message:   message + ": " + err.Error(),
	})
}
