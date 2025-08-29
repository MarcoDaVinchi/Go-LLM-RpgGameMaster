package clients

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

type EmbeddingClient interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	GetEmbedder() *embeddings.EmbedderImpl
}

type LlmEmbeddingClient struct {
	embedder *embeddings.EmbedderImpl
}

// NewLlmEmbeddingClient creates a new LlmEmbeddingClient instance using the specified model.
// It initializes an OpenAI embedding model and configures it with the provided modelName.
//
// Parameters:
//   - modelName: The name of the OpenAI model to use for embeddings
//
// Returns:
//   - *LlmEmbeddingClient: A new client instance
//   - error: An error if initialization fails
func NewLlmEmbeddingClient(modelName string) (*LlmEmbeddingClient, error) {
	embedLlm, err := openai.New(openai.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding client: %w", err)
	}

	embedder, err := embeddings.NewEmbedder(embedLlm)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder based on embedding client: %w", err)
	}

	return &LlmEmbeddingClient{
		embedder: embedder,
	}, nil
}

func (c *LlmEmbeddingClient) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return c.embedder.EmbedDocuments(ctx, texts)
}

func (c *LlmEmbeddingClient) GetEmbedder() *embeddings.EmbedderImpl {
	return c.embedder
}
