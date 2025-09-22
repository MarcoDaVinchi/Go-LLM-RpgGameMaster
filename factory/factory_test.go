package factory

import (
	"testing"

	"go-llm-rpggamemaster/config"
	factoryinterface "go-llm-rpggamemaster/factory/interface"
	"go-llm-rpggamemaster/interfaces"
)

func TestProviderFactory(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		LLMProviders: []config.ProviderConfig{
			{
				Name:  "ollama",
				Model: "llama3.1",
			},
			{
				Name:   "openai",
				Model:  "gpt-3.5-turbo",
				APIKey: "test-key",
			},
		},
		EmbeddingProviders: []config.ProviderConfig{
			{
				Name:  "ollama",
				Model: "nomic-embed-text",
			},
			{
				Name:   "openai",
				Model:  "text-embedding-ada-002",
				APIKey: "test-key",
			},
		},
		DefaultLLM:       "ollama",
		DefaultEmbedding: "ollama",
	}

	// Create provider factory
	providerFactory := NewProviderFactory(cfg)

	// Check that the factory implements the interface
	if _, ok := providerFactory.(factoryinterface.ProviderFactory); !ok {
		t.Error("providerFactory does not implement ProviderFactory interface")
	}

	// Test creating Ollama LLM provider
	ollamaProvider, err := providerFactory.CreateInferenceProvider("ollama", "llama3.1")
	if err != nil {
		t.Errorf("Failed to create Ollama LLM provider: %v", err)
	}

	if ollamaProvider.Name() != "ollama" {
		t.Errorf("Expected Ollama provider name 'ollama', got '%s'", ollamaProvider.Name())
	}

	// Check that the provider implements the InferenceProvider interface
	if _, ok := ollamaProvider.(interfaces.InferenceProvider); !ok {
		t.Error("OllamaProvider does not implement InferenceProvider interface")
	}

	// Test creating OpenAI LLM provider
	openaiProvider, err := providerFactory.CreateInferenceProvider("openai", "gpt-3.5-turbo")
	if err != nil {
		t.Errorf("Failed to create OpenAI LLM provider: %v", err)
	}

	if openaiProvider.Name() != "openai" {
		t.Errorf("Expected OpenAI provider name 'openai', got '%s'", openaiProvider.Name())
	}

	// Check that the provider implements the InferenceProvider interface
	if _, ok := openaiProvider.(interfaces.InferenceProvider); !ok {
		t.Error("OpenAIProvider does not implement InferenceProvider interface")
	}

	// Test creating Ollama embedding provider
	ollamaEmbeddingProvider, err := providerFactory.CreateEmbeddingProvider("ollama", "nomic-embed-text")
	if err != nil {
		t.Errorf("Failed to create Ollama embedding provider: %v", err)
	}

	if ollamaEmbeddingProvider.Name() != "ollama" {
		t.Errorf("Expected Ollama embedding provider name 'ollama', got '%s'", ollamaEmbeddingProvider.Name())
	}

	// Check that the provider implements the VectorEmbeddingProvider interface
	if _, ok := ollamaEmbeddingProvider.(interfaces.VectorEmbeddingProvider); !ok {
		t.Error("OllamaEmbeddingProvider does not implement VectorEmbeddingProvider interface")
	}

	// Test creating OpenAI embedding provider
	openaiEmbeddingProvider, err := providerFactory.CreateEmbeddingProvider("openai", "text-embedding-ada-002")
	if err != nil {
		t.Errorf("Failed to create OpenAI embedding provider: %v", err)
	}

	if openaiEmbeddingProvider.Name() != "openai" {
		t.Errorf("Expected OpenAI embedding provider name 'openai', got '%s'", openaiEmbeddingProvider.Name())
	}

	// Check that the provider implements the VectorEmbeddingProvider interface
	if _, ok := openaiEmbeddingProvider.(interfaces.VectorEmbeddingProvider); !ok {
		t.Error("OpenAIEmbeddingProvider does not implement VectorEmbeddingProvider interface")
	}

	// Test creating default LLM provider
	defaultLLMProvider, err := providerFactory.CreateDefaultLLMProvider()
	if err != nil {
		t.Errorf("Failed to create default LLM provider: %v", err)
	}

	if defaultLLMProvider.Name() != "ollama" {
		t.Errorf("Expected default LLM provider name 'ollama', got '%s'", defaultLLMProvider.Name())
	}

	// Test creating default embedding provider
	defaultEmbeddingProvider, err := providerFactory.CreateDefaultEmbeddingProvider()
	if err != nil {
		t.Errorf("Failed to create default embedding provider: %v", err)
	}

	if defaultEmbeddingProvider.Name() != "ollama" {
		t.Errorf("Expected default embedding provider name 'ollama', got '%s'", defaultEmbeddingProvider.Name())
	}
}
