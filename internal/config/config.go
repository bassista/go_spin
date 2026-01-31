package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const ENV_PREFIX = "GO_SPIN"

// Config holds all application configuration (immutable after load)
type Config struct {
	Server   ServerConfig
	Data     DataConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutDownTimeout time.Duration
}

type DataConfig struct {
	FilePath        string
	PersistInterval time.Duration
}

type SecurityConfig struct {
	APIKey  string
	GinMode string
}

// LoadConfig loads configuration from file, env vars and validates required fields.
// Returns error if validation fails (fail-fast).
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (that's okay in production)")
	}

	confPath := getEnvOrDefault(ENV_PREFIX+"_CONFIG_PATH", "./config")
	viper.AddConfigPath(confPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("server.port", "8084")
	viper.SetDefault("server.read_timeout_secs", 10)
	viper.SetDefault("server.write_timeout_secs", 10)
	viper.SetDefault("server.idle_timeout_secs", 120)
	viper.SetDefault("data.file_path", "./config/data/config.json")
	viper.SetDefault("data.persist_interval_secs", 5)

	// Environment variables automatically override config file values
	viper.AutomaticEnv()
	viper.SetEnvPrefix(ENV_PREFIX)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults and env vars")
		} else {
			return nil, fmt.Errorf("config file error: %w", err)
		}
	}

	// Build immutable config struct
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvOrViperString("PORT", "server.port"),
			ReadTimeout:     time.Duration(viper.GetInt("server.read_timeout_secs")) * time.Second,
			WriteTimeout:    time.Duration(viper.GetInt("server.write_timeout_secs")) * time.Second,
			IdleTimeout:     time.Duration(viper.GetInt("server.idle_timeout_secs")) * time.Second,
			ShutDownTimeout: time.Duration(viper.GetInt("server.shutdown_timeout_secs")) * time.Second,
		},
		Data: DataConfig{
			FilePath:        viper.GetString("data.file_path"),
			PersistInterval: time.Duration(viper.GetInt("data.persist_interval_secs")) * time.Second,
		},
		Security: SecurityConfig{
			APIKey:  os.Getenv("API_KEY"),
			GinMode: getEnvOrDefault("GIN_MODE", "release"),
		},
	}

	// Fail-fast validation
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	fmt.Println("All configuration loaded successfully!")

	return cfg, nil
}

// validate checks required configuration fields
func (c *Config) validate() error {
	if c.Security.APIKey == "" {
		return fmt.Errorf("API_KEY environment variable is required")
	}
	if c.Data.FilePath == "" {
		return fmt.Errorf("data.file_path configuration is required")
	}
	if c.Data.PersistInterval <= 0 {
		return fmt.Errorf("data.persist_interval_secs must be positive")
	}
	return nil
}

// getEnvOrDefault returns env var value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvOrViperString returns env var if set, otherwise viper config value
func getEnvOrViperString(envKey, viperKey string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return viper.GetString(viperKey)
}
