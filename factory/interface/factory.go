package factoryinterface

import (
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/retrievers"
)

// ProviderFactory is responsible for creating provider instances
type ProviderFactory interface {
	CreateInferenceProvider() (interfaces.InferenceProvider, error)
	CreateEmbeddingProvider() (interfaces.VectorEmbeddingProvider, error)
	CreateRetriever(embedder interfaces.VectorEmbeddingProvider, retrieverType string) (retrievers.Retriever, error)
}
