package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ugjb/api-gateway/config"
)

// RateLimiter manages rate limiting
type RateLimiter struct {
	config      *config.Config
	redisClient *redis.Client
	localLimits map[string]*clientLimit
	mu          sync.RWMutex
	useRedis    bool
}

// clientLimit tracks requests for a client using token bucket algorithm
type clientLimit struct {
	tokens       int
	lastRefill   time.Time
	mu           sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.Config) (*RateLimiter, error) {
	rl := &RateLimiter{
		config:      cfg,
		localLimits: make(map[string]*clientLimit),
	}

	// Try to connect to Redis for distributed rate limiting
	if cfg.Redis.Host != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err == nil {
			rl.redisClient = redisClient
			rl.useRedis = true
		}
		// If Redis is unavailable, fall back to in-memory rate limiting
	}

	// Start cleanup goroutine for local limits
	if !rl.useRedis {
		go rl.cleanupRoutine()
	}

	return rl, nil
}

// Close closes the rate limiter resources
func (rl *RateLimiter) Close() error {
	if rl.redisClient != nil {
		return rl.redisClient.Close()
	}
	return nil
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.RateLimit.Enabled {
			c.Next()
			return
		}

		// Get client identifier (IP address or user ID)
		clientID := rl.getClientID(c)

		allowed, remaining, resetTime, err := rl.allow(c.Request.Context(), clientID)
		if err != nil {
			// Log error but don't fail the request
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RateLimit.RequestsPerMin))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request should be allowed based on rate limits
func (rl *RateLimiter) allow(ctx context.Context, clientID string) (bool, int, time.Time, error) {
	if rl.useRedis {
		return rl.allowRedis(ctx, clientID)
	}
	return rl.allowLocal(clientID)
}

// allowRedis implements distributed rate limiting using Redis
func (rl *RateLimiter) allowRedis(ctx context.Context, clientID string) (bool, int, time.Time, error) {
	key := fmt.Sprintf("ratelimit:%s", clientID)
	window := time.Minute
	limit := int64(rl.config.RateLimit.RequestsPerMin)

	now := time.Now()
	windowStart := now.Truncate(window)

	pipe := rl.redisClient.Pipeline()

	// Increment counter
	incr := pipe.Incr(ctx, key)

	// Set expiry on first request
	pipe.ExpireAt(ctx, key, windowStart.Add(window))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := incr.Val()
	remaining := int(limit - count)
	if remaining < 0 {
		remaining = 0
	}

	resetTime := windowStart.Add(window)
	allowed := count <= limit

	return allowed, remaining, resetTime, nil
}

// allowLocal implements local in-memory rate limiting using token bucket
func (rl *RateLimiter) allowLocal(clientID string) (bool, int, time.Time, error) {
	rl.mu.Lock()
	limit, exists := rl.localLimits[clientID]
	if !exists {
		limit = &clientLimit{
			tokens:     rl.config.RateLimit.RequestsPerMin,
			lastRefill: time.Now(),
		}
		rl.localLimits[clientID] = limit
	}
	rl.mu.Unlock()

	limit.mu.Lock()
	defer limit.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(limit.lastRefill)

	// Refill tokens based on elapsed time
	if elapsed >= time.Minute {
		limit.tokens = rl.config.RateLimit.RequestsPerMin
		limit.lastRefill = now
	} else {
		tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RateLimit.RequestsPerMin))
		limit.tokens += tokensToAdd
		if limit.tokens > rl.config.RateLimit.RequestsPerMin {
			limit.tokens = rl.config.RateLimit.RequestsPerMin
		}
		if tokensToAdd > 0 {
			limit.lastRefill = now
		}
	}

	// Check if request can be allowed
	allowed := limit.tokens > 0
	if allowed {
		limit.tokens--
	}

	remaining := limit.tokens
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time
	resetTime := limit.lastRefill.Add(time.Minute)

	return allowed, remaining, resetTime, nil
}

// getClientID returns a unique identifier for the client
func (rl *RateLimiter) getClientID(c *gin.Context) string {
	// Prefer user ID if authenticated
	if claims, ok := GetUserFromContext(c); ok {
		return fmt.Sprintf("user:%s", claims.UserID)
	}

	// Fall back to IP address
	// Check X-Forwarded-For header for proxy scenarios
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return fmt.Sprintf("ip:%s", xff)
	}

	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// cleanupRoutine periodically cleans up old entries from local limits
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.config.RateLimit.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes stale entries from local limits
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, limit := range rl.localLimits {
		limit.mu.Lock()
		// Remove entries that haven't been accessed in 10 minutes
		if now.Sub(limit.lastRefill) > 10*time.Minute {
			delete(rl.localLimits, clientID)
		}
		limit.mu.Unlock()
	}
}
