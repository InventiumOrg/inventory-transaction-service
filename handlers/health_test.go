package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	router := gin.New()
	router.GET("/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["status"] != "UP" {
		t.Fatalf(`expected status "UP", got %q`, body["status"])
	}
	if body["service"] != "inventory-transaction-service" {
		t.Fatalf(`unexpected service: %q`, body["service"])
	}
}
