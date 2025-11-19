package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Environment string
	ServiceName string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	RabbitMQ    RabbitMQConfig
	Auth        AuthConfig
	Logging     LoggingConfig
	Tracing     TracingConfig
	Metrics     MetricsConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxRequestSize  int64
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL          string
	Exchange     string
	ExchangeType string
	Queue        string
	RoutingKey   string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	Issuer        string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled     bool
	ServiceName string
	JaegerURL   string
	SampleRate  float64
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool
	Port    string
	Path    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("APP_ENV", "development"),
		ServiceName: getEnv("SERVICE_NAME", "airport-service"),

		Server: ServerConfig{
			Port:            getEnv("APP_PORT", "8080"),
			ReadTimeout:     parseDuration(getEnv("SERVER_READ_TIMEOUT", ""), 15*time.Second),
			WriteTimeout:    parseDuration(getEnv("SERVER_WRITE_TIMEOUT", ""), 15*time.Second),
			ShutdownTimeout: parseDuration(getEnv("SERVER_SHUTDOWN_TIMEOUT", ""), 30*time.Second),
			MaxRequestSize:  parseInt64(getEnv("SERVER_MAX_REQUEST_SIZE", ""), 10*1024*1024), // 10MB default
		},

		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "airport"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", ""), 25),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", ""), 5),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", ""), 5*time.Minute),
			ConnMaxIdleTime: parseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", ""), 5*time.Minute),
		},

		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           parseInt(getEnv("REDIS_DB", ""), 0),
			MaxRetries:   parseInt(getEnv("REDIS_MAX_RETRIES", ""), 3),
			PoolSize:     parseInt(getEnv("REDIS_POOL_SIZE", ""), 10),
			MinIdleConns: parseInt(getEnv("REDIS_MIN_IDLE_CONNS", ""), 2),
		},

		RabbitMQ: RabbitMQConfig{
			URL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:     getEnv("RABBITMQ_EXCHANGE", "airport-events"),
			ExchangeType: getEnv("RABBITMQ_EXCHANGE_TYPE", "topic"),
			Queue:        getEnv("RABBITMQ_QUEUE", ""),
			RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", ""),
		},

		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", ""),
			JWTExpiration: parseDuration(getEnv("JWT_EXPIRATION", ""), 24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "airport-services"),
		},

		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},

		Tracing: TracingConfig{
			Enabled:     parseBool(getEnv("TRACING_ENABLED", "false")),
			ServiceName: getEnv("TRACING_SERVICE_NAME", ""),
			JaegerURL:   getEnv("JAEGER_URL", "http://localhost:14268/api/traces"),
			SampleRate:  parseFloat64(getEnv("TRACING_SAMPLE_RATE", ""), 0.1),
		},

		Metrics: MetricsConfig{
			Enabled: parseBool(getEnv("METRICS_ENABLED", "true")),
			Port:    getEnv("METRICS_PORT", "9090"),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.IsProduction() {
		if c.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD is required in production")
		}
	}
	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

func parseInt64(value string, defaultValue int64) int64 {
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intValue
	}
	return defaultValue
}

func parseBool(value string) bool {
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	return false
}

func parseFloat64(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return defaultValue
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return defaultValue
}
