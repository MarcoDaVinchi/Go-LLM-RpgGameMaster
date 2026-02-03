package factory

import (
	"testing"

	"go-llm-rpggamemaster/config"
	factoryinterface "go-llm-rpggamemaster/factory/interface"
	"go-llm-rpggamemaster/interfaces"
)

func TestProviderFactory(t *testing.T) {
	cfg := &config.Config{
		InferenceModel: config.LLModel{
			Name:   "gpt-4o-mini",
			Url:    "https://routerai.ru/v1",
			Type:   config.ModelTypeRouterAI,
			ApiKey: "test-key",
		},
		EmbeddingModel: config.LLModel{
			Name:   "text-embedding-ada-002",
			Url:    "https://routerai.ru/v1",
			Type:   config.ModelTypeRouterAI,
			ApiKey: "test-key",
		},
		VectorRetriever: config.VectorRetriever{
			Name: "qdrant",
			Url:  "http://localhost:6333",
			Type: config.RetrieverTypeQdrant,
		},
	}

	providerFactory := NewProviderFactory(cfg)

	if _, ok := providerFactory.(factoryinterface.ProviderFactory); !ok {
		t.Error("providerFactory does not implement ProviderFactory interface")
	}

	t.Run("CreateInferenceProvider", func(t *testing.T) {
		provider, err := providerFactory.CreateInferenceProvider()
		if err != nil {
			t.Errorf("Failed to create inference provider: %v", err)
			return
		}

		if provider.Name() != "routerai" {
			t.Errorf("Expected provider name 'routerai', got '%s'", provider.Name())
		}

		if _, ok := provider.(interfaces.InferenceProvider); !ok {
			t.Error("Provider does not implement InferenceProvider interface")
		}
	})

	t.Run("CreateEmbeddingProvider", func(t *testing.T) {
		provider, err := providerFactory.CreateEmbeddingProvider()
		if err != nil {
			t.Errorf("Failed to create embedding provider: %v", err)
			return
		}

		if provider.Name() != "routerai" {
			t.Errorf("Expected provider name 'routerai', got '%s'", provider.Name())
		}

		if _, ok := provider.(interfaces.VectorEmbeddingProvider); !ok {
			t.Error("Provider does not implement VectorEmbeddingProvider interface")
		}
	})

	t.Run("CreateRetriever", func(t *testing.T) {
		embedder, _ := providerFactory.CreateEmbeddingProvider()
		retriever, err := providerFactory.CreateRetriever(embedder, "sqlite")
		if err != nil {
			t.Errorf("Failed to create retriever: %v", err)
			return
		}

		if retriever == nil {
			t.Error("Expected retriever but got nil")
		}
	})
}

func TestDeprecatedProviders(t *testing.T) {
	cfg := &config.Config{
		InferenceModel: config.LLModel{
			Name:   "gpt-4",
			Type:   config.ModelTypeOpenAI,
			ApiKey: "test-key",
		},
	}

	providerFactory := NewProviderFactory(cfg)

	t.Run("OpenAIDeprecated", func(t *testing.T) {
		_, err := providerFactory.CreateInferenceProvider()
		if err == nil {
			t.Error("Expected error for deprecated OpenAI provider, got nil")
		}
	})

	cfg.InferenceModel.Type = config.ModelTypeOllama
	providerFactory = NewProviderFactory(cfg)

	t.Run("OllamaDeprecated", func(t *testing.T) {
		_, err := providerFactory.CreateInferenceProvider()
		if err == nil {
			t.Error("Expected error for deprecated Ollama provider, got nil")
		}
	})
}
