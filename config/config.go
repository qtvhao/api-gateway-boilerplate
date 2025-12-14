package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Environment string        `mapstructure:"environment"`
	Port        int           `mapstructure:"port"`
	Server      ServerConfig  `mapstructure:"server"`
	JWT         JWTConfig     `mapstructure:"jwt"`
	RateLimit   RateLimitConfig `mapstructure:"rate_limit"`
	Redis       RedisConfig   `mapstructure:"redis"`
	CORS        CORSConfig    `mapstructure:"cors"`
	OPA         OPAConfig     `mapstructure:"opa"`
	Services    ServicesConfig `mapstructure:"services"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey      string        `mapstructure:"secret_key"`
	TokenDuration  time.Duration `mapstructure:"token_duration"`
	RefreshDuration time.Duration `mapstructure:"refresh_duration"`
	Issuer         string        `mapstructure:"issuer"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	RequestsPerMin  int           `mapstructure:"requests_per_min"`
	BurstSize       int           `mapstructure:"burst_size"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// OPAConfig holds Open Policy Agent configuration
type OPAConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	PolicyPath string `mapstructure:"policy_path"`
	BundleURL  string `mapstructure:"bundle_url"`
}

// ServicesConfig holds backend service endpoints
type ServicesConfig struct {
	ProjectManagement     ServiceEndpoint `mapstructure:"project_management"`
	GoalManagement        ServiceEndpoint `mapstructure:"goal_management"`
	HRManagement          ServiceEndpoint `mapstructure:"hr_management"`
	EngineeringAnalytics  ServiceEndpoint `mapstructure:"engineering_analytics"`
	WorkforceWellbeing    ServiceEndpoint `mapstructure:"workforce_wellbeing"`
	WebUI                 ServiceEndpoint `mapstructure:"web_ui"`
}

// ServiceEndpoint represents a backend service endpoint
type ServiceEndpoint struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/api-gateway")

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults and environment variables
	}

	// Override with environment variables
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// General
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8060)

	// Server
	viper.SetDefault("server.read_timeout", 15*time.Second)
	viper.SetDefault("server.write_timeout", 15*time.Second)
	viper.SetDefault("server.idle_timeout", 60*time.Second)

	// JWT
	viper.SetDefault("jwt.secret_key", "change-me-in-production")
	viper.SetDefault("jwt.token_duration", 15*time.Minute)
	viper.SetDefault("jwt.refresh_duration", 7*24*time.Hour)
	viper.SetDefault("jwt.issuer", "ugjb-api-gateway")

	// Rate Limiting
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_min", 100)
	viper.SetDefault("rate_limit.burst_size", 20)
	viper.SetDefault("rate_limit.cleanup_interval", 1*time.Minute)

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// CORS
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	viper.SetDefault("cors.expose_headers", []string{"Content-Length", "X-Request-ID"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 12*3600)

	// OPA
	viper.SetDefault("opa.enabled", true)
	viper.SetDefault("opa.policy_path", "./policies")
	viper.SetDefault("opa.bundle_url", "")

	// Backend Services
	viper.SetDefault("services.project_management.base_url", "http://localhost:8061")
	viper.SetDefault("services.project_management.timeout", 30*time.Second)

	viper.SetDefault("services.goal_management.base_url", "http://localhost:8062")
	viper.SetDefault("services.goal_management.timeout", 30*time.Second)

	viper.SetDefault("services.hr_management.base_url", "http://localhost:8063")
	viper.SetDefault("services.hr_management.timeout", 30*time.Second)

	viper.SetDefault("services.engineering_analytics.base_url", "http://localhost:8064")
	viper.SetDefault("services.engineering_analytics.timeout", 30*time.Second)

	viper.SetDefault("services.workforce_wellbeing.base_url", "http://localhost:8065")
	viper.SetDefault("services.workforce_wellbeing.timeout", 30*time.Second)

	viper.SetDefault("services.web_ui.base_url", "http://host.docker.internal:3000")
	viper.SetDefault("services.web_ui.timeout", 30*time.Second)
}

func validateConfig(cfg *Config) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", cfg.Port)
	}

	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key cannot be empty")
	}

	if cfg.Environment == "production" && cfg.JWT.SecretKey == "change-me-in-production" {
		return fmt.Errorf("JWT secret key must be changed in production")
	}

	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RequestsPerMin <= 0 {
			return fmt.Errorf("requests per minute must be positive")
		}
		if cfg.RateLimit.BurstSize <= 0 {
			return fmt.Errorf("burst size must be positive")
		}
	}

	return nil
}
