package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/api-gateway/config"
	"github.com/api-gateway/handlers"
	"github.com/api-gateway/middleware"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for the API Gateway
func SetupRoutes(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	// Health check endpoints (no authentication required)
	health := handlers.NewHealthHandler(logger)
	router.GET("/health", health.Health)
	router.GET("/health/ready", health.Ready)
	router.GET("/health/live", health.Live)

	// Create proxy handler
	proxy := handlers.NewProxyHandler(cfg, logger)

	// ============================================
	// External Services (no authentication)
	// Configure these in config.yaml under external_services
	// ============================================

	// Ollama LLM API routes
	router.POST("/api/generate", proxy.ProxyToExternalServiceWithPath("ollama", "/api/generate"))
	router.POST("/api/chat", proxy.ProxyToExternalServiceWithPath("ollama", "/api/chat"))
	router.POST("/api/embeddings", proxy.ProxyToExternalServiceWithPath("ollama", "/api/embeddings"))

	// Docker Registry V2 API
	router.Any("/v2/*path", proxy.ProxyToExternalService("docker_registry"))

	// ============================================
	// API Routes
	// ============================================

	// API version 1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication)
		public := v1.Group("/public")
		{
			public.GET("/status", health.Status)
		}

		// Protected routes (authentication required)
		// Add your authenticated routes here
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Example: proxy to a backend service (configure in config.yaml under services)
			_ = proxy // proxy handler available for use
		}

		// Admin routes (require admin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.RequireRoles("admin"))
		{
			admin.GET("/system/status", health.SystemStatus)
		}
	}

	// ============================================
	// Frontend Catch-all (WebUI proxy)
	// ============================================
	// Proxies all unmatched routes to the frontend dev server (e.g., Vite)
	// Supports WebSocket upgrades for HMR (Hot Module Replacement)
	router.NoRoute(proxy.ProxyWithWebSocket("frontend"))
}
