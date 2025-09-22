package interfaces

import (
	"context"
)

// InferenceProvider defines the interface for LLM providers
type InferenceProvider interface {
	// GenerateResponse generates a response using the provider
	GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)

	// Name returns the name of the provider
	Name() string
}

// VectorEmbeddingProvider defines the interface for embedding providers
type VectorEmbeddingProvider interface {
	// EmbedDocuments embeds a list of documents
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)

	// Name returns the name of the provider
	Name() string
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
