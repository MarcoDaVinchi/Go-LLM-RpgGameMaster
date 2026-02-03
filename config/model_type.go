package config

import (
	"fmt"
	"strings"
)

type ModelType uint8

const (
	ModelTypeUnknown ModelType = iota
	ModelTypeOpenAI
	ModelTypeOllama
	ModelTypeRouterAI
)

func (t ModelType) String() string {
	switch t {
	case ModelTypeOpenAI:
		return "OpenAI"
	case ModelTypeOllama:
		return "Ollama"
	case ModelTypeRouterAI:
		return "RouterAI"
	default:
		return "Unknown"
	}
}

// Go
//
//nolint:gocritic // Unmarshall needs to be pointer receiver
func (t *ModelType) UnmarshalText(text []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(text))) {
	case "openai":
		*t = ModelTypeOpenAI
	case "ollama":
		*t = ModelTypeOllama
	case "routerai":
		*t = ModelTypeRouterAI
	default:
		*t = ModelTypeUnknown
		return fmt.Errorf("invalid model type: %s", text)
	}
	return nil
}

func (t ModelType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type LLModel struct {
	Name   string    `mapstructure:"name"`
	Url    string    `mapstructure:"url"`
	Type   ModelType `mapstructure:"type"`
	ApiKey string    `mapstructure:"api_key"`
}
