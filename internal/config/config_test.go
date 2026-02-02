package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:               8080,
			ReadTimeout:        10 * time.Second,
			WriteTimeout:       10 * time.Second,
			IdleTimeout:        120 * time.Second,
			ShutDownTimeout:    5 * time.Second,
			RequestTimeout:     1000 * time.Millisecond,
			CORSAllowedOrigins: "*",
		},
		Data: DataConfig{
			FilePath:          "/tmp/config.json",
			PersistInterval:   5 * time.Second,
			SchedulingEnabled: true,
			SchedulingPoll:    30 * time.Second,
		},
		Misc: MiscConfig{
			GinMode:      "release",
			SchedulingTZ: "Local",
			RuntimeType:  "docker",
		},
	}

	if err := cfg.validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestConfig_Validate_EmptyFilePath(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutDownTimeout: 5 * time.Second,
			RequestTimeout:  1000 * time.Millisecond,
		},
		Data: DataConfig{
			FilePath:        "",
			PersistInterval: 5 * time.Second,
			SchedulingPoll:  30 * time.Second,
		},
		Misc: MiscConfig{
			SchedulingTZ: "Local",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("expected error for empty file path")
	}
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"too high port", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port:            tt.port,
					ReadTimeout:     10 * time.Second,
					WriteTimeout:    10 * time.Second,
					IdleTimeout:     120 * time.Second,
					ShutDownTimeout: 5 * time.Second,
					RequestTimeout:  1000 * time.Millisecond,
				},
				Data: DataConfig{
					FilePath:        "/tmp/config.json",
					PersistInterval: 5 * time.Second,
					SchedulingPoll:  30 * time.Second,
				},
				Misc: MiscConfig{
					SchedulingTZ: "Local",
				},
			}

			err := cfg.validate()
			if err == nil {
				t.Errorf("expected error for port %d", tt.port)
			}
		})
	}
}

func TestConfig_Validate_InvalidPersistInterval(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutDownTimeout: 5 * time.Second,
			RequestTimeout:  1000 * time.Millisecond,
		},
		Data: DataConfig{
			FilePath:        "/tmp/config.json",
			PersistInterval: 0,
			SchedulingPoll:  30 * time.Second,
		},
		Misc: MiscConfig{
			SchedulingTZ: "Local",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("expected error for zero persist interval")
	}
}

func TestConfig_Validate_InvalidTimeouts(t *testing.T) {
	tests := []struct {
		name            string
		readTimeout     time.Duration
		writeTimeout    time.Duration
		idleTimeout     time.Duration
		shutdownTimeout time.Duration
	}{
		{"zero read timeout", 0, 10 * time.Second, 120 * time.Second, 5 * time.Second},
		{"zero write timeout", 10 * time.Second, 0, 120 * time.Second, 5 * time.Second},
		{"zero idle timeout", 10 * time.Second, 10 * time.Second, 0, 5 * time.Second},
		{"zero shutdown timeout", 10 * time.Second, 10 * time.Second, 120 * time.Second, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     tt.readTimeout,
					WriteTimeout:    tt.writeTimeout,
					IdleTimeout:     tt.idleTimeout,
					ShutDownTimeout: tt.shutdownTimeout,
					RequestTimeout:  1000 * time.Millisecond,
				},
				Data: DataConfig{
					FilePath:        "/tmp/config.json",
					PersistInterval: 5 * time.Second,
					SchedulingPoll:  30 * time.Second,
				},
				Misc: MiscConfig{
					SchedulingTZ: "Local",
				},
			}

			err := cfg.validate()
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestConfig_Validate_InvalidSchedulingPoll(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutDownTimeout: 5 * time.Second,
			RequestTimeout:  1000 * time.Millisecond,
		},
		Data: DataConfig{
			FilePath:        "/tmp/config.json",
			PersistInterval: 5 * time.Second,
			SchedulingPoll:  0,
		},
		Misc: MiscConfig{
			SchedulingTZ: "Local",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("expected error for zero scheduling poll")
	}
}

func TestConfig_Validate_InvalidTimezone(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutDownTimeout: 5 * time.Second,
			RequestTimeout:  1000 * time.Millisecond,
		},
		Data: DataConfig{
			FilePath:        "/tmp/config.json",
			PersistInterval: 5 * time.Second,
			SchedulingPoll:  30 * time.Second,
		},
		Misc: MiscConfig{
			SchedulingTZ: "Invalid/Timezone",
		},
	}

	err := cfg.validate()
	if err == nil {
		t.Error("expected error for invalid timezone")
	}
}

func TestConfig_Validate_ValidTimezones(t *testing.T) {
	timezones := []string{"Local", "UTC", "America/New_York", "Europe/Rome"}

	for _, tz := range timezones {
		t.Run(tz, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     10 * time.Second,
					WriteTimeout:    10 * time.Second,
					IdleTimeout:     120 * time.Second,
					ShutDownTimeout: 5 * time.Second,
					RequestTimeout:  1000 * time.Millisecond,
				},
				Data: DataConfig{
					FilePath:        "/tmp/config.json",
					PersistInterval: 5 * time.Second,
					SchedulingPoll:  30 * time.Second,
				},
				Misc: MiscConfig{
					SchedulingTZ: tz,
				},
			}

			if err := cfg.validate(); err != nil {
				t.Errorf("expected valid timezone %s, got error: %v", tz, err)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with env var set
	_ = os.Setenv("TEST_ENV_VAR", "custom_value")
	defer func() { _ = os.Unsetenv("TEST_ENV_VAR") }()

	result := getEnvOrDefault("TEST_ENV_VAR", "default_value")
	if result != "custom_value" {
		t.Errorf("expected 'custom_value', got '%s'", result)
	}

	// Test with env var not set
	result = getEnvOrDefault("NONEXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("expected 'default_value', got '%s'", result)
	}
}

func TestGetEnvOrViperPort_FromEnv(t *testing.T) {
	_ = os.Setenv("TEST_PORT", "9090")
	defer func() { _ = os.Unsetenv("TEST_PORT") }()

	port, err := getEnvOrViperPort("TEST_PORT", "server.port")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if port != 9090 {
		t.Errorf("expected 9090, got %d", port)
	}
}

func TestGetEnvOrViperPort_InvalidEnv(t *testing.T) {
	_ = os.Setenv("TEST_PORT_INVALID", "not_a_number")
	defer func() { _ = os.Unsetenv("TEST_PORT_INVALID") }()

	_, err := getEnvOrViperPort("TEST_PORT_INVALID", "server.port")
	if err == nil {
		t.Error("expected error for invalid port")
	}
}
