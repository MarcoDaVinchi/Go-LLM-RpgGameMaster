package factory

import (
	"fmt"
	config "go-llm-rpggamemaster/config"
	factoryinterface "go-llm-rpggamemaster/factory/interface"
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/providers/ollama"
	"go-llm-rpggamemaster/providers/openai"
	"go-llm-rpggamemaster/retrievers"

	"github.com/tmc/langchaingo/embeddings"
)

// providerFactory implements the ProviderFactory interface
type providerFactory struct {
	cfg *config.Config
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(cfg *config.Config) factoryinterface.ProviderFactory {
	return &providerFactory{
		cfg,
	}
}

// CreateLLMProvider creates an LLM provider instance based on the provider type and model name
func (f *providerFactory) CreateInferenceProvider() (interfaces.InferenceProvider, error) {
	inferenceModel := f.cfg.InferenceModel
	baseURL := inferenceModel.Url
	apiKey := inferenceModel.ApiKey
	modelName := inferenceModel.Name
	providerType := inferenceModel.Type

	switch inferenceModel.Type {
	case config.ModelTypeOllama:
		return ollama.NewOllamaProvider(modelName, baseURL)
	case config.ModelTypeOpenAI:
		return openai.NewOpenAIProvider(modelName, apiKey, baseURL)
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
	case config.ModelTypeOllama:
		return ollama.NewOllamaEmbeddingProvider(modelName, baseURL)
	case config.ModelTypeOpenAI:
		return openai.NewOpenAIEmbeddingProvider(modelName, apiKey, baseURL)
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}
}

func (f *providerFactory) CreateRetriever() (retrievers.Retriever, error) {
	embedderImpl, ok := embedder.(*embeddings.EmbedderImpl)
	if !ok {
		return nil, fmt.Errorf("invalid embedder type")
	}

	switch retrieverType {
	case "sqlite":
		// Find the DB path from configuration
		dbPath := "base.db"
		return retrievers.NewSQLiteRetrieverWithPath(embedderImpl, dbPath)
	case "qdrant":
		return retrievers.NewQdrantRetriever(embedderImpl)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}
