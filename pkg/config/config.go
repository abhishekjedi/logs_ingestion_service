package config

import (
	"log"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func NewDefaultConfigProvider() *koanf.Koanf {
	k := koanf.New(".")

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Optional, gitignored overlay for local secrets (e.g. google_client_secret).
	// Keys here override config.yaml.
	if _, err := os.Stat("config.local.yaml"); err == nil {
		if err := k.Load(file.Provider("config.local.yaml"), yaml.Parser()); err != nil {
			log.Fatalf("Failed to load config.local.yaml: %v", err)
		}
	}

	return k
}
