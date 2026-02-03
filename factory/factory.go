package factory

import (
	"fmt"

	config "go-llm-rpggamemaster/config"
	factoryinterface "go-llm-rpggamemaster/factory/interface"
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/providers/routerai"
	"go-llm-rpggamemaster/retrievers"
)

type providerFactory struct {
	cfg *config.Config
}

func NewProviderFactory(cfg *config.Config) factoryinterface.ProviderFactory {
	return &providerFactory{
		cfg,
	}
}

func (f *providerFactory) CreateInferenceProvider() (interfaces.InferenceProvider, error) {
	inferenceModel := f.cfg.InferenceModel
	baseURL := inferenceModel.Url
	apiKey := inferenceModel.ApiKey
	modelName := inferenceModel.Name
	providerType := inferenceModel.Type

	switch inferenceModel.Type {
	case config.ModelTypeRouterAI:
		return routerai.NewRouterAIProvider(modelName, apiKey, baseURL)
	case config.ModelTypeOpenAI, config.ModelTypeOllama:
		return nil, fmt.Errorf("provider type %s is deprecated, use routerai", providerType)
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}
}

func (f *providerFactory) CreateEmbeddingProvider() (interfaces.VectorEmbeddingProvider, error) {
	embeddingModel := f.cfg.EmbeddingModel
	baseURL := embeddingModel.Url
	apiKey := embeddingModel.ApiKey
	modelName := embeddingModel.Name
	providerType := embeddingModel.Type

	switch providerType {
	case config.ModelTypeRouterAI:
		return routerai.NewRouterAIProvider(modelName, apiKey, baseURL)
	case config.ModelTypeOpenAI, config.ModelTypeOllama:
		return nil, fmt.Errorf("provider type %s is deprecated, use routerai", providerType)
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}
}

func (f *providerFactory) CreateRetriever(embedder interfaces.VectorEmbeddingProvider, retrieverType string) (retrievers.Retriever, error) {
	switch retrieverType {
	case "sqlite":
		dbPath := "base.db"
		return retrievers.NewSQLiteRetrieverWithPath(embedder, dbPath)
	case "qdrant":
		return retrievers.NewQdrantRetriever(embedder)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}
