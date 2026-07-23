package config

import "github.com/knadh/koanf/v2"

// AppConfig holds process-level settings not tied to a specific client.
type AppConfig struct {
	// BaseURL is the externally reachable base used to build per-service ingest
	// URLs handed back at registration time.
	BaseURL string
	// Port is the TCP port the HTTP server listens on.
	Port string
}

func NewAppConfig(k *koanf.Koanf) AppConfig {
	base := k.String("app.base_url")
	if base == "" {
		base = "http://localhost:8080"
	}
	port := k.String("app.port")
	if port == "" {
		port = "8080"
	}
	return AppConfig{BaseURL: base, Port: port}
}

// Addr returns the listen address, e.g. ":8080".
func (c AppConfig) Addr() string {
	return ":" + c.Port
}
