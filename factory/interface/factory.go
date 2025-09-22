package factoryinterface

import (
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/retrievers"
)

// ProviderFactory is responsible for creating provider instances
type ProviderFactory interface {
	// CreateLLMProvider creates an LLM provider instance
	CreateInferenceProvider() (interfaces.InferenceProvider, error)

	// CreateEmbeddingProvider creates an embedding provider instance
	CreateEmbeddingProvider() (interfaces.VectorEmbeddingProvider, error)

	// CreateRetriever creates a retriever instance
	CreateRetriever() (retrievers.Retriever, error)
}
