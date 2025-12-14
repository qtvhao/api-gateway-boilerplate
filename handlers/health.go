package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger    *zap.Logger
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		startTime: time.Now(),
	}
}

// Health returns basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready returns readiness status (for Kubernetes readiness probe)
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if the service is ready to accept traffic
	// Add any additional checks here (database, external services, etc.)

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Live returns liveness status (for Kubernetes liveness probe)
func (h *HealthHandler) Live(c *gin.Context) {
	// Basic liveness check - service is running
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Status returns detailed status information
func (h *HealthHandler) Status(c *gin.Context) {
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, gin.H{
		"service":   "api-gateway",
		"status":    "healthy",
		"version":   "1.0.0",
		"uptime":    uptime.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SystemStatus returns detailed system status (admin only)
func (h *HealthHandler) SystemStatus(c *gin.Context) {
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, gin.H{
		"service":   "api-gateway",
		"status":    "healthy",
		"version":   "1.0.0",
		"uptime":    uptime.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"environment": gin.Mode(),
		"endpoints": gin.H{
			"project_management":     "configured",
			"goal_management":        "configured",
			"hr_management":          "configured",
			"engineering_analytics":  "configured",
			"workforce_wellbeing":    "configured",
		},
	})
}
