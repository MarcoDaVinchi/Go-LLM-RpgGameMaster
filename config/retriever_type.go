package config

import (
	"fmt"
	"strings"
)

type RetrieverType uint8

const (
	RetrieverTypeUnknown RetrieverType = iota
	RetrieverTypeSqlite
	RetrieverTypeQdrant
	RetrieverTypePostgres
)

func (t RetrieverType) String() string {
	switch t {
	case RetrieverTypeSqlite:
		return "Sqlite"
	case RetrieverTypeQdrant:
		return "Qdrant"
	case RetrieverTypePostgres:
		return "Postgres"
	default:
		return "Unknown"
	}
}

func (t *RetrieverType) UnmarshalText(text []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(text))) {
	case "sqlite":
		*t = RetrieverTypeSqlite
	case "qdrant":
		*t = RetrieverTypeQdrant
	case "postgres":
		*t = RetrieverTypePostgres
	default:
		*t = RetrieverTypeUnknown
		return fmt.Errorf("invalid retriever type: %s", text)
	}
	return nil
}

func (t RetrieverType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type VectorRetriever struct {
	Name string        `mapstructure:"name"`
	Url  string        `mapstructure:"url"`
	Type RetrieverType `mapstructure:"type"`
}
