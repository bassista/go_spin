package logger

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestWithComponent(t *testing.T) {
	entry := WithComponent("test-component")
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	// Check that the component field is set
	if val, ok := entry.Data["component"]; !ok {
		t.Error("expected component field to be set")
	} else if val != "test-component" {
		t.Errorf("expected component 'test-component', got '%v'", val)
	}
}

func TestLoggerInit(t *testing.T) {
	// Test that Logger is initialized
	if Logger == nil {
		t.Fatal("expected Logger to be initialized")
	}

	// Test that Logger has the expected output
	if Logger.Out != os.Stdout {
		t.Error("expected Logger output to be os.Stdout")
	}
}

func TestLoggerInitWithEnvLogLevel(t *testing.T) {
	// Save original value
	origLevel := Logger.GetLevel()

	tests := []struct {
		name          string
		envValue      string
		expectedLevel logrus.Level
	}{
		{"debug level", "debug", logrus.DebugLevel},
		{"info level", "info", logrus.InfoLevel},
		{"warn level", "warn", logrus.WarnLevel},
		{"error level", "error", logrus.ErrorLevel},
		{"DEBUG uppercase", "DEBUG", logrus.DebugLevel},
		{"invalid level", "invalid", origLevel}, // should keep original
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original
			Logger.SetLevel(logrus.InfoLevel)

			// Set env var
			if tt.envValue != "" {
				_ = os.Setenv("LOG_LEVEL", tt.envValue)
			}

			// Simulate init logic
			if level := os.Getenv("LOG_LEVEL"); level != "" {
				if parsedLevel, err := logrus.ParseLevel(level); err == nil {
					Logger.SetLevel(parsedLevel)
				}
			}

			if tt.envValue != "invalid" {
				if Logger.GetLevel() != tt.expectedLevel {
					t.Errorf("expected level %v, got %v", tt.expectedLevel, Logger.GetLevel())
				}
			}

			// Cleanup
			_ = os.Unsetenv("LOG_LEVEL")
		})
	}

	// Restore original level
	Logger.SetLevel(origLevel)
}

func TestWithComponentMultiple(t *testing.T) {
	entry1 := WithComponent("component-a")
	entry2 := WithComponent("component-b")

	if entry1.Data["component"] == entry2.Data["component"] {
		t.Error("expected different component values for different entries")
	}
}
