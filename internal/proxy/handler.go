package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"deepseek-proxy/internal/config"
)

func NewProxyHandler(cfg *config.Config) http.Handler {
	deepseekURL, err := url.Parse("https://api.deepseek.com")
	if err != nil {
		log.Fatalf("failed to parse deepseek url: %v", err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Set the target URL
			req.URL.Scheme = deepseekURL.Scheme
			req.URL.Host = deepseekURL.Host

			// Update Authorization/API Key if provided in env
			if cfg.DeepSeekAPIKey != "" {
				// Anthropic format uses x-api-key
				if req.Header.Get("x-api-key") != "" {
					req.Header.Set("x-api-key", cfg.DeepSeekAPIKey)
				}
				// OpenAI format uses Authorization: Bearer
				if req.Header.Get("Authorization") != "" || strings.HasPrefix(req.URL.Path, "/v1/") {
					req.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)
				}
			}

			// If it's Anthropic, prepend /anthropic to path if not present (to map to DeepSeek's anthropic endpoint)
			// Claude code typically sends requests to /v1/messages
			if strings.HasPrefix(req.URL.Path, "/v1/messages") {
				req.URL.Path = "/anthropic" + req.URL.Path
				if req.Header.Get("x-api-key") == "" && cfg.DeepSeekAPIKey != "" {
					req.Header.Set("x-api-key", cfg.DeepSeekAPIKey)
				}
			}

			// Intercept and modify the request body to enforce deepseek-v4-pro
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil && len(bodyBytes) > 0 {
					var bodyData map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
						// Override model
						bodyData["model"] = "deepseek-v4-pro"

						// Marshal back
						newBodyBytes, err := json.Marshal(bodyData)
						if err == nil {
							req.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
							req.ContentLength = int64(len(newBodyBytes))
							req.Header.Set("Content-Length", strconv.Itoa(len(newBodyBytes)))
						} else {
							req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
						}
					} else {
						req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				}
			}

			// We shouldn't send the original host header
			req.Host = deepseekURL.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			// Allows us to modify the response if needed, but standard proxying is fine for both streaming and non-streaming
			return nil
		},
	}

	return proxy
}
