package factoryinterface

import (
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/retrievers"
)

// ProviderFactory is responsible for creating provider instances
type ProviderFactory interface {
	// CreateLLMProvider creates an LLM provider instance
	CreateLLMProvider(providerType, modelName string) (interfaces.LLMProvider, error)

	// CreateEmbeddingProvider creates an embedding provider instance
	CreateEmbeddingProvider(providerType, modelName string) (interfaces.EmbeddingProvider, error)

	// CreateDefaultLLMProvider creates the default LLM provider based on configuration
	CreateDefaultLLMProvider() (interfaces.LLMProvider, error)

	// CreateDefaultEmbeddingProvider creates the default embedding provider based on configuration
	CreateDefaultEmbeddingProvider() (interfaces.EmbeddingProvider, error)

	// CreateRetriever creates a retriever instance
	CreateRetriever(retrieverType string, embedder interface{}) (retrievers.Retriever, error)

	// CreateDefaultRetriever creates the default retriever based on configuration
	CreateDefaultRetriever(embedder interface{}) (retrievers.Retriever, error)
}
