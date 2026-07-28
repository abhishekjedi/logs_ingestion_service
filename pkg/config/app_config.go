package config

import "github.com/knadh/koanf/v2"

type AppConfig struct {
	BaseURL string

	Port string

	FrontendURL string
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
	frontend := k.String("app.frontend_url")
	if frontend == "" {
		frontend = "http://localhost:5173"
	}
	return AppConfig{BaseURL: base, Port: port, FrontendURL: frontend}
}

func (c AppConfig) Addr() string {
	return ":" + c.Port
}
