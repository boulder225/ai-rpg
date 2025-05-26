package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Context  ContextConfig  `json:"context"`
	AI       AIConfig       `json:"ai"`
	Logging  LoggingConfig  `json:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	CORS         CORSConfig    `json:"cors"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string        `json:"url"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	SSLMode         string        `json:"ssl_mode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL            string        `json:"url"`
	Password       string        `json:"password"`
	DB             int           `json:"db"`
	MaxRetries     int           `json:"max_retries"`
	DialTimeout    time.Duration `json:"dial_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	PoolSize       int           `json:"pool_size"`
	MinIdleConns   int           `json:"min_idle_conns"`
	MaxConnAge     time.Duration `json:"max_conn_age"`
	PoolTimeout    time.Duration `json:"pool_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
	Enabled        bool          `json:"enabled"`
}

// ContextConfig holds context manager configuration
type ContextConfig struct {
	MaxActions      int           `json:"max_actions"`
	CacheTimeout    time.Duration `json:"cache_timeout"`
	PersistInterval time.Duration `json:"persist_interval"`
	EventQueueSize  int           `json:"event_queue_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxContextAge   time.Duration `json:"max_context_age"`
}

// AIConfig holds AI integration configuration
type AIConfig struct {
	Provider           string        `json:"provider"`
	APIKey             string        `json:"api_key"`
	Model              string        `json:"model"`
	MaxTokens          int           `json:"max_tokens"`
	Temperature        float64       `json:"temperature"`
	Timeout            time.Duration `json:"timeout"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	RateLimitRequests  int           `json:"rate_limit_requests"`
	RateLimitDuration  time.Duration `json:"rate_limit_duration"`
	EnableCaching      bool          `json:"enable_caching"`
	CacheTTL           time.Duration `json:"cache_ttl"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8080),
			Host:         getEnvString("HOST", "0.0.0.0"),
			ReadTimeout:  getEnvDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
			CORS: CORSConfig{
				AllowedOrigins:   getEnvStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
				AllowedMethods:   getEnvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders:   getEnvStringSlice("CORS_ALLOWED_HEADERS", []string{"*"}),
				ExposedHeaders:   getEnvStringSlice("CORS_EXPOSED_HEADERS", []string{}),
				AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
				MaxAge:           getEnvInt("CORS_MAX_AGE", 86400),
			},
		},
		Database: DatabaseConfig{
			URL:             getEnvString("POSTGRES_URL", "postgres://rpguser:rpgpass@localhost:5432/rpgdb?sslmode=disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
			SSLMode:         getEnvString("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL:            getEnvString("REDIS_URL", "localhost:6379"),
			Password:       getEnvString("REDIS_PASSWORD", ""),
			DB:             getEnvInt("REDIS_DB", 0),
			MaxRetries:     getEnvInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:    getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:    getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:   getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:       getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:   getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxConnAge:     getEnvDuration("REDIS_MAX_CONN_AGE", 30*time.Minute),
			PoolTimeout:    getEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:    getEnvDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute),
			Enabled:        getEnvBool("REDIS_ENABLED", false),
		},
		Context: ContextConfig{
			MaxActions:      getEnvInt("CONTEXT_MAX_ACTIONS", 50),
			CacheTimeout:    getEnvDuration("CONTEXT_CACHE_TIMEOUT", 30*time.Minute),
			PersistInterval: getEnvDuration("CONTEXT_PERSIST_INTERVAL", 5*time.Minute),
			EventQueueSize:  getEnvInt("CONTEXT_EVENT_QUEUE_SIZE", 1000),
			CleanupInterval: getEnvDuration("CONTEXT_CLEANUP_INTERVAL", 6*time.Hour),
			MaxContextAge:   getEnvDuration("CONTEXT_MAX_AGE", 30*24*time.Hour), // 30 days
		},
		AI: AIConfig{
			Provider:           getEnvString("AI_PROVIDER", "claude"),
			APIKey:             getEnvString("AI_API_KEY", ""),
			Model:              getEnvString("AI_MODEL", "claude-3-sonnet-20240229"),
			MaxTokens:          getEnvInt("AI_MAX_TOKENS", 1000),
			Temperature:        getEnvFloat("AI_TEMPERATURE", 0.7),
			Timeout:            getEnvDuration("AI_TIMEOUT", 30*time.Second),
			MaxRetries:         getEnvInt("AI_MAX_RETRIES", 3),
			RetryDelay:         getEnvDuration("AI_RETRY_DELAY", 1*time.Second),
			RateLimitRequests:  getEnvInt("AI_RATE_LIMIT_REQUESTS", 60),
			RateLimitDuration:  getEnvDuration("AI_RATE_LIMIT_DURATION", 1*time.Minute),
			EnableCaching:      getEnvBool("AI_ENABLE_CACHING", true),
			CacheTTL:           getEnvDuration("AI_CACHE_TTL", 10*time.Minute),
		},
		Logging: LoggingConfig{
			Level:      getEnvString("LOG_LEVEL", "info"),
			Format:     getEnvString("LOG_FORMAT", "json"),
			Output:     getEnvString("LOG_OUTPUT", "stdout"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// For more complex parsing, you might want to use a JSON array
		return []string{value}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}
	
	if c.AI.APIKey == "" {
		return fmt.Errorf("AI API key is required")
	}
	
	if c.Context.MaxActions <= 0 {
		return fmt.Errorf("context max actions must be positive")
	}
	
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return getEnvString("ENV", "development") == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return getEnvString("ENV", "development") == "production"
}

// GetDatabaseConfig returns database configuration string
func (c *Config) GetDatabaseConfig() string {
	return c.Database.URL
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
