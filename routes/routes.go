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
			// Example: proxy to a backend service
			// protected.Any("/users/*path", proxy.ProxyToService("user_service"))
			// protected.Any("/orders/*path", proxy.ProxyToService("order_service"))
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

	// Catch-all for undefined routes
	router.NoRoute(handlers.NotFound)
}
