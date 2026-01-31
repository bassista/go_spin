package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const ENV_PREFIX = "GO_SPIN"

// Config holds all application configuration (immutable after load)
type Config struct {
	Server ServerConfig
	Data   DataConfig
	Misc   MiscConfig
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutDownTimeout time.Duration
}

type DataConfig struct {
	FilePath        string
	PersistInterval time.Duration
}

type MiscConfig struct {
	GinMode           string
	SchedulingEnabled bool
	SchedulingPoll    time.Duration
	SchedulingTZ      string
	RuntimeType       string // "docker" o "memory"
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
	viper.SetDefault("server.port", 8084)
	viper.SetDefault("server.read_timeout_secs", 10)
	viper.SetDefault("server.write_timeout_secs", 10)
	viper.SetDefault("server.idle_timeout_secs", 120)
	viper.SetDefault("server.shutdown_timeout_secs", 5)
	viper.SetDefault("data.file_path", "./config/data/config.json")
	viper.SetDefault("data.persist_interval_secs", 5)
	viper.SetDefault("misc.gin_mode", "release")
	viper.SetDefault("misc.scheduling_enabled", true)
	viper.SetDefault("misc.scheduling_poll_interval_secs", 30)
	viper.SetDefault("misc.scheduling_timezone", "Local")
	viper.SetDefault("misc.runtime_type", "docker")

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
	port, err := getEnvOrViperPort("PORT", "server.port")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:            port,
			ReadTimeout:     time.Duration(viper.GetInt("server.read_timeout_secs")) * time.Second,
			WriteTimeout:    time.Duration(viper.GetInt("server.write_timeout_secs")) * time.Second,
			IdleTimeout:     time.Duration(viper.GetInt("server.idle_timeout_secs")) * time.Second,
			ShutDownTimeout: time.Duration(viper.GetInt("server.shutdown_timeout_secs")) * time.Second,
		},
		Data: DataConfig{
			FilePath:        viper.GetString("data.file_path"),
			PersistInterval: time.Duration(viper.GetInt("data.persist_interval_secs")) * time.Second,
		},
		Misc: MiscConfig{
			GinMode:           viper.GetString("misc.gin_mode"),
			SchedulingEnabled: viper.GetBool("misc.scheduling_enabled"),
			SchedulingPoll:    time.Duration(viper.GetInt("misc.scheduling_poll_interval_secs")) * time.Second,
			SchedulingTZ:      viper.GetString("misc.scheduling_timezone"),
			RuntimeType:       viper.GetString("misc.runtime_type"),
		},
	}

	// Fail-fast validation
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	fmt.Println("All configuration loaded successfully")

	return cfg, nil
}

// validate checks required configuration fields
func (c *Config) validate() error {
	if c.Data.FilePath == "" {
		return fmt.Errorf("data.file_path configuration is required")
	}
	if c.Data.PersistInterval <= 0 {
		return fmt.Errorf("data.persist_interval_secs must be positive")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be a valid TCP port (1-65535)")
	}
	if c.Server.ShutDownTimeout <= 0 {
		return fmt.Errorf("server.shutdown_timeout_secs must be positive")
	}
	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("server.idle_timeout_secs must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server.write_timeout_secs must be positive")
	}
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server.read_timeout_secs must be positive")
	}
	if c.Misc.SchedulingPoll <= 0 {
		return fmt.Errorf("misc.scheduling_poll_interval_secs must be positive")
	}
	if c.Misc.SchedulingTZ != "" && c.Misc.SchedulingTZ != "Local" {
		if _, err := time.LoadLocation(c.Misc.SchedulingTZ); err != nil {
			return fmt.Errorf("misc.scheduling_timezone is invalid: %w", err)
		}
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

func getEnvOrViperPort(envKey, viperKey string) (int, error) {
	if value := os.Getenv(envKey); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%s must be a valid integer TCP port: %w", envKey, err)
		}
		return port, nil
	}
	return viper.GetInt(viperKey), nil
}
