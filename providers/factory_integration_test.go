package providers

import (
	"context"
	"testing"

	"go-llm-rpggamemaster/config"
	factory "go-llm-rpggamemaster/factory"
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/providers/ollama"
	"go-llm-rpggamemaster/providers/openai"
)

func TestFactoryIntegration(t *testing.T) {
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
	providerFactory := factory.NewProviderFactory(cfg)

	// Test creating all provider combinations
	testCases := []struct {
		name          string
		providerType  string
		modelName     string
		shouldSucceed bool
	}{
		{"Ollama LLM", "ollama", "llama3.1", true},
		{"OpenAI LLM", "openai", "gpt-3.5-turbo", true},
		{"Invalid LLM", "invalid", "model", false},
		{"Ollama Embedding", "ollama", "nomic-embed-text", true},
		{"OpenAI Embedding", "openai", "text-embedding-ada-002", true},
		{"Invalid Embedding", "invalid", "model", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test LLM provider creation
			llmProvider, err := providerFactory.CreateInferenceProvider(tc.providerType, tc.modelName)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Failed to create LLM provider: %v", err)
				} else {
					// Verify the provider implements the interface
					if _, ok := llmProvider.(interfaces.InferenceProvider); !ok {
						t.Error("Provider does not implement InferenceProvider interface")
					}

					// Verify the provider name
					if llmProvider.Name() != tc.providerType {
						t.Errorf("Expected provider name '%s', got '%s'", tc.providerType, llmProvider.Name())
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				}
			}

			// Test embedding provider creation
			embeddingProvider, err := providerFactory.CreateEmbeddingProvider(tc.providerType, tc.modelName)
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Failed to create embedding provider: %v", err)
				} else {
					// Verify the provider implements the interface
					if _, ok := embeddingProvider.(interfaces.VectorEmbeddingProvider); !ok {
						t.Error("Provider does not implement VectorEmbeddingProvider interface")
					}

					// Verify the provider name
					if embeddingProvider.Name() != tc.providerType {
						t.Errorf("Expected provider name '%s', got '%s'", tc.providerType, embeddingProvider.Name())
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				}
			}
		})
	}

	// Test default provider creation
	t.Run("Default Providers", func(t *testing.T) {
		defaultLLMProvider, err := providerFactory.CreateDefaultLLMProvider()
		if err != nil {
			t.Errorf("Failed to create default LLM provider: %v", err)
		}

		if defaultLLMProvider.Name() != "ollama" {
			t.Errorf("Expected default LLM provider name 'ollama', got '%s'", defaultLLMProvider.Name())
		}

		defaultEmbeddingProvider, err := providerFactory.CreateDefaultEmbeddingProvider()
		if err != nil {
			t.Errorf("Failed to create default embedding provider: %v", err)
		}

		if defaultEmbeddingProvider.Name() != "ollama" {
			t.Errorf("Expected default embedding provider name 'ollama', got '%s'", defaultEmbeddingProvider.Name())
		}
	})
}

// Test that the providers can be used in a typical application flow
func TestProviderUsageFlow(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		LLMProviders: []config.ProviderConfig{
			{
				Name:  "ollama",
				Model: "llama3.1",
			},
		},
		EmbeddingProviders: []config.ProviderConfig{
			{
				Name:  "ollama",
				Model: "nomic-embed-text",
			},
		},
		DefaultLLM:       "ollama",
		DefaultEmbedding: "ollama",
	}

	// Create provider factory
	providerFactory := factory.NewProviderFactory(cfg)

	// Create providers
	llmProvider, err := providerFactory.CreateDefaultLLMProvider()
	if err != nil {
		t.Fatalf("Failed to create default LLM provider: %v", err)
	}

	embeddingProvider, err := providerFactory.CreateDefaultEmbeddingProvider()
	if err != nil {
		t.Fatalf("Failed to create default embedding provider: %v", err)
	}

	// Verify we can get provider names
	if llmProvider.Name() == "" {
		t.Error("LLM provider name is empty")
	}

	if embeddingProvider.Name() == "" {
		t.Error("Embedding provider name is empty")
	}

	// Verify we can call methods (they will fail without a running service, but we can check method existence)
	ctx := context.Background()

	// Test that GenerateSimpleResponse method exists (will fail without service)
	_, _ = llmProvider.GenerateSimpleResponse(ctx, "test")

	// Test that EmbedDocuments method exists (will fail without service)
	_, _ = embeddingProvider.EmbedDocuments(ctx, []string{"test"})
}

// Test interface compliance for all provider implementations
func TestProviderInterfaceCompliance(t *testing.T) {
	// Test that OllamaProvider implements InferenceProvider
	var _ interfaces.InferenceProvider = (*ollama.OllamaProvider)(nil)

	// Test that OpenAIProvider implements InferenceProvider
	var _ interfaces.InferenceProvider = (*openai.OpenAIProvider)(nil)

	// Test that OllamaEmbeddingProvider implements VectorEmbeddingProvider
	var _ interfaces.VectorEmbeddingProvider = (*ollama.OllamaEmbeddingProvider)(nil)

	// Test that OpenAIEmbeddingProvider implements VectorEmbeddingProvider
	var _ interfaces.VectorEmbeddingProvider = (*openai.OpenAIEmbeddingProvider)(nil)
}
