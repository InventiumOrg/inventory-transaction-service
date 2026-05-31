package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// Recovery is a structured-logging panic recovery middleware. It logs the
// panic with a stack trace via slog and returns a 500 ErrorResponse-style body.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered in HTTP handler",
					slog.Any("panic", r),
					slog.String("path", c.Request.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"status":    http.StatusInternalServerError,
					"error":     "Internal Server Error",
					"message":   "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}
