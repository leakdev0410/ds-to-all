package main

import (
	"log"
	"net/http"

	"deepseek-proxy/internal/config"
	"deepseek-proxy/internal/proxy"
)

func main() {
	cfg := config.LoadConfig()

	handler := proxy.NewProxyHandler(cfg)

	log.Printf("Starting DeepSeek AI Proxy server on port %s", cfg.Port)
	log.Printf("Listening for requests to translate Claude/Codex API to DeepSeek")
	if cfg.DeepSeekAPIKey == "" {
		log.Printf("Warning: DEEPSEEK_API_KEY environment variable is not set. The proxy will rely on the client's provided key.")
	} else {
		log.Printf("DEEPSEEK_API_KEY is configured. It will override client keys.")
	}

	err := http.ListenAndServe(":"+cfg.Port, handler)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
