package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ugjb/api-gateway/config"
	"go.uber.org/zap"
)

// ProxyHandler handles reverse proxy operations
type ProxyHandler struct {
	config  *config.Config
	logger  *zap.Logger
	proxies map[string]*httputil.ReverseProxy
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(cfg *config.Config, logger *zap.Logger) *ProxyHandler {
	handler := &ProxyHandler{
		config:  cfg,
		logger:  logger,
		proxies: make(map[string]*httputil.ReverseProxy),
	}

	// Initialize proxies for each backend service
	handler.initProxies()

	return handler
}

// initProxies initializes reverse proxies for all backend services
func (p *ProxyHandler) initProxies() {
	services := map[string]string{
		"project_management":    p.config.Services.ProjectManagement.BaseURL,
		"goal_management":       p.config.Services.GoalManagement.BaseURL,
		"hr_management":         p.config.Services.HRManagement.BaseURL,
		"engineering_analytics": p.config.Services.EngineeringAnalytics.BaseURL,
		"workforce_wellbeing":   p.config.Services.WorkforceWellbeing.BaseURL,
		"web_ui":                p.config.Services.WebUI.BaseURL,
	}

	for serviceName, baseURL := range services {
		target, err := url.Parse(baseURL)
		if err != nil {
			p.logger.Error("Failed to parse service URL",
				zap.String("service", serviceName),
				zap.String("url", baseURL),
				zap.Error(err),
			)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		// Customize the director to modify the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			p.modifyRequest(req, target)
		}

		// Custom error handler
		proxy.ErrorHandler = p.errorHandler

		// Custom response modifier
		proxy.ModifyResponse = p.modifyResponse

		p.proxies[serviceName] = proxy
	}
}

// modifyRequest modifies the request before sending to backend service
func (p *ProxyHandler) modifyRequest(req *http.Request, target *url.URL) {
	req.Host = target.Host
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme

	// Add/forward headers
	req.Header.Set("X-Forwarded-Host", req.Host)
	req.Header.Set("X-Origin-Host", target.Host)

	// Forward original client IP
	if clientIP := req.Header.Get("X-Real-IP"); clientIP == "" {
		req.Header.Set("X-Real-IP", req.RemoteAddr)
	}

	// Add gateway identifier
	req.Header.Set("X-Gateway", "ugjb-api-gateway")
}

// modifyResponse modifies the response from backend service
func (p *ProxyHandler) modifyResponse(resp *http.Response) error {
	// Add custom headers to response
	resp.Header.Set("X-Gateway", "ugjb-api-gateway")

	return nil
}

// errorHandler handles errors from the reverse proxy
func (p *ProxyHandler) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	p.logger.Error("Proxy error",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.Error(err),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)

	response := fmt.Sprintf(`{"error":"Bad Gateway","message":"Failed to reach backend service: %s"}`, err.Error())
	w.Write([]byte(response))
}

// ProxyToService returns a handler that proxies requests to a specific backend service
func (p *ProxyHandler) ProxyToService(serviceName, targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy, exists := p.proxies[serviceName]
		if !exists {
			p.logger.Error("Proxy not found for service", zap.String("service", serviceName))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Service configuration not found",
			})
			return
		}

		// Log the proxy request
		p.logger.Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("target", targetPath),
		)

		// Replace path parameters in target path
		finalPath := p.replacePathParams(targetPath, c)

		// Store original path and set new path for backend
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = finalPath

		// Set timeout for backend request
		timeout := p.getServiceTimeout(serviceName)
		c.Request = c.Request.WithContext(c.Request.Context())

		// Add timeout handling
		done := make(chan bool, 1)
		go func() {
			proxy.ServeHTTP(c.Writer, c.Request)
			done <- true
		}()

		select {
		case <-done:
			// Request completed successfully
		case <-time.After(timeout):
			p.logger.Error("Backend request timeout",
				zap.String("service", serviceName),
				zap.String("path", originalPath),
				zap.Duration("timeout", timeout),
			)
			if !c.Writer.Written() {
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"error":   "Gateway Timeout",
					"message": "Backend service did not respond in time",
				})
			}
		}
	}
}

// replacePathParams replaces path parameters (e.g., :id) with actual values from context
func (p *ProxyHandler) replacePathParams(path string, c *gin.Context) string {
	// Get all path parameters
	for _, param := range c.Params {
		placeholder := ":" + param.Key
		path = strings.ReplaceAll(path, placeholder, param.Value)
	}
	return path
}

// getServiceTimeout returns the configured timeout for a service
func (p *ProxyHandler) getServiceTimeout(serviceName string) time.Duration {
	switch serviceName {
	case "project_management":
		return p.config.Services.ProjectManagement.Timeout
	case "goal_management":
		return p.config.Services.GoalManagement.Timeout
	case "hr_management":
		return p.config.Services.HRManagement.Timeout
	case "engineering_analytics":
		return p.config.Services.EngineeringAnalytics.Timeout
	case "workforce_wellbeing":
		return p.config.Services.WorkforceWellbeing.Timeout
	case "web_ui":
		return p.config.Services.WebUI.Timeout
	default:
		return 30 * time.Second
	}
}

// ProxyWebUI returns a handler that proxies all requests to the Web UI
func (p *ProxyHandler) ProxyWebUI() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy, exists := p.proxies["web_ui"]
		if !exists {
			p.logger.Error("Proxy not found for web_ui service")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Web UI service configuration not found",
			})
			return
		}

		// Log the proxy request
		p.logger.Debug("Proxying request to Web UI",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// NotFound handles 404 errors
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Not Found",
		"message": "The requested endpoint does not exist",
		"path":    c.Request.URL.Path,
	})
}

// MethodNotAllowed handles 405 errors
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error":   "Method Not Allowed",
		"message": fmt.Sprintf("Method %s is not allowed for this endpoint", c.Request.Method),
		"path":    c.Request.URL.Path,
	})
}

// logRequestBody logs the request body for debugging (use carefully in production)
func (p *ProxyHandler) logRequestBody(req *http.Request) {
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			p.logger.Debug("Request body", zap.String("body", string(bodyBytes)))
			// Restore the body for the actual request
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
}
