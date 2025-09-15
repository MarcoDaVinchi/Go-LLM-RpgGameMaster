package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("QDRANT_URL")
	}()

	// Test loading config from default path when config.yaml doesn't exist
	cfg, err := LoadConfigFromDefaultPath()
	if err != nil {
		t.Errorf("Failed to load config from default path: %v", err)
	}

	// Check that we have the expected default configuration
	if len(cfg.LLMProviders) != 2 {
		t.Errorf("Expected 2 LLM providers, got %d", len(cfg.LLMProviders))
	}

	if len(cfg.EmbeddingProviders) != 2 {
		t.Errorf("Expected 2 embedding providers, got %d", len(cfg.EmbeddingProviders))
	}

	if cfg.DefaultLLM != "ollama" {
		t.Errorf("Expected default LLM 'ollama', got '%s'", cfg.DefaultLLM)
	}

	if cfg.DefaultEmbedding != "ollama" {
		t.Errorf("Expected default embedding 'ollama', got '%s'", cfg.DefaultEmbedding)
	}

	// Test loading config from YAML file
	// Create a temporary config file
	tempConfig := `llm_providers:
- name: ollama
  model: llama3.1
- name: openai
  model: gpt-3.5-turbo
  api_key: test-key

embedding_providers:
- name: ollama
  model: nomic-embed-text
- name: openai
  model: text-embedding-ada-002
  api_key: test-key

default_llm: openai
default_embedding: openai

retrievers:
- name: sqlite
  db_path: test.db

default_retriever: sqlite`

	// Write temporary config file
	err = os.WriteFile("test_config.yaml", []byte(tempConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}

	// Clean up temporary file
	defer os.Remove("test_config.yaml")

	// Load config from file
	cfg, err = LoadConfig("test_config.yaml")
	if err != nil {
		t.Errorf("Failed to load config from file: %v", err)
	}

	// Check that we have the expected configuration from file
	if len(cfg.LLMProviders) != 2 {
		t.Errorf("Expected 2 LLM providers, got %d", len(cfg.LLMProviders))
	}

	if len(cfg.EmbeddingProviders) != 2 {
		t.Errorf("Expected 2 embedding providers, got %d", len(cfg.EmbeddingProviders))
	}

	if cfg.DefaultLLM != "openai" {
		t.Errorf("Expected default LLM 'openai', got '%s'", cfg.DefaultLLM)
	}

	if cfg.DefaultEmbedding != "openai" {
		t.Errorf("Expected default embedding 'openai', got '%s'", cfg.DefaultEmbedding)
	}

	// Check specific provider configurations
	if cfg.LLMProviders[0].Name != "ollama" {
		t.Errorf("Expected first LLM provider name 'ollama', got '%s'", cfg.LLMProviders[0].Name)
	}

	if cfg.LLMProviders[1].Name != "openai" {
		t.Errorf("Expected second LLM provider name 'openai', got '%s'", cfg.LLMProviders[1].Name)
	}

	if cfg.EmbeddingProviders[0].Name != "ollama" {
		t.Errorf("Expected first embedding provider name 'ollama', got '%s'", cfg.EmbeddingProviders[0].Name)
	}

	if cfg.EmbeddingProviders[1].Name != "openai" {
		t.Errorf("Expected second embedding provider name 'openai', got '%s'", cfg.EmbeddingProviders[1].Name)
	}
}
