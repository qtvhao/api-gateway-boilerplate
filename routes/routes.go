package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ugjb/api-gateway/config"
	"github.com/ugjb/api-gateway/handlers"
	"github.com/ugjb/api-gateway/middleware"
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

	// Proxy root path and static assets to Web UI
	router.GET("/", proxy.ProxyWebUI())
	router.GET("/src/*filepath", proxy.ProxyWebUI())
	router.GET("/@vite/*filepath", proxy.ProxyWebUI())
	router.GET("/@react-refresh", proxy.ProxyWebUI())
	router.GET("/vite.svg", proxy.ProxyWebUI())
	router.GET("/assets/*filepath", proxy.ProxyWebUI())
	router.GET("/node_modules/*filepath", proxy.ProxyWebUI())

	// API version 1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication)
		public := v1.Group("/public")
		{
			public.GET("/status", health.Status)
		}

		// Project Management Service routes
		projectMgmt := v1.Group("/projects")
		projectMgmt.Use(middleware.AuthMiddleware(cfg))
		{
			// Task Management
			projectMgmt.POST("/tasks", proxy.ProxyToService("project_management", "/api/v1/tasks"))
			projectMgmt.GET("/tasks", proxy.ProxyToService("project_management", "/api/v1/tasks"))
			projectMgmt.GET("/tasks/:id", proxy.ProxyToService("project_management", "/api/v1/tasks/:id"))
			projectMgmt.PUT("/tasks/:id", proxy.ProxyToService("project_management", "/api/v1/tasks/:id"))
			projectMgmt.DELETE("/tasks/:id", proxy.ProxyToService("project_management", "/api/v1/tasks/:id"))
			projectMgmt.PATCH("/tasks/:id/status", proxy.ProxyToService("project_management", "/api/v1/tasks/:id/status"))

			// Sprint Management
			projectMgmt.POST("/sprints", proxy.ProxyToService("project_management", "/api/v1/sprints"))
			projectMgmt.GET("/sprints", proxy.ProxyToService("project_management", "/api/v1/sprints"))
			projectMgmt.GET("/sprints/:id", proxy.ProxyToService("project_management", "/api/v1/sprints/:id"))
			projectMgmt.PUT("/sprints/:id", proxy.ProxyToService("project_management", "/api/v1/sprints/:id"))
			projectMgmt.DELETE("/sprints/:id", proxy.ProxyToService("project_management", "/api/v1/sprints/:id"))
		}

		// Goal Management Service routes
		goalMgmt := v1.Group("/goals")
		goalMgmt.Use(middleware.AuthMiddleware(cfg))
		{
			// Objectives
			goalMgmt.POST("/objectives", proxy.ProxyToService("goal_management", "/api/v1/objectives"))
			goalMgmt.GET("/objectives", proxy.ProxyToService("goal_management", "/api/v1/objectives"))
			goalMgmt.GET("/objectives/:id", proxy.ProxyToService("goal_management", "/api/v1/objectives/:id"))
			goalMgmt.PUT("/objectives/:id", proxy.ProxyToService("goal_management", "/api/v1/objectives/:id"))
			goalMgmt.DELETE("/objectives/:id", proxy.ProxyToService("goal_management", "/api/v1/objectives/:id"))

			// Key Results
			goalMgmt.POST("/key-results", proxy.ProxyToService("goal_management", "/api/v1/key-results"))
			goalMgmt.GET("/key-results", proxy.ProxyToService("goal_management", "/api/v1/key-results"))
			goalMgmt.GET("/key-results/:id", proxy.ProxyToService("goal_management", "/api/v1/key-results/:id"))
			goalMgmt.PUT("/key-results/:id", proxy.ProxyToService("goal_management", "/api/v1/key-results/:id"))
			goalMgmt.DELETE("/key-results/:id", proxy.ProxyToService("goal_management", "/api/v1/key-results/:id"))
			goalMgmt.PATCH("/key-results/:id/progress", proxy.ProxyToService("goal_management", "/api/v1/key-results/:id/progress"))
		}

		// HR Management Service routes
		hrMgmt := v1.Group("/hr")
		hrMgmt.Use(middleware.AuthMiddleware(cfg))
		{
			// Employee Management
			hrMgmt.POST("/employees", proxy.ProxyToService("hr_management", "/api/v1/employees"))
			hrMgmt.GET("/employees", proxy.ProxyToService("hr_management", "/api/v1/employees"))
			hrMgmt.GET("/employees/:id", proxy.ProxyToService("hr_management", "/api/v1/employees/:id"))
			hrMgmt.PUT("/employees/:id", proxy.ProxyToService("hr_management", "/api/v1/employees/:id"))
			hrMgmt.DELETE("/employees/:id", proxy.ProxyToService("hr_management", "/api/v1/employees/:id"))

			// Resource Allocation
			hrMgmt.POST("/allocations", proxy.ProxyToService("hr_management", "/api/v1/allocations"))
			hrMgmt.GET("/allocations", proxy.ProxyToService("hr_management", "/api/v1/allocations"))
			hrMgmt.GET("/allocations/:id", proxy.ProxyToService("hr_management", "/api/v1/allocations/:id"))
			hrMgmt.PUT("/allocations/:id", proxy.ProxyToService("hr_management", "/api/v1/allocations/:id"))
			hrMgmt.DELETE("/allocations/:id", proxy.ProxyToService("hr_management", "/api/v1/allocations/:id"))
		}

		// Engineering Analytics Service routes
		analytics := v1.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(cfg))
		{
			// Metrics
			analytics.GET("/metrics", proxy.ProxyToService("engineering_analytics", "/api/v1/metrics"))
			analytics.GET("/metrics/team/:teamId", proxy.ProxyToService("engineering_analytics", "/api/v1/metrics/team/:teamId"))
			analytics.GET("/metrics/project/:projectId", proxy.ProxyToService("engineering_analytics", "/api/v1/metrics/project/:projectId"))

			// KPIs
			analytics.GET("/kpis", proxy.ProxyToService("engineering_analytics", "/api/v1/kpis"))
			analytics.GET("/kpis/:id", proxy.ProxyToService("engineering_analytics", "/api/v1/kpis/:id"))
			analytics.POST("/kpis", proxy.ProxyToService("engineering_analytics", "/api/v1/kpis"))

			// Dashboards
			analytics.GET("/dashboards", proxy.ProxyToService("engineering_analytics", "/api/v1/dashboards"))
			analytics.GET("/dashboards/:id", proxy.ProxyToService("engineering_analytics", "/api/v1/dashboards/:id"))
		}

		// Workforce Wellbeing Service routes
		wellbeing := v1.Group("/wellbeing")
		wellbeing.Use(middleware.AuthMiddleware(cfg))
		{
			// Wellbeing Metrics
			wellbeing.GET("/metrics", proxy.ProxyToService("workforce_wellbeing", "/api/v1/metrics"))
			wellbeing.GET("/metrics/employee/:employeeId", proxy.ProxyToService("workforce_wellbeing", "/api/v1/metrics/employee/:employeeId"))
			wellbeing.POST("/metrics", proxy.ProxyToService("workforce_wellbeing", "/api/v1/metrics"))

			// Burnout Predictions
			wellbeing.GET("/burnout/predictions", proxy.ProxyToService("workforce_wellbeing", "/api/v1/burnout/predictions"))
			wellbeing.GET("/burnout/predictions/:employeeId", proxy.ProxyToService("workforce_wellbeing", "/api/v1/burnout/predictions/:employeeId"))

			// Interventions
			wellbeing.GET("/interventions", proxy.ProxyToService("workforce_wellbeing", "/api/v1/interventions"))
			wellbeing.POST("/interventions", proxy.ProxyToService("workforce_wellbeing", "/api/v1/interventions"))
		}

		// Admin routes (require admin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.RequireRoles("admin", "system_admin"))
		{
			admin.GET("/users", proxy.ProxyToService("hr_management", "/api/v1/admin/users"))
			admin.GET("/system/status", health.SystemStatus)
		}
	}

	// Catch-all for undefined routes
	router.NoRoute(handlers.NotFound)
}
