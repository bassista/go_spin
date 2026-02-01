package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Default level
	Logger.SetLevel(logrus.InfoLevel)

	// Override from env, e.g., LOG_LEVEL=debug
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsedLevel, err := logrus.ParseLevel(strings.ToLower(level)); err == nil {
			Logger.SetLevel(parsedLevel)
		}
	}
}

// WithComponent adds a component field to the logger
func WithComponent(component string) *logrus.Entry {
	return Logger.WithField("component", component)
}
