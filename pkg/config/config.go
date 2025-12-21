// Package config provides configuration management for the coordination engine.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port        int    `json:"port"`
	MetricsPort int    `json:"metrics_port"`
	LogLevel    string `json:"log_level"`

	// Kubernetes configuration
	Kubeconfig string `json:"kubeconfig,omitempty"`
	Namespace  string `json:"namespace"`

	// External service URLs
	MLServiceURL string `json:"ml_service_url"`
	ArgocdAPIURL string `json:"argocd_api_url,omitempty"` // Optional, auto-detected

	// HTTP client configuration
	HTTPTimeout time.Duration `json:"http_timeout"`

	// Feature flags
	EnableCORS      bool     `json:"enable_cors"`
	CORSAllowOrigin []string `json:"cors_allow_origin,omitempty"`

	// Performance tuning
	KubernetesQPS   float32 `json:"kubernetes_qps"`
	KubernetesBurst int     `json:"kubernetes_burst"`
}

// Default configuration values
const (
	DefaultPort            = 8080
	DefaultMetricsPort     = 9090
	DefaultLogLevel        = "info"
	DefaultNamespace       = "self-healing-platform"
	DefaultMLServiceURL    = "http://aiops-ml-service:8080"
	DefaultHTTPTimeout     = 30 * time.Second
	DefaultKubernetesQPS   = 50.0
	DefaultKubernetesBurst = 100
	DefaultEnableCORS      = false
)

// Valid log levels
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"fatal": true,
	"panic": true,
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnvAsInt("PORT", DefaultPort),
		MetricsPort:     getEnvAsInt("METRICS_PORT", DefaultMetricsPort),
		LogLevel:        getEnv("LOG_LEVEL", DefaultLogLevel),
		Kubeconfig:      getEnv("KUBECONFIG", ""),
		Namespace:       getEnv("NAMESPACE", DefaultNamespace),
		MLServiceURL:    getEnv("ML_SERVICE_URL", DefaultMLServiceURL),
		ArgocdAPIURL:    getEnv("ARGOCD_API_URL", ""),
		HTTPTimeout:     getEnvAsDuration("HTTP_TIMEOUT", DefaultHTTPTimeout),
		EnableCORS:      getEnvAsBool("ENABLE_CORS", DefaultEnableCORS),
		CORSAllowOrigin: getEnvAsSlice("CORS_ALLOW_ORIGIN", []string{"*"}),
		KubernetesQPS:   getEnvAsFloat32("KUBERNETES_QPS", DefaultKubernetesQPS),
		KubernetesBurst: getEnvAsInt("KUBERNETES_BURST", DefaultKubernetesBurst),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
//
//nolint:gocyclo // complexity acceptable for comprehensive config validation
func (c *Config) Validate() error {
	var errors []string

	// Validate port numbers
	if c.Port < 1 || c.Port > 65535 {
		errors = append(errors, fmt.Sprintf("invalid port: %d (must be 1-65535)", c.Port))
	}
	if c.MetricsPort < 1 || c.MetricsPort > 65535 {
		errors = append(errors, fmt.Sprintf("invalid metrics_port: %d (must be 1-65535)", c.MetricsPort))
	}
	if c.Port == c.MetricsPort {
		errors = append(errors, "port and metrics_port cannot be the same")
	}

	// Validate log level
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		errors = append(errors, fmt.Sprintf("invalid log_level: %s (must be debug, info, warn, error, fatal, or panic)", c.LogLevel))
	}

	// Validate namespace
	if c.Namespace == "" {
		errors = append(errors, "namespace cannot be empty")
	}

	// Validate ML service URL
	if c.MLServiceURL == "" {
		errors = append(errors, "ml_service_url cannot be empty")
	}
	if !strings.HasPrefix(c.MLServiceURL, "http://") && !strings.HasPrefix(c.MLServiceURL, "https://") {
		errors = append(errors, fmt.Sprintf("ml_service_url must start with http:// or https://: %s", c.MLServiceURL))
	}

	// Validate ArgoCD URL if provided
	if c.ArgocdAPIURL != "" {
		if !strings.HasPrefix(c.ArgocdAPIURL, "http://") && !strings.HasPrefix(c.ArgocdAPIURL, "https://") {
			errors = append(errors, fmt.Sprintf("argocd_api_url must start with http:// or https://: %s", c.ArgocdAPIURL))
		}
	}

	// Validate HTTP timeout
	if c.HTTPTimeout < 1*time.Second {
		errors = append(errors, fmt.Sprintf("http_timeout too short: %s (must be >= 1s)", c.HTTPTimeout))
	}
	if c.HTTPTimeout > 5*time.Minute {
		errors = append(errors, fmt.Sprintf("http_timeout too long: %s (must be <= 5m)", c.HTTPTimeout))
	}

	// Validate Kubernetes client settings
	if c.KubernetesQPS <= 0 {
		errors = append(errors, fmt.Sprintf("kubernetes_qps must be positive: %f", c.KubernetesQPS))
	}
	if c.KubernetesBurst <= 0 {
		errors = append(errors, fmt.Sprintf("kubernetes_burst must be positive: %d", c.KubernetesBurst))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}

// getEnvAsFloat32 gets an environment variable as a float32 or returns a default value
func getEnvAsFloat32(key string, defaultVal float32) float32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		return defaultVal
	}
	return float32(value)
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}

// getEnvAsDuration gets an environment variable as a duration or returns a default value
func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}

// getEnvAsSlice gets an environment variable as a comma-separated slice or returns a default value
func getEnvAsSlice(key string, defaultVal []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	parts := strings.Split(valueStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}
