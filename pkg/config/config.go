package config

import (
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func NewDefaultConfigProvider() *koanf.Koanf {
	k := koanf.New(".")

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	return k
}
