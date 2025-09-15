package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProviderConfig represents the configuration for a provider
type ProviderConfig struct {
	Name     string `yaml:"name"`
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key,omitempty"`
	BaseURL  string `yaml:"base_url,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

// RetrieverConfig represents the configuration for a retriever
type RetrieverConfig struct {
	Name       string `yaml:"name"`
	DBPath     string `yaml:"db_path,omitempty"`
	Collection string `yaml:"collection,omitempty"`
}

// Config represents the overall configuration structure
type Config struct {
	LLMProviders       []ProviderConfig  `yaml:"llm_providers"`
	EmbeddingProviders []ProviderConfig  `yaml:"embedding_providers"`
	DefaultLLM         string            `yaml:"default_llm"`
	DefaultEmbedding   string            `yaml:"default_embedding"`
	Retrievers         []RetrieverConfig `yaml:"retrievers"`
	DefaultRetriever   string            `yaml:"default_retriever"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigFromDefaultPath loads the configuration from the default path
func LoadConfigFromDefaultPath() (*Config, error) {
	// Try to load from config.yaml in the current directory
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("config", "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("no config file found")
		}
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}
