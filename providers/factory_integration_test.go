package providers

import (
	"testing"

	"go-llm-rpggamemaster/config"
	"go-llm-rpggamemaster/factory"
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/providers/routerai"
)

func TestFactoryIntegration(t *testing.T) {
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
	}

	providerFactory := factory.NewProviderFactory(cfg)

	t.Run("Create Inference Provider", func(t *testing.T) {
		provider, err := providerFactory.CreateInferenceProvider()
		if err != nil {
			t.Errorf("Failed to create inference provider: %v", err)
			return
		}
		if provider.Name() != "routerai" {
			t.Errorf("Expected provider name 'routerai', got '%s'", provider.Name())
		}
	})

	t.Run("Create Embedding Provider", func(t *testing.T) {
		provider, err := providerFactory.CreateEmbeddingProvider()
		if err != nil {
			t.Errorf("Failed to create embedding provider: %v", err)
			return
		}
		if provider.Name() != "routerai" {
			t.Errorf("Expected provider name 'routerai', got '%s'", provider.Name())
		}
	})
}

func TestProviderInterfaceCompliance(t *testing.T) {
	var (
		_ interfaces.InferenceProvider       = (*routerai.RouterAIProvider)(nil)
		_ interfaces.VectorEmbeddingProvider = (*routerai.RouterAIProvider)(nil)
	)
}
