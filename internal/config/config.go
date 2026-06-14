package config

import (
	"os"
)

type Config struct {
	Port           string
	DeepSeekAPIKey string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	return &Config{
		Port:           port,
		DeepSeekAPIKey: apiKey,
	}
}
