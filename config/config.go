package config

import (
	"log"
	"os"
)

type Config struct {
	Port      string
	Providers map[string]ProviderConfig
}

type ProviderConfig struct {
	BaseURL string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	cfg := &Config{
		Port: port,
		Providers: map[string]ProviderConfig{
			"openai": {
				BaseURL: getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com"),
			},
			"anthropic": {
				BaseURL: "https://api.anthropic.com",
			},
			"google": {
				BaseURL: getEnvOrDefault("GOOGLE_BASE_URL", "https://generativelanguage.googleapis.com"),
			},
			"grok": {
				BaseURL: getEnvOrDefault("GROK_BASE_URL", "https://api.x.ai"),
			},
		},
	}

	log.Printf("配置加载完成 - 端口: %s", cfg.Port)
	for provider, config := range cfg.Providers {
		log.Printf("Provider %s: %s", provider, config.BaseURL)
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}