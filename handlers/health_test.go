package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestRouter() (*gin.Engine, *HealthHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger, _ := zap.NewDevelopment()
	handler := NewHealthHandler(logger)
	return router, handler
}

func TestHealth(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/health", handler.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestReady(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/health/ready", handler.Ready)

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestLive(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/health/live", handler.Live)

	req, _ := http.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestStatus(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/api/v1/public/status", handler.Status)

	req, _ := http.NewRequest("GET", "/api/v1/public/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", response["service"])
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotEmpty(t, response["uptime"])
}

func TestSystemStatus(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/api/v1/admin/system/status", handler.SystemStatus)

	req, _ := http.NewRequest("GET", "/api/v1/admin/system/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", response["service"])
	assert.NotNil(t, response["endpoints"])
}
