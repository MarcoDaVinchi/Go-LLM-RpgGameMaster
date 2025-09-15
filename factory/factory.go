package factory

import (
	"fmt"
	"go-llm-rpggamemaster/config"
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
		cfg: cfg,
	}
}

// CreateLLMProvider creates an LLM provider instance based on the provider type and model name
func (f *providerFactory) CreateLLMProvider(providerType, modelName string) (interfaces.LLMProvider, error) {
	switch providerType {
	case "ollama":
		// Find the base URL from configuration
		baseURL := ""
		for _, provider := range f.cfg.LLMProviders {
			if provider.Name == "ollama" {
				baseURL = provider.BaseURL
				break
			}
		}
		return ollama.NewOllamaProvider(modelName, baseURL)
	case "openai":
		// Find the API key and base URL from configuration
		apiKey := ""
		baseURL := ""
		for _, provider := range f.cfg.LLMProviders {
			if provider.Name == "openai" {
				apiKey = provider.APIKey
				baseURL = provider.BaseURL
				break
			}
		}
		return openai.NewOpenAIProvider(modelName, apiKey, baseURL)
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}
}

// CreateEmbeddingProvider creates an embedding provider instance based on the provider type and model name
func (f *providerFactory) CreateEmbeddingProvider(providerType, modelName string) (interfaces.EmbeddingProvider, error) {
	switch providerType {
	case "ollama":
		// Find the base URL from configuration
		baseURL := ""
		for _, provider := range f.cfg.EmbeddingProviders {
			if provider.Name == "ollama" {
				baseURL = provider.BaseURL
				break
			}
		}
		return ollama.NewOllamaEmbeddingProvider(modelName, baseURL)
	case "openai":
		// Find the API key and base URL from configuration
		apiKey := ""
		baseURL := ""
		for _, provider := range f.cfg.EmbeddingProviders {
			if provider.Name == "openai" {
				apiKey = provider.APIKey
				baseURL = provider.BaseURL
				break
			}
		}
		return openai.NewOpenAIEmbeddingProvider(modelName, apiKey, baseURL)
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}
}

// CreateDefaultLLMProvider creates the default LLM provider based on configuration
func (f *providerFactory) CreateDefaultLLMProvider() (interfaces.LLMProvider, error) {
	defaultProvider := f.cfg.DefaultLLM
	for _, provider := range f.cfg.LLMProviders {
		if provider.Name == defaultProvider {
			return f.CreateLLMProvider(provider.Name, provider.Model)
		}
	}

	// If no default provider found, create a default Ollama provider
	return f.CreateLLMProvider("ollama", "llama3.1")
}

// CreateDefaultEmbeddingProvider creates the default embedding provider based on configuration
func (f *providerFactory) CreateDefaultEmbeddingProvider() (interfaces.EmbeddingProvider, error) {
	defaultProvider := f.cfg.DefaultEmbedding
	for _, provider := range f.cfg.EmbeddingProviders {
		if provider.Name == defaultProvider {
			return f.CreateEmbeddingProvider(provider.Name, provider.Model)
		}
	}

	// If no default provider found, create a default Ollama provider
	return f.CreateEmbeddingProvider("ollama", "nomic-embed-text")
}

// CreateRetriever creates a retriever instance based on the retriever type
func (f *providerFactory) CreateRetriever(retrieverType string, embedder interface{}) (retrievers.Retriever, error) {
	embedderImpl, ok := embedder.(*embeddings.EmbedderImpl)
	if !ok {
		return nil, fmt.Errorf("invalid embedder type")
	}

	switch retrieverType {
	case "sqlite":
		// Find the DB path from configuration
		dbPath := "base.db"
		for _, retriever := range f.cfg.Retrievers {
			if retriever.Name == "sqlite" {
				if retriever.DBPath != "" {
					dbPath = retriever.DBPath
				}
				break
			}
		}
		return retrievers.NewSQLiteRetrieverWithPath(embedderImpl, dbPath)
	case "qdrant":
		return retrievers.NewQdrantRetriever(embedderImpl)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}

// CreateDefaultRetriever creates the default retriever based on configuration
func (f *providerFactory) CreateDefaultRetriever(embedder interface{}) (retrievers.Retriever, error) {
	defaultRetriever := f.cfg.DefaultRetriever
	return f.CreateRetriever(defaultRetriever, embedder)
}
