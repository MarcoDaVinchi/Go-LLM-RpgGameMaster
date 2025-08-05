package clients

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	defaultEmbeddingModel = "nomic-embed-text"
)

type EmbeddingClient interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	GetEmbedder() *embeddings.EmbedderImpl
}

type OllamaEmbeddingClient struct {
	embedder *embeddings.EmbedderImpl
}

func NewOllamaEmbeddingClient() (*OllamaEmbeddingClient, error) {
	embedLlm, err := ollama.New(ollama.WithModel(defaultEmbeddingModel))
	if err != nil {
		return nil, err
	}

	embedder, err := embeddings.NewEmbedder(embedLlm)
	if err != nil {
		return nil, err
	}

	return &OllamaEmbeddingClient{
		embedder: embedder,
	}, nil
}

func (c *OllamaEmbeddingClient) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return c.embedder.EmbedDocuments(ctx, texts)
}

func (c *OllamaEmbeddingClient) GetEmbedder() *embeddings.EmbedderImpl {
	return c.embedder
}
