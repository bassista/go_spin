package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig(confPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(confPath)

	// Defaults to allow running without config file
	viper.SetDefault("data.file_path", "./config/data/config.json")

	// Environment variables automatically override config file values
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GO_SPIN")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults and env vars")
		} else {
			log.Fatalf("Config file error: %v", err)
		}
	}
	// Environment variables like GO_SPIN_SERVER_PORT will override everything server.port
}
