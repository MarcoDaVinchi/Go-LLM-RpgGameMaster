package interfaces

import (
	"context"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	// GenerateResponse generates a response using the provider
	GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)

	// GenerateSimpleResponse generates a simple response using the provider
	GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error)

	// Name returns the name of the provider
	Name() string
}

// EmbeddingProvider defines the interface for embedding providers
type EmbeddingProvider interface {
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
